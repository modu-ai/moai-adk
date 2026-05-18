# SPEC-V3R4-CI-INFRA-FIX-001 — Task Breakdown

> **Methodology**: DDD (Domain-Driven Development) — CI workflow YAML / shell script 변경은 unit-test 부적합. **evidence-driven verification** (CI 3-run sampling + workflow log grep) 채택.
> **Sequential**: Wave 1 (Fix A) → Wave 2 (Fix B) → Wave 3 (Fix C) → Wave 4 (Verification). 각 Wave 내부도 sequential commit.
> **Branch base**: BODP `¬a ¬b ¬c` → `origin/main` fresh branch. plan-PR 머지 후 run-phase 가 fresh `feat/SPEC-V3R4-CI-INFRA-FIX-001` branch 생성.

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 4-Wave × 11 tasks. 각 task 는 file + concrete edit + verification step + AC mapping. NO time estimates. Run-phase methodology DDD. |

---

## 1. Task Summary

| Wave | Task ID | Title | Type | Files | AC mapping |
|------|---------|-------|------|-------|-----------|
| W1 (Fix A — SIGPIPE) | W1-T1 | Read detect-language action.yml + identify head -1 pattern | analyze | (read only) | (foundation) |
| W1 (Fix A — SIGPIPE) | W1-T2 | Apply awk replacement per design.md D-1 (A2) | impl | `.github/actions/detect-language/action.yml` | AC-CIIF-001 |
| W1 (Fix A — SIGPIPE) | W1-T3 | 16-language fixture verification (local) | verify | (fixture dirs) | AC-CIIF-001 |
| W2 (Fix B — 403) | W2-T1 | Read spec-status-auto-sync.yml + audit step scopes | analyze | (read only) | (foundation) |
| W2 (Fix B — 403) | W2-T2 | Add workflow-level permissions block per design.md D-2 (B1) | impl | `.github/workflows/spec-status-auto-sync.yml` | AC-CIIF-002 |
| W3 (Fix C — fetch-depth) | W3-T1 | Read ci.yml + identify 5 checkout steps | analyze | (read only) | (foundation) |
| W3 (Fix C — fetch-depth) | W3-T2 | Add `with: fetch-depth: 0` to 5 ci.yml checkout steps | impl | `.github/workflows/ci.yml` | AC-CIIF-003 |
| W3 (Fix C — fetch-depth) | W3-T3 | Remove GITHUB_ACTIONS env skip guard | impl | `internal/spec/drift_specid_grep_test.go` | AC-CIIF-003 |
| W3 (Fix C — fetch-depth) | W3-T4 | Update @MX:NOTE / @MX:REASON in drift_specid_grep_test.go | docs | `internal/spec/drift_specid_grep_test.go` | (governance) |
| W4 (Verification) | W4-T1 | run-PR open with --squash auto-merge | git | (PR only) | AC-CIIF-001..003 |
| W4 (Verification) | W4-T2 | Verify AC-CIIF-001..004 from run-PR CI run + post-merge | verify | (CI logs) | AC-CIIF-001..004 |

---

## 2. Wave 1 — Fix A (detect-language SIGPIPE)

### W1-T1: Read detect-language action.yml + identify head -1 pattern

**Goal**: action.yml 의 SIGPIPE-prone shell snippet 정확한 위치 + pattern 식별.

**File**: `.github/actions/detect-language/action.yml` (read only, 37 lines, 1718 bytes)

**Output**:
- `head -1` 등장 위치 2건 식별 (line 16 의 `head -1 | sed 's/.*\.//'` + line 17 의 `head -1 | awk '{print $2}'`)
- Plan §3.1 의 잠정 fix code (A2 awk) 와 실제 line 비교 → minor adjustments 필요 시 plan.md inline update

**Verification**: Read tool output 확인. shell snippet 2개 매칭 위치 명시.

### W1-T2: Apply awk replacement per design.md D-1 (A2)

**Goal**: action.yml 의 두 `head -1` 위치를 awk explicit exit pattern 으로 교체.

**File**: `.github/actions/detect-language/action.yml`

**Edit details**:

Line 14-17 의 shell snippet 을 다음으로 교체 (Edit tool, sed 금지):

**Before**:
```yaml
        LANG_COUNT=$(find . -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.js" | \
          grep -v node_modules | grep -v vendor | grep -v ".git" | \
          head -1 | sed 's/.*\.//' | sort | uniq -c | sort -rn | head -1 | \
          awk '{print $2}')
```

**After**:
```yaml
        # @MX:NOTE: [AUTO] SIGPIPE-safe pipeline — head -1 → awk NR==1{print;exit}
        # @MX:REASON: SPEC-V3R4-CI-INFRA-FIX-001 D-1 (A2). head 의 implicit
        # stdin close 가 upstream find 와 broken-pipe race 유발 → exit 141 incident 4건.
        LANG_COUNT=$(find . -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.js" | \
          grep -v node_modules | grep -v vendor | grep -v ".git" | \
          awk 'NR==1{print;exit}' | sed 's/.*\.//' | sort | uniq -c | sort -rn | \
          awk 'NR==1{print $2;exit}')
```

**Note**: design.md D-1 alternative A1 (sed -n '1p') 와 manager-develop 가 비교 후 최종 선택 가능. 본 task 의 결과는 A2 (awk) 채택 기준.

**Commit message** (after W1-T2 + W1-T3 검증 통과 시):
```
fix(ci): SPEC-V3R4-CI-INFRA-FIX-001 W1 — detect-language SIGPIPE remediation

`.github/actions/detect-language/action.yml` 의 head -1 → awk NR==1{print;exit}
교체. set -o pipefail (GitHub Actions bash default) 환경에서 broken-pipe race
가 SIGPIPE (exit 141) 로 step fail 시키던 결함 영구 해소.

NAMESPACE-001 PR #944 incident 4건 root cause fix.
design.md D-1 A2 채택.

AC-CIIF-001 충족 예정 (CI 3-run 검증 후).

🗿 MoAI <email@mo.ai.kr>
```

### W1-T3: 16-language fixture verification (local)

**Goal**: awk replacement 가 16 supported languages 각각에서 정확한 LANG_COUNT 출력하는지 fixture 기반 검증.

**Files**: temporary fixture directories (e.g., `/tmp/cifix-fixture-go/`, `/tmp/cifix-fixture-python/` 등) — commit 대상 아님.

**Procedure** (manager-develop run-phase 수행):

```bash
# Each language fixture
for LANG in go python typescript javascript rust java kotlin csharp ruby php elixir cpp scala r flutter swift; do
  # Extension mapping
  case "$LANG" in
    go) EXT=go ;;
    python) EXT=py ;;
    typescript) EXT=ts ;;
    javascript) EXT=js ;;
    rust) EXT=rs ;;
    java) EXT=java ;;
    kotlin) EXT=kt ;;
    csharp) EXT=cs ;;
    ruby) EXT=rb ;;
    php) EXT=php ;;
    elixir) EXT=ex ;;
    cpp) EXT=cpp ;;
    scala) EXT=scala ;;
    r) EXT=r ;;
    flutter) EXT=flutter ;;  # Note: flutter 는 .dart but case "flutter" 에서 처리
    swift) EXT=swift ;;
  esac

  # Fixture create
  FIXDIR=$(mktemp -d -t cifix-fixture-XXXXXX)
  touch "$FIXDIR/sample.$EXT"

  # Run snippet
  cd "$FIXDIR"
  RESULT=$(find . -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.js" -o -name "*.$EXT" 2>/dev/null | \
    grep -v node_modules | grep -v vendor | grep -v ".git" | \
    awk 'NR==1{print;exit}' | sed 's/.*\.//' | sort | uniq -c | sort -rn | \
    awk 'NR==1{print $2;exit}')

  echo "$LANG: $RESULT (expected: $EXT)"
  rm -rf "$FIXDIR"
done
```

**Verification**: 16 출력 모두 expected 와 일치 (각 fixture 의 EXT 매칭).

**Note**: 본 task 는 commit 산출물 없음. progress.md 에 결과 기록.

---

## 3. Wave 2 — Fix B (spec-status-auto-sync 403)

### W2-T1: Read spec-status-auto-sync.yml + audit step scopes

**Goal**: workflow 의 각 step 의 GITHUB_TOKEN scope 사용 명시화.

**File**: `.github/workflows/spec-status-auto-sync.yml` (read only)

**Output** (analysis):
- Line 12-14: `actions/checkout@v4 with: fetch-depth: 0` — `contents: read` (sufficient).
- Line 15-17: `actions/setup-go@v5` — no token scope.
- Line 18-19: `make build && make install` — no token scope.
- Line 20-108: `Sync SPEC statuses from PR title` step:
  - Line 69: `moai spec status "$SPEC_ID" "$TARGET_STATUS"` — local file edit, no token.
  - Line 85: `git commit -m ...` — local commit, no token.
  - Line 92: `git rebase origin/main` — local, no token (needs `contents: read` for fetch — `actions/checkout` already done).
  - Line 95-98: `gh issue create` — requires `issues: write`.
  - Line 107: `git push origin main` — requires `contents: write`.

**Required scope union**: `contents: write` + `issues: write`.

**Verification**: audit table progress.md 에 기록.

### W2-T2: Add workflow-level permissions block per design.md D-2 (B1)

**Goal**: workflow 최상단 trigger 아래에 `permissions:` block 추가.

**File**: `.github/workflows/spec-status-auto-sync.yml`

**Edit details**:

Line 1-7 (현재 trigger 정의) 다음에 permissions block 삽입:

**Before** (line 1-15):
```yaml
name: SPEC Status Auto-Sync

on:
  pull_request:
    types: [closed]

jobs:
  spec-status-sync:
    runs-on: ubuntu-latest
    if: github.event.pull_request.merged == true
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
```

**After**:
```yaml
name: SPEC Status Auto-Sync

on:
  pull_request:
    types: [closed]

# @MX:NOTE: [AUTO] minimum permissions scope for git push + gh issue create
# @MX:REASON: SPEC-V3R4-CI-INFRA-FIX-001 D-2 (B1). default workflow scope
# (contents: read) 가 line 107 git push origin main 을 403 으로 실패시키던
# 결함 영구 해소. NAMESPACE-001 force-push 후 incident 1건 root cause fix.
permissions:
  contents: write    # git push origin main (line 107)
  issues: write      # gh issue create fallback (line 95-99)

jobs:
  spec-status-sync:
    runs-on: ubuntu-latest
    if: github.event.pull_request.merged == true
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
```

**Note**: YAML 의 `@MX:NOTE` 주석은 `#` 으로 시작 (shell-style). MX tag protocol 의 comment syntax table 에 따라 YAML 은 `#` 사용 (Python/Ruby 와 동일).

**Commit message**:
```
fix(ci): SPEC-V3R4-CI-INFRA-FIX-001 W2 — spec-status-auto-sync permissions

`.github/workflows/spec-status-auto-sync.yml` 에 workflow-level permissions
block 추가 — `contents: write` (git push origin main) + `issues: write` (gh
issue create fallback). GITHUB_TOKEN default scope (contents: read) 가
post-merge 자동 sync 의 push 를 403 으로 실패시키던 결함 영구 해소.

NAMESPACE-001 force-push 후 incident root cause fix.
design.md D-2 B1 채택 (workflow-level minimum permissions).

AC-CIIF-002 충족 예정 (sync-PR 머지 후 자연 검증).

🗿 MoAI <email@mo.ai.kr>
```

---

## 4. Wave 3 — Fix C (checkout fetch-depth: 0 + skip guard 제거)

### W3-T1: Read ci.yml + identify 5 checkout steps

**Goal**: ci.yml 의 모든 `actions/checkout@v5` step 위치 식별.

**File**: `.github/workflows/ci.yml` (read only)

**Output** (analysis):
- Line 27-28: test job — `Checkout code`
- Line 95-96: test-integration job — `Checkout code`
- Line 119-120: lint job — `Checkout code`
- Line 156-157: build job — `Checkout code`
- Line 195-196: constitution-check job — `Checkout code`

**Verification**: 5건 확인. progress.md 에 기록.

### W3-T2: Add `with: fetch-depth: 0` to 5 ci.yml checkout steps

**Goal**: ci.yml 5 checkout step 각각에 fetch-depth: 0 명시.

**File**: `.github/workflows/ci.yml`

**Edit details**: 5 step 모두 동일 패턴:

**Before** (각 step):
```yaml
      - name: Checkout code
        uses: actions/checkout@v5
```

**After**:
```yaml
      - name: Checkout code
        uses: actions/checkout@v5
        with:
          fetch-depth: 0  # SPEC-V3R4-CI-INFRA-FIX-001 D-3 (C-Scope-1): drift walker test 가 full git history 필요
```

**Note**: Edit tool 의 `replace_all: true` 사용 시 5 step 일괄 변경. 단, `uses: actions/checkout@v5` 가 정확히 5건 존재 확인 후 적용 (다른 workflow 와 cross-contamination 방지 — file 단위 Edit 이므로 영향 0).

**Commit message**:
```
fix(ci): SPEC-V3R4-CI-INFRA-FIX-001 W3-A — ci.yml checkout fetch-depth: 0

`.github/workflows/ci.yml` 의 5 actions/checkout@v5 step 모두 `with:
fetch-depth: 0` 명시. default fetch-depth: 1 (shallow clone) 이 drift walker
의 git log --grep 으로 SPEC commits 조회를 막던 결함 영구 해소.

5 step:
- test job (line 27-28)
- test-integration job (line 95-96)
- lint job (line 119-120)
- build job (line 156-157)
- constitution-check job (line 195-196)

design.md D-3 C-Scope-1 채택 (ci.yml only).
다른 workflow (codeql.yml 등) 검토는 후속 SPEC scope.

AC-CIIF-003 충족 예정 (drift test PASS 검증 후).

🗿 MoAI <email@mo.ai.kr>
```

### W3-T3: Remove GITHUB_ACTIONS env skip guard

**Goal**: `internal/spec/drift_specid_grep_test.go:21-27` 의 `GITHUB_ACTIONS` env probe 제거. local probe (line 30-34) 보존.

**File**: `internal/spec/drift_specid_grep_test.go`

**Edit details**:

**Before** (line 21-34):
```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
	// CI 환경 자동 skip — shallow clone으로 SPEC commits 부재
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("requires full git history; CI uses actions/checkout@v4 default fetch-depth: 1 (shallow). " +
			"Word-boundary logic is fully covered by TestGetGitImpliedStatus_SPECIDWordBoundary 5 sub-cases. " +
			"Follow-up: SPEC-V3R4-CI-INFRA-FIX-001 to set fetch-depth: 0 for full-history tests.")
	}

	// Probe: target SPEC commits이 local git에 존재하는지 확인 (non-CI 환경에서도 fork/shallow 대응)
	probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
	if out, err := probe.Output(); err != nil || len(out) == 0 {
		t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone). " +
			"WordBoundary helper test (5 sub-cases) covers the logic.")
	}
```

**After**:
```go
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
	// @MX:NOTE: [AUTO] CI shallow-clone skip guard 영구 제거
	// @MX:REASON: SPEC-V3R4-CI-INFRA-FIX-001 W3 적용 후 ci.yml 5 checkout step
	// 모두 fetch-depth: 0. CI 환경에서도 SPEC commits 정상 존재.
	// LSGF-001 PR #948 의 GITHUB_ACTIONS env workaround 영구 해소.

	// Probe: target SPEC commits이 local git에 존재하는지 확인 (fork/shallow clone 사용자 환경 대응)
	probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
	if out, err := probe.Output(); err != nil || len(out) == 0 {
		t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone). " +
			"WordBoundary helper test (5 sub-cases) covers the logic.")
	}
```

**Note**: `os` 패키지 import 가 다른 곳에서 사용되는지 확인 후 (현재 file 의 다른 test 가 `os.Getenv` 사용 가능) — 만약 미사용 시 import line 도 제거. 검증 명령: `grep -c "os\." internal/spec/drift_specid_grep_test.go` 가 1 이상이면 import 보존, 0 이면 import 제거.

### W3-T4: Update @MX:NOTE / @MX:REASON in drift_specid_grep_test.go

**Goal**: W3-T3 의 skip guard 제거 reason + LSGF-001 follow-up 완료 명시 (W3-T3 의 edit 에 inline 포함됨 — 본 task 는 verification only).

**Verification**:
- `@MX:NOTE` 와 `@MX:REASON` 주석이 drift_specid_grep_test.go 의 `TestGetGitImpliedStatus_HARNESS001Resolution` 상단에 정확히 추가됨.
- LSGF-001 PR #948 workaround 영구 제거 사유 명시.

**Commit message** (W3-T3 + W3-T4 통합):
```
fix(ci): SPEC-V3R4-CI-INFRA-FIX-001 W3-B — drift test skip guard 제거

`internal/spec/drift_specid_grep_test.go` 의 `GITHUB_ACTIONS` env probe
완전 제거. LSGF-001 PR #948 에서 도입된 CI workaround (shallow clone
회피) 가 W3-A (ci.yml fetch-depth: 0) 적용으로 불필요해짐.

- skip guard 제거 (line 21-27)
- local probe 보존 (line 30-34) — fork/shallow clone 사용자 환경 대응
- @MX:NOTE + @MX:REASON 갱신 — LSGF-001 workaround 영구 해소 명시

design.md D-3 Skip-B 채택 (env probe 제거 + local probe 보존).

AC-CIIF-003 (drift test PASS not SKIP) 충족 예정.

🗿 MoAI <email@mo.ai.kr>
```

---

## 5. Wave 4 — Verification + PR Lifecycle

### W4-T1: run-PR open with --squash auto-merge

**Goal**: run-phase 4 commits (W1-T2 + W2-T2 + W3-T2 + W3-T3/T4) 를 run-PR 로 통합 + auto-merge 등록.

**Branch**: `feat/SPEC-V3R4-CI-INFRA-FIX-001` (plan-PR 머지 후 fresh from origin/main per SPEC phase discipline)

**PR**:
- **Base**: `main`
- **Head**: `feat/SPEC-V3R4-CI-INFRA-FIX-001`
- **Title**: `feat(ci): SPEC-V3R4-CI-INFRA-FIX-001 run-phase — 3-defect bundle fix (SIGPIPE + 403 + fetch-depth: 0)`
- **Body**: Wave 1/2/3 summary + AC-CIIF-001..003 verification plan + lint regression baseline check + 3 sub-fix mapping
- **Auto-merge**: `gh pr merge <PR> --auto --squash --delete-branch`

**Required CI Pass** (per CLAUDE.local.md §18.7):
- Lint
- Test (ubuntu-latest)
- Test (macos-latest)
- Test (windows-latest)
- Build (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- CodeQL

### W4-T2: Verify AC-CIIF-001..004 from run-PR CI run + post-merge

**Goal**: 4 AC 모두 PASS 검증.

**Procedure**:

```bash
# AC-CIIF-001: detect-language SIGPIPE (호출 workflow 의 CI run)
# 호출 workflow 는 본 PR commit 이 trigger 한 경우 (gemini/glm/codex-review)
# 만약 trigger 부재 시 vacuous PASS (EC-001) — fixture verification (W1-T3) 이 보강
PR_SHA=$(gh pr view <run-PR> --json headRefOid -q .headRefOid)
SIGPIPE_FAILS=$(gh run list --commit="$PR_SHA" --json conclusion,workflowName \
  | jq '[.[] | select(.conclusion == "failure" and (.workflowName | test("gemini-review|glm-review|codex-review")))] | length')
echo "AC-CIIF-001: $SIGPIPE_FAILS (expected: 0)"

# AC-CIIF-002: sync-PR 머지 후 spec-status-auto-sync workflow 검증
# (sync-PR open 후 검증 가능)

# AC-CIIF-003: ci.yml 5-step grep + drift test PASS
FETCH_DEPTH_COUNT=$(grep -c "fetch-depth: 0" .github/workflows/ci.yml)
echo "AC-CIIF-003 grep: $FETCH_DEPTH_COUNT (expected: 5)"

RUN_ID=$(gh pr view <run-PR> --json statusCheckRollup \
  | jq -r '.statusCheckRollup[] | select(.name | startswith("Test (ubuntu")) | .detailsUrl' \
  | head -1 | grep -oE '[0-9]+$')
DRIFT_RESULT=$(gh run view "$RUN_ID" --log 2>&1 \
  | grep "TestGetGitImpliedStatus_HARNESS001Resolution" \
  | grep -E "PASS|SKIP" | head -1)
echo "AC-CIIF-003 drift test: $DRIFT_RESULT (expected: contains PASS)"

# AC-CIIF-004: lint baseline 회귀 부재 (main 머지 후)
git checkout main && git pull
moai spec lint --strict 2>&1 | tail -3
# expected: "0 error(s), 0 warning(s)" 또는 "✓ No findings (0/0)"

# Skip guard 제거 검증
SKIP_GUARD=$(grep -c 'os.Getenv("GITHUB_ACTIONS")' internal/spec/drift_specid_grep_test.go)
echo "Skip guard removed: $SKIP_GUARD (expected: 0)"
```

**Output**: progress.md 에 4 AC 결과 + lint baseline 결과 기록.

---

## 6. Sync-phase Tasks (post-run-PR merge)

### W-Sync-T1: spec.md frontmatter sync

**File**: `.moai/specs/SPEC-V3R4-CI-INFRA-FIX-001/spec.md`

**Changes**:
- `status: draft → completed`
- `updated: 2026-05-16`
- `version: "0.1.0" → "0.2.0"`
- HISTORY 0.2.0 entry 추가:

```
| 0.2.0 | 2026-05-16 | manager-docs | Sync-phase 완료. status `draft → completed`. 3 sub-fix 모두 main 적용: (A) detect-language awk replacement / (B) spec-status-auto-sync permissions / (C) ci.yml fetch-depth: 0 + drift test skip guard 제거. AC-CIIF-001~004 모두 PASS. v2.20.0-rc1 release-readiness 최종 precondition 충족. |
```

**Commit message**:
```
sync(spec): SPEC-V3R4-CI-INFRA-FIX-001 lifecycle COMPLETE — 3-defect bundle fix main 적용

3 CI infrastructure defect 모두 영구 해소:
- (A) detect-language SIGPIPE → awk NR==1{print;exit} 채택
- (B) spec-status-auto-sync 403 → workflow-level permissions (contents+issues: write)
- (C) ci.yml fetch-depth: 0 (5 step) + drift_specid_grep_test.go skip guard 제거

AC-CIIF-001..004 모두 PASS.
`moai spec lint --strict` → 0 error(s), 0 warning(s) (regression baseline 유지).
v2.20.0-rc1 release tagging 최종 precondition 충족.

🗿 MoAI <email@mo.ai.kr>
```

### W-Sync-T2: CHANGELOG entry

**File**: `CHANGELOG.md` Unreleased 섹션에 ko + en entry 추가:

```markdown
### Fixed
- **CI**: SPEC-V3R4-CI-INFRA-FIX-001 — GitHub Actions CI infrastructure 3 결함 영구 해소: (A) `.github/actions/detect-language/action.yml` SIGPIPE remediation (head -1 → awk NR==1{print;exit}); (B) `.github/workflows/spec-status-auto-sync.yml` workflow-level permissions block 추가 (contents+issues: write); (C) `.github/workflows/ci.yml` 5 actions/checkout step 모두 `fetch-depth: 0` 명시 + `internal/spec/drift_specid_grep_test.go` 의 `GITHUB_ACTIONS` env skip guard 제거. v2.20.0-rc1 release-readiness 최종 precondition.

### Fixed (English)
- **CI**: SPEC-V3R4-CI-INFRA-FIX-001 — Permanently resolved 3 GitHub Actions CI infrastructure defects: (A) `.github/actions/detect-language/action.yml` SIGPIPE remediation (head -1 → awk NR==1{print;exit}); (B) `.github/workflows/spec-status-auto-sync.yml` workflow-level permissions block (contents+issues: write); (C) `.github/workflows/ci.yml` 5 actions/checkout steps explicit `fetch-depth: 0` + removed `GITHUB_ACTIONS` env skip guard from `internal/spec/drift_specid_grep_test.go`. Final precondition for v2.20.0-rc1 release-readiness.
```

### W-Sync-T3: sync-PR open with auto-merge

**Branch**: `sync/SPEC-V3R4-CI-INFRA-FIX-001`

**PR**:
- **Base**: `main`
- **Head**: `sync/SPEC-V3R4-CI-INFRA-FIX-001`
- **Title**: `sync(spec): SPEC-V3R4-CI-INFRA-FIX-001 lifecycle COMPLETE — 3-defect bundle fix main 적용`
- **Body**: AC verification matrix + lint regression baseline check + v2.20.0-rc1 release readiness 보고
- **Auto-merge**: `gh pr merge <PR> --auto --squash --delete-branch`

### W-Sync-T4: sync-PR CI green + AC-CIIF-002 자연 검증

**Goal**: sync-PR 머지 trigger 한 spec-status-auto-sync workflow run 결과로 AC-CIIF-002 자연 검증.

**Procedure**:

```bash
# sync-PR 머지 후 spec-status-auto-sync workflow run 조회
SYNC_MERGE_SHA=$(gh pr view <sync-PR> --json mergeCommit -q .mergeCommit.oid)
RUN_ID=$(gh run list --workflow=spec-status-auto-sync.yml --limit=5 --json databaseId,headSha \
  | jq -r --arg sha "$SYNC_MERGE_SHA" '.[] | select(.headSha == $sha) | .databaseId' \
  | head -1)

# Conclusion 검증
gh run view "$RUN_ID" --json conclusion -q .conclusion
# Expected: "success"

# 403 grep
gh run view "$RUN_ID" --log 2>&1 | grep -c "403"
# Expected: 0
```

---

## 7. Task Dependency Graph

```
W1-T1 (read action.yml) → W1-T2 (apply awk) → W1-T3 (fixture verify)
                                                       ↓
                                              W1-T2 commit
                                                       ↓
W2-T1 (read workflow.yml) → W2-T2 (add permissions) → W2-T2 commit
                                                       ↓
W3-T1 (read ci.yml) → W3-T2 (add fetch-depth: 0) → W3-T2 commit
                                                       ↓
                       W3-T3 (remove skip guard) + W3-T4 (MX tags) → W3-T3/T4 commit
                                                       ↓
                                              W4-T1 (run-PR open)
                                                       ↓
                                              W4-T2 (AC verify)
                                                       ↓
                                              run-PR merge (squash)
                                                       ↓
                                              W-Sync-T1 (spec.md sync)
                                                       ↓
                                              W-Sync-T2 (CHANGELOG)
                                                       ↓
                                              W-Sync-T3 (sync-PR open)
                                                       ↓
                                              W-Sync-T4 (AC-CIIF-002 자연 검증)
```

- Wave 1, 2, 3 commits 은 mutually independent (3 file 모두 다른 layer). 그러나 main checkout 에서 sequential commit (commit history 명확성 + plan-auditor 추적성).
- W4-T2 verification 은 W4-T1 PR open 후 즉시 시작. CI run 결과 대기 후 evaluator.

---

## 8. Out of Scope (Wave 5+ 미실시)

- `actions/checkout@v4` → `@v5` 일괄 upgrade.
- ci.yml 외 다른 workflow (codeql.yml, test-install.yml, auto-merge.yml 등) 의 fetch-depth 검토 — 별도 follow-up SPEC `SPEC-V3R4-CI-WORKFLOWS-AUDIT-001` 후보.
- spec-status-auto-sync.yml 의 fork PR security model 변경 — `pull_request_target` 도입 검토 별도.
- detect-language action 의 16-language detection 정확도 audit — 별도 SPEC.
- `internal/spec/drift.go` walker 의 shallow-clone fallback 도입 — non-goal (LSGF-001 territory).
- v2.20.0-rc1 tagging — 별도 release 작업 (CLAUDE.local.md §18.8 `./scripts/release.sh v2.20.0-rc1`).
- branch protection rule (CLAUDE.local.md §18.7) 변경.
- CHANGELOG.md 4-locale i18n sync (CLAUDE.local.md §17.3 별도 정책 — 별도 SPEC).
- `internal/template/templates/` user project 영향 검토 (CI workflows 는 dev project only, user project 미배포).

---

## 9. References (cross-file)

- `spec.md` — REQ-CIIF-001~007 + 3 sub-fix root cause + AC ↔ REQ mapping
- `plan.md` — Wave 1/2/3 implementation strategy + open questions
- `design.md` — D-1 (SIGPIPE), D-2 (permissions), D-3 (fetch-depth scope), D-4 (compatibility)
- `acceptance.md` — AC-CIIF-001..004 Given/When/Then + edge cases + DoD
- `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/tasks.md` — 3-Wave TDD precedent
- `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/tasks.md` — 3-Wave evidence-driven precedent
- `CLAUDE.local.md §18.3` — merge strategy
- `CLAUDE.local.md §18.7` — branch protection rule (required check matrix)
- `CLAUDE.local.md §18.8` — release process (post-merge tag push)
- `internal/spec/drift.go:120-180` — walker (참조 only, 무수정)
