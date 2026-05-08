# MT5 MCP Server - 16 Ferramentas Completas

Servidor gRPC + JSON-RPC que integra MetaTrader 5 com Claude via Model Context Protocol. Todas 16 ferramentas implementadas e testadas.

## 📊 Ferramentas Disponíveis

### Fase 1: Core (6 ferramentas)

| Tool | Descrição | Entrada | Saída |
|------|-----------|---------|-------|
| `account-info` | Informações da conta MT5 | - | `{balance, equity, free_margin, margin_level, currency}` |
| `quote` | Cotação atual de um símbolo | `symbol` | `{symbol, bid, ask, spread}` |
| `place-order` | Abrir posição | `symbol, volume, type, price` | `{ticket, symbol, volume, open_price}` |
| `close-position` | Fechar posição | `ticket` | `{ticket, closed_price, profit_loss}` |
| `orders-list` | Listar todas posições | - | `[{ticket, symbol, volume, open_price, ...}]` |
| `get-candles` | Histórico OHLC | `symbol, timeframe, count` | `[{time, open, high, low, close, volume}]` |

### Fase 2: Evolução (4 ferramentas)

| Tool | Descrição | Entrada | Saída |
|------|-----------|---------|-------|
| `modify-order` | Modificar SL/TP de ordem | `ticket, stop_loss, take_profit` | `{ticket, stop_loss, take_profit}` |
| `pending-order-details` | Detalhe de ordem pendente | `symbol` | `[{ticket, type, price, volume, ...}]` |
| `symbol-properties` | Propriedades do símbolo | `symbol` | `{symbol, digits, tick_size, contract_size, ...}` |
| `margin-calculator` | Calcular margem necessária | `symbol, volume` | `{symbol, volume, margin_required, percentage}` |

### Fase 3: Análise (3 ferramentas)

| Tool | Descrição | Entrada | Saída |
|------|-----------|---------|-------|
| `position-details` | Detalhe completo de posição | `symbol` | `[{ticket, volume, current_price, profit, swap, ...}]` |
| `account-equity-history` | Histórico de equity | `from_timestamp, to_timestamp, granularity` | `[{timestamp, equity, balance}]` |
| `balance-drawdown` | Redução máxima de saldo | `since_timestamp` | `{max_equity, current_equity, drawdown_percent}` |

### Fase 4: Avançado (3 ferramentas)

| Tool | Descrição | Entrada | Saída |
|------|-----------|---------|-------|
| `order-fill-analysis` | Análise de execução | `ticket` | `{ticket, fill_price, slippage, execution_latency_ms}` |
| `market-hours` | Horários de mercado | `symbol` | `{open_time, close_time, timezone, is_closed}` |
| `tick-data` | Dados de tick (bid/ask) | `symbol, duration_seconds` | `[{timestamp, bid, ask}]` |

## 🏗️ Arquitetura

```
JSON-RPC Request (HTTP POST)
    ↓
MCP Server (cmd/mcp-mt5-server/main.go)
    ├── MCP Tool Layer (internal/services/mcp/*.go)
    ├── Daemon Handler Layer (internal/services/daemon/*.go)
    └── MT5 Service Layer (internal/services/mt5/*.go)
            ↓
       gRPC Daemon (daemon-service)
            ↓
       MT5 WebAPI HTTP Client
```

### Camadas

1. **MCP Tool** (`mcp/*.go`): Interface JSON-RPC, parâmetros, formatação
2. **Daemon Handler** (`daemon/*.go`): Conversão request/response, gRPC calls
3. **MT5 Service** (`mt5/*.go`): Lógica de negócio, integrações, mock data

## 🚀 Build e Deploy

```bash
# Build servidor
make build

# Rodar servidor
./bin/mcp-mt5-server

# Rodar testes integração
go test ./cmd/mcp-mt5-server -v

# Verificar todas 16 tools
curl -X POST http://localhost:9090/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"account-info","params":{},"id":1}'
```

## ✅ Testes

- **Integração**: 16 testes validam todas as ferramentas
- **Unit**: Testes por serviço (quote, order, position, etc)
- **Mock**: Todas ferramentas funcionam sem MT5 real

```bash
go test ./cmd/mcp-mt5-server -v -timeout 15s
```

Resultado esperado: **16/16 PASS** (account-info, quote, place-order, close-position, orders-list, get-candles, modify-order, pending-order-details, symbol-properties, margin-calculator, position-details, account-equity-history, balance-drawdown, order-fill-analysis, market-hours, tick-data)

## 📋 Implementação

- ✅ Todas 10 features de evolução (Phases 2-4)
- ✅ Todos 6 core tools preservados (Phase 1)
- ✅ Testes integração validam 16 tools
- ✅ Tratamento de erros multilíngue (pt-BR)
- ✅ Decimal precision com shopspring/decimal
- ✅ Mock data realista para desenvolvimento

## 🔧 Configuração

Via `config.yaml` ou variáveis env:

```yaml
mt5:
  server: "localhost:8228"
  login: "123456"
  password: "senha"
  timeout: 30

http:
  port: 9090

grpc:
  port: 50051
```

## 📝 Próximos Passos

1. Integração com servidor MT5 real
2. Persistência de fila (SQLite)
3. Autenticação e autorização
4. Monitoramento (Prometheus/Grafana)
5. Documentação por tool (schemas, exemplos)
