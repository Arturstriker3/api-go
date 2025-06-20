package api

import (
	"encoding/json"
	"fmt"
	"gomailer/config"
	"gomailer/internal/email"
	"log"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler struct {
	channel *amqp.Channel
}

func NewHandler(cfg *config.Config) (*Handler, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Handler{
		channel: ch,
	}, nil
}

func (h *Handler) QueueEmail(c *gin.Context) {
	var emailData email.EmailData

	if err := c.ShouldBindJSON(&emailData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate email data
	if len(emailData.To) == 0 {
		c.JSON(400, gin.H{"error": "Recipient list is empty"})
		return
	}
	if emailData.Subject == "" {
		c.JSON(400, gin.H{"error": "Subject is required"})
		return
	}
	if emailData.Body == "" {
		c.JSON(400, gin.H{"error": "Body is required"})
		return
	}

	// Convert email data to JSON
	body, err := json.Marshal(emailData)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to encode email data"})
		return
	}

	// Publish to queue
	err = h.channel.Publish(
		"",           // exchange
		"email_queue", // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		c.JSON(500, gin.H{"error": "Failed to queue email"})
		return
	}

	c.JSON(202, gin.H{"message": "Email queued successfully"})
}

func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	router.POST("/email", handler.QueueEmail)

	return router
} 