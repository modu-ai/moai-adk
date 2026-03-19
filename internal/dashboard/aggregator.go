package dashboard

import (
	"sort"
	"time"
)

// Aggregator computes statistics from raw task metrics.
type Aggregator struct{}

// NewAggregator creates a new metrics aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// ComputeSPECCosts groups metrics by SPEC ID and computes costs.
func (a *Aggregator) ComputeSPECCosts(metrics []TaskMetric) []SPECCost {
	specMap := make(map[string]*SPECCost)

	for _, m := range metrics {
		if m.SpecID == "" {
			continue
		}
		sc, ok := specMap[m.SpecID]
		if !ok {
			sc = &SPECCost{SpecID: m.SpecID}
			specMap[m.SpecID] = sc
		}
		sc.TotalTokens += m.InputTokens + m.OutputTokens
		sc.InputTokens += m.InputTokens
		sc.OutputTokens += m.OutputTokens
		sc.Duration += time.Duration(m.DurationMS) * time.Millisecond
		sc.AgentCalls++

		// Track phase costs
		found := false
		for i := range sc.Phases {
			if sc.Phases[i].Phase == m.Phase {
				sc.Phases[i].Tokens += m.InputTokens + m.OutputTokens
				sc.Phases[i].Duration += time.Duration(m.DurationMS) * time.Millisecond
				sc.Phases[i].AgentCalls++
				found = true
				break
			}
		}
		if !found {
			sc.Phases = append(sc.Phases, PhaseCost{
				Phase:      m.Phase,
				Tokens:     m.InputTokens + m.OutputTokens,
				Duration:   time.Duration(m.DurationMS) * time.Millisecond,
				AgentCalls: 1,
			})
		}
	}

	var result []SPECCost
	for _, sc := range specMap {
		result = append(result, *sc)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalTokens > result[j].TotalTokens
	})
	return result
}

// ComputeAgentStats computes per-agent performance statistics.
func (a *Aggregator) ComputeAgentStats(metrics []TaskMetric) []AgentStats {
	agentMap := make(map[string]*agentAccumulator)

	for _, m := range metrics {
		if m.AgentName == "" {
			continue
		}
		acc, ok := agentMap[m.AgentName]
		if !ok {
			acc = &agentAccumulator{name: m.AgentName}
			agentMap[m.AgentName] = acc
		}
		acc.totalCalls++
		if m.Success {
			acc.successCount++
		} else {
			acc.failureCount++
		}
		acc.totalTokens += m.InputTokens + m.OutputTokens
		acc.totalDurationMS += m.DurationMS
	}

	var result []AgentStats
	for _, acc := range agentMap {
		stats := AgentStats{
			AgentName:    acc.name,
			TotalCalls:   acc.totalCalls,
			SuccessCount: acc.successCount,
			FailureCount: acc.failureCount,
		}
		if acc.totalCalls > 0 {
			stats.SuccessRate = float64(acc.successCount) / float64(acc.totalCalls) * 100
			stats.AvgTokens = acc.totalTokens / int64(acc.totalCalls)
			stats.AvgDuration = float64(acc.totalDurationMS) / float64(acc.totalCalls) / 1000.0
		}
		result = append(result, stats)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalCalls > result[j].TotalCalls
	})
	return result
}

// ComputeTrends analyzes quality and cost trends.
func (a *Aggregator) ComputeTrends(metrics []TaskMetric, period string) *TrendReport {
	if len(metrics) == 0 {
		return &TrendReport{Period: period}
	}

	// Sort by timestamp
	sorted := make([]TaskMetric, len(metrics))
	copy(sorted, metrics)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	report := &TrendReport{
		Period:    period,
		StartDate: sorted[0].Timestamp,
		EndDate:   sorted[len(sorted)-1].Timestamp,
	}

	// Count unique SPECs
	specSet := make(map[string]bool)
	for _, m := range sorted {
		if m.SpecID != "" {
			specSet[m.SpecID] = true
		}
	}
	report.SPECsCreated = len(specSet)

	// Compute success rate as quality proxy
	successCount := 0
	for _, m := range sorted {
		if m.Success {
			successCount++
		}
	}
	if len(sorted) > 0 {
		report.AvgQuality = float64(successCount) / float64(len(sorted)) * 100
	}

	// Determine trends (simple: compare first half vs second half)
	mid := len(sorted) / 2
	if mid > 0 {
		firstHalfSuccess := countSuccess(sorted[:mid])
		secondHalfSuccess := countSuccess(sorted[mid:])
		firstRate := float64(firstHalfSuccess) / float64(mid)
		secondRate := float64(secondHalfSuccess) / float64(len(sorted)-mid)

		if secondRate > firstRate+0.05 {
			report.QualityTrend = "improving"
		} else if secondRate < firstRate-0.05 {
			report.QualityTrend = "declining"
		} else {
			report.QualityTrend = "stable"
		}
	}

	return report
}

// ComputeBudgetStatus calculates budget utilization.
func (a *Aggregator) ComputeBudgetStatus(metrics []TaskMetric, totalBudget int64, alertPct float64) BudgetStatus {
	var usedTokens int64
	for _, m := range metrics {
		usedTokens += m.InputTokens + m.OutputTokens
	}

	remaining := totalBudget - usedTokens
	if remaining < 0 {
		remaining = 0
	}

	usagePct := 0.0
	if totalBudget > 0 {
		usagePct = float64(usedTokens) / float64(totalBudget) * 100
	}

	return BudgetStatus{
		TotalBudget:     totalBudget,
		UsedTokens:      usedTokens,
		RemainingTokens: remaining,
		UsagePercent:    usagePct,
		AlertThreshold:  alertPct,
		IsAlert:         usagePct >= alertPct,
	}
}

type agentAccumulator struct {
	name            string
	totalCalls      int
	successCount    int
	failureCount    int
	totalTokens     int64
	totalDurationMS int64
}

func countSuccess(metrics []TaskMetric) int {
	count := 0
	for _, m := range metrics {
		if m.Success {
			count++
		}
	}
	return count
}
