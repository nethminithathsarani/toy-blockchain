package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func peerListURL(peer string) string {
	peer = strings.TrimRight(peer, "/")

	if !strings.HasPrefix(peer, "http://") &&
		!strings.HasPrefix(peer, "https://") {

		peer = "http://" + peer
	}

	return peer + "/peers"
}

func fetchPeerList(peer string) ([]string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	response, err := client.Get(peerListURL(peer))

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"peer returned status %d",
			response.StatusCode,
		)
	}

	var peersResponse PeersResponse

	if err := json.NewDecoder(response.Body).Decode(
		&peersResponse,
	); err != nil {
		return nil, err
	}

	return peersResponse.Peers, nil
}
func registerWithPeer(
	peer string,
	nodeAddress string,
) error {
	registrationJSON, err := json.Marshal(
		PeerRegistration{
			Address: nodeAddress,
		},
	)

	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		peerListURL(peer),
		bytes.NewReader(registrationJSON),
	)

	if err != nil {
		return err
	}

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	response, err := client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK &&
		response.StatusCode != http.StatusAccepted {

		return fmt.Errorf(
			"peer returned status %d",
			response.StatusCode,
		)
	}

	return nil
}

// DiscoverPeers asks known peers for their peer lists.
// Newly discovered peers are added and queried as well.
func (n *Node) DiscoverPeers() error {
	queue := n.PeerSnapshot()
	visited := make(map[string]bool)

	var lastError error
	successfulRequest := false

	for len(queue) > 0 {
		peer := queue[0]
		queue = queue[1:]

		if visited[peer] {
			continue
		}

		visited[peer] = true

		discoveredPeers, err := fetchPeerList(peer)

		if err != nil {
			lastError = err

			fmt.Println(
				"Peer discovery failed for",
				peer,
				":",
				err,
			)

			continue
		}

		successfulRequest = true

		if n.Config.AdvertiseAddress != "" {
			if err := registerWithPeer(
				peer,
				n.Config.AdvertiseAddress,
			); err != nil {
				fmt.Println(
					"Failed to register with peer",
					peer,
					":",
					err,
				)
			}
		}

		for _, discoveredPeer := range discoveredPeers {
			if n.AddPeer(discoveredPeer) {
				queue = append(
					queue,
					discoveredPeer,
				)

				fmt.Println(
					"Peer discovered:",
					discoveredPeer,
				)
			}
		}
	}

	if !successfulRequest && lastError != nil {
		return fmt.Errorf(
			"peer discovery failed: %w",
			lastError,
		)
	}

	return nil
}
