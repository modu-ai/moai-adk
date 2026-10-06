// Package bugreport is the detection and attribution half of the opt-in
// improvement-participation pipeline (SPEC-FEEDBACK-PARTICIPATION-001): the
// closed error-kind taxonomy, the payload schema, the fingerprint, and the
// fail-open capture entry point.
//
// The package imports the standard library (except os/exec and net/http) plus
// internal/config, and nothing else from this module (design.md section 1) —
// that is what lets every recover site, the hook registry, and the CLI import
// it without forming a cycle, and it is enforced by the static reachability
// guard (REQ-ANON-025, AC-025).
package bugreport

// Kind is the closed enumeration of error kinds the pipeline accepts
// (REQ-ANON-006). A signal is accepted only from a registered emit site, and
// only for one of these six kinds; every other value is a programming error
// the validators refuse.
type Kind string

// The six kinds. The register in design.md section 2 maps each to its default
// attribution verdict, its derivation mode, and the exact code sites that emit
// it.
const (
	// KindPanic is a recovered panic: the main goroutine's deferred recover
	// in cmd/moai/main.go and the registered recover sites in internal/.
	KindPanic Kind = "panic"

	// KindHookHandlerFailure is a hook handler that returned an error
	// (internal/hook/registry.go's handler-error branch).
	KindHookHandlerFailure Kind = "hook_handler_failure"

	// KindHookTimeout is a hook dispatch that hit its deadline
	// (registry.go's ErrHookTimeout branch). It is attributed ambiguous and
	// stays local.
	KindHookTimeout Kind = "hook_timeout"

	// KindInternalError is a violated internal invariant, wrapped by the
	// bugreport internal marker at the call site.
	KindInternalError Kind = "internal_error"

	// KindTemplateDeployFailure is a shipped-template deploy failure
	// (template sentinels mapped to closed tokens by internal/cli).
	KindTemplateDeployFailure Kind = "template_deploy_failure"

	// KindHarnessDefect is a shipped asset that failed to render or validate
	// (renderer/validator sentinels mapped to closed tokens by internal/cli).
	KindHarnessDefect Kind = "harness_defect"
)

// allKinds is the closed set, in registration order. AllKinds returns a copy
// so callers cannot mutate the enumeration.
var allKinds = []Kind{
	KindPanic,
	KindHookHandlerFailure,
	KindHookTimeout,
	KindInternalError,
	KindTemplateDeployFailure,
	KindHarnessDefect,
}

// Valid reports whether k is one of the six closed members.
func (k Kind) Valid() bool {
	for _, member := range allKinds {
		if k == member {
			return true
		}
	}
	return false
}

// AllKinds returns the six kinds in their pinned order. The return value is a
// copy: mutating it must not reach the enumeration.
func AllKinds() []Kind {
	out := make([]Kind, len(allKinds))
	copy(out, allKinds)
	return out
}
