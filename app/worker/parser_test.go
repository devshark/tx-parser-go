package worker_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/devshark/tx-parser-go/api"
	"github.com/devshark/tx-parser-go/app/internal/repository"
	"github.com/devshark/tx-parser-go/app/worker"
)

// MockBlockchainClient implements blockchain.BlockchainClient for testing
type MockBlockchainClient struct {
	latestBlockNumber int64
	blocks            map[int64]*api.Block
	mu                sync.RWMutex
}

func (m *MockBlockchainClient) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.latestBlockNumber, nil
}

func (m *MockBlockchainClient) GetBlockByNumber(ctx context.Context, number int64) (*api.Block, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blocks[number], nil
}

func TestNewParserWorker(t *testing.T) {
	mockBC := &MockBlockchainClient{}
	mockRepo := repository.NewInMemoryRepository() // it's fine to use this as the mock, it's just using memory anyway
	logger := log.Default()

	worker := worker.NewParserWorker(mockBC, mockRepo, logger)

	if worker == nil {
		t.Fatal("NewParserWorker returned nil")
	}
}

func TestParserWorker_Run(t *testing.T) {
	mockBC := &MockBlockchainClient{
		latestBlockNumber: 10,
		blocks: map[int64]*api.Block{
			1: {Number: 1, Transactions: []api.Transaction{{From: "0x1", To: "0x2", Hash: "0x999", Value: 100}}},
			2: {Number: 2, Transactions: []api.Transaction{{From: "0x2", To: "0x3", Hash: "0x888", Value: 200}}},
		},
	}
	mockRepo := repository.NewInMemoryRepository()
	logger := log.Default()

	worker := worker.NewParserWorker(mockBC, mockRepo, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mockRepo.Subscribe(ctx, "0x1")
	mockRepo.Subscribe(ctx, "0x2")

	err := worker.Run(ctx, 0, 100*time.Millisecond)

	<-time.After(1 * time.Second)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if val, _ := mockRepo.GetLastParsedBlock(ctx); val != 10 {
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
		if subscribed, _ := mockRepo.IsSubscribed(ctx, c.address); !subscribed {
			t.Errorf("Expected %s to be subscribed", c.address)
		}

		if txs, _ := mockRepo.GetTransactions(ctx, c.address); len(txs) != c.size {
			t.Errorf("Expected %d transactions for %s, got %d", c.size, c.address, len(txs))
		}
	}
}
