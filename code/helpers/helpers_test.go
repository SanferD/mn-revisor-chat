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
var isLocalhostURLTestCases = []struct {
	url      string
	expected bool
}{
	{url: "http://localhost", expected: true},
	{url: "https://localhost", expected: true},
	{url: "http://127.0.0.1", expected: true},
	{url: "https://127.0.0.1", expected: true},
	{url: "http://[::1]", expected: true},
	{url: "https://[::1]", expected: true},
	{url: "http://localhost:8080", expected: true},
	{url: "https://localhost:8443", expected: true},
	{url: "http://example.com", expected: false},
	{url: "https://example.com", expected: false},
}

var base64EncodeTestCases = []struct {
	content  string
	expected string
}{
	{content: "Hello, World!", expected: "SGVsbG8sIFdvcmxkIQ=="},
	{content: "12345", expected: "MTIzNDU="},
	{content: "Lorem ipsum dolor sit amet", expected: "TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQ="},
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

	t.Run("IsLocalhostURL", func(t *testing.T) {

		for _, tc := range isLocalhostURLTestCases {
			result := IsLocalhostURL(tc.url)
			assert.Equal(t, tc.expected, result, "unexpected result for URL: "+tc.url)
		}
	})

	t.Run("Base64Encode", func(t *testing.T) {
		for _, tc := range base64EncodeTestCases {
			result := Base64Encode(tc.content)
			assert.Equal(t, tc.expected, result, "unexpected result for content: "+tc.content)
		}
	})
}
