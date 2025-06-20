package main

import (
	"gomailer/config"
	"gomailer/internal/api"
	"gomailer/internal/email"
	"gomailer/internal/queue"
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

	// Setup router
	router := api.SetupRouter(handler)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.API.Port)
		if err := router.Run(":" + cfg.API.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
} 