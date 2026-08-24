# t238 — SPEC 문서 리뷰 대응 (D1-D8)

카드 t238. PR #1606 (SPEC-CODEX-SESSION-MSG-001) CodeRabbit 지적 7건 + 확장 1건에 대한 문서 수정.
5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

## Claim

D1-D8 전부 수정 완료. 커밋하지 않았다(오케스트레이터 몫).

변경 파일 5개 (전부 문서):

| 파일 | 관련 findings |
|---|---|
| `.moai/specs/SPEC-CODEX-SESSION-MSG-001/acceptance.md` | D2, D3, D4, D5, D6, D7 |
| `.moai/specs/SPEC-CODEX-SESSION-MSG-001/progress.md` | D1, D2, D3, D6, D7 |
| `.moai/specs/SPEC-CODEX-SESSION-MSG-001/plan.md` | D6 |
| `.moai/specs/SPEC-CODEX-SESSION-MSG-001/design.md` | D8 |
| `.moai/reports/t187/verdict.md` | D1 |

`spec.md`는 **무변경**(`git diff -- spec.md` = 0행) — frontmatter `status`/`version` 미변경.
`internal/**` **무변경** — 작업 시작 시점의 `git status` 목록과 종료 시점 목록이 동일(형제 레인 소관 10개 파일 + 신규 3개, 내가 만든 변경 0).

가장 중요한 발견: **AC-CSM-008 / AC-CSM-010 / AC-CSM-012 / 그리고 progress.md의 `-run` 명령이 문서에 실린 형태 그대로는 아무것도 관측하지 않았다**(vacuous). CodeRabbit의 지적은 절반만 맞았고 처방 방향은 반대였다 — 상세는 D2/D7.

## Evidence

측정은 전부 이 트리에서, 아래 Baseline-attribution의 HEAD 기준으로 실행했다.

### D1 — 로컬 환경 값 커밋됨 (수정)

누출 인벤토리 실측(수정 전):

```
$ grep -noE '/Users/goos[^ "`)]*' <5개 문서>
.moai/reports/t187/verdict.md:95:/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187
.moai/specs/.../progress.md:132:/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187
.moai/specs/.../progress.md:140:/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187

$ grep -noE '"pid":[0-9]+|goos\.local' .moai/specs/.../progress.md
140:"pid":27305
140:goos.local
```

수정 후:

```
$ grep -n "Users/goos\|27305\|goos.local" .moai/specs/.../progress.md
rc=1   (매치 없음)
$ grep -n "worktrees/t187" .moai/reports/t187/verdict.md
3:Card: t187 · Branch: `WT-codex-session-msg` · Worktree: `.claude/worktrees/t187`
95:(`.claude/worktrees/t187`), after the rebase onto
```

- 절대경로 → 저장소 상대경로 `.claude/worktrees/t187`
- `"pid":27305` → `"pid":<pid>`, `"host":"goos.local"` → `"host":"<host>"`
- progress.md 140행은 **축어 e2e 증거 블록** 안이다. 조용한 편집은 기록을 거짓으로 만들므로, 블록 바로 위에 편집 고지를 붙였다: 어떤 세 값을 자리표시자로 바꿨는지 명시하고 그 외 바이트는 원본임을 단언. 증거 무결성은 값을 남겨서가 아니라 **치환을 공개해서** 지킨다.

**`~/go/bin/moai`는 의도적으로 유지**(acceptance.md:29 ×2, progress.md:130). 홈 상대경로라 계정명이 없고, Go 표준 설치 경로이며, 해당 절차 지시문(rm+cp 재설치 규율)이 그 경로를 필요로 한다.

### D2 + D7 — 이스케이프된 파이프 (CodeRabbit 처방 방향이 반대였음)

CodeRabbit: "파이프라인 구분자의 백슬래시를 제거하고, grep 패턴 안의 `\|`는 정규식 교대이므로 유지하라."

**측정 결과 두 지시가 모두 반대다.** GFM은 표 셀 안(인라인 코드 포함)에서 `\|`를 `|`로 해제한다 — 파이프라인 구분자와 grep 패턴 **양쪽 모두**. 따라서:

- 파이프라인 구분자는 이미 동작하는 `|`로 렌더된다 → 백슬래시를 지우면 표가 깨진다(고칠 것이 없다).
- grep 패턴의 교대는 **깨져 있다** — BRE `grep`이 리터럴 `|`를 받는다. CodeRabbit이 "유지하라"고 한 쪽이 정확히 망가진 쪽이다.

vacuous 실증 — 매치가 **67행 존재하는** 디렉터리에 렌더된 형태를 실행:

```
$ grep -rn "exec.Command|codex-jobs|app-server|net.Listen|http.Listen" internal/cli/ | grep -v _test | wc -l
0
$ grep -rEn "exec.Command|codex-jobs|app-server|net.Listen|http.Listen" internal/cli/ | grep -v _test | wc -l
67
```

AC-CSM-012도 같은 형태로 깨져 있었다(이쪽은 vacuous PASS가 아니라 **거짓 FAIL**):

```
$ grep -c "재시작|restart" .claude/rules/moai/core/moai-mcp-tools.md   →  0   (AC는 ≥1 요구 → 거짓 실패)
$ grep -cE "재시작|restart" .claude/rules/moai/core/moai-mcp-tools.md  →  1
```

`-run` 패턴(Go RE2)은 방향이 반대다 — RE2에서 `\|`는 **리터럴 파이프**라 렌더된 `|` 쪽이 옳다. 두 형태를 실행:

```
$ go test ./internal/cli/ -run 'TestSessionMsg|TestMoaiMCPServer_RegistrationMatchesCatalog' -v | grep -c '^--- PASS'
9

$ go test ./internal/cli/ -run 'TestSessionMsg\|TestMoaiMCPServer_RegistrationMatchesCatalog' -v | tail -3
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.256s [no tests to run]
```

이스케이프된 형태는 **0개 테스트를 고르고 exit 0으로 통과**한다 — vacuous의 최악 형태. 다만 표 소스의 `\|`는 렌더 시 `|`가 되므로 `-run` 행은 **소스 수정이 불필요**했다(이미 옳다). 확인만 하고 두었다.

적용한 수정 — **표 이스케이프(`\|`)는 유지하고 플래그에 `-E`를 추가**해서, 렌더된 명령이 그대로 실행 가능하게:

| 위치 | 수정 |
|---|---|
| acceptance.md AC-CSM-008 | `grep -rn` → `grep -rEn` + 이유 각주 |
| acceptance.md AC-CSM-012 | `grep -c` → `grep -cE` + 이유 각주 |
| progress.md M1 E4 경계 grep (59행) | `grep -rn` → `grep -rEn` |
| progress.md M1 AC-CSM-008 (62행) | `grep -rn` → `grep -rEn` |
| progress.md M2 AC-CSM-008 (86행) | `grep -rn` → `grep -rEn` |
| progress.md M3 AC-CSM-012 (114행) | `grep -c` → `grep -cE` |
| progress.md M2 배터리 `-run` (95행) | 소스 무변경(이미 옳음) — 선택 건수 주석만 추가 |

**변환 후 렌더 형태 전량 실행 — 기록된 결과와 전부 일치, 새 불일치 0:**

```
AC-CSM-008  grep -rEn ... internal/sessionmsg/ internal/cli/mcp_session_msg.go | grep -v _test   → 0행 (기록: 0행)
AC-CSM-010  grep -rn 'os.Getenv("' ... | grep -v _test                                          → 0행 (기록: 0행)
AC-CSM-010  grep -cE '^[[:space:]]*DefaultSessionMsg[A-Za-z]*[[:space:]]*=' defaults.go          → 6   (신규 기록: 6)
AC-CSM-012  grep -c "session_msg" internal/mcp/catalog.go                                        → 4   (기록: 4)
AC-CSM-012  grep -cE "재시작|restart" 본품 / 미러                                                → 1 / 1 (기록: 1/1)
AC-CSM-013  grep -c "session-msg-e2e" progress.md                                                → 3   (기록: ≥1)
M1 E4       grep -rEn 'AskUserQuestion|mcp__askuser' internal/sessionmsg/ | grep -v _test        → 0행 (기록: 0행)
AC-CSM-015  go test ./internal/cli/ -run TestSessionMsgToolsRegisteredWithHintsAndDiscipline -v  → --- PASS (0.01s)
```

**중요 — 기록 자체는 거짓이 아니었다.** 셸에서 큰따옴표 안의 `\|`는 grep에 BRE 교대로 전달되어 정상 동작한다(대조 측정: `grep -rn "exec.Command\|..." internal/cli/` → 67). 즉 실행자가 당시 얻은 증거는 유효했고, 깨진 것은 **문서에 렌더되어 독자가 복사·실행하는 형태**뿐이다. 그래서 관측값은 하나도 바뀌지 않았다.

`-run` 건수만 예외: 기록된 `8`은 M2 커밋 `26e248ce6` 기준으로 정확하다. 현재 트리는 형제 레인이 추가한 `TestSessionMsgPollHandlerRejectsTraversalIDs`(HEAD에 없음 — `git grep` 확인) 때문에 **9**를 고른다. 기록된 8을 고치지 않고, 선택 집합만 늘었다는 주석을 붙였다.

### D3 — AC-CSM-015 검증 명령이 너무 약함 (수정)

기존: `grep -c "a reply is not user approval" internal/cli/mcp_session_msg.go` ≥ 1.

이 grep이 겨누는 파일에는 **도구 설명이 아예 없다**:

```
$ grep -n "a reply is not user approval" internal/cli/mcp_session_msg.go
28:const sessionMsgDisciplineShortForm = "... a reply is not user approval."

$ grep -n "Discipline" internal/cli/mcp_server.go
403: mcp.WithDescription("Register this session ... "+sessionMsgDisciplineShortForm),
413: mcp.WithDescription("List registered ... "+sessionMsgDisciplineShortForm),
421: mcp.WithDescription("Send a short message ... "+sessionMsgDisciplineShortForm),
436: mcp.WithDescription("Claim pending messages ... "+sessionMsgDisciplineShortForm),
```

발행되는 설명 4개는 전부 `mcp_server.go`에 있고 상수를 이어붙인다. 따라서 그 grep의 값은 `mcp_server.go`를 어떻게 바꾸든 **1로 불변** — 발행 설명을 한 번도 관측하지 않는다.

교체: `go test ./internal/cli/ -run TestSessionMsgToolsRegisteredWithHintsAndDiscipline -v`. 테스트 본문 확인 결과 ListTools 실측으로 `session_msg_` 접두 도구를 순회하며 **각각의 `tool.Description`**에 토큰 포함을 단언하고, 마지막에 4개 도구가 전부 발견됐는지 확인한다. 실행 → `--- PASS`.

§D 매트릭스, §F AC-CSM-015 시나리오, progress.md M2 매트릭스 3곳 모두 교체. base-0 실측 각주는 유지.

### D4 — §G 추적성 서술이 거짓 (수정)

"나머지 13 요구는 각자 전용 AC"는 §D 매트릭스와 맞지 않는다. 매트릭스에서 직접 재도출:

- **공유 AC**: AC-CSM-003 ← REQ-005·006 / AC-CSM-008 ← REQ-010 + REQ-001의 "포트 없음" 조항 / AC-CSM-012 ← REQ-013·015
- **복수 AC**: REQ-013 → AC-007 + AC-012 / REQ-014 → AC-011 + AC-015
- **전용 1:1은 8건**: REQ-002→001, 003→002, 004→014, 007→004, 008→005, 009→006, 011→009, 012→010
- AC-CSM-013은 요구가 아닌 카드 납품 목표 — 커버리지 계산에서 제외

리드 요약에는 REQ-013(AC-007+AC-012)과 REQ-015가 빠져 있었다. 위 도출이 매트릭스 전량 대조 결과다. 공유·복수를 모두 허용하는 서술로 교체.

### D5 — §G Definition of Done 체크리스트 (수정)

progress.md에 **인용된 증거(명령+관측 출력)가 있는 항목만** 체크했다.

**체크 8건** — 근거:

| 항목 | progress.md 근거 |
|---|---|
| REQ↔AC 매핑 | §D 매트릭스(D4에서 정정·재도출) — 문서 내부 속성이라 명령 증거가 아닌 매트릭스 자체가 근거 |
| sessionmsg + cli 테스트 초록 | §E.2 M1 E3(패키지 전체 `ok ... 0.484s`), M2 배터리(`8 test 전부 --- PASS`) |
| 커버리지 ≥85% | §E.2 M1 E3 `coverage: 86.9% of statements` |
| lint 0건 | §E.2 M1 E5 / M2 E5 `0 issues.` |
| windows 빌드 exit 0 | §E.2 M1 E2 `E2_FINAL_OK` / M2 E2 `E2_M2_BUILDS_OK` |
| 가드·경계·하드코딩 grep 0건 | §E.2 M1 E4, M2 AC-008·009·010 행 |
| 미러 + make build | §E.2 M3 `catalog.yaml updated successfully`, 미러 diff 0행 |
| e2e 왕복 + 위생 | §E.2 M4 3-7단계 축어, 9단계 `HYGIENE_UNCHANGED` |

**미체크 2건** — 각 항목에 사유 괄호를 달았다:

- **MX 태그**: progress.md에 MX 검증 명령·출력이 **인용돼 있지 않다**. 태그 자체는 소스에 존재하나(`grep -rn "@MX:" internal/sessionmsg/ internal/cli/mcp_session_msg.go` → 10행, `store.go`/`agent.go`에 fan_in≥3 `@MX:ANCHOR`) 이 체크리스트의 근거 규칙은 progress.md 인용이므로 체크하지 않았다.
- **5절 보고 형식**: 마일스톤별 보고는 리드에게 갔고 progress.md에 그 형식의 보고문이 없다(`grep -n "5절\|Residual" progress.md` → 매치 0). 인용 가능한 증거 부재.

### D6 — plan.md M1 임계값 5 → 6 (수정)

```
$ grep -nE '^[[:space:]]*DefaultSessionMsg[A-Za-z]*[[:space:]]*=' internal/config/defaults.go
389:	DefaultSessionMsgMessageTTL = 24 * time.Hour
392:	DefaultSessionMsgClaimTTL = 10 * time.Minute
395:	DefaultSessionMsgAgentOfflineMinutes = 30
398:	DefaultSessionMsgPollBatch = 16
401:	DefaultSessionMsgMaxTextBytes = 65536
406:	DefaultSessionMsgMaxParts = 8
```

plan.md에 여섯 번째(`DefaultSessionMsgMaxParts` 8, REQ-CSM-005 부분 수 상한) 추가.

의존 계수 정합: AC-CSM-010의 "임계값 5종" → "6종 **선언**"으로 고치고, 검증 명령을 무앵커 `grep -c "DefaultSessionMsg" ≥ 5`에서 **선언 행만 세는 앵커 정규식 = 6**으로 강화. progress.md M1 매트릭스에도 선언 6 / 무앵커 12를 함께 기록.

### D8 — design.md role 열거형 주석 부정확 (수정)

proto3 JSON은 열거형을 **이름**으로 직렬화하므로 A2A v1 ProtoJSON의 `role`은 `"ROLE_USER"`/`"ROLE_AGENT"`다. 소문자 단축형은 그 열거형의 "JSON 형태"가 아니다.

오케스트레이터 판정대로 **온디스크 값은 유지**하고 이탈을 명시했다. Go 상수 **무변경**.

- §3.2 주석: 소문자 단축형이 브로커의 온디스크 값이며 A2A ProtoJSON과 다르다는 점, 상호운용 시 명시적 변환 필요를 서술
- §3.1 정합 표: `role` 행에 "**값은 의도적 이탈**" 표기 + 표 아래 각주 신설 — 필드명·카디널리티는 정합이나 값은 비정합, 경계 변환(`"user" ↔ "ROLE_USER"`) 필요, 값 변경은 온디스크 계약 변경이라 범위 밖
- 기계적 가드로 형제 레인의 단언을 인용: `internal/sessionmsg/envelope_test.go:93-94` `TestEnvelopeA2AAlignment` 하위 `role serializes as the lowercase short form`이 `"user"`/`"agent"` 직렬 값을 고정

### t241 규칙 — 변경한 AC의 변이 실증

| AC | 변이 | 옛 형태 | 새 형태 |
|---|---|---|---|
| AC-CSM-008 | `mcp_session_msg.go` 사본에 금지 토큰(`exec.Command`) 1줄 주입 | `0` → **통과(vacuous)** | `1` → **실패** ✓ |
| AC-CSM-010 | `defaults.go` 사본에서 선언 1개 삭제 | 무앵커 `11` ≥ 5 → **통과** | 앵커 `5` ≠ 6 → **실패** ✓ |
| AC-CSM-012 | 문서 사본에서 재시작/restart 토큰 제거 | (렌더 형태는 원본에서도 0 — 거짓 실패) | `0` < 1 → **실패** ✓ |
| AC-CSM-015 | `mcp_server.go`의 한 `WithDescription`에서 `+sessionMsgDisciplineShortForm` 제거 | `mcp_session_msg.go`의 값은 상수 1건으로 **불변 → 통과** | ListTools 순회 단언 → **실패**(미실행, Gaps 참조) |

앞의 셋은 스크래치 사본으로 실제 실행해 위 수치를 관측했다.

## Baseline-attribution

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187` (`pwd` + `git rev-parse --show-toplevel` 일치 확인)
- 브랜치: `WT-codex-session-msg`
- HEAD: `f33cd05649f27f6ba0c44db95505c3e303283a52`
- 작업 트리 상태: 형제 레인의 **미커밋** 코드 변경 10개 파일 + 신규 3개 포함. 따라서 `internal/**` 대상 측정값은 HEAD가 아니라 **HEAD + 형제 레인 미커밋 변경** 기준이다 — 특히 `-run` 9건, `TestSessionMsgPollHandlerRejectsTraversalIDs` PASS가 그렇다.
- grep 구현: `ugrep 7.8.4 aarch64-apple-macosx` (`grep --version`). BRE `\|` 교대를 지원함을 대조 측정(67행)으로 확인 — 위 D2 결론은 "BRE가 `\|`를 못 읽는다"가 아니라 "**렌더 결과인 리터럴 `|`를 BRE가 교대로 읽지 않는다**"이다.
- 모든 명령은 이 실행에서, 이 트리에서 직접 실행했다. 인용 출력은 축어.

## Gaps (관측하지 않은 것)

- **AC-CSM-015 변이를 실행하지 않았다.** 실행하려면 `internal/cli/mcp_server.go`를 편집해야 하는데 `internal/**`는 범위 밖이다. 변이가 옛 grep을 통과한다는 것은 **구조적으로** 증명했다(토큰이 `mcp_session_msg.go`에 상수 1건으로만 존재 → 그 파일의 계수는 `mcp_server.go` 변경에 불변). 새 테스트가 그 변이를 잡는다는 것은 테스트 본문의 도구별 순회 단언을 읽어 확인했을 뿐, 실행 관측은 아니다.
- **`go test ./...` 미실행**(리포 규율). 실행한 것은 `internal/cli/`의 두 `-run` 필터뿐. 전 패키지 판정은 CI 몫.
- **GFM 렌더를 내가 직접 재측정하지 않았다.** `gh api markdown -f mode=gfm` 결과는 오케스트레이터 측정을 인용했다. 다만 그 결론에 의존하는 부분(`\|` → `|`)은 두 방향 모두 실행으로 뒷받침된다 — 리터럴 `|` BRE는 0을 반환하고 `-E`는 67을 반환한다.
- **커버리지·lint·windows 빌드를 재측정하지 않았다.** D5 체크는 progress.md 인용 증거를 근거로 했을 뿐, 내가 다시 돌려 확인한 값이 아니다.
- **미러(`internal/template/templates/`) 문서는 손대지 않았다.** SPEC 산출물과 리포트는 미러 대상이 아니다. 다만 AC-CSM-011/012가 미러 파일을 grep하므로, 미러 쪽 `moai-mcp-tools.md`의 `-cE` 값(1)은 확인했다.
- **`git status` 기준선을 작업 전에만 캡처했다.** 형제 레인이 작업 중 파일을 더 건드렸다면 내 "internal 무변경" 주장은 목록 동일성으로만 뒷받침된다(내용 diff 대조는 하지 않았다).

## Residual-risk

- **축어 블록 편집이라는 선례.** progress.md 140행 치환은 고지를 달았지만, 축어 증거를 사후 편집하는 행위 자체가 선례가 된다. 이번 건은 비식별화라는 좁은 사유이고 고지로 감사 가능성을 유지했으나, 같은 논리가 "불편한 값"으로 확장되면 기록이 무의미해진다.
- **`-run` 선택 건수는 계속 표류한다.** 형제 레인 테스트가 늘 때마다 같은 명령이 다른 수를 고른다. 주석으로 완화했을 뿐 구조적 해결은 아니다.
- **다른 SPEC에도 같은 결함이 있을 가능성이 높다.** 표 셀 안 `grep "a\|b"`는 이 리포의 SPEC 문서 관행이다. 이 카드는 SPEC-CODEX-SESSION-MSG-001만 고쳤다 — 카탈로그 전역 스윕은 별도 카드 소관이다.
- **AC-CSM-010 앵커 정규식은 `defaults.go`의 `var (...)` 블록 들여쓰기에 의존한다.** 선언이 다른 형태(예: 한 줄 `var X, Y = ...`)로 바뀌면 계수가 어긋난다. 현재 6개 선언이 전부 동일 형태임은 확인했다.
- **AC-CSM-012의 `재시작|restart` 계수는 여전히 단어 존재만 본다.** 재시작 문장이 다른 맥락으로 바뀌어도 통과한다 — `-E` 수정은 "동작하지 않던 것을 동작하게" 했을 뿐 판정 강도를 올린 것은 아니다.
- **커밋하지 않았다.** 트리는 형제 레인의 코드 변경과 섞여 있으므로, 커밋 시 경로를 명시해 스테이징해야 한다(`git add` 대상: 위 5개 문서만).
