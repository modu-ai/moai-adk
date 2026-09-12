# iter6 dispatch brief — SPEC-MOAI-GATEWAY-001 0.6.0 → 0.7.0

Orchestrator dispatch record (session 650cbd95, 2026-09-11). Not an audit report.

Revise SPEC-MOAI-GATEWAY-001 from 0.6.0 to 0.7.0. The iter5 plan audit (`.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md`) returned FAIL 0.83. The operator explicitly authorized one more bounded fix followed by a delta-only re-audit of G5-B1, G5-B2 and G5-B3. Replace contradicting sentences; do not annotate beside them. SPEC prose stays Korean in clean native written register.

At start, invoke Skill("moai-workflow-spec") for the GEARS format.

## 0. Tree, ownership, prohibitions

- Worktree root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. Branch `WT-unified-gateway`, HEAD `81c1d58f9` (fast-forwarded from `ed71054d3` by the orchestrator; +65 commits, no local commits).
- **You are the sole writer of `.moai/specs/SPEC-MOAI-GATEWAY-001/`.** Do not message any agent. Do not touch code or `.moai/reports/`. Run no git command that changes state (read-only `git rev-list`, `git show`, `git diff` are fine).
- Use `/usr/bin/grep`. The shell `grep` is a ugrep wrapper that silently skips files.
- **Open every cited code line before citing it.** Where the code differs from this brief, trust the code and report the difference.
- Budget is fixed: 24 live REQ / 24 AC, ceiling 25/25. **Add no new REQ or AC number.** Tombstones unchanged. Every fix fits inside REQ-MG-019, REQ-MG-021, REQ-MG-022, AC-MG-003, AC-MG-018 and their design/plan text.
- If a fix cannot be written without an operator decision not given below, stop that item and return a blocker report for it; finish the rest.

## 1. Reading budget — locate with `/usr/bin/grep -n`, read only these ranges (line numbers are 0.6.0)

- iter5 report: `plan-audit-iter5.md:117-178` (D1-D7). Everything else is restated below.
- `spec.md`: frontmatter; HISTORY 0.6.0 entry (`:229-305`); §C (`:354-367`); REQ-MG-019 (`:486-496`); REQ-MG-021 (`:497-568`); REQ-MG-022 (`:569-575`); §E (`:623-652`).
- `design.md`: §3.2 (`:132-160`); §6.1 (`:472-529`); §6.5 (`:587-624`); §6.6 (`:625-667`); §7.4 (locate).
- `acceptance.md`: AC-MG-001 (`:18-36`); AC-MG-003 (`:37-70`); AC-MG-012 (`:164-170`); AC-MG-018 (`:219-280`).
- `plan.md`: M1 (`:91-125`); M3 (`:139-168`); M7 (`:210-247`); §H decision list (locate decisions 10-12).
- `research.md`: `:910-920`, `:960-970`.
- `progress.md` (small).

## 2. Save order, so an interruption shows

Finish one file, save it, then the next: `spec.md` body → `design.md` → `acceptance.md` → `plan.md` → `research.md` → `progress.md`. **Leave `version: "0.6.0"` and HISTORY untouched until every other edit is saved.** Then bump to `"0.7.0"` and add the HISTORY 0.7.0 entry as the final write.

## 3. G5-B1 — scrub inherited GLM keys from the Claude child env (operator: fix (i), with a Z_AI_API_KEY split)

Code facts verified by the orchestrator at `81c1d58f9`:
- The Claude child env is built from the launcher's inherited process env: `buildEnvForLaunch(effectiveEffort, os.Environ())` (`internal/cli/launcher.go:837`); `buildEnvForLaunch` (`:1180-1200`) only replaces `CLAUDE_CODE_EFFORT_LEVEL`. The GLM branch is `buildEnvForGLMLaunch(..., os.Environ())` (`:833-834`).
- Today's launch-time tmux clear is `applyCCMode` → `clearTmuxSessionEnv` (`launcher.go:224`); decision 12 removes it for gateway launches.
- Every key the two tmux GLM writers put into the tmux session env is inside the 14-key GLM cleanup key set (`design.md` §6.1): `buildTmuxInjectVars` (`internal/cli/glm.go:507-539`: AUTH_TOKEN, BASE_URL, four DEFAULT_*_MODEL slots, DISABLE_EXPERIMENTAL_BETAS, API_TIMEOUT_MS, MOAI_STATUSLINE_CONTEXT_SIZE, AUTO_COMPACT_WINDOW, MAX_CONTEXT_TOKENS) and `glmTmuxKeys` (`internal/hook/glm_tmux.go:20-30`). So the scrub can key off that single set definition — do not introduce an ad hoc list.
- Today's `moai glm` sets `Z_AI_API_KEY` in its own process env (`internal/cli/glm.go:386`) because `moai glm tools` registers Z.AI MCP servers — by default in the user-scope `~/.claude.json` (`internal/cli/glm_tools.go:188`, `:371-376`) — whose HTTP header is the literal `Bearer ${Z_AI_API_KEY}` expanded by Claude Code from its process env (`glm_tools.go:58-60`, `:404-411`; expansion is per code comment, not run live).
- GLM upstream auth in this SPEC comes from the GLM `CredentialRef` reading `internal/glmcred` in the gateway child (`design.md:536`, `:753`), not from the Claude child env.

Write:
1. **REQ-MG-021** — add that every gateway launcher removes the GLM cleanup key set from the *inherited* process env before assembling the Claude child env. Process-env projection: all 14 keys are deleted (there is no backup restore in a process env; `MOAI_BACKUP_AUTH_TOKEN` is deleted too). The scrub applies to the inherited base, **before** the launcher's own additions (loopback `ANTHROPIC_BASE_URL`, `MOAI_LAUNCH_PROVIDER`, session access token, effort key, other launcher-set keys), so those survive.
2. **Z_AI_API_KEY — operator decision:** for `moai cc` and `moai gpt` the inherited `Z_AI_API_KEY` is removed (a user-scope Z.AI MCP entry would otherwise call paid Z.AI from a session the user did not point at GLM — REQ-MG-022). For gateway `moai glm` it is kept, but only as the value the launcher sets from the GLM credential store for MCP tool authentication, never an inherited value. Rewrite REQ-MG-021's "GLM credential must not appear in the Claude child env" clause so this narrow exception is explicit, names why (the MCP header expansion above), and keeps every other GLM credential/routing key out of all three Claude child envs. Record in §E / design §6.5 that under `moai glm` the Z.AI key is present in the Claude process env for MCP tools.
3. **ANTHROPIC_AUTH_TOKEN** — the inherited value is removed in all three launchers. The session access token carrier key is still undecided until M0 (`design.md:151-156`): state both branches (carrier = `ANTHROPIC_AUTH_TOKEN` → the launcher sets its own value after the scrub; carrier = separate header → no `ANTHROPIC_AUTH_TOKEN` in the child env, and the `design.md:589-596` "user token rides loopback requests" residual changes accordingly — update that bullet). Do not claim any Anthropic credential sourcing you did not read in the SPEC.
4. **`moai glm` model slots** — the gateway `moai glm` Claude child must not carry inherited GLM `ANTHROPIC_DEFAULT_*_MODEL` IDs. If the SPEC's glm initial-model mechanism (REQ-MG-019, M3, §6.7 moved to PICKER) actually requires slot keys set by the launcher, say so precisely; if that creates a contradiction, report it rather than resolving it.
5. **design.md** — §3.2 step 6: add the scrub as the first sub-step. Correct `design.md:618-619` so the "does not reach the Claude child" claim cites the REQ-MG-021 scrub; keep the residual that a `claude` the user starts in a new tmux pane still inherits the stale keys.
6. **AC-MG-018 (a)** — add a variant: launcher process env pre-seeded with a Z.AI `ANTHROPIC_BASE_URL`, a GLM `ANTHROPIC_AUTH_TOKEN`, GLM `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL` IDs, one layer-(b) key, `MOAI_BACKUP_AUTH_TOKEN`, and `Z_AI_API_KEY`. Judge the exec env: for `moai cc` and `moai gpt` none of those inherited values appears; for `moai glm` none appears except `Z_AI_API_KEY`, whose value equals the credential-store fixture value, not the pre-seeded one. The loopback base URL and `MOAI_LAUNCH_PROVIDER` are present in all three. Name a mutant it kills (e.g. a builder that passes `os.Environ()` through).

## 4. G5-B2 — couple the core release to the PICKER sibling (operator: option (a))

- In `plan.md` M3 "출시 판단 입력" (`:164-166`), add that exposure and release also wait for `SPEC-MOAI-GATEWAY-PICKER-001` (proposed), with the reason: until it lands, an empty-default `moai cc` starts on Claude Code's persisted `/model` choice (probe 1, `research.md` §15), so a gateway model ID can become the start model while the provider is recorded as `claude` (REQ-MG-026), and a plain `claude` can send that ID to Anthropic. Mirror the coupling wherever the GPT-AUTH coupling is stated (design §7.4, spec §E or §G — locate).
- In REQ-MG-019 or §E, state the core-only behaviour for the empty default in one or two sentences and point to the coupling. Keep AC-MG-001's empty-case exclusion, with a pointer to the sibling and the coupling.

## 5. G5-B3 — reword REQ-MG-022 (operator: option (a), no new flag)

Replace "…in-process로 한정하는 것으로 막는다" (`spec.md:573-575`) with a statement that the in-process limit narrows but does not close that path, naming the residuals: `teammateMode` lives in a shared project file that a later non-gateway SessionStart can rewrite to `tmux` (`internal/hook/session_start.go` — re-anchor at `81c1d58f9`), the read timing is unmeasured, and gateway launches do not clear stale tmux GLM keys; closing it belongs to `SPEC-MOAI-GATEWAY-TEAMMATE-001` (proposed). Keep it consistent with design §6.5 / §6.6. Do not adopt `--teammate-mode`.

## 6. Advisories to fold in

- **G5-A1**: AC-MG-018 (a) — add a no-`TMUX` variant starting from `teammateMode: "auto"`, asserting `in-process` after the SessionStart chain.
- **G5-A2**: AC-MG-003 (c) — add variants `stream: null`, `max_tokens: 0`, and a single `system`-role message; update every "여섯 변형" count (e.g. `plan.md:117`) to the new number.
- **G5-A3**: `research.md:917` — re-point "결정 10의 게이트(`plan.md` M1)" to `SPEC-MOAI-GATEWAY-PICKER-001`.
- **G5-A4**: `plan.md:108` — name the tracked test-fixture path for the raw capture and the tracked evidence path for the gate's comparison result, following repository conventions (check an existing gateway-adjacent `testdata/` layout before naming).

## 7. Re-anchor citations moved by the fast-forward

Only two SPEC-cited files (cited by full path) changed between `ed71054d3` and `81c1d58f9`:
- `internal/hook/session_end.go`: +1 line at 16, +4 lines after old 261. Old lines 16-261 shift +1; old lines ≥262 shift +5. Known citations: `design.md:456` (`:93`→`:94`, and "정의 `:646`"→`:651`), `design.md:457` (`:634-640`→`:639-645`), `acceptance.md:277` (`:728-731`→`:733-736`), `acceptance.md:314` (`:653`→`:658`), `research.md:124` (`:671`→`:676`), `research.md:125` (`:696`→`:701`), `research.md:261` (`:95-104`→`:96-105`), `research.md:375` and `:393` (`:735`→`:740`), `research.md:450` (`:733-744`→`:738-749`), `research.md:480` (`:682`→`:687`), `research.md:512` (`:634-640`→`:639-645`). Open each target line and confirm before writing.
- `.github/workflows/ci.yml`: +1 after old 67, +1 after old 90. `design.md:213` `:108-110`→`:110-112`.
- Any citation you touch for the fixes above: open the line at `81c1d58f9` first.

## 8. progress.md and HISTORY

- `progress.md` §E.1: add the 0.7.0 line (operator decisions 2026-09-11: B1 fix (i) with the Z_AI_API_KEY split, B2 (a), B3 (a); A1-A4 folded), add the iter5 report path, set baseline HEAD `81c1d58f9` with the measuring command `git rev-list --count --left-right origin/develop...HEAD` and the output you observe (run it), and state "0.7.0 작성 완료, 감사 미실시".
- HISTORY 0.7.0 (final write): same decisions and dispositions, concise.

## 9. Return (to the orchestrator, English)

- Per file: what changed (section + new line ranges).
- Verbatim output of: `moai spec lint SPEC-MOAI-GATEWAY-001`; a live REQ and AC count command of your choice with its output; `wc -c` of the six artifacts.
- Every place where code differed from this brief, and any contradiction or blocker left unresolved.
