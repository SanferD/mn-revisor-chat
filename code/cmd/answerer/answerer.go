package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/indexers"
	"code/infrastructure/loggers"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
	"code/infrastructure/vectorizers"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	agent      core.Agent
	chunkStore core.ChunksDataStore
	indexer    core.SearchIndex
	logger     core.Logger
	vectorizer core.Vectorizer
)

func init() {
	ctx := context.Background()

	var err error

	log.Println("initializing settings")
	mySettings, err := settings.GetSettings()
	if err != nil {
		log.Fatalf("error on get settings: %v\n", err)
	}

	log.Println("initializing loggers")
	logger, err = loggers.InitializeMultiLogger(mySettings.DoLogToStdout)
	if err != nil {
		log.Fatalf("error on initializing multilogger: %v\n", err)
	}

	logger.Info("initializing bedrock helpers")
	var bedrockHelper *vectorizers.BedrockHelper
	bedrockHelper, err = vectorizers.InitializeBedrockHelper(ctx, mySettings.EmbeddingModelID, mySettings.FoundationModelID, mySettings.ContextTimeout)
	if err != nil {
		logger.Fatal("error on initializing bedrock helper: %v", err)
	}
	vectorizer = bedrockHelper
	agent = bedrockHelper

	logger.Info("initializing s3 helpers")
	chunkStore, err = stores.InitializeS3Helper(ctx, mySettings.MainBucketName, mySettings.RawPathPrefix, mySettings.ChunkPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initializing s3 helpers: %v", err)
	}

	logger.Info("initializing opensearch helpers")
	indexer, err = indexers.InitializeOpenSearchIndexerHelper(ctx, mySettings.OpensearchUsername, mySettings.OpensearchPassword, mySettings.OpensearchDomain, mySettings.DoAllowOpensearchInsecure, mySettings.OpensearchIndexName, mySettings.ContextTimeout, logger)
	if err != nil {
		logger.Fatal("error on initializing opensearch indexer helper: %v", err)
	}

}

func HandleRequest(ctx context.Context, snsEvent events.SNSEvent) error {
	var err error
	var prompt string
	logger.Info("processing snsEvent", snsEvent)
	for _, record := range snsEvent.Records {
		prompt = record.SNS.Message
		if err = application.Answer(ctx, prompt, chunkStore, agent, indexer, vectorizer, logger); err != nil {
			return err
		}
	}
	logger.Info("done processing snsEvent", snsEvent)
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
