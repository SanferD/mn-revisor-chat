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

var tests = []testCase{
	{statute: core.TestStatute1, chunks: []core.Chunk{core.Chunk11, core.Chunk12}},
	{statute: core.TestStatute2, chunks: []core.Chunk{core.Chunk21}},
}

func TestHelpers(t *testing.T) {
	t.Run("statutes 2 subdivision chunks", func(t *testing.T) {
		for _, test := range tests {
			chunks := Statute2SubdivisionChunks(test.statute)
			assert.Equal(t, test.chunks, chunks, "chunks are not the same")
		}
	})
}
