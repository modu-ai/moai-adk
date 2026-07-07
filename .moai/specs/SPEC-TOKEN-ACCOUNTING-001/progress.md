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
- M2 (귀속 레이어) 완료 — `Attribution` struct + `Attribute(progressPath, transcriptRootDir, specsDir, activeSessionUUID)` 공개 진입점 신규 추가. session-set 합산(REQ-TA-005), lineage 부재 폴백(REQ-TA-006), high/low 신뢰도 한정자(REQ-TA-007) 구현. M1 `SumSession`+`finalize` 재사용(중복 금지). 공유-UUID 교차검사(specsDir 스캔).
- AC PASS: AC-TA-006(session-set+high, count=2) / AC-TA-007(lineage 부재→low 폴백, count=1) / edge §D.2(공유 UUID→low 강등) — `go test -run TestAttributionConfidence` PASS. REQ-TA-013(absent transcript skip-and-continue) / extractSessionUUIDs 6 subcase PASS.
- coverage: 90.1% (M1 90.0% 유지). race clean. golangci-lint 0 issues. gofmt clean. build host/windows exit 0.
- @MX: `Attribute` ANCHOR (M3/M4 다운스트림 통합 지점), `determineConfidence` NOTE (신뢰도 휴리스틱).
- M3 (§I writer + Section Map SSOT + era 무충돌 회귀테스트) 완료 — TDD(cycle_type=tdd). 신규 파일: `internal/tokenusage/section_writer.go`(`BuildSectionI`/`WriteSectionI`/`SectionIHeading` + `applySectionI` 멱등 replace-or-append) + `section_writer_test.go`(7 subcase). `internal/spec/era_token_section_test.go`(`TestEraUnchangedByTokenSection`, AC-TA-009). Section Map SSOT(`spec-frontmatter-schema.md`)에 `## §I Token Accounting` 행 추가.
- AC PASS: AC-TA-008(`## §I Token Accounting` + `tokens_spent:` grep hit, fixture writer 실행 후) / AC-TA-009(`go test ./internal/spec/ -run TestEraUnchangedByTokenSection` PASS — §I 추가 전후 ClassifyEra == V3R6 H-4 불변) / AC-TA-011(era.go/audit.go `§E.N`+SHA 토큰 미개명 source guard — diff 공백). M1/M2 회귀 없음(`go test ./internal/tokenusage/... ./internal/spec/...` — 본 SPEC 추가 테스트만 신규, 기존 전부 PASS).
- §I 실측 값 미기록(sync-phase 소관) — 본 마일스톤은 writer/스키마/회귀테스트만. §I placeholder는 그대로.
- coverage: tokenusage 패키지 M1+M2+M3 합산 (M3 코드 신규 — `go test -cover`로 E3에서 실측). race clean. golangci-lint 0 issues(NEW). build host/windows exit 0.
- @MX: `WriteSectionI` ANCHOR (sync-close 통합 지점), `BuildSectionI` NOTE (필드-스키마 1:1 매핑).
- M4 (audit 표면) 완료 — `AuditResult.TokensSpent *int` (JSON `tokens_spent,omitempty`) 신규 필드 + `parseTokensSpentFromSectionI` 파서. M3 `BuildSectionI` 출력 포맷(`- tokens_spent: <int>`) round-trip 파싱. 단일 SPEC / `--filter-spec` 감사 시 노출; 다중 SPEC 감사는 nil (정밀도-정직 — 신뢰도 한정자가 다른 SPEC간 합산은 misleading). 미기록 §I → nil (fabricate 금지, REQ-TA-012).
- AC PASS: AC-TA-010(`TestSpecAuditTokensSpent` 3 subcase: §I 있음→1860 / §I 없음→nil / §I 있으나 tokens_spent 라인 없음→nil) + `TestParseTokensSpentFromSectionI` 6 boundary subcase — `go test -run 'TestSpecAuditTokensSpent|TestParseTokensSpentFromSectionI' ./internal/spec/` PASS. AC-TA-012(중립성 grep 0 files) PASS.
- coverage: 88.2% (internal/spec, M3-debt TestCloseSubjectDoctrineAmendment 제외 시 통과). internal/tokenusage 90.1% 유지. golangci-lint 0 issues. build host/windows exit 0.
- 로컬 heading 상수 `sectionIHeading` 사용 (M4 worktree가 M3 `section_writer.go` 이전 base여서 회수 당시 cross-package import 불가; coupling 주석으로 문서화 — M3 heading 변경 시 lockstep 갱신 필요, sync-phase에서 `tokenusage.SectionIHeading` 재사용으로 DRY 정리 후보). CLI human-readable summary에 `Tokens spent:` 라인 추가(nice-to-have).
- source_session_id: ae640eb5-a14e-4c13-ad86-2b7d8afc464f (run-phase M1-M4 세션 — token-attribution lineage, REQ-TA-005 session-set 합산용)

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-07-08
- milestones: M1(transcript 파서) + M2(session-set 귀속) + M3(§I writer + Section Map SSOT + era 무충돌 회귀) + M4(audit 표면 tokens_spent) 전부 완료
- ac_pass: AC-TA-001..012 (12/12 PASS, SHOULD AC-TA-011 포함) — `go test ./internal/tokenusage/... ./internal/spec/...` 실측 PASS
- coverage: internal/tokenusage 90.1%, internal/spec 88.2% (per §E.2 manager-develop E3 증거)
- build: host exit 0, `GOOS=windows GOARCH=amd64` exit 0; golangci-lint 0 issues; go vet 0; gofmt clean; `-race` clean
- note: shared-checkout 독립 재검증 + L1 격리 worktree 회수 (runtime 배치 → ff/cherry-pick)

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready (→ completed on sync commit)
- sync_complete_at: 2026-07-08
- sync_artifacts: spec.md frontmatter in-progress→completed + progress §E.3/§E.4 signals + §I 실측값 + CHANGELOG entry
- token_self_measure: §I = 30,304,046 tokens_spent / cache_hit_ratio 0.9229 / session-set / high / 2 sessions — 본 SPEC 도입 메커니즘의 첫 실측 적용 (dogfood). throwaway `cmd/selfmeasure` 헬퍼로 `Attribute()`+`WriteSectionI()` 실행 후 헬퍼 삭제.
- sync_commit_sha: f88d0226f
- residual: plan-phase 세션 UUID 미회수 (측정 누락分); CLI write 경로 미연결 (현재 library-only — 후속 SPEC에서 `moai` CLI §I write 연동)

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

- tokens_spent: 30304046
- tokens_input: 2301120
- tokens_output: 474670
- tokens_cache_creation: 0
- tokens_cache_read: 27528256
- cache_hit_ratio: 0.9229
- token_attribution: session-set
- token_attribution_confidence: high
- token_session_count: 2
