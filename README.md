# Email Microservice

A Go-based microservice for handling email sending operations with RabbitMQ integration.

## Features
- Email sending service
- RabbitMQ queue integration
- REST API endpoints
- Scalable microservice architecture

## Prerequisites
- Go 1.21 or later
- RabbitMQ server
- SMTP server configuration

## Project Structure
```
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   └── routes/
│   ├── config/
│   ├── email/
│   └── queue/
├── pkg/
│   ├── models/
│   └── utils/
├── go.mod
└── README.md
```

## Setup Instructions
1. Install Go from https://golang.org/dl/
2. Clone this repository
3. Install dependencies: `go mod tidy`
4. Set up environment variables (see `.env.example`)
5. Run the service: `go run cmd/api/main.go`

## API Endpoints
- POST /api/v1/email - Queue a new email
- GET /api/v1/status - Check service status

## Environment Variables
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port
- `SMTP_USERNAME` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `RABBITMQ_URL` - RabbitMQ connection URL
- `API_PORT` - API server port 