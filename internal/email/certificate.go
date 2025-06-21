package email

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// CertificateEmailService handles sending certificate emails using the existing email infrastructure
type CertificateEmailService struct {
	emailService *Service
}

// NewCertificateEmailService creates a new certificate email service
func NewCertificateEmailService(emailService *Service) *CertificateEmailService {
	return &CertificateEmailService{
		emailService: emailService,
	}
}

// SendCertificateEmail sends the CA certificate via the existing email infrastructure
func (c *CertificateEmailService) SendCertificateEmail(action string) error {
	// Check if we should send email (Docker + recipient configured)
	recipient := os.Getenv("CERTIFICATE_EMAIL_RECIPIENT")
	if recipient == "" {
		recipient = os.Getenv("SMTP_USER") // Fallback to SMTP_USER
	}
	
	if !c.isDockerEnvironment() || recipient == "" {
		return fmt.Errorf("certificate email not configured (need CERTIFICATE_EMAIL_RECIPIENT or SMTP_USER in Docker)")
	}

	// Read CA certificate
	caCert, err := os.ReadFile("certs/ca-cert.pem")
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}

	// Create email content based on action
	subject, body := c.createEmailContent(string(caCert), action)

	// Create email request using the existing structure
	emailData := &EmailData{
		To:      []string{recipient},
		Subject: subject,
		Body:    body,
	}

	// Send via existing email infrastructure
	return c.emailService.SendEmail(emailData)
}

// isDockerEnvironment checks if running inside Docker
func (c *CertificateEmailService) isDockerEnvironment() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// createEmailContent creates the email subject and body with improved formatting
func (c *CertificateEmailService) createEmailContent(caCert, action string) (string, string) {
	now := time.Now()
	expiryDate := now.Add(365 * 24 * time.Hour)
	
	var subject string
	var headerIcon, actionText, statusBadge string
	
	switch action {
	case "RENEWED":
		subject = "🔄 GoMailer TLS Certificate Renewed"
		headerIcon = "🔄"
		actionText = "renewed"
		statusBadge = "RENEWED"
	default:
		subject = "🔐 GoMailer TLS Certificate Generated"
		headerIcon = "🔐"
		actionText = "generated"
		statusBadge = "NEW"
	}

	// Clean and format certificate
	cleanCert := strings.TrimSpace(caCert)
	
	body := c.buildEmailBody(EmailTemplateData{
		HeaderIcon:    headerIcon,
		StatusBadge:   statusBadge,
		ActionText:    actionText,
		Certificate:   cleanCert,
		GeneratedDate: now.Format("2006-01-02 15:04:05"),
		ExpiryDate:    expiryDate.Format("2006-01-02 15:04:05"),
		Action:        action,
	})

	return subject, body
}

// EmailTemplateData holds the data for email template
type EmailTemplateData struct {
	HeaderIcon    string
	StatusBadge   string
	ActionText    string
	Certificate   string
	GeneratedDate string
	ExpiryDate    string
	Action        string
}

// buildEmailBody creates a clean, professional email body
func (c *CertificateEmailService) buildEmailBody(data EmailTemplateData) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>GoMailer TLS Certificate</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 20px;">
    
    <!-- Header -->
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; border-radius: 10px; text-align: center; margin-bottom: 30px;">
        <h1 style="margin: 0; font-size: 28px;">%s GoMailer TLS Certificate</h1>
        <p style="margin: 10px 0 0 0; font-size: 16px; opacity: 0.9;">Certificate has been successfully %s</p>
        <div style="background: rgba(255,255,255,0.2); display: inline-block; padding: 8px 16px; border-radius: 20px; margin-top: 15px; font-weight: bold; font-size: 14px;">
            STATUS: %s
        </div>
    </div>

    <!-- Certificate Info -->
    <div style="background: #f8f9fa; border-left: 4px solid #28a745; padding: 20px; margin-bottom: 25px; border-radius: 5px;">
        <h2 style="margin-top: 0; color: #28a745;">📋 Certificate Information</h2>
        <table style="width: 100%%; border-collapse: collapse;">
            <tr>
                <td style="padding: 8px 0; font-weight: bold; width: 30%%;">Generated:</td>
                <td style="padding: 8px 0;">%s</td>
            </tr>
            <tr>
                <td style="padding: 8px 0; font-weight: bold;">Expires:</td>
                <td style="padding: 8px 0;">%s</td>
            </tr>
            <tr>
                <td style="padding: 8px 0; font-weight: bold;">Validity:</td>
                <td style="padding: 8px 0;">1 Year</td>
            </tr>
            <tr>
                <td style="padding: 8px 0; font-weight: bold;">Organization:</td>
                <td style="padding: 8px 0;">GoMailer</td>
            </tr>
            <tr>
                <td style="padding: 8px 0; font-weight: bold;">DNS Names:</td>
                <td style="padding: 8px 0;">localhost, gomailer, *.gomailer.local</td>
            </tr>
        </table>
    </div>

    <!-- Quick Setup -->
    <div style="background: #e3f2fd; border-left: 4px solid #2196f3; padding: 20px; margin-bottom: 25px; border-radius: 5px;">
        <h2 style="margin-top: 0; color: #1976d2;">⚡ Quick Setup</h2>
        <ol style="margin: 0; padding-left: 20px;">
            <li style="margin-bottom: 8px;"><strong>Save the certificate</strong> below as <code style="background: #fff; padding: 2px 6px; border-radius: 3px; color: #d63384;">ca-cert.pem</code></li>
            <li style="margin-bottom: 8px;"><strong>Replace</strong> your existing certificate file</li>
            <li style="margin-bottom: 8px;"><strong>Configure</strong> your client application to use this CA certificate</li>
            <li><strong>Restart</strong> your client application</li>
        </ol>
    </div>

    <!-- Certificate Content -->
    <div style="margin-bottom: 25px;">
        <h2 style="color: #495057;">📄 CA Certificate (ca-cert.pem)</h2>
        <div style="background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 5px; padding: 0;">
            <div style="background: #e9ecef; padding: 10px; border-bottom: 1px solid #dee2e6; font-weight: bold; color: #495057;">
                ca-cert.pem
                <button style="float: right; background: #007bff; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer; font-size: 12px;" onclick="copyToClipboard()">📋 Copy</button>
            </div>
            <pre id="certificate" style="margin: 0; padding: 15px; overflow-x: auto; font-family: 'Courier New', monospace; font-size: 12px; line-height: 1.4; background: #ffffff; white-space: pre-wrap; word-wrap: break-word;">%s</pre>
        </div>
    </div>

    <!-- Client Examples -->
    <div style="background: #fff3cd; border-left: 4px solid #ffc107; padding: 20px; margin-bottom: 25px; border-radius: 5px;">
        <h2 style="margin-top: 0; color: #856404;">💻 Client Integration Examples</h2>
        
        <h3 style="color: #856404; margin-bottom: 10px;">NestJS/Node.js</h3>
        <pre style="background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 13px; margin-bottom: 15px;"><code>import * as fs from 'fs';
import * as tls from 'tls';

const caCert = fs.readFileSync('ca-cert.pem');
const socket = tls.connect({
  host: 'localhost',
  port: 9001,
  ca: [caCert],
  rejectUnauthorized: true
});</code></pre>

        <h3 style="color: #856404; margin-bottom: 10px;">Go Client</h3>
        <pre style="background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 13px;"><code>caCert, _ := ioutil.ReadFile("ca-cert.pem")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

conn, err := tls.Dial("tcp", "localhost:9001", &tls.Config{
    RootCAs: caCertPool,
})</code></pre>
    </div>

    %s

    <!-- Footer -->
    <div style="background: #f8f9fa; padding: 20px; border-radius: 5px; text-align: center; color: #6c757d; margin-top: 30px;">
        <p style="margin: 0; font-size: 14px;">
            <strong>GoMailer Certificate System</strong><br>
            This is an automated message. Please keep this certificate secure and do not share it publicly.
        </p>
    </div>

    <script>
    function copyToClipboard() {
        const cert = document.getElementById('certificate');
        const textArea = document.createElement('textarea');
        textArea.value = cert.textContent;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        alert('Certificate copied to clipboard!');
    }
    </script>

</body>
</html>`, 
		data.HeaderIcon,
		data.ActionText,
		data.StatusBadge,
		data.GeneratedDate,
		data.ExpiryDate,
		data.Certificate,
		c.getActionAlert(data.Action))
}

// getActionAlert returns action-specific alert message
func (c *CertificateEmailService) getActionAlert(action string) string {
	switch action {
	case "RENEWED":
		return `    <!-- Renewal Alert -->
    <div style="background: #f8d7da; border-left: 4px solid #dc3545; padding: 20px; margin-bottom: 25px; border-radius: 5px;">
        <h2 style="margin-top: 0; color: #721c24;">⚠️ Important: Certificate Renewal</h2>
        <p style="margin-bottom: 15px; color: #721c24;">
            <strong>This is an automatic certificate renewal.</strong> Your previous certificate will expire soon.
        </p>
        <div style="background: #fff; padding: 15px; border-radius: 5px; border: 1px solid #f5c6cb;">
            <p style="margin: 0; color: #721c24;">
                <strong>Action Required:</strong> Please update your client applications with this new certificate 
                as soon as possible to avoid connection issues.
            </p>
        </div>
    </div>`
	default:
		return `    <!-- Welcome Message -->
    <div style="background: #d4edda; border-left: 4px solid #28a745; padding: 20px; margin-bottom: 25px; border-radius: 5px;">
        <h2 style="margin-top: 0; color: #155724;">🎉 Welcome to GoMailer TLS</h2>
        <p style="margin: 0; color: #155724;">
            Your TLS certificate has been generated successfully! You can now establish secure connections 
            to your GoMailer server using this CA certificate.
        </p>
    </div>`
	}
} 