package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arturstriker3/api-go/config"
	"github.com/Arturstriker3/api-go/internal/email"
	"github.com/Arturstriker3/api-go/internal/queue"
	"github.com/Arturstriker3/api-go/internal/tcp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CertificateNotification struct {
	Action          string `json:"action"`
	Timestamp       string `json:"timestamp"`
	CertificatePath string `json:"certificate_path"`
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("🔴 Failed to load configuration: %v", err)
	}

	// Initialize email service
	emailService := email.NewEmailService(cfg)

	// Initialize certificate email service
	certEmailService := email.NewCertificateEmailService(emailService)

	// Initialize consumer
	consumer, err := queue.NewConsumer(cfg, emailService)
	if err != nil {
		log.Fatalf("🔴 Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Setup(); err != nil {
		log.Fatalf("🔴 Failed to setup consumer: %v", err)
	}

	if err := consumer.StartConsuming(); err != nil {
		log.Fatalf("🔴 Failed to start consuming: %v", err)
	}

	// Initialize TCP server
	tcpServer, err := tcp.NewServer(cfg, emailService)
	if err != nil {
		log.Fatalf("🔴 Failed to create TCP server: %v", err)
	}

	// Start certificate notification watcher
	go startCertificateNotificationWatcher(certEmailService)

	// Start metrics server in a separate goroutine
	go func() {
		metricsPort := cfg.Metrics.Port
		if metricsPort == "" {
			metricsPort = "9091" // Default metrics port
		}
		log.Printf("🟡 Starting metrics server on port %s", metricsPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			log.Printf("🔴 Metrics server error: %v", err)
		}
	}()

	// Start TCP server
	go func() {
		log.Printf("🟢 Starting TCP server on port %s", cfg.TCP.Port)
		if err := tcpServer.Start(); err != nil {
			log.Fatalf("🔴 Failed to start TCP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🟡 Shutting down servers...")
	
	// Graceful shutdown
	if err := tcpServer.Stop(); err != nil {
		log.Printf("🔴 Error stopping TCP server: %v", err)
	}
}

// startCertificateNotificationWatcher monitors for certificate notifications and sends emails
func startCertificateNotificationWatcher(certEmailService *email.CertificateEmailService) {
	log.Println("🔍 Certificate notification watcher started")
	
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()
	
	var lastProcessed time.Time

	for range ticker.C {
		notificationFile := "certs/certificate_notification.json"
		
		// Check if notification file exists
		stat, err := os.Stat(notificationFile)
		if err != nil {
			continue // No notification file
		}
		
		// Check if this is a new notification
		if stat.ModTime().After(lastProcessed) {
			// Read notification
			data, err := os.ReadFile(notificationFile)
			if err != nil {
				log.Printf("🔴 Error reading certificate notification: %v", err)
				continue
			}
			
			var notification CertificateNotification
			if err := json.Unmarshal(data, &notification); err != nil {
				log.Printf("🔴 Error parsing certificate notification: %v", err)
				continue
			}
			
			log.Printf("🟡 Processing certificate notification: %s", notification.Action)
			
			// Send certificate email
			if err := certEmailService.SendCertificateEmail(notification.Action); err != nil {
				log.Printf("🟡 Warning: Could not send certificate email: %v", err)
			} else {
				log.Printf("✅ Certificate email sent successfully for action: %s", notification.Action)
			}
			
			// Update last processed time
			lastProcessed = stat.ModTime()
			
			// Optionally remove the notification file after processing
			if err := os.Remove(notificationFile); err != nil {
				log.Printf("🟡 Warning: Could not remove notification file: %v", err)
			}
		}
	}
} 