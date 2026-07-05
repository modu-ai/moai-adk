# SPEC-PRECOMMIT-001 — Implementation Plan

## §A Context

Symmetric extension of the pre-push installer (`internal/cli/hook_install.go`). The
run phase mirrors the existing `PrePushInstaller` machinery for the commit tier and
adds a fast-subset `#!/bin/sh` pre-commit hook. Tier: **S** (mechanical mirror of a
proven pattern; < 300 LOC; low risk; no new abstractions).

**Tier note (D7):** this SPEC sits at the UPPER Tier-S file-count boundary — the run
phase touches ~5-6 files (template + `preCommitHookContent` const + byte-identity test +
installer tests + `init.go` + `update.go`) and uses the separate-`acceptance.md` layout
rather than the lean 2-file AC-inline form. Tier S is retained as the tie-breaker on the
mechanical-mirror / <300 LOC / no-new-abstraction axis, NOT on file count alone; the file
count nudges toward the S/M boundary but the risk profile and absence of new abstractions
resolve to S (per the "default to the simpler mode" tie-breaker in spec-workflow.md).

**PRESERVE (do NOT modify):**
- `prePushHookContent` constant + `PrePushInstaller` + `InstallPrePushHook` +
  `installPrePushHookOptional` + `fileHasMoaiMarker` (the shared helper is REUSED, not changed).
- `internal/template/templates/.git_hooks/pre-push` + `TestPrePushTemplateMatchesConstant`.
- The `ErrUserHookExists` sentinel (REUSED verbatim by the pre-commit installer).
- All config layers (`defaults.go` / `types.go` / `validation.go` / `git-strategy.yaml`) — REQ-PC-017 is config-agnostic.

**EXTEND / ADD:**
- New template file `internal/template/templates/.git_hooks/pre-commit`.
- New `preCommitHookContent` constant + `PreCommitInstaller` type +
  `NewPreCommitInstaller` + `InstallPreCommitHook` + `installPreCommitHookOptional`
  in `internal/cli/hook_install.go` (or a sibling `hook_install_precommit.go`).
- New byte-identity test (mirror of `TestPrePushTemplateMatchesConstant`) +
  installer unit tests (mirror of the pre-push installer tests).
- One-line wiring in `init.go` (after the pre-push call, ~L473) and `update.go` (~L886).

## §B Known Issues (Tier S — filtered to relevant categories)

- **B1 Cross-platform build (REQ-PC-019):** installer uses only portable primitives
  (`os`/`bufio`/`filepath`/`strings`/`errors`/`io`) — the same set `PrePushInstaller`
  uses. No syscall, so no `//go:build` split needed. Verify `GOOS=windows GOARCH=amd64 go build ./...`.
- **B6 spec-lint Out of Scope heading:** spec.md uses `### Out of Scope — <topic>`
  H3 sub-headings under an `## Exclusions` H2 (satisfies `OutOfScopeRule`). Do not
  collapse to a bare H2.
- **B8 Working-tree hygiene:** the working tree carries large uncommitted changes from
  parallel sessions and origin/main is ahead. Commit ONLY the pre-commit files with a
  path-limited `git add` (template, hook_install code, tests, init.go, update.go). Do
  NOT `git add -A`; do NOT touch anything outside this SPEC's scope.
- **B10 Template-First order (REQ-PC-016):** edit `internal/template/templates/.git_hooks/pre-commit`
  FIRST, then `make build`, then confirm the byte-identity test passes. The constant
  and the template MUST be authored to be byte-identical (same trailing newline, same
  indentation, same content).

## §C Pre-flight

```bash
git rev-parse HEAD                       # baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...   # cross-platform baseline
go test ./internal/cli/ -run 'PrePush|InstallPrePush|TemplateMatchesConstant' -count=1  # pre-push suite green pre-change
ls internal/template/templates/.git_hooks/                   # confirm pre-push exists, pre-commit absent
```

## §D Constraints (DO NOT VIOLATE)

- Reuse `ErrUserHookExists`; do not introduce a second sentinel.
- Mirror the marker-in-first-3-lines logic exactly (may reuse `fileHasMoaiMarker` with
  a marker-string parameter, or add a `fileHasMarker(path, marker)` generalization —
  implementer's choice, but keep the pre-push path behavior byte-identical).
- Template neutrality: NO internal SPEC IDs / REQ tokens / dates / SHAs / macOS paths
  in the distributed `pre-commit` template (CLAUDE.local.md §15/§25).
- `make build` after template edit; byte-identity MUST hold.
- Path-limited commits only (B8). No `--no-verify`, no `--amend`, no force-push.
- Conventional Commits: `feat(SPEC-PRECOMMIT-001): ...`.

## §E Self-Verification (run-phase deliverables)

manager-develop MUST report:
- **E1** AC PASS/FAIL matrix for AC-PC-001 .. AC-PC-015 with the verification command + actual output per row.
- **E2** `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- **E3** Coverage for `internal/cli` (new installer paths covered ≥ 85% for the added functions).
- **E5** `golangci-lint run` — NEW issues vs baseline distinguished.
- **E6** New commit SHA(s) + push state.

## §F Milestones (priority-ordered, no time estimates)

- **M1 — Template + constant (byte-identity core).**
  1. Author `internal/template/templates/.git_hooks/pre-commit` (fast-subset POSIX sh).
  2. Add `preCommitHookContent` constant to `hook_install.go`, byte-identical to the template.
  3. `make build` to regenerate the embedded FS.
  4. Add the byte-identity test (mirror of `TestPrePushTemplateMatchesConstant`) → passes.

  Illustrative hook sketch (final byte content authored at run phase; must be neutral POSIX sh):

  ```sh
  #!/bin/sh
  # MoAI-ADK pre-commit hook — fast subset (gofmt -l + go vet on staged Go files)
  # Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit
  # Push-tier full CI (make ci-local) is the pre-push hook's job — NOT duplicated here.
  set -eu

  if [ "${SKIP_MOAI_PRECOMMIT:-0}" = "1" ]; then
      printf '[pre-commit] SKIP_MOAI_PRECOMMIT=1 -- bypass requested\n' >&2
      exit 0
  fi

  # Staged Go files (Added/Copied/Modified), NUL-safe iteration recommended at impl time.
  STAGED_GO="$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)"
  [ -z "$STAGED_GO" ] && exit 0        # no Go staged -> fast no-op

  # gofmt format check (skip gracefully if gofmt absent -> non-Go env)
  if command -v gofmt >/dev/null 2>&1; then
      NEED_FMT="$(printf '%s\n' "$STAGED_GO" | xargs gofmt -l 2>/dev/null || true)"
      if [ -n "$NEED_FMT" ]; then
          printf '\n[pre-commit] FAILED: the following staged files need gofmt:\n%s\n' "$NEED_FMT" >&2
          printf '[pre-commit] Hint: gofmt -w <files> && git add <files>\n' >&2
          printf '[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit\n' >&2
          exit 1
      fi
  fi

  # go vet on affected packages (skip gracefully if go absent)
  if command -v go >/dev/null 2>&1; then
      PKGS="$(printf '%s\n' "$STAGED_GO" | xargs -n1 dirname | sort -u | sed 's#^#./#')"
      if ! go vet $PKGS >/dev/null 2>&1; then
          printf '\n[pre-commit] FAILED: go vet reported issues in staged packages.\n' >&2
          printf '[pre-commit] Hint: go vet ./...   (fix, then re-commit)\n' >&2
          printf '[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit\n' >&2
          exit 1
      fi
  fi

  exit 0
  ```

- **M2 — Installer type + optional wrapper.**
  1. Add `PreCommitInstaller` + `NewPreCommitInstaller` + `InstallPreCommitHook`
     (marker-based create/overwrite/preserve, `0755`, reuse `ErrUserHookExists`).
  2. Add `installPreCommitHookOptional` (non-fatal, mirror of the pre-push optional wrapper).
  3. Add installer unit tests: fresh-repo create (AC-PC-001), foreign-hook preserve
     (AC-PC-002), MoAI-marker overwrite (AC-PC-003), `--no-hooks` skip (AC-PC-010),
     install assertion — POSIX shebang + marker + 4 tokens (AC-PC-012), non-fatal
     install failure on unwritable `.git/hooks/` (AC-PC-014), config-agnostic install
     across pre_commit=skip/warn/enforce (AC-PC-015).

- **M3 — Wiring + hook-behavior tests + verification.**
  1. Wire `installPreCommitHookOptional` into `init.go` (after the pre-push call) and
     `update.go` (after the pre-push call), gated by the same `no-hooks` flag.
  2. Add hook-behavior tests (script-level, using a temp git repo): staged un-gofmt'd
     file blocks (AC-PC-005), `SKIP_MOAI_PRECOMMIT=1` bypass (AC-PC-006), no staged Go
     no-op (AC-PC-008), `go vet` failure blocks (AC-PC-009), no `make ci-local` token
     present (AC-PC-011), toolchain-absent (`go`/`gofmt` off PATH) graceful exit 0 with
     staged Go files (AC-PC-013).
  3. Full verification batch: `go test ./...`, cross-platform build (AC-PC-007),
     byte-identity test (AC-PC-004), lint.

## §G Anti-Patterns

- Re-architecting `fileHasMoaiMarker` in a way that changes pre-push behavior.
- Reading the `pre_commit` config value at install time to gate installation — the installer is config-agnostic (REQ-PC-017); the deferred follow-up is a hook-run-time severity dial mirroring `resolvePrePushAction`, NOT an install gate.
- Adding `make ci-local` / golangci-lint / full test to the commit hook (violates the 2-tier design, REQ-PC-014).
- Editing the Go constant without editing the template (or vice versa) — byte-identity test will fail.
- `git add -A` sweeping unrelated parallel-session changes into the commit.

## §H Cross-References

- spec.md § Cross-References (full list).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier S minimal-form delegation).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier S = < 300 LOC, < 5 files).
