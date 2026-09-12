# iter4 dispatch brief — SPEC-MOAI-GATEWAY-001 (re-issue)

Orchestrator dispatch record. Not an audit report.

- The first dispatch, `spec-gateway-iter4`, failed on a session usage limit before writing anything. Its idle notice arrived at 2026-09-10T11:37:37Z with the message "You've hit your session limit · resets 9:20pm (Asia/Seoul)".
- The orchestrator then measured the tree:
  - the six artifacts are byte-identical to 0.3.2: spec 36811, plan 24815, acceptance 27705, design 41292, research 52599, progress 1634 (184,856 B in total);
  - `version: "0.3.2"`, and `moai spec lint` is clean;
  - 25 live REQs, 25 ACs, 0 orphans;
  - git HEAD is `d060e0d13`, and the only changes are the three untracked `.moai/` directories.
- It re-issued the brief at 2026-09-10 22:03 KST, after the reset, as `spec-gateway-iter4b`. The re-issue changes two things:
  - it narrows the reading;
  - it adds a per-file save order, so that a second interruption cannot leave the work invisible.

---

Revise SPEC-MOAI-GATEWAY-001 from 0.3.2 to 0.4.0. This is the iter4 delta. The iter3 plan audit returned FAIL 0.84 against the 0.85 threshold. The operator made one explicit exception to the three-iteration cap: one more revision, followed by an audit of the changed parts only. Replace contradicting sentences; do not annotate beside them.

At start, invoke Skill("moai-workflow-spec") for the GEARS format.

## 0. Tree, ownership, prohibitions

- Worktree root, and your CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. Branch `WT-unified-gateway`, HEAD `d060e0d13`. There are no commits.
- **You are the sole writer of `.moai/specs/SPEC-MOAI-GATEWAY-001/`.** Do not message any agent. Do not touch code or `.moai/reports/`. Run no git command that changes state.
- Use `/usr/bin/grep`. The shell `grep` is a ugrep wrapper that silently skips files.
- **Open every cited code line before citing it.** Where the code differs from this brief, trust the code and report the difference.

## 1. Budget your reading. Do not read whole artifacts first.

A previous attempt ran out of session budget while still reading and wrote nothing. Read ONLY these:

- the advisory section of `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter3.md` (G3-A1 through G3-A10). G3-B1 and G3-B2 are fully restated below;
- in `spec.md`: frontmatter and HISTORY; §A around `:180-200`; the REQ-MG-008, REQ-MG-018, REQ-MG-021, and REQ-MG-022 blocks; and `:355-400`;
- in `design.md`: §5.4 and §6;
- in `acceptance.md`: AC-MG-006 and AC-MG-018;
- in `plan.md`: milestone M7 and the measurement items;
- in `research.md`: `:1-20` and `:600-615`;
- `progress.md` (it is small).

Locate each block with `/usr/bin/grep -n` first, then read only that range.

## 2. Save order, so an interruption shows

- Finish one file, save it, then move to the next. Order: `spec.md` body → `design.md` → `acceptance.md` → `plan.md` → `research.md` → `progress.md`.
- **Leave `version: "0.3.2"` and the HISTORY section untouched until every other edit is saved.** Then bump the version to `"0.4.0"` and add the HISTORY 0.4.0 entry as the final write. If the version on disk still reads 0.3.2, the pass did not finish.

## 3. G3-B1, operator decision: tmux session env gets the gateway address only

These code facts were verified by the orchestrator this run.

- **`moai glm` inside tmux writes GLM routing into the tmux session environment today.**
  - `applyGLMMode` calls `injectTmuxSessionEnv` at `internal/cli/launcher.go:278-284`. The comment there calls it the "`moai cg` path".
  - `buildTmuxInjectVars` (`internal/cli/glm.go:507-518`) includes `ANTHROPIC_AUTH_TOKEN` (the GLM key), `ANTHROPIC_BASE_URL` (Z.AI), and the `ANTHROPIC_DEFAULT_*_MODEL` slots.
  - `injectTmuxSessionEnvVia` (`glm.go:546-567`) sends the token through `mgr.InjectSensitiveEnv`, never argv (CWE-214).
- **`ensureTmuxGLMEnv`** is defined at `internal/hook/glm_tmux.go:78` and called at `internal/hook/session_start.go:638`.
  - It injects only when TMUX is set, `teammateMode == "tmux"`, and `settings.local.json` holds `ANTHROPIC_AUTH_TOKEN`.
  - Under REQ-MG-021 that file holds no GLM token after a gateway launch, so this writer is naturally inert there.
- **The two tmux clearers disagree about `ANTHROPIC_AUTH_TOKEN`.**
  - The CLI `clearTmuxSessionEnv` (`glm.go:574-585`, whose key list is `buildTmuxClearVars`) EXCLUDES it. Its comment: "it may be an OAuth token that must survive mode switches".
  - The SessionEnd `clearTmuxSessionEnv` (`internal/hook/session_end.go:634-665`, list `glmEnvVarsToClean` = AUTH_TOKEN, BASE_URL, OPUS, SONNET, HAIKU) CLEARS it. Its comment: "the user's real Claude credential is stored in ~/.claude/ ... not in the tmux environment, so unsetting the tmux var is always safe". It runs **unconditionally** at `session_end.go:93`, apart from the TMUX check.
  - `applyCCMode` calls the CLI clearer (`launcher.go:224`).
- **Unmeasured:** whether tmux teammate panes actually inherit the tmux session env for Claude routing. This is inferred from code comments and warning text, not run.

The contract to write. Fold it into existing REQ/AC text; do not add a REQ or an AC.

1. **What a gateway launch writes to tmux** (all three launchers, inside tmux).
   - It writes ONLY `ANTHROPIC_BASE_URL` pointing at the loopback gateway, and `MOAI_LAUNCH_PROVIDER`.
   - It also writes the session access-token carrier if that turns out to be delivered through tmux. The carrier is deferred to M0/T09. If it is `ANTHROPIC_AUTH_TOKEN`, it goes through the sensitive channel, never argv.
   - **It never writes a GLM credential or the Z.AI URL.**
   - For gateway launches, this replaces `applyGLMMode`'s tmux injection.
2. **Stale-key clearing at launch.**
   - Before injecting, a gateway launch removes the stale GLM routing keys that `buildTmuxInjectVars` writes.
   - A stale GLM `ANTHROPIC_AUTH_TOKEN` in tmux must be overwritten or removed, whatever the carrier decision turns out to be.
   - Resolve the two clearers' disagreement explicitly. Cite both comments and state which rationale the gateway contract adopts, and why. If you cannot, state the dependency honestly and tie it to M0.
3. **Hook suppression when `MOAI_LAUNCH_PROVIDER` is present.**
   - `ensureTmuxGLMEnv` makes no tmux writes. This is defense in depth.
   - The SessionEnd tmux clear (`session_end.go:93`) is suppressed. Teammate panes are Claude sessions that run their own SessionEnd, so an unconditional clear when one teammate exits would erase the loopback address the other live panes rely on. Teammates receive `MOAI_LAUNCH_PROVIDER` through the tmux injection.
   - Add both to the While suppression list in REQ-MG-021 and to the table in design §5.4.
4. **Gateway lifetime, folded into the existing REQ-MG-008.**
   - The gateway child keeps serving until the lead Claude session has ended AND no tmux teammate pane that was given its address is still alive.
   - Where teammate liveness cannot be determined, the contract is loud failure: the teammate's loopback request fails with a connection error. Silent bypass is structurally impossible, because tmux holds no Z.AI address.
   - Record the case where a teammate outlives the gateway as residual risk.
   - Add two plan measurement tasks: teammate tmux-env inheritance, and teammate-pane liveness detection. These are tasks, not `[NEEDS CLARIFICATION]` markers.
5. **Consistency.** Update every sentence that this decision contradicts or leaves open:
   - `spec.md` §A around `:188-189`, which names both surfaces;
   - "교체하는 것은 mode switch의 env 주입 단계", around `:196`;
   - "GLM credential은 gateway child에만", around `:365-367`: generalize it so it covers the tmux surface;
   - REQ-MG-022's bypass-path clause: add the tmux path;
   - REQ-MG-018: teammate UX is preserved through the gateway;
   - design §5.4 and §6: add the tmux surface;
   - AC-MG-018 (a)/(c): extend with tmux-surface checks using a fake `tmux.SessionManager` (`injectTmuxSessionEnvVia` already accepts one);
   - AC-MG-006, the lifecycle AC for REQ-MG-008: extend with the teammate-liveness and loud-failure contract.

## 4. G3-B2: the `cleanupGLMSettingsLocal` suppression check in AC-MG-018 (c) is vacuous

- The launch strip has already removed `ANTHROPIC_BASE_URL`, so the function returns early at `session_end.go:728-731` whether or not it is suppressed. The check therefore passes either way.
- The fix:
  - Seed a fixture whose `settings.local.json` `env` still HAS `ANTHROPIC_BASE_URL` plus a representative GLM key when SessionEnd runs. This models a concurrent session rewriting the file.
  - Assert that those keys survive under `MOAI_LAUNCH_PROVIDER`.
  - Add a no-signal control: the same fixture without `MOAI_LAUNCH_PROVIDER` must have the keys removed.

## 5. Advisories

- **G3-A1 (verified).** `ensureTeammateMode` deletes the legacy `CLAUDE_CODE_TEAMMATE_DISPLAY` from `settings.local.json` `env`, but only when it rewrites `teammateMode`. It returns early when the value already matches (`internal/hook/session_start.go:1033-1047`). Reflect this wherever AC-MG-018's exclusion or strip reasoning depends on that key.
- **G3-A2 (verified).** The release-PR artifact has `retention-days: 7` (`.github/workflows/release-pr-multi-os.yml:219`). Name who reads the Windows verdict, and where it is recorded durably within that window.
- **G3-A7.** Change "both values changed" to "either".
- **G3-A3 to A6 and A8 to A10.** Apply each unless it conflicts with the decisions above, and report a disposition for each.
- **Leftovers.**
  - `research.md` around `:611`, "`MOAI_HOME` 우회": add the `filepath.IsAbs` condition (`internal/paths/paths.go:69`).
  - `research.md:12`: the list of re-read revisions omits 0.3.2.
  - `progress.md`: it does not mention 0.3.2.

## 6. Invariants

- Frontmatter at the end: `version: "0.4.0"`, `status: draft`, `updated: 2026-09-10`. All 12 required fields present.
- The HISTORY 0.4.0 entry is written last. Earlier entries stay unaltered.
- REQ 25/25 and AC 25/25. `[NEEDS CLARIFICATION]` = 0.
- No capability claim regresses into asserting unmeasured behavior. Tmux inheritance, teammate liveness, and settings-env precedence all stay labeled unmeasured.
- Korean, clean native written register. Identifiers, paths, and code stay verbatim.

## 7. Self-check before you return (run it, quote the output verbatim)

- No sentence permits a GLM credential or the Z.AI URL in tmux under a gateway launch. Past-tense descriptions of current code are fine.
- `ensureTmuxGLMEnv` and the SessionEnd `clearTmuxSessionEnv` both appear in the REQ-MG-021 While clause and in design §5.4.
- REQ-MG-008 carries the teammate-lifetime contract and the loud-failure fallback.
- AC-MG-018 (c) has the seeded `ANTHROPIC_BASE_URL` fixture and its no-signal control.
- The traceability parse shows 25 live REQs, 25 ACs, 0 orphans, and 0 references to undefined or retired REQs.
- `[NEEDS CLARIFICATION]` count per file.
- `moai spec lint SPEC-MOAI-GATEWAY-001`, output verbatim.
- `/usr/bin/grep -n '^version' .moai/specs/SPEC-MOAI-GATEWAY-001/spec.md` prints `0.4.0`.

## 8. Return

- A disposition, with the new file:line, for each of G3-B1 items 1–5, G3-B2, G3-A1 through A10, and the three leftovers.
- How the two clearers' `ANTHROPIC_AUTH_TOKEN` disagreement was resolved, or the dependency recorded instead.
- REQ and AC counts, and the traceability result.
- The self-check outputs and the lint output, verbatim.
- Anything not satisfied, and why.
