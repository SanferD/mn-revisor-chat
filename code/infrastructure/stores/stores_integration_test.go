package stores

import (
	"code/infrastructure/settings"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	s3Helper, err := InitializeS3Helper(ctx, mySettings.BucketName, mySettings.RawPathPrefix, mySettings.ContextTimeout, mySettings.LocalEndpoint)
	assert.NoError(t, err, "error on InitializeS3Helper: %v", err)

	t.Run("test PutTestFile", func(t *testing.T) {
		fileName := "my-file.txt"
		fileContents := "some test contents."
		body := strings.NewReader(fileContents)
		err := s3Helper.PutTextFile(ctx, fileName, body)
		assert.NoError(t, err, "error on PutTextFile with fileName=%s: %v", fileName, err)
	})
}
