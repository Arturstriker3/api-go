# GoMailer

<div align="center">

[🇧🇷 Português](README.md) | [🇺🇸 English](#english)

</div>

# English

A microservice for handling email sending through a RabbitMQ queue, built with Go. Provides a secure TCP interface for service integration.

## Features

- TCP server for service integration
- RabbitMQ integration for reliable message queuing
- SMTP email sending with HTML support
- Environment-based configuration
- Docker support for RabbitMQ
- Prometheus metrics and Grafana dashboards

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

4. Start the infrastructure using Docker Compose:

```bash
docker-compose up -d
```

5. Run the application:

```bash
go run cmd/main.go
```

The service will start the TCP server on port 9000 (default) and metrics on port 9091.

## Environment Variables

### Required Variables

- `SMTP_USER`: SMTP server username (required)
- `SMTP_PASSWORD`: SMTP server password (required)
- `SMTP_FROM`: Email address to send from (required)
- `TCP_AUTH_SECRET`: Secret key for TCP authentication (required)

### Optional Variables with Defaults

- `SMTP_HOST`: SMTP server host (default: "smtp.gmail.com")
- `SMTP_PORT`: SMTP server port (default: 587)
- `RABBITMQ_HOST`: RabbitMQ host (default: "localhost")
- `RABBITMQ_PORT`: RabbitMQ port (default: "5672")
- `RABBITMQ_USER`: RabbitMQ username (default: "admin")
- `RABBITMQ_PASSWORD`: RabbitMQ password (default: "admin")
- `TCP_PORT`: TCP server port (default: "9000")
- `METRICS_PORT`: Prometheus metrics port (default: "9091")

## TCP Integration

To integrate other services with GoMailer, you can use the provided TCP client:

```go
package main

import (
    "log"
    "os"
    "gomailer/pkg/client"
)

func main() {
    // Create email client
    emailClient := client.NewEmailClient(
        os.Getenv("GOMAILER_HOST"),     // Service host
        os.Getenv("GOMAILER_PORT"),     // TCP port
        os.Getenv("GOMAILER_AUTH_SECRET"), // Authentication key
    )

    // Prepare email request
    request := &client.EmailRequest{
        To:      []string{"recipient@example.com"},
        Subject: "TCP Test",
        Body:    "<h1>Hello</h1><p>This is a TCP test</p>",
    }

    // Send email
    if err := emailClient.SendEmail(request); err != nil {
        log.Fatalf("Error sending email: %v", err)
    }
}
```

To use the client in another project:

1. Add GoMailer as a dependency:

```bash
go get github.com/Arturstriker3/gomailer
```

2. Configure environment variables in your service:

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=your-secret-here
```

### NestJS Integration Example

Here's how to integrate GoMailer in a NestJS application:

1. Create a TCP client service:

```typescript
// src/services/gomailer.service.ts
import { Injectable, OnModuleInit } from "@nestjs/common";
import { Socket } from "net";

interface EmailRequest {
  to: string[];
  subject: string;
  body: string;
}

@Injectable()
export class GomailerService implements OnModuleInit {
  private client: Socket;
  private connected: boolean = false;

  constructor() {
    this.client = new Socket();
  }

  async onModuleInit() {
    await this.connect();
  }

  private connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client.connect(
        {
          host: process.env.GOMAILER_HOST || "localhost",
          port: parseInt(process.env.GOMAILER_PORT || "9000"),
        },
        () => {
          this.connected = true;
          // Send authentication
          const auth = { secret: process.env.GOMAILER_AUTH_SECRET };
          this.client.write(JSON.stringify(auth));
          resolve();
        }
      );

      this.client.on("error", (error) => {
        this.connected = false;
        reject(error);
      });

      this.client.on("close", () => {
        this.connected = false;
      });
    });
  }

  async sendEmail(request: EmailRequest): Promise<void> {
    if (!this.connected) {
      await this.connect();
    }

    return new Promise((resolve, reject) => {
      this.client.write(JSON.stringify(request));

      this.client.once("data", (data) => {
        const response = JSON.parse(data.toString());
        if (response.error) {
          reject(new Error(response.error));
        } else {
          resolve();
        }
      });
    });
  }

  onModuleDestroy() {
    if (this.client) {
      this.client.destroy();
    }
  }
}
```

2. Register the service in your module:

```typescript
// src/app.module.ts
import { Module } from "@nestjs/common";
import { ConfigModule } from "@nestjs/config";
import { GomailerService } from "./services/gomailer.service";

@Module({
  imports: [
    ConfigModule.forRoot(), // For environment variables
  ],
  providers: [GomailerService],
  exports: [GomailerService],
})
export class AppModule {}
```

3. Use the service in your controllers/services:

```typescript
// src/controllers/email.controller.ts
import { Controller, Post, Body } from "@nestjs/common";
import { GomailerService } from "../services/gomailer.service";

@Controller("email")
export class EmailController {
  constructor(private readonly gomailerService: GomailerService) {}

  @Post()
  async sendEmail(
    @Body() emailData: { to: string[]; subject: string; body: string }
  ) {
    try {
      await this.gomailerService.sendEmail(emailData);
      return { message: "Email queued successfully" };
    } catch (error) {
      throw new Error(`Failed to send email: ${error.message}`);
    }
  }
}
```

4. Configure environment variables in your `.env`:

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=your-secret-here
```

The NestJS service handles:

- Automatic connection management
- Authentication with GoMailer
- Reconnection on failures
- Clean shutdown
- Type safety with TypeScript

## Monitoring

The service exposes Prometheus metrics and includes a pre-configured Grafana dashboard:

- Prometheus metrics: http://localhost:9091/metrics
- Grafana dashboard: http://localhost:3000 (default credentials: admin/admin)

The dashboard includes:

- Email queue and sending rates
- Queue size and processing latency
- TCP connection metrics
- Error rates

## Architecture

The service follows a clean architecture pattern with the following components:

- `cmd/main.go`: Application entry point
- `config/`: Configuration structures and environment handling
- `internal/email/`: Email sending service
- `internal/queue/`: RabbitMQ consumer implementation
- `internal/tcp/`: TCP server for service integration
- `pkg/client/`: TCP client for external integration

## Error Handling

The service implements robust error handling:

- Environment variable validation
- Queue connection error handling
- SMTP sending error handling with message requeuing
- TCP connection authentication and validation
- Graceful shutdown on system signals

## Development

To run the service in development mode:

1. Copy and configure environment variables:

```bash
cp env.example .env
# Edit .env with your settings
```

2. Start the infrastructure:

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
6. Configure firewalls to allow only trusted TCP connections

## License

MIT License
