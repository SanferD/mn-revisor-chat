package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

type BedrockHelper struct {
	client       *bedrockagentruntime.Client
	timeout      time.Duration
	agentID      string
	agentAliasID string
}

func InitializeBedrockHelper(ctx context.Context, agentAliasID, agentID string, timeout time.Duration) (*BedrockHelper, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error on loading default config: %v", err)
	}
	client := bedrockagentruntime.NewFromConfig(cfg)

	return &BedrockHelper{
		client:       client,
		timeout:      timeout,
		agentID:      agentID,
		agentAliasID: agentAliasID,
	}, nil
}

func (bedrockHelper *BedrockHelper) Ask(ctx context.Context, prompt, sessionID string) (string, error) {
	return bedrockHelper.invokeAgent(ctx, prompt, sessionID)
}

func (bedrockHelper *BedrockHelper) invokeAgent(ctx context.Context, prompt, sessionID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, bedrockHelper.timeout)
	defer cancel()
	input := bedrockagentruntime.InvokeAgentInput{
		AgentAliasId: aws.String(bedrockHelper.agentAliasID),
		AgentId:      aws.String(bedrockHelper.agentID),
		InputText:    aws.String(prompt),
		SessionId:    aws.String(sessionID),
	}
	output, err := bedrockHelper.client.InvokeAgent(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("error on invoke agent: %v", err)
	}
	var textBuilder strings.Builder
	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			content := string(v.Value.Bytes)
			textBuilder.WriteString(content)
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)

		default:
			fmt.Println("union is nil or unknown type")

		}
	}

	return textBuilder.String(), nil
}
