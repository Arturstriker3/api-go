package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	fmt.Println("🔐 Generating TLS certificates for GoMailer...")

	// Create certs directory
	if err := os.MkdirAll("certs", 0755); err != nil {
		log.Fatal("Failed to create certs directory:", err)
	}

	// Generate private key
	fmt.Println("Generating private key...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal("Failed to generate private key:", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Province:           []string{"Dev"},
			Locality:           []string{"Local"},
			Organization:       []string{"GoMailer"},
			OrganizationalUnit: []string{"Dev"},
			CommonName:         "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:              []string{"localhost"},
	}

	// Generate certificate
	fmt.Println("Generating self-signed certificate...")
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal("Failed to create certificate:", err)
	}

	// Save certificate
	certOut, err := os.Create("certs/server.crt")
	if err != nil {
		log.Fatal("Failed to create cert file:", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		log.Fatal("Failed to write certificate:", err)
	}

	// Save private key
	keyOut, err := os.Create("certs/server.key")
	if err != nil {
		log.Fatal("Failed to create key file:", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatal("Failed to marshal private key:", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		log.Fatal("Failed to write private key:", err)
	}

	// Copy certificate as CA cert
	fmt.Println("Generating CA certificate...")
	if err := copyFile("certs/server.crt", "certs/ca-cert.pem"); err != nil {
		log.Fatal("Failed to create CA cert:", err)
	}

	fmt.Println("✅ Certificates generated successfully!")
	fmt.Println("")
	fmt.Println("📁 Generated files:")
	fmt.Println("  - certs/server.key  (Private key - keep secure!)")
	fmt.Println("  - certs/server.crt  (Server certificate)")
	fmt.Println("  - certs/ca-cert.pem (CA certificate for client validation)")
	fmt.Println("")
	fmt.Println("🚀 You can now start the GoMailer TLS server!")
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, input, 0644)
} 