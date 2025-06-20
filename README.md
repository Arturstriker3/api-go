# GoMailer

A microservice for handling email sending through a RabbitMQ queue, built with Go.

## Features

- REST API for queuing emails
- RabbitMQ integration for reliable message queuing
- SMTP email sending with HTML support
- Environment-based configuration
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

   - Copy the example environment file:

   ```bash
   cp env.example .env
   ```

   - Edit the `.env` file with your configuration

4. Start RabbitMQ using Docker Compose:

```bash
docker-compose up -d
```

5. Run the application:

```bash
go run cmd/main.go
```

The service will start on the configured port (default: 8080).

## Environment Variables

### Required Variables

- `SMTP_USER`: SMTP server username (required)
- `SMTP_PASSWORD`: SMTP server password (required)
- `SMTP_FROM`: Email address to send from (required)

### Optional Variables with Defaults

- `SMTP_HOST`: SMTP server host (default: "smtp.gmail.com")
- `SMTP_PORT`: SMTP server port (default: 587)
- `RABBITMQ_HOST`: RabbitMQ host (default: "localhost")
- `RABBITMQ_PORT`: RabbitMQ port (default: "5672")
- `RABBITMQ_USER`: RabbitMQ username (default: "admin")
- `RABBITMQ_PASSWORD`: RabbitMQ password (default: "admin")
- `API_PORT`: API server port (default: "8080")

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

## Architecture

The service follows a clean architecture pattern with the following components:

- `cmd/main.go`: Application entry point
- `config/`: Configuration structures and environment handling
- `internal/api/`: HTTP API handlers
- `internal/email/`: Email sending service
- `internal/queue/`: RabbitMQ consumer implementation

## Error Handling

The service implements robust error handling:

- Environment variable validation
- Input validation for email requests
- Queue connection error handling
- SMTP sending error handling with message requeuing
- Graceful shutdown on system signals

## Development

To run the service in development mode:

1. Copy and configure environment variables:

```bash
cp env.example .env
# Edit .env with your settings
```

2. Start RabbitMQ:

```bash
docker-compose up -d
```

3. Run the service:

```bash
go run cmd/main.go
```

## Production Deployment

For production deployment:

1. Build the binary:

```bash
go build -o gomailer cmd/main.go
```

2. Set up environment variables in your production environment
3. Configure a process manager (e.g., systemd)
4. Set up proper monitoring and logging
5. Use a production-grade SMTP service

## License

MIT License
