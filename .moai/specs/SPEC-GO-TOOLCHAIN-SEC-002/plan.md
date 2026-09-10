# Implementation Plan — SPEC-GO-TOOLCHAIN-SEC-002

> Tier S, Class C (global change). Card t610 (Factory lane-8). Artifacts: spec.md + plan.md +
> acceptance.md + progress.md. acceptance.md is kept at Tier S to carry the RED-now evidence
> ledger, mirroring the precedent SPEC-GO-TOOLCHAIN-SEC-001.

## A. Context

Bump the `go.mod` `go` directive from `go 1.26.4` to `go 1.26.8` so that `govulncheck ./...`
reports 0 affecting vulnerabilities (8 standard-library findings → 0). This is Go full review
2026-09-10, finding F01 (evidence item SEC-01). The go.mod single source of truth for CI came
from SPEC-GO-TOOLCHAIN-SEC-001, so `go.mod` line 3 is the only file this run phase edits.

## B. Known Issues / Ground Truth (baseline on `d3b7d438d`, evidence `.moai/reports/t610/baseline/`)

- Command record: `.moai/reports/t610/baseline/commands.md`.
- `go.mod:3` = `go 1.26.4`. There is one go.mod, no `toolchain` directive, and no go.work.
- CI: every `actions/setup-go@v7` step reads `go-version-file: go.mod`. No hard-coded Go
  version exists in CI, the Makefile, or a release config (`ci-go-version-refs.txt`).
- Local toolchain: the installed Go is go1.26.0 (recorded at `commands.md:22-24`) and
  `GOTOOLCHAIN=auto`. The effective go1.26.4 is auto-acquired from the directive
  (`goversion-auto.txt`). go1.26.6 and go1.26.8 download on demand (`goversion-go1.26.6.txt`,
  `goversion-go1.26.8.txt`).
- govulncheck on the same tree, varying only the toolchain: go1.26.4 exits 3 with 8 affecting
  findings. go1.26.6 exits 0 with 0 affecting. go1.26.8 exits 0 with 0 affecting
  (`govulncheck-*.log`, `govulncheck-exits.txt`).
- The card base has already fallen behind local `develop`. `git rev-parse --verify develop` →
  `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`, while `git merge-base develop HEAD` →
  `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1`. This is why the diff ACs use the merge-base form
  (acceptance.md § D.0a).
- Precedent lesson (SPEC-GO-TOOLCHAIN-SEC-001 §E.2.1): on go1.26.x, a `toolchain` directive
  equal to a full-patch `go` directive is redundant. `go build` then fails with "updates to
  go.mod needed". This is why D3 keeps the directive-only form.

## C. Pre-flight (run-phase entry)

- `govulncheck` is available in the run-phase environment. It already ran to produce the
  baseline.
- Network access to the Go toolchain proxy is available, so `GOTOOLCHAIN=auto` can download
  go1.26.8.
- `git status --short` shows only `.moai/reports/t610/` and this SPEC directory as local
  changes before M1.

## D. Constraints

- **Scope guard.** The run-phase commits touch `go.mod` line 3 plus evidence and SPEC
  artifacts only. No other tracked file changes (AC-GTS2-003, AC-GTS2-007). The sync phase
  adds CHANGELOG.md and the 4 project documents of M5.
- **Verification scope.** Run `make build`, `go vet ./...`, and the tests of packages the run
  phase selects as affected. The selection MUST include `internal/web` tests, as a regression
  guard for the net/http server surface: go1.26.7 changes net/http (#80927), and
  `internal/web` runs an `http.Server` (`internal/web/server.go:230`). Those tests do not
  cover the #80927 path itself. That path needs unencrypted HTTP/2, which this repo does not
  configure (see § D2). Do NOT run the full local test suite (`go test ./...`). The full-suite
  verdict comes from CI on the lead's `origin/develop` push after M4.
- **[HARD] `internal/cli` slot.** Any execution of `internal/cli` tests needs the lead's
  explicit slot approval first. Without that approval, `internal/cli` is excluded from the
  local package selection, and the exclusion is recorded in progress.md §E.2.
- **Evidence instrument contract.** Each judged command's output goes to a file under
  `.moai/reports/t610/`, and its exit code is recorded unpiped (for example
  `<cmd> > <file> 2>&1; echo $? > <file>.exit`). Never pipe judged output through
  `| head`, `| tail`, or `| grep` before its exit code is captured.
- **No toolchain override in judged runs.** The judged govulncheck and build runs use the
  bumped `go.mod` with `GOTOOLCHAIN` unset or `auto`, recorded with
  `go -C <worktree-root> env GOTOOLCHAIN GOMOD` in the same evidence file. A
  `GOTOOLCHAIN=go1.26.x` override proves the scanner, not the repository change.
- **Integration.** No card PR (repo-local git-flow policy). The lane reports its local
  `develop` merge SHA, and the lead batch-pushes `origin/develop`.
- **No source changes.** If a compile error or test regression is attributable to the
  toolchain change, STOP and return a blocker report. The fix would be a separate SPEC.

## D-DESIGN. Decisions (lane plan decisions — the Implementation Kickoff Approval gate may override)

### D1 — Toolchain acquisition: GOTOOLCHAIN auto-switch from the go.mod directive

**Decision.** Acquire go1.26.8 through `GOTOOLCHAIN=auto` reading the `go 1.26.8` directive.
Do not install Go locally.

**Rationale.** The build today already works this way. The installed Go is go1.26.0, yet the
effective toolchain is go1.26.4, taken from the directive (spec.md § A fact 3). The target
version downloads on demand (`go: downloading go1.26.8 (darwin/arm64)`,
`goversion-go1.26.8.txt`). CI follows the same directive through `go-version-file: go.mod`, so
local and CI stay on one source.

**Alternative.** Install go1.26.8 locally (Homebrew or golang.org download) and pin with
`GOTOOLCHAIN=local`. Rejected: it moves setup work outside the repo, it is per machine, and it
reintroduces the drift the single source of truth exists to prevent.

**Residual risk.** An environment with `GOTOOLCHAIN=local` and an older installed Go, or with
no network path to the toolchain proxy, fails with `go: go.mod requires go >= 1.26.8`. This is
recorded as a risk. Fixing it is out of scope (spec.md § E).

### D2 — Target version: go 1.26.8 (floor go1.26.6)

**Decision.** Target `go 1.26.8`, the latest 1.26 patch release. go1.26.6 is the proven floor.

**Rationale.** Both versions clear all 8 findings (spec.md § A fact 4, identical govulncheck
output). The target is justified by what 1.26.6→1.26.8 actually adds, not by being "latest".

**Delta 1.26.6→1.26.8 (measured).** It adds no security fixes. 1.26.7 fixes a net/http
ReadHeaderTimeout regression after unencrypted HTTP/2 handoff (#80927), a behavior fix on the
path the 1.26.6 security change (GO-2026-6089) touched. 1.26.8 fixes cgo flags (#80851),
debug/elf PPC relocations (#81113), an openbsd-only os test (#80889), a netbsd/amd64-only
runtime AVX issue (#80827), and a compiler test (#81152). None of the 1.26.8 fixes reaches the
shipped binary:

- #80851: the release artifact is built with `CGO_ENABLED=0` (`.goreleaser.yml:12`).
- #81113: no tracked Go file imports `debug/elf` (`git grep -n '"debug/elf"' -- '*.go'` → no
  output; judged on stdout, since the exit value was not independently observable through the
  tool). The package reads ELF on any host, so the release platforms are not the
  reason.
- #80889 and #80827: the release targets are linux, darwin, and windows
  (`.goreleaser.yml:13-16`).
- #81152: a compiler test.

The one 1.26.6→1.26.8 change in a package this project uses is 1.26.7's net/http #80927,
"ReadHeaderTimeout remains active after unencrypted HTTP/2 handoff". This repo does not
configure unencrypted HTTP/2. Measured on `d3b7d438d`:

- `git grep -n -E 'UnencryptedHTTP2|h2c|x/net/http2' -- '*.go' go.mod` → no output (judged on
  stdout; the exit value was not independently observable through the tool).
- `git grep -n -E 'ReadHeaderTimeout|Protocols' -- internal/web` →
  `internal/web/server.go:230:		ReadHeaderTimeout: 10 * time.Second,`, exit 0 (no `Protocols`
  field on that server).

So the #80927 handoff path is not configured here. The `internal/web` tests in § D stay in
the selection as a general net/http regression guard. They are not a check of that path.

Evidence is under `.moai/reports/t610/baseline/`. `go-release-history.html` labels 1.26.6
"includes security fixes"; 1.26.7 and 1.26.8 read "includes fixes to …" without a security
label. Milestone issue lists come from `go1.26.7-milestone-issues.json` and
`go1.26.8-milestone-issues.json`, and the CGO setting from `cgo-enabled-refs.txt`.

**Gap.** The individual issue diffs were not read. Whether Go's `http.Server` serves
unencrypted HTTP/2 by default was not measured, so the conclusion above rests on the repo grep
(no h2c configuration), not on the Go default. The impact of #80927 on this project is a Gap,
not a measured claim.

**Alternative.** Exactly `go 1.26.6`, the review report's minimum. Smaller distance from
go1.26.4, but it lands already behind the latest patch. go1.27.x is not an alternative: a
minor bump is out of scope.

**Edge.** If govulncheck under go1.26.8 reports a new affecting finding published after this
plan, the target stays "0 affecting". The run phase returns a blocker naming the finding and
its fix version rather than silently picking another version.

### D3 — Directive form: only the `go` directive on line 3, no `toolchain` line

**Decision.** Change `go.mod` line 3 from `go 1.26.4` to `go 1.26.8`. Add no `toolchain`
directive.

**Rationale.** This matches the precedent's final form and the go.mod single source of truth
read by `go-version-file`. A `toolchain` line equal to the full-patch `go` directive is
redundant on go1.26.x and breaks `go build` (precedent §E.2.1).

**Accepted consequence.** The `go` directive is also the minimum Go version for any module
importing this one. This module ships a CLI binary and is not consumed as a library
dependency, so the stricter minimum is accepted, as it was in the precedent.

**Alternative.** Keep `go 1.26.4` and add `toolchain go1.26.8`. That raises the build toolchain
without raising importers' minimum. Rejected: it splits the version across two directives
where the single source of truth wants one, and the importer-minimum concern does not apply
here.

**Gap.** How `go-version-file` treats a `toolchain` directive was not verified in this plan,
and the rejection above does not depend on it. The precedent's §E.2.1 observation covers only
a `toolchain` line equal to the `go` directive, not this split form.

### D4 — Global landing: merge alone in its own integration window

**Decision.** Once merged, the directive changes every lane's and every CI job's build
toolchain. The change therefore lands **alone** in its own integration window, not batched
with other card merges. Before-and-after `go version` and govulncheck output are recorded.

**Rationale.** If other card merges land in the same window, a CI regression on
`origin/develop` cannot be attributed to the toolchain versus those merges. Landing alone
keeps the attribution clean.

**Alternative.** Batch with the other cards in the next window. Fewer windows, but a mixed
CI failure would need bisecting across merges.

## E. Self-Verification (run-phase)

1. AC-GTS2-001 first. `go -C <worktree-root> version`, after the edit, reports go1.26.8, and
   `go -C <worktree-root> env GOTOOLCHAIN GOMOD` is in the same evidence file.
   AC-GTS2-002..007 evidence is not judged until this gate passes.
2. `go.mod` directive and committed diff against the merge-base with `develop`: exactly the
   directive line (AC-GTS2-002, AC-GTS2-003).
3. govulncheck without an override: exit 0, 0 affecting, 8 IDs absent, the "modules you
   require" line present (AC-GTS2-004).
4. `make build` exit 0 and `go version -m bin/moai` shows go1.26.8, kept as evidence
   (AC-GTS2-005).
5. `go vet ./...` plus the selected package tests with a non-empty swept set (AC-GTS2-006).
6. Scope guard: no committed tracked-file change outside `go.mod`, this SPEC directory, and
   `.moai/reports/t610/` (AC-GTS2-007).
7. At M4, AC-GTS2-001..008 (items 1–6 plus the M5 doc counts) are re-measured on the absorbed
   tree with `HEAD` and the tree id recorded, and the 5 sync paths' card-side numstat is
   compared with the M5 sync commit (acceptance.md § D.0b).

## F. Milestones (ordered by decision reversibility, no time estimates)

Execution order is M1 → M2 → M3 → M5 → M4. Sync completes in the card worktree before the
integration merge (CLAUDE.local.md:393). The milestone numbers are kept for AC traceability.

- **M1 — Directive bump (D2 + D3, most likely to change at kickoff).** Edit `go.mod` line 3
  to `go 1.26.8` and commit. Pass the AC-GTS2-001 gate, then capture AC-GTS2-002 and
  AC-GTS2-003 (the diff ACs read commits, not the working tree).
- **M2 — Security verdict.** Run govulncheck under the bumped go.mod with no override and
  write the output and exit code to `.moai/reports/t610/`. Judge AC-GTS2-004 against the
  baseline list.
- **M3 — Build and regression evidence.** Record `git status --porcelain --untracked-files=no`
  before and after `make build`. `make build` regenerates tracked files: `templ generate` over
  `internal/web` (8 tracked `*_templ.go`) and `gen-catalog-hashes.go --all`
  (`internal/template/catalog.yaml`). Any tracked file it changes is reported in progress.md
  §E.2 and not committed in this work. Then run `go version -m bin/moai` (kept),
  `go vet ./...`, and the selected package tests (the `internal/cli` slot rule applies).
  Judge AC-GTS2-005, AC-GTS2-006, and AC-GTS2-007.
- **M4 — Integration (D4), after M5.**
  1. Take the integration window.
  2. Absorb local `develop` into the card branch with `git merge develop`
     (CLAUDE.local.md:389).
  3. Re-measure AC-GTS2-001..008 on the absorbed tree and record `git rev-parse HEAD` and
     `git rev-parse 'HEAD^{tree}'` (acceptance.md § D.0b). Always re-measure. M1–M3 evidence
     is never the post-merge verdict, because the M5 sync commit changes the tree after M3.
  4. Record the card-side numstat of the 5 M5 sync paths and confirm it matches the M5 sync
     commit's own numstat for those paths (acceptance.md § D.0b step 5).
  5. Merge alone in the dedicated window, report the local `develop` merge SHA, and read the
     CI verdict after the lead's push.
- **M5 — Sync (runs before M4).** Update the 5 doc mentions (`.moai/project/product.md:244,300`,
  `.moai/project/structure.md:132`, `.moai/project/codemaps/overview.md:6`,
  `.moai/project/codemaps/modules.md:6`) to 1.26.8 (AC-GTS2-008). Add a CHANGELOG security
  entry and the status transition.

## G. Anti-Patterns to Avoid

- Judging the security AC from a `GOTOOLCHAIN=go1.26.8` override run instead of the bumped
  go.mod. That proves the scanner, not the change.
- Accepting a "0 affecting" result while the effective toolchain is still go1.26.4 (skipping
  AC-GTS2-001).
- Adding a `toolchain` directive, or touching `require` lines "while here".
- Editing workflows to pin 1.26.8. They already follow go.mod.
- Running `go test ./...` locally, or `internal/cli` tests without the lead's slot approval.
- Piping judged output through `head`, `tail`, or `grep` before capturing its exit code.
- Reading a green test run without checking the swept set is non-empty.
- Judging the diff ACs against a fixed SHA or the moving `develop` tip instead of the
  merge-base form, or on the working tree instead of commits.
- Reusing pre-absorption evidence as the post-merge verdict.
- Committing tracked files that `make build` regenerated.

## H. Cross-References

- spec.md § C (requirements), § E (exclusions)
- acceptance.md (AC matrix, RED-now evidence ledger, post-absorption re-measure, closure gates)
- `.moai/reports/t610/plan-audit.md` (plan-audit iteration 1, resolved by revision 0.1.1)
- `.moai/reports/t610/plan-audit-2.md` (plan-audit iteration 2, PASS; residuals R1–R6 folded into revision 0.1.2)
- `.moai/specs/SPEC-GO-TOOLCHAIN-SEC-001/` (precedent: go-version-file SSOT, toolchain directive lesson)
- `.moai/reports/t610/baseline/` (baseline evidence; `commands.md` is the command record)
- `.claude/rules/local/gitflow-lane-protocol.md` (integration window, lead batch push)
- `CLAUDE.local.md` §4.1 (:389 absorb local develop, :393 sync before merge)
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, §4 (swept-set, two-cell AC discipline, evidence pinning)
