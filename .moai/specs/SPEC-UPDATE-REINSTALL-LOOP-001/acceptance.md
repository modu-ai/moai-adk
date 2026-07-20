# Acceptance Criteria — SPEC-UPDATE-REINSTALL-LOOP-001

All criteria are observable via test output, filesystem state, or `git status`.
Each AC maps to a GEARS requirement in spec.md §B.

## §D AC Matrix

| AC | REQ | Severity | Summary |
|----|-----|----------|---------|
| AC-RIL-001 | REQ-RIL-001 | Critical | `.claude/rules/moai/design` no longer on `DeprecatedPaths` |
| AC-RIL-002 | REQ-RIL-003 | Critical | Build-time collision guard: DeprecatedPaths ∩ template FS = ∅ |
| AC-RIL-003 | REQ-RIL-002 | Critical | Repeated `moai update` on a v3 project is zero-net-change |
| AC-RIL-004 | REQ-RIL-004 | High | Count + category-split guards updated atomically and pass |
| AC-RIL-005 | REQ-RIL-007 | High | `settings.json` user keys survive clean-reinstall |
| AC-RIL-006 | REQ-RIL-008 | High | ALL `.moai/config/sections/*.yaml` survive clean-reinstall (user.yaml name, language.yaml, design.yaml, …) |
| AC-RIL-007 | REQ-RIL-009 | High | Clean-reinstall config handling matches normal-path protection |
| AC-RIL-008 | REQ-RIL-005 | Medium | Symlink / nested-`.git` aborts BEFORE any destruction, names the path — **[SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]** |
| AC-RIL-009 | REQ-RIL-006 | Medium | No half-migrated end state — **[SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]** |
| AC-RIL-010 | REQ-RIL-010 | Medium | Model pin does not silently downgrade (policy RESOLVED: merge-preserving — user model survives) |

## §D.1 Detailed Scenarios

### AC-RIL-001 — collision entry removed
- **Given** the `DeprecatedPaths` registry in `internal/defs/dirs.go`,
- **When** the registry is inspected,
- **Then** it contains no entry equal to `.claude/rules/moai/design`,
- **And** every remaining entry is a path the v3 template does NOT ship.
- **Evidence**: a test enumerating `defs.DeprecatedPaths` finds no `.claude/rules/moai/design` entry.

### AC-RIL-002 — build-time collision guard (regression-proof)
- **Given** the embedded template filesystem and `defs.DeprecatedPaths`,
- **When** the guard test runs (`go test ./internal/...`),
- **Then** it asserts the intersection is empty and PASSES,
- **And** re-adding ANY colliding entry (a path the template ships) makes the guard FAIL.
- **Evidence**: guard test passes on the fixed tree; a deliberately re-inserted colliding entry
  reproduces a FAIL (demonstrated in the test's own negative-path fixture or a temporary local
  check). This is a real executed test, not a token-presence grep.

### AC-RIL-003 — repeated update is zero-net-change (loop eliminated)
- **Given** a v3 project whose working tree contains `.claude/rules/moai/design/` and a v3
  `system.yaml`,
- **When** `moai update` is run twice in succession,
- **Then** neither run removes any deprecated path (no `Removed N deprecated paths` with N>0 on
  the design dir),
- **And** the working tree is unchanged after each run (`git status --porcelain` empty for
  template-managed paths).
- **Evidence**: behavioral test / integration test driving the update path twice and asserting
  removal count 0 and tree stability. NOT satisfied by a code-token grep.

### AC-RIL-004 — count guards updated atomically
- **Given** the `DeprecatedPaths` slice after the collision entry is removed,
- **When** `go test ./internal/defs/...` runs,
- **Then** `TestDeprecatedPathsTotalCount` and `TestDeprecatedPathsCategorySplit` PASS with the
  updated totals (total 40; Category B 28; A 9, C 3 unchanged),
- **And** the count recited in `dirs.go` (@MX:REASON) and `v2_detection.go` header comments
  matches the new total.
- **Evidence**: passing test output + grep showing no stale `41` count claim in the two source
  comments.

### AC-RIL-005 — settings.json user keys preserved
- **Given** a v3 project whose `.claude/settings.json` carries a user-customized value (e.g. a
  non-default `effortLevel` and a user-added permission),
- **When** a clean-reinstall executes on that project,
- **Then** the customized values are still present in `.claude/settings.json` afterward.
- **Evidence**: before/after read of the two keys shows them retained.

### AC-RIL-006 — all sections/*.yaml preserved (issue #1084 scope)
- **Given** a v3 project whose `.moai/config/sections/` carries user-populated values across
  MULTIPLE section files — at minimum a non-empty operator `name` in `user.yaml`, a non-default
  `conversation_language` in `language.yaml`, and a customized token in `design.yaml`,
- **When** a clean-reinstall executes,
- **Then** every one of those `sections/*.yaml` files is unchanged (`user.yaml` `name` not blanked
  to `""`, `language.yaml` / `design.yaml` values not reset to template defaults),
- **And** no user-populated `sections/*.yaml` file is silently overwritten by its `.tmpl` default.
- **Evidence**: before/after read of the `name` field AND the language + design values (issue #1084
  explicitly reports language.yaml/design.yaml loss — the AC MUST exercise more than user.yaml).

### AC-RIL-007 — clean-reinstall matches normal-path protection
- **Given** the normal `moai update` path's config protection (3-way merge of `settings.json`
  and `.moai/config/sections/*.yaml`),
- **When** the clean-reinstall path handles the same files,
- **Then** the resulting file content is equivalent to what the normal path would produce
  (user keys preserved, template additions merged in) — the clean path is not a
  lower-protection bypass.
- **Evidence**: a test comparing the clean-reinstall outcome against the normal-path merge
  outcome for a fixture with user customizations.

### AC-RIL-008 — pre-flight abort before destruction  [SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]
> This AC is SPLIT to the follow-up `SPEC-UPDATE-PREFLIGHT-SAFETY-001` (R2 pre-flight hardening).
> It carries NO Definition-of-Done weight in this SPEC; retained verbatim so the follow-up can lift it.
- **Given** a v3 project with a symlink (or a nested `.git`) inside a PRESERVE scan root,
- **When** a clean-reinstall is attempted,
- **Then** the operation aborts with an error that names the offending path,
- **And** no deprecated path has been removed and no template has been redeployed (filesystem
  state identical to before the attempt).
- **Evidence**: before/after filesystem snapshot shows zero removals; error message contains the
  offending path.

### AC-RIL-009 — no half-migrated end state  [SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001]
> This AC is SPLIT to the follow-up `SPEC-UPDATE-PREFLIGHT-SAFETY-001` (R2 pre-flight hardening).
> It carries NO Definition-of-Done weight in this SPEC; retained verbatim so the follow-up can lift it.
- **Given** any failure during the clean-reinstall after the destructive step,
- **When** the operation returns an error,
- **Then** the project is NOT left with deprecated paths removed + templates reinstalled but
  PRESERVE inventory unrestored (either fully restored, or the failure occurs before destruction).
- **Evidence**: fault-injection test asserting the invariant across the failure surface.

### AC-RIL-010 — model pin does not silently downgrade (policy RESOLVED: merge-preserving)
- **Given** a v3 project whose `.claude/settings.json` explicitly sets a higher-capability
  `model` (e.g. `opus`),
- **When** a clean-reinstall / update runs,
- **Then** the configured `model` is still present in `.claude/settings.json` afterward (NOT
  replaced by the template's `"model": "sonnet"` pin).
- **Status**: RESOLVED (plan-audit iter-1) — policy is option (b), merge-preserving. The template
  pin is KEPT; the user's `model` survives as a by-product of M1's `settings.json` merge (no
  separate template edit). This AC therefore verifies the user model SURVIVES; it is a
  by-product check of AC-RIL-005/007, not an independent template-change verification.

## §D.2 Definition of Done

- [ ] AC-RIL-001..004 pass (loop broken + guarded + counts consistent) — Critical/High, MUST.
- [ ] AC-RIL-005..007 pass (config preservation on clean-reinstall; AC-RIL-006 covers ALL
      `sections/*.yaml`, not only user.yaml) — High, MUST.
- [ ] AC-RIL-010 passes (user model survives via the merge-preserving settings handling) — Medium, MUST.
- [x] AC-RIL-008..009 DEFERRED — SPLIT to `SPEC-UPDATE-PREFLIGHT-SAFETY-001` (R2 pre-flight
      hardening); carry no DoD weight in this SPEC (plan-audit iter-1 decision).
- [ ] `go test ./...` green; `golangci-lint run` clean; template change (if any) followed by `make build`.
- [ ] `dirs_test.go` count guards and the source-comment counts updated in the same change.
