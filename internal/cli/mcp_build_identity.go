package cli

// SPEC-AUDIT-BUILD-IDENTITY-001 — the ONE build-identity assembly point shared
// by the three audit entry points (codex_audit, glm_audit, audit_multi;
// REQ-ABI-007). A verdict that does not name the binary that produced it
// cannot be re-attributed after the fact, which is how a 259-commit-lagging
// install kept serving audit verdicts that read as current-tree findings.

import (
	"context"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/binlag"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// auditBuildIdentity returns the flat build-identity fields every audit
// verdict carries: the producing binary's build commit, and (only when that
// commit is a strict ancestor of the reviewed tree's HEAD) the rebuild
// advisory naming both commits.
//
// Fail-open by construction (REQ-ABI-005): an unusable build commit yields
// ("", "") and the audit is untouched — a dev build reports no identity
// rather than a fake one. The lag comparison is obtained exclusively through
// binlag.Evaluate, the one ancestry comparison in the tree (REQ-ABI-006);
// nothing here re-implements it.
//
// When no reviewed tree was named (audit_multi's documented normal call —
// resolveOptionalToolProjectRoot returns "" deliberately), the COMPARISON
// falls back to the process working directory, the same precedent
// checkBinaryFreshness follows (internal/cli/doctor.go:521). The fallback
// value lives and dies inside this function: callers keep handing the
// backends exactly the projectRoot they resolved before, so an absent
// project_root still fans out with no cwd at all.
//
// @MX:NOTE: unusable commit ("", "none", "unknown") short-circuits BEFORE the
// comparison — no commit metadata means no identity AND no identity-derived
// advisory, mirroring binlag's own not-applicable leniency
// (internal/binlag/binlag.go:108-110).
// @MX:SPEC: SPEC-AUDIT-BUILD-IDENTITY-001
func auditBuildIdentity(ctx context.Context, projectRoot string) (buildCommit, buildLag string) {
	rawCommit := version.GetCommit()
	buildCommit = normalizeBuildCommit(rawCommit)
	if buildCommit == "" {
		// No commit metadata: a verdict carrying build_lag here would be an
		// identity-derived warning about a binary that identifies nothing.
		return "", ""
	}

	// The comparison needs A tree; the caller-named root when there is one,
	// otherwise the process working directory (comparison ONLY — see above).
	dir := strings.TrimSpace(projectRoot)
	if dir == "" {
		if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
		// Getwd failure keeps dir empty: binlag treats it as not-applicable,
		// so the fallback never turns into a hard failure (fail-open).
	}

	v := binlag.Evaluate(ctx, binlag.Request{
		Dir:           dir,
		BinaryCommit:  rawCommit,
		BinaryVersion: version.GetVersion(),
	})
	return buildCommit, binlag.Advisory(v)
}

// normalizeBuildCommit maps the three build-metadata placeholders to the
// empty string, following the exact set binlag treats as not-applicable
// (internal/binlag/binlag.go:108-110). "none" and "unknown" look like
// identity to a reader while identifying nothing — carrying them would
// revive the misattribution this field exists to close.
func normalizeBuildCommit(commit string) string {
	c := strings.TrimSpace(commit)
	if c == "" || c == "none" || c == "unknown" {
		return ""
	}
	return c
}
