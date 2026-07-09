---
description: "manager-develop 위임 Prompt Template — Tier M/L SPEC run-phase 5-section 표준. SPEC 위임 작성 시에만 로드."
paths: ".moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md"
---

# manager-develop 위임 Prompt Template

## Applicability

[ZONE:Evolvable] [HARD] The Section A-E 5-section delegation template defined in this rule is **REQUIRED for Tier M and Tier L SPEC delegations** and **OPTIONAL for Tier S** delegations. Tier S SPECs (≤300 LOC, <5 files affected, 2 artifacts per the LEAN workflow) MAY use minimal delegation prompts (~500-800 tokens) covering only:

## cycle_type Mode Reference

Per the canonical agent catalog policy, the `manager-develop` agent operates in one of three `cycle_type` modes selected per the run-phase task profile:

| cycle_type | Loop pattern | When to use | Iteration contract | Canonical reference |
|------------|--------------|-------------|---------------------|---------------------|
| `ddd` | ANALYZE-PRESERVE-IMPROVE | Existing codebases with minimal test coverage (< 10% per quality.yaml `development_mode: ddd` selection); characterization-test-first preservation of behavior | No fixed iteration limit; one cycle per logical refactoring chunk | `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase DDD Mode |
| `tdd` | RED-GREEN-REFACTOR | Default — all new development work, brownfield projects with pre-RED analysis (≥ 10% coverage per quality.yaml `development_mode: tdd` selection); test-first development | No fixed iteration limit; one cycle per behavior specification | `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase TDD Mode |
| `autofix` | **DIAGNOSE-PATCH-VERIFY** | CI auto-fix loop after `scripts/ci-watch/run.sh` detects a failing required check; semantic-failure-safe patching of lint / build / type errors | **Maximum 3 iterations** per PR push (per-PR-push counter, not per-session); escalation to `AskUserQuestion` after iteration 3 with no auto-resume timeout | `.claude/rules/moai/workflow/ci-autofix-protocol.md` |

### cycle_type=autofix DIAGNOSE-PATCH-VERIFY pattern

Each iteration of the autofix loop executes the three-step DIAGNOSE-PATCH-VERIFY pattern:

1. **DIAGNOSE**: Read the failing CI check output (provided by the orchestrator from `scripts/ci-watch/run.sh`). Identify the root cause — lint rule violation, build error, type error, missing dependency, etc.
2. **PATCH**: Apply a minimal fix that addresses the root cause without expanding scope. The autofix loop MUST NOT modify `.env`, `.env.*`, credentials files, secrets, or `scripts/ci-watch/run.sh` or any Wave 2 infrastructure scripts.
3. **VERIFY**: Re-run the failing check locally; if exit 0, push the patch as a new commit on the PR branch. If still failing, increment the iteration counter and repeat from DIAGNOSE.

### autofix escalation contract

After 3 iterations without success, the loop MUST halt and the orchestrator MUST trigger an `AskUserQuestion` blocking call presenting the user with at least: (a) continue with manual investigation, (b) revert the offending change and re-plan, (c) abort with structured failure report. Semantic failures (data race, deadlock, panic, test assertion failure) MUST NOT be auto-patched without human approval.

### Logged at

Every autofix iteration MUST be logged to `.moai/logs/ci-autofix/` with timestamp, patch summary, and CI result.



- Goal (single-paragraph task description)
- Deliverables (concrete file/commit list)
- Constraints (PRESERVE list, forbidden commands)
- Self-verification (AC PASS/FAIL matrix)

When applying the minimal form for Tier S, Section B (Known Issues B1-B12) MAY be filtered to relevant categories only or omitted entirely if no listed risk applies. Section C (Pre-flight) MAY be reduced to the single most-relevant baseline command.

When the SPEC tier is M or L, the full Section A-E template SHOULD be applied; Section B (known issues) MAY filter B1-B12 categories by domain relevance (e.g., a documentation-only SPEC may omit B1 cross-platform build tags).

Tier classification reference: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (S/M/L).

> [ZONE:Evolvable] [HARD] 모든 Tier M/L의 `manager-develop` subagent 위임 prompt는 본 템플릿의 5개 섹션 (Context / Known Issues / Pre-flight / Constraints / Self-Verification Deliverables)을 포함해야 한다. Tier S는 위 Applicability 절의 minimal form을 사용해도 무방. 누락 시 (Tier M/L에서) 재위임 반복 위험 증가.

본 rule은 메타-분석 결과에서 도출된 위임 품질 개선 사항을 표준화한다. 1-pass 위임으로 결함 사전 차단 목표.

## 1. 표준 위임 Prompt 5-Section 구조

### Section A — Context (위치 + 분기 + SPEC 산출물 경로)

명시 의무:
- 작업 위치 (project root absolute path)
- 현재 branch + HEAD SHA (manager가 추가 commit을 어디에 쌓을지 명확화)
- SPEC 산출물 경로 (`.moai/specs/SPEC-XXX/{spec,plan,acceptance,progress}.md`) + 라인 카운트
- plan-auditor verdict (PASS score + 재실행 권장 여부)
- 기존 인프라 (PRESERVE 대상 + EXTEND 대상)

### Section B — Known Issues 자동 주입 (가장 중요)

[ZONE:Evolvable] [HARD] 다음 12 카테고리의 known issues는 위임 prompt에 자동 포함되어야 한다. 누락 = 재위임 위험.

**B1. Cross-platform Build Tags**
- syscall 패키지 사용 시 build tag 강제
- 권장: `//go:build !windows` + `//go:build windows` 파일 분리
- 검증: `GOOS=windows GOARCH=amd64 go build ./...` 통과 의무

**B2. Cross-SPEC 정책 충돌 사전 스캔**
- 영향 받는 패키지의 retired/superseded SPEC 확인 (예: 이전 harness retirement)
- `grep -r "Retired\|TestHarnessRetirement\|deprecation-marker" internal/<pkg>` 실행
- 충돌 발견 시: SPEC 본문에서 reversal 명시 또는 새 SPEC scope 정의

**B3. C-HRA-008 / Subagent Boundary Discipline**
- `internal/harness/`, `internal/hook/` 등 subagent 도메인 코드에 AskUserQuestion 호출 금지
- 검증: `grep -rn 'AskUserQuestion\|mcp__askuser' <pkg> | grep -v "_test.go" | grep -v "// "` 가 0 매치
- CI guard test 의무: `<pkg>/subagent_boundary_test.go`

**B4. Frontmatter Canonical Schema**
- `created:`/`updated:`/`tags:` 사용 (snake_case alias 금지)
- 참조: `.claude/rules/moai/development/spec-frontmatter-schema.md`

**B5. CI 3-tier 인지**
- spec-lint, golangci-lint, Test (per OS) 각각 별도 fail 가능
- pre-existing baseline vs NEW defect classification

**B6. spec-lint Heading 규약**
- `## Out of Scope` (h2) 만으로는 `MissingExclusions` ERROR
- `### <X.Y> Out of Scope` (h3) sub-section 필요

**B7. observer.go / capture path resolution**
- `input.CWD` empty → `os.Getwd()` fallback 시 working dir 누수 (`internal/hook/.moai/` anomaly 원인)
- 권장: `$CLAUDE_PROJECT_DIR` 우선 사용

**B8. Working Tree Hygiene**
- runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/`) 변경 금지
- session_end의 `cleanupBogusRootDir` 의존 (`{}/`  literal directory는 cleanup 대상)
- 무관 untracked files은 commit 포함 금지 (`git add` specific path만)

**B9. Git Commit + Push 자체 수행 (Hybrid Trunk 1-person OSS)**
- manager-develop은 본 SPEC scope 내 commit + push 자체 수행 권장 (main 직진 — Hybrid Trunk 1-person OSS policy, Tier S/M)
- Conventional Commits format 의무 (`feat(SPEC-...): M{N} <subject>`)
- M별 분리 commit + 마지막 push 또는 M별 push 둘 다 허용
- `--no-verify` 사용 절대 금지 (pre-commit hook warn-only는 정상)
- 예외: (a) parallel session race 발생 시 orchestrator가 push 수행, (b) AC PASS-WITH-DEBT 상태에서 사용자 확인 필요 시 orchestrator 위임, (c) explicit blocker report 시
- 본 rule이 manager-docs에는 적용 **안 됨** — manager-docs는 /moai sync workflow에서 commit + push가 deliverable 자체

**B10. Untouched Paths PRESERVE (Scope Discipline)**
- 본 SPEC plan.md §A.5 PRESERVE list 외 working tree 변경 절대 금지
- parallel manager-develop instance 진행 중일 때 특히 주의 (다른 디렉토리 scope 손대지 말 것)
- runtime-managed files (`.moai/harness/*`, `.moai/state/*`, `.moai/cache/*`) 손대지 말 것
- 무관 SPEC 디렉토리 (다른 SPEC plan-phase artifacts) 손대지 말 것
- parallel session research/audit 산출물 (`.moai/research/*`) 손대지 말 것

**B11. AskUserQuestion 금지 (Subagent Boundary)**
- subagent는 사용자와 직접 상호작용 금지 (CLAUDE.md §8 + askuser-protocol.md §Orchestrator–Subagent Boundary)
- Blocker 발견 시 structured blocker report 반환 (orchestrator가 AskUserQuestion 수행 + re-delegate)
- Blocker report format: 4-옵션 + 각 옵션의 변경/영향/위험/ETA 명시
- free-form prose 질문 절대 금지 (response body에 "? 어떻게 진행할까요?" 패턴 금지)

**B12. Sync-phase CHANGELOG emission discipline (manager-docs only)**
- Before drafting CHANGELOG entries, `Read` every implementation file referenced in the SPEC plan.md (do NOT rely on plan.md description alone — plan-phase placeholders may diverge from final implementation).
- Before appending to `CHANGELOG.md` `[Unreleased]` section, run `grep -c '<SPEC-ID>' CHANGELOG.md` — if the count is ≥1, halt emission and return blocker report (avoid duplicate entries from parallel BATCH-SYNC sessions).
- Verify file paths claimed in CHANGELOG match actual `ls <package-path>` output before committing.
- Verify AC count in CHANGELOG matches `acceptance.md` (SSOT) — NOT `progress.md` (which may include deferred AC).
- Origin: an earlier CHANGELOG cleanup root cause analysis (BATCH-SYNC line hallucination incident).

### Section C — Pre-flight Check List (착수 전 의무 검증)

위임 받은 manager-develop가 코드 변경 전 실행:

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build 가능성 사전 확인
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. 기존 lint baseline 측정 (NEW vs pre-existing 구분 위해)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. PRESERVE 대상 파일 list 출력
ls <PRESERVE_GLOB>

# 5. 영향 패키지 retired/superseded SPEC 확인
grep -r "Retired\|TestHarnessRetirement\|superseded" internal/<target_pkg> || echo "no conflicts"
```

### Section D — Constraints (DO NOT VIOLATE)

각 위임 prompt에 explicit list:
- PRESERVE 대상 파일 enumeration (Brownfield strategy 적용 시)
- 무관 untracked/modified 파일 list (변경 금지)
- 금지 명령 (`--no-verify`, `--amend`, force-push to main, …)
- 사용 의무 명령 (Conventional Commits, `🗿 MoAI` trailer, …)
- C-HRA-008 같은 binary constraint (grep 0 매치)

### Section E — Self-Verification Deliverables

> Each E-item is reported per the verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — see `.claude/rules/moai/core/verification-claim-integrity.md` §3.

When manager-develop reports completion, it MUST include self-verification of the following items:

**E1. AC Binary PASS/FAIL Matrix**
| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-XXX-001 | PASS | `go test -run TestX ./pkg` | `PASS — ok  pkg 0.5s` |

**E2. Cross-Platform Build result**
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**E3. Coverage measurement (≥85% threshold per package)**
```
$ go test -cover ./internal/<pkg>/...
```

**E4. Subagent Boundary Grep (C-HRA-008 family)**
```
$ grep -rn 'AskUserQuestion' <pkg> | grep -v "_test.go" | grep -v "// "
(no output expected)
```

**E5. Lint Status (distinguish NEW vs baseline)**
```
$ golangci-lint run --timeout=2m
# On NEW issues, report explicitly; mark pre-existing baseline separately
```

**E6. Branch HEAD + Push state**
- List of new commit SHAs
- Result of `git push origin <branch>`

**E7. Blocker Report (if any)**
- When the delegation prompt did not specify a needed user decision, report it as a structured blocker (NEVER call AskUserQuestion)

## 2. 위임 Prompt 작성 Workflow (오케스트레이터 입장)

```
1. Section A 구성 (SPEC 산출물 경로 + 현재 git 상태)
2. Section B 자동 주입 (lessons memory에서 keyword 매칭으로 12 카테고리 select)
3. Section C 표준 pre-flight checks (위 그대로 복사)
4. Section D constraints (SPEC + PRESERVE list + working tree 상태에서 추출)
5. Section E deliverables (위 그대로 복사)
```

## 3. Anti-Patterns

본 템플릿 미준수 케이스 — 재위임 위험 증가:

- Section B 누락 → cross-platform build / cross-SPEC 충돌 사후 발견
- Section E 누락 → orchestrator가 직렬 검증 (~10분 추가 손실)
- "implement the SPEC" 같은 1-liner 위임 → manager-develop이 prompt 부족으로 가정 누락
- PRESERVE 대상 enumeration 누락 → 의도치 않은 파일 수정

## 4. 관련 Layer (메타-분석 §3)

본 rule은 Layer A (위임 Prompt 품질 향상)의 표준화. 후속 Layer:

- Layer B (병렬 위임 — Agent Teams): 별도 rule 필요
- Layer C (Background CI watch): `gh pr checks --watch` 패턴 표준화
- Layer D (검증 병렬화): orchestrator self-discipline
- Layer F (lessons 자동 capture): SubagentStop hook 확장

## 5. Verification (this rule applied to itself)

On the next SPEC delegation, measure by phase ordering (not wall-clock targets, per `agent-common-protocol.md` § Time Estimation):
- 1-pass success rate — Priority High target (baseline: 33%)
- Re-delegation count — Priority Medium target (baseline: 3)
- Overall completion — track by milestone progression, not duration

---

Version: 1.0.0
Status: Active — applies to all manager-develop delegations
