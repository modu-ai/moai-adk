# t89 — Codex 듀얼 하네스 M5 plan-phase verdict

Class C (design change — 전체 3단계 진행). 이 문서는 plan 단계 종결 기록.
worktree t89 / branch `WT-agent-toml-dual` (base origin/main @ 4b2f203fe).

- SPEC: `SPEC-CODEX-DUAL-AGENTS-001` (Tier M, era V3R6, status draft →
  run 진입)
- 커밋: `aa22605be` (plan artifacts 5파일 1,007행) + `255bf95b7` (§F
  mode selection) — 미푸시(통합은 리드 몫)
- evidence: `.moai/specs/SPEC-CODEX-DUAL-AGENTS-001/` + 감사 보고서
  `audit-plan.md` (3차 반복 전체 보존)

## 판정 요약

| 항목 | 결과 |
|---|---|
| plan-audit | **PASS (iter-3, 0.92)** — 궤적 0.86→0.92→0.92 무회귀, must-pass 7/7. Tier M 상한(2) 도달 후 리드 오버라이드로 3차 델타 감사(잔여가 토큰 1개) |
| 킥오프 | 발동 — 리드 일괄 승인(2026-08-22) + iter-3 PASS 조건 충족 |
| 핵심 설계 | **Option A 리드 승인**: template `.md` = 중립 코어, `.md` 발행 = identity([HARD] 바이트 동일이 구조적으로 보장), TOML = 변환. 매니페스트 `internal/template/agentemit/agents-codex.yaml` |
| 매핑표 | 14개 의미 클래스 전부 처분 확정(emit/consequence/documented-drop/deferred-to-M1/correspondence-note) — mcp__moai__* → `mcp_servers=["moai"]` 정확히 7개 에이전트, 미확정 값집합은 probe-우선·미확정 필드 생략(리드 비준 4건) |

## 감사 라운드 상세

1. **iter-1 FAIL 0.86** — clarification gate(마커 4개) + 검증된 재고 테이블
   오류 5셀(super-advisor mcp 11, sync-auditor 5, union 20/21, Web
   클래스 +builder-harness, DesignSync manager-design 단독). 교차모델
   audit_multi는 2차 백엔드가 둘 다 엉뚱한 대상 검토(codex=primary
   무관 파일+verdict 자기모순, GLM=환각 API) — in-session 증거만 채택.
2. **수정 라운드** — 기계적(D2-D5) manager-spec 적용(재측정 grep로 값
   확정), 마커 4건 리드 비준 디폴트로 치환(`grep -c 'NEEDS
   CLARIFICATION'` → 0), Option A 승인 기록.
3. **iter-2 FAIL 0.92** — 잔여: §A.3 row 9 스테일 토큰 "19 distinct"
   (정답 20; 1차 지시의 rows 4/8/9 중 9 누락). 상한 도달.
4. **리드 오버라이드 → iter-3 PASS 0.92** — 토큰 수정 grep 검증(0건)
   + N1 인용 정리 + 무변경 회귀 청결(mtime 상관 확인).

## run 인계 사항 (감사관 watch items)

- §A.4 DECIDED 4건이 MS2를 구속: probe 우선, 미확정 필드는 생략
  (t91 §1 silent-ignore — 잘못된 키는 조용히 무시됨)
- MS3의 AC-010 deploy fixture 테스트가 `.codex/` 루트 배포 전제의
  하중 지점 (CLAUDE.local.md §2.3 관리 뿌리에 `.codex/` 없음)
- §B.1 로컬↔템플릿 `.md` 드리프트 6파일은 골든 테스트가 유일한 고정
- codex-cli 0.147.0 고정 — 업그레이드 시 P-01..P-04 재실행

## Phase 4

`serial` (manager-develop, cycle_type=tdd) — 코딩 중심 Tier M, MS1→MS4
의존 사슬(emitter→probes→mass emission→close-out). progress.md §F에
입력·평가·근거 기록 완료. Plan Audit Gate skip: PASS + 0.92 ≥ 0.80 +
해시 무변경 (3조건 계약, §F에 기록).

## Gaps

- 3차 감사는 리드 오버라이드로 상한(2/2)을 넘어 수행됨 — 근거: 잔여
  토큰 1개 + 감사관 자체 제안 스코프. 오버라이드 승인 메시지가 근거 기록.
- audit_multi 워크트리 사각(2차 백엔드가 primary에서 실행)은 이번
  판정에 0 기여 — 리드가 후보 카드로 통합 제안 중이라고 회신.
