package blockchain

func (bc *Blockchain) NextNonce(senderAddress string) int {
	highestNonce := 0
	for _, currentBlock := range bc.Blocks {
		for _, tx := range currentBlock.Transactions {
			if tx.SenderAddress == senderAddress &&
				tx.Nonce > highestNonce {

				highestNonce = tx.Nonce
			}
		}
	}
	for _, tx := range bc.PendingTransactions {
		if tx.SenderAddress == senderAddress &&
			tx.Nonce > highestNonce {

			highestNonce = tx.Nonce
		}
	}
	return highestNonce + 1
}
