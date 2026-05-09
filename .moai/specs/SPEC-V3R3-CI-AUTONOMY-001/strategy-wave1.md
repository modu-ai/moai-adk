# Wave 1 Execution Strategy (Phase 1 Output)

> Audit trail. manager-strategy output for Wave 1 of SPEC-V3R3-CI-AUTONOMY-001.
> Generated: 2026-05-05. Methodology: TDD. Mode: Standard.

## Architecture Decisions

### Module Topology

`scripts/ci-mirror/` directory layout (mirrored under `internal/template/templates/`):

```
scripts/ci-mirror/
├── run.sh                    # Entry: language detection + dispatch + progress orchestration
├── cross-compile.sh          # Cross-OS/ARCH go build matrix
├── lib/
│   ├── _common.sh           # Helpers: log_step, abort, posix_now, detect_languages
│   ├── go.sh                # Full Go pipeline (vet + lint + test -race + cross-compile)
│   ├── python.sh, node.sh, rust.sh, java.sh, kotlin.sh, csharp.sh, ruby.sh,
│   ├── php.sh, elixir.sh, cpp.sh, scala.sh, r.sh, flutter.sh, swift.sh
│   └── (smoke modules; skip if toolchain missing)
└── README.md
```

### Why Shell, Not Go

- Pre-push runs BEFORE binary build (binary may itself be broken)
- Zero-build-time: shell ships in `.git_hooks/`, runnable on fresh clone
- Per-language extension without recompiling moai-adk
- Go binary remains installer/orchestrator role

### SSoT Pattern

`.github/required-checks.yml`:
```yaml
version: 1
branches:
  main:
    contexts: [Lint, Test (ubuntu-latest), Test (macos-latest), Test (windows-latest), Build (linux/amd64), CodeQL]
  release/*:
    contexts: [Lint, Test (ubuntu-latest), Test (macos-latest)]
```

Consumed by `internal/config/required_checks.go::LoadRequiredChecks()` → `internal/cli/branch_protection.go::RenderBranchProtectionJSON()` → `gh api -X PUT`.

### Cross-Platform Constraints

- POSIX sh shebang `#!/bin/sh`, no bash-isms (no `[[ ]]`, no `${var//}`, no `local` outside subshells, no `set -o pipefail`)
- `printf '%s\n'` instead of `echo`
- `command -v` instead of `which`
- Date: `date +%s` only (epoch seconds)
- BSD/GNU `sed -i` divergence avoided (no in-place edits in hook)
- Quote every `"$VAR"`
- Exit codes: 0 success, 1 pre-flight, 2 lint/test, 3 build, 99 internal

### 16-Language Detection Algorithm

Priority chain (mirrors `internal/config/language_markers.go`):
1. go.mod → go
2. package.json → node
3. pyproject.toml/setup.py → python
4. Cargo.toml → rust
5. pom.xml/build.gradle/build.gradle.kts → java/kotlin (gradle.kts → kotlin)
6. *.csproj/*.sln → csharp
7. Gemfile → ruby
8. composer.json → php
9. mix.exs → elixir
10. CMakeLists.txt → cpp (heuristic with .cpp source presence)
11. build.sbt → scala
12. DESCRIPTION → r
13. pubspec.yaml → flutter
14. Package.swift → swift

Multi-language: dispatch sequentially, fail-fast on first non-zero. No marker → silent skip exit 0.

### Branch Protection Template

Stored as Go template at `internal/template/templates/.github/branch-protection.json.tmpl`:
```
{
  "required_status_checks": {
    "strict": true,
    "contexts": [{{range $i, $c := .Contexts}}{{if $i}}, {{end}}{{printf "%q" $c}}{{end}}]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {...},
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_conversation_resolution": true
}
```

## TDD Breakdown Per Task

(See tasks.md for the table; this section captures RED/GREEN/REFACTOR detail.)

### W1-T07 — Required Checks SSoT

- RED: `TestLoadRequiredChecks` table cases for main + release/*; `TestNoHardcodedContexts` grep assertion (only loader file matches `Test (ubuntu-latest)` literal in internal/cli|internal/cmd).
- GREEN: `LoadRequiredChecks(projectRoot string) (RequiredChecks, error)` via yaml.v3.
- REFACTOR: extract `MatchBranch(name string) []string` for release/* glob if simple; otherwise keep inline.

### W1-T06 — Branch Protection Wiring

- RED:
  - `TestRenderBranchProtectionJSON` (jsondiff against fixture)
  - `TestApplyBranchProtection_GhMissing` (LookPath fail → exact remediation message)
  - `TestApplyBranchProtection_AuthFailure` (fake gh in tempdir PATH exits 4)
- GREEN: `RenderBranchProtectionJSON(checks, branch)`, `ApplyBranchProtection(ctx, owner, repo, branch)` with `ghClient` interface for test injection.
- REFACTOR: extract gh invocation into `ghClient` interface for fake/real swap.

### W1-T04 — Makefile

- RED: `Makefile.ci-local.test.sh` runs `make -n ci-local`, asserts presence of golangci-lint, vet, test -race, cross-compile invocation strings.
- GREEN: add ci-local + pr-merge targets. ci-local streams progress via `printf '[ci-local] step N/4: ...' >&2`. pr-merge reads `PR=N STRATEGY=squash|merge` env vars.
- REFACTOR: extract cross-compile matrix to `scripts/ci-mirror/cross-compile.sh`.

### W1-T01 — run.sh + _common.sh

- RED: `run_test.sh` 3 cases (no marker / go marker dispatches / dispatched module exit code propagates).
- GREEN: implement run.sh with `detect_languages` loop + `MOAI_CI_LIB_DIR` override. _common.sh provides `log_step`, `safe_mktemp`, `posix_now`.
- REFACTOR: timestamp prefix in log_step.

### W1-T02 — lib/go.sh

- RED: `go_test.sh` 3 cases (passing project / failing test / golangci-lint not on PATH graceful skip).
- GREEN: vet → lint → test -race → cross-compile build matrix (linux/amd64, darwin/amd64, darwin/arm64, windows/amd64). All gated on `command -v` toolchain availability.
- REFACTOR: extract `cross_compile_matrix()` function.

### W1-T03 — 14 Languages

- RED: structural conformance loop (every lib/*.sh has shebang + `command -v` + silent skip on no toolchain). Behavioral spot-check on python/node/rust/swift.
- GREEN: each module 5-15 LOC following the template:
  ```sh
  #!/bin/sh
  set -eu
  if ! command -v <tool> >/dev/null 2>&1; then
    printf '[ci-mirror][<lang>] toolchain not installed — skipping\n' >&2
    exit 0
  fi
  <tool> <args> || exit 2
  ```
- REFACTOR: audit duplication; resist premature `_lang_runner.sh` extraction unless duplication is severe.

### W1-T05 — pre-push Hook + Installer

- RED:
  - `TestInstallPrePushHook` — installs hook, asserts file at `.git/hooks/pre-push` is executable + contains `scripts/ci-mirror/run.sh` literal
  - `TestInstallHooks_NoHooksFlag` — noHooks=true skips
  - `TestInstallHooks_PreservesUserHook` — pre-existing user hook (no MoAI marker) → not overwritten
  - `prepush_e2e_test.sh` — git init tempdir, install, commit, push --dry-run; assert hook ran + result
- GREEN: hook script with SKIP_MOAI_PREPUSH env var override. installer copies embedded template → 0755 mode, idempotent (checks first 3 lines for `# MoAI-ADK pre-push hook` marker).
- REFACTOR: extract hook content into a constant referenced by installer + test.

## Risk Mitigations

R1 (POSIX): _common.sh wrappers; shellcheck -s sh in CI.
R2 (Windows git-bash): no path translation issues (POSIX form via `git rev-parse --show-toplevel`); `make` availability checked by `moai doctor`.
R3 (gh auth): error message includes full rendered JSON for copy-paste; sentinel `ErrGhUnavailable`.
R4 (16-lang fixture): spot-check 4, structural conformance for the rest.
R5 (installer idempotency): MoAI marker check; `--no-hooks` flag.
R6 (cross-compile speed): Wave 1 always cross-compiles; speed optimization deferred to REFACTOR.
R7 (--no-verify self-log): scoped honestly to invocation log; full bypass detection deferred to Wave 2.

## Implementation Order Rationale

1. **W1-T07** (SSoT) first — smallest, no deps, blocks W1-T06; validates yaml.v3 round-trip
2. **W1-T06** (branch protection) — depends on W1-T07; resolves Confirmer interface architectural subtlety early
3. **W1-T04** (Makefile) — needed by W1-T01, W1-T05
4. **W1-T01** (run.sh + _common.sh) — establishes shell helper conventions
5. **W1-T02** (lib/go.sh) — heaviest module; validates cross-compile script
6. **W1-T03** (14 languages) — repetitive once go.sh pattern set; batch in commits of 4-5
7. **W1-T05** (pre-push hook + installer) — last; e2e test exercises entire stack

## Commit Pacing (Phase 3 manager-git target)

- Commit 1 (W1-T07): "feat(ci-mirror): add required-checks SSoT + loader"
- Commit 2 (W1-T06): "feat(github): wire branch protection from SSoT with graceful auth failure"
- Commit 3 (W1-T04): "feat(make): add ci-local target with progress streaming"
- Commit 4 (W1-T01): "feat(ci-mirror): add run.sh dispatcher with 16-lang detection"
- Commit 5 (W1-T02): "feat(ci-mirror): add Go pipeline (vet/lint/test/cross-compile)"
- Commit 6 (W1-T03): "feat(ci-mirror): add 14 lightweight language modules"
- Commit 7 (W1-T05): "feat(hooks): add POSIX pre-push hook + installer with bypass log"
- Commit 8: "docs: update CHANGELOG.md for SPEC-V3R3-CI-AUTONOMY-001 Wave 1"

Each commit individually green per `make ci-local` (bootstrap exception: commits 1-3 use simpler local check; commits 5+ run real `make ci-local` against themselves).
