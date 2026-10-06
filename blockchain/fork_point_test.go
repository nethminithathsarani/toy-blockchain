package blockchain

import (
	"testing"

	"toy-blockchain/block"
)

func TestFindForkPoint(t *testing.T) {
	genesis := block.Block{
		Index: 0,
		Hash:  "genesis",
	}

	sharedBlock := block.Block{
		Index: 1,
		Hash:  "shared-block",
	}

	localBlock := block.Block{
		Index: 2,
		Hash:  "local-block",
	}

	candidateBlock := block.Block{
		Index: 2,
		Hash:  "candidate-block",
	}
	local := []block.Block{
		genesis,
		sharedBlock,
		localBlock,
	}

	candidate := []block.Block{
		genesis,
		sharedBlock,
		candidateBlock,
	}
	forkPoint := FindForkPoint(local, candidate)

	if forkPoint != 1 {
		t.Fatalf(
			"expected fork point 1, got %d",
			forkPoint,
		)
	}
}
