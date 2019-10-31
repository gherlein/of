package v2_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	of "github.com/cisco-cx/of/pkg/v2"
	http "github.com/cisco-cx/of/wrap/http/v2"
	snmp "github.com/cisco-cx/of/wrap/snmp/v2"
)

// Test Handler
func TestHandlerRun(t *testing.T) {

	// AM server
	addr := "localhost:14932"

	cfg := of.SNMPConfig{
		AMAddress:     "http://" + addr,
		AMTimeout:     5 * time.Second,
		ListenAddress: "localhost:44932",
		Version:       "Handler Test",
	}

	// Start fake AM Server.
	hc := &of.HTTPConfig{ListenAddress: addr, ReadTimeout: cfg.AMTimeout, WriteTimeout: cfg.AMTimeout}

	srv := http.NewServer(hc)
	srv.HandleFunc("/-/healthy", func(w of.ResponseWriter, r of.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	go func() {
		err := srv.ListenAndServe()
		require.NoError(t, err)
	}()
	defer srv.Shutdown()
	time.Sleep(time.Second)
	service := initService(t)
	h := snmp.Handler{
		SNMP:   service,
		Log:    service.Log,
		Config: &cfg,
	}

	go h.Run()
	time.Sleep(2 * time.Second)

	// Test Version
	checkResponse(t, 200, "http://"+cfg.ListenAddress, "Handler Test")

	// Test health check
	checkResponse(t, 500, "http://"+cfg.ListenAddress+"/health", "Received status code '502' does not match expected status code '200'")

	// Test metrics
	message := getResponse(t, 200, "http://"+cfg.ListenAddress+"/metrics")
	require.Contains(t, message, "promhttp_metric_handler_requests_total")

	// Test Status
	checkResponse(t, 200, "http://"+cfg.ListenAddress+"/api/v2/status", "{ AlertManager Client for SNMP Traps {https://github.com/cisco-cx/am-client-snmp} success}")

	// Test Posting SNMP event
	dataBytes, err := json.Marshal(TrapEvents())
	require.NoError(t, err)

	data := bytes.NewBuffer(dataBytes)
	c := http.NewClient()
	req, err := http.NewRequest("Post", "http://"+cfg.ListenAddress+"/api/v2/events", data)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "test")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

// HTTP client to hit server and check response.
func checkResponse(t *testing.T, expectedStatusCode int, u, msg string) {
	all := getResponse(t, expectedStatusCode, u)
	require.Equal(t, msg, all)
}

// HTTP client to hit server and check response.
func getResponse(t *testing.T, expectedStatusCode int, u string) string {

	c := http.NewClient()
	res, err := c.Get(u)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, res.StatusCode)

	all, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	return string(all)
}