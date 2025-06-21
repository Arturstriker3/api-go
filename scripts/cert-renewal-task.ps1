# GoMailer Certificate Auto-Renewal PowerShell Script
# This script should be run via Windows Task Scheduler

param(
    [string]$GoMailerPath = "C:\gomailer"
)

$LogFile = "$GoMailerPath\logs\cert-renewal.log"
$CertRenewScript = "$GoMailerPath\scripts\auto-renew-certs.go"

# Create logs directory if it doesn't exist
New-Item -ItemType Directory -Force -Path "$GoMailerPath\logs" | Out-Null

# Function to log with timestamp
function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "$timestamp - $Message"
    Write-Host $logEntry
    Add-Content -Path $LogFile -Value $logEntry
}

Write-Log "🟡 Starting certificate renewal check..."

# Check if auto-renew script exists
if (-not (Test-Path $CertRenewScript)) {
    Write-Log "🔴 Certificate renewal script not found: $CertRenewScript"
    exit 1
}

# Change to GoMailer directory
Set-Location $GoMailerPath

try {
    # Run the certificate renewal check
    $process = Start-Process -FilePath "go" -ArgumentList "run", "-tags", "renew_certs", "scripts/auto-renew-certs.go" -Wait -PassThru -NoNewWindow
    
    if ($process.ExitCode -eq 0) {
        # Check if certificates were actually renewed by checking file modification time
        $certFile = "$GoMailerPath\certs\server.crt"
        if (Test-Path $certFile) {
            $certModified = (Get-Item $certFile).LastWriteTime
            $fiveMinutesAgo = (Get-Date).AddMinutes(-5)
            
            if ($certModified -gt $fiveMinutesAgo) {
                Write-Log "🟢 Certificates were renewed! Sending reload signal to GoMailer..."
                
                # Try to restart GoMailer Windows service
                try {
                    if (Get-Service -Name "GoMailer" -ErrorAction SilentlyContinue) {
                        Restart-Service -Name "GoMailer" -Force
                        Write-Log "🟡 GoMailer service restarted"
                    } else {
                        Write-Log "🟡 GoMailer service not found or not running as Windows service"
                    }
                } catch {
                    Write-Log "🟡 Could not restart GoMailer service: $($_.Exception.Message)"
                }
            } else {
                Write-Log "🟢 Certificate check completed - no renewal needed"
            }
        }
    } else {
        Write-Log "🔴 Certificate renewal check failed with exit code: $($process.ExitCode)"
        exit 1
    }
} catch {
    Write-Log "🔴 Error during certificate renewal: $($_.Exception.Message)"
    exit 1
}

Write-Log "🟡 Certificate renewal check completed" 