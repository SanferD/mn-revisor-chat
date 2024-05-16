package stores

import (
	"bytes"
	"code/core"
	"code/helpers"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Helper struct {
	client          *s3.Client
	bucketName      string
	rawPathPrefix   string
	chunkPathPrefix string
	timeout         time.Duration
}

func InitializeS3Helper(ctx context.Context, bucketName, rawPathPrefix, chunkPathPrefix string, timeout time.Duration, endpointURL *string) (*S3Helper, error) {
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
	return &S3Helper{client: client, bucketName: bucketName, rawPathPrefix: rawPathPrefix, chunkPathPrefix: chunkPathPrefix, timeout: timeout}, nil
}

func (s3Helper *S3Helper) PutStatute(ctx context.Context, statute core.Statute) error {
	fileName := s3Helper.statuteToFileName(statute)
	statuteJsonBytes, err := json.Marshal(statute)
	if err != nil {
		return fmt.Errorf("error when converting statute to json: %v", err)
	}
	body := bytes.NewReader(statuteJsonBytes)
	key := s3Helper.getChunkObjectKey(fileName)
	return s3Helper.putFile(ctx, key, body)
}

func (s3Helper *S3Helper) GetStatute(ctx context.Context, key string) (core.Statute, error) {
	prefix := s3Helper.chunkPathPrefix + "/"
	if !strings.HasPrefix(key, prefix) {
		return core.Statute{}, fmt.Errorf("key doesn't have the correct prefix, prefix='%s', key='%s'", prefix, key)
	}
	contents, err := s3Helper.getObject(ctx, key)
	if err != nil {
		return core.Statute{}, fmt.Errorf("error on reading object from s3: %v", err)
	}
	var statute core.Statute
	if err := json.Unmarshal([]byte(contents), &statute); err != nil {
		return core.Statute{}, fmt.Errorf("error on unmarshalling json object: %v", err)
	}
	return statute, nil
}

func (s3Helper *S3Helper) DeleteStatute(ctx context.Context, statute core.Statute) error {
	fileName := s3Helper.statuteToFileName(statute)
	key := s3Helper.getChunkObjectKey(fileName)
	if err := s3Helper.deleteObject(ctx, key); err != nil {
		return fmt.Errorf("error on deleting object: %v", err)
	}
	return nil
}

func (s3Helper *S3Helper) PutTextFile(ctx context.Context, fileName string, body io.Reader) error {
	key := s3Helper.getRawObjectKey(fileName)
	return s3Helper.putFile(ctx, key, body)
}

func (s3Helper *S3Helper) GetTextFile(ctx context.Context, fileName string) (string, error) {
	key := s3Helper.getRawObjectKey(fileName)
	return s3Helper.getObject(ctx, key)
}

func (s3Helper *S3Helper) DeleteTextFile(ctx context.Context, fileName string) error {
	key := s3Helper.getRawObjectKey(fileName)
	return s3Helper.deleteObject(ctx, key)
}

func (s3Helper *S3Helper) putFile(ctx context.Context, key string, body io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, s3Helper.timeout)
	defer cancel()
	putObjectInput := &s3.PutObjectInput{Bucket: aws.String(s3Helper.bucketName), Key: aws.String(key), Body: body}
	if _, err := s3Helper.client.PutObject(ctx, putObjectInput); err != nil {
		return fmt.Errorf("error on PutObject: %v", err)
	}
	return nil
}

func (s3Helper *S3Helper) getObject(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s3Helper.timeout)
	defer cancel()
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

func (s3Helper *S3Helper) deleteObject(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, s3Helper.timeout)
	defer cancel()
	deleteObjectInput := &s3.DeleteObjectInput{Bucket: aws.String(s3Helper.bucketName), Key: aws.String(key)}
	if _, err := s3Helper.client.DeleteObject(ctx, deleteObjectInput); err != nil {
		return fmt.Errorf("error on deleting object from s3: %v", err)
	}
	return nil

}

func (s3Helper *S3Helper) statuteToFileName(statute core.Statute) string {
	return statute.Chapter + "." + statute.Section + " " + statute.Title
}

func (s3Helper *S3Helper) getRawObjectKey(fileName string) string {
	return s3Helper.rawPathPrefix + "/" + fileName
}

func (s3Helper *S3Helper) getChunkObjectKey(fileName string) string {
	return s3Helper.chunkPathPrefix + "/" + fileName
}
