/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 *	Information used in project from: https://pkg.go.dev/
***********************************************************************/
package main

import (
	"fmt"
	"log"
	"net"
)

const (
	IP = "127.0.0.1"
	PORT = ":10740"
	SIZE_OF_BUFF = 1024
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close connection upon function exit

	// Read in data sent by the connection
	buffer := make([]byte, SIZE_OF_BUFF)		// Creates a buffer slice
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return
	}
	for i := 0; i < bytesRead; i++ {
		fmt.Print(string(buffer[i]))
	}
	fmt.Println("")
}

func main() {
	// Start the server
	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ln.Close() // Close upon exiting main
	fmt.Println("Server is listening on port ", PORT)

	// Listen for and handle connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}
