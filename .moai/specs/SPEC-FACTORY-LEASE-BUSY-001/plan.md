# Plan — SPEC-FACTORY-LEASE-BUSY-001 (card t1916)

Stateless plan artifact (no `status:` field — the SPEC's lifecycle lives in `spec.md` frontmatter).

## §A Context

- **Card**: t1916 (reissue of t1743), P1. **Tier**: M (3-artifact set + progress + decision-index;
  `decision_gate: on` per `.moai/config/sections/interview.yaml`). **Worktree**:
  `.moai/worktrees/t1916`, branch `WT-factory-next-raced`, plan-start HEAD `8249bff8c3`.
- **Methodology**: TDD (RED-GREEN-REFACTOR) — `.moai/config/sections/quality.yaml`
  `constitution.development_mode: tdd`.
- **§A.1 Problem.** Two axes, both in `internal/cli`: (1) the lease path's claim-failure
  classification merges host-load deadline expiry and record-write contention into
  `errFactoryRecordBusy` (`factory_card.go` 492-510) and maps both onto the `raced` token at exit 4
  (`factory_card.go` 1600-1614), masking the real cause; (2) the `factory status` read path
  (`newFactoryStatusCommand`, `factory_card.go` 3335+) is observed slow; unmeasured.
- **§A.2 Operative contract and classification honesty.** The reissued card text labels this work
  `Class B` — a label mismatch. The operative contract (leader dispatch 03:16Z reissue + leader
  constraint 03:26Z) is the original card scope WITH the SPEC binding: plan → run → sync,
  SPEC-FACTORY-LEASE-BUSY-001 authored in plan. No ceremony reduction; the session record and
  `.moai/reports/t1916/file-map-survey.md` §D carry the actual contract.
- **§A.3 Sequencing precondition (binding, leader directive).** Run-phase implementation is **HELD
  until PR #1859 lands**. Survey evidence (measured, not inferred):
  `.moai/reports/t1916/file-map-survey.md` — FILE overlap YES on `internal/cli/factory_card.go`
  (and `mcp_factory_card.go` via the MCP classification relay); hunk-level no collision today, but
  #1859's claim-path hunks (991-1044 reshaped, +138 inserted at 1288, its own old base) bracket this
  fix's claim-call-site region (1133-1220), so a same-file rebase is the expected integration shape
  after it lands. The survey file is machine-local and untracked; it is cited here per the leader's
  directive and is not an authority-register artifact. The hold is a sequencing precondition, NOT a
  scope reduction: requirements and ACs are fully specified now. Resume is coordinated by the
  leader. Milestone M0 encodes the gate.
- **§A.4 PRESERVE list.** SPEC-FACTORY-ATOMIC-LEASE-001 artifacts and its AC-FAL-007/015 timing
  bounds; the lease-bound constants (`factory_card.go` 68-76); the queue-lock refusal sites
  (`factory_card.go` 1527-1528, 1567, 1570) and their tests; the serial-slot / quota-hold /
  backend-skip / foreign-worktree refusal texts; `factory status` output fields and JSON keys; every
  other card's artifacts.

## §B Known Issues (domain-filtered)

- **B2 (cross-SPEC conflict).** PR #1859 (OPEN, active) shares `factory_card.go` +
  `mcp_factory_card.go` — the §A.3 hold. SPEC-FACTORY-ATOMIC-LEASE-001 is in-progress amendment;
  its amendment owns the adjacent timing machinery — do not touch its bounds or its AC text.
- **Pinned-test flips are part of the change, not collateral.**
  `TestFactoryLeaseRecordStallBounded` (its doc comment and its `raced` assertion) and
  `TestFactoryLeaseMidClaimStallBounded` (its `raced` assertion) pin the current misclassification
  (record stall → `raced`); REQ-FLB-003/004 legitimately flip them. The queue-lock pins —
  `TestFactoryLeaseQueueLockStallBounded`, subtests `nominated` and `bare` — stay untouched and
  keep passing (out of scope, decision-index Q4).
- **B6 (spec-lint heading convention).** `spec.md` carries `### Out of Scope —` H3 sub-headings with
  `-` bullets (satisfies `OutOfScopeRule`).
- **B8/B10 (hygiene, scope discipline).** Stage by explicit pathspec; no runtime-managed files; no
  other card's artifacts; no worktree-external writes.
- **B12** applies to manager-docs at sync (CHANGELOG count + live-identifier check).

## §C Pre-flight (before the first implementation commit, after M0 opens)

```bash
git branch --show-current && git rev-parse HEAD        # expect WT-factory-next-raced, post-rebase HEAD
gh pr view 1859 --json state,mergedAt                  # M0 gate: state == MERGED before M1 starts
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5          # baseline: NEW vs pre-existing
```

Plus the LSP baseline capture required by `quality.yaml` `lsp_quality_gates.plan.require_baseline:
true`.

## §D Constraints

- Plan phase: documents only (this artifact set). Implementation begins only after M0.
- No lease-bound resizing; no exit-code changes beyond decision-index Q1's resolution (the
  published default — keep exit 4 for every class — holds until the operator answers).
- Conventional Commits with the card id in every commit body and the `🗿 MoAI` trailer; landing per
  the gitflow lane protocol (`moai factory complete` → integration merge; the leader batches
  remote pushes). Never `--no-verify`.
- One classification implementation shared by the CLI verb and the MCP handler — no relay-side
  re-classification in `mcp_factory_card.go`.
- TDD: every behavior change lands RED-first; E8 carries the verbatim pre-GREEN failing output.

## §E Self-Verification deliverables

Per `manager-develop-prompt-template.md` §E and the verification-claim-integrity 5-section format:
E1 AC PASS/FAIL matrix (all eleven AC-FLB ids, command + verbatim output + HEAD per row); E2
cross-platform build (`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`); E3 coverage on
the touched packages (`go test -cover ./internal/cli/...` and any `internal/homestate` touch)
against the 85% target; E4 subagent-boundary grep on touched dirs; E5 lint status (NEW vs baseline);
E6 commit SHAs + landing state; E7 blocker reports (never prose questions); E8 verbatim RED outputs
per TDD milestone. Ordering attribution for AC-FLB-008: the baseline commit precedes the first
read-path change commit — git is the only sequencing witness.

## §F Milestones (ordered by decision-reversibility: the classification contract first, mechanical steps last)

- **M0 — Hold gate (precondition, no code).** Verify PR #1859 is merged and local `main` carries
  it; obtain the leader's resume signal; rebase this card's branch; re-run the §C pre-flight.
  Priority High. Nothing else starts before M0 closes.
- **M1 — Cause-class contract + refusal surface (RED).** Fix the class table — race loss /
  record-write contention / claim-deadline expiry — and the reason tokens; write the failing tests
  per class through the existing seams (`factoryLeaseClaimDeadlineOverride`, `flHoldRecord`,
  `factoryLeaseBeforeClaim`). Deliverable: RED outputs for AC-FLB-001..004 assertions plus the
  token/class table recorded in progress.md §E.2 (run-phase evidence; this plan file's body is not
  manager-develop-editable, so the table lands there — the classes themselves are already normative
  in spec.md REQ-FLB-001). Priority High. (Decision-reversibility:
  the token names and class boundaries are the decisions most likely to move in review — they lead.)
- **M2 — GREEN reclassification.** Split `factoryClaimFailure`'s marking (492-510) into the
  contention and deadline classes; map them at the nominated switch (1600-1614) and keep the race
  family on `raced`; update the pinned assertions in `TestFactoryLeaseRecordStallBounded` and
  `TestFactoryLeaseMidClaimStallBounded`; extend `waitable()` (`factory_card.go` 1270-1276) so the
  new record-contention and claim-deadline tokens are wait-retryable exactly as `raced` is today,
  keeping the `--wait` loop waiting through transient contention (REQ-FLB-008, AC-FLB-011);
  preserve one-line rendering (REQ-FLB-006) and exit 4 (REQ-FLB-007); the MCP relay inherits
  unchanged. Priority High.
- **M3 — Slot-read race-claim hygiene.** Verify the lease path attributes no race loss to a
  serial-slot outcome (REQ-FLB-009): the slot-hold refusal keeps its `serial-slot` token and text
  (`factorySerialInFlightExcluding`, `factory_card.go` 330, refusing at 1368), and the peek's
  store-read failure keeps failing open with no race claim in any rendered line
  (`factoryNextMayLease`, `factory_card.go` 1868-1878). Test via the existing serial-slot arm of
  the lease tests plus a forced peek store-read failure (AC-FLB-007). No reclassification of a
  genuine serial-slot outcome — that would collide with REQ-FLB-010's byte-identical preservation.
  Priority Medium.
- **M4 — Status baseline + profile (evidence-first; ordering-attribution).** Capture
  `.moai/reports/t1916/status-latency-baseline.md` — N ≥ 20 `factory status` invocations, host
  1-minute load at or below the core count (the card-t1746 amendment's load-condition convention),
  p50/p95 — committed in its own commit BEFORE any read-path change. The recipe is executable end
  to end and identical across revisions: (1) **binary** — built per revision from the tree under
  measurement (`go build -o <scratch>/moai-<revision> ./cmd/moai`) and invoked by that path, never
  the installed `moai` (whose build lags this tree); (2) **isolated root** — one fresh `mktemp -d`
  project directory serves both runs (`FACTORY_ROOT="$(mktemp -d)"`), never the primary checkout's
  live store (`factory status` is read-only, so one root keeps the fixture and the measurement
  comparable); (3) **fixed fixture** — seeded once into that root by the real seed verbs before the
  first run: `moai todo add <fixture-card>` then `moai factory assign <fixture-card> --run
  <fixture-run>` (executed with the baseline binary against the isolated root; the exact seed
  command lines, including flags, are recorded verbatim in the baseline artifact and reissued
  unchanged for the post-change run); (4) **exact command** — the invocation line is recorded
  verbatim in the baseline artifact and reissued verbatim for the post-change measurement with
  exactly one token changed: the binary path, the sole revision-varying measurement parameter —
  every other argument, environment variable, and fixture input is byte-identical across the two
  runs. Profile the read path
  (`EffectiveFactoryLease` / `LoadPure` / `OpenFactory`+`ListCards` / `ReadRecordUnavailable`) and
  name the dominant cost. Priority High.
- **M5 — Status read-path fix + improvement measurement.** Remove or restructure the dominant cost;
  re-measure with the M4 recipe; record the AC-FLB-009 outcome. Priority High.
- **M6 — Regression sweep + sync handoff.** Scoped lease-path suite green with updated pins
  (AC-FLB-010); cross-platform builds; MX tag updates for the changed functions; hand off to
  manager-docs. Priority Medium.

## §G Anti-patterns (refuse these)

- Re-tokening the queue-lock `raced` refusal (out of scope; Q4).
- Resizing the claim deadline, busy timeout, or wait cap while "fixing" the classification.
- Splitting the classification between the CLI verb and the MCP handler.
- Editing #1859's hunks during the rebase; resolving rebase conflicts by reclassifying their code.
- Writing a latency improvement number without the M4 baseline; citing the baseline from a prior
  run or another tree (carry-over is not a baseline).

## §H Cross-references

- `spec.md` REQ-FLB-001..012; `acceptance.md` AC-FLB-001..011 and the REQ→AC traceability table.
- `decision-index.md` Q1-Q4; `.moai/reports/t1916/file-map-survey.md` (§A.3 sequencing evidence).
- SPEC-FACTORY-ATOMIC-LEASE-001 (claim machinery, AC-FAL-007), SPEC-TODO-AUTO-PICK-001 §C.2
  (`raced` definition), SPEC-FACTORY-LEASE-RENEW-001 (status lease fields).
