package block_test

import (
	"testing"

	"toy-blockchain/block"
	"toy-blockchain/ledger"
)

func TestMerkleRoot(t *testing.T) {

	transactions := []ledger.Transaction{

		{
			Sender:   "Alice",
			Receiver: "Bob",
			Amount:   10,
		},

		{
			Sender:   "Bob",
			Receiver: "Charlie",
			Amount:   5,
		},
	}

	root := block.CalculateMerkleRoot(
		transactions,
	)

	if root == "" {

		t.Fatal(
			"Merkle root should not be empty",
		)

	}

	t.Log(
		"Merkle Root:",
		root,
	)

}
func merkleProofTransactions() []ledger.Transaction {
	return []ledger.Transaction{
		{
			Sender:   "Alice",
			Receiver: "Bob",
			Amount:   10,
		},
		{
			Sender:   "Bob",
			Receiver: "Charlie",
			Amount:   5,
		},
		{
			Sender:   "Charlie",
			Receiver: "Alice",
			Amount:   3,
		},
	}
}

func TestValidMerkleProof(t *testing.T) {
	transactions := merkleProofTransactions()

	root := block.CalculateMerkleRoot(transactions)

	// Generate and verify a proof for every transaction.
	for index, transaction := range transactions {
		proof, err := block.GenerateMerkleProof(
			transactions,
			index,
		)

		if err != nil {
			t.Fatalf(
				"failed to generate proof for index %d: %v",
				index,
				err,
			)
		}

		if !block.VerifyMerkleProof(
			transaction,
			proof,
			root,
		) {
			t.Fatalf(
				"expected valid proof for transaction index %d",
				index,
			)
		}
	}
}

func TestMerkleProofRejectsChangedTransaction(t *testing.T) {
	transactions := merkleProofTransactions()

	root := block.CalculateMerkleRoot(transactions)

	proof, err := block.GenerateMerkleProof(
		transactions,
		0,
	)

	if err != nil {
		t.Fatal(err)
	}

	changedTransaction := transactions[0]
	changedTransaction.Amount = 99

	if block.VerifyMerkleProof(
		changedTransaction,
		proof,
		root,
	) {
		t.Fatal("expected changed transaction proof to fail")
	}
}

func TestMerkleProofRejectsInvalidIndex(t *testing.T) {
	transactions := merkleProofTransactions()

	_, err := block.GenerateMerkleProof(
		transactions,
		len(transactions),
	)

	if err == nil {
		t.Fatal("expected invalid transaction index error")
	}
}
func TestMerkleProofRejectsChangedNonce(t *testing.T) {
	transactions := merkleProofTransactions()

	root := block.CalculateMerkleRoot(transactions)

	proof, err := block.GenerateMerkleProof(
		transactions,
		0,
	)

	if err != nil {
		t.Fatal(err)
	}

	changedTransaction := transactions[0]
	changedTransaction.Nonce++

	if block.VerifyMerkleProof(
		changedTransaction,
		proof,
		root,
	) {
		t.Fatal(
			"expected transaction with changed nonce to fail proof verification",
		)
	}
}
