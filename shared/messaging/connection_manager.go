package messaging

import (
	"encoding/json"
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
	log.Printf("Added WebSocket connection for user: %s", userID)
}

// Remove unregisters a user connection
func (cm *ConnectionManager) Remove(userID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.connections, userID)
	log.Printf("Removed WebSocket connection for user: %s", userID)
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
	log.Printf("===== ATTEMPTING TO SEND WS MESSAGE =====")
	log.Printf("UserID: %s, MessageType: %s", userID, message.Type)

	cm.mutex.RLock()
	wrapper, exists := cm.connections[userID]
	cm.mutex.RUnlock()

	if !exists {
		log.Printf("❌ ERROR: No WebSocket connection found for user: %s", userID)
		log.Printf("Available connections: %d", len(cm.connections))
		return ErrConnectionNotFound
	}
	log.Printf("✓ WebSocket connection found for user: %s", userID)
	log.Printf("Preparing message payload...")

	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	// Send the full message structure including Type
	res := map[string]interface{}{
		"success": true,
		"type":    message.Type,
		"data":    message.Data,
	}

	// Log message details before sending
	if messageData, err := json.Marshal(res); err == nil {
		log.Printf("Message payload size: %d bytes", len(messageData))
	} else {
		log.Printf("WARNING: Could not marshal message for logging: %v", err)
	}

	log.Printf("Writing message to WebSocket for user: %s", userID)
	err := wrapper.conn.WriteJSON(res)
	if err != nil {
		log.Printf("❌ ERROR: Failed to write JSON to WebSocket for user %s: %v", userID, err)
		log.Printf("===== WS SEND FAILED =====")
	} else {
		log.Printf("✅ SUCCESS: Message successfully sent to user %s via WebSocket", userID)
		log.Printf("✅ MessageType: %s, UserID: %s", message.Type, userID)
		log.Printf("===== WS SEND SUCCESSFUL =====")
	}
	return err
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

// BroadcastToAll sends a message to all connected users
func (cm *ConnectionManager) BroadcastToAll(message contracts.WSMessage) {
	cm.mutex.RLock()
	userIDs := make([]string, 0, len(cm.connections))
	for userID := range cm.connections {
		userIDs = append(userIDs, userID)
	}
	cm.mutex.RUnlock()

	log.Printf("Broadcasting message to %d users", len(userIDs))
	for _, userID := range userIDs {
		if err := cm.SendMessage(userID, message); err != nil {
			log.Printf("Failed to broadcast to user %s: %v", userID, err)
		}
	}
}