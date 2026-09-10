# t481 — 병합 창 재측정 기록 (2026-09-04)

- 창: lane-8 (acquire 기록 — 홀더 세션 `6f01ec8e`, 브랜치 `WT-cli-cover-suppression`)
- 병합할 팁: `ce106e3a9` — **측정 트리 `206135c404982be272a31110013092213051cd95`**
- develop 팁(창 획득 시점 재판독): `8b391bc8c` — 흡수 시도 결과 "Already up to date." (창 잡는 시점까지 develop 불변 확인)

## 측정

```
$ go test -count=1 -cover -timeout 900s ./internal/cli/...   # 트리 206135c40, exit 0 (백그라운드 태스크 출력 직독)
ok  github.com/modu-ai/moai-adk/internal/cli  526.766s  coverage: 80.6% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/agentlint   3.953s  coverage: 86.7% of statements
... (17개 전체 — 원문: window-remeasure-sweep.txt)
```

## 카드 측정과의 대조

17개 패키지의 커버리지 수치가 **전부 동일**(main 80.6% 포함) — 실행 시간만 다르다(루트 438.373s → 526.766s). 수치 재현 확인. §6 판정(main 80.6% 미달, `internal/cli/harness` 80.9% 미달)에 변화 없음.

## 귀속

- 측정 시점 트리는 `206135c40`(커밋 `ce106e3a9`) — 이 기록을 커밋하면 트리가 움직인다.
- 따라서 병합 트리와 측정 트리의 차이는 **본 증거 파일뿐**임을 `git diff --stat ce106e3a9..<병합 후 팁>`으로 확인한다(t470 선례의 델타 귀속 형태 — "재측정은 병합할 tip에서" 규율에 따라 측정 뒤 증거 커밋을 얹으므로, 트리 동일성 대신 델타=증거파일뿐로 귀속).
- Go 소스 축 차이 0 — 커버리지 수치의 develop 정착 귀속 성립.
