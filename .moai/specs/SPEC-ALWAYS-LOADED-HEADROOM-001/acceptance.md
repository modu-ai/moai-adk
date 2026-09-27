# SPEC-ALWAYS-LOADED-HEADROOM-001 — 인수 조건

모든 AC 는 Given-When-Then 이며 이진 판정된다. 명령은 워크트리 루트에서 실행한다. 판정서는 `.moai/reports/t1226/verdict.md`, 기계 판독용 후보 표는 `.moai/reports/t1226/candidates-init.tsv` 와 `.moai/reports/t1226/candidates-live.tsv` 다.

---

## §D. AC 매트릭스

| AC | 대상 REQ | 판정 | 심각도 |
|---|---|---|---|
| AC-ALH-001 | REQ-ALH-001, REQ-ALH-014 | 기계적 | MUST-PASS |
| AC-ALH-002 | REQ-ALH-002 | 기계적 | MUST-PASS |
| AC-ALH-003 | REQ-ALH-003, REQ-ALH-004, REQ-ALH-005, REQ-ALH-015 | 기계적 | MUST-PASS |
| AC-ALH-004 | REQ-ALH-006 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-005 | REQ-ALH-007, REQ-ALH-008 | 기계적 | MUST-PASS |
| AC-ALH-006 | REQ-ALH-009 | 기계적 | MUST-PASS |
| AC-ALH-007 | REQ-ALH-010, REQ-ALH-011, REQ-ALH-012 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-008 | REQ-ALH-013 | 기계적 | MUST-PASS |

### 후보 표 TSV 열 규약

두 TSV 는 머리 줄 하나와 후보 행으로 이루어지며 열은 다음 순서다. AC-ALH-003·004·006 이 이 열 번호로 판정한다.

| 열 | 이름 | 값 |
|---:|---|---|
| 1 | `file` | 18경로 중 하나 |
| 2 | `section` | 절 제목(문단 후보는 `절 제목 ¶n`) |
| 3 | `chars` | 정수 자수 또는 `미측정` |
| 4 | `mech` | `M1` · `M1p` · `M2` |
| 5 | `bind` | 후보 안 구속 조항 줄 개수(정수) |
| 6 | `c1` | 조건 1 판정 `Y`/`N` |
| 7 | `c2` | 조건 2 판정 `Y`/`N`/`NA` |
| 8 | `c3` | 조건 3 판정 `Y`/`N`/`NA` |
| 9 | `c4` | 조건 4 판정 `Y`/`N`/`NA` |
| 10 | `verdict` | `ADMIT` · `REJECT` |
| 11 | `reason` | 기각 사유(허용 행은 `-`) |
| 12 | `evidence` | 증거 파일 경로(`.moai/reports/t1226/` 아래) |

`NA` 는 M2 행처럼 목적지가 없는 기제에서만 쓴다.

---

### AC-ALH-001 — 판정서 머리의 고정 항목과 폐기 수치 격리

**Given** 판정서가 존재할 때,
**When** 첫 20줄과 본문 전체를 아래 명령으로 읽을 때,
**Then** 첫 20줄에 네 항목이 모두 있고, 197,897 과 198,361 이 나오는 모든 줄에 `폐기` 가 함께 있다.

```bash
head -20 .moai/reports/t1226/verdict.md | grep -F -c -e '레인 백엔드: Claude Opus 5.5 (claude-opus-5-5)'   # 1 이상
head -20 .moai/reports/t1226/verdict.md | grep -F -c -e '7fe658815eb0d4110b9acadad56e5a85bee3ed3f'       # 1 이상
head -20 .moai/reports/t1226/verdict.md | grep -E -c '199,?111'                                          # 1 이상
head -20 .moai/reports/t1226/verdict.md | grep -E -c '49,?111'                                           # 1 이상
grep -E '197,?897|198,?361' .moai/reports/t1226/verdict.md | grep -v -c '폐기'                            # 0
```

FAIL 조건: 네 `grep -c` 중 하나라도 0, 또는 마지막 명령이 0 이 아니다. 측정 명령 전문은 판정서 첫 절의 fenced 블록이 아래 AC-ALH-002 의 `S_live` 블록과 문자 단위로 같아야 한다(검토).

### AC-ALH-002 — 두 표면을 각각 쟀다

**Given** 이 트리에서 빌드한 바이너리와 격리된 scratch 디렉터리가 있을 때,
**When** `S_live` 는 아래 블록으로, `S_init` 은 같은 18경로를 `moai init` 산출 트리에서 잴 때,
**Then** 판정서에 `## S_init` 과 `## S_live` 두 절이 있고, 각 절에 명령·기준 커밋·관측 `total` 값이 적혀 있다.

`S_live` 측정 블록(기준선 `199111 total`, `7fe658815`):

```bash
wc -m CLAUDE.md AGENTS.md \
  .moai/config/sections/user.yaml .moai/config/sections/language.yaml \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/moai-constitution.md \
  .claude/rules/moai/core/moai-mcp-tools.md \
  .claude/rules/moai/core/native-idiom-and-register.md \
  .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/skill-routing.md | tail -1
```

`S_init` 산출 절차(판정서에 실제 실행한 명령을 그대로 옮긴다):

```bash
go build -o "$SCRATCH/moai" ./cmd/moai
mkdir -p "$SCRATCH/init-home" "$SCRATCH/init-proj"
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && HOME="$SCRATCH/init-home" "$SCRATCH/moai" init "$SCRATCH/init-proj" --non-interactive
# 이어서 "$SCRATCH/init-proj" 안에서 위 S_live 블록과 같은 18경로를 wc -m
```

`moai init` 의 사용법은 `moai init [project-name] [--flags]` 이다. 인자가 절대경로를 받는지 이름만 받는지는 plan 단계에서 확인하지 않았다 — run 단계가 `--help` 로 확인한 뒤 실제로 쓴 형태를 판정서에 적는다. `HOME` 격리는 실제 사용자 설정 오염을 막기 위한 것이며, 빠뜨렸다면 그 사실을 판정서 Gaps 에 적는다.

```bash
grep -c -E '^## S_init' .moai/reports/t1226/verdict.md    # 1
grep -c -E '^## S_live' .moai/reports/t1226/verdict.md    # 1
```

FAIL 조건: 두 절 중 하나가 없거나, 한 절에 `total` 관측 줄 또는 명령이 없거나, `S_init` 측정이 기존 설치 바이너리(`~/go/bin/moai`)로 이뤄졌다. 설치 바이너리는 이 트리의 코드를 담는다는 보장이 없다.

### AC-ALH-003 — 후보 표의 완결성과 허용 규칙

**Given** 두 TSV 가 존재할 때,
**When** 아래 `awk` 를 각 TSV 에 돌릴 때,
**Then** 출력이 `BAD=0` 이다.

```bash
for f in .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv; do
  awk -F'\t' 'NR>1 {
    if (NF!=12) bad++;
    if ($10=="ADMIT" && $5+0>0) bad++;                         # REQ-ALH-005: 구속 줄을 담은 후보 허용 금지
    if ($10=="ADMIT" && $6!="Y") bad++;                        # 조건 1
    if ($10=="ADMIT" && $4=="M1" && ($7!="Y"||$8!="Y"||$9!="Y")) bad++;  # 조건 2~4
    if ($1 ~ /AGENTS\.md/ && $4=="M1" && $10=="ADMIT") bad++;  # REQ-ALH-004
    if ($10=="REJECT" && ($11=="-"||$11=="")) bad++;           # 기각 사유 필수
    if ($12=="") bad++;
  } END {print FILENAME" BAD="bad+0}' "$f"
done
```

조건 4 의 증거(REQ-ALH-015)는 행마다 `evidence` 파일에 라이브·템플릿 두 트리의 목적지 `wc -m` 출력을 담아야 하며, 이미 40,000자를 넘은 파일(`worktree-integration.md` · `session-handoff-examples.md` · `kanban-dispatch-detail.md` · `spec-workflow.md`, 그리고 측정 시점에 새로 넘은 파일)이 목적지인 `ADMIT` 행은 0건이어야 한다(검토 — 목적지 경로가 `reason`/`evidence` 에 적힌다).

FAIL 조건: 어느 TSV 든 `BAD` 가 0 이 아니다.

### AC-ALH-004 — M2 자수는 실제 압축 시도로 쟀다

**Given** 두 TSV 의 `mech == M2` 행이 있을 때,
**When** 아래 명령으로 `ADMIT` 인 M2 행의 증거 파일을 확인할 때,
**Then** 출력이 `MISSING=0` 이고, `chars` 가 `미측정` 인 행은 모두 판정서 `## Gaps` 절에 나열된다.

```bash
for f in .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv; do
  awk -F'\t' 'NR>1 && $4=="M2" && $10=="ADMIT" {print $12}' "$f" \
    | while read -r p; do [ -f "$p" ] || echo "missing $p"; done
done | wc -l | awk '{print "MISSING="$1}'
```

증거 파일은 scratch 사본에 가한 압축 diff 와, 그 사본에서 동결 다중집합 sha256 이 `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` 로 유지됨을 보인 출력을 함께 담는다(검토). 수율을 곱해 채운 값은 허용하지 않는다.

FAIL 조건: `MISSING` 이 0 이 아니다, 또는 `미측정` 행이 Gaps 절에 없다.

### AC-ALH-005 — `P절` 재조정이 하나로 닫혔다

**Given** 절 분할 스크립트가 `.moai/reports/t1226/` 아래 커밋돼 있을 때,
**When** 그 스크립트를 `git archive 172ef22eb` 로 풀어 낸 원 트리에서 다시 돌리고 판정서를 읽을 때,
**Then** 판정서에 재조정 선택 줄이 정확히 하나 있고 F 가 단일값으로 적혀 있다.

```bash
ls .moai/reports/t1226/*.py .moai/reports/t1226/*.sh 2>/dev/null | wc -l                 # 1 이상
grep -c -E '^P절 재조정: \((a|b)\)' .moai/reports/t1226/verdict.md                        # 1
grep -c -E '^F = [0-9]{3},?[0-9]{3}$' .moai/reports/t1226/verdict.md                     # 1
grep -c -F '/tmp/' .moai/reports/t1226/verdict.md                                         # 0
```

FAIL 조건: 네 값 중 하나라도 기대와 다르다. 판정서는 재실행 출력 원문(kanban-dispatch.md 행 포함)을 함께 담아 (a)·(b) 를 가른 근거를 보인다(검토).

### AC-ALH-006 — `T_min` 산술이 후보 표에서 재현된다

**Given** 판정서의 표면별 `A_adm`·`R`·현재 합계·`T_min` 줄과 두 TSV 가 있을 때,
**When** TSV 의 `ADMIT` 행 자수 합을 구해 판정서 값과 대조할 때,
**Then** 표면마다 `A_adm` 이 행 합과 같고 `T_min = 현재 합계 − A_adm + R` 이 성립한다.

```bash
awk -F'\t' 'NR>1 && $10=="ADMIT" && $3 ~ /^[0-9]+$/ {s+=$3} END {print "A_adm_init="s}' .moai/reports/t1226/candidates-init.tsv
awk -F'\t' 'NR>1 && $10=="ADMIT" && $3 ~ /^[0-9]+$/ {s+=$3} END {print "A_adm_live="s}' .moai/reports/t1226/candidates-live.tsv
grep -E '^(A_adm|R|T_min|current)_(init|live) = ' .moai/reports/t1226/verdict.md
```

판정서는 표면마다 `current_<s> = N` · `A_adm_<s> = N` · `R_<s> = N` · `T_min_<s> = N` 네 줄을 이 철자로 적는다(쉼표 없는 정수). `R` 줄 옆에는 포인터 줄 원문과 그 `wc -m` 출력이 있어야 한다.

FAIL 조건: 어느 표면이든 네 줄 중 하나가 없거나, `A_adm` 이 TSV 합과 다르거나, 산술이 성립하지 않는다.

### AC-ALH-007 — 판정 토큰과 상신 절차

**Given** 판정서가 완성됐을 때,
**When** 표면별 판정 줄과 상신 절을 읽을 때,
**Then** 표면마다 판정 토큰이 정확히 하나이고, 토큰이 규칙(REQ-ALH-010)과 맞으며, `S_init` 이 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 이면 상신 절이 (a)~(d) 네 항목을 모두 담는다.

```bash
grep -c -E '^verdict_init = (ACHIEVABLE|STRUCTURALLY-INFEASIBLE-UNDER-FREEZE|UNDETERMINED)$' .moai/reports/t1226/verdict.md  # 1
grep -c -E '^verdict_live = (ACHIEVABLE|STRUCTURALLY-INFEASIBLE-UNDER-FREEZE|UNDETERMINED)$' .moai/reports/t1226/verdict.md  # 1
# S_init 이 INFEASIBLE 인 경우에만:
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' .moai/reports/t1226/verdict.md | grep -c -E '^### \((a|b|c|d)\)'      # 4
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' .moai/reports/t1226/verdict.md | grep -c -E '결정했다|승인했다|해제한다$' # 0
```

규칙 대조: `T_min < 150000` ⇔ `ACHIEVABLE`. `UNDETERMINED` 는 `미측정` 행 자수 합 `U` 에 대해 `T_min − U < 150000 ≤ T_min` 일 때만 쓴다(`U` 를 판정서에 `U_<s> = N` 으로 적는다). 상신 절의 권고 문장은 `RECOMMEND:` 로 시작한다(REQ-ALH-012, 검토).

FAIL 조건: 토큰 줄 개수가 1 이 아니다, 토큰이 규칙과 어긋난다, INFEASIBLE 인데 (a)~(d) 가 4개가 아니다, 또는 상신 절에 레인의 결정 서술이 있다.

### AC-ALH-008 — 계수 파일과 미러를 건드리지 않았다

**Given** 기준 커밋 `7fe658815` 와 카드 HEAD 가 있을 때,
**When** 18개 계수 파일과 그 템플릿 미러에 대해 `git diff` 를 돌리고 동결 다중집합을 다시 잴 때,
**Then** `git diff` 가 비어 있고(종료 코드 0) 해시가 기준선과 같다.

```bash
git diff --quiet 7fe658815 HEAD -- CLAUDE.md AGENTS.md \
  .moai/config/sections/user.yaml .moai/config/sections/language.yaml \
  .claude/rules/moai/core .claude/rules/moai/workflow \
  internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl \
  internal/template/templates/.moai/config/sections/user.yaml.tmpl \
  internal/template/templates/.moai/config/sections/language.yaml.tmpl \
  internal/template/templates/.claude/rules/moai; echo "exit=$?"     # exit=0
```

동결 다중집합은 `SPEC-ALWAYS-LOADED-DIET-002/acceptance.md` AC-ALD2-002 의 파이프라인 전문으로 재며 기대값은 `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` 이다(이 트리, 2026-09-27 manager-spec 재실행으로 확인).

FAIL 조건: `exit` 가 0 이 아니거나 해시가 다르다.

---

## §D.1 판정 게이트

- MUST-PASS 8개 — AC-ALH-001 · AC-ALH-002 · AC-ALH-003 · AC-ALH-004 · AC-ALH-005 · AC-ALH-006 · AC-ALH-007 · AC-ALH-008. 하나라도 FAIL 이면 SPEC 은 FAIL 이다.
- 품질 게이트: `moai spec lint SPEC-ALWAYS-LOADED-HEADROOM-001` 가 `No findings` 로 끝난다.
- 이 SPEC 의 판정은 「150,000 을 달성했는가」가 아니다. 측정이 완결됐고 판정 토큰이 규칙대로 붙었는가다. `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 와 `UNDETERMINED` 도 AC 를 통과할 수 있다.

## §D.2 추적성

| REQ | AC |
|---|---|
| REQ-ALH-001 | AC-ALH-001 |
| REQ-ALH-002 | AC-ALH-002 |
| REQ-ALH-003 | AC-ALH-003 |
| REQ-ALH-004 | AC-ALH-003 |
| REQ-ALH-005 | AC-ALH-003 |
| REQ-ALH-006 | AC-ALH-004 |
| REQ-ALH-007 | AC-ALH-005 |
| REQ-ALH-008 | AC-ALH-005 |
| REQ-ALH-009 | AC-ALH-006 |
| REQ-ALH-010 | AC-ALH-007 |
| REQ-ALH-011 | AC-ALH-007 |
| REQ-ALH-012 | AC-ALH-007 |
| REQ-ALH-013 | AC-ALH-008 |
| REQ-ALH-014 | AC-ALH-001 |
| REQ-ALH-015 | AC-ALH-003 |

## §D.3 경계 사례

- **`S_init` 의 18경로가 다르다** — `moai init` 트리에 18경로 중 일부가 없거나 경고가 세는 집합이 다르면, 판정서는 그 차이를 기록하고 실제 존재하는 경로 목록으로 잰 값과 누락 경로를 함께 적는다. 목록을 조용히 바꾸지 않는다.
- **후보가 두 기제에 걸친다** — 한 절의 일부는 M1, 나머지는 M2 로 갈 수 있으면 두 행으로 나눠 적고 자수를 이중 계상하지 않는다.
- **포인터 재유입이 제거량보다 크다** — `R` 이 해당 후보 자수를 넘는 후보는 순감이 음수이므로 `REJECT`(사유 `net-negative`)로 적는다.
