# t332 card sweep — batch 5

Cards: t313 t315 t319 t320 t323 t324 t325 t327 t329 t337  (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t313

**Premise (one sentence).** `EnterWorktree` uses `origin/HEAD` (which pointed at `main`) as the
implicit branch base, so every card worktree was built from `main` instead of the git-flow-mandated
`develop`, and the fix needs both a durable branch-base config surface and a `moai web` UI to set
it.

**Premise verdict.** `holds` — the underlying defect (base-branch mismatch, lead's emergency
`git remote set-head` fix) is the documented trigger for SPEC-WORKTREE-BASEREF-001, which landed
(see Landing verdict). The card's own "미검증" callout (whether Claude Code's `fresh` mode actually
reads `origin/HEAD`) is exactly what the landed SPEC's implementation (`git_strategy.worktree_base_branch`
config key + the hook that aligns `origin/HEAD`) was built to settle from first principles rather than
assumption — I did not re-verify that specific runtime behavior myself; see Gaps.

**Landing verdict.** `landed`
- commit: `62ff3c2e6` (merge(WT-worktree-baseref): integrate card t313 — configurable card-worktree
  base branch (SPEC-WORKTREE-BASEREF-001))
- pinned ref: `ee50984abe4f11ac337382b48a26328f091e200a`
- `--is-ancestor` exit: 0
- branch + tip (in-flight only): — (worktree `.claude/worktrees/t313` @ `3fd8b5072` still exists
  but its tip commit is itself one of the 11 commits absorbed by the merge — the worktree was simply
  never disposed post-landing)

**Claim.** t313 is fully landed on `develop` via SPEC-WORKTREE-BASEREF-001 (11 commits: plan → config
schema key → hook alignment → doctor diagnostic → `git worktree add` base-branch wiring → `moai web`
free-text field → docs → sync/backfill → merge → post-merge doctor restamp). The `contains t295`
relation the card records is consistent with the merged SPEC's scope (base-branch-aware worktree
creation covers the "checkout an existing branch" path t295 names).

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt313\b' --oneline
f5a834fef fix(cli): restamp doctor golden pass count after the t313 merge
62ff3c2e6 merge(WT-worktree-baseref): integrate card t313 — configurable card-worktree base branch (SPEC-WORKTREE-BASEREF-001)
3fd8b5072 chore(SPEC-WORKTREE-BASEREF-001): backfill sync_commit_sha
b0d179de1 docs(SPEC-WORKTREE-BASEREF-001): sync-phase artifacts and 3-phase close
92102de1e docs(SPEC-WORKTREE-BASEREF-001): run-phase evidence and verdict (t313)
7d46e69c9 docs(worktree): document the stored card-worktree base branch and its two consumers
80d9e7e5b feat(web): expose worktree_base_branch in the console as a free-text field
c59e74232 feat(cli): add the Worktree Base Branch doctor diagnostic
97aef573d feat(cli): pass the configured base branch to git worktree add
26cc9ba90 feat(hook): align refs/remotes/origin/HEAD from the configured worktree base branch
a9c61cf56 feat(config): add git_strategy.worktree_base_branch schema key and neutral default
e717133cb feat(SPEC-WORKTREE-BASEREF-001): plan-phase artifacts (Tier M, 3 artifacts)

$ git merge-base --is-ancestor 62ff3c2e6 ee50984abe4f11ac337382b48a26328f091e200a; echo $?
0
```

**Baseline-attribution.** Measured against `origin/develop` pinned SHA
`ee50984abe4f11ac337382b48a26328f091e200a`, in this run.

**Gaps.** I did not re-run `EnterWorktree` end-to-end to confirm the landed implementation actually
produces a `develop`-based worktree at runtime — the card's own "실제로 만들어 실측할 것" instruction
was addressed inside the SPEC's own run-phase evidence (`92102de1e`), which I did not open in full. I
also did not check whether the `moai web` free-text field (`80d9e7e5b`) is wired end-to-end to a
consumer, or whether it is only stored.

**Residual-risk.** The still-present, undisposed `.claude/worktrees/t313` worktree is stale
housekeeping, not a functional risk — its tip is already an ancestor of the merge. If the `moai web`
field from `80d9e7e5b` is store-only (no consumer), a related but separate defect could remain
open under a different card.

**Proposed disposition.** `already-landed` — rests on the `--is-ancestor` exit 0 evidence above.

**Overlap candidates.** t319 (same interview-schema/config-surface class — t319's card text
explicitly notes "카드 t313 이 같은 표면에 새 항목을 얹는다"). No other in-scope id references t313
in its own text within this batch.

---

### t315

**Premise (one sentence).** t303's SPEC-SYNC-STRATEGY-KEY-001 audit left two carry-forward defects
(D6: a v3.3.0-scoped fallback-sentinel removal, D7a: a v3.2.0 release-notes obligation for the
`main_direct` → `github-flow` default flip) that have no owning card, so this card exists to hold
them until the right release-prep moment.

**Premise verdict.** `holds` — both carry-forwards are still open in the landed t303 sync commit's
own text, which explicitly assigns them to "card t315" by name.

**Landing verdict.** `not-landed`
- commit: — (no delivering commit found; two commits *mention* t315 by name as a forward pointer,
  neither delivers it)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree named t315 in `00-worktree-list.txt`)

**Claim.** t315 has not been worked. Both mentions found are the *origin* of the carry-forward
(t351's AC wording fix records "카드 t315" as the future remover of the sentinel; t303's own
terminal-transition commit says "the two carry-forwards D6/D7a remain card t315's"), not evidence of
execution.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt315\b' --oneline
60a6b2b97 docs(spec): refine AC-SYK-012.1 so it stops contradicting AC-SYK-003 (t351)
ed68889e3 docs(SPEC-SYNC-STRATEGY-KEY-001): terminal transition implemented -> completed (t303)

$ git show -s --format=%B 60a6b2b97 | grep -n t315
Also records in D.5 that the raw-count clause must move from 1 to 0 when v3.3.0
removes the sentinel (card t315), and that AC-SYK-003 becomes obsolete then.

$ git show -s --format=%B ed68889e3 | grep -n t315
Open debt is unchanged: OBS-2 and OBS-3 stay open (OBS-3 is card t333's
trigger axis), and the two carry-forwards D6/D7a remain card t315's.

$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt315\b' --oneline
(no output)
```

**Baseline-attribution.** Measured against both pinned SHAs, in this run.

**Gaps.** I did not open t303's full verdict.md to re-confirm the D6/D7a text verbatim beyond what
the two commit messages already quote — the two mentions found are internally consistent, so I did
not treat this as a gap requiring a third source.

**Residual-risk.** None specific to landing status — this is a not-yet-started card, correctly
queued.

**Proposed disposition.** `keep` — two concrete carry-forward obligations, each with a clear
originating SPEC and defect id, not yet actioned.

**Overlap candidates.** none observed in-batch. Cross-batch: t303 (origin SPEC, not in-scope list —
already closed) and t351/t333 mentioned by name inside the card text but neither is in
`inscope-all.txt`.

---

### t319

**Premise (one sentence).** `tab_schema.json` (the interview schema for `moai-workflow-project`) has
no pointer anywhere telling a consumer to read it, so it may be a dead file whose internal
correctness nobody checks.

**Premise verdict.** `holds` — I re-ran the card's own claimed scan and got the same counts.

**Landing verdict.** `not-landed`
- commit: — (no delivering commit)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t319)

**Claim.** The file exists at two mirrored locations, has exactly one Go-code reference (a
neutrality test that does not consume it as a schema), and zero references from its owning skill's
`SKILL.md`.

**Evidence.**
```
$ find . -iname "tab_schema.json" -not -path "*/node_modules/*"
./.claude/skills/moai-workflow-project/schemas/tab_schema.json
./internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json

$ grep -rln "tab_schema" --include="*.go" internal/
internal/template/internal_content_leak_test.go

$ grep -rln "tab_schema" .claude/skills/moai-workflow-project/
(no output)

$ grep -n "schema" .claude/skills/moai-workflow-project/SKILL.md
106:See [configuration schema and language fields](references/configuration.md) for full field reference and supported language metadata.
195:4. Validate rendered output against schema or existing conventions

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt319\b' --oneline
(no output)
```

**Baseline-attribution.** Scoped scan: `internal/` for Go references, `.claude/skills/moai-workflow-project/`
for skill-body references, both against worktree HEAD `6165f9f5e`.

**Gaps.** I did not scan `.claude/agents/` or other skills outside `moai-workflow-project` for a
reference (the card's own scan claims 10 non-consumer references across SPEC docs/reports/manifest
hashes — I did not enumerate those 10 myself, only reproduced the two counts that matter for the
holds/falsified decision). I also did not determine whether an *agent* (not a skill/code file) is
told out-of-band (e.g. in its spawn prompt) to read this file.

**Residual-risk.** If some agent's runtime prompt (not visible to a static grep) does reference this
file, the "orphan" framing would be wrong even though the static evidence holds.

**Proposed disposition.** `needs-operator-decision` — the card itself frames this as needing a
decision (retire the file vs. add a pointer), consistent with the file's actual orphan status
observed here.

**Overlap candidates.** t313 (adds a new entry to a config surface the card describes as sharing the
same "표면"). t316 is named in the card text (the key-mismatch defect this orphan status may explain)
but is not in `inscope-all.txt`.

---

### t320

**Premise (one sentence).** `moai integration release` returns an `ERROR` with an empty message body
when the calling session does not hold the lock (e.g. after being evicted by another lane's
`--force acquire`), leaving the caller unable to read why it failed.

**Premise verdict.** `falsified` — the sentinel error for exactly this "not held" case carries a
non-empty message, and has since the feature's original commit, predating the 2026-08-27
observation.

```
$ grep -n "ErrIntegrationLockNotHeld\|func ReleaseIntegrationLock" internal/kanban/integration_lock.go
52:var ErrIntegrationLockNotHeld = errors.New("no release integration window is held")
215:func ReleaseIntegrationLock(projectRoot, sessionID string, force bool) (released *IntegrationLock, err error) {

$ git show b2ad9158c -- internal/kanban/integration_lock.go | grep -n "ErrIntegrationLockNotHeld\|no release"
113:+// ErrIntegrationLockNotHeld is returned by ReleaseIntegrationLock when no
117:+var ErrIntegrationLockNotHeld = errors.New("no release integration window is held")
258:+		return nil, ErrIntegrationLockNotHeld

$ git show 3f3465369 -- internal/kanban/integration_lock.go | grep -n "ErrIntegrationLockNotHeld\|no release"
(no output — the t298 M2/M3 fix did not touch this sentinel)

$ sed -n '223,244p' internal/cli/integration.go
(RunE returns `err` directly on the not-held path; cobra's default error
printer renders err.Error() — "no release integration window is held" — not
an empty string)
```
The card's own hypothesized cause (eviction → not-held → empty-message error) maps directly onto
this code path, and that path's message is not empty in this tree, nor was it ever empty since the
lock feature's inception commit `b2ad9158c`.

**Landing verdict.** `not-landed`
- commit: — (no delivering commit — expected, since this is a fresh defect report, not a claim of a
  prior fix)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t320)

**Claim.** As currently understood and sourced, the card's premise does not match the code in this
tree. Either (a) the lane's original observation came from a different/older binary (v3.1.2 per the
worker instructions' banned-column caveat, which is plausible since CLI binaries lag source), or (b)
the empty message came from a code path other than the hypothesized eviction/not-held one that I did
not examine (e.g. a `--force` release path, or an unhandled panic/recover swallowing output).

**Evidence.** See the falsifying commands above, plus:
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt320\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt320\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/kanban/integration_lock.go` and `internal/cli/integration.go` at
worktree HEAD `6165f9f5e`; the two cited historical commits (`b2ad9158c`, `3f3465369`) via `git show`.

**Gaps.** I did not check the `--force` release branch of `ReleaseIntegrationLock` for a separate
empty-message path, nor did I check whether the actually-*installed* `moai` binary (v3.1.2, per the
worker instructions) matches this source tree's `integration_lock.go` — the lane's observation was
against a running binary, not this source. I also did not check `internal/cli/fang.go`'s error
rendering wrapper for any message-stripping behavior under specific flag combinations.

**Residual-risk.** The card's *diagnosis* may be wrong while its *observation* (an empty ERROR
message was genuinely seen) stays true — meaning a real defect could exist on a code path this scan
did not reach. A `falsified` verdict here is about the stated cause, not a claim that no bug exists.

**Proposed disposition.** `needs-operator-decision` — the falsified cause suggests re-scoping rather
than dropping outright: either narrow the card to "confirm against the exact binary version used" or
re-open it as "audit every ReleaseIntegrationLock error path for a possible empty-message case."

**Overlap candidates.** none observed in-batch. Cross-batch: t298 (the SPEC this card explicitly
declines to fold into) is not in `inscope-all.txt` (already closed per memory).

---

### t323

**Premise (one sentence).** The catalog-hash integrity mechanism hashes only the single `SKILL.md`
(or `skill.md`) file inside a skill directory entry, so changes to any other file in that directory
(`schemas/`, `references/`, `scripts/`) are shipped via `//go:embed` but never move the catalog hash
that is supposed to attest integrity.

**Premise verdict.** `holds` — the hashing function's directory-branch logic matches the claim
exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t323)

**Claim.** `resolveHashSourcePath` in `internal/template/scripts/gen-catalog-hashes.go` resolves a
directory catalog entry to exactly one file (`SKILL.md` first, `skill.md` fallback) and hashes only
that file; no code path aggregates the rest of the directory tree into the hash.

**Evidence.**
```
$ sed -n '100,134p' internal/template/scripts/gen-catalog-hashes.go
... (directory branch: `for _, candidate := range []string{"SKILL.md", "skill.md"}` returns the
     first found path; non-directory branch hashes the file directly) ...

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt323\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/template/scripts/gen-catalog-hashes.go` at worktree HEAD
`6165f9f5e`.

**Gaps.** I did not check whether the catalog's `path` field for a directory-shaped entry can ever
point at a non-skill directory with a different resolution rule, nor did I check the consumer side
(what reads and trusts the catalog hash at runtime) to see how severe the blind spot actually is in
practice.

**Residual-risk.** If some other mechanism (not found by this scoped read) separately verifies
non-SKILL.md files, the severity framing in the card could be overstated even though the mechanical
claim is correct.

**Proposed disposition.** `keep` — mechanical claim reproduced exactly as stated; the card frames the
choice (a)/(b)/(c) as needing an operator decision, which I did not adjudicate.

**Overlap candidates.** t317 is named in the card text ("같은 병의 다른 증상") but is not in
`inscope-all.txt`. No other in-scope batch-5 card touches catalog hashing.

---

### t324

**Premise (one sentence).** `develop` currently has no branch protection at all, so the git-flow
model's CI status checks run only after lanes have already pushed directly to `develop` (detection,
not prevention), and re-enabling required-status protection needs a co-design with the no-card-PR
lane-push model rather than a simple toggle.

**Premise verdict.** `holds` — the cited 404 reproduces exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t324)

**Claim.** `develop` branch protection is confirmed absent right now, matching the card's cited
evidence verbatim.

**Evidence.**
```
$ gh api repos/modu-ai/moai-adk/branches/develop/protection
{"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection#get-branch-protection","status":"404"}
gh: Branch not protected (HTTP 404)

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt324\b' --oneline
(no output)
```

**Baseline-attribution.** Live GitHub API state for `modu-ai/moai-adk`, queried in this run (not a
tree-scoped measurement — branch protection is GitHub-side config, not a file).

**Gaps.** I did not check `main`'s current protection settings for comparison, and did not enumerate
which specific CI workflows currently exist as candidate required-status checks (the card names this
as an open design question, which I did not attempt to resolve).

**Residual-risk.** GitHub-side settings can change between this read and any future action on this
card — this is a live-state fact, not a tree-pinned one (§2.1 moving-ref concern: the 404 is current
as of this run, not a durable baseline).

**Proposed disposition.** `keep` — operator-flagged decision card (`[운영자 판정 2026-08-27]` prefix
in the card text itself), premise reproduces exactly, design questions remain genuinely open.

**Overlap candidates.** t325 (both concern the integrity of the `develop`→`main` promotion path
under the git-flow no-card-PR model; t324 is about protecting `develop` itself, t325 is about a
workflow that could bypass the `main`-entry gate this protection sits alongside).

---

### t325

**Premise (one sentence).** `spec-status-auto-sync.yml` fires on any `pull_request: closed` event
with no base-branch filter and pushes to `main` with `contents: write`, so a PR targeting `develop`
being closed could trigger a push to `main` — a hypothesis, not yet observed firing.

**Premise verdict.** `holds` (as an unverified-but-structurally-confirmed hypothesis, matching the
card's own framing — the card explicitly labels this "미검증 가설, 검증 필요" and I did not push the
verification further than the card already had).

**Landing verdict.** `not-landed`
- commit: — (one commit *mentions* t325 as a still-unverified hypothesis, does not deliver it)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t325)

**Claim.** The workflow trigger and push target match the card's description exactly: no `branches:`
filter under `on.pull_request`, and `git push origin main` present in the job body.

**Evidence.**
```
$ sed -n '1,14p' .github/workflows/spec-status-auto-sync.yml
name: SPEC Status Auto-Sync
on:
  pull_request:
    types: [closed]
permissions:
  contents: write    # git push origin main (line 107)
  issues: write      # gh issue create fallback (line 95-99)

$ grep -n "git push origin main\|branches:" .github/workflows/spec-status-auto-sync.yml
9:# (contents: read) 가 line 107 git push origin main 을 403 으로 실패시키던
12:  contents: write    # git push origin main (line 107)
16:# commit + git push origin main; cancelling mid-push would leave the
123:            git push origin main

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt325\b' --oneline
d723ca29e docs(t314): separate the three verdict gaps, mark the third as awaiting observation

$ git show -s --format=%B d723ca29e | grep -n t325
Now stated as three items with distinct character: unverified (release-PR
filter, no open PR to measure), unverified hypothesis (spec-status-auto-sync,
card t325), and awaiting observation (first firing of spec-lint /
```

**Baseline-attribution.** `.github/workflows/spec-status-auto-sync.yml` at worktree HEAD `6165f9f5e`;
the one mentioning commit via `git show` against pinned develop.

**Gaps.** I did not trace the job body past line 123 to confirm which ref is actually checked out
before the push (the card itself says this is the exact next step — "발화 시 어느 ref 로 push 하는지
코드 경로를 따라갈 것" — and I stopped at reproducing the structural facts already in the card, per
the restraint instruction).

**Residual-risk.** If `actions/checkout@v7` with `fetch-depth: 0` in this job checks out the
PR-closed event's base ref (not always `main`) before the commit+push steps, the actual runtime
behavior could differ from what the trigger/permission facts alone suggest — this needs the
step-by-step trace the card calls for, not yet done here or by the referenced commit.

**Proposed disposition.** `keep` — hypothesis remains open and structurally plausible; card correctly
declines to overstate to a confirmed finding.

**Overlap candidates.** t324 (both touch the integrity of the `develop`→`main` boundary under
git-flow). t314 is named in the originating commit but is not in `inscope-all.txt`.

---

### t327

**Premise (one sentence).** `treeDirty` (the function deciding whether a stamp anchors to a commit or
to a dirty-tree fingerprint) checks raw `git status --porcelain` dirtiness without applying the
"described-worthy" filter that the SPEC-GRAPH-FRESHNESS-CADENCE-001 (t322) audit found elsewhere, so
a dirty-but-not-described-worthy tree (e.g. only `testdata/` changed) still gets denied a `--commit`
anchor.

**Premise verdict.** `holds`, with one citation correction: the card's cited location
(`internal/config/provenance.go:201`) does not exist — `internal/config/provenance.go` is only 104
lines and defines no `treeDirty` symbol. The actual function is `internal/mx/provenance.go:225`, and
its behavior matches the mechanism the card describes exactly (raw `git status --porcelain` dirty
check, no described-worthy filtering).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t327)

**Claim.** The mechanism claim is correct; the file:line citation in the card is stale/wrong and
should be corrected to `internal/mx/provenance.go:223-227` before this card is worked.

**Evidence.**
```
$ wc -l internal/config/provenance.go
104 internal/config/provenance.go
$ grep -n "treeDirty" internal/config/provenance.go
(no output)

$ grep -rn "treeDirty" internal/mx/provenance.go
184:// that, and treeDirty's emptiness test depends on it (CR round-2 3855149357).
223:// treeDirty reports whether any file under the given repo-relative roots has
225:func treeDirty(root string, roots []string) bool {
249:	if treeDirty(projectRoot, describedRoots) {

$ sed -n '223,227p' internal/mx/provenance.go
// treeDirty reports whether any file under the given repo-relative roots has
// uncommitted changes (staged, unstaged, or untracked) versus HEAD.
func treeDirty(root string, roots []string) bool {
	args := append([]string{"status", "--porcelain", "--"}, roots...)
	return gitOut(root, args...) != ""
}

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt327\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/config/provenance.go` and `internal/mx/provenance.go` at worktree
HEAD `6165f9f5e`.

**Gaps.** I did not read SPEC-GRAPH-FRESHNESS-CADENCE-001's D.1/E sections that the card says already
document the deferral and remediation direction — I verified only the mechanism claim, not the SPEC
cross-reference's exact wording.

**Residual-risk.** If the card is worked using the stale `internal/config/provenance.go:201` citation
without first re-locating the function, the implementer could edit the wrong file or waste time
searching a 104-line file for a symbol that isn't there.

**Proposed disposition.** `keep`, with a note attached before dispatch: correct the location citation
to `internal/mx/provenance.go:223-227`.

**Overlap candidates.** t322 (the originating SPEC that deferred this scope, not in `inscope-all.txt`
— already closed/landed per project memory). No other in-scope batch-5 card touches this function.

---

### t329

**Premise (one sentence).** `.moai/docs/git-workflow-doctrine.md` §18.12 cites `internal/bodp/relatedness.go`
by function and file name for its BODP decision matrix, but that package does not exist in the tree —
the doctrine's actual SSOT for the default base-branch recommendation is a different, live file
(`branch-origin-protocol.md`), and this card's scope is narrowly to redirect the citation, not to
change the `origin/main` default itself.

**Premise verdict.** `holds` — reproduced exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t329)

**Claim.** `internal/bodp/` has zero tracked files on pinned `develop`; the doctrine's actual live
SSOT (`branch-origin-protocol.md:26`) carries the `origin/main` default-recommendation language the
card quotes.

**Evidence.**
```
$ git ls-tree -r --name-only ee50984abe4f11ac337382b48a26328f091e200a -- internal/bodp/ | wc -l
0

$ grep -n "relatedness.go\|internal/bodp" .moai/docs/git-workflow-doctrine.md
402:`internal/bodp/relatedness.go` `Check()` 함수가 다음 3개 시그널을 평가한다:
412:`internal/bodp/relatedness.go` `applyMatrix()` — SignalB 우선순위 dominates A/C:

$ grep -n "When no signal fires" .claude/rules/moai/development/branch-origin-protocol.md
26:- [ZONE:Evolvable] [HARD] The recommended base MUST be derived from the signals below, not assumed. When no signal fires, the recommendation is `origin/main` — team-safe, because it reflects the latest merged state rather than whatever the local checkout happens to hold.

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt329\b' --oneline
(no output)
```

**Baseline-attribution.** `git ls-tree` against pinned develop `ee50984abe4f11ac337382b48a26328f091e200a`;
`.moai/docs/git-workflow-doctrine.md` and `.claude/rules/moai/development/branch-origin-protocol.md`
at worktree HEAD `6165f9f5e`.

**Gaps.** I did not search the rest of the doctrine document tree for other instances of the "same
class" the card flags at the end ("이 문서군이 코드 경로를 인용하는 다른 지점들") — that is explicitly
left open by the card itself as follow-on scope.

**Residual-risk.** None specific — this is a narrowly-scoped, cleanly-reproduced documentation-drift
finding with an explicit scope boundary already stated in the card.

**Proposed disposition.** `keep` — reproduces exactly, scope is already well-bounded by the card
itself.

**Overlap candidates.** none observed in-batch (no other batch-5 card touches
`git-workflow-doctrine.md` or `internal/bodp`).

---

### t337

**Premise (one sentence).** On Windows, the anchor guard's `isProcessAlive` unconditionally returns
`true`, so a stamped-but-actually-dead session PID is never corrected back to reclaimable — the one
code path that still contradicts the declared TREAT-AS-LIVE invariant, and it predates and was not
touched by t298's fix.

**Premise verdict.** `holds` — reproduced exactly, including the "diff 0" claim against t298.

**Landing verdict.** `in-flight-unlanded`
- commit: — (no delivering commit on pinned develop)
- pinned ref: —
- `--is-ancestor` exit: 1 (tip `c72a517c3` is NOT an ancestor of pinned develop)
- branch + tip (in-flight only): `WT-windows-stamp-liveness` @ `c72a517c3` (from
  `.moai/reports/t332/00-worktree-list.txt`, worktree `.claude/worktrees/t337`)

**Claim.** A live worktree exists for this card with unmerged work; the underlying code claim
(unconditional `true` on Windows, not touched by t298) is independently verified as still accurate on
`develop`.

**Evidence.**
```
$ cat internal/session/anchor_pid_windows.go
//go:build windows
...
func isProcessAlive(pid int) bool {
	_ = pid
	return true
}
...
func probeProcessLiveness(pid int) (alive bool, determined bool) {
	_ = pid
	return false, false
}

$ grep -rn "isProcessAlive\|sessionPIDFromEnv" internal/session/*.go | grep -v _test.go
internal/session/anchor.go:80:			alive := e.PID > 0 && isProcessAlive(e.PID)
internal/session/session_pid.go:54:	pidIsAlive              = isProcessAlive
internal/session/session_pid.go:74:	if pid, ok := sessionPIDFromEnv(os.Getenv(config.EnvMoaiSessionPID)); ok {

$ git log ee50984abe4f11ac337382b48a26328f091e200a --oneline -- internal/session/anchor_pid_windows.go
8ff3e0823 fix(worktree): repair the worktree reaper — three-valued merge detection, lock-aware anchor guard, clean --stale inventory (t209) (#1638)
cf749fafe fix(worktree): guard done against live sessions anchored in the target tree

$ git merge-base --is-ancestor c72a517c3 ee50984abe4f11ac337382b48a26328f091e200a; echo $?
1
```
Neither commit touching `anchor_pid_windows.go` mentions or was authored by t298 — consistent with
the card's "diff 0건" attribution claim (t298's own commits, checked separately in the t320 entry
above, do not appear in this file's history either).

**Baseline-attribution.** `internal/session/anchor_pid_windows.go`, `anchor.go`, `session_pid.go` at
worktree HEAD `6165f9f5e`; file history and worktree-liveness check against pinned develop
`ee50984abe4f11ac337382b48a26328f091e200a`.

**Gaps.** I did not open the live worktree (`.claude/worktrees/t337`) to inspect what work is already
in progress there, per the restraint instruction and because this is a read-only sweep of the primary
tree's cards, not an audit of in-flight branches. I also did not verify the `probeProcessLiveness`
honest-undetermined path is actually wired to replace `isProcessAlive` anywhere (the card frames this
as the fix direction, not yet confirmed done).

**Residual-risk.** Since work is already underway on a live branch, this sweep's read-only verdict
could be stale by the time it's read — the in-flight branch may already contain a fix for exactly
this gap. The disposition proposal below accounts for that explicitly.

**Proposed disposition.** `already-landed` is NOT proposed (verified not an ancestor of develop, exit
1); given a live worktree already exists and the premise independently holds, `needs-operator-decision`
is proposed only to confirm whether the in-flight branch should simply be pushed to completion rather
than treated as a queue item — this rests on the `--is-ancestor` exit 1 evidence plus the existing
worktree entry.

**Overlap candidates.** none observed in-batch (no other batch-5 card touches
`internal/session/anchor_pid_windows.go`). Cross-batch: t298 is named directly in the card text but
is not in `inscope-all.txt` (already landed/closed per project memory).
