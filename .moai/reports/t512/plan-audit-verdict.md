# Plan-Audit Verdict — SPEC-WORKTREE-GUARD-HEREDOC-DOC-001 (card t512, GH #1659)

- Auditor: plan-auditor (adversarial, security-adjacent lens)
- Tier M · PASS threshold 0.80
- Audited tree: `.claude/worktrees/t512` · branch `WT-guard-heredoc` · HEAD `6a46c0edb` (카드 커밋 0개)
- Date: 2026-09-07

**최종 판정: PASS — 0.91 / 1.00 (Tier M 임계 0.80 상회)**

차단 결함(Major) 0건. Minor 5건, 관측(note) 4건 — 전부 run/sync에서 흡수 가능한
정합성·충실도 항목이며, 어떤 AC의 판별력도 무효화하지 않는다.

---

## Claim

1. SPEC 4종 + progress.md 의 트리 대비 사실관계(독트린 인용 행, 미러 패리티,
   미러 테스트 존재, 앵커 RED 값)가 전부 참이다.
2. 피벗(pivot)이 정직하다 — SPEC 어디에도 가드를 "고친다"는 주장이 없고,
   REQ-WGHD-003이 문서=수리 프레임을 금지한다.
3. 두 구분자(quoted 불확성 / unquoted 확장) 케이스가 M1 소절 초안과 앵커에
   양쪽으로 실려 있다.
4. 미러 규율(같은 커밋 바이트 동일 + 템플릿 중립)이 명시적이며, 로컬판 인용
   여부의 비대칭은 "양쪽 모두 중립"으로 더 엄격하게 해소돼 있다.
5. 범위가 Go 변경 전체(branch_guard.go 포함)를 근거와 함께 밖으로 뺐다.
6. REQ 7건 ↔ 작업 T1-T4/M1-M3 ↔ AC 7행 추적이 완결이고, blame-free [HARD]가
   M2와 AC 양쪽에 실려 있다.
7. 프론트매터 12 정식 필드 + 형제 무상태(statelessness) + §E.1 신호가 규격대로다.

## Evidence (본 세션 직접 측정 — 전부 이 트리, HEAD `6a46c0edb`)

### 트리·구조 검증

| 측정 | 명령 | 관측 |
|---|---|---|
| toplevel/브랜치/HEAD | `git rev-parse --show-toplevel` · `git branch --show-current` · `git rev-parse --short HEAD` | `.claude/worktrees/t512` · `WT-guard-heredoc` · `6a46c0edb` — 디스패치 전제와 일치 |
| 카드 커밋 0개 | `git log --oneline -3` | 선두 커밋 `6a46c0edb` (develop 병합 커밋), 그 위 카드 커밋 없음 |
| 미러 존재+패리티 | `cmp <live> <mirror>` | exit 0 (바이트 동일) — REQ-WGHD-004 베이스라인 참 |
| branch_guard.go 존재 | `ls internal/hook/branch_guard.go` | 존재 (27,375바이트) — 편집 금지 대상 실재 확인 |

### 독트린 인용 행 대조 (SPEC §1.2 / research §2 의 행번호 주장)

`grep -n 'too complex to verify'` — **두 트리 모두 동일 행번호에서 3건**, 소스 구현 0건:

```
.claude/rules/moai/workflow/worktree-integration.md:481   (소속 표 — The Claude Code binary, "No — there is no source here that implements or configures it")
.claude/rules/moai/workflow/worktree-integration.md:485   (branch_guard.go 는 별개 가드 — 편집 무효)
.claude/rules/moai/workflow/worktree-integration.md:491   (거절 전문 인용)
internal/template/templates/.../worktree-integration.md:481,485,491  (미러 — 바이트 동일이므로 동일 행)
```

SPEC 이 주장한 `:481/:485/:491` 이 **그대로** 맞았다. `internal/`·`.claude/hooks/`
스윕에서 소스 히트 0 — "결함의 몸은 바이너리" 소속 판별의 재측정 성립.

### 앵커 재측정 (acceptance.md RED-now 셀 대조)

| 앵커 | acceptance.md 기재 | 본 세션 재측정 | 일치 |
|---|---|---|---|
| `grep -c "brace expansion"` (AC-002) | `0` / rc=1 | `0` / rc=1 | YES |
| `grep -c "unquoted"` (AC-002) | `0` / rc=1 | `0` / rc=1 | YES |
| `grep -c "re-measured"` (AC-003) | `0` / rc=1 | `0` / rc=1 | YES |
| `grep -c "1659" <live> <mirror>` (AC-005) | `:0` ×2 / rc=1 | `:0` ×2 / rc=1 | YES (단, 출력 순서 불일치 — Minor F1) |
| `grep -c "SPEC-WORKTREE-GUARD-HEREDOC"` (AC-005) | `0` / rc=1 | `0` / rc=1 | YES |
| `ls reporter-reply-1659.md upstream-draft-claude-code.md` (AC-006/007) | 부재 ×2 / rc=1 | 부재 ×2 / rc=1 | YES (단, stderr/stdout 구분 — Minor F3) |

### 미러 테스트 기계론 (research §4 주장 대조)

`rule_template_mirror_test.go` 실측: `TestRuleTemplateMirrorDrift` **:137** ·
sentinel `RULE_TEMPLATE_MIRROR_DRIFT` · allowlist **:57** 에
`".claude/rules/moai/workflow/worktree-integration.md"` 명시 — research.md 가
기재한 행번호 3개가 전부 정확했다.

### 중립성 정밀 측정 (광의 패턴의 오탐 분별)

`grep -cE "1659|SPEC-[A-Z]|2026-09"` 를 양쪽 트리에 돌리면 **7건**이 나오지만,
전원이 `SPEC-XXX`/`SPEC-ID` **제네릭 플레이스홀더**다
(행 45/387/436/520/523/525/527) — 내부 SPEC ID가 아니므로 중립성 위반이 아니다.
AC의 스코프드 앵커(`1659`, `SPEC-WORKTREE-GUARD-HEREDOC`)가 올바르게 좁혀져 있다.
plan.md T1 삽입 초안(63-93행)도 광의 패턴으로 0건 — 삽입물 자체가 중립이다.

###develop 선행 관측

`git rev-list --count --left-right develop...HEAD` → `23 0` (develop이 23커밋
앞섬, `e0c904f58`). `git log HEAD..develop -- <독트린 2파일>` → **0커밋** —
선진 23커밋 중 어느 것도 M1 편집 대상 파일을 건드리지 않았다 (병합 창 충돌
위험 없음).

## Baseline-attribution

- 전 측정은 **이 실행, 이 트리**(`.claude/worktrees/t512`, HEAD `6a46c0edb`,
  카드 커밋 0개)에서 직접 수행됐다. SPEC 문서들이 핀한 트리와 동일 트리다.
- 디스패치 전제 "HEAD `6a46c0edb` = local develop tip"은 plan 작성 시점엔 참이었으나
  현재 develop은 `e0c904f58`로 +23 진행됐다. SPEC 문서 자체는 "plan 작성 시점 HEAD,
  카드 커밋 0개"로 정확히 규정해 어떤 핀도 현재-팁 주장을 하지 않으므로 문서 내
  모순은 없다(관측 N1).
- 재현 원장 `.moai/reports/t512/repro-heredoc-brace.md`는 존재하며 명령 verbatim +
  거절 문면 verbatim + 트리/HEAD 핀을 운반한다(축적된 1차 증거 — 본 감사는 이를
  재실행하지 않고 원장 대조로 검증했다. 미실행 증명·프로브 2 대조군 기재 포함).
- 미추적 상태: `git status --short` → `?? .moai/reports/t512/`,
  `?? .moai/specs/SPEC-.../` — research.md 관측 차이 1의 기재와 일치하며 plan.md
  M1 커밋 규약이 원장 동봉을 이미 규정했다.

## Findings

### Major — 없음

### Minor

**F1 — AC-WGHD-005 RED 셀의 출력 순서가 실측과 다르다**
(acceptance.md:153-156). 명령은 live 먼저 나열하는데 셀의 출력은
mirror 먼저 적혀 있다. 본 세션 실측 출력 순서는 **live 먼저**(인자 순서대로)다.
카운트 자체는 양쪽 :0으로 참이지만 verification-completeness §2.1 "verbatim =
raw 출력" 요건에 비춰 손봐야 한다. → run/sync에서 셀 출력을 실측 순서로 정정.

**F2 — AC-WGHD-002의 결합 앵커 임계(≥1)는 quoted-불릿 탈락 뮤턴트를 기계적으로 못 잡는다**
(acceptance.md:91-92 vs plan.md:98). acceptance.md(SSOT)는 두 앵커 각각 ≥1만
요구한다. quoted 불릿만 떼는 뮤턴트는 `unquoted` ≥1, `brace expansion` =1(≥1),
`re-measured` ≥1로 **전 앵커 통과**한다 — AC가 명시한 "한쪽 뮤턴트가 이 AC를
이기지 못한다"는 설계 문장은 unquoted-탈락 방향에만 참이다. plan.md T1의
`>= 2`가 양쪽 불릿이 각각 정확히 1개씩 `brace expansion`을 실으므로 **양방향
단일-불릿 뮤턴트를 모두 잡는 값**이다. 리뷰 셀(AC-002 리뷰 문단 (a))이 이를
보완하므로 치명적이지 않으나, SSOT 숫자는 ≥2로 상향해야 기계층이 독립 선다.
→ acceptance.md AC-002 GREEN을 "각각 ≥1, `brace expansion`은 ≥2"로 확정 권고.

**F3 — AC-WGHD-006/007의 ls RED 셀이 stderr를 stdout처럼 적었다** (acceptance.md:178-181).
BSD ls의 `No such file or directory`는 stderr로 나온다(본 세션 실측 동일).
rc=1은 정확하나 출력 스트림 구분 표기가 없다. §2.1 충실도 관점에서 "stderr 2행 +
rc=1"로 정정.

**F4 — AC-WGHD-002의 "§2.0" 교차참조가 수형(懸空)이다** (acceptance.md:92).
acceptance.md에는 §2.0 절이 없다 — 지시 대상은 서문 문단(뮤턴트-저항 설계
문단)이다. 참조를 "본 문서 서문 문단"으로 정정.

**F5 — T1 삽입 초안의 규범 문장이 독트린 관측 문체를 넘선다** (plan.md:83-84 —
"The fold already applied to command substitutions in that same position is the
correct treatment for the whole body."). 이 문장은 수정 방향 1(M3 (c))의 논거로,
독트린(관측 기록)보다 업스트림 초안의 자리가 맞다. 가드 약화는 아니다(증명
가능한 불확성 위치의 정밀화 방향이고, unquoted 쪽은 "거절이 옳다"로 경계를
유지) — 그러나 REQ-WGHD-002가 요구하는 "중립 문체" 관점에서 run-phase에서
관측형으로 완화할 것을 권고한다. 예: "the fold applied to command substitutions
in that same position is consistent with treating the whole body as inert."
불변(두 구분자 의미 + loosening-danger 경계)은 건드리지 않는 문장 단위 조정으로
plan.md가 이미 허용한다(60-61행).

### Observations (결함 아님 — run/sync 참고)

**N1 — develop +23 (`e0c904f58`)**: 병합 창에서 `git merge origin/develop` 흡수가
필요하다(레인 프로토콜상 이미 절차에 있음). 독트린 파일 충돌 위험 0 (실측).
SPEC 핀들은 전부 `6a46c0edb` 트리-상태 핀으로 유효하다.

**N2 — 광의 중립성 grep 금지**: run-phase가 `SPEC-[A-Z]` 같은 넓은 패턴으로
중립성을 검증하면 기존 `SPEC-XXX` 플레이스홀더 7건에 오탐낸다. AC의 스코프드
앵커를 그대로 쓸 것.

**N3 — research §2 히트 경로 표기**: 3건의 히트 실측 위치는
`internal/template/templates/` 미러 쪽이다(swept set이 `internal/`이므로).
research는 live 경로를 적었다 — 바이트 동일이라 행번호까지 동일해 실질 동치이나,
미래 독자를 위해 "양 트리 동일 행"으로 표기하면 더 정확하다.

**N4 — Tier 표기 정직성 긍정**: spec.md §1.4가 LOC·파일 수는 Tier S 범위임을
스스로 밝히고 디스패치 지시로 Tier M을 채택한 근거를 본문에 남겼다 — 프론트매터
`tier: M`과 일치. 위장 없음.

## 축별 판정 (디스패치 8축)

| 축 | 판정 | 근거 |
|---|---|---|
| 1. two-cell | PASS | 원장(거절 RED) 존재+핀 확인; AC 5종 RED-now 4요소(cmd/verbatim/rc/SHA) 갖춤; AC-004 regression-guard 정직 분류; AC-001 불변형의 뮤턴트-관측 RED도 채택 의무로 명시(F2의 임계 비대칭만 Minor) |
| 2. pivot honesty | PASS | 가드 "수리" 주장 0; REQ-WGHD-003·Goal 3·§1.4가 문서≠수리 명시; M1 초안에 약화 조언 0(오히려 unquoted 거절을 "옳은 동작"으로 명시); 허용 우회는 거절 메시지 안내+Write 도구뿐 |
| 3. 양 구분자 | PASS | quoted 불확성(plan.md:69-74) + unquoted 확장(:75-79) 양쪽; AC-002가 unquoted-탈락 뮤턴트를 앵커로 잡음 — quoted-탈락 방향만 F2 |
| 4. 미러 규율 | PASS | REQ-004/005 + cmp exit 0 베이스라인 + 같은 커밋 규약; 로컬-인용 허용 비대칭은 "양쪽 중립"으로 더 엄격히 해소됐음을 REQ-005 근거에 명시 |
| 5. scope | PASS | Non-Goals §3.1이 모든 Go 변경(branch_guard.go 포함)을 근거와 함께 배제; plan §5 금지 목록; t511·원장 편집 금지; 원장은 경로 인용만 |
| 6. traceability | PASS | REQ 7건 전원이 작업 "Implements" + AC "Implements"로 양방향 연결; blame-free [HARD]가 plan.md:153-155 + AC-006(e) 양쪽에 존재 |
| 7. frontmatter/statelessness | PASS | spec.md 12 정식 필드(별칭 0, 옵션 tier/issue_number는 스키마 허용); plan/acceptance/research `status:` 0; progress.md §E.1 `plan_status: audit-ready` + 핀 |
| 8. 사실 대조 | PASS | :481/:485/:491 실측 일치(양 트리); 원장 수치와 SPEC 인용 일치; 양 트리에 독트린 파일 존재 + 바이트 동일; 미러 테스트 :137/:57 실측 일치 |

## Gaps

- 거절 프로브 재실행은 하지 않았다 — 원장이 명령+verbatim+핀을 운반하고, 디스패치가
  재실행을 허용(필수 아님)으로 규정했기 때문이다. 바이너리 거절의 현재-세션 재현성은
  원장 의존으로 남는다(세션 동작이라 커밋 트리로 재증명 불가 — 본질적 한계).
- `go test ./internal/template/ -run TestRuleTemplateMirrorDrift`의 현재 PASS는
  실행하지 않았다 — plan-phase 기준점은 `cmp` exit 0(실행, 통과)으로 충족되고,
  테스트 실행은 run-phase T2 검증 라인이므로 이중 실행이 아니다.
- develop +23 커밋의 **내용 적합성**(t508/t527 등이 이 SPEC과 충돌하는지)은 파일
  접촉 0커밋 실측으로 충돌 없음만 확인했다; 의미적 충돌 여부는 병합 창에서 재판정.
- #1659/#1658 이슈 본문은 읽지 않았다 — research §6의 제보자 인용은 디스패치
  확립 사실로 수용했다(연구 규칙상 재연구 제외 명시).

## Residual-risk

- **F2 미수정 시**: sync-auditor가 SSOT(≥1)만 기계 집행하면 quoted 불릿 결여 판이
  앵커층을 통과할 수 있다. 리뷰 셀이 잡아주지만 리뷰는 사람/에이전트 판정이라
  기계 불변보다 약하다. 한 줄 수정(≥2)으로 봉합된다.
- **바이너리 재현의 비내구성**: 원장의 거절 재현은 Claude Code 세션 동작이라
  이 트리의 어떤 커밋으로도 재증명할 수 없다 — 업스트림 초안(M3)이 이 증거를
  인용할 때 "세션 관측"으로 등재하는 표현이 유지되어야 한다(plan/acceptance는
  이미 그렇게 쓰여 있다).
- **develop 진행**: 병합 창까지 추가 레인이 독트린 파일을 건드리면 본 감사의
  "충돌 위험 0"이 만료된다 — 흡수 후 M1 전에 `git log develop..` 로 재확인 권장.
- M1 초안 문안의 문장 단위 재배열 허용(plan.md:60-61)은 합리적 유연성이지만,
  AC-002의 앵커 키워드(`brace expansion` ×2 / `unquoted` / `re-measured`)가
  재배열에서 살아남는 것이 기계층의 전부다 — 문안 확정 시 앵커 보존을 우선할 것.

---

## 스코어 (Tier M)

| 차원 | 가중 | 점수 |
|---|---|---|
| 트리 사실관계 정확도 (행번호·테스트·앵커 전원 실측 일치) | 0.30 | 1.00 |
| two-cell·수용 규율 (정직 분류 포함, F2·F3 공제) | 0.25 | 0.80 |
| 피벗 정직성·안전 경계 (loosening-danger) | 0.20 | 1.00 |
| 추적성·범위·프론트매터 규격 | 0.15 | 1.00 |
| 내부 정합성 (F1·F4·F5 공제) | 0.10 | 0.70 |

**합계: 0.30×1.00 + 0.25×0.80 + 0.20×1.00 + 0.15×1.00 + 0.10×0.70 = 0.91**

**PASS (≥ 0.80)** — run-phase 진입 가능. F1-F5는 run/sync에서 흡수하는
비차단 정정 권고다.

## 가장 하중 높은 약점 (단일)

**acceptance.md AC-WGHD-002의 결합 앵커 임계 ≥1이 quoted-불릿 탈락 뮤턴트를
기계적으로 통과시킨다** (acceptance.md:91-92). SPEC이 선언한 뮤턴트-저항 설계는
unquoted-탈락 방향에서만 성립하고, 반대 방향은 plan.md:98의 `>= 2`만이 잡는다.
acceptance.md가 SSOT인 이상, 임계를 ≥2로 올리는 한 줄이 이 감사에서 나온
정정 중 판별력에 실제로 닿는 유일한 것이다.

---

# Iteration 2 — 수리 검증 (verification-only, 2026-09-07)

**최종 판정: PASS — 0.98 / 1.00 (Tier M 임계 0.80 상회). Minor 5건 전원
착지 확인, 신규 결함 0건.** 본 iter2 판정이 iter1 판정 이후 산출물 변경에 대한
재판정이다(artifact-hash 불일치 → skip policy 3조건 미충족 — progress.md:10 이
이 사실을 정직하게 선기록했다).

## Claim

디스패치가 보고한 5건의 수리 착지점(acceptance.md F2 :93-99 + intro :65-70,
F1 :156-164, F3 :184-195, F4 흡수, F5 plan.md:82-83 + :97 주석)이 실제 착지했고,
progress.md §E.1 resolution map(:14-26)이 착지된 내용과 일치하며, ≥2 문턱이 이
트리에서 충족 가능하다.

## Evidence (본 세션 직접 측정 — 동일 트리, HEAD `6a46c0edb` 재확인 불변)

**F2 — 착지 확인 + 뮤턴트 산수 대조.** 현재 plan.md T1 초안에서
`brace expansion`의 행 분포: quoted 불릿 **2행**(:71 "and no brace expansion",
:72 "cannot be brace expansion"), unquoted 불릿 **1행**(:76 "and brace
expansion") — 합계 3. acceptance.md:93-98의 근거 서술("quoted 문장에 2회,
unquoted 문장에 1회, 합계 3")과 **정확히 일치**. 양방향 뮤턴트 재계산:
quoted-탈락 → 1 < 2 실패 ✓ / unquoted-탈락 → brace 앵커 2≥2 통과하나
`unquoted` 앵커 0으로 실패 ✓ — 두 방향 모두 기계층이 독립 선다. 문턱 충족
가능성: 삽입 3행 ≥ 2 (기존 독트린 0행) — 충족 가능. plan.md:97 주석
(">= 2 (quoted 문장 2회 + unquoted 문장 1회 = 3 기준...)")도 동일 산수로
두 문서 정렬 확인. 소개 문단(:65-70)도 양방향 서술로 정정됨.

**F1 — 착지 확인.** acceptance.md:156-164 RED 셀이 인자 순서(live 먼저,
mirror 나중)로 출력을 기재 + 핀 주석 병기. iter1 본 감사의 실측 출력 순서와
일치(내가 재측정해 확정한 순서).

**F3 — 착지 확인 + 귀속 관측 재현.** acceptance.md:184-191 셀이 stdout 공백 /
stderr 2행 / exit 1로 분리 기재. 본 세션 재현: `ls <두 초안경로> 2>/dev/null`
→ stdout 완전 공백 + rc=1 — 셀의 귀속 문장(:193-195)과 일치.

**F4 — 소멸 확인.** `grep -n "2.0" acceptance.md` → 유일 적중 `:9 phase:
"v3.2.0"`(버전 문자열) — "§2.0" 절 참조 0건. F2 재작성에서 흡수됨.

**F5 — 착지 확인.** plan.md:81-83 브리지 문단이 관측형으로 재기술됨
("command substitutions in the same position are already folded") —
"is the correct treatment" 처방 소멸. unquoted 불릿의 loosening-danger 경계
(:78-79 "correct behavior, not a defect")와 마무리 문단의 "record of
observations, not a specification" 불변 유지. 처방 논거는 M3 수정 방향 1에
보존(plan.md:169 실측).

**discriminating 앵커 재측정 (파이프 없는 rc, `6a46c0edb` 재핀)**:
`grep -c "brace expansion" <live>` → `0` / rc=1 · `grep -c "unquoted" <live>`
→ `0` / rc=1 — 문턱 ≥2 승격 후에도 RED-now 상태 불변(0 < 2), 트리 무변경
(HEAD `6a46c0edb` 재확인, untracked SPEC-artifact 편집만 존재).

**progress.md resolution map 대조**: :14-26 표의 5행이 실제 착지 내용과
전원 일치(F5 행의 인용문까지 plan.md:83과 verbatim 동일). :10의
"skip-eligible 아님" 선언은 skip policy 3조건 중 3(artifact-hash 불변)의
정확한 적용이다.

## Baseline-attribution

iter1과 동일 트리(HEAD `6a46c0edb`, 카드 커밋 0개)에서의 재측정. 변경된 것은
untracked SPEC artifact 3종(acceptance/plan/progress)뿐이며, RED-now 셀의
측정 대상(독트린 파일)은 건드리지 않았다 — 모든 RED 핀의 유효성 유지.

## 정정 기록 (감사자 측)

iter1 F2 서술의 "각 불릿이 정확히 1개씩 운반"은 부정확했다 — quoted 불릿은
2개를 운반했다(합계 3). iter1의 기계적 결론(≥1은 quoted-탈락 뮤턴트를 통과,
plan의 ≥2는 양방향을 잡음)은 그대로 성립하지만, per-불릿 계수의 정확한 값은
본 섹션이 정정하며 수리 측의 근거 서술(2/1=3)이 옳았다.

## Gaps

- spec.md와 research.md는 이번 수리에서 변경되지 않았다고 진술했고, 본 검증도
  변경 표면(acceptance/plan/progress)만 재독했다 — 두 파일의 전문 재감사는
  iter1 판정을 그대로 승계한다.
- 거절 프로브 재실행은 계속 원장 의존(세션 동작 — 커밋으로 재증명 불가).
- `TestRuleTemplateMirrorDrift` 실행은 run-phase T2 라인에 유지(이중 실행 회피).

## Residual-risk

- 표기 소소점: acceptance.md:194 "비영exit"은 "0이 아닌 exit"의 어색한 축약 —
  의미는 명확하고 판정에 영향 없음(cosmetic).
- research.md §2의 히트 경로 표기(live 경로 인용, 실측 위치는 미러 경로 —
  바이트 동일·동일 행번호로 실질 동치)는 iter1 관측 N3 그대로 잔존.
- 병합 창까지 develop 추가 진행 시 iter1의 "독트린 파일 충돌 0"이 만료 —
  흡수 후 재확인은 레인 절차에 이미 있음.

## 스코어 (iter2)

| 차원 | 가중 | 점수 (iter1 → iter2) |
|---|---|---|
| 트리 사실관계 정확도 | 0.30 | 1.00 → 1.00 |
| two-cell·수용 규율 | 0.25 | 0.80 → **1.00** (F2 착지) |
| 피벗 정직성·안전 경계 | 0.20 | 1.00 → 1.00 (F5 경계 불변 실측) |
| 추적성·범위·프론트매터 | 0.15 | 1.00 → 1.00 |
| 내부 정합성 | 0.10 | 0.70 → **0.85** (F1/F3/F4 착지; cosmetic 1건 + 관측 N3 잔존 공제) |

**합계: 0.30 + 0.25 + 0.20 + 0.15 + 0.085 = 0.985 → 절사 0.98**

**PASS (≥ 0.80)** — run-phase 진입 가능. 잔여 항목은 cosmetic 표기와 본질적
residual 뿐, 차단 또는 점수-유효 결함 없음. iter1의 최하중 약점(SSOT 임계)은
SSOT 쪽에서 봉합됐다.
