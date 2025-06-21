#!/bin/bash

# GoMailer Certificate Auto-Renewal Cron Script
# This script should be run via cron to check and renew certificates

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/var/log/gomailer-cert-renewal.log"
CERT_RENEW_SCRIPT="$SCRIPT_DIR/auto-renew-certs.go"

# Function to log with timestamp
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Check if auto-renew script exists
if [ ! -f "$CERT_RENEW_SCRIPT" ]; then
    log "🔴 Certificate renewal script not found: $CERT_RENEW_SCRIPT"
    exit 1
fi

log "🟡 Starting certificate renewal check..."

# Change to the GoMailer directory
cd "$SCRIPT_DIR/.." || exit 1

# Run the certificate renewal check
if go run -tags renew_certs scripts/auto-renew-certs.go; then
    # Check if certificates were actually renewed by comparing timestamps
    if [ -f "certs/server.crt" ] && [ "$(find certs/server.crt -mmin -5)" ]; then
        log "🟢 Certificates were renewed! Sending reload signal to GoMailer..."
        
        # Try to reload certificates in running GoMailer instance
        # This assumes GoMailer is running as a systemd service
        if systemctl is-active --quiet gomailer; then
            systemctl reload gomailer
            log "🟡 GoMailer service reloaded"
        else
            log "🟡 GoMailer service not running or not managed by systemd"
        fi
    else
        log "🟢 Certificate check completed - no renewal needed"
    fi
else
    log "🔴 Certificate renewal check failed"
    exit 1
fi

log "🟡 Certificate renewal check completed" 