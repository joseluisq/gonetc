// Package gonetc is a simple client interface that wraps Go net.
package gonetc

import (
	"fmt"
	"net"
)

// NetClient defines a client network.
type NetClient struct {
	network string
	address string
	conn    net.Conn

	// It specifies the maximum size of bytes per read (2048 by default).
	MaxReadBytes int
}

// New creates a new client network instance. Parameters are the same as Go `net.Dial`.
func New(network string, address string) *NetClient {
	return &NetClient{
		network:      network,
		address:      address,
		MaxReadBytes: 2048,
	}
}

// Connect establishes a new network connection.
func (c *NetClient) Connect() error {
	conn, err := net.Dial(c.network, c.address)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// readData reads data from current connection.
func (c *NetClient) readData(respHandler func(data []byte, err error, done func())) {
	if c.conn == nil {
		respHandler(make([]byte, 0), fmt.Errorf("no network connection to read"), func() {})
		return
	}
	if c.MaxReadBytes < 0 {
		respHandler(make([]byte, 0), fmt.Errorf("max read bytes can not be negative"), func() {})
		return
	}
	var quit = make(chan struct{})
	var buf = make([]byte, c.MaxReadBytes)
	for {
		select {
		case <-quit:
			return
		default:
			n, err := c.conn.Read(buf)
			respHandler(buf[:n], err, func() {
				close(quit)
			})
		}
	}
}

// Conn gets the inner Go net connection instance.
func (c *NetClient) Conn() net.Conn {
	return c.conn
}

// Listen listens for incoming response data.
func (c *NetClient) Listen(respHandler func(data []byte, err error, done func())) {
	c.readData(respHandler)
}

// Write writes bytes to current client network connection. It also provides an optional data response handler.
// When a `respHandler` function is provided then three params are provided: `data []byte`, `err error`, `done func()`.
// The `done()` function param acts as a callback completion in order to finish the current write execution.
func (c *NetClient) Write(data []byte, respHandler func(data []byte, err error, done func())) (n int, err error) {
	if c.conn == nil {
		return 0, fmt.Errorf("no network connection to write")
	}
	n, err = c.conn.Write(data)
	if err == nil && respHandler != nil {
		c.Listen(respHandler)
	}
	return n, err
}

// Close closes current client network connection.
func (c *NetClient) Close() error {
	if c.conn == nil {
		return fmt.Errorf("no network connection to close")
	}
	return c.conn.Close()
}
