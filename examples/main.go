package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/joseluisq/gonetc"
)

func main() {
	// Code example using IPC Unix Sockets

	// 1. Create a simple listening Unix socket with echo functionality
	// using the `socat` tool -> http://www.dest-unreach.org/socat/
	// Then execute the following commands on your terminal:
	//	rm -f /tmp/mysocket && socat UNIX-LISTEN:/tmp/mysocket,fork exec:'/bin/cat'

	// 2. Now just run this client code example in order to exchange data with current socket.
	//	go run examples/main.go

	// 2.1 Connect to the listening socket

	fmt.Println("=== First example")

	sock := gonetc.New("unix", "/tmp/mysocket")
	err := sock.Connect()
	if err != nil {
		log.Fatalln("unable to communicate with socket:", err)
	}

	// 2.2 Send some sequential data to current socket (example only)
	pangram := strings.Split("The quick brown fox jumps over the lazy dog", " ")
	for _, word := range pangram {
		log.Println("client data sent:", word)
		_, err := sock.Write([]byte(word), func(resp []byte, err error, done func()) {
			if err != nil {
				log.Fatalln("unable to write data to socket", err)
			}
			log.Println("client data received:", string(resp))
			done()
		})
		if err != nil {
			log.Fatalln("unable to write to socket:", err)
		}
	}

	sock.Close()

	fmt.Println("=== Second example (reconnection)")

	err = sock.Connect()
	if err != nil {
		log.Fatalln("unable to communicate with socket:", err)
	}

	// 2.2 Send some sequential data to current socket (example only)
	pangram = strings.Split("The quick brown fox jumps over the lazy dog", " ")
	for _, word := range pangram {
		log.Println("client data sent:", word)
		_, err := sock.Write([]byte(word), func(resp []byte, err error, done func()) {
			if err != nil {
				log.Fatalln("unable to write data to socket", err)
			}
			log.Println("client data received:", string(resp))
			done()
		})
		if err != nil {
			log.Fatalln("unable to write to socket:", err)
		}
	}

	sock.Close()
}
