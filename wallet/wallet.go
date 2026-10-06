package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
)

type Wallet struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func NewWallet() (*Wallet, error) {
	publicKey, privateKey, err := GenerateEd25519KeyPair()

	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (w *Wallet) GetPublicKey() string {
	return hex.EncodeToString(w.PublicKey)
}

func (w *Wallet) GetAddress() string {
	return AddressFromPublicKey(w.PublicKey)
}
