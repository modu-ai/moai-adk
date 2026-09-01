// Package cli — doctor_agentemit_embed.go
//
// `moai doctor` check: do the .codex agent artifacts EMBEDDED IN A BUILT
// BINARY still match the committed emission set?
//
// The .codex/agents/moai/*.toml artifacts are machine-emitted from the .md
// source layer by internal/template/agentemit, and they reach users as
// //go:embed assets compiled into the binary. Nothing in `go test` can judge
// that axis: the test binary is recompiled on every run, so //go:embed reads
// the same committed bytes the test compares against — both sides move
// together and the comparison is a tautology. A stale embed exists only in an
// ALREADY-BUILT binary, which `go test` never sees.
//
// This check therefore judges a build artifact, not a source tree: it asks a
// built binary to deploy its own templates into a scratch directory outside
// the repository and compares the .codex TOMLs it produced against the
// committed ones. It never rebuilds the binary under test — rebuilding would
// restore the same tautology.
package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// agentEmitEmbedCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const agentEmitEmbedCheckName = "Agent Emit Embed"

// committedEmissionRelDir is the committed emission set, relative to the
// project root under check. Its presence IS the applicability predicate.
//
// A deployed project's own .codex/agents/moai/*.toml set is deliberately NOT
// substitutable here: that set is a deployment OUTPUT of this repository's
// templates, not a committed artifact of the tree under check. Keying the
// predicate on it would turn this into a different check.
const committedEmissionRelDir = "internal/template/templates/.codex/agents/moai"

// deployedEmissionRelDir is where a binary's own deployment path writes the
// emitted artifacts — the extraction surface this check reads.
const deployedEmissionRelDir = ".codex/agents/moai"

// embedCheckBinEnvKey overrides the judgment-target binary path, so the same
// judgment can be aimed at an installed binary rather than ./bin/moai. This
// is the seam `make embed-check BIN=<path>` drives.
const embedCheckBinEnvKey = "MOAI_EMBED_CHECK_BIN"

// defaultEmbedCheckBinRel is the judgment target when no override is given,
// resolved against the project root under check.
const defaultEmbedCheckBinRel = "bin/moai"

// projectRootMarker is the directory whose presence marks a project root.
const projectRootMarker = ".moai"

// emissionExtractor asks a binary to materialize its own embedded templates
// and returns the directory they were written to, plus a cleanup func.
type emissionExtractor func(binPath string) (dir string, cleanup func(), err error)

// checkAgentEmitEmbed is the registered doctor entry point.
func checkAgentEmitEmbed(cwd string, verbose bool) DiagnosticCheck {
	return checkAgentEmitEmbedAgainst(cwd, "", extractEmissionViaInit, verbose)
}

// checkAgentEmitEmbedAgainst is checkAgentEmitEmbed with the judgment target
// and the extraction path injected, so the comparison can be exercised
// without building or running a real binary.
//
// binPath == "" resolves the target from the env override, then from the
// project root's bin/moai.
//
// @MX:WARN: [AUTO] 10 if-branches, each an early return carrying a distinct
// verdict (not applicable / no judgment target / extraction failure /
// comparison failure / cardinality shortfall / stale bytes / match). A branch
// inserted at the wrong point silently reclassifies one of the others.
// @MX:REASON: [AUTO] the branch ORDER is the contract: applicability is
// decided before a judgment target is demanded, and the cardinality gate runs
// before the byte comparison. Reordering either makes "0 comparisons -> pass"
// reachable again — the vacuity this check exists to close.
// @MX:SPEC: SPEC-CI-DOCTOR-BIN-001 (no-judgment-target branch verdict;
// branch structure from SPEC-AGENT-EMIT-LINEAGE-001)
func checkAgentEmitEmbedAgainst(cwd, binPath string, extract emissionExtractor, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: agentEmitEmbedCheckName}

	// --- applicability -------------------------------------------------
	//
	// doctor hands every check the raw os.Getwd() value and resolves no root
	// of its own, so locating the project root is this check's own job.
	root, ok := findEmbedCheckRoot(cwd)
	if !ok {
		check.Status = uikit.CheckOK
		if _, inProject := nearestProjectRoot(cwd); inProject {
			check.Message = fmt.Sprintf("not applicable: no committed emission set at %s/", committedEmissionRelDir)
		} else {
			check.Message = "not applicable: no committed agent-emit artifacts (not a MoAI project root)"
		}
		if verbose {
			check.Detail = "this check judges the moai-adk repository's own emitted .codex artifacts; " +
				"a deployed project carries none and is skipped"
		}
		return check
	}

	// Non-empty by construction: it is what the walk matched on.
	committed := committedEmissionSet(root)

	// --- applicable: no readable judgment target is an informational skip --
	//
	// A tree without bin/moai cannot host this judgment at all — the CI jobs
	// that run `go test` never build a binary, so reporting a failed judgment
	// there classifies "nothing to judge" as a judgment outcome. Precedent:
	// checkBinaryFreshness (t184) reports ok + a reason for every "cannot
	// judge" input instead of gating doctor on it. REQ-CDB-003 is untouched:
	// the fail paths below fire exactly when a readable binary IS present.
	if binPath == "" {
		binPath = resolveEmbedCheckBin(root)
	}
	if info, err := os.Stat(binPath); err != nil || info.IsDir() {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("skipped: no readable binary to judge at %s — %d committed artifacts unjudged; build one with `make build` or aim the check elsewhere with %s=<path>",
			binPath, len(committed), embedCheckBinEnvKey)
		if verbose {
			check.Detail = fmt.Sprintf("the judgment target would have been %s, compared against the committed %s/ set", binPath, committedEmissionRelDir)
		}
		return check
	}

	dir, cleanup, err := extract(binPath)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("could not extract embedded artifacts from %s: %v", binPath, err)
		return check
	}

	compared, differing, uncompared, err := compareEmission(committed, dir)
	if err != nil {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("comparison failed: %v", err)
		return check
	}

	// --- cardinality gate ------------------------------------------------
	//
	// A partially-successful extraction must not pass by comparing a subset.
	if compared < len(committed) {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("compared %d/%d artifacts — %s carries no embedded counterpart for: %s",
			compared, len(committed), filepath.Base(binPath), strings.Join(uncompared, ", "))
		return check
	}

	if len(differing) > 0 {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf("%s embeds stale agent-emit artifacts (%d/%d compared): %s",
			filepath.Base(binPath), compared, len(committed), strings.Join(differing, ", "))
		check.Detail = fmt.Sprintf("the binary at %s was built before the current %s/ — rebuild with `make build`. "+
			"If the committed artifacts are themselves stale, regenerate with `make agents-emit`.",
			binPath, committedEmissionRelDir)
		return check
	}

	check.Status = uikit.CheckOK
	check.Message = fmt.Sprintf("%d/%d embedded agent-emit artifacts match the committed set (%s)",
		compared, len(committed), filepath.Base(binPath))
	if verbose {
		check.Detail = fmt.Sprintf("judged %s against %s/ without rebuilding it", binPath, committedEmissionRelDir)
	}
	return check
}

// findEmbedCheckRoot walks up from start to the nearest ancestor carrying the
// COMMITTED EMISSION SET. doctor passes each check the raw working directory,
// so a run from a subdirectory must still resolve the same root — otherwise an
// applicable tree silently reports "not applicable".
//
// The walk does NOT stop at the nearest .moai/-bearing ancestor. Measured in
// this repository: a package directory can carry a bare .moai/state/ left by a
// test side effect, and anchoring there misjudges the applicable repository
// root as skippable. The applicability predicate names the committed set, so
// the committed set is what the walk looks for.
func findEmbedCheckRoot(start string) (string, bool) {
	return walkUp(start, func(dir string) bool {
		return len(committedEmissionSet(dir)) > 0
	})
}

// nearestProjectRoot walks up to the nearest ancestor carrying a .moai/
// directory. It phrases the not-applicable reason and decides nothing.
func nearestProjectRoot(start string) (string, bool) {
	return walkUp(start, func(dir string) bool {
		info, err := os.Stat(filepath.Join(dir, projectRootMarker))
		return err == nil && info.IsDir()
	})
}

// walkUp returns the nearest ancestor of start (start included) satisfying
// match, walking toward the filesystem root.
func walkUp(start string, match func(dir string) bool) (string, bool) {
	dir := start
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", false
		}
		dir = wd
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	for {
		if match(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// committedEmissionSet lists the committed .toml artifacts under root, sorted.
func committedEmissionSet(root string) []string {
	matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(committedEmissionRelDir), "*.toml"))
	if err != nil {
		return nil
	}
	sort.Strings(matches)
	return matches
}

// resolveEmbedCheckBin picks the judgment target: the env override when set,
// otherwise the project root's bin/moai.
func resolveEmbedCheckBin(root string) string {
	if v := strings.TrimSpace(os.Getenv(embedCheckBinEnvKey)); v != "" {
		return v
	}
	return filepath.Join(root, filepath.FromSlash(defaultEmbedCheckBinRel))
}

// compareEmission compares every committed artifact against its extracted
// counterpart. It returns how many paths it actually compared, which of those
// differed, and which committed artifacts had no counterpart at all.
func compareEmission(committed []string, extractedRoot string) (compared int, differing, uncompared []string, err error) {
	extractedDir := filepath.Join(extractedRoot, filepath.FromSlash(deployedEmissionRelDir))
	for _, c := range committed {
		base := filepath.Base(c)
		got, readErr := os.ReadFile(filepath.Join(extractedDir, base))
		if readErr != nil {
			uncompared = append(uncompared, base)
			continue
		}
		want, readErr := os.ReadFile(c)
		if readErr != nil {
			return compared, differing, uncompared, fmt.Errorf("read committed %s: %w", base, readErr)
		}
		compared++
		if !bytes.Equal(got, want) {
			differing = append(differing, base)
		}
	}
	sort.Strings(differing)
	sort.Strings(uncompared)
	return compared, differing, uncompared, nil
}

// extractEmissionViaInit asks the binary under test to deploy its own
// embedded templates into a scratch directory and returns that directory.
//
// The scratch lives under the OS temp root, never inside the repository:
// this deployment writes a whole project (git init and hook installation
// included), and a mis-aimed target directory would contaminate the tree the
// check is judging.
//
// @MX:WARN: [AUTO] this spawns the binary under judgment and lets it deploy
// an entire project (git init and hook installation included). The scratch
// target MUST stay outside the repository, and os.RemoveAll is aimed only at
// the os.MkdirTemp result. Tagged WARN rather than ANCHOR: the subprocess
// stays inside this process tree, so the "external system" clause does not
// apply — what this carries is a hazard marker, not an anchor.
// @MX:REASON: [AUTO] REQ-AEL-002 forbids every check this SPEC introduces
// from writing inside the repository tree; a mis-aimed target directory would
// contaminate the very tree being judged and then be deleted.
// @MX:SPEC: SPEC-AGENT-EMIT-LINEAGE-001
func extractEmissionViaInit(binPath string) (string, func(), error) {
	base, err := os.MkdirTemp("", "moai-embed-check-")
	if err != nil {
		return "", func() {}, fmt.Errorf("scratch dir: %w", err)
	}
	cleanup := func() {
		// Best-effort scratch removal: a failure here leaves a temp dir the
		// OS reclaims, and must not change the check's verdict.
		if rmErr := os.RemoveAll(base); rmErr != nil {
			fmt.Fprintf(os.Stderr, "doctor: could not remove embed-check scratch %s: %v\n", base, rmErr)
		}
	}

	// The command runs WITH ITS WORKING DIRECTORY SET TO THE SCRATCH, so a
	// relative target path (the `make embed-check` default `bin/moai`) would
	// otherwise be looked up inside the scratch and fail as "not found".
	// Resolve it against the caller's working directory first.
	execPath := binPath
	if !filepath.IsAbs(execPath) {
		if abs, absErr := filepath.Abs(execPath); absErr == nil {
			execPath = abs
		}
	}

	target := filepath.Join(base, "extract")
	cmd := exec.Command(execPath, "init", target, "--non-interactive")
	cmd.Dir = base
	// AGENTEMIT_UPDATE is the regeneration switch of the emitter's golden
	// path. It has no role here, but scrubbing it keeps this check's verdict
	// independent of the environment it happens to inherit.
	cmd.Env = append(os.Environ(), "AGENTEMIT_UPDATE=")
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return "", cleanup, fmt.Errorf("%s init: %w (%s)", filepath.Base(binPath), runErr, boundedTail(out))
	}
	return target, cleanup, nil
}

// boundedTail trims command output to a short tail for error messages.
func boundedTail(b []byte) string {
	s := strings.TrimSpace(string(b))
	const max = 300
	if len(s) > max {
		return "..." + s[len(s)-max:]
	}
	return s
}
