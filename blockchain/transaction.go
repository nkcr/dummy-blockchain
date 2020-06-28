package blockchain

// NewTransaction returns a new transaction
func NewTransaction(sender, receiver string, amount int) *Transaction {
	return &Transaction{
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
	}
}

// Transaction represents a crypto currency transaction
type Transaction struct {
	Sender   string
	Receiver string
	Amount   int
}
