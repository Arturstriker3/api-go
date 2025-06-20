package main

import (
	"fmt"
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
	cfg := config.NewConfig()

	// Set SMTP credentials from environment variables
	cfg.SMTP.User = os.Getenv("SMTP_USER")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	cfg.SMTP.From = os.Getenv("SMTP_FROM")

	if cfg.SMTP.User == "" || cfg.SMTP.Password == "" || cfg.SMTP.From == "" {
		log.Fatal("SMTP credentials not set. Please set SMTP_USER, SMTP_PASSWORD, and SMTP_FROM environment variables")
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
		if err := router.Run(":" + cfg.API.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on port %s", cfg.API.Port)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
} 