package event

import (
	"goplay2/globals"
	"net"
	"os"
)

func RunEventServer() {
	// listen for incoming connections.
	l, err := net.Listen("tcp", ":60003")
	if err != nil {
		globals.ErrLog.Println("Error listening:", err.Error())
		return
	}
	// Close the listener when the application closes.
	defer l.Close()
	for {
		// listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			globals.ErrLog.Println("Error accepting: ", err.Error())
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
			globals.ErrLog.Printf("Event error : %v", err)
			break
		}
	}
}
