# t456 — 병합 트리 재측정 (창 밖 선수행)

리드 지시: *"호명 전에 흡수·재측정을 창 밖에서 끝내 두시면 창을 비우지 않습니다."*
그리고 §4.1 규율: **병합 전 검증을 병합 후 근거로 재사용하지 않는다.** 이 문서가 그 재측정이다.

## 흡수

```
흡수 전:  develop 4e4607abe · 뒤처짐 134 / 앞섬 5
병합:     55881565f  Merge branch 'develop' into WT-statusline-landed
충돌:     CHANGELOG.md 1건 — 최상단에 양쪽이 각자 항목을 추가
해소:     둘 다 보존 (t456 항목 + develop 의 SPEC-PROJECT-CONTINUATION-KEY-001 항목).
          어느 쪽도 버리지 않았고 마커는 0.
미해결 경로: 0 (`git diff --name-only --diff-filter=U` 무출력)
```

## 재측정 — 판정 바이너리는 병합 트리에서 다시 빌드했다

```
$ go build -o bin/moai-t456 ./cmd/moai      → BUILD_RC=0
```

병합 전 바이너리를 재사용하지 않는다. 이 규율이 실제로 값을 한 사례가 §정정 1
(`verdict.md`)이다 — 구현 이전 빌드로 잰 매트릭스가 통과처럼 보였다.

### R1 — 렌더 경로 git spawn: 2 (baseline 유지)

```
$ ./bin/moai-t456 statusline < payload-wt.json     # 캐시 부재
🔄 TODO: 39/4                                       ← 주석 없음 (정확)

기록된 spawn:
rev-parse --git-dir --show-toplevel
status --porcelain --branch
→ 2
```

non-self-invocable 이름으로 재 자식을 격리했다. 병합 전과 동일.

### R2 — 판정 확보 시 주석이 붙는다

```
$ ./bin/moai-t456 statusline --refresh-landed --board-root <worktree>
{"landed":23,"ref":"origin/develop","measured":true,"fetched_at":1788412455}

$ ./bin/moai-t456 statusline < payload-wt.json
🔄 TODO: 39/4 ✓23
```

두 상태가 갈린다 — 부재는 안 붙고 확보는 붙는다. 병합 전 매트릭스와 같은 판별력.

### R3 — 테스트

```
$ go test -count=1 -timeout 600s ./internal/statusline/...   → ok 13.928s  EXIT=0
$ go test -count=1 -timeout 900s ./internal/cli/             → (별도 기록)
```

## Gaps — 이 재측정이 관측하지 않은 것

- **교대 실행 지연은 다시 재지 않았다.** base arm(`/tmp/t456_moai_base`)은 흡수 **전**
  트리의 빌드라, 병합 트리의 after 와 짝지으면 134커밋의 다른 변경까지 델타에 섞인다.
  병합 전 트리에서 잰 `+1.53ms (+1.0%)` 가 이 카드 변경에 귀속되는 값이고, 병합 트리의
  절대 지연은 **미측정**이다. 렌더 경로 spawn 이 2로 유지된 것이 회귀 없음의 기계적 근거다.
- **`golangci-lint` 미실행**, `go vet` 만 돌렸다.
- **CI 판정 없음.** 로컬 darwin 단일 플랫폼이며 이 브랜치는 push 되지 않았다.
- 상속된 `agents-emit-check` 적색은 그대로다 — 리드 판정으로 **t443 소관**이며 이 카드가
  고치지 않는다. `make build` 가 막히면 `bin/moai` 가 동결되고 doctor 의 `Agent Emit Embed`
  가 stale embed 를 보고해 doctor 테스트가 실패하는 하류까지 같은 뿌리다(lane-1 확정).
