package node

import (
	"encoding/json"
	"net/http"

	"toy-blockchain/block"
)

type ChainResponse struct {
	Blocks []block.Block `json:"blocks"`
}

func (n *Node) handleChain(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}
	n.mu.RLock()
	defer n.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(ChainResponse{
		Blocks: n.Blockchain.Blocks,
	})
}
