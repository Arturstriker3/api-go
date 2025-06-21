package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Arturstriker3/api-go/config"
	"github.com/Arturstriker3/api-go/internal/email"
	"github.com/Arturstriker3/api-go/internal/metrics"
)

type Server struct {
	config      *config.Config
	listener    net.Listener
	handler     *Handler
	authSecret  string
	tlsConfig   *tls.Config
	certMutex   sync.RWMutex
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
		return fmt.Errorf("🔴 Both TCP and TLS are disabled. Enable at least one with TCP_ENABLED=true or TCP_TLS_ENABLED=true")
	}

	var listener net.Listener
	var err error
	
	address := fmt.Sprintf(":%s", s.config.TCP.Port)
	
	// Determine which mode to use based on configuration
	if s.config.TCP.TLS.Enabled && s.config.TCP.Enabled {
		// Both enabled - prioritize TLS for security
		log.Printf("🟡 Both TCP and TLS are enabled. Using TLS for security (TCP_ENABLED will be ignored)")
		log.Printf("💡 Set TCP_ENABLED=false to disable insecure TCP completely")
	}
	
	if s.config.TCP.TLS.Enabled {
		// Load TLS certificates
		cert, err := tls.LoadX509KeyPair(s.config.TCP.TLS.CertPath, s.config.TCP.TLS.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		// Check certificate expiry
		if len(cert.Certificate) > 0 {
			x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
			if err == nil {
				daysUntilExpiry := time.Until(x509Cert.NotAfter).Hours() / 24
				metrics.TLSCertificateExpiry.Set(daysUntilExpiry)
				log.Printf("📅 Certificate expires in %.0f days (%s)", daysUntilExpiry, x509Cert.NotAfter.Format("2006-01-02"))
			}
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   "localhost", // For development
		}

		s.tlsConfig = tlsConfig // Store reference for hot reload

		listener, err = tls.Listen("tcp", address, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to start TLS server: %w", err)
		}
		
		log.Printf("🔒 TLS Server listening on port %s (SECURE)", s.config.TCP.Port)
		log.Printf("📜 Using certificate: %s", s.config.TCP.TLS.CertPath)
		log.Printf("🛡️  All connections will be encrypted")
		
		// Start certificate watcher for hot reload
		s.StartCertificateWatcher()
	} else if s.config.TCP.Enabled {
		listener, err = net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("failed to start TCP server: %w", err)
		}
		
		log.Printf("🟡 TCP Server listening on port %s (INSECURE)", s.config.TCP.Port)
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

		// Log connection type and update metrics
		if s.config.TCP.TLS.Enabled {
			if tlsConn, ok := conn.(*tls.Conn); ok {
				log.Printf("🔒 New TLS connection from %s", conn.RemoteAddr())
				// Log TLS details
				state := tlsConn.ConnectionState()
				log.Printf("   Cipher Suite: %s", tls.CipherSuiteName(state.CipherSuite))
				log.Printf("   TLS Version: %x", state.Version)
				
				// Update TLS metrics
				metrics.TLSConnections.Inc()
				defer metrics.TLSConnections.Dec()
				
				go s.handleConnection(tlsConn, true)
			}
		} else {
			log.Printf("🟡 New insecure TCP connection from %s", conn.RemoteAddr())
			
			// Update TCP metrics
			metrics.TCPConnections.Inc()
			defer metrics.TCPConnections.Dec()
			
			go s.handleConnection(conn, false)
		}
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn, isTLS bool) {
	defer conn.Close()
	
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
					if isTLS {
						// TLS auth error - no specific metric for now
					} else {
						metrics.TCPAuthErrors.Inc()
					}
					sendError(conn, "Invalid authentication")
					return
				}
				
				authenticated = true
				if isTLS {
					// TLS auth success - no specific metric for now
				} else {
					metrics.TCPAuthSuccess.Inc()
				}
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

// ReloadCertificates reloads TLS certificates without restarting the server
func (s *Server) ReloadCertificates() error {
	if !s.config.TCP.TLS.Enabled {
		return fmt.Errorf("TLS is not enabled")
	}

	// Load new certificates
	cert, err := tls.LoadX509KeyPair(s.config.TCP.TLS.CertPath, s.config.TCP.TLS.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to load new certificates: %w", err)
	}

	// Update certificate expiry metric
	if len(cert.Certificate) > 0 {
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			daysUntilExpiry := time.Until(x509Cert.NotAfter).Hours() / 24
			metrics.TLSCertificateExpiry.Set(daysUntilExpiry)
			log.Printf("🟢 Certificate reloaded - expires in %.0f days (%s)", daysUntilExpiry, x509Cert.NotAfter.Format("2006-01-02"))
		}
	}

	// Thread-safe certificate update
	s.certMutex.Lock()
	s.tlsConfig.Certificates = []tls.Certificate{cert}
	s.certMutex.Unlock()

	log.Println("✅ TLS certificates reloaded successfully")
	return nil
}

// StartCertificateWatcher monitors certificate files for changes
func (s *Server) StartCertificateWatcher() {
	if !s.config.TCP.TLS.Enabled {
		return
	}

	go func() {
		var lastModTime time.Time
		
		// Get initial modification time
		if stat, err := os.Stat(s.config.TCP.TLS.CertPath); err == nil {
			lastModTime = stat.ModTime()
		}

		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		for range ticker.C {
			if stat, err := os.Stat(s.config.TCP.TLS.CertPath); err == nil {
				if stat.ModTime().After(lastModTime) {
					log.Println("🔍 Certificate file changed, reloading...")
					if err := s.ReloadCertificates(); err != nil {
						log.Printf("🔴 Failed to reload certificates: %v", err)
					}
					lastModTime = stat.ModTime()
				}
			}
		}
	}()

	log.Println("🔍 Certificate watcher started - monitoring for changes every 30s")
} 