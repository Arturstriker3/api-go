# GoMailer

A microservice for handling email sending through a RabbitMQ queue, built with Go.

## Features

- REST API for queuing emails
- RabbitMQ integration for reliable message queuing
- SMTP email sending with HTML support
- Configurable via environment variables
- Docker support for RabbitMQ

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- SMTP server credentials (e.g., Gmail SMTP)

## Setup

1. Clone the repository:

```bash
git clone <repository-url>
cd gomailer
```

2. Install dependencies:

```bash
go mod download
```

3. Set up environment variables:

```bash
export SMTP_USER=your-email@example.com
export SMTP_PASSWORD=your-smtp-password
export SMTP_FROM=your-email@example.com
```

4. Start RabbitMQ using Docker Compose:

```bash
docker-compose up -d
```

5. Run the application:

```bash
go run cmd/main.go
```

The service will start on port 8080 by default.

## API Usage

### Queue an Email

```http
POST /email
Content-Type: application/json

{
  "to": ["recipient@example.com"],
  "subject": "Hello",
  "body": "<h1>Hello World</h1><p>This is a test email.</p>"
}
```

Response:

```json
{
  "message": "Email queued successfully"
}
```

## Configuration

The service can be configured through environment variables:

- `SMTP_USER`: SMTP server username
- `SMTP_PASSWORD`: SMTP server password
- `SMTP_FROM`: Email address to send from

Default configuration (can be modified in config/config.go):

- SMTP Server: smtp.gmail.com:587
- RabbitMQ: localhost:5672 (credentials: admin/admin)
- API Port: 8080

## Architecture

The service follows a clean architecture pattern with the following components:

- `cmd/main.go`: Application entry point
- `config/`: Configuration structures
- `internal/api/`: HTTP API handlers
- `internal/email/`: Email sending service
- `internal/queue/`: RabbitMQ consumer implementation

## Error Handling

The service implements robust error handling:

- Input validation for email requests
- Queue connection error handling
- SMTP sending error handling with message requeuing
- Graceful shutdown on system signals

## Development

To run the service in development mode:

1. Start RabbitMQ:

```bash
docker-compose up -d
```

2. Run the service:

```bash
go run cmd/main.go
```

## Production Deployment

For production deployment:

1. Build the binary:

```bash
go build -o gomailer cmd/main.go
```

2. Set up environment variables
3. Configure a process manager (e.g., systemd)
4. Set up proper monitoring and logging
5. Use a production-grade SMTP service

## License

MIT License
