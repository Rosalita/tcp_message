package main

import (
	"bufio"
	"log"
	"net"
	"os"

	"strings"
	"sync"
)

type handlerFunc func(*bufio.ReadWriter)

type endpoint struct {
	listener net.Listener
	handler  map[string]handlerFunc
	m        sync.RWMutex
}

// Exported global variables used for mocking values in unit tests
var (
	// Port is the port number the hub will listen on
	Port string

	// TestMode is env var that can be set to "on" it is used by unit tests
	TestMode string
)

func (e *endpoint) listen(port string) error {

	// This function is test mode aware
	TestMode = os.Getenv("TEST_MODE")

	var err error
	e.listener, err = net.Listen("tcp", port)
	if err != nil {
		return err
	}

	log.Println("Listening on", e.listener.Addr().String())

	// Test mode returns after checking listening is ok
	if TestMode == "on" {
		return nil
	}

	for {
		conn, err := e.listener.Accept()
		if err != nil {
			log.Println("Failed to accept a connection request:", err)
			continue
		}

		go e.handleMessages(conn)
	}
}

func (e *endpoint) handleMessages(conn net.Conn) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	e.routeMessage(rw)
}

func (e *endpoint) routeMessage(rw *bufio.ReadWriter) {

	// This function is test mode aware
	TestMode = os.Getenv("TEST_MODE")

	msgType, err := rw.ReadString('\n')

	switch {
	case err != nil:
		log.Println("\nError reading message type. Got: '"+msgType+"'\n", err)
		return
	}

	msgType = strings.Trim(msgType, "\n ")
	log.Printf("Received message type %s\n", msgType)

	e.m.RLock()
	handleCommand, ok := e.handler[msgType]
	e.m.RUnlock()
	if !ok {
		log.Println("Message Type '" + msgType + "' is not recognised.")
		return
	}

	// only handle commands outside of test mode
	if TestMode != "on" {
		handleCommand(rw)
		return
	}

}

func main() {

	err := startHub()
	if err != nil {
		log.Println("Error:", err)
	}

	log.Println("Server done.")
}

func newEndpoint() *endpoint {
	return &endpoint{
		handler: map[string]handlerFunc{
			"IDENTITY": handleIdentityMessage,
		},
	}
}

func handleIdentityMessage(rw *bufio.ReadWriter) {

	log.Println("I'm handling an identity message")

	s, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Cannot read from connection.\n", err)
	}
	s = strings.Trim(s, "\n ")
	log.Printf("data received from client: %s\n", s)

	_, err = rw.WriteString("Thank you for connecting to me.\n")
	if err != nil {
		log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
}

func startHub() error {

	Port = ":61000"

	endpoint := newEndpoint()

	return endpoint.listen(Port)
}
