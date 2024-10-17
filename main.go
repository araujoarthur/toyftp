package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":21")
	if err != nil {
		log.Fatalf("Error setting up the listener on port 21: %v", err)
	}
	defer listener.Close()

	fmt.Println("FTP Server Listening to Port 21")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept a connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
