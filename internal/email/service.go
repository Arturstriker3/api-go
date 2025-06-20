package email

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Arturstriker3/api-go/config"
	"github.com/Arturstriker3/api-go/internal/metrics"
	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	To        []string  `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	QueuedAt  time.Time `json:"queued_at"`
}

type Service struct {
	config   *config.Config
	dialer   *gomail.Dialer
	channel  *amqp.Channel
}

func NewEmailService(cfg *config.Config) *Service {
	dialer := gomail.NewDialer(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password,
	)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Failed to open channel: %v", err))
	}

	// Declare the queue
	_, err = ch.QueueDeclare(
		"email_queue", // queue name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to declare queue: %v", err))
	}

	return &Service{
		config:  cfg,
		dialer:  dialer,
		channel: ch,
	}
}

// QueueEmail adds the email to the RabbitMQ queue
func (s *Service) QueueEmail(data *EmailData) error {
	if len(data.To) == 0 {
		metrics.EmailErrors.Inc()
		return fmt.Errorf("recipient list is empty")
	}

	// Add timestamp when queueing
	data.QueuedAt = time.Now()

	// Convert email data to JSON
	body, err := json.Marshal(data)
	if err != nil {
		metrics.EmailErrors.Inc()
		return fmt.Errorf("failed to marshal email data: %w", err)
	}

	// Publish to queue
	err = s.channel.Publish(
		"",           // exchange
		"email_queue", // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		metrics.EmailErrors.Inc()
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	metrics.EmailsQueued.Inc()
	return nil
}

// SendEmail sends the email directly via SMTP (used by the consumer)
func (s *Service) SendEmail(data *EmailData) error {
	if len(data.To) == 0 {
		metrics.EmailErrors.Inc()
		return fmt.Errorf("recipient list is empty")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.SMTP.From)
	m.SetHeader("To", data.To...)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", data.Body)

	if err := s.dialer.DialAndSend(m); err != nil {
		metrics.EmailErrors.Inc()
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Calculate delivery time if timestamp exists
	if !data.QueuedAt.IsZero() {
		deliveryTime := time.Since(data.QueuedAt).Seconds()
		metrics.EmailDeliveryTime.Observe(deliveryTime)
	}

	metrics.EmailsSent.Inc()
	return nil
} 