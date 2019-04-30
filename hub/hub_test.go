package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	flag.Parse()
	// Test setup
	Port = ":61000"

	// Run tests
	exitCode := m.Run()

	// Test teardown
	os.Exit(exitCode)
}

func TestCanCreateNewEndpoint(t *testing.T) {

	endpoint := newEndpoint()

	assert.Nil(t, endpoint.listener)
	assert.NotNil(t, endpoint.handler)
	assert.NotNil(t, endpoint.handler["IDENTITY"])
}

func TestHubCanListen(t *testing.T) {

	SetTestMode()

	var logged bytes.Buffer
	log.SetOutput(&logged)

	ep := newEndpoint()
	err := ep.listen(Port)

	assert.Equal(t, nil, err)
	assert.Contains(t, logged.String(), "Listening on [::]:61000")

	UnsetTestMode()
}

func TestRouteMessage(t *testing.T) {

	SetTestMode()

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	var logged bytes.Buffer
	log.SetOutput(&logged)

	ep := newEndpoint()

	validMessageType := "IDENTITY\n"
	unknownMessageType := "UNKNOWN\n"
	missingMessageType := ""

	tests := []struct {
		msgType     string
		loggedLine1 string
		loggedLine2 string
	}{
		{validMessageType, "Received message type IDENTITY", ""},
		{unknownMessageType, "Received message type UNKNOWN", "Message Type 'UNKNOWN' is not recognised."},
		{missingMessageType, "EOF", ""},
	}

	for _, test := range tests {
		rw.WriteString(test.msgType)
		rw.Flush()

		ep.routeMessage(rw)

		assert.Contains(t, logged.String(), test.loggedLine1)
		assert.Contains(t, logged.String(), test.loggedLine2)

		logged.Reset()

	}

	UnsetTestMode()

}

// Test helper functions
func SetTestMode() {
	os.Setenv("TEST_MODE", "on")
}

func UnsetTestMode() {
	os.Unsetenv("TEST_MODE")
}
