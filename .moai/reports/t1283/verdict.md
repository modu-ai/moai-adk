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

## resume-pointer

- **이월(리드 전언, t1259 run 착수 첫 일)**: Claude 2.1.283 은 `AGENTS.local.md` 를 primary 에서도 발견하지
  못하고 `CLAUDE.local.md` 는 워크트리에서 조상 탐색으로 로드되며, Codex 0.157.0 도 `AGENTS.local.md` 를
  발견하지 못한다(LOCAL_HEAD/TAIL 0·0). 근거: t1243 M1(agent-44, 커밋 `9f32f8077`, 증거
  `.claude/worktrees/t1243/.moai/reports/t1243/m1/evidence.md`). 따라서 `CLAUDE.local.md` 를 폐기하고
  `AGENTS.local.md` 만 두면 **워크트리 레인이 로컬 지침을 전혀 받지 못한다** — t1259 SPEC 의 이관 경로(두
  파일 동시 존재 금지, 한 릴리스 경고 대체)가 이 실측과 충돌하는지를 run 착수 첫 일로 판정한다.
