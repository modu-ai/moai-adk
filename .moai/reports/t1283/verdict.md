# 판정서 — 카드 t1283 (detail 동반 파일 parity 레지스트리)

트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1283` · 브랜치 `WT-detail-parity-registry` ·
기준 HEAD `e765d33ec`(t1064 병합 포함, develop 대비 `0 0`) · 클래스 B(run, plan 생략).

발행 경위: t1064 sync-audit 이 「`main-checkout-branch-guard-detail.md` 가 어느 parity 레지스트리에도
없다」를 Gap 으로 냈고, 레인이 그것을 직접 확인해 리드에 카드 요청 → 리드 발행.

---

## ① 측정 — 구멍이 카드가 예상한 것보다 넓다

두 가드의 선택 논리를 먼저 읽은 것이 이 측정의 성립 조건이었다. **바이트 가드는 전수 검사가 아니라 명시
allowlist** 다 — `rule_template_mirror_test.go` 자기 주석: *"Allowlist is intentionally explicit (no
glob) so that adding a new mirrored file is a deliberate code change visible in PR review."* 따라서
「테스트가 안 죽는다」가 「보호받는다」를 뜻하지 않는다. 이 구별을 세우지 않으면 5개 파일을 안전한 것으로
오독한다.

`.claude/rules/moai/**` 의 `*-detail.md` 동반 파일 **9개 전수**:

| 파일 | 미러 | 두 사본 | parity 등록 | stub 등록 |
|---|---|---|---|---|
| `core/moai-constitution-detail.md` | 있음 | **다름** | 없음 | 없음 |
| `core/verification-claim-integrity-detail.md` | 있음 | **다름** | 없음 | **있음** |
| `workflow/cross-session-messaging-detail.md` | 있음 | **다름** | 없음 | 없음 |
| `workflow/main-checkout-branch-guard-detail.md` | 있음 | **다름** | 없음 | **있음** |
| `core/native-idiom-and-register-detail.md` | 있음 | 같음 | 없음 | 없음 |
| `workflow/context-window-management-detail.md` | 있음 | 같음 | 없음 | 없음 |
| `workflow/goal-directive-detail.md` | 있음 | 같음 | 없음 | 없음 |
| `workflow/kanban-dispatch-detail.md` | 있음 | 같음 | 없음 | 없음 |
| `workflow/skill-routing-detail.md` | 있음 | 같음 | 없음 | 없음 |

**9개 전부 어느 가드에도 없다.** 카드가 지목한 「stub 은 등록, detail 은 미등록」 형태는 그중 2쌍이고,
나머지 7쌍은 양쪽 다 미등록이다.

**부수 발견 — `kanban-dispatch-detail.md` 는 「깨끗해서 같음」이 아니다.** 같음으로 분류된 5개 중 이 파일만
템플릿 미러가 내부 토큰을 **4행** 싣고 있다(나머지 4개는 0행). 즉 양쪽이 똑같이 오염된 상태이며, 바이트
가드에 넣으면 그 오염이 고정되고, 정리하면 바이트 동일이 깨져 sanitized_pair 로 가야 한다. 이 카드 범위
밖으로 남긴다.

## ② 등록 — 운영자 판정으로 「갈리는 4쌍만」

운영자 게이트(2026-09-27)에서 네 안 중 **갈리는 4쌍만 sanitized_pair 에 등록**을 택했다. 같은 5쌍은
지금 우연히 일치하는 상태로 남으며 별 카드 후보로 보고한다.

등록 후 첫 실행에서 **3/4 통과, 1건 FAIL**:

```
PASS  moai-constitution-detail.md            net one-sided=1 (tolerance 4)
PASS  verification-claim-integrity-detail.md
PASS  cross-session-messaging-detail.md
FAIL  main-checkout-branch-guard-detail.md   net one-sided=5 (tolerance 4)
      SANITIZED_PAIR_PARITY_DRIFT: 로컬에만 있고 미러에 없는 독트린 5행
```

**가드가 붙자마자 실재 드리프트를 잡았다.** 그 파일은 t1064 가 고친 파일이고, t1064 에서 감사와 레인이
detail 미러의 등가성을 **육안 diff 로만** 판정해 둘 다 「등가」라고 했던 대상이다. t1064 판정서가 Gap 으로
「detail 등가성 근거는 육안 diff 뿐」이라 적어 둔 지점이 정확히 여기다.

**원인 분리 — 주범은 t1064 가 아니다.** 실제 diff 를 읽어 두 갈래로 갈랐다:

- **(가) 선행 정화 누락(로컬 전용 5행 중 대부분)**: 미러에서 통째로 빠진 `Origin:` 출처 블록 **4행**과
  `path (REQ-6 backward compat).` **1행**. t1064 이전부터 있던 상태다.
- **(나) t1064 가 만든 잡음(내용 동일, 줄 경계만 다름)**: 미러 중립화 때 identity-axis 문단을 다시 감싸면서
  로컬과 줄 경계가 달라졌다. 정규화가 `<SPEC-ID>`·`<DATE>` 로 치환하므로 토큰 자체는 문제가 아니고,
  **줄 단위 비교에서 한쪽 전용으로 잡히는 것**이 문제였다.

**처방은 관행을 따랐다 — 새로 만들지 않았다.** 통과하는 등록 파일이 이미 답을 갖고 있다:
`verification-claim-integrity.md` 는 출처를 로컬 **1행**(`> Provenance: …`)으로 두고 미러에서는 빼며,
1행이라 허용 오차에 흡수된다. 그래서 (가)는 4행 블록을 같은 형태의 **1행**으로 모았고(세 SPEC ID 와 두
REQ 범위 전부 보존), (나)는 미러를 로컬과 같은 줄 경계로 다시 감쌌다.

결과: **net one-sided 5 → 1**(허용 4), 13개 서브테스트 전부 PASS.

## ③ 변이 검출 — 등록이 공허하지 않음

변이는 **새로 등록한 파일**에 걸었다. 이미 등록돼 있던 항목에서만 죽는다면 새 항목 4개는 미증명으로 남기
때문이다. `cross-session-messaging-detail.md` 미러에서 본문 연속 6행을 제거:

```
$ <mirror 에서 본문 6행 제거>
--- FAIL: TestSanitizedPairParity/cross-session-messaging-detail.md
    normalized diff — 10 local-only, 4 template-only, ~4 reword pairs, net one-sided=6 (tolerance 4)
    SANITIZED_PAIR_PARITY_DRIFT: … carries 6 content line(s) of doctrine present in the LOCAL copy
    but ABSENT from the template mirror …
$ <복원>
$ go test ./internal/template/ -run 'SanitizedPair|Leak|Neutral|MirrorDrift' -count=1
ok      github.com/modu-ai/moai-adk/internal/template    1.816s
```

제거한 행 수 6과 보고된 `net one-sided=6` 이 일치한다 — 가드가 **세는 대상을 실제로 세고 있다**.
복원 후 네 가드 전부 초록.

## ④ 동반 과제 — t1064 판정서 정정

t1064 의 재감사 기록이 `PASS-WITH-DEBT` 로 남아 있었다(감사 최종분이 병합 `e765d33ec` 뒤에 도착).
최종 **PASS(부채 없음)** 와 점수(Functionality 93 · Security 88 · Craft 86 · Consistency 94,
must-pass 는 Functionality + Security 두 축), 그리고 감사자가 둘째 부모 측정을 독립 재현한 뒤
`[blocking-for-merge]` 등급·해당 문장·D2 를 철회한 경위를 기록했다. 관측 순서는 지우지 않고 최종 상태를
덧붙이는 형태로 정정했다.

---

## Baseline-attribution

- 트리 `.claude/worktrees/t1283` · 브랜치 `WT-detail-parity-registry` · 기준 `e765d33ec`
- 위 ①~④ 의 모든 수치와 출력은 **이번 실행에서 이 트리에 대해** 낸 명령의 것이다. 인용해 옮긴 것은 없다.
- 변경 파일 4개: `sanitized_pair_parity_test.go`(레지스트리 4항목 추가) ·
  `main-checkout-branch-guard-detail.md` 로컬(출처 1행화) 및 미러(줄 경계 정렬) ·
  `.moai/reports/t1064/verdict.md`(④).

## Gaps

- **같은 5쌍은 여전히 무가드다.** 운영자 판정으로 범위 밖이며, 지금 일치하는 것은 규율의 결과일 뿐 가드의
  보장이 아니다. 별 카드 후보.
- **`kanban-dispatch-detail.md` 의 미러 내부 토큰 4행은 손대지 않았다** — 범위 밖 부수 발견.
- **(가)의 `Origin:` 블록을 미러로 전파하지 않았다.** 관행대로 미러에서 빼는 쪽을 택했고, 로컬을 1행으로
  모아 오차에 들어가게 했다. 「출처를 미러에도 중립 문구로 싣는다」는 대안은 채택하지 않았다.
- **전 패키지 스위트 미실행** — 변경 영향 범위(`internal/template`)만 쟀다.
- t1064 의 열린 Gap(살아 있는 `manager-git` 서브에이전트 실 Bash 발사 · `MOAI_BRANCH_GUARD_EXEMPT`
  환경변수 축)은 이 카드가 닫지 않는다.

## Residual-risk

- 허용 오차 4행은 이 카드가 정한 값이 아니다. `main-checkout-branch-guard-detail.md` 가 1행으로 통과하므로
  **여유가 3행뿐**이다 — 그 파일에 로컬 전용 4행이 더 붙으면 다시 죽는다. 그것이 가드의 의도지만, 다음
  사람이 「왜 갑자기 죽나」로 읽지 않도록 이 여유 폭을 여기에 적어 둔다.
- 줄 경계 정렬은 내용이 아니라 형식에 의존한다. 미러를 다시 감싸는 편집이 들어오면 같은 형태로 재발할 수
  있다.


---

## 감사 지적 반영 — 정정 4건 (2026-09-27, sync-audit 후)

감사자 모델 귀속: `auditor-model: claude-opus-5[1m]`. 감사자 스스로 **런타임 자기 선언이며 독립 측정이
아니라는 한계**를 고지했다. GLM 서빙 징후는 없다고 보고했으나 그 판단 역시 자기 선언 층에 있다.
독립 관측은 병합 전 리드에게 요청한다(t1064 에서는 서브에이전트 트랜스크립트의 `.message.model` 과 스폰
`.meta.json` 두 경로가 그 역할을 했다).

### 정정 1 — 「관행을 따랐다」는 근거를 넘은 주장이었다 (철회)

위 ② 절과 커밋 `b6c1c3f4a` 의 메시지는 출처 1행화를 *"follows the house convention instead of
inventing one"* 이라고 적었다. **성립하지 않는다.** 감사 측정:

- 카드 이전 규칙 트리에서 `^> Provenance:` 는 **1개 파일**(`verification-claim-integrity.md`)뿐.
- 내가 버린 `^Origin:` 형식은 **4개 파일**에 살아 있음(`agent-common-protocol-reference.md`,
  `orchestration-mode-selection.md`, `archived-agent-rejection.md`, `lifecycle-sync-gate.md`).

즉 **더 흔한 형식을 더 드문 쪽으로 바꾸고** 그것을 관행이라 불렀다. n=1 에서의 일반화다. 주장을 철회한다.

**「1행화가 유일한 통과 경로였다」는 문장도 쓰지 않는다** — 아래 정정 2가 측정으로 반증한다.

### 정정 2 — 제3의 경로를 채택했다 (형식 되돌림, 여유폭 확대)

내가 상정한 갈래는 둘(1행화 유지 / 4행 유지하며 다른 방식으로 오차 진입)이었고, 감사가 **후자가 실제로
작동함을 변이로 측정**했다. 그 처방을 채택했다:

- **로컬**: 4행 `Origin:` 블록을 원상 복구(`e765d33ec` 판과 같은 형태). 세 SPEC ID·세 REQ 범위 그대로.
- **미러**: 같은 자리에 같은 행 수의 **중립 산문 대응 블록**을 넣었다(내부 식별자 없음). 4행이
  `local-only` 로 남지 않고 **reword 쌍으로 상계**된다(`localExcess = len(localOnly) - len(tmplOnly)`).

측정 결과:

```
$ go test ./internal/template/ -run TestSanitizedPairParity -count=1 -v
main-checkout-branch-guard-detail.md: normalized diff —
  10 local-only, 10 template-only, ~10 reword pairs, net one-sided=0 (tolerance 4)
  → within reword tolerance
$ go test ./internal/template/ -run 'SanitizedPair|Leak|Neutral|MirrorDrift' -count=1
ok      github.com/modu-ai/moai-adk/internal/template    1.149s
$ grep -cE 'SPEC-…-[0-9]{3}|REQ-[A-Z0-9]|card t[0-9]+' <미러>   → 0
$ grep -coE 'SPEC-WORKTREE-BRANCH-GUARD(-OPTIN|-DISCRIM)?-001' <로컬>   → 4
```

**`net one-sided=0`** — 착지분(`net=1`)보다 여유폭이 한 행 넓다. 동시에 다수 형식을 유지하고, 미러가
provenance 를 **침묵으로 잃지 않고** 중립 문구로 보유한다. 잔여 위험의 「여유폭 3행」은 **4행**으로 갱신된다.

### 정정 3 — 부수 발견의 수치가 틀렸다 (4행 → 카드 id 2행)

`kanban-dispatch-detail.md` 미러의 내부 토큰을 **4행**이라 적었다. 재측정 결과 **그 4행은 전부 거짓
양성**이다 — 내 패턴 `SPEC-[A-Z0-9-]+` 가 `<SPEC-ID>` **플레이스홀더** 안쪽을 물었다(`/moai run
<SPEC-ID>` 같은 일반 산문). 앵커를 붙이면 내부 SPEC ID 는 **0**이다.

감사가 잡은 것은 다른 클래스이며 그쪽이 실재한다 — **내부 카드 id 2행**(`26:t133`, `186:card t224`).
로컬 사본도 같은 2행이므로 **양쪽이 동일하게 오염**돼 있다. 따라서 결론(이 파일만 오염 · 정리하면 바이트
동일이 깨지므로 바이트 등록 불가)은 살아남고 **근거가 교체**된다: SPEC ID 4행이 아니라 카드 id 2행.

```
$ grep -cE 'card t[0-9]+|\bt[0-9]{2,4}\b' <로컬>  → 2
$ grep -cE 'card t[0-9]+|\bt[0-9]{2,4}\b' <미러>  → 2
$ grep -coE 'SPEC-[A-Z0-9]+(-[A-Z0-9]+)+-[0-9]{3}' <미러>  → 0
```

**이 오류의 성격**: t1064 판정서 §E.1 (10) 에 내가 직접 「앵커 없이 센 값은 과다계상이며 재측정해도 같은
값이 나와 검산으로 안 잡힌다」고 적었다. 같은 함정에 같은 사람이 걸렸다. 교훈을 기록한 것이 면제가 되지
않는다는 실사례로 남긴다.

### 정정 4 — 가드는 양방향이다 (내 의심이 기각됨)

감사 의뢰서에서 「가드가 `local-only` 축만 세는 일방향일 수 있다」를 의심 축으로 넘겼다. **측정으로
기각됐다.** 감사가 신규 등록 항목으로 양방향 변이를 실행했다 — 미러에서 6행 삭제 → `net one-sided=6`
FAIL(`present in the LOCAL copy but ABSENT from the template mirror`), 로컬에서 8행 삭제 →
`net one-sided=-6` FAIL(`present in the TEMPLATE mirror but ABSENT from the local copy`). 코드 근거는
`sanitized_pair_parity_test.go:219` 의 `case -localExcess > structuralDriftToleranceLines`.

**내 부족은 도달 범위가 아니라 그 범위를 재지도 보고하지도 않은 것이다.** ③ 절의 변이는 한 방향만
확인했고, 판정서는 그 한계를 적지 않았다. 이제 적는다.

### 감사가 추가로 낸 발견 2건 — 이 카드가 고치지 않는다

- **`sanitized_pair_parity_test.go` 의 `runtime-recovery-doctrine.md` 주석이 스테일하다.** 주석은 로컬이
  "Origin line" 을 보유한다고 적지만 **양쪽 어디에도 `Origin:` 줄이 없다**(각 0행). 리드에 별 카드 후보로
  보고한다.
- **커밋 `b6c1c3f4a` 메시지의 "both REQ ranges" 는 부정확하다** — 실제 범위는 **세 개**
  (`REQ-WBG-001..013`, `REQ-1..6`, `REQ-WBG-D-001..008`). 세 개 다 보존돼 손실은 없고 서술 수만 틀렸다.
  커밋 메시지는 고치지 않고(이력 무결성) 이 절에 정정으로 남긴다.

### 감사가 확인해 준 것 (전부 감사 자신의 실행)

- **렌즈 1 PASS**: `find .claude -name '*-detail.md'` → 정확히 9개, `.claude/agents/`·`.claude/skills/`·
  중첩 디렉터리에 누락 **0**. 내 루프가 하위 디렉터리 5개를 손으로 열거한 것이 결과적으로 트리를 전부
  덮었음을 독립 확인했다. 4열 행렬도 내 표와 전부 일치. `rule_template_mirror_test.go:41-43` 이 명시
  allowlist 임을 읽고, 9개 중 어느 것도 그 목록에 없음을 확인해 **「통과」가 「미측정」이었다는 판단이
  옳다**고 판정했다.
- **렌즈 2**: 등록 4건 전부 통과. Origin 붕괴에서 SPEC ID 3개·REQ 범위 3개 **손실 0**. 미러 재줄바꿈은
  공백 정규화 후 `cmp` **바이트 동일** — 단어 손실 없이 줄 경계만 움직였음을 증명.
- **④ 충족**: `git diff e765d33ec HEAD -- .moai/reports/t1064/verdict.md` 로 최종 PASS 기록·점수 4개·
  관측 순서 보존(덮어쓰지 않고 이어 붙임)·철회 기록 3건 전부 실재 확인.
- **`git show --stat b6c1c3f4a`**: 5 files changed, 162 insertions(+), 9 deletions(-) — 커밋 메시지가
  주장하는 항목 전부가 내용에 있음(위 "both REQ ranges" 오기 1건 제외).
- **재실행**: `golangci-lint` v2.1.6 `0 issues.` · `gofmt -l internal/template/` 0행 ·
  `go test ./internal/template/ -count=1` exit 0 (패키지 전량 91.162s) · parity 13/13.
- **복원 증명**: 감사의 모든 변이 실험이 sha256 일치 + 빈 `git status --porcelain` 으로 원복 확인됨.

### 갱신된 잔여 위험

- **여유폭은 이제 4행**(`net one-sided=0`, tolerance 4). 정정 2로 3행에서 넓어졌다. 감사가 3행 상태를
  「유예된 실패에 가깝다」고 평했고, 그 평가를 받아들여 여유폭을 넓히는 쪽을 택한 것이다. 다만 오차 0을
  요구하면 정상 §25 사문화가 적색이 되므로 tolerance 자체는 건드리지 않는다.
- **범위 밖 5쌍은 현재 동일성을 무엇도 강제하지 않는다** — 어느 가드에도 없으므로 **한쪽만 편집되면
  조용히 갈라진다**. 별 카드 후보로 남긴 판단은 유지하되, 함정이 실재한다는 감사 표현을 그대로 싣는다.

---

## resume-pointer

- **이월(리드 전언, t1259 run 착수 첫 일)**: Claude 2.1.283 은 `AGENTS.local.md` 를 primary 에서도 발견하지
  못하고 `CLAUDE.local.md` 는 워크트리에서 조상 탐색으로 로드되며, Codex 0.157.0 도 `AGENTS.local.md` 를
  발견하지 못한다(LOCAL_HEAD/TAIL 0·0). 근거: t1243 M1(agent-44, 커밋 `9f32f8077`, 증거
  `.claude/worktrees/t1243/.moai/reports/t1243/m1/evidence.md`). 따라서 `CLAUDE.local.md` 를 폐기하고
  `AGENTS.local.md` 만 두면 **워크트리 레인이 로컬 지침을 전혀 받지 못한다** — t1259 SPEC 의 이관 경로(두
  파일 동시 존재 금지, 한 릴리스 경고 대체)가 이 실측과 충돌하는지를 run 착수 첫 일로 판정한다.
