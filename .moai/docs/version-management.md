# Version Management

> Relocated from `CLAUDE.local.md` §5 to reduce launch-time context weight. CLAUDE.local.md retains a 1-line pointer.

### Single Source of Truth

- [HARD] `go.mod` module version + git tags are the authoritative sources
- [HARD] `pkg/version/version.go` holds the version a build falls back to when ldflags inject none. It is a package-level `var` (it has to be, for `-X` to reach it), and every bump commit rewrites it by hand — it is not derived from git tags.

**Version Reference:**
- Authoritative Source: Git tags (e.g., `v1.0.0`)
- Build-time injection: `Makefile:20` defines `-X .../pkg/version.Version=$(VERSION)`, used by `make build` (`Makefile:36`) and `make install` (`Makefile:72`); release builds inject the same symbol at `.goreleaser.yml:22`. A binary from either path never shows the fallback.
- Fallback exposure: a bare `go install ./cmd/moai` carries no ldflags, so such a binary reports whatever `pkg/version/version.go` says. That hand-build is the only path where the fallback is user-visible, and it is why the line is a version stamp rather than a derived value.
- Config Display: `.moai/config/sections/system.yaml` (updated by release process)

### [HARD] Pre-release Versioning — SemVer 2.0.0

Canonical pre-release form: **`vX.Y.Z-rc.N`** (dot before the number, `N` starting at `0`).

```
v3.1.0-rc.0   v3.1.0-rc.1   ...   v3.1.0-rc.10      → then v3.1.0
```

The dot is not cosmetic. SemVer splits a pre-release on `.` and compares each
identifier separately, comparing a purely numeric identifier **numerically**.
So `rc.9 < rc.10` orders correctly. The older undotted form (`rc9`, `rc10`) is a
single alphanumeric identifier and compares ASCII-lexically, which puts `rc10`
*before* `rc9` — the failure surfaces only once a line reaches its tenth
candidate, which is exactly when a release is least forgiving.

Rules:

- [HARD] New pre-release tags use `-rc.N`. A numeric identifier carries no
  leading zero (`rc.1`, never `rc.01`) — SemVer forbids it and
  `scripts/release.sh` rejects it.
- Legacy undotted tags (`v3.0.0-rc12` and the eleven before it) stay valid and
  are NOT retagged; `scripts/release.sh` still accepts that form so the existing
  history remains reproducible. Do not author new tags in it.
- Enforcement lives in `scripts/release.sh` Validation 1, which implements the
  official SemVer 2.0.0 grammar (pre-release + optional build metadata).
- Local testing needs no tag at all — see Build Version Injection below.

### Build Version Injection

Version is injected at build time using ldflags:

```bash
# Build with version injection
go build -ldflags="-X github.com/modu-ai/moai-adk/pkg/version.Version=v1.0.0"

# Makefile handles this automatically
make build VERSION=v1.0.0
```

**Local pre-release testing needs no git tag.** `VERSION` is injected via ldflags
at build time, so a release candidate can be built and installed locally without
tagging, pushing, or invoking `scripts/release.sh`:

```bash
make build   VERSION=v3.1.0-rc.0   # bin/moai
make install VERSION=v3.1.0-rc.0   # $GOPATH/bin/moai
make release-local VERSION=v3.1.0-rc.0   # dist copy + version.json
```

Tagging is a separate, remote-facing act performed only by the release harness.

### Files Requiring Version Sync

When releasing new version, update:

The two groups below are different kinds of work. A bump commit **rewrites** every path under Version Stamps; it does not touch Release Artifacts, which are written fresh for each release.

**Version Stamps:**
- README.md (Version line)
- README.ko.md (Version line)
- README.ja.md (Version line)
- README.zh.md (Version line)
- .moai/config/sections/system.yaml (moai.version)
- internal/template/templates/.moai/config/config.yaml (moai.version)
- docs-site/hugo.toml (version + `releaseDate` — the date is not a version token, so no check reads it; it still has to be updated by hand)
- pkg/version/version.go (the no-ldflags fallback — see Single Source of Truth above)

`internal/template/templates/.moai/config/sections/system.yaml.tmpl` is **not** a stamp: it renders `{{.Version}}` rather than carrying a literal version, so a bump leaves it alone. It is named here only so the next reader does not add it back.

**Release Artifacts:**
- CHANGELOG.md (New version entry — **English-only**; Korean lives in `.moai/release-notes/vX.Y.Z.ko.md`, NOT in CHANGELOG.md)
- .moai/release-notes/vX.Y.Z.ko.md (Korean release notes — consumed by the GitHub release body and the docs-site `ko` changelog page; `vX.Y.Z` is a placeholder, not a path on disk)

A guard test reads the Version Stamps list and fails when it names a path that is not in the working tree: `internal/cli/version_sync_list_test.go`, run by the existing `go test ./...` in CI.

The guarantee it establishes is **partial**. It catches the list naming a path that does not exist. It **does not detect** a stamp site that is absent from the list — which is the direction that actually bit us: the v3.1.3 bump missed `docs-site/hugo.toml`, and nothing said so. Closing that direction is card t392.

### Release Process

Driven by the `hns-release-specialist` harness (`/harness:release`) under the PR-mandatory regime (`.moai/docs/git-local-workflow-doctrine.md` §18/§23). High-level:

1. Phase 4 — update `CHANGELOG.md` (**English-only**, commit-complete: cross-check `git log vPREV..HEAD`, no user-facing commit omitted; `docs:`/`chore:`/`style:`/`test:`/merge/typo commits excluded).
2. Phase 4.5 — author `.moai/release-notes/vX.Y.Z.ko.md` (Korean counterpart).
3. Phase 5 — human gate (release/abort).
4. Phase 6 — `release/vX.Y.Z` PR → **merge commit (NOT squash)** → `MOAI_RELEASE_VIA_HARNESS=1 ./scripts/release.sh vX.Y.Z` (annotated tag + provenance trailer, triggers GoReleaser; `verify-provenance` gates GoReleaser).
5. Phase 7 — GoReleaser publishes a `changelog.use: github` English body; overwrite with English (CHANGELOG) + Korean (`.moai/release-notes/vX.Y.Z.ko.md`) via `gh release edit vX.Y.Z --notes-file <merged>`.

[HARD] `CHANGELOG.md` English-only; GitHub Release English-first then Korean. A manual `git tag` + `make release VERSION=` (the older flow below) is NOT the release path under the PR-mandatory regime — `scripts/release.sh` is; the `make build VERSION=` snippet in §Build Version Injection above remains valid for local build-only injection.

> Legacy (build-only, not release): `make build VERSION=1.0.0` injects the version into a local binary via ldflags; it does NOT tag, push, or publish.
