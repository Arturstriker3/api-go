# GoMailer TLS Certificate Generation Script for Windows
# This creates self-signed certificates for development/testing

Write-Host "Generating TLS certificates for GoMailer..." -ForegroundColor Green

# Create certs directory if it doesn't exist
if (!(Test-Path "certs")) {
    New-Item -ItemType Directory -Path "certs"
    Write-Host "Created certs directory" -ForegroundColor Yellow
}

# Check if OpenSSL is available
try {
    $null = Get-Command openssl -ErrorAction Stop
    Write-Host "OpenSSL found" -ForegroundColor Green
} 
catch {
    Write-Host "OpenSSL not found. Please install OpenSSL first:" -ForegroundColor Red
    Write-Host "   Option 1: Install Git (includes OpenSSL)" -ForegroundColor Yellow
    Write-Host "   Option 2: Download from https://slproweb.com/products/Win32OpenSSL.html" -ForegroundColor Yellow
    Write-Host "   Option 3: Use Chocolatey: choco install openssl" -ForegroundColor Yellow
    exit 1
}

try {
    # Generate private key
    Write-Host "Generating private key..." -ForegroundColor Cyan
    & openssl genrsa -out certs/server.key 4096

    # Generate certificate signing request
    Write-Host "Generating certificate signing request..." -ForegroundColor Cyan
    & openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=US/ST=Dev/L=Local/O=GoMailer/OU=Dev/CN=localhost"

    # Generate self-signed certificate
    Write-Host "Generating self-signed certificate..." -ForegroundColor Cyan
    & openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt

    # Generate CA certificate (for client validation)
    Write-Host "Generating CA certificate..." -ForegroundColor Cyan
    Copy-Item certs/server.crt certs/ca-cert.pem

    # Clean up CSR file
    Remove-Item certs/server.csr -ErrorAction SilentlyContinue

    Write-Host "Certificates generated successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Generated files:" -ForegroundColor Yellow
    Write-Host "  - certs/server.key  (Private key - keep secure!)" -ForegroundColor White
    Write-Host "  - certs/server.crt  (Server certificate)" -ForegroundColor White
    Write-Host "  - certs/ca-cert.pem (CA certificate for client validation)" -ForegroundColor White
    Write-Host ""
    Write-Host "You can now start the GoMailer TLS server!" -ForegroundColor Green
} 
catch {
    Write-Host "Error generating certificates: " -ForegroundColor Red -NoNewline
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
} 