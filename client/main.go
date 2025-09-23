/***********************************************************************
 *	Seth Ek
 *	Networks
 *	Chatbot V1
 *	Information used in project from: 
 * https://pkg.go.dev/
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

func login(conn net.Conn, command string) bool {
	// Send the login message
	message := []byte(command)
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
		return false
	}

	// Listen for the server response
	buffer := make([]byte,SIZE_OF_BUFF)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return false
	}

	// Parse response
	response := buffer[0:bytesRead]
	data := string(response)
	
	// Validate login
	if data != "1" {
		fmt.Println("> Login unsuccessful.")
		return false
	}
	fmt.Println("> You are logged in!")
	return true
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
	_, err := conn.Write(message)
	if err != nil {
		log.Println(err)
		return false
	}

	// Listen for the server response
	buffer := make([]byte,SIZE_OF_BUFF)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return false
	}

	// Parse response
	response := buffer[0:bytesRead]
	data := string(response)
	
	// Validate login
	if data != "1" {
		fmt.Println("> Account creation unsuccessful.")
		return false
	}
	fmt.Println("> New account created! Please login.")
	return true
}

func start(conn net.Conn) {
	fmt.Println("******************************************************************")
	fmt.Print("Hello! Welcome to Seth Ek's chatbot V1.\n\nAvailable commands:\n")
	fmt.Print("login \"UserID\" \"Password\"\nnewuser \"UserID\" \"Password\"\nsend \"message\"\nlogout\n")
	fmt.Print("\nPlease enter commands as shown above. You must begin with login.\n")
	fmt.Println("******************************************************************")

	reader := bufio.NewReader(os.Stdin)
	loggedIn := false

	// Wait for input from user
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')     	// Get input
		inputString := strings.TrimSpace(input) 	// Trim whitespace here
		command := strings.Fields(inputString)  	// strings.Fields creates a slice of strings seperated by a space

		// Execute the command
		switch command[0] {
		case "login":
			if loggedIn {
				fmt.Println("> You are already logged in.")
			} else if len(command) == 3 {
				loggedIn = login(conn, inputString)
			} else {
				fmt.Println("> You must provided a username and password.")
			}
		case "newuser":
			if len(command) != 3 {
				fmt.Println("> You must provided a username and password.")
			} else if len(command[1]) < 3 || len(command[1]) > 32 {
				fmt.Println("> Your username must be between 3 and 32 characters.")
			} else if len(command[2]) < 4 || len(command[2]) > 8 {
				fmt.Println("> Your password must be between 4 and 8 characters.")
			} else {
				newuser(conn, inputString)
			}
		case "send":
			if loggedIn {
				fmt.Println("> send command chosen")
			} else {
				fmt.Println("> You must login before sending a message.")
			}
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
		log.Fatal(err) // This will kill thr program
	}

	// Run the app
	start(conn)

	conn.Close()
}
