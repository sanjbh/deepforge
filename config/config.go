package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	//LLM
	GeminiAPIKey  string `mapstructure:"GEMINI_API_KEY" validate:"required,min=20"`
	GeminiModel   string `mapstructure:"GEMINI_MODEL" validate:"required"`
	OllamaModel   string `mapstructure:"OLLAMA_MODEL" validate:"required"`
	GeminiBaseURL string `mapstructure:"GEMINI_BASE_URL" validate:"required,url"`
	OllamaBaseURL string `mapstructure:"OLLAMA_BASE_URL" validate:"required,url"`
	//Search
	HowManySearches int `mapstructure:"HOW_MANY_SEARCHES" validate:"required,min=1,max=10"`

	//Email
	SendGridAPIKey string `mapstructure:"SENDGRID_API_KEY"`
	FromEmail      string `mapstructure:"FROM_EMAIL" validate:"required_with=SendGridAPIKey"`
	ToEmail        string `mapstructure:"TO_EMAIL" validate:"required_with=SendGridAPIKey"`

	//Observability
	ServiceName    string `mapstructure:"SERVICE_NAME" validate:"required"`
	ServiceVersion string `mapstructure:"SERVICE_VERSION" validate:"required"`
	LogLevel       string `mapstructure:"LOG_LEVEL" validate:"required"`
	OTLPEndpoint   string `mapstructure:"OTLP_ENDPOINT" validate:"required"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Explicitly bind each env var — required for viper to map
	// env vars to struct fields via mapstructure tags
	viper.BindEnv("GEMINI_API_KEY")
	viper.BindEnv("GEMINI_MODEL")
	viper.BindEnv("GEMINI_BASE_URL")
	viper.BindEnv("OLLAMA_BASE_URL")
	viper.BindEnv("OLLAMA_MODEL")
	viper.BindEnv("HOW_MANY_SEARCHES")
	viper.BindEnv("SENDGRID_API_KEY")
	viper.BindEnv("FROM_EMAIL")
	viper.BindEnv("TO_EMAIL")
	viper.BindEnv("SERVICE_NAME")
	viper.BindEnv("SERVICE_VERSION")
	viper.BindEnv("LOG_LEVEL")
	viper.BindEnv("OTLP_ENDPOINT")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("warning: .env not found, using environment variables\n")
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) EmailEnabled() bool {
	return c.SendGridAPIKey != ""
}
