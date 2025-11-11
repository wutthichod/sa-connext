package messaging

import (
	"encoding/json"
	"log"
	"net/smtp"

	"github.com/wutthichod/sa-connext/shared/contracts"
)

type EmailConsumer struct {
	rb        *RabbitMQ
	queueName string
	from      string
	password  string
	smtpHost  string
	smtpPort  string
}

// NewEmailConsumer creates a new consumer for email events
func NewEmailConsumer(rb *RabbitMQ, queueName, from, password string) *EmailConsumer {
	return &EmailConsumer{
		rb:        rb,
		queueName: queueName,
		from:      from,
		password:  password,
		smtpHost:  "smtp.gmail.com",
		smtpPort:  "587",
	}
}

// Start begins consuming messages from the queue
func (ec *EmailConsumer) Start() error {
	msgs, err := ec.rb.Channel.Consume(
		ec.queueName,
		"",
		true,  // auto-ack (set to false for manual ack if needed)
		false, // not exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			var event contracts.EmailEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("Failed to unmarshal EmailEvent:", err)
				continue
			}

			if err := ec.sendEmail(&event); err != nil {
				log.Printf("Failed to send email to %s: %v", event.To, err)
				continue
			}

			log.Printf("Email sent to %s successfully!", event.To)
		}
	}()

	log.Printf(" [*] EmailConsumer listening on queue: %s", ec.queueName)
	return nil
}

// sendEmail sends a general-purpose email using SMTP
func (ec *EmailConsumer) sendEmail(event *contracts.EmailEvent) error {
	to := []string{event.To}
	subject := event.Subject
	body := event.Body

	msg := []byte("To: " + event.To + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", ec.from, ec.password, ec.smtpHost)
	return smtp.SendMail(ec.smtpHost+":"+ec.smtpPort, auth, ec.from, to, msg)
}
