package queues

import (
	"code/infrastructure/settings"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQS(t *testing.T) {
	mySettings, err := settings.GetSettings()
	assert.NoError(t, err, "error on GetSettings: %v", err)

	ctx := context.Background()

	t.Run("testing SQSHelper", func(t *testing.T) {
		sqsHelper, err := InitializeSQSHelper(ctx, mySettings.URLSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
		assert.NoError(t, err, "error on InitializeSQSHelper: %v", err)

		const msg1 = "msg-1"
		const msg2 = "msg-2"
		const msg3 = "msg-3"
		msgs := []string{msg1, msg2, msg3}

		t.Run("test can send and receive messages", func(t *testing.T) {
			for _, msg := range msgs {
				err := sqsHelper.SendBody(ctx, msg)
				assert.NoError(t, err, "error on SendBody: %v", err)
			}
			for _, msg := range msgs {
				queueMessage, err := sqsHelper.ReceiveQueueMessage(ctx)
				assert.NoError(t, err, "error on ReceiveQueueMessage: %v", err)
				assert.NotEmpty(t, queueMessage, "queueMessage on ReceiveQueueMessage is nil but should be %s", msg)
				assert.Equal(t, msg, queueMessage.Body, "msg=%s != body=%s", msg, queueMessage.Body)

				err = sqsHelper.DeleteQueueMessage(ctx, queueMessage)
				assert.NoError(t, err, "error on DeleteQueueMessage: %v", err)
			}
			queueMessage, err := sqsHelper.ReceiveQueueMessage(ctx)
			assert.NoError(t, err, "error on ReceiveQueueMessage: %v", err)
			assert.Empty(t, queueMessage, "queueMessage on ReceiveQueueMessage should be nil")
		})

		t.Run("test can purge", func(t *testing.T) {
			for _, msg := range msgs {
				err := sqsHelper.SendBody(ctx, msg)
				assert.NoError(t, err, "error on SendBody: %v", err)
			}
			err := sqsHelper.Purge(ctx)
			assert.NoError(t, err, "error on Purge")
			queueMessage, err := sqsHelper.ReceiveQueueMessage(ctx)
			assert.NoError(t, err, "error on ReceiveQueueMessage: %v", err)
			assert.Empty(t, queueMessage, "queueMessage on ReceiveQueueMessage should be nil")
		})
	})

	t.Run("testing URLSQSHelper", func(t *testing.T) {
		urlSQSHelper, err := InitializeURLSQSHelper(ctx, mySettings.URLSQSARN, mySettings.ContextTimeout, mySettings.LocalEndpoint)
		assert.NoError(t, err, "error on InitializeSQSHelper: %v", err)

		const url1 = "https://url1.com"
		const url2 = "https://url2.com"
		const url3 = "https://url3.com"
		urls := []string{url1, url2, url3}

		t.Run("test send and receive URL queue messages", func(t *testing.T) {
			for _, url := range urls {
				err := urlSQSHelper.SendURL(ctx, url)
				assert.NoError(t, err, "error on SendBody: %v", err)
			}
			for _, url := range urls {
				urlQueueMessage, err := urlSQSHelper.ReceiveQueueMessage(ctx)
				assert.NoError(t, err, "error on ReceiveQueueMessage: %v", err)
				assert.NotEmpty(t, urlQueueMessage, "queueMessage on ReceiveQueueMessage is nil but should be %s", url)
				assert.Equal(t, url, urlQueueMessage.Body, "url=%s != body=%s", url, urlQueueMessage.Body)

				err = urlSQSHelper.DeleteQueueMessage(ctx, urlQueueMessage)
				assert.NoError(t, err, "error on DeleteQueueMessage: %v", err)
			}
			queueMessage, err := urlSQSHelper.ReceiveQueueMessage(ctx)
			assert.NoError(t, err, "error on ReceiveQueueMessage: %v", err)
			assert.Empty(t, queueMessage, "queueMessage on ReceiveQueueMessage should be nil")
		})

		t.Run("test can clear", func(t *testing.T) {
			for _, url := range urls {
				err := urlSQSHelper.SendURL(ctx, url)
				assert.NoError(t, err, "error on SendBody: %v", err)
			}
			err := urlSQSHelper.Clear(ctx)
			assert.NoError(t, err, "no error on clear expected, got %v", err)
			urlQueueMessage, err := urlSQSHelper.ReceiveQueueMessage(ctx)
			assert.NoError(t, err, "error on ReceiveQueueMessage: %v", err)
			assert.Empty(t, urlQueueMessage, "queueMessage on ReceiveQueueMessage should be nil")
		})
	})
}
