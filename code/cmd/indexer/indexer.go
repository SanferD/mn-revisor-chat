package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/indexers"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/settings"
	"code/infrastructure/types"
	"code/infrastructure/vectorizers"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	logger          core.Logger
	chunksDataStore core.ChunksDataStore
	vectorizer      core.Vectorizer
	searchIndex     core.SearchIndex
	toIndexQueue    core.Queue
)

func init() {
	ctx := context.Background()

	var err error
	mySettings, err := settings.GetSettings()
	if err != nil {
		log.Fatalf("error getting settings: %v\n", err)
	}
	if logger, err = loggers.InitializeMultiLogger(mySettings.DoLogToStdout); err != nil {
		log.Fatalf("error initializing logger: %v\n", err)
	}
	if toIndexQueue, err = queues.InitializeSQSHelper(ctx, mySettings.ToIndexSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint); err != nil {
		logger.Fatal("error initializing url queue: %v", err)
	}
	if vectorizer, err = vectorizers.InitializeBedrockHelper(ctx, mySettings.EmbeddingModelID, mySettings.ContextTimeout); err != nil {
		logger.Fatal("error initializing bedrock helper: %v", err)
	}
	if searchIndex, err = indexers.InitializeOpenSearchIndexerHelper(ctx, mySettings.OpensearchUsername, mySettings.OpensearchPassword, mySettings.OpensearchDomain, mySettings.DoAllowOpensearchInsecure, mySettings.OpensearchIndexName, mySettings.ContextTimeout); err != nil {
		logger.Fatal("error initializing opensearch indexer helper: %v", err)
	}
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		logger.Info("processing", record)
		var event types.S3EventMessage
		json.Unmarshal([]byte(record.Body), &event)
		chunkObjectKey := event.Detail.Object.Key
		chunkFileNameParts := strings.Split(chunkObjectKey, "/")
		chunkFileName := chunkFileNameParts[len(chunkFileNameParts)-1]
		if err := application.Index(ctx, chunkFileName, chunksDataStore, vectorizer, searchIndex, logger); err != nil {
			return err
		}
		logger.Info("deleing message by handle=%s", record.ReceiptHandle)
		if err := toIndexQueue.DeleteMessageByHandle(ctx, record.ReceiptHandle); err != nil {
			return fmt.Errorf("error deleting message by handle: %v", err)
		}
		logger.Info("done processing record", record)
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
