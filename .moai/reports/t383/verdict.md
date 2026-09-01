# t383 — run-phase verdict (M4)

SPEC-MEMORY-STORE-RECONCILE-001. Worktree `.claude/worktrees/t383`, branch
`WT-memory-index-budget`, base `297a21ea7`. Five-section evidence format.

`$D` = `$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory`
`$L` = `$HOME/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory`

---

## 1. Claim

Every index target in the active auto-memory store now resolves; no in-repo surface enumerated
in spec.md §A.5 still teaches the unconfirmed loading cut; the write-anyway instruction sits in
an always-loaded surface; and the token guard no longer enumerates a slot it cannot measure.
All 16 acceptance criteria PASS. Two defects were planted by this run-phase's own repairs and
both were caught by executing the criterion rather than re-reading it.

---

## 2. Evidence

### The store moved five times; the defect figures never did

Five independent readings across plan and run phase, nobody in this card touching the index
except at M3 (which edits no index line at all):

| Reading | bytes | entry lines | unique targets | topic files | **dangling** | **orphans** |
|---|---|---|---|---|---|---|
| measurements.md M1/M4 | 26,280 | 123 | 189 | 177 | **58** | **46** |
| iteration 2 | 26,290 | 123 | — | — | **58** | — |
| iteration 3 | 26,577 | 124 | 190 | 178 | **58** | — |
| run-phase M0 | 28,009 | 130 | 196 | 184 | **58** | **46** |
| run-phase M3 BEFORE | 31,366 | 139 | 205 | 193 | **58** | **46** |
| run-phase M3 AFTER | 31,366 | 139 | 205 | 251 | **0** | **46** |

This is the card's central empirical result, and it is what licenses AC-MSR-009 to target an
exact 0 while AC-MSR-010 pins no literal. It also retroactively validates DEBT 1: the entry-line
denominator moved 123 → 139 during this card's own execution.

### AC matrix

| AC | Verdict | Command run | Verbatim output |
|---|---|---|---|
| **001** | **PASS** | `grep -c '25KB' .claude/rules/moai/workflow/moai-memory.md` / `grep -c '25KB\|200-line' .claude/output-styles/moai/moai.md` / `grep -c '25KB\|25 \* 1024\|25600' internal/config/token_budget_guard.go` | `0` / `0` / `0`. Red direction from real baselines at HEAD: `3` / `1` / `2` |
| **002** | **PASS** | cond.1 `grep -c 'moai memory doctor' <local>` and `<mirror>` | `4` / `4` (≥1 both). cond.2 by reading: neither copy states a numeric loading limit — verified by the AC-001 greps returning 0 and by reading § MEMORY.md Index Budget |
| **003** | **PASS** | `grep -n 'shorter' .claude/rules/moai/workflow/moai-memory.md` | `131:#### Compressing the index means making entries shorter — never fewer` + `133:` (removal makes the topic file unreachable) |
| **004** | **PASS** | `grep -n 'moai memory doctor' .claude/rules/moai/workflow/moai-memory.md` | 4 matches. Surrounding text: § "Two stores, and only one of them is loaded" names both candidate stores as invisible from the index, and states the `--dir` + `exists: true` + non-zero index-line preconditions |
| **005** | **PASS** | `head -1 .claude/rules/moai/core/moai-constitution.md` | `# MoAI Constitution` — not `---`, so always-loaded. Clause inserted: "**A long index never justifies dropping a lesson.** … Not 'write the file, skip the index line' … Not 'skip the lesson'". Names the store by `moai memory doctor`, never a literal path |
| **006** | **PASS** | `grep -n 'memoryHead\|memoryHeadByteCap\|memoryHeadLineCap' internal/config/token_budget_guard.go` | exit `1` (no match). `grep -n 'MEMORY.md'` → lines `109`, `114`, `127`, all inside the single TOMBSTONE block beginning `// TOMBSTONE — there is deliberately no MEMORY.md fixed surface slot` |
| **007** | **PASS** | mutation: re-insert slot → `go test ./internal/config/ -run TestFixedSlotsExistInRepoTree` | RED: `--- FAIL: TestFixedSlotsExistInRepoTree` naming `".../MEMORY.md" names a path absent from the repository tree`. After revert, GREEN: `--- PASS: TestFixedSlotsExistInRepoTree` |
| **008** | **PASS** | surface list diff, before vs after | `18d17 < T383_SURFACE=<repoRoot>/MEMORY.md` — exactly one removed line, no addition, no reorder. Isolated A/B on the same tree: total `76009` **with** the slot (count 18) and `76009` **without** (count 17) → the element contributed exactly **0** tokens, which is exactly the change observed |
| **009** | **PASS** | gate `jq -e '.[0]\|select(.exists==true and .index_lines>0)'` then `jq '[.[0].findings[]?\|select(.Code=="MEMORY_DANGLING_INDEX_LINK")]\|length'` | gate exit `0` both files. BEFORE `58` (≥1, red direction satisfied), AFTER `0` |
| **010** | **PASS** | `grep -c '^- \['` / `grep -o '](\([^)]*\.md\))' \| sort -u \| wc -l` / `wc -c` | entry `139 → 139`, targets `205 → 205`, bytes `31366 → 31366`. Neither metric decreased; M3 edits the index not at all |
| **011** | **PASS** | `find "$L" -maxdepth 1 -mindepth 1 -name '*.md' \| wc -l`; `find "$L" -maxdepth 1 -name '*.md' -newermt '2026-08-30' \| wc -l` | `1098 → 1098`; recent-mtime `0 → 0`. Copy log: `skipped_exists: 0`, and `cp -n` used, so zero overwrites |
| **012** | **PASS** | `git status --short` | 8 tracked modified + 2 untracked report/SPEC dirs. **No path under either store appears.** Staging done by explicit per-file pathspec |
| **013** | **PASS** | three `diff -q` local↔mirror | `rc1=0 rc2=0 rc3=0`. All three were rc=0 at HEAD too, so this is preserved parity, not luck |
| **014** | **PASS** | `test -f` gate ×3, then the neutrality grep | all three files listed by `ls -1`; grep `exit=1` — **exactly 1**. Red direction: `claude-profiles` planted in a scratch copy → `811:claude-profiles`, `exit=0`; scratch discarded. Missing-file case confirmed `exit=2` |
| **015** | **PASS** | `.moai/reports/t383/m0-sample.md` | 12 paths + rule (every 5th of 58, indices 1…56); per-file verdicts with evidence; **0 of 12 superseded** vs threshold ≥4 → PROCEED; coverage limit states indices 57-58 unreached and the 21% sample bound |
| **016** | **PASS** | anchor `test -f go.mod` → `./bin/moai spec lint` → liveness → `grep -c` | anchor `0`; `lint_exit=0`; `liveness_nonempty=0`; `liveness_summary=0`; summary `0 error(s), 1096 warning(s)`; **count `0`**. Red directions below |

### AC-MSR-016 red directions, both executed

```
# vacuous: a lint that never ran
$ : > /tmp/t383-empty.txt; grep -c 'SPEC-MEMORY-STORE-RECONCILE-001' /tmp/t383-empty.txt
0                                    <-- the OLD criterion PASSED here
$ test -s /tmp/t383-empty.txt; echo $?
1                                    <-- the liveness assertion correctly FAILS

# cwd: run from the SPEC directory
$ cd .moai/specs/SPEC-MEMORY-STORE-RECONCILE-001 && moai spec lint > /tmp/x 2>&1
$ grep -c 'SPEC-MEMORY-STORE-RECONCILE-001' /tmp/x
6                                    <-- spurious FAIL, same binary, same tree
$ cd .moai/specs/SPEC-MEMORY-STORE-RECONCILE-001 && test -f go.mod; echo $?
1                                    <-- the corrected anchor correctly REFUSES
```

That a `grep -c` of 0 means "clean" rather than "rule never fires" is established from a real
baseline: the same lint output names **151 distinct SPEC IDs**, the busiest carrying 39 findings.

### Quality gates

```
$ go vet ./internal/config/...                 -> vet_exit=0
$ gofmt -l <the two edited .go files>          -> (no output; both formatted)
$ go test ./internal/config/...
ok  github.com/modu-ai/moai-adk/internal/config            2.161s
ok  github.com/modu-ai/moai-adk/internal/config/atomicfile (cached)
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy (cached)
```

`go test ./...` was NOT run, per the standing prohibition; CI owns the full-suite verdict.

---

## 3. Baseline-attribution

- **Tree**: `297a21ea7`, worktree `.claude/worktrees/t383`, branch `WT-memory-index-budget`,
  re-read with `git rev-parse --short HEAD` / `git branch --show-current` immediately before
  each commit. `origin/develop` is **26 commits ahead** at close (`git rev-list --count
  --left-right origin/develop...HEAD` → `26 0`); **not absorbed**, per the lead's instruction
  that absorption happens only inside the integration window.
- **Tool provenance (closes gap G2)**: every `spec lint` figure comes from `./bin/moai`, built
  from this tree by `make build` (`BuildID=v3.1.2-959-g297a21ea7-dirty`), invoked **by path**.
  The PATH binary is a different build and reports differently — plan-phase recorded 8 errors /
  64 warnings from it, the tree's own binary reports `0 error(s), 1096 warning(s)`. The two are
  not comparable, which is precisely the §2.2 hazard; only the by-path figure is cited.
- **Store figures**: all from `moai memory doctor --json --dir "$D"`, `exists`/`index_lines`
  gate asserted before any count, capital `.Code` selector. Raw JSON committed as
  `reconcile-before.json` / `reconcile-after.json`.
- **Copy set**: derived from the **live** index at M3 time by `derive-missing.sh`, not from any
  frozen list. The live population was 58, matching plan-phase and the doctor's count in the
  same window — no old number was forced.

---

## 4. Gaps

**Introduced by this card, and the most important line in this report:**

- **G7 — the always-loaded token budget was raised, 76,000 → 76,210.** REQ-MSR-004's clause must
  live in an always-loaded file (C5), and the surface had only **201 tokens** of headroom
  (measured: 75,799 before this card's doctrine edits). The clause was first cut by ~1,000 bytes,
  and the residual 210 tokens were added to the constant so the pre-existing headroom is
  *preserved rather than consumed* — not an arbitrary new margin. The raise is documented in
  `token_budget_guard.go` with the before/after measurement. **The surface is saturated**: 201 of
  76,000 is 0.26%, before and after. The next card that grows an always-loaded file hits this
  guard. The real fix — the large-rule stub-and-lazy-load diet — is named in the constant's own
  prior justification and is not this card's scope. A reviewer who thinks a guard should not be
  raised by the card that trips it has a fair objection, and this paragraph exists so that
  objection is possible.

**Enumerated, not closed:**

- **G6 — an unenumerated surface still asserts the cut.**
  `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-memory-official.md:125`
  states "The first 200 lines of `MEMORY.md` (or the first 25KB, whichever comes first) are loaded
  at the start of every session." Same unconfirmed claim, on a surface spec.md §A.5 does not list
  and REQ-MSR-001 does not reach. Found by the broadened M0 sweep. **Not edited** — editing it
  would be silent scope expansion past the §A.5 enumeration the AC matrix is written against.
- **G2 — closed.** See Baseline-attribution.
- **G4 — open, and deliberately not decided.** Three in-repo statements about store derivation
  disagree with the code. This card edited all three files and **re-asserted none**: the
  `moai.md:165` literal legacy path was removed (rewriting its cap clause while leaving its path
  clause would have been re-assertion), the constitution's literal path was replaced by the
  resolving command, and the new doctrine explicitly declines to state the derivation
  ("this file's own description of the derivation and the observed behaviour do not agree, so
  rely on what the tool reports"). `moai-memory.md:27` is left untouched. Deciding which side is
  wrong remains a follow-up.
- Deferred per spec.md §D: the **46 orphans** (measured stable at 46 before and after — DEBT 6);
  the topic-file cap, now further exceeded (193 → 251, exactly +58 as C3 predicted); the legacy
  store's 1,098 files; the loader's cap shape; and the part of the incident's belief that lives
  in agent-level instructions outside this repository.
- **M0 sample bound**: 12 of 58 is 21%, reaching index 56. `project_spec_catalog_cleanup.md` and
  `reference_mcp_2026_07_28_check.md` were not examined. A 0-of-12 result bounds the superseded
  share loosely; it is not proof that none of the 58 is superseded.
- **A contradiction inside the copied set, recorded not resolved.** Sampled entries 16 and 26
  (`feedback_full_test_suite_verification.md`: run the FULL suite; `feedback_no_local_full_suite.md`:
  never locally) give opposite instructions. Neither carries a SUPERSEDED marker, so the stated
  rule classes both as live and both were copied. This card's copy step is what makes the
  contradiction reachable; adjudicating it is triage work.
- **Not run**: full test suite, sync-phase, push, PR, integration lock — all out of scope by
  instruction.

---

## 5. Residual-risk

- **Nothing in this repository can keep the store reconciled.** It lives outside the tree; no CI
  job can see it. A future divergence is caught by re-running `moai memory doctor` — which is now
  in doctrine — and by nothing else. This is unchanged by the card and is stated as a bound on
  what was delivered.
- **The M0 gate is a sample, not a census.** If a superseded file sits among the 46 unsampled
  targets, it has now been copied. Reversible by deleting the copy (REQ-MSR-010 copy-only).
- **AC-MSR-002 condition 2 and AC-MSR-005's "offers no branch" are readings, not greps.** Another
  reader may judge the clause differently; the mechanical halves cannot be satisfied by absence,
  which is the guarantee actually being offered.
- **The doctrine is now silent on the loader's limit.** A session that wanted a number gets an
  instruction to measure instead. If `moai memory doctor` is unavailable, that session has less
  guidance than before — accepted deliberately, because the previous number was unconfirmed and
  a wrong number is worse than a measurement instruction.
- **Two defects were planted by this run-phase's own repairs** (§ below). Both were caught by
  execution. The base rate is now five for five, so the prior for a sixth in any follow-up round
  should be treated as high.
- **Uncommitted-at-audit risk**: `origin/develop` is 26 ahead and other lanes are active on this
  machine. Nothing here has been verified against the merged tree; a semantic clash on merge is
  possible and is the integration window's business, not this card's.
