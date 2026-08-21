# Acceptance — SPEC-SVG-QUALITY-ABSORB-001

> One falsifiable criterion per requirement, plus the ones that guard against a
> restructure quietly dropping what it was supposed to preserve.

## §D AC Matrix

| AC ID | REQ | Severity | Summary |
|---|---|---|---|
| AC-REQ-1 | REQ-1 | MUST-PASS | All six connector rules present, each with its number |
| AC-REQ-2 | REQ-2 | MUST-PASS | Every archetype maps to a node ceiling |
| AC-REQ-3a | REQ-3 | MUST-PASS | Accessible-SVG contract stated with a copyable skeleton |
| AC-REQ-3b | REQ-3 | MUST-PASS | `check-svg.mjs` fails a missing accessible name AND passes a present one |
| AC-REQ-4 | REQ-4 | MUST-PASS | 14 anti-patterns, each an observable symptom |
| AC-REQ-5 | REQ-5 | MUST-PASS | Four dials with enumerated values and a stated default each |
| AC-REQ-6a | REQ-6 | MUST-PASS | Roles carry light and dark values; inversion stated as a rule |
| AC-REQ-6b | REQ-6 | MUST-PASS | One-accent discipline survives the restructure |
| AC-REQ-7a | REQ-7 | MUST-PASS | Every listed type has both sample renders committed |
| AC-REQ-7b | REQ-7 | MUST-PASS | No type without a sample pair appears on the list |
| AC-REQ-7c | REQ-7 | MUST-PASS | Every entry is image-path-scoped; none touches locale-synced text |
| AC-REQ-8a | REQ-8 | MUST-PASS | Template mirrors updated in the same commit; `make build` clean |
| AC-REQ-8b | REQ-8 | MUST-PASS | Attribution recorded where the absorbed rules land |
| AC-REQ-9 | REQ-9 | MUST-PASS | No absorbed rule introduces a view-time asset fetch |
| AC-BUDGET | REQ-2 | SHOULD | `SKILL.md` stays within its progressive-disclosure budget |

## §D.1 Severity / Traceability

Every REQ carries at least one MUST-PASS. AC-BUDGET is SHOULD because it
constrains how the content is distributed between L2 and L3 rather than whether
the content is correct. 14 criteria against the Tier M ceiling of 16.

## §D.2 Given-When-Then

### AC-REQ-1 — six connector rules, each with a number

**Given** `references/authoring.md` §2, **when** read, **then** all six rules are
present and each carries its numeric form: elbow `r=8`, the 6–10px mask gap, the
`L·k/(N+1)` fan with its ≥12px floor, no-overlap with bridge/hop on cross, no
routing behind a non-endpoint node, and mask-must-not-overlap-a-following-node.
**Why numeric**: the sibling card (B-7) turns these into executable checks; a
rule phrased as a preference cannot be asserted.

### AC-REQ-2 — every archetype has a ceiling

**Given** the archetype list and the budget table, **when** cross-referenced,
**then** every archetype named in `archetypes.md` appears in the table with a
node ceiling, and the table introduces no archetype that does not exist.

### AC-REQ-3a — the contract is copyable, not described

**Given** the accessibility section, **when** read, **then** it carries a
skeleton showing `role`, `aria-labelledby` pointing at a prefixed ID, `<title>`
as the first child of `<svg>`, and a `<desc>` — and states the prefixed-ID rule
(IDs must be unique within a host document that may embed several diagrams).

### AC-REQ-3b — the checker fails in the right direction

**Given** an SVG with no accessible name, **when** `check-svg.mjs` runs, **then**
it exits non-zero and names the missing part. **Given** an SVG carrying the full
contract, **then** it exits zero.
**Why both**: a check that only ever passes proves nothing, and this is the one
absorbed item with an executable guard — it is also the one that can regress
silently once the docs are written.

### AC-REQ-4 — anti-patterns are observable

**Given** § Red Flags, **when** read, **then** 14 entries are present and each
names something to look for in the rendered output rather than stating a
preference. An entry a reader cannot check against a diagram fails this AC.

### AC-REQ-5 — dials have values and defaults

**Given** the Frame section, **when** read, **then** format, size, detail, and
audience each list their allowed values and state what the skill assumes when
the caller says nothing.
**Why defaults matter**: a dial with no default forces a question on every
invocation, which is how a four-dial contract becomes four extra round trips.

### AC-REQ-6a — roles, both modes, one rule

**Given** the palette section, **when** read, **then** each semantic role carries
a light and a dark value, and the inversion is expressed as a rule rather than a
second hand-maintained table.
**Why a rule**: two tables drift; the source's own inversion (same alphas, RGB
flipped) is stateable in one sentence.

### AC-REQ-6b — the one-accent discipline survived

**Given** the restructured section, **when** searched, **then** the focal
discipline still forbids a second accent.
**Why a separate AC**: the restructure rewrites the section that currently holds
this rule. The rule disappearing would not fail AC-REQ-6a, and the resulting
diagrams would look plausible while losing the single thing that makes the focal
point read.

### AC-REQ-7a — every listed type has its evidence

**Given** the final exception list, **when** each entry is checked, **then** both
a mermaid render and an absorbed-rule render exist under
`.moai/reports/t165/samples/` and the entry references them.

### AC-REQ-7b — nothing on the list without a pair

**Given** the candidate set considered, **when** compared against the final list,
**then** no type appears on the list without a committed sample pair.
**An empty list passes this AC.** If no comparison showed a decisive difference,
the correct outcome is no carve-out and an unchanged routing table.

### AC-REQ-7c — the carve-out stays in its lane

**Given** each entry, **when** read, **then** it is scoped to the image-output
path, and no entry alters how locale-synced text is authored. The docs-site
4-locale path is verified unchanged.

### AC-REQ-8a — mirrors move with the source

**Given** the run-phase diff, **when** filtered to `.claude/skills/`, **then**
every changed file has a corresponding change under
`internal/template/templates/`, and `make build` exits zero.

### AC-REQ-8b — attribution is present

**Given** the files carrying absorbed rules, **when** searched, **then**
attribution to `cathrynlavery/diagram-design` v2.6.1 (MIT) is present.
**Why MUST-PASS**: MIT's attribution duty is a licence obligation. Omitting it is
a defect of a different kind from a missing rule.

### AC-BUDGET — SKILL.md stays inside its budget

**Given** `SKILL.md` after the additions, **when** measured, **then** it remains
within the progressive-disclosure budget for an L2 body, with A-2's table and
A-4's detail living at L3 and the L2 body carrying the rule plus a pointer.

### AC-REQ-9 — the no-external-asset contract stays checkable

**Given** the changed skill files, **when** searched for a view-time fetch —
`fonts.googleapis.com`, `fonts.gstatic.com`, `@import`, and any `http`-scheme
`url(...)` outside a fragment reference — **then** there are zero matches.
**Why an AC and not just an exclusion**: a prohibition nothing tests drifts. The
absorbed source uses a font CDN, so the rule most likely to arrive by copy is
exactly the one this skill forbids.

## §D.3 Residual Risk

- **"Better" in REQ-7 is a human judgement.** The AC requires two artifacts to
  exist and be referenced; it cannot require that the judgement between them was
  correct. A wrong carve-out is durable — every future diagram of that type
  becomes an image with an image's maintenance cost.
- **The 14 anti-patterns are counted, not validated.** AC-REQ-4 checks that each
  is phrased observably; whether the set is the *right* 14 rests on the source's
  authority, not on measurement here.
- **`check-svg.mjs` checks structure, not usefulness.** A `<desc>` reading
  "diagram" satisfies the contract mechanically while telling a screen-reader
  user nothing. The AC catches absence, not vacuity.
- **Complexity budgets are inherited numbers.** They come from the source's
  per-type specs; this SPEC does not re-derive them against MoAI's own
  archetypes, so a ceiling may be tighter or looser than this skill warrants.
