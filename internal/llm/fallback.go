package llm

import (
	"context"
	"fmt"

	"github.com/invopop/jsonschema"
)

type FallbackProvider struct {
	primary   Provider
	secondary Provider
}

func NewFallbackProvider(primary, secondary Provider) *FallbackProvider {
	return &FallbackProvider{
		primary:   primary,
		secondary: secondary,
	}
}

func (f *FallbackProvider) Name() string {
	return fmt.Sprintf("%s -> %s", f.primary.Name(), f.secondary.Name())
}

func (f *FallbackProvider) Generate(
	ctx context.Context,
	systemPrompt string,
	userPrompt string) (string, error) {

	var response string
	var err1, err2 error

	response, err1 = f.primary.Generate(ctx, systemPrompt, userPrompt)
	if err1 != nil {
		response, err2 = f.secondary.Generate(ctx, systemPrompt, userPrompt)
		if err2 != nil {
			return "", fmt.Errorf("both providers failed — primary: %w, secondary: %w", err1, err2)
		}
	}

	return response, nil
}

func (f *FallbackProvider) GenerateStructured(
	ctx context.Context,
	systemPrompt string,
	userPrompt string,
	schema *jsonschema.Schema) (string, error) {

	var response string
	var err1, err2 error

	response, err1 = f.primary.GenerateStructured(ctx, systemPrompt, userPrompt, schema)
	if err1 != nil {
		response, err2 = f.secondary.GenerateStructured(ctx, systemPrompt, userPrompt, schema)
		if err2 != nil {
			return "", fmt.Errorf("both providers failed — primary: %w, secondary: %w", err1, err2)
		}
	}
	return response, nil
}
