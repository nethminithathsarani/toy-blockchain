package node

import (
	"encoding/json"
	"net/http"
)

type MempoolResponse struct {
	Size int `json:"size"`
}

func (n *Node) handleMempool(
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
	size := len(n.Blockchain.PendingTransactions)
	n.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(MempoolResponse{
		Size: size,
	})
}
