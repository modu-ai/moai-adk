# SPEC-CAPACITY-RECLAIM-001 — Progress

status: draft
plan_status: audit-ready
plan_complete_at: 2026-10-14T02:03:04Z

## §E.1 Plan-phase Audit-Ready Signal

Plan artifacts (spec.md, plan.md, acceptance.md, decision-index.md) are revised to SPEC v0.2.0 on card worktree HEAD `e2658f30127375e75e0c606d09ca97be2c551cf2` (short `e2658f3012`, branch `WT-capacity-reclaim`). This is the FINAL authorized repair round: leader decision d-20261010T080627Z-987a allows exactly one repair inside fix_scope and one final delta re-audit after it; a FAIL on that audit returns the card to the operator.

Basis. The revision answers the iter3 verdict (`.moai/reports/t1631/plan-audit-iter3.md`: FAIL 0.56, 11 blocking, MP-8) restricted to card A's subset of the operator's narrowed scope (decision recorded on the card queue record, 2026-10-10): D1 (AC-CAP-008's approved-branch four-element cell EV-007, command unescaped, verbatim output, exit per the §0 rule, tree pin), D2 (dissolved — its criterion moved to card B with the sweep scope), D3 (AC-CAP-004 restored to RB with a four-element cell; the G-EXIT premise withdrawn), D4 (the §0 exit-capture rule restated from this session's runner observations; the `bash -c` wrapper withdrawn; no "runner-captured" label without a runner-printed status), D7 (plan §D.1 rotation mechanics deleted; retention ownership and the ordering guarantee stated instead), D8 (new AC-CAP-009: the archive-before-rewrite characterization test on the default branch), D9 (new AC-CAP-012: the SessionStart split behavior), D11 (§F absent-key residual rewritten: no notice exists in that case). D5, D6, and D10 left this SPEC by excision, not by rewrite: the worktree-sweep scope (disposal narrowing, last-activity field, report fields, `moai worktree list`, bulk-run evidence, 10 GiB target) moved to card B; the doctor cap read-path (`internal/cli/doctor_disk.go:155`, `:261`) and the compiled-constant removal moved to card D; the inbox rotation moved to card t1653.

Absorb and pins. The card worktree merged local `main` (merge commit `e2658f3012`, 689 commits, zero conflicts, tree clean) and every pin this SPEC carries was re-measured on that tree in this session. Of the measured paths, only `internal/config/defaults.go` (constants moved 451→456, 455→460) and the template `workflow.yaml` (hygiene mode line 318→331) differ from the iter3-audited tree `ab9f5a6d3`; all other cited lines reproduce unchanged. The template `.gitignore` still holds no rule for the state files or archives; the five state files are still tracked, now totalling 132,585,731 B across 191 files.

Scope addition flagged for operator ratification. The merged tree carries a new tracked monthly archive (`learning-history/archive/2026-09.jsonl.gz`, 2,142,019 B; six `.gz` archives now tracked, 4,929,798 B total), which pushed the post-untracking residual to 5,230,544 B — above the operator's 5,000,000-byte completion criterion 1. REQ-CAP-010 (new) adds the archive untracking (local copies kept; template rule `.moai/harness/learning-history/archive/*.jsonl.gz`), restoring the criterion (residual 300,746 B). This widens the M2 untracking set from five to eleven files and was not in the operator's enumerated list; it is forced by the operator's own PASS definition on the merged tree and is recorded in spec §A v0.2.0, §E A14, and §F.

Q3. The clone-measurement gate is satisfied: the per-clone recipe was executed on 2026-10-10 and recorded at `.moai/reports/t1631/q3-clone-measurement.md`; the untracking is authorized by the narrowed-scope decision. B1 lands through the card's normal landing path.

Counts. 10 requirements (REQ-CAP-001–010) and 13 acceptance criteria (AC-CAP-001–013), against the Tier M ceilings of 16 and 16. AC renumbering and the v0.1.2→v0.2.0 mapping are recorded in acceptance.md §0.

v0.2.1 — the authorized one-clause D12 repair (operator decision d-20261014T025128Z-0aa1, kind ceiling-exception: repair only D12, then exactly one re-audit; another FAIL returns the card to hold): acceptance.md AC-CAP-012 clause (1) now seeds one `verify` residue-class target in the fixture project and requires its `"unit":"gc"` / `"mode":"apply"` audit row, making the cache/GC apply half observable (D12, `.moai/reports/t1631/plan-audit-final.md`); premise cells EV-015 (no hook fixture seeds the class) and EV-016 (the GC's decision-gated apply-row write, `gc.go:395–404`) were observed on this tree `111493b71`, EV-008 re-observed there unchanged, and the evidence carriers they cite were updated accordingly. Nothing outside the clause changed: REQ-CAP-010 and every other REQ, AC, and milestone are untouched, no `internal/` code changed, and the REQ-CAP-010 ratification remains with the operator. The re-audit is the next act.

Plan-audit outcome (2026-10-14): the single authorized re-audit PASSED — score 0.86 against the Tier M threshold 0.80, zero blocking defects, MP-1..9 all pass, codex required backend `pass` (receipt `rcpt-3e5c619ef067d45b28756f2f`), D12 mutant closed. Verdict: `.moai/reports/t1631/plan-audit-final-2.md`. Skip-eligibility: plan-artifact hash `8ecff932e63f1cb5c708ba7b2166c52ef1c3a104ca959f80519545d22a76690c` stable at HEAD `e993fff98999e66a549971017b04d3e449725177`. Run entry authorized by decision d-20261014T025128Z-0aa1 ("then exactly one re-audit"; PASS advances the card).

## §E.2 Run-phase Evidence

Run entered on tree `e993fff98999e66a549971017b04d3e449725177` (short `e993fff989`, branch `WT-capacity-reclaim`, clean except pre-existing progress.md plan-phase edit). All commands below ran in this worktree from its root; exit codes are the runner observations described in acceptance.md §0 (this runner prints `Exit code N` for a non-zero bare invocation and `(Bash completed with no output)` for empty output; below, "exit" labelled *convention* means the runner showed no status line).

### Pre-flight measurements (plan §C, run as written)

| Command | Observed output (verbatim) |
|---|---|
| `git rev-parse HEAD` | `e993fff98999e66a549971017b04d3e449725177` |
| `git status --short` | ` M .moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md` (plan-phase edit, pre-existing) |
| `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| tail -1` | `132585731 total` |
| `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive` | the eleven paths of EV-002, reproduced identically |
| `grep -n -E 'usage-log\|lessons-inbox' internal/hygiene/sinks.go` | `90:` lessons-inbox owner exception; `91:` usage-log owner exception (matches AC-CAP-006/007 green premises) |
| `grep -rn -E 'DefaultProfileMaxBytes' internal` | 5 lines: `defaults.go:460` (const), `doctor_disk_test.go:183`, `clean_home.go:243`, `doctor_disk.go:155`, `doctor_disk.go:261` |
| `grep -n -E 'mode: (report\|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` | `331:        mode: report              # report \| apply (auto path only)` (EV-007 reproduced) |
| `grep -rn 'TestRetentionPruneArchivesBeforeRewrite' internal` | (empty) — EV-006 re-observed, exit 1 by convention, runner showed `(Bash completed with no output)` |
| `grep -rn 'TestSessionStartAutoPathSplit' internal` | (empty) — EV-008 re-observed on `e993fff989` (superset of the `111493b71` pin; no `internal/` path differs), exit 1 by convention |
| `grep -rn 'profile_max_bytes' internal/hook` | (empty) — EV-009 re-observed, exit 1 by convention |

Additional premises re-observed: `grep -n -E 'usage-log\|lessons\|corrupt' internal/template/templates/.gitignore` → (empty), exit 1 by convention (EV-001 reproduced). Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (verbatim tails: `BUILD-MAIN-OK` / `BUILD-WINDOWS-OK` echoed after each).

Affected-family test baseline (before any edit, this tree): `go test ./internal/harness/ -count=1` → first run FAIL, then three consecutive `ok` runs (the sole first-run failure coincided with the template suite running concurrently in a parallel shell; judged machine contention, not a code defect — green on repeated runs per the concurrency-testing discipline). `go test ./internal/template/ -count=1` → FAIL, exactly `--- FAIL: TestRuleDateProvenance` (`.claude/rules/moai/core/hooks-system.md:368` date-provenance leak) — **pre-existing baseline red**, unrelated to this SPEC's template edits (`.moai/config/sections/*` and `.gitignore`), left untouched per scope discipline. `go test ./internal/hook/ -count=1` → `ok` (316.7s). `go test ./internal/hygiene/ -count=1` → `ok`.

### M0/B0 baseline (measured pre-B0 on `e993fff989`; recorded as the B0 progress.md-only commit, which precedes the M2 untracking commit B1)

The numbers below were measured before any tracking change, from the worktree root.

- `wc -c .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl` → `90331575 .moai/harness/usage-log.jsonl` / `6681068 .moai/lessons-inbox.jsonl` (`97012643 total`).
- `wc -c .moai/harness/learning-history/archive/*.jsonl.gz .moai/harness/learning-history/archive/*.corrupt` → 2026-04: `149`; 2026-05: `12995`; 2026-06: `1371`; 2026-07: `874561`; 2026-08: `1898703`; 2026-09: `2142019`; 2026-07.corrupt: `4179609`; 2026-08.corrupt: `14398585`; 2026-09.corrupt: `11764350`; `35272342 total`.
- `shasum -a 256` (three `.corrupt`): `0714d958ee1e53ba5f7de039f078f939629c75dfacebd99fc9eea638483b6e66  .moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt`; `61e77c0040d4206270d6905b4d9c9e6717eb8908639d212624b369c50640dc19  .moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt`; `8d0df96f247642307f543abec72264f02666cd98280ad53fc21dc17b969af687  .moai/harness/learning-history/archive/2026-09.jsonl.gz.corrupt`.
- `shasum -a 256` (six `.jsonl.gz`): `a9dc51bfb65410b29fda88f4193a786b1ce0b5eb516f790fd33abeb19ded7d0c` (2026-04); `3f9b775f4b6dcbdbc5c5e652fc90ba43c3af4888efcee285dead59af34b71f31` (2026-05); `f0cd00b598ac35cb8b02b7027ada85906d543f86c5cf674bf7c8da28f3bff053` (2026-06); `373f749fe95d1d3b8e64871f2a32a8f8fe78164ed65a7f02205314a92f3812ac` (2026-07); `0347b1a115ac00343725c938a5737c5e713055a897e5ab03760fca0e1e396172` (2026-08); `4be02485537e55925c973d300cbb1cab276429574273f39645b92dcafea80e9b` (2026-09).
- `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive` → identical to EV-002 (eleven paths).
- B0 SHA: recorded below after the commit lands.

### RED / mutant observations (E8; captured BEFORE the implementation that flips each)

- **EV-006 premise + mutant (AC-CAP-009).** Absence premise re-observed (table above). Test `TestRetentionPruneArchivesBeforeRewrite` written against the unchanged prune → `ok github.com/modu-ai/moai-adk/internal/harness 0.842s` (both subtests PASS — characterization of the default branch, as planned). Mutant: `prune` reordered to `overwriteWithEvents` before `archiveEvents` → FAIL, verbatim: `--- FAIL: TestRetentionPruneArchivesBeforeRewrite/failed_archive_leaves_log_byte-identical — log was rewritten despite the failed archive (pruneErr=retention: 아카이브 실패: 월별 아카이브 2026-08 기록 실패: 아카이브 임시 파일 생성: open .../archive/archive-452868282.jsonl.gz.tmp: permission denied): before 297 bytes, after 141 bytes`. Mutant reverted (`git diff` on retention.go empty); test green again (`ok ... 0.586s`). `retention.go` carries no change in the M1 commit.
- **EV-008 premise + behavioral RED (AC-CAP-012, clauses 1-2).** Test written against the pre-flip template (mode: report) → FAIL, verbatim: `no rotator row carries mode=apply (shipped template default did not reach the pass); rows=[{Unit:rotator Mode:report Outcome:summary Path:} {Unit:gc Mode:report Outcome:summary Path:}]`. After the M1 flip → PASS.
- **EV-001 premise + pre-edit RED (AC-CAP-003).** `TestTemplateGitignoreIgnoresLocalArtifacts` extended with the AC-CAP-003 cases, run before the template `.gitignore` edit → FAIL with exactly the four positive cases (`usage-log` / `lessons-inbox` / `*.corrupt` / monthly `.gz`: `ignored=false, want true`); all controls (tier-promotions, product.md, scaffolds) passed.
- **AC-CAP-010 RED.** `TestHomeCleanKeepFloors` written and run against the age-only backups selection → FAIL both fixtures, verbatim: `backup candidates = [removed-old-a removed-old-b removed-old-c removed-old-d removed-old-e], want [removed-old-a removed-old-b]` (fixture i) and `backup candidates = [removed-old-1 removed-old-2 removed-old-3], want [removed-old-2 removed-old-3]` (fixture ii); the release arm produced no failure (already correct, as the acceptance predicts).
- **EV-009/EV-010 premises + compile RED (AC-CAP-011).** Premises re-observed (table above). Final-form `TestProfileCapWarnsWithoutDeleting` run before the loader/advisory existed → build failure, verbatim: `undefined: hygiene.LoadProfileCapMaxBytes` / `undefined: hygiene.ProfileCapAdvisories` (×2 each). Post-green mutant (loader with compiled fallback) observed after M3 — see the M3 section below.

### M1 — decision-gated tests + template default flip

- M1 commit SHA: `7fe94f712` (`feat(SPEC-CAPACITY-RECLAIM-001): M1 decision-gated tests + hygiene default apply flip (card t1631)`). Known defect: its trailer carries `🗺 MoAI` (U+1F5FA) instead of `🗿 MoAI` (U+1F5FF) — an escape-sequence authoring mistake; B9 forbids amend, so the defect is recorded here for the lane (a pre-integration recommit is the lane's call). All later commits carry the correct literal trailer.
- Template `workflow.yaml` hygiene mode flipped `report` → `apply` (line 331) with the comment updated to state the new default; `make build` re-embedded the templates (exit 0).
- `go test ./internal/hook/ -run '^TestSessionStartAutoPathSplit$' -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 0.897s` (clauses 1-2 PASS: rotator apply rows, seeded verify target's `unit:gc / mode:apply` row, path-scan clause, fixture home byte-identical).
- `go test ./internal/harness/ -run '^TestRetentionPruneArchivesBeforeRewrite$' -count=1` → `ok` (0.833s).

### M0/B0 — baseline commit landed

- B0 SHA: `7d77361ac` (`docs(SPEC-CAPACITY-RECLAIM-001): M2 pre-change baseline (card t1631)`); `git show --name-only --format= 7d77361ac` → `.moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md` only.

### M2 — template ignore rules + untracking

- Template `.gitignore` gained the four REQ-CAP-001 rules (`.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, `.moai/**/*.corrupt*`, `.moai/harness/learning-history/archive/*.jsonl.gz`); `make build` re-embedded (exit 0).
- `go test ./internal/template/ -run '^TestTemplateGitignoreIgnoresLocalArtifacts$' -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 3.447s` (GREEN; pre-edit RED recorded above).
### M2 green runs (AC-CAP-004 / AC-CAP-001 / AC-CAP-005 / AC-CAP-013)

- B1 SHA: `0a6b31f10` (template .gitignore rules + extended test + five-file untracking); B1b SHA: `fb682d251` (six `.jsonl.gz` untracking per `d-20261014T033100Z-e411`).
- AC-CAP-013: `git merge-base 7d77361ac46161862d686adc10184821d10fd4b2 0a6b31f107eac483c44885c96d08f1dc4ff6b056` → `7d77361ac46161862d686adc10184821d10fd4b2` (= B0, so B0 is an ancestor of B1); `git show --name-only --format= 7d77361ac46161862d686adc10184821d10fd4b2` → `.moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md` only. (Known-bad-input controls EV-012/013/014 recorded in acceptance.md.)
- AC-CAP-004: `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive` → empty (runner: `(Bash completed with no output)`, exit 0 convention). All eleven files untracked.
- AC-CAP-001: `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1` → `300746 total` (< 5,000,000; equals the §B arithmetic prediction exactly).
- AC-CAP-005: post-B1b `wc -c` → `90331575` / `6681068` (both exist, non-zero, sizes unchanged); `shasum -a 256` over the nine archive paths → all nine digests byte-identical to the B0 baseline (list above).
- `git status --short` after B1b: clean except the two pending M3 cli test files (untracked by design until the M3 commit) — the untracked local state files are now ignored.

### M3 — profile cap warning-only + keep floors + shipped default

- `internal/hygiene/profile_cap.go` (new): `LoadProfileCapMaxBytes` (HOME-tier `state.profile_max_bytes`, no compiled fallback; file-absent or key-absent → ok=false; non-positive → config-invalid error) and `ProfileCapAdvisories` (sizes profile dirs, one advisory per over-cap profile naming the profile and the key, time-boxed via `context.WithTimeout`, deletes nothing, symlink-unfollowing).
- `internal/config/defaults.go`: `ProfileCapAdvisoryTimeout = 2 * time.Second` (the advisory's time-box, thresholds-in-defaults rule).
- `internal/cli/clean_home.go`: the compiled-cap candidate loop removed (category 0 comment rewritten: the cap is advisory-only; reclaim is `moai clean --home --force` on the age rules alone) — `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` → empty (AC-CAP-011 green clause; exit 1 by convention, runner showed no output). Category 4 backups floor: sort by mtime desc, keep the three newest, select the rest only when also older than the window (A12).
- `internal/hook/session_start_hygiene.go`: `runSessionHygiene` now ends with `runProfileCapAdvisory()` — the time-boxed warning-only advisory; every failure (absent key, unreadable config, deadline) logged and swallowed.
- `internal/template/templates/.moai/config/sections/state.yaml`: `state: {}` → documented `state:` block with `profile_max_bytes: 8589934592`; `make build` re-embedded (exit 0).
- Incident, caught by the template suite (recorded as a NEW failure introduced and repaired inside this milestone): the first versions of the two template comments named `SPEC-CAPACITY-RECLAIM-001`, and the full `go test ./internal/template/ -count=1` run failed `TestTemplateNoInternalContentLeak` with exactly `class=C1-spec-id-prefix | match=SPEC-CAPACITY-RECLAIM-001` on `templates/.gitignore` and `templates/.moai/config/sections/state.yaml` — templates are generic deployables and must not carry internal SPEC references. Both comments rewritten without the ID; `make build` re-embedded; `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$|^TestTemplateGitignoreIgnoresLocalArtifacts$' -count=1` → `ok` (3.855s). Post-fix full template suite: FAIL with exactly one test — `TestRuleDateProvenance` (`.claude/rules/moai/core/hooks-system.md:368`, the pre-existing baseline red recorded at pre-flight; untouched per scope discipline). The internal-isolation guard did its job.
- Full hook suite after the advisory wiring: `go test ./internal/hook/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 324.499s` (the only other test touching the pass stubs `hygieneRunFn` wholesale and is unaffected).
- `internal/config/testdata/shipped_key_inventory.yaml`: `state.profile_max_bytes` added as W/reader — the REQ-CKH-008 anti-rot guard (`TestShippedConfigKeysHaveReaders`) caught the new shipped key and the entry is its required triage row.
- Tests: `go test ./internal/cli/ -run '^TestHomeCleanKeepFloors$|^TestProfileCapWarnsWithoutDeleting$' -count=1` → `ok` (1.594s, post-revert); `go test ./internal/hook/ -run '^TestSessionStartAutoPathSplit$' -count=1` → `ok` (0.912s; clause 3 PASS — fixture-2 arm silent, fixture-1 arm warns naming `p-big` + `state.profile_max_bytes`, home2 byte-identical); `go test ./internal/hygiene/ -count=1` → `ok`; `go test ./internal/config/ -count=1` → `ok`.
- Mutant (AC-CAP-011, compiled-fallback loader): `LoadProfileCapMaxBytes` temporarily returning `(config.DefaultProfileMaxBytes, true, nil)` on key-absent → FAIL verbatim: `key-absent state.yaml resolved to (5368709120, true); the loader must carry no compiled fallback`. Reverted → `ok` again. Note: the mutant only bites the key-absent-file-present arm; the first fixture version (no state.yaml at all) returned early at file-absence and the mutant was unreachable — fixture 2 now writes a state.yaml without the key, and the file-absent case is asserted separately (also no fallback).
- Known Gaps for AC-CAP-011: the acceptance's second mutant direction ("a cap that selects a candidate fails fixture 1") is not observable with the AC's 6 MiB fixture — the compiled cap it would have to exceed is 5 GiB, so a reverted scan (cap-driven selection restored) would still select nothing on this fixture. The scan-side conversion is instead evidenced by: EV-010 (the `:243` comparison existed) + the post-change grep printing nothing + fixture-1 asserting zero projects candidates with a young entry.

### M3 commit + CLI-path fixture runs (AC-CAP-002 / AC-CAP-007; tool provenance per VCI §2.2)

- M3 commit SHA: `9a6fd4906`. Binary rebuilt from the clean committed tree (`make build`, BuildID `...-860-g9a6fd49065`), so the fixture verdicts below are attributed to build `9a6fd49065` judging tree `9a6fd4906` — same commit, no lag.
- Fixture: `/tmp/t1631-hygiene-fixture-37517` with `workflow.hygiene.audit_log_max_bytes: 1024`, one registered sink `rule-load-audit.jsonl` at 4096 bytes, and `.moai/lessons-inbox.jsonl` (39 bytes, SHA-256 `41ee8b8d76731005728537425a25a2939da914d0d479a34916744abc91ade358`).
- AC-CAP-002: `(cd fixture) <tree>/bin/moai clean --audit-logs --apply` → printer reported `rotated: .../rule-load-audit.jsonl` plus per-sink skips; `grep -n '"mode":"apply"' .moai/logs/hygiene-audit.jsonl` → 21 rows, row 1 verbatim: `{"ts":"2026-10-14T05:00:57Z","unit":"rotator","mode":"apply","outcome":"rotated","path":".../rule-load-audit.jsonl"}`. PASS.
- AC-CAP-007: post-run `wc -c` → `39` (unchanged); `shasum -a 256` → `41ee8b8d76731005728537425a25a2939da914d0d479a34916744abc91ade358` (unchanged); `grep -c 'lessons-inbox' .moai/logs/hygiene-audit.jsonl` → `0` (no audit row names it). PASS (held guard).
- AC-CAP-006: `grep -n -E 'usage-log' internal/hygiene/sinks.go` → single match `91: "usage-log.jsonl": "internal/harness/retention.go (30-day prune to monthly gz)"` (owner exception; `sinkRegistry` at `sinks.go:25-50` names it nowhere). HELD.

### E3 coverage (changed packages, this run, this tree `9a6fd4906`)

- `go test -cover ./internal/harness/ -count=1` → `coverage: 86.4% of statements` (≥85 target).
- `go test -cover ./internal/hygiene/ -count=1` → `coverage: 79.9% of statements` (package-wide, below the 85 target — reported honestly; the new `profile_cap.go` is exercised mainly from the cross-package tests measured next).
- Cross-package coverage of the new file: `go test -coverprofile -coverpkg=./internal/hygiene -run '^TestProfileCapWarnsWithoutDeleting$' ./internal/cli/` → `LoadProfileCapMaxBytes 75.0%`, `ProfileCapAdvisories 60.9%`, `profileSizeWithDeadline 78.6%` (uncovered arms: the error-degradation paths — unreadable config dir, parse failure, mid-walk deadline, unreadable entries).
- `go test -cover ./internal/config/ -count=1` → `coverage: 84.9% of statements` (the change is one additive const + one inventory row; no new branch).
- `go test -cover -run '^TestHomeClean|^TestProfileCap' ./internal/cli/` → `coverage: 5.0% of statements` — scoped-run number only: the full cli suite is structurally red in card worktrees (dispatch Section C; CI is the verdict surface), so no full-package cli coverage figure is claimable locally.
- `go test -cover ./internal/hook/ -count=1` → `coverage: 87.7% of statements` (511.850s; the full hook suite passed again in the same run).

### M4 — final gates (this run, tree `9a6fd4906` + §E.2/M4 progress edits uncommitted)

- `go vet ./internal/cli/ ./internal/hygiene/ ./internal/hook/ ./internal/config/ ./internal/harness/ ./internal/template/` → clean (no output).
- `go build ./...` → exit 0 (`BUILD-MAIN-OK`); `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`BUILD-WINDOWS-OK`).
- `golangci-lint run --timeout=2m ./internal/cli/... ./internal/hygiene/... ./internal/hook/... ./internal/config/... ./internal/harness/... ./internal/template/...` → `0 issues.` (no NEW lint findings; the pre-existing baseline red `TestRuleDateProvenance` is a test failure, not a lint issue, and is recorded above).
- Subagent-boundary grep (E4; no new agent surfaces added by this SPEC): `grep -rn 'AskUserQuestion\|mcp__askuser'` over every touched file (`internal/hygiene/profile_cap.go`, `internal/hook/session_start_hygiene.go`, `internal/cli/clean_home.go`, and the four new/extended test files) → no match (exit 1 by convention, runner showed no output).

### Post-close card-review repair round (leader-dispatched; codex advisory scope=card, base `ec193e4777`)

Three findings, all three mechanically confirmed on the tree before fixing, all three fixed in one round:

- **P1 — M3 keep-3 floor broke two pre-existing clean-home fixtures (both FAIL reproduced first, verbatim: `aged entry should be deleted under --force: .../backups/removed-old (stat err=<nil>)` and `control assertion failed: removed-plain should be deleted (stat err=<nil>)`).** The fixtures seeded only two `removed-*` dirs each, so keep-3 protected the asserted-deletion targets. Fix preserves each test's intent: `clean_home_test.go` seeds two aged fillers (`removed-keep-1/2`, ranked newer than `removed-old`, still beyond the 30-day window) so `removed-old` ranks 4th — beyond the floor AND expired — and the fillers join the survive list as the floor's positive demonstration; `clean_home_carveout_test.go` seeds three aged fillers with distinct mtimes (`removed-mixed` ranks 4th, so the whole-dir carve-out skip — not the rank floor — is its deciding protector; `removed-plain` ranks 5th and is deleted, keeping the --force-ran control).
- **P2#1 — Windows path prefix in the split test:** `string(filepath.Separator)` appended after `filepath.ToSlash` yields `.../logs\` on Windows, never matching audited row paths. Both occurrences replaced with the literal `"/"`.
- **P2#2 — the loader read only the HOME tier while the shipped default lands in the PROJECT tier:** `LoadProfileCapMaxBytes` now reads HOME first (semantics unchanged), then falls back to the project sections `<projectDir>/.moai/config/sections/state.yaml` (canonical constants `defs.MoAIDir` + `defs.SectionsSubdir` — no hand-rolled path) when the home tier is absent-or-key-absent; both absent still yields ok=false / no warning (D11 preserved). The hook pass threads its `projectDir` through; the test gains a third arm (home key-absent + project cap set → cap 1048576, advisory fires, no candidates) and the absent-key arm now passes an empty project dir to pin the chain's termination.

Re-verification after the round (this tree): `go test ./internal/cli/ -run 'TestCleanHome|TestHomeClean|TestProfileCap' -count=1` → `ok` (13 `--- PASS`, zero `--- FAIL`); `go test ./internal/hygiene/ -count=1` → `ok`; `go test ./internal/hook/ -run 'TestSessionStart' -count=1` → `ok` (a first parallel-shell FAIL did not reproduce, judged machine contention per the repeated-run discipline). Full hook suite: two re-run attempts both died to machine load (load average 76 from foreign sessions' concurrent suites — first run exceeded the 600s foreground bound, second killed exit 144 with an empty log); the repair's hook delta is two lines (a test literal + the advisory's projectDir threading), both covered by the green SessionStart family run, and the full hook suite passed on the pre-repair tree at M4 (`ok`, coverage 87.7%) — the full-suite verdict rides CI per the card-worktree doctrine. gofmt clean; `go vet` clean; `golangci-lint run` on the touched packages → `0 issues.` Fix commit: the single `fix(SPEC-CAPACITY-RECLAIM-001)` card-review round commit carrying this note (`b6ebf54b0`).

### Post-close sync-audit repair round (sync audit FAIL 0.78, `.moai/reports/t1631/{sync-audit.md,verdict.md}`, codex receipt `rcpt-be44957bd889025fc6fd9c6f`; 13 ACs held; two verified P2 defects in the profile-cap advisory)

- **P2-1 (starvation, fixed):** the advisory walk ran as step 1 of `computeDeferredAdvisory` (via `hygieneRunFn` → `runSessionHygiene`) inside the goroutine whose results join at `deferredScanJoinBound` (250ms, `session_start.go:1780`) — at the shipped 8 GiB default the targeted profiles are exactly the walks exceeding 250ms, so the warning AND every sibling deferred step were dropped each session. Fix: `runSessionHygiene` now installs the advisory through `profileCapSpawnFn` (default: its OWN goroutine via `go profileCapAdvisoryFn(projectDir)`), so the joined window is never consumed by the walk; the walk keeps its own time-box (`config.ProfileCapAdvisoryTimeout`, distinct from the 250ms join) and degrades to a skip notice on expiry. Warning-only semantics and the project-tier fallback are unchanged. `TestSessionStartDeferredSiblingsProceedDuringCapWalk` proves it: the advisory body stubbed to block until released; while blocked, `computeDeferredAdvisory` still returns AND the sibling `pruneTelemetry` ran (the aged `usage-YYYY-MM-DD.jsonl` fixture file is pruned). The split test installs a synchronous spawn for its clause-3 determinism (documented in its comment).
- **P2-2 (silent suppression, fixed):** `profileSizeWithDeadline` returned sub-path read errors, and the caller's non-deadline `continue` dropped the advisory with clean nil (auditor-reproduced: chmod-000 subdir → advisories=0 err=nil), contradicting the "Unreadable entries are skipped, never fatal" contract. Fix: WalkDir error arms (unreadable directories) and `Info()` failures are SKIPPED and COUNTED; only deadline/context errors abort. An over-cap profile now warns on its partial sum, and the advisory names the skipped count (`; measurement partial: N entries unreadable (skipped)`). Test: the unreadable-subdirectory arm asserts the advisory fires naming `prof-a`, the key, and `measurement partial`.
- First-cut defect caught by the new test and fixed in-round: the WalkDir error arm skipped without incrementing the count (advisory fired, `measurement partial` absent) — the counter now covers both arms.
- Re-verification (this tree): `go test ./internal/hook/ -run 'TestSessionStart' -count=1` → `ok` (26.7s; both cap tests PASS); `go test ./internal/cli/ -run 'TestCleanHome|TestHomeClean|TestProfileCap' -count=1` → `ok` (all four `TestProfileCapWarnsWithoutDeleting` arms PASS); `go test ./internal/hygiene/ -count=1` → `ok`; gofmt clean; `go vet` clean; `golangci-lint run` on internal/cli, internal/hook, internal/hygiene → `0 issues.` Fix commit: the single `fix(SPEC-CAPACITY-RECLAIM-001)` sync-audit round commit carrying this note.

## §E.3 Run-phase Audit-Ready Signal

run_status: complete
run_complete_at: 2026-10-14T05:05:00Z
run_tree: 9a6fd49065 (M3/M4 verification baseline; final commits listed below)

AC matrix (AC-CAP-001..013; full command + verbatim output records above and in acceptance.md §D.0):

| AC | Class | Verdict | Evidence anchor |
|---|---|---|---|
| AC-CAP-001 | MS | PASS | `300746 total` (< 5,000,000) — §E.2 M2 green runs |
| AC-CAP-002 | MS | PASS | 21 `"mode":"apply"` rows, outcome `rotated` on the seeded sink — §E.2 CLI-path fixture runs |
| AC-CAP-003 | RB | PASS | extended test RED pre-edit (4 positive cases) → `ok` post-edit — §E.2 RED + M2 |
| AC-CAP-004 | RB | PASS | `git ls-files` of the eleven paths → empty — §E.2 M2 green runs |
| AC-CAP-005 | RG | HELD | eleven local copies exist; nine archive digests byte-identical to B0 — §E.2 M2 green runs |
| AC-CAP-006 | RG | HELD | `sinks.go:91` owner exception only; no registry entry — §E.2 CLI-path fixture runs |
| AC-CAP-007 | RG | HELD | inbox 39 B + SHA unchanged; 0 audit rows name it — §E.2 CLI-path fixture runs |
| AC-CAP-008 | RB | PASS | template `workflow.yaml:331` now `mode: apply` (EV-007 was the RED) — §E.2 M1 |
| AC-CAP-009 | RB | PASS | test green vs unchanged prune; mutant (rewrite-before-archive) FAIL recorded, reverted — §E.2 RED |
| AC-CAP-010 | MS | PASS | `TestHomeCleanKeepFloors` RED pre-change (both fixtures) → `ok` — §E.2 RED + M3 |
| AC-CAP-011 | RB | PASS | `DefaultProfileMaxBytes` gone from clean_home.go (grep empty); `TestProfileCapWarnsWithoutDeleting` green; compiled-fallback mutant FAIL recorded — §E.2 M3 (one mutant direction unobservable, recorded as Gap) |
| AC-CAP-012 | RB | PASS | behavioral RED pre-flip (report-mode summary rows) → clauses 1-3 green (gc apply row for the seeded verify target; path-scan; home byte-identical; advisory arms) — §E.2 RED + M1 + M3 |
| AC-CAP-013 | MS | PASS | `git merge-base B0 B1` = B0 full SHA; `git show --name-only B0` = progress.md only — §E.2 M2 green runs |

Residual risks / Gaps for the sync phase:

1. AC-CAP-011's second mutant direction is not observable with the AC's 6 MiB fixture (details in §E.2 M3); scan-side conversion rests on EV-010 + post-change grep-empty + fixture-1 zero-candidates.
2. The M1 commit `7fe94f712` carries a `🗺 MoAI` trailer (U+1F5FA) instead of `🗿 MoAI` (U+1F5FF) — escape-sequence authoring mistake, amend forbidden by B9; the lane may recommit the branch tip before integration if it wants the trailer byte-exact. All later commits carry the literal correct trailer.
3. Pre-existing baseline red `TestRuleDateProvenance` (template package, `.claude/rules/moai/core/hooks-system.md:368` date leak) is untouched — out of this SPEC's scope; CI surface owner should triage it separately.
4. Full cli package suite was not run locally (structurally red in card worktrees); `internal/cli` verdict rides CI, as the dispatch's Section C anticipated.
5. Apply-by-default (AC-CAP-008) changes every new project's SessionStart auto path from report to apply for the non-destructive categories — this is the operator's approved Q1 verdict (8a62); existing projects keep their local workflow.yaml (mode: report) until they adopt the new template.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_


## §E.4 Sync-phase Audit-Ready Signal

Recorded by manager-docs in the single sync commit (3-phase close, plan→run→sync). The commit cannot cite its own hash, so `sync_commit_sha` holds the placeholder and the real SHA is backfilled in a following commit. This section records what was observed at commit time; the sync-audit verdict belongs to the auditor.

sync_complete_at: 2026-10-14T05:26:04Z
sync_commit_sha: "2a805d76b57dbfea8d3b3375336dd29f97169957"
sync_status: complete
audit_ready: true
b12_self_test_a: duplicate-entry guard before emission — `grep -c 'SPEC-CAPACITY-RECLAIM-001' CHANGELOG.md` printed `0` (grep exit 1, no-match), so the emission proceeded with no existing entry to duplicate
b12_self_test_b: AC counter over acceptance.md (the SSOT) — §D.0 lists exactly AC-CAP-001..013, all live, none carrying a `[RETIRED]`/`[REF]` reserved token; `live=13 excluded=0 ambiguous=0`; the CHANGELOG entry states the same count (13 acceptance criteria AC-CAP-001..013)
b12_self_test_c: every file path the CHANGELOG entries cite was checked with `ls -l` in this pass — internal/hygiene/profile_cap.go, internal/hook/session_start_hygiene.go, internal/cli/clean_home.go, internal/template/templates/.moai/config/sections/state.yaml, internal/template/templates/.gitignore, internal/config/defaults.go, and this progress.md all exist; entries were drafted from the actual diff (`git diff --name-only e993fff989..804180777c`), not from plan.md descriptions
changelog_entry_position: CHANGELOG.md `## [Unreleased]` — one entry at the top of `### Added` (profile-cap advisory) and three entries at the top of `### Changed` (template hygiene default apply, per-machine state untracking, backup keep-3 floor), each ending with the `🗿 MoAI` trailer
mx_tag_sync: 3 `@MX:NOTE: [AUTO]` tags added, each with an `@MX:SPEC: SPEC-CAPACITY-RECLAIM-001` sub-line — `LoadProfileCapMaxBytes` and `ProfileCapAdvisories` in `internal/hygiene/profile_cap.go` (no-compiled-fallback and warning-only invariants), and the backups keep-3 floor block in `internal/cli/clean_home.go` (A12 floor invariant). 0 tags updated, 0 removed. No `@MX:TODO` (both new exported functions are tested — cross-package coverage 75.0%/60.9% per §E.2 E3); no `@MX:WARN` (no goroutines, no complexity trigger); no `@MX:ANCHOR` (fan_in < 3). gofmt clean, `go vet` clean after the tag edits
frontmatter_status_transitions.spec_md: `status: in-progress -> completed`; `updated: 2026-10-14` already the sync date, unchanged
frontmatter_status_transitions.plan_md: no frontmatter block and no status field; nothing to refresh, not edited
frontmatter_status_transitions.acceptance_md: no frontmatter block and no status field; nothing to refresh, not edited
frontmatter_status_transitions.progress_md: header `status:` line still reads `status: draft` (plan-phase value; the run phase did not flip it) — outside this dispatch's allowed surface (§E.4 only), left untouched and reported to the lane
known_defect_note: a stale duplicate `## §E.3 Run-phase Audit-Ready Signal` heading carrying `_<pending run-phase>_` sits below the populated §E.3 (run-phase surface, outside the sync-phase allowed scope, left untouched and reported); era classification is unaffected — the parser matches the populated §E.3 heading and the §E.4 `sync_commit_sha` field

### AC matrix (acceptance.md is the SSOT; 13 criteria; verdicts transcribed from §E.3, not re-run at sync)

| AC | Class | Verdict | Evidence anchor (§E.2) |
|---|---|---|---|
| AC-CAP-001 | MS | PASS | `300746 total` (< 5,000,000) — M2 green runs |
| AC-CAP-002 | MS | PASS | 21 `"mode":"apply"` rows, `rotated` on the seeded sink — CLI-path fixture runs |
| AC-CAP-003 | RB | PASS | extended test RED pre-edit (4 positive cases) → `ok` post-edit — RED + M2 |
| AC-CAP-004 | RB | PASS | `git ls-files` of the eleven paths → empty — M2 green runs |
| AC-CAP-005 | RG | HELD | eleven local copies exist; nine archive digests byte-identical to B0 — M2 green runs |
| AC-CAP-006 | RG | HELD | `sinks.go:91` owner exception only; no registry entry — CLI-path fixture runs |
| AC-CAP-007 | RG | HELD | inbox 39 B + SHA unchanged; 0 audit rows name it — CLI-path fixture runs |
| AC-CAP-008 | RB | PASS | template `workflow.yaml:331` now `mode: apply` — M1 |
| AC-CAP-009 | RB | PASS | test green vs unchanged prune; mutant FAIL recorded, reverted — RED |
| AC-CAP-010 | MS | PASS | `TestHomeCleanKeepFloors` RED pre-change (both fixtures) → `ok` — RED + M3 |
| AC-CAP-011 | RB | PASS | `DefaultProfileMaxBytes` gone from clean_home.go; fixture test green; compiled-fallback mutant FAIL recorded (one mutant direction unobservable, recorded as Gap) — M3 |
| AC-CAP-012 | RB | PASS | behavioral RED pre-flip → clauses 1-3 green — RED + M1 + M3 |
| AC-CAP-013 | MS | PASS | `git merge-base B0 B1` = B0 full SHA; B0 touches only progress.md — M2 green runs |

## §F Phase 4 Mode Selection

Inputs: tier=M, scope ≈10 files (template .gitignore, workflow.yaml template, state.yaml template, gitignore_local_artifacts_test.go, retention_test.go, session_start_hygiene tests, clean_home.go, session_start_hygiene.go, state loader), domains=2 (Go code + template assets), language mix=Go+YAML, concurrency benefit=LOW (coding-heavy, interdependent milestone order), Agent Teams prereqs=not requested.
Decision: serial — one manager-develop carries M1→M0→M2→M3→M4 in order.
Justification: coding-heavy work takes the sequential path per the Anthropic coding-task caveat; milestones share files and ordering constraints (B0 before B1), so fan-out adds coordination without parallelism. Sweep preconditions (≥30 mechanical files) are not met.
Mode Selection recorded before the first run-phase Agent() spawn (grep token: serial).
