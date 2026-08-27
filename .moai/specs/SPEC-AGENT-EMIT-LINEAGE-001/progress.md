# PROGRESS: SPEC-AGENT-EMIT-LINEAGE-001

## §E.1 Plan-phase Audit-Ready Signal

- card: t317 · worktree `.claude/worktrees/t317` (branch `WT-agent-emit-lineage`, base `48eb945df`)
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M 산출물 3종 + 워크플로 공통 progress.md)
- Tier: **M** — v0.4.0 재판정(v0.3.0 까지 S). 영향 파일 **전수 열거 5건** ≥ Tier S 의 `< 5 files` 경계 → 실격, Tier M 밴드 `5 - 15` 에 해당. 열거·산술은 plan.md §B.1/§B.2. plan-auditor PASS 임계 0.75 → **0.80**(iter-2 종합 0.82 는 이 임계도 넘는다)
- REQ **7** / AC **7** (Tier M 상한 각 16, 두 축 독립). **개수는 v0.2.0 이후 불변** — v0.3.0·v0.4.0 수리는 전부 기존 본문 확장이며, Tier 를 올린 축은 항목 수가 아니라 파일 수다. AC 중 4건(AC-AEL-001/002/003/006)이 뮤테이션 확립
- 근거 문서: `.moai/reports/t317/measurement.md` (실측 1-10), `.moai/reports/t317/plan-audit-iter1.md`, `.moai/reports/t317/plan-audit-iter2.md`
- 미해결 결정: **없음**. v0.3.0 에서 자동 호출 지점이 운영자 결정으로 종결됐고(`moai doctor` 항목 편입), v0.4.0 이 그 결정에서 **파생되는** 적용가능성 거동(배포 프로젝트 = 적용 불가 → `ok`, 종료 코드 불변)을 요구 층에 못박았다. 반영 지점은 REQ-AEL-004 「Applicability」 + AC-AEL-003 의 v0.4.0 4게이트
- 감사 이력: iter-1 D1-D6 전건 RESOLVED → iter-2 PASS-WITH-DEBT 0.82, 신규 D7·D8 → v0.4.0 에서 수리 → **iter-3 PASS 0.90 (0.74 → 0.82 → 0.90 단조 상승), 감사 iteration 상한 도달 — plan-phase 감사 종료.** 신규 D9-D12 는 전건 optional, blocking 0
- v0.5.0 델타(감사 후 편집, **재감사 대상 아님**): D9 단독 수리. AC-AEL-003 에 「하위 디렉터리 앵커」 게이트 1개 + REQ-AEL-004 적용가능성 술어에 기준점 해석 절 1개. 요구/수락 개수 불변(7 / 7), Tier M 불변
- **부채로 지고 가는 iter-3 결함 3건 (OPEN — 운영자 승인 하에 run-phase 진입)** — run-phase 가 이것들을 새 발견으로 놀라지 않도록 여기 이름으로 남긴다:
  - **D10 (minor, OPEN)** — `plan.md §B.1` 의 "전수 열거 5건" 이 `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` 3본을 빠뜨린다. doctor 항목을 하나 더하면 이 3본은 **반드시** 재생성된다(항목 행 + 그룹 카운터 `8 ok, 3 warn, 0 fail` 을 담고 있음). 실제 편집 파일은 **8건**이며 Tier 는 M 그대로. → M1 종료 조건에 골든 3본 재생성을 포함하고, 갱신 diff 가 새 항목 행 + 카운터 증가에 한정되는지 확인할 것
  - **D11 (minor, OPEN)** — `plan.md §B.1` 의 "이 리포의 doctor 항목은 파일 1개에 사는 것이 규약" 은 과잉 주장이다. MoAI-ADK 그룹 13개 중 자기 파일 보유 2건(`doctor_mcp_version.go:39`, `doctor_disk.go:66`), `doctor.go` 인라인 6건(`:417,:478,:495,:641,:941,:987`). 파일 #2·#3 은 규약이 아니라 **선택 가능한 형태**이며, 어느 쪽을 골라도 Tier 는 M
  - **D12 (minor, OPEN)** — `spec.md` 의 `golden_test.go:285` 인용은 `:284-285` 가 정확하다(`if` 는 `:284`, `want 11` 메시지가 `:285`). 표기 정정만 남았다
- 상태: `draft` — run-phase 미착수. Implementation Kickoff Approval 은 감사 PASS 로 대체되지 않는다

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
