package vectorizers

import (
	"code/core"
	"code/infrastructure/settings"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	chunk          core.Chunk
	vectorDocument core.VectorDocument
}

var testCases = []test{
	{chunk: core.Chunk11, vectorDocument: core.VectorDocument11},
	{chunk: core.Chunk12, vectorDocument: core.VectorDocument12},
}

func TestVectorizers(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	mySettings, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	bedrockHelper, err := InitializeBedrockHelper(ctx, mySettings.EmbddingModelID, mySettings.ContextTimeout)
	assert.NoError(err, "error on initialize bedrock helper: %v", err)

	t.Run("test vectorize chunk", func(t *testing.T) {
		for _, testCase := range testCases {
			vectorDocument, err := bedrockHelper.VectorizeChunk(ctx, testCase.chunk)
			assert.NoError(err, "error on vectorize chunk: %v", err)
			assert.Equal(testCase.vectorDocument, vectorDocument, "vector documents are not equal")
		}
	})
}
