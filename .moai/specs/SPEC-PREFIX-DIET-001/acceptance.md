---
id: SPEC-PREFIX-DIET-001
title: "acceptance — 세션 시작 prefix 다이어트 2단계"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
---

# acceptance.md — SPEC-PREFIX-DIET-001

## §D. AC 매트릭스

- 분류: **RB** = 릴리스 차단(새 의무 — RED-now 필수), **RG** = 회귀 가드(이미 성립하는 성질을 지킨다 — RED-now 를 요구하지 않음).
- **트리 핀**: 이 문서의 모든 RED-now 측정은 커밋 `5d5ff1aae`(워크트리 `t1450`, 브랜치 `WT-prefix-diet-stage2`)에서 이 plan 실행이 직접 돌린 것이다. 문서 수준 핀이며 개별 행에 핀이 없으면 이 핀이 묶는다.
- 명령은 적힌 그대로 워크트리 루트에서 실행하며 단일 호출이다. exit 코드는 별도 칸이다. `go test` 의 초록 판정은 언제나 `-v` 출력의 이름 붙은 `--- PASS: <이름> ` 줄이며, 패키지 요약 `ok` 줄은 테스트를 0개 골라도 찍히므로 근거가 되지 않는다. `-run` 은 `^…$` 로 앵커한다.
- 표 셀 안의 `\|` 는 마크다운 표 이스케이프이며, 실제 명령에서는 `|` 하나다.
- 출력 길이가 한계(50줄·2KB)를 넘는 명령은 파일로 리다이렉트하고 종료 코드와 꼬리만 기록한다.
- 증거 원문은 실행 단계에서 `.moai/reports/t1450/verdict.md` 와 `progress.md` §E.2 로 **반출**한다. `.moai/state/` 는 머신 로컬이라 인용 대상이 아니다.
- 이 plan 의 RED-now 는 증거 원장 EL-1 ~ EL-9 에 있고, 표의 RED-now 칸은 원장 ID 로 가리킨다.

| AC | REQ | 분류 | 기준 | RED-now (원장 ID) | 초록 조건 — 명령 → 기대 출력 (exit) |
|---|---|---|---|---|---|
| AC-PFD-001 | REQ-PFD-001, REQ-PFD-002 | RB | 출력 스타일 예산 테스트 존재, 파일별 상수 ≤ 앵커 크기 | EL-2: 테스트 부재, exit 1 | `grep -rl TestOutputStylesCharBudget internal/template` → 테스트 파일 경로 1줄 (0). `go test ./internal/template/ -run '^TestOutputStylesCharBudget$' -count=1 -v` → `--- PASS: TestOutputStylesCharBudget ` 과 `output-style=moai <N>`, `output-style=moai-easy <N>`, `output-style=moai-learn <N>` 3줄 (0) |
| AC-PFD-002 | REQ-PFD-002 | RB | 세 본문 모두 앵커보다 작고 M0 목표 이하 | EL-1: 29243 / 28517 / 62593 | 파일별 `python3 -c "import sys;t=open(sys.argv[1],encoding='utf-8').read();print(len(t.encode('utf-16-le'))//2)" <파일>` → 각각 EL-1 값보다 작고 `progress.md` §E.2 에 기록된 파일별 상수 이하 (0). 실행 단계가 세 값을 축자 기록 |
| AC-PFD-003 | REQ-PFD-003, REQ-PFD-004 | RB | 구속 원장 존재·완전 | EL-3: 고정물 부재, exit 1 | `go test ./internal/template/ -run '^TestOutputStyleBindingLedger$' -count=1 -v` → `--- PASS: TestOutputStyleBindingLedger ` (0). 앵커 단위가 행 없이 남은 픽스처·`binding`/`normative` 행이 `dropped` 인 픽스처·`dropped` 에 이유 또는 생존 단위 참조가 없는 픽스처 하위 테스트가 각각 `--- PASS` (실패를 기대하는 단언이 통과) |
| AC-PFD-004 | REQ-PFD-003, REQ-PFD-004 | RB | 변이 — 구속 줄 삭제와 의무 약화가 판정을 바꾼다 | EL-2 와 같은 경로(테스트 부재) | 같은 테스트의 변이 하위 테스트 `--- PASS`: (a) `binding` 단위의 `[HARD]` 토큰을 지운 변이 본문 → 원장 테스트 `--- FAIL`, (b) `MUST NOT` 을 `SHOULD NOT` 으로 바꾼 `rewrite` 행 → `--- FAIL`, (c) 단위를 `moai-easy.md` 에서 `.claude/skills/` 로 옮긴 변이 → `--- FAIL`. 변이 입력에서 FAIL 이 실제로 관측된 출력을 기록(`verification-completeness.md` §1.1) |
| AC-PFD-005 | REQ-PFD-005 | RB | 핸드오프 동결 단위 바이트 불변 | EL-2: 테스트 부재, exit 1 | `go test ./internal/template/ -run '^TestOutputStyleHandoffUnitsFrozen$' -count=1 -v` → `--- PASS: TestOutputStyleHandoffUnitsFrozen ` (0), 대상 3개(`moai.md` §6, `moai.md` §8, `moai-easy.md` Banner 7)의 해시가 앵커 고정물과 같음. 동결 단위 안 한 글자를 바꾼 변이 → `--- FAIL` |
| AC-PFD-006 | REQ-PFD-006 | RB | Localization 표 4개 로케일 셀 패리티 | EL-2: 테스트 부재, exit 1 | `go test ./internal/template/ -run '^TestOutputStyleLocalizationTableParity$' -count=1 -v` → `--- PASS: TestOutputStyleLocalizationTableParity ` (0): 세 파일의 표 행·4개 로케일 셀이 앵커 고정물과 같음. 셀 하나를 바꾼 변이 → `--- FAIL` |
| AC-PFD-007 | REQ-PFD-007 | RG | Template-First: 로컬 사본 동일, 기존 출력 스타일 테스트 유지 | EL-4: `diff -rq` 무출력 (0). EL-6: 5개 `--- PASS` | `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` → 출력 없음 (0). `go test ./internal/template/ -run '^TestOutputStylesEncoding$|^TestOutputStylesFallbackDocsContract$|^TestOutputStylesExactlyThree$|^TestOutputStylesFrontmatterSchema$|^TestOutputStylesTemplateLiveParity$' -count=1 -v` → `TestOutputStylesEncoding`, `TestOutputStylesFallbackDocsContract`, `TestOutputStylesExactlyThree`, `TestOutputStylesFrontmatterSchema`, `TestOutputStylesTemplateLiveParity` 각각 `--- PASS: <이름> ` (0). `make build` → (0) |
| AC-PFD-008 | REQ-PFD-008 | RG | 신규 SPEC ID·카드 id 미추가 | EL-9: `moai.md` 15, `moai-easy.md` 0, `moai-learn.md` 0 | `grep -cE 'SPEC-[A-Z]\|\bt1[0-9]{3}\b' internal/template/templates/.claude/output-styles/moai/moai.md` → 15 이하. 같은 식을 `moai-easy.md`·`moai-learn.md` 에 → `0` (exit 1 — 0건이면 grep 이 1 을 낸다) |
| AC-PFD-009 | REQ-PFD-009 | RB(값이 0.02 가 아닐 때) / RG(유지일 때) | 결정된 값이 원본·로컬 두 곳에 같다 | EL-5: 두 곳 `0.02` | `grep -n '"skillListingBudgetFraction"' internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → 두 줄 모두 `progress.md` 의 `skill-listing-decision=<D>` 값 (0). `python3 -c "import json;json.load(open('.claude/settings.json'))"` → 출력 없음 (0). 결정 줄이 없으면 값은 `0.02` 그대로여야 한다 |
| AC-PFD-010 | REQ-PFD-010 | RB(값을 낮췄을 때) | 스킬 이름 나열 수 전/후와 운영자 처분 기록 | EL-8: `progress.md` 에 해당 줄 없음(`grep -c` 0, exit 1) | `grep -c 'skill-listing-count\[' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `6` 이상 (0): `before#1..3`, `after#1..3` 줄. `grep -c '^operator-disposition:' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `1` (0). 값을 낮추지 않았으면 `skill-listing-decision=unchanged` 줄 1개가 이 AC 를 대신한다 |
| AC-PFD-011 | REQ-PFD-011 | RB | 에이전트 설명 예산 테스트와 상수 | EL-2: 테스트 부재, exit 1 | `go test ./internal/template/ -run '^TestAgentDescriptionBudget$' -count=1 -v` → `--- PASS: TestAgentDescriptionBudget ` (0), 합계 ≤ 총합 상수 < 앵커 11,155, 모든 단일 ≤ 상한 상수. 상한을 넘긴 변이 설명 → `--- FAIL` 이고 메시지에 파일 이름과 크기 |
| AC-PFD-012 | REQ-PFD-012 | RG | 라우팅 의미 보존, 편집한 에이전트의 설명 로컬 사본 동일 | EL-7: `NOT for:` 개수 1×10·2(`manager-spec`)·3(`super-advisor`) | `grep -c 'NOT for:' internal/template/templates/.claude/agents/moai/*.md` → 파일별 개수가 EL-7 과 같거나 많음 (0). `go test ./internal/template/ -run '^TestAgentFrontmatterAudit$' -count=1 -v` → `--- PASS: TestAgentFrontmatterAudit ` (0). 편집한 각 에이전트의 `description:` 블록이 로컬 사본과 같음 — `plan.md` §C 의 추출 명령을 템플릿 파일과 로컬 파일에 각각 돌려 `<scratch>/a.txt`·`<scratch>/b.txt` 로 리다이렉트한 뒤 `diff <scratch>/a.txt <scratch>/b.txt` → 출력 없음 (0) |
| AC-PFD-013 | REQ-PFD-013, REQ-PFD-014 | RB | 첫 턴 토큰 전/후 측정 기록과 비악화 | EL-8: `progress.md` 에 해당 줄 없음 | `grep -c 'first-turn-input-tokens\[' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `6` 이상 (0): `anchor#1..3` 과 `final#1..3` 필수, 파일을 바꾼 마일스톤은 `M<n>#1..3` 추가. `grep -c '"cache_creation_input_tokens"' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → 위 줄 수 이상 (0): 줄마다 원문 `usage` JSON 동반. `grep -c '^result: first-turn-nonregression=true' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `1` (0): `final` 중앙값 ≤ `anchor` 중앙값 |
| AC-PFD-014 | REQ-PFD-015 | RG | 제외 경로 무변경 | EL-9b: 앵커 대비 빈 출력 (0) | `git diff --name-only 5d5ff1aae HEAD -- .claude/rules internal/template/templates/.claude/rules CLAUDE.md internal/template/templates/CLAUDE.md AGENTS.md internal/template/templates/AGENTS.md.tmpl` → 출력 없음 (0). 범위 판정은 병합 전 평가 전용이며, 이 명령은 `5d5ff1aae` 가 조상일 때만 의미가 있다 — 실행 단계가 `git merge-base --is-ancestor 5d5ff1aae HEAD` (exit 0)를 함께 기록한다 |
| AC-PFD-015 | REQ-PFD-004, REQ-PFD-007, REQ-PFD-011 | RG | 패키지 단위 회귀 없음 | M0 가 앵커의 `FAIL` 이름 목록과 exit 를 기록(EL 에 없음 — 실행 단계 몫) | `go test ./internal/template/ ./internal/config/ -count=1` (파일로 리다이렉트, 종료 코드와 `FAIL` 이름만 기록) → M0 앵커 기록과 `FAIL` 이름 집합이 같거나 더 작음 (앵커가 0 이면 0). 새 `FAIL` 이름 1개라도 있으면 회귀 |
| AC-PFD-016 | (plan 단계 산출물) | RG | SPEC lint 오류 0, `progress.md` §E 4절 | 이 plan 이 아래 §D.3 에 기록 | `moai spec lint .moai/specs/SPEC-PREFIX-DIET-001/spec.md` → `0 error(s)` (0). `grep -c '^## §E\.' .moai/specs/SPEC-PREFIX-DIET-001/progress.md` → `4` (0) |

## §D.1 증거 원장 (RED-now, 트리 핀 `5d5ff1aae`)

- **EL-1** — 크기(UTF-16), 같은 식을 세 파일에 반복한 스크립트로 측정:
  `moai-easy.md 29243`, `moai-learn.md 28517`, `moai.md 62593`. (코드포인트: 29181 / 28418 / 62470 — 리더 인용값과 같다. 단위가 다르다.)
- **EL-2** — `grep -rl TestOutputStylesCharBudget internal/template` 등 5개 테스트 이름(`TestOutputStylesCharBudget`, `TestOutputStyleBindingLedger`, `TestOutputStyleHandoffUnitsFrozen`, `TestOutputStyleLocalizationTableParity`, `TestAgentDescriptionBudget`) 각각 stdout 없음, exit 1.
- **EL-3** — `ls internal/template/testdata/output_style_ledger.json` → stdout 없음, stderr `ls: internal/template/testdata/output_style_ledger.json: No such file or directory`, exit 1.
- **EL-4** — `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` → stdout 없음, exit 0.
- **EL-5** — `grep -n '"skillListingBudgetFraction"' internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` →
  `internal/template/templates/.claude/settings.json.tmpl:414:  "skillListingBudgetFraction": 0.02,` 과 `.claude/settings.json:410:  "skillListingBudgetFraction": 0.02,`, exit 0.
- **EL-6** — `go test ./internal/template/ -run '^TestOutputStylesEncoding$|^TestOutputStylesFallbackDocsContract$|^TestOutputStylesExactlyThree$|^TestOutputStylesFrontmatterSchema$|^TestOutputStylesTemplateLiveParity$' -count=1 -v` → `--- PASS: TestOutputStylesEncoding `, `--- PASS: TestOutputStylesFallbackDocsContract `, `--- PASS: TestOutputStylesExactlyThree `, `--- PASS: TestOutputStylesFrontmatterSchema `, `--- PASS: TestOutputStylesTemplateLiveParity ` 5줄(이 plan 은 앵커 패턴 `^TestOutputStyles` 로 같은 5개를 관측했고 실행 단계가 위 분기별 패턴으로 재관측한다), `ok  github.com/modu-ai/moai-adk/internal/template 0.451s`.
- **EL-7** — `grep -c 'NOT for:' internal/template/templates/.claude/agents/moai/*.md` → 파일별 `builder-harness 1 · e2e-tester 1 · manager-design 1 · manager-develop 1 · manager-docs 1 · manager-git 1 · manager-lead 1 · manager-spec 2 · manager-todo 1 · plan-auditor 1 · super-advisor 3 · sync-auditor 1`(출력 순서는 달랐다). 에이전트 `description:` 블록 UTF-16 합: 템플릿 11,155, 로컬 11,146.
- **EL-8** — `progress.md` 에는 `skill-listing-count[`·`first-turn-input-tokens[` 줄이 없다(이 plan 은 §E.2 를 플레이스홀더로만 둔다).
- **EL-9** — `grep -cE 'SPEC-[A-Z]|\bt1[0-9]{3}\b' <파일>` → `moai.md 15`, `moai-easy.md 0`, `moai-learn.md 0`.
- **EL-9b** — `git diff --name-only 5d5ff1aae HEAD -- <제외 경로 7개>` → stdout 없음, exit 0(plan 실행 시점 HEAD 는 `5d5ff1aae` 자체).

## §D.2 첫 턴 입력 토큰·스킬 발견 측정 프로토콜 (전/후 동일 명령)

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

조건 이름: `anchor`(편집 전), `M1`·`M2`·`M3`·`M4`·`M5`(파일을 바꾼 마일스톤), `final`. 각 줄에 측정 시점 HEAD SHA 를 같이 적는다.

규칙:

1. **앵커 재현 먼저(M0)**: 편집 전 트리에서 3회. 세 값 중 어느 것도 154,219 의 ±1% 안에 없으면 편집하지 않고 리더에게 세 값을 보고한다(REQ-PFD-014). 리더의 154,219 와 이 SPEC 의 합계 정의가 같은지는 이 재현으로 확인한다 — 같다는 가정을 하지 않는다.
2. **같은 명령, 같은 cwd, 같은 모델.** 스타일·설정을 바꿔 보는 조건은 `--settings` 파일에 키를 추가해(`{"disableAllHooks":true,"outputStyle":"MoAI-Learn"}` 등) 만든다. 스타일 이름은 해당 `output-styles/moai/*.md` 의 frontmatter `name:` 값을 읽어 쓴다.
3. **분산**: 조건당 3회의 최솟값·중앙값·최댓값을 기록한다. 마일스톤의 기대 감소폭이 3회 범위(최댓값−최솟값)보다 작으면 그 마일스톤의 토큰 효과는 "측정 불가"로 보고하고 크기 지표(UTF-16)만 주장한다.
4. **비악화(AC-PFD-013)**: `final` 중앙값 ≤ `anchor` 중앙값일 때만 `result: first-turn-nonregression=true` 를 적는다.
5. 측정 한계는 같은 절에 적는다: 훅을 끈 측정이라 SessionStart 훅 주입 맥락은 포함하지 않고, 계정·시각에 따른 cache 상태가 값에 영향을 줄 수 있다.

**스킬 발견 프로브**(REQ-PFD-010, 값을 낮춘 경우만) — 같은 `--settings` 파일로:

```
claude -p "List every skill name available to you in your skill listing, one per line, no commentary." --output-format json --model claude-opus-5-5 --settings <scratch>/hooks-off.json
```

`jq -r .result <결과 파일>` 의 비어 있지 않은 줄 수를 `skill-listing-count[before#n]`(0.02 상태)와 `skill-listing-count[after#n]`(후보값 상태)로 3회씩 기록한다. 이 대리 지표는 이름 누락만 잡고, 설명이 잘리는 품질 저하나 모델이 목록 대신 기억으로 답하는 오차는 잡지 못한다 — 같은 줄 옆에 한계를 적고, 운영자 처분(`operator-disposition:`)은 이 숫자와 한계를 함께 본 뒤의 판단이다.

## §D.3 plan 단계 검증 기록

- `moai spec lint .moai/specs/SPEC-PREFIX-DIET-001/spec.md` → `✓ No findings — all SPEC documents are valid` (exit 0). 이 plan 의 마지막 편집 뒤에 판독했다(최초 판독은 acceptance.md 부재·AC 연결 없음·앵커 없는 `-run` 패턴으로 경고 16+3건이었고 수리 뒤 0건).
- SPEC ID 점검(`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`): `SPEC-PREFIX-DIET-001` → `PASS`. `.moai/specs/` 에 `SPEC-PREFIX-DIET-*` 는 이전에 없었다(`ls | grep -E 'PREFIX|DIET'` 결과에 해당 이름 없음).

## §D.4 시나리오 (Given-When-Then)

- **AC-PFD-003** — Given 앵커 본문 단위 전부가 원장에 행으로 있고, When 한 `binding` 단위를 원장 행 없이 본문에서 지우면, Then 원장 테스트가 그 단위의 출처 제목을 대며 실패한다.
- **AC-PFD-005** — Given `moai.md` §8 `Session Handoff [HARD]` 의 해시가 앵커 고정물과 같고, When 그 절 안 한 글자를 바꾸면, Then 동결 테스트가 해당 절 이름과 두 해시를 대며 실패한다.
- **AC-PFD-009** — Given 운영자가 0.01 로 판정했고, When `settings.json.tmpl` 과 `.claude/settings.json` 의 키를 읽으면, Then 두 곳 모두 `0.01` 이고 `.claude/settings.json` 은 유효한 JSON 이다.
- **AC-PFD-013** — Given 앵커 3회 측정이 154,219 의 ±1% 안에 하나라도 있고, When 모든 마일스톤 뒤 `final` 3회를 재면, Then 중앙값이 `anchor` 중앙값 이하이고 각 실행의 원문 `usage` JSON 이 기록에 있다.

## §D.5 Definition of Done

- AC-PFD-001 ~ AC-PFD-016 이 위 명령 그대로 재실행되어 축자 출력이 `.moai/reports/t1450/verdict.md` 에 반출되어 있다.
- 구속 원장 테스트가 초록이고, 원장 앞면은 편집 커밋보다 앞선 별도 커밋에 있다.
- 템플릿이 원본이고 `make build` 뒤 출력 스타일 로컬 사본이 템플릿과 같다.
- 착지는 리더의 일괄 push 배치 경계에서 이루어진다(prompt cache 무효화는 착지 직후 세션에서 한 번 발생한다는 점을 완료 보고에 적는다).
- 완료 보고는 Claim / Evidence / Baseline-attribution / Gaps / Residual-risk 5개 절을 쓰고, `spec.md` §H 의 Gap 중 이 실행이 닫은 것과 못 닫은 것을 나눠 적는다.
