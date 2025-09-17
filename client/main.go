package main

import (
	"log"
	"net"
)

const (
	SOCKET       = "127.0.0.1:10740"
	SIZE_OF_BUFF = 1024
)

func main() {
	// Connect to the socket via tcp
	conn, err := net.Dial("tcp", SOCKET)
	if err != nil {
		log.Fatal(err)
	}

	// Write a basic message
	message := []byte("Hello from the client!")
	_, err = conn.Write(message)
	if err != nil {
		log.Println(err)
	}

	conn.Close()
}
