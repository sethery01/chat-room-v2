/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 * 	October 24, 2025
 *	Information used in project from: https://pkg.go.dev/
 *	server/main.go
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
	SOCKET       = "127.0.0.1:10740"
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
		
		// Check if username and password match
		if user[0] == username && user[1] == password {
			log.Println("User logged in as: " + username)
			return true
		}
	}
	return false
}

func login(command []string) bool {
	username := command[1]
	password := command[2]
	return validateUser(username,password)
}

func newuser(command []string) bool {
	username := command[1]
	password := command[2]
	
	userExists := validateUser(username, password)
	if userExists {
		log.Printf("User %s already exists.\n",username)
		return false
	} else {
		// Open the users file for appending
		file, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return false
		}
		defer file.Close()	// Close file after func exit

		// Write the new user to the EOF
		user := fmt.Sprintf("\n(%s, %s)",username,password)
		_, err = file.Write([]byte(user))
		if err != nil {
			log.Println(err)
			return false
		}
	}
	log.Print("User added: " + username)
	return true
}

func handleConnection(conn net.Conn) {
	log.Println("New connection from: " + conn.RemoteAddr().String())
	defer conn.Close() // Close connection upon function exit
	loggedIn := false
	activeUser := ""

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
				activeUser = command[1]
				message = []byte("1")
			} else {
				message = []byte("0")
			}
			sendMessage(conn,message)
		case "newuser":
			created := newuser(command)
			var message []byte
			if created {
				message = []byte("1")
			} else {
				message = []byte("0")
			}
			sendMessage(conn,message)
		case "send":
			if loggedIn {
				data = data[5:]
				message := []byte(fmt.Sprintf("%s: %s",activeUser,data))
				log.Println(string(message))
				sendMessage(conn, message)
			} else {
				log.Println("User is not logged in.")
			}
		case "logout":
			if loggedIn {
				log.Println("Terminating connection: " + conn.RemoteAddr().String())
				return
			} else {
				log.Println("User chose logout but nobody is logged in.")
			}
		default:
			log.Println("Invalid command")
		}
	}
}

func main() {
	// Start the server
	ln, err := net.Listen("tcp", SOCKET)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ln.Close() // Close upon exiting main
	fmt.Println("Server is listening on " + ln.Addr().String())

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
