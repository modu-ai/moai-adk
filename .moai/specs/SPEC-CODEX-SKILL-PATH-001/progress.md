# progress.md — SPEC-CODEX-SKILL-PATH-001

Card t468 · Tier S · worktree `.claude/worktrees/t468` (branch `WT-codex-path-parser`, base `d592b0551`).

## §E.1 Plan-phase Audit-Ready Signal

2026-09-03 plan-audit iter-2: D1-D4 해소 확인 + D6(AC-CSP-001-09를 -04(`~user`)/-08(bare `~`)로 접어 8-AC 상한 복원, fixture 행·REQ 열거는 유지)·D7(문서 카운트 정합) 적용.

2026-09-03 plan-audit iter-1 (PASS-WITH-DEBT 0.875) 터치업 적용: D1(REQ-CSP-004 `filepath.IsAbs` 우선 한정자 + AC-04 이중-OS 예외 형태로 교체 + plan §E windows 테스트실행 판정/fixture 절대경로 규칙), D2(AC-06 실제 missing + symlink-loop 쌍으로 재작성 — no-finding 자세 보존), D3(테스트 수 census 항목), D4(AC-CSP-001-09 `~`/`~user` + M1 fixture 행).

plan_status: audit-ready
plan_complete_at: 2026-09-03
plan_tree_sha: d592b0551
artifacts: spec.md (7 REQ / 8 AC, GEARS + Given-When-Then inline), plan.md (M1-M3 single-pass), progress.md (this file)
status: draft (initial `(none) → draft`, manager-spec)

Plan-phase evidence captured on tree `d592b0551`:

- RED-now probes (throwaway /tmp program, verbatim):
  - `probe1 stat-tilde err=stat ~/definitely-not-a-real-path-t468: no such file or directory isNotExist=true`
  - `probe2 stat-backslash err=stat C:\\Users\\goos\\SKILL.md: no such file or directory isNotExist=true`
- Green baseline: `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1 -v` → both PASS (this run, this tree).
- Plan-phase correction recorded in spec.md HISTORY: t451's error-taxonomy constraint is ALREADY implemented in `codexStaleSkillFinding` — carried as PRESERVE (REQ-CSP-005), not as new work.

## §E.2 Run-phase Evidence

Milestones: M1 RED `9e8e19946` · M2 GREEN `fa72e39a8` · M3 doc-sweep `aa6c97580` (branch `WT-codex-path-parser`, cycle_type=tdd).

### E8 — RED evidence (captured BEFORE implementation, tree `9e8e19946`)

Command: `go test ./internal/cli/ -run 'CodexSkillPath' -count=1 -v`
Census (D3): `go test ./internal/cli/ -run 'CodexSkillPath' -count=1 -v 2>&1 | grep -c '^=== RUN'` → **8** (positive; no `[no tests to run]`).
Verdict: **5 FAIL / 3 PASS** — the 3 PASS are the pre-classified guards (AC-02 vacuous-green-today, AC-05, AC-06), matching spec.md §D.1 two-cell table.

Verbatim failure excerpts (the right stated reason in every case — raw `os.Stat` ENOENT counting the shape as missing, remove directive attached):

- AC-01 `TestCodexSkillPath_HomeRelativeExistingNotMissing` — `an existing ~-relative registration was reported missing: {Status:warn Message:~/.codex/config.toml: 1 stale skill entry … — remove the stale entries or restore the skill files}`
- AC-03 `TestCodexSkillPath_RelativeNotMissingDistinctClassification` — `relative declaration inflated the missing count: "~/.codex/config.toml: 2 stale skill entries"` (+ 2 further errors: missing count lost; distinct relative classification absent from Detail)
- AC-03 `TestCodexSkillPath_RelativeOnlyNonDestructiveFinding` — `a relative-only config was advised to remove entries: {Status:warn Message:… 1 stale skill entry …}`
- AC-04 `TestCodexSkillPath_BackslashAndOtherUserHomeNotMissing` — `oddly-formed declarations inflated the missing count: "~/.codex/config.toml: 3 stale skill entries"`
- AC-08 `TestCodexSkillPath_ExpansionUsesUserHomeSeam` — `~ expansion did not use the pinned user home seam: {Status:warn Message:… 2 stale skill entries …}`

### E1 — AC matrix (all measurements this run, tree `aa6c97580` unless noted)

Command (shared): `go test ./internal/cli/ -run 'CodexSkillPath' -count=1 -v` → 8/8 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 1.308s`.

| AC | Test | Status | Command (selector) | Actual output | Tree |
|---|---|---|---|---|---|
| AC-CSP-001-01 | HomeRelativeExistingNotMissing | PASS | `-run 'CodexSkillPath_HomeRelativeExistingNotMissing'` | `--- PASS … (0.01s)` | aa6c97580 |
| AC-CSP-001-02 | HomeRelativeMissingStillCounted | PASS | `-run 'CodexSkillPath_HomeRelativeMissingStillCounted'` | `--- PASS … (0.01s)` | aa6c97580 |
| AC-CSP-001-03 | RelativeNotMissingDistinctClassification + RelativeOnlyNonDestructiveFinding | PASS | `-run 'CodexSkillPath_Relative'` | both `--- PASS` | aa6c97580 |
| AC-CSP-001-04 | BackslashAndOtherUserHomeNotMissing | PASS | `-run 'CodexSkillPath_BackslashAndOtherUserHomeNotMissing'` | `--- PASS … (0.01s)` | aa6c97580 |
| AC-CSP-001-05 | AbsoluteExistingAndMissing + StaleHomeSkillsReported + HealthyHomeSkillsNoFinding | PASS | `-run 'TestCodexSkillPath_AbsoluteExistingAndMissing|TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding'` | all `--- PASS` / `ok` | aa6c97580 |
| AC-CSP-001-06 | RealMissingWithSymlinkLoopIndeterminate + IndeterminateStatNotMissing | PASS | `-run 'CodexSkillPath_RealMissingWithSymlinkLoopIndeterminate|TestCheckCodexWiring_IndeterminateStatNotMissing'` | all `--- PASS` | aa6c97580 |
| AC-CSP-001-07 | structural — source scan of the new fixture set | PASS | inspection: every fixture path is `t.TempDir()`-derived or a declared-shape literal (`~/…`, `skills/…`, `skills\…`, `~otheruser/…`, `~`); home pinned via `stubCodexHome` (seam + CODEX_HOME blank) in all 8 tests; zero reads of the developer home, zero writes outside temp roots | n/a (structural) | aa6c97580 |
| AC-CSP-001-08 | ExpansionUsesUserHomeSeam | PASS | `-run 'CodexSkillPath_ExpansionUsesUserHomeSeam'` | `--- PASS … (0.00s)` | aa6c97580 |

Fixture sweep (lead-mandated): `grep -n '/nonexistent' internal/cli/doctor_codex_test.go` → no output (exit 1). All 12 former literal sites now derive from `absentSkillPath(t, …)` / shared `absentRoot` temp roots. Width-band tests (MessageWidthStaysInBand, RenderedPanelStaysInBand) PASS unchanged with the longer temp paths — their band assertions bound the rendered Message/panel width, which cites config paths and counts, not the fixture skill paths.

Regression: `go test ./internal/cli/ -run 'TestCheckCodexWiring' -count=1` → `ok … 0.799s` (19 tests). `go test ./internal/codexwiring/ -count=1` → `ok`.

### E2 — Cross-platform builds (tree `aa6c97580`)

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- Windows TEST-RUN verdict: deferred to CI (named Gap — a build verdict proves compilation only, not classification).

### E3 — Coverage (informational vs 85%)

- `go test -cover ./internal/codexwiring/ -count=1` → `ok … coverage: 88.2% of statements` (tree `aa6c97580`)
- `go test -cover -timeout=25m ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 608.802s coverage: 80.5% of statements` (exit 0, tree `aa6c97580`). The full package suite PASSES under coverage — an earlier attempt FAILED only on the default 10m test timeout (601.349s), the package's known long-suite floor, not on any test. 80.5% is the pre-existing package-wide figure (the touched code's behavior is covered by the 27-test selector set, all green); the 85% package target remains the quality.yaml SSOT concern for the package as a whole, not a regression introduced by this SPEC.

### E4 — Subagent boundary

`grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/doctor_codex.go internal/cli/doctor_codex_test.go internal/codexwiring/skills.go` → no matches (exit 1).

### E5 — Lint

- Pre-flight baseline (tree `f813cdb0e`): `golangci-lint run --timeout=2m` → 1 pre-existing errcheck in `internal/template/catalog_tree_hash.go` (untouched by this SPEC).
- Post-change (tree `aa6c97580`): `golangci-lint run --timeout=5m ./internal/cli/... ./internal/codexwiring/...` → `0 issues.` — NEW findings: 0.

### E6 — Commits + push state

- `9e8e19946` M1 (tests + spec frontmatter draft→in-progress) · `fa72e39a8` M2 (classification + expansion) · `aa6c97580` M3 (SkillEntry.Path doc posture) · progress/backfill commits follow this file.
- Push state: **not pushed — lane protocol** (develop integration + push is the lead's window; the lane never pushes).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: "f943364f0"   # run-phase closing commit (backfilled per D3 placeholder pattern)
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 7
preserve_list_rows:
  - "CODEX_HOME precedence via resolveCodexHomeDir (CodexHomeHonoured + CodexHomeConfigRead PASS)"
  - "stat-error taxonomy: only fs.ErrNotExist feeds missing (AC-06 pair PASS)"
  - "indeterminate-only config stays silent — no-finding-when-missing==0 posture preserved (HealthyHomeSkillsNoFinding, AbsentHomeConfigSilent PASS)"
  - "no second home resolver — ~ expansion reuses the codexUserHomeDir seam (AC-08 PASS)"
  - "read-only, fail-open in every direction (unresolvable user home counts indeterminate, never missing)"
  - "unspecified-enabled declared split preserved (UnspecifiedEnabledReportedSeparately PASS)"
  - "Message/rendered-panel width bands preserved (MessageWidthStaysInBand + RenderedPanelStaysInBand PASS)"
l44_pre_commit_fetch: "not-run — lane worktree off develop; no origin-facing claim made in run-phase"
l44_post_push_fetch: "not-run — no push (lane protocol: lead batch-pushes origin/develop)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: "exit 0 (go build ./…, tree aa6c97580)"
  windows: "exit 0 (GOOS=windows GOARCH=amd64 go build ./…, tree aa6c97580)"
  windows_test_run: "deferred-to-CI (named gap; build proves compilation only)"
total_run_phase_files: 5   # doctor_codex.go, doctor_codex_test.go, skills.go (doc comment), spec.md frontmatter, progress.md
m1_to_mN_commit_strategy: "per-milestone commits: M1 9e8e19946, M2 fa72e39a8, M3 aa6c97580, + progress/backfill"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-03
sync_commit_sha: "135e325f5"   # backfilled in the immediately following commit (D3 self-reference workaround)
sync_status: complete
b12_self_test_a_pre_emission_grep: "grep -c 'SPEC-CODEX-SKILL-PATH-001' CHANGELOG.md → 0 before append (duplicate guard PASS)"
b12_self_test_b_ac_count_match: "8 AC identifiers in spec.md §D (Tier S inline SSOT; no acceptance.md by design) == 8 referenced in the CHANGELOG entry"
b12_self_test_c_file_path_verification: "ls verified: internal/cli/doctor_codex.go, internal/cli/doctor_codex_test.go, internal/codexwiring/skills.go — all present"
changelog_entry_position: "CHANGELOG.md [Unreleased] → Fixed (topmost entry)"
frontmatter_status_transitions:
  - "spec.md: in-progress → completed on the single sync commit (3-phase close merged; spec.md frontmatter status + updated only, no body edits)"
user_facing_docs_judgment: "NO README/docs-site change — the delta is an advisory-behavior refinement inside `moai doctor`'s stale-skill check; the doctor's user-facing contract ('reports stale entries') is unchanged in wording, only its false-positive behavior on non-absolute path shapes changed. docs-site documents doctor at the command level, not the per-finding advisory level."
canary_compliance_check: "n/a — SPEC defines no forward-looking canary policy"
```

Backfill record: commit `docs(SPEC-CODEX-SKILL-PATH-001): backfill sync_commit_sha=135e325f5 (t468)` replaces the placeholder above with the sync commit's real SHA (`135e325f5`).

## §F Phase 4 Mode Selection

2026-09-03 run-phase entry (lane dispatch: Kickoff 승인 통과, lead-1). Phase 1 스킵 근거: plan-audit iter-2 PASS 0.9375 (Tier S 문턱 0.75 상회) + 산출물 해시 불변(plan 커밋 8650286b8, 트리 clean) + 판정 PASS — 3조건 전부 충족.

Input parameters: tier S · scope 3 code files (`internal/cli/doctor_codex.go`, `internal/cli/doctor_codex_test.go`, `internal/codexwiring/skills.go` 주석 1건) · domains 1 (Go backend, internal/cli) · file language mix 100% Go · concurrency benefit LOW (coding-heavy) · Agent Teams 미요청 (--team 없음).

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | 의미 변경 수반(분류 로직 + 테스트) — typo/단일행 아님 |
| serial | **yes** | coding-heavy 단일 도메인; Anthropic coding-task parallelism caveat; M1(RED)→M2(GREEN)→M3 마일스톤 체인이 본질적으로 순차 |
| fanout | no | coding-heavy — 병렬화 주의권 고지 단일 도메인 |
| sweep | no | 3파일, 기계적-균일 변환이 아닌 의미 작업; ~30파일 미달 |

Decision: serial (단일 manager-develop spawn, cycle_type=tdd)

Justification: Tier S coding-heavy SPEC이 단일 도메인에 머물고 마일스톤 체인이 엄격히 순차(RED fixture → GREEN 분류기 → 보고 스윕)라서 manager-develop 1회 위임이 §E 귀속 행렬과 함께 전 체인을 운반한다. Boundary case 없음 — 결정트리 기본 분기.

## Sync-audit fold record (2026-09-03, lane-8 orchestrator)

- Sync-audit verdict: **PASS-WITH-DEBT 91/100** (harmonic mean; Func 95 / Sec 95 / Craft 82 / Cons 93) — `.moai/reports/t468/sync-verdict.md` (감사자 본문, 무수정 보존). 독립 확인: 27/27 PASS 0 SKIP, vet/gofmt/lint clean, AC 8/8, PRESERVE 7/7, stress points (a)-(e) 결함 없음.
- **D1 [MINOR][blocking] 해소** — CHANGELOG.md:330의 Windows-호스트 분류 절이 규범과 정반대 서술. 수정 커밋 `2a3615b34` (manager-docs 재위임, 1파일 1행). orchestrator 델타 검증: diff-stat 1파일 + 절 재독 — REQ-CSP-004/`IsAbs`-우선 구현과 일치. Windows 런타임 처분은 여전히 CI 판정 대기(§E.3 기존 갭).
- D2/D3 [INFO][optional] 수용 부채 — 감사자 자신이 Tier S 충분으로 판정. 직접 단위테스트 부재(분류기)/indeterminate-only 침묵 셀.
- **프로세스 이벤트 기록 (일인-작성자 규율)** — run-phase E3 증거 커밋 `6e972f585`가 3-phase close(`135e325f5`) 이후이자 sync-audit 창 도중에 착지. 문서 전용(코드 0건)이라 감사자가 코드-동일성 diff로 스스로 브리지 — 판정 기반 불변. 순서 수용/접을지는 리드 판정 사항으로 완료 보고에 운반됨.
