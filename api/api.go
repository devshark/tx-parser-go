package api

import (
	"context"
	"math/big"
	"time"
)

// Block represents an Ethereum block
type Block struct {
	Number       *big.Int      `json:"number"`
	Hash         string        `json:"hash"`
	ParentHash   string        `json:"parentHash"`
	Nonce        string        `json:"nonce"`
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	// Add more fields as needed, such as:
	// GasUsed      *big.Int       `json:"gasUsed"`
	// GasLimit     *big.Int       `json:"gasLimit"`
	// Difficulty   *big.Int       `json:"difficulty"`
	// Miner        string         `json:"miner"`
}

// Transaction represents an Ethereum transaction within a block
type Transaction struct {
	Hash             string   `json:"hash"`
	From             string   `json:"from"`
	To               string   `json:"to"`
	Value            *big.Int `json:"value"`
	Gas              uint64   `json:"gas"`
	GasPrice         *big.Int `json:"gasPrice"`
	Input            string   `json:"input"`
	Nonce            uint64   `json:"nonce"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      *big.Int `json:"blockNumber"`
	TransactionIndex uint     `json:"transactionIndex"`
}

// Parser interface as defined in the requirements
type Parser interface {
	GetCurrentBlock() int64
	Subscribe(address string) bool
	GetTransactions(address string) []Transaction
	Parse(ctx context.Context) error
}
