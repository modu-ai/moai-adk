# t82 M1 — 절 분류와 압축 가능성 파일럿

SPEC-AGENTS-MD-CANON-001 M1. 측정일 2026-08-22 · 워크트리 `.claude/worktrees/t82` ·
`WT-agents-md-diet` @ `58f0bdd43`.

M1 은 측정 파일럿이다. `AGENTS.md` 를 쓰지 않고, 룰 텍스트를 옮기지 않는다. 산출물은 투영값과
go/no-go 판정이다.

## 판정 — GO, 단 조건부

| 정지 조건 | 기준 | 투영 | 판정 |
|---|---:|---:|---|
| Arm A — 계약이 천장에 맞는가 | 24,576 B | **11,881 B** | 통과 (천장의 48 %) |
| Arm B — 다이어트가 래칫 천장에 닿는가 | 66,371 tok | 필요 감축 **7,806 tok** vs 확보 가능 **10,670 tok** | 통과 (여유 +2,864) |

**조건**: Arm B 의 통과는 `AGENTS.md` 가 **측정 투영값(약 11.9 KB) 근처로 착지할 때**만 성립한다.
천장(24,576 B) 까지 부풀면 필요 감축이 10,980 tok 으로 올라가고, 한 번도 스텁 분할된 적 없는
파일 9개만으로는 **310 tok 부족**하다. 그 경우 M4 는 이미 스텁 분할된 5개 파일 중 하나 이상에
2차 패스를 추가해야 한다(그 파일들의 R3 잔량 68,115 B — 5 %만 회수해도 851 tok 이라 메우기는
쉽다). 이 조건을 M2 에 넘긴다.

출력 스타일(`.claude/output-styles/moai/moai.md`) 2차 패스는 **필요 없다** — `spec.md` §E.2 가
M1 결과로 결정하라고 한 사안이며, 위 두 경로 모두 출력 스타일을 건드리지 않고 성립한다.

---

## 1. 절 블록 확장 — 라인 프록시 정정

`AC-AMC-003` 은 라인 단위 grep 을 거부하고 마커를 **절 블록**으로 확장할 것을 요구한다.
확장 규칙과 구현: `.moai/reports/t82/clause-blocks.py` (연속 규칙·펜스 블록·제목 부착 마커 처리를
docstring 에 명시).

```
python3 .moai/reports/t82/clause-blocks.py > .moai/reports/t82/blocks.json
python3 .moai/reports/t82/summarize.py
```

| 항목 | 라인 프록시 (`spec.md` §A.4) | 절 블록 실측 |
|---|---:|---:|
| 마커 수 (룰 14개 + `CLAUDE.md`) | 97 | 97 |
| 바이트 | 32,543 | **51,639** |

절 블록 확장은 프록시를 **+58.7 %** 끌어올린다. `spec.md` §A.4 가 "16개가 `:` 로 끝나 본문이
전혀 계수되지 않는다, 그 크기에 상한이 없다"고 경고한 그 미계수분이 실제로 이 크기였다.

파일별:

| 파일 | 절 블록 수 | 바이트 |
|---|---:|---:|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | 23 | 11,188 |
| `.claude/rules/moai/workflow/session-handoff.md` | 11 | 9,890 |
| `.claude/rules/moai/core/moai-constitution.md` | 10 | 7,809 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 15 | 5,613 |
| `.claude/rules/moai/core/askuser-protocol.md` | 9 | 3,586 |
| `.claude/rules/moai/workflow/skill-routing.md` | 4 | 2,870 |
| `.claude/rules/moai/workflow/context-window-management.md` | 5 | 2,408 |
| `CLAUDE.md` | 4 | 2,190 |
| `.claude/rules/moai/workflow/cache-aware-execution.md` | 5 | 1,872 |
| `.claude/rules/moai/workflow/cross-session-messaging.md` | 4 | 1,553 |
| `.claude/rules/moai/workflow/main-checkout-branch-guard.md` | 2 | 979 |
| `.claude/rules/moai/core/native-idiom-and-register.md` | 2 | 848 |
| `.claude/rules/moai/core/verification-claim-integrity.md` | 3 | 833 |

always-loaded 룰 14개 중 **13개만** `[HARD]` 마커를 갖는다. `moai-mcp-tools.md` (7,357 B),
`goal-directive.md` (6,600 B) 는 0개다 — 즉 두 파일은 전량이 R3 이고 다이어트 대상으로만 존재한다.

### 1.1 프록시의 과다 계수 주장은 실제보다 크다

`spec.md` §A.4 는 "93개 중 15개가 절이 아니라 산문 언급"이라고 적었다. 절 블록을 **읽고** 판정한
결과는 다르다:

- 구조적으로 절 첫머리가 아닌 마커는 **10개**다(제목 부착 6개를 절로 재분류한 뒤).
- 그중 실제로 의무를 담지 않는 순수 항해 주석은 **1개**뿐이다 —
  `kanban-dispatch.md:7` 의 detail companion 안내문("스텁은 모든 [HARD] 룰과 포인터를 유지한다").
- 나머지 9개는 문장 중간에 마커가 놓였을 뿐 **진짜 의무**다(`cache-aware-execution` 지시 6-10,
  `session-handoff` Block 1 / Block 5 필드 규격, `kanban-dispatch` 유휴 통지 조항,
  `CLAUDE.md` §14 백그라운드·동시성 조항).

그리고 `moai-constitution.md` 의 Agent Core Behaviors 6개는 **제목 줄에 마커가 붙어 있어**
라인 프록시에서는 제목 한 줄(약 50 B)로만 계수됐다. 이들의 본문 전체가 의무이며, 절 블록으로
확장하면 6개 합계 5,117 B 다. 프록시의 과소 계수가 가장 심한 지점이다.

**정정**: 프록시의 오차는 §A.4 가 서술한 "과다 15 / 과소 16" 이 아니라, 사실상 **과다 1 / 과소 96**
방향이다. 과다 계수 우려는 실측으로 거의 소멸했고, 과소 계수만 남는다.

---

## 2. Codex 관련 / Claude 전용 분류

축: 그 절이 **Codex 가 갖지 않은 메커니즘**을 구속하는가. Claude 전용으로 분류해도 어느 쪽
하네스의 구속 표면도 줄지 않는다 — 그 절은 Claude 쪽 always-loaded 룰에 그대로 남는다
(`design.md` §1.1).

분류는 파일 단위가 아니라 **절 단위**다. `design.md` §1.1 의 6-파일 상한(14,360 B)은 파일 소속을
프록시로 쓴 값이고, 이번 실측은 그 6개 파일 안의 개별 절을 각각 판정했다.

명령:

```
python3 .moai/reports/t82/classify_report.py            # 소계
python3 .moai/reports/t82/classify_report.py --table    # 절별 표 (97행, 미분류 0)
```

| 클래스 | 블록 | 바이트 |
|---|---:|---:|
| **C — Codex 관련** | **35** | **16,135** |
| K — Claude 전용 | 61 | 35,147 |
| P — 산문 언급(의무 아님) | 1 | 357 |
| 합계 | 97 | 51,639 |

파일별:

| 파일 | C 블록 | C 바이트 | K 블록 | K 바이트 | P 바이트 |
|---|---:|---:|---:|---:|---:|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | 10 | 4,932 | 12 | 5,899 | 357 |
| `.claude/rules/moai/workflow/session-handoff.md` | 0 | 0 | 11 | 9,890 | 0 |
| `.claude/rules/moai/core/moai-constitution.md` | 7 | 5,905 | 3 | 1,904 | 0 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 9 | 1,995 | 6 | 3,618 | 0 |
| `.claude/rules/moai/core/askuser-protocol.md` | 0 | 0 | 9 | 3,586 | 0 |
| `.claude/rules/moai/workflow/skill-routing.md` | 0 | 0 | 4 | 2,870 | 0 |
| `.claude/rules/moai/workflow/context-window-management.md` | 0 | 0 | 5 | 2,408 | 0 |
| `CLAUDE.md` | 1 | 393 | 3 | 1,797 | 0 |
| `.claude/rules/moai/workflow/cache-aware-execution.md` | 2 | 675 | 3 | 1,197 | 0 |
| `.claude/rules/moai/workflow/cross-session-messaging.md` | 0 | 0 | 4 | 1,553 | 0 |
| `.claude/rules/moai/workflow/main-checkout-branch-guard.md` | 2 | 979 | 0 | 0 | 0 |
| `.claude/rules/moai/core/native-idiom-and-register.md` | 1 | 423 | 1 | 425 | 0 |
| `.claude/rules/moai/core/verification-claim-integrity.md` | 3 | 833 | 0 | 0 | 0 |

### 2.1 6-파일 상한과의 대조

`design.md` §1.1 이 상한으로 지목한 6개 파일의 절 블록 합계는 22,179 B 이고, 그중 Claude 전용은
**21,504 B (97.0 %)**, Codex 관련은 675 B(`cache-aware-execution` 지시 7·8) 뿐이다. 파일 소속
프록시는 이 6개에 한해 거의 정확했다 — 상한을 거의 전부 실제로 쓸 수 있다.

그러나 Claude 전용 총량의 **38.8 %(13,643 B)** 는 그 6개 **밖**에 있다
(`kanban-dispatch` 5,899, `agent-common-protocol` 3,618, `moai-constitution` 1,904 …). `kanban-dispatch.md` 가 특히 그렇다: 칸반 리드/레인 역할·디스패치·`/clear` 인계는 Claude
Code 세션 메커니즘이고, 워크트리 규율·검증 부하·증거 판독은 하네스 일반 원칙이라 같은 파일 안에서
갈린다. **파일 단위 분류였다면 어느 쪽으로 잡아도 틀렸을 지점**이다.

### 2.2 경계 판정 — 보수적으로 C

다음은 메커니즘이 반반이라 판정이 갈릴 수 있는 절이다. 전부 **C 로 잡았다** — 계약을 부풀리는
방향이므로 Arm A 에 대해 안전한 오차다.

| id | 절 | 판정 근거 |
|---|---|---|
| B066 | 워크트리는 런처로 진입 | 금지 조항(`git worktree add` 금지)은 일반적이나 형식 표는 Claude 도구(`EnterWorktree`) |
| B069 | 새 카드는 새 워크트리 | 같음 |
| B074 | env 스크럽 검증 형식 | 근거가 Claude Code 가드지만 형식 자체는 리포 전반 규율 |
| B010 | 읽기 전용 검증 배치 | 단일 턴 다중 Bash 는 Claude 형식이나 "직렬화하지 말라"는 일반 원칙 |
| B004 | XML 은 에이전트 간 전용 | 사용자 대면 절반은 일반, 에이전트 간 절반은 Claude |

반대로 **K 로 잡았지만 일반 원칙이 섞인** 절도 있다. B050(피어에게 이 세션이 못 하는 일을
시키지 말 것 — 권한 세탁 금지)과 B063(리드는 읽은 증거로 카드를 전진시킨다)이 그렇다. 둘 다
일반 원칙 형태는 이미 B036(검증 주장 무결성)·B032(범위 규율)로 계약에 들어가므로 중복 배제했다.

---

## 3. 압축 파일럿

`plan.md` §E M1 이 지정한 두 파일 — `kanban-dispatch.md`(마커 23개, 분포의 상단)와
`native-idiom-and-register.md`(2개, 하단). `AGENTS.md` 에 들어갈 절만 재작성했다(C 11개, 5,355 B);
Claude 전용 절은 `AGENTS.md` 에 들어가지 않으므로 잴 압축이 없다.

재작성본: `.moai/reports/t82/pilot-compressed.md` · 계측: `.moai/reports/t82/pilot_measure.py`

| id | 출처 | 이전 B | 이후 B | 비율 |
|---|---|---:|---:|---:|
| B034 | `native-idiom-and-register.md`:8 | 423 | 257 | 0.608 |
| B064 | `kanban-dispatch.md`:98 | 679 | 423 | 0.623 |
| B066 | `kanban-dispatch.md`:127 | 676 | 335 | 0.496 |
| B067 | `kanban-dispatch.md`:139 | 454 | 208 | 0.458 |
| B068 | `kanban-dispatch.md`:141 | 246 | 179 | 0.728 |
| B069 | `kanban-dispatch.md`:143 | 693 | 349 | 0.504 |
| B070 | `kanban-dispatch.md`:145 | 550 | 224 | 0.407 |
| B071 | `kanban-dispatch.md`:159 | 513 | 243 | 0.474 |
| B072 | `kanban-dispatch.md`:171 | 488 | 195 | 0.400 |
| B073 | `kanban-dispatch.md`:173 | 379 | 257 | 0.678 |
| B074 | `kanban-dispatch.md`:179 | 254 | 178 | 0.701 |
| **합계** | | **5,355** | **2,848** | **0.532** |

**집계 압축비 0.5318 — 46.8 % 감축.** 절별 분산: 최소 0.400, 최대 0.728, 평균 0.552,
중앙값 0.504, 표준편차 0.119 (`AC-AMC-005`).

압축은 전부 **의무 밖 텍스트 제거**다: 인라인 근거, 사고 기록 참조, 상호참조 포인터, 예시 표.
비율이 높은 쪽(B068 0.728, B074 0.701, B073 0.678)은 원문이 이미 의무 한 문장에 가까워 깎을 것이
없던 절이고, 낮은 쪽(B072 0.400, B070 0.407)은 근거 서술이 의무와 같은 줄에 실려 있던 절이다.

### 3.1 `AC-AMC-004` — 주어 / 양태 / 범위 보존

| id | 주어 | 양태 | 범위 | 판정 |
|---|---|---|---|---|
| B034 | 모든 사용자 대면 표면 | MUST read / prohibited | `conversation_language ≠ en` | 보존 |
| B064 | `gh pr checks` 의 CodeRabbit 행 | 두 조건 BOTH 성립 시에만 계수 | PR 리뷰 증거 판독 | 보존 |
| B066 | 카드 작업 / 워크트리 | never bare `git worktree add` | 카드 워크트리 생애주기 | 보존 |
| B067 | `moai worktree done` | L2 전용 (closes … only) | L1/L2 폐기 경로 | 보존 |
| B068 | 워크트리 폐기 | Dispose no worktree … until | 미푸시 브랜치 | 보존 |
| B069 | 새 카드 | MUST exit first / never reuse | 카드 전환 | 보존 |
| B070 | 카드 워크트리 브랜치 | carry `WT-` … never the card id | 브랜치 명명 | 보존 |
| B071 | 세 운반자 | mandatory | 추적성 | 보존 |
| B072 | 레인 검증 | scoped to the change | 로컬 검증 범위 | 보존 |
| B073 | 백그라운드 부하 | Never spawn | 검증 레시피 | 보존 |
| B074 | env 스크럽 검증 | one compound invocation | 워크트리 내부 | 보존 |

폐기(revert)한 절은 없다. 다만 세 건에서 **의도적으로 범위를 좁히지 않기 위해** 원문 인접
텍스트를 끌어오지 않았다: B070 의 슬러그 형태 표(토큰 3개·24자 상한)와 B074 의 `env -u` 거부
근거, B064 의 `Review rate limited` 처리는 모두 해당 `[HARD]` 절 **밖**에 있어 제외했다. M2 가
계약에 이들을 포함하려면 별도 판단이 필요하다 — 지금 넣으면 범위 확대(REQ-AMC-003 위반)다.

---

## 4. Arm A — 천장 재도출과 투영

`AC-AMC-006` 은 라인 프록시가 아니라 **절 블록 수치**로 천장을 재도출할 것을 요구한다.

계측: `.moai/reports/t82/project.py`

| 단계 | 바이트 | 근거 |
|---|---:|---|
| 절 블록 `[HARD]` 총량 | 51,639 | `clause-blocks.py` 실측 |
| 차감: Claude 전용 61블록 | −35,147 | 절별 분류 |
| 차감: 산문 언급 1블록 | −357 | 의무 아님 |
| = Codex 관련 원문 (35블록) | **16,135** | |
| × 측정 압축비 0.5318 | 8,581 | 파일럿 11절 |
| + 문서 구조 | +3,300 | **가정** — 아래 참조 |
| = 투영 `AGENTS.md` | **11,881** | |
| 천장 (REQ-AMC-004) | 24,576 | 미변경 |
| 예산 대비 여유 | 20,886 | 하한 8,192 |

**문서 구조 3,300 B 는 측정치가 아니라 명시된 가정이다.** 산정: 절 35개를 약 14개 절(section)로
묶고, 절당 제목 한 줄(~40 B) + 항해용 서술 1-2문장(~150 B) = ~190 B, 여기에 문서 서두
(제목·자기충족성 선언·전역 `~/.codex/AGENTS.md` 경고 포인터) ~600 B. 추적표는 `AGENTS.md` 가
아니라 `progress.md` 산출물이므로(`plan.md` M2) 포함하지 않았다.

### 4.1 민감도 — 분류가 크게 틀려도 Arm A 는 버틴다

- **압축비 0 (원문 그대로)**: 16,135 + 3,300 = 19,435 B — 여전히 천장 아래.
- **손익분기**: 천장을 채우려면 Codex 관련 원문이 **40,004 B** 여야 한다. 실측 16,135 B 의
  **2.5배**. 내 분류가 절반쯤 틀려 Claude 전용을 대거 C 로 옮겨도 천장은 유지된다.
- 반대로 **분류를 하지 않으면 불가능**하다: 절 블록 전량 51,639 B 는 압축 후 27,464 B, 구조를
  더하면 30,764 B 로 천장(24,576)뿐 아니라 예산(32,768)에도 2,004 B 밖에 남기지 못한다.
  분류는 선택이 아니라 필수 지렛대다.

**천장 자체는 24,576 B 로 유지할 것을 권고한다.** 투영이 천장의 48 % 에 불과해 낮출 여지가
있지만, 천장을 조이는 것은 SPEC 변경이고 M1 의 소관 밖이다. 다만 §5 가 보이듯 **Arm B 는
`AGENTS.md` 의 실제 크기에 민감**하므로, M2 는 천장이 아니라 투영값 근처를 목표로 삼아야 한다.

---

## 5. Arm B — 다이어트가 래칫 천장에 닿는가

### 5.1 기준선

리드 지시에 따라 **71,212 가 아니라** 가드 재현 실측을 쓴다. `surface_r3.py` 는
`token_budget_guard.go` 의 `alwaysLoadedSurface()` + `measureAlwaysLoaded()` 를 그대로 재현한다
(파일별 `len/4` 내림 합산, `MEMORY.md` head 캡, frontmatter 한정 `paths:` 판정):

```
python3 .moai/reports/t82/surface_r3.py
# surface files: 17   guard tokens (sum of per-file len/4): 71207
# surface bytes: 284850
```

**71,207 tok / 284,850 B.** pre-flight 의 71,212 와 5 tok 차이는 파일별 내림(`len/4`) 대 총합
나눗셈(`total//4`) 차이이며, 가드가 실제로 계산하는 값은 71,207 이다. 어느 쪽을 쓰든 결론은
바뀌지 않는다(아래 두 경우 모두 여유가 5 tok 보다 훨씬 크다).

`AGENTS.md` 는 아직 열거되지 않으므로(REQ-AMC-013 ¶2 가 M5 에 확장을 요구) 이 값은 **계약층
이전** 표면이다. 계약층은 순증이다.

### 5.2 필요 감축

| 경우 | `AGENTS.md` | 표면+계약 | 필요 감축 |
|---|---:|---:|---:|
| 측정 투영 | 11,881 B = 2,970 tok | 74,177 | **7,806 tok** (31,224 B) |
| REQ-AMC-004 천장 | 24,576 B = 6,144 tok | 77,351 | **10,980 tok** (43,920 B) |

천장 경우의 10,980 은 pre-flight 가 제시한 10,985 와 5 tok 차(기준선 차이 그대로)다.

### 5.3 확보 가능량 — 실측 선례로 가격 매김

이동 가능한 R3(룰 14개 + `CLAUDE.md` 에서 R1 절 블록과 R2 라인을 뺀 잔량):
**161,482 B = 40,370 tok** (`surface_r3.py`).

필요 감축은 그 R3 의 **19.3 %**(측정 투영) / **27.2 %**(천장)다.

스텁+지연 companion 패턴의 **실측 선례** 두 건:

| 파일 | 이전 | 이후 (always-loaded) | 감축 | 커밋 |
|---|---:|---:|---:|---|
| `goal-directive.md` | 25,755 | 6,531 | **74.6 %** | `6422046bb` |
| `kanban-dispatch.md` | 21,003 | 13,027 | **38.0 %** | `a203a7c3a` |

보수적으로 낮은 쪽(38.0 %)을 쓴다.

**가장 조이는 경계** — 한 번도 스텁 분할된 적 없는 always-loaded 파일 9개에만 전체파일 38 % 를
적용한다(이미 companion 이 있는 5개 파일의 2차 패스는 이 경계에서 **제외**).

| 파일 | 바이트 |
|---|---:|
| `CLAUDE.md` | 20,523 |
| `moai-constitution.md` | 18,958 |
| `cross-session-messaging.md` | 16,672 |
| `verification-claim-integrity.md` | 13,140 |
| `context-window-management.md` | 13,009 |
| `main-checkout-branch-guard.md` | 11,865 |
| `moai-mcp-tools.md` | 7,357 |
| `skill-routing.md` | 5,825 |
| `native-idiom-and-register.md` | 4,967 |
| **합계** | **112,316** |

38.0 % → 42,680 B = **10,670 tok**.

| 필요 감축 | 확보 가능 | 판정 |
|---|---:|---|
| 7,806 tok (측정 투영) | 10,670 | **통과, 여유 +2,864** |
| 10,980 tok (천장) | 10,670 | 부족 −310 |

천장 경우의 −310 은 이미 스텁된 5개 파일(R3 잔량 68,115 B)에서 **5 %만 회수해도 851 tok** 이라
쉽게 메워진다. 그래서 Arm B 는 통과이되, §판정의 조건이 붙는다.

### 5.4 이 투영이 무엇이 아닌지

- 파일럿의 0.532 는 **절 압축비**이지 R3 이전 비율이 아니다. Arm B 는 절 압축비를 쓰지 않고,
  실측 스텁 분할 선례(38 %)를 R3 위에 얹는다. 두 지렛대는 별개다.
- 38 % 를 **전체파일** 비율로 재고 **한 번도 손대지 않은 파일에만** 적용한 것은 두 번 보수적이다:
  R3 는 전체파일보다 작으므로 R3 에 38 % 를 적용하면 전체파일 38 % 보다 적게 나오고,
  이미 스텁된 파일의 잔여 여지를 0 으로 뒀다.
- 그럼에도 이것은 **투영이지 측정이 아니다**. 실제 이전은 M4 가 파일별로 재야 한다.

---

## 6. 검증

```
go test -count=1 ./internal/config/ -run 'Budget|AlwaysLoaded' -v
```

```
--- PASS: TestAlwaysLoadedTokenBudget (0.00s)
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.00s)
    --- PASS: .../over-budget (0.00s)
    --- PASS: .../under-budget (0.00s)
--- PASS: TestAlwaysLoadedSurfaceEnumeration (0.01s)
--- PASS: TestMeasureAlwaysLoaded_WithMemory (0.00s)
--- PASS: TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.637s
```

Go 코드는 변경하지 않았다. `alwaysLoadedSurface()` 확장(M5)과 `t.Logf`(M5)는 손대지 않았고,
`AlwaysLoadedTokenBudget` 은 76,000 그대로다.

---

## 7. 미검증 (Gaps)

1. **문서 구조 3,300 B 는 가정이다.** M2 가 실제로 재야 한다. 3,300 이 두 배로 틀려도 Arm A 는
   통과하지만(6,600 + 8,581 = 15,181 < 24,576), Arm B 의 필요 감축은 825 tok 늘어난다.
2. **Arm B 의 38 % 는 이전 비율의 실측이 아니라 선례의 이식이다.** 두 개의 데이터 포인트뿐이며,
   두 파일 모두 이번 대상 9개 파일과 다른 문서다.
3. **R2 측정은 라인 단위 조악 추정이다** (10,023 B, 룰+`CLAUDE.md`). `MUST`/`shall` 을 포함한
   줄을 통째로 세므로 서술 문장을 과다 계수하고, 다중 줄 의무 본문은 과소 계수한다.
   `design.md` §1 의 11,095 과 같은 계열의 프록시다. R3 가용량 161,482 B 는 이 오차를 안고 있다.
4. **v3.2 통합 브랜치가 아직 없다.** `AC-AMC-018` 은 통합 브랜치 수치를 요구하고, pre-flight 는
   오늘 시점 live ref 4개가 모두 71,212(가드 기준 71,207)로 동일함을 확인했을 뿐이다. 형제 카드가
   룰을 추가하면 기준선이 올라가고 필요 감축도 같은 폭으로 올라간다 — 76,000 상향이 정확히 그렇게
   발생했다.
5. **`char/4` 추정.** 가드와 같은 식이라 가드에 대해서는 정합하지만 실제 tokenizer 와 ±15 % 오차.
   Arm A 는 바이트 기준이라 무관하고, Arm B 만 이 오차 안에 있다.
6. **분류는 판단이다.** 97개 중 경계 판정 7건을 §2.2 에 명시했다. `AC-AMC-003` 의 미분류 0 은
   충족했으나(97/97), 분류 자체의 정합성은 M2 의 추적표에서 다시 검증된다.

## 8. 잔여 위험 (Residual risk)

- **재팽창.** `kanban-dispatch.md` 는 2026-08-17 에 13,027 B 로 다이어트됐고 오늘 **25,915 B** 다
  — 5일 만에 다이어트 이전(21,003)보다 커졌다. 다이어트는 되돌아가며, 이번 SPEC 의 래칫이
  그 되돌림을 잡는 유일한 장치다. Arm B 의 여유 2,864 tok 은 재팽창 한 번이면 사라질 크기다.
- **`AGENTS.md` 크기 압력.** M2 가 천장(24,576)을 목표로 삼으면 Arm B 가 −310 으로 뒤집힌다.
  천장은 상한이지 목표가 아니라는 `plan.md` §A 의 서술이 여기서 실질적 구속이 된다.
- **분류 드리프트.** 이후 룰 편집이 새 `[HARD]` 절을 추가하면 어느 쪽으로 분류할지 판정이 필요하고,
  그 판정 없이는 `AGENTS.md` 가 조용히 계약을 잃는다. M5 의 바이트 가드는 크기만 보지 **완전성**은
  보지 않는다.
- **전역 `~/.codex/AGENTS.md`.** 예산 32,768 중 사용자 파일이 먼저 소비한다. 투영 11,881 이면
  약 20,886 B 가 남아 웬만한 개인 문서를 흡수하지만, 이는 리포가 기계적으로 볼 수 없는 슬라이스다
  (`spec.md` §D.3).

---

## 9. 산출물

| 파일 | 역할 |
|---|---|
| `.moai/reports/t82/clause-blocks.py` | 마커 → 절 블록 확장 (가드 표면 재현 포함) |
| `.moai/reports/t82/blocks.json` | 절 블록 97개 (파일·줄 범위·바이트·본문) |
| `.moai/reports/t82/summarize.py` | 절 블록 파일별 집계 |
| `.moai/reports/t82/dump.py` | 절 블록 사람 판독용 덤프 |
| `.moai/reports/t82/classification.tsv` | 절별 C/K/P 판정 + 근거 (97행) |
| `.moai/reports/t82/classify_report.py` | 분류 소계 + `--table` 절별 표 |
| `.moai/reports/t82/pilot-compressed.md` | 파일럿 재작성본 11절 |
| `.moai/reports/t82/pilot_measure.py` | 파일럿 이전/이후 바이트 + 분산 |
| `.moai/reports/t82/surface_r3.py` | 가드 표면 재현 + 파일별 R1/R2/R3 |
| `.moai/reports/t82/project.py` | Arm A / Arm B 투영 |

절별 분류표(97행) 원문은 `python3 .moai/reports/t82/classify_report.py --table` 로 재생성한다.
