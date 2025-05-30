# Key-Value Store Server

## Overview

This project implements a centralized key-value store server in Go. The server allows multiple clients to connect via TCP and perform operations such as adding, retrieving, updating, and deleting key-value pairs. It is designed to handle concurrent client connections efficiently.

## Features

- **Concurrent Client Support**: Manages multiple clients simultaneously using goroutines.
- **Key-Value Operations**:
  - `PUT`: Add a new key-value pair.
  - `GET`: Retrieve values associated with a key.
  - `UPDATE`: Update an existing key-value pair.
  - `DELETE`: Remove a key-value pair.
- **Client Management**:
  - Tracks active and dropped clients.
  - Handles client disconnections gracefully.
- **Thread-Safe Implementation**: Ensures safe access to shared resources.

## Project Structure

- `constants.go`: Contains command types and connection constants used across the server.
- `server_api.go`: Defines the `KeyValueServer` interface.
- `server_impl.go`: Implements the server logic, including client handling and request processing.
- `server_test.go`: Comprehensive test suite for server functionality.
- `kvstore/kv_api.go`: Defines the `KVStore` interface for key-value operations.
- `kvstore/kv_impl.go`: Implements the in-memory key-value store logic.
- `srunner/srunner.go`: Starts the server and listens for client connections.
- `crunner/crunner.go`: A simple client program to test the server.

## Key Components

### KeyValueServer Interface

The `KeyValueServer` interface defines the following methods:
- `Start(port int) error`: Starts the server on the specified port.
- `CountActive() int`: Returns the number of currently connected clients.
- `CountDropped() int`: Returns the number of clients dropped by the server.
- `Close()`: Shuts down the server and closes all client connections.

### Server Implementation

The server is implemented in `server_impl.go` and includes:
- **Client Management**: Tracks active and dropped clients using channels and maps.
- **Request Handling**: Processes client commands and interacts with the key-value store.
- **Concurrency**: Uses goroutines for managing client connections and processing requests.

### Key-Value Store

The key-value store is implemented in `kvstore/kv_impl.go` and provides:
- `Put(key string, value []byte)`: Adds a new key-value pair.
- `Get(key string) [][]byte`: Retrieves all values associated with a key.
- `Update(key string, oldValue, newValue []byte)`: Updates a specific value for a key.
- `Delete(key string)`: Removes a key and its associated values.

## Usage

### Starting the Server

To start the server, use the `srunner` program:

```bash
go run srunner/srunner.go
```

### Interacting with the Server

Clients can connect to the server via TCP and send commands in the following format:
- `PUT:key:value` - Adds a new key-value pair.
- `GET:key` - Retrieves all values associated with a key.
- `UPDATE:key:oldValue:newValue` - Updates a specific value for a key.
- `DELETE:key` - Removes a key and its associated values.

You can use the `crunner` program to test the server:

```bash
go run crunner/crunner.go
```

Alternatively, you can use tools like telnet or netcat to manually send commands to the server

```bash
echo "PUT:exampleKey:exampleValue" | nc localhost 9999
```

### Stopping the Server

If you are running the server using the `srunner` program, you can terminate the process manually (e.g., using `Ctrl+C` in the terminal).

## Running Tests

This project includes a comprehensive test suite to verify the functionality of both the key-value store and the server
- server_test.go: Tests server functionality including:
- Basic server operations
- Concurrent client handling
- Slow client behavior
- Client connection tracking

To run the tests, use the following command:

```bash
go test 
```