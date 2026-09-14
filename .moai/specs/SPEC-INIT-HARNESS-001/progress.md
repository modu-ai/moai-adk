# progress.md — SPEC-INIT-HARNESS-001

## §E.1 Plan-phase Audit-Ready Signal

- card: t585 · tree: `.claude/worktrees/t585` · branch: `WT-init-harness-q` · HEAD at authoring: `a404132e7` (tree `aebfa5bea`)
- tier: L · artifacts: 6 (spec / plan / acceptance / design / research / progress)
- SPEC ID 자가검증: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS` (2026-09-14 실행, verbatim `PASS`)
- ID 충돌: `.moai/specs/` 디렉터 조회 0건 + 카탈로그 grep 0건 (SPEC-INIT-HARNESS-001)
- frontmatter: 12 캐노닉 필드 전부 + tier/depends_on/related_specs 옵션 — 스네이크 케이스 별칭 0
- depends_on 사전 점검: SPEC-INIT-QUIET-WIZARD-001 `status: completed` (2026-09-14 frontmatter 직접 판독) — 성립

### FO-PLAN-1 (병렬 리서치 팬아웃) 스킵 기록

레인 결정으로 FO-PLAN-1 병렬 조사 팬아웃은 **스킵** — 표준 단일 에이전트 디스커버리(manager-spec 자체 Grep/Read)로 수행. 사유 3개: (1) 리드의 이 공유 코드 표면에 대한 직렬 지시, (2) 카드가 전수 사전 조사를 이미 보유(`.moai/reports/init-tui-audit-20260909.md` — 표면 소속 지도·결함 17건·운영자 결정 포함), (3) 조사 표면이 좁다(배포기·미러·위저드·update·doctor의 하니스 축 한 지점). 대가로 모든 file:line을 이 트리 HEAD `a404132e7`에서 재측정했고 카드 전제 13건 중 2건 반증·1건 드리프트를 잡았다(research.md §1).

### Decision Point 1 (plan 후보 검토 게이트) 레인 노트

DP1은 리드 디스패치 승인 + 카드 본문(운영자 승인 조사 C3의 배출 의도)으로 레인 측 충족 — 본 SPEC의 후보 방향(harness 3-way 배포, codex 단독 본체)은 카드 텍스트가 곧 운영자 의도다. 이 SPEC에 대한 운영자 게이트는 본 산물 + plan-audit 이후의 Implementation Kickoff Approval이며 manager-spec이 요청하지 않는다.

### plan-phase 판단 메모 (리드 보고 항목)

- 카드 전제 반증 2건: commandemit 16종 착지됨(`e7d2a1658`, t503) — both 모드 신규 공사 소멸 / AGENTS.md 결속표 3행 존재(t523 `0755cc7f5`) — 조사의 "0행"은 낡은 측정.
- 명칭 드리프트 1건: 플래그는 `--agent`가 아니라 `--llm`(SPEC-CODEX-WIRING-001) — AC-CW-004 핀 축은 동일.
- 미해결 2건(운영자 소유 — plan.md §I): `--llm codex` 값 의미 변경 수용 여부 / claude 단독 트리밍(디스패치 문구 vs 조사 운영자 결정 "claude 단독=현행" 불일치 — 본 SPEC은 조사 결정 채택).

### iter1 개정 (plan-audit FAIL 0.94 → 수리)

audit iter1 FAIL 0.94 → D2(REQ-IH-003 바이트 동일 모순 → 파일집·로직 보존으로 재작성, REQ-IH-002/008 제외 명시 + AC-IH-005 정렬)+D3(가드 앵커 internal/config/token_budget_guard.go:94 정정)+D4(AC-IH-008 단일 호출형 + 실측 verbatim)+D5(스킬 로더 부분 공개 판정선 추가)+D6(REQ-IH-002 M1-산물 서술) 수리 완료, D7 관례 적합 무수리; MP-7 NC-1/NC-2는 리드 경유 운영자 대기.
NC-1·NC-2 운영자 확정(2026-09-14, 리드 경유) — NC-1 채택(codex-only deployment)·NC-2 현행 유지, plan.md §I 기록 + 본문 마커 RESOLVED 해제; plan_status 유지.

plan_complete_at: 2026-09-14T20:24:31+09:00 (plan 산물 검증 시점 실측 `date` 값)
plan_status: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop가 슬롯 지문(홈 지문 규약 REQ-IQW-014/016 준용)과 함께 기록>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs가 sync_commit_sha와 함께 기록>_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier=L · scope≈12파일(internal/cli/wizard/{questions,types,translations}.go, internal/cli/init.go, internal/config/defaults.go, internal/template/{deployer,skill_mirror}.go, templates AGENTS.md·llm.yaml, 신규·갱신 테스트) · 도메인 3(CLI 위저드 / config / 템플릿 배포기) · 언어 혼합 Go+YAML+Markdown · concurrency benefit=LOW(coding-heavy) · agent-team 사전요건=미요청(n/a)
- 모드 평가: direct=미선정(다중 파일 의미 변경) / serial=**선정** / fanout=미선정(coding-heavy — Anthropic caveat + 리드 직렬 지시) / sweep=미선정(기계적 균일 변환이 아닌 신규 코드 작업)
- Decision: serial
- 정당화: codex 단독 배포 필터는 단일 표면(deployer+harnessFS 래퍼)에 걸린 신규 코드 작업이다. manager-develop 단일 스폰으로 plan.md M1-M5 마일스퀀스를 직렬 실행한다 — Anthropic의 coding-task 직렬 기본과 리드 디스패치의 직렬 지시(같은 init 마법사 표면 병렬 금지)가 모두 serial을 가리킨다.
- Boundary case: 없음(serial 폴백 도달)
- Implementation Kickoff Approval: PASSED 2026-09-14 — 리드 경유 운영자 확정 2건(plan.md §I: NC-1 `--llm codex` 재정의 수용 / NC-2 claude 현행 유지) + 리드의 run 직행 지시(fanout 일괄 사전승인 정책 — 개별 Kickoff 질문 생략 선언) + plan-audit iter2 PASS 1.00(`.moai/reports/t585/plan-audit-iter2.md`, 로컬).
- Plan Audit Gate skip 판정: verdict PASS · score 1.00 ≥ 0.85(Tier L) · plan-artifact 해시 불변(iter2 감사 이후 해시 대상 5파일 무변경 — §F/§E.*는 해시 대상 아님) 3조건 충족 → run Phase 1 재실행 skip.
