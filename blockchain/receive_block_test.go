package blockchain

import (
	"testing"

	"toy-blockchain/ledger"
)

func TestAddReceivedBlock(t *testing.T) {
	chainA := NewBlockchain()
	chainB := NewBlockchain()

	tx := createSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)
	if !chainA.AddTransaction(tx) {
		t.Fatal("expected Node A transaction to be accepted")
	}

	if !chainB.AddTransaction(tx) {
		t.Fatal("expected Node B transaction to be accepted")
	}
	ledgerA := ledger.NewLedger(
		CalculateBalances(
			chainA.Blocks,
			chainA.InitialBalances,
		),
	)

	chainA.MinePendingTransactions(ledgerA)

	if len(chainA.Blocks) != 2 {
		t.Fatalf(
			"expected Node A to have 2 blocks, got %d",
			len(chainA.Blocks),
		)
	}

	minedBlock := chainA.Blocks[1]
	if err := chainB.AddReceivedBlock(minedBlock); err != nil {
		t.Fatalf("expected Node B to accept block: %v", err)
	}
	if len(chainB.Blocks) != 2 {
		t.Fatalf(
			"expected Node B to have 2 blocks, got %d",
			len(chainB.Blocks),
		)
	}

	if chainB.Blocks[1].Hash != minedBlock.Hash {
		t.Fatal("expected Node B to append the received block")
	}

	if len(chainB.PendingTransactions) != 0 {
		t.Fatalf(
			"expected Node B pending pool to be empty, got %d",
			len(chainB.PendingTransactions),
		)
	}
}
