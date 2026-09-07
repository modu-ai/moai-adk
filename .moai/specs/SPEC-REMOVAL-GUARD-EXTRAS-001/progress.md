# progress.md — SPEC-REMOVAL-GUARD-EXTRAS-001

Card: t511 · Branch `WT-danger-guard-regex` · Base `0b1e27877` (`origin/develop`) · Tier M

> §F Phase 4 Mode Selection은 오케스트레이터 소관 — 첫 런 페이즈 `Agent()` 스폰 전에 이 파일에 기록된다 (orchestrator-owned placeholder).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
plan_phase_artifacts: spec.md, plan.md, acceptance.md (Tier M set; research.md·progress.md 카드 관행 동반)
measured_baseline: 라이브 RED 3클래스(heredoc 데이터·unquoted 데이터·실행형 스크래치) 전부 extras 정규식 경로로 재현(verbatim 메시지 동일, tree 0b1e27877, research.md §R.2) · baseline 테스트 `ok ... 0.669s` (§R.1) · 결함 라인 사본 3곳 실측(디스패치 2곳 + testdata 신규 발견, §R.4) · plan 페이즈 코드·템플릿 변경 0건 (SPEC 아티팩트만 저작)

**사후 감사 개정 (post-audit-revision, 2026-09-07)** — plan-audit PASS 0.87 (S3 x4)에 대한 표적 응답. 판정 파일 `.moai/reports/t511/plan-audit-verdict.md`는 읽기 전용 보존. 네 해결:

- **F2 (AC-001 vs AC-008 자기 모순)**: 테스트 fixture는 R1의 실공간 예시 형태 행만 재현하고 패턴 리터럴 행은 모든 fixture에서 의도적 부재; 패턴 텍스트 필요 시 Go 프로그램적 구성 → plan.md:111 (M2.2 fixture 규율), acceptance.md:40 (AC-001 green path), acceptance.md:47 (AC-008 green path 도달 조건 명시)
- **F2c (스코프 grep 실측)**: `internal/` + `.moai/config/` 스코프 grep이 사전 수정 트리에서 정확히 3매치 실측 → acceptance.md:28 ([LEDGER-GREP3]), research.md §R.7.1
- **F1 (REQ-RGE-010 측정 AC 부재)**: 실측 AC 흡수 선택 — AC-010이 `git diff --numstat` deleted=1 added=0로 REQ-RGE-009+010을 함께 판정 (판별 뮤턴트 선택 실행) → acceptance.md:49
- **F3 (AC-009 RED 셀 4요소)**: 사후 감사 시점 라이브 재현 실행, verbatim 거부 기록 → acceptance.md:26 ([LEDGER-R4]), acceptance.md:48, research.md §R.7.2
- **F4 (docs-site 4로케일 낡은 예시)**: config-sections.md 131행 4로케일 실측 확인, sync가 같은 변경으로 갱신 의무화 → plan.md:128 (M3 §2), research.md §R.7.3

## §E.2 Run-phase Evidence

(pending — manager-develop, M1-M3 완료 시 채움)

## §E.3 Run-phase Audit-Ready Signal

(pending — manager-develop)

## §E.4 Sync-phase Audit-Ready Signal

(pending — manager-docs; sync_commit_sha는 sync 커밋 이후 backfill)

## §F Phase 4 Mode Selection

- Input parameters: tier=M · scope(files)=4(보안 yaml 3사본 + 테스트 1파일) · domains=2(hook 테스트, template/config) · language mix=Go 테스트 + YAML · concurrency benefit=LOW(코딩 중심) · agent-team prereqs=미요청
- Mode evaluation: direct=미선정(의미 변경 + 테스트 수반) · serial=**선정** · fanout=미선정(코딩 중심 — Anthropic coding-task caveat) · sweep=미선정(기계적 대량 변형 아님)
- Decision: serial
- Justification: 3 yaml 1행 제거 + 배포 정책 회귀 테스트 + 뮤턴트 2회의 코딩 중심 소규모 변경이라 단일 manager-develop 순차 스폰이 유일한 합리적 축이다. 병렬화 이익이 없어 fanout은 부적합하고, 균일 변형 대량 작업이 아니어서 sweep도 아니다. 카드 t511을 전 구간(plan→run→sync) 책지는 Factory 레인 구조와도 직렬 스폰이 정합이다. plan-audit iter2 PASS 0.93(아티팩트 해시 현재) 이후 첫 run 페이즈 스폰이다.
