# SPEC-ALWAYS-LOADED-DIET-002 — 진행 기록

카드 t1175 · Tier L · 워크트리 `.claude/worktrees/t1175` · 브랜치 `WT-rules-diet`

---

## §E.1 Plan-phase Audit-Ready Signal

- 상태: plan phase 산출물 작성 완료 (`status: draft`)
- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `design.md` · `research.md` · `progress.md` (Tier L 6종)
- SPEC ID 정규식 검사: Bash 실행, 출력 `PASS`
- 기준선(이 실행, 이 트리, 2026-09-25):
  - 18파일 `wc -m` 합계 = `246943 total`
  - 구속 조항 줄 = 170, sha256 `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`
- `[NEEDS CLARIFICATION]` **0건** — 작성 시점 2건이 리드 판정으로 해소됐다(`plan.md §B`):
  - B-1 `CLAUDE.md` 템플릿↔라이브 의미 분기 → **보존**, AC-ALD2-003 판정에서 제외. 실측으로 라이브 루트가 낡은 쪽임이 확정됐다(`moai cg --help` 가 은퇴 통지를 출력, `moai migrate cg` 존재). 별도 카드.
  - B-2 중복 제거와 동결의 경계 → 기본 입장 유지, **REQ-ALD2-013 으로 승격**.
- 리드 판정으로 추가된 항목: **범위 이탈** 잔여 위험(`acceptance.md §D.3`)과 재배치 표의 **범위 의존** 칸(AC-ALD2-004, `design.md §4`).
- **v0.2.0 (리드 판정)** — 파일당 40,000자 축 편입: REQ-014·015·016, AC-ALD2-009, `spec.md §D` 신규 Out of Scope 1건.
  - 이 축이 `design.md §2` 의 "새 companion 을 만들지 않는다" 결정과 충돌해 **설계를 개정**했다. 목적지는 존재가 아니라 **수용량**으로 고른다.
  - 실측: 계획대로면 셋이 한도 초과(`kanban-dispatch-detail` 57,036 · `session-handoff-examples` 51,616 · `agent-common-protocol-reference` 46,207). 셋째는 **이 카드가 만드는 신규 초과**이며 리드 보고에는 없던 항목이다.
  - 개정된 배정(`design.md §2.2`): 신규 companion 3개 생성, 이미 초과한 파일에는 쓰지 않는다. 마일스톤 M6.5.
  - 범위: 이미 초과한 룰 4개는 **수리하지 않고 악화만 금지**한다(REQ-015·016).
- 기준선 추가: 40,000자 초과 룰 파일 **4개**(두 트리 동일).
- **범위 이탈 공개를 양방향으로 확장**(`acceptance.md §D.3`) — 종전 공개는 축소 방향(참조 대상이 떠남)만 예시로 들었다. 확대 방향(범위를 **좁히던** 비구속 문장이 지워져 남은 조항이 넓어짐)을 방향 2 로 추가했고, 그것이 **결손이 아니라 증가**여서 손실 탐지기에 걸리지 않음을 명시했다. AC-ALD2-004 범위 의존 칸은 Q1(도달)/Q2(폭)의 두 질문을 묻도록 확장. 이 구멍은 감사가 아니라 **SPEC 작성자가 자기 공개에서 찾았고**, 발견 경위를 §D.3 에 남겼다.
- REQ 16개 / AC 9개(MUST-PASS 8). `moai spec lint SPEC-ALWAYS-LOADED-DIET-002` exit 0, findings 0.
- plan-audit iteration 1 은 ABORTED(`.moai/reports/t1175/plan-audit-iter1-aborted.md`) — 혼합 세대 읽기. iteration 2 는 커밋 SHA 를 입력으로 받는다.
- 커밋 대기 → plan-audit iteration 2

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
