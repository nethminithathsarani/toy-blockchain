package node

import "fmt"

// SyncFromPeer synchronizes this node with one peer.
func (n *Node) SyncFromPeer(peer string) error {
	status, err := fetchPeerStatus(peer)

	if err != nil {
		return fmt.Errorf(
			"failed to fetch peer status: %w",
			err,
		)
	}

	fmt.Println(
		"Synchronization started with:",
		peer,
	)

	// Safely read the local chain status.
	n.mu.RLock()
	localHeight := len(n.Blockchain.Blocks) - 1
	localHeadHash := n.Blockchain.Blocks[localHeight].Hash
	n.mu.RUnlock()

	// Both nodes already have the same chain head.
	if status.Height == localHeight &&
		status.HeadHash == localHeadHash {

		fmt.Println(
			"Already synchronized with:",
			peer,
		)

		return nil
	}

	// The peer is not ahead, but its head is different.
	// This may be a competing fork.
	if status.Height <= localHeight {
		return n.resolveForkFromPeer(peer)
	}

	// The peer is ahead, so first try downloading only
	// the blocks missing from the local chain.
	missingBlocks, err := fetchMissingBlocks(
		peer,
		localHeight+1,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to fetch missing blocks: %w",
			err,
		)
	}

	fmt.Println(
		"Missing blocks downloaded:",
		len(missingBlocks),
		"from",
		peer,
	)

	// Validate and append the missing blocks in order.
	for _, missingBlock := range missingBlocks {
		n.mu.Lock()

		err := n.Blockchain.AddReceivedBlock(
			missingBlock,
		)

		if err != nil {
			n.mu.Unlock()

			// The peer may be on a different fork.
			// Download its complete chain and attempt reorganization.
			return n.resolveForkFromPeer(peer)
		}

		n.seenBlocks[missingBlock.Hash] = true

		for _, tx := range missingBlock.Transactions {
			if tx.ID != "" {
				n.seenTransactions[tx.ID] = true
			}
		}

		n.mu.Unlock()
	}

	// Confirm that the local node reached the peer's height.
	n.mu.RLock()
	finalHeight := len(n.Blockchain.Blocks) - 1
	n.mu.RUnlock()

	if finalHeight != status.Height {
		return fmt.Errorf(
			"sync incomplete: local height %d, peer height %d",
			finalHeight,
			status.Height,
		)
	}

	fmt.Println(
		"Synchronization completed with:",
		peer,
		"at height",
		finalHeight,
	)

	return nil
}

// resolveForkFromPeer downloads the peer's complete chain
// and asks the blockchain package to resolve the fork.
func (n *Node) resolveForkFromPeer(
	peer string,
) error {
	fmt.Println(
		"Competing chain detected from:",
		peer,
	)

	candidate, err := fetchPeerChain(peer)

	if err != nil {
		return fmt.Errorf(
			"failed to fetch peer chain: %w",
			err,
		)
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.Blockchain.ResolveFork(candidate); err != nil {
		return fmt.Errorf(
			"failed to resolve fork: %w",
			err,
		)
	}

	n.rebuildSeenState()

	fmt.Println(
		"Stronger chain adopted from:",
		peer,
	)

	return nil
}

// rebuildSeenState recreates the de-duplication maps
// after the node switches to a different chain.
func (n *Node) rebuildSeenState() {
	n.seenBlocks = make(map[string]bool)
	n.seenTransactions = make(map[string]bool)

	for _, currentBlock := range n.Blockchain.Blocks {
		if currentBlock.Hash != "" {
			n.seenBlocks[currentBlock.Hash] = true
		}

		for _, tx := range currentBlock.Transactions {
			if tx.ID != "" {
				n.seenTransactions[tx.ID] = true
			}
		}
	}

	for _, tx := range n.Blockchain.PendingTransactions {
		if tx.ID != "" {
			n.seenTransactions[tx.ID] = true
		}
	}
}

func (n *Node) SyncWithPeers() error {
	var lastError error

	peers := n.PeerSnapshot()

	for _, peer := range peers {
		err := n.SyncFromPeer(peer)

		if err == nil {
			return nil
		}

		lastError = err
	}

	return lastError
}
