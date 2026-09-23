# t1094 판정서 — Opus 5.5 docs-site 4-locale + README 반영

- card: t1094 · Tier S · branch `WT-opus55-docs` · base 로컬 develop `08113ff0f`
- 정본: t1089 `SPEC-MODEL-OPUS55-001` (spec.md §A.2·§C, plan.md §C.2·§C.3·§C.6·§C.8) — 읽기만 함
- 상태: **t1089 착지(`e52ba05e7`) 흡수·대조 완료, sync-audit PASS-WITH-DEBT 87.7 (차단 0) — 병합 창 요청 단계** (순번: t1112 다음)

## 1. 공식 근거 (이 실행에서 재조회)

| 사실 | 출처 |
|---|---|
| `claude-opus-5-5`, 1M 컨텍스트, 128K 출력, $4/$20, 기본 effort `medium`, 최소 캐시 프롬프트 512 토큰 | platform.claude.com/docs/en/models/opus-5-5/overview |
| `opus` 별칭 → Opus 5.5, Claude Code v2.1.280 이상 필요 | code.claude.com/docs/en/model-config |
| xhigh/max 지원: Opus 5.5 · Opus 5 · Sonnet 5 · Opus 4.8 · Opus 4.7 (+Fable 5.1/5) | 같은 페이지 |
| 기본 effort: Opus 5.5 `medium`, Opus 4.7 `xhigh`, 나머지 effort 지원 모델 `high` | 같은 페이지 |

## 2. 전수 측정 (편집 전)

명령: `grep -rnE 'Opus 5([^.0-9]|$)|opus-5([^-]|$)'` (opus-5-5 제외)

| 범위 | 파일 | 줄 |
|---|---|---|
| docs-site ko | 18 | 30 |
| docs-site en | 12 | 25 |
| docs-site ja | 10 | 22 |
| docs-site zh | 9 | 21 |
| README 4종 | 4 | 32 (각 8) |

ko 에만 있는 적중 파일 6개(claude-code/_index, context-window, features-overview, how-claude-code-works, multi-llm/_index, moai-run)는 **기존 로케일 드리프트**이며 이 카드에서 새로 만든 것이 아니다.

## 3. 분류와 처분

### 치환 (현재 동작 서술)
- 컨텍스트 임계 표·다이어그램 `Opus 5 (1M)` / `(Opus 5, GLM-5.3)` → Opus 5.5 — token-budget, tokenomics-overview, statusline, moai-run
- 모델 표: commands, how-claude-code-works, context-window(`Opus 5.5 / Opus 5 / Opus 4.8`), multi-llm/_index(`5.5 / 5 / 4.8`), model-policy 라인업 `opus` 행 → `Claude Opus 5.5`
- xhigh/max 지원 목록에 Opus 5.5 추가 (Opus 5 유지 — 여전히 선택 가능)
- init-wizard 화면 예시 3줄 `Opus 5 (` → `Opus 5.5 (`
- prompt-caching 표에 `Claude Opus 5.5 | 1M | 512` 행 추가 (Opus 5 행 유지)
- model-policy 라인업 제목 날짜 `(2026-08)` → `(2026-09)` + 「기본 effort」 단락 신설: Opus 5.5 기본 `medium`, 다른 모델 대부분 `high`, 위저드·웹 콘솔이 `medium` 권장, `opus`→5.5 는 CC v2.1.280+

### 보존 + 귀속 명시 (측정 사실)
- no-haiku-3tier 벤치마크 표(Opus 5 × 5 effort)·단가줄·해설 — 원문 유지, 표 아래에 「Opus 5 에서 측정, Opus 5.5 미재측정」 한 문장 추가 (4 로케일)
- README 벤치마크 표·해설 — 원문 유지, 해설 뒤에 귀속 + 현재 `opus`=Opus 5.5 · v2.1.280+ · 기본 effort `medium` 한 문장 추가 (4종)

### 보존 (손대지 않음)
- profile-matrix, comparison 의 벤치마크 인용 문장 — 측정 사실
- `_index.md` / README 52행 스크린샷 캡션(「Plan을 Opus 5 high로」) — 그 화면의 사실 기록
- 날짜가 박힌 라인업 서술 — claude-code/_index(2026년 8월 기준), features-overview(2026-08 기준), tools-reference(2026년 8월 현재) — 날짜 스냅샷으로 참
- agentic/_index 「Opus 4.7+/4.8/5 는 서브에이전트를 자동 spawn 하지 않음」 — Opus 5.5 에 대한 확인 근거 없음, 옮기지 않음
- CHANGELOG·릴리스 노트 — 대상 아님

## 4. 검증

| 항목 | 명령 | 결과 |
|---|---|---|
| 잔여 현재-동작 적중 | 위 grep 에서 보존 목록 제외 | 0 건 (exit 1) |
| diff 규모 | `git diff --stat` | 37 files, +76 / −44 |
| hugo 빌드 | `hugo --minify --quiet` | exit 0, warn/error 0 줄 |
| 금지 URL·Mermaid LR | 추가 줄 grep | 0 |
| 본문 이모지 | 추가 줄 U+1F300–1FAFF·2600–27BF | 0 |
| 변경 경로 | `git diff --name-only` | docs-site/content/{ko,en,ja,zh} + README 4종 뿐 |

## 5. 미관측 (Gaps)
- Vercel 프리뷰 렌더 미확인 (push 금지)
- `scripts/docs-i18n-check.sh` 는 이 트리에 없음 — 로케일 패리티 기계 검사 불가
- 실제 `moai init` 위저드 화면은 띄워 보지 않음 — 코드 문자열 대조로 대신함 (§7). (종전 Gap 「t1089 라벨 미착지」는 §7 에서 착지본 대조로 해소)

## 6. 잔여 위험 · 리드 판단 요청
1. **범위 밖 드리프트 발견**: model-policy 라인업 표의 Fable 행(`Fable 5`, 256K)과 Sonnet 5 컨텍스트(200K)는 공식 문서(Fable 5.1 1M, Sonnet 5 1M)와 다르다. prompt-caching 표도 같은 두 행이 낡았다. 이 카드는 Opus 5.5 만 다뤄 손대지 않았다 — 별도 카드 후보.
2. ~~init-wizard 화면 예시 불일치~~ — §7 에서 해소.
3. 날짜 스냅샷 라인업(8월 기준)은 참이지만 독자가 「최신」으로 읽을 수 있다. 갱신 여부는 Fable 5.1 반영과 함께 결정하는 편이 맞다.
4. **(sync-audit F1)** 「high 티어는 호출 빈도가 가장 낮은 두 에이전트에 `max`」 서술이 cli.md · introduction.md · what-is-moai-adk.md · tokenomics-overview.md × 4 로케일 = 16곳에 남아 있다. `profile_matrix.go:291`(max 는 어느 칸에도 없음)과 모순이며, 이 카드가 init-wizard 만 고쳐 페이지 간 불일치가 생겼다. t1089 이전부터의 낡은 서술 — 별도 카드 후보.

## 7. t1089 착지 후 대조 (흡수 `e52ba05e7`)

- 흡수: `176d8b658` → 흡수 커밋 `74e713709` (충돌 0), `e52ba05e7` → 충돌 0 (README 4종 자동 병합, 귀속 문장 각 1개 유지)
- t1089 최종 동작과 문서 대조:
  - 위저드 세션 effort `medium - 균형 (권장)` (`internal/cli/profile_setup_translations.go:272`), 웹 `f.effort_level.opt.medium` = `중간 (권장)` (`internal/web/assets/i18n.js:1357`) → model-policy 「위저드·웹 콘솔이 `medium` 권장」 문장과 일치
  - `max` 는 settings `effortLevel` 이 아닌 `--effort max` 실행 인자로 전달 (`internal/cli/launch_effort_settings.go`) → docs-site·README 어디에도 `effortLevel`/`--effort` 서술 없음 (`grep -rnE "effortLevel|--effort|CLAUDE_CODE_EFFORT_LEVEL"` 0건), 모순 없음
  - 위저드 성능 티어 라벨 (`internal/cli/wizard/questions.go:121-123`, `translations.go:93-95/183-185/273-275`) 과 init-wizard 화면 예시가 달랐음 → 4 로케일 화면 3줄을 각 로케일 코드 문구로 교체(Max / Medium 권장 / Low, 범위·플랜 포함)
  - 표의 「두 에이전트에 `max`」는 `profile_matrix.go:291` 「`max` is absent from every cell」과 모순, 「Low 는 Opus `low`」는 같은 주석(대부분 `medium`)과 모순 → Max/Low 행을 코드 기준으로 수정, Medium 에 권장 표기
- 흡수 + 수정 후 `hugo --minify --quiet` exit 0, warn/error 0

## 8. sync-audit 대응 (`.moai/reports/t1094/sync-audit.md`, PASS-WITH-DEBT 87.7, 차단 0)

| 발견 | 처분 |
|---|---|
| F1 max 서술 16곳 | 범위 밖 — §6-4 로 리드에게 상신 |
| F2 화면 제목 「성능 티어 선택」≠ 코드 「모델 정책 선택」 | 수리 — 4 로케일 제목을 `questions.go:106` / `translations.go:90·180·270` 문구로 교체 |
| F3 profile-matrix·comparison 에 미재측정 문장 없음 | 유지 — 두 페이지 모두 해당 문장 안에서 「Opus 5」를 측정 모델로 이미 명시. 귀속은 충족, 문장 추가는 보류 |
| F4 Max 행이 Medium 과 구분 안 됨 | 수리 — 「Medium 과 같되 `builder-harness`·`e2e-tester` 만 한 단계 높은 effort」 (`profile_matrix.go:333-334` high/medium vs `:348-349` medium/low 확인) |
| F5/F6 판정서 상태·Gaps·§6 배치 | 수리 — 머리말·Gaps·§6 갱신 (이 절 포함) |
