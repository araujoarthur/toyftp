package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var dataConnListener net.Listener

/* HANDLERS */
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Greets the client
	conn.Write([]byte("220 FTP Server Ready\r\n")) // <CRLF> is required by TELNET iirc

	// Data Connection Variable
	var dataConn net.Conn

	// Command processing loop
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf) // Read command from buffer
		if err != nil {
			log.Printf("Failed to read from connection: %v", err)
		}

		input := string(buf[:n])
		fmt.Printf("Received: %s", input)

		if len(input) >= 4 && input[:4] == "USER" {
			conn.Write([]byte("331 Username OK, need password\r\n"))
		} else if len(input) >= 4 && input[:4] == "EPRT" {
			dataConn, err = handleEPRTCommand(conn, input)
			if err != nil {
				fmt.Printf("Error in EPRT command: %v\n", err)
			}
		} else if len(input) >= 4 && input[:4] == "PASS" {
			conn.Write([]byte("230 User logged in\r\n"))
		} else if len(input) >= 4 && input[:4] == "QUIT" {
			conn.Write([]byte("Goodbye\r\n"))
			break
		} else if len(input) >= 4 && input[:4] == "PASV" {
			handlePASVCommand(conn)
		} else if len(input) >= 4 && input[:4] == "STOR" {
			filename := strings.TrimSpace(input[5:])
			if filename == "" {
				conn.Write([]byte("501 No filename given\r\n"))
			} else {
				fmt.Printf("Received filename %s\n", filename)
				handleStorCommand(conn, filename, dataConn)
			}
		} else if len(input) >= 4 && input == "LIST" {
			handleListCommand(conn)
		} else {
			conn.Write([]byte("502 Command not implemented\r\n"))
		}
	}
}

func handleEPRTCommand(conn net.Conn, command string) (dataConn net.Conn, err error) {
	protocol, ip, port, err := parseEPRT(command)
	if err != nil {
		conn.Write([]byte("500 Syntax error in EPRT command \r\n"))
		return nil, err
	}

	var network string
	if protocol == "1" {
		network = "tcp4"
	} else if protocol == "2" {
		network = "tcp6"
	} else {
		conn.Write([]byte("522 Network protocol not supported\r\n"))
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	dataConn, err = net.Dial(network, fmt.Sprintf("[%s]:%d", ip, port))

	if err != nil {
		conn.Write([]byte("425 can't open data connection\r\n"))
		return nil, err
	}

	conn.Write([]byte("200 EPRT command successful\r\n"))

	return dataConn, nil
}

func handlePASVCommand(conn net.Conn) {
	var err error
	dataConnListener, err = net.Listen("tcp", ":0") // it'll listen on a random port chosen by the server (it's what :0 does, let's the OS pick)
	if err != nil {
		log.Printf("Error setting up data connection listener: %v\n", err)
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	// Retrieve the port chosen by the server
	addr := dataConnListener.Addr().String()
	_, port, _ := net.SplitHostPort(addr)
	portInt, _ := strconv.Atoi(port)

	// Converts ports to high byte, low byte (ftp format)
	portHigh := portInt / 256
	portLow := portInt % 256

	ip := "127.0.0.1"
	response := fmt.Sprintf("227 Entering Passive Mode (%s, %d, %d)", ip, portHigh, portLow)
	conn.Write([]byte(response))
	fmt.Println("PASV Enabled, listening on port ", port)
}

// LIST command
func handleListCommand(conn net.Conn) {
	files, err := os.ReadDir(".")
	if err != nil {
		conn.Write([]byte("550 Failed to list directory\r\n"))
		return
	}

	conn.Write([]byte("150 Here comes the directory listing\r\n"))
	for _, file := range files {
		fileInfo, _ := file.Info()
		conn.Write([]byte(fileInfo.Name() + "\r\n"))
	}

	conn.Write([]byte("226 Directory send OK\r\n"))
}

// RETR command
func handleStorCommand(conn net.Conn, filename string, dconn net.Conn) {

	// Accepts the data connection
	conn.Write([]byte("150 Opening data connection\r\n"))
	/*dataConn, err := dataConnListener.Accept()
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}*/

	defer dconn.Close()

	// Creates the file on serverside
	file, err := os.Create(filename)
	if err != nil {
		conn.Write([]byte("550 Failed to create file \r\n"))
		return
	}

	defer file.Close()

	_, err = io.Copy(file, dconn)
	if err != nil {
		conn.Write([]byte("550 Failed to receive file\r\n"))
		return
	}

	conn.Write([]byte("226 Transfer complete \r\n"))
}
