package agents

import (
	"context"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
)

const systemPrompt = `
	You are a helpful research assistant. Given a query, come up with a set of web searches to perform to best answer the query. 
	Output {N} search terms to query for.
`

type PlannerAgent struct {
	provider        llm.Provider
	howManySearches int
}

func NewPlannerAgent(provider llm.Provider, howManySearches int) *PlannerAgent {
	return &PlannerAgent{
		provider:        provider,
		howManySearches: howManySearches,
	}
}

func (p *PlannerAgent) Plan(ctx context.Context, query string) (*models.WebSearchPlan, error) {
	schema := jsonschema.Reflect(&models.WebSearchPlan{})
	userPrompt := fmt.Sprintf("Query: %s, Generate %s searches", query, p.howManySearches)

	plan, err := p.provider.GenerateStructured(ctx, systemPrompt, userPrompt, schema)
	if err != nil {
		return nil, err
	}

	return jsonschema.Unmarshal(plan, schema)
}
