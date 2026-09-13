# Sync Audit — SPEC-TODO-LAND-AUTO-DONE-001 (card t684, Tier M)

Auditor: sync-auditor (independent, fresh-judgment)
Date: 2026-09-13
Tree: worktree `.claude/worktrees/t684`, branch `WT-auto-done-on-land`, HEAD `0673b3ac0`
Diff scope: `74d872aaf..HEAD` (merge-base; `origin/develop` has since advanced to `5e0f71175` with other cards' work — the `origin/develop..HEAD` diff reverses that work and is NOT this card's scope, per gitflow-lane-protocol §8)

## Verdict

**Overall Verdict: PASS** (score 91.9/100, harmonic mean)
Must-pass dimensions (Functionality, Security): both PASS. No Critical/High findings.

## Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence (verbatim, this run, this tree) |
|-----------|--------|-------|---------|------------------------------------------|
| Functionality | 40% | 95/100 | PASS | `go test ./internal/kanban/ -run 'AutoDone\|Landed' -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 6.286s` (35 PASS, 0 FAIL, 0 SKIP, verified by `-v` count); `go test ./internal/cli/ -run 'TestTodoAutoDone' -count=1 -v` → 21 PASS, 0 FAIL, 0 SKIP incl. `TestLandedPredicate_PinnedCorpusReproduction` (ran 0.25s, not skipped) |
| Security | 25% | 95/100 | PASS | Boundary grep `grep -n "AskUserQuestion\|mcp__askuser" internal/cli/todo_autodone.go internal/kanban/autodone_scan.go` → no matches; fetch-counter test proves 0 network fetches without `--fetch`; `--end-of-options` guards both operand-shaped git calls (consistent with todo_landed.go's measured practice); log written 0o600/dir 0o700 |
| Craft | 20% | 88/100 | PASS | `gofmt -l` on all 10 touched Go files → empty; `go vet ./internal/kanban/ ./internal/cli/` → exit 0; `golangci-lint run ./internal/kanban/... ./internal/cli/... --timeout=4m` → 37 findings, all in untouched files (filtered grep on touched files → 0 hits; matches §E.3 claim of "0 new, 37 baseline"); coverage measured this run: `AutoDoneDecide`/`AutoDoneDistinctTexts`/`AutoDoneSkipReasons`/`LandedAttributions` 100%, `planAutoDone` 100%, `runTodoAutoDone` 89.7%, `applyAutoDoneCloses` 92.3%, `ScanLandedSubjects` 86.7% |
| Consistency | 15% | 90/100 | PASS | Close-line contract reuses the canonical `done <id> landing=landed` prefix + `source=auto-land`; seam usage (`todoRunCommand`/`todoGitOutput`) matches todo family; file modes 0o600/0o700 match `state_dir.go`; `"completed"` literal matches `todo.go:735` `recordFactoryCardState`; `gitEndOfOptions` phantom-coordinate discipline respected (no `merge-base` spelled in comments — verified); verb registered with `permittedVerbAdditions` SPEC citation |

Harmonic mean: 4 / (1/95 + 1/95 + 1/88 + 1/90) = **91.9/100**

## Mechanical verification log (all re-executed by this audit, this run)

| # | Command | Result |
|---|---------|--------|
| 1 | `git rev-parse origin/develop` / `git merge-base origin/develop HEAD` | `5e0f71175` / `74d872aaf` — true scope established |
| 2 | `git diff 74d872aaf..HEAD --stat` | 22 files, +3094/−12 — matches dispatch scope exactly |
| 3 | `go test ./internal/kanban/ -run 'AutoDone\|Landed' -count=1` | `ok ... 6.286s`; -v: 35 PASS / 0 FAIL / 0 SKIP |
| 4 | `go test ./internal/cli/ -run 'TestTodoAutoDone' -count=1` | `ok ... 29.939s`; -v: 21 PASS / 0 FAIL / 0 SKIP |
| 5 | `gofmt -l` (10 touched files) | clean |
| 6 | `go vet ./internal/kanban/ ./internal/cli/` | exit 0 |
| 7 | `go test ./internal/cli/ -run 'TestAuditLagUsesBinlagSeam'` | PASS — the ancestry baseline holds with the new coordinate |
| 8 | `grep -rn "\.ArchiveCard(" --include="*.go" internal/ cmd/ pkg/` (non-test) | exactly 2 hits: `todo.go:705`, `todo_autodone.go:388` |
| 9 | `git log 7835148d3 --format=%s \| grep -c '(t68,'` | 1 — exactly one comma-form subject for t68 at the pinned corpus commit |
| 10 | `golangci-lint run ./internal/kanban/... ./internal/cli/... --timeout=4m` | 37 findings, 0 in touched files |
| 11 | coverage (`go test -coverprofile` + `go tool cover -func`) | decision core 100%/100%/86.7%/66.7%(argv default branch); CLI verb core ≥89.7% |
| 12 | `GOOS=windows GOARCH=amd64 go build ./...` / `GOOS=linux ...` | both exit 0 |

## Highest-risk claims — adversarial results

### (a) Reissued-id collision fail-closed (AC-AD-004/005) — fixture-premise deviation JUDGED SOUND

Premise verified in source: `internal/kanban/todo_identity.go:293-313` (`ensureRecordIdentities`) refuses, on every whole-record write, any record holding the same card id live AND archived. The AC "given" (predecessor archived + reissued live in one record) is therefore genuinely unwritable through the store — the deviation is real, not evasive.

The resolution is fail-closed in both directions:
- AC-AD-004: predecessor injected via direct SQL (`injectArchivedPredecessor`); observable (`skip t902 reason=ambiguous-id`, card stays live) proven end-to-end by `TestTodoAutoDone_CollisionSkipsAmbiguous` — PASS in this run. No mutation is issued for the skipped card, so no write crosses the invariant.
- AC-AD-005: the decision is proven twice — unit row in `TestAutoDoneDecide` and the CLI `--dry-run` run printing `form=sha-recorded` end-to-end over the injected reissue state (PASS in this run). The non-dry write is refused loudly (`duplicate card identity`, asserted), which is the store correctly protecting integrity; in the field the live record holds no duplicate (the invariant guarantees it), and every close test proves that path applies normally.

**Judgment: the AC's intent — "the recorded SHA is the one evidence form that names this card" — is verified at the decision layer and through the full CLI pipeline (dry-run). The deviation leaves nothing material unverified. AC-AD-004/005 = PASS.**

### (b) Corpus re-pin 309/38 → 310/37 — VERIFIED, only t68 moved

Mechanical set-algebra proof: the corpus test pins the residual SET exactly and `mentioned == 347` constant; `residual' = residual − {t68}` ⟹ `attributed' = attributed + {t68}` and nothing else moved. The test passed in this run (not skipped — 0.25s). Independent subject check: `git log 7835148d3 --format=%s | grep '(t68,'` → exactly 1: `merge: Factory Mode -f N worker fan-out (t68, SPEC-FACTORY-WORKER-FANOUT-001)` — genuinely the comma-form trailing parenthetical (id opens the group, non-card qualifier after the comma). The negation guard moved no corpus id (residual set identical apart from t68).

### (c) mcp_build_identity_test.go baseline row — DECLARED ADDITION, not a smuggled exemption

`todo_autodone.go:340` is exactly `_, err := todoGitOutput("merge-base", "--is-ancestor", gitEndOfOptions, sha, ref)` — the scan's form-1 reachability question, the same referential-integrity class as the pre-existing `todo_landed.go:217` declaration. The sweep's semantics are exact-set both directions: a REMOVED declared coordinate also fails (`baseline hit ... MISSING`), so the row cannot decay into an exemption. The declaration comment carries the same rationale pattern as its predecessors. `TestAuditLagUsesBinlagSeam` PASS in this run. The author also respected the phantom-coordinate discipline: the word `merge-base` is not spelled in any new comment (verified — no extra sweep hits).

### (d) Close-surface exclusivity (AC-AD-016) — VERIFIED REPO-WIDE

The AC's structural test globs `todo*.go` only; I extended the check to the whole tree: `grep -rn "\.ArchiveCard(" --include="*.go" internal/ cmd/ pkg/` (non-test) → exactly 2 call sites (`internal/cli/todo.go:705` done, `internal/cli/todo_autodone.go:388` auto-done). The control-run clause (landed transitions nothing; State/position/text unchanged, no archive entry, only the landing field written) is asserted by the test and passed. The exclusivity claim is true of the tree, not just the test's window.

## AC coverage map (all 17 re-observed PASS in this run)

| AC | Test (PASS this run) | AC | Test (PASS this run) |
|----|----------------------|----|----------------------|
| 001 | `TestTodoAutoDone_LiveFilter` | 010 | `TestTodoAutoDone_ExecutionLog` |
| 002 | `TestTodoAutoDone_FormRecordedSHA` | 011 | `TestTodoAutoDone_ReversalRow` |
| 003 | `TestTodoAutoDone_FormSubjectAttribution` | 012 | `TestTodoAutoDone_DryRunByteIdentity` |
| 004 | `TestTodoAutoDone_CollisionSkipsAmbiguous` | 013 | `TestTodoAutoDone_Idempotence` |
| 005 | `TestTodoAutoDone_CollisionRecordedSHAOverrides` | 014 | `TestTodoAutoDone_NoLandingColumnWrites` |
| 006 | `TestTodoAutoDone_SpecNotCompletedSkips` (+unreadable edge) | 015 | `TestTodoAutoDone_FetchBoundary` (both subtests) |
| 007 | `TestTodoAutoDone_NegationAttributesNothing` (+subject-stream-only unit) | 016 | `TestTodoAutoDone_CloseSurfaceExclusivity` (both subtests) |
| 008 | `TestTodoAutoDone_InconclusiveNeverCloses` | 017 | `TestTodoAutoDone_FalseNegativeShapes` (+comma-form unit + corpus re-pin) |
| 009 | `TestTodoAutoDone_CloseLineContract` (incl. factory state) | DoD | `TestTodoAutoDone_HelpDocumentsContract` + mirror/diff checks below |

DoD doc-step verified directly: both todo.md mirrors byte-identical (`diff` → identical), lead-only [HARD] clause present in both, `git fetch origin develop && git rev-parse origin/develop` confirmation named, no lane named as runner; docs-site 4 locales each +4 lines (parity preserved); CHANGELOG entry cites the SPEC-ID and 17 ACs; `catalog.yaml` hash regenerated.

## Findings

- F1 [Low] [optional] `internal/cli/todo_autodone_test.go:225` — the decision-table row named "recorded SHA closes through a collision" builds facts via `shaReachableFacts`, which sets `DistinctTexts: 1`; the row's name promises the collision shape its facts do not carry. Behavior is right (form 1 closes before the gate is consulted) and the CLI dry-run test covers the true DistinctTexts=2 shape — but a mutant relocating the collision gate ahead of form 1 would survive this specific row. Required fix: set `DistinctTexts: 2` in that row.
- F2 [Low] [optional] `internal/cli/todo_autodone_test.go:445-452` — the comment (and progress.md §E.2) claims the queue record is "unchanged (Mutate's byte-identity contract)" after the identity-invariant refusal, but the test asserts only the refusal error, not record byte-identity. Required fix: capture `recordBytes` before the non-dry run and assert equality after the error.
- F3 [Low] [optional] `internal/cli/todo_autodone.go:543-574` — `appendAutoDoneReversal` names the LAST "closed" row for the id; after close→undone→manual done→undone, the second reversal row names the already-reversed first closure (stale `original_at`). Audit-log precision only, no correctness impact. Required fix: skip ids whose latest log row is already "reversed".
- F4 [Low] [optional] commits `a5c0c0eca`, `286a6713a` — the AGENTS.md §3 traceability carrier "card id in every commit message on the branch" is missed on 2 of 5 commits (SPEC-ID present in all 5, so the card is recoverable one hop away via spec.md frontmatter `author: ... (card t684)`). Required fix: none retroactively (no history rewrite); carry `(t684)` in any remaining commits on the branch.
- F5 [Info] [optional] `internal/cli/todo_autodone.go:297` — literal `"completed"` where `kanban.StatusCompleted` exists; neighborhood-consistent with `todo.go:735`'s identical literal, so no action required.
- F6 [Info] [optional] §D.2's "subagent-boundary guard extended to the new verb" — no dedicated static guard test names `todo_autodone.go`; satisfied in substance (zero AskUserQuestion hits grep-verified; the verb is non-interactive so the guard-extension trigger does not fire).
- F7 [Info] [optional] AC-AD-015's "evaluates against the post-fetch ref position" is proven structurally (code order fetch→rev-parse→scan, `todo_autodone.go:201-217`) plus counter==1, but the fixture has no configured remote, so the fetch fails and a SUCCESSFUL fetch that moves the ref is never exercised end-to-end. Required fix: a fixture with a real local remote would close this; current residual is low.

All findings are [optional]-class: none affects correctness or any requirement the SPEC states; none is blocking.

## Gaps (what this audit did NOT observe)

- Full-package `go test ./internal/cli/` and the whole-repo suite — machine-load discipline (bounded selectors only, per dispatch); the full-suite verdict belongs to CI on the develop push.
- CI verdict on origin/develop — lanes never push; the lead's batch push creates it. Not mine to run.
- A successful-fetch end-to-end (F7).
- docs-site hugo build/4-locale verify recipe (the +4-line diffs are parity-consistent by inspection; the site build runs on the docs harness, not in this audit).

## Residual risks

1. The negation guard is subject-stream-only; a body negation with a clean attributing subject still attributes (SPEC §D.1 recorded edge; SPEC-TODO-LANDING-ATTRIBUTION-001's domain).
2. M1's field reach: while the identity invariant holds, no store-writable record can carry DistinctTexts > 1, and the counter reads only the local record — a cross-STORE reissue (predecessor in another queue) presents DistinctTexts=1 locally, so M1 does not fire there. progress.md §E.3's "guards cross-store/divergent-queue reissue shapes" phrasing overstates the single-record counter's reach; the guard effectively protects legacy/corrupt/injected records. The scan still closes safely in that shape only when form-1 evidence exists.
3. Four READMEs enumerate todo verbs without auto-done (disclosed in §E.4 as outside module scope — follow-up card territory).
4. The stale-binary hazard is loud, not silent: a binary predating this SPEC fails with `unknown command "auto-done"`, so nothing can misfire quietly.

## Baseline attribution

All evidence rows above were measured in this audit run, in this worktree, at HEAD `0673b3ac0`, against diff base `74d872aaf` (merge-base with origin/develop). The AC verdicts are re-executed observations, not carry-overs from progress.md §E.

🗿 MoAI
