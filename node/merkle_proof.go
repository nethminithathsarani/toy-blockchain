package node

import (
	"encoding/json"
	"net/http"
	"strconv"

	"toy-blockchain/block"
	"toy-blockchain/ledger"
)

type MerkleProofResponse struct {
	BlockIndex       int                     `json:"block_index"`
	TransactionIndex int                     `json:"transaction_index"`
	Transaction      ledger.Transaction      `json:"transaction"`
	MerkleRoot       string                  `json:"merkle_root"`
	Proof            []block.MerkleProofStep `json:"proof"`
	Verified         bool                    `json:"verified"`
}

func (n *Node) handleMerkleProof(
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

	blockText := r.URL.Query().Get("block")
	transactionID := r.URL.Query().Get("transaction")

	if blockText == "" || transactionID == "" {
		http.Error(
			w,
			"block and transaction are required",
			http.StatusBadRequest,
		)
		return
	}

	blockIndex, err := strconv.Atoi(blockText)

	if err != nil || blockIndex < 0 {
		http.Error(
			w,
			"invalid block index",
			http.StatusBadRequest,
		)
		return
	}

	n.mu.RLock()

	if blockIndex >= len(n.Blockchain.Blocks) {
		n.mu.RUnlock()

		http.Error(
			w,
			"block not found",
			http.StatusNotFound,
		)
		return
	}

	selectedBlock := n.Blockchain.Blocks[blockIndex]

	transactionIndex := -1

	for index, transaction := range selectedBlock.Transactions {
		if transaction.ID == transactionID {
			transactionIndex = index
			break
		}
	}

	if transactionIndex == -1 {
		n.mu.RUnlock()

		http.Error(
			w,
			"transaction not found in block",
			http.StatusNotFound,
		)
		return
	}

	proof, err := block.GenerateMerkleProof(
		selectedBlock.Transactions,
		transactionIndex,
	)

	if err != nil {
		n.mu.RUnlock()

		http.Error(
			w,
			"failed to generate Merkle proof",
			http.StatusInternalServerError,
		)
		return
	}

	transaction := selectedBlock.Transactions[transactionIndex]

	response := MerkleProofResponse{
		BlockIndex:       blockIndex,
		TransactionIndex: transactionIndex,
		Transaction:      transaction,
		MerkleRoot:       selectedBlock.MerkleRoot,
		Proof:            proof,
		Verified: block.VerifyMerkleProof(
			transaction,
			proof,
			selectedBlock.MerkleRoot,
		),
	}

	n.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
