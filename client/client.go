package main

import (
	"bufio"
	"flag"
	"log"
	"net"
)

func main() {

	msgType := flag.String("msgType", "", "The message type to send, IDENTITY, LIST or RELAY.")
	flag.Parse()

	err := startClient("localhost", *msgType)
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("Client done.")
	return

}

const (
	port = ":61000"
)

func startClient(ip, messageType string) error {

	rw, err := openConnection(ip + port)
	if err != nil {
		return err
	}

	_, err = rw.WriteString(messageType + "\n")
	if err != nil {
		return err
	}
	_, err = rw.WriteString("Additional data.\n")
	if err != nil {
		return err
	}

	log.Println("Sending message to the hub.")
	err = rw.Flush()
	if err != nil {
		return err
	}

	response, err := rw.ReadString('\n')
	if err != nil {
		return err
	}

	log.Println("reply from hub:", response)

	return nil
}

func openConnection(address string) (*bufio.ReadWriter, error) {
	log.Println("Connecting to hub " + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}
