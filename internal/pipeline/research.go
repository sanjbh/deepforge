package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/sanjbh/deepforge/internal/agents"
	"github.com/sanjbh/deepforge/internal/models"
	"golang.org/x/sync/errgroup"
)

type ResearcherPipeline struct {
	planner  *agents.PlannerAgent
	searcher *agents.SearchAgent
	writer   *agents.WriterAgent
	emailer  *agents.EmailAgent
}

func NewResearcherPipeline(
	planner *agents.PlannerAgent,
	searcher *agents.SearchAgent,
	writer *agents.WriterAgent,
	emailer *agents.EmailAgent) *ResearcherPipeline {

	return &ResearcherPipeline{planner: planner, searcher: searcher, writer: writer, emailer: emailer}
}

func (p *ResearcherPipeline) Run(ctx context.Context, query string) error {
	planner, err := p.planner.Plan(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get planner: %w", err)
	}

	log.Println("Created planner successfully")

	g, gCtx := errgroup.WithContext(ctx)
	results := make([]models.SearchResult, len(planner.Searches))

	for i, item := range planner.Searches {
		i, item := i, item
		g.Go(func() error {
			result, err := p.searcher.Search(gCtx, item)
			if err != nil {
				return fmt.Errorf("failed to search: %w", err)
			}
			results[i] = *result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to get results: %w", err)
	}

	log.Println("Created searches successfully")

	reportData, err := p.writer.Write(ctx, query, results)
	if err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}
	log.Println("Created reports successfully")

	if err := p.emailer.Send(ctx, reportData); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Println("Sent email successfully")
	return nil
}
