package wallet

import (
	"crypto/sha256"
	"encoding/hex"
)

// AddressFromPublicKey creates a blockchain address from a public key.
func AddressFromPublicKey(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)

	return hex.EncodeToString(hash[:])
}
