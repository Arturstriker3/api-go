package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type EmailClient struct {
	host       string
	port       string
	authSecret string
}

type EmailRequest struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewEmailClient(host, port, authSecret string) *EmailClient {
	return &EmailClient{
		host:       host,
		port:       port,
		authSecret: authSecret,
	}
}

func (c *EmailClient) SendEmail(request *EmailRequest) error {
	// Conectar ao servidor
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.host, c.port), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to email service: %w", err)
	}
	defer conn.Close()

	// Enviar autenticação
	auth := struct {
		Secret string `json:"secret"`
	}{
		Secret: c.authSecret,
	}

	authBytes, err := json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	_, err = conn.Write(authBytes)
	if err != nil {
		return fmt.Errorf("failed to send auth data: %w", err)
	}

	// Enviar requisição de email
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	_, err = conn.Write(requestBytes)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}

	// Ler resposta
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response Response
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("email service error: %s", response.Error)
	}

	return nil
} 