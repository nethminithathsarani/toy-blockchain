package blockchain

import (
	"testing"

	"toy-blockchain/ledger"
)

func TestDoubleSpendPrevention(t *testing.T) {

	bc := NewBlockchain()

	balances := CalculateBalances(bc.Blocks, bc.InitialBalances)

	l := ledger.NewLedger(balances)

	tx1 := createSignedTransaction(
		t,
		"Alice",
		"Bob",
		80,
		1,
	)

	tx2 := createSignedTransaction(
		t,
		"Alice",
		"Charlie",
		80,
		2,
	)

	bc.AddTransaction(tx1)
	bc.AddTransaction(tx2)

	bc.MinePendingTransactions(l)

	minedBlock := bc.Blocks[len(bc.Blocks)-1]

	if len(minedBlock.Transactions) != 1 {
		t.Fatalf(
			"expected only 1 transaction to be mined, got %d",
			len(minedBlock.Transactions),
		)
	}
}
