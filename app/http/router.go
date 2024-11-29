package http

import (
	"log"
	"net/http"

	"github.com/devshark/tx-parser-go/app/internal/blockchain"
	"github.com/devshark/tx-parser-go/app/internal/repository"
)

func NewServeMux(bcClient blockchain.BlockchainClient, repo repository.Repository, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	handler := &httpHandler{
		bcClient: bcClient,
		repo:     repo,
		logger:   logger,
	}

	mux.HandleFunc("GET /healthz", handler.HandleHealthCheck)
	mux.HandleFunc("GET /block/current", handler.GetCurrentBlock)
	mux.HandleFunc("GET /transactions/{address}", handler.GetTransactions)
	mux.HandleFunc("POST /subscribe/{address}", handler.PostSubscribeAddress)

	return mux
}
