package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pires/go-proxyproto"
)

func handleClient(client net.Conn, target string) {
	defer client.Close()

	server, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("Error connecting to target: %v", err)
		return
	}
	defer server.Close()

	go func() {
		if _, err := io.Copy(server, client); err != nil {
			log.Printf("Error copying data from client to server: %v", err)
		}
	}()

	if _, err := io.Copy(client, server); err != nil {
		log.Printf("Error copying data from server to client: %v", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbServers := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if strings.HasPrefix(pair[0], "DB_SERVER_") {
			port := strings.TrimPrefix(pair[0], "DB_SERVER_")
			dbServers[port] = pair[1]
		}
	}

	for port, target := range dbServers {
		go func(port, target string) {
			listener, err := net.Listen("tcp", ":"+port)
			if err != nil {
				log.Fatalf("Error starting listener on port %s: %v", port, err)
			}
			fmt.Println("Starting listener on port", port)
			proxyListener := &proxyproto.Listener{Listener: listener}

			for {
				client, err := proxyListener.Accept()
				if err != nil {
					log.Printf("Error accepting connection: %v", err)
					continue
				}
				go handleClient(client, target)
			}
		}(port, target)
	}
	// Mantén el programa en ejecución
	select {}
}
