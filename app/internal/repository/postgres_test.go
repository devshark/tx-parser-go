// File: app/internal/repository/postgres_test.go

package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/devshark/tx-parser-go/api"
	"github.com/devshark/tx-parser-go/app/internal/repository"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	if testing.Short() {
		t.Skip("Skipping test in short mode, as it requires a real database")
	}

	// Use environment variables for connection details in a real scenario
	connStr := "user=postgres password=postgres host=localhost port=5433 dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func setupPostgresNewPostgresRepository(t *testing.T, ctx context.Context) (repository.Repository, *sql.DB, func()) {
	db, cleanup := setupTestDB(t)

	repo := repository.NewPostgresRepository(db)
	if repo == nil {
		t.Fatal("NewPostgresRepository returned nil")
	}

	repo.(*repository.PostgresRepository).CreateTables(ctx)

	return repo, db, func() {
		repo.(*repository.PostgresRepository).DropTables(ctx)
		cleanup()
	}
}

func TestPostgresCreateTables(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, db, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	// Verify tables were created
	tables := []string{"transactions", "subscriptions", "blocks"}
	for _, table := range tables {
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", table).Scan(&exists)
		if err != nil {
			t.Fatalf("Error checking if table %s exists: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s was not created", table)
		}
	}
}

func TestPostgresPostgresSaveTransaction(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, db, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	tx := api.Transaction{
		Hash:  "0x123",
		From:  "0xabc",
		To:    "0xdef",
		Value: 100,
	}

	err = repo.SaveTransaction(ctx, tx.From, tx)
	if err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// Verify transaction was saved
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM transactions WHERE hash = $1", tx.Hash).Scan(&count)
	if err != nil {
		t.Fatalf("Error checking if transaction was saved: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 transaction, got %d", count)
	}
}

func TestPostgresPostgresGetTransactions(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, _, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	// Save some transactions
	txs := []api.Transaction{
		{Hash: "0x123", From: "0xabc", To: "0xdef", Value: 100},
		{Hash: "0x456", From: "0xabc", To: "0xghi", Value: 200},
		{Hash: "0x789", From: "0xdef", To: "0xabc", Value: 300},
	}

	for _, tx := range txs {
		err = repo.SaveTransaction(ctx, tx.From, tx)
		if err != nil {
			t.Fatalf("Failed to save transaction: %v", err)
		}
	}

	// Get transactions
	retrievedTxs, err := repo.GetTransactions(ctx, "0xabc")
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(retrievedTxs) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(retrievedTxs))
	}
}

func TestPostgresPostgresSubscribe(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, db, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	address := "0xabc"
	err = repo.Subscribe(ctx, address)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Verify subscription
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscriptions WHERE address = $1", address).Scan(&count)
	if err != nil {
		t.Fatalf("Error checking if address was subscribed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 subscription, got %d", count)
	}
}

func TestPostgresPostgresIsSubscribed(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, _, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	address := "0xabc"
	err = repo.Subscribe(ctx, address)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	subscribed, err := repo.IsSubscribed(ctx, address)
	if err != nil {
		t.Fatalf("Failed to check subscription: %v", err)
	}
	if !subscribed {
		t.Errorf("Expected address to be subscribed")
	}

	notSubscribed, err := repo.IsSubscribed(ctx, "0xdef")
	if err != nil {
		t.Fatalf("Failed to check subscription: %v", err)
	}
	if notSubscribed {
		t.Errorf("Expected address to not be subscribed")
	}
}

func TestPostgresGetLastParsedBlock(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, db, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	// Insert a block
	_, err = db.ExecContext(ctx, "INSERT INTO blocks (block_number) VALUES ($1)", 12345)
	if err != nil {
		t.Fatalf("Failed to insert block: %v", err)
	}

	lastBlock, err := repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get last parsed block: %v", err)
	}
	if lastBlock != 12345 {
		t.Errorf("Expected last block to be 12345, got %d", lastBlock)
	}
}

func TestPostgresUpdateLastParsedBlock(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, _, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	// Update last parsed block
	newBlock := int64(54321)
	err = repo.UpdateLastParsedBlock(ctx, newBlock)
	if err != nil {
		t.Fatalf("Failed to update last parsed block: %v", err)
	}

	// Verify the update
	lastBlock, err := repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get last parsed block: %v", err)
	}
	if lastBlock != newBlock {
		t.Errorf("Expected last block to be %d, got %d", newBlock, lastBlock)
	}

	// Test updating with a lower block number
	lowerBlock := int64(12345)
	err = repo.UpdateLastParsedBlock(ctx, lowerBlock)
	if err != nil {
		t.Fatalf("Failed to update last parsed block with lower number: %v", err)
	}

	// The block number should remain the same as before
	lastBlock, err = repo.GetLastParsedBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get last parsed block: %v", err)
	}
	if lastBlock != newBlock {
		t.Errorf("Expected last block to remain %d, got %d", newBlock, lastBlock)
	}
}

func TestPostgresConcurrentOperations(t *testing.T) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, db, cleanup := setupPostgresNewPostgresRepository(t, ctx)
	defer cleanup()

	// Concurrent subscriptions
	concurrency := 10
	address := "0xabc"
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			errChan <- repo.Subscribe(ctx, address)
		}()
	}

	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent subscription failed: %v", err)
		}
	}

	// Verify only one subscription exists
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscriptions WHERE address = $1", address).Scan(&count)
	if err != nil {
		t.Fatalf("Error checking subscriptions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 subscription, got %d", count)
	}

	// Concurrent transaction saves
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			tx := api.Transaction{
				Hash:  fmt.Sprintf("0x%d", i),
				From:  address,
				To:    "0xdef",
				Value: int64(i * 100),
			}
			errChan <- repo.SaveTransaction(ctx, tx.From, tx)
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent transaction save failed: %v", err)
		}
	}

	// Verify all transactions were saved
	txs, err := repo.GetTransactions(ctx, address)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}
	if len(txs) != concurrency {
		t.Errorf("Expected %d transactions, got %d", concurrency, len(txs))
	}
}
