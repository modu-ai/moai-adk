# SPEC-CODEX-CONFORMANCE-001 — progress.md

status: draft

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09
plan_audit: iter 1 FAIL 0.875 (MP-8 fail, blocking 6, required backend codex fail — receipt rcpt-f829a5596d422b6433546552, verdict `.moai/reports/t1607/plan-audit.md`); iter 2 repair applied — fixes 1-7(D1-D8) + lane inputs (D6 CI-trigger correction, 0.161.0 하위명령 실측 해소); iter 2 재감사는 감사자 429 사망으로 미완료; plan amendment round (2026-10-08T17:12:42Z) — 레인 실측(2026-10-08T16:56Z+) 접기: 제너레이터 출력 형상(39파일·v1/v2 분할·소비 8종 전부 존재, 이전 감사의 "이름 누락"은 ls 별칭 인공물로 반증) + vendoring 구성 결정(소비 부분집합, decision-index Q7 기본 적용); iter 2 FAIL 0.94 (MP-8 잔여 AC-CONF-005 + N2/N3/N4, receipt rcpt-2f21e144af2e5bef92bb82f6) — round-3 repairs applied (N4 005 채용 지연+기준선 원리 정정+Q4 앵커 갱신, N2 양세계 family-green 완결 경로, N3 죽은 출처 참조 제거, O2 스탬프 UTC 정정, O3 스테이징 디렉터)
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md (Tier M set + decision gate on)
red_now_ledger: acceptance.md §B (LEDGER-1/2/3 — tree 81786284e; AC-CONF-006의 RED 채용은 M3 E8 시점)
open_items: decision-index Q4 (EVIDENCE-NEEDED — 0.161.0 login-status 출력; 미인증 형태 포획은 AC-CONF-005의 최소 필수 관측으로 강화, keyring 변형만 기록 갭 허용)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~6 files (fixture dir + 2 test consumers + 1 ops doc + SPEC artifacts); domains 1 (Go CLI adapter + its tests); language mix Go + machine-generated fixtures; concurrency benefit LOW (coding-heavy, M1 fixture regeneration gates M2/M3); Agent Teams prereqs n/a (explicit-request-only).
- Mode evaluation: direct — not selected (semantic multi-milestone work); serial — selected; fanout — not selected (not research-heavy, single domain); sweep — not selected (semantic work, not mechanical-uniform bulk); agent-team — not selected (no operator request).
- Decision: serial
- Justification: single-domain coding-heavy conformance work with a strict milestone dependency chain — sequential manager-develop delegation with per-milestone commits is the safe default per the coding-task parallelism caveat; no other mode's selection criteria are unambiguously met.
- Kickoff record (autonomous form): decision record: decided_by=lane-21 evidence_refs=.moai/reports/t1607/plan-audit.md (iter-4 PASS 1.0, must_pass 0, blocking 0; codex required backend pass, governing receipt rcpt-f16b809f87c02e929297a2e9 fresh on the unchanged final state; plan_artifact_hash 93f2e64f8bbb38ba93a0ec664ddfa00796321566a9300da3f7223b115f34b795 via the canonical runtime.ComputeHash recipe — name:len:NUL+bytes+NUL per planArtifactNames order — unchanged since verdict; the earlier 21aba6a7 figure was the concatenation recipe and is superseded per gate review) ladder_path=plan-to-run Kickoff autonomous form — verdict PASS + 1.0 ≥ 0.80 Tier M threshold + artifact hash unchanged + no open blocker — run-phase entry approved.
