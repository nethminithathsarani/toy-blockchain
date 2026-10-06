package node

import (
	"encoding/json"
	"net/http"

	"toy-blockchain/blockchain"
)

type BalancesResponse struct {
	Balances map[string]int `json:"balances"`
}

func (n *Node) handleBalances(
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

	balances := blockchain.CalculateBalances(
		n.Blockchain.Blocks,
		n.Blockchain.InitialBalances,
	)

	n.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(BalancesResponse{
		Balances: balances,
	})
}
