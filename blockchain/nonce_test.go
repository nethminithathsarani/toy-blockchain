package blockchain

import (
	"strings"
	"testing"

	"toy-blockchain/ledger"
)

func TestRejectReplayedTransaction(t *testing.T) {
	bc := NewBlockchain()

	tx := createSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)
	if !bc.AddTransaction(tx) {
		t.Fatal("expected first transaction submission to be accepted")
	}
	if bc.AddTransaction(tx) {
		t.Fatal("expected replayed transaction to be rejected")
	}
	if len(bc.PendingTransactions) != 1 {
		t.Fatalf(
			"expected 1 pending transaction, got %d",
			len(bc.PendingTransactions),
		)
	}
}
func TestValidateChainRejectsRepeatedNonce(t *testing.T) {
	bc := NewBlockchain()

	tx := createSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)

	// Simulate receiving a block containing the same transaction twice.
	bc.AddBlock([]ledger.Transaction{tx, tx})

	err := bc.ValidateChain()

	if err == nil {
		t.Fatal("expected chain with repeated nonce to be rejected")
	}

	if !strings.Contains(err.Error(), "invalid nonce") {
		t.Fatalf("expected invalid nonce error, got: %v", err)
	}
}
