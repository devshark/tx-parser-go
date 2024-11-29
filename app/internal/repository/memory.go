package repository

import (
	"context"
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

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		transactions: make(map[string][]api.Transaction),
	}
}
func (r *InMemoryRepository) SaveTransaction(ctx context.Context, address string, tx api.Transaction) error {
	address = strings.TrimSpace(strings.ToLower(address))

	if address == "" {
		return ErrEmptyAddress
	}

	r.muTx.Lock()
	defer r.muTx.Unlock()

	// skip if tx hash already exists
	txs, ok := r.transactions[address]
	if ok {
		for _, _tx := range txs {
			if strings.EqualFold(_tx.Hash, tx.Hash) {
				return nil
			}
		}
	}

	r.transactions[address] = append(txs, tx)

	return nil
}

func (r *InMemoryRepository) GetTransactions(ctx context.Context, address string) ([]api.Transaction, error) {
	address = strings.TrimSpace(strings.ToLower(address))

	if address == "" {
		return nil, ErrEmptyAddress
	}

	r.muTx.RLock()
	defer r.muTx.RUnlock()

	return r.transactions[address], nil
}

// Subscribe creates a new transaction slice for the given address if it doesn't exist; does not overwrite existing address' transactions
func (r *InMemoryRepository) Subscribe(ctx context.Context, address string) error {
	address = strings.TrimSpace(strings.ToLower(address))

	if address == "" {
		return ErrEmptyAddress
	}

	r.muTx.Lock()
	defer r.muTx.Unlock()

	if _, exists := r.transactions[address]; !exists {
		r.transactions[address] = make([]api.Transaction, 0)
	}

	return nil
}

func (r *InMemoryRepository) IsSubscribed(ctx context.Context, address string) (bool, error) {
	address = strings.TrimSpace(strings.ToLower(address))

	if address == "" {
		return false, ErrEmptyAddress
	}

	r.muTx.RLock()
	defer r.muTx.RUnlock()

	_, exists := r.transactions[address]

	return exists, nil
}

func (r *InMemoryRepository) GetLastParsedBlock(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.lastParsedBlock, nil
}

func (r *InMemoryRepository) UpdateLastParsedBlock(ctx context.Context, blockNumber int64) error {
	if blockNumber < 0 {
		return ErrNegativeBlock
	}

	if blockNumber < r.lastParsedBlock {
		return ErrInvalidBlock
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastParsedBlock = blockNumber
	return nil
}
