# t492 — always-loaded 표면 구조 최적화 분석

카드: t492 · 브랜치: `WT-rules-always-loaded-diet` · 워크트리: `.claude/worktrees/t492`

**모든 수치의 측정 트리: `cc78d1479091dc45834565dccdc9ad91d2f56c68`** (로컬 develop 흡수 tip, fast-forward, 충돌 0). 이 SHA 밖에서 잰 값은 이 문서에 없다. 배차문이 전한 `main` 기준 수치(32,800 · 185,368 · VCI 8,224 등)는 **일절 사용하지 않았다**.

---

## 0. Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. 측정 트리 `cc78d1479` 에서 always-loaded 표면은 **77,723 토큰 / 17 entries** 이고 예산 77,600 대비 **여유 −123**, 즉 가드가 **FAIL** 이다.
2. 재현 스크립트의 총계가 가드 출력과 **바이트 단위로 일치**하므로 이 문서의 항목별 분해는 두 번째 측정 경로가 아니다.
3. always-loaded 14본은 **14/14 companion 을 갖고 있으나 이관율이 20%~74% 로 균일하지 않다**. `verification-claim-integrity.md` 는 이관율 20% 로 최저이며, 그 안의 §2.1 단일 블록이 **14,796 B(3,699 tok, 표면의 4.8%)** — 표면 전체에서 가장 큰 단일 이동 가능 블록이다.
4. `AGENTS.md`(3,693 tok)는 rules 트리의 압축 미러이며, 표본 21개 조항 전부가 rules 트리에 대응을 갖는다. Claude 세션에서 그 비용은 `CLAUDE.md:9` 의 `@AGENTS.md` 한 줄에서 나온다.
5. `kanban-dispatch.md` 는 **최대 파일이지만 최대 이동 블록을 갖고 있지 않다** — 최대 절이 5,694 B 이고 16개 절에 고르게 분산돼 있다. 배차문이 이 파일을 1순위 후보로 지목한 것은 파일 크기 기준이며, 블록 이관 기준으로는 1순위가 아니다.

### Evidence

**E1 — 가드 실측** (`go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v`):

```
=== RUN   TestAlwaysLoadedTokenBudget
    token_budget_guard_test.go:69: always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)
    token_budget_guard_test.go:73: always-loaded surface = 77723 tokens, exceeds budget 77600 (overflow 123); surface has 17 entries — trim always-loaded rules or raise AlwaysLoadedTokenBudget with justification
--- FAIL: TestAlwaysLoadedTokenBudget (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/config	0.449s
```

**E2 — 재현 동일성** (`bash .moai/reports/t492/measure-surface.sh <worktree> .moai/reports/t492/breakdown.tsv`):

```
SCRIPT_TOTAL_TOKENS=77723 SCRIPT_ENTRIES=17
```

`77723 == 77723`, `17 == 17`. 스크립트는 `internal/config/token_budget_guard.go` 의 `alwaysLoadedSurface` + `measureAlwaysLoaded` 를 그대로 옮긴 것이다: `.claude/rules/moai/**/*.md` 중 frontmatter 에 `paths:` 키가 없는 것(정렬) + 고정 3슬롯, 파일당 `floor(bytes/4)`, 그 합.

**E3 — 절 크기 측정**: `bash .moai/reports/t492/section-sizes.sh <file>` (코드 펜스 내부의 `#` 을 헤딩으로 오인하지 않도록 펜스 상태를 추적한다).

### Baseline-attribution

| 측정 | 명령 | 관측 |
|---|---|---|
| 표면 총계 | `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` | `77723 tokens … headroom -123, 17 entries` / `FAIL` |
| 재현 총계 | `bash .moai/reports/t492/measure-surface.sh` | `SCRIPT_TOTAL_TOKENS=77723 SCRIPT_ENTRIES=17` |
| 트리 | `git rev-parse HEAD` | `cc78d1479091dc45834565dccdc9ad91d2f56c68` |
| VCI 크기 | `wc -c .claude/rules/moai/core/verification-claim-integrity.md` | `26629` |
| VCI §2.1 | `section-sizes.sh` | `14796  ### 2.1 Moving-ref attribution …` |

리드가 같은 트리에서 `77,723 / −123 / 17 entries / FAIL` 을 독립 재현했다. 두 측정은 서로 다른 세션에서 독립적으로 나왔고 일치한다.

### Gaps — 관측하지 않은 것

- **조항 전수 대응은 확인하지 않았다.** AGENTS.md ↔ rules 트리 대응은 **표본 21개**를 확인했을 뿐이며, "모든 조항이 대응을 갖는다"는 전수 주장이 **아니다**. 전수 검증 방법은 §4 에 적었다.
- **`kanban-dispatch.md` 를 제외한 파일들의 조건부화 가능성은 개별 판단하지 않았다.** §3 의 조건부화 축은 kanban / cross-session 두 파일에 대해서만 논거를 세웠다.
- **t471(lane-3)의 델타는 반영되지 않았다.** lane-3 이 `kanban-dispatch.md` § Deputy dispatch surface 를 순증 ≤ 0 목표로 편집 중이며, 그 착지 후 `kanban-dispatch.md` 의 34,268 B 는 달라질 수 있다.
- **t473 의 상환(+1,044)은 이 트리에 없다.** 그 브랜치는 미푸시 상태로 창을 기다린다.
- **companion 파일의 내용 정합성은 검사하지 않았다.** 존재와 크기만 쟀다.
- **템플릿 미러 3건의 드리프트 원인은 조사하지 않았다**(§6 부수 발견).

### Residual-risk

- 이 문서의 절 크기는 **바이트**이고 가드는 **파일당 `floor(bytes/4)`** 를 센다. 절 여러 개를 옮길 때 파일 단위 내림 때문에 예상 절감이 최대 1 토큰 어긋날 수 있다. 절감 추정치는 이 오차를 무시할 수 있는 규모다.
- 이관 제안(§4)의 절감은 **stub 에 남길 분량을 내가 정한 값**에 의존한다. 실제 run 단계에서 stub 이 더 두꺼워지면 절감은 줄어든다. 그래서 §4 는 보수/공격 두 층으로 제시했다.
- 표면은 계속 자란다. 리드가 인용한 **하루 성장 +493** 전례가 있으므로, 여기서 만든 여유는 영구적이지 않다.

---

## 1. 산출물 1 — 17 entries 전수 실측표

측정 트리 `cc78d1479`. 토큰 = `floor(bytes/4)`, 가드와 동일.

| # | 경로 | 바이트 | 토큰 | 표면 비중 | companion | 이관율 |
|---:|---|---:|---:|---:|---:|---:|
| 1 | `.claude/output-styles/moai/moai.md` | 66,255 | 16,563 | 21.3% | — (고정 슬롯) | — |
| 2 | `.claude/rules/moai/workflow/kanban-dispatch.md` | 34,268 | 8,567 | 11.0% | 34,657 | 50% |
| 3 | `.claude/rules/moai/core/agent-common-protocol.md` | 26,829 | 6,707 | 8.6% | 30,712 | 53% |
| 4 | `.claude/rules/moai/core/verification-claim-integrity.md` | 26,629 | 6,657 | 8.6% | 6,860 | **20%** |
| 5 | `.claude/rules/moai/core/askuser-protocol.md` | 21,822 | 5,455 | 7.0% | 15,748 | 42% |
| 6 | `.claude/rules/moai/workflow/session-handoff.md` | 21,197 | 5,299 | 6.8% | 40,449 | 66% |
| 7 | `CLAUDE.md` | 19,766 | 4,941 | 6.4% | — (고정 슬롯) | — |
| 8 | `.claude/rules/moai/core/moai-constitution.md` | 16,506 | 4,126 | 5.3% | 7,687 | 32% |
| 9 | `.claude/rules/moai/workflow/cross-session-messaging.md` | 16,332 | 4,083 | 5.3% | 10,645 | 39% |
| 10 | `AGENTS.md` | 14,774 | 3,693 | 4.8% | — (고정 슬롯) | — |
| 11 | `.claude/rules/moai/workflow/context-window-management.md` | 8,882 | 2,220 | 2.9% | 7,548 | 46% |
| 12 | `.claude/rules/moai/workflow/main-checkout-branch-guard.md` | 7,800 | 1,950 | 2.5% | 8,471 | 52% |
| 13 | `.claude/rules/moai/workflow/cache-aware-execution.md` | 6,980 | 1,745 | 2.2% | 5,881 | 46% |
| 14 | `.claude/rules/moai/workflow/goal-directive.md` | 6,925 | 1,731 | 2.2% | 19,228 | **74%** |
| 15 | `.claude/rules/moai/core/moai-mcp-tools.md` | 6,403 | 1,600 | 2.1% | 7,010 | 52% |
| 16 | `.claude/rules/moai/workflow/skill-routing.md` | 5,595 | 1,398 | 1.8% | 1,359 | 20% |
| 17 | `.claude/rules/moai/core/native-idiom-and-register.md` | 3,952 | 988 | 1.3% | 2,705 | 41% |
| | **합계** | | **77,723** | 100% | | |

rules 14본 바이트 합계 = **210,120 B**. 리드가 정정 메시지에서 전한 `origin/develop 210,120 B` 와 일치한다 — 다만 그 값은 **바이트**이고 가드가 세는 **77,723 토큰(17 entries)** 과는 **모집단도 단위도 다르다**. 두 수치를 같은 문장에서 비교하지 말 것.

### 조건부화 가능 여부

| 파일 | 조건부화 축 | 판정 | 근거 |
|---|---|---|---|
| `moai.md` (output-style) | 없음 | **불가** | 활성 output style 은 정의상 전 세션 로드. 축소만 가능 |
| `CLAUDE.md` | 없음 | **불가** | 프로젝트 지시문 슬롯 |
| `AGENTS.md` | `@`-import 제거 | **가능** | §4 C1 — Codex 는 import 없이 읽고, Claude 는 rules 트리로 같은 조항을 받는다 |
| `kanban-dispatch.md` | SessionStart 조건부 주입 / Skill | **가능(기전 신설)** | §3 |
| `cross-session-messaging.md` | 동일 | **부분 가능** | §3 — 안전 조항은 잔류 필요 |
| 나머지 10본 | `paths:` | **불가** | 전 세션이 언제든 해당 상황에 진입 |

---

## 2. 산출물 2 — 중복 조항과 SSOT 지정안

**중복 제거는 SSOT 를 하나 정하고 나머지를 cross-reference 로 바꾸는 것이지 조항 삭제가 아니다.** 아래 어느 항목도 조항을 없애지 않는다.

### D1 — `AGENTS.md` 전체 ↔ rules 트리 (최대)

AGENTS.md 7개 절이 rules 트리를 압축 미러한다. 표본 21개 조항의 대응(전부 실측):

| AGENTS.md 절 | 바이트 | rules 트리 대응 |
|---|---:|---|
| §1 Evidence and verification claims | 1,497 | `core/verification-claim-integrity.md` (`no unobserved-claim`, `Baseline-integrity attribution`, `Evidence-bearing report format`) |
| §2 Git, branches, shared checkout | 2,092 | `workflow/main-checkout-branch-guard.md` (`MUST NOT change branch state`, `Re-read branch and commit state`) + `core/agent-common-protocol.md` (`sweep prohibition`, `git add -A`) |
| §3 Worktrees | 1,602 | `workflow/kanban-dispatch.md` § Isolation is entered, never provisioned |
| §4 How verification is run | 1,498 | `workflow/kanban-dispatch.md` (`verification is scoped to the card`, `Never spawn background load`, `Merge Risk:`) |
| §5 Core behaviors (6개) | 3,178 | `core/moai-constitution.md` § Agent Core Behaviors (같은 6개, 5,650 B — `sed -n '177,287p' … \| wc -c`) |
| §6 Output, language, format | 1,599 | `core/moai-constitution.md` § Response Language + `core/native-idiom-and-register.md` + `core/agent-common-protocol.md` (`Never use time predictions`) |
| §7 Tools and command output | 1,879 | `core/agent-common-protocol.md` (`Read a file before using Edit`, `Maintain effectiveness without MCP`) + `workflow/cache-aware-execution.md` (`Keep command output bounded`, `Prefer the quiet form`) |

**SSOT 지정**: rules 트리가 SSOT, `AGENTS.md` 는 Codex 전용 압축 사본. 이는 이미 AGENTS.md 자신이 선언한 구조다 — *"Obligations are carried from `.claude/rules/moai/**` and `CLAUDE.md`, which remain the source of truth"*. 즉 **문서상 SSOT 는 이미 정해져 있고, 문제는 그 사본이 Claude 컨텍스트에도 실린다는 것**뿐이다.

### D2 — `kanban-dispatch.md` § Isolation ↔ `AGENTS.md` §3

같은 의무(런처 경유 진입 / `moai worktree done` 은 L2 전용 / 미푸시 브랜치 폐기 금지 / 새 카드는 새 트리 / `WT-` 접두 슬러그 / 추적성 담체)를 **5,694 B 와 1,602 B 두 밀도로** 둘 다 always-loaded 에 싣는다. D1 이 해결되면 자동 소멸한다.

### D3 — AskUserQuestion 채널 독점 — **중복 아님(기각)**

grep 계수(`askuser-protocol` 41 / `agent-common-protocol` 15 / `kanban-dispatch` 3 / `moai-constitution` 1)는 **언급 횟수이지 중복이 아니다.** 원문 대조 결과 `moai-constitution` 과 `agent-common-protocol` 은 이미 cross-reference 로 정리돼 있고, 겹치는 실질은 `### Subagent Prohibitions` 3개 불릿 **약 400 B** 뿐이다. `agent-common-protocol` 쪽은 그 위에 lane spawn authority(askuser-protocol 에 없는 내용)를 얹고 있어 통합 시 오히려 손실이 난다. **기각.**

### D4 — 증거 규율 3본 — **중복 아님(기각)**

`verification-claim-integrity`(주장 무결성) / `agent-common-protocol`(검증 배치 실행) / `kanban-dispatch` § Completion is read(리드의 판독 의무)는 서로 다른 행위자·다른 시점을 규율한다. 표현이 닮았을 뿐 대체 불가. **기각.**

> 배차문이 제시한 중복 후보 3축 중 **2축이 실측으로 기각**됐다. 남은 실질 중복은 D1/D2 이며, 둘은 같은 뿌리다.

---

## 3. 조건부화 축 — 기전은 실재하는가

배차문의 질문: *"SessionStart 조건부 주입 같은 기전으로 `paths:` 없이도 조건화할 수 있는가."*

**답: 기전은 실재하며 이미 테스트로 검증돼 있다.** `internal/hook/handoff_inject_test.go` 가 SessionStart 훅의 `HookSpecificOutput.AdditionalContext` 조건부 주입을 검증한다 — `source == "clear"` 일 때만 주입, `handoff.mode == manual` 이면 미주입. 조건 판정 후 컨텍스트를 주입하는 경로가 이미 프로덕션에 있다.

kanban 모드 판정 신호도 실재한다: `MOAI_KANBAN` 계열 환경변수(`internal/config/envkeys.go`, `internal/cli/kanban.go`).

### 그러나 [HARD] 두 가지 대가가 따른다

**(가) 측정 표면의 사각지대.** 훅으로 주입된 컨텍스트는 가드가 세지 않는다. kanban 리드 세션에서는 실제 비용이 그대로인데 가드는 8,567 토큰이 줄었다고 보고한다. `token_budget_guard.go` 의 t453 주석이 「측정 대상 변경」을 한 번 기각한 것과 같은 위험이다. **조건부화를 채택한다면 주입 표면을 별도 예산으로 함께 세는 가드 확장이 동반돼야 한다** — 그러지 않으면 이 카드는 다이어트가 아니라 계측기 눈속임이 된다.

**(나) 배포 표면과의 정합.** `kanban-dispatch.md` 는 템플릿 미러가 존재하는 배포 파일이다. 조건부 주입으로 옮기면 사용자 프로젝트에서도 훅이 그 역할을 해야 한다.

**그래서 조건부화는 이 카드에서 실행하지 않는다**(§5). 기전 실재는 확인됐고, 채택 여부는 가드 확장을 포함한 별도 카드의 결정이다.

---

## 4. 산출물 3 — 절감 추정치와 측정 방법

**목표선은 0 이 아니라 +123 부터다.** 현재 여유 **−123**.

### C2 — `verification-claim-integrity.md` §2.1 → companion 이관 (권고 1순위)

§2.1(14,796 B) 내부 분해 — 실측:

| 블록 | 줄 | 바이트 | 처분 |
|---|---|---:|---|
| 도입([HARD] 조항 + "pin every ref 가 답이 아니다") | 48–53 | 1,123 | **stub 잔류** |
| The four tests | 54–76 | 2,679 | 이관 |
| Classification and remedy are two separate steps ([HARD]) | 77–80 | 456 | **stub 잔류** |
| The four remediation branches (R1–R4 표) | 81–104 | 2,560 | 층에 따라 |
| The exemption marker (문법 + [HARD] 사유 필수) | 105–119 | 1,102 | **stub 잔류** |
| The five grounded instances | 120–133 | 3,061 | 이관 |
| Detection limits (L1–L7) | 134–145 | 2,876 | 이관 |
| The divergence figure, split by carrier | 146–152 | 939 | 이관 |

| 층 | 이관 바이트 | Δ 토큰 | 이관 후 여유 |
|---|---:|---:|---:|
| **A (보수)** — 4 tests · 5 instances · limits · divergence | 9,555 | **−2,389** | **+2,266** |
| **B (공격)** — A + R1–R4 표 | 12,115 | **−3,028** | **+2,905** |

**stub 에 남는 [HARD] 의무(문면 확인 — lane-3 경고 대응):**
1. `[ZONE:Evolvable] [HARD] A claim decided against a moving ref … carries no baseline` — 도입 블록에 원문 그대로 잔류.
2. `[HARD] The tests return a class; the class does not name the remedy.` — 잔류.
3. 예외 마커의 `The reason is mandatory and non-empty.` — 잔류.
4. 이관 블록을 가리키는 [HARD] 지시문 1줄을 신설: *"이 술어를 적용하기 전에 companion § Moving-ref predicate 를 읽는다"*.

**목적지 사전 확인**(「지우기 전에 이미 있는지부터 재라」): `verification-claim-integrity-detail.md`(6,860 B)의 현재 헤딩은 5-section 설명 2개 + 사례 2개뿐이며 **§2.1 관련 내용이 없다**. 중복 착지 없음.

**A 층만으로도 적자가 해소되고 여유 +2,266 이 남는다.**

### C1 — `CLAUDE.md:9` 의 `@AGENTS.md` 제거

| 항목 | 값 |
|---|---|
| Δ 토큰 | **−3,693** (AGENTS.md 슬롯이 Claude 표면에서 빠짐) |
| 조항 손실 | **없음**(조건부) — 표본 21개 전부 rules 트리 대응 확인. 전수 확인이 전제 |
| Codex 영향 | **없음** — Codex 는 `AGENTS.md` 를 import 없이 직접 읽는다. `CodexContractByteCeiling` 가드는 그대로 유지 |
| 동반 변경 | `alwaysLoadedSurface` 고정 슬롯에서 `AGENTS.md` 제거 + `TestFixedSlotsExistInRepoTree` 갱신 |
| 위험 | **교리 결정** — `CLAUDE.md` §0 이 "imported below" 라고 선언하므로 문면이 함께 바뀐다. 운영자 게이트 사안 |

**[HARD] C1 의 전제 검증 방법**(전수, 손 열거 금지 — 「손 열거 판별식은 자기 결함을 재생산한다」):

AGENTS.md 의 모든 [HARD]/MUST/MUST NOT 문장을 기계적으로 추출해 각각의 rules 트리 대응을 강제로 채우는 표를 만든다. 문구 일치가 아니라 **요구 일치**로 판정하며(「문구 말고 요구로 훑어라」), 대응이 비는 행이 하나라도 남으면 C1 은 기각된다. 이 추출·대조는 run 단계 작업이다.

### C3 / C4 — 조건부화 (이 카드에서 실행하지 않음)

| 후보 | Δ 토큰 | 전제 |
|---|---:|---|
| `kanban-dispatch.md` 조건부 주입 | −8,567 | 가드의 주입-표면 확장이 선행 |
| `cross-session-messaging.md` 부분 조건부 | −4,083 중 일부 | 안전 [HARD] 3조항 잔류 필요 |

### 측정 방법 (모든 층 공통)

```bash
# 1. 편집 전 baseline — 반드시 편집 커밋 앞 커밋에 남긴다 (VCI §2.3)
go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
bash .moai/reports/t492/measure-surface.sh "$(git rev-parse --show-toplevel)" before.tsv

# 2. 편집 후 — 같은 두 명령
go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
bash .moai/reports/t492/measure-surface.sh "$(git rev-parse --show-toplevel)" after.tsv

# 3. 귀속 — 항목별 델타
diff before.tsv after.tsv
```

[HARD] 스크립트 총계가 가드 출력과 일치하는지를 **매번 먼저** 보인다. 일치하지 않으면 그 회차의 항목별 분해는 근거가 아니다.

---

## 5. 산출물 4 — 실행을 이 카드에서 할지 분리할지

**판단: C2 는 이 카드에서 실행하고, C1 · C3 · C4 는 분리한다.**

### C2 를 이 카드에서 하는 근거

1. **적자가 지금 나 있다.** develop 가드가 FAIL 이고, C2 A 층 하나로 −123 → +2,266 이 된다. 적자 해소를 다음 카드로 미룰 근거가 없다.
2. **기전 변경이 없다.** 파일 한 개 안에서 stub ↔ 기존 companion 으로 블록을 옮기는 것뿐이다. 가드·훅·템플릿 어느 것도 건드리지 않는다.
3. **범위가 한 파일이고 목적지가 비어 있다.** 다른 레인과 충돌하지 않는다 — lane-3 의 t471 은 `kanban-dispatch.md` § Deputy dispatch surface 안으로 스스로 범위를 좁혔고, t473 은 `moai.md` 와 `main-checkout-branch-guard.md` 를 건드린다. **셋 다 VCI 와 겹치지 않는다.**
4. **검증이 값싸고 결정적이다.** 가드 테스트 하나가 판정한다.

### C1 을 분리하는 근거

1. **교리 결정이다.** `CLAUDE.md` §0 의 선언을 바꾸므로 운영자 게이트 사안이며, 레인이 열 수 없다.
2. **전제 검증이 이 카드 범위를 넘는다.** 표본 21개는 가설을 세우기에 충분하지만 전수 대조는 별도 작업량이다. **표본으로 전수를 주장하지 않겠다**는 것이 이 카드의 입장이다.
3. **가드 코드 변경을 동반한다.** 고정 슬롯 목록과 그 테스트를 함께 고쳐야 하므로 문서 카드가 아니라 코드 카드가 된다.

### C3 / C4 를 분리하는 근거

**가드 확장이 선행되지 않으면 조건부화는 다이어트가 아니라 계측기 눈속임이다**(§3 가). 순서가 뒤집히면 되돌리기 어렵다.

### 제안 후속 카드

| 카드 | 내용 | 선행 |
|---|---|---|
| C1 카드 | AGENTS.md 조항 전수 대조 → `@`-import 제거 판단 | 없음(조사부터) |
| 가드 확장 카드 | 훅 주입 표면을 별도 예산으로 계수 | 없음 |
| C3 카드 | `kanban-dispatch.md` 조건부 주입 | 가드 확장 카드 |
| 템플릿 드리프트 카드 | §6 부수 발견 3건 | 없음 |

---

## 6. 부수 발견 — 이 카드 범위 밖

### F1 — 템플릿 미러 드리프트 3건 (live > template)

| 파일 | live | template | 차 |
|---|---:|---:|---:|
| `workflow/cross-session-messaging.md` | 16,332 | 16,269 | −63 |
| `workflow/main-checkout-branch-guard.md` | 7,800 | 7,575 | −225 |
| `workflow/main-checkout-branch-guard-detail.md` | 8,471 | 8,109 | −362 |

`CLAUDE.local.md` §2 [HARD] Template-First 위반. **조치하지 않았다**(범위 규율).

### F2 — `agent-common-protocol.md` 의 부정확한 §-인용

`agent-common-protocol.md` § Subagent Prohibitions 이 `kanban-dispatch.md § Lane spawn authority` 를 인용하나, `kanban-dispatch.md` 에 그 이름의 절은 없다 — 해당 문단(line 238)은 `## Factory Mode — the card travels whole` **안의 볼드 문단**이다. 대상은 실재하므로 dangling 은 아니고, §-단위 좌표가 부정확하다. **조치하지 않았다.**

### F3 — 배차문 수치의 오차 구조

배차문의 `main` 기준 값과 develop 실측의 차 24,752 B 중 **18,405 B(74%)가 `verification-claim-integrity.md` 한 파일**이다(8,224 → 26,629, `f7662da98` t300 이 §2.3 을 추가). 리드의 정정 메시지는 이를 "13% 과소평가"로 전했는데, **오차는 균등하지 않고 한 파일에 집중돼 있다.** 파일별 수치를 쓰는 쪽에서는 이 편차가 중요하다.

---

## 7. 재현 자산

| 파일 | 용도 |
|---|---|
| `.moai/reports/t492/measure-surface.sh` | 가드 계수기 재현 — 총계가 가드 출력과 일치함을 보이는 데 쓴다 |
| `.moai/reports/t492/section-sizes.sh` | 파일의 `##`/`###` 절 바이트 (코드 펜스 인식) |
| `.moai/reports/t492/companion-check.sh` | always-loaded 각 룰의 companion 존재·크기 |
| `.moai/reports/t492/breakdown.tsv` | 17 entries 항목별 바이트/토큰 (트리 `cc78d1479`) |
| `.moai/reports/t492/build-{stub,companion}.sh` | live 쌍에 적용한 이관(먼저 실행) |
| `.moai/reports/t492/build-mirror.sh` | 위 둘을 하나로 합친 판 — 템플릿 미러에 적용. 처리 대상 stub 에서 블록을 뽑으므로 중립화 미러가 자기 문구를 유지한다 |
| `.moai/reports/t492/sync-companion-meta.py` | companion 3개 메타 갱신 — 대상 문자열 부재 시 종료코드 1(무음 no-op 방지) |
| `.moai/reports/t492/moved-block-{A,B}-*.md` | 이관된 원문 2블록 (감사용) |
| `.moai/reports/t492/stub-pointer.md` · `companion-insert-header{,-template}.md` | 신설 문단 원문 |

---

## 8. C2 실행 기록 (운영자 승인 후 · 보수층 한정)

승인 범위: **C2 보수층**. 공격층(R1–R4 표 이관)은 승인 범위 밖이므로 **손대지 않았다** — R1–R4 표와 비용표는 stub 에 그대로 있다.

### 8.1 순서 귀속 (VCI §2.3 — 이 편집이 고치는 바로 그 파일의 조항)

baseline 은 **변경 커밋보다 앞선 자기 커밋 `4113cb667`**(§0 Evidence 의 가드 원문 `77723 … headroom -123 … FAIL`)에 이미 착지해 있었다. 커밋 그래프가 순서의 증인이며, 이 문서의 주장은 같은 커밋 쌍에 얹혀 있지 않다.

### 8.2 가드 전/후 원문

**전** (트리 `cc78d1479`, `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v`):

```
    token_budget_guard_test.go:69: always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)
--- FAIL: TestAlwaysLoadedTokenBudget (0.01s)
```

**후** (같은 명령, 같은 워크트리, 편집 후):

```
    token_budget_guard_test.go:69: always-loaded surface = 75483 tokens (budget 77600, headroom 2117, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.00s)
```

**−123 → +2,117**. entries 는 17 로 불변(파일 신설·삭제 없음).

### 8.3 절감 실측이 예측과 다른 이유 — 전수 귀속

| 항목 | 바이트 |
|---|---:|
| 이관(4 tests + 5 instances + L1–L7 + divergence) | −9,555 |
| 신설 [HARD] companion 포인터 | +600 |
| `the predicate below` → `the predicate` (지시 대상 소실 수정) | −6 |
| **stub 순증** | **−8,961** |

stub 26,629 → 17,668 B. 토큰 `floor(26629/4)=6657` → `floor(17668/4)=4417`, **Δ = −2,240**. 가드 실측 `77,723 → 75,483 = −2,240` 과 일치한다.

**예측치 −2,389 는 포인터 비용(600 B ≈ 150 tok)을 계산에 넣지 않은 값이었다.** 실측이 예측보다 149 토큰 적다. 예측을 폐기하고 실측을 쓴다.

### 8.4 [HARD] 잔류 의무 문면 확인 — lane-3 경고 대응

> 「가드는 항목 수를 세지 내용을 세지 않는다. stub 이 지나치게 얇아져 의무가 companion 에만 남는 이관은 가드를 통과하면서 실질을 잃는다.」

**기계적 확인 — 이관 블록의 의무 문장 개수:**

```
grep -c '\[HARD\]' moved-block-A-four-tests.md moved-block-B-instances-limits-divergence.md
  → 0, 0
grep -c 'MUST'     moved-block-A-four-tests.md moved-block-B-instances-limits-divergence.md
  → 0, 0
```

**이관된 9,555 B 안에 [HARD] 도 MUST 도 하나도 없다.** 옮겨간 것은 절차·사례·탐지한계뿐이다.

**stub 의 `[HARD]` 총수: 7 → 8** (`git show HEAD:<stub> | grep -c` vs 현재). 하나도 빠지지 않았고 신설 포인터가 하나 늘었다.

**잔류 4건 문면 인용** (`.claude/rules/moai/core/verification-claim-integrity.md`, 편집 후 행번호):

- **:50** — `[ZONE:Evolvable] [HARD] A claim decided against a **moving ref** … carries no baseline in the sense §2 requires.`
- **:54** *(신설)* — `[HARD] The predicate is applied, not recalled. Before remediating any moving-ref or moving-coordinate claim, read `verification-claim-integrity-detail.md` § Moving-ref predicate and run its four tests in order; they return one of two classes — **ANCHOR** … or **SUBJECT** … Reaching a remedy below without having run the tests is indiscriminate pinning by another name.`
- **:58** — `[HARD] The tests return a **class**; the class does not name the remedy. There are **two classes and four remedies** — ANCHOR selects between R1 and R2, SUBJECT between R3 and R4.`
- **:94** — `- **The reason is mandatory and non-empty.** A bare marker would make "silence the warning" cheaper than "pin the SHA", inverting the incentive this clause sets.`

신설 포인터가 두 클래스 이름(ANCHOR / SUBJECT)을 stub 안에서 정의하므로, 아래 R1–R4 표의 `Class` 열은 tests 가 없어도 stub 만으로 읽힌다.

### 8.5 목적지 사전 확인 — 「지우기 전에 이미 있는지부터 재라」

편집 전 `verification-claim-integrity-detail.md`(6,860 B)의 `^#` 헤딩 전수:

```
# Verification-Claim Integrity — Detail Companion
## What each of the five sections contains       (### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)
## Cross-references (each remains the single source of truth for its own subject)
## Worked example — defect-claim hazard
## Worked example — retention-claim hazard
```

**§2.1 관련 내용 없음.** 중복 착지 없이 순증으로 붙였다(6,860 → 17,278 B).

### 8.6 끊어진 인용 확인

`.claude/rules/` 전체에서 이관된 하위절(`grounded instances` / `the four tests` / `Detection limits`)을 인용하는 파일: **0건**(VCI 두 파일 제외). rules 트리 안 dangling 참조 없음.

리포 전체로 넓히면 `CHANGELOG.md` 와 SPEC 문서들이 이 하위절들을 언급하나, **그것들은 당시 착지 사실의 역사 기록이므로 고치지 않는다** — 고치면 기록이 거짓이 된다.

### 8.7 템플릿 미러 — Template-First [HARD]

`internal/template/templates/.claude/rules/moai/core/` 의 두 미러에 **같은 이관을 적용**했다. `build-mirror.sh` 는 처리 대상 파일에서 블록을 뽑으므로 미러의 중립화 문구(Instance 2 / Instance 4)가 보존된 채 companion 으로 함께 이동했다. companion 헤더는 카드 id 를 뺀 중립판(`companion-insert-header-template.md`)을 썼다.

**미러 정합 전수 귀속** (`diff live template | grep -c '^[<>]'`):

| | 편집 전 | 편집 후 |
|---|---:|---:|
| stub | 25 | 21 |
| companion | 4 | 10 |
| **합** | **29** | **31** |

`31 = 29 − 4(중립화 2쌍이 stub→companion 이동) + 4(같은 2쌍이 companion 에 도착) + 2(헤더 중립화, 의도적 신설)`. **모든 diff 라인이 귀속된다.**

> 이 정합 확인 과정에서 **자체 결함 1건을 적발해 수정**했다: `§ §2.1` 중복 기호 수정을 live 에만 적용하고 미러에는 적용하지 않아 두 파일이 갈렸다(원인: 미러 빌드가 읽는 포인터 원본 파일이 미수정 상태였음). diff 라인수가 기대치와 2줄 어긋난 것이 단서였다. 포인터 원본과 미러 양쪽을 고쳐 21/10 으로 정렬했다.

### 8.8 검증

```
go test ./internal/template/... ./internal/config/...
ok  github.com/modu-ai/moai-adk/internal/template            24.936s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit    1.473s
ok  github.com/modu-ai/moai-adk/internal/config               4.253s
ok  github.com/modu-ai/moai-adk/internal/config/atomicfile     0.773s
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy     1.791s
```

`make build` 실행 — `catalog.yaml updated successfully (12899 bytes)` 후 `git status` 에 `internal/template/catalog.yaml` 변경 없음(카탈로그는 agents/skills 를 덮고 rules 는 덮지 않는다).

**미검증**: 전 패키지 스위트는 로컬에서 돌리지 않았다(§4.1 규율). 전수 판정은 CI 몫이다.

### 8.9 이 실행이 만들지 못한 것

- **여유 +2,117 은 t471 델타(+564 B ≈ +141 tok, 리드 전달)를 반영하지 않은 값이다.** t471 이 착지하면 +2,117 − 141 ≈ +1,976 이 된다. **다만 그 −141 은 내가 재지 않았고 리드가 전한 값이므로 내 측정이 아니다** — 착지 후 재측정이 필요하다.
- t473 의 +1,044 도 이 트리에 없다. 셋이 모두 착지하면 여유는 대략 +3,000 대가 되나, **그 합산은 예측이지 측정이 아니다.**
- **표면 성장은 계속된다.** 하루 +493 전례가 있으므로 이 여유는 영구적이지 않다. C1(−3,693) 이 다음 상환 수단으로 남아 있다.
