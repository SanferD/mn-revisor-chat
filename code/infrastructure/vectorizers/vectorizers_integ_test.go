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

type vectorizeTestCase struct {
	body           string
	vectorDocument core.VectorDocument
}

var testCases = []test{
	{chunk: core.Chunk11, vectorDocument: core.VectorDocument11},
	{chunk: core.Chunk12, vectorDocument: core.VectorDocument12},
}

var vectorizeTest = vectorizeTestCase{
	body:           core.Prompt,
	vectorDocument: core.PromptVD,
}

const dcPrompt = "I want to open a dry cleaning business. By when should I register my business with the commissioner of revenue?"
const expectedSubdivision = "§ 115B.49, subd. 4"

func TestVectorizers(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	mySettings, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	bedrockHelper, err := InitializeBedrockHelper(ctx, mySettings.EmbeddingModelID, mySettings.FoundationModelID, mySettings.ContextTimeout)
	assert.NoError(err, "error on initialize bedrock helper: %v", err)

	t.Run("test vectorize chunk", func(t *testing.T) {
		for _, testCase := range testCases {
			vectorDocument, err := bedrockHelper.VectorizeChunk(ctx, testCase.chunk)
			assert.NoError(err, "error on vectorize chunk: %v", err)
			assert.Equal(testCase.vectorDocument, vectorDocument, "vector documents are not equal")
		}
	})

	t.Run("test vectorize", func(t *testing.T) {
		vd, err := bedrockHelper.Vectorize(ctx, vectorizeTest.body)
		assert.NoError(err, "error on vectorize: %v", err)
		assert.Equal(vd, vectorizeTest.vectorDocument, "vector documents are not equal")
	})

	t.Run("test AskWithChunks", func(t *testing.T) {
		answer, err := bedrockHelper.AskWithChunks(ctx, dcPrompt, core.DCChunks)
		assert.NoError(err, "found error on ask with chunks: %v", err)
		assert.Contains(answer, expectedSubdivision)
	})
}
