package wallet

import (
	"testing"

	"toy-blockchain/ledger"
)

func TestVerifyTransactionProtectsAddressAndNonce(t *testing.T) {
	// Keep the original wallet registry and restore it after the test.
	originalWallets := Wallets
	Wallets = make(map[string]*Wallet)

	defer func() {
		Wallets = originalWallets
	}()

	aliceWallet, err := NewWallet()
	if err != nil {
		t.Fatalf("failed to create Alice wallet: %v", err)
	}

	Wallets["Alice"] = aliceWallet

	tx := ledger.Transaction{
		Sender:   "Alice",
		Receiver: "Bob",
		Amount:   20,
		Nonce:    1,
	}

	signedTx, err := SignTransaction(tx, aliceWallet)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	t.Run("valid transaction", func(t *testing.T) {
		if !VerifyTransaction(signedTx) {
			t.Error("expected valid signed transaction to pass")
		}
	})

	t.Run("changed sender address", func(t *testing.T) {
		tamperedTx := signedTx
		tamperedTx.SenderAddress = "fake-address"

		if VerifyTransaction(tamperedTx) {
			t.Error("expected changed sender address to fail")
		}
	})

	t.Run("changed nonce", func(t *testing.T) {
		tamperedTx := signedTx
		tamperedTx.Nonce++

		if VerifyTransaction(tamperedTx) {
			t.Error("expected changed nonce to fail")
		}
	})
}
