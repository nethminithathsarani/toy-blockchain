package wallet

import (
	"encoding/hex"

	"toy-blockchain/ledger"
)

// VerifyTransaction verifies the sender identity, address,
// transaction payload, and Ed25519 signature.
func VerifyTransaction(tx ledger.Transaction) bool {
	publicKeyBytes, err := hex.DecodeString(tx.PublicKey)
	if err != nil {
		return false
	}

	expectedAddress := AddressFromPublicKey(publicKeyBytes)

	if tx.SenderAddress != expectedAddress {
		return false
	}
	expectedID, err := TransactionID(tx)

	if err != nil {
		return false
	}

	if tx.ID != expectedID {
		return false
	}
	payload, err := TransactionSigningPayload(tx)
	if err != nil {
		return false
	}

	return Verify(
		string(payload),
		tx.Signature,
		tx.PublicKey,
	)
}
