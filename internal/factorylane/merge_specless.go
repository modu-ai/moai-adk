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
// every other change after the audited commit makes the verdict stale, except
// a plain absorb of the integration branch that leaves the card's own diff
// byte-identical — card t2121, absorbOnlyChange). When verdict.md is absent the
// latest sync-audit*.md (highest -iter<N>) is the evidence file instead.
//
// The leader's default for a SPEC-less card, operator not yet confirmed; the
// SPEC path (checkSyncAudit) is untouched.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditverdict"
)

var (
	specLessCardIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	specLessSHAPattern    = regexp.MustCompile(`^audited_sha: ([0-9a-f]{12,40})$`)

	specLessSyncAuditPattern = regexp.MustCompile(`^sync-audit(?:-iter(\d+))?\.md$`)
)

// specLessMaxVerdictBytes bounds the verdict file the condition reads.
const specLessMaxVerdictBytes = int64(4 << 20)

// checkSyncAuditVerdictFile evaluates condition (a) for a SPEC-less card. repoDir
// is the card's tree (where the verdict file lives); git runs in that same tree.
func checkSyncAuditVerdictFile(git GitRunner, repoDir, card, develop string) (string, bool) {
	if !specLessCardIDPattern.MatchString(card) || strings.Contains(card, "..") {
		return fmt.Sprintf("card id %q is not a safe path segment, so its verdict file .moai/reports/<card-id>/verdict.md cannot be located", card), false
	}
	rel := specLessEvidenceRel(repoDir, card)
	path := filepath.Join(repoDir, filepath.FromSlash(rel))
	refuse := func(format string, args ...any) (string, bool) {
		return fmt.Sprintf("card %s has no SPEC, so its merge evidence is %s (or, when verdict.md is absent, the latest .moai/reports/%s/sync-audit*.md) — %s. "+
			"Produce that file on the card branch with the machine lines `verdict: PASS` (or `verdict: PASS-WITH-DEBT`) and `audited_sha: <full commit SHA the audit read>` "+
			"(the audited commit must be HEAD, differ from HEAD only under .moai/reports/%s/, or differ only by absorbing %s with the card's own diff unchanged)",
			card, rel, card, fmt.Sprintf(format, args...), card, develop), false
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

	shas := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if m := specLessSHAPattern.FindStringSubmatch(line); m != nil {
			shas[m[1]] = true
		}
	}

	// Card t1571: the verdict admission is the shared auditverdict predicate —
	// the same rule the contract, the kickoff evaluator, and the
	// card-transition guard use — so a verdict cannot pass this reader and
	// fail another site. PhaseSync keeps the label-only rule (PASS,
	// PASS-WITH-DEBT) plus the unconditional required_backend_fail refusal
	// (REQ-ACR-005) and the conflicting-decision-key refusal.
	fields := auditverdict.Parse(raw)
	// A spec-less card's merge evidence claims a CLOSED audit: a recorded
	// backend receipt that is not "pass" (fail or inconclusive) refuses the
	// verdict regardless of its own label — the PASS+INCONCLUSIVE mixed shape
	// from the t1557 gate repro. A malformed receipt line is refused
	// fail-closed; a verdict file with no receipt lines is the legacy shape
	// and admits on the label checks alone. (The convergence_overall: fail
	// refusal is owned by the shared auditverdict predicate itself — card
	// t1571 gate round 3 moved it there so every admission site is covered.)
	for _, v := range fields.Receipt.Backends {
		if v != "pass" {
			return refuse("%s is inadmissible: a recorded backend receipt reads %q — the audit record is not closed", path, v)
		}
	}
	if fields.MalformedReceipts > 0 {
		return refuse("%s is inadmissible: %d malformed receipt line(s)", path, fields.MalformedReceipts)
	}
	if ok, reason := auditverdict.Admit(fields, auditverdict.PhaseSync, 0, false); !ok {
		return refuse("%s is inadmissible: %s", path, reason)
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
				if absorbed, why := absorbOnlyChange(git, develop, auditedFull, head, card); !absorbed {
					return refuse("audited_sha %s is stale: HEAD %s differs from it outside %s (first changed path: %s) and the change is not a plain absorb of %s: %s", auditedFull, head, evidenceDir, changed, develop, why)
				}
				break
			}
		}
	}
	return fmt.Sprintf("%s: verdict: %s, audited_sha %s binds HEAD %s — the SPEC-less card's audit record reads closed", rel, fields.Label, auditedFull, head), true
}

// specLessEvidenceRel names the card's evidence file: verdict.md when it exists,
// else the latest sync-audit*.md (highest -iter<N>; a name without -iter counts
// as iteration 1), else verdict.md so the refusal names the primary file.
func specLessEvidenceRel(repoDir, card string) string {
	dir := ".moai/reports/" + card + "/"
	primary := dir + "verdict.md"
	if _, err := os.Lstat(filepath.Join(repoDir, filepath.FromSlash(primary))); err == nil {
		return primary
	}
	entries, err := os.ReadDir(filepath.Join(repoDir, filepath.FromSlash(dir)))
	if err != nil {
		return primary
	}
	best, bestN := "", -1
	for _, e := range entries {
		m := specLessSyncAuditPattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n := 1
		if m[1] != "" {
			n, _ = strconv.Atoi(m[1])
		}
		if n > bestN || (n == bestN && e.Name() > best) {
			best, bestN = e.Name(), n
		}
	}
	if best == "" {
		return primary
	}
	return dir + best
}

// absorbOnlyChange reports whether HEAD differs from the audited commit only by
// absorbing the integration branch, with the card's own diff untouched.
//
//	(a) every first-parent commit in audited..HEAD is either a merge whose extra
//	    parents all lie inside develop (an absorb of main, not of another
//	    branch) or touches nothing outside the card evidence directory;
//	(b) the card's diff against its merge-base with develop is the same at the
//	    audited commit and at HEAD, outside the evidence directory.
//
// (b) is the load-bearing proof: an evil merge — a conflict-resolution edit to a
// card-owned file, or a smuggled edit to any other path — changes the card diff
// (every path where HEAD differs from develop is in the card diff by
// construction), so it cannot pass. The comparison uses -U0 patches with the
// blob ids ("index" lines) dropped and the hunk positions blanked: blob ids of
// a file main also touched legitimately change on absorb, and hunk line numbers
// shift when main inserts lines earlier in the file, while the removed/added
// lines themselves are what the audit read. Anything unresolvable fails closed.
func absorbOnlyChange(git GitRunner, develop, audited, head, card string) (bool, string) {
	if strings.TrimSpace(develop) == "" {
		return false, "the integration branch is unknown"
	}
	dev, err := git.Git("rev-parse", "--verify", "--quiet", develop+"^{commit}")
	if err != nil || strings.TrimSpace(dev) == "" {
		return false, fmt.Sprintf("the integration ref %s cannot be resolved (fail-closed)", develop)
	}
	dev = strings.TrimSpace(dev)
	evidenceDir := ".moai/reports/" + card + "/"

	log, err := git.Git("rev-list", "--first-parent", "--parents", audited+".."+head)
	if err != nil {
		return false, fmt.Sprintf("the commits after the audited one could not be listed: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if len(f) > 2 { // a merge: every merged-in parent must already be in develop
			for _, p := range f[2:] {
				if _, err := git.Git("merge-base", "--is-ancestor", p, dev); err != nil {
					return false, fmt.Sprintf("merge %s brings in %s, which is not part of %s", f[0], p, develop)
				}
			}
			continue
		}
		names, err := git.Git("diff-tree", "--no-commit-id", "--name-only", "-r", "-z", "--root", f[0])
		if err != nil {
			return false, fmt.Sprintf("commit %s could not be inspected: %v", f[0], err)
		}
		for _, p := range strings.Split(names, "\x00") {
			if p != "" && !strings.HasPrefix(p, evidenceDir) {
				return false, fmt.Sprintf("commit %s changes %s after the audit", f[0], p)
			}
		}
	}

	before, err := cardPatch(git, dev, audited, evidenceDir)
	if err != nil {
		return false, err.Error()
	}
	after, err := cardPatch(git, dev, head, evidenceDir)
	if err != nil {
		return false, err.Error()
	}
	if before != after {
		return false, "the card's own diff against " + develop + " differs between the audited commit and HEAD"
	}
	return true, ""
}

// cardPatch returns the normalized card diff of tip against its merge-base with dev.
func cardPatch(git GitRunner, dev, tip, evidenceDir string) (string, error) {
	base, err := git.Git("merge-base", dev, tip)
	if err != nil || strings.TrimSpace(base) == "" {
		return "", fmt.Errorf("no merge-base between %s and %s: %v", dev, tip, err)
	}
	out, err := git.Git("diff", "--no-ext-diff", "--no-color", "--no-renames", "-U0", strings.TrimSpace(base), tip, "--", ".", ":(exclude)"+evidenceDir)
	if err != nil {
		return "", fmt.Errorf("the card diff at %s could not be computed: %v", tip, err)
	}
	var b strings.Builder
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(l, "index "):
			continue
		case strings.HasPrefix(l, "@@"):
			b.WriteString("@@\n")
		default:
			b.WriteString(l)
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
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
