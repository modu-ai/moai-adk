---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Acceptance Criteria"
version: "0.12.0"
status: completed
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

### A.1 판정 명령 저작 규칙 (공허 GREEN·거짓 실패 방지)

아래 5개 규칙은 본 SPEC에서 **실제로 공허 GREEN(규칙 1·2·3·5) 또는 거짓 실패(규칙 1·4)를 만들어낸 원인**을 코드화한 것이다. 새 판정 명령을 쓰거나 기존 명령을 고칠 때 매번 적용한다. 공통 뿌리는 하나다 — 판정 명령이 **무엇을 실제로 읽고 무엇을 실제로 판정하는지** 확인하지 않은 것.

1. **표 셀에 `-E` 교대 금지.** 마크다운 표 셀에서 `|`는 셀 구분자라 `\|`로 escape해야 하는데, `grep -E`에서 `\|`는 교대가 아니라 **리터럴 `|` 문자**다. 두 제약이 충돌하므로 교대가 필요하면 코드블록(§D.2.1 / §D.5.1)에 두거나 `-e` 반복(`grep -E -e 'A' -e 'B'`)으로 파이프를 제거한다. `grep -c`/`-l`/`-r` 같은 **BRE** 모드에서만 표 셀의 `\|`가 정상 교대로 동작한다. 실측 피해: AC-APO-024b·050·061·071·071b가 **공허 GREEN**, AC-APO-023이 **부당 FAIL**.
2. **대상 파일의 언어와 패턴의 언어를 일치시킨다.** `.claude/skills/**` · `.claude/agents/**` · `.claude/rules/**`는 CLAUDE.md §9(*Commands, Agents, Skills Instructions: Always English*)에 따라 **영어 전용**이므로 한국어 패턴은 그 파일이 담을 수 **있는** 어떤 문자열과도 매치하지 못한다 — 결과가 항상 0이라 `== 0` 판정이 자동 통과한다. 실측 피해: AC-APO-024b.
3. **부정형 판정은 RED fixture로 증명한다.** `== 0` 판정은 "0이 나왔다"만으로 유효하지 않다. **금지 대상 문자열을 실제로 심은 사본에서 명령이 FAIL함**을 보이고, 원본에서 PASS함을 보이는 RED/GREEN 왕복이 있어야 근거로 인정한다(§E 품질 게이트의 중립성 항목과 동일 규범).
4. **가드의 정규식을 면제 없이 재구현하지 않는다.** 기계 가드(`TestTemplateNoInternalContentLeak` 등)는 **탐지 패턴과 면제 목록을 한 쌍으로** 소유한다. AC가 그 패턴만 베끼고 면제(`pedagogicalAllowlist` 등)를 빼면, 그 AC는 **거짓 실패만 생산할 수 있고 가드가 놓친 것은 하나도 잡지 못한다** — 특히 스코프까지 더 좁으면(변경 파일 ⊂ 전체 트리) 순수 손실이다. 판정은 **가드 실행을 권위로 삼고**(`go test -run <Guard>`), 보조 스캔은 *가드가 구조적으로 알 수 없는 클래스*(예: 본 SPEC 고유 토큰 계열, 가드가 CI에서 강제하지 않는 opt-in 티어)로만 좁혀 **조기 경보**로 표기한다. 보조 스캔은 변경 **라인**(`git diff -U0`) 스코프를 쓴다 — 파일 스코프는 건드린 파일의 **선행** 토큰까지 끌어와 거짓 실패를 만든다. 실측 피해: AC-APO-061이 allowlist 등재된 교육용 예시 2건에 대해 곧 거짓 실패할 예정이었고, AC-APO-071은 가드 S2의 `requireHexLetter` 정련을 빠뜨려 십진 상수를 오탐하는 형태였다. **보조 스캔을 붙이기 전에 "가드가 이미 덮는가"를 클래스 단위로 실측 대조하고, 덮는다면 붙이지 않는다** — 거짓 양성만 내는 중복 검사는 검사가 없는 것보다 나쁘다.
5. **이식성을 확인한다.** 판정 명령은 BSD `grep`(macOS `/usr/bin/grep`)과 `ugrep` 양쪽에서 **같은 값**을 내야 한다. `\b`(word boundary)를 한 패턴에 **여러 번** 쓰면 `ugrep`이 조용히 0을 반환하거나 정지하는 사례를 실측했다 — 그 자체가 또 하나의 공허 GREEN 경로다. `\b` 대신 POSIX 문자클래스(`[^a-z]` 등)를 쓴다.

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
| AC-APO-020 | 020 | MUST | **파일 앵커 2항 동시 충족** — 판정 명령 `CMD-020`(§D.2.1). (a) 병렬 렌즈 구조 heading 실재 == 1(baseline 1 — `plan/spec-assembly.md:172`), (b) verdict 소유자를 **이름으로 지명한** 문장 실재 == 1(baseline 1 — `:176`). (b)가 소유자명(`plan-auditor`)까지 포함해야 하는 이유: `single binding PASS/FAIL` 같은 **소유자 없는 토큰**만 세면 "verdict가 렌즈로 넘어갔다"는 **정반대 문장에도 통과**한다(토큰 존재 ≠ 요구사항 성립). **prose-judged 잔여(반사적 강화 금지)**: 렌즈 서술의 *적정성*(렌즈가 실제 read-only인지, fan-out 폭이 3-5인지)은 본 명령이 판정하지 않는다 — fan-out 규범은 AC-APO-028/030이, 적정성은 리뷰어 독해가 담당한다 |
| AC-APO-021 | 021 | MUST | 판정 명령 `CMD-021`(§D.2.1) == 1 — `plan/spec-assembly.md`에 `single writer, single-turn parallel Write` 구문 실재(baseline 1 — `:69`). **파일 앵커 필수**(재귀형 금지). 본 AC는 *그 구문이 존재할 것* 자체가 요구사항이므로 구문 앵커가 곧 요구사항 판정이며 대리 토큰 검사가 아니다. 산출물 **개수** 일치는 본 AC가 아니라 AC-APO-046 소관 |
| AC-APO-022 | 022 | SHOULD | **파일 앵커 필수** — 판정 명령 `CMD-022`(§D.2.1) 2항 동시 충족: (a) RED-stage drafter pool heading == 1(baseline 1 — `run/task-decomposition.md:72`), (b) 동일 파일 내 단일 적용자 서술 ≥ 1(baseline 1 — `:74` `remains the only writer`). 종전 판정은 트리 전역 존재형이라 `plan.md` M2 작업 3이 지정한 표면(`run/task-decomposition.md`)이 **아닌 어느 파일에서 매치돼도 GREEN**이었다(AC-APO-012/014가 실측한 이설·무관인용 해저드와 동형). 등급 SHOULD 유지 — 본 수정은 판정 정밀화이지 등급 변경이 아니다 |
| AC-APO-023 | 023 | MUST | run Phase 13/16/17 축약 구조 서술 존재 **AND** 판정 명령 `CMD-023`(§D.2.1) == 2 (baseline 2 — **Phase 16**(Active Quality Evaluation) `Maximum 3 fix-evaluate cycles` @:190, **Phase 17**(TRUST 5 Static Verification) `Maximum 3 review iterations` @:230. Phase 13은 반복 상한을 갖지 않는다. 둘 중 하나라도 삭제되면 1로 떨어져 FAIL). **구 명령 폐기 사유(실측)**: 종전 표 셀 형태 `grep -cE 'Maximum 3 (fix-evaluate cycles\|review iterations)'`는 `grep -E`에서 `\|`가 교대가 아니라 **리터럴 `|` 문자**이므로 `Maximum 3 fix-evaluate cycles|review iterations`라는 단일 연접 문자열을 찾았고, 현행 트리에서 **0**을 반환해 정상 구현을 **부당 FAIL**시켰다(정상 교대형은 2 반환). 파이프 없는 `-e` 반복형으로 교체 — §D.2.1 서두의 표-셀 파이프 금지 규칙 참조 |
| AC-APO-024 | 024 | MUST | 판정 명령 `CMD-024`(§D.2.1) 3항 동시 충족: (a) drafter 표 행 수 == 5 (D1..D5 — **구조 카운트**이므로 토큰 존재 검사가 아니다; drafter가 4개로 줄거나 6개로 늘면 즉시 FAIL), (b) 단일 적용자 서술 ≥ 1 (`single writer of every final artifact`), (c) **disjoint-writer 변형 서술 == 0** (`disjoint`, 대소문자 무시, baseline 0). (c)는 사용자 결정 D2(disjoint-writer 불채택)의 **유일한 기계 가드**다 |
| AC-APO-024b | 024b | MUST | **2항 동시 충족.** (1) Phase 12 서술이 **현행 write-concurrency 규칙과 독립**임이 명시(현행 `[HARD]` 절대 금지형 규칙을 그대로 둔 채 성립) — baseline 충족(`sync/doc-execution.md:138`). (2) 규칙 완화의 진행 여부를 전제로 삼는 서술 0건 — 판정 명령 `CMD-024b`(§D.2.1) 4항 전량 충족(2a 총 언급 == 1 / 2b 독립선언 줄 밖 언급 == 0 / 2c 의존마커→규칙어 근접 == 0 / 2d 규칙어→의존마커 근접 == 0). **구 명령 폐기 사유 2건(각각 독립 실측)**: (i) **언어 불일치로 공허** — 구 패턴은 한국어(`규칙 완화가 전제`)를 찾았으나 대상 `doc-execution.md`는 CLAUDE.md §9(*Commands, Agents, Skills Instructions: Always English*)에 따라 **영어 전용**이라 그 패턴은 이 파일이 담을 수 **있는** 어떤 문자열과도 매치 불가였다. 금지 대상 영어 문장 7종을 심은 fixture에서 구 명령은 **0**을 반환(공허 GREEN 실측). (ii) **`\|` 리터럴 파이프** — `grep -E`에서 `\|`는 교대가 아니라 **리터럴 `|` 문자**이므로 구 패턴은 두 절이 한 줄에 연접해야 하는 단일 패턴이었다. **한국어** 금지 문장 `규칙 완화가 전제이다`에 대해서조차 구 명령은 0, 정상 교대형(`|`)은 1을 반환해 (i)과 무관하게 독립 확인됐다. **잔여 취약점(공개 — clean 아님)**: `CMD-024b`는 `write-concurrency`를 명명하지 **않으면서** 마커 어휘(`prerequisite`/`pending`/`depends on` 등) **밖**의 표현을 쓰는 의존 주장(예: `cannot ship until the rule is relaxed`)을 잡지 못한다(fixture R13이 4항 전부 통과함을 실측). 따라서 (2)는 *부재의 증명이 아니라 회귀 tripwire*이며, 1차 통제는 (1)의 적극적 독립 선언 + 리뷰어 독해다 |
| AC-APO-025 | 025 | SHOULD | **파일 앵커 필수** — 판정 명령 `CMD-025`(§D.2.1) ≥ 1: `sync/quality-gates-quality.md`에 `Per-package fan-out` 구문 실재(baseline 1 — `:280`, Phase 10 Step 0.7.3 Test Generation). 종전 판정은 트리 전역 존재형이라 표면 앵커가 없어 배선 이설 시 공허해진다. 등급 SHOULD 유지 |
| AC-APO-026 | 026 | MUST | sync Phase 1/7에 `moai verify` 스냅샷 소비 서술 + 신선도(키 일치/TTL) 조건 동반 |
| AC-APO-027 | 027 | SHOULD | **3개 표면 파일별 판정** — 판정 명령 `CMD-027`(§D.2.1). `plan.md` M2 작업 8이 지정한 표면은 run Phase 9 / run Phase 18 / sync Phase 9 **3곳**이므로 각각 ≥ 1: (a) `run/phase-execution.md`(Phase 9 Pre-Implementation MX Context Scan, baseline 1 — `:397`), (b) `run/task-decomposition.md`(Phase 18 MX Tag Update, baseline 1 — `:260`), (c) `sync/quality-gates-quality.md`(Phase 9 Step 0.6.2 Scan Modified Files, baseline 1 — `:182`). **재귀형 `≥1` 및 재귀형 `== 3` 모두 금지** — 전자는 3곳 중 1곳만 있어도, 후자는 3건이 한 파일에 몰려도 통과한다(AC-APO-012가 실측한 동형 해저드). 판정에 `read-only` 동반을 요구하는 이유: 샤딩이 write fan-out으로 변질되면 REQ-APO-028 위반이므로 "샤딩됨"만으로는 요구사항이 성립하지 않는다. 등급 SHOULD 유지 |
| AC-APO-028 | 028 | MUST | 판정 명령 `CMD-028`(§D.2.1) 2항 동시 충족. (a) **부정형 — nesting 의존 서술 == 0**: `subagent nesting` / `nested subagent` 언급 중 부정어(`not`/`never`/`rather than`/`instead of`/`without`)의 지배를 받지 **않는** 것이 0건(baseline 0 — 현행 8건 전량이 `scaling, not subagent nesting` 형태). 단순 `grep -c 'subagent nesting' == 0`은 **정상 문장까지 FAIL**시키므로 사용 불가다. (b) **긍정형 — 파일별 커버리지**: fan-out 도입 5개 파일 각각 오케스트레이터-launch 서술 ≥ 기대 건수(spec-assembly 1 / task-decomposition 3 / phase-execution 1 / doc-execution 1 / quality-gates-quality 2). **부분 기계화 명시(과신 금지)**: (b)는 *파일별* 카운트이지 *사이트별* 카운트가 아니다 — 한 파일에 fan-out 2곳이 있고 그중 1곳만 규범을 달아도 (b)는 통과한다. REQ-APO-028의 "모든" 전칭은 신규 fan-out 블록 추가 시 **리뷰어 독해**로 보증하며, 명령은 기존 규범의 회귀 탐지용 tripwire다 |
| AC-APO-029 | 029 | MUST | 게이트 토큰 4종(`Decision Point 1` / `Implementation Kickoff Approval` / `gate-sync-1` / `gate-sync-2`) 모두 편집 후에도 존재 |
| AC-APO-030 | 030 | MUST | 판정 명령 `CMD-030`(§D.2.1) 2항 동시 충족. (a) **부정형 — fan-out 서술 줄의 `AskUserQuestion` 지시 == 0**(baseline 0). 열거자(`read-only` **AND** `Agent()`를 동시에 포함하는 줄, 현행 9줄)가 다소 넓은 것은 **안전한 방향**이다 — `== 0` 부정형 판정에서 열거자가 넓으면 판정이 더 **엄격**해질 뿐 느슨해지지 않는다(반대로 긍정형 "모든 사이트" 판정에서 넓은 열거자는 오탐을 만든다). 파일 전역 `grep -c AskUserQuestion == 0`은 오케스트레이터 게이트의 **정당한** 언급(예: `plan/spec-assembly.md` 19건)까지 잡으므로 성립 불가. (b) **긍정형 — 파일별 blocker report 규범 커버리지** ≥ 기대 건수(spec-assembly 1 / task-decomposition 3 / phase-execution 1 / doc-execution 1 / quality-gates-quality 2). **부분 기계화 명시**: (b)는 AC-APO-028(b)와 동일하게 파일별이지 사이트별이 아니다 |

#### D.2.1 Group 2 판정 명령 블록

[HARD] **교대(alternation)가 필요한 판정 명령은 표 셀에 두지 않는다.** 마크다운 표 셀에서 `|`는 셀 구분자라 `\|`로 escape해야 하는데, `grep -E`에서 `\|`는 교대가 아니라 **리터럴 `|` 문자**다. 이 이중 제약이 AC-APO-024b(구판)를 **공허 GREEN**으로, AC-APO-023(구판)을 **부당 FAIL**로 만든 직접 원인이다. 따라서 교대가 필요하면 (a) 아래 코드블록에 두거나 (b) `-e` 반복(`grep -E -e 'A' -e 'B'`)으로 파이프를 **아예 제거**한다. 표 셀 안 `\|`는 `grep -c`/`grep -l`/`grep -r` 같은 **BRE** 모드에서만 정상 교대로 동작한다(§D.4/§D.5의 AC-APO-049/054/070/076이 그 경우).

**이식성**: 아래 결과는 BSD `grep`(macOS `/usr/bin/grep`)과 `ugrep` **양쪽에서 동일**함을 실측했다. `\b`(word boundary)를 **한 패턴에 여러 번** 쓰면 `ugrep`이 조용히 0을 반환하거나 정지하는 사례를 실측했으므로 — 그 자체가 또 하나의 공허 GREEN 경로다 — 아래 패턴은 `\b`를 쓰지 않고 POSIX 문자클래스만 사용한다. 모든 명령은 리포 루트에서 실행하며 기재된 baseline은 M2 완료 시점(`HEAD`) 실측이다. `grep -c`는 0건일 때 **exit status 1**을 반환하므로 종료 코드가 아니라 **출력 숫자**로 판정한다.

```bash
W=.claude/skills/moai/workflows

# CMD-020  (a)==1  (b)==1
grep -c '^#### Parallel Review Lenses' $W/plan/spec-assembly.md
grep -cE 'binding PASS/FAIL stays with .plan-auditor.' $W/plan/spec-assembly.md

# CMD-021  ==1
grep -c 'single writer, single-turn parallel Write' $W/plan/spec-assembly.md

# CMD-022  (a)==1  (b)>=1
grep -c '^### RED-stage Drafter Pool' $W/run/task-decomposition.md
grep -c 'remains the only writer' $W/run/task-decomposition.md

# CMD-023  ==2   (구판의 표-셀 `\|` 리터럴 파이프 버그를 -e 반복으로 제거)
grep -cE -e 'Maximum 3 fix-evaluate cycles' -e 'Maximum 3 review iterations' \
  $W/run/task-decomposition.md

# CMD-024  (a)==5  (b)>=1  (c)==0
grep -cE '^\| D[1-5] \|' $W/sync/doc-execution.md
grep -c  'single writer of every final artifact' $W/sync/doc-execution.md
grep -ci 'disjoint' $W/sync/doc-execution.md

# CMD-024b  (2a)==1  (2b)==0  (2c)==0  (2d)==0
grep -ciE 'write.?concurrency' $W/sync/doc-execution.md
grep -inE 'write.?concurrency' $W/sync/doc-execution.md \
  | grep -civ 'Independent of the write-concurrency rule'
grep -ciE '(prerequisite|precondition|blocked on|pending|awaiting|contingent|depends on|dependent on|conditional on|predicated on)[^.]{0,80}(rule|relax|revis|amend|loosen|concurren)' \
  $W/sync/doc-execution.md
grep -ciE '(rule|relax|revis|amend|loosen|concurren)[^.]{0,80}(prerequisite|precondition|blocked on|pending|awaiting|contingent|depends on|dependent on|conditional on|predicated on)' \
  $W/sync/doc-execution.md

# CMD-025  >=1
grep -c 'Per-package fan-out' $W/sync/quality-gates-quality.md

# CMD-027  (a)(b)(c) 각각 >=1  — 파일별 판정(재귀형 금지)
grep -ciE -e 'shard.*read-only' -e 'read-only.*shard' $W/run/phase-execution.md
grep -ciE -e 'shard.*read-only' -e 'read-only.*shard' $W/run/task-decomposition.md
grep -ciE -e 'shard.*read-only' -e 'read-only.*shard' $W/sync/quality-gates-quality.md

# CMD-028 (a)  ==0  — 부정어의 지배를 받지 않는 nesting 언급
grep -oinE -e '.{0,30}subagent nesting' -e '.{0,30}nested subagent' \
  $W/plan/spec-assembly.md $W/run/task-decomposition.md $W/run/phase-execution.md \
  $W/sync/doc-execution.md $W/sync/quality-gates-quality.md $W/sync/quality-gates-context.md \
  | grep -civE -e 'not[^a-z]' -e 'never[^a-z]' -e 'rather than' -e 'instead of' -e 'without'

# CMD-028 (b)  파일별 >= 1 / 3 / 1 / 1 / 2
for f in plan/spec-assembly.md run/task-decomposition.md run/phase-execution.md \
         sync/doc-execution.md sync/quality-gates-quality.md; do
  printf '%s %s\n' "$f" "$(grep -c 'scaling, not subagent nesting' "$W/$f")"
done

# CMD-030 (a)  ==0  — fan-out 서술 줄에 AskUserQuestion 지시가 없을 것
grep -nE 'read-only' \
  $W/plan/spec-assembly.md $W/run/task-decomposition.md $W/run/phase-execution.md \
  $W/sync/doc-execution.md $W/sync/quality-gates-quality.md $W/sync/quality-gates-context.md \
  | grep -E 'Agent\(\)' | grep -cE 'AskUserQuestion'

# CMD-030 (b)  파일별 >= 1 / 3 / 1 / 1 / 2
for f in plan/spec-assembly.md run/task-decomposition.md run/phase-execution.md \
         sync/doc-execution.md sync/quality-gates-quality.md; do
  printf '%s %s\n' "$f" "$(grep -c 'blocker report' "$W/$f")"
done
```

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
| AC-APO-050 | 050 | MUST | **본문(body) 스코프 판정** — frontmatter 블록 제외: `B=$(awk 'NR>1 && /^---$/{f=1;next} f' .claude/agents/moai/manager-docs.md); grep -ciE -e 'nextra' -e 'wcag' -e 'page.?speed' -e 'lighthouse' <<< "$B"` == 0 (본문 기준 baseline **7**). 판정 근거는 **stdout 수치**이며 종료코드가 아니다 — `grep -c`는 0건일 때 exit 1을 반환한다. **구 명령 폐기 사유 2건(실측)**: (i) 최초형 `grep -ciE "nextra\|wcag\|page.?speed\|lighthouse"`는 `-E` 모드에서 `\|`가 교대가 아니라 리터럴 파이프라 `nextra|wcag|page.?speed|lighthouse`라는 **단일 연접 문자열**을 찾았고 **0을 반환 → 공허 GREEN**이었다. 파이프 없는 `-e` 반복형으로 교체(§D.2.1 서두 규칙). (ii) 그 `-e` 반복형은 **파일 전역 스코프**였고, 이 형태는 **AC-APO-065와 상호 배타**라 원리적으로 충족 불가였다 — M3 본문 다이어트 완료 시점(`8f0426f4b`) 전역 실측은 **1**이며 유일 잔존 매치는 `manager-docs.md:6`의 frontmatter `description:` 행(`... API docs, Nextra, technical writing ...`)이다. 전역형으로 0에 도달하려면 AC-APO-065가 **무변경으로 고정한 frontmatter 블록**을 수정해야 하므로 두 MUST는 동시에 만족될 수 없었다. REQ-APO-050이 요구하는 대상은 **본문의 docs-플랫폼 어휘**이므로 스코프를 본문으로 재한정한다 — 본문 기준 **7 → 0**으로 REQ는 이미 충족되어 있다. 향후 이 명령을 전역 스코프로 "복원"해서는 안 된다(AC-APO-065 충돌 재발). **중립성 가드와 무관함(혼동 방지)**: 내부 유출 가드는 `internal/template/templates/` 만 스캔하고 이 토큰들에 대응 클래스도 없으므로 AC-APO-061/071과 달리 **allowlist 인지 처리 대상이 아니다** |
| AC-APO-051 | 051 | MUST | `e2e-tester.md`에 비-호스트 OS 레시피 본문 부재 **AND** `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` 실재(`test -f`) **AND** `e2e-tester.md`가 해당 경로를 참조 **AND** 템플릿 미러 `internal/template/templates/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` 존재 + 0-diff |
| AC-APO-052 | 052 | MUST | **서식 관용(formatting-tolerant) 판정**: `grep -cE 'squash.{0,6}merge.{0,6}rebase' .claude/agents/moai/manager-git.md` == 1 (baseline 3 — `origin/main` L126 주석 / L163 auto-merge / L191 manual). **구 명령 폐기 사유(실측)**: 종전 리터럴형 `grep -c 'squash \| merge \| rebase'`는 **서식 무관용**이라 백틱 재서식에 깨진다 — `8f0426f4b` 실측 **0**, `origin/main` 실측 **3**. HEAD의 잔존 해석 규칙은 L32에서 각 토큰이 백틱으로 감싸인 형태(`` `squash` `` / `` `merge` `` / `` `rebase` ``)로 재서식되어 리터럴 "공백-파이프-공백" 패턴과 불일치한다. 즉 **0은 규칙이 사라졌다는 뜻이 아니라 명령이 서식 변화를 못 따라간 것**이며, 실질은 REQ-APO-052가 요구한 **해석 규칙 3 → 1 축약이 그대로 달성**되어 있다. 백틱 유무를 흡수하는 `.{0,6}` 관용 정규식으로 교체하면 HEAD **1** / baseline **3** 으로 정상 판정된다(양 리비전 실측 확인). 리터럴 파이프를 정규식에서 제거해 마크다운 표 셀 파손·렌더 불일치 위험도 함께 해소한다. **보존 확인(REQ-APO-068)**: baseline L163·L191의 두 운용 경로는 HEAD에서 각각 Late-Branch Phase C 인라인 머지(L114 `gh pr merge <PR> --squash --delete-branch`)와 PR Auto-Merge 4단계(L170 `gh pr merge --<merge_method> --delete-branch`)로 **존치**하며, `gh pr merge` 총 3건은 양 리비전 동일 — 위반 없음 |
| AC-APO-053 | 053 | MUST | `builder-harness.md`의 `Model/effort escalation` 중복 문장 1개 이하, model-policy 표 재진술 0건 |
| AC-APO-054 | 054 | MUST | `grep -l "verification-batch-pattern\|Parallel Execution" .claude/agents/moai/*.md` 결과 파일 수 ≥ 8 (현재 1) |
| AC-APO-055 | 055 | MUST | `wc -l .claude/agents/moai/*.md` 합계와 파일별 값이 **모두** `spec.md` §D.2 표의 적용 분기 상한 이하. 합계 상한은 분기 조건부 — **분기 A ≤ 1997** / **분기 B ≤ 2017**(`manager-spec.md` 230 → 250 차이). 적용 분기는 AC-APO-043 게이트 결과가 결정하며, 분기 B 적용은 "MUST 미달성 + 사유 기록"이 아니라 **정상 적용**이다. **합계 상한은 개별 상한의 산술 합이며 파생값이다** — 개별 상한이 §D.2.1 절차로 재보정되면 합계도 함께 재계산하고, 감축률 주장(§D.2 각주)도 갱신한다. **판정 시 주의**: 상한 충족이 **공백 전용 압축**으로 달성되지 않았음을 확인한다(`spec.md` §D.2.1 [HARD] 조항) — `git diff` 상 제거된 행이 실질 내용 없이 빈 줄뿐인 구간은 이 AC의 충족 근거로 인정하지 않는다 |

### D.4 Group 4 — 불변식 (REQ-APO-060..068)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-060 | 060 | MUST | 편집된 미러 쌍 전량 `diff` 결과가 Pre-flight baseline 차이 외 0 |
| AC-APO-061 | 061 | MUST | **2항 동시 충족 — 가드가 권위, 보조는 조기 경보**(§D.5.1). (a) `CMD-061(a)` **PASS**: `TestTemplateNoInternalContentLeak`이 C1-C8/S1-S3 클래스 **와 `pedagogicalAllowlist` 면제를 함께** 소유하므로 *무엇이 leak인지*의 SSOT다. (b) `CMD-061(b)` == 0: 본 SPEC 고유 토큰(`SPEC-AGENT-PARALLEL-OPT-001` / `REQ-APO-` / `AC-APO-`)을 **변경된 라인**(`git diff -U0`)에서만 스캔. **구 명령 전면 폐기 사유 2건**: (i) `-E` + `\|` 리터럴 파이프로 **공허 GREEN**이었다(정상 교대형 실측 2). (ii) 더 심각하게 — 구 명령은 **가드의 C1/C2 정규식을 면제 목록 없이 재구현**하면서 스코프는 오히려 좁았다(변경 파일 ⊂ 전체 트리). 이 조합은 **거짓 실패만 생산할 수 있고 가드가 놓친 것은 하나도 잡지 못한다**(§A.1 규칙 5). 실증: 정상 교대형이 잡은 2건은 `templates/.claude/agents/moai/manager-spec.md`(`:160`,`:175`)의 `SPEC-V3R6-SPEC-ID-VALIDATION-001`인데, 이는 가드 `pedagogicalAllowlist`에 *"Demonstrates SPEC ID regex validation pre-write self-check pattern"* 사유로 **등재된 정당한 교육용 예시**이고 `origin/main`에 이미 존재한다(가드 PASS 실측). M3 대상이 `manager-spec.md`이므로 이 거짓 실패는 곧 발생할 예정이었다. (b)가 additive인 근거는 §D.5.1 참조 — 가드는 `SPEC-AGENT-PARALLEL-OPT-001`에 대응 클래스가 **없고**, `REQ-APO-`를 잡는 S3는 `skillBodyScoped`라 본 SPEC 대상인 agents/workflows에서 **미발화**한다 |
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
| AC-APO-070 | 070 | MUST | 4개 동시 충족: (a) 본 SPEC frontmatter `partially_supersedes: [SPEC-DWF-CODEMAPS-PILOT-001]`, (b) `spec.md`가 superseded AC를 **ID로 인용** — `AC-DCP-010` [REF](`acceptance.md:79` / `progress.md:86`) + 그 소유 요구사항 `REQ-DCP-009/010`, (c) 정식 grep 문구 `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing 이 더 이상 성립하지 않음을 명시, (d) 파일럿 SPEC 아티팩트에 supersession 주석 추가로 상호 참조 성립 |
| AC-APO-071 | 071 | MUST | **2항 동시 충족 — 가드가 권위**(§D.5.1). (a) `CMD-071(a)`: 가드 **PASS** **AND** 스캐너가 `.js`를 실제로 읽음(`leakTextExtensions`에 `".js"` 등재 ≥ 1 — M1이 추가). (b) `CMD-071(b)` == 0: 본 SPEC 고유 토큰만 3개 배포 스크립트에서 스캔(해당 파일들은 `.claude/workflows/` 라 S3 `skillBodyScoped`가 미발화하므로 이 구간은 실재하는 가드 사각지대다). **손수 쓴 정규식 전면 폐기 사유 2건**: (i) `-E` + `\|` 리터럴 파이프로 공허했다(실측 0 — 다만 정상 교대형으로도 0이라 잠복 상태였다). (ii) 더 중요하게 — 구 정규식은 가드 **C1/C2/S2를 면제 없이 재구현**한 것이라 `.js`가 `leakTextExtensions`에 등재된 지금 **가드에 완전히 포섭**되며, 게다가 S2의 `requireHexLetter` 정련을 빠뜨려 **가드가 올바르게 제외하는 십진 상수까지 오탐**한다(실증: `const maxBytes = 10485760;` → 구 형태 1 매치, 가드 S2는 제외). 일반형 `SPEC-[A-Z0-9-]+-[0-9]{3}` 제외 방침(`spec.md` §F.8.3-a)은 이제 가드가 직접 소유한다. **선행 조건**: AC-APO-072 PASS — (a)의 `.js` 등재 확인이 그 왕복의 정적 대응물이다 |
| AC-APO-071b | 071 | SHOULD | **존치 — 가드 중복 여부 실측 검증 완료.** 본 AC가 덮는 3클래스는 AC-APO-071과 달리 가드에 포섭되지 **않는다**: (α) 날짜 `20[0-9]{2}-[0-9]{2}-[0-9]{2}` — 가드 `S1-internal-date`가 유사 패턴을 갖지만 `strictLeakClasses` 소속이고 그 티어는 `MOAI_TEMPLATE_LEAK_STRICT=1` **opt-in**이다(`.github/` · `Makefile` 전수 grep 결과 **어디에도 미설정** → CI 미강제). (β) `/Users/` — **대응 클래스가 아예 없다**. (γ) SHA `[0-9a-f]{9,40}` — 가드 S2는 `{7,8}` + 후행 구두점/EOL 요구라 9자 이상 연속 hex와 **겹치지 않는다**. 따라서 본 AC는 "가드 정규식 재구현"이 아니라 **가드가 CI에서 강제하지 않는 잔여 구간의 유일한 커버리지**이며, 이것이 REQ-APO-071의 5개 금지 클래스를 온전히 채우는 부분이다. **manual-only / CI-unenforced 표기 의무**(`spec.md` §F.8.3-a 귀결 3): "CI green"을 이 3클래스의 근거로 인용하지 않는다. 판정 **2항 동시 충족**: (i) 배포된 3개 파일에 대해 판정 명령 `CMD-071b`(§D.5.1) **== 0**(구 표-셀 형태는 `-E` + `\|` 리터럴 파이프로 공허했다 — 정상 교대형 실측도 0이라 판정 결과는 불변이나 잠복 공허성은 제거) — REQ-APO-071이 금지한 5개 클래스 전량이 MUST 수준으로 커버되도록 하는 조항이며, SHOULD 등급이라는 이유로 면제되지 **않는다**. (ii) 그 결과가 `progress.md`에 **CI-unenforced 라벨과 함께** 기록됨 — 기록 의무는 "CI green"을 이 3개 클래스의 근거로 오인용하지 못하게 하는 장치다 |
| AC-APO-072 | 072 | MUST | `leakTextExtensions`에 `".js": true` 존재 (`grep -n '".js"' internal/template/internal_content_leak_test.go` ≥ 1) **AND** 시나리오 8의 RED/GREEN 왕복이 관측됨 — 미중립 스크립트 심었을 때 FAIL, 중립화 후 PASS. **왕복 직접 관측 완료(2026-07-25, 격리 worktree)**: probe `redgreen-probe.js`(C1 토큰 `SPEC-V3R6-REDGREEN-PROBE`)를 `internal/template/templates/.claude/workflows/`에 심고 실행 → `TestTemplateNoInternalContentLeak` 이 `class=C1-spec-id-prefix` 로 **FAIL**, probe 제거 후 **PASS**(전후 baseline 모두 `ok`). C1 클래스는 always-on(strict 티어 아님)이고 `.js`는 `leakTextExtensions`에 등재돼 있어 RED가 공허하지 않음을 저작 전 판독으로 확인했다. 이로써 close 당시의 PASS-WITH-DEBT 유예 사유("템플릿 트리 변형이 필요해 read-only 감사 범위 밖")는 **무효**다 |
| AC-APO-072b | 062/069 | MUST | `TestSplitHarnessNamespaceNoLeak` PASS **AND** 차단 유효성 확인: `hns-release-update-run.js`를 템플릿에 심고 실행 시 `SPLIT_HARNESS_NAMESPACE_LEAK`으로 FAIL (심은 파일은 제거). **왕복 직접 관측 완료(2026-07-25, 격리 worktree)**: probe `.claude/workflows/hns-release-update-run.js`를 템플릿 트리에 심은 상태에서 `TestSplitHarnessNamespaceNoLeak` 이 sentinel `SPLIT_HARNESS_NAMESPACE_LEAK` 과 dev-only 배포 금지 메시지를 내며 **FAIL**, probe 제거 후 **PASS**. probe 파일명은 `splitHarnessAgentPrefixes`(`hns-release-update` 포함)를 저작 전 판독해 선택했으므로 RED가 공허하지 않다. 이로써 close 당시의 PASS-WITH-DEBT 유예 사유("템플릿 트리 변형이 필요해 read-only 감사 범위 밖")는 **무효**다 |
| AC-APO-073 | 073 | MUST | `internal/cli/update/plan/plan.go`의 user-owned 판정이 3개 generic 스크립트에 대해 false 반환 — 접두사 `hns-`/`harness-` 미매치로 확인. 보존 목록 소스 무변경(`git diff` 0줄) |

#### D.5.1 Group 4/5 중립성 판정 명령 블록

**[HARD] 무엇이 leak인지는 가드가 권위다.** `TestTemplateNoInternalContentLeak`은 클래스 집합(C1-C8 / S1-S3)과 **면제 목록(`pedagogicalAllowlist`)을 함께** 소유한다. 판정 명령이 가드의 정규식만 베끼고 면제를 빼면 **거짓 실패만 생산할 수 있고 가드가 놓친 것은 하나도 잡지 못한다**(§A.1 규칙 5). 따라서 아래 (a)는 가드 실행이며, (b)는 **가드가 구조적으로 알 수 없는 클래스**로만 좁힌 보조 스캔 — **조기 경보이지 verdict가 아니다**. 새 교육용 예시를 추가하는 유지보수자는 가드 allowlist에 등재하면 되고, (b)가 그것을 막지 않는다.

**보조 스캔이 additive임의 근거(가드 클래스 전수 대조 실측)**:

- `SPEC-AGENT-PARALLEL-OPT-001` → **어떤 가드 클래스에도 매치되지 않음**(C1은 `V3R[2-6]|AGENCY|WORKTREE`만, C1c는 `DB-SYNC-RELOC|PROJECT-DB-HINT`만 열거).
- `REQ-APO-` / `AC-APO-` → `S3-req-ac-token-any-prefix`가 매치하나 **`skillBodyScoped: true`** 라 `.claude/skills/` 하위에서만 발화한다. C2의 도메인 열거(`ATR|WO|COORD|UNP|LNC|TII`)에도 `APO`는 없다.
- 본 SPEC의 템플릿 대상은 `.claude/agents/`(M3)와 `.claude/workflows/`(M1) — **둘 다 비-skill 표면**이라 S3가 발화하지 않는다. 따라서 (b)가 덮는 구간은 실재한다.

```bash
# CMD-061 (a) — 권위 판정. PASS 필수.
go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1

# CMD-061 (b) — 보조(조기 경보). 변경된 **라인**만 스코프. ==0
#   파일 스코프(--name-only)를 쓰면 본 SPEC이 건드린 파일의 **선행** 토큰까지 걸려
#   거짓 실패가 난다(실측: 파일 스코프 2 / 라인 스코프 0 — 그 2건은 가드 allowlist에
#   등재된 manager-spec.md 교육용 regex 예시다).
git diff -U0 origin/main...HEAD -- internal/template/templates/ \
  | grep '^+' | grep -v '^+++' \
  | grep -cE -e 'SPEC-AGENT-PARALLEL-OPT-001' -e 'REQ-APO-[0-9]' -e 'AC-APO-[0-9]'

# CMD-071 (a) — 권위 판정 + 스캐너가 .js를 실제로 읽는지(공허 green 차단).
go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
grep -c '"\.js":' internal/template/internal_content_leak_test.go   # >=1 (leakTextExtensions 등재)

# CMD-071 (b) — 보조. 3개 배포 스크립트는 .claude/workflows/ 라 S3 미발화. ==0
grep -lE -e 'SPEC-AGENT-PARALLEL-OPT-001' -e 'REQ-APO-[0-9]' -e 'AC-APO-[0-9]' \
  internal/template/templates/.claude/workflows/*.js | wc -l

# CMD-071b — CI 미강제 3클래스 수동 점검. ==0
#   S1(날짜)·S2(sha 7-8자)는 strictLeakClasses 소속이고 그 티어는
#   MOAI_TEMPLATE_LEAK_STRICT=1 opt-in이다(.github/ · Makefile 전수 grep 결과 **미설정**).
#   `/Users/`는 대응 클래스가 아예 없고, sha {9,40}은 S2의 {7,8}과 겹치지 않는다.
grep -nE '20[0-9]{2}-[0-9]{2}-[0-9]{2}|/Users/|[0-9a-f]{9,40}' \
  internal/template/templates/.claude/workflows/*.js | wc -l
```

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
