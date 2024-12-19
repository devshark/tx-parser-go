package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/devshark/tx-parser-go/api"
)

var (
	ErrEmptyAddress  = errors.New("address cannot be empty")
	ErrNegativeBlock = errors.New("block number cannot be negative")
	ErrInvalidBlock  = errors.New("block number is not valid")
)

// type Repository interface {
// 	BlockRepository
// 	SubscriberRepository
// 	TransactionRepository
// }

// type implRepository struct {
// 	BlockRepository
// 	SubscriberRepository
// 	TransactionRepository
// }

// func NewRepository(blockRepo BlockRepository, txRepo TransactionRepository, subRepo SubscriberRepository) Repository {
// 	return &implRepository{
// 		BlockRepository:       blockRepo,
// 		TransactionRepository: txRepo,
// 		SubscriberRepository:  subRepo,
// 	}
// }

// Repository interface for data storage
type BlockRepository interface {
	GetLastParsedBlock(ctx context.Context) (int64, error)
	UpdateLastParsedBlock(ctx context.Context, blockNumber int64) error
}

type SubscriberRepository interface {
	Subscribe(ctx context.Context, address string) error
	IsSubscribed(ctx context.Context, address string) (bool, error)
}

type TransactionRepository interface {
	SaveTransaction(ctx context.Context, address string, tx api.Transaction) error
	GetTransactions(ctx context.Context, address string) ([]api.Transaction, error)
}

func CleanAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}

func ValidateAddress(address string) (string, error) {
	cleanAddress := CleanAddress(address)
	if cleanAddress == "" {
		return cleanAddress, ErrEmptyAddress
	}

	return cleanAddress, nil
}

func ValidateBlock(ctx context.Context, nextBlock int64) (bool, error) {
	if nextBlock < 0 {
		return false, ErrNegativeBlock
	}

	return true, nil
}
