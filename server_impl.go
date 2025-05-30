package keyvalue

import (
	"bufio"
	"bytes"
	"fmt"
	"net"

	"github.com/Behyna/keyvalue-server/kvstore"
)

type keyValueServer struct {
	store          kvstore.KVStore
	listener       net.Listener
	clients        map[net.Conn]*client
	requests       chan request
	clientEvents   chan clientEvent
	closeChan      chan struct{}
	activeClients  int
	droppedClients int
	closed         bool
}

type request struct {
	command      []byte
	conn         net.Conn
	responseChan chan [][]byte
}

type clientEvent struct {
	eventType clientEventType
	conn      net.Conn
	client    *client
}

type clientEventType int

type client struct {
	conn     net.Conn
	response chan []byte
}

// New creates and returns (but does not start) a new KeyValueServer.
func New(store kvstore.KVStore) KeyValueServer {
	return &keyValueServer{
		store:        store,
		requests:     make(chan request),
		clientEvents: make(chan clientEvent),
		closeChan:    make(chan struct{}),
		clients:      make(map[net.Conn]*client),
	}
}

// Start keyValueServer.
func (kvs *keyValueServer) Start(port int) error {
	if kvs.closed {
		return fmt.Errorf("server already closed")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	kvs.listener = listener

	go kvs.manager()

	go kvs.acceptConnections()

	return nil
}

func (kvs *keyValueServer) Close() {
	if kvs.closed {
		return
	}
	kvs.closed = true

	close(kvs.closeChan)

	if kvs.listener != nil {
		kvs.listener.Close()
	}
}

func (kvs *keyValueServer) CountActive() int {
	return kvs.activeClients
}

func (kvs *keyValueServer) CountDropped() int {
	return kvs.droppedClients
}

func (kvs *keyValueServer) manager() {
	for {
		select {
		case <-kvs.closeChan:
			for _, c := range kvs.clients {
				c.conn.Close()
				close(c.response)
			}
			kvs.clients = make(map[net.Conn]*client)
			return

		case evt := <-kvs.clientEvents:
			switch evt.eventType {
			case clientConnect:
				kvs.clients[evt.conn] = evt.client
				kvs.activeClients++
			case clientDisconnect:
				if _, ok := kvs.clients[evt.conn]; ok {
					delete(kvs.clients, evt.conn)
					kvs.activeClients--
					kvs.droppedClients++
				}
			}

		case req := <-kvs.requests:
			kvs.processRequest(req)
		}
	}
}

func (kvs *keyValueServer) acceptConnections() {
	for {
		conn, err := kvs.listener.Accept()
		if err != nil {
			select {
			case <-kvs.closeChan:
				return
			default:
				continue
			}
		}

		kvs.handleNewClient(conn)
	}
}

func (kvs *keyValueServer) handleNewClient(conn net.Conn) {
	c := &client{
		conn:     conn,
		response: make(chan []byte, 500),
	}

	kvs.clientEvents <- clientEvent{
		eventType: clientConnect,
		conn:      conn,
		client:    c,
	}

	go kvs.clientWriter(c)
	go kvs.clientReader(c)
}

func (kvs *keyValueServer) clientWriter(c *client) {
	defer kvs.disconnectClient(c)

	for response := range c.response {
		if !bytes.HasSuffix(response, []byte("\n")) {
			response = append(response, '\n')
		}

		_, err := c.conn.Write(response)
		if err != nil {
			return
		}
	}
}

func (kvs *keyValueServer) clientReader(c *client) {
	defer kvs.disconnectClient(c)

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		cmd := make([]byte, len(line))
		copy(cmd, line)

		respChan := make(chan [][]byte, 1)
		kvs.requests <- request{
			command:      cmd,
			conn:         c.conn,
			responseChan: respChan,
		}

		responses := <-respChan
		for _, resp := range responses {
			select {
			case c.response <- resp:
			default:
			}
		}
	}
}

func (kvs *keyValueServer) disconnectClient(c *client) {
	select {
	case kvs.clientEvents <- clientEvent{
		eventType: clientDisconnect,
		conn:      c.conn,
	}:
	case <-kvs.closeChan:
		return
	}

	c.conn.Close()
}

func (kvs *keyValueServer) processRequest(req request) {
	line := req.command
	responses := make([][]byte, 0)

	_, ok := kvs.clients[req.conn]
	if !ok {
		req.responseChan <- responses
		return
	}

	parts := bytes.SplitN(line, []byte(CommandSeparator), 2)
	if len(parts) == 0 {
		req.responseChan <- responses
		return
	}

	command := string(parts[0])

	switch command {
	case PutCommand:
		if len(parts) >= 2 {
			keyValue := bytes.SplitN(parts[1], []byte(CommandSeparator), 2)
			if len(keyValue) == 2 {
				kvs.store.Put(string(keyValue[0]), keyValue[1])
			}
		}
	case GetCommand:
		if len(parts) == 2 {
			values := kvs.store.Get(string(parts[1]))
			responses = generateResponse(parts[1], values)
		}
	case UpdateCommand:
		if len(parts) == 2 {
			params := bytes.SplitN(parts[1], []byte(CommandSeparator), 3)
			if len(params) == 3 {
				kvs.store.Update(string(params[0]), params[1], params[2])
			}
		}
	case DeleteCommand:
		if len(parts) == 2 {
			kvs.store.Delete(string(parts[1]))
		}
	}

	req.responseChan <- responses
}

func generateResponse(key []byte, values [][]byte) [][]byte {
	responses := make([][]byte, 0)
	for _, val := range values {
		line := make([]byte, 0, len(key)+1+len(val))
		line = append(line, key...)
		line = append(line, ':')
		line = append(line, val...)
		responses = append(responses, line)
	}

	return responses
}
