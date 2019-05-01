package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

// Exported global variables used for mocking values in unit tests
var (
	// Port is the port number the hub will listen on
	Port string
)

func startClient(ip, cmd, m, to string) error {

	messageSent := false

	rw, err := openConnection(ip + Port)
	if err != nil {
		return err
	}

	for {

		if messageSent == false {
			err = sendMessageToHub(rw, cmd, m, to)

			_, err := readResponseFromHub(rw)
			if err != nil {
				return err
			}

			messageSent = true
		}

		readResponseFromHub(rw)
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

func sendMessageToHub(rw *bufio.ReadWriter, cmd, m, to string) error {

	_, err := rw.WriteString(cmd + "\n")
	if err != nil {
		return err
	}
	if m != "" {
		_, err = rw.WriteString(m + "\n")
		if err != nil {
			return err
		}
	}

	if to != "" {
		_, err = rw.WriteString(to + "\n")
		if err != nil {
			return err
		}
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

	if response != "" {
		log.Println("response from hub:", response)
	}

	return response, nil
}

func main() {

	Port = ":61000"

	cmd := flag.String("cmd", "", "The command client will perform, can be IDENTITY, LIST or RELAY.")
	m := flag.String("m", "", "Message to relay, required param for RELAY")
	to := flag.String("to", "", "comma separated string of IDs (1,2,3) to relay message to, required param for RELAY")
	flag.Parse()

	*cmd = strings.ToUpper(*cmd)

	err := startClient("localhost", *cmd, *m, *to)
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("Client done.")
	return

}
