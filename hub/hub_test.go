package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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

		ep.connectedUsers[uint64(1)] = rw

		ep.routeMessage(uint64(1))
		assert.Contains(t, logged.String(), test.logged)
		logged.Reset()

	}

	UnsetTestMode()
}

func TestConnectUser(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	ep := newEndpoint()

	userID := ep.connectUser(rw)

	assert.Equal(t, uint64(1), userID)
	assert.Equal(t, 1, len(ep.connectedUsers))
	assert.NotNil(t, ep.connectedUsers[uint64(1)])
	assert.Equal(t, uint64(2), ep.nextID)

}

func TestHandleIdentity(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	ep := newEndpoint()

	ep.connectedUsers[uint64(1)] = rw

	tests := []struct {
		userID   uint64
		response string
	}{
		{uint64(1), "Your identity is: 1\n"},
		{uint64(2), "Your identity is: 2\n"},
		{uint64(3), "Your identity is: 3\n"},
	}

	for _, test := range tests {
		ep.connectedUsers[test.userID] = rw

		ep.handleIdentity(test.userID)
		result := buffer.String()
		assert.Equal(t, test.response, result)

		buffer.Reset()
	}

}

func TestHandleList(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	epOneConnection := newEndpoint()
	epOneConnection.connectUser(rw)

	epTwoConnections := newEndpoint()
	epTwoConnections.connectUser(rw)
	epTwoConnections.connectUser(rw)

	epThreeConnections := newEndpoint()
	epThreeConnections.connectUser(rw)
	epThreeConnections.connectUser(rw)
	epThreeConnections.connectUser(rw)

	tests := []struct {
		endpoint *endpoint
		userID   uint64
		response string
	}{
		{epOneConnection, uint64(1), "no other users are connected"},
		{epTwoConnections, uint64(1), "The following user(s) are connected [2]\n"},
		{epThreeConnections, uint64(2), "The following user(s) are connected [1 3]\n"},
	}

	for _, test := range tests {

		test.endpoint.handleList(test.userID)

		result := buffer.String()
		assert.Equal(t, test.response, result)

		buffer.Reset()
	}
}

func TestHandleRelay(t *testing.T) {

	var buffer1 bytes.Buffer
	rw1 := bufio.NewReadWriter(bufio.NewReader(&buffer1), bufio.NewWriter(&buffer1))

	var buffer2 bytes.Buffer
	rw2 := bufio.NewReadWriter(bufio.NewReader(&buffer2), bufio.NewWriter(&buffer2))

	var buffer3 bytes.Buffer
	rw3 := bufio.NewReadWriter(bufio.NewReader(&buffer3), bufio.NewWriter(&buffer3))

	epThreeConnections := newEndpoint()
	epThreeConnections.connectUser(rw1)
	epThreeConnections.connectUser(rw2)
	epThreeConnections.connectUser(rw3)

	tooManyRecipients := ""
	for i := 1; i < 257; i++ {
		tooManyRecipients += strconv.Itoa(i) + ","
	}

	tests := []struct {
		endpoint *endpoint
		senderID uint64
		message  string
		to       string
		response string
	}{
		{epThreeConnections, uint64(3), "foo message", "1,2", "Message from 3 : foo message\n"},
		{epThreeConnections, uint64(3), "bar message", tooManyRecipients, ""},
	}

	for _, test := range tests {

		senderRw := test.endpoint.connectedUsers[test.senderID]
		request := fmt.Sprintf("%s\n%s\n", test.message, test.to)
		senderRw.WriteString(request)
		senderRw.Flush()

		test.endpoint.handleRelay(test.senderID)

		user1result := buffer1.String()
		assert.Equal(t, test.response, user1result)

		user2result := buffer2.String()
		assert.Equal(t, test.response, user2result)

		user3result := buffer3.String()
		assert.Equal(t, "", user3result)

		buffer1.Reset()
		buffer2.Reset()
		buffer3.Reset()
	}

}

// Test helper functions
func SetTestMode() {
	os.Setenv("TEST_MODE", "on")
}

func UnsetTestMode() {
	os.Unsetenv("TEST_MODE")
}
