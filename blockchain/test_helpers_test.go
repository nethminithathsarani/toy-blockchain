package blockchain

import (
	"testing"

	"toy-blockchain/ledger"
	"toy-blockchain/wallet"
)

// createSignedTransaction signs with -- and registers into -- the real
// wallet.Wallets registry (not a private map disconnected from it), because
// VerifyTransaction now looks up the sender's registered wallet to confirm
// the embedded public key actually belongs to that sender. Tests need to
// exercise the same registry the production code checks against.
func createSignedTransaction(
	t *testing.T,
	sender string,
	receiver string,
	amount int,
	nonce int,
) ledger.Transaction {

	if _, exists := wallet.Wallets[sender]; !exists {

		w, err := wallet.NewWallet()

		if err != nil {
			t.Fatal(err)
		}

		wallet.Wallets[sender] = w
	}

	tx := ledger.Transaction{
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
		Nonce:    nonce,
	}

	signed, err := wallet.SignTransaction(tx, wallet.Wallets[sender])
	if err != nil {
		t.Fatal(err)
	}

	return signed
}
