---
card: t2147
reviewed_sha: 8d1891b97a1ff899ed1342966b5555d65e8a9a29
reviewer: lane-28 self-review (Claude)
codex: cited separately in codex-receipt.md (rule d-20261106T080039Z-305c, at most 2 per stage)
---

# t2147 card review

## Diff reviewed (f9d8866951..8d1891b97a; bump commit 801eba9d92)

| File | Change |
|---|---|
| `pkg/version/version.go` | `v3.2.0` → `v3.2.1` |
| `.moai/config/sections/system.yaml` | `template_version` and `version`: `v3.2.0` → `v3.2.1` |
| `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` | Release badge `v3.1.2` → `v3.2.1`. The badge had regressed to `v3.1.2` in 840053635f. The historical sentence at line 260, "From v3.2.0", is unchanged. |
| `docs-site/hugo.toml` | `version` → `v3.2.1`; `releaseDate` → `2026-11-07` (see Debt) |

The template `system.yaml.tmpl` takes its value from `{{.Version}}`, so it was not edited and `make build` was not needed.

The edits follow the v3.2.0 bump `a85be12449`, cross-checked against v3.1.3 `eba919e44b`. The extra test changes in v3.2.0 are not needed here: the stamp registry already covers the v3.2.1 citations, and the role-rules test reads `version.GetVersion()`.

## Lane observations

- Lane runs: `go build ./...` exited 0, and `go test -count=1 ./pkg/version/` passed (`ok`).
- Auditor runs:
  - `TestVersionStamp*`, `TestVersionSync*` and `TestRoleRules*` pass.
  - `TestReleasePreflight*` does not exist in this repo.
  - The trial merge onto the batch-FF candidate `f05814ad66` with `git merge-tree` is conflict-free.

## Debt (non-blocking)

- **D1:** `docs-site/hugo.toml:56` has `releaseDate = "2026-11-07"`. That is the bump day, not a confirmed release date. Fix it on tag day, or have the leader accept it. The docs site auto-deploys from main.
- **Out of card scope:**
  - `CHANGELOG.md` has no `## [3.2.1]` section, and `scripts/release.sh:210-216` stops a formal tag without one.
  - `.moai/docs/version-management.md:146-150` still names a removed `make plugin-emit` step.
