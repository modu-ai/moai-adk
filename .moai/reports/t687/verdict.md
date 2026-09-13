# t687 판정서 — #1631 잔여: gate가 프로젝트 자체 scripts.lint를 실행하지 않음

- **카드**: t687 (Tier S, Class B — 재현 → 수리 → sync)
- **브랜치**: `WT-scripts-lint` (기점 `e4ecbf854` = 로컬 develop)
- **커밋**: `add08e10c` feat(quality): run the project's own scripts.lint as the Node lint axis
  + `2730c5520` fix(quality): guard scripts.lint resolution to the Node lint axis (감사 F1 수리)
- **독립 감사**: sync-auditor 초판 **FAIL(6.6/10, F1 차단)** → 델타 재판정 **PASS(9.0/10)** (`sync-audit.md` 동봉)

## 1. Claim (주장)

package.json에 `scripts.lint`를 선언한 Node 프로젝트에서 게이트가 이제 그 명령을 lint 축으로
실행한다. 우선순위(카드가 정하라 한 설계 판단): scripts.lint 존재 시 설정 파일 게이트 3엔트리
(eslint/biome/oxlint)를 **대체** — 프로젝트의 자기 선언이 게이트의 추측보다 우선하고, 둘 다
돌리면 이중 lint. watch-prone lint 스크립트는 타임아웃까지 매다는 대신 이유를 남기며 스킵.
스크립트가 없으면 기존 동작 그대로(불변). [HARD] 안전성: 해석된 스텝은 기존과 동일한
executeStep→runStep 경로(프로세스 그룹 격리·LintTimeout·stepEnv git 스크럽 상속 — 감사자가
코드로 확인). 언어 가드로 Go 등 타 언어 축의 침묵 대체를 차단(감사 F1).

## 2. Evidence (증거 — 명령 + 그대로의 출력)

**RED (수리 전 `e4ecbf854`, /tmp materialize):**

```
$ go test -C /tmp/t687-base ./internal/hook/quality/ -run 'Lint' -count=1 -v   → EXIT=1
--- FAIL: TestScriptsLintViolationFailsGate
--- FAIL: TestScriptsLintCleanProjectPasses
--- FAIL: TestScriptsLintSupersedesConfigEntries
--- FAIL: TestWatchProneLintScriptSkippedVisibly
--- FAIL: TestRealNpmScriptsLintViolationFailsGate
(--- PASS 18 — 불변 보존증명 TestNoScriptsLintLeavesLintAxisUnchanged 포함; 5+18=23 전량 선택)
```

**F1 회귀 RED (감사 지적 결함이 있는 `add08e10c`, /tmp materialize):**

```
$ go test ... -run 'TestGoToolchainKeepsGolangciLintBesideScriptsLint' -v
--- FAIL: TestGoToolchainKeepsGolangciLintBesideScriptsLint (0.26s)   ← 게이트가 조용히 통과
```

**GREEN (최종 트리 `2730c5520`, 이번 실행):**

```
scripts.lint 7테스트 전부 --- PASS (선택자 실행)
$ go test ./internal/hook/quality/            → ok  ... 21.051s
$ golangci-lint run ./internal/hook/quality/  → 0 issues.
$ go vet ./internal/hook/quality/             → exit 0
$ gofmt -l internal/hook/quality/             → (빈 출력) / $ go build ./... → 성공
```

감사자 독립 재측정: `ok 20.562s`·`0 issues.`·vet·gofmt 청결 + 자체 A/B 바이너리 재실행 —
하이브리드(go.mod+package.json) 픽스처에서 `EXIT=1`, `quality gate failed: golangci-lint`,
`npm run lint` 행 부재(1차 침묵 대체 관측의 정확한 역전 = F1 봉합), Node 전용 픽스처는
`npm run lint: executed`(과차단 없음).

## 3. Baseline-attribution (baseline 귀속)

- 초판 RED: `e4ecbf854` materialize, 이번 실행.
- F1 회귀 RED: `add08e10c` materialize, 이번 실행. 감사자 초판도 동일 결함을 /tmp 가짜
  바이너리 A/B로 독립 입증.
- 최종 GREEN: `2730c5520` 워크트리, 이번 실행 + 감사자 재측정(같은 트리).
- 감사 이력: 초판 FAIL(add08e10c, F1 차단·6.6/10) → 델타 재판정 PASS(2730c5520, 9.0/10).

## 4. Gaps (미검증)

1. Windows: 셸 스크립트 fake 바이너리 테스트 자체 skip — 실제 Windows는 CI 몫.
2. `-race`·커버리지 수치: 이 카드에서 미실행(패키지 평시 게이트는 CI가 대행).
3. 전체 스위트·크로스플랫폼 매트릭스: develop push 후 CI 판정.
4. 감시형 스크립트 판별은 REQ-HGT-002의 테스트 스텝용 술어 계승 — chokidar·nodemon류는
   미포착(F4, 후속).

## 5. Residual-risk (잔여 위험)

- **F4 [Low, 후속]** watch-prone 판별이 `--watch`/bare-vitest 토큰 한정 — nodemon 류 감시
  lint는 LintTimeout false red 가능. 후속 카드 권고.
- **F3 [Low, 미관]** watch-prone 스킵 행이 지연 시딩이라 요약 표 순서 흔들림.
- **공시 행**: 대체 발생을 요약에 알리는 별도 행 — 감사자 판정 "현 시점 발생하지 않는 상태의
  방어라 불요", 별도 결정 사항으로 남김.
- 실행 도중 package.json 수정 시 시딩·실행 해석 갈라짐 — 병적 상황, 미미.

후속 카드 권고: F4(감시형 판별 확장) 단건.
