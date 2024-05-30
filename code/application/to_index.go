package application

import (
	"code/core"
	"context"
	"fmt"
)

func Index(ctx context.Context, chunkID string, chunksDataStore core.ChunksDataStore, vectorizer core.Vectorizer, searchIndex core.SearchIndex, logger core.Logger) error {
	logger.Info("initializing index")
	searchIndex.SetupIndexIfNecessary(ctx)

	logger.Info("getting chunk with chunkID=%s", chunkID)
	chunk, err := chunksDataStore.GetChunk(ctx, chunkID)
	if err != nil {
		return fmt.Errorf("error on get chunk: %v", err)
	}
	logger.Info("vectorizing chunk")
	vectorDocument, err := vectorizer.VectorizeChunk(ctx, chunk)
	if err != nil {
		return fmt.Errorf("error on vectorize chunk: %v", err)
	}
	logger.Info("adding vector document of chunk to index")
	if err = searchIndex.AddVectorDocument(ctx, vectorDocument); err != nil {
		return fmt.Errorf("error on adding vector document: %v", err)
	}
	logger.Info("successfully indexed chunk wtih name chunkID=%s", chunkID)
	return nil
}
