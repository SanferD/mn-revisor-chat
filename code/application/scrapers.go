package application

import (
	"code/core"
	"code/helpers"
	"context"
	"fmt"
	"strings"
)

func ScrapeRawPage(ctx context.Context, objectKey string, rawDataStore core.RawDataStore, chunksDataStore core.ChunksDataStore, urlQueue core.URLQueue, scraper core.MNRevisorStatutesScraper, logger core.Logger) error {

	// get text file
	logger.Info("getting text file \"%s\"", objectKey)
	contents, err := rawDataStore.GetTextFile(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("error getting text file from s3: %v", err)
	}

	// get page kind
	logger.Info("getting page kind")
	pageKind, err := scraper.GetPageKind(strings.NewReader(contents))
	if err != nil {
		return fmt.Errorf("error getting page kind while parsing: %v", err)
	}

	doDelete := true

	// parse page
	switch pageKind {
	case core.StatutesChaptersTable, core.StatutesChaptersShortTable, core.StatutesSectionsTable:
		// extract urls
		logger.Info("found page kind %v, extracting urls", pageKind)
		urls, err := scraper.ExtractURLs(strings.NewReader(contents), pageKind)
		if err != nil {
			return fmt.Errorf("error extracting urls from statutes chatpers table: %v", err)
		}

		// put scraped urls in the url queue for crawling
		for _, url := range urls {
			logger.Info("sending url \"%s\"", url)
			if err := urlQueue.SendURL(ctx, url); err != nil {
				logger.Error("error putting url: %v", err)
				doDelete = false
				continue
			}
		}
	case core.Statutes:
		// extract statute
		logger.Info("found page kind %v, extracting statutes", pageKind)
		statute, err := scraper.ExtractStatute(strings.NewReader(contents))
		if err != nil {
			return fmt.Errorf("error on extracting statutes: %v", err)
		}
		if len(statute.Title) == 0 {
			logger.Info("statute is empty, skipping...")
		}

		// put subdivision chunks into data store
		for _, chunk := range helpers.Statute2SubdivisionChunks(statute) {
			logger.Info("putting chunk in data store, chunk=%s", chunk)
			if err := chunksDataStore.PutChunk(ctx, chunk); err != nil {
				return fmt.Errorf("error on putting chunk into chunk store: %v", err)
			}
		}
	default:
		return fmt.Errorf("unsupported page kind: %v", pageKind)
	}

	// delete the scraped page if no error
	if doDelete {
		logger.Info("deleting raw text file \"%s\"", objectKey)
		if err := rawDataStore.DeleteTextFile(ctx, objectKey); err != nil {
			return fmt.Errorf("error deleting text file: %v", err)
		}
		logger.Info("scraping %s done", objectKey)
	} else {
		logger.Info("skipping deleting raw text file \"%s\", some errors found", objectKey)
		return fmt.Errorf("scraping '%s' completed with errors", objectKey) // sqs will not delete message
	}

	// success
	return nil
}
