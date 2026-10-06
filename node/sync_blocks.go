package node

import (
	"encoding/json"
	"net/http"
	"strconv"

	"toy-blockchain/block"
)

type MissingBlocksResponse struct {
	Blocks []block.Block `json:"blocks"`
}

func (n *Node) handleMissingBlocks(
	w http.ResponseWriter,
	r *http.Request,
) {
	fromText := r.URL.Query().Get("from")

	from, err := strconv.Atoi(fromText)

	if err != nil || from < 0 {
		http.Error(
			w,
			"invalid starting block index",
			http.StatusBadRequest,
		)
		return
	}
	n.mu.RLock()

	if from > len(n.Blockchain.Blocks) {
		n.mu.RUnlock()

		http.Error(
			w,
			"starting index is beyond chain",
			http.StatusBadRequest,
		)
		return
	}

	blocks := append(
		[]block.Block(nil),
		n.Blockchain.Blocks[from:]...,
	)

	n.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(MissingBlocksResponse{
		Blocks: blocks,
	})
}
