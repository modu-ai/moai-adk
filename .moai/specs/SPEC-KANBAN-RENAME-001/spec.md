---
id: SPEC-KANBAN-RENAME-001
title: "Rename Factory Mode to Kanban Mode across the moai-adk-go surface (KANBAN M0)"
version: "0.5.1"
status: completed
created: 2026-08-10
updated: 2026-08-14
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: cli
lifecycle: spec-anchored
tags: "rename, refactor, cli, template-mirror, behavior-preserving"
tier: L
related_specs: [SPEC-FACTORY-MODE-001]
---

## HISTORY

- **v0.5.1** (2026-08-11) — **Cross-reference correction inside the v0.5.0 entry below. No requirement, criterion, command, or measured figure changes.** The entry grounded the preservation of the `SPEC-FACTORY-MODE-001` citation on `REQ-KR-012`, whose protection is scoped to `AC-FM-*` identifiers in test comments and test function names — not to a SPEC identifier in a `.moai/project/` prose line. The accurate grounding is `REQ-KR-024`, which scopes those two documents to the renamed **package path, flag, and file names**; a citation to another SPEC is none of the three, so the rename never asked for it. Corrected in place rather than only retold, on the same principle the v0.4.0 entry applied to a false measurement: a wrong citation left standing in a live artifact is the defect, not its retelling. The same clause is corrected in `acceptance.md` `AC-KR-028`, which is where an auditor meets it. One figure in the same sentence is corrected with it — "one measurement short" reads **two**, which is what the entry's own numbers already said (5 bare-word baseline lines against 3 in-scope; 2 post-rename against a target of 0). `acceptance.md` moves to `0.5.1` and gains a `## HISTORY` section; `plan.md` carries no occurrence and stays at `0.5.0`.
- **v0.5.0** (2026-08-11) — **Run-phase amendment: two criteria whose premises were falsified by running them.** Neither is a scope change; no requirement or criterion is added, removed, or renumbered, so the budget position stays at **25 of 25 requirements and 28 of 25 criteria**. **(1)** `AC-KR-028`'s third command — an unbounded `grep -niI factory` over the two project documents — carried a target of `0` that **cannot be reached**. Measured after the rename landed at `768024f30`: the count is **2**, and both surviving matches are the substring `FACTORY` inside the citation `SPEC-FACTORY-MODE-001`, on the two lines `REQ-KR-024` governs. A SPEC identifier is outside what that requirement asks to change — it scopes these documents to the renamed **package path, flag, and file names**, and a citation to another SPEC is none of the three — and rewriting one would name a SPEC that does not exist and orphan the preserved `.moai/specs/SPEC-FACTORY-MODE-001/` record. The package references on both lines are already renamed. The defect is in the control, not the implementation: of the five baseline lines, **three carry only the in-scope token** (`modules.md` 157, 158, 161) and **two are dual-token** (`modules.md` 246, `structure.md` 139), and the v0.3.0 control counted all five as if each carried only the in-scope token — setting the target two measurements short. The command is bounded with `| grep -v 'SPEC-FACTORY-MODE-001'`, which reads **3** at baseline and **0** now, and still fails on a stale package reference; the reason the unbounded form cannot reach `0` is recorded beside it so a later reader does not "restore" it. `plan.md` M2 exit and M4 carry the same command and are corrected with it. **(2)** `REQ-KR-020` and its criterion rested on a **false fact about the catalog**. `plan.md` §B-3 asserted that renaming `factory.md` → `kanban.md` inside `.claude/skills/moai/` changes the `moai` skill's hash, and `AC-KR-020` accordingly expected a **non-empty** `catalog.yaml` diff. Measured: `make build` exits 0 and prints `catalog.yaml updated successfully (12403 bytes)`, after which `git status --porcelain` is **empty** — the file did not change — and `git log --oneline d39e3cdc6..HEAD -- internal/template/catalog.yaml` returns **0** commits. Read from the generator (`internal/template/scripts/gen-catalog-hashes.go` `resolveHashSourcePath`), a skill directory's hash resolves to its root `SKILL.md` **alone**, so a `workflows/`-only rename produces no delta; the plan had conflated one hash *per* directory with one hash *of* the directory. The `make build` obligation is kept, the commit obligation becomes conditional, and the criterion is re-keyed onto what actually decides the requirement — `make build` exiting 0 and `go test ./internal/template/... -count=1` passing (measured `ok … 20.081s`, `FAIL` count `0`). The stated hazard does not arise: `.claude/skills/moai/SKILL.md` carries zero mode tokens on either side, and `grep -rn 'workflows/factory'` over `.claude/`, `internal/template/templates/`, and `.moai/project/` returns **0**, so nothing dangles. **Artifacts:** `spec.md`, `plan.md`, and `acceptance.md` move to `0.5.0`; `design.md`, `research.md`, and `progress.md` are not edited and stay at `0.4.0` / `0.3.0`. Every figure above was measured in the `kanban` worktree at HEAD `768024f30`.
- **v0.4.0** (2026-08-11) — **Plan-audit delta fix (D1-D4), all four in the verification layer, none adding a requirement or a criterion.** The budget position is therefore unchanged and re-measured after the edits: **25 of 25 requirements and 28 of 25 criteria**. **(D1)** `REQ-KR-011` binds every test function naming a renamed production identifier — measured, **sixteen** of them — while its only criteria, `AC-KR-001` and `AC-KR-005`, reach **seven** through their `-run` patterns. The other **nine** carry no `$TOK` token in their names either, so after a production rename the completion grep of `AC-KR-021` reads `0` with nine test functions still announcing a mode that does not exist. `AC-KR-001` gains a bare-word grep bounded to the six surface test files (baseline **16**, target **0**), `plan.md` M1 step 6 enumerates the nine, and `design.md` §F.3 records the blind spot as its fourth. The six-file bound is the two-file reasoning of `AC-KR-028` at a slightly wider scope: the tree-wide false-positive population that forces `$TOK` to be token-scoped does not exist across six named files. **(D2)** §A.1 declared two guards — a name-existence grep and an absent-`[no tests to run]` assertion — for every `-run`-keyed criterion, and two of the four carried neither. v0.3.0 had applied them to the criteria keyed on post-rename names and exempted those keyed on rename-invariant substrings (`AC-KR-002`'s `PassThroughBoundary`, `AC-KR-009`'s `Path`), which cannot go vacuous *from the rename* — true, and not what the rule says. Both now carry both guards. `AC-KR-009` also gains the command that makes `REQ-KR-010` falsifiable: it is that requirement's **sole** criterion and covers it by an absence claim ("old path simply absent from the code"), which a run that cannot fail does not decide; an old-path grep over `internal/kanban/` is added against a measured baseline of **3** lines. **(D3)** The v0.3.0 rationale for `AC-KR-012`'s count-invariance check rested on a false premise — that the trailing filter discards *every* assertion line — and the check does not do what that premise called for. Measured: **22 of 226** assertion lines (**9.7%**) carry a filtered token, so 90% remain visible to the first command; and a `+`/`-` **count** comparison catches deletion and addition but not a **weakened predicate at constant line count**, which contributes one `+` and one `-` exactly as a rename does. The premise is corrected in item (4) of the v0.3.0 entry below rather than restated correctly only here — a false measurement left standing in a live artifact is the defect, not its retelling — the correct figures are carried in `acceptance.md`, and the predicate-weakening gap is recorded as a residual risk in `design.md` §D. **(D4)** Three test function names mix an `AC-FM-*` citation with a mode token (`TestACFM022a_Factory…` ×2, `TestACFM023c_Factory…`), so `REQ-KR-011` required renaming what `REQ-KR-012` protected and neither said which prevailed. `REQ-KR-012` gains a clause scoping its protection to the citation substring alone; the two requirements bind different substrings of one identifier and were never actually in conflict. `AC-KR-013` — that requirement's only entry in §D — was blind to precisely those three names, because a Go identifier cannot hold a hyphen and its `AC-FM-` grep (baseline **50**) matches none of the `ACFM022a` / `ACFM023c` forms; a second command holds that count at its invariant **3**. **Artifact versions:** `spec.md`, `plan.md`, `acceptance.md`, `design.md`, and `progress.md` move to `0.4.0`; `research.md` stays at `0.3.0` because it was not edited, and every measurement above was taken fresh in the worktree rather than transcribed from it.
- **v0.3.0** (2026-08-11) — **Plan-audit delta fix at 0.848 against the Tier L threshold of 0.85 — a marginal FAIL concentrated entirely in the verification layer.** Nine repairs, none of which adds a requirement or a criterion, so the budget position of v0.2.0 is unchanged: 25 of 25 requirements and 28 of 25 criteria, re-measured after the edits. **(1)** `AC-KR-020` used a **ref-less** `git diff` on `internal/template/catalog.yaml`, which is empty once the file is committed — and `plan.md` M3 instructs exactly that commit — so the criterion reported no stat line and a hash count of zero at the moment it runs while its own text asserted "and the file is committed"; anchored to `d39e3cdc6..HEAD`, and no other criterion carries the ref-less form. **(2) The repair that mattered most:** `go test -run` with a pattern matching nothing exits 0 and prints `PASS` (measured, `go1.26.4`), and `AC-KR-001` / `AC-KR-005` were both keyed on **post-rename** test names — so `REQ-KR-011`, which traces only to those two, was verified exclusively by criteria that pass when it is not done; both now pair the run with a name-existence grep and an absent-`[no tests to run]` assertion. **(3)** `AC-KR-025`'s help-text `factory` grep was **vacuous**: measured, `moai cc --help` never documented the flag, so the check returned zero before the rename too, and its missing positive control is what concealed that; reduced to the exit-0 smoke it can actually decide, since making it discriminating would need a twenty-sixth requirement. **(4)** `AC-KR-012`'s trailing `grep -viE 'kanban|factory'` discards ~~every~~ **[corrected at v0.4.0: 22 of 226, 9.7%]** assertion line in these tests, so a deleted or weakened assertion was invisible and `REQ-KR-013` had no criterion able to fail; a `+`/`-` count-invariance check on `t.Error`/`t.Fatal` lines was added, filter-independent, against a measured baseline of 226. **(5)** `AC-KR-028`'s control said "2 files" while its command counted **lines** (measured 3); `-l` added, and the disclosed line-granularity gap of `research.md` §H.4 **closed** with a bare-word grep bounded to the two named files (baseline 5) — the tree-wide false-positive objection that forces `$TOK` to be token-scoped does not arise across two files, which is the ground the v0.2.0 disclosure had borrowed. **(6)** Five references handed scope and cross-references to `SPEC-KANBAN-MULTISESSION-001`, which is **not under `.moai/specs/`** — superseded and preserved read-only — while the budget prose already named the three real siblings; each scope is now assigned to the sibling that owns it (§C, §E). **(7)** `NOTICE.md` does not exist anywhere in this tree, so §C asserted an attribution living nowhere and §D.1 listed two falsification targets that resolve to nothing, one of them returning zero vacuously; both corrected. **(8)** `progress.md` carried "26 on the real surface", contradicting its own D4 row and the measured 28. **(9)** `SPEC-KANBAN-BOARD-001` v0.3.0 resolved the `.moai/state/kanban/` path collision by moving **its** board state to `.moai/state/kanban-board/`; `REQ-KR-009` is deliberately unchanged and the coexistence is now recorded here rather than left to be re-inferred as a collision (§C).
- **v0.2.0** (2026-08-10) — **Promoted from Tier M to Tier L, and the budget position disclosed for the first time.** Measured at promotion: 25 requirements (`grep -cE '^\*\*REQ-KR-[0-9]{3}\*\*' spec.md` → `25`) and 28 acceptance criteria (`grep -cE '^\*\*AC-KR-[0-9]{3}\*\*' acceptance.md` → `28`), against a Tier M ceiling of 16 and 16 — nine and twelve over. The governing rule (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier) reads an overage as a signal to tier up or split, never to relax the budget, so the tier is raised. At Tier L the requirements fit **exactly** at 25 of 25 and the criteria remain **3 over** at 28 of 25; the residual is carried as a disclosed debt for the plan auditor to rule on, on the same terms `SPEC-KANBAN-BOOTSTRAP-001` carries its four-criterion overage. The revision closes a **disclosure gap** rather than a requirement defect: at v0.1.1 no artifact stated a count against a ceiling anywhere, which is why a nine-over/twelve-over condition survived an independent audit at 0.92 — nothing in the document invited the comparison. No requirement or criterion is added, removed, renumbered, or reworded. The promotion adds the two artifacts Tier L requires — `design.md` (the decisions, each with its rejected alternative) and `research.md` (the commands and observed outputs those decisions rest on) — and Tier L additionally raises the plan-auditor PASS threshold from 0.80 to 0.85.
- **v0.1.1** (2026-08-10) — Plan-audit delta fix (D1-D10). Frontmatter `tags` corrected from a YAML sequence to the comma-separated string the decoder declares (`internal/spec/lint.go` `Tags string`) — the sequence form made every `moai spec` verb refuse the SPEC. Completion-grep scope widened to `.moai/project/` after a whole-tree run found two stale surface files outside the enumerated 26 (§A.5). Added REQ-KR-024 (project docs) and REQ-KR-025 (bare `-f` residue). Go non-test file count corrected 7 → 8. The `-k` collision probe was run at plan-phase and returned no collision (§A.6).
- **v0.1.0** (2026-08-10) — Initial plan-phase authoring. Surface inventory measured at worktree HEAD `d39e3cdc6`; no-deprecation-alias rationale verified against the tag history and CHANGELOG.

---

## §A. Context

`SPEC-FACTORY-MODE-001` (completed) shipped **Factory Mode**: a `--factory` / `-f` entry switch on the `moai cc` and `moai glm` session launchers that opens a session pre-armed to drive a plan→run→verify→sync chain, backed by a session-scoped state record and a `factory_chain` goal preset.

The name has since been judged wrong for what the feature is becoming. The follow-on work — a six-column multi-session board with a `lead` session and N `run` sessions — is a **kanban** board, not a factory line. Renaming now, while the feature exists only in release candidates, costs one mechanical pass. Renaming later costs a deprecation window.

This SPEC is the rename **and nothing else**. The six-column multi-session orchestration lives in three follow-on SPECs, each of which declares this one in its `dependencies:` and writes every identifier in its post-rename form: `SPEC-KANBAN-BOARD-001` (the six columns, the card record, the board state store), `SPEC-KANBAN-WORKTREE-001` (per-card worktree lifecycle, holder liveness, mutual exclusion), and `SPEC-KANBAN-BOOTSTRAP-001` (session topology, bootstrap and the entry switch, the dispatch protocol). No board is designed here.

The predecessor that once carried all three, `SPEC-KANBAN-MULTISESSION-001`, was superseded during this session for being unauditable at that size; it is preserved read-only outside `.moai/specs/` and is not a live owner of anything. References to it were removed at v0.3.0 (§HISTORY).

### A.1 Rename mapping (decided; not re-litigated in this SPEC)

| Before | After |
|---|---|
| `--factory` / `-f` entry flag | `--kanban` / `-k` |
| `MOAI_FACTORY` env var | `MOAI_KANBAN` |
| `MOAI_FACTORY_SPEC` env var | `MOAI_KANBAN_SPEC` |
| `internal/factory/` package | `internal/kanban/` |
| `.moai/state/factory/` state dir | `.moai/state/kanban/` |
| `FACTORY_MODE_UNSUPPORTED_BACKEND` sentinel | `KANBAN_MODE_UNSUPPORTED_BACKEND` |
| `factory_chain` goal preset | `kanban_chain` |
| `workflows/factory.md` skill doc | `workflows/kanban.md` |
| "Factory Mode" prose | "Kanban Mode" |

### A.2 No deprecation alias — verified basis

No hidden `-f` alias and no deprecation-warning path are added. The basis was measured, not assumed:

- `internal/cli/factory.go` is **absent** from `v3.0.1` (the latest stable tag) and **present** only in `v3.1.0-rc.0` and `v3.1.0-rc.1`. Command: `git cat-file -e <tag>:internal/cli/factory.go`.
- `CHANGELOG.md` contains **zero** case-insensitive occurrences of `factory`. Command: `grep -ci factory CHANGELOG.md` → `0`.

No released user can depend on `-f`, so an alias would be dead code carrying a permanent maintenance cost.

### A.3 Measured surface (worktree HEAD `d39e3cdc6`)

Measured with the token-scoped pattern defined in §D.1. Twenty-eight files carry the Factory Mode surface: twenty-six under `internal/` and `.claude/`, plus two under `.moai/project/` (§A.5).

**Go, non-test (8):** `internal/cli/factory.go` (143 LOC), `internal/cli/cc.go`, `internal/cli/glm.go`, `internal/cli/cg.go`, `internal/cli/launcher_blockcap_infinite.go`, `internal/config/envkeys.go`, plus the `internal/factory/` package (`record.go` 184 LOC, `revision.go` 184 LOC). The count is eight files, one per path enumerated here.

**Go, test (6):** `internal/cli/cc_test.go`, `cg_test.go`, `glm_test.go`, `launcher_blockcap_infinite_test.go`, `internal/factory/record_test.go`, `revision_test.go`.

**Harness docs (6 local + 6 template mirrors):** `skills/moai/workflows/factory.md`, `skills/moai/workflows/moai.md`, `skills/moai/workflows/run.md`, `skills/moai/workflows/run/mode-orchestration.md`, `skills/moai/workflows/sync/quality-gates-quality.md`, `rules/moai/workflow/goal-directive.md`.

**docs-site: zero.** `grep -rni factory docs-site/content/` returns four hits, all the string `ExecutionFactory` inside a Claude Code best-practices example table (one per locale). None is MoAI Factory Mode. **This SPEC therefore commissions no docs-site work.**

### A.4 Mirror classification is NOT uniform byte-parity — measured

A blanket "the local and template `.claude/**` counterparts are byte-identical" premise is **false at HEAD** and must not be encoded as an invariant. Measured with `diff`:

| Pair | State |
|---|---|
| `workflows/factory.md` | byte-identical |
| `workflows/run.md` | byte-identical |
| `rules/moai/workflow/goal-directive.md` | byte-identical |
| `workflows/moai.md` | **sanitized pair** (local carries `Updated: 2026-07-09` + `Source: SPEC-MOAI-001.`; template strips both) |
| `workflows/run/mode-orchestration.md` | **sanitized pair** (local carries `Updated: 2026-03-30` + a line-wrap difference) |
| `workflows/sync/quality-gates-quality.md` | **sanitized pair** (local carries three SPEC-ID-bearing paragraphs the template strips) |

The sanitization exists because §25 Template Internal-Content Isolation forbids SPEC IDs and internal dates in template source. Forcing byte-parity would *re-introduce* the forbidden content and trip the neutrality guard. The correct invariant is **delta preservation**: the rename must leave each pair's `diff` unchanged except for the renamed tokens (§D.2).

The classification is time-varying. It is re-measured at run-phase as a pre-flight step rather than trusted from this document.

### A.5 Two surface files live outside `internal/` and `.claude/` — measured

The v0.1.0 draft scoped the completion grep to `internal/`, `.claude/`, and `internal/template/templates/`. Run whole-tree, the same pattern returns two further files:

| File | Lines | What it says |
|---|---|---|
| `.moai/project/codemaps/modules.md` | 157, 158, 161, 246 | a `### internal/factory` section heading, a role line documenting `moai cc -f` / `moai glm -f`, and an entry-point line naming `internal/cli/factory.go` |
| `.moai/project/structure.md` | 139 | a package-count paragraph naming `internal/factory` as one of the `internal/` top-level packages |

These are not unrelated vocabulary. Each names a package path and a flag that the rename deletes, so after the rename both documents describe a package that does not exist and a flag that does not work.

**Resolution: they are in scope.** The completion grep of §D.1 is widened to include `.moai/project/`, and a milestone step updates both files (REQ-KR-024). The alternative — excluding them on the rationale that `/moai project` regenerates them at sync-phase — was rejected on two grounds: neither file is template-mirrored, so nothing inside this SPEC's Definition of Done regenerates them; and both passages are hand-authored Korean prose carrying measurement narrative, which a regeneration pass would not reliably reproduce. Five lines of edits is cheaper than a documented staleness.

`.moai/specs/` stays out of the grep scope, so the closed `SPEC-FACTORY-MODE-001/` record is untouched (§C).

### A.6 The `-k` collision question is answered — probed, not assumed

`claude --help` defines no `-k` short flag. Measured at plan-phase in this worktree:

```
$ claude --help 2>&1 | grep -E '(^|[^-])-k[ ,]' ; echo "exit=$?"
exit=1                       # no match

$ claude --help 2>&1 | grep -oE '(^|[[:space:]])-[a-zA-Z][,[:space:]]' | tr -d ' ,' | sort -u
-c -d -h -n -p -r -v -w
```

The full short-flag set is `-c -d -h -n -p -r -v -w`; `-k` is free. REQ-KR-003 keeps the M0 gate anyway, because the `claude` CLI surface drifts between versions and the run-phase tree may sit on a different one — but M0 now *re-confirms* a recorded answer rather than discovering an unknown. One limitation is recorded with the probe: the pattern matches `-k ` and `-k,` renderings and would miss a `-k=<value>` form, so a null result is strong but not exhaustive.

---

## §B. Requirements (GEARS)

> Requirement count: 25 (`REQ-KR-001` … `REQ-KR-025`) — **at the Tier L ceiling of 25 exactly, with no headroom.** Any requirement added to this SPEC from here forces a split; it cannot be absorbed, because there is no tier above L. That is a live constraint on the next editor, not a formality: an implementer who discovers a twenty-sixth obligation during run-phase must either fold it into an existing requirement — which is the bundling defect `SPEC-KANBAN-BOARD-001` v0.2.0 records, where a split later deleted a rule along with the mechanism it had been bundled with — or carve the surface into a second SPEC. Acceptance criteria: 28 (`AC-KR-001` … `AC-KR-028`), **exceeding the Tier L ceiling of 25 by three**. Tier L is the top tier, so no further promotion is available; the excess is reported rather than absorbed, and re-bundling three criteria to fit the ceiling would merge separable observations back into single verdicts — precisely what §D.1.1 and `AC-KR-026` exist to keep apart. Whether to carry the excess, split this SPEC, or accept a bundled criterion is the orchestrator's and the plan auditor's decision, not this document's. Precedent: `SPEC-KANBAN-BOOTSTRAP-001` carries a four-criterion overage on the same terms. `REQ-KR-024` and `REQ-KR-025` were added at v0.1.1 and are grouped with the surface they bind (§B.4, §B.5) rather than appended at the end.

### B.1 Entry surface

**REQ-KR-001** — The session launchers shall accept `--kanban` and `-k` as the Kanban Mode entry switch, with the same parse semantics the prior switch carried: an optional following SPEC identifier is consumed, both tokens are stripped before the launcher seam, and a token appearing after a bare `--` is forwarded verbatim rather than treated as the switch.

**REQ-KR-002** — The launchers shall not accept `--factory` or `-f` as an alias, and shall emit no deprecation notice for either token, on the verified basis of §A.2.

**REQ-KR-003** — **Where** the `claude` CLI already defines a `-k` short flag, the run-phase implementer shall record the collision and surface it before proceeding, because the launcher strips `-k` before pass-through and would otherwise silently shadow the underlying flag. The plan-phase probe found no collision (§A.6); M0 re-confirms that result against the run-phase tree rather than rediscovering it.

### B.2 Identifier surface

**REQ-KR-004** — The `internal/factory` package shall be renamed to `internal/kanban`, with its package clause, its import paths at every call site, and its doc comment updated to match.

**REQ-KR-005** — The file `internal/cli/factory.go` shall be renamed to `internal/cli/kanban.go`, and its exported and unexported identifiers shall be renamed: `parseFactoryFlag` → `parseKanbanFlag`, `enterFactoryMode` → `enterKanbanMode`, `recordFactorySession` → `recordKanbanSession`, `rejectFactoryOnCG` → `rejectKanbanOnCG`, `factoryFlagLong` → `kanbanFlagLong`, `factoryFlagShort` → `kanbanFlagShort`, `factoryUnsupportedBackendSentinel` → `kanbanUnsupportedBackendSentinel`.

**REQ-KR-006** — The helper `captureEnvState` shall not be renamed, because its name carries no mode-specific token.

**REQ-KR-007** — The environment-variable names shall be carried as constants in `internal/config/envkeys.go` — `EnvMoaiFactory` → `EnvMoaiKanban` holding `"MOAI_KANBAN"`, and `EnvMoaiFactorySpec` → `EnvMoaiKanbanSpec` holding `"MOAI_KANBAN_SPEC"` — and no call site shall inline either string literal.

**REQ-KR-008** — The unsupported-backend sentinel shall be the literal `KANBAN_MODE_UNSUPPORTED_BACKEND`, and the mixed-backend rejection on `moai cg` shall continue to carry it in the returned error.

**REQ-KR-009** — The session state directory shall be `.moai/state/kanban/`, expressed through the package's existing path-segment constant rather than a new literal.

**REQ-KR-010** — The renamed surface shall migrate no pre-existing records under `.moai/state/factory/`. A record is session-scoped and best-effort by design — a launch never depends on one, and an unreadable record resolves in the safe direction (run the check rather than skip it) — so orphaned records under the old path are inert rather than a correctness hazard.

### B.3 Test surface

**REQ-KR-011** — Test function names that name a renamed production identifier shall be renamed to match it, so that a reader can still find the test from the function under test.

**REQ-KR-012** — The `AC-FM-*` acceptance-criterion identifiers appearing in test comments **and in test function names** shall **not** be renamed. They are citations to the acceptance criteria of the closed `SPEC-FACTORY-MODE-001`; rewriting them would break traceability to a record this SPEC does not amend. **Where** an identifier carries both a citation and a mode token — the three functions `TestACFM022a_FactoryRaisesBlockCapUnconditionally`, `TestACFM022a_FactoryCapReplacesPreexistingEntry`, and `TestACFM023c_FactoryEnvReachesChildEnvironment` in `internal/cli/launcher_blockcap_infinite_test.go` — this protection is scoped to the **citation substring alone** (`ACFM022a`, `ACFM023c`), and the mode token in the same identifier is still renamed under REQ-KR-011. The two requirements pull in opposite directions on exactly these three names and neither said which prevailed; the resolution is that they are not in conflict, because they bind different substrings of one identifier. `TestACFM022a_KanbanRaisesBlockCapUnconditionally` satisfies both.

**REQ-KR-013** — The behavioral assertions of the existing tests shall be preserved unchanged. Only names and prose shall move; no assertion shall be added, weakened, or removed, because this SPEC claims behavior preservation and a changed assertion would make that claim unverifiable.

### B.4 Harness documentation surface

**REQ-KR-014** — The skill document `skills/moai/workflows/factory.md` shall be renamed to `workflows/kanban.md` on both the local and the template side, and every cross-reference naming the old path shall be updated.

**REQ-KR-015** — The five sibling harness documents that reference the contract — `workflows/moai.md` (run→sync chaining policy), `workflows/run.md` (routing table), `workflows/run/mode-orchestration.md` (Verify Exit Gate), `workflows/sync/quality-gates-quality.md` (Step 0.55.0 dedup gate), and `rules/moai/workflow/goal-directive.md` (block-cap trigger 2) — shall carry the renamed contract name, flag, and mode prose.

**REQ-KR-016** — The goal preset shall be named `kanban_chain` wherever the contract document names it, and its arming rules, bounds, and blast-radius disclosure shall be carried over unchanged.

**REQ-KR-025** — The six contract documents shall carry no bare `-f` short-flag token on either the local or the template side. The completion grep of §D.1 cannot detect this residue: its pattern carries `--factory` but no `-f` alternative, because a bare `-f` added to a tree-wide pattern would match `rm -f`, `grep -f`, and `git commit -f`. The check is therefore file-scoped to the six documents (AC-KR-026). Without it, an implementer who renames `--factory` → `--kanban` and leaves `-f` untouched passes every other criterion while shipping documentation that advertises a flag the launcher no longer accepts. Measured residue at HEAD `d39e3cdc6`: **8 occurrences** — `workflows/factory.md` ×2, `workflows/moai.md` ×1, `rules/moai/workflow/goal-directive.md` ×1, and the same three on the template side.

### B.5 Template-mirror and build surface

**REQ-KR-017** — The template source under `internal/template/templates/` shall be edited before its local counterpart, per the Template-First rule.

**REQ-KR-018** — **While** applying the rename to a mirrored pair, the implementer shall preserve that pair's measured sanitization delta: a pair that was byte-identical shall remain byte-identical, and a sanitized pair shall retain exactly the same stripped content it carried before, so that the rename neither widens nor closes the pre-existing gap.

**REQ-KR-019** — The renamed template file `templates/.claude/skills/moai/workflows/kanban.md` shall contain no SPEC identifier, no REQ or AC token, no audit citation, no internal date, and no commit SHA — including no occurrence of the string `SPEC-KANBAN-RENAME-001`.

**REQ-KR-020** — **When** template source has been edited, the implementer shall run `make build`, and shall commit the regenerated `internal/template/catalog.yaml` **if it changed**. The catalog carries one `hash:` per skill directory, but that hash is computed from the directory's root `SKILL.md` **alone** — `internal/template/scripts/gen-catalog-hashes.go` `resolveHashSourcePath` stats the entry path and, when it is a directory, returns `filepath.Join(absPath, "SKILL.md")` as the sole hash source rather than walking the tree. One hash *per* skill directory is therefore not one hash *of* that directory, and a rename confined to `workflows/` cannot move it. This SPEC's rename is exactly that case: `.claude/skills/moai/SKILL.md` carries **zero** `factory` or `kanban` occurrences (`grep -ciE 'factory|kanban' .claude/skills/moai/SKILL.md` → `0`), so **no catalog delta is expected**. The obligation to run `make build` stands regardless — it is what proves the claim rather than assumes it, and it also rebuilds the binary from the edited template FS.

**REQ-KR-024** — The two project documents that name the renamed package — `.moai/project/codemaps/modules.md` and `.moai/project/structure.md` — shall carry the renamed package path, flag, and file names at the lines measured in §A.5. Neither is template-mirrored, so neither is subject to the delta-preservation invariant of REQ-KR-018.

### B.6 Completion surface

**REQ-KR-021** — The rename shall be complete when the token-scoped grep of §D.1 returns zero matches across `internal/`, `.claude/`, and `.moai/project/`. (`internal/template/templates/` is a subtree of `internal/` and is therefore already covered; `.moai/specs/` is excluded so the closed `SPEC-FACTORY-MODE-001/` record is preserved.)

**REQ-KR-022** — The verification shall run the **full** test suite (`go test ./...`), not an affected-packages subset, because a prior run-phase in this repository missed a cross-cutting template guard by testing narrowly.

**REQ-KR-023** — The rename shall introduce no functional change. No flag semantics, no state-record schema field, no gate ordering, and no goal bound shall differ from the pre-rename behavior.

---

## §C. Exclusions

### Out of Scope — the kanban board itself

- The six columns, the card record, and the board state store. That is `SPEC-KANBAN-BOARD-001`, which builds on the surface this SPEC renames.
- The per-card worktree lifecycle, holder liveness, and assignment mutual exclusion. That is `SPEC-KANBAN-WORKTREE-001`.
- The `lead` / `run` session roles, the session topology, bootstrap, and any orchestration or dispatch across N sessions. That is `SPEC-KANBAN-BOOTSTRAP-001`.
- The board state store's location. `SPEC-KANBAN-BOARD-001` `REQ-KB-005` and its §A.3(e) place board state at `.moai/state/kanban-board/`, resolved at the **primary checkout**, and explicitly decline to reuse, relocate, or amend the `.moai/state/kanban/` that `REQ-KR-009` gives the **per-tree** session record. The two were briefly one directory name with two occupants under two resolution rules; the sibling moved and this SPEC's path is deliberately unchanged, because per-tree resolution is correct for a session-scoped best-effort record (`REQ-KR-010`). Stated here so a later reader finds the coexistence recorded rather than inferring a collision that is already resolved.
- Any change to the number, ordering, or semantics of the human gates the chain already carries.

### Out of Scope — compatibility machinery

- A hidden `-f` / `--factory` alias, a deprecation warning path, or a migration shim. Excluded on the verified basis of §A.2.
- Migration of pre-existing records under `.moai/state/factory/` (REQ-KR-010).

### Out of Scope — historical records

- `.moai/specs/SPEC-FACTORY-MODE-001/` is preserved verbatim as the closed record of the feature's original delivery. It is not renamed, not rewritten, and not scanned by the §D.1 completion grep. Note the boundary: `.moai/specs/` is excluded, but `.moai/project/` is **in scope** (§A.5, REQ-KR-024) — the two directories are not treated alike.
- The `AC-FM-*` identifiers in test comments (REQ-KR-012).

### Out of Scope — unrelated "factory" vocabulary

- Generic software-pattern uses of the word across roughly 110 files — `clientFactory` in `internal/lsp/core`, the "Interface + Factory for Single Implementation" anti-pattern example, the Apache-2.0 attribution to `revfactory/harness`, the `ExecutionFactory` example in the docs-site best-practices tables, and the `@MX:ANCHOR` renderer-factory comment. None names Kanban Mode. The `revfactory/harness` attribution does **not** live in a `NOTICE.md`: measured at v0.3.0, `find . -maxdepth 3 -iname 'NOTICE*' -not -path './.git/*'` returns nothing — there is no such file anywhere in this tree — and the attribution is instead carried in five files under `.moai/research/`, a directory outside every grep scope this SPEC defines (`research.md` §H.2). The v0.2.0 text named `NOTICE.md` as its home and was wrong about a path, not about the exclusion. The §D.1 grep is token-scoped precisely so this vocabulary is excluded by construction rather than by a 110-entry allowlist.

### Out of Scope — the pre-commit gate defect

- The repository's `.git/hooks/pre-commit` invokes `moai gate`, which currently exits non-zero on pre-existing ast-grep findings and blocks all commits. The documented bypass is `SKIP_MOAI_PRECOMMIT=1`. That defect is tracked separately and is not fixed here.

---

## §D. Verification surfaces

### D.1 The completion grep (token-scoped)

The completion criterion is a grep over **Kanban Mode tokens**, not over the bare word `factory`. A bare-word grep would require an allowlist of roughly 110 files of unrelated vocabulary, which is a judgment call wearing the costume of a mechanical check.

```bash
TOK='MOAI_FACTORY|EnvMoaiFactory|factoryFlag|factoryUnsupportedBackend|FACTORY_MODE_UNSUPPORTED_BACKEND|[Ff]actory [Mm]ode|--factory|parseFactoryFlag|enterFactoryMode|recordFactorySession|rejectFactoryOnCG|internal/factory|factory_chain|workflows/factory|state/factory|package factory|[Ff]actory (contract|chain|dedup|verify stage|session|state record|pipeline)'
grep -rlniIE "$TOK" internal/ .claude/ .moai/project/
```

**Baseline (HEAD `d39e3cdc6`): 28 files** — 26 under `internal/` + `.claude/`, plus the two under `.moai/project/` identified in §A.5. **Target: 0 files.**

This fenced block is the **sole definition** of `$TOK`. `acceptance.md` AC-KR-021 carries a byte-identical copy and AC-KR-027 checks that the two copies have not drifted. The v0.1.0 draft's two copies differed — the spec.md copy carried an extra backtick-delimited `` `factory` `` alternative the acceptance.md copy omitted. Both returned 26, so the drift was inert, but it was a live divergence surface; the extra alternative was dropped so the two copies match.

The pattern was falsified against unrelated trees before adoption: run against `internal/lsp`, `internal/tui`, `internal/hook`, `internal/core`, `.claude/skills/moai/references/anti-patterns.md`, and `docs-site/`, it returns **zero** matches while the bare word still returns 9 files in `internal/lsp` and 3 occurrences in the anti-patterns reference (`research.md` §D.2). A pattern that matched the unrelated vocabulary would produce a criterion no rename could satisfy.

Two targets were dropped from that list at v0.3.0 because they were doing no work. The v0.2.0 list named `NOTICE.md`, which does not exist anywhere in this tree, and `references/anti-patterns.md`, which does not resolve from the repository root — the real path is the one written above. A falsification run against a nonexistent path returns zero **vacuously**, so two of the seven targets established nothing. The falsification itself survives on the five real ones, re-run at v0.3.0 and recorded in `research.md` §D.2; only the list is corrected.

### D.1.1 What the token grep cannot see

Three residues are invisible to this pattern by construction, and each carries its own criterion rather than a pattern extension:

- **A bare `-f`.** Adding `-f` as a global alternative would match `rm -f`, `grep -f`, and `git commit -f` across the whole tree. The check is file-scoped to the six contract documents instead (REQ-KR-025, AC-KR-026).
- **Anything under `.moai/specs/`.** Deliberately excluded so the closed `SPEC-FACTORY-MODE-001/` record survives verbatim (§C, AC-KR-024).
- **A test function name.** Added at v0.4.0. The pattern carries production identifiers, path fragments, and mode prose; a CamelCase identifier matches none of them — `TestCC_FactoryWritesStateRecord` is not `internal/factory`, not `parseFactoryFlag`, and not `[Ff]actory [Mm]ode`, which needs a space the identifier does not have. Measured, **sixteen** test functions carry the token and the grep sees none. The check is a bare-word grep bounded to the six surface test files (REQ-KR-011, AC-KR-001), on the same reasoning as the two-file bound of AC-KR-028: the ~110-file false-positive population that forces `$TOK` to be token-scoped does not exist across six named files. `design.md` §F.3 records this and the `modules.md` line-granularity residue as the third and fourth blind spots; both are closed by a bounded bare-word grep rather than by widening `$TOK`.

### D.2 Mirror delta preservation

For each of the six mirrored pairs, the `diff` output taken before the rename, with `factory`→`kanban` substitution applied to it, must equal the `diff` output taken after. This holds the three byte-identical pairs identical and the three sanitized pairs sanitized by exactly the same amount.

### D.3 Neutrality

`internal/template/internal_content_leak_test.go` and `.github/workflows/template-neutrality-check.yaml` are the mechanical authority. This SPEC adds one directed check (REQ-KR-019) but does not reimplement the guard's regex — a hand-rolled reimplementation without the guard's exemption list is a false-failure machine.

---

## §E. Cross-references

- `SPEC-FACTORY-MODE-001` — the closed SPEC that delivered the surface being renamed. Preserved as a historical record.
- `SPEC-KANBAN-BOARD-001` — the six columns, the card record, and the board state store. Declares this SPEC in `dependencies:`, and places its own state at `.moai/state/kanban-board/` rather than the `.moai/state/kanban/` of `REQ-KR-009` (§C).
- `SPEC-KANBAN-WORKTREE-001` — the per-card worktree lifecycle, holder liveness, and assignment mutual exclusion. Declares this SPEC in `dependencies:`.
- `SPEC-KANBAN-BOOTSTRAP-001` — session topology, bootstrap, the entry switch, and the dispatch protocol. Declares this SPEC in `dependencies:`.

Together these three are the follow-on that motivates the rename; none is authored here. They replace the single `SPEC-KANBAN-MULTISESSION-001` that preceded them, which is superseded and preserved read-only outside `.moai/specs/`.
- `CLAUDE.local.md` §2 (Template-First), §14 (env constants), §25 (Template Internal-Content Isolation).
