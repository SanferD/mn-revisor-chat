package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/scrapers"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
	"code/infrastructure/types"
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	urlQueue       core.URLQueue
	rawEventsQueue core.RawEventsQueue
	logger         core.Logger
	rawStore       core.RawDataStore
	chunksStore    core.ChunksDataStore
	scraper        core.MNRevisorStatutesScraper
)

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
		logger.Fatal("error on initialize-sqs: %v", err)
	}

	rawEventsQueue, err = queues.InitializeSQSHelper(ctx, mySettings.RawEventsSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initialize-sqs: %v", err)
	}

	s3Helper, err := stores.InitializeS3Helper(ctx, mySettings.MainBucketName, mySettings.RawPathPrefix, mySettings.ChunkPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initializing s3-helper: %v", err)
	}
	rawStore = s3Helper
	chunksStore = s3Helper

	scraper, err = scrapers.InitializeScraper()
	if err != nil {
		logger.Fatal("error on initializing scraper: %v", err)
	}

}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	var err error
	for _, record := range sqsEvent.Records {
		logger.Info("processing", record)
		var event types.S3EventMessage
		json.Unmarshal([]byte(record.Body), &event)
		err = application.ScrapeRawPage(ctx, event.Detail.Object.Key, rawStore, chunksStore, urlQueue, scraper, logger)
		if err != nil {
			logger.Fatal("error on scraping raw page: %v", err)
		}
		if err = rawEventsQueue.DeleteMessageByHandle(ctx, record.ReceiptHandle); err != nil {
			logger.Fatal("error on deleting event from raw events queue: %v", err)
		}
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
