# Plan — SPEC-CONFIG-DEAD-SWEEP-001

> **v0.2.0.** Scope narrowed from three targets to two. The plan below is a **re-run** plan:
> the v0.1.0 work landed as `4c88bbce9` and was partially reverted by `7171880a9`
> (spec.md §B.0). Milestones are ordered by decision-reversibility — the scope decision that
> was overtaken leads, mechanical deletions follow.

## §A. Context

Two dead non-ralph config surfaces are removed in a behavior-preserving sweep: `research.yaml`
(local + template + Go plumbing) and the `state_dir` key (template + Go field + orphaned
constant). Evidence re-measured 2026-08-14 in the `chore/config-dead-sweep-resume` worktree at
`ddfe2253f`; the v0.1.0 line numbers had drifted and are corrected in spec.md §B.

The third v0.1.0 target, `cache.yaml`, is **dropped** — see M0.

The ralph half of the dead-config space is owned by `SPEC-RALPH-CONFIG-REDESIGN-001` (spec.md §H).

## §B. Known Issues

- **The tree is half-reverted, not pristine.** The local project already documents a removal
  (`state.yaml` carries a comment naming this SPEC) that the template contradicts. Any run that
  assumes a symmetric starting state will misread its own diff. M0 pins the starting state.
- **Two CI guards fail on a partial application.** `TestAuditLoaderCompleteness` breaks if the
  loader is removed while the template file remains; `TestAuditRegistry_AllRegisteredStructsExist`
  breaks if the registry row is removed while `knownSections` still lists `research`. Each pair
  lands atomically (M2, M3).
- **A comment will dangle.** `internal/config/loader.go:295` describes another loader as
  "parallel to loadResearchSection". Deleting `loadResearchSection` leaves that comment referring
  to nothing — reword it in the same commit, or the sweep creates the exact class of stale comment
  M5 exists to fix.
- **`DefaultStateDir` becomes orphaned.** Its only reader is the dead-field population at
  `defaults.go:742`. Once that line goes, the constant at `defaults.go:150` has zero readers and
  must go too — otherwise the sweep trades a dead YAML key for a dead Go constant.

## §C. Pre-flight (before M1)

- Run the AC-CDS-011 four-observation check. If the tree has moved since 2026-08-14, stop and
  re-measure before editing.
- Capture the **full** `go test ./...` baseline. A pre-existing failure elsewhere must be recorded
  now, or it will be misattributed to this sweep at M6.
- Confirm `make build` baseline is green.
- Re-grep for new consumers of `cfg.Research` / `cfg.State.StateDir` landed since 2026-08-14.

## §D. Constraints

- Template-First: `internal/template/templates/.moai/config/sections/` edited first, `make build`,
  then the local copy.
- `internal/config/cache_config.go` and the template `cache.yaml` are **frozen** (M0).
- `internal/settings/testdata/sections/research.yaml` and its consuming tests are **preserved**.
- No behavior change to `moai update`, `moai web`, or `findStateDir()`.
- Coupled edits land in a single commit each (M2, M3) — no intermediate red-guard state on a
  pushed commit.

## §E. Self-Verification (plan-phase)

- [x] SPEC ID regex PASS — see §E.1 evidence in `progress.md`.
- [x] Frontmatter 12 canonical fields present; `version` bumped to 0.2.0, `updated` to 2026-08-14.
- [x] Out of Scope section carries ≥1 `### Out of Scope — <topic>` H3 with `-` bullets (4 now).
- [x] Every live REQ uses a GEARS pattern; the retired REQ-CDS-003 is marked, not deleted.
- [x] Every live AC is Given-When-Then, binary-testable, and carries a measured baseline.
- [x] No implementation detail prescribed — symbol names appear as removal targets (evidence),
      not as designs for new code.

## §F. Milestones

### M0 — Scope decision: drop the cache.yaml target (LEAD — this is the reversal)

**Decision: `cache.yaml` leaves this SPEC's scope; REQ-CDS-003 is retired, not deleted.**

The v0.1.0 draft treated `cache.yaml` as dead on the strength of an `@MX:DEBT` marker saying
`LoadCacheConfig` had no caller. That reading was correct about the *loader* and wrong about the
*file*: `SPEC-WEBCONF-SIMPLIFY-001` M3 subsequently decided the baked template YAML is retained
for runtime consumption, stated at `internal/settings/sectionroute.go:88-89` — *"Their config keys
persist in the baked template YAML for runtime consumption (REQ-WC-003)."*

Two framings to avoid, both of which would produce a wrong plan:

- **Wrong:** "cache is a live seam write path, so keep it." It is not — `cache` is absent from the
  `sectionRoutes` map and resolves to `RouteExcluded` by zero value. Its tab and web write path
  are gone.
- **Wrong:** "the loader is dead, so the file is dead." REQ-WC-003 retains the file independently
  of whether a Go loader consumes it.

Concrete effect: no file under the cache surface is touched (spec.md §E), and AC-CDS-014 asserts
that as a positive check rather than trusting the absence of an edit.

### M1 — Template-First deletion of research.yaml

- Delete `internal/template/templates/.moai/config/sections/research.yaml`.
- Delete `.moai/config/sections/research.yaml`.
- Remove the doc row at
  `internal/template/templates/.claude/rules/moai/core/settings-management.md:101`
  (`| research.yaml | research | cfg.Research |`). The **local** copy of that doc is already clean —
  the first run's edit survived the revert — so no local doc edit is needed. Do not "restore
  symmetry" by re-adding it.
- Run `make build`.

### M2 — Remove the research loader chain (atomic with M1's file deletion)

Lands in the same commit as M1 so `TestAuditLoaderCompleteness` never sees a loader-less file:

- `internal/config/loader.go:279` — delete `loadResearchSection`; `loader.go:77` — delete the call.
- `internal/config/loader.go:295` — reword the "parallel to loadResearchSection" comment.
- `internal/config/slice.go:31` — delete the `"research"` registration.
- `internal/config/types.go` — delete `Config.Research` (`:31`), `ResearchConfig` (`:837-838`),
  `researchFileWrapper` (`:1361-1363`).
- `internal/config/resolver.go:805` — delete the `researchFileWrapper` branch.
- `internal/config/defaults.go` — delete `NewDefaultResearchConfig` (`:362-364`) and its call
  site (`:349`).

### M3 — Remove the audit-registry row (atomic with its test-list edit)

- `internal/config/audit_registry.go:39` — delete `"research": "ResearchConfig"`.
- `internal/config/audit_registry_test.go:56` — delete `"research"` from `knownSections`.

Both in one commit: removing either alone turns `TestAuditRegistry_AllRegisteredStructsExist` red.
Do **not** add `research` to `acknowledgedUnloadedSections` or `yamlAuditExceptions` (AC-CDS-012) —
an allowlist entry would preserve the dishonest surface this SPEC removes.

### M4 — Remove state_dir (option a, unchanged from v0.1.0)

**Decision: option (a) — remove the dead field, keep the hardcoded literal as SSOT.**

- **Option (a) (chosen):** smallest diff; `findStateDir()` and `state_guard.go` already hardcode
  `.moai/state/` and that is the de-facto SSOT. Cost: a user who set `state.state_dir` loses the
  knob — but it never worked, so nothing observable changes.
- **Option (b) (rejected):** wire `findStateDir()` / `state_guard.go` to read the field. Changes
  runtime path resolution, more code, and no evidence of demand.

Concrete changes:

- `internal/template/templates/.moai/config/sections/state.yaml:5` — remove `state_dir`.
  The `state:` key must remain (as `state: {}` or with a comment) so the file stays valid YAML
  with a top-level `state` key.
- `.moai/config/sections/state.yaml` — **UNCHANGED** (already `state: {}`).
- `internal/config/types.go:682` — remove the `StateDir` field.
- `internal/config/defaults.go:742` — remove `StateDir: DefaultStateDir,`.
- `internal/config/defaults.go:150` — remove the now-orphaned `DefaultStateDir` constant.
- `internal/cli/state.go:210-211`, `internal/worktree/state_guard.go:25` — **UNCHANGED**.
- `internal/goal/state.go:17`, `ChainStateDir`, `detectStateDir`, `navigatorDetectStateDir`,
  `convergenceStateDir`, `routingStateDir`, `ensureStateDir` — **UNCHANGED** (unrelated symbols).

### M5 — Stale comment fixes (fold-in)

- `internal/config/loader.go` — the `learning` comment: replace "Legacy sub-system (out-of-scope)"
  with a live description citing `internal/cli/hook.go`. `learning` is LIVE; only the comment is wrong.
- `internal/config/audit_registry.go` — the observability row: replace "no Go loader yet" with
  "partial direct-read" citing `observability_master.go`.

Locate both by content, not by the v0.1.0 line numbers (`:345` / `:75`) — the surrounding code
has shifted. Record the line actually edited.

### M6 — Verify

- `go build ./...` exits 0.
- `go test ./...` exits 0, compared against the M-pre-flight baseline.
- `go test ./internal/config/ -run 'TestAuditLoaderCompleteness|TestAuditRegistry|TestStructYAMLSymmetry' -count=1` green.
- `go test ./internal/settings/ -count=1` green (fixtures preserved).
- Re-run the AC greps; confirm the symmetry case count is still 7.
- `moai update` smoke test on `/tmp/test-project` — remaining sections merge, `cache.yaml` still deploys.

## §G. Anti-Patterns

- **AP-1: delete the template `cache.yaml`.** Violates REQ-WC-003. The v0.1.0 draft called for
  exactly this; it is now forbidden (M0).
- **AP-2: describe `cache` as a live seam write path.** It is `RouteExcluded`. The retention
  rationale is baked-template runtime consumption, not console editability.
- **AP-3: allowlist `research` to make a guard green.** Suppression is not removal (AC-CDS-012).
- **AP-4: remove the registry row without the `knownSections` entry** (or the loader without the
  template file). Each pair is atomic.
- **AP-5: cite `4c88bbce9`'s evidence as this run's evidence.** That run was partially reverted;
  its output does not describe the current tree (REQ-CDS-010).
- **AP-6: "restore symmetry" by re-adding the local doc row or the local `cache.yaml`.** The local
  tree's cleanliness is the *surviving* half of the first run, not a defect.
- **AP-7: remove `learning` because the stale comment says "legacy".** The comment is wrong;
  `learning` is LIVE. Only the comment is fixed.
- **AP-8: cross into `ralph.yaml`.** Collision with `SPEC-RALPH-CONFIG-REDESIGN-001`.
- **AP-9: skip `make build` after template edits.** The embedded FS goes stale and the deleted
  file still ships in the binary.
- **AP-10: delete `internal/settings/testdata/sections/research.yaml`.** It backs the seam-rejection
  test, which stays meaningful after the section is gone.

## §H. Cross-References

- `acceptance.md` — AC-CDS-003..018; AC-CDS-001/002/006 retired.
- `spec.md` §B.0 — the land-then-revert history; §C REQ-CDS-003 — the retirement record.
- `SPEC-WEBCONF-SIMPLIFY-001` — REQ-WC-003, the superseding decision (`internal/settings/sectionroute.go:88-89`).
- `SPEC-RALPH-CONFIG-REDESIGN-001/plan.md` — the ralph half; sequence or isolate.
- `CLAUDE.local.md` §2 (Template-First), §25 (Template Internal-Content Isolation).
