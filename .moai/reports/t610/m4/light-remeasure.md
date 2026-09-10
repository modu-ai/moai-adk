# t610 M4 — 흡수 후 가벼운 재측정 (lane-8)

- 흡수: 카드 브랜치 `WT-go-1266`에서 `git merge --no-edit develop` 실행. 충돌은 없었고 ort 전략으로 병합 커밋이 만들어졌다(develop 쪽 211개 파일 반영).
- 흡수한 로컬 develop 끝: `f3c4ca50ce9b1ae95711c4452cd3a65ffd27449d` (리드 지명값 `f3c4ca50c`와 일치)
- 병합 트리: `git rev-parse HEAD 'HEAD^{tree}'` → HEAD `180bc24f1e7516624c1b6400861268000f9af63d`, tree `c85c42aeaf90f104f3b8a36f27bbfd8542e1da65`
- `git merge-base develop HEAD` → `f3c4ca50ce9b1ae95711c4452cd3a65ffd27449d`. 흡수로 merge-base가 develop 끝으로 옮겨졌다(흡수 전 `d3b7d438d`).
- 창: `moai integration acquire --name lane-8` → `release-integration window acquired by 7e33c4bd-a91e-4a67-bf0c-15ad255da456 on WT-go-1266`

아래 명령은 모두 이 병합 트리에서 도구 호출 하나당 명령 하나로 실행했다. git 명령의 종료 코드는 도구가 보고한 값이다. 도구는 0이 아닌 종료만 오류로 표시하며, 아래 명령은 모두 오류 표시 없이 끝났다.

## AC-GTS2-002 — 지시어

```text
$ sed -n 3p go.mod
go 1.26.8
$ grep -c '^toolchain' go.mod
0
```

## AC-GTS2-003 — go.mod 한 줄 변경

```text
$ git diff --numstat develop...HEAD -- go.mod
1	1	go.mod
$ git diff develop...HEAD -- go.mod
@@ -1,6 +1,6 @@
 module github.com/modu-ai/moai-adk
 
-go 1.26.4
+go 1.26.8
 
 require (
```

develop이 흡수된 뒤에도 go.mod 변경은 이 헌크 하나다. 흡수한 develop 쪽에는 go.mod 변경이 없었다.

## AC-GTS2-007 — 변경 범위 (M4 형태, sync 경로 5개 제외 포함)

```text
$ git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610' ':!CHANGELOG.md' ':!.moai/project/product.md' ':!.moai/project/structure.md' ':!.moai/project/codemaps/overview.md' ':!.moai/project/codemaps/modules.md'
(출력 없음)
```

대조군: 제외 조건 없이 경로 목록을 파일로 저장했다(`git diff --name-only --output=.moai/reports/t610/m4/ac007-control-unexcluded.txt develop...HEAD`).
- `wc -l`로 센 경로 수는 `90`이다.
- 허용 집합(go.mod, sync 경로 5개, SPEC 디렉터리, `.moai/reports/t610/`)을 벗어난 줄은 `grep -c -v -E '<허용 집합 정규식>'` 결과 `0`이다.

따라서 제외 조건을 넣었을 때의 빈 출력은 명령이 헛돈 결과가 아니다. (이 절 초안에는 손으로 셈한 "91개"가 적혀 있었다. 기계로 다시 세어 `90`으로 바로잡았다.)

## § D.0b 5단계 — sync 경로 numstat 대조

```text
$ git diff --numstat develop...HEAD -- CHANGELOG.md .moai/project/product.md .moai/project/structure.md .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md
1	1	.moai/project/codemaps/modules.md
1	1	.moai/project/codemaps/overview.md
2	2	.moai/project/product.md
1	1	.moai/project/structure.md
1	0	CHANGELOG.md
$ git show --numstat --format= 4516fbe40 -- (같은 5개 경로)
1	1	.moai/project/codemaps/modules.md
1	1	.moai/project/codemaps/overview.md
2	2	.moai/project/product.md
1	1	.moai/project/structure.md
1	0	CHANGELOG.md
```

두 출력이 같다. 흡수가 이 경로들을 건드리지 않았고, 카드 쪽 변경은 sync 커밋 하나뿐이다. EC-5는 걸리지 않는다.

## AC-GTS2-008 — 문서 표기

```text
.moai/project/product.md            1.26.4=0 1.26.8=2
.moai/project/structure.md          1.26.4=0 1.26.8=1
.moai/project/codemaps/overview.md  1.26.4=0 1.26.8=1
.moai/project/codemaps/modules.md   1.26.4=0 1.26.8=1
```

`1.26.4`는 합계 0건, `1.26.8`은 합계 5건이다. 편집 전 대조군(2건)은 `sync/ac008-evidence.md`에 있다.

## 무거운 검사

AC-001(도구 체인), AC-004(govulncheck), AC-005(make build, go version -m), AC-006(go vet, 대표 테스트)은 부하가 겹치지 않게 하나씩 돌렸다. 결과는 같은 디렉터리의 `ac00*` 파일과 `.exit` 파일에 있다.

| AC | 명령 | 관측 출력 | 종료 코드 | 파일 |
|---|---|---|---|---|
| AC-GTS2-001 | `go version`; `go env GOTOOLCHAIN GOMOD GOVERSION` | `go version go1.26.8 darwin/arm64`; `auto` / 워크트리 `go.mod` / `go1.26.8` | 0, 0 | `ac001-go-version.txt`, `ac001-go-env.txt` |
| AC-GTS2-004 | `GOMAXPROCS=2 timeout 900 govulncheck ./...` (GOTOOLCHAIN override 없음) | `Your code is affected by 0 vulnerabilities.` / `0 vulnerabilities in packages you import and 3` / `vulnerabilities in modules you require` | 0 | `ac004-govulncheck.log` |
| AC-GTS2-004 ID | 8개 ID 각각 `grep -c` | 새 로그 8개 모두 `0`, 기준선 로그 `baseline/govulncheck-auto.log` 각 `2` | — | (명령 출력 인라인, 아래) |
| AC-GTS2-005 | `timeout 900 make build`; `go version -m bin/moai` | 빌드 마지막 줄 `go build ... -X .../version.Commit=180bc24f1 ... -o bin/moai ./cmd/moai`; 첫 줄 `bin/moai: go1.26.8` | 0(백그라운드 작업 완료 알림 기준), 0 | `ac005-make-build.log`, `ac005-go-version-m.txt` |
| AC-GTS2-005 추적 파일 | `git status --porcelain --untracked-files=no` (빌드 후) | 출력 없음 | 0 | — |
| AC-GTS2-006 vet | `timeout 900 go vet ./...` | 출력 0바이트. 대조: `go list ./...` → `140` 패키지 | 0, 0 | `ac006-go-vet.log`, `ac006-vet-pkgset.list` |

| AC-GTS2-006 tests | `go test -count=1 -timeout 600s ./internal/web/... ./internal/update/... ./internal/goal/... ./pkg/...` | `ok` 5줄: internal/web 19.295s, internal/update 7.519s, internal/goal 1.824s, pkg/models 0.356s, pkg/version 2.130s. `grep -c -E 'no test files|no tests to run|^FAIL|^---'` → `0` | 0 | `ac006-go-test.log` |

검증 분담: 레인은 `go test ./...`와 `internal/cli` 테스트를 돌리지 않았다. 전 패키지 판정은 리드가 develop tip에서 일괄로 돌리는 전체 실행과 `origin/develop` CI의 몫이다.

ID 대조 출력(원문):

```text
GO-2026-6218 new=0 baseline=2
GO-2026-6091 new=0 baseline=2
GO-2026-6090 new=0 baseline=2
GO-2026-6089 new=0 baseline=2
GO-2026-6088 new=0 baseline=2
GO-2026-5972 new=0 baseline=2
GO-2026-5856 new=0 baseline=2
GO-2026-5026 new=0 baseline=2
```

범위 밖 관측(판정하지 않음): 병합 트리 빌드에서도 `vcs.revision=2213871afb7d655f411c46a34fd38bb8153278fc`, `vcs.modified=true`가 찍힌다. ldflags의 Commit은 `180bc24f1`이다.
