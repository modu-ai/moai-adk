# sync-audit.md — SPEC-PRECOMMIT-GATE-SCOPE-001 (card t461)

- auditor: sync-auditor (독립 재검증 — 레인 보고를 전제로 삼지 않고 재유도)
- 측정 트리: `WT-precommit-gate-scope` @ `914cfce8e` (본 워크트리, 2026-09-03)
- 감사 범위: run 커밋 `8347656b8`(M1) / `b18de874c`(M2) / `8ab11ed99`(M3+M4) / `3f6a3c4c2`(M6), catalog revert `0d26f8a00`, sync `b0cd51195` + backfill `914cfce8e` — base `a7f30b373` 대비

---

## Verdict

**PASS**

must-pass 방화벽(Functionality + Security) 독립 통과, blocking finding 0건. 전수 AC 재검증에서 레인 보고와 배치되는 관측은 없다. 차이점은 §결함 목록 F1(문서 수치 부정확, optional)뿐이다.

## 차원 점수 (harmonic mean)

| Dimension | Score | Verdict | Evidence (발췌 — 기계 검증 출력) |
|-----------|-------|---------|----------|
| Functionality (40%) | 96 | PASS | 아래 AC 매트릭스 전수 — 10/10 PASS, 전부 본 감사 자체 실행 |
| Security (25%) | 96 | PASS | 신규 입력 표면 0 — 마커 비교 `os.Getenv(config.EnvPreCommitMarker) == "1"` 상수 비교, 훅 안내문 전부 상수 printf, seam 기록은 `sectionRootKeys` 화이트리스트gate 등록으로 경로 고정. 템플릿 중립성 grep(SPEC ID/날짜/SHA/카드 id) 0 히트 |
| Craft (20%) | 93 | PASS | `go vet ./internal/{cli,config,hook,settings,web}/...` exit 0. `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 모두 ok. RED-GREEN 4층(config/runner/twin/web) 증거 + 뮤탄트 프로브 + fixture 정직성(must-replace) 확인. 레인 전량 internal/cli `ok … 429.520s`(run2 로그 재독) |
| Consistency (15%) | 92 | PASS | Conventional Commits + 카드 id + `🗿 MoAI` 트레일러 전 커밋 확인. `gateFields()`가 `mcpFields()`/`crossSessionFields()` 선례와 동일 패턴, `EnvPreCommitMarker` envkeys 상수 패턴 준수, 러너/타입 코멘트는 영문(해당 파일 관례), schema.go 신설 코멘트는 국문(해당 파일 관례) — 파일 내 일관 유지. 감점: F1 |

**harmonic mean = 94.1**

## AC 매트릭스 (전수 재유도 — 판정 명령 병기)

| AC | 상태 | 본 감사의 판정 근거 (명령 → 관측) |
|----|------|------|
| AC-001 (REQ-003 안내 5문자열) | **PASS** | `grep -c` 템플릿 훅: `.moai/config/sections/gate.yaml`=3, `gate.pre_commit.enabled`=3, `gate.enabled`=1, `gate.skip_tests`=1, `gate.disabled_steps`=1, `SKIP_MOAI_PRECOMMIT=1`=5. 실제 훅 stderr 전수 단언: `go test ./internal/cli/ -run 'TestPrecommitE2E' -count=1` → `TestPrecommitE2EOptInBlocksAndGuides PASS` (6문자열 전부 출력 존재) |
| AC-002 (무관한 실패가 커밋을 안 막음) | **PASS** | `go test ./internal/cli/ -run 'TestPrecommitE2EDefaultAllowsUnrelatedCommit' -count=1` → PASS(실제 git commit + fixture). 공허 초록 방지: 뮤탄트 프로브 — 구 훅(`a7f30b373`, 마커 export 없음)을 같은 시나리오에 넣으면 exit 1(`.moai/reports/t461/mutant_probe_old_hook_blocks.txt` 재독) — 초록은 마커 export에 의존함이 입증됨 |
| AC-003 (twin byte identity) | **PASS** | `go test ./internal/cli/ -run TestPreCommitTemplateMatchesConstant -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 1.361s` |
| AC-004 (gate.yaml의 update 생존) | **PASS** | 전제 반전 검증: `update_template_sync.go`에서 `BackupMoaiConfig`(:251/:403) → `CleanMoaiManagedPaths`(:315) → `RestoreMoaiConfigRetained`(:482) 순서 직독 — 소거-전 백업 + 3-way 병합 복원이 기존 장치로 존재. `go test ./internal/cli/ -run 'TestUpdateGateYAML' -count=1` → 2건 PASS(실제 `runTemplateSyncWithReporter` 경유, 손편집값+web 작성값+사용자 키/주석 유지, 신규 키 전달) |
| AC-005 (t237 소관 무변경) | **PASS** | `git diff a7f30b373 3f6a3c4c2 -- internal/template/templates/.git_hooks/pre-commit` 전수 판독 — 변경은 헤더 주석 + heavy-gate 블록뿐. `go vet`/`STAGED_GO`/`BT_TAGS`/`PKGS`/`diff-filter` 줄은 ± 어느 쪽에도 없음 (grep 히트 3건 전부 헤더 코멘트의 "go vet" 문자 언급) |
| AC-006 (opt-in 실행·차단 + 단독 불변) | **PASS** | `go test ./internal/cli/ -run 'TestRunGate|TestPrecommitE2E' -count=1` → `TestRunGatePreCommitOptInRunsHeavyGate PASS`, `TestRunGatePreCommitOptInPassingProjectPasses PASS`, `TestPrecommitE2EOptInBlocksAndGuides PASS`(차단+안내), 단독 불변: `TestRunGateStandaloneUnchanged PASS` + `TestRunGateGateDisabledStillShortCircuitsUnderMarker PASS` |
| AC-007 (non-moai 무음 통과) | **PASS** | `go test ./internal/cli/ -run 'TestPrecommitE2ENoMoaiOnPathPasses' -count=1` → PASS |
| AC-008 (기존 설치 반영) | **PASS** | `go test ./internal/cli/ -run 'TestUpdateReplacesMarkerHook' -count=1` → PASS(교체된 훅이 `MOAI_PRECOMMIT=1 moai gate` 운반). 이 테스트의 RED(`gate_yaml_preserve_run1.log` 82행 FAIL)→GREEN 궤적 확인 |
| AC-009 (마커 하 기본 OFF 러너 계약) | **PASS** | `TestRunGatePreCommitMarkerSkipsHeavyGate PASS`(exit 0 + remedy 안내, "quality gate failed" 부재 — heavy 미실행) vs `TestRunGateStandaloneUnchanged PASS`(마커 없으면 기존 계약대로 실패) — 대비가 REQ-001의 기계 증명. RED 선관측: `red_gate_precommit.log`(마커 하에서 heavy 실행됐던 적색) 재독 |
| AC-010 (moai web 편집 표면) | **PASS** | 코드 직독: `SectionGate` 신설+`AllSections()`/`SchemaSectionIDs()` 등록, `seamField(SectionGate,"gate",TypeBool,"gate","pre_commit","enabled")`, `sectionRoutes["gate"]=RouteSeam`, `sectionRootKeys["gate"]`, `consoleTabs()` 13번째 탭 + `schemaSectionMetas()` 패널, i18n 4-locale 키(en/ko/ja/zh 각 4키). `go test ./internal/settings/ -run Gate` 3건 PASS / `go test ./internal/web/ -run Gate` 4건 PASS — `TestApplySchemaEditsGateSeamRoundTrip`(주석·미모델링 키 보존) + `TestGateSavePath`(gate.yaml만 기록) + `TestGateI18nKeysInAllLocales`. RED: `red_settings_gate.log`(undefined: SectionGate), `red_web_gate.log`(4 probe 미렌더) 재독 |

## 결함 목록

- **F1** [Low] [optional] `.moai/specs/SPEC-PRECOMMIT-GATE-SCOPE-001/progress.md` §E.2 동기화 산출물 절 - docs-site 섹션 수 치가 부정확하다. 기록은 "ko 8→9, en/ja/zh 9→10"이나 본 감사 실측은 `grep -c '^## '` 4파일 모두 **사전 7 → 사후 8**(8/8/8/8 동일). §E.4의 패리티 기록("4파일 × 2히트 동일")과 모순되는 유일한 지점이며, AC-relevant 속성(4-locale 패리티 + 페이지당 신설 섹션 1개)은 실측상 충족된다. - Required fix: §E.2의 4개 행을 `7 → 8`로 정정 (또는 실측 8/8/8/8 등가 표기로 대체). 후속 커밋 1건 소관.
- **F2** [Info] [optional] `.moai/reports/t461/e2e_precommit_run1.log`의 RED는 하네스 결함 적색(`fatal: not a git repository`)이지 AC-002의 의미적 적색이 아니다. 의미적 공허성 방지는 뮤탄트 프로브(구 훅 차단)가 대신 담당 — 증거 품질 기록용으로 남긴다. 수정 불요.
- **F3** [Info] [optional] `progress.md` §E.3 `run_commit_sha: pending-backfill-run` 플레이스홀더 잔존 - 다수 완료 SPEC(`SPEC-HARNESS-EVOLVE-001` 등)이 동일 상태로 종결한 리포 전역 관행이고 M1..MN SHA는 §E.2 `m1_to_mN_commit_strategy`에 기록돼 있다. 본 카드의 결함 아님.
- **F4** [Info] [optional] acceptance.md D.3의 "make build 성공"이 본 트리에서 미충족 - `agents-emit-check`가 t443 소관 `sync-auditor.toml` 드리프트로 실패(`make_build.log`에 원문 기록, 수리 금지 지시 준수). 대체 검증(go build + GOOS=windows + catalog 직접 재계산)이 §E.2에 기록됨. t443 귀속 — 본 카드 감점 아님.

## Gaps (본 감사가 관측하지 않은 것)

- 전체 스위트(`go test ./...`) 로컬 실행 — 리포 규율상 CI 소관. 본 감사는 표적 `-run` 필터 + 레인의 전량 internal/cli 로그(`ok … 429.520s`, run2) 재독으로 대체했다.
- 실제 브라우저에서의 `moai web` UI 조작 — Go/템플릿 층(패널 렌더·저장 경로·i18n 키)까지 검증했으며 라이브 UI 세션은 아니다.
- 실제 사용자 설치에 대한 라이브 `moai update` — 테스트가 `t.TempDir()`에서 실제 파이프라인(`runTemplateSyncWithReporter`)을驱动하는 것으로 대체됐다.
- hugo 배포(deploy) — 빌드(exit 0, WARN/ERROR 0행)까지 검증, 배포 표면은 별도.

## Residual-risk

- 훅이 자신이 호출하는 모든 `moai gate`에 `MOAI_PRECOMMIT=1`을 실는 구조상, 사용자가 셸에서 이 변수를 수출하면 opt-in 전 단독 실행도 skip 포스처를 따른다 — opt-in 전 기본 포스처와 동일 상태이고, opt-in(true) 시에는 분기 조건(`!cfg.PreCommitEnabled`)이 어차피 거짓이라 무차별. 특권 표면 없음.
- PCP-005(c) 플립: 새 훅 본문을 실은 첫 릴리스가 컷되기 전까지 v3.1.2-era 훅은 `hookUserModified`(backup+notice)로 귀속된다. 재-pin 절차(`pinnedReleasedHookTag` 재지정 + 원 단정 복원)가 테스트 코멘트에 기록돼 있으나, 이는 리포 자동화 밖의 릴리스 체크리스트 항목으로 남는다.
- `runTemplateSyncAt` 드라이버가 `os.Chdir`을 쓰는 관계로 AC-004 테스트 패밀리는 직렬 실행이다(t.Parallel 없음) — 선례(llm.yaml)와 동일 트레이드로 의도됨.

## 유산·귀속 상태 (본 카드 감점 제외 확인)

- `TestAlwaysLoadedTokenBudget` 적색: 본 카드 diff가 `.claude/` 아래 0 파일(`git diff a7f30b373 3f6a3c4c2 --stat` 실측) — 측정 표면이 base와 동일, 귀속 타당.
- `make build` agents-emit 드리프트: t443 소관, `make_build.log` 원문 기록 확인. 본 카드는 수리하지 않았다(지시 준수).
- catalog revert `0d26f8a00`: 본 트리의 sync-auditor 해시 `f1b4487f…`가 `origin/develop` 값과 동일함을 직접 비교 확인. t443의 `545d03d9…`는 배제, t191의 moai 스킬 갱신(`f005e873…`)은 유지 — 커밋 메시지의 주장과 일치. catalog.yaml에 `.git_hooks`/`gate.yaml` 항목이 없음도 확인(본 카드 템플릿 2파일은 catalog 재계산 대상 아님).

## 종합

결정된 메커니즘(축 (b) + 메커니즘 1)이 계약대로 착지했다: 러너 분기점 1개(`internal/cli/gate.go:81`, 마커 게이트 + `!cfg.PreCommitEnabled`), `config.Enabled` 무플립, 기본 false(`internal/config/defaults.go` `NewDefaultGateConfig`), twin byte identity 유지, t237 소관 구간 무침범, `moai web` 표면은 seam 화이트리스트 등록 2건(누락 시 소리 내는 실패)까지 갖췄다. AC-004는 카드가 가정한 wipe 결함이 존재하지 않음을 실측으로 반전시키고 keep+pin 테스트로 계약을 고정한, 전제를 기계로 반증한 모범 사례다. 유일한 자기 결함은 progress.md §E.2의 문서 수치(F1)로 optional이다.
