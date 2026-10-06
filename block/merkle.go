package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"toy-blockchain/ledger"
)

type MerkleProofStep struct {
	Hash     string `json:"hash"`
	Position string `json:"position"`
}

func hashTransaction(tx ledger.Transaction) string {
	data, err := json.Marshal(tx)

	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

func CalculateMerkleRoot(
	transactions []ledger.Transaction,
) string {

	if len(transactions) == 0 {
		return ""
	}

	var hashes []string

	// Hash every transaction
	for _, tx := range transactions {

		hashes = append(
			hashes,
			hashTransaction(tx),
		)
	}

	// Build tree
	for len(hashes) > 1 {

		var newLevel []string

		for i := 0; i < len(hashes); i += 2 {

			left := hashes[i]

			right := left

			if i+1 < len(hashes) {
				right = hashes[i+1]
			}

			data := left + right

			hash := sha256.Sum256([]byte(data))

			newLevel = append(
				newLevel,
				hex.EncodeToString(hash[:]),
			)

		}

		hashes = newLevel
	}

	return hashes[0]
}
func GenerateMerkleProof(
	transactions []ledger.Transaction,
	transactionIndex int,
) ([]MerkleProofStep, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("cannot create proof for empty transaction list")
	}

	if transactionIndex < 0 ||
		transactionIndex >= len(transactions) {

		return nil, fmt.Errorf("transaction index out of range")
	}

	hashes := make([]string, len(transactions))

	for i, tx := range transactions {
		hashes[i] = hashTransaction(tx)
	}

	proof := []MerkleProofStep{}
	currentIndex := transactionIndex

	for len(hashes) > 1 {
		var siblingIndex int
		var position string

		if currentIndex%2 == 0 {
			// Current hash is on the left.
			// Its sibling must be placed on the right.
			siblingIndex = currentIndex + 1
			position = "right"

			// Duplicate the final hash when the level has an odd count.
			if siblingIndex >= len(hashes) {
				siblingIndex = currentIndex
			}
		} else {
			// Current hash is on the right.
			// Its sibling must be placed on the left.
			siblingIndex = currentIndex - 1
			position = "left"
		}

		proof = append(proof, MerkleProofStep{
			Hash:     hashes[siblingIndex],
			Position: position,
		})

		var newLevel []string

		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i]
			right := left

			if i+1 < len(hashes) {
				right = hashes[i+1]
			}

			combinedHash := sha256.Sum256(
				[]byte(left + right),
			)

			newLevel = append(
				newLevel,
				hex.EncodeToString(combinedHash[:]),
			)
		}

		hashes = newLevel
		currentIndex /= 2
	}

	return proof, nil
}
func VerifyMerkleProof(
	transaction ledger.Transaction,
	proof []MerkleProofStep,
	expectedRoot string,
) bool {
	if expectedRoot == "" {
		return false
	}

	currentHash := hashTransaction(transaction)

	for _, step := range proof {
		var data string

		switch step.Position {
		case "left":
			data = step.Hash + currentHash

		case "right":
			data = currentHash + step.Hash

		default:
			return false
		}

		combinedHash := sha256.Sum256([]byte(data))

		currentHash = hex.EncodeToString(
			combinedHash[:],
		)
	}

	return currentHash == expectedRoot
}
