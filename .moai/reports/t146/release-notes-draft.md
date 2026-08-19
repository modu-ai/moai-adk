# t146 — v3.1.1 2차 배치 릴리즈 노트 초안

**상태: 초안 (DRAFT).** 배치 PR이 머지돼야 릴리즈 범위가 확정되므로 이 문서는 확정본이 아니다.
버전 스탬프 실제 적용과 태깅은 `/harness:release` 소관 — 이 카드는 CHANGELOG.md를 편집하지 않았다.

- 범위: `4100d8767..5798bdc2e` (PR #1582 머지 이후 `release/v3.1.1`에 쌓인 분량)
- 측정 시각 기준 `origin/release/v3.1.1` = `5798bdc2ec630f83704551e4e99c4291ec723026`
- 작업 트리: `WT-t146` (`.claude/worktrees/t146`, base `origin/main`)

---

## 0. 카드 전제 정정 — "23커밋"은 실측과 다름

카드 본문은 이 배치를 23커밋이라 적었으나, 실측값은 그보다 크다.

| 세는 방식 | 명령 | 값 |
|---|---|---|
| 전체 | `git rev-list --count 4100d8767..5798bdc2e` | **57** |
| first-parent (릴리즈 브랜치 위의 통합 커밋 열) | `git rev-list --count --first-parent …` | **34** |
| 머지 제외 (실제 내용 커밋) | `git rev-list --count --no-merges …` | **42** |

23이라는 숫자는 셋 중 어느 것과도 맞지 않는다. 아래 분류는 **머지 제외 42커밋**을 기준으로 했다
(머지 커밋 15개는 워크트리 통합 기록이라 릴리즈 노트에 담을 내용이 없다).

추가 실측 2건:

- **CHANGELOG.md는 이 배치에서 한 줄도 바뀌지 않았다.** `git diff --stat 4100d8767 5798bdc2e -- CHANGELOG.md` → 출력 없음.
  현재 `[Unreleased]`에 들어 있는 3개 SPEC 항목(MOAI-HOME-PATHS / MOAI-CLEAN-HOME / ALWAYS-LOADED-DIET)은
  **1차 배치(PR #1582)의 산물**이며 v3.1.1의 일부다. 아래 초안은 그 3개를 유지한 채 2차 배치분을 덧붙이는 형태다.
- **README 배지 4종과 `docs-site/hugo.toml`은 이미 v3.1.1로 올라가 있다** (`024536e38`).
  `pkg/version/version.go`만 `v3.1.0`에 머물러 있다.

---

## 1. `[Unreleased]` 초안 (CHANGELOG.md 삽입용, 영어)

> 삽입 위치: 기존 `## [Unreleased]` 블록. 기존 3개 SPEC 항목은 그대로 두고, 각 카테고리에 아래 항목을 **덧붙인다**.
> Keep a Changelog는 Added / Changed / Deprecated / Removed / Fixed / Security만 인정하므로,
> 룰·독트린 교정과 문서 스윕은 별도 카테고리를 만들지 않고 **Changed** 아래에 묶었다.

```markdown
## [Unreleased]

### Added

- **`moai-domain-design-dna` — a reference-design deconstruction skill.** Distilled from the MIT-licensed
  `zanwei/design-dna` methodology: a three-dimension Design DNA profile (`design_system` / `design_style` /
  `visual_effects`) is extracted from a screenshot, an image set, or a URL, and a self-contained artifact is
  generated from that profile. `SKILL.md` (L2) carries the extraction rules, the generation priority order, and
  the delivery gate; the field schema and the effect-implementation patterns are L3 references loaded on demand.
  Overlapping material is cross-referenced rather than restated — `moai-ref-ui-polish` owns component finish,
  `dataviz` owns chart palettes, `moai-domain-html-report` owns the report render surface. Registered in
  `skill-routing.md` §1.1 (orchestrator-direct) and in `catalog.yaml`; MIT attribution recorded in
  `.claude/rules/moai/NOTICE.md`. (`6a64a0bad`)

- **`moai-ref-ui-polish` gains a motion-reasoning layer and an executable design audit.** Two new L3 references,
  both loaded on demand: `references/motion-principles.md` (the three decision passes, three motion layers, the two
  1/3 rules, stagger budgets, four personality archetypes, Disney's 12 principles with UI numeric ranges, and the
  emotion-to-motion map — distilled from LottieFiles' MIT `motion-design` skill), and
  `references/design-audit.md` (motion-gap patterns, a three-stack `prefers-reduced-motion` probe, accessibility
  and layout-property checks, and duration/easing inventories — distilled from the MIT `genjutsu` plugin). The L2
  body gains a required `prefers-reduced-motion` branch — the skill previously had no accessibility escape hatch —
  the compositor-vs-layout framing for why `transform`/`opacity` outperform layout properties, and the mobile
  no-hover doctrine with its desktop inverse. Detection patterns carry a signal rating, so a match is a candidate
  rather than a defect. MIT attributions recorded in `.claude/rules/moai/NOTICE.md`. (`f34462ef6`, `16899d982`)

- **[SPEC-GOAL-STOPFAILURE-CLEAR-001](.moai/specs/SPEC-GOAL-STOPFAILURE-CLEAR-001/spec.md) — an armed
  `/moai goal` now disarms itself when the turn dies on an unrecoverable API error.** A revoked credential, an
  org-refused OAuth account, or an exhausted balance ends the turn that was doing the work; the goal previously
  survived it, still armed, and every subsequent turn-end found the condition unmet with nothing advancing it —
  the idle spin to the ceiling that `goal-directive.md` already names as the cost of arming a goal with nothing
  running. The disarm rides `StopFailure` rather than the goal evaluator, because a turn that dies on an API error
  does not end through `Stop` at all. Only `authentication_failed`, `oauth_org_not_allowed`, and `billing_error`
  disarm: `rate_limit`, `overloaded`, and `server_error` resolve on retry and the goal is exactly the state that
  must survive to see it, `max_output_tokens` is classified withheld-recoverable by `runtime-recovery-doctrine.md`
  §1, and an unknown or empty type is not evidence of unrecoverability. Disarming on any of those would destroy
  live state on a self-fixing condition — the opposite defect, and a quieter one. This is the loop's fourth exit
  and the only one that does not need turns to keep completing; the turn ceiling, the runtime block cap, and the
  stagnation guard keep their semantics unchanged. Context overflow, the third case upstream names, has no value
  in the documented `error_type` enum and is deliberately not covered. Written reproduction-first.
  (`eb597b319` plan, `34a94cc80` run)

- **The Kanban bootstrap notice now states the recommended backend mix and the per-lane agent cap.** The
  SessionStart notice (all four locales) recommends lead on GLM, plan on Claude, run on GLM, sync on Claude — the
  judgment and review lanes ride Claude while the implementation lane and the always-on lead ride GLM — with an
  explicit note that any other mix, one backend everywhere included, works just as well. A second line states that
  each companion lane can run up to 10 agents concurrently. (`00af8d334`, `3a3e711ca`)

### Changed

- **`manager-kanban` is renamed `manager-lead`, and the role widens from kanban-only to kanban + factory lead
  coordination.** Role A (Tier L in-session hierarchical fan-out — sole `Agent`-carrier, depth-2 seal) is
  semantically unchanged. Role B now covers both cross-session modes: Kanban Mode (`-k`, the
  lead > plan > run > sync chain, with the plan lane fanning out per-card SPEC authoring to parallel workers) and
  Factory Mode (`-f`, lanes `lane-1..N` each carrying a card through plan → run → sync in-session), lanes running
  up to 10 concurrent agents. The rename is mechanical and complete across live wiring: the agent file and its
  template mirror, `catalog.yaml`, the name-keyed Go maps (agentlint allowlist, delegation map, v4 manifest agent
  tiers), the depth-seal CI guard (`manager_kanban_depth_test.go` → `manager_lead_depth_test.go`), `CLAUDE.md` §4
  and §15, `moai-mcp-tools.md`, statusline fixtures, `workflow.yaml`, and `ci.yml`. Companion sessions are renamed
  to the bare role names `plan` / `run` / `sync`, and factory workers are labelled `lane-<n>`. The docs surface
  follows across all four locales: the `advanced/manager-kanban` page is retired in favour of `advanced/manager-lead`
  (a verified strict superset — all 11 old H2 sections carried over plus one added), slugs migrate with permanent
  `vercel.json` redirects, and the README agent catalog is corrected in all four files. Historical records
  (prior CHANGELOG entries, closed SPEC bodies, the "renamed from" sentences on the new pages) deliberately keep
  the old name. (`310d75dd2`, `d73e9a669`, `59a2ddfe3`, `f35f74a45`, `5a0660ba2`, `c365350ad`, `948a565e8`,
  `c25a418ab`, `c76566817`, `0daaada1c`, `0e5239066`, `05378d75a`)

- **The Kanban lead session is named by its bare role, `lead`, rather than `lead-<run-id>`.** The lead was the
  last session carrying a run id in its name; companions stopped carrying one under the one-machine-one-run policy.
  Nothing functional was downstream of the round-trip — companion names carry no run id, the board and the
  per-session records key on the Claude session id, and `MOAI_KANBAN_LEAD_ADDR` is a conventional display path —
  so `leadRunID` now adopts a still-set `MOAI_KANBAN_ID` and mints one when there is none, with no new state file
  introduced to carry a display-only value. A legacy `lead-<run-id>` name is still adopted, so an operator pasting
  an old launch line lands on the run they meant; a bump number is explicitly not adopted (`lead-2` is the second
  live lead on this machine, not run 2). Collision handling mirrors the companion side: a second live lead takes
  `lead-1`, keeping every session addressable by name alone. (`c326eb4e0`)

- **The GLM surface is retargeted at GLM-5.3, and `glm-5.2` is withdrawn from the tier-slot closed set.**
  z.ai's GLM-5.3 documentation settles the question the previous overlay was written around: reasoning is always
  on, disabling it is no longer supported, and `reasoning_effort` takes `{low, high, max}` with `max` as the
  default. The pre-5.3 thinking-off state is therefore unreachable, not merely weakest, so the collapse floor moves
  to a wire-real level: `GLMState{Low,High,Max}` replace `GLMState{ThinkingOff,ReasoningHigh,ReasoningMax}`,
  `GLMReasoningEffortLow` is added, and `GLMThinkingDisabledVal` is removed so no caller can reach the retired
  state. `CollapseClaudeEffortToGLM` keeps its 5-to-3 shape and totality clause — only the floor's identity
  changes. Tier effort defaults become `high=high`, `medium=high`, `low=low`, `fable=max`; the template `llm.yaml`
  mirrors all four tiers onto `glm-5.3`; i18n option keys and labels follow in all four locales. **One behavior
  change rides along**: `glmReasoningEnvVarsForEffort` now emits `ANTHROPIC_REASONING_EFFORT=low` for Claude effort
  `low`, where it previously emitted no key at all — under GLM-5.3 the absent key is the failing case.
  `DefaultGLM52` stays declared, so an existing `llm.yaml` naming `glm-5.2` still loads and still resolves a
  context window. (`d7a629869`)

- **Statusline layout: GitHub counts fold into the repo segment, and the `[WT]` marker moves behind the branch
  glyph.** The repo line reads `🔀 <issues> / <prs> → 📡 owner/name, <ahead>/<behind> | [WT] 🅱️ branch +N`:
  GitHub activity points at the repository it belongs to, ahead/behind demote from branch arrows (`↑N ↓N`, zeros
  omitted) to an always-on bare pair on the repo, and the session line drops its own `🔀` segment. The `SegmentGitHub`
  gate governs the prefix; repo and branch survive it. Per a follow-up operator instruction the counts prefix is
  then removed from the repo segment, and the TODO counter tightens from `6 / 12` to `6/12`.
  (`81dc1e044`, `d13bd8a8f`, `44447c1b8`)

- **The Factory Mode session-start notice drops the no-model-override paragraph.** The rule itself is unchanged
  and stays documented in the factory skill; only its per-session-start repetition is retired. (`d41f3c0fc`)

- **v3.1.1 documentation bundle — 13 pages × 4 locales, with a version sync and a mermaid audit.** Korean drafts
  landed (README.ko plus 9 replaced pages and 3 new `cli-reference` pages: graph, tokens, memory) and en/ja/zh were
  fully re-derived; `_meta.yaml` ×4 and the menu map gained the 3 new entries. Four addenda ride along:
  terminology (the "factory mode" prose of the time rewritten across affected pages), a mermaid audit (22 fix files
  applied; residual slash-leading labels 0), a version sync (`hugo.toml` v3.1-rc.2 → v3.1.1 with releaseDate
  2026-08-18, README badges ×4, and the statusline/update/FAQ/feedback example displays), and the backend-mix
  recommendation documented in the README kanban section and the kanban-mode page in all four locales. The launcher
  pages are corrected in the same line of work: the flag table had claimed `-f` / `--factory` was retired and now
  errors, which is false as of v3.1.1 — `-f` is the dedicated factory entry, with a `-k <N>` compatibility row.
  Gates measured: heading parity 12 pages × 4 locales, warning-free `hugo` build with sitemap, URL blacklist 0,
  mermaid TD-only, body emoji limited to the UI-glyph documentation class. (`024536e38`, `c25a418ab`)

- **Always-loaded rule surface put back under budget via stub + companion splits.** The budget guard measured
  75,992 / 76,000 tokens (8 headroom) on `release/v3.1.1`. Restructured without raising the constant:
  `agent-common-protocol.md` 38,136 → 27,102 B, `askuser-protocol.md` 30,154 → 23,539 B,
  `kanban-dispatch.md` 23,728 → 23,421 B, and `output-styles/moai/moai.md` 65,238 → 63,406 B, with the relocated
  material moving to `paths:`-scoped companions. Every `[ZONE]` / `[HARD]` clause, discriminator table, and
  cross-referenced section heading stays inline; only rationale and post-emission explanatory prose relocates.
  Measured 70,538 tokens after, headroom 5,462. A second pass trims `session-handoff.md` 23,465 → 21,197 B under
  the same render-surface ruling, leaving the 6-block skeleton byte-identical. Template mirrors byte-identical
  throughout. (`3f31135c3`, `b796fddf3`)

- **Doctrine corrections against measured upstream and in-tree behavior.** Six rule edits, each pinned to an
  observation rather than to a reading: `ListAgents` no longer claims to report the live set, so the kanban
  dispatch cycle stops reading an incomplete session listing as an empty role (the `[HARD]` clause keeps its force —
  only what counts as evidence of emptiness narrows); the in-process mailbox collision that silently eats a
  bare-name dispatch is named, with the `routing` object identified as the only discriminator and a `[HARD]` rule
  to read the send's result; `model-policy.md` stops naming `glm-5.2` as the GLM default; the two Claude Code
  2.1.234 native `/goal` clearing behaviors are recorded as native-only, which `/moai goal` does not have;
  `archived-agent-rejection.md` §C states the `general-purpose` availability precondition its migration table
  depends on; and the worktree rule names the `worktree.baseRef` trap that makes `"head"` the wrong fix — no
  setting changed and no `baseRef` shipped in the template, since shipping `"head"` as a default would change every
  distributed user's worktree base to satisfy one repository's release-lane procedure.
  (`afff28d6a`, `87679d694`, `ad854ac6c`, `0b4f7d652`, `9a9778b30`, `2ceeb7b36`)

### Fixed

- **SPEC-HARNESS-GATE-TEST-001 — `moai gate` no longer hangs on
  a watch-mode Node test script.** The Node toolchain's test step was hardcoded to `npm test --`, which never exits
  for packages whose `test` script is a bare `vitest` or carries a `--watch` / `--watchAll` token; the step
  therefore always died at `TestTimeout` with the suite at 0%, and consumer projects bypassed the gate via
  `SKIP_MOAI_PRECOMMIT` for exactly this reason across three consecutive cards. `resolveNodeTestStep` now rewrites
  the step immediately before execution, reading `package.json` scripts from the resolved project dir, per a
  three-tier resolution: `scripts.test:run` present → `npm run test:run`; a watch-prone test script → the runner's
  non-watch flag appended (`vitest --run`, `jest --ci`); anything else → `npm test` unchanged. Every `package.json`
  parse failure falls back to the third tier, and non-Node toolchains are untouched. A follow-up correction drops
  the appended flags from tier (i): on a turbo monorepo root, `npm run test:run -- --passWithNoTests` expands to
  `turbo run test:run --passWithNoTests`, which turbo rejects at the argument parser before any test runs;
  `--passWithNoTests` stays on tiers (ii) and (iii), where the runner is known to accept it.
  (`f1ebd634a`, `692d44586`)

- **A raw `*exec.ExitError` at the CLI exit-code seam no longer silences a real failure.** `*exec.ExitError`
  happens to satisfy the `ExitCoder` interface (it has an `ExitCode() int` method), so a `%w`-wrapped subprocess
  failure reaching `cmd/moai` made `main` adopt the subprocess's raw exit code and fang's error handler suppress
  the error box — a silent-failure pair observed twice (in `moai worktree done`, and in the `moai hook
  worktree-create` fix below), both previously patched at the call site only. Fixed at the seam:
  `internal/cli.ResolveExitCode` refuses a raw `*exec.ExitError` chain-wide, then matches intentional `ExitCoder`
  carriers; `main.go` and `fang.go` resolve through it. The producer-side half is a new
  `internal/execerr.StatusDetail` that describes a subprocess failure (exit status plus captured stderr) without
  chaining the raw type, adopted at 24 wrap sites across cli / spec / verify / github / worktree / hook / core.
  Guards are `errors.As`-chain-based rather than message-based, and both failed against the pre-fix seam.
  (`8bcfd506d`)

- **`moai hook worktree-create` no longer fails with zero diagnostics on a non-repo cwd.** Measured before: a
  `WorktreeCreate` whose cwd sits outside any repository exited 128 with zero bytes on both stdout and stderr —
  and since this hook gates every isolated agent spawn, the user saw a dead spawn with no cause at all. The message
  was being built correctly and then discarded by the `*exec.ExitError` seam described above. The stderr text now
  stays printable and the exit status travels as data: measured after, rc 1 and 272 bytes on stderr naming both the
  directory and git's "not a git repository" text. The regression test asserts both properties, so a future rewrap
  fails loudly rather than going quiet. (`207d2e993`)

- **`moai-ref-ui-polish` local copy resynced with its template mirror.** The exit-animation easing row had
  diverged: the local copy still carried the pre-audit `ease-out` wording while the shipped template carried the
  corrected `ease-in`. Resolved template-to-local, since the template is the source of truth per the Template-First
  cycle and is the side that received the intentional correction — without which the subsequent motion-knowledge
  additions would have been silently reverted by the next `moai update`. (`99469f36e`)
```

---

## 2. 버전 스탬프 위치 (적용은 `/harness:release` 소관)

`[HARD]` 이 카드는 아래 어느 것도 적용하지 않았다. 위치와 현재 값만 표시한다.

| # | 파일 | 현재 값 (배치 tip `5798bdc2e` 기준) | 스탬프 시 필요한 값 |
|---|---|---|---|
| 1 | `CHANGELOG.md` | `## [Unreleased]` | `## [3.1.1] - <태그 날짜>` |
| 2 | `pkg/version/version.go:8` | `Version = "v3.1.0"` | `v3.1.1` — **유일하게 미갱신** |
| 3 | `docs-site/hugo.toml:55-56` | `version = "v3.1.1"` / `releaseDate = "2026-08-18"` | 이미 완료. 태그 날짜와 다르면 `releaseDate` true-up |
| 4 | `README.md` / `.ko` / `.ja` / `.zh` 배지 | `badge/Release-v3.1.1` ×4 | 이미 완료 (4/4 실측) |
| 5 | git 태그 + GoReleaser | — | `scripts/release.sh` 경유 |

보조 사항:

- `Makefile:6`의 `VERSION`은 `git describe --tags --abbrev=0` 유래라 태그를 달면 자동 추종한다. 손댈 필요 없음.
- CHANGELOG 하단에 비교 링크 정의(`[3.1.1]: …`)가 존재하지 않는다 — 이 리포는 그 관례를 쓰지 않으므로 추가하지 말 것.
- `[Unreleased]` 헤딩을 버전 헤딩으로 바꿀 때, 1차 배치 3개 SPEC 항목도 **함께** v3.1.1로 들어간다.
  1차/2차를 나눠 두 버전으로 쪼개려면 그건 별도 결정이다.
- 3.1.0 항목은 `### Summary` 문단을 앞에 두는 형식을 쓴다. 3.1.1도 같은 형식을 원하면 §3의 요약을 옮겨 쓰면 된다.

---

## 3. GitHub 릴리즈 노트용 요약 (이중언어)

### 한국어

**MoAI-ADK v3.1.1**

이번 릴리즈의 축은 셋이다.

**리드 세션이 이름을 갖는다.** `manager-kanban`이 `manager-lead`로 바뀌면서 역할이 칸반 전용에서
칸반 + 팩토리 리드 조율로 넓어졌다. 동반 세션은 `plan` / `run` / `sync`라는 맨 이름을 쓰고, 팩토리 워커는
`lane-<n>`으로 붙는다. 리드 세션 자체도 `lead-<run-id>`가 아니라 그냥 `lead`다 — 이름에 run id를 실어 나를
이유가 실제로는 없었다. 한 머신에 리드가 둘이면 두 번째가 `lead-1`을 집으므로, 모든 세션은 여전히 이름만으로
주소가 된다. 예전 `lead-<run-id>` 형태도 계속 받아들이므로 옛 실행 줄을 붙여 넣어도 의도한 run에 안착한다.

**GLM 표면이 GLM-5.3에 맞춰진다.** GLM-5.3에서는 추론을 끌 수 없고 `reasoning_effort`가 `{low, high, max}`를
받는다. 그래서 예전의 thinking-off 상태는 "가장 약한 설정"이 아니라 **닿을 수 없는 설정**이며, 그것을 요구하는
요청은 실패한다. 붕괴 바닥을 실제 와이어 값으로 옮기고 `glm-5.2`를 티어 슬롯 집합에서 내렸다. 동반 동작 변경
하나: Claude effort `low`에서 이제 `ANTHROPIC_REASONING_EFFORT=low`를 실제로 내보낸다(예전에는 키를 아예
빼먹었고, GLM-5.3에서는 그 부재가 실패 케이스다).

**조용히 죽던 실패 셋이 소리를 낸다.** `moai gate`의 Node 테스트 단계는 watch 모드 스크립트에서 늘 타임아웃까지
매달렸고, 소비 프로젝트는 그 때문에 게이트를 우회하고 있었다 — 이제 실행 직전에 `package.json`을 읽어 3단계로
실행형 명령을 고른다. `moai hook worktree-create`는 저장소 밖 cwd에서 stdout·stderr 0바이트로 죽어 스폰이
왜 실패했는지 아무도 알 수 없었다. 뿌리는 같았다: `*exec.ExitError`가 우연히 `ExitCoder`를 만족해 두 경계가
그것을 "의도된 종료 코드"로 오인했다. 이번에는 호출부가 아니라 경계에서 고쳤다.

이 밖에 `/moai goal`이 복구 불가능한 API 오류로 턴이 죽으면 스스로 무장을 해제한다(재시도로 풀리는 오류에서는
해제하지 않는다 — 그쪽이 더 조용한 결함이다), 레퍼런스 디자인을 Design DNA로 분해하는 `moai-domain-design-dna`
스킬이 추가되며, `moai-ref-ui-polish`가 모션 추론 계층과 실행 가능한 디자인 감사 스위트를 얻는다. 문서는
13페이지 × 4로케일로 정리됐다.

### English

**MoAI-ADK v3.1.1**

Three themes dominate this release.

**The lead session gets a name.** `manager-kanban` becomes `manager-lead`, and the role widens from kanban-only to
kanban + factory lead coordination. Companion sessions take the bare role names `plan` / `run` / `sync`, and
factory workers are labelled `lane-<n>`. The lead itself is now simply `lead` rather than `lead-<run-id>` — nothing
functional was ever downstream of carrying a run id in a name. When two leads share a machine the second takes
`lead-1`, so every session stays addressable by name alone, and a legacy `lead-<run-id>` launch line is still
adopted so an operator pasting an old command lands on the run they meant.

**The GLM surface is retargeted at GLM-5.3.** Under GLM-5.3 reasoning cannot be disabled and `reasoning_effort`
takes `{low, high, max}`. The old thinking-off state is therefore not the weakest setting but an unreachable one,
and a request still asking for it fails. The collapse floor moves to a wire-real level and `glm-5.2` is withdrawn
from the tier-slot closed set. One behavior change rides along: Claude effort `low` now actually emits
`ANTHROPIC_REASONING_EFFORT=low`, where it previously emitted no key at all — under GLM-5.3 the absent key is the
failing case.

**Three failures that used to die quietly now make noise.** `moai gate`'s Node test step hung to the timeout on any
watch-mode script, and consumer projects were bypassing the gate because of it — the step is now resolved to a
run-form command from `package.json` immediately before execution, across three tiers. `moai hook worktree-create`
exited with zero bytes on both stdout and stderr from a non-repo cwd, leaving a dead agent spawn with no cause at
all. Both shared one root: `*exec.ExitError` happens to satisfy `ExitCoder`, so two seams mistook it for a
deliberate exit code. This time the fix is at the seam, not at the call site.

Elsewhere: `/moai goal` disarms itself when a turn dies on an unrecoverable API error — and deliberately does not
disarm on errors that resolve on retry, which would be the quieter defect; a new `moai-domain-design-dna` skill
deconstructs a reference design into a portable Design DNA profile; `moai-ref-ui-polish` gains a motion-reasoning
layer and an executable design-audit suite; and the documentation lands 13 pages across 4 locales.

---

## 4. 분류에서 뺀 커밋과 그 이유

| 커밋 | 이유 |
|---|---|
| `ce310ca82` | 테스트 격리 수정 (팩토리 환경변수 오염). 사용자 대면 동작 변화 없음 |
| `e7aeec088` | statusline forge 카운트 스텁 CLI 커버리지 추가. 테스트 전용 |
| `f2326031c` | C1 리뷰 반영 테스트 개명 + `[WT]` 마커 위치 재적용. 상위 항목에 흡수 |
| `05378d75a` | t118 중간 스냅샷(폐기된 planner/leader 명명 포함). 최종 상태는 `310d75dd2`/`d73e9a669`가 대표 |
| `0e5239066` | 429로 중단된 부분 랜딩. `c25a418ab`가 완결하므로 문서 항목에 흡수 |
| 머지 커밋 15개 | 워크트리 통합 기록. 릴리즈 노트에 담을 내용 없음 |

## 5. 미검증 / 잔여 위험

- **범위가 확정이 아니다.** 배치 PR이 아직 열리지 않았고 `origin/release/v3.1.1`은 계속 움직일 수 있다.
  릴리즈 시점에 `git log --no-merges <base>..<tip>`을 다시 돌려 이 초안과 대조할 것.
- **커밋 메시지를 1차 근거로 삼았다.** 분류와 서술은 커밋 본문과 diffstat 기준이며, 각 항목의 런타임 동작을
  직접 실행해 확인하지는 않았다. 특히 GLM-5.3 와이어 동작(`ANTHROPIC_REASONING_EFFORT=low` 실제 반영)은
  이 배치에서도 실측되지 않았다 — 기존 SPEC-MODEL-TIER-PLANTYPE-001 유보 사항이 그대로 살아 있다.
- **1차 배치 3개 SPEC 항목의 `sync_commit_sha`가 placeholder다** (`pending-backfill-*`). 태깅 전에 backfill이
  끝났는지 확인이 필요하다 — 이 카드 범위 밖.
- **`024536e38`의 "36 commands" 표기가 릴리즈 시점 확인 대기 상태**로 남아 있다(원시 help 카운트 52,
  빌트인·별칭 포함). 문서 카드가 리드에게 보고한 항목.
- **`SPEC-HARNESS-GATE-TEST-001`은 `.moai/specs/` 아래에 디렉터리가 없다.** 배치 tip 기준 이 ID는
  `internal/hook/quality/gate.go`와 `gate_node_resolution_test.go` 두 파일에서만 참조된다
  (`git grep -ln "SPEC-HARNESS-GATE-TEST-001" 5798bdc2e`). 그래서 초안에서는 다른 SPEC 항목과 달리
  `.moai/specs/…/spec.md` 링크를 걸지 않고 ID만 남겼다 — 링크를 걸면 깨진다. SPEC 문서 자체가 누락된 것인지
  다른 리포에 있는 것인지는 이 카드 범위 밖이며, 별도 확인이 필요하다.
