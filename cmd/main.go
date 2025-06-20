package main

import (
	"gomailer/config"
	"gomailer/internal/api"
	"gomailer/internal/email"
	"gomailer/internal/queue"
	"gomailer/internal/tcp"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Initialize API handler
	handler, err := api.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create API handler: %v", err)
	}

	// Initialize TCP server
	tcpServer, err := tcp.NewServer(cfg, emailService)
	if err != nil {
		log.Fatalf("Failed to create TCP server: %v", err)
	}

	// Setup router
	router := api.SetupRouter(handler)

	// Start the servers in goroutines
	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.API.Port)
		if err := router.Run(":" + cfg.API.Port); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

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