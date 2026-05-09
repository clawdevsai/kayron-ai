#!/bin/bash
# MT5 MCP Server - Start Script
# Opens MT5 Terminal (if needed) + Starts MCP Server

CONFIG="${1:-config-real-ftmo.yaml}"
MT5_PATH="C:/Program Files/FTMO Global Markets MT5 Terminal/terminal64.exe"
WAIT_SECONDS=5

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/bin/mcp-mt5-server"
LOG_FILE="$SCRIPT_DIR/mcp-server.log"

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ERROR:${NC} $1"
}

show_help() {
    cat << EOF
MT5 MCP Server - Start Script

Usage: $0 [config-file]

Examples:
  $0                          # Start with default config-real-ftmo.yaml
  $0 config-ftmo-demo.yaml    # Start with specific config
  $0 --help                   # Show this help

Config Files:
  - config-real-ftmo.yaml     (recommended) Real MT5 WebAPI connection
  - config-ftmo-demo.yaml     (mock data)

EOF
}

# Parse arguments
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    show_help
    exit 0
fi

# Check MT5 path exists
if [[ ! -f "$MT5_PATH" ]]; then
    error "MT5 não encontrado em: $MT5_PATH"
    exit 1
fi

# Check MCP binary exists
if [[ ! -f "$BINARY" ]]; then
    error "MCP binary não encontrado: $BINARY"
    exit 1
fi

log "=========================================="
log "MT5 MCP Server"
log "=========================================="
log ""

log "Abrindo MT5 Terminal..."
log "Caminho: $MT5_PATH"

# Launch MT5 in background (detached)
"$MT5_PATH" > /dev/null 2>&1 &
MT5_PID=$!

log "MT5 iniciado (PID: $MT5_PID)"
log "Aguardando ${WAIT_SECONDS}s para inicializar..."
sleep $WAIT_SECONDS

log ""
log "=========================================="
log "Iniciando MCP Server"
log "=========================================="
log "Config: $CONFIG"
log "HTTP Server: http://localhost:8080/rpc"
log "gRPC Daemon: localhost:50051"
log ""

log "⚠️  CONFIGURAR NO MT5:"
log "  1. Ferramentas → Opções"
log "  2. Aba: Consultores Especialistas"
log "  3. ✅ Marque 'Permitir negociações automatizadas'"
log "  4. Procure 'WebAPI' ou 'Servidor'"
log "  5. ✅ Ative WebAPI"
log "  6. Anote a porta (padrão: 8228)"
log ""

log "Logs: $LOG_FILE"
log ""

export MCP_CONFIG="$CONFIG"
export MCP_LOG_LEVEL="info"

"$BINARY" 2>&1 | tee -a "$LOG_FILE"
