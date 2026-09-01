---
id: SPEC-CODEX-REVIEW-TARGET-001
title: "구현 계획 — codex native review/start target 계약"
version: "0.3.0"
created: 2026-09-01
---

# SPEC-CODEX-REVIEW-TARGET-001 — 구현 계획

## §A 맥락

트리 `.claude/worktrees/t399` @ `442da4f06` (= `origin/develop`), 브랜치 `WT-codex-native-branch`. 카드 t399, 이슈 #1632.

마일스톤은 **되돌리기 어려운 결정을 앞에** 둔다 — M1 이 이 SPEC 에서 유일하게 논쟁적인 결정이고, 나머지는 그 결정의 기계적 귀결이다.

---

## §B 되돌리기 어려운 결정 — 해석 실패 시의 처리 (M1 에서 고정)

REQ-CRT-004 는 **무엇이 일어나면 안 되는지**를 고정했다(review/start 미전송, uncommittedChanges 미대체, 원인 명명). 무엇이 일어나야 하는지는 후보가 둘이고, 이 선택이 도구의 계약을 바꾸므로 착지 후 되돌리기 비싸다.

| 후보 | 형태 | 근거 | 대가 |
|---|---|---|---|
| **(가) 원인을 명명한 `inconclusive`** (권장) | `inconclusiveReview("cannot resolve a base branch in <root>")` — `handleCodexAudit` 의 `applyGateUnmet(out, root)` 를 **통과한다** | fail-open 계약을 건드리지 않는다. `codex_audit` 은 오늘 codex 관련 모든 실패를 구조화 결과로 되돌린다. required 게이트를 켠 프로젝트는 기존과 같은 화면(`GateUnmet` 주석이 달린 `inconclusive`)을 본다 | `inconclusive` 가 하나 늘어난다. 그 값을 통과와 구별해 보이는 일은 t284 소관이라, 소비자에게 얼마나 잘 보이는지는 이 카드 밖 |
| (나) 도구 오류 | `toolErr("codex_audit", err)` — `IsError: true` 인 `CallToolResult` 로 **조기 반환**한다 | `resolveToolProjectRoot` 선례와 같은 모양 — 잘못된 입력은 fallback 하지 않고 거절 | ① 선례가 다루는 것은 **호출자가 고칠 수 있는 입력**이다. base 해석 불가는 트리 상태이지 호출자 입력이 아니다. ② **`applyGateUnmet` 을 아예 지나지 않는다** — `toolErr` 는 그 호출 앞에서 반환하므로, `workflow.audit.gates.codex: required` 인 프로젝트가 보는 것이 구조화 결과에서 Go 오류로 **바뀐다**. 관측 표면 자체가 달라지는 변화이며 문구 수정이 아니다 |

권장은 **(가)**. 이유: 호출자는 `target=baseBranch` 라는 적법한 enum 값을 넘겼을 뿐이고, 고칠 수 있는 것이 없다. 도구 오류는 "네 입력을 고쳐라"는 신호인데 고칠 입력이 없으면 그 신호는 거짓이다. 여기에 (나)의 대가 ②가 더해진다 — required 게이트 소비자의 화면을 바꾸는 것은 이 카드가 요청받은 범위 밖의 효과다.

다만 (가)를 택하면 `inconclusive` 하나가 늘어나므로, **원인 문자열이 반드시 서로 구별 가능해야 한다** — `"codex binary not found in PATH"` 와 base 해석 불가가 구별되지 않으면 이 SPEC 이 만든 것이 새로운 공허함이다. AC-CRT-004 가 관측 필드를 후보별로 고정하고 그 구별 가능성까지 단언하는 이유가 이것이다.

운영자가 (나)를 지시하면 REQ-CRT-004 의 금지 조항 둘(review/start 미전송 · uncommittedChanges 미대체)은 그대로이고, REQ-CRT-008 의 관측 필드가 `Summary` 에서 `IsError` + 오류 텍스트로 옮겨간다. AC-CRT-004 의 표에서 (나) 행이 유효해지고 (가) 행이 무효해진다.

---

## §C 새로 필요한 것 — 브랜치 이름 해석기

`resolveReviewMergeBase` 는 **merge-base SHA** 를 돌려준다. codex `baseBranch` 는 **브랜치 이름**을 요구한다. 같은 폴백 사슬을 이름 층위에서 읽는 함수가 하나 필요하다.

사슬은 `resolveReviewMergeBase` 와 **같은 순서**를 이름 층위에서 읽는다(spec.md §A.7 의 정렬 결정). 설정 키는 읽지 않는다.

1. **원격 기본 head** — `git -C <root> symbolic-ref --short refs/remotes/origin/HEAD` 의 결과에서 `origin/` 접두사를 뗀 이름
2. **`main`** — `origin/main` 또는 로컬 `main` 중 하나라도 ref 로 해석되면 `main`
3. 전부 실패 → 해석 불가 (§B)

**2단계는 원래 두 단계였다 — 합쳤다.** `resolveReviewMergeBase` 의 사슬은 `origin/main` 과 `main` 을 별개 단계로 두지만, 그것은 **merge-base 계산의 인자**로서 서로 다른 ref 를 가리키기 때문이다. 이름 층위로 옮기면 두 단계가 **같은 문자열 `main`** 을 낸다. 구별할 수 없는 두 단계는 어떤 AC 도 가를 수 없으므로(관측 불가능한 분기), 하나로 둔다. 원래 사슬이 표현하던 것 — "원격에 없으면 로컬이라도 본다" — 은 2단계의 `또는` 이 그대로 보존한다.

배치: `internal/cli/mcp_review_material.go` 에 둔다. 그 파일이 이미 `runReviewGit` 과 폴백 사슬을 갖고 있어 새 파일도 새 git 헬퍼도 필요 없다(단순성 사다리 2단).

**[HARD] 해석 가능성을 확인하지 않은 이름은 반환하지 않는다.** 1단계는 `origin/HEAD` 에서 접두사를 떼어 `develop` 같은 이름을 만드는데, **그 이름 자체가 이 트리에서 ref 로 해석된다는 보장이 없다**(원격 추적 ref 는 있고 로컬 브랜치는 없을 수 있다). 2단계도 마찬가지다. 각 단계는 반환 직전에 그 이름이 ref 로 해석되는지 확인하고, 확인에 실패하면 다음 단계로 넘어간다. 확인 없이 이름을 돌려주면 codex 가 해석하지 못할 값으로 리뷰를 시도하고, 그 실패는 이 SPEC 이 닫은 자리에서 다시 `inconclusive` 로 나타난다.

확인의 정확한 형태(`git rev-parse --verify --quiet <name>` 인지, `refs/heads/` + `refs/remotes/origin/` 두 곳을 보는지)는 run 이 고른다. **codex 가 어떤 이름을 수용하는지는 별개 축이고 AC-CRT-010 이 잰다** — 서버가 해석 가능한 이름을 보내는 것과 codex 가 그것을 받아들이는 것은 다른 주장이며, 이 SPEC 은 둘 다 관측한다.

---

## §D 마일스톤

### M1 — 회귀선 고정 + 해석 실패 처리 결정 (우선순위 High)

- AC-CRT-003 (uncommittedChanges 회귀선) 검사를 추가하고 **변경 전 트리에서 초록**임을 관측. 초록이 아니면 즉시 보고하고 중단.
- §B 의 (가)/(나) 결정을 운영자에게 확인하고 고정. 이후 마일스톤은 이 결정에 의존한다.
- 산출: 회귀선 검사 1건, 결정 기록.

### M2 — RED 확립: 계약 층 + 왕복 층 (우선순위 High)

**M2a — 계약 층 RED.** AC-CRT-001 · 002 · 004 · 005 · 006b 를 구현하는 검사를 **프로덕션 변경 없이** 추가. `-v` 로 실행해 `=== RUN` + `--- FAIL` 을 관측하고 `.moai/reports/t399/red/` 에 출력 보존.

**M2b — 왕복 층 RED (AC-CRT-010).** 라이브 검사를 같은 마일스톤에서 추가한다. 계약 층과 함께 두는 이유: 이것이 이 SPEC 이 **얻을 수 있는** 가장 강한 RED 이고(계약 판독에서 유도한 예측이 아니라 실 codex 가 요청을 거절하는 **관측**이 되므로), 늦게 두면 프로덕션 변경이 이미 들어간 뒤라 그 관측을 다시 얻을 수 없다.

[HARD] 이 시점까지 그 관측은 **존재하지 않는다.** 거절은 스키마 판독에서 예측한 것이고, 예측을 관측으로 바꾸는 것이 M2b 의 산출물이다 — 안티패턴 6 이 금지하는 "RED 를 서술로 대체하기"가 바로 이 자리에서 일어날 수 있다.

- 선례 재사용: `codex_review_gate_live_test.go:33` 의 skip 3조건, `codex_live_protocol_probe_test.go` 의 `probeLiveCodex` / `probeInstallRunner` / `probeWriteTranscript`.
- 픽스처는 `probeSeedRepo` 의 변형 — 미커밋 변경이 아니라 **base 브랜치 + 그로부터 갈라진 HEAD** 가 필요하다.
- 선례와 같이 `turn/started` 에서 세션을 끊어 리뷰 turn 전체를 청구서에서 뺀다.
- 산출: JSON-RPC 거절 응답 본문을 `.moai/reports/t399/red/` 에 보존.

산출(M2 전체): RED 로그 2종, progress.md §E.2 인용.

### M3 — 해석기 + 요청 조립 교정 (우선순위 High)

- §C 의 해석기 추가 (2단계 사슬 + ref 해석 가능성 확인).
- `coerceCodexReviewTarget` 을 variant 인지형으로 교정: `uncommittedChanges` 만 bare string 리프트를 허용하고, `baseBranch` 는 해석된 이름을 실으며, required 필드를 채울 수 없는 variant 는 직렬화하지 않는다(REQ-CRT-005).
- `buildCodexReviewParams` / `handleCodexAudit` 에서 해석에 필요한 프로젝트 루트가 조립부까지 닿는지 확인 — `handleCodexAudit` 은 이미 `root` 를 `params["cwd"]` 로 넘기므로 새 배관이 필요한지 먼저 읽는다.
- M2 의 RED 가 전부 초록으로 뒤집히는지 관측 — **계약 층과 왕복 층 둘 다.** 계약 층만 초록이 되고 AC-CRT-010 이 여전히 붉으면, 요청은 스키마를 만족하는데 codex 가 그래도 거절한다는 뜻이며, 그때는 §C 의 이름 해석 규칙이 틀린 것이다(spec.md §F 의 미측정 항목이 답으로 돌아온 경우).

### M4 — 속성형 단언과 공허 검사 교정 (우선순위 Medium)

- AC-CRT-006 / 006b 를 acceptance.md §B 표의 `분류` 열을 순회하는 두 속성 단언으로 작성. 각각 순회한 행 수를 먼저 세고 판정한다.
- AC-CRT-007: `TestCodexAudit_AdversarialDispatchesTurnStart` 의 `baseBranch` 인자를 처리 — 그 검사 안에서 "adversarial 은 target 을 싣지 않는다"를 단언하거나, 인자를 native 검사로 옮긴다.

### M5 — 도구 표면 서술 (우선순위 Low)

- AC-CRT-009: `internal/cli/mcp_server.go:255` 의 `target` 설명에 서버 해석 사실과 해석 원천을 적는다.
- 기계적 변경. 마지막에 둔다.

---

## §E 자기 검증

| 항목 | 명령 |
|---|---|
| 대상 테스트 | `go test ./internal/cli/ -run 'TestCodexAudit_\|TestCodexRPC_\|TestCoerceCodexReviewTarget\|TestResolveCodexBaseBranch' -count=1 -v` |
| 라이브 왕복 (AC-CRT-010) | `go test ./internal/cli/ -run 'TestCodexLive_ReviewStartBaseBranch' -count=1 -v` — skip 이면 **미관측**으로 기록(통과 아님) |
| 정적 검사 (darwin) | `go vet ./internal/cli/...` |
| 정적 검사 (windows) | `GOOS=windows go vet ./internal/cli/...` |
| 범위 침범 확인 | `git diff --stat` — `internal/cli/mcp_convergence.go` 가 목록에 없어야 한다 |
| RED 보존 | `ls .moai/reports/t399/red/` |

로컬 전 패키지 수트(`go test ./...`)는 돌리지 않는다 — 부하 규율(CLAUDE.local §4). 전 패키지 판정은 CI 몫.

---

## §F 안티패턴

1. **반환값으로 판정하기.** 스텁은 요청과 무관하게 스크립트를 되돌려주므로 `pass` 가 나왔다는 사실은 요청이 옳았다는 증거가 아니다. 관측 대상은 `sess.sent[2]` 다.
2. **`inconclusive` 를 통과로 세기.** 이 카드가 고치려는 결함 그 자체.
3. **enum 에 `commit` / `custom` 추가하기.** REQ-CRT-005 는 안전 요구이지 기능 요구가 아니다(spec.md §E).
4. **`disagreement_flag` 나 백엔드 수 건드리기.** t284 축. `mcp_convergence.go` 에 손대는 순간 중복 설계다.
5. **해석 가능성 확인 없이 이름 반환하기.** §C [HARD] 항목. `origin/HEAD` 에서 접두사를 뗀 문자열은 그 자체로 ref 라는 보장이 없다.
6. **RED 를 서술로 대체하기.** "이 검사는 현행 구현에서 실패할 것이다"는 예측이지 관측이 아니다.
7. **AC-CRT-010 의 skip 을 통과로 세기.** codex 부재로 건너뛴 실행은 그 AC 에 대해 아무것도 관측하지 않은 것이다. 0매칭 초록의 한 형태이며, 이 SPEC 이 §A 에서 금지한 바로 그 사고다.
8. **계약 층만으로 회귀선을 세웠다고 보고하기.** spec.md §0 의 방어선은 두 층이다. 계약 층 초록은 "요청이 스키마를 만족한다"까지만 말한다.
9. **`worktree_base_branch` 를 codex 경로에만 배선하기.** spec.md §A.7 의 정렬 결정을 뒤집는 일이며, 그 결정을 바꾸려면 GLM 경로까지 함께 바꾸는 별도 카드다.

---

## §G 참조

- spec.md §A (측정), §E (범위 밖)
- acceptance.md §C (RED 확립 규율)
- `.moai/reports/t399/discovery.md`
