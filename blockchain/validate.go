package blockchain

import (
	"fmt"
	"strings"

	"toy-blockchain/block"
	"toy-blockchain/wallet"
)

func (bc *Blockchain) ValidateChain() error {
	// Check blockchain is not empty
	if len(bc.Blocks) == 0 {
		return fmt.Errorf("blockchain is empty")
	}

	// Check genesis block
	genesis := bc.Blocks[0]

	// Genesis block must have fixed properties
	if genesis.Index != 0 {
		return fmt.Errorf("invalid genesis index")
	}

	expectedPreviousHash := "0000000000000000000000000000000000000000000000000000000000000000"

	if genesis.PreviousHash != expectedPreviousHash {
		return fmt.Errorf("invalid genesis previous hash")
	}

	// Validate genesis Merkle root
	calculatedRoot := block.CalculateMerkleRoot(genesis.Transactions)
	if calculatedRoot != genesis.MerkleRoot {
		return fmt.Errorf("block 0: invalid Merkle root")
	}

	// Validate genesis hash
	if block.CalculateHash(genesis) != genesis.Hash {
		return fmt.Errorf("block 0: invalid genesis hash")
	}

	// Tracks the difficulty every block SHOULD have been mined at, replaying
	// the same retargeting rule AddBlock uses in production (see
	// AdjustDifficulty). Without this, ValidateChain only checked that a
	// block's hash matched its OWN recorded Difficulty field -- which means
	// a rebuilt chain could simply declare Difficulty: 0 on every block
	// (trivial to "mine") and still pass validation, defeating the point of
	// proof-of-work for anything loaded from a file or offered via
	// ResolveFork.
	expectedDifficulty := genesis.Difficulty
	lastNonceBySender := make(map[string]int)

	// Check remaining blocks
	for i := 1; i < len(bc.Blocks); i++ {

		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]
		// Verify every transaction signature in the current block.
		for txIndex, tx := range current.Transactions {
			if !wallet.VerifyTransaction(tx) {
				return fmt.Errorf(
					"block %d transaction %d: invalid signature",
					i,
					txIndex,
				)
			}
			expectedNonce := lastNonceBySender[tx.SenderAddress] + 1

			if tx.Nonce != expectedNonce {
				return fmt.Errorf(
					"block %d transaction %d: invalid nonce %d, expected %d",
					i,
					txIndex,
					tx.Nonce,
					expectedNonce,
				)
			}

			lastNonceBySender[tx.SenderAddress] = tx.Nonce
		}

		// Check Merkle root
		calculatedRoot := block.CalculateMerkleRoot(current.Transactions)
		if calculatedRoot != current.MerkleRoot {
			return fmt.Errorf("block %d: invalid Merkle root", i)
		}

		// Check block hash
		calculatedHash := block.CalculateHash(current)
		if calculatedHash != current.Hash {
			return fmt.Errorf("block %d: invalid hash", i)
		}

		// Check previous hash link
		if current.PreviousHash != previous.Hash {
			return fmt.Errorf("block %d: invalid previous hash link", i)
		}

		// Check block order
		if current.Index != previous.Index+1 {
			return fmt.Errorf("block %d: invalid index", i)
		}

		// Check timestamp order
		if current.Timestamp < previous.Timestamp {
			return fmt.Errorf("block %d: invalid timestamp order", i)
		}

		// Check that this block was mined at the difficulty the retargeting
		// rule actually required at this point in the chain -- not just
		// whatever difficulty the block happens to claim for itself.
		if current.Difficulty != expectedDifficulty {
			return fmt.Errorf(
				"block %d: unexpected difficulty %d (expected %d)",
				i, current.Difficulty, expectedDifficulty,
			)
		}

		// Check proof of work
		prefix := strings.Repeat("0", current.Difficulty)
		if !strings.HasPrefix(current.Hash, prefix) {
			return fmt.Errorf("block %d: invalid proof of work", i)
		}

		// Advance the expected difficulty exactly as AddBlock would after
		// mining this block, using the real chain seen so far.
		tracker := Blockchain{
			Blocks:     bc.Blocks[:i+1],
			Difficulty: expectedDifficulty,
		}
		tracker.AdjustDifficulty()
		expectedDifficulty = tracker.Difficulty
	}

	// Check transaction balances
	if !bc.ValidateBalances() {
		return fmt.Errorf("invalid balances: negative balance or invalid transaction")
	}

	return nil
}
