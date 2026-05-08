package models

// ModifyOrderResult represents the result of modify order operation
type ModifyOrderResult struct {
	Ticket   int64
	Status   string
	ErrorMsg string
}
