//go:build generate_certs

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	fmt.Println("🟡 Generating TLS certificates for GoMailer...")

	// Create certs directory if it doesn't exist
	if err := os.MkdirAll("certs", 0755); err != nil {
		fmt.Printf("🔴 Error creating certs directory: %v\n", err)
		os.Exit(1)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("🔴 Error generating private key: %v\n", err)
		os.Exit(1)
	}

	// Save private key
	keyFile, err := os.Create("certs/server.key")
	if err != nil {
		fmt.Printf("🔴 Error creating key file: %v\n", err)
		os.Exit(1)
	}
	defer keyFile.Close()

	keyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(keyFile, keyPEM); err != nil {
		fmt.Printf("🔴 Error encoding private key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🟡 Generating private key...")
	fmt.Println("🟡 Generating self-signed certificate...")

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"GoMailer"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost", "gomailer", "*.gomailer.local"},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		fmt.Printf("🔴 Error creating certificate: %v\n", err)
		os.Exit(1)
	}

	// Save certificate
	certFile, err := os.Create("certs/server.crt")
	if err != nil {
		fmt.Printf("🔴 Error creating certificate file: %v\n", err)
		os.Exit(1)
	}
	defer certFile.Close()

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	if err := pem.Encode(certFile, certPEM); err != nil {
		fmt.Printf("🔴 Error encoding certificate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🟡 Generating CA certificate...")

	// Copy server certificate as CA certificate for client validation
	if err := copyFile("certs/server.crt", "certs/ca-cert.pem"); err != nil {
		fmt.Printf("🔴 Error creating CA certificate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Certificates generated successfully!")
	fmt.Println("")
	fmt.Println("📁 Generated files:")
	fmt.Println("- certs/server.key (Private key - keep secure!)")
	fmt.Println("- certs/server.crt (Server certificate)")
	fmt.Println("- certs/ca-cert.pem (CA certificate for client validation)")
	fmt.Println("")

	// Create notification file for the API to detect new certificates
	if isDockerEnvironment() {
		if err := createCertificateNotification("NEW"); err != nil {
			fmt.Printf("🟡 Warning: Could not create certificate notification: %v\n", err)
		} else {
			fmt.Println("🟡 Certificate notification created for email service")
		}
	}

	fmt.Println("🟢 You can now start the GoMailer TLS server!")
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, input, 0644)
}

func isDockerEnvironment() bool {
	// Check if running inside Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func createCertificateNotification(action string) error {
	// Create a notification file that the API can detect
	notification := fmt.Sprintf(`{
	"action": "%s",
	"timestamp": "%s",
	"certificate_path": "certs/ca-cert.pem"
}`, action, time.Now().Format(time.RFC3339))

	return os.WriteFile("certs/certificate_notification.json", []byte(notification), 0644)
} 