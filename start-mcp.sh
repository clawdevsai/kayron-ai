#!/bin/bash
# Start MT5 MCP Server
# Usage: ./start-mcp.sh [config-file] [--help]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${1:-config-ftmo-demo.yaml}"
BINARY="$SCRIPT_DIR/bin/mcp-mt5-server"
LOG_FILE="$SCRIPT_DIR/mcp-server.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

show_help() {
    cat << EOF
Start MT5 MCP Server

Usage: ./start-mcp.sh [options]

Options:
  [config-file]    Config file path (default: config-ftmo-demo.yaml)
  --help           Show this help message
  --build          Build binary before starting
  --verbose        Enable verbose logging

Environment Variables:
  MT5_PASSWORD     MetaTrader 5 password (required if not in config)
  MT5_LOGIN        MetaTrader 5 account login (required if not in config)
  MCP_PORT         Server port (default: 8080)
  MCP_LOG_LEVEL    Log level: debug, info, warn, error (default: info)

Examples:
  ./start-mcp.sh
  ./start-mcp.sh config-ftmo-demo.yaml
  ./start-mcp.sh my-config.yaml --verbose
  ./start-mcp.sh --build
EOF
}

log_info() {
    echo -e "${GREEN}[MCP]${NC} $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Parse arguments
BUILD=false
VERBOSE=false

for arg in "$@"; do
    case "$arg" in
        --help) show_help; exit 0 ;;
        --build) BUILD=true ;;
        --verbose) VERBOSE=true ;;
    esac
done

log_info "MT5 MCP Server Startup"
log_info "Config: $CONFIG_FILE"

# Check config exists
if [[ ! -f "$SCRIPT_DIR/$CONFIG_FILE" ]]; then
    log_error "Config file not found: $CONFIG_FILE"
    exit 1
fi

# Build if requested
if [[ "$BUILD" == true ]]; then
    log_info "Building binary..."
    cd "$SCRIPT_DIR"
    go build -o ./bin/mcp-mt5-server ./cmd/mcp-mt5-server
    BINARY="$SCRIPT_DIR/bin/mcp-mt5-server"
fi

# Try .exe extension on Windows/Git Bash
if [[ ! -f "$BINARY" && -f "$BINARY.exe" ]]; then
    BINARY="$BINARY.exe"
fi

# Check binary exists
if [[ ! -f "$BINARY" ]]; then
    log_error "Binary not found: $BINARY. Run './start-mcp.sh --build' to compile."
    exit 1
fi

log_info "Binary: $BINARY"

# Set environment
export MCP_CONFIG="$CONFIG_FILE"
export MCP_LOG_LEVEL="${MCP_LOG_LEVEL:-info}"
if [[ "$VERBOSE" == true ]]; then
    export MCP_LOG_LEVEL="debug"
fi

# Start server
log_info "Starting server..."
log_info "Logs: $LOG_FILE"

"$BINARY" 2>&1 | tee -a "$LOG_FILE"
