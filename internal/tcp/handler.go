package tcp

import (
	"encoding/json"
	"gomailer/config"
	"gomailer/internal/email"
	"log"
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
	var emailData email.EmailData
	if err := json.Unmarshal(message, &emailData); err != nil {
		log.Printf("Error parsing email data: %v", err)
		return createErrorResponse("Invalid email data format")
	}

	if err := h.emailService.SendEmail(&emailData); err != nil {
		log.Printf("Error sending email: %v", err)
		return createErrorResponse("Failed to send email")
	}

	return createSuccessResponse()
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

func createSuccessResponse() []byte {
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Email queued successfully",
	}
	responseBytes, _ := json.Marshal(response)
	return responseBytes
} 