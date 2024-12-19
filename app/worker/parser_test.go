package worker_test

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devshark/tx-parser-go/api"
	"github.com/devshark/tx-parser-go/app/internal/repository"
	"github.com/devshark/tx-parser-go/app/worker"
)

// MockBlockchainClient implements blockchain.BlockchainClient for testing
type MockBlockchainClient struct {
	initialBlockNumber      int64
	latestBlockNumber       int64
	blocks                  map[int64]*api.Block
	mu                      sync.RWMutex
	getLastParsedBlockCalls atomic.Int32
}

func (m *MockBlockchainClient) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// the first time we call GetLastParsedBlock, we need to return the first block number
	// because the worker will start from the first block and work it's way up to the latest block
	// all subsequent calls will just return the latest block number
	m.getLastParsedBlockCalls.Add(1)
	if m.getLastParsedBlockCalls.Load() == 1 {
		return m.initialBlockNumber, nil
	}

	return m.latestBlockNumber, nil
}

func (m *MockBlockchainClient) GetBlockByNumber(ctx context.Context, number int64) (*api.Block, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blocks[number], nil
}

func TestNewParserWorker(t *testing.T) {
	mockBC := &MockBlockchainClient{}

	mockTxRepo := repository.NewInMemoryTransactionRepository()
	mockSubRepo := repository.NewInMemorySubscriberRepository()
	mockBlockRepo := repository.NewInMemoryBlockRepository()
	logger := log.Default()

	worker := worker.
		NewParserWorker(mockBC, mockTxRepo, mockSubRepo, mockBlockRepo).
		WithCustomLogger(logger)

	if worker == nil {
		t.Fatal("NewParserWorker returned nil")
	}
}

func TestParserWorker_Run(t *testing.T) {
	mockBC := &MockBlockchainClient{
		initialBlockNumber: 0,
		latestBlockNumber:  10,
		blocks: map[int64]*api.Block{
			1: {Number: 1, Transactions: []api.Transaction{{From: "0x1", To: "0x2", Hash: "0x999", Value: 100}}},
			2: {Number: 2, Transactions: []api.Transaction{{From: "0x2", To: "0x3", Hash: "0x888", Value: 200}}},
		},
	}

	mockTxRepo := repository.NewInMemoryTransactionRepository()
	mockSubRepo := repository.NewInMemorySubscriberRepository()
	mockBlockRepo := repository.NewInMemoryBlockRepository()
	logger := log.Default()

	worker := worker.
		NewParserWorker(mockBC, mockTxRepo, mockSubRepo, mockBlockRepo).
		WithCustomLogger(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mockSubRepo.Subscribe(ctx, "0x1")
	mockSubRepo.Subscribe(ctx, "0x2")

	err := worker.Run(ctx, 100*time.Millisecond)

	<-time.After(200 * time.Millisecond)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if val, _ := mockBlockRepo.GetLastParsedBlock(ctx); val != 10 {
		t.Errorf("Expected last parsed block to be 10, got %d", val)
	}

	cases := []struct {
		address string
		size    int
	}{
		{"0x1", 1},
		{"0x2", 2},
	}

	for _, c := range cases {
		if subscribed, _ := mockSubRepo.IsSubscribed(ctx, c.address); !subscribed {
			t.Errorf("Expected %s to be subscribed", c.address)
		}

		if txs, _ := mockTxRepo.GetTransactions(ctx, c.address); len(txs) != c.size {
			t.Errorf("Expected %d transactions for %s, got %d", c.size, c.address, len(txs))
		}
	}
}
