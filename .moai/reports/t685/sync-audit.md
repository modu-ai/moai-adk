# t685 sync-audit — 중첩 툴체인 게이트 (run every detected toolchain)

- 카드: t685 (Factory lane-7, Tier M, Class B) — GH #1680 residual
- 대상 커밋: `7063a4c83` (branch `WT-nested-toolchain`, base `5e0f71175` = origin/develop head)
- 감사자: sync-auditor (독립 감사, 1차 반복)
- 감사일: 2026-09-13, baseline: 이 워크트리(`.claude/worktrees/t685`) HEAD `7063a4c83`, clean tree

## 판정 (Verdict)

**PASS — harmonic mean 8.87 / 10** (Functionality 9 · Security 9 · Craft 8.5 · Consistency 9)

필수 통과 차원(Functionality, Security) 모두 독립 충족. 차단 결함(blocking) 없음 — 발견 결함 5건 전부 optional.

## 기계적 검증 근거 (이번 실행, 이 트리)

| 주장 | 명령 | 관측 결과 (verbatim) |
|---|---|---|
| 패키지 테스트 통과 | `go test ./internal/hook/quality/` | `ok github.com/modu-ai/moai-adk/internal/hook/quality 21.499s` |
| 커버리지 ≥ 85% | `go test -count=1 -cover ./internal/hook/quality/` | `ok ... 21.023s coverage: 90.2% of statements` |
| 린트 클린 | `golangci-lint run ./internal/hook/quality/` | `0 issues.` (exit 0) |
| 정적 분석 | `go vet ./internal/hook/quality/` | exit 0, 출력 없음 |
| 포맷 | `gofmt -l internal/hook/quality/` | 빈 출력 (exit 0) |
| 전체 빌드 | `go build ./...` | exit 0 |

모든 명령은 `unset MOAI_KANBAN … &&` 단일 복합 호출로 환경 스크럽 후 직접 실행했고, 캐시 무시(-count=1) 실행도 별도로 확인했다.

### RED 재검증 — 레인 주장의 독립 재현

레인이 보고한 "pre-fix 트리 `5e0f71175`에서 행동 테스트 3건 실패"를 그대로 믿지 않고, `git archive 5e0f71175`로 /tmp에 base 트리를 materialize한 뒤 신규 테스트 파일에서 API 레벨 테스트(detectToolchains 참조 — base에서 컴파일 불가)만 제외한 행동 테스트 3건을 그대로 옮겨 실행했다. 결과 — 3건 모두 실패, 레인 주장과 정확히 일치:

```
--- FAIL: TestNestedGoToolchainReachedDespiteTopLevelNode
    go shim log is empty — the gate ran zero Go toolchain steps (GH #1680 T0: silent zero coverage)
--- FAIL: TestNestedGoDefectFailsGate
    gate passed a nested Go module that does not compile (GH #1680 T1)
--- FAIL: TestNestedGoVetFailsAtModuleRootNotTop
    gate passed a non-compiling nested module
FAIL  github.com/modu-ai/moai-adk/internal/hook/quality  0.603s
```

base 트리의 게이트 요약이 이 결함의 본질을 그대로 보여준다: 5단계 전부 Node(eslint/biome/oxlint/npm test)이고 전부 skipped인 채 게이트는 통과 — Go 모듈은 요약 어디에도 없다. silent-zero-coverage 형상 재현 완료. HEAD에서는 동일 테스트 3건이 통과한다(패키지 전체 `ok`에 포함).

### F1 재현 (감사 중 발견, 아래 참조)

HEAD 사본(/tmp)에서 루트 Node `scripts.typecheck` + 중첩 Go 조합으로 실행:

```
typecheck row: outcome=skipped command="" reason="no default for this language; set gate.typecheck.command to enable one"
INVERSION REPRODUCED: root Node ran `npm run typecheck` (exit 0) but the shared row reports outcome=skipped
```

## 차원 점수

| 차원 | 점수 | 근거 한 줄 |
|---|---|---|
| Functionality | 9/10 | 카드 요구(전 툴체인 수집·실행, 모듈 루트 결박, root 우선순위 보존, 단일 언어 불변)를 RED 재검증 + GREEN + API 테스트 5건으로 End-to-End 확인; F1 요약 표기 결함만 감점 |
| Security | 9/10 | 신규 주입면 없음(step dir는 walk 파생 경로, 바이너리는 정적 테이블, `stepEnv()` GIT_DIR 스크럽(t560) 유지); F4는 기존 노출의 강화형이지 새 방향이 아님 |
| Craft | 8.5/10 | 커버리지 90.2%(목표 85%), lint 0, shim 기반 관찰 테스트로 공허 녹색 차단 — 다만 dispatch가 지목한 다중 엔트리 typecheck 표기 영역을 테스트가 놓친 것(F1)과 주석 중복(F2) |
| Consistency | 9/10 | `runToolchain*Step` 분해·`appendReason`/`withSummary` 재사용·영어 주석·REQ 참조 보존으로 파일 관습 준수; 래퍼 `detectToolchain()`을 t559 테스트 호환용으로 의도 보존(프로덕션 호출자 0, 테스트 고정 다수 확인) |

harness mean = 4 / (1/9 + 1/9 + 1/8.5 + 1/9) = **8.87**

## 발견 결함 (Findings)

- **F1 [Low] [optional]** `internal/hook/quality/gate.go:560` — typecheck 요약 행 라벨 충돌(다중 엔트리). `typecheckStepName`("typecheck")은 언어 무관 단일 라벨로 한 번만 seed되는데, `runToolchainStaticSteps`는 엔트리마다 같은 행에 기록한다 — `markSkipped`(기본 분기)와 `markExecuted` 모두 무조건 덮어쓴다(last-writer-wins). 루트 Node가 `npm run typecheck`을 실행해 통과해도, 뒤이은 중첩 Go(기본값 없음)의 skip이 행을 `skipped`로 뒤집는 것을 /tmp HEAD 사본에서 기계적으로 재현했다. 게이트 판정은 모든 순서에서 안전하다(실행 실패는 즉시 early-return, skip은 pass 측) — 그러나 "모든 결과는 보고된다, 결코 침묵 없다"는 이 축 자신의 계약과 어긋나게 실행을 skip으로 오보고한다. 수정 제안: 엔트리별 라벨(예: `typecheck (apps/id)`)을 쓰거나, executed 행을 skip으로 덮지 못하게 한다. Required fix: `gate.go` runToolchainStaticSteps의 typecheck 분기에서 행 라벨에 엔트리 루트를 붙이거나 markSkipped의 executed-덮어쓰기를 금지.
- **F2 [Low] [optional]** `internal/hook/quality/gate.go:1138-1143` — `executeStep` 문서 주석 3줄이 그대로 이중 삽입됨(기존 3줄 + 동일 3줄 + 신규 dir 단락). 외견 결함. Required fix: 중복 3줄 삭제.
- **F3 [Low] [optional]** `internal/hook/quality/gate.go:632-650, 776` — 래퍼 `detectToolchain()`은 중첩 엔트리 중 tableIdx 최솟값을 반환하는데 `detectedRoot`는 `found[0].root`(walk 순서)에 결박된다. 중첩 이언어 모듈 2개에서 walk 순서 ≠ 테이블 순서(예: `alpha-app/Cargo.toml`이 `zebra-app/package.json`보다 먼저 발견, 테이블은 Node가 Rust보다 앞)면 반환 엔트리와 결박 루트가 어긋난다 — 구 코드는 반환 매치의 루트에 결박했다. 현재 무해: `stepDir`의 유일한 프로덕션 소비자는 `anyConfigFileExists`이고 그 호출자도 전부 테스트다(Run은 엔트리별 dir 사용). 단, `detectToolchains` 주석의 "keeping stepDir's resolution meaningful"은 과장이라 정정 대상. Required fix: 래퍼가 선택한 엔트리의 루트를 결박하거나 주석을 실제 동작에 맞게 정정.
- **F4 [Low] [optional, 보안 인접]** `internal/hook/quality/gate.go:1229` — walk가 임베디드 git 저장소/워크트리 루트를 경계로 보지 않는다. `sourceScanSkipDirs`는 의존성·빌드·캐시 이름만 포함하고 `.claude`/`.moai`/`.git 파일을 품은 디렉터리는 제외 대상이 아니다. Node 루트 프로젝트의 `.claude/worktrees/<card>/` 아래(깊이 ≤4)에 Go 모듈이 있으면 다른 카드의 진행 중 트리에서 `go vet`이 돌고, 그 트리 상태 때문에 게이트가 실패할 수 있다. t559 단일 승자 스캔에도 같은 노출이 있었으나(그때는 더 심각하게 승자가 되었음) 이번 변경으로 해당 모듈 전부가 실행된다 — 기존 노출의 강화형. Required fix: `.git` 항목(.git 파일 포함)을 품은 디렉터리에서 walk를 멈추거나 `.claude`/`.moai`를 skip 집합에 추가.
- **F5 [Info] [optional]** `internal/hook/quality/gate.go:1206` — `cachedStagedFiles` 단일 비키드 캐시 확인. 루트 간 공유는 올바르다: `git diff --cached`(`cmd.Dir`=툴체인 루트, GIT_DIR 스크럽 유지)는 임의 하위디렉터리에서 둘러싼 저장소를 해석하고 `hasStagedExt`는 확장자만 읽어 루트 스코핑이 무관하기 때문. 단, 프로젝트 안에 별도 git 저장소(서브모듈·임베디드 repo)가 마커와 함께 걸리면 그 루트의 `git`은 자기 저장소로 해석되고, 캐시 선점 순서에 따라 (보통 빈) staged 집합이 전 실행의 확장자 필터를 오염시킬 수 있다 — 보수 방향(skip)이며, 구 코드도 중첩 엔트리에 프로젝트 상단 집합을 썼으므로 기존 형상. 잔여 위험으로 유지.

확인하여 결함 아님(카드 명세대로): 루트에 go.mod+package.json이 함께 있으면 테이블 선순위 1개만 수집 — "root-level first match keeps existing precedence"의 의도된 동작이며 래퍼와 Run 양쪽이 이에 의존. walk 오류 경로는 클로저가 nil/SkipDir만 반환해 `walkErr`가 항상 nil이라 `_ =` 폐기 안전. 깊이 상한(`DefaultGateMarkerScanDepth=4`)·모듈 루트 불하강·동일 언어 중복 수집 방지(`detected` 맵) 모두 확인. ast-grep 프로젝트 레벨 1회 실행 확인.

## 범위 규율 (Scope)

- diff는 `internal/hook/quality`에 한정: gate.go 1개 + 신규 테스트 1개 + 기존 테스트 9개의 기계적 시그니처 갱신(`""` 삽입) — 전부 확인 결과 동작 변경 없음.
- 템플릿 트리 변경 없음. 커밋 메시지 Conventional Commits + `Card: t685` + 🗿 MoAI 추적성 충족.
- drive-by 리팩터 없음. 죽은 코드(`anyConfigFileExists`/`stepDir`는 프로덕션 호출자 없음)도 삭제하지 않고 보존 — 범위 규율상 올바른 선택.

## 잔여 위험 (Residual risk)

- F5의 임베디드 저장소 staged 집합 오염(저확률, 보수 방향).
- F4의 worktree/임베디드 repo walk 진입 — moai-adk-go 자체는 루트 go.mod가 있어 Go가 루트에서 감지되고 walk가 다른 Go를 건너뛰므로 본 저장소는 안전하나, Node 루트 + 워크트리 내 Go 조합 사용자에게 도달 가능.
- 다중 엔트리 요약에서 typecheck 외 행은 라벨이 언어별로 유일해 충돌이 없으나, 향후 툴체인 테이블에 동명 단계가 추가되면 F1과 같은 충돌이 재발할 수 있다(테이블 확장 시 라벨 유일성 체크 권장).

## Gaps (이번에 관측하지 않은 것)

- Windows에서의 shim 테스트 경로(테스트가 자체 skip) — 타 플랫폼 판정은 CI 몫.
- 전체 스위트·darwin/windows 매트릭스 — develop push 후 CI가 통합 판정한다(§4.1 규율).
- 실제 npm/eslint 설치 환경에서의 Node 경로 동작(픽스처는 config 부재로 skip되는 경로만 검증).

## 권고 (Recommendations)

1. F1을 후속 소형 카드로: 다중 엔트리 typecheck 라벨 분리 + 재현 테스트(이 감사의 /tmp 재현 테스트를 그대로 이식 가능).
2. F4: walk 경계에 `.git` 항목 보유 디렉터리 추가 — F5와 같은 패치로 함께 다룰 수 있다.
3. F2/F3는 다음 이 파일 편집 시점에 수반 정리(주석 중복 제거, detectedRoot 주석 정정).
