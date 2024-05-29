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
	logger    core.Logger
}

type SearchResponse struct {
	Hits struct {
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type ErrorResponse struct {
	Error struct {
		RootCause []struct {
			Type string `json:"type"`
		} `json:"root_cause"`
	} `json:"error"`
}

func InitializeOpenSearchIndexerHelper(ctx context.Context, username, password, domain string, doInsecureSkipVerify bool, indexName string, timeout time.Duration, logger core.Logger) (*OpenSearchIndexerHelper, error) {
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

	transport := &http.Transport{}
	if doInsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	opensearchCfg.Transport = transport

	client, err := opensearch.NewClient(opensearchCfg)
	if err != nil {
		return nil, fmt.Errorf("error on creating new opensearch client: %v", err)
	}

	osiHelper := &OpenSearchIndexerHelper{client: client, indexName: indexName, timeout: timeout, logger: logger}
	if err := osiHelper.createIndex(ctx); err != nil {
		return nil, fmt.Errorf("error on creating index: %v", err)
	}
	return osiHelper, nil
}

func (osiHelper *OpenSearchIndexerHelper) createIndex(ctx context.Context) error {
	logger := osiHelper.logger
	ctx, cancel := context.WithTimeout(ctx, osiHelper.timeout)
	defer cancel()

	// Create the index if it does not exist
	logger.Debug("creating index...")
	settings := strings.NewReader(indexSettingsForKNNEmbeddings)
	req := opensearchapi.IndicesCreateRequest{
		Index: osiHelper.indexName,
		Body:  settings,
	}
	createResp, err := req.Do(ctx, osiHelper.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %v, response=%s", err, createResp.String())
	}
	if createResp.IsError() {
		var values ErrorResponse
		contents, err := io.ReadAll(createResp.Body)
		if err != nil {
			return fmt.Errorf("failed to readall from response body: %v", err)
		}
		if err = json.Unmarshal(contents, &values); err != nil {
			return fmt.Errorf("failed to unmarshal string: %v", err)
		}
		if values.Error.RootCause[0].Type == "resource_already_exists_exception" {
			logger.Debug("index already exists...")
			return nil
		}
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

func (osiHelper *OpenSearchIndexerHelper) FindMatchingChunkIDs(ctx context.Context, vectorDocument core.VectorDocument) ([]string, error) {
	results, err := osiHelper.search(ctx, vectorDocument.Vector, findMatchesK)
	if err != nil {
		return nil, fmt.Errorf("error on search: %v", err)
	}
	var chunkIDs []string = make([]string, 0, len(results))
	for _, result := range results {
		chunkIDs = append(chunkIDs, result.ID)
	}
	return chunkIDs, nil
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

func (osiHelper *OpenSearchIndexerHelper) search(ctx context.Context, embeddings []float64, k int) ([]core.VectorDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, osiHelper.timeout)
	defer cancel()
	body := map[string]interface{}{
		"size": k,
		"query": map[string]interface{}{
			"script_score": map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"script": map[string]interface{}{
					"source": "knn_score",
					"lang":   "knn",
					"params": map[string]interface{}{
						"field":       "vector",
						"query_value": embeddings,
						"space_type":  "cosinesimil",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("error on encoding search request: %v", err)
	}
	searchRequest := opensearchapi.SearchRequest{
		Index: []string{osiHelper.indexName},
		Body:  &buf,
	}
	response, err := searchRequest.Do(ctx, osiHelper.client)
	if err != nil {
		return nil, fmt.Errorf("error on Do search request: %v", err)
	}
	defer response.Body.Close()

	if response.IsError() {
		return nil, fmt.Errorf("received error response from opensearch api: %s", response.String())
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(response.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("error on decoding search response: %v", err)
	}

	results := make([]core.VectorDocument, len(searchResponse.Hits.Hits))
	for i, hit := range searchResponse.Hits.Hits {
		var source core.VectorDocument
		if err := json.Unmarshal(hit.Source, &source); err != nil {
			return nil, fmt.Errorf("error on unmarshalling hit source: %v", err)
		}
		results[i] = source
	}

	return results, nil
}
