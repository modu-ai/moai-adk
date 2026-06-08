# Acceptance Criteria — SPEC-PREPUSH-WIRING-001

**AC count**: 9 (AC-PPW-001 .. AC-PPW-009)

All grep idioms reference EXISTING file paths and EXISTING test names verified against the
live tree. Comment lines are excluded where presence-of-invocation is asserted. Byte-parity
is verified via the canonical test, not a hand grep diff.

---

## AC-PPW-001 — Convention invocation present in the distributed template (REQ-PPW-001)

**Given** the distributed pre-push template,
**When** searching for the `moai hook pre-push` invocation excluding comment lines,
**Then** the invocation is present.

```bash
grep -n 'moai hook pre-push' internal/template/templates/.git_hooks/pre-push \
  | grep -v '^[0-9]*:[[:space:]]*#'
# Expected: at least one non-comment match (the invocation line).
```

**Given** the `make ci-local` step,
**Then** it is retained unchanged in the template.

```bash
grep -n 'make -C "\$REPO_ROOT" -s ci-local' internal/template/templates/.git_hooks/pre-push
# Expected: 1 match (existing ci-local step preserved).
```

---

## AC-PPW-002 — Convention block placed after ci-local fail-check, before final exit (REQ-PPW-002 / D.1)

**Given** the template script,
**When** comparing the line number of the `moai hook pre-push` invocation against the final
`exit 0`,
**Then** the invocation appears BEFORE the final `exit 0` (not dead code after it).

```bash
# Line of the moai invocation (non-comment):
inv=$(grep -n 'moai hook pre-push' internal/template/templates/.git_hooks/pre-push \
        | grep -v '^[0-9]*:[[:space:]]*#' | head -1 | cut -d: -f1)
# Line of the LAST 'exit 0' (final exit):
fin=$(grep -n '^exit 0' internal/template/templates/.git_hooks/pre-push | tail -1 | cut -d: -f1)
[ -n "$inv" ] && [ -n "$fin" ] && [ "$inv" -lt "$fin" ] && echo "PASS: invocation($inv) before final exit($fin)"
# Expected: PASS line printed.
```

---

## AC-PPW-003 — moai-presence guard wraps the convention block (REQ-PPW-003)

**Given** the convention block,
**When** searching for the `command -v moai` guard,
**Then** the guard is present so non-moai projects skip the block.

```bash
grep -n 'command -v moai' internal/template/templates/.git_hooks/pre-push
# Expected: 1 match (the presence guard).
```

---

## AC-PPW-004 — Default no-op behavior preserved (REQ-PPW-004)

**Given** the shipped template git-convention config,
**When** reading `enforce_on_push`,
**Then** it remains `false` (engine self-gates → wired hook is a no-op by default).

```bash
grep -n 'enforce_on_push: false' internal/template/templates/.moai/config/sections/git-convention.yaml
# Expected: 1 match (default unchanged — out-of-scope to flip).
```

**Given** the Go engine,
**Then** the self-gate `isEnforceOnPushEnabled()` is unchanged (no Go logic modified).

```bash
grep -n 'func isEnforceOnPushEnabled' internal/cli/hook_pre_push.go
# Expected: 1 match at the original definition (engine untouched).
```

---

## AC-PPW-005 — git stdin → commit-message translation present (REQ-PPW-005)

**Given** the convention block,
**When** searching for the git-log subject extraction used to translate ref ranges into
commit messages,
**Then** a `git log --format=%s` (or `--pretty=format:%s`) translation is present.

```bash
grep -nE "git log .*(--format=%s|--pretty=format:%s|--format='%s')" \
  internal/template/templates/.git_hooks/pre-push
# Expected: at least one match (range → subject-lines translation).
```

**Given** the convention block,
**Then** a `while read` ref-update loop reading the four git-supplied fields is present.

```bash
grep -nE 'while[[:space:]]+read[[:space:]]+(-r[[:space:]]+)?[a-z_]+[[:space:]]+[a-z_]+[[:space:]]+[a-z_]+[[:space:]]+[a-z_]+' \
  internal/template/templates/.git_hooks/pre-push
# Expected: 1 match (the `while read -r local_ref local_oid remote_ref remote_oid` loop).
# The optional `(-r[[:space:]]+)?` group tolerates the shellcheck-recommended `read -r` flag;
# without it the regex false-negatives on the correct `read -r ...` idiom (found at run-phase
# independent verification — the implementation correctly uses `read -r`).
```

---

## AC-PPW-006 — New-branch zero-SHA edge case handled (REQ-PPW-006)

**Given** the translation loop,
**When** a ref-update line reports an all-zero remote SHA (new remote ref),
**Then** the block uses a locally-new-commit enumeration rather than a `<zero>..<local>` diff.

```bash
# The block must reference the all-zero SHA sentinel AND a not-remotes / new-commit form.
grep -nE '0{7,}' internal/template/templates/.git_hooks/pre-push
# Expected: at least one match (zero-SHA sentinel compared per ref).
grep -nE '(--not[[:space:]]+--remotes|--remotes)' internal/template/templates/.git_hooks/pre-push
# Expected: at least one match (new-branch range enumerates commits not on remotes).
```

> Run-phase note: the exact range idiom (e.g. `git log --format=%s "$local_oid" --not --remotes`)
> is chosen during GREEN; this AC asserts the zero-SHA sentinel handling + new-commit
> enumeration are both present, not a specific verbatim string.

---

## AC-PPW-007 — Branch-delete zero-local-SHA skipped (REQ-PPW-007)

**Given** the translation loop,
**When** a ref-update line reports an all-zero local SHA (branch deletion),
**Then** the loop skips that ref (a `continue` guarded by the zero-local-SHA test).

```bash
# The delete-skip path must reference the all-zero local-SHA sentinel AND a continue:
grep -nE '0{7,}' internal/template/templates/.git_hooks/pre-push
# Expected: >=1 match (zero-SHA sentinel — shared with AC-PPW-006).
grep -nE '\bcontinue\b' internal/template/templates/.git_hooks/pre-push
# Expected: >=1 match (skip path for the zero-local-SHA delete ref).
#
# Note (Tier S presence-level debt): AC-PPW-005/006/007 are presence greps — they assert the
# tokens exist, not that `continue` is syntactically bound to the zero-local-SHA test. The
# behavioral gate for the delete / new-branch edges is the run-phase smoke push plus the Go
# suite (AC-PPW-009). Accepted as documented Tier S debt per plan-auditor iter-1 D2.
```

---

## AC-PPW-008 — git stdin captured before ci-local (REQ-PPW-008 / D.2)

**Given** the script,
**When** comparing the line where stdin is captured against the `make ci-local` line,
**Then** the capture occurs BEFORE `make ci-local` so the ref-update stdin survives.

```bash
# Line where stdin is captured into a variable (cat into a var):
cap=$(grep -nE '\$\(cat\)|`cat`|=\"\$\(cat\b' internal/template/templates/.git_hooks/pre-push | head -1 | cut -d: -f1)
# Line of the ACTUAL make ci-local invocation (NOT the line-2 header comment, which also
# contains the substring "ci-local" — anchor to the exact invocation token):
ci=$(grep -n 'make -C "\$REPO_ROOT" -s ci-local' internal/template/templates/.git_hooks/pre-push | head -1 | cut -d: -f1)
[ -n "$cap" ] && [ -n "$ci" ] && [ "$cap" -lt "$ci" ] && echo "PASS: stdin capture($cap) before ci-local($ci)"
# Expected: PASS line printed.
```

---

## AC-PPW-009 — Mirror parity + install assertion + suite green (REQ-PPW-009/010/011)

**Given** the template and the Go constant,
**When** running the canonical byte-parity test,
**Then** it passes (template ↔ constant byte-identical).

```bash
go test ./internal/cli/ -run TestPrePushTemplateMatchesConstant -count=1
# Expected: ok (exit 0). Canonical byte-parity verification — NOT a hand grep diff.
```

**Given** the dev-repo root copy (hand-maintained leg),
**Then** it is byte-identical to the template.

```bash
diff internal/template/templates/.git_hooks/pre-push .git_hooks/pre-push && echo ROOT_BYTE_IDENTICAL
# Expected: ROOT_BYTE_IDENTICAL printed (no diff output).
```

**Given** a fresh repo install,
**When** running the install assertion test (after extending `wantStrings`),
**Then** the installed hook contains the `moai hook pre-push` invocation token and the
existing tokens.

```bash
go test ./internal/cli/ -run 'TestInstallPrePushHook' -count=1 -v
# Expected: all TestInstallPrePushHook* subtests PASS (FreshRepo asserts moai hook pre-push token).
# Caveat (vacuous-guard): -run 'TestInstallPrePushHook' is a prefix match that DOES capture
# TestInstallPrePushHook_FreshRepo / _SkipFlag / _PreservesUserHook / _OverwritesMoaiHook.
# Confirm "--- PASS: TestInstallPrePushHook_FreshRepo" appears in -v output (not "no tests to run").
```

**Given** the full affected packages,
**Then** the suite is green and `embedded.go` regenerated.

```bash
make build && go test ./internal/cli/ ./internal/template/... -count=1
# Expected: build succeeds, all packages ok.
```

---

## Definition of Done

- [ ] AC-PPW-001 .. AC-PPW-009 all pass on the post-change tree.
- [ ] `TestPrePushTemplateMatchesConstant` green (template ↔ constant byte-identical).
- [ ] Root `.git_hooks/pre-push` byte-identical to template (hand-maintained leg synced).
- [ ] `TestInstallPrePushHook_FreshRepo` asserts the `moai hook pre-push` token and passes
      (verified via `-v`, not vacuous).
- [ ] `make build` run; `embedded.go` regenerated; no hand-edits to it.
- [ ] No change to `runPrePush` Go logic; `enforce_on_push` template default stays `false`.
- [ ] Template block contains no internal markers (SPEC ID / REQ / dates / SHAs / OS paths).
- [ ] `golangci-lint run ./internal/cli/...` clean.

## Edge Cases Covered

| Edge case | Handling | AC |
|-----------|----------|-----|
| `moai` not installed | `command -v moai` guard skips block | AC-PPW-003 |
| `enforce_on_push` disabled (default) | engine self-gates → no-op | AC-PPW-004 |
| New branch push (remote SHA all-zeros) | enumerate locally-new commits | AC-PPW-006 |
| Branch delete (local SHA all-zeros) | skip ref (`continue`) | AC-PPW-007 |
| `make ci-local` fails | hook exits before convention block | AC-PPW-002 |
| stdin consumed by ci-local | captured into variable before ci-local | AC-PPW-008 |
| Empty stdin (no ref-update lines) | `$(cat)` exits 0 under `set -e`; loop iterates zero times | AC-PPW-008 / D.2 |
