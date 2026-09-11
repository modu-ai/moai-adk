# t525 로컬 develop 흡수 기록

트리: `.claude/worktrees/t525`, 브랜치 `WT-speclint-red`. 흡수 전 HEAD `7151990276d171feada1aa9e18e582279fa19c4b`. 창은 필요 없는 흡수(카드 브랜치 안)이며, 리드 판정으로 진행했다. 흡수 전 판독은 `../absorb-precheck/summary.md`.

## 흡수 두 번

| 차수 | 흡수 대상(로컬 develop) | 병합 커밋 | 부모 | 트리 |
|---|---|---|---|---|
| 1 | `f6b121a9ccd27e9962ad5c0f48b420ef8b4c528b` | `e7522225ef7eaa56eba12acabcbd8cf2faf92bf9` | `715199027` · `f6b121a9c` | `5201183928cedc4d30621c5c0140e8e82f88bd62` |
| 2 | `81c1d58f9cf7045594ee61d5e4ff380948ce9eba` | `42c17b5b697f73696cf511e41e4965f82106aadb` | `e7522225e` · `81c1d58f9` | `3b9e8016ff875607b20d4ad3857e143c09af4ead` |

- 1차 직전 develop 이 판독 시점(`1412c5738`)에서 `f6b121a9c` 로 움직였다. 델타는 t247 보고서와 `internal/web` 뿐이고, 충돌 예측은 판독 때와 같았다(`spec-lint.yml` content, `spec_lint_test.go` add/add). 새 tip 에서 규칙 코드 집합을 다시 모았더니 34개로 같았다(추가·삭제 0).
- 2차는 리드 알림(로컬 develop 이 t464·t661 lint 수리로 이동)에 따라 측정 전에 했다. 델타는 t464·t661 보고서와 `internal/cli/main_test.go` 뿐이고 `git merge-tree` 예측 트리가 실제 병합 트리 `3b9e8016` 와 같았다. 충돌 없음.

## 충돌 해결 (1차)

### `.github/workflows/spec-lint.yml` — 리드 결정 A안

- develop 의 `Select SPEC lint policy` 단계와 `Run error-only SPEC lint` 단계를 유지했다.
- strict 쪽 단계는 명령만 `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` 로 바꾸고, 이름을 `Run baseline-gated SPEC lint` 로 바꿨다. 조건 `steps.policy.outputs.strict == 'true'` 는 그대로다.
- t525 의 트리거 경로 추가(`pull_request`·`push` 두 목록의 `.moai/spec-lint-baseline.json`)는 자동 병합으로 들어왔다.
- 선택 단계의 `case` 패턴에 `.moai/spec-lint-baseline.json` 을 더했다. 기준선만 바꾼 develop push 가 error-only 로 빠져 기준선을 읽지 않는 구멍을 막기 위해서다.
- error-only 단계 주석의 "Strict mode remains mandatory" 를 "The baseline gate remains mandatory … the baseline file …" 로 고쳤다. 명령이 바뀌어 옛 문구가 사실과 달라졌기 때문이다.
- 확인: 충돌 표지 0, `git diff --check` exit 0. Ruby YAML 로드로 단계 목록과 트리거를 읽었다 — 단계 6개(checkout, setup-go, Fetch integration refs, Select SPEC lint policy, Run baseline-gated SPEC lint, Run error-only SPEC lint), 두 트리거 경로 목록 모두 `.moai/spec-lint-baseline.json` 포함, push 브랜치 `main,develop`, case 줄 `.moai/specs/*|internal/spec/*|.moai/spec-lint-baseline.json) spec_changed=true ;;`.

### `internal/cli/spec_lint_test.go` — 합집합

- 원본 확인: `git rev-parse :2:` = `92f586a05`, `:3:` = `7d556ae0a`. 작업에 쓴 사본의 `git hash-object` 가 두 값과 같다.
- 방법: import 두 블록의 합집합 12개를 경로순으로 정렬하고, 본문은 develop 쪽 다음에 브랜치 쪽을 그대로 이어 붙였다.
- 확인: 최상위 함수 41개 = develop 18 + 브랜치 23, 이름 중복 0, 두 원본 본문이 병합 파일 안에 바이트 그대로 들어 있음(각 1), 충돌 표지 0, `gofmt -l` 출력 없음, `git diff --check` exit 0.
- 컴파일: 경쟁 `go (test|build|vet)` 프로세스 없음(pgrep exit 1) 확인 뒤 `go test -count=1 -run '^$' ./internal/cli/` → `ok … 1.037s [no tests to run]`, exit 0. 테스트는 돌리지 않았다.

## 흡수 트리 측정 (트리 `3b9e8016`)

- 판정 바이너리: 이 트리(HEAD `42c17b5b6`)에서 `go build -o <세션 스크래치>/moai-t525-absorb2 ./cmd/moai` 로 빌드해 경로로 호출했다. 설치본은 쓰지 않았다.
- 실행 위치: 워크트리 루트. 부하 10:36 7.34 → 11:00 23.60(다른 세션 포함).

| 실행 | 파일 | exit | 요약 줄 |
|---|---|---|---|
| `spec lint` | `lint-default.txt` | 0 | `0 error(s), 3133 warning(s)` |
| `spec lint --strict` | `lint-strict.txt` | 0 | `0 error(s), 3133 warning(s)` |
| `spec lint --baseline .moai/spec-lint-baseline.json` | `lint-baseline.txt` | 0 | `baseline: OK` · `inventory: 3133 warning(s) total (advisory included), 0 non-advisory tracked across 0 recorded rule(s)` · `recorded at: 9e1744469 (2026-09-08)` |
| `spec lint --json` | `lint.json` (stderr 0바이트) | 0 | 발견 3336 |

`census.txt` (JSON 집계):

- 심각도: info 203, warning 3133, error 0. **advisory 가 아닌 warning 0.**
- `DuplicateAcceptanceID` 발견 **0**. 전 era 에서 0 이므로 V3R6 대상도 0.
- 출력된 코드 15개, 전부 warning+advisory 이거나 info(`OwnershipTransitionUnmeasured` 203)다: CoverageIncomplete 2010, FrontmatterInvalid 14, InvalidREQID 6, LegacyEARSKeyword 48, MissingExclusions 27, ModalityMalformed 180, ModalityUnjudged 514, MovingRefUnpinned 116, OwnershipTransitionInvalid 1, REQTableRowsRejected 84, StatusGitConsistency 18, StatusTokenUnrecognized 7, StatusTransitionInvalid 102, SyncSHASlotFormat 6.
- 출력 코드 15개 중 소스 수집 집합(34개)에 없는 코드 0 → 흡수 전 판독의 "상수 경유 선언 누락" Gap 은 출력된 코드에 대해서는 닫혔다.

판독:

- 리드 조건 "`DuplicateAcceptanceID` 인구 0 이면 M3.4 진입" 충족.
- `--strict` 와 `--baseline` 이 exit 0 인 이유는 흡수 트리의 warning 이 전부 advisory 이기 때문이다. 기준선 파일의 `rules {}` 가 비어 있어도 초과할 non-advisory 규칙이 없다.
- 로컬 기본 실행(= CI 의 error-only 단계와 같은 명령)의 녹색 출력에 경고 총수 줄 `0 error(s), 3133 warning(s)` 이 있다.

## 미검증

- CI 에서 error-only 단계·baseline 단계가 실제로 어떤 조건으로 선택되고 녹색 로그에 무엇이 남는지는 원격 실행으로만 확인할 수 있다.
- 출력되지 않은 코드(발견 0인 규칙)가 소스 수집에서 빠졌는지는 이 방법으로 가를 수 없다.
- 병합 트리의 `internal/cli` 테스트는 컴파일만 했고 실행하지 않았다.
