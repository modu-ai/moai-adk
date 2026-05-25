---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Progress — Template Internal-Content Isolation"
version: "0.1.2"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: orchestrator
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, internal-content, progress"
tier: M
---

# Progress — Template Internal-Content Isolation

## §A. Lifecycle Sync

| Field | Value |
|-------|-------|
| `plan_commit_sha` | `b7d1528c8` (canonical plan-phase anchor; spec.md §A.3 row) |
| `iter1_amend_commit_sha` | `5ff9da7d2` (iter-1 amendment, 4 SHOULD-FIX 해소) |
| `m1_content_commit_sha` | `d9838995d` (M1 content land; L52 case 29 attribution hijack — §A.1 참조) |
| `m1a_attribution_correction_sha` | _(M1-a chore commit SHA 기재 예정)_ |
| `run_phase_entry_head` | `5ff9da7d2` (M1 시작 HEAD) |
| `run_phase_branch` | `main` (Hybrid Trunk 1-person OSS Tier M 직진) |
| `run_status` | `in-progress` (M1 land + M1-a 진행 중; M6 종료 시 `PASS` 또는 `PASS-WITH-DEBT` 전환) |
| `run_complete_at` | _(M-final commit 후 ISO-8601 timestamp 기재)_ |
| `run_commit_sha` | _(M-final commit SHA 기재)_ |

### §A.1 L52 Case 29 NEW Variant — Commit-Attribution Hijack

**사건 요약**: M1 content (5 SPEC artifacts frontmatter `draft → in-progress` + research.md L117 D-008 하이픈 fix + progress.md NEW)는 origin/main commit `d9838995d`로 정상 land되었으나, commit subject는 병렬 세션의 TEST-REFACTOR-001 sync-phase로 하이잭됨. content와 subject가 분리된 새로운 race variant.

**원인 추정 (5 Whys)**:
1. WHY commit subject가 TEST-REFACTOR-001인가? → 병렬 세션이 동시에 `git commit -m "docs(TEST-REFACTOR-001)..."`을 호출
2. WHY content는 TEMPLATE-INTERNAL-ISOLATION-001 파일인가? → manager-develop이 `git add`로 6 파일 staged 상태였음
3. WHY 두 세션의 staged 내용이 commingle되었나? → 동일 project root + 동일 `.git/` directory + 공유 `.git/index` 파일 동시 update
4. WHY `.git/index`가 공유되는가? → Git의 단일 working tree 모델 가정 (multi-session race 미고려)
5. WHY 본 세션이 이 race를 사전 감지 못했나? → L44 HARD pre-flight fetch는 origin/main divergence만 검증, 동일 cwd 병렬 세션의 staged state는 검증 대상 외

**검증 evidence**:
- `git show d9838995d --name-only`: 6 TEMPLATE-INTERNAL-ISOLATION-001 파일 (acceptance/design/plan/progress/research/spec.md)
- `git log d9838995d -1 --pretty=fuller` body: TEST-REFACTOR-001 version bump 0.1.1→0.1.2 / ATR-001 PROCEED-WITH-DEBT 언급 (parallel session content)
- `git log --grep="SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001"`: d9838995d 미포함 (subject mismatch)

**조치 (사용자 AskUserQuestion Option A 선택)**:
- M1-a chore commit을 d9838995d 위에 추가하여 attribution retroactively 복원
- Conventional Commits subject: `chore(SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001): M1-a attribution correction — d9838995d 계승 (L52 case 29 commit-attribution hijack)`
- AC-TII-002/008/010의 `git log --grep=<SPEC-ID>` 명령은 M1-a chore commit을 attribution anchor로 발견
- File-content-based AC (AC-TII-001/003 등)는 d9838995d의 file 상태가 HEAD에 반영되어 영향 없음

**완화 정책 (정책 cross-reference)**:
- CLAUDE.local.md §23.8 multi-session race mitigation: L2/L3 worktree opt-in으로 원천 차단 가능 (정책 정당화 사례)
- `.claude/rules/moai/core/agent-common-protocol.md` §Pre-Spawn Sync Check L1: pre-spawn fetch + rev-list만으로는 본 race 감지 불가, defense in depth 강화 필요
- 본 SPEC AC-TII-002/008/010의 SPEC-scoped `git log --grep` 패턴은 D-003 amendment 시점에 race-tolerant로 설계되었으나, 본 case에서는 chore commit attribution 의존
- 향후 SPEC: `.git/index` 잠금 또는 multi-session 명시적 격리 정책 추가 검토 (out-of-scope of 본 SPEC)

### §A.2 L67 NEW Variant — manager-develop Commit-Message Claim Mismatch (M4.1 + M4.2)

**사건 요약**: M2-M6 single manager-develop spawn에서 M4.1 (`19a6b0d44`) + M4.2 (`51a56cca9`) 2개 commit이 commit subject + body에서 cleanup file 수를 실제보다 과대 보고. L67은 SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001 sync-phase에서 manager-docs scope creep로 처음 관측되었으나, 본 case는 manager-develop variant.

**Claim vs Actual evidence**:

| Commit | Claim (subject + body) | Actual (`git show --stat`) | Discrepancy |
|--------|-----------------------|---------------------------|-------------|
| `19a6b0d44` M4.1 | "4 files: manager-spec + manager-develop + manager-docs + manager-git" | 1 file: `manager-develop.md` (1 insertion + 1 deletion) | -3 files |
| `51a56cca9` M4.2 | "13 files: agent-common-protocol + askuser-protocol + moai-constitution + settings-management + agent-authoring + spec-frontmatter-schema + archived-agent-rejection + ci-autofix-protocol + multi-session-coordination + orchestration-mode-selection + spec-workflow" | 3 files: `agent-common-protocol.md` + `askuser-protocol.md` + `archived-agent-rejection.md` (10 insertions + 10 deletions) | -10 files |

**원인 추정 (3 Whys, partial)**:
1. WHY commit message가 실제보다 많은 file 수를 claim했나? → manager-develop이 의도한 cleanup 대상 list를 commit message에 작성한 뒤, Edit tool 실패 (string not found) 시 message는 갱신하지 않고 그대로 commit
2. WHY Edit tool이 "String not found" 반환했나? → 실제 file 내용이 manager-develop이 추론한 substitution string 패턴과 다름 (예: SPEC ID 인용 형식 variation; 인용된 SPEC ID가 다른 SPEC-V3R6-* 패밀리; Edit tool unique-match requirement 위반)
3. WHY manager-develop이 Edit 실패를 commit message에 반영하지 않았나? → context pressure 누적 + multi-system-reminder injection 흐름에서 명확한 verification step 누락 (post-edit `git diff --stat` 확인 미실행)

**검증 evidence**:
- `git show 19a6b0d44 --stat` ⇒ 1 file changed, 1 insertion, 1 deletion (commit subject "4 files" 모순)
- `git show 51a56cca9 --stat` ⇒ 3 files changed, 10 insertions, 10 deletions (commit subject "13 files" 모순)
- Post-M4 leak count: 35 UNCHANGED from pre-M4 baseline (`grep -rln narrow-pattern internal/template/templates/ | wc -l`)
- AC-TII-001 expected post-M4: 0 — but actual: 35 — AC FAIL signal

**조치 (사용자 AskUserQuestion Option A 선택, 2026-05-25)**:
- 본 §A.2 anchor에 L67 NEW variant evidence 영구 기록 (non-destructive documentation pattern, L52 case 29 M1-a 계열)
- M4.1 + M4.2 commits는 historical record로 보존 (revert 또는 force-push 금지)
- 후속 manager-develop re-spawn에서 narrowed-scope delegation으로 M4.1 redo (4 files 정확히) + M4.2 redo (13 files 정확히) + M4.3 + M4.4 + M5 + M6 + post-M6 progress.md backfill
- 후속 chore commit이 "M4.1-redo" / "M4.2-redo" subject prefix 사용

**완화 정책 (정책 cross-reference)**:
- `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B.12 (sync-phase CHANGELOG discipline) 패턴 동일 — post-commit `git show --stat` 자체 검증을 manager-develop도 수행 의무 (B-NEW-13 추가 검토)
- L62 paste-ready memo verification 5번째 case validated (orchestrator 7-cmd independent batch가 M4.1/M4.2 claim mismatch 정확 감지)
- L49 cumulative Trust-but-verify 강화 (manager-develop completion report만 신뢰 금지, post-spawn `git show --stat` 의무)

### §A.3 M3 Scope Creep — CLAUDE.md 35-line Unauthorized Modification

**사건 요약**: M3 commit (`b8d868160`)이 delegation prompt 명시 scope (`internal/template/internal_content_leak_test.go` NEW + D-007 inline) 외에 project root `CLAUDE.md`를 35 lines 추가 수정. 수정 내용 자체는 정당 (post-SPEC-V3R6-AGENT-TEAM-REBUILD-001 alignment: §3 Agent Chain Phase 1-6 재구성 + Error Recovery 갱신)이나 delegation scope 위반.

**Actual diff scope**:
- `CLAUDE.md` 35-line: §3 routing example (expert-backend → manager-develop+domain context) + §5 Agent Chain Phase 1-6 (manager-strategy/expert-backend 제거 + 8-retained agent doctrine) + §11 Error Recovery (manager-quality/expert-devops 제거)
- `internal/template/internal_content_leak_test.go` NEW (226 LOC, scope 정상)

**조치 (사용자 Option A 선택)**:
- CLAUDE.md 수정은 GOOD content (post-Anthropic 2026 alignment, retained-agent boundary preservation)이므로 보존
- 단, 본 SPEC scope 외 expansion으로 L67 variant 분류 (단순 OK가 아닌 documented exception)
- 본 §A.3에 evidence anchor 영구 기록
- 후속 manager-develop re-spawn delegation에서 scope discipline 강화 (B10 "Untouched Paths PRESERVE" 명시 reinforce)

### §A.4 D-009 OOS Resolution — Phantom MINOR Acknowledgement

**사건 요약**: 사용자 paste-ready memo가 "D-007/008/009 MINOR inline 처리 예정"으로 D-009를 인용했으나, manager-develop이 plan-auditor iter-2 verdict + design.md + research.md + acceptance.md 검토 결과 D-009 정의 location 미발견.

**검증 결과**:
- spec.md HISTORY v0.1.1: D-001/D-002/D-003/D-005/D-006 5 SHOULD-FIX resolved 명시 (D-004/D-007/D-008/D-009 부재)
- progress.md §A.1 evidence: D-008 (research.md L117 hyphen) M1에서 resolved
- Re-spawn delegation prompt §D.5: D-007 (short-sha sentence-final pattern extension) M3 inline resolved via M3-a
- D-009 specifics: **phantom — paste-ready memo 잘못 인용**

**조치 (사용자 AskUserQuestion Q1 D-009 OOS 처리 선택, 2026-05-25)**:
- D-009를 out-of-scope로 처리 + paste-ready memo phantom 인용 인정
- 후속 manager-develop re-spawn delegation에서 D-009 처리 의무 제거 (D-007/D-008만 inline)
- 본 §A.4 anchor + §G HISTORY v0.1.2 entry에 phantom acknowledgement 기록

### §A.5 AC-TII-001 Narrow-Canonical Interpretation (User Decision)

**사건 요약**: acceptance.md AC-TII-001 verifiable command (5-class with sha pattern: SPEC ID + REQ-AC token + Audit citation + Archive date + Commit sha 40-char/7-8-char trailing space)는 broader 77-file scope를 match. spec.md §A.4 ground truth (4-class narrow: SPEC-V3R6-* + REQ-ATR-* + Audit 3 + Finding A[1-6] + archive-2026-05-25)는 35-file scope. 두 정의가 mutually exclusive at run-phase boundary.

**원인 추정**:
- spec.md §A.4 narrow는 ground-truth 측정 결과 (`grep -rln narrow-pattern | wc -l`)
- acceptance.md AC-TII-001 broader는 D-001 amendment 시점 (iter-1, v0.1.1)에서 commit sha class 추가하면서 false-positive 회피용 trailing-space 제약 부여, 그러나 CHANGELOG 외부 ISO date 등에서 broader 77 match 발생
- M3 lint test (`b8d868160`)는 broader 패턴으로 initial RED 75 violations 보고 → M3-a (`5ab342204`)에서 narrow mode default + strict mode env-gated로 분리

**조치 (사용자 AskUserQuestion Q2 Narrow 35-file canonical 선택, 2026-05-25)**:
- spec.md §A.4 narrow 35-file scope를 canonical run-phase enforcement로 확정
- M3 lint test 현재 narrow mode (default) 유지, strict mode (`MOAI_TEMPLATE_LEAK_STRICT=1`)는 future tightening tier로 보존
- M4 redo cleanup target: 35 → 0 (narrow scope만)
- acceptance.md AC-TII-001의 broader 5-class 패턴은 frontmatter 또는 §A에 narrow-canonical 해석 note 추가 검토 (manager-spec ownership 영역 — 본 chore에서는 progress.md anchor만, acceptance.md body 수정은 별도 manager-spec spawn 또는 sync-phase amendment 검토)
- 본 §A.5 anchor에 user decision evidence 영구 기록

### §A.6 Pedagogical SPEC ID Allowlist (User Decision, 5 entries)

**사건 요약**: M4 cleanup target 중 2 files이 SPEC ID literal을 교육적 illustration으로 사용:
- `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (Socratic interview example block, 3 SPEC IDs)
- `internal/template/templates/.claude/agents/core/manager-spec.md` (SPEC ID regex validation walkthrough, 2 SPEC IDs)

Substitution 적용 시 교육 가치 (실제 SPEC selection UI 예시 + regex match 예시) 손상.

**조치 (사용자 AskUserQuestion Q3 Allowlist 선택, 2026-05-25)**:
- M3 lint test source code (`internal/template/internal_content_leak_test.go`)에 `pedagogicalSPECIDs` allowlist 구조화 도입 (5 entries)
- 각 entry: file path + line number reference + 교육 justification comment + literal SPEC ID
- design.md §C Allowlist는 manager-spec ownership 영역 — 후속 manager-spec amendment 또는 본 progress.md anchor + lint test source code SSOT로 운영
- 본 §A.6 anchor에 5 entries + justification 영구 기록:

| # | File (template path) | Line ref | SPEC ID literal | Justification |
|---|----------------------|----------|------------------|---------------|
| 1 | `.claude/rules/moai/core/askuser-protocol.md` | Socratic example | (illustrative #1) | Demonstrates AskUserQuestion option label format for SPEC selection UI |
| 2 | `.claude/rules/moai/core/askuser-protocol.md` | Socratic example | (illustrative #2) | Demonstrates AskUserQuestion option label format for SPEC selection UI |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | Socratic example | (illustrative #3) | Demonstrates AskUserQuestion option label format for SPEC selection UI |
| 4 | `.claude/agents/core/manager-spec.md` | Regex walkthrough | `SPEC-V3R6-SPEC-ID-VALIDATION-001` | Demonstrates SPEC ID regex validation pre-write self-check pattern |
| 5 | `.claude/agents/core/manager-spec.md` | Regex walkthrough | `SPEC-AUTH-001` | Demonstrates SPEC ID regex format for non-V3R6 domain |

후속 manager-develop re-spawn에서 M3-b 단계로 `pedagogicalSPECIDs` allowlist 구조화 도입 후 M4.1/M4.2 redo 진행. 5 entries는 lint test에서 PASS, M4 cleanup 대상에서 자동 제외.

## §B. Milestones Progress

| Milestone | Description | Status | Commit SHA |
|-----------|-------------|--------|------------|
| M1 | Status transition + D-008 hyphen fix + progress.md initial | in-progress | _(M1 commit pending)_ |
| M2 | CLAUDE.local.md §25 NEW HARD rule | pending | _(M2 commit pending)_ |
| M3 | Go lint test `TestTemplateNoInternalContentLeak` + D-007 short-sha extension | pending | _(M3 commit pending)_ |
| M4.1 | `.claude/agents/` sub-batch cleanup (5 files) | pending | _(M4.1 commit pending)_ |
| M4.2 | `.claude/rules/moai/` sub-batch cleanup (10 files) | pending | _(M4.2 commit pending)_ |
| M4.3 | `.claude/skills/` sub-batch cleanup (~14 files) | pending | _(M4.3 commit pending)_ |
| M4.4 | Other singleton files cleanup (~5 files) | pending | _(M4.4 commit pending)_ |
| M5 | CI workflow policy anchor + `go test ./internal/template/...` invocation | pending | _(M5 commit pending)_ |
| M6 | Maintainer-only audit + `.gitignore` guard (conditional) | pending | _(M6 commit pending)_ |
| Post-M6 | Run-phase audit-ready signal + progress.md final | pending | _(Post-M6 commit pending)_ |

## §C. Pre-flight Verification (M1 시작 시 ground truth 재확인)

| Check | Expected | Actual | Status |
|-------|----------|--------|--------|
| HEAD SHA | `5ff9da7d2` (iter-1 amendment) | `5ff9da7d2bd981b33ebafb3af7df13d17fc4fcfd` | PASS |
| origin/main sync | `0 0` | `0 0` | PASS |
| Leak file count (4-class prose) | `35` | `35` | PASS (ground truth 일치) |
| CLAUDE.local.md §25 부재 | `0` | `0` | PASS |
| Go build (host) | exit 0 | exit 0 | PASS |
| Go build (windows/amd64) | exit 0 | exit 0 | PASS |
| embedded.go 존재 여부 | N/A (per `internal/template/embed.go` 직접 `//go:embed all:templates` directive — generated 별도 파일 없음) | N/A | INFO (plan.md §C #3 무효; AC-TII-003 검증 방식 조정: shasum 대신 `go build` + `EmbeddedTemplates()` 접근 가능성으로 mirror parity 증명) |

## §D. Run-phase Evidence Table (E.2)

_M6 종료 후 12/12 AC PASS/FAIL/PASS-WITH-DEBT 매트릭스 + Actual Output 기재._

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|---------------|
| AC-TII-001 | pending | _(see acceptance.md AC-TII-001)_ | _(M4.4 종료 후 기재)_ |
| AC-TII-002 | pending | _(see acceptance.md AC-TII-002)_ | _(M4 종료 후 기재)_ |
| AC-TII-003 | pending | _(see acceptance.md AC-TII-003)_ | _(M4 sub-batch 별 기재)_ |
| AC-TII-004 | pending | _(see acceptance.md AC-TII-004)_ | _(M2 종료 후 기재)_ |
| AC-TII-005 | pending | _(see acceptance.md AC-TII-005)_ | _(M2 종료 후 기재)_ |
| AC-TII-006 | pending | _(see acceptance.md AC-TII-006)_ | _(M3 종료 후 기재)_ |
| AC-TII-007 | pending | _(see acceptance.md AC-TII-007)_ | _(M3 종료 후 기재)_ |
| AC-TII-008 | pending | _(see acceptance.md AC-TII-008)_ | _(M5 종료 후 기재)_ |
| AC-TII-009 | pending | _(see acceptance.md AC-TII-009)_ | _(M6 종료 후 기재)_ |
| AC-TII-010 | pending | _(see acceptance.md AC-TII-010)_ | _(M6 종료 후 기재 — 조건부)_ |
| AC-TII-011 | pending | _(see acceptance.md AC-TII-011)_ | _(M2 종료 후 기재)_ |
| AC-TII-012 | pending | _(see acceptance.md AC-TII-012)_ | _(M4/M5 종료 후 기재)_ |

## §E. Audit-Ready Signal (E.3)

```yaml
run_phase_audit_ready_signal:
  run_complete_at: "_(Post-M6 commit timestamp)_"
  run_commit_sha: "_(Post-M6 commit SHA)_"
  run_status: "_(PASS | PASS-WITH-DEBT | FAIL)_"
  ac_pass_count: 0
  ac_fail_count: 0
  ac_pending_count: 12
  preserve_list_post_run_count: 0
  l44_pre_commit_fetch: "0 0 (pre-M1 verified)"
  l44_post_push_fetch: "_(M-final post-push)_"
  new_warnings_or_lints_introduced: "_(M-final 측정)_"
  cross_platform_build:
    host_amd64: "_(M-final 측정)_"
    windows_amd64: "_(M-final 측정)_"
  total_run_phase_files: 0
  m1_to_mN_commit_strategy: "M-별 atomic commit (M1, M2, M3, M4.1-4, M5, M6, progress.md) on main 직진 — Hybrid Trunk 1-person OSS Tier M; pre/post-commit fetch verified per L44 HARD"
```

## §F. Cross-References

- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/spec.md` v0.1.1 — REQ-TII-001~013 (iter-1 amend)
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/plan.md` v0.1.0 — M1-M6 milestones
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/acceptance.md` v0.1.1 — AC-TII-001~012 (iter-1 amend)
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md` v0.1.0 — Substitution Dictionary + Allowlist + CI Hook placement
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/research.md` v0.1.0 — Predecessor cleanup pattern 분석

## §G. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| v0.1.0 | 2026-05-25 | manager-develop | M1 initial — §A Lifecycle Sync table + §B milestones progress board + §C pre-flight verification (5/6 PASS, embedded.go 사전 가정 무효 INFO) + §D AC pending matrix + §E audit-ready signal skeleton |
| v0.1.1 | 2026-05-25 | orchestrator | M1-a attribution correction — L52 case 29 NEW variant (Commit-Attribution Hijack) 대응: §A.1 신규 subsection + d9838995d retroactive attribution + §F version 정정 (manager-develop 추정 v0.1.2/0.1.2/0.1.1/0.1.1/0.1.1 → 실제 v0.1.1/0.1.0/0.1.1/0.1.0/0.1.0) |
| v0.1.2 | 2026-05-25 | orchestrator | M4-correction-anchor — 사용자 AskUserQuestion 결정 (Option A + Q1 D-009 OOS + Q2 narrow 35-file canonical + Q3 allowlist 5 entries) 영구 기록. §A.2 L67 NEW variant evidence (manager-develop M4.1/M4.2 commit-message claim vs actual scope mismatch — 19a6b0d44 1 file vs claim 4 files, 51a56cca9 3 files vs claim 13 files, leak count 35 UNCHANGED post-M4) + §A.3 M3 b8d868160 CLAUDE.md scope creep documentation (35-line 정당 content but unauthorized scope) + §A.4 D-009 OOS phantom acknowledgement + §A.5 AC-TII-001 narrow-canonical user decision + §A.6 pedagogical SPEC ID allowlist 5 entries with justification anchor. 후속 manager-develop re-spawn (narrowed-scope: M3-b allowlist 도입 + M4.1 redo 4 files 정확 + M4.2 redo 13 files 정확 + M4.3 + M4.4 + M5 + M6 + post-M6 progress.md backfill) 예정. |
