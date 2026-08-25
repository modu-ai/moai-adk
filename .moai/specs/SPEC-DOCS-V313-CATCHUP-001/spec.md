---
id: SPEC-DOCS-V313-CATCHUP-001
title: "v3.1.3 release documentation catch-up — docs-site 4 locales + README 4 files"
version: "0.1.0"
status: draft
created: 2026-08-26
updated: 2026-08-26
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, i18n, v3.1.3, catchup"
tier: M
---

# SPEC: v3.1.3 문서 따라잡기 — CHANGELOG 26항목의 docs-site 4로케일·README 4파일 반영

## HISTORY

- 0.1.0 (2026-08-26): plan-phase 최초 작성. 카드 t274 (Class C, Tier M). baseline 트리 `e07a6d0f4` (worktree t274, branch WT-v313-docs)에서 CHANGELOG `[3.1.3]` 26항목 전수 조사 완료 — 검증된 격차 표는 §1. v3.1.3 릴리즈(#1602 계열, CHANGELOG 상 2026-08-24) 이후 문서가 코드를 따라가지 못한 격차를 닫는다.

## 1. 문제 — 측정된 형태 (검증된 격차 표)

CHANGELOG.md `## [3.1.3] - 2026-08-24` (177–306행)의 항목을 전수 추출했다: **Added 13 + Changed 4 + Fixed 9 = 26 항목** (`grep -n '^## \[' CHANGELOG.md` → `[3.1.3]` 177행, 다음 버전 `[3.1.2]` 307행; `[Unreleased]` 8행 항목은 범위 밖). 각 항목의 문서화 존재 여부를 docs-site 4로케일(`docs-site/content/{ko,en,ja,zh}/`, 파일 목록 체크섬 `98d2b226e6569dd7b07a8ce9ee4d3e5c` ×4 — 4로케일 파일 파리티 100%)과 README 4파일(`README.ko.md` ko-canonical + `README.md`/`README.ja.md`/`README.zh.md`, H2 12개 ×4)에 대해 grep으로 관측했다. 모든 관측은 이 워크트리(baseline `e07a6d0f4`)에서 이번 실행으로 얻은 것이다 (VCI §2).

판정 분류: **D** = 이미 문서화(작업 불필요) · **U** = 기존 페이지 갱신 · **N** = 신규 페이지 필요(구조 소관 — 별도 승인) · **NA** = 문서 표면 없음(결함 수정·내부 개선).

### 1.1 Added 13항목

| ID | 항목 | 판정 | 대상 · 근거(관측된 명령 결과) |
|----|------|------|-------------------------------|
| A1 | 루트 `AGENTS.md` standing contract (harness 공통 계약, 빌드 가드, 11 문서 stub화) | **N** | `grep -rln 'AGENTS\.md' docs-site/content` → 0파일; README 4파일 `grep -rn 'AGENTS\.md'` → 0행. 문서화 경로가 없음 — 신규 페이지 후보 |
| A2 | 11 에이전트 듀얼 게시 (`.codex/agents/moai/*.toml`, `agentemit` 결정적 생성) | **N** | `grep -rln '\.codex/agents' docs-site/content README…` → 0파일. A1과 동일 주제(codex dual-harness) |
| A3 | `.agents/skills` 스킬 미러 (codex-cli용, 사용자 저장소 밖) | **N** | `grep -rln '\.agents/skills' …` → 0파일. A1과 동일 주제 |
| A4 | `internal/codexadapter` — Codex 훅 어댑터 라이브러리 (11-이벤트 표, 아직 호출부 없음) | **N** | `grep -rln 'codexadapter\|hook adapter' …` → 0파일. A1과 동일 주제 — "아직 호출부 없음"이므로 문서 깊이는 승인 시 결정 |
| A5 | `/moai feedback` 스크러빙 계약 (`moai feedback scrub`/`queue` 동사, 취약점 분류기, 재시도 큐, `feedback.auto_submit`) | **U** | `moai-feedback.md` 95행에 auto_submit 게이트는 문서화됨(4로케일). 그러나 `grep -n 'scrub\|queue' ko/utility-commands/moai-feedback.md` → scrub/queue 동사·분류기·재시도 큐 0행 → 기존 페이지 갱신 |
| A6 | `workflow.todo.enabled` — 백로그 큐 끄기 스위치 (부재 시 on) | **U** | `grep -rln 'todo\.enabled' docs-site/content README…` → 0파일 → `advanced/config-sections.md`(+ `moai-todo.md` 크로스링크) 갱신 |
| A7 | `moai todo` 큐 자기 분석 (add 시 측정, `moai todo analyze`) | **D** | `grep -n 'analyze\|측정\|중복' ko/utility-commands/moai-todo.md` → 95·97·99·166·195·196행 — "기록만 남김", "정확 중복 거절", "Jaccard 0.80"까지 완전 문서화(4로케일 존재) |
| A8 | MCP 5도구 선택적 `project_root` (워크트리 감사·재시작 주의 포함) | **D** | `grep -n 'project_root' ko/guides/mcp-server.md` → 96–114행 — 6도구 나열, 워크트리 사례 표, "재시작 전까지 예전 동작" 주의까지 완전 문서화(4로케일 존재) |
| A9 | `moai-domain-svg-infographic` 커넥터 지오메트리 검사 (`SVG070`–`SVG074`) | **NA** | `grep -rln 'SVG06\|SVG07\|aria-labelledby' …` → 0파일. SVG 규칙 세부는 원래 docs-site 문서 표면 밖(skill 내부 린트 규칙) — v3.1.3 미반영이 아니라 문서화된 적이 없음 |
| A10 | SVG 산출물 접근성 이름 (`SVG060`–`SVG064`) | **NA** | A9와 동일 근거 |
| A11 | `moai-domain-design-dna` 다이어그램 프로파일 (`.design-dna/` 지속, mermaid/drawio 임포터) | **U** | `grep -n 'design-dna\|diagram' ko/advanced/skill-guide.md` → 159행 스킬 소개 행만 존재 — 프로파일·지속·임포터 미반영 → 기존 행 갱신. README 4파일 745행 design-dna 절도 동일 갱신 대상 |
| A12 | `moai update`/`moai init` 스킬 미러 symlink→copy 폴백 통지 | **U** | `grep -n 'symlink' ko/cli-reference/update.md ko/getting-started/init-wizard.md` → 0행 → `cli-reference/update.md`에 통지 한 줄 (경량) |
| A13 | `/moai gate` typecheck 축 (#1592) | **D** | `grep -n '…타입…' ko/utility-commands/moai-gate.md` → 7·10·23·64·88행 — "린트·포맷·타입 검사·테스트" 4축을 이미 서술. #1592로 코드가 문서를 따라잡은 형태 — 문서 갱신 불필요 |

### 1.2 Changed 4항목

| ID | 항목 | 판정 | 대상 · 근거 |
|----|------|------|-------------|
| C1 | 에이전트 model/effort 매트릭스 → judgment-weighted 정책 (심사·조율 행 high, `manager-spec`/`manager-develop` 전 열 medium, `manager-docs` sonnet/low, 어떤 행도 max 아님) | **U** | `grep -n 'max\|medium\|sonnet/low' ko/advanced/profile-matrix.md` → 27–35행 매트릭스가 구 정책(`manager-spec opus/high`, `manager-develop opus/max`, `super-advisor opus/max`, `manager-docs opus/medium`), 47행 "max 두 행 배정", 51행 max 별칭 서술 → `profile-matrix.md` 갱신. **`multi-llm/model-policy.md` 112–123행에 같은 매트릭스 중복 존재 — 두 페이지 모두 갱신** |
| C2 | `manager-lead` 매트릭스 합류 (기존 inherit sentinel) | **U** | `profile-matrix.md` 27–35행 표에 `manager-lead` 행 없음 → C1 갱신에 포함 |
| C3 | GLM reasoning-effort 상한 `max`로 상향 (low 제외 전 effort → reasoning max, 무설정 기본 max) | **U** | `grep -rn 'reasoning' docs-site/content/ko/multi-llm/` → 0행. `model-policy.md` 201행 "GLM effort 오버레이" 개념 언급만 있고 Claude effort→GLM reasoning 매핑·상한은 미서술 → `model-policy.md` 갱신 |
| C4 | `ANTHROPIC_*` 환경변수 나열 완전판 | **NA** | `grep -n 'ANTHROPIC_' ko/cli-reference/launchers.md` → 59행 (GLM 자격증명 주입 설명). 인식 키 목록 전체를 나열하는 사용자 문서 표면이 없음 — 내부 개선, 문서화 의무 없음 |

### 1.3 Fixed 9항목

| ID | 항목 | 판정 | 대상 · 근거 |
|----|------|------|-------------|
| F1 | 홈 디렉터리 하위 상태 조회가 `~/.moai/state`로 귀결되던 결함 (조회가 프로젝트 루트에서 멈춤) | **NA** | `moai-clean.md`는 `--home` allowlist 중심(208–249행) — 결함 수정이지 기능 변화 아님 |
| F2 | 관리 루트의 symlink가 `moai update`를 벽돌로 만들던 결함 | **NA** | 결함 수정 이력 — 문서 의무 아님 (A12 갱신 시 함께 언급할 수 있으나 의무 아님) |
| F3 | codex/GLM 감사 백엔드가 안 읽은 코드에 판정 내리던 결함 (diff 수집 실패 → `inconclusive`, 백엔드 미호출) | **U** | `grep -c 'inconclusive' ko/advanced/multi-model-audit.md` → `0`. 수렴 절차(57행)에 diff-수집-실패 케이스 없음 → `multi-model-audit.md` 수렴 절차 갱신 |
| F4 | 명시된 `inconclusive` 판정을 pass로 합성하던 결함 | **U** | F3과 동일 grep — `inconclusive` 0행 → F3 갱신에 포함 |
| F5 | `audit_multi` 수렴 판정이 감사된 트리 아래 기록되지 않던 결함 | **U** | F3/F4와 동일 페이지 — `mcp-server.md` 96–114행(project_root)은 이미 문서화돼 있으므로 `multi-model-audit.md` 수렴 절차에 한 절 추가 |
| F6 | `project_root` 심볼릭 링크 경계 우회 결함 (canonicalize) | **NA** | 내부 경계 검사 디테일 — A8의 `mcp-server.md` 워크트리 사례 표가 사용자 관점을 이미 다룸 |
| F7 | `moai init` 수집·폐기되던 답변의 실제 적용 (autonomy tier, 4 워크플로 토글, audit/codex 선택) | **D** | `init-wizard.md` 7행 "정한 값은 전부 YAML로 저장" + 44행 Page 3 서술 — fix 후 문서와 이미 일치 |
| F8 | 웹 콘솔 서버 기동 중 SIGTERM 즉사 결함 | **NA** | `grep -n 'SIGTERM\|signal' ko/advanced/moai-web-console.md` → 0행. 내부 결함 수정 — 사용자 가시 문서 표면 아님 |
| F9 | constitution의 `agent-authoring` 교차참조 복구 | **NA** | 내부 rules 참조 복구 — 사용자 문서와 무관 |

### 1.4 항목 외 — version SSOT 갭 (카드 지시 조사, t272 잔여 재확인)

26항목과 별개로, i18n 규칙 §7 release-sync 의무(“모든 버전 표시는 릴리즈 PR에서 함께 갱신”)가 v3.1.3에서 지켜지지 않았음을 관측했다:

| ID | 표면 | 관측 | 판정 |
|----|------|------|------|
| V1 | `docs-site/hugo.toml` 55–56행 | `version = "v3.1.2"`, `releaseDate = "2026-08-21"` (스테일) | **U** — v3.1.3 / 2026-08-24로 갱신 |
| V2 | README 4파일 491행 statusline 예시 | `🗿 v3.1.2` ×4 (스테일 표시) | **U** — v3.1.3으로 |
| V3 | README 4파일 766행 update-prompt 예시 | `🗿 v3.1.1 -> 🗿 v3.1.2` ×4 | **U** — 최신 릴리즈 기준 예시로 |

README 배지(24행)는 4파일 모두 `Release-v3.1.3`으로 이미 올바르다. **이 갭의 원인(릴리즈 프로세스가 hugo.toml·예시 표시를 동기화하지 않는 구조)은 별도 카드 권장 사항이며 이 SPEC이 흡수하지 않는다** — 이 SPEC은 증상(스테일 값)만 바로잡는다 (§6 Out of Scope 참조).

### 1.5 집계

- 26항목 = **D 4** (A7·A8·A13·F7) + **U 10** (A5·A6·A11·A12·C1·C2·C3·F3·F4·F5) + **N 4** (A1–A4, codex dual-harness 주제로 통합 가능) + **NA 8** (A9·A10·C4·F1·F2·F6·F8·F9)
- 항목 외: version SSOT 갭 3건 (V1–V3)
- 작업 대상: U 10항목 + V1–V3 (기존 페이지/파일 갱신) + N 4항목(신규 페이지 — 승인 관문)

## 2. 요구사항 (GEARS)

> GEARS 키워드(When/While/Where/shall/shall not)는 프로토콜 토큰으로 영문을 유지한다.

- **REQ-DVC-001** (Ubiquitous): Every documentation change produced under this SPEC shall land in all four locales — docs-site `content/{ko,en,ja,zh}` and README `{README.ko.md, README.md, README.ja.md, README.zh.md}` — in the same pull request. A canonical edit without its derived counterparts is a locale-parity failure.
- **REQ-DVC-002** (When): When a gap-table item classified **U** (§1) is documented, the harness shall update the existing docs-site page(s) and README section(s) it maps to — canonical ko authored first, en/ja/zh derived from it — without creating new pages.
- **REQ-DVC-003** (Where): Where a gap-table item classified **N** (§1: A1–A4, codex dual-harness) would require a new docs-site page, the harness shall not create the page, nor touch navigation config (per-locale `content/<locale>/_meta.yaml`, `data/menu/main.yaml`, `layouts/partials/menu.html` SVG cases), until the operator approves the new page explicitly; on denial the item shall be recorded as deferred to a separate card.
- **REQ-DVC-004** (While): While this SPEC is in run phase, the version SSOT surfaces shall read the v3.1.3 release values — `docs-site/hugo.toml` `params.version = "v3.1.3"` with `params.releaseDate = "2026-08-24"`, and the README in-example version displays (statusline `🗿 v…` line, update-prompt example) in all four README files — with historical citations ("introduced in vX.Y.Z"류) left untouched.
- **REQ-DVC-005** (While): While this SPEC is in run phase, the harness shall not modify any Go source under `internal/`/`pkg/`/`cmd/`, any template under `internal/template/templates/`, or any hook script — documentation surfaces only.
- **REQ-DVC-006** (When): When an item classified **NA** (§1) is left undocumented, the skip and its §1 rationale shall remain recorded in this SPEC's gap table — no entry of the 26 shall be dropped silently.
- **REQ-DVC-007** (When): When the run phase completes, the hns-oss-docs-verify recipe shall pass in full — warning-free hugo build, sitemap existence, URL-blacklist grep 0 hits, Mermaid LR/RL grep 0 hits, 4-locale file-existence and section parity, README 4-file heading parity, body-emoji scan 0 hits.
- **REQ-DVC-008** (When): When run-phase work begins, the §1 gap table's presence/absence cells shall be re-verified against the then-current tree — every cell re-observed by its named command, no carry-over from plan-phase observations (VCI §2 baseline-integrity attribution).

## 3. 제약

- ko canonical 체인: docs-site와 README 모두 ko가 정본, en/ja/zh는 파생 (i18n 규칙 §1). 번역본 안에서 정본 콘텐츠를 "고치지" 않는다 — 불일치는 격차 표에 기록한다.
- Mermaid는 TD/TB만 (`flowchart TD`/`graph TB`), LR/RL 금지. 본문 emoji 금지 — `{{</* icon <name> */>}}` shortcode 사용. 타이포그래픽 기호(→ ← ✓ ✗)와 orchestrator-banner 예시 코드블록 내 브랜딩 emoji는 보존.
- URL은 `adk.mo.ai.kr` 단독 허용 (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` 금지).
- 강조 마커 간격 규칙: `**단어** (Word)` — 괄호 설명은 마커 밖.
- docs-site 디자인 컴포넌트·`moai-brand.css`는 FROZEN — 이 SPEC은 콘텐츠만 만진다.
- 게시(commit/push/PR 병합)는 human-gated — run-phase는 편집까지만, 푸시는 sync-phase 관문.

## 4. Tier 분류

**Tier M.** 근거: 문서 전용이지만 4로케일 × (docs-site 페이지 6종 + README 4파일)의 동시 갱신, 항목별 매핑·파생 검증, 그리고 verify 레시피 7축 종료 게이트가 필요하다. Go 코드 신규·복잡 아키텍처 결정 없음 → Tier L 아님. 단일 파일 수정이 아니므로 Tier S 아님.

## 5. 변경 대상 파일 (전부 baseline `e07a6d0f4` 기준)

**U 항목 (기존 페이지 갱신 — 4로케일 각각):**
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` — A5 (scrub/queue 동사, 취약점 분류기, 재시도 큐)
- `docs-site/content/{ko,en,ja,zh}/advanced/config-sections.md` — A6 (`workflow.todo.enabled`)
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md` — A6 크로스링크 (선택)
- `docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md` — A11 (design-dna 행 갱신)
- `docs-site/content/{ko,en,ja,zh}/cli-reference/update.md` — A12 (symlink 폴백 통지)
- `docs-site/content/{ko,en,ja,zh}/advanced/profile-matrix.md` — C1·C2 (매트릭스 표 재작성, manager-lead 행)
- `docs-site/content/{ko,en,ja,zh}/multi-llm/model-policy.md` — C1 중복 매트릭스 표 + C3 (GLM reasoning 매핑)
- `docs-site/content/{ko,en,ja,zh}/advanced/multi-model-audit.md` — F3·F4·F5 (inconclusive 케이스, 판정 기록 트리)

**version SSOT:**
- `docs-site/hugo.toml` — V1 (`params.version`, `params.releaseDate`)
- `README.ko.md` + `README.md` + `README.ja.md` + `README.zh.md` — V2·V3 (491·766행 예시) 및 A11 (design-dna 절 745행)

**N 항목 (신규 페이지 — 승인 관문 후에만; 승인 시):**
- `docs-site/content/{ko,en,ja,zh}/advanced/` 하위 신규 페이지 (가칭 `codex-dual-harness.md`) + `content/<locale>/_meta.yaml` ×4 + `data/menu/main.yaml` + `layouts/partials/menu.html` — A1–A4. **승인 전에는 이 파일 어디도 손대지 않는다.**

## 6. Out of Scope

### Out of Scope — Go 코드·템플릿·훅 변경
- `internal/`, `pkg/`, `cmd/`의 모든 Go 소스, `internal/template/templates/` 하위 모든 파일, 훅 스크립트. 이 카드는 문서 전용(Class C docs card)이며 코드 표면은 REQ-DVC-005가 금지한다.
- 예외 없음 — 문서화 과정에서 발견한 코드 결함은 별도 카드로 보고된다.

### Out of Scope — 릴리즈 프로세스 개선 (version SSOT 재발 방지)
- V1–V3의 원인인 "릴리즈 PR이 hugo.toml·README 예시 버전을 동기화하지 않는" 프로세스 결함의 구조적 수정(릴리즈 체크리스트·검증 스크립트 추가 등). 이 SPEC은 스테일 값을 바로잡을 뿐이며, 재발 방지는 별도 카드 권장로 operator에게 보고된다 (t272 종결 시 지목된 t204 release workstream 귀속 가능성 포함).

### Out of Scope — CHANGELOG `[Unreleased]` 항목
- `CHANGELOG.md` 8행 `## [Unreleased]` 섹션의 모든 항목(2026-08-24 이후 작업). 이 SPEC의 기준은 `[3.1.3]` 확정 항목 26개뿐이다.

### Out of Scope — 신규 페이지 무승인 착수
- N 항목(A1–A4)의 페이지 생성·내비게이션 설정 변경은 REQ-DVC-003의 operator 승인 없이는 진행하지 않는다. 승인 대기 또는 거부 시, 해당 항목은 "deferred — separate card"로 격차 표에 기록되고 SPEC은 U 항목으로 완결된다.
