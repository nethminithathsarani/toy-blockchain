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

func normalizeNodeURL(address string) string {
	address = strings.TrimRight(address, "/")

	if !strings.HasPrefix(address, "http://") &&
		!strings.HasPrefix(address, "https://") {

		address = "http://" + address
	}

	return address
}
func FetchNextNonce(
	nodeAddress string,
	senderAddress string,
) (int, error) {
	requestURL := normalizeNodeURL(nodeAddress) +
		"/nonce?address=" + senderAddress

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	response, err := client.Get(requestURL)

	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"node returned status %d",
			response.StatusCode,
		)
	}

	var nonceResponse NonceResponse

	if err := json.NewDecoder(response.Body).Decode(
		&nonceResponse,
	); err != nil {
		return 0, err
	}

	return nonceResponse.Nonce, nil
}
func SubmitTransaction(
	nodeAddress string,
	tx ledger.Transaction,
) error {
	transactionJSON, err := json.Marshal(tx)

	if err != nil {
		return err
	}

	requestURL := normalizeNodeURL(nodeAddress) +
		"/transactions"

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Post(
		requestURL,
		"application/json",
		bytes.NewReader(transactionJSON),
	)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf(
			"transaction rejected with status %d",
			response.StatusCode,
		)
	}

	return nil
}
