# setup.ps1 — Orbitron Bot Windows Developer Setup
# Equivalent of setup.sh for Windows.
# Usage: .\setup.ps1
#Requires -Version 5

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$GO_VERSION = "1.24.4"

function Write-Info  { param($m) Write-Host "[setup] $m" -ForegroundColor Green  }
function Write-Warn  { param($m) Write-Host "[setup] $m" -ForegroundColor Yellow }
function Write-Fatal { param($m) Write-Host "[setup] ERROR: $m" -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Orbitron Universal Setup (Windows)"    -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green

# ── 1. Check/Install Go ───────────────────────────────────────────────────────
$goOk = $false
try {
    $installedVer = (go version 2>$null) -replace "go version go(\S+).*", '$1'
    if ($installedVer -eq $GO_VERSION) {
        Write-Info "Go $GO_VERSION is already installed."
        $goOk = $true
    } else {
        Write-Warn "Found Go $installedVer, need $GO_VERSION. Installing locally..."
    }
} catch {}

if (-not $goOk) {
    $goZip      = "go$GO_VERSION.windows-amd64.zip"
    $goUrl      = "https://go.dev/dl/$goZip"
    $localGoDir = Join-Path $PSScriptRoot "local_go"

    Write-Info "Downloading $goZip ..."
    Invoke-WebRequest -Uri $goUrl -OutFile $goZip -UseBasicParsing

    Write-Info "Extracting to .\local_go ..."
    if (Test-Path $localGoDir) { Remove-Item $localGoDir -Recurse -Force }
    New-Item -ItemType Directory -Path $localGoDir | Out-Null
    Expand-Archive -Path $goZip -DestinationPath $localGoDir -Force
    Remove-Item $goZip

    $env:PATH = "$localGoDir\go\bin;$env:PATH"
    Write-Info "Go installed locally: $(go version)"
}

# ── 2. Check Node.js ──────────────────────────────────────────────────────────
$nodeOk = $false
try {
    $nodeVer = node --version 2>$null
    $npmVer  = npm --version 2>$null
    if ($nodeVer -and $npmVer) {
        Write-Info "Node.js $nodeVer and npm $npmVer already installed."
        $nodeOk = $true
    }
} catch {}

if (-not $nodeOk) {
    Write-Warn "Node.js/npm not found."
    Write-Warn "Please install Node.js 20+ from https://nodejs.org/ and re-run this script."
    Write-Fatal "Node.js is required to build the frontend."
}

# ── 3. Create config.toml if absent ──────────────────────────────────────────
$CONFIG_FILE = Join-Path $PSScriptRoot "config.toml"
if (Test-Path $CONFIG_FILE) {
    Write-Info "config.toml already exists. Skipping creation."
} else {
    Write-Warn "config.toml not found. Creating a minimal template..."
    $secureKey = Read-Host "  Enter your L1 wallet private key (hex, no 0x prefix)" -AsSecureString
    $PRIV_KEY  = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto(
        [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureKey)
    )

    $configContent = @"
[webui]
enabled = true
listen  = "127.0.0.1:8080"
# IMPORTANT: change this secret before running in production
jwt_secret = "changeme123"

[ui]
language = "en"

[auth]
private_key = "$PRIV_KEY"

[log]
level  = "info"
format = "pretty"
"@
    Set-Content -Path $CONFIG_FILE -Value $configContent -Encoding UTF8
    Write-Info "config.toml created."
}

# ── 4. Build Frontend ─────────────────────────────────────────────────────────
$webDir = Join-Path $PSScriptRoot "internal\webui\web"
if (Test-Path $webDir) {
    Write-Info "Building Vue 3 Frontend..."
    Push-Location $webDir
    try {
        npm install --silent
        npm run build
    } finally {
        Pop-Location
    }
    Write-Info "Frontend built."
} else {
    Write-Warn "Directory internal\webui\web not found — frontend build skipped."
}

# ── 5. Build Go Binary ────────────────────────────────────────────────────────
Write-Info "Running go mod tidy..."
Push-Location $PSScriptRoot
try {
    go mod tidy

    $BIN_NAME = "orbitron-polytrade-bot.exe"
    Write-Info "Compiling $BIN_NAME ..."

    go build `
        -ldflags "-s -w" `
        -o $BIN_NAME `
        .\cmd\bot\

    if (Test-Path $BIN_NAME) {
        Write-Info "Backend built: $BIN_NAME"
    } else {
        Write-Fatal "Build failed — $BIN_NAME not found."
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Run: " -NoNewline; Write-Host ".\$BIN_NAME --config config.toml" -ForegroundColor Yellow
Write-Host ""
