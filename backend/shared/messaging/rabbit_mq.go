package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitMQ connects and returns a RabbitMQ client
func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %v", err)
	}

	return &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}, nil
}

// Close safely closes channel and connection
func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQ) SetupQueue(queueName, exchangeName, exchangeType, routingKey string, durable bool, args amqp.Table) (string, error) {
	// Declare exchange
	if err := r.Channel.ExchangeDeclare(exchangeName, exchangeType, durable, false, false, false, nil); err != nil {
		return "", fmt.Errorf("failed to declare exchange %s: %v", exchangeName, err)
	}

	// Declare queue
	q, err := r.Channel.QueueDeclare(queueName, durable, false, false, false, args)
	if err != nil {
		return "", fmt.Errorf("failed to declare queue %s: %v", queueName, err)
	}

	// Bind queue to exchange
	if err := r.Channel.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
		return "", fmt.Errorf("failed to bind queue %s to exchange %s: %v", queueName, exchangeName, err)
	}

	return q.Name, nil
}

// DeclareExchange declares an exchange
func (r *RabbitMQ) DeclareExchange(name, kind string, durable bool) error {
	return r.Channel.ExchangeDeclare(
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
	q, err := r.Channel.QueueDeclare(
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
	return r.Channel.QueueBind(queue, routingKey, exchange, false, nil)
}

// PublishMessage publishes a JSON message to an exchange with a routing key
func (r *RabbitMQ) PublishMessage(ctx context.Context, exchange, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	return r.Channel.PublishWithContext(ctx,
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

// ConsumeMessages starts consuming messages from a queue
func (r *RabbitMQ) ConsumeMessages(queue string, handler func(msg []byte) error) error {
	msgs, err := r.Channel.Consume(queue, "", false, false, false, false, nil)
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
