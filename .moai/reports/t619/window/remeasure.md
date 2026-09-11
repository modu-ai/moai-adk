# t619 통합 창 재측정 — 로컬 develop 흡수 뒤 병합 트리

- 카드: t619 · SPEC-TOOLPOLICY-DRIFT-GUARD-001 · 레인 lane-7
- 창: `moai integration acquire --name lane-7` → `release-integration window acquired by 6a7596db-e474-4923-82f9-ff16499a11da on WT-toolpolicy-drift` (exit 0, 설정 드리프트 적중 없음)
- 측정일: 2026-09-11

## 1. 주장

로컬 develop `2b6c82f40` 을 흡수한 병합 트리(흡수 병합 커밋 `c74340c42`, 트리 `a52b4835ac16f7b0bcc38bd735748ae7fff28add`)에서 Go 1.26.8 로 다시 잰 결과, 이 카드의 검사는 모두 통과하고 AC-TDG-009 도 이제 통과한다. 적용 권한 파일은 develop 과 바이트 동일하다.

## 2. 증거

| 항목 | 명령 | 관측 |
|---|---|---|
| 흡수 대상 최신성 | `git fetch origin develop` 후 `git rev-list --count --left-right origin/develop...develop` | `0	9` (로컬 develop 이 원격보다 9 앞섬, 뒤처짐 0 → 로컬 develop 흡수) |
| 흡수 | `git merge develop` | `CONFLICT (content): Merge conflict in CHANGELOG.md` 1건. `[Unreleased] / ### Fixed` 첫 줄에 t619·t610 두 항목이 겹침 → 두 항목 모두 유지(t619 먼저). 해결 뒤 충돌 표시 `0`, 두 SPEC 링크 각 `1`, `git diff --name-only --diff-filter=U` 출력 없음 |
| 흡수 뒤 분기 | `git rev-list --count --left-right develop...HEAD` | `0	20` |
| 툴체인 | `go version`, `go env GOTOOLCHAIN` | `go version go1.26.8 darwin/arm64`, `auto` |
| toolpolicy 패키지 | `go test -count=1 -cover -v ./internal/config/toolpolicy/...` (환경변수 scrub) | `toolpolicy_rc=0`, `--- PASS` `70`, `--- FAIL`/`--- SKIP` `0`, `coverage: 89.1% of statements` — `toolpolicy-test.txt` |
| 드리프트 검사 | `make tool-policy-drift-check` | `make_rc=0`, `ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.365s` — `drift-check.txt` |
| AC-TDG-009 (리드 승인 1회) | `go test -count=1 -timeout 600s ./internal/cli/ -run 'TestToolPolicyList_QueryFilters' -v` (환경변수 scrub) | `cli_rc=0`, 하위 6개 PASS(`filter_ask` 포함), `ok  	github.com/modu-ai/moai-adk/internal/cli	1.293s` — `cli-queryfilters.txt` |
| vet | `go vet ./internal/config/toolpolicy/...` | `vet_rc=0` — `vet.txt` |
| lint | `golangci-lint run --timeout=3m ./internal/config/toolpolicy/...` | `lint_rc=0`, `0 issues.` — `lint.txt` |
| 적용 권한 불변 | `git diff --exit-code --stat develop HEAD -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` | 출력 없음, exit 0 |
| 카드 기여 범위 | `git diff --stat develop HEAD` | 13 파일, 2189 추가 / 108 삭제. `.claude/settings.json` 없음 |

## 3. 기준선 귀속

- 측정 트리: `WT-toolpolicy-drift` 흡수 병합 커밋 `c74340c42` (sync 커밋 `aa40907ae`, SHA 기록 `3465af90a` 위에 로컬 develop `2b6c82f40` 흡수).
- 툴체인: go1.26.8 darwin/arm64 (GOTOOLCHAIN=auto, `go.mod` 의 `go 1.26.8`).
- 이 문서를 담는 증거 커밋은 코드와 설정을 바꾸지 않으므로 위 측정 대상 파일은 증거 커밋 뒤에도 같다.

## 4. 갭

- 전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다. 판정은 리드 일괄 push 뒤 `origin/develop` CI 가 낸다.
- darwin/arm64 한 플랫폼만 쟀다. windows·linux 는 CI 몫이다.
- `internal/cli` 는 이름 지정 테스트 1개만 돌렸다(리드 승인 범위).

## 5. 잔여 위험

- develop 에 다른 레인 병합이 먼저 들어오면 이 흡수 뒤 트리와 달라진다. 병합 직후 develop 병합 커밋 트리와 이 브랜치 tip 트리의 동일성으로 확인한다.
- AC-TDG-009 는 progress.md 에 FAIL(선재 결함 귀속)로 남아 있다. 이 문서는 t660 흡수 뒤 PASS 로 바뀌었다는 관측만 기록하며 SPEC 산출물은 고치지 않는다.
