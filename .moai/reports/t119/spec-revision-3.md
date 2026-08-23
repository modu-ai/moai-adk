# SPEC-TODO-ANALYSIS-001 — F5 개정 (iteration 3, 재감사 없음)

- 기준 커밋: `fbdfd8363` (직전 개정 + plan-audit PASS 0.920 이 고정된 커밋). 편집 시작 시점 워킹트리 clean.
- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t119`, branch `WT-todo-auto-analysis`
- 범위: **F5 한 건만**. 커밋·푸시 하지 않음.
- 예산: REQ **15** (불변), AC **16** (불변), Out of Scope H3 **6** (불변). REQ·AC 신설 0건, 기존 AC 주장 내용 변경 0건.
- 변경 규모: 3파일 / `+11 −7` (`git diff --stat`)

---

## 1. 무엇을 바꿨는가

감사관이 준 두 갈래 중 **후자(산문을 요구사항 수준으로 낮추기)** 를 택했고, 그 안에서 **표식 이름도 바꾸는 쪽**을 골랐다.

### 1.1 이름을 바꾼 이유 (지시문이 요구한 선택 근거)

`unreviewed` → **`machine-only`**.

이름을 유지하고 §B.3 산문에만 의미를 명시하는 선택도 허용됐지만, 택하지 않았다. 이유는 **읽는 사람이 다르기 때문**이다:

- §B.3 산문을 읽는 것은 이 SPEC 을 읽는 사람이다.
- `unreviewed` 라는 문자열을 읽는 것은 `moai todo list` 출력을 보는 **운영자**다. 운영자는 SPEC 을 열지 않는다.

F5 의 결함은 "산문이 요구사항보다 강한 성질을 주장한다"인데, 그 과잉 주장을 실제로 운영자에게 전달하는 매체는 산문이 아니라 **렌더되는 토큰 그 자체**다. `unreviewed` 는 문자 그대로 "검토되지 않음"을 말하고, 뒤집으면 표식이 없을 때 "검토됐음"을 함의한다 — REQ-TA-013 이 실제로 확인하는 것은 `source: agent` 소견의 존재 유무뿐인데도. 산문만 고치면 SPEC 안의 모순은 사라지지만 **운영자에게 전달되는 거짓말은 그대로 남는다.**

`machine-only` 를 고른 이유:

- 술어를 그대로 옮긴 이름이다 — "기계 소견만 있고 에이전트 소견은 없다". REQ-TA-013 의 조건절과 1:1 대응하고, 검토 여부에 대해 아무 주장도 하지 않는다.
- 스키마가 이미 정의한 어휘를 재사용한다(`source ∈ {mechanical, agent}`). `source: mechanical` 을 아는 운영자는 `machine-only` 를 사전 없이 읽는다.
- 짧아서 `list` 의 들여쓴 소견 줄에 인라인으로 붙는다(REQ-TA-011 의 렌더 형태를 바꾸지 않는다).

### 1.2 편집 목록

| # | 무엇 | file:line (개정 후) |
|---|---|---|
| A | §B.3 표의 L2 `주체` 열: `manager-lead 가 relate 로 기록` → **`relate` 를 호출하는 누구나 — 관행상 `manager-lead` (REQ-TA-008 은 행위자를 제약하지 않는다)**. F5 가 지목한 "§B.3 이 L2 를 manager-lead 에 배정한다"는 바로 그 문장이다 | `spec.md:101` |
| B | §B.3 본문 1문단을 3문단으로 교체. 표식이 추적하는 것이 **검토가 아니라 에이전트 출처 기록의 유무**임을 명시하고, REQ-TA-008 의 무제약을 지면에 올리고, CLI 가 호출자를 강제할 수 없다는 한계를 §B.1 의 잔여 위험과 같은 뿌리로 묶고, 그럼에도 표식이 무엇을 잡고 무엇을 못 잡는지를 갈랐다. 삭제한 과잉 주장: "L2 부재를 다루는 **유일한 정직한 방법**" | `spec.md:103`, `spec.md:105`, `spec.md:107` |
| C | REQ-TA-013 의 표식 토큰 교체 + 꼬리절 추가: `; the mark records the ABSENCE of an agent-sourced finding for the pair, not that any review took place.` **술어(조건절)는 한 글자도 바뀌지 않았다** — State-driven `While … shall` 패턴 유지 | `spec.md:147` |
| D | Out of Scope — 의미 기반 중복 탐지 불릿의 토큰 교체 + 한 문장 보강(에이전트 없는 호출은 `source: agent` 소견을 만들지 않으므로 표식이 그대로 남는다) | `spec.md:173` |
| E | AC-TA-011 제목의 토큰 교체 | `acceptance.md:113` |
| F | AC-TA-011 Then 절의 토큰 교체. **주장 구조는 불변** — ①에 표식이 붙고 ②에서 걷힌다 | `acceptance.md:118` |
| G | M4 `runTodoList` 항목의 토큰 교체 | `plan.md:104` |

### 1.3 부수 정정 (내가 만든 문장의 구두점)

B/C/D 편집 직후 한 문장에 em-dash 가 2~3개 겹치는 구간이 생겨 같은 줄에서 정리했다 — 새 내용이 아니라 내가 방금 쓴 문장의 가독성 수선이다.

- REQ-TA-013: `machine-only` 뒤 세 번째 em-dash → 세미콜론 (`spec.md:147`)
- Out of Scope 불릿: 두 번째 em-dash → 마침표 + 새 문장 (`spec.md:173`)

---

## 2. 토큰 치환 전수 확인 (실측)

이름을 바꿨으므로 SPEC 안에서 이름이 갈리지 않았는지 전수 확인했다.

```
$ grep -rc "unreviewed" .moai/specs/SPEC-TODO-ANALYSIS-001/{spec,acceptance,plan,progress}.md
spec.md:0        acceptance.md:0        plan.md:0        progress.md:0

$ grep -ro "unreviewed" .moai/specs/SPEC-TODO-ANALYSIS-001/ | wc -l
0

$ grep -ro "machine-only" .moai/specs/SPEC-TODO-ANALYSIS-001/ | wc -l
7
```

- **잔존 `unreviewed` 0건** — 디렉터리 전체(`-r`) 기준. 개정 전 6건(`spec.md` 3 / `acceptance.md` 2 / `plan.md` 1, `progress.md` 0)이 전부 치환됐다.
- `machine-only` **7건 / 6줄** — `spec.md:105` 한 줄에 2회 등장하므로 줄 수와 건수가 1 차이 난다. 위치: `spec.md` 103·105(×2)·147·173, `acceptance.md` 113·118, `plan.md` 104.
- `progress.md` 는 개정 전에도 이 토큰이 0건이라 치환 대상이 없었다(아래 §4 참조).

## 3. 예산·구조 불변 확인 (실측)

```
$ grep -c '^- \*\*REQ-TA-' spec.md            → 15   (상한 16, 신설 0)
$ grep -c '^### AC-TA-' acceptance.md          → 16   (상한 16, 신설 0)
$ grep -c '^### Out of Scope — ' spec.md       → 6    (불변)
$ git diff --stat                              → 3 files, +11 −7
$ git status --short                           → M spec.md / M acceptance.md / M plan.md
```

- **REQ 신설 0 / AC 신설 0.** 지시된 금지 준수.
- **기존 AC 의 주장 내용 변경 0.** AC-TA-011 은 토큰 문자열만 바뀌었고 Given/When/Then 의 구조·개수·방향(② 역방향 `relate t1 t5`)은 그대로다. 나머지 15개 AC 는 무변경.
- REQ-TA-013 은 **조건절(술어) 무변경**, 결과절의 토큰 + 설명 꼬리절만 변경. GEARS State-driven 패턴 유지.
- `spec.md` frontmatter, `§F 이력`, `progress.md` **미변경**.

`moai spec lint` 전량 실행 결과는 §5 참조.

---

## 4. 범위 밖으로 남긴 것 (리드 판단 필요)

지시문의 "범위 밖은 한 글자도 건드리지 마세요"에 따라 **의도적으로 하지 않은** 것 3가지. 커밋 주체가 리드이므로 판단을 넘긴다.

1. **`spec.md` frontmatter `version` 미변경** — 현재 `"0.2.0"`, `updated: 2026-08-23` 그대로. 본문이 실질 변경됐으므로 `0.2.1` 이 맞지만, frontmatter 는 지정된 편집 대상 밖이고 PASS 0.920 이 붙은 내용이다.
2. **`§F 이력` 행 미추가** — 같은 이유. 추가한다면 `| 0.2.1 | 2026-08-23 | manager-spec | plan-audit F5 반영: §B.3 을 REQ 수준으로 하향, 표식 `unreviewed` → `machine-only` 개명. |` 형태가 될 것이다.
3. **`progress.md` §E.1 미변경** — 이 토큰이 0건이라 치환 대상이 아니었고, iteration 3 신호를 §E.1 에 추가하는 것은 지정 범위 밖이다. 현재 §E.1 은 iteration 2 상태(REQ 15 / AC 16)를 기술하며, 그 수치는 이번 개정 후에도 **여전히 정확**하다 — 다만 표식 이름이 §E.1 에 등장하지 않으므로 불일치도 없다.

리드가 1~3 중 하나라도 원하면 별도 지시를 주면 반영한다. 세 건 모두 내용이 아니라 메타데이터이므로, 커밋 메시지에서 대신 기록하는 선택도 가능하다.

---

## 5. 검증 (실행한 것 / 실행하지 않은 것)

**실행한 것**

| 확인 | 명령 | 결과 |
|---|---|---|
| F5 정본 확인 | `awk '/^### F5/,/^### F6/' .moai/reports/t119/plan-audit.md` | 지시문 요약과 일치, `spec.md:101/135/143` 지목 |
| 개정 전 토큰 위치 전수 | `grep -rn "unreviewed" .moai/specs/SPEC-TODO-ANALYSIS-001/` | 6건 / 4파일 중 3파일 |
| 개정 후 잔존 | `grep -ro "unreviewed" … \| wc -l` | **0** |
| 개정 후 신규 토큰 | `grep -ro "machine-only" … \| wc -l` | **7** (6줄) |
| REQ / AC / OoS 수 | `grep -c` 3종 | **15 / 16 / 6** — 전부 불변 |
| 변경 범위 | `git diff --stat`, `git status --short` | 3파일 `+11 −7`, `progress.md` 미포함 |
| SPEC lint | `moai spec lint` (전량) | 이 SPEC 관련 finding **0건**, 카탈로그 전체 **0 error** |

**실행하지 않은 것 (Gaps)**

- `go test` 를 한 건도 돌리지 않았다. plan-phase 이고 구현이 없다. 코드 관련 주장 없음.
- **재감사를 받지 않았다.** Tier M 상한 2회를 이미 소진했고, 지시에 따라 이 수정은 재감사 없이 간다. 따라서 이 개정이 F5 를 감사관 기준으로 닫았는지는 **미검증**이다 — 감사관이 요구한 두 조건("검토가 아니라 에이전트 출처 기록의 유무임을 그대로 적을 것", "표식 이름을 그 의미에 맞게 바꿀 것)을 문면상 둘 다 충족했다는 것이 내 판단이지, 감사관의 판정이 아니다.
- 커밋·푸시하지 않았다. `internal/`·`.claude/` 미변경.

**잔여 위험**

- `machine-only` 는 **이름과 술어를 일치시켰을 뿐 행위자 구멍 자체를 막지 않는다.** 아무 호출자나 `source: agent` 소견을 써서 표식을 끌 수 있다는 사실은 그대로이고, 이제 그것이 §B.3 에 명시돼 있다. 감사관이 제시한 전자(행위자 제약 REQ)를 택하지 않은 결과이며, 지시에 따른 선택이다.
- 이름 변경이 **PASS 0.920 이 붙은 내용을 건드린다.** 판정 근거였던 `unreviewed` 토큰은 이제 SPEC 어디에도 없다. AC-TA-011 의 주장 구조는 보존됐으므로 판정의 실질은 유지된다고 보지만, 이는 재감사로 확인된 것이 아니다.
- run-phase 구현자는 `machine-only` 를 **CLI 렌더 문자열**로 그대로 쓰게 된다. 이 이름이 실제 `list` 출력에서 어떻게 읽히는지는 구현 후 운영자 확인이 필요하다.

---

## 6. 변경 파일

| 파일 | 변경 |
|---|---|
| `.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` | §B.3 표 L2 주체 열(`:101`), §B.3 본문 3문단(`:103`,`:105`,`:107`), REQ-TA-013(`:147`), Out of Scope 불릿(`:173`) |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/acceptance.md` | AC-TA-011 제목(`:113`), Then 절(`:118`) — 토큰만 |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/plan.md` | M4 `runTodoList` 항목(`:104`) — 토큰만 |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/progress.md` | **미변경** (§4 참조) |
