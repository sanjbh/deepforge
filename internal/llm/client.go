package llm

import (
	"context"
	"encoding/json"
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
		return "", fmt.Errorf("[%s] failed to generate: %w", p.name, err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("[%s] no choices returned", p.name)
	}

	return response.Choices[0].Message.Content, nil
}

func (p *OpenAICompatibleProvider) GenerateStructured(
	ctx context.Context,
	systemPrompt string,
	userPrompt string,
	schema *jsonschema.Schema) (string, error) {

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("[%s] failed to marshal schema: %w", p.name, err)
	}

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
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "response",
				Schema: json.RawMessage(schemaBytes),
				Strict: true,
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("[%s] failed to generate structured: %w", p.name, err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("[%s] no choices returned", p.name)
	}

	return response.Choices[0].Message.Content, nil
}
