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
)

const (
	SOCKET       = "127.0.0.1:10740"
	SIZE_OF_BUFF = 1024
)

func sendAndReceive(conn net.Conn, message []byte) string {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
		return "0"
	}

	// Listen for the server response
	buffer := make([]byte, SIZE_OF_BUFF)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return "0"
	}

	// Parse response
	response := buffer[0:bytesRead]
	data := string(response)

	return data
}

func sendMessage(conn net.Conn, message []byte) {
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
	}
}

func validateConn(conn net.Conn) string {
	buffer := make([]byte, SIZE_OF_BUFF)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return "-1"
	}

	// Parse response
	response := buffer[0:bytesRead]
	data := string(response)

	return data
}

func login(conn net.Conn, command string, user string) bool {
	// // Send the login message
	message := []byte(command)
	data := sendAndReceive(conn, message)

	// Validate login
	switch data {
	case "1":
		fmt.Println("> Login unsuccessful. User: " + user + " doesn't exist.")
		return false
	case "2":
		fmt.Println("> Login unsuccessful. " + user + " is already logged in.")
		return false
	default:
		fmt.Println("> You are logged in!")
		return true
	}
}

func logout(conn net.Conn) bool {
	// Send the logout message
	message := []byte("logout")
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func newuser(conn net.Conn, command string) bool {
	// Send the newuser message
	message := []byte(command)
	data := sendAndReceive(conn, message)

	// Validate login
	if data != "1" {
		fmt.Println("> Account creation unsuccessful.")
		return false
	}
	fmt.Println("> New account created! Please login.")
	return true
}

func send(conn net.Conn, command string) {
	// Send the message to be echoed to the user
	message := []byte(command)
	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("> Error:", err)
	}
}

func start(conn net.Conn) {
	fmt.Println("******************************************************************")
	fmt.Print("Hello! Welcome to Seth Ek's chatbot V2.\n\nAvailable commands:\n")
	fmt.Print("login \"UserID\" \"Password\"\nnewuser \"UserID\" \"Password\"\nsend all \"message\"\nsend \"UserID\" \"message\"logout\n")
	fmt.Print("\nPlease enter commands as shown above. You must begin with login.\n")
	fmt.Println("******************************************************************")

	reader := bufio.NewReader(os.Stdin)
	loggedIn := false

	// Wait for input from user
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')     // Get input
		inputString := strings.TrimSpace(input) // Trim whitespace here
		command := strings.Fields(inputString)  // strings.Fields creates a slice of strings seperated by a space

		if len(command) == 0 {
			continue
		}

		// Execute the command
		switch command[0] {
		
		case "login":
			if loggedIn {
				fmt.Println("> You are already logged in.")
			} else if len(command) == 3 {
				loggedIn = login(conn, inputString, command[1])
				if loggedIn {
					// Goroutine: Listen for server messages continuously
					go func() {
						buffer := make([]byte, SIZE_OF_BUFF)
						for {
							bytesRead, err := conn.Read(buffer)
							if err != nil {
								fmt.Println("\n[Server disconnected]")
								os.Exit(0)
							}
							message := strings.TrimSpace(string(buffer[:bytesRead]))
							if message != "" {
								fmt.Printf("\r> %s\n> ", message)
							}
						}
					}()
				}
			} else {
				fmt.Println("> You must provided a username and password.")
			}
		
		case "newuser":
			if loggedIn {
				fmt.Println("> Denied. You cannot create a newuser while logged in.")
			} else if len(command) != 3 {
				fmt.Println("> You must provided a username and password.")
			} else if len(command[1]) < 3 || len(command[1]) > 32 {
				fmt.Println("> Your username must be between 3 and 32 characters.")
			} else if len(command[2]) < 4 || len(command[2]) > 8 {
				fmt.Println("> Your password must be between 4 and 8 characters.")
			} else {
				newuser(conn, inputString)
			}
		
		case "send":
			if !loggedIn {
				fmt.Println("> You must login before sending a message.")
			} else if len(command) < 3 {
				fmt.Println("> You must include an argument and a message.")
			} else if len(strings.Join(command[2:], " ")) > 256 {
				fmt.Println("> Your message must be 1-256 characters long.")
			} else {
				send(conn, inputString)
			}
		
		case "who":
			sendMessage(conn, []byte("who"))
		
		case "logout":
			if loggedIn {
				loggedOut := logout(conn)
				if loggedOut {
					fmt.Println("> See you next time!")
					return
				}
				fmt.Println("> Error logging out.")
			} else {
				fmt.Println("> You must login before logging out.")
			}
		default:
			fmt.Println("> Invalid command")
		}
	}
}

func main() {
	// Connect to the socket via tcp
	conn, err := net.Dial("tcp", SOCKET)
	if err != nil {
		fmt.Println("Please start the server first.")
		log.Fatal(err) // This will kill thr program
	}

	// Run the app if connection accepted
	data := validateConn(conn)
	if data == "0" {
		start(conn)
	} else {
		fmt.Println("Connection refused. Server full.")
	}

	conn.Close()
}
