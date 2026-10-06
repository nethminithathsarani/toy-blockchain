package node

import (
	"testing"

	"toy-blockchain/ledger"
	"toy-blockchain/wallet"
)

func createNodeSignedTransaction(
	t *testing.T,
	sender string,
	receiver string,
	amount int,
	nonce int,
) ledger.Transaction {
	t.Helper()

	senderWallet, err := wallet.NewWallet()

	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	tx := ledger.Transaction{
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
		Nonce:    nonce,
	}

	signedTx, err := wallet.SignTransaction(
		tx,
		senderWallet,
	)

	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	return signedTx
}
