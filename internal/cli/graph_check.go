package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/graph"
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

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Report per-layer staleness (codemaps / mx-index / edges) numerically",
		Long: `Check the freshness of the three gated graph layers and exit accordingly.

Per layer the report names the layer, the metric kind used, the measured
integer value, the configured threshold, and a verdict (fresh | stale |
absent). An artifact without a provenance block is freshness-unjudgeable and
reported absent — never silently fresh. Absent is a distinct, FAILING verdict
for untracked layers (fresh-worktree state); the bootstrap a CI job performs
(moai mx scan + moai graph build) refreshes those layers to head first.

Metrics (per layer, by tracking status):
  codemaps  described-source-diff        files whose content differs from the
                                         stamped generation commit (endpoint
                                         diff; reverted churn counts zero)
  mx-index  inventory-content-diff       scanner-read files whose content hash
                                         differs from the stamped inventory
  edges     source-fingerprint-mismatch  source sets whose fingerprint moved
                                         since the stamped build

No filesystem mtime is read anywhere — a fresh worktree checkout resets every
mtime, which an mtime metric would misread as freshly regenerated.

Exit codes: 0 all fresh · 1 stale or absent · 2 system error.

Thresholds are configured in gate.yaml (graph_freshness section).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveGraphRoot(rootArg)
			if err != nil {
				return err
			}

			th, cfgErr := graphCheckThresholds(projectRoot)
			if cfgErr != nil {
				// A PRESENT-but-malformed gate.yaml is a system error (exit 2,
				// CR round-2 3855149230): silently checking under defaults
				// could pass a tree the operator's (unreadable) config meant
				// to fail. Absent or partial config keeps defaults.
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph check: config load failed: %v\n", cfgErr)
				return &exitCodeError{code: 2, msg: fmt.Sprintf("graph check: gate.yaml unreadable: %v", cfgErr)}
			}

			res, err := graph.CheckFreshness(projectRoot, th)
			if err != nil {
				// System error — exit 2 per the 0/1/2 contract. The message
				// prints explicitly: the ExitCoder vehicle's styled box is
				// suppressed by the fang error handler, so without this the
				// process would exit 2 silently.
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph check: system error: %v\n", err)
				return &exitCodeError{code: 2, msg: fmt.Sprintf("graph check: %v", err)}
			}

			if asJSON {
				data, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					return &exitCodeError{code: 2, msg: fmt.Sprintf("graph check: marshal: %v", err)}
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
			for _, l := range res.OffendingLayers() {
				_, _ = fmt.Fprintf(errs, "graph check: layer %s verdict=%s value=%d threshold=%d — %s\n",
					l.Layer, l.Verdict, l.Value, l.Threshold, l.Reason)
			}
			// REQ-GF-004: stale or absent exits 1, naming the offending layer.
			return &exitCodeError{code: 1, msg: "graph freshness check failed (stale or absent layer)"}
		},
	}

	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON report on stdout")

	return cmd
}

// graphCheckThresholds resolves thresholds from gate.yaml graph_freshness.
// Absent or partial config keeps the reasoned defaults; a PRESENT but
// malformed gate.yaml is an error the caller surfaces as exit 2 — a broken
// config must not silently pass a tree its intended thresholds would fail.
func graphCheckThresholds(projectRoot string) (graph.Thresholds, error) {
	th := graph.DefaultThresholds()
	moaiDir := filepath.Join(projectRoot, ".moai")

	// Pre-validate a present gate.yaml (CR round-2 3855149230): the shared
	// loader logs-and-continues on parse errors (defaults contract), so the
	// checker validates the file itself — a malformed file must exit 2, not
	// silently check under defaults.
	gatePath := filepath.Join(moaiDir, "config", "sections", "gate.yaml")
	if data, err := os.ReadFile(gatePath); err == nil {
		var probe map[string]any
		if yamlErr := yaml.Unmarshal(data, &probe); yamlErr != nil {
			return th, fmt.Errorf("gate.yaml present but unparseable: %w", yamlErr)
		}
	}

	loader := config.NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		return th, nil // absent/partial config: documented defaults path
	}
	gf := cfg.Gate.GraphFreshness
	if gf.CodemapsChangedFiles > 0 {
		th.CodemapsChangedFiles = gf.CodemapsChangedFiles
	}
	if gf.MXIndexChangedFiles > 0 {
		th.MXIndexChangedFiles = gf.MXIndexChangedFiles
	}
	return th, nil
}
