package queues

import (
	"code/core"
	"code/infrastructure/settings"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	msg1 = "msg-1"
	msg2 = "msg-2"
	msg3 = "msg-3"
	url1 = "http://hello.com"
)

func TestQueues(t *testing.T) {
	assert := assert.New(t)
	myS, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	ctx := context.Background()
	t.Run("test SQSHelper", func(t *testing.T) {
		sqsHelper, err := InitializeSQSHelper(ctx, myS.URLSQSARN, myS.ContextTimeout, myS.LocalEndpoint)
		assert.NoError(err, "error on initialize sqs helper: %v", err)
		t.Run("test SendMessage, ReceiveMessage, DeleteMessage", func(t *testing.T) {
			var err error
			var msg core.QueueMessage
			err = sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg1})
			assert.NoError(err, "error on send message 1: %v", err)
			msg, err = sqsHelper.ReceiveMessage(ctx)
			assert.NoError(err, "error on receive message: %v", err)
			assert.False(msg.IsEmpty, "message should not be empty")
			assert.Equal(msg1, msg.Body, "messages should be equal")
			err = sqsHelper.DeleteMessage(ctx, msg)
			assert.NoError(err, "error on delete message: %v", err)
			msg, err = sqsHelper.ReceiveMessage(ctx)
			assert.NoError(err, "error on receive message: %v", err)
			assert.True(msg.IsEmpty, "message should be empty")
		})

		t.Run("test Purge", func(t *testing.T) {
			ctx := context.Background()
			sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg1})
			sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg2})
			sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg3})
			sqsHelper.purge(ctx)
			msg, _ := sqsHelper.ReceiveMessage(ctx)
			assert.True(msg.IsEmpty, "message should've been empty but isn't")
		})

		t.Run("test SendURL, DeleteEvent", func(t *testing.T) {
			var err error
			var msg core.QueueMessage
			ctx := context.Background()
			err = sqsHelper.SendURL(ctx, url1)
			assert.NoError(err, "error on send url: %v", err)
			msg, _ = sqsHelper.ReceiveMessage(ctx)
			err = sqsHelper.DeleteMessageByHandle(ctx, msg.Handle)
			assert.NoError(err, "error on delete event: %v", err)
		})

		t.Run("test Clear", func(t *testing.T) {
			var err error
			var msg core.QueueMessage
			ctx := context.Background()
			sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg1})
			sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg2})
			sqsHelper.SendMessage(ctx, core.QueueMessage{Body: msg3})
			err = sqsHelper.Clear(ctx)
			assert.NoError(err, "error on clear: %v", err)
			msg, _ = sqsHelper.ReceiveMessage(ctx)
			assert.True(msg.IsEmpty, "message should be empty")
		})
	})
}
