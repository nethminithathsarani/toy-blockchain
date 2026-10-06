package node

import (
	"encoding/json"
	"fmt"
	"net/http"

	"toy-blockchain/block"
)

type BlockResponse struct {
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
	Message   string `json:"message"`
}

func (n *Node) handleBlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method == http.MethodGet {
		n.handleMissingBlocks(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	sourcePeer := r.Header.Get("X-Node-Address")

	var receivedBlock block.Block

	if err := json.NewDecoder(r.Body).Decode(
		&receivedBlock,
	); err != nil {
		http.Error(
			w,
			"invalid block JSON",
			http.StatusBadRequest,
		)
		return
	}

	n.mu.Lock()

	if n.seenBlocks[receivedBlock.Hash] {
		n.mu.Unlock()

		fmt.Println(
			"Duplicate block ignored:",
			receivedBlock.Hash,
		)

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(BlockResponse{
			Accepted:  false,
			Duplicate: true,
			Message:   "block already known",
		})
		return
	}

	if err := n.Blockchain.AddReceivedBlock(
		receivedBlock,
	); err != nil {
		n.mu.Unlock()

		// If the block came from a peer and does not connect to our
		// current chain, download that peer's chain and resolve the fork.
		if sourcePeer != "" {
			fmt.Println(
				"Competing block received; synchronizing with:",
				sourcePeer,
			)

			if syncErr := n.SyncFromPeer(sourcePeer); syncErr != nil {
				http.Error(
					w,
					"block rejected and synchronization failed: "+
						syncErr.Error(),
					http.StatusConflict,
				)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)

			json.NewEncoder(w).Encode(BlockResponse{
				Accepted:  true,
				Duplicate: false,
				Message:   "competing chain synchronized",
			})
			return
		}

		http.Error(
			w,
			"invalid block: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	n.seenBlocks[receivedBlock.Hash] = true

	for _, tx := range receivedBlock.Transactions {
		if tx.ID != "" {
			n.seenTransactions[tx.ID] = true
		}
	}

	n.mu.Unlock()

	fmt.Println(
		"Block accepted:",
		receivedBlock.Index,
		receivedBlock.Hash,
	)

	n.forwardBlock(receivedBlock, sourcePeer)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(BlockResponse{
		Accepted:  true,
		Duplicate: false,
		Message:   "block accepted",
	})
}
