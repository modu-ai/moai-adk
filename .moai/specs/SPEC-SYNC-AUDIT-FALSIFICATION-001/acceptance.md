# acceptance.md — SPEC-SYNC-AUDIT-FALSIFICATION-001

> Tier M plan-phase artifact. Given-When-Then acceptance criteria for REQ-SAF-001 / REQ-SAF-002 / REQ-SAF-003 (the three IMPs) plus the template-mirror and neutrality gates.

## §D AC Matrix

| AC ID | REQ | Severity | Class | Traceability |
|-------|-----|----------|-------|--------------|
| AC-SAF-001 | REQ-SAF-001 (IMP-1) | blocking | functionality | spec.md §C REQ-SAF-001, §H AC-SAF-001 |
| AC-SAF-002 | REQ-SAF-002 (IMP-3) | blocking | functionality | spec.md §C REQ-SAF-002, §H AC-SAF-002 |
| AC-SAF-003 | REQ-SAF-003 (IMP-6) | blocking | functionality | spec.md §C REQ-SAF-003, §H AC-SAF-003 |
| AC-SAF-004 | (template mirror) | blocking | craft | spec.md §D, §H AC-SAF-004 |
| AC-SAF-005 | (template neutrality) | blocking | craft | spec.md §D, §H AC-SAF-005 |

## §D.1 Severity model

- **blocking** — must PASS for the SPEC to close.
- **optional** — reported but not closing-blocking.

All five ACs here are `blocking` (Tier M close requires all PASS).

## §D.2 Given-When-Then (binary-testable)

### AC-SAF-001 — IMP-1 AC-mechanism falsification

**Given** a fixture SPEC with a high-blast-radius AC whose stated mechanism is intentionally false (e.g. AC asserts "the cost-cap bound rejects goals exceeding the cap" but the implementation does not enforce the cap — modeled on INFINITE-GOAL-001 AC-011) AND the fixture SPEC's test suite exits 0 (vacuous pass)
**When** the sync-auditor evaluates the fixture SPEC
**Then** (a) the sync-auditor's `### Dimension Scores` Functionality row records FAIL (not PASS), AND (b) the `### Findings` list contains a blocking entry naming the AC ID, the stated mechanism, the probe input, and the observed outcome.

**Verification path**: fixture-based test. Run-phase authors a small fixture (fixture SPEC + fixture code with a deliberately-false mechanism) and runs sync-audit against it. The fixture lives under `.moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/fixtures/` (NOT in the template mirror — fixtures are repo-local dev assets per CLAUDE.local.md §2 Local-Only Files). Assertion is a grep against the sync-audit output: `grep -E 'Functionality.*FAIL'` AND `grep -E 'AC-[0-9]+.*mechanism.*false'` (or equivalent). The fixture-based run is the binary-testable evidence.

**Falsification check**: AC-SAF-001 is FALSIFIABLE — if IMP-1 is NOT implemented (sync-auditor only runs `go test ./...`), the fixture's test suite exits 0 and sync-audit records Functionality = PASS, failing the AC. This is the intended asymmetric design (spec.md §F AP-1).

### AC-SAF-002 — IMP-3 VCI §1.1 surface-3 binding for Findings

**Given** a sync-audit invocation that emits at least one entry under `### Findings`
**When** the finding is a defect / debt / drift identification claim (i.e. it asserts that something exists or is broken in the audited SPEC / code)
**Then** (a) the finding cites verbatim tool output as its Evidence (the command + the observed output, not a summary), AND (b) where tool output cannot be obtained, the finding carries an `unverified-premise` marker and is downgraded to `optional` severity (never emitted as blocking without tool output), AND (c) each blocking finding is structured per the VCI §3 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

**Verification path**: binary-testable via a grep against the sync-auditor.md body — the agent body MUST contain a normative clause binding `### Findings` emission to VCI §1.1 surface 3. Run: `grep -E 'verification-claim-integrity.md.*§1.1.*surface 3|surface 3.*verification-claim-integrity.md' .claude/agents/moai/sync-auditor.md` returns ≥1 match. (Body-level grep is the AC verifier because the obligation itself lives in the agent body — testing the body text is the correct level of verification for a doc-only obligation.)

**Falsification check**: AC-SAF-002 is FALSIFIABLE — if IMP-3 is NOT implemented (no VCI cross-reference in the body), the grep returns 0 matches and the AC fails.

### AC-SAF-003 — IMP-6 AC-class coverage minimums

**Given** a fixture SPEC whose acceptance.md §D AC matrix declares ACs across classes {functionality, security}, with the high-blast-radius AC in the security class
**When** the sync-auditor selects the Functionality-dimension sample
**Then** the sample includes (a) at least one security-class AC (the mandatory high-blast-radius one), AND (b) at least one functionality-class AC — it does NOT pass by sampling only the functionality class.

**Verification path**: binary-testable via a grep against the sync-auditor.md body — the agent body MUST contain a normative clause requiring "≥1 AC sampled per AC class present in the SPEC's acceptance.md §D AC matrix, with the high-blast-radius AC mandatory regardless of class". Run: `grep -E 'class.*sample|sample.*class' .claude/agents/moai/sync-auditor.md` returns ≥1 match. The fixture-based run is a stronger verifier but not required for a Tier-M doc-only SPEC; the body-level grep is the minimum binary verifier.

**Falsification check**: AC-SAF-003 is FALSIFIABLE — if IMP-6 is NOT implemented (no per-class sampling minimum in the body), the grep returns 0 matches and the AC fails.

### AC-SAF-004 — Template mirror byte-identity

**Given** run-phase has edited both `.claude/agents/moai/sync-auditor.md` AND `internal/template/templates/.claude/agents/moai/sync-auditor.md`
**When** the verifier runs `diff .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md`
**Then** the diff exits 0 with no output (the two files are byte-identical).

**Verification path**: `diff .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md && echo OK` — the literal command. Documented in plan.md §E E-extra-1.

**Falsification check**: trivially falsifiable — any byte divergence fails the AC.

### AC-SAF-005 — Template Content Neutrality

**Given** run-phase has authored the obligation prose in the template mirror
**When** the verifier runs `grep -E 'REQ-SAF|SPEC-SYNC-AUDIT-FALSIFICATION' internal/template/templates/.claude/agents/moai/sync-auditor.md`
**Then** the grep returns 0 matches (no SPEC IDs / REQ tokens leaked into the distributed template).

**Verification path**: the grep above + the CI guard `template-neutrality-check.yaml` (CLAUDE.local.md §25.3 5-item pre-commit self-check is the human-executed version; the CI guard is the automated safety net). Documented in plan.md §E E-extra-2.

**Falsification check**: trivially falsifiable — any SPEC-ID / REQ-token leak in the mirror fails the AC.

## §D.3 Edge cases

- **E-1 — SPEC marks NO high-blast-radius AC**: IMP-1 obligation degrades to "falsify at least one AC's stated mechanism" (still stricter than status quo). AC-SAF-001 remains testable (the fixture explicitly marks one).
- **E-2 — domain lacks a dedicated verification tool**: IMP-3 degrades per REQ-SAF-002 option (b) — `unverified-premise` marker + `optional` severity. AC-SAF-002's grep verifier still PASSes (the clause is present in the body).
- **E-3 — SPEC declares fewer than 4 canonical AC classes**: IMP-6 minimum is "≥1 AC per present class", not "per canonical class". AC-SAF-003 covers this (the fixture declares 2 classes, not 4).
- **E-4 — negative probe is not feasible for the chosen high-blast-radius AC**: REQ-SAF-001 allows a positive probe that observes the mechanism producing the asserted outcome. AC-SAF-001's fixture uses a negative probe (the more stringent form).

## §D.4 Indirect verification (tooling)

- `moai spec lint .moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/spec.md` exits 0 (or advisory-only warnings). Documented in plan.md §E E-extra-3.
- `make build` succeeds after M4 (template embedded FS regenerated). Required for the distributed binary to carry the updated agent body.

## §D.5 Closure gate (Definition of Done)

ALL five ACs (AC-SAF-001 through AC-SAF-005) PASS + plan.md §E E-extra-1..E-extra-4 all green + sync-phase CHANGELOG entry present + frontmatter `status: completed`.

## §D.6 Forward-looking checks (post-close, not blocking)

- **FL-1**: the next `/moai review` sweep on completed SPECs should show a reduced rate of PASS-WITH-DEBT that should-have-been-FAIL (the root-cause analysis's metric). Tracked post-close, not a closure gate.
- **FL-2**: IMP-2 / IMP-4 / RC5 follow-up SPECs should be linked back to this SPEC via `related_specs` when authored.

## §D.7 Traceability

| REQ | AC | GWT section | Fixture / grep verifier |
|-----|----|------------|-------------------------|
| REQ-SAF-001 (IMP-1) | AC-SAF-001 | §D.2 AC-SAF-001 | fixture under `.moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/fixtures/` (repo-local dev asset) |
| REQ-SAF-002 (IMP-3) | AC-SAF-002 | §D.2 AC-SAF-002 | body grep on `.claude/agents/moai/sync-auditor.md` |
| REQ-SAF-003 (IMP-6) | AC-SAF-003 | §D.2 AC-SAF-003 | body grep on `.claude/agents/moai/sync-auditor.md` |
| (template) | AC-SAF-004 | §D.2 AC-SAF-004 | `diff` live vs mirror |
| (template) | AC-SAF-005 | §D.2 AC-SAF-005 | `grep -E 'REQ-SAF\|SPEC-SYNC-AUDIT-FALSIFICATION' internal/template/templates/.claude/agents/moai/sync-auditor.md` returns 0 |
