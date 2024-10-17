package main

import (
	"fmt"
	"strconv"
	"strings"
)

/* PARSERS */
func parseEPRT(command string) (protocol string, ip string, port int, err error) {
	parts := strings.Split(strings.TrimSpace(command[6:len(command)-3]), "|")
	if len(parts) != 3 {
		fmt.Println("parts: ", parts, len(parts))
		return "", "", 0, fmt.Errorf("invalid EPRT command format")
	}
	protocol = parts[0]
	ip = parts[1]
	port, err = strconv.Atoi(parts[2])

	if err != nil {
		return "", "", 0, fmt.Errorf("invalid port number")
	}

	return protocol, ip, port, nil
}
