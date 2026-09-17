# SPEC-SYNC-GATE-VERDICT-001 — Progress

card t783 · branch `WT-syncgate-hook` · base `a404132e7`

## Plan-phase Record

- Authored 2026-09-14 by manager-spec in the card worktree (`.claude/worktrees/t783`),
  Tier M, 3 artifacts (spec.md / plan.md / acceptance.md) + this progress.md.
- All four target files read in full on the base tree; every reproduction fact re-measured
  in this run (see §E.1 preface table). Baseline discipline: the audit's line numbers are
  main-based (`2213871af`); all targets located by symbol/phrase on develop.
- Key plan-phase finding recorded for the auditor: card t783's H01 machinery was already
  delivered by SPEC-SYNC-GATE-FAILSTATE-001 (card t624, completed 2026-09-11); t783's H01
  residual is the card-mandated three-arm EXECUTION proof on the current tree plus a
  conditional minimal repair (REQ-SGV-003). SX-R05 is genuinely unresolved (measured: old
  trio live in the template copy; BOTH copies carry the relationship paragraph's two stale
  clauses — the local copy already contradicts itself) and is this SPEC's substantive edit.
  H03 is wording-verified only (baseline phrases demonstrably present on `2213871af`,
  absent on both develop copies).
- **v0.2.0 amendment (2026-09-14)** — plan-audit PASS 0.88 (threshold 0.80), 6 findings,
  `.moai/reports/t783/plan-audit.md`. F1 (High) resolved as **option A**: the Phase 8
  relationship paragraph's two stale clauses ("its CRITICAL-only stop gate below"; "a HIGH
  finding that Phase 8 reports only as a warning") are aligned OUT of BOTH copies in M3 —
  the freeze is dropped, the clauses leave the text (AC-SGV-006/009 removal greps evidence
  the removal; option B's grep-watched survival was rejected as the vacuous-green shape).
  F2: all REQ bodies reflowed SHALL-first. F3: write-ordering clause declared
  consumed-from-FAILSTATE-001 (torn-write shims own it; AC-SGV-003 states the boundary).
  F4: AC-SGV-007(b) parity given a mechanical proxy (normalized-file diff, exit 0).
  F5: hook neutrality tightened to no-NEW-card-IDs-on-edited-lines (AC-SGV-007(e)).
  F6: baseline-hook gate-layout note added (plan.md B10 + pre-flight). Artifact hash
  changed ⇒ the run-phase Plan Audit Gate re-executes (skip-eligibility intentionally
  invalidated).
- **v0.2.1 amendment (2026-09-14)** — iter-2 verdict PASS 0.94, F1-F6 verified closed; one
  residual Low (F7) folded in: the stale-clause removal patterns have 0 hits on the
  `2213871af` baseline doc (measured this run; the relationship paragraph postdates the
  baseline via t624's M3 `c0e56ab09`), so their pattern-proof is re-scoped to the CURRENT
  pre-M3 tree (1 hit each in both current copies), and the `2213871af` baseline positive
  control applies to the H01 hook-state arms and the H03 absence pattern only. Amended:
  AC-SGV-006 Given, AC-SGV-005 Given (absence-pattern scoping, same class), acceptance §A,
  plan M1 row. No milestone work started (plan-phase amendment only).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-14

Plan-phase self-check (re-affirmed at v0.2.0): frontmatter carries the canonical 12 fields +
`tier: M` + `related_specs` (no snake_case aliases); REQ count 8 / AC count 9 — inside Tier M
ceilings (≤16 / ≤16); acceptance.md carries Given-When-Then per AC; plan.md names
Template-First ordering and the divergence-integration requirement (deliberate merge, no
verbatim cp) as explicit M3 steps and anti-patterns; no `make build` anywhere in the plan;
evidence convention `.moai/reports/t783/` (untracked, primary checkout) named in REQ-SGV-008
and AC-SGV-008; the 계기 observer contract is encoded as the M1 positive controls
(AC-SGV-001 baseline-hook reproduction; AC-SGV-005 baseline-proven grep patterns).

## Implementation Kickoff Approval

- Approved by the operator on 2026-09-18, relayed by the kanban lead session (dispatch for card t783). Progression mode: autonomous — run (M1→M4 serial) then sync, without intermediate stops.
- Scope of the approval: run-phase entry only. It is not authorization for a PR, a push (lead batch-pushes develop), `make build` (REQ-SGV-008), or any destructive operation.
- Tree at approval: local develop `ca2dae9a6` absorbed (includes t602 `693cf3eb9`, which edited both hook copies in the java/kotlin/ruby/php/scala/R fast-check branches only). Re-check after absorption: hook pair `cmp` exit 0; `git log --since=2026-09-14` on both doc copies → 0 commits; spec/plan/acceptance unchanged since `2275e0241`.
- Plan Audit Gate: prior report lacks machine hash metadata (cache miss) and the tree moved (t602), so Phase 1 re-executes as iteration 3 before M1.

## §E.2 Run-phase Evidence

Authored by manager-develop, 2026-09-18, card t783. Tree: worktree `.claude/worktrees/t783`,
branch `WT-syncgate-hook`, run entry HEAD `4d277d1d9`, code-final HEAD `23e8cd61e`. Evidence
files are untracked under `.moai/reports/t783/run/` (index: `evidence-index.md`).

**Diff-range base (plan-audit F8).** `CARD_BASE=$(git merge-base develop HEAD)` →
`ca2dae9a6428e62e380c3104d4b2cb208fccca73`, re-derived at pre-flight and again at M4. Non-vacuity
control: `git diff --name-only $CARD_BASE..HEAD` → 6 paths (4 SPEC artifacts + 2 doc copies).
`a404132e7` is used only for ancestry (`git merge-base --is-ancestor` exit 0) and as a diff-range
positive control (7378-byte hook diff carrying t602's hunks).

**Fixture recipe (frozen: `run/harness.sh`).** `mktemp -d /tmp/t783-fixture.XXXXXX` git repo,
`go.mod` + `main.go`, HEAD subject `docs(SPEC-FIXTURE-001): sync-phase artifacts` with a `.go`
delta, `main.go` calling an undefined symbol (vet and build both fail), stdin `{}`,
`CLAUDE_PROJECT_DIR` = fixture, EXIT-trap cleanup. The harness unsets `MOAI_SYNC_GATE_BLOCKING`
and `MOAI_AUTONOMY_TIER`: the calling session had `MOAI_AUTONOMY_TIER=fully-autonomous`, which
would have forced advisory mode. A pass-through `go` shim records the outcome record at check
start. B10: the same recipe drove the `2213871af` hook to its checks, so no adaptation was needed
(recorded, not assumed: `m1-control/control-go-invocations.txt` shows vet and build ran).

| AC | Status | Command | Actual output (verbatim key lines) | Evidence |
|---|---|---|---|---|
| AC-SGV-001 (control first) | PASS | `harness.sh control run/baseline-hook.sh` (baseline hook byte-identical to prior extract, `cmp` 0) | call1 `{"hookSpecificOutput":{"hookEventName":"Stop","decision":"block","reason":"go vet failed"},...}` bytes=237 exit=0; call2 bytes=0 exit=0 | `run/m1-control/` |
| AC-SGV-001 (arm A) | PASS | `harness.sh arms .claude/hooks/moai/sync-phase-quality-gate.sh` | call1 bytes=237 exit=0; call2 bytes=237 exit=0; `cmp_exit=0`; gate log `mode=blocking decision=block-redelivered`; go invocations after arm A: one vet/build pair | `run/m2-arms/`, `run/m2-arms-r2/` |
| AC-SGV-002 (arm B) | PASS | same harness, `main.go` repaired in work tree, HEAD `b16c3254…` before and after | call3 bytes=0 exit=0; new run line `decision=allow go vet=0 go build=0`; record `b16c3254… pass 453af84b…` (fail record carried `e3b0c442…`) | `run/m2-arms/arms-armB-*` |
| AC-SGV-003 (arm C) | PASS | same harness, calls 4 and 5 on the unchanged tree | call4/call5 bytes=0 exit=0 in all three runs; record-at-check-start `<sha> running <wci>`; failing run payload `<sha> block 1 1` + block JSON | `run/m2-arms*/arms*-armC-*`, `*-record-at-check-start.txt` |
| AC-SGV-004 | PASS (branch a) | `git diff $CARD_BASE..HEAD -- <both hook paths>` | 0 bytes | `run/m4-hook-diff-cardbase.txt` |
| AC-SGV-005 | PASS | `phrases.sh` + hook diff | "vulnerability scan runs automatically" 0/0 (baseline doc 1); "not a vulnerability scan" 2/2, hook `679:# not a vulnerability scan.`; "deps_modified" 1/1; hook diff 0 bytes | `run/m4-analysis.txt` |
| AC-SGV-006 | PASS | `phrases.sh` (template / local) | "Only CRITICAL findings block" 0/0; "HIGH findings are reported as warnings" 0/0; "Continue with warning" 0/0; "CRITICAL-only stop gate" 0/0; "reports only as a warning" 0/0; "Continue by approved exception" 1/1; template 147 / local 168 carry "finding ID, rationale, scope, approver, expiry, and review condition"; template "security-decision-contract" 0. Pre-M3 controls: stale clauses 1/1, trio 1 in template | `run/m4-analysis.txt` |
| AC-SGV-007 | PASS | `cmp` hooks; `parity.sh`; `git diff -U0 $CARD_BASE..HEAD -- <doc>`; `git log $CARD_BASE..HEAD` | (a) `cmp_exit=0`; (b) `diff_exit=0` bytes=0 (pre-M3 control `diff_exit=1` bytes=1100; one excluded delta line, local only); (c) SPEC-ID/card/date/SHA/audit/rule-path hits 0 over 7 added lines (controls on progress.md 2/10/7/7/7); (d) every hunk inside the Step 0.55.1..Phase 9 region; (e) 0 added hook lines; template commit `ff0031e2d` precedes local `23e8cd61e` | `run/m3-parity*/`, `run/m4-*` |
| AC-SGV-008 | PASS (location corrected after sync-audit F1) | `git ls-files .moai/reports/t783`; `git log --name-only $CARD_BASE..HEAD`; grep over evidence; export check | Location: the evidence was first written only to the worktree's `.moai/reports/t783/run/`, and the first PASS did not measure the PRIMARY-checkout location. Sync-audit F1 caught this. The orchestrator then exported it without committing: `cp -Rn <worktree>/.moai/reports/t783/run /Users/goos/MoAI/moai-adk-go/.moai/reports/t783/`, with `diff -rq` → exit 0 and no output. Re-measured by manager-develop: `ls /Users/goos/MoAI/moai-adk-go/.moai/reports/t783/run \| wc -l` → `41` (worktree run dir also `41`); `diff -rq <worktree run> <primary run>` redirected to a file → `exit=0`, file bytes `0`. Other checks: tracked 0 bytes; `.gitignore:229:.moai/reports/*`; report paths in card commits 0 (control: spec paths 15); `make build` hits in evidence 0 (control spec.md 3) | `run/m4-reports-*`, `run/m4-makebuild-hits.txt` |
| AC-SGV-009 | PASS | `phrases.sh` | "sync-auditor FAIL" 1/1; "additional lens" 1/1; stale clauses 0/0 | `run/m4-analysis.txt` |

**Invariants.** Hook pair byte-identical (`cmp_exit=0`) and unchanged; no scanner added; the
`deps_modified` block and the per-language `case` are untouched (empty hook diff).

**Finding (not an AC failure).** In `m2-arms-r2`, `go build ./...` wrote an untracked binary into
the fixture after the work-tree id was computed. As a result call4 re-ran the checks once, and
call5 did not. With the binary ignored (`m2-arms-ignored`), call4 ran no checks. Stdout stayed
empty and no stale block was re-delivered. A repair that recomputes the id after the checks would
loosen the gate: an edit made during the check window could be recorded as passed, which
REQ-SGV-005 forbids. No hook change was made; this is left for the lead to consider as a card.

**Regression surface (catalog blocker — RESOLVED by commit `9fa9bc40b`, see §E.3).** First
measurement: `go test -count=1 ./internal/template/` exit=1 with three failures. `TestManifestHashFormat` (`CATALOG_HASH_UNSTABLE: moai stored hash=0afffd09…, computed
hash=a9f58803…`) and `TestCatalogHashCoversSkillSubfiles` (`CATALOG_HASH_SKINNY`) are caused by
the M3 template doc edit: with the pre-M3 template doc swapped back, both PASS (worktree restored,
`git status` clean). `TestGTDCanonicalSurfaceGolden` (`templates/.claude/commands/moai/todo.md is
not a thin gtd compatibility path`) also FAILs on the pre-M3 doc, so it was already failing before
this card. The repair (`internal/template/catalog.yaml` hash regen) was outside this card's
4-file scope fence and was returned as a blocker. The orchestrator resolved it as option (a),
regeneration owned by this card, following the t628 precedent that a known red never rides into
develop. The regeneration landed in `9fa9bc40b`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-18
run_commit_sha: 23e8cd61e0e6188bde65363e0765d840f998e993  # run-phase code-final commit; catalog cascade recorded separately as catalog_hash_commit_sha
run_code_final_sha: 23e8cd61e0e6188bde65363e0765d840f998e993
run_status: audit-ready
ac_pass_count: 9
ac_fail_count: 0
card_base: ca2dae9a6428e62e380c3104d4b2cb208fccca73
hook_repair_applied: false
catalog_hash_commit_sha: 9fa9bc40ba95695ec943bd4558d594711d88a38c
open_blocker: none
preexisting_failure_not_attributable: TestGTDCanonicalSurfaceGolden
new_warnings_or_lints_introduced: none-measured (doc-only change; no Go lint run)
pushed: false
```

**Catalog hash cascade (resolved).** Commands and outputs below were measured and reported by the
orchestrator on this worktree. The generator was not re-run by manager-develop.

- `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai --dry-run` →
  `a9f58803381ade51be00472add9be95b9468affd406466066f5900c130d5533e`. This matches the failing
  test's `computed hash=` byte for byte (all 64 characters).
- The generator was then written without `--dry-run`. `git diff --numstat -- internal/template/catalog.yaml` → `1 1`. The diff
  touches only the moai hash line, `0afffd09…` → `a9f58803…`; manager-develop re-observed this with
  `git diff -U0` before committing `9fa9bc40b`.
- `go test -count=1 -v -run 'TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles' ./internal/template/`
  → exit 0, with exactly two `--- PASS` lines (`TestManifestHashFormat`,
  `TestCatalogHashCoversSkillSubfiles`).
- `go test -count=1 ./internal/template/` → exit 1. The sole failure is `TestGTDCanonicalSurfaceGolden`
  (`templates/.claude/commands/moai/todo.md is not a thin gtd compatibility path`).

**Pre-existing, card-independent failure.** `TestGTDCanonicalSurfaceGolden` reads only 5 template
files plus a skill count.
- Scoped to those 5 files and the test file, `git diff --name-only ca2dae9a6..HEAD` → 0 bytes.
- The same range without a pathspec → 6 paths, which is the non-vacuity control.
- The test also failed with the pre-M3 template doc swapped back (`run/m4-three-tests-premis3-control.txt`).

## §E.4 Sync-phase Audit-Ready Signal

Authored by manager-docs, 2026-09-18, card t783. Tree: worktree `.claude/worktrees/t783`,
branch `WT-syncgate-hook`, run-phase code-final HEAD `23e8cd61e`.

```yaml
sync_status: audit-ready
sync_complete_at: 2026-09-18
sync_commit_sha: e25e6440d8174ac69ec335ef267b06de6dfa71a5
```

**CHANGELOG entry.** `CHANGELOG.md` `[Unreleased]` section, grouped with the sibling
sync-phase-quality-gate entries (card t602/t603/t664), inserted immediately before the t603
entry. Pre-emission self-test: `grep -c 'SPEC-SYNC-GATE-VERDICT-001' CHANGELOG.md` → 0 before
this commit (no duplicate risk from a parallel BATCH-SYNC session). AC count cross-check:
9 distinct `AC-SGV-[0-9]{3}` identifiers in `acceptance.md` (AC-SGV-001..009), all live (none
marked `[RETIRED]`/`[REF]`); the CHANGELOG entry's narrative covers all three findings (H01,
H03, SX-R05) that those 9 ACs verify. Every file path cited in the entry
(`.claude/hooks/moai/sync-phase-quality-gate.sh`,
`.claude/skills/moai/workflows/sync/quality-gates-quality.md`, their template mirrors,
`internal/template/catalog.yaml`, `.moai/reports/t783/run/`) verified to exist via `ls` before
commit.

**gate-sync-1 (pre-sync quality) / gate-sync-2 (doc scope) disposition.** Both HUMAN GATEs are
satisfied by the operator's 2026-09-18 Implementation Kickoff Approval, which explicitly
authorized autonomous progression through run AND sync without intermediate stops (progress.md
§ Implementation Kickoff Approval, relayed by the kanban lead). The sync scope is narrow and
doc/hook-verification-only (4 files: hook pair + doc pair, no user-facing command or flag
changed) — no README/docs-site edit is warranted, consistent with the operator's approval scope.
No AskUserQuestion round was run for this sync close; none was required given the recorded
approval.

**MX tag validation disposition: skipped.** `git diff --name-only "$CARD_BASE"..HEAD` (where
`CARD_BASE=$(git merge-base develop HEAD)`) shows 0 changed Go/source files — this SPEC's diff
touches only 4 SPEC artifacts, 2 doc-workflow markdown files, and (via the catalog-hash cascade)
`internal/template/catalog.yaml`, a generated hash file. No `@MX:*` annotation surface exists in
this diff to validate.

**Residuals carried forward from run-phase (§E.2/§E.3), restated for the close record:**
- `TestGTDCanonicalSurfaceGolden` in `internal/template` — pre-existing, card-independent
  failure (0-byte diff on its 5 input files + test file since `CARD_BASE`; also failed with the
  pre-M3 template doc swapped back).
- Arm-C side finding (not an AC failure): on a fixture where `go build ./...` leaves an
  untracked binary, the work-tree identity changes after a passing check and the next call
  re-runs the checks once more instead of suppressing re-delivery; stdout stayed empty, no stale
  block re-delivered. Left unrepaired — recomputing the identity after the checks would loosen
  the gate. Candidate for a follow-up card.
- No run against an installed `moai` binary (`make build` explicitly out of scope, REQ-SGV-008).
- No remote CI result yet — this repository batch-pushes `develop` from the kanban lead
  (`CLAUDE.local.md` §4.1); this sync commit stays local pending that push.

**Status transition.** `spec.md` frontmatter `status: in-progress → completed` (passing through
`implemented` in this same commit per the 3-phase close) and `updated: 2026-09-18`. No SPEC body
content (spec.md/plan.md/acceptance.md §A-§H) was modified — frontmatter only.

## §F Phase 4 Mode Selection

Logged by the lane orchestrator (lane-4) before the first run-phase Agent() spawn.

Input parameters: tier M · scope 4 target files (+ 4 SPEC artifacts) · domains 3 (hook script pair, workflow doc pair, SPEC artifacts) · language mix shell + markdown · concurrency benefit LOW (milestone-ordered: M1 freezes the fixture M2 consumes; M2's verdict gates its own repair; M3/M4 verify surfaces M2 may touch) · Agent Teams prereqs not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | not selected | multi-file edit + execution evidence — not the trivial case |
| serial | **selected** | coding/doc-edit work per Anthropic's coding-task parallelism caveat; sequential milestone dependencies; single-writer file surfaces |
| fanout | not selected | concurrency benefit LOW; write surfaces overlap the verification milestones |
| sweep | not selected | 4 files, non-uniform semantic edits — not the mechanical-uniform case |

Decision: serial

Justification: this is a verification-and-docs SPEC whose milestones are sequentially dependent (M1 fixture → M2 arms → conditional repair → M3 doc alignment → M4 close). One manager-develop spawn carrying M1→M4 in order is the simple mode that satisfies every dependency; no higher-concurrency mode meets its own entry criteria.

