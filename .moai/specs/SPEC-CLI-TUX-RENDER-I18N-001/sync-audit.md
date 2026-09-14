# Sync-Audit Report — SPEC-CLI-TUX-RENDER-I18N-001 (card t756)

- Auditor: sync-auditor (독립 판정, orchestration 결과 미인용)
- 감사일: 2026-09-14
- 대상 트리: `WT-tux-render` @ `3edcd2391` (base `a404132e7`, tree clean)
- Harness: standard (Tier M)
- **Overall Verdict: FAIL** — 점수 95.7/100. 차단 결함 1건(F1, 문서 정확성). **코드 구현 자체는 PASS 등급**이며, FAIL 의 원인은 sync 산출물(CHANGELOG)의 사실 오류다. 수리는 문장 1개 수정 규모다.

---

## 1. 판정 요약

구현은 우수하다. AC-TRI-001..009 전건을 본 감사가 직접 재실행하여 PASS 를 관측했고, REQ-TRI-008 의 zero-repair fence 는 커밋 범위 전체에서 빈 diff 로 확인됐으며, M3 의 유일한 생산 코드 수리(acceptEdits 고지문 현지화)는 4-로케일 표 완비·앵커 토큰 보존·호출부 로케일 연결까지 모두 올바르다. 3-phase close 도 정확히 sync 커밋(`4c4534419`)에 실렸다.

FAIL 은 한 가지다. **CHANGELOG 엔트리가 고지문의 명령 표면과 출력 스트림을 모두 잘못 기술했다.** 엔트리는 "`moai init`/`update` 가 stderr 에 출력하는 고지문"이라 쓰고 있으나, 본 트리의 기계적 사실은 (1) 고지문은 `moai profile setup` / `moai profile --setup` 에서만 발화한다 — `profile.go:59-63` 이 REQ-ITI-001 에 따라 init/update 는 프로필 진입 자체가 없다고 명시하고, (2) 출력 스트림은 stdout 이다 — 호출부가 `cmd.OutOrStdout()`(`profile_setup.go:387`)이고, 생산 코드 전체에서 `SetOut` 리다이렉트가 0건이며, 흡수 통합 테스트가 `run.stdout` 에서 고지문을 관측한다. spec.md REQ-TRI-006 본문도 "표준 출력 고지문"으로 바르게 기술하고 있어, CHANGELOG 는 닫는 SPEC 자신의 문장과도 모순된다.

## 2. Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence (본 감사 직접 관측) |
|-----------|--------|-------|---------|------------------------------|
| Functionality | 40% | 100/100 | PASS | `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run 'TestPtyCapture_ConfirmGapBudget\|TestConfirmAlignmentSweep\|TestPtyCapture_I18nKoSweep' -count=1` → PASS (sweep "2 huh.NewConfirm site(s), all left-aligned" + mutant 거부 관측, gap budget 4 서브테스트, ko sweep 4 서브테스트). `go test ./internal/cli/ -run 'TestEmitAcceptEditsConfirmationAnchor\|TestHuhV1NonRegressionGuard' -count=1` → 양건 PASS ("swept 255 production source files + go.mod", mutant 2경로 거부 관측). `go test ./internal/cli/wizard/ -count=1` → `ok … 3.920s`. `TestLayout_NoBlankBetweenFields`·`TestOptionDescriptionColumn_DisplayWidthAligned` -v PASS. `TestPtyCapture_SkipWithoutGate` 게이트 부재 환경에서 관측 가능한 SKIP 확인. zero-repair diff: `git diff --stat a404132e7..HEAD -- internal/cli/wizard/wizard.go internal/cli/wizard/downgrade_confirm.go` → **빈 출력**. 기준 프레임 8매 실재(`evidence/baseline/`) + `baseline-downgrade-confirm-ko.txt` 직접 판독(제목/설명/빈 1행/`예 아니오` 좌측 정렬). |
| Security | 25% | 100/100 | PASS | 로케일 표는 정적 `map[string]string`, 출력이 `fmt.Fprintln(out, txt)` — 포맷 문자열 없음. `locale` 인자는 맵 키 선택 또는 영어 폴백으로만 소비되고 경로·명령·쿼리 구성에 쓰이지 않는다(주입 표면 없음). 시크릿·신규 입출력 표면 없음. `golangci-lint run ./internal/cli/...` → `0 issues`. |
| Craft | 20% | 95/100 | PASS | `go test ./internal/cli/wizard/ -cover -count=1` → `coverage: 93.6% of statements` (역치 85 이상). GOOS=windows 빌드 ok. 가드 3종 모두 비공공(non-vacuous) 설계 — 스윕 가드는 0-site 전제 단언("found 0 … asserts nothing")과 in-test mutant 경유, huh v1 가드는 go.mod·import 이중 mutant. 편차 사항(M2 래퍼 미실행, 전제값 2→1)이 §E.3 deviations 에 정직히 기록됨. 감점: 앵커 테스트 말단의 `"no"` 단언이 공허 F3. |
| Consistency | 15% | 78/100 | PASS(조건부) | 커밋 전체가 Conventional Commits + card t756 명기, CHANGELOG 형식·위치·영어·trailer 는 형제 엔트리와 동일, 테스트는 기존 `ptycaptest` 관례 재사용, SPEC 산문 한국어+영어 식별자 준수. plan-audit iter2 PASS 1.0 — 미해결 결함 0. `.moai/reports/` diff 0, diff 파일 목록이 예상 범위와 정확히 일치. **감점: F1/F2 — CHANGELOG·progress.md §E.4·residual-defects.md 가 코드·SPEC 본문과 모순되는 스트림/표면 서술을 배포 기록에 남김.** |

가중합: 0.40×100 + 0.25×100 + 0.20×95 + 0.15×78 = **95.7/100**. Verdict 는 점수와 무관하게 차단 결함 존재로 **FAIL** (plan-audit iter1 의 "차단 결함은 러브릭 점수로 상쇄 불가" 원칙과 동일).

## 3. 감사 청구 7건 검증 결과

| # | 청구 | 판정 | 본 감사 관측 |
|---|------|------|--------------|
| 1 | AC 재실행 | **PASS** | §2 Functionality 증거 참조. 전부 본 감사의 자체 실행 출력이다. |
| 2 | zero-repair-diff | **PASS** | 두 파일 diff 빈 출력(exit 0). M2 래퍼는 실제로 실행되지 않았고 diff 어디에도 래퍼 타입이 없다 — fence 유지. |
| 3 | M3 품질 | **PASS** | 로케일 표 en(폴백 상수)+ko/ja/zh, 앵커 토큰 `acceptEdits`·`settings.local.json` 이 4 로케일 번역문 안에 그대로 존재, 호출부 `result.ConversationLang` 의 값역이 `languageOptionLabels`(`profile_options.go:14-19`)의 `en/ko/ja/zh` 와 정확히 일치, 미지 로케일 영어 폴백 + 단일 줄 출력 검사("unknown" 포함). 스트림·동작 변경 없음. |
| 4 | 3-phase close | **PASS** | `git show 4c4534419:.moai/specs/.../spec.md` → `5:status: completed` — 전환이 sync 커밋에 탑승. `sync_commit_sha: "4c4534419"` 백필 확인, §E.4 신호 완비. |
| 5 | CHANGELOG | **FAIL** | 형식·섹션·영어·trailer 는 적합하고 "render unchanged"·"93.6%"·"lint 0"·"9/9"·"GOOS=windows" 등 측정 청구는 본 감사 재측정과 일치. 그러나 첫 문장의 표면(`moai init`/`update`)+스트림(stderr) 서술이 관측 사실과 불일치 — F1. |
| 6 | 일관성 | **PASS(F2 예외)** | plan-audit 종결 확인, reports/ 무기록, stray 파일 0. 단 stderr 오기가 progress.md §E.4 와 residual-defects.md 4행에도 반복됨 — F2. |
| 7 | 보안 | **PASS** | §2 Security 증거 참조. 정적 표+Fprintln, 주입 표면 없음. |

## 4. Findings

- **F1** [major] [blocking] `CHANGELOG.md:12` — acceptEdits 고지문을 "`moai init`/`update` 가 stderr 에 출력"한다고 기술. 관측 사실: (1) 고지문은 `runProfileSetup` 경유 `moai profile setup`/`moai profile --setup` 에서만 발화 — `internal/cli/profile.go:59-63` 이 REQ-ITI-001 로 init/update 의 프로필 진입 부재를 명시하고 plan-audit 1차가 `init.go:617` 주석으로 대조 확인했음. (2) 스트림은 stdout — `internal/cli/profile_setup.go:387` `cmd.OutOrStdout()`, 생산 코드 `SetOut` 0건, `profile_setup_absorb_test.go` 가 `run.stdout` 에서 고지문 계수. spec.md REQ-TRI-006 은 "표준 출력 고지문"으로 바르게 기술. — Required fix: 해당 문장을 "`moai profile setup` 이 stdout 에 출력하는 고지문"으로 정정(표면·스트림 2토큰).
- **F2** [minor] [blocking] `.moai/specs/.../progress.md` §E.4 `canary_compliance_check`("undocumented stderr detail") 및 `evidence/residual-defects.md` 4행("acceptEdits stderr 고지문") — F1 과 동일 원인의 스트림 오기가 SPEC 기록에 반복. — Required fix: "stderr" → "stdout" 2곳 정정. (참고: 카드 판정문 원문 `.moai/reports/init-tui-audit-20260909.html` 에서 "stderr" 문자열은 검색 0건 — 이 오기의 출처는 카드 판정문이 아니다.)
- **F3** [minor] [optional] `internal/cli/profile_setup_acceptEdits_test.go:96-99` — `strings.Contains(strings.ToLower(buf.String()), "no")` 는 "Note:" 접두 자체가 "no"를 포함하므로 어떤 출력에도 통과하는 공허 단언이다. en 줄의 "no override written" 의미 사실은 어디에도 엄밀히 고정돼 있지 않다(위 3개 앵커 단언이 사실상 대신 지키고 있어 해롭지는 않음). — Required fix(선택): 단언을 `will be written` 같은 구체 프래그먼트로 교체.
- **F4** [minor] [optional] `internal/cli/launcher.go:1110-1126` — `syncPermissionModeToSettingsLocal` doc 주석 블록이 그대로 2회 반복돼 있다. **본 카드 diff 밖의 기존 결함**으로 범위 밖이며 수리 대상 아님 — 향후 별도 카드 후보로 기록만 남긴다.

## 5. Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: SPEC-CLI-TUX-RENDER-I18N-001 의 코드 구현은 AC-TRI-001..009 전건을 충족하며 REQ-TRI-008 fence 가 유지됐다. 단, sync 산출물 CHANGELOG 가 고지문의 명령 표면과 출력 스트림에 관한 관측 불가능한(거짓) 청구를 1문장에 2건 담고 있어 배포 기록으로는 부적합하다.
- **Evidence**: §2 표의 전 셀(본 감사의 자체 명령 실행 출력 발췈) + §4 결함별 판독 근거. 핵심 관측 — 테스트 4회 재실행 전부 PASS, zero-repair diff 빈 출력, `git show 4c4534419:…spec.md` status 전행 확인, `golangci-lint run ./internal/cli/...` → `0 issues`, `-cover` → `93.6% of statements`, `GOOS=windows go build ./internal/cli/` 성공, 기준 프레임 8매 실재 및 ko 프레임 1매 직접 판독.
- **Baseline-attribution**: 모든 수치는 이번 실행, 이 트리(`WT-tux-render` @ `3edcd2391`)에서 본 감사가 실행한 명령의 관측 출력이다. §E.3 이 기록한 93.6%·lint 0·windows pass 는 본 감사의 재측정과 독립 일치했다.
- **Gaps**: (1) `internal/cli` 전체 스위트·`-race`·linux 크로스빌드 미실행 — 공유 머신 규율상 CI 몫(§E.3 이 linux 를 not-run 으로 정직 기록). (2) `internal/cli` 패키지 커버리지 미측정(§E.3 기록과 일치). (3) 기준 프레임 8매 중 ko downgrade 1매만 직접 판독하고 나머지 7매는 존재·크기 확인만. (4) AC-TRI-002/008 의 mutant 관측은 검사 자체의 in-test mutant 경로를 통한 것으로, 본 감사가 외부 mutant 를 별도 작성해 관측한 것은 아니다(스윕 함수가 데이터 주입형이라 in-test 경로가 곧 생산 경로다). (5) 선행 SPEC(t586) 산물의 동작 재검증은 본 SPEC 범위 밖으로 전제로만 사용.
- **Residual-risk**: (1) huh 버전 상승 시 `field_confirm.go:261-263` 인용과 `alignmentWindow` 스윕 휴리스틱이 스테일해질 수 있다(양가드 모두 0-site/무매칭 전제 단언을 갖춰 공공 통과는 차단돼 있다). (2) F3 의 공허 단언 때문에 en 고지문의 "no override" 의미가 느슨하게 고정돼 있다. (3) F1 수리 전까지 공개 CHANGELOG 가 사용자에게 잘못된 명령·스트림을 안내한다 — 수리가 창에 잡히기 전 develop 으로 병합되면 오기가 배포 노트에 잔존한다. (4) `ConversationLang` 값역이 settings 스키마(`FieldOptionDefs("conversation_lang")`) 의존 — 신규 로케일 추가 시 맵 미정의 키는 영어 폴백으로 조용히 동작한다(의도된 우아한 강등이나 로케일 추가 카드에서 표 동기화가 전제된다).

## 6. 수리 경로 (re-audit scope)

F1+F2 는 동일 원인의 문서 정정이다 — CHANGELOG 문장 1개(표면+스트림), progress.md §E.4 1단어, residual-defects.md 1단어. 수리 커밋 후 재감사는 **이 결함 델타만** 재판정한다(전수 재감사 아님). 코드·테스트·AC 판정에는 영향이 없으므로 §2의 Functionality/Security/Craft 재측정은 불요다.
