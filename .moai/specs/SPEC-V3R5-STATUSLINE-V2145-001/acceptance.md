# SPEC-V3R5-STATUSLINE-V2145-001 Acceptance Criteria

All acceptance criteria are binary (PASS/FAIL) and verifiable via a single shell command or `grep` predicate. Verification batch SHOULD execute in a single orchestrator turn per agent-common-protocol §Parallel Execution.

## Milestone 1 — Disappearing Hotfix

### AC-SLV-001 (REQ-SLV-001): DEBUG_STATUSLINE default 0 in template + rendered

**Verification**:
```bash
grep -E 'DEBUG_STATUSLINE:-0' \
  internal/template/templates/.moai/status_line.sh.tmpl \
  .moai/status_line.sh
```

**PASS**: Both files contain `DEBUG_STATUSLINE:-0` (or equivalent — `DEBUG_STATUSLINE:-0`).
**FAIL**: Either file contains `DEBUG_STATUSLINE:-1` or omits the variable.

### AC-SLV-002 (REQ-SLV-002): Debug fork guarded by explicit opt-in

**Verification**:
```bash
grep -A1 -E '\$\{DEBUG_STATUSLINE:-0\}' \
  internal/template/templates/.moai/status_line.sh.tmpl \
  .moai/status_line.sh \
  | grep -E '= "1"'
```

**PASS**: Both files use the `if [ "${DEBUG_STATUSLINE:-0}" = "1" ]` guard pattern.
**FAIL**: The debug block executes unconditionally or uses a different default.

### AC-SLV-003 (REQ-SLV-003): Dead `echo ""` lines removed

**Verification**:
```bash
# Count echo "" lines in the template (should be 0)
grep -c '^[[:space:]]*echo ""[[:space:]]*$' \
  internal/template/templates/.moai/status_line.sh.tmpl
```

**PASS**: Output is `0` (no `echo ""` lines remain around `exec moai statusline` calls).
**FAIL**: Output is `>= 1`.

### AC-SLV-004 (REQ-SLV-004): `statusLine.padding` documented in all 4 docs-site locales

**Verification**:
```bash
for locale in ko en ja zh; do
  grep -l 'statusLine.padding' \
    "docs-site/content/${locale}/advanced/statusline.md" \
    || echo "MISSING: ${locale}"
done
```

**PASS**: 4 paths echoed (one per locale), zero `MISSING:` lines.
**FAIL**: Any `MISSING:` line printed.

### AC-SLV-005 (REQ-SLV-005): No user-specific absolute path in rendered wrapper

**Verification**:
```bash
grep -E '/Users/[^/]+/go/bin/moai' .moai/status_line.sh && echo "FAIL" || echo "PASS"
```

**PASS**: Output `PASS` (no `/Users/<name>/go/bin/moai` literal found).
**FAIL**: Output `FAIL` (hardcoded user-specific path detected).

## Milestone 2 — PR Segment Addition

### AC-SLV-010 (REQ-SLV-010): `StdinData.PR` field exists with correct JSON tags

**Verification**:
```bash
grep -E '^\s+PR\s+\*PRInfo\s+`json:"pr"`' internal/statusline/types.go && \
grep -E 'type PRInfo struct' internal/statusline/types.go && \
grep -E '^\s+Number\s+int\s+`json:"number"`' internal/statusline/types.go && \
grep -E '^\s+URL\s+string\s+`json:"url"`' internal/statusline/types.go && \
grep -E '^\s+ReviewState\s+string\s+`json:"review_state"`' internal/statusline/types.go
```

**PASS**: All 5 grep commands return exit 0 (all 5 lines present).
**FAIL**: Any grep returns exit 1.

### AC-SLV-011 (REQ-SLV-011): `WorkspaceInfo.Repo` field exists

**Verification**:
```bash
grep -E '^\s+Repo\s+\*RepoInfo\s+`json:"repo"`' internal/statusline/types.go && \
grep -E 'type RepoInfo struct' internal/statusline/types.go && \
grep -E '^\s+Host\s+string\s+`json:"host"`' internal/statusline/types.go && \
grep -E '^\s+Owner\s+string\s+`json:"owner"`' internal/statusline/types.go && \
grep -E '^\s+Name\s+string\s+`json:"name"`' internal/statusline/types.go
```

**PASS**: All 5 grep commands return exit 0.
**FAIL**: Any grep returns exit 1.

### AC-SLV-012 (REQ-SLV-012): PR segment default off

**Verification**:
```bash
go test -run 'TestPRSegment.*DefaultOff|TestPRSegment.*Disabled' \
  ./internal/statusline/... -v
```

**PASS**: Test exits 0 with at least one `--- PASS:` line for a default-off scenario.
**FAIL**: Test exits non-zero or no matching test name found.

### AC-SLV-013 (REQ-SLV-013): PR segment render format `#<number> ⌥<state>`

**Verification**:
```bash
go test -run 'TestPRSegment.*Format|TestPRSegment.*Render' \
  ./internal/statusline/... -v
```

**PASS**: Test exits 0; test source contains the literal format characters `#` and `⌥` (or `⌥`) in expected-output strings.

```bash
grep -E '"#[0-9]+ ⌥' internal/statusline/renderer_test.go
```

**FAIL**: Either go test fails or grep returns exit 1.

### AC-SLV-014 (REQ-SLV-014): Review-state color coding for all 5 cases

**Verification**:
```bash
go test -run 'TestPRSegment.*Color|TestPRReviewState' \
  ./internal/statusline/... -v -count=1
```

**PASS**: Test exits 0 with PASS lines for at least the 5 cases (`approved` → green, `pending` → yellow, `changes_requested` → red, `draft` → gray, unknown → default).
**FAIL**: Any subtest fails OR fewer than 5 review_state cases covered.

Additional sentinel:
```bash
grep -cE 'approved|pending|changes_requested|draft' internal/statusline/renderer_test.go
```

PASS: count >= 4 (all 4 known states appear in test source).

### AC-SLV-015 (REQ-SLV-015): No segment emitted when PR absent

**Verification**:
```bash
go test -run 'TestPRSegment.*Absent|TestPRSegment.*Nil|TestPRSegment.*ZeroNumber' \
  ./internal/statusline/... -v
```

**PASS**: Test exits 0; at least 2 subtests cover (a) `pr == nil` and (b) `pr.Number == 0`.
**FAIL**: Test fails or absence cases not covered.

### AC-SLV-016 (REQ-SLV-016): `SegmentPR` constant + builder branch

**Verification**:
```bash
grep -E '^\s+SegmentPR\s+=\s+"pr"' internal/statusline/types.go && \
grep -E 'SegmentPR' internal/statusline/builder.go && \
grep -E 'SegmentPR' internal/statusline/renderer.go
```

**PASS**: All 3 greps return exit 0 (constant defined in types.go, referenced in both builder.go and renderer.go).
**FAIL**: Any grep returns exit 1.

### AC-SLV-017 (REQ-SLV-017): Coverage ≥85% on changed files

**Verification**:
```bash
go test -coverprofile=/tmp/statusline-cover.out ./internal/statusline/... && \
go tool cover -func=/tmp/statusline-cover.out | grep -E '(types|builder|renderer)\.go' | \
  awk '{gsub("%",""); if ($NF < 85.0) print "FAIL: " $0; else print "PASS: " $0}'
```

**PASS**: All matched lines start with `PASS:`.
**FAIL**: Any line starts with `FAIL:`.

Note: AC-SLV-017 measures per-function coverage. Aggregate package coverage should also pass via existing TRUST 5 quality gate.

## Milestone 3 — docs-site 4-locale sync

### AC-SLV-020 (REQ-SLV-020): Korean canonical section exists

**Verification**:
```bash
grep -E '## PR ?세그먼트|## PR Segment|segments\.pr' \
  docs-site/content/ko/advanced/statusline.md
```

**PASS**: Korean file contains a heading section for PR segment AND the `segments.pr` config key reference.
**FAIL**: Either grep miss.

Additional sentinel:
```bash
grep -E 'review_state|approved|pending|changes_requested|draft' \
  docs-site/content/ko/advanced/statusline.md | wc -l
```

PASS: count >= 4 (all 4 review_state values documented in Korean canonical).

### AC-SLV-021 (REQ-SLV-021): 4-locale parity

**Verification**:
```bash
for locale in ko en ja zh; do
  if grep -q 'segments.pr' "docs-site/content/${locale}/advanced/statusline.md"; then
    echo "PASS: ${locale}"
  else
    echo "FAIL: ${locale}"
  fi
done
```

**PASS**: All 4 lines start with `PASS:`.
**FAIL**: Any line starts with `FAIL:`.

### AC-SLV-022 (REQ-SLV-022): docs CI passes + URL blacklist clean

**Verification**:
```bash
# Existence check first — script may live at scripts/docs-i18n-check.sh or docs-site/scripts/
test -x scripts/docs-i18n-check.sh && bash scripts/docs-i18n-check.sh || \
  echo "SKIP: docs-i18n-check.sh not found (acceptable if docs-site uses different CI mechanism)"

# URL blacklist enforcement
! grep -rE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr' \
  docs-site/content/{ko,en,ja,zh}/advanced/statusline.md
```

**PASS**: First command returns 0 OR prints SKIP. Second command (negated grep) returns 0 (no blacklisted URLs found).
**FAIL**: First command exits non-zero (docs CI script fails) OR second command exits non-zero (blacklisted URL detected).

## Cross-Milestone

### AC-SLV-100: Zero-regression on existing 14 segments

**Verification**:
```bash
go test ./internal/statusline/... -count=1 -v 2>&1 | \
  grep -cE '^--- (PASS|FAIL):' && \
go test ./internal/statusline/... -count=1
```

**PASS**: All existing tests pass (no FAIL lines in verbose output); aggregate `go test` exits 0.
**FAIL**: Any pre-existing test fails.

### AC-SLV-101: golangci-lint baseline does not regress

**Verification**:
```bash
golangci-lint run --timeout=2m ./internal/statusline/... 2>&1 | tail -20
```

**PASS**: Zero new issues introduced (compare against pre-change baseline measured at run-phase Section C of the manager-develop prompt template per `manager-develop-prompt-template.md` §C.3).
**FAIL**: New lint issues introduced.

### AC-SLV-102: `make build` regenerates embedded templates successfully

**Verification**:
```bash
make build && \
git diff --quiet internal/template/embedded.go || \
  echo "embedded.go regenerated — expected after template edit"
```

**PASS**: `make build` exits 0. Either `embedded.go` is unchanged (template byte-identical) or shows a delta consistent with the M1 template edit.
**FAIL**: `make build` exits non-zero.

### AC-SLV-103: Conventional Commits format

**Verification**:
```bash
git log --format='%s' origin/main..HEAD | \
  grep -vE '^(feat|fix|docs|style|refactor|perf|test|chore|revert)(\(.+\))?:' \
  && echo "FAIL" || echo "PASS"
```

**PASS**: Output `PASS` (all commits on branch follow Conventional Commits).
**FAIL**: Any commit message lacks the type prefix.

## Definition of Done

All of the following must be true:

- [ ] All 17 REQ-level ACs (AC-SLV-001..022) PASS
- [ ] All 4 cross-milestone ACs (AC-SLV-100..103) PASS
- [ ] `go test ./internal/statusline/... -cover` reports ≥85% on changed files
- [ ] `golangci-lint run --timeout=2m ./...` baseline does not regress
- [ ] `make build` exits 0 and `internal/template/embedded.go` is in sync with template source
- [ ] 4 docs-site locale files present (ko canonical + en/ja/zh translations)
- [ ] No `Users/goos` literal in `.moai/status_line.sh` or template source
- [ ] No `DEBUG_STATUSLINE:-1` literal in shell wrappers
- [ ] PR opened with title `feat(statusline): align with Claude Code v2.1.145 + fix disappearance` and labels `type:feat` `area:statusline` `priority:P1`
- [ ] PR body references SPEC-V3R5-STATUSLINE-V2145-001 and Closes any associated GitHub issue
- [ ] sync-phase Manager-docs `/moai sync` updates progress.md status `draft → completed`

## Verification Batch Recipe (single-turn parallel execution)

Per agent-common-protocol §Parallel Execution Read-only verification batching, the orchestrator SHOULD invoke the following 7 commands as separate Bash tool calls within the same response turn after manager-develop reports run-phase completion:

```bash
# 1. Functional — full test suite
go test ./internal/statusline/... -count=1

# 2. Coverage on changed files
go test -coverprofile=/tmp/statusline-cover.out ./internal/statusline/...

# 3. Subagent boundary discipline (sanity — no agent code in statusline package)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/statusline/ | grep -v "_test.go" | grep -v "// " || echo "PASS: no AskUserQuestion misuse"

# 4. Sentinel scan — no hardcoded user-path
grep -rE '/Users/[a-z]+/' .moai/status_line.sh internal/template/templates/.moai/status_line.sh.tmpl || echo "PASS: no hardcoded user paths"

# 5. CLI smoke — moai binary still works
go run ./cmd/moai --version

# 6. Benchmark micro-suite — render latency baseline (optional)
go test -bench=. -benchmem -run=^$ ./internal/statusline/... -benchtime=3x

# 7. Lint baseline
golangci-lint run --timeout=2m ./internal/statusline/...
```

All 7 commands are read-only (no shared file mutation) and safe to parallelize. Expected total wall-time: ~30-60 s if executed in parallel; ~5 min if serial.
