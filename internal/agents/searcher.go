package agents

import (
	"context"
	"fmt"

	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
	"github.com/sanjbh/deepforge/internal/search"
	"github.com/sanjbh/deepforge/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type SearchAgent struct {
	provider     llm.Provider
	searchClient *search.DuckDuckGoClient
	obs          *observability.Observability
}

func NewSearchAgent(provider llm.Provider, searchClient *search.DuckDuckGoClient, obs *observability.Observability) *SearchAgent {
	return &SearchAgent{
		provider:     provider,
		searchClient: searchClient,
		obs:          obs,
	}
}

const searcherSystemPrompt = `
	You are a research assistant. Given a search term and raw search results,
	produce a concise summary of the results. The summary must be 2-3 paragraphs 
	and less than 300 words. Capture the main points. Write succinctly.
	Do not include any commentary other than the summary itself.
`

func (s *SearchAgent) Search(ctx context.Context, item models.WebSearchItem) (*models.SearchResult, error) {
	ctx, span := s.obs.Tracer.Start(ctx, "SearchAgent.Search")
	defer span.End()

	s.obs.Logger.Info("creating searches",
		zap.String("query", item.Query),
		zap.Any("item", item),
	)

	searchResult, err := s.searchClient.Search(ctx, item.Query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.obs.Logger.Error("search failed",
			zap.String("query", item.Query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	userPrompt := fmt.Sprintf(`
		Search term: %s
		Reason for searching: %s
		Raw search results: %s
	`,
		item.Query, item.Reason, searchResult,
	)

	summary, err := s.provider.Generate(ctx, searcherSystemPrompt, userPrompt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.obs.Logger.Error("search summarization failed",
			zap.String("query", item.Query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search summarization failed: %w", err)
	}

	s.obs.Logger.Info("search results created successfully",
		zap.String("query", item.Query),
	)

	return &models.SearchResult{
		Query:   item.Query,
		Summary: summary,
	}, nil

}
