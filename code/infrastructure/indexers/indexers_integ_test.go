package indexers

import (
	"code/core"
	"code/infrastructure/loggers"
	"code/infrastructure/settings"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	vectorDocument core.VectorDocument
}

var tests []TestCase = []TestCase{
	{vectorDocument: core.VectorDocument11},
	{vectorDocument: core.VectorDocument12},
}

var expectedChunkIDs []string = []string{"1a.34.1", "1a.34.2a"}

func TestIndexers(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	mySettings, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	logger, err := loggers.InitializeMultiLogger(true)
	assert.NoError(err, "error on initializing multilogger: %v", err)
	osiHelper, err := InitializeOpenSearchIndexerHelper(ctx, mySettings.OpensearchUsername, mySettings.OpensearchPassword, mySettings.OpensearchDomain, mySettings.DoAllowOpensearchInsecure, mySettings.OpensearchIndexName, mySettings.ContextTimeout, logger)
	assert.NoError(err, "error on creating opensearch indexer helper: %v", err)

	t.Run("test can setup index", func(t *testing.T) {
		err = osiHelper.SetupIndexIfNecessary(ctx)
		assert.NoError(err, "error on setting up index if necessary: %v", err)
		err = osiHelper.SetupIndexIfNecessary(ctx)
		assert.NoError(err, "error on setting up index a second time if necessary: %v", err)
	})

	t.Run("test can AddVectorDocument", func(t *testing.T) {
		for _, test := range tests {
			err = osiHelper.AddVectorDocument(ctx, test.vectorDocument)
			assert.NoError(err, "error on adding statute: %v", err)
		}
	})

	t.Run("test can FindMatches", func(t *testing.T) {
		for _, test := range tests {
			chunkIDs, err := osiHelper.FindMatchingChunkIDs(ctx, test.vectorDocument)
			assert.NoError(err, "error on finding matches: %v", err)
			assert.Equal(expectedChunkIDs, chunkIDs, "did not find matching chunkids")
		}
	})

}
