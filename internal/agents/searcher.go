package agents

import (
	"context"
	"fmt"

	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
	"github.com/sanjbh/deepforge/internal/search"
)

type SearchAgent struct {
	provider     llm.Provider
	searchClient *search.DuckDuckGoClient
}

func NewSearchAgent(provider llm.Provider, searchClient *search.DuckDuckGoClient) *SearchAgent {
	return &SearchAgent{
		provider:     provider,
		searchClient: searchClient,
	}
}

const searcherSystemPrompt = `
	You are a research assistant. Given a search term and raw search results,
	produce a concise summary of the results. The summary must be 2-3 paragraphs 
	and less than 300 words. Capture the main points. Write succinctly.
	Do not include any commentary other than the summary itself.
`

func (s *SearchAgent) Search(ctx context.Context, item models.WebSearchItem) (*models.SearchResult, error) {
	searchResult, err := s.searchClient.Search(ctx, item.Query)
	if err != nil {
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
		return nil, fmt.Errorf("search summarisation failed: %w", err)
	}

	return &models.SearchResult{
		Query:   item.Query,
		Summary: summary,
	}, nil

}
