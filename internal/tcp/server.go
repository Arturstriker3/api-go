package tcp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/Arturstriker3/api-go/config"
	"github.com/Arturstriker3/api-go/internal/email"
	"github.com/Arturstriker3/api-go/internal/metrics"
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
	// Check if any TCP service is enabled
	if !s.config.TCP.Enabled && !s.config.TCP.TLS.Enabled {
		return fmt.Errorf("❌ Both TCP and TLS are disabled. Enable at least one with TCP_ENABLED=true or TCP_TLS_ENABLED=true")
	}

	var listener net.Listener
	var err error
	
	address := fmt.Sprintf(":%s", s.config.TCP.Port)
	
	// Determine which mode to use based on configuration
	if s.config.TCP.TLS.Enabled && s.config.TCP.Enabled {
		// Both enabled - prioritize TLS for security
		log.Printf("⚠️  Both TCP and TLS are enabled. Using TLS for security (TCP_ENABLED will be ignored)")
		log.Printf("💡 Set TCP_ENABLED=false to disable insecure TCP completely")
	}
	
	if s.config.TCP.TLS.Enabled {
		// Load TLS certificates
		cert, err := tls.LoadX509KeyPair(s.config.TCP.TLS.CertPath, s.config.TCP.TLS.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   "localhost", // For development
		}

		listener, err = tls.Listen("tcp", address, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to start TLS server: %w", err)
		}
		
		log.Printf("🔒 TLS Server listening on port %s (SECURE)", s.config.TCP.Port)
		log.Printf("📜 Using certificate: %s", s.config.TCP.TLS.CertPath)
		log.Printf("🛡️  All connections will be encrypted")
	} else if s.config.TCP.Enabled {
		listener, err = net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("failed to start TCP server: %w", err)
		}
		
		log.Printf("⚠️  TCP Server listening on port %s (INSECURE)", s.config.TCP.Port)
		log.Printf("💡 Consider enabling TLS with TCP_TLS_ENABLED=true")
		log.Printf("🔒 For production, set TCP_ENABLED=false and TCP_TLS_ENABLED=true")
	}
	
	s.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			metrics.TCPErrors.Inc()
			continue
		}

		// Log connection type
		if s.config.TCP.TLS.Enabled {
			if tlsConn, ok := conn.(*tls.Conn); ok {
				log.Printf("🔒 New TLS connection from %s", conn.RemoteAddr())
				// Log TLS details
				state := tlsConn.ConnectionState()
				log.Printf("   Cipher Suite: %s", tls.CipherSuiteName(state.CipherSuite))
				log.Printf("   TLS Version: %x", state.Version)
			}
		} else {
			log.Printf("⚠️  New insecure TCP connection from %s", conn.RemoteAddr())
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

	authenticated := false

	for {
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		message := buffer[:n]
		
		if !authenticated {
			// Primeiro, tentar autenticar
			var auth struct {
				Secret string `json:"secret"`
			}
			if err := json.Unmarshal(message, &auth); err == nil && auth.Secret != "" {
				if auth.Secret != s.authSecret {
					metrics.TCPAuthErrors.Inc()
					sendError(conn, "Invalid authentication")
					return
				}
				
				authenticated = true
				metrics.TCPAuthSuccess.Inc()
				sendSuccess(conn, "Authentication successful")
				continue
			}
			
			// Se não é autenticação, rejeitar
			sendError(conn, "Authentication required")
			return
		}

		// Se já autenticado, processar mensagem de email
		response := s.handler.HandleMessage(message)
		conn.Write(response)
	}
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

func sendSuccess(conn net.Conn, message string) {
	response := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	responseBytes, _ := json.Marshal(response)
	conn.Write(responseBytes)
} 