package models

import "fmt"

// WebSearchItem is one planned search.
type WebSearchItem struct {
	Reason string `json:"reason"`
	Query  string `json:"query"`
}

// WebSearchPlan is the full output of the PlannerAgent.
type WebSearchPlan struct {
	Searches []WebSearchItem `json:"searches"`
}

// ReportData is the full output of the WriterAgent.
type ReportData struct {
	ShortSummary      string   `json:"short_summary"`
	MarkdownReport    string   `json:"mark_down_report"`
	FollowUpQuestions []string `json:"follow_up_questions"`
}

// SearchResult carries one search's output through the pipeline.
// The Query field gives the WriterAgent richer context than
// a raw string slice would.
type SearchResult struct {
	Query   string `json:"query"`
	Summary string `json:"summary"`
}

// AgentError wraps errors with the agent name for clean observability.
type AgentError struct {
	Agent   string
	Message string
	Err     error
}

func (e *AgentError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Agent, e.Message, e.Err)
}

func (e *AgentError) Unwrap() error {
	return e.Err
}
