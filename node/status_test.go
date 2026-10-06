package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"toy-blockchain/blockchain"
)

func TestStatusEndpoint(t *testing.T) {
	chain := blockchain.NewBlockchain()

	n := NewNode(
		Config{Address: "localhost:8001"},
		chain,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/status",
		nil,
	)

	response := httptest.NewRecorder()
	n.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.Code,
		)
	}
	var status StatusResponse

	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if status.Height != 0 {
		t.Fatalf("expected height 0, got %d", status.Height)
	}

	if status.HeadHash != chain.Blocks[0].Hash {
		t.Fatal("expected head hash to match genesis block")
	}
}
