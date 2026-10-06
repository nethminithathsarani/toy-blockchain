package blockchain

import (
	"testing"

	"toy-blockchain/ledger"
)

func TestResolveForkReturnsOrphanedTransaction(t *testing.T) {
	localChain := NewBlockchain()

	orphanedTx := createSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)

	if !localChain.AddTransaction(orphanedTx) {
		t.Fatal("expected local transaction to be accepted")
	}
	localLedger := ledger.NewLedger(
		CalculateBalances(
			localChain.Blocks,
			localChain.InitialBalances,
		),
	)

	localChain.MinePendingTransactions(localLedger)

	if len(localChain.Blocks) != 2 {
		t.Fatal("expected local chain to contain one mined block")
	}
	candidateChain := NewBlockchain()

	candidateTx := createSignedTransaction(
		t,
		"Charlie",
		"Bob",
		10,
		1,
	)

	if !candidateChain.AddTransaction(candidateTx) {
		t.Fatal("expected candidate transaction to be accepted")
	}

	candidateLedger := ledger.NewLedger(
		CalculateBalances(
			candidateChain.Blocks,
			candidateChain.InitialBalances,
		),
	)

	candidateChain.MinePendingTransactions(
		candidateLedger,
	)

	candidateChain.AddBlock(nil)
	if err := localChain.ResolveFork(
		candidateChain.Blocks,
	); err != nil {
		t.Fatalf("expected fork resolution to succeed: %v", err)
	}
	if len(localChain.Blocks) != 3 {
		t.Fatalf(
			"expected selected chain to have 3 blocks, got %d",
			len(localChain.Blocks),
		)
	}

	if localChain.Blocks[2].Hash != candidateChain.Blocks[2].Hash {
		t.Fatal("expected candidate chain to be selected")
	}
	if len(localChain.PendingTransactions) != 1 {
		t.Fatalf(
			"expected 1 orphaned transaction in pending pool, got %d",
			len(localChain.PendingTransactions),
		)
	}

	if localChain.PendingTransactions[0].ID != orphanedTx.ID {
		t.Fatal("expected local orphaned transaction to return to pending")
	}
}
