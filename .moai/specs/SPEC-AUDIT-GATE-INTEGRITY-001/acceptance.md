# SPEC-AUDIT-GATE-INTEGRITY-001 — Acceptance Criteria

> 모든 AC는 기계 검증 가능 (명령 + 기대 출력). `verification-claim-integrity.md` §3.2 — Evidence는 verbatim 출력이며 요약 불가.
>
> 명령 표기 규약 (iter-1 D1/D2 교훈): 표 셀 안에서는 ERE(`-E`) alternation을 쓰지 않는다 — ERE에서 `\|`는 리터럴 파이프(영구 false-fail), 비이스케이프 `|`는 markdown 표를 깨뜨린다. 대신 (a) `&&` 연쇄 개별 `grep -c`, 또는 (b) BRE(플레인 grep) `\|` alternation(single-quote 안에서 shell-safe + table-safe + BRE-correct)만 사용한다. 파이프라인이 필요한 AC는 임시 파일 2-스텝으로 분해한다.

## §A Given-When-Then 시나리오

### 시나리오 1 — D7 BLOCKING이 Verdict를 강제한다 (R1)

- **Given**: superseded 상태의 SPEC을 reconciliation 절 없이 참조하는 SPEC 아티팩트
- **When**: plan-auditor가 plan-phase 감사를 수행하면
- **Then**: 보고서 `## Must-Pass Results`에 `MP-5 ... FAIL` 행이 기록되고, aggregate score가 threshold 이상이어도 `Verdict: FAIL`이며, 해당 finding이 `## Defects Found`에 severity=critical로 나타난다

### 시나리오 2 — hierarchical 모드 보고서가 worked example 형식을 따른다 (R2)

- **Given**: `harness.yaml`에 `evaluator_mode: hierarchical` 설정
- **When**: sync-auditor가 구현을 평가하면
- **Then**: 보고서가 `### Hierarchical-Mode Output Example`에 제시된 형식(sub-criterion 행, canonical anchor 0.25/0.50/0.75/1.00, per-dimension aggregation, must-pass verdict)을 따르고, 각 차원 Evidence 셀에 프로젝트 언어 자동감지로 선택된 기계 검증 명령의 verbatim 출력이 인용된다

### 시나리오 3 — manager-spec이 실행된 Bash 출력으로 SPEC ID를 검증한다 (R4)

- **Given**: canonical regex 검증 대상인 신규 SPEC ID 후보 (임의의 후보 문자열)
- **When**: manager-spec이 spec.md `Write`를 준비하면
- **Then**: 응답 본문에 실행된 Bash one-liner와 그 verbatim 출력(`PASS` 또는 `FAIL`)이 인용되고, `FAIL`이면 Write가 중단되고 blocker report가 반환된다

### 시나리오 4 — 두 보고서 스트림이 상호 참조된다 (R3)

- **Given**: Phase 0.5 게이트가 skip-eligibility를 평가하는 상황
- **When**: 오케스트레이터가 spec-workflow.md § Report Persistence를 참조하면
- **Then**: "plan-phase review stream"의 final-iteration verdict가 run-gate 입력이고, artifact-hash 검사는 plan-artifact hash(plan artifacts SHA-256) 재계산·대조이며, "run-gate stream" date-file은 verdict 기록 표면임이 명시되어 있고, plan-auditor.md에서도 동일한 두 스트림 구분과 상호 참조를 발견한다

## §D AC Matrix

모든 명령은 리포지토리 루트에서 실행. 경로 축약: `L-PA` = `.claude/agents/moai/plan-auditor.md`, `L-SA` = `.claude/agents/moai/sync-auditor.md`, `L-MS` = `.claude/agents/moai/manager-spec.md`, `L-WF` = `.claude/rules/moai/workflow/spec-workflow.md`, `T-*` = 대응 template mirror (`internal/template/templates/` 하위 동일 경로).

| AC | REQ | 검증 명령 | 기대 출력 |
|----|-----|-----------|-----------|
| AC-AGI-001 | AGI-001 | `grep -c '(MP-5)\|(MP-6)' L-PA` (BRE alternation) | `≥ 2` (baseline 0; iter-1 D1 수정 — ERE `\|` 금지) |
| AC-AGI-002a | AGI-002 | `grep -c 'must-pass-equivalent' L-PA` | `≥ 1` (baseline 0) |
| AC-AGI-002b | AGI-002 | `grep -c 'severity=critical' L-PA` | `≥ 1` (baseline 0 실측 — fold 절 토큰, D8) |
| AC-AGI-003 | AGI-003 | `grep -c 'MP-5' L-PA && grep -c 'MP-6' L-PA` | 각각 `≥ 2` (M5 정의 + Must-Pass Results 표 행; baseline 각 0) |
| AC-AGI-004a | AGI-004 | `grep -c 'evaluator_mode: hierarchical' L-SA` | `≥ 2` (baseline 1 — 기존 HRN-003 절 + 신규 선택 규칙) |
| AC-AGI-004b | AGI-004 | `grep -c 'sub-criteria refinement' L-SA` | `≥ 1` (baseline 0 실측 — 두 모델 관계 절 토큰, D8) |
| AC-AGI-005 | AGI-005 | `grep -c '### Hierarchical-Mode Output Example' L-SA` | `= 1` (baseline 0) |
| AC-AGI-006a | AGI-006 | `grep -c 'project-language auto-detection' L-SA` | `≥ 1` (baseline 0 실측 — 언어 자동감지 규약 토큰) |
| AC-AGI-006b | AGI-006 | `grep -c 'go test' L-SA && grep -c 'pytest' L-SA && grep -c 'npm test' L-SA && grep -c 'cargo' L-SA` | 각각 `≥ 1` (baseline 각 0 실측 — 4개 언어 동등 예시, Go는 그중 하나) |
| AC-AGI-006c | AGI-006 | `grep -c 'verbatim' L-SA` | `≥ 1` (baseline 0 — Evidence verbatim 인용 의무 문구) |
| AC-AGI-007 | AGI-007 | `grep -c 'moai-ref-owasp-checklist' L-SA && grep -c 'moai-ref-testing-pyramid' L-SA` | 각각 `≥ 1` (frontmatter skills; baseline 각 0; D9 — 정확 카운트 `= 1` 완화) |
| AC-AGI-008a | AGI-008 | `grep -c 'plan-phase review stream' L-WF && grep -c 'run-gate stream' L-WF` | 각각 `≥ 1` (baseline 각 0) |
| AC-AGI-008b | AGI-008 | `grep -c 'plan-phase review stream' L-PA && grep -c 'run-gate stream' L-PA` | 각각 `≥ 1` (baseline 각 0) |
| AC-AGI-009a | AGI-009 | `grep -c 'final-iteration verdict' L-WF` | `≥ 1` (baseline 0) |
| AC-AGI-009b | AGI-009 | `grep -c 'plan-artifact hash' L-WF` | `≥ 1` (baseline 0 실측 — D4 hash 메커니즘 토큰) |
| AC-AGI-010a | AGI-010 | `grep -cE '=~ \^SPEC' L-MS` | `≥ 1` (baseline 0 — 실행형 Bash one-liner 존재; 파이프 없는 단일 패턴이라 `-E` 안전) |
| AC-AGI-010b | AGI-010 | `grep -c 'mentally' L-MS` | `= 0` (baseline 1 — iter-1 감사 실측: self-check 섹션 단독 출현 확인) |
| AC-AGI-011 | AGI-011 | `grep -c 'L32 chain context' L-MS` | `= 0` (baseline 1 — 역사 서사 압축) |
| AC-AGI-012a | AGI-012 | `grep -c '(MP-5)\|(MP-6)' T-PA && grep -c 'must-pass-equivalent' T-PA && grep -c 'plan-phase review stream' T-PA && grep -c 'run-gate stream' T-PA` | 순서대로 `≥ 2`, `≥ 1`, `≥ 1`, `≥ 1` (mirror 다중-토큰, D7) |
| AC-AGI-012b | AGI-012 | `grep -c '### Hierarchical-Mode Output Example' T-SA && grep -c 'project-language auto-detection' T-SA && grep -c 'verbatim' T-SA` | 순서대로 `= 1`, `≥ 1`, `≥ 1` (mirror 다중-토큰, D7) |
| AC-AGI-012c | AGI-012 | `grep -cE '=~ \^SPEC' T-MS && grep -c 'mentally' T-MS` | 순서대로 `≥ 1`, `= 0` (mirror `mentally` baseline 1 실측 — 제거 검증, D7) |
| AC-AGI-012d | AGI-012 | `grep -c 'plan-phase review stream' T-WF && grep -c 'run-gate stream' T-WF && grep -c 'final-iteration verdict' T-WF && grep -c 'plan-artifact hash' T-WF` | 각각 `≥ 1` (mirror 다중-토큰, D7) |
| AC-AGI-012e | AGI-012 | `grep -rc 'SPEC-AUDIT-GATE-INTEGRITY\|REQ-AGI' internal/template/templates/.claude/agents/moai/ internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md > /tmp/agi-neutrality.txt; awk -F: '{s+=$2} END {print s+0}' /tmp/agi-neutrality.txt` | `0` (Neutrality; D9 — `s+0`로 zero-match 빈 출력 방지) |
| AC-AGI-013 | AGI-012 | `make build; echo "exit=$?"` | `exit=0` |
| AC-AGI-014 | 전체 | `command -v moai > /dev/null; echo "tool=$?"; moai spec lint > /tmp/agi-lint.log 2>&1; grep 'SPEC-AUDIT-GATE-INTEGRITY-001' /tmp/agi-lint.log > /tmp/agi-lint-self.txt; grep -c 'ERROR' /tmp/agi-lint-self.txt` | `tool=0` AND 최종 grep `0` (iter-2 N1 교정 — SPEC-범위 ERROR-급만 판정: `moai spec lint` exit code는 리포 전역이라 타 SPEC pre-existing ERROR로 exit=1이 정상이며 본 AC의 판정 입력이 아님; 본 SPEC의 WARNING은 비계수 — E6 참조) |

## §D.1 Edge Cases

- **E1 (병렬 세션 선점)**: pre-flight baseline grep이 이미 기대값이면 해당 REQ는 verify-then-SKIP — 중복 편집 금지, blocker report로 보고
- **E2 (mentally 토큰)**: iter-1 감사 실측으로 live/mirror 각 1건(self-check 섹션 단독) 확인됨. 편집 후 신규 출현이 발견되면 `grep -n 'mentally'` 출력을 인용하고 scope 판단
- **E3 (Must-Pass 표 서식 변형)**: plan-auditor 보고서 템플릿의 MP-5/6 행이 `[PASS/FAIL/N/A]` 3-상태를 갖는 경우 — D7/D8 검증 verb가 실행 불능(대상 파일 부재 등)일 때 N/A 허용 여부는 MP-4 선례(N/A auto-pass)를 따라 명문화
- **E4 (make build 실패)**: template 편집 후 build 실패 시 mirror 편집을 revert하지 말고 원인(YAML frontmatter 손상 등) 진단 후 수정 — 커밋 전 build green 필수
- **E5 (언어 편향 회귀)**: sync-auditor 차원-명령 표에서 특정 언어가 "PRIMARY"/"기본"으로 승격되거나 4개 언어 예시 중 일부가 누락되면 MP-4 회귀 — AC-AGI-006b의 4-토큰 동등 확인이 회귀 가드
- **E6 (StatusGitConsistency WARNING — lifecycle-예정 상태, 명시 예외)**: 본 SPEC의 plan-phase 수정 커밋 이력으로 `moai spec lint`가 본 SPEC에 `StatusGitConsistency WARNING`('draft' vs git-implied)을 방출하는 것은 run-phase status 전이(`draft → in-progress → …`)가 진행되기 전까지 예상되는 정상 상태다 (iter-2 N1 실측: WARNING 1건). AC-AGI-014는 ERROR-급만 계수하므로 이 WARNING은 판정에 미포함 — 본 SPEC 명명 finding이 ERROR로 나타나는 경우에만 FAIL

## §D.2 Quality Gate

- 전 AC 판정은 단일 턴 병렬 검증 배치로 실행하고 verbatim 출력을 file-redirect contract에 따라 보존
- plan-auditor 독립 감사 PASS ≥ 0.80 (Tier M threshold — frontmatter `tier: M` 명시로 확정)
- 커밋은 pathspec 한정 — 무관 파일 무포함을 `git show --stat`로 확인

## §D.3 Definition of Done

1. AC-AGI-001..014 전 행 PASS (verbatim 증거 인용; 매트릭스 25행)
2. live 4파일 + mirror 4파일 편집 완료, `make build` exit 0
3. Neutrality grep 0 (AC-AGI-012e)
4. milestone별 pathspec 커밋 완료, 무관 변경 무포함
5. progress.md §E.2/§E.3에 run-phase 증거 기록 (manager-develop 소관)
