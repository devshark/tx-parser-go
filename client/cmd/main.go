package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devshark/tx-parser-go/client"
)

var sampleAddresses = []string{
	"0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5",
	"0x6eaE5e2d47f1CbB5979734812521579921d37C9A",
}

func main() {
	logger := log.Default()

	parserClient := client.NewClient("http://localhost:8080")

	currentBlock := parserClient.GetCurrentBlock()
	logger.Printf("current block: %d", currentBlock)

	for _, address := range sampleAddresses {
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
		case <-time.After(5 * time.Second):
			for _, address := range sampleAddresses {
				transactions := parserClient.GetTransactions(address)
				logger.Printf("%d transactions for %s\n", len(transactions), address)
				// logger.Printf("%d transactions for %s: %+v\n", len(transactions), address, transactions)
			}
		}
	}

}
