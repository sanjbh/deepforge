package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
	"github.com/sanjbh/deepforge/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type WriterAgent struct {
	provider llm.Provider
	obs      *observability.Observability
}

func NewWriterAgent(provider llm.Provider, obs *observability.Observability) *WriterAgent {
	return &WriterAgent{
		provider: provider,
		obs:      obs,
	}
}

const writerSystemPrompt = `
	You are a senior researcher tasked with writing a cohesive report for a 
	research query. You will be provided with the original query and summarised 
	search results from a research assistant. First come up with an outline, 
	then generate the report. The final output should be in markdown format, 
	lengthy and detailed. Aim for at least 1000 words.
`

func (w *WriterAgent) Write(ctx context.Context, query string, results []models.SearchResult) (*models.ReportData, error) {
	ctx, span := w.obs.Tracer.Start(ctx, "WriterAgent.Write")
	defer span.End()

	w.obs.Logger.Info("generating reports",
		zap.String("query", query),
		zap.Any("item", results),
	)

	schema := jsonschema.Reflect(&models.ReportData{})
	var writerUserPrompt strings.Builder

	fmt.Fprintf(&writerUserPrompt, "Original query: %s\n\n", query)
	fmt.Fprintf(&writerUserPrompt, "Summarized search results:\n\n")
	for _, sr := range results {
		fmt.Fprintf(
			&writerUserPrompt,
			"Search %s\nSummary: %s\n\n",
			sr.Query, sr.Summary,
		)
	}

	result, err := w.provider.GenerateStructured(ctx, writerSystemPrompt, writerUserPrompt.String(), schema)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		w.obs.Logger.Error("failed to generate report",
			zap.String("query", query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	var reportData models.ReportData
	if err = json.Unmarshal([]byte(result), &reportData); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		w.obs.Logger.Error("failed to unmarshal report",
			zap.Any("short_summary", reportData.ShortSummary),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	w.obs.Logger.Info("report created successfully",
		zap.String("query", query),
	)

	return &reportData, nil
}
