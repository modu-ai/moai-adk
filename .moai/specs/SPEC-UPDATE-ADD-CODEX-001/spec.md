---
id: SPEC-UPDATE-ADD-CODEX-001
title: "moai update --add-codex 신설 + CLAUDE.md→AGENTS.md 구조 전환 + init --force codex-add 경로 철폐"
version: "0.2.0"
status: in-progress
created: 2026-09-09
updated: 2026-09-09
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/codexwiring, internal/template/templates/AGENTS.md, internal/template/templates/CLAUDE.md"
lifecycle: spec-anchored
tags: "update, codex, wiring, agents-md, init-force, template-mirror, byte-ceiling"
tier: M
era: V3R6
related_specs: [SPEC-CODEX-WIRING-001, SPEC-UPDATE-MIRROR-HEAL-001, SPEC-UPDATE-VERSION-FLAG-001, SPEC-INIT-HARNESS-PROMPT-001]
---

# SPEC: moai update --add-codex 신설 + CLAUDE.md→AGENTS.md 구조 전환 + init --force codex-add 경로 철폐

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-09 | manager-spec | 최초 작성 — 카드 t589. 워크트리 실측 5종(바이트 상한 가드, codexwiring 계약, update 플래그 면, init --force 테스트 인벤터리, .agents/skills 존재)을 근거로 REQ 14건·AC 14건 확정. 카드의 스테일 인용 2건(validator.go:258-284, phase_test/validator_test)을 §1.3에 정정 기록 |
| 0.2.0 | 2026-09-09 | manager-spec | plan-audit iter1(PASS-with-fixes 0.86, `.moai/reports/t589/plan-audit.md`) 차단 2건 수리 — F1: REQ-UAC-005 의 기계 판정 확보(§D.4 #1 을 구체 테스트명으로 승격 + plan §D1 거부 전파 구속 추가; 형제 래퍼가 ErrValidationRefused 까지 삼키는 mutant 봉쇄). F2: 템플릿 CLAUDE.md `@AGENTS.md` import 유지 요구를 AC-UAC-011 에 추가(헤딩 카운트만으론 import 행 상실을 못 잡음). F5(권고): plan §A 에 t585 공유 표면 조율 문장 1줄 |

## 1. 문제 — 측정된 형태

기존 claude 프로젝트에 codex 하네스를 **덧붙이는**(additive) sanctioned 경로가 없다. 오늘의 유일한 경로는 `moai init --force --agent both` 다 — 즉 **재초기화를 통한 파괴적 우회**다. 2026-09-09 감사(`.moai/reports/init-tui-audit-20260909.md` §7)가 관측한 대로: init --force 는 `.moai/` 만 백업하고 사용자가 편집한 CLAUDE.md 는 경고 후 덮어쓴다.

**측정 트리**: `.claude/worktrees/t589` (branch `WT-add-codex-verb`) @ `5caddeb2d`. 이 SPEC 의 모든 수치는 이 트리의 실측이다. 아래 `bin/moai` 는 이 트리에서 `go build -o bin/moai ./cmd/moai` 로 빌드한 것이다.

```console
$ ./bin/moai update --add-codex ; echo exit=$?

   ERROR

  Unknown flag: --add-codex.

  Try --help for usage.
exit=1
```

### 1.1 실측 요약 (트리 `5caddeb2d`)

| # | 측정 대상 | 관측값 | 근거 |
|---|---|---|---|
| M1 | 템플릿 AGENTS.md 바이트 | **15,415 B** (`wc -c`) | 상한 24,576 B 대비 여유 **9,161 B** |
| M2 | 바이트 상한 가드 | `CodexContractByteCeiling = 24576` (`internal/config/token_budget_guard.go:95`). 발화 면은 **Go 테스트**다 — `TestCodexContractByteCeiling`(`internal/config/token_budget_guard_test.go:83`)가 위반 시 빌드를 실패시키며, 측정 대상은 **살아있는 루트 AGENTS.md + 템플릿 미러 둘 다**다(`contractDocuments`) | |
| M3 | `codexwiring.Wire` 계약 | `func Wire(projectRoot string, out, warn io.Writer) (Result, error)` (`internal/codexwiring/wire.go:43`). 쓰는 것: `.codex/hooks.json`(병합 렌더 + 화이트리스트 사전검증, 위반 시 무기록 거부), `.codex/config.toml`(`EnsureMCPTable`/`EnsureStatusLine` — create-if-absent **병합**, 사용자 내용 보존), `.moai/state/codex-wiring.json`(신뢰 사이드카, hooks.json 을 쓴 실행에서만). 불변 재생성은 무기록·무기록 재기록(멱등 내장). `RefreshWiring`(wire.go:57)이 존재 게이트 래퍼고 `wireProject` 본체에는 게이트가 없다 | |
| M4 | update 플래그 면 | `check, shell-env, config, force, yes, templates-only, binary, dry-run, no-hooks, restore, verbose, profile, version` — `--agent` 도 `--add-codex` 도 없다. **주의: update 에는 이미 다른 의미의 `--force` 가 있다**(버전 매치 skip 해제 등 — init 의 것과 무관). wiring 경로는 `update.go:507` `refreshCodexWiringBestEffort`(게이트 버전)가 template sync 직후에 이미 호출된다 | |
| M5 | init --force 테스트 인벤터리 | 카드가 지목한 `internal/cli/phase_test.go`·`internal/cli/validator_test.go`는 **이 트리에 없다**(§1.3). 실측된 고정 테스트: `init_test.go`(`--force` 플래그 목록 52행), `init_agent_flag_test.go`(both 시맨틱·`resolveAgentWiring`·`wireCodexUnlessClaude` 호출 검증 161행), `init_agent_wizard_test.go`, `init_agent_wizard_precedence_test.go`, `codex_auth_ladder_test.go`, `tokens_test.go` | |
| M6 | `.agents/skills` 존재 | **16개 `moai-*` 발행 세트가 이 트리에 존재한다**(`ls internal/template/templates/.agents/skills/`) — commandemit 이 착지돼 있으므로 `--add-codex` 가 오늘 배포 가능한 스킬 면이 완성돼 있다 | |

init 쪽 참고 실측: `--agent` 플래그는 `internal/cli/init.go:132`, 폐쇄 집합 검증은 `validateInitFlags`(init.go:335), `--force` 플래그는 init.go:83("Reinitialize an existing project (backs up current .moai/)"), already-initialized 가드는 init.go:843. init 의 codex 배선은 `wireCodexUnlessClaude`(init.go:191)가 **게이트 없는** `codexwiring.Wire`(init.go:195)를 이미 호출한다 — `--add-codex` 가 필요로 하는 호출은 이것과 동일하다.

### 1.2 설계 결정 기록 (재논의 금지 — plan.md §D 와 쌍을 이룬다)

| # | 결정 | 근거 |
|---|---|---|
| D1 | `--add-codex` 는 `codexwiring.Wire`(**게이트 없는 입구**)를 호출한다. 존재 게이트(`RefreshWiring`) 우회는 이 동사의 경로에 한정된다 | 동사의 존재 이유가 "없는 것을 만드는 것"이므로 게이트가 목적을 무효로 한다 |
| D2 | `--add-codex` 는 `.mcp.json` 을 건드리지 않는다 | codex MCP 는 `.codex/config.toml` 로 간다. claude 쪽 `.mcp.json` 엔트리는 init 이 이미 프로비저닝했다(init.go:236 부근, "Provisioned the moai MCP server entry"). 카드의 "기존 .mcp.json 을 망가뜨리지 않는다" 요구는 **비간섭으로 충족**된다 |
| D3 | `## MOAI:LEARNED-WORKFLOW` 제목을 AGENTS.md 에 **미러링하지 않는다** | Tier-4 curator 쓰기 표면은 CLAUDE.md 앵커(`internal/harness/curator` `TierSurfaceMap`)이고 병합 보호 목록도 CLAUDE.md 앵커다(`internal/merge/strategies.go:484-485`). 무음 미러는 두 계약 모두를 깬다. 운영자 HARD-2 의 "mirroring 도 omission 도 아닌 명시" 충족: 미러 금지를 REQ-UAC-012 로 명세하고, codex 세션이 학습 다이제스트를 못 보는 격차를 §5 에 기록하며, curator 이중 쓰기 확장은 후속 SPEC 후보로 뺀다. 추가 근거: AGENTS.md 에는 24,576 B 상한이 있는데 미러는 그 예비(9,161 B)를 소모한다 |
| D4 | `init --force` 는 **축소**(안내 + help 정비)하며 완전 제거하지 않는다 | --force 에는 재초기화라는 정당한 역할이 남는다( help 텍스트가 이미 그렇게 쓰여 있다 ). 제거는 기존 사용자를 깨는 breaking change라 별도 운영자 결정 사안으로 기록한다(§6). 철폐의 대상은 "codex-add 경로"로서의 --force 다 |
| D5 | no pure wrapper — 템플릿 CLAUDE.md 의 18개 `## N.` 제목(§0–§17)을 모두 유지한다 | 운영자 HARD-1. universal 절(§6 품질 게이트 포인터 등)의 **본문**만 AGENTS.md 로 옮기고 제목은 남긴다 |
| D6 | 5개 codex 강화 섹션의 제목을 plan.md §D6 에서 미리 고정한다 | AC-UAC-009 의 grep 앵커를 SPEC 단계에서 확정해, "무엇을 쓸지"가 run-phase 재량이 되지 않게 한다 |

### 1.3 카드 인용 정정 — 스테일 좌표 2건

카드 원문의 좌표 2건이 이 트리에서 해결되지 않는다. run-phase 의 재측정 항목으로 넘긴다(plan.md §C).

1. **`validator.go:258-284`(.moai/ 백업, os.Rename)** — `internal/cli/validator.go` 는 존재하지 않는다. `rg -n "os.Rename" internal/cli/` 재탐색에서 `internal/cli/memory.go`·`internal/cli/update/deploy/deploy.go`·`update_destructive_registry.go` 가 후보로 관측됐으나, 측정 명령의 `-r` 플래그 오용(치환 모드)으로 출력이 오염돼 정밀 좌표가 아니다. run-phase 에서 클린 재측정이 필요하다. M3 의 구현이 백업 메커니즘을 바꾸지 않으므로(안내 추가 + 테스트 갱신이 전부) 이 좌표는 진행을 막지 않는다.
2. **`phase_test` / `validator_test`(internal/cli)** — 부재하다. 실제 고정 테스트는 §1.1 M5 의 6개 파일이다. REQ-UAC-014 는 이 실측 목록을 따르고, run-phase 진입 시 plan.md §C 로 재인벤토리한다.

## 2. 원인 — codex-add 가 --force 로 좁혀진 구조

SPEC-CODEX-WIRING-001 은 배선 생성을 init(--agent codex|both)에 두고, update 에는 **보존적** 경로(REQ-CW-009 존재 게이트)만 두었다. 이 설계 아래에서 wiring 파일이 없는 기존 claude 프로젝트는 update 로는 codex 를 영영 받을 수 없고, init 쪽은 "이미 초기화된 프로젝트" 가드(init.go:843) 뒤에 --force 가 있어야 재진입이 가능하다. 두 문이 서로 맞물리며 유일한 통로가 `init --force --agent both` 가 된 것이다 — 그리고 그 통로는 CLAUDE.md 를 경고 후 덮어쓴다(§1 감사 인용).

즉 결함은 배선 코드의 부재가 아니라 **덧붙임(additive) 동사의 부재**다. 부품(`codexwiring.Wire`)은 존재하고 init 이 이미 게이트 없이 호출하고 있다(§1.1). 빠진 것은 그것을 기존 프로젝트에서 부르는 update 쪽 문이다.

## 3. 이 카드의 대표 mutant

아래 mutant 들이 이 SPEC 의 AC 를 통과하려면 AC 가 너무 얕은 것이다 (verification-completeness §2 mutant probe).

- **M-1 "no-op 플래그"**: `--add-codex` 를 파싱만 하고 아무것도 호출하지 않는 mutant. AC-UAC-001(플래그 존재)은 통과하지만 AC-UAC-002(산출물 생성)에서 잡혀야 한다.
- **M-2 "게이트 우회 실패"**: `Wire` 대신 `RefreshWiring` 을 호출하는 mutant. 이미 wiring 이 있는 프로젝트(AC-UAC-005/006)에서는 통과하지만 claude-only 프로젝트(AC-UAC-002)에서 잡혀야 한다 — 존재 게이트 우회가 이 SPEC 의 핵심 행위다.
- **M-3 ".mcp.json 재프로비저닝"**: claude 쪽 `.mcp.json` 을 재기록하는 mutant. 아티팩트 검사(AC-UAC-002)는 통과하지만 AC-UAC-004(바이트 동일)에서 잡혀야 한다.
- **M-4 "마커 미러"**: LEARNED-WORKFLOW 제목을 AGENTS.md 에 복제하는 mutant. 콘텐츠가 늘었다는 이유로 AC-UAC-009 를 통과하지만 AC-UAC-012 에서 잡혀야 한다.
- **M-5 "pure wrapper"**: CLAUDE.md 를 `@AGENTS.md` 한 줄로 바꾸는 mutant. AC-UAC-011(18 제목 불변)에서 잡혀야 한다 — 운영자 HARD-1 의 기계적 형태다.

## 4. 요구사항 (GEARS)

### 신설 동사 — update --add-codex (M1)

- **REQ-UAC-001** (Event-driven) — **When** the operator runs `moai update --add-codex` in an existing project, the update command shall invoke the ungated `codexwiring.Wire` so that `.codex/hooks.json`, `.codex/config.toml` (`[mcp_servers.moai]` + `[tui].status_line`), and `.moai/state/codex-wiring.json` are created. **Where** the project already carries wiring, the same invocation shall behave as today's regeneration pass (no duplicate writes).
- **REQ-UAC-002** (Unwanted) — **When** `--add-codex` is absent, the update command shall keep the existence-gated path (`codexwiring.RefreshWiring`) unchanged — a claude-only project gets nothing created (SPEC-CODEX-WIRING-001 REQ-CW-009 계승, 회귀 금지).
- **REQ-UAC-003** (Unwanted) — The `--add-codex` path shall not read, write, or reposition `.mcp.json` (결정 D2).
- **REQ-UAC-004** (State-driven) — **While** the project's wiring is already current (hooks.json unchanged, config.toml already merged), a re-run of `moai update --add-codex` shall write nothing new and print no re-trust guidance (the unchanged-regeneration silence `wireProject` already implements).
- **REQ-UAC-005** (Event-driven) — **When** the rendered hooks.json fails the whitelist gate (`codexadapter.ValidateConfig` → `ErrValidationRefused`), the `--add-codex` invocation shall fail loud naming the violating keys, with no wiring bytes written.
- **REQ-UAC-006** (Event-driven) — **When** `--add-codex` is combined with `--dry-run`, the update command shall preview the wiring actions and write nothing.
- **REQ-UAC-007** (Event-driven) — **When** `--add-codex` is combined with `--check`, the update command shall reject the combination fail-loud by extending the existing `validateUpdateVersionConflicts` mutual-exclusion validator (informational flag × mutation flag).
- **REQ-UAC-008** (Ubiquitous) — The `moai update` help text shall describe `--add-codex` as the sanctioned additive codex path for an existing project.

### AGENTS.md 계약 강화 + universal 이동 (M2)

- **REQ-UAC-009** (Ubiquitous) — `internal/template/templates/AGENTS.md` shall carry the five codex reinforcements under the five section titles pinned in plan.md §D6: (a) Codex web console guidance, (b) hook-event coverage delta (codex fires fewer than claude's 11 hook events — the delta is explained, not just counted), (c) a `.moai/config/sections/` pointer, (d) a `moai` CLI verb table, (e) statusline tokens.
- **REQ-UAC-010** (Ubiquitous) — `internal/template/templates/AGENTS.md` shall stay ≤ 24,576 bytes (`CodexContractByteCeiling`), and `TestCodexContractByteCeiling` shall stay green (fail-closed 가드 — 잘림은 무신호이므로 이 가드가 유일한 신호다).
- **REQ-UAC-011** (Event-driven) — **When** a universal clause of the template CLAUDE.md (census: §6 quality-gate pointer; part of §7) is relocated into AGENTS.md, the template CLAUDE.md shall keep all 18 `## N.` section headings (§0–§17) — the claude-only orchestration sections stay verbatim, and a pure `@AGENTS.md` wrapper is prohibited (결정 D5, 운영자 HARD-1).
- **REQ-UAC-012** (Unwanted) — `internal/template/templates/AGENTS.md` shall not carry a `## MOAI:LEARNED-WORKFLOW` heading; the Tier-4 curator write surface (`internal/harness/curator` `TierSurfaceMap`) and the merge-protection list (`internal/merge/strategies.go:484`) shall stay CLAUDE.md-anchored (결정 D3, 운영자 HARD-2 의 명시적 결론).

### init --force codex-add 철폐 (M3)

- **REQ-UAC-013** (Event-driven) — **When** init runs with `--agent codex|both` against an already-initialized project (the `--force` reinit path), init shall print guidance naming `moai update --add-codex` as the sanctioned additive path, before proceeding with the reinit the operator explicitly asked for. The reinit is not blocked — the codex-add *purpose* is redirected, the reinit *capability* remains (결정 D4).
- **REQ-UAC-014** (Ubiquitous) — The pinned contract tests covering `--force` / `--agent both` semantics — the run-phase re-inventory of the six files measured in §1.1 M5 (`init_test.go`, `init_agent_flag_test.go`, `init_agent_wizard_test.go`, `init_agent_wizard_precedence_test.go`, `codex_auth_ladder_test.go`, `tokens_test.go`) — shall be updated in the same change that narrows the semantics.

## 5. 알려진 구속 조건

1. **AGENTS.md 바이트 예비는 유한하다.** 이 트리 기준 여유 9,161 B. plan.md §D6 의 5 섹션 + universal 이동분의 예산을 §D6 가 함께 고정한다(목표 총 증분 ≤ 6,500 B, M2 후 목표 ≤ 22,000 B). 초과 시 본문을 늘리지 말고 포인터화한다 — AGENTS.md 는 codex 가 상한 아래에서 읽고 초과분을 **조용히** 자르는 문서다(token_budget_guard.go 주석 인용).
2. **codex 세션은 학습 다이제스트를 못 본다 (기록된 격차).** 결정 D3 에 따라 LEARNED-WORKFLOW 블록은 CLAUDE.md 에만 산다. codex 세션이 이 항상-로드 다이제스트를 보려면 curator Tier-4 표면의 AGENTS.md 확장이 필요하고, 그것은 병합 보호 목록·바이트 상한 예산까지 함께 다루는 별도 SPEC 이다. 이 카드는 그 격차를 만들지도 메우지도 않는다 — 기록만 한다.
3. **update 의 `--force` 와 init 의 `--force` 는 다른 플래그다.** 이 카드는 update 쪽 --force 를 건드리지 않는다(REQ-UAC-002 의 보존 축과 무관).
4. **템플릿 내용 언어는 영어다.** 5 섹션 본문은 배포 템플릿 관례(16 프로그래밍 언어 중립 + 영어 본문)를 따른다. 템플릿 중립성 CI 가드(`template-neutrality-check.yaml`)가 같은 변경을 검사한다 — SPEC ID·내부 날짜·커밋 SHA 는 템플릿에 넣지 않는다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — CLAUDE.md 순수 래퍼 전환

- 템플릿 CLAUDE.md 를 `@AGENTS.md` 한 줄로 바꾸는 전환(§1-§17 삭제). 운영자 HARD-1 로 금지됐고, operator 가 나중에 원하면 별도 결정이다(REQ-UAC-011 은 이 전환을 봉쇄한다).

### Out of Scope — MOAI:LEARNED-WORKFLOW curator 이중 쓰기

- Tier-4 curator 쓰기 표면을 AGENTS.md 로 확장하거나 마커를 미러링하는 작업. 결정 D3 — 병합 보호 목록(`internal/merge/strategies.go`)과 바이트 상한 예산을 함께 다루는 별도 SPEC 의 소관이다.

### Out of Scope — init --force 플래그 완전 제거

- `--force` 플래그 자체의 삭제. 결정 D4 — 재초기화 역할이 남아 있고 제거는 breaking change다. codex-add 목적의 사용은 REQ-UAC-013 의 안내로 축소된다.

### Out of Scope — init --force 재초기화 시 CLAUDE.md 백업 확대

- 재초기화 경로가 CLAUDE.md 도 백업하도록 바꾸는 작업. 감사가 지적한 파괴성 자체의 수리는 이 카드의 해법인 `--add-codex`(덧붙임 경로)가 우회하며, 백업 정책 수리는 별도 소관이다.

### Out of Scope — init 사이드 구조 전환

- init 시점의 harness 질문·3-way 배포 확장은 형제 카드 t585 의 소관이다. 이 카드는 update 사이드만 소유한다.

### Out of Scope — .agents/skills 신규 발행

- 16개 발행 세트는 이 트리에 이미 존재한다(§1.1 M6). 이 카드는 새 스킬을 발행하거나 commandemit 을 바꾸지 않는다.

## 7. 미검증 항목 (Gaps)

- `init --force` 재초기화 시 `.moai/` 백업의 정확한 호출부 좌표 — 이 트리에서 미확정(§1.3 #1). M3 구현은 백업 메커니즘을 변경하지 않으므로 블로커가 아니지만, run-phase 가 클린 재측정한다(plan.md §C).
- 5 섹션의 실제 바이트 소모 — §D6 예산은 목표치이며, M2 종료 시 `wc -c` 실측으로 확정된다(AC-UAC-010 이 상한의 판정자다).
- 루트 계약 쌍(`AGENTS.md`/`CLAUDE.md`)과 템플릿 쌍의 바이트 동일성 강제 가드 존재 여부 — 미확인. 없다면 M2 후 루트 쌍은 다음 `moai update` 까지 구판으로 남는다(잔여 위험, plan.md §H).
