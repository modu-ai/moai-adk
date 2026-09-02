# Sync-Audit Verdict (iter-2, delta only): SPEC-BACKLOG-JSON-DISCLOSURE-001

카드: t395 · 트리 `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `41a3a500b`
iter-1 감사 대상: `e8a30fa10` · 이번 델타: `e8a30fa10..41a3a500b` (커밋 1개, 코드 0줄)
측정 시점: 2026-09-02 · 플랫폼: darwin/arm64
iter-1 보고서 `.moai/reports/t395/sync-audit-verdict.md` — 덮어쓰지 않음

**Overall Verdict: PASS**
**Overall Score: 91.2 / 100** — iter-1 89.9 → **+1.3** (단조 증가, 점수 퇴행 없음)

**D1 · D4 는 실제로 닫혔다 — 옮겨진 것이 아니다.** 거짓 전칭 문장이 두 곳 모두에서 사라졌고,
그 자리를 함수별 실측 열거가 대신한다. 다만 그 교체 문장 안에서 **세 번째 불일치**를 찾았다(D8).
리드가 "세 번째를 있을 법한 쪽으로 취급하라"고 한 그대로였다. 한 단어짜리 오산이며 공개
산출물(CHANGELOG)이 아니라 내부 `progress.md` 에만 있다.

---

## 트리 확인 — HEAD 는 핀대로, 그러나 divergence 가 움직였다

`git rev-parse HEAD` → `41a3a500b47f4ab150a3494391bf04e9ff01a381` — 지시받은 핀과 일치.
작업트리 clean, 델타 커밋 1개.

**보고 의무 사항**: `git rev-list --count --left-right origin/develop...HEAD` 가
iter-1 의 `0 11` 에서 **`56 12`** 로 바뀌었다. 브랜치 내용이 변한 게 아니라 **origin/develop 이
56 커밋 전진**했다. SPEC 결함은 아니지만 병합 창 판단에 걸리는 사실이라 적는다:

- 더 이상 fast-forward 가 아니다.
- 내가 잰 커버리지·린트는 그 56 커밋을 **포함하지 않는** 트리의 값이다.
- 그리고 아래 §"세 번째 불일치 해소"가 보이듯, **develop 병합이 `internal/cli` 커버리지 분모를
  바꾼 전례가 이미 이 카드 안에 있다.** 다시 병합하면 cli 수치는 또 움직일 수 있다.

델타가 코드를 건드리지 않았음을 확인: `git diff --stat e8a30fa10..HEAD -- internal/ .claude/`
→ 출력 없음. 따라서 iter-1 에서 채점한 Functionality / Security 는 재측정 대상이 아니다.

---

## Dimension Scores

| Dimension | iter-1 | iter-2 | 근거 |
|---|---|---|---|
| Functionality (40%) | 92 | **92** | 코드 델타 0 — 재측정 안 함(지시대로). iter-1 실측 유지 |
| Security (25%) | 95 | **95** | 코드 델타 0 — 동일 |
| Craft (20%) | 85 | **85** | 코드 델타 0. D2/D3(AC-BJD-016 자동 검증 부재)은 수용된 부채로 커밋 메시지에 명시됨 — 재론 안 함 |
| Consistency (15%) | 82 | **91** | D1·D4 닫힘. 감점 잔여는 D8(내부 문서 오산) 하나뿐 |

가중합: 92·0.4 + 95·0.25 + 85·0.2 + 91·0.15 = 36.8 + 23.75 + 17 + 13.65 = **91.2**
Must-pass firewall(Functionality + Security): 통과. blocking 결함 없음 → **PASS**.

---

## D1 — 닫혔는가, 옮겨졌을 뿐인가: **닫혔다**

### 사라진 것

거짓 전칭문 `every changed function at or above 86.7%` 이 **두 곳 모두에서** 제거됐다.
`grep -rn 'at or above 86' CHANGELOG.md .moai/specs/SPEC-BACKLOG-JSON-DISCLOSURE-001/` → 0건.
문구를 부드럽게 다듬은 게 아니라 명제 자체가 없어졌고, 그 자리에 **함수별 실측 열거**가 들어왔다.

### 들어온 것 — 전 항목 이번 실행에서 재측정

| 산출물이 주장하는 값 | 내 실측 | 일치 |
|---|---|---|
| `todo_disclosure.go` 100% | `discloseNonAuthoritativeBacklogJSON` 100.0% · `discloseQueueLayout` 100.0% | ✓ |
| `InspectBacklogArchiveVouch` 100% | **100.0%** (이번에 처음 직접 측정 — iter-1 미측정분) | ✓ |
| `runTodoList` 93.2% | 93.2% | ✓ |
| `runTodoPR` 88.1% | 88.1% | ✓ |
| `newTodoWhyCmd` 86.7% | 86.7% | ✓ |
| `runTodoHistory` 72.7% "the floor" | 72.7% — 변경 함수 중 최솟값 맞음 | ✓ |
| kanban 86.5% / template 86.3% | 86.5% (2회 확인) / 86.3% | ✓ |

즉 공개 산출물의 정량 주장은 이제 **한 건도 반박되지 않는다**.

### 리드가 따로 확인을 요청한 것 — §E.2 Gaps 의 추론 철회

옛 문장이 기대던 추론 `so the change cannot plausibly account for the shortfall` 은 제거됐다.
새 문장은 그 추론을 되살리지 않고 뒤집는다 — `the changed code cannot be used to argue the
shortfall away`. 게다가 "위에 있는 함수들에 대해서조차 그것은 측정이 아니라 추론"이라는
단서를 **일반화해서** 유지했다. 올바른 철회다. 변경 함수 하나가 패키지 수치 아래에 앉은 이상
그 추론은 실제로 성립 불가이고, 산출물이 그것을 인정한다.

---

## D4 — 닫혔는가: **닫혔다**

추가된 절: "`moai todo --help` is a tool-path reader whose `Long` text still names `backlog.json`
as the queue, and that string ships inside the binary. It is deliberately out of scope here and
carded separately."

이 HEAD 에서 두 절 모두 재확인:

- **help 텍스트**: `internal/cli/todo.go:96` `Long:` 이 여전히
  `Operate the kanban backlog queue at .moai/state/todo/backlog.json.` 로 시작하고,
  `:101` 이 홈폴백을 `~/.moai/todo/<project-key>/backlog.json` 로 단언한다. 절의 서술과 일치하며,
  오히려 **한 건 적게** 말한다(사이트가 둘인데 절은 하나로 뭉뚱그림 — 과장이 아니라 축소이므로 결함 아님).
- **바이너리 탑재**: iter-1 에서 바이너리 내 `state/todo/backlog.json` 잔존 2건 중 1건이 이
  `Long` 문자열임을 확인했다(나머지 1건은 `todo-queue-storage.md` 의 export 통제행).

한계 문단의 **다른** 주장들이 이 편집으로 깨졌는지도 봤다 — 안 깨졌다. `cat` 독자 절, 두 개
측정된 사각 열거, 미러 파리티 "scoped, not whole-file", Windows 미측정 범위(008/009/010,
011 제외 — iter-1 에서 확인한 대로 정확), `todo-queue-storage.md` byte-identical 주장 모두
문구 그대로 남았고 새 절은 그 사이에 삽입됐을 뿐이다.

---

## 세 번째 불일치(80.2 vs 80.3) — 귀속으로 남기지 않고 **해소했다**

리드 질문: 귀속만 하고 해결하지 않은 것이 잘못된 판단인가, 비례적 비용으로 확정할 수 있으면 하라.

**답: 재실행 없이 확정했다. 두 수치 모두 옳고, 서로 다른 트리를 잰 값이다.**

```
$ git diff --stat d7813ec2b 205f91f67 -- internal/cli/
 internal/cli/memory.go            |   2 +-
 internal/cli/memory_drain.go      | 178 +++++++++++++++++++++++++++
 internal/cli/memory_drain_test.go | 245 ++++++++++++++++++++++++++++++++++++++
```

run-phase 배치는 `4a4bbe396` 위(= 작업 내용 `d7813ec2b`)에서 쟀다. 그 **뒤에** 병합 커밋
`205f91f67` 이 develop 을 들여오면서 `internal/cli` 에 **production 코드 178행**
(`memory_drain.go`, SPEC-AGENT-MEMORY-DRAIN-001)이 추가됐다. 커버리지 분모가 달라진 것이다.
내가 잰 80.3% 는 **병합 이후** 트리의 값이고, 서로 다른 두 패키지를 잰 두 수치이므로
어느 쪽도 틀리지 않았다. Go 커버리지는 트리가 고정되면 결정적이며, 나는 80.3% 를 독립된 두
실행에서 얻었다.

**판단**: 재실행 없이 한쪽을 고르지 않은 것은 **옳았다** — 그때 골랐다면 다른 트리의 수치를
같은 트리의 오차로 오진했을 것이다. 다만 이제 원인이 밝혀졌으니 현재 문구는 한 곳이
부정확하다: `run-scoped rather than exact` 는 **측정 잡음/비결정성**을 암시하는데, 실제 원인은
**분모가 다른 트리**다. 값은 트리별로 정확하다.

**리드에게 주는 숫자**: 병합 대상 트리(`41a3a500b`)의 `internal/cli` 커버리지는 **80.3%** 다.
권장 문구(선택): "cli 80.2% — run-phase 측정값이며 `internal/cli/memory_drain.go` 를 들여온
develop 병합 이전 트리 기준. 병합 후 트리에서는 80.3%." 이러면 `not exact` 라는 오해가
사라지고 두 수치가 모두 살아난다.

---

## Findings

### D8 — [MINOR] [optional] §E.2 Gaps 의 대역 개수가 바로 위 표와 어긋난다 (세 번째 불일치)

`progress.md` §E.2 Gaps, D1 교체 문장:

> the new file is at 100% and **four of the changed functions sit between 86.7% and 93.2%**,
> but `runTodoHistory` is at 72.7%

이번 실행에서 잰 변경 함수 전량:

| 함수 | 값 | [86.7, 93.2] 대역? |
|---|---|---|
| `discloseNonAuthoritativeBacklogJSON` (신규 파일) | 100.0% | — (문장이 "the new file"로 별도 처리) |
| `discloseQueueLayout` (신규 파일) | 100.0% | — (동일) |
| `InspectBacklogArchiveVouch` | **100.0%** | **아니오 — 대역 위** |
| `runTodoList` | 93.2% | 예 |
| `runTodoPR` | 88.1% | 예 |
| `newTodoWhyCmd` | 86.7% | 예 |
| `runTodoHistory` | 72.7% | 아니오 — 문장이 따로 지목 |

대역 안은 **3개**지 4개가 아니다. 네 번째 후보인 `InspectBacklogArchiveVouch` 는 100.0% 라
대역 밖이다 — 이것이 내가 이번 라운드에 그 함수를 직접 잰 이유다. 값이 90%대였다면 "four" 가
맞았겠지만, 100.0% 로 실측됐다.

(참고로 `archiveTablesPresent` 가 86.7% 라 세면 4가 되지만, 그 함수는 이 SPEC 이 **바꾸지
않았다** — diff 에서 컨텍스트 줄로만 등장한다. 그렇게 세면 "changed functions" 쪽이 틀려진다.)

**성격**: 같은 문서 안 두 줄 위 커버리지 행이 반박하는 개수다. 다만 (a) 공개 산출물이 아닌
내부 `progress.md` 이고, (b) CHANGELOG 의 대응 문장은 개수를 세지 않고 **함수를 이름과 값으로
열거**하므로 이 오산을 갖지 않으며, (c) 문장의 하중을 지는 부분 — `runTodoHistory` 72.7% 가
패키지 수치 아래라는 것 — 은 정확하다. 결론의 방향이 뒤집히지 않는다.

**Required fix**: `four` → `three`. 혹은 개수를 빼고 CHANGELOG 처럼 열거로 바꾼다.

**FAIL 로 올리지 않은 이유**: finding-consumption 규율상 blocking 은 정확성이나 SPEC 이 실제로
말한 요구에 걸리는 것을 뜻한다. 이건 내부 문서의 한 단어 오산이고 어떤 결론도 뒤집지 않으므로
optional 이다. 다만 리드가 "자기 자신과 논쟁하는 문서를 내보내느니 창을 붙들겠다"고 했으므로
그대로 옮긴다 — **한 단어 수정이며, 넣을지는 리드 판단**이다. 내가 이걸로 FAIL 을 만들면
그건 이 규율이 막으려는 과잉 교정이 된다.

### 새로 만들어진 모순 — 없음

리드 지시대로 metrics 영역을 적대적으로 훑었다. CHANGELOG 쪽 정량 주장은 위 표대로 **전부
실측과 일치**하며 서로 어긋나는 문장을 찾지 못했다. §E.4 도 확인 — 델타가 건드리지 않았고
`b12_self_test_*` 는 CHANGELOG 항목의 **존재/중복/경로**를 검사하지 커버리지 수치를 검사하지
않으므로, 이번 편집으로 무효가 된 self-test 는 없다. §D.6 traceability 표, §D.5 severity 표,
AC 매트릭스도 이 델타의 손이 닿지 않았다.

---

## Gaps — 이번 라운드에 관측하지 **않은** 것

- **지시에 따라 재측정하지 않음**: AC-BJD-008 falsifiability, AC-BJD-010 WAL Given,
  AC-BJD-007 범위, 미수리 트리 대비 AC-BJD-012/015, stdout byte-identity, lint. iter-1 실측이
  그대로 유효하며 코드 델타가 0이라 결과가 달라질 경로가 없다.
- **`GOOS=windows GOARCH=amd64 go build ./...`**: metrics 영역에 "windows/amd64 builds green"
  주장이 있으나 iter-1·iter-2 모두 내가 재지 않았다. 델타가 문구 편집뿐이라 이 주장이 깨질 경로는
  없지만, **검증된 바 없음**을 명시해 둔다.
- **56 커밋 앞선 origin/develop**: 그 트리에서는 아무것도 재지 않았다. 병합 후 수치(특히 cli
  커버리지)는 다시 움직일 수 있다 — 이 카드 안에서 이미 한 번 그렇게 됐다.
- **D2/D3/D5/D6/D7**: 수용된 부채로 재론하지 않았다.

## Residual risk

- D8 을 두면 `progress.md` 가 두 줄 간격으로 자기 표와 어긋난 채 남는다. 결론은 안전하지만,
  이 카드에서 같은 영역의 오류가 **세 번 연속** 나왔다는 사실 자체가 신호다 — 셋 다 "세어서 요약한
  문장"에서 났고, 실측 열거로 바꾼 CHANGELOG 쪽에서는 한 번도 안 났다. 요약하지 말고 열거하는
  편이 이 영역에서는 더 안전하다.
- develop 이 56 커밋 앞서 있어, 병합 트리에서의 재측정 없이 지금 수치를 병합 후 근거로
  재사용하면 CLAUDE.local.md §4.1 규율 5("병합 전 검증을 병합 후 근거로 재사용하지 않는다")에
  걸린다.

---

## 감사 위생

- 커밋하지 않았다. 작업트리 clean, HEAD `41a3a500b` 불변.
- 이 보고서는 `sync-audit-verdict-iter2.md` 신규 파일이며 iter-1 보고서를 덮어쓰지 않았다.
- 임시 커버리지 프로파일은 `/tmp` 에서 생성 후 삭제. primary 체크아웃의 살아있는 큐는
  읽지도 건드리지도 않았다.
