package errors

// Portuguese (pt-BR) error messages for all MT5 failure modes
const (
	// Connection & Disconnection
	PTBRTerminalDisconnected   = "Terminal desconectado"
	PTBRReconnecting           = "Reconectando ao terminal"
	PTBRConnectionRefused      = "Conexão recusada pelo terminal"
	PTBRConnectionTimeout      = "Tempo limite de conexão excedido"

	// Authentication & Credentials
	PTBRInvalidCredentials     = "Credenciais inválidas"
	PTBRAuthenticationFailed   = "Falha na autenticação"
	PTBRUnauthorized           = "Não autorizado"
	PTBRCredentialsExpired     = "Credenciais expiradas"

	// Trading Operations
	PTBRInsufficientMargin     = "Saldo insuficiente para a margem"
	PTBRInsufficientBalance    = "Saldo insuficiente"
	PTBRMarginCall             = "Chamada de margem acionada"
	PTBRStopOutTriggered       = "Stop out acionado"

	// Symbols & Quotes
	PTBRSymbolNotFound         = "Símbolo não encontrado"
	PTBRUnknownSymbol          = "Símbolo desconhecido"
	PTBRQuoteUnavailable       = "Cotação indisponível"
	PTBRNoQuote                = "Sem cotação disponível"

	// Price & Timing
	PTBRPriceGapping           = "Abertura de preço detectada"
	PTBRSlippageExceeded       = "Deslizamento de preço excedido"
	PTBRDeadlineExceeded       = "Tempo limite excedido"
	PTBROrderTimeout           = "Tempo limite do pedido excedido"

	// Order Operations
	PTBROrderRejected          = "Pedido rejeitado"
	PTBROrderCancelled         = "Pedido cancelado"
	PTBROrderClosed            = "Pedido fechado"
	PTBRDuplicateOrder         = "Pedido duplicado"
	PTBRInvalidOrderType       = "Tipo de pedido inválido"

	// Position Operations
	PTBRPositionNotFound       = "Posição não encontrada"
	PTBRPositionClosed         = "Posição já foi fechada"
	PTBRCannotClosePosition    = "Não é possível fechar a posição"
	PTBRPartialCloseUnsupported = "Fechamento parcial não suportado"

	// System & Internal
	PTBRInternalError          = "Erro interno do terminal"
	PTBRDataAccessError        = "Erro ao acessar dados"
	PTBRQueueProcessingError   = "Erro ao processar fila"
	PTBRIdempotencyError       = "Erro na verificação de idempotência"

	// Generic
	PTBRErrorUnknown           = "Erro desconhecido"
	PTBROperationFailed        = "Operação falhou"
	PTBRRetryAvailable         = "Tente novamente"
)

// ErrorMessageMap maps error patterns to pt-BR messages
var ErrorMessageMap = map[string]string{
	"disconnected":           PTBRTerminalDisconnected,
	"connection refused":     PTBRConnectionRefused,
	"connection timeout":     PTBRConnectionTimeout,
	"invalid credentials":    PTBRInvalidCredentials,
	"authentication failed":  PTBRAuthenticationFailed,
	"unauthorized":           PTBRUnauthorized,
	"insufficient margin":    PTBRInsufficientMargin,
	"insufficient balance":   PTBRInsufficientBalance,
	"symbol not found":       PTBRSymbolNotFound,
	"unknown symbol":         PTBRUnknownSymbol,
	"quote unavailable":      PTBRQuoteUnavailable,
	"no quote":               PTBRNoQuote,
	"price gap":              PTBRPriceGapping,
	"slippage exceeded":      PTBRSlippageExceeded,
	"deadline exceeded":      PTBRDeadlineExceeded,
	"order timeout":          PTBROrderTimeout,
	"order rejected":         PTBROrderRejected,
	"order cancelled":        PTBROrderCancelled,
	"order closed":           PTBROrderClosed,
	"duplicate order":        PTBRDuplicateOrder,
	"invalid order type":     PTBRInvalidOrderType,
	"position not found":     PTBRPositionNotFound,
	"position closed":        PTBRPositionClosed,
	"cannot close position":  PTBRCannotClosePosition,
	"partial close unsupported": PTBRPartialCloseUnsupported,
	"internal error":         PTBRInternalError,
	"data access error":      PTBRDataAccessError,
	"queue processing error": PTBRQueueProcessingError,
	"idempotency error":      PTBRIdempotencyError,
}

// GetPTBRMessage returns Portuguese error message for English error string
func GetPTBRMessage(englishError string) string {
	if ptbrMsg, ok := ErrorMessageMap[englishError]; ok {
		return ptbrMsg
	}
	return PTBRErrorUnknown
}
