# t665 — `moai todo done` 이 기록된 착지 증거를 읽지 않는다

- 카드: t665 (구 t661 에서 id 충돌로 재발행)
- 브랜치: `WT-todo-landing-record`
- base: 로컬 develop `fb8bfff95`
- 출처: 리드 실측 2026-09-12 — `landed t555 --sha c9a9e8866` 기록 성공 직후 `done t555` 가 `landing=unknown`

---

## 1. 두 갈래 판정 — (b) 이다

카드가 먼저 가르라고 한 두 갈래:

| 갈래 | 내용 | 판정 |
|---|---|---|
| (a) | `landed` 가 쓴 곳과 `done` 이 읽는 곳이 다른 저장 위치인가 | **아니다** |
| (b) | `done` 이 그 필드를 아예 조회하지 않는가 | **그렇다** |

근거는 코드 판독이 아니라 **저장소 직접 판독**이다. `TestLandingEvidenceSurvivesArchive_StoreLevel` 이
엔진 데이터베이스를 열어 두 시점의 컬럼을 SQL 로 읽는다:

- `landed --sha` 직후 — `SELECT landing IS NULL FROM items WHERE id = 't1'` → false (기록 존재)
- `done` 직후 — `SELECT landing IS NULL FROM archived_items WHERE id = 't1'` → false (증거 보존)

즉 `landed` 가 쓴 레코드는 아카이브 행까지 그대로 따라간다. `ArchiveCard` 가 항목을 통째로 복사하고
두 카드 보유 테이블이 같은 `landing` 컬럼을 갖기 때문이다. 저장 측에는 결함이 없다.

결함은 **읽는 쪽 두 곳**이 비어 있다는 것이다.

1. `done` 의 판정값은 `--require-landed` 플래그가 있을 때만 계산된다. 플래그가 없으면
   `kanban.LandingUnknown` 그대로 출력된다 — 저장된 레코드는 조회하지 않는다.
2. `todo history` 는 착지 증거를 **어떤 형태로도 렌더링하지 않는다**. 아카이브를 읽는 유일한
   표면인데도 배달 커밋을 돌려주지 않으므로, 닫힌 카드의 배달 커밋은 데이터베이스를 직접
   열어야만 복구된다 — 이것이 보고된 "추적성 상실" 의 실체다.

플래그를 켠 경로조차 저장된 증거가 아니라 `GitLandedQuerier` 의 커밋 메시지 조회로 답한다.
따라서 어느 경로에서도 운영자가 기록한 SHA 는 판정에 도달하지 않았다.

---

## 2. 재현 (RED) — 수리 전, 이 트리 이 런

```
go test ./internal/cli/ -run 'TestLandingEvidenceSurvivesArchive_StoreLevel|TestLandingEvidenceRoundTrip' -count=1 -timeout 600s
```

```
--- FAIL: TestLandingEvidenceRoundTrip (1.35s)
    todo_landing_roundtrip_test.go:85: done reported landing=unknown for a card carrying recorded evidence: "done t1 landing=unknown"
    todo_landing_roundtrip_test.go:96: history t1 does not return the recorded SHA e4ebedfebc980d3c57d8529b7cea842053484be5: "t1\tarchived\tqueued\tcard whose landing must round-trip"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.435s
```

같은 실행에서 store-level 시험은 통과했다 — 그것이 (a) 를 기각한 관측이다.

---

## 3. 수리

읽는 쪽만 고쳤다. 저장 형식, `landed` 의 쓰기 경로, 아카이브 이관은 건드리지 않았다.

**`internal/cli/todo.go` — `done`**

- 뮤테이션 콜백 안에서 `rec.Items[at].Landing` 을 읽어 둔다 (아카이브가 행을 옮기기 전).
- `todoDoneLandingSuffix` 가 **배달 SHA 를 담고 자체 검증을 통과한** 레코드에 한해
  `sha=<커밋> source=operator` 를 만든다. 관측된 ref 위치만 가진 레코드는 아무것도
  덧붙이지 않는다 — 그 레코드는 배달 커밋을 주장하지 않으며, 그것을 배달로 렌더링하는 것이
  REQ-TLE-013 이 막으려는 혼동이다.
- 플래그가 없을 때 판정값은 `landed` 가 된다. 검증된 레코드는 **이전에 얻어 저장해 둔 답**이므로
  `unknown`("조회가 돌지 않았다") 은 그 상태에 대한 정직한 보고가 아니다.
- 출처(`source=operator`)가 값과 함께 이동하므로, 저장된 운영자 주장이 이번 런의 조회 결과로
  오독되지 않는다.
- `--require-landed` 경로의 판정값은 **바꾸지 않았다** — git 조회가 여전히 그 가드의 답이고,
  `not-landed` 거절도 그대로다. 저장된 증거는 접미사로만 붙는다.

**`internal/cli/todo_history.go` — `history`**

- `live` / `archived` 두 줄 모두 카드 텍스트 **앞에** 착지 열을 갖는다. 이 파일의 계약이 이미
  "텍스트가 마지막이므로 열 추가는 꼬리를 읽는 소비자에게 무해하다" 를 확장 지점으로 명시한다.
- 네 상태를 구분한다: `landing=<SHA>` / `landing=ref-head` / `landing=-` / `landing=malformed`.
  "기록 없음" 과 "배달 커밋을 주장하지 않는 기록" 은 카드에 대한 서로 다른 사실이라 하나의
  대시로 합치지 않았다.
- SHA 는 **전체 길이**로 렌더링한다. `pr` 이 7자로 줄이는 것은 정렬된 표에서 열 너비가 희소
  자원이기 때문이고, `history` 는 탭 구분 비정렬 줄이라 같은 제약이 없다. 이 열의 목적 자체가
  두 번째 조회 없이 바로 쓸 수 있는 값을 돌려주는 것이다.

**문서** — `.claude/skills/moai/workflows/todo.md` 의 `done` / `history` 행을 갱신하고
템플릿 미러에 같은 변경만 적용했다(두 사본은 다른 문단에서 의도적으로 다르므로 통째 `cp` 는
되돌렸다).

---

## 4. 회귀 — 왕복으로 세웠다

`landed --sha X` → `done` → `history` 가 X 를 돌려주면 GREEN. "landing 필드가 존재한다" 는
단정은 쓰지 않았다.

```
go test ./internal/cli/ -run 'TestLandingEvidenceSurvivesArchive_StoreLevel|TestLandingEvidenceRoundTrip|TestLandingBackfillPathForAlreadyClosedCard' -count=1 -v -timeout 600s
--- PASS: TestLandingEvidenceSurvivesArchive_StoreLevel (1.24s)
--- PASS: TestLandingEvidenceRoundTrip (1.40s)
--- PASS: TestLandingBackfillPathForAlreadyClosedCard (1.85s)
ok  	github.com/modu-ai/moai-adk/internal/cli	5.474s
```

### 전량 실행 (`internal/cli`)

```
go test ./internal/cli/ -count=1 -timeout 1800s
ok  	github.com/modu-ai/moai-adk/internal/cli	1606.392s        EXIT=0   (--- FAIL 0건)
```

**측정 순서에 관한 정직한 기록.** 이 전량 실행은 `todo_history.go` 의 `--help` 본문(cobra
`Long` 문자열)에 새 열 설명을 추가하기 **전** 트리에서 시작됐다. 그 편집은 도움말 문자열
뿐이고 이를 단정하는 시험은 저장소에 없지만(`grep` 으로 확인), 전량 결과를 편집 후 트리의
근거로 그대로 쓰지는 않는다. 편집이 닿는 범위를 덮는 재측정을 따로 돌렸다:

```
go test ./internal/cli/ -run 'Todo|Landing|Landed|History' -count=1 -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/cli	296.178s
```

증거 파일: `.moai/reports/t665/full-run/cli-full.txt`, `cli-full-exit.txt`,
`todo-scoped-rerun.txt`.

### 기존 계약 시험

기존 계약 시험 3건(`TestTodoHistoryReportsLiveCard`, `TestTodoHistoryReportsArchivedCard`,
`TestTodoHistoryDegradesWithoutArchiveTables`)은 열이 하나 늘어 기대값을 `landing=-` 포함으로
갱신했다. 의도된 계약 확장이며, 변경 이유를 시험 본문 주석에 남겼다.

---

## 5. 이미 닫힌 8장의 백필 경로 — 성립한다

`undone → landed → done` 이 실제로 증거를 복구하는지 픽스처 큐에서 측정했다
(`TestLandingBackfillPathForAlreadyClosedCard`, 위에서 PASS). **실행은 하지 않았다** — 큐 쓰기는
리드 소관이다.

운영자용 절차 (카드 1장당):

```
moai todo undone <id>                  # 아카이브에서 살아있는 큐로 복귀
moai todo landed <id> --sha <배달커밋>  # 증거 기록 (ref 미지정이면 통합 ref 로 해석)
moai todo done <id>                    # 재마감 — 이제 landing=landed sha=... source=operator
moai todo history <id>                 # 확인 — landing=<배달커밋>
```

대상 8장: t525 t550 t552 t554 t555 t556 t559 t560.

**주의 두 가지.**

1. `landed` 는 배달 SHA 가 해당 ref 에 실제로 있는지 확인하고, 확인되지 않으면 기록을 거절한다.
   따라서 8장 각각의 배달 커밋을 먼저 확정해야 하며, 이 절차는 그 확정을 대신하지 않는다.
2. `undone` 은 카드를 살아있는 큐로 되돌린다. 여러 레인이 같은 큐를 동시에 읽는 상태에서
   8장이 잠깐 큐에 다시 보이므로, 배치가 조용한 시점에 한 번에 처리하는 편이 안전하다.

---

## 6. 격리 방법 (카드 [HARD])

실제 큐에는 쓰지 않았다. 두 가지를 썼다.

**(1) 저장소 위치 파악 — 셸 실험실.** 임시 디렉터리(`…/scratchpad/t665-lab`)에서 이 트리로
빌드한 `bin/moai` 를 실행했다. 큐 루트 해석이 임시 루트를 감지해 홈 큐 생성을 거절하고
project-local 큐로 내려가므로, 이것만으로 격리가 성립한다:

```
moai todo: the launch directory is inside the temporary root /tmp, so no home queue was created
under ~/.moai/db/<project-key>/todo; continuing against the project-local queue at …/t665-lab
→ t1 1
→ 저장소: …/t665-lab/.moai/state/todo/backlog.db  (SQLite)
```

다만 `landed` 는 ref 를 해석할 수 있는 git 저장소를 요구하는데, 워크트리 격리 세션의 가드가
워크트리 바깥을 겨누는 셸 git 을 거부하므로 이 실험실에서는 왕복을 완성할 수 없었다.

**(2) 왕복 재현 — Go 픽스처.** 그래서 실제 측정 수단은 기존 테스트 픽스처다:
`todoFixture` 가 `t.TempDir()` + `CLAUDE_PROJECT_DIR` 재지정 + 임시 git 저장소를 만들고,
`newLandedFixture` 가 카드 id 를 언급하는 커밋 3개와 아무것도 언급하지 않는 tip 을 고정한다.
큐도 git 이력도 테스트별 임시 트리 안에만 존재하므로 실제 큐와 접점이 없다.

---

## 7. 5-섹션 판정

**Claim.** `done` 이 기록된 착지 증거를 읽지 않아 아카이브가 배달 커밋을 잃는 결함을 읽기
측면에서 수리했고, 왕복 회귀와 백필 경로를 측정했다.

**Evidence.** §2 (RED 실측 출력), §4 (GREEN 실측 출력), §5 (백필 PASS),
`go vet ./internal/cli/` exit 0, `golangci-lint run ./internal/cli/...` → `0 issues.`,
전량 실행 → `.moai/reports/t665/full-run/cli-full.txt` + `cli-full-exit.txt`.

**Baseline-attribution.** 전부 이 워크트리(`.claude/worktrees/t665`), 브랜치
`WT-todo-landing-record`, base 로컬 develop `fb8bfff95` 에서 2026-09-12 에 측정. 부하 게이트
측정 직전 값: RED/GREEN 구간 `load averages: 6.31–15.75` (임계 30 미만).

**Gaps.**
- 실제 큐(홈 `~/.moai/db/<project-key>/todo`)에서는 아무것도 실행하지 않았다 — 백필도,
  왕복도. 카드 지시에 따른 것이며, 따라서 "실제 큐에서도 동일하게 동작한다" 는 미검증이다.
- 8장 각각의 배달 커밋을 확정하지 않았다. §5 절차는 그 값이 이미 있다고 전제한다.
- `--require-landed` 와 저장 증거가 **불일치**하는 경우(조회는 not-landed, 기록은 landed)를
  별도 시험으로 세우지 않았다. 현 동작은 조회가 이기고 거절한다 — 의도한 우선순위지만
  시험으로 고정하지는 않았다.
- 다른 읽기 표면(`todo pr`, 웹 콘솔, statusline)이 이 열 추가에 영향받는지는 보지 않았다.
  `pr` 은 자기 렌더러를 그대로 쓰므로 영향이 없어야 하지만, 이는 코드 판독에 근거한 기대이지
  측정이 아니다.

**Residual-risk.**
- `history` 의 줄 모양이 바뀌었다. 텍스트를 마지막 필드로 읽는 소비자는 무해하지만, 네 번째
  탭 필드를 **텍스트로** 읽는 소비자가 저장소 밖에 있다면 깨진다. 저장소 안에서는
  `grep -rn "todo history" --include='*.go' --include='*.sh' --include='*.js' --include='*.ts'`
  (워크트리·테스트 제외) 가 기계 소비자를 하나도 돌려주지 않았고, 문서 소비자는 스킬 문서
  두 벌(로컬 + 템플릿 미러)뿐이라 둘 다 갱신했다. 시험 3건도 갱신했다. 저장소 **밖**은
  확인할 방법이 없다.
- `done` 의 판정값이 플래그 없이도 `landed` 가 될 수 있게 됐다. 운영자가 잘못된 SHA 를
  기록해 두었다면 그 오류가 이제 판정값까지 전파된다 — `landed` 의 ref 소속 검증이 1차
  방어이고, `source=operator` 가 그 값의 출처를 노출하는 것이 2차 방어다.
- t648(큐 통일)과 같은 영역이다. 충돌은 관측되지 않았으나 두 변경이 같은 파일을 만진다면
  병합 시 재측정이 필요하다.
