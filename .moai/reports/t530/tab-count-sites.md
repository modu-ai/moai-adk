# t530 — 설정 탭 수가 손으로 적혀 있는 자리 전수 목록

카드 t530 의 plan-phase 산출물. 이 파일은 **열거 그 자체가 산출물**이라는 카드 요구를 충족하기 위한 것이며,
run-phase 는 이 목록을 체크리스트로 소비한다. 목록이 불완전하면 수정도 불완전하다.

- 측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, base `1d150a27d`
- 측정 일자: 2026-09-12
- SPEC: `SPEC-DOCS-TABCOUNT-DRIFT-001`

---

## 1. 기준값 — 이 트리에서 다시 잰 탭 수

**14.** 카드가 들고 온 숫자를 그대로 옮긴 것이 아니라 이 트리에서 다시 쟀다.
근거의 성격은 **원천 1(구현) + 미러 1(테스트 리터럴) + 둘의 일치를 확인하는 테스트 1** 이다 —
**독립인 세 측정이 아니다.** 측정 3은 1과 2를 비교하는 테스트 그 자체이고, 측정 2의 `wantTabOrder` 는
`consoleTabs()` 를 미러하도록 손으로 적은 리터럴이라 독립성이 약하다. 근거로 충분하되 독립성을 부풀려 적지 않는다 —
같은 사실을 여러 곳에서 따로 센 것처럼 보이게 하는 과장이 바로 이 카드가 다루는 결함이기 때문이다
(`spec.md §1.1` 과 같은 문장, 같은 뜻).

| # | 성격 | 명령 | 관측된 출력 |
|---|---|---|---|
| 1 | 원천(구현) | `sed -n '30,90p' internal/web/schemaform.go \| grep -c 'LabelKey:'` | `14` |
| 2 | 미러(테스트 리터럴) | `grep -n -A 20 'wantTabOrder' internal/web/tab_layout_test.go` | 리터럴에 id 14개: identity, language, launch, llm, workflow, git-worktree, audit, codex, agentfm, report, mcp, crosssession, feedback, gate |
| 3 | 1과 2의 일치 확인 | `go test ./internal/web/ -run 'TestConsoleTabsOrder'` | `ok github.com/modu-ai/moai-adk/internal/web 0.794s` |

`consoleTabs()` 는 `internal/web/schemaform.go:34` 에서 시작하고 탭 리터럴은 36–84 행에 걸쳐 있다.

**14 는 상수가 아니라 이 base 에서의 실측치다.** 탭을 추가하는 다른 카드가 이 값을 바꾼다.
그래서 본 카드의 해법은 "14 로 고치기" 가 아니라 "손으로 적힌 수를 없애고, 남는 곳은 기계가 지키게 하기" 다.

## 2. 탭 이름 정본 (콘솔이 실제로 렌더하는 라벨)

`internal/web/assets/i18n.js` 의 `LabelKey` 값이 정본이고, `schemaform.go` 의 `Baseline` 은 그 폴백이다.

| 순서 | id | en 라벨 |
|---|---|---|
| 1 | identity | Identity |
| 2 | language | Language |
| 3 | launch | LLM |
| 4 | llm | **GLM Settings** |
| 5 | workflow | Workflow |
| 6 | git-worktree | Git & Worktree |
| 7 | audit | Audit |
| 8 | codex | Codex |
| 9 | agentfm | Agents |
| 10 | report | Report |
| 11 | mcp | MCP |
| 12 | crosssession | Cross-Session |
| 13 | feedback | Feedback |
| 14 | gate | Quality Gate |

---

## 3. A군 — 수가 틀린 자리 (9 라고 적혀 있음): 4자리

전부 `docs-site/content/<locale>/cli-reference/web.md:53`, `/settings` 엔드포인트 표의 한 행이다.

| # | 파일 | 행 | 로케일 | 적힌 형태 | 표기 | 제거 가능? |
|---|---|---|---|---|---|---|
| A1 | `docs-site/content/ko/cli-reference/web.md` | 53 | ko | `설정 9개 탭` | 숫자 | 예 |
| A2 | `docs-site/content/en/cli-reference/web.md` | 53 | en | `The nine settings tabs` | **낱말** | 예 |
| A3 | `docs-site/content/ja/cli-reference/web.md` | 53 | ja | `設定 9 タブ` | 숫자 | 예 |
| A4 | `docs-site/content/zh/cli-reference/web.md` | 53 | zh | `设置九个标签页` | **한자 수사** | 예 |

## 4. B군 — 수는 맞지만(14) 손으로 적힌 자리: docs-site 8자리

전부 `docs-site/content/<locale>/advanced/moai-web-console.md`. 로케일마다 두 곳이다.

| # | 파일 | 행 | 로케일 | 적힌 형태 | 표기 | 제거 가능? |
|---|---|---|---|---|---|---|
| B1 | `docs-site/content/ko/advanced/moai-web-console.md` | 25 | ko | `(14개 탭)` | 숫자 | 예 |
| B2 | `docs-site/content/ko/advanced/moai-web-console.md` | 128 | ko | `14개 탭이 세로 목록으로` | 숫자 | 예 — 바로 아래 이름 목록이 있음 |
| B3 | `docs-site/content/en/advanced/moai-web-console.md` | 25 | en | `(14 tabs)` | 숫자 | 예 |
| B4 | `docs-site/content/en/advanced/moai-web-console.md` | 128 | en | `unfolds fourteen tabs` | **낱말** | 예 — 바로 아래 이름 목록이 있음 |
| B5 | `docs-site/content/ja/advanced/moai-web-console.md` | 25 | ja | `（14 タブ）` | 숫자 | 예 |
| B6 | `docs-site/content/ja/advanced/moai-web-console.md` | 128 | ja | `14 のタブが` | 숫자 | 예 — 바로 아래 이름 목록이 있음 |
| B7 | `docs-site/content/zh/advanced/moai-web-console.md` | 25 | zh | `（14 个标签页）` | 숫자 | 예 |
| B8 | `docs-site/content/zh/advanced/moai-web-console.md` | 128 | zh | `14 个标签页的纵向列表` | 숫자 | 예 — 바로 아래 이름 목록이 있음 |

## 5. C군 — 수는 맞지만(14) 손으로 적힌 자리: README 8자리

| # | 파일 | 행 | 로케일 | 적힌 형태 | 표기 | 제거 가능? |
|---|---|---|---|---|---|---|
| C1 | `README.md` | 414 | en | `splits into fourteen tabs: Identity, …` | **낱말** + 이름 14개 목록 | 예 — 목록이 이미 뒤따름 |
| C2 | `README.ko.md` | 414 | ko | `… 열네 개 탭으로 나뉜다` | **한국어 수사** + 이름 목록 | 예 |
| C3 | `README.ja.md` | 414 | ja | `… の 14 タブに分かれる` | 숫자 + 이름 목록 | 예 |
| C4 | `README.zh.md` | 414 | zh | `分成十四个标签页：…` | **한자 수사** + 이름 목록 | 예 |
| C5 | `README.md` | 750 | en | `` `moai web` `` 표 행 — `14-tab settings` | 숫자 | 예 |
| C6 | `README.ko.md` | 750 | ko | 표 행 — `14-탭 설정` | 숫자 | 예 |
| C7 | `README.ja.md` | 750 | ja | 표 행 — `14 タブ設定` | 숫자 | 예 |
| C8 | `README.zh.md` | 750 | zh | 표 행 — `14 标签页设置` | 숫자 | 예 |

**총 20자리 / 12파일.** A군 4 + B군 8 + C군 8.

## 6. D군 — 같은 결함의 두 번째 사례: 탭 **이름** 드리프트 (12자리)

수와 별개로, 탭 이름 목록도 손으로 적혀 있고 이미 한 칸 어긋나 있다.
4번 탭의 실제 렌더 라벨은 `GLM Settings` / `GLM 설정` / `GLM設定` / `GLM设置`
(`internal/web/assets/i18n.js:229,1096,1852,2608`) 인데, 문서 8자리는 전부 `3rd Party LLM` 계열로 적혀 있다.

| # | 파일 | 행 | 적힌 이름 |
|---|---|---|---|
| D1 | `README.md` | 414 | `3rd Party LLM` |
| D2 | `README.ko.md` | 414 | `3rd Party LLM` |
| D3 | `README.ja.md` | 414 | `3rd Party LLM` |
| D4 | `README.zh.md` | 414 | `3rd Party LLM` |
| D5 | `docs-site/content/ko/advanced/moai-web-console.md` | 133 | `서드파티 LLM(3rd Party LLM)` |
| D6 | `docs-site/content/en/advanced/moai-web-console.md` | 133 | `3rd Party LLM` |
| D7 | `docs-site/content/ja/advanced/moai-web-console.md` | 133 | `サードパーティ LLM（3rd Party LLM）` |
| D8 | `docs-site/content/zh/advanced/moai-web-console.md` | 133 | `第三方 LLM（3rd Party LLM）` |
| D9 | `docs-site/content/ko/advanced/moai-web-console.md` | 165 | `서드파티 LLM 탭` (번호 목록 밖 산문) |
| D10 | `docs-site/content/en/advanced/moai-web-console.md` | 165 | `The 3rd Party LLM tab` (번호 목록 밖 산문) |
| D11 | `docs-site/content/ja/advanced/moai-web-console.md` | 165 | `サードパーティ LLM タブ` (번호 목록 밖 산문) |
| D12 | `docs-site/content/zh/advanced/moai-web-console.md` | 165 | `第三方 LLM 标签页` (번호 목록 밖 산문) |

D9~D12 는 번호 목록이 아니라 산문이라 **기계 가드(N2)가 보지 않는다** — 같은 변경에서 사람이 함께 고친다.
표 행으로 올려 둔 이유가 그것이다: 행이 아니면 열거 계수에 들지 않아, 문단을 통째로 지워도 완전성 검사가 통과한다.
실측(base `1d150a27d`): `grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM'` 대상 8파일 → **12행**
= D1~D12 와 같은 수.

나머지 13개 이름은 렌더 라벨과 일치한다. 어느 쪽을 고쳐야 하는지(문서를 코드에 맞출 것인가,
코드 라벨 자체가 틀렸는가)는 plan.md §C 에서 갈래 1(문서를 코드에 맞춘다)로 결정되었다 (2026-09-12).

## 7. 오탐 — 가드가 반드시 걸러야 할 자리

두 부류로 나뉜다. **7-A 는 대상 파일 밖**이라 파일 목록으로 걸러지고, **7-B 는 대상 파일 안**이라 목록으로도
수-명사 인접으로도 걸러지지 않는다 — 명시 허용 목록이 있어야 한다.

### 7-B. 화이트리스트 **안**의 정당한 계수 — 허용 목록 항목 (1항목, 4로케일)

| 파일 | 행 | 내용 | 왜 고칠 수 없나 |
|---|---|---|---|
| `docs-site/content/ja/advanced/moai-web-console.md` | 149 | `…2 つのタブを行き来し…` | codex 패널의 **존재 이유**를 설명하는 문장. 세는 대상이 설정 탭 총수가 아니라 **감사 탭과 MCP 탭 둘**이다. codex 패널 반경은 `spec.md §4`·§6 이 범위 밖으로 선언했다 |
| `docs-site/content/ko/advanced/moai-web-console.md` | 149 | `…탭 두 곳을 오가며…` | 위와 같음. 현재 숫자 스윕에는 안 걸리지만(수사가 낱말) 같은 문단이다 |
| `docs-site/content/en/advanced/moai-web-console.md` | 149 | `…visiting two tabs…` | 위와 같음 |
| `docs-site/content/zh/advanced/moai-web-console.md` | 149 | `…在两个标签页之间来回翻…` | 위와 같음 |

**식별 방법은 줄 번호가 아니라 내용이다** — 편집이 줄 번호를 움직인다. 가드는
`advanced/moai-web-console.md` 안에서 `codex` 토큰을 포함하는 줄을 허용한다.
실측(base `1d150a27d`): 낱말 경계를 붙인 숫자 스윕 15행 중 이 1행(ja:149)만 열거 밖이다.

[HARD] **허용 규칙은 반드시 파일 범위를 가진다 — `advanced/moai-web-console.md` 안에서만 적용한다.**
파일 범위 없이 "`codex` 를 포함하는 줄" 로 쓰면 **README 4본의 `:414` 가 함께 면제된다** — 그 줄이 탭 이름 목록에
`Codex` 를 담고 뒤 문장에 소문자 `codex` 도 쓰기 때문이다. 즉 열거된 C1~C4 자리가 가드에서 조용히 사라진다.
수리 중 실측으로 확인했다: 파일 범위 없이 적용하면 낱말 축 적중 6행 중 README 3행이 사라져 3행만 남고,
파일 범위를 붙이면 6행이 그대로 남는다. 규칙 하나의 범위가 이 카드의 절반을 무력화할 수 있다는 뜻이다.

**허용 규칙은 1개지만, 그 규칙이 면제하는 줄 수는 그보다 많다.** 실측(base `1d150a27d`):
`grep -ci codex` → 4로케일 각 **4행**, 합계 **16행**. 규칙 수(1)와 면제 표면(16)은 다른 값이며,
가드는 둘 다 보고한다(`acceptance.md` AC-TCD-007).

### 7-C. 낱말 축을 열었을 때의 오탐 — 서수 접두 제외로 닫힌다

낱말 수사 축(`nine`/`fourteen`/`열네`/`十四`/`九个` …)을 가드에 넣으면 아래 한 부류가 추가로 걸린다.

| 파일 | 행 | 내용 | 왜 오탐인가 | 어떻게 닫나 |
|---|---|---|---|---|
| `docs-site/content/zh/advanced/moai-web-console.md` | 165 | `第**三**方 LLM 标签页…` | `第三方`(서드파티)의 `三` 가 수사로 읽힌다. 세는 말이 아니라 **서수 접두어**다 | 수사 바로 앞이 `第` 이면 매치에서 제외 (실측: 제외 후 이 줄 적중 `0`) |

이 한 부류를 닫으면 **허용되지 않은 오탐은 0** 이다 — §7-B 의 `codex` 허용 규칙을 적용한 상태에서 잰 값이다(아래 재현 명령 5).
낱말 축을 닫아 두어야 할 실측 근거는 **없다**(이전 판이 근거로 들었던 세 자리는 전부 §7-B 의 `codex` 허용 규칙이 이미 덮는다).

### 7-A. 장치로 걸러지는 오탐 — 파일 목록 / 낱말 경계 / 수-명사 인접

`README.md` 는 대상 12파일 **안**이다. 아래 README 행들이 걸러지는 이유는 파일 목록이 아니라
**낱말 경계**와 **수-명사 인접** 조건이다. 표의 마지막 열이 어느 장치가 닫는지를 밝힌다.

| 파일 | 행 | 내용 | 왜 오탐인가 | 닫는 장치 |
|---|---|---|---|---|
| `README.md` | 698 | `glm-5.3 stays selec**tab**le…` | 낱말 안의 `tab` 부분 문자열 | **낱말 경계** `\btabs?\b` (실측: 경계 없이 적중, 붙이면 무적중) |
| `README.md` | 422 | `measured nine forms … all nine proved reproducible` | SVG 인포그래픽 형태 수. 탭과 무관 | **수-명사 인접** (탭 명사가 없다) |
| `README.ko.md` | 422 | `… 아홉 가지 형태를` | 위와 같음 | 위와 같음 |
| `README.md` | 418 | `Eleven ref skills` | 스킬 수 | 위와 같음 |
| `README.md` | 486 | `Eleven of the twelve are agents` | 에이전트 수 | 위와 같음 |
| `docs-site/content/ja/claude-code/extensibility/plugins.md` | 100 | `4 タブを持つプラグインマネージャー` | **Claude Code 플러그인 매니저**의 탭. 수+탭 명사가 붙어 있어 다른 장치로는 못 막는다 | **파일 목록** (대상 12파일 밖) |
| `docs-site/content/*/core-concepts/kanban-board-terms.md` | 9, 29 | `nine terms` / `아홉 낱말` / `九个词` | 칸반 용어 수 | **파일 목록** |
| `docs-site/content/*/advanced/statusline.md` | 203 | `열여섯 개 키` / `十六个键` | statusline 세그먼트 수 | **파일 목록** |
| `docs-site/content/*/advanced/moai-web-console.md` | 137 | `codex 설정 12개` / `twelve scattered codex settings` | codex 패널 내부 설정 수 — t509 반경. 본 카드 범위 밖 | **허용 목록**(`codex` 토큰) + 수-명사 인접 |

## 8. 스크린샷 — 별건으로 분리

`assets/images/moai-web-settings.png` 는 README 4본(각 411행)에서 참조되며, 11탭 시절 이미지다.
t509 가 alt 텍스트에서 수를 없앴을 뿐 이미지는 그대로다. 재생성 절차는 저장소 어디에도 기록되어 있지 않다.
본 카드에 접지 않는 근거와 후속 카드 제안은 `spec.md` §5 에 있다.

---

## 재현 명령

모든 수치는 `git merge-base develop HEAD` = `1d150a27d` 에서 잰 것이다(측정 시점의 값 — 이 값은 develop 을 흡수하면
전진하므로 리터럴로 고정하지 않고 읽는 시점에 다시 구한다). 편집이 행 번호를 움직이므로, 위 표의 `행` 열은
**그 base 에서의 위치**이며 그 뒤의 트리에서는 다시 재야 한다. 파일 경로는 움직이지 않으므로
완전성 검사(AC-TCD-011)는 A/B/C/D 표 행에서 뽑은 경로만 본다.

```bash
# 1) 기준값 — 원천 1 + 미러 1 + 둘의 일치 테스트 1 (독립인 세 측정이 아니다)
sed -n '30,90p' internal/web/schemaform.go | grep -c 'LabelKey:'
grep -n -A 6 'var wantTabOrder' internal/web/tab_layout_test.go
go test ./internal/web/ -run 'TestConsoleTabsOrder'

# 2) A~C군 20자리 — 열거에서 뽑은 정확 리터럴 집합으로 센다 (창 너비 문제 없음, 오탐 없음)
grep -rnF -f .moai/reports/t530/count-literals.txt \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | wc -l
# base 실측: 20 (= 위 A/B/C 표 20행과 집합 동일)

# 3) 가드 층 숫자 스윕 — 낱말 경계 필수. base 실측 15행 = 열거 14 + 허용 1(ja:149)
grep -rnE '[0-9]+[^.。]{0,12}(\btabs?\b|탭|タブ|标签页)' \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md

# 4) D군 12자리
grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM' \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | wc -l
# base 실측: 12

# 5) 낱말 축 — 서수 접두(第) 제외 + codex 허용 규칙(파일 범위 필수) 적용 후 남는 적중
#    base 실측: 6행 = A2, A4, B4, C1, C2, C4 (열거된 낱말 표기 6자리와 정확히 일치)
#    허용되지 않은 오탐: 0
grep -rniE '(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|thirteen|fourteen|fifteen|sixteen|seventeen|eighteen|nineteen|twenty|하나|둘|셋|넷|다섯|여섯|일곱|여덟|아홉|열|열한|열두|열세|열네|열다섯|열여섯|한|두|세|네|[一二三四五六七八九十])[^.。]{0,12}(\btabs?\b|탭|タブ|标签页)' \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md \
  | sed 's/第[一二三四五六七八九十]/第X/g' \
  | grep -iE '(nine|fourteen|열네|十四|九|[一二三四五六七八九十]|one|two|three|four|five|six|seven|eight|ten|eleven|twelve|thirteen|fifteen|sixteen|하나|둘|셋|넷|다섯|여섯|일곱|여덟|아홉|열|한|두|세|네)[^.。]{0,12}(\btabs?\b|탭|タブ|标签页)' \
  | grep -vE '^docs-site/content/[a-z]+/advanced/moai-web-console\.md:[0-9]+:.*codex'
```

위 5번의 `grep -vE` 가 **파일 경로까지 포함한** 패턴인 점이 핵심이다. `grep -v 'codex'` 로 줄이면
README 3행이 함께 사라져 6행이 3행으로 줄어든다(실측). 허용 규칙에서 파일 범위를 빼면 안 되는 이유다.
