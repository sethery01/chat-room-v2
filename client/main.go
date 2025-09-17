package main

import (
	"log"
	"net"
	"fmt"
	"os"
	"bufio"
	"strings"
)

const (
	SOCKET       = "127.0.0.1:10740"
	SIZE_OF_BUFF = 1024
)

func start(conn net.Conn) {
	fmt.Println("******************************************************************")
	fmt.Print("Hello! Welcome to Seth Ek's chatbot V1.\n\nAvailable commands:\n")
	fmt.Print("login \"UserID\" \"Password\"\nnewuser \"UserID\" \"Password\"\nsend \"message\"\nlogout\n")
	fmt.Print("\nPlease enter commands as shown above. You must begin with login.\n")
	fmt.Println("******************************************************************")

	reader := bufio.NewReader(os.Stdin)

	// Wait for input from user
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		command := strings.TrimSpace(input)
		fmt.Println("> Your command: " + command)
	}
}

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

	// Run the app
	start(conn)

	conn.Close()
}
