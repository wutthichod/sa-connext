package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
	uri     string
	mu      sync.RWMutex
}

// NewRabbitMQ connects and returns a RabbitMQ client
func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		uri: uri,
	}

	if err := rmq.connect(); err != nil {
		return nil, err
	}

	return rmq, nil
}

// connect establishes a new connection and channel
// This method acquires a write lock - do not call while holding a lock
func (r *RabbitMQ) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.connectUnsafe()
}

// connectUnsafe establishes a new connection and channel without acquiring locks
// Caller must hold the write lock
func (r *RabbitMQ) connectUnsafe() error {
	conn, err := amqp.Dial(r.uri)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create channel: %v", err)
	}

	// Close existing connection/channel if any
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}

	r.conn = conn
	r.Channel = ch

	return nil
}

// ensureChannel ensures the channel is open, recreating it if necessary
func (r *RabbitMQ) ensureChannel() error {
	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	// Check if channel is closed
	if ch == nil || ch.IsClosed() {
		log.Println("Channel is closed, attempting to reconnect...")
		r.mu.Lock()
		defer r.mu.Unlock()

		// Double-check after acquiring write lock
		if r.Channel != nil && !r.Channel.IsClosed() {
			return nil
		}

		// Check if connection is also closed
		if r.conn == nil || r.conn.IsClosed() {
			log.Println("Connection is also closed, reconnecting...")
			if err := r.connectUnsafe(); err != nil {
				return fmt.Errorf("failed to reconnect: %v", err)
			}
			log.Println("Successfully reconnected to RabbitMQ")
			return nil
		}

		// Connection is open, just recreate channel
		newCh, err := r.conn.Channel()
		if err != nil {
			// Connection might have closed between check and channel creation
			log.Println("Failed to create channel, reconnecting...")
			if err := r.connectUnsafe(); err != nil {
				return fmt.Errorf("failed to reconnect: %v", err)
			}
			log.Println("Successfully reconnected to RabbitMQ")
			return nil
		}

		r.Channel = newCh
		log.Println("Successfully recreated channel")
		return nil
	}

	return nil
}

// Close safely closes channel and connection
func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Channel != nil {
		r.Channel.Close()
		r.Channel = nil
	}
	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}
}

func (r *RabbitMQ) SetupQueue(queueName, exchangeName, exchangeType, routingKey string, durable bool, args amqp.Table) (string, error) {
	if err := r.ensureChannel(); err != nil {
		return "", err
	}

	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	// Declare exchange
	if err := ch.ExchangeDeclare(exchangeName, exchangeType, durable, false, false, false, nil); err != nil {
		return "", fmt.Errorf("failed to declare exchange %s: %v", exchangeName, err)
	}

	// Declare queue
	q, err := ch.QueueDeclare(queueName, durable, false, false, false, args)
	if err != nil {
		return "", fmt.Errorf("failed to declare queue %s: %v", queueName, err)
	}

	// Bind queue to exchange
	if err := ch.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
		return "", fmt.Errorf("failed to bind queue %s to exchange %s: %v", queueName, exchangeName, err)
	}

	return q.Name, nil
}

// DeclareExchange declares an exchange
func (r *RabbitMQ) DeclareExchange(name, kind string, durable bool) error {
	if err := r.ensureChannel(); err != nil {
		return err
	}

	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	return ch.ExchangeDeclare(
		name,
		kind,
		durable,
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	)
}

// DeclareQueue declares a queue and returns its name
func (r *RabbitMQ) DeclareQueue(name string, durable bool, args amqp.Table) (string, error) {
	if err := r.ensureChannel(); err != nil {
		return "", err
	}

	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	q, err := ch.QueueDeclare(
		name,
		durable,
		false, // delete when unused
		false, // exclusive
		false, // noWait
		args,
	)
	if err != nil {
		return "", err
	}
	return q.Name, nil
}

// BindQueue binds a queue to an exchange with a routing key
func (r *RabbitMQ) BindQueue(queue, exchange, routingKey string) error {
	if err := r.ensureChannel(); err != nil {
		return err
	}

	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	return ch.QueueBind(queue, routingKey, exchange, false, nil)
}

// PublishMessage publishes a JSON message to an exchange with a routing key
func (r *RabbitMQ) PublishMessage(ctx context.Context, exchange, routingKey string, message interface{}) error {
	// Ensure channel is open before publishing
	if err := r.ensureChannel(); err != nil {
		return fmt.Errorf("failed to ensure channel is open: %v", err)
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	err = ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)

	// If publish failed due to closed channel, try once more after reconnecting
	if err != nil {
		if ch.IsClosed() {
			log.Println("Channel closed during publish, reconnecting and retrying...")
			if err := r.ensureChannel(); err != nil {
				return fmt.Errorf("failed to reconnect: %v", err)
			}

			r.mu.RLock()
			ch = r.Channel
			r.mu.RUnlock()

			err = ch.PublishWithContext(ctx,
				exchange,
				routingKey,
				false, false,
				amqp.Publishing{
					ContentType:  "application/json",
					DeliveryMode: amqp.Persistent,
					Body:         body,
				},
			)
		}
	}

	return err
}

// ConsumeMessages starts consuming messages from a queue
func (r *RabbitMQ) ConsumeMessages(queue string, handler func(msg []byte) error) error {
	if err := r.ensureChannel(); err != nil {
		return err
	}

	r.mu.RLock()
	ch := r.Channel
	r.mu.RUnlock()

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("failed to handle message: %v", err)
				msg.Nack(false, true) // requeue if failed
				continue
			}
			msg.Ack(false)
		}
	}()

	return nil
}
