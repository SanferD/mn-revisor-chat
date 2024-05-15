package stores

import (
	"code/helpers"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Helper struct {
	client        *s3.Client
	bucketName    string
	rawPathPrefix string
	timeout       time.Duration
}

func InitializeS3Helper(ctx context.Context, bucketName string, rawPathPrefix string, timeout time.Duration, endpointURL *string) (*S3Helper, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading default config: %v", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpointURL != nil {
			o.BaseEndpoint = aws.String(*endpointURL)
			if helpers.IsLocalhostURL(*endpointURL) {
				o.UsePathStyle = true
			}
		}
	})
	return &S3Helper{client: client, bucketName: bucketName, rawPathPrefix: rawPathPrefix, timeout: timeout}, nil
}

func (s3Helper *S3Helper) PutTextFile(ctx context.Context, fileName string, body io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, s3Helper.timeout)
	defer cancel()
	key := s3Helper.getRawObjectKey(fileName)
	putObjectInput := &s3.PutObjectInput{Bucket: aws.String(s3Helper.bucketName), Key: aws.String(key), Body: body}
	if _, err := s3Helper.client.PutObject(ctx, putObjectInput); err != nil {
		return fmt.Errorf("error on PutObject: %v", err)
	}
	return nil
}

func (s3Helper *S3Helper) GetTextFile(ctx context.Context, fileName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s3Helper.timeout)
	defer cancel()
	key := s3Helper.getRawObjectKey(fileName)
	getObjectInput := &s3.GetObjectInput{Bucket: aws.String(s3Helper.bucketName), Key: aws.String(key)}
	getObjectOutput, err := s3Helper.client.GetObject(ctx, getObjectInput)
	if err != nil {
		return "", fmt.Errorf("error on get object from s3: %v", err)
	}
	bytes, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return "", fmt.Errorf("error on reading contents of get object body: %v", err)
	}
	return string(bytes), nil
}

func (s3Helper *S3Helper) DeleteTextFile(ctx context.Context, fileName string) error {
	ctx, cancel := context.WithTimeout(ctx, s3Helper.timeout)
	defer cancel()
	key := s3Helper.getRawObjectKey(fileName)
	deleteObjectInput := &s3.DeleteObjectInput{Bucket: aws.String(s3Helper.bucketName), Key: aws.String(key)}
	if _, err := s3Helper.client.DeleteObject(ctx, deleteObjectInput); err != nil {
		return fmt.Errorf("error on deleting object from s3: %v", err)
	}
	return nil
}

func (s3Helper *S3Helper) getRawObjectKey(fileName string) string {
	return s3Helper.rawPathPrefix + "/" + fileName
}
