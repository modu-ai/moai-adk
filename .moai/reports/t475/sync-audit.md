# sync-audit.md — SPEC-CODEMAPS-REFRESH-002 (카드 t475)

감사자: sync-auditor (독립 재측정) · 감사 일자: 2026-09-08
감사 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `f527f8f9f`, `git status --porcelain` 무출력
카드 base / merge-base: `52f863f36`
평가 프로필: `.moai/config/evaluator-profiles/default.md` (평면 가중 방식 — `harness.yaml`에 `evaluator_mode: hierarchical` 없음). `audit_model` 키 부재 → Claude 단독 감사, 교차 모델 백엔드 미호출.

---

## 판정

**PASS-WITH-DEBT · 종합 93 / 100** (조화평균 92.9 — 두 방식이 갈리지 않는다)

- **차단 findings: 0**
- **비차단 findings: 4** (F1 부채 · F2~F4 관측)
- must-pass 방화벽(Functionality · Security) 양쪽 모두 임계 통과

이 카드가 실제로 내놓은 산출물 — 6개 codemaps 문서, 재스탬프된 `provenance.json`, 증거 파일 — 은 내가 재측정한 범위 전부에서 정확하다. 부채는 산출물이 아니라 **증거 파일 §④의 표시 상태**에 있다.

---

## 차원별 점수

| 차원 | 점수 | 판정 | 증거 (verbatim) |
|---|---|---|---|
| Functionality (40%) | 93 | PASS | AC 13항목 전수 독립 재측정 — 11 PASS / 2 regression-guard / 0 FAIL. 후보 20 재유도 `Zero=48 A=5 B=14` (+C 1), 판정 행 20, `comm` 양방향 무출력. 인용 경로 `unique normalized: 175 / absent: 0`. 히트-0 `48 → 39` 원소 단위 일치. 식별자 10행 · 좌표 10/10 해석. omission 15/15 ≥1 · fold 5/5 = 0 |
| Security (25%) | 100 | PASS | Go 프로덕션 코드 diff 0줄, `.moai/config/` 변경 0건(`git diff --name-only 52f863f36 HEAD -- .moai/config/` 무출력), 비밀정보 스캔 적중 0(산문 내 "token" 오탐 1건만), 의존성 무변경. Critical/High 0 |
| Craft (20%) | 85 | PASS (커버리지 N/A) | Go 0줄이라 85% 커버리지 임계에 측정면이 없다(REQ-CM2-012) — undecidable disposition 적용, FAIL 아님. 재측정한 수치 20여 개가 전부 재현. 감점 3건: F1·F2·F3 |
| Consistency (15%) | 95 | PASS | 변경 27경로 전부 허용 접두사 4개(`.moai/project/codemaps/`·`.moai/reports/t475/`·`.moai/specs/SPEC-CODEMAPS-REFRESH-002/`·`CHANGELOG.md`) 안. 커밋 7건 전부 Conventional + 카드 id `t475` 운반. 브랜치 `WT-codemaps-stale`(슬러그형, 카드 id 미포함) |

가중 계산: `93×0.40 + 100×0.25 + 85×0.20 + 95×0.15 = 93.45`

---

## Claim / Evidence / Baseline-attribution

기준선 귀속: 아래 모든 측정은 **이 감사 실행에서, 이 트리(HEAD `f527f8f9f`)에 대해** 직접 실행한 명령의 출력이다. run/sync가 기록한 수치를 인용한 것이 아니라 재실행한 값이며, 그래서 일치가 곧 독립 확증이다.

### 게이트 (AC-CM2-010)

```
$ ./bin/moai graph check ; echo EXIT=$?
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
EXIT=1
```

stale 계층 0개. 종료 코드 1의 원인은 `absent` 두 계층이며, `internal/graph/check.go:142-149`를 직접 읽어 확인했다 — `Failed()`가 `l.Verdict != VerdictFresh`를 전부 실패로 센다. AC-CM2-010이 명시적으로 합격 저해 요인이 아니라고 규정한다.

### 후보 규칙 재유도 (AC-CM2-002)

SPEC §A.3(a)의 A층·B층 명령을 **재생성 전 사본**(`pre-regen/`)을 우주로 삼아 그대로 실행:

```
=== universe=pre : Zero=48 A=5 B=14 ===
=== universe=post : Zero=39 A=1 B=7 ===
```

A층 5개(`core/git`·`settings/yamlpatch`·`stateanchor`·`template/agentemit`·`template/commandemit`) + B층 14개 + C층 `internal/chain` = **20**, `candidates.txt`와 원소 단위 일치(`comm -23`/`comm -13` 양방향 무출력). 판정 행 수 `grep -cE '^\| \`[^\`]+\` \| (fold|omission) \|'` → **20**.

인용한 부모 히트 수도 전부 재현: `internal` 236 · `internal/cli` 64 · `internal/template` 19 · `internal/kanban` 10 · `internal/statusline` 7 · `internal/web` 5 · `internal/harness` 4 · `internal/settings` 3.

15개 인용 좌표 전수를 `sed -n '<n>p' … | grep -qF '<인용문>'`로 재확인 — **15/15 exit 0**.

### 편입과 접힘 규율 (AC-CM2-004 / §A.3(a1))

```
omission 15개 → grep -rl -F 히트 파일 수
internal/settings/yamlpatch 3 · internal/stateanchor 4 · internal/template/agentemit 4
internal/template/commandemit 5 · internal/chain 4 · codex_skills_disable.go 1
codex_skills_prune.go 1 · integration_settings_drift.go 3 · skills.go 1
update_mirror_heal.go 2 · kanban/settings_drift.go 2 · statusline/state_anchor.go 2
published_skills.go 2 · skill_mirror_repair.go 2 · web/codexmirror.go 2

fold 5개 → 전부 0
internal/core/git 0 · doctor_hook_delivery.go 0 · step_git_env.go 0
prlink_landedref.go 0 · fieldsets_codex_templ.go 0
```

**독립 확증의 결정적 형태**: 재생성 **후** 트리에서 후보 규칙을 다시 돌렸을 때 히트-0으로 남은 후보 집합이 fold 5개와 **원소 단위로 정확히 같다**(commandemit 흡수 해제로 재부상한 3개 생성 파일 제외). 즉 되돌림은 실재하고 완전하며, 같은 문단의 omission 서술은 손상되지 않았다.

### 정확성 층 (AC-CM2-006 / 007 / 008)

`check_citations.go`의 정본 규약 3요소 + `normalizeCitedPath` 나머지 규칙을 **독립 재구현**해 실행:

```
raw tokens: 373
unique normalized: 175
absent: 0
```

커밋된 `cited-paths-table.txt`(175행, `absent` 0회)와 **집합 동일** — `comm` 양방향 무출력. 게이트 `citations` 계층 `value=0`과도 일치.

히트-0 재실행 39개는 §⑥ 표의 열거와 원소 단위 일치. 히트를 얻은 9개도 재현(`chain`·`taskledger`·`git/convention`·`hook/testutil`·`lsp/aggregator`·`settings/yamlpatch`·`stateanchor`·`agentemit`·`commandemit`).

식별자 추출 10행(0행 아님 — 공허한 통과 아님), 10개 좌표 전부 `sed -n '<n>p'`로 직접 확인. 예: `internal/hook/types.go:18` → `type EventType string`, `internal/cli/mcp_server.go:158` → `add("session_list", mcp.NewTool(`.

### 재스탬프 (AC-CM2-009)

```
$ cat .moai/project/codemaps/provenance.json
  "commit_sha": "52f863f3666c9ec754253a06b96ed1fe844f1590"
  "tree_root":  ".../worktrees/t475"
$ git merge-base --is-ancestor 52f863f36… origin/develop ; echo $?
0
```

`provenance.json`이 기록한 값이 merge-base이며 `origin/develop`의 조상이다.

### 범위 위생 (AC-CM2-011)

`git diff --name-only 52f863f36 HEAD`의 27경로를 허용 접두사 필터로 걸렀을 때 잔여 0. Go·config·`gate.yaml` 변경 0건. REQ-CM2-011의 `Where` 전제 좌표도 재확인 — `gate.yaml:72-78`에 `enabled: true` / `blocking: false` / `codemaps_changed_files: 40`.

### 관측 리포트 (AC-CM2-012)

3항목 전부 존재. ③의 42개 목록은 내가 계산한 잔여집합(Zero_pre 48 − A층 5 − C층 1)과 **원소 단위로 정확히 일치**(차이는 뒤 문단의 글롭 표기 `internal/harness/` 3건뿐). ①은 앵커 귀속(t476 / `2026-09-03T18:18:34Z`)과 5일 누적을 담는다. `tree_root`는 "보고하지 않는 것" 절에서 비항목으로 명시 — 항목으로 올리지 않았으므로 AC의 금지에 걸리지 않는다.

---

## 리드가 제시한 4개 flag에 대한 판정

### Flag 1 — AC-CM2-003의 실행면 → **PASS** (전제 정정 포함)

두 가지를 분리해야 한다.

**(a) 인용된 AC 문구는 폐기본이다.** 리드가 인용한 `6개 문서가 전부 재생성 대상으로 보고된다`는 문자열은 **현재 트리 어디에도 없다**(`/usr/bin/grep -rn "6개 문서가 전부" .moai/` → exit 1). 그것은 v0.1.0 문구이며 plan-audit iter-1 D3이 잡아 v0.1.1에서 수리됐다. 현행 `acceptance.md:146`은 다음과 같다:

> **Then** `overview.md` / `modules.md` / `dependencies.md` / `entry-points.md` / `data-flow.md` **5개**가 재생성 대상으로 보고되고, `ls .moai/project/codemaps/`가 7항목(문서 6 + provenance.json)을 낸다.

따라서 "6문서 재생성이 애초에 달성 불가능했다"는 얽힘은 **이미 닫힌 것**이며, 이 카드가 물려받은 결함이 아니라 이 카드가 REQ-CM2-014/AC-CM2-003a로 **닫은** 결함이다.

**(b) 스킬을 호출하지 않은 것이 AC 위반인가 — 아니다.** 실행면을 직접 읽었다. `.claude/skills/moai/workflows/codemaps.md`가 규정하는 것은 이것이다:

> The orchestrator generates the maps directly (no Agent() spawn) from the Phase 2 analysis — replacing the former generation delegation spawn.

즉 `/moai codemaps --force`의 실행면 자체가 **오케스트레이터가 직접 문서를 쓰는 것**이다. 별도의 기계 생성기는 존재하지 않는다 — `moai graph`의 서브커맨드는 `build`/`query`/`check`/`stamp`/`report` 다섯뿐이고, `codemaps-gen`이라는 문자열은 `internal/graph/*_test.go`의 픽스처 상수로만 나온다(비테스트 코드 적중 0). 그러므로 "스킬을 호출했는가"와 "스킬이 규정한 워크플로를 수행했는가"는 산출물에 대해 구분되는 두 상태가 아니다. 스킬 본문이 산출 5개를 열거하고, 실행자가 그 5개를 현재 트리에서 다시 썼으며, `ls`가 7항목을 낸다 — Then 절이 요구하는 것 전부가 성립한다.

**남는 것은 좁은 Gap 하나**: 슬래시 디스패처 경로를 실제로 태웠을 때 **다른** 문서 집합이 나오는지는 관측되지 않았다. 실행자가 이를 스스로 Gap으로 신고한 것은 옳은 처분이며, 그 미관측이 AC의 Then을 무너뜨리지는 않는다. **PASS.**

### Flag 2 — fold 규율 간극 → **(a) 되돌림 실재·완전 확인 · (b) 간극 주장 참**

**(a) 되돌림은 실재하고 완전하다.** 위 §"편입과 접힘 규율"의 두 측정이 이를 독립적으로 세운다: fold 5개 전부 히트 0, omission 15개 전부 ≥1, 그리고 재생성 후 후보 규칙이 뱉는 히트-0 잔여가 fold 5개와 **원소 단위로 정확히 같다**. 세 번째가 특히 강한 증거다 — 되돌림이 부분적이었다면 이 집합이 fold 집합보다 작았을 것이고, 과도했다면 omission이 섞였을 것이다. 어느 쪽도 아니다.

**(b) 간극 주장은 참이다.** AC 전수를 읽어 확인했다. `AC-CM2-004`는 omission만 구속하고, `AC-CM2-007`은 "omission 판정 단위가 이 목록에 남아 있으면 FAIL"만 규정하며 **fold가 목록에 남아 있을 것을 요구하지 않는다**. `AC-CM2-002`는 판정표만 본다. 따라서 fold 단위가 산문을 얻어도 FAIL하는 AC가 없다 — 되돌리지 않았어도 13개 AC가 전부 통과했을 것이라는 실행자의 진술은 정확하다.

이 간극을 잡은 것은 AC가 아니라 실행자가 스스로 세운 불변 조건(progress.md §E.3 "fold 판정 단위 산문 무변경")이었다. **후속 카드 재료로 반드시 이관되어야 한다** — 이 카드에서 AC를 사후 편집해 간극을 닫지 않은 것은 옳은 처분이다(사후 편집은 감사를 무의미하게 만든다).

### Flag 3 — AC-id 정규식 → **실재하는 결함 계열 · 이 카드는 그 위에 서 있지 않다**

재현했다:

```
canonical  AC-([A-Z0-9]+-)*[0-9]+     → 12   (AC-CM2-003a 가 AC-CM2-003 으로 붕괴)
suffixed   AC-([A-Z0-9]+-)*[0-9]+[a-z]? → 13
progress.md AC 매트릭스 행 수           → 13
```

**결함은 실재하며 배포된다.** 이 정규식은 세 곳에 산다:
- `.claude/agents/moai/plan-auditor.md:228` (로컬)
- `internal/template/templates/.claude/agents/moai/plan-auditor.md:228` (배포 템플릿)
- `internal/template/templates/.codex/agents/moai/plan-auditor.toml:224` (방출본)

즉 접미 문자를 쓰는 모든 SPEC에서 plan-auditor의 AC 전수 조사가 조용히 1개씩(또는 그 이상) 적게 센다. 이 카드의 소관 밖이며(허용 3경로 + CHANGELOG 밖), open item으로 이관된 것이 옳다.

**이 카드가 잘못된 수 위에 서 있지는 않다.** 실행자는 13을 채택했고, CHANGELOG도 `13 criteria — 11 PASS, 2 regression-guard, 0 FAIL`로 13을 싣는다. 12를 썼다면 AC 하나가 CHANGELOG에서 누락된 채 통과했으리라는 sync의 진술은 정확하다.

**형제 REQ 건은 부분 확증에 그친다.** `internal/spec/lint.go:604`의 `reqIDPattern`은 접미 문자를 허용하지 않아 `REQ-CM2-003a`를 거절하는 것이 맞다. 다만 그 거절은 **조용하지 않다** — `:844`가 `InvalidREQID` finding을 발행한다. "조용히 무시한다"는 서술이 성립하려면 추출 층(`reqLinePattern:607` 등)에서 `doc.REQs`에 애초에 실리지 않아야 하는데, **나는 추출 경로를 끝까지 추적하지 않았다**(Gap). 검증 층에 한해서는 침묵이 아니다.

### Flag 4 — 3건의 직전 판 정정 → **셋 다 옳다 · 새 거짓 주장 미검출**

각각을 문서가 명시한 명령으로 다시 쟀다.

| 정정 | 문서 주장 | 내 재측정 | 판정 |
|---|---|---|---|
| 패키지 단위 엣지 | `1638` 미재현, 이 판 `345` | `345` (dedup 후에도 345), 테스트 import 포함 시 `408` | **확증** |
| 최대 비테스트 파일 | `session_start.go` 61KB, 5위, 상위 둘은 templ 생성 | `168KB fieldsets_templ.go` · `121KB screens_templ.go` · `89KB mcp_codex.go` · `75KB config/types.go` · `61KB session_start.go` | **확증** |
| `docs-truth.md` §4.1 | 렌더 그룹 4개, 직전 판의 손 5분류는 부재, `skills` 누락 | `./bin/moai --help` 그룹 헤더 정확히 4개(`COMMANDS`/`LAUNCH COMMANDS`/`PROJECT COMMANDS`/`TOOLS`), TOOLS 33개 verb가 문서 표와 **순서까지 동일**, `skills`는 `internal/cli/root.go:177` `rootCmd.AddCommand(newSkillsCmd())`로 등록된 라이브 루트 명령 | **확증** |

**정정이 새 거짓을 들여왔는가 — 검출되지 않았다.** 정정과 함께 재작성된 문서의 다른 수치를 표본이 아니라 가능한 한 전수로 다시 쟀고, 전부 재현했다:

```
Go 패키지 총수 139 → 139        임베드 템플릿 581 → 581
internal 최상위 65 → 65         테스트 1812 / 비테스트 1114 → 비 1.63:1
AddCommand 비테스트 207 → 207   rootCmd.AddCommand 62 → 62
internal/cli 비테스트 파일 279 → 279   .claude/commands/moai/*.md 16 → 16
.claude/agents/moai/*.md 11 → 11
internal/git 루트 비테스트 importer 0 → 0, convention 1 → 1
```

팬텀 `cmd/templ` 수리도 확증된다 — 인용 경로 175개 중 absent 0이며, 내 독립 재구현도 같은 답을 낸다.

---

## Findings (구조화 결함 목록)

### F1 [Medium] [optional·부채] `.moai/reports/t475/codemaps-accuracy-verification.md:341-546` — §④의 diff 증거가 되돌림 **이전** 캡처이며 그 사실이 표기돼 있지 않다

§④가 표시하는 `diff -u` 출력은 HEAD에서 재현되지 않는다. 정량화했다.

**(1) 문서 전체 diff 규모 표(:361-369)** — 실측과 어긋난다:

| 문서 | §④ 표기 | HEAD 재측정 |
|---|---|---|
| overview.md | 115 | 114 |
| modules.md | 282 | 280 |
| dependencies.md | 169 | 168 |
| entry-points.md | **135** | **104** |
| data-flow.md | **198** | **193** |
| docs-truth.md | 120 | 120 |

**(2) AREA 블록이 표시하는 추가 행 75개 중 8개가 배포 문서에 없다.** 8개 전부 fold 단위 산문이며, §④-b가 되돌렸다고 적은 바로 그것들이다:

```
NOT-IN-FINAL: | `doctor*` | 15 | 진단 — … doctor_hook_delivery.go …
NOT-IN-FINAL: `moai doctor`의 hook delivery 점검(`internal/cli/doctor_hook_delivery.go`)이 …
NOT-IN-FINAL: 렌더 쪽 `internal/web/fieldsets_codex_templ.go`는 `a-h/templ`이 …   (외 3행)
NOT-IN-FINAL: | `internal/kanban` | 35 | … PR 링크(착지 ref 3단 해석 사슬 `prlink…
```

확증: `/usr/bin/grep -rn 'doctor_hook_delivery' .moai/project/codemaps/` → **exit 1**(배포 문서에 없음). 그런데 같은 문자열이 §④:387·:411에는 `+` 행으로 실려 있다.

**왜 문제인가**: 증거 파일이 표시하는 명령 출력을 독자가 배포 트리에서 재실행하면 다른 답이 나온다. 이 저장소 자신의 `verification-claim-integrity.md` §2가 요구하는 baseline 귀속(어느 트리에서 잰 값인가)이 §④에는 없다.

**왜 차단이 아닌가**: ① AC-CM2-005는 "구간마다 행이 존재하고 그 행이 `diff -u` 출력을 담을 것"을 요구하며, 캡처 시점에는 진짜 출력이었다 — 조작이 아니라 **표기 누락**이다. ② §④-b(:548-574)가 fold 5개 각각에 대해 무엇을 넣었다 지웠는지 표로 명시하므로, §④ → §④-b를 순서대로 읽은 독자는 진상을 복원한다. 실제로 내가 검출한 8행이 그 표와 1:1로 대응한다. ③ 배포된 문서 자체는 옳다.

**필요한 수리(리드 판단)**: §④ 머리에 한 줄 — "이 절의 diff는 §④-b의 되돌림 **이전** 트리(`e397ec00d` 계열)에서 캡처했다. HEAD에서는 fold 단위 8행이 빠진 더 작은 출력이 나온다." 문서 본문은 손대지 않는다.

### F2 [Low] [optional] `acceptance.md:255` / `progress.md §E.3` — AC-CM2-009의 regression-guard 처분이 실측을 과소평가한다

acceptance.md의 근거는 "현재는 `merge-base`와 `HEAD`가 같아(둘 다 `52f863f36`) 이 AC가 물지 않는다"이며, 이는 **저작 시점**에는 참이었다. 그러나 M4 실행 시점의 HEAD는 `cd03be0d3`으로 merge-base보다 3커밋 앞서 있었고, 실측상 조상이 아니다:

```
$ git merge-base --is-ancestor cd03be0d3 origin/develop ; echo $?
1
$ git log --oneline 52f863f36..cd03be0d3 | wc -l
3
```

즉 bare HEAD를 스탬프했다면 AC-CM2-009가 실제로 **떨어뜨렸을** 것이다 — 이 AC는 실행 시점에 판별력을 가졌다. regression-guard로 기록한 것은 **보수적 방향**이므로(얻지 않은 통과를 주장하지 않았다) 거짓 초록을 만들지 않는다. 다만 얻은 증거를 과소 보고했다.

### F3 [Low] [optional] `spec.md:376` — §D 요약표가 AC-CM2-012를 "관측 리포트 **2항목**"으로 적는다

`acceptance.md` AC-CM2-012와 `REQ-CM2-013`은 **3항목**을 요구하며, §A.3(c)·`REQ-CM2-013 ③`은 세 번째(잔여 42 목록)의 부재가 FAIL임을 두 곳에서 못박는다. §D 요약표 행만 `2항목`으로 남아 있다 — v0.1.4가 ③을 비선택으로 복원할 때 갱신되지 않은 잔재로 보인다.

결과에는 영향이 없다(실행자가 3항목을 냈고, 판정면은 acceptance.md다). 그러나 요약표만 읽는 독자는 요구를 과소 산정한다.

### F4 [Medium] [optional·범위 밖·이관] `plan-auditor.md:228` (3사본) — AC-id 정규식이 접미 문자 id를 흡수한다

Flag 3 판정 참조. 배포 템플릿과 codex 방출본에도 실려 있어 사용자 프로젝트 전체에 미친다. 이 카드의 허용 경로 밖이므로 여기서 고칠 수 없고, open item으로 이관된 것이 옳다. 리드가 후속 카드를 정할 재료다.

---

## Gaps (관측하지 않은 것)

- **Go 테스트 스위트 미실행.** 지시에 따른 것이며 Go diff가 0줄이라 측정면이 없다. 커버리지 85% 임계도 같은 이유로 판정하지 않았다(N/A, FAIL 아님).
- **CI 판정 부재.** 미푸시 7커밋, PR 없음 — CI 실행이 존재하지 않는다. 초록도 빨강도 관측하지 않았다.
- **6문서 산문 전수 사실 검증은 하지 않았다.** diff 4257행 중 내가 재측정한 것은 수치 주장 약 20건과 인용 경로 175개(존재만), 인용 식별자 10개(좌표 해석)다. 서술적 문장(책임 기술·구조 해설)의 사실성은 표본조차 아니라 **미검증**이다.
- **인용 경로는 존재만 확인했다.** AC-CM2-006이 요구하는 것이 존재이므로 이는 요건 충족이지만, "그 경로가 그 문맥에서 옳게 인용됐는가"는 다른 질문이며 답하지 않았다.
- **`reqLinePattern` 추출 경로 미추적** (F4 형제 건). 검증 층의 `InvalidREQID` 발행만 확인했다.
- **§④ 캡처 시점을 특정하지 않았다.** HEAD에서 재현되지 않는다는 것만 세웠고, `e397ec00d`에서 정확히 재현되는지는 재지 않았다.
- **`/moai codemaps --force`의 디스패처 경로 미실행** (Flag 1(b)의 잔여 Gap). 스킬 본문이 산출 5개를 선언한다는 것과 기계 생성기가 부재한다는 것만 확인했다.
- **pre-regen 사본이 진짜 재생성 전 상태인지는 원리상 재확인 불가.** 사본은 M1 커밋(`295fd8f11`)에 실려 있고 재생성은 M2(`e397ec00d`)이므로 커밋 그래프가 순서를 증언한다 — 이것이 얻을 수 있는 최선의 증거이며, 사본 내용 자체를 앵커 트리와 대조하지는 않았다.

---

## Residual-risk (관측했음에도 틀릴 수 있는 것)

1. **문서 산문의 사실성.** 수치는 전부 재현됐지만 codemaps의 가치는 서술에 있다. 20개 단위의 "책임" 기술이 코드의 실제 책임과 어긋나 있을 가능성은 이 감사가 배제하지 못한다 — 인용 좌표가 해석된다는 것은 인용이 존재한다는 뜻이지 서술이 옳다는 뜻이 아니다.
2. **fold 판정 5건의 질적 타당성.** 부모 산문 인용이 실재함은 확인했으나, 그 인용문이 정말 그 단위의 책임을 "담는가"는 판단이다. 특히 `internal/hook/quality/step_git_env.go`의 근거(`data-flow.md:L58` "린터·포매터·게이트 요약")는 인용문이 짧아 포섭 범위가 넓게 읽힌다.
3. **`internal/core/git`의 fold 근거가 경로 표기 차이에 의존한다.** 근거는 부모가 `core/git`(접두어 없는 형태)로 기술한다는 것인데, 그렇다면 히트-0은 서술 공백이 아니라 표기 규약의 산물이다. 이 판정은 옳아 보이지만, 같은 논리가 다른 단위에도 적용되면 히트 기반 후보 규칙 자체의 정밀도 문제로 번진다 — 이 카드의 결함은 아니고 규칙의 성질이다.
4. **재스탬프가 붉음을 지우지만 낡음을 지우지는 않는다.** 게이트 `value=0`은 앵커가 `52f863f36`으로 옮겨진 결과이며, 이 카드가 정확성 층을 독립으로 세운 것이 바로 그 때문이다. 다음 앵커까지 같은 5일 누적이 반복되면 값은 다시 40을 넘는다 — verdict.md 관측 ①이 이를 리드에게 이관했고, 이 카드는 아무것도 바꾸지 않았다.
5. **§④ 부채가 후속 감사를 오도할 수 있다.** F1을 표기하지 않으면 다음 독자가 §④를 배포 상태의 서술로 읽고, 이미 되돌려진 fold 산문이 문서에 있다고 믿을 수 있다.
6. **F4가 이 감사에도 적용된다.** 나 역시 AC 열거에 접미사 허용 정규식을 썼기에 13을 얻었다. 정본 규약을 그대로 따랐다면 12를 세고 AC-CM2-003a를 통째로 놓쳤을 것이다.

---

## Recommendations (우선순위 순)

1. **F1 — §④ 머리에 캡처 시점 한 줄 추가.** 유일한 실질 부채이며 수리 비용이 한 줄이다. 리드가 병합 전 필수로 볼지 후속으로 넘길지 정한다.
2. **F4 — 후속 카드 발주.** `AC-([A-Z0-9]+-)*[0-9]+` 세 사본(로컬·템플릿·codex 방출본)에 `[a-z]?` 추가. 배포 표면이므로 `make agents-emit` 동반 필요.
3. **fold 규율 AC 신설 — 후속 카드.** Flag 2(b)가 참으로 확정됐으므로, "fold 판정 단위가 재생성 후에도 히트 0인가"를 직접 묻는 AC를 codemaps 계열 SPEC 템플릿에 넣는다. 이번에는 실행자의 자율 규율이 잡았고, 다음에는 잡히지 않을 수 있다.
4. **F2·F3 — 기록만.** 결과에 영향이 없고 사후 편집이 오히려 감사 이력을 흐린다.
5. **잔여 42개 처분 — 후속 카드.** verdict.md 관측 ③이 이관 조건을 충족했다. `internal/harness/*` 12 · `internal/lsp/*` 5 · `internal/navigator/*` 5가 세 부모에 몰려 있어 접힘 정책 질문 3회로 22개가 정리될 개연성이 있다(개연성이지 판정이 아니다).

---

## 병합 관점 결론

차단 결함 0건. 카드가 선언한 것 — 게이트 종결, 20개 전수 판정, 15개 편입, 3항목 정확성 검증, merge-base 재스탬프, 관측 3항목 — 전부가 내가 이 트리에서 다시 실행한 명령의 출력으로 확증된다. 범위 위생은 깨끗하고, Go·설정·임계값은 손대지 않았다.

부채 F1은 배포 산출물이 아니라 증거 파일의 표기에 있으며, 같은 파일의 §④-b가 그 정보를 이미 담고 있다. 병합을 막을 근거로 보지 않는다 — 처분은 리드 몫이다.
