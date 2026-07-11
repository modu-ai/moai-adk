package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// spec_archive_test.go — CLI surface for the SPEC auto-archive capability
// (SPEC-SESSIONSTART-PERF-001 M2, REQ-SSP-008 / REQ-SSP-009).

// withProjectRoot points findProjectRootFn at dir for the duration of the test.
func withProjectRoot(t *testing.T, dir string) {
	t.Helper()

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })
}

// writeArchiveSPEC creates a SPEC dir whose frontmatter `updated:` date drives the
// grace-window decision (the archive command falls back to it when git has no
// history for the SPEC — which is the case in a bare t.TempDir()).
func writeArchiveSPEC(t *testing.T, baseDir, specID, status, updated string) {
	t.Helper()

	dir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}

	body := fmt.Sprintf(`---
id: %s
title: "Archive fixture"
version: "0.1.0"
status: %s
created: 2020-01-01
updated: %s
author: test
priority: P1
phase: "v3.0.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "fixture"
---

# %s
`, specID, status, updated, specID)

	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
}

func runArchiveCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := newSpecArchiveCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return out.String(), err
}

func specExists(t *testing.T, baseDir, specID string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(baseDir, ".moai", "specs", specID, "spec.md"))
	return err == nil
}

// long ago, comfortably outside any sane grace window
const staleDate = "2021-03-04"

func recentDate() string {
	return time.Now().AddDate(0, 0, -3).Format("2006-01-02")
}

// TestSpecArchiveCmd_DryRunMovesNothing is AC-SSP-008a: --dry-run reports the
// eligible set and touches no file.
func TestSpecArchiveCmd_DryRunMovesNothing(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	writeArchiveSPEC(t, base, "SPEC-DRY-001", "completed", staleDate)

	out, err := runArchiveCmd(t, "--dry-run")
	if err != nil {
		t.Fatalf("archive --dry-run: %v\n%s", err, out)
	}

	if !strings.Contains(out, "SPEC-DRY-001") {
		t.Errorf("dry-run must report the eligible SPEC; got:\n%s", out)
	}
	if !specExists(t, base, "SPEC-DRY-001") {
		t.Error("dry-run must not move anything")
	}
	if _, err := os.Stat(filepath.Join(base, ".moai", "archive")); !os.IsNotExist(err) {
		t.Error("dry-run must not create the archive tree")
	}
}

// TestSpecArchiveCmd_RequiresConfirmation is the destructive-path safety guard:
// a bare `moai spec archive` must not silently relocate directories. The operator
// confirms with --yes (or previews with --dry-run) first.
func TestSpecArchiveCmd_RequiresConfirmation(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	writeArchiveSPEC(t, base, "SPEC-NOCONFIRM-001", "completed", staleDate)

	out, err := runArchiveCmd(t)
	if err == nil {
		t.Fatalf("bare `spec archive` with eligible SPECs must refuse without --yes; got:\n%s", out)
	}
	if !specExists(t, base, "SPEC-NOCONFIRM-001") {
		t.Fatal("the unconfirmed path must not move anything")
	}
	if !strings.Contains(out, "--yes") {
		t.Errorf("the refusal must name the confirmation flag; got:\n%s", out)
	}
}

// TestSpecArchiveCmd_ApplyMovesEligible is AC-SSP-008b (with the confirmation flag).
func TestSpecArchiveCmd_ApplyMovesEligible(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	writeArchiveSPEC(t, base, "SPEC-MOVE-001", "completed", staleDate)
	writeArchiveSPEC(t, base, "SPEC-ACTIVE-001", "in-progress", staleDate)

	out, err := runArchiveCmd(t, "--yes")
	if err != nil {
		t.Fatalf("archive --yes: %v\n%s", err, out)
	}

	if specExists(t, base, "SPEC-MOVE-001") {
		t.Error("eligible SPEC should have been moved out of .moai/specs")
	}
	if !specExists(t, base, "SPEC-ACTIVE-001") {
		t.Error("non-terminal SPEC must stay put")
	}

	moved := filepath.Join(base, ".moai", "archive", "specs", "2021", "SPEC-MOVE-001", "spec.md")
	if _, statErr := os.Stat(moved); statErr != nil {
		t.Errorf("expected archived copy at %s: %v", moved, statErr)
	}
	if !strings.Contains(out, "SPEC-MOVE-001") {
		t.Errorf("the moved set must be reported; got:\n%s", out)
	}
}

// TestSpecArchiveCmd_GraceDaysFlag is AC-SSP-009a: --grace-days N gates eligibility.
func TestSpecArchiveCmd_GraceDaysFlag(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	old := time.Now().AddDate(0, 0, -60).Format("2006-01-02")
	writeArchiveSPEC(t, base, "SPEC-SIXTY-001", "completed", old)

	// Grace 90d → 60d-old SPEC is INSIDE the window → not eligible.
	out, err := runArchiveCmd(t, "--dry-run", "--grace-days", "90")
	if err != nil {
		t.Fatalf("archive --grace-days 90: %v\n%s", err, out)
	}
	if strings.Contains(out, "SPEC-SIXTY-001") {
		t.Errorf("60d-old SPEC must not be eligible at 90d grace; got:\n%s", out)
	}

	// Grace 30d → the same SPEC is now past the window → eligible.
	out, err = runArchiveCmd(t, "--dry-run", "--grace-days", "30")
	if err != nil {
		t.Fatalf("archive --grace-days 30: %v\n%s", err, out)
	}
	if !strings.Contains(out, "SPEC-SIXTY-001") {
		t.Errorf("60d-old SPEC must be eligible at 30d grace; got:\n%s", out)
	}
}

// TestSpecArchiveCmd_DefaultGraceDaysIs90 is AC-SSP-009b: flag absent → configured
// default (90).
func TestSpecArchiveCmd_DefaultGraceDaysIs90(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	writeArchiveSPEC(t, base, "SPEC-DEFAULTS-001", "completed", staleDate)

	out, err := runArchiveCmd(t, "--dry-run", "--json")
	if err != nil {
		t.Fatalf("archive --dry-run --json: %v\n%s", err, out)
	}

	var payload struct {
		GraceDays  int `json:"grace_days"`
		Candidates []struct {
			SPECID string `json:"spec_id"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, out)
	}

	if payload.GraceDays != 90 {
		t.Errorf("default grace_days = %d, want 90", payload.GraceDays)
	}
	if len(payload.Candidates) != 1 || payload.Candidates[0].SPECID != "SPEC-DEFAULTS-001" {
		t.Errorf("unexpected candidates: %+v", payload.Candidates)
	}
}

// TestSpecArchiveCmd_JSONReportsEraFinal: the JSON surface exposes the grandfather
// flag so an operator (or a reviewing agent) can audit the plan before applying it —
// the direct mitigation for the verification-claim-integrity.md §5 incident.
func TestSpecArchiveCmd_JSONReportsEraFinal(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	// No progress.md → ClassifyEra H-1 → V2.x → era-final (grandfather-protected).
	writeArchiveSPEC(t, base, "SPEC-GRANDFATHER-001", "completed", staleDate)

	out, err := runArchiveCmd(t, "--dry-run", "--json")
	if err != nil {
		t.Fatalf("archive --dry-run --json: %v\n%s", err, out)
	}

	var payload struct {
		Candidates []struct {
			SPECID   string `json:"spec_id"`
			EraFinal bool   `json:"era_final"`
			Status   string `json:"status"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if len(payload.Candidates) != 1 {
		t.Fatalf("want 1 candidate, got %+v", payload.Candidates)
	}
	if !payload.Candidates[0].EraFinal {
		t.Error("era_final must be surfaced as true for a grandfather-era SPEC")
	}
	if payload.Candidates[0].Status != "completed" {
		t.Errorf("status = %q, want completed", payload.Candidates[0].Status)
	}
}

// TestSpecArchiveCmd_NothingEligible: an empty plan is a clean exit, not an error,
// and it must not demand confirmation.
func TestSpecArchiveCmd_NothingEligible(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	writeArchiveSPEC(t, base, "SPEC-YOUNG-001", "completed", recentDate())
	writeArchiveSPEC(t, base, "SPEC-BUSY-001", "in-progress", staleDate)

	out, err := runArchiveCmd(t)
	if err != nil {
		t.Fatalf("an empty plan must exit cleanly without --yes: %v\n%s", err, out)
	}
	if !specExists(t, base, "SPEC-YOUNG-001") || !specExists(t, base, "SPEC-BUSY-001") {
		t.Fatal("nothing was eligible; nothing may move")
	}
}

func TestSpecArchiveCmd_InvalidGraceDays(t *testing.T) {
	base := t.TempDir()
	withProjectRoot(t, base)

	if _, err := runArchiveCmd(t, "--dry-run", "--grace-days", "-1"); err == nil {
		t.Fatal("negative --grace-days must be rejected")
	}
}

// TestSpecArchiveCmd_NoAskUserQuestion is the C-HRA-008 subagent-boundary guard:
// CLI code never prompts the user. Confirmation is a flag, not a prompt.
func TestSpecArchiveCmd_NoAskUserQuestion(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("spec_archive.go")
	if err != nil {
		t.Fatalf("read spec_archive.go: %v", err)
	}

	for _, banned := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(string(src), banned) {
			t.Errorf("spec_archive.go references %q — CLI code must never prompt the user", banned)
		}
	}
}
