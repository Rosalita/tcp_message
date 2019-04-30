package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/Rosalita/tcp_message/mockhub"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	flag.Parse()
	// Test setup

	// Create a hub for these client tests to interact with
	createHubForTests()

	// Run tests
	exitCode := m.Run()

	// Test teardown
	os.Exit(exitCode)
}

func createHubForTests() {
	Port = ":61000"
	endpoint := mockhub.NewEndpoint()
	go endpoint.Listen(Port)

}

func TestOpenConnection(t *testing.T) {

	validAddress := "localhost:61000"
	missingPort := "localhost"
	badPort := "localhost:1"

	tests := []struct {
		address string
		errMsg  string
	}{
		{validAddress, ""},
		{missingPort, "dial tcp: address localhost: missing port in address"},
		{badPort, "dial tcp 127.0.0.1:1: connect: connection refused"},
	}

	for _, test := range tests {

		_, err := openConnection(test.address)

		if test.errMsg == "" {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, test.errMsg, err.Error())
		}
	}
}

func TestSendMessageToHub(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	tests := []struct {
		cmd      string
		m        string
		to       string
		expected string
	}{
		{"IDENTITY", "", "", "IDENTITY\n"},
		{"LIST", "", "", "LIST\n"},
		{"RELAY", "message", "1,2", "RELAY\nmessage\n1,2\n"},
		{"", "", "", "\n"},
	}

	for _, test := range tests {
		sendMessageToHub(rw, test.cmd, test.m, test.to)
		result := buffer.String()
		assert.Equal(t, test.expected, result)
		buffer.Reset()
	}
}

func TestReadResponseFromHub(t *testing.T) {

	var buffer bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))

	validResponse := "This is a valid response from hub \n"
	missingNewLine := "This response doesn't end in a new line"
	blankResponse := ""

	tests := []struct {
		hubResponse string
		err         error
		expected    string
	}{
		{validResponse, nil, "This is a valid response from hub \n"},
		{missingNewLine, errors.New("EOF"), ""},
		{blankResponse, errors.New("EOF"), ""},
	}

	for _, test := range tests {

		rw.WriteString(test.hubResponse)
		rw.Flush()

		result, err := readResponseFromHub(rw)

		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)

		buffer.Reset()
	}
}
