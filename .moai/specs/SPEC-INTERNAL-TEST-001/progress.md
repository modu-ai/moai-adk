---
id: SPEC-INTERNAL-TEST-001
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-TEST-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-08: plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 저작 완료 (manager-spec, `status: draft`). 발견사항 F1-F5의 파일·라인 앵커는 저작 시점 HEAD에서 read-only 재확인됨. plan-auditor 감사 대기.

## §E.2 Run-phase Evidence

### AC-TEST-002a (M1, F1) — PASS
```
$ go test -run TestRunHookEvent_ReadInputError -count=1 ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	0.583s
```

### AC-TEST-002b (M1, F1) — PASS (SIGSEGV 제거, 커버리지 측정 가능)
```
$ go test -cover -count=1 ./internal/cli/
coverage: 71.6% of statements
```
(이전: SIGSEGV로 커버리지 측정 불능. 이제 측정 가능. 6개 Doctor/Status TUI 사전 부채 FAIL 잔존 — 본 SPEC scope 외.)

### AC-TEST-003a (M3, F2) — PASS
```
$ go test -run TestAuthoringDocHasEffortMatrix -count=1 ./internal/cli/agentlint/
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.434s
```

### AC-TEST-003b (M3, F2) — PASS (0)
```
$ awk '/expectedAgents := \[\]string\{/,/\}/' internal/cli/agentlint/agent_lint_test.go | grep -cE 'expert-|researcher|manager-strategy|manager-quality|manager-project|manager-cycle|builder-platform'
0
```

### AC-TEST-003c (M3, F2) — PASS (template ↔ local parity, doc 미편집)
```
$ diff internal/template/templates/.claude/rules/moai/development/agent-authoring.md .claude/rules/moai/development/agent-authoring.md
exit=0 (identical)
```

### AC-TEST-004a (M2, F3) — PASS
```
$ go test -run TestCloseSubjectDoctrineAmendment -count=1 ./internal/spec/
ok  	github.com/modu-ai/moai-adk/internal/spec	0.442s
```

### AC-TEST-004b (M2, C-2) — PASS (템플릿 미러 미생성)
```
$ ls internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md
No such file or directory
```

### AC-TEST-005a (M4, F4) — PASS (coverage ≥ 85%)
```
$ go test -cover -count=1 ./internal/migration/
ok  	github.com/modu-ai/moai-adk/internal/migration	0.453s	coverage: 85.5% of statements
```

### AC-TEST-005b (M4, F4) — PASS (0 skip placeholders)
```
$ grep -rn 't.Skip("waiting for migration package implementation")' internal/migration/ | wc -l
0
```

### AC-TEST-006a (M5, F5) — PASS (파일 존재)
```
$ ls internal/constitution/canary_test.go internal/constitution/contradiction_test.go
(internal/constitution/canary_test.go, internal/constitution/contradiction_test.go 존재)
```

### AC-TEST-006b (M5, F5) — PASS-WITH-DEBT (ok, coverage 67.5% < 85%)
```
$ go test -cover -count=1 ./internal/constitution/
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.427s	coverage: 67.5% of statements
```
F5 scope(canary.go 87% / contradiction.go 95%) 완료. package 41.3%→67.5% 상승.
85% 미달 사유: pipeline.go(8함수 0%, integration-level) + human_oversight.go 일부가
F5 scope 외 사전 부채. plan.md §G 예견: "도달 불가 시 실측 수치 + 사유 문서화".

### AC-TEST-001 (M6, headline) — NOT MET (0 panic ✓, 7 pre-existing FAIL)
```
$ go test ./internal/... -count=1
(7 FAIL: 6× internal/cli Doctor/Status TUI + 1× internal/statusline env-flake)
```
0 panic. 7 FAIL 전부 사전 부채:
- 6× internal/cli TestDoctor/TestStatus TUI 렌더링 (M1 SIGSEGV 수정 전까지 마스킹됨)
- 1× internal/statusline TestBuild_WritesContextUsageWithSessionID (context_window_size
  환경 의존, SPEC Out of Scope + project_ac_css_001_rescan_debt.md 기록 pre-existing)

### Cross-platform build (B1) — PASS
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./internal/migration/ → exit 0
```

### Lint (E5) — PASS (0 issues, NEW vs baseline 구분)
```
$ golangci-lint run --timeout=2m
0 issues.
```

### Subagent boundary (E4, B3) — PASS (my edits added 0 AskUserQuestion to production code)
```
$ git diff main --name-only | grep -E "\.go$" | grep -v "_test.go"
(empty — production .go 무접촉)
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-07-08"
run_commit_sha: "84019cbf1"
run_status: "PASS-WITH-DEBT"
ac_pass_count: 11
ac_fail_count: 1   # AC-TEST-001 headline (7 pre-existing FAIL, F1-F5 scope 외)
preserve_list_post_run_count: 5   # hook.go + migration/*.go + constitution/{canary,contradiction}.go — 전부 무변경
l44_pre_commit_fetch: true
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues (수정 후)
cross_platform_build:
  darwin: "exit 0"
  windows: "exit 0 (./internal/migration/)"
total_run_phase_files: 8   # coverage_test.go + lifecycle-sync-gate.md + agent_lint_test.go + 4×migration/*_test.go + 2×constitution/*_test.go
m1_to_mN_commit_strategy: "per-milestone (M1=0fce3e256 M2=629d8bd0d M3=140d95a48 M4=2abc66f09 M5=2files M5lint=84019cbf1)"
residual_debt:
  - "AC-TEST-001 headline: 6× internal/cli Doctor/Status TUI FAIL (SIGSEGV 수정으로 노출된 사전 부채)"
  - "AC-TEST-001 headline: 1× internal/statusline env-flake FAIL (pre-existing, SPEC Out of Scope)"
  - "AC-TEST-006b: internal/constitution package coverage 67.5% < 85% (pipeline.go 0% integration-level, F5 scope 외)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 0.95 Mode Selection

- 입력: tier=S(선언)/M(실측, D1 권장), files≈11, domains=5+(cli/spec/migration/constitution/rules+template), language-mix=Go test+markdown, concurrency-benefit=LOW(coding-heavy)
- 평가: Mode 1 trivial X | Mode 2 background X(write 작업) | Mode 3 agent-team X(prereqs 미확보) | Mode 4 parallel X(coding-heavy caveat) | Mode 5 sub-agent **선택** | Mode 6 workflow X(<30 files, multi-rule)
- Decision: sub-agent (Mode 5)
- 근거: Anthropic coding-task parallelism caveat — 테스트 저작 + doc 정합은 coding-heavy로 순차 sub-agent가 안전. domain이 다수라도 coding-heavy면 Mode 4/6 우회하고 Mode 5 기본.

## §G IGGDA Kickoff Predicate

- (a) intent clarity 100%: PASS (resume 메시지 + AskUserQuestion 승인)
- (b) plan-auditor PASS: PASS (PASS-WITH-DEBT 0.91, skip-eligible, F1-F5 전부 CONFIRMED at HEAD)
- (c) Tier S or M: PASS (tier S 선언 / M 실측 — 양 쪽 모두 조건 만족)
- (d) no dangerous keywords: FAIL ("migration" 키워드 매치 — §H.3 critical-infra 목록 over-inclusive; internal/migration 패키지 대상이나 database migration 아님에도 매치)
- verdict: explicit-gate (조건 d FAIL) → mandatory blocking AskUserQuestion 발행 → 사용자 승인 획득 ("승인 — 즉시 run 진입 (권장)")
- timestamp: 2026-07-08 / source_session_id: fe5490b1
