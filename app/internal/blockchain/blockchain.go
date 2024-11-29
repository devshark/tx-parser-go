package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/devshark/tx-parser-go/api"
)

// BlockchainClient interface for interacting with Ethereum
type BlockchainClient interface {
	GetLatestBlockNumber(ctx context.Context) (int64, error)
	GetBlockByNumber(ctx context.Context, number int64) (*api.Block, error)
	// Add more methods as needed
}

// uses big.Int to parse the hex string
func HexToInt64(hexStr string) (int64, error) {
	bigInt := new(big.Int)

	// Parse the hex string (without the "0x" prefix)
	bigInt.SetString(hexStr[2:], 16)

	return bigInt.Int64(), nil
}

func HexToInt[T int | uint | uint64](hexStr string) (T, error) {
	// Remove the "0x" prefix if it exists
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Parse the hexadecimal string to int64
	unixTimestamp, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hexadecimal string: %w", err)
	}

	return T(unixTimestamp), nil
}

func HexToTime(hexStr string) (time.Time, error) {
	unixTimestamp, err := HexToInt64(hexStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse hexadecimal string: %w", err)
	}

	// Convert Unix timestamp to time.Time
	return time.Unix(unixTimestamp, 0), nil
}
