---
id: SPEC-UPDATE-YAML-PRESERVE-001
title: "moai update — preserve YAML comments, key order, scalar style, and user-added keys across config merge"
version: 0.1.0
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: high
phase: plan
module: cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, update, yaml, merge, config, preservation, regression"
issue_number: 1243
related_specs: [SPEC-CLI-TUX-V3-003, SPEC-V3R6-UPDATE-NOISE-001, SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002]
---

# SPEC-UPDATE-YAML-PRESERVE-001

## §A Problem / Motivation

GitHub issue **#1243** reported that `moai update` silently reverted local customizations in `.moai/config/sections/*.yaml`. Root-cause analysis isolated three independent defects. Two are already handled and are excluded here (see §B). The third — the subject of this SPEC — is unfixed.

### Defect A — the `map[string]any` round-trip destroys everything but values

`internal/cli/update/backup/merge.go` decodes YAML into `map[string]any`, merges the maps, then re-emits with `yaml.Marshal`:

- `MergeYAML3Way` — `merge.go:20-35` (unmarshal at `:21-31`, marshal at `:34`)
- `MergeYAMLDeep` — `merge.go:116-130` (unmarshal at `:117-124`, marshal at `:129`)

The map is a lossy intermediate representation: `yaml.Node`'s `HeadComment` / `LineComment` / `FootComment`, the source key order, the scalar quoting style, and the source indentation all exist only on the node tree and are discarded the moment the document becomes a `map[string]any`.

**Reproduced against the repo's pinned `gopkg.in/yaml.v3 v3.0.1`**, unmarshal→marshal of `internal/template/templates/.moai/config/sections/cache.yaml`:

```
comments in=16 out=0
```

```yaml
# BEFORE (source, 22 lines)             # AFTER (round-trip, 5 lines)
# Anthropic Prompt Caching strategy.    cacheStrategy:
# ...16 comment lines total...              enabled: true
cacheStrategy:                              min_cacheable_tokens: 2048
  # Toggle cache_control injection.         session_ttl: 1h
  enabled: true                             spec_ttl: 5m
  session_ttl: "1h"
  spec_ttl: "5m"
  min_cacheable_tokens: 2048
```

Four distinct losses in one round trip: **all 16 comments dropped**, **keys alphabetized** (`enabled, min_cacheable_tokens, session_ttl, spec_ttl` — source order was `enabled, session_ttl, spec_ttl, min_cacheable_tokens`), **quoting stripped** (`"1h"` → `1h`), **indent widened** (2-space → yaml.v3's default 4-space).

The loss occurs on **every successful merge**, not only on fallback. It is idempotent — a file already stripped stays stripped — which explains why the reported run visibly changed only those files that still carried comments in the user's committed tree.

### Defect B — the 3-way path silently deletes user-added keys

`DeepMerge3Way` deliberately drops keys present in the user's file but absent from the new template (`merge.go:95-97`):

```go
// Keys only in old (not in new template) are dropped:
// they were removed from the template, so we don't carry them forward.
```

A key the user added by hand is indistinguishable from a key the template retired, so both are deleted with no notice.

The asymmetry is the tell: the 2-way fallback `DeepMergeMaps` **preserves** old-only keys (`merge.go:168-171`, `// Only exists in old, add it`). The 3-way path — the one that runs when template defaults are available, i.e. the normal path — is strictly more destructive than its own degraded fallback.

A test currently pins the destructive behavior. `internal/cli/update_yaml_test.go:591-603`, subtest **`user_added_key_not_in_base_preserved`**, asserts the exact opposite of its own name:

```go
name: "user added key not in base preserved",
...
if contains(s, "user_added") {
    t.Errorf("expected user_added to be dropped (not in new template), got %s", s)
}
```

### Defect C — no preservation contract exists

`grep -rin "comment" internal/cli/update/backup/*_test.go internal/cli/update_yaml_test.go` returns **zero matches**. There is no test anywhere asserting that a merge preserves comments, key order, scalar style, or indentation. Establishing that contract is half the work of this SPEC: without it, a future refactor silently regresses the fix.

### Blast radius

Both entry points are reached from `restore.go` on the normal `moai update` path: `MergeYAML3Way` at `restore.go:121`, `MergeYAMLDeep` at `restore.go:139` (2-way fallback) and `restore.go:207` (legacy backup format). Every `.moai/config/sections/*.yaml` file in every user project passes through one of them. Across the 31 shipped section templates the comment payload is substantial — `llm.yaml` carries 119 comment lines, `lsp.yaml.tmpl` 88, `quality.yaml.tmpl` 84.

## §B Scope

**In Scope**:

- Replace the `map[string]any` round-trip in `MergeYAML3Way`, `MergeYAMLDeep`, `DeepMerge3Way`, and `DeepMergeMaps` with a `*yaml.Node` tree merge that preserves comments, key order, and scalar style.
- Pin the encoder indentation to 2 spaces (`enc.SetIndent(2)`).
- Reverse the old-only-key drop policy in the 3-way path to preserve-and-report, aligning it with `DeepMergeMaps`.
- Amend `internal/cli/update_yaml_test.go:591-603` and its misnamed subtest to assert preservation.
- Establish a comment/format-preservation test contract, including a golden round-trip test over every section template under `internal/template/templates/.moai/config/sections/`.
- Revise the string-assertion table in `internal/cli/update_yaml_test.go` (from `TestMergeYAML3Way` at `:471`) that conflicts with node-based output.
- Explicitly declare the handling of each YAML edge case: anchors/aliases, multi-document streams, sequence merge policy, null vs empty-map nodes, and merge keys (`<<`).

**Out of Scope**:

### Out of Scope — `.gitignore` user-pattern loss
- The `.gitignore` user-pattern regression is a duplicate of issue **#1131**, already fixed on `main` (`b4ef28e5d`) and pending release. This SPEC does not touch `.gitignore` handling.

### Out of Scope — clean-reinstall fallback reporting
- The clean-reinstall path's failure to report 3-way merge fallbacks was fixed in **PR #1247** (recorder wiring in `internal/cli/update_clean_install.go`). This SPEC does not touch the fallback recorder or its wiring.

### Out of Scope — `SaveTemplateDefaults` base derivation
- `SaveTemplateDefaults` (`internal/cli/update/backup/backup.go:149-192`) derives the 3-way merge BASE from the NEW embedded template rather than from the old template the user actually had installed, violating the 3-way premise. This is a real defect and is **deferred to a follow-up SPEC** — the decision and its rationale are recorded in `plan.md` §E Decision D5. It is named here so the deferral is explicit rather than implicit.

### Out of Scope — merge decision semantics
- The 3-way decision logic itself (`old == base` → take new; `old != base` → keep old; system-field always-new) is preserved byte-for-byte in meaning. This SPEC changes the representation the decision operates on, not the decision.

### Out of Scope — non-YAML config formats
- JSON settings files (`settings.json`, `settings.local.json`) merge through `internal/cli/settings.go` helpers and are untouched.

## §C Requirements (GEARS)

### Preservation

**REQ-UYP-001** — **When** `MergeYAML3Way` or `MergeYAMLDeep` completes a merge successfully, the merge function shall emit output in which every comment (`HeadComment`, `LineComment`, `FootComment`) present in the source document that supplied each surviving node is retained.

**REQ-UYP-002** — **When** `MergeYAML3Way` or `MergeYAMLDeep` completes a merge successfully, the merge function shall emit mapping keys in the key order of the document that supplied the mapping's structure, and shall not reorder them alphabetically.

**REQ-UYP-003** — **When** a scalar value is carried through a merge unchanged, the merge function shall preserve that scalar's source quoting style, so that `session_ttl: "1h"` does not become `session_ttl: 1h`.

**REQ-UYP-004** — The merge function shall emit block-mapping output at an indentation width of 2 spaces, by configuring the encoder with `SetIndent(2)`.

**REQ-UYP-005** — **While** a merge produces no change to any value, the merge function shall emit output byte-identical to the input template document.

### Key retention

**REQ-UYP-006** — **When** a key is present in the user's document and absent from the new template document, `DeepMerge3Way` shall retain that key and its value in the merged output.

**REQ-UYP-007** — **When** `DeepMerge3Way` retains a key that is absent from the new template, the merge function shall report the retained key path on stderr as an advisory notice, so the user learns that a key no longer shipped by the template survived.

**REQ-UYP-008** — The 3-way merge path shall not delete any key that the 2-way fallback path (`DeepMergeMaps`) would preserve.

### Decision preservation

**REQ-UYP-009** — The merge function shall preserve the existing 3-way decision semantics: a value equal to its base takes the new template value; a value differing from its base keeps the user value; a key with no base entry keeps the user value.

**REQ-UYP-010** — The merge function shall continue to take the new template value unconditionally for the system fields `version` and `template_version`, on both the 3-way and 2-way paths.

**REQ-UYP-011** — **When** the merge function encounters a YAML document it cannot parse, the merge function shall return a wrapped error using `fmt.Errorf("...: %w", err)`, preserving the existing caller fallback behaviour in `restore.go`.

### Edge cases

**REQ-UYP-012** — **When** a source document contains a YAML anchor or alias node, the merge function shall carry the node through without expanding the alias into its resolved value.

**REQ-UYP-013** — **When** a source document contains a merge key (`<<`), the merge function shall treat it as an ordinary mapping key and shall not resolve its semantics.

**REQ-UYP-014** — **When** a merged mapping value is a sequence node, the merge function shall replace the sequence wholesale rather than appending or element-wise merging.

**REQ-UYP-015** — **When** a source document is a multi-document stream (more than one document node), the merge function shall merge the first document and return a wrapped error naming the unsupported input rather than silently discarding the trailing documents.

**REQ-UYP-016** — **When** a node is an explicit null and its counterpart is an empty mapping, the merge function shall treat them as distinct values and shall not coerce one into the other.

### Test contract

**REQ-UYP-017** — The test suite shall contain a golden round-trip test that, for every `*.yaml` and `*.yaml.tmpl` file under `internal/template/templates/.moai/config/sections/`, asserts a no-user-edit merge preserves comment count, indentation width, key order, and scalar quoting.

**REQ-UYP-018** — The test suite shall not contain any assertion that requires a user-added key to be dropped from 3-way merge output.

**REQ-UYP-019** — The test suite shall place all YAML fixtures inside `_test.go` files or `t.TempDir()` directories, and shall not add fixture files under `internal/template/templates/`.

## §D Constraints

- **Pinned dependency**: `gopkg.in/yaml.v3 v3.0.1` — no new YAML library may be introduced. The `yaml.Node` API is already available in this version.
- **Template neutrality** (CLAUDE.local.md §25): no SPEC ID, REQ token, internal date, or commit SHA may appear in any file under `internal/template/templates/`.
- **Test isolation** (CLAUDE.local.md §6): every temporary directory via `t.TempDir()`; no writes into the project root from a test.
- **Go conventions**: `snake_case.go` filenames, `fmt.Errorf("...: %w", err)` error wrapping, English code comments and godoc.
- **Coverage**: `internal/cli/update/backup` must hold at or above its current package coverage after the change.
- **API stability**: the four exported function signatures (`MergeYAML3Way`, `MergeYAMLDeep`, `DeepMerge3Way`, `DeepMergeMaps`) are consumed by tests in four packages; signature changes cascade and must be justified in `plan.md`.

## §E Success Criteria

- A `moai update` over a project whose `.moai/config/sections/*.yaml` files are unmodified template copies produces a byte-identical tree.
- A `moai update` over a project with user-edited values preserves those values **and** every comment, key order, and quoting style in the file.
- A user-added key absent from the new template survives the merge and is reported on stderr.
- The full suite is green: `go test ./internal/cli/... ./internal/template/...`.

## §F Cross-References

- Issue **#1243** — the originating report.
- Issue **#1131** / commit `b4ef28e5d` — the `.gitignore` sibling defect (out of scope).
- PR **#1247** — the clean-reinstall fallback-reporting sibling defect (out of scope).
- `internal/cli/update/backup/merge.go` — the file under change.
- `internal/cli/update/backup/restore.go:121,139,207` — the three call sites.
- `internal/cli/update/backup/backup.go:149-192` — `SaveTemplateDefaults`, deferred (plan.md §E D5).
