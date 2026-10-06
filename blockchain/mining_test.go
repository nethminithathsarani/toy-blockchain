package blockchain

import (
	"strings"
	"testing"
	"toy-blockchain/ledger"
)

func TestMiningDifficulty(t *testing.T) {

	bc := NewBlockchain()

	tx := createSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)

	if !bc.AddTransaction(tx) {
		t.Fatal("expected transaction to be accepted")
	}

	l := ledger.NewLedger(CalculateBalances(bc.Blocks, bc.InitialBalances))

	bc.MinePendingTransactions(l)

	if len(bc.Blocks) < 2 {
		t.Fatal("expected a mined block to be added")
	}

	minedBlock := bc.Blocks[len(bc.Blocks)-1]

	difficulty := 4

	prefix := strings.Repeat("0", difficulty)

	if !strings.HasPrefix(minedBlock.Hash, prefix) {
		t.Fatalf(
			"hash %s does not satisfy difficulty %d",
			minedBlock.Hash,
			difficulty,
		)
	}
}

func TestRejectPendingDoubleSpend(t *testing.T) {
	bc := NewBlockchain()

	if !bc.AddTransaction(createSignedTransaction(
		t,
		"Alice",
		"Bob",
		80,
		1,
	)) {
		t.Fatal("expected first transaction to be accepted")
	}

	if bc.AddTransaction(createSignedTransaction(
		t,
		"Alice",
		"Bob",
		30,
		2,
	)) {
		t.Fatal("expected overspending transaction to be rejected")
	}

	if len(bc.PendingTransactions) != 1 {
		t.Fatalf("expected 1 pending transaction, got %d", len(bc.PendingTransactions))
	}
}
