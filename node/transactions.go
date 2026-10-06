package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"toy-blockchain/ledger"
)

type TransactionResponse struct {
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
	Message   string `json:"message"`
}

func (n *Node) handleTransaction(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	sourcePeer := r.Header.Get("X-Node-Address")

	var tx ledger.Transaction

	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(
			w,
			"invalid transaction JSON",
			http.StatusBadRequest,
		)
		return
	}

	n.mu.Lock()

	// Already known: unlock, respond and do not forward.
	if n.seenTransactions[tx.ID] {
		n.mu.Unlock()
		fmt.Println(
			"Duplicate transaction ignored:",
			tx.ID,
		)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(TransactionResponse{
			Accepted:  false,
			Duplicate: true,
			Message:   "transaction already known",
		})
		return
	}

	// Invalid: unlock, reject and do not forward.
	if !n.Blockchain.AddTransaction(tx) {
		n.mu.Unlock()

		http.Error(
			w,
			"invalid transaction",
			http.StatusBadRequest,
		)
		return
	}

	// Valid and new: remember it, then unlock.
	n.seenTransactions[tx.ID] = true
	n.mu.Unlock()
	fmt.Println(
		"Transaction accepted:",
		tx.ID,
	)
	// Forward only the newly accepted transaction.
	n.forwardTransaction(tx, sourcePeer)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(TransactionResponse{
		Accepted:  true,
		Duplicate: false,
		Message:   "transaction accepted",
	})
}
