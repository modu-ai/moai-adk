package cli

// plugin_probe.go — the post-install list-surface probe
// (SPEC-INIT-SHRINK-001 design §2.4): the tri-state install-outcome signal
// this card adds BESIDE the t1435 install step. The step is fail-open by
// contract and returns nil in every case, so its caller cannot tell an
// install from a skip; the probe reads the OBSERVABLE list surface instead:
//
// opted-out        the t1435 opt-out is set (--no-plugin or
//                  MOAI_SKIP_PLUGIN_INSTALL=1|true). Decided before the
//                  step runs; no snapshot is taken.
// confirmed        for every tool the step acted on, the plugin ref
//                  (moai@moai-adk) is present in the post-execution
//                  surface AND absent from the pre-execution snapshot —
//                  the only diff that demonstrates THIS run's install.
// not-demonstrated every other outcome, decided only by the observable
//                  diff (present in both reads, absent from both, an
//                  unreadable surface, a timed-out read, any probe error).
//                  Conservative by construction; never names a cause
//                  inside the step.
//
// The list surfaces are read through the same REQ-017 runner seam the step
// and the doctor probe use, bounded by the doctor probe's timeout class:
// `claude plugin list` for the Claude tool, `codex plugin list --json` for
// the Codex tool (the doctor's read pattern). Under a test binary the
// default runner refuses, so every probe read degrades to
// not-demonstrated — inert by construction, exactly like the step.

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// installProbeOutcome is the tri-state verdict of design §2.4.
type installProbeOutcome string

const (
	probeOutcomeOptedOut        installProbeOutcome = "opted-out"
	probeOutcomeConfirmed       installProbeOutcome = "confirmed"
	probeOutcomeNotDemonstrated installProbeOutcome = "not-demonstrated"
)

// pluginSurfaceState is one tool's list-surface read for the plugin ref.
type pluginSurfaceState int

const (
	surfaceAbsent     pluginSurfaceState = iota // readable, ref not listed
	surfacePresent                              // readable, ref listed
	surfaceUnreadable                           // the read failed, timed out, or parsed wrong
)

// pluginInstallOptedOut reports whether the step will skip by the opt-out
// (the same rule runPluginInstallStep applies).
func pluginInstallOptedOut(opts pluginInstallOptions) bool {
	return opts.NoPlugin || pluginOptOutFromEnv()
}

// probeTimeout is the per-read bound: the doctor probe's timeout-constant
// class.
func probeTimeout() time.Duration { return config.DefaultPluginVersionProbeTimeout }

// readPluginListSurface reads one tool's installed-plugin list surface and
// reports whether the moai plugin ref is listed. Claude runs `plugin list`
// through the runner and matches the ref string in the combined output
// (design §2.4: the probe matches only the ref string — no coupling to the
// tool's internal install layout); Codex runs `plugin list --json` and
// matches the JSON entry, the doctor probe's read pattern. Any failure
// resolves to surfaceUnreadable.
func readPluginListSurface(tool pluginTool, projectRoot string, run pluginCommandRunner) pluginSurfaceState {
	bin, err := resolvePluginBinary(tool, projectRoot)
	if err != nil {
		return surfaceUnreadable
	}
	args := []string{"plugin", "list"}
	if tool == pluginToolCodex {
		args = append(args, "--json")
	}
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout())
	defer cancel()
	env := os.Environ()
	if tool == pluginToolCodex {
		if home, _ := resolveCodexHomeDir(); home != "" {
			env = envWithCodexHome(env, home)
		}
	}
	out, runErr := run.Run(ctx, bin, args, env)
	if runErr != nil {
		return surfaceUnreadable
	}
	if tool == pluginToolCodex {
		var list struct {
			Installed []struct {
				PluginID string `json:"pluginId"`
			} `json:"installed"`
		}
		if json.Unmarshal(out, &list) != nil {
			return surfaceUnreadable
		}
		for _, p := range list.Installed {
			if p.PluginID == pluginRef {
				return surfacePresent
			}
		}
		return surfaceAbsent
	}
	if strings.Contains(string(out), pluginRef) {
		return surfacePresent
	}
	return surfaceAbsent
}

// snapshotPluginListSurfaces captures the pre-execution snapshot: per acted
// tool, whether the surface lists the plugin ref.
func snapshotPluginListSurfaces(tools []pluginTool, projectRoot string, run pluginCommandRunner) map[pluginTool]pluginSurfaceState {
	snapshot := make(map[pluginTool]pluginSurfaceState, len(tools))
	for _, tool := range tools {
		snapshot[tool] = readPluginListSurface(tool, projectRoot, run)
	}
	return snapshot
}

// diffPluginListSurfaces reads the post-execution state and returns the
// verdict (design §2.4): confirmed only when EVERY acted tool's ref went
// absent-in-pre → present-in-post; every other combination — including any
// unreadable read — is not-demonstrated.
func diffPluginListSurfaces(pre map[pluginTool]pluginSurfaceState, tools []pluginTool, projectRoot string, run pluginCommandRunner) installProbeOutcome {
	for _, tool := range tools {
		// A snapshot always fills every acted tool; a missing pre-entry
		// (zero value surfaceAbsent) treats the ref as absent before — the
		// conservative reading of an incomplete snapshot.
		before := pre[tool]
		after := readPluginListSurface(tool, projectRoot, run)
		if before != surfaceAbsent || after != surfacePresent {
			return probeOutcomeNotDemonstrated
		}
	}
	return probeOutcomeConfirmed
}

// runInitPluginInstallProbed is runInit's install step call: it runs the
// t1435 step exactly as before (fail-open, opt-out honored, guidance to
// stderr) and returns the probe's verdict on the observable outcome. The
// step's own behavior is unchanged — the probe never re-runs or repairs it.
func runInitPluginInstallProbed(out io.Writer, wiring agentWiring, projectRoot string, noPlugin bool) installProbeOutcome {
	opts := newPluginInstallOptions(pluginToolsForHarness(wiring), projectRoot, noPlugin)
	if pluginInstallOptedOut(opts) {
		_ = runPluginInstallStep(out, opts) // no-ops by the step's own contract
		return probeOutcomeOptedOut
	}
	pre := snapshotPluginListSurfaces(opts.Tools, opts.ProjectRoot, pluginRunner)
	_ = runPluginInstallStep(out, opts)
	return diffPluginListSurfaces(pre, opts.Tools, opts.ProjectRoot, pluginRunner)
}
