package gonetc

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"testing"
	"time"
)

// unixSocketPath is a default Unix socket path used on tests.
const unixSocketPath = "/tmp/mysocket"

// unixSocketDelay defines milliseconds pause in order to wait until
// the listening server (`socat`) is ready to accept connections.
const unixSocketDelay = 500

// listeningSocket defines a listening unix socket.
type listeningSocket struct {
	cmd *exec.Cmd
	wg  *sync.WaitGroup
}

// createListeningSocket creates a new listening unix socket using `socat` tool.
func createListeningSocket() (*listeningSocket, error) {
	exec.Command("rm", "-rf", unixSocketPath).Run()

	var out bytes.Buffer
	cmd := exec.Command("socat", "UNIX-LISTEN:"+unixSocketPath+",fork", "exec:'/bin/cat'")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go cmd.Wait()
	time.Sleep(unixSocketDelay * time.Millisecond)

	return &listeningSocket{
		wg:  &wg,
		cmd: cmd,
	}, nil
}

// close method closes current socket connection signaling it to finish.
func (s *listeningSocket) close() error {
	return s.cmd.Process.Signal(os.Interrupt)
}

func TestNew(t *testing.T) {
	type args struct {
		network        string
		unixSocketPath string
	}
	tests := []struct {
		name string
		args args
		want *NetClient
	}{
		{
			name: "valid unix socket client instance",
			args: args{
				network:        "unix",
				unixSocketPath: "/tmp/mysocket",
			},
			want: &NetClient{
				network:      "unix",
				address:      "/tmp/mysocket",
				MaxReadBytes: 2048,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.network, tt.args.unixSocketPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNetClient_Connect(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tests := []struct {
		name           string
		network        string
		unixSocketPath string
		wantErr        bool
	}{
		{
			name:           "invalid unix socket connection",
			network:        "unix",
			unixSocketPath: unixSocketPath + "xyz",
			wantErr:        true,
		},
		{
			name:           "valid unix socket connection",
			network:        "unix",
			unixSocketPath: unixSocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.network, tt.unixSocketPath)
			if err := c.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("NetClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if c.conn == nil {
					t.Errorf("Connect() = conn: %v, want not nil", c.conn)
				}
			}
		})
	}

	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestNetClient_Write(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	type args struct {
		data        []byte
		respHandler func(data []byte, err error, done func())
	}
	tests := []struct {
		name           string
		network        string
		unixSocketPath string
		args           args
		socketNil      bool
		wantN          int
		wantErr        bool
	}{
		{
			name:           "valid unix socket write without handler",
			network:        "unix",
			unixSocketPath: unixSocketPath,
			args: args{
				data:        []byte(nil),
				respHandler: nil,
			},
			wantN: 0,
		},
		{
			name:           "valid unix socket write with handler",
			network:        "unix",
			unixSocketPath: unixSocketPath,
			args: args{
				data:        []byte("ñ"),
				respHandler: func(data []byte, err error, done func()) { done() },
			},
			wantN: 2,
		},
		{
			name:           "nil socket connection reference",
			network:        "unix",
			unixSocketPath: unixSocketPath,
			socketNil:      true,
			args: args{
				data: []byte(nil),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.network, tt.unixSocketPath)
			// Check nil socket references on demand
			if !tt.socketNil {
				if err := c.Connect(); (err != nil) != tt.wantErr {
					t.Errorf("NetClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			gotN, err := c.Write(tt.args.data, tt.args.respHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("NetClient.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("NetClient.Write() = %v, want %v", gotN, tt.wantN)
			}
		})
	}

	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestNetClient_Close(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tests := []struct {
		name           string
		network        string
		unixSocketPath string
		socketNil      bool
		wantErr        bool
	}{
		{
			name:           "close current socket connection",
			network:        "unix",
			unixSocketPath: unixSocketPath,
		},
		{
			name:           "close invalid socket connection",
			network:        "unix",
			unixSocketPath: unixSocketPath,
			socketNil:      true,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.network, tt.unixSocketPath)
			if !tt.socketNil {
				if err := c.Connect(); (err != nil) != tt.wantErr {
					t.Errorf("NetClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("NetClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestNetClient_readData(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	type fields struct {
		network      string
		address      string
		Conn         net.Conn
		MaxReadBytes int
	}
	type args struct {
		respHandler func(data []byte, err error, done func())
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		socketNil         bool
		wantErr           bool
		wantConnErr       bool
		wantData          []byte
		closeListenSocket bool
	}{
		{
			name: "valid socket response",
			fields: fields{
				network: "unix",
				address: unixSocketPath,
			},
			args: args{
				respHandler: func(data []byte, err error, done func()) { done() },
			},
		},
		{
			name: "error on invalid socket connection",
			fields: fields{
				network: "unix",
				address: unixSocketPath,
			},
			args: args{
				respHandler: func(data []byte, err error, done func()) { done() },
			},
			socketNil: true,
			wantErr:   true,
		},
		{
			name: "invalid max read bytes value",
			fields: fields{
				network:      "unix",
				address:      unixSocketPath,
				MaxReadBytes: -1,
			},
			args: args{
				respHandler: func(data []byte, err error, done func()) { done() },
			},
			wantErr: true,
		},
		{
			name: "interrupted socket connection",
			fields: fields{
				network: "unix",
				address: unixSocketPath,
			},
			args: args{
				respHandler: func(data []byte, err error, done func()) { done() },
			},
			closeListenSocket: true,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NetClient{
				network:      tt.fields.network,
				address:      tt.fields.address,
				conn:         tt.fields.Conn,
				MaxReadBytes: tt.fields.MaxReadBytes,
			}
			// Check nil socket references on demand
			if !tt.socketNil {
				if err := c.Connect(); (err != nil) != tt.wantConnErr {
					t.Errorf("NetClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if tt.closeListenSocket {
				if err := lsock.close(); err != nil {
					t.Errorf("%v", err)
					return
				}
				time.Sleep(300 * time.Millisecond)
			}
			c.readData(func(data []byte, err error, done func()) {
				if len(data) != len(tt.wantData) {
					t.Errorf("NetClient.readData(respHandler) data = %v, want %v", string(data), string(tt.wantData))
				}
				if (err != nil) != tt.wantErr {
					t.Errorf("NetClient.readData(respHandler) error = %v, wantErr %v", err != nil, tt.wantErr)
				}
				tt.args.respHandler(data, err, done)
			})
		})
	}
	if err := lsock.close(); err != nil {
		return
	}
}

func TestNetClient_Conn(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	tests := []struct {
		name           string
		network        string
		unixSocketPath string
		wantNil        bool
		wantErr        bool
	}{
		{
			name:    "invalid socket connection",
			wantNil: true,
		},
		{
			name:           "return valid socket connection",
			network:        "unix",
			unixSocketPath: unixSocketPath,
			wantNil:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.network, tt.unixSocketPath)
			if !tt.wantNil {
				if err := c.Connect(); (err != nil) != tt.wantErr {
					t.Errorf("NetClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			got := c.Conn()
			if (got == nil) != tt.wantNil {
				t.Errorf("NetClient.Conn() = Nil %v, wantNil %v", got == nil, tt.wantNil)
			}
			c.Close()
		})
	}
	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}
