package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devshark/tx-parser-go/client"
	"github.com/devshark/tx-parser-go/pkg/env"
)

func main() {
	config := NewConfig()
	logger := log.Default()

	parserClient := client.NewClient(config.parserUrl)

	currentBlock := parserClient.GetCurrentBlock()
	logger.Printf("current block: %d", currentBlock)

	for _, address := range config.subscribeAddresses {
		parserClient.Subscribe(address)
		logger.Printf("subscribed to %s", address)
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

	// Run forever until we kill the program
	for {
		select {
		case <-stop:
			return
		case <-time.After(config.fetchFrequency):
			for _, address := range config.subscribeAddresses {
				transactions := parserClient.GetTransactions(address)
				logger.Printf("%d transactions for %s\n", len(transactions), address)
				// logger.Printf("%d transactions for %s: %+v\n", len(transactions), address, transactions)
			}
		}
	}

}

type Config struct {
	parserUrl          string
	fetchFrequency     time.Duration
	subscribeAddresses []string
}

func NewConfig() *Config {
	return &Config{
		parserUrl:          env.GetEnv("PARSER_URL", "http://localhost:8081"),
		fetchFrequency:     env.GetEnvDuration("FETCH_FREQUENCY", 5*time.Second),
		subscribeAddresses: env.GetEnvValues("SUBSCRIBE_ADDRESSES"),
	}
}
