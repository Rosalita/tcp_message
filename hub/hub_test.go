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

func TestNewEndpoint(t *testing.T) {

	endpoint := newEndpoint()

	assert.Nil(t, endpoint.listener)
	assert.Equal(t, endpoint.nextID, uint64(1))
}

func TestListen(t *testing.T) {

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
		msgType string
		logged  string
	}{
		{validMessageType, "Received message type IDENTITY"},
		{unknownMessageType, "Received message type UNKNOWN"},
		{missingMessageType, "Error reading message type"},
	}

	for _, test := range tests {
		rw.WriteString(test.msgType)
		rw.Flush()

		ep.routeMessage(rw)
		assert.Contains(t, logged.String(), test.logged)
		logged.Reset()

	}

	UnsetTestMode()
}

func TestHandleIdentity(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	ep := newEndpoint()

	ep.handleIdentity(rw)

	assert.Equal(t, 1, len(ep.connectedUsers))
	assert.Equal(t, uint64(1), ep.connectedUsers[0])
	assert.Equal(t, uint64(2), ep.nextID)

	result := buffer.String()
	assert.Equal(t, "1\n", result)

	buffer.Reset()
	ep.handleIdentity(rw)

	assert.Equal(t, 2, len(ep.connectedUsers))
	assert.Equal(t, uint64(1), ep.connectedUsers[0])
	assert.Equal(t, uint64(2), ep.connectedUsers[1])
	assert.Equal(t, uint64(3), ep.nextID)

	result = buffer.String()
	assert.Equal(t, "2\n", result)

}

func TestHandleList(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	ep := newEndpoint()
	ep.handleList(rw)

	result := buffer.String()
	assert.Equal(t, "[]\n", result)

	buffer.Reset()
	ep.handleIdentity(rw)
	buffer.Reset()
	ep.handleList(rw)

	result = buffer.String()
	assert.Equal(t, "[1]\n", result)

	buffer.Reset()
	ep.handleIdentity(rw)
	buffer.Reset()
	ep.handleList(rw)

	result = buffer.String()
	assert.Equal(t, "[1 2]\n", result)
}

func TestHandleRelay(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	ep := newEndpoint()
	ep.handleRelay(rw)

	result := buffer.String()
	assert.Equal(t, "", result)

}

// Test helper functions
func SetTestMode() {
	os.Setenv("TEST_MODE", "on")
}

func UnsetTestMode() {
	os.Unsetenv("TEST_MODE")
}
