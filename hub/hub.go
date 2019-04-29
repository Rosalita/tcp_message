package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type handlerFunc func(*bufio.ReadWriter)

type endpoint struct {
	listener net.Listener
	handler  map[string]handlerFunc
	m        sync.RWMutex
}

const (
	port = ":61000"
)


func (e *endpoint) listen() error {
	var err error
	e.listener, err = net.Listen("tcp", port)
	if err != nil {
		return err
	}
	log.Println("Listening on", e.listener.Addr().String())
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
	defer conn.Close()

	for {

		msgType, err := rw.ReadString('\n')
		switch {
		case err == io.EOF:
			log.Println("Reached EOF - closing connection.\n   ---")
			return
		case err != nil:
			log.Println("\nError reading command. Got: '"+msgType+"'\n", err)
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
		handleCommand(rw)
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

	endpoint := newEndpoint()

	return endpoint.listen()
}
