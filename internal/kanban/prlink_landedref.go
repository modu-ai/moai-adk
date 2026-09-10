// prlink_landedref.go — the landed ref's three-level resolution chain and its
// answering-level report (SPEC-TODO-LANDING-ATTRIBUTION-001 REQ-TLA-007..009,
// REQ-TLA-012, M2).
//
// Milestone 1 repaired the predicate; this milestone repairs WHAT IT ASKS.
// The landed question used to fall from an empty configuration straight to a
// compiled-in constant (MUT-CONFIG-ONLY) — so a repository whose own git
// metadata already records its integration branch, and whose primary checkout
// simply does not configure the key, asked every landing question about a
// branch its work never reaches, silently. The repository already knows the
// answer: `git symbolic-ref refs/remotes/origin/HEAD`.
//
// The chain is ordered, and the order is not interchangeable: a project that
// CONFIGURES the key keeps its current behaviour exactly (C-4), so level 2 is
// consulted only when level 1 yields nothing. Level 2 READS the symref and
// never writes a repository ref (REQ-TLA-012) — the write form of
// symbolic-ref, `remote set-head`, and `update-ref` are all absent from this
// path, and the subprocess census in the tests asserts that shape.
package kanban

import (
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// LandedRefLevel names WHICH level of the resolution chain answered. The
// caller discloses it so an operator sees where the ref came from rather than
// inferring it (REQ-TLA-011): a level below the configured key is the
// exceptional path, and a silent fallback is how MUT-SILENT-FALLBACK hid.
type LandedRefLevel int

const (
	// LandedRefConfigured — level 1: the project configured the key.
	LandedRefConfigured LandedRefLevel = 1
	// LandedRefOriginHEAD — level 2: refs/remotes/origin/HEAD answered.
	LandedRefOriginHEAD LandedRefLevel = 2
	// LandedRefDefault — level 3: the compiled-in default.
	LandedRefDefault LandedRefLevel = 3
)

// landedRefGitRun is the seam chain level 2 reads through. It is a variable
// so tests can record the invocations the resolution path makes — the census
// behind REQ-TLA-012's writes-nothing property runs through it.
var landedRefGitRun CommandRunner = func(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

// LandedRefForWithLevel resolves the landed ref through the ordered
// three-level chain and reports which level answered:
//
//  1. the configured git_strategy.worktree_base_branch (C-4: a project that
//     configures the key keeps its current behaviour exactly — level 2 is
//     never consulted),
//  2. refs/remotes/origin/HEAD — the integration branch the repository
//     itself records, read through a local ref (C-3: no network I/O),
//  3. DefaultLandedRef — on any level-2 failure (REQ-TLA-009).
//
// No failure path aborts the invocation: an unresolvable symref behaves as
// one that is absent.
func LandedRefForWithLevel(projectRoot string) (string, LandedRefLevel) {
	if base := strings.TrimSpace(config.LoadWorktreeBaseBranch(projectRoot)); base != "" {
		return "origin/" + base, LandedRefConfigured
	}
	// The READ form of symbolic-ref: exactly one operand after the subcommand,
	// the ref to read. The write form (a second operand) is never issued here
	// — REQ-TLA-012's census asserts exactly that shape on this path.
	if out, err := landedRefGitRun("git", "-C", projectRoot,
		"symbolic-ref", "refs/remotes/origin/HEAD"); err == nil {
		if sym := strings.TrimSpace(out); sym != "" {
			return strings.TrimPrefix(sym, "refs/remotes/"), LandedRefOriginHEAD
		}
	}
	return DefaultLandedRef, LandedRefDefault
}
