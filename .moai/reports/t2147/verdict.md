auditor-model: claude-opus-5-5

# t2147 sync 단계 독립 감사 (iter1, Claude 단독 — Codex는 2단계)

verdict: PASS-WITH-DEBT
audited_sha: 8d1891b97a1ff899ed1342966b5555d65e8a9a29
score: 0.91

- 카드: t2147 (Class A, v3.2.1 버전 올림, 결정 d-20261106T174106Z-9fbc), SPEC 없음
- 트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2147`, 브랜치 `WT-version-bump-to`
- 감사 범위: `f9d8866951..8d1891b97a` (801eba9d92 버전 올림, 8d1891b97a 진행 기록)
- 반복 이력: 1회차 (이전 감사 없음)

## 판정 조건 — Codex 필수 게이트

`moai verify audit-plan` 결과 Codex 게이트가 `required`다 (`cross_model_required: true`, `enforced_required: ["codex"]`). 이 파일의 PASS-WITH-DEBT는 Claude 측 판정이고, Codex 리뷰가 통과해야 확정된다. Codex 영수증은 2단계에서 별도 파일 `.moai/reports/t2147/codex-receipt.md`에 기록하며, 이 파일은 그 뒤로 고치지 않는다. Codex가 지적을 내거나 실행되지 않으면 카드 전체 판정은 FAIL로 본다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | 버전 스탬프 7곳(version.go, system.yaml 2키, README 4개, hugo.toml) 모두 v3.2.1. `TestVersionStampRegistry` PASS — 등록부가 새 토큰 v3.2.1 기준으로 전 추적 파일을 쓸어 통과. `TestVersionSyncListNamesOnlyExistingPaths` PASS, `TestRoleRules*` PASS, `pkg/version` PASS |
| Security (25%) | 100/100 | PASS | 문자열 상수만 바뀜. 추가 줄 비밀값 grep 0건(exit 1), 입력 처리 경로 변경 없음 |
| Craft (20%) | 88/100 | PASS | `go build ./...` exit 0, `gofmt -l` 출력 없음, `go vet ./pkg/version/` exit 0. 감점: releaseDate가 확인되지 않은 추정값(F1) |
| Consistency (15%) | 92/100 | PASS | 이전 올림(a85be12449, eba919e44b)과 같은 파일 집합. 커밋 메시지 Conventional + 카드 id. 감점: version-management.md의 plugin-emit 절이 퇴역한 파이프라인을 가리킴(F3, 범위 밖) |

필수 통과 차원(Functionality, Security) 모두 기준 충족.

## 지적 사항

- F1 [Low] [optional] [in-diff] `docs-site/hugo.toml:56` — `releaseDate = "2026-11-07"`은 감사 당일 날짜일 뿐 확정된 출시일이 아니다. v3.2.0 올림은 CHANGELOG 섹션 날짜(2026-10-23)와 일치시켰는데, v3.2.1은 CHANGELOG 섹션도 태그도 아직 없다. docs-site는 main 착지 시 자동 배포되므로 태그 전에 v3.2.1이 공개 표기될 수 있다. 필요한 조치: 태그를 찍는 날 releaseDate를 실제 날짜로 맞추거나, 리더가 이 날짜를 수용한다고 기록한다. (부채 D1)
- F2 [Low] [optional] [outside] `CHANGELOG.md:8` — `## [3.2.1]` 섹션이 없다(`## [Unreleased]`만 있음). `scripts/release.sh:210-216`은 정식 버전 태그에서 이 섹션이 없으면 `die`한다. 카드 범위가 태그 없음·CHANGELOG 제외이고, 버전 올림 시험 중 이 섹션을 요구하는 것은 없다(아래 증거). 태그 전에 릴리스 산출물(CHANGELOG 섹션, `.moai/release-notes/v3.2.1.ko.md`)을 따로 써야 한다. 이 카드의 결함은 아니다.
- F3 [Info] [optional] [outside] `.moai/docs/version-management.md:146-150` — "매 올림마다 `make plugin-emit`" 절차를 적고 있으나, 이 트리에는 `plugins/` 디렉터리도 Makefile의 plugin 대상도 없다(플러그인 퇴역 이후 문서가 갱신되지 않음). 이 카드가 해당 절차를 건너뛴 것은 올바르다. 문서 정리는 별도 카드 몫이다.
- F4 [Info] [optional] [outside] README 배지가 올림 직전 `v3.1.2`였다. v3.2.0 올림(a85be12449)은 `v3.2.0`으로 썼으나 `git log -S 'Release-v3.1.2' -- README.md`에서 840053635f(t1720, #1932)가 `v3.1.2`를 다시 들여온 것으로 나온다. 이 카드가 v3.2.1로 바로잡았으므로 조치는 필요 없고, 회귀 경로로만 기록한다.

차단(blocking) 지적은 없다.

## 증거 (5절 형식)

### Claim
1. 버전 올림은 이전 올림과 같은 스탬프 집합을 모두 v3.2.1로 바꿨다.
2. 이번 올림에는 등록부 시험·role_rules 시험·CHANGELOG 변경이 필요 없다.
3. 지정된 시험이 모두 통과한다(TestReleasePreflight는 존재하지 않음).
4. batch FF 커밋 f05814ad66 위로 올려도 충돌이 없고 v3.2.1 보유 파일 집합이 같다.

### Evidence

카드 diff 범위:
```
$ git diff --stat f9d8866951..HEAD
 .moai/config/sections/system.yaml |  4 ++--
 .moai/reports/t2147/progress.md   | 19 +++++++++++++++++++
 README.ja.md                      |  2 +-
 README.ko.md                      |  2 +-
 README.md                         |  2 +-
 README.zh.md                      |  2 +-
 docs-site/hugo.toml               |  4 ++--
 pkg/version/version.go            |  2 +-
 8 files changed, 28 insertions(+), 9 deletions(-)
```

이전 올림 비교:
```
$ git show --stat a85be12449   (v3.2.0)
 .moai/config/sections/system.yaml           |  4 +-
 CHANGELOG.md                                |  2 +
 README.ja.md / README.ko.md / README.md / README.zh.md  | 2 +- 각각
 docs-site/hugo.toml                         |  4 +-
 internal/cli/version_stamp_registry_test.go | 83 +++++++++++++++++++++++++++--
 internal/hook/role_rules_fallback_test.go   |  3 +-
 pkg/version/version.go                      |  2 +-
$ git show --stat eba919e44b   (v3.1.3)
 .moai/config/sections/system.yaml | 4 ++--
 README.ja.md / README.ko.md / README.md / README.zh.md | 2 +- 각각
 pkg/version/version.go            | 2 +-
 6 files changed, 7 insertions(+), 7 deletions(-)
```
v3.2.0의 등록부 +83은 v3.2.0이 rc 시기에 67개 파일에서 인용돼 있었기 때문이고(등록부 주석 54-55행), role_rules 시험은 그때 리터럴을 `version.GetVersion()`으로 바꿔 이후 올림에서 손댈 필요가 없어졌다(`internal/hook/role_rules_fallback_test.go:123`).

시험 실행 (`export GOCACHE=.../scratchpad/gocache-lane28 && ...`, 한 번에 하나씩):
```
$ go test -count=1 ./pkg/version/
exit=0
ok  	github.com/modu-ai/moai-adk/pkg/version	0.096s

$ go test -count=1 -run TestVersionStamp ./internal/cli/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	7.418s

$ go test -count=1 -v -run TestVersionStampRegistry ./internal/cli/
=== RUN   TestVersionStampRegistryShape
--- PASS: TestVersionStampRegistryShape (0.00s)
=== RUN   TestVersionStampRegistry
--- PASS: TestVersionStampRegistry (8.56s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	10.772s

$ go test -count=1 -v -run TestReleasePreflight ./internal/cli/
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.643s [no tests to run]

$ go test -count=1 -v -run TestVersionSync ./internal/cli/
=== RUN   TestVersionSyncListNamesOnlyExistingPaths
--- PASS: TestVersionSyncListNamesOnlyExistingPaths (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.445s

$ go test -count=1 -run TestRoleRules ./internal/hook/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/hook	1.824s
```

빌드·형식·정적 검사·보안:
```
$ go build ./...                 → build exit=0 (출력 없음)
$ gofmt -l pkg/version/          → 출력 없음, exit=0
$ go vet ./pkg/version/          → vet exit=0
$ git diff f9d8866951..HEAD | grep -E '^\+' | grep -i -E 'token|secret|password|api[_-]?key|BEGIN .*PRIVATE'
secret-grep matches exit=1   (일치 없음)
```

v3.2.1 보유 파일(제외 그룹 밖) — 등록부가 쓸어보는 모집단:
```
$ git grep -l 'v3\.2\.1' | grep -v -E '^\.moai/(reports|specs|release-notes)/|^CHANGELOG\.md$|_test\.go$'
.moai/config/sections/system.yaml
README.ja.md
README.ko.md
README.md
README.zh.md
docs-site/hugo.toml
internal/bugreport/identity.go
internal/bugreport/spool.go
pkg/version/version.go
$ grep -n "internal/bugreport" internal/cli/version_stamp_registry_test.go
143:	{path: "internal/bugreport/identity.go", class: versionStampClassProse},
144:	{path: "internal/bugreport/spool.go", class: versionStampClassProse},
```
두 bugreport 파일은 이미 산문으로 등록돼 있어 새 등록이 필요 없다. 남은 `v3.2.0` 리터럴(74개 파일)은 주석·예시·골든 fixture(`.goreleaser.yml:54`, `design.yaml:3`, `internal/bugreport/spool.go:51-52`, `internal/feedback/testdata/*.golden` 등)로, 스탬프가 아니다.

템플릿:
```
$ grep -n "version" internal/template/templates/.moai/config/sections/system.yaml.tmpl
6:  version: "{{.Version}}"
9:  template_version: "{{.Version}}"
```
템플릿 수정이 없으므로 `make build`는 필요 없다.

CHANGELOG·릴리스 스크립트:
```
$ grep -n "^## \[" CHANGELOG.md | head -3
8:## [Unreleased]
31:## [3.2.0] - 2026-10-23
1084:## [3.1.3] - 2026-08-24
scripts/release.sh:216: die "CHANGELOG.md missing section '## [$CHANGELOG_VERSION]' ... Add release notes first."
```

batch FF 커밋 기준 사전 병합(ref 없이 트리만 계산):
```
$ git merge-base --is-ancestor f05814ad66 HEAD   → exit=1 (HEAD는 f05814ad66 위가 아님)
$ git merge-base --is-ancestor f9d8866951 f05814ad66 → exit=0
$ git diff --name-only f9d8866951 f05814ad66 | grep <스탬프 7개 + 등록부 시험>  → exit=1 (겹침 없음)
$ git merge-tree --write-tree --name-only f05814ad66 HEAD
31553d16c623190d3df52dfbfa74d91ea1f8650e      (충돌 줄 없음)
$ git grep -l 'v3\.2\.1' 31553d16c6... -- <제외 그룹 제외>
  → 위와 같은 9개 파일
```

감사 계획:
```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2147
"config_status": "ok", "model": "codex",
backends: claude gate off, codex gate required, glm gate off
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
```

### Baseline-attribution
- 모든 Go 측정은 이 트리(HEAD 8d1891b97a1ff899ed1342966b5555d65e8a9a29)의 소스를 이번 실행에서 컴파일해 얻었다(§2.2 첫째 방식 — 시험 바이너리가 곧 트리 빌드).
- `moai verify audit-plan`은 설치본 `/Users/goos/go/bin/moai`로 돌렸다. MCP 서버 배너는 v3.2.1-rc.2 (commit 04d8c27a7e)다. 이 실행은 workflow 설정만 읽었고 코드를 판정하지 않았다.

### Gaps
- `TestReleasePreflight`는 이 저장소에 없다(`no tests to run`). 버전 올림을 검증하는 별도 preflight 시험은 관측되지 않았다.
- 전체 시험 묶음은 돌리지 않았다(지시에 따라 범위 한정). CI가 맡는다.
- docs-site `hugo` 빌드는 돌리지 않았다. 바뀐 것은 hugo.toml 두 값뿐이다.
- 사전 병합 트리 31553d16c6에서 `TestVersionStampRegistry`를 실제로 돌리지 않았다. 보유 파일 집합이 같다는 것만 관측했다.
- `scripts/release.sh` 사전 검증은 실행하지 않았다(태그 범위 밖).
- 설치본 moai의 정확한 commit은 `moai version` 출력이 상자 형식이라 이번에 기록하지 못했다. 판정에는 쓰지 않았다.
- Codex 게이트는 아직 실행되지 않았다(2단계).

### Residual-risk
- f05814ad66 위로 착지할 때 그 사이 다른 카드가 제외 그룹 밖 파일에 `v3.2.1`(rc 표기 포함)을 새로 인용하면 등록부 시험이 깨질 수 있다. 현재 f05814ad66 기준으로는 해당 없음.
- releaseDate 추정값이 그대로 배포되면 docs-site가 실제와 다른 출시일을 보인다(F1).
- 태그 시점에 CHANGELOG 섹션이 없으면 release.sh가 멈춘다(F2) — 이 카드 밖이지만 출시 경로의 다음 단계에서 막힌다.

## 부채
- D1 (F1): 출시일 확정 시 `docs-site/hugo.toml` releaseDate 갱신 또는 수용 기록.

## Codex 영수증
Codex 리뷰와 영수증 id는 2단계에서 `.moai/reports/t2147/codex-receipt.md`에 따로 기록한다.
