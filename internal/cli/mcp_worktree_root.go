package cli

// mcp_worktree_root.go — SPEC-MCP-WORKTREE-UNTRACKED-001.
//
// A repository that keeps .moai/ untracked gives every linked worktree a tree
// with no .moai of its own. validateProjectRoot's existing test (".moai is a
// directory") rejects such a tree, and omitting project_root silently acts on
// the primary checkout. This file holds the second, git-backed acceptance
// branch, the git-free "config-orphaned" predicate, and the audit-gate routing
// that keeps the primary's explicit workflow.audit.gates binding on such a tree.
//
// Every git inspection here runs scrubbed (no inherited GIT_DIR / GIT_WORK_TREE
// / GIT_COMMON_DIR / GIT_INDEX_FILE / GIT_CEILING_DIRECTORIES, LC_ALL=C), is
// decided by exit status and output shape — never message text — and is never
// retried through an alternative invocation: an inspection that cannot
// complete is a rejection (validator) or a fail-closed gate (gate read).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// The git inspection helpers, the config-orphaned predicate, and primary
// identification live in internal/auditreceipt so the hook package reaches the
// same answer (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-001). These names keep the
// call sites of this package unchanged.
var runScrubbedGit = auditreceipt.RunScrubbedGit

func singleGitPath(out string) (string, error) { return auditreceipt.SingleGitPath(out) }

func identifyPrimaryCheckout(dir string) (string, []auditreceipt.WorktreeEntry, error) {
	return auditreceipt.IdentifyPrimaryCheckout(dir)
}

func isConfigOrphanedRoot(root string) bool { return auditreceipt.IsConfigOrphanedRoot(root) }

// gateAssumedRequiredNote is the gate_unmet / residual-note wording for a
// config-orphaned root whose primary checkout could not be identified
// (REQ-MWU-012), distinguishable from a primary that declares `required`.
const gateAssumedRequiredNote = auditreceipt.GateAssumedRequiredNote

// validateLinkedWorktreeRoot is the second acceptance branch of
// validateProjectRoot, reached only when canonical has no .moai directory
// (REQ-MWU-002..007). It accepts canonical when it is the top level of a linked
// worktree listed without a prunable mark, in a repository whose primary
// checkout has a .moai directory; every other shape is rejected with the
// failed condition named. raw is the caller's spelling, used in messages.
func validateLinkedWorktreeRoot(raw, canonical string) (string, error) {
	reject := func(reason string) (string, error) {
		return "", fmt.Errorf("project_root %q has no .moai directory and is not an accepted linked worktree: %s", raw, reason)
	}

	out, err := runScrubbedGit(canonical, "rev-parse", "--path-format=absolute", "--show-toplevel")
	if err != nil {
		return reject("git could not inspect it as a worktree (not a git repository, an unregistered or unreadable worktree, or git unavailable)")
	}
	top, err := singleGitPath(out)
	if err != nil {
		return reject("git could not inspect it as a worktree (unexpected output)")
	}
	if top != canonical {
		return reject("it is not the top level of a worktree (a subdirectory of " + top + ")")
	}

	primary, entries, err := identifyPrimaryCheckout(canonical)
	if err != nil {
		return reject(err.Error())
	}
	if primary == canonical {
		return reject("it is the repository's primary checkout, which has no .moai directory")
	}
	if info, statErr := os.Stat(filepath.Join(primary, ".moai")); statErr != nil || !info.IsDir() {
		return reject("its primary checkout " + primary + " has no .moai directory")
	}
	for _, e := range entries[1:] {
		if e.Path != canonical {
			continue // other entries — stale or not — are not a reason to reject
		}
		if e.Prunable {
			return reject("it is listed as a prunable worktree")
		}
		return canonical, nil
	}
	return reject("it is not a registered worktree of " + primary)
}

// resolveAuditGates returns the workflow.audit.gates block that governs an
// audit of root (REQ-MWU-011/012). A root that is not config-orphaned reads its
// own workflow.yaml exactly as before and runs no git. A config-orphaned root
// reads the gate of its primary checkout; when that primary cannot be
// identified the codex gate is treated as `required` and assumedNote says why.
func resolveAuditGates(root string) (gates config.AuditGates, assumedNote string) {
	if !isConfigOrphanedRoot(root) {
		return workflowAuditPins(root).Gates, ""
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return config.AuditGates{Codex: config.AuditGateRequired}, gateAssumedRequiredNote
	}
	return workflowAuditPins(primary).Gates, ""
}

// receiptCodexGateRequired is the receipt-id exposure read of recordAuditReceipt,
// routed the same way. auditreceipt.CodexGateRequired itself is unchanged (the
// hook-side receipt guard keeps reading the session tree's own config).
func receiptCodexGateRequired(root string) bool {
	if !isConfigOrphanedRoot(root) {
		return auditreceipt.CodexGateRequired(root)
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return true
	}
	return auditreceipt.CodexGateRequired(primary)
}

// withRootBlock returns data re-shaped as a JSON object carrying an added
// "_root" key, with every existing field kept verbatim (numbers are decoded as
// json.Number so their text does not change). Data that does not marshal to a
// JSON object is returned unchanged.
func withRootBlock(data any, rootBlock map[string]any) any {
	b, err := json.Marshal(data)
	if err != nil {
		return data
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil || m == nil {
		return data
	}
	m["_root"] = rootBlock
	return m
}

// worktreeWarning is the _root.worktree_warning text on a config-orphaned
// worktree (REQ-MWU-013, reworded by SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-015):
// the answer now comes from the union catalogue and the primary's state store,
// and `sources` lists what was actually read.
const worktreeWarning = "project_root is a linked worktree whose repository does not track .moai; " +
	"the SPEC catalogue answered here is the union of this worktree's and its primary checkout's .moai/specs " +
	"(each record names its source), and state is kept in the primary checkout's .moai/state under this " +
	"worktree's tree identity — `sources` lists what was actually read"

// Tool-description notes for a config-orphaned worktree (REQ-WSR-016).
const (
	worktreeCatalogueDescNote = " On a linked worktree of a repository that does not track .moai, the catalogue is " +
		"the union of the worktree's and the primary checkout's .moai/specs: each record and finding names its source " +
		"(worktree or primary), and a SPEC present in both is reported once, from the worktree, with the primary copy " +
		"named as shadowed. When that primary cannot be identified only the worktree is read and _root says so."
	worktreeStateDescNote = " On a linked worktree of a repository that does not track .moai, this tool's state is " +
		"kept in the primary checkout's .moai/state under the worktree's own tree identity; when that primary cannot " +
		"be identified nothing is written and the result says why."
)

// Catalogue source tags (REQ-WSR-011).
const (
	catalogueSourceWorktree = "worktree"
	catalogueSourcePrimary  = "primary"
)

// stateRootBlock is the _root block of a state tool (verify_snapshot,
// verify_trend): the provenance map, plus — on a config-orphaned root — the
// store root actually read (REQ-WSR-015).
func stateRootBlock(root, source, store string) map[string]any {
	prov := rootProvenanceMap(root, source)
	if isConfigOrphanedRoot(root) {
		prov["sources"] = []map[string]string{{"source": "store", "dir": store}}
	}
	return prov
}

// catalogueView names the catalogues a SPEC catalogue tool answers over. For a
// root that is not config-orphaned it is that root alone, exactly as before.
// For a config-orphaned worktree it is the worktree plus its primary checkout
// (REQ-WSR-011), or the worktree alone with primaryErr set when the primary
// cannot be identified (REQ-WSR-014).
type catalogueView struct {
	root       string
	primary    string
	orphaned   bool
	primaryErr error
}

func resolveCatalogueView(root string) catalogueView {
	if !isConfigOrphanedRoot(root) {
		return catalogueView{root: root}
	}
	primary, err := auditreceipt.StoreRoot(root)
	return catalogueView{root: root, primary: primary, orphaned: true, primaryErr: err}
}

// hasPrimary reports whether the primary catalogue is read.
func (v catalogueView) hasPrimary() bool { return v.orphaned && v.primaryErr == nil }

// rootBlock is the catalogue tool's _root block: the provenance map, plus on a
// config-orphaned root the catalogue sources actually read and, when the
// primary could not be identified, a statement that its catalogue was not read
// and why (REQ-WSR-014/015). The existing `warning` key is never touched.
func (v catalogueView) rootBlock(source string) map[string]any {
	prov := rootProvenanceMap(v.root, source)
	if !v.orphaned {
		return prov
	}
	sources := []map[string]string{{"source": catalogueSourceWorktree, "dir": v.root}}
	if v.hasPrimary() {
		sources = append(sources, map[string]string{"source": catalogueSourcePrimary, "dir": v.primary})
	} else {
		prov["primary_catalogue"] = "not read: " + v.primaryErr.Error()
	}
	prov["sources"] = sources
	return prov
}

// catalogueSpecIDs returns the SPEC directory names under root/.moai/specs.
func catalogueSpecIDs(root string) map[string]bool {
	ids := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "specs"))
	if err != nil {
		return ids
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "SPEC-") {
			ids[e.Name()] = true
		}
	}
	return ids
}

// shadowedPrimaryCopies names the primary copies of SPEC IDs present in both
// catalogues; the worktree copy is the one reported (REQ-WSR-012).
func (v catalogueView) shadowedPrimaryCopies() []map[string]string {
	out := []map[string]string{}
	if !v.hasPrimary() {
		return out
	}
	primaryIDs := catalogueSpecIDs(v.primary)
	ids := []string{}
	for id := range catalogueSpecIDs(v.root) {
		if primaryIDs[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		out = append(out, map[string]string{
			"spec_id": id,
			"source":  catalogueSourcePrimary,
			"path":    filepath.Join(v.primary, ".moai", "specs", id),
		})
	}
	return out
}

// toJSONMap re-shapes a value as a JSON object map (numbers kept verbatim).
func toJSONMap(v any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

// unionSpecDocs answers spec_progress over the union catalogue: every worktree
// record tagged `worktree`, every primary record whose SPEC ID the worktree
// does not also carry tagged `primary` (REQ-WSR-011/012/013).
func (v catalogueView) unionSpecDocs() ([]map[string]any, error) {
	wRecs, err := spec.ListDocs(v.root)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(wRecs))
	seen := map[string]bool{}
	for _, rec := range wRecs {
		m := toJSONMap(rec)
		m["source"] = catalogueSourceWorktree
		out = append(out, m)
		seen[filepath.Base(filepath.Dir(rec.Path))] = true
	}
	if !v.hasPrimary() {
		return out, nil
	}
	pRecs, err := spec.ListDocs(v.primary)
	if err != nil {
		return nil, err
	}
	for _, rec := range pRecs {
		if seen[filepath.Base(filepath.Dir(rec.Path))] {
			continue // shadowed by the worktree copy
		}
		m := toJSONMap(rec)
		m["source"] = catalogueSourcePrimary
		out = append(out, m)
	}
	return out, nil
}

// unionAudit runs spec.Audit per catalogue and merges the results: the counts
// add, the primary run excludes every SPEC ID the worktree carries (so a
// shadowed SPEC counts once), and every finding names its source
// (REQ-WSR-011/012).
func (v catalogueView) unionAudit(opts spec.AuditOptions) (*spec.AuditResult, []map[string]any, error) {
	opts.BaseDir = v.root
	wRes, err := spec.Audit(opts)
	if err != nil {
		return nil, nil, err
	}
	merged := *wRes
	findings := tagFindings(wRes.DriftFindings, catalogueSourceWorktree, nil)
	if v.hasPrimary() {
		worktreeIDs := catalogueSpecIDs(v.root)
		pOpts := opts
		pOpts.BaseDir = v.primary
		pOpts.ExcludeSpecs = worktreeIDs
		pRes, err := spec.Audit(pOpts)
		if err != nil {
			return nil, nil, err
		}
		merged.TotalSpecs += pRes.TotalSpecs
		merged.Grandfathered += pRes.Grandfathered
		merged.ModernEraClean += pRes.ModernEraClean
		if merged.TotalSpecs != 1 {
			merged.TokensSpent = nil
		} else if wRes.TotalSpecs == 0 {
			merged.TokensSpent = pRes.TokensSpent
		}
		findings = append(findings, tagFindings(pRes.DriftFindings, catalogueSourcePrimary, worktreeIDs)...)
	}
	merged.DriftFindings = nil
	return &merged, findings, nil
}

// tagFindings renders findings as JSON maps carrying their catalogue source,
// dropping any whose SPEC ID is in skip (a copy shadowed by the worktree).
func tagFindings(fs []spec.DriftFinding, source string, skip map[string]bool) []map[string]any {
	out := make([]map[string]any, 0, len(fs))
	for _, f := range fs {
		if skip[f.SpecID] {
			continue
		}
		m := toJSONMap(f)
		m["source"] = source
		out = append(out, m)
	}
	return out
}
