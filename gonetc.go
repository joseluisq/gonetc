// Package gonetc is a simple client interface that wraps Go net.
package gonetc

import (
	"fmt"
	"io"
	"net"
)

// NetClient defines a client network.
type NetClient struct {
	network string
	address string
	netResp chan netResp
	// It holds the current Go net inner connection instance.
	Conn net.Conn
}

// netResp defines the client network response pair.
type netResp struct {
	data []byte
	err  error
}

// netReader reads network response data.
func netReader(r io.Reader, resp chan<- netResp) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			resp <- netResp{
				data: make([]byte, 0),
				err:  err,
			}
			return
		}
		resp <- netResp{
			data: buf[0:n],
			err:  err,
		}
	}
}

// New creates a new client network instance. Parameters are the same as Go `net.Dial`.
func New(network string, address string) *NetClient {
	return &NetClient{
		network: network,
		address: address,
	}
}

// Connect establishes a new network connection.
func (c *NetClient) Connect() error {
	conn, err := net.Dial(c.network, c.address)
	if err != nil {
		return err
	}
	resp := make(chan netResp)
	go netReader(conn, resp)
	c.Conn = conn
	c.netResp = resp
	return nil
}

// Write writes bytes to current client network connection. It also provides an optional data response handler.
// When a `respHandler` function is provided then three params are provided: `data []byte`, `err error`, `done func()`.
// The `done()` function param acts as a callback completion in order to finish the current write execution.
func (c *NetClient) Write(data []byte, respHandler func(data []byte, err error, done func())) (n int, err error) {
	if c.Conn == nil {
		return 0, fmt.Errorf("no available unix network connection to write")
	}
	n, err = c.Conn.Write(data)
	if err == nil && respHandler != nil {
		var res netResp
		quit := make(chan struct{})
	loop:
		for {
			select {
			case <-quit:
				break loop
			case res = <-c.netResp:
				respHandler(res.data, res.err, func() {
					close(quit)
				})
			}
		}
	}
	return n, err
}

// Close closes current client network connection.
func (c *NetClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
