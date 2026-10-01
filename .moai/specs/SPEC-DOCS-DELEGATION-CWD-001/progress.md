# progress.md — SPEC-DOCS-DELEGATION-CWD-001

## §A 현재 상태

- 카드: t1387 · SPEC 상태: `draft` (plan-phase 산출물 방출)
- 트리: `.moai/worktrees/t1387` · 브랜치: `WT-docs-delegation-cwd` · 작성 시점 HEAD: `e927266be`
- Tier M 산출물 4종 + decision-index 작성 완료 (2026-10-01, manager-spec)

## §B 결정 요약

- 수리 부류 확정: (a) 양상태 문서+화해 절차 + (c) 콘텐츠 가드. (b) 레인 직접 실행은 흐름에서 배제 — 소유 예외는 트리거+기록 의무의 최후 수단으로만 성문화(REQ-DSC-008). 근거: spec.md §A.4 트리거 통제 불가 부정 판정.
- 문서 착지지: 절차 본문은 paths 스코프 `kanban-dispatch-mechanics.md` § Isolation(부담 0), 의무 문장만 항상 적재 `agent-common-protocol.md` § User Interaction Boundary lane-session 단락(≤600B).
- 가드: `internal/template/docs_delegation_lane_flow_test.go` — 섹션+마커 콘텐츠 스캔, M1→M2 커밋 순서로 RED-first 를 그래프에 새김.
- **미러 안전 계측 정정(v0.1.1 — 프로브 2건 실측)**: 원 인용 `TestRuleTemplateMirrorDrift` 는 착지 쌍에 도달하지 않는다 — 프로브 1(C1 mechanics description 변이 → 테스트 `ok 0.360s` GREEN 유지)로 기계적 증명; mechanics 는 바이트 패리티 미등록, `agent-common-protocol.md` 는 §25 정화 쌍 제외. 교체 기제: `TestTemplateNoInternalContentLeak` + `TestSanitizedPairParity` — 프로브 2(C2 SPEC-ID 주입 → `FAIL`, 위반 라인 `class=C1-spec-id-prefix | match=SPEC-DOCS-DELEGATION-CWD-001`)로 RED 관측, 복원 sha256 전후 일치. AC-DSC-004 처분: (a) 유도 변이 RED(교체 기제 기준) + 기제 정정 기록.

## §C 열린 항목

- [해소 2026-10-01] decision-index Q1(배제 비준 APPROVED)·Q2(2회+거부 증거 CONFIRMED) — 운영자 판정이 kick-off AskUserQuestion 라운드(2026-10-01, card t1387 lane-1 pane direct answer)로 착지해 decision-index.md 의 Operator verdict 슬롯에 기입됨. D2 클로저 — kick-off 진입 가능 상태.
- `feedback_sync_dispatch_tree_anchor.md` 메모리 파일 미발견(양 저장소 검색) — spec.md §G 미검증 항목. run/sync 에서 발견되면 §A.3 에 보강.

## §E 진행 신호

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft-authored (plan-auditor 감사 대기)
- artifacts: spec.md · plan.md · acceptance.md · progress.md · decision-index.md (5파일, Tier M + decision gate on)
- frontmatter: 12 정규 필드 검증 완료 — SPEC ID 정규식 Bash 실행 PASS (`SPEC-DOCS-DELEGATION-CWD-001`), 기존 카탈로그 중복 없음(`ls .moai/specs | grep DELEGATION` 0건)
- lint: `moai spec lint SPEC-DOCS-DELEGATION-CWD-001` → "✓ No findings — all SPEC documents are valid" (1차: 0 error 14 warning → REQ↔AC 추적성 열 + `-run` 패턴 앵커 수리 후 0 findings; v0.1.1 델타 편집 뒤 재실행 0 findings — 2026-10-01 본 트리 실측)
- plan-audit r1: FAIL 0.875(MP-8 가림) — `.moai/reports/t1387/plan-audit-r1.md`. D1 델타 수리 완료(AC-004 처분 (a)+기제 교체, AC-005/006/007/008 RED-now 본 인 재측정치로 추가), D3·D4·D6 접수, D2는 kick-off 전 운영자 직답 유지. 재감사는 델타 스코프로 재위임 대기.

## §E.2 Run-phase Evidence

**트리 특이사항(격리 스폰)** — run-phase는 런타임 격리 워크트리 `.claude/worktrees/agent-a6ac0c2224378b493`(브랜치 `worktree-agent-a6ac0c2224378b493`, 기점 `f22e2d7ac`, card t1387 lane의 스폰)에서 수행됐다. 본 트리에는 plan-phase 산출물이 없어 SPEC 5파일을 카드 트리의 r2 감사본에서 바이트 동일 복사해 왔다(shasum 실측: spec.md `0938ee59…70a4`, plan.md `38ff1fb7…94c8`, acceptance.md `d24507cd…dbdb` — r2 감사 핀과 일치). 레인의 화해는 본 SPEC이 성문화하는 절차 본문 그대로 수행된다 — 카드의 첫 실측 사례. 커밋은 본 트리 브랜치에 쌓이며 push·병합은 하지 않는다(레인 화해 대기).

### M1 — 화해 절차 본문 + 항상 적재 의무 문장 (커밋 그래프 1번째)

- 편집 4파일: `.claude/rules/moai/workflow/kanban-dispatch-mechanics.md` + C2 미러, `.claude/rules/moai/core/agent-common-protocol.md` + C2 미러 — C1↔C2 편집 후 diff 바이트 동일 실측.
- AC-DSC-001 PASS: `grep -c "Reconciling an isolated specialist spawn" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `1`, exit=0 (RED-now 핀: `0` + exit=1).
- AC-DSC-002 PASS: `grep -c "runtime decision" .claude/rules/moai/core/agent-common-protocol.md` → `1`, exit=0 (lane-session 단락 27행 안 적중).
- AC-DSC-005 PASS: `grep -c "ownership exception"` → `1`; `grep -c "structurally"` → `1`; 양성 대조 `grep -c "twice"` → `2`.
- AC-DSC-006 PASS: `grep -c "git merge --ff-only" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `1`, exit=0. **범위 해석 기록**: 금지형 grep(`git -C` / `--git-dir`)은 소절 스코프로 측정 — 소절 내부 0적중(`sed` 추출 후 grep exit=1 ×2). 파일 전역은 기존 측정 인용(guard 거부 형상)이 6건(`git -C`, 행 50/79/81/91/101/104)과 1건(`--git-dir`, 행 81) 존재해 전역 0적중은 불가능 방향(impossible)이고, 신설 소절이 이들을 추가하지 않는 것이 criterion 의 실제 주장이다. 본 인 재측정: 편집 전 파일에 이미 6/1 건 존재(행 번호 실측).
- AC-DSC-007 PASS (N2 지시대로 UNESCAPED 형태): `grep -cE "implemented .* completed|3-phase close" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `1`, exit=0; 양성 대조 `## Integration into the release branch` → `1`.
- 항상 적재 예산: `agent-common-protocol.md` 증분 **487 바이트**(17,284 → 17,771) — ≤600B 목표 이내.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소유>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소유>_
