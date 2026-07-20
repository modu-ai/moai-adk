# Acceptance Criteria — SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001

All criteria are observable via mechanical commands. `<mirror>` = `internal/template/templates/.claude/rules/moai/core/zone-registry.md`; `<source>` = `.claude/rules/moai/core/zone-registry.md`.

## AC-ZRP-001 — Neutralized mirror carries zero forbidden tokens (REQ-ZRP-001)

- **Given** the Strategy-A neutralized mirror,
- **When** it is scanned for the default-tier leak classes and the neutrality binary classes,
- **Then** every scan returns zero matches.

Verifiable:
```bash
# Default-tier leak class C1 (whole-tree SPEC-ID) — expect 0
grep -cE '\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b' <mirror>   # → 0
# §25 internal dates (strict-tier S1) — expect 0
grep -cE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b' <mirror>               # → 0
# Neutrality binary classes — expect 0 each
grep -cE '/Users/|CLAUDE\.local\.md|PR #[0-9]+|\| *[A-Za-z]+ \(canonical\)' <mirror>  # → 0
```

## AC-ZRP-002 — Mirror exists with full CONST content preserved (REQ-ZRP-002)

- **Given** the template tree,
- **When** the mirror path is inspected,
- **Then** the file exists and carries the same count of `CONST-*` entries as the source (115), proving neutralization did not drop registry content.

Verifiable:
```bash
test -f <mirror> && echo EXISTS
[ "$(grep -cE '^- id: CONST-' <mirror>)" = "$(grep -cE '^- id: CONST-' <source>)" ] && echo COUNT_MATCH   # 115 == 115
```

## AC-ZRP-003 — Build embeds the mirror (REQ-ZRP-003)

- **Given** the mirror under `internal/template/templates/`,
- **When** `make build` runs,
- **Then** the build succeeds and the recompiled binary's embedded FS resolves `.claude/rules/moai/core/zone-registry.md`.

Verifiable:
```bash
make build   # exit 0
# Embedded-FS resolution is proven transitively by AC-ZRP-004 (moai init deploys from the embedded FS).
```

## AC-ZRP-004 — Deployed project consumes the registry without break (REQ-ZRP-004)

- **Given** the freshly built binary,
- **When** `moai init` deploys into a fresh temp directory and the three registry-consuming commands run there,
- **Then** the file is present and none of the commands hit the "not found" break.

Verifiable:
```bash
TMP=$(mktemp -d); <built-moai> init "$TMP"/proj
test -f "$TMP"/proj/.claude/rules/moai/core/zone-registry.md && echo DEPLOYED
cd "$TMP"/proj
<built-moai> constitution list        # returns entries (non-empty), no load error
<built-moai> doctor                    # zone-registry check NOT "not found"/Warn-for-missing
```

Pass condition (all three OBSERVABLE):
1. `constitution list` returns registry entries (non-empty stdout, no load error).
2. `doctor` does not print `zone-registry.md not found` (the check is no longer downgraded to Warn-for-missing).
3. **`spec lint`'s registry-skip failure mode is eliminated by file presence.** `spec lint` silently skips the registry dimension ONLY when `detectRegistryPath(cwd)` returns `""` for an ABSENT file, and it emits no verbose/debug line surfacing the resolved path (verified: `spec_lint.go` has no such flag). The single observable proxy is therefore the deployed file: `test -f "$TMP"/proj/.claude/rules/moai/core/zone-registry.md` → EXISTS. With the file present, `detectRegistryPath` returns a non-empty path and the absence-skip branch is unreachable. (The prior "spec lint resolves a non-empty registry path" sub-assertion was dropped — it is not observable from CLI output.)

## AC-ZRP-005 — All six template CI guards + package tests green (REQ-ZRP-005)

- **Given** the mirror, the sanitized-pair registration, and the governance-token exemption,
- **When** the template test package runs (default and strict leak tiers),
- **Then** all six guards (§spec.md A.4.1) pass, including the two previously-missed guards.

Verifiable:
```bash
go test ./internal/template/...                                             # PASS (all six guards)
MOAI_TEMPLATE_LEAK_STRICT=1 go test -run TestTemplateNoInternalContentLeak ./internal/template/...  # PASS
# Explicitly pin the two previously-missed guards (D1/D2):
go test -run 'TestRuleProvenanceAudit|TestRuleProvenanceRecurrenceBackstop' ./internal/template/...  # PASS
go test -run 'TestRuleDateProvenance|TestRuleDateProvenanceRecurrenceBackstop' ./internal/template/...  # PASS (mirror has 0 ISO dates → RULE_DATE_PROVENANCE_LEAK absent)
```

## AC-ZRP-006 — Sanitized-pair drift guard green (REQ-ZRP-006)

- **Given** `.claude/rules/moai/core/zone-registry.md` registered in `sanitizedPairPaths`,
- **When** `TestSanitizedPairParity` runs,
- **Then** it passes: after token normalization the source and mirror doctrine are identical or within the 4-line reword tolerance (the only divergence is the 4 normalized tokens), with no `SANITIZED_PAIR_PARITY_DRIFT`.

Verifiable:
```bash
go test -run TestSanitizedPairParity ./internal/template/...   # PASS, no SANITIZED_PAIR_PARITY_DRIFT
```

## AC-ZRP-007 — Governance-token guard green via single-path file-level exemption (REQ-ZRP-007)

- **Given** the mirror carrying all 119 `CONST-V3R` occurrences and the `governanceTokenFileAllowlist` single-path entry for `zone-registry.md`,
- **When** `TestRuleProvenanceAudit` runs, and separately when a synthetic governance token is placed in a DIFFERENT rules file,
- **Then** the guard is green on the mirror (119 CONST occurrences exempted) AND still fires on the synthetic token elsewhere — proving the exemption is single-path, not a class-wide disable.

Verifiable:
```bash
# (a) Guard green WITH the 119 CONST occurrences present in the mirror.
grep -coE '\bCONST-V3R[0-9]-[0-9]+\b' internal/template/templates/.claude/rules/moai/core/zone-registry.md  # → 119 (content preserved)
go test -run TestRuleProvenanceAudit ./internal/template/...   # PASS, no RULE_GOVERNANCE_TOKEN_LEAK

# (b) Single-path property — the allowlist is exactly one entry (zone-registry.md).
grep -A3 'governanceTokenFileAllowlist = map' internal/template/rule_provenance_audit_test.go  # exactly 1 path key

# (c) Guard still fires elsewhere: append a synthetic CONST token to any OTHER rules file, run the guard,
#     observe RULE_GOVERNANCE_TOKEN_LEAK, then revert. (Manual RED probe; the raw-regex fire is already
#     pinned deterministically by TestRuleProvenanceRecurrenceBackstop's CONST-V3R6-099 leaky case.)
```

- **Byte-parity guard untouched**: `zone-registry.md` is NOT enrolled in `workflowOptMirroredPaths` / `lateBranchMirroredPaths`, so `TestRuleTemplateMirrorDrift` neither requires byte-identity nor flags the intentional divergence. Confirm no accidental enrollment.
- **Pedagogical allowlist**: `zone-registry.md` is NOT in the leak-test pedagogical allowlist; the mirror must be genuinely clean (no reliance on an exemption).
- **Strict-mode dates**: even though the default leak tier ignores bare dates, the mirror removes them so the strict-tier run (AC-ZRP-005) is also clean.
- **CONST-token guard is unconditional**: `TestRuleProvenanceAudit` has NO strict-mode gate — it runs on every `go test`. The 119 CONST occurrences would fail it verbatim; only the file-level exemption (AC-ZRP-007) keeps it green. This is the D1 blocking path.
- **C1b is skill-body-scoped, not whole-tree**: the `CONST-V3R`-matching class in `internal_content_leak_test.go` (C1b) is `skillBodyScoped: true`, so it does NOT fire on this `.claude/rules/` file — the CONST tokens trip ONLY guard #5, which is why a single exemption there suffices.
- **Deploy genericity**: no `catalog.yaml` edit; the file must deploy purely via the `fs.WalkDir` mechanism (AC-ZRP-004 proves this).

## Quality gate / Definition of Done

- [ ] AC-ZRP-001 .. AC-ZRP-007 all pass with cited command output.
- [ ] `go test ./internal/template/...` green (all SIX guards: default + strict leak, neutrality, byte-parity, sanitized-pair, governance-token provenance, date provenance).
- [ ] `go build ./...` green.
- [ ] Source `.claude/rules/moai/core/zone-registry.md` unchanged (Strategy A — `git diff` shows no change to the source).
- [ ] Deployed temp project: file present + 2 commands (`constitution list`, `doctor`) operate without the "not found" break; `spec lint` absence-skip eliminated by file presence.
- [ ] No `CONST-*` entry added/removed/semantically changed (119 occurrences preserved; kept green by the file-level governance-token exemption, not by removal).
- [ ] Governance-token exemption is single-path (zone-registry.md only) — guard still fires on governance tokens in other rules files.
