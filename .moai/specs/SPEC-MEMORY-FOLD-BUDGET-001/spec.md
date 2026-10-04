---
id: SPEC-MEMORY-FOLD-BUDGET-001
title: "Memory hygiene pass 1 — card-done fold into the archive index, byte-aware index budget, link-class report"
version: "0.2.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
author: GOOS행님
priority: P1
phase: "v3.2.0 target"
module: "internal/cli (memory, todo) + internal/hook/memo/taxonomy + internal/hook (session start) + internal/config"
lifecycle: spec-anchored
tags: "memory-hygiene, memory-fold, index-budget, reachability, session-start, link-report, card-t1502"
tier: M
related_specs: [SPEC-MEMORY-INDEX-FOLD-001, SPEC-MEMORY-DIET-001, SPEC-MEMORY-STORE-RECONCILE-001]
---

## HISTORY

- 2026-10-04 — v0.2.0 — manager-spec — Plan delta for plan-audit iteration 1 (FAIL 0.69, 12 blocking findings D1-D12, 8 optional D13-D20; report `.moai/reports/t1502/plan-audit.md`) and the leader ruling (`.moai/reports/t1502/leader-disposition.md`). Scope split (option A): the `relink` verb, the nearest-name suggestion, the repo-relative-link rewrite and the similarity threshold (old REQ-MFB-014, old M6, OD-8, OD-10) move to a follow-up card; the doctor keeps classifying and reporting link classes accurately (card item 3 reduced to "report"). Closed in this revision: D1 (§1.5 index set redefined on the doctor's own secondary-index rule; fifth lossy mutant; refusal when the archive index would fall under the threshold), D3 (doctrine edit dropped from this card), D4 (execution bound, seeded-panic and blocked-read cells), D5 (hook-package home sandbox and a never-opened cell), D6 (decision index classes, defaults, OD-6 row), D7 (STRONG lines filed verbatim, description rewrite deleted), D9 (named key `index_link_targets`), D10 (differential assertion instead of a golden recording), D12 (single-file RED cells); D2 and D11 left with the split-out half; optional D13-D20 per `plan.md` §L. REQ count 15 → 14 (old REQ-MFB-014 removed, old REQ-MFB-015 renumbered REQ-MFB-014); AC count 15 → 14. Budget finding code `MEMORY_INDEX_BUDGET_OVER` renamed `MEMORY_INDEX_BUDGET_AT_CAP` (D19).
- 2026-10-04 — v0.1.0 — manager-spec — Plan-phase artifacts authored for card t1502 (Tier M: spec.md + plan.md + acceptance.md; plus progress.md, the conditional decision-index.md because `interview.decision_gate` is `on`, and a synthetic `fixtures/store-A/` read by the acceptance commands). Operator directive v3.2 redesign step 1. Tree baseline: HEAD `2f492df19`. `status: draft`.

---

## Prior-Art Review

Three completed SPECs cover the memory index. None covers a mechanical fold, a byte-aware budget, or a link-class report, so the verdict for this SPEC is **NEW** (no amendment, no supersede): all three are `completed`, their bodies are immutable post-completion, and each did a one-time manual job this SPEC turns into a repeatable command.

| Prior SPEC | Status | What it did | Relationship to this SPEC | Verdict |
|---|---|---|---|---|
| `SPEC-MEMORY-INDEX-FOLD-001` (v1.0.0, Tier S) | completed | One-time MANUAL repair: appended 3 dropped index lines verbatim to the secondary indexes and recorded a no-additional-fold judgment. No code. Measured in characters (`python3 len()`) and records that `wc -m` returns bytes on the measuring machine. | This SPEC mechanizes the act that SPEC performed by hand — append verbatim lines to a secondary index, then verify containment — and keeps its discipline that `MEMORY.md` edits never lose a link. Its unit (characters) is not overridden: this SPEC reports characters, loaded-content characters and bytes side by side (§1.4). | new |
| `SPEC-MEMORY-DIET-001` (v1.0.0, Tier M) | completed | One-time prune of always-loaded surfaces; its REQ-3 pruned `MEMORY.md` by hand through `_archive/`. | Distinct mechanism: that SPEC chose entries by judgment; this one folds by a mechanical, card-keyed rule and never prunes or archives a topic file. | new |
| `SPEC-MEMORY-STORE-RECONCILE-001` (v0.3.0, Tier M) | completed | Corrected the budget premise (its M2: the 25,600-byte cut is not confirmed), removed the vacuous token-guard slot, and copied the dangling targets between the two stores. Its REQ-MSR-008 forbids the always-loaded token guard (`internal/config/token_budget_guard.go`) from carrying a constant that encodes the unconfirmed cut. Its exclusion "index dieting" declined entry shortening because the diet could not be verified by an entry-line metric that sees only a subset of the file's unique targets. | **Constraint-bearing.** (1) REQ-MSR-008 binds the token guard only; this SPEC does not touch the guard, and its byte cap is an operator-tunable advisory threshold on a different surface, stated as a conservative proxy (§1.4), never as the loader's cut. (2) The verification gap that exclusion cited is closed here: invariants (b) and (c) of §1.5 count unique link targets file-wide across the whole index set, not anchored entry lines. (3) This SPEC does not settle the loader's cut shape (Out of Scope below). | new |

---

## 1. Background and measured premises

### 1.1 What the card asks

Memory hygiene pass 1 (operator directive, v3.2 redesign step 1): (1) a `moai memory fold --card <id>` that moves a closed card's open-work line(s) from `MEMORY.md` into the card-archive index, wired into card close; (2) a byte-aware index budget in `moai memory doctor` with a warning at 80 % and a SessionStart warning; (3) accurate classification and **reporting** of dangling and repo-relative links. Item (3) is reduced by the leader ruling to "report": the repair of those links (a `relink` verb) moves to a follow-up card (Out of Scope). Index lines are shortened or moved and **never deleted**: the lesson-loss prohibition of `.claude/rules/moai/workflow/moai-memory.md` § Compressing the index means making entries shorter — never fewer, and § Admission (acceptance is reachability, never size), stay in force.

### 1.2 Measured premises (this tree, HEAD `2f492df19`, 2026-10-04)

| Ref | Premise | Evidence |
|---|---|---|
| P1 | `fold` is absent from the memory command. | `grep -c -i fold internal/cli/memory.go` prints `0`; positive control `grep -c -i doctor internal/cli/memory.go` prints `8`. `bin/moai memory fold --card t9001 --dir <fixture>` exits 1 with `Unknown flag: --card` (acceptance.md E1). |
| P2 | Doctor measures lines only. | `internal/cli/memory.go:263` (`rep.IndexLines`), render at `:309`; the report JSON has no byte, character or link-target-count key (E2). |
| P3 | Cap constants. | `internal/config/defaults.go:500` `DefaultMemoryIndexLineCap = 200`; `:502` `DefaultMemoryTopicFileCap = 50`. `AuditIndex` (`internal/hook/memo/taxonomy/audit.go:161`) has exactly one non-test caller, `internal/cli/memory.go:274`, which passes `0` (the default cap); it fires `MEMORY_INDEX_OVERFLOW` when lines exceed the cap, only inside doctor, never in a hook. |
| P4 | Card close has three paths, not two, and they share only a record method. | `rec.ArchiveCard` is called at `internal/cli/todo.go:1104` (`todo done`, reached as `moai gtd done` too — `bin/moai gtd done --help` renders the same help body), `internal/cli/todo_autodone.go:397` (`todo auto-done`) and `internal/cli/todo_auto.go:331` (the `todo --auto` cycle). The only shared code is `BacklogRecord.ArchiveCard` (`internal/factory/backlog_store.go:397`), which runs inside the locked `Mutate` callback. A side effect there would run before the queue write is known to have landed, so the seam is each path's post-`Mutate` success point, calling one shared helper. |
| P5 | SessionStart has a visible advisory surface. | `appendAdditionalContext` (`internal/hook/session_start_binary_lag.go:96`) is used by the binary-lag advisory (`session_start.go:599`) and the guard-liveness advisory (`:613`). The `HookOutput.Data` map is `json:"-"` and reaches nobody (`session_start.go` comment at the deferred-scan join), so the budget line must use `additionalContext`. |
| P6 | Doctor's link bookkeeping collapses distinct repo-relative targets. | `internal/hook/memo/taxonomy/linkage.go:99` keys targets by `filepath.Base`. On `fixtures/store-A`, two distinct repo-relative targets (`.moai/reports/t9002/verdict.md`, `.moai/reports/t9006/verdict.md`) yield ONE finding, `index links verdict.md but no such file exists` (E2), naming neither path. The same base-name keying means a repo-relative target whose base name equals an existing store file counts as that store file for the doctor; the fixture has no such target, and §1.5 classifies by full text, so this is the one place the checker and the doctor can differ (Residual risk, plan.md §G). |
| P7 | The doctor qualifies a secondary index only at a link threshold. | `secondaryIndexLinkThreshold = 3` (`internal/hook/memo/taxonomy/linkage.go:56`), applied at `:147`, counts the distinct store files a file links that exist, excluding itself and `MEMORY.md`. Scratch measurements: (i) the fixture with its archive index cut to one link reports 5 `MEMORY_ORPHAN_NOT_INDEXED` instead of 2, so fold must never create a fresh archive index; (ii) the lossy mutant "file the `t9001` line into a linked topic file" leaves `MEMORY.md` at 821 bytes, passes the v0.1.0 invariants, and the doctor reports 3 orphans against the baseline's 2 (acceptance.md §5). |
| P8 | `moai memory archive` is not a model for the fold's apply path. | It applies immediately with no preview and selects the store as `stores[0]` even when absent (`memory.go:361`); its line remover deletes whole lines by base name (`memory.go:419-458`). `moai memory drain` is the preview-by-default (`--yes` applies) convention the fold follows. Reusable from it: `memoryCandidateStores`, and the link regex `markdownLinkTargetCLI` (`memory.go:463`). |
| P9 | The store-derivation sites are pinned. | `TestHomeJoinSiteCountIsPinned` (`internal/hook/home_isolation_test.go`) holds an explicit allowlist of files that join the home directory with a `.claude/projects` slug; a new hook-side site needs an allowlist row and a real-home guard test. |
| P10 | The hook package's `TestMain` does not isolate the home directory. | `internal/hook/main_test.go:67-96` unsets `CLAUDE_PROJECT_DIR`, `MOAI_PROFILE_LEASE_TOKEN` and `CLAUDE_CONFIG_DIR` and sandboxes `MOAI_HOME`; `grep -c "USERPROFILE" internal/hook/main_test.go` prints `0` (acceptance.md E7a). `internal/cli/main_test.go:264` defines `sandboxUserHomeDir`, which is why the card-close wiring is safe there. The new SessionStart advisory runs inside every existing test that drives the SessionStart handler. |
| P11 | A doctor-style read of an index file that blocks on read blocks the caller. | A FIFO named `MEMORY.md` in a temporary store: `timeout 5 ./bin/moai memory doctor --dir <store>` exited 124 with empty stdout (measured on this tree, scratch store, not the real one). Consequence: a fold or advisory that reads the same file without a bound can stall the command that hosts it. |

### 1.3 Card-reported figures (not re-measured here)

The card reports a 1,465-file store, a 25,735-byte `MEMORY.md` at 102.9 % of "25KB", 27 of 38 open-work lines being closed cards, 278 unreachable files, and dangling lines. Those figures describe the operator's real store, which this card is forbidden to read; they are motivation, dated 2026-10-04, **no requirement or acceptance criterion here depends on them**, and the same applies to the investigation note's per-section figures. Every acceptance command runs against `fixtures/store-A/` or a test temporary directory.

### 1.4 The budget-unit premise (stated, not papered over)

The card asks for a byte budget. What is known and unknown:

- Known (doctrine, quoted): the host announced that the index truncates at 200 lines or 25KB under Claude Code 2.1.83, and later changed the over-limit warning to measure loaded content, excluding frontmatter and HTML comments (2.1.211) — `moai-memory.md` § MEMORY.md Index Budget.
- Known (a completed SPEC): `SPEC-MEMORY-STORE-RECONCILE-001` M2 (`spec.md:61`) — a 26,280-byte index (18,463 characters, 163 lines) loaded whole with its final line present, so a raw-byte cut at 25,600 is **not confirmed**; a character cap, a line-only cap, or a larger byte cap each remain possible.
- **Unknown, and unmeasured by this SPEC:** the unit and size of the loader's actual cut, and which of "25KB" = 25,000 or 25,600 bytes the host means.

Therefore the design is: the doctor reports four measures — raw **bytes**, **characters** (Unicode code points), **loaded-content characters** (a reconstruction of the documented exclusion: leading YAML frontmatter and HTML comments removed; informational only, unverified) and **lines**. The 80 % warning is **keyed on raw bytes and on lines**, because bytes is never smaller than characters, so it warns earliest and errs toward a false alarm on CJK-heavy indexes, which is the safe side for an advisory. Raw bytes is a **conservative proxy, not a confirmed cut**; every finding says so in its own text. The cap values are configuration constants (C-3) with per-invocation flag overrides; the byte cap takes the smaller reading of "25KB" (25,000) for the same reason. Because the thresholds are advisory (they emit warnings and never enforce), the 25,000-versus-25,600 ambiguity changes only when a warning appears, never what is lost. The finding at 100 % of the byte cap is named `MEMORY_INDEX_BUDGET_AT_CAP`, not "over", because the cap is a proxy: a store loaded whole at 26,280 bytes is on record. Decisions: plan.md §A.1 (OD-3, OD-4).

### 1.5 Reachability model (the primary invariant)

For a store directory S (the directory holding `MEMORY.md` and the topic files at its top level):

- **Link target**: the text between `](` and `)` of every markdown link whose target ends in `.md`, on any line shape of any file (grouped lines and mid-line links included — never only lines beginning `- [`).
- **Store-local target**: a bare file name, optionally `./`-prefixed. **Absolute target**: begins with `/`. **Repo-relative target**: any other target containing a path separator.
- **Resolved link of a file f**: a store-local target of f naming an existing store file other than f itself and other than `MEMORY.md`; links are counted as distinct files.
- **Index set I(S)** = `MEMORY.md` ∪ every store file carrying at least `secondaryIndexLinkThreshold` resolved links. This is the doctor's own rule: the constant is defined at `internal/hook/memo/taxonomy/linkage.go:56` (value 3 at the pinned tree) and applied at `:147`, and this SPEC reuses that constant rather than a second number, so the checker and the doctor cannot disagree on which files carry reachability. (v0.1.0 defined I(S) as the files `MEMORY.md` links, which admitted a lossy transformation; see P7 and acceptance.md §5.)
- **Reachable set R(S)** = the store files that appear as a store-local target in any member of I(S).
- **Target set T(S)** = the distinct full target strings occurring in any member of I(S). Counted by full text, never by base name (P6).

Every operation that rewrites `MEMORY.md` or an index (the fold) must satisfy, measured before versus after on the same store:

- **(a)** R(after) ⊇ R(before) — no file becomes unreachable.
- **(b)** T(after) ⊇ T(before) — no link target string disappears from the index set. This SPEC never rewrites a target, so the comparison is over raw strings with no exception.
- **(c)** every line removed from `MEMORY.md` has an equal-target line (same set of link targets on one line) in another member of I(after).

A shrink in bytes, characters or lines is the motive and **never satisfies any criterion of this SPEC**; on `fixtures/store-A` a correct fold and four different lossy transformations produce the same size reduction (acceptance.md §5).

### 1.6 Store resolution

One store per operation, always named in the output. Precedence: an explicit `--dir`; otherwise the first candidate from `memoryCandidateStores` (the profile `CLAUDE_CONFIG_DIR` key, then the default `~/.claude` key, each for the session working directory and, inside a linked worktree, the primary checkout) whose directory contains `MEMORY.md`. Which candidate the host actually loads inside a worktree is unsettled by this repository's own doctrine (the rule and the code comments disagree); this SPEC does not settle it, so it names the store it acted on instead of asserting it is the loaded one.

---

## 2. Requirements (GEARS)

### 2.1 Fold

- **REQ-MFB-001** — **When** an operator runs `moai memory fold --card <id>`, the fold command shall resolve exactly one store (§1.6), print `store: <dir> (<origin>)` as the first line of output, compute a fold plan, and write no file unless `--yes` is also given; `--json` emits the plan as one JSON object. `<id>` is `t<digits>` or bare digits normalized to `t<digits>`; any other value is refused before the store is read.
- **REQ-MFB-002** — The fold command shall classify each `MEMORY.md` line against the card id and move only **STRONG** lines: a line is STRONG when it begins `- [`, its first link title begins with the card id as a whole token, and at least one of its link targets contains the card id as a whole token while no target of the line contains a different card id; a line meeting exactly one of the two conditions is **AMBIGUOUS**; a line naming the card id only elsewhere in its text is a **MENTION**. AMBIGUOUS and MENTION lines are listed as kept, each with its reason, and are never moved. A whole token is bounded by the string edge or a character outside `[0-9A-Za-z]`.
- **REQ-MFB-003** — The fold command shall file each STRONG line **verbatim** — every byte of the line, title, targets and trailing text — at the end of the **archive index**, in the original order, where the archive index is the store-root file with the greatest name matching `project_card_archive_<YYYY>_<MM>.md` that `MEMORY.md` itself links as a store-local target. The command shall not create an archive index and shall not alter any other file.
- **REQ-MFB-004** — **While** applying (`--yes`), the fold command shall (1) append the lines to the archive index, (2) re-read the archive index and confirm each moved line's link-target set is present on one line, then (3) rewrite `MEMORY.md` without those lines, each file replaced atomically so that no reader observes a partial file; it shall abort without writing `MEMORY.md` when `MEMORY.md` or the archive index differs from the content read when the plan was computed, and shall abort before any write when the resulting store would violate invariant (a), (b) or (c) of §1.5. After every apply, invariants (a), (b) and (c) hold.
- **REQ-MFB-005** — **When** `fold --yes` runs for a card already folded, or after an apply that stopped between steps (1) and (3), the fold command shall complete without duplicating an archive line: a STRONG line whose link-target set already appears on one line of the archive index is removed from `MEMORY.md` without a second append, and a repeat run on a fully folded card exits 0 reporting that nothing remains to fold.
- **REQ-MFB-006** — The fold command shall respond to edge inputs as follows: no `MEMORY.md` line names the card → exit 0, write nothing, report `no line`; only AMBIGUOUS or MENTION lines → exit 0, write nothing, list them; no archive index exists, or the greatest-named one is not linked from `MEMORY.md` → exit non-zero, write nothing, and name the file to create or link; the archive index would carry, after the fold, fewer resolved links than the doctor's secondary-index threshold (§1.5) → exit non-zero, write nothing, and name the archive index, its resolved-link count after the fold and the threshold; an invalid id → exit non-zero before any read.

### 2.2 Card-close wiring

- **REQ-MFB-007** — **Where** fold-on-done is enabled — an environment-variable gate that the operator sets truthy, and that is off when unset — **when** a queue close path has archived a card and its `Mutate` call has returned success, the CLI shall, outside the queue lock, run fold apply for that card id against one store and write one stderr line naming the store, the archive file and the number of lines filed. The step shall be bounded by an execution bound (a configuration constant, C-3): **when** it has not finished within the bound the CLI shall abandon it and begin no further write. A failure of the step — an error, a panic, a store file that blocks on read, or the bound being reached — shall produce at most one stderr line and shall not change the close path's stdout, exit status or queue record, nor delay the close path's return beyond the bound. **Where** the gate is disabled the close paths shall perform no memory read and no memory write. The close paths are `todo done` (also `gtd done`), `todo auto-done` (once per closed card) and the `todo --auto` cycle.

### 2.3 Doctor budget

- **REQ-MFB-008** — **When** `moai memory doctor` reports a store whose `MEMORY.md` exists, it shall report, in text and in `--json`, the raw byte count, the character count, the loaded-content character count and the line count, as `index_bytes`, `index_chars`, `index_loaded_chars` and the existing `index_lines`.
- **REQ-MFB-009** — **When** the byte count or the line count of `MEMORY.md` reaches the warn percentage of its cap (integer test `value*100 >= warnPercent*cap`), the doctor shall emit `MEMORY_INDEX_BUDGET_WARN` for that axis; **when** the byte count reaches the byte cap it shall emit `MEMORY_INDEX_BUDGET_AT_CAP` instead of the warning for the byte axis; on the line axis the existing `MEMORY_INDEX_OVERFLOW` (lines above the cap) is emitted instead of the warning and is otherwise unchanged. The line cap given to the budget check also governs `MEMORY_INDEX_OVERFLOW`, so one line cap is in force per invocation. Each budget finding shall name the axis, the measured value, the cap, the percentage and the basis (`raw bytes: conservative proxy; the loader's cut is unconfirmed`). The byte cap, line cap and warn percentage shall be configuration values (C-3), overridable per invocation by `--byte-cap`, `--line-cap` and `--warn-percent`, and shall not appear as literals in the check.
- **REQ-MFB-010** — The doctor shall leave the topic-file count, the topic-file cap (50, overridable by `--cap`) and the `MEMORY_TOPIC_COUNT_OVER_CAP` finding unchanged. This SPEC neither raises, removes nor reinterprets that cap (its noise at large store sizes is recorded in Out of Scope).

### 2.4 SessionStart warning

- **REQ-MFB-011** — **When** a session starts and the store resolved for it (§1.6) holds a `MEMORY.md` at or above the warn percentage on bytes or lines, the SessionStart hook shall add exactly one line, prefixed `[moai:memory-budget]`, to the session's `additionalContext` and to no other output channel, naming the store path, the larger of the two percentages, the measure it is keyed on, and `moai memory doctor`; below the threshold, with no store, on any read error, or with `MOAI_MEMORY_AUDIT=0`, it shall add nothing. The check shall be bounded by a join bound (a configuration constant, C-3), read one file, never block or fail session start, and never read a path outside the stores derived for the session.
- **REQ-MFB-012** — The SessionStart budget check shall run on every session start — startup, resume, clear and compact alike — with no path, changed-file, branch or source condition.

### 2.5 Link-class report

- **REQ-MFB-013** — The doctor shall classify every link target of every member of the index set (§1.5) as store-local, absolute or repo-relative; shall report a store-local target with no file as `MEMORY_DANGLING_INDEX_LINK` naming the target; shall report each distinct repo-relative target as `MEMORY_REPO_RELATIVE_LINK` naming the full target and shall not report it as dangling; and shall report the number of distinct link targets of `MEMORY.md`, counted by full text, as `index_link_targets` in text and in `--json`. The doctor shall not rewrite, retarget or suggest a replacement for any link.

### 2.6 Global invariant

- **REQ-MFB-014** — No operation introduced by this SPEC shall delete a topic file, move a topic file, rewrite a link target, or remove a link; the only index-line removal is the fold's `MEMORY.md` removal under REQ-MFB-004 (c); and every automated test and acceptance command of this SPEC shall operate only on a store under a test temporary directory or on `fixtures/store-A/` read-only — never on the operator's real memory store.

---

## 3. Constraints

- **C-1 Real store off limits.** Implementation, tests, acceptance commands and every run-phase step of this card never read, modify or write the operator's real memory directory. Applying the feature to the real store is a separate step, taken after the leader confirms. Tests isolate `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR` and `MOAI_HOME` to temporary directories and assert the resolved store sits beneath the temporary root (precedent: `TestMemoryCandidateStores_DoNotEscapeToRealHome`, `internal/cli/memory_test.go`). This binds the pre-existing hook-package tests too, which run the new advisory: the hook package's `TestMain` sandboxes the home directory (P10, plan.md M5).
- **C-2 Advisory paths fail open and are bounded.** The SessionStart check and the card-close wiring never block, never delay beyond their named bound (REQ-MFB-007, REQ-MFB-011), never alter a close path's stdout or exit status, and never call `AskUserQuestion` (`.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline; `internal/cli/CLAUDE.md`, `internal/hook/CLAUDE.md` subagent boundary).
- **C-3 Configuration, not literals.** Thresholds, caps and execution bounds are constants in `internal/config/defaults.go`; the gate's environment-variable name is a constant in `internal/config/envkeys.go`. Config must have a real consumer (the typed-memory-config precedent recorded at `defaults.go:496-499` removed inert config).
- **C-4 Concurrency.** `MEMORY.md` is also written by the host's native memory subsystem, so a read-modify-write can lose an update. The content re-check of REQ-MFB-004 narrows the window; it does not close it, and the residual is named in plan.md §G rather than claimed away.
- **C-5 Cross-platform.** Path handling uses `filepath`; `GOOS=windows GOARCH=amd64 go build ./...` passes.
- **C-6 No doctrine or template edit.** This card edits no rule file and no template mirror. The `moai-memory.md` pair already differs (`cmp` exit 1 at line 200, a deliberate machine-specific-value difference) and no requirement here needs a doctrine change; the doctrine pointer is a follow-up (Out of Scope).
- **C-7 No time estimates** anywhere in the artifacts; ordering is by phase and priority.

---

## 4. Exclusions

### Out of Scope — link repair (split out to a follow-up card)

- The `moai memory relink` verb and everything it carried: retargeting a dangling store-local link to its nearest file, the nearest-name suggestion on a dangling finding, the token-set similarity measure and its threshold (`DefaultMemoryLinkRepairMinSimilarity`, 0.75 in v0.1.0, supported by one measured pair only), the rewrite of a repo-relative link to an absolute path, and the decision of what a repo-relative link becomes (v0.1.0 OD-8 and OD-10). The follow-up card must also carry the card-id-token guard that the audit found necessary: two names differing only in the card id must never qualify as nearest (audit finding D11). This card keeps only the doctor's classification and report (REQ-MFB-013).

### Out of Scope — doctrine pointer

- The paragraph in `.claude/rules/moai/workflow/moai-memory.md` § MEMORY.md Index Budget (and its template mirror) naming the new doctor figures and the fold verb. Dropped from this card because no requirement or criterion asks for it and the pair already differs; the follow-up must state its own measurable pair-parity criterion and name the template-leak guard (`internal/template/card_id_leak_test.go`) and `make build`.

### Out of Scope — re-indexing unreachable topic files

- Appending index lines for the files no index reaches (the investigation's `reindex` proposal), and any triage of orphans. Follow-up card.

### Out of Scope — store merging and the second store

- Merging the profile store with the legacy `~/.claude` store, and any SessionStart notice about the non-loaded store. Follow-up card.

### Out of Scope — duplicate detection, merging and promotion

- Clustering `feedback_*` files, merging near-duplicates, and promoting recurring lessons to guards or rules.

### Out of Scope — memory language policy

- The Korean-versus-English question for the index and topic files. This SPEC edits no wording of any entry (the fold moves lines verbatim).

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

The acceptance criteria live in `acceptance.md` (AC-MFB-001 … AC-MFB-014). The primary criterion is **reachability**: after any fold over a temporary store, invariants (a), (b) and (c) of §1.5 hold, on the index set the doctor itself uses; size is never a gate.

## 6. Cross-references

- `.claude/rules/moai/workflow/moai-memory.md` § MEMORY.md Index Budget, § Compressing the index means making entries shorter — never fewer, § Admission.
- `.claude/rules/moai/development/verification-completeness.md` § 1.2 (three-part check spec), § 1.3 (continued firing), § 2 (two-cell adoption), § 2.1 (RED-now four elements).
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.2 (tool provenance), § 2.3 (ordering attribution).
- `plan.md` §A.1 (decisions and defaults), `decision-index.md` (questions the interview left unsettled and their dispositions).
