package blockchain

import (
	"math/big"
	"testing"

	"toy-blockchain/block"
)

func TestCalculateChainWork(t *testing.T) {
	blocks := []block.Block{
		{
			Difficulty: 1,
		},
		{
			Difficulty: 2,
		},
	}

	work := CalculateChainWork(blocks)

	expected := big.NewInt(272)

	if work.Cmp(expected) != 0 {
		t.Fatalf(
			"expected chain work 272, got %s",
			work.String(),
		)
	}
}
