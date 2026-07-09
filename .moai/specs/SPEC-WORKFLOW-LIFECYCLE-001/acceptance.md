# SPEC-WORKFLOW-LIFECYCLE-001 — Acceptance Criteria

> 모든 AC는 기계 검증 가능 (명령 + 기대 출력). `verification-claim-integrity.md` §3.2 — Evidence는 verbatim 출력이며 요약 불가.
>
> 명령 표기 규약 (SPEC-AUDIT-GATE-INTEGRITY-001 iter-1 D1/D2 교훈 계승): 표 셀 안에서는 ERE(`-E`) alternation을 쓰지 않는다. 대신 (a) `&&` 연쇄 개별 `grep -c`, 또는 (b) BRE(플레인 grep) `\|` alternation(single-quote 안에서 shell-safe + table-safe + BRE-correct)만 사용한다.

## §A Given-When-Then 시나리오

### 시나리오 1 — completed SPEC이 in-place 개정 전이로 재개방된다 (R1)

- **Given**: `status: completed`인 SPEC이 사후 개정 필요 (예: 하위 호환성 깨진 API 명세 재수정)
- **When**: manager-spec이 `amendment_of: <self-ID>` frontmatter 필드 설정 + HISTORY `## Amendments` 행 추가 + `status: completed → in-progress` 전이를 수행하면
- **Then**: spec-frontmatter-schema.md § Status Transition Ownership Matrix에 신규 행이 명시되어 전이 소유자가 manager-spec임이 확인되고, `## Amendments` 섹션에 prior version + SHA + rationale + scope가 기록되며, `moai spec audit`는 해당 SPEC을 여전히 V3R6 modern era로 분류한다

### 시나리오 2 — 미충족 depends_on이 run-phase 진입을 차단한다 (R2)

- **Given**: SPEC-X의 `depends_on: [SPEC-Y-001]` 선언 + SPEC-Y-001의 `status: in-progress` (completed 아님)
- **When**: orchestrator가 `/moai run SPEC-X`를 Phase 0.5 Plan Audit Gate의 첫 sub-step에서 depends_on pre-flight를 실행하면
- **Then**: pre-flight는 plan-auditor 호출 전에 실패하고 orchestrator는 AskUserQuestion으로 3-option(wait / override `--ignore-deps` + log / abort) blocker를 방출하며, `--ignore-deps` 선택 시 `.moai/logs/depends-on-override.log`에 미충족 dep ID + rationale이 기록된다

### 시나리오 3 — Tier L SPEC 감사 시 plan-auditor가 design.md와 research.md를 입력으로 읽는다 (R3)

- **Given**: `tier: L` 선언된 SPEC (또는 tier 부재 → Tier L backward-compat)
- **When**: plan-auditor가 Input Contract를 평가하면
- **Then**: plan-auditor.md § Input Contract에 "Tier L: design.md + research.md are required inputs"와 "Tier-differentiated input contract" 리터럴 토큰이 명시되어, Tier L 감사에서 5-file 전체가 입력으로 소비됨이 계약적으로 보장된다

### 시나리오 4 — plan-artifact hash subject가 Go 구현과 정합된다 (R3c)

- **Given**: Phase 0.5 Plan Audit Gate가 skip-eligibility의 artifact-hash 검사를 수행
- **When**: orchestrator가 spec-workflow.md § Report Persistence의 hash subject list 서술을 참조하면
- **Then**: 4-file 집합 `{spec.md, plan.md, acceptance.md, tasks.md}`가 `internal/runtime/audit_cache.go` `planArtifactNames`와 verbatim 정합으로 명문화되어 있고, design.md/research.md는 `manual-skip judgment inputs`으로 별도 서술되며, amendment시 `cache-invalidating event`로 작동함이 서술되어 있다

## §D AC Matrix

모든 명령은 리포지토리 루트에서 실행. 경로 축약: `L-SFS` = `.claude/rules/moai/development/spec-frontmatter-schema.md`, `L-WF` = `.claude/rules/moai/workflow/spec-workflow.md`, `L-PA` = `.claude/agents/moai/plan-auditor.md`, `L-MS` = `.claude/agents/moai/manager-spec.md`, `T-*` = 대응 template mirror.

### R1 — delta-spec 수명주기

| AC | REQ | 검증 명령 | 기대 출력 |
|----|-----|-----------|-----------|
| AC-WFL-001 | WFL-001 | `grep -c 'amendment_of' L-SFS` | `≥ 1` (baseline 0 — Optional 필드 서술 추가) |
| AC-WFL-002 | WFL-001 | `grep -c 'completed → in-progress' L-SFS && grep -c 'amendment' L-SFS` | 순서대로 `≥ 1`, `≥ 2` (Status Enum 각주 + Ownership Matrix 행) |
| AC-WFL-003 | WFL-002 | `grep -c '## Amendments' L-MS` | `≥ 1` (HISTORY sub-section 작성 가이드) |
| AC-WFL-004 | WFL-002 | `grep -c 'prior completed' L-SFS \|\| grep -c 'prior_completed' L-SFS` (BRE alternation) | `≥ 1` (HISTORY 행 필드 서술) |
| AC-WFL-005 | WFL-003 | `grep -c 'in-place amendment' L-SFS` | `≥ 1` (commit subject 패턴 리터럴) |
| AC-WFL-006 | WFL-004 | `grep -c 'amendment in progress' L-WF \|\| grep -c 'amendment transition' L-WF` (BRE) | `≥ 1` (audit era 상호작용 서술) |

### R2 — depends_on run-phase 집행

| AC | REQ | 검증 명령 | 기대 출력 |
|----|-----|-----------|-----------|
| AC-WFL-007 | WFL-005 | `grep -c 'Depends_on Pre-flight' L-WF` | `≥ 1` (baseline 0 — sub-step 이름 리터럴) |
| AC-WFL-008 | WFL-005 | `grep -c 'depends_on' L-WF && grep -c 'Phase 0.5' L-WF` | 각각 `≥ 2` (Phase 0.5 확장 + 기존 언급) |
| AC-WFL-009 | WFL-006 | `grep -c 'fulfillment' L-WF` | `≥ 1` (baseline 0 — fulfillment 정의 토큰) |
| AC-WFL-010 | WFL-006 | `grep -c 'status: completed' L-WF` | `≥ 1` (충족 유일 상태 명시) |
| AC-WFL-011 | WFL-007 | `grep -c 'ignore-deps' L-WF` | `≥ 1` (override flag 리터럴) |
| AC-WFL-012 | WFL-007 | `grep -c 'depends-on-override.log' L-WF` | `≥ 1` (override log path 리터럴) |

### R3 — Tier L 산물 + plan-auditor 입력 계약 + hash 정합

| AC | REQ | 검증 명령 | 기대 출력 |
|----|-----|-----------|-----------|
| AC-WFL-013 | WFL-008 | `grep -c 'design\.md' L-SFS && grep -c 'research\.md' L-SFS` | 각각 `≥ 1` (Optional `tier:` 필드에 5-file 열거) |
| AC-WFL-014 | WFL-008 | `grep -c 'Tier L' L-SFS` | `≥ 2` (Status Enum이 아닌 Optional Fields 절 내 Tier L 5-file 서술 + 기존 다른 언급) |
| AC-WFL-015 | WFL-009 | `grep -c 'Tier-differentiated input contract' L-PA` | `≥ 1` (baseline 0 — 리터럴 토큰) |
| AC-WFL-016 | WFL-009 | `grep -c 'Tier L: design.md + research.md are required inputs' L-PA` | `≥ 1` (baseline 0 — 리터럴 토큰) |
| AC-WFL-017 | WFL-009 | `grep -c 'design\.md' L-PA && grep -c 'research\.md' L-PA` | 각각 `≥ 1` (baseline 0 — Input Contract 신규 언급) |
| AC-WFL-018 | WFL-010 | `grep -c 'planArtifactNames' L-WF` | `≥ 1` (baseline 0 — Go 구현 이름 인용) |
| AC-WFL-019 | WFL-010 | `grep -c 'tasks\.md' L-WF` | `≥ 1` (4-file 집합 열거에 tasks.md 포함) |
| AC-WFL-020 | WFL-010 | `grep -c 'manual-skip judgment inputs' L-WF` | `≥ 1` (baseline 0 — design/research 서술 토큰) |
| AC-WFL-021 | WFL-010 | `grep -c 'V3R4' L-WF` | `≥ 1` (tasks.md 잔재 서술) |
| AC-WFL-022 | WFL-011 | `grep -c 'cache-invalidating event' L-WF` | `≥ 1` (baseline 0 — amendment시 hash 변경 서술 토큰) |

### Cross-cutting — Template Mirror 동기화

| AC | REQ | 검증 명령 | 기대 출력 |
|----|-----|-----------|-----------|
| AC-WFL-023 | WFL-012 | `grep -c 'Tier-differentiated input contract' T-PA && grep -c 'design\.md' T-PA && grep -c 'research\.md' T-PA` | 각각 `≥ 1` (mirror plan-auditor Tier-differentiated) |
| AC-WFL-024 | WFL-012 | `grep -c 'amendment_of' T-SFS && grep -c 'in-place amendment' T-SFS` | 각각 `≥ 1` (mirror spec-frontmatter-schema) |
| AC-WFL-025 | WFL-012 | `grep -c 'Depends_on Pre-flight' T-WF && grep -c 'ignore-deps' T-WF && grep -c 'manual-skip judgment inputs' T-WF && grep -c 'cache-invalidating event' T-WF` | 각각 `≥ 1` (mirror spec-workflow 다중-토큰) |
| AC-WFL-026 | WFL-012 | `grep -c 'amendment_of' T-MS && grep -c '## Amendments' T-MS` | 각각 `≥ 1` (mirror manager-spec) |
| AC-WFL-027 | WFL-012 | `make build; echo "exit=$?"` | `exit=0` |
| AC-WFL-028 | WFL-012 | `grep -rc 'SPEC-WORKFLOW-LIFECYCLE\|REQ-WFL' internal/template/templates/.claude/agents/moai/ internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md > /tmp/wfl-neutrality.txt; awk -F: '{s+=$2} END {print s+0}' /tmp/wfl-neutrality.txt` | `0` (Neutrality — `s+0`로 zero-match 빈 출력 방지) |
| AC-WFL-029 | 전체 | `command -v moai > /dev/null; echo "tool=$?"; moai spec lint > /tmp/wfl-lint.log 2>&1; grep 'SPEC-WORKFLOW-LIFECYCLE-001' /tmp/wfl-lint.log > /tmp/wfl-lint-self.txt; grep -c 'ERROR' /tmp/wfl-lint-self.txt` | `tool=0` AND 최종 grep `0` (ERROR-급만 판정 — WARNING은 비계수; 선행 SPEC 패턴 계승) |
| AC-WFL-030 | WFL-012 | `git diff --name-only HEAD -- .moai/config/astgrep-rules/security/credentials.yml \|\| echo "no-change"; git diff --cached --name-only -- .moai/config/astgrep-rules/security/credentials.yml \|\| echo "no-staged-change"` | 둘 다 "no-change" / "no-staged-change" (`credentials.yml` 무접촉 서약 검증) |

## §D.1 Edge Cases

- **E1 (병렬 세션 선점)**: pre-flight baseline grep이 이미 기대값이면 해당 REQ는 verify-then-SKIP — 중복 편집 금지, blocker report로 보고
- **E2 (amendment_of 해석)**: self-referential 값(SPEC-X-001이 `amendment_of: SPEC-X-001`)과 parent-ID 값(SPEC-X-002가 `amendment_of: SPEC-X-001`)이 모두 유효하다 — 전자는 in-place, 후자는 successor. AC는 둘 중 하나의 패턴을 서술에서 발견하면 PASS
- **E3 (depends_on 사이클)**: A가 B에게 depend하고 B가 A에게 depend하는 순환은 본 SPEC이 다루지 않는다 (Out of Scope). pre-flight는 각 dep을 개별 조회할 뿐 순회하지 않는다
- **E4 (Tier L에 tasks.md 존재?)**: 일부 V3R4-era grandfathered SPEC은 tasks.md를 가질 수 있다 (planArtifactNames에 포함된 이유). V3R6 Tier L 신규 SPEC은 tasks.md를 갖지 않는다 — hash는 존재하지 않는 파일을 무시한다 (Go 구현이 이미 그렇게 동작)
- **E5 (make build 실패)**: template 편집 후 build 실패 시 mirror 편집을 revert하지 말고 원인(YAML frontmatter 손상 등) 진단 후 수정 — 커밋 전 build green 필수
- **E6 (StatusGitConsistency WARNING — lifecycle-예정 상태, 명시 예외)**: 본 SPEC의 plan-phase 수정 커밋 이력으로 `moai spec lint`가 본 SPEC에 `StatusGitConsistency WARNING`('draft' vs git-implied)을 방출하는 것은 run-phase status 전이(`draft → in-progress → …`)가 진행되기 전까지 예상되는 정상 상태다. AC-WFL-029는 ERROR-급만 계수하므로 이 WARNING은 판정에 미포함
- **E7 (credentials.yml이 git untracked로 이미 존재)**: 본 SPEC은 credentials.yml을 전혀 편집하지 않는다. AC-WFL-030은 "본 SPEC의 커밋이 credentials.yml을 건드리지 않았음"을 검증 (untracked 상태는 무관)

## §D.2 Quality Gate

- 전 AC 판정은 단일 턴 병렬 검증 배치로 실행하고 verbatim 출력을 file-redirect contract에 따라 보존
- plan-auditor 독립 감사 PASS ≥ 0.85 (Tier L threshold — frontmatter `tier: L` 명시로 확정)
- 커밋은 pathspec 한정 — 무관 파일 무포함을 `git show --stat`로 확인
- `credentials.yml` 무접촉 — `git diff --name-only HEAD..origin/main` 또는 `git show --stat`에서 해당 경로가 절대 나타나지 않아야 함

## §D.3 Definition of Done

1. AC-WFL-001..030 전 행 PASS (verbatim 증거 인용; 매트릭스 30행)
2. live 4파일 + mirror 4파일 편집 완료, `make build` exit 0
3. Neutrality grep 0 (AC-WFL-028)
4. credentials.yml 무접촉 (AC-WFL-030)
5. milestone별 pathspec 커밋 완료, 무관 변경 무포함
6. progress.md §E.2/§E.3에 run-phase 증거 기록 (manager-develop 소관)
