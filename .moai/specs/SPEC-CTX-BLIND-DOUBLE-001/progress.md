# 진행 기록 — SPEC-CTX-BLIND-DOUBLE-001

카드 t539 · Tier M · Class C · 워크트리 `.claude/worktrees/t539` · 브랜치 `WT-ctx-blind-double` · 기준 `52f863f36`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-08
tier: M
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - research.md
  - progress.md
requirements: 16   # REQ-CBD-001..016 (Tier M 상한 16)
acceptance_criteria: 15   # AC-CBD-001..015 (Tier M 상한 16)
needs_clarification: 0   # 해소됨 2026-09-08 — 운영자 결정 "측정만 이 카드, 수리는 후속 카드"
research_source: .moai/reports/t539/sweep.md
baseline_head: 52f863f36
spec_version: "0.2.0"
plan_audit_iter1: FAIL 0.84 (MP-7 해소 게이트) — D0~D12 반영 완료
```

plan 이전 스윕은 완료돼 있다. 뮤턴트 두 건(M1 검출 / M2 생존)이 각각 제외 근거와 수리 근거의
실측 사례로 고정됐다.

plan-audit 1회차(FAIL 0.84)의 must-pass 실패였던 해소 게이트 마커 1건은 **해소됐다** —
운영자 결정 2026-09-08, "측정만 이 카드, 수리는 후속 카드"(`plan.md` §F "수리의 소속 — 해소됨").
해소 게이트 마커는 산출물 전체에서 0건이다. 나머지 지적 D1~D10 · D12 는 산출물에 반영했고,
D11(감사 보고서 경로 관례)은 오케스트레이터 재량으로 남겼다.

## §E.2 Run-phase Evidence

> 증거 정본은 `.moai/reports/t539/verdict.md`(5절 형식). 이 표는 AC 별 판정만 옮긴다.
> 모든 `-run` 셀렉터 인용에는 같은 셀렉터의 `-list` 건수가 붙는다(REQ-CBD-013).

### M1 — 판별식 확정 + 스윕 산출물 고정

| AC | 분류 | 검증 명령 | 관측 출력 | 판정 |
|---|---|---|---|---|
| AC-CBD-001 | regression-guard | `/usr/bin/grep -n '판별식' research.md` | `92:## 4. 판별식 — 네 축 (SPEC 확정본)` + `101:**수리 대상 = (c) ∧ (d).**` + `99:` (d) 축 근거 `codex_task.go:124` | 깨지지 않음 (통과로 기록하지 않음) |
| AC-CBD-002 | regression-guard | `go run .moai/reports/t539/ctxsweep/main.go internal` → `cmp` + `awk` 재집계 | `cmp_exit=0`(바이트 동일) · method 127/26/29 · funclit 196/52/12 · `/usr/bin/grep -rl ctxsweep internal/` → exit 1(무출력) | 깨지지 않음 |
| AC-CBD-003 | regression-guard | `/usr/bin/grep -n '옳게 눈멂\|teeGLMDoer' research.md` | `89:` goal 러너 참조 0건 · `182/183:` goal fakeRunner + hook 핸들러 · `57:` teeGLMDoer 위임 예외 | 깨지지 않음 |

M1 로그: `.moai/reports/t539/m1-rerun.log`. 코드 변경 0줄.

### M2 — GLM audit 경로 취소 가드

| AC | 검증 명령 | 관측 출력 | 판정 |
|---|---|---|---|
| AC-CBD-004 | 가드 없이 `mcp_glm.go:305` 뮤턴트 주입 → `-run GLM` / `-run Audit` / `-run Converg` | `ok 11.610s` · `ok 31.061s` · `ok 1.314s` (**생존**) · `-list` 223/99/32 · 되돌림 후 `git diff --stat` 무출력 | PASS |
| AC-CBD-005 | 가드 추가 후 같은 뮤턴트 재주입 | `--- FAIL: TestGLMAudit_CancelledContext_IsNotSwallowed (0.45s)` — `a cancelled audit returned the canned success verdict "pass"` (**사망**) · 넓은 셀렉터 `-run GLM`/`-run Audit` 도 FAIL · 되돌림 후 초록 | PASS |
| AC-CBD-006 | `go test ./internal/cli/ -run TestGLMTask -count=1 -timeout 1800s` | `ok 1.276s` · `-list` 16 · `glm_task_bg_context_test.go` diff 0줄 | PASS |
| AC-CBD-007 | 세 셀렉터 재실행 + `-list` 로그 보존 | `ok 8.837s` / `ok 20.907s` / `ok 0.980s` · `-list` 225/101/32 (기준선 223/99/32 대비 **감소 없음**, +2/+2/+0) · `list-{GLM,Audit,Converg}.log` 실재 · `mcp_glm_test.go` diff 0줄(`stubGLMDoer` 눈먼 채 유지) | PASS |

부가 측정: `-run GLM -race` → `ok 11.937s` · `go vet ./internal/cli/...` → exit 0 · `golangci-lint ./internal/cli/...` → `0 issues.`

M2 로그: `.moai/reports/t539/m2-mutant.log`. **프로덕션 코드 변경 0줄** — `:305` 은 처음부터
`NewRequestWithContext` 였다. 없던 것은 가드다. 신규 파일은 `internal/cli/mcp_glm_audit_ctx_test.go` 하나.

### M3 — 다섯 후보군 84건 뮤턴트 판정 (측정만, 수리 없음)

씨앗 11개, 후보군 5개. 판정: **생존 49 · 검출 34 · seam 부재 1 = 84**.

| 씨앗 | 주입 지점 | `-list` | 관측 출력 | 판정 |
|---|---|---:|---|---|
| 1A | `update/orchestrator.go:55` | 88 | `ok 6.334s` | 생존 |
| 1B | (씨앗 없음 — `IsUpdateAvailable(string)` 에 ctx 파라미터 부재) | — | 인터페이스 시그니처 | seam 부재 |
| 2 | `statusline/builder.go:362` | 318 | `ok 23.423s` | 생존 |
| 3a | `lsp/aggregator/aggregator.go:174` | 19 | `FAIL 0.495s` — `got 1 diagnostics, want 0` 외 1건 | **검출** |
| 3b | `lsp/core/manager.go:366` | 97 | `ok 0.305s` | 생존 |
| 3c | `lsp/transport/request.go:53` | 32 | `FAIL 180.201s` — `TestCallWithTimeout_DeadlineExceeded` 행(hang) | **검출** |
| 4a | `core/project/initializer.go:434` | 118 | `ok 1.609s` | 생존 |
| 4b | `cli/mirror_notice.go:37,:42` | 161·32·51 | `ok 15.564s` / `ok 4.384s` / `ok 0.321s` | 생존 |
| 5a | `cli/branch_protection.go:65` | 5 | `ok 0.862s` | 생존 |
| 5b | `github/pr_reviewer.go:110` | 136 | `ok 0.244s` | 생존 |
| 5c | `guardstate/evaluate.go:180` | 47 | `ok 0.228s` | 생존 |

| AC | 검증 명령 | 관측 출력 | 판정 |
|---|---|---|---|
| AC-CBD-008 | `/usr/bin/grep -E '^- \`internal/[^\`]+_test\.go:[0-9]+\` \`' plan.md \| wc -l` + 귀속표 대조 | `84` · 미측정 **0건** · 로그 5개 + `m3-attribution.md` 실재 | PASS |
| AC-CBD-009 | 생존 후보 수리 diff 부재 + 후속 카드 요청 | 프로덕션 diff 0줄 · `verdict.md` §7 에 후속 카드 요청 6항목(F1~F6) | PASS |
| AC-CBD-010 | 검출 후보 수리 diff 부재 + 판정 기록 | 3a·3c CLOSE "옳게 눈멂 / 위층이 컨텍스트를 본다" · `m3-lsp.log` | PASS |

M3 로그: `m3-update.log` · `m3-statusline.log` · `m3-lsp.log` · `m3-template-deployer.log` ·
`m3-gh-client.log` · 귀속표 `m3-attribution.md`.

**귀속의 한계(과장하지 않는다)**: 씨앗은 11개이고 나머지 구성원은 같은 seam 뒤에 있다는 이유로
귀속됐다. "84건 전부에 판정이 귀속된다"는 참이지만 "84건 전부가 자기 뮤턴트를 가졌다"는 거짓이다
(`m3-attribution.md` §Gaps, `verdict.md` §4).

### 전역

| AC | 검증 명령 | 관측 출력 | 판정 |
|---|---|---|---|
| AC-CBD-011 | `verdict.md` 의 모든 `-run` 인용에 `-list` 건수 동반 | 인용 전부에 건수 기재 · 0건 셀렉터 근거 0건 | PASS |
| AC-CBD-012 | `git diff --stat` / `git status --short` | 둘 다 무출력 · 커밋 어디에도 뮤턴트 없음 | PASS |
| AC-CBD-013 | 부재 주장의 근거 명령 | 전부 `/usr/bin/grep` 절대경로 · 셸 맨 `grep` 근거 0건 | PASS |
| AC-CBD-014 | 채택 근거 형태 | M2 = 생존→수리→사망 삼단 · 커버리지 단독 채택 0건 | PASS |
| AC-CBD-015 | `git log --format=%s dbc1f7125..HEAD` | 커밋 제목 전부 `t539` 포함 · `verdict.md` 실재 · 인용 로그 전부 실재 | PASS |

**불변식 — 프로덕션 코드 변경 0줄**:
`git diff --stat dbc1f7125..HEAD -- 'internal/***.go' ':(exclude)internal/**/*_test.go'` → 무출력.
전체 변경 파일 19개 중 Go 파일은 `internal/cli/mcp_glm_audit_ctx_test.go` 하나(테스트 전용).

빌드/린트/vet: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 ·
`go vet` (건드린 7개 패키지) exit 0 · `golangci-lint run` → `0 issues.` (기준선도 `0 issues.` → NEW 0건).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: pending-backfill-run   # M3 커밋은 자기 SHA 를 참조할 수 없다 (D3 backfill 창)
run_status: complete
ac_pass_count: 12        # AC-CBD-004..015
ac_fail_count: 0
ac_regression_guard_count: 3   # AC-CBD-001..003 — 통과로 기록하지 않음 (undecidable disposition)
preserve_list_post_run_count: 0   # PRESERVE 목록 위반 0
l44_pre_commit_fetch: n/a         # WT 브랜치는 push 하지 않는다 (레인 규율) — fetch/divergence 판정 불필요
l44_post_push_fetch: n/a          # push 없음
new_warnings_or_lints_introduced: 0   # lint-baseline.log "0 issues." → lint-final.log "0 issues."
cross_platform_build:
  host: pass       # go build ./... exit 0
  windows: pass    # GOOS=windows GOARCH=amd64 go build ./... exit 0
  note: 크로스 빌드는 테스트 파일을 컴파일하지 않으므로 이 카드의 산출물(테스트 전용)에 대해서는
        회귀 부재의 방증일 뿐 직접 근거가 아니다
total_run_phase_files: 19        # 보고서 17 + progress.md + spec.md ... Go 파일은 1개(테스트 전용)
production_code_lines_changed: 0 # 불변식 — git diff 로 확인
m1_to_mN_commit_strategy: milestone 당 1커밋 (M1 / M2 / M3), 전 제목에 t539, push 없음
evidence_root: .moai/reports/t539/
verdict: .moai/reports/t539/verdict.md   # 5절 형식 (Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)
followup_card_requests: 6        # F1~F6 (verdict.md §7) — 생존 뮤턴트는 수리하지 않고 후속 카드로
known_gaps: 7                    # verdict.md §4 — 최대 항목: 84건 개별 뮤턴트 미측정(씨앗 11개로 귀속)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

- plan_complete_at: 2026-09-07T17:34:14Z
- plan_status: audit-ready
- plan_audit: iteration 2/2 · 0.91 · must-pass 7/7 · E1/E2 literal fixes applied by orchestrator (verified by grep)
- kickoff_approval: 2026-09-08 operator — approved, autonomous progression

## Phase 4 Mode Selection

- date: 2026-09-08 · tier: M · harness: standard · development_mode: tdd
- mode: serial (sub-agent) — single manager-develop spawn, milestones M1→M2→M3 in order; mutants require an exclusive writer per tree, which rules out parallel write-capable agents
- progression: autonomous (operator, kickoff gate) — carried by orchestrator continuation; `/moai goal` NOT armed (background-agent waits would spin idle turns against the stop-goal evaluator; arm-only hazard per goal-directive.md)
- plan audit gate: plan-audit.md iteration 2/2 · 0.91 · must-pass 7/7 (read, not trusted: E1/E2 literal fixes verified by grep in this session)
