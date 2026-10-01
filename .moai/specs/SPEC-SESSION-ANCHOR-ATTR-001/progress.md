# progress.md — SPEC-SESSION-ANCHOR-ATTR-001

## 카드 연계

- **카드**: t1339 — "Bash 워크트리 세션 앵커 교차 레인 오판독 수리" (등록 2026-09-29T07:40:40Z)
- **배차**: 리드 디스패치 — t1337+t1339 동일 근원 통합 수리, 저장소 내 수리 가능분 + 재현/측정 스위치 + 회피 규율 문서(부분 산출 허가) + 레인-병렬 재현 환경 심각도 판정
- **SPEC**: SPEC-SESSION-ANCHOR-ATTR-001 (Tier M)
- **병목 연계**: autonomy-bottleneck-proposal-20260929.md bottleneck #7 → P7 (추가 자율화의 전제로 시퀀싱)

## Phase 1 SKIP 근거

리드가 계획 단계 진입 전 조사 팬아웃(plan-research-fanout 3렌즈 + 종합)을 이 레인에서 선실행하여 산출물을 /tmp에 전달했다. 본 SPEC의 research.md가 그 종합본을 축어 보존하므로, 표준 Phase 1(계획 내 재조사)은 SKIP — 중복 조사는 컨텍스트 낭비이며 1차 기록은 이미 수집·교차검증됨. 렌즈 원본 3종의 경로는 research.md 헤더 인용.

## Decision Point 1 흡수

계획-검토 후보 게이트(Decision Point 1)는 lane 도크트린에 따라 factory Implementation Kickoff 게이트에 흡수된다 — 운영자가 레인 창(pane)에서 직접 답한다. 별도의 중간 승인 라운드는 두지 않는다.

---

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-30
- artifacts: spec.md, plan.md, acceptance.md, research.md, spec-compact.md, progress.md (Tier M + research + compact)
- plan_audit: PASS 1.00 (반복 2/2, Tier M 임계 0.80) — 1차 FAIL 0.71(D1~D4) → 수정 → 델타 재감사
- plan_audit_reports: .moai/reports/plan-audit/SPEC-SESSION-ANCHOR-ATTR-001-review-1.md, -review-2.md (gitignored 로컬 런타임 기록)
- plan_audit_model: glm-5.3-flash (GLM 프록시 레인 실사용 모델, 1차는 audit_multi로 codex gpt-6.1-sol/high 수렴 병행)
- FO-PLAN-2 렌즈 생략 근거: 당일 실측 429 레이트리밋 압박(워크플로 중 1에이전트 429·t1347 쿼터 기록) + plan-auditor --deep 자체 증거 수집으로 대체 — fallback 절(single plan-auditor path) 적용
- DP2/3/3.5 흡수: 개발환경(워크트리 이미 진입)·다음 행위(run, 배차 지시)·실행 모드(serial, 레인 기본)는 factory Kickoff 게이트에서 일괄 노출
- MX 계획: plan.md는 감사 PASS 해시 고정 유지 — MX 대상(신규 exported 함수 @MX:NOTE, RelocateSession @MX:ANCHOR 후보, 미테스트 public @MX:TODO)은 run 위임 프롬프트로 전달
- run 진입 시 Phase 1 skip 예상: PASS 1.00 ≥ 0.80 + 아티팩트 해시 불변 (skip 계약 3조건 중 2개, verdict+score+hash)

## §E.2 Run-phase Evidence

### M1 — W1 거부 귀속 가능화 (2026-09-30, HEAD 78df22755 기준 작업)

- REQ-SAA-001/002 구현: `harness.Event`에 `cwd`·`worktree_path` additive omitempty 필드 추가(internal/harness/types.go — 행 스키마의 기계적 소재라 M1 봉투 내 캐스케이드), failure_observer.go에 WorktreeGuardRefusal 행 전용 attribution(session_id/cwd/트리 경로, 미해상 시 `unknown` 마커, 행 유실 없음). 다른 카테고리 행은 기존 형태 유지(REQ 스코프).
- RED 증거(E8): 신규 3테스트 컴파일 실패 — `undefined: guardRefusalWorktreePath` / `undefined: guardRefusalAttribution` / `harness.Event has no field or method Cwd` (구현 전 실측).
- GREEN: `go test ./internal/hook/ -run '^(TestGuardRefusalWorktreePath|TestClassifyError_GuardRefusal_UnknownMarkers|TestRecordToolFailureEvent_GuardRefusalRowFields|…기존 guard군)$'` — 전부 PASS, ok.
- 기존 살생 변이 2건 테스트(TestClassifyError_GuardRefusal_AnchorOnly, _BeatsOOM) 무수정 GREEN — REQ-SAA-010.
- draft→in-progress 전환(spec.md status+updated만) 본 M1 커밋에 동승.

### M2 — W2 재배치 감사 + 소유 타당성 (2026-09-30)

- REQ-SAA-003: `Registry.RelocateSession`이 cwd 재작성 시 감사 행(session_id/from/to/trigger/timestamp)을 `<project-root>/.moai/logs/anchor-relocation-audit.jsonl`에 추가 — 내부 구현은 신설 `RelocateSessionWithOptions`(anchor_relocate_audit.go)로 위임, 모든 진입점이 감사. fail-open(기록 실패가 재배치를 막지 않음).
- REQ-SAA-004: 케이스표 4팔 구현(self/other-live/unreadable/registry-only + 잔여 dead-holder 기록). lock reason card-id/pid 토큰 파싱은 `parseLockPID`/신설 `ParseLockCardID` 재사용. other-live·registry-only는 advisory 플래그 + 진행; unreadable은 기존 fail-closed 준수(플래그 없음).
- REQ-SAA-005: `workflow.anchor_relocation_guard.enabled` opt-in 키(config types.go+defaults.go — BranchGuard/AgentStopGuard 선례 동형, 기본 false, 템플릿 중립). on + 플래그 → 거부 + 감사행 Refused:true, `ErrRelocationRefused`.
- REQ-SAA-006 보존: 두-패스 후보 순서(relocateRegistryCandidatesFrom)·fail-open 무수정 — 기존 8 RUN 테스트 GREEN.
- 훅 연결: cwd_changed_relocate.go에 git-context seam(`relocationGitContext` 패키지 변수 — sessionProcessLiveness 선례 동형), `relocationTargetContext`(git rev-parse --show-toplevel + worktree list --porcelain fail-open), `anchorRelocationGuardEnabled`(nil-safe). cwdChangedHandler cfg 필드 + NewCwdChangedHandlerWithConfig, deps.go 갱신, handleWorktreeMove에 cfg 전달.
- RED 증거(E8): (a) session측 — 구현 전 컴파일 실패: `undefined: RelocationAudit/RelocationOptions/RelocateSessionWithOptions/ParseLockCardID/ErrRelocationRefused` 등 audit 표면 전체 미존재; (b) hook측 — `undefined: relocationGitContext` ×3, `too many arguments in call to relocateSessionCwd` ×3, `undefined: anchorRelocationGuardEnabled` ×3 (구현 전 실측).
- GREEN: session 신규 5테스트 + 기존 RelocateSession 3테스트, hook 신규 4테스트 + 기존 two-pass 8 RUN — 전부 PASS. session·config 패키지 전체 ok.

### M3 — W3 앵커 트레이스 스위치 (2026-10-01)

- REQ-SAA-009: `EnvAnchorTrace = "MOAI_ANCHOR_TRACE"` 상수 envkeys.go 선언(유일 리터럴). E4 grep 실측: literal 1건(envkeys.go:574) + 주석 4건 — 코드 하드코딩 0.
- REQ-SAA-007: 신설 internal/session/anchor_trace.go — `TraceAnchorDecision`이 게이트 on 시 `.moai/logs/anchor-trace.jsonl`에 행(session_id/pid/cwd/timestamp/decision/detail) 추가. 계측점 3곳: (1) `RelocateSessionWithOptions`(decision=relocate — trigger/owner/flagged/refused detail), (2) `AnchorDecision`(decision=anchor_decision — 판정 프로세스 pid+judged tree, session_id unknown 마커), (3) hook `checkBranchState` Seam A 앵커 읽기(decision=branch_guard.anchor_read — fail-open 경로도 기록).
- REQ-SAA-008: 게이트 off 시 env 조회 1회 외 오버헤드 0 — 부정 경로 테스트 + 양성 대조 병행(AC-008 요구: 같은 fixture 게이트 on 대조). truthy="1"/"true"(대소문자 무시).
- REQ-SAA-012 보존: branch_guard 트레이스 행의 로그 디렉터는 기존 audit-log projectDir(CLAUDE_PROJECT_DIR 체인) 사용, git-context는 input.CWD 체인 그대로.
- RED 증거(E8): `undefined: AnchorTraceEvent/AnchorTracePath/AnchorTraceEnabled`(session), `undefined: session.AnchorTraceEvent/AnchorTracePath`(hook) — 구현 전 실측.
- GREEN: session 신규 4테스트 + hook 신규 2테스트 전부 PASS; session 패키지 전체 ok, hook branch-guard군 ok.

### M4 — W4 회피 규율 문서 (2026-10-01)

- REQ-SAA-014 4요소를 worktree-integration-ops.md 신설 절에 문서화: (a) 복구 절차(ExitWorktree keep + 재-EnterWorktree — feedback_worktree_cd_wedge 선례), (b) 런타임 경계 선언(방출 앵커는 Claude Code 런타임 소유, 내부 keying은 unknown 기록), (c) 무음 경로-없는-명령 위험(t741 실측), (d) 정정된 t1337(재시작 없이 지속)/t1339(오트리 쓰기 0건·수 분 오판 후 복귀) 기록 + 측정 계측기 안내(MOAI_ANCHOR_TRACE + 재배치 감사 로그). 세션-종료 메모리 갱신은 후속 주석 명시.
- Template-First 준수: 템플릿 선수정 → 로컬 동기(cp) → `make build`(catalog.yaml 재생성 — 변경 0 실측).
- 검증: diff 템플릿↔로컬 = 0, `TestTemplateNeutralityAudit` ok, `TestRuleTemplateMirrorDrift` PASS, 4요소 presence grep 템플릿·로컬 양면 t1339=3/3.
- AC-010 RED-now(`grep -c t1339` = 0) → green-path 전환 완료.

### M5 — W5 심각도 일치 + 종합 검증 (2026-10-01)

- W5: spec.md §A.3 심각도 논거 3축(무음 변이 실측/처분 가드 오귀속/표준 운용 재현)과 구현 산출물 대조 — 모순 0(무음 측정은 W3 트레이스가, 처분 가드 오귀속 판독은 W2 registry-only 플래그가, 기본 동작 불변은 advisory 도크트린이 각각 대응; 부분 산출 한계는 spec.md §A.4 그대로).
- E2 빌드: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 + `go vet ./internal/...` exit 0 (HEAD 455e069f6+lfx, 최종 트리).
- E1 AC 매트릭스: AC-001..AC-010 전부 PASS — §D 매트릭스 선택자 실측(관측 RUN 수 인용, `[no tests to run]` 0건): AC-001/002 군 14 RUN ok, AC-003~005 session 6+hook 7 RUN ok, AC-006 회귀(anchor_relocate_test.go·cwd_changed_relocate_anchor_test.go **무수정** — base 대비 diff 0 실측) ok, AC-007/008 트레이스 4+2 RUN ok(부정+양성 대조 병행), AC-009 (a)살생 변이 2건 유지[파일 diff 순수 추가 실측]/(b)Seam A/[c]감사 로그 CLAUDE_PROJECT_DIR 앵커/(d)env 리터럴 1건(envkeys.go 선언)/(e)LiveAnchoredSessions 6 RUN/(f)RefuseMutation 1 RUN 전부 ok, AC-010 diff 템플릿↔로컬 0 + 중립성 ok.
- E3 커버리지: `go test ./internal/session/ -cover` → **86.1%** (ok 36.8s); `go test ./internal/hook/ -cover -timeout=24m` → **86.8%** (ok 1198.9s, TRUST 5 85% 충족).
- 전체 hook 스위트 최종 실측: -cover 풀 실행(1198.9s, 무타임아웃)에서 적색은 **선존재 2건뿐** — TestMaybeSet1MAutoCompactWindow(glm-5.2)·TestMaybeDeclareGLMContextWindow(glm-4.5-air), t1368 모델 매트릭스 유입 baseline 결함(코드경로 session_start.go·session_start_test.go·statusline/memory.go base..HEAD diff 0 입증) — 본 SPEC 불가항목, 리드 통합 전 수리 필요. 중간 실행의 타이밍 3건 적색은 격리 재실행 전부 PASS로 부하 요인 판정.
- E4 경계 grep: B3 명령 raw 24건 전수 확인 → 23건 주석 + 1건 pre_tool.go:834 `input.ToolName == "AskUserQuestion"`(관측 비교문 — user_decision_capture 서브파이프라인, API 호출 아님; 본 SPEC 미터치 파일, base 대비 diff 0). 신규 위반 0.
- E5 린트(CI판 golangci-lint v2.1.6): 변경 패키지 4종 scoped 실행 → **0 issues**. 전체 리포 실행은 병행 부하로 timeout(측정 조건 기록) — scoped 결과가 변경면 커버. 중간 실측 3건 errcheck(test `os.Unsetenv` 미확인 복귀)는 본 마일스톤에서 수정(`_ =` 명시) — 최종 신규 0.

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-10-01
- run_commit_sha: pending-backfill-run
- run_status: complete
- ac_pass_count: 10
- ac_fail_count: 0
- ac_pass_with_debt_count: 0
- preserve_list_post_run_count: 6 (anchor_lock.go 동작·two-pass 순서·Seam A 분리·worktreeGuardAnchor 핀·LiveAnchoredSessions 읽기 경로·RefuseMutationFromNonCanonicalTree — 전부 무수정 실측: base 대비 해당 파일 diff 0 또는 기존 테스트 GREEN)
- l44_pre_commit_fetch: n/a (레인 미push — 리포 로컬 git-flow, 통합은 리드 일괄)
- l44_post_push_fetch: n/a (상동)
- new_warnings_or_lints_introduced: 0 (중간 3건 errcheck — M5에서 수정, 최종 scoped lint 0 issues)
- cross_platform_build.darwin_arm64: pass (go build ./... exit 0)
- cross_platform_build.windows_amd64: pass (GOOS=windows GOARCH=amd64 exit 0)
- total_run_phase_files: 25 (base 78df22755..HEAD — Go 소스 15 + 테스트 6 + rules doc 2 + SPEC 아티팩트 2)
- f2_correction_note: 위 24 → 25 정정 — sync-audit F2: 수정된 worktree_guard_refusal_test.go가 테스트 분해에서 미계수였음 (.moai/reports/t1339/sync-audit-1.md F2, 2026-10-01 manager-docs sync 단계 적용)
- m1_to_mN_commit_strategy: per-milestone 5커밋 (M1 attribution → M2 audit+guard → M3 trace → M4 docs → M5 wrap-up)
- pre_existing_baseline_note: internal/hook GLM 컨텍스트 윈도우 테스트 2건 적색 — t1368 유입 선존재(코드경로 diff 0 입증), 본 SPEC 스코프 외, 리드 통합 전 수리 필요

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete
- sync_complete_at: 2026-10-01
- sync_commit_sha: pending-backfill-sync
- sync_audit: PASS 0.97 (harmonic mean, Tier M 임계 0.85 — must-pass Functionality 1.00·Security 1.00) — report `.moai/reports/t1339/sync-audit-1.md` (cold-auditor fallback 경로, audit 1, audited_sha 4072b84a74d28ee504b446fc584ef47fe352cc76)
- changelog_entry_position: CHANGELOG.md `[Unreleased]` § Added 선두 항목
- b12_self_test_a: 사전 배출 grep `SPEC-SESSION-ANCHOR-ATTR-001` in CHANGELOG.md = 0건 (중복 항목 가드 통과)
- b12_self_test_b: AC 카운터(acceptance.md) → live=10 excluded=0 ambiguous=0 — CHANGELOG 항목의 10건(AC-001..AC-010)과 일치
- b12_self_test_c: CHANGELOG 항목이 이름 대는 파일 경로 전수 `ls` 실측 존재 확인 (16경로, 누락 0)
- frontmatter_status_transitions.spec: in-progress → completed (본 sync 커밋에 동승, updated: 2026-10-01 — 3-phase close로 completed 전이가 sync 커밋에 병합됨)
- frontmatter_status_transitions.plan_acceptance: 해당 없음 — plan.md·acceptance.md는 frontmatter 미보유(status 축 stateless 규정)
- mx_folded: 3-phase close에 따라 MX 태그 검증은 별도 Mx-phase가 아니라 sync 하위 단계로 흡수 — sync-audit Craft 0.95가 신규 @MX:NOTE 어노테이션(anchor_relocate_audit.go·anchor_trace.go) 포함 표면을 실측 커버
- sync_audit_f2_correction: §E.3 total_run_phase_files 24 → 25 정정 (상세는 §E.3 f2_correction_note)

## §F Phase 4 Mode Selection

- 입력 파라미터: tier M · scope 7~9파일 (internal/hook 3 + internal/session 2 + docs 1 + 테스트) · 도메인 수 2 (hook/session + rules doc) · 언어 혼합 Go 위주 · concurrency benefit LOW (coding-heavy)
- 모드 평가: direct 미선정(다중 파일·의미 변경) / fanout 미선정(coding-heavy — Anthropic 병렬화 주의사항) / sweep 미선정(기계적 대량 변형 아님) / **serial 선정**
- Decision: serial (단일 manager-develop 순차 위탁, per-milestone 커밋)
- 근거: Anthropic coding-task parallelism caveat — 코딩 과업은 연구와 달리 진병렬화가 드묾. 앵커 수리는 hook/session 패키지에 걸치는 의미 변경이라 단일 작성자 순차가 안전. 리서치는 이미 plan 단계에서 fanout으로 소진.
- Kickoff: 운영자 직답 승인(진입 승인·자율 진행) 2026-09-30 — 이 기록이 그 게이트의 결정 레코드
- Phase 1 skip 근거(위임 Section A에도 기재): verdict PASS 1.00(≥0.80) + 아티팩트 해시 불변(판정 후 spec/plan/acceptance/research 무변경 — progress.md는 해시 대상 아님)
