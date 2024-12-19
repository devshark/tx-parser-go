package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/devshark/tx-parser-go/api"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl string
	logger  *log.Logger
	client  Doer
}

func NewClient(baseUrl string) *Client {
	return &Client{
		baseUrl: baseUrl,
		logger:  log.Default(),
		client:  http.DefaultClient,
	}
}

func (c *Client) WithCustomHttpDoer(client Doer) *Client {
	c.client = client

	return c
}

type CurrentBlockResponse struct {
	BlockNumber int64 `json:"block_number"`
}

type AddressTransactionsResponse struct {
	Transactions []api.Transaction `json:"transactions"`
}

func (c *Client) GetCurrentBlock() int {
	url := fmt.Sprintf("%s/block/current", c.baseUrl)

	var currentBlockResponse CurrentBlockResponse

	err := c.get(url, &currentBlockResponse)
	if err != nil {
		c.logger.Printf("error: %v\n", err)

		return 0
	}

	return int(currentBlockResponse.BlockNumber)
}

func (c *Client) GetTransactions(address string) []api.Transaction {
	url := fmt.Sprintf("%s/transactions/%s", c.baseUrl, address)

	var addressTransactionsResponse AddressTransactionsResponse

	err := c.get(url, &addressTransactionsResponse)
	if err != nil {
		c.logger.Printf("error: %v\n", err)

		return nil
	}

	return addressTransactionsResponse.Transactions
}

func (c *Client) Subscribe(address string) bool {
	url := fmt.Sprintf("%s/subscribe/%s", c.baseUrl, address)

	if err := c.postNoContent(url, nil, http.StatusAccepted); err != nil {
		c.logger.Printf("error subscribing address: %v\n", err)

		return false
	}

	return true
}

func (c *Client) get(url string, response any) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	req.Header.Add("User-Agent", "go-client/v1")

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return fmt.Errorf("json Decode: %w", err)
	}

	return nil
}

func (c *Client) postNoContent(url string, body io.Reader, expectedStatus int) error {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("NewRequest: %w", err)
	}

	req.Header.Add("User-Agent", "go-client/v1")

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	if res.StatusCode != expectedStatus {
		return fmt.Errorf("status code: expected %d, got %d", expectedStatus, res.StatusCode)
	}

	return nil
}
