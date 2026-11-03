// Package review holds the provider review / audit adapter layer of
// AgenticFactory (Issue #1948, SPEC-AGENTIC-FACTORY-001 REQ-AF-013, -018).
//
// Codex is reached only through an explicit opt-in key; the GLM managed
// adapter is excluded. Types here are minimal on purpose and are meant to be
// aligned with internal/agenticfactory/core once that package lands.
package review

import (
	"context"
	"errors"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// CodexOptInKey is the single place the opt-in key name lives. The name is
// provisional (SPEC design.md §6); lane-17 owns the final spelling.
const CodexOptInKey = "workflow.agentic_factory.codex.enabled"

// Provider names.
const (
	ProviderClaude = "claude"
	ProviderCodex  = "codex"
	ProviderGLM    = "glm"
)

// Kind names the review being asked for.
type Kind string

const (
	KindCodeReview     Kind = "code-review"
	KindSecurityReview Kind = "security-review"
	KindAudit          Kind = "audit"
	KindQA             Kind = "qa"
)

// Verdict is a provider's judgement of one HEAD.
type Verdict string

const (
	VerdictPass         Verdict = "PASS"
	VerdictPassWithDebt Verdict = "PASS-WITH-DEBT"
	VerdictFail         Verdict = "FAIL"
	VerdictBlocked      Verdict = "BLOCKED"
	VerdictSkipped      Verdict = "SKIPPED"
)

// Request asks a provider to judge one full HEAD SHA.
type Request struct {
	HeadSHA core.HeadSHA
	Kind    Kind
}

// Result is what a provider reports for the HEAD it actually judged.
type Result struct {
	HeadSHA core.HeadSHA
	Verdict Verdict
}

// Provider is one review / audit backend.
type Provider interface {
	Review(ctx context.Context, req Request) (Result, error)
}

var (
	// ErrNotConfigured: the provider needs an explicit setting that is absent.
	ErrNotConfigured = errors.New("review: provider not configured")
	// ErrBinaryAbsent: the provider is opted in but its binary is not installed.
	ErrBinaryAbsent = errors.New("review: provider binary absent")
)

// Lookup reads one configuration key; ok is false when the key is unset.
type Lookup func(key string) (value any, ok bool)

// ProviderEnabled is the only provider activation predicate. Codex is on only
// when its opt-in key is the boolean true; Claude independent audit does not
// depend on any key; GLM and unknown providers are never enabled here.
func ProviderEnabled(name string, lookup Lookup) bool {
	switch name {
	case ProviderClaude:
		return true
	case ProviderCodex:
		if lookup == nil {
			return false
		}
		v, ok := lookup(CodexOptInKey)
		b, isBool := v.(bool)
		return ok && isBool && b
	default:
		return false
	}
}

// CodexDeps are the two outward calls a Codex review can make.
type CodexDeps struct {
	LookPath func(file string) (string, error)
	Run      func(ctx context.Context, req Request) (Result, error)
}

// Codex adapts the Codex reviewer behind the opt-in predicate.
type Codex struct {
	lookup Lookup
	deps   CodexDeps
}

// NewCodex wires the adapter. Nothing is called at construction time.
func NewCodex(lookup Lookup, deps CodexDeps) *Codex {
	return &Codex{lookup: lookup, deps: deps}
}

// Review makes zero outward calls unless the opt-in key is true.
func (c *Codex) Review(ctx context.Context, req Request) (Result, error) {
	if !ProviderEnabled(ProviderCodex, c.lookup) {
		return Result{}, ErrNotConfigured
	}
	if _, err := c.deps.LookPath("codex"); err != nil {
		return Result{}, ErrBinaryAbsent
	}
	return c.deps.Run(ctx, req)
}

// Executor runs one independent Claude audit; the real one is wired later.
type Executor interface {
	Execute(ctx context.Context, req Request) (Result, error)
}

// ClaudeAuditor is the independent-audit adapter that needs no opt-in.
type ClaudeAuditor struct{ exec Executor }

// NewClaudeAuditor wires the adapter to an executor.
func NewClaudeAuditor(exec Executor) *ClaudeAuditor { return &ClaudeAuditor{exec: exec} }

// Review delegates to the executor; HEAD binding is enforced by Record.Admit.
func (a *ClaudeAuditor) Review(ctx context.Context, req Request) (Result, error) {
	return a.exec.Execute(ctx, req)
}
