/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 * 	October 24, 2025
 *	Go net library: 	https://pkg.go.dev/net
 * 	Go sync library: 	https://pkg.go.dev/sync
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
	"sync"
)

// TODO:
// Handle file i/o async DONE
// Handle multiple users logging in async... i.e. cannot be two sessions of Seth

// GLOBALS
const (
	SOCKET       	= "127.0.0.1:10740"
	SIZE_OF_BUFF 	= 1024
	MAX_CLIENTS		= 3
)

// Globals for multithreading and mutex locks
var activeUsers = make(map[string]net.Conn)	// Maps usernames to their actual connection
var activeConns int							// Tracks active conns
var userMutex sync.RWMutex					// Read/Write mutex lock for activeUsers
var fileMutex sync.Mutex					// Mutex for file I/O with users.txt
var connMutex sync.Mutex					// Mutex for activeConns int


func sendMessage(conn net.Conn, message []byte) {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
	}
}

func validateUser(username, password string) bool {
	// Lock users.txt
	fileMutex.Lock()
	defer fileMutex.Unlock()
	
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

func login(command []string, conn net.Conn) int {
	username := command[1]
	password := command[2]

	// Lock shared activeUsers
	userMutex.Lock()
	defer userMutex.Unlock()

	// Check if user is already active
	_, loggedIn := activeUsers[username]
	if loggedIn {
		return 2 	// User already active in chatroom
	}
	
	// Try to login
	validUser := validateUser(username,password)
	if !validUser {
		return 1 	// User doesn't exist
	}

	// Set active user and return success
	activeUsers[command[1]] = conn
	return 0
}

func newuser(command []string) bool {
	username := command[1]
	password := command[2]
	
	userExists := validateUser(username, password)
	if userExists {
		log.Printf("User %s already exists.\n",username)
		return false
	} else {
		// Lock users.txt
		fileMutex.Lock()
		defer fileMutex.Unlock()

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
	defer conn.Close() // Close connection upon function exit
	
	// Attempt connection limit check
	connMutex.Lock()
	if activeConns >= MAX_CLIENTS {
		connMutex.Unlock()
		log.Println("Connection refused from:", conn.RemoteAddr().String())
		conn.Close()
		return
	}

	// Connection allowed!
	activeConns++
	log.Printf("New connection from: %s Active connections: %d\n", conn.RemoteAddr().String(), activeConns)
	connMutex.Unlock()

	// Make sure number of conns is updated after exit
	defer func() {
		connMutex.Lock()
		activeConns--
		log.Printf("Connection closed. Active connections: %d\n", activeConns)
		connMutex.Unlock()
	}()

	loggedIn := false
	errorCode := 0
	activeUser := ""

	// Handle accepted connection
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
			errorCode = login(command, conn)
			message := []byte(fmt.Sprintf("%d", errorCode))
			if errorCode == 0 {
				loggedIn = true
				activeUser = command[1] 
				log.Printf("New login. Active Users: %v\n", activeUsers)
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
				userMutex.Lock()
				delete(activeUsers, activeUser)
				log.Println("Terminating connection: " + conn.RemoteAddr().String())
				log.Printf("Active Users: %v\n", activeUsers)
				userMutex.Unlock()
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
