# t572 Sync-Audit Verdict — SPEC-OWNERSHIP-SILENCE-001 (2026-09-09)

- Auditor: sync-auditor (independent post-implementation skeptical evaluation)
- Tree: `.claude/worktrees/t572` · branch `WT-ownership-lint-silent` · HEAD `9acd8b11792cb84ef1957f3ab36334c381e32d91` (verified at audit start AND re-read before this verdict — unchanged)
- Base: `3ac58b5a1` (origin/develop tip at dispatch) · commit stack: 8 (plan ×1, orchestrator-direct ×1, run M1-M3 ×3, body-repair ×1, sync close ×1, backfill ×1)
- Audit constraint honored: read-only except this file. No commits, no pushes.

---

## Verdict

**PASS — 0.94** (harmonic mean). Blocking findings: **0**. Should-fix: **0**. Advisory: **3** (below).

## Dimension Scores

| Dimension | Score | Basis |
|-----------|-------|-------|
| Functionality | 0.95 | AC 8/8 re-judged against observed evidence; runtime-green closure observed (below); suspicious branch emits exactly the designed finding, three designed silences untouched (diff-scope read + preservation tests GREEN) |
| Security | 0.95 | No trust-boundary or input-validation surface touched; measurement-state statement can never masquerade as violation (REQ-OWN-005 pinned by test + m2 mutant); parser untouched per plan §H.4 |
| Craft | 0.94 | Plain-Info idiom consistent with file siblings (Advisory unset asserted by test); message carries all 5 required elements incl. `emptyOrValue` "(none)" path (2-fixture design makes m3 observable); stale-comment repair (REQ-OWN-009) landed in the same diff; intent-change comments present |
| Consistency | 0.93 | Rule-doc byte-twins aligned (cmp_rc=0, old trigger 0 both copies); CHANGELOG claims match measured facts (199, rc unchanged, 8/8); one stale subordinate clause in the CHANGELOG entry (advisory A1 below) |

Harmonic mean: 4 / (1/0.95 + 1/0.95 + 1/0.94 + 1/0.93) = **0.9424 ≈ 0.94** — above the Tier M threshold (0.80).

## Runtime-Green Closure Record (mandated check #2 — the sync-phase residual, observed not inferred)

manager-docs' own Gap (`.moai/reports/t572/sync-evidence.md` Residual-risk: "백필 후 코퍼스 lint에서 본 SPEC의 OwnershipTransitionUnmeasured = 0건이 확인되면 닫힌다") is **CLOSED by observation**:

- **Command**: `go run ./cmd/moai spec lint --strict` (background; `go run` builds from this tree, so the judging build = measured tree per VCI §2.2 second coordinate)
- **Tree**: `9acd8b117` (HEAD, post-backfill — the tree a future lint will see)
- **Verbatim output tail** (full output: `/tmp/t572-syncaudit-corpus-lint.txt`, 1,471,404 bytes — see Gaps G2 for export constraint):

```
0 error(s), 4718 warning(s)
exit status 1
EXIT=1
```

- **SPEC-OWNERSHIP-SILENCE-001 findings** (`grep 'OWNERSHIP-SILENCE' | awk '{print $1,$2}' | sort | uniq -c`):

```
  10 WARNING CoverageIncomplete
  10 WARNING ModalityUnjudged
```

- **OwnershipTransitionUnmeasured corpus total**: `199` — **self-SPEC count: 0** (`grep -c` returned 0, rc=1 zero-match)
- **Judgment**: the SPEC carries ZERO OwnershipTransition\* findings — Invalid 0, Unmeasured 0, Unreachable 0. The repaired rule measured this card's own sync transition (commit `26878d787`, trailer `Authored-By-Agent: manager-docs`) against the matrix, found the expected owner (manager-docs for the close transition) matched, and passed it **at runtime** — not from trailer presence + matrix mapping, but from the rule's own emitted output. Inherited rc=1 is unchanged and error count is now 0 (the pre-existing MissingExclusions error was repaired by `91c7079dc` before this measurement — exclusion-fix evidence corroborated and now re-observed).

## Findings

### Blocking — none

### Should-fix — none

### Advisory

- **A1 — CHANGELOG subordinate clause is stale on the current tree** (`CHANGELOG.md:12`, landed in `26878d787`): "measured rc=1 before and after (the single error on this tree is a pre-existing plan-phase artifact finding, not this change)". The parenthetical was accurate at its measurement time (run-phase postlint, 1 error) but the tree now carries **0 errors** — the MissingExclusions error it describes was repaired by `91c7079dc` (manager-spec body repair) before sync landed. The load-bearing claim ("Info findings never change the exit status, rc=1 before and after") remains true and re-observed. No action required for this card; fix opportunistically if the entry is ever revised.
- **A2 — fresh lint evidence lives in /tmp only** (`/tmp/t572-syncaudit-corpus-lint.txt`): the dispatch constrained the auditor to a single permitted write (this verdict file), so the 1.4 MB full output could not be exported under `.moai/reports/t572/`. The verdict carries the tail + counts; run-phase's own export (`postlint-strict.txt`, 199 unmeasured on other SPECs, self 0) remains the tracked full-text record for the code behavior, and this audit's fresh run re-confirms both counts on the post-backfill tree.
- **A3 — template neutrality guard not re-executed by this audit**: `go test ./internal/template/...` GREEN is cited from the run-phase ledger (run-evidence.md AC-OWN-006 (4), 34.687s). Twin byte-identity (`cmp_rc=0`) and zero old-trigger tokens were re-observed by this audit; the guard suite itself was not re-run (cost under lane load vs. unchanged inputs since its GREEN run — the twin files are the only changed inputs and both were re-verified directly).

## Mandated-check record (all 7 executed)

1. **AC-by-AC re-judgment** — AC-OWN-001 RED 4-element cell verified in run-evidence.md (verbatim FAIL, exit 1, tree `b642479ec`, right-reason: silent nil branch, stated in-test doc comment); spot re-runs this audit: `go test ./internal/spec/ -run TestOwnershipTransition -count=1` → `ok ... 1.258s` exit 0; `go vet ./internal/spec/` clean; `GOOS=windows GOARCH=amd64 go build ./internal/spec/` exit 0; `cmp -s` twins rc=0; `grep -c "subject prefix"` both copies print 0 (exit ignored per F4 disposition).
2. **Runtime-green closure** — observed, above.
3. **Mutation non-vacuity** — run-evidence.md m1/m2/m3 each carry verbatim `--- FAIL:` output with real assertion messages (m1: "got 0: []" ×2 + strict-safe "got none"; m2: HasErrors flipped true + severity rejection; m3: `"(none)"` literal missing), each with exit code 1 and revert confirmation (`MUTANT` marker count 0 + diff-stat scoped to the 2 intended files). Detection records, not pass reports. No surviving mutant — why-acceptable slot in plan §G carried as designed.
4. **Scope discipline** — `git diff 3ac58b5a1..HEAD --name-only -- internal/` = exactly `internal/spec/lint_ownership.go`, `internal/spec/lint_ownership_test.go`, `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md`. Full name-only diff: 18 files, all in-card (SPEC artifacts ×4, evidence ×9, CHANGELOG ×1, rule-doc local twin ×1, internal ×3). No `manager-develop.md`/`.toml`, no `lint.go`, no parser changes. Diff hunks read: the three silent-nil sites (`rec == nil`, `ownerNone`, unrecognized actor) and sibling finding blocks are absent from the diff; only the trailer-less branch, its comment block, and the two stale doc-comments (`:179`, `:367` regions) changed.
5. **Docs alignment** — rewritten Cross-Reference section (loaded and read from the local twin; twins byte-identical): trailer WHO SSOT stated ("The commit subject is never consulted as the WHO signal"), three finding codes with severities, three designed silences with reasons, strict escalation stated Warning-only. CHANGELOG claims cross-checked against measurement: 199 (verified: per-spec list 199 lines, 199 unique; corpus grep 199), rc unchanged (re-observed rc=1), 8/8 AC (this audit). A1 advisory noted.
6. **Commit hygiene** — `git log --format='%h %(trailers:key=Authored-By-Agent,valueonly)' 3ac58b5a1..HEAD` over all 8 commits: `b642479ec`=manager-spec, `0538585ac`=orchestrator-direct, `c7b940e45`=manager-develop, `a6274068e`=manager-develop, `1649bff43`=manager-develop, `91c7079dc`=manager-spec, `26878d787`=manager-docs, `9acd8b117`=manager-docs — exactly the dispatch expectation (spec ×2, develop ×3, orchestrator-direct ×1, docs ×2). Subjects canonical: SPEC-ID present in 7; `0538585ac` carries the card id (`docs(t572):` — evidence-baseline note, not a SPEC-phase commit); sync subject carries "3-phase close". Ownership boundaries: body repair `91c7079dc` correctly owned by manager-spec (B4 re-delegation); sync commit touched spec.md frontmatter `status`+`updated` only (diff read); backfill `9acd8b117` is the D3 placeholder resolution owned by manager-docs (diff read: one field line).
7. **Four-dimension scoring** — above.

## Evidence-bearing report (VCI §3)

**Claim**: SPEC-OWNERSHIP-SILENCE-001 as landed on `9acd8b117` satisfies AC-OWN-001..008 with observed evidence; the sync-phase runtime-green residual is closed; the change is scope-clean and docs-aligned. Score 0.94, PASS.

**Evidence**: this file § Runtime-Green Closure Record + § Mandated-check record (each item names its command and verbatim output); upstream ledgers re-read and spot-verified: run-evidence.md (RED 4-element cell, m1-m3 verbatim FAILs, corpus export), sync-evidence.md, exclusion-fix-evidence.md, unmeasured-per-spec.txt (199 lines / 199 unique / self 0), postlint-strict.txt (tail matches run-phase claims), golangci-lint.txt (1 inherited errcheck, file untouched by this card's diff — attribution to t577 axis confirmed via zero-line diff on `zz_t528_overacceptance_test.go`).

**Baseline-attribution**: every fresh observation in this audit was run in this run, against this tree (`WT-ownership-lint-silent` @ `9acd8b117`, re-read immediately before the verdict and unchanged). Run-phase claims were verified against their pinned trees (`b642479ec`, `c7b940e45`) via ledger inspection, not re-measurement where the measurement is historical by nature (RED on pre-fix tree, mutants). The one claim carried from a prior measurement and NOT re-measured: template neutrality guard suite (A3).

**Gaps**:
- G1 — plan-phase content debt on this SPEC's own directory: 10 CoverageIncomplete + 10 ModalityUnjudged warnings (observed in this audit's lint run, matching exclusion-fix-evidence.md). Deliberately out of scope (corpus-wide debt class); recorded, not judged against this card.
- G2 — fresh corpus-lint full text at `/tmp/t572-syncaudit-corpus-lint.txt` (single-write constraint). Tail + counts preserved above.
- G3 — template neutrality guard suite cited from ledger, not re-run (A3).
- G4 — inherited errcheck (t577 axis) and develop CI spec-lint red left un-re-measured — outside this card's judgment axis per acceptance §D.1 AC-OWN-007 and spec.md §6.

**Residual-risk**:
- The 199-count and rc-stability were re-observed at `9acd8b117`; future SPEC additions will grow the unmeasured count until the convention-enforcement follow-up card (spec §3.1 recommendation) lands — the growth is advisory noise by construction (Warning-only strict promotion), but the corpus number in the CHANGELOG is a point-in-time measurement, not an invariant.
- `lint.go` is under active change on develop (t518 axes, plan §H.1). The merge into develop re-measures everything; this verdict binds this tree, not the merged one (per the develop-integration doctrine, the merged tree must be re-measured or proven tree-identical).
- The mock-injection seam (`getOwnershipTransitionRunner`) means unit tests prove the emission branch and its matrix logic, while corpus behavior is proven separately by the runtime lint observation above — both legs were observed here, so the combined claim is covered; but the corpus leg depends on git history shape (trailer presence), which the backfill window preserves by construction.

---

Verdict rendered: **PASS, 0.94**. Card t572 sync-phase audit closed.
