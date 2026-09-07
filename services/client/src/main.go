package main

import (
	"errors"
	"os"
	"os/signal"
	"context"
	"syscall"

	client "github.com/7574-sistemas-distribuidos/tp-nivelador/src/client"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

func loadConfig() (client.ClientConfig, error) {
	agencyId := os.Getenv("AGENCY_ID")
	if agencyId == "" {
		return client.ClientConfig{}, errors.New("AGENCY_ID environment variable is required")
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		return client.ClientConfig{}, errors.New("SERVER_HOST environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return client.ClientConfig{}, errors.New("SERVER_PORT environment variable is required")
	}

	inputFile := os.Getenv("INPUT_FILE")
	if inputFile == "" {
		return client.ClientConfig{}, errors.New("INPUT_FILE environment variable is required")
	}

	outputFile := os.Getenv("OUTPUT_FILE")
	if outputFile == "" {
		return client.ClientConfig{}, errors.New("OUTPUT_FILE environment variable is required")
	}

	var batchSize int
	batchSizeStr := os.Getenv("BATCH_SIZE")
	if batchSizeStr != "" {
		batchSize = 0
		// Implementacion manual de atoi 
		for _, char := range batchSizeStr {
			batchSize *= 10 
			if char < '0' || char > '9' {
				return client.ClientConfig{}, errors.New("BATCH_SIZE should be numeric")
			}
			batchSize += int(char - '0')
		}

		if batchSize == 0 {
			return client.ClientConfig{}, errors.New("BATCH_SIZE should be positive")
		}
	} else {
		batchSize = 1
	}

	return client.ClientConfig{
		ServerHost: serverHost,
		ServerPort: serverPort,
		AgencyId:   agencyId,
		InputFile:  inputFile,
		OutputFile: outputFile,
		BatchSize:  batchSize,
	}, nil
}

func run() int {	
	config, err := loadConfig()
	if err != nil {
		logger.Error("load-config", logger.Fail, "err", err)
		return 1
	}
	
	client, err := client.NewClient(config)
	if err != nil {
		logger.Error("client-new", logger.Fail, "err", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	errorChannel := make(chan error)
	go func() {
		errorChannel<-client.Run()	
	}()
	
	select {
		case err := <-errorChannel:
			if err != nil && !client.IsStop() {
				logger.Error("client-run", logger.Fail, "err", err)
				return 1
			}
			return 0

		case <-ctx.Done():
			logger.Info("SIGTERM-client", logger.InProgress)
			client.Stop() 

			err := <-errorChannel
			if err != nil && !client.IsStop() {
				logger.Error("client-run", logger.Fail, "err", err)
				return 1
			}
			logger.Info("SIGTERM-client", logger.Success)
			return 0
	}
}

func main() {
	os.Exit(run())
}
