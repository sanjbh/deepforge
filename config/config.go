package config

import (
	"fmt"
	"os"

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
	HowManySearches  int    `mapstructure:"HOW_MANY_SEARCHES" validate:"required,min=1,max=10"`
	SearXNGBaseURL   string `mapstructure:"SEARXNG_BASE_URL" validate:"required,url"`
	ResultsPerSearch int    `mapstructure:"RESULTS_PER_SEARCH" validate:"min=1,max=20"`
	DeepForgeQuery   string `mapstructure:"DEEPFORGE_QUERY" validate:"required"`

	//Email
	SendGridAPIKey string `mapstructure:"SENDGRID_API_KEY"`
	FromEmail      string `mapstructure:"FROM_EMAIL" validate:"required_with=SendGridAPIKey"`
	ToEmail        string `mapstructure:"TO_EMAIL" validate:"required_with=SendGridAPIKey"`
	MailhogHost    string `mapstructure:"MAILHOG_HOST"`
	MailhogPort    int    `mapstructure:"MAILHOG_PORT"`

	//Observability
	ServiceName    string `mapstructure:"SERVICE_NAME" validate:"required"`
	ServiceVersion string `mapstructure:"SERVICE_VERSION" validate:"required"`
	LogLevel       string `mapstructure:"LOG_LEVEL" validate:"required"`
	OTLPEndpoint   string `mapstructure:"OTLP_ENDPOINT" validate:"required"`
}

func Load() (*Config, error) {
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
	viper.BindEnv("SEARXNG_BASE_URL")
	viper.BindEnv("RESULTS_PER_SEARCH")
	viper.BindEnv("DEEPFORGE_QUERY")
	viper.BindEnv("MAILHOG_HOST")
	viper.BindEnv("MAILHOG_PORT")

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	/* if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("warning: .env not found, using environment variables\n")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	} */

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

func (c *Config) MailHogEnabled() bool {
	return c.MailhogHost != "" && c.MailhogPort != 0
}
