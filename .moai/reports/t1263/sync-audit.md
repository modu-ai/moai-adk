# t1263 sync-audit — 수렴 노트 문구와 차단 동작의 일치

- 카드: t1263 (Class B, Tier S, SPEC 없음)
- 대상: `git diff 6b1e9bbd4 adbe6cffe` (브랜치 `WT-review-gate-wording`, HEAD `adbe6cffe`)
- 감사자: sync-auditor (독립, 소스 읽기 전용)
- 일자: 2026-09-26

## 판정

**PASS-WITH-DEBT** — 가중 점수 91.8 / 조화 평균 91.4

must-pass 차원(Functionality·Security)은 모두 통과했다. 발견 사항 5건은 모두 optional이다. 부채로 남는 것은 이번 diff가 만들지 않은 인접 경로 두 곳(F1·F2)이며, 후속 카드로 다룰 것을 권한다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | 대상 결함(required split)과 같은 수정으로 함께 고쳐진 경로(required FAIL + advisory PASS)에서 노트가 `required-backend FAIL: …` 로 시작하고, advisory 전용 충돌은 `(advisory, NOT a block)` 를 유지한다(아래 프로브 P1·P3·P4). 대상 테스트 `exit=0`. 감점 사유는 이번 diff 밖에 있는 인접 경로 P5·P7(F1·F2) |
| Security (25%) | 100/100 | PASS | 변경 범위는 문자열 조립 분기 하나다. 외부 입력·경로·권한·비밀 표면이 없고, 백엔드 이름과 게이트 토큰만 서식화한다 |
| Craft (20%) | 85/100 | PASS | 새 테스트가 옛 코드에서 RED임을 오버레이로 재현(`exit=1`, 단언 3건 실패). `gofmt -l` 출력 없음, `go vet` 출력 0바이트. 감점: P3 경로를 고정하는 케이스가 없고(F3), `describeRequiredFails` doc 주석이 낡았다(F4) |
| Consistency (15%) | 92/100 | PASS | 로컬·템플릿 SKILL.md 바이트 동일(`cmp exit=0`), catalog 해시 재생성 후 `internal/template` 테스트 통과. 기존 `describeRequiredFails` 를 재사용해 노트 접두 형식이 비분할 경로와 같다 |

## 경로별 동작 확인 (converge 전 경로)

오버레이(`go test -overlay`)로 트리를 건드리지 않고 프로브 테스트를 주입해 관측했다. `block=` 은 `multi_review_gate.go` 의 `blockReason` 출력이며, 게이트는 `overall_verdict == fail` 일 때만 이것을 내보낸다.

| 경로 | overall | 노트 | 판단 |
|---|---|---|---|
| P1 required split | fail | `required-backend FAIL: codex; cross-model disagreement: pass=[claude(required)] fail=[codex(required)]` | 일치 (이번 수정 대상) |
| P2 required FAIL만, 분할 없음 | fail | `required-backend FAIL: claude, codex` | 일치 (`describeRequiredFails`, 기존) |
| P3 required FAIL + advisory PASS | fail | `required-backend FAIL: codex; cross-model disagreement: pass=[glm(advisory)] fail=[codex(required)]` | 일치 — 옛 코드는 여기서도 `NOT a block` 이라고 했으므로 이번 수정이 함께 고쳤다 |
| P4 required PASS + advisory FAIL | pass | `cross-model disagreement (advisory, NOT a block): …` | 일치 (게이트 ALLOW) |
| P5 required FAIL + 백엔드 내부 synthesis note | fail | `cross-model disagreement detected; see per_backend_verdicts for details \| claude: signals diverged` | **불완전** — 모순 문구는 없지만 실패 백엔드 이름이 빠지고, 없는 cross-model 불일치를 말한다 (F1, 기존 결함) |
| P6 required split + synthesis note | fail | `required-backend FAIL: codex; … \| codex: signals diverged` | 일치 (append 경로 정상) |
| P7 required PASS + advisory FAIL + 명시 required codex inconclusive → `enforceRequiredGateUnmet` | fail | `required gate unmet (…): codex \| cross-model disagreement (advisory, NOT a block): …` | **혼동 소지** — 차단 사유가 앞에 오지만 차단되는 결과 안에 `NOT a block` 이 남는다 (F2, 기존 결함) |

`multi_review_gate.go` 는 노트를 해석하지 않고 `blockReason` 에서 그대로 되울릴 뿐이며, 차단 여부는 `OverallVerdict` 만으로 정한다. 따라서 문구 수정이 게이트 동작을 바꾸지 않는다는 점도 확인했다.

## 발견 사항 (structured defect-list)

- **F1** [Medium] [optional] `internal/cli/mcp_convergence.go:278-283, 454-457` — required FAIL(분할 없음)에 백엔드 내부 synthesis note가 붙으면 `disagreement=true` 가 되어 `describeDisagreement` 로 가고, pass 목록이 비어 일반 문구 `cross-model disagreement detected; …` 를 낸다. 결과: overall=fail·게이트 BLOCK인데 노트와 Stop 사유 어디에도 실패 백엔드 이름이 없다(AC-AMM-007 "residual_risk_note records which backend(s) failed" 위반 소지). 이번 diff가 만든 결함은 아니다(일반 분기는 기존 코드). 신뢰도: 높음(P5 실측). — Required fix: `converge` Step 3 에서 `requiredFails` 가 있으면 `describeRequiredFails` 를 항상 앞에 두고, 일반 불일치 문구는 실제 pass/fail 분할이 있을 때만 붙이도록 후속 카드로 처리한다.
- **F2** [Low] [optional] `internal/cli/mcp_convergence.go:846-850` (`enforceRequiredGateUnmet`) — 명시적 required 게이트 미충족이 overall을 fail로 뒤집어도 기존 advisory 노트 `(advisory, NOT a block)` 가 뒤에 그대로 붙는다. 그 절 자체는 advisory 충돌에 대해서는 사실이고 차단 사유가 먼저 나오지만, 차단된 Stop 사유 안에 `NOT a block` 이 들어 있어 이번 카드가 없애려던 종류의 혼동이 남는다. 신뢰도: 높음(P7 실측), 심각도는 읽는 사람에 따라 다름. 기존 결함. — Required fix: overall이 fail로 뒤집힐 때 advisory 절을 `(advisory)` 로만 줄이거나 `advisory conflict:` 로 바꾸는 방안을 후속 카드에서 검토한다.
- **F3** [Low] [optional] `internal/cli/convergence_note_wording_test.go:21-42` — 케이스가 2개(required split, advisory 전용)뿐이다. 같은 수정이 고친 P3(required FAIL + advisory PASS, overall=fail)은 고정되지 않았다. 새 분기를 `filterRequired(vs)` 대신 다른 조건(예: required PASS 존재 여부)으로 바꾸는 회귀는 현재 테스트를 통과한다. 신뢰도: 중간. — Required fix: P3 케이스 한 줄(`wantIn: "required-backend FAIL: codex"`, `wantNotIn: "NOT a block"`)을 추가한다.
- **F4** [Low] [optional] `internal/cli/mcp_convergence.go:415-418` — `describeRequiredFails` doc 주석이 "no-split case … where disagreement_flag is false 에서 쓰인다"고 적는데, 이번 diff 이후 분할 경로(`describeDisagreement`)에서도 호출된다. 신뢰도: 높음. — Required fix: 주석을 "required FAIL 이름을 나열한다; 분할·비분할 양쪽에서 쓴다"로 고친다.
- **F5** [Info] [optional] `internal/cli/mcp_convergence.go:427-431` — `describeDisagreement` 의 doc 주석이 `collectSynthesisNotes` 위에 붙어 있다(기존). 이번 diff와 무관하며 기록만 남긴다. — Required fix: 주석을 `describeDisagreement` 선언 바로 위로 옮긴다.

### 다른 표면의 `NOT a block` 문구 점검

- `.claude/skills/moai-ref-cross-model-audit/SKILL.md:195` (템플릿 미러 동일) "Disagreement is advisory, NOT a block" — 불변식 설명이며 "유일한 차단 모양은 required FAIL(case 2/3)"을 같은 문단에서 밝힌다. 사실과 맞으므로 수정 대상이 아니다.
- `docs-site/content/{en,ja,zh}/advanced/autonomous-loops.md:109` — "불일치 자체는 하드 블록하지 않는다"는 서술. 사실과 맞다.
- `.moai/specs/SPEC-AUDIT-MULTI-MODEL-001/design.md:135,142` — case 4(advisory) 예시. 사실과 맞다.
- `.moai/specs/SPEC-CODEX-DUAL-AGENTS-001/audit-plan.md:119` — 옛 노트를 인용한 과거 실행 기록이다. 이력이므로 고치지 않는 것이 옳다.
- 옛 문구를 단언하던 기존 테스트는 없다(`git grep "NOT a block" -- '*.go'` 결과가 새 테스트와 소스 두 곳뿐).

## Claim / Evidence / Baseline / Gaps / Residual-risk

**Claim**
1. required FAIL을 포함한 분할에서 노트는 더 이상 `advisory, NOT a block` 이라고 하지 않고 `required-backend FAIL: <이름>` 으로 시작한다.
2. advisory 전용 충돌은 `(advisory, NOT a block)` 을 유지하고 overall=pass다.
3. 새 테스트는 옛 코드에서 실패하고 새 코드에서 통과한다.
4. 로컬·템플릿 SKILL.md는 바이트 동일하고 catalog 해시가 맞다.

**Evidence** (이 실행, 이 트리에서 관측한 원문)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED GIT_DIR GIT_WORK_TREE && go test -count=1 -run 'TestConvergence|Convergence|Disagree|RequiredGate|MultiReview|AuditMulti|Participant|Divergence' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	4.349s
exit=0

$ unset … && go test -count=1 ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	63.741s
exit=0

$ unset … && go vet ./internal/cli/
(출력 0바이트)
exit=0

$ gofmt -l internal/cli/mcp_convergence.go internal/cli/convergence_note_wording_test.go
(출력 없음)

$ cmp .claude/skills/moai-ref-cross-model-audit/SKILL.md internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md
exit=0
```

RED 재현(옛 `mcp_convergence.go` 를 `git show 6b1e9bbd4:…` 로 꺼내 `-overlay` 로 대체, 트리 무변경):

```
--- FAIL: TestConvergenceNoteMatchesBlockDecision/required_split_blocks (0.00s)
    convergence_note_wording_test.go:51: residual_risk_note = "cross-model disagreement (advisory, NOT a block): pass=[codex(required)] fail=[claude(required)]", want it to contain "required-backend FAIL: claude"
    convergence_note_wording_test.go:56: … must not contain "NOT a block"
    convergence_note_wording_test.go:56: … must not contain "advisory"
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.761s
exit=1
```

경로 프로브 P1~P7 원문은 위 「경로별 동작 확인」 표에 옮겼다(`go test -overlay … -run TestAuditorProbeT1263 -v`, `exit=0`).

**Baseline-attribution**: HEAD `adbe6cffe` (브랜치 `WT-review-gate-wording`, `git status` clean), 기준 커밋 `6b1e9bbd4`. 모든 명령은 워크트리 `.claude/worktrees/t1263` 에서 env 스크럽과 한 복합 호출로 실행했다. 프로브 파일과 원시 출력은 세션 scratchpad(`…/scratchpad/t1263/`)에만 있고 트리에는 쓰지 않았다.

**Gaps**
- `./internal/cli/` 전체 스위트는 돌리지 않았다(지시 범위 밖, 판정은 develop CI 몫). 정규식에 걸리지 않는 cli 테스트가 이 문구에 기대는지는 `git grep` 으로만 확인했다(`.go` 파일 적중은 새 테스트와 소스뿐).
- darwin 외 플랫폼, `golangci-lint` 는 돌리지 않았다.
- 교차 모델 감사(`audit_multi`)는 호출하지 않았다 — 한 분기 문구 수정에 대한 Tier S 감사이고, 경로 동작은 프로브로 직접 관측했다.
- docs-site ko 로케일은 영어 키워드 grep 으로만 찾았으므로 한국어 표현의 동등 서술은 확인하지 않았다(en/ja/zh 서술이 사실과 맞으므로 위험은 낮다).

**Residual-risk**
- F1·F2의 인접 경로에서 Stop 사유가 여전히 불완전하거나 혼동을 줄 수 있다. 차단 동작 자체는 옳고 문구만의 문제다.
- F3 때문에 P3 경로의 문구 회귀는 현재 테스트로 잡히지 않는다.
- 노트 문자열을 파싱하는 외부 소비자(스크립트·대시보드)가 `cross-model disagreement (advisory` 접두를 키로 쓰고 있었다면 분할 사례에서 매칭이 바뀐다. 저장소 안에서는 그런 소비자를 찾지 못했다.

## 재감사 (F2 한정, a0b3f1ece)

- 범위: `git diff 0062a1e7d a0b3f1ece` 만 본다(`internal/cli/mcp_convergence.go` +9/−3, `internal/cli/convergence_note_wording_test.go` +21). F1과 이 diff 밖의 것은 다시 감사하지 않았다.
- 일자: 2026-09-26, 감사자: sync-auditor (소스 읽기 전용)

### F2 판정: 닫힘

수정 내용은 두 가지다. `describeDisagreement` 의 advisory 문구를 상수 `advisoryDisagreementQualifier` (`" (advisory, NOT a block)"`)로 뽑았고, `enforceRequiredGateUnmet` 가 overall을 fail로 뒤집을 때 앞선 노트에서 그 상수를 첫 번째 한 번 지운다(`strings.Replace(…, 1)`). 분할 내용 자체는 남고, 게이트 라벨 `glm(advisory)` 이 목록에 그대로 있어 어느 백엔드가 advisory였는지도 잃지 않는다.

**`enforceRequiredGateUnmet` 가 받을 수 있는 노트 모양 전부에서 확인했다.** 호출자는 `runMultiAudit` 한 곳(`mcp_convergence.go:762`)이고 입력은 곧바로 앞의 `converge(verdicts)` 결과다. 그 사이에 노트를 고치는 코드는 없다. 오버레이로 프로브 테스트를 주입해 각 모양의 전·후를 관측했다(트리 무변경).

| 모양 | 전 (overall, 상수 출현 수) | 후 (overall, 상수 출현 수) | 앞선 노트 내용 보존 |
|---|---|---|---|
| advisory 전용 분할 | pass, 1 | fail, 0 | 보존 |
| advisory 분할 + synthesis note 2건 | pass, 1 | fail, 0 | 보존 (synthesis 절 원문 유지) |
| 빈 노트 | pass, 0 | fail, 0 | 해당 없음 — `" \| "` 없이 미충족 절만 |
| required 분할 (상수 없음) | fail, 0 | fail, 0 | 바이트 그대로 |
| 일반 불일치 문구 (synthesis만) | pass, 0 | fail, 0 | 바이트 그대로 |

**상수가 두 번 나올 수 있는가 — 없다.** `describeDisagreement` 는 `converge` 한 번에 한 번만 불리고 노트의 맨 앞에 온다(`mcp_convergence.go:278-292`). 뒤에 붙는 것은 synthesis note뿐인데, 비테스트 코드에서 그 값을 만드는 곳은 `describeSignalDivergence`(`mcp_codex.go:1758-1771`) 하나이고 형식이 `"codex signals diverged: <source>=<verdict>, …; adopted <verdict>"` 로 고정돼 상수를 담을 수 없다. 따라서 첫 번째 한 번만 지우는 `Replace(…, 1)` 로 충분하고, 그 한 번은 늘 `describeDisagreement` 가 쓴 자리다. 프로브에서도 모든 모양의 출현 수가 0 또는 1이었다.

**남은 발신처.** 비테스트 Go 코드에서 `NOT a block` 을 담은 것은 상수 선언(`mcp_convergence.go:447`)과 `converge` doc 주석의 case 4 설명(`:179`)뿐이다. 주석은 `converge` 단독 동작으로서는 사실이고, 바로 아래 `:185-188` 이 `enforceRequiredGateUnmet` 가 결과를 덮을 수 있음을 밝힌다. 노트를 내보내는 경로는 이제 상수 하나를 거친다.

### 새 서브테스트의 RED/GREEN

`required gate unmet flips advisory note` 는 상수가 아니라 리터럴 `"NOT a block"` 으로 단언하므로, 상수 이름이나 값을 바꾸는 회귀도 잡는다. `0062a1e7d` 의 `mcp_convergence.go` 로 오버레이하면 이 서브테스트만 실패하고 나머지 세 서브테스트는 통과한다 — 테스트가 이 수정 하나를 가른다는 뜻이다.

### 발견 사항 (재감사 diff 한정)

- **F6** [Low] [optional] `internal/cli/convergence_note_wording_test.go:56-72` — 새 서브테스트는 미충족 절이 있고 `NOT a block` 이 없음만 단언한다. 앞선 노트를 통째로 버리는 회귀(예: `note += ""`)도 통과한다. 이번 수정의 목적은 "분할은 남기고 문구만 뺀다"이므로 그 절반은 고정되지 않았다. 신뢰도: 높음(단언 목록을 읽어 확인, 변이 실행은 하지 않음). — Required fix: `wantIn` 에 `"cross-model disagreement: pass=[claude(required)] fail=[glm(advisory)]"` 한 줄을 추가한다.

blocking 발견 사항은 없다.

### 갱신된 점수와 판정

| 차원 | 이전 | 재감사 후 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90 | 93 | P7(F2) 해소, 5개 노트 모양 전부 관측. F1은 재감사하지 않았으므로 감점 유지 |
| Security (25%) | 100 | 100 | 변경은 상수 추출과 문자열 치환 하나. 입력 표면 없음 |
| Craft (20%) | 85 | 85 | RED 재현 확인(+), F6 신규(−). F3·F4는 이 범위에서 재확인하지 않았다 |
| Consistency (15%) | 92 | 92 | 상수 도입이 문구 단일 출처를 만든다. 기존 서식과 같다 |

가중 점수 **93.0**, 조화 평균 **92.2**.

**종합 판정: PASS-WITH-DEBT** — must-pass(Functionality·Security) 통과, F2 닫힘. 부채로 남는 것은 이번 재감사가 다루지 않은 F1(직전 보고 기준 미해소)과 optional F6이다.

### Claim / Evidence / Baseline / Gaps / Residual-risk

**Claim**
1. F2가 닫혔다 — 명시 required 게이트 미충족으로 fail로 뒤집힌 결과의 노트에 `NOT a block` 이 남지 않는다.
2. 상수는 노트에 많아야 한 번 나오므로 `Replace(…, 1)` 로 빠짐없이 지워진다.
3. 새 서브테스트는 `0062a1e7d` 소스에서 실패하고 HEAD에서 통과한다.
4. 비테스트 코드에서 이 문구를 노트로 내보내는 곳은 상수 하나뿐이다.

**Evidence** (이 실행, 이 트리에서 관측한 원문)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED GIT_DIR GIT_WORK_TREE && go test -count=1 -run 'TestConvergence|Convergence|Disagree|RequiredGate|GateUnmet|MultiReview|AuditMulti|Participant|Divergence' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	3.431s
exit=0

$ unset … && go vet ./internal/cli/
exit=0
(출력 0바이트)
```

RED 재현(`git show 0062a1e7d:internal/cli/mcp_convergence.go` 를 scratchpad로 꺼내 `-overlay` 로 대체):

```
=== RUN   TestConvergenceNoteMatchesBlockDecision/required_gate_unmet_flips_advisory_note
    convergence_note_wording_test.go:70: residual_risk_note = "required gate unmet (explicitly configured required, no verdict): codex | cross-model disagreement (advisory, NOT a block): pass=[claude(required)] fail=[glm(advisory)]", must not contain "NOT a block" on a failing verdict
--- FAIL: TestConvergenceNoteMatchesBlockDecision (0.00s)
    --- FAIL: TestConvergenceNoteMatchesBlockDecision/required_gate_unmet_flips_advisory_note (0.00s)
    --- PASS: TestConvergenceNoteMatchesBlockDecision/required_split_blocks (0.00s)
    --- PASS: TestConvergenceNoteMatchesBlockDecision/required_fail_against_advisory_pass_blocks (0.00s)
    --- PASS: TestConvergenceNoteMatchesBlockDecision/advisory-only_conflict_does_not_block (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.595s
exit=1
```

노트 모양 프로브(HEAD, 오버레이로 `zz_f2_probe_test.go` 추가, `-run 'TestZZF2ProbeNoteShapes|TestConvergenceNoteMatchesBlockDecision' -v`):

```
[advisory-only split]
  before(overall=pass, qualifierCount=1): "cross-model disagreement (advisory, NOT a block): pass=[claude(required)] fail=[glm(advisory)]"
  after (overall=fail, qualifierCount=0): "required gate unmet (explicitly configured required, no verdict): codex | cross-model disagreement: pass=[claude(required)] fail=[glm(advisory)]"
[advisory split + synthesis notes]
  before(overall=pass, qualifierCount=1): "cross-model disagreement (advisory, NOT a block): pass=[claude(required)] fail=[glm(advisory)] | claude: claude signals diverged: a=pass, b=fail; adopted pass | codex: codex signals diverged: x=pass, y=fail; adopted inconclusive"
  after (overall=fail, qualifierCount=0): "required gate unmet (explicitly configured required, no verdict): codex | cross-model disagreement: pass=[claude(required)] fail=[glm(advisory)] | claude: claude signals diverged: a=pass, b=fail; adopted pass | codex: codex signals diverged: x=pass, y=fail; adopted inconclusive"
[empty note]
  before(overall=pass, qualifierCount=0): ""
  after (overall=fail, qualifierCount=0): "required gate unmet (explicitly configured required, no verdict): codex"
[required split (no qualifier)]
  before(overall=fail, qualifierCount=0): "required-backend FAIL: glm; cross-model disagreement: pass=[claude(required)] fail=[glm(required)]"
  after (overall=fail, qualifierCount=0): "required gate unmet (explicitly configured required, no verdict): codex | required-backend FAIL: glm; cross-model disagreement: pass=[claude(required)] fail=[glm(required)]"
[generic disagreement (synthesis only)]
  before(overall=pass, qualifierCount=0): "cross-model disagreement detected; see per_backend_verdicts for details | claude: claude signals diverged: a=pass, b=fail; adopted pass"
  after (overall=fail, qualifierCount=0): "required gate unmet (explicitly configured required, no verdict): codex | cross-model disagreement detected; see per_backend_verdicts for details | claude: claude signals diverged: a=pass, b=fail; adopted pass"
--- PASS: TestZZF2ProbeNoteShapes (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.905s
exit=0
```

남은 발신처(`grep -rn 'advisory, NOT\|NOT a block' internal pkg cmd`, 템플릿 제외) — 비테스트 적중은 `internal/cli/mcp_convergence.go:179`(doc 주석)과 `:447`(상수 선언) 두 줄뿐이다.

**Baseline-attribution**: HEAD `a0b3f1ece` (브랜치 `WT-review-gate-wording`, 측정 시점 `git status --short` 출력 없음), 비교 기준 `0062a1e7d`. 모든 go 명령은 워크트리 `.claude/worktrees/t1263` 에서 env 스크럽과 한 복합 호출로 실행했다. 옛 소스·오버레이 JSON·프로브 테스트는 세션 scratchpad(`…/scratchpad/f2/`)에만 있고 트리에는 쓰지 않았다. 이 절을 덧붙인 것이 이 재감사가 트리에 남긴 유일한 변경이다.

**Gaps**
- F1, 그리고 직전 보고의 F3·F4·F5가 `0062a1e7d` 에서 처리됐는지는 재확인하지 않았다(지시 범위 밖).
- `runMultiAudit` 의 `gateAssumedNote` 접두(`mcp_convergence.go:763-765`)가 붙는 경로는 프로브하지 않았다. 그 접두는 치환 뒤에 앞에 붙으므로 이번 수정과 상호작용하지 않는다고 코드로만 판단했다.
- F6 는 단언 목록을 읽어 판단했고 변이 테스트로 재현하지 않았다.
- `./internal/cli/` 전체 스위트, `golangci-lint`, darwin 외 플랫폼은 돌리지 않았다(판정은 develop CI 몫).
- 교차 모델 감사(`audit_multi`)는 호출하지 않았다 — 한 함수의 문자열 치환에 대한 범위 한정 재감사이고, 동작은 프로브로 직접 관측했다.

**Residual-risk**
- 앞으로 synthesis note를 만드는 새 생산자가 자유 텍스트를 싣게 되면, 상수가 두 번 나오거나 synthesis 쪽 문구가 대신 지워질 수 있다. 지금은 생산자가 하나이고 형식이 고정돼 있다.
- F6 때문에 "분할 내용 보존" 쪽 회귀는 현재 테스트로 잡히지 않는다.
- 노트 문자열을 키로 쓰는 외부 소비자가 있다면 fail 결과에서 `(advisory, NOT a block)` 이 사라지는 것이 매칭을 바꾼다. 저장소 안에서는 그런 소비자를 찾지 못했다(직전 보고와 같다).
