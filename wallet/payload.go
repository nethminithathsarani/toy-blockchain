package wallet

import (
	"encoding/json"

	"toy-blockchain/ledger"
)

type transactionPayload struct {
	Sender        string `json:"sender"`
	SenderAddress string `json:"sender_address"`
	Receiver      string `json:"receiver"`
	Amount        int    `json:"amount"`
	Nonce         int    `json:"nonce"`
}

func TransactionSigningPayload(
	tx ledger.Transaction,
) ([]byte, error) {

	return json.Marshal(transactionPayload{
		Sender:        tx.Sender,
		SenderAddress: tx.SenderAddress,
		Receiver:      tx.Receiver,
		Amount:        tx.Amount,
		Nonce:         tx.Nonce,
	})
}
