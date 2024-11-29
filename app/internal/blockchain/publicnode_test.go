package blockchain_test

import (
	"context"
	"log"
	"reflect"
	"testing"

	"github.com/devshark/tx-parser-go/app/internal/blockchain"
)

func TestNewPublicNodeClient(t *testing.T) {
	// Test case 1: Valid URL
	validURL := "https://ethereum-rpc.publicnode.com/"
	client := blockchain.NewPublicNodeClient(validURL, log.Default())

	if client == nil {
		t.Fatal("NewPublicNodeClient returned nil for valid URL")
	}

	// Verify that the returned client implements BlockchainClient interface
	if _, ok := client.(blockchain.BlockchainClient); !ok {
		t.Fatal("Returned client does not implement BlockchainClient interface")
	}

	// Test case 2: Empty URL
	emptyURL := ""
	emptyClient := blockchain.NewPublicNodeClient(emptyURL, log.Default())

	if emptyClient == nil {
		t.Fatal("NewPublicNodeClient returned nil for empty URL")
	}

	// Test case 3: Invalid URL
	invalidURL := "not-a-valid-url"
	invalidClient := blockchain.NewPublicNodeClient(invalidURL, log.Default())

	if invalidClient == nil {
		t.Fatal("NewPublicNodeClient returned nil for invalid URL")
	}

	// Verify that all returned clients have the same type
	if reflect.TypeOf(client) != reflect.TypeOf(emptyClient) || reflect.TypeOf(client) != reflect.TypeOf(invalidClient) {
		t.Fatal("Returned clients have different types")
	}
}

func TestPublicNodeClientMethods(t *testing.T) {
	url := "https://ethereum-rpc.publicnode.com/"
	client := blockchain.NewPublicNodeClient(url, log.Default())

	// Test GetLatestBlockNumber method
	t.Run("GetLatestBlockNumber", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("GetLatestBlockNumber returned an error: %v", err)
		}
	})

	// Test GetBlockByNumber method
	t.Run("GetBlockByNumber", func(t *testing.T) {
		ctx := context.Background()
		blockNumber := int64(21291150) // Use a known block number
		block, err := client.GetBlockByNumber(ctx, blockNumber)
		if err != nil {
			t.Fatalf("GetBlockByNumber returned an error: %v", err)
		}
		if block == nil {
			t.Fatal("GetBlockByNumber returned nil block")
		}

		if block.Number != blockNumber {
			t.Fatalf("Expected block number %d, got %d", blockNumber, block.Number)
		}
	})
}

func TestPublicNodeClientErrorCases(t *testing.T) {
	// Use an invalid URL to simulate network errors
	invalidURL := "https://invalid-url.example.com/"
	client := blockchain.NewPublicNodeClient(invalidURL, log.Default())

	// Test GetLatestBlockNumber with invalid URL
	t.Run("GetLatestBlockNumber_InvalidURL", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.GetLatestBlockNumber(ctx)
		if err == nil {
			t.Fatal("Expected error for invalid URL, got nil")
		}
	})

	// Test GetBlockByNumber with invalid URL
	t.Run("GetBlockByNumber_InvalidURL", func(t *testing.T) {
		ctx := context.Background()
		blockNumber := int64(1000000)
		_, err := client.GetBlockByNumber(ctx, blockNumber)
		if err == nil {
			t.Fatal("Expected error for invalid URL, got nil")
		}
	})

	// Test GetBlockByNumber with negative block number
	t.Run("GetBlockByNumber_NegativeBlockNumber", func(t *testing.T) {
		ctx := context.Background()
		blockNumber := int64(-1)
		_, err := client.GetBlockByNumber(ctx, blockNumber)
		if err == nil {
			t.Fatal("Expected error for negative block number, got nil")
		}
	})
}
