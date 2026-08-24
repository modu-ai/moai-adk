# SPEC-CODEX-VERDICT-SYNTH-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- 작성자: manager-spec
- Tier: **S** · 방법론: TDD
- 산출물: `spec.md` (REQ 4건 = 결함 3 + 회귀 방어 1, 지배 원칙 §0, 명명 속성 P-CONS) · `plan.md` (M1~M4) · `acceptance.md` (**AC 6건** · 속성형 2건 — AC-CVS-001 은 서식 corpus 8건(§B), AC-CVS-006 은 조합 corpus 8행(§B-2, M2 가드))
- 착수 트리: `.claude/worktrees/t229`, base `origin/main` @ `294b4b6ab`
- **일차 근거**: `.moai/reports/t229/premise-revision.md` §2 의 7행 실측표. `cause.md` 는 부분 스테일이며 어긋나면 premise-revision 이 이긴다
- v0.3.0 개정 (리드 판정): Tier M→S · §A.2 를 프로세스 랙 → **바이너리 랙**(설치본 `a1b1ca696`, `origin/main` 대비 259 커밋 뒤짐)으로 교체 · 보수 채택 순서를 결함에서 유지보수 메모로 강등 · AC 를 서식 열거형 → **속성형** 전환 · 후속 카드 t248 명시
- v0.4.0 개정 (plan-audit iter1 = PASS-WITH-DEBT 0.7625 수리):
  - **D1 (critical/blocking)** — 보수 채택 강등의 전제가 **2신호 구현에만** 성립함을 §A.5 에 명시. M2 가 세 번째 신호를 들이면 `Verdict: fail` + `PASS 0.95` → `pass` 세탁이 가능하므로 **AC-CVS-006** 을 신설해 채택값으로 고정. 요구사항은 강등 상태 유지(구현 형태 자유), plan.md §D M2 에 착수 게이트 추가
  - **D1 보강 (리드 추가 지시)** — 보수 규칙을 **순서 무관 집합 연산**으로 재서술. spec.md §A.5 에 명명된 속성 **P-CONS**(채택값 = 신호 **집합**의 최댓값, `fail` > `inconclusive` > `pass`)를 SPEC 본문 규칙으로 신설하고, "나중 신호가 앞선 것을 덮지 않는다" 류 순서 서술을 명시적으로 금지(신호 셋에만 유효 → 넷째에서 재발). AC-CVS-006 을 쌍 열거형 → **조합 corpus(§B-2, K1~K8) 순회 단일 단언**으로 재작성. 리드가 준 세 쌍은 K1·K3·K4 로 편입되어 증인이 됨. mutant (f) 순서 의존 · (g) **쌍 특수화** 추가, K1/K2(같은 집합·다른 순서) · K5(`scored × bullet`) · K6(3-신호) · K8(갈리지 않는 행)이 각각 증인
  - **D2 (major/blocking)** — §A.6 의 "게이트 차단" 논거가 측정상 거짓임을 확인(`isBlockVerdict` 는 `fail` 접두사만 차단, `codex_review_gate.go:116-117` / 종단 `:109`). **보고 정확성** 논거로 교체. AC-CVS-003 은 유지, mutant (b) 서술만 정정
  - **D3 (blocking)** — AC-CVS-001 의 RED 기대를 C1·C5·C7 → **C1~C8 전부**로 정정
  - **D9 (blocking)** — 스테일 인용 정정: `mcp_codex.go:1152`→`1155`, `mcp_convergence.go:368`→`367`. 추가로 감사 보고서에서 인용한 2건도 자체 실측으로 재정정: `codex_review_rpc_test.go:120`→`119`, `mcp_convergence.go:125-128`→`126-129`
  - 부수: **D4** 개방 항목 종결(`review-output.schema.json` 은 저장소에 파일로 존재하지 않음 — 실측), **D7** 모드 배선 증인을 도달 불가한 C7 → 도달 가능한 C5 로 교체
  - 미착수(리드 지시 범위 밖): D5(`Where`→`When` 패턴 정밀도) · D6(복합 REQ 분할) · D8(PR 경로 명시)
- 착수 순서 확정: 이 SPEC → t234 (= GitHub #1632)
- plan-audit iter2 대기

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
