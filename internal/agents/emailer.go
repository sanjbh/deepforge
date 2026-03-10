package agents

import (
	"context"
	"fmt"

	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
	"github.com/sanjbh/deepforge/internal/tools"
	"github.com/sanjbh/deepforge/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type EmailAgent struct {
	provider llm.Provider
	sender   tools.EmailSender
	obs      *observability.Observability
}

func NewEmailAgent(provider llm.Provider, sender tools.EmailSender, obs *observability.Observability) *EmailAgent {
	return &EmailAgent{provider: provider, sender: sender, obs: obs}
}

const emailerSystemPrompt = `
	You are able to convert a markdown research report into clean, 
	well presented HTML. Return only the HTML body content, 
	no additional commentary.
`

func (e *EmailAgent) Send(ctx context.Context, report *models.ReportData) error {
	ctx, span := e.obs.Tracer.Start(ctx, "EmailAgent.Send")
	defer span.End()

	e.obs.Logger.Info(
		"sending email",
		zap.Any("short_summary", report.ShortSummary),
	)

	userPrompt := fmt.Sprintf(
		`Please convert the following markdown report into clean, well presented HTML.\n
				Return only the HTML body content, no additional commentary.\n\n%s`,
		report.MarkdownReport,
	)
	htmlBody, err := e.provider.Generate(ctx, emailerSystemPrompt, userPrompt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		e.obs.Logger.Error("failed to generate HTML body using LLM", zap.Error(err))
		return fmt.Errorf("failed to generate HTML body: %w", err)
	}

	subjectLine := fmt.Sprintf("Research Report: %s", report.ShortSummary)

	if err := e.sender.Send(subjectLine, htmlBody); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		e.obs.Logger.Error("failed to send email", zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.obs.Logger.Info("sent email", zap.String("subject", subjectLine))
	return nil
}
