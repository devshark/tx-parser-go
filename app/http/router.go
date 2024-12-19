package http

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/devshark/tx-parser-go/app/internal/blockchain"
	"github.com/devshark/tx-parser-go/app/internal/repository"
)

func NewRouter(
	bcClient blockchain.BlockchainClient,
	transactionRepo repository.TransactionRepository,
	subscriberRepo repository.SubscriberRepository,
	logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	handler := &httpHandler{
		bcClient:        bcClient,
		transactionRepo: transactionRepo,
		subscriberRepo:  subscriberRepo,
		logger:          logger,
	}

	mux.HandleFunc("GET /healthz", handler.HandleHealthCheck)
	mux.HandleFunc("GET /block/current", handler.GetCurrentBlock)
	mux.HandleFunc("GET /transactions/{address}", handler.GetTransactions)
	mux.HandleFunc("POST /subscribe/{address}", handler.PostSubscribeAddress)

	return mux
}

func NewHttpServer(httpHandlers http.Handler, port int64, httpReadTimeout, httpWriteTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           httpHandlers,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		ReadHeaderTimeout: 0,
		MaxHeaderBytes:    1 << 20,
	}
}
