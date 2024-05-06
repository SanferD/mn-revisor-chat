package trigger_scraper

import (
	"code/application"
	"code/core"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
	"context"
	"log"

	_ "github.com/aws/aws-lambda-go/lambda"
)

var (
	urlQueue     core.URLQueue
	seenURLStore core.SeenURLStore
	logger       core.Logger
)

func HandleRequest(ctx context.Context) error {
	mySettings, err := settings.GetSettings()
	if err != nil {
		log.Fatalf("error on GetSettings: %v", err)
	}

	logger, err = loggers.InitializeMultiLogger(mySettings.DoLogToStdout)
	if err != nil {
		log.Fatalf("error on initialize multilogger: %v\n", err)
	}

	urlQueue, err = queues.InitializeURLSQSHelper(ctx, mySettings.URLSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initialize-sqs: %v", err)
	}

	table1, err := stores.InitializeTable1(ctx, mySettings.Table1ARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initialize-table1: %v", err)
	}
	seenURLStore = table1

	err = application.TriggerCrawler(ctx, urlQueue, table1, logger)
	if err != nil {
		logger.Fatal("error on trigger-crawler: %v", err)
	}
	return nil
}

func main() {
	// lambda.Start(HandleRequest)
	HandleRequest(context.Background())
}
