# SPEC-CONFIG-TIER-RESOLVE-001 — Implementation Plan

## §A Context

- **Baseline**: HEAD `865cd8aa2` = `origin/main` (ff-synced, clean). Plan-phase executes in the main checkout (spec-workflow.md Step 1 HARD — no worktree).
- **Parent lineage**: This SPEC is slice (a) of the 3-way split of `SPEC-CONFIG-TIER-PERSIST-001`, recorded in the parent's §L Split Branches. The split carving preserves the parent's exact tier-resolution semantics; see `spec.md` § Split Lineage for the REQ-CTP → REQ-CTR mapping.
- **Code under change**:
  - `internal/config/merge.go` — `MergeAll` value-selection path (`:137` walk, `:149-152` zero-skip, `:155-162` strict-mode rejection), plus the `isZero` helper disposition (`:200`).
  - `internal/config/source.go` — `iota` block (`:26-28`) and `AllSources()` literal slice (`:103-114`).
  - `internal/config/source_test.go` — `TestSourceOrdering` `expectedOrder` literal (`:104-119`); `TestAllSources` (`:89-102`) needs no edit.
  - `internal/permission/resolver.go` — `tiers` literal slice (`:225-234`).
  - `internal/config/loader.go` — local-tier read (currently zero matches on `config/local\|localDir\|SrcLocal`).
- **Out of scope** (full list in `spec.md` §C): atomic/mode-preserving writes (slice b, closed); malformed-section contract, gitignore merge, fallback visibility (slice c, resident in parent); the SrcUser/SrcProject relative order; template content.

## §B Known Issues

- **B1 — `Priority()` has zero non-test consumers.** Measured on the baseline tree: `grep -rn '\.Priority()' --include='*.go' internal/ | grep -v '_test.go'` returns no output. Reordering the `iota` block alone therefore changes no production resolution while breaking `TestAllSources`'s `AllSources()[i].Priority() == i` invariant. The three-site reorder in M2 is load-bearing, not stylistic.
- **B2 — `resolver.go` does NOT iterate the `Source` enum.** It carries its own literal `tiers` slice at `:225-234`. The reorder reaches permission resolution only because that slice is named as one of the three M2 edit sites. AC-CTR-007 is load-bearing: it fails if M2 reorders `source.go` but leaves `resolver.go` alone.
- **B3 — `TestSourceOrdering` already exists and is green.** Its current `expectedOrder` literal is `[SrcPolicy, SrcUser, SrcProject, SrcLocal, SrcPlugin, SrcSkill, SrcSession, SrcBuiltin]`. M2's obligation is to UPDATE that literal in the same commit as the three-site reorder, NOT to author a duplicate guard and NOT to replace the literal with an enum-derived comparison (the latter is tautological and catches nothing).
- **B4 — The `isZero` disposition is a caller-count decision, not a blanket delete/keep.** After M1 removes the zero-skip at `merge.go:149-152`, `isZero` may retain other callers or may become an orphan. REQ-CTR-005a/b states both outcomes non-contradictorily; AC-CTR-006 enforces the decision is consistent with the caller count, not that either outcome is chosen unconditionally. Do NOT blindly delete `isZero` on M1 commit.
- **B5 — R1 is High/High and the mitigation is the M0 enumeration deliverable.** A previously-ignored false that disables a depended-on feature is the single largest risk in this SPEC. The M0 enumeration is a milestone deliverable, not a footnote — it MUST be reviewed before M1 lands.
- **B6 — M1 must land before M3.** Wiring the local tier (M3) while falsey values are still skipped (M1 not yet in force) produces an opt-in that cannot opt out. REQ-CTR-012 makes the milestone ordering normative; AC-CTR-010 (local tier can disable) cannot pass unless M1 landed first.
- **B7 — Cross-SPEC policy conflict scan clean.** `grep -r "Retired\|superseded\|SPEC-V3R2-RT-005" internal/config/ internal/permission/` — `SPEC-V3R2-RT-005` is the 8-tier system this SPEC corrects (cited at `source.go:5` `@MX:ANCHOR` with `fan_in=71`); it is the origin, not a conflict. No retired/superseded SPEC conflicts with the tier-resolution scope.
- **B8 — Subagent boundary: n/a.** This SPEC touches `internal/config/` and `internal/permission/`, neither of which is subagent-domain code; no `AskUserQuestion` grep applies.

## §C Pre-flight (run before any code change)

```bash
# 1. Branch + baseline
git branch --show-current                                   # expect: main (or a feature branch — repo-local PR policy applies)
git rev-parse HEAD                                          # expect: 865cd8aa2 (baseline)

# 2. Cross-platform build
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Confirm the three ordering sites at their current line numbers
grep -n 'SrcPolicy\|SrcUser\|SrcProject\|SrcLocal' internal/config/source.go | head -20
grep -n 'SrcPolicy\|SrcUser\|SrcProject\|SrcLocal' internal/permission/resolver.go | head -20

# 5. Confirm the local-tier-read gap (zero matches today)
grep -n "config/local\|localDir\|SrcLocal" internal/config/loader.go || echo "zero matches (expected — M3 adds the read)"

# 6. Confirm `isZero`'s current caller count
grep -n 'isZero(' internal/config/merge.go
```

## §D Constraints

- **PRESERVE** — every other pair of tier orderings (REQ-CTR-007). `SrcPolicy` MUST remain the numerically highest-priority tier. The SrcUser/SrcProject relative order is explicitly out of scope and MUST NOT be touched.
- **PRESERVE** — the existing `TestSourceOrdering` guard structure (REQ-CTR-008). Update its `expectedOrder` literal in the same commit as the three-site reorder; do NOT delete, weaken, or replace it with an enum-derived comparison.
- **PRESERVE** — `TestAllSources`'s `AllSources()[i].Priority() == i` invariant. Reordering the `iota` block and `AllSources()` TOGETHER restores this invariant; the test needs no edit. Reverting only one of the two is an incomplete change and is not a valid falsification.
- **Forbidden commands** — `--no-verify` (a warn-only pre-commit hook is normal); `--amend` after push; force-push to `main`.
- **Required commands** — Conventional Commits format (`feat(SPEC-CONFIG-TIER-RESOLVE-001): M{N} <subject>`); commit messages in English per `language.yaml`.
- **Repo-local PR policy** — `main` is protected (`enforce_admins: true`); all tiers use Route B (PR), self-merge allowed, CI must pass. The orchestrator handles PR strategy after plan-audit; this plan-phase does NOT commit or push.
- **Scope discipline** — touch only `internal/config/{merge,source,source_test,loader,loader_test}.go`, `internal/permission/resolver.go` (+ its tests), and the M0 enumeration deliverable. Do NOT touch `internal/template/templates/**`, the unrelated `.moai/config/sections/llm.yaml` modification in the working tree, or any sibling SPEC directory.

## §E Milestones

Order by decision-reversibility (highest-change-likelihood first), per the SPEC builder plan-ordering rule.

### M0 — falsey-value blast-radius enumeration (Milestone Deliverable, NOT a footnote)

**Deliverable**: a table at `.moai/specs/SPEC-CONFIG-TIER-RESOLVE-001/falsey-key-inventory.md` tabulating every key in `internal/config/defaults.go` and every shipped `.moai/config/sections/*.yaml` whose effective value changes once falsey values win.

Per row:
- file (`defaults.go` / `sections/<name>.yaml`)
- key path (e.g. `Workflow.BranchGuard.Enabled`)
- current effective value (the value that wins today because falsey is skipped)
- value that will win after M1 (the explicit falsey value)
- direction (disable / blank / zero / empty-collection)
- reviewed (Y/N — a human review column; the table is reviewed BEFORE M1 lands, per R1/R2 mitigation)

**Why M0 leads** — R1 (High/High) and R2 (Medium/High) are both mitigated by this table or not at all. A previously-ignored false that disables a depended-on feature is the single largest risk in this SPEC; the enumeration MUST exist and be reviewed before the merge-semantics change ships.

**Work item**: produce the table. No code change. No AC beyond the table's existence (AC-CTR-004).

### M1 — falsey-wins semantics (REQ-CTR-001..005b) — ONE revertable commit

The merge-semantics change, shipped alone so it can be reverted without unwinding M2/M3 (NFR-CTR-006).

- **REQ-CTR-001** — replace the zero-skip at `merge.go:149-152` with a presence check: select the first tier in priority order in which the key is present, regardless of whether that value is falsey.
- **REQ-CTR-002** — record the explicit falsey value as the winning value with that tier as `Provenance.Source`. `Get("explicitly.false")` MUST return `ok=true, value=false`.
- **REQ-CTR-003** — decouple `PolicyOverrideRejected` from `winningValue != nil`. The rejection block at `merge.go:155-162` MUST fire for a falsey policy value under `strict_mode` on the same terms as a non-falsey policy value.
- **REQ-CTR-004** — absent keys stay omitted; absence (`ok=false`) and falsey-presence (`ok=true, value=false`) MUST remain distinguishable in the merged result.
- **REQ-CTR-005a/b** — after removing the zero-skip, count `isZero`'s remaining callers in `internal/config/merge.go`. Retain the helper if at least one caller whose semantics genuinely require zero-detection remains; remove it if it is an orphan. Do NOT blindly delete.

**Tests** (TDD RED-GREEN-REFACTOR — `quality.yaml` `development_mode: tdd`):
- `TestMergeAll_ExplicitFalseWins` (AC-CTR-001)
- `TestMergeAll_AbsentKeyOmittedFalseyKeyPresent` (AC-CTR-002)
- `TestMergeAll_StrictModeRejectsOverrideOfFalseyPolicy` (AC-CTR-003)
- `isZero` disposition consistency shell check (AC-CTR-006)

**Why this ordering** — M1 MUST land before M3 (REQ-CTR-012). Wiring the local tier while falsey values are still skipped produces an opt-in that cannot opt out.

### M2 — 3-site SrcLocal reorder (REQ-CTR-006..009) — ONE commit, all three sites or it is incomplete

The tier-ordering change, shipped as a single commit that touches the enum and its guards only (NFR-CTR-005) so it can be reverted independently of M1.

- **REQ-CTR-006** — reorder `SrcLocal` above `SrcProject` at ALL THREE sites in ONE commit:
  1. the `iota` block in `internal/config/source.go` (so `SrcLocal.Priority() < SrcProject.Priority()`),
  2. `AllSources()` literal slice (`source.go:103-114`, which `merge.go:137` walks),
  3. `tiers` literal slice in `internal/permission/resolver.go:225-234`.
- **REQ-CTR-007** — every other pair of tier orderings stays unchanged; `SrcPolicy` remains numerically highest.
- **REQ-CTR-008** — UPDATE the `TestSourceOrdering` `expectedOrder` literal (`source_test.go:104-119`) to the new order IN THE SAME COMMIT. Retain the literal; do NOT replace it with an enum-derived comparison. The new literal reads `[SrcPolicy, SrcUser, SrcLocal, SrcProject, SrcPlugin, SrcSkill, SrcSession, SrcBuiltin]`.
- **REQ-CTR-009** — show `resolver.go` resolves correctly under the new ordering. It does NOT iterate the `Source` enum; the reorder reaches it only because its `tiers` slice is one of the three named sites. The assertion is load-bearing.

**Tests**:
- `TestSourceOrdering_LocalOutranksProject` (AC-CTR-005) — asserts BOTH (1) `MergeAll` merge outcome observes the `AllSources()` literal [load-bearing], AND (2) `SrcLocal.Priority() < SrcProject.Priority()` [under-reach catcher].
- `TestSourceOrdering` + `TestAllSources` green with updated literal (AC-CTR-006-guard).
- `TestPermissionResolver_*` plus new `TestPermissionResolver_LocalOutranksProject` (AC-CTR-007) — load-bearing: fails if step 2 reorders `source.go` but leaves `resolver.go` alone.
- No-ordinal-cast regression guard (AC-CTR-008) — already green, retained.

**Why a single commit** — reverting fewer than all three sites is an incomplete change (REQ-CTR-006). The `iota` block alone leaves both literal slices resolving in the old order while breaking `TestAllSources`.

### M3 — local-tier reachability (REQ-CTR-010..012) — applied AFTER M1

- **REQ-CTR-010** — typed `Loader` reads `.moai/config/local/*.yaml` and applies it above `.moai/config/sections/*.yaml`.
- **REQ-CTR-011** — absent `local/` directory → no warning, no error, identical behaviour to today.
- **REQ-CTR-012** — local tier applied AFTER REQ-CTR-001 is in force, so it can both enable AND disable a setting.

**Tests**:
- `TestLoader_LocalTierOverridesSections` (AC-CTR-009) — `sections/workflow.yaml=false`, `local/workflow.yaml=true` → loaded value is `true`.
- `TestLoader_LocalTierCanDisable` (AC-CTR-010) — `sections/workflow.yaml=true`, `local/workflow.yaml=false` → loaded value is `false`. Cannot pass unless M1 landed first (REQ-CTR-012 ordering).
- `TestLoader_NoLocalDirIsSilent` (AC-CTR-011) — absent `local/` directory → identical `*Config`, no warning.

### Regression guards (REQ-CTR-013/014) — AC-only, no milestone work item

- **REQ-CTR-013** — `.moai/config/local/` remains in `.gitignore` (AC-CTR-013). Already satisfied at `b9fc75016`.
- **REQ-CTR-014** — `CLAUDE.local.md` §22.9 carries no `BLOCKER (gitignore)` note (AC-CTR-014). Already satisfied.

No milestone work item follows; both are retained as regression guards only.

## §F Anti-Patterns

- **AP-1 — `iota`-only reorder.** Changing only the `iota` block changes no production resolution (Priority has zero non-test consumers) while breaking `TestAllSources`. The three-site reorder in M2 is load-bearing.
- **AP-2 — Deleting `TestSourceOrdering`'s literal and replacing it with `AllSources()` itself.** That is tautological — the test would compare `AllSources()` to `AllSources()` and catch nothing. REQ-CTR-008 requires the literal be UPDATED, not replaced.
- **AP-3 — Wiring the local tier before M1 lands.** Produces an opt-in that cannot opt out (REQ-CTR-012 violation). M3 MUST follow M1.
- **AP-4 — Blindly deleting `isZero` on M1 commit.** The disposition is a caller-count decision (REQ-CTR-005a/b). Count remaining callers first; retain or remove accordingly.
- **AP-5 — Skipping the M0 enumeration.** R1 (High/High) is mitigated by the enumeration table or not at all. The table is a milestone deliverable, not a footnote.
- **AP-6 — Including unrelated changes in the M1 commit.** NFR-CTR-006 requires the merge-semantics change ship alone so it can be reverted independently of M2/M3. Bundling the reorder or the local-tier wiring into M1 defeats the revertability property.

## §G Cross-References

- `spec.md` §D — REQ-CTR-001..014 (the requirements this plan implements).
- `acceptance.md` §B — AC-CTR-001..014 (the acceptance criteria this plan's milestones satisfy).
- Parent `SPEC-CONFIG-TIER-PERSIST-001` §L — split-branch mapping (this child = slice (a)).
- Sibling `SPEC-CONFIG-ATOMIC-WRITE-001` — slice (b), closed. Owns the atomic-write helper; this child's tier resolution is orthogonal to write mechanics.
- `SPEC-V3R2-RT-005` — the 8-tier system this SPEC corrects; `source.go:5` `@MX:ANCHOR` records `fan_in=71`.
- `CLAUDE.local.md` §22.9 — maintainer opt-in doctrine; its former `BLOCKER (gitignore)` note is already cleared (REQ-CTR-014).

## §H Self-Verification (manager-develop E1-E8, attribution per VCI §3)

When manager-develop reports completion, it MUST include:
- **E1** — AC Binary PASS/FAIL matrix (AC-CTR-001..014), each row naming the command + verbatim observed output.
- **E2** — Cross-platform build (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`).
- **E3** — Coverage (`go test -cover ./internal/config/... ./internal/permission/...`).
- **E4** — Subagent-boundary grep (n/a for this SPEC's packages, but state it).
- **E5** — Lint status (NEW vs pre-existing baseline).
- **E6** — Branch HEAD + push state (per repo-local PR policy, Route B).
- **E7** — Blocker report (if any; NEVER call AskUserQuestion).
- **E8** — RED failure output (TDD verbatim pre-GREEN evidence).

Each E-item names (a) the command, (b) the observed output, (c) the baseline-attribution (this run, this tree, HEAD SHA). Per VCI §3, summarized evidence like "all tests passed" is NOT acceptable.
