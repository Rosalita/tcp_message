package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)

// Exported global variables used for mocking values in unit tests
var (
	// Port is the port number the hub will listen on
	Port string
)

func startClient(ip, cmd string) error {

	messageSent := false

	rw, err := openConnection(ip + Port)
	if err != nil {
		return err
	}

	for {

		if messageSent == false {
			err = sendMessageToHub(rw, cmd, "Hello hub")

			_, err := readResponseFromHub(rw)
			if err != nil {
				return err
			}

			messageSent = true
		}

	}
}

func openConnection(address string) (*bufio.ReadWriter, error) {
	log.Println("Connecting to hub " + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

func sendMessageToHub(rw *bufio.ReadWriter, msgType, msg string) error {

	_, err := rw.WriteString(msgType + "\n")
	if err != nil {
		return err
	}
	_, err = rw.WriteString(msg + "\n")
	if err != nil {
		return err
	}

	log.Println("Sending message to the hub.")
	err = rw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func readResponseFromHub(rw *bufio.ReadWriter) (string, error) {
	response, err := rw.ReadString('\n')
	if err != nil {
		return "", err
	}

	log.Println("response from hub:", response)
	return response, nil
}

func main() {

	Port = ":61000"

	cmd := flag.String("cmd", "", "The command client will perform, IDENTITY, LIST or RELAY.")
	flag.Parse()

	err := startClient("localhost", *cmd)
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("Client done.")
	return

}
