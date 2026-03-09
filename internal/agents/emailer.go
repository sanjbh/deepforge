package tools

import (
	"context"
	"fmt"

	"github.com/sanjbh/deepforge/internal/llm"
	"github.com/sanjbh/deepforge/internal/models"
)

type EmailAgent struct {
	provider llm.Provider
	sender   *EmailSender
}

func NewEmailAgent(provider llm.Provider, sender *EmailSender) *EmailAgent {
	return &EmailAgent{provider: provider, sender: sender}
}

const emailerSystemPrompt = `
	You are able to convert a markdown research report into clean, 
	well presented HTML. Return only the HTML body content, 
	no additional commentary.
`

func (e *EmailAgent) Send(ctx context.Context, report models.ReportData) error {
	userPrompt := fmt.Sprintf(
		`Please convert the following markdown report into clean, well presented HTML.\n
				Return only the HTML body content, no additional commentary.\n\n%s`,
		report.MarkdownReport,
	)
	htmlBody, err := e.provider.Generate(ctx, emailerSystemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate HTML body: %w", err)
	}

	if err := e.sender.Send(report.ShortSummary, htmlBody); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
