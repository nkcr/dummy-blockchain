package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/xerrors"
)

// NewBlockchain creates a new blockchain
func NewBlockchain(address string) *Blockchain {
	blockchain := &Blockchain{
		Chain:        make([]*Block, 0),
		Transactions: make([]*Transaction, 0),
		Nodes:        make([]*Node, 0),
		Address:      address,
	}

	blockchain.CreateBlock(0, [32]byte{})

	return blockchain
}

// Blockchain represents a node holding a chain of blocks.
type Blockchain struct {
	Chain        []*Block
	Transactions []*Transaction
	Nodes        []*Node
	Address      string
}

// CreateBlock creates a block and appends it to the chain
func (b *Blockchain) CreateBlock(proof int, prevHash [32]byte) *Block {
	block := NewBlock(len(b.Chain), proof, prevHash, b.Transactions)

	b.Chain = append(b.Chain, block)
	b.Transactions = make([]*Transaction, 0)

	return block
}

// GetPreviousBlock returns the last block stored. This function panics if the
// chain is empty.
func (b *Blockchain) GetPreviousBlock() *Block {
	return b.Chain[len(b.Chain)-1]
}

// ProofOfWork calculates the right nounce, ie. the proof
func (b *Blockchain) ProofOfWork(prevProof int) int {
	newProof := 0

	for {
		hashOperation := sha256.Sum256([]byte(fmt.Sprintf("%d",
			newProof*newProof-prevProof*prevProof)))

		if strings.HasPrefix(hex.EncodeToString(hashOperation[:]), "0000") {
			break
		}
		newProof++
	}

	return newProof
}

// IsCHainValid checks that the given chain is valid
func (b *Blockchain) IsCHainValid(blocks []*Block) (bool, error) {
	if len(blocks) == 0 {
		return false, xerrors.Errorf("chain is empty")
	}

	prevBlock := blocks[0]

	for _, block := range blocks[1:] {
		// 1: check the prev hash: the prevHash of the block should be the same
		// as the hash of the previous block
		prevHash, err := prevBlock.Hash()
		if err != nil {
			return false, xerrors.Errorf("failed to get hash: %v", err)
		}
		if bytes.Compare(block.PrevHash[:], prevHash[:]) != 0 {
			return false, nil
		}

		// 2: check the proof: we apply the same hashOperation as in the
		// ProofOfWork function and check if it returns a correct hash
		prevProof := prevBlock.Proof
		proof := block.Proof
		hashOperation := sha256.Sum256([]byte(fmt.Sprintf("%d", proof*proof-
			prevProof*prevProof)))
		if !strings.HasPrefix(hex.EncodeToString(hashOperation[:]), "0000") {
			return false, nil
		}

		prevBlock = block
	}

	return true, nil
}

// AddTransaction adds a new transaction to the list of transactions. Returns
// the block index of the block that will contain the transaction.
func (b *Blockchain) AddTransaction(t *Transaction) int {
	b.Transactions = append(b.Transactions, t)

	return b.GetPreviousBlock().Index + 1
}

// AddNode adds a new node to the list of nodes
func (b *Blockchain) AddNode(node *Node) {
	b.Nodes = append(b.Nodes, node)
}

// ReplaceChain checks the chains on all the other nodes and replace the current
// chain if it finds one longest than ours. Returns if the chain has been
// updated or not.
func (b *Blockchain) ReplaceChain() (bool, error) {
	maxLen := len(b.Chain)
	var longestChain []*Block

	for _, node := range b.Nodes {
		url := node.GetHTTP() + "/get_chain"
		resp, err := http.Get(url)
		if err != nil {
			return false, xerrors.Errorf("failed to call on '%s': %v", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, xerrors.Errorf("wrong status code: %s", resp.Status)
		}

		var chainResp GetCHainResponse
		err = json.NewDecoder(resp.Body).Decode(&chainResp)
		if err != nil {
			return false, xerrors.Errorf("failed to decode response: %v", err)
		}

		if chainResp.Numblocks > maxLen {
			maxLen = chainResp.Numblocks
			longestChain = chainResp.Blockchain.Chain
		}
	}

	if longestChain != nil {
		b.Chain = longestChain
		return true, nil
	}

	return false, nil

}
