package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/graph"
)

// graph check exit codes (the 0/1/2 contract as named constants — CR
// round-3: bare literals at four sites drifted from their documented
// meaning). The fresh path returns nil (exit 0) per cobra convention.
const (
	// exitStaleOrAbsent: any layer stale or absent (REQ-GF-004).
	exitStaleOrAbsent = 1
	// exitSystemError: system error — not-comparable stamp, unreadable
	// config, marshal failure (spec-gap exit-2 completion).
	exitSystemError = 2
)

// newGraphCheckCmd 'moai graph check' answers per-layer staleness numerically
// with exit-code discipline: 0 all fresh, 1 any layer stale/absent, 2 system
// error (REQ-GF-001/004). Thresholds come from gate.yaml graph_freshness
// (calibratable) with reasoned defaults.
//
// @MX:ANCHOR: [AUTO] moai graph check — exit-code contract consumed by moai gate and CI graph-freshness job
// @MX:REASON: CI and gate consume only the exit code; a reporting-only implementation silently disarms both consumers (the 740-commit silent-drift class)
func newGraphCheckCmd() *cobra.Command {
	var rootArg string
	var asJSON bool
	var eventArg string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Report per-layer staleness (codemaps / mx-index / edges / citations) numerically",
		Long: `Check the freshness of the gated graph layers and exit accordingly.

Per layer the report names the layer, the metric kind used, the measured
integer value, the configured threshold, and a verdict (fresh | stale |
absent). An artifact without a provenance block is freshness-unjudgeable and
reported absent — never silently fresh. Absent is a distinct, FAILING verdict
for untracked layers (fresh-worktree state); the bootstrap a CI job performs
(moai mx scan + moai graph build) refreshes those layers to head first.

Metrics (per layer, by tracking status):
  codemaps   described-source-diff         files whose content differs from
                                          the content anchor — the point the
                                          codemaps body last actually changed,
                                          not the stamped commit (endpoint
                                          diff; reverted churn counts zero).
                                          A re-stamp over an untouched body
                                          therefore does not reset the window.
  mx-index   inventory-content-diff        scanner-read files whose content
                                          hash differs from the stamped
                                          inventory
  edges      source-fingerprint-mismatch   source sets whose fingerprint
                                          moved since the stamped build
  citations  positive-cited-path-absence   source paths positively cited in
                                          .moai/project/codemaps/*.md that do
                                          not exist in the working tree
                                          (accuracy, not freshness — paths on
                                          blockquote lines are negative
                                          citations and exempt)

No filesystem mtime is read anywhere — a fresh worktree checkout resets every
mtime, which an mtime metric would misread as freshly regenerated.

Exit codes: 0 all fresh · 1 stale or absent · 2 system error.

One narrow exception turns a stale report into exit 0 plus a warning: on a
pull_request event (--event, else GITHUB_EVENT_NAME), when the codemaps layer
is the ONLY non-fresh layer, this change left the codemaps body and its
provenance stamp exactly as the first parent carried them, and the BASE alone
(value minus this change's measured contribution) is already at or over the
codemaps threshold. The JSON report states the origin as drift_origin:
inherited (base >= threshold), crossed-by-change (base < threshold <= value),
originated (contribution >= threshold), or anchor-moved (the change touched the
body or stamp, or that could not be measured, so value is not measured from the
base's window); only inherited downgrades, with verdict=stale kept and
verdict_class=warn added. crossed-by-change, originated, anchor-moved, a push
event (main, release/*), an absent layer, edges or citations failures, and an
unmeasured contribution all still exit 1. gate.graph_freshness.inherited_codemaps_drift
= block restores exit 1 for the inherited case too.

Thresholds are configured in gate.yaml (graph_freshness section).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveGraphRoot(rootArg)
			if err != nil {
				return err
			}

			th, driftPolicy, cfgErr := graphCheckThresholds(projectRoot)
			if cfgErr != nil {
				// A PRESENT-but-malformed gate.yaml is a system error (exit 2,
				// CR round-2 3855149230): silently checking under defaults
				// could pass a tree the operator's (unreadable) config meant
				// to fail. Absent or partial config keeps defaults.
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph check: config load failed: %v\n", cfgErr)
				return &exitCodeError{code: exitSystemError, msg: fmt.Sprintf("graph check: gate.yaml unreadable: %v", cfgErr)}
			}

			res, err := graph.CheckFreshness(projectRoot, th)
			if err != nil {
				// System error — exit 2 per the 0/1/2 contract. The message
				// prints explicitly: the ExitCoder vehicle's styled box is
				// suppressed by the fang error handler, so without this the
				// process would exit 2 silently.
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph check: system error: %v\n", err)
				writeUnreachableStampRecovery(cmd.ErrOrStderr(), err)
				return &exitCodeError{code: exitSystemError, msg: fmt.Sprintf("graph check: %v", err)}
			}

			// Classify before rendering so the JSON report carries verdict_class.
			// The verdict itself is never rewritten (card t2001, R208).
			res.ClassifyVerdicts(graphCheckEvent(eventArg), driftPolicy, th)

			if asJSON {
				data, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					return &exitCodeError{code: exitSystemError, msg: fmt.Sprintf("graph check: marshal: %v", err)}
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				out := cmd.OutOrStdout()
				for _, l := range res.Layers {
					reason := ""
					if l.Reason != "" {
						reason = "  (" + l.Reason + ")"
					}
					_, _ = fmt.Fprintf(out, "%-9s metric=%s value=%d threshold=%d verdict=%s%s\n",
						l.Layer, l.Metric, l.Value, l.Threshold, l.Verdict, reason)
				}
			}

			if !res.Failed() {
				return nil
			}
			errs := cmd.ErrOrStderr()
			if !res.Blocking() {
				// Every non-fresh layer was downgraded to a warning: announce it
				// and exit 0. The verdict above stays stale.
				writeInheritedDriftWarnings(cmd.OutOrStdout(), errs, asJSON, res.Warnings())
				for _, l := range res.Warnings() {
					writeLayerAttribution(errs, l)
				}
				return nil
			}
			for _, l := range res.OffendingLayers() {
				_, _ = fmt.Fprintf(errs, "graph check: layer %s verdict=%s value=%d threshold=%d — %s\n",
					l.Layer, l.Verdict, l.Value, l.Threshold, l.Reason)
				writeLayerAttribution(errs, l)
			}
			// REQ-GF-004: stale or absent exits 1, naming the offending layer.
			return &exitCodeError{code: exitStaleOrAbsent, msg: "graph freshness check failed (stale or absent layer)"}
		},
	}

	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON report on stdout")
	cmd.Flags().StringVar(&eventArg, "event", "", "CI event being judged (pull_request | push); defaults to GITHUB_EVENT_NAME. Only a pull_request may downgrade an inherited codemaps drift to a warning; any other event, or none, blocks")

	return cmd
}

// writeUnreachableStampRecovery names the recovery for the one system error
// whose fix is a WORKFLOW step rather than a git one: the codemaps stamp
// resolves as an object but is not an ancestor of HEAD, so the two trees share
// no comparison window and freshness was never measured.
//
// The block is specific to that state on purpose. A stamp git cannot resolve at
// all needs history (a deeper fetch), not a regenerated body, and printing the
// regenerate-and-restamp advice there would send the reader down a path that
// cannot work. Both halves of the recovery are stated because the half that is
// easy to reach for — re-stamping alone — is the one the freshness contract
// exists to defeat: a stamp rewritten over an untouched body would turn a stale
// tree green without regenerating anything.
func writeUnreachableStampRecovery(errs io.Writer, err error) {
	if !errors.Is(err, graph.ErrStampUnreachable) {
		return
	}
	_, _ = fmt.Fprint(errs,
		"graph check: unreachable stamp — the codemaps provenance names a commit that exists here\n"+
			"  but is not an ancestor of this checkout's HEAD, so freshness unmeasured (no value is reported).\n"+
			"  recovery: regenerate the codemaps body (/moai codemaps), then stamp it at a commit reachable\n"+
			"  from HEAD (moai graph stamp codemaps). A bare re-stamp over an unchanged body is NOT a fix.\n")
}

// writeLayerAttribution renders a failing layer's attribution beneath its
// verdict line: the change's own described-worthy contribution (REQ-GFC-007)
// and the paths driving the cumulative count (REQ-GFC-008).
//
// This is the surface that matters operationally. CI runs `moai graph check`
// and the lane reads its stderr, so a report that carries the attribution in
// the struct but renders only the count leaves the reader exactly where this
// SPEC found them: unable to tell an inherited red from one they caused.
//
// The contribution is REPORTED here. The exit code is decided by the cumulative
// count, except for the one narrow downgrade of graph.ClassifyVerdicts (card
// t2001, user request R208): a pull_request whose only non-fresh layer is a
// codemaps drift it did not originate is announced, not failed.
func writeLayerAttribution(errs io.Writer, l graph.LayerReport) {
	// The measurement window comes first: without it a reader cannot tell
	// which two points the count spans, and the codemaps layer no longer
	// measures from the stamped commit (SPEC-GRAPH-GATE-RESTAMP-001).
	// Reported only — the exit code is still decided by the count alone.
	if l.ContentAnchor != "" {
		anchor := l.ContentAnchor
		if len(anchor) > 9 {
			anchor = anchor[:9]
		}
		_, _ = fmt.Fprintf(errs, "  measured from: %s (%s)\n", anchor, l.ContentAnchorSource)
	}
	switch {
	case l.Contribution != nil:
		base := l.ContributionBase
		if len(base) > 9 {
			base = base[:9]
		}
		_, _ = fmt.Fprintf(errs, "  contribution: %d described-worthy file(s) vs first parent %s (%s)\n",
			*l.Contribution, base, driftOriginSentence(l))
	case l.ContributionAbsentReason != "":
		// Absent, and said so. A fabricated 0 would read as the inheriting
		// signature on every checkout that has no comparison base.
		_, _ = fmt.Fprintf(errs, "  contribution: unmeasured — %s\n", l.ContributionAbsentReason)
	}
	for _, path := range l.DrivingPaths {
		_, _ = fmt.Fprintf(errs, "    %s\n", path)
	}
	if l.DrivingPathsOmitted > 0 {
		_, _ = fmt.Fprintf(errs, "    ... and %d more (listing bounded)\n", l.DrivingPathsOmitted)
	}
}

// graphCheckThresholds resolves thresholds from gate.yaml graph_freshness.
// Absent or partial config keeps the reasoned defaults; a PRESENT but
// malformed gate.yaml is an error the caller surfaces as exit 2 — a broken
// config must not silently pass a tree its intended thresholds would fail.
func graphCheckThresholds(projectRoot string) (graph.Thresholds, string, error) {
	th := graph.DefaultThresholds()
	policy := config.DefaultGraphFreshnessInheritedCodemapsDrift
	moaiDir := filepath.Join(projectRoot, ".moai")

	// Pre-validate a present gate.yaml (CR round-2 3855149230): the shared
	// loader logs-and-continues on parse errors (defaults contract), so the
	// checker validates the file itself — a malformed file must exit 2, not
	// silently check under defaults.
	gatePath := filepath.Join(moaiDir, "config", "sections", "gate.yaml")
	if data, err := os.ReadFile(gatePath); err == nil {
		var probe map[string]any
		if yamlErr := yaml.Unmarshal(data, &probe); yamlErr != nil {
			return th, policy, fmt.Errorf("gate.yaml present but unparseable: %w", yamlErr)
		}
	}

	loader := config.NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		return th, policy, nil // absent/partial config: documented defaults path
	}
	gf := cfg.Gate.GraphFreshness
	if gf.CodemapsChangedFiles > 0 {
		th.CodemapsChangedFiles = gf.CodemapsChangedFiles
	}
	if gf.MXIndexChangedFiles > 0 {
		th.MXIndexChangedFiles = gf.MXIndexChangedFiles
	}
	// The inherited-drift policy: empty keeps the default; a value other than
	// warn|block is a malformed config and, like a malformed file, a system
	// error — a typo must not silently pick a side of a gate.
	switch gf.InheritedCodemapsDrift {
	case "":
	case graph.InheritedDriftWarn, graph.InheritedDriftBlock:
		policy = gf.InheritedCodemapsDrift
	default:
		return th, policy, fmt.Errorf("gate.graph_freshness.inherited_codemaps_drift %q must be %q or %q",
			gf.InheritedCodemapsDrift, graph.InheritedDriftWarn, graph.InheritedDriftBlock)
	}
	return th, policy, nil
}

// graphCheckEvent resolves the CI event the check is judging: the --event flag
// when given, else GITHUB_EVENT_NAME. An unknown or empty event is returned
// as-is and the policy treats it as blocking.
func graphCheckEvent(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	return os.Getenv("GITHUB_EVENT_NAME")
}

// driftOriginSentence states, as a fact, where the codemaps count came from.
// The base's figure is LayerReport.BaseCount, measured directly; it is never
// derived as value - contribution, because Value is the union of the base's and
// the change's file sets and a subtraction undercounts when they overlap. When
// BaseCount is unmeasured (nil) no base number is printed. The wording follows
// graph's three-way DriftOrigin and never calls a count "inherited" unless the
// base carried it.
func driftOriginSentence(l graph.LayerReport) string {
	if l.Contribution == nil {
		return "origin not stated"
	}
	switch l.DriftOrigin {
	case graph.DriftInherited:
		if l.BaseCount == nil {
			return fmt.Sprintf("inherited — threshold %d, this change added %d (the base's own count was not measured)", l.Threshold, *l.Contribution)
		}
		return fmt.Sprintf("inherited — the base alone carries %d >= threshold %d, this change added %d", *l.BaseCount, l.Threshold, *l.Contribution)
	case graph.DriftCrossedByChange:
		if l.BaseCount == nil {
			return fmt.Sprintf("crossed by this change — threshold %d, this change added %d and took the count over the line (the base's own count was not measured)", l.Threshold, *l.Contribution)
		}
		return fmt.Sprintf("crossed by this change — the base carried %d < threshold %d, this change added %d and took the count over the line", *l.BaseCount, l.Threshold, *l.Contribution)
	case graph.DriftOriginated:
		return fmt.Sprintf("originated by this change — it alone is %d >= threshold %d", *l.Contribution, l.Threshold)
	case graph.DriftAnchorMoved:
		return "origin not attributable — this change touched the codemaps body or stamp (or that could not be ruled out), so its measurement window is not the base's"
	default:
		return "origin not stated"
	}
}

// writeInheritedDriftWarnings announces the layers that were downgraded from
// block to warn. In text mode the GitHub annotation goes to stdout, where the
// runner reads workflow commands; in --json mode stdout must stay a single JSON
// document, so the annotation goes to stderr instead. The text never offers a
// bare re-stamp as a remedy: a stamp rewritten over an untouched body turns a
// stale tree green without regenerating anything.
func writeInheritedDriftWarnings(out, errs io.Writer, asJSON bool, warns []graph.LayerReport) {
	dst := out
	if asJSON {
		dst = errs
	}
	for _, l := range warns {
		contribution := "unmeasured"
		if l.Contribution != nil {
			contribution = fmt.Sprintf("%d", *l.Contribution)
		}
		_, _ = fmt.Fprintf(dst,
			"::warning title=graph-freshness::layer %s is %s (value=%d >= threshold=%d), inherited: the base alone carries the count (%s), this change contributed %s described-worthy file(s), and every other layer is fresh. Regenerate the %s body (/moai codemaps) and stamp it; a bare re-stamp is not a fix.\n",
			l.Layer, l.Verdict, l.Value, l.Threshold, driftOriginSentence(l), contribution, l.Layer)
	}
}
