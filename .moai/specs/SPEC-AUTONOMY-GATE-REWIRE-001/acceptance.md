# acceptance.md — SPEC-AUTONOMY-GATE-REWIRE-001

모든 AC 는 Given-When-Then 이며 명령과 기대 출력으로 판정한다. `BASE` 는 `plan.md §C` 가 정의한 SHA(M0 에서 `progress.md §E.2` 에 기록)다 — 브랜치 이름을 쓰지 않는다. 명령은 워크트리 루트에서 실행한다.

## §A. 공용 정의

```bash
BASE=<progress.md §E.2 에 기록된 SHA>
T=internal/template/templates

# 기존 파일 편집 집합 (design.md §2 의 2~14행; NC-2 가 (b) 면 clarity-interview.md 추가)
EDIT_SET="CLAUDE.md
.claude/rules/moai/core/askuser-protocol.md
.claude/rules/moai/workflow/goal-directive.md
.claude/rules/moai/workflow/orchestration-mode-selection.md
.claude/skills/moai/SKILL.md
.claude/skills/moai/workflows/moai.md
.claude/skills/moai/workflows/plan.md
.claude/skills/moai/workflows/plan/spec-assembly.md
.claude/skills/moai/workflows/run.md
.claude/skills/moai/workflows/goal.md
.claude/skills/moai/workflows/sync.md
.claude/skills/moai/workflows/sync/doc-execution.md
.claude/skills/moai/workflows/sync/delivery.md"

# G1 발화 지점 (research.md §1.2 의 E, 10개)
E_SET="CLAUDE.md
.claude/rules/moai/workflow/orchestration-mode-selection.md
.claude/rules/moai/core/askuser-protocol.md
.claude/rules/moai/workflow/goal-directive.md
.claude/skills/moai/SKILL.md
.claude/skills/moai/workflows/moai.md
.claude/skills/moai/workflows/plan.md
.claude/skills/moai/workflows/plan/spec-assembly.md
.claude/skills/moai/workflows/run.md
.claude/skills/moai/workflows/goal.md"

strip()   { awk '/^<!-- moai:contract-mode-start id="[a-z0-9-]+" -->$/{s=1} !s{print} /^<!-- moai:contract-mode-end -->$/{s=0}' "$1"; }
extract() { awk '/^<!-- moai:contract-mode-start id="[a-z0-9-]+" -->$/{s=1}  s{print} /^<!-- moai:contract-mode-end -->$/{s=0}' "$1"; }
SSOT=.claude/rules/moai/workflow/contract-autonomy.md
```

## §B. 수용 기준

### AC-GR-001 — guided 경로 바이트 보존 (REQ-GR-001·002·003)

- **Given** `EDIT_SET` 의 각 파일과 그 템플릿 사본(26개)
- **When** contract-mode 블록을 `strip` 으로 걷어낸다
- **Then** 결과가 `git show "$BASE:<path>"` 와 바이트 동일하다

```bash
for f in $EDIT_SET; do for p in "$f" "$T/$f"; do strip "$p" | cmp -s - <(git show "$BASE:$p") || echo "DRIFT $p"; done; done
```

기대 출력: 없음(빈 출력). 추가로 `for f in $EDIT_SET; do extract "$f" | grep -c 'moai:contract-mode-start'; done` 의 합이 ≥ 1 이어야 한다 — 블록이 하나도 없으면 이 AC 는 공허하게 통과하므로 FAIL 로 본다.

분류: **불변 가드**(현재 트리에서도 녹색). 판정식 자체의 반증 능력은 변이 탐침으로 확인했다(`§C` EV-5).

### AC-GR-002 — 로컬↔템플릿 블록 동등 (REQ-GR-070)

- **Given** `EDIT_SET` 의 각 파일 쌍과 신규 SSOT
- **When** 두 사본에서 블록을 `extract` 하고, SSOT 는 통째로 비교한다
- **Then** 모든 쌍이 동일하다

```bash
for f in $EDIT_SET; do diff -q <(extract "$f") <(extract "$T/$f") >/dev/null || echo "BLOCK-DIFF $f"; done; cmp -s "$SSOT" "$T/$SSOT" || echo "SSOT-DIFF"
```

기대 출력: 없음.

### AC-GR-003 — 승계 분기 불변 (REQ-GR-070)

- **Given** 기준 ref 에서 이미 갈라져 있던 로컬·템플릿 쌍
- **When** 블록을 걷어낸 두 사본의 diff 와 기준 ref 두 사본의 diff 를 비교한다
- **Then** 두 diff 가 같다

```bash
for f in $EDIT_SET; do cmp -s <(diff <(git show "$BASE:$f") <(git show "$BASE:$T/$f")) <(diff <(strip "$f") <(strip "$T/$f")) || echo "INHERITED-DRIFT $f"; done
```

기대 출력: 없음.

### AC-GR-004 — Frozen·제외 파일과 헌법 레지스트리 불변 (REQ-GR-050·090)

- **Given** REQ-GR-090 목록과 그 템플릿 사본
- **When** `BASE` 와 비교하고 헌법 검증을 돌린다
- **Then** diff 가 없고 헌법 검증이 기준선과 같은 OK 를 낸다

```bash
git diff --quiet "$BASE" -- .claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/zone-registry.md .claude/agents .claude/output-styles .claude/rules/moai/workflow/ci-autofix-protocol.md .claude/rules/moai/workflow/context-window-management.md .claude/rules/moai/core/agent-common-protocol.md "$T/.claude/rules/moai/core/moai-constitution.md" "$T/.claude/rules/moai/core/zone-registry.md" "$T/.claude/agents" "$T/.claude/output-styles" "$T/.claude/rules/moai/workflow/ci-autofix-protocol.md" "$T/.claude/rules/moai/workflow/context-window-management.md" "$T/.claude/rules/moai/core/agent-common-protocol.md"
moai constitution validate
```

기대: 첫 명령 exit 0. 둘째 명령 exit 0, 출력 첫 줄이 `constitution validate: OK — no drift or violations detected` 로 시작하고 검사 항목 수가 `BASE` 에서 잰 값과 같다(ca1d5dc43 에서는 `97 of 101`).

### AC-GR-005 — G1 발화 지점 전수 전환 (REQ-GR-011)

- **Given** `E_SET` 10개 파일과 템플릿 사본
- **When** 각 파일의 블록 수를 센다
- **Then** 20개 사본 모두 1 이상이다

```bash
for f in $E_SET; do for p in "$f" "$T/$f"; do n=$(grep -c '^<!-- moai:contract-mode-start id=' "$p"); [ "$n" -ge 1 ] || echo "NO-BLOCK $p"; done; done
```

기대 출력: 없음.

보조 판정 — 분류의 빈틈 없음: `BASE` 에서 `grep -rlE "Kickoff" CLAUDE.md .claude/rules .claude/skills/moai .claude/output-styles .claude/agents` 가 낸 파일 집합이 `research.md §1.2` 의 E ∪ R ∪ H ∪ 로컬 전용 목록(M0 재분류 반영)과 같다. 다르면 FAIL — 새 Kickoff 위치가 분류되지 않은 것이다.

### AC-GR-006 — SSOT 의 등가 조항과 게이트 처분표 (REQ-GR-011·020·021·022·030·040·051·052)

- **Given** 신규 SSOT
- **When** 필수 절 제목과 게이트 이름을 찾는다
- **Then** 모두 있다

```bash
for s in "Implementation Kickoff Approval" "Socratic interview" "approach approval" "assumption" "plan-audit" "SPEC quality gate" "Documentation Scope" "question-channel monopoly" "sync-auditor" "Report-Before-Ask" "escalate_on"; do grep -qF "$s" "$SSOT" || echo "MISSING: $s"; done; grep -cE '^## ' "$SSOT"
```

기대: `MISSING` 줄 없음, 제목 수 ≥ 8 (`design.md §3` 의 8개 절). **[A1 감사 통과본으로 재확인]** (`escalate_on` 등 필드명)

분류: 존재 검사다. 내용의 옳음(예: 에스컬레이션이 질문이 아님)은 plan-auditor·sync-auditor 검토가 판정한다.

### AC-GR-007 — 생명주기 순서 (REQ-GR-060·061·062)

- **Given** run.md·sync.md 의 블록과 SSOT
- **When** 단계명을 등장 순서대로 뽑는다
- **Then** run 은 `Discovery RED GREEN Qualification`, sync 는 `Closure Integration Push`, SSOT 는 일곱 단계 전부가 이 순서다

```bash
extract .claude/skills/moai/workflows/run.md | grep -oE '\b(Discovery|RED|GREEN|Qualification)\b' | awk '!seen[$0]++' | paste -sd' ' -
extract .claude/skills/moai/workflows/sync.md | grep -oE '\b(Closure|Integration|Push)\b' | awk '!seen[$0]++' | paste -sd' ' -
grep -oE '\b(Discovery|RED|GREEN|Qualification|Closure|Integration|Push)\b' "$SSOT" | awk '!seen[$0]++' | paste -sd' ' -
```

기대 출력(세 줄, 템플릿 사본도 동일):

```text
Discovery RED GREEN Qualification
Closure Integration Push
Discovery RED GREEN Qualification Closure Integration Push
```

### AC-GR-008 — 상시 가드 테스트 (REQ-GR-001·072·073)

- **Given** `internal/template/contract_mode_blocks_test.go`
- **When** 불량 픽스처(짝 불일치 / evolvable 구간 안의 블록 / SPEC ID 를 담은 블록 / 검사 대상 블록 0개)로 먼저 실행하고, 이어 실제 템플릿 트리로 실행한다
- **Then** 불량 픽스처 네 경우 모두 실패가 관측되고, 실제 트리에서는 통과한다

```bash
go test ./internal/template/ -run 'TestContractModeBlocks' -count=1 -v
```

기대: 최종 실행에서 `--- PASS: TestContractModeBlocks` 와 하위 테스트 PASS, 그리고 `[no tests to run]` 가 **없어야** 한다(빈 선택은 FAIL). RED 원문(불량 픽스처에서의 `--- FAIL`)은 `progress.md §E.2` 에 원문으로 남긴다.

### AC-GR-009 — 템플릿 중립성 (REQ-GR-072)

- **Given** 템플릿 사본의 모든 블록과 템플릿 SSOT
- **When** 금지 클래스 패턴을 찾고, 기존 중립성 테스트를 돌린다
- **Then** 적중 0, 테스트 통과

```bash
{ for f in $EDIT_SET; do extract "$T/$f"; done; cat "$T/$SSOT"; } | grep -nE 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|\bREQ-[A-Z]|\bAC-[A-Z]{2,}-[0-9]|\bt[0-9]{3,4}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\bG(1|2|3|4|5|7|8|11|14|20)\b'
go test ./internal/template/ -run 'TestTemplateNeutrality|TestInternalContentLeak' -count=1
```

기대: 첫 명령 출력 없음(exit 1). 둘째 명령 `ok`. 둘째 명령이 `[no tests to run]` 를 내면 실제 테스트 이름을 `go test -list` 로 확인해 교체한다 — 빈 선택은 통과가 아니다.

### AC-GR-010 — always-loaded 예산 (REQ-GR-080)

- **Given** always-loaded 편집 파일 3개(`CLAUDE.md`, `askuser-protocol.md`, `goal-directive.md`)
- **When** `BASE` 대비 문자 수 증가를 잰다
- **Then** 합계 증가 ≤ `design.md §4` 상한(NC-6 결정값, 기본 1,500)이고 각 파일 < 40,000

```bash
for f in CLAUDE.md .claude/rules/moai/core/askuser-protocol.md .claude/rules/moai/workflow/goal-directive.md; do a=$(git show "$BASE:$f" | LC_ALL=en_US.UTF-8 wc -m); b=$(LC_ALL=en_US.UTF-8 wc -m < "$f"); echo "$f $a $b $((b-a))"; done
```

기대: 네 번째 열 합 ≤ 상한, 세 번째 열 모두 < 40000.

### AC-GR-011 — 블록 첫 문장이 적용 조건 (REQ-GR-003)

- **Given** 모든 블록(로컬·템플릿)
- **When** 시작 마커 다음의 첫 비어 있지 않은 줄을 본다
- **Then** 그 줄이 `` Where `workflow.autonomy.mode: contract` `` 로 시작한다

```bash
for f in $EDIT_SET; do for p in "$f" "$T/$f"; do awk '/^<!-- moai:contract-mode-start/{w=1; next} w&&NF{ if ($0 !~ /^Where `workflow\.autonomy\.mode: contract`/) print FILENAME": "$0; w=0 }' "$p"; done; done
```

기대 출력: 없음. **[A1 감사 통과본으로 재확인]** (설정 키 경로)

### AC-GR-012 — 로컬 사본의 evolvable 구간 비중첩 (REQ-GR-001)

- **Given** 로컬 편집 파일(가드 테스트는 템플릿만 본다)
- **When** evolvable 구간 안에 시작 마커가 있는지 본다
- **Then** 없다

```bash
for f in $EDIT_SET; do awk '/moai:evolvable-start/{e=1} /moai:evolvable-end/{e=0} e&&/moai:contract-mode-start/{print FILENAME": "NR}' "$f"; done
```

기대 출력: 없음.

### AC-GR-013 — plan-audit FAIL 자동 수리 상한 (REQ-GR-030·031)

- **Given** `spec-assembly.md` 의 `contract-audit-retry`·`contract-quality-gate` 블록
- **When** 블록 내용을 본다
- **Then** 두 블록 모두 `audit_retries` 와 `budget_default` 를 언급하고, 도구 호출 지시어 `AskUserQuestion` 을 담지 않는다

```bash
awk '/id="contract-(audit-retry|quality-gate)"/{s=1} s{print} /moai:contract-mode-end/{s=0}' .claude/skills/moai/workflows/plan/spec-assembly.md | grep -cE 'audit_retries|budget_default'
awk '/id="contract-(audit-retry|quality-gate)"/{s=1} s{print} /moai:contract-mode-end/{s=0}' .claude/skills/moai/workflows/plan/spec-assembly.md | grep -c 'AskUserQuestion'
```

기대: 첫 명령 ≥ 2, 둘째 명령 `0`. **[A1 감사 통과본으로 재확인]** (`budget.audit_retries`, `workflow.autonomy.escalation.budget_default`)

### AC-GR-014 — sync 확인 질문 제거 (REQ-GR-040·041)

- **Given** `sync.md`·`sync/doc-execution.md`·`sync/delivery.md` 의 블록
- **When** 블록 내용을 본다
- **Then** 블록들이 `gate-sync-2` 를 이름으로 가리키고, 「에스컬레이션」 경로를 명시하며, `AskUserQuestion` 을 담지 않는다

```bash
for f in .claude/skills/moai/workflows/sync.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/workflows/sync/delivery.md; do extract "$f"; done | grep -cE 'gate-sync-2|escalat'
for f in .claude/skills/moai/workflows/sync.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/workflows/sync/delivery.md; do extract "$f"; done | grep -c 'AskUserQuestion'
```

기대: 첫 명령 ≥ 3, 둘째 `0`.

### AC-GR-015 — 서명 게이트 문언 (REQ-GR-010·013·014·015)

- **Given** `run.md` 의 `contract-signing-run` 블록과 `spec-assembly.md` 의 `contract-signing-review`·`contract-draft` 블록
- **When** 명령·상태 이름을 찾는다
- **Then** run 블록에 `moai contract verify`, 서명 검토 블록에 `moai contract sign`, 초안 블록에 `contract.yaml` 이 있다

```bash
extract .claude/skills/moai/workflows/run.md | grep -c 'moai contract verify'
awk '/id="contract-signing-review"/{s=1} s{print} /moai:contract-mode-end/{s=0}' .claude/skills/moai/workflows/plan/spec-assembly.md | grep -c 'moai contract sign'
awk '/id="contract-draft"/{s=1} s{print} /moai:contract-mode-end/{s=0}' .claude/skills/moai/workflows/plan/spec-assembly.md | grep -c 'contract.yaml'
```

기대: 세 명령 모두 ≥ 1. **[A1 감사 통과본으로 재확인]** (CLI 동사, 파일명)

### AC-GR-016 — 빌드와 범위 테스트 (전 REQ 공통)

- **Given** 편집이 끝난 트리
- **When** 빌드와 템플릿 패키지 테스트를 돌린다
- **Then** 둘 다 성공

```bash
make build
go test ./internal/template/... -count=1
```

기대: `make build` exit 0(선행 `agents-emit-check`·`commands-emit-check` 포함), 테스트 `ok`. 전체 스위트는 CI 가 판정한다.

## §C. 증거 원장 — RED-now 셀 (트리 `ca1d5dc43`)

M0 에서 `BASE` 가 정해지면 같은 명령을 `BASE` 에서 다시 재고 이 표를 `progress.md §E.2` 에 갱신한다.

| id | 대상 AC | 명령 (단일 호출) | stdout | exit | 빨간 이유 |
|---|---|---|---|---|---|
| EV-1 | AC-GR-005, AC-GR-011 | `grep -rlc 'moai:contract-mode-start' CLAUDE.md .claude internal/template/templates` | (빈 출력) | 1 | 블록이 아직 없다 — 이 작업이 만든다 |
| EV-2 | AC-GR-002, AC-GR-006 | `ls .claude/rules/moai/workflow/contract-autonomy.md internal/template/templates/.claude/rules/moai/workflow/contract-autonomy.md internal/template/contract_mode_blocks_test.go` | `ls: … No such file or directory` ×3 (stderr) | 1 | SSOT 와 가드 테스트가 아직 없다 |
| EV-3 | AC-GR-007 | `grep -c 'Discovery' .claude/skills/moai/workflows/run.md` | `0` | 1 | 생명주기 서술이 run 문서에 없다 |
| EV-4 | AC-GR-007 | `grep -c 'Qualification' .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/sync.md` | `…/run.md:0` `…/sync.md:0` | 1 | 같은 이유 |
| EV-5 | AC-GR-001 (판정식 반증 능력) | scratch 파일 3개에 `strip` 적용 후 `cmp` — 원본 복원 사례와 블록 밖 한 단어를 바꾼 변이 사례 | 복원 사례 `STRIP-RESTORES`; 변이 사례 `mutant cmp exit=1` | 0 / 1 | AC-GR-001 은 불변 가드라 RED-now 가 없다. 대신 판정식이 블록 밖 변경을 잡아낸다는 것을 변이로 확인했다 |
| EV-6 | AC-GR-004 | `moai constitution validate` | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)` | 0 | 불변 가드 기준선 |

EV-1~EV-4 는 녹색 경로가 M1~M5 이고 녹색 출력은 각 AC 의 「기대」다. AC-GR-008 의 RED 는 run phase M6 에서 불량 픽스처로 관측한다(현재는 테스트 파일 자체가 없다 — EV-2).

## §D. 품질 게이트와 완료 정의

### §D.1 MUST-PASS

AC-GR-001, 002, 003, 004, 005, 007, 008, 009, 016.

### §D.2 추적성

| REQ | AC |
|---|---|
| REQ-GR-001 | AC-GR-008, AC-GR-012 |
| REQ-GR-002 | AC-GR-001 |
| REQ-GR-003 | AC-GR-001, AC-GR-011 |
| REQ-GR-010 | AC-GR-015 |
| REQ-GR-011 | AC-GR-005, AC-GR-006 |
| REQ-GR-013 | AC-GR-015 |
| REQ-GR-015 | AC-GR-015 |
| REQ-GR-020 | AC-GR-006 (+ 검토) |
| REQ-GR-022 | AC-GR-006 (+ 검토) |
| REQ-GR-030 | AC-GR-013, AC-GR-006 |
| REQ-GR-040 | AC-GR-014, AC-GR-006 |
| REQ-GR-050 | AC-GR-004, AC-GR-006 |
| REQ-GR-052 | AC-GR-006 (+ 검토) |
| REQ-GR-060 | AC-GR-007 |
| REQ-GR-062 | AC-GR-007 |
| REQ-GR-063 | AC-GR-006 (+ 검토) — Push 조건 문언 |
| REQ-GR-070 | AC-GR-002, AC-GR-003 |
| REQ-GR-072 | AC-GR-009, AC-GR-008 |
| REQ-GR-073 | AC-GR-008 |
| REQ-GR-080 | AC-GR-010 |
| REQ-GR-090 | AC-GR-004 |

요구사항 21개 전부가 하나 이상의 AC 에 매핑된다.

### §D.3 기계로 판정하지 않는 것

- 블록 문장이 정확히 무엇을 **지시하는지**(예: 에스컬레이션이 질문이 아니라 보고라는 것)는 문자열 존재로만 부분 판정한다. 의미의 옳음은 plan-auditor 와 sync-auditor 의 검토 몫이다(REQ-GR-020~022, 051~052, 063~064).
- guided 세션의 모델이 contract 블록을 무시한다는 것은 이 AC 들이 증명하지 않는다 — 텍스트 불변만 증명한다(`design.md §6` 잔여 위험).

### §D.4 완료 정의

- MUST-PASS 전부 PASS, 나머지 AC PASS 또는 사유가 적힌 PASS-WITH-DEBT.
- `plan.md §B` 의 NC 전부 해소되어 기록됨.
- **[A1 감사 통과본으로 재확인]** 표시 항목 전부가 M0 에서 A1 감사 통과본과 대조됨.
- `progress.md §E.2` 에 `BASE` SHA, 재측정 원장, AC 표가 원문 증거와 함께 있음.
