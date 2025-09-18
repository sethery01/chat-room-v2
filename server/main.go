/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 *	Information used in project from: https://pkg.go.dev/
***********************************************************************/
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	IP           = "127.0.0.1"
	PORT         = ":10740"
	SIZE_OF_BUFF = 1024
)

func sendMessage(conn net.Conn, message []byte) {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
	}
}

func validateUser(username, password string) bool {
	// Open the users file
	file, err := os.Open("users.txt")
	if err != nil {
		log.Println(err)
		return false
	}
	defer file.Close()	// Close file after func exit

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Parse the txt file for the username and password
		line := scanner.Text()
		line = strings.TrimPrefix(line, "(")
		line = strings.TrimSuffix(line, ")")
		user := strings.Split(line,",")
		user[0] = strings.TrimSpace(user[0])
		user[1] = strings.TrimSpace(user[1])
		log.Println("User in validate func:", user[0], user[1])
		
		// Check if username and password match
		if user[0] == username && user[1] == password {
			log.Println("Returning true")
			return true
		}
	}
	return false
}

func login(command []string) bool {
	username := command[1]
	password := command[2]
	log.Println(username,password)
	return validateUser(username,password)
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close connection upon function exit
	loggedIn := false

	for {
		// Read in data sent by the connection
		buffer := make([]byte, SIZE_OF_BUFF)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}

		// Parse request
		request := buffer[0:bytesRead]
		data := string(request)
		command := strings.Fields(data) 

		// Execute the command
		switch command[0] {
		case "login":
			loggedIn = login(command)
			var message []byte
			if loggedIn {
				message = []byte("1")
			} else {
				message = []byte("0")
			}
			sendMessage(conn,message)
		case "newuser":
			fmt.Println("> newuser command chosen")
		case "send":
			if loggedIn {
				fmt.Println("> send command chosen")
			} else {
				fmt.Println("> You must login before sending a message.")
			}
		case "logout":
			if loggedIn {
				fmt.Println("> See you next time!")
				return
			} else {
				fmt.Println("> You must login before logging out.")
			}
		default:
			fmt.Println("> Invalid command")
		}
	}
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
