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
	response, err := f.primary.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		_, err2 := f.secondary.Generate(ctx, systemPrompt, userPrompt)
		if err2 != nil {
			return "", fmt.Errorf("both providers failed — primary: %w, secondary: %v", err, err2)
		}
	}

	return response, nil
}

func (f *FallbackProvider) GenerateStructured(
	ctx context.Context,
	systemPrompt string,
	userPrompt string,
	schema *jsonschema.Schema) (string, error) {
	response, err := f.primary.GenerateStructured(ctx, systemPrompt, userPrompt, schema)
	if err != nil {
		f.secondary.GenerateStructured(ctx, systemPrompt, userPrompt, schema)
	}
	return response, nil
}
