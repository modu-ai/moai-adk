# SPEC-EVIDENCE-PATH-EXCEPTION-001 — 측정 기록

> Tier L 산출물. **이 SPEC이 인용하는 수치의 정본**이다. 명령과 관측 출력을 짝지어 적고,
> 실패할 수 있는 양성 대조를 함께 적는다. `spec.md`의 각 절은 여기를 가리킨다.
>
> [HARD] 이 문서는 `status:` frontmatter를 담지 않는다(`spec-frontmatter-schema.md`
> § Artifact Statelessness).

## §0 귀속 (baseline-attribution)

| 축 | 값 |
|---|---|
| 측정 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1039` |
| 브랜치 | `WT-evidence-path` |
| HEAD | `116820f40` |
| `.gitignore` 줄수 | **417** |
| 보조 트리(폭 수치 전용) | primary 체크아웃 `/Users/goos/MoAI/moai-adk-go` (`main` 체크아웃, `.gitignore` **319**줄) |
| 측정 시점 | 2026-09-20, iteration 2 |

[HARD] **두 트리는 서로 다른 답을 낸다.** 줄번호·줄수 인용은 트리를 명명하고, 폭 수치
(`find` 계열)는 primary에서만 의미가 있다 — 이 워크트리의 `.moai/reports`에는 54개 파일만 있다.

[HARD] **primary 체크아웃 수치는 다른 세션이 쓰는 트리의 스냅샷이며 드리프트한다.**
레인이 304를 쟀고 iter1 감사가 305를 쟀다 — 불일치가 아니라 드리프트다. 어떤 인수조건도 이
수치들에 대해 정확-일치를 단언하지 않는다.

---

## §1 `.gitignore` 규칙 지형 — `.moai/reports`를 건드리는 전량

### 명령과 출력

```
$ grep -n 'moai/reports' .gitignore
107:# Card evidence under .moai/reports/** is durable audit evidence, not operational
109:!.moai/reports/**/*.log
227:.moai/reports/*
228:!.moai/reports/plan-audit/
229:.moai/reports/plan-audit/*
230:!.moai/reports/plan-audit/.gitkeep
234:!.moai/reports/t338/
235:.moai/reports/t338/*
236:!.moai/reports/t338/ac-count-baseline.txt
237:!.moai/reports/t528/
238:.moai/reports/t528/*
239:!.moai/reports/t528/probe/
240:.moai/reports/t528/probe/*
241:!.moai/reports/t528/probe/nondecl-bullets.txt
242:!.moai/reports/t530/
243:.moai/reports/t530/*
244:!.moai/reports/t530/count-literals.txt
245:!.moai/reports/t229/
246:.moai/reports/t229/*
247:!.moai/reports/t229/live-probe-body.txt
294:.moai/reports/plan-audit/*.md
295:!.moai/reports/plan-audit/.gitkeep
376:.moai/reports/*.md
```

**23행.** 양성 대조로도 쓰인다 — 0보다 크므로 계측기가 파일을 읽고 있다.

### 주석 좌표 (본문에 `moai/reports`가 없어 위 grep에 안 나오는 것 포함)

```
$ sed -n '222,233p' .gitignore
(222) **/.mink/auth/
(223)
(224) # Card/audit reports are local-only artifacts (operator directive 2026-09-14):
(225) # evidence for lead verdicts stays on disk, never on the remote. Only the
(226) # plan-audit scaffold ships, mirroring the template's reports philosophy.
(227) .moai/reports/*
...
(231)
(232) # Narrow exceptions: test guard fixtures under reports stay tracked — CI reads
(233) # them (operator-approved 2026-09-14). Reports themselves remain local-only.
```

**블랭킷 주석은 `:224-226`이다** — `:227`은 블랭킷 규칙, `:228`은 negation이며 둘 다 주석이
아니다. (iter1 `spec.md`가 `:225-228`로 적었던 것의 정정.)

### 무리 분류 — `spec.md` §1.5의 표가 여기서 나왔다

| 무리 | 줄 | 성격 |
|---|---|---|
| A | 107-108 주석 / 109 규칙 | 죽은 negation과 살아 있다고 읽히는 주석 |
| B | 224-226 주석 / 227 / 228-230 | 블랭킷 + plan-audit 스캐폴드 카브아웃 |
| C | 232-233 주석 / 234-236, 237-241, 242-244, 245-247 | 3단 관용구 4벌 (t338 / t528 / t530 / t229) |
| D | 290-293 주석 / 294-295 | B의 **중복** 카브아웃 |
| E | 376 | 깊이 1 전용 |

관용구 줄 범위는 위 grep 출력에서 직접 읽은 것이다: t338 `234-236`, t528 `237-241`(중첩
디렉터리라 5줄), t530 `242-244`, t229 `245-247`. **iter1 `spec.md`(`:233-236`, `:237-240`,
`:243-245`, `:246-248`)와 레인 기록은 셋 이상이 한 칸씩 어긋나 있었다.**

---

## §2 어느 규칙이 무엇을 결정하는가 — `-v` 측정

[HARD] `-v`는 **판정**이 아니라 **귀속**을 읽는 용도다(negation 적중에도 exit 0).

```
$ git check-ignore -v --no-index .moai/reports/plan-audit/verdict.md
.gitignore:294:.moai/reports/plan-audit/*.md	.moai/reports/plan-audit/verdict.md

$ git check-ignore -v --no-index .moai/reports/tZZZ/verdict.md
.gitignore:227:.moai/reports/*	.moai/reports/tZZZ/verdict.md

$ git check-ignore -v --no-index .moai/reports/top.md
.gitignore:376:.moai/reports/*.md	.moai/reports/top.md

$ git check-ignore -v --no-index .moai/reports/t1039/plan-audit.md
.gitignore:227:.moai/reports/*	.moai/reports/t1039/plan-audit.md

$ git check-ignore -v --no-index .moai/reports/t338/ac-count-baseline.txt
.gitignore:236:!.moai/reports/t338/ac-count-baseline.txt	.moai/reports/t338/ac-count-baseline.txt
```

세 가지 결론:

1. **무리 D(`:294`)가 `plan-audit/verdict.md`를 결정한다** — 후보 삽입 지점(`:227` 직후)보다
   **아래**에 있다. `AC-EPE-004`가 재는 바로 그 경로다.
2. **무리 E(`:376`)는 깊이 1 전용이다** — `*`가 `/`를 넘지 않으므로 `<card>/verdict.md`에
   닿지 않는다. 깊이 2 프로브는 `:227`이 결정한다.
3. **무리 C는 발화한다** — 마지막 프로브가 negation 줄을 이름 붙인다. 계측기가 두 방향을
   모두 내므로 실패할 수 있다.

### 죽은 negation

```
$ git check-ignore --no-index -q .moai/reports/tZZZ/probe.log ; echo $?
0     # 무시됨 — :109는 발화하지 않았다
```

양성 대조: 위의 `t338/ac-count-baseline.txt`(negation이 실제로 발화하는 경로).
음성 대조: `README.md` → `rc=1`.

---

## §3 이 카드 자신의 실사례 — 감사 판정서가 반출되지 않았다

```
$ git status --porcelain --ignored -- .moai/reports/t1039
!! .moai/reports/t1039/

$ git status --porcelain -- .moai/reports/t1039
                                    # 0행 — --ignored 없이는 보이지 않는다

$ ls -1 .moai/reports/t1039/
lane-measurements.md
plan-audit.md
```

`plan-auditor.md:601`의 [HARD] Export mandate("an audit is complete only when its verdict is
exported")가 **이 SPEC을 심사한 감사에 대해 거짓**이다. 감사자는 규정된 목적지에 규정대로
썼고, 거짓인 것은 그 목적지가 추적된다는 **전제**다.

[HARD] **두 번째 명령의 0행은 부재가 아니라 미측정이다.** `--ignored` 없는 `git status`는
무시된 경로에 대해 구조상 침묵한다. 이 침묵을 「문제 없음」으로 읽는 것이 이 결함이 오래
살아남은 경로 중 하나다.

---

## §4 문서 표면 인벤토리

### 4.1 저장소 산문 — 리터럴 주장

```
$ grep -rlnE 'citation target is a (\*\*)?tracked(\*\*)? path' --include='*.md' --include='*.toml' .
.moai/specs/SPEC-HIERARCHICAL-TEAM-001/progress.md      ← 완료 SPEC의 진행 기록, 수정 대상 아님
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
.claude/rules/moai/core/agent-common-protocol-reference.md
.claude/rules/moai/core/agent-common-protocol.md
```

양성 대조(같은 정규식, 존재하지 않는 토큰): `0`.

**5건 사전 존재 / 수정 표면 4건.** 리터럴 grep은 `reference.md`의 볼드 표기
(`a **tracked** path`)를 놓치므로 볼드 허용 정규식이 필요하다 — 레인의 리터럴 측정이 3건만
보고한 이유다.

**자기-적중 주의**: 지금 같은 명령을 돌리면 **6행**이다. 6번째가 이 SPEC 자신의 `spec.md`이며,
§1.3이 이 정규식을 인용하기 때문이다. 자기-적중은 수정 표면이 아니다.

### 4.2 도달 표면 — 낱말 grep으로는 절반만 보인다

```
$ grep -n 'tracked' .claude/agents/moai/manager-lead.md
63:  ... export of the deciding lines to the tracked `.moai/reports/<card-id>/` ...
152: ... write ... to the tracked path `.moai/reports/<card-id>/M<n>.<AC-id>.log` ...
154: ... and only the tracked path does. ...

$ grep -n 'M<n>-report' .claude/agents/moai/manager-lead.md
161:M<n>: <AC-id-1>=PASS, ... | evidence: .moai/reports/<card-id>/M<n>-report.md | fold-at: <ISO-8601>

$ grep -n 'tracked' .claude/output-styles/moai/moai.md
354:   └─ evidence: .moai/reports/<card-id>/<check>.log  (tracked; exported before citing — ...)
561:📎 Evidence: .moai/reports/<card-id>/<check>.log  (tracked; exported before citing — ...)
583, 620: (무관 — "tracked checklist", "tracked items")
```

**결정적 관측**: `tracked` 3행 중 `M<n>-report.md`를 이름 붙이는 행은 **0행**이다.

```
$ grep -n 'tracked' .claude/agents/moai/manager-lead.md | grep -c 'M<n>-report'
0
```

iter1의 `AC-EPE-012`가 바로 이 형태였고, 따라서 **착수 전부터 초록**이었다. `:161`은
`M<n>-report.md`를 해소 가능한 인용 경로로 제시하지만 「tracked」라는 낱말을 담지 않고,
그것을 추적 주장으로 만드는 것은 `:154`의 `"and only the tracked path does"`다.

정정된 판정식의 RED-now:

| 프로브 | 값 |
|---|---|
| ``grep -c 'tracked path `.moai/reports/<card-id>/M<n>\.<AC-id>\.log`' manager-lead.md`` | **1** |
| `grep -c 'only the tracked path does' manager-lead.md` | **1** |
| `grep -c '(tracked; exported before citing' moai.md` | **2** |
| 양성 대조 `grep -c 'ZZZNOPE' manager-lead.md` | **0** |

### 4.3 생존해야 하는 표면 (수정 대상 아님)

```
$ sed -n '601p' .claude/agents/moai/plan-auditor.md
[HARD] **Export mandate — an audit is complete only when its verdict is exported.** ...
  `.moai/reports/<card-id>/plan-audit.md` (or `plan-audit-iter<N>.md` ...)

$ sed -n '108p' .claude/agents/moai/sync-auditor.md
[HARD] **Export mandate — an audit is complete only when its verdict is exported.** ...
  `.moai/reports/<card-id>/sync-audit.md` (or `sync-audit-verdict*.md` ...)
```

둘 다 **진짜 판정서**를 반출하므로 문언 수정 대상이 아니다. 폭 (A)면 참이 되고 (B)면 거짓으로
남는다 — `AC-EPE-013`이 그 상태를 기록하게 한다.

---

## §5 docs-site 인벤토리 — 로케일 내성 측정

### 5.1 영어 토큰만 겨눈 측정 (불충분)

```
$ grep -rnE 'tracked' --include='*.md' docs-site/ | grep -cE 'moai/reports'
5     # 3파일: en/advanced/{token-budget,agent-guide,manager-lead}.md
```

이 수치만 보면 영어 페이지 3개의 문제로 읽힌다. **그것이 iter1이 이 표면을 놓친 경로다.**

### 5.2 로케일 내성 측정

```
$ grep -rnE '(tracked|추적|追跡|跟踪|追踪)' --include='*.md' docs-site/content/ \
    | grep -E '\.moai/reports' | wc -l
20

$ (같은 파이프, 파일별)
docs-site/content/en/advanced/agent-guide.md     1
docs-site/content/en/advanced/manager-lead.md    2
docs-site/content/en/advanced/token-budget.md    2
docs-site/content/ja/advanced/agent-guide.md     1
docs-site/content/ja/advanced/manager-lead.md    2
docs-site/content/ja/advanced/token-budget.md    2
docs-site/content/ko/advanced/agent-guide.md     1
docs-site/content/ko/advanced/manager-lead.md    2
docs-site/content/ko/advanced/token-budget.md    2
docs-site/content/zh/advanced/agent-guide.md     1
docs-site/content/zh/advanced/manager-lead.md    2
docs-site/content/zh/advanced/token-budget.md    2
```

**12파일 / 20행 / 4로케일 × 3페이지 계열.**

양성 대조: `grep -rl 'moai/reports' docs-site/ | wc -l` → **41** (계측기 도달 확인).
음성 대조: 같은 프레임에 존재하지 않는 토큰 → **0**.

### 5.3 로케일별 번역 양상 (인벤토리 grep이 로케일 내성이어야 하는 이유)

| 로케일 | "tracked path" 표기 | 플레이스홀더 |
|---|---|---|
| en | `the tracked path` | `.moai/reports/<card-id>/` |
| ko | `추적 경로` | **`.moai/reports/<카드-id>/`** ← 플레이스홀더도 번역됨 |
| ja | `追跡対象のパス` | `.moai/reports/<card-id>/` (일부 `<カード-id>`) |
| zh | `受版本跟踪的路径` / `受跟踪的路径` | `.moai/reports/<card-id>/` (일부 `<卡片-id>`) |

`ko`가 플레이스홀더 자체를 번역하므로, **경로 토큰을 리터럴로 겨눈 grep도 로케일을 놓칠 수
있다**. §5.2의 프레임은 `\.moai/reports` 접두만 겨누어 이 함정을 피한다.

### 5.4 기계 하위 점검의 RED-now (iteration 3에서 전량 재측정)

> **정정 기록.** iteration 2의 이 절은 하위 점검을
> `grep -cE '(<check>\.log|M<n>-report\.md|M<n>\.<AC-id>\.log)'`로 적고 그 출력을 **18**로
> 실었다. 두 가지가 틀렸다. ① 그 정규식은 이 인벤토리에서 **0행**을 낸다 — 로케일이
> 플레이스홀더 자체를 번역하므로(`M<milestone>` / `M<마일스톤>` / `M<マイルストーン>` /
> `M<里程碑>`) `M<n>` 리터럴은 어디에도 없고, `M<n>-report.md`도 본문에는 구체 예시
> `M2-report.md`로 적혀 있으며, `<check>.log`는 docs-site에 존재하지 않는다. ② **18은 어떤
> 명령의 출력도 아니었다** — `20 − 2`(§8의 「나머지 2행」)로 **계산한** 값이 fenced 명령+출력
> 블록 안에 관측값으로 놓여 있었다. 아래는 전부 이 트리에서 다시 돌려 출력을 옮긴 것이다.

인벤토리는 **두 형태로 남김 없이 분할된다**. `AC-EPE-022`의 하위 점검은 이 두 갈래다.

```
$ grep -rnE '(tracked|추적|追跡|跟踪|追踪)' --include='*.md' docs-site/content/ \
    | grep -E '\.moai/reports' | wc -l
      20
$ (같은 파이프) | grep -cE '\.moai/reports/[^/ `]+/M[^ `]*\.(log|md)'      ← S1 (형태 i)
8
$ (같은 파이프) | grep -cE '\.moai/reports/[^/ `]+/`'                       ← S2 (형태 ii)
12
$ (같은 파이프) | grep -cE 'verdict\.md|plan-audit.*\.md|sync-audit.*\.md'  ← 판정서를 이름 붙인 행
0
$ (같은 파이프) | (원장 L-R1 — 잔차)                                        ← iteration 4 추가
       0
$ (같은 파이프) | (원장 L-R2 — 중복)                                        ← iteration 4 추가
0
$ (같은 파이프) | grep -cE 'ZZZNOPE_NOT_A_TOKEN'                            ← 음성 대조
0
```

8 + 12 = 20 = 인벤토리 전량. **판정서 파일을 이름 붙이는 행은 0행**이므로, 20행 전부가 위반이고
정상 행은 없다.

[HARD] **정정 (iteration 4, D16) — 이 항등식은 「남김 없음」의 증거가 아니다.** 종전 문장은
`8 + 12 = 20`을 「하위 점검이 인벤토리의 어느 부분도 놓치지 않는다」의 증거로 적었는데, 그
추론은 성립하지 않는다: 양쪽에 걸리는 행 1개와 어느 쪽에도 안 걸리는 행 1개가 함께 있어도
합은 똑같이 20이다. 증거가 되는 것은 **잔차 `L-R1` = 0과 중복 `L-R2` = 0의 직접 측정**이고,
위 블록에 그 두 줄을 실었다. 결론(분할이 배타적이고 남김 없다)은 참이며 바뀌지 않는다 —
바뀐 것은 무엇이 그 결론을 떠받치는가다. `acceptance.md` `AC-EPE-022`가 두 잔차 점검을
PASS 조건으로 싣는다.

S1이 잡는 8행 (`cut -d: -f1,2`, 이번 실행):

```
docs-site/content/{en,ko,ja,zh}/advanced/manager-lead.md:122
docs-site/content/{en,ko,ja,zh}/advanced/manager-lead.md:123
```

S2가 잡는 12행:

```
docs-site/content/{en,ja,zh}/advanced/agent-guide.md:186   (ko는 :212 — 문단 구조가 다르다)
docs-site/content/{en,ja,zh}/advanced/token-budget.md:50   (ko는 :96)
docs-site/content/{en,ja,zh}/advanced/token-budget.md:102  (ko는 :129)
```

`ja/advanced/agent-guide.md`는 `:187`이다. 로케일별 줄번호 차이는 번역 길이에서 오며, 두
프로브는 줄번호를 전제하지 않는다.

### 5.5 iter1 감사와의 차이

iter1 감사는 같은 축을 **19행**으로 보고했다. 이 측정은 **20행**이다. 차이 1행은
`ko/advanced/agent-guide.md`의 「추적 경로」 문장이며, 정규식 폭 차이에서 온다. 둘 다 이
트리에서 잰 값이고 **어느 쪽도 인수조건의 정확-일치 대상이 아니다** — `AC-EPE-022`가 재는
것은 위반 행의 0으로의 도달이다.

---

## §6 폭 수치 (primary 체크아웃 — 드리프트함)

```
$ find <primary>/.moai/reports -mindepth 2 -maxdepth 2 -type f -name 'verdict.md' | wc -l
305
$ (같은 find, 감사 계열 파일명)                                                      155
$ find <primary>/.moai/reports -mindepth 3 -name 'verdict.md' | wc -l                 19
$ (같은 프레임, 존재하지 않는 파일명 — 양성 대조)                                      0
```

깊이 3+ 19건 내역: 18건 `t264/rescued/**`(구조 아카이브), 1건
`.moai/reports/lead/merge-batch-20260913/verdict.md`(리드 배치 판정서 — 깊이 제한이 놓치는
**살아 있는** 형태).

```
$ git ls-files .moai/reports | wc -l
54
```

---

## §7 예산 측정 (Tier 판정 입력)

```
$ grep -oE 'REQ-EPE-[0-9]+[a-z]?' spec.md | sort -u | wc -l        (iter1) 17 → (iter2) 18
$ grep -cE '^\*\*AC-EPE-[0-9]+\*\*' acceptance.md                  (iter1) 20 → (iter2) 23
$ (양성 대조: 존재하지 않는 접두 REQ-ZZZ)                                             0
```

수정 표면 **20파일** — 파일 단위로 열거하면:

```
독트린 루트 2   .claude/rules/moai/core/agent-common-protocol.md
                .claude/rules/moai/core/agent-common-protocol-reference.md
도달 표면 루트 2 .claude/agents/moai/manager-lead.md
                .claude/output-styles/moai/moai.md
템플릿 미러 4   위 네 파일의 internal/template/templates/ 대응본 (ls로 4건 전부 확인)
docs-site 12    4로케일 × {agent-guide.md, manager-lead.md, token-budget.md} (§5.2 측정)
합계            2 + 2 + 4 + 12 = 20
```

Tier M 밴드는 파일 5-15, REQ/AC 16/16 — **세 축 모두 초과**.

> **정정 기록(iteration 3, D11).** iteration 2는 이 합계를 **16**으로 적었고 산식을
> `§4.1 4 + §4.2 2 + 템플릿 미러 4 + §5.2 12 − 도달 표면 중복 2`로 썼다. 그 산식의 구성 항을
> 그대로 더해도 **20**이 나온다(4+2+4+12−2 = 20) — 합계만 어긋나 있었고, `spec.md` §1.3(d)의
> 층별 표(2+2+4+12)도 같은 20을 낸다. 위 열거는 중복 없이 파일을 직접 세는 형태로 다시 썼다.
> **Tier 결론은 바뀌지 않는다** — 20 > 15이므로 16보다 오히려 더 확실하다.
>
> 기계 재생성물 `internal/template/templates/.codex/agents/moai/manager-lead.toml`은 이 수에
> 들어가지 않는다 — 손편집 대상이 아니라 `make agents-emit`의 산출물이며, `AC-EPE-015`가
> 그 축을 따로 잰다.

---

## §8 미측정 (Gaps)

- **`main` 트리의 `.gitignore` 커밋본을 git으로 읽지 못했다.** 워크트리 가드가 `git show
  main:.gitignore`를 거부한다. 319줄은 **primary 체크아웃의 워킹 사본**을 비-git `wc -l`로
  잰 값이며, 커밋본과의 동일성은 확인하지 않았다.
- **예외 규칙의 제자리 효과를 재지 않았다.** `.gitignore`를 편집해야 하는 run 단계의 일이고
  plan 단계의 쓰기 범위 밖이다. `AC-EPE-004`(누수)·`AC-EPE-005`(깊이)의 실제 결과는 미측정이며,
  관측된 것은 **현재 규칙 하의 상태**뿐이다.
- **docs-site 빌드를 돌리지 않았다.** 4-로케일 편집이 warning-free 빌드를 유지하는지는 run
  단계에서 처음 측정된다(AC-EPE-023).
- **docs-site 20행 각각의 대체 문언은 아직 없다.** 기계적으로는 §5.4가 20행을 S1 8 + S2 12로
  남김 없이 분할하고, 판정서를 이름 붙이는 행이 0행임까지 확인했다 — 즉 **정상 행은 없고 20행
  전부가 위반이다**. 미측정인 것은 「각 행을 무엇으로 바꿀 것인가」이며, 그 판단과 로케일별
  자연스러움 확인은 M5의 일이다(`AC-EPE-023`).
- **`:294`를 제거했을 때의 귀결**은 재지 않았다 — 무리 D 정리가 범위 밖이기 때문이다.
