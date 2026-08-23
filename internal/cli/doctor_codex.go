package cli

// doctor_codex.go — SPEC-CODEX-WIRING-001 M4, the `moai doctor` "Codex
// Wiring" diagnostic (REQ-CW-010 / AC-CW-012).
//
// Advisory and fail-open (checkBinaryFreshness t184 precedent): the check
// reports, it never gates — an inactive project (no wiring files) is an
// informational skip, and every finding is an action directive rather than a
// claim about Codex's undocumented internal trust store (spec §H: the
// sidecar divergence is an INDIRECT signal; the advice says what to DO).

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

// reTrustAdvice is the divergence action directive (AC-CW-012 token).
const reTrustAdvice = "run codex /hooks to re-trust the changed hooks"

// checkCodexWiring verifies the Codex wiring of the project at root:
// hooks.json presence + whitelist validity, sidecar-hash divergence, the
// moai binary's PATH resolution, and the config.toml [mcp_servers.moai]
// table's presence/canonical shape. A project without wiring files is an
// informational skip (the update path's opt-in marker, REQ-CW-009).
func checkCodexWiring(root string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Codex Wiring"}

	hooksRaw, hooksErr := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
	cfgRaw, cfgErr := os.ReadFile(filepath.Join(root, codexwiring.ConfigRelPath))
	if hooksErr != nil && cfgErr != nil {
		check.Status = uikit.CheckOK
		check.Message = "not wired (claude-only project) — skipped"
		return check
	}

	var problems []string
	var advice []string

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
