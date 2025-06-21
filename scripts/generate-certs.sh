#!/bin/bash

# GoMailer TLS Certificate Generation Script
# This creates self-signed certificates for development/testing

echo "🔐 Generating TLS certificates for GoMailer..."

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate private key
echo "Generating private key..."
openssl genrsa -out certs/server.key 4096

# Generate certificate signing request
echo "Generating certificate signing request..."
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=US/ST=Dev/L=Local/O=GoMailer/OU=Dev/CN=localhost"

# Generate self-signed certificate
echo "Generating self-signed certificate..."
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt

# Generate CA certificate (for client validation)
echo "Generating CA certificate..."
cp certs/server.crt certs/ca-cert.pem

# Set proper permissions
chmod 600 certs/server.key
chmod 644 certs/server.crt certs/ca-cert.pem

echo "✅ Certificates generated successfully!"
echo ""
echo "📁 Generated files:"
echo "  - certs/server.key  (Private key - keep secure!)"
echo "  - certs/server.crt  (Server certificate)"
echo "  - certs/ca-cert.pem (CA certificate for client validation)"
echo ""
echo "🚀 You can now start the GoMailer TLS server!"

# Clean up CSR file
rm certs/server.csr 