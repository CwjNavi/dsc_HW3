package main

import (
	"fmt"
	"net"
	"net/rpc"
)

// KeyValueStore represents the server's key-value store
type KeyValueStore struct {
	store map[string]string
}

// Args defines the arguments for the RPC methods
type Args struct {
	Key   string
	Value string // Optional for setting key-value pairs
}

// Response defines the response structure
type Response struct {
	Value string
	Error string
}

// Get retrieves a value by key
func (kv *KeyValueStore) Get(args *Args, reply *Response) error {
	value, exists := kv.store[args.Key]
	if !exists {
		reply.Error = "Key not found"
		return nil
	}
	reply.Value = value
	return nil
}

// Start the server
func main() {
	kv := &KeyValueStore{store: map[string]string{
		"notexampleKey": "exampleValue", // Sample data
	}}

	// Register the key-value store as an RPC service
	err := rpc.Register(kv)
	if err != nil {
		fmt.Println("Error registering KeyValueStore:", err)
		return
	}

	// Listen for incoming connections
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is running on port 1234...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
