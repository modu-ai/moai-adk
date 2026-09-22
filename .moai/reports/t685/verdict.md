# t685 판정서 — GH #1680 잔여: heavy gate 중첩 툴체인 전량 실행

- **카드**: t685 (Tier M, Class B — 재현 → 수리 → sync)
- **브랜치**: `WT-nested-toolchain` (기점 `5e0f71175` = origin/develop)
- **커밋**: `252ee621d` fix(quality): run every detected toolchain, not just the first marker match
- **독립 감사**: sync-auditor **PASS, 조화평균 8.87/10** (`sync-audit.md` 동봉, 발견 5건 전부 optional)

## 1. Claim (주장)

최상단 일치에서 단락회전하던 heavy gate의 마커 탐지가 **일치하는 툴체인 전부를 수집**해 각자의
모듈 루트에서 실행하도록 바뀌었다. 이슈 #1680의 T0/T1 재현(최상단 package.json + 중첩
apps/id/go.mod)에서 Go 툴체인 도달=true, 단일 언어 프로젝트의 단계 목록은 불변 — 카드 [HARD]
회귀 요구 두 항을 모두 충족한다.

## 2. Evidence (증거 — 명령 + 그대로의 출력, 이번 실행)

**RED (수리 전 트리 `5e0f71175`에서 관측):**

```
$ go test ./internal/hook/quality/ -run 'TestNestedGo' -v
--- FAIL: TestNestedGoToolchainReachedDespiteTopLevelNode — go shim log is empty
--- FAIL: TestNestedGoDefectFailsGate — gate passed a nested Go module that does not compile
--- FAIL: TestNestedGoVetFailsAtModuleRootNotTop — gate passed a non-compiling nested module
```

수리 전 요약은 5단계 전부 Node(`typecheck/eslint/biome/oxlint/npm test`)뿐 — Go 스텝 행 자체가
없는 silent-zero-coverage 형상. RED 사유가 결함 그 자체(wrong-reason red 아님).

**GREEN (커밋 트리에서 관측):**

```
$ go test ./internal/hook/quality/          → ok  github.com/modu-ai/moai-adk/internal/hook/quality 20.188s
$ go test -count=1 -cover ./internal/hook/quality/  → ok ... coverage: 90.2% of statements  (감사자 측정)
$ golangci-lint run ./internal/hook/quality/ → 0 issues.
$ go vet ./internal/hook/quality/            → exit 0
$ go build ./...                             → exit 0
$ go test ./internal/cli/ -run 'TestPreCommit|TestPrecommit|Relocation' -count=1
                                             → ok  github.com/modu-ai/moai-adk/internal/cli 21.675s
```

셰임 로그로 관찰된 실행(공허 녹색 차단 — 녹색 경로 스텝은 출력이 없어서 로그가 관찰 유일 수단):

```
/private/var/.../001/apps/id vet ./...    ← 중첩 모듈 루트에서 vet 실행 (루트 바인딩)
```

T1(컴파일 불가 중첩 모듈)은 게이트가 실패하며 `undefined` + `probe.go` 진단을 그대로 운반.

**구현 요약** (`internal/hook/quality` 한정, 소스 1 + 신규 테스트 1 + 기계 갱신 9):

- `detectToolchains() []detectedToolchain` — 최상단 첫 일치(기존 우선순위 보존) + 경계 있는
  walk(`DefaultGateMarkerScanDepth`, `sourceScanSkipDirs` 생략, 모듈 루트 불하강)로 언어당
  1엔트리 수집
- `Run()` — 엔트리별 정적 단계(vet→typecheck→lint) → ast-grep 1회(프로젝트 단위) → 엔트리별
  테스트 단계
- `executeStep`/`runStep` — 루트 `dir` 명시 매개변수(config·staged·source 범위와 cwd가 전부
  그 루트를 따름), 빈 값은 기존 프로젝트 dir 폴백
- `detectToolchain()` — t559 테스트 호환 래퍼(최상단 우선, 없으면 테이블 순위 1위)

## 3. Baseline-attribution (baseline 귀속)

- RED: `5e0f71175`(워크트리 작업 전 HEAD, 이번 실행). 감사자가 같은 SHA를 /tmp에 materialize해
  행동 테스트 3건을 이식·실행 → 3건 모두 실패로 **독립 재현** 확인.
- GREEN: `7063a4c83`(1차 커밋)에서 전 배측정 → F2 주석 정리 amend 후 `252ee621d`에서 테스트
  ok 20.188s · 린트 0건 · vet 재측정(위 출력). 감사자의 커버리지 90.2%·ok 21.499s는
  `7063a4c83` 기준(주석-only 차이라 미영향).

## 4. Gaps (미검증 — 명시적으로 관측하지 못한 것)

1. Windows: 셰임·스크립트 페이크 바이너리 테스트는 Windows에서 자체 skip — 실제 Windows 경로는
   CI 몫.
2. 전체 스위트·darwin/windows 매트릭스: develop push 후 CI 판정(레인 로컬 미실행 — lane-local
   검증 규정).
3. 실제 npm/eslint가 설치된 환경의 Node 다중 툴체인 경로는 모의 바이너리로만 검증.
4. 최상단에 두 언어 마커가 동시에 있는 폴리글랏 레포는 기존 우선순위(1개) 그대로 — 이슈·카드
   범위 밖으로 유지(이슈 본문도 "root-level first match keeps existing precedence"로 명시).

## 5. Residual-risk (잔여 위험 — 관측했음에도 남을 수 있는 것)

- **F1 [Low]** 다중 엔트리에서 `typecheck` 요약 행이 나중 기록으로 덮임(루트 Node 실행 → 중첩 Go
  skip 표기로 역전). 판정은 전 순서 안전(실패는 early-return)하나 보고 계약 위반 — 감사자 /tmp
  재현 테스트 존재, **후속 소형 카드 권고**.
- **F4 [Low·보안 인접]** walk가 임베디드 git repo·`.claude/worktrees`를 경계로 보지 않음 —
  Node 루트 + 워크트리 내 Go 모듈 레포에서 타 카드 트리 게이트 실행 가능(t559 노출의 강화형).
  본 저장소는 루트 go.mod라 안전. F4+F5 같은 패치에서 walk 경계 보강 권고.
- **F3 [Low]** 래퍼 반환 엔트리와 `detectedRoot` 불일치 가능(walk 순서 vs 테이블 순위) — 현재
  무해(`stepDir` 프로덕션 소비자 0, Run은 엔트리별 dir).
- **F5 [Info]** `cachedStagedFiles` 루트 간 공유 캐시는 올바름 확인 — 임베디드 별도 repo의
  저확률 예외만 잔여.

후속 카드 권고: (a) F1 typecheck 행 엔트리별 표기, (b) F4+F5 walk 경계 보강, (c) F2/F3는 다음
해당 파일 편집 시 수반 정리(F2는 이번 amend로 반영 완료).
