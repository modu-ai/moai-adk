package bugreport

// register.go — the kind-to-verdict-to-site register (design.md section 2).
// Each row states the kind's default verdict rule, its derivation mode
// (derived inside bugreport, or asserted by the call site), the closed
// tokens the kind accepts with their verdicts, and the verified emit-site
// leads. The go/parser walk (TestEveryRecoverSiteReportsOrIsAllowlisted) is
// the AUTHORITATIVE recover-site inventory — the site lists here are leads;
// an unlisted site surfaces as a guard failure, never as a silent omission.

// Derivation names the two modes a register row's verdict can carry.
type Derivation string

const (
	// DerivationDerived: the verdict is derived inside bugreport from a type
	// or sentinel bugreport can import (standard library, internal/config, or
	// the marker type defined here).
	DerivationDerived Derivation = "derived"

	// DerivationCallSite: the call site asserts the kind (and, for template
	// and harness kinds, a closed token), because the sentinel lives in a
	// package bugreport cannot import.
	DerivationCallSite Derivation = "call-site-asserted"
)

// KindRow is one register row.
type KindRow struct {
	Kind Kind
	// VerdictRule is the verdict the kind carries when no other row applies:
	// moai for the two inherently-moai kinds, ambiguous for everything else
	// (hook handler errors without a marker, timeouts, unmatched template and
	// harness chains all stay local).
	VerdictRule Verdict
	Derivation  Derivation
	// Tokens are the closed tokens the kind's detail admits, each with the
	// verdict its register row assigns.
	Tokens        []TemplateToken
	TokenVerdicts map[TemplateToken]Verdict
	// Sites are the verified emit-site leads (file:line at the time of the
	// design pass; the walk is authoritative).
	Sites []string
}

// kindRegister is the six-row register, in the pinned kind order.
var kindRegister = []KindRow{
	{
		Kind:        KindPanic,
		VerdictRule: VerdictMoai,
		Derivation:  DerivationCallSite,
		Sites: []string{
			"cmd/moai/main.go:1",
			"internal/guardliveness/evaluator.go:160",
			"internal/cli/codex_stop_chain.go:426",
			"internal/cli/memory_fold.go:1011",
			"internal/resilience/circuit.go:229",
			"internal/navigator/route/run.go:51",
			"internal/navigator/fix/request.go:105",
			"internal/escalation/detector.go:137",
			"internal/hook/session_start_guard_liveness.go:205",
			"internal/hook/user_decision_capture.go:92",
			"internal/hook/session_start_binary_lag.go:72",
			"internal/hook/navigator_detect.go:158",
			"internal/hook/navigator_detect.go:180",
			"internal/hook/session_start_memory_budget.go:125",
			"internal/cli/update_integrity_probe.go:144",
		},
	},
	{
		Kind:        KindHookHandlerFailure,
		VerdictRule: VerdictAmbiguous,
		Derivation:  DerivationCallSite,
		Sites:       []string{"internal/hook/registry.go:133"},
	},
	{
		Kind:        KindHookTimeout,
		VerdictRule: VerdictAmbiguous,
		Derivation:  DerivationCallSite,
		Sites:       []string{"internal/hook/registry.go:125"},
	},
	{
		Kind:        KindInternalError,
		VerdictRule: VerdictMoai,
		Derivation:  DerivationDerived,
		Sites:       []string{"internal/cli/preference/cmd.go:221"},
	},
	{
		Kind:        KindTemplateDeployFailure,
		VerdictRule: VerdictAmbiguous,
		Derivation:  DerivationCallSite,
		Tokens:      []TemplateToken{TokenPathTraversal, TokenNotFound, TokenPreserveIntegrity},
		TokenVerdicts: map[TemplateToken]Verdict{
			TokenPathTraversal:     VerdictMoai,
			TokenNotFound:          VerdictMoai,
			TokenPreserveIntegrity: VerdictMoai,
		},
		Sites: []string{
			"internal/cli/update_template_sync.go:535",
			"internal/cli/update_clean_install.go:599",
		},
	},
	{
		Kind:        KindHarnessDefect,
		VerdictRule: VerdictAmbiguous,
		Derivation:  DerivationCallSite,
		Tokens:      []TemplateToken{TokenMissingKey, TokenUnexpandedToken, TokenInvalidJSON},
		TokenVerdicts: map[TemplateToken]Verdict{
			TokenMissingKey:      VerdictMoai,
			TokenUnexpandedToken: VerdictAmbiguous,
			TokenInvalidJSON:     VerdictAmbiguous,
		},
		Sites: []string{
			"internal/cli/update_template_sync.go:439",
		},
	},
}

// KindRegister returns the six register rows in pinned order.
func KindRegister() []KindRow {
	out := make([]KindRow, len(kindRegister))
	copy(out, kindRegister)
	return out
}

// RecoverAllowlist returns the reasoned allowlist entries: recover sites the
// walk must accept without a capture call. The reason IS the entry's note.
var recoverAllowlist = []struct {
	Site   string
	Reason string
}{
	{
		// The recover defends a closed-channel send race after Close() and
		// drops one trace entry by design — benign by construction.
		Site:   "internal/hook/trace/writer.go:Write",
		Reason: "closed-channel send race after Close; drops one trace entry by design",
	},
}

// AllowlistedRecoverSites returns the allowlist's site identifiers
// ("file:function").
func AllowlistedRecoverSites() []string {
	out := make([]string, 0, len(recoverAllowlist))
	for _, a := range recoverAllowlist {
		out = append(out, a.Site)
	}
	return out
}
