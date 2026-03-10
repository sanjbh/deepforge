package pipeline

import (
	"context"
	"fmt"

	"github.com/sanjbh/deepforge/internal/agents"
	"github.com/sanjbh/deepforge/internal/models"
	"github.com/sanjbh/deepforge/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ResearcherPipeline struct {
	planner  *agents.PlannerAgent
	searcher *agents.SearchAgent
	writer   *agents.WriterAgent
	emailer  *agents.EmailAgent
	obs      *observability.Observability
}

func NewResearcherPipeline(
	planner *agents.PlannerAgent,
	searcher *agents.SearchAgent,
	writer *agents.WriterAgent,
	emailer *agents.EmailAgent,
	obs *observability.Observability,
) *ResearcherPipeline {

	return &ResearcherPipeline{planner: planner, searcher: searcher, writer: writer, emailer: emailer, obs: obs}
}

func (p *ResearcherPipeline) Run(ctx context.Context, query string) error {
	ctx, span := p.obs.Tracer.Start(ctx, "ResearchPipeline.Run")
	defer span.End()

	p.obs.Logger.Info("starting research pipeline",
		zap.String("query", query),
	)

	planner, err := p.planner.Plan(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.obs.Logger.Error("failed to get planner",
			zap.Any("query", query),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get planner: %w", err)
	}

	p.obs.Logger.Info("planner created", zap.String("query", query))

	g, gCtx := errgroup.WithContext(ctx)
	results := make([]models.SearchResult, len(planner.Searches))

	for i, item := range planner.Searches {
		i, item := i, item
		g.Go(func() error {
			result, err := p.searcher.Search(gCtx, item)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				p.obs.Logger.Error("failed to search", zap.String("query", query), zap.Error(err))
				return fmt.Errorf("failed to search: %w", err)
			}
			results[i] = *result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.obs.Logger.Error("failed to get results", zap.Error(err))
		return fmt.Errorf("failed to get results: %w", err)
	}

	p.obs.Logger.Info("created searches successfully", zap.String("query", query))

	reportData, err := p.writer.Write(ctx, query, results)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.obs.Logger.Error("failed to write result", zap.String("query", query), zap.Error(err))
		return fmt.Errorf("failed to write results: %w", err)
	}

	p.obs.Logger.Info("created reports successfully", zap.Any("short_summary", reportData.ShortSummary))

	if err := p.emailer.Send(ctx, reportData); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.obs.Logger.Error("failed to send email", zap.Any("short_summary", reportData.ShortSummary), zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}
	p.obs.Logger.Info("sent email successfully")
	return nil
}
