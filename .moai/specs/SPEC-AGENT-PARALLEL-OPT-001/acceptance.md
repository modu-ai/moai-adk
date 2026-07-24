---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Acceptance Criteria"
version: "0.7.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md, .claude/rules/moai/workflow, .claude/workflows, internal/template/templates, internal/template/internal_content_leak_test.go"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
---

## §A 판정 원칙

- 모든 AC는 **기계 검증 가능**하거나(grep / diff / `wc -l` / `go test`), 명시적 구조 검사로 판정한다.
- 판정 근거는 명령의 **verbatim 출력**이어야 한다. 요약·추정·이월 수치는 근거로 인정하지 않는다.
- 라인 수 AC는 `wc -l` 출력의 정확한 정수로 판정한다. 근사치("약 340줄") 금지.
- MUST 등급 AC 1건이라도 FAIL이면 SPEC은 close 불가다.

---

## §B Given-When-Then 시나리오

### 시나리오 1 — 런타임 미지원 환경에서의 graceful degradation

**Given** 배포 사용자가 3개 스크립트 파일은 수령했으나 런타임이 dynamic workflow를 지원하지 않고(구버전 Claude Code),
**And** `plan.md`가 `plan-research-fanout.js`를 capability-gate 형태로 참조할 때,
**When** 사용자가 `/moai plan "<feature>"`를 실행하면,
**Then** 워크플로우는 기존 단일 Explore 리서치 경로로 진행되고,
**And** 오류·경고·중단이 발생하지 않으며,
**And** 산출되는 SPEC 산출물 집합은 배선 이전과 동일하다.

파일 부재 환경(스크립트를 수동 삭제한 사용자)에서도 동일 결과여야 한다 — gate는 "파일 존재 AND 런타임 지원" 두 조건을 모두 확인한다.

### 시나리오 2 — 병렬 심사 후 단일 구속 verdict

**Given** run Phase 품질 단계가 병렬 증거 수집 + 단일 verdict 구조로 재구조화되어 있을 때,
**When** 4개 차원 read-only 심사가 병렬 완료되면,
**Then** 최종 PASS/FAIL verdict는 `sync-auditor` 1개 에이전트가 산출하며,
**And** 스크립트가 계산한 집계 점수는 증거로만 인용되고 verdict를 대체하지 않는다.

### 시나리오 3 — 다이어트 후 SSOT 도달성 보존

**Given** `manager-spec.md`에서 12-field frontmatter 스키마 블록이 제거되었을 때,
**When** 독자가 해당 위치를 읽으면,
**Then** `spec-frontmatter-schema.md`를 가리키는 교차참조가 존재하고,
**And** 그 경로의 파일이 실재하며,
**And** 12개 필드 정보가 그 파일에서 조회 가능하다.

### 시나리오 4 — Template-First 위반 탐지

**Given** 미러가 존재하는 파일이 편집되었을 때,
**When** 로컬 파일과 템플릿 미러를 `diff`하면,
**Then** §C Pre-flight에서 확정한 baseline 차이 외의 신규 차이가 0이다.

---

### 시나리오 5 — 배포가 harness Runner 격리를 깨뜨리지 않음

**Given** 3개 generic fan-out 스크립트가 `internal/template/templates/.claude/workflows/`에 배포되어 있을 때,
**When** `go test ./internal/template/...`를 실행하면,
**Then** `TestSplitHarnessNamespaceNoLeak`이 통과하고,
**And** `hns-release-update-run.js` 등 dev-only Runner를 템플릿에 심으면 여전히 FAIL한다(차단 유효성 확인),
**And** `moai update`가 3개 스크립트를 template-managed로, 사용자의 `hns-*` Runner를 user-owned로 분류한다.

### 시나리오 6 — 중립성 판정이 공허하지 않음

**Given** `leakTextExtensions`에 `.js`가 추가되어 있을 때,
**When** 중립화되지 않은 스크립트(`REQ-ATR-018` 등 포함)를 템플릿에 심고 leak 테스트를 실행하면,
**Then** 테스트가 **FAIL**하고(스캐너가 실제로 `.js`를 읽었음의 증거),
**And** 중립화 후 재실행하면 PASS한다.

## §C 엣지 케이스

| # | 케이스 | 기대 동작 |
|---|---|---|
| E1 | manifest 교집합 판정이 모호(glob 패턴 중첩) | 직렬화가 기본값 — 판정 불가는 "겹침"으로 취급 |
| E2 | fan-out 중 1개 drafter가 blocker report 반환 | 나머지 결과는 유지, 오케스트레이터가 해당 항목만 재위임 |
| E3 | `moai verify` 스냅샷이 TTL 만료 | 스냅샷 인용 금지, 해당 검사 재실행 |
| E4 | 라인 상한을 충족하나 교차참조 누락 | REQ-APO-068 FAIL — 상한 충족만으로 PASS 불가 |
| E5 | 템플릿 편집 중 SPEC ID 문자열 유입 | 중립성 가드 FAIL, 즉시 제거 |
| E6 | D3 게이트 grep에서 마커 소비자 발견 | REQ-APO-043은 출력 계약 보존 + 주변 산문만 축약으로 전환 |
| E7 | docs-site 주장 정합화 시 4-locale 중 일부 누락 | 4-locale 동시 수정 의무 위반 — FAIL |
| E8 | fan-out 폭이 5를 초과 | Mode 4 상한 위반 — 3-5로 조정 |
| E9 | `.js` 확장자 추가 없이 중립성 green 주장 | 공허 통과 — AC-APO-071 증거 무효, FAIL |
| E10 | 중립화 중 스크립트 실행 로직 변경 | 범위 위반 — 헤더·주석만 대상 |
| E11 | 배포 후 `hns-*` 차단이 함께 완화됨 | dev-only 격리 훼손 — AC-APO-072b FAIL |
| E12 | 3개 스크립트가 user-owned 보존 집합에 편입됨 | `moai update`가 갱신 불가 — AC-APO-073 FAIL |

---

## §D AC 매트릭스

### D.1 Group 1 — fan-out 배선 (REQ-APO-010..016)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-010 | 010 | MUST | `grep -c "plan-research-fanout" .claude/skills/moai/workflows/plan.md` ≥ 1 (baseline 1 — `plan.md` 실측). **재귀형 금지**: REQ-APO-010이 지정한 표면은 `plan.md` 단일이므로 파일 앵커가 필수다. 재귀형(`grep -rn … workflows/`)은 배선이 다른 파일로 이동해도 GREEN이라 요구 표면을 검증하지 못한다(AC-APO-012가 실제로 겪은 이설 해저드와 동형). `grep -c`는 0건일 때 **exit status 1**을 반환하므로 종료 코드가 아니라 **출력 숫자**로 판정한다 |
| AC-APO-011 | 011 | MUST | 3개 참조 각각에 capability-gate 문구 동반 — gate 조건이 **파일 존재 AND 런타임 지원** 두 가지를 모두 명시하고, 참조 건수와 gate 문구 건수 일치 |
| AC-APO-012 | 012 | MUST | **2항 동시 충족**(어느 한쪽 회귀도 독립적으로 FAIL): (a) `grep -c "sync-audit-4dim" .claude/skills/moai/workflows/sync.md` ≥ 1 (baseline 1 — `sync.md:56`), (b) `grep -c "sync-audit-4dim" .claude/skills/moai/workflows/run/task-decomposition.md` ≥ 1 (baseline 1 — `:104`). **재귀형 `≥ 1건` 금지**: REQ-APO-012는 2개 표면을 요구하는데 재귀 `≥1`은 한쪽만 있어도 통과한다(실측 반증 — `task-decomposition.md` 배선 삭제 후에도 재귀 카운트 1로 GREEN 잔존). 재귀형에 `== 2`를 걸어도 두 매치가 한 파일에 몰리면 통과하므로 **파일별 판정만이 유효**하다. `grep -c`는 0건일 때 **exit status 1**을 반환하므로 종료 코드가 아니라 **출력 숫자**로 판정한다 |
| AC-APO-013 | 013 | MUST | 각 스크립트 참조 지점 인근에 verdict 소유권 보존 문장 존재(auditor가 verdict 소유) |
| AC-APO-014 | 014 | MUST | `grep -c "codemaps-extract" .claude/skills/moai/workflows/codemaps.md` ≥ 1 (baseline 1 — `codemaps.md:83`), 그리고 high-count 스코핑 문구 동반. **재귀형 금지(공허 GREEN 실측 확인)**: `harness-builder.md:81`이 REQ-APO-014과 **무관한 선례 인용**으로 `codemaps-extract`를 포함하므로, 재귀 `≥1`은 요구 표면인 `codemaps.md` 배선을 삭제해도 그 무관 인용 1건으로 GREEN을 유지한다 — 즉 재귀형은 대상 표면을 전혀 검증하지 못한다. `grep -c`는 0건일 때 **exit status 1**을 반환하므로 종료 코드가 아니라 **출력 숫자**로 판정한다 |
| AC-APO-015 | 015 | MUST | zero-orphan: 3개 스크립트명 각각이 `.claude/skills/moai/workflows/` 하위에서 최소 1건 매치 (3/3). **의도적으로 느슨한 판정 — 강화 금지**: 본 AC의 목적은 *고아 스크립트 탐지*(배포했으나 아무 데서도 참조되지 않는 `.js`)이지 표면별 커버리지가 아니다. 스크립트당 `≥1`이 이미 per-item 판정이며, 어느 파일에서 매치되는지는 본 AC의 관심사가 아니다. 요구 표면별 배선 검증은 AC-APO-010(`plan.md`) / 012(`sync.md` + `run/task-decomposition.md`) / 014(`codemaps.md`)가 소유한다 |
| AC-APO-016 | 016 | MUST | docs-site 4-locale `workflows.md`의 파이프라인 투입 주장이 **참임이 검증**됨 — AC-APO-015(zero-orphan) AND AC-APO-069(배포 존재) 동시 PASS가 근거. 배선/배포 미완 시에는 4개 로케일 동시 정정으로 대체 판정 |

### D.2 Group 2 — 재구조화 (REQ-APO-020..030 + 024b)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-020 | 020 | MUST | plan Phase 11 서술에 병렬 read-only 심사 + plan-auditor 단일 verdict 구조 명시 |
| AC-APO-021 | 021 | MUST | plan Phase 10 단일 writer 유지 명시 + 단일 턴 병렬 Write 지시 존재 |
| AC-APO-022 | 022 | SHOULD | RED 단계 drafter pool + 단일 applier 구조 서술 존재 |
| AC-APO-023 | 023 | MUST | run Phase 13/16/17 축약 구조 서술 존재 **AND** `grep -cE 'Maximum 3 (fix-evaluate cycles\|review iterations)' .claude/skills/moai/workflows/run/task-decomposition.md` == 2 (baseline 2 — **Phase 16**(Active Quality Evaluation) `Maximum 3 fix-evaluate cycles` @:190, **Phase 17**(TRUST 5 Static Verification) `Maximum 3 review iterations` @:230. Phase 13은 반복 상한을 갖지 않는다. 둘 중 하나라도 삭제되면 1로 떨어져 FAIL) |
| AC-APO-024 | 024 | MUST | sync Phase 12에 5개 read-only drafter 구조 서술 + 단일 `manager-docs` 순차 적용 명시; disjoint-writer 변형 서술 0건 |
| AC-APO-024b | 024b | MUST | Phase 12 서술이 **현행 write-concurrency 규칙과 독립**임이 명시(현행 `[HARD]` 절대 금지형 규칙을 그대로 둔 채 성립) **AND** 규칙 완화의 진행 여부를 전제로 삼는 서술 0건 — `grep -cE "규칙 완화(가|를) (전제|선행)\|write-concurrency 개정.*(선행\|의존)" .claude/skills/moai/workflows/sync/doc-execution.md` == 0 (baseline 0 — 실측 확인). 주의: `grep -c`는 0건일 때 **exit status 1**을 반환하므로 종료 코드가 아니라 **출력 숫자**로 판정한다 |
| AC-APO-025 | 025 | SHOULD | sync Phase 10 패키지별 fan-out 서술 존재 |
| AC-APO-026 | 026 | MUST | sync Phase 1/7에 `moai verify` 스냅샷 소비 서술 + 신선도(키 일치/TTL) 조건 동반 |
| AC-APO-027 | 027 | SHOULD | MX 스캔 샤딩 서술 존재 |
| AC-APO-028 | 028 | MUST | 모든 신규 fan-out 서술이 오케스트레이터 launch로 기술 — subagent nesting 의존 문구 0건 |
| AC-APO-029 | 029 | MUST | 게이트 토큰 4종(`Decision Point 1` / `Implementation Kickoff Approval` / `gate-sync-1` / `gate-sync-2`) 모두 편집 후에도 존재 |
| AC-APO-030 | 030 | MUST | drafter/judge 서브에이전트 서술에 blocker report 반환 규범 존재, `AskUserQuestion` 호출 지시 0건 |

### D.3 Group 3 — 본문 다이어트 (REQ-APO-040..055)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-040 | 040 | MUST | `plan-auditor.md` 내 12-field 열거 블록 1개 이하 (현재 2개: MP-3 + FC-1..12) |
| AC-APO-041 | 041 | MUST | `manager-spec.md`에 12-field 열거 블록 0개, `spec-frontmatter-schema.md` 교차참조 ≥ 1건 |
| AC-APO-042 | 042 | MUST | `grep -c "Chain-of-Verification" .claude/agents/moai/plan-auditor.md` == 0 |
| AC-APO-043 | 043 | MUST | **선행 게이트 (판별형)**: `grep -rn 'decomposition:' --include='*.go' --include='*.sh' --include='*.yaml' internal/ .github/ .claude/hooks/` 를 실행하고 출력을 `progress.md`에 verbatim 기록. **기대 출력량 0-5줄**(plan-phase 참고 실측 0건) — 이 범위를 넘으면 명령이 잘못 좁혀진 것이므로 재작성. 비판별형(`decomposition\|segment match trace` 전역 grep, 12,133-match)은 **사용 금지**(§F DoD와 동일 규범). 분기: **소비자 0건** → 마커 강제 제거 + 블록 축약(라인 상한 분기 A); **소비자 ≥1건** → 마커 출력 계약 보존 + 주변 산문만 축약(라인 상한 분기 B). 어느 분기든 실행 Bash 검사는 존치 |
| AC-APO-044 | 044 | MUST | Step 5 체크리스트 항목 중 Step 4 서술의 축자 재진술 0건 |
| AC-APO-045 | 045 | MUST | `manager-spec.md`에 GEARS/EARS 패턴 표 0개, `moai-workflow-spec` 교차참조 ≥ 1건 |
| AC-APO-046 | 046 | MUST | `manager-spec.md` Step 4의 산출물 개수 서술이 실제 열거 개수와 일치 |
| AC-APO-047 | 047 | MUST | `manager-develop.md`에 DDD/TDD 전문 2회 기술 부재 — 공통 골격 + 모드 차이 구조 |
| AC-APO-048 | 048 | MUST | "one atomic change" 제약에 패키지 내부 한정 수식어 존재 |
| AC-APO-049 | 049 | MUST | 3개 판정 동시 충족: (a) **선택 규칙 산문 소멸** — `grep -ci "two scoring models\|scoring model selection" .claude/agents/moai/sync-auditor.md` == 0 (baseline 2 — L44 `## Scoring Model Selection`, L46 `Two scoring models exist`). 잔존 모델을 설명하는 일반 표현(예: `## Scoring`)은 허용된다. (b) `grep -c "^## Evaluation Report" .claude/agents/moai/sync-auditor.md` == 1 (baseline 2 — L67 평면형 + L178 계층형). (c) **정확히 1개 모델만 잔존** — 두 **정의 마커**의 합이 1: `M=$(grep -c "^## HRN-003 Hierarchical Scoring Protocol" .claude/agents/moai/sync-auditor.md); N=$(grep -c "^### Dimension Scores" .claude/agents/moai/sync-auditor.md); test $((M+N)) -eq 1`. **baseline M=1(:130 정의 heading), N=1(:71 평면 모델 report) → 합 2 → FAIL**. `M`은 반드시 heading 앵커(`^## `)여야 한다 — 앵커 없는 `grep -c "HRN-003 Hierarchical Scoring Protocol"`는 2를 반환하며(:49는 정의가 아니라 산문 cross-reference), 그 형태를 쓰면 "계층형 유지 + 평면형 제거" 분기가 M=2·N=0·합 2로 **부당 FAIL** 한다. 앵커 적용 시 두 분기 모두 합 1 → PASS |
| AC-APO-050 | 050 | MUST | `grep -ciE "nextra\|wcag\|page.?speed\|lighthouse" manager-docs.md` == 0 |
| AC-APO-051 | 051 | MUST | `e2e-tester.md`에 비-호스트 OS 레시피 본문 부재 **AND** `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` 실재(`test -f`) **AND** `e2e-tester.md`가 해당 경로를 참조 **AND** 템플릿 미러 `internal/template/templates/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` 존재 + 0-diff |
| AC-APO-052 | 052 | MUST | `grep -c 'squash | merge | rebase' .claude/agents/moai/manager-git.md` == 1 (baseline 3 — L126 주석 / L163 auto-merge / L191 manual). **해석 규칙**만 1회화 대상이며, L163·L191의 두 운용 경로와 각각의 `gh pr merge --<merge_method>` 명령 템플릿은 **보존**(제거 시 REQ-APO-068 위반) |
| AC-APO-053 | 053 | MUST | `builder-harness.md`의 `Model/effort escalation` 중복 문장 1개 이하, model-policy 표 재진술 0건 |
| AC-APO-054 | 054 | MUST | `grep -l "verification-batch-pattern\|Parallel Execution" .claude/agents/moai/*.md` 결과 파일 수 ≥ 8 (현재 1) |
| AC-APO-055 | 055 | MUST | `wc -l .claude/agents/moai/*.md` 합계와 파일별 값이 **모두** `spec.md` §D.2 표의 적용 분기 상한 이하. 합계 상한은 분기 조건부 — **분기 A ≤ 1907** / **분기 B ≤ 1927**(`manager-spec.md` 230 → 250 차이). 적용 분기는 AC-APO-043 게이트 결과가 결정하며, 분기 B 적용은 "MUST 미달성 + 사유 기록"이 아니라 **정상 적용**이다 |

### D.4 Group 4 — 불변식 (REQ-APO-060..068)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-060 | 060 | MUST | 편집된 미러 쌍 전량 `diff` 결과가 Pre-flight baseline 차이 외 0 |
| AC-APO-061 | 061 | MUST | `git diff --name-only origin/main...HEAD -- internal/template/templates/ \| xargs -r grep -nE "SPEC-(V3R[2-6]\|AGENCY\|WORKTREE)-[A-Z0-9-]+\|(REQ\|AC)-(ATR\|WO\|COORD\|UNP\|LNC\|TII)-[0-9]{3}\|REQ-APO-\|AC-APO-"` == 0. 정규식은 CI 가드 클래스(C1/C2)에 정렬; `REQ-APO-`/`AC-APO-`는 본 SPEC 고유 토큰으로 추가 |
| AC-APO-062 | 062 | MUST | `ls internal/template/templates/.claude/workflows/` 결과에 `hns-*` / `harness-*` 접두 파일 0개 (generic fan-out 3개만 존재) |
| AC-APO-063 | 063 | MUST | 시나리오 3 통과 — 스크립트 부재 시 fallback 경로가 문서상 완결 |
| AC-APO-064 | 064 | MUST | `go test ./...` exit 0; template-neutrality CI guard green |
| AC-APO-065 | 065 | MUST | `git diff` 상 `.claude/agents/moai/*.md` frontmatter 블록 변경 0줄 |
| AC-APO-066 | 066 | MUST | archived 12개 에이전트명이 편집 산출물에서 신규 매치 0건 |
| AC-APO-067 | 067 | MUST | `go test ./internal/template/...` exit 0 (`split_namespace_test.go`, `internal_content_leak_test.go` 포함) |
| AC-APO-068 | 068 | MUST | 제거된 각 중복 블록 위치에 SSOT 교차참조 존재, 참조 경로 파일 전량 실재 |

### D.5 Group 5 — 배포 (REQ-APO-069..073)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-069 | 069 | MUST | `ls internal/template/templates/.claude/workflows/` 에 3개 파일 존재; `make build` 후 embedded 트리에서도 조회 가능 |
| AC-APO-070 | 070 | MUST | 4개 동시 충족: (a) 본 SPEC frontmatter `partially_supersedes: [SPEC-DWF-CODEMAPS-PILOT-001]`, (b) `spec.md`가 superseded AC를 **ID로 인용** — `AC-DCP-010`(`acceptance.md:79` / `progress.md:86`) + 그 소유 요구사항 `REQ-DCP-009/010`, (c) 정식 grep 문구 `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing 이 더 이상 성립하지 않음을 명시, (d) 파일럿 SPEC 아티팩트에 supersession 주석 추가로 상호 참조 성립 |
| AC-APO-071 | 071 | MUST | 배포된 3개 파일에 대해 **CI 정렬 정규식** `grep -nE "SPEC-(V3R[2-6]\|AGENCY\|WORKTREE)-[A-Z0-9-]+\|(REQ\|AC)-(ATR\|WO\|COORD\|UNP\|LNC\|TII)-[0-9]{3}\|[0-9a-f]{7,8}([[:space:].,;:!?]\|$)"` == 0. 정규식은 `internal_content_leak_test.go` C1/C2/S2 클래스에 정렬됨 — 일반형 `SPEC-[A-Z0-9-]+-[0-9]{3}`는 **의도적으로 제외**한다(`spec.md` §F.8.3-a: `SPEC-FOO-001`은 CI 미매치 일반 플레이스홀더이며 중립화 대상이 아니다). **선행 조건**: AC-APO-072 PASS (미충족 시 본 AC는 공허하여 무효) |
| AC-APO-071b | 071 | SHOULD | **manual-only / CI-unenforced 표기 의무 이행** (`spec.md` §F.8.3-a 귀결 3). 아래 클래스는 CI 가드가 강제하지 **않으므로** 수동 점검이며, "CI green"을 근거로 인용하지 않는다: 내부 날짜 `20[0-9]{2}-[0-9]{2}-[0-9]{2}`, 메인테이너 절대경로 `/Users/`, 9자 이상 SHA `[0-9a-f]{9,40}`(CI S2는 `{7,8}`로 **겹치지 않음**). 판정 **2항 동시 충족**: (i) 배포된 3개 파일에 대해 `grep -nE "20[0-9]{2}-[0-9]{2}-[0-9]{2}\|/Users/\|[0-9a-f]{9,40}"` **== 0** — REQ-APO-071이 금지한 5개 클래스 전량이 MUST 수준으로 커버되도록 하는 조항이며, SHOULD 등급이라는 이유로 면제되지 **않는다**. (ii) 그 결과가 `progress.md`에 **CI-unenforced 라벨과 함께** 기록됨 — 기록 의무는 "CI green"을 이 3개 클래스의 근거로 오인용하지 못하게 하는 장치다 |
| AC-APO-072 | 072 | MUST | `leakTextExtensions`에 `".js": true` 존재 (`grep -n '".js"' internal/template/internal_content_leak_test.go` ≥ 1) **AND** 시나리오 8의 RED/GREEN 왕복이 관측됨 — 미중립 스크립트 심었을 때 FAIL, 중립화 후 PASS |
| AC-APO-072b | 062/069 | MUST | `TestSplitHarnessNamespaceNoLeak` PASS **AND** 차단 유효성 확인: `hns-release-update-run.js`를 템플릿에 심고 실행 시 `SPLIT_HARNESS_NAMESPACE_LEAK`으로 FAIL (심은 파일은 제거) |
| AC-APO-073 | 073 | MUST | `internal/cli/update/plan/plan.go`의 user-owned 판정이 3개 generic 스크립트에 대해 false 반환 — 접두사 `hns-`/`harness-` 미매치로 확인. 보존 목록 소스 무변경(`git diff` 0줄) |

### D.6 Group 6 — 배포 정합성 (REQ-APO-074..078)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-074 | 074 | MUST | 3항 동시 충족: (a) `grep -c "MoAI does not ship any saved workflows by default" .claude/rules/moai/workflow/dynamic-workflows.md` == 0 (baseline 1, L80 — 배포 후 거짓이 되는 전칭 주장). (b) **무한정 전칭형 소멸** — `grep -c "the user-owned \`.claude/workflows/\` directory is not template-managed" .claude/rules/moai/workflow/dynamic-workflows.md` == 0 (baseline 1, L80). `not template-managed` 문구 **자체는 금지되지 않는다** — `hns-*` / `harness-*` 한정 서술로는 보존되어야 하며(`design.md` §E R5), L131의 "사용자 자신이 검증한 스크립트" 서술도 여전히 참이다. (c) 개정문에 MoAI-shipped generic fan-out과 user-owned `hns-*`/`harness-*` 구분 서술 존재 |
| AC-APO-075 | 075 | MUST | `diff .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` == 0-diff (개정 전 baseline도 0-diff — 실측 확인) |
| AC-APO-076 | 076 | MUST | 대상 3개 파일을 **명시 열거**해 판정(glob은 out-of-scope `hns-*` Runner를 포함하므로 금지): `grep -l "user-owned workflows" internal/template/templates/.claude/workflows/{plan-research-fanout,sync-audit-4dim,codemaps-extract}.js \| wc -l` == 0 **AND** 동일 명령을 로컬 `.claude/workflows/` 경로로 실행해도 == 0. baseline 2 (`plan-research-fanout.js:36`, `sync-audit-4dim.js:38`). `grep -c` 다중 파일 형태는 파일당 `path:count` 줄을 출력하므로 == 0 비교에 부적합 |
| AC-APO-077 | 077 | MUST | `plan.md` M1 작업 순서가 **템플릿 원본 우선 → 로컬 파생** 임이 문서상 확인(로컬 선편집 후 복사 서술 0건) **AND** `dynamic-workflows.md` 개정문에 `moai update`가 3개 스크립트의 로컬 사본을 덮어쓴다는 서술 존재 |
| AC-APO-078 | 078 | MUST | 전체 경로 `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` 가 3곳에 문자열로 존재: (a) `spec.md` §E.2 파일 인벤토리, (b) `plan.md` M4 작업 3, (c) **`spec.md` frontmatter `module:`**. `design.md`는 자체 스코프에 한정된 별도 `module:` 값을 갖는 것이 정상이며 본 판정 대상이 아니다 |

---

## §E 품질 게이트

| 게이트 | 기준 |
|---|---|
| 테스트 | `go test ./...` green |
| 커버리지 | 회귀 없음(본 SPEC은 Go 코드 무변경이므로 baseline 유지) |
| 린트 | `golangci-lint run` green |
| SPEC 린트 | `moai spec lint` 본 SPEC 디렉터리 대상 0 errors |
| 템플릿 중립성 | CI guard green **AND** 스캐너가 `.js`를 실제로 읽음이 RED/GREEN 왕복으로 입증(공허 green 금지) |
| 미러 파리티 | 편집 쌍 전량 0-diff |
| 격리 불변식 | `TestSplitHarnessNamespaceNoLeak` green + 차단 유효성 확인 |

---

## §F Definition of Done

- [ ] MUST 등급 AC 전량 PASS, verbatim 명령 출력으로 근거 제시
- [ ] SHOULD 등급 AC 미충족 시 사유가 `progress.md`에 기록
- [ ] §B 시나리오 6개 전량 통과 (구 Group 1 철회로 write-concurrency 시나리오 2개 제거)
- [ ] 사용자 결정 D1 / D2 / D3 반영 확인 — D3는 **판별형** 게이트 grep 출력이 `progress.md`에 verbatim 기록(비판별형 12,133-match 형태 금지)
- [ ] 구 Group 1(write-concurrency) 철회 확인 — `agent-common-protocol.md` / `CLAUDE.md`의 write-concurrency 문장 `git diff` 0줄
- [ ] 후속 SPEC 이관 기록 확인 — `spec.md` §C + §G에 이관 범위와 사유 명시
- [ ] `spec.md` §D.2 라인 상한 전량 충족(실측 `wc -l` 인용) — `manager-spec.md`는 D3 게이트 분기에 대응하는 상한 적용
- [ ] Template-First 순서 준수 감사 통과
- [ ] HUMAN GATE 4종 보존 확인
- [ ] verdict 소유권 불변 확인
- [ ] `SPEC-DWF-CODEMAPS-PILOT-001` supersession 상호 참조 성립
- [ ] `progress.md` §E.2 / §E.3 run-phase 증거 기록
