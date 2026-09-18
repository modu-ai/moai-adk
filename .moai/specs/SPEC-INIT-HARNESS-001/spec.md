---
id: SPEC-INIT-HARNESS-001
title: "moai init 하니스 3-way 배포 — agent_wiring 선택이 배포를 지배한다"
version: "0.1.0"
status: in-progress
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/cli/wizard, internal/template, internal/core/project, internal/cli/doctor"
lifecycle: spec-anchored
tags: "init, wizard, harness, codex, deployer-filter, skill-mirror, llm-harness, disclosure, t585"
tier: L
depends_on: [SPEC-INIT-QUIET-WIZARD-001]
related_specs: [SPEC-INIT-QUIET-WIZARD-001, SPEC-INIT-HARNESS-PROMPT-001, SPEC-CODEX-WIRING-001, SPEC-CODEX-COMMAND-SKILLS-001, SPEC-AUT-PERMMODES-001]
---

## HISTORY

- 2026-09-14 — v0.1.0 초안 (manager-spec, 카드 t585 plan 단계). 근거: `.moai/reports/init-tui-audit-20260909.md` §운영자 결정 반영 + §2차 결정 반영(읽기전용 참조), 선행 SPEC `.moai/specs/SPEC-INIT-QUIET-WIZARD-001/`(completed, 2026-09-12 착지), 이 트리 전수 재측정(research.md §0 — 워크트리 HEAD `a404132e7`, 트리 `aebfa5bea`). 카드 전제 13건 중 2건 반증(commandemit 착지, 결속표 부재)·1건 명칭 드리프트(`--agent` → `--llm`) 확인 — research.md §1 표.

---

## §1 문제

`moai init`의 하니스 질문(`agent_wiring`)과 `--llm` 플래그는 오늘 **배포 범위를 지배하지 않는다**. 배포기(`internal/template/deployer.go:160-284`)는 임베디드 템플릿 FS를 무조건 전부 순회·기록하므로, `--llm codex`로 초기화해도 `.claude/` 전체·`CLAUDE.md`·`.mcp.json`이 심어진다. codex 선택의 실제 효과는 (1) `.mcp.json` 런타임 프로비저닝 생략, (2) `.codex/` 배선(`internal/codexwiring` — hooks.json·config.toml·신뢰 사이드카)뿐이다. 사용자가 "Codex"를 골랐는데 Claude 전용 트리가 통째로 깔리는 것은 질문이 약속하는 것과 배포가 하는 것이 갈라진 상태다.

반대 방향의 갈라짐도 있다: claude·both 선택의 결과는 이미 사실상 동일하고(둘 다 전면 배포 + codex both일 때 배선 추가), 하니스 3값이 배포 수준에서 의미 있는 차이를 만들지 않는다.

이 SPEC은 2026-09-09 init/update 전수 조사(`.moai/reports/init-tui-audit-20260909.md`)의 운영자 결정 2("하네스 질문 추가 — claude 단독=현행 / both=`.claude`+codex 전체 / codex 단독=AGENTS.md+codex만")와 최종 카드 분할 C3("하네스 3-way 배포")를 실행한다.

### §1.1 2026-09-09 조사 이후 이미 착지한 것 (이 SPEC이 새로 하지 않는 것)

조사 시점 대비 이 트리(`a404132e7`)에서 아래가 착지돼 있어, C3의 "질문 신설"과 "both 보강"은 대부분 기존 자산 확인으로 흡수된다:

| 조사의 C3 전제 | 이 트리 실측 | 판정 |
|---|---|---|
| 하니스 질문 신설 필요 | `agent_wiring` 질문 존재 — `internal/cli/wizard/questions.go:375-387`, 3옵션 claude/codex/both, `--llm` 플래그 우선(SPEC-INIT-HARNESS-PROMPT-001) | 이미 존재 |
| both = commandemit 16종 착지 필요 | `internal/template/commandemit/` 착지(`e7d2a1658`, t503) + 게시 스킬 16종 실파일(`internal/template/templates/.agents/skills/`) | 이미 착지 |
| both = AGENTS.md 보강 | 템플릿 CLAUDE.md:9 `@AGENTS.md` 임포트 + 중립 결속표(`templates/AGENTS.md:21-25`, 3행) + 훅 6/11 공개(`:268-271`) | 이미 착지(t523 포함) |
| codex 단독 4축 (필터·캐노닉·정합·플래그 판정) | 미착지 — **이 SPEC의 본체** | 신규 |

## §2 배경 계약

- 선행 SPEC-INIT-QUIET-WIZARD-001(completed): init 위저드는 정확히 4문항(`conversation_language`, `user_name`, `agent_wiring`, `autonomy_tier`) — 이 순서. 제거 14문항은 기본값 위임. reconfigure(`moai update -c`) 12문항 불변(D1). 이 SPEC은 위저드 **문항 수와 순서를 바꾸지 않는다** — `agent_wiring`의 설명문·옵션 설명만 배포 결과를 서술하도록 고친다. 형제 카드 t753의 목표 형상(4문항)과 정합.
- SPEC-INIT-HARNESS-PROMPT-001(completed): `--llm` 플래그(claude|codex|both, 기본 claude)와 위저드 답변의 단일 해석 지점 `resolveAgentWiringWithWizard`(`internal/cli/init.go:737-739`), 플래그 > 위저드 우선순위. AC-CW-004("플래그 부재 ≡ `--llm claude` — `.codex/` 배선 없음")는 이 SPEC에서도 유지된다.
- SPEC-CODEX-WIRING-001(completed): codex 배선 3종(`.codex/hooks.json`, `.codex/config.toml`, 신뢰 사이드카 — `internal/codexwiring/codexwiring.go:29-38`), `moai tool enable codex`(기존 프로젝트 codex 추가), update 시 배선 리프레시(REQ-CW-009).
- SPEC-CODEX-COMMAND-SKILLS-001(completed): 16개 `/moai` 커맨드의 codex 게시 스킬(`.agents/skills/moai-<command>/SKILL.md`) — 템플릿에 실파일로 존재, update 모드에서 사용자 파일 보호(R-011).
- 스킬 미러(`internal/template/skill_mirror.go`): 카탈로그 원본은 `.claude/skills/`(34개), 미러는 `.agents/skills/<name>` → `../../.claude/skills/<name>` **상대 심볼릭 링크**(`MirrorLinkTarget` :174-176). `//go:embed`가 심볼릭 링크를 조용히 누락하므로(:11-17) 원본은 실파일, 링크는 배포 시 생성이다. **원본이 없는 배포에서 링크는 전부 끊긴다** — codex 단독 설계의 핵심 구속 조건.

## §3 요구사항 (GEARS)

> 동결 목록(클로드 전용 런타임 기능 — 조사 §표면 소속 지도 :73 기준, 본문에서 "claude 전용 기능 동결 목록"이라 칭한다): 질문 채널(AskUserQuestion), Agent 소환, 스킬 로더(deferred-m1), Task 도구군, output style, 슬래시 명령, Workflow 스크립트, 훅 6/11 이벤트(Notification·PostToolUseFailure·TeammateIdle·TaskCompleted 등 Claude 전용 5종 미발화 포함).

- **REQ-IH-001** (Ubiquitous): The init harness resolver shall resolve the harness selection exactly once, from the `--llm` flag over the wizard answer, onto the closed set {claude, codex, both} with default claude.
  - 현행 `resolveAgentWiringWithWizard`(`internal/cli/init.go:737-739`)·`normalizeAgentWiring`(:159-166)의 단일 해석 지점과 닫힌 집합을 유지한다. 해석 지점의 신규 소비자는 배포기 선택(REQ-IH-005)과 영속화(REQ-IH-002)다.

- **REQ-IH-002** (Ubiquitous): The init flow shall persist the resolved harness value to `llm.yaml` key `llm.harness` on every init run — claude 포함 3값 모두 기록.
  - 기본값 claude도 명시 기록한다(암묵 부재 < 명시 기록 — doctor/update가 부재 키를 "claude 취급"으로 추론하지 않게). `internal/config/defaults.go`에 claude 시딩을 추가하는 것은 M1 산물이다. update의 3-way 병합(RestoreMoaiConfigRetained)이 이 키의 생존 경로다 — M4가 AC-IH-015로 증명한다.

- **REQ-IH-003** (While harness = claude): The deployer shall keep the claude-only deployment file-set and deployment logic unchanged — 신규 필터·재매핑 없음. Content updates mandated by REQ-IH-002 (`llm.harness` 기록) and REQ-IH-008 (AGENTS.md 공개 보강) are explicitly excluded from this invariance claim.
  - 현행 claude 배포가 심는 표면 집합(`.claude/` 전체 + `AGENTS.md` + `.codex/agents/*.toml`(11개) + `.agents/skills/` 게시 16종 + 미러 링크 + `.mcp.json`)은 그대로 유지된다 — 이 요구가 보존하는 것은 "무엇을 심는가"(파일집)와 "어떤 로직으로 심는가"이지, 심어지는 파일의 내용 불변이 아니다. 조사 운영자 결정("claude 단독=현행")에 따른 것으로, 디스패치 문구 "`.claude/` 표면만"과의 불일치는 research.md §1 · plan.md NC-2 기록.

- **REQ-IH-004** (While harness = both): The deployer shall produce today's claude surfaces plus the Codex wiring artifacts, with `.mcp.json` provisioning forced on — 현행 유지.
  - codex 배선 3종은 `wireCodexUnlessClaude`(`internal/cli/init.go:191-198` → `codexwiring.Wire`)가 이미 수행한다. 이 SPEC은 both 배포에 아무것도 추가하지 않고 고정 테스트로 보존한다(AC-IH-006).

- **REQ-IH-005** (While harness = codex): The deployer shall deploy only the universal and Codex surfaces — 프로젝트 루트에 `.claude/` 아래 경로가 하나도 생기지 않으며, `CLAUDE.md`, `.mcp.json`, `.claudeignore`, `.moai/status_line.sh`도 쓰지 않는다.
  - 배포 대상(보편+codex): `AGENTS.md`, `.codex/**`(템플릿 TOML 11개 + 배선 3종), `.agents/skills/**`(REQ-IH-006), `.moai/**`(status_line.sh 제외 — config sections 등 보편), `.gitignore`, `.git_hooks/`, `.github/`, `.worktreeinclude`. 배포 제외(claude 전용): `.claude/**`, `CLAUDE.md`, `.mcp.json`, `.claudeignore`, `.moai/status_line.sh(.tmpl)`. 정본 표는 design.md §3.
  - `.mcp.json` 제외 근거: codex는 `.codex/config.toml`이 MCP 설정 표면이고 현행 규칙도 codex에 `.mcp.json` 프로비저닝을 건너뛴다(`internal/cli/init.go:980-987`).

- **REQ-IH-006** (While harness = codex): The deployment shall materialize the skill catalog canonically at `.agents/skills/` as real directories — `.claude/skills`를 가리키는 심볼릭 링크를 만들지 않으며, 카탈로그 원본(`.claude/skills/**`, 34개)의 내용을 `.agents/skills/<name>/**` 경로로 재매핑해 기록한다.
  - 게시 16종(`moai-<command>`)은 이미 `.agents/skills/` 실파일이라 재매핑 대상 아님. 카탈로그 34개 이름과 게시 16개 이름의 교집합은 오늘 0개(research.md §4 실측) — 그래도 충돌 정책은 REQ-IH-007로 고정한다.

- **REQ-IH-007** (When a codex-only remap target path is occupied by a non-symlink entry): The deployer shall leave it untouched, skip the remap for that skill, and report a warning — `mirrorOneSkill`의 skip+report 선례(`internal/template/skill_mirror.go:221-229`)를 따른다.

- **REQ-IH-008** (While harness = codex): The deployed `AGENTS.md` shall explicitly disclose every capability in the claude 전용 기능 동결 목록 — 숨김 없음.
  - 현행 템플릿 AGENTS.md는 결속표 3행(질문 채널·Task 도구군·design-sync, `:21-25`), 스킬 로더 산문(`:27-33` — 주소·도달 축은 이미 공개), 훅 6/11(`:268-271`)을 갖는다. 보강 대상: Agent 소환·output style·슬래시 명령·Workflow 스크립트의 신규 공개 4건과, 스킬 로더의 미등가(deferred) 상태 명시 1건 — 후자는 표 행이 아니라 기존 산문 뒤 포인터 보강으로 충족한다(판정 근거: acceptance.md AC-IH-007). 바이트 상한 가드(`internal/config/token_budget_guard.go:94` `CodexContractByteCeiling` 24,576B; 현 17,059B)를 넘지 않는다.

- **REQ-IH-009** (Ubiquitous): The shipped documentation shall document the three harness selections and their deployment consequences — README init 절과 docs-site init 가이드에 codex 단독의 제약(claude 전용 기능 미제공)을 서술한다.

- **REQ-IH-010** (When `moai update` runs on a codex-only project): The update flow shall re-deploy through the codex-only surface set, reading `llm.harness` — `.claude/`를 부활시키지 않으며, `CleanMoaiManagedPaths`는 부재 경로를 조용히 건너뛴다.
  - update 재배포 경로는 `runCleanReinstall`(`internal/cli/update.go:340-393` 주변) + `internal/cli/update/deploy/deploy.go:51`. codex 배선 리프레시(REQ-CW-009, `internal/cli/update.go:494-502`)는 하니스와 무관하게 유지.

- **REQ-IH-011** (When `moai doctor` diagnoses a codex-only project): The doctor shall suppress or downgrade claude-surface findings (settings.json, hooks, statusline, `.claude/rules`, 미러→`.claude` 검사) while keeping the Codex checks (hooks 신뢰, config.toml, skills 등록) — `llm.harness`를 판정 입력으로 읽는다.

- **REQ-IH-012** (When init runs non-interactively with `--llm codex`): The deployment shall equal the interactive codex selection's file set — 플래그 경로와 위저드 경로의 결과 동일성.
  - 비대화형 MCP 비대칭(REQ-IQW-006)은 그대로다: codex는 어느 경로에서도 `.mcp.json`을 쓰지 않으므로 이 SPEC과 충돌하지 않는다.

- **REQ-IH-013** (Ubiquitous): The wizard's `agent_wiring` question shall carry exactly three options with deployment-consequence descriptions, synced across ko/ja/zh translations — 옵션 `Value`(claude/codex/both)는 동결하고 Label·Desc만 고친다. reconfigure 12문항은 불변(SPEC-INIT-QUIET-WIZARD-001 D1).

- **REQ-IH-014** (When the `--llm` value is outside the closed set): `validateInitFlags` shall reject it fail-loud — 현행 유지(`internal/cli/init_agent_flag_test.go:38-58`).

## §4 제약

- 위저드는 4문항 유지(질문 추가/삭제 금지 — t753 목표 형상·t583 D1 공동 구속).
- AC-CW-004(플래그 부재 ≡ claude, `.codex/` 배선 없음)와 기존 `init_agent_flag_test.go` 전체는 무변경 통과.
- 기존 사용자 프로젝트에 마이그레이션 절차를强요하지 않는다: `llm.harness` 키가 없는 기존 프로젝트는 claude로 해석되어 update 행동이 변하지 않는다.
- AGENTS.md 바이트 상한 24,576B 가드 유지.
- 템플릿 변경은 Template-First(`internal/template/templates/` 우저작 + `make build`). 16개 프로그래밍 언어 중립·4 로케일 번역 동기(ko/en/ja/zh — 위저드 번역은 ko/ja/zh 3로케일 + 영어 원본).

## §5 Out of Scope

### Out of Scope — t589 (구조 전환 · update --add-codex)

- `moai update --add-codex` 신설 동사와 CLAUDE.md→AGENTS.md 구조 전환, `--force` codex 추가 경로 철폐 (조사 §2차 결정 반영이 t589로 분할).
- LSEL curator의 AGENTS.md 마커 수확 경로 확장 — `## MOAI:LEARNED-WORKFLOW`는 CLAUDE.md 전용(`templates/CLAUDE.md:189`)으로 둔다. codex-only 사용자 프로젝트는 LSEL 루프의 Claude 런타임 자체가 없어 수확 대상이 아니며, 구조 전환 시의 판정은 t589 몫.

### Out of Scope — t753 (init 잔여 수리)

- F1 워크트리 "예" 답변 drop 수리(`internal/cli/init_workflow_flags.go:41-44` — 이 트리에서도 동일 앵커 확인).
- F4 비대화형 MCP 보장 호출 부재(`internal/cli/init.go:703-712` 대화형 전용 기본값 + `:980-987` — 조사의 `:542-567·911`은 t583 이전 행번호).
- 웹 콘솔의 `llm.harness` 표면 노출 여부 판정.

### Out of Scope — 기타

- codex 단독 → claude 역방향 전환 동사.
- 이미 초기화된 프로젝트의 하니스 전환(re-init 또는 t589 경로로만).
- claude 단독 배포에서 codex 표면(`.codex/agents`, `AGENTS.md`, `.agents/skills` 미러) 제거 — 조사 운영자 결정 "claude 단독=현행"에 따라 트리밍하지 않는다.
- 위저드 렌더링·그룹 라벨 재구성(t586), `.mcp.json`의 codex 포맷 변환(codex는 config.toml 사용 — 변환 없음).
- `moai update -c`(reconfigure) 문항 변경 — 12문항 불변.

## §6 교차 참조

- 선행·형제: SPEC-INIT-QUIET-WIZARD-001(위저드 4문항 프레임), SPEC-INIT-HARNESS-PROMPT-001(질문+플래그 해석), SPEC-CODEX-WIRING-001(codex 배선·AC-CW-*), SPEC-CODEX-COMMAND-SKILLS-001(게시 스킬·R-011).
- 후행: t589(update --add-codex·구조 전환), t753(F1·F4·web 표면), t586(렌더링).
- 조사: `.moai/reports/init-tui-audit-20260909.md` (읽기전용 근거 — primary 체크아웃, 커밋 여부와 무관하게 참조 전용).
