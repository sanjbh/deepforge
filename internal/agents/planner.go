package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
)

const plannerSystemPromptTemplate = `
	You are a helpful research assistant. Given a query, come up with a set of web searches to perform to best answer the query. 
	Output %d search terms to query for.
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
	userPrompt := fmt.Sprintf("Query: %s, Generate %d searches", query, p.howManySearches)

	plannerSystemPrompt := fmt.Sprintf(plannerSystemPromptTemplate, p.howManySearches)
	plan, err := p.provider.GenerateStructured(ctx, plannerSystemPrompt, userPrompt, schema)
	if err != nil {
		return nil, fmt.Errorf("planner failed to generate search plan: %w", err)
	}

	var webSearchPlan models.WebSearchPlan
	err = json.Unmarshal([]byte(plan), &webSearchPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal web search plan: %w", err)
	}

	return &webSearchPlan, nil
}
