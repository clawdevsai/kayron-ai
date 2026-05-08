package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MT5Error represents an MT5-specific error
type MT5Error struct {
	Code    string
	Message string
	Details string
}

// Error implements the error interface
func (e *MT5Error) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// ToGRPCStatus converts MT5Error to gRPC status
func (e *MT5Error) ToGRPCStatus() *status.Status {
	grpcCode := CodeToGRPC(e.Code)
	return status.New(grpcCode, e.Message)
}

// Error codes
const (
	ErrAuthenticationFailed = "AUTH_FAILED"
	ErrConnectionFailed     = "CONNECTION_FAILED"
	ErrAccountNotFound      = "ACCOUNT_NOT_FOUND"
	ErrInvalidSymbol        = "INVALID_SYMBOL"
	ErrInvalidVolume        = "INVALID_VOLUME"
	ErrInsufficientMargin   = "INSUFFICIENT_MARGIN"
	ErrPositionNotFound     = "POSITION_NOT_FOUND"
	ErrInvalidPrice         = "INVALID_PRICE"
	ErrOrderRejected        = "ORDER_REJECTED"
	ErrNetworkError         = "NETWORK_ERROR"
	ErrTimeout              = "TIMEOUT"
	ErrInternal             = "INTERNAL_ERROR"
)

// CodeToGRPC maps error codes to gRPC status codes
func CodeToGRPC(code string) codes.Code {
	switch code {
	case ErrAuthenticationFailed:
		return codes.Unauthenticated
	case ErrConnectionFailed, ErrNetworkError, ErrTimeout:
		return codes.Unavailable
	case ErrAccountNotFound, ErrPositionNotFound, ErrInvalidSymbol:
		return codes.NotFound
	case ErrInvalidVolume, ErrInvalidPrice:
		return codes.InvalidArgument
	case ErrInsufficientMargin, ErrOrderRejected:
		return codes.PermissionDenied
	case ErrInternal:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// NewMT5Error creates a new MT5Error
func NewMT5Error(code, message, details string) *MT5Error {
	return &MT5Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common error constructors
func AuthenticationFailed(details string) *MT5Error {
	return NewMT5Error(ErrAuthenticationFailed, GetMessage(ErrAuthenticationFailed), details)
}

func ConnectionFailed(details string) *MT5Error {
	return NewMT5Error(ErrConnectionFailed, GetMessage(ErrConnectionFailed), details)
}

func AccountNotFound(accountID string) *MT5Error {
	return NewMT5Error(ErrAccountNotFound, GetMessage(ErrAccountNotFound), accountID)
}

func InvalidSymbol(symbol string) *MT5Error {
	return NewMT5Error(ErrInvalidSymbol, GetMessage(ErrInvalidSymbol), symbol)
}

func InsufficientMargin(required, available float64) *MT5Error {
	return NewMT5Error(ErrInsufficientMargin, GetMessage(ErrInsufficientMargin),
		fmt.Sprintf("Required: %.2f, Available: %.2f", required, available))
}

// GetMessage returns Portuguese-BR error message
func GetMessage(code string) string {
	messages := map[string]string{
		ErrAuthenticationFailed: "Falha na autenticação com o servidor MT5",
		ErrConnectionFailed:     "Falha na conexão com o servidor MT5",
		ErrAccountNotFound:      "Conta não encontrada",
		ErrInvalidSymbol:        "Símbolo inválido ou não suportado",
		ErrInvalidVolume:        "Volume inválido",
		ErrInsufficientMargin:   "Margem insuficiente",
		ErrPositionNotFound:     "Posição não encontrada",
		ErrInvalidPrice:         "Preço inválido",
		ErrOrderRejected:        "Ordem rejeitada pelo servidor",
		ErrNetworkError:         "Erro de rede",
		ErrTimeout:              "Tempo limite excedido",
		ErrInternal:             "Erro interno do servidor",
	}
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "Erro desconhecido"
}
