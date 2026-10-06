package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
)

func Verify(
	data string,
	signature string,
	publicKeyText string,
) bool {
	publicKeyBytes, err := hex.DecodeString(publicKeyText)

	if err != nil {
		return false
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return false
	}

	publicKey := ed25519.PublicKey(publicKeyBytes)

	return VerifyEd25519(
		[]byte(data),
		signature,
		publicKey,
	)
}
