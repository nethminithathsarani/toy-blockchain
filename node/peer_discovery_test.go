package node

import (
	"net/http/httptest"
	"testing"

	"toy-blockchain/blockchain"
)

func TestDiscoverPeersFromSingleSeed(t *testing.T) {
	nodeB := NewNode(
		Config{
			Address: "node-b",
		},
		blockchain.NewBlockchain(),
	)

	serverB := httptest.NewServer(nodeB.Handler())
	defer serverB.Close()

	nodeA := NewNode(
		Config{
			Address: "node-a",
			Peers: []string{
				serverB.URL,
			},
		},
		blockchain.NewBlockchain(),
	)

	serverA := httptest.NewServer(nodeA.Handler())
	defer serverA.Close()

	nodeC := NewNode(
		Config{
			Address: "node-c",
			Peers: []string{
				serverA.URL,
			},
		},
		blockchain.NewBlockchain(),
	)

	if err := nodeC.DiscoverPeers(); err != nil {
		t.Fatalf(
			"expected peer discovery to succeed: %v",
			err,
		)
	}

	// Node C should retain Node A and discover Node B.
	peers := nodeC.PeerSnapshot()

	if len(peers) != 2 {
		t.Fatalf(
			"expected Node C to know 2 peers, got %d",
			len(peers),
		)
	}

	knownPeers := make(map[string]bool)

	for _, peer := range peers {
		knownPeers[peer] = true
	}

	if !knownPeers[serverA.URL] {
		t.Fatal("expected Node C to retain seed Node A")
	}

	if !knownPeers[serverB.URL] {
		t.Fatal(
			"expected Node C to discover Node B through Node A",
		)
	}

	// Node C should announce itself to Node A.
	nodeAKnowsC := false

	for _, peer := range nodeA.PeerSnapshot() {
		if peer == "node-c" {
			nodeAKnowsC = true
			break
		}
	}

	if !nodeAKnowsC {
		t.Fatal(
			"expected Node A to learn Node C through registration",
		)
	}

	// Node C should also announce itself to discovered Node B.
	nodeBKnowsC := false

	for _, peer := range nodeB.PeerSnapshot() {
		if peer == "node-c" {
			nodeBKnowsC = true
			break
		}
	}

	if !nodeBKnowsC {
		t.Fatal(
			"expected Node B to learn Node C through registration",
		)
	}
}

func TestAddPeerRejectsSelfAndDuplicates(t *testing.T) {
	n := NewNode(
		Config{
			Address: "localhost:8001",
			Peers: []string{
				"localhost:8002",
			},
		},
		blockchain.NewBlockchain(),
	)

	if n.AddPeer("localhost:8001") {
		t.Fatal("expected node to reject its own address")
	}

	if n.AddPeer("localhost:8002") {
		t.Fatal("expected duplicate peer to be rejected")
	}

	if n.AddPeer("localhost:8002/") {
		t.Fatal(
			"expected peer with trailing slash to be treated as duplicate",
		)
	}

	if !n.AddPeer("localhost:8003") {
		t.Fatal("expected new peer to be added")
	}

	peers := n.PeerSnapshot()

	if len(peers) != 2 {
		t.Fatalf(
			"expected 2 unique peers, got %d",
			len(peers),
		)
	}
}
