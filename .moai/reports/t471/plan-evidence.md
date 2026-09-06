# Plan-Phase Evidence — card t471

- **SPEC-ID**: SPEC-LEAD-DEPUTY-001
- **Split decision**: 단일 SPEC (Tier M), 마일스톤 순서 Axis A(상주 deputy) → B(회차 보고 위임+파일 분할) → C(idle 통지). 근거: 세 축이 같은 두 표면(manager-lead.md, kanban-dispatch.md + detail)을 편집 / B는 A의 deputy가 전제 / C는 A의 감시 임무 변경 / A는 단독 출하 가능해 "Tier L 번들 금지·A 단독 가치" [HARD] 충족. 상세: spec.md §1.4
- **Tier**: M (마크다운 전용, 예상 6-8 파일, Go 변경 0)
- **Artifacts**: `.moai/specs/SPEC-LEAD-DEPUTY-001/{spec.md, plan.md, acceptance.md, progress.md}`
- **Commit**: (아래 기록)
- **AC 요약**: AC-LDP-001 기계 판정형 성공 지표(리드 턴 툴 배치 수, fixed window, stated baseline/target, 세션 로그 계수 레시피) · AC-LDP-002 idle-통지 경계(일정 힌트≠완료 증거) · AC-LDP-003 상주 spawn · AC-LDP-004 RECOMMEND-only · AC-LDP-005 delivery-shape 상속 · AC-LDP-006 deputy 귀속 · AC-LDP-007 파일 분할 · AC-LDP-008 템플릿 중립성 · AC-LDP-009 depth seal+Go 비접촉 · AC-LDP-010 위임 불가 3경계 보존
- **Audit status**: 미실행 (plan-auditor는 run-gate에서 Tier M 0.80 threshold로 실행). verdict 예정: `.moai/reports/t471/verdict.md`
- **Self lint**: `moai spec lint .moai/specs/SPEC-LEAD-DEPUTY-001/spec.md` → `✓ No findings — all SPEC documents are valid` (exit 0, this run, this tree)
- **측정 귀속**: §D.0 RED-now 기준값 전부 `[리드 자체 계수]` (lead-1 session 2026-09-03 self-count)
- **Depends_on**: SPEC-LEAD-DEBOTTLENECK-001 — completed → fulfilled
