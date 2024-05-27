package agents

import (
	"code/infrastructure/settings"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const prompt = "Someone burnt down my house. What are some applicable laws?"
const phoneNumber = "12223334444"

func TestAgents(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	mySettings, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	bedrockHelper, err := InitializeBedrockHelper(ctx, mySettings.BedrockAgentAliasID, mySettings.BedrockAgentID, mySettings.ContextTimeout)
	assert.NoError(err, "error on initialize bedrock helper: %v", err)
	t.Run("can ask", func(t *testing.T) {
		_, err := bedrockHelper.Ask(ctx, prompt, phoneNumber)
		assert.NoError(err, "error on asking bedrock: %v", err)
	})
}
