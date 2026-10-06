package wallet

import (
	"crypto/ed25519"
	"testing"
)

func TestGenerateEd25519KeyPair(t *testing.T) {
	publicKey, privateKey, err := GenerateEd25519KeyPair()

	if err != nil {
		t.Fatal("failed to generate Ed25519 key pair:", err)
	}

	if len(publicKey) != ed25519.PublicKeySize {
		t.Fatalf(
			"expected public key size %d, got %d",
			ed25519.PublicKeySize,
			len(publicKey),
		)
	}

	if len(privateKey) != ed25519.PrivateKeySize {
		t.Fatalf(
			"expected private key size %d, got %d",
			ed25519.PrivateKeySize,
			len(privateKey),
		)
	}
}
func TestEd25519SignAndVerify(t *testing.T) {
	publicKey, privateKey, err := GenerateEd25519KeyPair()

	if err != nil {
		t.Fatal("failed to generate key pair:", err)
	}

	message := []byte("Alice sends 20 coins to Bob")

	signature := SignEd25519(message, privateKey)

	if !VerifyEd25519(message, signature, publicKey) {
		t.Fatal("valid signature was rejected")
	}
}
func TestEd25519RejectsChangedMessage(t *testing.T) {
	publicKey, privateKey, err := GenerateEd25519KeyPair()

	if err != nil {
		t.Fatal("failed to generate key pair:", err)
	}

	originalMessage := []byte("Alice sends 20 coins to Bob")
	changedMessage := []byte("Alice sends 200 coins to Bob")

	signature := SignEd25519(originalMessage, privateKey)

	if VerifyEd25519(changedMessage, signature, publicKey) {
		t.Fatal("signature was accepted for a changed message")
	}
}
