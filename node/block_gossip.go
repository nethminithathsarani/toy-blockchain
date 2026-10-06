package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"toy-blockchain/block"
)

func blockURL(peer string) string {
	peer = strings.TrimRight(peer, "/")

	if !strings.HasPrefix(peer, "http://") &&
		!strings.HasPrefix(peer, "https://") {

		peer = "http://" + peer
	}

	return peer + "/blocks"
}

func (n *Node) forwardBlock(
	newBlock block.Block,
	excludedPeer string,
) {
	blockJSON, err := json.Marshal(newBlock)

	if err != nil {
		fmt.Println("Failed to encode block:", err)
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
			blockURL(peer),
			bytes.NewReader(blockJSON),
		)

		if err != nil {
			fmt.Println("Failed to create block request:", err)
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
				"Failed to forward block to",
				peer,
				":",
				err,
			)
			continue
		}

		fmt.Println(
			"Block forwarded:",
			newBlock.Index,
			newBlock.Hash,
			"to",
			peer,
		)

		response.Body.Close()
	}
}
