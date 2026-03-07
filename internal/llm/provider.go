package llm

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
)

const (
	BaseURLGemini = "https://generativelanguage.googleapis.com/v1beta/openai/"
	BaseURLOllama = "http://localhost:11434/v1"
	BaseURLGroq   = "https://api.groq.com/openai/v1"
	BaseURLOpenAI = "https://api.openai.com/v1"
)

// Provider is the interface every agent programs against.
// Gemini and Ollama are both hidden behind this boundary.
type Provider interface {
	// Generate sends a prompt and returns raw text.
	// Used by SearchAgent and EmailAgent.
	Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error)

	// GenerateStructured constrains the response to a JSON schema.
	// Used by PlannerAgent and WriterAgent.
	// Schema is generated automatically from Go structs via invopop/jsonschema.
	GenerateStructured(ctx context.Context, systemPrompt string, userPrompt string, schema *jsonschema.Schema) (string, error)

	// Name returns the provider name for logging and tracing.
	Name() string
}

type ProviderConfig struct {
	Name    string `validate:"required"`
	APIKey  string `validate:"required"`
	BaseURL string `validate:"required"`
	Model   string `validate:"required"`
}

type OpenAICompatibleProvider struct {
	client *openai.Client
	model  string
	name   string
}

var validate = validator.New()

func NewProvider(cfg *ProviderConfig) (*OpenAICompatibleProvider, error) {
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	config := openai.DefaultConfig(cfg.APIKey)
	config.BaseURL = cfg.BaseURL

	return &OpenAICompatibleProvider{
		client: openai.NewClientWithConfig(config),
		model:  cfg.Model,
		name:   cfg.Name,
	}, nil
}
