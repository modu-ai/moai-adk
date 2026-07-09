# SPEC-WORKFLOW-LIFECYCLE-001 — Research

> Tier L research artifact. 본 파일은 모든 gap 주장을 codebase grep/Bash 실측으로 뒷받침한다 (`verification-claim-integrity.md` §1.1 surface 3 — "defect claim은 도구 검증 전까지 가설"). 2026-07-09 기준 실측.

## §A Research Methodology

모든 아티팩트는 다음 우선순위로 실측:
1. **Grep/Bash 직접 실행** — `grep -rn`, `sed -n`, `cat -n`을 통한 verbatim 증거 수집
2. **file:line 인용** — 라인 번호 drift 가능성 인지, content-token 앵커 병기
3. **Go 코드 실측 우선** — doctrine 서술과 Go 구현이 다를 경우 **Go를 정합의 SSOT**로 취급 (Go가 실제 동작이므로)
4. **Hypothesis-as-defect 회피** — memory 회상으로 gap을 주장하지 않음 (feedback_hypothesis_as_defect); 모든 주장은 이 세션에서 직접 실행한 명령 출력에 기반

## §B Codebase Reality — R1 (delta-spec 수명주기)

### B.1 Status Enum (8값, `amended` 부재)

**명령**: `grep -n 'ValidStatuses' internal/spec/status.go` + `sed -n '34,45p'`

**실측** (`internal/spec/status.go:34-44`):
```go
// ValidStatuses defines all allowed status values
var ValidStatuses = []string{
	"draft",
	"planned",
	"in-progress",
	"implemented",
	"completed",
	"superseded",
	"archived",
	"rejected",
}
```

**발견**: `amended` 값 부재. 8개 상태만 존재. `completed`는 terminal status (아래 B.2 확인).

### B.2 completed no-drift predicate

**명령**: `grep -n 'completed' internal/spec/audit.go` + `sed -n '350,380p'`

**실측** (`internal/spec/audit.go:355-377`):
```go
	// If status is already completed, no drift.
	if specStatus == "completed" {
		return nil
	}
	// ... otherwise check §E.2 + §E.4 + sync_commit_sha predicate
```

**발견**: `completed` 상태에서는 drift detection이 조기 종료. 이는 completed가 terminal이므로 추가 drift 검사가 무의미하다는 정합적 동작. **R1의 in-place 개정은 completed에서 in-progress로 전이하므로 이 predicate가 발화하지 않고 정상 drift detection이 재개** — Go 변경 불필요.

### B.3 Status Transition Ownership Matrix — `completed → *` 경로 부재

**명령**: `grep -n 'completed\|Status Transition' .claude/rules/moai/development/spec-frontmatter-schema.md` + 라인 범위 실측

**실측** (`spec-frontmatter-schema.md` § Status Transition Ownership Matrix):

| Transition | Owning agent | Canonical commit subject pattern |
|------------|--------------|----------------------------------|
| `(none) → draft` | manager-spec | `feat(SPEC-{ID}): plan-phase artifacts ...` |
| `draft → in-progress` | manager-develop (on M1) | `fix(SPEC-{ID}): M1 ...` |
| `in-progress → implemented → completed` | manager-docs (sync commit) | `docs(SPEC-{ID}): sync-phase artifacts` |
| `* → superseded` | manager-spec | `feat(SPEC-{NEW-ID}): supersedes SPEC-{OLD-ID}` |
| `* → archived` | manager-docs | `chore(specs): archive SPEC-{ID}` |
| `* → rejected` | orchestrator, recorded by manager-docs | `chore(SPEC-{ID}): rejected per <rationale>` |

**발견**: `completed → *` 경로 전무. `* → superseded/archived/rejected`는 있으나 `completed → in-progress` (amendment) 경로는 부재. 이것이 R1의 gap — completed SPEC의 사후 개정 전이가 정식으로 소유자/커밋 패턴과 함께 정의되어 있지 않다. 본 SPEC이 이 경로를 추가 (REQ-WFL-003).

### B.4 superseded_by / partially_superseded_by 이미 존재

**명령**: `grep -rn 'superseded_by\|partially_superseded' .claude/agents/moai/manager-spec.md`

**실측** (`.claude/agents/moai/manager-spec.md:217-218`):
```
- `superseded_by: SPEC-NEW-001` — When status=superseded.
- `partially_superseded_by: [SPEC-A-001]` — Partial supersession.
```

**발견**: 두 필드는 이미 Optional 필드로 문서화. 본 SPEC의 `amendment_of:`는 세 번째 lineage 필드로 추가되며, 의미가 다름:
- `superseded_by`: wholesale replacement (원본이 폐기됨)
- `partially_superseded_by`: 부분 대체 (여러 후속이 원본의 일부를 대체)
- `amendment_of:`: lineage link (원본 또는 자기 자신을 개정) — replacement가 아닌 revision

## §C Codebase Reality — R2 (depends_on run-phase 강제)

### C.1 depends_on의 유일 소비자 — BODP Signal A

**명령**: `grep -rn 'depends_on' internal/` + `grep -n 'checkSignalA\|DependsOn' internal/bodp/relatedness.go`

**실측** (`internal/bodp/relatedness.go:128, 177-180`):
```go
// checkSignalA detects code-level dependency: SPEC depends_on heuristic OR
// ...
DependsOn []string `yaml:"depends_on"`
```

`internal/bodp/relatedness.go:180` (함수 시그니처): `parseDependsOn`는 `.moai/specs/<specID>/spec.md`를 읽어 depends_on 리스트를 파싱.

`internal/bodp/relatedness_test.go:30` (주석): "depends_on list at .moai/specs/<specID>/spec.md under repoRoot."

**발견**: depends_on의 유일 소비자는 BODP Signal A (branch-origin 결정). "이 SPEC이 현재 브랜치 이름과 코드 의존 관계가 있는가?"를 판단할 때 사용. **run-phase 진입 허가와 무관** — BODP는 "브랜치 만들까?"를 결정하지 "run 할까?"를 결정하지 않음.

### C.2 Phase 0.5 / runtime audit gate의 depends_on 참조 부재

**명령**: `grep -rn 'depends_on' internal/runtime/`

**실측**: (empty)

**발견**: `internal/runtime/audit_cache.go`, `audit_gate.go`, `audit_report.go` 어디에서도 depends_on 참조 없음. Phase 0.5 Plan Audit Gate는 depends_on을 소비하지 않는다 — 이것이 R2의 gap. depends_on은 declarative-only이며 run-phase 강제 메커니즘이 전무.

### C.3 spec-frontmatter-schema.md의 depends_on Optional 필드 설명

**명령**: `grep -n 'depends_on' .claude/rules/moai/development/spec-frontmatter-schema.md`

**실측** (`spec-frontmatter-schema.md` § Optional Fields):
```
| `depends_on` | list | SPEC IDs this SPEC depends on. Used by BODP signal A. |
```

**발견**: doctrine은 "Used by BODP signal A"라고만 서술 — run-phase 강제 언급 없음. 본 SPEC이 이 설명을 확장하여 "BODP Signal A + Phase 0.5 Depends_on Pre-flight Check"로 명문화 (REQ-WFL-005).

## §D Codebase Reality — R3 (Tier L 산물 + plan-auditor 입력 계약 + hash drift)

### D.1 Tier L 5-file 집합 — spec-workflow.md에 서술 존재

**명령**: `grep -n 'Tier L\|5 files\|5-file' .claude/rules/moai/workflow/spec-workflow.md`

**실측** (`spec-workflow.md:137`):
```
| L (Large) | > 1000 LOC or constitutional | > 15 files | **5 files**: spec.md + plan.md + acceptance.md + design.md + research.md | 0.85 |
```

**발견**: Tier L 5-file 집합은 이미 spec-workflow.md § SPEC Complexity Tier에 서술됨. R3a의 gap은 "서술이 없다"가 아니라 "서술이 SSOT(spec-frontmatter-schema.md `tier:` 필드)에 명시적이지 않다" — 독자가 다른 파일로 교차 참조해야 한다. 본 SPEC이 `tier:` 필드 설명에 5-file 목록을 직접 임베드 (REQ-WFL-008).

### D.2 plan-auditor Input Contract — design.md/research.md 언급 부재

**명령**: `grep -n 'design\.md\|research\.md' .claude/agents/moai/plan-auditor.md`

**실측**: (empty — design.md/research.md에 대한 언급이 plan-auditor.md 전체에서 0건)

**명령**: `sed -n '468,480p' .claude/agents/moai/plan-auditor.md`

**실측** (`plan-auditor.md § Input Contract, L470-476`):
```
This agent receives one input: the absolute path to the SPEC directory (e.g., `.moai/specs/SPEC-AUTH-001/`).

The agent reads `spec.md` as the primary input. It may also read `acceptance.md` and `plan.md` for cross-reference.

If the caller passes additional context (author reasoning, prior conversation), the agent MUST ignore it and state: "Reasoning context ignored per M1 Context Isolation."
```

**발견**: Input Contract는 "spec.md primary + MAY read acceptance/plan"으로 제한. **design.md와 research.md는 Tier L 산물임에도 계약에서 누락**. 이것이 R3b의 gap. 본 SPEC이 Tier-differentiated 계약으로 재작성 (REQ-WFL-009).

### D.3 plan-artifact hash subject list — doctrine vs Go drift

**명령**: `grep -n 'planArtifactNames' internal/runtime/audit_cache.go` + `sed -n '60,70p'`

**실측** (`internal/runtime/audit_cache.go:61-68`):
```go
// planArtifactNames is the ordered list of plan artifact file names to hash.
// OPEN QUESTION Q1 resolution: whitespace-insensitive, sorted by filename.
var planArtifactNames = []string{
	"acceptance.md",
	"plan.md",
	"spec.md",
	"tasks.md",
}
```

**명령**: `sed -n '310,325p' .claude/rules/moai/workflow/spec-workflow.md`

**실측** (`spec-workflow.md:313-319`, § Phase Transitions skip policy):
```
   artifact-hash unchanged since that verdict — no plan-phase artifact
   (spec.md / plan.md / acceptance.md / research.md / design.md) has been
   modified since the audit that produced the verdict (equivalently on
   Route B: no plan-PR commit has landed since that verdict). Note: the
   mechanical hash subject is the `ComputeHash` 4-file plan-artifact set
   (spec.md / plan.md / acceptance.md / tasks.md — see § Report
   Persistence); research.md / design.md changes are a conservative input
   to the manual skip judgment, not part of the mechanical plan-artifact
   hash.
```

**발견**: doctrine은 이미 4-file vs 5-file의 차이를 인지하고 서술 (선행 SPEC-AUDIT-GATE-INTEGRITY-001 M3.4의 산물). 그러나 서술이 산만하고 § Report Persistence(정식 SSOT surface)가 아니라 § Phase Transitions skip policy에 숨어 있다. 본 SPEC은 이 서술을 § Report Persistence로 끌어올리고 (REQ-WFL-010), 4-file 집합을 verbatim Go 정합으로 고정하며, design/research를 `manual-skip judgment inputs` 리터럴 토큰으로 명명.

### D.4 tasks.md는 Tier L 산물 아님

**명령**: `grep -n 'tasks\.md' .claude/rules/moai/workflow/spec-workflow.md` (Tier 정의 표)

**실측**: Tier L 5-file 집합 = `spec.md + plan.md + acceptance.md + design.md + research.md` — **tasks.md는 없음**.

**발견**: `audit_cache.go` `planArtifactNames`에는 `tasks.md`가 포함되지만 Tier L 5-file 집합에는 없다. tasks.md는 V3R4-era plan artifact 이름 (과거 SPEC들 중 `.moai/specs/SPEC-V3R4-*/tasks.md` 형태로 잔존). V3R6 Tier L은 design.md + research.md로 대체. hash에는 backward compat을 위해 tasks.md가 남아 있음 — V3R6 신규 SPEC이 tasks.md를 안 갖으면 hash는 존재하지 않는 파일을 무시(Go 구현이 이미 그렇게 동작). 이 사실을 본 SPEC이 서술 (REQ-WFL-010).

## §E Cross-Verification — R1+R3 교차 (amendment시 hash 변경)

### E.1 spec.md가 hash subject이므로 amendment시 cache invalidate

**명령**: 위 D.3 `planArtifactNames` 확인 — `spec.md` 포함.

**발견**: `spec.md`는 4-file hash subject에 포함. R1의 in-place amendment는 `spec.md`를 수정하므로 (HISTORY `## Amendments` 행 추가 + version bump), hash가 변경 → cached PASS verdict invalidate → 다음 `/moai run`에서 Phase 0.5 재실행. 이는 자연스러운 결과이지만 doctrine에 명시적 서술이 필요 (REQ-WFL-011, 리터럴 토큰 `cache-invalidating event`).

## §F Findings Summary

| ID | Gap | Evidence | Covered REQ |
|----|-----|----------|-------------|
| F-1 | completed → * 전이 경로 Status Transition Ownership Matrix 부재 | `spec-frontmatter-schema.md` Matrix에서 completed 출발 경로 0건 | WFL-001, WFL-003 |
| F-2 | depends_on run-phase 강제 메커니즘 부재 | `internal/runtime/`에서 depends_on 참조 0건 (BODP만 소비) | WFL-005, WFL-006, WFL-007 |
| F-3 | Tier L 5-file 집합이 SSOT `tier:` 필드에 명시적이지 않음 | spec-workflow.md:137에만 서술; spec-frontmatter-schema.md Optional Fields에는 산물 목록 부재 | WFL-008 |
| F-4 | plan-auditor Input Contract에서 design/research 누락 | `plan-auditor.md` 전체에서 design.md/research.md 언급 0건 | WFL-009 |
| F-5 | plan-artifact hash subject list의 doctrine-vs-Go drift | doctrine 5-file vs Go 4-file; 서술이 § Report Persistence가 아닌 § Phase Transitions에 산재 | WFL-010 |
| F-6 | amendment시 hash invalidation이 doctrine에 명시적이지 않음 | spec.md가 hash subject이지만 amendment시 cache 거동 서술 부재 | WFL-011 |

## §G Tool Inventory (본 research에 사용)

- `grep` (BSD grep 2.6.0 + ugrep 7.5.0 호환 — D1/D2 교훈 계승: 표 셀 내 ERE `\|` 금지)
- `sed -n` (라인 범위 추출)
- `cat -n` (file:line 인용)
- `awk` (집계)

## §H Gaps (미검증 — 명시)

- **G1**: R1 amendment의 실제 runtime 검증은 후속 Go 구현 필요 — 본 SPEC은 doc-only이므로 "amendment_of: 필드가 주어졌을 때 audit이 올바르게 분류하는가?"는 런타임 관측 불가 (doctrine 서술만으로 동작이 보장되지 않음 — 후속 SPEC이 Go를 배선할 때 검증)
- **G2**: R2 depends_on pre-flight의 실제 orchestrator 구현은 미배선 — 본 SPEC은 spec-workflow.md 서술만 추가; `/moai run`이 실제로 이 sub-step을 실행하는지는 런타임 관측 불가
- **G3**: R3 plan-auditor가 Tier L 감사에서 실제로 design/research를 읽는지는 런타임 관측 불가 — 본 SPEC은 계약을 명시할 뿐, plan-auditor agent의 runtime 행동은 다음 감사 호출 시 관측 가능
- **G4**: `amendment_of:` 필드의 순환 참조 (A가 B를, B가 A를)는 본 SPEC이 다루지 않음 — R2와 마찬가지로 사이클 탐지는 Out of Scope

## §I Residual Risk

- amendment 전이가 in-progress를 재사용하므로, 일반적인 진행 중 작업과 amendment를 구분하는 것은 오직 `amendment_of:` 필드와 HISTORY `## Amendments` 행에 의존. 이 필드들이 누락되면 감사/드리프트 탐지가 "amendment"로 인식하지 못한다 — plan-auditor MP-3 frontmatter 검사가 이를 잡아내야 한다 (후속 검증에서 확인 필요).
- doctrine-vs-Go drift (F-5)를 Go 정합으로 명문화하므로, 향후 Go가 5-file로 확장되지 않는 한 design/research 변경은 여전히 수동 skip 판단에만 영향을 미친다 — 자동 cache invalidation이 필요하면 별도 SPEC이 Go를 변경해야 한다.
- `--ignore-deps` override가 남용되면 depends_on 집행이 사실상 무력화될 수 있다 — `.moai/logs/depends-on-override.log` 감사 가능성을 높이는 후속 모니터링 권장 (본 SPEC 범위 밖).
