package blockchain

import (
	"toy-blockchain/block"
	"toy-blockchain/ledger"
)

func FindForkPoint(
	local []block.Block,
	candidate []block.Block,
) int {
	shorterLength := len(local)

	if len(candidate) < shorterLength {
		shorterLength = len(candidate)
	}

	lastCommonIndex := -1

	for i := 0; i < shorterLength; i++ {
		if local[i].Hash != candidate[i].Hash {
			break
		}

		lastCommonIndex = i
	}

	return lastCommonIndex
}
func ConfirmedTransactionIDs(
	blocks []block.Block,
) map[string]bool {
	confirmed := make(map[string]bool)

	for _, currentBlock := range blocks {
		for _, tx := range currentBlock.Transactions {
			if tx.ID != "" {
				confirmed[tx.ID] = true
			}
		}
	}

	return confirmed
}
func OrphanedTransactions(
	localBlocks []block.Block,
	forkPoint int,
	confirmedIDs map[string]bool,
) []ledger.Transaction {
	orphaned := []ledger.Transaction{}

	for i := forkPoint + 1; i < len(localBlocks); i++ {
		for _, tx := range localBlocks[i].Transactions {
			if tx.ID == "" {
				continue
			}

			if confirmedIDs[tx.ID] {
				continue
			}

			orphaned = append(orphaned, tx)
		}
	}

	return orphaned
}
