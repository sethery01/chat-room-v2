/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 *	Informatin used in poject from: https://pkg.go.dev/net
***********************************************************************/
package main

import (
	"fmt"
	"log"
	"net"
)

const (
	PORT = ":10740"
)

func handleConnection(conn net.Conn) {

}

func main() {
	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Server is listening on port ", PORT)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		go handleConnection(conn)
	}
}
