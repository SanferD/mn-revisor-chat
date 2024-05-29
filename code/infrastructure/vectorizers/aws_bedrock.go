package vectorizers

import (
	"code/core"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const defaultClientTimeout = 20 * time.Second
const promptAugment1 = "answer the following prompt: "
const promptAugment2 = "\nuse the following knowledge:\n"

var emptyVD = core.VectorDocument{}

type BedrockHelper struct {
	client            *bedrockruntime.Client
	timeout           time.Duration
	embeddingModelID  string
	foundationModelID string
}

type Embeddings struct {
	Embedding []float64 `json:"embedding"`
}

// Define a custom type for parsing the output
type ModelResponse struct {
	Type       string `json:"type"`
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Stop       string `json:"stop"`
}

func InitializeBedrockHelper(ctx context.Context, embeddingModelID, foundationModelID string, timeout time.Duration) (*BedrockHelper, error) {
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

	return &BedrockHelper{
		client:            client,
		timeout:           timeout,
		embeddingModelID:  embeddingModelID,
		foundationModelID: foundationModelID,
	}, nil
}

func (bedrockHelper *BedrockHelper) AskWithChunks(ctx context.Context, prompt string, chunks []core.Chunk) (string, error) {
	// build augmented prompt
	var augmentedPrompt strings.Builder
	augmentedPrompt.WriteString("\n\nHuman: ")
	augmentedPrompt.WriteString(prompt)
	augmentedPrompt.WriteString("\n\nAssistant:\n")
	for _, chunk := range chunks {
		augmentedPrompt.WriteString(chunk.Body)
		augmentedPrompt.WriteString("\n")
	}

	// build input body
	body, err := json.Marshal(map[string]interface{}{
		"prompt":               augmentedPrompt.String(),
		"max_tokens_to_sample": 300,
		"temperature":          0.5,
		"top_k":                250,
		"top_p":                1,
		"stop_sequences":       []string{"\n\nHuman:"},
		"anthropic_version":    "bedrock-2023-05-31",
	})
	if err != nil {
		return "", fmt.Errorf("error on json.Marshal: %v", err)
	}

	// create InvokeModelInput
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("anthropic.claude-v2"),
		Body:        body,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("*/*"),
	}

	// set context and invoke model
	ctx, cancel := context.WithTimeout(ctx, bedrockHelper.timeout)
	defer cancel()

	output, err := bedrockHelper.client.InvokeModel(ctx, input)
	if err != nil {
		return "", fmt.Errorf("error invoking model: %v", err)
	}

	// Parse the response
	var response ModelResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("error parsing model response: %v", err)
	}

	return response.Completion, nil

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
