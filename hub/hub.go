package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"

	"strings"
	"sync"
)

type handlerFunc func(*bufio.ReadWriter)

type endpoint struct {
	listener       net.Listener
	connectedUsers map[uint64]*bufio.ReadWriter
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

	userID := e.connectUser(rw)
	e.routeMessage(userID)
}

func (e *endpoint) connectUser(rw *bufio.ReadWriter) (userID uint64) {

	userID = e.nextID

	e.m.RLock()
	e.connectedUsers[userID] = rw
	e.m.RUnlock()

	e.nextID++

	return userID

}

func (e *endpoint) handleIdentity(userID uint64) {
	log.Println("Handling identity message")

	e.m.RLock()
	rw := e.connectedUsers[userID]
	e.m.RUnlock()

	response := fmt.Sprintf("Your identity is: %d\n", userID)

	_, err := rw.WriteString(response)
	if err != nil {
		log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
}

func (e *endpoint) handleList(userID uint64) {
	log.Println("Handling list message")

	connectedUsers := []uint64{}

	e.m.RLock()
	rw := e.connectedUsers[userID]

	for user := range e.connectedUsers {
		if user == userID {
			continue
		}
		connectedUsers = append(connectedUsers, user)
	}
	e.m.RUnlock()

	sort.Slice(connectedUsers, func(i, j int) bool { return connectedUsers[i] < connectedUsers[j] })

	response := ""
	if len(connectedUsers) < 1 {
		response = "no other users are connected"
	} else {
		response = fmt.Sprintf("The following user(s) are connected %d\n", connectedUsers)
	}

	_, err := rw.WriteString(response)
	if err != nil {
		log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
}

func (e *endpoint) handleRelay(senderID uint64) {
	log.Println("Handling relay message")

	e.m.RLock()
	rw := e.connectedUsers[senderID]
	e.m.RUnlock()

	message, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Cannot read from connection.\n", err)
	}
	message = strings.Trim(message, "\n ")
	log.Printf("message received from client: %s\n", message)

	to, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Cannot read from connection.\n", err)
	}
	to = strings.Trim(to, "\n ")
	log.Printf("to IDs received from client: %s\n", to)

	recipients := []uint64{}

	ids := strings.Split(to, ",")
	for _, id := range ids {
		userID, err := strconv.ParseUint(id, 10, 64)

		if err != nil {
			log.Println(err)
		}
		recipients = append(recipients, userID)
	}

	if len(recipients) > 255 {
		log.Println("too many message recipients, max 255")
		return
	}

	for _, userID := range recipients {

		e.m.RLock()
		if e.connectedUsers[userID] == nil {
			log.Printf("user %d is not connected", userID)
			continue
		}
		rw := e.connectedUsers[userID]
		e.m.RUnlock()

		response := fmt.Sprintf("Message from %d : %s\n", senderID, message)

		_, err := rw.WriteString(response)
		if err != nil {
			log.Println("Cannot write to connection.\n", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}
	}

}

func (e *endpoint) routeMessage(userID uint64) {

	TestMode = os.Getenv("TEST_MODE")

	e.m.RLock()
	rw := e.connectedUsers[userID]
	e.m.RUnlock()

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
			e.handleIdentity(userID)
		case "LIST":
			e.handleList(userID)
		case "RELAY":
			e.handleRelay(userID)
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
		connectedUsers: make(map[uint64]*bufio.ReadWriter),
		nextID:         uint64(1),
	}
}

func startHub() error {

	Port = ":61000"

	endpoint := newEndpoint()

	return endpoint.listen(Port)
}
