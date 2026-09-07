---
id: SPEC-DOCS-CODEX-WIRING-CALLOUT-001
title: "docs-site moai doctor 페이지에 Codex Wiring 진단 콜아웃 4로케일 반영"
version: "0.2.0"
status: completed
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.1.4"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, codex, doctor"
tier: M
---

# SPEC: docs-site `moai doctor` 페이지 — Codex Wiring 진단 콜아웃 4로케일 반영

## HISTORY

- 0.1.0 (2026-09-08): plan-phase 최초 작성. 카드 t535 (Class C, Tier M — 원인은 이미 확립돼 있다). 워크트리 t535 (branch `WT-docs-codex-callout`, baseline `a849d99d2` = origin/develop tip)에서 이번 실행으로 격차를 재측정해 §1에 고정했다 (VCI §2). 원인은 이미 확립돼 있다 — `moai doctor` 의 "Codex Wiring" 진단이 v3.1.4 문서화 콜아웃 관례 시대에 배포됐는데 docs-site 에 콜아웃 절이 없다. 작업은 docs-site 4로케일 `cli-reference/doctor.md` 에 H2 절을 하나씩 추가하는 것뿐이다.
- 0.2.0 (2026-09-08): plan-audit iter1 **PASS 0.86** (Tier M 기준 0.80, 반복 1/3 — `.moai/reports/t535/plan-audit-iter1.md`). 차단 D1 + 선택 D2–D4 적용. D1: acceptance §D.1 추적 맵에 `AC-DWC-002→REQ-DWC-014` 보완 + `REQ-DWC-015→AC-DWC-013` 간접 적용 선언. D2: plan §A.2 에 zh 헤딩 `诊断` 용어 근거 명시(ko 정본 한자 대응 + 본문 9회 관측). D3: REQ-DWC-010 에 progress.md 라이프사이클 기록 예외 명시(AC-DWC-013 정합). D4: acceptance §A 에 공유 RED-now 셀 규약 명시. 오탈자 2건(`moa doctor`→`moai doctor`, Class 라벨)과 가드 거부형 루프 명령 2곳의 평문 분해는 iter1 이전 커밋(`80401bece`)에서 이미 적용됨.

## 1. 문제 — 측정된 형태

아래 관측은 전부 이 워크트리(baseline `a849d99d2`)에서 이번 실행으로 수행한 것이다.

### 1.1 결함 (RED-now)

`/usr/bin/grep -rn 'Codex Wiring' docs-site/` → **0 히트, exit 1**. docs-site 어디에도 Codex Wiring 진단이 문서화돼 있지 않다.

주의: 이 리포의 셸 `grep` 은 ugrep 래퍼로 `-I`(바이너리 추정)·`--ignore-files`(gitignore) 조건을 조용히 적용해 부재 판정을 왜곡한다. 부재 재검증은 반드시 `/usr/bin/grep` 으로 한다 (요구사항 REQ-DWC-014·AC-DWC-001에 규약 명시).

### 1.2 대조군 (양성 바늘 — 같은 범위)

`Home Disk Usage` 는 4로케일 doctor.md 전부에서 히트한다 (이번 실행 관측: ko·en·ja·zh 각 2히트). 즉 파일 자체는 grep-도달 가능하고 콜아웃 관례도 살아 있으며, 빠진 것은 Codex Wiring 하나뿐이다.

### 1.3 콜아웃 관례 (닮을 대상)

`docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md` 는 각각 정확히 7개 H2 절을 가진다 (4로케일 패리티 — 이번 실행 관측 `grep -c '^## '` → 각 7). ko 기준 절 순서:

```
11: ## 개요
17: ## 플래그
26: ## 하위 명령어
39: ## Home Disk Usage 진단 {{< new-badge v3.1.1 >}}
55: ## Hook Delivery 진단 {{< new-badge v3.1.4 >}}
67: ## 종료 코드
78: ## 예시
```

새로 배포되는 검사는 배지 달린 H2 절로 문서화하는 관례가 이미 두 절(Home Disk Usage, Hook Delivery)에서 성립돼 있다. 특히 Hook Delivery 절이 이미 `v3.1.4` 배지를 쓰고 있어, 이번 절이 이 페이지의 두 번째 `v3.1.4` 배지가 된다.

### 1.4 배치 대상 — 진단의 실제 동작 (internal/cli/doctor_codex.go 관측)

"Codex Wiring" 진단(`internal/cli/doctor_codex.go` 173행 `DiagnosticCheck{Name: "Codex Wiring"}`)은 조언형(advisory)·fail-open·읽기 전용이며 스스로 고치지 않는다. 검사 항목과 코드에 박힌 수정 지시문:

| 대상 | 관측 (코드 상수·행) | 지시문 |
|------|---------------------|--------|
| `.codex/hooks.json` 존재 + 키 화이트리스트 | 키 하나라도 어긋나면 codex가 파일 전체를 조용히 무시 | (조언형 보고) |
| sidecar 해시 발산 | `reTrustAdvice` (46행): `"run codex /hooks to re-trust the changed hooks"` | `codex /hooks` 재신뢰 |
| moai 바이너리 PATH | 생성된 훅 명령이 PATH 해석 없이는 발화 불가 | (조언형 보고) |
| `.codex/config.toml` `[mcp_servers.moai]` 테이블 | 사용자 소유 — 보고만, 수리 안 함 | (조언형 보고) |
| `.agents/skills` 스킬 미러 | 부재/댕글링 → `mirrorRedeployAdvice` (79행) | `run moai update --templates-only --force --yes` |
| 미배선 프로젝트 | `initCodexAdvice` (52행): `"run moai init --agent codex"` | `moai init --agent codex` |
| 사용자측 `~/.codex/config.toml`(`$CODEX_HOME/config.toml`) `[[skills.config]]` 스테일 경로 | 939행: `"remove the stale entries or restore the skill files"` — enabled 상태 4분할 조언형 | 스테일 항목 제거 또는 스킬 파일 복원 |
| **`enabled` 키 형태 불량** | 798행 `codexSeverityFatal` — **이 검사의 유일한 fatal 등급** | 각 항목에 `enabled = true` 또는 `enabled = false` (bare TOML 불리언) 명시 |

두 가지 측정된 미묘함을 SPEC 이 구현자에게 그대로 전달한다:

1. **`moai clean --codex-skills` 는 doctor 의 지시문이 아니다.** 이 명령은 실재한다(`internal/cli/clean.go:42,70` — "ghost" `[[skills.config]]` 등록 수거)지만 `internal/cli/doctor_codex.go` 는 `codex-skills` 문자열을 전혀 담지 않는다 (이번 실행 `grep -n 'codex-skills' internal/cli/doctor_codex.go` → 0히트). 콜아웃은 스테일 항목 조언을 코드 문구 그대로 적고, 유령 등록 수거는 `moai clean --codex-skills` **별도 동사**로만 언급해야 한다 — doctor 가 그 명령을 가리킨다고 쓰면 허위 기술이다.
2. **fatal 의 버전 범위.** codex-cli 0.153.4 에서 `enabled` 누락·비불리언이면 모든 호출이 exit 1 (SPEC-CODEX-ENABLED-FATAL-001 실측). 코드 주석(791행)이 "observed on that release only"로 한정하므로 콜아웃도 그 한정을 유지한다. fatal 외의 모든 발견은 조언형이며 doctor 종료 코드 0을 유지한다.

claude-only 미배선 프로젝트(코덱스 없는 머신)는 조용한 정보성 스킵 — "un-nagging" 불변. 절반 배선(에이전트 정의는 있고 배선 파일 없음)은 명명된 상태(SPEC-CODEX-PARTIAL-WIRING-001)다.

### 1.5 버전 귀속 — 왜 v3.1.3 이 아니라 v3.1.4 인가 (구현자가 "고쳐서는 안 되는" 근거)

`CHANGELOG.md` 의 `## [Unreleased]` 는 8–574행이고 마지막 릴리스 절은 `## [3.1.3] - 2026-08-24` (575행)다 — 이번 실행 관측. Codex Wiring 검사를 배포한 SPEC-CODEX-WIRING-001 (CHANGELOG 239행)과 그 확장들(PARTIAL-WIRING t499, MIRROR-DOCTOR t498, ENABLED-FATAL t508, STALE-SPLIT-FOURTH t534, GHOST-SKILLS-PRUNE t506, SKILL-DISABLE t502)이 전부 Unreleased 안에 있다. 즉 이 검사는 **v3.1.4 에 처음 배포된다**.

`new-badge` shortcode(`docs-site/layouts/shortcodes/new-badge.html`)는 버전 문자열을 자유 형식 인자로 받는다. 배지는 반드시 `{{< new-badge v3.1.4 >}}` 다. **`v3.1.3` 으로 바꾸면 허위 귀속** — 검사가 [3.1.3] 릴리스에 존재하지 않았다. v3.1.3 시점 문서만 읽은 구현자가 정렬을 이유로 배지를 3.1.3으로 "수정"하는 사고를 이 절이 차단한다.

### 1.6 범위 판정 — 이 검사만 빠져 있다

Codex Wiring은 콜아웃 관례 성립 시대(v3.1.1+)에 배포된 검사 중 이 콜아웃이 없는 **유일한** 검사다 (이번 실행 확인 — §1.1 부재 + §1.2 대조군). doctor 자체는 v3.0.0-rc11 (SPEC-DOCTOR-PROMOTION-001)에 태어났고 그 이전 검사들은 관례 이전 산물이다. binary-lag SPEC들(SPEC-BINLAG-INVOCATION-001 t366, SPEC-BINARY-LAG-VISIBILITY-001 t326)은 doctor 에 새 검사를 추가하지 않았다(t326 명시: 비교는 이미 `moai doctor` Binary Freshness 행 안에 있었다). 관례 이전 검사들의 문서화는 더 큰 별도 작업이며 이 SPEC 의 범위 밖이다(Out of Scope).

### 1.7 갭이 catch-up SPEC 을 살아남은 이유

SPEC-DOCS-V313-CATCHUP-001 (카드 t274, status: completed 2026-08-26)이 CHANGELOG `[3.1.3]` 절까지만 문서를 따라잡았고 **[Unreleased] 를 명시적으로 범위 밖**으로 두었다 — Codex Wiring 검사가 정확히 그 위치에 있다. 완료된 catch-up SPEC 이 있었음에도 갭이 남은 구조적 원인이며, 재발 방지를 위해 카드 t538(보류 중인 v3.1.3+ 문서 catch-up)이 이 SPEC 착지 후 이 콜아웃을 "이미 반영"으로 다뤄야 한다(§5 조정 노트).

## 2. 요구사항 (GEARS)

> GEARS 키워드(When/While/Where/shall/shall not)는 프로토콜 토큰으로 영문을 유지한다.

- **REQ-DWC-001** (Ubiquitous): Every content change produced under this SPEC shall land in all four locales — `docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md` — in the same change set. ko is canonical, en is derived from ko, ja/zh are derived in parallel; a partial-locale landing (ko만, 또는 일부 로케일 누락) is prohibited.
- **REQ-DWC-002** (Ubiquitous): Each of the four doctor.md files shall carry the new section as exactly one H2 with the callout convention — ko `## Codex Wiring 진단 {{< new-badge v3.1.4 >}}`, en `## Codex Wiring check {{< new-badge v3.1.4 >}}`, ja·zh 동등 형태(배지는 verbatim) — raising each file's H2 count from 7 to 8.
- **REQ-DWC-003** (While): While this SPEC is in run phase, the new section shall be placed adjacent to the existing check callout sections — immediately after the "Hook Delivery 진단|check" section and before the "종료 코드"/"Exit codes" section — in all four locales.
- **REQ-DWC-004** (Ubiquitous): Every behavioral claim in the new section shall be supported by `internal/cli/doctor_codex.go` as observed in §1.4. The four repair directives shall read code-faithfully: `run moai init --agent codex`, `run codex /hooks to re-trust the changed hooks`, `run moai update --templates-only --force --yes`, and the stale-entry advice `remove the stale entries or restore the skill files`. The section shall not attribute `moai clean --codex-skills` to the doctor's advice; ghost collection may be mentioned only as a separate `moai clean --codex-skills` verb (internal/cli/clean.go).
- **REQ-DWC-005** (Ubiquitous): The section shall describe the fatal grade exactly: the unusable `enabled` key is the ONE fatal finding (codex 0.153.4 measured, per SPEC-CODEX-ENABLED-FATAL-001 — 코드 주석의 "observed on that release only" 한정 유지); the correct shape is `enabled = true|false` as a bare TOML boolean on every `[[skills.config]]` entry; all other findings are advisory/fail-open and keep doctor's exit code 0.
- **REQ-DWC-006** (When): When the diagnosed project is claude-only and unwired on a codex-less machine, the section shall describe the silent informational skip (the "un-nagging" invariant) — the check nags nobody on a codex-less setup.
- **REQ-DWC-007** (Ubiquitous): The section shall cross-link to the per-locale Codex Dual Harness page using the locale-prefixed absolute path form (`/ko/advanced/codex-dual-harness`, `/en/advanced/codex-dual-harness`, `/ja/advanced/codex-dual-harness`, `/zh/advanced/codex-dual-harness`) — 대상 페이지는 4로케일 모두 실재(이번 실행 관측).
- **REQ-DWC-008 (shall not)**: The badge for this section shall not read `v3.1.3` — the check ships first in v3.1.4 (§1.5 귀속 근거 유지). Rewriting the badge to match a v3.1.3-era reading of the docs is a false attribution and a review-blocking defect.
- **REQ-DWC-009 (While)**: While this SPEC is in run phase, docs-site conventions shall hold: no decorative body emoji (`{{< icon ... >}}` shortcode만 허용); Mermaid는 도입 시 TD-only; URLs는 `adk.mo.ai.kr` 허용 리스트만; 강조 표기 간격 규칙(`**단어** (Word)` — 괄호는 마커 밖); facts·figures·commands·code blocks는 로케일 간 verbatim 보존.
- **REQ-DWC-010 (While)**: While this SPEC is in run phase, the harness shall not modify any Go source under `internal/`/`pkg/`/`cmd/`, any template under `internal/template/templates/`, any `docs-site/layouts/`/`shortcodes/` file, any sidebar/menu/nav file (`_meta.yaml`, `data/menu/main.yaml`, `menu.html` — doctor 페이지는 이미 메뉴에 있다), or any other docs page. The write surface is exactly the four doctor.md files (본 SPEC 디렉터리의 progress.md 등 라이프사이클 진행 기록은 콘텐츠 쓰기 표면이 아니므로 이 제한 밖이다 — AC-DWC-013 과의 정합, plan-audit iter1 D3).
- **REQ-DWC-011 (When)**: When a behavioral claim in the section cannot be traced to `internal/cli/doctor_codex.go` (or `internal/cli/clean.go` for the ghost-collection verb), the claim shall be dropped or reworded to its observed citation — no invented behavior enters user documentation.
- **REQ-DWC-012 (When)**: When run-phase work begins, the §1.1 absence cell shall be re-verified against the then-current tree with `/usr/bin/grep -rn 'Codex Wiring' docs-site/` — no carry-over from plan-phase observations (VCI §2).
- **REQ-DWC-013 (When)**: When the run phase completes, the hns-oss-docs-verify axes shall report zero NEW violations — (1) warning-free hugo build, (2) 4-locale section-count parity (doctor.md H2 = 8 ×4), (3) URL-blacklist grep 0 hits, (4) Mermaid LR/RL 0 hits, (5) body-emoji scan 0 new hits. Axes already green at the baseline stay green.
- **REQ-DWC-014 (Ubiquitous)**: Every absence-type verification in this SPEC shall use `/usr/bin/grep` (이 리포의 셸 `grep` 은 ugrep 래퍼로 gitignore·바이너리 추정을 조용히 적용해 부재 판정을 무효화한다), and every AC's starting cell carries the RED-now observation from §1.
- **REQ-DWC-015 (When)**: When the run encounters any surface belonging to the reserved card t538 (docs-site v3.1.3+ catch-up) beyond this callout, the harness shall not absorb it — it records the finding in progress.md and leaves t538's scope intact.

## 3. 제약

- ko canonical 체인 (i18n 규칙 §17.3): ko 작성 → en 파생 → ja/zh 병렬 파생. 파생본에서 정본 콘텐츠를 고치지 않는다 — 불일치는 ko 쪽을 고쳐 재파생.
- 4로케일 동일 변경 집합 (§17.3 동시 업데이트 의무). 이 문서 규모에서는 `translation_status: pending` 유예(5,000단어 이상) 대상이 아니다.
- 명령·버전·TOML 표기는 로케일 간 번역하지 않는다: `moai init --agent codex`, `codex /hooks`, `moai update --templates-only --force --yes`, `moai clean --codex-skills`, `enabled = true`, `0.153.4`.
- 게시(commit/push/PR)는 human-gated — run-phase는 편집까지만.
- 이 SPEC 은 docs-site 콘텐츠만 만진다 — `moai-brand.css`·디자인 컴포넌트 FROZEN.

## 4. Tier 분류

Tier M — 산출물 3종(spec/plan/acceptance), 쓰기 표면 4파일, LOC 영향 ~80–160행(로케일당 ~20–40행), Go 코드 0줄. REQ 15 / AC 14 — Tier M 상한(16/16) 이내.

## 5. 조정 노트 — 카드 t538 (범위 흡수 금지)

예약 카드 t538 = docs-site v3.1.3+ 4로케일 catch-up (현재 보류). 이 카드가 콜아웃을 착지시키면 t538의 measure-first 전수 조사는 doctor.md Codex Wiring 콜아웃을 **이미 반영(D)**으로 판정해야 한다. 반대 방향은 금지다: 이 SPEC 은 t538의 다른 어떤 표면도 흡수하지 않는다(REQ-DWC-015). run-phase에서 t538 소관 표면을 발견하면 progress.md에 기록만 한다.

## Out of Scope

### Out of Scope — 관례 이전 doctor 검사들의 문서화

- doctor 가 v3.0.0-rc11 (SPEC-DOCTOR-PROMOTION-001) 때부터 가진 기존 검사들(Constitution Registry, Binary Freshness, Home Disk Usage 이전 항목 등)의 콜아웃·문서화 보강은 더 큰 별도 작업이다 — 이 SPEC 은 Codex Wiring 한 절만 추가한다 (§1.6 범위 판정).

### Out of Scope — README·CHANGELOG·내부 코드

- README* 는 `moai doctor` 를 3회 언급하지만 per-check 콜아웃 표면이 없다 — 관측만 기록하고 편집하지 않는다.
- CHANGELOG.md 는 sync-phase 소관이다 — run-phase가 만지지 않는다.
- `internal/**` Go 변경, `docs-site/layouts/**`·`shortcodes/**` 변경, 사이드바/메뉴/네비 파일 변경 — 전부 금지 (REQ-DWC-010).

### Out of Scope — t538 소관 표면

- t538(보류 중인 v3.1.3+ docs catch-up)의 다른 어떤 페이지·갭도 이 SPEC 이 흡수하지 않는다 — doctor.md Codex Wiring 콜아웃 하나가 이 SPEC 의 전부다.

## 관련 SPEC

- 선행(문서 대상): SPEC-CODEX-WIRING-001, SPEC-CODEX-PARTIAL-WIRING-001, SPEC-CODEX-MIRROR-DOCTOR-001, SPEC-CODEX-ENABLED-FATAL-001, SPEC-CODEX-STALE-SPLIT-FOURTH-001, SPEC-CODEX-GHOST-SKILLS-PRUNE-001, SPEC-CODEX-SKILL-DISABLE-001
- 방법론 모델: SPEC-DOCS-V313-CATCHUP-001 (같은 도메인, 같은 구조 — §1.7)
- 관례 이력: SPEC-DOCTOR-PROMOTION-001 (doctor 탄생), SPEC-BINLAG-INVOCATION-001 / SPEC-BINARY-LAG-VISIBILITY-001 (새 검사 아님 — §1.6)
