# Acceptance — SPEC-UPDATE-VERSION-FLAG-001

> Verification layer. Given-When-Then scenarios (binary-testable). Each
> AC-UVF-* binds to its corresponding REQ-UVF-* in spec.md. Severity,
> traceability, and closure gates in §D.1..§D.7.

## §D Acceptance Criteria (Given-When-Then)

### AC-UVF-001 — flag registered
**Given** the `moai update` cobra command tree,
**When** `moai update --help` is invoked,
**Then** the output lists `--version string` with a description naming
"release tag (stable / rc / previous version)".

### AC-UVF-002 — install specific tag (stable, rc, previous)
**Given** an `httptest.Server` mock standing in for the GitHub API,
capturing every outbound request, and serving a release for tag
`v3.0.0`, `v3.1.0-rc1`, and `v2.14.0`,
**When** `moai update --version <tag>` is invoked for each,
**Then** the binary is replaced by the asset attached to that tag,
verified against the tag's published checksum, and the command exits 0;
AND for the `v3.0.0` case the captured outbound request satisfies
`req.Host == "api.github.com"` AND
`req.URL.Path == "/repos/modu-ai/moai-adk/releases/tags/v3.0.0"`,
directly falsifying REQ-UVF-005's host-confinement claim on the
constructed tag-resolution URL (not only on the env-var rejection
path covered by AC-UVF-005).

### AC-UVF-003 — v-prefix normalization
**Given** a user invocation `moai update --version 3.0.0` (no `v`),
**When** tag resolution runs,
**Then** the HTTP call targets
`/repos/modu-ai/moai-adk/releases/tags/v3.0.0` (with `v`), and an
equivalent `moai update --version v3.0.0` invocation hits the same
URL.

### AC-UVF-004 — default behavior preserved (regression anchor)
**Given** a recorded golden trace of `moai update` (no `--version`)
HTTP calls and control flow captured before this SPEC,
**When** `moai update` is invoked post-SPEC with no `--version`,
**Then** the HTTP calls, branch decisions, and exit code are
byte-identical to the golden trace (this AC MUST land before any
`--version` branch is added, making the regression anchor
unambiguous).

### AC-UVF-005 — allowlist confinement
**Given** `--version v3.0.0` and a `MOAI_UPDATE_URL` env var pointing
at `https://evil.example.com/`,
**When** the update subsystem starts,
**Then** the env var is rejected fail-closed by the existing
`validateUpdateURL` before any `--version`-related HTTP call, and the
command exits non-zero naming the allowlist violation.

### AC-UVF-006 — checksum verification mandatory, no bypass
**Given** a downloaded binary whose computed checksum does NOT match
the tag's published checksum,
**When** `moai update --version <tag>` runs to completion of the
verification step,
**Then** the binary is NOT installed, the downloaded bytes are
discarded, and `moai update --help` shows no `--skip-checksum` /
`--insecure` flag.

### AC-UVF-007 — mutual-exclusion matrix
**Given** `--version v3.0.0` combined with `--check`, `--templates-only`,
`--restore`, or `--dry-run`,
**When** the command is invoked,
**Then** each combination exits non-zero with a usage error naming the
conflicting pair, before any network call; AND given `--version` +
`--binary`, `--version` + `--force`, `--version` + `--yes`, the
command proceeds (no conflict).

> Note: the "orthogonal; passed through" flags (`--no-hooks`,
> `--verbose`, `--profile`) are transitively covered by AC-UVF-004
> (default-behavior preservation) — they carry no `--version`-specific
> interaction, so they need not be enumerated here.

### AC-UVF-008 — tag not found (HTTP 404)
**Given** a mock GitHub API returning HTTP 404 for
`/releases/tags/v9.9.9-nope`,
**When** `moai update --version v9.9.9-nope` is invoked,
**Then** the command exits non-zero with a structured error naming the
tag and directing the user to `moai update --check` or the releases
page, with no download attempted.

### AC-UVF-009 — release has no binary asset
**Given** a mock GitHub release for tag `v3.1.0-rc2` whose `assets[]`
contains no archive matching the running `GOOS/GOARCH`,
**When** `moai update --version v3.1.0-rc2` is invoked,
**Then** the command exits non-zero naming the tag, the detected
platform (`<GOOS>/<GOARCH>`), and the assets that WERE present, and
the filesystem is untouched.

### AC-UVF-010 — checksum mismatch
**Given** a mock binary download whose SHA256 does NOT match the
tag's `checksums.txt` entry,
**When** `moai update --version <tag>` runs,
**Then** the downloaded bytes are discarded, the running binary path
is unchanged (byte-identical to pre-invocation), and the command
exits non-zero with a CWE-345-referencing error naming the tag.

### AC-UVF-011 — network failure (resolution and download phases)
**Given** a mock network failure during (a) tag resolution and (b)
binary download,
**When** `moai update --version <tag>` runs in each phase,
**Then** the command exits non-zero with a wrapped error that names
the phase (`resolution` vs `download`) and the underlying network
error, and no partial install is left on disk.

### AC-UVF-012 — dev/RC branch interaction
**Given** a running binary whose version reports as a `go-v`-prefixed
dev build,
**When** `moai update --version v3.0.0` is invoked,
**Then** the explicit tag `v3.0.0` is resolved directly via
`/releases/tags/v3.0.0` (NOT the dev-branch `/releases` list
endpoint); AND given the same dev build with default `moai update`
(no `--version`), the dev-branch `/releases` endpoint is still used
(no regression).

### AC-UVF-013 — downgrade confirmation
**Given** a running binary at `v3.2.0` and an invocation
`moai update --version v3.0.0` on an interactive TTY,
**When** the downgrade is detected,
**Then** the user is prompted to confirm; AND given `--yes` OR a
non-TTY stdin, the prompt is skipped and the downgrade proceeds.

### AC-UVF-014 — re-exec after install
**Given** a successful `moai update --version v3.0.0` install (no
`--binary`),
**When** the install step completes,
**Then** the process re-execs into the newly installed binary (via
`reexecNewBinary` equivalent), setting `MOAI_SKIP_BINARY_UPDATE=1`,
so subsequent template sync runs against the new binary's embedded
templates.

### AC-UVF-015 — no template-mirror touch
**Given** the merged PR diff for this SPEC,
**When** `git diff --name-only origin/main...HEAD` is run,
**Then** no path under `internal/template/templates/` appears in the
diff (REQ-UVF-015).

### AC-UVF-016 — 4-locale docs-site update
**Given** the merged PR for this SPEC,
**When** the docs-site pages are inspected,
**Then** `cli-reference/update.md` and `getting-started/installation.md`
exist and are updated in all four locales (ko, en, ja, zh), and the
`hns-oss-docs-verify` exit-gate recipe passes without warning.

## §D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-UVF-004 | BLOCKER | regression anchor; its failure invalidates the entire SPEC |
| AC-UVF-005 | BLOCKER | allowlist escape = RCE path |
| AC-UVF-006 | BLOCKER | checksum bypass = RCE path |
| AC-UVF-010 | BLOCKER | defect-path, security-critical |
| AC-UVF-002 | BLOCKER | core happy path |
| AC-UVF-007 | BLOCKER | UX safety; bad combos must fail fast |
| AC-UVF-008 | HIGH | defect-path |
| AC-UVF-009 | HIGH | defect-path, rc reality |
| AC-UVF-011 | HIGH | defect-path |
| AC-UVF-012 | HIGH | dev-build regression risk |
| AC-UVF-001 | MEDIUM | discoverability |
| AC-UVF-003 | MEDIUM | normalization polish |
| AC-UVF-013 | MEDIUM | UX safety on downgrade |
| AC-UVF-014 | MEDIUM | parity with existing re-exec |
| AC-UVF-015 | MEDIUM | scope hygiene |
| AC-UVF-016 | MEDIUM | docs obligation |

## §D.2 Traceability Matrix

| REQ | Covered by AC |
|-----|---------------|
| REQ-UVF-001 | AC-UVF-001 |
| REQ-UVF-002 | AC-UVF-002 |
| REQ-UVF-003 | AC-UVF-003 |
| REQ-UVF-004 | AC-UVF-004 |
| REQ-UVF-005 | AC-UVF-005 |
| REQ-UVF-006 | AC-UVF-006 |
| REQ-UVF-007 | AC-UVF-007 |
| REQ-UVF-008 | AC-UVF-008 |
| REQ-UVF-009 | AC-UVF-009 |
| REQ-UVF-010 | AC-UVF-010 |
| REQ-UVF-011 | AC-UVF-011 |
| REQ-UVF-012 | AC-UVF-012 |
| REQ-UVF-013 | AC-UVF-013 |
| REQ-UVF-014 | AC-UVF-014 |
| REQ-UVF-015 | AC-UVF-015 |
| REQ-UVF-016 | AC-UVF-016 |

Every REQ has exactly one AC; coverage is 16/16 = 100%.

## §D.3 Indirect Verification

The following invariants are indirectly verified by the AC set above
(no separate AC needed):

- **Allowlist host confinement in the success path**: AC-UVF-002
  directly asserts the outbound Host (`api.github.com`) and Path
  (`/repos/modu-ai/moai-adk/releases/tags/v3.0.0`) on the constructed
  tag-resolution URL; AC-UVF-005 covers the env-var rejection path.
  Together the two ACs falsify REQ-UVF-005 on both the success path
  and the env-var-escape path. (The prior reliance on AC-UVF-002's
  mock host being `127.0.0.1` under-verified host confinement — the
  strengthened AC-UVF-002 now asserts Host/Path directly.)
- **No partial install on any defect path**: AC-UVF-008, AC-UVF-009,
  AC-UVF-010, AC-UVF-011 each assert filesystem-untouched as a
  post-condition.
- **Default-behavior preservation under all flag additions**:
  AC-UVF-004 covers the no-`--version` baseline; the mutual-exclusion
  matrix (AC-UVF-007) ensures the new flag cannot silently combine
  with existing flags to alter the default path.

## §D.4 Closure Gates (Definition of Done)

- [ ] All 16 AC-UVF-* PASS with observed evidence (test output cited).
- [ ] All BLOCKER-severity AC PASS.
- [ ] `go test ./internal/cli/...` exits 0 with the new tests.
- [ ] `go test ./...` exits 0 (no regressions outside `internal/cli`).
- [ ] `golangci-lint run` exits 0.
- [ ] `git diff --name-only origin/main...HEAD` shows no
  `internal/template/templates/` path (AC-UVF-015).
- [ ] `hns-oss-docs-verify` exit-gate recipe passes (AC-UVF-016).
- [ ] PR merged with ≥4 CI checks green.

## §D.5 Forward-Looking Checks (Deferred to Future SPECs)

- **Range specifiers** (`--version ">=3.0"`, `--version latest`):
  explicitly out of scope (spec.md §G). Future SPEC may revisit.
- **Non-GitHub mirror** (`--source`): explicitly out of scope (spec.md
  §G). Owned by SPEC-SEC-HARDEN-005 successor.
- **`install.sh` retirement`: out of scope (spec.md §G). Script remains
  the bootstrap path.

## §D.6 Test Strategy Notes

- **TDD**: every BLOCKER AC gets a RED test first (reproduction for
  defect paths, characterization for AC-UVF-004), then GREEN, then
  REFACTOR.
- **Mock GitHub API**: tests use an `httptest.Server` stub for
  `/releases/tags/<tag>` responses, including 404, empty-assets, and
  checksum-mismatch fixtures.
- **Re-exec verification**: AC-UVF-014 uses a subprocess test pattern
  (the test binary re-execs itself and asserts `MOAI_SKIP_BINARY_UPDATE`
  is set in the child).
- **Downgrade prompt**: AC-UVF-013 uses a pseudo-TTY stub for the
  interactive case and a pipe for the non-TTY case.

## §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: ≥85% coverage on touched files (`internal/cli/update.go`,
  `internal/cli/deps.go`); characterization test (AC-UVF-004) anchors
  behavior preservation.
- **Readable**: error messages name the offending tag, phase, and
  remediation; no raw hex dumps in user-facing output.
- **Unified**: `gofmt` + `golangci-lint` clean; flag-help text follows
  the existing `moai update` flag description style.
- **Secured**: checksum MANDATORY (AC-UVF-006, AC-UVF-010); allowlist
  MANDATORY (AC-UVF-005); CWE-345 referenced in the checksum-mismatch
  error.
- **Trackable**: Conventional Commits (`feat(update): ...`); SPEC-ID
  referenced in the PR body.
