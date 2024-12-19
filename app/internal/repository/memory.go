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

	if blockNumber < r.lastParsedBlock {
		return ErrInvalidBlock
	}

	r.lastParsedBlock = blockNumber

	return nil
}

// func NewInMemoryRepository() Repository {
// 	return &InMemoryRepository{
// 		transactions: make(map[string][]api.Transaction),
// 	}
// }
// func (r *InMemoryRepository) SaveTransaction(ctx context.Context, address string, tx api.Transaction) error {
// 	address = strings.TrimSpace(strings.ToLower(address))

// 	if address == "" {
// 		return ErrEmptyAddress
// 	}

// 	r.muTx.Lock()
// 	defer r.muTx.Unlock()

// 	// skip if tx hash already exists
// 	txs, ok := r.transactions[address]
// 	if ok {
// 		for _, _tx := range txs {
// 			if strings.EqualFold(_tx.Hash, tx.Hash) {
// 				return nil
// 			}
// 		}
// 	}

// 	r.transactions[address] = append(txs, tx)

// 	return nil
// }

// func (r *InMemoryRepository) GetTransactions(ctx context.Context, address string) ([]api.Transaction, error) {
// 	address = strings.TrimSpace(strings.ToLower(address))

// 	if address == "" {
// 		return nil, ErrEmptyAddress
// 	}

// 	r.muTx.RLock()
// 	defer r.muTx.RUnlock()

// 	return r.transactions[address], nil
// }

// // Subscribe creates a new transaction slice for the given address if it doesn't exist; does not overwrite existing address' transactions
// func (r *InMemoryRepository) Subscribe(ctx context.Context, address string) error {
// 	address = strings.TrimSpace(strings.ToLower(address))

// 	if address == "" {
// 		return ErrEmptyAddress
// 	}

// 	r.muTx.Lock()
// 	defer r.muTx.Unlock()

// 	if _, exists := r.transactions[address]; !exists {
// 		r.transactions[address] = make([]api.Transaction, 0)
// 	}

// 	return nil
// }

// func (r *InMemoryRepository) IsSubscribed(ctx context.Context, address string) (bool, error) {
// 	address = strings.TrimSpace(strings.ToLower(address))

// 	if address == "" {
// 		return false, ErrEmptyAddress
// 	}

// 	r.muTx.RLock()
// 	defer r.muTx.RUnlock()

// 	_, exists := r.transactions[address]

// 	return exists, nil
// }

// func (r *InMemoryRepository) GetLastParsedBlock(ctx context.Context) (int64, error) {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()

// 	return r.lastParsedBlock, nil
// }

// func (r *InMemoryRepository) UpdateLastParsedBlock(ctx context.Context, blockNumber int64) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	if blockNumber < 0 {
// 		return ErrNegativeBlock
// 	}

// 	if blockNumber < r.lastParsedBlock {
// 		return ErrInvalidBlock
// 	}

// 	r.lastParsedBlock = blockNumber
// 	return nil
// }
