# 🔄 GoMailer Certificate Auto-Renewal System

Automatic self-signed certificate renewal system for production **without downtime**.

## 📋 **Available Scripts**

### **1. Initial Generation**

```bash
# Generate self-signed certificates for development
go run -tags generate_certs scripts/generate-self-signed-certs.go
```

### **2. Automatic Renewal**

```bash
# Check and renew certificates (if needed)
go run -tags renew_certs scripts/auto-renew-certs.go
```

### **3. Automation via Cron (Linux/Mac)**

```bash
# Make executable
chmod +x scripts/cert-renewal-cron.sh

# Add to crontab (runs daily at 2 AM)
crontab -e
# Add line:
0 2 * * * /path/to/gomailer/scripts/cert-renewal-cron.sh
```

### **4. Automation via Task Scheduler (Windows)**

```powershell
# Run manually
PowerShell -ExecutionPolicy Bypass -File "scripts\cert-renewal-task.ps1"

# Or configure in Task Scheduler:
# - Trigger: Daily at 2:00 AM
# - Action: PowerShell -ExecutionPolicy Bypass -File "C:\path\to\gomailer\scripts\cert-renewal-task.ps1"
```

## ⚙️ **How It Works**

### **🔍 Automatic Verification**

- ✅ Checks if certificates exist
- ✅ Calculates days until expiration
- ✅ Automatically renews if < 30 days
- ✅ Keeps valid certificates if > 30 days

### **🔄 Zero Downtime Renewal**

1. **Secure Generation**: New certificates in temporary directory
2. **Automatic Backup**: Old certificates saved in `certs/backup_TIMESTAMP/`
3. **Atomic Replacement**: Instant file swap
4. **Zero Downtime**: API continues working during the process

### **📁 File Structure**

```
certs/
├── server.crt          # Current certificate
├── server.key          # Current private key
├── ca-cert.pem         # CA certificate
├── backup_1234567890/  # Automatic backup
│   ├── server.crt
│   ├── server.key
│   └── ca-cert.pem
└── temp/               # Temporary directory (removed after use)
```

## 🚀 **Production Configuration**

### **Option 1: Cron Job (Linux)**

```bash
# Edit crontab
crontab -e

# Add (runs daily at 2 AM)
0 2 * * * /path/to/gomailer/scripts/cert-renewal-cron.sh

# Check logs
tail -f /path/to/gomailer/logs/cert-renewal.log
```

### **Option 2: Windows Task Scheduler**

1. Open **Task Scheduler**
2. **Create Basic Task**
3. **Name**: GoMailer Certificate Renewal
4. **Trigger**: Daily at 2:00 AM
5. **Action**: Start a program
   - **Program**: `PowerShell`
   - **Arguments**: `-ExecutionPolicy Bypass -File "C:\path\to\gomailer\scripts\cert-renewal-task.ps1"`

### **Option 3: Docker Cron**

```dockerfile
# Add to Dockerfile
RUN apt-get update && apt-get install -y cron
COPY scripts/cert-renewal-cron.sh /etc/cron.daily/gomailer-certs
RUN chmod +x /etc/cron.daily/gomailer-certs
```

### **Docker Setup Requirements**

Before running with Docker, create a `.env` file based on `env.example`:

```bash
# Copy the example
cp env.example .env

# Edit with your configurations
nano .env
```

**Required variables for Docker:**

```env
# SMTP Configuration
SMTP_HOST=your-smtp-host
SMTP_PORT=587
SMTP_USER=your-email@domain.com
SMTP_PASSWORD=your-password
SMTP_FROM=your-email@domain.com

# RabbitMQ Configuration
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=admin

# TCP/TLS Configuration
TCP_PORT=9000
TCP_AUTH_SECRET=your-secret-key
TCP_ENABLED=false
TCP_TLS_ENABLED=true
TCP_TLS_CERT_PATH=certs/server.crt
TCP_TLS_KEY_PATH=certs/server.key
TCP_TLS_CA_PATH=certs/ca-cert.pem

# Metrics
METRICS_PORT=9091
```

### **Option 4: Kubernetes CronJob**

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: gomailer-cert-renewal
spec:
  schedule: "0 2 * * *" # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: cert-renewal
              image: gomailer:latest
              command:
                [
                  "go",
                  "run",
                  "-tags",
                  "renew_certs",
                  "scripts/auto-renew-certs.go",
                ]
          restartPolicy: OnFailure
```

## 📊 **Monitoring**

### **Prometheus Metrics**

- `gomailer_tls_certificate_expiry_days` - Days until expiration

### **Logs**

```bash
# View renewal logs
tail -f logs/cert-renewal.log

# Check last execution
ls -la certs/backup_*
```

### **Grafana Alerts**

Configure alerts when:

- Certificate expires in < 7 days
- Renewal fails
- Renewal process didn't execute

## 🔧 **Advanced Configuration**

### **Change Renewal Threshold**

```go
// In scripts/auto-renew-certs.go, line 45
renewThreshold := 30.0  // Change to desired days
```

### **Configure Custom Domains**

```go
// In scripts/auto-renew-certs.go, line 88
DNSNames: []string{"localhost", "mydomain.com"},
```

### **API Hot Reload**

To reload certificates without restarting:

```go
// Implement in internal/tcp/server.go
func (s *Server) ReloadCertificates() error {
    // Reload TLS certificates
    cert, err := tls.LoadX509KeyPair(s.config.TCP.TLS.CertPath, s.config.TCP.TLS.KeyPath)
    if err != nil {
        return err
    }

    // Update TLS configuration
    s.tlsConfig.Certificates = []tls.Certificate{cert}
    log.Println("🔄 TLS certificates reloaded successfully")
    return nil
}
```

## ⚠️ **Security Considerations**

### **✅ Self-Signed Advantages**

- ✅ **Full Control**: You manage the renewal
- ✅ **No Dependencies**: Doesn't depend on external services
- ✅ **Automatic Renewal**: Own renewal system
- ✅ **Zero Downtime**: Seamless replacement

### **⚠️ Limitations**

- ⚠️ **Browsers**: Show untrusted certificate warning
- ⚠️ **Clients**: Need to configure to accept self-signed certificates
- ⚠️ **Public Production**: Not recommended for public APIs

### **🔒 For Public Production**

If you need publicly trusted certificates:

1. Use **Let's Encrypt** with Certbot
2. Use **AWS Certificate Manager**
3. Use **Cloudflare SSL**
4. Use corporate certificates

## 🎯 **Summary**

This system allows using **self-signed certificates in production** with:

- ✅ **Automatic renewal** (30 days before expiration)
- ✅ **Zero downtime** (atomic replacement)
- ✅ **Automatic backup** (rollback if needed)
- ✅ **Complete logs** (audit and debug)
- ✅ **Multi-platform** (Linux, Windows, Docker, Kubernetes)

**Ideal for**: Internal APIs, microservices, corporate environments where full control is more important than public certificate trust.

## 🔗 **Client Connection (NestJS Example)**

### **What NestJS Needs:**

- ✅ **Only `ca-cert.pem`** - To validate the server
- ❌ **No certificate generation** needed
- ❌ **No server certificates** required

### **Simple NestJS Connection:**

```typescript
import * as tls from "tls";
import * as fs from "fs";

// Connect to GoMailer
const client = tls.connect({
  host: "localhost",
  port: 9000,
  // ONLY CA cert to validate server
  ca: fs.readFileSync("certs/ca-cert.pem"),
  rejectUnauthorized: true, // Validate server
});

client.on("secureConnect", () => {
  console.log("🔒 Connected to GoMailer via TLS");

  // 1. Authenticate
  client.write(JSON.stringify({ secret: "your-secret" }));

  // 2. Send email
  client.write(
    JSON.stringify({
      to: "user@example.com",
      subject: "Test",
      body: "Hello from NestJS!",
    })
  );
});
```

### **Certificate Renewal Impact:**

- ✅ **Old `ca-cert.pem` continues working** (most cases)
- ✅ **NestJS stays online** during GoMailer renewal
- ✅ **Hot reload** - GoMailer automatically reloads certificates
- ⚠️ **Rare case**: If CA changes completely, copy new `ca-cert.pem`

### **NestJS File Structure:**

```
nestjs-app/
├── certs/
│   └── ca-cert.pem     # ← ONLY this file (copied from GoMailer)
├── src/
│   └── gomailer.service.ts
└── .env
```

**It's like HTTPS**: The website has certificates, your browser only validates! 🌐🔒
