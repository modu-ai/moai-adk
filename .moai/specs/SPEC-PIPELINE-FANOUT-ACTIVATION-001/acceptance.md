# Acceptance — SPEC-PIPELINE-FANOUT-ACTIVATION-001

Fifteen criteria, within the Tier M budget of 16 that this SPEC introduces.

Each criterion carries a **Covers:** line naming the requirement(s) it verifies, so the
requirement-to-criterion mapping is explicit rather than inferred.

## Conventions binding every criterion

- **Content-token anchoring only.** No judging command references a line number. Three of the ten
  fan-out sites moved line position between the originating audit and this plan phase, so line
  anchors are known to be unstable here.
- **Baselines are observed, not assumed.** Every command below was executed in the worktree
  `.claude/worktrees/pipeline-fanout` at HEAD `903f899d1` during plan authoring, and the recorded
  baseline is that run's output.
- **Non-vacuity.** Each criterion's baseline differs from its expected post-change state, so no
  criterion can pass without the change having landed.
- **Shell portability.** Commands use explicit file lists and quoted expansions. They avoid `ls`
  entirely — the interactive shell here aliases `ls` to `ls -la`, so an anchored `ls | grep '^name'`
  always returns zero matches. Unquoted variable word-splitting is also avoided, because the shell
  is zsh, which does not word-split unquoted expansions by default; a first attempt at the
  discretionary-token census failed for exactly that reason.
- **Working directory.** All relative paths resolve from the worktree root.

---

## AC-PFA-001 — Fan-Out Index exists in all three local routers

**Covers:** REQ-PFA-001

**Given** the three phase-router skills,
**When** the index heading is searched for,
**Then** each router carries exactly one Fan-Out Index section.

```bash
for f in plan run sync; do
  echo "$f=$(grep -c '^## Fan-Out Index' ".claude/skills/moai/workflows/$f.md")"
done
```

- Observed baseline: `plan=0`, `run=0`, `sync=0`
- Expected after: `plan=1`, `run=1`, `sync=1`

---

## AC-PFA-002 — Fan-Out Index exists in all three template mirrors

**Covers:** REQ-PFA-001, REQ-PFA-009

**Given** the three mirrored routers,
**When** the same heading is searched for,
**Then** each mirror carries exactly one Fan-Out Index section.

```bash
for f in plan run sync; do
  echo "$f=$(grep -c '^## Fan-Out Index' "internal/template/templates/.claude/skills/moai/workflows/$f.md")"
done
```

- Observed baseline: `plan=0`, `run=0`, `sync=0`
- Expected after: `plan=1`, `run=1`, `sync=1`

---

## AC-PFA-003 — Every Fan-Out ID appears in BOTH its router index and its own site file (local)

**Covers:** REQ-PFA-002

**Given** the ten fan-out sites and the three index tables,
**When** each canonical ID is counted separately in its router file and in its site file,
**Then** each ID appears at least once in each — proving the site-to-index pairing, not merely that
the ID exists somewhere.

```bash
# id : router file : site file
while IFS='|' read -r id router site; do
  r=$(grep -co "$id" ".claude/skills/moai/workflows/$router")
  s=$(grep -co "$id" ".claude/skills/moai/workflows/$site")
  echo "$id router=$r site=$s"
done <<'EOF'
FO-PLAN-1|plan.md|plan.md
FO-PLAN-2|plan.md|plan/spec-assembly.md
FO-RUN-1|run.md|run/phase-execution.md
FO-RUN-2|run.md|run/task-decomposition.md
FO-RUN-3|run.md|run/task-decomposition.md
FO-RUN-4|run.md|run/task-decomposition.md
FO-SYNC-1|sync.md|sync.md
FO-SYNC-2|sync.md|sync/quality-gates-quality.md
FO-SYNC-3|sync.md|sync/quality-gates-quality.md
FO-SYNC-4|sync.md|sync/doc-execution.md
EOF
```

- Observed baseline: every line reads `router=0 site=0`
- Expected after: every line has `router>=1` AND `site>=1`
- **Why this replaced a tree-wide count.** The original form counted each ID across the whole
  workflow tree and asserted `>= 2`. That would pass if both occurrences sat inside one index table,
  which does not demonstrate the pairing REQ-PFA-002 actually requires. FO-PLAN-1 and FO-SYNC-1 have
  router and site in the same file, so their two counts read from one file; the other eight are
  genuine cross-file assertions.

---

## AC-PFA-004 — Every Fan-Out ID appears in BOTH its router index and its own site file (template)

**Covers:** REQ-PFA-002, REQ-PFA-009

**Given** the mirrored workflow tree,
**When** the same paired census runs against it,
**Then** the same pairing holds on the template side.

```bash
while IFS='|' read -r id router site; do
  r=$(grep -co "$id" "internal/template/templates/.claude/skills/moai/workflows/$router")
  s=$(grep -co "$id" "internal/template/templates/.claude/skills/moai/workflows/$site")
  echo "$id router=$r site=$s"
done <<'EOF'
FO-PLAN-1|plan.md|plan.md
FO-PLAN-2|plan.md|plan/spec-assembly.md
FO-RUN-1|run.md|run/phase-execution.md
FO-RUN-2|run.md|run/task-decomposition.md
FO-RUN-3|run.md|run/task-decomposition.md
FO-RUN-4|run.md|run/task-decomposition.md
FO-SYNC-1|sync.md|sync.md
FO-SYNC-2|sync.md|sync/quality-gates-quality.md
FO-SYNC-3|sync.md|sync/quality-gates-quality.md
FO-SYNC-4|sync.md|sync/doc-execution.md
EOF
```

- Observed baseline: every line reads `router=0 site=0`
- Expected after: every line has `router>=1` AND `site>=1`

---

## AC-PFA-005 — Each router index lists exactly its own phase's IDs

**Covers:** REQ-PFA-001, REQ-PFA-002

**Given** the three index tables,
**When** each router file is searched for ID prefixes,
**Then** a router references only its own phase's IDs, confirming the index is phase-scoped rather
than a copied global list.

```bash
for f in plan run sync; do
  p=$(grep -co 'FO-PLAN-' ".claude/skills/moai/workflows/$f.md")
  r=$(grep -co 'FO-RUN-'  ".claude/skills/moai/workflows/$f.md")
  s=$(grep -co 'FO-SYNC-' ".claude/skills/moai/workflows/$f.md")
  echo "$f plan=$p run=$r sync=$s"
done
```

- Observed baseline: `plan plan=0 run=0 sync=0` / `run plan=0 run=0 sync=0` / `sync plan=0 run=0 sync=0`
- Expected after: `plan` has `plan>=2` and `run=0` and `sync=0`; `run` has `run>=4` and `plan=0` and
  `sync=0`; `sync` has `sync>=4` and `plan=0` and `run=0`
- Note: the `plan` router hosts FO-PLAN-1 inline as well as in its table, so its count exceeds the
  table's two rows; the criterion is the zero cross-phase counts.

---

## AC-PFA-006 — No discretionary fan-out phrasing remains in the local sites

**Covers:** REQ-PFA-003

**Given** the seven local files hosting the ten sites,
**When** the discretionary tokens are counted per file,
**Then** every fan-out occurrence is gone and only the one non-fan-out occurrence remains.

```bash
for f in \
 .claude/skills/moai/workflows/plan.md \
 .claude/skills/moai/workflows/plan/spec-assembly.md \
 .claude/skills/moai/workflows/run/phase-execution.md \
 .claude/skills/moai/workflows/run/task-decomposition.md \
 .claude/skills/moai/workflows/sync.md \
 .claude/skills/moai/workflows/sync/quality-gates-quality.md \
 .claude/skills/moai/workflows/sync/doc-execution.md ; do
  echo "$(grep -cE 'orchestrator MAY|MAY be sharded' "$f")  $f"
done
```

- Observed baseline:
  ```
  1  .claude/skills/moai/workflows/plan.md
  2  .claude/skills/moai/workflows/plan/spec-assembly.md
  1  .claude/skills/moai/workflows/run/phase-execution.md
  3  .claude/skills/moai/workflows/run/task-decomposition.md
  1  .claude/skills/moai/workflows/sync.md
  2  .claude/skills/moai/workflows/sync/quality-gates-quality.md
  1  .claude/skills/moai/workflows/sync/doc-execution.md
  ```
  Total 11 = 10 fan-out sites + 1 unrelated tier-judgment skip condition in `spec-assembly.md`.
- Expected after: `0` for every file except `spec-assembly.md`, which stays at `1` — that surviving
  occurrence is the out-of-scope tier-judgment clause and must NOT be edited.

---

## AC-PFA-007 — No discretionary fan-out phrasing remains in the template mirrors

**Covers:** REQ-PFA-003, REQ-PFA-009

**Given** the seven mirrored files,
**When** the same census runs,
**Then** the same post-state holds, confirming the promotion landed on both sides.

```bash
for f in \
 .claude/skills/moai/workflows/plan.md \
 .claude/skills/moai/workflows/plan/spec-assembly.md \
 .claude/skills/moai/workflows/run/phase-execution.md \
 .claude/skills/moai/workflows/run/task-decomposition.md \
 .claude/skills/moai/workflows/sync.md \
 .claude/skills/moai/workflows/sync/quality-gates-quality.md \
 .claude/skills/moai/workflows/sync/doc-execution.md ; do
  echo "$(grep -cE 'orchestrator MAY|MAY be sharded' "internal/template/templates/$f")  $f"
done
```

- Observed baseline: identical to AC-PFA-006's baseline (`1 / 2 / 1 / 3 / 1 / 2 / 1`) — measured, not
  assumed
- Expected after: identical to AC-PFA-006's expected state

---

## AC-PFA-008 — The fail-open fallback survives at every site, both sides

**Covers:** REQ-PFA-004

**Given** each promoted site,
**When** the fallback markers are counted,
**Then** the count is unchanged from baseline on both sides — the promotion added an obligation but
removed no escape hatch.

```bash
RE='no error, no warning|runs unchanged|identical output|runs as one pass|path unchanged|serially as before|serial path'
for f in \
 .claude/skills/moai/workflows/plan.md \
 .claude/skills/moai/workflows/plan/spec-assembly.md \
 .claude/skills/moai/workflows/run/phase-execution.md \
 .claude/skills/moai/workflows/run/task-decomposition.md \
 .claude/skills/moai/workflows/sync.md \
 .claude/skills/moai/workflows/sync/quality-gates-quality.md \
 .claude/skills/moai/workflows/sync/doc-execution.md ; do
  echo "local=$(grep -cE "$RE" "$f") tmpl=$(grep -cE "$RE" "internal/template/templates/$f") $f"
done
```

- Observed baseline:
  ```
  local=1 tmpl=1  plan.md
  local=1 tmpl=1  plan/spec-assembly.md
  local=1 tmpl=1  run/phase-execution.md
  local=4 tmpl=4  run/task-decomposition.md
  local=1 tmpl=1  sync.md
  local=1 tmpl=1  sync/quality-gates-quality.md
  local=1 tmpl=1  sync/doc-execution.md
  ```
- Expected after: every file's local and template counts are **greater than or equal to** their own
  baseline, and `sync/quality-gates-quality.md` rises to **at least 2** on both sides.
- **Why that one file must rise.** A per-site inspection during plan authoring found that nine of the
  ten sites carry a fallback sentence but FO-SYNC-2 (the sync MX-tag-scan sharding site) carries
  none — the single baseline hit in that file belongs to FO-SYNC-3. Promoting FO-SYNC-2 without
  adding a fallback would leave a conditional obligation with no escape hatch, so REQ-PFA-004
  requires one to be added, and this criterion is what detects its absence.
- Rationale for the relational form elsewhere: a promotion may reword a fallback sentence, so exact
  equality would be brittle. What must hold is that no fallback is lost and the missing one appears.

---

## AC-PFA-009 — D-1 resolved in the local plan-auditor

**Covers:** REQ-PFA-006

**Given** the Retry Loop Contract,
**When** the contradicting clause and the agreeing clause are counted,
**Then** the contradiction is gone and the delta phrasing is present.

```bash
echo "contradiction=$(grep -c 'Full audit PLUS regression check' .claude/agents/moai/plan-auditor.md)"
echo "delta=$(grep -ci 'enumerated defect delta' .claude/agents/moai/plan-auditor.md)"
```

- Observed baseline: `contradiction=1`, `delta=1`
- Expected after: `contradiction=0`, `delta>=2` (the pre-existing defect-list clause, plus the
  rewritten Retry Loop Contract clause)

---

## AC-PFA-010 — D-1 resolved in the template plan-auditor

**Covers:** REQ-PFA-006, REQ-PFA-009

**Given** the mirrored agent file,
**When** the same counts are taken,
**Then** the same post-state holds.

```bash
echo "contradiction=$(grep -c 'Full audit PLUS regression check' internal/template/templates/.claude/agents/moai/plan-auditor.md)"
echo "delta=$(grep -ci 'enumerated defect delta' internal/template/templates/.claude/agents/moai/plan-auditor.md)"
```

- Observed baseline: `contradiction=1`, `delta=1`
- Expected after: `contradiction=0`, `delta>=2`

---

## AC-PFA-011 — Verdict authority is preserved through the D-1 fix

**Covers:** REQ-PFA-007

**Given** the delta-scope rewrite,
**When** the verdict-authority clause is searched for,
**Then** it is still present on both sides — the cost reduction did not weaken who owns the verdict.

```bash
echo "local=$(grep -ci 'verdict authority stays with this agent' .claude/agents/moai/plan-auditor.md)"
echo "tmpl=$(grep -ci 'verdict authority stays with this agent' internal/template/templates/.claude/agents/moai/plan-auditor.md)"
```

- Observed baseline: `local=1`, `tmpl=1`
- Expected after: `local>=1`, `tmpl>=1`
- Non-vacuity note: this is a preservation criterion, so baseline and expected state coincide by
  design. Its discriminating power comes from being run *after* the M1.3 rewrite — a rewrite that
  dropped the clause would move it to `0`.

---

## AC-PFA-012 — Tier size budget is stated in the SSOT, both sides

**Covers:** REQ-PFA-008, REQ-PFA-009

**Given** the SPEC Complexity Tier section,
**When** the budget marker and its three ceilings are searched for,
**Then** the budget is present with all three tier values, on both sides.

```bash
for p in ".claude" "internal/template/templates/.claude"; do
  f="$p/rules/moai/workflow/spec-workflow.md"
  echo "$p marker=$(grep -c 'REQ/AC budget' "$f") ceilings=$(grep -cE '\b(8|16|25)\b' "$f")"
done
grep -n -A 8 'REQ/AC budget' .claude/rules/moai/workflow/spec-workflow.md
```

- Observed baseline: `marker=0` on both sides; the `grep -n -A 8` prints nothing
- Expected after: `marker>=1` on both sides, and the printed block shows the three ceilings 8 / 16 /
  25 together with the statement that they apply independently to the requirement count and to the
  acceptance-criterion count
- The `ceilings` count is context only; the marker count and the printed block are the decision.

---

## AC-PFA-013 — The four intentional neutralization divergences are preserved

**Covers:** REQ-PFA-009

**Given** the template internal-content isolation doctrine,
**When** each local-only token is counted on both sides,
**Then** every one remains local-only — the mirror edits were semantic, not a copy.

```bash
check() { echo "$(grep -c "$1" "$2")/$(grep -c "$1" "internal/template/templates/$2")  $1"; }
check '88 pre-v3 SPECs'  .claude/skills/moai/workflows/plan.md
check 'Updated: 2026-05-25' .claude/skills/moai/workflows/plan.md
check 'internal/spec/ears.go' .claude/agents/moai/plan-auditor.md
check 'EARSModalityRule'      .claude/agents/moai/plan-auditor.md
```

- Observed baseline: `1/0` for all four
- Expected after: `1/0` for all four — unchanged
- This criterion is direction-aware: it asserts local-only-ness, never byte-identity between the two
  copies. A blind `cp` from local to template would flip every row to `1/1` and fail it.

---

## AC-PFA-014 — Scope discipline: no forbidden content added, no adjacent drift swept

**Covers:** REQ-PFA-010

**Given** the completed change,
**When** the template mirrors are scanned for forbidden classes and the two named out-of-scope sites
are inspected,
**Then** no forbidden content was introduced and neither out-of-scope site moved.

```bash
echo "-- this SPEC's own tokens leaked into any template mirror --"
grep -rnE 'SPEC-PIPELINE-FANOUT-ACTIVATION-001|REQ-PFA-|AC-PFA-' \
  internal/template/templates/.claude/ | wc -l | tr -d ' '

echo "-- SPEC-ID-shaped tokens in the mirrors this SPEC edits (delta check) --"
grep -rnE 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}' \
  internal/template/templates/.claude/skills/moai/workflows/ \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md | wc -l | tr -d ' '

echo "-- out-of-scope A: stale local hook description untouched --"
grep -c 'exit 2 to block sync completion' .claude/skills/moai/workflows/sync.md

echo "-- out-of-scope B: codemaps fan-out not promoted --"
grep -c 'orchestrator MAY' .claude/skills/moai/workflows/codemaps.md
```

- Observed baseline: own-token leak `0`; SPEC-ID-shaped tokens `12`; stale hook description `1`;
  codemaps `1`
- Expected after: own-token leak still `0`; SPEC-ID-shaped tokens still `12` (no increase); stale
  hook description still `1`; codemaps still `1`
- **Why the second scan is a delta check, not an absolute zero.** The twelve baseline matches are
  pre-existing. Eleven are documentation placeholders (`SPEC-AUTH-001`, `SPEC-BUG-042`) which are
  legitimately neutral example identifiers. The twelfth, in the mirrored plan-auditor at the
  LEAN-workflow clause, is a genuine internal identifier that predates this SPEC. It is a real
  pre-existing neutrality issue, **out of scope here**, and is deliberately left alone rather than
  silently swept into this diff. Asserting an absolute `0` would either fail on day one or invite
  exactly that out-of-scope cleanup, so the criterion asserts no *increase* instead.
- The first scan is the one that binds this SPEC's own additions, and it is an absolute zero.

---

## AC-PFA-015 — No promoted obligation is detached from its GEARS condition

**Covers:** REQ-PFA-005, REQ-PFA-003

**Given** the ten promoted fan-out clauses,
**When** every obligation verb is examined together with the text preceding it on its line,
**Then** each obligation is preceded by a bolded GEARS modifier, and none stands unconditional.

The criterion has two parts. Part A proves the promotion happened at all; Part B proves it was not
over-applied. **Part B alone would be vacuous** — it returns 0 both before the change (because no
obligation exists yet) and after a correct change — so Part A is what gives Part B meaning. Both
must be read together.

```bash
for side in "" "internal/template/templates/"; do
  A=0; B=0
  for f in \
   .claude/skills/moai/workflows/plan.md \
   .claude/skills/moai/workflows/plan/spec-assembly.md \
   .claude/skills/moai/workflows/run/phase-execution.md \
   .claude/skills/moai/workflows/run/task-decomposition.md \
   .claude/skills/moai/workflows/sync.md \
   .claude/skills/moai/workflows/sync/quality-gates-quality.md \
   .claude/skills/moai/workflows/sync/doc-execution.md ; do
    a=$(grep -hcE 'orchestrator shall' "$side$f"); A=$((A+a))
    b=$(grep -hoE '^.*orchestrator shall' "$side$f" | grep -vcE '\*\*(Where|While|When)\*\*'); B=$((B+b))
  done
  echo "side='${side:-local}' totalObligations=$A unguarded=$B"
done
```

- Observed baseline:
  ```
  side='local' totalObligations=0 unguarded=0
  side='internal/template/templates/' totalObligations=0 unguarded=0
  ```
- Expected after: `totalObligations >= 10` on **both** sides (Part A — the promotion landed at all
  ten sites), and `unguarded = 0` on **both** sides (Part B — no obligation is condition-free).

**Selector design notes, both empirically established rather than assumed.**

1. **Line-scoped, not sentence-scoped.** The obvious form — extract the sentence around the
   obligation with `[^.!?]*orchestrator shall[^.!?]*` — is **wrong here**, and was rejected after
   testing it. These files cite dotted filenames inside the condition clause, and the `.` in
   `orchestration-mode-selection.md` terminates the extracted fragment early, severing the leading
   `**Where**`. A correctly-guarded clause is then reported as unguarded. Verified on a synthetic
   line carrying a real citation: the sentence form flagged it (false positive, count 1) while the
   line-scoped form passed it (count 0).
2. **Bolded modifier, not bare word.** The guard anchors on `**Where**` / `**While**` / `**When**`
   rather than the bare words, because "when" and "where" appear constantly in ordinary prose in
   these files and would make the guard trivially satisfiable. All existing conditions in these
   files are already bolded, so this matches house style rather than imposing a new one.
3. **Falsification check.** On a synthetic fixture holding two guarded and two unguarded
   obligations, the selector returned exactly 2 and named the correct two lines. A selector that
   could not distinguish them would have returned 0 or 4.

---

## Definition of Done

- All fifteen criteria evaluated against their recorded baselines, with the observed post-change
  output cited verbatim for each.
- `make build` run after the mirror edits, so the embedded template filesystem is regenerated.
- The full test suite passes, confirming no template-guard regression (this change touches only
  markdown, so a guard failure would indicate a neutrality or namespace violation).
- Requirement count (10) and acceptance-criterion count (15) both inside the Tier M ceiling of 16
  that this SPEC introduces.

## Post-change behavioural indicator — not an acceptance criterion

The audit's success indicator for this milestone is that script-backed fan-out invocations rise
above zero (`plan-research-fanout.js` and `sync-audit-4dim.js`, currently 0 across 761 transcripts).
That is only observable once post-change transcripts accumulate, so it is deliberately **not** a
criterion here. Recording it separately keeps the documentation-only nature of this SPEC's
verification honest: every criterion above proves the text changed, none proves the behaviour did.
