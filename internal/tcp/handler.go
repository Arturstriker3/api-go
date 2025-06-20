package tcp

import (
	"encoding/json"
	"log"

	"github.com/Arturstriker3/api-go/config"
	"github.com/Arturstriker3/api-go/internal/email"
	"github.com/Arturstriker3/api-go/internal/metrics"
)

type Handler struct {
	config       *config.Config
	emailService *email.Service
}

func NewHandler(cfg *config.Config, emailService *email.Service) *Handler {
	return &Handler{
		config:       cfg,
		emailService: emailService,
	}
}

func (h *Handler) HandleMessage(message []byte) []byte {
	// Check for authentication message
	var authData struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(message, &authData); err == nil && authData.Secret != "" {
		if authData.Secret == h.config.TCP.AuthSecret {
			return createSuccessResponse("Authentication successful")
		}
		return createErrorResponse("Invalid authentication secret")
	}

	// Handle email message
	var emailData email.EmailData
	if err := json.Unmarshal(message, &emailData); err != nil {
		log.Printf("Error parsing email data: %v", err)
		metrics.EmailErrors.Inc()
		return createErrorResponse("Invalid email data format")
	}

	if err := h.emailService.QueueEmail(&emailData); err != nil {
		log.Printf("Error queueing email: %v", err)
		metrics.EmailErrors.Inc()
		return createErrorResponse("Failed to queue email")
	}

	metrics.EmailsQueued.Inc()
	return createSuccessResponse("Email queued successfully")
}

func createErrorResponse(message string) []byte {
	response := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}
	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

func createSuccessResponse(message string) []byte {
	response := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	responseBytes, _ := json.Marshal(response)
	return responseBytes
} 