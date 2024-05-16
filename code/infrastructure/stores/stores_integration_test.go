package stores

import (
	"code/core"
	"code/infrastructure/settings"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testStatute = core.Statute{
	Chapter: "1a", Section: "34", Title: "not a real statute",
	Subdivisions: []core.Subdivision{
		{Number: 1, Heading: "hello", Content: "some sample text"},
		{Number: 2, Heading: "world", Content: "more sample text"},
	},
}

func TestTable1(t *testing.T) {
	mySettings, err := settings.GetSettings()
	assert.NoError(t, err, "error on GetSettings: %v", err)

	ctx := context.Background()

	table1, err := InitializeTable1(ctx, mySettings.Table1ARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	assert.NoError(t, err, "error on InitializeTable1: %v", err)

	const url1 = "https://url1.com"
	const url2 = "https://url2.com"

	t.Run("test Put, Has, DeleteAll seen urls", func(t *testing.T) {
		hasURL, err := table1.HasURL(ctx, url1)
		assert.NoError(t, err, "error on HasURL url=%s: %v", url1, err)
		assert.False(t, hasURL, "url=%s not in table1", url1)

		err = table1.PutURL(ctx, url1)
		assert.NoError(t, err, "error on PutURL url=%s: %v", url1, err)

		hasURL, err = table1.HasURL(ctx, url1)
		assert.NoError(t, err, "error on HasURL url=%s: %v", url1, err)
		assert.True(t, hasURL, "url=%s not in table1", url1)

		hasURL, err = table1.HasURL(ctx, url2)
		assert.NoError(t, err, "error on HasURL url=%s: %v", url2, err)
		assert.False(t, hasURL, "url=%s in table1 but shouldn't be", url2)

		err = table1.PutURL(ctx, url2)
		assert.NoError(t, err, "error on PutURL url=%s: %v", url2, err)

		err = table1.DeleteAll(ctx)
		assert.NoError(t, err, "error on DeleteAll: %v", err)

		for _, url := range []string{url1, url2} {
			hasURL, err := table1.HasURL(ctx, url)
			assert.NoError(t, err, "error on HasURL after DeleteAll, url=%s: %v", url, err)
			assert.False(t, hasURL, "url=%s in table but shouldn't be", url)
		}
	})
}

func TestS3Helper(t *testing.T) {
	mySettings, err := settings.GetSettings()
	assert.NoError(t, err, "error on GetSettings: %v", err)

	ctx := context.Background()

	s3Helper, err := InitializeS3Helper(ctx, mySettings.BucketName, mySettings.RawPathPrefix, mySettings.ChunkPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	assert.NoError(t, err, "error on InitializeS3Helper: %v", err)

	t.Run("test PutTextFile, GetTextFile, DeleteTextFile", func(t *testing.T) {
		fileName := "my-file.txt"
		fileContents := "some test contents."
		body := strings.NewReader(fileContents)

		// put returns no errors
		err := s3Helper.PutTextFile(ctx, fileName, body)
		assert.NoError(t, err, "error on PutTextFile with fileName=%s: %v", fileName, err)

		// get returns the file contents
		foundContents, err := s3Helper.GetTextFile(ctx, fileName)
		assert.NoError(t, err, "error no get text file with fileName=%s: %v", fileName, err)
		assert.Equal(t, fileContents, foundContents, "get text file contents are not the same.")

		// delete followed by get confirms that the file is not found
		err = s3Helper.DeleteTextFile(ctx, fileName)
		assert.NoError(t, err, "error on delete text file with fileName=%s: %v", fileName, err)
		_, err = s3Helper.GetTextFile(ctx, fileName)
		assert.Error(t, err, "get text file was supposed to receive an error after deletion, fileName=%s", fileName)
	})

	t.Run("test PutStatute, GetStatute, DeleteStatute", func(t *testing.T) {

		// put returns no error
		err := s3Helper.PutStatute(ctx, testStatute)
		assert.NoError(t, err, "error on put statute: %v", err)

		// get retrieves the statute
		fileName := testStatute.Chapter + "." + testStatute.Section + " " + testStatute.Title
		key := s3Helper.getChunkObjectKey(fileName)
		statute, err := s3Helper.GetStatute(ctx, key)
		assert.NoError(t, err, "error on get statute: %v", err)
		assert.Equal(t, testStatute, statute, "error statutes are not equal but should be")

		// delete returns no error
		err = s3Helper.DeleteStatute(ctx, statute)
		assert.NoError(t, err, "delete returned an error")

		// get after delete returns error
		_, err = s3Helper.GetStatute(ctx, key)
		assert.Error(t, err, "did not get an error after deleting statute")
	})

}
