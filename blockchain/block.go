package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"golang.org/x/xerrors"
)

// Hash represents the hash type. We define it to have a more convenient hash
// representation in JSON.
type Hash [32]byte

// MarshalJSON implements json.Encoder.
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h[:]))
}

// UnmarshalJSON implements json.Encoder
func (h *Hash) UnmarshalJSON(data []byte) error {
	var hashStr string
	err := json.Unmarshal(data, &hashStr)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal hash string: %v", err)
	}

	buf, err := hex.DecodeString(hashStr)
	if err != nil {
		return xerrors.Errorf("failed to decode Hash hex: %v", err)
	}

	if len(buf) != 32 {
		return xerrors.Errorf("len of hash should be == 32: %d", len(buf))
	}

	for i := range buf {
		h[i] = buf[i]
	}

	return nil
}

// String returns a string representation of a hash
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// NewBlock creates a new block
func NewBlock(index int, proof int, prevHash Hash, txs []*Transaction) *Block {
	return &Block{
		Index:        index,
		Timestamp:    time.Now().UnixNano(),
		Proof:        proof,
		PrevHash:     prevHash,
		Transactions: txs,
	}
}

// Block represents a block on the chain
type Block struct {
	Index        int
	Timestamp    int64
	Proof        int
	PrevHash     Hash
	Transactions []*Transaction
}

// Hash outputs the hash of the JSON representation of the block
func (b Block) Hash() ([32]byte, error) {
	encodedBlock, err := json.Marshal(b)
	if err != nil {
		return [32]byte{}, xerrors.Errorf("failed to marshal to json: %v", err)
	}

	return sha256.Sum256(encodedBlock), nil
}
