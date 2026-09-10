# (다) 문면 정정안 — 수집 주체와 반출 절차

상태: **초안. 커밋은 M1 재개 시.** 리드 판정(2026-09-02)에 따라 실질은 (가)로 해결됐고,
이 문서는 SPEC 문면이 실제와 어긋난 채 남는 것만 막는다.

측정 트리: `.claude/worktrees/t401` · `WT-analysis-pull` @ `6352897a5`

## 무엇이 어긋났나

SPEC 은 baseline·falsifier 창을 **오케스트레이터가 자기 호출로 모은다**고 전제한다. 칸반 레인
분업에서는 성립하지 않는다:

```
레인 경계 [HARD]: 레인은 운영자에게 묻지 않는다
  → AskUserQuestion 호출 0건
  → 관측 행 0건
  → 창을 채울 수 없다
```

행을 만들려고 질문을 지어내는 길은 **이 카드가 걷어내는 공허한 증거 그 자체**라 닫혀 있다
(리드가 (나)를 기각한 근거).

여기에 착지 위치 문제가 붙는다. 관측자는 `CLAUDE_PROJECT_DIR`(미설정 시 cwd)로 로그 경로를
정한다(`internal/hook/askuser_observer.go:156`). 리드 세션은 primary 에서 돌므로 리드의 호출은
**primary 의 `.moai/logs/askuser-observations.jsonl`** 에 쌓이고 카드 워크트리에서는 보이지 않는다.

## 고칠 좌표 4곳

### 1. `plan.md:99-106` — M0 수집 주체 명시

현행은 "collect a `push`-mode control sample …" 로 시작해 **누가** 모으는지 말하지 않는다.
`Copy the window's rows to …` 라는 복사 단계는 있으나 원본이 어디인지 없다.

추가할 문장(취지):

> 창을 모으는 주체는 **실제로 묻는 세션**이면서 동시에 **관측자가 실제로 배선된 세션**이다.
> 칸반 분업에서 묻는 쪽은 카드의 레인이 아니라 운영자 채널을 소유한 리드 세션이다(레인은
> `feedback_lane_cannot_open_operator_gate` [HARD] 에 따라 운영자에게 묻지 않으므로 호출을
> 만들지 못한다). 그러나 묻는 것은 **필요조건일 뿐 충분조건이 아니다.** 행은 그 세션 안에서
> 관측자가 살아 있을 때만 남는다 — `AskUserQuestion` matcher 가 설정에 추가되기 **전부터**
> 돌고 있던 세션은 그 matcher 를 집어 들지 않으므로, 묻기는 묻고 아무것도 기록하지 않는다.
> 따라서 수집 세션은 matcher 착지 **이후에 시작된** 세션이다. 행은 묻는 세션의
> `CLAUDE_PROJECT_DIR`(미설정 시 cwd) 하위 `.moai/logs/askuser-observations.jsonl` 에 쌓이며,
> 이는 일반적으로 카드 워크트리가 **아니다**. 실제로 물은 세션에서 로그가 비어 있으면 그것은
> **관측자 미배선의 증거**이지, 라벨 붙은 선택지가 제시된 적 없다는 증거가 결코 아니다 —
> 진짜 부재와 공허한 초록을 가르는 지점이다. 창을 목적으로 질문을 지어내는 것은 금지한다 —
> 그렇게 채운 창은 표본이 아니라 장식이다.

두 번째 조건을 붙이는 근거는 리드 세션의 **부정 관측**이다(2026-09-02, primary 체크아웃에서
리드가 측정한 것으로 보고받음 — 이 문서 작성자의 자체 측정이 아니다). 리드가 실제로 4문항
`AskUserQuestion` 호출을 한 뒤 primary 체크아웃과 develop 양쪽에서 `askuser-observations.jsonl`
을 찾았으나 어느 쪽에도 파일이 없었다. PreToolUse matcher 의 wire shape 자체는 이 카드에서
따로 초록으로 확인됐으므로, 남는 설명은 하나다 — matcher 추가 시점에 이미 돌고 있던 세션은
그것을 집어 들지 않는다. 즉 관측자는 **설정에는** 배선돼 있으나 **그 살아 있는 세션에는**
배선돼 있지 않았다.

### 2. `plan.md:313-316` — M6 verify 경로 + 플레이스홀더

현행 verify 가 `.moai/logs/<observer>.jsonl` 을 읽는다. 두 가지가 틀렸다:

- `<observer>` 는 플레이스홀더다. 실제 파일명은 `askuser-observations.jsonl`
  (`internal/hook/askuser_observer.go:26` 상수, 실측 확인)
- 이 트리의 `.moai/logs/` 를 읽는데, 묻는 세션이 다른 트리에 있으면 **영원히 빈 파일**이다.
  빈 파일에 `select(.mode=="pull")` 를 걸면 `violations: 0` 이 나오고 **위반 없음처럼 읽힌다** —
  공허한 초록

고칠 방향: **반출된 산출물**을 읽게 한다.

```bash
jq -s '[.[] | select(.mode=="pull")] | {n: length, violations: ([.[] | select(.label_present==true)] | length)}' \
  .moai/reports/t401/pull-window.jsonl
```

그리고 `n` 하한(20) 미달을 gap 으로 처리하는 현행 문장은 유지한다 — 그것이 빈 파일의 공허한
초록을 막는 장치다.

### 3. `acceptance.md` AC-JFM-023 surface 2 / AC-JFM-018 — 반출 출처 검증 추가

지금은 반출된 JSONL 의 행 수와 양성 수만 본다. 반출 자체가 **귀속 없는 주장**이 될 수 있다 —
어느 세션이 어느 구간에 모은 것인지 파일만으로는 알 수 없다.

추가할 요구: 반출 산출물 옆에 `<name>.provenance.md` 를 두고 아래를 적는다.

| 항목 | 이유 |
|---|---|
| 원본 절대경로 | 어느 트리의 로그였는지 |
| 묻는 세션의 session_id | 누가 모았는지 |
| 수집 구간(첫 행·마지막 행 timestamp) | 언제 |
| 행 수 · `label_present:true` 수 | 반출 시점 실측 |
| 반출 명령 | 재현 경로 |

그리고 **AC 는 provenance 파일의 존재와 행 수 일치를 함께 단언**한다. 파일만 있고 출처가 없으면
`verification-claim-integrity.md §2` 의 미귀속 주장이다.

### 4. `spec.md` — REQ 문면

수집 주체를 못박는 REQ 가 없다. 신규 REQ 1건(취지):

> Where the recorded-window evidence is collected by a session other than the one implementing the
> SPEC, the exported artifact MUST carry its provenance record, and the acceptance criteria MUST
> assert both the artifact and its provenance.

번호는 M1 재개 시 다음 빈 번호로 부여한다(현재 REQ 24건이므로 REQ-JFM-025 예상 — 부여 직전
`grep -c '^\*\*REQ-JFM-'` 로 재확인할 것. 오늘 리드가 토큰 출현 횟수를 항목 수로 오독한 전례가
있으므로 선택자를 고정한다).

## 이 정정이 만드는 부수 효과

M0 의 RED-now 셀 중 `ls .moai/reports/t401/baseline-push-window.jsonl` → 부재/exit 1 은 그대로
유효하다. provenance 파일이 추가되면 RED-now 셀도 두 파일 부재로 늘어난다 — M1 재개 시
**이 트리에서 재측정**할 것. 기억한 값은 baseline 이 아니다.

## 하지 않을 것

- M0 을 다시 열지 않는다. 리드 판정대로 실질은 (가)로 해결됐고 문면만 고친다.
- 레인 예외((나))를 쓰지 않는다. 지어낸 질문으로 채운 창은 표본이 아니다.
- `<observer>` 플레이스홀더를 남겨두지 않는다 — 상수 실측값으로 고정한다.
