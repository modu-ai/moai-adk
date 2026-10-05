// Resolution: UPGRADE — CLAUDE.md 40,000-char budget check per coding-standards.md.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// instructionsLoadedHandler processes InstructionsLoaded events.
// It validates character budget compliance for CLAUDE.md and rules files.
type instructionsLoadedHandler struct{}

// NewInstructionsLoadedHandler creates a new InstructionsLoaded event handler.
func NewInstructionsLoadedHandler() Handler {
	return &instructionsLoadedHandler{}
}

// EventType returns EventInstructionsLoaded.
func (h *instructionsLoadedHandler) EventType() EventType {
	return EventInstructionsLoaded
}

// Handle processes an InstructionsLoaded event. It checks that loaded files
// comply with the 40,000 character budget per coding-standards.md.
func (h *instructionsLoadedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Official stdin field for InstructionsLoaded is file_path;
	// instruction_file_path is the legacy MoAI field name (kept as fallback).
	instructionPath := input.FilePath
	if instructionPath == "" {
		instructionPath = input.InstructionFilePath
	}

	slog.Info("instructions loaded",
		"session_id", input.SessionID,
		"file_path", instructionPath,
		"load_reason", input.LoadReason,
		"globs", input.Globs,
		"trigger_file_path", input.TriggerFilePath,
	)

	// The slog record above does not reach a reader HERE: resolveLoggingDecision
	// in internal/cli/logging.go keeps every `moai hook` invocation off
	// stdout/stderr, unconditionally, so both stay clean for the hook JSON
	// contract. It is an Info record, so it also sits below the sink's level
	// gate (.moai/logs/hook-runtime.log admits warn and above) — it is written
	// nowhere at all.
	//
	// The persistent record is therefore an audit row, written here — without it
	// the three host-supplied fields are observable nowhere.
	// The destination is the write-side project root, never input.CWD: a
	// session whose cwd is a subdirectory would otherwise grow a stray
	// <subdir>/.moai/logs/ tree (card t1160).
	appendRuleLoadAudit(resolveProjectRoot(input), RuleLoadAuditRecord{
		SessionID:       input.SessionID,
		FilePath:        instructionPath,
		LoadReason:      input.LoadReason,
		Globs:           input.Globs,
		TriggerFilePath: input.TriggerFilePath,
	})

	// Check character budget for the loaded instruction file
	if instructionPath != "" {
		if err := h.checkCharacterBudget(instructionPath); err != nil {
			return &HookOutput{
				SystemMessage: err.Error(),
			}, nil
		}
	}

	// Also check CLAUDE.md in CWD as a fallback
	if input.CWD != "" {
		claudeMDPath := filepath.Join(input.CWD, "CLAUDE.md")
		if _, err := os.Stat(claudeMDPath); err == nil {
			if budgetErr := h.checkCharacterBudget(claudeMDPath); budgetErr != nil {
				slog.Warn("CLAUDE.md exceeds budget", "path", claudeMDPath, "error", budgetErr)
				// Don't block on CLAUDE.md budget violations, just log
			}
		}
	}

	// Aggregate session-start budget (SPEC-INSTRUCTIONS-BUDGET-001,
	// REQ-INSTRBUDGET-002). Advisory only: the message names the overshoot;
	// the event is never blocked, the exit path stays zero, and the per-file
	// checks above are unchanged. The set roots at the write-side project
	// root (resolveProjectRoot), never input.CWD (t1160 precedent). An empty
	// derivation skips the comparison entirely.
	if root := resolveProjectRoot(input); root != "" {
		if set := instructionFileSet(root); len(set) > 0 {
			if agg := aggregateInstructionChars(root); agg > sessionCharBudget {
				return &HookOutput{
					SystemMessage: fmt.Sprintf(
						"instruction files aggregate %d chars, over the %d-char session budget across %d files; diet the always-loaded surface (advisory, does not block)",
						agg, sessionCharBudget, len(set)),
				}, nil
			}
		}
	}

	return &HookOutput{}, nil
}

// checkCharacterBudget verifies that a file does not exceed the 40,000 character limit.
func (h *instructionsLoadedHandler) checkCharacterBudget(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// File may have been deleted or is inaccessible - log and continue
		slog.Debug("failed to read instruction file", "path", filePath, "error", err)
		return nil
	}

	// Count UTF-8 characters
	charCount := utf8.RuneCount(data)

	// Check budget limit (40,000 characters per coding-standards.md)
	const charBudget = 40000
	if charCount > charBudget {
		return fmt.Errorf("%s exceeds 40,000 char budget at %d; split content per coding-standards.md",
			filePath, charCount)
	}

	return nil
}

// sessionCharBudget is the aggregate character budget for the session-start
// instruction-file set (the CLAUDE.md @-import closure plus the always-loaded
// rules under .claude/rules/moai/). It sits file-locally next to charBudget
// per plan decision D2: this package treats its budgets as handler-local, and
// the in-file precedent is the per-file constant itself. Changing this value
// is an operator act, not an implementation one.
//
// @MX:NOTE: 210000 is the operator ruling, not a derived limit — do not tune
// it to quiet the arrival advisory; the diet it calls for is a separate card.
// @MX:SPEC:SPEC-INSTRUCTIONS-BUDGET-001
const sessionCharBudget = 210000

// instructionFileSet derives the instruction-file set for projectRoot
// mechanically at metric time (SPEC-INSTRUCTIONS-BUDGET-001 §D.1 — never a
// hand-written path list):
//
//  1. the CLAUDE.md anchor;
//  2. the transitive closure of `^@` import lines in the anchor — repo-
//     relative paths resolved under projectRoot; missing members are skipped,
//     imports escaping projectRoot are not followed, and the traversal is
//     cycle-safe via a visited set;
//  3. every *.md under .claude/rules/moai/ whose frontmatter — when the file
//     starts with a --- block — lacks a top-level `paths:` key (the
//     always-loaded rule set).
//
// Returns absolute cleaned paths in sorted order, duplicates removed. A
// hardcoded enumeration of the paths would fail the fixture-growth tests.
func instructionFileSet(projectRoot string) []string {
	seen := make(map[string]bool)
	var set []string

	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		set = append(set, p)
	}

	// Anchor + transitive @-import closure (breadth-first, cycle-safe).
	queue := []string{filepath.Join(projectRoot, "CLAUDE.md")}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if seen[cur] {
			continue
		}
		data, err := os.ReadFile(cur)
		if err != nil {
			continue // missing/unreadable member: skip silently-and-continue
		}
		add(cur)
		for _, imp := range parseImports(string(data)) {
			if resolved := resolveUnder(projectRoot, imp); resolved != "" && !seen[resolved] {
				queue = append(queue, resolved)
			}
		}
	}

	// Always-loaded rules under .claude/rules/moai/.
	rulesDir := filepath.Join(projectRoot, ".claude", "rules", "moai")
	_ = filepath.WalkDir(rulesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable subtree member: skip
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if ruleFileAlwaysLoaded(path) {
			add(path)
		}
		return nil
	})

	sort.Strings(set)
	return set
}

// parseImports extracts the repo-relative targets of `^@` import lines.
// Everything from the first whitespace after the path token onward is
// ignored; a line with nothing usable after the @ is skipped.
func parseImports(content string) []string {
	var imports []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "@") {
			continue
		}
		fields := strings.Fields(trimmed[len("@"):])
		if len(fields) == 0 {
			continue
		}
		imports = append(imports, fields[0])
	}
	return imports
}

// resolveUnder resolves a repo-relative import path under projectRoot,
// refusing absolute paths and any target that escapes projectRoot.
func resolveUnder(projectRoot, rel string) string {
	if rel == "" || filepath.IsAbs(rel) {
		return ""
	}
	cleaned := filepath.Clean(filepath.Join(projectRoot, rel))
	escaped, err := filepath.Rel(projectRoot, cleaned)
	if err != nil || escaped == ".." || strings.HasPrefix(escaped, ".."+string(filepath.Separator)) {
		return ""
	}
	return cleaned
}

// ruleFileAlwaysLoaded reports whether a rules-tree markdown file belongs to
// the always-loaded surface: it has no frontmatter block, or its frontmatter
// lacks a top-level `paths:` key. A file without frontmatter is
// always-loaded; malformed frontmatter (opening --- with no closer) is also
// always-loaded — no top-level `paths:` is observable. An unreadable file is
// never always-loaded (skipped silently per REQ-INSTRBUDGET-003).
func ruleFileAlwaysLoaded(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return true // no frontmatter block
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			break // closing delimiter: frontmatter ended without paths:
		}
		if strings.HasPrefix(lines[i], "paths:") {
			return false // paths:-scoped: loads on demand, not session-start
		}
	}
	return true
}

// aggregateInstructionChars sums utf8.RuneCount over the derived
// instruction-file set, skipping unreadable members (REQ-INSTRBUDGET-003).
// Rune count matches the per-file checkCharacterBudget unit (plan decision
// D4) so the two budgets measure the same unit.
func aggregateInstructionChars(projectRoot string) int {
	total := 0
	for _, path := range instructionFileSet(projectRoot) {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		total += utf8.RuneCount(data)
	}
	return total
}

// ruleLoadAuditFileName is the InstructionsLoaded observation log under
// <projectRoot>/.moai/logs/ — one JSONL row per rule/instruction load,
// shaped like agent-stop-audit.jsonl.
const ruleLoadAuditFileName = "rule-load-audit.jsonl"

// RuleLoadAuditRecord is one observed instruction-load event. LoadReason,
// Globs, and TriggerFilePath are the host-supplied fields (HookInput, v2.1.69+)
// that the handler previously dropped, which left `paths:` glob matching
// unobservable at runtime.
type RuleLoadAuditRecord struct {
	Timestamp       string   `json:"timestamp"`
	SessionID       string   `json:"session_id"`
	FilePath        string   `json:"file_path"`
	LoadReason      string   `json:"load_reason,omitempty"`
	Globs           []string `json:"globs,omitempty"`
	TriggerFilePath string   `json:"trigger_file_path,omitempty"`
}

// appendRuleLoadAudit appends one record to <projectRoot>/.moai/logs/.
// Every failure path is silent-and-continue, matching the agent-stop-audit
// precedent: an audit failure may never fail the observed hook event.
func appendRuleLoadAudit(projectRoot string, rec RuleLoadAuditRecord) {
	if projectRoot == "" {
		return
	}
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(rec)
	if err != nil {
		return
	}
	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(logsDir, ruleLoadAuditFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.Write(append(line, '\n'))
}
