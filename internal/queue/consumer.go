package queue

import (
	"encoding/json"
	"fmt"
	"gomailer/config"
	"gomailer/internal/email"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	emailService *email.Service
}

func NewConsumer(cfg *config.Config, emailService *email.Service) (*Consumer, error) {
	amqpURI := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		emailService: emailService,
	}, nil
}

func (c *Consumer) Setup() error {
	// Declare the queue
	queue, err := c.channel.QueueDeclare(
		"email_queue", // queue name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS
	err = c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	return nil
}

func (c *Consumer) StartConsuming() error {
	msgs, err := c.channel.Consume(
		"email_queue", // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			var emailData email.EmailData
			if err := json.Unmarshal(msg.Body, &emailData); err != nil {
				log.Printf("Error decoding message: %v", err)
				msg.Nack(false, false)
				continue
			}

			if err := c.emailService.SendEmail(&emailData); err != nil {
				log.Printf("Error sending email: %v", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
			log.Printf("Email sent successfully to %v", emailData.To)
		}
	}()

	return nil
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
} 