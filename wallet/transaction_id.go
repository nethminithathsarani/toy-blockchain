package wallet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"toy-blockchain/ledger"
)

type transactionIDPayload struct {
	Sender        string `json:"sender"`
	SenderAddress string `json:"sender_address"`
	Receiver      string `json:"receiver"`
	Amount        int    `json:"amount"`
	Nonce         int    `json:"nonce"`
	PublicKey     string `json:"public_key"`
	Signature     string `json:"signature"`
}

func TransactionID(
	tx ledger.Transaction,
) (string, error) {
	data, err := json.Marshal(transactionIDPayload{
		Sender:        tx.Sender,
		SenderAddress: tx.SenderAddress,
		Receiver:      tx.Receiver,
		Amount:        tx.Amount,
		Nonce:         tx.Nonce,
		PublicKey:     tx.PublicKey,
		Signature:     tx.Signature,
	})

	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:]), nil
}
