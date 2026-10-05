# SPEC 검토 보고서: SPEC-DOCTOR-TEST-CWD-ISOLATION-001

- Iteration: 1/1
- Verdict: **FAIL**
- Overall Score: **0.67**
- Tier S PASS threshold: **0.75**
- Plan Artifact Hash: `73f80cd864fe2973bc70abcd08ff54c0d0cb74416c67597f21d18042fc68362e`
- Audit export: plan-auditor가 판정 텍스트를 반환했으나 상위 실행 규칙 때문에 파일을 직접 쓸 수 없었다. 이 파일은 lane이 반환된 판정과 증거를 그대로 보존하기 위해 작성했다. 감사나 SPEC 재실행은 하지 않았다.

## Must-Pass Results

| 기준 | 판정 | 증거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | PASS | `spec.md:L51,L56,L62,L68`의 `001`~`004`가 순차적이며 중복이 없음 |
| MP-2 GEARS 형식 | **FAIL** | `REQ-DTC-002`와 `REQ-DTC-003`이 각각 두 개의 독립 규범 문장을 포함함 |
| MP-3 YAML frontmatter | PASS | `spec.md:L2-L13`; strict SPEC lint가 `[]`, exit 0 |
| MP-4 언어 중립성 | N/A | `internal/cli`의 Go 테스트만 다루는 단일 언어 SPEC |
| MP-5 D7 교차 SPEC | PASS | 자기 SPEC 참조만 존재하며 상태가 `draft`; retired/superseded/archived 참조 없음 |
| MP-6 D8 교차 플랫폼 | PASS | `syscall_matches=0` |
| MP-7 clarification gate | PASS | `plan.md` 존재, Tier S라 `research.md` 없음, `[NEEDS CLARIFICATION` 일치 0 |
| MP-8 RED 재실행 | PASS | E-RED-001과 E-RED-002가 기준 SHA에서 exit 1로 재현됨 |
| MP-9 교차 산출물 순서 | GAP | 반환된 감사 결과에 CN-4/MP-9 판정이 포함되지 않아 PASS로 기록하지 않음 |

## Category Scores

| 평가 항목 | 점수 | 근거 |
|---|---:|---|
| Clarity | 0.50 | `AC-DTC-001`의 기준 SHA와 기대 exit가 모순되며 cleanup 표현도 모호함 |
| Completeness | 1.00 | HISTORY, 문제·범위, 요구사항, inline AC, Out of Scope, frontmatter가 모두 존재함 |
| Testability | 0.50 | AC-DTC-001의 기준 상태 모순, assertion·복원 mutant 공백, E-RED-002 stdout 요약 |
| Traceability | 1.00 | 두 AC의 `Covers`가 실재 REQ를 가리키고 모든 REQ가 형식상 연결됨 |

조화평균:

```text
4 / (1/0.50 + 1/1.00 + 1/0.50 + 1/1.00) = 0.6667 ≈ 0.67
```

## Defects Found

D1. **MP2-GEARS-MULTICLAUSE** — `spec.md:L58-L60,L64-L66` — `REQ-DTC-002`와 `REQ-DTC-003`이 각각 두 독립 규범 문장을 한 REQ ID에 묶어 MP-2를 위반함. **Severity: critical; Class: blocking.** 각 REQ를 하나의 `shall` 절로 정리하거나 독립 REQ로 분리하고 번호 및 `Covers`를 갱신해야 함.

D2. **AC-BASELINE-CONTRADICTION** — `spec.md:L79-L82,L101-L117` — AC-DTC-001은 기준 SHA `74d872aafbd90235e67163a5bc233f7c8a934491`을 Given으로 고정하면서 exit 0을 요구하지만, 같은 SHA의 E-RED-001과 감사 재실행은 exit 1을 반환함. **Severity: major; Class: blocking.** AC의 Given을 구현 descendant로 명시하거나 기준 SHA를 RED ledger에만 둬야 함.

D3. **RED-LEDGER-NONVERBATIM** — `spec.md:L124-L128` — release-blocking E-RED-002가 원문 stdout 대신 요약과 마지막 verdict만 기록함. `.claude/rules/moai/development/verification-completeness.md §2.1`의 verbatim stdout 요소가 빠짐. **Severity: major; Class: blocking.** 전체 stdout을 fenced ledger 또는 영속 raw evidence로 보존하고 E-RED-002가 직접 인용해야 함.

D4. **AC-MUTANT-COVERAGE-GAP** — `spec.md:L64-L71,L84-L92,L133-L135` — `Covers`는 REQ-DTC-003/004를 포함하지만 AC-DTC-002는 테스트 이름·exit·변경 파일만 확인함. 기존 assertion 삭제 mutant나 framework cleanup 없는 수동 `os.Chdir` mutant를 거부하지 못함. **Severity: major; Class: blocking.** assertion 보존과 framework-managed CWD 복원을 판정하는 실행 가능한 조건을 추가해야 함.

D5. **REQ-CLEANUP-WORDING-AND-HOW** — `spec.md:L68-L71` — 임시 디렉터리를 복원하는지 제거하는지 불명확하고, `by the Go test framework`가 requirement layer에 구현 수단을 고정함. **Severity: major; Class: blocking.** 임시 디렉터리 제거와 원래 CWD 복원이라는 관찰 가능한 결과로 다시 쓰고 구현 수단은 plan/constraint로 옮겨야 함.

## Evidence

### Claim

- 기준 SHA에서 E-RED-001과 E-RED-002가 의도한 RED로 재현됨.
- 정확한 selector가 9개 테스트를 선택함.
- frontmatter와 기계적 SPEC lint에 오류가 없음.
- lifecycle drift, D7 BLOCKING, D8 BLOCKING, clarification marker가 없음.
- MP-2와 acceptance testability 결함 때문에 최종 판정은 FAIL임.

### Evidence 1 — baseline and tree

Command:

```bash
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 rev-parse --show-toplevel && git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 rev-parse HEAD && git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 branch --show-current
```

Observed output:

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675
74d872aafbd90235e67163a5bc233f7c8a934491
WT-doctor-red
```

### Evidence 2 — strict SPEC lint

Command:

```bash
moai spec lint SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json
```

Observed output and exit code:

```json
[]
```

`exit 0`

### Evidence 3 — lifecycle audit

Command/Tool invocation:

```text
mcp__moai__spec_audit(
  filter_spec="SPEC-DOCTOR-TEST-CWD-ISOLATION-001",
  include_grandfathered=true,
  project_root="/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675"
)
```

Observed output:

```json
{
  "total_specs": 1,
  "grandfathered": 0,
  "modern_era_clean": 1,
  "drift_findings": [
    {
      "spec_id": "SPEC-DOCTOR-TEST-CWD-ISOLATION-001",
      "era": "V3R6",
      "finding_type": "EraAutoDetected",
      "severity": "INFO",
      "details": {"heuristic_matched": "H-5 (modern phase or created date)"}
    }
  ]
}
```

`exit 0`

### Evidence 4 — structural counts

Observed output:

```text
REQ=REQ-DTC-001,REQ-DTC-002,REQ-DTC-003,REQ-DTC-004
AC=AC-DTC-001,AC-DTC-002
counts req=4 ac=2
syscall_matches=0
SPEC_refs=SPEC-DOCTOR-TEST-CWD-ISOLATION-001
D7 SPEC-DOCTOR-TEST-CWD-ISOLATION-001 exists=True status=draft
```

GEARS clause count:

```text
REQ-DTC-001 shall_count=1 when=True while=False shall_not=False
REQ-DTC-002 shall_count=2 when=False while=True shall_not=True
REQ-DTC-003 shall_count=2 when=False while=False shall_not=True
REQ-DTC-004 shall_count=1 when=True while=False shall_not=False
```

AC traceability:

```text
AC-DTC-001 given=True when=True then=True covers=['REQ-DTC-001', 'REQ-DTC-002', 'REQ-DTC-004'] valid=True
AC-DTC-002 given=True when=True then=True covers=['REQ-DTC-001', 'REQ-DTC-002', 'REQ-DTC-003', 'REQ-DTC-004'] valid=True
uncovered=[]
weasel=[]
```

### Evidence 5 — E-RED-001 re-execution

Command:

```bash
MOAI_EMBED_CHECK_BIN=/usr/bin/false go test ./internal/cli -count=1 -run '^TestDoctorCmd_Execution$'
```

Observed output and exit code:

```text
--- FAIL: TestDoctorCmd_Execution (18.67s)
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	19.761s
FAIL
```

`exit 1`

### Evidence 6 — E-RED-002 re-execution

Command:

```bash
MOAI_EMBED_CHECK_BIN=/usr/bin/false go test ./internal/cli -count=1 -run '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$'
```

Observed output and exit code:

```text
--- FAIL: TestRunDoctor_WithExport (22.01s)
    coverage_improvement_test.go:715: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_WithFix (20.25s)
    coverage_improvement_test.go:737: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_Verbose (23.05s)
    coverage_improvement_test.go:777: runDoctor error: doctor: 1 check(s) failed
  ○ Go Runtime
  ✓ Go Runtime
  ○ Git
  ✓ Git
  ○ Claude Code
  ✓ Claude Code
  ○ GitHub CLI
  ✓ GitHub CLI
  ○ ast-grep CLI
  ✓ ast-grep CLI
  ○ MoAI Config
  ✓ MoAI Config
  ○ Claude Config
  ✓ Claude Config
  ○ MoAI Version
  ✓ MoAI Version
  ○ Binary Freshness
  ✓ Binary Freshness
  ○ MCP Scope Duplicates
  ✓ MCP Scope Duplicates
  ○ MCP Server Version
  ✓ MCP Server Version
  ○ Agent Emit Embed
  ✗ Error: could not extract embedded artifacts from /usr/bin/false: false init: exit status 1 ()
  ○ Constitution Registry
  ✓ Constitution Registry
  ○ Harness 5-Layer
  ✓ Harness 5-Layer
  ○ Migration
  ✓ Migration
  ○ Plugin Deployment
  ✓ Plugin Deployment
  ○ Home Disk Usage
  ✓ Home Disk Usage
  ○ Hooks Config
  ✓ Hooks Config
  ○ Hook Wiring
  ✓ Hook Wiring
  ○ Hook Delivery
  ✓ Hook Delivery
  ○ Hook opt-in:
  ✓ Hook opt-in:
  ○ Slash Commands
  ✓ Slash Commands
  ○ Skills Allowlist
  ✓ Skills Allowlist
  ○ MX Tag Config
  ✓ MX Tag Config
  ○ Worktree State
  ✓ Worktree State
  ○ Worktree Base Branch
  ✓ Worktree Base Branch
  ○ BODP Config
  ✓ BODP Config
  ○ Telemetry Config
  ✓ Telemetry Config
  ○ Glamour Cache
  ✓ Glamour Cache
  ○ Codex Wiring
  ✓ Codex Wiring
--- FAIL: TestRunDoctor_AllFlags (20.33s)
    coverage_improvement_test.go:4930: unexpected error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_VerboseAndDetail (21.23s)
    coverage_improvement_test.go:5754: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_ExportMode (23.62s)
    coverage_improvement_test.go:5804: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_Execution (24.58s)
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_ExportFlag (21.17s)
    integration_test.go:176: doctor --export error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_VerboseExecution (19.36s)
    integration_test.go:202: doctor --verbose error: doctor: 1 check(s) failed
FAIL
github.com/modu-ai/moai-adk/internal/cli	196.322s
FAIL
```

`exit 1`

### Evidence 7 — non-empty selector

Command:

```bash
go test ./internal/cli -list '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$'
```

Observed output:

```text
TestRunDoctor_WithExport
TestRunDoctor_WithFix
TestRunDoctor_Verbose
TestRunDoctor_AllFlags
TestRunDoctor_VerboseAndDetail
TestRunDoctor_ExportMode
TestDoctorCmd_Execution
TestDoctorCmd_ExportFlag
TestDoctorCmd_VerboseExecution
ok  	github.com/modu-ai/moai-adk/internal/cli	0.804s
```

### Evidence 8 — current worktree after audit

Command:

```bash
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 status --short && git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 rev-parse HEAD
```

Observed output:

```text
?? .moai/reports/t675/
?? .moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/
74d872aafbd90235e67163a5bc233f7c8a934491
```

## Baseline-attribution

측정 대상은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675`의 `WT-doctor-red` 브랜치이며, 모든 위 명령은 `74d872aafbd90235e67163a5bc233f7c8a934491` HEAD에서 실행됐습니다. 감사가 계산한 plan-artifact hash는 `73f80cd864fe2973bc70abcd08ff54c0d0cb74416c67597f21d18042fc68362e`입니다.

## Gaps

- 구현 전 감사이므로 GREEN 결과는 실행하지 않았습니다.
- 저장소 전체 `go test ./...`는 실행하지 않았습니다.
- `go vet`, `golangci-lint`, coverage는 run-phase 검증 항목이라 실행하지 않았습니다.
- 감사 반환문에 MP-9의 CN-4 결과가 없어 해당 기준을 PASS로 주장하지 않았습니다.
- 테스트 전후 complete porcelain-v2 tree key는 별도로 측정하지 않았습니다. 따라서 검증 명령이 파일을 전혀 건드리지 않았다고 주장하지 않습니다.

## Residual-risk

- REQ와 AC를 고친 뒤에도 구현자가 assertion을 약화하거나 9개 중 일부만 격리할 수 있습니다. 수정된 AC에는 assertion 보존, CWD 복원, non-empty selector를 함께 판정하는 실행 가능한 조건이 필요합니다.
- `/usr/bin/false` RED 명령은 현재 Darwin worktree에서 재현됐지만, 다른 운영체제에서 같은 shell 명령이 동일하게 실행된다는 점은 이번 감사 범위에서 검증하지 않았습니다.
- Tier S의 감사 ceiling은 1회이므로 이번 FAIL을 자동 iteration 2로 넘기지 않습니다. 결함을 수정한 뒤 별도 감사 승인이 필요합니다.

## Recommendation

1. `REQ-DTC-002`와 `REQ-DTC-003`을 각각 하나의 GEARS `shall` 절로 정리하거나 독립 REQ로 분리하고 AC의 `Covers`를 갱신합니다.
2. AC-DTC-001에서 기준 커밋과 구현 descendant의 실행 상태를 구분합니다.
3. E-RED-002의 전체 stdout을 원문으로 보존하거나 영속 evidence 경로를 연결합니다.
4. 9개 테스트의 기존 assertion 보존과 framework-managed CWD 복원을 확인하는 binary 조건을 추가합니다.
5. REQ-DTC-004는 임시 디렉터리 제거와 원래 CWD 복원이라는 결과로 쓰고 구현 수단은 plan/constraint로 이동합니다.

**최종 판정: FAIL, 0.67.**
