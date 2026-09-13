# SPEC Review Report: SPEC-MOAI-GATEWAY-001
Iteration: 5 — operator-authorized final iteration (iter4 FAIL 0.80 with STOP signal → operator chose "scope split": one core revision plus this audit). No further iteration is authorized.
Verdict: FAIL
Overall Score: 0.83 (Tier L PASS threshold 0.85). iter4 0.80 → iter5 0.83: no score regression, so this report emits no STOP signal. Escalation to the operator is still required because this was the last authorized iteration.

Reasoning context ignored per M1 Context Isolation. The dispatch's description of what 0.6.0 claims was treated as claims to check against the artifacts, not as evidence.

Audited artifacts (Tier L, all five read by targeted range): `spec.md` 62636 B, `plan.md` 43728 B, `acceptance.md` 41008 B, `design.md` 63762 B, `research.md` 71162 B, plus `progress.md` 4104 B. Worktree `.claude/worktrees/moai-proxy-unified`, branch `WT-unified-gateway`, HEAD `ed71054d3`. The SPEC directory is untracked.

---

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency.** 26 definition lines `REQ-MG-001`…`026`, contiguous, no duplicates. Two are tombstones: `REQ-MG-007` (`spec.md:413`) and `REQ-MG-020` (`spec.md:492`). The live pattern `^\*\*REQ-MG-[0-9]{3}\*\* \(` matches 24 IDs (E3).
- [PASS] **MP-2 GEARS compliance.** Judged on the requirement layer only (`spec.md` §D). Every live REQ carries a GEARS label and form. The REQs changed in 0.6.0:
  - `REQ-MG-006` When + Ubiquitous (`:406`)
  - `REQ-MG-008` Event-driven (`:417`)
  - `REQ-MG-018` Ubiquitous (`:477`)
  - `REQ-MG-019` Ubiquitous (`:486`)
  - `REQ-MG-021` Ubiquitous + While (`:497`)
  - `REQ-MG-022` Unwanted (`:569`)
  - `REQ-MG-023` Event-driven (`:577`)

  Tombstones declare that they carry no pattern. ACs are Given-When-Then and were not graded against this rubric. Both lint builds report no findings (E1).
- [PASS] **MP-3 frontmatter.** `spec.md:2-14`: `id`, `title`, `version: "0.6.0"` (quoted), `status: draft`, `created: 2026-09-10`, `updated: 2026-09-11`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (string), `tier: L`. No rejected aliases.
- [N/A] **MP-4 language neutrality.** The SPEC targets moai-adk-go's own Go internals (`module:` internal/gateway, internal/cli, internal/kanban, internal/hook), not template-bound content.
- [PASS] **MP-5 D7.** The verb emitted no BLOCKING result, only five SHOULD "not found" notices, each explained:
  - `SPEC-MOAI-GPT-AUTH-001`, `SPEC-MOAI-CG-RETIRE-001`, `SPEC-MOAI-GATEWAY-PICKER-001` and `SPEC-MOAI-GATEWAY-TEAMMATE-001` are marked "(제안)" (proposed). `progress.md` states the two new sibling directories were deliberately not created.
  - `SPEC-MOAI-PROXY-001` is the preserved iter1 report path (`spec.md:739`).
- [PASS, judgment] **MP-6 D8.** Read literally, the verb prints BLOCKING (`syscall` = 6 hits, `//go:build`/exemption = 0, E4). This audit follows the iter1–iter4 judgment: every mention preserves the existing POSIX `syscall.Exec` (`spec.md:399-400`, §E `:625`). The SPEC introduces no new syscall use. Stated explicitly so the literal BLOCKING is not read as silently absorbed.
- [PASS] **MP-7 clarification gate.** `NEEDS CLARIFICATION` count is 0 in all six files (E4).

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Child-env assembly (`design.md:142-144`) lists additions only. Two readings are both "compliant" and route differently: pass the inherited env through (as today's `buildEnvForLaunch` does, `internal/cli/launcher.go:1180-1200`), or strip inherited GLM keys (G5-B1). The core-only outcome for a persisted `/model` default is unstated (G5-B2) |
| Completeness | 0.85 | 0.75–1.0 | All sections and eight Out-of-Scope H3 headings present (`spec.md:666-733`); sibling routing complete. Gaps: the release-coupling input exists for GPT-AUTH (`plan.md:164-166`) but not for PICKER; the stale-tmux residual consequence is misdescribed (`design.md:618-619`) |
| Testability | 0.85 | 0.75–1.0 | `AC-MG-003` (c) now kills all three iter4 mutants (`acceptance.md:57-62`), with a reaching control (`:64-65`). Gaps: the `REQ-MG-021` ban on a GLM credential in the Claude child env (`spec.md:520-522`) has no judgment, because `AC-MG-018` (a) checks only the child base URL (`acceptance.md:247`); outside-tmux `teammateMode` is unjudged (G5-A1); minor recognizer mutants remain (G5-A2) |
| Traceability | 0.90 | 0.75–1.0 | 24 live REQs, 24 live AC headers, and the REQ set cited by AC headers equals the live REQ set (`diff` exit 0, E3). Deductions: the two unjudged clauses above |

Harmonic mean = 4 / (1/0.75 + 1/0.85 + 1/0.85 + 1/0.90) ≈ 0.834, reported as 0.83.

---

## Answers to the orchestrator questions

### 1. Is the split clean?

**Mostly yes.** No live text depends on moved content:
- Every remaining `REQ-MG-020` mention is historical, a tombstone, or a move pointer: `spec.md:410` (the invariant moved into `REQ-MG-006`), `spec.md:213` (0.5.0 HISTORY), `plan.md:345` and `:418` (under the "[0.6.0 이관]" banner).
- `§6.7` survives only as pointers (`design.md:487`, `design.md:668-671`).
- Decision 10 appears only under its banner (`plan.md:405-418`) and in M3's hand-off sentence (`plan.md:152-155`).
- The tmux injection contract is gone from live REQs; `REQ-MG-021` now forbids tmux writes and clears (`spec.md:532-534`).
- `AC-MG-005` keeps its dependency: the `~/.claude/settings.json` exclusion is live in `REQ-MG-006` (`spec.md:409-411`).
- `AC-MG-003` still exercises picker `s`, but M1 declares that a single custom picker entry is test input, not the product picker (`plan.md:115-116`). That is acceptable.
- One stale pointer: `research.md:917` still names "the decision-10 gate (`plan.md` M1)" (G5-A3).

**Not stated: what a core-only release does.** Nothing couples the core release to the PICKER sibling, and M3 keeps today's omit-`--model`-when-empty behaviour (`plan.md:152-155`). In a core-only release, three consequences follow (G5-B2):
- In the default profile, a `/model <GPT or GLM ID>` in any gateway session is persisted by Claude Code. `research.md` §15 records the persistence.
- The next default-profile `moai cc` starts on that non-Claude model.
- A plain `claude` start then sends a gateway-only model ID to Anthropic.

Decision 10 existed to remove exactly this. Moving it out is legitimate scoping only if core either states that it cannot ship before the sibling, or records this as an accepted residual.

### 2. Is the M1 entry gate for decision 9 a genuine gate?

**Yes.**
- **When it runs:** the first task of M1; no recognizer or fixture is built before it (`plan.md:100-101`).
- **Failing input:** a raw, unprocessed capture (`:102`) taken without the four suppression flags (`:105`) whose validation request differs from the four conditions, or another captured request that meets all four.
- **Who sees the red:** implementation stops and a blocker goes to the orchestrator; the SPEC is fixed before code (`:110-111`). The rule is restated normatively in `REQ-MG-023` (`spec.md:590-594`).
- The capture must move into a tracked fixture (`:108`), and the client version is recorded (`:112`).

**The iter4 mutants are killed.**
- Mutant 2 (absent `stream` treated as `false`) is killed by the absent-`stream` variant (`acceptance.md:59`).
- Mutant 3 (no `role` check) is killed by the single-`assistant` variant (`:61`).
- A tools-ignoring recognizer is killed by the tools variant (`:62`).
- Mutant 1 (requires JSON `false`) is now either correct or caught by the capture gate before implementation.

The control makes the variants meaningful (`:64-65`). The ordering fix is present: the validation branch comes before registry and credential checks (`design.md:239`, `:254`). Residual mutants are minor (G5-A2).

### 3. Decision C

**No contradiction with locked decisions or live REQs.**
- `REQ-MG-018` declares the teammate UX change as an explicit exception (`spec.md:479-482`).
- Hook tmux suppression is retained and consistent with the no-write rule (`spec.md:545-558`).
- The non-regression clause "clears no less than today's `moai cc`" is scoped to the `settings.local.json` key list (`spec.md:512-514`). Gateway `moai cc` does clear fewer tmux keys than today (`internal/cli/launcher.go:224`), and the SPEC records that (`design.md:616-621`).

**The stale-GLM-keys residual is a blocker as written, not an acceptable residual.** Its consequence is misdescribed. `design.md:618-619` asserts that stale tmux GLM keys "do not reach the gateway session's Claude child". Nothing in the SPEC makes that true (G5-B1):
- A pane created after GLM keys were injected into the tmux session starts its shell with `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL` and the `ANTHROPIC_DEFAULT_*_MODEL` slots (`internal/cli/glm.go:507-514`).
- Today's child env is `os.Environ()` with only the effort key replaced (`launcher.go:837`, `:1180-1200`).
- The gateway child-env step lists additions only (`design.md:142-144`).

Separately, `REQ-MG-022` says the in-process limit "막는다" (blocks) the tmux-pane bypass (`spec.md:573-575`). `design.md:611-615` itself records that a later non-gateway SessionStart rewrites the shared `teammateMode` (G5-B3).

**Is `teammateMode: "in-process"` measured to prevent panes? No.** It is unmeasured, and the SPEC says so in `spec.md` §E (`:640-643`), `spec.md:535-536`, `spec.md:575` and `research.md` §16.2. M7 carries a stop-on-negative measurement (`plan.md:238-243`). The docs quote establishes what the setting means, not what Claude Code 2.1.267 does. The only overstatement is the word "막는다" in `REQ-MG-022`.

### 4. Honest counts and anchors

- Counts are honest: 24/24 under the 25/25 ceiling (E3), consistent in `plan.md` §I (`:451-452`), `progress.md:31`, `spec.md:303`, `spec.md:655` and `acceptance.md:412`.
- `progress.md:33` records "0.6.0 작성 완료, 감사 미실시" (authoring complete, not yet audited), and its baseline HEAD `ed71054d3` matches the measured HEAD.
- Earlier HISTORY entries appear intact in kind: the 0.5.0 entry still states decision 10 (`spec.md:221`), and the 0.6.0 entry explicitly declines to edit the 0.1.0 D1 paragraph (`spec.md:244-245`). Byte-level non-alteration cannot be verified because the directory is untracked (Gap).
- The `ensureTeammateMode` citations are exact: comment `:989`, definition `:995`, `desired := "auto"` `:1021`, tmux branch `:1023`, caller comment `:630` (E5). The comment-versus-code discrepancy is real and correctly recorded.

### 5. Unmeasured behaviour still presented as supported

Two items:
- The `design.md:618-619` absence claim (G5-B1).
- `REQ-MG-022`'s "막는다" (G5-B3).

Everything else is fenced as unmeasured.

---

## Defects Found (structured defect-list)

D1. **G5-B1 — inherited GLM keys reach the gateway session's Claude child.** `design.md:142-144`, `design.md:616-621`, `spec.md:520-522`, `acceptance.md:247`.
- **Mechanism.** Decision 12 removes the only launch-time clearing of tmux GLM keys (`launcher.go:224`). The SPEC then asserts those keys cannot reach the gateway session's Claude child. That assertion depends on scrubbing inherited GLM keys from the child env, and no REQ, design step or AC requires it:
  - §3.2 step 6 lists only additions.
  - Today's builder passes `os.Environ()` through (`launcher.go:837`, `:1180-1200`).
  - `REQ-MG-021` forbids a GLM *credential* in the Claude child env, but `AC-MG-018` (a) judges only the child base URL.
  - The four `ANTHROPIC_DEFAULT_*_MODEL` slot keys are not forbidden anywhere.
- **Failure scenario.**
  1. A tmux session carries GLM keys from a pre-upgrade `moai glm` launch (`glm.go:507-514`, written via `launcher.go:279`), or from a non-gateway session's `ensureTmuxGLMEnv`. After the upgrade, no gateway launch clears them.
  2. The user opens a new pane and runs gateway `moai cc`. The Claude child inherits `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL=<GLM IDs>` and the GLM `ANTHROPIC_AUTH_TOKEN`; only `ANTHROPIC_BASE_URL` is overridden.
  3. Every alias-model request (`opus`/`sonnet`/`haiku` — the form subagents use) carries a GLM ID. The registry routes by exact match (`REQ-MG-011`), so a `moai cc` session makes paid Z.AI calls the user did not select (`REQ-MG-022`), and the GLM key travels to loopback.
- **Evidence type.** Inference from code and `design.md`, plus tmux session-environment semantics (new panes inherit the session environment) and Claude Code's documented slot variables. Not run live.
- **Severity: critical. Class: blocking** (correctness of `REQ-MG-022`; an unsupported absence claim).
- **Operator decision:** not needed for fix (i); needed only for option (ii).
- **Required fix, one of:**
  - (i) Add to `REQ-MG-021` and `design.md` §3.2 step 6 that the gateway launcher removes the GLM cleanup key set and `Z_AI_API_KEY` from the *inherited* Claude child env before exec. Add an `AC-MG-018` (a) variant with the launcher process env pre-seeded with a Z.AI base URL, a GLM token and GLM slot IDs, asserting none appears in the exec env for `moai cc`, `moai gpt` and `moai glm`. Correct `design.md:618-619`.
  - (ii) Restore a launch-time tmux GLM-key clear for gateway launches. This amends decision 12 and needs the operator.

D2. **G5-B2 — the core-only release outcome for a persisted `/model` default is unstated, and release is not coupled to the PICKER sibling.** `spec.md:486-490`, `plan.md:152-155`, `spec.md:409-411`, `acceptance.md:28-30`.
- **The unstated outcome.** `REQ-MG-019` says `moai cc` starts with the configured Claude default. M3 keeps omitting `--model` when that default is empty, and `AC-MG-001` excludes the empty case (`acceptance.md:28-30`). In a core-only release, Claude Code's persisted `/model` choice therefore decides the default-profile start model, and nothing says so.
- **Failure scenario.**
  1. The user runs `/model gpt-5.6-sol` in any gateway session and confirms; Claude Code persists the choice (`research.md` §15, probe 1).
  2. The next default-profile `moai cc` starts on GPT. It records `MOAI_LAUNCH_PROVIDER=claude` and kanban backend `claude` (`REQ-MG-026`) while its first turn goes to GPT.
  3. A plain `claude` in any project sends `gpt-5.6-sol` to Anthropic.

  Decision 10 existed to prevent exactly this. M3 routes release timing to the GPT-AUTH sibling (`plan.md:164-166`) but not to the PICKER sibling.
- **Severity: major. Class: blocking** (`REQ-MG-019`'s stated start model does not hold in the documented normal state of an empty default; the core-only consequence is silent).
- **Operator decision: required.**
- **Required fix, one of:**
  - (a) Add a release-coupling input like `plan.md:164-166`: core is not exposed or released until `SPEC-MOAI-GATEWAY-PICKER-001` lands.
  - (b) Record the core-only behaviour in `REQ-MG-019` and §E as an accepted residual: the persisted default governs an empty-default `moai cc`, and a gateway model ID reaches the global Claude config through Claude Code's own write. Make `REQ-MG-019`'s `moai cc` clause conditional on a non-empty default.

D3. **G5-B3 — `REQ-MG-022` overstates the in-process mitigation.** `spec.md:573-575` versus `design.md:611-615` and `:616-621`.
- **The contradiction.** `REQ-MG-022` says limiting gateway teammates to in-process "막는다" (blocks) the tmux-pane Z.AI bypass. `design.md` §6.5 records that `teammateMode` lives in a shared project file, and that a later non-gateway SessionStart inside tmux rewrites it to `tmux` (`internal/hook/session_start.go:1021-1023`). Stale tmux GLM keys are never cleared.
- **Failure scenario.**
  1. In one project and tmux session, a gateway `moai cc` lead is running.
  2. A `moai cg` or plain `claude` session starts and sets `teammateMode: tmux`.
  3. If the lead reads the setting at spawn time (unmeasured, `design.md:612-613`), a split-pane teammate starts. It gets no gateway address, because the gateway writes nothing to tmux, but does get the stale Z.AI URL and GLM token from the tmux session env. It calls Z.AI directly — the path `REQ-MG-022` says is blocked.
- **Severity: major. Class: blocking** (inconsistency between `spec.md` and `design.md`; a requirement-level mitigation claim contradicted by the design's own residual).
- **Operator decision:** needed only for option (b).
- **Required fix, one of:**
  - (a) Reword `REQ-MG-022` to name the residual (shared-file overwrite, unmeasured read timing, stale tmux keys) instead of "막는다".
  - (b) Operator chooses a per-session mechanism: the launcher passes `claude --teammate-mode in-process` (documented, experimental, `research.md:967`) so the gateway session stops depending on the shared file. Add an argv judgment to `AC-MG-018` (a).

D4. **G5-A1 — outside-tmux `teammateMode` unjudged.** `spec.md:529-531`, `acceptance.md:250-262`.
- `REQ-MG-021` requires `in-process` "inside or outside tmux". `AC-MG-018` (a)'s fixture and control are both inside tmux (`TMUX` set).
- An implementation that forces `in-process` only when `TMUX` is set passes. Outside tmux it leaves today's `auto`, which per the docs quote (`research.md:964-965`) opens iTerm2 split panes when the `it2` CLI is installed.
- **Severity: minor. Class: optional** (cheap). **Fix:** add a no-`TMUX` variant starting from `teammateMode: "auto"`.

D5. **G5-A2 — residual recognizer mutants.** `acceptance.md:57-62`, `design.md:294`.
- The design table lists `stream: null` and the string `"false"` as not recognized. `AC-MG-003` (c) has no variant for `stream: null`, `max_tokens: 0`, or a single `system`-role message.
- A truthiness recognizer (`!stream`) accepts `null` and survives. A `max_tokens <= 1` recognizer survives.
- **Severity: minor. Class: optional.** **Fix:** add those three variants.

D6. **G5-A3 — stale pointer.** `research.md:917` still names "the decision-10 gate (`plan.md` M1)". M1 moved that gate to the PICKER sibling (`plan.md:121`).
- **Severity: minor. Class: optional.** **Fix:** re-point it to the sibling.

D7. **G5-A4 — capture evidence location unnamed.** `plan.md:108` requires the raw capture to move into "a tracked test fixture" but names no path. It also names no tracked evidence path for the gate's comparison result.
- **Severity: minor. Class: optional.**

---

## G4 Disposition Verification

| iter4 item | Claimed disposition | Verified | Evidence |
|---|---|---|---|
| G4-B1 tmux ownership | teammate sibling; removed from core by decision 12 | RESOLVED in core; the replacement introduces G5-B1 and G5-B3 | `spec.md:532-534`, `:705-717`; `design.md:625-666` |
| G4-B2 recognizer evidence | core fix + M1 entry gate | RESOLVED | `spec.md:590-594`; `design.md` §4.1 table `:294`; `plan.md:100-112`; `acceptance.md:44-65`; `research.md` §15.6; `spec.md` §E `:644-652` |
| G4-B3 empty default / resume | picker sibling | MOVED (listed `spec.md:695-697`); the core-side consequence is unstated → G5-B2 | `plan.md:152-155`; `acceptance.md:28-30` |
| G4-B4 picker scheme | picker sibling | MOVED; AC-MG-002 tombstoned | `spec.md:690-694`; `acceptance.md:33-35`; `design.md:668-671` |
| G4-A1 AC-MG-001 When / independence | fixed in core | RESOLVED | `acceptance.md:18-26` |
| G4-A2 `--model=X` | picker sibling | MOVED | `spec.md:700`; `acceptance.md:28-30` |
| G4-A3 `s` item scope | picker sibling | MOVED | `spec.md:700` |
| G4-A4 guidance AC / Where | picker sibling | MOVED (REQ-MG-020 tombstone) | `spec.md:492-496`, `:698-701` |
| G4-A5 tmux token value | teammate sibling | MOVED; the core judgment is zero tmux calls | `spec.md:714`; `acceptance.md:255` |
| G4-A6 4xx before validation | fixed in core | RESOLVED | `design.md:239`, `:254` |
| G4-A7 REQ-MG-008 absence claim | fixed in core | RESOLVED — the absence claim is replaced by judgeable obligations plus an explicit non-claim | `spec.md:419-424`; `acceptance.md:107-108` |
| G4-A8 `CLAUDE_CONFIG_DIR` tmux | teammate sibling | MOVED | `spec.md:715` |
| G4-A9 teammate model ID | teammate sibling | MOVED | `spec.md:715` |
| G4-A10 `-p` profile settings | picker sibling | MOVED | `spec.md:701` |

## Regression Check (iter4 → iter5)

- No iter4 blocker survives unchanged in core, so there is no stagnation.
- The score rose 0.80 → 0.83, so there is no STOP signal.
- G5-B1 and G5-B3 are side effects of decision 12. G5-B2 is the core-side remainder of moving G4-B3.

---

## Claim

0.6.0 resolves or cleanly routes all four iter4 blockers and all ten iter4 advisories. It still carries three blocking defects:
- G5-B1 (critical): inherited GLM keys reach the gateway `moai cc` Claude child.
- G5-B2 (major): the core-only release outcome for a persisted `/model` default is unstated and not coupled to the PICKER sibling.
- G5-B3 (major): `REQ-MG-022` overstates the in-process mitigation.

The aggregate score is 0.83 against 0.85. Verdict FAIL.

## Evidence

E1 — lint, two builds:
```
$ ~/go/bin/moai spec lint SPEC-MOAI-GATEWAY-001 ; echo exit=$?
✓ No findings — all SPEC documents are valid
exit=0
$ ~/go/bin/moai version | tail -2
 v3.2.0-rc.7   moai_cp/20260910_130400-275-ged71054d3-dirty   built 2026-09-10T19:18:41Z
$ /tmp/moai-gw060/moai spec lint SPEC-MOAI-GATEWAY-001 ; echo exit=$?
✓ No findings — all SPEC documents are valid
exit=0
$ /tmp/moai-gw060/moai version
 v3.1.3   none   built unknown
```

E2 — baseline (this run):
```
HEAD ed71054d3 ; branch WT-unified-gateway
wc -c: acceptance 41008, design 63762, plan 43728, progress 4104, research 71162, spec 62636, total 286400
```

E3 — counts and traceability:
```
live REQ ids (^\*\*REQ-MG-[0-9]{3}\*\* \(, sort -u)          : 24
REQ ids cited by live AC headers (^\*\*AC-MG-[0-9]{3}\*\* \() : 24
live AC headers                                              : 24
diff live_req.txt ac_req.txt → (no output), diff_exit=0
```

E4 — D7/D8 verbs and markers:
```
SHOULD: SPEC-MOAI-CG-RETIRE-001 not found
SPEC-MOAI-GATEWAY-001 status=status: draft
SHOULD: SPEC-MOAI-GATEWAY-PICKER-001 not found
SHOULD: SPEC-MOAI-GATEWAY-TEAMMATE-001 not found
SHOULD: SPEC-MOAI-GPT-AUTH-001 not found
SHOULD: SPEC-MOAI-PROXY-001 not found
syscall=6 gobuild=0
NEEDS CLARIFICATION: spec.md:0 plan.md:0 acceptance.md:0 design.md:0 research.md:0 progress.md:0
```

E5 — code premises (HEAD `ed71054d3`):
```
internal/hook/session_start.go:630:	// When outside tmux, fall back to "auto" (in-process display).
internal/hook/session_start.go:989:// - Outside tmux → removes override (project default "auto" applies)
internal/hook/session_start.go:995:func ensureTeammateMode(projectDir string) string {
internal/hook/session_start.go:1021:	desired := "auto"
internal/hook/session_start.go:1023:		desired = "tmux"
internal/cli/launcher.go:223-225: applyCCMode → clearTmuxSessionEnv() (warning on failure)
internal/cli/launcher.go:837:		launchEnv = buildEnvForLaunch(effectiveEffort, os.Environ())
internal/cli/launcher.go:1180-1200: buildEnvForLaunch copies base, replacing only CLAUDE_CODE_EFFORT_LEVEL
internal/cli/glm.go:507-514: buildTmuxInjectVars → ANTHROPIC_AUTH_TOKEN, ANTHROPIC_BASE_URL, ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL
internal/cli/launcher.go:279: injectTmuxSessionEnv(glmConfig, apiKey)   (only non-test caller in internal/cli)
```

## Baseline-attribution

- Every artifact read, grep and code read was taken in this run against worktree `.claude/worktrees/moai-proxy-unified` at HEAD `ed71054d3`, with the untracked SPEC directory at the sizes in E2 (matching the orchestrator's 286400 B).
- Lint attribution: the installed build carries commit `ed71054d3` (dirty, which includes the untracked SPEC directory). `/tmp/moai-gw060/moai` carries no commit stamp (`v3.1.3 none built unknown`), so its tree-build provenance is not verifiable from the binary; its result agrees.
- The Claude Code agent-teams docs content is attributed to the orchestrator's 2026-09-11 fetch and to `research.md` §16.2. I did not re-fetch.

## Gaps

- **Claude-only audit.** Cross-model audit (`audit_multi`, codex, GLM) was not invoked; `audit_model` is not set under `.moai/config/sections/` (grep returned nothing).
- **Not run live.** The G5-B1 and G5-B3 scenarios are inferences. Tmux session-env inheritance into new panes, and Claude Code honouring inherited `ANTHROPIC_DEFAULT_*_MODEL`, are documented semantics not executed here.
- **HISTORY non-alteration unverifiable.** The directory is untracked and no prior hash was recorded.
- **Probe sources not re-opened.** The probe READMEs and `mock.py` were not read; `research.md` §15.6's quotation of them was taken as given.
- **Partial reading.** Only the ranges cited above were read. Not read in full:
  - `design.md` §5 and §7–§10.
  - `acceptance.md` AC-MG-004, AC-MG-007…017, AC-MG-019…021 and AC-MG-023…025 (headers only).

## Residual-risk

- If G5-B1 is fixed by scrubbing, a GLM key arriving under an inherited name outside the cleanup set would still pass through. The scrub should key off the single cleanup-set definition, not an ad hoc list.
- Even with G5-B3 reworded, the split-pane bypass remains possible until the TEAMMATE sibling lands or a per-session flag is adopted.
- The M1 gate protects the recognizer only for the captured client version; later Claude Code versions reopen it (acknowledged at `design.md:337`).

---

## Recommendation (final authorized iteration → human decision)

This was the fifth and last authorized iteration, and it ended in FAIL. The operator chooses one of:

1. **Explicit override: one more bounded fix plus a delta-only re-audit of G5-B1, G5-B2 and G5-B3.** Recommended, for these reasons:
   - The three fixes are small and local: one REQ clause, one design step and one AC variant for B1; one release-coupling or residual paragraph for B2; one sentence or one flag decision for B3.
   - None needs a new REQ or AC number; each fits inside `REQ-MG-021`, `REQ-MG-019`, `REQ-MG-022` and `AC-MG-018` (a).
   - The operator must answer two questions first:
     - G5-B2: couple the core release to `SPEC-MOAI-GATEWAY-PICKER-001`, or accept the persisted-default residual in writing.
     - G5-B3: reword only, or adopt a per-session `--teammate-mode in-process`.
   - G5-B1 fix (i) needs no operator input; option (ii) would amend decision 12.
2. **PASS-with-debt.** Not recommended for G5-B1: it is a paid-routing correctness defect, and debt there ships an unguarded path. If chosen anyway, bind G5-B1 as a hard M7 entry gate: the child-env scrub requirement and its AC variant are written before any launch code. Record G5-B2 and G5-B3 as release-blocking notes.
3. **Further scope reduction.** Not recommended. The three blockers are side effects of the last split. Another split would move more core safety into proposed siblings without removing the inherited-env path, which exists in core regardless.

The optional findings G5-A1…A4 can be folded into whichever fix pass happens. None of them alone justifies a FAIL.
