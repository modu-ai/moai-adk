# t457 verdict — gofmt 미포맷 수정 + 포맷 게이트 부재 판정

카드: t457 (Class B — plan 생략) · 브랜치: `WT-gofmt-drift` (develop 팁 `5107bbfff` 위) · 2026-09-03

## Claim

1. 배차 상정 "10건"과 달리, **전체 트리 gofmt 미포맷은 154건**이며 lane-9 base `5a8449859`에서도 동일하게 154건 — 오래된 부채이고 lane-9의 "10건"은 좁은 스코프 측정이었다.
2. "게이트가 있는데 통과했다"가 아니라 **포맷 게이트 자체가 존재하지 않는다** — CI의 golangci-lint 활성 셋에 gofmt 포매터가 없고, `moai gate`도 gofmt를 검사하지 않는다. 배차가 말한 "더 큰 축"이 맞다.
3. 154건 전부 `gofmt -w` 적용 → 재측정 0건. 구문 변화 없음(+801/−762 대칭, 재정렬만).

## Evidence

### 재측정 (배차 지시: 이 트리에서 다시 잴 것)

| 측정 | 대상 트리 | gofmt | 결과 |
|---|---|---|---|
| 전체 트리 | develop 팁 `5107bbfff` (= 본 브랜치 베이스) | go1.26.4 (homebrew) | **154건** |
| lane-9 base | `5a8449859` (`git archive` → /tmp 추출) | go1.26.4 | **154건** |
| 버전 교차 | 동일 트리 | go1.25.6 / go1.26.0 (Cellar 병존) | **둘 다 154건** |

→ 배차 예상 "10건"은 재현되지 않음. 43커밋 유입 구간(`5a8449859..5107bbfff`)에서 바뀐 Go 파일은 13개뿐 — 부채 유입이 아니라 lane-9 측정 스코프의 문제. *"확인했고 그 수치는 재현되지 않았다"*를 기록으로 남긴다.

### 게이트 질문의 답 — 왜 통과해 왔는가

- **CI Lint job** (`.github/workflows/ci.yml:422-458`): `golangci-lint run` (v2.1.6). `.golangci.yml`은 `default: none` + errcheck/govet/ineffassign/staticcheck/unused 5개 — **gofmt/goimports 포매터 미활성**. CI에 gofmt 잡은 별도로 없음.
- **`moai gate`** (`internal/cli/gate.go`): gofmt/format 언급 0회 → 포맷 무검사.
- **pre-commit fast-subset** (`internal/cli/hook_install_precommit.go`, gofmt -l on staged — 2026-07-05 `52b5e4bf5` 도입): **훅 파일은 설치돼 있다** — `primary .git/hooks/pre-commit` (3245 B, 2026-08-19 21:53, gofmt 6회 참조). **그러나 리포 로컬 설정 `.git/config`의 `core.hooksPath = /dev/null`이 훅 검색 경로를 무효화해 실행되지 않는다** (본 워크트리에서 `git rev-parse --git-path hooks` → `/dev/null`로 측정). 즉 설치와 활성화가 분리된 상태. 추가로 훅이 활성화돼도 `SKIP_MOAI_PRECOMMIT=1` 환경변수로 우회 가능 — 우회 경로가 남아 있으면 활성화만으로는 게이트가 아니다.
- ⚠️ 측정 함정 기록: 워크트리 세션에서 `ls .git/hooks/`는 **항상 실패한다** — 워크트리의 `.git`는 gitdir 포인터 파일이므로. 훅 존재 확인은 `git rev-parse --git-common-dir`/primary 경로로. 본 verdict 초판은 이 함정으로 "미설치"로 잘못 적었다가 리드 대조로 정정했다.
- 최근 커밋 파일들(`gate_graph_freshness_test.go`, `inbox_lifecycle_test.go`)은 gofmt-클린 — **훅이 작동해서가 아니라**(실행 무효 상태) 최근 착지 레인들이 클린하게 들여온 것. 8/19 이전 대량 착지분이 그대로 부채로 남았다.
- **결론: 현재 활성인 포맷 게이트는 0개다.** CI·`moai gate`에는 포맷 검사가 없고, 설치된 pre-commit 훅은 `core.hooksPath=/dev/null`로 실행 무효. 게이트가 없는 한 `gofmt -w` 한 번으로는 다시 쌓인다 (Residual-risk 권고 참조).

### 파일 분류 (배차 지시: 생성/의도적 미포맷 팄별)

- **testdata 2건** — `internal/navigator/astx/testdata/enrich/src/auth/{handler,token}.go`: 파서 입력 픽스처. 포맷 적용 후 `go test ./internal/navigator/...` 전부 통과로 **동작 불변 확인 → 포함** (제외 불요).
- **나머지 152건**: 실코드+테스트. diff 성격은 구조체 필드/슬라이스 리터럴 주석의 탭 정렬(tabwriter) 차이뿐 — 생성기 출력물이나 의도적 미포맷 없음. `gofmt -w` 일괄 적용 안전.

### 수정 + 검증

- `gofmt -w .` → `gofmt -l .` 재측정 **0건**.
- `go build ./...` → OK.
- 테스트 (건드린 23패키지 전부): navigator/mx/template/hook/harness/spec/config/chain/sandbox/session/statusline/goal/runtime/core/permission/constitution/settings/resilience/paths/migration/graph/pkg/models 통과 + `internal/cli` 17패키지 전부 ok (백그라운드 완료, exit 0, FAIL 라인 0). template(2)/spec(1)의 실패는 아래 사전 존재 3표면.
- `golangci-lint run` (로컬 v2.10.1): 1건 `catalog_tree_hash.go:60` errcheck — **사전 존재** (gofmt 목록 밖 + `git diff` 0, 내 변경 밖).

### 사전 존재 실패 3표면 (gofmt 무관, 대조로 확정)

`TestManifestHashFormat`(template), `TestGoldenCommittedArtifactsMatchEmission`(template/agentemit), `TestCatalogHashParity`(spec) — 전부 `sync-auditor.md` catalog 해시 1건(stored `f1b4487f…` ≠ computed `545d03d9…`). **gofmt 변경 없는 develop 팁 원본 트리(/tmp/t457-devtip)에서 동일 실패 재현** → 내 변경과 무관. 소관은 t443(lane-14)이며 본 카드보다 앞 순번이라 그 착지 후 소멸 예상 (리드 판정).

## Baseline-attribution

- 측정·수정 대상: `WT-gofmt-drift` @ `5107bbfff` (워크트리 `.claude/worktrees/t457`, 로컬 develop 팁 fast-forward 흡수 완료)
- gofmt: `/opt/homebrew/bin/gofmt` (go1.26.4); 1.25.6/1.26.0 Cellar 교차 검증 포함
- 모든 수치는 본 턴의 명령 출력에서 직접 관측 (lane-9 수치 미인용 — 대신 불일치 사실로 기록)

## Gaps

- `internal/cli` 테스트: **완료** — 백그라운드 런 exit 0, 17패키지 전부 ok (600s+ 소요, 대형 패키지 특성).
- CI 동일 버전(v2.1.6) golangci-lint로의 재측정 미수행 (로컬은 v2.10.1) — errcheck 1건이 CI에서도 잡히는지는 CI 판정 몫.
- 전체 스위트(`go test ./...`) 로컬 미실행 — 레포 규율(부하 사고 2026-08-15)에 따라 건드린 23패키지로 스코프 제한, 전 판정은 CI 몫.

## Residual-risk

- **재축적**: 활성 포맷 게이트가 0개라 gofmt 미포맷은 다시 쌓인다. 권고(운영자 결정 사항, 본 카드 범위 밖) — **t457 착지 직후가 켜는 시점**(그 전에 켜면 CI가 154건 전부 적색): (a) `.golangci.yml`에 `formatters.enable: [gofmt]` 추가(로컬↔CI 셋 일치 원칙 부합, 가장 작음), 또는 (b) CI에 `gofmt -l` step 추가, 또는 (c) `core.hooksPath=/dev/null` 해제로 설치된 pre-commit 훅 활성화 — 단 `SKIP_MOAI_PRECOMMIT=1` 우회가 남아 활성화만으로는 완전하지 않음.
- **통합 순서**: 리드 판정으로 **마지막**(lane-13(t442·t449) → … → lane-7(t458) → 본 카드). 근거: 재포맷은 재실행으로 해소되는 유일한 충돌 종류 — 마지막에 들어가면 뒤 카드 12장의 개별 흡수가 필요 없고 본 카드가 전부 흡수한 뒤 `gofmt -w` 재실행으로 끝난다. **창을 받으면 흡수 후 `gofmt -l` 재측정 → 재포맷 커밋** — 현 커밋(`1444583bf`)을 그대로 병합하면 충돌. 본 커밋 베이스는 `5107bbfff`이며 로컬 develop은 계속 진행 중(리드 기준 `d63dee78d` 미푸시 11) — 흡수 필수.
- testdata 2건은 포맷 적용해 파서 테스트 통과를 확인했으나, 향후 테스트가 "포맷 안 된 입력"을 기대로 바뀌면 재조정 필요할 수 있음.
