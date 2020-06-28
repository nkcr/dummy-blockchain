package blockchain

// GetCHainResponse is the response sent to a get chain request
type GetCHainResponse struct {
	Numblocks  int
	Blockchain *Blockchain
}
