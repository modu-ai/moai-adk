# t530 — 설정 탭 수가 손으로 적혀 있는 자리 전수 목록

카드 t530 의 plan-phase 산출물. 이 파일은 **열거 그 자체가 산출물**이라는 카드 요구를 충족하기 위한 것이며,
run-phase 는 이 목록을 체크리스트로 소비한다. 목록이 불완전하면 수정도 불완전하다.

- 측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, base `1d150a27d`
- 측정 일자: 2026-09-12
- SPEC: `SPEC-DOCS-TABCOUNT-DRIFT-001`

---

## 1. 기준값 — 이 트리에서 다시 잰 탭 수

**14.** 서로 독립인 세 측정이 일치한다. 카드가 들고 온 숫자를 그대로 옮긴 것이 아니라 이 트리에서 다시 쟀다.

| # | 명령 | 관측된 출력 |
|---|---|---|
| 1 | `sed -n '30,90p' internal/web/schemaform.go \| grep -c 'LabelKey:'` | `14` |
| 2 | `grep -n -A 20 'wantTabOrder' internal/web/tab_layout_test.go` | 리터럴에 id 14개: identity, language, launch, llm, workflow, git-worktree, audit, codex, agentfm, report, mcp, crosssession, feedback, gate |
| 3 | `go test ./internal/web/ -run 'TestConsoleTabsOrder'` | `ok github.com/modu-ai/moai-adk/internal/web 0.794s` |

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
| B1 | `.../ko/advanced/moai-web-console.md` | 25 | ko | `(14개 탭)` | 숫자 | 예 |
| B2 | `.../ko/advanced/moai-web-console.md` | 128 | ko | `14개 탭이 세로 목록으로` | 숫자 | 예 — 바로 아래 이름 목록이 있음 |
| B3 | `.../en/advanced/moai-web-console.md` | 25 | en | `(14 tabs)` | 숫자 | 예 |
| B4 | `.../en/advanced/moai-web-console.md` | 128 | en | `unfolds fourteen tabs` | **낱말** | 예 — 바로 아래 이름 목록이 있음 |
| B5 | `.../ja/advanced/moai-web-console.md` | 25 | ja | `（14 タブ）` | 숫자 | 예 |
| B6 | `.../ja/advanced/moai-web-console.md` | 128 | ja | `14 のタブが` | 숫자 | 예 — 바로 아래 이름 목록이 있음 |
| B7 | `.../zh/advanced/moai-web-console.md` | 25 | zh | `（14 个标签页）` | 숫자 | 예 |
| B8 | `.../zh/advanced/moai-web-console.md` | 128 | zh | `14 个标签页的纵向列表` | 숫자 | 예 — 바로 아래 이름 목록이 있음 |

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

## 6. D군 — 같은 결함의 두 번째 사례: 탭 **이름** 드리프트 (8자리)

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

목록 밖 산문에도 같은 이름이 4자리 더 있다 — `docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md:165`
(`서드파티 LLM 탭` / `The 3rd Party LLM tab` / `サードパーティ LLM タブ` / `第三方 LLM 标签页`).
목록과 함께 고쳐야 하며, 기계 가드는 목록만 본다(산문은 사람 손 몫 — `spec.md` §7 잔여 위험).

나머지 13개 이름은 렌더 라벨과 일치한다. 어느 쪽을 고쳐야 하는지(문서를 코드에 맞출 것인가,
코드 라벨 자체가 틀렸는가)는 plan.md §C 에서 갈래 1(문서를 코드에 맞춘다)로 결정되었다 (2026-09-12).

## 7. 오탐 — 가드가 반드시 걸러야 할 자리

수만 보고 긁는 가드는 아래를 잘못 잡는다. 그래서 가드는 **수 + 탭 명사** 를 붙여서 본다.

| 파일 | 행 | 내용 | 왜 오탐인가 |
|---|---|---|---|
| `README.md` | 422 | `measured nine forms … all nine proved reproducible` | SVG 인포그래픽 형태 수. 탭과 무관 |
| `README.ko.md` | 422 | `… 아홉 가지 형태를` | 위와 같음 |
| `README.md` | 418 | `Eleven ref skills` | 스킬 수 |
| `README.md` | 486 | `Eleven of the twelve are agents` | 에이전트 수 |
| `docs-site/content/ja/claude-code/extensibility/plugins.md` | 100 | `4 タブを持つプラグインマネージャー` | **Claude Code 플러그인 매니저**의 탭. 수+탭 명사가 붙어 있으므로 파일 화이트리스트로만 걸러진다 |
| `docs-site/content/*/core-concepts/kanban-board-terms.md` | 9, 29 | `nine terms` / `아홉 낱말` / `九个词` | 칸반 용어 수 |
| `docs-site/content/*/advanced/statusline.md` | 203 | `열여섯 개 키` / `十六个键` | statusline 세그먼트 수 |
| `docs-site/content/*/advanced/moai-web-console.md` | 137 | `codex 설정 12개` / `twelve scattered codex settings` | codex 패널 내부 설정 수 — t509 반경. 본 카드 범위 밖 |

## 8. 스크린샷 — 별건으로 분리

`assets/images/moai-web-settings.png` 는 README 4본(각 411행)에서 참조되며, 11탭 시절 이미지다.
t509 가 alt 텍스트에서 수를 없앴을 뿐 이미지는 그대로다. 재생성 절차는 저장소 어디에도 기록되어 있지 않다.
본 카드에 접지 않는 근거와 후속 카드 제안은 `spec.md` §5 에 있다.

---

## 재현 명령

```bash
# 기준값 (셋 다 돌려서 일치를 본다)
sed -n '30,90p' internal/web/schemaform.go | grep -c 'LabelKey:'
grep -n -A 6 'var wantTabOrder' internal/web/tab_layout_test.go
go test ./internal/web/ -run 'TestConsoleTabsOrder'

# A~C군 재열거 (수 + 탭 명사 인접)
grep -rnE '([0-9]+|nine|ten|eleven|twelve|thirteen|fourteen|fifteen|九|十四|열네|아홉)[^.。]{0,4}(tabs?|개 탭|-탭|タブ|个标签页|标签页)' \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/ko/cli-reference/web.md docs-site/content/en/cli-reference/web.md \
  docs-site/content/ja/cli-reference/web.md docs-site/content/zh/cli-reference/web.md \
  docs-site/content/ko/advanced/moai-web-console.md docs-site/content/en/advanced/moai-web-console.md \
  docs-site/content/ja/advanced/moai-web-console.md docs-site/content/zh/advanced/moai-web-console.md
```
