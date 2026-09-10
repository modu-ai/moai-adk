# 싱크 감사 보고서 — SPEC-GO-TOOLCHAIN-SEC-002 (카드 t610, Tier S)

- 감사자: sync-auditor (독립 감사, 읽기 전용)
- 측정 트리: worktree `.claude/worktrees/t610`, 브랜치 `WT-go-1266`, HEAD `7f7bb9a35`
- merge-base `develop...HEAD`: `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1` (로컬 `develop` 은 `7beba0342` 로 이동했으나 merge-base 는 카드 base 그대로)
- 평가 프로필: SPEC 에 `evaluator_profile` 없음 → 내장 기본값 (Functionality 40 / Security 25 / Craft 20 / Consistency 15, must-pass = Functionality + Security)
- 교차 모델 감사: `audit_model` 설정 없음. MCP 백엔드는 호출하지 않았다 — codex/glm 의 `baseBranch` 대상은 원격 기본 헤드(main)로 해석돼 이 카드의 develop 기준 diff 를 리뷰하지 못하고, `uncommittedChanges` 는 미추적 파일 1개뿐이다. Gap 으로 기록한다.

## 판정

**Overall Verdict: PASS** — blocking finding 0건, must-pass(Functionality·Security) 모두 통과.

| Dimension | Score | Verdict | Evidence (이번 감사에서 직접 실행) |
|-----------|-------|---------|-----------------------------------|
| Functionality (40%) | 93/100 | PASS | AC-001..008 전부 판별력 있는 증거 위에 PASS. 재측정: `go -C <root> version` → `go version go1.26.8 darwin/arm64` exit 0; `go env GOTOOLCHAIN GOMOD` → `auto` / `…/t610/go.mod`; `sed -n 3p go.mod` → `go 1.26.8`; `grep -c '^toolchain' go.mod` → `0`; `grep -c '1\.26\.4'` 4개 문서 → `0` 각각; `grep -c '1\.26\.8'` → `2,1,1,1` (합 5). M4 재측정·CI 판정은 리드 몫(대기) |
| Security (25%) | 90/100 | PASS | 기준선 `govulncheck-auto.log` 에서 8개 ID 각 `2` 건 (대조군 8/8 적중) vs 판정 로그 `0` 건 ×8; `diff baseline/govulncheck-go1.26.8.log run/ac004-govulncheck.log` → exit 0 (바이트 동일); 비증거 diff 25개 변경 줄에 비밀 패턴 grep → `0` (exit 1). govulncheck 재실행은 하지 않음(Gap) |
| Craft (20%) | 84/100 | PASS | 증거 계약 준수(모든 판정 명령에 `.exit`, AC-001 캡처 17:22:37 이 AC-002..007 캡처보다 먼저). 증거 위생 결함 3건(F2·F3·F4). Go 코드 변경 없음 → 커버리지 해당 없음 |
| Consistency (15%) | 82/100 | PASS | 범위 보완 diff 0 바이트, sync 5경로 numstat = sync 커밋 numstat, 커밋 제목 Conventional + 카드 id. codemaps 헤더 귀속 모순(F1) |

- 가중 조화평균: **88.6** / 비가중 조화평균: 87.0

## Findings (구조화된 결함 목록)

| ID | Severity | Blocking | 위치 | 결함 | 증거 | 권장 수정 |
|----|----------|----------|------|------|------|-----------|
| F1 | should-fix | non-blocking | `.moai/project/codemaps/overview.md:3-7`, `.moai/project/codemaps/modules.md:3-7` | 헤더가 "모든 수치는 아래 트리(t592, HEAD `e7bd89ee3`)에서 직접 잰 것이고, 다른 트리·다른 시점에서 옮겨온 값은 없습니다"라고 단언하는데, 바로 그 헤더 블록의 `**Go**: 1.26.8` 은 그 트리에서 잰 값이 아니다. 헤더가 스스로 금지한 "옮겨온 값"이 됐다(측정되지 않은 값 주장). 값 자체는 현재 트리에서 참이므로 SPEC 요구(REQ-GTS2-008)의 정확성에는 영향이 없다. 원인은 plan.md § F M5 가 토큰만 바꾸라고 지시하면서 이 헤더 계약을 고려하지 않은 데 있다 | `git show e7bd89ee3:go.mod \| sed -n 3p` → `go 1.26.4`; `merge-base --is-ancestor e7bd89ee3 HEAD` → exit 0 | Go 토큰 옆에 귀속 한 줄을 붙인다(예: "Go 버전만 t610 `41f445fa5` 의 `go.mod:3` 기준으로 갱신"). 또는 다음 `/moai codemaps` 실행에서 재측정해 헤더를 새 트리로 갱신한다 |
| F2 | advisory | non-blocking | `.moai/reports/t610/sync/ac008-evidence.md` (Control 절, Files 목록) | 증거 파일 `control-git-show.exit` (`0`) 을 인용하지만 그 파일이 없다. 해석되지 않는 증거 경로는 귀속되지 않은 주장이다(VCI §2) | `ls ../sync/control-git-show.exit` → `No such file or directory` | 인용을 지우거나, `git show` exit 를 실제 파일로 남긴다 |
| F3 | advisory | non-blocking | `.moai/reports/t610/sync/ac008-evidence.md` (Control 절 끝 문단) | "spec.md § D.0 E-08" 인용이 남아 있다. D.0 원장은 acceptance.md 에 있다. 백필 커밋 `7f7bb9a35` 은 progress.md 쪽만 고쳤다 | `git show --stat 7f7bb9a35` → `progress.md` 1개 파일만 변경 | 인용을 `acceptance.md § D.0 E-08` 로 고친다 |
| F4 | advisory | non-blocking | `.moai/reports/t610/sync/sync-paths-numstat.txt` | 파일이 커밋되지 않았다(미추적). 이 파일은 sync 커밋 단독의 numstat 이며, § D.0b step 5 의 M4 비교 증거가 아니다 | `git status --short` → `?? .moai/reports/t610/sync/sync-paths-numstat.txt` | M4 증거와 함께 커밋하거나, M4 파일로 대체된다는 점을 명시한다 |
| F5 | advisory | non-blocking | 카드 커밋 6개 본문 전부 | `🗿 MoAI` 줄이 마지막 문단에 있어 git 기본 trailer 파서가 `Authored-By-Agent` 를 인식하지 못한다. 소유권 lint 는 자체 정규식으로 파싱하므로 lint 판정에는 영향이 없다 | `git log --format='%h [%(trailers:key=Authored-By-Agent,valueonly)]'` → 6개 전부 `[]`; `internal/spec/lint_ownership.go:256` 정규식 `(?mi)^\s*Authored-By-Agent:` | 저장소 전반의 관례 문제다(카드 결함 아님). `Authored-By-Agent` 를 마지막 문단으로 옮기는 관례를 별도 카드로 검토한다 |
| F6 | advisory | non-blocking | `.moai/reports/t610/run/ac004-govulncheck-env.txt` | 판정 스캔 증거에 `GOTOOLCHAIN`·`GOMOD` 만 있고 `GOVERSION` 이 없다. 스캔 시점의 유효 툴체인은 AC-001(같은 트리, 약 36초 전 캡처)과 기준선 go1.26.8 로그와의 바이트 일치로 추론된다. AC 문구는 충족한다 | 캡처 시각 `ac001 17:22:37`, `ac004-env 17:23:13`; 위 `diff` exit 0 | M4 재측정 때 같은 파일에 `go env GOVERSION` 도 기록한다 |
| F7 | advisory | non-blocking | `CHANGELOG.md:12` | "reachable from the `go.mod` build directive" 는 부정확하다. govulncheck 의 도달성은 코드 호출 경로 기준이지 지시어 기준이 아니다. 악용 가능성·배포 바이너리 주장은 없다 | 항목 본문 | "reported by govulncheck against the go1.26.4 standard library" 정도로 다듬는다 |

## 질문별 판단

1. **AC 판별력** — 모두 공허하지 않다.
   - AC-004: 같은 grep 을 기준선 로그에 돌리면 ID마다 `2` 건(8/8 적중), 새 로그에서는 `0` 건. `vulnerabilities in modules you require` 대조군은 `1` 건이다.
   - AC-006: `go list` 로 선택된 패키지가 정확히 5개이고 `ok` 가 5줄, `no test files|no tests to run|^FAIL|^---` 은 `0` 건이다.
   - AC-007: 제외 경로를 뺀 보완 diff 가 M1·M3 모두 0 바이트이고, 이번 감사에서 HEAD 에 § D.0b 형태로 재실행해도 빈 출력이다. 제외하지 않은 전체 diff 에는 경로 87개가 나온다.
   - AC-008: 편집 전 `403bac94b:.moai/project/product.md` 에서 `2` 건(exit 0)이 나와 grep 이 토큰을 읽는다는 것이 확인된다.
2. **범위** — `git diff --name-only develop...HEAD` 에서 SPEC 디렉터리와 `.moai/reports/t610` 을 뺀 나머지는 정확히 `go.mod`, `CHANGELOG.md`, `product.md`, `structure.md`, `codemaps/overview.md`, `codemaps/modules.md` 다. `go.sum` 은 없다.
3. **검증 분담의 정직성** — `go test ./...` 나 `internal/cli` 테스트를 실행했다고 주장하는 산출물은 없다. `run/commands.md:59`, progress.md §E.2·§E.3, 커밋 `741ab8a38` 본문이 모두 "실행하지 않음"을 명시한다. `ac006-vet-pkgset.list` 에 `internal/cli` 가 보이는 것은 `go vet ./...` 의 패키지 집합이지 테스트 실행이 아니다.
4. **CHANGELOG** — ID 8개와 순서가 기준선과 일치한다. 악용 가능성·배포 바이너리 취약 주장은 없고, 오히려 명시적으로 부인한다. `GOTOOLCHAIN=local` 문장은 plan.md § D1 잔여 위험과 일치하며 옳다. `### Fixed` 배치는 이 파일에 `### Security` 헤딩이 한 번도 없다는 관례와 맞는다(`grep -c '^### Security'` → `0`). 문구 하나만 부정확하다(F7).
5. **codemaps 헤더** — 측정되지 않은 값 주장이다(F1, should-fix). 값이 틀린 것은 아니고 귀속이 틀렸으므로 blocking 은 아니다.
6. **소유권·수명주기**
   - 상태 전이 커밋과 trailer:
     - `e6158cb36` `(none)→draft` — manager-spec
     - `41f445fa5` `draft→in-progress` — manager-develop
     - `4516fbe40` `in-progress→completed` — manager-docs
   - 두 전이의 frontmatter diff 는 `status:` 한 줄씩뿐이다.
   - `sync_commit_sha: 4516fbe40` 은 `7f7bb9a35` 에서 백필됐다.
   - `moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` 결과는 `✓ No findings` (exit 0)이고, `--json` 은 `[]` 이다.
   - lint 바이너리 `84fa4ece4` 는 trailer 누락 Info 룰 커밋 `5caddeb2d` 를 포함한다(`merge-base --is-ancestor` exit 0). 따라서 `[]` 는 "측정 불가"가 아니라 "trailer 인식됨"을 뜻한다.
7. **범위 밖** — `bin/moai` 의 vcs.revision 불일치는 판정하지 않았다.

## 5절 증거 보고

### Claim
SPEC-GO-TOOLCHAIN-SEC-002 의 run·sync 산출물은 AC-GTS2-001..008 을 판별력 있는 증거로 충족한다. 카드 diff 범위는 SPEC 이 허용한 경로 안에 있다. 소유권 전이는 올바른 커밋과 trailer 에 실렸다. blocking 결함은 없다.

### Evidence (이번 감사 실행, 원문)
```
$ go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610 version
go version go1.26.8 darwin/arm64            (exit=0)
$ go -C <root> env GOTOOLCHAIN GOMOD
auto
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610/go.mod   (exit=0)
$ sed -n 3p go.mod
go 1.26.8
$ grep -c '^toolchain' go.mod
0                                            (grep_exit=1)
$ env | grep -E '^(GOTOOLCHAIN|GOFLAGS)='
                                             (exit=1, 설정 없음)
$ git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610' ':!CHANGELOG.md' ':!.moai/project/product.md' ':!.moai/project/structure.md' ':!.moai/project/codemaps/overview.md' ':!.moai/project/codemaps/modules.md'
                                             (빈 출력)
$ git diff --numstat develop...HEAD -- CHANGELOG.md .moai/project/product.md .moai/project/structure.md .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md
1	1	.moai/project/codemaps/modules.md
1	1	.moai/project/codemaps/overview.md
2	2	.moai/project/product.md
1	1	.moai/project/structure.md
1	0	CHANGELOG.md
$ git show --numstat --format= 4516fbe40 -- (같은 5경로)
(위와 동일한 5줄)
$ grep -c '1\.26\.4' product.md structure.md codemaps/overview.md codemaps/modules.md
overview.md:0  modules.md:0  structure.md:0  product.md:0
$ grep -c '1\.26\.8' (같은 4개)
product.md:2  structure.md:1  overview.md:1  modules.md:1
$ diff .moai/reports/t610/baseline/govulncheck-go1.26.8.log .moai/reports/t610/run/ac004-govulncheck.log
                                             (diff_exit=0)
$ git show e7bd89ee3:go.mod | sed -n 3p
go 1.26.4
$ moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002
✓ No findings — all SPEC documents are valid   (lint_exit=0)
$ moai spec lint --json .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002
[]
$ git merge-base --is-ancestor 5caddeb2d 84fa4ece4
                                             (ancestor_exit=0)
$ git status --short
?? .moai/reports/t610/sync/sync-paths-numstat.txt
```
레인 증거 원문(판독만): `run/ac004-ids.txt` 에서 ID 8개 각 `0`, `run/ac004-ids-control.txt` 에서 각 `2`; `run/ac004-govulncheck.log.exit` → `govulncheck_exit=0`; `run/ac005-go-version-m.txt:6` → `bin/moai: go1.26.8`; `run/ac005-make-build.log.exit` → `make_build_exit=0`; `run/ac006-go-test.log` 에 `ok` 5줄; `run/ac006-swept.txt` 의 빈 스윕 grep → `0`.

### Baseline-attribution
- 이번 감사의 모든 명령은 이 실행에서 worktree `t610` HEAD `7f7bb9a35` 에 대해 돌렸다.
- 레인 run 증거는 M1 커밋 `41f445fa5` (tree `e6ec2991e`)에서 캡처됐다(파일 헤더 기준). 이후 커밋은 Go 코드를 바꾸지 않으므로 그 증거는 현재 HEAD 의 Go 트리에 여전히 적용된다.
- sync AC-008 증거는 커밋 전 작업 트리(HEAD `403bac94b`)에서 캡처됐고, 이번 감사가 HEAD `7f7bb9a35` 에서 같은 결과를 재관측했다.
- lint 판정은 `/Users/goos/go/bin/moai` v3.2.0-rc.5 (`84fa4ece4`) 로 얻었다.

### Gaps
- govulncheck 를 재실행하지 않았다. 레인 스캔 이후 게시된 advisory(EC-2)는 관측하지 않았다. M4 재측정이 이를 덮는다.
- `make build`, `go vet`, `go test` 는 재실행하지 않았다(지시된 제약). AC-005·006 은 레인 증거 판독에 의존한다.
- M4(흡수 후 재측정, § D.0b step 5 비교, 단독 통합 창)와 `origin/develop` CI 판정은 아직 없다. 리드 몫이다.
- codex/glm 교차 모델 감사는 실행하지 않았다(대상 diff 해석 불가, 위 참조).
- 레인이 M1 에서 AC-003 을 실제로 `develop...HEAD` 형태로 캡처했는지는 `run/commands.md` 기록에 의존한다. 이번 감사에서는 go.mod diff(단일 hunk `-go 1.26.4`/`+go 1.26.8`, 1+/1-)를 재관측했다.

### Residual-risk
- 통합 후 CI 매트릭스(darwin/windows/linux)가 go1.26.8 에서 깨질 수 있다. 레인은 darwin/arm64 에서 5개 패키지만 돌렸다.
- `GOTOOLCHAIN=local` 이면서 옛 Go 가 설치됐거나 툴체인 프록시에 닿지 못하는 환경은 빌드에 실패한다(plan.md § D1 에 기록된 위험, 범위 밖).
- go1.26.7 net/http #80927 의 실제 영향은 저장소 grep(h2c 설정 없음)에만 근거하며 측정되지 않았다(plan.md § D2 Gap).
- F1 이 방치되면 codemaps 헤더의 "옮겨온 값 없음" 단언이 틀린 채로 남는다. 다음 codemaps 재생성 전까지 이 문서의 다른 수치에 대한 신뢰도가 떨어진다.

## Recommendations
- F1 은 M4 전에 카드 브랜치의 새 커밋(manager-docs)으로 귀속 한 줄을 붙이는 방식이 가장 싸다. 이 경우 § D.0b step 5 비교 대상 "M5 sync 커밋" numstat 과 카드 쪽 numstat 이 달라지므로, 리드가 비교 기준을 두 커밋 합으로 명시해야 EC-5 오탐이 나지 않는다. 그 비용이 크면 F1 을 후속 codemaps 재생성 카드로 넘긴다.
- F2·F3·F4 는 증거 파일만 고치는 일이다. M4 증거 커밋 때 함께 정리하면 된다.
- M4 재측정 파일에 `go env GOVERSION` 을 추가한다(F6).
