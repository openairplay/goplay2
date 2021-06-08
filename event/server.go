package event

import (
	"fmt"
	"net"
	"os"
)

func RunEventServer() {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", ":60003")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	// Close the listener when the application closes.
	defer l.Close()
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleEventConnection(conn)
	}
}

func handleEventConnection(conn net.Conn) {
	defer conn.Close()
	// Handle connections in a new goroutine.
	for {
		var buffer [4096]byte
		_, err := conn.Read(buffer[:])
		if err != nil {
			fmt.Printf("Event error : %v", err)
			break
		}
	}
}

