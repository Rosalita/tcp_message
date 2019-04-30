package mockhub

import (
	"log"
	"net"
)

type Endpoint struct {
	listener net.Listener
}

var (
	Port string
)

// Listen starts listening for tcp messages
func (e *Endpoint) Listen(port string) error {

	var err error
	e.listener, err = net.Listen("tcp", port)
	if err != nil {
		return err
	}

	log.Println("Mock hub is listening on", e.listener.Addr().String())

	return nil
}


// NewEndpoint creates a new endpoint
func NewEndpoint() *Endpoint {
	return &Endpoint{}
}
