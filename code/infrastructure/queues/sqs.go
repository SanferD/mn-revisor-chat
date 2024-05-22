package queues

import (
	"code/core"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	smithy "github.com/aws/smithy-go"
)

type SQSHelper struct {
	client   *sqs.Client
	queueURL string
	timeout  time.Duration
}

var emptyQMsg = core.QueueMessage{IsEmpty: true}

func InitializeSQSHelper(ctx context.Context, queueARN string, timeout time.Duration, endpoint *string) (*SQSHelper, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error on loading default config: %v", err)
	}
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		if endpoint != nil {
			o.BaseEndpoint = aws.String(*endpoint)
		}
	})
	queueURL, err := getURLFromARN(ctx, client, strings.TrimSpace(queueARN), timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting QueueURL for QueueARN='%s': %v", queueARN, err)
	}
	return &SQSHelper{client: client, queueURL: queueURL, timeout: timeout}, nil
}

func getURLFromARN(ctx context.Context, client *sqs.Client, queueARN string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})
	for paginator.HasMorePages() {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("error on fetching NextPage: %v", err)
		}
		for _, queueURL := range page.QueueUrls {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			getQueueAttributesInput := &sqs.GetQueueAttributesInput{
				QueueUrl: &queueURL,
				AttributeNames: []types.QueueAttributeName{
					types.QueueAttributeNameQueueArn,
				},
			}
			attribOutput, err := client.GetQueueAttributes(ctx, getQueueAttributesInput)
			if err != nil {
				var apiErr smithy.APIError
				if errors.As(err, &apiErr) && apiErr.ErrorCode() == "AccessDenied" {
					continue // 403 permission error, try next queue
				} else {
					return "", err // return on any other error
				}
			}

			_queueArn := strings.TrimSpace(attribOutput.Attributes[string(types.QueueAttributeNameQueueArn)])
			if _queueArn == queueARN {
				return queueURL, nil
			}
		}
	}
	return "", fmt.Errorf("queue with queue-arn='%s' not found", queueARN)
}

func (sqsHelper *SQSHelper) SendURL(ctx context.Context, url string) error {
	return sqsHelper.SendMessage(ctx, core.QueueMessage{Body: url})
}

func (sqsHelper *SQSHelper) DeleteMessageByHandle(ctx context.Context, handle string) error {
	return sqsHelper.DeleteMessage(ctx, core.QueueMessage{Handle: handle})
}

func (sqsHelper *SQSHelper) Clear(ctx context.Context) error {
	return sqsHelper.purge(ctx)
}

func (sqsHelper *SQSHelper) SendMessage(ctx context.Context, queueMessage core.QueueMessage) error {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	sendMsgInp := sqs.SendMessageInput{QueueUrl: &sqsHelper.queueURL, MessageBody: &queueMessage.Body}
	_, err := sqsHelper.client.SendMessage(ctx, &sendMsgInp)
	if err != nil {
		return fmt.Errorf("sqs send message error: %v", err)
	}
	return nil
}

func (sqsHelper *SQSHelper) ReceiveMessage(ctx context.Context) (core.QueueMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	recvMsgInp := sqs.ReceiveMessageInput{QueueUrl: &sqsHelper.queueURL, MaxNumberOfMessages: 1}
	recvMsgOut, err := sqsHelper.client.ReceiveMessage(ctx, &recvMsgInp)
	if err != nil {
		return emptyQMsg, fmt.Errorf("error on sqs receive message: %v", err)
	}
	if len(recvMsgOut.Messages) == 0 {
		return emptyQMsg, nil
	}
	message := recvMsgOut.Messages[0]
	ret := core.QueueMessage{Body: *message.Body, Handle: *message.ReceiptHandle}
	return ret, nil
}

func (sqsHelper *SQSHelper) DeleteMessage(ctx context.Context, queueMessage core.QueueMessage) error {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	deleteMsgInp := &sqs.DeleteMessageInput{QueueUrl: &sqsHelper.queueURL, ReceiptHandle: &queueMessage.Handle}
	_, err := sqsHelper.client.DeleteMessage(ctx, deleteMsgInp)
	if err != nil {
		return fmt.Errorf("error on sqs delete message: %v", err)
	}
	return nil
}

func (sqsHelper *SQSHelper) purge(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	purgeQueueInp := sqs.PurgeQueueInput{QueueUrl: &sqsHelper.queueURL}
	_, err := sqsHelper.client.PurgeQueue(ctx, &purgeQueueInp)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok && apiErr.ErrorCode() == "PurgeQueueInProgress" {
			return nil
		}
		return fmt.Errorf("error purging queue for queueURL=%s: %v", sqsHelper.queueURL, err)
	}
	return nil
}
