package llm

import (
	"context"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
)

func (p *OpenAICompatibleProvider) Name() string {
	return p.name
}

func (p *OpenAICompatibleProvider) Generate(
	ctx context.Context,
	systemPrompt string,
	userPrompt string) (string, error) {

	response, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

func (p *OpenAICompatibleProvider) GenerateStructured(
	ctx context.Context,
	systemPrompt string,
	userPrompt string,
	schema *jsonschema.Schema) (string, error) {

	response, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate structured: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return response.Choices[0].Message.Content, nil
}
