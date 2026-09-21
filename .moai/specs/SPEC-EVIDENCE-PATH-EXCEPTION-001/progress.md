# SPEC-EVIDENCE-PATH-EXCEPTION-001 — 진행 기록

- **SPEC**: SPEC-EVIDENCE-PATH-EXCEPTION-001
- **카드**: t1039
- **상태**: `in-progress` (run 단계 M1 착수, 2026-09-21)
- **Tier**: **L** (iteration 2에서 M → L 재분류, `spec.md` §0의 판별식)
- **트리**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1039`, 브랜치 `WT-evidence-path`,
  HEAD `116820f40`

---

## §E.1 Plan-phase Audit-Ready Signal

**Claim** — plan 단계 산출물 5종(Tier L: `spec.md` / `plan.md` / `acceptance.md` / `design.md` /
`research.md` + `progress.md`)이 저작됐고, 폭 경계 축이 **미해소인 채로** 닫힐 수 있는 구조를
갖췄으며, iteration 1 감사의 blocking 6건과 iteration 2 감사의 blocking 3건(D10·D11·D12) +
optional 2건(D13·D14)이 전부 닫혔다. 결함별 내역은 §G의 iteration 기록.

**Evidence** (이 트리, HEAD `116820f40`, 이번 실행):

- SPEC ID 정규식 자가 점검:
  ```
  $ ID="SPEC-EVIDENCE-PATH-EXCEPTION-001"
  $ [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
  PASS
  ```
- 예산 측정(Tier 판정 입력):
  ```
  $ grep -oE 'REQ-EPE-[0-9]+[a-z]?' spec.md | sort -u | wc -l   → 18
  $ grep -cE '^\*\*AC-EPE-[0-9]+\*\*' acceptance.md             → 23
  $ (양성 대조: 존재하지 않는 접두 REQ-ZZZ)                        → 0
  ```
  Tier L 상한 25/25 이내. 수정 표면 **20파일**(`spec.md` §1.3(d) — iteration 3이 16에서 정정;
  층별 2+2+4+12. Tier 결론 불변).
- `.gitignore` 규칙 지형 전량 재측정(`grep -n 'moai/reports' .gitignore` → 23행) 및 `-v` 귀속
  측정 5건 — 전문은 `research.md` §1·§2.
  - **`:294`가 `plan-audit/verdict.md`를 결정한다**(후보 삽입 지점보다 아래).
  - **`:376`은 깊이 1 전용**이라 `<card>/verdict.md`에 닿지 않는다.
  - 관용구 줄 범위 정정: t338 `234-236` / t528 `237-241` / t530 `242-244` / t229 `245-247`.
  - 블랭킷 주석 좌표 정정: `:224-226`(iter1의 `:225-228`은 오류).
- 이 카드 자신의 실사례 측정(`spec.md` §1.0):
  ```
  $ git status --porcelain --ignored -- .moai/reports/t1039   → !! .moai/reports/t1039/
  $ git status --porcelain -- .moai/reports/t1039             → 0행 (구조적 침묵)
  $ ls -1 .moai/reports/t1039/                                → lane-measurements.md, plan-audit.md
  ```
- docs-site 로케일 내성 인벤토리: **12파일 / 20행 / 4로케일 × 3페이지 계열**.
  영어 토큰만 겨눈 grep은 **3파일 5행**만 본다. 양성 대조 `grep -rl 'moai/reports' docs-site/`
  → **41**. 전문은 `research.md` §5.
- 정정된 `AC-EPE-012` 판정식의 RED-now: P-a **1**, P-b **1**, P-c **2**, 양성 대조 **0**
  (`research.md` §4.2).
- 신설 `AC-EPE-021`의 RED-now: `grep -c 'reports/lead' .gitignore` → **0**,
  양성 대조 `grep -c 'moai/reports' .gitignore` → **23**.
- 폭 수치(primary 체크아웃, `find`): `verdict.md` 깊이 2 → **305**; 감사 계열 깊이 2 → **155**;
  `verdict.md` 깊이 3+ → **19**. 양성 대조 → **0**.

**Baseline-attribution** — 모든 `grep`/`sed`/`ls`/`check-ignore`/`git status`는 이
워크트리(HEAD `116820f40`)에서, 모든 `find` 폭 수치는 primary 체크아웃
`/Users/goos/MoAI/moai-adk-go` 에서, 이번 실행에 수행됐다. 두 트리는 서로 다른 답을 내며
(이 워크트리의 `.moai/reports`는 54개 파일뿐), 그래서 각각 명명했다.

**Gaps** (plan 단계에서 의도적으로 닫지 않은 것):

- **폭 경계 (A)/(B)** — 운영자 소관, 미해소. `spec.md` §3.0이 두 독법의 수치와 귀결을 싣고,
  §3.4가 영향 범위를 **두 등급**으로 전수 열거(결정 의존 AC-EPE-002·AC-EPE-013, 참조 의존
  REQ-EPE-010)한다. plan 단계는 열린 채로 닫힌다.
- **예외 규칙의 제자리 효과** — 미측정. `spec.md` §5는 픽스처 가설이며 규칙 순서(특히 무리 D)를
  재현하지 않는다. REQ-EPE-008이 run 단계에 재측정을 의무화하고 픽스처 인용을 금지한다.
- **`.gitignore` 최종 형태와 삽입 위치** — 미결정. `design.md`가 후보와 각각의 귀결을 적었고,
  고르는 것은 M1의 제자리 측정이다.
- **docs-site 20행의 행 단위 판정과 빌드** — 미측정. 기계 하위 점검은 18을 내지만 나머지 2행의
  판정은 M5가 행 단위로 읽는다. warning-free 빌드는 run 단계에서 처음 측정된다.
- **`main` 트리 `.gitignore`의 커밋본** — 워크트리 가드가 `git show main:.gitignore`를
  거부한다. 319줄은 primary 워킹 사본의 비-git `wc -l` 값이다.
- **서브에이전트 쓰기 범위 가드** — 범위 밖(`spec.md` §7). 카드화 여부는 운영자 판정 대기.
- **소스↔소스 방출 검사의 사각(노후 바이너리)** — 이 카드에서 촉발되지 않으므로 미측정.
  `plan.md` §F M4의 검증 주의에 run 단계용 상기로 기록.

**Residual-risk**:

- **이 SPEC 자신의 근거가 반출돼 있지 않다.** `.moai/reports/t1039/lane-measurements.md`와
  `.moai/reports/t1039/plan-audit.md`는 `:227`로 무시된다 — 이 카드가 열려는 바로 그 규칙이다.
  iter1 감사 판정서가 `plan-auditor.md:601`의 [HARD] Export mandate를 만족하지 못하며, 이것이
  이 결함의 **6번째 실사례**이자 유일하게 이 카드 자신의 절차에서 나온 것이다.
  「증거가 반출됐다」고 주장하지 않는다. **워크트리를 폐기하면 두 파일의 유일본이 사라진다.**
- **Tier 재분류가 다음 iteration의 채점 기준을 바꾼다.** Tier L의 PASS 임계는 0.85(Tier M은
  0.80)이고 산출물이 5종이다. 범위를 정직하게 잡은 대가로 통과선이 올라간다는 것은 의도된
  교환이지만, iter1과 같은 기준으로 비교되지 않는다는 사실은 기록해 둔다.
- primary 체크아웃 폭 수치는 다른 세션이 쓰는 트리의 스냅샷이며 드리프트한다(레인 304 →
  iter1 감사 305 → 이번 305).
- docs-site 행 수도 드리프트한다. iter1 감사 19행 / 이번 20행 — 정규식 폭 차이이며 어느 쪽도
  인수조건의 정확-일치 대상이 아니다.
- `.gitignore` 줄수·줄번호는 트리 의존적이다(이 워크트리 417, `main` 트리 319). 귀속 없이
  인용되면 다른 트리에서 조용히 빗나간다.
- 선행 시도가 이 워크트리의 `.gitignore`를 파괴한 전례가 있고, 그것을 막은 장치는 없었다.

---

## §E.2 Run-phase Evidence

> 귀속: 모든 측정은 이 워크트리
> `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1039`, 브랜치 `WT-evidence-path`,
> 2026-09-21 이번 실행. 착수 HEAD `116820f40`, 종료 HEAD `733cd5a02`.
> 도구 귀속(§2.2): `moai spec lint` 는 이 트리에서 빌드한 `./bin/moai`
> (`moai_cp/20260910_130400-2148-g429a7b3d4`, `-dirty` 접미 없음)로 실행했다 —
> 판정 빌드의 커밋이 측정 대상 트리의 조상도 후손도 아닌 **같은 커밋**이다.

### §E.2.0 폭 경계 확정 — `REQ-EPE-006`의 [UNRESOLVED] 해소

**운영자 판정: 독법 (B).** 되살리는 것은 `verdict.md` **한 파일**, 깊이는 **카드 디렉터리 한
단계**(`.moai/reports/<card-id>/verdict.md`). 감사 판정서 계열(독법 A)은 **채택되지 않았다.**
`t264/rescued` 아카이브도 들어오지 않는다.

출처: 이 run 단계 배차문의 「Operator decision — the width axis is now CLOSED」 절, 2026-09-21.
`AC-EPE-019`가 요구하는 것은 (a) 「운영자가 명시했고 그 출처가 기록돼 있다」이며, 이 항목이 그
기록이다. `spec.md` §3.3의 **권고는 (A)였고 운영자는 (B)를 골랐다** — 권고가 결정이 아니라는
것이 이 자리에서 실제로 성립했다.

[HARD] **이 선택은 결함을 남기기로 하는 결정이며 실수가 아니다.** 귀결은 §E.3 Gaps에 적는다.

### §E.2.1 AC PASS/FAIL 행렬 (23건 전량)

| AC | 처분 | 판정 | 검증 명령 | 관측 출력 |
|---|---|---|---|---|
| 001 | release-blocking | **PASS** | `git check-ignore --no-index -q .moai/reports/tZZZ/verdict.md` | `rc=1` (착수 전 `rc=0`). 양성 대조 `report.md` → `rc=0` |
| 002 | width-parameterized → (B) 확정으로 **평가 가능** | **PASS** | 같은 P1을 (B) 확정 집합 `{verdict.md}` 에 적용 | `rc=1`. 양성 대조 `evidence.md` → `rc=0` |
| 003 | regression-guard | **PASS** | `git check-ignore --no-index` 에 `report.md` `evidence.md` `pr-body.md` `probe.log` 동시 투입 | 네 경로 **전원** 무시 목록에 출력됨. 양성 대조 `verdict.md` → 목록에 **없음**(`rc=1`) |
| 004 | regression-guard | **PASS** | `git check-ignore -v --no-index .moai/reports/plan-audit/verdict.md` | `.gitignore:319:.moai/reports/plan-audit/*.md` → `rc=0`. 음성 대조 `.gitkeep` → `rc=1` |
| 005 | regression-guard | **PASS** | `git check-ignore --no-index -q .moai/reports/tZZZ/sub/verdict.md` | `rc=0`. 양성 대조 `tZZZ/verdict.md` → `rc=1` |
| 006 | post-change-only | **PASS** | `wc -l .gitignore` + `git check-ignore -v --no-index .../tZZZ/verdict.md` | `450`; `.gitignore:259:!.moai/reports/*/verdict.md` — 이 카드가 삽입한 줄. 양성 대조 `report.md` → `.gitignore:258:.moai/reports/*/*` |
| 007 | regression-guard | **PASS** | `git ls-files .moai/reports \| wc -l` + `git status --porcelain .moai/reports` | `54` (착수와 동일), status **0행** — ` D` 행 없음 |
| 008 | release-blocking | **PASS** | `grep -n '!\.moai/reports/\*\*/\*\.log' .gitignore` | 0행, `rc=1` (착수 전 `1`). 양성 대조 `!\.moai/reports/t338/` → `2` |
| 009 | regression-guard | **PASS** | `git check-ignore --no-index -q .moai/reports/tZZZ/probe.log` | `rc=0` — 철회 전과 동일. 양성 대조 `verdict.md` → `rc=1` |
| 010 | release-blocking | **PASS** | `grep -c 'must resolve post-merge' .gitignore` | `0` (착수 전 `1`). 양성 대조 `grep -c 'moai/reports'` → `31` |
| 011 | post-change-only | **PASS** | `grep -rlnE 'citation target is a (\*\*)?tracked(\*\*)? path' --include='*.md' --include='*.toml' .` | 6파일 → **2파일**. 남은 둘은 완료 SPEC의 진행 기록과 이 SPEC 자신의 `spec.md`(자기-적중) — 둘 다 수정 표면이 아니다. 양성 대조(존재하지 않는 토큰) → `0` |
| 012 | release-blocking | **PASS** | P-a/P-b/P-c/P-d 네 프로브 | `1→0`, `1→0`, `2→0`, `1→0`. 양성 대조 `grep -c 'moai/reports' manager-lead.md` → `5`, 음성 대조 `ZZZNOPE` → `0` |
| 013 | width-parameterized (귀결 기록형) | **PASS** | 감사 계열 6개 파일명을 P1에 동시 투입 | `plan-audit.md` `plan-audit-iter1.md` `plan-audit-iter4.md` `sync-audit.md` `sync-audit-verdict.md` `review-verdict.md` **전원** 무시 목록에 출력. `verdict.md`는 목록에 없음. (B) 귀결표와 일치 — 두 [HARD] 문장은 **거짓인 채로 남는다**. 잔존 결함은 §E.3 Gaps에 기록 |
| 014 | post-change-only | **PASS** | 미러 4종에 좁혀진 문언 grep + 네 프로브 재측정 | 규칙 미러 2종에 좁혀진 문장 각 `1`; 미러 네 프로브 `0/0/0/0`. 양성 대조 `5`, 음성 대조 `0` |
| 015 | post-change-only | **PASS** | `make agents-emit-check` | `ok github.com/modu-ai/moai-adk/internal/template/agentemit`, **exit 0** |
| 016 | post-change-only | **PASS** | `git diff -U0 -- internal/template/templates/` hunk에 금지 클래스 grep | `SPEC-` `REQ-EPE` `AC-EPE` 날짜 SHA `CLAUDE.local` `/Users/` `t1039` → **전원 0**. 양성 대조 같은 hunk `grep -c 'evidence'` → `18` |
| 017 | post-change-only | **PASS** | 선행 SPEC frontmatter + HISTORY 판독 | `related_specs:` `:15`, `partially_superseded_by:` `:16`, HISTORY 2026-09-21 행 `1` |
| 018 | regression-guard | **PASS** | `git diff -U0 429a7b3d4~1 -- <선행 spec.md> \| grep '^-' \| grep -c 'REQ-ECC'` | `0`. 삭제 행 자체가 **0**(순수 추가). 양성 대조 추가 행 → `30` |
| 019 | process-record | **PASS** | §E.2.0 판독 | 운영자가 (B)를 명시했고 출처가 기록됨 — (a) 갈래로 닫힘 |
| 020 | process-record | **PASS** | 폭 무관 AC 실제 평가 | 001·003~012·014~018·021~023 어느 것도 (B)/(A) 값을 입력으로 요구하지 않았다. `REQ-EPE-010`의 **참조 의존** 분류도 확인됨 — 아래 §E.2.3 |
| 021 | release-blocking | **PASS** | `grep -n 'reports/lead' .gitignore` | `248:# The live shape it therefore misses is .moai/reports/lead/<batch>/verdict.md —` (착수 전 `0`). 양성 대조 `31` |
| 022 | release-blocking | **PASS** | 로케일 내성 인벤토리 + 판정식 원장 `L-S1`/`L-S2`/`L-R1`/`L-R2` | `S1 8→0`, `S2 12→0`, `L-R1 0→0`, `L-R2 0→0`. 인벤토리 `20→16`이고 **16행 전원이 판정서를 이름 붙인다**(`L-VD 0→16`). 양성 대조 41파일, 음성 대조 `0` |
| 023 | post-change-only | **PASS** | 로케일 패리티 + `hugo --gc --minify` | 페이지 stem 3종 전부 count **4**; 빌드 exit 0, WARN/ERROR **0행**(로그 18행), KO 188 / EN 186 / JA 186 / ZH 186 |

**23/23 PASS. FAIL 0, 평가 불가 0.**

### §E.2.2 `AC-EPE-012`가 요구하는 대체 문언 인용

네 프로브를 0으로 만든 것은 삭제가 아니라 **대체**다. 지우기만 하면 「무엇이 참인가」가 사라지므로
인용 없는 0은 미완이라고 `acceptance.md`가 못 박았다. 대체 문언은 다음과 같다.

- **P-a·P-b 자리**(`manager-lead.md` Step 1) — "the lines that decided the verdict … are written
  into the tracked verdict file `.moai/reports/<card-id>/verdict.md`, and let the AC row name
  **that** file. The verdict file is the only tracked name under a card directory; a sibling
  artifact written beside it stays ignored, so citing one produces a path that resolves nowhere
  off this machine." / 그리고 "… and only the verdict file does."
- **P-c 자리**(`moai.md` 배너 2곳) —
  "`📎 Evidence: .moai/reports/<card-id>/verdict.md  (the one tracked name; deciding lines carried
  into it — see agent-common-protocol.md § Evidence export)`"
- **P-d 자리**(`manager-lead.md` Context-Folding 능력 줄) — "the deciding lines carried into the
  tracked verdict file `.moai/reports/<card-id>/verdict.md`"
- **독트린 본문**(`agent-common-protocol.md` § Evidence export) — "The one tracked citation target
  is the **verdict file** … The directory around it is not tracked: the ignore rules re-include
  that single filename and nothing else, so no other artifact under a card directory may be
  described as tracked."

의무의 이름이 **export before citing → carry the deciding evidence into the verdict** 로 바뀌었다.
후반부(반출하지 않기로 한 것은 인용하지 않는다)는 힘이 그대로이며, 문언만 「스크래치에 남겨 둔
것을 인용하지 않는다」로 좁혀졌다.

### §E.2.3 제자리 측정이 plan 단계 가설을 어떻게 갈랐는가 (`REQ-EPE-008`)

[HARD] 아래 결론은 전부 **이 트리의 실제 450줄 `.gitignore`에서 이번 실행에** 관측한 것이다.
`spec.md` §5의 4줄 픽스처 표는 어느 행의 baseline으로도 인용되지 않았다.

1. **삽입 위치는 `design.md` 기본 후보(무리 B 직후)가 아니라 `:227`과 `:228` 사이다.** 그 자리에
   넣으면 블랭킷 바로 아래에 오고, **무리 B 자신의 `plan-audit/*` 재-제외가 누수를 닫는다.**
   `design.md` §2가 (ㄱ)의 근거로 지목한 것은 아래쪽 무리 D(`:294` → 이동 후 `:319`)였는데,
   실측은 **두 겹이 모두 닫는다**는 것이었다.
2. **(ㄱ)의 전제가 직접 측정됐다.** `design.md` §2.2는 「무리 D는 `plan-audit/*.md`만 겨누므로
   `.md` 아닌 파일이 집합에 들어오면 (ㄱ)이 깨진다」를 **미측정 전제**로 적었다. 이번에 비-`.md`
   경로로 그 전제를 갈랐다:
   ```
   $ git check-ignore -v --no-index .moai/reports/plan-audit/verdict.txt
   .gitignore:254:.moai/reports/plan-audit/*	.moai/reports/plan-audit/verdict.txt
   ```
   결정한 것은 무리 D가 아니라 **무리 B(`:254`)** 다. 즉 무리 D를 정리하는 별도 카드가 나중에
   `:319`를 지워도 `plan-audit/` 봉쇄는 유지된다. **`REQ-EPE-007`은 새 줄 없이 충족됐다.**
3. **블랭킷이 살아 있음을 따로 보였다.** `AC-EPE-006`의 양성 대조는 `report.md`에 블랭킷
   (`.moai/reports/*`)이 나오기를 기대했으나, 실제로 그 경로를 결정하는 것은 이 카드가 넣은
   `:258:.moai/reports/*/*` 였다. 블랭킷만이 결정하는 깊이-1 경로로 다시 쟀다:
   ```
   $ git check-ignore -v --no-index .moai/reports/stray.txt
   .gitignore:227:.moai/reports/*	.moai/reports/stray.txt
   ```
   **대조의 의도(블랭킷 생존 확인)는 충족되고 대조의 문언(어느 경로로 재는가)은 빗나갔다.**
   문언을 만족시키려 결과를 고쳐 쓰지 않고, 빗나간 사실과 대체 프로브를 함께 적는다.
4. **무리 E(`:376` → 이동 후 `:401`)는 여전히 깊이 1 전용**이며 `<card>/verdict.md`에 닿지 않는다
   (`.moai/reports/top.md` → `:401`이 결정). `plan.md` M1-3이 「적지 않으면 다음 사람이 다시 잰다」
   고 한 항목이다.
5. **P2 지상 진실로 폭 전체를 한 번에 확인했다.** 실파일 4개를 만들어 재고 즉시 지웠다:
   ```
   $ git status --porcelain --untracked-files=all .moai/reports/tZZZPROBE
   ?? .moai/reports/tZZZPROBE/verdict.md
   ```
   `report.md` · `probe.log` · `sub/verdict.md` 는 나타나지 않았다 — 되살아나는 것은 깊이 2의
   `verdict.md` 하나뿐이라는 것이 P1(규칙 귀속)과 P2(add 대상) 두 경로에서 같은 답을 냈다.
   프로브는 측정 직후 제거했고(`ls` → No such file), `.moai/reports` status 는 다시 0행이다.

### §E.2.4 `AC-EPE-023`이 요구하는 로케일 확인 (grep으로 재지 않는 항목)

[HARD] 기록되지 않으면 PASS가 아닌 항목이므로 여기 적는다. 네 로케일의 수정 문장을 각각 읽고
확인했다:

- **en** — 원문. "carry it into the verdict" / "the only tracked name under a card directory".
- **ko** — 영어 구문을 그대로 옮기지 않고 한국어 문어로 다시 썼다. 「판정서 안에 적고」,
  「카드 디렉터리에서 추적되는 이름은 판정서 하나뿐이라」, 「스크래치에 남겨 둔 것을 인용해서는
  안 됩니다」. 소제목도 「증거 저장과 반출」 → 「증거 저장과 판정문 기록」으로 바꿔 본문과 맞췄다.
- **ja** — 「判定書に書き込み」, 「カード ディレクトリで追跡される名前は判定書ひとつだけで」,
  「スクラッチに残したものを根拠として差し出してはいけません」. 소제목 「証拠の保存とエクスポート」
  → 「証拠の保存と判定書への書き込み」.
- **zh** — 「写进受版本跟踪的判定书」, 「卡片目录里受跟踪的名字只有这份判定书」,
  「留在暂存区的材料，绝不拿来当判定依据」. 소제목 「先采集，再导出」 → 「先采集，再写进判定书」.

**영어 문장이 비-영어 페이지에 그대로 들어간 곳은 없다.** 경로·파일명·코드 토큰
(`.moai/reports/<card-id>/verdict.md`, `progress.md`, `/compact`)은 주소이므로 네 로케일 모두
축자 유지했다.

### §E.2.5 인벤토리 20 → 16은 축소가 아니라 주장 병합이다

`AC-EPE-022`는 정확 행 수를 단언하지 않지만, 행 수가 줄어든 이유를 적지 않으면 다음 사람이
「계측기가 도달을 잃었나」를 다시 재야 한다. 줄어든 4행은 로케일당 1행씩이며, `manager-lead.md`
의 두 주장(1번 항목의 반출 문장 + 2번 항목의 「추적 경로뿐」 문장)이 **판정서 하나를 가리키는 한
주장으로 합쳐지면서** 「tracked」 토큰을 담은 행이 2→1이 된 결과다. 계측기 도달은 불변이다 —
양성 대조 41파일이 착수 전과 같다.

### §E.2.6 마일스톤 순서 일탈 1건

`plan.md` §F는 M2 → M3 순이지만, 실행은 **M1 → M3 → M2 → M4 → M5 → M6** 이었다. M1과 M3가 같은
파일(`.gitignore`)을 만지므로 연속 처리해 두 번의 읽기 사이에 파일이 바뀔 창을 없앴다 — 이
워크트리의 `.gitignore`가 선행 시도에서 파괴됐고 그것을 적발한 신호가 「같은 파일을 두 번 읽었는데
규칙 줄번호가 움직였다」였기 때문이다. **범위 일탈이 아니라 순서 일탈**이며, M3 종료 시점에
`AC-EPE-009`(철회 후 동작 무변화)를 재측정해 M1의 예외가 깨지지 않았음을 확인했다.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-21
run_commit_sha: 733cd5a02
run_base_sha: 116820f40
run_status: complete
ac_pass_count: 23
ac_fail_count: 0
ac_unevaluable_count: 0
preserve_list_post_run_count: 2   # plan-auditor.md / sync-auditor.md — 미수정 확인(git diff --stat 공백)
l44_pre_commit_fetch: not-performed   # 레인은 push하지 않는다(리드 일괄) — 아래 Gaps 참조
l44_post_push_fetch: not-performed
new_warnings_or_lints_introduced: 0   # 1건 도입 후 같은 run 안에서 닫음 — 아래 Gaps 참조
cross_platform_build:
  darwin: exit-0
  windows_amd64: exit-0
total_run_phase_files: 24
m1_to_mN_commit_strategy: per-milestone-commit   # 8 commits, no amend, no force-push
gitignore_lines: 417 -> 450   # +36 / -3, 이 워크트리 기준
```

### Gaps — 이번 실행에서 닫지 못했거나 의도적으로 열어 둔 것

1. **[HARD] 폭 독법 (B)의 귀결 — 두 [HARD] Export mandate가 거짓인 채로 남는다.**
   `plan-auditor.md:601`과 `sync-auditor.md:108`은 "an audit is complete only when its verdict is
   exported"를 [HARD]로 규정하고 반출 목적지를 각각
   `.moai/reports/<card-id>/plan-audit.md`(및 `plan-audit-iter<N>.md`)와
   `.moai/reports/<card-id>/sync-audit.md`(및 `sync-audit-verdict*.md`)로 못 박는다. (B)는 그
   파일명들을 되살리지 않으므로 **이 카드가 착지한 뒤에도 두 문장은 거짓이다.** 측정:
   여섯 파일명 전원이 P1 무시 목록에 출력되고 `verdict.md`만 목록에 없다.

   **구체적 귀결 — 이 카드 자신의 증거가 여전히 반출되지 않는다.** 이 SPEC을 네 번 심사한
   plan-audit 판정서 4건과 레인 측정 기록 1건은 전부 무시된 채로 남는다:
   ```
   $ git status --porcelain --untracked-files=all --ignored .moai/reports/t1039
   !! .moai/reports/t1039/lane-measurements.md
   !! .moai/reports/t1039/plan-audit-iter2.md
   !! .moai/reports/t1039/plan-audit-iter3.md
   !! .moai/reports/t1039/plan-audit-iter4.md
   !! .moai/reports/t1039/plan-audit.md
   ```
   **이 워크트리를 폐기하면 다섯 파일의 유일본이 사라진다.**

   [HARD] **운영자는 이 대가를 알고 (B)를 골랐다**(§E.2.0). 따라서 이것은 미처리 결함이 아니라
   **기록된 선택**이며, 이 카드는 그것을 메우려고 예외를 넓히지 않았다 — 결정된 교환을 조용히
   수선하는 것이 이 카드가 막으려는 실패 형태 그 자체다. 독트린 문언 쪽(두 문장이 거짓을
   주장하지 않도록 고치는 일)은 **별도 카드 t1059**가 맡는다. 다만 이 기록은 **t1059가 착지하기
   전의 사실**이다 — 위 다섯 파일은 이 글을 쓰는 시점에 무시된 상태다.

2. **SPEC 본문이 이 결정을 아직 반영하지 못했다 — 소유권 경계 때문이며, blocker 보고로 넘긴다.**
   `spec.md` §2.2 `REQ-EPE-006`은 여전히 `[UNRESOLVED — 운영자 소관]`으로 읽히고, iter4 감사가
   지목한 D20·D21·D22 세 건도 닫히지 않았다. 이 편집들은 `spec.md` / `acceptance.md` /
   `design.md`의 **본문**이며, 이 에이전트의 소유권 경계가 명시적으로 금지하는 표면이다
   (`spec-frontmatter-schema.md` § Forbidden ownership crossings — frontmatter `status:` 와
   `updated:` 만 허용). 제안 문안을 담은 blocker 보고를 완료 보고에 포함했다.
   **런타임 동작에는 영향이 없다** — 예외·문언·미러·docs-site는 전부 (B)로 구현돼 있고,
   남은 것은 산문이 그 사실을 서술하는 방식이다.

3. **`internal/spec` 패키지 테스트가 FAIL한다 — develop에서 상속된 것이며 이 카드가 만들지 않았다.**
   실패 단언은 `ac_count_clause_test.go:479`
   `SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md: present in the snapshot but no longer matched by
   the corpus glob` 하나다. 그 파일은 카드 t1036의 분할(병합 `8f87f2359`, 이 브랜치 분기점
   `116820f40`의 **조상**)이 지웠고, 디렉터리에는 `spec.md`만 남아 있다. 귀속 측정:
   `git log 116820f40~1..HEAD -- .moai/specs/SPEC-MODEL-PROFILE-MATRIX-002` → **0 커밋**,
   양성 대조로 같은 범위를 이 SPEC 디렉터리에 걸면 **2 커밋**. 이 카드의 `acceptance.md`는
   같은 출력의 「reported, not failed」 목록(66건)에만 나타난다.
   **이 실패를 고치는 것은 범위 밖이고, 고치지 않은 채 남겨 둔다는 사실을 여기 적는다.**

4. **Go AC 카운터와 이 SPEC의 자기 점검이 1건 어긋난다.** 위 테스트 출력은 이 카드의
   `acceptance.md`를 `COUNT 24`로 읽는데, `acceptance.md` §B.10의 자기 점검과 이번 재측정은
   **23**이다(`grep -cE '^\*\*AC-EPE-[0-9]+\*\*'` → 23; 고유 식별자 `AC-EPE-001`~`023` → 23).
   어느 인수조건의 PASS 조건도 아니고 이 카드가 만든 차이도 아니다. **두 계수기가 서로 다른
   것을 세고 있다는 관측만 남기고, 어느 쪽이 옳은지는 재지 않았다.**

5. **도입했다가 같은 run 안에서 닫은 경고 1건.** M6 첫 판의 HISTORY 항목이 요구사항 ID를 굵은
   글머리로 세워 lint 가 그것을 두 번째 정의로 읽었다 — `DuplicateREQID` 2건 + `ModalityUnjudged`
   2건, 총 `0 error(s), 4 warning(s)`. 형태를 고쳐 `733cd5a02`에서 `No findings`로 닫았다.
   그 0이 곧 선행 SPEC의 baseline이 0이었다는 측정이기도 하다. **순 증가는 0이지만 「한 번도
   경고를 내지 않았다」는 주장은 거짓이므로 이렇게 적는다.**

6. **이 SPEC 자신의 lint 경고 31건은 착수 전 상태 그대로다.** `./bin/moai spec lint
   SPEC-EVIDENCE-PATH-EXCEPTION-001` → `0 error(s), 31 warning(s)`, exit 0. 전량이
   `ModalityMalformed` / `ModalityUnjudged` / `CoverageIncomplete` / `REQTableRowsRejected` 이며
   `spec.md` 327~490행의 요구사항 산문(한국어라 영어 EARS 키워드에 걸리지 않는다)과 AC 교차참조가
   `acceptance.md`에 사는 구조에서 나온다. run 단계가 `spec.md`에서 만진 것은 frontmatter의
   `status:` 와 `updated:` 두 줄뿐이고 어느 경고도 그 두 줄을 가리키지 않는다.
   **plan 단계 기준선이며 이 카드가 고칠 범위가 아니다.**

7. **`main` 트리 커밋본 `.gitignore`(319줄)를 git으로 읽지 못하는 상태가 그대로다.** 워크트리
   가드가 `git show main:.gitignore`를 거부한다(`research.md` §8의 plan-phase Gap). 이번 실행도
   같은 제약 아래 있었고, 모든 줄수·줄번호는 **이 워크트리 기준**으로만 적었다.

8. **docs-site 경고 계수기를 실제 경고로 발화시키지 못했다.** `hugo` 빌드는 WARN/ERROR 0행을
   냈고, 계수 정규식이 발화할 수 있음은 **합성 문자열**로만 확인했다(`printf 'WARN …' | grep -icE
   '^(WARN|ERROR)'` → `1`). 이 트리에 실제 경고를 주입해 재지는 않았다 — 그것은 docs-site를
   의도적으로 깨는 일이라 이 카드의 범위 밖이다. **따라서 「빌드가 warning-free다」는 관측이고,
   「계수기가 이 빌드의 경고를 잡을 수 있다」는 합성 대조에 기댄 추론이다.**

9. **push·CI 판정이 없다.** 이 레인은 `git push`를 수행하지 않았다(리포 규율: develop push는
   리드 일괄). 따라서 **깨끗한 환경의 전체 스위트 판정도, darwin/windows 매트릭스 CI 판정도
   이 기록에는 없다.** 여기 있는 것은 전부 로컬 조기 신호다.

### Residual-risk — 관측했는데도 여전히 틀릴 수 있는 것

- **예외가 규칙 **순서**에 의존한다.** `:258`/`:259`가 무리 B의 `plan-audit` 카브아웃보다 **위**에
  있어야 누수가 닫힌다. 나중에 누가 이 블록을 아래로 옮기면 `plan-audit/verdict.md`가 다시
  열리고, 그 사실은 어떤 테스트도 내지 않는다. 주석에 「must stay ABOVE」를 적었지만 **주석은
  기계가 아니다** — 기계 가드는 이 카드의 범위 밖이다(`spec.md` §7).
- **깊이 1 제한이 살아 있는 판정서 1건을 놓친다.** `.moai/reports/lead/<batch>/verdict.md`.
  의도된 배제이고 주석이 이름 붙였지만, 리드 배치 판정서는 **지금도 반출되지 않는다.**
- **폭 수치는 primary 체크아웃에서만 의미가 있고 드리프트한다.** 리드가 2026-09-21에 재측정한
  값은 깊이-2 `verdict.md` **319**(배차문의 304는 전날 값 — 차이는 명령 형태가 아니라 트리
  성장으로 확인됨). 이 워크트리에서 같은 find 는 다른 답을 낸다. **어떤 인수조건도 이 수치에
  정확-일치를 단언하지 않으며, 이 문단도 수치를 트리와 날짜와 함께만 적는다.**
- **문언 좁히기는 산문이므로 회귀 가드가 없다.** `AC-EPE-011`·`AC-EPE-012`는 이번 실행의
  일회성 측정이며, 다음에 누가 넓은 주장을 다시 쓰는 것을 막는 기계는 없다.
- **`.gitignore`는 이 카드가 고치려는 바로 그 파일이고, 선행 시도에서 파괴된 전례가 있다.**
  이번 실행은 제자리 실험을 하지 않고 최종 형태를 한 번에 적용한 뒤 매 편집마다
  `wc -l` + `git diff --stat`으로 증분을 확인했다(417 → 442 → 450, +25 / +11-3). 그럼에도
  **파괴를 막은 것은 규율이지 장치가 아니다.**

---

## §E.4 Sync-phase Audit-Ready Signal

> 귀속: 모든 측정은 이 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1039`,
> 브랜치 `WT-evidence-path`, 착수 HEAD `4cfd112fa`, 2026-09-21 이번 실행.
> 도구 귀속(§2.2): `moai spec lint` 는 `./bin/moai`(빌드 커밋 `429a7b3d4`)로 실행했다.
> 그 커밋과 HEAD 사이에 **Go 파일이 0개**이므로(양성 대조: 같은 범위 전체 파일 5개, 전부 마크다운)
> 판정 빌드의 lint 코드는 측정 대상 트리의 lint 코드와 같다.

```yaml
sync_complete_at: 2026-09-21
sync_commit_sha: 619774c89
sync_base_sha: 4cfd112fa
sync_status: complete
b12_self_test_a: pass        # grep -c 'SPEC-EVIDENCE-PATH-EXCEPTION-001' CHANGELOG.md → 0 (중복 없음)
b12_self_test_b: pass-with-observation   # 계수기 24, 실제 인수조건 23 — 아래 §E.4.2가 차이를 설명
b12_self_test_c: pass        # CHANGELOG가 주장한 경로 전량 ls 확인
changelog_entry_position: "[Unreleased] → ### Changed (선두)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"
  plan_md: n/a               # Artifact Statelessness — status: 필드를 담지 않음
  acceptance_md: n/a
  design_md: n/a
  research_md: n/a
  updated_field: unchanged   # 이미 2026-09-21
mx_tag_validation: not-applicable   # 이 카드의 소스 파일 변경 0건
docs_surfaces_updated: [CHANGELOG.md]
docs_surfaces_deliberately_skipped: [README(4로케일), docs-site(M5에서 완료)]
canary_compliance_check:
  subject: "이 SPEC이 정한 정책이 이 SPEC 자신의 sync 시점에 성립하는가"
  verdict_md_unignored: true
  siblings_still_ignored: true
  own_evidence_files_still_ignored: true   # 의도된 대가 — 아래 Gaps 1번
```

### §E.4.1 착지 동작 재측정 (이번 실행)

**Claim** — run 단계가 기록한 ignore 동작이 sync 시점에도 같다.

**Evidence** — `git check-ignore --no-index -q <경로>` (rc 0 = 무시됨):

```
.moai/reports/t272/verdict.md       rc=1   # README가 인용하는 실제 판정서 — 되살아났다
.moai/reports/tZZZ/verdict.md       rc=1   # 예외 발화
.moai/reports/tZZZ/report.md        rc=0   # 형제는 그대로 무시
.moai/reports/plan-audit/verdict.md rc=0   # 누수 없음
.moai/reports/tZZZ/sub/verdict.md   rc=0   # 깊이 1 한정
```

**Baseline-attribution** — 이 워크트리, HEAD `4cfd112fa`, 이번 실행. 레인의 run 단계 측정
(HEAD `7939e38b9`)과 리드의 독립 재현이 같은 다섯 값을 냈고, 이번이 세 번째 관측이다.

**Gaps** — 첫 행(`t272`)은 배차문에 없던 추가 측정이며, README 범위 판정의 근거로만 쓴다.

### §E.4.2 AC 계수 불일치 — §E.3 Gaps 4번 / §J.3이 열어 둔 Gap을 닫는다

**Claim** — 계수기가 24를 읽는 이유는 `acceptance.md:413`의 픽스처 문자열 안에 있는 `AC-001`이다.

**Evidence**:

```
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l
24
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u
AC-001 AC-EPE-001 … AC-EPE-023          # 24번째 토큰이 AC-001
$ grep -n 'AC-001' acceptance.md
413:$ printf '...`.moai/reports/<card-id>/M<n>.AC-001.log`'    | (L-R1)
$ grep -cE 'AC-ZZZZ-[0-9]+' acceptance.md   # 양성 대조
0
```

**판정**: 실제 인수조건은 **23**(`AC-EPE-001`~`023`, 결번 없음). 24번째는 회귀 가드의 `printf`
프로브 안에 축자로 들어 있는 리터럴이지 인수조건이 아니다.

[HARD] **주장의 범위를 좁혀 적는다.** 여기서 돌린 것은 **정규식 계수기**이고 Go 계수기가 아니다.
Go 계수기의 내부를 읽지 않았으므로, 같은 원인이라는 것은 **모순되지 않는다**까지이지
**확립됐다**가 아니다. §J.3의 금지 방향은 그대로 지켜졌다 — 어느 쪽 계수에 맞추려고도
문서를 고치지 않았다.

### §E.4.3 품질 게이트

```
$ ./bin/moai spec lint SPEC-EVIDENCE-PATH-EXCEPTION-001
0 error(s), 31 warning(s)      exit 0
```

run 단계 기준선 31과 같다. sync 단계가 `spec.md`에서 만진 것은 frontmatter `status:` 한 줄뿐이고
어느 경고도 그 줄을 가리키지 않는다.

### Gaps — 이번 sync가 닫지 못했거나 의도적으로 열어 둔 것

1. **[HARD] 두 [HARD] Export mandate는 여전히 거짓이다.** `plan-auditor.md:601`,
   `sync-auditor.md:108`. 독법 (B)의 기록된 대가이며(§E.2.0, §E.3 Gaps 1), 이 sync는
   **그것을 메우려고 예외를 넓히지 않았다.** 독트린 문언 수리는 **t1059**이고 **아직 착지하지
   않았다**. 이 카드 자신의 증거 5건은 이 글을 쓰는 시점에도 무시 상태다 —
   **워크트리를 폐기하면 유일본이 사라진다.**
2. **push·CI 판정이 없다.** 레인은 push하지 않는다(리드 일괄). 여기 있는 것은 전부 로컬
   조기 신호이고, 깨끗한 환경의 전체 스위트 판정도 darwin/windows 매트릭스 판정도 없다.
3. **`internal/spec` 테스트 실패가 그대로다.** develop에서 상속된 것이며(§E.3 Gaps 3) 이 카드의
   범위 밖이다. sync 단계도 고치지 않았다.
4. **형제 SPEC의 `updated:`가 갱신되지 않았다.** `SPEC-EVIDENCE-CITATION-CANON-001`은 본문에
   2026-09-21 HISTORY 항목을 얻었는데 frontmatter `updated:`는 `2026-08-31`이다. `status:`는
   `completed` 그대로이므로 전이가 아니라 **비전이 frontmatter 정정**이고, 그 소유자는
   manager-spec이다(`spec-frontmatter-schema.md` § Non-transition frontmatter corrections).
   **고치지 않고 기록만 남긴다.**
5. **docs-site 빌드를 이 sync에서 다시 돌리지 않았다.** M5의 `hugo` exit 0 / 경고 0을 재사용했고,
   이번 실행이 docs-site 파일을 하나도 건드리지 않았다는 사실이 그 재사용의 근거다. 새로 재지
   않았다는 것을 여기 적는다.
6. **워크트리 가드 거부 1건.** 네 산출물의 frontmatter를 `for f in …; do sed …` 한 줄로 읽으려던
   호출이 거부됐다(런타임 계산 값이 옵션 자리에 설 수 있는 형태). 평범한 개별 명령으로 나눠
   같은 것을 읽었고, 우회로를 만들지 않았다.

### Residual-risk — 관측했는데도 여전히 틀릴 수 있는 것

- **예외가 규칙 순서에 의존한다.** 누가 이 블록을 무리 B의 `plan-audit` 카브아웃 아래로 옮기면
  `plan-audit/verdict.md`가 다시 열리고, **어떤 테스트도 그 사실을 내지 않는다.** 주석에
  「must stay ABOVE」가 있지만 주석은 기계가 아니다.
- **문언 좁히기는 산문이라 회귀 가드가 없다.** 다음에 누가 넓은 주장을 다시 쓰는 것을 막는 기계는
  없다.
- **깊이 1 제한이 `.moai/reports/lead/<batch>/verdict.md`를 놓친다.** 의도된 배제이지만 리드 배치
  판정서는 지금도 반출되지 않는다.
- **CHANGELOG 항목은 이 sync가 읽은 것만 담는다.** 위 Gaps 2번 때문에 CI가 무엇을 말할지는
  이 항목에 없다.

---

## §J SPEC 본문 blocker 해소 (manager-spec, 2026-09-21)

> §E.3 Gaps 2번이 blocker로 넘긴 SPEC **본문** 편집을 이 항목이 닫는다. 소유권 경계상
> `spec.md` / `acceptance.md` / `design.md` 본문은 manager-spec의 표면이므로, 오케스트레이터
> 재위임으로 이 에이전트가 수행했다. 섹션 문자 `J`는 `spec-frontmatter-schema.md`
> § progress.md Section Map의 할당 규칙에 따라 새로 집은 것이다 — `§E.*` 네임스페이스는
> era.go가 파싱하므로 재사용하지 않는다.
>
> **귀속**: 이 워크트리 `.claude/worktrees/t1039`, 브랜치 `WT-evidence-path`,
> 착수 HEAD `7939e38b9`. **런타임 변경 0** — 고친 것은 산문뿐이다.

### §J.1 무엇을 고쳤나 (5건)

| # | 자리 | 처분 |
|---|---|---|
| 1 | `spec.md` §2.2 `REQ-EPE-006` | `[UNRESOLVED — 운영자 소관]` → **해소, 독법 (B)**. 출처(2026-09-21 run 단계 배차문)와 귀결(두 [HARD] Export mandate 잔존, t1059 소관)을 함께 명기 |
| 2 | `spec.md` §3.2 / `acceptance.md` `AC-EPE-002` [HARD] 블록 / `design.md` §1.3 D19 | **「(B) 아래에서는 해당 없음」으로 표시.** 삭제하지 않았다 — (A)가 검토된 뒤 채택되지 않았다는 사실이 읽혀야 한다 |
| 3 | `acceptance.md` §B.9 D15 [HARD] | **범위 한정.** §B.9 원장 규약이며 §B.5 P-프로브 셀에 소급되지 않음을 명시하고, 어떤 판정식이 표 셀에 올 수 있는지의 조건을 적었다 |
| 4 | `acceptance.md` §B.10 regression-guard 행 | **빠진 절을 채웠다.** 하위 점검의 차단력은 그것을 실은 AC의 처분에서 나온다 — `L-R1`·`L-R2`는 단독으로 릴리스를 막는다 |
| 5 | `acceptance.md` §B.1 / `AC-EPE-001` / `AC-EPE-003` | `REQ-EPE-005`를 이름 붙였다. 커버리지는 이미 실질적이었고 없던 것은 교차참조뿐 |

부수 정합: `spec.md` §3 머리말 · §3.4 · §7 범위 밖 2항목 · §8 잔여 위험 2항목, `design.md` §6.
전부 **(B) 확정의 귀결을 서술**하는 것이고, 결정 이전 문언은 보존했다.

### §J.2 [HARD] 넓히지 않았다

두 [HARD] Export mandate(`plan-auditor.md:601`, `sync-auditor.md:108`)는 **거짓인 채로 남는다.**
이 카드 자신의 plan-audit 판정서 4건과 `lane-measurements.md`도 무시된 채다. §E.3 Gaps 1번이
그 측정 기록이며, 이 개정은 **그 기록을 그대로 두었다** — 예외를 넓혀 문장을 참으로 만드는 것은
결정된 교환을 조용히 수선하는 것이고, 그것이 이 카드가 막으려는 실패 형태 그 자체다.
독트린 문언 수리는 **t1059**가 맡으며 **아직 착지하지 않았다**.

### §J.3 AC 계수 불일치 — 23으로 판정, 계수기 동작은 미설명

`acceptance.md`의 인수조건 수는 **23**이다. 고유 `AC-EPE-*` 토큰 23개
(`AC-EPE-001`~`AC-EPE-023`, **결번 없이 연속**)이며, §B.10 처분표의 합(6+6+7+2+2)과 일치한다.
Tier L 상한은 25이므로 **예산은 어느 쪽으로 읽어도 초과하지 않는다.**

[HARD] **Go 계수기가 24를 읽는 이유는 이 개정이 확립하지 않았다.** 그것은 판정이 아니라
**Gap**이며, 별도 카드가 맡는다. 이 판정을 만족시키려고 AC를 더하거나 빼거나 문서를 고치지
않았다 — **이해하지 못한 계수기에 문서를 맞추는 것**이 금지된 방향이다.

---

## §G Plan-phase Iteration Log

> 섹션 문자 `G`는 `spec-frontmatter-schema.md` § progress.md Section Map의 할당 규칙에 따라
> 새로 집은 것이다 — `§E.*` 네임스페이스는 era.go가 파싱하므로 재사용하지 않는다.
> (`§F`는 Mode Selection, `§H`는 Recursive Self-Diagnosis, `§I`는 Token Accounting에 이미 할당.)

### iteration 1 — 2026-09-20

최초 저작. Tier M으로 분류. plan-audit 결과 **FAIL, 0.725**(Tier M 임계 0.80 미달).
must-pass 7건은 전원 PASS(방화벽은 깨끗); FAIL은 집계 점수 + blocking 결함 6건에서 나왔다.
차원별: Clarity 0.75 / Completeness 0.70 / Testability 0.70 / Traceability 0.75.

### iteration 2 — 2026-09-20

blocking 6건 전부 + optional 3건 전부를 닫았다. 결함별 수정 내역:

| 결함 | 등급 | 무엇이었나 | 무엇을 했나 |
|---|---|---|---|
| **D1** | critical | docs-site 12파일/19행이 같은 거짓 주장을 담는데 SPEC이 포함도 제외도 하지 않음(`docs-site` 언급 0회) | **범위 안으로 편입.** `spec.md` §1.3(c)에 로케일 내성 인벤토리(12파일/20행/4로케일×3계열), §1.3(d)에 층별 합계표 신설. REQ-EPE-017(문언 갱신)·REQ-EPE-018(4-로케일 동시 수정) 신설, AC-EPE-022·023 신설, `plan.md` M5 신설. `spec.md` §7에 「비-주장 언급 29파일」 경계 명시. §1.3의 「수정 표면 4건」은 리터럴 주장 기준임을 명시하고 전체 합계 16으로 재서술 |
| **D2** | major | 순서 분석이 5무리 23행 중 2무리만 담음; 관용구 줄번호 3/4가 한 칸씩 어긋남 | `spec.md` §1.5를 **다섯 무리 전량 표**로 재작성(A 107-109 / B 224-230 / C 232-247 / D 290-295 / E 376). `-v` 측정으로 **`:294`가 `AC-EPE-004`의 대상 경로를 현재 결정**함을 명시하고, `:376`이 깊이 1 전용이라 `<card>/verdict.md`에 닿지 않음을 측정과 함께 서술. 관용구 범위 정정(t338 234-236 / t528 237-241 / t530 242-244 / t229 245-247 — 레인 기록도 틀렸으므로 복사하지 않고 직접 측정). `plan.md` M1.3을 「삽입 위치를 측정으로 정한다」로 재작성하고 `:294` 중복 여부를 측정 항목으로 세움. `design.md` §2~§3이 봉쇄 경로 (ㄱ)/(ㄴ)와 삽입 위치 후보를 설계 |
| **D3** | major | REQ 17(상한 16) / AC 20(상한 16) — Tier M 예산 양축 초과 | **Tier L로 재분류.** `spec.md` §0에 판별식과 3축 측정표 신설(파일 16>15, REQ 18>16, AC 23>16). 요구사항을 수를 맞추려 지우지 않았다 — 접은 것은 `REQ-EPE-008a` 하나이고 그것은 규모 축소가 아니라 **우발적 분리의 복원**(제자리 재측정과 「픽스처는 그 측정이 아니다」는 하나의 의무)이며 D7도 함께 닫는다. 금지의 3중 명문화는 유지. frontmatter `tier: L`, 산출물에 `design.md`·`research.md` 추가 |
| **D4** | major | `REQ-EPE-010`이 「깊이만 결정」이라면서 문언에 `verdict.md`(폭 보유 토큰)를 박음 | 문언을 **폭 중립**으로 재작성 — 「한 단계 깊이만 되살린다, 파일명 집합은 REQ-EPE-006이 정하고 이 요구사항은 참조만 한다」. `spec.md` §3.4를 **두 등급**(결정 의존 / 참조 의존)으로 재작성하고 `REQ-EPE-010`을 참조 의존 행으로 명시 — 「폭과 무관」 목록에서 빼되 결정 의존으로 승격시키지도 않았다. `AC-EPE-020`에 참조 의존 분류의 재확인 항목 추가. **폭 축은 닫지 않았다** |
| **D5** | major | `AC-EPE-012`가 공허 — `manager-lead.md`의 `tracked` 3행 중 `M<n>-report.md`를 이름 붙이는 행이 0행이라 착수 전부터 초록 | **세 프로브로 재-앵커.** P-a `:152`의 직접 추적 주장(RED-now 1), P-b `:154`의 `"and only the tracked path does"`(RED-now 1) — **이 문장이 `:161`의 접기 행을 추적 주장으로 만든다**, P-c `moai.md`의 `(tracked; exported before citing`(RED-now 2). 대체 문언의 §E.2 인용을 PASS 조건에 추가(지우기만 하면 「무엇이 참인가」가 사라진다). `REQ-EPE-002`의 좌표에 `:154` 추가하고 왜 `:161`이 낱말 grep에 안 잡히는지 명시. 세 행을 직접 읽고 잰 값이며 레인 요약을 인용하지 않았다 |
| **D6** | major | `REQ-EPE-011`(깊이 제한을 주석에 적을 의무)을 검증하는 AC 없음 | **`AC-EPE-021` 신설** — `grep -n 'reports/lead' .gitignore` ≥1행이고 놓치는 형태를 이름 붙일 것. RED-now **0**, 양성 대조 **23**. `REQ-EPE-011`에 「형태를 이름 붙여야 한다」를 명문화. `AC-EPE-021`/`005`/`010`이 각각 다른 대상(주석 존재 / 깊이 동작 / `:107-108` 주석)을 잰다는 구분을 AC 본문에 기록 |
| **D7** | minor | `REQ-EPE-008a` 접미 하위 ID | D3의 접기로 해소 — `REQ-EPE-008`의 `shall not` 절이 됐다. 재번호가 아니라 병합이므로 결번이 생기지 않는다 |
| **D8** | minor | §1.3 grep 블록이 자기-적중을 누락 | 블록 하단에 주석 1행 추가 — 지금 같은 명령은 6행이고 6번째가 이 SPEC 자신이며 수정 표면이 아니다. `AC-EPE-011`에도 같은 주의 추가 |
| **D9** | minor | 블랭킷 주석 좌표 `:225-228` ↔ 실제 `:224-226` | `spec.md` §1.1 표에서 `:224-226`으로 정정. `sed -n '222,233p'`로 직접 확인(`research.md` §1) |

리드 지시 두 건도 함께 반영:

- **A(자기-실사례를 동기 절로)** — `spec.md` **§1.0** 신설. 감사 판정서가 규정된 목적지에
  쓰였는데 그 목적지가 무시되어 `plan-auditor.md:601`의 [HARD] 문장이 **자신을 판정한 감사에
  대해 거짓**임을, `git status --porcelain --ignored`·`ls`·`check-ignore -v` 측정과 함께 적었다.
  `--ignored` 없는 `git status`의 0행이 부재가 아니라 미측정이라는 함정도 같은 절에 기록.
  잔여 위험(§8)에도 남겼다 — 여기서는 **결함이 실재한다는 증거**, 거기서는 **유일본 손실 경고**.
- **B(Tier 근거 재도출)** — 종전의 「bounded, enumerable edit surface」류 서술을 그대로 옮기지
  않고 `spec.md` §0에서 **지금 서 있는 표면으로부터 3축을 다시 재어** 도출했다. 결론은 Tier L.

**감사가 「건드리지 말라」고 한 것은 건드리지 않았다**: §A 판별식 규약과 두 함정 봉쇄, 픽스처
금지의 3중 명문화, 트리 귀속 규율, 드리프트 수치를 AC에 넣지 않은 판단, `AC-EPE-002`/`013`의
폭 매개변수화, §4의 선행 SPEC 화해 3요소, §3.0의 양쪽-독법 귀결표. **폭 축은 여전히 열려 있다.**

### iteration 3 — 2026-09-20 (이 개정)

iter2 판정은 **PASS-WITH-DEBT, 0.85**(Tier L 임계 0.85 — 여유 0.00)였고 blocking 3건 + optional
2건을 남겼다. 차원별: Clarity 0.85 / Completeness 0.85 / **Testability 0.70(두 판 연속 정체)** /
Traceability 1.00. blocking을 들고 run 단계에 들어가지 않는다는 규율에 따라 전량을 닫았다.

| 결함 | 등급 | 무엇이었나 | 무엇을 했나 |
|---|---|---|---|
| **D10** | critical | `AC-EPE-022`의 기계 하위 점검이 **착수 전부터 0**(공허)이고, RED-now로 적힌 **18이 어떤 명령의 출력도 아님** | **하위 점검을 두 갈래로 재구성.** S1 = `\.moai/reports/[^/ \`]+/M[^ \`]*\.(log\|md)` (형태 i, RED-now **8**), S2 = `\.moai/reports/[^/ \`]+/\`` (형태 ii, RED-now **12**). **8 + 12 = 20 = 인벤토리 전량**이라는 항등식을 PASS 구조에 넣어 「하위 점검이 인벤토리의 어느 부분도 놓치지 않는다」를 기계로 보이게 했다. 추가로 **판정서 파일을 이름 붙이는 행 = 0**을 측정해, 20행 전부가 위반이고 그대로 둘 행이 없음을 확정(진짜 빨간 관측). 음성 대조 0, 양성 대조 41 동반. `acceptance.md`·`research.md` §5.4의 수치를 **명령을 다시 돌려 출력을 옮기는 방식으로** 전량 교체했고, `plan.md` M5 2항도 같은 형태로 고쳤다 |
| **D11** | major | Tier를 결정하는 수정 파일 수 **16**이 자기 구성 행(2+2+4+12=**20**)과 모순 | 네 곳(`spec.md` §0·§1.3(d), `research.md` §7, `plan.md` §A)을 **20**으로 정정. `research.md` §7의 산식도 중복 빼기 없이 **파일을 직접 열거**하는 형태로 다시 썼다(그 산식의 구성 항을 그대로 더해도 20이었다). `spec.md` §1.3(d) 표에 **실제 경로 열**을 추가해 합계가 다시 어긋나면 눈에 띄게 했다. 기계 재생성물 `.codex/…/manager-lead.toml`은 **세지 않는다**는 처분을 명문화(손편집 대상이 아니며 `AC-EPE-015`가 그 축을 잰다). **Tier 결론 불변 — 20 > 15로 더 확실해졌다** |
| **D12** | major | `REQ-EPE-002`의 본문(비-판정서 **파일** 3종)과 좌표 목록(**디렉터리**를 주석하는 `manager-lead.md:63` 포함)이 어긋나고, `:63`에 닿는 AC가 없음 | **본문을 넓히는 쪽으로 해소.** `REQ-EPE-002`를 **두 형태**(i 비-판정서 파일명 / ii 파일명 없는 `<card-id>/` 디렉터리 단독)로 재작성 — 이 둘이 곧 `AC-EPE-022`가 docs-site에서 세는 두 형태이므로, 같은 형태를 한 표면에서는 위반으로 세고 다른 표면에서는 비워 두는 비대칭이 사라진다. `AC-EPE-012`에 **P-d** 추가(`grep -c 'the deciding lines to the tracked' …manager-lead.md` → RED-now **1**, 미러에도 1건), 양성 대조 5 / 음성 대조 0 동반. `plan.md` M2의 `:63` 항목을 「함께 본다」에서 **P-d가 겨누는 좌표**로 승격 |
| **D13** | minor(optional) | `AC-EPE-002` 독법 (A)가 계열을 구체 파일명 `plan-audit-iter1.md`로 열거 — 문자 그대로면 **iter2 판정서가 되살아나지 않음** | (A) 집합을 **계열 표기**(`plan-audit-iter<N>.md`)로 바꾸고, 구현이 파일명당 negation이 아니라 **계열 글롭** `!.moai/reports/*/plan-audit-iter*.md`가 된다는 것을 명시. P1 프로브는 계열의 실재 구성원 **둘 이상**(iter1 + iter2)을 겨누도록 요구 — 하나만 재면 계열이 아니라 그 파일 하나만 검증된다 |
| **D14** | minor(optional) | 23개 AC 중 다수가 RED-now도 regression-guard 분류도 담지 않음 — 처분이 **적혀 있지 않음** | `acceptance.md` **§B.10 처분 분류표** 신설 — 23개 전량을 5부류로 배정(release-blocking 7 / regression-guard 6 / post-change-only 6 / width-parameterized 2 / process-record 2, 합 23). `verification-completeness.md` §2·§2.1의 기존 의무를 **명시**할 뿐 새 의무를 만들지 않는다. 덤으로 `AC-EPE-008`에 실측 RED-now(1, 양성 대조 2)를 채워 release-blocking으로 올렸다 |

**남겨야 할 관측 두 가지** — 이번 결함이 어디서 왔는지가 다음 사람에게 쓸모 있기 때문에 적는다.

1. **수리가 자기가 고치던 결함 계열을 재생산했다.** iter1의 D5는 「`AC-EPE-012`가 착수 전부터
   초록이라 공허하다」였다. 그 D5를 닫은 같은 개정이, D1(docs-site 편입)의 수리로 도입한
   `AC-EPE-022`에 **똑같은 성질**을 실었다 — grep 파이프에 이어 붙인 하위 점검이 착수 전부터 0.
   자기 비난으로 적는 것이 아니라, **grep 파이프를 상대로 인수조건을 쓰는 자리가 두 번 연속
   틀어진 지점**이라는 사실이 다음 저자에게 필요하기 때문이다. 이번 수리가 그 자리에 놓은
   대비책은 「토큰을 리터럴로 겨누지 말 것」과 「분할 항등식(S1+S2=전량)으로 누락을 기계로
   드러낼 것」 둘이다.
2. **지어낸 수치가 서술 층이 아니라 검증 층에 있었다.** 보고서 칸을 채우려는 압력이 수치를
   만들어 내고 하필 판정을 결정하는 값이 지어진다는 것은 이 저장소가 이미 기록한 실패
   형태다. 이번이 새로운 점은 **그것이 보고서가 아니라 인수조건 안에 있었다**는 것이다.
   인수조건은 작업을 **서술**하는 문서가 아니라 작업이 판정되는 **기준**이므로, 거짓 수치가
   거기 있으면 보고를 틀리게 하는 데 그치지 않고 **틀린 기준을 세운다** — 그리고 그 기준은
   run 단계가 통과했다고 말하는 순간까지 아무 신호도 내지 않는다.

**열린 채로 둔 것**(iteration 3이 건드리지 않았다):

- **폭 축(`REQ-EPE-006`)은 여전히 운영자 소관이며 미해소다.** 이 개정의 어떤 변경도 (A)/(B)를
  확정하지 않는다 — D13의 계열 표기는 **(A)를 택했을 때의 구현 형태**를 정확히 적은 것이지
  (A)를 고른 것이 아니다((B)에서는 그 글롭이 존재하지 않는다). plan은 이 축이 열린 채로 닫힌다.
- **예외 규칙의 제자리 효과**는 `.gitignore` 편집을 요구하므로 run 단계의 일이다.
- **서브에이전트 쓰기 범위 가드 부재**는 범위 밖(운영자·리드 판정).
- **docs-site 20행 인벤토리를 19가 아니라 20으로 읽는 판결**과 `AC-EPE-022`의 정확-수치
  비단언 문장은 그대로 두었다.
- **`spec.md`의 규모를 이유로 한 재구조화는 하지 않았다** — iter2 감사가 「결함 아님」으로
  판정했고, 없는 규칙을 만들지 않는다.

**감사가 「그대로 두라」고 한 것은 그대로 두었다**: §A 판별식 규약과 두 함정 봉쇄, 픽스처-금지
4중 명문화, 트리 귀속 [HARD] 블록, §1.5 다섯 무리 표, §1.0의 자기-실사례와 「0행 = 미측정」
경고, §3.0 귀결표와 §3.4 두 등급 봉쇄, D5 재-앵커의 `:154 → :161` 논증, `AC-EPE-021`의
RED-now/대조 쌍, `design.md` §2의 (ㄱ)/(ㄴ) 양립 설계.

**이번 개정이 쓴 파일**: `spec.md` / `plan.md` / `acceptance.md` / `research.md` / `progress.md`
(이 파일) 5종뿐. `.gitignore`(417줄 무손상), `.moai/reports/t1039/**`(레인 기록·iter1·iter2
판정서), 저장소 본체는 **읽기만 했다**.

### iteration 4 — 2026-09-21 (이 개정)

iter3 판정은 **PASS-WITH-DEBT, 0.90**(Tier L 임계 0.85 — 여유 +0.05)이었고 단조 증가
0.725 → 0.85 → 0.90, must-pass 7/7, iter2 blocking 3건 전량 해소가 측정으로 확인됐다.
감사는 iteration 4를 쓰지 말고 신규 3건을 **실행 게이트 debt**로 두라고 권고했다. **레인은
그 권고를 뒤집었다** — 감사 자신이 blocking 으로 이름 붙인 결함은 debt 가 아니고, 같은 저작
위치에서 네 번째 실패가 예고된 상태에는 기록이 아니라 판정이 필요하다. 다만 수리 범위는
좁게 유지했다(재작성 아님).

| 결함 | 등급 | 무엇이었나 | 무엇을 했나 |
|---|---|---|---|
| **D15** | major | `AC-EPE-022`의 **정본 carrier가 markdown 표 셀**이라 S1이 구성상 `0`을 냄. 실측: 실행본 → 8(exit 0), 표 셀 이스케이프본 → 0(exit 1). `(log\|md)`가 ERE에서 리터럴 `log|md`가 된다 | **carrier를 옮겼다.** §B.9에 **판정식 원장**(fenced, id 부여 `L-S1`/`L-S2`/`L-VD`/`L-R1`/`L-R2`)을 신설하고, 표는 command 대신 **id만 인용**한다. 이스케이프를 다시 고르는 수리를 명시적으로 거부했고, 거부 이유(같은 층의 세 번째 시도)를 본문에 남겼다. `plan.md` M5에 「원장에서 복사해 실행한다」 [HARD] 추가 |
| **D16** | major | 분할 항등식 `8 + 12 = 20`이 배타·남김없음을 **함의하지 않고**, green 시점 형태 `0 + 0 = 0`은 **동어반복**. 두 점검 어느 쪽도 분류하지 않는 위반 형태(`report.md` 계열)가 조용히 생존 | **항등식을 PASS 구조에서 뺐다.** `L-R1`(잔차) / `L-R2`(중복)를 신설해 PASS 조건을 **넷**으로 만들었다 — `S1=0 · S2=0 · L-R1=0 · L-R2=0`. 둘 다 착수 전 0인 **regression-guard**로 분류(공허가 아니라 부류의 정의). `L-R1` 변이 프로브 양방향 확인: `report.md` 형태 → **1**, 정상 S1 행 → **0**. `research.md` §5.4·`plan.md` M5의 같은 추론도 정정 |
| **D17** | minor | `AC-EPE-023`이 §B.10에서 release-blocking 인데 자기 RED-now 가 「**계측 불가**」를 선언 | **post-change-only 로 한 칸 이동.** 그 부류의 정의가 `AC-EPE-023`의 자기 서술과 축자 일치한다. 합계 **23 불변**(6+6+7+2+2). 고친 방향은 「자기 서술을 문언 그대로 따른다」이고, 측정 불가를 측정된 것처럼 고쳐 쓰는 반대 방향이 아니다 |
| **D18** | minor(optional) | `AC-EPE-002` P1 의 「구성원 **둘 이상**」 요구가 폭-매개변수화 AC 안에서 무조건문으로 읽힘 | **(A) 조건 안으로 넣었다** — (B)에서는 계열 글롭이 존재하지 않아 겨눌 구성원이 없다는 것을 명시. **폭을 고르지 않았다**: 어느 독법에서 요구가 발화하는지만 적었다 |
| **D19** | minor(optional) | `design.md` §1.3이 계열-글롭 재해석을 따라 갱신되지 않음 | §1.3 채택 절에 [HARD] 계열 예외 한 블록 추가(`!.moai/reports/*/plan-audit-iter*.md`), `acceptance.md`와 양방향 교차참조. 여기서도 「(A)를 택했을 때의 구현 형태」로 조건 서술 |

**[HARD] 같은 저작 위치에서 세 판 연속 실패했다 — 이번 수리가 다른 종류인 이유.**

위 §G가 이미 「grep 파이프를 상대로 인수조건을 쓰는 자리가 두 번 연속 틀어졌다」를 기록했다.
D15는 그 자리의 **세 번째** 실패다. 계보는 이렇다:

1. **iter1 D5** — `AC-EPE-012`가 착수 전부터 초록(**공허**).
2. **iter2 D10** — `AC-EPE-022`가 착수 전부터 0(**공허**) + RED-now 18이 **지어낸 값**.
3. **iter3 D15** — `AC-EPE-022`가 실측으로 빨갛지만, **정본 carrier가 실행 시점에 `0`을 낸다.**

무게는 매 판 내려왔지만(공허+거짓 → 공허 → 둘 다 아님) **자리는 같다**. 앞의 두 수리는 둘 다
**같은 층의 치환**이었다 — 정규식을 더 나은 정규식으로 바꿨다. 그래서 세 번째가 왔다: 정규식이
옳아져도 **그 정규식이 놓인 자리**(markdown 표 셀)가 실행본을 훼손하는 층은 손대지 않았기
때문이다.

이번 수리는 치환이 아니라 **층 이동**이다. D15는 문자열을 고치는 대신 정본이 사는 자리를
표 셀 → fenced 원장으로 옮겨, 이스케이프가 개입할 여지 자체를 없앴다. D16도 같은 성격이다 —
더 나은 항등식을 고르는 대신 **항등식이라는 증명 형태를 버리고** 잔차 재측정으로 바꿨다.
치환이면 네 번째가 오고, 층을 바꾸면 그 자리의 실패 경로가 닫힌다는 것이 이 구분의 요지다.

**[HARD] 감사 권고를 뒤집은 근거.** iter3 감사는 iteration 4를 열지 말라고 권고했고, 그
권고의 논거(집계가 여유를 두고 임계를 넘었다)는 참이다. 그러나 감사는 같은 문서에서 D15·D16을
**blocking**으로, D15를 「`AC-EPE-022` 실행 **전에** 닫는다」로 적었다 — 즉 debt 로 두면 run
단계가 그 AC 를 실행할 수 없다. 실행 불가 상태로 plan 을 닫는 것과 debt 를 안고 닫는 것은
다른 일이고, 전자는 닫는 것이 아니다.

**열린 채로 둔 것**(iteration 4가 건드리지 않았다):

- **폭 축(`REQ-EPE-006`)은 여전히 운영자 소관이며 미해소다.** D18·D19는 둘 다 폭에 가장
  가까이 간 수리이므로 특히 조심해서 **조건 서술**로만 적었다 — 「(A)를 택한 경우」이지
  (A)를 고른 것이 아니다. §3.0 귀결표·§3.3「권고는 결정이 아니다」·`REQ-EPE-006`의
  `[UNRESOLVED]`·`AC-EPE-019`의 암묵-확정 FAIL 조건·§B.10의 width-parameterized 분류는
  **한 글자도 건드리지 않았다**.
- **예외 규칙의 제자리 효과**는 여전히 미측정(run 단계).
- **서브에이전트 쓰기 범위 가드 부재**는 범위 밖.
- **`spec.md`는 이번 개정에서 열지 않았다** — 규모를 이유로 한 재구조화도, 다른 수정도 없다.

**감사가 「그대로 두라」고 한 것은 그대로 두었다**: §A 판별식 규약과 세 함정 봉쇄, 픽스처-금지
4중 명문화, 트리 귀속 [HARD] 블록, §1.5 다섯 무리 표, §1.0의 자기-실사례와 「0행 = 미측정」
경고, §3.0 귀결표와 §3.4 두 등급 봉쇄, D5 재-앵커의 `:154 → :161` 논증, D12를 「본문 넓히기」로
해소한 판단, `AC-EPE-021`의 RED-now/대조 쌍, `design.md` §2의 (ㄱ)/(ㄴ) 양립 설계,
§B.10 분류표의 나머지 22칸.

**이번 개정이 쓴 파일**: `acceptance.md` / `design.md` / `plan.md` / `research.md` /
`progress.md`(이 파일) 5종. **`spec.md`는 열지 않았다.** `.gitignore`(417줄 무손상),
`.moai/reports/t1039/**`(레인 기록·iter1·iter2·iter3 판정서), 저장소 본체는 **읽기만 했다**.
