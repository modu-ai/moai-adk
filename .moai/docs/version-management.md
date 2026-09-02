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

### Local RC Numbering

Local (untagged, ldflags-injected) rc builds cut from `develop` — the RC testbed
flow whose build commands live in `.claude/rules/local/gitflow-lane-protocol.md`
§9 — carry their own `rc.N` policy. The release-tag progression above
(`rc.0 … rc.N → vX.Y.Z`) governs release tags; this section governs the local
candidate builds that never become tags. (SPEC-RC-TESTBED-001)

**Increment / reset policy:**

- `N` starts at `0` for each new target `vX.Y.Z`.
- `N` increments by exactly 1 per candidate build cut for that target — every
  `make build VERSION=vX.Y.Z-rc.N` (likewise `make install` /
  `make release-local`) for that target is one candidate.
- `N` resets to `0` only when the target `X.Y.Z` changes: the next release
  line's `N` starts fresh regardless of how far the previous line climbed —
  `v3.1.4-rc.7` is followed by `v3.1.5-rc.0`, never by `v3.1.5-rc.8`.

**No git tags for local rc builds.** A local rc build creates no tag and pushes
nothing; tagging remains a remote-facing act performed only by the release
harness (see Release Process below).

- **counter-precedent (predates this rule):** the `v3.1.0-rc.0` / `v3.1.0-rc.1`
  local *annotated* tags ("Local-only release candidate … NOT pushed" in the tag
  messages) were cut before the ldflags-only rule was codified. They are recorded
  here rather than silently rewritten — the v3.1.3 line ran entirely tagless, and
  every line since follows the no-tags rule.

**The version string does not order builds.** The monotone build identity is
`BUILD_ID` (`Makefile:19` — `git describe --tags --dirty`, injected via ldflags
into `pkg/version.BuildID`), not `VERSION`. An explicit rc VERSION reads HIGHER
than a later default build, so comparing version strings reaches the opposite
conclusion about which binary is newer. Incident `SPEC-BINARY-LAG-VISIBILITY-001`:
the installed `v3.1.2` binary was actually newer than the prior `v3.1.3-rc.5`
build, while every version-string comparison says the reverse. Read `BUILD_ID`
(and the embedded commit SHA) to decide which binary is newer; never order builds
by version string.

**`moai update` resets `git-strategy.yaml`.** The
`.moai/config/sections/git-strategy.yaml` keys — `rc_version_format: vX.Y.Z-rc.N`
among them, with `workflow: git-flow` and the develop/release branch keys — are
reset to template defaults by every `moai update` and must be re-applied after
each update (re-application procedure: CLAUDE.local.md §2.3).

The clean-reinstall runbook (`rm -f` + `cp`, the exit-137 guard) is owned by
`.claude/rules/local/gitflow-lane-protocol.md` §9 and is not duplicated here.

### Files Requiring Version Sync

When releasing new version, update:

The two groups below are different kinds of work. A bump commit **rewrites** every path under Version Stamps; it does not touch Release Artifacts, which are written fresh for each release.

**Version Stamps:**
- README.md (Version line)
- README.ko.md (Version line)
- README.ja.md (Version line)
- README.zh.md (Version line)
- .moai/config/sections/system.yaml (moai.version)
- docs-site/hugo.toml (version + `releaseDate` — the date is not a version token, so no check reads it; it still has to be updated by hand)
- pkg/version/version.go (the no-ldflags fallback — see Single Source of Truth above)

`internal/template/templates/.moai/config/sections/system.yaml.tmpl` is **not** a stamp: it renders `{{.Version}}` rather than carrying a literal version, so a bump leaves it alone. It is named here only so the next reader does not add it back.

**Release Artifacts:**
- CHANGELOG.md (New version entry — **English-only**; Korean lives in `.moai/release-notes/vX.Y.Z.ko.md`, NOT in CHANGELOG.md)
- .moai/release-notes/vX.Y.Z.ko.md (Korean release notes — consumed by the GitHub release body and the docs-site `ko` changelog page; `vX.Y.Z` is a placeholder, not a path on disk)

A guard test reads the Version Stamps list and fails when it names a path that is not in the working tree: `internal/cli/version_sync_list_test.go`, run by the existing `go test ./...` in CI.

The registry check added by card t392 sweeps every tracked file for the authoritative version token and excludes six groups, each for its own reason. The counts below are observations pinned to tree `051f209b0`, not constants — the check holds none of them. When a group's measured count stops matching its reason, re-derive the enumeration and rewrite this table; the clause to watch is the changelog-pages one, which hides nothing today.

| Excluded group | Why this group alone is excluded | Token files hidden |
|---|---|---|
| `.moai/reports/` | Session measurement records — time-pinned observations that a bump rewriting them would forge | 61 |
| `.moai/specs/` | SPEC bodies — historical narrative, immutable once complete | 62 |
| `.moai/release-notes/` | Written fresh for each release, not rewritten by a bump | 1 |
| `CHANGELOG.md` | Same — absent from the numstat of bump commit `61921f1ba` | 1 |
| `*_test.go` | Test header comments and SPEC frontmatter fixtures — no refresh obligation | 4 |
| the per-locale changelog pages under the docs-site content tree (a `changelog*` glob) | Change-history pages | 0 |

The six groups are disjoint and together hide 129 of the 163 tracked files carrying the authoritative token at that tree; the sweep sees the remaining 34.

Two checks now stand side by side. One catches the list naming a **path that does not exist**
(t388). The other catches a file carrying the authoritative version token that the **registry
does not name**, and a registered stamp that **does not carry the authoritative token** (t392).
The registry lives in `internal/cli/version_stamp_registry_test.go`.

Things still go uncaught. **At least the following remain, and this list is not exhaustive.**

1. A file that is neither in the registry nor carrying the authoritative token — an
   unregistered site left holding only an aged-out token matches no predicate.
2. A genuine stamp site registered as `prose` — completeness passes because it is registered,
   freshness is skipped because it is `prose`, and the documentation cross-check passes because
   it is in neither stamp set. All three assertions are blind to it.
3. A stamp inlined inside a file the exclusion set hides — had the same fixture lived in a
   `*_test.go` file rather than in `testdata/*.golden`, this predicate would not see it.
4. A stamp that is not a version token at all, such as `releaseDate` in `docs-site/hugo.toml`.
5. A site that renders the version rather than carrying it, such as
   `internal/template/templates/.moai/config/sections/system.yaml.tmpl`.
6. A file the repository does not track — outside the sweep's population.

None of this means the list can no longer rot.

Who maintains the registry, and when: the author of the commit that adds or removes a file
edits the registry in that same commit. **A version bump does not touch the registry** — it
rewrites the seven stamp files and adds or removes nothing. Between a file landing and the
registry edit, the check fails naming the path: a new token-carrying file is reported as
unregistered, and a deleted registered path is reported as unresolved.

### Release Process

Driven by the `hns-release-specialist` harness (`/harness:release`) under the PR-mandatory regime (`.moai/docs/git-local-workflow-doctrine.md` §18/§23). High-level:

1. Phase 4 — update `CHANGELOG.md` (**English-only**, commit-complete: cross-check `git log vPREV..HEAD`, no user-facing commit omitted; `docs:`/`chore:`/`style:`/`test:`/merge/typo commits excluded).
2. Phase 4.5 — author `.moai/release-notes/vX.Y.Z.ko.md` (Korean counterpart).
3. Phase 5 — human gate (release/abort).
4. Phase 6 — `release/vX.Y.Z` PR → **merge commit (NOT squash)** → `MOAI_RELEASE_VIA_HARNESS=1 ./scripts/release.sh vX.Y.Z` (annotated tag + provenance trailer, triggers GoReleaser; `verify-provenance` gates GoReleaser).
5. Phase 7 — GoReleaser publishes a `changelog.use: github` English body; overwrite with English (CHANGELOG) + Korean (`.moai/release-notes/vX.Y.Z.ko.md`) via `gh release edit vX.Y.Z --notes-file <merged>`.

[HARD] `CHANGELOG.md` English-only; GitHub Release English-first then Korean. A manual `git tag` + `make release VERSION=` (the older flow below) is NOT the release path under the PR-mandatory regime — `scripts/release.sh` is; the `make build VERSION=` snippet in §Build Version Injection above remains valid for local build-only injection.

> Legacy (build-only, not release): `make build VERSION=1.0.0` injects the version into a local binary via ldflags; it does NOT tag, push, or publish.
