# t489 판정 — errcheck `fmt.Fprintf` 미검사 수리 (SPEC-ERRCHECK-CATALOG-HASH-001)

> 카드: t489 · 브랜치: `WT-errcheck-catalog-hash` · 측정 트리: 본 워크트리 `.claude/worktrees/t489`
> base: `615d18c1f` (= origin/develop, CI run 34014859880가 판정한 그 트리)
> 커밋 체인: RED baseline `325758c57` → 수리 `9f341f371` → 본 판정서 (§2.3 baseline-first 순서를 커밋 그래프가 증언)
> blob: 수리 전 `735b7d7b4856088dc7702a56fd450c0800327666` / 수리 후 `104a95236e4510aa8510f4532a5aa7d9031ee1e0`
> 관측: 2026-09-06 · lane-15 본 세션 실측 (단일 측정자 — 레인이 모든 측정 수행)

## Claim (주장)

- **C1**: `internal/template/catalog_tree_hash.go:60`의 errcheck 발견(`fmt.Fprintf` 반환값 미검사)은 명시적 폐기 `_, _ =` + `hash.Hash` never-error 계약 주석으로 수리됐다. 로컬 4회 측정: **RED 1 → GREEN 0 → 뮤턴트 1 → 재확인 0**. `go build ./internal/template/` 클린.
- **C2**: 수리 방향의 근거 — `:60`의 쓰기 대상은 `sha256.New()`(`hash.Hash`)이고, Go stdlib은 `hash.Hash.Write`가 오류를 반환하지 않는다고 문서로 보장한다. 따라서 오류는 구조적으로 항상 nil이고, 전파는 죽은 경로(REQ-003)다. 「린터를 조용히 시킨 것」이 아니라 존재할 수 없는 오류 경로의 정직한 표시다.
- **C3**: 첫 편집의 `_ =`(1개 폐기)는 `fmt.Fprintf` 반환값 2개 때문에 컴파일 오류였다(`assignment mismatch: 1 variable but fmt.Fprintf returns 2 values` — go build 실측). `_, _ =`가 유효형이며 AC-SCOPE의 `_ =` 술어는 부분 문자열로 그대로 성립한다 — SPEC 문서 수정 없이 해결.
- **C4**: AC 5건 전부 충족 (AC-GREEN · AC-MUTANT · AC-ORDERING · AC-SCOPE · AC-EVIDENCE — 아래 Evidence 참조).

## Evidence (증거 — 명령 + 실측 출력, 본 세션 본 트리)

### Run A — RED (수리 전 트리 `615d18c1f`)

```
$ golangci-lint run ./internal/template/... --timeout=2m
internal/template/catalog_tree_hash.go:60:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
		fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)
		           ^
1 issues:
* errcheck: 1
```

CI 대조: run `34014859880` · 잡 Lint · golangci-lint v2.1.6 — 동일 파일·동일 줄·동일 카운트. 상세: `red-baseline.md`.

### 수리 커밋 전 빌드 검증

```
$ go build ./internal/template/          # `_ =` 1개 폐기 상태
internal/template/catalog_tree_hash.go:62:7: assignment mismatch: 1 variable but fmt.Fprintf returns 2 values
$ go build ./internal/template/          # `_, _ =` 정정 후 — 무출력, exit 0
```

### Run B — GREEN (수리 커밋 `9f341f371`)

```
$ golangci-lint run ./internal/template/... --timeout=2m
0 issues.
```

### Run C — 뮤턴트 (HEAD `9f341f371`에서 파일만 base blob으로 복원: `git checkout 615d18c1f -- internal/template/catalog_tree_hash.go`)

```
$ golangci-lint run ./internal/template/... --timeout=2m
internal/template/catalog_tree_hash.go:60:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
		fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)
		           ^
1 issues:
* errcheck: 1
```

→ 검사가 그 파일을 실제로 본다 — GREEN이 공허하지 않음이 반대 방향에서 재확인됐다.

### Run D — 재확인 (`git checkout HEAD -- internal/template/catalog_tree_hash.go` 복원 후)

```
$ golangci-lint run ./internal/template/... --timeout=2m
0 issues.
```

### AC-SCOPE 폐쇄 검사

```
$ git diff 615d18c1f..HEAD -- '*.go'
```

→ 브랜치 전체 Go diff는 정확히 1훙크(위 수리 형상뿐)이며 `//nolint` 지시문 0건. 신규 테스트 파일 없음, `.golangci.yml` 변경 없음.

## Baseline-attribution (귀속)

- **명령**: `golangci-lint run ./internal/template/... --timeout=2m` — 매 회 파이프 없이 단독 실행, 판별식은 발견 행 수와 요약 계수(종료코드 아님).
- **트리 상태**: Run A = `615d18c1f`(blob `735b7d7b4`) · Run B = `9f341f371`(blob `104a95236`) · Run C = `9f341f371` + 파일만 base blob · Run D = `9f341f371`(파일 HEAD와 바이트 동일 — `git status`에서 Go 파일 침묵으로 확인). 2026-09-06 본 세션 실측.
- **도구**: golangci-lint v2.10.1 (`/opt/homebrew/bin`). CI v2.1.6과 바이너리 상이하나 커밋된 `.golangci.yml`이 linter 셋과 errcheck 설정(`check-blank: false`)을 양쪽 동일 고정(설정 파일 1-16·31-35행이 이 근거를 문서화). 컴파일 판정은 go(로컬 툴체인)가 수행.
- **plan-audit**: 3회 반복, 최종 **PASS 1.00**(Tier M 임계 0.80), 누적 결함 7건(D1-D5, N1-N2) 전부 해결 — `plan-audit.md`.

## Gaps (미검증)

- develop push 이후 **원격 CI의 재판정은 미관측** — 리드 일괄 push 후 판독 소관이다.
- 전체 테스트 스위트 미실행 — **의도된 범위**(AC-SCOPE)다. 두 반환값의 명시 폐기는 의미 변화가 불가능하고, 패키지 동작은 기존 테스트가 계속 보증한다.
- AC-GREEN의 패키지 요약 조항: golangci-lint는 이슈 0일 때 `* errcheck: 0`이 아니라 `0 issues.`를 인쇄한다 — 본 판정은 그 모양으로 충족됐음을 기록한다(파일 범위 카운트 0 + 전체 이슈 0).

## Residual-risk (잔여 위험)

- v2.10.1 대 v2.1.6 바이너리 차이가 단일 발견 판정을 뒤집을 가능성은 낮지만 0이 아니다 — 최종 그린은 develop push 후 원격 CI에서 확인된다.
- sync 단계는 미수행 — 수리 카드로서 문서 영향이 없다고 판단하나, sync 필요성 판단과 병합 창은 리드 소관이다.
- 로컬 `develop` ref가 `ce4f96869`(t472 병합)로 origin/develop의 조상이 아닌 것이 관측됐다(`git merge-base --is-ancestor` exit 1) — 레인 소관 밖이므로 보고만 한다. 카드 브랜치는 origin/develop에서 떴다.
