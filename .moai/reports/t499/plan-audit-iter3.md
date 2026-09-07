# SPEC 감사 보고서: SPEC-CODEX-PARTIAL-WIRING-001

- 반복: **iter3 / 3** (리드가 Tier M 상한을 넘겨 승인한 최종 라운드)
- 판정: **PASS**
- 총점: **0.89** (Tier M 임계 0.80). 조화평균 0.896, 산술평균 0.90 — 둘 다 임계 상회
- 점수 추이: iter1 0.76(FAIL) → iter2 0.84(FAIL, 뮤턴트 관문) → **iter3 0.89(PASS)** — 회귀 없음, STOP 신호 미발화

## 0. 감사 조건 (측정 귀속)

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`, HEAD `77739dbcb`
- **`internal/`은 픽스처 트리 `ace1c5440`과 바이트 동일** — 실측 `git status --short -- internal/` 무출력, `git diff --stat ace1c5440 -- internal/` 무출력. 따라서 acceptance.md가 `ace1c5440`에 고정한 RED-now 관측은 이 HEAD에서 그대로 재현 가능하며, 나는 재현했다(§3)
- 감사 대상 아티팩트 해시(판정 시점 13:32): `spec.md` `1ed5b1c5…`, `acceptance.md` `5b7962ba…`, `plan.md` `14f987d1…`, `progress.md` `4abe0312…`
  > **주의 — 감사 중 아티팩트가 두 번 움직였다.** 12:5x 판독분(`acceptance.md` `7167847c…`, `spec.md` `7d1f449e…`)과 13:32 판독분이 다르다(저자 세션이 동시 편집 중). 아래 모든 인용은 **13:32 해시** 기준이며, 결함 D1은 그 해시에서 직접 재확인했다. 이후 착지분은 이 판정의 범위 밖이다
- Tier M 입력 계약대로 `spec.md` + `plan.md` + `acceptance.md`를 읽었다(`progress.md`는 D3 회귀 점검용으로 추가 판독). 저자의 추론 맥락은 M1 Context Isolation에 따라 무시했다 — 리드 메시지가 전한 "저자의 주장"은 **주장으로만** 취급하고 전부 소스에서 다시 쟀다
- 감사 방식: Claude 단독(`audit_model` 다중 백엔드 미요청)

## 1. Must-Pass 결과 (7/7 통과)

| # | 기준 | 판정 | 증거 |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | **PASS** | 실측 `grep -o 'REQ-CPW-[0-9]*' spec.md \| sort -u` → `001 002 003 004 005 006 007 008 009 010 011` — 연속 11개, 결번·중복·제로패딩 불일치 없음 |
| MP-2 | GEARS 준수(요구 계층) | **PASS** | `spec.md` §C 11행 전부 유형 라벨 + SHALL/SHALL NOT. Ubiquitous 3(001/005/008), Event-driven 4(002/003/011 + 006는 State-driven `…인 동안`), Unwanted 4(004/007/009/010). **AC의 Given-When-Then은 검증 계층이므로 여기서 감점하지 않았다**(M3 §Scope) — Group 4에서 별도 채점 |
| MP-3 | YAML frontmatter | **PASS** | 12필드 전수 확인(`id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags`), 타입 적합. snake_case 별칭 0. `phase: "v3.1.5 target"`은 금지된 생애주기 토큰(plan/run/sync/mx)이 아님. `tier: M` / `related_specs`는 선택 필드(후자는 코퍼스 302개 SPEC이 쓰는 확립된 관례 — 실측) |
| MP-4 | §22 언어 중립성 | **N/A(자동 통과)** | 단일 언어(Go) 범위 SPEC — 다중 언어 도구를 규정하지 않음 |
| MP-5 | D7 교차-SPEC 화해 | **PASS** | 참조 2건 실측: `SPEC-CODEX-WIRING-001` → `status: completed`, `SPEC-CODEX-SIDECAR-GUARD-001` → `status: completed`. retired/superseded/archived 없음 → BLOCKING 소견 0 |
| MP-6 | D8 크로스플랫폼 | **PASS(auto)** | 실측 `grep -c syscall` → 4개 아티팩트 전부 `0` |
| MP-7 | [NEEDS CLARIFICATION] 관문 | **PASS** | 실측 `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-PARTIAL-WIRING-001/` → 무출력 |

## 2. 축별 점수

| 축 | 점수 | 밴드 | 증거 |
|---|---|---|---|
| Clarity | 0.90 | 1.0–0.75 사이 | 등가 단언이 `check.Message != wantHalfWiredAbsentMessage` 한 줄로 명시돼 해석 여지가 없다(`acceptance.md` (d)). 감점 요인은 (b)"Message가 agent 정의의 존재를 진술"이 기계 판정 불가 문장으로 남은 것 — 잔여분이 (d′)에는 명시됐으나 (b)에는 명시되지 않았다(§5 R2) |
| Completeness | 0.85 | 0.75 밴드 상단 | 필수 절 전부 존재, `### Out of Scope —` H3 3개 각각 구체 불릿 보유(실측 `grep -c '^### Out of Scope'` → `3`). REQ 11 / AC 9 모두 Tier M 상한 16 이내. 감점은 **D1**(DoD가 뮤턴트를 7종으로 세고 M6′를 빠뜨림) |
| Testability | 0.85 | 0.75 밴드 상단 | 릴리스 게이트 5건의 RED-now를 내가 직접 재실행해 재현(§3). 뮤턴트 8종이 등가 단언의 공허화 경로(시험이 구현 상수를 import 하는 tautology 포함)를 닫는다(§4). 감점은 D1 + (d′) 세 토큰 열거의 잔여분 + AC-CPW-005의 `grep` 하위절이 이미 초록인데 표기되지 않은 점(§5 R4) |
| Traceability | 1.00 | 1.0 | 전체 ID 대조(축약 표기에 기대지 않음): AC→REQ 9행이 인용한 REQ가 전부 실재하고, REQ 11개가 전부 최소 1개 AC에 덮인다(001→AC1/2, 002→AC1, 003→AC2, 004→AC1/2, 005→AC5, 006→AC3/4, 007→AC6, 008→AC7, 009→AC2, 010→AC8, 011→AC2, + AC9는 전체). 고아 AC 0, 미덮 REQ 0 |

> **`moai spec lint`의 침묵은 어느 주장에도 쓰지 않았다.** 이 SPEC의 요구는 전부 표 행이고 lint의 REQ 수집기는 목록 항목만 잡는다 — 실측 `internal/spec/lint_req_widen.go:59` `reqLineWidePattern` = ``^\s*[-*]\s+…``. 따라서 REQ 기반 4규칙이 이 SPEC을 **한 번도 방문하지 않으며**, 수집된 REQ는 0건이다. 위 Traceability 1.00은 lint가 아니라 전체 ID 손대조로 얻은 값이다.

## 3. 리드가 물은 다섯 가지

### (1) 등가 단언은 잘 형성됐고 뮤턴트에 빈틈이 없는가 — **그렇다**

소스로 확인한 성립 근거:

- **부재 갈래는 평문 리터럴을 대입하는 자리다.** 현행 `internal/cli/doctor_codex.go:99-103`의 조기 반환이 `check.Message = "not wired (claude-only project) — skipped"`(`:101`)로 리터럴을 대입한다. 파일 관용구도 같다 — `:195` wired-OK도 평문 리터럴, 가변부(경로)는 Detail로 내려간다. 즉 등가 단언이 요구하는 "합성 없는 대입"이 이 갈래에 이미 존재하는 모양이다
- **합성 경로는 등가를 깨는 쪽으로만 작동한다.** `:200-201` — `problems`가 하나라도 차면 `Status = CheckWarn` + `Message = joinCodexSummaries(problems)`. 즉 M7(훑기 누출)은 (a)와 (d)를 동시에 깨뜨린다. 저자가 (a)와 (d)를 "같은 사실의 두 얼굴"이라 적은 것은 정확하다
- **간접 판독 경로 없음.** (d′)는 시험 파일 리터럴 자체를 검사하고, (d)는 런타임 Message를 그 리터럴과 비교한다. 어느 쪽도 구현 상수를 우회 참조하지 않는다

내가 시도한 반례 뮤턴트 4종 중 3종은 잡히고 1종은 잡히지 않는다:

| 시도한 뮤턴트 | 결과 |
|---|---|
| 시험이 자기 리터럴 대신 **구현 상수를 import**해 `check.Message == impl.halfWiredAbsentMessage`로 쓴다(등가가 tautology가 됨) | **잡힌다** — (d)의 코드 스니펫이 "시험 파일 쪽 리터럴"을 명시하고, 무엇보다 그 구현에서는 M6·M6′가 둘 다 GREEN이 되어 §D.3 [HARD] 절이 모든 AC의 PASS 기록을 금지한다. **M6/M6′의 진짜 역할이 이 tautology 차단이다** |
| 부재 갈래를 조기 반환 없이 본 경로로 흘려보낸다 | **잡힌다** — M7. `:200-201` 합성으로 (a)·(d)·(e) 동시 FAIL |
| 부재 갈래 Message를 200 rune로 늘린다 | **잡힌다** — `joinCodexSummaries`의 "선두 요약 하나가 상한을 넘으면 통째로 낸다"는 의도적 예외(`:222-224` 주석) 때문에 존재 갈래에서도 절단으로 숨지 않는다. AC-CPW-007이 두 갈래를 다 재므로 M5는 양쪽에서 RED |
| **지시문을 Message가 아니라 `check.Detail`에 싣는다** | **잡히지 않는다** — 아홉 기준 전부 만족(REQ-CPW-009는 Message만 구속). 다만 실측상 피해가 작다: Detail은 `--verbose`에서만 렌더된다(`internal/cli/doctor_render.go:134-139`, `if verbose { … if c.Detail != "" }`). §D-1이 막으려던 "평시 `moai doctor`에 상시 경고 행"은 발생하지 않으므로 **결함이 아니라 판단 사항**으로 기록한다(§5 R3). plan.md §B-3이 이 렌더 조건을 이미 정확히 알고 있다 |

### (2) 등가의 한계에 대한 저자의 정정은 정확하고 정직하게 기록됐는가 — **그렇다**

- 정확하다. 등가는 *구현이 리터럴에서 벗어나는 것*을 완전히 막지만 *리터럴로 무엇을 고르는가*는 구속하지 못한다. 저자가 리드의 "완전 차단" 읽기를 반박한 것이 옳다
- 정직하게 기록됐다. `acceptance.md` (d′) 하단 주석이 "구멍이 사라진 것이 아니라 검토 지점이 옮겨간 것"이라고 명시하고, 잔여분의 성립 조건("저자가 지시문을 리터럴로 고르고 시험 리터럴도 같게 적으면 (d)는 초록이다")까지 적는다. 과장 폐쇄 주장 없음
- 다만 (d) 본문의 "**원리상 하나도 통과하지 못한다**"는 단독으로 읽으면 과장이다. 바로 아래 (d′) 주석이 정확히 그 범위를 좁히므로 **결함으로 세지 않고 판단 사항으로 기록**한다(§5 R1의 문구 축)

### (3) M6′는 등가에서 RED이고 종전 금지 목록에서 GREEN인가 — **그렇다(전환 정당)**

M6′ 예시 문구 `codex agent definitions present, no wiring files — wire them once codex is available`를 세 축으로 대조:

- iter2 (d) 처리(부분문자열 `moai init` 금지): 이 문구에 `moai init` 없음 → **GREEN(통과)**
- 현행 (d′) 세 토큰(`moai init` / `run ` / `install`): 셋 다 없음 → **GREEN** — 즉 모양 검사만으로는 못 잡는다
- 현행 (d) 등가: 시험 리터럴과 다름 → **RED**

세 관측이 갈리므로 M6′는 실제 판별자다. 그리고 M6′가 (d′)까지 통과한다는 사실이 **왜 (d)가 필요한지**의 증명이다 — 전환의 근거는 성립한다.

### (4) D2의 `stubCodexHome` 고정이 등가를 잘 정의되게 했는가 — **그렇다. 다만 리드의 읽기는 절반만 맞다**

소스 대조:

- `resolveCodexHomeDir()`(`internal/cli/mcp_codex.go:1758-1767`)는 `CODEX_HOME` → `codexUserHomeDir()` 순으로 읽는다. `stubCodexHome`(`internal/cli/doctor_codex_test.go:76`)은 **둘 다** 고정한다(`t.Setenv(codexHomeEnvVar, "")` + seam 교체) — 헬퍼 주석이 그 이유를 적고 있다
- 리드의 읽기("고정하지 않으면 `codexStaleSkillFinding`이 Message에 붙어 등가가 깨진다")는 **누출 구현에 한해** 참이다. 올바른 구현에서는 이 갈래가 훑기에 도달하지 않으므로 home 내용과 무관하게 등가가 성립한다 — 저자의 0.3.2 정정이 맞고, 아티팩트가 그 강한 형태를 채택했다(하위 케이스 (i)/(ii) 둘 다에서 등가 요구)
- 고정이 여전히 필요한 이유는 둘이다: (a) 누출 구현이 **결정적으로** 실패하게 만들려면 개발자 머신의 실제 `~/.codex/config.toml`이 판정에 끼어들면 안 된다, (b) M7이 RED가 되려면 (ii) 낡은 home이 반드시 있어야 한다. 아티팩트가 이 둘을 모두 적고 있다
- 픽스처 실행 가능성 확인(읽음): `writeCodexHomeConfig`(`:95`)는 `filepath.Join(home, ".codex")`에 config를 쓰고 **home 루트**를 반환하며, `resolveCodexHomeDir`는 `home + codexHomeDirName`을 조립한다 — `stubCodexHome(t, writeCodexHomeConfig(t, …))` 합성이 그대로 성립한다. `absentSkillPath`(`:132`)도 Windows 안전 형태

### (5) 앞선 네 수리 영역 재검증 — **네 건 전부 유지**

| iter2 결함 | 재검증 결과 | 증거 |
|---|---|---|
| D3 — lint 무결 주장이 공허 | **RESOLVED** | `progress.md:34-38`이 인용을 철회하고 원인을 정확히 진단. 원인 진술을 소스로 확인: `internal/spec/lint_req_widen.go:59`의 `^\s*[-*]\s+` 앵커 — 표 행은 매치 0. `spec.md`의 요구 표기는 lint를 만족시키려고 비틀지 않았다(옳은 처분). **이 판정서도 lint 침묵을 어떤 근거로도 쓰지 않았다** |
| D4 — 좌표 오기 2건 | **RESOLVED** | 인용된 좌표를 전수 재측정(트리 `77739dbcb`, `internal/`은 `ace1c5440`과 동일): `:46 reTrustAdvice` ✓, `:52 initCodexAdvice` ✓, `:92 wired := hooksErr` ✓, `:101` unwired-skip 리터럴 ✓, `:189 codexStaleSkillFinding()` 호출 ✓, `:193-201` 합성 분기 ✓, `:195` wired-OK 리터럴 ✓, `:376 func codexStaleSkillFinding` ✓, 시험 `:76/:95/:132` ✓, `:390/:404` · `:437/:450` ✓(`grep -n 'checkCodexWiring(t.TempDir()'` → `155: 404: 450: 552: 567:`), `internal/codexwiring/wire.go:51-56` = `RefreshWiring` 본문 전체 ✓. **오기 0건** |
| D5 — RED-now 3필드 분해 | **RESOLVED** | 다섯 셀 전부 명령/stdout/`exit:` 분해. 명령이 단일 호출 형태이고 종료코드가 별도 필드 — §2.1 네 요소(명령·verbatim stdout·종료코드·트리 SHA) 충족(SHA는 문서 머리 고정) |
| D1·D2 — 뮤턴트 관문 | **RESOLVED** | (3)·(4) 참조. M6/M6′/M7가 세 축을 각각 잡는다 |
| D6(optional) — stdout 바이트 재현 | **RESOLVED** | 문서 머리에 "`(cached)`·시간 토큰은 바이트 비교 대상 아님" 인용 범위 절 신설 |

**RED-now 독립 재현(내가 실행함, 트리 `77739dbcb` / `internal/` = `ace1c5440`):**

```
grep -c "func TestCheckCodexWiring_HalfWiredCodexAbsent" internal/cli/doctor_codex_test.go   → 0
grep -c "func TestCheckCodexWiring_HalfWiredCodexInstalled" …                                 → 0
grep -c "func TestCheckCodexWiring_HalfWiredCountIndependent" …                               → 0
grep -c "func TestCheckCodexWiring_HalfWiredReadOnly" …                                       → 0
grep -c "func TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand" …                        → 0

go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexAbsent > /tmp/t499a.txt 2>&1
  exit: 0 · stdout: ok  github.com/modu-ai/moai-adk/internal/cli  (cached) [no tests to run]
go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand > /tmp/t499b.txt 2>&1
  exit: 0 · stdout: ok  github.com/modu-ai/moai-adk/internal/cli  (cached) [no tests to run]
```

§D.0 판독표대로 (grep `0` · `[no tests to run]`) = **정당한 RED-now**. 다섯 셀 전부 인용대로다.

## 4. 결함 (Defects Found)

**D1. DoD가 뮤턴트를 7종으로 세고 M6′를 빠뜨린다 — `acceptance.md`:224 — Severity: major — Class: blocking**

- 관측(해시 `5b7962ba…`, 13:32): `- [ ] AC-CPW-009 뮤턴트 **7종** 각각 RED 관측 … **M6·M7의 RED는 이번 라운드 수리의 합격 조건이므로 생략 불가**`
- 같은 문서 `:196`(§D.3 제목)은 **8종**, `:210` [HARD] 절은 **M6·M6′·M7**, `spec.md` §F는 **8종**, `plan.md` §E M3은 **8종(M1~M7, M6′ 포함)** — 즉 네 곳이 8을 말하고 DoD 한 곳만 7을 말한다
- **왜 blocking인가**: DoD는 run-phase 저자가 실제로 체크하며 일하는 목록이다. 이 줄대로 일하면 M6′는 심지 않아도 체크가 채워지고, 어떤 기계 검사도 그 누락을 보지 못한다. 그리고 **M6′는 이번 라운드 전환(금지 목록 → 등가 단언)의 유일한 증거 기반**이다 — 그것을 빠뜨린 DoD는 이 라운드의 수리가 실제로 작동하는지 확인하지 않은 채 끝난다. 이름만 들어간 수리를 금지한다는 §D.3의 [HARD] 절과 직접 모순된다
- **필요한 수정(한 줄)**: `뮤턴트 **7종**` → `뮤턴트 **8종**`, `**M6·M7의 RED는**` → `**M6·M6′·M7의 RED는**`

그 외 결함 없음. iter1·iter2에서 지적된 항목 중 미해결로 남은 것은 없다.

## 5. 판단 사항(결함 아님) — run-phase로 넘어가는 잔여 위험

**R1. 리터럴 선택 자체는 여전히 사람 판정이다.** 등가 단언은 구현이 리터럴에서 벗어나는 것만 막는다. (d′)의 세 토큰을 모두 피한 지시문을 리터럴로 고르면 아홉 기준 전부 초록이다. 아티팩트가 이 잔여를 명시하고 있으므로 결함이 아니다 — 다만 그 판단 지점이 **diff의 한 줄**로 옮겨간 것이 이 전환의 실질이므로, 검토자가 그 줄을 실제로 읽어야만 값이 실현된다.

**R2. (b)"Message가 agent 정의의 존재를 진술"도 같은 계열의 사람 판정인데, 잔여로 명시되지 않았다.** (d′)에는 잔여 주석이 붙었고 (b)에는 없다. 실질 위험은 R1과 동일하고 대응도 같다.

**R3. 부재 갈래의 `Detail`은 (c)·(e) 외에 구속되지 않는다.** 지시문을 Detail에 실으면 아홉 기준과 REQ-CPW-009(Message만 구속)를 모두 만족한다. 실측 완화: Detail은 `--verbose`에서만 렌더되므로(`internal/cli/doctor_render.go:134-139`) §D-1의 un-nagging 계약은 평시 `moai doctor`에서 깨지지 않는다. 요구를 추가할 만한 사안은 아니고, run-phase가 **의도적으로** 선택하고 그 선택을 적으면 된다.

**R4. AC-CPW-005의 `grep -nE '\b11\b'` 하위절은 오늘 이미 초록이다.** 실측: `grep -nE '\b11\b' internal/cli/doctor_codex.go` → 무출력(0건). 이 AC의 RED-now는 전적으로 "시험이 없다" 쪽이 지고 있고, grep 절은 회귀 가드가 릴리스 게이트 안에 얹혀 있는 형태다. 실제로 실패할 수 있으므로(구현자가 11을 박으면) 공허하지는 않다. 다만 표기가 없어, 시험만 초록이 된 시점에 "AC-CPW-005가 뒤집혔다"고 적힐 여지가 있다.

### run-phase DoD가 추가로 실어야 하는 것

1. **D1 수정 착지 확인** — 위 한 줄. 이것이 없으면 M6′ 누락이 아무에게도 보이지 않는다
2. **고른 리터럴을 `progress.md` §E.2에 verbatim 인용** — R1/R2의 사람 판정 지점을 증거 표면에 올려 검토자가 읽게 만든다(구현 상수 + 시험 리터럴이 **함께** 바뀐 diff와 같이)
3. **부재 갈래 `check.Detail`의 최종 내용도 함께 인용** — R3의 선택을 의도적인 것으로 기록
4. **AC-CPW-005 판정 시 두 절을 분리 보고** — 시험 3케이스(RED→GREEN)와 grep 가드(green-now 유지)를 따로 적는다

## 6. 권고

**PASS.** Must-pass 7/7, 총점 0.89(임계 0.80), 점수 회귀 없음(0.76 → 0.84 → 0.89). 요구 계층·검증 계층·추적성 모두 건전하고, iter1·iter2의 지적 열 건이 전부 해소됐다. 등가 단언으로의 전환은 M6′라는 실제 판별자로 정당화되며, tautology 경로(시험이 구현 상수를 참조)까지 뮤턴트 관문이 닫는다.

**단, D1은 run-phase 진입 전에 반드시 착지해야 한다.** 네 번째 감사 라운드는 필요 없다 — 결함이 한 줄이고, 수정이 옳은지는 diff에서 눈으로 확인되며 어떤 재측정도 요구하지 않는다. 부채로 넘기는 것도 반대한다: 이 한 줄이 곧 이번 라운드 수리의 검사 자체이므로, 넘기면 검사 없는 수리가 된다. 리드는 이 수정의 착지만 확인하고 Implementation Kickoff Approval로 진행하면 된다.

## 7. 인용 규율 기록

- 이 판정서의 모든 `file:line`은 트리 `77739dbcb`(= `internal/` 기준 `ace1c5440`)에서 **판정 직전에 재측정**했다. 이 카드에서 좌표 오기가 두 번 났으므로 전수 재측정했고, 오기 0건이다
- 종료코드는 `<cmd> > out.txt 2>&1; echo $?`로 읽었다(파이프 뒤 `$?` 금지)
- `moai spec lint`의 침묵은 어떤 주장의 근거로도 쓰지 않았다(수집된 REQ 0건 — §2 각주)
- 적용한 정책 규칙: `.claude/rules/moai/development/verification-completeness.md` §1.1(빈 스윕), §2(두 셀), §2 뮤턴트 프로브, §2.1(RED-now 네 요소), §4(SHA 고정); `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3·§2
