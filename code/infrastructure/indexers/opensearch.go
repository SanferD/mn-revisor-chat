package indexers

import (
	"bytes"
	"code/core"
	"code/helpers"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	opensearch "github.com/opensearch-project/opensearch-go/v4"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
)

const findMatchesK = 100
const subdivisionKNNIndexName = "subdivision-knn"
const indexSettingsForKNNEmbeddings = `{
	"settings": {
		"index": {
			"knn": true,
			"knn.algo_param.ef_search": 100,
			"knn.algo_param.ef_construction": 200,
			"knn.algo_param.m": 16
		}
	},
	"mappings": {
		"properties": {
			"vector": {
				"type": "knn_vector",
				"dimension": 1024
			}
		}
	}
}`

type OpenSearchIndexerHelper struct {
	client    *opensearch.Client
	indexName string
	timeout   time.Duration
}

func InitializeOpenSearchIndexerHelper(ctx context.Context, username, password, domain string, doInsecureSkipVerify bool, indexName string, timeout time.Duration) (*OpenSearchIndexerHelper, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	opensearchCfg := opensearch.Config{
		Addresses: []string{domain},
		Username:  username,
		Password:  password,
	}
	if !helpers.IsLocalhostURL(domain) {
		awsCfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading aws default config: %v", err)
		}
		opensearchCfg.Signer, err = requestsigner.NewSignerWithService(awsCfg, "es")
		if err != nil {
			return nil, fmt.Errorf("error on creating new signer for 'es' service: %v", err)
		}
	}
	if doInsecureSkipVerify {
		opensearchCfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client, err := opensearch.NewClient(opensearchCfg)
	if err != nil {
		return nil, fmt.Errorf("error on creating new opensearch client: %v", err)
	}

	osiHelper := &OpenSearchIndexerHelper{client: client, indexName: indexName, timeout: timeout}
	if err := osiHelper.createIndex(ctx); err != nil {
		return nil, fmt.Errorf("error on creating index: %v", err)
	}
	return osiHelper, nil
}

func (osiHelper *OpenSearchIndexerHelper) createIndex(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, osiHelper.timeout)
	defer cancel()
	// Check if the index already exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{osiHelper.indexName},
	}
	existsResp, err := existsReq.Do(ctx, osiHelper.client)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %v", err)
	}
	if existsResp.StatusCode == 200 {
		// Index already exists
		return nil
	}

	// Create the index if it does not exist
	settings := strings.NewReader(indexSettingsForKNNEmbeddings)
	req := opensearchapi.IndicesCreateRequest{
		Index: osiHelper.indexName,
		Body:  settings,
	}
	createResp, err := req.Do(ctx, osiHelper.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}
	if createResp.IsError() {
		return fmt.Errorf("failed to create index: response=%s", createResp.String())
	}

	return nil
}

func (osiHelper *OpenSearchIndexerHelper) AddVectorDocument(ctx context.Context, vectorDocument core.VectorDocument) error {
	if err := osiHelper.addToIndex(ctx, vectorDocument); err != nil {
		return fmt.Errorf("error on adding to index: %v", err)
	}
	return nil
}

func (osiHelper *OpenSearchIndexerHelper) addToIndex(ctx context.Context, vectorDocument core.VectorDocument) error {
	ctx, cancel := context.WithTimeout(ctx, osiHelper.timeout)
	defer cancel()
	docBytes, err := json.Marshal(vectorDocument)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %v", err)
	}
	req := opensearchapi.IndexRequest{
		Index:      osiHelper.indexName,
		DocumentID: vectorDocument.ID,
		Body:       bytes.NewReader(docBytes),
	}
	resp, err := req.Do(ctx, osiHelper.client)
	if err != nil {
		return fmt.Errorf("failed to add document to index: %v", err)
	}
	if resp.IsError() {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to add document to index: received status-code=%d: failed to read response body: %v, response=%s", resp.StatusCode, err, resp.String())
		}
		return fmt.Errorf("failed to add document to index: received status-code=%d, body=%s", resp.StatusCode, string(bytes))
	}
	return nil
}
