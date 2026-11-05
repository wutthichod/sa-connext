# Script to open Windows Firewall ports for WiFi network access
# Run this script as Administrator

Write-Host "Opening Windows Firewall ports for network access..." -ForegroundColor Green

# Open NodePort 30080 for Kubernetes API Gateway
Write-Host "Opening port 30080 (Kubernetes API Gateway)..." -ForegroundColor Yellow
try {
    New-NetFirewallRule -DisplayName "Kubernetes API Gateway (NodePort 30080)" `
        -Direction Inbound `
        -Protocol TCP `
        -LocalPort 30080 `
        -Action Allow `
        -ErrorAction Stop
    Write-Host "✓ Port 30080 opened successfully" -ForegroundColor Green
} catch {
    if ($_.Exception.Message -like "*already exists*") {
        Write-Host "✓ Port 30080 rule already exists" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to open port 30080: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Open port 3000 for Next.js Frontend
Write-Host "`nOpening port 3000 (Next.js Frontend)..." -ForegroundColor Yellow
try {
    New-NetFirewallRule -DisplayName "Next.js Dev Server (Port 3000)" `
        -Direction Inbound `
        -Protocol TCP `
        -LocalPort 3000 `
        -Action Allow `
        -ErrorAction Stop
    Write-Host "✓ Port 3000 opened successfully" -ForegroundColor Green
} catch {
    if ($_.Exception.Message -like "*already exists*") {
        Write-Host "✓ Port 3000 rule already exists" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to open port 3000: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Firewall Configuration Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "`nYour services are now accessible from other devices on the same WiFi network:" -ForegroundColor White
Write-Host "  Frontend: http://172.20.10.14:3000" -ForegroundColor Yellow
Write-Host "  Backend:  http://172.20.10.14:30080" -ForegroundColor Yellow
Write-Host "`nPress any key to continue..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
