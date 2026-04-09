package observe

import "sort"

// PatternThresholds defines observation count thresholds for pattern classification.
type PatternThresholds struct {
	// Heuristic is the minimum observation count required for heuristic classification.
	Heuristic int
	// Rule is the minimum observation count required for rule classification.
	Rule int
	// HighConfidence is the minimum observation count required for high-confidence classification.
	HighConfidence int
}

// DefaultThresholds returns the default pattern classification thresholds.
func DefaultThresholds() PatternThresholds {
	return PatternThresholds{
		Heuristic:      3,
		Rule:           5,
		HighConfidence: 10,
	}
}

// PatternDetector detects repeating patterns from a list of observations.
type PatternDetector struct {
	thresholds PatternThresholds
}

// NewPatternDetector creates a pattern detector with the specified thresholds.
func NewPatternDetector(thresholds PatternThresholds) *PatternDetector {
	return &PatternDetector{thresholds: thresholds}
}

// Detect analyzes a list of observations to detect patterns and returns them sorted by count descending.
// Observations are grouped by Agent:Target key, and each group is classified based on its observation count.
func (d *PatternDetector) Detect(observations []*Observation) []*Pattern {
	if len(observations) == 0 {
		return nil
	}

	// Group by Agent:Target key
	groups := make(map[string][]*Observation)
	order := make([]string, 0) // preserve insertion order for stable sort
	for _, obs := range observations {
		key := obs.Agent + ":" + obs.Target
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], obs)
	}

	// Build patterns per group
	patterns := make([]*Pattern, 0, len(groups))
	for _, key := range order {
		obs := groups[key]
		p := &Pattern{
			Key:          key,
			Count:        len(obs),
			Observations: obs,
		}

		// Compute FirstSeen and LastSeen
		p.FirstSeen = obs[0].Timestamp
		p.LastSeen = obs[0].Timestamp
		for _, o := range obs[1:] {
			if o.Timestamp.Before(p.FirstSeen) {
				p.FirstSeen = o.Timestamp
			}
			if o.Timestamp.After(p.LastSeen) {
				p.LastSeen = o.Timestamp
			}
		}

		// Threshold-based classification
		p.Classification = d.classify(p.Count)

		patterns = append(patterns, p)
	}

	// Sort by count descending (stable sort preserves insertion order for equal counts)
	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}

// classify returns the pattern classification based on observation count.
func (d *PatternDetector) classify(count int) PatternClassification {
	switch {
	case count >= d.thresholds.HighConfidence:
		return ClassHighConfidence
	case count >= d.thresholds.Rule:
		return ClassRule
	case count >= d.thresholds.Heuristic:
		return ClassHeuristic
	default:
		return ClassObservation
	}
}
