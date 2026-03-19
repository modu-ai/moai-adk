// Package dashboard provides observability and cost tracking for MoAI-ADK
// SPEC workflows, agent performance, and quality trends.
package dashboard

import "time"

// SPECCost tracks the cost metrics for a single SPEC implementation.
type SPECCost struct {
	SpecID       string        `json:"spec_id"`
	TotalTokens  int64         `json:"total_tokens"`
	InputTokens  int64         `json:"input_tokens"`
	OutputTokens int64         `json:"output_tokens"`
	Duration     time.Duration `json:"duration"`
	AgentCalls   int           `json:"agent_calls"`
	Phases       []PhaseCost   `json:"phases"`
}

// PhaseCost tracks costs for a single workflow phase.
type PhaseCost struct {
	Phase        string        `json:"phase"` // "plan", "challenge", "run", "sync"
	Tokens       int64         `json:"tokens"`
	Duration     time.Duration `json:"duration"`
	AgentCalls   int           `json:"agent_calls"`
}

// AgentStats tracks performance statistics for an agent type.
type AgentStats struct {
	AgentName    string  `json:"agent_name"`
	TotalCalls   int     `json:"total_calls"`
	SuccessCount int     `json:"success_count"`
	FailureCount int     `json:"failure_count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgTokens    int64   `json:"avg_tokens"`
	AvgDuration  float64 `json:"avg_duration_seconds"`
}

// TrendReport summarizes quality trends over time.
type TrendReport struct {
	Period       string        `json:"period"` // "daily", "weekly", "monthly"
	StartDate    time.Time     `json:"start_date"`
	EndDate      time.Time     `json:"end_date"`
	SPECsCreated int           `json:"specs_created"`
	SPECsCompleted int         `json:"specs_completed"`
	AvgQuality   float64       `json:"avg_quality_score"`
	QualityTrend string        `json:"quality_trend"` // "improving", "stable", "declining"
	CostTrend    string        `json:"cost_trend"`
	Entries      []TrendEntry  `json:"entries"`
}

// TrendEntry represents a single data point in a trend.
type TrendEntry struct {
	Date         time.Time `json:"date"`
	SPECCount    int       `json:"spec_count"`
	TotalTokens  int64     `json:"total_tokens"`
	QualityScore float64   `json:"quality_score"`
}

// BudgetStatus represents the current budget utilization.
type BudgetStatus struct {
	TotalBudget    int64   `json:"total_budget"`
	UsedTokens     int64   `json:"used_tokens"`
	RemainingTokens int64  `json:"remaining_tokens"`
	UsagePercent   float64 `json:"usage_percent"`
	AlertThreshold float64 `json:"alert_threshold"`
	IsAlert        bool    `json:"is_alert"`
}

// TaskMetric represents a single task metric entry from task-metrics.jsonl.
type TaskMetric struct {
	Timestamp    time.Time `json:"timestamp"`
	SessionID    string    `json:"session_id"`
	SpecID       string    `json:"spec_id,omitempty"`
	AgentName    string    `json:"agent_name"`
	Phase        string    `json:"phase"`
	Action       string    `json:"action"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	DurationMS   int64     `json:"duration_ms"`
	Success      bool      `json:"success"`
	ErrorMsg     string    `json:"error_msg,omitempty"`
}

// DashboardSummary is the top-level dashboard view.
type DashboardSummary struct {
	ActiveSPECs    int          `json:"active_specs"`
	CompletedSPECs int          `json:"completed_specs"`
	TotalTokensUsed int64       `json:"total_tokens_used"`
	TopAgents      []AgentStats `json:"top_agents"`
	Budget         BudgetStatus `json:"budget"`
	RecentActivity []TaskMetric `json:"recent_activity"`
}
