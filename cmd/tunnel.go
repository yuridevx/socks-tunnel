package main

import (
	"context"
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

func runTunnel(c *cli.Context) error {
	// Override config values with command line flags if they are set
	// ...

	listener, err := net.Listen("tcp", config.LocalAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", config.LocalAddr, err)
	}
	defer listener.Close()
	log.Printf("Listening on %s", config.LocalAddr)

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

		go handleConnection(clientConn, config.RemoteAddr, config.Username, config.Password)
	}
}

func handleConnection(clientConn net.Conn, remoteAddr, username, password string) {
	defer clientConn.Close()
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
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() {
		_, err := io.Copy(remoteConn, clientConn)
		if err != nil {
			cancel(err)
			log.Printf("Failed to copy data to remote server: %v", err)
			return
		}
		cancel(ErrCanceled)
	}()
	go func() {
		_, err := io.Copy(clientConn, remoteConn)
		if err != nil {
			cancel(err)
			log.Printf("Failed to copy data to client: %v", err)
			return
		}
		cancel(ErrCanceled)
	}()
	<-ctx.Done()
}
