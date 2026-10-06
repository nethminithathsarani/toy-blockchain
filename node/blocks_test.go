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

func TestMineAndGossipBlockBetweenTwoNodes(t *testing.T) {
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

	transactionRequest := httptest.NewRequest(
		http.MethodPost,
		"/transactions",
		bytes.NewReader(transactionJSON),
	)

	transactionResponse := httptest.NewRecorder()

	nodeA.Handler().ServeHTTP(
		transactionResponse,
		transactionRequest,
	)

	if transactionResponse.Code != http.StatusAccepted {
		t.Fatalf(
			"expected transaction status 202, got %d: %s",
			transactionResponse.Code,
			transactionResponse.Body.String(),
		)
	}
	mineRequest := httptest.NewRequest(
		http.MethodPost,
		"/mine",
		nil,
	)

	mineResponse := httptest.NewRecorder()

	nodeA.Handler().ServeHTTP(
		mineResponse,
		mineRequest,
	)

	if mineResponse.Code != http.StatusCreated {
		t.Fatalf(
			"expected mining status 201, got %d: %s",
			mineResponse.Code,
			mineResponse.Body.String(),
		)
	}
	if len(chainA.Blocks) != 2 {
		t.Fatalf(
			"expected Node A to have 2 blocks, got %d",
			len(chainA.Blocks),
		)
	}

	if len(chainB.Blocks) != 2 {
		t.Fatalf(
			"expected Node B to have 2 blocks, got %d",
			len(chainB.Blocks),
		)
	}

	if chainA.Blocks[1].Hash != chainB.Blocks[1].Hash {
		t.Fatal("expected both nodes to have the same block")
	}

	if len(chainA.PendingTransactions) != 0 {
		t.Fatal("expected Node A pending pool to be empty")
	}

	if len(chainB.PendingTransactions) != 0 {
		t.Fatal("expected Node B pending pool to be empty")
	}
}
func TestCompetingBlockTriggersAutomaticSynchronization(t *testing.T) {
	// Create the local chain with one transaction and one mined block.
	localChain := blockchain.NewBlockchain()

	localTransaction := createNodeSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)

	if !localChain.AddTransaction(localTransaction) {
		t.Fatal("expected local transaction to be accepted")
	}

	localLedger := ledger.NewLedger(
		blockchain.CalculateBalances(
			localChain.Blocks,
			localChain.InitialBalances,
		),
	)

	localChain.MinePendingTransactions(localLedger)

	// Create a different and stronger peer chain.
	peerChain := blockchain.NewBlockchain()

	peerTransaction := createNodeSignedTransaction(
		t,
		"Charlie",
		"Bob",
		10,
		1,
	)

	if !peerChain.AddTransaction(peerTransaction) {
		t.Fatal("expected peer transaction to be accepted")
	}

	peerLedger := ledger.NewLedger(
		blockchain.CalculateBalances(
			peerChain.Blocks,
			peerChain.InitialBalances,
		),
	)

	peerChain.MinePendingTransactions(peerLedger)
	peerChain.AddBlock(nil)

	peerNode := NewNode(Config{}, peerChain)

	peerServer := httptest.NewServer(peerNode.Handler())
	defer peerServer.Close()

	localNode := NewNode(Config{}, localChain)

	// Send the peer's newest block to the local node.
	// It cannot connect directly because Peer Block 1 is missing.
	receivedBlock := peerChain.Blocks[len(peerChain.Blocks)-1]

	blockJSON, err := json.Marshal(receivedBlock)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/blocks",
		bytes.NewReader(blockJSON),
	)

	// Tell the local node which peer sent the block.
	request.Header.Set("X-Node-Address", peerServer.URL)

	response := httptest.NewRecorder()

	localNode.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status 202, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	// The local node should now have adopted the peer's chain.
	if len(localChain.Blocks) != len(peerChain.Blocks) {
		t.Fatalf(
			"expected %d blocks, got %d",
			len(peerChain.Blocks),
			len(localChain.Blocks),
		)
	}

	localHead := localChain.Blocks[len(localChain.Blocks)-1].Hash
	peerHead := peerChain.Blocks[len(peerChain.Blocks)-1].Hash

	if localHead != peerHead {
		t.Fatal("expected local node to adopt the peer chain")
	}

	// The transaction from the orphaned local block should return.
	if len(localChain.PendingTransactions) != 1 {
		t.Fatalf(
			"expected 1 orphaned transaction, got %d",
			len(localChain.PendingTransactions),
		)
	}

	if localChain.PendingTransactions[0].ID != localTransaction.ID {
		t.Fatal("expected local orphaned transaction to return to pending")
	}
}
