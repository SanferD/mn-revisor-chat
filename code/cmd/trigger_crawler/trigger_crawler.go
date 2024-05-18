package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	logger         core.Logger
	seenURLStore   core.SeenURLStore
	urlQueue       core.URLQueue
	rawEventsQueue core.RawEventsQueue
)

type URLsEvent struct {
	Urls []string `json:"urls"`
}

func init() {
	ctx := context.Background()

	mySettings, err := settings.GetSettings()
	if err != nil {
		log.Fatalf("error on GetSettings: %v", err)
	}

	logger, err = loggers.InitializeMultiLogger(mySettings.DoLogToStdout)
	if err != nil {
		log.Fatalf("error on initialize multilogger: %v\n", err)
	}

	urlQueue, err = queues.InitializeSQSHelper(ctx, mySettings.URLSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on url initialize-sqs: %v", err)
	}

	rawEventsQueue, err = queues.InitializeSQSHelper(ctx, mySettings.RawEventsSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on raw-events initialize-sqs: %v", err)
	}

	seenURLStore, err = stores.InitializeTable1(ctx, mySettings.Table1ARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initialize-table1: %v", err)
	}

}

func HandleRequest(ctx context.Context, urlsEvent URLsEvent) error {
	logger.Info("Received event", urlsEvent)
	if err := application.TriggerCrawler(ctx, urlsEvent.Urls, urlQueue, rawEventsQueue, seenURLStore, logger); err != nil {
		logger.Fatal("error on trigger-crawler: %v", err)
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
