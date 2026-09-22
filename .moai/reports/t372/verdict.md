# SPEC-STRESS-INVARIANT-VERDICT-001 — Sync-phase Verdict

Card t372 · worktree `.claude/worktrees/t372` · branch `WT-stress-invariant-guard` ·
base `origin/develop` = `b9149857c` · run-phase HEAD at sync entry = `0fa8606fe`.

Two measuring parties are distinguished throughout, because the strongest evidence in this card is
an independent re-run of the binding mutant and folding it into a single anonymous "we measured"
would erase exactly what makes it independent:

- **[MD]** — `manager-develop`, run-phase, recorded verbatim at `.moai/reports/t372/run-evidence.md`.
- **[ORCH]** — the orchestrator, re-running AC-SIV-008's mutant itself, after and independently of
  [MD], against this same tree.

Every measurement below carries its party and the tree it was taken against. No figure is carried
over from another package, tree, or point in time. The t370 measurements
(`.moai/reports/t370/verdict.md`, `.moai/reports/t370/measurements.md`) are consumed as **given
ground truth** and are re-derived nowhere in this card.

---

## 1. Claim

`TestConcurrencyStress`'s verdict criterion has moved off lock acquisition and onto the four queue
invariants; acquisition latency now has its own **constant-coherence** guard,
`TestBoardLockWaitBudgetCoversSerializedMutations`. Both separated criteria were demonstrated
capable of turning RED under a planted mutant, each mutant was reverted, and each restoring GREEN
was recorded. 13 of 14 acceptance criteria are PASS; **AC-SIV-013 (the closure gate) is OPEN at
merge by design**, so the SPEC closes sync-phase at `implemented`, not `completed`.

**What this claim does not say** is set out verbatim in §5. In particular it does not say the CI
flake is fixed, does not compare against any pre-repair rate, and does not say the budget suffices
on any machine.

---

## 2. Evidence

### 2.1 Per-AC matrix

Party column names who ran the verification. `[MD]` rows are attributable to
`.moai/reports/t372/run-evidence.md` §2-§7 against run-phase start HEAD `3cd1a09f1`; the `[ORCH]`
row is this sync-phase re-run against `0fa8606fe`.

| AC | Covers | Status | Party | Verification + observed |
|---|---|---|---|---|
| AC-SIV-001 | REQ-SIV-001 | PASS | [MD] | `go test -race -count=1 -v -run '…' ./internal/kanban/` → `--- PASS: TestStressAddClassificationToleratesStarvation (3.37s)`; log `2/2 adds starved under a seeded holder, all satisfying IsBoardLockHeld, 0 hard failures` |
| AC-SIV-002 | REQ-SIV-002, REQ-SIV-003 | PASS | [MD] | `grep -n 'strings.Contains' internal/kanban/backlog_concurrency_test.go` → one hit, line 343, inside the AC-SIV-005 sub-test's **message** assertion. Classification (`classifyStressAdd`) decides on `err == nil` / `IsBoardLockHeld(err)` only — no text matching in the classification path |
| AC-SIV-003 | REQ-SIV-004, REQ-SIV-005 | PASS | [MD] | source read `backlog_concurrency_test.go:205-231` — four assertions anchored to `issuedCount := len(issued)`; `wantTotal` no longer exists in the file |
| AC-SIV-004 | REQ-SIV-006 | PASS | [MD] | `grep -n 't.Skip' internal/kanban/backlog_concurrency_test.go` → one hit, line 202, inside a **comment**; no `t.Skip` call, no starvation conditional guarding any of the four |
| AC-SIV-005 | REQ-SIV-007 | PASS | [MD] | `go test -race -count=1 -v -run '…' ./internal/kanban/` → `--- PASS: TestStressZeroProgressFloorFailsTotalStarvation (1.68s)`; zero-success outcome rejected, 1-success outcome admitted |
| AC-SIV-014 | REQ-SIV-008 | PASS | [MD] | source read + package `-race` run — `successes + starved + len(hardFailures) == stressWriters * stressAddsPerWriter`; no clock, no fraction, no percentage |
| AC-SIV-006 | REQ-SIV-009, REQ-SIV-011 | PASS | [MD] | `go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/` → `--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)`; success message claims coherence only (verbatim, run-evidence §4) |
| AC-SIV-007 | REQ-SIV-010 | PASS | [MD] | `sed -n '95,120p' internal/kanban/board_lock_wait_test.go \| grep -nE 'time\.(Now\|Since\|Sleep)\|go func'` → **no output**; floor built from the two stress constants, which the budget expression does not supply |
| **AC-SIV-008** | **REQ-SIV-013 latency direction** | **PASS — binding** | **[MD] + [ORCH]** | §2.2 below — measured **twice, independently**. [MD]: selector-scoped RED with the old guard GREEN in the same run, plus a whole-package confirmation. [ORCH]: whole-package run of 389 tests, exactly one FAIL |
| **AC-SIV-009** | **REQ-SIV-013 invariant direction** | **PASS — binding** | [MD] | §2.3 below — two reverted mutants; RED at invariant **(d)** on the first, at **(b)** + **(c)** on the second |
| AC-SIV-010 | REQ-SIV-012 | PASS | [MD] | the `TestConcurrencyStress` `t.Logf` line reports starved count + back-derived per-mutation cost; no `t.Error`/`t.Fatal` gated on either |
| AC-SIV-011 | REQ-SIV-014 | PASS | [MD] + [ORCH] | run-evidence §8 states all four limits; §5 of **this** report restates them in substance |
| AC-SIV-012 | REQ-SIV-016 | PASS | [MD] | `git diff --stat HEAD` taken after both mutants reverted → 3 files; `board_store.go` comment-only, verified by a non-comment-line grep returning empty |
| **AC-SIV-013** | REQ-SIV-015 | **OPEN at merge** | — | requires ≥5 non-cancelled post-landing develop heads descended from the landing commit. Those heads do not exist. Deliberately unclaimed — see §5 clause 2 |

Counts: **13 PASS · 0 FAIL · 1 OPEN**, against 14 acceptance criteria counted from
`acceptance.md` (`grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' … | sort -u | wc -l` → `14`, non-zero).

### 2.2 AC-SIV-008 — measured twice, independently

**[MD], run-phase** (`run-evidence.md` §5, tree `3cd1a09f1`). Census taken **before** the mutant was
planted (`grep -rn 'boardLockHeadroom' internal/kanban/`, three test files cover the constant).
Mutant on the constant axis, `boardLockHeadroom = 5` → `4`. Selector-scoped run:

```
$ go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
    board_lock_wait_test.go:103: constant coherence broken: … (1.32s budget < 1.584s floor) …
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.461s
```

Selector match count 2 (two `=== RUN` lines), non-zero. The old guard GREEN in the same run is what
attributes the RED to the new guard alone.

**[ORCH], sync-phase, this tree `0fa8606fe`** — an independent re-run, deliberately widened from a
selector to the **whole package**, so that every guard in the census had a chance to fire rather
than only the two the `TestBoardLockWaitBudget` prefix selects.

Guard census first (the pre-plant enumerating command, run by [ORCH]):

```
$ grep -rln 'boardLockHeadroom\|boardLockWaitBudget' --include='*_test.go' internal/
internal/kanban/integration_lock_cross_test.go
internal/kanban/backlog_concurrency_test.go
internal/kanban/board_lock_wait_test.go
```

Three files. Mutant planted (`boardLockHeadroom` 5 → 4), then:

```
$ go test -race -count=1 -v ./internal/kanban/    # exit 1
```

- **Swept set: 389 `=== RUN` lines** (`grep -c '^=== RUN'` on the captured stream). Non-zero and
  package-wide — an empty or selector-narrowed sweep cannot masquerade as this observation.
- **Exactly ONE test failed**, by name:

```
=== CONT  TestBoardLockWaitBudgetCoversSerializedMutations
    board_lock_wait_test.go:103: constant coherence broken: the lock policy budgets 10 supported writers x 4 headroom = 40 serialized mutations, while the stress test serializes 8 x 6 = 48 (1.32s budget < 1.584s floor). Lowering either policy constant, or raising either stress constant past that product, fails this guard. The per-mutation cost cancels on both sides, so the relation is cost-independent and asserts nothing about the wait any real machine needs — the CI -race per-mutation cost observed by t370 was 42-105ms against the declared 33ms.
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
```

- **`TestBoardLockWaitBudgetDerivedFromNamedInputs` stayed `--- PASS` in the same run**
  (`--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)`). **That is the attribution**:
  the two guards share a `TestBoardLockWaitBudget` prefix, so a bare RED would not say which one
  discriminated.
- Mutant reverted; both guards `--- PASS` again; tracked tree clean afterwards
  (`git status --short internal/` empty).
- Baseline targeted run before mutation: 5 `=== RUN`, all PASS, back-derived per-mutation cost
  **14.8 ms** (elapsed 710 ms / 48 mutations).

Evidence stream: `.moai/reports/t372/mutant-headroom4-orchestrator.log` — **gitignored** by
`.gitignore:106 (*.log)` and therefore worktree-local, not committed. The load-bearing lines are
reproduced verbatim above precisely so the claim survives the file not being re-openable at audit
time; this is recorded as a Gap in §6, not papered over.

### 2.3 AC-SIV-009 — invariant direction ([MD], run-evidence §6)

Two mutants, both scale-gated to engage only at the 48-add fan-out, both planted in
`internal/kanban/backlog_store.go` and both reverted.

Mutant 1 — upward `last_seq` advance above the item count:

```
    backlog_concurrency_test.go:228: invariant (d) mark consistency: last_seq = 56, want 48 (distinct issued ids)
--- FAIL: TestConcurrencyStress (0.75s)
```

Mutant 2 — an item dropped after its id was issued (lost update):

```
    backlog_concurrency_test.go:218: invariant (b) no lost update: id t2 was issued but is not in the queue
    backlog_concurrency_test.go:223: invariant (c) count consistency: stored items = 44, want 47 (distinct issued ids) — 3 updates were lost
    … 47 succeeded, 1 starved (tolerated), 0 hard failures …
--- FAIL: TestConcurrencyStress (2.33s)
```

Mutant 2 is the single strongest observation in the card, and it was not staged: the run happened
to produce a genuine starved add, and **the invariants fired anyway**. Starvation tolerated *and*
the invariant criterion red, in the same run, is the direct demonstration that the two criteria
were separated rather than that the rule was switched off. Selector match count 1 on each run;
whole-package runs under each mutant fired exactly one test of the twelve-file census; both
reverted with restoring GREEN recorded.

### 2.4 Sync-phase checks run by [ORCH] on this tree

| Check | Command | Result |
|---|---|---|
| duplicate-entry guard (B12-1) | `grep -c 'SPEC-STRESS-INVARIANT-VERDICT-001' CHANGELOG.md` | `0` — no prior entry, emission permitted |
| AC count (B12-2) | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' …/acceptance.md \| sort -u \| wc -l` | `14`, non-zero — matches the 14 rows in §2.1 |
| README / docs-site reach | `grep -rln 'TestConcurrencyStress\|boardLockWaitBudget\|boardLockHeadroom' README*.md docs-site/` | **no hits** — no user-facing surface names any changed symbol |
| changed paths | `git diff --stat b9149857c..0fa8606fe` | 3 source files (`backlog_concurrency_test.go`, `board_lock_wait_test.go`, `board_store.go` +9/-0 comment-only) + 5 SPEC/report artifacts |

---

## 3. Baseline-attribution

| Figure | Command + observed output | Party | Tree |
|---|---|---|---|
| 13 PASS / 1 OPEN AC matrix | the per-row commands in §2.1 | [MD] | `3cd1a09f1` |
| package suite green, unmutated | `go test -race -count=1 ./internal/kanban/` → `ok … 24.890s` | [MD] | `3cd1a09f1` |
| vet clean | `go vet ./internal/kanban/...` → no output, exit 0 | [MD] | `3cd1a09f1` |
| gofmt clean | `gofmt -l internal/kanban/` → no output | [MD] | `3cd1a09f1` |
| AC-SIV-008 RED, package-wide, 389 swept, 1 FAIL | `go test -race -count=1 -v ./internal/kanban/` under `boardLockHeadroom` 5→4 → exit 1, one `--- FAIL` | **[ORCH]** | `0fa8606fe` |
| per-mutation cost 14.8 ms | back-derived, 710 ms elapsed / 48 mutations, local darwin `-race` | **[ORCH]** | `0fa8606fe` |
| per-mutation cost 15.696085 ms | `TestConcurrencyStress` `t.Logf`, local darwin `-race` | [MD] | `3cd1a09f1` |
| CI `-race` band 42-105 ms; 12-of-14 red; 0-of-14 invariant reds | cited, not measured here | t370 | `origin/develop` = `1e5199b88` |

Both local per-mutation figures — [ORCH]'s 14.8 ms and t370's 17.5 ms local darwin figure — sit far
below the 34.4 ms threshold at which the budget is exhausted. **This machine was always in the
passing band.** The 42-105 ms CI `-race` band was never reproduced locally, and no load was
generated to reproduce it (`spec.md` §E puts load generation out of scope).

---

## 4. Status transitions and artifacts carried by the sync commit

- `spec.md` frontmatter `status: in-progress → implemented` and `updated:` refreshed.
  **NOT `completed`** — AC-SIV-013 is open (§5 clause 2).
- `plan.md` and `acceptance.md` carry no frontmatter and are status-stateless by schema
  (`spec-frontmatter-schema.md` § Artifact Statelessness); neither is touched.
- `progress.md` §E.4 Sync-phase Audit-Ready Signal populated. `sync_commit_sha` is recorded as
  `pending-backfill-sync` — a commit cannot cite its own hash — and is backfilled in the
  immediately following commit.
- No SPEC body content (§A-§F of `spec.md`, any section of `plan.md` / `acceptance.md`) was
  modified at sync. No implementation source file was modified at sync.
- MX Tag validation performed as a sync sub-step: no `@MX:` annotation was added or altered. The
  change is confined to `_test.go` files plus a comment; the MX tag families
  (`NOTE`/`WARN`/`ANCHOR`/`TODO`) address exported production surface, and no exported production
  function was added or changed.

### CHANGELOG decision — **emit**, under `### Changed`

The change is test-only plus a comment, so the question is real rather than automatic. It is
emitted, on this repository's own precedent and for one reason beyond precedent:

- **Precedent.** `SPEC-CI-FLAKE-SERIES-001` (card t278) is three pure test-only flake fixes and
  carries a full `[Unreleased]` entry. `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354) — this card's
  direct predecessor — carries one under `### Fixed`. Test-only closure is not an exemption here.
- **Beyond precedent.** This card changes *what a test means*. A future reader of a green
  `TestConcurrencyStress` needs to know that green no longer asserts what it asserted before t354's
  successor landed — starvation is now tolerated. That is exactly the kind of thing a changelog
  exists to record, and it is invisible in the diff to anyone not reading the SPEC.

**Section `### Changed`, not `### Fixed`.** `Fixed` would assert the flake is fixed. It is not
established as fixed: AC-SIV-013 is open, and §5 clause 1 forecloses any before/after claim. What
changed is the verdict criterion — a `Changed` fact — so `Changed` is the honest section.

---

## 5. Non-claims (REQ-SIV-014, all four, carried verbatim in substance)

1. **No before/after improvement claim exists, in any quantity.** t370 never measured the
   **pre**-repair firing rate, and it is unrecoverable. No statement of the form "the flake rate
   improved from X to Y", at any magnitude or with any hedge, is available. The strongest sentence
   this card supports is: *the verdict criterion moved to the invariants, and under that criterion
   it is green.*

2. **A single green run cannot close this card.** Two green `-race` runs already existed
   post-repair — one green `Race Test` job (`51daada00`) and one run in which `TestConcurrencyStress`
   was green inside a job reddened by a **different** test (`c6aa61346`); calling the second a green
   run overstates the source. **This is exactly where card t354 stopped.** Closure requires
   AC-SIV-013 — `TestConcurrencyStress` green under the invariant criterion across **≥5**
   non-cancelled develop heads descended from the landing commit — which stays **OPEN at merge**.
   The SPEC therefore ends sync-phase at `implemented`, never `completed`.

3. **A green observation window does not evidence that the invariants still fire.** The invariant
   criterion was red in **0 of the 14** observed runs *before* this change, so a post-landing window
   of greens under the new criterion is fully consistent with the invariants having been switched
   off. It evidences only that no *new* failure mode was introduced. The burden of showing the
   invariants still fire belongs solely to the §2.3 mutant evidence, and no number of green heads
   substitutes for it.

4. **The tolerated error class is wider than contention, on the CI platform.**
   `internal/kanban/board_lock_unix.go:41-43` maps **every** `unix.Flock` failure to a bare
   `ErrBoardLockHeld` — `ENOLCK`, `EINTR`, and `EBADF` included, not only `EWOULDBLOCK`. The Windows
   substrate is narrower by contrast (`board_lock_windows.go` discriminates via `os.IsExist`), so
   the tolerance is widest on exactly the platform the CI `Race Test` job runs. Separately,
   `IsBoardLockHeld` is `errors.Is`, which traverses the `errors.Join(mutErr, relErr)` that `Mutate`
   returns, so a future joined error carrying a lock-held branch alongside a real defect would be
   tolerated wholesale. **This is the largest weakening the change introduces.** It is recorded as
   an out-of-scope non-claim and a follow-up candidate (§7), and is deliberately **not** repaired
   here — narrowing the sentinel is production behaviour.

Fifth, smaller, stated for completeness: **the REQ-SIV-007 zero-progress floor admits 1 success in
48**, where t370 measured real starvation at 3-7 of 48 — roughly 40× weaker than observed
behaviour. The REQ-SIV-008 conservation identity is the compensating control. A fractional or
percentage floor was **deliberately refused**: it would be a load sensor in accounting clothing and
would recreate the flake on the next slower runner.

**Local green is not the verdict.** [ORCH]'s measured local cost is 14.8 ms and t370's local figure
17.5 ms, both far below the 34.4 ms threshold — this machine was always in the passing band. The
42-105 ms CI band was never reproduced locally and no load was generated to reproduce it.

---

## 6. Gaps — what was explicitly NOT observed

- **No CI run was read or triggered at sync.** Every figure in this report is local darwin. The
  `Race Test` job on `origin/develop` has not run against this tree, and reading CI was outside the
  delegated scope.
- **No full-repository suite.** Verification is scoped to `internal/kanban` per this repository's
  local-full-suite prohibition. The whole-package verdict for the rest of the tree belongs to CI.
- **No `golangci-lint` run**, at run-phase or at sync. `go vet ./internal/kanban/...` and
  `gofmt -l` are the only static checks executed (both clean, [MD]).
- **Cross-platform build not exercised.** `GOOS=windows` / `GOOS=linux` build and vet were run at
  neither phase. The change being test-only plus a comment is an argument, not an observation.
- **[ORCH]'s mutant evidence stream is gitignored.** `.moai/reports/t372/mutant-headroom4-orchestrator.log`
  matches `.gitignore:106 (*.log)` and is not in the commit; it exists only inside this worktree and
  will not survive its disposal. The decisive lines are reproduced verbatim in §2.2 for that reason,
  but the raw 389-test stream is not re-openable at audit time by any later reader.
- **`.moai/reports/t370/**` — the cited ground truth of this entire SPEC — is untracked.** It is
  absent from `origin/develop` and from all git history (`git log --all -- .moai/reports/t370` →
  empty). Copies exist in the primary checkout, so the immediate loss risk is low, but until card
  t370's own lane commits them, `spec.md` §A and this report cite paths that resolve for nobody
  else. **Not staged here** — they are another card's artifacts, and committing them under this
  card's SHA would misattribute their provenance. Reported to the lead instead (§7 item 4).
- **The pre-repair firing rate was not measured** and is unrecoverable (§5 clause 1).
- **AC-SIV-013 is unmeasured by construction** — it needs post-landing develop heads that do not
  yet exist.
- **The 48-add fan-out under real starvation was observed once, incidentally** (§2.3 mutant 2, 1
  starved of 48). No run was staged to produce starvation at the parent test's scale.
- **The AC-SIV-009 mutants were scale-gated.** Ungated versions of the same shapes were not run, so
  the behaviour of the other eleven census files under an ungated mutation is unobserved.
- **The sync-phase quality gate was not re-run after the CHANGELOG edit.** This report's §2.4 checks
  were taken before the sync commit's own content existed.

---

## 7. Residual risk and follow-up candidates (recorded, NOT fixed)

Risks:

- **The tolerated class is wider than contention** (§5 clause 4). A future `ENOLCK`/`EINTR`
  regression, or a joined error carrying a lock-held branch alongside a genuine defect, is absorbed
  as starvation and the test stays green.
- **The zero-progress floor is ~40× weaker than observed starvation** (§5, fifth clause).
- **The constant-coherence margin is 4.2%** (66 ms on 1.584 s). The guard is a narrow tripwire, and
  it is cost-independent by construction — **no per-mutation cost regression of any size can make it
  fire**. Claiming it protects the latency budget is the overclaim to avoid.
- **`integration_lock_cross_test.go` sizes a 500 ms release timeout against `boardLockWaitBudget` in
  a comment.** It did not fire under either party's `headroom 5→4` mutant (budget 1.32 s, 500 ms =
  37.9%), but a deeper budget reduction could make that timeout marginal with no guard saying so.
- **Local green is not the verdict.** The CI `Race Test` job on `origin/develop` is.

Follow-up candidates, each accepted by the lead as a separate card and **deliberately not repaired
in this SPEC**:

1. **A pre-existing vacuous guard.** `internal/kanban/board_lock_wait_test.go:36-41` computes
   `floor` from the **identical expression** as `recomputed` two lines above
   (`time.Duration(boardLockSupportedWriters) * boardLockCIMutationCost * boardLockHeadroom`), and
   `boardLockWaitBudget == recomputed` has already been asserted — so `if boardLockWaitBudget < floor`
   is **unreachable for any input values**. It is pre-existing t354-era code. This card replaced that
   vacuity with a real inequality in the NEW guard and deliberately did not touch the old branch
   (`spec.md` §E, Out of Scope — repairing the existing derivation guard).
2. **The over-broad `ErrBoardLockHeld` sentinel on Unix** (§5 clause 4). Narrowing it is production
   behaviour with its own regression surface.
3. **A naming friction.** `plan.md` prescribes the guard name
   `TestBoardLockWaitBudgetCoversSerializedMutations` **verbatim**, while REQ-SIV-009 forbids
   "covers" framing in the guard's *messages* — which comply. A CI reader nonetheless sees "Covers"
   on a failure line, which reads as a coverage claim the guard does not make. **Not renamed
   unilaterally**: the name is contract text. Recorded for the lead's judgment.
4. **`.moai/reports/t370/**` is untracked everywhere** (§6). It is this SPEC's cited SSOT. Card
   t370's lane should commit it, or the lead should rule on adopting it.

---

## 8. Report artifacts staged in the sync commit, and why

| Artifact | Staged? | Reason |
|---|---|---|
| `.moai/reports/t372/verdict.md` | **yes** | this card's sync verdict — the deliverable |
| `.moai/reports/t372/plan-audit.md` | **yes** | this card's own plan-audit, cited by `spec.md` §F and by HISTORY 0.2.0/0.3.0; untracked before this commit, so the citation would otherwise resolve nowhere |
| `.moai/reports/t372/mutant-headroom4-orchestrator.log` | **no** | gitignored by `.gitignore:106 (*.log)`. Not force-added — the ignore rule is the repository's, not this card's, to override. Decisive lines reproduced verbatim in §2.2; the loss is recorded as a Gap (§6) |
| `.moai/reports/t370/{verdict,measurements,preflight-t354-overlap}.md` | **no** | another card's artifacts. Committing them under this card's SHA would misattribute provenance. Reported as §7 item 4 |

`.moai/reports/t372/run-evidence.md` was already committed by run-phase (`0fa8606fe`).

---

## 9. Post-sync 리드 판정 반영 (창 대기 중 추가)

sync 커밋 이후 리드가 §7 의 판정 요청 3건에 답했다. 그 판정과, 판정이 요구한
재현 근거를 여기에 남긴다. 이 절의 관측은 전부 **오케스트레이터가 이 트리에서
직접 실행**한 것이다(트리 `.claude/worktrees/t372`, 브랜치 `WT-stress-invariant-guard`,
측정 시점 HEAD `bdbd3c3f4`).

### 9.1 기준점 인용 결함 — 자기 보고

통합 요청 메시지에 `origin/develop = c36d2672b` 라고 적었으나, **이 값은 내가 잰 것이
아니라 리드의 kickoff 메시지 본문에서 옮겨 적은 것**이다. 출처 라벨 없이 옮긴 순간
내 측정처럼 읽혔고, 리드가 그렇게 읽고 정정했다.

재측정:

```
$ git fetch origin develop && git rev-parse --short origin/develop
94a1b0552
$ git rev-list --count --left-right origin/develop...HEAD
33	8
$ git cat-file -t c36d2672b
commit
$ git merge-base --is-ancestor c36d2672b origin/develop && echo YES
YES
```

`c36d2672b` 는 **존재하지 않는 SHA 가 아니라 낡은 SHA** 다 — `origin/develop` 의
조상이다. 그래서 더 위험하다: 존재하지 않는 값은 즉시 터지지만, 조상인 낡은 값은
조용히 통과한다. 결함은 값이 낡은 것이 아니라 **건네받은 값과 잰 값을 구별하지 않고
인용한 것**이다.

리드의 `b64043481` 도 이미 낡았다(내 fetch 시점 `94a1b0552`). 창 순번이 오는 시점에
다시 잰다.

### 9.2 판정 1 — 가드 이름 유지, 마찰은 기록

리드 판정: `TestBoardLockWaitBudgetCoversSerializedMutations` **유지**. REQ-SIV-009 이
금지한 것은 가드 **메시지**의 "covers" 표현이고 테스트 이름은 메시지가 아니므로,
현행 이름은 요구사항 문면을 위반하지 않는다. `plan.md` 가 이름을 못박은 이상 단독
개명은 아티팩트와 코드를 어긋나게 만들며, 개명하려면 `plan.md` 부터 고쳐야 하고
그것은 이 카드 범위 밖이다.

**남는 마찰(기록)**: CI 로그를 읽는 사람은 실패 줄에서 "Covers" 를 본다. 메시지 본문은
소거 사실을 명시하므로 오해가 정정되지만, 실패 줄만 스캔하는 독자에게는 "예산이
충분함을 검사한다"로 읽힐 여지가 남는다. 개명하려면 `plan.md` §F M1 step 2 와 함께
고쳐야 한다.

### 9.3 판정 2 — 후속 후보 2건, 재현 근거 포함 (수집만, 발행은 운영자)

#### 후보 A — 구조적으로 붉어질 수 없는 선행 가드

`internal/kanban/board_lock_wait_test.go` 의 `TestBoardLockWaitBudgetDerivedFromNamedInputs`
안에서, `floor` 가 12 줄 위 `recomputed` 와 **완전히 같은 식**이다:

```
$ sed -n '26,41p' internal/kanban/board_lock_wait_test.go
	recomputed := time.Duration(boardLockSupportedWriters) *
		boardLockCIMutationCost * boardLockHeadroom
	if boardLockWaitBudget != recomputed {
		...
	}
	floor := time.Duration(boardLockSupportedWriters) *
		boardLockCIMutationCost * boardLockHeadroom
	if boardLockWaitBudget < floor {
		t.Errorf("budget %v < headroom floor %v", boardLockWaitBudget, floor)
	}
```

`budget == recomputed` 는 바로 위에서 이미 단언됐고 `floor == recomputed` 이므로,
`budget < floor` 는 **어떤 입력값 조합으로도 참이 될 수 없다.** 이 `t.Errorf` 는
도달 불가다.

관측된 뒷받침 — `boardLockHeadroom` 을 5→4 로 낮춘 뮤턴트(예산이 실제로 얇아진 상태)
에서도 이 테스트는 초록으로 남았다:

```
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
```

(전문: `.moai/reports/t372/mutant-headroom4-orchestrator.md`)

성격: **가드가 있는데 구조적으로 못 붉어지는** 형태 — t241 계열의 정확한 사례다.
이 카드는 새 가드에서 floor 를 예산 식에 없는 항(직렬화 변이 수 48)에서 뽑아
공허성을 재생산하지 않았지만, **옛 분기는 그대로 남았다**(t354-era 코드, 범위 밖).

#### 후보 B — Unix 에서 경합보다 넓은 관용 술어

```
$ sed -n '41,44p' internal/kanban/board_lock_unix.go
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return nil, ErrBoardLockHeld
	}
$ grep -n -A2 'func IsBoardLockHeld' internal/kanban/board_lock.go
28:func IsBoardLockHeld(err error) bool {
29-	return errors.Is(err, ErrBoardLockHeld)
30-}
```

`unix.Flock` 이 **어떤 errno 로 실패하든** 맨 `ErrBoardLockHeld` 로 매핑된다 — 경합
(`EWOULDBLOCK`)뿐 아니라 `ENOLCK`·`EINTR`·`EBADF` 도 같다. Windows 쪽은
`os.IsExist` 로 구분해 좁다(`board_lock_windows.go`). **CI 가 도는 쪽이 넓은 쪽이다.**
게다가 `IsBoardLockHeld` 는 `errors.Is` 라 `Mutate` 가 반환하는 `errors.Join` 을
관통하므로, 락-held 분기와 실제 결함이 함께 묶인 에러가 미래에 생기면 흡수된다.

이 카드가 관용 클래스를 판정에서 빼면서 **이 폭이 그대로 관용 범위가 됐다** — 이
변경이 들여온 가장 큰 약화이고, 수리하지 않고 비주장으로만 기록했다(§5).

### 9.4 판정 3 — t370 증거 착지 (지시대로 별도 커밋)

리드 판정: 인용이 영구히 해소되지 않는 것이 출처 오귀속보다 큰 결함이므로, **커밋
메시지가 출처를 명시하면 오귀속이 아니다.** t370 보고 3파일을 이 병합에 포함하되
별도 커밋으로 분리했다.

```
$ git log --all --oneline -- .moai/reports/t370   → (빈 출력)
$ git ls-files .moai/reports/t370 | wc -l          → 0
$ cmp <primary>/.moai/reports/t370/verdict.md .moai/reports/t370/verdict.md
(무출력 = 바이트 동일; measurements.md · preflight-t354-overlap.md 도 동일)
```

착지 커밋: `6b9c74a3e` — `docs(t370): land investigation evidence cited by t372
(originating card t370, no live lane)`. 카드 커밋과 섞지 않았다.

**규약이 답해야 할 질문(lane-8 t375 소관)**: 발행 레인이 없는 카드의 증거를 누가
착지시키는가. 이 사례의 형태는 — 저작은 t370, 착지는 t372, 출처는 커밋 메시지가
보존, 카드 커밋과 분리. 일반화 여부는 그 규약의 몫이다.

### 9.5 리드가 승인한 구조 — AC-SIV-013

`implemented` 에서 멈추고 `completed` 로 가지 않는 것, 그대로 유지한다. **이 카드는
병합만으로 종결되지 않는다.** 종결 선언은 리드가 관측 창(착지 후 develop head 5개
이상)을 채운 뒤 한다. 로컬 초록이 근거가 아니라는 것(실측 14.8ms, CI 붉음 42~105ms
미재현·미시도)도 그대로 유지한다.
