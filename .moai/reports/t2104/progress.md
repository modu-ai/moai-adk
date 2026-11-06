# t2104 progress (GitHub #1938)

## Claim
The premise "no formal path moves an expired leased row from a dead run into a live run" is false at this base: `moai factory runs --retire <dead> --rebind-to <live>` does it. The real residual is ordering: assigning the card into the live run first creates a second row, and the consolidation then refuses.

## Evidence
- Probe (throwaway, deleted): `go test ./internal/cli -run TestZZT2104Probe -v`
  - pre=false: `rebound run-src into run-tgt (owner dead) / moved cards: t1 / abandoned cards: none`; rows: `run-tgt:leased:lane-9`
  - pre=true: `factory run consolidation refused: card t1: run run-tgt already holds a assigned row diverging on lease_holder, lease_expires_at ...`; rows: `run-src:leased:lane-9 ; run-tgt:assigned:lane-1`
- Earlier probe on the base: `assign --reassign --run <dead>` -> `run "run-dead" is not the currently active factory run`; `assign t1 --run <live>` creates a new row, `status` lists both.
- Pinned by `internal/cli/factory_runs_rebind_expired_t2104_test.go` (2 tests): `go test ./internal/cli -run 'TestFactoryRunsRebind(MovesExpired|RefusesCardAlready)' -v` -> both PASS (8.248s).
- Doc: `.moai/docs/factory-expired-lease-dead-run.md`.

## Baseline-attribution
Tree: worktree t2104, HEAD ee68245c57 (t2091 tip 27adc925ef absorbed). Tests above ran in this run against this tree.

## Gaps
- Harm of the two-row state to `factory complete` / status counting NOT measured; status was observed to print both rows only.
- Tests pin existing behavior; there was no RED step because no code changed (decision: document + pin, per the card's "no harm shown" branch).
- A worktree-guard refusal of a heredoc command (probe creation) occurred; the probe was rewritten with the Write tool instead.
- Windows/CI not run; focused tests only.

## Residual-risk
A dead run holding a non-terminal row cannot be retired without the rebind verb; an operator who assigned first is left with two rows and needs a decision. No code fix proposed (minimal-change); if complete is shown to misbehave on two rows, a follow-up card should add a retire-verb.
