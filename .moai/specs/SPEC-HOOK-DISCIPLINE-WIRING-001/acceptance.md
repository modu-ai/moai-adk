# Acceptance Criteria — SPEC-HOOK-DISCIPLINE-WIRING-001

> GEARS-format AC. Each criterion is verifiable by a concrete shell / grep / go-test command an auditor can run from the repo root.
> Status: draft.

## D. AC Matrix

| AC | Maps to REQ | Severity | Summary |
|----|-------------|----------|---------|
| AC-HDW-001 | REQ-HDW-001, 008 | MUST | status-transition registered in settings.json.tmpl PostToolUse (correct quoting + timeout) |
| AC-HDW-002 | REQ-HDW-003, 004, 005 | MUST | sync-gate language-auto-detect: non-Go fixture skips Go tooling gracefully (exit 0); no-marker fixture silent-passes |
| AC-HDW-003 | REQ-HDW-006, 008 | MUST | sync-gate registered in settings.json.tmpl Stop (correct quoting + timeout) |
| AC-HDW-004 | REQ-HDW-007 | MUST | team-ac-verify NOT registered (grep settings.json.tmpl returns 0) + file still present |
| AC-HDW-005 | REQ-HDW-001, 006 | MUST | existing handle-*.sh handlers still present (coexistence, no regression) |
| AC-HDW-006 | REQ-HDW-009 | MUST | template-neutrality CI guard passes (internal-content leak ONLY — NOT a Go-bias guard; see scope note) |
| AC-HDW-007 | REQ-HDW-010 | MUST | local↔template settings + sync-gate parity (manual mirror; NOT via make build); dev-intent keys untouched |
| AC-HDW-008 | REQ-HDW-002, 006 | MUST | advisory-only: wired default does not exit 2 (warn-first config asserted) |
| AC-HDW-009 | REQ-HDW-003, 009 | MUST | Go-bias boundedness: `go vet`/`golangci-lint`/`go.mod` tokens appear ONLY inside the Go case branch (the REAL automated language-neutrality guard) |

---

## AC-HDW-001 — status-transition registered in PostToolUse

| Field | Value |
|-------|-------|
| **Given** | The run-phase wiring milestone M2 is complete and `internal/template/templates/.claude/settings.json.tmpl` has been edited. |
| **When** | An auditor inspects the PostToolUse hooks array of settings.json.tmpl. |
| **Then** | The array contains a command entry pointing at `status-transition-ownership.sh`, using the platform-conditional quoting pattern (`bash "$CLAUDE_PROJECT_DIR/..."` for windows, `"$CLAUDE_PROJECT_DIR/..."` otherwise) with a `timeout` field. |
| **Evidence** | `grep -c 'status-transition-ownership.sh' internal/template/templates/.claude/settings.json.tmpl` returns ≥ 1 AND `grep -n -B3 'status-transition-ownership.sh' internal/template/templates/.claude/settings.json.tmpl \| grep -c 'CLAUDE_PROJECT_DIR'` returns ≥ 1. |
| **Pass criterion** | Both grep counts ≥ 1 AND the entry sits inside the `"PostToolUse"` array (verify via `grep -n '"PostToolUse"\|status-transition-ownership' internal/template/templates/.claude/settings.json.tmpl` showing status-transition line-number after the PostToolUse line and before the next event key). |

---

## AC-HDW-002 — sync-gate language-auto-detect (graceful skip + silent pass)

| Field | Value |
|-------|-------|
| **Given** | M1 is complete and `.claude/hooks/moai/sync-phase-quality-gate.sh` has been language-generalized, exposing `detect_language()` as a directly-invocable shell function (per design.md §1.2 + plan.md M1 — the function MUST be source-able / callable without first passing the sync-phase-commit git gate). |
| **When** | (a) `detect_language()` is invoked directly against a fixture dir containing only `package.json` (no go.mod); (b) the FULL hook runs in a **real git repo** whose HEAD is a sync-phase commit with a non-`.go` file delta and a `package.json` present (no go.mod); (c) the FULL hook runs in a real git repo whose HEAD is a sync-phase commit but with NO recognized language marker. |
| **Then** | (a) `detect_language()` returns `node` (NOT `go`); (b) the hook reaches language detection (does NOT short-circuit at the git gate), classifies the project as `node`, does NOT invoke `go vet`/`go test`, and exits 0; (c) the hook reaches detection, finds no marker, silent-passes (exit 0, no toolchain). |
| **Evidence** | (a) directly-invocable: `D=$(mktemp -d); echo '{"name":"x"}' > "$D/package.json"; ( cd "$D" && source "$OLDPWD/.claude/hooks/moai/sync-phase-quality-gate.sh" 2>/dev/null; detect_language "$D" )` prints `node`. (b) real-git-repo reaching detection: ```G=$(mktemp -d); ( cd "$G" && git init -q && echo '{"name":"x"}' > package.json && echo 'doc' > README.md && git add -A && git -c user.email=a@b.c -c user.name=t commit -q -m 'docs(SPEC-X): sync-phase' && CLAUDE_PROJECT_DIR="$G" bash "$OLDPWD/.claude/hooks/moai/sync-phase-quality-gate.sh" > out.json 2>/dev/null; echo "exit:$?"; grep -c '"go vet"' out.json )``` returns `exit:0` AND the `"go vet"` grep count is `0` (Go toolchain NOT invoked) — OR the emitted JSON carries a language field == `node`. (c) silent pass: same harness but with NO marker file → `exit:0` and JSON decision `skip`/`no recognized language marker`. |
| **Pass criterion** | (a) `detect_language` returns `node`; (b) the real-git-repo run reaches detection (`out.json` is the language-branch JSON, NOT the `"not a sync-phase commit"` short-circuit JSON) AND `"go vet"` is NOT present in output AND exit 0; (c) silent-pass exit 0 with a no-marker decision. ALSO (static, currently RED, non-vacuous): `grep -c 'command -v' .claude/hooks/moai/sync-phase-quality-gate.sh` returns ≥ 1 (graceful-skip guard) AND `grep -c 'package.json\|pyproject.toml\|requirements.txt\|Cargo.toml' .claude/hooks/moai/sync-phase-quality-gate.sh` returns ≥ 3 (marker detection for Node/Python/Rust). **Why the prior version was vacuous (D1)**: a `mktemp -d` is NOT a git repo, so the hook's `git log -1` gate (lines 23-32) short-circuits to `exit 0` ("not a sync-phase commit") BEFORE reaching `detect_language()` — the old exit:0 proved nothing about language-neutrality. This version forces a real git repo with a sync-phase HEAD commit (and adds a direct `detect_language()` unit path) so detection is actually exercised. |

---

## AC-HDW-003 — sync-gate registered in Stop

| Field | Value |
|-------|-------|
| **Given** | M2 is complete and settings.json.tmpl has been edited. |
| **When** | An auditor inspects the Stop hooks array of settings.json.tmpl. |
| **Then** | The array contains a command entry pointing at `sync-phase-quality-gate.sh`, coexisting with the existing `handle-stop.sh` and conditional `handle-harness-observe-stop.sh` entries, using the platform-conditional quoting pattern with a `timeout`. |
| **Evidence** | `grep -c 'sync-phase-quality-gate.sh' internal/template/templates/.claude/settings.json.tmpl` returns ≥ 1 AND `awk '/"Stop":/{f=1} /"SubagentStop":/{f=0} f' internal/template/templates/.claude/settings.json.tmpl \| grep -c 'sync-phase-quality-gate.sh'` returns ≥ 1 (entry is inside the Stop block, before SubagentStop). |
| **Pass criterion** | Both counts ≥ 1 AND `awk '/"Stop":/{f=1} /"SubagentStop":/{f=0} f' internal/template/templates/.claude/settings.json.tmpl \| grep -c 'handle-stop.sh'` still returns ≥ 1 (coexistence — handle-stop.sh not displaced). |

---

## AC-HDW-004 — team-ac-verify NOT registered + file preserved

| Field | Value |
|-------|-------|
| **Given** | M2 is complete. |
| **When** | An auditor greps settings.json.tmpl for `team-ac-verify` and checks the hook file on disk. |
| **Then** | settings.json.tmpl contains zero references to `team-ac-verify.sh` AND the hook file still exists in both local and template mirror. |
| **Evidence** | `grep -c 'team-ac-verify' internal/template/templates/.claude/settings.json.tmpl` returns `0` AND `test -f .claude/hooks/moai/team-ac-verify.sh && test -f internal/template/templates/.claude/hooks/moai/team-ac-verify.sh && echo present` prints `present`. |
| **Pass criterion** | grep count is exactly `0` AND both `test -f` checks succeed (`present` printed). |

---

## AC-HDW-005 — existing handlers coexist (no regression)

| Field | Value |
|-------|-------|
| **Given** | M2 is complete. |
| **When** | An auditor verifies the pre-existing hook wrapper entries remain in settings.json.tmpl. |
| **Then** | `handle-post-tool.sh` (PostToolUse), `handle-stop.sh` (Stop), and `handle-task-completed.sh` (TaskCompleted) entries are all still present and unmodified. |
| **Evidence** | `for h in handle-post-tool.sh handle-stop.sh handle-task-completed.sh; do echo "$h: $(grep -c "$h" internal/template/templates/.claude/settings.json.tmpl)"; done` shows each count ≥ 1. |
| **Pass criterion** | All three counts ≥ 1 AND the settings.json.tmpl still renders as valid JSON for a non-windows context (`moai init` into a temp dir succeeds and the produced `.claude/settings.json` parses with `jq . >/dev/null`). |

---

## AC-HDW-006 — template-neutrality CI guard passes (internal-content leak ONLY; NOT a Go-bias guard)

> **Scope clarification (D3)**: the neutrality CI guard catches **internal-content leaks only** (its detection classes are C1 macOS-bias-path / C2 V3R-sigil / C4 feedback-memory-ref / C5 CLAUDE.local-ref / C6 PR#-ref / C7 SHA / C8 GOOS-preserve / spec-id-date — empirically confirmed from the test's own subtest names). There is **NO language-bias / Go-bias detection class**. Proof: the current (un-generalized) `sync-phase-quality-gate.sh` hardcodes `go vet` 4× and the neutrality guard passes GREEN today. Therefore passing this AC does NOT prove 16-language neutrality — Go-bias / language-neutrality is verified separately by AC-HDW-002 (static marker-detection grep + runtime non-Go fixture) and AC-HDW-009 (case-block boundedness), NOT by this CI guard.

| Field | Value |
|-------|-------|
| **Given** | M1 and M2 are complete (sync-gate generalized, settings.json.tmpl edited). |
| **When** | The template-neutrality CI guard tests run. |
| **Then** | The guard finds zero internal-content leak (internal SPEC-ID / internal REQ-AC token / internal path / V3R-sigil / feedback-memory ref / CLAUDE.local ref / PR# / SHA) in the edited template assets, and the named tests pass. |
| **Evidence** | `go test ./internal/template/ -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak' -v -count=1 2>&1 \| grep -E '^(--- PASS\|--- FAIL\|ok\|no tests to run)'` shows `--- PASS` for BOTH `TestTemplateNoInternalContentLeak` and `TestTemplateNeutralityAudit` and does NOT print `no tests to run`. AND a targeted scan: `grep -nE 'SPEC-(V3R6\|AGENCY\|WORKTREE)-\|(REQ\|AC)-(ATR\|WO\|COORD\|UNP\|LNC\|TII)-[0-9]{3}' internal/template/templates/.claude/settings.json.tmpl internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` returns no matches. |
| **Pass criterion** | The bare-pipe `-run` alternation EXECUTES both named tests (verified via `-v` output showing two `--- PASS` lines and ZERO `no tests to run`) and both pass AND the targeted grep returns no matches (exit 1 from grep = no match = pass). **Vacuity note (D2)**: the escaped-pipe form `-run 'TestX\|TestY'` matches a LITERAL `\|` (RE2 escapes the pipe) and yields `[no tests to run]` exiting 0 vacuously — this AC mandates the bare-pipe alternation `-run 'TestX|TestY'` and the verified function name `TestTemplateNoInternalContentLeak` (present at `internal/template/internal_content_leak_test.go:283`; `TestTemplateNeutralityAudit` at `template_neutrality_audit_test.go:321`). Single-package path `./internal/template/` (NOT `./internal/template/...`) confines the run to that package. |

---

## AC-HDW-007 — local↔template parity + dev-intent keys untouched

| Field | Value |
|-------|-------|
| **Given** | M2 is complete and the implementer has **manually mirrored** the template sync-gate into the local `.claude/` tree (per CLAUDE.local.md §2: `make build` regenerates the EMBEDDED template `go:embed → embedded.go`, NOT the local `.claude/` working tree; local sync is `moai update` or a manual copy). |
| **When** | An auditor diffs the local and template sync-gate hooks and inspects the local settings.json dev-intent keys. |
| **Then** | The local `.claude/hooks/moai/sync-phase-quality-gate.sh` and `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` are byte-identical (because the implementer manually mirrored them — NOT because make build produced them); the wiring entries exist in both the local `.claude/settings.json` and the template tmpl; and the dev-intent keys (`defaultMode`, `env.PATH`, `teammateMode` if present) are unchanged. |
| **Evidence** | `diff -q .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh; echo "diff:$?"` prints `diff:0`. AND `grep -c 'status-transition-ownership.sh\|sync-phase-quality-gate.sh' .claude/settings.json` returns ≥ 2. AND `git diff .claude/settings.json \| grep -E '^[-+].*"defaultMode"\|^[-+].*"teammateMode"\|^[-+].*"PATH"'` returns no lines (dev-intent keys not in the diff). |
| **Pass criterion** | `diff:0` (byte-identity via manual mirror) AND local settings grep ≥ 2 AND the dev-intent git-diff grep returns no added/removed lines for those keys. **Note (D5)**: do NOT assert "after `make build` the local and template are byte-identical" — `make build` does not touch the local `.claude/` tree; byte-identity is achieved by the implementer's explicit manual mirror step (M2). |

---

## AC-HDW-008 — advisory-only (warn-first, no exit 2 in wired default)

| Field | Value |
|-------|-------|
| **Given** | M1+M2 complete. |
| **When** | The wired `sync-phase-quality-gate.sh` is invoked on a passing sync-phase scenario in the default (no `--skip-hook`) configuration. |
| **Then** | The hook does not block the Stop event by exiting 2 under the advisory/warn-first configuration shipped by this SPEC; status-transition-ownership.sh likewise exits 0 on every non-`--skip-hook` path. |
| **Evidence** | status-transition: `echo '{"tool_name":"Edit","tool_input":{"file_path":"/tmp/x/.moai/specs/SPEC-FOO-001/spec.md"}}' \| bash .claude/hooks/moai/status-transition-ownership.sh >/dev/null 2>&1; echo "exit:$?"` returns `exit:0`. sync-gate advisory: the script's blocking `exit 2` path is gated such that, per this SPEC's warn-first config, the registered usage does not enable it — verify the spec.md §Exclusions item 1 (exit-2 deferred) is documented AND `grep -c 'exit 2' .claude/hooks/moai/sync-phase-quality-gate.sh` is reported (presence of the dormant path is acceptable; what matters is it is not enabled by the wired default — assert via a passing-scenario invocation returning exit 0). |
| **Pass criterion** | status-transition returns `exit:0` on the SPEC-artifact Edit payload AND a passing-scenario sync-gate invocation returns exit 0 (advisory) AND spec.md §Exclusions item 1 documents the exit-2 deferral. |

---

## AC-HDW-009 — Go-bias boundedness (the REAL automated language-neutrality guard)

> **Why this AC exists (D3)**: the template-neutrality CI guard (AC-HDW-006) has NO Go-bias detection class, so it cannot prove 16-language neutrality. This AC supplies the missing automated guard for the central technical risk: it asserts Go-specific tokens are confined to the Go case branch, so a non-Go project never executes Go tooling. This is a structural (case-block-bounded) assertion, complementary to AC-HDW-002's runtime fixture.

| Field | Value |
|-------|-------|
| **Given** | M1 is complete — `sync-phase-quality-gate.sh` has been generalized into per-language `case` branches (per design.md §1.2). |
| **When** | An auditor extracts the Go case branch (the block from `go)` or `"go")` to its closing `;;`) and scans for Go-specific tokens both inside and outside that block. |
| **Then** | Every occurrence of `go vet`, `golangci-lint`, and the `go test` invocation token appears ONLY inside the Go case branch; none appears in the common/default path. (The `go.mod` marker string is the one permitted exception OUTSIDE the case block — it legitimately appears inside `detect_language()` as a detection marker, equal-weight alongside `package.json` / `pyproject.toml` / `Cargo.toml`.) |
| **Evidence** | `awk '/^[[:space:]]*("?go"?\))/{inblk=1} inblk{print} /;;/{if(inblk)inblk=0}' .claude/hooks/moai/sync-phase-quality-gate.sh > /tmp/goblk.txt`; then `TOTAL=$(grep -cE 'go vet\|golangci-lint\|go test ' .claude/hooks/moai/sync-phase-quality-gate.sh); INBLK=$(grep -cE 'go vet\|golangci-lint\|go test ' /tmp/goblk.txt); echo "total:$TOTAL inblk:$INBLK"` — the `total` and `inblk` counts MUST be EQUAL (every Go-tool token is inside the Go block). AND `detect_language` must weight `go.mod` equally with the other 3 markers (`grep -c 'go.mod\|package.json\|pyproject.toml\|requirements.txt\|Cargo.toml' .claude/hooks/moai/sync-phase-quality-gate.sh` ≥ 4, no single marker privileged as "PRIMARY"). |
| **Pass criterion** | `total == inblk` for the Go-tool token set (zero Go-tool tokens outside the Go case branch) AND the 4 language markers are present with equal weighting (no "PRIMARY"/"default-language" comment privileging Go). This AC is RED before M1 implementation (the current un-generalized hook has `go vet` 4× outside any case branch). |

---

## D.6 Definition of Done

- [ ] AC-HDW-001..009 모두 PASS (각 evidence 명령 재현 가능)
- [ ] `go test ./...` GREEN (전체 회귀 없음)
- [ ] neutrality CI guard GREEN (bare-pipe `-run`, 두 named test 모두 `--- PASS`, `no tests to run` 미출력 — AC-HDW-006)
- [ ] Go-bias boundedness: Go-tool 토큰 전부 Go case branch 내부 (`total == inblk` — AC-HDW-009)
- [ ] settings.json.tmpl이 유효 JSON으로 렌더 (`moai init` smoke)
- [ ] local↔template parity (diff 0, **manual mirror** — make build 아님 — AC-HDW-007)
- [ ] dev-intent 키 미접촉 (git diff 검증)
- [ ] `team-ac-verify.sh` 파일 보존 + 미등록
- [ ] exit-2 blocking 경로 미활성화 (warn-first)

## D.7 Edge Cases

1. **Windows 렌더링**: `.Platform == "windows"`일 때 `bash "$CLAUDE_PROJECT_DIR/..."` 형태로 두 신규 엔트리가 렌더되어야 함 (기존 패턴 미러).
2. **jq 부재**: status-transition-ownership.sh는 jq 부재 시 이미 graceful no-op(allow). 회귀 없음 확인.
3. **언어 마커 복수 존재** (예: go.mod + package.json 모노레포): sync-gate는 우선순위 첫 매치(go.mod → Go)를 따르거나, 감지된 모든 언어의 toolchain을 순차 실행하는 결정을 design.md에 명시 — AC는 "최소한 감지된 첫 언어를 처리하고 비매칭 도구는 skip"으로 검증.
4. **markdown-only sync 커밋**: 기존 Go-delta skip 로직은 언어 일반화 후에도 "코드 파일 델타 0이면 skip" 형태로 유지.
