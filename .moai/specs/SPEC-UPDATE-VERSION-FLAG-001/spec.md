---
id: SPEC-UPDATE-VERSION-FLAG-001
title: "Add --version <tag> flag to moai update for release-tag binary install"
version: "1.0.0"
status: draft
created: 2026-08-03
updated: 2026-08-03
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "cli, update, binary, security, release"
tier: M
---

# SPEC-UPDATE-VERSION-FLAG-001

## §A Overview

### §A.1 Problem

The `moai update` command can only move the moai binary forward to GitHub's
`/releases/latest` (which auto-excludes pre-releases). A user who needs to
install a **specific** release tag — pin to a known-stable version, switch
to an rc for testing, or roll back to a previous version after a regression
— has no first-class CLI path today. The only mechanism is the external
`install.sh --version VERSION` script, which is not surfaced from `moai`
itself and which bypasses the in-process checksum-verified downloader.

### §A.2 Solution

Add a `--version <tag>` flag to `moai update` that resolves `<tag>` to a
specific GitHub Release asset within the existing `api.github.com`
allowlist, downloads it through the existing checksum-verified binary
download path (`deps.go` `buildAutoUpdateFunc` / `EnsureUpdate` /
`UpdateOrch`), and installs it as a replacement for the running binary.
One flag covers three use cases: stable pin, rc switching, and
downgrade/rollback.

### §A.3 Goals

- Unify release-tag binary install inside `moai update` (no external script).
- Cover stable / rc / previous-version tags with a single flag.
- Preserve every existing default-`moai update` invariant (stable-only
  auto-update, checksum verification, allowlist).

### §A.4 Non-Goals

- Config/template rollback via `--version` (template state is the existing
  update flow's job: `--restore`, template sync, 3-way merge). Because the
  moai binary `go:embed`s templates, installing an rc binary naturally
  carries rc templates — no separate template-version integration is in
  scope.
- A new downloader. The existing checksum-verified path is reused.

## §B Background and Verified Context

The following facts were verified against the source tree this session
(cite, do not re-derive):

- `internal/cli/deps.go:34` defines
  `githubLatestReleaseURL = githubReleasesURL + "/latest"`, used at
  `deps.go:391`. GitHub's `/releases/latest` endpoint auto-EXCLUDES
  pre-releases, so rc is NOT picked up by the default update flow today.
  This is an intended safety property — preserved by this SPEC.
- `internal/cli/update.go:62-72` registers `moai update` flags today:
  `--check`, `--config`, `--force`, `--yes`, `--templates-only`,
  `--binary`, `--dry-run`, `--no-hooks`, `--restore`, `--verbose`,
  `--shell-env`, `--profile`. There is NO version flag.
- `install.sh:294-300` already supports `--version VERSION`; default
  fetches the latest `tag_name`. `install.sh:113-115` constructs the
  archive and checksum URLs from the resolved version.
- `deps.go:443` `buildAutoUpdateFunc` and `EnsureUpdate` carry the
  checksum-verified binary-download path. Reused unchanged.
- `deps.go:53` `allowedUpdateHost = "api.github.com"`;
  `MOAI_UPDATE_URL` is validated against the allowlist at `deps.go:408`
  `validateUpdateURL`. `--version` resolution MUST stay on this allowlist.
- `deps.go:387` has an existing **Dev/RC version branch**: when the
  *current* version looks like a dev/rc/`go-v` build, `apiURL` is set to
  `githubReleasesURL` (the unfiltered list endpoint) instead of the
  `/latest` filter. This is the path that lets a dev build self-update at
  all. `--version` MUST NOT silently break this branch.
- Checksum verification is MANDATORY with no bypass
  (`docs-site/content/en/cli-reference/update.md:45-66`, CWE-345 policy).

## §C Requirements (GEARS notation)

### REQ-UVF-001 (Ubiquitous) — flag registration

The `moai update` command shall register a `--version <tag>` string flag
that accepts a GitHub release tag (stable, rc, or previous version) and
defaults to the empty string.

### REQ-UVF-002 (Event-driven) — install specific tag

**When** a user invokes `moai update --version <tag>`, the update
subsystem shall resolve `<tag>` to a specific GitHub Release via
`GET /repos/modu-ai/moai-adk/releases/tags/<tag>` on the
`api.github.com` host, download the matching binary asset through the
existing checksum-verified download path, and install it as a
replacement for the running binary.

### REQ-UVF-003 (Capability gate) — v-prefix normalization

**Where** the supplied `<tag>` lacks a leading `v` prefix, the update
subsystem shall normalize it to the canonical `v<prefix>` form before
the tag-resolution HTTP call, reusing the v/V-prefix-stripping
discipline of `normalizeVersionMajor`
(`internal/cli/v2_detection.go:253`) — the existing function extracts
the major integer; full-semver canonicalization to the `v`-prefixed
form is new behavior the implementer must write.

### REQ-UVF-004 (State-driven) — preserve default stable-only behavior

**While** `--version` is not supplied, the `moai update` command shall
remain behavior-identical to today: it shall fetch GitHub's
`/releases/latest` (which auto-excludes pre-releases) and shall NOT
surface rc or pre-release tags to the user.

### REQ-UVF-005 (Capability gate) — allowlist confinement

**Where** `--version <tag>` triggers tag resolution, the update
subsystem shall confine every HTTP request to the `api.github.com` host
on the `https` scheme (the existing `allowedUpdateScheme` /
`allowedUpdateHost` allowlist), and shall reject any environment or
input that would escape the allowlist, fail-closed, before any network
call.

### REQ-UVF-006 (Ubiquitous) — checksum verification mandatory

The update subsystem shall verify the downloaded binary against the
release's published checksum before installation, and shall provide no
`--skip-checksum` or analogous bypass for `--version` installs.

### REQ-UVF-007 (Event-driven) — mutual-exclusion matrix

**When** `--version` is supplied together with `--check`, `--templates-only`,
`--restore`, or `--dry-run`, the update subsystem shall exit non-zero
with a usage error naming the conflict, before any network call. The
combination with `--binary` is permitted (binary-only install of the
requested tag); the combination with `--force` is permitted (force
re-install of the requested tag even when the running version already
matches).

### REQ-UVF-008 (Event-detected) — tag not found

**When** the tag-resolution HTTP call returns HTTP 404, the update
subsystem shall exit non-zero with a structured error naming the
offending tag and directing the user to `moai update --check` or the
GitHub releases page, without attempting any download.

### REQ-UVF-009 (Event-detected) — release has no binary asset

**When** the resolved release carries no installable binary asset for
the running `GOOS/GOARCH` (e.g. an rc whose GoReleaser run produced
none), the update subsystem shall exit non-zero with a structured error
naming the tag, the detected platform, and the assets that WERE present,
without writing anything to the filesystem.

### REQ-UVF-010 (Event-detected) — checksum mismatch

**When** the computed checksum of the downloaded binary does not match
the published checksum for the tag, the update subsystem shall discard
the downloaded bytes, shall NOT install the binary, and shall exit
non-zero with a CWE-345-referencing error naming the tag.

### REQ-UVF-011 (Event-detected) — network failure

**When** a network failure occurs during tag resolution or binary
download, the update subsystem shall exit non-zero with a wrapped error
that distinguishes the phase (resolution vs. download) and names the
underlying error, without partial installation.

### REQ-UVF-012 (State-driven) — dev/RC branch interaction

**While** the *currently running* binary is itself a dev/rc/`go-v`
build, the update subsystem shall honor an explicit `--version <tag>`
request by resolving `<tag>` directly (NOT by falling back to the
existing dev-branch `/releases` list endpoint), so the dev-branch
self-update path continues to work for default `moai update` and
`--version` overrides it explicitly.

### REQ-UVF-013 (Event-detected) — downgrade intent confirmation

**When** the resolved tag is older than the currently running version
(a downgrade/rollback), the update subsystem shall, in interactive
mode, prompt the user to confirm the downgrade before installation;
**Where** `--yes` is supplied OR stdin is not a TTY, the prompt shall
be skipped and the downgrade shall proceed.

### REQ-UVF-014 (Ubiquitous) — re-exec after install

The `moai update` command shall re-exec the process into the
newly installed binary after a `--version` install succeeds, mirroring
the existing re-exec semantics of `reexecNewBinary`, so that subsequent
template sync (when not suppressed by `--binary`) runs against the new
binary's embedded templates.

### REQ-UVF-015 (Ubiquitous) — no template-mirror touch

The `--version` flag's implementation shall touch no path under
`internal/template/templates/` (this is moai-adk-go's own Go code under
`internal/cli/`, not a distributed template; §25 template-neutrality is
N/A, but the no-touch invariant is enforced at run phase via `git diff`
verification).

### REQ-UVF-016 (Capability gate) — 4-locale docs-site update

**Where** the `--version` flag ships to users, the docs-site pages
`cli-reference/update.md` and `getting-started/installation.md` shall
be updated in all four locales (ko / en / ja / zh) within the same
release, documenting the flag semantics, the mutual-exclusion matrix,
and the stable-vs-rc behavior.

## §D Constraints

- **Tier M**: ≤16 REQ / ≤16 AC. This SPEC carries 16 REQ and 16 AC.
- **TDD**: test-first (RED-GREEN-REFACTOR) per `quality.yaml`
  `constitution.development_mode: tdd`. Reproduction-first for every
  defect path (REQ-UVF-008..011).
- **Security**: checksum verification MANDATORY (CWE-345); allowlist
  confinement MANDATORY (REQ-UVF-005); no `MOAI_UPDATE_URL` escape.
- **Behavior preservation**: default `moai update` (no `--version`)
  MUST be byte-for-byte behavior-identical to today (REQ-UVF-004,
  proved by AC-UVF-004).
- **Docs-site 4-locale**: same-PR 4-locale obligation per
  `hns-oss-docs-i18n-rules`.

## §E Open Questions

None blocking. All six Key Design Questions raised by the orchestrator
are resolved in §F below with rationale; none require user escalation.

## §F Design Decisions (resolves orchestrator's Key Design Questions)

### §F.1 Q1 — Flag-name collision (RECOMMENDATION: `--version`)

`--version` in Go CLIs conventionally means "print version and exit".
`moai` already exposes this at the **root** command (`moai --version`).
`moai update --version <tag>` is a **subcommand** flag, not a root flag,
so cobra's parser disambiguates by command context — there is no parser
collision. The semantic risk is real but bounded: a user typing
`moai update --version` with no argument receives a "flag needs an
argument" error from cobra (not a silent version-print), which already
disambiguates intent.

The SPEC recommends **`--version`** over `--tag` / `--release` /
`moai update switch <tag>`, on these grounds:

1. **User mental model**: `install.sh --version` is the prior art the
   user base already knows; reusing the same flag name carries that
   familiarity into the CLI.
2. **Discoverability**: `--version` is the most-guessed flag for
   "install a specific version" in CLIs that have it (curl, rustup,
   pyenv).
3. **Subcommand form rejected**: `moai update switch <tag>` introduces
   a new subcommand tree, conflicting with `moai update`'s current
   no-subcommand shape and forcing a larger parser change.

Run-phase implementer MAY propose `--tag` as an alias during M1 if
user-testing surfaces confusion; the canonical name stays `--version`.

### §F.2 Q2 — Tag → URL resolution

`<tag>` → `https://api.github.com/repos/modu-ai/moai-adk/releases/tags/<tag>`.
This endpoint is on the allowlist (host = `api.github.com`, scheme =
`https`). v-prefix normalization (REQ-UVF-003) reuses the
`normalizeVersionMajor` discipline: accept both `v3.0.0` and `3.0.0`,
reject ambiguous forms like `go-v3.0.0` (those belong to the dev-branch
path and are addressed by REQ-UVF-012).

### §F.3 Q3 — Flag interaction matrix

| Other flag | `--version` allowed? | Behavior |
|---|---|---|
| `--check` | NO | mutual exclusion (REQ-UVF-007) |
| `--templates-only` | NO | mutual exclusion |
| `--restore` | NO | mutual exclusion |
| `--dry-run` | NO | mutual exclusion |
| `--binary` | YES | install only the binary of the requested tag, skip template sync |
| `--force` | YES | force re-install even when running version already matches |
| `--yes` | YES | skip downgrade confirmation (REQ-UVF-013) |
| `--no-hooks` | YES | orthogonal; passed through |
| `--verbose` | YES | orthogonal; passed through |
| `--profile` | YES | orthogonal; persisted after re-exec |

### §F.4 Q4 — Error handling

Each defect path has its own REQ + AC pair: REQ-UVF-008 (tag 404),
REQ-UVF-009 (no binary asset), REQ-UVF-010 (checksum mismatch),
REQ-UVF-011 (network failure). All four exit non-zero, none leave the
filesystem in a partial state, all name the offending tag in the error.

### §F.5 Q5 — Default-behavior preservation

AC-UVF-004 proves this: a golden-file characterization test asserts
that `moai update` with no `--version` produces byte-identical HTTP
calls and control flow to the pre-SPEC baseline.

### §F.6 Q6 — RC asset reality

Precondition stated in REQ-UVF-009: a tag whose GitHub Release carries
binary assets (GoReleaser default) installs; one without fails clearly
with the asset list surfaced. No SPEC-level guarantee that every rc tag
has assets — that is a release-engineering concern outside this SPEC.

## §G Out of Scope

### Out of Scope — Config/template rollback via `--version`

- Template-state rollback (3-way merge to a prior template set,
  `--restore` semantics for templates). The moai binary `go:embed`s
  templates, so a `--version` install naturally carries that tag's
  templates; explicit template-version selection is the existing update
  flow's job.
- `moai-adk` config downgrade (`.moai/config/*` rewinding to a prior
  shape). Owned by SPEC-UPDATE-DATA-SURVIVAL-001 `--restore`, not here.

### Out of Scope — Checksum bypass

- A `--skip-checksum` flag, an `--insecure` flag, or any other pathway
  that disables the CWE-345 verification for `--version` installs.
  Checksum verification is MANDATORY (REQ-UVF-006).

### Out of Scope — Tag aliasing / semver range specifiers

- `--version latest`, `--version ">=3.0"`, `--version stable`, or any
  non-literal-tag specifier. The flag accepts ONE literal tag only.
  Range resolution is a future capability, not this SPEC.

### Out of Scope — Distribution outside `api.github.com`

- A `--source` / `--mirror` flag for installing from a non-GitHub
  source. The allowlist (REQ-UVF-005) is MANDATORY; broadening it is a
  separate security SPEC (cf. SPEC-SEC-HARDEN-005).

### Out of Scope — `install.sh` removal

- Retiring the external `install.sh --version` script. The script
  remains the bootstrapping path (used before `moai` itself is
  installed); this SPEC only adds the in-process CLI path.

### Out of Scope — Build-metadata / semver-plus tags

- Accepting build-metadata tags such as `v3.0.0+build.5` (the shape
  `v2_detection.go:247` names). The tag-charset validation
  (`^v?[0-9A-Za-z.\-]+$`) intentionally excludes the `+` separator,
  because accepting build-metadata invites ambiguity (two tags that
  differ only in `+build.N` resolve the same release asset). Tags
  carrying `+` are rejected at validation; broadening the charset to
  admit them is a separate SPEC.

## §H HISTORY

- 2026-08-03 — v1.0.0 — initial draft (plan-phase), author: manager-spec.
  Scope locked by orchestrator: binary-only; all release tags (stable +
  rc + previous versions). Related prior SPECs referenced, not
  duplicated.
