package cli

// doctor_codex.go — SPEC-CODEX-WIRING-001 M4, the `moai doctor` "Codex
// Wiring" diagnostic (REQ-CW-010 / AC-CW-012).
//
// Advisory and fail-open (checkBinaryFreshness t184 precedent): the check
// reports, it never gates — an inactive project on a codex-less machine is an
// informational skip, and every finding is an action directive rather than a
// claim about Codex's undocumented internal trust store (spec §H: the
// sidecar divergence is an INDIRECT signal; the advice says what to DO).
//
// The check READS only. It never creates, repairs, or removes a .codex/ file,
// and never touches the user-layer ~/.codex/config.toml: wiring creation is
// the explicit `moai init --agent codex` opt-in (REQ-CW-009) and a user-owned
// table is the doctor's to report, never the writer's to repair (REQ-CW-005).

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// codexWiringLookPath is the PATH-resolution seam (stubbed in tests so the
// check's verdict never depends on the machine running the test suite).
var codexWiringLookPath = exec.LookPath

// codexWiringUserHomeDir is the home-directory seam. The stale-skill
// sub-check reads the USER-layer ~/.codex/config.toml, which lives outside
// the project root; the seam keeps tests off the developer's real home
// without an environment mutation (t.Setenv("HOME", …) pollutes parallel
// tests, CLAUDE.local.md §6).
var codexWiringUserHomeDir = os.UserHomeDir

// reTrustAdvice is the divergence action directive (AC-CW-012 token).
const reTrustAdvice = "run codex /hooks to re-trust the changed hooks"

// initCodexAdvice is the unwired-project action directive. `moai init` is the
// ONLY path that creates wiring — RefreshWiring is an existence gate that
// creates nothing (REQ-CW-009), so an unwired project stays unwired until the
// user opts in explicitly.
const initCodexAdvice = "run moai init --agent codex"

// codexHomeConfigDisplay is how the user-layer config is named in findings.
const codexHomeConfigDisplay = "~/.codex/config.toml"

// checkCodexWiring verifies the Codex wiring of the project at root:
// hooks.json presence + whitelist validity, sidecar-hash divergence, the
// moai binary's PATH resolution, and the config.toml [mcp_servers.moai]
// table's presence/canonical shape — plus, when Codex is in play at all, the
// user-layer skill registrations. A project without wiring files on a machine
// without codex is an informational skip (the opt-in marker, REQ-CW-009).
func checkCodexWiring(root string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Codex Wiring"}

	hooksRaw, hooksErr := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
	cfgRaw, cfgErr := os.ReadFile(filepath.Join(root, codexwiring.ConfigRelPath))

	wired := hooksErr == nil || cfgErr == nil
	_, codexPathErr := codexWiringLookPath("codex")
	codexInstalled := codexPathErr == nil

	if !wired && !codexInstalled {
		// Claude-only project on a claude-only machine: nothing about Codex
		// is in play, so the check stays silent. This is the un-nagging
		// invariant — it must survive every addition below.
		check.Status = uikit.CheckOK
		check.Message = "not wired (claude-only project) — skipped"
		return check
	}

	var problems []string
	var advice []string

	if !wired {
		// Codex IS installed but this project was never wired. Silence here
		// costs the user the whole integration with no signal: the MoAI MCP
		// server unregistered and every generated hook dead. The directive
		// rides IN the message because Detail renders only under --verbose.
		problems = append(problems, fmt.Sprintf(
			"codex is installed but this project is not wired (%s and %s both absent) — the MoAI MCP server is not registered and the generated hooks cannot fire here; %s",
			codexwiring.HooksRelPath, codexwiring.ConfigRelPath, initCodexAdvice))
	} else {
		// hooks.json: presence, whitelist validity, sidecar divergence.
		if hooksErr != nil {
			problems = append(problems, fmt.Sprintf("%s missing (wiring active but the hook layer is gone)", codexwiring.HooksRelPath))
		} else {
			if violations, verr := codexadapter.ValidateConfig(hooksRaw); verr != nil {
				problems = append(problems, fmt.Sprintf("hooks.json unparseable: %v", verr))
			} else if len(violations) > 0 {
				// Codex silently disables the whole file on one stray key (t83
				// Finding D) — this is the observability backstop.
				names := make([]string, len(violations))
				for i, v := range violations {
					names[i] = v.Error()
				}
				problems = append(problems, "hooks.json fails the key whitelist ("+strings.Join(names, "; ")+") — Codex would silently ignore the file")
			}

			if sidecar, present, serr := codexwiring.LoadSidecar(root); serr != nil {
				check.Detail = fmt.Sprintf("sidecar unreadable: %v", serr)
			} else if present {
				sum := sha256.Sum256(hooksRaw)
				if sidecar.HooksSHA256 != hex.EncodeToString(sum[:]) {
					// The directive rides IN the message (not Detail) so a plain
					// `moai doctor` surfaces it — Detail only renders under
					// --verbose (AC-CW-012 clause 2's plain-doctor reading).
					problems = append(problems, "hooks.json differs from the last generated content (sidecar hash mismatch) — "+reTrustAdvice)
					advice = append(advice, reTrustAdvice)
				}
			} else if verbose {
				check.Detail = "no trust sidecar recorded (never wired by this generator)"
			}
		}

		// moai binary PATH resolution: the generated hook commands are
		// `moai hook ...` — an unresolvable binary means none of them can fire.
		if _, lerr := codexWiringLookPath("moai"); lerr != nil {
			problems = append(problems, "moai binary not found on PATH — the generated hook commands cannot fire")
		}

		// config.toml: table presence + canonical shape (drift is REPORTED; the
		// writer never repairs a user-owned table, REQ-CW-005).
		if cfgErr != nil {
			problems = append(problems, fmt.Sprintf("%s missing (wiring active but the MCP registration is gone)", codexwiring.ConfigRelPath))
		} else {
			status := codexwiring.InspectMCPTable(cfgRaw)
			switch {
			case !status.Present:
				problems = append(problems, "[mcp_servers.moai] table missing from config.toml")
			case !status.Canonical:
				problems = append(problems, "[mcp_servers.moai] table differs from the canonical registration (user-owned; left untouched)")
			}
		}
	}

	// User-layer skill registrations. Reached only when Codex is in play at
	// all (wired project, or codex on PATH) — the guard above already
	// returned for the claude-only case.
	if finding := codexStaleSkillFinding(); finding != "" {
		problems = append(problems, finding)
	}

	if len(problems) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical)"
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = strings.Join(problems, "; ")
	if len(advice) > 0 {
		check.Detail = strings.Join(advice, "; ")
	} else if verbose && check.Detail == "" {
		check.Detail = "advisory check — rerun `moai init --agent codex` to refresh the wiring"
	}
	return check
}

// codexStaleSkillFinding reports the user-layer [[skills.config]] entries
// whose declared path no longer exists — registrations Codex neither prunes
// nor complains about, so nothing else surfaces them.
//
// Fail-open in every direction: an unresolvable home, an absent or unreadable
// config, or a config declaring no entries all yield "" (a silent skip), so a
// missing input never becomes a finding. The function READS only.
func codexStaleSkillFinding() string {
	home, err := codexWiringUserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return ""
	}
	entries := codexwiring.ParseSkillEntries(raw)
	if len(entries) == 0 {
		return ""
	}

	var missingEnabled, missingDisabled int
	for _, e := range entries {
		if e.Path == "" {
			// An entry declaring no path says nothing about a file's
			// existence — counted in the total, never as a missing path.
			continue
		}
		if _, serr := os.Stat(e.Path); serr != nil {
			if e.Enabled {
				missingEnabled++
			} else {
				missingDisabled++
			}
		}
	}
	missing := missingEnabled + missingDisabled
	if missing == 0 {
		return ""
	}

	// The enabled/disabled split is the severity axis, so it is quantified
	// rather than summed away: a disabled missing path is stale bookkeeping
	// that breaks nothing today, while an enabled one is a live registration
	// pointing at a file that is not there.
	return fmt.Sprintf(
		"%s declares %d of %d [[skills.config]] entries whose path no longer exists (%d enabled, %d disabled) — remove the stale entries or restore the skill files",
		codexHomeConfigDisplay, missing, len(entries), missingEnabled, missingDisabled)
}
