package application

import (
	"code/core"
	"context"
	"fmt"
)

func InvokeTriggerCrawler(ctx context.Context, triggerCrawlerInvoker core.Invoker, logger core.Logger) error {
	logger.Info("determining if trigger crawler is already running")
	isRunning, err := triggerCrawlerInvoker.IsTriggerCrawlerAlreadyRunning(ctx)
	if err != nil {
		return fmt.Errorf("error on checking if trigger crawler is already running: %v", err)
	}
	if isRunning {
		logger.Info("trigger crawler is already running, done...")
		return nil
	}

	logger.Info("invoking trigger crawler")
	if err := triggerCrawlerInvoker.InvokeTriggerCrawler(ctx); err != nil {
		return fmt.Errorf("error on invoking trigger crawler: %v", err)
	}
	logger.Info("invoke trigger crawler success")
	return nil
}
