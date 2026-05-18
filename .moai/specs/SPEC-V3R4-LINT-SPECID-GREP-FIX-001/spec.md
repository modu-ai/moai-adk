---
id: SPEC-V3R4-LINT-SPECID-GREP-FIX-001
title: "walker SPEC-ID grep precision fix — substring collision 영구 차단"
version: "0.3.0"
status: completed
created: 2026-05-16
updated: 2026-05-16
author: manager-spec
priority: P1
phase: "v3.0.0 R4 — Foundation Cleanup"
module: "internal/spec/drift.go"
lifecycle: spec-anchored
tags: "v3r4, spec-lint, walker, grep-precision, substring-collision, status-git-consistency, foundation"
issue_number: null
related_specs:
  - SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
  - SPEC-V3R4-SPECLINT-DEBT-001
  - SPEC-V3R4-SPECLINT-DEBT-002
depends_on: []
breaking: false
bc_id: []
related_theme: "Foundation Cleanup — Lint walker precision"
target_release: v2.20.0-rc1
---

# SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — walker SPEC-ID grep precision fix

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.3.0   | 2026-05-16 | manager-docs | Sync-phase 완료. status `in-progress → completed`. PR #948 (run-phase) `be5e58179` 머지 후 walker word-boundary 필터 main에 영구 적용. V3R4-HARNESS-001/002/003 3건 false-positive WARNING 영구 해소 (AC-LSGF-001 본질 달성). |
| 0.2.0   | 2026-05-16 | manager-develop | Run-phase 완료. walker SPEC-ID word-boundary 매칭 구현 (Approach B: 2-pass post-filter), unit test 4건 신설 (`TestGetGitImpliedStatus_SPECIDWordBoundary` 등), regression test 1건. TDD 5-Wave cycle (RED → GREEN → REFACTOR → 통합 → CI validation) 완료. PR #948 OPEN. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. `internal/spec/drift.go` `getGitImpliedStatus` walker가 `git log --grep=<SPEC-ID>` substring 매칭을 사용해 SPEC-ID prefix 충돌 commit(특히 NAMESPACE supersede commit)을 walker first match로 채택하던 결함을 영구 차단. AC binary 0 (`moai spec lint --strict` 0 ERROR / 0 WARNING) 목표. PR #944 (SPEC-V3R4-HARNESS-NAMESPACE-001 plan-PR) sync drift 사건의 follow-up. |

---

## 1. Goal

`internal/spec/drift.go` `getGitImpliedStatus` walker가 SPEC-ID를 git log에서 검색할 때 사용하는 `--grep=<SPEC-ID>` substring 매칭을 **단어 경계 (word-boundary) 정밀 매칭**으로 격상하여, `SPEC-V3R4-HARNESS-001` 검색이 `SPEC-V3R4-HARNESS-NAMESPACE-001` 같이 prefix가 겹치는 다른 SPEC commit을 walker first match로 채택하는 false-positive `StatusGitConsistency` warning을 영구 차단한다.

### 1.1 배경

**Trigger 사건**: 2026-05-16 sync-phase에서 `moai spec lint --strict` 실행 시 다음 3건의 `StatusGitConsistency` WARNING이 발생:

```
WARNING SPEC-V3R4-HARNESS-001 frontmatter status 'completed' disagrees with git-implied status 'planned'
WARNING SPEC-V3R4-HARNESS-002 frontmatter status 'completed' disagrees with git-implied status 'planned'
WARNING SPEC-V3R4-HARNESS-003 frontmatter status 'completed' disagrees with git-implied status 'planned'
```

3건 SPEC 모두 frontmatter `status: completed`이며 lifecycle도 완료(superseded by NAMESPACE-001)된 상태임에도 git-implied status가 `planned`로 추론되어 일관성 검사가 실패.

**근본 원인 (code-level 검증 완료)**:

`internal/spec/drift.go:119-121`:
```go
cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges",
    "--grep="+specID, fmt.Sprintf("-%d", gitLogWindowSize))
```

`git log --grep=<pattern>`은 commit message(subject + body) 전체에서 **substring 매칭**을 수행한다. `specID = "SPEC-V3R4-HARNESS-001"` 검색 시:

- 매칭 commit (newest-first):
  - `ea1c10647 plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — harness namespace governance + verb policy (#944)` ← walker first match (false)
  - `19957efd8 sync(specs): V3 final status closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed) (#927)`
  - `d40cb9b8b plan(SPEC-V3R4-HARNESS-003): Embedding-Cluster Classifier (...)`
  - `2e27c14f8 chore(post-V3R4-HARNESS-001): my-harness/* 인스턴스화 + brain/IDEA-004 + config 정리 (...)`
- ea1c10647 commit subject는 `NAMESPACE-001` SPEC을 지칭하지만, 본문(body)에 supersede되는 `HARNESS-001/002/003` SPEC들이 언급되었기 때문에 substring `SPEC-V3R4-HARNESS-001` 매칭에 걸려 walker first 결과로 채택됨.
- `ClassifyPRTitle("plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 ...")` → `("plan-merge", "planned")` 반환.
- frontmatter `status: completed` ≠ git-implied `"planned"` → false-positive WARNING.

**진짜 완료 시그널 commit (`e8e38b17b sync(SPEC-V3R4-HARNESS-001): status transition draft → implemented`)은 ea1c10647보다 더 오래된 commit으로, newest-first 순회에서 ea1c10647이 먼저 채택되어 도달하지 못한다.**

같은 패턴이 `SPEC-V3R4-HARNESS-002`, `SPEC-V3R4-HARNESS-003`에도 동일 발생 (NAMESPACE-001 supersede commit이 dominant).

### 1.2 사용자 결정 사항

다음 결정은 plan-phase에서 잠정 채택된다. run-phase에서 plan-auditor 결과에 따라 조정될 수 있다:

1. **방향 (root cause 정밀화)**: `lint.skip` 우회는 금지. drift.go walker를 단어 경계 정밀 매칭으로 격상하여 근본 원인을 해소한다.
2. **접근법**: 3개 후보(A: regex word-boundary grep, B: 2-pass post-filter, C: A+B 조합) 중 plan-phase 잠정 채택은 **Approach B (2-pass post-filter)**. 이유: ① git log POSIX BRE/ERE 호환성 리스크 회피, ② 기존 `ExtractSPECIDs` 정규식 재사용 (외부 의존성 0), ③ N=50 commits × O(1) regex match는 1ms 이하로 성능 영향 미미. run-phase에서 plan-auditor 검토 후 최종 확정.
3. **transitions.go 변경 금지**: `transitions.go:27 {"sync", transition{"sync-merge", "completed"}}` 는 `strings.HasPrefix` 사용 → `sync(spec):`와 `sync(specs):` 모두 매칭하므로 plural 처리는 이미 정상. SPEC scope 외.
4. **`shouldSkipCommitTitle` 메커니즘 유지**: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001에서 도입한 `chore(spec):` skip 필터는 walker filter 단계로 동일하게 유지 (functional regression 0).
5. **새 의존성 도입 금지**: stdlib `regexp` + 기존 `ExtractSPECIDs` (transitions.go:97-116)만 사용.

---

## 2. Scope

### 2.1 In Scope

- `internal/spec/drift.go` `getGitImpliedStatus` walker에 단어 경계 SPEC-ID 매칭 추가 (Approach B 잠정).
- 회귀 방지를 위한 unit test 신설 (`TestGetGitImpliedStatus_SPECIDWordBoundary` 등).
- `moai spec lint --strict` 실행 시 현재 3 WARNING 발생 → 0 WARNING 해소 확인.
- 기존 `shouldSkipCommitTitle` 필터와 새 word-boundary 필터의 순서 결정 및 문서화 (`@MX:NOTE` 갱신).
- 이 SPEC의 plan/run/sync 3-PR cycle.

### 2.2 Out of Scope

- `transitions.go` 변경 (sync plural은 이미 정상 동작).
- 기존 SPEC frontmatter content 수정 (HARNESS-001/002/003 등 — walker fix 후 자동 해소).
- 새 lint rule 추가 또는 기존 rule severity 변경.
- 다른 walker (예: `manager-docs` 의 git log 사용) 전체 audit.
- 197 SPECs 전체 마이그레이션 또는 frontmatter cleanup.
- `lint.skip` 우회 또는 grace window 도입.
- 새 외부 의존성 도입 (regex 외).

---

## 3. Requirements (EARS Format)

### REQ-LSGF-001 (Ubiquitous)

The walker (`internal/spec/drift.go:getGitImpliedStatus`) SHALL perform SPEC-ID matching with **exact word-boundary semantics** so that substring collisions (e.g., `SPEC-V3R4-HARNESS-001` matching `SPEC-V3R4-HARNESS-NAMESPACE-001`) cannot pollute the walker result.

**Acceptance**: AC-LSGF-001, AC-LSGF-002

### REQ-LSGF-002 (Event-Driven)

**WHEN** a SPEC has both a substring-matched non-target commit (e.g., supersede commit referencing the SPEC-ID in body) and a genuine target commit (e.g., `sync(SPEC-XXX):` for that exact SPEC-ID), **THE walker SHALL** return the genuine commit's classified status, not the substring-matched commit's status.

**Acceptance**: AC-LSGF-002, AC-LSGF-004

### REQ-LSGF-003 (State-Driven)

**WHILE** iterating git log results in newest-to-oldest order, **THE walker SHALL** discard commits that do not contain an **exact match** for the target SPEC-ID (regardless of whether the SPEC-ID appears in the commit subject or body), and continue to the next commit candidate.

**Acceptance**: AC-LSGF-004

### REQ-LSGF-004 (Unwanted)

**IF** a commit body lists multiple SPEC-IDs in a supersede/closeout context (e.g., `plan(spec): SPEC-X-NAMESPACE-001 — supersedes SPEC-X-001, SPEC-X-002, SPEC-X-003`), **THEN** that commit SHALL NOT determine the walker's status decision for SPECs other than the one identified in the commit subject's primary scope (as extracted by `ExtractSPECIDs` of the first prefix-scope token).

**Acceptance**: AC-LSGF-002, AC-LSGF-004

### REQ-LSGF-005 (Backward-Compatible)

The existing `shouldSkipCommitTitle` chore-skip mechanism (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001) SHALL remain functionally unchanged after the fix. The new word-boundary filter SHALL compose with the chore-skip filter without behavioral drift.

**Acceptance**: AC-LSGF-003

---

## 4. Exclusions (What NOT to Build)

1. **transitions.go 변경 금지** — `sync(spec):`/`sync(specs):` plural 처리는 `strings.HasPrefix("sync", ...)` 로 이미 정상 동작. 본 SPEC은 transitions.go에 손대지 않는다.
2. **lint.skip 우회 금지** — `.moai/specs/SPEC-V3R4-HARNESS-{001,002,003}/spec.md`에 `lint.skip: [StatusGitConsistency]` 를 추가하여 WARNING을 mask하는 방식은 명시적으로 거부한다. 근본 원인 fix만 허용.
3. **새 lint rule 신설 금지** — `StatusGitConsistency` rule severity, scope, threshold 모두 불변.
4. **새 외부 의존성 금지** — `regexp` (stdlib) + `ExtractSPECIDs` (기존, transitions.go:97-116) 외 도입 없음.
5. **HARNESS-001/002/003 spec.md 자체 수정 금지** — frontmatter 내용 변경 없이 walker fix만으로 WARNING이 0이 되어야 한다 (그래야 진짜 fix). 만약 frontmatter touch가 필요해지면 그것은 fix가 미완료라는 신호.
6. **grace window/소프트 임계값 도입 금지** — binary AC (0 WARNING) 유지.
7. **기존 197 SPECs frontmatter mass-update 금지** — walker fix 후 자동 해소되는지 검증만 수행. 잔존 WARNING이 있다면 별도 follow-up SPEC.

---

## 5. Constraints

- **Code language**: Go (stdlib `regexp` + `bufio` + `strings` + `exec`).
- **Comment language**: Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`).
- **Schema**: 12-field canonical frontmatter (per SPEC-V3R4-SPECLINT-DEBT-002 SSOT `.claude/rules/moai/development/spec-frontmatter-schema.md`).
- **TDD discipline**: Wave 1 RED test 먼저 → Wave 2 GREEN implementation → Wave 3 REFACTOR.
- **Performance budget**: walker 1회 호출당 N=50 commits × O(1) regex match ≪ 5ms (현재 base ~30ms 대비 무시 가능).
- **MX tag**: `@MX:NOTE` / `@MX:REASON` 갱신 필수 (walker filter 변경 시 사유 문서화).

---

## 6. Stakeholders

- **Author**: manager-spec (이 SPEC의 plan-phase 작성자)
- **Implementer (run-phase)**: manager-develop (TDD cycle)
- **Validator**: plan-auditor (plan-phase audit), `moai spec lint --strict` (AC 검증)
- **End user**: MoAI sync workflow (`/moai sync` 사용자가 `moai spec lint --strict` 실행 시 false-positive 0건 보장)

---

## 7. References

- Bug evidence: `git log main --oneline --no-merges --grep=SPEC-V3R4-HARNESS-001 -5` 출력 (substring 매칭 결과 ea1c10647 walker first).
- Lessons #16 (sync-prefix-correction) — 본 SPEC은 그 학습 체인의 종착점.
- PR #944 sync drift incident — direct trigger.
- CLAUDE.local.md §18.12 BODP — orchestrator pre-evaluated `¬a ¬b ¬c` → base `origin/main` fresh branch.
- `.claude/rules/moai/workflow/spec-workflow.md` — SPEC phase discipline.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SSOT for 12-field frontmatter.
- `internal/spec/drift.go:96-197` — walker code + chore skip filter.
- `internal/spec/transitions.go:15-117` — ClassifyPRTitle + ExtractSPECIDs.

---

## 8. AC ↔ REQ Mapping (summary)

| AC ID         | REQ ID(s)               | Verification |
|---------------|-------------------------|--------------|
| AC-LSGF-001   | REQ-LSGF-001            | `moai spec lint --strict` 0 ERROR / 0 WARNING |
| AC-LSGF-002   | REQ-LSGF-001, REQ-LSGF-002, REQ-LSGF-004 | walker unit test `TestGetGitImpliedStatus_SPECIDWordBoundary` |
| AC-LSGF-003   | REQ-LSGF-005            | regression `TestClassifyPRTitle_ChoreSpecUnchanged` |
| AC-LSGF-004   | REQ-LSGF-002, REQ-LSGF-003, REQ-LSGF-004 | new walker unit test (substring vs exact 비교) |
| AC-LSGF-005   | (all)                   | `go test ./internal/spec/... -race` 0 fail |

상세 시나리오는 `acceptance.md` 참조.
