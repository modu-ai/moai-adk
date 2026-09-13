# SPEC Review Report: SPEC-INIT-TUX-I18N-001

- 카드: t586 · 감사 단계: plan · Iteration: 1/2 (Tier M 상한 2)
- 감사 대상 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `d0ec7921fd743608773bde75e5c7aa3cb0860644` (착수 시 `git rev-parse HEAD` 로 일치 확인)
- 대상 산출물: `.moai/specs/SPEC-INIT-TUX-I18N-001/{spec.md,plan.md,acceptance.md}` (Tier M 입력 3종), 재현 근거 `.moai/reports/t586/verdict.md`
- 작성자 추론 맥락은 M1 격리 원칙에 따라 배제했다. 호출문의 결정 D1~D6 과 미결 질문 Q1~Q4 는 판정 기준으로만 썼다.

**Verdict: FAIL**
**Overall Score: 0.67** (네 차원 조화평균, Tier M 통과 기준 0.80)

판정을 가르는 것은 MP-7(미해결 `[NEEDS CLARIFICATION]` 4건)이다. 점수와 무관하게 FAIL 이며, 그 밖에도 blocking 결함이 10건 있다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — 판정 층: 요구 층(`spec.md` §B). `grep -Eo 'REQ-ITI-[0-9]+' spec.md | sort | uniq -c` 결과 `REQ-ITI-001`~`REQ-ITI-015` 가 빠짐없이 연속이고 3자리 패딩이 일정하다. 정의 줄은 번호마다 하나다(`spec.md:118-144`). 여러 번 세어지는 번호(001·005·007·008)는 본문 교차 언급이다.
- **[PASS] MP-2 GEARS 형식** — 판정 층: 요구 층만. AC 의 Given-When-Then 은 검증 층 형식이라 여기서 채점하지 않았다. 15개 REQ 모두 다섯 패턴 중 하나에 들어간다. Event-driven 은 `spec.md:118,128,132`, State-driven 은 `:120,144`, Ubiquitous 는 `:124-127,133,138,139`, Unwanted 는 `:134`("shall not contain"), Where 는 `:140` 이다. 라벨 표기 문제 두 건(D16)은 형식 위반이 아니라 optional 로 분류했다.
- **[PASS] MP-3 YAML frontmatter** — `spec.md:2-13` 에 12개 필드가 모두 있다. `version: "0.1.0"` 은 따옴표 semver, `status: draft` 는 enum 값, `created`/`updated: 2026-09-11` 은 ISO 날짜, `priority: P1`, `lifecycle: spec-anchored`, `tags` 는 쉼표 문자열이다. 거부되는 별칭(`created_at`·`labels` 등)은 없다. `tier: M`·`related_specs` 는 선택 필드다. 트리에서 빌드한 바이너리의 lint 도 frontmatter 오류 0 을 보고했다(아래 증거 E-3).
- **[N/A] MP-4 언어 중립성** — 대상은 `internal/cli`·`internal/cli/wizard` Go 코드뿐이고, `internal/template/templates/**` 는 바뀌지 않는다(`spec.md:110`). 16개 프로그래밍 언어 도구 열거가 필요한 SPEC 이 아니다.
- **[PASS] MP-5 D7 교차 SPEC** — 참조 SPEC 7개가 모두 존재하고 모두 `status: completed` 다(증거 E-4). retired·superseded·archived 참조가 없으므로 D7 BLOCKING 은 없다. 다만 **참조 목록에 없는** 완료 SPEC 과의 계약 충돌이 따로 있다(D2, 일관성 결함).
- **[PASS] MP-6 D8 크로스 플랫폼** — `grep -c syscall` 결과 spec.md·plan.md·acceptance.md 모두 0 이다. D8-4 에 따라 자동 PASS.
- **[FAIL] MP-7 명확화 게이트** — `plan.md:103-106` 에 `[NEEDS CLARIFICATION: …]` 4건이 남아 있다(Q1 대화 언어 이중 질문, Q2 도움말 문장형 대 키별 라벨, Q3 다운그레이드 확인창 언어 우선순위, Q4 흡수된 프로필 위저드의 단계 표시). `research.md` 는 없다. 명확화 게이트 결함으로 `## Defects Found` D1 에 critical 로 옮겼다.

---

## Category Scores (0.0-1.0, rubric-anchored)

| 차원 | 점수 | 기준 구간 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분의 REQ 는 해석이 하나다. 다만 REQ-ITI-009 "resolved from existing persisted configuration"(`spec.md:132`)은 우선순위를 정하지 않고(Q3), REQ-ITI-005 "every option value set from the shared settings schema"(`spec.md:125`)는 `model_policy` 예외(`acceptance.md:29`)와 어긋난다(D5). |
| Completeness | 0.75 | 0.75 | 필수 절(HISTORY `spec.md:22`, 배경 §A, 요구 §B, 제외 범위 `### Out of Scope — …` 6개 `spec.md:158-183`)과 frontmatter 는 갖췄다. 그러나 v1 흡수 조사가 소스 스캔 가드 6건과 `schemaSelectOptions`·`huh.ErrUserAborted` 를 빠뜨렸고(D3·D4), `progress.md` 가 없다(D11). |
| Testability | 0.50 | 0.50 | pty AC 에 양의 존재 조건·캡처 시점·실패 경로 정리 검사가 없다(D8). AC-ITI-008 은 존재하지 않는 기존 테스트에 기댄다(D7). AC-ITI-009 는 예외 목록이 정의되지 않았고(D13), AC-ITI-010 은 언어 우선순위를 가르지 못한다(D10). |
| Traceability | 0.75 | 0.75 | 모든 REQ 에 `maps` 선언이 있고 고아 AC 가 없다(`acceptance.md:72`). 트리 빌드 lint 는 커버리지 경고 0 이며, 그 검사가 공허하지 않다는 것은 뮤턴트로 확인했다(E-3). 다만 REQ-ITI-012(다운그레이드 확인창)·REQ-ITI-013(프로필 위저드)·REQ-ITI-010(다운그레이드 `ja`/`zh`)은 AC 가 일부만 덮는다(D9). |

조화평균 = 4 / (1/0.75 + 1/0.75 + 1/0.50 + 1/0.75) = 0.667.

---

## Defects Found (structured defect-list)

D1. MP7-CLARIFY — plan.md:L103-106 — `[NEEDS CLARIFICATION]` 4건이 미해결이다(Q1~Q4). 이 네 답은 REQ-ITI-001·008·009·010·011 과 골든 출력을 직접 바꾼다. 명확화 게이트 결함. — Severity: critical — Class: blocking — Required fix: 리드·운영자 답을 받아 각 답을 spec.md REQ, plan.md 마일스톤, acceptance.md AC 에 반영하고 마커 4개를 지운다.

D2. CN-XSPEC-TUIM — spec.md:L15, L74, L96-106; plan.md:L83 — 이 SPEC 은 `huh_theme.go`(`moaiHuhTheme`/`moaiHuhStyles`)를 삭제한다. 그런데 완료 상태인 SPEC-CLI-TUI-MODERNIZE-001 은 두 팩토리의 존속을 현행 계약으로 갖고 있다. REQ-TUIM-040("two separate factories"), REQ-TUIM-045("indirection variables"), AC-TUIM-026(`grep -n "func moaiHuhStyles" internal/cli/huh_theme.go` → 1 hit), AC-TUIM-029(`var huhThemeIsDark` → 1 hit)가 그것이다. `related_specs` 와 §A.6 관계 표 어디에도 이 SPEC 이 없고, 대체 기록도 없다. status 가 completed 라 D7 BLOCKING 규칙에는 걸리지 않지만 교차 SPEC 계약 모순이다. — Severity: major — Class: blocking — Required fix: `related_specs` 와 §A.6 에 SPEC-CLI-TUI-MODERNIZE-001 을 넣는다. 그리고 REQ-TUIM-040/041/045·AC-TUIM-026/029 의 v1 쪽 절이 이 SPEC 에서 종료된다는 사실을 요구 층에 명시한다(과거 SPEC 본문은 고치지 않는다).

D3. CENSUS-SRCSCAN — plan.md:L17, L78, L117 — 계획은 `profile_setup.go` 를 읽는 소스 스캔 가드로 `profile_setup_removed_questions_test.go:61,87,116` 3건만 든다. 실제로는 9건이다. 나머지 6건은 `profile_setup_model_policy_test.go:38`(`t.ModelPolicyTitle`·`existingPrefs.ModelPolicy` 존재), `profile_setup_nested_test.go:27`(`persistProjectConfig`·`yaml.Marshal` 부재), `:80`(`permissionMode == defaultPermissionMode`), `:117`(model_policy 인라인 리터럴), `profile_setup_projectconfig_test.go:145`(`&developmentMode`), `schema_bridge_test.go:121`(`&userName`…`&developmentMode` 바인딩 9개)이다. 폼이 v2 `Question{ID,Type}` 정의로 옮겨 가면 양성 가드는 실패하고, 음성 가드(`NewMultiSelect`, `&statuslineTheme` 등)는 표현이 바뀌어 공허해진다. 계획의 뮤턴트(`plan.md:117` "제거된 질문 문자열 재삽입")는 옛 문자열만 잡을 뿐 v2 형태로 되살아난 질문은 못 잡는다. 이 가드들의 비공허성을 요구하는 AC 도 없다. — Severity: major — Class: blocking — Required fix: 9개 스캔을 전수로 적는다. 각 가드를 질문 정의 파일과 저장 경로 파일 가운데 어느 쪽으로 재조준할지 정하고, "v2 표현으로 되살린 제거 질문(예: `ID: \"statusline_theme\"`)이 가드를 실패시킨다"는 뮤턴트 조건을 AC 로 둔다.

D4. CENSUS-V1TYPE — spec.md:L79 — §A.4 는 `schemaSelectOptions` 를 "v1 폼과 무관한 저장·정규화 함수"로 분류해 남긴다. 그러나 `profile_setup.go:124` 는 `[]huh.Option[string]`(v1 타입)을 돌려주고 `huh.NewOption` 을 쓴다. 따라서 이 함수를 그대로 두면 REQ-ITI-004 의 v1 import 0 을 이룰 수 없다. 이 반환형의 `.Key`/`.Value` 에 기대는 테스트도 4개 파일에 있다(`profile_setup_projectconfig_test.go:160`, `profile_setup_schema_options_test.go:49,68,92,108`, `profile_setup_nested_test.go:107`). 취소 판별 `huh.ErrUserAborted`(`profile_setup.go:338,445`)도 조사 목록에 없다. — Severity: major — Class: blocking — Required fix: §A.4 에 "시그니처가 바뀌는 표면" 항목을 두어 `schemaSelectOptions` 의 처분(버전 중립 `{Label,Value}` 또는 v2 타입)과 영향받는 테스트 4개 파일, 그리고 `ErrUserAborted` 에서 위저드 취소 오류로의 이관을 적는다.

D5. REQ005-AC006-MISMATCH — spec.md:L125; acceptance.md:L29 — REQ-ITI-005 는 옵션 값 집합을 "every … from the shared settings schema that the web console also reads" 에서 얻으라고 한다. 그런데 `model_policy` 는 스키마에서 빠졌고(`internal/settings/schema.go:349-351` "model_policy was removed from the console … only the UI field is gone"), AC-ITI-006 은 `template.ValidModelPolicies()` 로 예외를 둔다. REQ 대로라면 AC 가 만족돼도 REQ 는 거짓이다. 또 REQ 는 "project config sync" 목적지 보존을 요구하지만, AC-ITI-006 은 `preferences.yaml` 과 `quality.yaml` 만 확인하고 `profile.SyncToProjectConfig`(`profile_setup.go:509`) 결과는 판정하지 않는다. — Severity: major — Class: blocking — Required fix: REQ-ITI-005 에 `model_policy` 예외를 명시한다. 언어 선택 필드 옵션의 출처(현재 인라인 `langOptions`, `profile_setup.go:321-326`)도 REQ 에 맞춰 정하고, AC-ITI-006 에 프로젝트 동기화 결과 파일·키 비교를 더한다.

D6. REQ006-PRESERVE-GAP — spec.md:L126; acceptance.md:L30-35 — 저장값 보존 목록에서 현재 코드가 지키는 동작 네 가지가 빠졌다. (1) 프로젝트 동기화에 nil 세그먼트 맵을 넘겨 `statusline.yaml` 을 보존한다(`profile_setup.go:505` `syncPrefs.StatuslineSegments = nil`). (2) `persistProjectConfig(cwd, developmentMode, "")` 의 빈 convention 으로 `git-convention.yaml` 을 보존한다(`:516`). (3) `StatuslineTheme` 을 0 값으로 두어 테마를 쓰지 않는다(`:480-486`). (4) 저장된 권한 모드가 비었을 때 `acceptEdits` 를 사전 선택한다(`:298-301`). AC-ITI-007 4번 사례는 `preferences.yaml` 쪽 세그먼트만 본다. — Severity: major — Class: blocking — Required fix: REQ-ITI-006 에 네 동작을 넣고 AC-ITI-007 에 대응 사례를 더한다(`statusline.yaml`·`git-convention.yaml` 바이트 동일, `statusline_theme` 키 부재, 빈 권한 모드의 사전 선택값).

D7. AC008-PREMISE — acceptance.md:L36 — AC-ITI-008 은 "기존 명령 계약 테스트(`profile_worktree_test.go`, `profile_setup_summary_test.go`, 취소 경로 테스트)" 실행으로 이름 기본값 `default` 저장과 취소 시 종료 상태 0 을 판정한다고 쓴다. 그러나 `profile_worktree_test.go` 의 테스트 4개는 헬퍼(`enterSessionWorktree`, `emitProfileScopeNotice`)를 직접 부르고, `profile_setup_summary_test.go` 는 `printProfileSummary` 만 부른다. 취소 경로는 `profile_setup_translations_test.go:23` 에서 `txt.SetupCancelled == ""` 키 존재만 본다. `runProfileSetup` 을 명령 수준에서 몰아 이름 기본값·취소 종료·요약 호출·worktree 정리를 판정하는 테스트는 없다. 인용한 테스트가 모두 초록이어도 뒤 절은 관측되지 않는다. — Severity: major — Class: blocking — Required fix: 위저드 실행 이음새를 주입하는 새 명령 수준 테스트를 AC 에 명시한다. `moai profile setup` 과 `moai profile --setup` 각각에 대해 취소·오류·성공 경로의 종료 상태, "setup cancelled" 출력, 기본/명시 프로필 경로, 요약 출력, worktree 정리 호출을 판정한다.

D8. PTY-VACUOUS — acceptance.md:L22, L47-49, L53; spec.md:L144 — pty 판정이 공허하게 초록이 될 경로가 네 가지 있다. (a) AC-ITI-014 의 "빈 줄 0 / 빈 카드 줄 0" 과 AC-ITI-015 의 "모든 옵션 줄의 열이 같다"는 빈 캡처나 옵션 줄 0개에서도 참이다. 필드 제목·옵션 줄 수 같은 양의 존재 조건이 없다. (b) 캡처 시점(어떤 문자열이 나타날 때까지 기다리는지, 제한 시간 초과 시 FAIL)이 정의되지 않았다. (c) AC-ITI-016 은 정상 종료 뒤 세션 0개만 확인한다. 그래서 REQ-ITI-015 의 "registered cleanup" 과 뒤에 붙인 `kill` 을 가르지 못한다(실패·타임아웃 경로 검사 없음). 세션 이름 충돌 방지와 기존 사용자 세션 보존(`kill-server`·접두사 일괄 삭제 금지)도 요구하지 않는다. (d) HOME 판정은 "하네스가 쓸 수 있는 경로 한정" 파일 목록 비교뿐이라 감시 경로가 정의되지 않았고, 내용 변경은 잡지 못한다. AC-ITI-002 는 자식 프로세스의 작업 디렉터리를 정하지 않았다. tmux 부재 시 FAIL/Gap 처리는 `plan.md:26` 에만 있고 AC 에는 없다. — Severity: major — Class: blocking — Required fix: 캡처마다 고유 기준 문자열과 최소 행 수를 먼저 단정하게 한다. 캡처 대기 조건과 타임아웃→FAIL 을 명시하고, 강제 실패 하위 테스트로 정리 동작을 증명한다. 세션명에 nonce 를 넣고 정확한 세션만 종료하며 외부 감시 세션이 살아남는지 확인한다. 임시 HOME·임시 cwd 를 명시하고 감시 경로를 열거하며, 가능하면 내용 해시를 비교한다. tmux 부재는 PASS 가 아니라고 AC 에 적는다.

D9. REQ-AC-PARTIAL — spec.md:L133, L138-140; acceptance.md:L41-49 — (1) REQ-ITI-012 "Every confirm field" 에는 M2 에서 v2 로 옮기는 다운그레이드 확인창이 포함되는데, AC-ITI-013 은 "위저드 그룹"만 보고 AC-ITI-010 골든은 버튼 정렬을 단정하지 않는다. (2) REQ-ITI-013/014 "A wizard form" 에는 흡수된 프로필 위저드가 포함되지만 AC-ITI-014 는 "첫 페이지 그룹"만 명시한다. (3) REQ-ITI-010 은 네 표면 × 네 로케일인데 AC-ITI-011 은 일반 폼 12개와 `update -c` `ko` 1개뿐이라, 다운그레이드 확인창의 `ja`/`zh` 는 판정되지 않는다. (4) AC-ITI-013/014 가 실제 init 질문을 쓰는지 합성 픽스처를 쓰는지 정하지 않았다. 확인형 질문은 t583 이 고치는 `Page3Questions`(`questions.go:370-505`)에 있어, 흡수 뒤 실제 표면에 확인형 질문이 남는다는 보장이 없다. — Severity: major — Class: blocking — Required fix: AC-ITI-013 에 다운그레이드 확인창을, AC-ITI-014/015 에 프로필 위저드를 넣는다. AC-ITI-011 에 다운그레이드 × 4로케일을 더하고, 픽스처가 실제 생성자(`buildConfirmField` 등)를 거친다는 도달성 조건을 둔다.

D10. SILENT-ASSUME — plan.md:L94; spec.md:L124, L132, L134; acceptance.md:L22, L37, L41 — 미결 질문의 답이 규범 텍스트에 조용히 들어가 있다. Q2: M7 이 `HelpSelect`/`HelpInput` 삭제를 작업으로 확정하고, REQ-ITI-011 의 키 참조 스윕이 사실상 삭제를 강제한다. Q4: REQ-ITI-004 "same huh v2 wizard engine as the init wizard" 와 AC-ITI-009 골든이 엔진의 단계 표시(`wizard.go:168,197` `stepperNote`)를 그대로 굳힌다. Q3: AC-ITI-010 은 `ko` 대 해석 불가 두 경우만 보므로 어떤 우선순위로 구현해도 통과한다. Q1: AC-ITI-002 는 프로필 위저드 고유 표식 없이 "대화 언어 질문 제목"만 보므로, init 위저드의 같은 질문이 떠도, 두 번 물어도 통과한다. — Severity: major — Class: blocking — Required fix: D1 해소 뒤 각 답을 AC 에 판별 가능한 사례로 넣는다. 예를 들어 AC-ITI-010 에 "프로젝트 `ja` + 프로필 `ko`" 사례를, AC-ITI-002 에 프로필 위저드 고유 문자열과 대화 언어 질문 출현 횟수를 둔다.

D11. MISSING-PROGRESS — .moai/specs/SPEC-INIT-TUX-I18N-001/ — `progress.md` 가 없다. `manager-spec.md:96` 은 "`progress.md` is emitted at every Tier", `:110` 은 [HARD] §E 네 제목 뼈대 생성을 요구한다. `plan.md:36` 과 `acceptance.md:104` 도 progress 기록을 전제로 한다. MCP `spec_audit` 도 이 부재 때문에 era 를 V2.x 로 자동 판정했다(E-5). — Severity: minor — Class: blocking — Required fix: §E.1~§E.4 제목만 있는 `progress.md` 를 추가하고 §E.1 만 채운다.

D12. AC005-SCOPE — acceptance.md:L28; spec.md:L124, L149 — REQ-ITI-004 는 모든 비테스트 Go 소스를 대상으로 하는데 AC-ITI-005 grep 은 `internal cmd pkg` 만 본다. 저장소에는 그 밖에 추적되는 비테스트 Go 파일이 12개 있다(`scripts/`, `.moai/scripts/`, `.moai/reports/**`). 거꾸로 `--include=*.go` 는 테스트 파일까지 포함해 REQ 보다 엄격하다. 또 `go mod tidy` 결과를 판정하지 않는다. huh v1 은 `go.mod:14` 에 직접 선언돼 있고, `go mod why -m github.com/charmbracelet/huh` 는 `internal/cli` 경로 하나뿐이며, bubbletea v1.3.10·bubbles v1.0.0·catppuccin(`go.mod:44-47`)은 직접 importer 가 0 이라 함께 빠진다. CI 에는 tidy 검사가 없고 `Makefile:129` 에만 있다. — Severity: minor — Class: blocking — Required fix: AC-ITI-005 의 검색 범위를 REQ 문구와 맞춘다(추적 파일 전체에서 `_test.go` 와 역사 증거만 제외). 그리고 `go mod tidy` 뒤 diff 없음, `go.mod` 에 `github.com/charmbracelet/huh` 줄 없음, `GOOS=windows` 크로스 빌드 통과를 AC 절로 더한다. lipgloss v1 은 직접 소비자가 있어 남는다는 점도 적는다.

D13. AC009-EXCEPTIONS — acceptance.md:L37 — "옵션 값 토큰과 고유명사는 예외 목록으로 명시"라고만 하고 목록이 없다. 언어 옵션 라벨(`English`, `Korean (한국어)`)은 번역하지 않는 것이 현재 설계라(`questions.go:59-61` 주석) 목록 없이는 판정자가 해석해야 한다. — Severity: minor — Class: blocking — Required fix: 예외 문자열 목록을 AC 본문이나 부록에 열거한다.

D14. RED-LEDGER-CMD — acceptance.md:L80 — L1 명령 `grep -rln '"github.com/charmbracelet/huh"' --include=*.go internal cmd pkg` 은 글롭을 인용하지 않아, 이 저장소의 zsh 에서 `no matches found: --include=*.go` 로 실패한다. 이번 감사에서 같은 형태의 명령이 실제로 그렇게 실패했다(E-1 첫 시도). 출력 칸도 원문 줄바꿈이 아니라 ` / ` 로 이어 붙인 요약이다. 이 grep 은 결함이 있을 때 exit 0, 없을 때 exit 1 이라 판정 방향도 명시해야 한다. — Severity: minor — Class: optional — Required fix: `--include='*.go'` 로 인용하고, 원문 출력과 판정 규칙(출력 0줄 = GREEN)을 적는다.

D15. TIER-SIZE — plan.md:L9 — 계획 스스로 테스트를 넣으면 변경 파일이 20개를 넘는다고 적는다. Tier 표(`spec-workflow.md:140-142`)의 L 기준은 15개 초과다. AC 16개는 M 상한 16과 같아서, D3·D6·D7·D9 를 고치면 상한을 넘는다. — Severity: minor — Class: optional — Required fix: 수정 뒤 AC 수를 다시 세어, 상한을 넘으면 Tier L 로 올리거나(`design.md`·`research.md` 추가) 하네스(REQ-ITI-015)를 별도 SPEC 으로 떼어 낸다.

D16. GEARS-LABEL — spec.md:L119, L140 — REQ-ITI-002 의 라벨 "Event-detected" 는 GEARS 패턴 이름이 아니다(구조는 Event-driven). REQ-ITI-014 는 데이터 조건("options carry descriptions")에 `Where` 를 쓰는데, GEARS 의 `Where` 는 기능 게이트·정적 설정 용도다. — Severity: minor — Class: optional — Required fix: 라벨을 "Event-driven" 으로 고치고, REQ-ITI-014 를 `When a select field's options carry descriptions, …` 나 Ubiquitous 조건절로 바꾼다.

D17. ACCEPT-STRUCTURE — acceptance.md:L5, L17 — 절 번호가 §A 에서 §D 로 건너뛴다. RED 원장(§D.3)에는 AC-ITI-005 한 행만 있고 나머지 MUST AC 의 RED 는 run 단계로 미뤘다(`acceptance.md:76`). run 단계 사전 점검(`plan.md:23-27`)이 이를 받으므로 결함이라기보다 기록이다. — Severity: minor — Class: optional — Required fix: 절 번호를 정리한다. 원하면 AC 마다 "RED 를 뜰 마일스톤"을 한 칸 더 둔다.

---

## 호출문 점검 항목별 결론

| 점검 항목 | 결론 | 근거 |
|---|---|---|
| REQ↔AC 매핑 | 형식상 완전, 의미상 부분 | E-3 (lint 커버리지 0 경고 + 뮤턴트 검출), D9 |
| 렌더 성질을 소스 grep 으로 대신하는 AC | 없음 | `acceptance.md:13` [HARD] 가 금지. AC-ITI-001·005 의 grep 은 비렌더 성질이라 허용 |
| pty 하네스의 공허 초록 방지(양성 대조, 크기 고정, env 게이트, 실패 정의) | 크기 80×30·env 게이트는 명시, 양성 대조·실패 정의·정리 실패 경로는 없음 | `acceptance.md:9`, D8 |
| v1 호출 지점 재조사 | `runProfileSetup` 4곳, v1 importer 5개, `moaiHuhTheme` 비테스트 소비 5곳 — SPEC 수치와 일치. 누락은 소스 스캔 6건, `schemaSelectOptions` v1 타입, `ErrUserAborted` | E-1, D3, D4 |
| 저장값 보존(@MX 주석 근처 모델 정규화, acceptEdits, development_mode) | 세 가지는 REQ-ITI-006/AC-ITI-007 에 있음. 프로젝트 동기화 nil 세그먼트·빈 convention·테마 0값·빈 권한 모드 사전 선택은 없음 | D6. 참고: `profile_setup.go:40-41` 은 `@MX:NOTE`/`@MX:REASON` 이며 `@MX:WARN` 이 아니다. `launcher.go:1103-1107` 은 주석뿐이라 SPEC 서술(`spec.md:62`)이 맞다 |
| t583 게이트 | t583 파일을 건드리는 M4~M7 은 모두 게이트 뒤에 있고, 흡수 트리 재측정도 요구됨 | `plan.md:32`, `:71`, `:23`. 게이트 앞 M1·M2 가 `wizard` 패키지에 새 파일을 둘 때 t583 의 미커밋 변경과 식별자가 충돌할 위험은 위험 표에 없다(운영 노트) |
| 소스 스캔 가드의 비공허성 | 보장되지 않음 | D3 |
| `go mod tidy` 영향과 검증 | 영향은 기술되지 않았고 검증 AC 도 없음 | D12 |
| Tier/규모 | REQ 15·AC 16 은 M 상한 안, 파일 수는 L 기준 | D15 |
| 관련 SPEC 모순 | 열거된 7개와는 모순 없음(주장 대조 E-4). 목록 밖 SPEC-CLI-TUI-MODERNIZE-001 과 모순 | D2 |
| 결정 D1~D6 준수 | D1(`spec.md:118`), D2(§A.3-A.4, 조사 불완전), D3(`spec.md:158-161`, `:148`), D4(`spec.md:163-165`), D5(`spec.md:140`), D6(`spec.md:144`, `acceptance.md:5-11`) 모두 반영 | 결정 자체는 다시 열지 않았다 |

---

## Evidence

### E-1 조사 재측정 (트리 `d0ec7921f`)

```
$ grep -rn 'runProfileSetup' --include='*.go' internal cmd pkg   (비테스트 호출·등록만)
internal/cli/update.go:187:				if err := runProfileSetup(cmd, nil); err != nil {
internal/cli/profile.go:62:		return runProfileSetup(cmd, args)
internal/cli/profile_setup.go:214:	RunE: runProfileSetup,
internal/cli/init.go:659:			if err := runProfileSetup(cmd, nil); err != nil {
(launcher.go:1104,1106 · profile_setup.go:34,221,231 은 주석)

$ grep -rln '"github.com/charmbracelet/huh"' --include='*.go' . --exclude-dir=.claude --exclude-dir=node_modules
internal/cli/huh_theme.go
internal/cli/update_version.go
internal/cli/update.go
internal/cli/profile_setup.go
internal/cli/init.go
(대조군: grep -rln '"charm.land/huh/v2"' … | wc -l → 2, 검색식 작동)

$ grep -rn 'moaiHuhTheme\|moaiHuhStyles' --include='*.go' internal cmd pkg
비테스트 소비: update.go:185, update_version.go:331, profile_setup.go:335, :442, init.go:657
테스트 소비: huh_theme_test.go:13, :45, :73

$ grep -rn 'ReadFile("profile_setup.go")' --include='*_test.go' internal/cli
profile_setup_projectconfig_test.go:145
profile_setup_model_policy_test.go:38
profile_setup_removed_questions_test.go:61, :87, :116
schema_bridge_test.go:121
profile_setup_nested_test.go:27, :80, :117

$ grep -rn 'schemaSelectOptions' --include='*.go' internal/cli   (정의·호출)
profile_setup.go:124: func schemaSelectOptions(t profileSetupText, field string, withEmpty bool) []huh.Option[string] {
테스트 호출: profile_setup_projectconfig_test.go:160, profile_setup_schema_options_test.go:49,68,92,108, profile_setup_nested_test.go:107
```

첫 시도에서 `--include=*.go` 를 인용하지 않았더니 zsh 가 `(eval):2: no matches found: --include=*.go` 로 명령 전체를 거부했다. 위 결과는 인용한 뒤 다시 잰 값이다(D14 근거).

### E-2 모듈 그래프

```
$ grep -n 'huh\|bubbletea\|bubbles\|catppuccin' go.mod
9:	charm.land/huh/v2 v2.0.3
14:	github.com/charmbracelet/huh v1.0.0
44:	github.com/catppuccin/go v0.3.0 // indirect
46:	github.com/charmbracelet/bubbles v1.0.0 // indirect
47:	github.com/charmbracelet/bubbletea v1.3.10 // indirect
$ go mod why -m github.com/charmbracelet/huh
# github.com/charmbracelet/huh
github.com/modu-ai/moai-adk/internal/cli
github.com/charmbracelet/huh
$ grep -rn '"github.com/charmbracelet/bubbletea"\|"github.com/charmbracelet/bubbles' --include='*.go' internal cmd pkg
(출력 없음, exit 0)
$ grep -rn 'mod tidy' .github/workflows/*.y*ml Makefile
Makefile:129:	go mod tidy
```

### E-3 SPEC lint 와 커버리지 수집기의 비공허성

설치된 바이너리(`moai-adk v3.2.0-rc.6`, commit `1bea05daa`)는 이 트리의 조상이 아니다: `git merge-base --is-ancestor 1bea05daa HEAD` → exit 1, `git rev-list --count HEAD..1bea05daa` → 17, merge-base `647ad0157`. 이 빌드가 판정한 결과를 트리 판정으로 쓸 수 없어서, 트리에서 경로를 지정해 빌드한 바이너리로 다시 쟀다.

```
$ go build -o <scratchpad>/moai-tree ./cmd/moai        → BUILD_EXIT=0
$ <scratchpad>/moai-tree version                        → v3.1.3 none built unknown   (ldflags 없이 빌드, 커밋은 빌드 시점 트리 d0ec7921f)
$ <scratchpad>/moai-tree spec lint SPEC-INIT-TUX-I18N-001   → LINT_TREE_EXIT=0
INFO  OwnershipTransitionUnmeasured  …spec.md 1  … commit d0ec7921f… has no Authored-By-Agent trailer — ownership transition unmeasured
0 error(s), 0 warning(s)
```

설치 바이너리의 결과도 같았다(`LINT_EXIT=0`, 같은 INFO 1건).

커버리지 경고 0 이 공허하지 않은지 확인하려고, 산출물 사본의 spec.md 에 AC 없는 `REQ-ITI-016` 을 덧붙인 뮤턴트를 만들어 lint 했다(원본은 건드리지 않았다).

```
$ <scratchpad>/moai-tree spec lint <scratchpad>/mutant/.moai/specs/SPEC-INIT-TUX-I18N-001/spec.md
WARNING  CoverageIncomplete  …spec.md  176  REQ REQ-ITI-016 is not referenced by any AC
0 error(s), 1 warning(s)
```

→ 수집기는 이 acceptance.md 의 `(maps REQ-…)` 형식을 읽는다. 원본의 경고 0 은 실제 커버리지 판정이다.

### E-4 관련 SPEC 상태와 주장 대조

```
SPEC-CLI-TUX-INIT-UPDATE-001 status: completed
SPEC-CLI-TUX-V3-002 status: completed
SPEC-CLI-WIZARD-RESTRUCTURE-001 status: completed
SPEC-I18N-GOVERNANCE-001 status: completed
SPEC-INIT-WIZARD-REPAIR-001 status: completed
SPEC-WEB-CONSOLE-002 status: completed
SPEC-WEB-CONSOLE-003 status: completed
```

§A.6 의 인용은 모두 실재했다. REQ-TUX2-006/008/012(`SPEC-CLI-TUX-V3-002/spec.md:55,57,64`), `### Out of Scope — wizard rendering engine / theme`(`SPEC-CLI-WIZARD-RESTRUCTURE-001/spec.md:100`), REQ-WIZ-007(`:48`), REQ-TUXIU-044(`SPEC-CLI-TUX-INIT-UPDATE-001/spec.md:93`), "no changes to wizard flow, pages, or translations"(`SPEC-INIT-WIZARD-REPAIR-001/spec.md:103`), REQ-WC2-006(`SPEC-WEB-CONSOLE-002/spec.md:72`), REQ-WC3-006(`SPEC-WEB-CONSOLE-003/spec.md:93`)이다.

목록 밖 충돌은 다음과 같다.

```
$ grep -n 'REQ-TUIM-040\|REQ-TUIM-045' .moai/specs/SPEC-CLI-TUI-MODERNIZE-001/spec.md
117:- **REQ-TUIM-040** (Ubiquitous): The huh v1 theme factory and the huh v2 theme factory **shall** remain two separate factories; …
122:- **REQ-TUIM-045** (Ubiquitous): Both factories **shall** continue to resolve the light/dark axis through their existing package-level indirection variables, …
$ grep -n 'AC-TUIM-026\|AC-TUIM-029' .moai/specs/SPEC-CLI-TUI-MODERNIZE-001/acceptance.md
116:| **AC-TUIM-026** | … | `grep -n "func moaiHuhStyles" internal/cli/huh_theme.go` → 1 hit **AND** … |
119:| **AC-TUIM-029** | … | `grep -n "var huhThemeIsDark" internal/cli/huh_theme.go` → 1 hit **AND** … |
status: completed
```

SPEC-INIT-HARNESS-PROMPT-001(completed)도 `init.go` 의 프로필 확인창을 언급하지만, REQ-IHP-005 가 "this requirement makes no claim about prompt totals"(`spec.md:104`)라고 스스로 밝혀 D1 과 충돌하지 않는다.

### E-5 MCP 교차 확인

- `mcp__moai__spec_audit(filter_spec=SPEC-INIT-TUX-I18N-001, project_root=<워크트리>)` → `{"era":"V2.x","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-1 (progress.md absent)"}}`. MCP 서버 빌드 `84fa4ece4` 는 HEAD 의 조상이다(`git merge-base --is-ancestor 84fa4ece4 HEAD` → exit 0). 즉 트리보다 뒤처진 빌드이며, 이 결과는 D11 의 보조 신호로만 썼다.
- `mcp__moai__audit_multi(project_root=<워크트리>, target=baseBranch, gates={claude: required, codex: advisory, glm: advisory})` → `overall_verdict: fail`. codex 는 `inconclusive`(fail-open)로 분류됐지만 본문으로 FAIL 소견 12건을 냈다. glm 은 `glm unavailable: z.ai response carried no text content`. codex 소견은 이 감사에서 직접 다시 재서 확인한 것만 채택했다. 채택한 것은 D2(교차 SPEC 충돌), D5 후반(프로젝트 동기화 미판정), D7(명령 수준 테스트 부재), D8 (c)(d)(세션명·HOME 내용), D11(progress.md), D12(검색 범위·tidy), D14(zsh 글롭 실패)다. "MUST AC 전부에 RED-now 증거 필요"라는 소견은 run 단계 사전 점검(`plan.md:23-27`)과 `acceptance.md:76` 의 이월 명시가 받고 있어 D17 optional 로만 반영했다.

---

## Baseline-attribution

- 모든 측정은 이번 실행에서 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t586`, HEAD `d0ec7921fd743608773bde75e5c7aa3cb0860644` 에서 했다. 착수 시 `git rev-parse HEAD` 출력이 호출문의 감사 커밋과 같았다.
- `git diff --stat e7a7d4bb3 HEAD` 는 SPEC 파일 3개(+412)뿐이다. 따라서 SPEC 이 인용한 조사 트리 `e7a7d4bb3` 와 감사 트리의 제품 코드는 같다.
- lint 판정 빌드: 트리에서 빌드한 `<scratchpad>/moai-tree`(빌드 트리 `d0ec7921f`). 설치 바이너리 `1bea05daa` 는 비조상이라 참고로만 썼다. MCP 서버 빌드는 `84fa4ece4`(조상, 뒤처짐).
- huh v2 공개 API 인용(`plan.md:18`)은 모듈 캐시 `/Users/goos/go/pkg/mod/charm.land/huh/v2@v2.0.3` 에서 직접 확인했다. `field_confirm.go:361` WithButtonAlignment, `:53` `buttonAlignment: lipgloss.Center`, `:261-264` `"\n"` 두 번, `field_select.go:256` Height, `form.go:284` WithKeyMap, `keymap.go:107` NewDefaultKeyMap, `theme.go:27`/`:104` FieldSeparator `"\n\n"`. 모두 일치한다.

## Gaps

- t583 의 미커밋 변경은 이 트리에 없어서 보지 못했다. t583 의 `WT-init-quiet-wizard` 브랜치 tip `120436f58` 은 develop 흡수 병합 하나뿐이다. 따라서 게이트 뒤의 질문 구조·확인형 질문 존속(D9 (4))은 판정하지 못했다.
- 제안된 pty 하네스·골든·새 테스트는 아직 없어서, 선택 테스트 수와 뮤턴트 실패를 실행으로 확인하지 못했다.
- `moai spec view SPEC-INIT-TUX-I18N-001` 은 트리 빌드에서 `Parse error: acceptance criteria section not found.`(exit 1)로 실패했다. 같은 `§D AC Matrix` 형식을 쓰는 완료 SPEC `SPEC-INIT-WIZARD-REPAIR-001` 도 똑같이 실패해서, 이 SPEC 고유 결함으로 세지 않았다. 원인(파서가 기대하는 절 제목)은 판정하지 않았다.
- 언어 선택 필드(`git_commit_lang` 등)의 스키마 옵션 정의 여부는 `internal/settings/schema.go:331-334` 의 필드 선언까지만 봤고, 옵션 목록 내용은 읽지 않았다(D5 수정 방향에만 영향).
- `schemaSelectOptions` 이외에 v1 타입을 매개변수·반환형으로 노출하는 헬퍼가 더 있는지는 `huh.` 접두 전수 스윕으로 확인하지 않았다(`internal/cli` 위저드 밖 비테스트 `huh.New…` 호출 33건이라는 개수만 셌다).

## Residual-risk

- codex 백엔드는 `inconclusive` 로 분류돼 수렴 판정에 참여하지 않았고 glm 은 응답하지 않았다. 교차 모델 의견은 사실상 claude 한 명의 판정이다.
- 뮤턴트 lint·버전 확인에 쓴 산출물(`<scratchpad>/lint*.txt`, `view-*.txt`, `mutant/`)은 세션 전용 임시 디렉터리에 있고 반출하지 않았다. 판정 근거는 이 보고서에 원문으로 옮긴 출력뿐이며, 그 파일 경로는 근거로 인용하지 않는다.
- 이 보고서는 워크트리에 커밋되지 않은 파일이다. 증거 커밋은 레인·리드 몫이다.

---

## Recommendation

FAIL. manager-spec 에 넘길 수정 순서는 다음과 같다.

1. **D1**: 리드가 Q1~Q4 답을 받는다. 그 답을 **D10** 방식으로 REQ·AC 에 판별 가능한 사례로 넣고 `plan.md:103-106` 마커를 지운다.
2. **D2**: `related_specs`·§A.6 에 SPEC-CLI-TUI-MODERNIZE-001 을 넣고, REQ-TUIM-040/041/045·AC-TUIM-026/029 의 v1 절 종료를 요구 층에 적는다.
3. **D3·D4**: §A.3-A.4 조사를 넓힌다. 소스 스캔 9건과 재조준 대상, `schemaSelectOptions` 시그니처 처분과 영향 테스트 4개 파일, `ErrUserAborted` 이관을 적고, 가드 비공허성 AC(v2 표현 뮤턴트)를 더한다.
4. **D5·D6**: REQ-ITI-005 에 `model_policy` 예외를 명시하고 프로젝트 동기화 판정을 넣는다. REQ-ITI-006·AC-ITI-007 에 보존 동작 네 가지를 더한다.
5. **D7**: AC-ITI-008 을 명령 수준 이음새 테스트로 다시 쓴다.
6. **D8·D9**: pty AC 에 양의 존재 조건, 대기·타임아웃, 강제 실패 정리 검사, 세션 소유, HOME/cwd 감시 범위를 넣는다. 다운그레이드 확인창·프로필 위저드·다운그레이드 `ja`/`zh` 커버리지와 실제 생성자 도달성을 더한다.
7. **D11·D12·D13**: `progress.md` 뼈대를 추가하고, AC-ITI-005 범위와 `go mod tidy` 판정을 더하고, AC-ITI-009 예외 목록을 열거한다.
8. 수정 뒤 AC 수가 16을 넘으면 **D15** 에 따라 Tier L 로 올리거나 하네스 SPEC 을 떼어 낸 다음 2회차 감사를 받는다. 2회차는 이 결함 목록 기준 델타 재감사다.

---

## Operational Notes (unverified)

- `inferred` — 게이트 앞 M1·M2 가 `internal/cli/wizard` 패키지에 새 파일(질문 세트·번역·확인창 헬퍼)을 두면, t583 이 같은 패키지 `translations.go`·`questions.go` 에 추가하는 식별자와 이름이 겹칠 수 있다. 근거 규칙: 같은 Go 패키지의 최상위 식별자는 파일이 달라도 한 이름공간을 쓴다. 측정 방법: 흡수 직후 `go build ./internal/cli/wizard/` 로 확인한다.
- `measured` — `moai spec view` 는 `§D AC Matrix` 형식 acceptance.md 에서 파싱 오류를 낸다(명령과 출력은 Gaps 에 기록). 반면 lint 커버리지 수집기는 같은 형식을 읽는다(E-3 뮤턴트). view 실패를 lint 커버리지 공허의 증거로 읽으면 안 된다.
- `assumption` — D4 수정에서 `schemaSelectOptions` 반환형을 버전 중립 구조체로 바꾸면 `internal/cli` → `internal/cli/wizard` 방향 제약(`spec.md:150`)을 지키기 쉬울 것이다. 설계 판단은 manager-spec 몫이다.
