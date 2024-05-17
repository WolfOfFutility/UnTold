package main

// This is intended to be used between multiple instances of the keystore, and will allow for redundant replication.
// ** Still a work in progress - Asymmetric Encryption needs to be in place before this can occur.

import (
	"net"
	"fmt"
)

type SocketAuth struct {
	Username string
	Token []byte
}

type SocketMessage struct {
	Auth SocketAuth
	Message []byte
}

// As a server, listen for incoming connections
func startServer() {
	// Listen for incoming connections on port 8080
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println(err)
        return
    }

    // Accept incoming connections and handle them
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }

        // Handle the connection in a new goroutine
        go handleConnection(conn)
    }
}

// As a server, handle a connection
func handleConnection(connection net.Conn) {
	// Close the connection when we're done
    defer connection.Close()

    // Read incoming data
    buf := make([]byte, 1024)
    _, err := connection.Read(buf)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Print the incoming data
    fmt.Printf("Received: %s", buf)
}

// As a client, create a connection
func createConnection() {
     // Connect to the server
     conn, err := net.Dial("tcp", "localhost:8080")
     if err != nil {
         fmt.Println(err)
         return
     }
 
     // Send some data to the server
     _, err = conn.Write([]byte("Hello, server!"))
     if err != nil {
         fmt.Println(err)
         return
     }
 
     // Close the connection
     defer conn.Close()
}

func closeConnection() {}