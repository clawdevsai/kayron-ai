# Start MT5 MCP Server
# Usage: .\start-mcp.ps1 -Config config-ftmo-demo.yaml -Build

param(
    [string]$Config = "config-ftmo-demo.yaml",
    [switch]$Build,
    [switch]$Verbose,
    [switch]$Help
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Binary = "$ScriptDir\mcp-mt5-server.exe"
$LogFile = "$ScriptDir\mcp-server.log"
$ConfigPath = "$ScriptDir\$Config"

function Show-Help {
    @"
Start MT5 MCP Server

Usage: .\start-mcp.ps1 [options]

Options:
  -Config <string>    Config file path (default: config-ftmo-demo.yaml)
  -Build              Build binary before starting
  -Verbose            Enable verbose logging
  -Help               Show this help message

Environment Variables:
  MT5_PASSWORD        MetaTrader 5 password (required if not in config)
  MT5_LOGIN           MetaTrader 5 account login (required if not in config)
  MCP_PORT            Server port (default: 8080)
  MCP_LOG_LEVEL       Log level: debug, info, warn, error (default: info)

Examples:
  .\start-mcp.ps1
  .\start-mcp.ps1 -Config config-ftmo-demo.yaml
  .\start-mcp.ps1 -Config my-config.yaml -Verbose
  .\start-mcp.ps1 -Build
"@
}

function Write-Info {
    param([string]$Message)
    Write-Host "[MCP] $Message" -ForegroundColor Green
    Add-Content -Path $LogFile -Value "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] [MCP] $Message"
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
    Add-Content -Path $LogFile -Value "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] [WARN] $Message"
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    Add-Content -Path $LogFile -Value "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] [ERROR] $Message"
}

if ($Help) {
    Show-Help
    exit 0
}

Write-Info "MT5 MCP Server Startup"
Write-Info "Config: $Config"

# Check config exists
if (-not (Test-Path $ConfigPath)) {
    Write-Error-Custom "Config file not found: $ConfigPath"
    exit 1
}

# Build if requested
if ($Build) {
    Write-Info "Building binary..."
    Push-Location $ScriptDir
    try {
        go build -o .\mcp-mt5-server.exe .\cmd\mcp-mt5-server
        Write-Info "Build successful"
    }
    catch {
        Write-Error-Custom "Build failed: $_"
        exit 1
    }
    finally {
        Pop-Location
    }
}

# Check binary exists
if (-not (Test-Path $Binary)) {
    Write-Error-Custom "Binary not found: $Binary"
    Write-Error-Custom "Run '.\start-mcp.ps1 -Build' to compile"
    exit 1
}

Write-Info "Binary: $Binary"

# Set environment
$env:MCP_CONFIG = $Config
$env:MCP_LOG_LEVEL = if ($Verbose) { "debug" } else { "info" }

Write-Info "Starting server..."
Write-Info "Logs: $LogFile"

# Ensure log file exists
if (-not (Test-Path $LogFile)) {
    New-Item -Path $LogFile -ItemType File -Force | Out-Null
}

# Start server with logging
& $Binary 2>&1 | Tee-Object -FilePath $LogFile -Append
