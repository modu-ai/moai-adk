# Acceptance — SPEC-WORKFLOW-RULE-SPLIT-001

측정 기준 트리: `.moai/worktrees/spec-workflow-split` @ `81786284e` (카드 t1586 worktree). 아래 RED 값은 이 트리에서 실제 관측한 값이다. 각 AC 는 관측 가능한 출력을 내는 명령으로 판정하며, 예산 단위는 hook 과 동일한 **runes**(`utf8.RuneCount`)로 통일한다 — `wc -c` bytes 와 혼용 금지.

---

## §D AC 매트릭스

| AC | 요구사항 | RED (현 트리) | GREEN (목표) |
|---|---|---|---|
| AC-01 | REQ-003 | core 40,081 runes (초과) | 양쪽 트리 core ≤ 39,999 (목표 ~28.8K) |
| AC-02 | REQ-004 | 쌍 byte-identical (현행 유지) | core·companion 양쪽 쌍 모두 `cmp` 무출력 |
| AC-03 | REQ-004 | mirror 테스트: core 등재 / companion 미등재 | mirror 테스트 exit 0 (companion 등재 포함) |
| AC-04 | REQ-001 | 등록 clause 3건 core 존재 | 3건 모두 `grep -F -c` ≥ 1 (core) + 앵커 헤딩 존재 |
| AC-05 | REQ-007 | Tier 파서 통과 (표 core 존재) | `go test ./internal/spec/` exit 0 + "Artifact set" 헤더 core 존재 |
| AC-06 | REQ-006 | `§ Report Persistence` 인용이 spec-workflow.md 를 향함 (4처) | 잔존 0건 + companion 인용 4처 + 앵커 링크 57건 전부 해소 |
| AC-07 | REQ-002, 008 | — (companion 부재) | 이동 블록 시그니처 원문 companion 존재 + detail-companion 계약 구색 |
| AC-08 | REQ-005 | embed manifest 에 companion 없음 | `make embed-manifest-check` exit 0 + embed 라인 존재 |
| AC-09 | REQ-007 | — | constitution/hook/template 회귀 배치 exit 0 |
| AC-10 | REQ-007 | — | `make build` exit 0 |

---

## §D.1 AC 상세

### AC-01 — 예산 판정 (runes, 양쪽 트리)

- **판정 명령** (4행 모두 관측, 전문을 progress.md §E.2 에 인용):

```bash
python3 -c "print(len(open('.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"
python3 -c "print(len(open('internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"
wc -m .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
go test ./internal/hook/ -run TestInstructionsLoaded -count=1   # 회귀만 — 실측 아님(픽스처 기반)
```

- **GREEN**: 파이썬 두 값 모두 `39999` 이하. `go test` exit 0.
- **RED**: `40081`, `40081` (실측 2026-10-08, `81786284e`).
- 예산 근거: `internal/hook/instructions_loaded.go:124` `const charBudget = 40000`. **CI 예산 테스트는 존재하지 않는다**(카드 후보명 `TestDeployedAlwaysLoadedCharBudget` — 전수 grep 0 적중, plan.md §A). 본 AC 가 유일한 예산 판정이다.

### AC-02 — mirror byte-parity (쌍 2개)

- **판정 명령**:

```bash
cmp .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md && echo PAIR1_OK
cmp .claude/rules/moai/workflow/spec-workflow-detail.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow-detail.md && echo PAIR2_OK
```

- **GREEN**: `PAIR1_OK` + `PAIR2_OK` 출력, exit 0.

### AC-03 — mirror 테스트 등재

- **판정 명령**: `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestDeclaredRuleMirrorForks' -count=1`
- **GREEN**: exit 0. `rule_template_mirror_test.go` 의 `workflowOptMirroredPaths` 가 신규 companion 경로를 포함한다(`grep -c "spec-workflow-detail" internal/template/rule_template_mirror_test.go` ≥ 1).

### AC-04 — 등록 clause·앵커 core 잔류 (mutant 방어 1)

- **판정 명령** (core 파일에 대해; 3 clause 는 zone-registry 원문과 byte 동일):

```bash
for c in "Create comprehensive specification using EARS format." \
         "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step" \
         "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"; do
  grep -F -c -- "$c" .claude/rules/moai/workflow/spec-workflow.md
done
grep -c "^## Plan Phase$" .claude/rules/moai/workflow/spec-workflow.md
grep -c "^## SPEC Phase Discipline$" .claude/rules/moai/workflow/spec-workflow.md
grep -c "the ONE.*authoritative skip contract\|authoritative skip contract" .claude/rules/moai/workflow/spec-workflow.md
```

- **GREEN**: clause 3행 각각 ≥ 1, 헤딩 2행 각각 1, skip 계약 ≥ 1.
- **최종 판정**: `make constitution-check` → `constitution-check: OK` (bin/moai 경유 실검증).

### AC-05 — Go Tier 파서 생존 (mutant 방어 2)

- **판정 명령**:

```bash
go test ./internal/spec/ -count=1
grep -c '"Artifact set"\|Artifact set' .claude/rules/moai/workflow/spec-workflow.md
```

- **GREEN**: go test exit 0 (`TierArtifactMissingRule`·`TierArtifactSetUnreadable` 회귀 포함), 헤더 존재 ≥ 1. 근거: `internal/spec/lint_tier_artifacts.go:36` 이 이 경로의 파일을 직접 읽는다.

### AC-06 — 크로스레퍼런스 해소 (mutant 방어 5)

- **판정 명령**:

```bash
# (a) 이동 섹션의 낡은 인용 잔존 0건 — 로컬+템플릿 전체
grep -rn "spec-workflow\.md.*§ Report Persistence" --include="*.md" .claude/ internal/ | wc -l          # → 0
# (b) 새 인용 4처 (2 소스 × 2트리)
grep -rln "spec-workflow-detail\.md § Report Persistence" .claude/agents/moai/plan-auditor.md .claude/skills/moai/workflows/run/phase-execution.md internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md | wc -l   # → 4
# (c) 앵커 링크 대상 헤딩이 core 에 존재
grep -c "^## Subcommand Classification (Pipeline vs Multi-Agent)$" .claude/rules/moai/workflow/spec-workflow.md   # → 1
grep -c "^### Mode Dispatch Cross-Reference$" .claude/rules/moai/workflow/spec-workflow.md                        # → 1
grep -c "^## Phase Overview$" .claude/rules/moai/workflow/spec-workflow.md                                        # → 1
# (d) 방법론 안내 갱신 (×2트리)
grep -rln "spec-workflow-detail" .claude/skills/moai/workflows/run/context-loading.md internal/template/templates/.claude/skills/moai/workflows/run/context-loading.md | wc -l   # → 2
# (e) core 안 이동-섹션 포인터 — 헤더 blockquote 포함 최소 6건 (5 블록 포인터 + skip 계약의 § Report Persistence 갱신)
grep -c "spec-workflow-detail" .claude/rules/moai/workflow/spec-workflow.md
```

- **GREEN**: (a) `0`, (b) `4`, (c) 각 `1`, (d) `2`, (e) `6` 이상.
- RED 대비: (a)는 현 트리에서 `4`(plan-auditor.md ×2, phase-execution.md ×2 — 실측).

### AC-07 — companion 원문 보존 + 계약 구색 (mutant 방어 3·6)

- **판정 명령** (companion 파일에 대해; 시그니처 = 이동 블록 각각의 고유 문장):

```bash
for s in "configured \`merge_method\`" \
         "Creating an L2/L3 worktree for plan (Step 1)" \
         "Best for existing projects with < 10% test coverage" \
         "Write a failing test describing desired behavior" \
         "Load the SPEC's frontmatter \`depends_on:\` list" \
         "Two report streams exist for plan audits" \
         "SPEC-AUDIT-SNAPSHOT-001 A1"; do
  grep -F -c -- "$s" .claude/rules/moai/workflow/spec-workflow-detail.md
done
grep -c "Detail companion of \`spec-workflow.md\`" .claude/rules/moai/workflow/spec-workflow-detail.md
grep -c "^paths:" .claude/rules/moai/workflow/spec-workflow-detail.md
```

- **GREEN**: 시그니처 7행 각각 ≥ 1, 계약 blockquote 1, frontmatter `paths:` 1.
- verbatim 보존 판정: M1 이동 시 각 블록을 원본 라인 범위에서 그대로 잘라 옮긴다. 판정은 위 시그니처 + sync-phase 리뷰가 원본(L범위는 plan.md §B M1)과 대조. core 잔류 블록의 보존은 AC-04/05/06의 core grep 이 대변한다.

### AC-08 — embed manifest 재생성

- **판정 명령**:

```bash
make embed-manifest-check
grep -c "spec-workflow-detail" internal/template/embed_manifest_gen.go
```

- **GREEN**: make exit 0, grep ≥ 1.

### AC-09 — 소비자 회귀 배치 (병렬 일괄)

- **판정 명령** (한 턴 multi-Bash):

```bash
go test ./internal/constitution/ -count=1
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak|TestDeclaredRuleMirrorForks|TestContractModeAlwaysLoadedBudget' -count=1
go test ./internal/hook/ -run TestInstructionsLoaded -count=1
go test ./internal/spec/ -count=1
```

- **GREEN**: 전부 exit 0.

### AC-10 — 최종 빌드 게이트

- **판정 명령**: `make build`
- **GREEN**: exit 0 (`embed-manifest-check`·`agents-emit-check`·`commands-emit-check`·`tool-policy-drift-check` 포함).

---

## §E 공통 조항

- 각 판정 명령의 출력은 progress.md §E.2 에 전문 인용(50줄 상한 재단 규칙은 agent-common-protocol § Parallel Execution 준수). "통과했다" 요약만으로는 Claim 아님(`verification-claim-integrity.md` §2).
- 재사용 금지: `moai verify run` 으로 재사용 결과를 인용할 경우 그 key 와 `recorded_at` 을 명시하고 "output not re-observed" 를 Gaps 에 기록한다(AGENTS.md §4).
- FAIL 시: AC-01 제외 전부 재측정 명령 재실행으로 재판정 가능. AC-04/05 FAIL 은 core 경계 위반(plan.md §2.2 표)이므로 분할 재조정 대상 — companion 으로의 추가 이동으로 해소 금지.
