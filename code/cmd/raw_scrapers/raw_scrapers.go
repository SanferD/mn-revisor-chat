package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/scrapers"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
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
	statutesStore  core.StatutesDataStore
	scraper        core.MNRevisorStatutesScraper
)

type s3EventMessage struct {
	Detail s3Detail `json:"detail"`
}

type s3Detail struct {
	Object s3Object `json:"object"`
}

type s3Object struct {
	Key string `json:"key"`
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
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

	rawEventsQueue, err = queues.InitializeSQSHelper(ctx, mySettings.RawEventsSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initialize-sqs: %v", err)
	}

	s3Helper, err := stores.InitializeS3Helper(ctx, mySettings.BucketName, mySettings.RawPathPrefix, mySettings.ChunkPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initializing s3-helper: %v", err)
	}
	rawStore = s3Helper
	statutesStore = s3Helper

	scraper, err = scrapers.InitializeScraper()
	if err != nil {
		logger.Fatal("error on initializing scraper: %v", err)
	}

	for _, record := range sqsEvent.Records {
		logger.Info("processing", record)
		var event s3EventMessage
		json.Unmarshal([]byte(record.Body), &event)
		err = application.ScrapeRawPage(ctx, event.Detail.Object.Key, rawStore, statutesStore, urlQueue, scraper, logger)
		if err != nil {
			logger.Fatal("error on scraping raw page: %v", err)
		}
		rawEventsQueue.DeleteMessage(ctx, record.ReceiptHandle)
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
