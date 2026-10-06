package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"toy-blockchain/block"
)

func statusURL(peer string) string {
	peer = strings.TrimRight(peer, "/")

	if !strings.HasPrefix(peer, "http://") &&
		!strings.HasPrefix(peer, "https://") {

		peer = "http://" + peer
	}

	return peer + "/status"
}
func missingBlocksURL(
	peer string,
	from int,
) string {
	peer = strings.TrimRight(peer, "/")

	if !strings.HasPrefix(peer, "http://") &&
		!strings.HasPrefix(peer, "https://") {

		peer = "http://" + peer
	}

	fromText := url.QueryEscape(strconv.Itoa(from))

	return peer + "/blocks?from=" + fromText
}
func chainURL(peer string) string {
	peer = strings.TrimRight(peer, "/")

	if !strings.HasPrefix(peer, "http://") &&
		!strings.HasPrefix(peer, "https://") {

		peer = "http://" + peer
	}

	return peer + "/chain"
}
func fetchPeerStatus(
	peer string,
) (StatusResponse, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	response, err := client.Get(statusURL(peer))

	if err != nil {
		return StatusResponse{}, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return StatusResponse{}, fmt.Errorf(
			"peer returned status %d",
			response.StatusCode,
		)
	}

	var status StatusResponse

	if err := json.NewDecoder(response.Body).Decode(
		&status,
	); err != nil {
		return StatusResponse{}, err
	}

	return status, nil
}
func fetchMissingBlocks(
	peer string,
	from int,
) ([]block.Block, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Get(
		missingBlocksURL(peer, from),
	)

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

	var missingResponse MissingBlocksResponse

	if err := json.NewDecoder(response.Body).Decode(
		&missingResponse,
	); err != nil {
		return nil, err
	}

	return missingResponse.Blocks, nil
}
func fetchPeerChain(
	peer string,
) ([]block.Block, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Get(chainURL(peer))

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

	var chainResponse ChainResponse

	if err := json.NewDecoder(response.Body).Decode(
		&chainResponse,
	); err != nil {
		return nil, err
	}

	return chainResponse.Blocks, nil
}
