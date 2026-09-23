# t1094 sync-audit — Opus 5.5 docs-site 4-locale + README 반영

- card: t1094 (docs 전용, 자체 SPEC 없음) · branch `WT-opus55-docs` · HEAD `4fe50d487`
- 감사 대상: 카드 커밋 `a29db12ec`, `4fe50d487` (병합 `74e713709`·`4586b1f00` 은 흡수이므로 제외)
- 기준: t1089 `SPEC-MODEL-OPUS55-001` (spec.md) + 착지 코드 + 공식 문서 2건
- 평가 프로필: `default` (harness.yaml `default_profile`) — 평면 가중 모드, must-pass = Functionality + Security
- 교차 모델 감사: `audit_model` 미설정 → Claude 단독 감사

## 종합 판정: **PASS-WITH-DEBT**

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 88/100 | PASS | AC1·AC2·AC3·AC5·AC6 충족. 차단 결함 0. 감점: 벤치마크 귀속 문장이 no-haiku·README 에만 있고 profile-matrix·comparison 에는 없음(F3), 위저드 화면 인용의 질문 제목 줄 불일치(F2) |
| Security (25%) | 100/100 | PASS | 추가 88줄에서 URL 0건, 금지 URL 0건, 비밀·토큰 패턴 0건 |
| Craft (20%) | 85/100 | PASS | hugo exit 0·경고 0. 커버리지는 docs 전용이라 해당 없음(N/A — FAIL 아님). 감점: 판정서 정확성 결함(F5·F6), Max 행 설명이 Medium 과 구분되지 않는 특징을 내세움(F4) |
| Consistency (15%) | 80/100 | PASS | Mermaid TD 유지, 본문 이모지 0, 4 로케일 동시 커밋. 감점: 이 카드가 init-wizard 에서 고친 「High 가 두 에이전트에 `max`」 서술이 다른 4개 페이지 × 4 로케일에 그대로 남아 페이지 간 모순이 생김(F1) |

- 조화평균: 4 / (1/88 + 1/100 + 1/85 + 1/80) = **87.7**
- 가중평균(참고): 89.2
- must-pass 방화벽: Functionality PASS, Security PASS → 통과
- FAIL 을 만드는 차단 결함 없음. 발견 사항은 전부 optional(선택) — 처리 여부는 리드 재량

## 인수조건별 대조

| AC | 판정 | 근거 |
|---|---|---|
| 1. 현재 동작 서술은 Opus 5.5 · v2.1.280+ · 기본 effort `medium` | PASS | model-policy 4 로케일 「기본 effort」 인용문, README 4종 246행, 임계표·statusline 다이어그램·명령 표·moai-run callout 치환 확인. 잔여 `Opus 5` 적중 66줄을 전수 분류한 결과 현재-동작 서술 누락 없음(벤치마크·스크린샷 캡션·날짜 스냅샷·xhigh 지원 목록·캐시 표의 Opus 5 행뿐) |
| 2. 측정·역사 사실 보존 | PASS (F3 선택 부채) | no-haiku 표·README DeepSWE 표·profile-matrix·comparison·캡션·「2026-08 기준」 서술 모두 원문 유지. 명시적 「Opus 5 에서 측정」 문장은 no-haiku(4)·README(4) 에만 추가 |
| 3. 착지 코드·SPEC 값 일치 | PASS | 아래 코드 대조 표 참조 |
| 4. docs-site HARD 규칙 | PASS | ko 정본 + en/ja/zh 동일 커밋, Mermaid LR/RL 0, 이모지 0, 금지 URL 0, 로케일 자연 관용(직역 흔적 없음). ko 전용 적중 파일은 파생 로케일에 해당 내용 자체가 없음을 확인 — 기존 드리프트이며 이 카드 결함 아님 |
| 5. 공식 사실 | PASS | 두 공식 페이지를 이 실행에서 재조회해 대조(아래) |
| 6. 빌드 | PASS | `hugo --minify --quiet` exit 0, 출력 0줄 |

### 코드 대조 (이 트리에서 읽음)

| 문서 서술 | 코드 위치 | 일치 |
|---|---|---|
| 위저드 화면 3줄 (Max / Medium (Recommended) / Low, 범위·플랜) — 4 로케일 | `internal/cli/wizard/questions.go:121-123`, `translations.go:93-95 / 183-185 / 273-275` | 일치 (라벨·설명 바이트 단위 일치) |
| ▸ 커서가 Medium | `questions.go:125` `Default: "medium"` | 일치 |
| 「위저드·웹 콘솔이 세션 effort 로 `medium` 권장」 | `profile_setup_translations.go:177/272/367/462`, `internal/web/assets/i18n.js:443/1357` | 일치 |
| Max 행 「감사·자문·조율 에이전트 `high` 유지」 | `profile_matrix.go:324-337` (plan/sync-auditor, super-advisor, design, lead = high) | 사실로는 참 (F4 참조) |
| Low 행 「대부분 Opus `medium`」 | `profile_matrix.go:354-367` (Opus 9행 중 6행 medium) | 일치 |
| `max` 셀 없음 | `profile_matrix.go:291` 주석 + 셀 전체 | init-wizard 는 일치, 다른 페이지는 불일치 (F1) |
| `max` 전달 경로(`--effort`) | `launch_effort_settings.go:68-71` | 문서에 `effortLevel`/`--effort` 서술 0건 → 모순 없음 |

### 공식 사실 대조 (이 실행에서 curl 재조회)

- `platform.claude.com/docs/en/models/opus-5-5/overview`: 「claude-opus-5-5 … Context window 1M tokens Max output 128K tokens Input pricing $4 / MTok Output pricing $20 / MTok」, 「Default effort medium」, 「The minimum cacheable prompt length is 512 tokens.」
- `code.claude.com/docs/en/model-config`: 「Opus 5.5 requires Claude Code v2.1.280 or later.」, 「high on every model that supports effort, except that Opus 5.5 defaults to medium, Opus 4.7 defaults to xhigh」, 「Opus 5.5, Opus 5, Sonnet 5, Opus 4.8, and Opus 4.7 low , medium , high , xhigh , max」
- 문서의 「다른 effort 지원 모델은 대부분 `high` 기본」은 Opus 4.7(`xhigh`) 예외를 포괄하는 정확한 표현
- 판정서 §6.1 의 전제(Fable 5.1 1M, Sonnet 5 1M)도 같은 페이지 비교표 「Claude Fable 5.1 1M 128K」「Claude Sonnet 5 1M 128K」로 확인됨

## 발견 사항 (구조화 결함 목록)

- **F1** [Medium] [optional] 확신도 높음 — `docs-site/content/{ko,en,ja,zh}/getting-started/cli.md` (ko:454, en/ja/zh:452), `getting-started/introduction.md` (ko:146·151, en:147·152, ja/zh:146·151), `core-concepts/what-is-moai-adk.md` (ko:528, en:526, ja:529, zh:527), `advanced/tokenomics-overview.md` (ko:57, en:106, ja:93, zh:63) — 「high 티어는 호출 빈도가 가장 낮은 두 에이전트에 `max`」, 「low 는 에이전틱 행을 모두 Opus `low`」 서술이 남아 있다. `profile_matrix.go:291` 「`max` is absent from every cell」과 모순이며, 이 카드가 `4fe50d487` 로 init-wizard 에서만 같은 서술을 고쳐 이제 페이지끼리 서로 다른 말을 한다. 서술 자체는 t1089 이전(t205 매트릭스 개정)부터 낡은 것이라 이 카드가 만든 결함은 아니고 카드 축(Opus 5.5)의 범위 밖이다. 다만 ko/en tokenomics-overview 는 이 카드가 손댄 파일이다. — Required fix: 별도 카드로 16개 지점을 `profile_matrix.go` 기준 문구(Max: 판정 행 `high` 유지 · `max` 셀 없음, Low: 대부분 Opus `medium`)로 맞춘다.
- **F2** [Low] [optional] 확신도 높음 — `docs-site/content/ko/getting-started/init-wizard.md:91` (en/ja/zh:93) — 화면 인용의 질문 제목 줄 「? 성능 티어 선택:」 / 「? Choose the performance tier:」가 실제 위저드 제목(`questions.go:106` "Select model policy", `translations.go:90` 「모델 정책 선택」, ja 「モデルポリシーを選択」, zh 「选择模型策略」)과 다르다. 커밋 `4fe50d487` 메시지는 「각 로케일의 위저드 문구를 인용한다」고 했지만 옵션 3줄만 해당된다. — Required fix: 제목 줄을 각 로케일 위저드 제목으로 바꾸거나, 화면 블록이 요약임을 밝힌다.
- **F3** [Low] [optional] 확신도 중간 — `docs-site/content/ko/advanced/profile-matrix.md:73`, `en|ja|zh/advanced/profile-matrix.md:15·62`, `*/getting-started/comparison.md` (ko:64, en/ja/zh:42) — Opus 5 벤치마크를 근거로 「그래서 모든 에이전틱 행에 Opus 를 둔다」고 설명하지만, 현재 `opus` 가 Opus 5.5 이고 재측정 전이라는 귀속 문장이 없다. 모델명을 문장 안에서 밝히고 있어 사실 오류는 아니며, no-haiku·README 와 처리 수준만 다르다. 코드 주석(`profile_matrix.go:273·280`)은 「(measured on Opus 5)」를 달았다. — Required fix: profile-matrix 4 로케일에 no-haiku 와 같은 한 문장을 추가할지 리드가 정한다.
- **F4** [Low] [optional] 확신도 높음 — `docs-site/content/ko/getting-started/init-wizard.md:99` (en/ja/zh:101) — Max 행 「감사·자문·조율 에이전트가 `high`를 유지」는 참이지만 Medium 열도 같은 행을 `high` 로 둔다(`profile_matrix.go:339-352`). Max 와 Medium 을 실제로 가르는 것은 `builder-harness`(high↔medium)와 `e2e-tester`(medium↔low) 두 행이다(profile-matrix 페이지 ko:50 도 그렇게 서술). 독자가 Max 의 차별점으로 오해할 수 있다. — Required fix: Max 행을 「Medium 보다 `builder-harness`·`e2e-tester` 를 한 단계 높임」처럼 차이를 드러내는 문구로 고친다.
- **F5** [Low] [optional] 확신도 높음 — `.moai/reports/t1094/verdict.md:5`, `:65` — 머리말 상태 「t1089 develop 착지 전」과 Gaps 의 「t1089 위저드·웹 라벨 아직 미착지」는 §7(착지 후 대조)로 이미 해소됐는데 갱신되지 않았다. 또 `:80` 의 「3.」 항목이 §6 목록에서 떨어져 §7 뒤에 붙어 있다. — Required fix: 머리말·Gaps 를 착지 후 상태로 고치고 §6 항목 3 을 제자리로 옮긴다.
- **F6** [Low] [optional] 확신도 높음 — `.moai/reports/t1094/verdict.md:71-79` — §7 은 「두 에이전트에 `max`」가 코드와 모순임을 찾아 init-wizard 만 고쳤다고 적었지만, 같은 서술이 16개 지점에 남아 있다는 사실(F1)을 잔여 위험에 기록하지 않았다. 발견한 결함 부류의 나머지 모집단을 재지 않은 누락이다. — Required fix: §6 잔여 위험에 F1 목록을 추가한다.

판정서의 나머지 수치는 이 실행에서 재현했다: 편집 전 적중 수(ko 18/30, en 12/25, ja 10/22, zh 9/21, README 4/32 — `git grep` 을 `a29db12ec^` 트리에 실행해 정확히 일치), diff 규모(판정서 제외 37 files, +76/−44), `effortLevel|--effort|CLAUDE_CODE_EFFORT_LEVEL` 0건, hugo exit 0.

## 권고

- 차단 결함이 없으므로 병합 창 요청 가능. F1 은 이 카드 축 밖의 기존 드리프트이므로 별도 카드로 분리하는 편이 범위 규율에 맞다.
- F2·F4 는 이 카드가 쓴 줄이라 같은 워크트리에서 적은 비용으로 고칠 수 있다(리드 재량).
- F5·F6 은 판정서 문서 정정이다.

## 근거 (Evidence)

| 명령 | 출력 |
|---|---|
| `cd docs-site && hugo --minify --quiet` (hugo v0.160.1+extended) | `hugo exit 0`, 로그 0줄, warn/error 0 |
| `git show a29db12ec 4fe50d487 -- docs-site README*` 의 추가 줄 추출 | 88줄 |
| 추가 줄 금지 URL `docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr` | 0 |
| 추가 줄 `(flowchart\|graph) (LR\|RL)` | 0 |
| 추가 줄 이모지 U+1F300–1FAFF · U+2600–27BF | 0 |
| 추가 줄 `https?://` / 비밀 패턴 | 0 / 0 |
| 빌드 후 `git status --short` | 비어 있음 (public/ 은 무시 대상) |
| 편집 전 `git grep` (a29db12ec^) | README 4/32 · en 12/25 · ja 10/22 · ko 18/30 · zh 9/21 |

## 미관측 (Gaps)

- Vercel 프리뷰 렌더는 보지 않았다(push 전).
- 실제 `moai init` 위저드를 띄워 화면 렌더(라벨·설명 구분자, 커서 위치)를 관측하지 않았다 — 코드 문자열 대조로 대신했다.
- `scripts/docs-i18n-check.sh` 가 이 트리에 없어 로케일 패리티는 grep 과 수동 대조로만 확인했다.
- 교차 모델(codex/GLM) 감사는 설정이 없어 실행하지 않았다.

## 잔여 위험

- 날짜 스냅샷(「2026-08 기준」 라인업)은 참이지만 독자가 최신으로 읽을 수 있다.
- model-policy 라인업의 Fable 5(256K)·Sonnet 5(200K) 행과 prompt-caching 표의 같은 행은 공식 문서(Fable 5.1 1M, Sonnet 5 1M)와 다르다 — 판정서 §6.1 이 이미 별도 카드 후보로 올렸다.
- 벤치마크 수치가 Opus 5.5 에서도 같은 순위를 보인다는 보장은 없다 — 문서는 「미재측정」이라고 밝혔다.
