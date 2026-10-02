package cli

// audit_plan_cmd.go — `moai verify audit-plan` (SPEC-AUDIT-MODEL-CONVERGE-001 M4,
// REQ-ACV-010..012 and REQ-ACV-016, design.md §D.4-§D.5).
//
// The verb prints the resolved audit backend plan of ONE tree as JSON, so an
// auditor learns which backends its audit must reach without interpreting the
// audit.model token itself. With --result-file it also compares that plan with an
// audit_multi result written to a file the caller names (--result-file <path>,
// M4b; the inline JSON argument of M4 is gone because the worktree-isolation guard
// refuses a command carrying braces or quotes) and reports which required gates
// the result leaves unmet.
//
// Both halves are read-only. The verb invokes no audit backend and writes no
// state, receipt, configuration or audit file. The checker is PURE: its only
// inputs are the resolved plan and the one file it is handed. It reads exactly
// that path once (regular files only, at most auditPlanResultMaxBytes), takes no
// session identifier and enumerates nothing under .moai/state, so a result cannot
// be a persisted file picked up from an earlier audit or a foreign tree.
//
// The configuration is read raw through loadWorkflowAuditSection — the loader
// that reports a failure — and never through a default-merged configuration: a
// merged read pairs the model token "claude" with the default gates and would
// hand "claude" to every project that never wrote one.
//
// @MX:NOTE: [AUTO] audit-plan verb — the read-only plan surface both auditors run before reaching a verdict; the old-binary signature in design §D.4 relies on the verb's config_status member or its audit-plan: stderr line, never on the exit code
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

func init() {
	verifyExtraCommands = append(verifyExtraCommands, newVerifyAuditPlanCmd)
}

const (
	// auditPlanMarker prefixes every error line the verb writes to stderr. It is
	// how a caller tells "the verb ran and refused" from "the verb does not exist
	// in this binary" (an unknown verb prints the verify group's help, exit 0).
	auditPlanMarker = "audit-plan:"

	// config_status values. unreadable is a distinct, non-passing state: the plan
	// is not guessed and not the default.
	auditPlanStatusOK         = "ok"
	auditPlanStatusAbsent     = "absent"
	auditPlanStatusUnreadable = "unreadable"

	// auditPlanUnknown is the tri-state member value of the unreadable state.
	auditPlanUnknown = "unknown"

	// Exit codes (internal/cli/CLAUDE.md): 1 user error, 2 system error.
	auditPlanExitUser   = 1
	auditPlanExitSystem = 2

	// per_backend_verdicts[].verdict values that answer a gate. They share the
	// overall verdict vocabulary; inconclusive is deliberately not one of them.
	auditVerdictPass = overallVerdictPass
	auditVerdictFail = overallVerdictFail
)

// auditPlanConvergenceCheck is the convergence_check object of the verb's output.
// Unmet names the enforced-required backends whose gate the result leaves unmet;
// Reason names every cause, a missing plan_source included.
type auditPlanConvergenceCheck struct {
	OK     bool     `json:"ok"`
	Unmet  []string `json:"unmet"`
	Reason string   `json:"reason,omitempty"`
}

// auditPlanOutput is the verb's stdout object (design.md §D.5).
type auditPlanOutput struct {
	ProjectRoot string `json:"project_root"`
	// ConfigRoot names the tree whose workflow.yaml was read when it is not
	// ProjectRoot: a config-orphaned worktree reads its primary checkout.
	ConfigRoot   string                  `json:"config_root,omitempty"`
	ConfigStatus string                  `json:"config_status"`
	Model        string                  `json:"model"`
	ModelSource  string                  `json:"model_source"`
	Backends     []config.AuditPlanEntry `json:"backends"`
	// CrossModelActive and CrossModelRequired are booleans, or the string
	// "unknown" when the configuration could not be read.
	CrossModelActive   any                        `json:"cross_model_active"`
	CrossModelRequired any                        `json:"cross_model_required"`
	EnforcedRequired   []string                   `json:"enforced_required"`
	Note               string                     `json:"note,omitempty"`
	ConvergenceCheck   *auditPlanConvergenceCheck `json:"convergence_check,omitempty"`
}

func newVerifyAuditPlanCmd(projectRoot *string) *cobra.Command {
	var resultFile string
	cmd := &cobra.Command{
		Use:   "audit-plan",
		Short: "Print the resolved audit backend plan of one tree (read-only); --result-file checks an audit_multi result file against it",
		Long: `Print the audit backend plan the tree's workflow.yaml resolves to, as one JSON
object: per backend (claude, codex, glm) the gate (off | advisory | required),
the source that decided it, and whether the operator wrote it; plus
cross_model_active, cross_model_required and enforced_required — the backends
whose unmet gate fails the verdict.

The verb is read-only: it invokes no audit backend and writes no state, receipt,
configuration or audit file.

IMPORTANT — name your own tree. --project-root defaults to $CLAUDE_PROJECT_DIR,
which in a worktree-isolated session names the PRIMARY checkout and would read
the wrong tree's configuration without saying so. A session working in a
worktree passes its own toplevel (git rev-parse --show-toplevel).

config_status is ok, absent (no workflow.yaml or no audit settings beyond the
backend pins: the distributed default plan) or unreadable (workflow.yaml exists
but cannot be read or parsed: cross_model_active is "unknown", no plan is
printed, and the state is never a pass).

--result-file <path> compares the plan with an audit_multi result held in a file
(the full result, or the digest of overall_verdict, gate_unmet, plan_source and,
per backend, backend, gate and verdict) and adds convergence_check {ok, unmet,
reason}. ok is true only when every enforced-required backend has a
per_backend_verdicts entry whose verdict is exactly "pass" or "fail" and whose
gate is "required", and — where any plan gate comes from the configuration — the
result's plan_source is "config".

The file is written by the auditor with its own Write tool, fresh for this check
(for example <toplevel>/.moai/state/audit-plan-result.json, overwritten each
time), and only the path is passed on the command line. The verb reads exactly
that path once: a regular file of at most 262144 bytes (256 KiB) holding one JSON
object. It takes no session id and reads nothing else under .moai/state. A file
that is missing, unreadable, a directory or pipe, empty, over the size bound, or
not a JSON object is an error, never a pass.

Output contract: stdout is one JSON object carrying config_status, or stderr
carries an "audit-plan:" line and stdout is empty. Exit 0 whenever a plan (or an
unreadable state) was reported; exit 1 for invalid configuration or an unusable
--result-file; exit 2 when no tree can be resolved.`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, _ []string) error {
			root, err := verifyResolveRoot(*projectRoot)
			if err != nil {
				return auditPlanFail(c, auditPlanExitSystem, "%v", err)
			}
			if info, statErr := os.Stat(root); statErr != nil || !info.IsDir() {
				return auditPlanFail(c, auditPlanExitSystem, "project root %q is not a directory", root)
			}

			out := auditPlanOutput{ProjectRoot: root}
			audit, readRoot, loadErr := loadAuditSectionForPlan(root)
			if loadErr != nil {
				out.ConfigStatus = auditPlanStatusUnreadable
				out.ModelSource = auditPlanUnknown
				out.Backends = []config.AuditPlanEntry{}
				out.CrossModelActive = auditPlanUnknown
				out.CrossModelRequired = auditPlanUnknown
				out.EnforcedRequired = []string{}
				out.Note = fmt.Sprintf("workflow.yaml unreadable, the plan is not guessed: %v", loadErr)
				if c.Flags().Changed("result-file") {
					out.ConvergenceCheck = &auditPlanConvergenceCheck{
						OK: false, Unmet: []string{},
						Reason: "the plan is unreadable, so there is nothing to compare the result with",
					}
				}
				return auditPlanEmit(c, out)
			}

			plan, err := config.ResolveAuditPlan(audit, config.AuditGates{})
			if err != nil {
				return auditPlanFail(c, auditPlanExitUser, "%v", err)
			}
			if readRoot != root {
				out.ConfigRoot = readRoot
			}
			out.ConfigStatus = auditPlanStatusAbsent
			if plan.Model != "" || plan.FromConfig() {
				out.ConfigStatus = auditPlanStatusOK
			}
			out.Model = plan.Model
			out.ModelSource = plan.ModelSource()
			out.Backends = plan.Backends
			out.EnforcedRequired = []string{}
			var active, required bool
			for _, e := range plan.Backends {
				if !e.Explicit {
					continue
				}
				if e.Gate == config.AuditGateRequired {
					out.EnforcedRequired = append(out.EnforcedRequired, e.Backend)
				}
				if e.Backend == auditPlanAnchorBackend {
					continue
				}
				active = active || e.Gate != config.AuditGateOff
				required = required || e.Gate == config.AuditGateRequired
			}
			out.CrossModelActive, out.CrossModelRequired = active, required

			if c.Flags().Changed("result-file") {
				data, readErr := readAuditResultFile(resultFile)
				if readErr != nil {
					return auditPlanFail(c, auditPlanExitUser, "%v", readErr)
				}
				check, checkErr := checkAuditResult(plan, string(data), auditEntryAnswered)
				if checkErr != nil {
					return auditPlanFail(c, auditPlanExitUser, "--result-file %s: %v", resultFile, checkErr)
				}
				out.ConvergenceCheck = &check
			}
			return auditPlanEmit(c, out)
		},
	}
	cmd.Flags().StringVar(&resultFile, "result-file", "", "path of a file the auditor wrote, fresh, holding an audit_multi result or its digest as one JSON object (at most 256 KiB); only the path is passed; adds convergence_check")
	return cmd
}

// auditPlanAnchorBackend is the in-session anchor; cross-model means any other
// backend taking part (design.md §D.5).
const auditPlanAnchorBackend = "claude"

// auditPlanFail writes the audit-plan: marker line to stderr and returns an
// error that carries the exit code; the root error handler renders nothing more
// for an exit-coded error, so the line is the whole output.
func auditPlanFail(c *cobra.Command, code int, format string, args ...any) error {
	msg := fmt.Sprintf("%s %s", auditPlanMarker, fmt.Sprintf(format, args...))
	_, _ = fmt.Fprintln(c.ErrOrStderr(), msg)
	return &exitCodeError{code: code, msg: msg}
}

func auditPlanEmit(c *cobra.Command, out auditPlanOutput) error {
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return auditPlanFail(c, auditPlanExitSystem, "marshal output: %v", err)
	}
	_, _ = fmt.Fprintln(c.OutOrStdout(), string(b))
	return nil
}

// loadAuditSectionForPlan reads the raw workflow.audit section that governs root
// through loadWorkflowAuditSection, routed as auditSectionForRoot routes it: a
// config-orphaned linked worktree reads its primary checkout. Unlike that
// helper it returns the failure, so an unreadable file is told apart from an
// absent one — and a primary that cannot be identified is a failure too, never a
// guess. readRoot names the tree actually read.
func loadAuditSectionForPlan(root string) (audit config.AuditConfig, readRoot string, err error) {
	readRoot = root
	if isConfigOrphanedRoot(root) {
		primary, _, perr := identifyPrimaryCheckout(root)
		if perr != nil {
			return config.AuditConfig{}, root, fmt.Errorf("%s has no workflow.yaml of its own and its primary checkout cannot be identified: %w", root, perr)
		}
		readRoot = primary
	}
	audit, err = loadWorkflowAuditSection(readRoot)
	if err != nil {
		path := filepath.Join(readRoot, ".moai", "config", "sections", "workflow.yaml")
		return config.AuditConfig{}, readRoot, fmt.Errorf("%s: %w", path, err)
	}
	return audit, readRoot, nil
}

// auditPlanResultMaxBytes bounds the result file: 256 KiB, about twelve times the
// largest full audit_multi result measured (design.md §D.5) and a few hundred
// times a digest, so no real result is refused and a runaway file is.
const auditPlanResultMaxBytes = 256 * 1024

// errAuditResultTooLarge reports an input longer than the cap.
var errAuditResultTooLarge = errors.New("larger than the size bound")

// readBoundedResult reads at most limit bytes from r. It asks the source for
// limit+1 bytes and no more, so an input longer than the cap is recognised after
// exactly one byte past it and the rest is never read.
func readBoundedResult(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errAuditResultTooLarge
	}
	return data, nil
}

// readAuditResultFile reads the one file the caller named, once. It stats the
// path first and refuses anything that is not a regular file — a directory, a
// named pipe, a device, a symlink resolving to one — before opening it, so a pipe
// cannot hang the verb; a symlink to a regular file is the file it names. The
// read itself is bounded (readBoundedResult).
func readAuditResultFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("--result-file %s: does not exist", path)
		}
		return nil, fmt.Errorf("--result-file %s: cannot be read: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("--result-file %s: not a regular file (mode %s)", path, info.Mode())
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("--result-file %s: cannot be read: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	data, err := readBoundedResult(f, auditPlanResultMaxBytes)
	if errors.Is(err, errAuditResultTooLarge) {
		return nil, fmt.Errorf("--result-file %s: larger than %d bytes (256 KiB)", path, auditPlanResultMaxBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("--result-file %s: cannot be read: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("--result-file %s: is empty", path)
	}
	return data, nil
}

// auditEntryPredicate decides whether one per_backend_verdicts entry answers a
// required gate. It is a parameter of checkAuditResult only so a test can run
// mutated variants over the same fixtures; production passes auditEntryAnswered.
type auditEntryPredicate func(verdict, gate any) bool

// auditEntryAnswered is the predicate of REQ-ACV-016: the verdict is a JSON
// string exactly "pass" or "fail" and the effective gate is the JSON string
// "required". Nothing is normalised — no case folding, no trimming — and a
// missing, null, empty, wrong-typed or any other value, inconclusive included,
// does not answer the gate.
func auditEntryAnswered(verdict, gate any) bool {
	v, ok := verdict.(string)
	if !ok || (v != auditVerdictPass && v != auditVerdictFail) {
		return false
	}
	g, ok := gate.(string)
	return ok && g == config.AuditGateRequired
}

// checkAuditResult compares plan with the audit_multi result JSON it is handed.
// An input that is not a JSON object is an error (never a pass). Otherwise the
// check is ok only when, for every enforced-required backend, each of the
// result's per_backend_verdicts entries for that backend satisfies answered and
// at least one exists, and — when any plan gate comes from the configuration —
// the result's plan_source is "config", the signal that the server that produced
// the result read the configuration at all. It performs no I/O.
func checkAuditResult(plan config.AuditPlan, resultJSON string, answered auditEntryPredicate) (auditPlanConvergenceCheck, error) {
	var top map[string]any
	if err := json.Unmarshal([]byte(resultJSON), &top); err != nil {
		return auditPlanConvergenceCheck{}, fmt.Errorf("not a JSON object: %w", err)
	}
	if top == nil {
		return auditPlanConvergenceCheck{}, fmt.Errorf("not a JSON object (got null)")
	}

	byBackend := map[string][]map[string]any{}
	if list, ok := top["per_backend_verdicts"].([]any); ok {
		for _, raw := range list {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if backend, ok := entry["backend"].(string); ok {
				byBackend[backend] = append(byBackend[backend], entry)
			}
		}
	}

	check := auditPlanConvergenceCheck{OK: true, Unmet: []string{}}
	var reasons []string
	for _, e := range plan.Backends {
		if !e.Explicit || e.Gate != config.AuditGateRequired {
			continue
		}
		entries := byBackend[e.Backend]
		if len(entries) == 0 {
			check.Unmet = append(check.Unmet, e.Backend)
			reasons = append(reasons, e.Backend+": no per_backend_verdicts entry")
			continue
		}
		for _, entry := range entries {
			if !answered(entry["verdict"], entry["gate"]) {
				check.Unmet = append(check.Unmet, e.Backend)
				reasons = append(reasons, fmt.Sprintf("%s: entry does not answer a required gate (verdict %s, gate %s)",
					e.Backend, auditJSONText(entry["verdict"]), auditJSONText(entry["gate"])))
				break
			}
		}
	}
	if plan.FromConfig() {
		if source, _ := top["plan_source"].(string); source != planSourceConfig {
			reasons = append(reasons, fmt.Sprintf("plan_source: %s, want %q (the server that produced the result did not read the configuration)",
				auditJSONText(top["plan_source"]), planSourceConfig))
		}
	}
	if len(reasons) > 0 {
		check.OK = false
		check.Reason = strings.Join(reasons, "; ")
	}
	return check, nil
}

// auditJSONText renders a decoded JSON value for a reason line; an absent member
// renders as "absent".
func auditJSONText(v any) string {
	if v == nil {
		return "absent or null"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
