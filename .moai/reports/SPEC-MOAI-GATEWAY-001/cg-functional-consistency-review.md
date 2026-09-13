# CG 일반 이전 흐름과 공개 안내의 기능 대조

## Claim

**Scoped functional verdict: PASS.** PFR-F1 delta 이후 선택 범위로 현재 CG 폐기 entry, 명시적 설정 이전, 렌더된 template, 네 언어의 공개 이전 페이지를 대조했다. 이 범위에서 새 blocking finding은 관측하지 않았다.

- 폐기된 `cg`는 root command에 등록되지 않으며 실행 entry·spawn 경계에서 이전 안내 오류로 끝난다. 이전 CG 역사 레코드는 읽은 바이트를 바꾸지 않는다.
- 임시 프로젝트의 legacy CG 설정은 현재 런처의 준비된 실행 경계를 통과하지 못한다. 명시적으로 `claude-only`로 이전한 대조군은 동일 경계에 도달한다. 실제 Claude·tmux를 실행한 시험이 아니라 주입한 실행 counter를 판정한 것이다.
- 네 언어 `cg-mode.md`에 실린 명령 네 개를 **문서에서 읽어** 각각 새 임시 프로젝트의 `newMigrateCGCommand.Execute`에 넣었다. preview는 원본을 보존하고, `--target claude-only --apply --accept-role-change`는 적용된다. 기존 영구 시험은 역할 변화 수락이 빠진 적용과 `claude-glm` 적용이 오류임을 판정한다.
- 새 Hugo 렌더의 네 언어 페이지도 실행한 네 명령을 그대로 표시하며 `tmux-env-security` anchor를 보존한다. 실제 embedded template 렌더의 이전 안내·일반 harness 모드·독립 감사 경로 유지 시험도 통과했다.

전체 CG-RETIRE SPEC 완료, 보안, 모든 문서 의미, 실제 설치·provider 동작에 대한 판정은 아니다. 제품·문서 원본을 수정하지 않았다.

## Evidence

현재 WT에서 기존 기능 시험을 읽고 필요한 항목만 선택했다.

```text
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-cg-functional-review-home GOCACHE=/tmp/gateway-protocol-review-cache go test ./internal/cli -run '^TestCGMigrationPreviewApplyAndGuards$|^TestCGRetired|^TestCGRetirement' -count=1 -v -timeout 30s
=== RUN   TestCGRetiredCommandHasNoRootRegistration
--- PASS: TestCGRetiredCommandHasNoRootRegistration (0.00s)
=== RUN   TestCGRetiredEntryAndModeHaveZeroEffects
--- PASS: TestCGRetiredEntryAndModeHaveZeroEffects (0.00s)
=== RUN   TestCGRetirementHistoricalRecordRemainsReadable
--- PASS: TestCGRetirementHistoricalRecordRemainsReadable (0.00s)
=== RUN   TestCGRetirementCompleteEntryShapesAndCounters
--- PASS: TestCGRetirementCompleteEntryShapesAndCounters (2.33s)
=== RUN   TestCGRetirementHelpDoesNotOfferCG
--- PASS: TestCGRetirementHelpDoesNotOfferCG (0.00s)
=== RUN   TestCGRetirementLiveHelpHasNoLaunchRecommendation
--- PASS: TestCGRetirementLiveHelpHasNoLaunchRecommendation (0.00s)
=== RUN   TestCGRetiredSpawnBoundaryDoesNotOpenWindow
--- PASS: TestCGRetiredSpawnBoundaryDoesNotOpenWindow (0.00s)
=== RUN   TestCGRetiredTopLevelDiagnostic
--- PASS: TestCGRetiredTopLevelDiagnostic (0.00s)
=== RUN   TestCGMigrationPreviewApplyAndGuards
--- PASS: TestCGMigrationPreviewApplyAndGuards (0.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.362s
```

Exit 0. runtime 시험과 template 시험은 독립 묶음으로 실행했다.

```text
$ GOCACHE=/tmp/gateway-protocol-review-cache go test ./internal/template -run '^TestCGEmbedded|^TestIsGLMBackend$' -count=1 -v -timeout 30s
=== RUN   TestCGEmbeddedRetirementPreservesRoutingAndAudit
--- PASS: TestCGEmbeddedRetirementPreservesRoutingAndAudit (0.00s)
=== RUN   TestIsGLMBackend
=== RUN   TestIsGLMBackend/team_mode=glm_(primary_moai_glm_signal)
=== RUN   TestIsGLMBackend/team_mode=cg_is_retired
=== RUN   TestIsGLMBackend/mode=glm_(defensive_dormant-field_OR)
=== RUN   TestIsGLMBackend/retired_cg_conflicts_with_mode=glm
=== RUN   TestIsGLMBackend/team_mode=claude_(legacy_non-GLM)
=== RUN   TestIsGLMBackend/team_mode=hybrid_(legacy_non-GLM)
=== RUN   TestIsGLMBackend/no_signal_(both_empty)
--- PASS: TestIsGLMBackend (0.00s)
    --- PASS: TestIsGLMBackend/team_mode=glm_(primary_moai_glm_signal) (0.00s)
    --- PASS: TestIsGLMBackend/team_mode=cg_is_retired (0.00s)
    --- PASS: TestIsGLMBackend/mode=glm_(defensive_dormant-field_OR) (0.00s)
    --- PASS: TestIsGLMBackend/retired_cg_conflicts_with_mode=glm (0.00s)
    --- PASS: TestIsGLMBackend/team_mode=claude_(legacy_non-GLM) (0.00s)
    --- PASS: TestIsGLMBackend/team_mode=hybrid_(legacy_non-GLM) (0.00s)
    --- PASS: TestIsGLMBackend/no_signal_(both_empty) (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.499s
```

Exit 0. 독립 문서 명령 probe:

```text
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-cg-functional-review-home GOCACHE=/tmp/gateway-protocol-review-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/cg-overlay.json ./internal/cli -run '^TestCGFunctionalPublishedMigrationCommands$' -count=1 -v -timeout 20s
=== RUN   TestCGFunctionalPublishedMigrationCommands
    cg_functional_doc_test.go:11: ko: four documented commands accepted; preview preserved source; explicit apply migrated
    cg_functional_doc_test.go:11: en: four documented commands accepted; preview preserved source; explicit apply migrated
    cg_functional_doc_test.go:11: ja: four documented commands accepted; preview preserved source; explicit apply migrated
    cg_functional_doc_test.go:11: zh: four documented commands accepted; preview preserved source; explicit apply migrated
--- PASS: TestCGFunctionalPublishedMigrationCommands (0.13s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.125s
```

Exit 0. overlay는 `cg_doc_commands_test.go`를 새 가상 시험 파일로 추가할 뿐 제품 파일을 대체하지 않는다. 임시 프로젝트는 `t.TempDir()`로 생성·정리되며 사용자 설정을 입력으로 쓰지 않는다.

```text
$ hugo --source docs-site --destination /tmp/gateway-cg-functional-docs --cacheDir /tmp/gateway-cg-functional-hugo-cache --noBuildLock --minify --quiet
```

Exit 0, stdout·stderr 빈 출력. 앞 문서 작성자의 생성 결과를 재사용하지 않고 새 destination에 빌드했다.

```text
$ python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/check_rendered_cg.py
ko: rendered four migration commands match executed source examples; historical anchor present
en: rendered four migration commands match executed source examples; historical anchor present
ja: rendered four migration commands match executed source examples; historical anchor present
zh: rendered four migration commands match executed source examples; historical anchor present
```

Exit 0. Python HTMLParser가 실제 생성 HTML의 `pre` 텍스트와 id를 읽는다. 비교 명령 집합은 기본 preview, claude-only preview, 수락 flag를 포함한 claude-only apply, claude-glm preview다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, 이번 부모 task 및 PFR delta에서 측정한 HEAD `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`.
- 주요 현재 source와 tests 및 네 언어 입력 9개 파일의 종료 시점 SHA-256은 `functional-review-iter2/cg-baseline-sha256.txt`에 보관했다. CG 검증 시작 전 모든 입력 해시를 찍지는 않았으므로, 이 목록으로 실행 전후 전체 파일 불변을 주장하지 않는다.
- 입력 보고서: `cg-retirement-runtime-verification.md`, `cg-template-verification.md`, `cg-docs-verification.md`. 보고서의 기존 수치를 이번 실행 수치로 전용하지 않았다.
- 읽은 source: `internal/cli/cg_retirement_test.go`, `migrate_cg.go`, `migrate_cg_test.go`, `internal/config/cg_migration.go`, `internal/template/cg_retirement_test.go`, 네 언어 `docs-site/content/<locale>/multi-llm/cg-mode.md`, 한국어 README의 이전 안내.
- probe·overlay·HTML 검사 스크립트는 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/`에 있다. 이 보고서는 `.moai/reports/SPEC-MOAI-GATEWAY-001/`에 별도로 수출한다.

## Gaps

- 실제 CLI 바이너리와 사용자의 프로젝트를 실행한 E2E는 아니다. `newMigrateCGCommand.Execute`와 준비된 launcher seam을 직접 호출했다.
- 156개 변경 문서 전체를 독립 의미 감사하지 않았다. 네 언어 핵심 이전 페이지의 실행 예시와 한국어 안내를 중심으로 대조했다.
- 실제 공급자, Claude/Codex, tmux 창, authentication, credential, 보안, 파일 장애 복구·동시 쓰기의 독립 재감사는 하지 않았다.
- 전체 template deploy/update/rollback과 사용자의 기존 로컬 수정 보존은 재검증하지 않았다.
- 브라우저 시각·접근성·전체 링크·외부 URL 상태를 검사하지 않았다. HTML 생성과 네 핵심 페이지의 command/anchor 판정만 수행했다.
- gateway/PICKER/TEAMMATE의 전체 출시 결합 조건과 CG-RETIRE 전체 AC 완료를 판정하지 않았다.

## Residual-risk

일반 실행 경계·명령 예시·렌더 출력이 맞아도 실제 사용자의 설치 상태와 다른 프로젝트 설정 조합은 추가 변수가 된다. 특히 `claude-glm` preview를 실제 혼합 팀 지원으로 읽지 않아야 한다. 현재 공개 한국어 문구와 실행 거절 시험은 이 제한을 같은 뜻으로 표현하지만 실제 TEAMMATE 기능의 지원 여부를 이번 검증으로 바꾸지 않는다.
