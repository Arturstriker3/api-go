//go:build renew_certs

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
	"path/filepath"
	"time"
)

const (
	renewalThreshold = 30 * 24 * time.Hour // 30 days
	certPath         = "certs/server.crt"
	keyPath          = "certs/server.key"
	caPath           = "certs/ca-cert.pem"
)

func main() {
	fmt.Println("🟡 Checking certificate expiration...")

	// Check if certificates exist
	if !fileExists(certPath) || !fileExists(keyPath) || !fileExists(caPath) {
		fmt.Println("🔴 Certificates not found. Generating initial certificates...")
		if err := generateCertificates(); err != nil {
			fmt.Printf("🔴 Failed to generate certificates: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("🟢 Initial certificates generated successfully!")
		return
	}

	// Check certificate expiration
	if shouldRenew, daysLeft := shouldRenewCertificate(); shouldRenew {
		fmt.Printf("🟡 Certificate expires in %d days. Generating new certificates...\n", int(daysLeft.Hours()/24))
		
		// Create backup
		if err := backupCertificates(); err != nil {
			fmt.Printf("🔴 Failed to backup certificates: %v\n", err)
			os.Exit(1)
		}

		// Generate new certificates
		if err := generateCertificates(); err != nil {
			fmt.Printf("🔴 Failed to renew certificates: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🟢 Certificates renewed successfully!")
		fmt.Println("🟡 Please restart the GoMailer service to load new certificates.")
	} else {
		daysLeft := int(time.Until(getCertificateExpiry()).Hours() / 24)
		fmt.Printf("🟢 Certificate is valid for %d more days. No renewal needed.\n", daysLeft)
	}
}

func generateCertificates() error {
	fmt.Println("🟡 Generating new TLS certificates...")

	// Create certs directory if it doesn't exist
	if err := os.MkdirAll("certs", 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %v", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Save private key
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %v", err)
	}
	defer keyFile.Close()

	keyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(keyFile, keyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization:  []string{"GoMailer"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", "gomailer", "*.gomailer.local"},
	}

	fmt.Println("🟡 Generating self-signed certificate...")

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Save certificate
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %v", err)
	}
	defer certFile.Close()

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	if err := pem.Encode(certFile, certPEM); err != nil {
		return fmt.Errorf("failed to encode certificate: %v", err)
	}

	// Copy server certificate as CA certificate for client validation
	if err := copyFile(certPath, caPath); err != nil {
		return fmt.Errorf("failed to create CA certificate: %v", err)
	}

	return nil
}

func backupCertificates() error {
	// Backup existing certificates
	files := []string{"server.crt", "server.key", "ca-cert.pem"}
	for _, file := range files {
		src := filepath.Join("certs", file)
		dst := filepath.Join("certs", fmt.Sprintf("backup_%d_%s", time.Now().Unix(), file))
		if fileExists(src) {
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to backup %s: %w", file, err)
			}
		}
	}
	return nil
}

func shouldRenewCertificate() (bool, time.Duration) {
	cert, err := loadCertificate(certPath)
	if err != nil {
		fmt.Printf("🔴 Error loading certificate: %v\n", err)
		return false, 0
	}
	daysLeft := time.Until(cert.NotAfter)
	return daysLeft <= renewalThreshold, daysLeft
}

func getCertificateExpiry() time.Time {
	cert, err := loadCertificate(certPath)
	if err != nil {
		fmt.Printf("🔴 Error loading certificate: %v\n", err)
		return time.Time{}
	}
	return cert.NotAfter
}

func loadCertificate(certFile string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	return x509.ParseCertificate(block.Bytes)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
} 