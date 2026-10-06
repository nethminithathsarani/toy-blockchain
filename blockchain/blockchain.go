package blockchain

import (
	"fmt"
	"time"
	"toy-blockchain/block"
	"toy-blockchain/ledger"
	"toy-blockchain/wallet"
)

type Blockchain struct {
	Blocks              []block.Block
	InitialBalances     map[string]int
	PendingTransactions []ledger.Transaction
	Difficulty          int
}

const (
	TargetBlockTime    int64 = 10
	AdjustmentInterval       = 5
)

func CreateGenesisBlock() block.Block {
	genesis := block.Block{
		Index:        0,
		Timestamp:    0,
		Transactions: []ledger.Transaction{},
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
		Nonce:        0,
		MerkleRoot:   "",
		Difficulty:   4,
	}
	// Calculate and assign the hash.
	genesis.Hash = block.CalculateHash(genesis)
	return genesis
}
func (bc *Blockchain) AddBlock(transactions []ledger.Transaction) {

	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	newBlock := block.Block{
		Index:        lastBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: lastBlock.Hash,
		Nonce:        0,
		MerkleRoot:   block.CalculateMerkleRoot(transactions),
		Difficulty:   bc.Difficulty,
	}

	attempts, duration := MineBlockConcurrent(
		&newBlock,
		bc.Difficulty,
		4,
	)

	fmt.Println("Mining attempts:", attempts)
	fmt.Println("Mining time:", duration)

	bc.Blocks = append(bc.Blocks, newBlock)
	bc.AdjustDifficulty()
}

// NewBlockchain creates a blockchain with the Genesis Block.
func NewBlockchain() *Blockchain {
	genesis := CreateGenesisBlock()

	return &Blockchain{
		Blocks: []block.Block{genesis},
		InitialBalances: map[string]int{
			"Alice":   100,
			"Bob":     100,
			"Charlie": 100,
		},
		PendingTransactions: []ledger.Transaction{},
		Difficulty:          4,
	}
}
func (bc *Blockchain) MinePendingTransactions(l *ledger.Ledger) {

	if len(bc.PendingTransactions) == 0 {
		return
	}

	tempBalances := make(map[string]int)

	for name, balance := range l.Balances {
		tempBalances[name] = balance
	}

	tempLedger := ledger.NewLedger(tempBalances)

	validTransactions := []ledger.Transaction{}

	for _, tx := range bc.PendingTransactions {
		if wallet.VerifyTransaction(tx) &&
			tempLedger.ValidateTransaction(tx) {

			validTransactions = append(validTransactions, tx)
			tempLedger.ApplyTransaction(tx)
		}
	}

	if len(validTransactions) == 0 {
		bc.PendingTransactions = []ledger.Transaction{}
		return
	}

	bc.AddBlock(validTransactions)

	for _, tx := range validTransactions {
		l.ApplyTransaction(tx)
	}

	bc.PendingTransactions = []ledger.Transaction{}
}
func (bc *Blockchain) AddTransaction(t ledger.Transaction) bool {
	if !wallet.VerifyTransaction(t) {
		return false
	}
	expectedNonce := bc.NextNonce(t.SenderAddress)

	if t.Nonce != expectedNonce {
		return false
	}
	balances := CalculateBalances(bc.Blocks, bc.InitialBalances)
	tempLedger := ledger.NewLedger(balances)

	for _, pending := range bc.PendingTransactions {
		if tempLedger.ValidateTransaction(pending) {
			tempLedger.ApplyTransaction(pending)
		}
	}

	if !tempLedger.ValidateTransaction(t) {
		return false
	}

	bc.PendingTransactions = append(bc.PendingTransactions, t)
	return true
}
func ValidateTransactionSignature(tx ledger.Transaction) bool {

	return wallet.VerifyTransaction(tx)

}
