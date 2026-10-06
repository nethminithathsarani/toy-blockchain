package ledger

type Transaction struct {
	ID            string
	Sender        string
	SenderAddress string
	Receiver      string
	Amount        int
	Nonce         int
	Signature     string
	PublicKey     string
}
