# SPEC Review Report: SPEC-MOAI-GATEWAY-001
Iteration: 4 (operator one-time exception to the Tier L 3-iteration cap; delta audit of 0.4.0 + 0.5.0)
Verdict: FAIL
Overall Score: 0.80 (Tier L PASS threshold 0.85)
Signal: STOP — aggregate score regressed from iter3 (0.84) to iter4 (0.80). Do not iterate unconditionally; see Recommendation.

Reasoning context ignored per M1 Context Isolation. `iter4-dispatch-brief.md` was not read. The orchestrator's scope statement (which versions are in the delta, where the probe evidence lives) was used only to locate artifacts; every claim below was re-measured on this tree.

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, branch `WT-unified-gateway`, HEAD `d060e0d13`. SPEC directory untracked, version `0.5.0`.

---

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — 26 definition lines = 25 live REQs plus the retired `REQ-MG-007` tombstone (`spec.md:327-329`), no duplicates. The tree-built `moai spec lint` reports no findings (Evidence E1, E2).
- [PASS] MP-2 EARS/GEARS compliance (judged on the requirement layer, `spec.md` §D) — every live REQ carries a GEARS pattern label and form. The changed REQs: `REQ-MG-008` When+While (`spec.md:339`), `REQ-MG-019` Ubiquitous (`:408`), `REQ-MG-020` Ubiquitous+Where (`:417`), `REQ-MG-021` Ubiquitous+While (`:434`), `REQ-MG-022` Unwanted (`:505`), `REQ-MG-023` two When clauses (`:512`, `:517`). ACs were not graded here. Advisory G4-A4 concerns `REQ-MG-020`'s Where condition, which is not decidable at runtime; that is not an MP-2 failure.
- [PASS] MP-3 YAML frontmatter — `spec.md:1-15` has id, title, quoted version `"0.5.0"`, status `draft`, created, updated, author, priority `P1`, phase, module, lifecycle `spec-anchored`, tags, and tier `L`. No rejected aliases.
- [N/A] MP-4 language neutrality — this SPEC targets moai-adk-go's own Go internals (`internal/gateway`, `internal/cli`, …), not template-bound multi-language content.
- [PASS] MP-5 D7 cross-SPEC — the verb emitted no BLOCKING. Three SHOULD notices, all explained: `SPEC-MOAI-GPT-AUTH-001` and `SPEC-MOAI-CG-RETIRE-001` are marked "(제안)" (proposed) throughout; `SPEC-MOAI-PROXY-001` is the preserved iter1 report path (E3).
- [PASS, judgment] MP-6 D8 — read literally, the verb prints BLOCKING (6 `syscall` hits in `spec.md`, 0 `//go:build`). This follows the iter1–iter3 judgment (`plan-audit-iter3.md:64-67`). Every `spec.md` mention concerns preserving the existing `syscall.Exec` in `launch_exec_posix.go`. New POSIX-only watch code carries an explicit build-tag split at `design.md:195`. The delta adds no new syscall use (E3).
- [PASS] MP-7 clarification gate — `NEEDS CLARIFICATION` count is 0 in all six files (E2). See G4-B4: an open design decision exists without a marker, which sidesteps this gate in spirit.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.70 | between 0.50 and 0.75 | New interpretation forks inside the delta. The initial model is undefined for an empty configured default and on resume (G4-B3, `spec.md:408-415`). The recognizer is undefined for an absent `stream` (G4-B2, `spec.md:525`, `design.md:275`). Ownership of the tmux keys is undefined across two gateway launches (G4-B1, `spec.md:463-475`). The picker scheme is open and unrouted (G4-B4, `design.md:688`) |
| Completeness | 0.85 | between 0.75 and 1.0 | All sections and Out-of-Scope H3s are present (`spec.md:596-631`). Gaps: the probes' traffic-suppression flags are missing from `spec.md:576` and `research.md:854`; residual risk `design.md:581` omits persistence of the signal |
| Testability | 0.80 | between 0.75 and 1.0 | G3-B2 closed with a real control (`acceptance.md:263`). Mutants survive `AC-MG-003` (G4-B2) and `AC-MG-001` (G4-A2, G4-A10). The tmux surface is tested for a single launch only (`acceptance.md:245-257`). The new `REQ-MG-020` guidance obligation has no AC (G4-A4) |
| Traceability | 0.90 | between 0.75 and 1.0 | 25 AC definitions; AC headers cite exactly the 25 live REQs (E2). Deduction: the guidance clause in `REQ-MG-020` and the resume behaviour implied by `REQ-MG-019` have no judging AC |

Harmonic mean ≈ 0.805, reported as 0.80.

---

## Answers to the six orchestrator questions

1. **G3-B1 / G3-B2 resolved? Yes, both.** All cited code lines were re-opened; there is no drift (E4). The 15-key tmux union was re-derived from code: `buildTmuxInjectVars` writes 11 keys (`internal/cli/glm.go:507-540`), `buildTmuxClearVars` removes 14 (`:595-622`), and their union is exactly the 15 keys in `design.md` §6.6. The claim that `ensureTmuxGLMEnv` is "not naturally inert" is correct. Its three guards are `glm_tmux.go:80`, `:101`, `:117`. `buildGLMTmuxEnvVars` does not check whether the token is a GLM key (`:39-58`). `ensureTeammateMode` (`session_start.go:631`) runs before `ensureTmuxGLMEnv` (`:638`). The `session_end.go:728` early return that makes the old (c) vacuous is exactly as cited. The cited files are also unchanged between HEAD and local `origin/develop` (E5). **The G3-B1 fix, however, introduces G4-B1.**
2. **Unmeasured behaviour presented as supported.** Mostly disciplined: T04, T15, real providers, picker Enter, `--model` precedence, `/v1/models` discovery and teammate tmux inheritance are all fenced as unmeasured (`spec.md:576`, `acceptance.md:24-31`). Overstatements found:
   - The probes ran with nonessential traffic, telemetry, error reporting and auto-update disabled, and that condition was dropped from the SPEC. Recognizer narrowness rests on the resulting request inventory (G4-B2).
   - The mock logged `stream` as `bool(req.get("stream"))`, so "`stream: false`" was never observed as a JSON value (G4-B2).
   - `REQ-MG-020` generalizes the `s` result from one custom picker entry to picker entries in general (G4-A3).
   - `REQ-MG-008`'s "no silent bypass path" sentence (G4-A7).
3. **A-VAL-2.1.267 narrowness and mutants.** The recognizer is not demonstrably narrow, and mutants survive in both directions (G4-B2). For `AC-MG-001`'s persisted-default judgment: an implementation that passes `--model` only for the separated form survives (G4-A2), and so does one that scrubs `model` from a profile's `settings.json` (G4-A10). The empty-default case is not defined at all (G4-B3).
4. **Folding and the 25/25 count.** The count is mechanically honest: 25 AC definitions, 25 distinct REQs in AC headers (E2). The judgment surface is not constant, and `plan.md` §I (`:393`) says so. `AC-MG-003` remains auditable: lettered sub-judgments with a control. `AC-MG-001` became the compressed kind §I warns about: two judgments with inconsistent When clauses and no independence sentence (G4-A1).
5. **§6.7 picker "Default" risk.** It does not contradict `REQ-MG-019`: an identical overlay across launchers keeps "differ only in initial model". It is still a blocker, because it is an unrouted decision that M1 depends on (`plan.md:98`) and that `design.md:677` itself says conflicts with `REQ-MG-022` and the §6.1/§6.6 slot-key cleanup (G4-B4). Routing it to a gate is enough; choosing a scheme is not required at plan phase.
6. **Leftover contradictions between files.** None live. `plan.md:335-336` keeps pre-0.5.0 text saying `s` was unmeasured, but explicitly labels it as a historical attribution record. `spec.md:92` is inside HISTORY. `plan.md:94` ("switching through the Go gateway not yet observed") agrees with `spec.md:266` and `research.md` §15.

---

## Defects Found (structured defect-list)

D1. **G4-B1** — `spec.md:463-475`, `spec.md:483-493`, `design.md:369`, `design.md:410-413`, `design.md:581`, `acceptance.md:245-257`
The tmux session env has no ownership model. The 0.4.0 contract writes fixed keys (`ANTHROPIC_BASE_URL`, `MOAI_LAUNCH_PROVIDER`, optional carrier) into the one environment each tmux session has, and suppresses the only hook that clears them.
- **Two gateway launches in one tmux session are ordinary, not hypothetical.** `--spawn`, whose behaviour `REQ-MG-002` preserves, runs `tmux new-window` in the caller's current session (`internal/cli/spawn.go:78-88`, which says "in the caller's current session"). `InjectEnv` writes session scope with no `-g` (`internal/tmux/session.go:206-208`).
- **Failure scenario 1.** Gateway A in window 1; `moai gpt --spawn` starts gateway B in window 2, whose launch clears and re-injects the keys. Agent Teams panes spawned afterwards by A's lead inherit B's loopback address and `MOAI_LAUNCH_PROVIDER=gpt`. Their requests are served by B, whose credential state may differ, and their hooks apply B's provider. `REQ-MG-008`'s rule "serve while panes that received *my* address are alive" cannot be enforced: A cannot know which panes received its address, and B's exit strands A's teammates.
- **Failure scenario 2.** After the last gateway exits, `MOAI_LAUNCH_PROVIDER` remains in the tmux env; `design.md:581` records only the loopback URL remaining. A later non-gateway Claude process in a new pane inherits the signal. That includes a direct `claude`, or `moai cg`, which stays live until the sibling SPEC removes it. Its SessionEnd `cleanupGLMSettingsLocal` and hook `clearTmuxSessionEnv` are then suppressed. For `moai cg` this leaves the GLM token and Z.AI URL it injected in the tmux env; today's SessionEnd (`session_end.go:93`) removes them. This contradicts `design.md:369` ("the existence of `MOAI_LAUNCH_PROVIDER` carries the fact that this is a gateway launch") and `design.md:410-413` (the "absent" row covers direct `claude` runs; pre-gateway sessions' cleanup is not cut off).
- **Evidence type.** Inference from code plus tmux session-environment semantics; not run live. Codex raised the same concurrency defect independently (fail-open backend, see Evidence E7); I verified its premise in `spawn.go`.
- **Severity: major. Class: blocking** (internal consistency of the `REQ-MG-021` While discriminant; `REQ-MG-008` lifetime contract cannot be enforced). **Tag: (O) operator decision** (alternative R).
- **Required fix.** Decide one of the following, then write it into `REQ-MG-021`/`REQ-MG-008`:
  - (i) at most one gateway per tmux session, enforced with an owner lock;
  - (ii) owner-tagged keys (gateway id + PID/fingerprint + port) with compare-and-clear at gateway exit, which also removes the dead-loopback residual;
  - (iii) declare multi-gateway tmux sessions out of scope, with an Out-of-Scope H3 and a launch-time refusal.

  In every case, state in `design.md` §5.2/§5.4 what a tmux-inherited signal means for a non-gateway process. Add to `AC-MG-018` (a) a two-launch sequence (A launch → B launch → A exit → new pane) and a check for the signal after exit.

D2. **G4-B2** — `spec.md:517-529`, `design.md:270-281`, `plan.md:104-108`, `acceptance.md:45-60`, `research.md:854`, `spec.md:576`
The evidence cannot pin the A-VAL-2.1.267 recognizer, and the ACs admit mutants in both directions.
- **(a) `stream` was never observed as a JSON `false`.** The probe mock logged `stream = bool(req.get("stream"))` (`.moai/state/gwprobe/mock.py:89`) and tools as `len(req.get("tools") or [])` (`:93`). `requests.jsonl` stores these derived fields, not request bodies. An absent `stream` and `false` produce the same record.
- **(b) Mutant 1.** A recognizer requiring JSON `false` passes an `AC-MG-003` fixture copied from these records (`plan.md:104`). If 2.1.267 actually omits `stream`, every real validation request is forwarded upstream: a paid call per `/model` switch, provider-shaped errors, and decision 9 silently not delivered.
- **Mutant 2.** A recognizer treating absent as `false` also passes. The REQ does not define absence, so neither mutant is detectably wrong.
- **Mutant 3.** A recognizer that skips the `role: user` check passes, because (c) varies only `max_tokens`, `stream` and message count.
- **(c) The narrowness rests on a conditioned inventory.** Narrowness is argued from the observed inventory of turn, title and validation requests (`design.md:279-281`; `plan.md` decision 9: narrow enough that real turns are not answered locally). Both probes ran with `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `DISABLE_TELEMETRY=1`, `DISABLE_ERROR_REPORTING=1` and `DISABLE_AUTOUPDATER=1` (gwprobe `README.md`, "What was run"). Neither `research.md:854` nor `spec.md` §E records this. The SPEC's own cleanup deletes `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` (`spec.md:450`), so production sessions run under a condition that was not measured. Other one-token non-streaming client requests in that condition were not inventoried; if one exists, it receives a fabricated local success.
- **Severity: major. Class: blocking** (`REQ-MG-023` correctness; a measured claim relayed without its conditions). **Tag: (D) debt-eligible only if bound to M1 as an entry gate**; otherwise fix at plan phase.
- **Required fix.**
  1. Record the four flags as probe conditions in `research.md` §15.1 and `spec.md` §E.
  2. Require M1 to capture raw request bodies and query strings through the gateway, without those flags, before freezing the fixture. Keep the fixture's source as a raw capture, not the derived log.
  3. Define absent-`stream` handling in `REQ-MG-023` and `design.md` §4.1 after the capture.
  4. Add `AC-MG-003` (c) variants: absent `stream`, single non-`user` message, and tools present, each with a stated expected outcome.

D3. **G4-B3** — `spec.md:408-415`, `plan.md:143-146`, `acceptance.md:17-28`
`REQ-MG-019` requires the initial model always as explicit `--model`, and defines it for `moai cc` as "the configured Claude default model". Two cases are left undefined.
- **(a) Empty configured model.** The launcher documents an empty resolved model as "the normal, intentional state" for the default profile (`internal/cli/launcher.go:743-751`; the flag is appended only when non-empty, `:768-770`). `plan.md:143-146` acknowledges the behaviour change but supplies no value.
- **Failure scenario.** Implementers diverge three ways, each colliding with other text:
  - omit `--model` when empty: violates "always", and the persisted `/model <id>` default then silently sets the start model, which decision 10 exists to prevent;
  - hardcode a catalog Claude default: an unstated product choice, and a constant the SPEC never names;
  - fail the launch: breaks default-profile `moai cc`.
- **(b) Resume.** `-c/--continue` runs `claude --continue [--model …]` through a spawn-and-wait `exec.Command(...).Run()` before any fresh exec (`launcher.go:792-812`). Under decision 10 a resumed session is re-pinned to the launcher's initial model, overriding the session's own `/model` or `s` choice; this is unmeasured (`research.md:901` only). The path is also absent from `design.md` §3.2's exec-only sequence (the launcher disappears at step 8), which governs gateway child ownership. Codex raised the resume path independently (E7); I verified it at `launcher.go:792-812`.
- **Severity: major. Class: blocking** (an obligation undefined in the documented normal state). **Tag: (O).**
- **Required fix.** State the initial-model value when the configured default is empty. Add a fresh-vs-resume rule (fresh / `--continue` / `--resume` / continue-fallback), listing resume precedence as unmeasured in `spec.md` §E. Add the empty-default and resume cases to `AC-MG-001`.

D4. **G4-B4** — `design.md:664-689`, `plan.md:98`, `acceptance.md:33-37`
The picker configuration scheme is an open decision that the SPEC itself says conflicts with other parts, and it is routed nowhere.
- `design.md:677` states that the current design collides in three places: `REQ-MG-022` (the "Default" row retargets to a pinned non-Claude model), the §6.1/§6.6 slot-key cleanup (the same key names), and `AC-MG-018`'s projection.
- `design.md:688` leaves the scheme as "a decision candidate". `plan.md` has no §H decision, M1 gate or clarification marker for it (grep for `6.7`, `picker 구성` and `결정 후보` over plan/acceptance/spec returns no routing).
- M1 nevertheless instructs building the three-provider overlay (`plan.md:98`). `AC-MG-002` compares item sets only (`acceptance.md:33-37`), so an overlay whose "Default" row routes `moai cc` users to GLM passes it.
- **Failure scenario.** The M1 implementer picks the slot-pin scheme, which is the only multi-entry scheme observed. A `moai cc` user choosing "Default (recommended)" is routed to Z.AI, and `AC-MG-002` stays green.
- **Severity: major. Class: blocking** (the same silent-assumption class as G3-B1). **Tag: (O).**
- **Required fix.**
  1. Add the picker scheme to `plan.md` §H as an open operator decision with an explicit gate before M1's overlay step.
  2. State the invariant the chosen scheme must satisfy (for example, the "Default" row must not resolve to a provider other than the launcher's initial provider, or the row must be absent).
  3. Strengthen `AC-MG-002` from item-set equality to a label → request `model` → provider mapping check, including the "Default" row.

D5. **G4-A1** — `acceptance.md:17-31` — `AC-MG-001` now carries judgments whose When clauses differ: "first `/v1/messages` request" (`:18`) versus "first turn request" (`:25-26`). The probes show the title-generation request arriving 2 ms before the turn request (gwprobe `requests.jsonl` ts `…274.328` vs `…274.33`). There is no independence sentence as in `AC-MG-018`. Severity minor. Class: optional. Fix: unify the When clauses to "first turn request" and add "each judgment is PASS/Gap independently".

D6. **G4-A2** — `internal/cli/launcher.go:704-715`, `acceptance.md:28` — Only the separated `--model X` and `-m X` forms are parsed. `--model=X` falls to `passThrough`. Under decision 10 that yields two `--model` arguments, and `AC-MG-001`'s "exactly once" check survives an implementation that ignores the `=` form. Severity minor. Class: optional. Fix: name the `--model=X` and `-m=X` forms in `AC-MG-001`.

D7. **G4-A3** — `spec.md:420-428` — The session-only behaviour of `s` is stated for "picker items". It was measured for a single `ANTHROPIC_CUSTOM_MODEL_OPTION` entry (gwprobe2 `README.md` observation 2); built-in and slot-pinned entries were not measured. Severity minor. Class: optional. Fix: scope the sentence to the measured entry kind and add the others to §E.

D8. **G4-A4** — `spec.md:417-432` — The new guidance obligation ("the system's guidance shall …") has no AC that judges it. The Where condition ("where the target version exhibits the behaviour measured in 2.1.267") is not mechanically decidable. Severity minor. Class: optional. Fix: bind the guidance to a documented artifact that an AC checks, or make the Where a static version gate.

D9. **G4-A5** — `acceptance.md:251` — The tmux surface accepts a remaining `ANTHROPIC_AUTH_TOKEN` as long as its value differs from the GLM marker. `REQ-MG-021` (`spec.md:466-468`) allows only deletion or the session access token. An implementation writing any other value passes. Severity minor. Class: optional.

D10. **G4-A6** — `design.md:241-243` — The flow diagram rejects registry misses and absent credentials with generic "4xx" before the validation-request branch. §4.1 requires typed 404 and 401 responses for validation-shaped requests, so an implementation following the diagram's order fails `AC-MG-003` (a)/(b). Severity minor. Class: optional. Fix: move the validation check first, or annotate the early 4xx steps.

D11. **G4-A7** — `spec.md:339-347` — "There is no path for this failure to become a silent bypass" is an absence claim derived only from the tmux session env contents. Pane shell env, tmux global env, and client behaviour on connection refusal are unmeasured. Severity minor. Class: optional. Fix: narrow it to "no path through the tmux session env".

D12. **G4-A8** — `design.md` §6.6 table — The union derivation is correct, but `CLAUDE_CONFIG_DIR` is a profile-isolation key, not a GLM key. Its only tmux writer is `applyCGMode`'s profile write (`launcher.go:355-361`). Clearing it on every gateway launch is a new behaviour for `moai glm`/`moai gpt`, and its effect on `-p` teammates is unexamined. Severity minor. Class: optional.

D13. **G4-A9** — `plan.md:231-234` — The M7 teammate measurement states its negative outcome only for inheritance, not for "which model ID teammates carry". Per §6.6 teammates no longer receive slot keys, so under a `moai glm` lead their alias requests may resolve to Claude IDs. Severity minor. Class: optional. Fix: define which observed model IDs count as negative for `REQ-MG-018`.

D14. **G4-A10** — `spec.md:417-419` versus `:431` — The write exclusion names `~/.claude/settings.json`, while 0.5.0 text speaks of "the config directory's `settings.json"`. Under `-p` profiles that directory is elsewhere. An implementation scrubbing `model` from the profile's `settings.json` satisfies `AC-MG-001` without `--model` precedence. Severity minor. Class: optional.

---

## Disposition — iter3 defects

| Defect | Status | Evidence |
|---|---|---|
| G3-B1 tmux session env contract | RESOLVED; introduced G4-B1 | `spec.md:463-475` (clear, inject, write ban), `:483-493` (While list now includes hook `clearTmuxSessionEnv` and `ensureTmuxGLMEnv`), `:511-513` (`REQ-MG-022` tmux bypass), `design.md` §5.4 table `:406-410`, §6.6, `acceptance.md:245-257`, `plan.md` decision 8 and M7. 15-key union re-derived from `glm.go:507-540` ∪ `:595-622`. All citations match (E4) |
| G3-B2 vacuous `cleanupGLMSettingsLocal` check | RESOLVED | `acceptance.md:262-270`: fixture keeps `ANTHROPIC_BASE_URL` and `ANTHROPIC_DEFAULT_OPUS_MODEL`, with a no-signal control that must delete both. Early return confirmed at `session_end.go:728` |
| G3-A1 `ensureTeammateMode` legacy delete | RESOLVED | `design.md` §6.3, `research.md:351`, `:401`, `:407-408`, `acceptance.md` exclusion paragraph. Code `session_start.go:1033-1047` |
| G3-A2 Windows verdict owner and record | RESOLVED | `acceptance.md` `AC-MG-006` timing paragraph, `acceptance.md` §D, `design.md` §3.5, `plan.md` M4 and decision 4 |
| G3-A3 progress/research revision lists | RESOLVED | `progress.md` §E.1 lists 0.3.2, 0.4.0, 0.5.0; `research.md:12` |
| G3-A4 `MOAI_HOME` IsAbs condition | RESOLVED | `research.md:435`, `:651` |
| G3-A5 HISTORY section reference | RESOLVED | `spec.md` HISTORY 0.4.0 corrects §2.1 → §2.2 |
| G3-A6 verification venue in `REQ-MG-009` | RESOLVED | `REQ-MG-009` (`spec.md:349-352`) has no venue wording; moved to `spec.md` §E |
| G3-A7 "either value" | RESOLVED | `design.md` §6.3 table, `research.md:426`; code `session_start.go:858-866` (`after != before`) |
| G3-A8 empty backup key | RESOLVED | `spec.md` `REQ-MG-021` layer (a), `design.md` §6.1, `acceptance.md` (a) empty-backup variant; code `launcher.go:406-411` |
| G3-A9 carrier-key condition | RESOLVED | `design.md` §6.5 first bullet |
| G3-A10 sub-judgment independence | RESOLVED | `acceptance.md` `AC-MG-018` closing paragraph |

## Regression Check (iter3 → iter4)

- G3-B1: RESOLVED, with a new defect on the same surface (G4-B1). This is not stagnation: the contract now exists, and its multi-launch and post-exit semantics are what is missing.
- G3-B2: RESOLVED.
- G3-A1 … G3-A10: all RESOLVED.
- Score: 0.84 → 0.80. The regression comes from 0.5.0 scope growth (decisions 9 and 10, the §6.7 risk), not from reopened iter3 defects.

---

## Claim

SPEC-MOAI-GATEWAY-001 0.5.0 closes both iter3 blockers and all ten iter3 advisories. It does not pass iter4: four new blocking defects sit in the delta (G4-B1 … G4-B4), and the aggregate score is 0.80 against the Tier L threshold of 0.85.

## Evidence

E1 — Lint using a binary built from this tree (tool provenance per the verification-claim-integrity rule, §2.2):
```
$ go build -o <scratchpad>/moai-tree ./cmd/moai ; echo "build exit=$?"
build exit=0
$ <scratchpad>/moai-tree spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
lint exit=0
```
The installed binary also reported no findings, but it is `v3.2.0-rc.5 list-974-g84fa4ece4`, and `84fa4ece4` is an ancestor of HEAD `d060e0d13`. The tree build is the attributed result.

E2 — Counts:
```
REQdef 26
AC defs: 25
distinct REQ in AC headers: 25
NEEDS CLARIFICATION: spec.md:0 plan.md:0 acceptance.md:0 design.md:0 research.md:0 progress.md:0
wc -c: acceptance 39151, design 61624, plan 36085, progress 3147, research 64679, spec 51951 (total 256637)
```

E3 — D7 / D8 verbs:
```
SHOULD: SPEC-MOAI-CG-RETIRE-001 not found
SPEC-MOAI-GATEWAY-001 status=draft
SHOULD: SPEC-MOAI-GPT-AUTH-001 not found
SHOULD: SPEC-MOAI-PROXY-001 not found
-- D8
syscall hits: 6
D8 BLOCKING: no //go:build or EXCL in spec.md
.moai/specs/SPEC-MOAI-GATEWAY-001/design.md:195:`launch_exec_posix.go`(`//go:build !windows`)와 `launch_exec_windows.go`(`//go:build windows`)처럼
```

E4 — Code citation checks (grep -n on HEAD `d060e0d13`):
```
internal/hook/session_start.go:623 ensureGLMCredentials(input…  :631 ensureTeammateMode(input…  :638 ensureTmuxGLMEnv(input…
internal/hook/session_start.go:995 func ensureTeammateMode   :1026 if current == desired
internal/hook/session_end.go:93 clearTmuxSessionEnv(ctx)  :628 "ANTHROPIC_AUTH_TOKEN is included"  :634 var glmEnvVarsToClean  :646 func clearTmuxSessionEnv  :653 "set-environment", "-u"  :682 func cleanupGLMSettingsLocal  :728 glmActive := env[…]
internal/hook/glm_tmux.go:80 TMUX==""  :101 teammateMode != "tmux"  :117 len(vars)==0  :132 InjectSensitiveEnv  :143 InjectEnv
internal/cli/launcher.go:224 clearTmuxSessionEnv() in applyCCMode  :272 "(moai cg path)"  :281 "may not have GLM credentials"  :359 tmux set-environment CLAUDE_CONFIG_DIR (applyCGMode)  :402 delete(m,"teammateMode")  :704 case "--model","-m"  :741 resolveMainSessionModel  :769 append "--model"
internal/cli/glm.go:571, :592 "intentionally excluded"  :557/:566 inject  :583 ClearEnv(buildTmuxClearVars())
internal/tmux/session.go:206-208 InjectEnv args {"set-environment", key, value} (no -g)  :365-367 ClearEnv {"set-environment","-u",key}
internal/cli/spawn.go:78 "defaultTmuxSpawn runs `tmux new-window` in the caller's current session"  :88 exec tmux new-window -d …
internal/cli/launcher.go:792-812 --continue: exec.Command(claudeBin, buildArgs(true)[1:]...).Run(), exit 1 → fresh launch
```

E5 — Drift of the cited code against the local develop ref (no fetch):
```
$ git rev-list --count --left-right origin/develop...HEAD
225	0
$ git diff --stat HEAD..origin/develop -- internal/cli/launcher.go internal/cli/glm.go internal/hook/glm_tmux.go internal/hook/session_end.go internal/hook/session_start.go internal/tmux/session.go
(empty)
```

E6 — Probe evidence (machine-local, untracked):
```
.moai/state/gwprobe/mock.py:89   stream = bool(req.get("stream"))
.moai/state/gwprobe/mock.py:91-93 log({... "stream": stream, "n_messages": len(msgs), ... "max_tokens": req.get("max_tokens"), "tools": len(req.get("tools") or []), ...
gwprobe/requests.jsonl validate rows: {"model": "gpt-5.6-sol", "stream": false, "n_messages": 1, "last_user": "Hi", "max_tokens": 1, "tools": 0, ...}
gwprobe/README.md "What was run": CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1, DISABLE_AUTOUPDATER=1, DISABLE_TELEMETRY=1, DISABLE_ERROR_REPORTING=1
gwprobe2/requests.jsonl GPT turn messages roles: user, system, assistant, user, system, assistant, user, system
gwprobe2/README.md observation 2: "The picker path sends no validation request, for a ANTHROPIC_CUSTOM_MODEL_OPTION entry"
grep NONESSENTIAL over SPEC files → research.md:454 spec.md:450 design.md:482 design.md:626 acceptance.md:219 (all cleanup-key contexts; none in probe-condition sections)
```

E7 — Cross-model second opinion (`mcp__moai__audit_multi`, `project_root` set to this worktree):
- `overall_verdict: fail`, `participant_count: 1`.
- Codex's verdict was parsed as `inconclusive` (fail-open), but its summary text reported FAIL with four findings: shared tmux env ownership, the unrouted picker scheme, recognizer over-match, and `--model` versus `--continue`. Each premise I adopted was re-verified in code before use (E4: `spawn.go:78-88`, `launcher.go:792-812`).
- GLM returned `inconclusive` ("no reviewable change": the SPEC is untracked).
- The tool reported `build_commit 84fa4ece4` behind HEAD.

## Baseline-attribution

- All file reads and greps: this tree at HEAD `d060e0d13`, in this run.
- Lint: a binary built from this tree in this run (E1).
- Probe facts: read from the probe files in this run. The probes themselves were run by another actor earlier and not re-run here.
- No figure is carried over from iter3.

## Gaps

- No runtime check of tmux behaviour. G4-B1's cross-launch overwrite and signal inheritance rest on code (`spawn.go`, `session.go`) plus tmux session-environment semantics, not on a live two-launch reproduction.
- I did not establish whether Claude Code 2.1.267 omits or sends `stream: false` in the validation request. The evidence cannot tell (G4-B2), and I did not re-probe.
- I did not inventory Claude Code requests with nonessential traffic enabled.
- I did not verify how kanban/factory lanes are placed in tmux sessions (only `--spawn` was read).
- Sections the delta did not touch (for example §2 types and §7 kanban) were not re-audited, apart from the cross-layer sweep of changed REQs.

## Residual-risk

- Even after G4-B1 through G4-B4 are fixed, the whole model-switching premise is proven only against a Python mock. M1's gateway reproduction may still overturn decisions 9 and 10.
- Tier L budget is 25/25 and `AC-MG-001`, `-003` and `-018` are dense. Each further fold raises the chance that a vacuous sub-judgment hides, as G3-B2 did.
- This is the fourth iteration, and the score regressed. Another in-place revision is likely to add surface faster than it closes it.

---

## Recommendation

FAIL with a STOP signal: the score regressed from 0.84 to 0.80 on an iteration that already exceeds the Tier L cap. Do not iterate unconditionally. For the orchestrator's user question:

1. **Scope reduction (recommended by this auditor).** Split out the picker/model-selection surface (decisions 9 and 10, `REQ-MG-019`/`020`/`023`'s validation clause, §6.7) and the tmux teammate surface (`REQ-MG-008` teammate lifetime, `REQ-MG-021` tmux block). Those two surfaces hold all four blockers, and both depend on M1/M7 measurements that do not exist yet. The core (supervisor, ingress, adapters, settings.local.json contract) had no blocker in this delta.
2. **PASS-with-debt.** Only defensible for G4-B2, and only if it is bound to M1 as an entry gate (raw capture before fixture freeze). G4-B1, G4-B3 and G4-B4 are operator decisions (O) and must not be carried as silent debt.
3. **Explicit override for a fifth iteration.** It would need, at minimum:
   - an operator decision on tmux ownership (G4-B1: one gateway per tmux session, compare-and-clear, or out of scope);
   - an operator decision on the initial model when the default is empty and on resume (G4-B3);
   - the picker scheme routed to a §H gate with its "Default"-row invariant, and `AC-MG-002` strengthened (G4-B4);
   - the probe conditions recorded, M1 raw-capture wording added, and the absent-`stream`, role and tools variants added (G4-B2);
   - optional fixes G4-A1 … A10.
