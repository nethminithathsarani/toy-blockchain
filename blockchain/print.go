package blockchain

import "fmt"

func (bc *Blockchain) PrintChain() {

	for _, block := range bc.Blocks {
		fmt.Println("----------------------------")
		fmt.Println("Index:", block.Index)
		fmt.Println("Transactions:", block.Transactions)
		fmt.Println("Nonce:", block.Nonce)
		fmt.Println("Previous Hash:", block.PreviousHash)
		fmt.Println("Hash:", block.Hash)
		fmt.Println("Merkle Root:", block.MerkleRoot)
		fmt.Println("Difficulty:", block.Difficulty)

	}
}
