# SPEC-STATUSLINE-OFFICIAL-USAGE-001 — progress

plan→run Kickoff (autonomous, auto-semantics §9.1)이 아래 차감을 **구속 조건**으로 복사함 (2026-10-14T23:30Z, lane-35/t1732, run tmvgvn). 두 sync 판정 소유자가 재독하며, 미처리 조건은 sync 판정을 FAIL로 상한합니다. 판정 출처: `.moai/reports/t1732/plan-audit-iter2.md` (PASS-WITH-DEBT 0.87, must_pass_failed 0, blocking 0, audited_sha a52a0e5c9, plan_artifact_hash e960c02a…837fe).

## Binding run conditions

> 갱신 (board d-20261014T235405Z-8b55 조건 ②, 2026-10-15T00:14Z): 아래 두 차감은 `f87f31dfd`에서 plan.md M4 스코프로 흡수되어 더 이상 run 구속 조건이 아닙니다. 처분 위치는 다음 AC이며, sync-audit이 이 AC들로 처분을 확인합니다.

- ~~sou-debt-1~~ → **AC-SOU-010** (acceptance.md §D; REQ-SOU-006): mid-read abandonment 핀 — plan.md §C M4 (blocked/slow fixture, 읽기 시작 후 취소) + §B D8 (가드 위치).
- ~~sou-debt-2~~ → **AC-SOU-009** (acceptance.md §D, release-blocking; REQ-SOU-003): relative `gitdir:` 포인터 케이스 — plan.md §C M4 (relative-pointer fixture) + §B D7 (Clean+Join 해소).

원 기록 (2026-10-14T23:30Z Kickoff가 복사, 이력 보존): sou-debt-1/2 원문은 plan-audit-iter2.md 10-11행 및 git 이력(17c3e8f72의 progress.md) 참조.

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_complete_at: 2026-10-14T23:43:52Z
plan_audit: PASS-WITH-DEBT 0.87 (iteration 2) — .moai/reports/t1732/plan-audit-iter2.md

## §F Phase 4 Mode Selection

- Input: tier M · scope ~10 files (usage.go·builder.go·renderer.go·types·tests·CLI comment·template verify) · domains 1-2 (statusline pkg + template mirror) · file mix Go+markdown · concurrency benefit LOW (coding-heavy)
- Mode evaluation: direct — not trivial, no / serial — **selected** / fanout — no (coding-heavy, Anthropic caveat) / sweep — no (not ≥30-file mechanical uniform)
- Decision: serial
- Justification: coding-heavy single-package implementation with a TDD cycle — per Anthropic's coding-task parallelism caveat the sequential path is the safe default; one writer per tree (factory lane discipline); milestones M1-M5 sequence through one manager-develop delegation with incremental commits. Kickoff gate met in its autonomous form (§9.1: PASS-WITH-DEBT 0.87, must_pass_failed 0, blocking 0, hash-bound, decision record on the card progress).

## §G Override and Refusal Record

- 2026-10-14T23:45:39Z SPEC-STATUSLINE-OFFICIAL-USAGE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit.md
- 2026-10-15T01:43:21Z SPEC-STATUSLINE-OFFICIAL-USAGE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit.md
- 2026-10-15T01:48:58Z SPEC-STATUSLINE-OFFICIAL-USAGE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit.md

## §E.2 Run-phase Evidence

Run executed 2026-10-15, tree `70df3df13` → `65e04f664` (branch `WT-10-14-class`, card worktree `.moai/worktrees/t1732`, TDD cycle_type). Commits: `180cecfd5` (M1 removal), `65e04f664` (M2+M3 fallback + n/a). Not pushed — the lane's factory_complete owns landing (B9).

### RED evidence (captured before GREEN, verbatim excerpts)

1. **Structural guard RED** (AC-SOU-002/006) — `go test ./internal/statusline/ -run 'TestRemovedCollectionSurfacesAbsent|TestRetiredUsageSymbolsAbsent' -count=1 -v` on the pre-removal tree:

   ```
   --- FAIL: TestRemovedCollectionSurfacesAbsent (0.35s)
       usage_guard_test.go:120: removed usage-collection surface(s) reappeared in production files:
           usage.go:23 matches api.anthropic.com
           usage.go:26 matches api.anthropic.com
           usage.go:29 matches oauth-2025-04-20
           usage.go:32 matches haiku-4-5
           usage.go:358 matches Claude Code-credentials
           usage.go:358 matches credentials.json (either credential file)
           ... (16 hits, all usage.go)
   --- FAIL: TestRetiredUsageSymbolsAbsent (1.07s)
       usage.go:78 matches NewUsageCollector
       ... (36 hits, all usage.go)
   ```

2. **E-ledger baseline re-observation** (acceptance §E cells, re-run on `70df3df13`, code-identical to the ledger's `0d057bd1e`): E1 → `internal/statusline/usage.go`, exit 0 · E2 → no output, exit 1 · E4 → 7 lines incl. `406: if b.usageProvider != nil && (input == nil || input.RateLimits == nil) {`, exit 0 · E5 → no output, exit 1 · E6 → 6 lines usage.go:114,118,357,362,402,404, exit 0 · E7 → `43: if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {`, exit 0 · E8 → 6 lines all usage.go, exit 0 · E9 → `77: if !filepath.IsAbs(gitFile) {`, exit 0.

3. **Behavioral RED** (7 tests, pre-M2) — `go test ./internal/statusline/ -run 'TestBuild_QuotaFallback' -count=1 -v` exit 1:

   ```
   --- FAIL: TestBuild_QuotaFallbackAbandonsMidRead (0.53s)
       usage_fallback_blocked_test.go:78: abandoned fallback should render n/a, got segment "🔋 5H: ░░░░░░░░░░ 0%"
   --- FAIL: TestBuild_QuotaFallbackNAWithoutRecord (0.25s)
       usage_fallback_test.go:135: 5H segment should render n/a with no snapshot record, got "🔋 5H: ░░░░░░░░░░ 0%"
   --- FAIL: TestBuild_QuotaFallbackFreshRecord (0.70s)
       usage_fallback_test.go:167: 5H segment should render the fresh record's 42%, got "🔋 5H: ░░░░░░░░░░ 0%"
   --- FAIL: TestBuild_QuotaFallbackResetRollover (0.67s)
   --- FAIL: TestBuild_QuotaFallbackWorktreeAbsoluteGitdir (0.13s)
   --- FAIL: TestBuild_QuotaFallbackWorktreeRelativeGitdir (0.81s)
   --- FAIL: TestBuild_QuotaFallbackDegradedMalformedGitdir (0.81s)
   --- PASS: TestBuild_QuotaFallbackNotReadWhenRateLimitsPresent (0.44s)   ← AC-SOU-005 preservation pre-change
   ```

   Every RED is red for the stated reason (0% bars where n/a/values are required); the mid-read test fails fast pre-M2 (no read exists) and its 3s deadline assertion is what an entry-check-only guard would trip post-M2.

### GREEN evidence (this run, HEAD `65e04f664`)

| # | Claim | Command | Verbatim observed output |
|---|-------|---------|--------------------------|
| 1 | Behavioral ACs flip green | `go test ./internal/statusline/ -run 'TestBuild_QuotaFallback\|TestRenderQuotaSegment\|TestFormatCapturedAge' -count=1 -v` | 10/10 PASS (Abandons 1.51s, NAWithoutRecord 1.17s, FreshRecord 1.03s, ResetRollover 2.53s, WorktreeAbsolute 1.71s, WorktreeRelative 1.78s, DegradedMalformed 1.21s, NotReadWhenRateLimits 1.12s, RenderQuotaSegment, FormatCapturedAge) → `ok ... 13.125s`, exit 0 |
| 2 | Full package suite | `go test ./internal/statusline/ -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/statusline	431.261s`, exit 0 |
| 3 | Race variant (new paths) | `go test ./internal/statusline/ -race -run 'Quota\|Snapshot\|Abandons' -count=2` | `ok ... 20.983s`, exit 0 |
| 4 | Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | `WIN_OK`, exit 0 (plain `go build ./...` also exit 0) |
| 5 | Structural guards | `-run 'TestRemovedCollectionSurfacesAbsent\|TestRetiredUsageSymbolsAbsent'` | both PASS, exit 0 |
| 6 | E1/E4/E6 green flips | same grep commands as the ledger | E1 exit 1 (no output) · E4 exit 1 · E6 exit 1 · E2 now 4 `n/a` sites in renderer.go · E5 wired (`usage_fallback.go:111 AggregateQuotaDirs`) |
| 7 | Template-First cycle (AC-SOU-007) | `make build` | exit 0; `git status --short internal/template/templates/` empty after build; template grep found only `usage_5h`/`usage_7d` segment-toggle keys (no removed-collection reference — no template change needed) |
| 8 | Coverage (AC-SOU-008) | `go test -cover ./internal/statusline/` | `ok ... 566.492s	coverage: 94.0% of statements`, exit 0 |
| 9 | Lint (E5) | `golangci-lint run --timeout=6m internal/statusline/...` | `0 issues.`, exit 0. Full-repo baseline run timed out under machine load (2m and 8m) after emitting 3 PRE-EXISTING errcheck findings in `internal/cli/factory_run_adopt.go` — unrelated package, untouched by this SPEC |
| 10 | Boundary grep (E4 §E) | `grep -rn 'AskUserQuestion' internal/statusline \| grep -v _test.go \| grep -v "// "` | empty, exit 1 |
| 11 | Push state (E6) | `git status -sb`; `git log origin/main..HEAD --oneline` | `## WT-10-14-class` (no upstream); `65e04f664`, `180cecfd5` local-only — not pushed |
| 12 | M1 suite baseline | `go test ./internal/statusline/ -count=1` at the M1 tree | `ok ... 557.295s`, exit 0 (AC-SOU-005 preservation arm) |

**Suite-stability note (recorded honestly):** the M1 suite pass 1 failed on 3 findings — two byte-equality builder tests read the LIVE board/git state (backlog `+0 M0 ?0` vs `+12 M9 ?1`, branch ahead `+12` moving between their sequential Build calls on a busy multi-lane machine) and one stale `45%/82%` assertion behind a removed collector fixture. Fix: inputs anchored at `t.TempDir()` + mocked git provider in those two tests; assertions updated. Pass 3 green (`557.295s`). The M2+M3 suite pass 1 failed only on the M1-interim `0%` assertion, which M3 intentionally flips to `n/a`; pass 2 green (`431.261s`).

### AC results matrix (E1)

| AC | Status | Verifying evidence (this run, HEAD `65e04f664`) |
|----|--------|--------------------------------------------------|
| AC-SOU-001 | PASS | usage.go deleted (E1 flip exit 1); GLM-form behavioral tests pass with no collector present; guard keeps the surfaces dead |
| AC-SOU-002 | PASS | `TestRemovedCollectionSurfacesAbsent` PASS (was RED with 16 hits); E1 green-path exit 1 |
| AC-SOU-003 | PASS | `TestBuild_QuotaFallbackNAWithoutRecord` + `ResetRollover` PASS; `TestRenderQuotaSegment` unknown/zero-value arms PASS; E2 flip (4 `n/a` sites) |
| AC-SOU-004 | PASS | `TestBuild_QuotaFallbackFreshRecord` PASS (42%/77% + `captured ` marker); `TestFormatCapturedAge` PASS |
| AC-SOU-005 | PASS | M1-tree suite green (`ok 557.295s`); RateLimits test functions unmodified; AC-SOU-005 byte-identical pin (`🔋 5H: ██░░░░░░░░ 23% (rolling)`) captured pre-change and PASSING both pre- and post-change; `TestBuild_QuotaFallbackNotReadWhenRateLimitsPresent` PASS |
| AC-SOU-006 | PASS | `TestRetiredUsageSymbolsAbsent` PASS (was RED, 36 hits); E6 flip exit 1; `mockUsageProvider`/`realisticUsage` references in builder_test.go: 0 |
| AC-SOU-007 | PASS | `make build` exit 0; embed tree clean post-build; template grep: no reference to the removed collection |
| AC-SOU-008 | PASS (SHOULD) | coverage 94.0% of statements (≥85% gate) |
| AC-SOU-009 | PASS | `TestBuild_QuotaFallbackWorktreeAbsoluteGitdir` + `WorktreeRelativeGitdir` PASS; `DegradedMalformedGitdir` PASS (fail-open, no cross-tree leak); E7/E9 baseline observed |
| AC-SOU-010 | PASS | `TestBuild_QuotaFallbackAbandonsMidRead` PASS (1.51s) — FIFO fixture blocks the read, context cancelled AFTER the read starts, Build returns promptly with n/a; entry-check mutant excluded (a pre-launch-only check leaves Build blocked → 3s deadline fails) |

### Gaps (explicitly not observed)

- golangci-lint never completed a FULL-repo run under this machine's load (2m and 8m budget timeouts; the 8m run emitted the 3 pre-existing `internal/cli/factory_run_adopt.go` errcheck findings, then hit its internal deadline, exit 4). The E5 verdict for this SPEC's scope rests on the scoped package run (`0 issues.`), which is the surface this SPEC touches.
- AC-SOU-005's byte-identical clause is judged through the pinned segment + the unmodified RateLimits test functions + the suite, not a golden-output snapshot (acceptance §F residual, unchanged).

### Residual risk

- The abandoned `os.ReadFile` (FIFO fixture; pathological records generally) finishes in its own goroutine whenever the read does — the named, accepted REQ-SOU-006 residual; the caller-visible contract (prompt return + n/a) is what the tests pin.
- The snapshot fallback renders whatever the tool's own local records carry; trust parity with the existing context bar (acceptance §F), no new surface introduced.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-15T05:20:00Z
run_commits: 180cecfd5 (M1), 65e04f664 (M2+M3)
run_audit: self-verified 10/10 AC PASS (matrix in §E.2), coverage 94.0%, package lint 0 issues, suite + race + windows green on HEAD 65e04f664; evidence per command above. Awaiting sync-phase audit (sync-auditor owns the binding verdict).

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-15T05:31:19Z
sync_commit_sha: 39489015f

**Binding-conditions disposition (sync re-read):** sou-debt-1 → **AC-SOU-010** (acceptance.md §D; REQ-SOU-006 mid-read abandonment pin) and sou-debt-2 → **AC-SOU-009** (acceptance.md §D, release-blocking; relative `gitdir:` pointer case) — both absorbed into plan.md M4 scope at `f87f31dfd` (header 갱신 note above) and **both disposed in run**: AC-SOU-010 PASS via `TestBuild_QuotaFallbackAbandonsMidRead` (context cancelled AFTER the read starts; 3s deadline excludes an entry-check-only mutant), AC-SOU-009 PASS via `TestBuild_QuotaFallbackWorktreeRelativeGitdir` + `TestBuild_QuotaFallbackDegradedMalformedGitdir` (§E.2 AC matrix, HEAD `65e04f664`). No unprocessed binding condition remains; the binding sync verdict is sync-auditor's.

**Sync scope (this commit):** spec.md frontmatter `in-progress → implemented → completed` (status + updated only; `updated:` already 2026-10-15, unchanged), this §E.4, and the CHANGELOG `[Unreleased]` entry — zero code changes (run phase owns code).

**Docs sweep (sync):** README 4-locale + docs-site statusline surfaces describe display semantics only — the 5H/7D rows ("요금제 사용률 + 초기화 시각" and the per-page rolling-window bullets) stay accurate under the new precedence (stdin `rate_limits` → local snapshot → n/a). Grep for the removed collection mechanism (haiku probe / api.anthropic usage / oauth usage / keychain / credentials) across `README*.md` + `docs-site/content/`: 0 mechanism hits — the only "Haiku" occurrences are context-window-class text (unrelated), and the only `api.*anthropic` occurrences are `api.z.ai/api/anthropic` GLM base URLs. No docs edit needed.

**Card-review:** recorded not-performed — codex 401 on both surfaces (`.moai/reports/t1732/card-review.md`); the sync-phase audit path is unchanged (sync-auditor owns the binding verdict).
