package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
)

type WriterAgent struct {
	provider llm.Provider
}

func NewWriterAgent(provider llm.Provider) *WriterAgent {
	return &WriterAgent{
		provider: provider,
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
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	var reportData models.ReportData
	if err = json.Unmarshal([]byte(result), &reportData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &reportData, nil
}
