package factorylane

// merge_specless.go — the sync-audit condition of the merge-readiness triple
// for a card that has no SPEC (SPEC-GITHUB-FLOW-DEFAULT-001 M2-B follow-up,
// card t1453, C5). A SPEC-less card (Class A/B) has no progress.md §E.4 sync
// record, so it closes on its audit verdict file instead:
//
//	.moai/reports/<card-id>/verdict.md
//
// The condition passes only when that file exists in the card's own tree, is
// readable, carries the machine line `verdict: PASS` (or PASS-WITH-DEBT — the
// verdicts the F1 audit gate counts as a pass, audit-artifact-convention.md
// § What), and carries an `audited_sha:` line that still describes the card:
// the audited commit is HEAD, or differs from HEAD only under the card's own
// evidence directory (a verdict file cannot name the commit that adds it, so
// the commit that lands the verdict is allowed to sit after the audited one —
// every other change after the audited commit makes the verdict stale).
//
// The leader's default for a SPEC-less card, operator not yet confirmed; the
// SPEC path (checkSyncAudit) is untouched.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	specLessCardIDPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	specLessVerdictPattern = regexp.MustCompile(`^verdict: (PASS|PASS-WITH-DEBT|FAIL)$`)
	specLessSHAPattern     = regexp.MustCompile(`^audited_sha: ([0-9a-f]{12,40})$`)
)

// specLessMaxVerdictBytes bounds the verdict file the condition reads.
const specLessMaxVerdictBytes = int64(4 << 20)

// checkSyncAuditVerdictFile evaluates condition (a) for a SPEC-less card. repoDir
// is the card's tree (where the verdict file lives); git runs in that same tree.
func checkSyncAuditVerdictFile(git GitRunner, repoDir, card string) (string, bool) {
	if !specLessCardIDPattern.MatchString(card) || strings.Contains(card, "..") {
		return fmt.Sprintf("card id %q is not a safe path segment, so its verdict file .moai/reports/<card-id>/verdict.md cannot be located", card), false
	}
	rel := ".moai/reports/" + card + "/verdict.md"
	path := filepath.Join(repoDir, filepath.FromSlash(rel))
	refuse := func(format string, args ...any) (string, bool) {
		return fmt.Sprintf("card %s has no SPEC, so its merge evidence is %s — %s. "+
			"Produce that file on the card branch with the machine lines `verdict: PASS` (or `verdict: PASS-WITH-DEBT`) and `audited_sha: <full commit SHA the audit read>` "+
			"(the audited commit must be HEAD, or differ from HEAD only under .moai/reports/%s/)",
			card, rel, fmt.Sprintf(format, args...), card), false
	}

	info, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return refuse("%s does not exist", path)
	case err != nil:
		return refuse("%s is unreadable: %v", path, err)
	case !info.Mode().IsRegular() || info.Size() > specLessMaxVerdictBytes:
		return refuse("%s is not a regular file of at most %d bytes", path, specLessMaxVerdictBytes)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return refuse("%s is unreadable: %v", path, err)
	}

	verdicts, shas := map[string]bool{}, map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if m := specLessVerdictPattern.FindStringSubmatch(line); m != nil {
			verdicts[m[1]] = true
		}
		if m := specLessSHAPattern.FindStringSubmatch(line); m != nil {
			shas[m[1]] = true
		}
	}
	verdict, ok := soleKey(verdicts)
	switch {
	case len(verdicts) == 0:
		return refuse("%s carries no `verdict:` line (verdict: PASS, PASS-WITH-DEBT or FAIL)", path)
	case !ok:
		return refuse("%s carries conflicting `verdict:` lines", path)
	case verdict != "PASS" && verdict != "PASS-WITH-DEBT":
		return refuse("%s reads verdict: %s, want verdict: PASS or verdict: PASS-WITH-DEBT", path, verdict)
	}
	audited, ok := soleKey(shas)
	switch {
	case len(shas) == 0:
		return refuse("%s carries no `audited_sha:` line", path)
	case !ok:
		return refuse("%s carries conflicting `audited_sha:` lines", path)
	}

	auditedFull, err := git.Git("rev-parse", "--verify", "--quiet", audited+"^{commit}")
	if err != nil || strings.TrimSpace(auditedFull) == "" {
		return refuse("audited_sha %s does not name a commit in this repository", audited)
	}
	auditedFull = strings.TrimSpace(auditedFull)
	headOut, err := git.Git("rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return refuse("HEAD could not be read to bind audited_sha %s: %v", audited, err)
	}
	head := strings.TrimSpace(headOut)
	if auditedFull != head {
		if _, err := git.Git("merge-base", "--is-ancestor", auditedFull, head); err != nil {
			var exitErr *GitExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode == 1 {
				return refuse("audited_sha %s is not an ancestor of HEAD %s — the verdict audited a different line of history", auditedFull, head)
			}
			return refuse("audited_sha %s could not be compared with HEAD %s: %v", auditedFull, head, err)
		}
		diffOut, err := git.Git("diff", "--name-only", "-z", auditedFull, head)
		if err != nil {
			return refuse("the change from audited_sha %s to HEAD %s could not be listed: %v", auditedFull, head, err)
		}
		evidenceDir := ".moai/reports/" + card + "/"
		for _, changed := range strings.Split(diffOut, "\x00") {
			if changed != "" && !strings.HasPrefix(changed, evidenceDir) {
				return refuse("audited_sha %s is stale: HEAD %s differs from it outside %s (first changed path: %s)", auditedFull, head, evidenceDir, changed)
			}
		}
	}
	return fmt.Sprintf("%s: verdict: %s, audited_sha %s binds HEAD %s — the SPEC-less card's audit record reads closed", rel, verdict, auditedFull, head), true
}

// soleKey returns the only key of m, and false when m does not hold exactly one.
func soleKey(m map[string]bool) (string, bool) {
	if len(m) != 1 {
		return "", false
	}
	for k := range m {
		return k, true
	}
	return "", false
}
