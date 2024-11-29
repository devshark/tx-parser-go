package main

import (
	"context"
	"errors"
	"fmt"
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
	shutdownTimeout  = 5 * time.Second
	httpReadTimeout  = 5 * time.Second
	httpWriteTimeout = 10 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := NewConfig()
	logger := log.Default()

	blockchainClient := blockchain.NewPublicNodeClient(config.publicNodeURL, logger)

	repository := repository.NewInMemoryRepository()

	parser := worker.NewParserWorker(blockchainClient, repository, logger)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.port),
		Handler:           httpHandler.NewServeMux(blockchainClient, repository, logger),
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		ReadHeaderTimeout: 0,
		MaxHeaderBytes:    1 << 20,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		logger.Printf("listening on port %d", config.port)

		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("http server closed: %v", err)
		}
	}()

	go func() {
		logger.Println("starting the parser worker")

		if err := parser.Run(ctx, int64(config.startBlock), config.jobSchedule); err != nil {
			logger.Fatalf("failed to run parser: %v", err)
		}

		logger.Println("parser worker stopped")
	}()

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

type Config struct {
	publicNodeURL string
	port          int
	startBlock    int
	jobSchedule   time.Duration
}

func NewConfig() *Config {
	return &Config{
		publicNodeURL: "https://ethereum-rpc.publicnode.com/",
		port:          8080,
		startBlock:    21292394,
		jobSchedule:   time.Duration(5 * time.Second),
	}
}
