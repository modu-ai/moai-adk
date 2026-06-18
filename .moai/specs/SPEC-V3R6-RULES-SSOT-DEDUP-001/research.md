# Research — SPEC-V3R6-RULES-SSOT-DEDUP-001

## §1. Research Question

**Does the `moai constitution` CLI read the zone-registry `clause:` field text, such that reducing
each `clause:` to a 3-5 word non-verbatim label (the originally-proposed bloat reduction) would
break the CLI?**

This is the gating question for REQ-SSD-009 / milestone M_Z. The answer determines whether the
zone-registry reduction is a safe structural change or a CLI-breaking one requiring deferral.

## §2. Method

Grepped `internal/` for zone-registry / constitution consumption (`grep -rn "zone-registry|clause:"
internal/ --include="*.go"`), then read the three consuming code paths in full:
`internal/cli/constitution.go`, `internal/constitution/rule.go`, `internal/constitution/validator.go`.

## §3. Finding — the `clause:` field is LOAD-BEARING for the CLI (3 consumers)

### §3.1 `moai constitution list` renders `Clause` (table + JSON)
- `internal/cli/constitution.go:206,220` — the JSON output struct serializes `Clause: e.Clause`.
- `internal/cli/constitution.go:415-417` (`renderConstitutionTable`) — prints `e.Clause`, truncated
  to 40 chars in table mode.
- **Impact of label reduction**: cosmetic only. A 3-5 word label would still render. `list` alone
  does NOT block the reduction.

### §3.2 `moai constitution amend --before <text>` does an EXACT-MATCH against `Clause`
- `internal/cli/constitution.go:503-505` (`runConstitutionAmend`):
  ```go
  if rule.Clause != before {
      return fmt.Errorf("clause mismatch: --before differs from current clause...")
  }
  ```
- **Impact of label reduction**: any future `amend` invocation would need `--before` = the label,
  not the verbatim clause. Lower-frequency operation; not a hard blocker but a behavior change.

### §3.3 `moai constitution validate` requires `clause` to be a VERBATIM SOURCE SUBSTRING (decisive)
- `internal/constitution/validator.go:262-273`:
  ```go
  normalizedClause := normalizeWhitespace(entry.Clause)
  normalizedSource := normalizeWhitespace(stripCodeFences(sourceContent))
  if normalizedClause != "" && !strings.Contains(normalizedSource, normalizedClause) {
      result.DriftCount++
      // ... SentinelDrift finding ...
  }
  ```
- The validate command (`internal/cli/constitution.go:246` Long-help: "Checks that every registry
  entry's clause exists in the source file") emits a `SentinelDrift` finding for **every** entry
  whose `clause:` is not found (after whitespace normalization, code-fence stripping) as a substring
  of the source file.
- **Impact of label reduction**: a 3-5 word label that is NOT a verbatim excerpt of the source HARD
  clause would produce a `SentinelDrift` for that entry → `validate` returns drift status (exit 1)
  → the drift detector is broken for all reduced entries. **This is the decisive constraint.**

### §3.4 `Rule.Validate()` only requires `Clause` non-empty (NOT a constraint)
- `internal/constitution/rule.go:69-71`:
  ```go
  if r.Clause == "" {
      return fmt.Errorf("rule %q: Clause field is empty", r.ID)
  }
  ```
- A 3-5 word label is non-empty → loading does NOT break. The constraint is `validate`, not load.

## §4. Conclusion

The originally-proposed "reduce each `clause:` to a 3-5 word label (NOT verbatim clause text)" is
**UNSAFE as-is**: it breaks `moai constitution validate` drift detection (§3.3) and changes
`amend --before` semantics (§3.2). The `Clause` field is verbatim-source-substring-coupled by
design — that coupling IS the drift detector.

### §4.1 Safe path (adopted in REQ-SSD-009 / design.md §7.1)
- Reduce each `clause:` to a **still-verbatim shorter excerpt** (first sentence / operative MUST
  clause). This keeps `validate` green while shrinking bloat.
- Narrow the over-broad `paths:` trigger (§7.2) — unconstrained, fully safe.

### §4.2 Blocked path → follow-up SPEC
- The full structural reduction to label-only entries (`{id, zone, file, anchor, label}` with a
  3-5 word label and NO verbatim clause) requires a `validator.go` change to make the drift check
  resolve the clause text from the source anchor instead of requiring it inline. This Go change is
  OUT OF SCOPE for this SPEC (spec.md §G — no Go source edit). Deferred to follow-up SPEC
  **`SPEC-V3R6-CONST-VALIDATE-LABEL-001`** (recommended, not created here).

## §5. CLI consumption evidence map (for run-phase re-confirmation)

| Consumer | File:line | Reads `Clause`? | Reduction impact |
|----------|-----------|-----------------|------------------|
| `constitution list` (table) | `internal/cli/constitution.go:415-417` | yes (truncated render) | cosmetic |
| `constitution list` (json) | `internal/cli/constitution.go:206,220` | yes | cosmetic |
| `constitution amend --before` | `internal/cli/constitution.go:503-505` | yes (exact match) | behavior change |
| `constitution validate` | `internal/constitution/validator.go:264` | yes (substring-in-source) | **BREAKS on non-verbatim label** |
| `Rule.Validate()` (load) | `internal/constitution/rule.go:69` | non-empty only | none |

## §6. Secondary research notes

- **zone-registry size**: 1019 lines, ~230 entries (verified `wc -l` + entry grep).
- **`paths:` trigger**: `.claude/**,.moai/specs/**,.claude/rules/**` (line 3) — loads on every
  `.moai/specs/**` edit, i.e. every SPEC-authoring turn (~13K-token load). Narrowing is safe
  (§7.2) pending run-phase verification that no rule depends on the registry during SPEC authoring.
- **mirror class**: zone-registry.md IS mirrored (`internal/template/templates/.claude/rules/moai/core/zone-registry.md`
  exists) — byte-parity (verify allowlist membership at run time).
- **CONST-V3R5-022** (the context-window threshold entry, T3): its de-dup must be a body/pointer
  edit, NOT a `clause:` blanking, for the same §3.3 reason.
