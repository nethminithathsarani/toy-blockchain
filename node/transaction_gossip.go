package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"toy-blockchain/ledger"
)

func transactionURL(peer string) string {
	peer = strings.TrimRight(peer, "/")

	if !strings.HasPrefix(peer, "http://") &&
		!strings.HasPrefix(peer, "https://") {

		peer = "http://" + peer
	}

	return peer + "/transactions"
}
func (n *Node) forwardTransaction(
	tx ledger.Transaction,
	excludedPeer string,
) {
	transactionJSON, err := json.Marshal(tx)

	if err != nil {
		fmt.Println("Failed to encode transaction:", err)
		return
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	peers := n.PeerSnapshot()

	for _, peer := range peers {
		if peer == excludedPeer {
			continue
		}

		request, err := http.NewRequest(
			http.MethodPost,
			transactionURL(peer),
			bytes.NewReader(transactionJSON),
		)

		if err != nil {
			fmt.Println("Failed to create peer request:", err)
			continue
		}
		request.Header.Set(
			"X-Node-Address",
			n.Config.AdvertiseAddress,
		)

		request.Header.Set(
			"Content-Type",
			"application/json",
		)
		response, err := client.Do(request)

		if err != nil {
			fmt.Println(
				"Failed to forward transaction to",
				peer,
				":",
				err,
			)
			continue
		}
		fmt.Println(
			"Transaction forwarded:",
			tx.ID,
			"to",
			peer,
		)
		response.Body.Close()
	}
}
