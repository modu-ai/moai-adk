auditor-model: claude-opus-5-5[1m]

# t1175 — SPEC-ALWAYS-LOADED-DIET-002 sync-phase 범위 한정 재감사

카드 t1175 · 워크트리 `.claude/worktrees/t1175` · 브랜치 `WT-rules-diet`
감사 대상 ref: `950fcc492` (수리 범위 `805d44bed..950fcc492` — `8e50ef148`·`7deb5b3b1`·`950fcc492`)
직전 감사: `.moai/reports/t1175/sync-audit.md` (FAIL 0.71, blocking F1~F4)
평가 프로필: 기본 프로필(Functionality + Security must-pass), 평면 가중 모드.

**HEAD 이동 여부: 없음.** 감사 시작 때 `950fcc492434d5c0822f246fe509c368b325a074`, 종료 때(2026-09-27 18:50:55 KST) 같은 SHA였고, 두 시점 모두 `git status --short` 출력이 비어 있었다. 이번 감사 창에서는 외부 쓰기가 관측되지 않았다.

---

## 판정 요약

| 차원 | 점수 | 판정 | 근거 |
|---|---:|---|---|
| Functionality (40%) | 88/100 | PASS | F1~F4 가 모두 해소됐다. AC-ALD2-001 기록값이 병합 이후 트리와 일치하고(198,361 재현), REQ-ALD2-001 이 AC 층과 같은 것을 요구하게 됐다. 수리로 새로 생긴 끊긴 앵커는 0건이다 |
| Security (25%) | 90/100 | PASS | 동결 다중집합 해시가 기준선과 바이트 동일하다(170줄). 수리 diff 의 구속 토큰 줄은 0건이다 |
| Craft (20%) | 78/100 | PASS | 완료 정의의 테스트 세트가 통과한다. 선택 항목 F5~F8 은 그대로 열려 있고, `progress.md` 재개 지점이 닫히지 않았다(N1) |
| Consistency (15%) | 90/100 | PASS | 수리는 두 트리에 같은 바이트로 들어갔다. 미러 hunk 수는 수리 전과 같고, codex `.toml` 은 방출 결과와 일치하며, 카탈로그 해시도 파일과 맞는다 |

가중 조화평균 = 1 / (0.40/88 + 0.25/90 + 0.20/78 + 0.15/90) ≈ **86.6**. must-pass 두 차원이 모두 통과했고, blocking 결함은 없다.

**판정은 PASS-WITH-DEBT 다.** 채무의 본체는 런타임 한도 150,000 에 대한 잔여 **48,361자**다. 개정된 REQ-ALD2-001 에 따라 이 잔여는 이 카드의 요구에서 빠졌고, 후속 카드 `t1226` 이 소유한다. 여기에 선택 항목 F5~F8 과 이번에 새로 기록한 N1·N2 가 더해진다.

### F1~F4 처분

| 결함 | 수리 커밋 | 판정 | 확인한 것 |
|---|---|---|---|
| F1 `sync-auditor` §Language Handling 끊긴 앵커 | `8e50ef148` | **해소** | 세 사본(C1 `.md`, C2 `.md`, C3 `.toml`)의 31/31/20행이 두 절에 각자의 파일을 붙인다. `§Skeptical Evaluation Stance` 는 `agent-common-protocol-reference.md:316` 에, `§Language Handling` 은 `agent-common-protocol.md:89` 에 있으며, 두 트리 모두 같다. C3 는 `agentemit` 골든 테스트가 방출 결과와 일치함을 확인했다(손편집이 아니다). 같은 잘못된 짝(`-reference.md` + Language Handling)은 트리 어디에도 남아 있지 않다 |
| F2 `moai-constitution-detail.md` § Parallel Execution 끊긴 포인터 | `8e50ef148` | **해소** | 두 트리의 37행에서 포인터가 사라졌고, 남은 대상 `dynamic-workflows.md` 에는 `## The Three Orchestration Primitives`(17행)가 두 트리 모두에 있다. 앵커 스윕 차집합에서 끊김 2건(라이브·템플릿)이 사라졌고 새로 생긴 항목은 없다. 재배치 표 R-70 에는 실제 처리(삭제, 정본 `dynamic-workflows.md`)가 개정 표시와 함께 기록됐다 |
| F3 REQ-ALD2-001 미개정 | `7deb5b3b1` | **해소** | REQ 가 「감축 + 잔여의 귀속 기록, 한도 달성은 `t1226` 소유」로 개정됐다. 종전 문구는 폐기 표시와 함께 축자로 보존됐다. 제목·H1·§A·§B 에 개정 3 주석이 붙었고, `§D.2` 추적성 행도 같은 내용으로 정렬됐다. REQ↔AC 대조에서 모순은 없다(아래 「교차 층 대조」) |
| F4 AC-ALD2-001 기록값이 병합 전 값 | `7deb5b3b1` | **해소** | `§AC-ALD2-001.2` 에 트리별 이력 표(기준선·병합 전·병합 트리·`8e50ef148`)가 들어갔다. 기록값 198,361 은 이번 감사에서 `950fcc492` 로 AC 블록을 축자 실행해 재현했다. `8e50ef148..950fcc492` 는 계수 대상 18경로를 하나도 건드리지 않으므로 두 ref 의 값이 같다는 것도 확인했다 |

### 교차 층 대조 (spec.md ↔ acceptance.md ↔ design.md ↔ plan.md)

- **REQ-ALD2-001 ↔ AC-ALD2-001**: REQ 의 세 요소(기준선 아래로 감축 · 잔여 기록 · 명령·트리·시점 귀속)가 AC 의 FAIL 조건 4(감축 ≤ 0) · Then 항목 (3) · FAIL 조건 2(귀속 부재)에 하나씩 대응한다. 층간 불일치는 해소됐다.
- **제목 ↔ REQ**: 제목과 H1 은 「감축, 150,000 잔여는 채무(후속 t1226)」를 말하고, REQ 도 같은 것을 말한다.
- **§A 계량기 표의 「겨눈다」(`spec.md:65`)**: 이 표는 두 계량기 가운데 어느 것을 겨누는지를 가르는 표이지 달성 약속이 아니다. 개정 3 과 충돌하지 않는다고 판정했다.
- **산술**: 246,943 − 198,361 = 48,582, 198,361 − 150,000 = 48,361, F 구간 대비 초과 27,833 / 26,666 / 25,498, 병합 기여 +454, `805d44bed` 의 +62, F2 의 −52 가 모두 맞는다. −52 는 삭제된 문자열 `` `moai-constitution-detail.md` § Parallel Execution; `` 의 문자 수와도 일치한다.
- **다른 AC 와의 충돌**: 수리 diff 는 어떤 AC 의 Given/When/Then 도 고치지 않았다. AC 는 9개이고(`^### AC-ALD2-` 계수 9), MUST-PASS 8개 지정도 그대로다. AC-ALD2-002 의 보조 진단 수치(198,361, F 최고값보다 25,498 위)는 갱신된 값과 맞는다.
- **plan.md**: 150,000 을 「폐기된 목표」로만 언급하므로 개정 3 과 모순되지 않는다. 다만 `t1226` 을 이름으로 적지는 않는다(판정에 영향 없음).

---

## Findings (structured defect-list)

직전 감사의 F1~F4 는 위 표대로 해소됐다. F5~F9 는 재개하지 않고 상태만 적는다.

- **F5** [Low] [optional] — **변화 없음.** 수리 diff 가 `agent-common-protocol-reference.md` 를 건드리지 않았다.
- **F6** [Low] [optional] — **변화 없음.**
- **F7** [Low] [optional] — **변화 없음.**
- **F8** [Low] [optional] — **변화 없음.** 앵커 스윕 차집합에 이 항목들의 변동이 없다.
- **F9** [Info] [optional] — **변화 없음.** `moai todo` 에서 `t1191`·`t1192` 는 여전히 `queued` 다.
- **N1** [Low] [optional] `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/progress.md:352-372`(「재개 지점」) — 2단계(병합 트리 재측정)는 이미 수행됐고, 결과는 `acceptance.md §AC-ALD2-001.2` 와 `remeasure/f4-ac001-ac002.md` 에 기록됐다. 그런데 진행 기록에는 그 사실을 가리키는 줄이 없고, 「지금 상태」 표도 여전히 HEAD `f385b6255`·미푸시 23 을 적고 있다. 값마다 「이 기록 시점」이라는 귀속이 붙어 있어 거짓은 아니다. 하지만 이어받는 사람이 이미 끝난 재측정을 다시 할 수 있다. 확신도 높음. — Required fix: 재개 지점 아래에 sync-phase 종결 한 줄을 둔다(재측정 결과 위치, 수리 커밋 세 개, 이 재감사 경로).
- **N2** [Info] [optional] 큐 카드 `t1226` 본문 — REQ-ALD2-001 개정 3 은 「한도 달성의 채무」를 `t1226` 에 넘긴다. 그러나 카드 본문의 범위는 `A_adm` 실측, 즉 잔여의 성격을 가리는 일이고, 한도 달성은 약속하지 않는다. 본문에는 병합 전 값 197,897 도 그대로 적혀 있다. `A_adm` 이 구조적 불가를 보이면 한도 달성에는 구속 조항 동결(REQ-ALD2-002·003)을 푸는 결정이 필요한데, 그 결정을 받을 카드는 아직 없다. 채무의 소유자는 있지만 청산 경로는 조건부로 열려 있는 셈이다. 확신도 중간. — Required fix: `t1226` 을 배차할 때 운영자 승인 아래 카드 본문을 병합 트리 값 198,361 로 고치고, 「결과가 구조적 불가이면 동결 완화 여부를 운영자에게 상신한다」를 범위에 넣는다.

blocking 결함: **0건.**

---

## Claim

1. F1·F2 의 끊긴 앵커는 두 트리와 codex 사본에서 모두 해소됐고, 수리가 새로 끊은 앵커는 없다.
2. 구속 조항 동결 다중집합은 `950fcc492` 에서 기준선과 바이트 동일하다(`d97b33d9…c6c3`, 170줄). 수리 diff 에는 구속 토큰을 가진 줄이 없다.
3. AC-ALD2-001 의 18경로 합계는 `950fcc492` 에서 198,361 이며, `acceptance.md §AC-ALD2-001.2` 기록값과 일치한다.
4. REQ-ALD2-001·제목·H1·AC-ALD2-001·`§D.2`·design.md R-70 사이에 모순이 없다. 다른 AC 의 판정 문언은 바뀌지 않았다.
5. 수리 대상 두 쌍의 미러 hunk 수는 수리 전과 같고, 카탈로그 해시는 템플릿 파일의 sha256 과 일치한다.
6. 파일당 40k 초과 목록은 두 트리 모두 같은 4개, 같은 크기로 유지된다.
7. 완료 정의의 테스트 세트와 SPEC lint 가 `950fcc492` 에서 통과한다.

## Evidence

**AC-ALD2-001 · AC-ALD2-002 축자 실행**(`acceptance.md` 30-45행과 190-205행을 `sed -n` 으로 그대로 떼어 실행. 떼어 낸 블록은 원문과 대조해 확인했다):

```
$ sed -n 30,45p acceptance.md > ac001.sh ; sed -n 190,205p acceptance.md > ac002.sh
$ bash ac001.sh
  198361 total
$ bash ac002.sh
d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3  -
$ grep -rhE '\[HARD\]|MUST|shall ' <같은 16경로> | wc -l
     170
$ git diff --stat 8e50ef148 950fcc492
 .../SPEC-ALWAYS-LOADED-DIET-002/acceptance.md      | 39 ++++++++++++++--------
 .moai/specs/SPEC-ALWAYS-LOADED-DIET-002/design.md  |  3 +-
 .moai/specs/SPEC-ALWAYS-LOADED-DIET-002/spec.md    | 22 +++++++-----
 internal/template/catalog.yaml                     |  2 +-
 4 files changed, 43 insertions(+), 23 deletions(-)
$ git diff -U0 805d44bed 950fcc492 -- .claude internal/template/templates CLAUDE.md AGENTS.md | grep -E '^[-+][^-+]' | grep -cE '\[HARD\]|MUST|shall '
0
```

**앵커 대상 제목**(두 트리):

```
.claude/rules/moai/core/agent-common-protocol-reference.md:316:### Skeptical Evaluation Stance
.claude/rules/moai/core/agent-common-protocol.md:89:## Language Handling
(dynamic-workflows.md) 17:## The Three Orchestration Primitives
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md:316:### Skeptical Evaluation Stance
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md:89:## Language Handling
(template dynamic-workflows.md) 17:## The Three Orchestration Primitives
```

**앵커 스윕 차집합**(`git archive` 로 `805d44bed` 와 `950fcc492` 를 스크래치에 풀고, 같은 파서로 「파일명 + § 제목」 인용을 전수 판독한 뒤 결과를 diff 했다. 파서 잡음이 섞인 절대 계수는 의미가 없고, 차집합만 판정에 쓴다):

```
citations 1732 broken 273   (805d44bed)
citations 1733 broken 271   (950fcc492)
== diff (< pre-repair only, > HEAD only)
42d41
< .claude/rules/moai/core/moai-constitution.md | moai-constitution-detail.md | Parallel Execution
170d168
< internal/template/templates/.claude/rules/moai/core/moai-constitution.md | moai-constitution-detail.md | Parallel Execution
```

`>` 행(HEAD 에만 있는 끊김)은 0건이다. F1 의 형태(「… §A … and §B」에서 둘째 § 앞에 파일명이 없는 경우)는 이 파서가 원리상 잡지 못하므로 직접 확인했다:

```
$ grep -n 'Language Handling' .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/template/templates/.codex/agents/moai/sync-auditor.toml
.claude/agents/moai/sync-auditor.md:31:> See `.claude/rules/moai/core/agent-common-protocol-reference.md` §Skeptical Evaluation Stance (…), and `.claude/rules/moai/core/agent-common-protocol.md` §Language Handling (…).
internal/template/templates/.codex/agents/moai/sync-auditor.toml:20:> See … (같은 문장)
internal/template/templates/.claude/agents/moai/sync-auditor.md:31:> See … (같은 문장)
$ grep -rn 'agent-common-protocol-reference.md[^|]*Language Handling' --include='*.md' --include='*.toml' . (reports·specs 제외)
→ 위 세 줄만 적중(수정된 줄 자신). 다른 파일의 잘못된 짝은 0건
```

**미러·카탈로그**:

```
moai-constitution.md live↔tpl hunk 수: 수리 전 1 → HEAD 1
sync-auditor.md live↔tpl hunk: 수리 전 {33,37d32 · 62,72d56 · 153c137} → HEAD 같은 세 hunk
$ shasum -a 256 internal/template/templates/.claude/agents/moai/sync-auditor.md
5beb50104c6d3257878ad5fb534e94ad7ae8cc13d85610d687cf6f7c0509012c
$ grep -n -A1 'path: templates/.claude/agents/moai/sync-auditor.md' internal/template/catalog.yaml
120-              hash: 5beb50104c6d3257878ad5fb534e94ad7ae8cc13d85610d687cf6f7c0509012c
$ grep -c 'moai-constitution' internal/template/catalog.yaml
0          (규칙 파일은 카탈로그 해시 대상이 아니다 — 추가 갱신 불필요)
```

**파일당 40k**(두 트리 같은 출력):

```
   61435 …/workflow/worktree-integration.md
   41034 …/workflow/kanban-dispatch-detail.md
   40797 …/workflow/spec-workflow.md
   41615 …/workflow/session-handoff-examples.md
```

**테스트·lint**(환경 스크럽 단일 호출, 출력은 스크래치 파일로 받아 꼬리만 인용):

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/spec/... ./internal/template/... ./internal/config/...
exit=0
ok  	github.com/modu-ai/moai-adk/internal/spec	116.241s
ok  	github.com/modu-ai/moai-adk/internal/template	66.451s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.337s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.085s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
ok  	github.com/modu-ai/moai-adk/internal/config	3.693s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.268s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.205s
$ go run ./cmd/moai spec lint .moai/specs/SPEC-ALWAYS-LOADED-DIET-002
exit=0
INFO      OwnershipTransitionUnmeasured  …/spec.md  1  … commit 5bfa134ca… (직전 감사와 같은 INFO)
0 error(s), 0 warning(s)
```

**SPEC 메타·AC 수·후속 카드**:

```
$ grep -c '^### AC-ALD2-' acceptance.md
9
spec.md: version: "0.10.0" · status: implemented
$ moai todo   (exit=0, 발췌)
t1226	queued	[리드 발행 09-26 · t1175 후속 · Tier M · 클래스 C] 상시 로드 템플릿 rules 풀에서 실제로 더 들어낼 수 있는 양(A_adm, 가용 풀)을 실측한다. t1175 는 합계를 197,897 자로 줄였고 …
t1191	queued	…
t1192	queued	…
```

## Baseline-attribution

- 모든 수치는 이번 감사 실행(2026-09-27, 18:4x–18:50 KST)에서 이 워크트리의 `950fcc492` 를 대상으로 직접 쟀다. 수리 전 비교 대상은 `git archive 805d44bed` 로 스크래치에 푼 트리다.
- AC 블록은 `acceptance.md` 원문의 줄 범위를 그대로 떼어 실행했다. 손으로 옮겨 적지 않았다.
- 오케스트레이터 측정(`remeasure/f4-ac001-ac002.md` 의 198,361, `remeasure/go-test-final.txt`)은 **인용하지 않고 재실행해 재현했다.** 두 값 모두 일치했다.
- `golangci-lint` 는 다시 돌리지 않았다. 수리 범위의 Go 파일 변경이 0개이므로(변경은 마크다운·TOML·YAML 뿐) lint 표면은 바뀌지 않는다. 이것은 판단이지 측정이 아니다.

## Gaps

- **교차 모델 2차 의견을 돌리지 않았다.** 이번 스폰에도 `mcp__moai__audit_multi`/`codex_audit`/`glm_audit` 가 노출되지 않았다.
- **앵커 스윕 파서는 「파일명 + § 제목」 형태만 본다.** F1 과 같은 연쇄 § 형태는 수리된 세 줄과 같은 짝 패턴 grep 으로만 확인했다. 연쇄 § 형태를 트리 전체에서 전수로 보지는 않았다.
- **`plan.md`·`progress.md` 전문을 대조하지 않았다.** `150,000`·`REQ-ALD2-001`·`197,897` 적중 줄과 재개 지점 절만 읽었다.
- **이 워크트리의 병합 창 이후 트리는 재지 않았다.** 로컬 develop 에 병합될 트리는 아직 없다.
- **`golangci-lint` 와 전체 테스트 스위트는 로컬에서 돌리지 않았다**(상시 금지 — 전 패키지 판정은 develop CI 몫).
- **F5~F8 의 내용은 다시 판독하지 않았다.** 수리 diff 가 해당 파일을 건드리지 않았다는 것만 확인했다.

## Residual-risk

- **기록값은 특정 ref 에 묶여 있다.** `§AC-ALD2-001.2` 의 198,361 은 `8e50ef148`(=`950fcc492` 와 계수 내용 동일)의 값이다. 병합 창 전에 develop 을 다시 흡수하면 이번처럼(+454) 합계가 다시 움직인다. 레인 규율대로 병합 트리에서 재측정해 이력 표에 한 행을 더하지 않으면, F4 와 같은 종류의 낡은 기록이 되풀이된다.
- **채무의 청산 경로가 조건부다(N2).** `t1226` 이 구조적 불가를 확립하면 150k 한도는 구속 조항 동결을 푸는 결정 없이는 닫히지 않는다. 그 결정을 받을 카드가 아직 없다.
- **도달 범위 축의 약화(직전 감사 F7·잔여 위험)는 이번 수리와 무관하게 그대로 남는다.** 해시는 이것을 원리상 볼 수 없다.
- **스윕 파서 잡음은 수작업으로 걸러졌다.** 차집합이 0이라는 판정은 수리 전후에 같은 파서를 썼다는 데 기대고 있으므로, 파서가 두 트리에서 똑같이 놓친 끊김은 이 방법으로 드러나지 않는다.

VERDICT: PASS-WITH-DEBT score=0.87
