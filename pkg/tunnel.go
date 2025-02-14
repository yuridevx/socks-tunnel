package pkg

import (
	"errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/proxy"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

var ErrCanceled = errors.New("operation canceled")

func RunTunnel(c *cli.Context) error {
	// Override config values with command line flags if they are set
	// ...

	listener, err := net.Listen("tcp", ConfigInst.LocalAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", ConfigInst.LocalAddr, err)
	}
	defer listener.Close()
	log.Printf("Listening on %s", ConfigInst.LocalAddr)

	// Create a channel to listen for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt

		// Close listener when interrupt signal is received
		listener.Close()
		log.Println("Shutting down...")
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "accept" {
				log.Println("Listener closed")
				return nil
			}

			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(clientConn, ConfigInst.RemoteAddr, ConfigInst.Username, ConfigInst.Password)
	}
}

func handleConnection(clientConn net.Conn, remoteAddr, username, password string) {
	auth := proxy.Auth{
		User:     username,
		Password: password,
	}
	dialer, err := proxy.SOCKS5("tcp", remoteAddr, &auth, proxy.Direct)
	if err != nil {
		log.Printf("Failed to create SOCKS5 dialer: %v", err)
		return
	}

	remoteConn, err := dialer.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("Failed to connect to remote SOCKS5 server: %v", err)
		return
	}
	go func() {
		io.Copy(remoteConn, clientConn)
		remoteConn.Close()
	}()
	go func() {
		io.Copy(clientConn, remoteConn)
		clientConn.Close()
	}()
}
