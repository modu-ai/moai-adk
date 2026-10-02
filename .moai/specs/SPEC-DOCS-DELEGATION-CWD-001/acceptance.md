# acceptance.md — SPEC-DOCS-DELEGATION-CWD-001

## §D AC 매트릭스

전 항 단일 명령 + 종료 코드로 기계 판정한다. RED-now 셀은 본 SPEC 작성 시점 트리(`.moai/worktrees/t1387`, HEAD `e927266be`, 2026-10-01)에서 관측 — 문서 신설 2항(AC-DSC-001/002)과 가드 신설 1항(AC-DSC-003)은 오늘 실제로 red 다. AC 하위 쌍은 `AC-DSC-XXXa/b` 관례를 따른다.

| AC | REQ | 검증 (Given-When-Then) | 분류 |
|---|---|---|---|
| **AC-DSC-001** | REQ-DSC-010, REQ-DSC-003, REQ-DSC-004, REQ-DSC-007 | Given 본 SPEC 이슈 트리, When `grep -c "Reconciling an isolated specialist spawn" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` 실행, Then exit 0 이고 출력 ≥ 1 — RED-now (관측 2026-10-01, 트리 `e927266be`): `0` + `exit=1`; 양성 대조 `grep -c "## Isolation"` 같은 파일 → `1` + `exit=0`. 소절의 5개 필수 단계 마커가 REQ-DSC-003(가드 안전 화해)·REQ-DSC-004(WT- 개명)·REQ-DSC-007(증거 수확)을 운반한다 | release-blocking |
| **AC-DSC-002** | REQ-DSC-001, REQ-DSC-002, REQ-DSC-009 | Given 같은 트리, When `grep -c "runtime decision" .claude/rules/moai/core/agent-common-protocol.md` 실행, Then exit 0 이고 lane-session 단락(§ User Interaction Boundary 안) 근방에서 적중 — RED-now (관측 2026-10-01, 트리 `e927266be`): `0` + `exit=1`; 양성 대조 `grep -c "Background Agent Execution"` → `2` + `exit=0` | release-blocking |
| **AC-DSC-003** | REQ-DSC-011, REQ-DSC-010 | Given 가드 테스트 착지, When `go test ./internal/template/ -run '^TestReconciliationProcedureDocumented$'` 실행, Then `--- PASS` — RED-now (관측 2026-10-01, 트리 `e927266be`): `ls internal/template/docs_delegation_lane_flow_test.go` → `No such file or directory` + `exit=1` (테스트 부재). 변이 프로브: 절차 소절 임시 삭제 → RED 재관측 → 복원 → GREEN (§1.1 관측된 실패 완료) | release-blocking |
| **AC-DSC-003a** | REQ-DSC-011 | Given AC-DSC-003 GREEN 트리, When 절차 소절의 "증거 수확" 단계 마커만 삭제, When 가드 재실행, Then FAIL — 필수 단계 마커 개별 검사의 변이 관측 | release-blocking |
| **AC-DSC-004** | REQ-DSC-010 | **처분: (a) 유도 변이 RED — 기제 교체 동반.** GREEN: 깨끗한 트리에서 두 계측 PASS — `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$' -count=1` → `ok … 0.857s`, `go test ./internal/template/ -run '^TestSanitizedPairParity$' -count=1` → `ok … 0.180s` (관측 2026-10-01, 트리 `e927266be`). RED (유도 변이, 관측 완료): C2 `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md` 에 SPEC-ID 토큰 주입 → 같은 누출 테스트 → `FAIL github.com/modu-ai/moai-adk/internal/template`, 위반 라인 `[1] templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md \| class=C1-spec-id-prefix \| match=SPEC-DOCS-DELEGATION-CWD-001`; 복원 후 sha256 전후 일치(`4e04d962…94fa`). **기제 정정 기록**: 원 인용 `TestRuleTemplateMirrorDrift` 는 이 쌍을 검사하지 않음이 프로브 1로 증명 — C1 mechanics description 변이 후에도 `ok 0.360s` 유지 관측; mechanics 는 어느 바이트 패리티 목록에도 미등록, `agent-common-protocol.md` 는 §25 정화 쌍으로 바이트 패리티 설계상 제외 | release-blocking |
| **AC-DSC-005** | REQ-DSC-008 | RED-now (관측 2026-10-01, 트리 `e927266be`): `grep -c "ownership exception" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `0` + `exit=1`; 트리거 마커 `grep -c "structurally"` → `0` + `exit=1`; 양성 대조 `grep -c "twice"` → `1` + `exit=0`. GREEN: 문서 완성 뒤 두 마커 각각 ≥1 — 소유 예외의 트리거+기록 구조 성문화 확인 | release-blocking |
| **AC-DSC-006** | REQ-DSC-003 | RED-now (관측 2026-10-01, 트리 `e927266be`): `grep -c "git merge --ff-only" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `0` + `exit=1` (소절과 정당 형상 모두 부재); 양성 대조 `grep -c "merge --no-ff"` → `1` + `exit=0`. GREEN: 문서 완성 뒤 ff-only ≥1, 그리고 금지형 grep(`git -C .` / `git -C ..` / `--git-dir`) 0적중 | release-blocking |
| **AC-DSC-007** | REQ-DSC-005, REQ-DSC-006 | RED-now (관측 2026-10-01, 트리 `e927266be`): `grep -cE "implemented .* completed\|3-phase close" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `0` + `exit=1`; 양성 대조 `grep -c "## Integration into the release branch"` → `1` + `exit=0`. GREEN: 문서 완성 뒤 close 계약 명명 ≥1 — 단일 싱크 커밋 요소(implemented→completed, §E.4, 3-phase close) 확인 | release-blocking |
| **AC-DSC-008** | REQ-DSC-006 | RED-now (pre-run 그래프 판독, 관측 2026-10-01, 트리 `e927266be`): `git rev-list --count HEAD -- .moai/specs/SPEC-DOCS-DELEGATION-CWD-001` → `0`, `git log --oneline -- <같은 경로>` → 빈 출력 — M1/M2 커밋 부재가 오늘의 red 다. GREEN: run-phase 뒤 count ≥2, 커밋 순서 M1(문서) → M2(가드), 각 주부가 카드 id 를 명명 — 순서 주장의 목격자는 커밋 그래프(VCI §2.3) | release-blocking |

## §D.1 간접 검증

- REQ-DSC-003/006/007/009(행위 절차)는 문서+가드로 간접 검증 — 레인 플로우 자체는 다음 실제 격리 스폰에서 최초 적용된다. 선례 정합 확인: t1373 이 수행한 순서(검증→개명→ff-only→수확)가 문서화 절차와 일치함을 plan-auditor 가 spec.md §A.2 와 대조.

## §D.2 품질 게이트 기준

- `moai spec lint` 0 error · `go vet ./internal/template/...` clean · 변경 패키지 테스트 GREEN (전체 스위트는 CI).
- 룰 미러: `make build` 후 `git status --short` 가 예상 4파일+테스트 1파일만.

## §D.3 Definition of Done

1. AC-DSC-001..007 전 항 GREEN 출력 인용(커밋 SHA 핀) + AC-DSC-008 그래프 판독.
2. progress.md §E.1 plan 신호 기입, run-phase에서 §E.2/§E.3, sync에서 §E.4.
3. decision-index.md Q1/Q2 가 계획 감사 시점까지 미확정이면 plan-auditor 가 clarification gate 로 플래그 — kick-off 전 운영자 확답.
