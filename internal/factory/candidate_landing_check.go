// candidate_landing_check.go — the SHARED landing check (card t1478,
// SPEC-CANDIDATE-CI-001 REQ-CCI-011, design.md D2/D10): the gate-5 seam
// both call sites wire. A merge may land only on a green candidate whose
// record is bound to THIS merge — same card and pinned SHA, same target
// branch (identity refuses outright), same tip (an advance voids the
// verification with re-candidate guidance), candidate still descending from
// the pinned SHA. The check reads the record it judges and the git state it
// needs; it never writes anything and never touches the window.
package factory

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// candidateCIRequired reads workflow.candidate_ci.enabled at the merge
// step's root — the fail-closed half of gate 5 (AC-CCI-011-1). An absent
// key and an unreadable config both read false: the step never invents a
// gate the project did not ask for, exactly like the cli side's
// candidateCIEnabled.
func candidateCIRequired(root string) bool {
	cfg, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil || cfg == nil {
		return false
	}
	return cfg.Workflow.CandidateCI.Enabled
}

// LandingCheckInput names one landing decision. Root is the PRIMARY
// checkout (the candidate record store); IntegrationWorktree is the tree
// the merge runs in (its branch tips are the merge's reality); Git
// overrides the git runner (nil = execGitIn over the integration worktree),
// which is how the step-level tests drive the check hermetically.
type LandingCheckInput struct {
	Root                string
	CardID              string
	PinnedSHA           string
	TargetBranch        string
	IntegrationWorktree string
	Git                 func(args ...string) (string, error)
}

func (in LandingCheckInput) git() (func(args ...string) (string, error), error) {
	if in.Git != nil {
		return in.Git, nil
	}
	if strings.TrimSpace(in.IntegrationWorktree) == "" {
		return nil, fmt.Errorf("no integration worktree to read %s's tips from", in.TargetBranch)
	}
	return func(args ...string) (string, error) {
		return execGitIn(in.IntegrationWorktree, args...)
	}, nil
}

// CandidateLandingCheck refuses anything that is not a green candidate
// record bound to the merge's actual target. Refusal order follows the
// record's own trust chain: identity first — the record read at this card's
// lookup path must name this card and this pinned SHA, and it must have been
// verified against this target (a record of another card, of another pinned
// SHA, or of another target is evidence about a different merge, and nothing
// about the verdict repairs that) — then tip equality (the verified tree must
// be the tree about to be merged), then the verdict (pending is not-green —
// fail-closed by default), then ancestry (an old green for an ancestor must
// not admit the new tip).
func CandidateLandingCheck(in LandingCheckInput) error {
	rec, err := ReadCandidateRecord(in.Root, in.CardID, in.PinnedSHA)
	if errors.Is(err, ErrCandidateRecordAbsent) {
		return fmt.Errorf("no candidate record for card %s at pinned %s — run moai integration candidate --card %s first (merging an unverified tree is what the gate exists to refuse)", in.CardID, shortSHAFull(in.PinnedSHA), in.CardID)
	}
	if err != nil {
		return fmt.Errorf("read the candidate record: %v", err)
	}

	// (0) RECORD IDENTITY (P2, card t1478 sync audit round 3): the record read
	// at this card's lookup path must name this card and this pinned SHA.
	// ReadCandidateRecord returns the key fields as written, so a valid record
	// of another card or of another pinned SHA stored at this path is evidence
	// about a different merge; it never admits, whatever its verdict.
	if rec.CardID != in.CardID || rec.PinnedSHA != in.PinnedSHA {
		return fmt.Errorf("identity mismatch: the candidate record stored for card %s at pinned %s names card %s at pinned %s — a record of another card or pinned SHA never admits; re-candidate with moai integration candidate --card %s", in.CardID, shortSHAFull(in.PinnedSHA), orUnset(rec.CardID), orUnset(shortSHAFull(rec.PinnedSHA)), in.CardID)
	}

	// (i) BRANCH IDENTITY (D10): a mismatch refuses OUTRIGHT — no
	// re-candidate guidance, because the record's branch itself must
	// change; the fix is a candidate built against the real target.
	if rec.IntegrationBranch != in.TargetBranch {
		return fmt.Errorf("target mismatch: candidate %s was verified against %s, this merge targets %s — a candidate verified against another target never admits; re-candidate against %s", shortSHAFull(rec.CandidateSHA), rec.IntegrationBranch, in.TargetBranch, in.TargetBranch)
	}

	git, err := in.git()
	if err != nil {
		return err
	}
	// (ii) TIP EQUALITY (D10): the target's current tip must still be the
	// candidate's first parent — the recorded integration tip. An advance
	// voids the verification and takes the same remedy the merge step's
	// cause 2 prescribes: re-measure, re-acquire, and re-candidate.
	tip, err := git("rev-parse", "refs/heads/"+in.TargetBranch)
	if err != nil {
		return fmt.Errorf("read the %s tip: %v", in.TargetBranch, err)
	}
	tip = strings.TrimSpace(tip)
	if rec.IntegrationTip != tip {
		return fmt.Errorf("stale candidate: the %s tip advanced past the candidate's base (recorded %s, now %s) — the verified tree is no longer the tree about to be merged; re-candidate with moai integration candidate --card %s", in.TargetBranch, shortSHAFull(rec.IntegrationTip), shortSHAFull(tip), in.CardID)
	}

	// (iii) VERDICT: only green admits. Pending is the push-time state and
	// red is a failed run — both refuse, naming what the record holds.
	if rec.Verdict != CandidateVerdictGreen {
		return fmt.Errorf("candidate %s verdict is %q (run %s, observed %s) — only a green candidate admits; observe the CI run with moai integration candidate --card %s --observe", shortSHAFull(rec.CandidateSHA), rec.Verdict, orUnset(rec.RunID), orUnset(rec.ObservedAt), in.CardID)
	}

	// (iv) ANCESTRY (D2): the candidate must still descend from the pinned
	// SHA — a stale green for an ancestor admits nothing.
	if _, err := git("merge-base", "--is-ancestor", in.PinnedSHA, rec.CandidateSHA); err != nil {
		return fmt.Errorf("candidate %s does not descend from the pinned %s — re-candidate", shortSHAFull(rec.CandidateSHA), shortSHAFull(in.PinnedSHA))
	}
	return nil
}

// @MX:ANCHOR: [AUTO] 12-character SHA shortener for candidate refusal messages.
// @MX:REASON: a refusal that names a candidate and a pinned SHA must truncate both the same way; a second length would make one message cite two spellings of one SHA.
func shortSHAFull(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// @MX:ANCHOR: [AUTO] empty-field placeholder for candidate refusal messages.
// @MX:REASON: the "(unset)" token is what every refusal shows for an empty field; changing it changes all of them at once.
func orUnset(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(unset)"
	}
	return s
}
