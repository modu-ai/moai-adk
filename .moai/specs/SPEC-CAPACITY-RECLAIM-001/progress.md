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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Inputs: tier=M, scope ≈10 files (template .gitignore, workflow.yaml template, state.yaml template, gitignore_local_artifacts_test.go, retention_test.go, session_start_hygiene tests, clean_home.go, session_start_hygiene.go, state loader), domains=2 (Go code + template assets), language mix=Go+YAML, concurrency benefit=LOW (coding-heavy, interdependent milestone order), Agent Teams prereqs=not requested.
Decision: serial — one manager-develop carries M1→M0→M2→M3→M4 in order.
Justification: coding-heavy work takes the sequential path per the Anthropic coding-task caveat; milestones share files and ordering constraints (B0 before B1), so fan-out adds coordination without parallelism. Sweep preconditions (≥30 mechanical files) are not met.
Mode Selection recorded before the first run-phase Agent() spawn (grep token: serial).
