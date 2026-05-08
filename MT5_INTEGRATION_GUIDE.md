# MT5 WebAPI Integration Guide

Guia para integrar servidor MT5 real com MCP server.

## 📋 Endpoints MT5 WebAPI Esperados

### Autenticação
Todos endpoints usam **HTTP Basic Auth** com `login` e `password` do MT5.

### Core Endpoints (Implementados)

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/api/account` | Informações da conta (balance, equity, margin) |
| GET | `/api/quote/{symbol}` | Cotação atual (bid/ask) |
| POST | `/api/order` | Abrir nova posição |
| POST | `/api/order/{ticket}/close` | Fechar posição |
| GET | `/api/orders?filter={filter}` | Listar posições (filter: open/closed/pending) |

### Evolução Endpoints (Novos)

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/symbols/{symbol}/candles?tf={timeframe}&count={count}` | Dados OHLC históricos |
| PUT | `/api/order/{ticket}` | Modificar SL/TP |
| GET | `/symbols/{symbol}/properties` | Propriedades símbolo (digits, tick_size, contract) |
| GET | `/symbols/{symbol}/hours` | Horários mercado |
| GET | `/symbols/{symbol}/ticks?duration={seconds}` | Dados tick (bid/ask histórico) |
| GET | `/api/equity/history?from={ts}&to={ts}` | Histórico equity |
| GET | `/symbols/{symbol}/positions` | Posições abertas símbolo |

## 🔌 Configuração

### Arquivo `config.yaml`

```yaml
mt5:
  server: "http://your-mt5-server:8228"  # URL WebAPI MT5
  login: "123456"                         # Número conta MT5
  password: "your_password"               # Senha API MT5
  timeout: 30                             # Timeout segundos

http:
  port: 9090

grpc:
  port: 50051
```

### Variáveis Ambiente

```bash
export MT5_SERVER="http://localhost:8228"
export MT5_LOGIN="123456"
export MT5_PASSWORD="senha"
export MT5_TIMEOUT="30"
export HTTP_PORT="9090"
export GRPC_PORT="50051"
```

## 🧪 Teste de Conexão

```bash
# Test account endpoint
curl -X GET http://your-mt5-server:8228/api/account \
  -H "Authorization: Basic $(echo -n 'login:password' | base64)" \
  -H "Content-Type: application/json"

# Test quote endpoint  
curl -X GET http://your-mt5-server:8228/api/quote/EURUSD \
  -H "Authorization: Basic $(echo -n 'login:password' | base64)"

# Test candles endpoint
curl -X GET "http://your-mt5-server:8228/symbols/EURUSD/candles?tf=H1&count=10" \
  -H "Authorization: Basic $(echo -n 'login:password' | base64)"
```

## 📊 Response Formats

### Account Info Response
```json
{
  "login": 123456,
  "balance": "10000.00",
  "equity": "10500.50",
  "margin": "2000.00",
  "free_margin": "8500.50",
  "margin_level": "525.25",
  "currency": "USD"
}
```

### Quote Response
```json
{
  "symbol": "EURUSD",
  "bid": "1.0950",
  "ask": "1.0952",
  "time": 1715000000
}
```

### Candles Response
```json
[
  {
    "time": 1715000000,
    "open": "1.0940",
    "high": "1.0960",
    "low": "1.0935",
    "close": "1.0950",
    "volume": 1000
  }
]
```

### Position Response
```json
{
  "ticket": 12345,
  "symbol": "EURUSD",
  "type": "BUY",
  "volume": "1.0",
  "open_price": "1.0900",
  "current_price": "1.0950",
  "profit_loss": "50.00",
  "swap": "0.25",
  "open_time": 1715000000
}
```

### Market Hours Response
```json
{
  "symbol": "EURUSD",
  "open_time": "0800",
  "close_time": "1700",
  "timezone": "GMT",
  "is_closed": false
}
```

### Tick Data Response
```json
[
  {
    "timestamp": 1715000000,
    "bid": "1.0950",
    "ask": "1.0952"
  },
  {
    "timestamp": 1715000100,
    "bid": "1.0951",
    "ask": "1.0953"
  }
]
```

## ⚠️ Tratamento de Erros

### HTTP Status Codes
- **200 OK** - Requisição bem-sucedida
- **201 Created** - Ordem criada
- **400 Bad Request** - Parâmetros inválidos
- **401 Unauthorized** - Credenciais inválidas
- **404 Not Found** - Símbolo/posição não existe
- **500 Server Error** - Erro servidor MT5

### Erro Response
```json
{
  "error": {
    "code": "INVALID_VOLUME",
    "message": "Volume must be positive",
    "details": "Provided: -1.0"
  }
}
```

## 🚀 Migração Mock → Real

### Step 1: Configurar servidor MT5
1. Instalar MT5 com API WebAPI habilitada
2. Gerar token API em MT5 settings
3. Testar endpoints com curl

### Step 2: Atualizar config
```bash
# config.yaml
mt5:
  server: "http://actual-mt5-server:8228"
  login: "your_login"
  password: "your_password"
```

### Step 3: Remover mocks
Services que ainda usam mock:
- `account_equity_history_service.go` - substituir geração mock por API
- `market_hours_service.go` - substituir por `client.GetMarketHours()`
- `order_fill_analysis_service.go` - substituir por API histórico
- `position_details_service.go` - substituir por `client.GetPositions()`
- `tick_data_service.go` - substituir por `client.GetTickData()`

### Step 4: Test
```bash
# Build
go build ./cmd/mcp-mt5-server/main.go

# Run
./mcp-mt5-server

# Test tools
go test ./cmd/mcp-mt5-server -v
```

## 🔐 Segurança

### Autenticação
- ✅ HTTP Basic Auth incluído em todos requests
- ⚠️ Use HTTPS em produção (não HTTP)
- 🔐 Armazene credenciais em secrets manager (não git)

### Exemplo Deployment Seguro
```bash
# Use environment variables + secrets manager
export MT5_PASSWORD=$(aws secretsmanager get-secret-value --secret-id mt5-password | jq -r .SecretString)

./mcp-mt5-server
```

## 📈 Monitoramento

### Métricas Importantes
- Latência endpoint MT5 (ms)
- Taxa erro conexão
- Número requisições/min
- Posições abertas / ordem count

### Logs
Todos requests/responses logados com:
- Timestamp
- Endpoint
- Status code
- Latência (ms)
- Erros

## 🐛 Troubleshooting

### Conexão Recusada
```
[CONNECTION_FAILED] Falha na conexão com o servidor MT5
```
**Solução**: Verificar URL MT5, porta, firewall

### Autenticação Falha
```
[AUTHENTICATION_FAILED] Invalid credentials
```
**Solução**: Verificar login/password MT5

### Símbolo Inválido
```
[INVALID_SYMBOL] Symbol not found
```
**Solução**: Verificar símbolo (ex: EURUSD vs eu)

### Timeout
```
context deadline exceeded
```
**Solução**: Aumentar timeout em config, ou verificar MT5 performance
