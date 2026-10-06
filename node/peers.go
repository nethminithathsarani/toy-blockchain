package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type PeersResponse struct {
	Peers []string `json:"peers"`
}
type PeerRegistration struct {
	Address string `json:"address"`
}

type PeerRegistrationResponse struct {
	Added   bool   `json:"added"`
	Message string `json:"message"`
}

// PeerSnapshot returns a safe copy of the current peer list.
func (n *Node) PeerSnapshot() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return append(
		[]string(nil),
		n.Config.Peers...,
	)
}

// AddPeer safely adds a new peer.
// It rejects empty addresses, this node's own address and duplicates.
func (n *Node) AddPeer(peer string) bool {
	peer = strings.TrimSpace(peer)
	peer = strings.TrimRight(peer, "/")

	if peer == "" {
		return false
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	nodeAddress := strings.TrimRight(
		n.Config.AdvertiseAddress,
		"/",
	)

	if peer == nodeAddress {
		return false
	}

	for _, existingPeer := range n.Config.Peers {
		if strings.TrimRight(existingPeer, "/") == peer {
			return false
		}
	}

	n.Config.Peers = append(
		n.Config.Peers,
		peer,
	)

	return true
}

func (n *Node) handlePeers(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {
	case http.MethodGet:
		peers := n.PeerSnapshot()

		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		json.NewEncoder(w).Encode(PeersResponse{
			Peers: peers,
		})

	case http.MethodPost:
		var registration PeerRegistration

		if err := json.NewDecoder(r.Body).Decode(
			&registration,
		); err != nil {
			http.Error(
				w,
				"invalid peer registration JSON",
				http.StatusBadRequest,
			)
			return
		}

		if registration.Address == "" {
			http.Error(
				w,
				"peer address is required",
				http.StatusBadRequest,
			)
			return
		}

		added := n.AddPeer(registration.Address)

		message := "peer already known"

		if added {
			message = "peer added"

			fmt.Println(
				"Peer registered:",
				registration.Address,
			)
		}

		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		if added {
			w.WriteHeader(http.StatusAccepted)
		}

		json.NewEncoder(w).Encode(
			PeerRegistrationResponse{
				Added:   added,
				Message: message,
			},
		)

	default:
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}
