package node

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"toy-blockchain/blockchain"
	"toy-blockchain/ledger"
	"toy-blockchain/wallet"
)

func TestTransactionEndpointRejectsDuplicate(t *testing.T) {
	chain := blockchain.NewBlockchain()
	n := NewNode(Config{}, chain)

	w, err := wallet.NewWallet()

	if err != nil {
		t.Fatal(err)
	}

	tx, err := wallet.SignTransaction(
		ledger.Transaction{
			Sender:   "Alice",
			Receiver: "Bob",
			Amount:   20,
			Nonce:    1,
		},
		w,
	)

	if err != nil {
		t.Fatal(err)
	}

	transactionJSON, err := json.Marshal(tx)

	if err != nil {
		t.Fatal(err)
	}
	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/transactions",
		bytes.NewReader(transactionJSON),
	)

	firstResponse := httptest.NewRecorder()

	n.Handler().ServeHTTP(firstResponse, firstRequest)

	if firstResponse.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status 202, got %d: %s",
			firstResponse.Code,
			firstResponse.Body.String(),
		)
	}

	if len(chain.PendingTransactions) != 1 {
		t.Fatalf(
			"expected 1 pending transaction, got %d",
			len(chain.PendingTransactions),
		)
	}
	secondRequest := httptest.NewRequest(
		http.MethodPost,
		"/transactions",
		bytes.NewReader(transactionJSON),
	)

	secondResponse := httptest.NewRecorder()

	n.Handler().ServeHTTP(secondResponse, secondRequest)

	if secondResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected duplicate status 200, got %d",
			secondResponse.Code,
		)
	}

	if len(chain.PendingTransactions) != 1 {
		t.Fatalf(
			"expected duplicate not to be added; got %d transactions",
			len(chain.PendingTransactions),
		)
	}
}
func TestTransactionGossipBetweenTwoNodes(t *testing.T) {
	chainB := blockchain.NewBlockchain()

	nodeB := NewNode(
		Config{
			Address: "node-b",
			Peers:   []string{"node-a"},
		},
		chainB,
	)

	serverB := httptest.NewServer(nodeB.Handler())
	defer serverB.Close()

	chainA := blockchain.NewBlockchain()

	nodeA := NewNode(
		Config{
			Address: "node-a",
			Peers:   []string{serverB.URL},
		},
		chainA,
	)
	w, err := wallet.NewWallet()

	if err != nil {
		t.Fatal(err)
	}

	tx, err := wallet.SignTransaction(
		ledger.Transaction{
			Sender:   "Alice",
			Receiver: "Bob",
			Amount:   20,
			Nonce:    1,
		},
		w,
	)

	if err != nil {
		t.Fatal(err)
	}

	transactionJSON, err := json.Marshal(tx)

	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/transactions",
		bytes.NewReader(transactionJSON),
	)

	response := httptest.NewRecorder()

	nodeA.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf(
			"expected Node A status 202, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}
	if len(chainA.PendingTransactions) != 1 {
		t.Fatalf(
			"expected Node A to have 1 pending transaction, got %d",
			len(chainA.PendingTransactions),
		)
	}

	if len(chainB.PendingTransactions) != 1 {
		t.Fatalf(
			"expected Node B to receive 1 pending transaction, got %d",
			len(chainB.PendingTransactions),
		)
	}
}
