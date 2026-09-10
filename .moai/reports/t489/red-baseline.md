# t489 RED baseline — errcheck `fmt.Fprintf` 미검사 (수리 전 측정)

> 카드: t489 · 브랜치: `WT-errcheck-catalog-hash` · 측정 트리: 본 워크트리 `.claude/worktrees/t489`
> base/HEAD: `615d18c1f` (= origin/develop, CI run 34014859880가 판정한 그 트리)
> 대상 파일 blob: `735b7d7b4856088dc7702a56fd450c0800327666` (`615d18c1f:internal/template/catalog_tree_hash.go`)
> 관측: 2026-09-06 · 본 세션(lane-15) 실측

## Claim (주장)

수리 전 트리(`615d18c1f`)에서 golangci-lint의 errcheck는 `internal/template/catalog_tree_hash.go:60:14` 에서 정확히 **1건**을 낸다. `./internal/template/...` 범위의 전체 발견은 이 1건뿐이다(다른 린터 발견 0).

## Evidence (증거 — 명령 + 실측 출력)

```
$ golangci-lint run ./internal/template/... --timeout=2m
internal/template/catalog_tree_hash.go:60:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
		fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)
		           ^
1 issues:
* errcheck: 1
```

- 종료코드는 1이지만 판별식은 매치 수다: errcheck 발견 행 **1행**, 요약 `* errcheck: 1`.
- CI 대조: run `34014859880` · 잡 Lint · golangci-lint v2.1.6 — 동일 파일·동일 줄(60:14)·동일 카운트(리드 직독, 배차문 인용).

## Baseline-attribution (귀속)

- **명령**: `golangci-lint run ./internal/template/... --timeout=2m` (파이프 없음, 본 세션에서 실행)
- **트리**: `.claude/worktrees/t489` @ `615d18c1f` — 2026-09-06 본 세션 실측
- **도구**: golangci-lint v2.10.1 (`/opt/homebrew/bin/golangci-lint`). CI의 v2.1.6과 바이너리가 다르지만, 커밋된 `.golangci.yml`(v2 형식)이 linter 셋(`default: none` + errcheck/govet/ineffassign/staticcheck/unused)과 errcheck 설정(`check-blank: false`)을 양쪽 동일하게 고정한다 — 이 버전 스큐 흡수는 설정 파일 자체가 문서화한 근거다(`.golangci.yml` 1-16행, 31-35행).
- **결함 성격 전제**: `:60`의 쓰기 대상은 `sha256.New()`(`hash.Hash`)이며, Go stdlib은 `hash.Hash`의 `Write`가 오류를 반환하지 않는다고 문서로 보장한다 — 수리 방향(`_ =` 명시 폐기)의 근거이며, 최종 판정은 본 카드 `verdict.md`에서 공식화한다.

## Gaps (미검증)

- 전체 테스트 스위트 미실행 — **의도된 범위**다. 레인 로컬 검증은 이 변경이 영향을 줄 수 있는 린터만 돌리고, 전수 판정은 CI 몫이다.
- develop push 이후 원격 CI의 재판정은 미관측 — 리드 일괄 push 후 판독 소관이다.

## Residual-risk (잔여 위험)

- v2.10.1 대 v2.1.6 바이너리 차이가 이 1건 판정 자체를 뒤집을 가능성은 낮지만 0이 아니다 — 최종 그린 판정은 develop push 후 원격 CI에서 확인된다.
