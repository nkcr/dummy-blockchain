package blockchain

import "fmt"

// NewNode returns a new node
func NewNode(host string, port int) *Node {
	return &Node{
		Host: host,
		Port: port,
	}
}

// Node represents a node running a blockchain
type Node struct {
	Host string
	Port int
}

// GetHTTP returns the url as http://<host>:<port>
func (n Node) GetHTTP() string {
	return fmt.Sprintf("http://%s:%d", n.Host, n.Port)
}
