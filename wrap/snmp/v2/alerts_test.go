package v2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	of "github.com/cisco-cx/of/pkg/v2"
	of_snmp "github.com/cisco-cx/of/pkg/v2/snmp"
	logger "github.com/cisco-cx/of/wrap/logrus/v2"
	snmp "github.com/cisco-cx/of/wrap/snmp/v2"
	uuid "github.com/cisco-cx/of/wrap/uuid/v2"
	yaml "github.com/cisco-cx/of/wrap/yaml/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

// Enforce AlertGenerator Interface
func TestAlertsInterface(t *testing.T) {
	var _ of_snmp.AlertGenerator = &snmp.Alerter{}
}

// Test multiple Alert firing.
func TestAlertFire(t *testing.T) {
	ag := newAlerter(t)
	for i := 0; i < 10; i++ {
		fireAlert(ag, i+1, t)
	}
}

// Test Alerts firing.
func fireAlert(ag *snmp.Alerter, count int, t *testing.T) {

	// All possible alerts for given configs and trapVars.
	alerts := ag.Alert([]string{"epc"})

	expectedAlert := []of.Alert{
		of.Alert{
			Labels: map[string]string{
				"alert_name":        "starCard",
				"alert_severity":    "error",
				"source_address":    "192.168.1.28",
				"source_hostname":   "localhost",
				"star_slot_num":     "14",
				"subsystem":         "epc",
				"vendor":            "cisco",
				"alert_fingerprint": "5dd1df6eff3119f4",
				"alert_oid":         ".1.3.6.1.4.1.8164.1.2.1.1.1",
			},
			Annotations: map[string]string{
				"event_id":                  "9dcc77fc-dda5-4edf-a683-64f2589036d6",
				"event_oid":                 ".1.3.6.1.4.1.8164.1.2.1.1.1",
				"event_type":                "error",
				"source_address":            "192.168.1.28",
				"source_hostname":           "localhost",
				"event_description":         "",
				"event_filebeat_timestamp":  "2019-04-26T03:46:57.941Z",
				"event_name":                "unknown",
				"event_oid_string":          "",
				"event_generated_time":      "(123) 0:00:01.23",
				"event_rawtext":             "SNMPTRAP timestamp=[2019-04-26T03:46:57Z] hostname=[localhost] address=[UDP/IPv6: [::1]:48381] pdu_security=[TRAP2, SNMP v3, user user-sha-aes128, context ] vars[.1.3.6.1.2.1.1.3.0 = Timeticks: (123) 0:00:01.23\t.1.3.6.1.6.3.1.1.4.1.0 = OID: .1.3.6.1.6.3.1.1.5.1\t.1.3.6.1.4.1.65000.1.1.1.1.1 = STRING: \"foo\"\t.1.3.6.1.4.1.65000.1.1.1.1.1 = STRING: \"bar\"]",
				"event_snmptrapd_timestamp": "2019-04-26T03:46:57Z",
				"event_vars_json":           "[{\"oid\":\".1.3.6.1.6.1.1.1.4.1\",\"oid_string\":\"1.3.6.1.6.1.1.1.4.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.1.1.1.4.1\",\"type\":\"\",\"value\":\".1.3.6.1.4.1.8164.1.2.1.1.1\"},{\"oid\":\".1.3.6.1.4.1.8164.1.2.1.1.1\",\"oid_string\":\"1.3.6.1.4.1.8164.1.2.1.1.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.8164.1.2.1.1.1\",\"type\":\"\",\"value\":\"14\"},{\"oid\":\".1.3.6.1.4.1.24961.2.103.1.1.5.1.2\",\"oid_string\":\"1.3.6.1.4.1.24961.2.103.1.1.5.1.2\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.24961.2.103.1.1.5.1.2\",\"type\":\"\",\"value\":\"package-load-failure\"},{\"oid\":\".1.3.6.1.2.1.1.3.0\",\"oid_string\":\"1.3.6.1.2.1.1.3.0\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.2.1.1.3.0\",\"type\":\"Timeticks\",\"value\":\"(123) 0:00:01.23\"},{\"oid\":\".1.3.6.1.6.3.1.1.4.1\",\"oid_string\":\"1.3.6.1.6.3.1.1.4.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.3.1.1.4.1\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.8164.2.13\"},{\"oid\":\".1.3.6.1.6.3.1.1.4.1.0\",\"oid_string\":\"1.3.6.1.6.3.1.1.4.1.0\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.3.1.1.4.1.0\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.8164.2.44\"},{\"oid\":\".1.3.6.1.4.1.8164.2.44\",\"oid_string\":\"1.3.6.1.4.1.8164.2.44\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.8164.2.44\",\"type\":\"STRING\",\"value\":\"foo\"},{\"oid\":\".1.3.6.1.6.3.1.1.4.1.1\",\"oid_string\":\"1.3.6.1.6.3.1.1.4.1.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.3.1.1.4.1.1\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.8164.2.45\"},{\"oid\":\".1.3.6.1.4.1.8164.2.45\",\"oid_string\":\"1.3.6.1.4.1.8164.2.45\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.8164.2.45\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.65000.1.1.1.1.1\"},{\"oid\":\".1.3.6.1.4.1.65000.1.1.1.1.1\",\"oid_string\":\"1.3.6.1.4.1.65000.1.1.1.1.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.65000.1.1.1.1.1\",\"type\":\"STRING\",\"value\":\"bar\"}]",
			},
			GeneratorURL: "http://www.oid-info.com/get/1.3.6.1.4.1.8164.2.44",
		},
	}

	var err error
	expectedAlert[0].StartsAt, err = time.Parse(time.RFC3339, "2019-04-26T03:46:57Z")
	require.NoError(t, err)

	require.Equal(t, expectedAlert, alerts)
	metrics := promMetrics(t)
	require.Contains(t, metrics, fmt.Sprintf("TestAlertFire_alerts_generated_count{alertType=\"firing\",alert_oid=\".1.3.6.1.4.1.8164.1.2.1.1.1\"} %d", count))
	require.Contains(t, metrics, "TestAlertFire_clearing_alert_count 0")
	require.Contains(t, metrics, fmt.Sprintf("TestAlertFire_unknown_cluster_ip_count 0"))

}

// Test multiple alert clearing.
func TestAlertClear(t *testing.T) {
	ag := newAlerter(t)
	for i := 0; i < 10; i++ {
		clearAlert(ag, i+1, t)
	}
}

// Test configs where device_identifier does not match.
func TestDeviceNotIdentified(t *testing.T) {
	ag := newAlerter(t)
	alerts := ag.Alert([]string{"device_not_found"})
	require.Len(t, alerts, 0)
}

// Test Alerts clearing.
func clearAlert(ag *snmp.Alerter, count int, t *testing.T) {

	// All possible alerts for given configs and trapVars.
	alerts := ag.Alert([]string{"nso"})

	endsAt, err := time.Parse(time.RFC3339, "2019-04-26T03:46:57Z")
	require.NoError(t, err)

	expectedAlertTemplate := of.Alert{
		Labels: map[string]string{
			"alert_severity":  "error",
			"alertname":       "nsoPackageLoadFailure",
			"source_address":  "nso1.example.org",
			"source_hostname": "nso1.example.org",
			"subsystem":       "nso",
			"vendor":          "cisco",
		},
		Annotations: map[string]string{
			"event_id":                  "9dcc77fc-dda5-4edf-a683-64f2589036d6",
			"event_type":                "clear",
			"source_address":            "nso1.example.org",
			"source_hostname":           "nso1.example.org",
			"event_description":         "",
			"event_filebeat_timestamp":  "2019-04-26T03:46:57.941Z",
			"event_name":                "unknown",
			"event_oid_string":          "",
			"event_generated_time":      "(123) 0:00:01.23",
			"event_rawtext":             "SNMPTRAP timestamp=[2019-04-26T03:46:57Z] hostname=[localhost] address=[UDP/IPv6: [::1]:48381] pdu_security=[TRAP2, SNMP v3, user user-sha-aes128, context ] vars[.1.3.6.1.2.1.1.3.0 = Timeticks: (123) 0:00:01.23\t.1.3.6.1.6.3.1.1.4.1.0 = OID: .1.3.6.1.6.3.1.1.5.1\t.1.3.6.1.4.1.65000.1.1.1.1.1 = STRING: \"foo\"\t.1.3.6.1.4.1.65000.1.1.1.1.1 = STRING: \"bar\"]",
			"event_snmptrapd_timestamp": "2019-04-26T03:46:57Z",
			"event_vars_json":           "[{\"oid\":\".1.3.6.1.6.1.1.1.4.1\",\"oid_string\":\"1.3.6.1.6.1.1.1.4.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.1.1.1.4.1\",\"type\":\"\",\"value\":\".1.3.6.1.4.1.8164.1.2.1.1.1\"},{\"oid\":\".1.3.6.1.4.1.8164.1.2.1.1.1\",\"oid_string\":\"1.3.6.1.4.1.8164.1.2.1.1.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.8164.1.2.1.1.1\",\"type\":\"\",\"value\":\"14\"},{\"oid\":\".1.3.6.1.4.1.24961.2.103.1.1.5.1.2\",\"oid_string\":\"1.3.6.1.4.1.24961.2.103.1.1.5.1.2\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.24961.2.103.1.1.5.1.2\",\"type\":\"\",\"value\":\"package-load-failure\"},{\"oid\":\".1.3.6.1.2.1.1.3.0\",\"oid_string\":\"1.3.6.1.2.1.1.3.0\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.2.1.1.3.0\",\"type\":\"Timeticks\",\"value\":\"(123) 0:00:01.23\"},{\"oid\":\".1.3.6.1.6.3.1.1.4.1\",\"oid_string\":\"1.3.6.1.6.3.1.1.4.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.3.1.1.4.1\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.8164.2.13\"},{\"oid\":\".1.3.6.1.6.3.1.1.4.1.0\",\"oid_string\":\"1.3.6.1.6.3.1.1.4.1.0\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.3.1.1.4.1.0\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.8164.2.44\"},{\"oid\":\".1.3.6.1.4.1.8164.2.44\",\"oid_string\":\"1.3.6.1.4.1.8164.2.44\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.8164.2.44\",\"type\":\"STRING\",\"value\":\"foo\"},{\"oid\":\".1.3.6.1.6.3.1.1.4.1.1\",\"oid_string\":\"1.3.6.1.6.3.1.1.4.1.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.6.3.1.1.4.1.1\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.8164.2.45\"},{\"oid\":\".1.3.6.1.4.1.8164.2.45\",\"oid_string\":\"1.3.6.1.4.1.8164.2.45\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.8164.2.45\",\"type\":\"OID\",\"value\":\".1.3.6.1.4.1.65000.1.1.1.1.1\"},{\"oid\":\".1.3.6.1.4.1.65000.1.1.1.1.1\",\"oid_string\":\"1.3.6.1.4.1.65000.1.1.1.1.1\",\"oid_uri\":\"http://www.oid-info.com/get/1.3.6.1.4.1.65000.1.1.1.1.1\",\"type\":\"STRING\",\"value\":\"bar\"}]",
			"event_oid":                 ".1.3.6.1.4.1.8164.2.44",
		},
		StartsAt:     time.Time{},
		EndsAt:       endsAt,
		GeneratorURL: "http://www.oid-info.com/get/1.3.6.1.4.1.8164.2.44",
	}

	expectedAlerts := make([]of.Alert, 3)
	OIDs := []map[string]string{
		map[string]string{
			"oid": ".1.3.6.1.4.1.24961.2.103.2.0.3",
			"fp":  "ef451c2db05fdd5d",
		},
		map[string]string{
			"oid": ".1.3.6.1.4.1.24961.2.103.2.0.4",
			"fp":  "14b93f772f8125e0",
		},
		map[string]string{
			"oid": ".1.3.6.1.4.1.24961.2.103.2.0.5",
			"fp":  "ab89464e06ae596f",
		},
	}

	for idx, val := range OIDs {
		newAlert := of.Alert{}
		alertJSON, _ := json.Marshal(expectedAlertTemplate)
		json.Unmarshal(alertJSON, &newAlert)
		newAlert.Labels["alert_oid"] = val["oid"]
		newAlert.Labels["alert_fingerprint"] = val["fp"]
		expectedAlerts[idx] = newAlert
	}

	// EndsAt is time.Now, so individually matching other components.
	require.Len(t, alerts, 3)
	require.ElementsMatch(t, expectedAlerts, alerts)
	metrics := promMetrics(t)
	for _, val := range OIDs {
		require.Contains(t, metrics, fmt.Sprintf("TestAlertClear_alerts_generated_count{alertType=\"clearing\",alert_oid=\"%s\"} %d", val["oid"], count))
	}
	require.Contains(t, metrics, fmt.Sprintf("TestAlertClear_clearing_alert_count %d", count))
	require.Contains(t, metrics, "TestAlertClear_unknown_cluster_ip_count 0")
}

// Test Unknown logging.
func TestUnknownLogging(t *testing.T) {
	ag := newAlerter(t)

	buf := &bytes.Buffer{}
	l := logger.New()
	l.SetOutput(buf)
	ag.Log = l
	ag.LogUnknown = true

	_ = ag.Unknown("unknown_logging")
	require.Contains(t, string(buf.Bytes()), "SNMPTrapOIDName=oid3 SNMPTrapOIDValue=.1.3.6.1.4.1.8164.2.44")

	ag.LogUnknown = false

	_ = ag.Unknown("unknown_logging")
	require.Contains(t, string(buf.Bytes()), "")
}

// Test Unknown forwarding.
func TestUnknownForwarding(t *testing.T) {
	ag := newAlerter(t)

	ag.ForwardUnknown = true

	alerts := ag.Unknown("unknown_logging")
	require.Len(t, alerts, 1)
	require.Equal(t, "324ad96bef5033e9", alerts[0].Labels[of_snmp.FingerprintText])
	require.Equal(t, "unknownSnmpTrap", alerts[0].Labels["alertname"])
	require.Equal(t, ".1.3.6.1.4.1.8164.2.44", alerts[0].Labels["alert_oid"])

	ag.ForwardUnknown = false

	alerts = ag.Unknown("unknown_logging")
	require.Len(t, alerts, 0)
}

// Test Auto Clear.
func TestAutoClear(t *testing.T) {
	ag := snmp.Alerter{}
	alert := of.Alert{}

	require.Equal(t, alert.EndsAt, time.Time{})

	ag.AutoClear(0, 0, &alert)
	require.Equal(t, alert.EndsAt, time.Time{})

	ag.AutoClear(10, 0, &alert)
	require.Equal(t, alert.EndsAt.Unix(), time.Now().Add(10*time.Minute).Unix())

	ag.AutoClear(0, 20, &alert)
	require.Equal(t, alert.EndsAt.Unix(), time.Now().Add(20*time.Minute).Unix())

	ag.AutoClear(10, 20, &alert)
	require.Equal(t, alert.EndsAt.Unix(), time.Now().Add(20*time.Minute).Unix())

}

// Preparing Alert generator.
func newAlerter(t *testing.T) *snmp.Alerter {
	// Prepare snmp.V2Config
	r := strings.NewReader(YamlContent)
	cfg := yaml.Configs{}
	err := cfg.Decode(r)
	require.NoError(t, err)
	configs := of_snmp.V2Config(cfg)

	//Preparing logger
	l := logger.New()

	mr := mibRegistry(t)

	cntr, cntrVec := snmp.InitCounters(t.Name(), l)
	ag := snmp.Alerter{
		Log:      l,
		Configs:  &configs,
		Receipts: TrapReceipts(),
		Value:    snmp.NewValue(trapVars(), mr),
		MR:       mr,
		U:        &uuid.FixedUUID{},
		Cntr:     cntr,
		CntrVec:  cntrVec,
	}
	return &ag
}

func TestFingerprint(t *testing.T) {
	ag := newAlerter(t)

	expectedAlert := of.Alert{
		Labels: map[string]string{
			"alertname":       "ISE_linkUp",
			"alert_severity":  "info",
			"source_address":  "dead:beef::1",
			"source_hostname": "randomhost",
			"subsystem":       "radware",
			"vendor":          "cisco",
			"alert_oid":       ".1.3.6.1.4.1.1872.2.5.7.0.154",
		},
	}
	require.Equal(t, "5dc17c71ffb8f3e9", ag.Fingerprint(expectedAlert))

	expectedAlert2 := of.Alert{
		Labels: map[string]string{
			"alert_oid":       ".1.3.6.1.4.1.1872.2.5.7.0.154",
			"alertname":       "Alteon_HAGroupNewMasterTrap",
			"alert_severity":  "warning",
			"source_address":  "dead:beef::1",
			"source_hostname": "randomhost",
			"subsystem":       "radware",
			"vendor":          "cisco",
		},
	}
	require.Equal(t, "2fdaaf908f8b8703", ag.Fingerprint(expectedAlert2))

	expectedAlert3 := of.Alert{
		Labels: map[string]string{
			"alert_oid":       ".1.3.6.1.4.1.8164.2.22",
			"alert_severity":  "critical",
			"alertname":       "starCardTempOverheat",
			"source_address":  "192.168.1.28",
			"source_hostname": "localhost",
		},
	}
	require.Equal(t, "a046e8b60b11f977", ag.Fingerprint(expectedAlert3))

	expectedAlert3.Labels["alert_fingerprint"] = "a046e8b60b11f977"
	require.Equal(t, "a046e8b60b11f977", ag.Fingerprint(expectedAlert3))
}

// Fetches current metrics.
func promMetrics(t *testing.T) string {
	ts := httptest.NewServer(promhttp.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	defer res.Body.Close()
	metrics, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	return string(metrics)
}
