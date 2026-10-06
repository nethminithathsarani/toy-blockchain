package node

import (
	"encoding/json"
	"fmt"
	"net/http"

	"toy-blockchain/block"
	"toy-blockchain/blockchain"
	"toy-blockchain/ledger"
)

type MineResponse struct {
	Block block.Block `json:"block"`
}

func (n *Node) handleMine(
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

	n.mu.Lock()

	if len(n.Blockchain.PendingTransactions) == 0 {
		n.mu.Unlock()

		http.Error(
			w,
			"no pending transactions",
			http.StatusBadRequest,
		)
		return
	}

	previousBlockCount := len(n.Blockchain.Blocks)

	balances := blockchain.CalculateBalances(
		n.Blockchain.Blocks,
		n.Blockchain.InitialBalances,
	)

	currentLedger := ledger.NewLedger(balances)

	n.Blockchain.MinePendingTransactions(currentLedger)

	if len(n.Blockchain.Blocks) == previousBlockCount {
		n.mu.Unlock()

		http.Error(
			w,
			"no valid transactions to mine",
			http.StatusBadRequest,
		)
		return
	}

	newBlock := n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1]

	n.seenBlocks[newBlock.Hash] = true

	for _, tx := range newBlock.Transactions {
		if tx.ID != "" {
			n.seenTransactions[tx.ID] = true
		}
	}

	n.mu.Unlock()

	fmt.Println(
		"Block mined:",
		newBlock.Index,
		newBlock.Hash,
	)

	n.forwardBlock(newBlock, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(MineResponse{
		Block: newBlock,
	})
}
