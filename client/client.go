package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/devshark/tx-parser-go/api"
)

type Client struct {
	baseUrl string
	logger  *log.Logger
}

func NewClient(baseUrl string) api.Parser {
	return &Client{
		baseUrl: baseUrl,
		logger:  log.Default(),
	}
}

type CurrentBlockResponse struct {
	BlockNumber int64 `json:"block_number"`
}

type AddressTransactionsResponse struct {
	Transactions []api.Transaction `json:"transactions"`
}

func (c *Client) GetCurrentBlock() int {
	url := fmt.Sprintf("%s/block/current", c.baseUrl)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", "go-client/v1")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Println("error getting current block:", err)
		return 0
	}

	defer res.Body.Close()

	var currentBlockResponse CurrentBlockResponse
	if err := json.NewDecoder(res.Body).Decode(&currentBlockResponse); err != nil {
		c.logger.Println("error decoding current block response:", err)
		return 0
	}
	return int(currentBlockResponse.BlockNumber)
}

func (c *Client) GetTransactions(address string) []api.Transaction {
	url := fmt.Sprintf("%s/transactions/%s", c.baseUrl, address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Println("error getting transactions for address:", address, err)
		return nil
	}

	req.Header.Add("User-Agent", "go-client/v1")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Println("error getting transactions for address:", address, err)
		return nil
	}

	defer res.Body.Close()

	var addressTransactionsResponse AddressTransactionsResponse
	if err := json.NewDecoder(res.Body).Decode(&addressTransactionsResponse); err != nil {
		c.logger.Println("error decoding transactions for address:", address, err)
		return nil
	}

	return addressTransactionsResponse.Transactions
}

func (c *Client) Subscribe(address string) bool {
	url := fmt.Sprintf("%s/subscribe/%s", c.baseUrl, address)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		c.logger.Println("error subscribing address:", address, err)
		return false
	}

	req.Header.Add("User-Agent", "go-client/v1")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Println("error subscribing address:", address, err)
		return false
	}

	return res.StatusCode == http.StatusAccepted
}
