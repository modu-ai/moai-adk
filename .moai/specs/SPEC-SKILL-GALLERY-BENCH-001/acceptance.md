# acceptance.md — SPEC-SKILL-GALLERY-BENCH-001

Verification layer. Every entry is `AC-XXX` labeled Given-When-Then and
binary-testable. The GEARS obligation lives in `spec.md` §D.1 (REQ-SGB-*); this file
does not restate requirements — it states the observable evidence that closes them.

## §D AC Matrix

### AC-001 — coverage (REQ-SGB-001)

**Given** the run-phase branch `WT-skillstead-gallery`
**When** reading `.moai/reports/t272/verdict.md`
**Then** exactly 9 verdict rows exist, one per TypePack form named in spec.md §D.2,
with no form skipped, deferred, or substituted.

Binary check: row count == 9 AND the form-name set equals the §D.2 set exactly AND
every row carries a non-empty settled-dials field (REQ-SGB-004; a row missing the
dial field = FAIL).

### AC-002 — artifact gate G1 (REQ-SGB-003)

**Given** the verdict table
**When** a row's verdict is `PRODUCIBLE` or `PARTIAL`
**Then** `.moai/reports/t272/artifacts/<form>.svg` exists and is non-empty.

Binary check: `test -s` per listed form.

### AC-003 — lint and render gates G2/G3 (REQ-SGB-003)

**Given** a `PRODUCIBLE` row
**When** reading `.moai/reports/t272/logs/<form>-lint.txt` and
`logs/<form>-render.txt`
**Then** each log follows the spec.md §D.2 log-format contract (command line +
verbatim output + explicit `exit=N` line); the lint log shows `check-svg.mjs`
`exit=0` with zero errors, and the render log shows `render.mjs` `exit=0` with the
PNG dimensions matching the 2x target and the browser executable + version
disclosed; and the row's settled-dials field records the four dials the artifact
was generated with (REQ-SGB-002).

Binary check: logs exist and each contains all three log-format parts, the
`exit=N` lines read `exit=0`, the settled-dials field is non-empty, and every
warning in the lint log has a triage note in the verdict row or log.

### AC-004 — failure taxonomy (REQ-SGB-005)

**Given** a row whose verdict is `PARTIAL` or `NOT-PRODUCIBLE`
**When** reading the row
**Then** it carries exactly one classification — `preset-gap` or `structural-limit` —
and at least one piece of observable evidence (a failed gate log, an archetype
mismatch observation, or a documented split-into-two-diagrams outcome) justifying it.

Binary check: every non-`PRODUCIBLE` row has classification + evidence citation.

### AC-005 — deviation naming (REQ-SGB-006)

**Given** a `PRODUCIBLE` or `PARTIAL` row
**When** the artifact is an equivalent-but-not-identical form
**Then** the row names the deviation in one sentence.

Binary check: rows flagged as equivalent-form-with-deviation carry a non-empty
deviation sentence (the absence of a deviation flag also satisfies this AC).

### AC-006 — evidence location (REQ-SGB-007)

**Given** all benchmark evidence
**When** listing `.moai/reports/t272/`
**Then** every artifact, log, and the verdict table resolve under that directory and
no evidence path cited by the docs deliverable points elsewhere.

Binary check: `verdict.md` path citations all start with `.moai/reports/t272/`.

### AC-007 — skill immutability (REQ-SGB-008)

**Given** the branch after M1 and after M2
**When** running `git diff --stat origin/main -- .claude/skills/moai-domain-svg-infographic/ internal/template/templates/.claude/skills/moai-domain-svg-infographic/`
**Then** the diff is empty.

Binary check: command output is empty at every commit point.

### AC-008 — docs grounding (REQ-SGB-009)

**Given** the docs emphasis content in all 4 locales
**When** reading each listed diagram type
**Then** each cites the evidence path of a `PRODUCIBLE` artifact; the docs and README
emphasis lists `PRODUCIBLE` forms only — `PARTIAL` and `NOT-PRODUCIBLE` outcomes are
recorded exclusively in the card evidence (`.moai/reports/t272/` verdict), never
surfaced in user-facing docs.

Binary check: cited paths exist AND every cited form's verdict row reads `PRODUCIBLE`.

### AC-009 — docs 4-locale parity (REQ-SGB-010)

**Given** the docs-site change
**When** running the 4-locale file-existence and per-page section-count parity checks
of Skill("hns-oss-docs-verify") §4
**Then** no NEW divergent page appears (ratchet clean) and
`docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md` all carry the emphasis.

Binary check: `comm -23 /tmp/parity-now.txt /tmp/parity-base.txt` output is empty.

### AC-010 — README chain (REQ-SGB-011)

**Given** the README change
**When** running `grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md`
**Then** counts are identical across the 4 files, the switcher header is intact in
all 4, and the touched section's H3 counts match.

Binary check: identical grep counts + switcher header present in all 4 files.

### AC-011 — verify recipe (REQ-SGB-012)

**Given** the full docs change set
**When** executing Skill("hns-oss-docs-verify") checks §1-§6
**Then** `build-clean` (warning-free hugo build + sitemap), URL blacklist, Mermaid
TD-only, body-emoji scan, and version-sync all pass; `locale-parity` per AC-009.

Binary check: each check's stated expected output observed; any FAIL blocks.

### AC-012 — environment preflight blocker (REQ-SGB-013)

**Given** a run environment lacking Node 18+ or a headless Chromium-family browser
**When** the preflight of plan.md §C executes
**Then** the run stops with a blocker report and no gate claim is emitted — no
evidence log in the spec.md §D.2 log-format contract exists under `logs/`.

Binary check: in the blocker scenario, `logs/` contains no log-format-contract file
and `verdict.md` contains no gate results (this AC is vacuously satisfied when the
environment passes preflight).

## §D.1 Severity

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-001, AC-003, AC-004, AC-008, AC-011 | BLOCKING | A benchmark that under-measures or docs that over-claim are the two failure modes this SPEC exists to prevent |
| AC-002, AC-006, AC-007, AC-009, AC-010 | BLOCKING | Evidence integrity and parity contracts are HARD rules |
| AC-005, AC-012 | SHOULD-FIX | Completeness of deviation naming; vacuous unless triggered |

## §D.2 Edge cases

- A form passes G2 but its PNG render (G3) fails on a browser quirk → verdict
  `PARTIAL` only if the SVG itself is complete; otherwise `NOT-PRODUCIBLE` with the
  render log as evidence.
- Two forms resolve to the same archetype with near-identical output → both still get
  independent artifacts and logs; no reuse of one form's evidence for another.
- The lint emits warnings that survive PNG inspection → recorded as triaged-noise per
  the skill's warning contract; not a gate failure.
- All nine forms come out `PRODUCIBLE` → docs emphasis lists all nine with citations;
  the taxonomy simply has no entries (valid outcome).
- Zero forms `PRODUCIBLE` → docs emphasis is not added at all; the deliverable
  degrades to the verdict table + a blocker report to the orchestrator (docs
  emphasis without measured forms would violate REQ-SGB-009).

## §D.3 Quality gates

- `hns-oss-docs-verify` recipe: `build-clean` 1.0, `locale-parity` 1.0 (must_pass).
- No commit touches the skill directories (AC-007).
- Conventional commits referencing `SPEC-SKILL-GALLERY-BENCH-001`; card id t272 in
  the PR title.

## §D.4 Definition of Done

All BLOCKING ACs pass with observed evidence; verdict table committed on the branch;
docs + README changes committed in the same PR; verify recipe results recorded.

## §D.5 Traceability

| REQ | ACs |
|-----|-----|
| REQ-SGB-001 | AC-001 |
| REQ-SGB-002 | AC-003 (dial record in verdict row) |
| REQ-SGB-003 | AC-002, AC-003 |
| REQ-SGB-004 | AC-001 (verdict table shape) |
| REQ-SGB-005 | AC-004 |
| REQ-SGB-006 | AC-005 |
| REQ-SGB-007 | AC-006 |
| REQ-SGB-008 | AC-007 |
| REQ-SGB-009 | AC-008 |
| REQ-SGB-010 | AC-009 |
| REQ-SGB-011 | AC-010 |
| REQ-SGB-012 | AC-011 |
| REQ-SGB-013 | AC-012 |

## §D.6 Indirect verification

Skills' internal test suites (`scripts/test-check-svg.mjs` over 42 fixtures) already
pin the lint both directions; the benchmark does not re-verify the skill's own
tooling — it consumes the gates as delivered.

## §D.7 Forward-looking checks

- If a follow-up card adds presets for `preset-gap` forms, re-run this SPEC's
  protocol for those forms only and update the docs emphasis accordingly.
