# Progress — SPEC-V3R6-DOCS-PREREWRITE-FLAG-001

Status: draft (plan-phase artifacts authored 2026-10-19, card t1726)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: iter2-defect-repaired-pending-re-audit
plan_revision: 0.1.2
plan_complete_at: 2026-10-19T03:44:13Z (0.1.2 iter-2 D7 수리 2026-10-19)
plan_commit_sha: pending-backfill-plan (커밋은 자기 SHA를 모르므로 플레이스홀더 — 커밋 후 이 SPEC 디렉터리를 건드리지 않는 후속 기록으로만 채운다)
iter1_audit: FAIL 0.94, blocking 4, receipt rcpt-25df35500d7f88da2fabbf8d (.moai/reports/t1726/plan-audit-iter1.md) — D1~D6 수리(b484ecacce), iter-2에서 재확인 해소.
iter2_audit: FAIL 0.88(점수 회귀 STOP), blocking 1 (D7), receipt rcpt-454d40d50ba7205d6fdcec4d (.moai/reports/t1726/plan-audit-iter2.md) — D7 수리 완료: AC-DPF-002 green path keep-절 수정 + STALE-3 부재 검증용 release-blocking AC-DPF-007 신설(RED-now 셀 E-05) + D8 확대(AC-DPF-006 허용 목록 합집합에 무pathspec committed diff 추가). REQ·기존 AC 식별자 불변, Q18 행 무변경. 재감사 대기.
open_decisions: Q18 (decision-index.md, FOUNDER, product-level, Operator verdict 비어 있음 — 이 행이 채워질 때까지 자율 Kickoff는 막히고, 판정은 리더 경유 운영자 라우팅을 기다린다). planner 기본값 (a)는 잠정 계획 입력일 뿐 자율 진입 승인이 아니다(plan.md F.1 M1 진입 선결 조건과 동일 문구). Q13·Q17은 운영자 결정 d-20261019T015908Z-220a(2026-10-19)로 기록됨(FOUNDER 라벨 + verdict 기록 형태).
evidence: RED-now 셀 E-01(트리 74ad197380)·E-02(재측정 트리 13cfafa454, iter-1 D4 — 74ad197380 대비 docs 바이트 무변경 확인)·E-05(iter-2 D7, 트리 b484ecacce — 이 실행, `git diff --name-only 74ad197380..HEAD`가 SPEC 아티팩트 5개 파일만 가리키는 docs 바이트 무변경 기반)와 baseline E-03·E-04는 acceptance.md 증거 장부에 있으며 모두 이 실행에서 측정됐다.

_런 페이즈 진행은 §E.2 이하 소관이다. manager-develop가 §E.2·§E.3을, manager-docs가 §E.4를 소유한다(이 파일의 그 소절은 plan 페이즈에서 채우지 않는다)._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
