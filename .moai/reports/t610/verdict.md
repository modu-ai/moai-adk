# t610 판정 기록 — Go 빌드 기준 1.26.4 → 1.26.8

- 카드: t610 (Go 전수 점검 F01 · 증거 항목 SEC-01 · P1 보안)
- SPEC: `SPEC-GO-TOOLCHAIN-SEC-002` (Tier S, Class C) — status `completed`
- 워크트리: `.claude/worktrees/t610`, 브랜치 `WT-go-1266`
- 카드 기준: 로컬 develop `d3b7d438d` (merge-base, 작성 시점까지 변동 없음)
- 작성: lane-8, 2026-09-10. 이 문서는 레인의 판정 기록이다. 병합 판정과 전 패키지 판정은 리드의 몫이다.

## 주장

`go.mod` 3행의 `go` 지시어를 `go 1.26.4`에서 `go 1.26.8`로 올렸다. 그 결과 govulncheck가 보고하던 표준 라이브러리 취약점 8개(GO-2026-6218, 6091, 6090, 6089, 6088, 5972, 5856, 5026)가 더 이상 보고되지 않는다. 다른 소스·워크플로 파일은 바꾸지 않았다.

## 증거

| 항목 | 명령 | 관측 출력 | 종료 코드 | 파일 |
|---|---|---|---|---|
| 기준선(1.26.4) | `GOTOOLCHAIN=auto GOMAXPROCS=2 govulncheck ./...` | `Your code is affected by 8 vulnerabilities from the Go standard library.` | 3 | `baseline/govulncheck-auto.log` |
| 대조(1.26.6) | 같은 트리, `GOTOOLCHAIN=go1.26.6` | `Your code is affected by 0 vulnerabilities.` | 0 | `baseline/govulncheck-go1.26.6.log` |
| 도구 체인 게이트 | `go -C <worktree> version` | `go version go1.26.8 darwin/arm64` | 0 | `run/ac001-toolchain.txt` |
| go.mod 변경 | `git diff --numstat develop...HEAD -- go.mod` | `1	1	go.mod` | 0 | `run/ac003-numstat.raw` |
| 취약점 재검사 | `GOMAXPROCS=2 timeout 590 govulncheck ./...` (override 없음) | `Your code is affected by 0 vulnerabilities.` | 0 | `run/ac004-govulncheck.log` |
| ID 부재 | 8개 ID 각각 `grep -c` | 새 로그 8개 모두 `0`, 기준선 로그 각 `2` | 1 ×8 / 0 ×8 | `run/ac004-ids.txt`, `run/ac004-ids-control.txt` |
| 빌드 | `make build` → `go version -m bin/moai` | `bin/moai: go1.26.8` | 0, 0 | `run/ac005-*.txt` |
| 정적 분석 | `go vet ./...` | 진단 0건(140 패키지) | 0 | `run/ac006-go-vet.log` |
| 대표 테스트 | `go test -count=1 -timeout 600s ./internal/web/... ./internal/update/... ./internal/goal/... ./pkg/...` | `ok` 5줄 | 0 | `run/ac006-go-test.log` |
| 변경 범위 | 제외 경로 여집합 `git diff --name-only develop...HEAD` | 0바이트 | 0 | `run/ac007-complement-m3.raw` |
| 문서 표기 | 문서 4개 `grep -c '1\.26\.4'` / `'1\.26\.8'` | 0 / 합계 5, 대조(편집 전 product.md) 2 | 1 / 0 | `sync/ac008-evidence.md` |
| plan 감사 | plan-auditor | 1회차 FAIL(blocking 3) → 2회차 PASS(0.86) | — | `plan-audit.md`, `plan-audit-2.md` |
| sync 감사 | sync-auditor | PASS, 가중 조화평균 88.6, blocking 0 | — | `sync-audit.md` |

경로는 모두 `.moai/reports/t610/` 기준이다. 명령 원문 색인은 `baseline/commands.md`, `run/commands.md`에 있다.

## 기준 귀속

- 모든 측정은 2026-09-10, 이 워크트리에서 레인이 직접 했다. 기준선은 `d3b7d438d`, run 측정은 M1 커밋 `41f445fa5`(tree `e6ec2991e`)에서 했다.
- 도구 체인 확보 방식: 설치된 brew go는 1.26.0이다. 그런데 `go env GOROOT`를 보면 이미 go.mod 지시어로 toolchain 모듈 캐시를 자동 전환해 쓰고 있었다. 1.26.6과 1.26.8은 on-demand 다운로드로 받아졌다. CI의 모든 workflow는 `setup-go`의 `go-version-file: go.mod`를 읽는다. 따라서 지시어 한 줄이 로컬과 CI를 함께 움직인다.

## 검증 분담 (리드 규칙 2026-09-10)

- **레인이 한 것:** 도구 체인 전환이 빌드를 깨지 않는다는 확인만 했다. `go build`(`make build`), `go vet ./...`, govulncheck, 대표 패키지 5개의 테스트다.
- **레인이 하지 않은 것:** `go test ./...`와 `internal/cli` 테스트는 돌리지 않았다. go 지시어는 모든 패키지에 걸리지만, 전 패키지 판정은 리드가 develop tip에서 일괄로 돌리는 전체 실행과 `origin/develop` CI의 몫이다.
- push와 바이너리 설치(`~/go/bin/moai`)는 하지 않았다.

## 확인하지 못한 것

- 전 패키지 테스트, darwin/arm64 이외 OS 매트릭스, golangci-lint.
- go1.26.7 #80927(암호화하지 않은 HTTP/2 전환 뒤 ReadHeaderTimeout)의 영향. 저장소에 h2c 설정이 없다는 것은 grep으로 확인했다. 다만 `internal/web` 테스트는 그 경로를 검사하지 않는다.
- 레인 스캔 이후에 새로 게시된 advisory.
- 이미 배포된 바이너리의 취약 여부. 심벌 도달성은 공격 재현이 아니다.

## 잔여 위험

- **M4 흡수:** develop을 흡수할 때 `require` 변경이 들어오면 govulncheck 결과가 달라질 수 있다. M4에서는 병합 트리에서 AC-001~008을 다시 재야 한다.
- **GOTOOLCHAIN=local 환경:** `GOTOOLCHAIN=local`이거나 toolchain 프록시에 닿지 못하는 환경은 "go.mod requires go >= 1.26.8"로 실패한다.
- **착지 영향:** 병합 뒤에는 모든 레인의 빌드 도구 체인이 1.26.8로 바뀐다. 그래서 단독 창에서 착지하고, 착지 뒤 모든 레인이 develop을 다시 흡수해야 한다.

## 보류 사항 (리드 결정)

- **sync-audit F1 (should-fix):** codemaps 두 문서의 머리말은 "다른 트리(t592 `e7bd89ee3`)에서 모든 값을 쟀다"고 적는다. 그 아래 Go 1.26.8은 그 트리에서 잰 값이 아니다.
- **sync-audit F7 (advisory):** CHANGELOG의 "reachable from the go.mod build directive" 문구가 부정확하다. 도달성은 코드 호출 경로 기준이다.
- 두 건을 지금 고치면 sync 경로 numstat이 sync 커밋 단독 값과 달라져 M4 비교 기준이 바뀐다. 그래서 고치지 않고 보류했다.
- **범위 밖 관측:** `go version -m bin/moai`의 `vcs.revision=2213871afb7d655f411c46a34fd38bb8153278fc`는 primary 체크아웃 main HEAD다. ldflags의 Commit은 `41f445fa5`로 맞다. 원문은 `run/ac005-go-version-m.txt`에 있다. 원인은 재지 않았다.
