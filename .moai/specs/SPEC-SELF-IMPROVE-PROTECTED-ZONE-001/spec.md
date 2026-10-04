---
id: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001
title: "Self-improvement protected zone — declare the checking apparatus in a manifest and block self-improvement identities from modifying it"
version: "0.1.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
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

## §A Problem — what the current guard actually covers

The card asks for the whole checking apparatus (safety guards, gate policy, auditor instructions and thresholds, the regression-test bundle, the apply/rollback machinery, budgets, logs) to be declared in a manifest and for the PreToolUse guard to stop self-improvement paths from modifying it. Today the only such guard is the harness-learner FROZEN-zone check in `internal/hook/pre_tool.go` (`frozenZonePrefixes`, `frozenInstructionFiles`, `checkHarnessFrozenZone`). Measured at base SHA e497f693608ac7ea45a08b06304dc585e927ff49 by driving the real handler (`moai hook pre-tool`) with 47 fixed cases (`evidence/probe-protected-zone.sh`):

### §A.1 What is blocked today

- Four controls are denied with their legacy sentinel: a `.claude/hooks/` write, a `.claude/rules/moai/` write, the instruction file `CLAUDE.md`, and a `.claude/agents/moai/` Edit. The guard reaches only six path prefixes and four instruction basenames.
- Nothing else is denied. All 20 probe rows naming a checking-apparatus path (safety guards, gate policy, auditor workflow and profiles, Go guard code, Go tests, CI workflow, apply/rollback code, logs, budget state, and the manifest's own future paths) return `allow` — except the CI workflow, which returns `ask`, a decision a subagent cannot answer.
- `HARNESS_FROZEN_CONFIG_VIOLATION` is declared in `internal/hook/pre_tool.go` and asserted by `internal/harness/sentinel_catalog_test.go`, but no code path emits it; `.moai/config/sections/*.yaml` is therefore unguarded although the sentinel's own comment says it is.

### §A.2 The guard does not survive the path forms a real tool call carries

The matching is a raw string-prefix test on the `file_path` exactly as received. The same `.claude/hooks/` write that is denied when sent as `.claude/hooks/moai/x.sh` is **allowed** when sent as an absolute path inside the project, with a leading `./`, through a `docs/../` detour, with different letter case (`.CLAUDE/Hooks/…`), or through a symlink into the directory (5 of 5 rows, all `allow`). Claude Code's Write and Edit tools take absolute file paths, so the one guard that exists is, in the common case, matching nothing. (A prior SPEC recorded the relative-versus-absolute concern as incidental and unmeasured; this run measured it.)

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

- **REQ-SIPZ-001** (declaration) — The protected-zone manifest shall declare the checking apparatus as named categories of repository-relative path entries — at minimum `safety_guards`, `gate_policy`, `auditor`, `regression_tests`, `apply_rollback`, `budgets` and `logs` — each entry written in exactly one of four forms: an exact path, a directory prefix ending in `/`, a raw string prefix ending in `*`, or a basename glob introduced by `**/`, and each category carrying a `runtime` flag that marks entries allowed to be absent from a checkout.
- **REQ-SIPZ-002** (layering and self-inclusion) — The effective protected zone shall be the union of the shipped manifest at `.moai/config/sections/protected-zone.yaml`, the optional project overlay at `.moai/project/protected-zone.yaml`, and the compiled baseline floor, where an overlay entry can only add to the zone, and both manifest files shall list themselves as protected entries.
- **REQ-SIPZ-003** (template neutrality) — The shipped manifest shall name only language-neutral, moai-managed paths, shall ship the `regression_tests` category empty, and shall carry no SPEC identifier, requirement token, audit citation, internal date, or commit hash.
- **REQ-SIPZ-004** (precedence and non-drift) — A path listed in the effective zone shall be protected regardless of any `[ZONE:Frozen]` or `[ZONE:Evolvable]` tag on the clause or file that governs it, and every entry of the compiled baseline and of the hard-coded frozen lists in `internal/harness/safety` and `internal/harness` shall be covered by the dogfood effective zone.
- **REQ-SIPZ-005** (file-tool block) — When a Write or Edit PreToolUse event arrives from a self-improvement identity, the guard shall deny the call if the normalized target path is covered by the effective zone.
- **REQ-SIPZ-006** (normalization) — The guard shall compare a target path to the zone only after rewriting backslashes to slashes, resolving relative input as the existing file-access check resolves it and expressing the result relative to the project root, applying a lexical clean, resolving the deepest existing ancestor through symlinks, normalizing Unicode to NFC, and folding ASCII case on every platform, and a path that escapes the project root shall be treated as outside the zone and left to the existing outside-project deny.
- **REQ-SIPZ-007** (shell mutation) — When a Bash PreToolUse event arrives from a self-improvement identity and a command segment pairs a mutating verb (`rm`, `unlink`, `mv`, `cp`, `tee`, `truncate`, `sed -i`, a `>` or `>>` redirection, `git rm`, `git checkout`, `git restore`, `git apply`) with an argument or redirection target that is covered by the zone after the same normalization, the guard shall deny the call.
- **REQ-SIPZ-008** (identity scope and cost) — The guard shall apply only to the compiled self-improvement identity set, initially `harness-learner`, and shall read no manifest file for any other identity or for any tool other than Write, Edit and Bash.
- **REQ-SIPZ-009** (invalid manifest fails closed) — When a manifest file is present but cannot be read, parsed or validated, the guard shall deny every Write, Edit and mutating-Bash event from a self-improvement identity and shall name `manifest=invalid` and the failing file in the denial reason.
- **REQ-SIPZ-010** (absent manifest degrades visibly) — When no manifest file exists, the guard shall evaluate against the compiled baseline floor alone and shall record the degraded state in the audit log.
- **REQ-SIPZ-011** (denial format and human route) — When the guard denies a call on a manifest entry, its reason shall begin with the sentinel `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION` and shall carry the identity, the project-relative path, `category=<name>` and `route=human`, bounded to 240 bytes with the sentinel kept intact, and the hook shall not invoke the user-question channel; a baseline match shall keep its existing sentinel and reason unchanged.
- **REQ-SIPZ-012** (the three proposal kinds) — The guard shall route a proposal that removes a check, lowers a threshold, or deletes a log to a human by the manifest membership of the proposal's target, and shall not classify a change by its content or direction.
- **REQ-SIPZ-013** (no agent-invocable bypass) — No environment variable, tool-call field, command-text token, or file written by a self-improvement identity shall suppress or downgrade a protected-zone denial; the override is a human edit made outside a self-improvement identity.
- **REQ-SIPZ-014** (audit row) — When the guard denies a call or evaluates in a degraded manifest state, it shall append one JSON line to `.moai/logs/protected-zone-audit.jsonl` recording timestamp, identity, tool, project-relative path, category, decision and manifest state, and a failed append shall not change the decision.
- **REQ-SIPZ-015** (continued firing) — The repository shall carry a CI-run liveness check that fails when the PreToolUse matcher group of the local or template settings no longer contains Write, Edit and Bash, when a shipped or dogfood manifest fails to parse, when a non-runtime manifest entry matches no existing path in its tree, or when the real handler no longer denies a known zone-member input.
- **REQ-SIPZ-016** (non-regression) — Decisions for callers outside the self-improvement identity set, for tools other than Write, Edit and Bash, and for paths matched by the compiled baseline shall be unchanged by this work, including the existing sentinels and reasons.

## §C Decisions and rejected alternatives

Each decision was reached from the §A measurements and the repository's reuse ladder (reuse before new). The operator's scope note — declaration, blocking, human routing, nothing broader, no new subsystem — bounds every row.

### §C.1 Manifest location and format

**Decision.** One YAML file shipped in the template at `.moai/config/sections/protected-zone.yaml`, plus an optional add-only project overlay at `.moai/project/protected-zone.yaml`, read by the hook directly (the config loader ignores unknown section files) and registered as a dedicated-loader section in the config completeness audit.

**Why.** The `.moai/config/sections/` convention already carries gate policy and thresholds, so the zone's declaration sits next to what it protects and is itself inside the zone. A managed root is regenerated from the template on every `moai update`, which heals a weakened shipped manifest; a project's own entries would be deleted by the same regeneration (`AGENTS.local.md` §2.3 documents the wipe), so project-specific entries live in `.moai/project/`, a root `moai update` does not touch, and can only add.

**Rejected.**
- *Extend `.claude/lsel/frozen-allowlist.json`.* Local-only (absent from the template), LSEL-specific, regex grammar compiled per call, and it would make a user-project guard depend on a dogfood file.
- *Extend `zone-registry.md`.* Clause-level records keyed to rule text, no path field, and the loader reads the first yaml fence as the entry list — extending it changes the `moai constitution` contract.
- *A compiled Go table only.* Not a manifest, cannot be reviewed or overlaid by a project, and every project-specific path would need a binary release.
- *Dogfood paths inside the shipped file.* `internal/hook/…` and the Go test globs are this repository's layout; shipping them is the internal-content leak the template-neutrality guard exists to stop, and a user project with its own `internal/hook/` would be silently protected.
- *A typed field in `workflow.yaml`.* Requires a new config struct, defaults, drift tests and template parity for one list the hook reads once.

### §C.2 Relationship to the existing lists

**Decision.** The compiled baseline (`frozenZonePrefixes`, `frozenInstructionFiles`) stays as the degraded-mode floor and keeps its eight legacy sentinels; the manifest repeats those entries (category `moai_managed`) so the declaration is complete, and a drift test asserts the manifest covers every baseline entry and every entry of the `internal/harness/safety` and `internal/harness` lists. Manifest-only matches emit one new sentinel, `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION`.

**Rejected.** Unifying the four lists into one derived source (a refactor across three packages and a bash script — scope the operator excluded; recorded as open question Q5); unifying the LSEL allowlist (different grammar, local-only).

### §C.3 Identity set

**Decision.** Compiled constant, initially `harness-learner` alone. **Rejected:** a manifest-declared identity list — every PreToolUse call of every agent would then have to read the manifest just to learn whether the guard applies, putting file I/O on the hottest path, and a parse failure would make the identity set unknown for everyone. Widening the set is a reviewed code change (Q1).

### §C.4 Failure behaviour

**Decision.** Present-but-invalid manifest → deny every identity write (fail closed), because the blast radius is confined to autonomous self-improvement writers and a guard that silently disables itself on a parse error is the defect to avoid. Absent manifest → baseline floor plus an audit row (a project that predates the file must not lose today's protection nor be locked out). Non-identity callers are never affected (REQ-SIPZ-008/016). **Rejected:** fail-open on a parse error (the family norm for opt-in guards — wrong here because the guard is not opt-in and its silent disablement is the threat); fail-closed for everyone (a corrupt manifest would stop humans and every agent); a compiled-only fallback with no disclosure.

### §C.5 Matching semantics

**Decision.** Four entry forms, no negation, no middle wildcards, ASCII case-folded on every platform (an over-block on a case-sensitive filesystem is harmless in a zone this small), NFC-normalized like the existing outside-project check, symlinks resolved on the deepest existing ancestor so a not-yet-created target is judged by where it would land. Relative input is resolved the way the existing file-access check resolves it — against the hook process's working directory (measured: a relative path under a fixture project root is judged against the process cwd, so the probe runs each case from inside its root). **Rejected:** regexes (compile cost per call; the LSEL file's `^…` anchoring has already caused a bypass once), full gitignore grammar (a new parser), platform-conditional case folding (the macOS/Windows/Linux difference then lives in test infrastructure no single CI leg covers).

### §C.6 What "modification" covers

**Decision.** Write and Edit on `file_path`; Bash with a mutating verb and a zone-covered argument or redirection target (token-level, reusing the existing shell segment splitter). **Not covered, by design and disclosed:** `MultiEdit` and `NotebookEdit` (the matcher never delivers them and an existing test forbids widening it — Q3), PowerShell (separate matcher, different syntax), and any Bash obfuscation (variables, command substitution, `python -c`, scripts, globs) — the shell rule under-matches and passes, the same stance the branch guard takes.

### §C.7 Routing to a human

**Decision.** The denial reason is the routing artifact: the sentinel plus `category=` and `route=human`, consumed by the orchestrator's existing hook-block round (`agent-common-protocol.md` § Hook Invocation Surface) and subagent blocker-report rule. The three kinds the card names map to manifest categories (checks → `safety_guards`, `regression_tests`, `auditor`, `apply_rollback`; thresholds → `gate_policy`, `budgets`, the auditor thresholds; logs → `logs`), and **any** modification of a listed path is routed, not only the three kinds — direction cannot be judged soundly by a hook (raising `rate_limit.max_per_week` loosens a check, lowering `tier_thresholds` loosens another). "Regardless of zone" is realized as manifest-over-tag precedence (REQ-SIPZ-004). **Rejected:** a content classifier for removal/lowering/deletion (a text-pattern inference is a hypothesis, not a verified defect — `verification-claim-integrity.md` §1 — and unsound in both directions); a signed approval token or approval file (a new subsystem and a new thing a self-improver could try to write); an environment-variable or flag bypass (the branch-guard doctrine records that an in-command export is a no-op — the sentinel axis was not re-tested there — and that the spawn-name exemption is reachable from a spawned agent, so the lesson is to offer none); wiring the hook to ask the user (hooks never call the question channel).

### §C.8 Continued firing

**Decision.** A CI-run liveness test (REQ-SIPZ-015) turns each way the guard can go quiet into a red: matcher drift, manifest parse failure, a renamed file leaving a dead entry (the zone shrinking silently), and a handler that stopped denying. A deployed binary older than the fix is the one case CI cannot see; it is owned by the existing binary-lag doctor mechanism and named in §F G6 rather than re-solved here.

### §C.9 Auditor definitions

This SPEC edits no `plan-auditor` or `sync-auditor` definition. The auditors' definition files are inside the zone (`.claude/agents/moai/` is a baseline entry), which is the whole interaction; the later audit-overhaul card changes those files through a human-reviewed path and is unaffected, except that the manifest must keep listing whatever files that card creates (Q4).

## §D Initial manifest content

The shipped manifest carries the neutral entries; the dogfood overlay carries this repository's own. "RED" names the probe rows that fail at base for that entry (`acceptance.md` §E).

| Category | Shipped (template) | Dogfood overlay | Probe rows |
|---|---|---|---|
| `safety_guards` | `.claude/hooks/`, `.claude/settings.json`, `.claude/settings.local.json` | `internal/hook/`, `internal/harness/safety/`, `internal/harness/frozen_guard.go`, `.claude/lsel/`, `.moai/hooks/` | R1, R6, R7, R11, R12 |
| `gate_policy` | `.moai/config/sections/`, `.moai/config/evaluator-profiles/`, `.claude/rules/moai/`, both manifest files | — | R2, R3, R4, R18, R19, R20 |
| `auditor` | `.claude/agents/moai/`, `.claude/workflows/sync-audit-4dim.js`, `.claude/skills/moai-*` | — | R4, R5 |
| `regression_tests` | empty | `**/*_test.go`, `**/*_test.sh`, `.github/workflows/`, `.claude/workflows/tests/` | R9, R10 |
| `apply_rollback` | empty (covered by `moai_managed`) | `internal/harness/applier.go`, `internal/harness/curator_dispatch.go`, `internal/harness/curator/`, `internal/harness/retention*`, `.claude/skills/hns-lsel-applier/`, `.claude/skills/hns-lsel-curator/` | R8, R17 |
| `budgets` | `.moai/harness/learning-history/rate-limit-state.json` (runtime) | `internal/harness/safety/rate_limit.go`, `internal/harness/curator/budget.go` | R16, R17 |
| `logs` | `.moai/harness/learning-history/`, `.moai/harness/usage-log.jsonl`, `.moai/logs/` (runtime) | `.moai/state/lsel/` (runtime) | R13, R14, R15 |
| `moai_managed` (baseline mirror) | `.claude/agents/moai/`, `.claude/skills/moai-*`, `.claude/rules/moai/`, `.claude/commands/`, `.claude/hooks/`, `.claude/output-styles/`, `**/CLAUDE.md`, `**/CLAUDE.local.md`, `**/AGENTS.md`, `**/AGENTS.local.md` | — | C1–C4 (must stay denied) |

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
- **G5 — the exemption-reach and lane-fence precedents.** The leader cited a case-fold finding (F6) of `SPEC-FACTORY-LANE-FENCE-001`; that SPEC is not in the base tree and `git grep` for "fold" in its commit `53cc68ca9` returned nothing, so no claim here rests on it. Case folding is decided from first principles in §C.5.
- **G6 — Windows and a stale binary are not measured.** The authoring host is darwin. Windows behaviour is specified by a pure-function table test and by the CI matrix, not observed here; a binary older than the fix is the deployment case of the continued-firing rule and is owned by the binary-lag doctor check, not by this SPEC.
- **G7 — latency is a paired comparison only.** Two back-to-back base runs moved the identity-write median from 110 ms to 239 ms with no code change; absolute numbers are not a baseline (`evidence/latency-base-e497f6936.txt`).
