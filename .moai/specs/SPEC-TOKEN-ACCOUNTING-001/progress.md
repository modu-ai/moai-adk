# Progress — SPEC-TOKEN-ACCOUNTING-001

> §E 섹션과 `§E.N` 하위 헤딩은 파서 load-bearing(`internal/spec/era.go`)이다. 헤딩 개명 금지.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-07
- tier: M
- artifacts: spec.md, plan.md, acceptance.md (+ progress.md skeleton)
- owner: manager-spec
- note: Token-Economy Epic 1/4. plan-auditor 게이트 대기.

## §E.2 Run-phase Evidence

- M1 (transcript 파서) 완료 — `internal/tokenusage` 신규 패키지 (`tokenusage.go` + `tokenusage_test.go`).
- API: `SumSession(path string) (Usage, error)` + `CacheHitRatio(input, cacheCreation, cacheRead int) float64` + `Usage` struct (tokens_input/output/cache_creation/cache_read/spent + cache_hit_ratio).
- AC PASS (shared checkout 독립 재검증): AC-TA-001(sum→1860) / AC-TA-002(usage부재→0) / AC-TA-003(malformed skip, no-panic) / AC-TA-004(ratio 경계+0분모) / AC-TA-005(read-only 불변) — `go test ./internal/tokenusage/...` 전부 PASS.
- coverage: 90.0% of statements. race clean (`go test -race`). golangci-lint 0 issues. gofmt clean. go vet 0.
- build: host exit 0, `GOOS=windows GOARCH=amd64` exit 0.
- 읽기 소스: transcript JSONL `message.usage` 4필드 합산 (statusline 현재-점유 스냅샷과 별개, spec.md §A.2). `~/.claude/projects/**` read-only.
- 회수 경위: 런타임이 manager-develop을 L1 격리 worktree(base f7b55e637)에 배치 → 코드를 shared checkout로 회수 후 독립 재검증. M2(귀속)/M3(§I writer)/M4(audit) 미착수 — 별도 spawn.
- @MX: `SumSession` ANCHOR (공개 추출 계약), `CacheHitRatio` NOTE (0-분모 규칙).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소유>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소유. sync_commit_sha 는 sync commit 시 기록>_

## §F Phase 0.95 Mode Selection

**Input parameters**:
- tier: M
- scope (file count): ~5-8 (신규 `internal/tokenusage` 2-3 파일 + `internal/spec/audit.go` / `internal/cli/spec_audit.go` 확장 + Section Map SSOT 1편집)
- domain count: 1 (internal Go 런타임/도구)
- file language mix: Go (+ markdown 1 doc)
- concurrency benefit: LOW (coding-heavy, 순차 의존 M1→M2→M3→M4)
- Agent Teams prereqs: 미충족 (harness level ≠ thorough, team env 미설정)

**Mode evaluation**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | 다중 파일 신규 코드 + 테스트 |
| 2 background | no | Write/Edit 수행 (read-only 아님) |
| 3 agent-team | no | 단일 도메인 + Agent Teams prereqs 미충족 |
| 4 parallel | no | coding-heavy (Anthropic parallelism caveat) |
| 5 sub-agent | **YES** | coding-heavy 단일 도메인 순차 의존 → 기본값 |
| 6 workflow | no | <30 파일 + mechanical 변환 아님 (신규 코드) |

**Decision: sub-agent** (Mode 5), milestone별 순차 실행 (M1→M2→M3→M4).

**Justification**: coding-heavy 단일 도메인 순차 의존 체인이므로 Anthropic coding-task parallelism caveat에 따라 Mode 5(순차 sub-agent)가 정답. milestone별 분할 실행으로 (a) 관측된 session-limit blast radius를 milestone 단위로 제한, (b) 병렬 세션(SPEC-MOAI-AGENTIC-LOOP-001 active) concurrent-commit race의 checkpoint를 milestone 간 확보.

## §G IGGDA Kickoff Predicate

- (a) intent clarity 100%: PASS — Token-Economy Epic 4-SPEC 구조 + SPEC-A Kickoff이 이전 세션 AskUserQuestion으로 확정, 본 세션 재확인.
- (b) plan-auditor PASS: PASS — PASS-WITH-DEBT 0.84 ≥ 0.80 (Tier M), D1+D2 해소 완료.
- (c) Tier S/M: PASS — Tier M.
- (d) no dangerous keywords / destructive scope: FAIL — "token" 키워드가 §H.3 security list 매칭(LLM-token false-positive) → explicit-gate. `--pr` 없음, 비파괴적 scope.
- **Verdict: explicit-gate** (조건 d FAIL) — 필수 blocking AskUserQuestion 발화됨, 사용자 승인(Option A: D1+D2 수정 후 run). timestamp: 2026-07-08.

## §I Token Accounting

_<pending sync-phase — 본 SPEC이 도입하는 신규 섹션. sync-close 시 token-accounting
메커니즘이 아래 필드를 채운다. era.go가 grep하지 않는 신규 top-level letter(§I)이므로
§E.N 파서와 무충돌. placeholder only — 값 미기록.>_

<!--
제안 필드 스키마 (run-phase에서 확정, sync-close 시 채움):
- tokens_spent: <int 합산>
- tokens_input: <int>
- tokens_output: <int>
- tokens_cache_creation: <int>
- tokens_cache_read: <int>
- cache_hit_ratio: <float [0,1]>
- token_attribution: session-set
- token_attribution_confidence: high | low
- token_session_count: <int>
-->
