package main

import (
	"fmt"

	"github.com/Behyna/keyvalue-server"
	"github.com/Behyna/keyvalue-server/kvstore"
)

const defaultPort = 9999

func main() {
	// Initialize the key-value store.
	store, _ := kvstore.CreateWithBackdoor()
	// Initialize the server.

	server := keyvalue.New(store)
	if server == nil {
		fmt.Println("New() returned a nil server. Exiting...")
		return
	}

	// Start the server and continue listening for client connections in the background.
	if err := server.Start(defaultPort); err != nil {
		fmt.Printf("KeyValueServer could not be started: %s\n", err)
		return
	}

	fmt.Printf("Started KeyValueServer on port %d...\n", defaultPort)

	// Block forever.
	select {}
}
