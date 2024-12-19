package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devshark/tx-parser-go/api"
	"github.com/devshark/tx-parser-go/app/internal/blockchain"
	"github.com/devshark/tx-parser-go/app/internal/repository"
	"github.com/devshark/tx-parser-go/pkg/retry"
)

type ParserWorker struct {
	blockchain      blockchain.BlockchainClient
	transactionRepo repository.TransactionRepository
	subscriberRepo  repository.SubscriberRepository
	blockRepo       repository.BlockRepository
	logger          *log.Logger
}

// NewParserWorker creates a new ParserWorker with required arguments
func NewParserWorker(
	blockchain blockchain.BlockchainClient,
	transactionRepo repository.TransactionRepository,
	subscriberRepo repository.SubscriberRepository,
	blockRepo repository.BlockRepository) *ParserWorker {
	return &ParserWorker{
		blockchain:      blockchain,
		transactionRepo: transactionRepo,
		subscriberRepo:  subscriberRepo,
		blockRepo:       blockRepo,
		logger:          log.Default(),
	}
}

func (p *ParserWorker) WithCustomLogger(logger *log.Logger) *ParserWorker {
	p.logger = logger

	return p
}

// Run method with improved concurrency and error handling
func (p *ParserWorker) Run(ctx context.Context, schedule time.Duration) error {
	// fetch latest block number first
	// if it fails, return the error so it will handle the recovery sequence
	lastParsedBlock, err := p.blockchain.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %w", err)
	}

	// If the context is cancelled, exit immediately
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(schedule):
			// Get the latest block number
			latestBlock, err := p.blockchain.GetLatestBlockNumber(ctx)
			if err != nil {
				return err
			}

			// p.logger.Printf("last parsed block: %d, latest block: %d", lastParsedBlock, latestBlock)

			for _blockNum := lastParsedBlock + 1; _blockNum <= latestBlock; _blockNum++ {
				go func(blockNum int64) {
					// Set up a retry loop to parse the block
					action := func() error { return p.parseBlock(ctx, blockNum) }
					if err := retry.Retry(ctx, action, retry.DefaultMaxAttempts); err != nil {
						// Log any errors that happen, but don't crash
						p.logger.Printf("failed to parse block %d: %v", blockNum, err)
					}

					// p.logger.Print("parsed block ", blockNum)
				}(_blockNum)
			}

			p.blockRepo.UpdateLastParsedBlock(ctx, latestBlock)
			// Get the last block number that we've parsed
			lastParsedBlock = latestBlock
		}
	}
}

// parseBlock parses a single block
func (p *ParserWorker) parseBlock(ctx context.Context, blockNum int64) error {
	block, err := p.blockchain.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return err
	}

	if block == nil {
		return nil
	}

	for _, tx := range block.Transactions {
		if err := p.processTx(ctx, tx); err != nil {
			return err
		}
	}

	return nil
}

// processTx processes a single transaction
func (p *ParserWorker) processTx(ctx context.Context, tx api.Transaction) error {
	addresses := []string{tx.From, tx.To}

	for _, addr := range addresses {
		if strings.TrimSpace(addr) == "" { // just skip immediately if address is empty
			continue
		}

		subscribed, err := p.subscriberRepo.IsSubscribed(ctx, addr)
		if err != nil {
			return err
		}

		if subscribed {
			if err = p.transactionRepo.SaveTransaction(ctx, addr, tx); err != nil {
				return err
			}
		}
	}

	return nil
}
