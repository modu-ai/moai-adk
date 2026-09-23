## Evaluation Report
SPEC: SPEC-MODEL-OPUS55-001 (card t1089, Tier M)
Overall Verdict: **FAIL** — 차단 결함 1건(F1, K4 수리의 운영자 `--effort` 우선순위 위반)과 문서 차단 결함 1건(F2). 나머지 16개 AC는 재측정으로 통과했다.

- 감사 트리: `.claude/worktrees/t1089`, 브랜치 `WT-opus-55-default`, HEAD `0842a0ac4`
- 카드 범위: `CARD_BASE=$(git merge-base develop HEAD)` → `1dbe5e2f3`, `git diff --stat 1dbe5e2f3..HEAD` → `85 files changed, 2613 insertions(+), 260 deletions(-)` (대조군 비어 있지 않음)
- Go 트리 동일성: `git diff --name-only eb629efb5..HEAD` → `.moai/specs/SPEC-MODEL-OPUS55-001/progress.md`, `spec.md` 두 줄뿐. K4 수리 커밋 이후 Go 코드는 바뀌지 않았다.
- 평가 프로필: `.moai/config/evaluator-profiles/default.md` (flat 가중 모드, must-pass = Functionality + Security)
- 근거 규칙: `verification-claim-integrity.md` §1.1 surface 2·3, `verification-completeness.md` §1.1(빈 스윕), `agent-common-protocol.md` § Skeptical Evaluation Stance

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 75/100 | PASS (must-pass 충족) | SPEC AC 16건 전부 재측정 통과(AC-012는 전체 cli 패키지만 귀속 증거, 아래 표). K4 수리의 `--` 경로 결함(F1)은 SPEC AC 밖이라 must-pass는 깨지지 않지만 점수를 0.75 앵커로 낮춘다: "All primary acceptance criteria pass; minor edge cases missing" |
| Security (25%) | 100/100 | PASS | 추가된 Go 라인 grep(`exec.Command\|password\|secret\|api_key\|sh -c\|unsafe`) → 산문 속 "token" 단어만 적중, 코드 적중 0. argv 분기는 `effort == template.EffortLevelMax` 정확 일치일 때만 `[--effort max]`를 만들고 셸을 거치지 않으므로(syscall.Exec) 프로필 값으로 인자 주입이 불가능하다 |
| Craft (20%) | 75/100 | FAIL (비 must-pass, 기존 기준선) | `go test -coverprofile ./internal/template/` → `coverage: 81.7% of statements` (85% 미달, progress E.2.3의 착수 전 값과 동일 — 이 카드가 낮춘 것 아님). 카드 신규 로직: `applyLaunchEffort 100.0%`, `launchEffortArgs 100.0%`, `appendCrossSessionSettings 100.0%`, `prepareKanbanSettings 93.3%`, `operatorSuppliedEffort 83.3%`(`--` 분기 미커버 = F1). `golangci-lint run ./internal/cli/ ./internal/cli/wizard/ ./internal/template/ ./internal/web/` → `0 issues.` exit 0. `go vet` exit 0. `gofmt -l` 17개 변경 Go 파일 → 무출력 |
| Consistency (15%) | 75/100 | PASS | 템플릿/로컬 쌍: 15개 공통 파일의 정규화 hunk 비교에서 차이는 이 카드와 무관한 기존 GLM-5.2/5.3 문구뿐. SAME 쌍 6개 `cmp` 전부 rc=0. 감점: SPEC 본문이 K4 런타임 변경을 부정(F2), `settings-management.md:93`이 max 경로를 잘못 서술(F3), progress의 N3 처분 서술 부정확(F5) |

가중 조화평균: 1 / (0.40/75 + 0.25/100 + 0.20/75 + 0.15/75) = **80.0** (비가중 조화평균도 80.0).

### Per-AC 재측정 표 (이 감사에서 직접 실행, HEAD `0842a0ac4`)

| AC | 명령 | 결정적 출력(verbatim) | 판정 |
|----|------|----------------------|------|
| 001 | `grep -nE 'ModelIDOpus55 = "claude-opus-5-5"\|"opus":[[:space:]]+ModelIDOpus55' internal/template/model_policy.go` | `51:const ModelIDOpus55 = "claude-opus-5-5"` / `78:	"opus":     ModelIDOpus55,` rc=0 | PASS |
| 002 | `grep -nE '"claude-opus-5":[[:space:]]+"opus",[[:space:]]*// superseded' internal/template/model_policy.go` | `98:	"claude-opus-5":     "opus", // superseded by ModelIDOpus55` rc=0 | PASS |
| 003 | `grep -rnw 'ModelIDOpus5' internal --include='*.go'` | 무출력, rc=1 | PASS |
| 004 | P4 v2 (`find … -exec awk …`, acceptance.md 원문 그대로) | 무출력. 양성 대조 `awk … internal/template/model_policy.go` → `1` | PASS |
| 005 | anchor/heading grep 3종 + `go test -run 'TestRegistrySyncGuard\|TestRegistrySyncMirrorsIdentical' -v ./internal/constitution/` | `:2`/`:2`, `:0`/`:0`, `:1`/`:1`; `--- PASS: TestRegistrySyncGuard (0.15s)`, `--- PASS: TestRegistrySyncMirrorsIdentical (0.00s)` | PASS |
| 006a | P7 awk (두 사본) | `006a=2` | PASS |
| 006b | `grep -cE 'Opus 5\.5[^\|]*medium\|medium[^\|]*Opus 5\.5'` 9개 파일 | constitution ×2 `:1`, model-policy ×2 `:4`, agent-authoring ×2 `:1`, dynamic-workflows ×2 `:1`, tech.md `:1` | PASS |
| 006c | P6 | 무출력, `rc006c=1` | PASS |
| 006d | P5 | 무출력, `rc006d=1` | PASS |
| **006e (N1)** | 음성: `grep -cE '^- \*\*Effort calibration\*\*:.*only for speed-critical or simple tasks'` 두 사본 / 양성: `…(Opus 5\.5[^.]*medium\|medium[^.]*Opus 5\.5)` | 음성 `:0`/`:0` `rcneg=1`; 양성 `:1`/`:1` `rcpos=0`. 35행 원문: "`medium` is Opus 5.5's default and MoAI's recommended session effort. Raise it per role (`high` / `xhigh` / `max`) … use `low` for speed-critical or simple tasks." | PASS — plan-audit iter-2 N1 해소 확인 |
| 007 | i18n grep 10종 + `go test -run 'TestModelOptLabelsEnglishUnified\|TestEffortOptRecommendationLabels' -v ./internal/web/` | 모델 라벨 `4`, 옛 라벨 `0`, medium 4개 로케일 각 `1`, runtime_default 4개 로케일 각 `1`; `--- PASS: TestModelOptLabelsEnglishUnified`, `--- PASS: TestEffortOptRecommendationLabels`; `TestResolveLaunchEffort` PASS | PASS |
| 008 | `profile_setup_translations.go` 로케일별 `grep -cE` 8종 | 8개 모두 `1`. 권장 표식은 opus[1m]·medium에만 존재(다른 effort 옵션 적중 0) | PASS |
| 009a/b/c | 가드 테스트 현재 상태 + `run-ac009-mutants.log` 판독 | `--- PASS: TestGetProfileText_OpusAliasValues`, `--- PASS: TestModelPolicyLabels_AgreeWithProfileMatrix`, wizard 패키지 `ok`. 로그: M-a `should reference Opus 5.5`, M-b `does not name "Opus 5.5"`, M-c `should reference Opus 6` / `does not name "Opus 6"` | PASS (변이는 로그 의존, F9) |
| 010 | `cmp -s` 6쌍 | 6개 모두 `rc=0` | PASS |
| 011 | `grep -n -i effort internal/template/templates/.claude/settings.json.tmpl` | 무출력, `rc011=1` | PASS |
| 012 | `go test -count=1 ./internal/template/ ./internal/cli/wizard/ ./internal/web/ ./internal/settings/ ./internal/constitution/` + cli 전체(슬롯 `go-test-cli`, 환경 정리 단일 복합 호출) | 5개 패키지: `ok … template 74.526s`, `ok … cli/wizard 4.335s`, `ok … web 25.482s`, `ok … settings 0.650s`, `ok … constitution 1.242s`. cli 대상 실행 27 `--- PASS`, 0 `--- FAIL`, `ok … internal/cli 1.078s`. **cli 전체**: 내 실행은 `panic: test timed out after 25m0s`(부하 18~27) + `--- FAIL: TestCC_FactoryEntryThroughRunCC/-f_lane-2 … AMBIGUOUS_FACTORY`; 해당 테스트 단독 실행 `--- PASS` (F6). 이전 증거 `run-k4-cli-full-scrubbed.log` = `ok … internal/cli 1327.513s` (Go 트리 동일한 `eb629efb5`) | PASS-귀속 / 전체 cli는 이 감사에서 완주 미관측 |
| 013 | `git diff develop...HEAD -- internal/template/profile_matrix.go` | 주석 두 줄만 변경(`(measured on Opus 5)`), `defaultProfileMatrix` 셀 무변경 | PASS |
| 014 | `git diff --name-only develop...HEAD -- CHANGELOG.md docs-site README*.md .moai/research .moai/docs .moai/reports ':!.moai/reports/t1089'` | 무출력. 동반: `.moai/specs`는 `SPEC-MODEL-OPUS55-001/*` 4개뿐 | PASS |
| 015 | `go test -run 'TestTemplateNoInternalContentLeak\|TestLanguageNeutrality\|TestLeakClassNoDateShaInDefaultTier' -v ./internal/template/` | 세 줄 `--- PASS`, `ok` | PASS |
| 016 | 카탈로그 해시 드라이런 비교 + agents/commands 소스 diff | `computed=47 stale=0`(양성 대조: `catalog.yaml`에 `hash:` 47줄); agents·commands·skills/moai·.codex·.agents diff `0` | PASS (`make build` 자체는 미실행 — 트리를 쓰기 때문) |

### K4 수리 검증 (점검 항목 2)

- 공식 근거 재확인: `curl code.claude.com/docs/en/model-config` 본문에서 "set `effortLevel` to `low`, `medium`, `high`, or `xhigh` as the default for models without one. `max` isn't accepted as a level in either key" 및 "`--effort` flag: pass a level name to set it for a single session" 확인. 설치된 CLI `claude --version` → `2.1.280 (Claude Code)`, `claude --help` → `--effort <level>  Effort level for the current session (low, medium, high, xhigh, max)`.
- 통과 확인: `TestLaunchEffortMaxTravelsAsArgvOnGeneralInjection`, `…OnKanbanInjection`, `TestLaunchEffortXHighStaysOnSettingsPath`, `TestLaunchEffortMaxDefersToOperatorEffortFlag`, `TestApplyLaunchEffort`(max 하위 테스트 포함), `TestClaudeLaunchEnvPreservesInheritedEffort` 전부 `--- PASS`. RED 로그 `run-k4-red.log`에 `settings effortLevel = max; max must never be written…` 확인.
- 확인된 불변식: max는 페이로드에 들어가지 않고 `[--effort max]`로 반환, low~xhigh는 settings 경로 유지, `buildEnvForClaudeLaunch`는 무변경, `CLAUDE_CODE_EFFORT_LEVEL` 주입 없음.
- **깨진 불변식: 운영자 `--effort` 우선순위** — F1 참조.

### Findings (structured defect-list)

- **F1** [Medium, 신뢰도 높음] [blocking] `internal/cli/launch_effort_settings.go:87-97` (`operatorSuppliedEffort`), 호출부 `internal/cli/crosssession_settings.go:90-91`, `internal/cli/kanban_settings.go:82-83` — 운영자 `--effort` 탐지가 MoAI 구분자 `--`에서 멈추는데, 런처는 그 `--`를 소비하고 뒤쪽을 Claude에 그대로 넘기며(`internal/cli/launcher.go:652-656`) moai가 주입하는 플래그는 argv 맨 뒤(즉 `--` 뒤)에 붙는다. `go test -overlay`로 트리를 건드리지 않고 실행한 프로브 결과:
  - 일반 경로 `appendCrossSessionSettings(root,"dev",["--","--effort","low"])`, 프로필 max → `general argv=[-- --effort low --effort max] effortFlags=2 last=max`
  - kanban + 일반 퍼널 → `kanban+funnel argv=[-- --effort low --settings …moai-kanban-….json --effort max --effort max] effortFlags=3 last=max settingsFlags=1`
  - 대조(구분자 없음) → `control argv=[--effort low] effortFlags=1 last=low`
  progress E.2.6의 "An operator-supplied `--effort` wins, so no second flag is added"는 `--` 형태에서 거짓이다. K4 이전에는 max가 무효한 settings 값이었으므로 운영자 `--effort low`(명시적 선택)가 이겼는데, 이제는 두 값이 경합한다. Claude Code가 두 값 중 무엇을 적용하는지는 관측하지 못했다(`claude --effort bogusA --effort bogusB --version`이 두 값 모두 경고 → 둘 다 검증됨). codex 백엔드도 같은 결함을 HIGH로 독립 보고했다.
  - Required fix: 운영자 `--effort`를 argv 전체(MoAI `--` 양쪽)에서 탐지하거나 Claude로 가는 꼬리만 기준으로 판정하고, 이미 argv 어디든 `--effort`가 있으면 주입하지 않도록 한다(kanban이 `--` 뒤에 넣은 플래그를 퍼널이 다시 붙이는 중복도 함께 해소). 테스트: 일반·kanban 두 경로에 `-- --effort low`, `-- --effort=low` 사례를 추가하고 최종 argv의 `--effort`가 정확히 1개(운영자 값)임을 단언. RED 먼저.
- **F2** [Medium] [blocking, 문서 전용] `.moai/specs/SPEC-MODEL-OPUS55-001/spec.md:85-86`, `plan.md:194` — spec §C는 "The one runtime change this SPEC makes is the `opus` alias target … Nothing else changes at runtime"라고 쓰고 plan §E K4는 "Not changed here"라고 쓰는데, `eb629efb5`가 max 전달 경로(런타임 argv)를 바꿨다. K4를 다루는 REQ/AC도 없다. SPEC은 `completed`인데 본문이 출하된 동작을 부정한다.
  - Required fix: manager-spec이 spec §C에 두 번째 런타임 변경(K4: max → `--effort max`, 운영자 `--effort` 우선)을 기록하고 K4 테스트 4종을 가리키는 AC를 추가하거나, 최소한 HISTORY와 §C 문장을 사실대로 고친다(재배차 경로: 오케스트레이터 → manager-spec).
- **F3** [Low, 신뢰도 높음] [optional] `.claude/rules/moai/core/settings-management.md:93` 및 템플릿 사본 — "The launcher passes the profile's effort as an `effortLevel` in the transient `--settings` file"는 max에 대해 이제 거짓이다. Required fix: "max는 settings 키가 받지 않으므로 `--effort max` 실행 플래그로 전달한다" 한 절 추가(템플릿 먼저, `make build`).
- **F4** [Low] [optional, 기존] `.claude/rules/moai/development/coding-standards.md:103`, `.claude/rules/moai/workflow/worktree-integration.md:456`(두 사본) — `effortLevel` 값에 `max`를 포함해 서술한다. 공식 문서("`max` isn't accepted as a level in either key") 및 이 카드의 K4 근거와 모순. 카드가 만든 결함은 아니지만 K4 잔여 목록에 없다. Required fix: 후속 카드로 두 사본 정정.
- **F5** [Low] [optional] `.moai/specs/SPEC-MODEL-OPUS55-001/progress.md` §E.4.2 N3 처분 — "The Known-limit note already exists at acceptance.md:61"이라 쓰지만 acceptance.md:61은 "Opus 5-era" 미탐지 한계이고, N3가 요구한 한계(`(superseded)`가 줄 어디에 있든 그 줄 전체가 면제됨)는 기록되지 않았다. 실측상 현재 `(superseded)` 사용처는 `claude-opus-5`가 실제로 대체된 두 줄(model-policy.md:19 두 사본, tech.md:14)뿐이라 오남용은 없다. Required fix: progress 서술 정정 또는 manager-spec이 acceptance에 한계 문장 추가.
- **F6** [Low] [optional] `progress.md` §E.2.6 / §E.4.1 — `TestCC_FactoryEntryThroughRunCC/-f_lane-2`(`AMBIGUOUS_FACTORY`) 실패를 "lane env leakage"로만 귀속하지만, 이 감사의 환경 정리 전체 실행에서도 같은 실패가 재현됐고 단독 실행은 통과했다 → 실행 순서/호스트 상태 의존 flake. 발생원 `internal/factorymsg/store.go:276`은 카드 diff 밖(`git diff --name-only develop...HEAD -- internal/factorymsg internal/kanban` → `0`)이라 이 카드 귀속은 아니다. Required fix: 귀속 문구를 "순서 의존 flake, 카드 무관"으로 정정하고 테스트 격리 후속 카드 검토.
- **F7** [Low] [optional] `progress.md` §E.4.3 잔여 목록 불완전 — (a) GLM만 언급하지만 일반 퍼널(`launcher.go:212-219`, "cc / glm / gpt all funnel through here")을 타는 모든 공급자가 max 프로필에서 `--effort max`를 받는다; F1이 없음; `internal/template` 커버리지 81.7%(<85%, 기존)가 E.2.3에만 있고 잔여에 없음. Required fix: 잔여 목록 보강.
- **F8** [Low, 신뢰도 중간] [optional] `.claude/rules/moai/development/model-policy.md:19`(두 사본), `.moai/project/tech.md:15` — "Other effort-capable models default to `high`". 공식 문서: "high on every model that supports effort, except that Opus 5.5 defaults to medium, Opus 4.7 defaults to xhigh". tech.md:14가 Opus 4.7을 여전히 지원 모델로 적으므로 부정확하다. Required fix: "(Opus 4.7은 `xhigh`)" 단서 추가.
- **F9** [Info] [optional] `.moai/reports/t1089/run-ac009-mutants.log` — M-c(별칭 → `claude-opus-6`) 변이는 세 번째 가드 `TestModelPolicyLabels_AgreeWithProfileMatrix` 도입 전 트리(`43aa64cba`)에서 실행돼 그 가드의 별칭 추종 변이 증거가 없다. M4에서 `Opus 5-5`로 자연 RED가 난 기록이 보조 근거다. Required fix: 없음(필요 시 M-c 재실행에 해당 테스트 포함).
- **F10** [Info] [optional, 기존] 로컬 `.claude/skills/moai-foundation-core/modules/token-optimization.md`, `.claude/skills/moai-workflow-spec/references/reference.md`는 GLM-5.2, 템플릿 사본은 GLM-5.3. 이 카드의 hunk는 양쪽이 동일하므로 카드 무관한 기존 드리프트다.

### 교차 모델 감사 (`mcp__moai__audit_multi`, target `baseBranch`, project_root = 이 워크트리)

- 기준 해석: `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop`, `git merge-base origin/develop HEAD` → `1dbe5e2f3` (카드 범위와 동일).
- codex: **fail** — F1과 같은 결함(HIGH), 그 외 K4 정상 경로·별칭 매핑은 결함 없음으로 확인.
- claude: **inconclusive** — 리포트/로그가 많은 diff가 크기 제한으로 잘려 Go hunk가 전달되지 않음. 결함 판정이 아니라 입력 절단.
- glm: **inconclusive** — "z.ai response carried no text content".
- 종합 `overall_verdict: fail`, `gate_unmet: claude`. 서버 빌드 `cd99336bf`는 HEAD의 조상(바이너리 지연 경고) — MCP 판정은 참고용이며 이 보고서의 판정은 위 실측에 근거한다.

### Recommendations

- F1을 먼저 수리하고(RED 먼저), 재감사는 F1·F2 델타로 한정한다: `operatorSuppliedEffort`/`launchEffortArgs` 변경 + 신규 테스트 + spec §C/HISTORY 정정.
- F3·F5·F7은 같은 sync 정정 커밋에 묶어도 되는 문서 한 줄 수정이다. F4·F6·F8은 후속 카드 후보.
- 전체 `internal/cli` 완주 판정은 병합 뒤 develop CI에 맡긴다(로컬 부하 18~27에서 25분 제한 초과).

### Gaps (이 감사에서 관측하지 못한 것)

- `internal/cli` 전체 패키지 완주 미관측: 내 실행은 25분 제한에서 중단(`running tests: TestTodoDoneUndone_NeverPrompt`). 완주 증거는 이전 실행의 한 줄 로그뿐이며 그 파일에는 실행 명령이 기록돼 있지 않다.
- Claude Code가 중복된 `--effort` 중 어느 값을 적용하는지 미관측. 실제 `claude --effort max` 세션, GLM·gpt 공급자에서의 `--effort max` 동작 미관측.
- AC-009 변이 3종은 재실행하지 않고 로그를 읽었다.
- `make build` 미실행(카탈로그 해시는 드라이런 비교로 대체). 골든 파일이 `-update-golden`으로 재생성됐는지는 출처 확인 불가.
- 웹 콘솔 실제 렌더링, TUI 위저드 대화형 실행 미관측. `internal/cli`·`internal/web` 패키지 전체 커버리지 미측정.
- `moai spec lint .moai/specs/SPEC-MODEL-OPUS55-001` → "✓ No findings" 관측, 단 파이프 때문에 moai 자체 종료 코드는 미포착.

### Residual risk

- F1 수리 전까지 `moai cc … -- --effort <x>` 형태로 명시 effort를 준 운영자는 max 프로필에서 자기 선택이 무시될 수 있다.
- K6(별칭 → `claude-opus-5-5[1m]`, Claude Code v2.1.280 미만에서 미지 모델)는 버전 하한 없이 남는다.
- `--effort max`는 공식 문서상 명시적 선택(환경변수와 같은 층)이다. 세션 중 `/effort` 변경이 허용되는지는 관측하지 않았다.
