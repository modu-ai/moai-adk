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

- **CI Lint job** (`.github/workflows/ci.yml:422-458`): `golangci-lint run` (v2.1.6). `.golangci.yml`은 `default: none` + errcheck/govet/ineffassign/staticcheck/unused 5개 — **gofmt/goimports 포매터 미활성**. CI에 gofmt 잡은 별도로 없음. → 포맷은 처음부터 검사 대상이 아니었다.
- **`moai gate`** (`internal/cli/gate.go`): gofmt/format 언급 0회 → 포맷 무검사.
- **pre-commit fast-subset** (`internal/cli/hook_install_precommit.go`, gofmt -l on staged — 2026-07-05 `52b5e4bf5` 도입): 설치기만 존재, **본 리포 `.git/hooks/pre-commit` 미설치** → 본 리포에선 한 번도 작동한 적 없음. 배포(사용자 프로젝트) 계층용.
- 최근 커밋 파일들(`gate_graph_freshness_test.go`, `inbox_lifecycle_test.go`)은 gofmt-클린 — 부채는 훅/의식 도입 이전 대량 착지분.
- **결론: 게이트가 없어서 쌓였다.** 게이트가 없는 한 `gofmt -w` 한 번으로는 다시 쌓인다 (Residual-risk 권고 참조).

### 파일 분류 (배차 지시: 생성/의도적 미포맷 팄별)

- **testdata 2건** — `internal/navigator/astx/testdata/enrich/src/auth/{handler,token}.go`: 파서 입력 픽스처. 포맷 적용 후 `go test ./internal/navigator/...` 전부 통과로 **동작 불변 확인 → 포함** (제외 불요).
- **나머지 152건**: 실코드+테스트. diff 성격은 구조체 필드/슬라이스 리터럴 주석의 탭 정렬(tabwriter) 차이뿐 — 생성기 출력물이나 의도적 미포맷 없음. `gofmt -w` 일괄 적용 안전.

### 수정 + 검증

- `gofmt -w .` → `gofmt -l .` 재측정 **0건**.
- `go build ./...` → OK.
- 테스트 (건드린 23패키지 전부): navigator/mx/template/hook/harness/spec/config/chain/sandbox/session/statusline/goal/runtime/core/permission/constitution/settings/resilience/paths/migration/graph/pkg/models 통과 + `internal/cli` 17패키지 전부 ok (백그라운드 완료, exit 0, FAIL 라인 0). template(2)/spec(1)의 실패는 아래 사전 존재 3표면.
- `golangci-lint run` (로컬 v2.10.1): 1건 `catalog_tree_hash.go:60` errcheck — **사전 존재** (gofmt 목록 밖 + `git diff` 0, 내 변경 밖).

### 사전 존재 실패 3표면 (gofmt 무관, 대조로 확정)

`TestManifestHashFormat`(template), `TestGoldenCommittedArtifactsMatchEmission`(template/agentemit), `TestCatalogHashParity`(spec) — 전부 `sync-auditor.md` catalog 해시 1건(stored `f1b4487f…` ≠ computed `545d03d9…`). **gofmt 변경 없는 develop 팁 원본 트리(/tmp/t457-devtip)에서 동일 실패 재현** → 내 변경과 무관. 알려진 잔여 드리프트(`4244c4a06` 소관 운영자 큐, lane-14 기록과 일치) — 본 카드 소관 아님.

## Baseline-attribution

- 측정·수정 대상: `WT-gofmt-drift` @ `5107bbfff` (워크트리 `.claude/worktrees/t457`, 로컬 develop 팁 fast-forward 흡수 완료)
- gofmt: `/opt/homebrew/bin/gofmt` (go1.26.4); 1.25.6/1.26.0 Cellar 교차 검증 포함
- 모든 수치는 본 턴의 명령 출력에서 직접 관측 (lane-9 수치 미인용 — 대신 불일치 사실로 기록)

## Gaps

- `internal/cli` 테스트: **완료** — 백그라운드 런 exit 0, 17패키지 전부 ok (600s+ 소요, 대형 패키지 특성).
- CI 동일 버전(v2.1.6) golangci-lint로의 재측정 미수행 (로컬은 v2.10.1) — errcheck 1건이 CI에서도 잡히는지는 CI 판정 몫.
- 전체 스위트(`go test ./...`) 로컬 미실행 — 레포 규율(부하 사고 2026-08-15)에 따라 건드린 23패키지로 스코프 제한, 전 판정은 CI 몫.

## Residual-risk

- **재축적**: 게이트 부재가 해소되지 않으면 gofmt 미포맷은 다시 쌓인다. 권고(운영자 결정 사항, 본 카드 범위 밖): (a) `.golangci.yml`에 `formatters.enable: [gofmt]` 추가(로컬↔CI 셋 일치 원칙 유지), 또는 (b) CI에 `gofmt -l` step 추가, 또는 (c) 본 리포에 pre-commit 설치기 적용. (a)가 `.golangci.yml`의 "로컬과 CI가 같은 셋" 취지와 부합.
- **병합 충돌**: 154파일 대량 리포맷이라 미착지 대기 카드들(t436·t438·t454, lane-9 8장 등)과 창 병합 시 충돌 가능. 본 브랜치는 미푸시 유일본 — 창 순서 판단은 리드 소관.
- testdata 2건은 포맷 적용해 파서 테스트 통과를 확인했으나, 향후 테스트가 "포맷 안 된 입력"을 기대로 바뀌면 재조정 필요할 수 있음.
