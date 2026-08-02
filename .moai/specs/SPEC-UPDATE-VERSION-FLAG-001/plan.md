# Plan — SPEC-UPDATE-VERSION-FLAG-001

> Implementation plan. Plan-phase ONLY — do NOT enter run phase from
> this document. Decisions most likely to change are ordered FIRST
> (Approach-First ordering per CLAUDE.md §7 Rule 1).

## §A Context

This SPEC adds a `--version <tag>` flag to `moai update` so a user can
install a specific release tag of the moai binary — covering stable, rc,
and downgrade/rollback targets — unified into the update CLI. Scope is
binary-only; template rollback stays on the existing `--restore` /
template-sync flow. See `spec.md` §A–§F for the full contract.

**Verified touch surface** (cite, do not re-derive):
- `internal/cli/update.go` — register `--version` flag, conflict
  validation in `validateUpdateFlags`, branch inside `runUpdate`.
- `internal/cli/deps.go` — add a tag-resolution helper that stays on the
  `api.github.com` allowlist; do NOT modify the existing
  `allowedUpdateHost` / `validateUpdateURL` contract.
- `internal/cli/v2_detection.go:253` `normalizeVersionMajor` — reuse
  discipline for v-prefix normalization (do NOT fork).
- `docs-site/content/{ko,en,ja,zh}/cli-reference/update.md` —
  4-locale docs update.
- `docs-site/content/{ko,en,ja,zh}/getting-started/installation.md` —
  cross-reference `--version` from the install page.

**No-touch surface** (REQ-UVF-015):
- `internal/template/templates/**` — out of scope (this is moai-adk-go's
  own Go code, not a distributed template).

## §B Known Issues

1. **`--version` flag-name collision risk** (orchestrator Q1). Resolved
   in spec.md §F.1: cobra subcommand scoping disambiguates; canonical
   name `--version`, aliasing to `--tag` deferred to run-phase judgment.
2. **Dev/RC branch interaction** (orchestrator Q, deps.go:387). Resolved
   in REQ-UVF-012: `--version` ALWAYS resolves the explicit tag; the
   existing dev-branch `/releases` list endpoint continues to serve
   default `moai update` for dev builds.
3. **Partial-install hazard on checksum mismatch**. Mitigation:
   REQ-UVF-010 mandates discard-before-install; AC-UVF-010 verifies the
   binary path is unchanged after a checksum-mismatch exit.
4. **Allowlist escape via tag content**. A `<tag>` containing URL
   metacharacters (e.g. `../`, `?`) could be used to attempt path
   traversal inside the GitHub API URL. Mitigation: run-phase validation
   rejects any `<tag>` that is not `^v?[0-9A-Za-z.\-]+$` before URL
   construction (this is an implementation detail, not a REQ-level
   constraint, but the implementer MUST add it).

## §C Pre-flight (Before Run Phase)

- [ ] Implementation Kickoff Approval gate passed (orchestrator
  `AskUserQuestion`).
- [ ] Plan-auditor verdict on this plan.md ≥ PASS (plan-audit gate).
- [ ] TDD cycle confirmed: RED first (reproduction tests for
  REQ-UVF-008..011 defect paths), then GREEN, then REFACTOR.
- [ ] Branch created from `origin/main` (Tier M PR-mandatory per
  CLAUDE.local.md §23).

## §D Constraints Carried Into Run Phase

- **Tier M ceiling**: ≤16 REQ / ≤16 AC (at ceiling; do not add more
  without a Tier bump).
- **Checksum verification MANDATORY**: no bypass flag (REQ-UVF-006).
- **Allowlist MANDATORY**: stay on `api.github.com` / `https`
  (REQ-UVF-005).
- **Default-behavior preservation**: golden characterization test for
  default `moai update` is a merge blocker (AC-UVF-004).
- **4-locale same-PR obligation**: ko/en/ja/zh in one PR
  (`hns-oss-docs-i18n-rules`).
- **TDD**: reproduction-first for every defect path. No defect-path REQ
  ships without a failing-then-passing test.

## §E Self-Verification (Plan-Phase)

This plan.md carries:
- [ ] All 16 REQ-UVF-* traceable to ≥1 AC-UVF-* (acceptance.md §D matrix).
- [ ] Every Key Design Question (§F in spec.md) resolved with rationale.
- [ ] Out of Scope section satisfies the `OutOfScopeRule` lint (5 H3
  sub-headings with `-` bullets in spec.md §G).
- [ ] No implementation detail (function names, signatures) beyond what
  is needed to locate the touch surface — HOW belongs to run phase.

## §F Milestones (priority-ordered, no time estimates)

Ordering rationale: decisions most likely to change lead (data-model /
URL shapes, user-facing UX), mechanical steps defer to the bottom.

### M1 — Tag resolution + URL shape (highest reversibility)
- Implement v-prefix normalization reusing `normalizeVersionMajor`
  discipline.
- Implement tag → `releases/tags/<tag>` URL construction.
- Implement `<tag>` character-class validation (Known Issue #4).
- Unit tests: normalization accepts `v3.0.0` + `3.0.0`, rejects
  `go-v3.0.0`, rejects URL-metacharacter tags.

### M2 — Flag registration + mutual-exclusion matrix (user-facing UX)
- Register `--version <tag>` in `update.go init()`.
- Extend `validateUpdateFlags` with the conflict matrix (REQ-UVF-007).
- Unit tests: every conflict cell exits non-zero before any network
  call.

### M3 — Default-behavior preservation (regression anchor)
- Capture golden characterization test proving default `moai update`
  (no `--version`) is byte-identical to pre-SPEC baseline
  (AC-UVF-004).
- This MUST land before any `--version` branch is added to `runUpdate`,
  so the regression anchor is unambiguous.

### M4 — Install path + checksum verification (security-critical)
- Wire `--version` into the existing `UpdateOrch` /
  checksum-verified download path. Reuse, do not fork.
- Reproduction-first tests for REQ-UVF-008 (404), REQ-UVF-009 (no
  asset), REQ-UVF-010 (checksum mismatch), REQ-UVF-011 (network
  failure).
- Verify no partial install on any defect path.

### M5 — Downgrade confirmation + re-exec (UX boundary)
- Implement interactive downgrade prompt (REQ-UVF-013).
- Verify `--yes` / non-TTY skip.
- Verify re-exec after install (REQ-UVF-014).

### M6 — Dev/RC branch interaction preservation
- Test that an explicit `--version` overrides the deps.go:387 dev-branch
  fallback (REQ-UVF-012).
- Test that default `moai update` on a dev build STILL uses the
  dev-branch `/releases` endpoint (no regression).

### M7 — Docs-site 4-locale update (mechanical, lowest reversibility)
- Update `cli-reference/update.md` × 4 locales.
- Update `getting-started/installation.md` × 4 locales (cross-reference).
- Run `hns-oss-docs-verify` exit-gate recipe before PR.

### M8 — PR + CI
- Open PR (Tier M, PR-mandatory per CLAUDE.local.md §23).
- CI guard: confirm `git diff` touches no `internal/template/templates/`
  path (REQ-UVF-015).
- Self-merge after 4 CI checks green (1-person OSS policy).

## §G Anti-Patterns (Run-Phase Prohibitions)

- **AP-UVF-001**: writing a parallel downloader. Reuse
  `buildAutoUpdateFunc` / `UpdateOrch`. (REQ-UVF-002)
- **AP-UVF-002**: adding a `--skip-checksum` bypass. Forbidden.
  (REQ-UVF-006)
- **AP-UVF-003**: broadening `allowedUpdateHost` to accommodate a
  non-GitHub source. Out of scope; cf. SPEC-SEC-HARDEN-005.
- **AP-UVF-004**: reusing the dev-branch `/releases` list endpoint for
  default `moai update` when `--version` is present — breaks
  REQ-UVF-012.
- **AP-UVF-005**: silently treating `--version` absence as
  `--version latest`. Default MUST stay `/releases/latest`
  (REQ-UVF-004).
- **AP-UVF-006**: touching `internal/template/templates/**`. Forbidden
  (REQ-UVF-015).

## §H Cross-References

- **spec.md** §A–§H — the WHAT and WHY (this plan.md is the HOW).
- **acceptance.md** §D — the AC-UVF-* matrix binding every REQ.
- **SPEC-UPDATE-DATA-SURVIVAL-001** — `--restore` contract (referenced,
  not duplicated).
- **SPEC-SEC-HARDEN-005** — `MOAI_RELEASES_DIR` / allowlist (referenced,
  not duplicated).
- **SPEC-UPDATE-REINSTALL-LOOP-002** — `normalizeVersionMajor`
  discipline source (referenced, not duplicated).
- **CLAUDE.local.md §23** — PR-mandatory 1-person OSS policy.
- **CLAUDE.local.md §25** — template-neutrality N/A here (no template
  touch); enforced instead via REQ-UVF-015 no-touch invariant.
