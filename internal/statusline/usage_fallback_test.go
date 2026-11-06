package statusline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SPEC-STATUSLINE-OFFICIAL-USAGE-001 — builder-level behavioral tests for the
// local context-usage snapshot fallback (plan M2/M4; AC-SOU-003, AC-SOU-004,
// AC-SOU-009, AC-SOU-010, AC-SOU-005). Every test feeds GLM-form stdin (no
// rate_limits) or Claude-form stdin through the real Builder.Build pipeline
// with the state anchor pointed at fixtures built under t.TempDir, so no test
// touches the host's records. Records are seeded through the real
// SessionTelemetryRecord schema (plan §F R3 — ad-hoc JSON would read as
// unparseable and falsely satisfy the n/a cases).

// quotaTestSessionIDs keep the seeded records distinct from the window-less
// record Build itself writes at the anchor on every render.
const quotaTestSessionID = "seed-primary"

// seedQuotaRecord writes one schema-v3 session telemetry record under
// projectRoot's state directory, the shape the quota aggregator reads.
func seedQuotaRecord(t *testing.T, projectRoot, sessionID string, fiveHour, sevenDay *QuotaWindowRecord, captured time.Time) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "state", contextUsageDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create context-usage dir: %v", err)
	}
	rec := SessionTelemetryRecord{
		SchemaVersion: 3,
		SessionID:     sessionID,
		WriterPID:     os.Getpid(),
		CapturedAt:    captured.Format(time.RFC3339Nano),
		FiveHour:      fiveHour,
		SevenDay:      sevenDay,
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal telemetry record: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, sessionID+".json"), data, 0o644); err != nil {
		t.Fatalf("write telemetry record: %v", err)
	}
}

// seedWorktreeAnchor turns worktreeRoot into a linked-worktree-shaped project
// root: .git is a regular file whose gitdir: line names the primary checkout's
// metadata for this worktree, and the primary's metadata carries the matching
// gitdir entry back (the real git shape, so QuotaStateDirs's enumeration is
// exercised, not bypassed). primaryGitdirLine is the raw first line written
// into worktreeRoot/.git — pass an absolute or a relative pointer per the case
// under test. Set primaryMetaGitdir to false to omit the primary-side entry
// (pure degradation fixtures).
func seedWorktreeAnchor(t *testing.T, primaryRoot, worktreeRoot, primaryGitdirLine string, primaryMetaGitdir bool) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(primaryRoot, ".git", "worktrees", "x"), 0o755); err != nil {
		t.Fatalf("create primary worktree metadata dir: %v", err)
	}
	if err := os.MkdirAll(worktreeRoot, 0o755); err != nil {
		t.Fatalf("create worktree root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreeRoot, ".git"), []byte(primaryGitdirLine+"\n"), 0o644); err != nil {
		t.Fatalf("write worktree .git file: %v", err)
	}
	if primaryMetaGitdir {
		gitdir := filepath.Join(worktreeRoot, ".git") + "\n"
		if err := os.WriteFile(filepath.Join(primaryRoot, ".git", "worktrees", "x", "gitdir"), []byte(gitdir), 0o644); err != nil {
			t.Fatalf("write primary gitdir entry: %v", err)
		}
	}
}

// quotaTestBuilder builds a Builder isolated from the host: git and version
// collection are mocked, so the only state surface a Build touches is the
// fixture anchor.
func quotaTestBuilder() Builder {
	return New(Options{
		GitProvider:    &mockGitProvider{},
		UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0"}},
		Mode:           ModeDefault,
		NoColor:        true,
	})
}

// glmFormStdin returns a GLM-form stdin payload (no rate_limits) anchored at
// projectDir — the input shape that today reaches the collector gate and after
// this SPEC reaches the snapshot fallback.
func glmFormStdin(projectDir string) string {
	return fmt.Sprintf(`{"workspace":{"project_dir":%q},"model":{"id":"glm-5.3","display_name":"GLM"}}`, projectDir)
}

// usageSegment extracts the pipe-separated bars segment that carries marker
// (e.g. "5H:"), so a 7D percentage on the same line cannot satisfy a 5H
// assertion.
func usageSegment(t *testing.T, out, marker string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, marker) {
			continue
		}
		for _, seg := range strings.Split(line, "|") {
			if strings.Contains(seg, marker) {
				return strings.TrimSpace(seg)
			}
		}
	}
	t.Fatalf("no segment containing %q in output:\n%s", marker, out)
	return ""
}

// futureResets returns a reset time comfortably ahead of the render, in the
// Unix-seconds unit the stdin schema and the records share.
func futureResets(d time.Duration) int64 {
	return time.Now().Add(d).Unix()
}

// TestBuild_QuotaFallbackNAWithoutRecord (AC-SOU-003): GLM-form stdin with no
// usable context-usage record renders n/a — never the old 0% bar.
func TestBuild_QuotaFallbackNAWithoutRecord(t *testing.T) {
	anchor := t.TempDir() // exists, but holds no records

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(glmFormStdin(anchor)))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "n/a") {
		t.Errorf("5H segment should render n/a with no snapshot record, got %q\noutput:\n%s", seg, out)
	}
	if strings.Contains(seg, "0%") {
		t.Errorf("5H segment must not render a 0%% bar, got %q", seg)
	}
	seg7 := usageSegment(t, out, "7D:")
	if !strings.Contains(seg7, "n/a") {
		t.Errorf("7D segment should render n/a with no snapshot record, got %q", seg7)
	}
	if strings.Contains(seg7, "0%") {
		t.Errorf("7D segment must not render a 0%% bar, got %q", seg7)
	}
}

// TestBuild_QuotaFallbackFreshRecord (AC-SOU-004): a fresh schema-v3 record's
// windows render their aggregated percentages plus the source record's
// captured-at age marker.
func TestBuild_QuotaFallbackFreshRecord(t *testing.T) {
	anchor := t.TempDir()
	captured := time.Now().Add(-2 * time.Minute)
	seedQuotaRecord(t, anchor, quotaTestSessionID,
		&QuotaWindowRecord{UsedPercentage: 42.5, ResetsAt: futureResets(2 * time.Hour)},
		&QuotaWindowRecord{UsedPercentage: 77.5, ResetsAt: futureResets(5 * 24 * time.Hour)},
		captured)

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(glmFormStdin(anchor)))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "42%") {
		t.Errorf("5H segment should render the fresh record's 42%%, got %q\noutput:\n%s", seg, out)
	}
	if !strings.Contains(seg, "captured ") {
		t.Errorf("5H segment should carry the captured-at age marker, got %q", seg)
	}
	seg7 := usageSegment(t, out, "7D:")
	if !strings.Contains(seg7, "77%") {
		t.Errorf("7D segment should render the fresh record's 77%%, got %q", seg7)
	}
	if !strings.Contains(seg7, "captured ") {
		t.Errorf("7D segment should carry the captured-at age marker, got %q", seg7)
	}
	if strings.Contains(usageSegment(t, out, "5H:"), "n/a") {
		t.Errorf("5H segment with a fresh record must not render n/a, got %q", seg)
	}
}

// TestBuild_QuotaFallbackResetRollover (AC-SOU-003, reset arm of REQ-SOU-004):
// a winning window whose reset time has passed renders n/a — never a stale
// percentage. The other window stays fresh and keeps rendering.
func TestBuild_QuotaFallbackResetRollover(t *testing.T) {
	anchor := t.TempDir()
	seedQuotaRecord(t, anchor, quotaTestSessionID,
		&QuotaWindowRecord{UsedPercentage: 88, ResetsAt: time.Now().Add(-1 * time.Hour).Unix()},
		&QuotaWindowRecord{UsedPercentage: 77.5, ResetsAt: futureResets(5 * 24 * time.Hour)},
		time.Now().Add(-2*time.Minute))

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(glmFormStdin(anchor)))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "n/a") {
		t.Errorf("rolled-over 5H window should render n/a, got %q\noutput:\n%s", seg, out)
	}
	if strings.Contains(seg, "%") {
		t.Errorf("rolled-over 5H window must not render any percentage, got %q", seg)
	}
	seg7 := usageSegment(t, out, "7D:")
	if !strings.Contains(seg7, "77%") {
		t.Errorf("still-fresh 7D window should keep rendering 77%%, got %q", seg7)
	}
}

// TestBuild_QuotaFallbackWorktreeAbsoluteGitdir (AC-SOU-009, absolute
// pointer): a render anchored at a linked-worktree root (.git a regular file
// with an absolute gitdir: line) reads the primary checkout's records — the
// only ones carrying a window — not just its own.
func TestBuild_QuotaFallbackWorktreeAbsoluteGitdir(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	wt := filepath.Join(base, "wt-card")
	seedQuotaRecord(t, primary, quotaTestSessionID,
		&QuotaWindowRecord{UsedPercentage: 42.5, ResetsAt: futureResets(2 * time.Hour)},
		&QuotaWindowRecord{UsedPercentage: 77.5, ResetsAt: futureResets(5 * 24 * time.Hour)},
		time.Now().Add(-2*time.Minute))
	seedWorktreeAnchor(t, primary, wt,
		"gitdir: "+filepath.Join(primary, ".git", "worktrees", "x"), true)

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(glmFormStdin(wt)))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "42%") {
		t.Errorf("worktree-rooted render should surface the primary record's 42%% (absolute gitdir), got %q\noutput:\n%s", seg, out)
	}
	seg7 := usageSegment(t, out, "7D:")
	if !strings.Contains(seg7, "77%") {
		t.Errorf("worktree-rooted render should surface the primary record's 77%% (absolute gitdir), got %q", seg7)
	}
}

// TestBuild_QuotaFallbackWorktreeRelativeGitdir (AC-SOU-009, relative pointer
// — absorbed sou-debt-2): git accepts relative worktree anchors; the fallback
// must resolve them via Clean+Join with the worktree root and surface the
// primary record's values exactly as the absolute case does.
func TestBuild_QuotaFallbackWorktreeRelativeGitdir(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	wt := filepath.Join(base, "wt-card")
	seedQuotaRecord(t, primary, quotaTestSessionID,
		&QuotaWindowRecord{UsedPercentage: 42.5, ResetsAt: futureResets(2 * time.Hour)},
		&QuotaWindowRecord{UsedPercentage: 77.5, ResetsAt: futureResets(5 * 24 * time.Hour)},
		time.Now().Add(-2*time.Minute))
	seedWorktreeAnchor(t, primary, wt,
		"gitdir: ../primary/.git/worktrees/x", true)

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(glmFormStdin(wt)))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "42%") {
		t.Errorf("relative gitdir pointer should resolve to the primary record's 42%%, got %q\noutput:\n%s", seg, out)
	}
	seg7 := usageSegment(t, out, "7D:")
	if !strings.Contains(seg7, "77%") {
		t.Errorf("relative gitdir pointer should resolve to the primary record's 77%%, got %q", seg7)
	}
}

// TestBuild_QuotaFallbackDegradedMalformedGitdir (AC-SOU-009 degradation arm,
// plan §F R5): a gitdir: pointer that names nothing degrades fail-open to the
// single-root scan — the worktree's own (window-less) records — so the
// segments render n/a and the primary's values never leak into the wrong
// project.
func TestBuild_QuotaFallbackDegradedMalformedGitdir(t *testing.T) {
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	wt := filepath.Join(base, "wt-lost")
	seedQuotaRecord(t, primary, quotaTestSessionID,
		&QuotaWindowRecord{UsedPercentage: 42.5, ResetsAt: futureResets(2 * time.Hour)},
		&QuotaWindowRecord{UsedPercentage: 77.5, ResetsAt: futureResets(5 * 24 * time.Hour)},
		time.Now().Add(-2*time.Minute))
	seedWorktreeAnchor(t, primary, wt,
		"gitdir: ../gone/.git/worktrees/x", false)

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(glmFormStdin(wt)))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "n/a") {
		t.Errorf("malformed gitdir should degrade to n/a (no local records), got %q\noutput:\n%s", seg, out)
	}
	if strings.Contains(seg, "42%") {
		t.Errorf("a dead pointer must never surface another tree's 42%%, got %q", seg)
	}
}

// TestBuild_QuotaFallbackNotReadWhenRateLimitsPresent (AC-SOU-005): with
// Claude-form stdin the bars render from rate_limits byte-for-byte as before
// and the snapshot fallback contributes nothing — the seeded record's values
// are absent. The exact 5H segment is pinned below (the AC-SOU-005
// byte-identical clause pins one full rendered segment, captured on the
// pre-change tree where the rate_limits path is untouched by this SPEC).
func TestBuild_QuotaFallbackNotReadWhenRateLimitsPresent(t *testing.T) {
	anchor := t.TempDir()
	seedQuotaRecord(t, anchor, quotaTestSessionID,
		&QuotaWindowRecord{UsedPercentage: 42.5, ResetsAt: futureResets(2 * time.Hour)},
		&QuotaWindowRecord{UsedPercentage: 77.5, ResetsAt: futureResets(5 * 24 * time.Hour)},
		time.Now().Add(-2*time.Minute))

	stdin := fmt.Sprintf(`{"workspace":{"project_dir":%q},"rate_limits":{"five_hour":{"used_percentage":23.5,"resets_at":1738425600},"seven_day":{"used_percentage":41.2,"resets_at":1738857600}}}`, anchor)

	out, err := quotaTestBuilder().Build(context.Background(), strings.NewReader(stdin))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	seg := usageSegment(t, out, "5H:")
	if !strings.Contains(seg, "23%") {
		t.Errorf("rate_limits 5H value must render, got %q\noutput:\n%s", seg, out)
	}
	if strings.Contains(seg, "42%") {
		t.Errorf("snapshot fallback must not be read (or rendered) when rate_limits are present, got %q", seg)
	}
	if strings.Contains(seg, "n/a") {
		t.Errorf("rate_limits 5H must not degrade to n/a, got %q", seg)
	}
	// AC-SOU-005 pin: the exact pre-change 5H segment for this stdin
	// (bar=10 blocks, NoColor). Captured verbatim on the pre-change tree.
	const want5H = "🔋 5H: ██░░░░░░░░ 23% (rolling)"
	if strings.TrimSpace(seg) != want5H {
		t.Errorf("AC-SOU-005 byte-identical pin:\n got 5H segment %q\nwant %q", seg, want5H)
	}
}

// TestRenderQuotaSegment pins the segment-level contract of
// renderQuotaSegment directly (the builder-level tests above exercise it
// end-to-end): fresh renders bar + reset + captured-at marker; unknown,
// reset, and absent aggregate all render n/a.
func TestRenderQuotaSegment(t *testing.T) {
	fresh := QuotaReading{
		State:          QuotaFresh,
		UsedPercentage: 42.5,
		ResetsAt:       futureResets(2 * time.Hour),
		CapturedAt:     time.Now().Add(-2 * time.Minute),
	}
	reset := QuotaReading{
		State:      QuotaReset,
		ResetsAt:   time.Now().Add(-1 * time.Hour).Unix(),
		CapturedAt: time.Now().Add(-2 * time.Minute),
	}

	tests := []struct {
		name        string
		reading     QuotaReading
		relReset    bool
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "fresh renders bar, reset, and captured marker",
			reading:     fresh,
			relReset:    true,
			wantContain: []string{"5H:", "42%", "captured "},
			wantAbsent:  []string{"n/a"},
		},
		{
			name:        "reset renders n/a with no percentage",
			reading:     reset,
			relReset:    true,
			wantContain: []string{"n/a"},
			wantAbsent:  []string{"88%", "%"},
		},
		{
			name:        "unknown renders n/a",
			reading:     QuotaReading{State: QuotaUnknown},
			relReset:    true,
			wantContain: []string{"n/a"},
			wantAbsent:  []string{"0%"},
		},
		{
			name:        "zero-value reading renders n/a",
			reading:     QuotaReading{},
			relReset:    false,
			wantContain: []string{"n/a"},
			wantAbsent:  []string{"0%"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderQuotaSegment("5H:", tt.reading, 10, nil, tt.relReset)
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("%q should contain %q, got %q", tt.name, want, got)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("%q must not contain %q, got %q", tt.name, absent, got)
				}
			}
		})
	}
}

// TestFormatCapturedAge pins the age-marker formatting (REQ-SOU-003).
func TestFormatCapturedAge(t *testing.T) {
	tests := []struct {
		name     string
		captured time.Time
		want     string
	}{
		{"zero reads unknown", time.Time{}, "unknown"},
		{"just now", time.Now().Add(-10 * time.Second), "<1m"},
		{"minutes", time.Now().Add(-5 * time.Minute), "5m"},
		{"hours", time.Now().Add(-2*time.Hour - 14*time.Minute), "2h 14m"},
		{"days", time.Now().Add(-50 * time.Hour), "2d 2h"},
	}
	for _, tt := range tests {
		if got := formatCapturedAge(tt.captured); got != tt.want {
			t.Errorf("%s: formatCapturedAge() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
