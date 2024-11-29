package repository_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/devshark/tx-parser-go/api"
	"github.com/devshark/tx-parser-go/app/internal/repository"
)

func TestNewInMemoryRepository(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
}

func TestSaveTransaction(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	// Test successful case
	tx := api.Transaction{
		Hash:  "0x123",
		From:  "0xabc",
		To:    "0xdef",
		Value: big.NewInt(100),
	}

	err := repo.SaveTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test saving transaction with empty From address
	txEmptyFrom := api.Transaction{
		Hash:  "0x456",
		From:  "",
		To:    "0xdef",
		Value: big.NewInt(200),
	}

	err = repo.SaveTransaction(ctx, txEmptyFrom)
	if err == nil {
		t.Fatal("Expected error when saving transaction with empty From address, got nil")
	}

	// Test saving transaction with empty To address
	txEmptyTo := api.Transaction{
		Hash:  "0x789",
		From:  "0xabc",
		To:    "",
		Value: big.NewInt(300),
	}

	err = repo.SaveTransaction(ctx, txEmptyTo)
	if err == nil {
		t.Fatal("Expected error when saving transaction with empty To address, got nil")
	}
}

func TestGetTransactions(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	address := "0xabc"
	tx1 := api.Transaction{Hash: "0x123", From: address, To: "0xdef", Value: big.NewInt(100)}
	tx2 := api.Transaction{Hash: "0x456", From: "0xghi", To: address, Value: big.NewInt(200)}

	repo.SaveTransaction(ctx, tx1)
	repo.SaveTransaction(ctx, tx2)

	// Test successful case
	txs, err := repo.GetTransactions(ctx, address)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}
	if !containsTransaction(txs, tx1) || !containsTransaction(txs, tx2) {
		t.Errorf("Transactions not found in result")
	}

	// Test getting transactions for non-existent address
	nonExistentAddress := "0xnonexistent"
	txs, err = repo.GetTransactions(ctx, nonExistentAddress)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("Expected 0 transactions for non-existent address, got %d", len(txs))
	}

	// Test getting transactions with empty address
	_, err = repo.GetTransactions(ctx, "")
	if err == nil {
		t.Fatal("Expected error when getting transactions for empty address, got nil")
	}
}

func TestSubscribe(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	// Test successful case
	address := "0xabc"
	err := repo.Subscribe(ctx, address)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test subscribing with empty address
	err = repo.Subscribe(ctx, "")
	if err == nil {
		t.Fatal("Expected error when subscribing with empty address, got nil")
	}

	// Test subscribing to already subscribed address
	err = repo.Subscribe(ctx, address)
	if err != nil {
		t.Fatalf("Unexpected error when subscribing to already subscribed address: %v", err)
	}
}

func TestIsSubscribed(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	address1 := "0xabc"
	address2 := "0xdef"

	repo.Subscribe(ctx, address1)

	// Test successful cases
	subscribed1, err := repo.IsSubscribed(ctx, address1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !subscribed1 {
		t.Error("Expected address1 to be subscribed")
	}

	subscribed2, err := repo.IsSubscribed(ctx, address2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if subscribed2 {
		t.Error("Expected address2 to not be subscribed")
	}

	// Test with empty address
	_, err = repo.IsSubscribed(ctx, "")
	if err == nil {
		t.Fatal("Expected error when checking subscription for empty address, got nil")
	}
}

func TestGetLastParsedBlock(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	// Test initial state
	initialBlock, err := repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if initialBlock != 0 {
		t.Errorf("Expected initial block to be 0, got %d", initialBlock)
	}

	// Test after update
	newBlock := int64(1000)
	err = repo.UpdateLastParsedBlock(ctx, newBlock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lastParsedBlock, err := repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if lastParsedBlock != newBlock {
		t.Errorf("Expected last parsed block to be %d, got %d", newBlock, lastParsedBlock)
	}
}

func TestUpdateLastParsedBlock(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	// Test successful update
	newBlock := int64(1000)
	err := repo.UpdateLastParsedBlock(ctx, newBlock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lastParsedBlock, err := repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if lastParsedBlock != newBlock {
		t.Errorf("Expected last parsed block to be %d, got %d", newBlock, lastParsedBlock)
	}

	// Test update with negative block number
	negativeBlock := int64(-1)
	err = repo.UpdateLastParsedBlock(ctx, negativeBlock)
	if err == nil {
		t.Fatal("Expected error when updating with negative block number, got nil")
	}

	// Test update with lower block number
	lowerBlock := int64(500)
	err = repo.UpdateLastParsedBlock(ctx, lowerBlock)
	if err == nil {
		t.Fatal("Expected error when updating with lower block number, got nil")
	}
}

// Helper function to check if a slice of transactions contains a specific transaction
func containsTransaction(txs []api.Transaction, tx api.Transaction) bool {
	for _, t := range txs {
		if reflect.DeepEqual(t, tx) {
			return true
		}
	}
	return false
}

func TestConcurrentSaveTransaction(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			tx := api.Transaction{
				Hash:  fmt.Sprintf("0x%d", i),
				From:  "0xabc",
				To:    "0xdef",
				Value: big.NewInt(int64(i)),
			}
			err := repo.SaveTransaction(ctx, tx)
			if err != nil {
				t.Errorf("Unexpected error in goroutine %d: %v", i, err)
			}
		}(i)
	}

	wg.Wait()

	txs, err := repo.GetTransactions(ctx, "0xabc")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(txs) != numGoroutines {
		t.Errorf("Expected %d transactions, got %d", numGoroutines, len(txs))
	}
}

func TestConcurrentSubscribe(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			address := fmt.Sprintf("0x%d", i)
			err := repo.Subscribe(ctx, address)
			if err != nil {
				t.Errorf("Unexpected error in goroutine %d: %v", i, err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		address := fmt.Sprintf("0x%d", i)
		subscribed, err := repo.IsSubscribed(ctx, address)
		if err != nil {
			t.Errorf("Unexpected error checking subscription for address %s: %v", address, err)
		}
		if !subscribed {
			t.Errorf("Expected address %s to be subscribed", address)
		}
	}
}

func TestConcurrentUpdateLastParsedBlock(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			err := repo.UpdateLastParsedBlock(ctx, int64(i))
			// exclude invalid block error
			if err != nil && !errors.Is(err, repository.ErrInvalidBlock) {
				t.Errorf("Unexpected error in goroutine %d: %v", i, err)
			}
		}(i)
	}

	wg.Wait()

	lastParsedBlock, err := repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if lastParsedBlock < int64(numGoroutines-1) {
		t.Errorf("Expected last parsed block to be at least %d, got %d", numGoroutines-1, lastParsedBlock)
	}
}

func TestConcurrentGetTransactions(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	// Prepare some transactions
	for i := 0; i < 1000; i++ {
		tx := api.Transaction{
			Hash:  fmt.Sprintf("0x%d", i),
			From:  "0xabc",
			To:    "0xdef",
			Value: big.NewInt(int64(i)),
		}
		repo.SaveTransaction(ctx, tx)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			txs, err := repo.GetTransactions(ctx, "0xabc")
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(txs) != 1000 {
				t.Errorf("Expected 1000 transactions, got %d", len(txs))
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentSaveAndGetTransactions(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			tx := api.Transaction{
				Hash:  fmt.Sprintf("0x%d", i),
				From:  "0xabc",
				To:    "0xdef",
				Value: big.NewInt(int64(i)),
			}
			err := repo.SaveTransaction(ctx, tx)
			if err != nil {
				t.Errorf("Unexpected error in save goroutine %d: %v", i, err)
			}
		}(i)

		go func() {
			defer wg.Done()
			_, err := repo.GetTransactions(ctx, "0xabc")
			if err != nil {
				t.Errorf("Unexpected error in get goroutine: %v", err)
			}
		}()
	}

	wg.Wait()

	txs, err := repo.GetTransactions(ctx, "0xabc")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(txs) != numGoroutines {
		t.Errorf("Expected %d transactions, got %d", numGoroutines, len(txs))
	}
}

func TestConcurrentSubscribeAndIsSubscribed(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 goroutines per iteration

	subscribeErrors := make(chan error, numGoroutines)
	isSubscribedErrors := make(chan error, numGoroutines)
	isSubscribedResults := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		address := fmt.Sprintf("0x%d", i)

		// Goroutine 1: Subscribe
		go func(addr string) {
			defer wg.Done()
			err := repo.Subscribe(ctx, addr)
			if err != nil {
				subscribeErrors <- fmt.Errorf("error subscribing address %s: %v", addr, err)
			}
		}(address)

		// Goroutine 2: IsSubscribed (should eventually return true)
		go func(addr string) {
			defer wg.Done()
			var subscribed bool
			var err error
			for j := 0; j < 10; j++ { // Retry up to 10 times
				time.Sleep(time.Millisecond * 10) // Small delay between retries
				subscribed, err = repo.IsSubscribed(ctx, addr)
				if err != nil {
					isSubscribedErrors <- fmt.Errorf("error checking subscription for address %s: %v", addr, err)
					return
				}
				if subscribed {
					isSubscribedResults <- true
					return
				}
			}
			isSubscribedResults <- false
		}(address)

		// Goroutine 3: IsSubscribed with invalid address (should fail)
		go func() {
			defer wg.Done()
			invalidAddress := "" // Empty address should cause an error
			_, err := repo.IsSubscribed(ctx, invalidAddress)
			if err == nil {
				isSubscribedErrors <- fmt.Errorf("expected error for invalid address, got nil")
			}
		}()
	}

	wg.Wait()
	close(subscribeErrors)
	close(isSubscribedErrors)
	close(isSubscribedResults)

	// Check for subscription errors
	for err := range subscribeErrors {
		t.Error(err)
	}

	// Check for IsSubscribed errors
	for err := range isSubscribedErrors {
		t.Error(err)
	}

	// Check IsSubscribed results
	successfulChecks := 0
	for result := range isSubscribedResults {
		if result {
			successfulChecks++
		}
	}

	if successfulChecks != numGoroutines {
		t.Errorf("Expected %d successful IsSubscribed checks, got %d", numGoroutines, successfulChecks)
	}

	// Final check: all addresses should be subscribed
	for i := 0; i < numGoroutines; i++ {
		address := fmt.Sprintf("0x%d", i)
		subscribed, err := repo.IsSubscribed(ctx, address)
		if err != nil {
			t.Errorf("Unexpected error checking final subscription status for address %s: %v", address, err)
		}
		if !subscribed {
			t.Errorf("Expected address %s to be subscribed in final check", address)
		}
	}
}
