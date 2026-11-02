package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M1 — shared fixture for the delta chain
// gate tests (AC-AEC-001c, 001d, 001e, 001f; card t2017).
//
// "Kickoff not passed - pre-implementation RED": these tests are written ahead
// of the implementation and may need rework if the SPEC changes.
//
// The world is built from the production writers wherever one exists: the
// second-review record comes from appendSecondReviewRecord, the receipt from
// auditreceipt.WriteReceipt, and the call goes through the real audit_multi
// handler with a fake backend seam. Nothing here adds or stubs a production
// symbol.
//
// ASSUMED WIRE NAMES. The SPEC (REQ-AEC-004/005) fixes what a record and the
// confirmation must carry, but not the JSON key names. The constants below are
// the only place this file commits to a name; EV-04 names `blocking_closed`
// for the confirmation, the record keys follow the plan.md M1 wording ("delta
// marker, delta base, extended receipt id, confirmed ids, relied-on
// verdict-file digests"). If the implementation or a SPEC amendment picks other
// names, change them here and nowhere else. Every test that depends on one of
// them says so in its doc comment.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
)

const (
	// aecConfirmArg is the audit_multi argument carrying the BLOCKING-closure
	// confirmation: a JSON object mapping each id to its closing commit SHA.
	aecConfirmArg = "blocking_closed"

	// Additive second-review record keys (REQ-AEC-005), all top-level.
	aecRecDeltaKey     = "delta"           // delta marker (bool true)
	aecRecDeltaBaseKey = "delta_base"      // the delta base (string)
	aecRecExtendsKey   = "extends_receipt" // audit receipt id of the extended record
	aecRecConfirmedKey = "confirmed_ids"   // ids the confirmation named ([]string)
	aecRecDigestsKey   = "verdict_digests" // relied-on FAIL file -> SHA-256 (object)

	aecCard = closuretest.Card
)

// aecWorld is one card tree with its evidence directory, receipt store and
// second-review records.
type aecWorld struct {
	t       *testing.T
	f       *closuretest.Fixture
	tree    string // canonical audited tree root (what the handler receives)
	store   string // receipt store root of that tree
	session string
	n       int
}

func newAECWorld(t *testing.T) *aecWorld {
	t.Helper()
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{aecCard: closuretest.SpecID})
	tree := auditreceipt.Canonical(f.CardDir)
	store, err := auditreceipt.StoreRoot(tree)
	if err != nil {
		t.Fatalf("receipt store root for %s: %v", tree, err)
	}
	return &aecWorld{t: t, f: f, tree: tree, store: store, session: "aec-sess"}
}

// commit adds one commit to the card tree and returns the new HEAD.
func (w *aecWorld) commit(msg string) string {
	w.t.Helper()
	w.n++
	w.f.Write(filepath.Join(w.f.CardDir, "src", fmt.Sprintf("aec%d.txt", w.n)), msg+"\n")
	w.f.Git(w.f.CardDir, "add", "src")
	w.f.Git(w.f.CardDir, "commit", "-m", msg)
	return w.f.Head()
}

func (w *aecWorld) head() string { return w.f.Head() }

// parent returns the first parent of sha.
func (w *aecWorld) parent(sha string) string {
	w.t.Helper()
	return w.f.GitOut(w.f.CardDir, "rev-parse", sha+"^")
}

// aecReceipt describes the receipt a record cites; the zero value is a good one.
type aecReceipt struct {
	absent    bool   // cite an id that was never written to the store
	tool      string // default audit_multi
	treeRoot  string // default the audited tree
	head      string // default the record's head SHA
	gateUnmet string // default empty
	codex     string // default pass
}

func (w *aecWorld) writeReceipt(recHead string, o aecReceipt) string {
	w.t.Helper()
	if o.absent {
		id, err := auditreceipt.NewReceiptID()
		if err != nil {
			w.t.Fatalf("receipt id: %v", err)
		}
		return id
	}
	r := auditreceipt.Receipt{
		Tool:         o.tool,
		TreeRoot:     o.treeRoot,
		RootSource:   auditreceipt.RootSourceArgument,
		CodexVerdict: o.codex,
		GateUnmet:    o.gateUnmet,
		HeadCommit:   o.head,
		TreeDigest:   "aec-tree-digest",
	}
	if r.Tool == "" {
		r.Tool = auditreceipt.ToolAuditMulti
	}
	if r.TreeRoot == "" {
		r.TreeRoot = w.tree
	}
	if r.CodexVerdict == "" {
		r.CodexVerdict = "pass"
	}
	if r.HeadCommit == "" {
		r.HeadCommit = recHead
	}
	id, err := auditreceipt.WriteReceipt(w.store, &r)
	if err != nil {
		w.t.Fatalf("write receipt: %v", err)
	}
	return id
}

// aecRecOpts describes one second-review record; the zero value is a
// full-scope record with a pass from claude, codex and glm and a good receipt.
type aecRecOpts struct {
	deltaBase string            // non-empty: a delta record from this base
	extends   string            // receipt id of the record it extends
	confirmed []string          // ids the record's confirmation named
	digests   map[string]string // relied-on FAIL file -> SHA-256
	backends  map[string]string // backend -> verdict; nil = all three pass; a missing key = no entry
	receipt   aecReceipt
}

type aecRec struct{ head, receipt string }

var aecAllPass = map[string]string{BackendClaude: "pass", BackendCodex: "pass", BackendGLM: "pass"}

// record files one second-review record at the current HEAD, written by the
// production writer, then (for a delta record) patched to carry the additive
// keys and the delta base as its scope base.
func (w *aecWorld) record(o aecRecOpts) aecRec {
	w.t.Helper()
	head := w.f.Head()
	rid := w.writeReceipt(head, o.receipt)
	backends := o.backends
	if backends == nil {
		backends = aecAllPass
	}
	var pbv []PerBackendVerdict
	for _, name := range []string{BackendClaude, BackendCodex, BackendGLM} {
		if v, ok := backends[name]; ok {
			pbv = append(pbv, PerBackendVerdict{Backend: name, Gate: "required", Verdict: v})
		}
	}
	res := ConvergenceResult{PerBackendVerdicts: pbv, ParticipantCount: len(pbv), AuditReceipt: rid}
	if err := appendSecondReviewRecord(MultiAuditConfig{ProjectRoot: w.tree, CardID: aecCard}, res, codexTargetBaseBranch); err != nil {
		w.t.Fatalf("appendSecondReviewRecord: %v", err)
	}
	if o.deltaBase != "" {
		w.patchLastRecord(func(m map[string]any) {
			scope, _ := m["scope"].(map[string]any)
			if scope == nil {
				scope = map[string]any{}
			}
			scope["base_sha"] = o.deltaBase
			if n, err := gitio.ChangedFileCount(w.f.CardDir, o.deltaBase, head); err == nil {
				scope["changed_files"] = n
			}
			m["scope"] = scope
			m[aecRecDeltaKey] = true
			m[aecRecDeltaBaseKey] = o.deltaBase
			m[aecRecExtendsKey] = o.extends
			m[aecRecConfirmedKey] = o.confirmed
			if o.digests != nil {
				m[aecRecDigestsKey] = o.digests
			}
		})
	}
	return aecRec{head: head, receipt: rid}
}

func (w *aecWorld) recordsPath() string {
	return filepath.Join(w.f.EvidenceDir, closure.SecondReviewFile)
}

func (w *aecWorld) patchLastRecord(mutate func(map[string]any)) {
	w.t.Helper()
	data, err := os.ReadFile(w.recordsPath())
	if err != nil {
		w.t.Fatalf("read records: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &m); err != nil {
		w.t.Fatalf("decode last record: %v", err)
	}
	mutate(m)
	b, err := json.Marshal(m)
	if err != nil {
		w.t.Fatalf("encode record: %v", err)
	}
	lines[len(lines)-1] = string(b)
	if err := os.WriteFile(w.recordsPath(), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		w.t.Fatalf("write records: %v", err)
	}
}

// keepRecords rewrites the record file keeping only the given 0-based lines.
func (w *aecWorld) keepRecords(keep ...int) {
	w.t.Helper()
	data, err := os.ReadFile(w.recordsPath())
	if err != nil {
		w.t.Fatalf("read records: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var out []string
	for _, i := range keep {
		out = append(out, lines[i])
	}
	if err := os.WriteFile(w.recordsPath(), []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		w.t.Fatalf("write records: %v", err)
	}
}

// records reads the record file back as generic maps (nil when absent).
func (w *aecWorld) records() []map[string]any {
	w.t.Helper()
	if _, err := os.Stat(w.recordsPath()); err != nil {
		return nil
	}
	return readSecondReviewLines(w.t, w.recordsPath())
}

func aecN(n int) *int { return &n }

// verdictFile writes a FAIL verdict file into the card evidence directory.
// count nil omits the blocking_count line; ids become blocking_defect lines.
func (w *aecWorld) verdictFile(name, audited string, count *int, ids ...string) string {
	w.t.Helper()
	body := "# audit verdict\n\nverdict: FAIL\naudited_sha: " + audited + "\n"
	if count != nil {
		body += fmt.Sprintf("blocking_count: %d\n", *count)
	}
	for _, id := range ids {
		body += "blocking_defect: " + id + "\n"
	}
	path := filepath.Join(w.f.EvidenceDir, name)
	w.f.Write(path, body)
	return path
}

func aecFileSHA256(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// aecCallOpts is one audit_multi call.
type aecCallOpts struct {
	deltaBase string
	confirm   map[string]any // nil omits the confirmation argument
	extra     map[string]any
	session   string
	verdicts  map[string]string // backend -> verdict for the fake legs; default pass
	noCard    bool              // omit card_id (the t1873 path)
}

// aecOut is what one call showed: which legs ran, the delta base each saw on
// its context (the seam every real leg measures from), and the decoded result.
type aecOut struct {
	ran       map[string]bool
	deltaSeen map[string]string
	raw       string
	result    ConvergenceResult
	isErr     bool
}

// call drives the real audit_multi handler over the card tree. The required set
// of the call is {codex, glm}: claude is gated off.
func (w *aecWorld) call(c aecCallOpts) aecOut {
	w.t.Helper()
	var mu sync.Mutex
	out := aecOut{ran: map[string]bool{}, deltaSeen: map[string]string{}}
	prev := backendCall
	backendCall = func(ctx context.Context, backend, _, _, _ string) ReviewOutput {
		mu.Lock()
		defer mu.Unlock()
		out.ran[backend] = true
		out.deltaSeen[backend] = deltaBaseFromContext(ctx)
		v := "pass"
		if x, ok := c.verdicts[backend]; ok {
			v = x
		}
		return ReviewOutput{Verdict: v, Summary: backend + ":" + v, Findings: []Finding{}, NextSteps: []string{}}
	}
	defer func() { backendCall = prev }()

	session := c.session
	if session == "" {
		session = w.session
	}
	args := map[string]any{
		"target":       codexTargetBaseBranch,
		"project_root": w.tree,
		"card_id":      aecCard,
		"session_id":   session,
		"gates":        map[string]any{"claude": "off", "codex": "required", "glm": "required"},
	}
	if c.noCard {
		delete(args, "card_id")
	}
	if c.deltaBase != "" {
		args["delta_base"] = c.deltaBase
	}
	if c.confirm != nil {
		args[aecConfirmArg] = c.confirm
	}
	for k, v := range c.extra {
		args[k] = v
	}
	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		w.t.Fatalf("handleAuditMulti: %v", err)
	}
	out.isErr = res.IsError
	for _, ct := range res.Content {
		if tc, ok := ct.(mcp.TextContent); ok {
			out.raw += tc.Text
		}
	}
	if !out.isErr {
		if err := json.Unmarshal([]byte(out.raw), &out.result); err != nil {
			w.t.Fatalf("decode audit_multi result: %v\n%s", err, out.raw)
		}
	}
	return out
}

func aecClosure(pairs ...string) map[string]any {
	m := map[string]any{}
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m
}

// aecChain is the happy chain every gate test starts from.
//
//	h0: R0 (full scope) ... F0 = plan-audit-1.md FAIL, audited h0, names B1
//	h1: R1 (delta h0..h1, confirmed B1)     (k1 closed B1)
//	h2: HEAD, the call point                (k2 is a later change)
type aecChain struct {
	h0, h1, h2 string
	k1, k2     string // closing commits
	r0, r1     aecRec
	f0         string // F0 path
}

// happyChain builds R0, R1 and F0 and leaves HEAD at h2. Options adjust the
// parts a row varies; the zero value is the unbroken, reviewed chain.
type aecChainOpts struct {
	r0, r1      aecRecOpts
	f0Ids       []string // default B1
	f0Count     *int     // default len(f0Ids)
	f0NoCount   bool
	noF0        bool
	r1Confirmed []string // default f0Ids

	// r1BaseParentOfH0 gives R1 a delta base that is a valid ancestor of HEAD but
	// is not R0's head: a gap in the chain with an otherwise valid last link.
	r1BaseParentOfH0 bool
}

func (w *aecWorld) happyChain(o aecChainOpts) aecChain {
	w.t.Helper()
	c := aecChain{}
	c.h0 = w.head()
	ids := o.f0Ids
	if ids == nil {
		ids = []string{"B1"}
	}
	if !o.noF0 {
		count := o.f0Count
		if count == nil {
			count = aecN(len(ids))
		}
		if o.f0NoCount {
			count = nil
		}
		c.f0 = w.verdictFile("plan-audit-1.md", c.h0, count, ids...)
	}
	c.r0 = w.record(o.r0)
	c.k1 = w.commit("close " + strings.Join(ids, ","))
	c.h1 = c.k1
	r1 := o.r1
	r1.deltaBase = c.h0
	if o.r1BaseParentOfH0 {
		r1.deltaBase = w.parent(c.h0)
	}
	r1.extends = c.r0.receipt
	r1.confirmed = o.r1Confirmed
	if r1.confirmed == nil {
		r1.confirmed = ids
	}
	if c.f0 != "" && r1.digests == nil {
		r1.digests = map[string]string{filepath.Base(c.f0): aecFileSHA256(w.t, c.f0)}
	}
	c.r1 = w.record(r1)
	c.k2 = w.commit("later change")
	c.h2 = c.k2
	return c
}

// wantScope asserts which scope every leg measured: want "" is the full scope,
// otherwise the delta base the legs must have seen on their context.
func wantScope(t *testing.T, out aecOut, want string) {
	t.Helper()
	if out.isErr {
		t.Fatalf("the call returned a tool error: %s", out.raw)
	}
	for _, b := range []string{BackendCodex, BackendGLM} {
		if !out.ran[b] {
			t.Errorf("required leg %s did not run", b)
			continue
		}
		if got := out.deltaSeen[b]; got != want {
			scope := "full scope"
			if want != "" {
				scope = "delta " + want
			}
			t.Errorf("leg %s measured delta base %q, want %q (%s)", b, got, want, scope)
		}
	}
}

// wantNames asserts the result text names at least one of the given commits (by
// its first 7 hex digits). ASSUMED carrier: the SPEC says "the result names the
// failing link" without fixing a field, so the whole result JSON is searched.
func wantNames(t *testing.T, out aecOut, shas ...string) {
	t.Helper()
	for _, s := range shas {
		if len(s) >= 7 && strings.Contains(out.raw, s[:7]) {
			return
		}
	}
	t.Errorf("the full-scope result does not name the failing link (looked for any of %v); result: %s", shas, out.raw)
}
