package node

import (
	"encoding/json"
	"net/http"
)

type StatusResponse struct {
	Height   int    `json:"height"`
	HeadHash string `json:"head_hash"`
}

func (n *Node) handleStatus(
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

	height := len(n.Blockchain.Blocks) - 1
	headHash := ""

	if len(n.Blockchain.Blocks) > 0 {
		headHash = n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1].Hash
	}

	n.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(StatusResponse{
		Height:   height,
		HeadHash: headHash,
	})
}
