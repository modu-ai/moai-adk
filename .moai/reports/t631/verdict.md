# t631 — verdict (lane-3, pre-integration)

Card: t631 · Branch: `WT-spec-ownership-ssot` · Base HEAD: `dfe6531ecf7f5af6c10a5410a03430dcddd14482`
(parents `d060e0d13` develop + `f66cdc918` t615) · Toolchain at measurement: `go1.26.4 darwin/arm64`
(pre-absorb; develop `296ba7aa5`+ moves to go 1.26.8 — every figure below is re-measured in the
integration window).

Mode: no SPEC; the card text is the spec (operator decision). Edits are Template-First, targeted,
applied identically to the template copy and then the local copy (no cp-mirroring).

## Claim

| ID | Site (both copies) | Before | After |
|---|---|---|---|
| SX-01 | `skills/moai/workflows/plan/spec-assembly.md:252` | thorough cross-validation invoked sync-auditor "in SPEC-review mode" — a mode sync-auditor does not have (post-implementation only; modes final-pass / per-iteration) | independent plan-auditor re-review (fresh spawn, first verdict/score/findings withheld); sync-auditor role boundary cited by path |
| SX-R03 | `skills/moai/workflows/run.md:191` | `Foreground` row forcing `run_in_background: false` — contradicts agent-common-protocol § Background Agent Execution (runtime chooses; safeguard is concurrency) | `Concurrency` row: single write-capable agent, orchestrator work read-only, foreground/background left to runtime |
| SX-R02 | `skills/moai/workflows/run.md:193` | ledger heading lettered `## §E Recursive Self-Diagnosis Log` — SSOT Section Map allocates `§H`, and `§E` is reserved for era.go-parsed sections | letter removed; row points at `spec-frontmatter-schema.md` § progress.md Section Map (grep recipe is letter-agnostic, unchanged) |
| SX-R06 | `skills/moai/workflows/sync.md:41` | manager-docs owns "`in-progress → implemented` transition" — SSOT matrix has manager-docs owning `in-progress → implemented → completed` on the single sync commit | transition stated by reference to § Status Transition Ownership Matrix; `§E.4` kept with a Section Map pointer |
| SX-06 | `skills/moai/workflows/plan.md:37` | already correct (`§E.1`) — fixed earlier by a re-copy, not by this card | `§E.1` kept with a Section Map pointer |
| — | `internal/template/catalog.yaml:9` | `moai` entry hash `1d23838d…` | `fa683eb7…` (generator dry-run value; one-line refresh, same shape as `bdbc09788`) |

## Evidence

- Reproduction on develop `f66cdc918` (before edits): `repro-sweep.txt`, `premise-checks.txt`,
  `sync-auditor-loose-check.txt`, `sibling-sweep.txt`.
- Old phrases absent after edits: grep over both workflow trees for `SPEC-review`,
  `run_in_background: false`, `§E Recursive Self-Diagnosis`, `background-write prohibition`,
  `` in-progress → implemented` transition `` → exit 1 (no match). Positive control: same grep on
  `git show HEAD:…/run.md` → count 2.
- New phrases present: `post-new-phrases.txt` (spec-assembly 1/1, run 2/2, sync 1/1 per copy).
- Local↔template parity: `post2-diff-*.txt` — spec-assembly and run diff 0 lines; sync (15) and plan
  (6) diffs are byte-identical to the pre-edit divergence files (`cmp` exit 0) — neutrality-driven
  divergence only.
- Cf format characters in the 8 edited files: `Cf=0`; control file containing one U+200B: `Cf=1`.
- Template neutrality of added lines (`template-diff.txt`): the only hit is the `<SPEC-ID>`
  placeholder already present at HEAD (`git show HEAD:… | grep -c` → 1); control grep on a file with
  internal IDs → 5.
- Tests (`go1.26.4`, this tree):
  - `guard-test-final.txt`: `TestPhaseSignalCitationsNameTheParsedSection`,
    `TestSpecArtifactSetIsTierScoped` → 2 PASS.
  - `cli-targeted-tests.txt`: 9 selected internal/cli tests that read the edited files → 9 PASS,
    0 FAIL/SKIP.
  - `pkg-tests-preabsorb.txt`: template/spec/skills/harness/lsp-config packages → template and spec
    FAILED on catalog hash parity (`CATALOG_HASH_DRIFT` entry `moai`), all others ok.
  - `pkg-tests-after-hash.txt` (after the catalog refresh): `internal/template` ok,
    `internal/spec` ok, exit 0.

## Baseline-attribution

All measurements above ran in this worktree against `dfe6531ec` plus the uncommitted edits listed in
Claim, toolchain `go1.26.4`. No figure is carried from another tree. The integration-window absorb
of local develop (≥ `296ba7aa5`, go 1.26.8) invalidates them as merge evidence until re-run.

## Gaps

- Not run: `internal/cli` full package (deliberately — local full-suite prohibition); only the 9 tests
  that read the edited files. `internal/harness` subpackages and `internal/lsp/config` ran as part
  of the pre-absorb package batch (ok), not re-run after the hash refresh (that refresh touches only
  `catalog.yaml`).
- Not run: `make build` (lanes do not build); `make commands-emit-check` not invoked directly — its
  golden tests (`internal/template/commandemit`) passed.
- CI (darwin/windows matrix) has not seen this change.

## Residual-risk

- **Operator choice partially reversed, recorded.** The operator first chose to replace `§E.1`/`§E.4`
  with Section Map references. The t408 guard (`internal/spec/progress_section_letter_guard_test.go`)
  requires the literal `progress.md §E.` in all four plan/sync copies as an anti-vacuity rule; the
  conflict was surfaced and the operator then chose "keep letter + add reference". The recommendation
  of the first option was made without checking existing guards — a process miss, not a code defect.
- SX-01 semantics: the new text defines what the re-review receives but no mechanism enforces "first
  verdict withheld" — it binds the orchestrator's spawn prompt only.
- **Sibling candidate, out of scope, not fixed:** `.claude/agents/harness/hns-release-update-specialist.md:152`
  and `:162` still instruct `run_in_background: false` (local-only dev harness, user-owned
  namespace). Reported to the lead.
- `catalog.yaml` hash is order-sensitive to any other card touching `templates/.claude/skills/moai/`;
  a conflicting refresh at merge time must be regenerated, not hand-picked.
