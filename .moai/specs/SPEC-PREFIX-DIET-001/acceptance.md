---
id: SPEC-PREFIX-DIET-001
title: "acceptance — 세션 시작 prefix 다이어트 2단계"
version: "0.2.0"
created: 2026-10-03
updated: 2026-10-03
---

# acceptance.md — SPEC-PREFIX-DIET-001

## §D. AC 매트릭스

- 분류: **RB** = 릴리스 차단(새 의무 — RED-now 필수), **RG** = 회귀 가드(이미 성립하는 성질을 지킨다 — RED-now 를 요구하지 않음).
- **트리 핀**: 모든 RED-now 측정은 커밋 `5d5ff1aae`(워크트리 `t1450`, 브랜치 `WT-prefix-diet-stage2`)에서 이 plan 실행이 직접 돌린 것이다. 문서 수준 핀이며 개별 행에 핀이 없으면 이 핀이 묶는다. 이후 plan 커밋은 `.moai/specs/SPEC-PREFIX-DIET-001/` 만 바꿨으므로 `HEAD` 에서 같은 값이 나온다.
- 명령은 워크트리 루트에서 적힌 그대로 실행하며 단일 호출이다. exit 코드는 별도 칸이다. `go test` 의 초록 판정은 언제나 `-v` 출력의 이름 붙은 `--- PASS: <이름> ` 줄이며, 패키지 요약 `ok` 줄은 테스트를 0개 골라도 찍히므로 근거가 되지 않는다. `-run` 은 `^…$` 로 앵커한다.
- 표 셀 안의 `\|` 는 마크다운 표 이스케이프이며, 실제 명령에서는 `|` 하나다.
- 출력 길이가 한계(50줄·2KB)를 넘는 명령은 파일로 리다이렉트하고 종료 코드와 꼬리만 기록한다. 증거 원문은 실행 단계에서 `.moai/reports/t1450/verdict.md` 와 `progress.md` §E.2 로 **반출**한다(`.moai/state/` 는 머신 로컬이라 인용 대상이 아니다).
- 범위 판정 명령의 기준 SHA(`<BASE>`)는 앵커 `5d5ff1aae` 이다. 실행 중 develop 을 흡수했다면 읽는 시점에 `git merge-base develop HEAD` 를 따로 실행해 출력된 SHA 를 `<BASE>` 로 쓰고 `progress.md` 에 기록한다. 병합 전 평가 전용이다.
- RED-now 칸의 EL-n 은 §D.1 증거 원장이다.

| AC | REQ | 분류 | 기준 | RED-now (원장 ID) | 초록 조건 — 명령 → 기대 출력 (exit) |
|---|---|---|---|---|---|
| AC-PFD-001 | REQ-PFD-001, REQ-PFD-002 | RB | 파일 전체 UTF-16 예산 테스트 존재, 상수 ≤ 앵커, 세 파일 모두 앵커보다 작음 | EL-2: 테스트 부재 exit 1. EL-1: 29243 / 28517 / 62593 | `grep -rl TestOutputStylesCharBudget internal/template` → 테스트 파일 경로 1줄 (0). `go test ./internal/template/ -run '^TestOutputStylesCharBudget$' -count=1 -v` → `--- PASS: TestOutputStylesCharBudget ` 과 `output-style=moai <N>`, `output-style=moai-easy <N>`, `output-style=moai-learn <N>` 3줄, 각 N 이 EL-1 값보다 작고 `progress.md` §E.2 의 파일별 상수 이하 (0). 파일별 `python3 -c "import sys;t=open(sys.argv[1],encoding='utf-8').read();print(len(t.encode('utf-16-le'))//2)" <파일>` 가 로그의 N 과 같음 (0) |
| AC-PFD-002 | REQ-PFD-003, REQ-PFD-004 | RB | 구속 원장 존재·완전, `verbatim` 은 바이트 동일, 구속 토큰 개수 보존 | EL-3: 고정물 부재 exit 1. EL-2: 테스트 부재 | `go test ./internal/template/ -run '^TestOutputStyleBindingLedger$' -count=1 -v` → `--- PASS: TestOutputStyleBindingLedger ` (0). 변이 하위 테스트가 각각 `--- PASS`(실패를 기대하는 단언이 통과): 행 없는 앵커 단위, `binding` 단위를 다른 종류로 표기, `binding`/`normative` 행의 `dropped`, 이유 없는 `dropped`, 생존 참조 없는 `dropped`, before≠after 인 `verbatim`, 구속 토큰이 있는 `dropped`, 토큰 하나 빠진 파일 합계. 각 변이 입력에서 실제 FAIL 이 관측된 출력을 기록한다 |
| AC-PFD-003 | REQ-PFD-003 | RB | 원장에 `rewrite` 행 0 (리더 결정 D3) | EL-12: 고정물 부재로 `FileNotFoundError` exit 1 | `python3 -c "import json;print(sum(1 for r in json.load(open('internal/template/testdata/output_style_ledger.json'))['rows'] if r['treatment']=='rewrite'))"` → `0` (0). `rewrite` 행 하나를 넣은 픽스처를 원장 테스트에 먹이면 `--- FAIL` 이고 하위 테스트 `TestOutputStyleBindingLedger/rewrite_row_rejected` 가 `--- PASS`(거부 단언 통과) |
| AC-PFD-004 | REQ-PFD-004 | RB | Go 테스트와 독립인 구속 토큰 개수 명령: 앵커와 같은 개수 | EL-10: `moai` 89/4/27/0 · `moai-easy` 33/0/0/0 · `moai-learn` 24/0/7/0 | 파일마다 `python3 -c "import re,sys;s=open(sys.argv[1],encoding='utf-8').read();print('HARD=%d MUST_NOT=%d MUST=%d shall=%d'%(s.count('[HARD]'),s.count('MUST NOT'),len(re.findall(r'MUST(?! NOT)',s)),s.count('shall ')))" <파일>` → EL-10 과 같은 값 (0). `[HARD]` 줄을 `dropped` 로만 줄이는 설계이므로 개수는 줄지 않는다. 한 토큰을 지운 변이 본문에서 같은 명령이 다른 값을 낸다는 관측을 기록한다 |
| AC-PFD-005 | REQ-PFD-005, REQ-PFD-006 | RB | 핸드오프 동결 단위 3개 바이트 불변, Localization 표 셀 패리티 | EL-2: 두 테스트 모두 부재 exit 1 | `go test ./internal/template/ -run '^TestOutputStyleHandoffUnitsFrozen$' -count=1 -v` → `--- PASS: TestOutputStyleHandoffUnitsFrozen ` (0): `moai.md` §6·§8, `moai-easy.md` Banner 7 해시가 앵커 고정물과 같음. `go test ./internal/template/ -run '^TestOutputStyleLocalizationTableParity$' -count=1 -v` → `--- PASS: TestOutputStyleLocalizationTableParity ` (0). 동결 절 안 한 글자 변이와 표 셀 하나 변이가 각각 `--- FAIL` 임을 관측해 기록 |
| AC-PFD-006 | REQ-PFD-007 | RG | Template-First: 로컬 사본 동일, 기존 출력 스타일 테스트 유지, 빌드 | EL-4: `diff -rq` 무출력 (0). EL-6: 5개 `--- PASS` | `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` → 출력 없음 (0). `go test ./internal/template/ -run '^TestOutputStylesEncoding$\|^TestOutputStylesFallbackDocsContract$\|^TestOutputStylesExactlyThree$\|^TestOutputStylesFrontmatterSchema$\|^TestOutputStylesTemplateLiveParity$' -count=1 -v` → 이름 5개 각각 `--- PASS: <이름> ` (0). `make build` → (0) |
| AC-PFD-007 | REQ-PFD-008 | RG | SPEC ID·카드 id·날짜·해시·언어명이 앵커보다 늘지 않음 | EL-9: SPEC/카드 15·0·0, 날짜 0·0·1, 16진 해시 0·0·0, 언어명 3·0·1 (`moai`·`moai-easy`·`moai-learn` 순) | 파일마다 다섯 식(식 단위 단일 호출) — `grep -cE 'SPEC-[A-Z]\|\bt1[0-9]{3}\b' <파일>`, `grep -cE '\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b' <파일>`, `grep -cE '\b[0-9a-f]{7,40}\b' <파일>`, `grep -cE '\b(Go\|Python\|Rust\|TypeScript\|JavaScript\|Java\|Kotlin\|Swift\|Ruby\|PHP\|Scala\|Elixir\|Dart)\b' <파일>` → 각 값이 EL-9 의 같은 칸 값 이하. 0건이면 grep 이 exit 1 을 내므로 exit 는 값 0 일 때 1, 그 외 0 으로 기록 |
| AC-PFD-008 | REQ-PFD-009 | RG | `skillListingBudgetFraction` `0.02` 불변, 설정 파일 무수정 | EL-5: 두 곳 `0.02`. EL-13: `git diff --numstat` 무출력 | `grep -n '"skillListingBudgetFraction"' internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → `…settings.json.tmpl:414:  "skillListingBudgetFraction": 0.02,` 과 `.claude/settings.json:410:  "skillListingBudgetFraction": 0.02,` (0). `git diff --numstat <BASE> HEAD -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → 출력 없음 (0) |
| AC-PFD-009 | REQ-PFD-010 | RB | 에이전트 설명 예산 테스트와 상수 | EL-2: 테스트 부재 exit 1 | `go test ./internal/template/ -run '^TestAgentDescriptionBudget$' -count=1 -v` → `--- PASS: TestAgentDescriptionBudget ` (0), 합계 ≤ 총합 상수 < 11,155, 모든 단일 ≤ 상한 상수 < 2,182. 상한을 넘긴 변이 설명 → `--- FAIL` 이고 메시지에 파일 이름과 크기 |
| AC-PFD-010 | REQ-PFD-011 | RG | 라우팅 의미 보존, 편집한 에이전트의 설명 로컬 사본 동일 | EL-7: `NOT for:` 개수 1×10·2(`manager-spec`)·3(`super-advisor`) | `grep -c 'NOT for:' internal/template/templates/.claude/agents/moai/*.md` → 파일별 개수가 EL-7 과 같거나 많음 (0). `go test ./internal/template/ -run '^TestAgentFrontmatterAudit$' -count=1 -v` → `--- PASS: TestAgentFrontmatterAudit ` (0). 편집한 각 에이전트마다 `plan.md` §C 의 추출 명령을 템플릿 파일과 로컬 파일에 돌려 `<scratch>/a.txt`·`<scratch>/b.txt` 로 리다이렉트한 뒤 `diff <scratch>/a.txt <scratch>/b.txt` → 출력 없음 (0) |
| AC-PFD-011 | REQ-PFD-012 | RB | 첫 턴 토큰 전/후 측정 기록, 앵커 재현 정지 조건, 비악화 | EL-8: `progress.md` 에 해당 줄 없음(`grep -c` 0, exit 1) | `grep -c 'first-turn-input-tokens\[' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `6` 이상 (0): `anchor#1..3`·`final#1..3` 필수, 파일을 바꾼 마일스톤은 `M<n>#1..3` 추가. `grep -c '"cache_creation_input_tokens"' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → 위 줄 수 이상 (0): 줄마다 원문 `usage` JSON 동반. `grep -c '^result: first-turn-nonregression=true' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `1` (0): `final` 중앙값 ≤ `anchor` 중앙값 |
| AC-PFD-012 | REQ-PFD-013 | RB | 경로 허용목록·내용 점검기가 최종 트리에서 PASS, 양성 대조에서 FAIL | EL-11: 앵커에서 docs 만 변경 → `surface-guard=PASS` exit 0, `docs` 금지 → exit 1, 에이전트 본문 변이 → exit 1 | `python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py <BASE>` → 변경 파일마다 `ok <표면> <경로>` 줄, 마지막 줄 `surface-guard=PASS` (0). 양성 대조(각각 FAIL·exit 1 을 관측해 기록하고 변이는 `git checkout --` 로 복원): `… <BASE> output-styles agents` (편집한 파일이 있으면 `VIOLATION forbidden-surface`), 템플릿 에이전트 본문에 한 줄 덧붙임 → `VIOLATION agent-body-or-nondescription-frontmatter-changed`, 출력 스타일 frontmatter 한 글자 변이 → `VIOLATION frontmatter-changed`, `settings.json.tmpl` 에 키 외 줄 변이 → `VIOLATION settings-line-other-than-skillListingBudgetFraction`, 허용목록 밖 파일(예: `.claude/rules/x.md` 새 파일) → `VIOLATION outside-allowlist` |
| AC-PFD-013 | REQ-PFD-013 | RG | 표면별 독립 명령: `outputStyle` 값, 설정 키, 스킬, 규칙, `CLAUDE.md`, `AGENTS*`, 훅 코드 | EL-13: `outputStyle` 두 줄, 나머지 모두 출력 없음 | `grep -n '"outputStyle"' internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → `…settings.json.tmpl:418:  "outputStyle": "MoAI-Easy",` 과 `.claude/settings.json:417:  "outputStyle": "MoAI-Easy",` (0). `git diff --numstat <BASE> HEAD -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → 출력 없음 (0). `git diff --name-only <BASE> HEAD -- .claude/skills internal/template/templates/.claude/skills .claude/rules internal/template/templates/.claude/rules CLAUDE.md internal/template/templates/CLAUDE.md AGENTS.md internal/template/templates/AGENTS.md.tmpl internal/hook` → 출력 없음 (0). `git merge-base --is-ancestor <BASE> HEAD` → 출력 없음 (0) |
| AC-PFD-014 | REQ-PFD-004, REQ-PFD-007, REQ-PFD-010 | RG | 패키지 단위 회귀 없음 | M0 가 앵커의 `FAIL` 이름 목록과 exit 를 기록(원장에 없음 — 실행 단계 몫) | `go test ./internal/template/ ./internal/config/ -count=1` (파일로 리다이렉트, 종료 코드와 `FAIL` 이름만 기록) → M0 앵커 기록과 `FAIL` 이름 집합이 같거나 더 작음 (앵커가 0 이면 0). 새 `FAIL` 이름 1개라도 있으면 회귀 |
| AC-PFD-015 | (plan 단계 산출물) | RG | SPEC lint 오류 0, `progress.md` §E 4절 | 이 plan 이 아래 §D.3 에 기록 | `moai spec lint .moai/specs/SPEC-PREFIX-DIET-001/spec.md` → `0 error(s)`, 경고 0 (0). `grep -c '^## §E\.' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `4` (0) |

## §D.1 증거 원장 (RED-now, 트리 핀 `5d5ff1aae`)

- **EL-1** — 파일 전체 UTF-16, 같은 식을 세 파일에 반복해 측정: `moai-easy.md 29243`, `moai-learn.md 28517`, `moai.md 62593`. (코드포인트 29181 / 28418 / 62470 — 단위가 다르다.)
- **EL-2** — `grep -rl <이름> internal/template` 을 `TestOutputStylesCharBudget`, `TestOutputStyleBindingLedger`, `TestOutputStyleHandoffUnitsFrozen`, `TestOutputStyleLocalizationTableParity`, `TestAgentDescriptionBudget` 에 각각 → stdout 없음, exit 1.
- **EL-3** — `ls internal/template/testdata/output_style_ledger.json` → stdout 없음, stderr `ls: internal/template/testdata/output_style_ledger.json: No such file or directory`, exit 1.
- **EL-4** — `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` → stdout 없음, exit 0.
- **EL-5** — `grep -n '"skillListingBudgetFraction"' internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → `internal/template/templates/.claude/settings.json.tmpl:414:  "skillListingBudgetFraction": 0.02,` 과 `.claude/settings.json:410:  "skillListingBudgetFraction": 0.02,`, exit 0.
- **EL-6** — `go test ./internal/template/ -run` 에 이름 접두 패턴을 줘서 관측한 결과: `--- PASS: TestOutputStylesEncoding `, `--- PASS: TestOutputStylesFallbackDocsContract `, `--- PASS: TestOutputStylesExactlyThree `, `--- PASS: TestOutputStylesFrontmatterSchema `, `--- PASS: TestOutputStylesTemplateLiveParity ` 5줄, `ok  github.com/modu-ai/moai-adk/internal/template 0.451s`. AC-PFD-006 의 분기별 앵커 패턴은 실행 단계가 재관측한다.
- **EL-7** — `grep -c 'NOT for:' internal/template/templates/.claude/agents/moai/*.md` → 파일별 `builder-harness 1 · e2e-tester 1 · manager-design 1 · manager-develop 1 · manager-docs 1 · manager-git 1 · manager-lead 1 · manager-spec 2 · manager-todo 1 · plan-auditor 1 · super-advisor 3 · sync-auditor 1`. 에이전트 `description:` 블록 UTF-16 합: 템플릿 11,155, 로컬 11,146.
- **EL-8** — `progress.md` 에는 `first-turn-input-tokens[` 줄이 없다(§E.2 는 플레이스홀더).
- **EL-9** — `moai.md` / `moai-easy.md` / `moai-learn.md`: `SPEC-[A-Z]|\bt1[0-9]{3}\b` 15 / 0 / 0, 날짜 `\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b` 0 / 0 / 1, 16진 해시 `\b[0-9a-f]{7,40}\b` 0 / 0 / 0, 언어명(13개 목록) 3 / 0 / 1.
- **EL-10** — 구속 토큰 개수(AC-PFD-004 명령): `moai.md` `HARD=89 MUST_NOT=4 MUST=27 shall=0`, `moai-easy.md` `HARD=33 MUST_NOT=0 MUST=0 shall=0`, `moai-learn.md` `HARD=24 MUST_NOT=0 MUST=7 shall=0`.
- **EL-11** — 점검기(`.moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py`) 관측: (a) `… 5d5ff1aae`, 변경 파일이 이 SPEC 디렉터리뿐 → 각 파일 `ok docs …`, `surface-guard=PASS`, exit 0. (b) `… 5d5ff1aae docs` → 각 파일 `VIOLATION forbidden-surface docs …`, `surface-guard=FAIL`, exit 1. (c) `internal/template/templates/.claude/agents/moai/manager-todo.md` 본문에 `mutant` 한 줄을 덧붙인 변이 → `VIOLATION agent-body-or-nondescription-frontmatter-changed agents internal/template/templates/.claude/agents/moai/manager-todo.md`, `surface-guard=FAIL`, exit 1 (관측 뒤 `git checkout --` 로 복원, `git status --short` 로 복원 확인).
- **EL-12** — AC-PFD-003 의 `python3 -c …` → stdout 없음, stderr 끝줄 `FileNotFoundError: [Errno 2] No such file or directory: 'internal/template/testdata/output_style_ledger.json'`, exit 1.
- **EL-13** — `grep -n '"outputStyle"' …` → `internal/template/templates/.claude/settings.json.tmpl:418:  "outputStyle": "MoAI-Easy",` 과 `.claude/settings.json:417:  "outputStyle": "MoAI-Easy",` exit 0. `git diff --numstat 5d5ff1aae HEAD -- <설정 두 파일>` → 출력 없음, exit 0. `git diff --name-only 5d5ff1aae HEAD -- <스킬·규칙·CLAUDE.md·AGENTS·훅 경로>` → 출력 없음, exit 0.

## §D.2 첫 턴 입력 토큰 측정 프로토콜 (전/후 동일 명령)

**준비**(스크래치 디렉터리 `<scratch>`, 머신 로컬):

```
<scratch>/hooks-off.json   내용: {"disableAllHooks": true}
```

**측정 명령** — 워크트리 루트에서, 조건마다 3회, 직렬로. 출력은 파일로:

```
claude -p ok --output-format json --model claude-opus-5-5 --settings <scratch>/hooks-off.json
```

**합계 정의** — 위 JSON 결과의 `usage` 안 세 필드의 합:

```
jq '.usage | .input_tokens + .cache_creation_input_tokens + .cache_read_input_tokens' <결과 파일>
```

**기록 형식**(`progress.md` §E.2, 한 실행당 한 줄 + 원문 `usage` JSON):

```
first-turn-input-tokens[anchor#1]=<합>   usage={"input_tokens":…,"cache_creation_input_tokens":…,"cache_read_input_tokens":…,…}
```

조건 이름: `anchor`(편집 전), `M2`·`M3`·`M4`·`M5`(파일을 바꾼 마일스톤), `final`. 각 줄에 측정 시점 HEAD SHA 를 같이 적는다.

규칙:

1. **앵커 재현 먼저(M0)**: 편집 전 트리에서 3회. 세 값 중 어느 것도 154,219 의 ±1% 안에 없으면 편집하지 않고 리더에게 세 값을 보고한다(REQ-PFD-012). 리더의 154,219 와 이 SPEC 의 합계 정의가 같은지는 이 재현으로 확인한다 — 같다는 가정을 하지 않는다.
2. **같은 명령, 같은 cwd, 같은 모델.** 스타일을 바꿔 보는 조건(`moai-learn.md`·`moai.md` 의 전/후)은 `--settings` 파일에 `outputStyle` 키를 추가해 만든다(`{"disableAllHooks":true,"outputStyle":"MoAI-Learn"}` 등). 스타일 이름은 해당 파일 frontmatter 의 `name:` 값을 읽어 쓴다. 이 조건들은 설정 파일을 수정하는 것이 아니라 측정 입력일 뿐이다.
3. **분산**: 조건당 3회의 최솟값·중앙값·최댓값을 기록한다. 마일스톤의 기대 감소폭이 3회 범위(최댓값−최솟값)보다 작으면 그 마일스톤의 토큰 효과는 "측정 불가"로 보고하고 크기 지표(UTF-16)만 주장한다.
4. **비악화(AC-PFD-011)**: `final` 중앙값 ≤ `anchor` 중앙값일 때만 `result: first-turn-nonregression=true` 를 적는다.
5. 측정 한계는 같은 절에 적는다: 훅을 끈 측정이라 SessionStart 훅 주입 맥락은 포함하지 않고, 계정·시각에 따른 cache 상태가 값에 영향을 줄 수 있다.

## §D.3 plan 단계 검증 기록

- `moai spec lint .moai/specs/SPEC-PREFIX-DIET-001/spec.md` — 결과는 **0 오류 + INFO 1건**(`OwnershipTransitionUnmeasured` — 커밋 `b30985c21` 에 `Authored-By-Agent` 트레일러 없음)이다. 0.1.0 판에서 "No findings" 라고 적은 앞선 기록은 부정확했고 plan-audit 1회차가 이를 정정했다. 이 INFO 는 커밋 메시지가 `🗿 MoAI` 줄로 끝나야 하므로(트레일러 블록은 마지막 단락이어야 인식된다) 의도적으로 남긴다. 이 0.2.0 판은 마지막 편집 뒤 재실행해 같은 형태(0 오류, INFO 1건, 경고 0)임을 확인한다 — 아래 커밋 직전 판독 참조. AC-PFD-015 의 기대는 "0 오류, 경고 0"(INFO 는 허용)으로 읽는다.
- SPEC ID 점검(`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`): `SPEC-PREFIX-DIET-001` → `PASS`. `.moai/specs/` 에 `SPEC-PREFIX-DIET-*` 는 이전에 없었다.

## §D.4 시나리오 (Given-When-Then)

- **AC-PFD-002** — Given 앵커 본문 단위 전부가 원장에 행으로 있고, When 한 `binding` 단위를 원장 행 없이 본문에서 지우면, Then 원장 테스트가 그 단위의 출처 제목을 대며 실패한다.
- **AC-PFD-003** — Given 원장에 `rewrite` 행이 0개이고, When 누군가 `[HARD]` 줄을 압축 재작성하며 `rewrite` 행을 추가하면, Then 원장 테스트가 그 행을 거부하며 실패한다.
- **AC-PFD-005** — Given `moai.md` §8 `Session Handoff [HARD]` 의 해시가 앵커 고정물과 같고, When 그 절 안 한 글자를 바꾸면, Then 동결 테스트가 해당 절 이름과 두 해시를 대며 실패한다.
- **AC-PFD-012** — Given 변경이 허용 표면 안에 있고 에이전트 본문·frontmatter 외 필드·설정 파일이 그대로일 때, When 점검기를 돌리면, Then `surface-guard=PASS` 이고, 에이전트 본문에 한 줄을 덧붙인 변이에서는 `surface-guard=FAIL` 과 exit 1 이 나온다.

## §D.5 Definition of Done

- AC-PFD-001 ~ AC-PFD-015 가 위 명령 그대로 재실행되어 축자 출력이 `.moai/reports/t1450/verdict.md` 에 반출되어 있다.
- 구속 원장 테스트가 초록이고, 원장 앞면(기준선)은 편집 커밋보다 앞선 별도 커밋에 있으며, 원장에 `rewrite` 행이 없다.
- `dropped` 행 전부의 목록(출처·이유·생존 단위)이 `progress.md` §E.2 에 반출되어 sync-audit 와 리더가 훑을 수 있다.
- 템플릿이 원본이고 `make build` 뒤 출력 스타일 로컬 사본이 템플릿과 같다.
- 착지는 리더의 일괄 push 배치 경계에서 이루어진다(prompt cache 무효화는 착지 직후 세션에서 한 번 발생한다는 점을 완료 보고에 적는다).
- 완료 보고는 Claim / Evidence / Baseline-attribution / Gaps / Residual-risk 5개 절을 쓰고, `spec.md` §H 의 Gap 중 이 실행이 닫은 것과 못 닫은 것을 나눠 적는다.
