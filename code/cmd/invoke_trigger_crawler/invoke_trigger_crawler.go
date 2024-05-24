package main

import (
	"code/application"
	"code/core"
	"code/infrastructure/loggers"
	"code/infrastructure/settings"
	"code/infrastructure/tasks"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	logger               core.Logger
	invokeTriggerCrawler core.Invoker
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

	invokeTriggerCrawler, err = tasks.InitializeECSHelper(ctx, mySettings.TriggerCrawlerTaskDfnArn, mySettings.TriggerCrawlerClusterArn, mySettings.SubnetIds, mySettings.SecurityGroupIds, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	if err != nil {
		logger.Fatal("error on initialize-ecs: %v", err)
	}

}

func HandleRequest(ctx context.Context) error {
	if err := application.InvokeTriggerCrawler(ctx, invokeTriggerCrawler, logger); err != nil {
		return err
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
