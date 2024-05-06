package application

import (
	"bytes"
	"code/core"
	"context"
	"fmt"
	"math/rand"
	"time"
)

const sleepSecondsDelta = 2
const sleepSecondsMin = 3

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

func Crawl(ctx context.Context, urlQueue core.URLQueue, seenURLStore core.SeenURLStore, dataStore core.DataStore, webClient core.WebClient, interruptWatcher core.InterruptWatcher, logger core.Logger) error {
	for !interruptWatcher.IsInterrupted() {
		var err error

		// sleep for politeness
		sleepSeconds := sleepSecondsMin + rand.Float32()*sleepSecondsDelta
		sleepDuration := time.Second * time.Duration(sleepSeconds)
		logger.Info("sleeping for %v", sleepDuration)
		time.Sleep(sleepDuration)

		// get URL
		logger.Info("receiving next url from queue")
		urlQueueMessage, err := urlQueue.ReceiveURLQueueMessage(ctx)
		if err != nil {
			logger.Error("error on receiveURL: %v", err)
			continue
		}
		if urlQueueMessage == nil {
			logger.Info("queue is empty")
			continue
		}
		url := urlQueueMessage.Body
		logger.Info("received next URL='%s'", url)

		// test that url isn't seen
		logger.Info("testing if URL='%s' is seen", url)
		hasURL, err := seenURLStore.HasURL(ctx, url)
		if err != nil {
			logger.Error("error on testing if URL is seen: %v", err)
			continue
		}
		if hasURL {
			// url is seen, delete the message
			logger.Info("URL='%s' is seen, skipping...", url)
			if err = urlQueue.DeleteURLQueueMessage(ctx, urlQueueMessage); err != nil {
				logger.Error("error on deleting queue message: %v", err)
			}
			continue
		}

		// get Web page
		logger.Info("getting HTML for url='%s'", url)
		htmlPageBytes, err := webClient.GetHTML(ctx, url)
		if err != nil {
			logger.Error("error on GetHTML for url='%s': %v", url, err)
			continue
		}
		htmlPageReader := bytes.NewReader(htmlPageBytes)

		// save web page to store
		logger.Info("saving HTML for url='%s'", url)
		fileName := getURLFileName(url)
		if err = dataStore.PutTextFile(ctx, fileName, htmlPageReader); err != nil {
			logger.Error("error on PutTextFile for url='%s': %v", url, err)
			continue
		}

		// update seen urls
		logger.Info("putting url='%s' as seen", url)
		if err = seenURLStore.PutURL(ctx, url); err != nil {
			logger.Error("error on PutURL for url='%s': %v", url, err)
			continue
		}

		// delete url from queue
		logger.Info("deleting url='%s' from queue", url)
		if err = urlQueue.DeleteURLQueueMessage(ctx, urlQueueMessage); err != nil {
			logger.Error("error on DeleteQueueMessage: %v", err)
			continue
		}

		// success
		logger.Info("url='%s' crawl done", url)
	}
	return nil
}
