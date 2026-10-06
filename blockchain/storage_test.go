package blockchain

import (
	"os"
	"path/filepath"
	"testing"
	"toy-blockchain/ledger"
)

func TestSaveAndLoadPreservesBalances(t *testing.T) {
	bc := NewBlockchain()

	if !bc.AddTransaction(createSignedTransaction(
		t,
		"Alice",
		"Bob",
		25,
		1,
	)) {
		t.Fatal("expected transaction to be accepted")
	}

	l := ledger.NewLedger(CalculateBalances(bc.Blocks, bc.InitialBalances))
	bc.MinePendingTransactions(l)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "chain.json")

	if err := bc.SaveToFile(filePath); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	balances := CalculateBalances(loaded.Blocks, loaded.InitialBalances)

	if balances["Alice"] != 75 {
		t.Fatalf("expected Alice to have 75, got %v", balances["Alice"])
	}

	if balances["Bob"] != 125 {
		t.Fatalf("expected Bob to have 125, got %v", balances["Bob"])
	}

	if balances["Charlie"] != 100 {
		t.Fatalf("expected Charlie to have 100, got %v", balances["Charlie"])
	}

	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected saved file to exist: %v", err)
	}
}
