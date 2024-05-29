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

func TestIndexers(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	mySettings, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	logger, _ := loggers.InitializeMultiLogger(true)
	osiHelper, err := InitializeOpenSearchIndexerHelper(ctx, mySettings.OpensearchUsername, mySettings.OpensearchPassword, mySettings.OpensearchDomain, mySettings.DoAllowOpensearchInsecure, mySettings.OpensearchIndexName, mySettings.ContextTimeout, logger)
	assert.NoError(err, "error on creating opensearch indexer helper: %v", err)

	t.Run("test can add document", func(t *testing.T) {
		for _, test := range tests {
			err = osiHelper.AddVectorDocument(ctx, test.vectorDocument)
			assert.NoError(err, "error on adding statute: %v", err)
		}
	})
}
