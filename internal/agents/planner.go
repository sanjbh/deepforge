package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
	"github.com/sanjbh/deepforge/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

const plannerSystemPromptTemplate = `
	You are a helpful research assistant. Given a topic, generate exactly %d web search queries to thoroughly research it.

	You MUST respond with valid JSON only. No explanation, no markdown, no code blocks.

	Each search item MUST have:
	- "reason": why this search is important
	- "query": the exact non-empty search term to use (this field must never be empty)

	Example output:
	{
	"searches": [
		{"reason": "Get an overview of the topic", "query": "async Rust programming basics"},
		{"reason": "Find advanced patterns", "query": "async Rust tokio advanced patterns 2025"}
	]
	}
`

type PlannerAgent struct {
	provider        llm.Provider
	howManySearches int
	obs             *observability.Observability
}

func NewPlannerAgent(provider llm.Provider, howManySearches int, obs *observability.Observability) *PlannerAgent {
	return &PlannerAgent{
		provider:        provider,
		howManySearches: howManySearches,
		obs:             obs,
	}
}

func (p *PlannerAgent) Plan(ctx context.Context, query string) (*models.WebSearchPlan, error) {
	ctx, span := p.obs.Tracer.Start(ctx, "PlannerAgent.Plan")
	defer span.End()

	p.obs.Logger.Info("planning searches",
		zap.String("query", query),
		zap.Int("howManySearches", p.howManySearches),
	)

	schema := jsonschema.Reflect(&models.WebSearchPlan{})
	userPrompt := fmt.Sprintf("Query: %s, Generate %d searches", query, p.howManySearches)

	plannerSystemPrompt := fmt.Sprintf(plannerSystemPromptTemplate, p.howManySearches)
	plan, err := p.provider.GenerateStructured(ctx, plannerSystemPrompt, userPrompt, schema)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.obs.Logger.Error("failed to plan searches",
			zap.String("query", query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("planner failed to generate search plan: %w", err)
	}

	var webSearchPlan models.WebSearchPlan
	err = json.Unmarshal([]byte(plan), &webSearchPlan)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.obs.Logger.Error("unmarshal web search plan",
			zap.Any("query", query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal web search plan: %w", err)
	}

	p.obs.Logger.Info("search plan created",
		zap.Int("num_searches", len(webSearchPlan.Searches)),
	)
	return &webSearchPlan, nil
}
