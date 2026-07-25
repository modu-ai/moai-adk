# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Design

Decisions are ordered by reversibility: the hardest-to-reverse and most-likely-to-change decisions come first. All five decisions below are now **settled** by user ruling (iteration 2); the alternatives are retained because a reader needs to know what was rejected and why.

---

## §A Decision 1 — Carve-out mechanism — SETTLED: hybrid (Option 4)

The guard must stop reporting the rows whose disposition is PRESERVE without stopping it from reporting a genuinely internal date added later.

**How many rows that is depends on M2.** Five of the six categories carry a plan-time-fixed disposition; `DC-5` does not — `spec.md` REQ-TDN-005 requires each of its 22 rows to be adjudicated individually, and `plan.md` M3 schedules a non-empty `DC-5` REMOVE subset. Writing `k` for the number of `DC-5` rows adjudicated REMOVE (`0 ≤ k ≤ 22`):

| Exit mechanism | Rows | At `k = 0` | At `k = 22` |
|---|---|---:|---:|
| Deleted by remediation (M3) | `80 + k` | 80 | 102 |
| Carved out by the guard (M4) | `100 − k` | 100 | 78 |
| **Total** | **180** | 180 | 180 |

The carve-out therefore sizes to `100 − k` rows, not a fixed 100. An earlier draft of this section asserted all 100 non-`DC-2a` rows were PRESERVE, which contradicted `spec.md:82`, REQ-TDN-005, and `plan.md` M3 step 2. The conclusion it supported — that AC-TDN-007 and AC-TDN-012 are jointly reachable — survives unchanged and is in fact *stronger* under the corrected premise: the identity holds for every `k`, so joint reachability is robust rather than conditional on a particular M2 outcome.

**Decision: Option 4 (hybrid).** Structural gates for `DC-1` and `DC-4`; a content-anchored allowlist for `DC-3`, `DC-2b`, and the PRESERVE subset of `DC-5`. Binding as REQ-TDN-010 in `spec.md` — this is a requirement, not a recommendation, so no downstream reader has to re-derive the choice.

**Admissibility boundary (accepted as stated).** A shape qualifies for a structural gate only where it is (a) mechanically decidable from the line's own syntax and (b) expected to recur in ordinary authoring. Everything else is an allowlist entry. `DC-1` (`^\s+updated:\s*"?20…`) and `DC-4` (a fixed file path) satisfy both. `DC-3` (a semantic deadline), `DC-2b` (a curatorial judgement about which directory mirrors third-party docs), and `DC-5` (per-row adjudication) satisfy neither.

### Enforcement surfaces (named, since the mechanism is decided)

| Component | Function | What it gates |
|---|---|---|
| Structural gate | `collectLeakViolations` (`internal_content_leak_test.go:593-634`) | `DC-1` (frontmatter `updated:` line shape) and `DC-4` (attribution file path) |
| Allowlist gate | `isDateAllowlisted` (new, sibling of `isPedagogicallyAllowed`) | `DC-3`, `DC-2b`, `DC-5-PRESERVE` |

Allowlist size: `DC-3` 13 + `DC-2b` 11 + `DC-5-PRESERVE` `(22 − k)` = **`24 + (22 − k)`** entries — between 24 (if M2 adjudicates every `DC-5` row REMOVE) and 46 (if none). The structural gate absorbs the other `DC-1` 48 + `DC-4` 6 = 54 rows and does not grow with `k`. Sizing the allowlist at a fixed 46 would assume `k = 0`, which REQ-TDN-005 forbids.

The structural gate must live in `collectLeakViolations` because that function is where per-class gating already happens (`skillBodyScoped`, `skillMoaiScoped`, `requireHexLetter`). It currently scans whole-file text via `FindAllString`, so adding a line-shape gate requires the scan loop to become line-aware — a real structural change, not a struct-field addition. That cost is the main price of Option 4 and is scheduled as M4.

### Options considered and rejected

**Option 1 — per-finding allowlist only** (copying the `pedagogicalAllowlist` shape). Rejected: 100 entries is 20× the existing 5-entry allowlist, and it violates REQ-TDN-012 — every future `NOTICE.md` import and every skill frontmatter bump would need a paired Go edit.

**Correction to a common assumption about that precedent.** `pedagogicalAllowlist` is often described as line-number-anchored because its entries carry `LineStart`/`LineEnd`. That is not how it enforces:

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

The match is `(File, SpecID)` — a content anchor. `LineStart`/`LineEnd` are documented in the struct comment as "diagnostic-only (recorded for human review and future drift detection)". The precedent is already content-anchored, so copying its enforcement shape imports no line-drift hazard. What Option 1 fails on is volume and REQ-TDN-012, not anchoring.

**Option 2 — structural narrowing only.** Rejected as a complete solution: it cannot express `DC-3` (a semantic deadline has no distinguishing syntax) or `DC-2b` (a curatorial judgement). Adopted as the *first half* of Option 4.

**Option 3 — semantic marker in source text** (e.g. a trailing `<!-- neutral-date: attribution -->`). Rejected: it pollutes the distributed template with harness-internal markup, which is itself an isolation-doctrine concern.

---

## §B Decision 2 — DC-1 disposition — SETTLED: preserve as authored

48 rows are skill and rule frontmatter fields:

```
  updated: "2026-03-30"
```

`.claude/rules/moai/development/skill-authoring.md` documents `updated:` as an ISO-date string field of the skill frontmatter schema.

**Decision: preserve as-is.** No schema change, no render-time neutralization. Binding as REQ-TDN-009.

Two alternatives were live and are now closed:

- *Preserve the key, neutralize the value at render time.* Would require touching the render path for `.md` files that are not currently templated — a far larger change than the neutrality problem warrants.
- *Remove the field from the schema.* Out of scope; that is a skill-authoring schema change, not a neutrality change (explicit in `spec.md` §5).

**Scope correction (measured).** The `DC-1` rule matches the frontmatter key `updated:` only. A `grep` for dated `created:` frontmatter lines returns `0`, and for dated `version:` lines `0`. Earlier prose describing `DC-1` as `updated:` / `created:` / `version:` overstated the rule's reach; `spec.md` REQ-TDN-001 and `plan.md` now describe it as `LS-FM` = the `updated:` key.

**Fenced-example exclusion.** One instance — `skill-authoring.md:89`, `  updated: "2026-01-28"` inside a fenced YAML example block — is *not* real frontmatter and does not carry the schema-break rationale. The classifier assigns it `LS-FM-FENCED` → `DC-5`, so the `DC-1` count is 48 rather than 49. The same file's line 45 (`- updated: ISO date as string (e.g., "2026-01-28")`) is `LS-OTHER` → `DC-5`; both rows belong to the same finding, and both are pedagogical.

### Recurring-cost note (why REQ-TDN-012 names DC-1 first)

Every skill frontmatter bump touches a `DC-1` line. Under Option 1 that would mean a Go allowlist edit per skill edit — the largest recurring tax of any option considered, and larger than the `NOTICE.md` case that is more commonly cited. The structural gate removes it entirely. `acceptance.md` AC-TDN-015 probes both cases for this reason.

---

## §C Decision 3 — CI enforcement — SETTLED: isolated target, gated on preconditions

**Decision:** add a strict-tier step to the existing `.github/workflows/template-neutrality-check.yaml`, scoped by `-run TestTemplateNoInternalContentLeak`, activated only after the finding count reaches zero. Binding as REQ-TDN-013 + REQ-TDN-015.

Preconditions, all verifiable:

| Precondition | Verified by |
|---|---|
| P1 — strict tier reports zero findings after remediation and carve-out | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` exits 0 (AC-TDN-012) |
| P2 — the carve-out satisfies REQ-TDN-012 for both DC-4 and DC-1 | synthetic future attribution line **and** synthetic frontmatter bump each produce no finding (AC-TDN-015 probes A and B) |
| P3 — narrow tier and the existing neutrality target remain green | AC-TDN-002 + AC-TDN-013 |

Only when P1-P3 hold is the workflow edited. If any fails, M6 closes as "not adopted" with the failing precondition recorded as a `precondition_failed: P<N>` line in `progress.md` — an executable check, not prose (AC-TDN-011).

### Correction: the "pre-existing unrelated failures" rationale is false

Iteration 1 justified the isolated-target convention by citing pre-existing failures in the `internal/template` package. **Measured, that claim does not hold:**

```
$ go test ./internal/template/...
package-wide exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	1.292s
```

The package is green. The claim originates in the workflow file's own in-file comment (written by an earlier SPEC) and was repeated here without verification. It is withdrawn.

The isolated-target convention still stands, on two grounds that *are* verified: (a) it is the user's settled decision (§3 of `spec.md`), and (b) it matches the existing step's shape in the same workflow (`-run TestTemplateNeutralityAudit`, confirmed present — AC-TDN-019's positive control).

**The workflow's comment is left in place, and not merely on scope grounds.** The measurement above is environment-dependent in a way this SPEC itself discovered: `output_styles_audit_test.go` resolves its root by ascending for a `.moai` marker, so a stray marker inside `internal/template/` — which the moai statusline creates in the cwd of any command run from there — makes `TestOutputStylesTemplateLiveParity` and `TestOutputStylesFallbackDocsContract` fail on nonexistent paths. A package that is green from the repo root can be red from a polluted cwd. Deleting a comment whose truth is conditional on invocation environment, on the strength of one green run, would be the larger error. The claim is withdrawn *as this SPEC's rationale*; it is not asserted false in general, and the comment stays.

### Future-legitimate-date regression risk

A CI-enforced strict tier that blocks a legitimate future entry is a net loss. P2 makes this concrete rather than aspirational: AC-TDN-015 shows that the *current* S1 pattern matches both a future `imported 2027-03-04` line and a future `updated: "2028-05-09"` line, so the hazard is measured, not hypothetical. Enforcement is not adopted until both probes pass.

---

## §D Decision 4 — Report cap — SETTLED: path-naming truncation

`limit := 50` truncates with `... 85 more (capped)`, hiding 63% of findings from anyone running the guard. The source comment justifies this by pointing at a `grep -rln` recipe in the isolation doctrine as the "real audit log". That indirection is what kept the true finding count invisible until a maintainer temporarily raised the cap by hand.

**Decision:** keep a console cap but always write the full listing to a file, and name that path in the truncation message. Binding as REQ-TDN-016.

Rejected alternatives: (a) removing the cap entirely — a 135-row failure is unreadable in a CI log; (b) making the cap an env-var-tunable constant with no file — the finding count stays invisible to anyone who does not know the env var exists.

AC-TDN-010 has two halves because the grep half alone is weak: renaming the literal would pass it. The injection recipe (append a synthetic dated line, observe the emitted path, revert) is the load-bearing half.

---

## §E Decision 5 — Local-mirror handling (mechanical, low change likelihood)

115 of the 116 affected template files have a counterpart in the local working tree. The naive move — copying the neutralized template file over the local copy, or the reverse — is prohibited: the local copies legitimately retain internal-development content that the template copies must not, and a blind copy in either direction breaks one of the two isolation policies.

The mirror-parity guard (`rule_template_mirror_test.go`) enforces byte-identity for an explicit, non-glob allowlist of rule files. Verified: none of the four date-bearing files under `templates/.claude/rules/moai/` (`NOTICE.md`, `development/spec-frontmatter-schema.md`, `development/skill-authoring.md`, `workflow/archived-agent-rejection.md`) appears in it. Byte-parity is therefore not a constraint on this remediation, and the local copies stay untouched by default (AC-TDN-020).

The one exception to check at run phase: if a remediating edit lands on a file that *is* in the mirror allowlist, that file's local copy must receive the identical edit in the same commit. AC-TDN-016 re-runs the membership check over the actual edited-file list at M3 rather than the four-file sample used for the plan-phase baseline.

---

## §F Cross-references

- `internal/template/internal_content_leak_test.go` — the guard (class definitions, `collectLeakViolations`, `isPedagogicallyAllowed`, the report cap)
- `internal/template/rule_template_mirror_test.go` — the byte-parity allowlist consulted in §E
- `.github/workflows/template-neutrality-check.yaml` — the isolated-target step convention referenced in §C
- `.moai/docs/template-internal-isolation-doctrine.md` — the C1-C8 content-class catalogue this SPEC's C3 class belongs to
- `classify.sh` (this directory) — the committed classifier implementing REQ-TDN-001
