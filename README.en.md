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
- `TCP_PORT`: TCP/TLS server port (default: "9000")
- `TCP_ENABLED`: Enable plain TCP (default: "true")
- `TCP_TLS_ENABLED`: Enable secure TLS (default: "false")
- `TCP_TLS_CERT_PATH`: TLS certificate path (default: "certs/server.crt")
- `TCP_TLS_KEY_PATH`: TLS private key path (default: "certs/server.key")
- `TCP_TLS_CA_PATH`: CA certificate path (default: "certs/ca-cert.pem")
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
go get github.com/Arturstriker3/api-go
```

2. Configure environment variables in your service:

For **plain TCP** (development):

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=your-secret-here
```

For **secure TLS** (production):

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=your-secret-here
GOMAILER_TLS_ENABLED=true
GOMAILER_REJECT_UNAUTHORIZED=false
GOMAILER_CA_PATH=certs/ca-cert.pem
```

**📁 Example files available:**

- `tcp.example` - Plain TCP configuration
- `tls.example` - Secure TLS configuration

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

## TLS Integration (Recommended for Production)

For secure connections with TLS encryption, follow these steps:

### 1. Generate TLS Certificates

```bash
# Generate self-signed certificates for development
go run scripts/generate-certs.go
```

### 2. Configure TLS Server

Set environment variables:

```env
# Disable plain TCP
TCP_ENABLED=false

# Enable secure TLS
TCP_TLS_ENABLED=true
TCP_TLS_CERT_PATH=certs/server.crt
TCP_TLS_KEY_PATH=certs/server.key
TCP_TLS_CA_PATH=certs/ca-cert.pem
```

### 3. Node.js Client with TLS

```javascript
const tls = require("tls");
const fs = require("fs");

const options = {
  host: process.env.GOMAILER_HOST,
  port: process.env.GOMAILER_PORT,
  rejectUnauthorized: process.env.NODE_ENV === "production",
  ca: process.env.GOMAILER_CA_PATH
    ? [fs.readFileSync(process.env.GOMAILER_CA_PATH)]
    : undefined,
};

const client = tls.connect(options, () => {
  console.log("🔒 TLS connection established");
  console.log("Authorized:", client.authorized);
  console.log("Cipher:", client.getCipher().name);

  // Send authentication (encrypted)
  const auth = { secret: process.env.GOMAILER_AUTH_SECRET };
  client.write(JSON.stringify(auth));

  // Send email (encrypted)
  const email = {
    to: ["recipient@example.com"],
    subject: "Secure TLS Email",
    body: "<h1>🔒 This message was sent securely via TLS</h1>",
  };
  client.write(JSON.stringify(email));
});

client.on("data", (data) => {
  console.log("📥 Encrypted response:", JSON.parse(data.toString()));
  client.destroy();
});
```

### 4. NestJS Integration with TLS

```typescript
// src/services/gomailer-tls.service.ts
import { Injectable, OnModuleInit } from "@nestjs/common";
import * as tls from "tls";
import * as fs from "fs";

@Injectable()
export class GomailerTLSService implements OnModuleInit {
  private client: tls.TLSSocket;
  private connected: boolean = false;

  private connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      const options = {
        host: process.env.GOMAILER_HOST || "localhost",
        port: parseInt(process.env.GOMAILER_PORT || "9000"),
        rejectUnauthorized: process.env.NODE_ENV === "production",
        ca: process.env.GOMAILER_CA_PATH
          ? [fs.readFileSync(process.env.GOMAILER_CA_PATH)]
          : undefined,
      };

      this.client = tls.connect(options, () => {
        console.log("🔒 TLS connection established");
        this.connected = true;

        // Send encrypted authentication
        const auth = { secret: process.env.GOMAILER_AUTH_SECRET };
        this.client.write(JSON.stringify(auth));
        resolve();
      });

      this.client.on("error", (error) => {
        this.connected = false;
        reject(error);
      });
    });
  }

  async sendEmail(request: EmailRequest): Promise<void> {
    if (!this.connected) {
      await this.connect();
    }

    // Data sent encrypted
    return new Promise((resolve, reject) => {
      this.client.write(JSON.stringify(request));
      // ... rest of implementation
    });
  }
}
```

### TCP vs TLS Comparison

| Aspect           | Plain TCP     | TLS                 |
| ---------------- | ------------- | ------------------- |
| **Encryption**   | ❌ None       | ✅ AES-256          |
| **Auth Secret**  | ⚠️ Plain text | ✅ Encrypted        |
| **Interception** | ❌ Vulnerable | 🛡️ Protected        |
| **Performance**  | 🟢 Fast       | 🟡 Minimal overhead |
| **Setup**        | 🟢 Simple     | 🟡 Certificates     |
| **Production**   | ❌ Insecure   | ✅ Recommended      |

### Security Configurations

#### Development

```env
TCP_ENABLED=true
TCP_TLS_ENABLED=false
```

#### Production (Recommended)

```env
TCP_ENABLED=false
TCP_TLS_ENABLED=true
```

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
