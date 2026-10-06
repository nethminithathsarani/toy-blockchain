package blockchain

import (
	"math/big"

	"toy-blockchain/block"
)

func CalculateChainWork(
	blocks []block.Block,
) *big.Int {
	totalWork := big.NewInt(0)
	base := big.NewInt(16)

	for _, currentBlock := range blocks {
		blockWork := new(big.Int).Exp(
			base,
			big.NewInt(int64(currentBlock.Difficulty)),
			nil,
		)

		totalWork.Add(totalWork, blockWork)
	}

	return totalWork
}
