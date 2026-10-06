package wallet

import "toy-blockchain/ledger"

func SignTransaction(
	tx ledger.Transaction,
	w *Wallet,
) (ledger.Transaction, error) {

	tx.PublicKey = w.GetPublicKey()
	tx.SenderAddress = w.GetAddress()

	payload, err := TransactionSigningPayload(tx)

	if err != nil {
		return tx, err
	}

	signature, err := Sign(
		string(payload),
		w.PrivateKey,
	)

	if err != nil {
		return tx, err
	}

	tx.Signature = signature

	transactionID, err := TransactionID(tx)

	if err != nil {
		return tx, err
	}

	tx.ID = transactionID

	return tx, nil
}
