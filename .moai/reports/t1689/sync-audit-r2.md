auditor-model: glm-5.3-flash
verdict: PASS
audited_sha: a8b850889929b22615761eb5d52b08551afcf767
required_backend: codex
required_backend_fail: false
audit_receipt: rcpt-c327b8cf1044182d4c759897,rcpt-3dadb4cd3258066aa64ab31a
overall_score: 0.93

# t1689 sync 감사 (r2 델타, 2026-10-18)

감사 범위: 라운드 1 FAIL(0.82)의 차단 결함 F1에 대한 수리 커밋 1건 — `a8b850889`(수정 +16/−2, 회귀테스트 +126). 트리 `WT-gate-fixture-lint` @ `a8b850889`, 작업 상태 clean. 카드 기여 전체(main 병합 기준 이후)는 5파일 — 코드 `internal/hook/quality/astgrep_gate.go`, 테스트 `astgrep_gate_walk_testdata_test.go`(r1)·`astgrep_gate_walk_ancestor_test.go`(신규), 증거 문서 `card-review.md`·`sync-audit.md`(r1).

## 차원 점수

| 차원 | 점수 | 판정 | 근거 |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | F1 수리를 같은 트리에서 RED(수리 전 코드 오버레이)·GREEN(HEAD) 양방향으로 재관측. 게이트 진입면까지 회귀 커버 |
| Security (25%) | 95/100 | PASS | 무사유 `ast-grep-ignore` 우회로(공소 통과)가 닫힘. 신규 입력·비밀·주입 면 없음 |
| Craft (20%) | 88/100 | PASS | 양성 대조군 포함 테이블형 회귀, gofmt·go vet 청결. F3(주석-코드 불일치)·F4(.git 커버리지 공백) 감점 — 둘 다 optional |
| Consistency (15%) | 95/100 | PASS | findings 측 제외(`astGrepExcludedPathPatterns`, :220) 불변 확인, 문서 주석이 상대경로 의미론으로 갱신, Conventional commit(카드 id + 🗿 트레일러) |

조화평균 0.93. 신규 결함 2건(F3·F4)은 모두 optional — 소비 규율상 임의 조치이며 PASS를 뒤집지 않는다.

## Claim

1. 라운드 1 차단 결함 F1(제외 조건이 절대경로 세그먼트를 검사해 상위 디렉터명이 `testdata` 등이면 스윕 0건·페어링 검사 공소 통과)은 `a8b850889`에서 수리되었다 — 제외가 이제 `filepath.Rel(projectDir, path)` 세그먼트에만 적용된다(astgrep_gate.go:394-400).
2. 이 감사관이 동일 트리에서 양방향을 실행 관측했다: 수리 전 코드(c0784fe9a 추출, go overlay)에서 신규 테스트 전체가 공소-스윕 서명으로 FAIL, HEAD 코드에서 전부 PASS. 라운드 1의 "RED 미관측" Gap은 해소되었다.
3. 감사 계획은 codex 단독 required이며, 수렴 결과 `overall_verdict: pass`(요구 백엔드 통과, findings 0)이다.

## Evidence

감사 계획 — `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1689`:
```
"config_status": "ok", "model": "codex", "model_source": "config",
claude: "off" / codex: "required" (explicit) / glm: "off" — 모두 config.gates,
"cross_model_active": true, "enforced_required": ["codex"]
```

수렴(mcp__moai__audit_multi, gates 인자 없음, project_root=자기 트리, target=baseBranch, focus="relative-segment scoping of walk exclusions and ancestor regression coverage", session_id=t1689-sync-r2) 다이제스트 — 오케스트레이터 점검용 멤버만:
```
overall_verdict: "pass"   participant_count: 1   disagreement_flag: null
per_backend_verdicts: [ {backend: codex, gate: required, verdict: pass, findings: 0} ]
fail_open_backends: []   plan_source: "config"   gate_unmet: 없음
audit_receipt: rcpt-c327b8cf1044182d4c759897
```

정지 가드 구제 호출(mcp__moai__codex_audit 단일 백엔드, 동일 트리·target=baseBranch — review_base: main, merge base bc5817d758):
```
verdict: "pass"   findings: []   audit_receipt: rcpt-3dadb4cd3258066aa64ab31a
요약 원문: "수정이 필요한 결함은 발견하지 못했어. 지침에 따라 테스트·빌드·재현 명령은 실행하지 않았어." (정적 검토만, 무실행 자체 공개)
```
→ stop 후크의 판정 줄 상관 검증 실패에 따른 재호출이다. r1과 달리 이번 두 codex 다리(수렴·구제)는 모두 실행 증거와 방향이 일치하며, 무실행 정적 통과라는 성격은 은닉하지 않는다 — 판정의 무게는 위 감사관의 RED/GREEN 실행 재현이 진다.

GREEN(HEAD 코드, 이번 실행) — `go test ./internal/hook/quality -run 'TestWalkSourceFiles_ExcludedNameAncestorStillSweeps|TestRunAstGrepGateV2_TestdataAncestorStillBlocks|TestWalkSourceFiles_PositiveControl_NoExcludedAncestor' -count=1 -v`:
```
--- PASS: TestWalkSourceFiles_ExcludedNameAncestorStillSweeps
    --- PASS: .../testdata  --- PASS: .../vendor  --- PASS: .../node_modules  --- PASS: .../__pycache__
--- PASS: TestRunAstGrepGateV2_TestdataAncestorStillBlocks
--- PASS: TestWalkSourceFiles_PositiveControl_NoExcludedAncestor
```
패키지 전체(비-race): `ok  github.com/modu-ai/moai-adk/internal/hook/quality  27.457s`

RED 재관측(오버레이, /tmp만 기입, 트리 무변경 — 실행 후 `git status --porcelain` 빈 출력): `git show c0784fe9a:internal/hook/quality/astgrep_gate.go` 추출본 확인 — 386행 `parts := strings.Split(filepath.ToSlash(path), "/")`, `filepath.Rel` 없음(수리 전 형태). `go test -overlay=/tmp/t1689-r2-red/overlay.json ... -count=1 -v`:
```
astgrep_gate_walk_ancestor_test.go:60: walkSourceFiles skipped the project under a "testdata" ANCESTOR; got []
    (vendor / node_modules / __pycache__ 하위테스트도 동일 서명으로 FAIL — 4건 전부)
astgrep_gate_walk_ancestor_test.go:92: RunAstGrepGateV2 passed vacuously for a project under a testdata-named ancestor
--- PASS: TestWalkSourceFiles_PositiveControl_NoExcludedAncestor   (대조군만 통과 = 결함이 제외 조건에 있음을 분리)
FAIL
```
→ 같은 트리, 같은 테스트, 생산 코드 1파일만 교체: RED→GREEN이 관측값으로 확정.

Craft 보조: `gofmt -l internal/hook/quality/` 출력 없음, `go vet ./internal/hook/quality/` OK.

## Baseline-attribution

- 트리: 상기 `audited_sha`(감사 개시 시점 HEAD). 모든 go 테스트·오버레이 프로브·grep은 이 트리에서 이번 실행으로 관측.
- 판정 빌드(§2.2): 설치 moai는 `v3.2.0-rc.36-t1822c`, 커밋 `ee981a9326` — `git merge-base --is-ancestor ee981a932 HEAD` exit 0, `git rev-list --count ee981a932..HEAD` = 424. 설치 바이너리는 본 트리의 엄격 조상(424커밋 뒤짐)이므로 moai CLI 실행 산출물을 증거로 쓰지 않았다(audit-plan 동사는 설정 조회이며 `plan_source: config`로 확인). `go test`는 트리 소스를 새로 컴파일하므로 판정 빌드 = 트리 자체. 수렴 결과의 `build_lag` 노트(build_commit ee981a932)도 동일 내용.
- codex 리뷰 기준: baseBranch(서버가 이 트리에서 재계산) — 카드 기여 전체가 검토 범위.

## Findings (라운드 1 처분 + 신규)

- F1 [P2] [blocking] `internal/hook/quality/astgrep_gate.go` — **repaired-verified, 닫음.** 수리: 제외를 `filepath.Rel(projectDir, path)` 상대 세그먼트에만 적용(:394-400), 문서 주석 갱신(:380-384). 검증: 위 RED/GREEN 양방향 재관측 + codex(required) pass·findings 0.
- F2 [P3] [optional] 동일 상위-이름 충돌 클래스 — **repaired-verified(부분).** vendor·node_modules·__pycache__ 상위 하위테스트 PASS로 같은 상대경로화가 클래스를 함께 수리. testdata는 F1 그 자체. `.git`만 표에서 누락 → F4로 이월.
- F3 [P3] [optional] (신규) `internal/hook/quality/astgrep_gate.go:395-397` — Rel 오류 폴백이 자기 주석·커밋 메시지와 반대다. `rel = path`는 절대경로 세그먼트 매칭(수리 전 술어)을 복원하므로, 주석의 "파일을 제외하지 않는다 — fail toward more checking"과 반대(덜 검사)로 떨어진다. WalkDir 경로는 projectDir 접두사 일관이라 Rel이 실패할 수 없어(주석 자체가 인정) 도달 불가 — 임의 조치. 필수 수정(채택 시): `relErr != nil`이면 제외 루프를 건너뛰게 하여 기술된 의도와 코드를 일치시킬 것.
- F4 [P3] [optional] (신규) `internal/hook/quality/astgrep_gate_walk_ancestor_test.go:44` — 상위 테이블이 5개 제외명 중 4개만 커버(`.git` 누락). 테이블에 한 항목 추가로 해소.

## Gaps

- 첫 go test 시도가 워크트리 격리 가드에 거부됨(중괄호 그룹+리다이렉트 복합 명령 — "too complex to verify"). §3.1에 따라 거부를 여기에 기재하며, 단순 연쇄 명령으로 재실행하여 두 관측(대상 테스트 -v, 패키지 전체) 모두 완료했다.
- codex 다리가 실행한 명령 유무를 결과가 기술하지 않는다(summary "Verdict: pass" 한 줄). 판정의 무게는 이 감사관의 실행 재현(RED/GREEN)이 진다.
- 저장소 전체 스위트·`-race` 미실행(슬롯 규칙; 범위 프로브만). 전역 판정은 통합 브랜치 CI 소관.
- 라이브 `moai gate` 종단 패리티 데모 미실행 — 이중 사유: main 선행 errcheck-3 블록(아래 잔여 위험 b) + 설치 바이너리 424커밋 지연.
- 카드 소속 SPEC/progress.md 부재 재확인(`ls .moai/specs/ | grep -i t1689` 0건) — 보고할 Binding run conditions 없음.

## Residual-risk

- (a) 프로젝트 **내부**의 `testdata` 디렉터는 계속 페어링 단속 면제 — 의도된 설계이며 findings 측 제외(`astGrepExcludedPathPatterns`, :209/:220)와 동일 규칙 미러(불변 확인). 상대 세그먼트 매칭이라 상위 경로 문제와 무관하게 유지된다.
- (b) 라이브 종단 데모는 main 선행 errcheck-3(`internal/cli/factory_run_adopt.go`, 카드 t1864) 착지 전까지 블록 — 파일 존재 재확인(8605바이트), 본 카드 diff는 internal/cli 미접촉(5파일 목록 상단).
- (c) 저장소 전역 테스트 판정은 통합 브랜치 CI 소속.
- (d) 참여자 1(codex), `disagreement_flag: null` — "백엔드 합의" 주장은 성립하지 않으며, 이 감사관의 실행 재현이 보강 역할을 한다. r1의 codex 내부 분열(r1 재현 동반 FAIL vs 무실행 정적 PASS)은 이번 라운드에서 두 codex 다리(수렴 pass + stop 가드 구제 정적 pass)가 모두 실행 증거와 방향이 일치하여 해소. 구제 다리는 무실행 정적 검토임을 스스로 공개 — 상쇄 근거로 쓰지 않는다.
- (e) 설치 moai 바이너리는 HEAD보다 424커밋 뒤처진 빌드 — moai CLI 산출물은 구형 코드를 서술함(§2.2 공개).
- (f) 수렴 도구 residual_risk_note 원문: (빈 문자열).
- 본 판정은 레인이 audit-plan 결과 점검(다이제스트 → `.moai/state/audit-plan-result.json` → `moai verify audit-plan --project-root <toplevel> --result-file <경로>`)을 통과시키기 전까지 최종이 아니다. 레인이 본 파일을 읽고 커밋할 것(감사관은 커밋하지 않음).
