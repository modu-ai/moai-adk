# t388 — measured baseline (lane-4, plan-phase)

Tree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t388
HEAD: 9328a5242 (== origin/develop at measurement time)
Branch: WT-version-sync-list

## Claim 1 — the doc list names a path that does not exist

Command:
    git cat-file -e origin/develop:internal/template/templates/.moai/config/config.yaml

Output:
    fatal: path 'internal/template/templates/.moai/config/config.yaml' does not exist in 'origin/develop'
    exit=128

Command:
    git ls-tree origin/develop internal/template/templates/.moai/config/ --name-only

Output:
    internal/template/templates/.moai/config/astgrep-rules
    internal/template/templates/.moai/config/evaluator-profiles
    internal/template/templates/.moai/config/sections

## Claim 2 — the actual bump target set is 7 files / 9 lines

Command:
    git show --stat 61921f1ba      # "chore: bump version to v3.1.4"

Observed changed paths + lines (from the diff body):
    pkg/version/version.go:8
    .moai/config/sections/system.yaml:45,47
    README.md:24
    README.ko.md:24
    README.ja.md:24
    README.zh.md:24
    docs-site/hugo.toml:55,56          # version + releaseDate

Attribution: this is the diff of the v3.1.4 bump commit itself, not a table
carried over from a report.

## Claim 3 — 4 omissions + 1 ghost

Doc section "Files Requiring Version Sync" (.moai/docs/version-management.md:71-78) lists:
    README.md, README.ko.md, CHANGELOG.md, .moai/release-notes/vX.Y.Z.ko.md,
    .moai/config/sections/system.yaml,
    internal/template/templates/.moai/config/config.yaml   <- ghost

Omitted vs the measured set: README.ja.md, README.zh.md, docs-site/hugo.toml,
pkg/version/version.go.

## Claim 4 — version.go is a contradiction, not a plain omission

Command:
    grep -n -i 'version\.go|ldflags|pkg/version' .moai/docs/version-management.md

Output (relevant):
    8:- [HARD] `pkg/version/version.go` reads from git tags at build time
    12:- Runtime Access: `pkg/version/version.go` via `git describe`

The doc asserts version.go is DERIVED. Every bump commit hand-edits
version.go:8. Both cannot be true as written.

## Claim 5 — the doc is stale

Command:
    git log --oneline -1 origin/develop -- .moai/docs/version-management.md

Output:
    6422046bb feat(SPEC-HARNESS-TOKEN-OPT-001): trim ~40KB always-loaded rules tax + A9 consume-path default (#1443)

## Gaps (explicitly NOT observed)

- Whether CHANGELOG.md / .moai/release-notes/*.ko.md belong in the same list as
  version stamps was NOT adjudicated here; the bump commit does not touch them,
  so the list currently conflates two different artifact classes.
- No guard test was run; none exists yet. The regression mechanism is unbuilt.
- Bump commits before eba919e44 were not inspected.

## Residual risk

- The set was measured from ONE bump commit (61921f1ba). A stamp site that
  exists but was missed by that commit too would not appear here. hugo.toml is
  precedent for exactly this shape (missed at v3.1.3, fixed later by t274).
