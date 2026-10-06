package node

import (
	"encoding/json"
	"net/http"
)

type NonceResponse struct {
	Nonce int `json:"nonce"`
}

func (n *Node) handleNonce(
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

	address := r.URL.Query().Get("address")

	if address == "" {
		http.Error(
			w,
			"sender address is required",
			http.StatusBadRequest,
		)
		return
	}

	n.mu.RLock()
	nonce := n.Blockchain.NextNonce(address)
	n.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(NonceResponse{
		Nonce: nonce,
	})
}
