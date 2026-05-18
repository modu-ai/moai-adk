# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — Design Rationale

> **Scope**: 본 design 문서는 17 SPEC frontmatter remediation 의 **decision framework** 만 다룬다.
> 코드 설계, 알고리즘, 아키텍처 결정은 본 SPEC scope 외 (metadata-only).

## 1. Why FOLLOWUP-002 is Necessary

### 1.1 LSGF-001 의도된 부수효과

`SPEC-V3R4-LINT-SPECID-GREP-FIX-001` 은 walker `git log --grep=<SPEC-ID>` substring 매칭 의 false-positive 문제 (V3R4-HARNESS-001 → NAMESPACE-001 collision) 를 영구 차단하기 위해 word-boundary 정밀 매칭으로 격상했다.

이 fix 의 영구적 부수효과로:
- substring 충돌로 가려져 있던 **진짜 status drift** 가 walker 에 의해 정확하게 보고됨
- main 에서 lint 가 17 WARNING 노출 — 모두 LSGF-001 이전부터 존재하던 real drift

LSGF-001 본문은 이 부수효과를 명시적으로 인정 (`SPEC-V3R4-LINT-SPECID-GREP-FIX-001/spec.md §1.1 Goal`) 하며, 17건 remediation 을 FOLLOWUP SPEC 으로 분리하기로 plan-phase 결정함.

### 1.2 SDF-001 precedent

`SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (SDF-001) 은 `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (LSKC-001) lint.skip 회피책 제거 후 노출된 64건 real drift 를 Pattern A~H 8 분류 + 5 Wave 처리로 해소했다.

FOLLOWUP-002 는 이 lineage 의 직계 후속:

```
LSKC-001 cleanup → 64 drift 노출 → SDF-001 (Wave 1-5) → 64 → 4 → 0
LSGF-001 walker fix → 17 drift 노출 → FOLLOWUP-002 (Wave 1-3) → 17 → 0
```

SDF-001 의 패턴 분류 모델은 본 SPEC 에서도 유효하나, FOLLOWUP-002 의 17건은 SDF-001 Pattern A (전체 50건 대부분) 가 아닌 **Walker visibility 한계** 패턴이 dominant. 따라서 본 SPEC 은 Category A (forward sync-up) + Category B (per-SPEC investigation) 의 2-카테고리 분류만 사용.

### 1.3 v2.20.0-rc1 release readiness gate

본 SPEC closure + 별도 `SPEC-V3R4-CI-INFRA-FIX-001` (CI workflow 결함 해소) closure 모두 완료되어야 `v2.20.0-rc1` 태그 발행 가능. lint 0 ERROR / 0 WARNING 은 release gate 의 핵심 indicator.

---

## 2. Why Per-SPEC Analysis is Required (vs Blanket Sync-up)

### 2.1 Bootstrapping bug 재발 방지

`SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (PR #930) 의 11건 일괄 sweep + lint-skip 등록 은 후속 LSKC-001 cleanup 에서 mass-skip 의 단점을 드러냈다. 모든 SPEC을 `lint.skip` 으로 처리하면 walker 의 detection 가치 자체가 사라진다.

따라서 본 SPEC 은:
- Category A (5건): 단순 frontmatter sync-up — 안전. Wave 2-A 단일 commit.
- Category B (12건): per-SPEC mechanism 결정 — Wave 1 evidence-driven analysis 필수.

### 2.2 Walker miss 원인의 다양성

Category B 12건 의 walker miss 원인은 단일 원인이 아님:

| 원인 클래스 | SPECs (가설) | 적절한 mechanism |
|-------------|--------------|------------------|
| Bulk closure visibility | B1, B7 | lint.skip with reason "synced via bulk-closure PR #XXX" |
| Prefix unrecognized (docs(sync), fix(hook)) | B3, B11 | sync(spec) commit 또는 lint.skip with prefix reason |
| Multi-SPEC sync commit 우산 | B6, B8, B9, B10 | sync(spec) commit (해당 SPEC 명시) |
| Subtask prefix mismatch (T-SPC002-*) | B4 | lint.skip with "subtask prefix walker miss" |
| Plan backfill overrode implementation | B5 | sync(spec) commit 또는 frontmatter clarify |
| Implementation event 부재 | B2 | frontmatter downgrade (외부 증거 검증 후) |
| SDF-001 chain self-drift | B12 | lint.skip with chain-rolled-up reason |

mechanism 선택은 **per-SPEC evidence** 에 기반해야 함. 일괄 lint.skip 은 미적절.

### 2.3 Audit trail 의 가치

각 lint.skip / sync-commit / downgrade 결정이 HISTORY entry 에 reason 으로 기록되면, 미래 walker 개선 시 (예: 가상의 SPEC-V3R4-WALKER-BULK-SYNC-RECOGNITION-001) 어떤 case 를 자동 해소 가능한지 reference materials 로 활용 가능.

---

## 3. Mechanism Boundary: sync-commit vs chore-commit vs lint.skip

### 3.1 sync(spec): vs chore(spec):

`internal/spec/transitions.go` (LSCSK-001 walker filter post-fix):

```go
transitions := []struct {
    prefix     string
    transition transition
}{
    {"plan", transition{"plan-merge", "planned"}},
    {"feat", transition{"feat-merge", "implemented"}},
    {"sync", transition{"sync-merge", "completed"}},  // <-- walker 가 인식
    // ...
}

shouldSkipCommitTitle(title)  // chore(spec):, chore(specs): → SKIP
```

따라서:
- `sync(spec): SPEC-X — closeout` → walker 가 `("sync-merge", "completed")` 인식 → git-implied `completed`
- `chore(spec): SPEC-X — sync` → walker 가 SKIP → git-implied 가 그 commit을 무시하고 다음 hit 확인

**Implication**: status 끌어올리려면 `sync(spec):` prefix 필수. `chore(spec):` 는 sweep commit 용도이며 status drift 해소 목적엔 부적합.

> **Lesson #16 (SDF-001 산출)**: `chore(spec):` 는 walker filter SKIP 대상이므로 sync commit은 `sync(spec):` / `feat(spec):` prefix 사용 필수.

### 3.2 sync-commit mechanism 의 한계

`sync(spec): SPEC-X — closeout under FOLLOWUP-002` commit 추가 시:
- walker 가 다음 lint 실행에서 해당 commit 을 인식 → SPEC-X git-implied 가 `completed` 로 격상
- 그러나 manually-created sync commit 은 retroactive 이므로 historical accuracy 가 약간 손상
- 또한 이 sync commit 자체가 SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 의 sync 단계 commit 이 아니라 본문 commit 이므로 본 SPEC 의 closeout 시점과 별개

**Trade-off**: walker visibility 회복 vs commit history accuracy.

본 SPEC 은 sync-commit mechanism 을 우선 채택하되, retroactive 성 명시 + HISTORY entry 의 인용 PR 참조로 trade-off 완화.

### 3.3 lint.skip mechanism 의 한계

`lint.skip: [StatusGitConsistency]` frontmatter 추가 시:
- walker 의 status drift detection 을 해당 SPEC 에서 영구 무력화
- 향후 frontmatter 가 또 다른 drift 상태가 되어도 lint 가 감지 못함
- 명확한 reason 기록 필수 (HISTORY entry)

**Trade-off**: 영구 false-positive 회피 vs detection 능력 손실.

본 SPEC 은 lint.skip 을 **walker visibility 한계가 명백한 경우에만** 사용 (예: docs(sync) prefix 미인식, T-SPC002-* subtask prefix 미인식, SDF-001 chain self-drift). 일반 sync visibility 케이스는 sync-commit 우선.

### 3.4 frontmatter downgrade mechanism 의 한계

frontmatter `status: implemented → planned` 같은 downgrade 시:
- walker 와 frontmatter 일치 → lint clean
- 그러나 SPEC 의 실제 완료 상태 misrepresentation 위험 (만약 외부 증거가 완료 입증)

**Trade-off**: lint clean vs SPEC accuracy.

본 SPEC 은 downgrade 를 **git log + 외부 증거 (CHANGELOG / project memory / merged PR) 모두 implementation 미완료를 입증할 때만** 채택. 단 하나라도 완료 증거 있으면 sync-commit 또는 lint.skip 선택.

---

## 4. Why 17 (Not 18, Not 16)

본 SPEC 의 inventory 는 정확히 17건 — `moai spec lint --strict` (main `139c4d9d0`) 결과의 binary fingerprint.

추가 (18번째) 가 노출되면:
- **본 SPEC scope 확장 금지** (REQ-SDF002-006 violations risk)
- 별도 `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-003` 발급

부족 (16건) 시:
- AC-SDF002-A-N / B-N 갯수 조정 + 미포함 SPEC 의 사유 inline 기록

본 SPEC 의 17 inventory 는 plan-phase 시점 main HEAD `139c4d9d0` 의 lint 출력에 anchor 되어 있음. run-phase 이전에 main 이 변경되어 새 commits 가 추가되면 inventory 가 변경될 수 있음. 그 경우 plan.md §3 갱신.

---

## 5. Why No Worktree

본 SPEC 의 모든 작업은 `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/` (본 SPEC 디렉토리) + `.moai/specs/SPEC-X/spec.md` (17 target frontmatter) + `CHANGELOG.md` 만 수정. 모두 markdown / YAML 파일.

Worktree 필요성:
- 코드 isolation: ❌ (Go 코드 미수정)
- 병렬 SPEC 작업 isolation: ❌ (단일 SPEC, 단일 세션)
- Agent isolation (`Agent(isolation: "worktree")`): ❌ (run-phase는 sequential, parallel teammate 불요)

따라서 `feedback_worktree_never_use` 정책 일관 적용 — main checkout 에서 직접 작업.

---

## 6. Why "FOLLOWUP-002" Naming

### 6.1 Lineage continuity

`SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (SDF-001) 이 직계 precedent. FOLLOWUP-002 는 같은 워크플로우 패턴 (drift 노출 → 분류 → 정밀 해소) 의 두 번째 wave.

명시적 lineage:
```
SDF-001 (LSKC-001 cleanup의 후속, 64건 → 0)
SDF-002 (LSGF-001 walker fix의 후속, 17건 → 0) ← 본 SPEC
```

### 6.2 Naming alternative 고려 + 기각

- ❌ `SPEC-V3R4-STATUS-DRIFT-002`: SDF-001 의 successor 라는 의미 약화
- ❌ `SPEC-V3R4-LSGF-CLEANUP-001`: LSGF-001 의 cleanup 처럼 보이지만, LSGF-001 자체는 cleanup 불필요 (walker fix 가 의도 결과)
- ❌ `SPEC-V3R4-LINT-DRIFT-SWEEP-003`: SWEEP 표현은 일괄 처리 함의 — 본 SPEC 의 per-SPEC analysis 와 어긋남
- ✅ `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002`: lineage 명시 + FOLLOWUP 표현으로 "직접 trigger 의 후속" 함의 보존

---

## 7. Risk Decision: lint.skip 채택 기준

본 SPEC 은 `lint.skip` 사용을 보수적으로 결정. 다음 조건 모두 충족 시에만 채택:

1. **Walker visibility 한계 명확**: walker 가 인식 못하는 commit prefix 또는 매칭 패턴 (예: docs(sync), fix(hook), T-SPC002-*) 가 명백한 implementation/sync event.
2. **외부 증거로 완료 입증**: CHANGELOG, project memory, merged PR 중 최소 하나가 SPEC 완료를 명시.
3. **HISTORY entry 에 reason 기록**: 채택 사유 + 인용 PR # + walker miss 원인 모두 inline 기록.
4. **sync-commit mechanism 으로 해소 가능한가? 검토 + 채택 거부 사유 기록**: sync-commit 이 우선이지만 retroactive commit 의 commit history accuracy 손상이 더 큰 경우 lint.skip 채택.

이 4 조건 미충족 시 sync-commit 또는 frontmatter downgrade 우선.

---

## 8. Open Question Resolution Strategy

plan.md §8 OQ-1, OQ-2, OQ-3, OQ-4 (4 개 open questions) 는 모두 **run-phase Wave 1 analysis 에서 evidence-driven resolution** 으로 처리.

원칙:
- plan-phase 가설 정확도 자체는 plan-auditor 비평가 (process driven 평가)
- run-phase 결정 변경 시 plan.md §3 Category B Analysis Table inline 갱신
- sync-phase HISTORY entry 에 final decisions consolidated

---

## 9. Cross-references

- `spec.md` — REQ + AC mapping
- `plan.md` — 3-Wave workflow + Category B per-SPEC analysis
- `acceptance.md` — Binary AC 시나리오
- `tasks.md` — Wave task breakdown
- `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/spec.md` — direct trigger SPEC
- `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/plan.md` — SDF-001 precedent (Pattern A~H model)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field SSOT
