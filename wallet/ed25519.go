package wallet

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

// GenerateEd25519KeyPair creates a new public and private key pair.
func GenerateEd25519KeyPair() (
	ed25519.PublicKey,
	ed25519.PrivateKey,
	error,
) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)

	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

// SignEd25519 signs data using a private key.
func SignEd25519(
	data []byte,
	privateKey ed25519.PrivateKey,
) string {
	signature := ed25519.Sign(privateKey, data)

	return hex.EncodeToString(signature)
}

// VerifyEd25519 checks a signature using a public key.
func VerifyEd25519(
	data []byte,
	signatureText string,
	publicKey ed25519.PublicKey,
) bool {
	signature, err := hex.DecodeString(signatureText)

	if err != nil {
		return false
	}

	return ed25519.Verify(publicKey, data, signature)
}
