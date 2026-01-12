package http

type CreateTransactionResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
