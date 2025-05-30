package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	defaultHost = "localhost"
	defaultPort = 9999
	timeout     = 5 * time.Second
)

func main() {
	host := flag.String("host", defaultHost, "Server hostname")
	port := flag.Int("port", defaultPort, "Server port")
	flag.Parse()

	serverAddr := fmt.Sprintf("%s:%d", *host, *port)

	runInteractive(serverAddr)
}

func runInteractive(serverAddr string) {
	fmt.Println("KV Store Client - Interactive Mode")
	fmt.Println("Enter commands or 'exit' to quit")
	fmt.Println("  exit - Exit the client")
	fmt.Println("Command format examples:")
	fmt.Println("  Put:mykey:myvalue")
	fmt.Println("  Get:mykey")
	fmt.Println("  Update:mykey:oldvalue:newvalue")
	fmt.Println("  Delete:mykey")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := scanner.Text()
		if strings.ToLower(command) == "exit" {
			break
		}

		if command == "" {
			continue
		}

		conn, err := connectToServer(serverAddr)
		if err != nil {
			continue
		}

		if !strings.HasSuffix(command, "\n") {
			command += "\n"
		}

		_, err = conn.Write([]byte(command))
		if err != nil {
			fmt.Printf("Error sending command: %v\n", err)
			conn.Close()
			continue
		}

		conn.SetReadDeadline(time.Now().Add(timeout))

		reader := bufio.NewReader(conn)
		responseCount := 0
		for {
			response, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						break
					} else {
						fmt.Printf("Error reading response: %v\n", err)
					}
				}
				break
			}

			responseCount++
			fmt.Printf("Response: %s", response)
		}

		if responseCount == 0 && strings.HasPrefix(strings.ToUpper(command), "GET") {
			fmt.Println("No values found for this key")
		}

		conn.Close()
	}
}

func connectToServer(serverAddr string) (net.Conn, error) {
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Error connecting to server at %s: %v\n", serverAddr, err)
		return nil, err
	}
	return conn, nil
}
