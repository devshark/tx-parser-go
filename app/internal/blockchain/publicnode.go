package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/devshark/tx-parser-go/api"
)

// publicNodeClient is a BlockchainClient for interacting with a public Ethereum node
type publicNodeClient struct {
	publicNodeURL string
	logger        *log.Logger
}

func NewPublicNodeClient(publicNodeURL string, logger *log.Logger) BlockchainClient {
	return &publicNodeClient{publicNodeURL: publicNodeURL, logger: logger}
}

type EthBlock struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Hash                  string `json:"hash"`
		ParentHash            string `json:"parentHash"`
		Sha3Uncles            string `json:"sha3Uncles"`
		Miner                 string `json:"miner"`
		StateRoot             string `json:"stateRoot"`
		TransactionsRoot      string `json:"transactionsRoot"`
		ReceiptsRoot          string `json:"receiptsRoot"`
		LogsBloom             string `json:"logsBloom"`
		Difficulty            string `json:"difficulty"`
		Number                string `json:"number"`
		GasLimit              string `json:"gasLimit"`
		GasUsed               string `json:"gasUsed"`
		Timestamp             string `json:"timestamp"`
		ExtraData             string `json:"extraData"`
		MixHash               string `json:"mixHash"`
		Nonce                 string `json:"nonce"`
		BaseFeePerGas         string `json:"baseFeePerGas"`
		WithdrawalsRoot       string `json:"withdrawalsRoot"`
		BlobGasUsed           string `json:"blobGasUsed"`
		ExcessBlobGas         string `json:"excessBlobGas"`
		ParentBeaconBlockRoot string `json:"parentBeaconBlockRoot"`
		TotalDifficulty       string `json:"totalDifficulty"`
		Size                  string `json:"size"`
		Uncles                []any  `json:"uncles"`
		Transactions          []struct {
			Type                 string   `json:"type"`
			ChainID              string   `json:"chainId"`
			Nonce                string   `json:"nonce"`
			Gas                  string   `json:"gas"`
			MaxFeePerGas         string   `json:"maxFeePerGas,omitempty"`
			MaxPriorityFeePerGas string   `json:"maxPriorityFeePerGas,omitempty"`
			To                   string   `json:"to"`
			Value                string   `json:"value"`
			AccessList           []any    `json:"accessList,omitempty"`
			Input                string   `json:"input"`
			R                    string   `json:"r"`
			S                    string   `json:"s"`
			YParity              string   `json:"yParity,omitempty"`
			V                    string   `json:"v"`
			Hash                 string   `json:"hash"`
			BlockHash            string   `json:"blockHash"`
			BlockNumber          string   `json:"blockNumber"`
			TransactionIndex     string   `json:"transactionIndex"`
			From                 string   `json:"from"`
			GasPrice             string   `json:"gasPrice"`
			BlobVersionedHashes  []string `json:"blobVersionedHashes,omitempty"`
			MaxFeePerBlobGas     string   `json:"maxFeePerBlobGas,omitempty"`
		} `json:"transactions"`
		Withdrawals []struct {
			Index          string `json:"index"`
			ValidatorIndex string `json:"validatorIndex"`
			Address        string `json:"address"`
			Amount         string `json:"amount"`
		} `json:"withdrawals"`
	} `json:"result"`
}

// GetLatestBlockNumber fetches the latest block number
func (c *publicNodeClient) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	reqBody := strings.NewReader(`{
		"jsonrpc": "2.0",
		"method": "eth_blockNumber",
		"params": [],
		"id": 1
	}`)

	req, err := http.NewRequestWithContext(ctx, "POST", c.publicNodeURL, reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  string `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Parse the hexadecimal value to int64
	blockNum, err := HexToInt64(result.Result)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block number: %w", err)
	}

	return blockNum, nil
}

// GetBlockByNumber fetches the block with the given number
func (c *publicNodeClient) GetBlockByNumber(ctx context.Context, number int64) (*api.Block, error) {
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x",true],"id":1}`, number)

	req, err := http.NewRequestWithContext(ctx, "POST", c.publicNodeURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result EthBlock

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	txs := make([]api.Transaction, len(result.Result.Transactions))
	for i, t := range result.Result.Transactions {
		value, err := HexToInt64(t.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value: %w", err)
		}

		transactionIndex, err := HexToInt[uint](t.TransactionIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction index: %w", err)
		}

		txs[i] = api.Transaction{
			Hash:             t.Hash,
			From:             t.From,
			To:               t.To,
			Input:            t.Input,
			Nonce:            t.Nonce,
			Value:            value,
			BlockHash:        t.BlockHash,
			TransactionIndex: transactionIndex,
		}
	}

	timeStamp, err := HexToTime(result.Result.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	blockNumber, err := HexToInt64(result.Result.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block number: %w", err)
	}

	return &api.Block{
		Number:       blockNumber,
		Hash:         result.Result.Hash,
		ParentHash:   result.Result.ParentHash,
		Nonce:        result.Result.Nonce,
		Timestamp:    timeStamp,
		Transactions: txs,
	}, nil
}
