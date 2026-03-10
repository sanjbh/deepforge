package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/sanjbh/deepforge/config"
	"github.com/sanjbh/deepforge/internal/agents"
	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/pipeline"
	"github.com/sanjbh/deepforge/internal/search"
	"github.com/sanjbh/deepforge/internal/tools"
	"github.com/sanjbh/deepforge/observability"
	"go.uber.org/zap"
)

func main() {
	var query string

	flag.StringVar(&query, "query", os.Getenv("DEEPFORGE_QUERY"), "research query")
	flag.Parse()

	if query == "" {
		log.Fatalf("query is required")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := observability.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	ctx := context.Background()

	tracer, shutdown, err := observability.NewTracer(
		ctx,
		cfg.ServiceName,
		cfg.ServiceVersion,
		cfg.OTLPEndpoint)

	if err != nil {
		logger.Fatal("failed to create tracer: %w", zap.Error(err))
	}
	defer shutdown()

	obs := observability.NewObservability(logger, tracer)
	defer obs.Sync()

	gemini, err := llm.NewProvider(&llm.ProviderConfig{
		Name:    "gemini",
		APIKey:  cfg.GeminiAPIKey,
		BaseURL: cfg.GeminiBaseURL,
		Model:   cfg.GeminiModel,
	})
	if err != nil {
		logger.Fatal("failed to create provider", zap.Error(err))
	}

	ollama, err := llm.NewProvider(&llm.ProviderConfig{
		Name:    "ollama",
		APIKey:  "ollama",
		BaseURL: cfg.OllamaBaseURL,
		Model:   cfg.OllamaModel,
	})
	if err != nil {
		logger.Fatal("failed to create provider", zap.Error(err))
	}

	provider := llm.NewFallbackProvider(gemini, ollama)

	searchClient := search.NewSearXNGClient(cfg.SearXNGBaseURL, cfg.ResultsPerSearch)

	var sender tools.EmailSender

	switch {
	case cfg.EmailEnabled():
		sender = tools.NewSendGridEmailSender(cfg.SendGridAPIKey, cfg.FromEmail, cfg.ToEmail)
	case cfg.MailHogEnabled():
		sender = tools.NewMailHogEmailSender(cfg.MailhogHost, cfg.MailhogPort, cfg.FromEmail, cfg.ToEmail)
	default:
		sender = tools.NewFileEmailSender("emails")
	}

	planner := agents.NewPlannerAgent(provider, cfg.HowManySearches, obs)
	searcher := agents.NewSearchAgent(provider, searchClient, obs)
	writer := agents.NewWriterAgent(provider, obs)
	emailer := agents.NewEmailAgent(provider, sender, obs)

	pipeLine := pipeline.NewResearcherPipeline(planner, searcher, writer, emailer, obs)

	if err := pipeLine.Run(ctx, query); err != nil {
		logger.Fatal("failed to run pipeLine", zap.Error(err))
	}

}
