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
	SOCKET       = "127.0.0.1:10740"
	SIZE_OF_BUFF = 1024
	MAX_CLIENTS  = 3
)

// Globals for multithreading and mutex locks
var activeUsers = make(map[string]net.Conn) // Maps usernames to their actual connection
var activeConns int                         // Tracks active conns
var userMutex sync.RWMutex                  // Read/Write mutex lock for activeUsers
var fileMutex sync.Mutex                    // Mutex for file I/O with users.txt
var connMutex sync.Mutex                    // Mutex for activeConns int

func sendMessage(conn net.Conn, message []byte) {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
	}
}

func validateUser(username, password string, newuser bool) bool {
	// Lock users.txt
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Open the users file
	file, err := os.Open("users.txt")
	if err != nil {
		log.Println(err)
		return false
	}
	defer file.Close() // Close file after func exit

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Parse the txt file for the username and password
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // skip empty lines
		}
		line = strings.TrimPrefix(line, "(")
		line = strings.TrimSuffix(line, ")")
		user := strings.Split(line, ",")
		if len(user) < 2 {
			continue // skip bad lines
		}
		user[0] = strings.TrimSpace(user[0])
		user[1] = strings.TrimSpace(user[1])

		// Check if username and password match for login case
		if user[0] == username && user[1] == password && !newuser {
			return true
		}

		// Check if user exists for newuser case
		if user[0] == username && newuser {
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
		return 2 // User already active in chatroom
	}

	// Try to login
	validUser := validateUser(username, password, false)
	if !validUser {
		return 1 // User doesn't exist
	}

	// Set active user, notify connected clients, and return success
	activeUsers[command[1]] = conn
	return 0
}

// FIX NEWUSER NOT CHECKING RIGHT
func newuser(command []string) bool {
	username := command[1]
	password := command[2]

	userExists := validateUser(username, password, true)
	if userExists {
		log.Printf("User %s already exists.\n", username)
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
		defer file.Close() // Close file after func exit

		// If file is empty, don't add a newline
		fileInfo, _ := file.Stat()
		prefix := ""
		if fileInfo.Size() > 0 {
			prefix = "\n"
		}

		// Write the new user to the EOF
		user := fmt.Sprintf("%s(%s, %s)", prefix, username, password)
		_, err = file.Write([]byte(user))
		if err != nil {
			log.Println(err)
			return false
		}
	}
	log.Print("User added: " + username)
	return true
}

func sendAll(message []byte) {
	userMutex.RLock()
	for username, conn := range activeUsers {
		log.Printf("Sending to %s\n", username)
		sendMessage(conn, message)
	}
	userMutex.RUnlock()
}

func sendUserID(currentConn net.Conn, message []byte, userID, activeUser string) {
	userMutex.RLock()
	defer userMutex.RUnlock()

	// Lookup user in the map
	conn, exists := activeUsers[userID]
	if !exists {
		log.Printf("Attempted to send message to %s, but user not found.\n", userID)
		sendMessage(currentConn, []byte(fmt.Sprintf("Error: %s is not online.", userID)))
		return
	}

	// send message
	sendMessage(conn, message)
	log.Printf("\"%s\" sent to %s from %s", string(message), userID, activeUser)
}

func who(conn net.Conn) {
	userMutex.RLock()
	defer userMutex.RUnlock()

	// Collect all active usernames
	var usernames []string
	for username := range activeUsers {
		usernames = append(usernames, username)
	}

	// Do some message formatting here
	var message string
	switch len(usernames) {
	case 0:
		message = "No active users.\n"
	case 1:
		message = fmt.Sprintf("Active user: %s\n", usernames[0])
	default:
		message = fmt.Sprintf("Active users: %s\n", strings.Join(usernames, ", "))
	}

	sendMessage(conn, []byte(message))
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close connection upon function exit

	// Attempt connection limit check
	connMutex.Lock()
	if activeConns >= MAX_CLIENTS {
		connMutex.Unlock()
		log.Println("Connection refused from:", conn.RemoteAddr().String())
		sendMessage(conn, []byte("-1"))
		conn.Close()
		return
	}

	// Connection allowed!
	activeConns++
	log.Printf("New connection from: %s Active connections: %d\n", conn.RemoteAddr().String(), activeConns)
	connMutex.Unlock()
	sendMessage(conn, []byte("0"))

	loggedIn := false
	errorCode := 0
	activeUser := ""

	// Make sure number of conns is updated after exit
	defer func() {
		connMutex.Lock()
		activeConns--
		log.Printf("Connection closed. Active connections: %d\n", activeConns)
		connMutex.Unlock()
		if activeUser != "" {
			userMutex.Lock()
			delete(activeUsers, activeUser)
			userMutex.Unlock()
			log.Printf("%s closed their connected unexpectedly", activeUser)
			sendAll([]byte(fmt.Sprintf("%s left!", activeUser)))
		}
	}()

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
			sendMessage(conn, message)
			if loggedIn {
				sendAll([]byte(fmt.Sprintf("%s joined!", activeUser)))
			}
		case "newuser":
			created := newuser(command)
			var message []byte
			if created {
				message = []byte("1")
			} else {
				message = []byte("0")
			}
			sendMessage(conn, message)
		case "send":
			data := strings.Join(command[2:], " ")
			message := []byte(fmt.Sprintf("%s: %s", activeUser, data))
			if command[1] == "all" {
				sendAll(message)
				log.Println(string(message))
			} else {
				sendUserID(conn, message, command[1], activeUser)
			}

		case "who":
			who(conn)

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
	// Ensure users.txt exists before server starts
	if _, err := os.Stat("users.txt"); os.IsNotExist(err) {
		file, err := os.Create("users.txt")
		if err != nil {
			log.Fatalf("Failed to create users.txt: %v", err)
		}
		file.Close()
		log.Println("Created users.txt")
	}

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
