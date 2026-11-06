package cli

// audit_delta_default.go — the default delta proposal of SPEC-AUDIT-ENGINE-
// CONVERGENCE-001 (REQ-AEC-001, REQ-AEC-002). An audit_multi call that names a
// card_id, carries no delta_base and targets baseBranch proposes the audited_sha
// of the card's newest exported FAIL verdict file; that proposal then goes through
// the chain gate (audit_delta_chain.go) exactly as an explicit SHA does. Any other
// verdict state proposes nothing, and the result says why. delta_base "full"
// opts out.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditverdict"
	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	// deltaBaseFull is the delta_base value that requests the full resolved base.
	deltaBaseFull = "full"

	// deltaDefaultPrefix leads the result's note on every card call that carried no
	// explicit delta_base.
	deltaDefaultPrefix = "default delta_base: "
)

// verdictPhases lists the verdict-file families in decision order: once a card
// has a sync-audit file, the newest of those decides; before that, the plan-audit
// files do (the split `factory rebind-evidence` makes by stage).
var verdictPhases = []string{"sync-audit", "plan-audit"}

// deltaFullNote is the note of a call that asked for the full scope.
func deltaFullNote() string {
	return deltaDefaultPrefix + `none - delta_base "` + deltaBaseFull + `" requested: the full resolved base scope is reviewed`
}

// newestAuditVerdictFile returns the card's newest exported verdict file by the
// rebind-evidence selection (highest -iter<N>, then latest modification time),
// applied to the first verdict family that has any file. "" when there is none.
func newestAuditVerdictFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, phase := range verdictPhases {
		var paths []string
		for _, e := range entries {
			if e.Type().IsRegular() && strings.HasPrefix(e.Name(), phase) && strings.HasSuffix(e.Name(), ".md") {
				paths = append(paths, filepath.Join(dir, e.Name()))
			}
		}
		if len(paths) == 0 {
			continue
		}
		if newest, nerr := homestate.NewestVerdictFile(paths); nerr == nil {
			return newest
		}
	}
	return ""
}

// defaultDeltaProposal returns the SHA to propose as the delta base (full commit
// SHA, or "" for none) and the note that names the verdict file and SHA it
// proposed or states why it proposed none.
//
// @MX:NOTE: [AUTO] only a proposal: the chain gate (evaluateDeltaChain) still decides whether the delta is reviewed; this function never admits one.
// @MX:SPEC: SPEC-AUDIT-ENGINE-CONVERGENCE-001
func defaultDeltaProposal(root, card string) (sha, note string) {
	none := func(format string, a ...any) (string, string) {
		return "", deltaDefaultPrefix + "none - " + fmt.Sprintf(format, a...)
	}
	if !secondReviewCardRe.MatchString(card) {
		return none("card_id %q is not a valid card id", card)
	}
	home, err := closure.ResolveEvidenceHome(root, card)
	if err != nil {
		return none("the card evidence home cannot be resolved: %v", err)
	}
	file := newestAuditVerdictFile(closure.EvidenceFor(home, card).Dir)
	if file == "" {
		return none("the card has no audit verdict file")
	}
	name := filepath.Base(file)
	raw, ok := readBoundedEvidence(file)
	if !ok {
		return none("the newest verdict file %s cannot be read in full", name)
	}
	f := auditverdict.Parse(raw)
	if hasString(f.DuplicateKeys, "verdict") {
		return none("the newest verdict file %s is not a FAIL (it carries conflicting verdict lines)", name)
	}
	if !strings.EqualFold(strings.TrimSpace(f.Label), "FAIL") {
		return none("the newest verdict file %s is not a FAIL (verdict: %s)", name, strings.TrimSpace(f.Label))
	}
	shas := auditedSHAs(raw)
	switch {
	case len(shas) == 0:
		return none("the newest verdict file %s carries no audited_sha line", name)
	case len(shas) > 1:
		return none("the newest verdict file %s carries conflicting audited_sha lines", name)
	}
	full, rerr := runReviewGit(root, "rev-parse", "--verify", "--quiet", shas[0]+"^{commit}")
	full = strings.TrimSpace(full)
	if rerr != nil || full == "" {
		return none("the audited_sha %s of %s is not a commit of the audited tree, so it is not an ancestor of HEAD", shas[0], name)
	}
	if anc, aerr := gitio.IsAncestor(root, full, "HEAD"); aerr != nil || !anc {
		return none("the audited_sha %s of %s is not an ancestor of HEAD", full, name)
	}
	return full, deltaDefaultPrefix + "proposed " + full + " from " + name + " (the audited_sha of the card's newest FAIL verdict)"
}
