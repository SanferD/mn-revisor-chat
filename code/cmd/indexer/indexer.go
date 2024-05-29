package main

import (
	"code/application"
	"code/core"
	"code/helpers"
	"code/infrastructure/indexers"
	"code/infrastructure/loggers"
	"code/infrastructure/queues"
	"code/infrastructure/settings"
	"code/infrastructure/stores"
	"code/infrastructure/types"
	"code/infrastructure/vectorizers"
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	var err error

	ctx := context.Background()

	log.Println("initializing settings")
	mySettings, err := settings.GetSettings()
	if err != nil {
		log.Fatalf("error getting settings: %v\n", err)
	}

	log.Println("initializing multi logger")
	if logger, err = loggers.InitializeMultiLogger(mySettings.DoLogToStdout); err != nil {
		log.Fatalf("error initializing logger: %v\n", err)
	}

	logger.Info("initializing s3 helper")
	if chunksDataStore, err = stores.InitializeS3Helper(ctx, mySettings.MainBucketName, mySettings.RawPathPrefix, mySettings.ChunkPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint); err != nil {
		logger.Fatal("error initializing s3 helper: %v", err)
	}

	logger.Info("initializing sqs helper")
	if toIndexQueue, err = queues.InitializeSQSHelper(ctx, mySettings.ToIndexSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint); err != nil {
		logger.Fatal("error initializing url queue: %v", err)
	}

	logger.Info("initializing bedrock helper")
	if vectorizer, err = vectorizers.InitializeBedrockHelper(ctx, mySettings.EmbeddingModelID, mySettings.ContextTimeout); err != nil {
		logger.Fatal("error initializing bedrock helper: %v", err)
	}

	logger.Info("initializing opensearch helper")
	if searchIndex, err = indexers.InitializeOpenSearchIndexerHelper(ctx, mySettings.OpensearchUsername, mySettings.OpensearchPassword, mySettings.OpensearchDomain, mySettings.DoAllowOpensearchInsecure, mySettings.OpensearchIndexName, mySettings.ContextTimeout, logger); err != nil {
		logger.Fatal("error initializing opensearch indexer helper: %v", err)
	}
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		logger.Info("processing", record)
		var event types.S3EventMessage
		json.Unmarshal([]byte(record.Body), &event)
		chunkID := helpers.ChunkObjectKeyToID(event.Detail.Object.Key)
		if err := application.Index(ctx, chunkID, chunksDataStore, vectorizer, searchIndex, logger); err != nil {
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
