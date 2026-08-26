# acceptance.md — SPEC-STAMP-REACHABILITY-001

Verification layer. Every AC is Given-When-Then, binary-testable, with a concrete verification command. All shell commands run from the worktree root unless stated. Version 0.2.0 (iteration-2) · 2026-08-27.

Baseline-attribution (v0.1.0 plan-phase observations, measured in THIS worktree at base `da791eb0a`):

- `git merge-base --is-ancestor 0d15864ae90b origin/main` → rc≠0 (orphan negative fixture live)
- `.moai/project/codemaps/provenance.json` `.commit_sha` = `410da655f39d…`; ancestor-of-origin/main rc=0 (green baseline live)
- No `--commit` flag exists on `moai graph stamp codemaps` (graph_stamp.go flags list: `--root` only); no reachability step exists in graph-freshness.yml (steps are checkout/setup-go/build/bootstrap/check only)

Iteration-2 additions to baseline-attribution (2026-08-27, HEAD advanced through 379b310a6/#1666; the reference script and fixtures were materialized under `/tmp/t291-guard*` scratch and every logic cell below was EXECUTED — outcomes quoted per-cell):

- D5 premise MEASURED before baking assertions: throwaway repo fed by a local source-path fetch of ONLY main (`git init /tmp/t291-guardobs && git -C /tmp/t291-guardobs fetch -q <this-worktree-abs-path> main`) → `git cat-file -e '0d15864ae90b^{commit}'` prints `fatal: Not a valid object name …` **rc=128** there, while the same command in this full-history worktree returns **rc=0**. The absence premise of AC-SP-003 is therefore observed fact, not assumption.
- Six logic cells observed via acceptance.md §A.1's canonical script (Branch labels reused below): A rc=1 with `::error::codemaps provenance stamp 0d15864ae90b is NOT an ancestor of PR base origin/main …`; B rc=0 `guard: stamp 410da655f… reachable from origin/main`; C rc=1 `::error::codemaps provenance names commit 0d15864ae90b which does not exist in this checkout's history`; D rc=0 skip-with-reason line; E rc=0 `guard: push event (no base ref) — object presence verified only`; F (mutant) rc=1 `fatal: Not a valid object name origin/`.
- Pre-M2 anatomy state observed: `grep -n "continue-on-error\|GITHUB_BASE_REF\|if:" .github/workflows/graph-freshness.yml` → zero matches.

## §A. Reference Guard Script (canonical text)

Every guard-shaped AC below verifies against THIS text: run-phase copies it verbatim into the workflow step (modulo workflow-env lines); evidence cells source it for execution. Where this section and the shipped step body disagree, the shipped step body must be brought to this contract or a blocker report returned.

```bash
#!/usr/bin/env bash
# Reference reachability guard for the codemaps provenance stamp.
# Canonical text: SPEC-STAMP-REACHABILITY-001 acceptance.md §A.1.
# Inputs: PROV path to provenance.json · REPO git root ·
#         GITHUB_BASE_REF target branch name (empty on push events).
set -euo pipefail
SHA=$(jq -r '.commit_sha // empty' "$PROV")
if [ -z "$SHA" ]; then
  echo "guard: provenance carries no commit anchor (dirty/content-fingerprint anchor) — reachability not judgeable, skipping"
  exit 0
fi
if ! git -C "$REPO" cat-file -e "${SHA}^{commit}" 2>/dev/null; then
  echo "::error::codemaps provenance names commit ${SHA} which does not exist in this checkout's history"
  exit 1
fi
if [ -z "${GITHUB_BASE_REF:-}" ]; then
  echo "guard: push event (no base ref) — object presence verified only"
  exit 0
fi
if ! git -C "$REPO" merge-base --is-ancestor "$SHA" "origin/${GITHUB_BASE_REF}" 2>/dev/null; then
  echo "::error::codemaps provenance stamp ${SHA} is NOT an ancestor of PR base origin/${GITHUB_BASE_REF} — orphan-bound stamp (a squash merge would strand it)"
  exit 1
fi
echo "guard: stamp ${SHA} reachable from origin/${GITHUB_BASE_REF}"
```

Conditioning mechanism pinned (plan.md §M2 decision): ONE unconditional step on both trigger events; the shell-internal emptiness test is the ONLY PR/push divergence point; GitHub-level step-level `if:` conditioning of reachability is prohibited (it would skip object-presence enforcement on push-to-main runs).

## §D. AC Matrix

### AC-SP-001 — Guard fails on non-ancestor stamp (RED-now)

**Given** a checkout whose history contains object `0d15864ae90b` and provenance naming that sha
**When** the §A.1 canonical script runs with `GITHUB_BASE_REF=main`
**Then** it exits non-zero with an error annotation naming both `0d15864ae90b` (or its resolved full sha) and the base ref.

Verify (canonical-script citation per §A.1):

```bash
printf '{"schema_version":1,"commit_sha":"0d15864ae90b","dirty":false}\n' > /tmp/pv-orphan.json
REPO="$PWD" PROV=/tmp/pv-orphan.json GITHUB_BASE_REF=main bash <reference-script> ; echo "rc=$?"
# expect: ::error::...NOT an ancestor of PR base origin/main... and rc=1
```

Observed (iteration-2): `rc=1`, message as quoted in baseline-attribution.
Mutation/vacuity guard: count ≥1 ancestry-branch hit across evidence; a selector matching zero executions is green-vacuous.

### AC-SP-002 — Guard passes on correct ancestor-named stamps (no false positive)

**Given** the tracked `.moai/project/codemaps/provenance.json` (`commit_sha` `410da655f…`)
**When** the §A.1 canonical script runs against origin/main
**Then** it exits 0 printing the reachability pass line.

Verify:

```bash
printf "{\"schema_version\":1,\"commit_sha\":\"$(jq -r '.commit_sha' .moai/project/codemaps/provenance.json)\",\"dirty\":false}\n" > /tmp/pv-current.json
REPO="$PWD" PROV=/tmp/pv-current.json GITHUB_BASE_REF=main bash <reference-script> ; echo "rc=$?"
```

Observed (iteration-2): `rc=0`, `guard: stamp 410da655f… reachable from origin/main`. Additional sweep: both `c9eed8ac6` and the delivering PR's eventual final stamp evaluate green.

### AC-SP-003 — Guard fails on object-absence (environment-dependent shape closed; premise measured)

**Given** a history-free context lacking the named object, built by fetching ONLY main from a LOCAL source path (no bare-remote fetch):

```bash
mkdir -p /tmp/stamp-guard-t && git init -q /tmp/stamp-guard-t \
  && git -C /tmp/stamp-guard-t fetch -q "$PWD" main && cd /tmp/stamp-guard-t
# FIRST assert the orphan truly absent there — measure, never assume:
git cat-file -e '0d15864ae90b^{commit}' ; echo "absence-premise rc=$?"   # expect rc=128
```

**When** the §A.1 canonical script runs there with the orphan-naming provenance

```bash
cp /tmp/pv-orphan.json . 2>/dev/null || true
REPO=/tmp/stamp-guard-t PROV=/tmp/stamp-guard-t/pv-orphan.json GITHUB_BASE_REF=main \
  bash <reference-script> ; echo "rc=$?"
```

**Then** it exits non-zero with the DISTINCT missing-object message (`does not exist in this checkout's history`) — different message class from AC-SP-001's ancestry failure — demonstrating the CI-side red shape that today manifests as generic exit 2.

Observed (iteration-2): absence premise `rc=128` (`fatal: Not a valid object name`); guard cell rc=1 with the missing-object message. Vacuity guard: if the absence-premise probe unexpectedly returns rc=0, the cell FAILS its own setup — stop, re-derive the fixture; do not proceed on an unverified history claim.

### AC-SP-004 — Anchorless provenance: skip-with-reason

**Given** a provenance fixture carrying `dirty:true`, a 64-hex `content_fingerprint`, and NO `commit_sha`
**When** the §A.1 canonical script runs
**Then** it exits 0 AND stdout carries the printed skip-with-reason line containing the greppable token `no commit anchor`. Absence of BOTH a failure and that reason line is a FAIL.

Observed (iteration-2): rc=0 with `guard: provenance carries no commit anchor … skipping`.

### AC-SP-005 — Explicit-commit stamps the named sha verbatim

**Given** a clean described-source tree and resolvable revision (merge-base form)
**When** `moai graph stamp codemaps --root "$TREE" --commit "$(git merge-base HEAD origin/main)"` executes against a `t.TempDir()` fixture repo
**Then** written provenance has `dirty:false` and `commit_sha` equal-bytes to the resolved full sha (`jq -r '.commit_sha' pv == git rev-parse <rev>`), schema_version 1, described_roots `[internal cmd pkg]`.
RED-now at v0.1.0: flag did not exist (flags list cited in baseline-attribution). Run-phase flips this cell.

### AC-SP-006 — Unresolvable revision rejected

**Given** a revision string resolving to nothing (`deadbeef42`)
**When** the stamp command runs with `--commit deadbeef42`
**Then** exit code ≠ 0, stderr names the resolution failure WITHOUT any absolute local path, and `.moai/project/codemaps/provenance.json` is byte-unchanged (cmp before/after).

### AC-SP-007 — Mixed-anchor combination rejected pre-write

**Given** a tree with ≥1 uncommitted change under internal/, cmd/, or pkg/
**When** stamp runs with a valid `--commit`
**Then** exit ≠ 0, path-free error, NO provenance file written (tmp neither renamed nor left behind), tree unchanged.
Unit-level: also asserts the underlying mx entry rejects without invoking the filesystem install path.

### AC-SP-008 — Default path characterized unchanged

**Given** the flagless invocation over (a) clean tree and (b) tree with described-root changes, snapshotted against pre-M1 behavior
**When** comparing outputs byte-for-byte (stdout shape, WriteFile+rename ordering, printed Describe line)
**Then** identical in all respects. Method: capture `pv.Describe()` strings and generated JSON key-set for both cases pre-edit and post-edit; equality required.

### AC-SP-009 — Freshness-contract regression lock

**Given** the existing suites `go test ./internal/cli/... ./internal/mx/... ./internal/graph/...`
**When** run after M1-M3 edits
**Then** zero failures; specifically-surfaced contracts (named selectors, all observed passing): codemaps changed-files endpoint diff (reverted churn counts zero), dirty-fingerprint anchoring, described_roots fidelity, threshold resolution from gate.yaml incl. malformed-gate.yaml exit 2, not-comparable system-error path, and the at-or-above staleness boundary itself (`count >= th.CodemapsChangedFiles` → stale; one-below → fresh). If a contract's test cannot be located by selector in run phase, the gap is ADDED as a new pinning test before proceeding (record which).
Anti-regression: package tests green ALSO at baseline (captured in plan-phase pre-flight) — no inherited-red masking.

### AC-SP-010 — Operator guidance present, 4 locales, greppable tokens

**Given** the docs-site cli-reference graph pages (en/ja/ko/zh)
**When** checking all four locales after M3
**Then** EACH file matches ALL THREE greppable token sets — the recipe (`merge-base HEAD origin/main`), the mode invocation (`graph stamp codemaps --commit`), and the prohibition (one of `branch-local` / `브랜치 로컬` / `ブランチローカル` / `分支本地` per locale idiom) — and prose states plainly WHY branch-local restamps re-orphan under squash merges; `hugo -s docs-site --minify --gc` completes warning-free.

```bash
for l in en ja ko zh; do f=docs-site/content/$l/cli-reference/graph.md
  grep -q 'merge-base HEAD origin/main' "$f" || echo "MISSING recipe $l"
  grep -q 'graph stamp codemaps --commit' "$f" || echo "MISSING cmd $l"
  grep -qE 'branch-local|브랜치 로컬|ブランチローカル|分支本地' "$f" || echo "MISSING ban $l"
done   # expect: no MISSING lines
```

Locale bodies need NOT be transliteration-equal — code spans stay untranslated; prohibition prose is native idiom per locale (oss-docs i18n rules).

### AC-SP-011 — Push-event anatomy: object-presence-only, ancestry gated on non-empty base ref (RED-now)

Three jointly-blocking parts; all three hold or the AC fails.

**(a) GREEN cell — push event does not evaluate ancestry.**
Given ancestor-named provenance AND an EMPTY `GITHUB_BASE_REF` (the push-event shape)
When the §A.1 canonical script runs
Then it exits 0 AND stdout names object-presence verification only AND carries NO ancestry output.

```bash
REPO="$PWD" PROV=/tmp/pv-current.json GITHUB_BASE_REF= bash <reference-script>
echo "push-green rc=$?"        # expect rc=0, line: guard: push event (no base ref)...
```

**Mutant-catch cell (discriminates the exact regression D1 names).** A guard variant that evaluates ancestry unconditionally against the empty base ref MUST fail this AC:

```bash
REPO="$PWD" PROV=/tmp/pv-current.json GITHUB_BASE_REF= bash <mutant-script> ; echo "mutant rc=$?"
# expect rc≠0 (fatal: Not a valid object name origin/) — i.e., the mutant turns a legitimate push red
```

Both cells OBSERVED (iteration-2): canonical rc=0 presence-only; mutant rc=1 — so any implementation equal to the mutant is caught by asserting (a)'s expected rc=0 + absence of ancestry output.

**(b) Static anatomy assertions over the SHIPPED `.github/workflows/graph-freshness.yml`** (execute after M2 lands; pre-M2 state observed as zero-matches):

```bash
grep -c 'continue-on-error' .github/workflows/graph-freshness.yml      # expect 0 — silent-slip escape hatch forbidden
grep -cE '^\s*- name:.*[Rr]eachab' .github/workflows/graph-freshness.yml # expect exactly 1 (single unconditional step)
grep -c 'GITHUB_BASE_REF:' .github/workflows/graph-freshness.yml       # expect ≥1 env plumbing line mapping ${{ github.base_ref }}
grep -c '\[ -z "\${GITHUB_BASE_REF:-}" \]' .github/workflows/graph-freshness.yml # expect exactly 1 (THE conditioning point)
```

**(c) Step-level conditional conditioning absent**:

```bash
awk '/- name:.*[Rr]eachab/,/- name:/' .github/workflows/graph-freshness.yml | grep -c '^      if:'
# expect 0 — no GitHub-level step-level `if:` may condition the reachability step out of either event
```

RED-now status: (a) is GREEN today when executed against §A.1's contract text but the SHIPPED anatomy claims (b)/(c) are pending M2 — recorded RED-now because no such step exists yet (`grep rc=1` zero-match observed pre-M2).

### AC-SP-012 — Judgeability inputs static inspection (REQ-SR-004 instrument)

**Given** the §A.1 canonical reference script now and the shipped workflow step body after M2
**When** statically inspected token-by-token
**Then** the positive allowance set is EXACTLY `{set, jq -r '.commit_sha', [ -z ... ], git cat-file -e, git merge-base --is-ancestor, echo, exit}` plus variable plumbing — and ZERO wall-clock/mtime tokens occur anywhere:

```bash
grep -nE -- '-mtime|-atime|-ctime| find |newermt|newer-than|date \+%|time\.Now|os\.Stat' \
  /path/to/reference-script.sh ; echo "forbidden-token rc=$?"   # expect zero matches
```

Observed (iteration-2): executed against the §A.1 text materialized at `/tmp/t291-guardref.sh` — zero matches, rc=1 from grep. Run-phase repeats identically against the shipped YAML step body (any match = AC FAIL: freshness signals may not enter the reachability judgment, REQ-GF-002 lineage).
Positive-side check: each allowed primitive occurs at least once (`grep -c` ≥1 for `cat-file -e`, `merge-base --is-ancestor`, `jq -r`) so the static inspection cannot pass on an empty script.

## §D.1 Severity & traceability

| AC | Severity | REQ covered | Direction |
|---|---|---|---|
| AC-SP-001 | Must | REQ-SR-001, REQ-SR-008 | guard red on bad input |
| AC-SP-002 | Must | REQ-SR-001, REQ-SR-008 | guard green on good input (mirror image) |
| AC-SP-003 | Must | REQ-SR-002 | guard red on missing object |
| AC-SP-004 | Must | REQ-SR-003 | honest skip path |
| AC-SP-005 | Must | REQ-SR-005 | explicit mode records named sha |
| AC-SP-006 | Must | REQ-SR-005 | invalid input rejected |
| AC-SP-007 | Must | REQ-SR-006 | mixed-anchor rejected |
| AC-SP-008 | Must | REQ-SR-007 | default preservation |
| AC-SP-009 | Must | REQ-SR-009 | regression lock |
| AC-SP-010 | Should | REQ-SR-010 | docs parity |
| AC-SP-011 | Must | REQ-SR-001, REQ-SR-004, REQ-SR-008 | push-event validity + anatomy + judgeability inputs (instrument basis) |
| AC-SP-012 | Must | REQ-SR-004 | static-inspection instrument |

Traceability both directions: REQ-SR-001→{001,002,011} · REQ-SR-002→{003} · REQ-SR-003→{004} · REQ-SR-004→{011,012} (+ plan.md §E4 review note) · REQ-SR-005→{005,006} · REQ-SR-006→{007} · REQ-SR-007→{008} · REQ-SR-008→{001,002,011} · REQ-SR-009→{009} · REQ-SR-010→{010}; every AC maps back into that list.

## §E. Edge Cases

| Case | Expected behavior |
|---|---|
| provenance.json unreadable/tracked-file corrupted | guard treats as failure-equivalent to anchorless-with-anomaly: fail the step (cannot verify what cannot be read — silent pass is the forbidden direction) — pinned by an extra unit row under AC-SP-001's evidence |
| Short-rev form passed to `--commit` (7-char) | accepted, resolves, FULL sha recorded |
| Ref-name form (`main`, `origin/main`) | accepted like any rev-parse-valid expression |
| `--commit` equal to current HEAD | legal no-op relative to default path (byte-identical output) |
| Push-to-main event with orphaned stamp landed anyway | object-presence clause fails main's own run; graph check still independently red (both signals may fire — independent layers, no suppression) |
| Fork PR where origin/<base_ref> differs across forks | guard resolves base_ref against the checkout's fetched refs; GitHub-hosted fork PRs carry base-repo refs with fetch-depth 0 — if ever unverifiable, the step fails safe (existence clause still binding), NEVER skips silently |

## §F. Quality Gates & Definition of Done

- Targeted package tests: zero failures on `./internal/{cli,mx,graph}/...` (TRUST5 Tested floor for touched packages; coverage on NEW code paths ≥85%).
- `go vet ./internal/cli/... ./internal/mx/...` clean; `gofmt -l` clean on touched files.
- All twelve ACs recorded PASS in progress.md §E.2 with command + verbatim output; five-section report format on the run-phase completion report. AC-SP-011(b)/(c) executions must postdate the M2 workflow edit (anatomy asserted on the landed artifact, not on intent).
- The delivering PR's own graph-freshness run observed green on CI (live witness of the guard's pass path AND the green-baseline premise) before sync close.
- Frontmatter transitions owned canonically: draft → in-progress (manager-develop, first run commit), implemented/completed (manager-docs, sync commit).
