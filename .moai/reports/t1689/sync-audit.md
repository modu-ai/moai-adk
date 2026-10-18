auditor-model: glm-5.3
verdict: FAIL
audited_sha: f4d1c6be1419acdda921057f3cc31148083e51f5
required_backend: codex
required_backend_fail: true
audit_receipt: rcpt-abdf8b923152bd9908e37f2f,rcpt-099a0b40753dad1017f8c941
overall_score: 0.82

# t1689 sync 감사 (r1, 2026-10-18)

감사 대상: 카드 t1689의 기여 3커밋 — cad5f4e52(수정), 5dcfc44ef(main 흡수), f4d1c6be1(카드리뷰 영수증 docs).
트리: `WT-gate-fixture-lint` @ `f4d1c6be1419acdda921057f3cc31148083e51f5`, merge-base `bc5817d758`,
변경 파일 3개(`.moai/reports/t1689/card-review.md`, `internal/hook/quality/astgrep_gate.go`,
`internal/hook/quality/astgrep_gate_walk_testdata_test.go`).

차원 점수(조화평균 0.82; must-pass 방화벽으로 전체 판정은 FAIL):

| 차원 | 점수 | 판정 | 근거 |
|-----------|-------|---------|----------|
| Functionality (40%) | 75/100 | FAIL | 카드 자체 결함(출하 fixture 오탐)은 단위 수준에서 해소 — 3개 신규 회귀테스트 PASS. 그러나 수정이 도입한 제외 조건이 절대경로 세그먼트를 검사해, 상위 디렉터리 이름이 `testdata`인 프로젝트는 페어링 검사가 통째로 공소화(F1, 기계 재현 확인) |
| Security (25%) | 80/100 | PASS | 신규 입력·비밀·주입 면 없음. F1은 정책 검사 우회면이지만 Critical/High 아님(P2) |
| Craft (20%) | 85/100 | PASS | 양성 대조군 포함 테스트, 실 fixture 특성화 테스트, 주석이 순서 충돌을 설명. 상위 디렉터리 경계 케이스만 부재(F1 수정 요구에 포함) |
| Consistency (15%) | 90/100 | PASS | findings 측 제외(`astGrepExcludedPathPatterns "/testdata/"`)와 동일 규칙 미러, 파일 스타일·Conventional commit(카드 id 명기) 정합 |

## Claim

1. 감사 계획: 이 트리의 계획은 codex 단독 required(claude/glm off, `cross_model_active: true`)이며, 수렴 결과 `overall_verdict=fail`(required-backend FAIL: codex) — 이 카드의 동기 게이트가 실패했다.
2. codex P2 지적(F1)은 참양성이다: `walkSourceFiles`의 제외 조건이 절대경로 전체를 검사하므로, 프로젝트 상위에 `testdata`라는 이름의 디렉터리가 있으면 스윕이 0건이 되어 `SUPPRESSION_WITHOUT_REASON` 단속이 무효화된다. 이 감사관의 오버레이 프로브가 독립 재현했다.
3. 카드가 막으려던 원래 결함(출하 fixture `internal/astgrep/testdata/fixtures/go/suppressed.go` 오탐)은 이 트리에서 해소되어 있다(단위 수준).

## Evidence

감사 계획 — `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1689`:
```
"config_status": "ok", "model": "codex", "model_source": "config",
codex: "gate": "required", "source": "config.gates", "explicit": true (claude/glm: "off"),
"cross_model_active": true, "enforced_required": ["codex"]
```

수렴(mcp__moai__audit_multi, gates 인자 없음, project_root=자기 트리, target=baseBranch, focus="quality-gate exclusion correctness and regression coverage", session_id=t1689-sync-r1) 다이제스트:
```
overall_verdict: "fail"   participant_count: 1   disagreement_flag: null
fail_open_backends: []    plan_source: "config"
residual_risk_note: "required-backend FAIL: codex"
per_backend_verdicts: [ {backend: codex, gate: required, verdict: fail, findings: 1 (P2)} ]
audit_receipt: rcpt-abdf8b923152bd9908e37f2f
```

범위 지정 패키지 테스트 — `go test ./internal/hook/quality -count=1`(비-race):
```
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	49.592s
```
신규 3종 회귀테스트(-v, -run 지정): `--- PASS: TestRunAstGrepGateV2_TestdataMarkerLayoutDoesNotBlock`, `--- PASS: TestWalkSourceFiles_SkipsTestdata`, `--- PASS: TestWalkSourceFiles_SkipsShippedFixture`.

F1 독립 재현(오버레이 프로브, `go test -overlay=/tmp/t1689-audit-repro/overlay.json ... -run TestAuditReproAncestorNamedTestdata -v`; 프로브 파일은 /tmp에만 존재, 트리 무기입 — 실행 후 `git status --short` 빈 출력으로 확인):
```
zz_audit_repro_test.go:38: probe: project under testdata-named ancestor swept 0 files
zz_audit_repro_test.go:39: probe: control project swept 1 files ([.../plain/proj/prod.go])
zz_audit_repro_test.go:41: probe: checkSuppressionPairing on control prod.go reports 1 violation(s)
zz_audit_repro_test.go:43: probe: violation type=SUPPRESSION_WITHOUT_REASON
```
→ 동일한 무사유 `ast-grep-ignore` 파일이 상위 `testdata` 경로에서는 검사 자체를 안 받고, 대조 경로에서는 위반이 잡힌다.

코드 근거: `internal/hook/quality/astgrep_gate.go:386-391`(절대경로 세그먼트 5종 비교, 388행에 `testdata` 추가), 사용점 `:88`(`sourceFiles := walkSourceFiles(projectDir)`), findings 측 미러 `:220-223`(`/testdata/`).

라운드 2(정지 가드 구제 호출, `mcp__moai__codex_audit` 단일 백엔드, 동일 트리·동일 기준 `review_base: main (merge base bc5817d758…)`):
```
verdict: "pass"   findings: []   audit_receipt: rcpt-099a0b40753dad1017f8c941
요약: "지적 사항 없음 … 테스트·빌드·재현 명령은 실행하지 않았어" (정적 검토만, 무실행)
```
→ codex 두 다리의 내부 분열(r1 FAIL-with-repro vs r2 PASS-static-only). r2가 무실행 정적 검토라고 스스로 명시한 이상, 이 감사관의 실행 재현(오버레이 프로브)과 r1의 재현을 뒤집지 못한다. 판정은 FAIL을 유지하며, 이 분열 자체를 잔여 위험에 기록한다.

## Baseline-attribution

- 트리: 위 `audited_sha`(감사 개시 시점 HEAD). 모든 go 테스트·프로브·grep은 이 트리에서 이번 실행으로 관측.
- 판정 빌드(§2.2): 호출된 `moai`는 `v3.2.0-rc.36-t1822c`, 커밋 `ee981a932`, 빌드 2026-10-15T21:56:00Z. `git merge-base --is-ancestor ee981a932 HEAD` exit 0, `git rev-list --count ee981a932..HEAD` = 422 — 설치 빌드는 이 트리의 엄격한 조상(422커밋 뒤짐). 따라서 이 수정 코드를 담지 않은 빌드이며, 라이브 `moai gate` 실행으로 증거를 삼지 않았다(잘못된 빌드 측정 방지). 수렴 도구 결과의 `build_lag` 노트도 동일 내용.
- audit_multi의 codex 리뷰 기준: baseBranch(서버가 이 트리에서 재계산한 merge base `bc5817d758` 계열) — 카드 기여 전체가 검토 범위.

## Findings (레인 처분용 결함 목록)

- F1 [P2] [blocking] `internal/hook/quality/astgrep_gate.go:388` — 제외 조건이 절대경로 전체 세그먼트를 검사한다. 프로젝트 상위 디렉터리가 문자 그대로 `testdata`이면(예: `/tmp/testdata/proj`) 스윕이 0건, `RunAstGrepGateV2` 1단계 페어링 검사가 공소 통과하며, 무사유 `ast-grep-ignore`가 `SUPPRESSION_WITHOUT_REASON`를 우회한다. codex가 재현테스트+변경 전 대조군으로 검증, 이 감사관이 오버레이 프로브로 독립 확증. 필수 수정: `filepath.Rel(projectDir, path)`로 얻은 프로젝트 내부 상대경로 세그먼트에만 제외를 적용하고, 상위 `testdata` 경로에서도 생산 파일이 스윕되고 위반이 보고되는 회귀테스트를 추가할 것.
- F2 [P3] [optional] `internal/hook/quality/astgrep_gate.go:388` — 동일한 상위-이름 충돌 형태가 `vendor`/`.git`/`node_modules`/`__pycache__`에 대해 main에 이미 존재(이 카드가 도입한 것 아님 — 카드는 기존 조건열에 `testdata`만 추가). F1 수정으로 조건 자체를 상대경로화하면 이 클래스 전체가 함께 수리된다. 별도 조치 불요(단, F1 수리 시 회귀테스트에 이름 몇 개를 곁들이면 저렴).

## Gaps

- RED 측(수정 전 커밋에서의 실패) 재실행 안 함 — 고정 커밋 트리 유지 목적. RED→GREEN 기록은 레인 관측치; 이 감사관은 GREEN 측(3테스트 PASS, 실 fixture 스윕 제외)만 독립 관측.
- 저장소 전체 스위트·`-race` 미실행(슬롯 미확보 규칙; 범위 프로브만). 전역 판정은 통합 브랜치 CI 소관.
- main 선행 errcheck-3(`internal/cli/factory_run_adopt.go`, t1864/lane-31) 미재현 — 파일 존재(8605바이트)와 이 카드가 internal/cli 미접촉(diff 3파일)만 확인, golangci-lint 실행 안 함.
- 라이브 `moai gate` 종단 패리티 데모 미실행(위 errcheck-3 블록 + 설치 바이너리가 수정 이전 빌드라는 이중 사유).
- 카드 소속 SPEC/progress.md 부재(그레프 0건; SubagentStart 훅의 SPEC-ZONE-SHELL-PARSING-001는 t1689 언급 0건으로 무관) — 보고할 Binding run conditions 없음.
- codex 지적 원문은 한국어로 반환됨; 본문 요약 번역은 이 감사관 작성.

## Residual-risk

- (a) 트리에서 확인: 이름이 `testdata`인 디렉터리는 이제 페어링 단속에서 면제(findings 측 제외와 동일 규칙, Go 도구체인 관례 정합) — `testdata` 아래에 생산 코드를 두는 프로젝트는 해당 경로 단속 상실. 카드 승인 옵션 (a)의 의도된 범위.
- (b) 라이브 종단 데모는 main 선행 errcheck-3(t1864) 착지 전까지 golangci-lint 단계에서 블록. 카드 결함 자체는 단위 수준 입증.
- (c) 저장소 전역 테스트 판정은 통합 브랜치 CI 소속.
- (신규, FAIL 동인) F1 — 상위 `testdata` 경로 전체 프로젝트 단속 공소화; 수정 전까지 잔존.
- 수렴 도구 residual_risk_note 원문: "required-backend FAIL: codex". 설치 moai 바이너리는 HEAD보다 422커밋 뒤처진 빌드 — moai CLI 산출물은 구형 코드를 서술함(§2.2 공개).
- codex 내부 분열: r1(수렴, 계획 적용)은 재현 동반 FAIL, r2(정지 가드 구제 호출)은 무실행 정적 PASS — 실행 재현이 없는 통과는 실행된 재현을 상쇄하지 못한다. 영수증 2건 모두 위 기재; 최종 줄 인용은 2건 모두.
- 본 판정은 레인이 audit-plan 결과 점검(다이제스트 → `.moai/state/audit-plan-result.json` → `moai verify audit-plan --result-file`)을 통과시키는 절차와 별개로, required 게이트 실패라는 측정값에 기반한다. 레인은 본 파일을 읽고 커밋할 것(감사관은 커밋하지 않음).
