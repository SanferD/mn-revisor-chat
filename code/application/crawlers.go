package application

import (
	"code/core"
	"context"
	"fmt"
)

func TriggerCrawler(ctx context.Context, urlQueue core.URLQueue, seenURLStore core.SeenURLStore, logger core.Logger) error {
	logger.Info("purging url queue")
	if err := urlQueue.Clear(ctx); err != nil {
		return fmt.Errorf("error on queue Purge: %v", err)
	}
	logger.Info("deleting all seen url items")
	if err := seenURLStore.DeleteAll(ctx); err != nil {
		return fmt.Errorf("error on table1 DeleteAll: %v", err)
	}
	logger.Info("sending '%s' to url queue", MNRevisorStatutesURL)
	if err := urlQueue.SendURL(ctx, MNRevisorStatutesURL); err != nil {
		return fmt.Errorf("error on queue SendBody: %v", err)
	}
	logger.Info("trigger success")
	return nil
}
