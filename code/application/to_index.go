package application

import (
	"code/core"
	"context"
	"fmt"
)

func Index(ctx context.Context, chunkFileName string, chunksDataStore core.ChunksDataStore, vectorizer core.Vectorizer, searchIndex core.SearchIndex, logger core.Logger) error {
	logger.Info("getting chunk with fileName=%s", chunkFileName)
	chunk, err := chunksDataStore.GetChunk(ctx, chunkFileName)
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
	logger.Info("successfully indexed chunk wtih name chunkFileName=%s", chunkFileName)
	return nil
}
