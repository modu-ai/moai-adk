# 2차 창 증거 — 카드 t536 (lane-3)

1차 창(병합 `08609c198`) 이후 생긴 문서 1본(`97e99e327`)을 실으려는 짧은 창이다.

## 흡수 직전 재측정 (배차문 값을 그대로 쓰지 않았다)
```
develop tip (재측정)        = 08609c198   ← 배차문과 같았으나 다시 쟀다
내 브랜치 vs 로컬 develop   = 1 1
흡수 커밋                   = c138520c6
```

## 판정식 — 코드 표면 무이동

코드 diff 가 0 이므로 `go test` 는 **이 변경에 대해 아무것도 말하지 않는다.**
그래서 테스트를 돌리지 않았고, 대신 흡수 후 이 브랜치가 develop 에 더하는 것이 문서 1파일뿐임을 diff 로 세운다.
흡수가 데려온 남의 코드 변경은 develop 이 이미 검증한 것이며 내 카드의 변경이 아니다.

```
git diff --stat 08609c198..HEAD
  .moai/reports/t536/merge/count-scope-rederivation.md | 40 ++++++++++
  1 file changed, 40 insertions(+)

git diff --name-only 08609c198..HEAD -- '*.go' '*.yaml' '*.yml' '*.json' '*.tmpl' Makefile
  (무출력)   ← 코드 표면 무이동

대조군: git diff --name-only 08609c198..HEAD | wc -l  → 1
  같은 명령이 경로 필터 없이는 1건을 찾는다 — 위 무출력은 침묵이 아니라 측정이다.
```

## 돌리지 않은 것과 그 이유 (간극이 아니라 판정)

- `go test ./internal/{kanban,web,cli}/` — **미실행.** 코드 diff 0 이라 이 변경을 재지 못한다. 1차 창에서 병합 트리 재측정 완료(vet 3패키지 · windows 빌드 · 테스트 3패키지 ok, 트리 동일성 `ce96e323d…` 로 이음).
- `go vet` / `GOOS=windows` 빌드 — **미실행.** 같은 이유.
- `golangci-lint` — 미실행(CI 몫, 이 카드 내내 그러했다).

안 돌린 것을 적지 않으면 읽는 쪽이 「빠뜨렸나」와 「일부러 안 했나」를 구별할 수 없다. 그 구분은 적혀 있어야 성립한다.
