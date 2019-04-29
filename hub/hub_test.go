package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateNewEndpoint(t *testing.T) {

	endpoint := newEndpoint()

	assert.Nil(t, endpoint.listener)
	assert.NotNil(t, endpoint.handler)
	assert.NotNil(t, endpoint.handler["IDENTITY"])
}

func TestListen(t *testing.T) {

	

	// to do need to mock out the port number

	// ep := newEndpoint()
	// err := ep.listen()
	// assert.Equal(t, nil, err)

}
