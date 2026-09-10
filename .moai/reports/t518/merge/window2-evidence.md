# 2차 창 증거 — 카드 t518 (lane-4)

이 파일은 **창 안에서 저작되고 창 안에서 커밋된다.** 기록이 새 커밋을 만들고 그것이
다시 병합 대상이 되는 재귀를, 창 안에서 끊기 위해서다
(일반형: `feedback_make_the_evidence_commit_inside_the_window_it_documents.md`).

## 창

```
moai integration acquire --name lane-4
  → acquired by b98c9746… on WT-spec-lint-axes
  ※ settings drift 없음 (1차 창과 달리 무경고)
```

## 흡수 직전 재측정 — 값이 또 움직였다

리드가 준 값과 일치했고, 1차 창에서 흡수한 tip 은 이미 낡아 있었다.

```
develop tip     = 921ed083c      (1차 창에서 흡수한 08609c198 이 아니다)
내 HEAD         = 465ed2672
develop..HEAD   = 24
HEAD..develop   = 4              ← develop 이 4커밋 앞섬 → 재흡수 필요
```

## 재흡수 — 충돌 0

```
git merge develop  → Merge made by the 'ort' strategy (충돌 없음)
재흡수 커밋        = 8c4f8a215
```

## 재흡수가 무엇을 들여왔는지 — 세었다

```
git diff --name-only 465ed2672..HEAD
  .moai/reports/t536/merge/count-scope-rederivation.md
  .moai/reports/t536/merge/window2-evidence.md

git diff --name-only 465ed2672..HEAD | grep -vc '^\.moai/reports/'
  0        ← .moai/reports/ 밖 변경 0파일
```

들어온 것은 lane-3 의 증거 문서 2개뿐이다. Go 소스·go.mod·설정·`.moai/specs/`
어느 것도 바뀌지 않았다.

## 재측정 — 무엇을 다시 재고 무엇을 잇는가

**다시 잰 것.** `internal/spec` 은 `.moai/specs/` 코퍼스를 읽는 테스트를 갖고 있어
코퍼스가 움직이면 결과가 달라질 수 있다. 실제 병합 트리 `8c4f8a215` 에서 재실행했다.

```
go test ./internal/spec/... -count=1   → exit 0
  ok github.com/modu-ai/moai-adk/internal/spec  67.978s
```

**이은 것.** `internal/cli` 와 `internal/kanban` 은 `465ed2672` 측정을 잇는다. 근거는
위 계수다 — 두 커밋 사이의 변경이 `.moai/reports/` 하위 마크다운 2파일뿐이고 그
경로는 어느 테스트도 읽지 않는다. 이것은 「안 바뀌었을 것이다」가 아니라
**변경 파일 전수 열거 + 비-reports 계수 0** 이라는 측정이다.

`465ed2672` 측정치(그 커밋의 `revision-record.md` 가 정본):

```
go vet ./internal/spec/... ./internal/cli/... ./internal/kanban/...  → exit 0, 출력 0줄
gofmt -l internal/spec/lint_req_table_test.go                        → 무출력
go test ./internal/kanban/... -count=1                               → ok 142.460s
go test ./internal/cli/...    -count=1 -timeout 900s                 → ok 461.510s
                                                                        (+ 하위 16패키지 ok)
```

## 잔여 미검증

- **CI 판정 없음** — 미푸시. 로컬 초록은 조기 신호이며 darwin/windows 매트릭스가 아니다
- `internal/cli` / `internal/kanban` 은 `8c4f8a215` 에서 **직접 실행하지 않았다.**
  위 계수로 이었을 뿐이며, 그 이음의 전제(비-reports 변경 0)가 이 파일에 적혀 있다
- 1차 창에서 검출된 `settings.json` 드리프트는 손대지 않았다 — 리드가 운영자에게 올린다
