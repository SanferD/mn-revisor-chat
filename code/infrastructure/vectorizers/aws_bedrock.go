package vectorizers

import (
	"code/core"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const defaultClientTimeout = 20 * time.Second

type BedrockHelper struct {
	client           *bedrockruntime.Client
	timeout          time.Duration
	embeddingModelID string
}

type Embeddings struct {
	Embedding []float64 `json:"embedding"`
}

var emptyVD = core.VectorDocument{}

func InitializeBedrockHelper(ctx context.Context, embeddingModelID string, timeout time.Duration) (*BedrockHelper, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading default config: %v", err)
	}
	customHTTPClient := &http.Client{
		Timeout: defaultClientTimeout,
	}
	customCfg := aws.Config{
		Region:      cfg.Region,
		Credentials: cfg.Credentials,
		HTTPClient:  customHTTPClient,
	}
	client := bedrockruntime.NewFromConfig(customCfg)

	return &BedrockHelper{client: client, timeout: timeout, embeddingModelID: embeddingModelID}, nil
}

func (bedrockHelper *BedrockHelper) VectorizeChunk(ctx context.Context, chunk core.Chunk) (core.VectorDocument, error) {
	embeddings, err := bedrockHelper.getEmbeddings(ctx, chunk.Body)
	if err != nil {
		return emptyVD, fmt.Errorf("error on get embeddings: %v", err)
	}
	return core.VectorDocument{ID: chunk.ID, Vector: embeddings}, nil
}

func (bedrockHelper *BedrockHelper) Vectorize(ctx context.Context, content string) (core.VectorDocument, error) {
	embeddings, err := bedrockHelper.getEmbeddings(ctx, content)
	if err != nil {
		return emptyVD, fmt.Errorf("error on get embeddings: %v", err)
	}
	return core.VectorDocument{ID: "", Vector: embeddings}, nil
}

func (bedrockHelper *BedrockHelper) getEmbeddings(ctx context.Context, inputText string) ([]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, bedrockHelper.timeout)
	defer cancel()
	body, err := json.Marshal(map[string]interface{}{
		"inputText": inputText,
	})
	if err != nil {
		return nil, fmt.Errorf("error on json.Marshal: %v", err)
	}
	invkInp := bedrockruntime.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(bedrockHelper.embeddingModelID),
		ContentType: aws.String("application/json")}
	invkOut, err := bedrockHelper.client.InvokeModel(ctx, &invkInp)
	if err != nil {
		return nil, fmt.Errorf("error on invoke model: %v", err)
	}
	var response Embeddings
	json.Unmarshal(invkOut.Body, &response)
	return response.Embedding, nil
}
