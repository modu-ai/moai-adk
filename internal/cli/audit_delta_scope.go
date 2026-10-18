package cli

// Card t1708 — delta scope for codex_audit / audit_multi.
//
// A re-audit of a card that was already reviewed up to some commit used to see
// the full diff from the resolved base, so pre-existing findings re-surfaced on
// every pass. The explicit optional input `delta_base` names that previously
// audited commit; every backend then measures from it instead of the resolved
// base. Absent ⇒ none of this code runs and behaviour is byte-identical.
//
// The base travels on the context (the withCodexAuditLeg precedent) rather than
// through backendCallFn, whose five-argument signature many tests assign.

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// deltaBaseRef is the Ref a delta scope reports in review_base.
const deltaBaseRef = "delta_base"

// deltaBaseDescription is the tool-schema text of the delta_base input.
const deltaBaseDescription = "Optional delta scope: a commit SHA, an ancestor of HEAD, that this change was already audited up to. With target 'baseBranch' the review measures `delta_base...HEAD` instead of the resolved base, so findings that predate it do not re-surface. Claude, GLM and codex native honour it exactly; codex adversarial (and the audit_multi codex leg) is told in the prompt only, and the result says so. An invalid value, or one used with another target, is a tool error — never a silent full-diff review. audit_multi writes no second-review record for a delta call. Omitted ⇒ unchanged behaviour."

type deltaBaseCtxKey struct{}

// withDeltaBase marks ctx as a delta audit from sha (a full, validated commit).
func withDeltaBase(ctx context.Context, sha string) context.Context {
	return context.WithValue(ctx, deltaBaseCtxKey{}, sha)
}

// deltaBaseFromContext returns the delta base carried on ctx, or "".
func deltaBaseFromContext(ctx context.Context) string {
	sha, _ := ctx.Value(deltaBaseCtxKey{}).(string)
	return sha
}

// deltaBaseHex bounds what a caller may name: abbreviated or full commit hex.
// Refusing everything else keeps option-shaped and revision-expression values
// away from git.
var deltaBaseHex = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)

// validateDeltaBase resolves raw to a full commit SHA in root and requires it to
// be an ancestor of HEAD. Any failure is an error: a delta_base that cannot be
// honoured must never degrade into a full-diff review that looks scoped.
func validateDeltaBase(root, raw string) (string, error) {
	sha := strings.TrimSpace(raw)
	if !deltaBaseHex.MatchString(sha) {
		return "", fmt.Errorf("delta_base %q is not a commit SHA (7-64 hex digits)", raw)
	}
	out, err := runReviewGit(root, "rev-parse", "--verify", "--quiet", sha+"^{commit}")
	full := strings.TrimSpace(out)
	if err != nil || full == "" {
		return "", fmt.Errorf("delta_base %s is not a commit in %s", sha, root)
	}
	// Nothing reachable from delta_base may be missing from HEAD: an empty
	// HEAD..delta_base range means delta_base is already part of HEAD's history.
	ahead, err := runReviewGit(root, "rev-list", "--count", "HEAD.."+full)
	if err != nil || strings.TrimSpace(ahead) != "0" {
		return "", fmt.Errorf("delta_base %s is not an ancestor of HEAD", full)
	}
	return full, nil
}

// deltaBaseArg reads the raw delta_base argument. GetString returns its default
// for any non-string value, which would read an array, object, number, bool or
// null as "absent" and silently widen the call to a full-scope review, so a
// present key that is not a string is an error here. Absent ⇒ "".
func deltaBaseArg(req mcp.CallToolRequest) (string, error) {
	raw, present := req.GetArguments()[deltaBaseRef]
	if !present {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("delta_base must be a string (a commit SHA), got %T", raw)
	}
	return s, nil
}

// readDeltaBase validates the optional delta_base argument against target. It
// returns "" when the argument is absent. A delta scope is a base-branch scope,
// so any other target is rejected rather than silently ignoring the value.
func readDeltaBase(root, raw, target string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	if target != codexTargetBaseBranch {
		return "", fmt.Errorf("delta_base needs target %q (got %q): there is no base to move for any other target", codexTargetBaseBranch, target)
	}
	return validateDeltaBase(root, raw)
}

// deltaReviewBase is the reviewBase a delta scope measures from.
func deltaReviewBase(sha string) reviewBase {
	return reviewBase{Ref: deltaBaseRef, MergeBase: sha}
}

// deltaProseReviewBase is the review_base text of a codex adversarial audit: the
// scope reaches codex as prompt text only, and the result says so.
func deltaProseReviewBase(sha string) string {
	return deltaBaseRef + " " + sha + " (prose-only scope: codex adversarial mode cannot be bounded mechanically)"
}

// deltaScopeNote is the residual_risk_note clause of a delta audit_multi call.
// It states only what the verdicts show: a claude or glm leg counts as having
// reviewed delta_base...HEAD when it produced a pass/fail verdict itself; the
// caller's in-session Claude anchor arrives with a scope the server cannot know.
// The codex clause appears only when a codex verdict is present.
func deltaScopeNote(sha string, verdicts []PerBackendVerdict) string {
	var reviewed []string
	anchor, codexRan := false, false
	for _, v := range verdicts {
		if v.Backend == BackendCodex {
			codexRan = true
		}
		if v.Backend != BackendClaude && v.Backend != BackendGLM {
			continue
		}
		switch {
		case v.Source == "in_session_anchor":
			anchor = true
		case v.Verdict == "pass" || v.Verdict == "fail":
			reviewed = append(reviewed, v.Backend)
		}
	}
	var clauses []string
	if len(reviewed) > 0 {
		clauses = append(clauses, strings.Join(reviewed, "/")+" reviewed "+sha+"...HEAD")
	}
	if anchor {
		clauses = append(clauses, "the in-session claude result is the caller's (scope not verified)")
	}
	if len(clauses) == 0 {
		clauses = append(clauses, "no claude or glm leg produced a verdict on it")
	}
	codex := "codex did not run on it"
	if codexRan {
		codex = "codex was scoped in prose only (not mechanically bounded)"
	}
	return "delta_base " + sha + ": " + strings.Join(clauses, "; ") +
		"; " + codex + "; no second-review record and no session state written"
}

// codexDeltaScopeClause extends the adversarial prompt for a delta audit. The
// adversarial path starts a turn (turn/start), which has no target field, so the
// base can only be asked for in words.
func codexDeltaScopeClause(sha string) string {
	return " Delta audit scope: this change was already reviewed up to commit " + sha +
		". Review ONLY what changed after it (`git diff " + sha + "...HEAD`); code and findings that predate that commit are out of scope, so do not report them."
}

// codexAdversarialPromptFor is the adversarial prompt for ctx: the standard one,
// plus the delta clause when ctx carries a delta base.
func codexAdversarialPromptFor(ctx context.Context, focus string) string {
	prompt := codexAdversarialReviewPrompt(focus)
	if sha := deltaBaseFromContext(ctx); sha != "" {
		prompt += codexDeltaScopeClause(sha)
	}
	return prompt
}

// collectReviewDiffCtx is collectReviewDiff for a backend that holds the call's
// context: a delta base on ctx replaces the resolved merge base of a baseBranch
// review, anything else is collectReviewDiff unchanged.
func collectReviewDiffCtx(ctx context.Context, root, target string) (string, error) {
	if sha := deltaBaseFromContext(ctx); sha != "" && target == codexTargetBaseBranch {
		return collectReviewDiffAt(root, deltaReviewBase(sha))
	}
	return collectReviewDiff(root, target)
}
