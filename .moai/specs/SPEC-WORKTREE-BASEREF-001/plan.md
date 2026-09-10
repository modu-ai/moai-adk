# SPEC-WORKTREE-BASEREF-001 — Implementation Plan

Card: t313 · Tier: M · Tree measured: worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`, HEAD `48eb945df`.
Evidence directory for later phases: `.moai/reports/t313/`.

---

## §A Design Decision Record

The decisions below are the reversible part of this plan and are stated first, per Rule 1 ordering. The milestones in §C are mechanical once these hold.

### D1 — The reproducible path is a consumer, not a pass-through key [decided, from measurement]

**Conflict.** The card asked for a "reproducible configuration path" for the worktree base branch. Measurement forecloses the obvious shape: `worktree.baseRef` "accepts only `"fresh"` or `"head"`, not arbitrary refs" (`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md:171`; local mirror same line). A Claude Code settings key naming `develop` is therefore impossible — there is no key that accepts it.

**Replacement.** The only handle `EnterWorktree`'s `fresh` mode reads is `refs/remotes/origin/HEAD`, which is local repository metadata and does not propagate to a fresh clone. So the stored setting lives in moai's own config, and a moai-owned consumer applies it to that handle at a predictable moment. The setting is reproducible (it is in a file, in git); the metadata mutation is derived from it rather than hand-applied.

**What this costs.** The mutation is still local, so a fresh clone gets the correct base only after the first session start in that clone. That is inherent to the handle, not to this design.

### D2 — Where the key lives: `git_strategy.worktree_base_branch` (root level) [decided]

The key is a git branch name, so `git_strategy` is the owning section. Within it, three placements were available:

| Placement | Against |
|---|---|
| `git_strategy.<mode>.worktree_base_branch` (per-mode) | Triples the value for one repository-wide fact; `ActiveModeProfile()` (`internal/config/types.go:171-183`) would then decide the base branch, coupling worktree creation to git-strategy mode for no benefit |
| `workflow.worktree.base_branch` | The neighbouring keys there (`auto_create` / `auto_merge` / `auto_cleanup` / `tmux_preferred`, pinned at `internal/web/dead_config_guard_test.go:72-75`) are automation toggles, not git refs |
| **`git_strategy.worktree_base_branch` (root)** | **Chosen** — peer of `mode` / `provider` / `github_username` (`.moai/config/sections/git-strategy.yaml:1-6`), one key for one repository-wide fact, and `typedField(SectionGitStrategy, "git_strategy", "worktree_base_branch", …)` maps to it with no new persistence machinery (`internal/settings/schema_sections.go:99-110`) |

#### D2.1 — The widget shape: `TypeText` [decided by operator ruling, 2026-08-27]

**The measurement that made this a question.** The operator asked for "three choices: main / develop / free-text". The schema's `FieldType` set is `select` / `multi-select` / `radio` / `text` / `int` / `float` / `bool` (`internal/settings/schema.go:105-113`). There is no combo widget — a closed option set and a free-text input are different types, and no existing field mixes them. The two available resolutions were (a) `TypeText` with `main` and `develop` named in the field description, and (b) a new combo widget (`TypeSelect` plus an "other…" escape revealing a text input).

**Ruling: (a) `TypeText`.** The `FieldType` set at `internal/settings/schema.go:105-113` stays exactly as it is — no eighth type is added, and no render branch is added near `schemaRadioRow`. The control is a free-text field; `main` and `develop` are named in its description as the two branches in common use. The "three choices" framing becomes two named suggestions plus free text.

**Why the closed-set alternative is rejected — and this reason must survive.** A `TypeSelect` with `main` / `develop` as its options would bake two repository-specific branch names into the **shipped** schema. `internal/settings/` is template-managed and reaches every downstream project, so a user whose default branch is `trunk`, `master`, or anything else could not select their own branch — the control would be unusable for them. That is a template-neutrality violation of exactly the kind CLAUDE.local.md §2.1 and §15 forbid, and it is the same reason REQ-WBR-003 forbids the shipped template default from naming a branch. This paragraph exists so a later reader does not "improve" the free-text field into a picker: the free text is not a shortcut taken for lack of a better widget, it is the only shape that stays neutral.

**What this costs, stated honestly.** One click. A user on `main` or `develop` types the branch name rather than selecting it, and a typo becomes possible where a picker would have made it impossible. REQ-WBR-009's pre-write resolvability check is what converts that typo from a corrupted `refs/remotes/origin/HEAD` into one diagnostic line and a no-op — it is the price of the neutral widget, not an optional extra.

Consequences recorded elsewhere: REQ-WBR-014 (free-text control), REQ-WBR-009 + AC-WBR-015 (resolvability check), REQ-WBR-012 (doctor reports the unresolvable case separately), AC-WBR-011 (now unconditional).

### D3 — When consumer 1 fires: SessionStart, inside the existing best-effort group

`Handle` (`internal/hook/session_start.go:66`) already runs four independent best-effort tasks in an `errgroup` (`:120-175`), each writing to its own local map and merging after `Wait()`. The alignment step is a fifth task of exactly that shape: it touches only `refs/remotes/origin/HEAD`, shares no file with Tasks 1-4, and its failure contract ("never returns a non-nil error"; `:176-181`) is already the group's contract.

The notice line surfaces on stderr, matching the existing precedent at `:94-98` (the empty-session_id warning) rather than inventing a new channel.

**Why announce.** The operator explicitly rejected silent mutation. A change to git metadata is a change to what every subsequent worktree inherits; the one-line notice is what makes it attributable rather than mysterious. REQ-WBR-006 keeps the no-op path silent so the notice stays a signal.

### D3.1 — Consumer 1 fires only from the primary checkout [decided, plan-audit iter-1 D4]

The alignment step mutates `refs/remotes/origin/HEAD`, which is repository-global, from a setting stored in a **tracked** per-worktree file. Left unnarrowed, every active lane is a writer of one shared handle. REQ-WBR-004's second clause gates the step on the primary checkout, discriminated by `git rev-parse --git-dir` vs `--git-common-dir` (`internal/cli/session_worktree.go:234-241`). The full analysis — the tracked-file measurement, why write-only-on-difference does not confine the hazard, and the residual risk that remains — is §E R1. Recorded here because it is a design decision, not merely a mitigation: the SPEC deliberately gives up firing consumer 1 in seven of eight lanes, and loses nothing by it, because one writer of a repository-global handle is sufficient.

### D3.2 — The read seam is the alignment-entry (configured-value) read [decided, plan-audit iter-2 N2]

**The ambiguity.** AC-WBR-016 half 1 asserts the alignment path's read seam is invoked exactly once per `Handle`, with a `Given` admitting **any** configured value including empty. "The read seam" had two available denotations and they disagree on precisely that empty branch: if the seam is the `refs/remotes/origin/HEAD` read, a compliant implementation short-circuits before it (M2 orders the helper `read config → no-op silently on empty → read origin/HEAD`, and REQ-WBR-005 forbids a git-metadata read on the empty path), records **0** invocations, and fails a criterion it satisfies. If the seam is the entry helper, it holds for every value.

**Ruling: the read seam is the alignment-ENTRY seam — the function-variable that reads the configured `git_strategy.worktree_base_branch` value — and nothing else.** The `origin/HEAD` read is a separate seam AC-WBR-016 does not assert on.

**Why this reading and not the narrowing alternative.** The rejected alternative was to narrow half 1's `Given` to a **non-empty** value. That resolves the contradiction, but it also makes the criterion evaporate exactly where it is most needed: with the setting unset — the shipped default (REQ-WBR-003), and therefore the common case — an implementation that never registers the errgroup task would pass. AC-WBR-016 exists solely to catch that failure (REQ-WBR-004; plan-audit iter-1 D1), so a resolution that lets it lapse on the default configuration is not a resolution. The entry-seam reading keeps the assertion "the task is wired into the group at all" true for every value.

**Consequence for the implementation shape.** The primary-checkout gate (D3.1) precedes the configured-value read, so from a linked worktree the entry seam is invoked **0** times, which is what AC-WBR-016 half 2 asserts. The two halves therefore read the same seam under one denotation, and it is stated in the criterion itself (`acceptance.md`, AC-WBR-016 preamble) so the run-phase test and the criterion cannot disagree — the audit's explicit instruction was that this must not be settled by silently writing the permissive test.

### D4 — Consumer 2's insertion point: the seam, not the caller

`materializeSessionWorktree` calls through the overridable seam `sessionWorktreeGitWorktreeAdd` (`internal/cli/session_worktree.go:53`, invoked at `:192`), whose real implementation is `gitWorktreeAddReal` (`:217`). The base operand is a property of the git invocation, so it belongs in the real implementation, and the seam's signature grows by one string parameter so tests can assert what was passed.

**Fail-back.** An unresolvable configured ref must degrade to today's no-operand call, not to an error: a worktree that exists on the wrong base is recoverable, a worktree that failed to materialize blocks the lane. "Unresolvable" is decided by the SAME predicate consumer 1 uses (REQ-WBR-009 — `git show-ref --verify refs/remotes/origin/<value>`), so the two consumers cannot disagree about whether a configured value is usable. Implement the predicate once and call it from both; a second, divergent resolvability rule is the defect this note exists to prevent.

---

## §B Gaps and unverified items

Stated explicitly rather than asserted (AGENTS.md §1).

- **G1 — The reflog observation is inherited, not re-measured.** §A.2's `branch: Created from origin/develop` line is quoted from the orchestrator's dispatch. `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop` WAS re-measured in this tree at HEAD `48eb945df`.
- **G2 — `EnterWorktree`'s read of `origin/HEAD` is inferred from behaviour, not from source.** Claude Code is not in this repository; the evidence is the observed base of this worktree plus the documented `fresh` semantics at `worktree-integration.md:171-173`. If a future runtime changes that behaviour, consumer 1 silently stops mattering — the doctor item (REQ-WBR-012) is what would surface it.
- **G3 — `ModeProfile` has no `develop_branch` field.** `.moai/config/sections/git-strategy.yaml:15-17` sets `develop_branch` / `release_branch_prefix` / `rc_version_format` under `manual`, but `ModeProfile` (`internal/config/types.go:106-132`) declares none of them. A typed load-and-write round trip through `SetSection("git_strategy")` (`internal/settings/sectionapply.go:98`) would therefore drop those three keys. NOT verified end to end and NOT in scope — but it is adjacent to the write path this SPEC uses, and run-phase should confirm the round trip before shipping the web control.
- **G4 — `moai doctor`'s check registry was read, not run.** The group listing at `internal/cli/doctor.go:200-226` was read; no `moai doctor` invocation was performed in this worktree.
- **G5 — No template mirror of `.moai/config/sections/git-strategy.yaml` exists as a plain file.** The template ships `git-strategy.yaml.tmpl` (`internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:1-9`) with Go-template placeholders. The new key must be added there, not to a non-existent plain mirror.
- **G6 — The rendered attribute order of a `TypeText` control was NOT measured.** `internal/web/dead_config_guard_test.go` asserts only on `name="<field>"` (`:48`, `:96`), so the presence pattern is measured, but whether the renderer emits `type="text" name=…` or `name=… type="text"` was not. AC-WBR-011's render half is written with a two-branch condition for that reason; run-phase MUST measure the emitted order and collapse the assertion to the single true form.
- **G7 — `git show-ref --verify` on a missing ref exits 128, not 1, in this tree.** Measured in worktree `.claude/worktrees/t313` at HEAD `48eb945df`: `git show-ref --verify refs/remotes/origin/develop` → rc 0; `git show-ref --verify refs/remotes/origin/nonexistent-xyz` → rc 128. The REQ-WBR-009 predicate must therefore test `rc == 0`, never `rc == 1` — a `rc == 1` test would classify every missing ref as an execution error rather than as unresolvable.

---

## §C Milestones

Ordered so that the decision-bearing work lands before the mechanical work.

### M1 — Schema and neutral default (data model)

- Add `WorktreeBaseBranch string \`yaml:"worktree_base_branch"\`` to `GitStrategyConfig` (`internal/config/types.go:143-167`), documented as repository-wide and empty-by-default.
- Add the key with an empty **value** to `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` beside `github_username` (`:5-9`), with a comment naming its effect and its neutral default. The comment MAY name `main` and `develop` as examples (house style, REQ-WBR-014); AC-WBR-002's grep strips comments and binds the value side of the colon only, so `worktree_base_branch: ""  # e.g. main, develop; empty = no action` passes. What MUST NOT happen is a shipped **value** naming a branch (REQ-WBR-003).
- `make build`; mirror to `.moai/config/sections/git-strategy.yaml`. The **local** file MAY carry `develop`; the **template** MUST NOT (REQ-WBR-003).
- Guard: a template-neutrality assertion that the shipped `.tmpl` does not carry `develop` for this key.

### M2 — Consumer 1 (SessionStart alignment)

- New helper in `internal/hook/` implementing the five-way behaviour of REQ-WBR-005..009: read config → no-op silently on empty → read `refs/remotes/origin/HEAD` → no-op silently on equal → **verify resolvability** (`git show-ref --verify refs/remotes/origin/<value>`) → on unresolvable, one diagnostic line and NO write → otherwise `git remote set-head origin <value>` + one stderr line → fail-open on any error.
- The resolvability predicate is a **single exported helper** consumed by both M2 and M3 (plan §A D4 fail-back). Use the plumbing form; `git branch --list` / `git branch -vv` are refused at the tool layer by BranchGuard's `\bgit\s+branch\b` over-match (CLAUDE.local.md §4.1.4) and would be discovered as a blocker mid-run.
- **Gate the whole helper on the primary checkout first (REQ-WBR-004, second clause)**: before reading the configured value, compare `git rev-parse --git-dir` against `git rev-parse --git-common-dir` — equal means primary checkout and the helper proceeds, different means linked worktree and the helper returns immediately having read nothing, written nothing, and emitted nothing. `inGitWorktreeReal` (`internal/cli/session_worktree.go:234-241`) is the existing implementation of that discriminant; reuse its shape rather than inventing a second test. This gate is what removes the multi-lane write contention plan §E R1 analyses.
- Register as a fifth `g.Go(...)` task in `Handle`'s errgroup (`internal/hook/session_start.go:120-175`), writing to its own local map.
- Tests: the five behaviours, plus a fail-open test asserting `Handle` returns a nil error when the git command fails, plus the two firing-point tests AC-WBR-016 demands — read-seam call count 1 from the primary checkout (half 1) and 0 from a linked worktree (half 2). Both assert on a seam call count, so the read path is a seam, not a direct `exec.Command`. The seam these two tests count is the **alignment-entry (configured-value) read**, per §A D3.2 — NOT the `origin/HEAD` read, which must stay a separate seam so half 1 holds for an empty value too.

### M3 — Consumer 2 (`git worktree add` base operand)

- Widen the `sessionWorktreeGitWorktreeAdd` seam (`internal/cli/session_worktree.go:51-53`) to carry a base string; `gitWorktreeAddReal` (`:217-219`) appends it as the final operand when non-empty.
- `materializeSessionWorktree` (`:181`) resolves the configured value and passes it; unresolvable or empty → no operand. Resolvability is decided by **calling M2's exported helper**, never by a second rule — REQ-WBR-011 now carries this at the requirement layer, so a divergent `git rev-parse --verify` or `git branch --list` implementation is a requirement violation, not merely a style deviation. Expose the helper through a seam so AC-WBR-008's third assertion (call count 1 with the configured value) is testable.
- Tests: a fake seam asserting the operand for set / empty / unresolvable, plus a real-git test asserting the created tree's reflog names the configured base.

### M4 — Doctor diagnostic

- `checkWorktreeBaseBranch(projectRoot, verbose) DiagnosticCheck` modelled on `checkWorktreeState` (`internal/cli/doctor.go:876-891`), registered in the same group (`:220`).
- Four states (REQ-WBR-012): `uikit.CheckOK` on empty; `CheckOK` on match; non-OK on mismatch with `Message` naming the repair command; a **separate** non-OK on unresolvable with `Message` naming the offending value and telling the user to correct the setting. Collapsing the last two into one message is a defect — the repairs differ.

### M5 — Web surface and the anti-dead-key guard

- `FieldDef` in `gitStrategyFields()` (`internal/settings/schema_sections.go:160-177`) with `Type: settings.TypeText` per D2.1, plus `I18nKey`/`Description` entries in `internal/web/assets/i18n.js`. The description names `main` and `develop` as the common values (REQ-WBR-014) — the names live in prose, never in an option set.
- `applyGitStrategyKey` case for `worktree_base_branch` (`internal/settings/sectionapply.go:170-175`, alongside the existing `mode` case).
- Regression guard modelled on `internal/web/dead_config_guard_test.go`: present in `settings.AllFields()`, present in the rendered HTML, and reaching a consumer.

### M6 — Documentation and template parity

- Update the worktree rule's base-branch section (`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md:169-192`) to state that moai's own creation path now honours the configured base, keeping the `baseRef` "fresh"/"head" statement intact.
- `make build`; mirror; verify no hook `.sh` was touched (and if one was, its `.sh.tmpl` twin carries the identical edit — REQ-WBR-016).

---

## §D File-by-file write list

| File | Milestone | Change |
|---|---|---|
| `internal/config/types.go` | M1 | `WorktreeBaseBranch` field on `GitStrategyConfig` (:143-167) |
| `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | M1 | New root key, empty value, neutral comment (:5-9) |
| `.moai/config/sections/git-strategy.yaml` | M1 | Local mirror (MAY carry `develop`) |
| `internal/hook/session_start.go` | M2 | Fifth errgroup task (:120-175) |
| `internal/hook/worktree_base_branch.go` (new) | M2 | Alignment helper + the shared resolvability predicate (REQ-WBR-009), exported for M3 |
| `internal/hook/worktree_base_branch_test.go` (new) | M2 | Five behaviours (empty / match / resolvable-mismatch / unresolvable / git-error) + fail-open |
| `internal/cli/session_worktree.go` | M3 | Seam signature (:51-53), `gitWorktreeAddReal` (:217-219), `materializeSessionWorktree` (:181-196) |
| `internal/cli/session_worktree_test.go` | M3 | Operand assertions + reflog test |
| `internal/cli/doctor.go` | M4 | `checkWorktreeBaseBranch` + registration (:220) |
| `internal/cli/doctor_test.go` | M4 | Match / mismatch / empty |
| `internal/settings/schema_sections.go` | M5 | `FieldDef` in `gitStrategyFields()` (:160-177) |
| `internal/settings/sectionapply.go` | M5 | `applyGitStrategyKey` case (:170-175) |
| `internal/web/assets/i18n.js` | M5 | Label + description keys |
| `internal/web/dead_config_guard_test.go` (or a sibling) | M5 | Live-key regression guard: present in `settings.AllFields()`, present in rendered HTML, reaches a consumer — the three-part conjunction of REQ-WBR-015, verified by mutation per AC-WBR-012 |
| `internal/settings/sectionapply_test.go` (or a `gitstrategy_roundtrip_test.go` sibling) | M5 | Typed round-trip test for REQ-WBR-013's preservation clause / AC-WBR-014: `manual.develop_branch`, `manual.release_branch_prefix`, `manual.rc_version_format` survive a `worktree_base_branch` write |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | M6 | Base-branch section (:169-192) |
| `.claude/rules/moai/workflow/worktree-integration.md` | M6 | Local mirror |

---

## §E Risks

### R1 — Git-metadata mutation blast radius (High)

`git remote set-head origin <branch>` changes `refs/remotes/origin/HEAD`, which is repository-global: every worktree of this repository, every concurrent session, and every subsequent `EnterWorktree` reads it. A session that starts while another lane is mid-creation changes that lane's base underneath it.

**The premise correction (plan-audit iter-1 D4).** An earlier version of this risk closed with *"that is a misconfiguration, not a race — one repository, one key"*. That sentence was **false**, and it was load-bearing, because it converted a live concurrency hazard into an accepted misconfiguration. Measured in this worktree at HEAD `48eb945df`:

```bash
git ls-files --error-unmatch .moai/config/sections/git-strategy.yaml ; echo "rc=$?"
# .moai/config/sections/git-strategy.yaml
# rc=0   → the file is TRACKED
```

A tracked file has one working-tree copy **per worktree**, and its content follows that worktree's branch. `refs/remotes/origin/HEAD` has one copy **per repository**. So the counts are inverted from what that sentence assumed: there are as many configured values as there are active card worktrees — eight concurrent lanes at audit time — all writing one shared handle. Two lanes on branches carrying different values are **not** misconfigured; each is correctly configured for its own branch. And write-only-on-difference does not confine that to a single transition: lane A's write creates precisely the difference lane B's next session start observes and reverses, so divergence is continuously re-created rather than resolved. The same re-creation happens in a single-lane repository whenever an external actor moves the handle (`git remote set-head -a`, a manual reset, a fresh clone), so "one write per configuration change" was an under-statement even there.

**The real steady-state invariant.** Silence holds only while every working tree that runs the alignment step carries the same configured value. That is not a property this SPEC can assume; it is one it must arrange.

**The narrowing that arranges it — REQ-WBR-004's second clause.** Consumer 1 fires **only from the primary checkout**, discriminated by `git rev-parse --git-dir` vs `--git-common-dir` (the same test `inGitWorktreeReal` already performs at `internal/cli/session_worktree.go:234-241`). Inside a linked worktree the alignment step performs no read, no write, and emits nothing (AC-WBR-016 half 2). The handle is repository-global, so one writer suffices and the SPEC loses nothing it claimed: the primary checkout's session start still aligns the handle every lane's `EnterWorktree` subsequently reads. Multi-lane write contention is removed outright rather than mitigated. **Consumer 2 is unaffected** — `moai cc -w` honours the configured base from any working tree (REQ-WBR-010 / REQ-WBR-011), because it passes an operand to its own `git worktree add` rather than mutating shared metadata.

Remaining mitigations, all still in force: the write-only-on-difference rule (REQ-WBR-006) keeps steady state write-free once the handle is aligned; the one-line notice (REQ-WBR-007) makes every write attributable; the resolvability check (REQ-WBR-009) skips the write entirely rather than pointing `refs/remotes/origin/HEAD` at a ref that does not exist — strictly worse than the defect this SPEC repairs, since every subsequent `EnterWorktree` would read a broken handle; the doctor item (REQ-WBR-012) lets any lane read the current state without writing.

**Residual risk, stated rather than dismissed.** Two writers remain possible in principle: a second *primary checkout* of the same repository (a separate clone has its own `refs/remotes/origin/HEAD`, so this is not the hazard it appears to be), and an external actor moving the handle by hand between session starts. Both produce one attributable notice line on the next primary-checkout session start, not a fight, because only one automatic writer exists. A lane that observes an unexpected base reads the doctor item; it does not race to correct it.

### R2 — `moai update` reverts the local key (Medium, inherited)

`.moai/config` is inside the `CleanMoaiManagedPaths` wipe root (CLAUDE.local.md §2.3), so `moai update` restores `git-strategy.yaml` to template defaults — which, by REQ-WBR-003, means the empty neutral value. The failure is quiet in the direction of *doing nothing*, not in the direction of a wrong branch. This is the same hazard the `git_strategy.manual.workflow: gitflow` key already carries, and CLAUDE.local.md §2.3 already prescribes the re-application step. This SPEC inherits the hazard and does not fix it.

### R3 — t316 adjacency (Medium)

Card t316 owns both copies of `tab_schema.json` and its worktree is active. This SPEC touches neither file (spec.md §C). The adjacency is real in one direction: if t316 adds a `git_strategy` interview tab, a later card may want `worktree_base_branch` on it. That is a follow-up note, not a work item here.

### R4 — Typed round-trip may drop unmodelled keys (Medium, G3)

Saving `git_strategy` through the typed path (`internal/settings/sectionapply.go:98`, `:129-135`) writes back a `GitStrategyConfig`, which has no fields for `manual.develop_branch` / `release_branch_prefix` / `rc_version_format`. Editing the new web control could therefore silently drop three keys this repository depends on. Run-phase MUST verify the round trip before M5 lands, and escalate rather than absorb if the drop is real. As of iter-2 this is no longer only a risk: preservation is required by REQ-WBR-013's preservation clause and verified by AC-WBR-014 (promoted from SHOULD to MUST), and plan §D names the test file. The risk entry remains because the underlying `ModeProfile` gap (§B G3) is still unrepaired — this SPEC requires that its own write path not expose it, not that the gap be closed.

### R5 — The doctor item reports a state it cannot repair (Low)

A mismatch reported after worktrees already exist does not fix those trees (spec.md §C). The `Message` must say what the user should do (re-run a session, or reset an empty tree per `worktree-integration.md:186-190`) rather than implying the tree is now correct.

---

## §F Cross-references

- `.claude/rules/moai/workflow/worktree-integration.md:169-192` — `worktree.baseRef` semantics and the card-worktree base-branch guidance
- `CLAUDE.local.md` §2 (Template-First), §2.3 (`moai update` wipe root and `.sh`/`.sh.tmpl` pairs), §4.1.1 (card worktrees cut from develop), §15 (template neutrality)
- `.claude/rules/local/gitflow-lane-protocol.md` §1 rule 1
- `internal/web/dead_config_guard_test.go` — the guard pattern REQ-WBR-015 models
