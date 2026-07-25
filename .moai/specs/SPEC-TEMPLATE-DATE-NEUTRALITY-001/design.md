# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Design

Design decisions are ordered by reversibility: the hardest-to-reverse and most-likely-to-change decisions come first.

---

## §A Decision 1 — Carve-out mechanism (highest change likelihood)

The guard must stop reporting the ~77 findings whose disposition is PRESERVE (DC-1 48 + DC-3 9 + DC-4 3 + the PRESERVE subset of DC-5), without stopping it from reporting a genuinely internal date added later. Four mechanisms are available. None is obviously correct; the trade-offs differ along maintenance cost, false-negative surface, and future-date behaviour.

### Option 1 — Per-finding allowlist, copying the `pedagogicalAllowlist` shape

Add a `dateAllowlist []dateAllowlistEntry{File, DateLiteral, Rationale}` consulted by `collectLeakViolations`, mirroring the existing `isPedagogicallyAllowed` gate.

**Correction to a common assumption about this precedent.** `pedagogicalAllowlist` is frequently described as line-number-anchored because its entries carry `LineStart` / `LineEnd`. That is not how it enforces. The enforcement path is:

```go
func isPedagogicallyAllowed(relPath, matched string) bool {
	for _, entry := range pedagogicalAllowlist {
		if entry.File == relPath && entry.SpecID == matched {
			return true
		}
	}
	return false
}
```

The match is `(File, SpecID)` — a content anchor. `LineStart` / `LineEnd` are documented in the struct comment as "diagnostic-only (recorded for human review and future drift detection)". So the precedent is already content-anchored, and copying it does **not** import a line-drift hazard.

- Pros: exactly mirrors an accepted in-repo precedent; zero false negatives (each entry names one literal in one file); reviewable in a PR diff.
- Cons: ~77 entries is an order of magnitude larger than the existing 5-entry allowlist. Violates REQ-TDN-012 for DC-4: a future `NOTICE.md` import entry would need a Go-file edit before it could be committed. Every skill frontmatter `updated:` bump would also need a paired allowlist edit — a per-commit tax on ordinary authoring.

### Option 2 — Structural narrowing of the S1 regex / class

Keep S1 as a whole-tree class but exclude the structural shapes that are legitimate, by extending `leakClass` with a line-context gate analogous to the existing `requireHexLetter` gate. Two sub-gates cover the bulk:

- exclude a match on a YAML frontmatter line whose key is `updated` / `created` / `version` (kills DC-1, 48 findings);
- exclude a match inside an attribution record file (kills DC-4, 3 findings).

- Pros: satisfies REQ-TDN-012 by construction — a future frontmatter bump or a future `NOTICE.md` import entry needs no code change. Reduces the residual allowlist to a handful of entries. Uses the same extension point the class struct already provides.
- Cons: widens the false-negative surface — a genuinely internal date that happens to sit on an `updated:` line is no longer caught. Requires the matcher to become line-aware; `collectLeakViolations` currently operates on whole-file text via `FindAllString`, so this is a real structural change to the scanning loop, not a field addition.

### Option 3 — Semantic marker in the source text

Require an inline marker on any legitimately-dated line (e.g. a trailing `<!-- neutral-date: attribution -->`), and gate S1 on marker absence.

- Pros: the justification lives next to the date; no central registry; self-documenting.
- Cons: pollutes the distributed template with harness-internal markup — itself an isolation-doctrine concern. Rejected on that ground unless the marker is invisible in rendered output and the doctrine explicitly admits it.

### Option 4 — Hybrid (structural gate + small residual allowlist)

Apply Option 2's structural gates for the two high-volume mechanical shapes (DC-1, DC-4), and Option 1's content-anchored allowlist for the low-volume judgement calls (DC-3's 9 deadline dates and the PRESERVE subset of DC-5).

- Pros: satisfies REQ-TDN-012 where it matters (recurring shapes), keeps the allowlist small enough to review, and keeps judgement calls explicit rather than swallowed by a broad structural rule.
- Cons: two mechanisms to understand instead of one; the boundary between "structural shape" and "judgement call" must be stated or it will drift.

### Recommendation

Option 4, with the boundary stated as: a structural gate is admissible only when the shape is (a) mechanically decidable from the line's own syntax and (b) expected to recur in ordinary authoring. Everything else is an allowlist entry. This is a recommendation for the audit and the user to accept or reject, not a settled decision.

---

## §B Decision 2 — DC-1 disposition (48 findings, schema-coupled)

DC-1 is the largest single category and the one the problem statement did not anticipate. These are skill and rule frontmatter fields:

```
  updated: "2026-03-30"
```

`.claude/rules/moai/development/skill-authoring.md` documents `updated:` as an ISO-date string field of the skill frontmatter schema. Three dispositions exist:

1. **Preserve as-is** (REQ-TDN-009 default). The field stays; the guard carves it out. The distributed template then carries the authoring date of each skill — arguably meaningful to a user judging skill freshness, arguably an internal work date.
2. **Preserve the key, neutralize the value** at render time (template variable, or a fixed sentinel). Keeps the schema valid; removes the date. Requires touching the render path for `.md` files that are not currently templated, which is a much larger change.
3. **Remove the field from the schema.** Out of scope — it would be a skill-authoring schema change, not a neutrality change.

Disposition 1 is the SPEC's default because it is the only one that does not force a schema or render-path change. Recorded as an open question because the "internal work date" reading is legitimate and only the user can settle it.

---

## §C Decision 3 — CI enforcement

Enforcement is gated on preconditions, not scheduled:

| Precondition | Verifiable by |
|---|---|
| P1 — strict tier reports zero findings after remediation and carve-out | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` exits 0 |
| P2 — the carve-out satisfies REQ-TDN-012 for at least DC-4 | adding a synthetic future attribution date does not produce a finding |
| P3 — narrow tier and the existing neutrality workflow remain green | their own runs |

Only when P1-P3 hold is the workflow edited. The edit follows the existing workflow's convention of running one isolated test target by name, because the `internal/template` package carries pre-existing unrelated failures and a package-wide green is not available as a gate.

**Future-legitimate-date regression risk.** A CI-enforced strict tier that blocks a legitimate future `NOTICE.md` entry is a net loss. P2 is the precondition that makes this concrete rather than aspirational: the enforcement decision is not taken until a synthetic future-date probe is shown to pass.

---

## §D Decision 4 — Report cap

`limit := 50` currently truncates with `... 85 more (capped)`, hiding 63% of findings from anyone running the guard. The source comment justifies this by pointing at a `grep -rln` recipe in the isolation doctrine as the "real audit log". That indirection is what made the true finding count invisible until a maintainer temporarily raised the cap by hand.

Options: (a) remove the cap; (b) make the cap a named constant overridable by env var; (c) keep the cap for console output but always write the full listing to a file whose path appears in the truncation message.

Option (c) satisfies REQ-TDN-016 without making a 135-row failure unreadable in CI logs. Option (b) is the smaller change and is acceptable if the env var is documented in the truncation message itself.

---

## §E Decision 5 — Local-mirror handling (mechanical, low change likelihood)

115 of the 116 affected template files have a counterpart in the local working tree. The naive move — copy the neutralized template file over the local copy, or the reverse — is prohibited: the local copies legitimately retain internal-development content that the template copies must not, and a blind copy in either direction breaks one of the two isolation policies.

The mirror-parity guard (`rule_template_mirror_test.go`) enforces byte-identity for an explicit, non-glob allowlist of rule files. Verified: none of the six date-bearing files under `templates/.claude/rules/moai/` appears in that allowlist. Byte-parity is therefore not a constraint on this remediation, and the local copies stay untouched by default.

The one exception to check at run phase: if a remediating edit lands on a file that *is* in the mirror allowlist, that file's local copy must receive the identical edit in the same commit.

---

## §F Cross-references

- `internal/template/internal_content_leak_test.go` — the guard (class definitions, `collectLeakViolations`, `isPedagogicallyAllowed`, the report cap)
- `internal/template/rule_template_mirror_test.go` — the byte-parity allowlist consulted in §E
- `.github/workflows/template-neutrality-check.yaml` — the isolated-target workflow convention referenced in §C
- `.moai/docs/template-internal-isolation-doctrine.md` — the C1-C8 content-class catalogue this SPEC's C3 class belongs to
