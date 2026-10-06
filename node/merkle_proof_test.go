package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"toy-blockchain/blockchain"
	"toy-blockchain/ledger"
)

func TestMerkleProofEndpoint(t *testing.T) {
	chain := blockchain.NewBlockchain()

	transaction := createNodeSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)

	if !chain.AddTransaction(transaction) {
		t.Fatal("expected transaction to be accepted")
	}

	currentLedger := ledger.NewLedger(
		blockchain.CalculateBalances(
			chain.Blocks,
			chain.InitialBalances,
		),
	)

	chain.MinePendingTransactions(currentLedger)

	n := NewNode(Config{}, chain)

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/merkle-proof?block=1&transaction=%s",
			transaction.ID,
		),
		nil,
	)

	response := httptest.NewRecorder()

	n.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	var proofResponse MerkleProofResponse

	if err := json.NewDecoder(response.Body).Decode(
		&proofResponse,
	); err != nil {
		t.Fatalf(
			"failed to decode proof response: %v",
			err,
		)
	}

	if proofResponse.Transaction.ID != transaction.ID {
		t.Fatal("expected requested transaction in response")
	}

	if proofResponse.MerkleRoot != chain.Blocks[1].MerkleRoot {
		t.Fatal("expected response Merkle root to match block")
	}

	if !proofResponse.Verified {
		t.Fatal("expected returned Merkle proof to be valid")
	}
}
