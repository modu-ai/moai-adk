package telemetry

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// underutilizedThreshold is the minimum number of uses required for a skill
// to be considered "utilized". Skills below this count are reported as underutilized.
const underutilizedThreshold = 3

// GenerateReport reads all telemetry JSONL files within the past `days` days
// and produces an aggregated Report.
//
// Returns an empty (non-nil) report when no telemetry data exists.
func GenerateReport(projectRoot string, days int) (*Report, error) {
	telDir := filepath.Join(projectRoot, ".moai", "evolution", "telemetry")

	entries, err := os.ReadDir(telDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Report{Days: days}, nil
		}
		return nil, fmt.Errorf("report: readdir: %w", err)
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -days)

	// Collect all records within the time window.
	var records []UsageRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "usage-") || !strings.HasSuffix(name, ".jsonl") {
			continue
		}

		dateStr := strings.TrimPrefix(name, "usage-")
		dateStr = strings.TrimSuffix(dateStr, ".jsonl")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		// Skip files that are entirely outside the window.
		if fileDate.Before(cutoff.Truncate(24 * time.Hour)) {
			continue
		}

		recs, err := readJSONLFile(filepath.Join(telDir, name))
		if err != nil {
			// Best effort: skip unreadable files
			continue
		}
		for _, r := range recs {
			if !r.Timestamp.Before(cutoff) {
				records = append(records, r)
			}
		}
	}

	return buildReport(days, records), nil
}

// buildReport constructs a Report from a flat slice of UsageRecords.
func buildReport(days int, records []UsageRecord) *Report {
	// Aggregate by skill_id.
	statsMap := make(map[string]*SkillStats)
	// sessionSkills maps session_id → set of skill_ids used in that session.
	sessionSkills := make(map[string]map[string]bool)

	for _, r := range records {
		if _, ok := statsMap[r.SkillID]; !ok {
			statsMap[r.SkillID] = &SkillStats{SkillID: r.SkillID}
		}
		s := statsMap[r.SkillID]
		s.Uses++
		switch r.Outcome {
		case OutcomeSuccess:
			s.Success++
		case OutcomePartial:
			s.Partial++
		case OutcomeError:
			s.Error++
		default:
			s.Unknown++
		}

		// Track co-occurrence via session.
		if r.SessionID != "" {
			if sessionSkills[r.SessionID] == nil {
				sessionSkills[r.SessionID] = make(map[string]bool)
			}
			sessionSkills[r.SessionID][r.SkillID] = true
		}
	}

	// Sort skills by uses descending for stable output.
	skills := make([]SkillStats, 0, len(statsMap))
	for _, s := range statsMap {
		skills = append(skills, *s)
	}
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Uses != skills[j].Uses {
			return skills[i].Uses > skills[j].Uses
		}
		return skills[i].SkillID < skills[j].SkillID
	})

	// Identify underutilized skills.
	var underutilized []UnderutilizedSkill
	for _, s := range skills {
		if s.Uses < underutilizedThreshold {
			underutilized = append(underutilized, UnderutilizedSkill{
				SkillID: s.SkillID,
				Uses:    s.Uses,
			})
		}
	}

	// Compute co-occurrence pairs.
	coMap := make(map[string]int) // "skillA|skillB" → count
	for _, skillSet := range sessionSkills {
		// Collect sorted skill IDs for canonical pair ordering.
		ids := make([]string, 0, len(skillSet))
		for id := range skillSet {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				key := ids[i] + "|" + ids[j]
				coMap[key]++
			}
		}
	}

	coOccurrences := make([]CoOccurrence, 0, len(coMap))
	for key, count := range coMap {
		parts := strings.SplitN(key, "|", 2)
		if len(parts) == 2 {
			coOccurrences = append(coOccurrences, CoOccurrence{
				SkillA: parts[0],
				SkillB: parts[1],
				Count:  count,
			})
		}
	}
	sort.Slice(coOccurrences, func(i, j int) bool {
		if coOccurrences[i].Count != coOccurrences[j].Count {
			return coOccurrences[i].Count > coOccurrences[j].Count
		}
		return coOccurrences[i].SkillA < coOccurrences[j].SkillA
	})

	return &Report{
		Days:          days,
		Skills:        skills,
		CoOccurrences: coOccurrences,
		Underutilized: underutilized,
	}
}

// String returns a human-readable text representation of the report.
func (r *Report) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Skill Usage Report (last %d days)\n", r.Days)
	sb.WriteString("=====================================\n")

	if len(r.Skills) == 0 {
		sb.WriteString("No telemetry data available.\n")
		return sb.String()
	}

	// Header
	fmt.Fprintf(&sb, "%-32s | %5s | %7s | %7s | %5s\n",
		"Skill", "Uses", "Success", "Partial", "Error")
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	for _, s := range r.Skills {
		fmt.Fprintf(&sb, "%-32s | %5d | %7d | %7d | %5d\n",
			s.SkillID, s.Uses, s.Success, s.Partial, s.Error)
	}

	if len(r.CoOccurrences) > 0 {
		sb.WriteString("\nTop Co-occurring Skills:\n")
		limit := 10
		if len(r.CoOccurrences) < limit {
			limit = len(r.CoOccurrences)
		}
		for _, co := range r.CoOccurrences[:limit] {
			fmt.Fprintf(&sb, "  %s + %s: %d times\n", co.SkillA, co.SkillB, co.Count)
		}
	}

	if len(r.Underutilized) > 0 {
		sb.WriteString("\nUnderutilized Skills (loaded but rarely invoked):\n")
		for _, u := range r.Underutilized {
			fmt.Fprintf(&sb, "  %s: %d uses in %d days\n", u.SkillID, u.Uses, r.Days)
		}
	}

	return sb.String()
}

// readJSONLFile reads all UsageRecords from a JSONL file.
// Lines that cannot be parsed are skipped (best effort).
func readJSONLFile(path string) ([]UsageRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("report: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var records []UsageRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var r UsageRecord
		if err := json.Unmarshal(line, &r); err != nil {
			// Skip malformed lines — best effort
			continue
		}
		records = append(records, r)
	}
	return records, scanner.Err()
}
