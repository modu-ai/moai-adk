package worktree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SuspectFlag indicates the orchestrator's response showed an empty worktreePath
// despite an isolation: "worktree" request.
type SuspectFlag struct {
	SnapshotID  string    `json:"snapshot_id"`
	AgentName   string    `json:"agent_name"`
	Reason      string    `json:"reason"`
	DetectedAt  time.Time `json:"detected_at"`
	PushBlocked bool      `json:"push_blocked"`
}

// SuspectFlag reason codes.
const (
	SuspectReasonEmptyWorktreePath = "empty_worktree_path"
	SuspectReasonStateDivergence   = "post_state_divergence"
)

// DivergenceLogEntry is a single divergence record written to the daily report.
type DivergenceLogEntry struct {
	Timestamp   time.Time    `json:"timestamp"`
	SnapshotID  string       `json:"snapshot_id"`
	AgentName   string       `json:"agent_name"`
	Divergence  Divergence   `json:"divergence"`
	SuspectFlag *SuspectFlag `json:"suspect_flag,omitempty"`
}

// AppendDivergenceLog appends a divergence entry to the daily markdown report
// at .moai/reports/worktree-guard/<YYYY-MM-DD>.md and writes a JSON sidecar
// at .moai/reports/worktree-guard/<YYYY-MM-DD>-<snapshot_id>.json.
func AppendDivergenceLog(projectRoot string, entry DivergenceLogEntry) (mdPath, jsonPath string, err error) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	dateStr := entry.Timestamp.Format("2006-01-02")
	reportDir := filepath.Join(projectRoot, DivergenceReportDir)
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir %s: %w", reportDir, err)
	}
	mdPath = filepath.Join(reportDir, dateStr+".md")
	jsonPath = filepath.Join(reportDir, fmt.Sprintf("%s-%s.json", dateStr, entry.SnapshotID))

	md := buildMarkdown(entry)
	f, err := os.OpenFile(mdPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", "", fmt.Errorf("open %s: %w", mdPath, err)
	}
	if _, werr := f.WriteString(md); werr != nil {
		_ = f.Close()
		return "", "", fmt.Errorf("write %s: %w", mdPath, werr)
	}
	if cerr := f.Close(); cerr != nil {
		return "", "", fmt.Errorf("close %s: %w", mdPath, cerr)
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal entry: %w", err)
	}
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", jsonPath, err)
	}

	return mdPath, jsonPath, nil
}

// WriteSuspectFlag writes a suspect flag file for the given snapshot ID.
func WriteSuspectFlag(projectRoot string, flag SuspectFlag) (string, error) {
	if flag.DetectedAt.IsZero() {
		flag.DetectedAt = time.Now().UTC()
	}
	flagDir := filepath.Join(projectRoot, StateDirRel)
	if err := os.MkdirAll(flagDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", flagDir, err)
	}
	flagPath := filepath.Join(projectRoot, fmt.Sprintf(SuspectFlagTemplate, flag.SnapshotID))
	data, err := json.MarshalIndent(flag, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal flag: %w", err)
	}
	if err := os.WriteFile(flagPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", flagPath, err)
	}
	return flagPath, nil
}

func buildMarkdown(e DivergenceLogEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s — Divergence detected\n\n", e.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Snapshot ID: %s\n", e.SnapshotID)
	fmt.Fprintf(&b, "- Agent: %s\n", coalesce(e.AgentName, "(unknown)"))
	fmt.Fprintf(&b, "- HeadChanged: %v", e.Divergence.HeadChanged)
	if e.Divergence.HeadChanged {
		fmt.Fprintf(&b, " (pre=%s -> post=%s)", short(e.Divergence.PreHeadSHA), short(e.Divergence.PostHeadSHA))
	}
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- BranchChanged: %v", e.Divergence.BranchChanged)
	if e.Divergence.BranchChanged {
		fmt.Fprintf(&b, " (pre=%s -> post=%s)", e.Divergence.PreBranch, e.Divergence.PostBranch)
	}
	fmt.Fprintln(&b)
	if len(e.Divergence.UntrackedAdded) > 0 {
		fmt.Fprintln(&b, "- UntrackedAdded:")
		for _, p := range e.Divergence.UntrackedAdded {
			fmt.Fprintf(&b, "  - %s\n", p)
		}
	}
	if len(e.Divergence.UntrackedRemoved) > 0 {
		fmt.Fprintln(&b, "- UntrackedRemoved:")
		for _, p := range e.Divergence.UntrackedRemoved {
			fmt.Fprintf(&b, "  - %s\n", p)
		}
	}
	if e.SuspectFlag != nil {
		fmt.Fprintf(&b, "- SuspectFlag: %s (push_blocked=%v)\n", e.SuspectFlag.Reason, e.SuspectFlag.PushBlocked)
	}
	fmt.Fprintln(&b)
	return b.String()
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func coalesce(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
