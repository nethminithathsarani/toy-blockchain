package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"toy-blockchain/blockchain"
)

func TestPeersEndpoint(t *testing.T) {
	chain := blockchain.NewBlockchain()

	n := NewNode(
		Config{
			Address: "localhost:8001",
			Peers: []string{
				"localhost:8002",
				"localhost:8003",
			},
		},
		chain,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/peers",
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

	var peersResponse PeersResponse

	if err := json.NewDecoder(response.Body).Decode(
		&peersResponse,
	); err != nil {
		t.Fatalf("failed to decode peers response: %v", err)
	}

	if len(peersResponse.Peers) != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			len(peersResponse.Peers),
		)
	}
}
func TestMempoolEndpoint(t *testing.T) {
	chain := blockchain.NewBlockchain()
	n := NewNode(Config{}, chain)

	request := httptest.NewRequest(
		http.MethodGet,
		"/mempool",
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

	var mempoolResponse MempoolResponse

	if err := json.NewDecoder(response.Body).Decode(
		&mempoolResponse,
	); err != nil {
		t.Fatalf("failed to decode mempool response: %v", err)
	}

	if mempoolResponse.Size != 0 {
		t.Fatalf(
			"expected mempool size 0, got %d",
			mempoolResponse.Size,
		)
	}
}
func TestBalancesEndpoint(t *testing.T) {
	chain := blockchain.NewBlockchain()
	n := NewNode(Config{}, chain)

	request := httptest.NewRequest(
		http.MethodGet,
		"/balances",
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

	var balancesResponse BalancesResponse

	if err := json.NewDecoder(response.Body).Decode(
		&balancesResponse,
	); err != nil {
		t.Fatalf("failed to decode balances response: %v", err)
	}

	if balancesResponse.Balances["Alice"] != 100 {
		t.Fatalf(
			"expected Alice balance 100, got %d",
			balancesResponse.Balances["Alice"],
		)
	}

	if balancesResponse.Balances["Bob"] != 100 {
		t.Fatalf(
			"expected Bob balance 100, got %d",
			balancesResponse.Balances["Bob"],
		)
	}

	if balancesResponse.Balances["Charlie"] != 100 {
		t.Fatalf(
			"expected Charlie balance 100, got %d",
			balancesResponse.Balances["Charlie"],
		)
	}
}
func TestChainEndpoint(t *testing.T) {
	chain := blockchain.NewBlockchain()
	n := NewNode(Config{}, chain)

	request := httptest.NewRequest(
		http.MethodGet,
		"/chain",
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

	var chainResponse ChainResponse

	if err := json.NewDecoder(response.Body).Decode(
		&chainResponse,
	); err != nil {
		t.Fatalf("failed to decode chain response: %v", err)
	}

	if len(chainResponse.Blocks) != 1 {
		t.Fatalf(
			"expected 1 block, got %d",
			len(chainResponse.Blocks),
		)
	}

	if chainResponse.Blocks[0].Index != 0 {
		t.Fatalf(
			"expected genesis index 0, got %d",
			chainResponse.Blocks[0].Index,
		)
	}

	if chainResponse.Blocks[0].Hash != chain.Blocks[0].Hash {
		t.Fatal("expected returned genesis hash to match blockchain")
	}
}
