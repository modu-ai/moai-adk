# SPEC-HANDOFF-GOALFIX-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_phase:
  status: audit-ready
  date: 2026-07-08
  artifacts: [spec.md, plan.md, acceptance.md, progress.md]
  tier: M
  baseline_head: 3d35cc18d
  baselines_measured: true   # '# /goal' 6/6/1/1, 're-set' 4/1/3 (live=template), FANOUT debt phrase 0
  spec_id_self_check: "decomposition: SPEC ✓ | HANDOFF ✓ | GOALFIX ✓ | 001 ✓ → PASS"
  plan_audit_iter1:
    verdict: PASS-WITH-DEBT
    score: 0.87
    mp: "4/4"
    fixes_applied:
      - "D1: AC-GF-004a false baseline corrected — grep -ci 'remind' SH baseline is 3 (remind ⊂ reminder, ceremonial-reminder prose), swapped to distinguishing token 'reminder obligation' (verified baseline 0); REQ-GF-003 now mandates the literal token"
      - "D2: REQ-GF-003 trigger re-grounded — detection via handoff auto-memory entry (resume + follow-up block persisted verbatim) or emission-condition re-derivation; acceptance.md Scenario 3 aligned"
      - "D3: 4 AC grep patterns (006a/007a/012a/012b) converted from backslash-escaped backticks to plain backticks in single quotes; darwin baselines reproduced 1/0/1/0; 006a converted to line-count form (pipe-free)"
      - "D4: related_specs frontmatter field dropped (non-canonical) — lineage retained in spec.md §A.4 + §E body prose"
      - "D6: AC-GF-003e strengthened from vacuous ≥1 to ≥3 (re-measured baseline 3)"
    scope_addition: "REQ-GF-009 goal-first bootstrap variant (user-approved Option A) + AC-GF-013a..c; tokens goal-first bootstrap / model discretion verified baseline 0"
    artifact_version: "0.1.1"
```

## §E.2 Run-phase Evidence

### Worktree reconciliation note (for orchestrator)

manager-develop executed in an isolated Claude-native worktree (`.claude/worktrees/agent-a6d58a0ffd99a64af`, branch `worktree-agent-a6d58a0ffd99a64af`). At entry the worktree pointed at `68bbcb2c4` (= `origin/main`) and lacked the SPEC's 3 unpushed plan-phase commits (`70d768c90`, `af082c2fb`, `f2f6b33bf`), which existed only on the shared checkout's local `main`. Recovery: `git rebase f2f6b33bf` fast-forwarded the worktree branch to the shared main HEAD (0 unique commits to replay, clean tree), bringing in the SPEC artifacts while leaving the 6 doctrine surfaces byte-identical (verified: the 3 plan commits touch none of the 6 surfaces — empty `git diff --stat 68bbcb2c4 f2f6b33bf` over the 6 paths). All baselines re-measured post-rebase in the worktree and matched the SPEC exactly. **Shared-checkout note**: `progress.md §F/§G` were orchestrator-authored UNCOMMITTED content in the shared checkout at session start; they are absent from `f2f6b33bf` (this worktree) so this evidence commit does NOT contain §F/§G — the orchestrator must preserve/re-apply its uncommitted §F/§G when reconciling the shared checkout against the pushed origin/main. The shared checkout also carries unrelated uncommitted files (`.moai/config/sections/system.yaml`, `SPEC-INTERNAL-PERF-001/progress.md`, `internal/statusline/*`, `pkg/version/version.go`) that were NOT touched by this run.

### AC matrix (baseline-delta; every check re-measured in the worktree post-rebase)

| AC | Command | Baseline | Actual Output | Status |
|----|---------|----------|---------------|--------|
| AC-GF-001a/b | `grep -o '# /goal' SH` / `T/SH` \| wc -l | 6 / 6 | `0` / `0` | PASS |
| AC-GF-001c/d | `grep -o '# /goal' OM` / `T/OM` \| wc -l | 1 / 1 | `0` / `0` | PASS |
| AC-GF-001e | `grep -o '# /goal' GD T/GD` \| wc -l | 0 | `0` (guard held) | PASS |
| AC-GF-002a-d | `grep -o 're-set' SH/OM/GD` (+ T/) \| wc -l | 4/1/3 | `0`/`0`/`0` (+ mirrors `0`) | PASS |
| AC-GF-003a | `grep -c '^## Post-Paste /goal Follow-up Block' SH` (+T/SH) | 0 | `1` / `1` | PASS |
| AC-GF-003b | `grep -c '^/goal <completion-condition>$' SH` | 0 | `1` | PASS |
| AC-GF-003c | `grep -c 'standalone message' SH` | 0 | `6` (≥2) | PASS |
| AC-GF-003d | structural: `/goal` skeleton between its own cut-line markers, OUTSIDE+AFTER main block; instruction line outside markers | n/a | PASS by inspection (§ Post-Paste /goal Follow-up Block skeleton) | PASS |
| AC-GF-003e | `grep -c 'machine-verifiable end-state' SH` | 3 | `7` (≥3) | PASS |
| AC-GF-004a | `grep -c 'reminder obligation' SH` | 0 | `4` (≥1) | PASS |
| AC-GF-004b | `grep -ci 'remind' GD` | 0 | `1` (≥1) | PASS |
| AC-GF-004c | reminder prose: NL status guidance NOT AskUserQuestion; timing post-Kickoff-Approval; rationale = model cannot invoke `/goal` | n/a | PASS by inspection (SH § reminder obligation + GD § Resumed-session reminder obligation) | PASS |
| AC-GF-005a | `grep -c '^## Paste-Time Activation Matrix' SH` (+T/SH) | 0 | `1` / `1` | PASS |
| AC-GF-005b | `grep -c 'code.claude.com/docs/en/interactive-mode' SH` | 0 | `2` (≥1) | PASS |
| AC-GF-005c | `grep -c 'code.claude.com/docs/en/goal' SH` | 0 | `3` (≥1) | PASS |
| AC-GF-005d | matrix 4-class; class (d) = `/goal`/`/effort`/`/clear` user-only; class (c) = `mode:`+Block5 `/moai` orchestrator-routed | n/a | PASS by inspection (4-class table present) | PASS |
| AC-GF-006a | `grep -c 'Block 1 `/goal` line' GD` | 1 | `0` | PASS |
| AC-GF-006b | `grep -c 'follow-up block' GD` | 0 | `2` (≥1) | PASS |
| AC-GF-006c | `grep -c 'Implementation Kickoff Approval' GD` | 3 | `4` (≥ baseline) | PASS |
| AC-GF-007a | `grep -c 'bare `ultracode` keyword or fan-out steering phrase' SH` (+T/SH) | 0 | `1` / `1` | PASS |
| AC-GF-008a | `grep -c 'Post-paste /goal instruction line' SH` (+T/SH) | 0 | `1` / `1` | PASS |
| AC-GF-008b | row carries 4 locale columns (en/ko/ja/zh); `/goal` verbatim (untranslated) in all 4 | n/a | PASS by inspection (Localization Table instruction-line row) | PASS |
| AC-GF-008c | cut-line MARKER rows unchanged (2) | 2 | `2` | PASS |
| AC-GF-009a | `grep -c 'follow-up block' OM` | 0 | `2` (≥1) | PASS |
| AC-GF-009b | `grep -c '…paste-ready budget… — 10 items' SH` = 1 AND `grep -c 'Pre-emit self-check (12 items)' OM` = 1 | 1/1 | `1` / `1` (SH checkbox count 10) | PASS |
| AC-GF-009c | `grep -c 'paste-ready budget' SH OM` each ≥1 | ≥1 | SH `5` / OM `1` | PASS |
| AC-GF-010a/b/c | `diff -q` SH↔T/SH, OM↔T/OM, GD↔T/GD | identical | all identical (exit 0) | PASS |
| AC-GF-010d | `make build` | exit 0 | `make build exit=0` (`/tmp/goalfix-build.log`) | PASS |
| AC-GF-011a/b | `grep -rn 'SPEC-HANDOFF-GOALFIX' / 'SPEC-V3R6-HANDOFF-GOAL-BINDING' internal/template/templates/` \| wc -l | 0 / 0 | `0` / `0` | PASS |
| AC-GF-012a | `grep -c 'Omitting the `/goal` re-set line' SH` | 1 | `0` | PASS |
| AC-GF-012b | `grep -c 'Omitting the post-paste `/goal` follow-up block' SH` | 0 | `1` | PASS |
| AC-GF-012c | `grep -c 'input start' SH` | 0 | `7` (≥1) | PASS |
| AC-GF-013a | `grep -c 'goal-first bootstrap' SH` (+T/SH) | 0 | `2` / `2` (≥1) | PASS |
| AC-GF-013b | `grep -c 'model discretion' SH` (+T/SH) | 0 | `1` / `1` (≥1) | PASS |
| AC-GF-013c | variant prose states all 5 clauses (two-step DEFAULT; effort keywords not documented to fire; model discretion; compact condition; IKA unaffected + `/goal` locale-verbatim) | n/a | PASS by inspection (§ Goal-first bootstrap variant (a)/(b)/(c)) | PASS |

### Gaps (not observed)
- Real runtime behavior of a pasted `/goal` follow-up block firing the goal loop is NOT machine-testable at the doctrine layer (documentation-only fix); ja/zh Localization-row naturalization quality is human-review residual.
- `spec-lint` / markdown lint on the edited surfaces not run (markdown surfaces are not lint-gated per acceptance.md §D.2); no Go source changed → no `go test` / `golangci-lint` delta expected.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_phase:
  run_complete_at: 2026-07-08
  run_commit_sha: a2af8a23af34e4f49560bb32c2e9f3bbd28b0d4d   # M1 (6 surfaces + draft->in-progress); this §E.2/§E.3 evidence commit is a separate chore commit
  run_status: audit-ready
  ac_pass_count: 13     # AC-GF-001..013 all PASS (all sub-items green; by-inspection items 003d/004c/005d/008b/013c confirmed)
  ac_fail_count: 0
  preserve_list_post_run_count: unchanged   # only the 6 declared doctrine surfaces + 3 SPEC frontmatter transitions touched in the worktree; no out-of-scope files modified
  l44_pre_commit_fetch: "git fetch origin main → origin/main 68bbcb2c4 at session start; worktree rebased onto shared f2f6b33bf (see §E.2 reconciliation note)"
  l44_post_push_fetch: "0 0"   # git push origin HEAD:main → 68bbcb2c4..d10c1bed8; post-push git rev-list --count --left-right origin/main...HEAD = 0 0 (synced)
  new_warnings_or_lints_introduced: 0   # doctrine-text-only; no Go source changed; make build exit 0
  cross_platform_build:
    go_build_all: n/a   # no internal/**/*.go changed; make build (re-embed) exit 0 is the build gate for this SPEC
    goos_windows_amd64: n/a   # no syscall/build-tag surface touched
  total_run_phase_files: 9   # 3 live doctrine + 3 template mirrors + spec.md + plan.md + acceptance.md (progress.md committed separately as evidence)
  m1_to_mN_commit_strategy: "2 commits — Commit A (M1) a2af8a23a: 6 surfaces + frontmatter; Commit B (chore): progress.md §E.2/§E.3 evidence. Both pushed to main (Route A Hybrid Trunk, Tier M). NO --no-verify, NO --amend, NO force-push."
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_phase:
  sync_complete_at: 2026-07-08
  sync_status: audit-ready
  sync_auditor_verdict: PASS
  sync_auditor_scores:
    overall: 0.94
    functionality: 0.96
    security: 0.95
    craft: 0.90
    consistency: 0.92
  residual_risk: "F1 debt recorded (goal-first bootstrap section coexistence of 'Implementation Kickoff Approval unaffected' with 'setting a goal starts a turn immediately' — coherent only under implicit assumption goal-first presumes run-phase resume already past Kickoff-Approval; future one-sentence clarification deferred, documented for audit trail; no functional impact to current doctrine)"
  changelog_entry_confirmed: true
  changelog_path: "CHANGELOG.md [Unreleased] / ### Added / SPEC-HANDOFF-GOALFIX-001 entry"
  sync_commit_sha: pending backfill via follow-up chore(SPEC-HANDOFF-GOALFIX-001): record sync_commit_sha commit
```

## §F Phase 0.95 Mode Selection

- tier: M / scope: 6 surfaces (3 live doctrine + 3 template mirrors) + make build / domain count: 1 (handoff doctrine) / language mix: 100% markdown / concurrency benefit: LOW (parity-coupled edits) / Agent Teams prereqs: not all met (harness standard)
- Mode evaluation: trivial=no (multi-surface semantic doctrine) / background=no (Write) / agent-team=no (1 domain < thresholds) / parallel=no (SSOT→render→mirror parity dependency, edits not independent) / **sub-agent=SELECTED** / workflow=no (6 surfaces ≪ ~30, not mechanical-uniform)
- Decision: sub-agent (Mode 5)
- Justification: doc-editing-heavy with cross-surface parity dependency (session-handoff.md SSOT → moai.md §8 render → 3 mirrors must carry consistent text); Anthropic coding-task parallelism caveat applies. Single manager-develop delegation is the safe default for Tier M doctrine work.

## §G IGGDA Kickoff Predicate

- (a) intent clarity 100%: PASS — user confirmed scope (SPEC 정식 진행) + goal-first variant (Option A) + reaffirmed "필요한 경우 goal-first bootstrap 처리" (2026-07-08)
- (b) plan-auditor PASS: PASS-WITH-DEBT 0.87 ≥ Tier M 0.80; D1-D4/D6 remediated + REQ-GF-009 added at f2f6b33bf
- (c) Tier S or M: PASS — Tier M
- (d) dangerous keywords / destructive scope: PASS — none matched; no --pr flag
- Verdict: auto-proceed — Implementation Kickoff Approval AskUserQuestion STILL ISSUED (blocking); user selected "run-phase 진입 (권장)" (2026-07-08)
