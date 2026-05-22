---
id: SPEC-V3R6-CODE-COMMENTS-EN-001
title: "Design — Mass migration of Korean comments to English"
version: "0.2.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: Medium
phase: "v3.0.0"
module: "internal/"
lifecycle: spec-anchored
tags: "code-quality, comments, internationalization, en-migration, mass-migration"
tier: L
type: design
---

# Design — SPEC-V3R6-CODE-COMMENTS-EN-001

## 1. Design Goals

본 SPEC의 설계 목표:

1. **Semantic preservation**: 의미 손실 없이 한국어 주석 → 영어 번역
2. **Identifier integrity**: SPEC-ID, REQ-ID, code identifier verbatim 보존
3. **Byte-identity of non-comment code**: 코드 구문, 들여쓰기, string literal 변경 0
4. **Validation reproducibility**: 결과를 grep + count로 binary 검증
5. **Wave-level rollback safety**: 각 Wave PR independent → revert 용이

---

## 2. Translation Methodology

### 2.1 Translation Granularity

**Per-comment block** (NOT per-file, NOT per-line):

- 1 logical comment block (continuous `//` lines OR 1 `/* */` block) = 1 translation unit
- Multi-line `//` 연속체는 sentence-aware translation (의미 단위 보존)

**Example unit**:

```go
// LSPFanInCounter는 powernap LSP 클라이언트를 사용하여 fan-in을 계산하는 구현체입니다.
// LSP 사용 불가 시 TextualFanInCounter로 fallback합니다 (REQ-SPC-004-020).
// @MX:ANCHOR: [AUTO] LSPFanInCounter — FanInCounter 인터페이스의 LSP 구현체
// @MX:REASON: fan_in >= 3 — Resolver.Resolve(), CLI mx_query.go, M6 sweep test 모두 이 구현체를 사용
```

→ Single translation unit (4 lines, semantically cohesive).

### 2.2 Translation Heuristics

| Pattern | Korean | English | Note |
|---------|--------|---------|------|
| Type description | `<Name>는 ~ 입니다.` | `<Name> is ~.` | Godoc convention |
| Function description | `~을 ~합니다.` | `~ does ~.` / `Returns ~.` | Active voice |
| State assertion | `~ 상태에서 ~ 반환합니다.` | `Returns ~ when in ~ state.` | Conditional |
| Fallback note | `~ 불가 시 ~ fallback` | `Falls back to ~ when ~ unavailable` | Standard |
| Reference | `(REQ-XXX)` | `(REQ-XXX)` | **VERBATIM** |
| @MX:NOTE | `@MX:NOTE [AUTO] X — Y` | `@MX:NOTE [AUTO] X — Y` (Y English) | Tag verbatim, desc English |
| @MX:REASON | `@MX:REASON: fan_in >= N — ...` | `@MX:REASON: fan_in >= N — ...` (desc English) | Counter verbatim |
| Inline TODO | `// TODO: 구현 필요` | `// TODO: implementation needed` | |
| Korean-only comment | `// 한국어 설명` | `// English description` | Direct semantic |
| Mixed identifier | `// SPEC-V3R6-X 처리` | `// SPEC-V3R6-X handling` | Identifier preserved |

### 2.3 Identifier Preservation Rules (REQ-CCE-004)

**MUST preserve verbatim**:

- `SPEC-[A-Z0-9-]+` (예: `SPEC-V3R6-CODE-COMMENTS-EN-001`)
- `REQ-[A-Z]+-\d+` (예: `REQ-CCE-001`)
- `AC-[A-Z]+-\d+` (예: `AC-CCE-001`)
- `MEMO-[A-Z0-9-]+` (예: `MEMO-V3R5-001`)
- `EXCL-[A-Z]+-\d+` (예: `EXCL-CCE-001`)
- Go identifier: `funcName`, `TypeName`, `varName`, `ConstName`
- Error codes: `EPERM`, `ENOENT`, `EOF`, etc.
- Library names: `cobra`, `viper`, `gh`, `tmux`, etc.
- Sentinel keys: `FROZEN_SENTINEL`, `HARNESS_FROZEN`, `MODE_PIPELINE_ONLY_UTILITY`, etc.

### 2.4 Technical Term Preservation (REQ-CCE-006)

**MAY preserve verbatim within English context**:

- Go keywords: `goroutine`, `defer`, `select`, `chan`, `interface`, `struct`
- Concepts: `mutex`, `atomic`, `context cancellation`, `fan-in`, `fan-out`
- Build tags: `//go:build !windows`, `//go:build windows`
- Linter directives: `//nolint`, `//revive:disable`

**Translation pattern** for technical comments:

```go
// Before (Korean):
// goroutine 누수 방지를 위해 context 취소 시 채널 닫음

// After (English):
// Close the channel on context cancellation to prevent goroutine leak
```

### 2.5 @MX Tag Translation Rules (REQ-CCE-002)

**Tag prefix verbatim**:

| Tag | Format |
|-----|--------|
| `@MX:NOTE` | `@MX:NOTE: [AUTO] <Name> — <Description in English>` |
| `@MX:WARN` | `@MX:WARN <Description in English>` (must include reason via @MX:REASON) |
| `@MX:REASON` | `@MX:REASON: <criterion> — <English explanation>` |
| `@MX:ANCHOR` | `@MX:ANCHOR: [AUTO] <Name> — <English description>` |
| `@MX:TODO` | `@MX:TODO: <English description>` |

**Examples**:

```go
// Before:
// @MX:NOTE: [AUTO] LSPReferencesClient — core.Client 의존 없이 mx 패키지 내부에서 LSP 참조 질의를 추상화.

// After:
// @MX:NOTE: [AUTO] LSPReferencesClient — Abstracts LSP reference queries inside the mx package without depending on core.Client.
```

```go
// Before:
// @MX:REASON: fan_in >= 3 — Resolver.Resolve(), CLI mx_query.go, M6 sweep test 모두 이 구현체를 사용

// After:
// @MX:REASON: fan_in >= 3 — Resolver.Resolve(), CLI mx_query.go, and M6 sweep test all use this implementation
```

### 2.6 Godoc Convention (REQ-CCE-003)

Go documentation tool 인식 패턴 준수:

```go
// FunctionName does X.
// It returns Y when Z.
func FunctionName() Y { ... }

// TypeName represents X.
//
// Fields:
//   - Field1: description
//   - Field2: description
type TypeName struct { ... }

// VariableName is the X for Y.
var VariableName = ...
```

---

## 3. Agent-based Batch Strategy

### 3.1 Single-Agent Sequential (Solo Mode)

**Use case**: Wave 1 (small, foundation), Wave 4 (small cleanup), Wave 7 (final cleanup)

```
manager-develop (Tier L Section A-E delegation)
  └─ Per file:
       Read → Identify spans → Translate → Edit (or MultiEdit)
  └─ Per Wave post-verification:
       AC matrix + cross-platform build + string preservation check
  └─ Commit (Conventional Commits) + Push + PR creation
```

### 3.2 Agent Teams 5+1+1 (Parallel Mode)

**Use case**: Wave 2 (CLI, 25 files), Wave 3 (Harness+Migration, 23 files), Wave 5 (Test A, 50 files), Wave 6 (Test B, 38 files)

per `.claude/rules/moai/workflow/agent-teams-pattern.md`:

```
LEADER (orchestrator, this session)
  ├─ Spawn reviewer (Phase 1, read-only, no isolation)
  ├─ Spawn implementer-1..5 (Phase 2, parallel, isolation: worktree NOT used per 2026-05-22 policy)
  │    ├─ implementer-1: internal/<wave_pkg_A>/**
  │    ├─ implementer-2: internal/<wave_pkg_B>/**
  │    ├─ implementer-3: internal/<wave_pkg_C>/**
  │    ├─ implementer-4: internal/<wave_pkg_D>/**
  │    └─ implementer-5: internal/<wave_pkg_E>/**
  └─ Spawn tester (Phase 3, after first implementer commits, no test file 충돌 책임)
```

**File ownership map (test-files Wave 5)**:

```
implementer-1 → internal/cli/auth/**/*_test.go      (~8 files)
implementer-2 → internal/cli/spec/**/*_test.go      (~8 files)
implementer-3 → internal/cli/worktree/**/*_test.go  (~8 files)
implementer-4 → internal/cli/<rest>/**/*_test.go    (~8 files)
implementer-5 → internal/template/**/*_test.go      (~10 files)
tester        → (no write — test files are owned by implementers in this SPEC since we are translating, not adding)
reviewer      → (read-only, semantic preservation review)
```

**Note**: 본 SPEC은 _adding_ tests가 아닌 _translating comments in_ test files이므로, tester role의 역할은 reviewer로 흡수. Implementer가 test 파일 자체를 영어화.

### 3.3 Activation Conditions

Agent Teams 5+1+1 use case is **OPTIONAL** — sequential solo mode가 default. Activation 조건:

1. `.moai/config/sections/workflow.yaml` `team.enabled: true` 확인
2. `.claude/settings.json` env `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` 확인
3. Wave files >= 25 (Wave 2/3/5/6 해당)
4. Wall-time 단축 필요성 사용자 확인

**Fallback**: Solo mode (single manager-develop sequential)

---

## 4. Validation Methodology

### 4.1 Per-File Validation

After each file edit:

```bash
gofmt -l <file>           # MUST be empty (no formatting drift)
go vet <file>             # No new vet errors
grep -c '[가-힣]' <file>  # Korean count for verification (in-scope spans translated)
```

### 4.2 Per-Wave Validation

per [acceptance.md](./acceptance.md) §6.2 wave-N-verify.sh script:

- AC-CCE-001/002/003 grep 0 matches (Wave scope)
- AC-CCE-005/006 cross-platform build PASS
- AC-CCE-012 diff scope `*.go` only

### 4.3 Final Validation (Wave 7 완료 후)

per [acceptance.md](./acceptance.md) §4 Definition of Done — 7 items.

### 4.4 Stash-Test Pattern for Baseline Comparison

EXCL-CCE-008 baseline 잔존 검증:

```bash
# Pre-Wave baseline capture
git stash --include-untracked -m "wave-N-baseline-$(date -u +%Y%m%dT%H%M%SZ)"
go test ./... 2>&1 | tee /tmp/baseline-test.log
git stash pop

# Post-Wave test (after translations applied)
go test ./... 2>&1 | tee /tmp/post-test.log

# Compare FAIL counts
BASELINE_FAILS=$(grep -cE "^--- FAIL|^FAIL" /tmp/baseline-test.log)
POST_FAILS=$(grep -cE "^--- FAIL|^FAIL" /tmp/post-test.log)
NEW_FAILS=$((POST_FAILS - BASELINE_FAILS))

echo "Baseline FAILs: $BASELINE_FAILS"
echo "Post FAILs: $POST_FAILS"
echo "NEW FAILs: $NEW_FAILS (MUST be 0)"
test "$NEW_FAILS" = "0" || { echo "Wave FAIL — NEW regression"; exit 1; }
```

---

## 5. Tooling and Tool Constraints

### 5.1 Allowed Tools (per C-CCE-001)

- **Read**: File reading
- **Edit / MultiEdit**: Comment editing (preferred)
- **Grep**: Korean span discovery + verification
- **Glob**: File listing
- **Bash**: Verification commands only (NOT for editing)

### 5.2 Prohibited Tools

- **sed / awk / perl**: Bulk regex replace (semantic loss risk, C-CCE-001)
- **Write**: Avoid for existing files (use Edit for diff-only changes)
- **Find -exec**: Avoid script-based bulk modification

### 5.3 Edit Tool Best Practice

For multi-span same-file edits, **MultiEdit** preferred over multiple Edit calls:

```
MultiEdit(file_path: ".../foo.go", edits: [
  {old_string: "// 한국어 1", new_string: "// English 1"},
  {old_string: "// 한국어 2", new_string: "// English 2"},
  ...
])
```

Atomic: all edits succeed or all fail. Reduces token usage and turn count.

---

## 6. Commit Strategy

### 6.1 Commit Granularity

**Per package (NOT per file)**:

```
feat(comments): translate Korean to English in internal/config (Wave 1)

- 4 files, ~80 Korean comment lines → English
- @MX:NOTE/REASON descriptions translated
- Identifier verbatim preserved (REQ-CCE-004)
- String literals byte-identical (REQ-CCE-005)
- go build + cross-platform PASS
- Lint NEW=0, test baseline preserved

Refs: SPEC-V3R6-CODE-COMMENTS-EN-001, REQ-CCE-001/002/003/004/005/008
```

### 6.2 Per-Wave Aggregation

각 Wave는 1-3 commits로 구성:

- Commit 1: Translation 적용 (대부분 files)
- (Optional) Commit 2: Verification fixes (test baseline 보정 등)
- Commit 3 (chore): `progress.md` 갱신 + spec.md HISTORY 추가

### 6.3 Conventional Commits Format

```
<type>(<scope>): <description>

[body]

[footer with SPEC ref]
```

**Type for this SPEC**: `feat(comments)` 또는 `chore(comments)` (no functional change → `chore` 권장).

**Final decision**: `chore(comments)` since non-functional (comments only).

```
chore(comments): translate Korean to English in internal/config (Wave 1)

Wave 1 of SPEC-V3R6-CODE-COMMENTS-EN-001 — Foundation packages (config/core/hook/spec, 9 files).

Per-Wave verification:
- AC-CCE-001/002/003 grep 0 matches in Wave scope
- AC-CCE-005/006 cross-platform build PASS
- AC-CCE-007 test baseline preserved (3 baseline FAILs from HARNESS-RENAME-001 — see EXCL-CCE-008)
- AC-CCE-008 lint NEW=0
- AC-CCE-011 identifier count preserved
- AC-CCE-012 diff scope: 9 *.go files only

Refs: SPEC-V3R6-CODE-COMMENTS-EN-001, REQ-CCE-001..008
Wave: 1/7
```

### 6.4 No `--no-verify` / `--amend`

Per CLAUDE.md Git Safety Protocol:

- ❌ NEVER skip pre-commit hooks (`--no-verify`)
- ❌ NEVER amend prior commits (`--amend`)
- ✅ Always create NEW commits

---

## 7. Rollback Strategy

### 7.1 Per-File Rollback (during Wave)

```bash
git checkout HEAD -- <file>  # Revert single file
```

### 7.2 Per-Wave Rollback (PR-level)

```bash
gh pr close <PR-number>      # Close PR without merge
git branch -D feat/SPEC-V3R6-CODE-COMMENTS-EN-001-wave-N
```

### 7.3 Post-Merge Rollback (catastrophic)

```bash
gh pr create --title "revert: Wave N from SPEC-V3R6-CODE-COMMENTS-EN-001" \
  --body "Revert due to <reason>" \
  --base main
# Manually create revert PR with git revert <merge-commit>
```

### 7.4 No-Op Recovery

Translation only — no schema/API changes. Rollback impact: **comments revert to Korean**. No code behavior change.

---

## 8. Integration Points

### 8.1 With CLAUDE.md §5 MX Tag Integration

본 SPEC은 MX Tag protocol과 **complementary** — `@MX:` tag prefix는 protocol 정의 그대로 보존, descriptions만 영어화.

### 8.2 With `.moai/config/sections/language.yaml`

```yaml
language:
    code_comments: en          # ← 이 정책의 강제 마이그레이션
    documentation: ko          # ← SPEC 본문은 한국어 유지
    conversation_language: ko
```

### 8.3 With Template-First Rule (CLAUDE.local.md §2)

`internal/template/templates/.claude/` mirror는 **이미 영어 정책** (별도 마이그레이션 불필요). 본 SPEC은 source code (`internal/`, `cmd/`, `pkg/`) only.

### 8.4 With CI Pipeline

CI quality gates (`scripts/ci-watch/run.sh`):

- `Test (ubuntu/macos/windows)` — AC-CCE-005/006/007 검증
- `Lint` — AC-CCE-008 검증
- `Build (linux/darwin/windows-amd64/...)` — AC-CCE-006 검증
- `CodeQL` — 영향 없음 (comment-only change)

Each Wave PR triggers these gates. 본 SPEC introduced FAILs는 baseline 잔존 외 0건이어야 한다.

---

## 9. Anti-Patterns to Avoid

### 9.1 Bulk sed/awk Replace

```bash
# ❌ FORBIDDEN
sed -i 's/한국어/English/g' internal/**/*.go

# Why: semantic loss, identifier corruption, multi-line break, no review
```

### 9.2 ML 자동번역 단독 사용

```
# ❌ AVOIDED
# DeepL API → bulk translate all Korean strings
# Why: 코드 컨텍스트 무지, 식별자 손상, 문법 오류 도입
```

본 SPEC은 **Agent (LLM with code context)** 기반 — DeepL/Google Translate 단독 사용 금지.

### 9.3 Single Mega-Commit

```
# ❌ AVOID
# All 267 files translated in a single commit

# Why:
# - Review burden (10K+ line diff)
# - Rollback impossible (single revert undoes everything)
# - Merge conflict cascade
```

본 SPEC은 **7-wave 분할** — Wave당 commit 1-3개.

### 9.4 String Literal Modification

```go
// ❌ FORBIDDEN
fmt.Println("한국어 메시지")  // → fmt.Println("English message")
// Why: 사용자 노출 메시지일 수 있음 (EXCL-CCE-001)
```

본 SPEC은 **string literal 보존** (REQ-CCE-005, AC-CCE-004).

---

## 10. Cross-references

- [spec.md](./spec.md) — Requirements + scope
- [acceptance.md](./acceptance.md) — Binary AC matrix
- [plan.md](./plan.md) — 7-wave execution plan
- [research.md](./research.md) — Codebase Korean inventory
- `CLAUDE.md` §5 — MX Tag Integration protocol
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — MX tag canonical
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E template
- `.claude/rules/moai/workflow/agent-teams-pattern.md` — 5+1+1 parallel
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — Read-only batch

---

Version: 0.2.0
Status: draft
Approach: Agent-based per-file batch translation, 7-wave split, Section A-E delegation MANDATORY
Cobra exception (EXCL-CCE-001 exception): N=14 entries handled in Wave 2 per OQ-CCE-001 Option B
