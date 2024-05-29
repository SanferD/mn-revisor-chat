package helpers

import (
	"code/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	statute core.Statute
	chunks  []core.Chunk
}

type chunkTestCase struct {
	objectKey string
	chunkID   string
}

var tests = []testCase{
	{statute: core.TestStatute1, chunks: []core.Chunk{core.Chunk11, core.Chunk12}},
	{statute: core.TestStatute2, chunks: []core.Chunk{core.Chunk21}},
}

var chunkTests = []chunkTestCase{
	{objectKey: "bucket/chunk/1.2.3.txt", chunkID: "1.2.3"},
	{objectKey: "bucket/chunk/4.12a.txt", chunkID: "4.12a"},
}

func TestHelpers(t *testing.T) {
	t.Run("statutes 2 subdivision chunks", func(t *testing.T) {
		for _, test := range tests {
			chunks := Statute2SubdivisionChunks(test.statute)
			assert.Equal(t, test.chunks, chunks, "chunks are not the same")
		}
	})

	t.Run("ChunkObjectKeyToChunkID extracts chunk id successfully", func(t *testing.T) {
		for _, test := range chunkTests {
			chunkID := ChunkObjectKeyToID(test.objectKey)
			assert.Equal(t, test.chunkID, chunkID, "chunk ids not equal")
		}
	})
}
