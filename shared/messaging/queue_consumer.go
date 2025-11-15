package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/wutthichod/sa-connext/shared/contracts"
)

type QueueConsumer struct {
	rb        *RabbitMQ
	connMgr   *ConnectionManager
	queueName string
}

func NewQueueConsumer(rb *RabbitMQ, connMgr *ConnectionManager, queueName string) *QueueConsumer {
	return &QueueConsumer{
		rb:        rb,
		connMgr:   connMgr,
		queueName: queueName,
	}
}

func (qc *QueueConsumer) Start() error {
	log.Printf("Starting consumer for queue: %s", qc.queueName)

	// Ensure channel is open before consuming
	log.Printf("Ensuring channel is open")
	if err := qc.rb.ensureChannel(); err != nil {
		log.Printf("ERROR: Failed to ensure channel is open: %v", err)
		return fmt.Errorf("failed to ensure channel is open: %v", err)
	}
	log.Printf("Channel is open and ready")

	qc.rb.mu.RLock()
	ch := qc.rb.Channel
	qc.rb.mu.RUnlock()

	log.Printf("Starting to consume from queue: %s", qc.queueName)

	// Verify queue exists by declaring it (idempotent operation)
	_, err := ch.QueueDeclarePassive(qc.queueName, true, false, false, false, nil)
	if err != nil {
		log.Printf("WARNING: Queue %s might not exist: %v", qc.queueName, err)
		log.Printf("Attempting to declare queue...")
		_, declareErr := ch.QueueDeclare(qc.queueName, true, false, false, false, nil)
		if declareErr != nil {
			log.Printf("ERROR: Failed to declare queue: %v", declareErr)
			return fmt.Errorf("failed to declare queue: %v", declareErr)
		}
		log.Printf("Queue declared successfully")
	} else {
		log.Printf("Queue exists and is ready")
	}

	msgs, err := ch.Consume(
		qc.queueName,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("ERROR: Failed to start consuming: %v", err)
		return err
	}
	log.Printf("Successfully started consuming from queue: %s", qc.queueName)

	go func() {
		log.Printf("Message processing goroutine started - waiting for messages...")
		messageCount := 0

		// Log that we're ready to receive messages
		log.Printf("Consumer is now actively listening for messages on queue: %s", qc.queueName)

		for msg := range msgs {
			messageCount++
			log.Printf("===== RECEIVED MESSAGE #%d =====", messageCount)
			log.Printf("RoutingKey: %s, Size: %d bytes", msg.RoutingKey, len(msg.Body))
			previewLen := 200
			if len(msg.Body) < previewLen {
				previewLen = len(msg.Body)
			}
			if previewLen > 0 {
				log.Printf("Message body preview (first %d chars): %s", previewLen, string(msg.Body[:previewLen]))
			}

			log.Printf("Unmarshaling AMQP message body")
			var msgBody contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &msgBody); err != nil {
				log.Printf("ERROR: Failed to unmarshal message body: %v", err)
				continue
			}
			log.Printf("AMQP message unmarshaled - OwnerID: %s, Data size: %d bytes", msgBody.OwnerID, len(msgBody.Data))

			userID := msgBody.OwnerID
			log.Printf("Processing message for user: %s", userID)

			var payload any
			if msgBody.Data != nil {
				log.Printf("Unmarshaling message payload for user: %s", userID)
				if err := json.Unmarshal(msgBody.Data, &payload); err != nil {
					log.Printf("ERROR: Failed to unmarshal payload for user %s: %v", userID, err)
					continue
				}
				log.Printf("Payload unmarshaled successfully for user: %s", userID)
			} else {
				log.Printf("WARNING: Message data is nil for user: %s", userID)
			}

			clientMsg := contracts.WSMessage{
				Type: msg.RoutingKey,
				Data: payload,
			}
			log.Printf("Created WebSocket message - Type: %s, for user: %s", clientMsg.Type, userID)

			log.Printf("===== GATEWAY: Sending to WebSocket =====")
			log.Printf("Target UserID: %s, MessageType: %s", userID, clientMsg.Type)
			if err := qc.connMgr.SendMessage(userID, clientMsg); err != nil {
				log.Printf("❌ GATEWAY ERROR: Failed to send message to user %s via WebSocket: %v", userID, err)
				log.Printf("===== GATEWAY: WS SEND FAILED =====")
			} else {
				log.Printf("✅ GATEWAY SUCCESS: Message successfully delivered to user %s via WebSocket", userID)
				log.Printf("✅ GATEWAY: Real-time message sent - UserID: %s, Type: %s", userID, clientMsg.Type)
				log.Printf("===== GATEWAY: WS SEND SUCCESSFUL =====")
			}
		}
		log.Printf("Message processing goroutine ended (channel closed)")
	}()

	log.Printf("Consumer started successfully for queue: %s", qc.queueName)
	return nil
}
