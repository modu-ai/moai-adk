# decision-index.md — SPEC-CI-VERDICT-INTEGRITY-001

Stateless on the status axis — the SPEC lifecycle lives in `spec.md` alone.
Authored at plan phase (card t1534). `interview.decision_gate: on` (`.moai/config/sections/interview.yaml`).

### Q1: Which context list should the SSoT's `release/*` branch key carry after the correction?

Label: FOUNDER

Class: implementation-level

Why unresolved: The card directs the SSoT correction ("required-checks.yml 실제 발행 이름으로 정정") without saying whether the `release/*` key mirrors the corrected `main` list or carries a distinct release-specific list. `RenderBranchProtectionJSON` resolves any branch key present in the SSoT (`internal/cli/branch_protection.go:100-105`), so the answer changes what a future protection apply to a release branch would require and what the watch loop treats as required for PRs based on `release/*`. No committed authority (project docs, prior SPEC HISTORY/Amendments rows, config sections, constitution) settles the question; the live check-run evidence (ledger E13) covers what publishes on main-bound PRs, not a release-base list.

Authority anchor: none reachable in the committed register — routed FOUNDER, never relabeled.

Default: `release/*` mirrors the corrected `main` list (rule: preserves current behavior — the SSoT's own Wave-3 comment records that `release/*` "now matches main exactly", so mirroring continues the file's existing structure).

Alternate: a distinct release-based list (for example dropping Build contexts that do not publish on release-base PRs).

Operator verdict: DEFAULT-APPLIED 2026-10-05T19:41:30Z manager-spec
