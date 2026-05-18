# Research — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.1   | 2026-05-15 | plan-audit remediation | P2-3 결함 반영: transitions.go 인용 라인 실측값으로 보정 (§7.1 `ClassifyPRTitle` 함수 본문 60-92 → 70-92, §2.3 step 4 `for _, rule := range transitionRules` 루프 75-92 → 84-88). 동시에 sibling SPEC 문서 일괄 0.1.1 bump (spec.md REQ-002 §3.4 재배치 + REQ-010 EARS Optional, design.md §3.2 build→refactor + §6.5 ack). |
| 0.1.0   | 2026-05-15 | manager-spec  | 초기 research. 7건 WARNING 증거 + sweep commit body 분석 + ClassifyPRTitle 동작 분석 + lint.skip 51건 카운트 + 대안 4건 평가 + 참조 file path/line. |

---

## 1. Evidence: 7건 WARNING (현재 main = bdcb57f8d)

### 1.1 `moai spec lint --strict` 출력 발췌

origin/main commit `bdcb57f8d` 머지 직후 (2026-05-15 11:57Z) 측정:

```text
WARN: SPEC-UTIL-001 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R2-CON-001 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R2-CON-002 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R2-CON-003 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R2-RT-001 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R2-SPC-003 (status: implemented) — git-implied status 'in-progress' [StatusGitConsistency]
WARN: SPEC-V3R4-HARNESS-003 (status: completed) — git-implied status 'in-progress' [StatusGitConsistency]

Total: 0 ERROR, 7 WARNING
```

`--strict` flag 적용 시 WARNING은 ERROR로 승격되어 CI fail 위험을 동반한다. 본 SPEC이 해소하지 않으면 `spec-lint` workflow가 적색을 유지하며 다른 SPEC PR 머지에 압력을 가한다 (선행 SPEC-V3R4-SPECLINT-DEBT-001 의 lessons).

### 1.2 7건 영향 SPEC 의 git history (sample)

7개 영향 SPEC 의 `git log --oneline --no-merges --grep=<specID> -10` 출력 요약:

| SPEC | matching commits (newest first) | latest title prefix |
|------|--------------------------------|---------------------|
| SPEC-UTIL-001 | bdcb57f8d, ...prev-impl-commit... | `chore(spec):` |
| SPEC-V3R2-CON-001 | bdcb57f8d, multiple prev commits | `chore(spec):` |
| SPEC-V3R2-CON-002 | bdcb57f8d, multiple prev commits | `chore(spec):` |
| SPEC-V3R2-CON-003 | bdcb57f8d, multiple prev commits | `chore(spec):` |
| SPEC-V3R2-RT-001 | bdcb57f8d, multiple prev commits | `chore(spec):` |
| SPEC-V3R2-SPC-003 | bdcb57f8d, multiple prev commits | `chore(spec):` |
| SPEC-V3R4-HARNESS-003 | bdcb57f8d, prev sync commit, prev impl commit | `chore(spec):` |

모든 7건이 `bdcb57f8d` (PR #930 sweep commit) 를 latest로 매칭 → fallback 경로 활성화 → false-positive WARNING.

---

## 2. Evidence: Sweep Commit (`bdcb57f8d`) Content

### 2.1 Commit 메타데이터

```text
commit bdcb57f8d5f22895f930ecc9e42d889a48b62305
chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean) (#930)

* chore(template): 신규 프로젝트 default effortLevel medium → xhigh
...

* chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)

## 8 status sweep (draft/planned/in-progress → implemented, 1건 draft → completed)
- SPEC-GITHUB-WORKFLOW, SPEC-V3R2-CON-002, SPEC-V3R2-CON-003: draft → implemented
- SPEC-V3R2-CON-001, SPEC-V3R2-SPC-003: planned → implemented
- SPEC-UTIL-001, SPEC-V3R2-RT-001: in-progress → implemented
- SPEC-V3R4-HARNESS-003: draft → completed

## 3 lint-skip 등록 (StatusGitConsistency)
- SPEC-CORE-001, SPEC-LOOP-001, SPEC-V3R4-CATALOG-001

PR #926/#917 일괄 sweep으로 이미 `status: completed`인 SPECs인데 lint는
git-implied status를 `implemented`로 인식한다 (sync 머지 이후 추가 commit이
SPEC 코드 경로에 누적된 영향). 이전 작업 결과를 보존하기 위해 spec.md
frontmatter에 `lint.skip: [StatusGitConsistency]`를 등록하여 false-positive
WARNING만 억제. status enum 자체는 그대로 `completed` 유지.
```

### 2.2 분석

- Commit title prefix: `chore(spec):` → transitions.go line 22 의 `{"chore(spec)", transition{"skip-meta", ""}}` 규칙 매칭
- Commit body: 11개 SPEC ID 모두 plain text로 언급 (SPEC-UTIL-001, SPEC-V3R2-CON-001 등) — `git log --grep=<specID>` 가 body 매칭으로 이 commit을 hit
- 8 sweep + 3 lint-skip = 11건 처리 commit
- 8 sweep 대상 중 7건이 본 SPEC의 WARNING 발생 SPEC (SPEC-GITHUB-WORKFLOW 제외, 이유는 추가 분석 필요 — 해당 SPEC은 별도 lint.skip 보유 가능성)

### 2.3 검증된 추론 체인

1. `git log main --oneline --no-merges --grep=SPEC-UTIL-001 -1` 실행 시 `bdcb57f8d` 의 commit hash + title 한 줄 반환
2. drift.go line 107-110 코드 경로 그대로 실행
3. line 129: `ClassifyPRTitle("chore(spec): status drift 11건 sweep + lint-skip 등록 (lint clean)")` 호출
4. transitions.go line 84-88 의 `for _, rule := range transitionRules` 루프 → line 22 매칭
5. `return "skip-meta", "", nil` 반환
6. drift.go line 134-136: `if status == ""` 진입 → `return "in-progress", nil`
7. lint.go line 902-911: `if fm.Status != gitStatus` (implemented != in-progress) → finding 추가 → WARNING 출력

위 7단계는 source code 직접 추적으로 verify됨. 추측 없음.

---

## 3. Evidence: ClassifyPRTitle Behavior

### 3.1 `transitions.go` 의 chore(spec) 분류 규칙

`internal/spec/transitions.go` lines 19-23:

```go
// Plan phase transitions
{"plan(spec)", transition{"plan-merge", "planned"}},
{"plan(specs)", transition{"plan-merge", "planned"}},
{"chore(spec)", transition{"skip-meta", ""}}, // Loop prevention
{"docs(spec-plan)", transition{"plan-merge", "planned"}},
```

`chore(spec):` 의 분류 결과는 `(category="skip-meta", status="")`. `// Loop prevention` 주석이 의도를 명확히 한다 — chore commit이 sweep 자체이므로 lifecycle 추론에서 제외해야 한다.

### 3.2 Sync 관련 prefix 분석 (OQ3 결정 근거)

| Prefix | transitions.go 위치 | classification | fallback 트리거? |
|--------|---------------------|---------------|------------------|
| `docs(sync)` | line 26 | `(sync-merge, "completed")` | 아니오 (status 비어있지 않음) |
| `sync` | line 27 | `(sync-merge, "completed")` | 아니오 |
| `sync(spec)` | 등록 안 됨 | `("unknown", "", nil)` | 예 (빈 status → fallback) |

`sync(spec):` prefix는 transitions.go에 등록되지 않은 unknown prefix이므로 빈 status를 반환하고 fallback에 의해 "in-progress"로 분류된다. 그러나 실무에서 `docs(sync):` 와 `sync:` 만 사용되며 (PR title 컨벤션 — CLAUDE.local.md §4 Conventional Commits 참조), `sync(spec):` 변형은 거의 등장하지 않는다.

**OQ3 결론**: `sync(spec):` 는 본 SPEC의 skip pattern에 포함하지 않는다. 별도 SPEC으로 transitions.go 규칙 확장 또는 skip pattern 확장 시 재검토.

### 3.3 ClassifyPRTitle 의 의도된 설계

`chore(spec):` → empty status 반환은 의도된 loop prevention 동작이다. 만약 이를 `(skip-meta, "implemented")` 또는 다른 valid status로 변경하면 다른 영향이 발생한다:

- 다른 곳에서 ClassifyPRTitle을 status 결정자로 사용한다면 chore commit이 implementation으로 오인됨
- status auto-update 메커니즘이 chore commit으로 status 전이를 트리거할 위험

따라서 본 SPEC은 `transitions.go` 를 건드리지 않고 walker 호출자 측에서 처리한다 — REQ-LSCSK-001 / AC-LSCSK-003 의 regression guard 근거.

---

## 4. Existing lint.skip Workaround

### 4.1 51개 SPEC의 회피책 카운트

`.moai/specs/*/spec.md` frontmatter에서 `lint.skip: [StatusGitConsistency]` 가 등록된 SPEC을 grep으로 카운트:

```bash
grep -lR "lint.skip" .moai/specs/ | xargs grep -l "StatusGitConsistency" | wc -l
# 결과: 51 (PR #930 머지 직후 측정)
```

51개 SPEC은 다음 history를 가진다:

- SPEC-V3R4-SPECLINT-DEBT-001 (M2 sync sweep) — 약 50건의 `lint.skip` 등록
- PR #930 sweep — 3건 추가 (SPEC-CORE-001, SPEC-LOOP-001, SPEC-V3R4-CATALOG-001)
- 그 외 일부 individual SPEC들이 자체적으로 등록

### 4.2 회피책의 본질적 문제

- **확산 압력**: 매 sweep 마다 새 SPEC들이 회피책에 합류 → 결국 모든 SPEC이 lint.skip 보유 → 검사 무력화
- **검사 의도 훼손**: `StatusGitConsistencyRule` 의 본래 목적 (frontmatter 갱신 누락 탐지) 이 SPEC 단위로 비활성화됨
- **유지보수 부담**: 새 SPEC 작성 시 lint.skip 등록을 잊을 위험 — checklist 의존

본 SPEC은 lint 엔진 자체를 fix하여 회피책 추가 확산을 멈춘다. **기존 51건 lint.skip은 제거하지 않는다** (REQ-LSCSK-009 / scope out-of-scope #3) — 별도 cleanup task로 분리. 다만 fix 머지 후 새 SPEC 작성 시에는 lint.skip 등록이 불필요해진다.

### 4.3 Cleanup 작업의 분리 정당화

51건 lint.skip 제거는 본 SPEC scope 밖인 이유:

1. **검증 부담**: 각 lint.skip 제거 시 해당 SPEC이 실제로 0 WARNING 인지 확인 필요 → 51건 × 검증 = 큰 작업량
2. **회귀 위험**: 본 SPEC fix가 모든 케이스를 커버한다는 보장은 v2.20.0-rc1 출시 후 1주일 anchor 확보 후 결론 가능
3. **분할 정복**: lint engine fix (본 SPEC) → 1주일 anchor → cleanup SPEC (별도) 순서가 안전

별도 cleanup SPEC 후보: `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (제안, 본 SPEC 머지 + 1주 anchor 후 plan-phase 진입).

---

## 5. 선행 SPEC Context

### 5.1 SPEC-V3R4-SPECLINT-DEBT-001 의 lifecycle

| PR | 단계 | Status | Date | Note |
|----|------|--------|------|------|
| #915 | plan | merged | 2026-05-15 | plan-auditor PASS 0.92 |
| #917 | run | merged | 2026-05-15 | ERROR 66 → 0, WARNING 140 → 0, 90 status sync + 51 lint.skip 등록 |
| #921 | sync | merged | 2026-05-15 | status completed, CHANGELOG entry, lessons #SLD logged |

선행 SPEC은 일괄 sweep으로 ERROR/WARNING 대량 부채를 0으로 reset했다. 이 과정에서 51개 SPEC에 `lint.skip` 을 등록했고, 90개 SPEC frontmatter status를 일괄 갱신했다.

### 5.2 본 SPEC과 선행 SPEC의 관계

| 측면 | 선행 SPEC (V3R4-SPECLINT-DEBT-001) | 본 SPEC (V3R4-LINT-STATUS-CHORE-SKIP-001) |
|------|-------------------------------------|------------------------------------------|
| 접근 | Data sweep (frontmatter 일괄 수정) | Engine fix (lint 로직 자체 개선) |
| Scope | 188개 SPEC frontmatter | 1개 Go 함수 + helper + tests |
| 회피책 | lint.skip 51건 등록 (회피책 사용) | 회피책 추가 확산 차단 (근본 해결) |
| Result | ERROR 66→0, WARN 140→0 | 잔존 7 WARN → 0 |
| 위험 | high (188 SPEC frontmatter 일괄 변경) | low (drift.go 1개 함수 수정 + 4 test case) |
| Lifecycle | completed (2026-05-15) | draft (본 SPEC) |

본 SPEC은 선행 SPEC이 해소한 1차 부채를 마무리하여 sweep 자기참조 loop를 닫는 cleanup SPEC이다.

### 5.3 PR #930 (`bdcb57f8d`) 의 결과 — 새 부채 생성

PR #930 은 PR #917 + #921 (선행 SPEC sync) 머지 이후 잔존 status drift 11건 sweep + lint-skip 3건 등록을 수행했다. 그러나 sweep commit 자체가 lint의 git-implied status를 `in-progress` 로 재인식하게 만들어 **7건 새 WARNING 을 발생시켰다**.

이는 sweep 도구가 작업할수록 작업 대상이 늘어나는 "치료가 병을 만든다" 시나리오. 본 SPEC이 이 루프를 끊는다.

---

## 6. Alternative Approaches Considered

### 6.1 Alternative A: lint.skip 메커니즘 확장

**접근**: 7건 영향 SPEC에 `lint.skip: [StatusGitConsistency]` 추가, 또는 글로벌 ignore list 도입.

**Rejected — Rationale**:

- 51건 → 58건으로 회피책 확산 (정확한 반대 방향)
- 매 sweep 마다 새 회피책 추가 압력 영구화
- 검사 무력화의 long-term 비용 > 단기 편의

### 6.2 Alternative B: ClassifyPRTitle의 의미 변경

**접근**: `transitions.go` line 22 의 `chore(spec)` 규칙을 `(skip-meta, "implemented")` 또는 다른 valid status로 변경.

**Rejected — Rationale**:

- ClassifyPRTitle의 standalone semantics 회귀 위험 (AC-LSCSK-003)
- ClassifyPRTitle 호출처가 drift.go 외에도 있을 수 있음 — sweep commit을 "implemented" 로 분류하면 다른 메커니즘 (status auto-update, transition validator) 이 오작동
- `skip-meta` 카테고리의 본래 의도 (lifecycle 추론에서 제외) 정면 위반

### 6.3 Alternative C: drift.go walker filter (CHOSEN)

**접근**: drift.go::getGitImpliedStatus 함수만 변경. transitions.go는 보존.

**Selected — Rationale**:

- 변경 범위 좁음 (1 함수 + 1 helper + 4 tests)
- ClassifyPRTitle의 의미 보존 → 다른 호출처 회귀 없음
- 회피책 누적 압력 해소
- Backward-compatible — signature/CLI/config 모두 변경 없음
- Low risk (failure mode 분석 §6 모두 안전 처리 확인)

### 6.4 Alternative D: 외부 설정화 + Engine fix 동시 진행

**접근**: walker filter + `.moai/config/sections/spec-lint.yaml` `git_status_skip_patterns:` array 동시 도입.

**Rejected — Rationale (Deferred)**:

- v2.20.0-rc1 scope에서 외부 설정화는 추가 회귀 위험 — config 파일 자체가 새 lint 대상이 됨
- 현재 skip pattern은 `chore(spec):` 단일 → externalization 의 이득 < 비용
- KISS principle — 미래 패턴 확장 시점에 externalization 결정 (plan.md §7 OQ2)
- 본 SPEC은 walker filter (Alternative C) 만 적용. Externalization은 별도 follow-up SPEC.

---

## 7. References

### 7.1 소스 파일

| Path | Lines | 관련 |
|------|-------|------|
| `internal/spec/drift.go` | 96-143 | `getGitImpliedStatus` — 본 SPEC 수정 대상 |
| `internal/spec/drift.go` | 134-136 | Fallback 결함 지점 (`if status == ""` → return "in-progress") |
| `internal/spec/lint.go` | 879-914 | `StatusGitConsistencyRule::Check` — 변경 없음, err skip 동작 확인 |
| `internal/spec/lint.go` | 897-900 | err != nil 시 return nil (finding 미발생) — walker error 호환 |
| `internal/spec/transitions.go` | 19-23 | `chore(spec)` 분류 규칙 — 보존 |
| `internal/spec/transitions.go` | 70-92 | `ClassifyPRTitle` 함수 본문 — AC-LSCSK-003 regression guard 대상 |
| `internal/spec/transitions.go` | 22 | `// Loop prevention` 주석 — chore skip 의도 명시 |

### 7.2 선행 SPEC

| SPEC | 관련 |
|------|------|
| `SPEC-V3R4-SPECLINT-DEBT-001` | 1차 lint debt sweep, 본 SPEC의 직접 선행 — `.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/spec.md` |

### 7.3 PR 이력

| PR | 머지 시점 | 관련 |
|----|----------|------|
| #913 | 2026-05-15 | lint baseline reset 시작 (선행 SPEC plan) |
| #917 | 2026-05-15 | 선행 SPEC run-phase 머지 (ERROR/WARN 일괄 sweep) |
| #921 | 2026-05-15 | 선행 SPEC sync-phase 머지 — `status: completed` |
| #926 | 2026-05-15 | V3 final status closeout sweep (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed) |
| #929 | 2026-05-15 | lint multi-REQ regex fix (comma-separated REQ 매핑 인식) — main `c3505543a` |
| #930 | 2026-05-15 | **본 SPEC의 직접 원인 commit** — 11건 status drift sweep + 3 lint.skip 등록 — main `bdcb57f8d` |

### 7.4 메모리 / Lessons

| Memory | 관련 |
|--------|------|
| `project_status_drift_sweep_pr_930_complete` | PR #930 머지 후 7건 WARN 재발견 + V3R4-LINT-STATUS-CHORE-SKIP-001 후보 식별 |
| `project_v3r4_speclint_debt_001_lifecycle_complete` | 선행 SPEC lifecycle 전체 컨텍스트 |
| `project_lint_multireq_regex_fix_complete` | lint engine 2-pass 재구성 사례 — 본 SPEC과 유사한 lint 엔진 단독 수정 패턴 |

### 7.5 외부 표준

| 표준 | 관련 |
|------|------|
| Conventional Commits | `chore(scope):` prefix 의 의미 정의 — 비기능적 메타데이터 변경 |
| EARS Format | 본 SPEC §3 Requirements 형식 |
| Go 1.23 testing | `testing` package, `t.TempDir()`, table-driven tests |

### 7.6 Rules

| Rule | 관련 |
|------|------|
| `.claude/rules/moai/workflow/spec-workflow.md` | SPEC Phase Discipline — Step 1 plan-in-main (BODP) |
| `.claude/rules/moai/workflow/mx-tag-protocol.md` | @MX:NOTE / @MX:ANCHOR 부착 가이드 |
| `.claude/rules/moai/languages/go.md` | Go 1.23+ test/lint 표준 |
| `CLAUDE.local.md §6` | Test isolation (`t.TempDir()`) 규칙 |
| `CLAUDE.local.md §18.12` | BODP — base branch 결정 프로토콜 |
