package main

import (
	"gomailer/config"
	"gomailer/internal/email"
	"gomailer/internal/queue"
	"gomailer/internal/tcp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize email service
	emailService := email.NewEmailService(cfg)

	// Initialize consumer
	consumer, err := queue.NewConsumer(cfg, emailService)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Setup(); err != nil {
		log.Fatalf("Failed to setup consumer: %v", err)
	}

	if err := consumer.StartConsuming(); err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	// Initialize TCP server
	tcpServer, err := tcp.NewServer(cfg, emailService)
	if err != nil {
		log.Fatalf("Failed to create TCP server: %v", err)
	}

	// Start metrics server in a separate goroutine
	go func() {
		metricsPort := cfg.Metrics.Port
		if metricsPort == "" {
			metricsPort = "9091" // Default metrics port
		}
		log.Printf("Starting metrics server on port %s", metricsPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Start TCP server
	go func() {
		log.Printf("Starting TCP server on port %s", cfg.TCP.Port)
		if err := tcpServer.Start(); err != nil {
			log.Fatalf("Failed to start TCP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
	
	// Graceful shutdown
	if err := tcpServer.Stop(); err != nil {
		log.Printf("Error stopping TCP server: %v", err)
	}
} 