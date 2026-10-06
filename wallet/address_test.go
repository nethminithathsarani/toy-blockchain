package wallet

import "testing"

func TestAddressFromPublicKey(t *testing.T) {
	publicKey := []byte("example-public-key")

	firstAddress := AddressFromPublicKey(publicKey)
	secondAddress := AddressFromPublicKey(publicKey)

	if firstAddress != secondAddress {
		t.Fatal("same public key produced different addresses")
	}

	if len(firstAddress) != 64 {
		t.Fatalf(
			"expected a 64-character address, got %d characters",
			len(firstAddress),
		)
	}
}
