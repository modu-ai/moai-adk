package cli

// audit_delta_chain.go — the delta chain gate of SPEC-AUDIT-ENGINE-CONVERGENCE-001
// (REQ-AEC-004). An audit_multi call that names a card_id and proposes a delta
// base reviews the delta scope only when (i) the card's server-written
// second-review records form an unbroken, reviewed chain ending at the proposed
// base, and (ii) the call's BLOCKING-closure confirmation is complete. Anything
// else is a full-scope review, and the result says which condition failed.
//
// Nothing here is read from a call argument except the confirmation itself: the
// chain comes from second-review.jsonl and the receipt store, the closure set
// from the FAIL verdict files in the card evidence directory. A chain that could
// be supplied by the caller, or rebuilt from verdict files alone, would let any
// caller buy a delta review.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/auditverdict"
	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// deltaGatePrefix opens the reason a refused delta carries in the result.
	deltaGatePrefix = "delta gate: full scope reviewed - "

	// maxVerdictFileBytes bounds one evidence file the gate reads.
	maxVerdictFileBytes = 1 << 20

	// maxClosureIDBytes bounds one confirmation id (the id is echoed into the
	// legs' focus text).
	maxClosureIDBytes = 128

	// deltaNoteNoRecord is the tail deltaScopeNote appends to a delta call's
	// note. It is true for a delta the gate did not admit, and false for one it
	// did; admittedDeltaNote swaps it.
	deltaNoteNoRecord = "; no second-review record and no session state written"
	// deltaNoteChainVerified replaces deltaNoteNoRecord on an admitted delta.
	deltaNoteChainVerified = "; chain-verified delta: one second-review record filed and the result persisted"
)

// auditedSHALine matches the machine line a verdict file carries, the same
// shape the factory card reader accepts.
var auditedSHALine = regexp.MustCompile(`^audited_sha: ([0-9a-f]{12,40})$`)

// deltaChainInput is what the gate decides on.
type deltaChainInput struct {
	Root     string            // the audited tree
	Card     string            // the card id (already a trust-boundary input)
	Proposed string            // the proposed delta base
	Required []string          // backends the call's plan resolves as required
	Closure  map[string]string // the confirmation: id -> closing commit ("" = inherited)
}

// deltaAdmission is what an admitted delta call files and forwards.
type deltaAdmission struct {
	Extends   string            // audit receipt id of the chain's last record
	Confirmed []string          // the closure-set ids, sorted
	Digests   map[string]string // relied-on FAIL verdict file -> SHA-256
	Closing   map[string]string // id -> closing commit as confirmed (may be "")
}

// requiredBackends lists the backends a call's gates resolve as required, with
// the same defaults runMultiAudit applies when it builds its fan-out.
func requiredBackends(g config.AuditGates) []string {
	var out []string
	for _, b := range []struct{ name, gate, dflt string }{
		{BackendClaude, g.Claude, config.AuditGateRequired},
		{BackendCodex, g.Codex, config.AuditGateRequired},
		{BackendGLM, g.GLM, config.AuditGateAdvisory},
	} {
		if gateOr(b.gate, b.dflt) == config.AuditGateRequired {
			out = append(out, b.name)
		}
	}
	return out
}

// evaluateDeltaChain applies REQ-AEC-004. It returns the admission, or nil and
// the reason the delta is refused (always starting with deltaGatePrefix). Every
// uncertainty refuses: a gate that cannot read its inputs reviews the full scope.
//
// @MX:NOTE: [AUTO] the only place a delta review is bought; the guarantee "a delta passes only on a verified chain" holds on the card_id path only.
func evaluateDeltaChain(in deltaChainInput) (*deltaAdmission, string) {
	refuse := func(format string, a ...any) (*deltaAdmission, string) {
		return nil, deltaGatePrefix + fmt.Sprintf(format, a...)
	}
	if !secondReviewCardRe.MatchString(in.Card) {
		return refuse("card_id %q is not a valid card id", in.Card)
	}
	head, err := gitio.Head(in.Root)
	if err != nil {
		return refuse("HEAD of the audited tree cannot be read: %v", err)
	}
	proposed := in.Proposed
	if full, rerr := runReviewGit(in.Root, "rev-parse", "--verify", "--quiet", proposed+"^{commit}"); rerr == nil && strings.TrimSpace(full) != "" {
		proposed = strings.TrimSpace(full)
	}

	home, err := closure.ResolveEvidenceHome(in.Root, in.Card)
	if err != nil {
		return refuse("the card evidence home cannot be resolved: %v", err)
	}
	paths := closure.EvidenceFor(home, in.Card)
	records, _, err := closure.LoadSecondReviews(paths.SecondReview)
	if err != nil {
		return refuse("the card's second-review records cannot be read: %v", err)
	}

	// (i) the chain: the newest full-scope record starts it; every later record
	// is a delta from the previous record's head.
	start := -1
	for i, r := range records {
		if !r.Delta && r.Target == closure.TargetBaseBranch {
			start = i
		}
	}
	if start < 0 {
		return refuse("no full-scope second-review record starts a chain for card %s", in.Card)
	}
	chain := records[start:]
	for i, r := range chain {
		who := shortSHA(r.HeadSHA)
		if r.HeadSHA == "" {
			return refuse("a chain record (%d of %d) carries no head commit", i+1, len(chain))
		}
		if i > 0 {
			prev := chain[i-1].HeadSHA
			if !r.Delta || r.Target != closure.TargetBaseBranch || r.DeltaBase != prev || r.Scope.BaseSHA != prev {
				return refuse("the chain is broken at record %s: its delta base %s is not the previous record's head %s", who, shortSHA(r.DeltaBase), shortSHA(prev))
			}
		}
		if why := recordReviewedRequired(r, in.Required); why != "" {
			return refuse("chain record %s: %s", who, why)
		}
		if why := recordReceiptRefusal(in.Root, r); why != "" {
			return refuse("chain record %s: %s", who, why)
		}
	}
	last := chain[len(chain)-1]
	if proposed != last.HeadSHA {
		return refuse("the proposed base %s is not the head of the chain's last record %s", shortSHA(proposed), shortSHA(last.HeadSHA))
	}

	// (ii) the closure set and the confirmation.
	files, why := tiedFailFiles(paths.Dir, chain)
	if why != "" {
		return refuse("%s", why)
	}
	if len(files) == 0 {
		return refuse("no FAIL verdict file is tied to the chain, so there is no closure basis")
	}
	latest := map[string]int{} // id -> chain index of its latest naming
	for _, f := range files {
		for _, id := range f.ids {
			if idx, ok := latest[id]; !ok || f.chainIdx > idx {
				latest[id] = f.chainIdx
			}
		}
	}
	union := make([]string, 0, len(latest))
	for id := range latest {
		union = append(union, id)
	}
	sort.Strings(union)
	var missing []string
	for _, id := range union {
		if _, ok := in.Closure[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return refuse("the %s confirmation does not name every BLOCKING id of the closure set (missing: %s)", "blocking_closed", strings.Join(missing, ", "))
	}
	for _, id := range union {
		idx := latest[id]
		// Inherited: the record that extended the latest naming's FAIL file
		// confirmed this id, so its closure was verified one link earlier.
		if idx+1 < len(chain) && hasString(chain[idx+1].ConfirmedIDs, id) {
			continue
		}
		if why := closingCommitRefusal(in.Root, head, proposed, chain[idx].HeadSHA, id, in.Closure[id]); why != "" {
			return refuse("%s", why)
		}
	}

	digests := map[string]string{}
	for _, f := range files {
		digests[f.name] = f.sha256
	}
	closing := map[string]string{}
	for _, id := range union {
		closing[id] = in.Closure[id]
	}
	return &deltaAdmission{Extends: last.AuditReceipt, Confirmed: union, Digests: digests, Closing: closing}, ""
}

// recordReviewedRequired reports why a record did not review every required
// backend. It iterates the required set, not the entries present, so an absent
// entry counts as not run.
func recordReviewedRequired(r closure.SecondReviewRecord, required []string) string {
	for _, name := range required {
		verdict, found := "", false
		for _, b := range r.Backends {
			if b.Backend == name {
				verdict, found = b.Verdict, true
				break
			}
		}
		switch {
		case !found:
			return "required backend " + name + " has no entry (not run)"
		case verdict != "pass" && verdict != "fail":
			return "required backend " + name + " recorded " + fmt.Sprintf("%q", verdict) + " (not run)"
		}
	}
	return ""
}

// recordReceiptRefusal cross-checks the record's receipt against the receipt
// store: it must exist, come from audit_multi on this tree, bind the record's
// head commit, and record a review that ran (no unmet gate, no inconclusive
// codex verdict).
func recordReceiptRefusal(root string, r closure.SecondReviewRecord) string {
	id := strings.TrimSpace(r.AuditReceipt)
	if id == "" {
		return "cites no audit receipt"
	}
	tree := auditreceipt.Canonical(root)
	store, err := auditreceipt.StoreRoot(tree)
	if err != nil {
		return "the receipt store cannot be resolved: " + err.Error()
	}
	rc, err := auditreceipt.ReadReceipt(store, id)
	switch {
	case err != nil:
		return "its receipt " + id + " is absent from the receipt store or unreadable"
	case rc.Tool != auditreceipt.ToolAuditMulti:
		return "its receipt " + id + " was recorded by " + rc.Tool + ", not audit_multi"
	case auditreceipt.Canonical(rc.TreeRoot) != tree:
		return "its receipt " + id + " names another tree root"
	case rc.HeadCommit != r.HeadSHA:
		return "its receipt " + id + " binds commit " + shortSHA(rc.HeadCommit) + ", not the record's head"
	case strings.TrimSpace(rc.GateUnmet) != "":
		return "its receipt " + id + " records an unmet gate"
	case strings.TrimSpace(rc.CodexVerdict) == VerdictInconclusive:
		return "its receipt " + id + " records an inconclusive codex verdict"
	}
	return ""
}

// failFile is one FAIL verdict file tied to a chain record.
type failFile struct {
	name     string
	ids      []string
	chainIdx int
	sha256   string
}

// tiedFailFiles reads every FAIL verdict file in the evidence directory whose
// audited_sha is the head of a chain record. A returned reason refuses the
// delta: a tied file with no usable blocking_count, or whose blocking_defect
// lines do not match it, cannot carry a closure set.
func tiedFailFiles(dir string, chain []closure.SecondReviewRecord) ([]failFile, string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ""
		}
		return nil, "the card evidence directory cannot be read: " + err.Error()
	}
	var out []failFile
	for _, e := range entries {
		if !e.Type().IsRegular() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, ok := readBoundedEvidence(filepath.Join(dir, e.Name()))
		if !ok {
			// A file the gate cannot read in full is not an empty closure set:
			// skipping it would let an oversized FAIL file hide its ids.
			return nil, "evidence file " + e.Name() + " cannot be read in full (unreadable or over the size bound)"
		}
		f := auditverdict.Parse(raw)
		if !strings.EqualFold(strings.TrimSpace(f.Label), "FAIL") || hasString(f.DuplicateKeys, "verdict") {
			continue
		}
		audited := singleAuditedSHA(raw)
		if audited == "" {
			continue
		}
		idx := -1
		for i, r := range chain {
			if strings.HasPrefix(r.HeadSHA, audited) {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		switch {
		case !f.BlockingKnown || hasString(f.DuplicateKeys, "blocking_count"):
			return nil, "FAIL verdict file " + e.Name() + " carries no single blocking_count"
		case len(f.BlockingDefects) != f.BlockingCount:
			return nil, fmt.Sprintf("FAIL verdict file %s lists %d blocking_defect lines but states blocking_count %d", e.Name(), len(f.BlockingDefects), f.BlockingCount)
		}
		sum := sha256.Sum256(raw)
		out = append(out, failFile{name: e.Name(), ids: uniqueStrings(f.BlockingDefects), chainIdx: idx, sha256: hex.EncodeToString(sum[:])})
	}
	return out, ""
}

// readBoundedEvidence reads a regular file of at most maxVerdictFileBytes.
func readBoundedEvidence(path string) ([]byte, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	raw, err := io.ReadAll(io.LimitReader(f, maxVerdictFileBytes+1))
	if err != nil || len(raw) > maxVerdictFileBytes {
		return nil, false
	}
	return raw, true
}

// singleAuditedSHA returns the one audited_sha a file carries, or "" when it
// carries none or two different ones (the file is then unreadable to the gate,
// as it is to the factory card reader).
func singleAuditedSHA(raw []byte) string {
	seen := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if m := auditedSHALine.FindStringSubmatch(strings.TrimRight(line, " \t\r")); m != nil {
			seen[m[1]] = true
		}
	}
	if len(seen) != 1 {
		return ""
	}
	for s := range seen {
		return s
	}
	return ""
}

// closingCommitRefusal checks the closing commit of an id that is not
// inherited: a commit inside proposed..HEAD that comes after the id's latest
// naming (namedAt, the head of the record the naming FAIL file is tied to).
func closingCommitRefusal(root, head, proposed, namedAt, id, commit string) string {
	commit = strings.TrimSpace(commit)
	if commit == "" {
		return "BLOCKING id " + id + " is new since the previous link and the confirmation gives no closing commit"
	}
	if !deltaBaseHex.MatchString(commit) {
		return "the closing commit given for " + id + " is not a commit SHA"
	}
	out, err := runReviewGit(root, "rev-parse", "--verify", "--quiet", commit+"^{commit}")
	full := strings.TrimSpace(out)
	if err != nil || full == "" {
		return "the closing commit " + shortSHA(commit) + " given for " + id + " is not a commit in the audited tree"
	}
	inHead, herr := gitio.IsAncestor(root, full, head)
	inBase, berr := gitio.IsAncestor(root, full, proposed)
	if herr != nil || berr != nil || !inHead || inBase {
		return "the closing commit " + shortSHA(full) + " given for " + id + " lies outside the delta " + shortSHA(proposed) + "..HEAD"
	}
	if after, aerr := gitio.IsAncestor(root, namedAt, full); aerr != nil || !after || full == namedAt {
		return "the closing commit " + shortSHA(full) + " given for " + id + " does not come after its latest naming at " + shortSHA(namedAt)
	}
	return ""
}

// admittedDeltaNote corrects the delta scope note for an admitted delta: the
// tail deltaScopeNote writes ("no second-review record ...") is only true for a
// delta the gate did not admit.
func admittedDeltaNote(note string) string {
	return strings.Replace(note, deltaNoteNoRecord, deltaNoteChainVerified, 1)
}

// confirmationFocus returns the focus text every leg of an admitted delta call
// receives: the caller's own focus, then the confirmation. The confirmation
// travels in the existing focus argument so each leg's request carries it.
func confirmationFocus(focus string, adm *deltaAdmission) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(focus))
	if b.Len() > 0 {
		b.WriteString("\n\n")
	}
	b.WriteString("BLOCKING-closure confirmation for this delta review (every id below was a BLOCKING defect of an earlier FAIL round; verify each is actually closed):")
	for _, id := range adm.Confirmed {
		b.WriteString("\n- " + id)
		if c := adm.Closing[id]; c != "" {
			b.WriteString(": closed by commit " + c)
		} else {
			b.WriteString(": closure confirmed in an earlier round")
		}
	}
	return b.String()
}

func hasString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
