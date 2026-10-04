---
id: SPEC-MEMORY-FOLD-BUDGET-001
title: "Memory hygiene pass 1 — card-done fold into the archive index, byte-aware index budget, link repair"
version: "0.1.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
author: GOOS행님
priority: P1
phase: "v3.2.0 target"
module: "internal/cli (memory, todo) + internal/hook/memo/taxonomy + internal/hook (session start) + internal/config"
lifecycle: spec-anchored
tags: "memory-hygiene, memory-fold, index-budget, reachability, session-start, link-repair, card-t1502"
tier: M
related_specs: [SPEC-MEMORY-INDEX-FOLD-001, SPEC-MEMORY-DIET-001, SPEC-MEMORY-STORE-RECONCILE-001]
---

## HISTORY

- 2026-10-04 — v0.1.0 — manager-spec — Plan-phase artifacts authored for card t1502 (Tier M: spec.md + plan.md + acceptance.md; plus progress.md, the conditional decision-index.md because `interview.decision_gate` is `on`, and a synthetic `fixtures/store-A/` read by the acceptance commands). Operator directive v3.2 redesign step 1. Tree baseline: HEAD `2f492df19`. `status: draft`.

---

## Prior-Art Review

Three completed SPECs cover the memory index. None covers a mechanical fold, a byte-aware budget, or link repair, so the verdict for this SPEC is **NEW** (no amendment, no supersede): all three are `completed`, their bodies are immutable post-completion, and each did a one-time manual job this SPEC turns into a repeatable command.

| Prior SPEC | Status | What it did | Relationship to this SPEC | Verdict |
|---|---|---|---|---|
| `SPEC-MEMORY-INDEX-FOLD-001` (v1.0.0, Tier S) | completed | One-time MANUAL repair: appended 3 dropped index lines verbatim to the secondary indexes and recorded a no-additional-fold judgment. No code. Measured in characters (`python3 len()`) and records that `wc -m` returns bytes on the measuring machine. | This SPEC mechanizes the act that SPEC performed by hand — append verbatim lines to a secondary index, then verify containment — and keeps its discipline that `MEMORY.md` edits never lose a link. Its unit (characters) is not overridden: this SPEC reports characters, loaded-content characters and bytes side by side (§1.4). | new |
| `SPEC-MEMORY-DIET-001` (v1.0.0, Tier M) | completed | One-time prune of always-loaded surfaces; its REQ-3 pruned `MEMORY.md` by hand through `_archive/`. | Distinct mechanism: that SPEC chose entries by judgment; this one folds by a mechanical, card-keyed rule and never prunes or archives a topic file. | new |
| `SPEC-MEMORY-STORE-RECONCILE-001` (v0.3.0, Tier M) | completed | Corrected the budget premise (its M2: the 25,600-byte cut is not confirmed), removed the vacuous token-guard slot, and copied the dangling targets between the two stores. Its REQ-MSR-008 forbids the always-loaded token guard (`internal/config/token_budget_guard.go`) from carrying a constant that encodes the unconfirmed cut. Its exclusion "index dieting" declined entry shortening because the diet could not be verified by an entry-line metric that sees only a subset of the file's unique targets. | **Constraint-bearing.** (1) REQ-MSR-008 binds the token guard only; this SPEC does not touch the guard, and its byte cap is an operator-tunable advisory threshold on a different surface, stated as a conservative proxy (§1.4), never as the loader's cut. (2) The verification gap that exclusion cited is closed here: invariants (b) and (c) of §1.5 count unique link targets file-wide across the whole index set, not anchored entry lines. (3) This SPEC does not settle the loader's cut shape (Out of Scope below). | new |

---

## 1. Background and measured premises

### 1.1 What the card asks

Memory hygiene pass 1 (operator directive, v3.2 redesign step 1): (1) a `moai memory fold --card <id>` that moves a closed card's open-work line(s) from `MEMORY.md` into the card-archive index, wired into card close; (2) a byte-aware index budget in `moai memory doctor` with a warning at 80 % and a SessionStart warning; (3) repair of dangling and repo-relative links. Index lines are shortened or moved and **never deleted**: the lesson-loss prohibition of `.claude/rules/moai/workflow/moai-memory.md` § Compressing the index means making entries shorter — never fewer, and § Admission (acceptance is reachability, never size), stay in force.

### 1.2 Measured premises (this tree, HEAD `2f492df19`, 2026-10-04)

| Ref | Premise | Evidence |
|---|---|---|
| P1 | `fold` is absent from the memory command. | `grep -c -i fold internal/cli/memory.go` prints `0`; positive control `grep -c -i doctor internal/cli/memory.go` prints `8`. `bin/moai memory fold --card t9001 --dir <fixture>` exits 1 with `Unknown flag: --card` (acceptance.md E1). |
| P2 | Doctor measures lines only. | `internal/cli/memory.go:263` (`rep.IndexLines`), render at `:309`; the report JSON has no byte or character key (E2). |
| P3 | Cap constants. | `internal/config/defaults.go:500` `DefaultMemoryIndexLineCap = 200`; `:502` `DefaultMemoryTopicFileCap = 50`. `AuditIndex` (`internal/hook/memo/taxonomy/audit.go:161`) has exactly one non-test caller, `internal/cli/memory.go:274`; `MEMORY_INDEX_OVERFLOW` fires only inside doctor, never in a hook. |
| P4 | Card close has three paths, not two, and they share only a record method. | `rec.ArchiveCard` is called at `internal/cli/todo.go:1104` (`todo done`, reached as `moai gtd done` too — `bin/moai gtd done --help` renders the same help body), `internal/cli/todo_autodone.go:397` (`todo auto-done`) and `internal/cli/todo_auto.go:331` (the `todo --auto` cycle). The only shared code is `BacklogRecord.ArchiveCard` (`internal/factory/backlog_store.go:397`), which runs inside the locked `Mutate` callback. A side effect there would run before the queue write is known to have landed, so the seam is each path's post-`Mutate` success point, calling one shared helper. |
| P5 | SessionStart has a visible advisory surface. | `appendAdditionalContext` (`internal/hook/session_start_binary_lag.go:96`) is used by the binary-lag advisory (`session_start.go:599`) and the guard-liveness advisory (`:613`). The `HookOutput.Data` map is `json:"-"` and reaches nobody (`session_start.go` comment at the deferred-scan join), so the budget line must use `additionalContext`. |
| P6 | Doctor's link bookkeeping collapses distinct repo-relative targets. | `internal/hook/memo/taxonomy/linkage.go:99` keys targets by `filepath.Base`. On `fixtures/store-A`, two distinct repo-relative targets (`.moai/reports/t9002/verdict.md`, `.moai/reports/t9006/verdict.md`) yield ONE finding, `index links verdict.md but no such file exists` (E2), naming neither path. |
| P7 | A new archive index with fewer than three resolvable links is invisible to the orphan audit. | `secondaryIndexLinkThreshold = 3` (`linkage.go:56`). Scratch measurement: the fixture with its archive index cut to one link reports 5 `MEMORY_ORPHAN_NOT_INDEXED` instead of 2. Consequence: fold must never create a fresh archive index (§2 REQ-MFB-003, OD-5). |
| P8 | `moai memory archive` is not a model for the fold's apply path. | It applies immediately with no preview and selects the store as `stores[0]` even when absent (`memory.go:361`); its line remover deletes whole lines by base name (`memory.go:419-458`). `moai memory drain` is the preview-by-default (`--yes` applies) convention the fold follows. Reusable from it: `memoryCandidateStores`, and the link regex `markdownLinkTargetCLI` (`memory.go:463`). |
| P9 | The store-derivation sites are pinned. | `TestHomeJoinSiteCountIsPinned` (`internal/hook/home_isolation_test.go`) holds an explicit allowlist of files that join the home directory with a `.claude/projects` slug; a new hook-side site needs an allowlist row and a real-home guard test. |

### 1.3 Card-reported figures (not re-measured here)

The card reports a 1,465-file store, a 25,735-byte `MEMORY.md` at 102.9 % of "25KB", 27 of 38 open-work lines being closed cards, 278 unreachable files, and dangling lines. Those figures describe the operator's real store, which this card is forbidden to read; they are motivation, dated 2026-10-04, **no requirement or acceptance criterion here depends on them**, and the same applies to the investigation note's per-section figures. Every acceptance command runs against `fixtures/store-A/` or a test temporary directory.

### 1.4 The budget-unit premise (stated, not papered over)

The card asks for a byte budget. What is known and unknown:

- Known (doctrine, quoted): the host announced that the index truncates at 200 lines or 25KB under Claude Code 2.1.83, and later changed the over-limit warning to measure loaded content, excluding frontmatter and HTML comments (2.1.211) — `moai-memory.md` § MEMORY.md Index Budget.
- Known (a completed SPEC): `SPEC-MEMORY-STORE-RECONCILE-001` M2 — a 26,280-byte index loaded whole with its final line present, so a raw-byte cut at 25,600 is **not confirmed**; a character cap, a line-only cap, or a larger byte cap each remain possible.
- **Unknown, and unmeasured by this SPEC:** the unit and size of the loader's actual cut, and which of "25KB" = 25,000 or 25,600 bytes the host means.

Therefore the design is: the doctor reports four measures — raw **bytes**, **characters** (Unicode code points), **loaded-content characters** (a reconstruction of the documented exclusion: leading YAML frontmatter and HTML comments removed; informational only, unverified) and **lines**. The 80 % warning is **keyed on raw bytes and on lines**, because bytes is never smaller than characters, so it warns earliest and errs toward a false alarm on CJK-heavy indexes, which is the safe side for an advisory. Raw bytes is a **conservative proxy, not a confirmed cut**; every finding says so in its own text. The cap values are configuration constants in `internal/config/defaults.go` (never literals in the check) with per-invocation flag overrides; the byte cap takes the smaller reading of "25KB" (25,000) for the same reason. Because the thresholds are advisory (they emit warnings and never enforce), the 25,000-versus-25,600 ambiguity changes only when a warning appears, never what is lost. Open decisions: OD-3, OD-4 (plan.md §A.1).

### 1.5 Reachability model (the primary invariant)

For a store directory S (the directory holding `MEMORY.md` and the topic files at its top level):

- **Link target**: the text between `](` and `)` of every markdown link whose target ends in `.md`, on any line shape of any file (grouped lines and mid-line links included — never only lines beginning `- [`).
- **Store-local target**: a bare file name, optionally `./`-prefixed. **Absolute target**: begins with `/`. **Repo-relative target**: any other target containing a path separator.
- **Index set I(S)** = `MEMORY.md` ∪ the store files `MEMORY.md` links as store-local targets.
- **Reachable set R(S)** = the store files that appear as a store-local target in any member of I(S) (one hop from `MEMORY.md` through its linked files).
- **Target set T(S)** = the distinct full target strings occurring in any member of I(S). Counted by full text, never by base name (P6).

Every operation that rewrites `MEMORY.md` or an index (fold, relink) must satisfy, measured before versus after on the same store:

- **(a)** R(after) ⊇ R(before) — no file becomes unreachable.
- **(b)** T(after) ⊇ T(before) — no link target disappears (for relink, over targets that resolve to an existing file, because its job is to retarget dead ones).
- **(c)** every line removed from `MEMORY.md` has an equal-target line (same set of link targets on one line) in another member of I(after).

A shrink in bytes, characters or lines is the motive and **never satisfies any criterion of this SPEC**; on `fixtures/store-A` a correct fold and a fold that silently deletes the lines produce the same size reduction (acceptance.md mutant record).

### 1.6 Store resolution

One store per operation, always named in the output. Precedence: an explicit `--dir`; otherwise the first candidate from `memoryCandidateStores` (the profile `CLAUDE_CONFIG_DIR` key, then the default `~/.claude` key, each for the session working directory and, inside a linked worktree, the primary checkout) whose directory contains `MEMORY.md`. Which candidate the host actually loads inside a worktree is unsettled by this repository's own doctrine (the rule and the code comments disagree); this SPEC does not settle it, so it names the store it acted on instead of asserting it is the loaded one.

---

## 2. Requirements (GEARS)

### 2.1 Fold

- **REQ-MFB-001** — **When** an operator runs `moai memory fold --card <id>`, the fold command shall resolve exactly one store (§1.6), print `store: <dir> (<origin>)` as the first line of output, compute a fold plan, and write no file unless `--yes` is also given (preview by default, as `moai memory drain`; `--json` emits the plan as one JSON object). `<id>` is `t<digits>` or bare digits normalized to `t<digits>`; any other value is refused before the store is read. Open decision OD-6 (plan.md §A.1); this text states the recommended default.
- **REQ-MFB-002** — The fold command shall classify each `MEMORY.md` line against the card id and move only **STRONG** lines: a line is STRONG when it begins `- [`, its first link title begins with the card id as a whole token, and at least one of its link targets contains the card id as a whole token while no target of the line contains a different card id; a line meeting exactly one of the two conditions is **AMBIGUOUS**; a line naming the card id only elsewhere in its text is a **MENTION**. AMBIGUOUS and MENTION lines are listed as kept, each with its reason, and are never moved. A whole token is bounded by the string edge or a character outside `[0-9A-Za-z]`. Open decision OD-7.
- **REQ-MFB-003** — The fold command shall file each STRONG line at the end of the **archive index** — the store-root file with the greatest name matching `project_card_archive_<YYYY>_<MM>.md` that `MEMORY.md` itself links as a store-local target — in the original order, replacing only the line's trailing description (the text after the final link when it begins with ` — `) by the `description` of the first link target's frontmatter when that target is a store file carrying a non-empty one, and leaving the title and every link target byte-for-byte unchanged; any other line shape is filed verbatim. The command shall not create an archive index. Open decision OD-5.
- **REQ-MFB-004** — **While** applying (`--yes`), the fold command shall (1) append the rewritten lines to the archive index, (2) re-read the archive index and confirm each moved line's link-target set is present on one line, then (3) rewrite `MEMORY.md` without those lines — each file replaced atomically (temp file in the same directory, then rename) — and shall abort without writing `MEMORY.md` when the SHA-256 of `MEMORY.md` or of the archive index differs from the value read when the plan was computed. After every apply, invariants (a), (b) and (c) of §1.5 hold.
- **REQ-MFB-005** — **When** `fold --yes` runs for a card already folded, or after an apply that stopped between steps (1) and (3), the fold command shall complete without duplicating an archive line: a STRONG line whose link-target set already appears on one line of the archive index is removed from `MEMORY.md` without a second append, and a repeat run on a fully folded card exits 0 reporting that nothing remains to fold.
- **REQ-MFB-006** — The fold command shall respond to edge inputs as follows: no `MEMORY.md` line names the card → exit 0, write nothing, report `no line`; only AMBIGUOUS or MENTION lines → exit 0, write nothing, list them; no archive index exists, or the greatest-named one is not linked from `MEMORY.md` → exit non-zero, write nothing, and name the file to create or link; an invalid id → exit non-zero before any read.

### 2.2 Card-close wiring

- **REQ-MFB-007** — **Where** fold-on-done is enabled (the environment variable named by `config.EnvMemoryFoldOnDone` is truthy; the compiled default, `config.DefaultMemoryFoldOnDone`, is disabled), **when** a queue close path has archived a card and its `Mutate` call has returned success, the CLI shall — outside the queue lock — run fold apply for that card id against one store and write one stderr line naming the store, the archive file and the number of lines filed; a failure of that step shall produce one stderr line and shall not change the close path's stdout, exit status, or queue record. **Where** the gate is disabled the close paths shall perform no memory read and no memory write. The close paths are `todo done` (also `gtd done`), `todo auto-done` (once per closed card) and the `todo --auto` cycle. Open decisions OD-1, OD-2.

### 2.3 Doctor budget

- **REQ-MFB-008** — **When** `moai memory doctor` reports a store whose `MEMORY.md` exists, it shall report, in text and in `--json`, the raw byte count, the character count, the loaded-content character count and the line count, as `index_bytes`, `index_chars`, `index_loaded_chars` and the existing `index_lines`.
- **REQ-MFB-009** — **When** the byte count or the line count of `MEMORY.md` reaches the warn percentage of its cap (integer test `value*100 >= warnPercent*cap`), the doctor shall emit `MEMORY_INDEX_BUDGET_WARN`; **when** the byte count reaches 100 % of the byte cap it shall emit `MEMORY_INDEX_BUDGET_OVER` instead for that axis (the line-axis over-cap case stays the existing `MEMORY_INDEX_OVERFLOW`, unchanged). Each finding shall name the axis, the measured value, the cap, the percentage and the basis (`raw bytes: conservative proxy; the loader's cut is unconfirmed`). The byte cap, line cap and warn percentage shall be configuration values in `internal/config/defaults.go`, overridable per invocation by `--byte-cap`, `--line-cap` and `--warn-percent`, and shall not appear as literals in the check. Open decisions OD-3, OD-4.
- **REQ-MFB-010** — The doctor shall leave the topic-file count, the topic-file cap (`DefaultMemoryTopicFileCap`, 50, overridable by `--cap`) and the `MEMORY_TOPIC_COUNT_OVER_CAP` finding unchanged. This SPEC neither raises, removes nor reinterprets that cap (its noise at large store sizes is recorded in Out of Scope).

### 2.4 SessionStart warning

- **REQ-MFB-011** — **When** a session starts and the store resolved for it (§1.6) holds a `MEMORY.md` at or above the warn percentage on bytes or lines, the SessionStart hook shall add exactly one line, prefixed `[moai:memory-budget]`, to the session's `additionalContext`, naming the store path, the larger of the two percentages, the measure it is keyed on, and `moai memory doctor`; below the threshold, with no store, on any read error, or with `MOAI_MEMORY_AUDIT=0`, it shall add nothing. The check shall be bounded by a join bound, read one file, and never block or fail session start. Open decision OD-9.
- **REQ-MFB-012** — The SessionStart budget check shall run on every session start — startup, resume, clear and compact alike — with no path, changed-file, branch or source condition.

### 2.5 Link repair

- **REQ-MFB-013** — The doctor shall classify every link target of `MEMORY.md` and of the files it links as store-local, absolute or repo-relative (§1.5); shall report a store-local target with no file as `MEMORY_DANGLING_INDEX_LINK` naming the target and, when exactly one store file scores at or above `config.DefaultMemoryLinkRepairMinSimilarity` (token-set Jaccard of the file names split on `_`, `-`, `.`), that file as the nearest name; shall report each distinct repo-relative target as `MEMORY_REPO_RELATIVE_LINK` naming the full target and shall not report it as dangling; and shall count targets by full text. Open decision OD-8.
- **REQ-MFB-014** — **When** an operator runs `moai memory relink` (preview by default; `--yes` applies; `--dir` and `--json` as `fold`), the command shall, for each link of `MEMORY.md` and of the files it links: retarget a dangling store-local link to its unique nearest file; rewrite a repo-relative link whose target exists under the project root to its absolute path; leave every other link unchanged and report it. It shall never remove a line or a link, and after apply invariant (a) and invariant (b) over resolvable targets of §1.5 hold, the count of link occurrences is unchanged, and the line count is unchanged. Open decisions OD-8, OD-10.

### 2.6 Global invariant

- **REQ-MFB-015** — No operation introduced by this SPEC shall delete a topic file, move a topic file, or remove a link; the only index-line removal is the fold's `MEMORY.md` removal under REQ-MFB-004 (c); and every automated test and acceptance command of this SPEC shall operate only on a store under a test temporary directory or on `fixtures/store-A/` read-only — never on the operator's real memory store.

---

## 3. Constraints

- **C-1 Real store off limits.** Implementation, tests, acceptance commands and every run-phase step of this card never read, modify or write the operator's real memory directory. Applying the feature to the real store is a separate step, taken after the leader confirms. Tests isolate `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR` and `MOAI_HOME` to temporary directories and assert the resolved store sits beneath the temporary root (precedent: `TestMemoryCandidateStores_DoNotEscapeToRealHome`, `internal/cli/memory_test.go`).
- **C-2 Advisory paths fail open.** The SessionStart check and the card-close wiring never block, delay unboundedly, alter a close path's stdout or exit status, or call `AskUserQuestion` (`.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline; `internal/cli/CLAUDE.md`, `internal/hook/CLAUDE.md` subagent boundary).
- **C-3 Configuration, not literals.** Thresholds are constants in `internal/config/defaults.go`; the gate's environment-variable name is a constant in `internal/config/envkeys.go`. Config must have a real consumer (the typed-memory-config precedent recorded at `defaults.go:496-499` removed inert config).
- **C-4 Concurrency.** `MEMORY.md` is also written by the host's native memory subsystem, so a read-modify-write can lose an update. The SHA-256 re-check of REQ-MFB-004 narrows the window; it does not close it, and the residual is named in plan.md §G rather than claimed away.
- **C-5 Cross-platform.** Path handling uses `filepath`; `GOOS=windows GOARCH=amd64 go build ./...` passes.
- **C-6 Doctrine mirror.** Any edit of `.claude/rules/moai/workflow/moai-memory.md` is mirrored byte-identically in `internal/template/templates/.claude/rules/moai/workflow/moai-memory.md`, in the re-measuring-command-first form with no machine-specific value (the `SPEC-MEMORY-STORE-RECONCILE-001` R4 discipline).
- **C-7 No time estimates** anywhere in the artifacts; ordering is by phase and priority.

---

## 4. Exclusions

### Out of Scope — re-indexing unreachable topic files

- Appending index lines for the files no index reaches (the investigation's `reindex` proposal), and any triage of orphans. Follow-up card.

### Out of Scope — store merging and the second store

- Merging the profile store with the legacy `~/.claude` store, and any SessionStart notice about the non-loaded store. Follow-up card.

### Out of Scope — duplicate detection, merging and promotion

- Clustering `feedback_*` files, merging near-duplicates, and promoting recurring lessons to guards or rules.

### Out of Scope — memory language policy

- The Korean-versus-English question for the index and topic files. This SPEC edits no wording of any entry beyond the fold's description replacement.

### Out of Scope — per-section budget and re-sectioning

- An open-work-section byte budget, and any heading-based parsing of `MEMORY.md`. The real store's headings are not guaranteed to follow one convention, and fold removes the main driver of open-work growth; revisit after the language-policy follow-up.

### Out of Scope — the topic-file cap

- Replacing the 50-file cap with reachability and index-size findings. REQ-MFB-010 freezes it; at large store sizes its advice is noise, which this card does not fix.

### Out of Scope — the loader's cut

- Determining whether the host cuts by bytes, characters, loaded content or lines, and the 25,000-versus-25,600 reading of "25KB". That needs a deliberate truncation experiment (as `SPEC-MEMORY-STORE-RECONCILE-001` also declined).

### Out of Scope — automation around the fold

- Creating the next month's archive index, routing a card-bound lesson to the lesson index at fold time, write-time (admission) enforcement, periodic consolidation, and applying the feature to the real store.

### Out of Scope — documentation surfaces

- The docs-site CLI reference (four locales) and README updates for the new verbs belong to the sync phase, not to this SPEC's requirements.

---

## 5. Success criteria

The acceptance criteria live in `acceptance.md` (AC-MFB-001 … AC-MFB-015). The primary criterion is **reachability**: after any fold or relink over a temporary store, invariants (a), (b) and (c) of §1.5 hold; size is never a gate.

## 6. Cross-references

- `.claude/rules/moai/workflow/moai-memory.md` § MEMORY.md Index Budget, § Compressing the index means making entries shorter — never fewer, § Admission.
- `.claude/rules/moai/development/verification-completeness.md` § 1.2 (three-part check spec), § 1.3 (continued firing), § 2 (two-cell adoption), § 2.1 (RED-now four elements).
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.2 (tool provenance).
- `plan.md` §A.1 (open decisions), `decision-index.md` (questions the interview left unsettled).
