# Acceptance — SPEC-CONFIG-DEAD-SWEEP-001

> **v0.2.0 amendment.** Every criterion below carries a **measured baseline** taken in the
> `chore/config-dead-sweep-resume` worktree at `ddfe2253f` on 2026-08-14. A criterion whose
> baseline already equals its target would be vacuous — it would have passed between commits
> `4c88bbce9` and `7171880a9` while proving nothing (spec.md §B.0). Where a baseline is
> already at target, that is stated explicitly rather than dressed as a check.

## §D.0 Retired criteria (v0.2.0)

| AC ID | Status | Reason |
|-------|--------|--------|
| AC-CDS-001 [RETIRED] | RETIRED | Asserted `cache.yaml` removal. REQ-CDS-003 is retired — `SPEC-WEBCONF-SIMPLIFY-001` REQ-WC-003 retains the template `cache.yaml` (`internal/settings/sectionroute.go:88-89`). |
| AC-CDS-002 [RETIRED] | RETIRED | Asserted `LoadCacheConfig` / `CacheConfig` removal. Same reason. |
| AC-CDS-006 [RETIRED] | RETIRED | Asserted `ValidSessionTTLs` survives an extraction that is no longer performed — `cache_config.go` is untouched, so there is nothing to assert. |

A superseding criterion, AC-CDS-014, replaces them: it asserts the cache surface is **unchanged**.

## §D. AC Matrix

| AC ID | REQ | Severity | Summary |
|-------|-----|----------|---------|
| AC-CDS-003 | REQ-CDS-004 | MUST | research.yaml removed from local + template |
| AC-CDS-004 | REQ-CDS-004 | MUST | Research loader, wrapper, type, field, default, registry row removed |
| AC-CDS-005 | REQ-CDS-005 | MUST | template state_dir key + StateConfig.StateDir + DefaultStateDir removed; hardcoded literal unchanged |
| AC-CDS-007 | REQ-CDS-007 | MUST | go build ./... && go test ./... exit 0 |
| AC-CDS-008 | REQ-CDS-007 | MUST | moai update still merges remaining sections (smoke test) |
| AC-CDS-009 | REQ-CDS-009 | MUST | TestAuditLoaderCompleteness green (no YAML_SECTION_NO_LOADER: research) |
| AC-CDS-010 | REQ-CDS-009 | MUST | TestAuditRegistry_AllRegisteredStructsExist green after the paired knownSections edit |
| AC-CDS-011 | REQ-CDS-010 | MUST | Starting state matches the half-reverted baseline (re-run precondition) |
| AC-CDS-012 | REQ-CDS-009 | MUST | No new allowlist/exception entry added for research |
| AC-CDS-013 | REQ-CDS-008 | MUST | template-neutrality-check CI green |
| AC-CDS-014 | REQ-CDS-003 | MUST | cache surface byte-unchanged (retirement is honoured, not silently reversed) |
| AC-CDS-015 | REQ-CDS-005b | MUST | Unrelated StateDir-substring symbols untouched |
| AC-CDS-016 | REQ-CDS-009 | MUST | TestStructYAMLSymmetry green, case count still 7 |
| AC-CDS-017 | REQ-CDS-004 | MUST | internal/settings research test fixtures preserved |
| AC-CDS-018 | REQ-CDS-006 | SHOULD | stale comments at loader.go and audit_registry.go corrected |

## §D.1 AC Definitions (Given-When-Then)

### AC-CDS-003 — research.yaml removed from both trees
**Given** the local project and the template source tree,
**When** `ls internal/template/templates/.moai/config/sections/research.yaml .moai/config/sections/research.yaml` runs,
**Then** both paths are absent (command exits non-zero with two "No such file" errors).

*Measured baseline:* **both files present, 680 bytes each.** This criterion fails today, so it is decidable.

### AC-CDS-004 — Research plumbing removed
**Given** the post-removal source tree,
**When**
```
grep -rn "ResearchConfig\|researchFileWrapper\|loadResearchSection\|\.Research\b" internal/ cmd/ pkg/ | grep -v "_test.go" | wc -l
```
runs,
**Then** the count is `0`;
**And when** `grep -n "research" internal/config/slice.go internal/config/audit_registry.go` runs,
**Then** zero matches;
**And when** `grep -n "research.yaml" internal/template/templates/.claude/rules/moai/core/settings-management.md` runs,
**Then** zero matches.

*Measured baseline:* the first count is **20** (spanning `slice.go:31`, `audit_registry.go:39`, `types.go:31/837/838/1361/1362/1363`, `defaults.go:349/362/363/364`, `loader.go:77/278/279/280/287/295`, `resolver.go:805`, and the template doc row at `settings-management.md:101`). Non-zero baseline → the criterion is decidable.

*Note:* the `.Research\b` pattern also matches the comment at `loader.go:295` ("parallel to loadResearchSection"), which must be reworded rather than left dangling — a comment referencing a deleted function is itself a stale-comment defect.

### AC-CDS-005 — state_dir removed, literal preserved
**Given** the post-removal source tree,
**When** `grep -c "state_dir" internal/template/templates/.moai/config/sections/state.yaml` runs,
**Then** the count is `0`;
**And when** `grep -n "StateDir" internal/config/types.go internal/config/defaults.go` runs,
**Then** zero matches (the field, its default-population, and the now-orphaned `DefaultStateDir` constant are all gone);
**And when** `grep -n "\.moai/state" internal/cli/state.go internal/worktree/state_guard.go` runs,
**Then** the hardcoded literal is still present in both files;
**And when** `grep -rn "\.State\.StateDir" internal/ cmd/ pkg/ | grep -v _test.go` runs,
**Then** zero matches.

*Measured baseline:* template `state_dir` count **1** (`state.yaml:5`); `grep -n "StateDir" types.go defaults.go` returns **3** lines (`types.go:682`, `defaults.go:150`, `defaults.go:742`); `.State.StateDir` non-test readers already **0** — that last sub-check is the *justification* for the removal, not a post-condition that can fail, and is recorded as such.

*Local-side note:* `.moai/config/sections/state.yaml` is **already** `state: {}` and needs no edit. Asserting its cleanliness would be vacuous; it is listed here as an unchanged precondition (AC-CDS-011).

### AC-CDS-007 — build + test green
**Given** the post-removal source tree,
**When** `go build ./...` runs, **Then** exit 0;
**And when** `go test ./...` runs, **Then** exit 0.

*Measured baseline:* the targeted config guards are green today —
`go test ./internal/config/ -run 'TestAuditLoaderCompleteness|TestStructYAMLSymmetry|TestAuditRegistry' -count=1` → `ok github.com/modu-ai/moai-adk/internal/config 0.550s`.
The full-suite baseline MUST be captured at run-phase start before any edit, because a pre-existing failure elsewhere would otherwise be misattributed to this sweep.

### AC-CDS-008 — moai update smoke test
**Given** a `/tmp/test-project` initialized from the rebuilt template,
**When** `moai update` runs, **Then** exit 0;
**And** the remaining section files merge without an error mentioning `research.yaml`;
**And** `cache.yaml` is still deployed (REQ-WC-003 retention observable end-to-end).

### AC-CDS-009 — loader-completeness guard green
**Given** the post-removal source tree,
**When** `go test ./internal/config/ -run TestAuditLoaderCompleteness -count=1` runs,
**Then** exit 0 and no `YAML_SECTION_NO_LOADER: research` line appears in the output.

*Measured baseline:* green today, because template `research.yaml` exists **and** `loadResearchSection` is registered (`slice.go:31`) so `loaded["research"]` is true. The guard turns **red** if the loader is removed while the template file remains — this is the coupling the criterion protects. To make it capable of failing rather than passing by accident, the run-phase MUST record the intermediate red (loader removed, file still present) before completing the pair, or explicitly state that the pair landed atomically in one edit.

### AC-CDS-010 — registry guard green after paired edit
**Given** the `research` row is removed from `YAMLToStructRegistry` (`internal/config/audit_registry.go:39`),
**When** `go test ./internal/config/ -run TestAuditRegistry -count=1` runs,
**Then** exit 0.

*Measured baseline:* green today. Removing the registry row **alone** fails with `registry missing section "research"`, because `knownSections` at `internal/config/audit_registry_test.go:56` still lists it. Both edits land together; the test-list edit is in scope.

### AC-CDS-011 — re-run precondition (distinguishes "never done" from "done then reverted")
**Given** the worktree before any edit,
**When** the following four observations are made,
**Then** all four hold — confirming the half-reverted starting state of spec.md §B.0:

| # | Observation | Expected |
|---|-------------|----------|
| 1 | `ls .moai/config/sections/cache.yaml` | absent (first run's deletion survived) |
| 2 | `grep -c state_dir .moai/config/sections/state.yaml` | `0` in the `state:` block; the file carries the removal comment |
| 3 | `ls internal/template/templates/.moai/config/sections/research.yaml` | present (revert restored it) |
| 4 | `grep -n state_dir internal/template/templates/.moai/config/sections/state.yaml` | present at line 5 |

**And** if any observation fails, the run aborts and reports the divergence rather than proceeding — the tree is not in the state this amendment measured.

*Why this AC exists:* a plain "research.yaml is absent" assertion would have passed in the window between `4c88bbce9` and `7171880a9` while the SPEC's work was, in fact, about to be undone. Pinning the asymmetric starting state is what makes the re-run's success attributable.

### AC-CDS-012 — no suppression substituted for removal
**Given** the post-removal source tree,
**When** `grep -n "research" internal/config/audit_loader_completeness_test.go internal/config/audit_registry.go` runs,
**Then** zero matches — `research` appears in neither `acknowledgedUnloadedSections`, `acknowledgedDedicatedLoaders`, nor `yamlAuditExceptions`.

*Measured baseline:* `research` is currently absent from all three allowlists (it is genuinely loaded). The criterion guards against the cheapest wrong fix: silencing a red guard by allowlisting instead of completing the removal.

### AC-CDS-013 — template-neutrality CI green
**Given** the PR is open against main,
**When** `template-neutrality-check.yaml` runs,
**Then** exit 0. Deleted template files trivially pass; the edited template `state.yaml` and `settings-management.md` must not gain a SPEC ID, REQ token, or commit SHA.

*Note:* the local `.moai/config/sections/state.yaml` comment DOES name this SPEC — that is permitted, because the local tree is outside the neutrality guard's path scope. The template `state.yaml` must NOT copy that comment.

### AC-CDS-014 — cache surface unchanged (retirement honoured)
**Given** the post-removal source tree,
**When** `git diff --name-only <base>..HEAD` is inspected,
**Then** none of the following appear:
`internal/config/cache_config.go`, `internal/config/cache_config_test.go`,
`internal/template/templates/.moai/config/sections/cache.yaml`,
`internal/config/testdata/cache-*/cache.yaml`,
`internal/template/cache_config_contract_test.go`;
**And when** `grep -n "cache" internal/config/audit_loader_completeness_test.go` runs,
**Then** the `acknowledgedDedicatedLoaders` entry is still present.

### AC-CDS-015 — unrelated StateDir symbols untouched
**Given** the post-removal source tree,
**When** `grep -rn "StateDir" internal/ cmd/ pkg/ | grep -v _test.go | wc -l` runs,
**Then** the count is exactly the baseline **minus 3** (`types.go:682`, `defaults.go:150`, `defaults.go:742`);
**And when** `grep -n "StateDir" internal/goal/state.go internal/worktree/state_guard.go internal/cli/chain.go` runs,
**Then** `goal.StateDir`, `StateDirRel`, and `ChainStateDir` are all still present and unmodified.

### AC-CDS-016 — symmetry guard green, scope unchanged
**Given** the post-removal source tree,
**When** `go test ./internal/config/ -run TestStructYAMLSymmetry -count=1` runs, **Then** exit 0;
**And when** `grep -c "structType:" internal/config/audit_struct_yaml_symmetry_test.go` runs,
**Then** the count is still `7`.

*Measured baseline:* **7** cases (constitution, context, interview, design, statusline, git-convention, gate). Neither `StateConfig` nor `ResearchConfig` is among them, so this guard is expected to be **unaffected**. The criterion asserts that expectation rather than assuming it — a drive-by addition or removal of a symmetry case would signal the sweep went out of bounds.

### AC-CDS-017 — settings test fixtures preserved
**Given** the post-removal source tree,
**When** `ls internal/settings/testdata/sections/research.yaml` runs, **Then** the file is present;
**And when** `go test ./internal/settings/ -count=1` runs, **Then** exit 0
(`TestWriteSectionViaSeamRejectsResearchPreservesFile` and the `seedTypedFixtures` callers still pass — they read from `testdata/`, never from the template).

### AC-CDS-018 — stale comments corrected
**Given** the post-fix source tree,
**When** the `learning` comment in `internal/config/loader.go` is read,
**Then** it no longer says "Legacy sub-system" and instead describes `learning` as live (consumed at `internal/cli/hook.go`);
**And when** the observability row comment in `internal/config/audit_registry.go` is read,
**Then** it no longer says "no Go loader yet" and instead reads "partial direct-read" (`observability_master.go` reads `enabled` live).

*Note:* line numbers are deliberately omitted from this criterion. The v0.1.0 version pinned `loader.go:345` and `audit_registry.go:75`, and the surrounding code has since shifted; the run-phase locates the comments by content, then records the line it actually edited.

## §D.2 Severity Definitions

- **MUST** — blocks run-phase completion; failure aborts the run.
- **SHOULD** — fixed in the same PR if cheap; otherwise recorded as `@MX:DEBT` with a follow-up SPEC.

## §D.3 Indirect Verification

- `golangci-lint run` exit code is NOT a load-bearing AC (repo carries baseline warnings); recorded as a sanity check, not a gate.
- `moai spec lint` exits on repo-global state, so its exit code cannot attribute a finding to this SPEC. Filter to this SPEC ID before citing it.
- AC-CDS-005's `.State.StateDir` sub-check has a baseline of zero and therefore cannot fail. It is recorded as the removal's *justification*, not as evidence of the removal's success.

## §D.4 Closure Gate

All MUST ACs MUST be PASS. The run-phase §E.2 evidence block MUST cite the verbatim command + output for each AC (`verification-claim-integrity.md` §2), and MUST establish its own baseline from the current tree — citing commit `4c88bbce9`'s evidence is prohibited (REQ-CDS-010), because that run's result was partially reverted.

## §D.5 Forward-Looking Checks

- After merge, `moai spec audit` MUST classify this SPEC as `era_final: true` (V3R6, 3-phase close) — not grandfather.
- If a future broad recovery commit restores these surfaces a second time, the correct response is a ratchet (a guard asserting their absence), not a third re-run of this SPEC. That ratchet is out of scope here (spec.md §E).
