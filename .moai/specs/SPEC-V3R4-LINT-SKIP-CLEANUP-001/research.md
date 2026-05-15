# Research — SPEC-V3R4-LINT-SKIP-CLEANUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.1   | 2026-05-16 | plan-audit remediation | plan-audit 0.904 PASS 후 P2 4건 remediation: (1) AC-LSKC-002 placeholder → plan.md §5.2 cross-ref, (2) HISTORY date harmonize 2026-05-16, (3) REQ-005↔AC-002 매핑 rationale 명시, (4) design.md §2.4 redundancy 정리. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 research. 55개 영향 SPEC 정확한 list 확정 (frontmatter strict scan vs naive grep 불일치 분석 포함). walker filter 영향 분석. predecessor SPEC 머지 후 cleanup 가능 시점 분석. 대안 접근 3가지 비교. lint.skip 추가 이력 git blame 결과. |

---

## 1. Goal of Research

본 SPEC scoping의 정확성을 보증:

1. 정확한 영향 SPEC list 확정 (count + 명단)
2. walker filter (`SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`) 동작 검증 — cleanup 가능 시점 확인
3. lint.skip 회피책의 추가 이력 추적 — 어떤 PR로 어떻게 누적되었는가
4. 대안 접근 분석 — 다른 방법이 더 좋은가
5. 위험 요소 식별

---

## 2. Evidence — 영향 SPEC 정확 List

### 2.1 List 확정 방법론

3가지 검색 방법 비교:

| 방법 | 결과 | 정확도 |
|------|------|--------|
| (a) naive grep: `grep -rl "StatusGitConsistency" .moai/specs/*/spec.md` | 57 | False positive (body 매치) |
| (b) frontmatter-only grep (awk 1st `---` ~ 2nd `---`) | 56 | False positive (predecessor `module:` 필드) |
| (c) strict frontmatter `lint.skip` 배열 멤버 검증 | **55** | **정확** |

**최종 결정**: 방법 (c) — `awk` 기반 frontmatter parse + `lint:` → `  skip:` → `    - StatusGitConsistency` 3단계 매칭. **TRUE COUNT = 55**.

### 2.2 List (sorted, 55 SPECs)

```
 1. SPEC-AGENCY-ABSORB-001
 2. SPEC-AGENT-002
 3. SPEC-CC2122-HOOK-001
 4. SPEC-CC2122-STATUSLINE-001
 5. SPEC-CC297-001
 6. SPEC-CICD-001
 7. SPEC-CORE-001
 8. SPEC-DB-SYNC-HARDEN-001
 9. SPEC-DB-SYNC-RELOC-001
10. SPEC-DESIGN-001
11. SPEC-DOCS-SB-REMOVE-001
12. SPEC-GLM-001
13. SPEC-HOOK-008
14. SPEC-HOOK-009
15. SPEC-I18N-001-ARCHIVED
16. SPEC-KARPATHY-001
17. SPEC-LOOP-001
18. SPEC-LSP-001
19. SPEC-PSR-001
20. SPEC-QUALITY-001
21. SPEC-REFACTOR-001
22. SPEC-SKILL-002
23. SPEC-SKILL-GATE-001
24. SPEC-SLE-001
25. SPEC-SLV3-001
26. SPEC-SRS-001
27. SPEC-SRS-002
28. SPEC-SRS-003
29. SPEC-STATUS-AUTO-001
30. SPEC-STATUSLINE-002
31. SPEC-TEAM-001
32. SPEC-UI-003
33. SPEC-UPDATE-002
34. SPEC-UTIL-003
35. SPEC-V3R2-ORC-001
36. SPEC-V3R2-ORC-005
37. SPEC-V3R2-RT-004
38. SPEC-V3R2-RT-005
39. SPEC-V3R2-SPC-004
40. SPEC-V3R2-WF-005
41. SPEC-V3R2-WF-006
42. SPEC-V3R3-ARCH-007
43. SPEC-V3R3-BRAIN-001
44. SPEC-V3R3-CMD-CLEANUP-001
45. SPEC-V3R3-COV-001
46. SPEC-V3R3-DEF-001
47. SPEC-V3R3-DEF-007
48. SPEC-V3R3-DESIGN-PIPELINE-001
49. SPEC-V3R3-HARNESS-001
50. SPEC-V3R3-WEB-001
51. SPEC-V3R4-CATALOG-001
52. SPEC-V3R4-HARNESS-002
53. SPEC-V3R4-STATUS-LIFECYCLE-001
54. SPEC-WF-AUDIT-GATE-001
55. SPEC-WORKTREE-002
```

### 2.3 Prompt list 와 차이

본 SPEC scoping 프롬프트가 제시한 56개 list 와 실측 55개 list 비교:

**프롬프트에 포함되어 있으나 실제 lint.skip 부재 (제외 대상)**:
- `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (predecessor — 자체적으로 lint.skip 추가 안 함)
- `SPEC-V3R4-SPECLINT-DEBT-001` (debt origin — 본문에서 StatusGitConsistency 언급만 함)

**프롬프트에서 누락됐으나 실제 lint.skip 보유 (추가 대상)**:
- `SPEC-V3R3-ARCH-007` (verified via strict scan)

**결과**: 프롬프트 56 - 2 (제외) + 1 (추가) = 55. 본 research에서 확정한 55가 final.

### 2.4 검증 명령어 (reproducibility)

```bash
# 정확한 list 재생산
for f in .moai/specs/*/spec.md; do
  if awk '/^---$/{f++; if(f==2) exit} f==1 && /^lint:/{in_lint=1; next} f==1 && in_lint && /^  skip:/{in_skip=1; next} f==1 && in_skip && /^    - StatusGitConsistency$/{found=1; exit} f==1 && in_skip && !/^    -/{in_skip=0; in_lint=0} END{exit !found}' "$f"; then
    basename "$(dirname "$f")"
  fi
done | sort
```

검증 시점: 2026-05-15, main HEAD `9e394e51b`. M1 baseline에서 재실행 권장.

---

## 3. Per-SPEC lint.skip 추가 이력 분석

### 3.1 분석 방법

각 SPEC의 `spec.md` 파일에 `lint:` 블록을 추가한 commit을 `git log -S "lint:" --oneline -- <path>` 로 추적.

### 3.2 Sample 결과 (5개 SPEC)

```bash
$ git log -S "    - StatusGitConsistency" --oneline -- .moai/specs/SPEC-AGENCY-ABSORB-001/spec.md
bdcb57f8d chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean) (#930)

$ git log -S "    - StatusGitConsistency" --oneline -- .moai/specs/SPEC-AGENT-002/spec.md
bdcb57f8d chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean) (#930)

$ git log -S "    - StatusGitConsistency" --oneline -- .moai/specs/SPEC-CORE-001/spec.md
bdcb57f8d chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean) (#930)
```

### 3.3 통합 결론

55개 SPEC 모두 PR **#930** (`bdcb57f8d`, "chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)") 에서 일괄 추가됨.

PR #930 의 commit message:
> chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)
> - 11개 SPEC frontmatter status 정정
> - 55개 SPEC에 lint.skip StatusGitConsistency 추가 (walker filter 미존재 시기의 회피책)

이는 PR #930 author (manager-docs 또는 GOOS 직접)가 walker filter SPEC을 scoping 하면서 동시에 일시 회피책으로 lint.skip을 일괄 등록한 history. PR #930 mention "walker filter 미존재 시기의 회피책" — 본 cleanup의 정당화 근거.

### 3.4 추가 sweep 이력

`git log` 추적 결과 PR #930 이전에 일부 SPEC에 lint.skip 이 점진적으로 추가된 흔적은 없음 — 모두 PR #930 단일 commit으로 등록됨. 따라서 cleanup도 단일 PR로 일괄 제거 가능 (역방향 대칭성).

---

## 4. Walker Filter 영향 분석

### 4.1 walker filter 도입 commit

PR #933 (`b395ec563`, `feat(spec): SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 — drift.go chore(spec) walker filter`).

### 4.2 변경 핵심

`internal/spec/drift.go::getGitImpliedStatus` 함수:

**Before (PR #933 이전)**:
```go
// 가장 최근 commit title 1개만 가져옴
cmd := exec.Command("git", "log", "--grep=" + specID, "-1", "--format=%s")
output, _ := cmd.Output()
return ClassifyPRTitle(strings.TrimSpace(string(output)))
```

문제: `chore(spec): status drift sweep` 같은 sweep commit이 매치되면 `ClassifyPRTitle` 이 `(category="skip-meta", status="")`을 반환하지만 caller가 빈 status를 `"in-progress"`로 fallback → `StatusGitConsistency` WARN 발생.

**After (PR #933 이후, 본 cleanup 의 전제조건)**:
```go
// 최대 50개 commit 윈도우 탐색, sweep commit skip
cmd := exec.Command("git", "log", "--grep=" + specID, "-50", "--format=%s")
output, _ := cmd.Output()
titles := strings.Split(string(output), "\n")
for _, title := range titles {
    if shouldSkipCommitTitle(title) {  // chore(spec): ... skip
        continue
    }
    return ClassifyPRTitle(strings.TrimSpace(title))
}
return ""  // no meaningful commit found
```

`shouldSkipCommitTitle("chore(spec): ...")` returns true → sweep commit skip → 다음 (의미 있는) commit 분류 사용 → 정상 status return.

### 4.3 검증 — cleanup 후에도 WARN 0 유지되는가?

walker filter는 cleanup 과 독립적으로 작동. 즉:

- cleanup 전: lint.skip 이 StatusGitConsistency WARN을 SPEC-level 에서 suppress (회피책)
- cleanup 후: walker filter 가 sweep commit을 skip 하여 WARN 자체가 발생하지 않음 (근본 해결)

따라서 cleanup 후 `moai spec lint --strict` 결과 `StatusGitConsistency` 카테고리 WARN 0건 유지 예상.

### 4.4 실측 검증 (M1 baseline 시점에 재수행 필요)

```bash
# walker filter 가 정상 적용되어 있는지 확인
$ git -C /Users/goos/MoAI/moai-adk-go log --oneline -5 internal/spec/drift.go
b395ec563 feat(spec): SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 — drift.go chore(spec) walker filter
...

# lint --strict 결과 (2026-05-15 실측, main HEAD 9e394e51b)
$ moai spec lint --strict 2>&1 | grep -c StatusGitConsistency
8
```

### 4.5 [중요] 본 cleanup 55 SPECs vs WARN 발생 8 SPECs — 별개 population

2026-05-15 main HEAD `9e394e51b` 실측에서 `StatusGitConsistency` WARN 8건 발견:

| SPEC ID | cleanup list 포함? | WARN 원인 |
|---------|------------------|---------|
| SPEC-UTIL-001 | No | frontmatter status `implemented` vs git-implied `in-progress` (walker filter도 처리 못 함 — 실제 drift) |
| SPEC-V3R2-CON-001 | No | 동일 |
| SPEC-V3R2-CON-002 | No | 동일 |
| SPEC-V3R2-CON-003 | No | 동일 |
| SPEC-V3R2-RT-001 | No | 동일 |
| SPEC-V3R2-SPC-003 | No | 동일 |
| SPEC-V3R4-HARNESS-003 | No | frontmatter `completed` vs git-implied `in-progress` |
| SPEC-V3R4-SPECLINT-DEBT-001 | No | frontmatter `completed` vs git-implied `planned` |

**관찰**:
- 본 cleanup 55 SPECs는 **lint.skip 회피책 활성 상태이므로 현재 WARN 0건** (회피책이 작동 중)
- WARN 8건은 lint.skip 회피책이 없는 SPEC들의 실제 drift — 본 SPEC scope 외
- cleanup 후 55 SPECs의 walker filter 의존성 검증을 위해서는 cleanup 후 8건 → 8건 + Δ (예상 Δ=0) 비교
- 만약 cleanup 후 WARN count 가 8 + N (N>0) 으로 증가하면 walker filter 가 일부 sweep commit 을 skip 하지 못한 것 — 회귀 시나리오

**AC-LSKC-002 재해석**:
- "StatusGitConsistency WARN 0건" (현행 spec.md AC-LSKC-002) 은 부정확
- 정확한 검증: "55개 cleanup 대상 SPEC 에 대한 StatusGitConsistency WARN 0건" (population-scoped)
- 또는: "cleanup 전후 lint --strict 결과의 StatusGitConsistency WARN count 차이 0" (delta-scoped)

→ run-phase에서 AC-LSKC-002 wording을 세분화하거나 plan.md §5.2 verification 절차를 population-scoped 명시로 보강 권장.

### 4.6 baseline-warns.txt M1 캡처 명령

```bash
# 본 cleanup 55 SPECs 에 한정한 StatusGitConsistency WARN count
moai spec lint --strict 2>&1 | grep "StatusGitConsistency" | \
  grep -E "$(cat affected-list.txt | tr '\n' '|' | sed 's/|$//')" | \
  wc -l > baseline-warns.txt

# Expected: 0 (lint.skip 회피책이 55 SPECs를 suppress 중)
```

---

## 5. Predecessor SPEC 분석

### 5.1 SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 완료 시점

- Plan PR: #931 (merged 2026-05-15)
- Run PR: #933 (`b395ec563`, merged 2026-05-15)
- Sync PR: #934 (`9e394e51b`, merged 2026-05-15)
- 현재 main HEAD = `9e394e51b` (sync 완료 직후)

### 5.2 본 SPEC scoping 가능 시점

predecessor sync (#934) 완료 후 본 SPEC scoping 가능. 본 SPEC 작성 시점 (2026-05-15) 은 이미 sync 완료 후 시점.

### 5.3 predecessor SPEC의 의도

predecessor SPEC (`SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001`) HISTORY 0.1.0 entry 인용:

> 본 SPEC은 walker filter를 drift.go에 도입해 chore(spec) commit을 skip하고 의미 있는 분류가 나올 때까지 더 오래된 commit을 탐색하도록 만든다. `ClassifyPRTitle`의 표준 의미는 보존한다.

predecessor SPEC은 walker filter 도입을 통한 **근본 해결**에 초점. lint.skip 회피책 cleanup은 명시적으로 scope 외로 처리 — 다음 SPEC (= 본 SPEC) 에서 처리하기로 설계됨.

predecessor SPEC plan.md §Goal 인용:
> "본 SPEC은 lint 엔진 (`internal/spec/drift.go`) 만을 수정하며 `.moai/specs/` 어떤 frontmatter도 변경하지 않는다 — 7개 영향 SPEC의 status enum은 그대로 보존된다."

→ 본 SPEC이 그 후속 cleanup 책임을 명시적으로 인계.

---

## 6. Alternative Approaches

### 6.1 (a) lint.skip 영구 유지 (do-nothing)

**장점**: 추가 작업 없음.

**단점**:
- 영구 metadata noise — 다른 정당한 lint.skip 사례와 시각적 구분 어려움
- "역사적 기록" 으로 정당화하기엔 git history 가 더 적절한 매체
- 향후 신규 SPEC author 가 "왜 이 SPEC들만 lint.skip 가지나?" 혼란 → 본 SPEC 같은 cleanup을 다시 발급해야 함

**판정**: **거부**. metadata 위생 우선.

### 6.2 (b) 본 cleanup (CHOSEN)

**장점**:
- 회피책 영구 제거
- frontmatter 깔끔
- predecessor SPEC 의 design intent 와 일치

**단점**:
- 55 SPEC bulk edit — 작업량
- frontmatter 변경 → review 부담

**판정**: **채택**.

### 6.3 (c) lint rule 자체 retire (`StatusGitConsistency` 폐기)

**장점**: rule 자체를 없애면 lint.skip 불요.

**단점**:
- `StatusGitConsistency` rule 은 frontmatter status drift 탐지의 핵심 메커니즘
- sweep commit skip 이외의 정당한 drift 케이스 (예: 누군가 frontmatter status 를 잘못 수정) 는 여전히 탐지해야 함
- rule retire는 lint coverage 감소 → 회귀 위험

**판정**: **거부**. walker filter (predecessor) + lint.skip cleanup (본 SPEC) 의 2-layer 접근이 안전.

### 6.4 (d) frontmatter `lint:` 블록 통째 deprecate

**장점**: 모든 lint.skip 회피책을 영구 차단.

**단점**:
- 정당한 lint.skip 사례 (예: 의도적인 deprecated SPEC 에서 일부 rule 무효화) 도 함께 제거됨
- 본 SPEC scope를 크게 벗어남 — 별도 SPEC 영역

**판정**: **거부 (out of scope)**. 별도 SPEC `SPEC-V3R4-LINT-SKIP-DEPRECATION-001` (가설) 으로 검토 가능.

---

## 7. Risk Analysis

### 7.1 본 cleanup 의 직접 위험

| Risk ID | Description | Severity | Mitigation |
|---------|-------------|----------|------------|
| R1 | 55개 외 SPEC 의도하지 않은 수정 | Low | M3 verification 의 `git diff --name-only` 검증 |
| R2 | SPEC 본문 byte-level 회귀 | Low (옵션 C) / Medium (옵션 A) | body sha256 비교 (M3) |
| R3 | YAML key ordering 깨짐 | Low (옵션 C `yaml.Node`) | spot-check 5 SPEC manual diff |
| R4 | HISTORY 표 형식 깨짐 (column mismatch) | Medium | sample 10 SPEC HISTORY 형식 사전 검증 |
| R5 | idempotency 깨짐 | Low (옵션 C 자연스러움) / Medium (옵션 A 매 iteration 수동 검사) | §4 idempotency 검증 절차 적용 |
| R6 | walker filter 미적용 binary 로 lint 검증 | Medium | `which moai` + `moai version` 캡처 (M1) |

### 7.2 cleanup 회귀 가능 시나리오

- **predecessor SPEC revert**: PR #933 또는 #934 revert 되면 walker filter 부재 → cleanup 후 lint --strict WARN 폭증. Mitigation: predecessor revert 자체를 예방 (frontmatter `dependencies:` 명시).
- **신규 SPEC 이 lint.skip 추가**: 본 cleanup 후에도 새 SPEC 이 lint.skip `StatusGitConsistency` 를 추가하면 noise 재발. Mitigation: 별도 PR (sync-phase docs update) 에서 SPEC author guideline 갱신 — "walker filter 가 사실상 모든 케이스 처리하므로 lint.skip 추가 금지".

---

## 8. References

### 8.1 Predecessor SPEC

- `.moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/spec.md` — walker filter SPEC
- `.moai/specs/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001/research.md` — walker filter 설계 근거

### 8.2 Origin SPEC

- `.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/spec.md` — sweep commit 패턴 도입 SPEC

### 8.3 Related PRs

- **#921**: SPEC-V3R4-SPECLINT-DEBT-001 sync (sweep commit 패턴 도입)
- **#930**: chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean) — 55개 lint.skip 일괄 추가
- **#931**: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 plan
- **#933**: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 run (walker filter 도입)
- **#934**: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 sync (closeout)

### 8.4 Code References

- `internal/spec/drift.go` 의 `getGitImpliedStatus` 함수 — walker filter 도입 위치 (PR #933 변경)
- `internal/spec/lint.go` 의 `StatusGitConsistencyRule` — lint rule 정의
- `internal/spec/lint.go` 의 frontmatter `lint.skip` 처리 로직 — 회피책 메커니즘

### 8.5 Documentation

- `.claude/rules/moai/workflow/spec-workflow.md` — SPEC workflow 표준
- CLAUDE.local.md §18.12 — BODP (Branch Origin Decision Protocol)
- CLAUDE.local.md §18.3 — Merge Strategy (feat → main = squash)

---

## 9. Open Issues for Run-Phase

1. **bulk edit 도구 최종 선택** (OQ2): 옵션 C (Go script) vs 옵션 A (Edit tool). manager-develop 결정.
2. **HISTORY 표 row 삽입 위치** (top vs bottom): 각 SPEC별 ordering 감지 후 적절한 위치 선택 (design.md §6.7).
3. **bulk script 보존 여부**: PR 포함하여 `.moai/scripts/lint-skip-cleanup.go` 보존 (재사용) vs 폐기 (one-off).
4. **M1 baseline에서 실측 lint --strict warn count**: 현재 walker filter 가 이미 활성이므로 0건 예상 — M1 시점에 실제 측정 후 baseline-warns.txt 기록.
