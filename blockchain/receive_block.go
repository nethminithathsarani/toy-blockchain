package blockchain

import "toy-blockchain/block"

func (bc *Blockchain) AddReceivedBlock(
	receivedBlock block.Block,
) error {
	candidateBlocks := make(
		[]block.Block,
		len(bc.Blocks),
		len(bc.Blocks)+1,
	)

	copy(candidateBlocks, bc.Blocks)

	candidateBlocks = append(
		candidateBlocks,
		receivedBlock,
	)
	candidateChain := &Blockchain{
		Blocks:              candidateBlocks,
		InitialBalances:     bc.InitialBalances,
		PendingTransactions: bc.PendingTransactions,
		Difficulty:          bc.Difficulty,
	}

	if err := candidateChain.ValidateChain(); err != nil {
		return err
	}
	bc.Blocks = candidateBlocks
	includedTransactions := make(map[string]bool)

	for _, tx := range receivedBlock.Transactions {
		includedTransactions[tx.ID] = true
	}

	remainingTransactions := bc.PendingTransactions[:0]

	for _, tx := range bc.PendingTransactions {
		if !includedTransactions[tx.ID] {
			remainingTransactions = append(
				remainingTransactions,
				tx,
			)
		}
	}

	bc.PendingTransactions = remainingTransactions
	bc.AdjustDifficulty()

	return nil
}
