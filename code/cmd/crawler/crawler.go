package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/clients"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
	"code/infrastructure/watchers"
	"context"
	"log"
)

var (
	urlQueue     core.URLQueue
	logger       core.Logger
	rawDataStore core.RawDataStore
	webClient    core.WebClient
)

const rawDataStorePathPrefix = "raw"

func main() {
	Crawl()
}

func Crawl() error {
	ctx := context.Background()

	var err error
	mySettings, err := settings.GetSettings()
	if err != nil {
		log.Fatalf("error getting settings: %v\n", err)
	}
	if logger, err = loggers.InitializeMultiLogger(mySettings.DoLogToStdout); err != nil {
		log.Fatalf("error initializing logger: %v\n", err)
	}
	log.Println(mySettings)
	table1, err := stores.InitializeTable1(ctx, mySettings.Table1ARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error initializing table1: %v", err)
	}
	if urlQueue, err = queues.InitializeURLSQSHelper(ctx, mySettings.URLSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint); err != nil {
		logger.Fatal("error initializing url queue: %v", err)
	}
	rawDataStore, err := stores.InitializeS3Helper(ctx, mySettings.BucketName, mySettings.RawPathPrefix, mySettings.ChunkPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error initializing s3: %v", err)
	}
	bgInterruptWatcher := watchers.InitializeBackgroundInterruptWatcher()
	webClient, err = clients.InitializeHTTPClientHelper()
	if err != nil {
		logger.Fatal("error initializing client: %v", err)
	}
	if err := application.Crawl(ctx, urlQueue, table1, rawDataStore, webClient, bgInterruptWatcher, logger); err != nil {
		logger.Fatal("error on crawl: %v", err)
	}
	return nil
}
