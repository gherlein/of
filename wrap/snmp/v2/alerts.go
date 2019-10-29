package v2

import (
	"fmt"
	"time"

	prommodel "github.com/prometheus/common/model"
	of "github.com/cisco-cx/of/pkg/v2"
	of_snmp "github.com/cisco-cx/of/pkg/v2/snmp"
	logger "github.com/cisco-cx/of/wrap/logrus/v2"
)

// Implements of_snmp.AlertGenerator
type Alerter struct {
	Log     *logger.Logger
	Configs *of_snmp.V2Config
	Source  *of.TrapSource
	Value   *Value
}

// Iterate through configs in configNames and generate all possible Alerts.
func (a *Alerter) Alert(cfgNames []string) ([]of.Alert, error) {

	var allAlerts = make([]of.Alert, 0)
	for _, cfgName := range cfgNames {

		var cfg of_snmp.Config
		var ok bool

		// Check if config exists
		if cfg, ok = (*a.Configs)[cfgName]; ok == false {
			a.Log.WithError(of.ErrConfigNotFound).Errorf("")
			continue
		}

		// Check through of_snmp.Config.Alerts to find a match.
		for aNum, alertCfg := range cfg.Alerts {
			// Check if alert is enabled.
			a.Log.Debugf("Checking Alert no. %d (%v) in config: %s", aNum, alertCfg.Name, cfgName)
			if a.enabled(cfg.Defaults.Enabled, alertCfg.Enabled) == false {
				// Printing alert no., since alert name can be nil.
				a.Log.Debugf("Alert no. %d (%v), not enabled in config: %s", aNum, alertCfg.Name, cfgName)
				continue
			}

			// Check if trap Vars have any alert matching firing conditions.
			fAlert, err := a.matchAlerts(cfg, alertCfg, of_snmp.Firing)
			if err == nil {
				fAlert.Annotations[string(of_snmp.EventTypeText)] = string(of_snmp.Firing)
				allAlerts = append(allAlerts, fAlert)
			}

			// Check if trap Vars have any alerts matching clearing conditions.
			cAlert, err := a.matchAlerts(cfg, alertCfg, of_snmp.Clearing)
			if err == nil {
				// Add end time to clearing alerts.
				cAlert.Annotations[string(of_snmp.EventTypeText)] = string(of_snmp.Clearing)
				cAlert.EndsAt = time.Now().Format(of.AMTimeFormat)
				allAlerts = append(allAlerts, cAlert)
			}
		}

	}
	return allAlerts, nil
}

// match trapVars with alerts in config.
func (a *Alerter) matchAlerts(cfg of_snmp.Config, alertCfg of_snmp.Alert, alertType of_snmp.EventType) (of.Alert, error) {

	var alert = of.Alert{}

	var selects []of_snmp.Select
	switch alertType {
	case of_snmp.Firing:
		selects = alertCfg.Firing["select"]
	case of_snmp.Clearing:
		selects = alertCfg.Clearing["select"]
	default:
		return alert, of.ErrUnknownEventType
	}

	// Check if trap Vars have any alerts matching select conditions.
	matched, err := a.selected(selects)
	if err != nil {
		a.Log.WithError(err).Errorf("Error while trying to match alert.")
		return alert, err
	}

	if matched == false {
		a.Log.WithError(of.ErrNoMatch).WithField("alertType", alertType).Debugf("")
		return alert, of.ErrNoMatch
	}

	a.Log.WithField("alertType", alertType).Debugf("Alert matched.")
	// Prepare alert since the selects have matched.

	// Preparing base alert.
	alert, err = a.prepareBaseAlert(&cfg)
	if err != nil {
		a.Log.WithError(err).Errorf("Error while preparing base alert.")
		return alert, err
	}

	// Apply alert specific labels.
	err = a.applyMod(&alert.Labels, alertCfg.LabelMods)
	if err != nil {
		a.Log.WithError(err).Errorf("Error while applying alert mods to labels.")
		return alert, err
	}

	// Apply alert specific annotations.
	err = a.applyMod(&alert.Annotations, alertCfg.LabelMods)
	if err != nil {
		a.Log.WithError(err).Errorf("Error while applying alert mods to annotations.")
		return alert, err
	}

	// Generator URL prefix.
	generatorURLPrefix := a.generatorURLPrefix(cfg.Defaults.GeneratorUrlPrefix, alertCfg.GeneratorUrlPrefix)

	// Apply select specfic changes.
	for _, sel := range selects {

		// Apply select specific annotations.
		err = a.applyMod(&alert.Annotations, sel.AnnotationMods)
		if err != nil {
			a.Log.WithError(err).Errorf("Error while applying alert mods to annotations.")
			return alert, err
		}

		// TODO : Decide on how to handle OIDs from multiple selects.
		// Update generator URL.
		if generatorURLPrefix != "" {
			// Prepend comma from second select onwards.
			seperator := ","
			if alert.GeneratorURL == "" {
				seperator = ""
			}
			alert.GeneratorURL += fmt.Sprintf("%s%s%s", seperator, generatorURLPrefix, sel.Oid)
		}
	}

	// Finger print the alert.
	fingerprint := a.fingerprint(alert)
	alert.Labels[of_snmp.FingerprintText] = fingerprint
	return alert, nil
}

// Check if give selects are applicable.
func (a *Alerter) selected(selects []of_snmp.Select) (bool, error) {
	// If none selects are mentioned, then match should fail.
	if len(selects) == 0 {
		return false, nil
	}

	allSelectsMatched := true
	for _, sel := range selects {
		// Find value based on the of_snmp.As type.
		resolvedValue, err := a.Value.ValueAs(sel.Oid, sel.As)
		if err != nil {
			a.Log.WithError(err).Errorf("Failed to resolve %s as %s.", sel.Oid, sel.As)
			return false, err
		}

		// Check if value is present in of_snmp.Config.Alerts[x].Select[y].Values
		valueFound := false
		for _, v := range sel.Values {
			if v == resolvedValue {
				valueFound = true
				break
			}
		}

		// If value is not found in of_snmp.Config.Alerts[x].Select[y].Values, stop checking for other selects.
		if valueFound == false {
			allSelectsMatched = false
			break
		}

	}

	return allSelectsMatched, nil
}

// Prepares the base alert based on keys under of_snmp.Config.Defaults
func (a *Alerter) prepareBaseAlert(cfg *of_snmp.Config) (of.Alert, error) {
	// Init new Alert
	alert := of.Alert{}
	alert.Labels = make(map[string]string)
	alert.Annotations = make(map[string]string)

	// Update source info.
	var found = false
	if cfg.Defaults.SourceType == of_snmp.ClusterType {
		// Iterate through clusters to check if IP matches with available IP.
		for clusterName, cluster := range cfg.Defaults.Clusters {
			for _, ip := range cluster.SourceAddresses {
				if ip == a.Source.Address {
					a.Log.Debugf("Found cluster name %s, for source IP : %s", clusterName, ip)
					found = true
					alert.Labels["source_address"] = clusterName
					alert.Labels["source_hostname"] = clusterName
					alert.Annotations["source_address"] = clusterName
					alert.Annotations["source_hostname"] = clusterName
					break
				}
			}
		}

	}

	// If no cluster is found or host type is not cluster.
	if found == false {
		a.Log.Debugf("Setting default source info for IP : %s", a.Source.Address)
		a.updateSource(&alert)
	}

	// Apply default mods to Labels
	err := a.applyMod(&alert.Labels, cfg.Defaults.LabelMods)
	if err != nil {
		a.Log.WithError(err).Errorf("Failed to apply default mods to labels.")
		return alert, err
	}

	// Apply default mods to Annotations
	err = a.applyMod(&alert.Annotations, cfg.Defaults.AnnotationMods)
	if err != nil {
		a.Log.WithError(err).Errorf("Failed to apply default mods to annotations.")
		return alert, err
	}

	return alert, nil
}

func (a *Alerter) applyMod(mapPtr *map[string]string, mods []of_snmp.Mod) error {

	// Init modifier
	m := Modifier{
		V: a.Value,
	}

	// Apply mods
	m.Map = mapPtr
	err := m.Apply(mods)
	if err != nil {
		return err
	}
	return nil
}

// Update source info based on of.TrapSource
func (a *Alerter) updateSource(alert *of.Alert) {
	alert.Labels["source_address"] = a.Source.Address
	alert.Labels["source_hostname"] = a.Source.Hostname
	alert.Annotations["source_address"] = a.Source.Address
	alert.Annotations["source_hostname"] = a.Source.Hostname
}

// Overwrite with alert specific prefix if available.
func (a *Alerter) generatorURLPrefix(defPrefix of_snmp.URLPrefix, alertPrefix of_snmp.URLPrefix) of_snmp.URLPrefix {
	if defPrefix == "" && alertPrefix == "" {
		return ""
	}

	// If alertPrefix is defined, use it.
	if alertPrefix != "" {
		return alertPrefix
	}

	return defPrefix
}

// Decide if alert should be sent or not based the of_snmp.Config.Defaults.Enabled and of_snmp.Config.Alerts[name].Enabled
//
// defaults.enabled 	alerts[n].enabled 	State
// Undefined 			Undefined 			Enabled
// Undefined 			false 				Disabled
// Undefined 			true 				Enabled
// false 				Undefined 			Disabled
// false 				false 				Disabled
// false 				true 				Disabled
// true 				Undefined 			Enabled
// true 				false 				Disabled
// true 				true 				Enabled
func (a *Alerter) enabled(defEnabled *bool, alertEnabled *bool) bool {
	// If both are undefined.
	if defEnabled == nil && alertEnabled == nil {
		return true
	}

	// If default is undefined, but alertEnabled is defined.
	if defEnabled == nil {
		return *alertEnabled
	}

	// If default is false.
	if *defEnabled == false {
		return false
	}

	// If alertEnabled is undefined.
	if alertEnabled == nil {
		return *defEnabled
	}

	// If defEnabled is true.
	return *alertEnabled
}

// Fingerprint the alert.
func (a *Alerter) fingerprint(al of.Alert) string {
	labels := make(prommodel.LabelSet)
	for k, v := range al.Labels {
		labels[prommodel.LabelName(k)] = prommodel.LabelValue(v)
	}
	return labels.Fingerprint().String()
}