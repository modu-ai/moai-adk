# SPEC Review Report: SPEC-GITIGNORE-ROOT-GUARD-001

Card: t377
Iteration: 1/1 (Tier S ceiling = 1)
Auditor: plan-auditor (independent, M1 context-isolated)
**Verdict: FAIL**
**Overall Score: 0.69** (Tier S PASS threshold 0.75)

> Export note: this file is a verbatim export of the iteration-1 judgment already
> delivered to the lead. No re-audit was performed and no finding was revised for
> this export. Anything not actually executed is recorded in § Gaps rather than
> reconstructed.

---

## Baseline-attribution

| Field | Value |
|---|---|
| Tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t377` (its own git root) |
| HEAD | `3f03d9c36faf49bdcb155d98a7009fc9d8dd9659` |
| Branch | `WT-gitignore-parity` |
| Audit subject | `.moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md` (Tier S; ACs inline; no `acceptance.md`, no `plan.md`) |
| Tooling | Bash CLI only — the moai MCP server failed to connect this session, so no `mcp__moai__*` call was made |

Every command in this report was executed in that session against that tree at that
commit. Three mutants were planted and all three restored: root `.gitignore`,
template `.gitignore`, and `spec.md` frontmatter. Final state observed:

```
$ git status --porcelain
?? .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/
$ git rev-parse HEAD
3f03d9c36faf49bdcb155d98a7009fc9d8dd9659
```

M1 Context Isolation: no author reasoning, prior draft, or conversation history was
passed to this audit. `spec.md` was the only SPEC artifact read.

## Version attribution

The audit was performed against the SPEC **before** two later edits:

1. the `§E` restructure into `### In Scope` / `### Out of Scope` headings, and
2. the `§B` strengthening that cites `internal/template/sanitized_pair_parity_test.go`.

Consequences, recorded rather than papered over:

- Edit 2 **resolves** the former defect D2b (sanitized-pair precedent omitted). D2b is
  struck below and retained only as a strike-through record.
- Every line number cited below falls before the old `§E` start (line 141), so a
  §E-confined edit leaves them valid. That the edit was so confined was **not
  observed** — see § Gaps.
- The `§E` judgment (no defect) rests on the lead's statement that content is
  unchanged, not on a re-read — see § Gaps.
- **Discrepancy on record.** The §E restructure was described as being needed to
  satisfy the `MissingExclusions` lint rule, but the lint run on the *old* form
  returned `No findings` at every severity, and that green was proven non-vacuous by
  mutation. The old §E already passed on this tree. The restructure is harmless; the
  stated reason does not match the measurement.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-GRG-001..005, sequential, no gap, no
  duplicate, consistent padding (spec.md:106-116).
- **[PASS] MP-2 GEARS format compliance** (requirement layer) — all five entries are
  Ubiquitous (`The X MUST …`) or Unwanted (REQ-GRG-004 `MUST NOT fire`). `MUST` in
  place of `shall` is this repository's consistent idiom and was **not** scored as a
  failure; the passive-voice subject omission is deducted under Clarity as D6 instead.
  Judgment made against the `REQ-XXX` layer in `spec.md`; the AC layer was graded
  under Group 4 (D1/D4/D5/D7), never here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with
  correct names and types; zero snake_case aliases; `phase: "v3.1.4 target"` is a
  release label, not a prohibited lifecycle stage; `tier: S` present as an optional
  field.
- **[N/A] MP-4 Section 22 language neutrality** — the SPEC names no language-specific
  tool and is scoped to one subsystem. N/A auto-passes per the MP-4 precedent.
- **[N/A] MP-5 D7 cross-SPEC reconciliation** — the body carries zero
  `SPEC-<DOMAIN>-NNN` references besides its own id, so the D7 verification verb has
  no subject to execute against. N/A per the MP-4 precedent.
- **[PASS] MP-6 D8 cross-platform discipline** — the literal `syscall` does not appear
  in the SPEC body; D8-4 auto-pass.
- **[N/A] MP-7 clarification gate** — neither `plan.md` nor `research.md` exists in
  the SPEC directory (`spec.md` is the only file), so the marker grep has no target.
  N/A per the MP-4 precedent, reason stated.

**No must-pass criterion failed.** The FAIL verdict derives from the dimension scores
and the blocking defects below.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ-GRG-001/003 passive with no named actor (spec.md:106,110); "declared set" defined only in prose. Otherwise unusually clear. |
| Completeness | 0.75 | 0.75 | All sections present; §B measurements reproduce 4/4. Deducted because §A's live-risk justification rests on a mechanism that was never verified (D3). |
| Testability | 0.50 | 0.50 | AC-GRG-001 selects an unrelated test (measured); AC-GRG-005 unsatisfiable as written (measured); AC-GRG-004 inherits 001's selector; AC-GRG-003's RED never observed. |
| Traceability | 0.75 | 0.75 | REQ-GRG-003 has no AC directly asserting it (D8). No orphan ACs; the other four REQs each carry ≥1 AC. |

Arithmetic mean (0.75 + 0.75 + 0.50 + 0.75) / 4 = 0.6875 → **0.69**, below the Tier S
threshold of 0.75. The mean is dominated by Testability: repairing D1 and D4 lifts
that dimension over the line, which is why this FAIL is a demand to rewrite the
acceptance criteria rather than a rejection of the design.

---

## Evidence — commands run and output observed

### E1 — §B measurement, re-measured independently (all four figures match)

```
$ grep -vE '^\s*#|^\s*$' .gitignore | sort -u > /tmp/t377_root.txt ; wc -l < /tmp/t377_root.txt
     177
$ grep -vE '^\s*#|^\s*$' internal/template/templates/.gitignore | sort -u > /tmp/t377_tpl.txt ; wc -l < /tmp/t377_tpl.txt
     135
$ comm -23 /tmp/t377_root.txt /tmp/t377_tpl.txt | wc -l
      44
$ comm -13 /tmp/t377_root.txt /tmp/t377_tpl.txt
.agents/skills/moai*
**/.mink/auth/
$ grep -n "^\.agents" .gitignore
130:.agents/
```

The SPEC's 177 / 135 / 44 / 2 reproduce exactly, the identity of the two
template-only rules matches §B, and the containment claim (`.agents/skills/moai*` ⊂
`.agents/`) holds. **The central design argument — that full rule-set parity is
unavailable — stands on independent measurement.**

### E2 — §A reproduction: the existing guard cannot catch a root-only regression

`internal/template/embed_gitignore_generated_test.go:32-46` reads only the
`EmbeddedTemplates()` `.gitignore`; no code path opens the root file. Reproduced
rather than inferred:

```
$ cp .gitignore /tmp/t377_root_backup.gitignore
$ grep -v '^\.moai/project/graph/$' .gitignore > /tmp/t377_mut.txt ; cp /tmp/t377_mut.txt .gitignore
$ grep -c '^\.moai/project/graph/$' .gitignore
0
$ go test ./internal/template/ -run TestEmbeddedGitignoreCoversGeneratedArtifacts -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.429s
$ git check-ignore -v .moai/project/graph/edges.jsonl ; echo "rc=$?"
rc=1
```

Restored:

```
$ cp /tmp/t377_root_backup.gitignore .gitignore ; git status --porcelain .gitignore
(no output)
$ git check-ignore -v .moai/project/graph/edges.jsonl
.gitignore:244:.moai/project/graph/	.moai/project/graph/edges.jsonl
```

**§A is correct.** No missed detection path was found.

### E3 — AC-GRG-001's selector is vacuous today (basis for D1)

```
$ go test ./internal/template/ -run TestGitignore -count=1 -v
=== RUN   TestGitignore_IgnoresSkillMirrorOnly
--- PASS: TestGitignore_IgnoresSkillMirrorOnly (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.258s
```

The selector matches an unrelated pre-existing test and does **not** match
`TestEmbeddedGitignoreCoversGeneratedArtifacts` (no `TestGitignore` substring). The
criterion is green before any work exists.

### E4 — AC-GRG-005's command cannot be satisfied as written (basis for D4)

```
$ grep -rn "generatedArtifactIgnoreRules" internal/template/ --include='*.go'
internal/template/embed_gitignore_generated_test.go:23:// generatedArtifactIgnoreRules are the .gitignore lines that must reach users.
internal/template/embed_gitignore_generated_test.go:26:var generatedArtifactIgnoreRules = []string{
internal/template/embed_gitignore_generated_test.go:48:	for _, rule := range generatedArtifactIgnoreRules {
```

Three lines — a comment mention, the declaration, and a use site. The command counts
occurrences, not declaration sites, so "→ exactly one declaration site" is
unreachable through it.

### E5 — the neutrality-enforcement claim is mechanically false for the cited SHA (basis for D2)

The SHA asymmetry §B describes is real:

```
$ grep -n "59a622c5a\|graph/" .gitignore
237:# which already happened once, on an unmerged branch (59a622c5a, +180,178).
244:.moai/project/graph/
$ grep -n "59a622c5a\|graph/" internal/template/templates/.gitignore
243:.moai/project/graph/
```

The root file's exact wording was planted into the template `.gitignore` and all
three guards run:

```
$ printf '# which already happened once, on an unmerged branch (59a622c5a, +180,178).\n' >> internal/template/templates/.gitignore
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.875s
$ go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.366s
```

To separate "guard blind to `.gitignore`" from "guard blind to this SHA", the same
line was re-planted with an 8-character SHA:

```
$ printf '# see commit 59a622c5 for the incident.\n' >> internal/template/templates/.gitignore
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
--- FAIL: TestTemplateNoInternalContentLeak (0.64s)
    template internal-content leak detected (1 occurrences, mode=strict):
      [1] templates/.gitignore | class=S2-short-sha-sentence-final | match=59a622c5
FAIL
```

So the walker **does** scan `templates/.gitignore` (it is registered in
`leakScannedDotfiles`) and the SHA class **does** fire — but the pattern is
`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`, and the cited `59a622c5a` is nine characters, so it
passes through. The doctrinal ground (§25 forbids commit SHAs) survives and the
byte-equality exclusion still holds; what does not hold is the claim of *mechanical
enforcement*, which was never observed. Restored:

```
$ cp /tmp/t377_tpl_backup.gitignore internal/template/templates/.gitignore
$ git status --porcelain internal/template/templates/.gitignore
(no output)
```

### E6 — §A's risk model has the direction inverted (basis for D3)

First half of the claim, **verified true**: `ManagedCleanTargets`
(`internal/cli/update/deploy/deploy.go:57-89`) is 7 entries, all under `.claude/`
(`settings.json`, `commands/moai`, `agents/moai`, `skills/moai*` glob, `rules/moai`,
`output-styles/moai`, `hooks/moai`); `.moai/config` is removed separately at
`deploy.go:187`. The root `.gitignore` is in neither. The SPEC's decision to discard
the `CleanMoaiManagedPaths` rationale is correct.

Second half, **false**: `moai update` does touch the root `.gitignore` — by merge,
not deletion.

```
$ grep -rn "gitignore" internal/cli/update/plan/plan.go
internal/cli/update/plan/plan.go:47:	case base == ".gitignore":
```

which returns `merge.EntryMerge`. On the ordinary template-sync path the file is
backed up before deploy (`internal/cli/update_template_sync.go:459`) and merged after
(`:510` → `MergeGitignoreFile`). Reading that function
(`internal/cli/update/merge/merge.go:57-97`) shows the result is **template content ∪
user-only lines**: the deployed template text is the base, and only backup lines
absent from the template set are appended.

The consequence inverts the SPEC's argument. A root-only rule loss is **repaired** by
the next `moai update`, whereas in the cited `.sh` / `.sh.tmpl` incident update was
the **aggressor** that reverted the deployed side. The SPEC states "The shape is the
same here"; it is the opposite. The risk remains real — the window between regression
and the next update is unbounded, and one `git add -A` inside it reopens the hole —
but the SPEC's sole stated justification for the work is wrong in a checkable way,
and this SPEC has already discarded one wrong rationale.

### E7 — spec-lint green, proven non-vacuous

```
$ moai spec lint .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md
✓ No findings — all SPEC documents are valid
```

Mutant probe to establish the linter actually reads this file:

```
$ sed -i '' 's/^priority: P2$/prioritee: P2/' .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md
$ moai spec lint .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md
SEVERITY  CODE                FILE                                               LINE  MESSAGE
WARNING   FrontmatterInvalid  .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md  1     Frontmatter required field missing: priority [grandfathered era — downgraded to warning]

0 error(s), 1 warning(s)
```

Restored; lint returns `No findings` again.

**Side observation (not a defect).** The `[grandfathered era — downgraded to warning]`
annotation shows this SPEC classifies as grandfathered — `progress.md` is absent, so
H-1 fires. During plan-phase, genuine frontmatter errors are therefore downgraded to
warnings, which makes a clean lint weaker evidence than it appears. Not this SPEC's
fault; worth knowing before citing the green.

### E8 — §E out-of-scope items (old form): no defect

```
$ grep -n "observability\|navigator\|moai/chain\|migrate-tx\|mink" .gitignore
(no output)
$ grep -n "observability\|navigator\|moai/chain\|migrate-tx\|mink" internal/template/templates/.gitignore
169:**/.mink/auth/
```

The four t373 paths are absent from **both** surfaces, so excluding them cannot affect
this SPEC's own correctness, and §E's closing paragraph already disclaims undeclared
paths. `**/.mink/auth/` is template-only with no root counterpart, matching §B.
Keeping a credential rule out of a generated-artifact guard SPEC is correct scope
discipline.

### E9 — the later §B claim, verified true

```
$ grep -rn "gitignore" internal/template/sanitized_pair_parity_test.go internal/template/rule_template_mirror_test.go
(no output)
```

The zero hit is not a vacuous sweep: the `sanitizedPairPaths` registry is populated
(7 entries, all `.claude/**` `.md` rule files) and `workflowOptMirroredPaths` likewise.
`.gitignore` is genuinely in neither the sanitized-pair tier nor the byte-parity tier.

---

## Defects Found

Severity ordering. `Class` follows the M6 finding-consumption discipline: **blocking**
findings affect correctness, internal consistency, or a criterion the SPEC itself
states; **optional** findings are surfaced and left to the orchestrator's discretion.

**D1** — `spec.md:121` — AC-GRG-001's verify command `go test ./internal/template/ -run TestGitignore -count=1`
selects the unrelated pre-existing `TestGitignore_IgnoresSkillMirrorOnly` and never
selects the guard under audit; the criterion is green before any work exists and stays
green after landing unless the implementer happens to name the new test `TestGitignore…`
— Severity: **critical** — Class: **blocking** — Evidence: E3 — Required fix: anchor the
selector to the real test name (`-run '^TestGitignoreDeclaredRulesBothSurfaces$'`) and
pin that name in the REQ or AC text so a rename breaks the criterion rather than
silently emptying it.

**D2** — `spec.md:79-80` — the claim that template neutrality is "enforced by
`template-neutrality-check.yaml`" is mechanically false for the 9-character SHA the
root file cites; the strict tier passes with the exact line planted — Severity:
**critical** — Class: **blocking** — Evidence: E5 — Required fix: limit the sentence to
its doctrinal ground (CLAUDE.local.md §25 commit-SHA prohibition) and drop the
enforcement clause, or record the observation that the guard's SHA class matches only
7-8 characters.

**D3** — `spec.md:54-58` — the `.sh` / `.sh.tmpl` precedent is cited with the risk
direction inverted; `MergeGitignoreFile` makes the targeted root-only regression
self-healing at the next `moai update`, where the cited incident had update as the
aggressor — Severity: **major** — Class: **blocking** — Evidence: E6 — Required fix:
withdraw "The shape is the same here" and restate the risk as the unbounded window
between regression and the next `moai update`. The SPEC remains justified under the
restatement; only the reason changes.

**D4** — `spec.md:135-136` — AC-GRG-005's grep counts occurrences (3 on the current
tree), not declaration sites, so the stated expectation of one is unreachable through
the given command — Severity: **major** — Class: **blocking** — Evidence: E4 — Required
fix: anchor to declaration syntax, e.g.
`grep -rn '^var generatedArtifactIgnoreRules' internal/template/ --include='*.go' | wc -l` → `1`.

**D5** — `spec.md:130-133` — AC-GRG-004 supplies no command of its own ("verify by
running it on the unmodified tree"), so it inherits D1's broken selector; as a
non-firing assertion it is additionally mutant-permeable — a mutant that skips the
assertion whenever divergence > 0 satisfies the AC while violating REQ-GRG-004 — Severity:
**minor** — Class: **optional** — Evidence: E3 plus the mutant-probe reasoning — Required
fix: give it an independent command paired with a §B re-measurement, so the pass is
observed under measured divergence rather than under a cited number.

**D6** — `spec.md:106,110` — REQ-GRG-001 and REQ-GRG-003 are passive ("MUST be
asserted", "The assertion MUST be over") with no named subject, leaving the GEARS
subject slot empty — Severity: **minor** — Class: **optional** — Required fix: "The guard
shall assert …".

**D7** — `spec.md:126-129` — AC-GRG-003's RED has never been observed; the
template-mutation-plus-`make build` failure is asserted, not measured, so the
criterion carries a green-path cell with no RED-now cell — Severity: **minor** — Class:
**optional** — Required fix: run that mutation once before implementation and record the
output in §A's format.

**D8** — `spec.md:110-111` — REQ-GRG-003 (line-discrete, never bytes, never full-set
equality) has no AC directly asserting it — Severity: **minor** — Class: **optional** —
Required fix: map it explicitly onto AC-GRG-004 or add a dedicated AC.

**~~D2b~~** — ~~`spec.md:80-82` — the `sanitized_pair_parity_test.go` precedent is not
cited, leaving unanswered the nearest in-repo alternative a reader meets on opening
`rule_template_mirror_test.go`~~ — **STRUCK: resolved by the later §B edit**, whose new
claim was independently verified (E9).

Also noted and **not** a defect: AC-GRG-002 is the one solid criterion in the set — §A
carries its command, its output, and its restore confirmation, and it reproduced here
(E2). AC-GRG-006 is observable but its pass condition rests on a human reading ("names
which file was missing which rule"); a literal string condition would make it binary.

---

## Gaps — what was NOT observed

Recorded as gaps rather than inferred, per the no-unobserved-claim invariant.

- **The post-edit SPEC text.** Neither the restructured `§E` nor the strengthened `§B`
  was re-read. The §E judgment (E8, no defect) rests on the lead's statement that
  content is unchanged.
- **That the §E edit was confined to §E.** Line-number validity for D1-D8 depends on
  this and it was not checked.
- **Lint on the new form.** `moai spec lint` was run only against the pre-edit file.
- **The `make build` path** (AC-GRG-003's mutation) — outside this audit's read-only
  scope, so the template-surface RED was never produced.
- **An actual `moai update` run.** D3 rests on reading `plan.go`,
  `update_template_sync.go`, and `merge.go`, not on executing the command, which is
  destructive in this tree.
- **The clean-reinstall branch** (`internal/cli/update_clean_install.go:403` and `:509`)
  — it may handle `.gitignore` differently from the template-sync path examined.
- **Real CI execution** of `template-neutrality-check.yaml`. The workflow YAML was read
  and its commands reproduced locally, including the step-scoped
  `MOAI_TEMPLATE_LEAK_STRICT=1`.
- **`plan.md`** — absent from the SPEC directory, so the Tier S input pair reduced to
  `spec.md` alone.

## Residual-risk

- If `gitignoreBackup` is empty on some path, or if clean-reinstall diverges from
  template-sync, D3's self-healing conclusion weakens. "The shape is the same here"
  would remain an unverified claim either way.
- D2 is measured against the guard's **current** pattern. A later widening to 9-character
  SHAs would make the SPEC's sentence true retroactively — but it was still written
  without observation, which is the defect.
- The 44 / 2 divergence figures go stale the moment any lane touches either
  `.gitignore`. They are pinned to `3f03d9c36` and must be re-measured, not re-cited.
- The 0.69 score is an arithmetic mean dominated by Testability 0.50. Repairing D1 and
  D4 lifts that dimension above the threshold, so this FAIL is a demand to rewrite the
  acceptance criteria, **not** a rejection of the design.

---

## Recommendation

The design is right and should not be changed. The 46-rule intentional divergence was
reproduced independently (E1), and under that constraint a declared-rule-set assertion
across both surfaces is the only honest instrument available — byte equality and
full-set equality are both genuinely foreclosed. What fails here is the justification
and the acceptance criteria, not the approach.

Four fixes, in order:

1. **D1 first.** Landing as-is would approve a guard whose own acceptance criteria pass
   by running an unrelated test — the SPEC would reproduce the exact defect class it
   was written to eliminate.
2. **D2 and D3** — rewrite the two justification sentences. Both conclusions survive
   intact; only the reasons change. This matters more than usual because the SPEC has
   already spent one wrong rationale (`CleanMoaiManagedPaths`), and a second wrong one
   in the same document is a pattern rather than a slip.
3. **D4** — make the command measure what the criterion claims.
4. The optional findings (D5-D8) are surfaced for the orchestrator's discretion and do
   not by themselves justify the FAIL.

Tier S caps plan-audit at one iteration, so a confirming re-audit after these fixes
requires an explicit decision to exceed the ceiling; the alternative is an operator
PASS-WITH-DEBT judgment. That choice belongs to the lead and the operator, not to this
agent. If a re-audit is ordered, it would be scoped to the enumerated defect delta
above plus a regression check over D1-D8, not a fresh full pass.
