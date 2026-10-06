package wallet

import (
	"testing"

	"toy-blockchain/ledger"
)

func TestTransactionIDIsDeterministic(t *testing.T) {
	w, err := NewWallet()

	if err != nil {
		t.Fatal(err)
	}

	tx := ledger.Transaction{
		Sender:   "Alice",
		Receiver: "Bob",
		Amount:   20,
		Nonce:    1,
	}

	first, err := SignTransaction(tx, w)

	if err != nil {
		t.Fatal(err)
	}

	second, err := SignTransaction(tx, w)

	if err != nil {
		t.Fatal(err)
	}
	if first.ID == "" {
		t.Fatal("expected transaction ID to be generated")
	}

	if first.ID != second.ID {
		t.Fatal("expected identical signed transactions to have the same ID")
	}

	if len(first.ID) != 64 {
		t.Fatalf(
			"expected 64-character transaction ID, got %d",
			len(first.ID),
		)
	}
}
func TestVerifyTransactionRejectsFakeID(t *testing.T) {
	w, err := NewWallet()

	if err != nil {
		t.Fatal(err)
	}

	tx := ledger.Transaction{
		Sender:   "Alice",
		Receiver: "Bob",
		Amount:   20,
		Nonce:    1,
	}

	signed, err := SignTransaction(tx, w)

	if err != nil {
		t.Fatal(err)
	}

	signed.ID = "fake-transaction-id"

	if VerifyTransaction(signed) {
		t.Fatal("expected transaction with fake ID to be rejected")
	}
}
