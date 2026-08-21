# Design — SPEC-FEEDBACK-AUTO-SUBMIT-001

> Tier L 설계 문서. 무엇을 만들 것인지(경계·계약·순서·실패 모드)를 정한다. 구현 함수명 수준의 세부는 plan.md 마일스톤이 갖는다.

## §1 경계 — 어디까지가 코드이고 어디부터가 산문인가

```
사용자 ──▶ 오케스트레이터 (스킬 본문: workflows/feedback.md)
                │  ① 필드 수집 (기존 AskUserQuestion :52)
                │  ② 본문 조립 (기존)
                ▼
          moai feedback scrub          ◀── 여기부터 Go (테스트 가능)
                │  stdin: 본문
                │  stdout: {verdict, body, findings, reason}
                │  side effect: .moai/logs/feedback-mask.log
                ▼
          오케스트레이터                ◀── 다시 산문
                │  ③ verdict != ok → 제출 중단 + SECURITY.md 경로 안내
                │  ④ auto_submit == false → 확인 게이트 (신규 AskUserQuestion)
                │  ⑤ gh issue create
                │  ⑥ 실패 시 moai feedback queue enqueue
                ▼
             GitHub issue
```

**설계상 가장 중요한 한 줄**: 코드가 소유하는 것은 ②→③ 사이의 **변환**뿐이다. ③④⑤⑥의 판단과 실행은 산문이 소유한다. 이 경계가 곧 `spec.md` §E.3 잔여 위험의 정확한 형태이며, 설계를 읽는 사람이 "스크러버를 넣었으니 강제된다"고 오해하지 않도록 다이어그램에 명시한다.

## §2 패키지 경계

| 패키지 | 책임 | 의존 |
|---|---|---|
| `internal/feedback` (신규) | 마스킹 변환, 취약점 분류, 판정 타입, 마스킹 로그, 재시도 큐 | `internal/hook`(패턴 정책), `internal/paths`, `internal/sandbox`(이름 어휘), `internal/atomicfile` |
| `internal/cli/feedback.go` (신규 서브커맨드) | stdin/stdout 배선, JSON 인코딩, 종료 코드 | `internal/feedback` |
| `internal/config` | `feedback.auto_submit` 스키마·기본값·접근자 | — |

**왜 `internal/hook`이 아닌가**: 훅 패키지가 CLI를 위해 존재하는 코드를 떠안게 되고, 이미 큰 훅 테스트 스위트에 무관한 테스트가 얹힌다. 반대 방향 의존(`feedback` → `hook`)만 두면 순환이 없다 — `hook`은 `feedback`을 모른다.

**왜 `internal/cli`에 직접 두지 않는가**: 변환 테스트가 cobra 명령 조립에 결합된다. 순수 함수로 두면 표 기반 테스트가 가능하다.

## §3 마스킹 파이프라인 — 순서가 결과를 바꾼다

```
원문 body
  │
  ├──▶ [분류] classify(원문)  ─────────────┐   ★ 원문을 본다
  │                                        │
  ▼                                        │
① env 값 정확일치 마스킹                    │
  ▼                                        │
② 시크릿 정규식 마스킹                      │
  ▼                                        │
③ 홈 경로 접두사 축약                       │
  ▼                                        ▼
마스킹 body ──────────────────▶ Result{verdict, body, findings, reason}
```

**[HARD] 분류는 마스킹 이전 원문을 본다.** 마스킹이 지우는 것이 정확히 분류기가 필요로 하는 신호(시크릿 패턴 적중, 키 파일 경로 언급)이기 때문이다. 마스킹 후 본문을 분류하면 "시크릿이 들어 있어서 위험한 보고"가 "시크릿이 없는 평범한 보고"로 보인다. 이 순서를 뒤집는 것은 조용한 미탐이며, 테스트로 잡기 어렵다 — 그래서 설계에 못박는다.

**변환 3단계의 순서 근거**:

1. **env 값 먼저**: 정확 일치이므로 가장 정밀하다. 값이 토큰 형태이기도 하다면 `findings`에 `env`로 귀속되는 편이 진단에 유용하다. 정규식이 먼저 잡으면 `secret`으로 잘못 귀속된다.
2. **시크릿 정규식 다음**: env로 걸러지지 않은 나머지를 훑는다.
3. **홈 경로 마지막**: 앞 두 단계가 만든 마스킹 토큰이 경로 매칭을 방해하지 않는 위치. 반대로 홈 축약을 먼저 하면 `/Users/x/.ssh/id_rsa` 같은 경로가 `~/.ssh/id_rsa`로 바뀌어 분류 신호로서의 절대성이 줄어든다(분류는 이미 원문을 봤으므로 실害는 없지만, 순서를 단순화할 이유도 없다).

**멱등성**: 파이프라인은 멱등이어야 한다 — 마스킹된 본문을 다시 스크럽해도 결과가 같아야 한다. 재시도 큐에서 꺼낸 본문(이미 마스킹됨)이 재전송 전에 다시 스크럽될 수 있기 때문이다.

## §4 타입 계약

```go
// internal/feedback
type Finding struct {
    Kind  string // "secret" | "env" | "homepath"
    Count int
}

type Result struct {
    Verdict  string    // "ok" | "blocked"
    Body     string    // 마스킹된 본문
    Findings []Finding // 종류·건수만. 원문 값 금지.
    Reason   string    // blocked 일 때만 비어 있지 않음
}
```

JSON 필드명은 소문자 스네이크(`verdict` / `body` / `findings` / `reason`). 스킬 본문이 `jq`로 읽을 수 있어야 하므로 **최상위는 배열이 아니라 객체 하나**이고, 진단 로그를 stdout에 섞지 않는다(사람용 메시지는 stderr).

**종료 코드 축의 분리**:

| 상황 | 종료 코드 | stdout | 호출자 행동 |
|---|---|---|---|
| 스크럽 성공, 통과 | 0 | `verdict: "ok"` | 게이트 → 제출 |
| 스크럽 성공, 차단 | 0 | `verdict: "blocked"` + reason | 제출 금지, 수동 경로 안내 |
| 도구 실패 | ≠ 0 | (JSON 없음) | 제출 금지 (fail-closed) |

`blocked`를 종료 코드로 인코딩하지 않는 이유: 도구 실패와 정책 차단이 같은 채널에서 섞이면 스킬 본문이 숫자 규약을 외워야 한다. 분리하면 [HARD] 조항이 두 문장으로 끝난다 — "0이 아니면 제출 금지, `verdict != ok`면 제출 금지".

## §5 온디스크 산출물 2종 — 형태가 다른 이유

| | 마스킹 로그 | 재시도 큐 |
|---|---|---|
| 경로 | `.moai/logs/feedback-mask.log` | `.moai/state/feedback/queue.json` |
| 형태 | 라인 지향 append | 단일 JSON + 형제 lock |
| 권한 | `0o600` | `0o600` |
| 선례 | `internal/config/log.go`(형태) + `failure_observer.go:156`(권한) | `internal/kanban/backlog_store.go` |
| 삭제 | 없음(추가만) | 성공 시 항목 삭제 필수 |
| 실패 정책 | fail-open (`slog.Warn` 강등) | fail-open (전송 자체는 이미 실패했으므로 큐 실패가 더 나쁜 상태를 만들지 않는다) |
| 동시성 | 인터리빙 무해 | `Mutate()` 잠금 필요 |

**하나로 통일하지 않는 이유**: 큐를 append-only로 만들면 "성공했는데 큐에 남아 있는" 상태를 표현할 수 없고, 로그를 잠금 있는 JSON으로 만들면 로깅이 스크럽을 블록할 수 있다(fail-open 위반). 요구가 다르므로 형태도 다르다.

**`.moai/logs/` 선택의 부수 이득**: 세 개의 청소 스윕(`internal/cli/update_cleanup.go:133`, `internal/worktree/state_guard.go:54`, `internal/cli/codex_review_gate.go:39`)이 이미 이 디렉터리를 제외한다.

**전송-1회 의미론이 필요해지면**: handoff의 claim-rename(`internal/hook/handoff_inject.go:116,252` + `consumed/` 아카이브 + `DefaultHandoffStaleTTL`)이 정확한 선례다. 이 SPEC은 재시도를 원하므로 채택하지 않는다.

## §6 취약점 분류기 설계

선례가 없는 유일한 부분이므로 설계를 명시한다.

**입력 신호 3종**(전부 원문 대상):

1. 시크릿 패턴 적중(REQ-4의 집합) — 본문이 실제 자격증명을 담고 있다.
2. 시크릿 보유 파일 부류 언급 — `denyPatterns`(`internal/hook/pre_tool.go:216-230`)의 `\.pem$`, `\.key$`, `\.ssh/.*`, `id_rsa.*`.
3. 취약점 어휘 — CVE/CWE 식별자 형태, 그리고 공개 채널 게시가 위험한 보고를 시사하는 표현군. 어휘는 신규 작성이며 `internal/feedback/classify.go`에 상수로 둔다(하드코딩 금지 규율 §14 — 리터럴을 함수 본문에 흩뿌리지 않는다).

**판정**: 신호 1은 단독으로 `blocked`. 신호 2·3은 조합 규칙으로 판정하되, **초기 임계값은 오탐 쪽으로 보수적으로** 잡는다 — 미탐 1건은 공개 채널 유출이고 오탐 1건은 사용자가 수동 경로로 가는 불편이다. 비용이 비대칭이다.

**거부 메시지**: `SECURITY.md`의 두 문장을 인용하고 Advisories URL을 그대로 싣는다. 정책을 재서술하지 않는다 — 재서술은 원문과 갈라진다.

**오탐 대조 테스트가 필수인 이유**: "전부 차단"하는 분류기는 미탐 테스트를 100% 통과한다. AC의 오탐 대조(평범한 버그 리포트 → `ok`)가 그 축퇴 구현을 배제하는 유일한 장치다.

## §7 확인 게이트의 표시 형태

`AskUserQuestion` 한 라운드, 옵션 3개:

| 옵션 | 라벨 | 결과 |
|---|---|---|
| 1 | 제출하지 않음 (권장) | `gh issue create` 미실행. 로컬 초안 경로 안내(`feedback.md:40`의 기존 초안 관례 재사용) |
| 2 | 이대로 제출 | 마스킹 본문 그대로 제출 |
| 3 | 본문 수정 후 제출 | "Other"로 수정본을 받고, **수정본을 다시 스크럽**한 뒤 재확인 |

`(권장)`은 첫 옵션에만 붙는다(askuser-protocol 규약). 기본값을 "제출하지 않음"에 두는 이유: 공개 채널 게시는 되돌리기 어렵고, `auto_submit: false`를 택한 사용자는 이미 신중 쪽을 고른 사람이다.

**옵션 3의 재스크럽이 §3 멱등성 요구의 출처다.**

질문 본문에는 마스킹된 본문 **전문**과 `findings` 요약(종류별 건수)을 함께 싣는다 — 무엇이 가려졌는지 모르면 사용자는 판단할 수 없다.

## §8 설정 판독 경로

`feedback.auto_submit`은 **Go 코드가 읽지 않는다**. 스킬 본문이 설정 파일을 읽어 분기한다 — `feedback.repository`가 이미 그 상태이며(프로덕션 호출자 0건), 키 인벤토리에서 `R`(reserved)로 분류돼 있다.

그럼에도 Go 측 필드·기본값·접근자를 두는 이유 3가지:

1. `TestShippedConfigKeysHaveReaders`가 등록을 요구한다.
2. 웹 콘솔의 seam 쓰기가 스키마 필드를 요구한다.
3. 마법사가 값을 기록할 때 기본값 상수가 단일 원천이어야 한다.

접근자(`FeedbackAutoSubmit()`)는 당장 호출자가 없을 수 있다 — 그 사실을 키 인벤토리 항목에 정직하게 기록한다(`evidence` 필드).

## §9 실패 모드 매트릭스

| 실패 | 탐지 | 결과 | 근거 |
|---|---|---|---|
| 스크러버 바이너리 부재 | 명령 실행 실패 | 제출 금지 | fail-closed (REQ-3) |
| 정책 로드 실패 | 종료 코드 ≠ 0 | 제출 금지 | fail-closed |
| 정규식 컴파일 실패(사용자 확장 패턴) | `compilePatterns`가 `slog.Warn` 후 건너뜀 | 그 패턴만 무시, 스크럽 계속 | 기존 `pre_tool.go:294` 태도 계승 |
| 마스킹 로그 쓰기 실패 | 파일 열기 실패 | 스크럽 정상 완료 | fail-open (REQ-8) |
| `HOME` 미설정 | `paths.Home()` 에러 | 홈 축약만 생략, 나머지 변환 수행 | 부분 실패가 전체 실패가 되지 않게 |
| 이슈 생성 실패 | `gh` 종료 코드 | 큐 적재 | REQ-9 |
| 큐 쓰기 실패 | `Mutate()` 에러 | 사용자에게 보고 + 본문을 화면에 출력 | 조용한 폐기 금지(REQ-9)의 최후 수단 |

## §10 후속 카드 후보 (이 SPEC 범위 밖)

- `gh issue create`를 Go 명령으로 감싸 강제 지점을 코드 안으로 옮기기 — §1 경계의 산문 구간을 줄인다.
- 중복 검색을 스크럽 이후로 옮기거나 마스킹된 제목으로 수행하기 — `spec.md` §E.3 두 번째 잔여 위험.
- `hook`의 `sensitiveContentPatterns` 원본에 `AIza` 반영(Write/Edit deny 판정 확대) — 기존 동작 변경이라 별도 판단.
- 스크러버를 `internal/telemetry` 트레이스에 적용.
