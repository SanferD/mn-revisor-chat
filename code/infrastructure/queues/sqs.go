package queues

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"code/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	smithy "github.com/aws/smithy-go"
)

type URLSQSHelper struct {
	SQSHelper
}

func InitializeURLSQSHelper(ctx context.Context, queueArn string, timeout time.Duration, endpoint *string) (*URLSQSHelper, error) {
	sqsHelper, err := InitializeSQSHelper(ctx, queueArn, timeout, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error on InitializeSQS: %v", err)
	}
	return &URLSQSHelper{SQSHelper: *sqsHelper}, nil
}

func (helper *URLSQSHelper) Clear(ctx context.Context) error {
	if err := helper.Purge(ctx); err != nil {
		return fmt.Errorf("error on Purge: %v", err)
	}
	return nil
}

func (helper *URLSQSHelper) SendURL(ctx context.Context, url string) error {
	if err := helper.SendBody(ctx, url); err != nil {
		return fmt.Errorf("error on SendBody: %v", err)
	}
	return nil
}

func (urlSQSHelper *URLSQSHelper) ReceiveURLQueueMessage(ctx context.Context) (*core.URLQueueMessage, error) {
	queueMessage, err := urlSQSHelper.ReceiveQueueMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("error on ReceiveURLQueueMessage: %v", err)
	}
	if queueMessage == nil {
		return nil, nil
	}
	return &core.URLQueueMessage{QueueMessage: *queueMessage}, nil
}

func (urlSQSHelper *URLSQSHelper) DeleteURLQueueMessage(ctx context.Context, urlQueueMessage *core.URLQueueMessage) error {
	if err := urlSQSHelper.DeleteQueueMessage(ctx, &urlQueueMessage.QueueMessage); err != nil {
		return fmt.Errorf("error on DeleteQueueMessage: %v", err)
	}
	return nil
}

type SQSHelper struct {
	client   *sqs.Client
	queueURL string
	timeout  time.Duration
}

func InitializeSQSHelper(ctx context.Context, queueArn string, timeout time.Duration, endpoint *string) (*SQSHelper, error) {
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
	queueURL, err := getURLFromARN(ctx, client, strings.TrimSpace(queueArn), timeout)
	if err != nil {
		return nil, fmt.Errorf("error getting QueueURL for QueueARN='%s': %v", queueArn, err)
	}
	return &SQSHelper{client: client, queueURL: queueURL, timeout: timeout}, nil
}

func getURLFromARN(ctx context.Context, client *sqs.Client, queueArn string, timeout time.Duration) (string, error) {
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
			if _queueArn == queueArn {
				return queueURL, nil
			}
		}
	}
	return "", fmt.Errorf("queue with queue-arn='%s' not found", queueArn)
}

func (sqsHelper *SQSHelper) Purge(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	_, err := sqsHelper.client.PurgeQueue(ctx, &sqs.PurgeQueueInput{QueueUrl: &sqsHelper.queueURL})
	if err != nil {
		return fmt.Errorf("error purging queue for queueURL=%s: %v", sqsHelper.queueURL, err)
	}
	return nil
}

func (sqsHelper *SQSHelper) SendBody(ctx context.Context, body string) error {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	_, err := sqsHelper.client.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: &sqsHelper.queueURL, MessageBody: &body})
	if err != nil {
		return fmt.Errorf("sqs SendMessage error: %v", err)
	}
	return nil
}

func (sqsHelper *SQSHelper) ReceiveQueueMessage(ctx context.Context) (*core.QueueMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	output, err := sqsHelper.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: &sqsHelper.queueURL, MaxNumberOfMessages: 1})
	if err != nil {
		return nil, fmt.Errorf("sqs ReceiveMesasge error: %v", err)
	}
	if len(output.Messages) == 0 {
		return nil, nil
	}
	message := output.Messages[0]
	sqsMessage := &core.QueueMessage{Body: *message.Body, ID: *message.MessageId, Handle: *message.ReceiptHandle}
	return sqsMessage, nil
}

func (sqsHelper *SQSHelper) DeleteQueueMessage(ctx context.Context, queueMessage *core.QueueMessage) error {
	return sqsHelper.DeleteMessage(ctx, queueMessage.Handle)
}

func (sqsHelper *SQSHelper) DeleteMessage(ctx context.Context, receiptHandle string) error {
	ctx, cancel := context.WithTimeout(ctx, sqsHelper.timeout)
	defer cancel()
	deleteMessageInput := &sqs.DeleteMessageInput{QueueUrl: &sqsHelper.queueURL, ReceiptHandle: &receiptHandle}
	_, err := sqsHelper.client.DeleteMessage(ctx, deleteMessageInput)
	if err != nil {
		return fmt.Errorf("error on DeleteMessage: %v", err)
	}
	return nil

}
