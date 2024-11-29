package repository

import (
	"context"
	"errors"

	"github.com/devshark/tx-parser-go/api"
)

var (
	ErrEmptyAddress  = errors.New("address cannot be empty")
	ErrNegativeBlock = errors.New("block number cannot be negative")
	ErrInvalidBlock  = errors.New("block number is not valid")
)

// Repository interface for data storage
type Repository interface {
	SaveTransaction(ctx context.Context, address string, tx api.Transaction) error
	GetTransactions(ctx context.Context, address string) ([]api.Transaction, error)
	Subscribe(ctx context.Context, address string) error
	IsSubscribed(ctx context.Context, address string) (bool, error)
	GetLastParsedBlock(ctx context.Context) (int64, error)
	UpdateLastParsedBlock(ctx context.Context, blockNumber int64) error
	// Add more methods as needed
}
