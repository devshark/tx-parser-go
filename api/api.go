package api

import (
	"time"
)

// Block represents an Ethereum block
type Block struct {
	Number       int64         `json:"number"`
	Hash         string        `json:"hash"`
	ParentHash   string        `json:"parentHash"`
	Nonce        string        `json:"nonce"`
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	// Add more fields as needed, such as:
	// GasUsed     int64       `json:"gasUsed"`
	// GasLimit    int64       `json:"gasLimit"`
	// Difficulty  int64       `json:"difficulty"`
	// Miner        string         `json:"miner"`
}

// Transaction represents an Ethereum transaction within a block
type Transaction struct {
	Hash             string `json:"hash"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            int64  `json:"value"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	BlockHash        string `json:"blockHash"`
	TransactionIndex uint   `json:"transactionIndex"`
	// Gas              uint64 `json:"gas"`
	// GasPrice         int64  `json:"gasPrice"`
	// BlockNumber      int64  `json:"blockNumber"`
}

// Parser interface as defined in the requirements
type Parser interface {
	// last parsed block
	GetCurrentBlock() int
	// add address to observer
	Subscribe(address string) bool
	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}
