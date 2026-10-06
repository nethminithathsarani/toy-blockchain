package node

import "net/http"

func (n *Node) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", n.handleStatus)
	mux.HandleFunc("/peers", n.handlePeers)
	mux.HandleFunc("/mempool", n.handleMempool)
	mux.HandleFunc("/balances", n.handleBalances)
	mux.HandleFunc("/chain", n.handleChain)
	mux.HandleFunc("/transactions", n.handleTransaction)
	mux.HandleFunc("/blocks", n.handleBlock)
	mux.HandleFunc("/mine", n.handleMine)
	mux.HandleFunc("/nonce", n.handleNonce)
	mux.HandleFunc("/merkle-proof", n.handleMerkleProof)
	return mux
}
func (n *Node) Start() error {
	return http.ListenAndServe(
		n.Config.Address,
		n.Handler(),
	)
}
