---
id: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001
title: "Self-improvement protected zone — declare the checking apparatus in a manifest and block calls carrying a self-improvement agent identity from modifying it"
version: "0.3.0"
status: draft
created: 2026-10-04
updated: 2026-10-05
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook,internal/config,internal/template"
lifecycle: spec-anchored
tier: M
tags: "protected-zone, self-improvement, pretooluse, guard, manifest, harness-learner, fail-closed, t1510"
---

# SPEC-SELF-IMPROVE-PROTECTED-ZONE-001

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-04 | manager-spec | Initial plan-phase draft (card t1510, operator decision D5, Class C). Measurements taken at tree SHA e497f693608ac7ea45a08b06304dc585e927ff49 with a binary built from that clean tree in this run; the evidence ledger is `acceptance.md` §E and the raw outputs are under `evidence/`. |
| 0.2.0 | 2026-10-04 | manager-spec | Plan-audit iteration 1 repair (FAIL 0.74, defects D1-D19). Entries gain a per-category `runtime_paths` list and a stated sweep tree (D2); `.claude/skills/moai/` and the full 21-entry baseline list are specified (D3); the denial reason has a fixed field order that survives truncation and a `next=return-blocker-report` instruction, and the human route is claimed only as far as it is observable (D4, D5, G8); root and target are resolved by the same symlink function and the pure lexical rules are split from the filesystem rules (D6, D11); basename-glob semantics defined (D10); REQ-007's 13 verb forms and REQ-013's four bypass forms are enumerated (D7, D9); identity reach reworded (D16); REQ-012 reworded (D19). The base-tree evidence was re-recorded with 67 probe cases. No requirement added; REQ count stays 16. |
| 0.3.0 | 2026-10-05 | manager-spec | Final delta repair before run (plan-audit iteration 2 FAIL 0.78, defects D20-D28; operator approved proceeding as PASS-with-debt, so this delta was not re-audited). The legacy local-instruction basename leaves the shipped manifest for the dogfood overlay (D20); the seven-category rule binds any manifest at the shipped path and the probe fixtures MS4/MS7/MS8 now carry all seven (D21); the real scope of the human route is stated in REQ-SIPZ-011, REQ-SIPZ-012, §C.7 and G8 — a baseline-matched file-tool denial keeps its legacy reason and carries no routing field (D23); the invalid-manifest reason carries the routing fields and the failing file (D26); `.moai/logs/` moves from `runtime_paths` to `paths` (D27). No requirement or criterion added; REQ count stays 16. |

## §A Problem — what the current guard actually covers

The card asks for the whole checking apparatus (safety guards, gate policy, auditor instructions and thresholds, the regression-test bundle, the apply/rollback machinery, budgets, logs) to be declared in a manifest and for the PreToolUse guard to stop self-improvement paths from modifying it. Today the only such guard is the harness-learner FROZEN-zone check in `internal/hook/pre_tool.go` (`frozenZonePrefixes`, `frozenInstructionFiles`, `checkHarnessFrozenZone`). Measured at base SHA e497f693608ac7ea45a08b06304dc585e927ff49 by driving the real handler (`moai hook pre-tool`) with 67 fixed cases (`evidence/probe-protected-zone.sh`):

### §A.1 What is blocked today

- Four controls are denied with their legacy sentinel: a `.claude/hooks/` write, a `.claude/rules/moai/` write, the instruction file `CLAUDE.md`, and a `.claude/agents/moai/` Edit. The guard reaches only six path prefixes and four instruction basenames.
- Nothing else is denied. All 21 probe rows naming a checking-apparatus path (R1-R21: safety guards, gate policy, auditor workflow and profiles, Go guard code, Go tests, CI workflow, apply/rollback code, logs, budget state, the `/moai` skill directory, and the manifest's own future paths) return `allow` — except the CI workflow, which returns `ask`, a decision a subagent cannot answer.
- `HARNESS_FROZEN_CONFIG_VIOLATION` is declared in `internal/hook/pre_tool.go` and asserted by `internal/harness/sentinel_catalog_test.go`, but no code path emits it; `.moai/config/sections/*.yaml` is therefore unguarded although the sentinel's own comment says it is.

### §A.2 The guard does not survive the path forms a real tool call carries

The matching is a raw string-prefix test on the `file_path` exactly as received. The same `.claude/hooks/` write that is denied when sent as `.claude/hooks/moai/x.sh` is **allowed** when sent as an absolute path inside the project, with a leading `./`, through a `docs/../` detour, with different letter case (`.CLAUDE/Hooks/…`), through a symlink into the directory, or when the project root itself is reached through a symlink (7 of 7 rows, all `allow`; the symlinked-root pair is not avoided by the fixture). Claude Code's Write and Edit tools take absolute file paths, so the one guard that exists is, in the common case, matching nothing. (A prior SPEC recorded the relative-versus-absolute concern as incidental and unmeasured; this run measured it.)

### §A.3 Which self-improvement paths exist, and which the PreToolUse guard can see

| Path | Writes through | Seen by PreToolUse? | Existing guard on it |
|---|---|---|---|
| `harness-learner` identity (Write/Edit/Bash) | Claude Code tools | Yes — but only when a spawn carries `agent_type` equal to that name | `frozenZonePrefixes`, `frozenInstructionFiles` (6 + 4 entries) |
| `moai-harness-learner` skill flow in the orchestrator's own session | `moai harness apply --execute` (a Bash command that runs a Go process) | The Bash command text only; the writes happen inside the process | `internal/harness/safety` L1 `frozenPrefixes` (7 entries) and the meta-harness `frozenPrefixes` (4 entries) |
| LSEL applier (`hns-lsel-applier` skill) | `.moai/hooks/lsel-apply.sh` running `git apply` | The Bash command text only | `.claude/lsel/frozen-allowlist.json` (9 regex `frozen_patterns`, 5 `execution_meta`); `internal/hook/`, `internal/harness/safety/` and the test files are in neither list |
| `builder-harness` agent | Claude Code tools into `.claude/agents/harness/`, `hns-*` skills | Yes, but it is a builder, not a self-improver, and writes the evolvable surface | none needed; not in the identity set |

No agent definition named `harness-learner` exists under `.claude/agents/` (`ls` of both directories), and the learner skill runs in the orchestrator's session where `agent_type` is empty. Whether any real flow spawns an agent under that name is unobserved (§F G1). What is measured is that the guard's identity gate is a spawn-name equality and that the main-branch-guard doctrine records `agent_type` as the spawn name verbatim for an Agent-tool spawn (`.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement).

### §A.4 Four independent frozen lists already exist

`internal/hook/pre_tool.go` (6 prefixes + 4 basenames), `internal/harness/safety/frozen_guard.go` + `frozen_rules.go` (7 prefixes, mirrored by a rule registry), `internal/harness/frozen_guard.go` (4 prefixes), and `.claude/lsel/frozen-allowlist.json` (regexes). They disagree — only the safety L1 list covers `.claude/settings.json` and the two `frozen_guard.go` sources, and the hook list covers `.claude/hooks/`, `.claude/commands/` and `.claude/output-styles/`, which the others do not. `zone-registry.md` is a fifth, but it is clause-level (CONST-* records keyed to rule text with `[ZONE:*]` tags) and carries no path field.

### §A.5 The hook only receives three tools

The PreToolUse matcher in `.claude/settings.json` and in `internal/template/templates/.claude/settings.json.tmpl` is `Write|Edit|Bash`, and existing template tests pin that string ("must not be renamed or widened"). `MultiEdit` and `NotebookEdit` never reach the guard, and PowerShell has its own separate matcher block.

## §B Requirements (GEARS)

- **REQ-SIPZ-001** (declaration) — The protected-zone manifest shall declare the checking apparatus as named categories — ASCII `[a-z0-9_]` names of at most 32 bytes, at minimum `safety_guards`, `gate_policy`, `auditor`, `regression_tests`, `apply_rollback`, `budgets` and `logs`, all seven present in the shipped manifest even when a list is empty, so that a file at the shipped path lacking any of the seven is invalid whatever else it declares (an extra category is valid, and the add-only overlay is not bound by the seven) — each category holding a `paths` list and an optional `runtime_paths` list of repository-relative entries, where an entry is written in exactly one of four forms: an exact path, a directory prefix ending in `/`, a raw string prefix ending in `*`, or a basename glob `**/<pattern>` whose `<pattern>` contains no `/` and is matched against the final path segment by the Go `path.Match` rules (`*`, `?`, and `[...]` classes, `[` always opening a class) with both sides case-folded, and where a `runtime_paths` entry is protected like any other but may be absent from the tree its file ships in.
- **REQ-SIPZ-002** (layering and self-inclusion) — The effective protected zone shall be the union of the shipped manifest at `.moai/config/sections/protected-zone.yaml`, the optional project overlay at `.moai/project/protected-zone.yaml`, and the compiled baseline floor, where an overlay entry can only add to the zone, and both manifest files shall list themselves as protected entries.
- **REQ-SIPZ-003** (template neutrality) — The shipped manifest shall name only language-neutral, moai-managed paths, shall ship the `regression_tests` category empty, and shall carry no SPEC identifier, requirement token, audit citation, internal date, commit hash, or the legacy local-instruction file name that the template-neutrality audit rejects in any shipped YAML — that entry lives in the dogfood overlay, and the compiled baseline floor keeps covering it in every other project.
- **REQ-SIPZ-004** (precedence and non-drift) — A path listed in the effective zone shall be protected regardless of any `[ZONE:Frozen]` or `[ZONE:Evolvable]` tag on the clause or file that governs it, and every entry of the compiled baseline and of the hard-coded frozen lists in `internal/harness/safety` and `internal/harness` shall be covered by the dogfood effective zone.
- **REQ-SIPZ-005** (file-tool block) — When a Write or Edit PreToolUse event carries an `agent_type` equal to a member of the compiled self-improvement identity set, the guard shall deny the call if the normalized target path is covered by the effective zone.
- **REQ-SIPZ-006** (normalization) — The guard shall compare a target path to the zone only after two groups of steps. The lexical steps touch no filesystem and are the same on every platform: rewrite backslashes to slashes; treat a leading `/`, a drive letter followed by `:` and `/`, and a leading `//` as absolute; express an absolute path relative to the project root by stripping the root prefix case-insensitively, and treat an absolute path that is not under the root as outside the zone; resolve relative input as the existing file-access check resolves it; clean `.` and `..` segments; normalize Unicode to NFC; fold ASCII case. The filesystem steps resolve the project root and the deepest existing ancestor of the target through symlinks with the same function — the root always, not only when the target resolved, which is the existing check's deliberate asymmetry and which this guard does not copy — and match the zone against both the lexical and the resolved relative forms, falling back to the lexical form when resolution fails; a path that escapes the root is outside the zone and is left to the existing outside-project deny.
- **REQ-SIPZ-007** (shell mutation) — When a Bash PreToolUse event carries an `agent_type` equal to a member of the compiled self-improvement identity set and a command segment pairs one of the thirteen mutating forms (`rm`, `unlink`, `mv`, `cp`, `tee`, `truncate`, `sed -i`, a `>` redirection, a `>>` redirection, `git rm`, `git checkout`, `git restore`, `git apply`) with an argument or redirection target that is covered by the zone after the same normalization, the guard shall deny the call.
- **REQ-SIPZ-008** (identity scope and cost) — The guard shall apply only to the compiled self-improvement identity set, initially `harness-learner`, and shall read no manifest file for any other identity or for any tool other than Write, Edit and Bash.
- **REQ-SIPZ-009** (invalid manifest fails closed) — When a manifest file is present but cannot be read, parsed or validated, the guard shall deny every Write, Edit and mutating-Bash event from a self-improvement identity, with a reason that carries, in this order, the sentinel, the identity, `manifest=invalid`, `route=human` and `next=return-blocker-report`, followed by the failing file's project-relative path as the only field that may be truncated.
- **REQ-SIPZ-010** (absent manifest degrades visibly) — When no manifest file exists, the guard shall evaluate against the compiled baseline floor alone and shall record the degraded state in the audit log.
- **REQ-SIPZ-011** (denial format) — When the guard denies a call on a manifest entry, its reason shall carry, in this order, the sentinel `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION`, the identity, `category=<name>`, `route=human` and `next=return-blocker-report`, followed by the project-relative path as the only field that may be truncated, so that the whole reason is at most 240 bytes and the four routing fields are never cut; the reason labels the denial for human routing and tells the denied agent to return a blocker report instead of retrying through another tool, and the hook shall not invoke the user-question channel; a Write or Edit target that the compiled baseline already matches shall keep its existing sentinel and reason unchanged, and so carries none of these routing fields.
- **REQ-SIPZ-012** (the three kinds) — The guard shall treat a denied call whose target is a manifest-listed check, threshold or log as the occasion for human routing, deciding by the manifest membership of the target and never by classifying the content or direction of the change; the routing is carried by the `route=human` and `next=return-blocker-report` fields on every denial except a baseline-matched Write or Edit denial, which keeps its legacy sentinel and reason and carries no routing field.
- **REQ-SIPZ-013** (no agent-invocable bypass) — No environment variable (in the hook's environment or as a prefix in the command text), no extra field in the tool-call input, no command-text token or comment, and no file written by a self-improvement identity shall suppress or downgrade a protected-zone denial; the override is a human edit made outside a self-improvement identity.
- **REQ-SIPZ-014** (audit row) — When the guard denies a call or evaluates in a degraded manifest state, it shall append one JSON line to `.moai/logs/protected-zone-audit.jsonl` recording timestamp, identity, tool, project-relative path, category, decision and manifest state, and a failed append shall not change the decision.
- **REQ-SIPZ-015** (continued firing) — The repository shall carry a CI-run liveness check that fails when the PreToolUse matcher group of the local or template settings no longer contains Write, Edit and Bash, when a shipped or dogfood manifest fails to parse, when a `paths` entry (not a `runtime_paths` entry) matches no existing path in the tree its file is swept against — the shipped manifest against the template tree, the project overlay against the local tree — or when the real handler no longer denies a known zone-member input.
- **REQ-SIPZ-016** (non-regression) — Decisions for callers outside the self-improvement identity set, for tools other than Write, Edit and Bash, and for paths matched by the compiled baseline shall be unchanged by this work, including the existing sentinels and reasons.

## §C Decisions and rejected alternatives

Each decision was reached from the §A measurements and the repository's reuse ladder (reuse before new). The operator's scope note — declaration, blocking, human routing, nothing broader, no new subsystem — bounds every row.

### §C.1 Manifest location and format

**Decision.** One YAML file shipped in the template at `.moai/config/sections/protected-zone.yaml`, plus an optional add-only project overlay at `.moai/project/protected-zone.yaml`, loaded by the `internal/config` package as a dedicated-loader section (registered in the config completeness audit) and called by the hook — one owner for parsing and validation, the hook only consumes the result and only after its identity and tool gates.

**Why.** The `.moai/config/sections/` convention already carries gate policy and thresholds, so the zone's declaration sits next to what it protects and is itself inside the zone. A managed root is regenerated from the template on every `moai update`, which heals a weakened shipped manifest; a project's own entries would be deleted by the same regeneration (`AGENTS.local.md` §2.3 documents the wipe), so project-specific entries live in `.moai/project/`, a root `moai update` does not touch, and can only add.

**Rejected.**
- *Extend `.claude/lsel/frozen-allowlist.json`.* Local-only (absent from the template), LSEL-specific, regex grammar compiled per call, and it would make a user-project guard depend on a dogfood file.
- *Extend `zone-registry.md`.* Clause-level records keyed to rule text, no path field, and the loader reads the first yaml fence as the entry list — extending it changes the `moai constitution` contract.
- *A compiled Go table only.* Not a manifest, cannot be reviewed or overlaid by a project, and every project-specific path would need a binary release.
- *Dogfood paths inside the shipped file.* `internal/hook/…` and the Go test globs are this repository's layout; shipping them is the internal-content leak the template-neutrality guard exists to stop, and a user project with its own `internal/hook/` would be silently protected.
- *A typed field in `workflow.yaml`.* Requires a new config struct, defaults, drift tests and template parity for one list the hook reads once.

### §C.2 Relationship to the existing lists

**Decision.** The compiled baseline (`frozenZonePrefixes`, `frozenInstructionFiles`) stays as the degraded-mode floor and keeps its eight legacy sentinels; the manifest repeats those entries (category `moai_managed`) so the declaration is complete — except the basename `CLAUDE.local.md`, which the template-neutrality audit rejects in any shipped YAML and which therefore sits in the dogfood overlay (§D) while the compiled floor covers it in every other project — and a drift test asserts the dogfood effective zone covers every baseline entry and every entry of the `internal/harness/safety` and `internal/harness` lists. Manifest-only matches emit one new sentinel, `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION`.

**The 21 entries the drift test sweeps, by list.** `internal/hook` (10): prefixes `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, `.claude/commands/`, `.claude/hooks/`, `.claude/output-styles/`; basenames `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `AGENTS.local.md`. `internal/harness/safety` (7): `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, `.claude/settings.json`, `.claude/settings.local.json`, `internal/harness/safety/frozen_guard.go`, `internal/harness/frozen_guard.go`. `internal/harness` (4): `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/skills/moai/`, `.claude/rules/moai/`. That is 21 list entries and 15 distinct strings; the one that only the `internal/harness` list carries, `.claude/skills/moai/` (the `/moai` skill directory — the raw prefix `.claude/skills/moai-` does not cover it), is an explicit manifest entry and probe row R21. An entry counts as covered when the dogfood effective zone matches a representative path of it: the entry followed by a probe segment for a directory or prefix, `<dir>/<name>` for a basename. The test asserts the list of 21, not only the count, so a removed or renamed compiled entry fails by name.

**Rejected.** Unifying the four lists into one derived source (a refactor across three packages and a bash script — scope the operator excluded; recorded as open question Q5); unifying the LSEL allowlist (different grammar, local-only).

### §C.3 Identity set

**Decision.** Compiled constant, initially `harness-learner` alone. **Rejected:** a manifest-declared identity list — every PreToolUse call of every agent would then have to read the manifest just to learn whether the guard applies, putting file I/O on the hottest path, and a parse failure would make the identity set unknown for everyone. Widening the set is a reviewed code change (Q1).

### §C.4 Failure behaviour

**Decision.** Present-but-invalid manifest → deny every identity write (fail closed), because the blast radius is confined to autonomous self-improvement writers and a guard that silently disables itself on a parse error is the defect to avoid. Absent manifest → baseline floor plus an audit row (a project that predates the file must not lose today's protection nor be locked out). Non-identity callers are never affected (REQ-SIPZ-008/016). **Rejected:** fail-open on a parse error (the family norm for opt-in guards — wrong here because the guard is not opt-in and its silent disablement is the threat); fail-closed for everyone (a corrupt manifest would stop humans and every agent); a compiled-only fallback with no disclosure.

### §C.5 Matching semantics

**Decision.** Four entry forms and no negation; a wildcard is legal only as the trailing `*` of a raw prefix and inside the basename pattern of a `**/` entry (its semantics are fixed in REQ-SIPZ-001: `path.Match` rules on the final segment, case-folded, `[` always a class), never mid-path. Matching is ASCII case-folded on every platform (an over-block on a case-sensitive filesystem is harmless in a zone this small), NFC-normalized like the existing outside-project check, symlinks resolved on the deepest existing ancestor so a not-yet-created target is judged by where it would land. Relative input is resolved the way the existing file-access check resolves it — against the hook process's working directory (measured: a relative path under a fixture project root is judged against the process cwd, so the probe runs each case from inside its root). **Rejected:** regexes (compile cost per call; the LSEL file's `^…` anchoring has already caused a bypass once), full gitignore grammar (a new parser), platform-conditional case folding (the macOS/Windows/Linux difference then lives in test infrastructure no single CI leg covers).

### §C.6 What "modification" covers

**Decision.** Write and Edit on `file_path`; Bash with a mutating verb and a zone-covered argument or redirection target (token-level, reusing the existing shell segment splitter). **Not covered, by design and disclosed:** `MultiEdit` and `NotebookEdit` (the matcher never delivers them and an existing test forbids widening it — Q3), PowerShell (separate matcher, different syntax), and any Bash obfuscation (variables, command substitution, `python -c`, scripts, globs) — the shell rule under-matches and passes, the same stance the branch guard takes.

### §C.7 Routing to a human

**Decision.** The denial reason is the routing artifact, and the SPEC claims only what that can carry. A PreToolUse deny is delivered to the subagent that made the call, not to the user; the existing hook-block round in `agent-common-protocol.md` § Hook Invocation Surface concerns stdout-JSON blocks from three named hook scripts and does not cover it. What does exist is the subagent rule that a subagent lacking authority returns a blocker report to the orchestrator (`agent-common-protocol.md` § Subagent Prohibitions, § Blocker Report Format). So, on a routed denial (scope below), the reason (REQ-SIPZ-011) labels the denial `route=human`, tells the agent `next=return-blocker-report`, and the audit row (REQ-SIPZ-014) is the one record that does not depend on the subagent behaving. Whether a denied subagent then actually returns a blocker report is unobserved (§F G8), and no rule currently names `HARNESS_FROZEN_*` denials as a trigger for an operator question. The three kinds the card names map to manifest categories (checks → `safety_guards`, `regression_tests`, `auditor`, `apply_rollback`; thresholds → `gate_policy`, `budgets`, the auditor thresholds; logs → `logs`), and **any** modification of a listed path is the occasion for routing, not only the three kinds — direction cannot be judged soundly by a hook (raising `rate_limit.max_per_week` loosens a check, lowering `tier_thresholds` loosens another). "Regardless of zone" is realized as manifest-over-tag precedence (REQ-SIPZ-004).

**Scope of the human route (what is routed and what is not).** The `route=human` and `next=return-blocker-report` fields ride three kinds of denial: a Write or Edit whose target only the manifest matches, every shell-mutation denial (new behaviour, so there is no legacy reason to keep), and an invalid-manifest denial. A Write or Edit target that the compiled baseline already matches keeps its legacy sentinel and reason byte for byte (REQ-SIPZ-011, REQ-SIPZ-016) and carries **no** routing field. That leaves the files the card cares most about — the auditor definitions under `.claude/agents/moai/`, the gate-policy rules under `.claude/rules/moai/`, and `.claude/hooks/` — on the legacy path: their denial is a recognizable `HARNESS_FROZEN_*` sentinel and nothing more, and this work does not route it. The distinction is asserted from both sides (AC-SIPZ-001 for the routed class, AC-SIPZ-005 for the baseline class). The alternative — appending routing fields after the unchanged legacy text and relaxing the byte-for-byte clause to a prefix — would change a decision surface REQ-SIPZ-016 freezes, so it is left to a later change.

**Rejected:** a content classifier for removal/lowering/deletion (a text-pattern inference is a hypothesis, not a verified defect — `verification-claim-integrity.md` §1 — and unsound in both directions); a signed approval token or approval file (a new subsystem and a new thing a self-improver could try to write); an environment-variable or flag bypass (the branch-guard doctrine records that an in-command export is a no-op — the sentinel axis was not re-tested there — and that the spawn-name exemption is reachable from a spawned agent, so the lesson is to offer none); wiring the hook to ask the user (hooks never call the question channel).

### §C.8 Continued firing

**Decision.** A CI-run liveness test (REQ-SIPZ-015) turns each way the guard can go quiet into a red: matcher drift, manifest parse failure, a renamed file leaving a dead entry (the zone shrinking silently), and a handler that stopped denying. A deployed binary older than the fix is the one case CI cannot see; it is owned by the existing binary-lag doctor mechanism and named in §F G6 rather than re-solved here.

### §C.9 Auditor definitions

This SPEC edits no `plan-auditor` or `sync-auditor` definition. The auditors' definition files are inside the zone (`.claude/agents/moai/` is a baseline entry), which is the whole interaction; the later audit-overhaul card changes those files through a human-reviewed path and is unaffected, except that the manifest must keep listing whatever files that card creates (Q4).

## §D Initial manifest content

The shipped manifest carries the neutral entries; the dogfood overlay carries this repository's own. Each cell lists `paths` entries first and `runtime_paths` entries after a semicolon. **Sweep trees:** the shipped file is swept against the template tree (`internal/template/templates/`), the overlay against the local tree; a `paths` entry must match something there, a `runtime_paths` entry need not. An entry belongs in `runtime_paths` exactly when it is created at init, update or run time rather than shipped as a template file — verified absent from the template tree for each one below (`settings.json` exists there only as `settings.json.tmpl`; `.moai/logs/` ships there as a directory holding a `.gitkeep`, so it is a `paths` entry). The overlay's `**/CLAUDE.local.md` is a `runtime_paths` entry because the legacy file is user-owned and absent from the local tree; it sits in the overlay, not the shipped file, because the template-neutrality audit rejects that file name in any shipped YAML (the compiled baseline floor still covers it where no overlay exists). The local `.moai/config/sections/protected-zone.yaml` is the shipped file byte for byte, so it needs no tree of its own. "Probe rows" names the rows that fail at base for that entry (`acceptance.md` §E).

| Category | Shipped (template) | Dogfood overlay | Probe rows |
|---|---|---|---|
| `safety_guards` | `.claude/hooks/`; `.claude/settings.json`, `.claude/settings.local.json` | `internal/hook/`, `internal/harness/safety/`, `internal/harness/frozen_guard.go`, `.claude/lsel/`, `.moai/hooks/` | R1, R6, R7, R11, R12 |
| `gate_policy` | `.moai/config/sections/`, `.moai/config/sections/protected-zone.yaml`, `.moai/config/evaluator-profiles/`, `.claude/rules/moai/`; `.moai/project/protected-zone.yaml` | — | R2, R3, R4, R18, R19, R20 |
| `auditor` | `.claude/agents/moai/`, `.claude/workflows/sync-audit-4dim.js`, `.claude/skills/moai-*` | — | R4, R5 |
| `regression_tests` | empty | `**/*_test.go`, `**/*_test.sh`, `.github/workflows/`, `.claude/workflows/tests/` | R9, R10 |
| `apply_rollback` | empty (covered by `moai_managed`) | `internal/harness/applier.go`, `internal/harness/curator_dispatch.go`, `internal/harness/curator/`, `internal/harness/retention*`, `.claude/skills/hns-lsel-applier/`, `.claude/skills/hns-lsel-curator/` | R8, R17 |
| `budgets` | ; `.moai/harness/learning-history/rate-limit-state.json` | `internal/harness/safety/rate_limit.go`, `internal/harness/curator/budget.go` | R16, R17 |
| `logs` | `.moai/logs/`; `.moai/harness/learning-history/`, `.moai/harness/usage-log.jsonl` | ; `.moai/state/lsel/` | R13, R14, R15 |
| `moai_managed` (baseline mirror) | `.claude/agents/moai/`, `.claude/skills/moai-*`, `.claude/skills/moai/`, `.claude/rules/moai/`, `.claude/commands/`, `.claude/hooks/`, `.claude/output-styles/`, `**/CLAUDE.md`; `**/AGENTS.md`, `**/AGENTS.local.md` | ; `**/CLAUDE.local.md` | C1–C4 (must stay denied), R21 |

Paths the identity legitimately writes stay outside the zone and are probed as controls: `.claude/agents/harness/`, `.claude/skills/hns-*`, `.moai/harness/main.md`, `.moai/specs/`.

## §E Out of Scope

### Out of Scope — tools the hook never receives
- `MultiEdit` and `NotebookEdit`: the shipped matcher is `Write|Edit|Bash` and an existing template test pins it. Widening it is a separate decision (Q3).
- PowerShell mutation: a separate matcher block and a different command syntax.

### Out of Scope — writers that do not pass through a PreToolUse file call
- The in-process Go apply and rollback writers behind `moai harness apply --execute` and `moai harness rollback`; they keep their own safety-pipeline L1.
- `.moai/hooks/lsel-apply.sh` and its `frozen-allowlist.json` — not edited or unified here.
- The orchestrator's own session when it runs the learner skill (`agent_type` empty).

### Out of Scope — detection beyond the manifest
- Content-level recognition of check removal, threshold lowering or log deletion outside manifest membership.
- Obfuscated shell mutation (variables, substitution, interpreters, generated scripts, globs).
- Windows 8.3 short names, junctions, UNC prefixes and alternate data streams.

### Out of Scope — adjacent work
- Any edit to `plan-auditor` or `sync-auditor` definitions (card t1500's territory).
- Unifying the four existing frozen lists into one source; wiring or retiring `HARNESS_FROZEN_CONFIG_VIOLATION`.
- A `moai doctor` manifest-health check; widening the identity set beyond `harness-learner`.
- `.moai/reports/**` verdict files and `.moai/lessons-inbox.jsonl` as zone members.

## §F Honest gaps

- **G1 — identity reach is unobserved.** The guard fires only when a spawn's `agent_type` equals `harness-learner`. No agent definition has that name and the learner skill runs in the orchestrator session. This SPEC makes the guard correct for that identity; it does not establish that any live flow uses it. Open question Q1.
- **G2 — two writer classes remain outside.** In-process Go writers and the LSEL applier are covered only by their own lists. The LSEL list contents above were read from the JSON file, not exercised at runtime in this run.
- **G3 — shell coverage under-matches.** Disclosed in §C.6.
- **G4 — the manifest cannot be wrong-proof.** A path nobody listed is unprotected; the liveness check catches a dead entry, not a missing one.
- **G5 — no precedent is cited for case folding.** A case-fold finding was mentioned to the author as precedent, but its SPEC is not in the base tree and could not be located, so no claim here rests on it. Case folding is decided from first principles in §C.5.
- **G6 — Windows and a stale binary are not measured.** The authoring host is darwin. Windows behaviour is specified by a pure-function table test and by the CI matrix, not observed here; a binary older than the fix is the deployment case of the continued-firing rule and is owned by the binary-lag doctor check, not by this SPEC.
- **G7 — latency is a paired comparison only.** Two back-to-back runs of the same build moved the identity-write median from 110 ms to 239 ms on an earlier measurement, so absolute numbers are not a baseline. The comparison is an interleaved A/B with an A/A control (`evidence/baseline-latency.sh`, recorded in `evidence/latency-ab-base-e497f6936.txt`: control median differences 0, 1 and 0 ms).
- **G8 — the human route depends on an unobserved subagent behaviour.** The denial reaches the subagent, not the user (§C.7). A human sees it only if the denied agent returns a blocker report and the orchestrator surfaces it; no such flow was observed, no rule names these denials as a trigger, and the audit row is a trail someone must read. The SPEC makes the denial legible and instructed; it does not make it certain to reach an operator. The route is also narrower than the manifest: a baseline-matched Write or Edit denial (auditor definitions, gate-policy rules, hooks, instruction files) keeps its legacy reason with no routing field (§C.7), so for the most check-critical files the human route is the legacy sentinel alone.
