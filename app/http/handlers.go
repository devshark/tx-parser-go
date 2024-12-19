package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/devshark/tx-parser-go/app/internal/blockchain"
	"github.com/devshark/tx-parser-go/app/internal/repository"
	"github.com/devshark/tx-parser-go/client"
)

type httpHandler struct {
	bcClient        blockchain.BlockchainClient
	transactionRepo repository.TransactionRepository
	subscriberRepo  repository.SubscriberRepository
	logger          *log.Logger
}

func (h *httpHandler) GetCurrentBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	block, err := h.bcClient.GetLatestBlockNumber(ctx)
	if err != nil {
		h.logger.Printf("Failed to get latest block number: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := &client.CurrentBlockResponse{
		BlockNumber: block,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *httpHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := r.PathValue("address")

	if strings.TrimSpace(address) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	transactions, err := h.transactionRepo.GetTransactions(ctx, address)
	if err != nil {
		h.logger.Printf("Failed to get transactions for address %s: %v", address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx := &client.AddressTransactionsResponse{
		Transactions: transactions,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (h *httpHandler) PostSubscribeAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := r.PathValue("address")

	if strings.TrimSpace(address) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.subscriberRepo.Subscribe(ctx, address); err != nil {
		h.logger.Printf("Failed to subscribe address %s: %v", address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *httpHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
