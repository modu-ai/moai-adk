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
- **v0.3.0 (plan-audit iter2 수리)** — 판정 FAIL 0.805 / Tier L 임계 0.85, MUST-PASS 7/7 통과, Clarity 0.65 가 하락 원인. 입력 커밋 `c95124a5a`, 보고서 `.moai/reports/t1175/plan-audit-iter2.md`.
  - **D5 는 설계 변경이다** — 감사의 "문서 편집, 설계 변경 없음" 분류를 받아들이지 않는다. 포인터 줄 재유입 항을 파일별 가중으로 실측해 투영을 gross/net 으로 재구성. **재유입은 이동량에 비례하므로 목표 집합마다 값이 다르다** — 종전 목표 8,796 / 개정 목표 9,019. 감사 추정(~16,300)은 과대계상이었고 예측한 5,443 부족은 성립하지 않으나, 종전 목표로는 여유가 10,857 → **2,061** 로 깎였다. 저재유입 파일 8개 목표를 상향해 복원: gross 117,200 · 재유입 9,019 · **net 108,181 · 여유 11,238**. 재유입이 실측 2배여도 통과(+2,220).
  - **D7** — 재배치 표 발동 조건이 "옮겼을 때"뿐이라 삭제 경로(`AGENTS.md` M2 · `CLAUDE.md` M1′, 계획의 11.1%)가 통째로 우회됐다. REQ-017(삭제 행 의무)·REQ-018(stub 포인터 줄 검사) 신설, AC-004 Given 을 제거 전반으로 확장. **Q1/Q2 는 이동 경로에 대해 옳았고 구멍은 질문이 아니라 발동 조건에 있었다.**
  - **D1** — 지시대로 한 건이 아니라 전수 스윕. 살아 있는 개정 전 진술 **2건**(감사가 찾은 `plan.md §D`, 작성자가 보류 중이던 `design.md §2`), 역사 서술 5건은 유지. 둘을 각각 다른 주체가 찾았고 **아무도 둘 다 찾지 못한 것**이 이 결함군의 성질이다.
  - D2(검사 허용 ≠ SPEC 허용) · D3(초과 파일을 허용 선례로 인용) · D4(트리 표기) · D6(§E ↔ MUST-PASS 일대일) · D8(표의 귀속 축 한정) · D10(앵커 수리 vs 초과 파일 선례 규칙) 수리.
- REQ 18개 / AC 9개(MUST-PASS 8) / 추적성 18행. `moai spec lint` exit 0, findings 0.
- plan-audit iteration 1 ABORTED(혼합 세대 읽기), iteration 2 FAIL 0.805 → 수리 완료, iteration 3 대기.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
