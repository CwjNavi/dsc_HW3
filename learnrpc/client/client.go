package main

import (
	"fmt"
	"net/rpc"
)

// Args and Response structs must match the server's definitions
type Args struct {
	Key   string
	Value string
}

type Response struct {
	Value string
	Error string
}

// Local cache for storing key-value pairs
var localCache = make(map[string]string)

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer client.Close()

	key := "exampleKey" // Key to request from the server
	var response Response

	// Check cache first
	if value, exists := localCache[key]; exists {
		fmt.Printf("Cache hit: %s -> %s\n", key, value)
		return
	}

	// If not in cache, query the server
	args := Args{Key: key}
	err = client.Call("KeyValueStore.Get", &args, &response)
	if err != nil {
		fmt.Println("RPC error:", err)
		return
	}

	// Handle the response
	if response.Error != "" {
		fmt.Println("Server response:", response.Error)
	} else {
		fmt.Printf("Server response: %s -> %s\n", key, response.Value)

		// Store the result in the local cache
		localCache[key] = response.Value
		fmt.Println("Value cached locally.")
	}
}
