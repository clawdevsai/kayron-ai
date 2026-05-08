package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MT5ErrorType represents types of MT5-specific errors
type MT5ErrorType int

const (
	ErrTypeUnknown MT5ErrorType = iota
	ErrTypeDisconnect
	ErrTypeTimeout
	ErrTypeInvalidCredentials
	ErrTypeMarginInsufficient
	ErrTypeSymbolNotFound
	ErrTypePriceGapping
	ErrTypeQuoteUnavailable
	ErrTypeOrderRejected
	ErrTypePositionClosed
)

// MT5Error represents a structured MT5 error
type MT5Error struct {
	Type      MT5ErrorType
	Message   string
	Original  error
	Code      codes.Code
	PTBRMsg   string
}

// Error implements error interface
func (e *MT5Error) Error() string {
	return e.Message
}

// ToGRPCStatus converts MT5Error to gRPC status
func (e *MT5Error) ToGRPCStatus() *status.Status {
	return status.New(e.Code, e.PTBRMsg)
}

// NewMT5Error creates a new MT5Error
func NewMT5Error(errType MT5ErrorType, msg, ptbrMsg string, original error) *MT5Error {
	code := mapErrorTypeToGRPCCode(errType)
	return &MT5Error{
		Type:     errType,
		Message:  msg,
		Original: original,
		Code:     code,
		PTBRMsg:  ptbrMsg,
	}
}

// mapErrorTypeToGRPCCode maps MT5ErrorType to gRPC status code
func mapErrorTypeToGRPCCode(errType MT5ErrorType) codes.Code {
	switch errType {
	case ErrTypeDisconnect:
		return codes.Unavailable
	case ErrTypeTimeout:
		return codes.DeadlineExceeded
	case ErrTypeInvalidCredentials:
		return codes.Unauthenticated
	case ErrTypeMarginInsufficient:
		return codes.InvalidArgument
	case ErrTypeSymbolNotFound:
		return codes.NotFound
	case ErrTypePriceGapping:
		return codes.FailedPrecondition
	case ErrTypeQuoteUnavailable:
		return codes.Unavailable
	case ErrTypeOrderRejected:
		return codes.FailedPrecondition
	case ErrTypePositionClosed:
		return codes.FailedPrecondition
	default:
		return codes.Internal
	}
}

// DetectMT5Error analyzes error string and returns typed MT5Error
func DetectMT5Error(err error, originalMsg string) *MT5Error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	// Detect error type from message patterns
	if msg == "context deadline exceeded" {
		return NewMT5Error(ErrTypeTimeout, msg, "Tempo limite excedido", err)
	}
	if msg == "connection refused" || msg == "no such host" {
		return NewMT5Error(ErrTypeDisconnect, msg, "Terminal desconectado", err)
	}
	if msg == "invalid credentials" || msg == "authentication failed" {
		return NewMT5Error(ErrTypeInvalidCredentials, msg, "Credenciais inválidas", err)
	}
	if msg == "insufficient margin" {
		return NewMT5Error(ErrTypeMarginInsufficient, msg, "Saldo insuficiente para a margem", err)
	}
	if msg == "symbol not found" || msg == "unknown symbol" {
		return NewMT5Error(ErrTypeSymbolNotFound, msg, "Símbolo não encontrado", err)
	}
	if msg == "price gap detected" {
		return NewMT5Error(ErrTypePriceGapping, msg, "Abertura de preço detectada", err)
	}
	if msg == "quote unavailable" {
		return NewMT5Error(ErrTypeQuoteUnavailable, msg, "Cotação indisponível", err)
	}

	// Default to internal error
	return NewMT5Error(ErrTypeUnknown, fmt.Sprintf("Erro MT5: %s", msg), "Erro interno do terminal", err)
}

// IsDisconnect checks if error is disconnection
func IsDisconnect(err error) bool {
	if mt5Err, ok := err.(*MT5Error); ok {
		return mt5Err.Type == ErrTypeDisconnect
	}
	return false
}

// IsTimeout checks if error is timeout
func IsTimeout(err error) bool {
	if mt5Err, ok := err.(*MT5Error); ok {
		return mt5Err.Type == ErrTypeTimeout
	}
	return false
}

// IsMarginError checks if error is margin-related
func IsMarginError(err error) bool {
	if mt5Err, ok := err.(*MT5Error); ok {
		return mt5Err.Type == ErrTypeMarginInsufficient
	}
	return false
}
