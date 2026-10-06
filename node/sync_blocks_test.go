package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"toy-blockchain/blockchain"
	"toy-blockchain/ledger"
)

func TestGetMissingBlocks(t *testing.T) {
	chain := blockchain.NewBlockchain()

	chain.AddBlock(nil)

	n := NewNode(Config{}, chain)

	request := httptest.NewRequest(
		http.MethodGet,
		"/blocks?from=1",
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
	var missingResponse MissingBlocksResponse

	if err := json.NewDecoder(response.Body).Decode(
		&missingResponse,
	); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(missingResponse.Blocks) != 1 {
		t.Fatalf(
			"expected 1 missing block, got %d",
			len(missingResponse.Blocks),
		)
	}

	if missingResponse.Blocks[0].Index != 1 {
		t.Fatalf(
			"expected block index 1, got %d",
			missingResponse.Blocks[0].Index,
		)
	}
}
func TestSyncFromPeer(t *testing.T) {
	peerChain := blockchain.NewBlockchain()

	peerChain.AddBlock(nil)

	peerNode := NewNode(
		Config{},
		peerChain,
	)

	peerServer := httptest.NewServer(
		peerNode.Handler(),
	)

	defer peerServer.Close()
	localChain := blockchain.NewBlockchain()

	localNode := NewNode(
		Config{},
		localChain,
	)

	if err := localNode.SyncFromPeer(
		peerServer.URL,
	); err != nil {
		t.Fatalf("expected sync to succeed: %v", err)
	}
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
		t.Fatal("expected synchronized head hashes to match")
	}
}
func TestSyncFromPeerResolvesFork(t *testing.T) {
	localChain := blockchain.NewBlockchain()

	orphanedTx := createNodeSignedTransaction(
		t,
		"Alice",
		"Bob",
		20,
		1,
	)

	if !localChain.AddTransaction(orphanedTx) {
		t.Fatal("expected local transaction to be accepted")
	}

	localLedger := ledger.NewLedger(
		blockchain.CalculateBalances(
			localChain.Blocks,
			localChain.InitialBalances,
		),
	)

	localChain.MinePendingTransactions(localLedger)

	peerChain := blockchain.NewBlockchain()

	peerTx := createNodeSignedTransaction(
		t,
		"Charlie",
		"Bob",
		10,
		1,
	)

	if !peerChain.AddTransaction(peerTx) {
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

	peerServer := httptest.NewServer(
		peerNode.Handler(),
	)

	defer peerServer.Close()

	localNode := NewNode(Config{}, localChain)

	if err := localNode.SyncFromPeer(
		peerServer.URL,
	); err != nil {
		t.Fatalf("expected fork sync to succeed: %v", err)
	}
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
		t.Fatal("expected local node to adopt peer chain")
	}
	if len(localChain.PendingTransactions) != 1 {
		t.Fatalf(
			"expected 1 orphaned transaction in pending pool, got %d",
			len(localChain.PendingTransactions),
		)
	}

	if localChain.PendingTransactions[0].ID != orphanedTx.ID {
		t.Fatal("expected orphaned transaction to return to pending")
	}

}
