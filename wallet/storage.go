package wallet

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type walletFile struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func SaveWallet(name string, w *Wallet) error {
	if err := os.MkdirAll("wallets", 0755); err != nil {
		return err
	}

	data := walletFile{
		PrivateKey: hex.EncodeToString(w.PrivateKey),
		PublicKey:  hex.EncodeToString(w.PublicKey),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		return err
	}

	path := filepath.Join("wallets", name+".json")

	return os.WriteFile(path, jsonBytes, 0600)
}

func LoadWallet(name string) (*Wallet, error) {
	path := filepath.Join("wallets", name+".json")

	jsonBytes, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var data walletFile

	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, err
	}

	privateKeyBytes, err := hex.DecodeString(data.PrivateKey)

	if err != nil {
		return nil, fmt.Errorf("invalid private key encoding: %w", err)
	}

	publicKeyBytes, err := hex.DecodeString(data.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("invalid public key encoding: %w", err)
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf(
			"invalid private key size: got %d bytes",
			len(privateKeyBytes),
		)
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf(
			"invalid public key size: got %d bytes",
			len(publicKeyBytes),
		)
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := ed25519.PublicKey(publicKeyBytes)

	derivedPublicKey := privateKey.Public().(ed25519.PublicKey)

	if !bytes.Equal(publicKey, derivedPublicKey) {
		return nil, fmt.Errorf(
			"public key does not match private key",
		)
	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}
