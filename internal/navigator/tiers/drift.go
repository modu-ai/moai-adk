package tiers

import (
	"os"
	"os/exec"
	"strings"

	"log/slog"
)

// tierContractCIEnv is the env-var that promotes contract drift from advisory
// to CI-failing (REQ-NS3-002). When set to "1", the CLI layer (M4.6) SHOULD
// exit non-zero on any drifted contract. The tiers.json emission is ALWAYS
// fail-open — drift findings never block the graph (REQ-NS3-002).
const tierContractCIEnv = "TIER_CONTRACT_CI"

// driftCIMode reports whether CI mode is active. When false, drift findings
// are advisory only (one diagnostic line to .moai/logs/navigator-sync.log).
// When true, drift findings additionally cause the calling CLI to exit
// non-zero — that exit decision is owned by the CLI layer (M4.6), NOT by the
// emission path.
//
// @MX:NOTE: [AUTO] env-read single point for drift-CI flag — kept isolated so the no-wall-clock invariant is mechanically auditable
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func driftCIMode() bool {
	return strings.TrimSpace(os.Getenv(tierContractCIEnv)) == "1"
}

// runDriftChecks runs each contract's validator_command (if non-empty) and
// mutates DriftStatus in place. Returns anyDrifted — true iff at least one
// contract's validator reported drift. Emission is unaffected (fail-open).
//
// Fail-open contract: a validator command that cannot run (missing binary,
// non-zero exit, timeout) collapses to DriftCollapsed (design.md §8).
func runDriftChecks(projectRoot string, nodes []ContractNode) bool {
	any := false
	for i := range nodes {
		nodes[i].DriftStatus = checkContractDrift(projectRoot, nodes[i])
		if nodes[i].DriftStatus == DriftCollapsed {
			any = true
		}
	}
	return any
}

// checkContractDrift runs the validator for one contract and returns the
// resulting DriftStatus. Empty validator → unknown (no signal). Exit 0 →
// aligned. Non-zero exit OR exec error → drifted (fail-open).
func checkContractDrift(projectRoot string, c ContractNode) DriftStatus {
	if strings.TrimSpace(c.ValidatorCommand) == "" {
		return DriftUnknown
	}
	cmd := buildValidatorCommand(projectRoot, c.ValidatorCommand)
	if cmd == nil {
		return DriftCollapsed
	}
	if err := cmd.Run(); err != nil {
		slog.Debug("tiers: contract drift detected",
			"identifier", c.Identifier, "validator", c.ValidatorCommand, "error", err)
		return DriftCollapsed
	}
	return DriftAligned
}

// buildValidatorCommand constructs an *exec.Cmd for the validator. The
// validator is the user-declared validator_command from
// .moai/project/blueprint/contracts.yaml (trusted project config, NOT
// untrusted input); `sh -c` permits shell features in the validator
// (pipelines, env expansion, redirection), consistent with Makefile-target
// trust. No untrusted input reaches this exec path — the validator string
// is authored by the project maintainer alongside contracts.yaml and is
// read verbatim from that file. The cwd is set to projectRoot so relative
// paths in the validator resolve.
func buildValidatorCommand(projectRoot, validator string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", validator)
	if projectRoot != "" {
		cmd.Dir = projectRoot
	}
	// Suppress validator stdout/stderr — drift findings log at debug, not the
	// user-facing surface. The validator's own output is its own concern.
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}
