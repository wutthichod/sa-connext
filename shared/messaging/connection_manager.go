package messaging

import (
	"errors"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/wutthichod/sa-connext/shared/contracts"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

// connWrapper is a wrapper around the websocket connection to allow for thread-safe operations
// WebSocket connections are not thread-safe by default
type connWrapper struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

type ConnectionManager struct {
	connections map[string]*connWrapper // userId -> connection
	mutex       sync.RWMutex
}

// NewConnectionManager initializes the manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*connWrapper),
	}
}

// Add registers a WebSocket connection for a user
func (cm *ConnectionManager) Add(userID string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.connections[userID] = &connWrapper{
		conn:  conn,
		mutex: sync.Mutex{},
	}
	log.Printf("Added connection for user %s", userID)
}

// Remove unregisters a user connection
func (cm *ConnectionManager) Remove(userID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.connections, userID)
}

// Get retrieves the WebSocket connection for a user
func (cm *ConnectionManager) Get(userID string) (*websocket.Conn, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	wrapper, exists := cm.connections[userID]
	if !exists {
		return nil, false
	}
	return wrapper.conn, true
}

// SendMessage sends a message safely to a connected user
func (cm *ConnectionManager) SendMessage(userID string, message contracts.WSMessage) error {
	cm.mutex.RLock()
	wrapper, exists := cm.connections[userID]
	cm.mutex.RUnlock()

	if !exists {
		return ErrConnectionNotFound
	}

	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	res := &contracts.Resp{
		Success: true,
		Data:    message.Data,
	}
	return wrapper.conn.WriteJSON(res)
}

func (cm *ConnectionManager) GetAllUserIDs() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	userIDs := make([]string, 0, len(cm.connections))
	for userID := range cm.connections {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}