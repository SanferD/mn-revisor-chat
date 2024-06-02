package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/comms"
	"code/infrastructure/indexers"
	"code/infrastructure/loggers"
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
	agent      core.Agent
	chunkStore core.ChunksDataStore
	indexer    core.SearchIndex
	logger     core.Logger
	vectorizer core.Vectorizer
	comm       core.Comms
)

var internalErrorResponse = events.APIGatewayProxyResponse{
	StatusCode: 500,
	Body:       "internal error",
}

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

	logger.Info("initializing comms helpers")
	comm, err = comms.InitializeSinchHelper(ctx, mySettings.SinchAPIToken, mySettings.SinchProjectID, mySettings.SinchVirtualPhoneNumber, mySettings.ContextTimeout)
	if err != nil {
		logger.Fatal("error on initializing sinch helper: %v", err)
	}

}

func HandleRequest(ctx context.Context, payload events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var err error
	logger.Info("processing payload %+v", payload)
	var whp types.SinchWebhookPayload
	if err := json.Unmarshal([]byte(payload.Body), &whp); err != nil {
		err = fmt.Errorf("error on unmarshalling json: %v", err)
		return internalErrorResponse, err
	}
	prompt := whp.Message.ContactMessage.TextMessage.Text
	phoneNumber := whp.Message.ChannelIdentity.Identity
	if err = application.Answer(ctx, prompt, phoneNumber, chunkStore, agent, indexer, vectorizer, comm, logger); err != nil {
		err = fmt.Errorf("error on getting answer from application: %v", err)
		return internalErrorResponse, err
	}
	logger.Info("done processing payload %+v", payload)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "success!",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
