# Plan-Phase Audit — SPEC-TABSCHEMA-AUTOBRANCH-001 (card t316), iteration 1

**Verdict: PASS**
**Aggregate score: 0.86** (harmonic mean of the four dimensions) — Tier S threshold **0.75**
**Iteration: 1 / 1** (Tier S ceiling)

Reasoning context ignored per M1 Context Isolation. The dispatch's characterization of the change,
its cited coordinates, and its attribution of author-declared gaps were all treated as claims to be
re-measured, not as inputs to the verdict.

---

## 0. Tree pinning and iteration provenance

| Check | Command | Output |
|---|---|---|
| Worktree root | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316` |
| HEAD (opening) | `git rev-parse --short HEAD` | `7ed6edb3e` |
| HEAD (closing re-pin) | `git rev-parse --short HEAD` | `7ed6edb3e` |
| Branch | `git branch --show-current` | `WT-tabschema-autobranch` |
| Prior reports | `ls -la .moai/reports/t316/` | `ls: .moai/reports/t316/: No such file or directory` |
| Tree state | `git status --porcelain` | `?? .moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001/` (only) |

This is genuinely iteration 1 — the report directory did not exist. No Tier-ceiling overrun.

**No mid-audit artifact drift.** The four SPEC files carried identical size and mtime at the opening
and closing `ls`, and `git status` showed no tracked-file modification at either end. `tab_schema.json`
(both copies) is byte-unmodified against HEAD. No process defect observed.

**Read-only mandate honoured.** This report is the only file created. No SPEC artifact, no source
file, and no `tab_schema.json` copy was modified. No `make build`, no `go test ./...`, no
integration-lock mutation was run.

---

## 1. Must-pass firewall

| # | Criterion | Result | Evidence |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | `spec.md:L53-L82` — REQ-TSA-001…009, sequential, zero-padded to 3, no gap, no duplicate. |
| MP-2 | GEARS format (requirement layer) | **PASS** | See §2. Judged against the nine `REQ-TSA-*` entries in `spec.md` §2 only; the `AC-TSA-*` Given-When-Then entries in `acceptance.md` are the verification layer and were graded under §4, never here. |
| MP-3 | YAML frontmatter validity | **PASS** | `spec.md:L1-L14` — all 12 canonical fields present, correct types, no rejected snake_case alias. Field-by-field in §3. |
| MP-4 | Language neutrality | **N/A (auto-pass)** | The SPEC scopes one JSON interview schema and one Go type file. It names no language-specific tooling and makes no multi-language claim, so the criterion is not executable. |
| MP-5 | D7 cross-SPEC reconciliation | **PASS — no BLOCKING finding** | `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md \| sort -u` → `SPEC-TABSCHEMA-AUTOBRANCH-001` (self only). The verb is executable and returns no external reference, so no `retired/superseded/archived` condition can be met. **However** an *unreferenced* in-progress sibling overlap was found — see D4; it is a disclosure defect, not a D7 BLOCKING, and it is reported at SHOULD-FIX rather than folded in at critical, because the D7 verb's trigger conditions are genuinely not met. |
| MP-6 | D8 cross-platform discipline | **PASS (auto)** | `grep -c 'syscall' spec.md` → `0`. D8-4 auto-PASS. |
| MP-7 | Clarification gate | **PASS** | `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001/` → rc=1, no output. (No `research.md` exists — Tier S; `plan.md` exists and is clean, so the criterion is executable rather than N/A.) |

No must-pass failure. No BLOCKING finding. The firewall does not force FAIL.

---

## 2. MP-2 detail — GEARS, per requirement

| REQ | Declared pattern | Verdict |
|---|---|---|
| REQ-TSA-001 | Ubiquitous | HOLDS — "The `tab_schema.json` interview definition **shall** bind…" |
| REQ-TSA-002 | Ubiquitous | HOLDS — "**shall not** contain…" is the GEARS canonical Unwanted negative form. The label says Ubiquitous; the pattern is Unwanted. Both are valid GEARS; the label mismatch is cosmetic and is not scored as a defect. |
| REQ-TSA-003 | State-driven | HOLDS — "**While** the operator's selected `git_strategy.mode` is `personal`, … **shall** present…" |
| REQ-TSA-004 | State-driven | HOLDS — same shape, `team`. |
| REQ-TSA-005 | Ubiquitous | HOLDS — "The change **shall** leave…". Generalized `<subject>` = "The change" (a noun), permitted. |
| REQ-TSA-006 | Ubiquitous | HOLDS |
| REQ-TSA-007 | Where (capability gate) | HOLDS — "**Where** the repository enforces the Template-First rule, … **shall** be applied…" |
| REQ-TSA-008 | Ubiquitous | HOLDS |
| REQ-TSA-009 | When (event-driven) | HOLDS — "**When** template neutrality is evaluated…, the change **shall** have introduced…" |

`grep -c 'IF\|THEN'` deprecation check: no `IF/THEN` modality anywhere in the requirement layer.
Nine of nine match a GEARS pattern.

---

## 3. MP-3 detail — frontmatter, field by field

`spec.md:L1-L14`:

| Field | Value | Type check |
|---|---|---|
| `id` | `SPEC-TABSCHEMA-AUTOBRANCH-001` | string ✓, matches directory name ✓ |
| `title` | quoted string | ✓ |
| `version` | `"0.1.0"` | quoted semver ✓ |
| `status` | `draft` | valid enum ✓ |
| `created` | `2026-08-27` | ISO date ✓ |
| `updated` | `2026-08-27` | ISO date ✓ |
| `author` | `manager-spec` | string ✓ |
| `priority` | `P2` | valid enum ✓ |
| `phase` | `"v3.1.3 target"` | string ✓ |
| `module` | `.claude/skills/moai-workflow-project/schemas` | string ✓ |
| `lifecycle` | `spec-anchored` | valid enum ✓ |
| `tags` | comma-separated string | ✓ |

No `created_at` / `updated_at` / `labels` / `spec_id` alias present. `grep -n 'tier' spec.md plan.md`
returned **no match in either file** — see D5.

---

## 4. Coordinate re-measurement (audit target 1)

Every figure below was measured by this audit on `7ed6edb3e`, not carried from the SPEC.

### 4.1 The nine `auto_branch` occurrences

`grep -n 'auto_branch' .claude/skills/moai-workflow-project/schemas/tab_schema.json`:

```
517:              "question": "Auto-create branches for Personal mode? (current: {{git_strategy.personal.auto_branch}})",
519:              "field": "git_strategy.personal.auto_branch",
534:              "current_value_path": "git_strategy.personal.auto_branch",
727:              "question": "Auto-create branches for Team mode? (current: {{git_strategy.team.auto_branch}})",
729:              "field": "git_strategy.team.auto_branch",
744:              "current_value_path": "git_strategy.team.auto_branch",
1006:              "question": "Auto-create a branch for each SPEC implementation? (current: {{git_strategy.{mode}.automation.auto_branch}})",
1008:              "field": "git_strategy.{mode}.automation.auto_branch",
1024:              "current_value_path": "git_strategy.{mode}.automation.auto_branch",
```

**HOLDS** — line-for-line identical to `acceptance.md` AC-TSA-002 and AC-TSA-003, including the
`1024` (not `1029`) coordinate.

### 4.2 The three batch conditions — and every other batch

Enumerated all 17 batch objects with their `condition`, question count, and `auto_branch` fields:

```
0.1   nq=3  cond=null                                                                     ab=[]
1.1   nq=3  cond=null                                                                     ab=[]
2.1   nq=3  cond=null                                                                     ab=[]
3.0   nq=1  cond=null                                                                     ab=[]
3.1   nq=4  cond={"field":"git_strategy.mode","operator":"equals","value":"manual"}       ab=[]
3.2   nq=1  cond={"field":"git_strategy.mode","operator":"equals","value":"manual"}       ab=[]
3.3   nq=4  cond={"field":"git_strategy.mode","operator":"equals","value":"personal"}     ab=['git_strategy.personal.auto_branch']
3.4   nq=3  cond={"field":"git_strategy.mode","operator":"equals","value":"personal"}     ab=[]
3.5   nq=2  cond={"field":"git_strategy.mode","operator":"equals","value":"personal"}     ab=[]
3.6   nq=4  cond={"field":"git_strategy.mode","operator":"equals","value":"team"}         ab=['git_strategy.team.auto_branch']
3.7   nq=3  cond={"field":"git_strategy.mode","operator":"equals","value":"team"}         ab=[]
3.8   nq=3  cond={"field":"git_strategy.mode","operator":"equals","value":"team"}         ab=[]
3.9   nq=2  cond={"field":"git_strategy.mode","operator":"equals","value":"team"}         ab=[]
3.10  nq=3  cond={"field":"git_strategy.mode","operator":"not_equals","value":"manual"}   ab=['git_strategy.{mode}.automation.auto_branch']
4.1   nq=2  cond=null                                                                     ab=[]
4.2   nq=4  cond=null                                                                     ab=[]
5.1   nq=3  cond=null                                                                     ab=[]
TOTAL batches 17
```

**HOLDS.** All three conditions are exactly as `spec.md` §1 states. Exactly three batches carry an
`auto_branch` question. **Every** condition in the file targets `git_strategy.mode`; there is no
condition on any other field and no operator other than `equals` / `not_equals`.

### 4.3 Struct-tag facts (`internal/config/types.go`)

```
type ModeProfile struct {
	Workflow          string `yaml:"workflow"`
	...
	Automation     AutomationConfig     `yaml:"automation"`
	BranchCreation BranchCreationConfig `yaml:"branch_creation"`
	CommitStyle    CommitStyleConfig    `yaml:"commit_style"`
	Hooks          HooksConfig          `yaml:"hooks"`
}

type AutomationConfig struct {
	AutoBranch bool `yaml:"auto_branch"`
	...
}

type GitStrategyConfig struct {
	...
	Manual   ModeProfile `yaml:"manual"`
	Personal ModeProfile `yaml:"personal"`
	Team     ModeProfile `yaml:"team"`
	// Deprecated: use ActiveModeProfile().Automation.AutoBranch instead.
	AutoBranch bool `yaml:"auto_branch"`      // types.go:158
	...
}
```

All three claims **HOLD**, and the three paths are correctly kept distinct:

1. `ModeProfile` carries **no** `auto_branch` field (verified by full struct body, not by grep) ⇒
   `git_strategy.personal.auto_branch` and `git_strategy.team.auto_branch` bind to nothing. Dead.
2. `AutomationConfig.AutoBranch` at `types.go:58` carries `yaml:"auto_branch"`, reached through
   `ModeProfile.Automation` (`yaml:"automation"`) ⇒ `git_strategy.{mode}.automation.auto_branch` is
   canonical. Live.
3. `GitStrategyConfig.AutoBranch` at `types.go:158` is a **bound but deprecated** top-level flat key.
   Live, legacy, and correctly excluded from scope by `spec.md:L116-117`.

The SPEC's insistence that the three not be conflated is correct and is itself correctly applied.

### 4.4 Pair identity, occurrence counts, JSON validity

| Measurement | Command | Output |
|---|---|---|
| pair identity | `diff -q LOCAL TMPL` | (empty) `diff_rc=0` |
| byte size | `ls -la` | both `52169` |
| `auto_branch` LOCAL | `grep -c 'auto_branch' LOCAL` | `9` |
| `auto_branch` TMPL | `grep -c 'auto_branch' TMPL` | `9` |
| dead personal | `grep -c 'git_strategy.personal.auto_branch' LOCAL` | `3` |
| dead team | `grep -c 'git_strategy.team.auto_branch' LOCAL` | `3` |
| canonical | `grep -c 'git_strategy.{mode}.automation.auto_branch' LOCAL` | `3` |
| JSON LOCAL | `python3 -m json.tool LOCAL > /dev/null` | `local_rc=0` |
| JSON TMPL | `python3 -m json.tool TMPL > /dev/null` | `template_rc=0` |
| total questions | AC-TSA-007 recipe | `TOTAL_QUESTIONS = 48` |

All **HOLD**. The independent per-batch sum (3+3+3+1+4+1+4+3+2+4+3+3+2+3+2+4+3) = 48 corroborates
the recipe. Per-batch baseline 3.3=4, 3.6=4, 3.10=3 **HOLDS** (`acceptance.md:L194`).

### 4.5 RED-now reproduction (independent of the SPEC's figures)

AC-TSA-001's recipe, run verbatim by this audit:

```
LOCAL                                    TMPL
personal=2                               personal=2
team=2                                   team=2
manual=0                                 manual=0
```

**HOLDS** — the SPEC's declared RED-now baseline is reproduced exactly on both copies. Not accepted;
re-measured.

---

## 5. AC-TSA-001 — adjudicating the `mode_admits` predicate (audit target 2)

The predicate is a reconstruction, so it was tested against the schema's actual condition semantics
rather than read.

**Faithful on every case the schema actually contains.**

| Predicate branch | Schema reality | Verdict |
|---|---|---|
| no `condition` ⇒ admits all | 7 batches have `condition: null`; none carries `auto_branch` | correct, and inert for this count |
| `condition.field != git_strategy.mode` ⇒ admits all | **no such condition exists** — all 10 conditions target `git_strategy.mode` | vacuously correct |
| `equals` ⇒ `m == value` | 9 batches; matches the schema's own `implementation_notes.conditional_batches` = *"Batches in Tab 3 show/hide based on git_strategy.mode selection"* | correct |
| `not_equals` ⇒ `m != value` | 1 batch (3.10) | correct |
| any other operator ⇒ admits (fallthrough) | **no other operator exists** in the file | vacuously correct |

**The one real incompleteness — and why it does not change a single counted value.** Tabs carry a
gate the predicate never inspects: `mode_condition`, a *tab-level* key distinct from the batch-level
`condition`. Measured:

```
tab_0_initialization  mode_condition="INITIALIZATION"  batches=['0.1']
tab_1_user_language   mode_condition=null              batches=['1.1']
tab_2_project_info    mode_condition=null              batches=['2.1']
tab_3_git_strategy    mode_condition=null              batches=['3.0'…'3.10']
tab_4_quality_reports mode_condition=null              batches=['4.1','4.2']
tab_5_system          mode_condition=null              batches=['5.1']
```

Two facts defuse it. First, `tab_3_git_strategy` — which owns every batch that matters — has
`mode_condition: null`, so no outer gate applies to any counted batch. Second, the only non-null
value, `"INITIALIZATION"`, is drawn from a different vocabulary entirely
(`navigation_flow.mode_specific_order` = `INITIALIZATION` / `SETTINGS`), i.e. an interview-flow gate,
not a `git_strategy.mode` gate — and its tab's sole batch `0.1` carries no `auto_branch` question.

**Adjudication: the predicate is a faithful reading for the question it is asked, and AC-TSA-001 is
sound as the primary criterion.** Its omission of `mode_condition` is a latent generality gap that
changes no value on this schema; it is reported at MINOR (D6), not as a defect in the criterion.

---

## 6. Mutant sweep (audit target 3)

Each mutant traced to the criterion that kills it, measured against the AC set as written.

| # | Mutant | Killed by | Verdict |
|---|---|---|---|
| a | delete only one of the two dead questions | **AC-TSA-001** → `personal=1, team=2` ≠ exact tuple; also AC-TSA-002 (`team_dead=3`≠0), AC-TSA-004 (`6`≠3), AC-TSA-007 (`47`≠46) | **caught, 4× redundant** |
| b | delete both dead **and** the canonical 3.10 question (over-deletion) | **AC-TSA-001** → `personal=0`; AC-TSA-003 (`0`≠3); AC-TSA-004 (`0`≠3); AC-TSA-007 (`45`≠46) | **caught, 4×** |
| c | rebind instead of delete (two live duplicates) | **AC-TSA-001** → `personal=2` (both objects now sit in admitting batches with a canonical field); AC-TSA-004 (`9`≠3); AC-TSA-007 (`48`≠46) | **caught.** AC-TSA-002 alone would *not* catch it — it reads `0` after a rebind. AC-TSA-001 is doing the work here, exactly as `plan.md` §E claims. |
| d | edit the local copy only (template reverted by next `moai update`) | **AC-TSA-005** `diff -q` rc≠0 | **caught** |
| e | edit template, **skip `make build`** (stale embedded asset) | **nothing in AC-TSA-001…008.** AC-TSA-005 compares the two *source* files; if M3 syncs local to template, it passes with the binary's embedded asset still stale. Caught only by the `Definition of Done` bullet *"`make build` run after the template edit, with its exit code recorded"* | **GAP → D2** |
| f | dangling trailing comma ⇒ invalid JSON | **AC-TSA-006** `json.tool` rc≠0 on the affected copy | **caught** |
| g | satisfy `grep -c == 3` by deleting one correct object plus an unrelated 3-occurrence block | **AC-TSA-001** (`team=2`) + **AC-TSA-002** (`team_dead=3`≠0); AC-TSA-004 alone is insufficient, which `plan.md` §E states outright | **caught** |
| h | silently **remove** some other question | **AC-TSA-007** → `45`≠46 | **caught** |
| h′ | silently **alter** some other question (adjacent mutant, not in the dispatch list) | **nothing.** All eight criteria pass: counts unchanged, greps unchanged, JSON valid, pair identical, neutrality unchanged. `REQ-TSA-006`'s second clause — *"alter no other question object"* — has no verifying criterion | **GAP → D1** |

Seven of the eight enumerated mutants die, most of them several times over. The AC set is genuinely
strong. Two escapes: (e) survives the numbered criteria but is caught by the DoD; (h′) survives the
whole artifact set.

---

## 7. Trailing-comma boundary claim (audit target 4)

The SPEC derived this by reading. It was verified by executing.

**Batch 3.3** — `sed -n '495,545p' LOCAL`, absolute line numbers restored:

```
515            },                                            ← preceding element's closing brace + comma
516            {                                             ← auto_branch object opens
517              "question": "Auto-create branches for Personal mode? …
519              "field": "git_strategy.personal.auto_branch",
534              "current_value_path": "git_strategy.personal.auto_branch",
535              "required": true
536            }                                             ← closes, NO trailing comma
537          ]                                               ← questions array closes
538        },
539        {
540          "batch_id": "3.4",
```

**Batch 3.6** — `sed -n '705,755p' LOCAL`:

```
725            },
726            {                                             ← auto_branch object opens
727              "question": "Auto-create branches for Team mode? …
729              "field": "git_strategy.team.auto_branch",
744              "current_value_path": "git_strategy.team.auto_branch",
745              "required": true
746            }                                             ← closes, NO trailing comma
747          ]
748        },
749        {
750          "batch_id": "3.7",
```

**HOLDS at both sites, on every sub-claim:**

- The object is the **final** element of its `questions` array (`]` immediately follows the closing brace).
- The line spans `516-536` and `726-746` are exact, as `plan.md:L33-34` states.
- The preceding element is `git_strategy.{personal,team}.github_integration` at both sites, as
  `plan.md:L37` states, and it is the `},` at L515 / L725 that must lose its comma.
- Each deleted object carries exactly **three** `auto_branch` occurrences (`question` / `field` /
  `current_value_path` — L517/519/534 and L727/729/744), which is the arithmetic AC-TSA-004's
  `9 − 6 = 3` derivation rests on. The derivation **HOLDS**.

The deletion recipe is correct.

---

## 8. AC-TSA-008 — is the delta-form neutrality assertion sound? (audit target 5)

`grep -n 'SPEC-\|20[0-9][0-9]-' TMPL`:

```
3:  "schema_updated": "2025-12-22",
625:              "question": "Branch prefix for Personal mode? … (e.g., feature/SPEC-) …",
915:              "question": "Branch prefix for Team mode? … (e.g., feature/SPEC-) …",
```

**HOLDS** — exactly three hits, at exactly the cited lines, and nothing else.

**Adjudication: the delta form is sound, and an absolute `0` would have been wrong.** The three hits
predate the card; asserting `0` would fail the criterion on content this SPEC neither owns nor is
permitted to touch (`spec.md` §4 scopes the change to two question objects), which would make the
criterion unsatisfiable-without-scope-violation — the worse failure mode.

**The licensing risk the target asks about is real but small, and is closed by a sibling criterion.**
A delta stated as *"the same three lines and no others"* is line-set-identity, not line-content
identity: a new violation appended to line 3, 625, or 915 would keep the line set the same and slip
through. But all three are outside both deletion regions (3 ≪ 516; 625 and 915 sit between the two
regions), and AC-TSA-007's exact total plus the deletion-only constraint in `plan.md` §B leave no
path by which the run phase edits them. Residual, not defect. Strengthening it to compare the three
**lines verbatim** rather than their line numbers would close it at zero cost (D7, MINOR).

---

## 9. Out-of-scope honesty — the manual-mode framing (audit target 6)

The claim under test (`spec.md:L96-108`): the manual-mode silence **predates** the change, and the
deletion removes nothing a manual operator currently gets.

Measured:

- Batch 3.3 condition: `equals personal` → does **not** admit `manual`.
- Batch 3.6 condition: `equals team` → does **not** admit `manual`.
- Batch 3.10 condition: `not_equals manual` → **excludes** `manual`.
- Predicate output, both copies, before the change: `manual=0`.

**The framing is accurate.** A manual-mode operator is admitted to neither deleted batch, so the
deletion cannot take anything away from them; `manual=0` before and `manual=0` after. The SPEC's
sharper sub-claim — that a manual operator's answer today would go to the dead path anyway — is
moot in the strongest possible way: they are never asked at all. The SPEC states this correctly at
`spec.md:L98-100` and does not overclaim. Deferring the design question to a separate card is the
right call and is not a scope dodge.

---

## 10. D7-adjacent finding — an undisclosed in-progress sibling on the same file

Not required by the D7 verb (which found zero external SPEC references), but material, so it was
pursued.

`grep -rln 'tab_schema' --exclude-dir=.git .` surfaced **SPEC-SYNC-STRATEGY-KEY-001** (card t303,
`status: in-progress`, `version: "0.2.0"`), which plans edits to this exact file:

```
plan.md:62      tab_schema.json: rebind the SPEC-workflow question to
                git_strategy.{mode}.automation.auto_branch, boolean options …
acceptance.md:82 … the SPEC-branching question is rebound to
                git_strategy.{mode}.automation.auto_branch … (grep for the field ≥ 1)
```

That looked at first like a live collision with `REQ-TSA-005` ("leave batch 3.10 byte-unchanged").
It is not. Measured:

```
$ grep -n 'spec_git_workflow' LOCAL          → (no output; count 0, not the 3 t303 records)
$ git log --oneline -5 -- LOCAL
63b4628a6 chore(sync): sync local mirrors and repo-local surfaces to the canonical key (t303)
$ git merge-base --is-ancestor HEAD origin/develop   → ancestor_rc=0
```

**t303's `tab_schema.json` rebind has already landed**, at commit `63b4628a6`, which is in this
worktree's committed history and an ancestor of `origin/develop`. Batch 3.10's canonical form at
L1006/1008/1024 — with its `"note"` line at L1011 and boolean options — *is* t303's landed output.
The line drift the sibling still records (`L1029` → measured `1024`) is the fingerprint of that edit.

Three consequences, and only the third is a defect:

1. **No collision.** The baseline is committed, not borrowed from an unmerged branch.
2. **No conflict in the reverse direction either.** After this card lands, t303's own criteria still
   pass: `spec_git_workflow` stays `0`, and the canonical-field grep stays `3` (≥1). Deleting the
   personal/team dead paths touches neither.
3. **Disclosure gap.** `spec.md` presents batch 3.10 as simply the pre-existing canonical batch and
   never says that its canonical form is recent, that it is another SPEC's landed output, or that
   that SPEC is still `in-progress`. `REQ-TSA-005` is therefore load-bearing for a reason the
   document does not state — a run-phase reader tempted to "tidy" 3.10 has no way to learn from
   these artifacts that they would be reverting t303. → **D4, SHOULD-FIX.**

---

## 11. Author-declared gaps, re-checked (audit target 8)

| Declared gap | Genuinely open? | Understated? |
|---|---|---|
| `make build` not run (M2's job) | **Yes, open.** No AC covers it; `plan.md:L45` assigns it to M2 | **Understated.** Presented as a sequencing note; it is in fact the only enumerated mutant (e) that escapes all eight criteria. → D2 |
| manual `automation` block confirmed structurally, not from a rendered config | **Yes, open.** Verified here from `ModeProfile` (all three modes share the type, so `Manual` does carry `Automation`) — a struct-level, not runtime, confirmation | Fairly stated. Harmless: nothing in this card depends on the manual profile's runtime shape, and `manual=0` holds either way |
| deletion boundary read, not executed | **Now closed by this audit** — §7 executed it at both sites; every sub-claim holds | Fairly stated at authoring time |
| neutrality asserted as a delta | **Yes, open by design.** §8 adjudicates it sound | Fairly stated, with the rationale given at `acceptance.md:L222-223` |

**One further gap the author did not declare:** the artifact set carries **no Gaps / Residual-risk
section**. The four gaps above live in the dispatch prose only; `progress.md` §E.1 records positive
claims and nothing unobserved. A run-phase reader working from the artifacts alone would not
encounter them. → **D3, SHOULD-FIX.**

---

## 12. Structure, requirements quality, traceability (audit target 7)

**Sections (Group 2).** HISTORY `spec.md:L18` ✓ · WHY (§1 Context) `L24` ✓ · WHAT (§3 Decision + §5
Surfaces) `L83`/`L130` ✓ · REQUIREMENTS `L51` with 9 entries ✓ · ACCEPTANCE CRITERIA — `acceptance.md`,
8 entries (Tier S may inline, a companion file is stronger) ✓ · Out of Scope `L94` with **three**
`### Out of Scope — <topic>` H3 sub-headings, each carrying specific `-` bullets ✓ (SC-6 PASS; this
is well above the bar — the entries name what is excluded *and* why).

**Requirements quality (Group 3).** No implementation detail leaks into the requirement layer: no
function name, no line number, no library version. Line-level coordinates are correctly confined to
`plan.md` §C and `acceptance.md`, which is where they belong. Normative text uses `shall` throughout;
no `should` / `may` / "reasonable" in any REQ. RQ-1…RQ-6 all PASS.

**Acceptance quality (Group 4).** All 8 ACs are Given-When-Then — the correct verification-layer
format, graded here and not under MP-2. Every one is binary: exact integers (`personal=1`,
`TOTAL_QUESTIONS = 46`, `= 3`, `= 0`), exit codes, or empty-output assertions. `grep -in
'appropriate\|adequate\|reasonable\|proper'` finds no weasel word in the criteria. Six of the eight
carry a runnable recipe and a verbatim RED-now or GREEN-now baseline, and this audit reproduced every
one of them. AC-1…AC-3 PASS.

**Consistency (Group 6).** No two requirements contradict. §4's exclusions do not collide with §2's
inclusions — REQ-TSA-005 (3.10 unchanged) and the §4 entry "any change to batch 3.10" are the same
constraint stated from both sides, which is coherent rather than redundant. `priority: P2` is
proportionate. CN-1…CN-3 PASS.

**Traceability (Group 4, AC-4/AC-5).** Coverage is complete in both directions and the mapping is
unambiguous:

| REQ | Covered by |
|---|---|
| 001 | AC-001, AC-003, AC-004 |
| 002 | AC-002 |
| 003 | AC-001 (`personal=1`) |
| 004 | AC-001 (`team=1`) |
| 005 | AC-003 *(partially — see D8)* |
| 006 | AC-007 *(count only — see D1)* |
| 007 | AC-005 *(identity only — see D2)* |
| 008 | AC-006 |
| 009 | AC-008 |

No orphaned AC, no uncovered REQ. **But the mapping is nowhere written down**:

```
$ grep -c 'REQ-TSA' acceptance.md   → 0
$ grep -c 'AC-TSA'  spec.md          → 0
$ grep -c 'REQ-TSA' plan.md          → 0
$ grep -c 'AC-TSA'  plan.md          → 9
```

Every REQ↔AC link is inferred by a reader, not asserted by the documents. `plan.md` cites ACs nine
times and REQs zero times, so even the plan carries only half the trace. → **D9, SHOULD-FIX.**

**Tier proportionality.** Four artifacts, 9 REQs and 8 ACs for a two-object JSON deletion is at the
upper end of Tier S but **not** inflated: the criteria are cheap greps and exact counts rather than
speculative requirements, and the mutant sweep in §6 shows the redundancy is load-bearing (mutant (c)
in particular dies only to AC-TSA-001, and mutant (g) only to the AC-001+AC-002 pair). No
over-engineering finding.

---

## 13. Category scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| **Clarity** | 0.95 | 1.0 band, minus a seam | Every REQ has a single reading; no pronoun ambiguity; the three `auto_branch` paths are held rigorously distinct (`spec.md:L37-45`, confirmed §4.3); §3 states the decision *and* its rejected alternative with the failure mode spelled out (`L88-90`). Deducted only for the REQ-005 "byte-unchanged" / AC-003 grep-count seam (D8). |
| **Completeness** | 0.90 | 1.0 band, minus three | All required sections present, three Out-of-Scope H3s with specific bullets, all 12 frontmatter fields valid. Deducted for: no `tier:` field (D5), no Gaps/Residual-risk section (D3), no t303 provenance disclosure (D4). |
| **Testability** | 0.85 | 0.75-1.0 boundary | 8 binary criteria, no weasel words, verbatim baselines this audit reproduced exactly, 7 of 8 enumerated mutants killed. Deducted for the two verification gaps: `make build` (D2) and alteration-vs-removal (D1), plus the AC-003 weakening (D8). |
| **Traceability** | 0.75 | 0.75 band | Semantic coverage is complete both directions with no orphan and no uncovered REQ — but *every* link is indirect: zero REQ citations in `acceptance.md`, zero AC citations in `spec.md`. The rubric's 0.75 band is "mapping is indirect"; here that is the general case, not the exception. (D9) |

**Aggregate = harmonic mean = 4 / (1/0.95 + 1/0.90 + 1/0.85 + 1/0.75) = 0.856 → 0.86.**

Harmonic mean is used deliberately per the skeptical-evaluation stance (it refuses to let the strong
Clarity score paper over the weakest dimension). The arithmetic mean would read 0.86 as well, so the
verdict is not sensitive to the choice.

**0.86 > 0.75 (Tier S threshold).**

---

## 14. Defects Found

```
D1. AC-COVERAGE-ALTERATION — acceptance.md:L164-195 (AC-TSA-007) vs spec.md:L68-69 (REQ-TSA-006)
    — REQ-TSA-006 requires the change "alter no other question object", but no criterion verifies
      alteration; AC-TSA-007 verifies only the question *count*. Mutant (h′) — silently editing an
      untouched question's text, options, or header — passes all eight criteria: counts hold, greps
      hold, JSON parses, the pair stays identical, neutrality is unchanged.
    — Severity: major — Class: blocking
    — Required fix: add a ninth criterion asserting the shape of the diff itself, e.g.
      `git diff --numstat -- <TMPL> <LOCAL>` shows zero added lines and exactly the expected removed
      count at each of the two sites, or `git diff -U0` touches only the two named line ranges.

D2. AC-COVERAGE-MAKE-BUILD — spec.md:L71-74 (REQ-TSA-007) vs acceptance.md:L134-144 (AC-TSA-005)
    — REQ-TSA-007 requires the template edit be "regenerated with `make build`", but AC-TSA-005
      verifies only source-file byte-identity. Mutant (e) — template edited, local synced, `make
      build` skipped — passes AC-TSA-005 while leaving the binary's embedded asset stale. Caught
      only by the Definition of Done bullet at acceptance.md:L231, which is a checklist item rather
      than a measured criterion. This is the one escaping mutant that reaches a shipped artifact.
    — Severity: major — Class: blocking
    — Required fix: promote the DoD bullet to a numbered criterion — record `make build` exit code 0
      AND assert the regenerated binary carries the edited asset (e.g. `go build` succeeds and a
      `strings`/extraction check on the built binary returns 3, not 9, `auto_branch` occurrences).

D3. NO-GAPS-SECTION — progress.md:L8-19 (§E.1)
    — The artifact set carries no Gaps / Residual-risk section. §E.1 records only positive claims
      ("SPEC ID regex self-check → PASS", "all eight criteria carry a measurement"). The four
      author-declared gaps confirmed open in §11 of this report exist nowhere in the artifacts, so a
      run-phase reader working from the SPEC directory alone never encounters them.
    — Severity: minor — Class: blocking
    — Required fix: add a `§E.1 Gaps` subsection to progress.md listing the four confirmed-open
      items (make-build not yet run; manual `automation` confirmed structurally not from a rendered
      config; neutrality asserted as a delta; and now-closed: deletion boundary, executed by this
      audit §7).

D4. UNDISCLOSED-SIBLING-PROVENANCE — spec.md:L35, L65-66, L115
    — Batch 3.10 is presented as the pre-existing canonical batch with no indication that its
      canonical form is recent, that it is SPEC-SYNC-STRATEGY-KEY-001's landed output (commit
      63b4628a6, "(t303)"), or that t303 remains `status: in-progress`. REQ-TSA-005 is therefore
      load-bearing for a reason the artifacts do not state; a run-phase agent tempted to normalize
      3.10 has no way to learn it would be reverting another card. Verified non-blocking: the
      commit is an ancestor of origin/develop, and t303's own criteria still pass after this change.
    — Severity: minor — Class: blocking
    — Required fix: one sentence in spec.md §1 or §4 naming SPEC-SYNC-STRATEGY-KEY-001 and commit
      63b4628a6 as the origin of batch 3.10's canonical binding, and noting that REQ-TSA-005 exists
      to protect it.

D5. TIER-FIELD-ABSENT — spec.md:L1-14 frontmatter vs plan.md:L3
    — plan.md declares "Tier S", but spec.md's frontmatter carries no `tier:` field
      (`grep -n 'tier' spec.md plan.md` → no match in either). Any mechanical tier reader — the
      plan-audit ceiling map `harness.plan_audit_tier_ceilings`, the tier-differentiated PASS
      threshold — defaults to Tier L, contradicting the plan's prose and the dispatch's stated 0.75
      threshold. This audit applied 0.75 as dispatched; the drift is real regardless of verdict.
    — Severity: minor — Class: blocking
    — Required fix: add `tier: S` to spec.md frontmatter.

D6. PREDICATE-IGNORES-TAB-GATE — acceptance.md:L15-17 (AC-TSA-001 mode_admits)
    — The predicate inspects only batch-level `condition` and never the tab-level `mode_condition`
      key, which does exist in this schema. Harmless here and changes no counted value (tab_3 has
      `mode_condition: null`; the sole non-null value "INITIALIZATION" belongs to a different
      vocabulary and gates a tab with no auto_branch question — both measured, §5), but the
      predicate is stated in general terms it does not fully honour.
    — Severity: minor — Class: optional
    — Required fix: one clause in the Given — "tab-level `mode_condition` is not a
      `git_strategy.mode` gate and is out of the predicate's scope" — so the omission is deliberate
      and recorded rather than latent.

D7. NEUTRALITY-DELTA-BY-LINE-NUMBER — acceptance.md:L199-223 (AC-TSA-008)
    — The delta is stated as "the same three lines and no others", which is line-set identity, not
      line-content identity: a new violation appended to line 3, 625, or 915 would preserve the set.
      Adjudicated sound overall (§8) and the risk is closed in practice by AC-TSA-007 plus the
      deletion-only constraint, but the assertion is weaker than its own intent.
    — Severity: minor — Class: optional
    — Required fix: assert the three matched lines verbatim (compare full grep output, not just the
      line numbers).

D8. AC-003-WEAKER-THAN-REQ-005 — acceptance.md:L96-109 vs spec.md:L65-66
    — REQ-TSA-005 requires batch 3.10's question/field/current_value_path be "byte-unchanged";
      AC-TSA-003 verifies `grep -c` of the path string == 3. A mutant that edited 3.10's question
      *text* around the path, its `header`, its `note` (L1011), or its `options` would keep the
      count at 3 and pass. Overlaps D1 and is closed by the same fix.
    — Severity: minor — Class: optional
    — Required fix: covered by D1's diff-shape criterion; alternatively assert the three lines
      verbatim rather than counting them.

D9. NO-EXPLICIT-REQ-AC-CITATION — acceptance.md (0 REQ-TSA refs), spec.md (0 AC-TSA refs)
    — Coverage is complete and unambiguous, but no document states the mapping: `grep -c 'REQ-TSA'
      acceptance.md` = 0, `grep -c 'AC-TSA' spec.md` = 0, `grep -c 'REQ-TSA' plan.md` = 0. Every
      link is reconstructed by the reader. This is what holds Traceability at the 0.75 band.
    — Severity: minor — Class: blocking
    — Required fix: append the covering REQ ID to each AC heading (e.g. "## AC-TSA-001 — the
      counting criterion (primary) — covers REQ-TSA-001, REQ-TSA-003, REQ-TSA-004").
```

**No BLOCKING-severity defect.** D1 and D2 are `major`; the remaining seven are `minor`. Five are
classed blocking (they touch a criterion the SPEC itself states) and four optional. Per M6, the
length of this list does not justify a FAIL and was not used to manufacture one.

---

## 15. Recommendation

**PASS.** The SPEC is unusually well-grounded for its size: every coordinate it cites was
re-measured on `7ed6edb3e` and every one **HOLDS** — the nine `auto_branch` lines, the three batch
conditions, all three struct-tag facts, the pair identity, both occurrence counts, the total question
count, the RED-now tuple on both copies, the neutrality triple, and the trailing-comma boundary
claim, which the author derived by reading and this audit confirmed by executing. Nothing DRIFTED and
nothing was FALSE. The central decision (delete, do not rebind) is correct and its rationale
survives adversarial reading: rebinding really would convert two dead questions into two live
duplicates on one key, and AC-TSA-001 is precisely the criterion that would catch a run-phase agent
doing it.

`PASS-WITH-DEBT` was considered and rejected. The debt is real but it is entirely verification-layer
strengthening on a change whose correctness is already pinned by AC-TSA-001 + AC-TSA-002 + AC-TSA-007
— which between them kill seven of the eight enumerated mutants, most several times over.

Route before or during run phase, cheapest first:

1. **D5** — add `tier: S` to frontmatter (one line; removes a mechanical/prose contradiction).
2. **D9** — append covering REQ IDs to the eight AC headings (eight edits; lifts Traceability out of
   the 0.75 band, the score's binding constraint).
3. **D1 + D8** — add one diff-shape criterion. This is the highest-value single addition: it closes
   the only mutant that escapes the entire artifact set (h′), and subsumes D8.
4. **D2** — promote the `make build` DoD bullet to a numbered criterion asserting the regenerated
   asset, not just the source pair. This is the only escape that reaches a shipped artifact.
5. **D3 + D4** — two short prose additions (a Gaps subsection; one sentence naming t303 and
   `63b4628a6` as batch 3.10's provenance).
6. **D6, D7** — optional; adopt only if the run phase is already editing `acceptance.md`.

Kickoff is not gated on any of these. Items 3 and 4 are the two worth doing before M1, because both
change what the run phase is required to measure.

---

## 16. Gaps — what this audit did NOT observe

- **`make build` was not run** and the embedded asset was not inspected, by explicit dispatch
  instruction (it would dirty the tree mid-audit). D2's severity is therefore assessed from the
  criterion set, not from an observed stale-asset failure.
- **`go test ./...` was not run**, by explicit dispatch instruction. No statement is made here about
  the suite's state on this tree.
- **`moai spec lint` was not run.** The DoD requires it in run phase; this audit did not pre-empt it,
  so no lint-engine verdict on this SPEC is claimed.
- **The post-change state was not measured** — nothing was edited, so every "after" figure in §6 is
  derived from the criteria as written plus the measured baseline, never observed.
- **The runtime interview was not exercised.** No consumer of `tab_schema.json` exists to exercise
  (`grep -rln 'tab_schema' --exclude-dir=.git .` returns only SPEC docs, reports,
  `.moai/manifest.json`, and `internal/template/internal_content_leak_test.go`; `grep -c 'tab_schema'
  .claude/skills/moai-workflow-project/SKILL.md` → `0`, confirming the SPEC's §4 claim). REQ-TSA-003
  and REQ-TSA-004, which speak about what the *interview presents*, are therefore verified through
  AC-TSA-001's static reconstruction of the condition semantics, not through any running interview.
- **t303's branch state was not inspected** beyond confirming `63b4628a6` is an ancestor of
  `origin/develop`. No claim is made about what else that in-progress SPEC may still land.
- **No MCP backend second opinion was obtained.** This is a Claude-anchor audit; `audit_multi` /
  `codex_audit` / `glm_audit` were not invoked, so the verdict rests on a single auditor.

## 17. Residual risk — what could still be wrong despite the above

- **AC-TSA-001's predicate is a reconstruction of semantics no code enforces.** It was validated
  against the schema's every condition and against `implementation_notes.conditional_batches`
  ("Batches in Tab 3 show/hide based on git_strategy.mode selection"), but with no runtime consumer
  there is no executable oracle. Should a consumer ever be written with different show/hide
  semantics, `personal=1, team=1, manual=0` could be the wrong target while remaining internally
  consistent. The risk is bounded by the deletion being subtractive: it removes questions bound to a
  path no struct field matches, which is true under any interview semantics.
- **The unverifiable clause of REQ-TSA-006** ("alter no other question object") remains open until
  D1 lands. A run-phase agent could satisfy all eight criteria while altering an untouched question.
- **A stale embedded asset would be invisible to the criteria** until D2 lands (mutant (e)).
- **The schema's self-declared counters are already inconsistent and this change widens the drift.**
  Measured: `total_settings: 60` vs computed 48 questions; `total_batches: 18` vs 17 actual batches
  (`spec.md` L6-8 of the JSON). Post-change the first becomes 60 vs 46. No criterion covers these
  counters and the SPEC never mentions them. Reported here rather than as a defect because the
  inconsistency **predates** the card by 12 and 1 respectively — this change does not create it —
  and no consumer reads the file. A future card that adds entries to this surface (the follow-up
  `spec.md:L126` anticipates) should reconcile them.
- **Line-number coordinates in `plan.md` §C will drift** the moment anything above line 516 changes.
  They are correct on `7ed6edb3e` and are anchored by `field` value as well as by line, so the
  drift is recoverable — but a stale re-read of that plan against a moved tree would mis-target.

---

*Audited on `7ed6edb3e`, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316`, branch
`WT-tabschema-autobranch`. Iteration 1 of 1 (Tier S).*
