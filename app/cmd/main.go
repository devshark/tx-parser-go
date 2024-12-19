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
	"github.com/devshark/tx-parser-go/pkg/env"
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

	txRepo := repository.NewInMemoryTransactionRepository()
	subRepo := repository.NewInMemorySubscriberRepository()
	blockRepo := repository.NewInMemoryBlockRepository()

	parser := worker.NewParserWorker(blockchainClient, txRepo, subRepo, blockRepo).WithCustomLogger(logger)

	router := httpHandler.NewRouter(blockchainClient, txRepo, subRepo, logger)
	server := httpHandler.NewHttpServer(router, config.port, httpReadTimeout, httpWriteTimeout)

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

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("http server failed to start: %v", err)
		}

		logger.Println("http server stopped")
	}()

	workerStopped := make(chan struct{})

	go func() {
		logger.Println("starting the parser worker")

		if err := parser.Run(ctx, config.jobSchedule); err != nil && !errors.Is(err, context.Canceled) {
			logger.Fatalf("failed to run parser: %v", err)
		}

		close(workerStopped)

		logger.Println("parser worker stopped")
	}()

	logger.Print("the app is running")

	<-stop

	log.Print("Shutting down...")
	// if Shutdown takes too long, cancel the context
	time.AfterFunc(shutdownTimeout, cancel)

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Shutdown", err)
	}

	<-workerStopped

	log.Print("Gracefully stopped.")
}

type Config struct {
	publicNodeURL string
	port          int64
	jobSchedule   time.Duration
}

func NewConfig() *Config {
	return &Config{
		publicNodeURL: env.GetEnv("PUBLIC_NODE_URL", "https://ethereum-rpc.publicnode.com/"),
		port:          env.GetEnvInt64("PORT", 8080),
		jobSchedule:   env.GetEnvDuration("JOB_SCHEDULE", 5*time.Second),
	}
}
