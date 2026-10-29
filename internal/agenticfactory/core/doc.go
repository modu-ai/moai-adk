// Package core holds the shared vocabulary of AgenticFactory (Issue #1948,
// SPEC-AGENTIC-FACTORY-001 design §2, §3, §4): execution states, pause
// reasons, delivery stages, ledger identifiers, and the error model.
//
// It is a leaf package: no I/O, no dependency on internal/factory. Other
// scopes (state, delivery, ledger, claim, policy, gate) consume these types.
package core
