package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"strings"
	"sync"
)

type handlerFunc func(*bufio.ReadWriter)

type endpoint struct {
	listener       net.Listener
	handler        map[string]handlerFunc
	connectedUsers []uint64
	nextID         uint64
	m              sync.RWMutex
}

// Exported global variables used for mocking values in unit tests
var (
	// Port is the port number the hub will listen on
	Port string

	// TestMode is env var that can be set to "on" it is used by unit tests
	TestMode string
)

func (e *endpoint) listen(port string) error {

	TestMode = os.Getenv("TEST_MODE")

	var err error
	e.listener, err = net.Listen("tcp", port)
	if err != nil {
		return err
	}

	log.Println("Listening on", e.listener.Addr().String())

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

func (e *endpoint) handleIdentity(rw *bufio.ReadWriter) {
	log.Println("Handling identity message")

	userID := e.nextID
	e.connectedUsers = append(e.connectedUsers, userID)
	e.nextID++

	response := fmt.Sprintf("%d\n", userID)

	_, err := rw.WriteString(response)
	if err != nil {
		log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
}

func (e *endpoint) handleList(rw *bufio.ReadWriter) {
	log.Println("Handling list message")

	response := fmt.Sprintf("%d\n", e.connectedUsers)

	_, err := rw.WriteString(response)
	if err != nil {
		log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
}

func (e *endpoint) handleRelay(rw *bufio.ReadWriter) {
	log.Println("Handling relay message")

	m, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Cannot read from connection.\n", err)
	}
	m = strings.Trim(m, "\n ")
	log.Printf("message received from client: %s\n", m)

	to, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Cannot read from connection.\n", err)
	}
	to = strings.Trim(to, "\n ")
	log.Printf("to IDs received from client: %s\n", to)

}

func (e *endpoint) routeMessage(rw *bufio.ReadWriter) {

	TestMode = os.Getenv("TEST_MODE")

	msgType, err := rw.ReadString('\n')

	if err != nil {
		log.Println("\nError reading message type. Got: '"+msgType+"'\n", err)
		return
	}

	msgType = strings.Trim(msgType, "\n ")
	log.Printf("Received message type %s\n", msgType)

	if TestMode != "on" {
		switch msgType {
		case "IDENTITY":
			e.handleIdentity(rw)
		case "LIST":
			e.handleList(rw)
		case "RELAY":
			e.handleRelay(rw)
		}
	}

	// e.m.RLock()
	// handleCommand, ok := e.handler[msgType]
	// e.m.RUnlock()

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
		nextID: uint64(1),
	}
}

func startHub() error {

	Port = ":61000"

	endpoint := newEndpoint()

	return endpoint.listen(Port)
}
