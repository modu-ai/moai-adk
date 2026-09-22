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

## run 단계 (2026-08-22 04:14 리드 판정 PASS)

커밋 5+정오타 1: M1 `7a7a05384`(emitter core, agentemit 패키지, 커버리지 93.5%) · M2 `abf08c1f0`(프로브 — sandbox 값집합 {read-only, workspace-write, danger-full-access} / effort 열거 {minimal,low,medium,high,xhigh} / model 생략 / 서브디렉터리 스캔 전부 실측 확정, 모델 호출 2회 한정) · M3 `445bfd3b5`(11개 TOML 발행+골든/임베드/배포 가드) · M3b `e6c2239e5`(측정 기반 수정) · MS4 `48032316d`(close-out+§E) · 정오타 `a0a426d7c`.

- AC 13/13 PASS (§E.2/§E.3 근거, 검증 스냅샷 `48032316d:run-final`)
- 크로스플랫폼 빌드 0 · lint 신규 0 · [HARD] .md 트리 무변경 (골든 테스트 기계 보장)
- **MS3b 사건**: SPEC이 가정한 `mcp_servers = ["moai"]` 배열형이 codex-cli 0.147.0에서 정확히 7개 운반 에이전트 파일을 통째로 거부("invalid type: sequence, expected a map" — kill-the-file 위험 실증). 실측으로 `[mcp_servers.moai]` 테이블형 확정 → 수정 → 재실증(0 malformed, 11개 등록, 위임 SMOKE-OK). 정오타 커밋이 R-009·§A.3 row 9 시정. 리드 평: "plan에서 가정→smoke가 포획→실측으로 수정" 설계 패턴의 완결 사례.

## sync 단계 + 종결 (2026-08-22)

경량 sync (리드 스코프): CHANGELOG [Unreleased] Added 항목(mcp_servers 테이블형 발견 포함) · spec.md `in-progress → completed` · §E.4 신호 + `sync_commit_sha: b637ca710` (pending-backfill 패턴 → `9d1913050`로 채움). 3-phase close 완료.

## 잔여 (리드 소관/후속)

- 통합: release/v3.1.3에 merge --no-ff 후 push — WT는 release PR 머지까지 유지
- moai spec audit 종결 후 확인 (서브컨텍스트에서 MCP 불가 — 리드 통합 시점 실행 권고)
- P-05(skills.config)·P-06(에이전트별 MCP 필터링) 미프로브 — 계획된 생략(M1-deferred/선택)
- ~/.codex/auth.json mtime 관측: 앰비언트 의심·미확정 (Gap 기록)
- 후속 카드: M1(t81, skills 정본화 — manifest class-6 deferred→emitted 전환), M3(t83, 훅 어댑터), M4(t88, 배선 생성기 — 에미터 검증기 재사용 시임)
