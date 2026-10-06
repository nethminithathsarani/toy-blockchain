package main

import (
	"fmt"
	"os"
	"strconv"

	"toy-blockchain/blockchain"
	"toy-blockchain/ledger"
	"toy-blockchain/node"
	"toy-blockchain/wallet"
)

func main() {

	bc, err := blockchain.LoadFromFile("chain.json")

	if err != nil {

		if _, statErr := os.Stat("chain.json"); statErr == nil {
			fmt.Println("Blockchain loading failed:", err)
			return
		}

		bc = blockchain.NewBlockchain()
	}

	if len(os.Args) < 2 {
		fmt.Println("Please provide a command")
		return
	}

	command := os.Args[1]

	switch command {

	case "add":

		if len(os.Args) < 5 {
			fmt.Println("Usage: add <sender> <receiver> <amount>")
			return
		}

		amount, err := strconv.Atoi(os.Args[4])

		if err != nil {
			fmt.Println("Invalid amount")
			return
		}

		tx := ledger.Transaction{
			Sender:   os.Args[2],
			Receiver: os.Args[3],
			Amount:   amount,
		}

		senderWallet, exists := wallet.GetWallet(tx.Sender)

		if !exists {
			fmt.Println("Sender wallet not found")
			return
		}
		tx.Nonce = bc.NextNonce(senderWallet.GetAddress())

		signedTx, err := wallet.SignTransaction(
			tx,
			senderWallet,
		)

		if err != nil {
			fmt.Println("Signing failed:", err)
			return
		}

		tx = signedTx

		if !bc.AddTransaction(tx) {
			fmt.Println("Transaction rejected")
			return
		}

		err = bc.SaveToFile("chain.json")

		if err != nil {
			fmt.Println("Error saving blockchain:", err)
			return
		}

		fmt.Println("Transaction added")
	case "submit":
		if len(os.Args) < 6 {
			fmt.Println(
				"Usage: submit <node-address> <sender> <receiver> <amount>",
			)
			return
		}

		nodeAddress := os.Args[2]
		sender := os.Args[3]
		receiver := os.Args[4]

		amount, err := strconv.Atoi(os.Args[5])

		if err != nil {
			fmt.Println("Invalid amount")
			return
		}

		senderWallet, exists := wallet.GetWallet(sender)

		if !exists {
			fmt.Println("Sender wallet not found")
			return
		}

		nonce, err := node.FetchNextNonce(
			nodeAddress,
			senderWallet.GetAddress(),
		)

		if err != nil {
			fmt.Println("Failed to fetch nonce:", err)
			return
		}

		tx := ledger.Transaction{
			Sender:   sender,
			Receiver: receiver,
			Amount:   amount,
			Nonce:    nonce,
		}

		signedTx, err := wallet.SignTransaction(
			tx,
			senderWallet,
		)

		if err != nil {
			fmt.Println("Signing failed:", err)
			return
		}

		if err := node.SubmitTransaction(
			nodeAddress,
			signedTx,
		); err != nil {
			fmt.Println("Submission failed:", err)
			return
		}

		fmt.Println(
			"Transaction submitted to",
			nodeAddress,
		)
	case "node":
		config, err := node.ParseConfig(os.Args[2:])

		if err != nil {
			fmt.Println("Invalid node configuration:", err)
			return
		}
		blockchainNode := node.NewNode(config, bc)

		if err := blockchainNode.DiscoverPeers(); err != nil {
			fmt.Println(
				"Initial peer discovery failed:",
				err,
			)
		}

		if err := blockchainNode.SyncWithPeers(); err != nil {
			fmt.Println(
				"Initial synchronization failed:",
				err,
			)
		}

		fmt.Println("Node listening on", config.Address)
		fmt.Println(
			"Known peers:",
			blockchainNode.PeerSnapshot(),
		)
		if err := blockchainNode.Start(); err != nil {
			fmt.Println("Node server failed:", err)
		}

	case "mine":

		currentBalances := blockchain.CalculateBalances(
			bc.Blocks,
			bc.InitialBalances,
		)

		l := ledger.NewLedger(currentBalances)

		bc.MinePendingTransactions(l)

		err = bc.SaveToFile("chain.json")

		if err != nil {
			fmt.Println("Save failed:", err)
			return
		}

		fmt.Println("Mining completed")

	case "print":

		bc.PrintChain()

	case "validate":

		err := bc.ValidateChain()

		if err != nil {
			fmt.Println("Blockchain is invalid:", err)
		} else {
			fmt.Println("Blockchain is valid")
		}

	case "balance":
		balances := blockchain.CalculateBalances(bc.Blocks, bc.InitialBalances)
		fmt.Println("Balances:", balances)
	case "wallet":

		if len(os.Args) < 3 {
			fmt.Println("Usage: wallet <name>")
			return
		}

		err := wallet.CreateWallet(os.Args[2])

		if err != nil {
			fmt.Println("Wallet creation failed")
			return
		}

		fmt.Println("Wallet created for", os.Args[2])

	default:

		fmt.Println("Commands:")
		fmt.Println("  add <sender> <receiver> <amount>")
		fmt.Println("  mine")
		fmt.Println("  print")
		fmt.Println("  validate")
		fmt.Println("  balance")
		fmt.Println(
			"  node -address <listen-address> " +
				"-advertise <peer-address> " +
				"-peers <peer1,peer2>",
		)
		fmt.Println("  submit <node-address> <sender> <receiver> <amount>")
	}
}
