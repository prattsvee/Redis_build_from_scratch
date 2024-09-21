package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("Listening on port :6379")

	// Create a new TCP server listening on port 6379
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer l.Close()

	// Continuously listen for incoming connections
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each connection concurrently
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}

		if value.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(value.array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		// Get the command and arguments
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		writer := NewWriter(conn)

		// Look up the handler for the command
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command:", command)
			writer.Write(Value{typ: "error", str: fmt.Sprintf("ERR unknown command '%s'", command)})
			continue
		}

		// Execute the command handler and write the response
		result := handler(args)
		writer.Write(result)
	}
}
