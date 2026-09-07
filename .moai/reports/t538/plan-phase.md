# t538 plan 페이즈 증거 — 측정-선행 스윕

- 카드: t538 · 브랜치: `WT-docs-v313-locales` · 베이스: `bce6d7e08` (= origin/develop 팁, 리드 좌표와 일치 재유도)
- SPEC: 미생성 — 이 파일이 manager-spec 위임의 입력
- 모든 실측은 2026-09-08, 워크트리 `.claude/worktrees/t538` @ `bce6d7e08` 에서 이번 실행으로 수행 (VCI §2)
- 재측정 도구: `/usr/bin/grep` (셸 grep 은 ugrep 래퍼 — 조용히 건너뜀, [[feedback_the_shell_grep_here_is_a_ugrep_wrapper_that_skips_silently]])

## 배차 전제 정정 1건 [중요]

리드 배차문의 전제 "docs-site 4-locale 에 v3.1.3 변경사항이 없습니다(codex 듀얼 하네스 포함)"는 **절반은 낡았다**:

- **codex 듀얼 하네스는 이미 문서화돼 있다**: `docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md` ×4 실재 (각 64행). 작성: t274 `SPEC-DOCS-V313-CATCHUP-001` (커밋 `175d63f3f`, 2026-08-26 착지, PR #1662). 갱신: t496 `SPEC-CODEX-EVENT-COVERAGE-001` (커밋 `732609dcf`, 2026-09-07, 12-이벤트 어댑터 커버리지 반영 — 페이지가 `RenderHooks`·`SubagentStart/Stop` 적응까지 서술)
- **v3.1.3 나머지 주요 항목도 전부 문서 존재** (spot-check, ×4 로케일): `/moai feedback` 스크러빙(utility-commands/moai-feedback.md) · `workflow.todo.enabled`+`moai todo analyze`(moai-todo.md) · `/moai gate` typecheck 축(moai-gate.md) · MCP `project_root`(advanced/multi-model-audit.md + guides/mcp-server.md) · design-dna 다이어그램 프로파일(advanced/skill-guide.md:159) · update symlink 폴백 통지(cli-reference/update.md:119) · model-policy 판정가중 정책+manager-lead 행+GLM reasoning 상한 `max`(multi-llm/model-policy.md:116/158/236-249) · `ANTHROPIC_*`(settings-json·security-notes·profile-matrix·factory-mode·launchers·multi-llm _index)
- `hugo.toml:55` `version = "v3.1.3"`, statusline.md·faq.md 스탬프 `🗿 v3.1.3` — 버전 동기축도 t274 가 해소
- docs-site `changelog/` 페이지는 **설계상 GitHub Releases 링크 페이지**(본문에 릴리스 노트를 옮겨 적지 않기로 명시) — 여기 v3.1.3 이 없는 것은 결함이 아님
- 결론: **v3.1.3 캐치업 축은 D(이미 반영)** — t535 verdict §C5 가 예고한 대로. 이 카드에서 codex 관련 신규 작업 없음

## 실제 갭 3건 — 이 카드의 실질

### G1. e2e desktop-native 레인 — en/zh 스테일 (ko·ja 는 완비)

`docs-site/content/{ko,en,ja,zh}/utility-commands/moai-e2e.md` 4-locale 대조:

| 축 | ko (201행) | ja (201행) | en (195행) | zh (195행) |
|---|---|---|---|---|
| 플래그 행 `--platform …desktop-native` | ✅ :58 | ✅ :58 | ❌ `web\|mobile\|desktop` | ❌ 동일 |
| 툴체인 매트릭스 3-OS 행 | ✅ :78-80 | ✅ :78-80 | ❌ 0행 | ❌ 0행 |
| 자동감지 표 desktop-native 행 | ✅ :98 | ✅ :98 | ❌ | ❌ |
| "대상 없음" 절 라우팅 설명 | ✅ :176 | ✅ :176 | ❌ **:170 deferral notice 잔존** ("native-desktop automation **is not yet provided**") | ❌ :170 동일 (中文) |
| 호스트 OS 규칙 문단 | ✅ 매트릭스 직후 | ✅ | ❌ | ❌ |

- **결함의 성격**: en·zh 는 기능이 착지한 뒤에도 "아직 제공되지 않는다"고 **기능 부재를 주장** — 부정확 문서. SPEC-DESKTOP-NATIVE-E2E-001 이 제거하라고 명시한 deferral 문장이 en·zh 에 생존
- ko·ja 는 완비(원어 용어 데스크탑-네이티브/デスクトップネイティブ — ASCII 토큰 grep 으로는 안 잡히므로 원어 토큰으로 재측정할 것). en·zh 만 수리 대상
- 리드가 카드 요지에 넣은 "desktop 지원"의 실체 = 이것. (desktop-native 는 v3.1.0대 착지로 v3.1.3 과는 별개 축 — 배차문의 버전 귀속만 부정확)
- 배지: ko·ja 행이 new-badge 없음 → 수리도 배지 없음 (착지 현실에 대한 정렬, 신규 기능 아님)

### G2. doctor.md 예시 블록 — ja·zh 2줄 부족

`docs-site/content/{l}/cli-reference/doctor.md` `## 例`/`## 示例` 블록:

- ko(:115-122)·en: `moai doctor permission` + `moai doctor sandbox` 예시 행 **있음** (ko:120-121, en:118-119)
- **ja(:106-111)·zh(:106-111)**: 예시 4행만 — `hook` 에서 끝남, `permission`·`sandbox` **누락**
- 명령 표(:32-33)에는 4-locale 모두 permission·sandbox 행 존재 → 갭은 예시 블록 한정
- t535 sync-audit F2 는 zh 만 지목 — **측정 정정: ja 도 누락**

### G3. doctor.md 기존 절 강조 간격 위반 3곳 (t535 F2 그대로 재측정 확인)

- `ko:41` · `ja:39` · `zh:39` — 기존 Home Disk Usage 절의 강조+괄호 패턴 (t535 run 이 신규 절은 0위반으로 착지했고 기존 절은 기록만 함)

## 범위 판정 (plan에서 결정할 사항)

- **선택 후보**: v3.1.3 Added 중 skill-guide.md 가 미언급한 축 — svg-infographic 접근성 이름(SVG060-064)·커넥터 기하(SVG070-074). 1-2행 추가면 가능. **권고: 포함** (v3.1.3 캐치업 카드의 존재 이유를 완결)
- **[HARD] 4-locale 동일 착지**: 변경 집합 = e2e en·zh + doctor ja·zh(+ko 간격 1행) + (선택) skill-guide ×4 — 전부 한 브랜치·한 커밋 체인. ko e2e 는 무변경(이미 정확)
- ko 정본 → en/zh 파생. CHANGELOG [Unreleased] Docs 항목 (t535 선례)

## Gaps (이번 실행에서 관측하지 않은 것)

- v3.1.3 **Fixed** 항목(동작 수리류)의 문서 요구성 전수 판정은 안 함 — 동작 수리는 통상 문서 표면 불요, t274 인벤토리가 판정한 축을 계승
- en/zh e2e 페이지의 문단 단위 전수 diff 는 run 페이즈 스페셜리스트 몫 — 본 측정은 축(level) 단위
- SVG060-074 가 skill-guide 이외 표면(예: 스킬 자체 문서)에서 이미 언급됐는지 전수 grep 아님 — skill-guide 기준으로만 판정

## Residual-risk

- 리드 배차 전제의 나머지 절반(desktop 지원)은 실재했으나 **v3.1.3 이 아니라 v3.1.0대 부채** — SPEC 이 이 귀속을 정확히 적어야 훗날 감사가 헷갈리지 않는다
- ko·ja 가 완비인 경로(어느 커밋에서)를 역추적하지 않았다 — 현재 트리 상태만이 카드의 근거이며 추적은 본 카드 불요

## 정정 부기 (2026-09-08, plan-audit iter1의 F2·F4 접수분 — 초판은 지우지 않고 덧붙임)

- **F4 (행수)**: 재측정 `/usr/bin/wc -l` = 64 ×4 — 초판의 "각 64행" 표기는 본 방식으로 유지된다. 감사 보고서의 "실측 63행"은 산출 방식이 기재돼 있지 않아 불일치로 기록 (본 파일의 64는 wc -l 기준임을 명시).
- **F2 (deferral 문장 계보 과장)**: G1 "결함의 성격" 문단의 「SPEC-DESKTOP-NATIVE-E2E-001 이 제거하라고 명시한 deferral 문장이 en·zh 에 생존」은 과장이었다 — 해당 SPEC 이 verbatim 제거를 명시한 문장("There is no opt-in automation path for `desktop-native`")의 대상은 workflow·agent 트리였고, en:170·zh:170 의 문장("native-desktop automation is not yet provided…")은 **같은 취지(deferral 통지)의 별개 문장**이다. 결함의 성격 자체(en·zh 문서가 이미 착지한 기능의 부재를 주장)는 동일하게 성립하며 G1 범위 판정은 변함없다.
