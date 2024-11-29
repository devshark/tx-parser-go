package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandler "github.com/devshark/tx-parser-go/app/http"
	"github.com/devshark/tx-parser-go/app/internal/blockchain"
	"github.com/devshark/tx-parser-go/app/internal/repository"
	"github.com/devshark/tx-parser-go/app/worker"
)

const (
	publicNodeURL    = "https://ethereum-rpc.publicnode.com/"
	shutdownTimeout  = 5 * time.Second
	httpReadTimeout  = 5 * time.Second
	httpWriteTimeout = 10 * time.Second
	port             = "8080"
	startBlock       = 21292394
	jobSchedule      = time.Duration(5 * time.Second)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.Default()

	blockchainClient := blockchain.NewPublicNodeClient(publicNodeURL, logger)

	repository := repository.NewInMemoryRepository()

	parser := worker.NewParserWorker(blockchainClient, repository, logger)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpHandler.NewServeMux(blockchainClient, repository, logger),
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		ReadHeaderTimeout: 0,
		MaxHeaderBytes:    1 << 20,
	}

	stop := make(chan os.Signal, 1)

	go func() {
		logger.Printf("listening on port %s", port)

		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("http server closed: %v", err)
		}
	}()

	go func() {
		logger.Println("starting the parser worker")

		if err := parser.Run(ctx, startBlock, jobSchedule); err != nil {
			logger.Fatalf("failed to run parser: %v", err)
		}

		logger.Println("parser worker stopped")
	}()

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	logger.Print("the app is running")

	<-stop

	log.Print("Shutting down...")
	// if Shutdown takes longer than 10, cancel the context
	time.AfterFunc(shutdownTimeout, cancel)

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Shutdown", err)
	}

	log.Print("Gracefully stopped.")
}
