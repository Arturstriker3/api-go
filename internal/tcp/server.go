package tcp

import (
	"encoding/json"
	"fmt"
	"gomailer/config"
	"gomailer/internal/email"
	"gomailer/internal/metrics"
	"log"
	"net"
)

type Server struct {
	config      *config.Config
	listener    net.Listener
	handler     *Handler
	authSecret  string
}

func NewServer(cfg *config.Config, emailService *email.Service) (*Server, error) {
	handler := NewHandler(cfg, emailService)
	return &Server{
		config:     cfg,
		handler:    handler,
		authSecret: cfg.TCP.AuthSecret,
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.config.TCP.Port))
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}
	s.listener = listener

	log.Printf("TCP Server listening on port %s", s.config.TCP.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			metrics.TCPErrors.Inc()
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	metrics.TCPConnections.Inc()
	defer metrics.TCPConnections.Dec()

	// Primeiro, autenticar a conexão
	authMsg := make([]byte, 1024)
	n, err := conn.Read(authMsg)
	if err != nil {
		log.Printf("Error reading auth message: %v", err)
		metrics.TCPErrors.Inc()
		return
	}

	var auth struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(authMsg[:n], &auth); err != nil {
		log.Printf("Error parsing auth message: %v", err)
		metrics.TCPErrors.Inc()
		sendError(conn, "Invalid auth format")
		return
	}

	if auth.Secret != s.authSecret {
		log.Printf("Invalid auth secret received")
		metrics.TCPErrors.Inc()
		sendError(conn, "Invalid authentication")
		return
	}

	// Após autenticação, ler a mensagem de email
	buffer := make([]byte, 4096)
	n, err = conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading message: %v", err)
		metrics.TCPErrors.Inc()
		return
	}

	response := s.handler.HandleMessage(buffer[:n])
	conn.Write(response)
}

func sendError(conn net.Conn, message string) {
	response := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}
	responseBytes, _ := json.Marshal(response)
	conn.Write(responseBytes)
} 