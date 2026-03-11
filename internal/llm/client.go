package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	req := openai.ChatCompletionRequest{
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
	}

	var response openai.ChatCompletionResponse

	err := p.withRetry(ctx, func() error {
		var e error
		response, e = p.client.CreateChatCompletion(ctx, req)
		return e
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

	req := openai.ChatCompletionRequest{
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
	}

	var response openai.ChatCompletionResponse
	err = p.withRetry(ctx, func() error {
		var e error
		response, e = p.client.CreateChatCompletion(ctx, req)
		return e
	})
	// response, err := p.client.CreateChatCompletion(ctx, req)

	if err != nil {
		return "", fmt.Errorf("[%s] failed to generate structured: %w", p.name, err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("[%s] no choices returned", p.name)
	}

	return response.Choices[0].Message.Content, nil
}

const (
	maxRetries = 3
	baseDelay  = 1 * time.Second
	maxDelay   = 10 * time.Second
)

func (p *OpenAICompatibleProvider) withRetry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := range maxRetries {
		err = fn()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if attempt == maxRetries-1 {
			return err
		}

		delay := min(baseDelay*(1<<attempt), maxDelay)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}
