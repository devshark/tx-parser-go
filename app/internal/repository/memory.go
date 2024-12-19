package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/devshark/tx-parser-go/api"
)

// InMemoryRepository implements the Repository interface
type InMemoryRepository struct {
	muTx            sync.RWMutex // mutex for transactions
	transactions    map[string][]api.Transaction
	mu              sync.RWMutex // mutex for lastParsedBlock
	lastParsedBlock int64
}

type InMemoryTransactionRepository struct {
	sync.RWMutex
	transactions map[string][]api.Transaction
}

type InMemorySubscriberRepository struct {
	sync.RWMutex
	subscribers map[string]struct{}
}

type InMemoryBlockRepository struct {
	sync.RWMutex
	lastParsedBlock int64
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string][]api.Transaction),
	}
}

func (r *InMemoryTransactionRepository) SaveTransaction(ctx context.Context, address string, tx api.Transaction) error {
	r.Lock()
	defer r.Unlock()

	cleanAddress, err := ValidateAddress(address)
	if err != nil {
		return fmt.Errorf("ValidateAddress: %w", err)
	}

	// skip if tx hash already exists
	txs, ok := r.transactions[cleanAddress]
	if ok {
		for _, _tx := range txs {
			if strings.EqualFold(_tx.Hash, tx.Hash) {
				return nil
			}
		}
	}

	r.transactions[cleanAddress] = append(txs, tx)

	return nil
}

func (r *InMemoryTransactionRepository) GetTransactions(ctx context.Context, address string) ([]api.Transaction, error) {
	r.RLock()
	defer r.RUnlock()

	cleanAddress, err := ValidateAddress(address)
	if err != nil {
		return nil, fmt.Errorf("ValidateAddress: %w", err)
	}

	return r.transactions[cleanAddress], nil
}

func NewInMemorySubscriberRepository() *InMemorySubscriberRepository {
	return &InMemorySubscriberRepository{
		subscribers: make(map[string]struct{}),
	}
}

// Subscribe creates a new transaction slice for the given address if it doesn't exist; does not overwrite existing address' transactions
func (r *InMemorySubscriberRepository) Subscribe(ctx context.Context, address string) error {
	r.Lock()
	defer r.Unlock()

	cleanAddress, err := ValidateAddress(address)
	if err != nil {
		return fmt.Errorf("ValidateAddress: %w", err)
	}

	if _, exists := r.subscribers[cleanAddress]; !exists {
		r.subscribers[cleanAddress] = struct{}{}
	}

	return nil
}

func (r *InMemorySubscriberRepository) IsSubscribed(ctx context.Context, address string) (bool, error) {
	r.RLock()
	defer r.RUnlock()

	cleanAddress, err := ValidateAddress(address)
	if err != nil {
		return false, fmt.Errorf("ValidateAddress: %w", err)
	}

	_, exists := r.subscribers[cleanAddress]

	return exists, nil
}

func NewInMemoryBlockRepository() *InMemoryBlockRepository {
	return &InMemoryBlockRepository{}
}

func (r *InMemoryBlockRepository) GetLastParsedBlock(ctx context.Context) (int64, error) {
	r.RLock()
	defer r.RUnlock()

	return r.lastParsedBlock, nil
}

func (r *InMemoryBlockRepository) UpdateLastParsedBlock(ctx context.Context, blockNumber int64) error {
	r.Lock()
	defer r.Unlock()

	if valid, err := ValidateBlock(ctx, blockNumber); err != nil {
		return err
	} else if !valid {
		return fmt.Errorf("%w: %w", ErrInvalidBlock, err)
	}

	if blockNumber < r.lastParsedBlock {
		return ErrInvalidBlock
	}

	r.lastParsedBlock = blockNumber

	return nil
}
