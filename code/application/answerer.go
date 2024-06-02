package application

import (
	"code/core"
	"context"
	"fmt"
)

func Answer(ctx context.Context, prompt, phoneNumber string, chunkStore core.ChunksDataStore, agent core.Agent, indexer core.SearchIndex, vectorizer core.Vectorizer, comms core.Comms, logger core.Logger) error {

	logger.Info("received prompt='%s'", prompt)

	logger.Info("vectorize prompt")
	promptVD, err := vectorizer.Vectorize(ctx, prompt)
	if err != nil {
		return fmt.Errorf("error vectorizing prompt: %v", err)
	}

	logger.Info("search index for matching chunk ids")
	chunkIDs, err := indexer.FindMatchingChunkIDs(ctx, promptVD)
	if err != nil {
		return fmt.Errorf("error finding matching chunk ids: %v", err)
	}

	if len(chunkIDs) == 0 {
		logger.Info("no matching chunks found")
	} else {
		logger.Info("getting chunks corresponding to matching chunk ids")
	}
	var chunks = make([]core.Chunk, 0, len(chunkIDs))
	for _, chunkID := range chunkIDs {
		logger.Info("getting chunk with chunkID=%s", chunkID)
		chunk, err := chunkStore.GetChunk(ctx, chunkID)
		if err != nil {
			return fmt.Errorf("error getting chunk for chunkID=%s: %v", chunkID, err)
		}
		chunks = append(chunks, chunk)
	}

	logger.Info("ask agent with prompt and chunks")
	answer, err := agent.AskWithChunks(ctx, prompt, chunks)
	if err != nil {
		return fmt.Errorf("error asking agent prompt with chunks: %v", err)
	}

	logger.Info("sending to phoneNumber=%s the answer=%s", phoneNumber, answer)
	if err = comms.SendMessage(ctx, phoneNumber, answer); err != nil {
		return fmt.Errorf("error on send message: %v", err)
	}

	logger.Info("answered prompt=%s", prompt)
	return nil
}
