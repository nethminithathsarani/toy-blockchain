package wallet

import "crypto/ed25519"

func Sign(
	data string,
	privateKey ed25519.PrivateKey,
) (string, error) {
	signature := SignEd25519(
		[]byte(data),
		privateKey,
	)

	return signature, nil
}
