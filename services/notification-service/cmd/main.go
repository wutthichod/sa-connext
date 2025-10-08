package main

import (
	"log"
	"github.com/wutthichod/sa-connext/shared/messaging"
)

func main() {
	rb, err := messaging.NewRabbitMQ("amqps://nsemvrni:JFrEKtzZLj9jmmwXKZ_VSNZoP5M6u8ON@gorilla.lmq.cloudamqp.com/nsemvrni");
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rb.Close()

	queueName, err := rb.SetupQueue(
		"email_queue",              // queue name
		"notification.exchange",    // exchange
		"direct",                   // exchange type
		"notification.email",       // routing key
		true,                       // durable
		nil,                        // args
	)
	if err != nil {
    	log.Fatalf("Failed to setup email queue: %v", err)
	}

	emailConsumer := messaging.NewEmailConsumer(rb, queueName,"","")
	if err := emailConsumer.Start(); err != nil {
		log.Fatalf("Failed to start email consumer: %v", err)
	}
	select {} // Block forever
}