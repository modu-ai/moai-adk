package homestate

// verdict_select.go — the one verdict-file selector. Every reader that needs
// "the card's sync-audit verdict" (the T13 guard, the rebind reader, the
// SPEC-less merge triple) goes through SelectCardVerdictFile, so a verdict
// cannot be found by one site and missed by another.

import (
	"os"
	"path/filepath"
)

// syncAuditPhase is the verdict phase whose file the doctrine names
// `verdict.md`.
const syncAuditPhase = "sync-audit"

// SelectCardVerdictFile returns the card's verdict file under
// <dir>/.moai/reports/<card>/: `verdict.md` when it is a regular file, else
// the highest-iteration `sync-audit*.md` (a name without `-iter<N>` is
// iteration 1; ties break to the latest modification time), else an error
// naming both accepted forms.
func SelectCardVerdictFile(dir, card string) (string, error) {
	if !ValidCardID(card) {
		return "", evidenceErr("card id %q is not a safe path segment", card)
	}
	if dir == "" {
		return "", evidenceErr("card has no worktree path")
	}
	reports := filepath.Join(dir, ".moai", "reports", card)
	canonical := filepath.Join(reports, "verdict.md")
	if info, err := os.Stat(canonical); err == nil && info.Mode().IsRegular() {
		return canonical, nil
	}
	matches, _ := filepath.Glob(filepath.Join(reports, syncAuditPhase+"*.md"))
	if len(matches) == 0 {
		return "", evidenceErr("no verdict file (verdict.md or sync-audit*.md) under %s", reports)
	}
	return newestVerdictFile(matches)
}

// locateVerdictFile finds the verdict file of one audit phase: the sync-audit
// phase uses the shared selector, any other phase keeps its `<phase>*.md`
// newest-iteration rule.
func locateVerdictFile(dir, cardID, phase string) (string, error) {
	if phase == syncAuditPhase {
		return SelectCardVerdictFile(dir, cardID)
	}
	pattern := filepath.Join(dir, ".moai", "reports", cardID, phase+"*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", evidenceErr("no %s verdict file under %s", phase, filepath.Dir(pattern))
	}
	return newestVerdictFile(matches)
}
