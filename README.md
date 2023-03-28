# gonetc ![devel](https://github.com/joseluisq/gonetc/workflows/devel/badge.svg) [![codecov](https://codecov.io/gh/joseluisq/gonetc/branch/master/graph/badge.svg)](https://codecov.io/gh/joseluisq/gonetc) [![Go Report Card](https://goreportcard.com/badge/github.com/joseluisq/gonetc)](https://goreportcard.com/report/github.com/joseluisq/gonetc) [![PkgGoDev](https://pkg.go.dev/badge/github.com/joseluisq/gonetc)](https://pkg.go.dev/github.com/joseluisq/gonetc)

> A simple [Go Network](https://golang.org/pkg/net/) wrapper client interface.

## Usage

```go
package main

import (
    "log"
    "strings"

    "github.com/joseluisq/gonetc"
)

func main() {
    // Code example using IPC Unix Sockets

    // 1. Create a simple listening Unix socket with echo functionality
    // using the `socat` tool -> http://www.dest-unreach.org/socat/
    // Then execute the following commands on your terminal:
    //  rm -f /tmp/mysocket && socat UNIX-LISTEN:/tmp/mysocket,fork exec:'/bin/cat'

    // 2. Now just run this client code example in order to exchange data with current socket.
    //  go run examples/main.go

    // 2.1 Connect to the listening socket
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
            // Finish the current write handling response if we are done
            done()
        })
        if err != nil {
            log.Fatalln("unable to write to socket:", err)
        }
    }

    sock.Close()

    // 3. Finally after running the client you'll see a similar output like:
    //
    // 2020/12/11 00:32:27 client data sent: The
    // 2020/12/11 00:32:27 client data received: The
    // 2020/12/11 00:32:28 client data sent: quick
    // 2020/12/11 00:32:28 client data received: quick
    // 2020/12/11 00:32:29 client data sent: brown
    // 2020/12/11 00:32:29 client data received: brown
    // 2020/12/11 00:32:30 client data sent: fox
    // 2020/12/11 00:32:30 client data received: fox
    // 2020/12/11 00:32:31 client data sent: jumps
    // 2020/12/11 00:32:31 client data received: jumps
    // 2020/12/11 00:32:32 client data sent: over
    // 2020/12/11 00:32:32 client data received: over
    // 2020/12/11 00:32:33 client data sent: the
    // 2020/12/11 00:32:33 client data received: the
    // 2020/12/11 00:32:34 client data sent: lazy
    // 2020/12/11 00:32:34 client data received: lazy
    // 2020/12/11 00:32:35 client data sent: dog
    // 2020/12/11 00:32:35 client data received: dog
}
```

## Contributions

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in current work by you, as defined in the Apache-2.0 license, shall be dual licensed as described below, without any additional terms or conditions.

Feel free to send some [Pull request](https://github.com/joseluisq/gonetc/pulls) or [issue](https://github.com/joseluisq/gonetc/issues).

## License

This work is primarily distributed under the terms of both the [MIT license](LICENSE-MIT) and the [Apache License (Version 2.0)](LICENSE-APACHE).

© 2020-present [Jose Quintana](https://git.io/joseluisq)
