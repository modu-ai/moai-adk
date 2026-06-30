# SPEC-SIMPLICITY-LADDER-002 — Acceptance Criteria

## §D. AC Matrix

| AC ID | REQ | Severity | Verification |
|-------|-----|----------|--------------|
| AC-SL2-001 | REQ-1.1 | MUST | LIVE constitution ladder is now a 7-rung ordered list (rung `7. Only then: write the minimum code that works.` present) |
| AC-SL2-002 | REQ-1.1 | MUST | The in-codebase-reuse rung is at **position 2** in LIVE (`2. Does a helper, util, type, or pattern already exist in this codebase? Reuse it.`) |
| AC-SL2-003 | REQ-1.1 | MUST | The five carried-over rungs are renumbered with per-rung CONTENT anchored (not just count): `^3\. Does the standard library`, `^4\. Does a native platform feature`, `^5\. Does an already-installed dependency`, `^6\. Can this be one line`, `^7\. Only then` — each grep == 1 in BOTH LIVE and template (guards against a content-swap that preserves numbering) |
| AC-SL2-004 | REQ-1.2 | MUST | The new rung is language-neutral (contains none of: `npm`, `pip`, `go.mod`, `package.json`, `requirements.txt`, `import`, `trait`, `interface`, or any single-language token) |
| AC-SL2-005 | REQ-1.3 | MUST | Safety carve-out paragraph (`Never simplify away …`) is byte-unchanged from the v001 baseline |
| AC-SL2-006 | REQ-1.3 | MUST | Quantitative 3x-LOC trigger (`Quantitative trigger: If implementation exceeds 3x …`) is byte-unchanged; intro line (`Simplicity decision ladder (apply in order …`) unchanged |
| AC-SL2-007 | REQ-3.1 | MUST | L247 framing prose adjusted to encompass reuse (contains `reuse-and-dependency-avoidance` or an equivalent reuse-inclusive framing); the language-neutral "standard library / native platform feature" sentence is retained |
| AC-SL2-008 | REQ-2.1 | MUST | Template mirror (`internal/template/templates/.../moai-constitution.md`) carries the SAME 7-rung ladder + framing edit |
| AC-SL2-009 | REQ-4.1 | MUST | `diff -q` clean: LIVE `moai-constitution.md` vs template mirror (byte-parity, per-file) |
| AC-SL2-010 | REQ-2.2 | MUST | `make build` succeeds (`gen-catalog-hashes --all` + `go build`); verification gate = `go test ./internal/template/...` passes AND `go build ./...` succeeds. NO `embedded.go` exists (directive embed `//go:embed all:templates`); catalog-hash regen for changed templates is an acknowledged side-effect, not the gate |
| AC-SL2-011 | REQ-4.2 | MUST | Both edited template mirrors (`moai-constitution.md` + `karpathy-quickref.md`) carry no internal token in edited content (`SPEC-SIMPLICITY-LADDER-002`, `REQ-`, `AC-`, internal date, commit SHA, `/Users/`) |
| AC-SL2-012 | (cross) | SHOULD | Irony guard: no new lint rule / hook / config knob / @MX change / Go logic introduced — pure doctrine edit |
| AC-SL2-013 | REQ-5.1/5.2 | MUST | `karpathy-quickref.md` line ~33 arrow enumeration LEADS with `in-codebase reuse` (`in-codebase reuse → stdlib → …`) AND the framing reads `reuse-and-dependency-avoidance decision ladder` (no `dependency-avoidance decision ladder`-only phrasing remains); the "capability-source ordering axis …" tail retained (LIVE + template) |
| AC-SL2-014 | REQ-5.3/4.1 | MUST | `diff -q` clean: LIVE `karpathy-quickref.md` vs its template mirror (byte-parity); ONLY line ~33 changed vs the v001 baseline (surgical — no full ladder restatement) |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Ladder is now 7 rungs with reuse at position 2 (AC-SL2-001, AC-SL2-002, AC-SL2-003)

```
GIVEN the file .claude/rules/moai/core/moai-constitution.md
  AND the section "### 4. Enforce Simplicity"
WHEN a reader reads the "Simplicity decision ladder" block
THEN the list has exactly 7 numbered rungs (1 through 7)
  AND rung 1 is YAGNI ("Does this need to be built at all?")
  AND rung 2 is the in-codebase-reuse rung
       ("Does a helper, util, type, or pattern already exist in this codebase? Reuse it.")
  AND the five carried-over rungs are anchored by CONTENT at their new numbers
       (not merely count) so a careless content-swap cannot pass
  (binary PASS, in BOTH LIVE and template:
     grep '^2\. Does a helper, util, type, or pattern already exist in this codebase' <file> == 1
     AND grep '^3\. Does the standard library' <file> == 1
     AND grep '^4\. Does a native platform feature' <file> == 1
     AND grep '^5\. Does an already-installed dependency' <file> == 1
     AND grep '^6\. Can this be one line' <file> == 1
     AND grep '^7\. Only then: write the minimum code that works' <file> == 1
     AND no '^8\.' rung exists)
```

### Scenario 2 — Carried-over rungs and safety paragraphs are preserved verbatim (AC-SL2-005, AC-SL2-006)

```
GIVEN the v001 ladder block (6 rungs + safety carve-out + 3x-LOC trigger)
WHEN the 7-rung edit is applied
THEN the safety carve-out paragraph "Never simplify away (safety carve-out): …" is byte-identical to v001
  AND the "Quantitative trigger: If implementation exceeds 3x …" paragraph is byte-identical to v001
  AND the intro line "Simplicity decision ladder (apply in order, before writing code — cheapest capability first):" is unchanged
  AND the five carried-over rung texts (standard library / native platform / installed dependency /
       one line / minimum code) are reworded only by their leading number
  (binary PASS: grep -F 'Never simplify away (safety carve-out)' <file> == 1
   AND grep -F 'If implementation exceeds 3x the estimated minimum viable LOC' <file> == 1)
```

### Scenario 3 — Framing prose encompasses reuse (AC-SL2-007)

```
GIVEN the L247 framing sentence (v001: "The ladder is the dependency-avoidance ordering axis …")
WHEN REQ-3 is applied
THEN the framing names the ladder a reuse-inclusive ordering axis
       (e.g. "reuse-and-dependency-avoidance ordering axis")
  AND the existing language-neutral sentence ("It is language-neutral: 'standard library' and
       'native platform feature' name whichever capability source the project's language provides …")
       is retained intact
  (binary PASS: grep -i 'reuse-and-dependency-avoidance' <file> >= 1
   AND grep -F "name whichever capability source the project's language provides" <file> == 1)
```

### Scenario 4 — LIVE and template mirror are byte-parallel (AC-SL2-008, AC-SL2-009)

```
GIVEN the LIVE rule and its template mirror both edited
WHEN diff -q is run on the two files
THEN the diff is clean (exit 0) — the two files are byte-identical
  (binary PASS:
     diff -q .claude/rules/moai/core/moai-constitution.md \
             internal/template/templates/.claude/rules/moai/core/moai-constitution.md ; echo $? == 0)
```

### Scenario 5 — Template mirrors stay neutral; build verifies via test+compile (AC-SL2-010, AC-SL2-011)

```
GIVEN both template mirrors edited and `make build` run
WHEN each template's edited content is scanned for internal-content leaks
THEN no internal SPEC ID, REQ/AC token, internal date, commit SHA, or /Users/ path appears in the edited content
  AND `make build` succeeds: `gen-catalog-hashes --all` regenerates the changed templates' catalog
       hashes, then `go build` embeds the live tree via `//go:embed all:templates`
       (there is NO `embedded.go` to regenerate — the embed is directive-based)
  AND the verification gate `go test ./internal/template/...` passes AND `go build ./...` exits 0
  (binary PASS:
     grep -nE 'SPEC-SIMPLICITY-LADDER-002|REQ-[A-Z0-9]|AC-SL2|/Users/' \
       internal/template/templates/.claude/rules/moai/core/moai-constitution.md == 0 matches
     AND grep -nE 'SPEC-SIMPLICITY-LADDER-002|REQ-[A-Z0-9]|AC-SL2|/Users/' \
       internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md == 0 matches
     AND go test ./internal/template/... exits 0
     AND go build ./... exits 0)
```

### Scenario 6 — karpathy-quickref cross-reference leads with in-codebase reuse (AC-SL2-013, AC-SL2-014)

```
GIVEN .claude/rules/moai/development/karpathy-quickref.md line ~33 (and its template mirror)
  AND the v001 line read: "the ordered dependency-avoidance decision ladder
       (stdlib → native platform feature → installed dependency → one line → minimum code) …"
WHEN REQ-5 is applied
THEN the arrow enumeration now LEADS with in-codebase reuse
       ("in-codebase reuse → stdlib → native platform feature → installed dependency → one line → minimum code")
  AND the framing reads "reuse-and-dependency-avoidance decision ladder"
       (no "dependency-avoidance decision ladder"-only phrasing remains)
  AND the "capability-source ordering axis that complements these LOC/abstraction checkpoint questions" tail is retained
  AND ONLY this one line changed (surgical — no full ladder restatement)
  AND LIVE and template mirror are byte-identical
  (binary PASS:
     grep -F 'in-codebase reuse → stdlib' karpathy-quickref.md == 1 (LIVE and template)
     AND grep -F 'reuse-and-dependency-avoidance decision ladder' karpathy-quickref.md >= 1
     AND grep -cF 'the ordered dependency-avoidance decision ladder (stdlib →' karpathy-quickref.md == 0
     AND diff -q <live-karpathy> <template-karpathy>; echo $? == 0)
```

## §D.2 Edge Cases

- **EC-1**: The new rung's "helper, util, type, or pattern" wording — verify it reads naturally for non-OOP languages too (e.g., a functional or systems language has "helper" and "util" and "type" and "pattern" concepts). It does — all four nouns are language-generic. No language-specific substitution needed.
- **EC-2**: Renumbering must not leave a gap or duplicate (no two rungs share a number; sequence is exactly 1..7 contiguous). Verify by `grep -oE '^[0-9]+\.' <file>` over the ladder block → `1. 2. 3. 4. 5. 6. 7.`.
- **EC-3**: The `karpathy-quickref.md` line ~33 carries an inline arrow enumeration + "dependency-avoidance" framing (§B.3) and IS edited (REQ-5). The edit is surgical (one line); verify the rest of karpathy-quickref.md is byte-unchanged (the "Simplicity First" checkpoint questions and all other lines untouched).
- **EC-4**: There is NO `internal/template/embedded.go`. The embed is directive-based (`//go:embed all:templates`, `embed.go:28`) — editing a template file is picked up automatically at the next `go build`. `make build` additionally runs `gen-catalog-hashes --all`, which regenerates the catalog hash for the changed template files; this is expected and is NOT a forbidden drift surface. Hand-editing `catalog.yaml` hashes instead of regenerating them via `make build` IS forbidden (AP-SL2-007).

## §D.3 Quality Gate Criteria

- `diff -q` BOTH file pairs → clean (exit 0): constitution LIVE vs template (AC-SL2-009) AND karpathy LIVE vs template (AC-SL2-014).
- `go test ./internal/template/...` passes (neutrality CI guard + embed/catalog-hash tests green) — AC-SL2-010/011.
- `go build ./...` exits 0 (the directive `//go:embed all:templates` embeds the edited templates at compile time; no `embedded.go` codegen).
- 7-rung grep + per-rung content anchors (Scenario 1) pass; PRESERVE grep assertions (Scenario 2) pass.
- Constitution framing-prose grep (Scenario 3) passes; karpathy enumeration-leads-with-reuse grep (Scenario 6) passes.
- Template-neutrality grep (Scenario 5) → 0 internal-token matches in BOTH template files.
- No new lint rule, hook, config file, @MX change, or Go logic introduced (AC-SL2-012 irony guard).

## §D.4 Definition of Done

- [ ] REQ-1: LIVE ladder is 7 rungs; in-codebase-reuse rung at position 2; rungs 3-7 are the renumbered v001 rungs 2-6 (AC-SL2-001/002/003).
- [ ] REQ-1.2: new rung is language-neutral (AC-SL2-004).
- [ ] REQ-1.3: safety carve-out + 3x-LOC trigger + intro line preserved verbatim (AC-SL2-005/006).
- [ ] REQ-3: framing prose adjusted to reuse-inclusive form; language-neutral sentence retained (AC-SL2-007).
- [ ] REQ-2.1: constitution template mirror carries the identical edit (AC-SL2-008).
- [ ] REQ-2.2: `make build` succeeds (`gen-catalog-hashes --all` + `go build`); gate = `go test ./internal/template/...` + `go build ./...` (no `embedded.go` regen) (AC-SL2-010).
- [ ] REQ-5: karpathy-quickref line ~33 enumeration leads with in-codebase reuse + reuse-inclusive framing, LIVE + template, surgical one-line edit (AC-SL2-013); karpathy LIVE vs template byte-parity (AC-SL2-014).
- [ ] REQ-4.1: `diff -q` clean for BOTH file pairs — constitution (AC-SL2-009) + karpathy (AC-SL2-014).
- [ ] REQ-4.2: both template mirrors' edited content carries no internal token (AC-SL2-011).
- [ ] Quality gates (§D.3) green: `go test ./internal/template/...` + `go build ./...`.
- [ ] Irony guard (AC-SL2-012): no new lint rule / hook / config knob / @MX change / Go logic.
- [ ] Status transitioned draft → in-progress (M1) → implemented → completed (sync) per the ownership matrix.

## §E. Lifecycle Audit Signals

(Plan-phase emits §E.1 only; run/sync phases populate §E.2-§E.4 in progress.md.)
