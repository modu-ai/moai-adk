# t529 판정서 — 세션 워크트리 이동이 배경 에이전트의 Bash 를 막는 건

- 카드: t529
- 브랜치: `WT-session-anchor-guard` (워크트리 `.claude/worktrees/t529`)
- 기반 SHA: `1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09` (origin/develop)
- 측정 시각: 2026-09-12T10:53~10:55Z
- 단계: **판정만**. 카드의 [HARD] 지시대로 수리는 하지 않았다.

---

## 결론 먼저

| 선택지 | 판정 |
|---|---|
| (a) 가드가 옳고 운영 규율로 막는다 | **부분 채택** — 옳지만 이것만으로는 불충분 |
| (b) 가드가 에이전트별 앵커를 봐야 한다 | **기각** — MoAI 의 코드가 아니다. 상류가 방금 닫은 구멍을 다시 열라는 요구가 된다 |
| (c) 막혔다는 사실이 산출물에 강제로 기록되게 한다 | **채택** — 그리고 카드가 예상한 것보다 싸다. 기록 층이 이미 있다 |

카드가 세운 전제 하나가 **반증됐다**: 이 거부는 조용하지 않다. 기계 층에 이미 남고 있다.

---

## F1 — 가드는 MoAI 것이 아니라 Claude Code 런타임 것이다

**Claim**: 거부를 내는 주체는 Claude Code 런타임이며, MoAI 에는 그 문자열의 생산자가 없다.

**Evidence**

```
$ grep -rn "isolated in the worktree\|shared checkout" --include="*.go" --include="*.sh" --include="*.md" --include="*.json" .
```

적중 20건 전부가 문서·SPEC·연구노트의 *인용*이다. `.go` / `.sh` 생산자는 0건.

상류의 설계 의도는 변경이력이 직접 말한다 (`.moai/research/cc-changelog-snapshot-2.1.233.md`, 2.1.216 구간):

```
- Fixed worktree-isolated subagents redirecting git into the shared checkout via `git -C`, `--git-dir`, or `GIT_DIR`/`GIT_WORK_TREE`
- Fixed subagents in background sessions bypassing the worktree-isolation guard and writing to the shared checkout
```

**Baseline-attribution**: 이 트리(`.claude/worktrees/t529`, base `1d150a27d`)에서 2026-09-12 에 실행한 grep 과 그 출력.

**함의 — (b) 기각의 근거**: 두 줄 모두 *가드를 우회하는 것*을 결함으로 취급해 닫았다. 그중 둘째 줄은 **배경 세션의 서브에이전트**가 바로 그 우회를 하던 것을 막은 기록이다. 즉 "에이전트가 자기 트리에 고정돼 있으니 세션 이동과 무관해야 한다"는 카드의 반증된 가정은, 상류에서는 *가정이 아니라 이미 거부된 설계안*이다. (b) 를 요구하는 것은 상류가 방금 봉한 경로를 다시 여는 일이고, 가드의 목적(한 세션이 여러 트리에 쓰는 것을 막음)과 정면으로 충돌한다. MoAI 쪽에 고칠 코드가 없다는 점은 그 위의 이야기다.

---

## F2 — 거부는 기록된다. 카드의 "조용하다"는 전제는 기계 층에서 거짓이다

**Claim**: 가드 거부는 `.moai/lessons-inbox.jsonl` 에 한 줄씩 남는다.

**Evidence** — 사전 스냅샷 → 거부 유발 → 사후 판독.

거부 유발 1 (cross-tree `git -C`):

```
$ git -C /Users/goos/MoAI/moai-adk-go status --porcelain
This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529,
but this command redirects git to the shared checkout via -C. Refusing to run it — ...
```

그 결과 남은 줄 (primary 체크아웃의 인박스, 9162행):

```json
{"timestamp":"2026-09-12T10:53:59Z","event_key":"tool_failure:Bash:UnknownFailure",
 "summary":"This session is isolated in the worktree .../t529, but this command redirects git to the shared checkout via -C. Refusing to run it — a worktree-isolated ses…",
 "source":"tool:Bash","v":1}
```

역사적 누적:

```
$ grep -c "Refusing to run" /Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl
2730
```

그리고 이것은 내 프로브만의 현상이 아니다 — 같은 날 **다른 세션**이 같은 거부를 맞은 줄이 인박스 9137행에 있다 (`.../worktrees/t606`, 10:36:13Z). 오늘 안에 실제로 재발한 사건이다.

**Baseline-attribution**: 이 트리에서, 2026-09-12 에, 프로브 전후로 같은 파일을 두 번 읽어 얻은 차이.

**기록 경로** (코드로 확인): `PostToolUseFailure` 훅 → `internal/hook/post_tool_failure.go:Handle` → `recordToolFailureEvent(input, category)`. 템플릿 배선은 `internal/template/templates/.claude/settings.json.tmpl:226` 의 `PostToolUseFailure` 항목.

**정정 — 증거 writer 는 이걸 못 본다.** 같은 프로브 동안 `.moai/evolution/telemetry/usage-2026-09-12.jsonl` 은 760줄에서 움직이지 않았다. `PostToolUse` 계열(`evidence_writer.go`)은 도구가 실제로 돈 경우에만 불린다. 거부를 잡는 것은 오직 `PostToolUseFailure` 하나다. 이 구분이 (c) 의 설계에 그대로 걸린다.

---

## F3 — 자진 보고는 우연이 아니라 규약이다. 다만 약하고, 하중을 받는 표면에 없다

**Claim**: lane-3 사례에서 에이전트가 이상을 보고한 것은 규약의 산물이다.

**Evidence**

- `.claude/agents/moai/sync-auditor.md:108` — [HARD] 판정서는 5절 형식(Claim / Evidence / Baseline-attribution / **Gaps** / Residual-risk)으로 내보내야 한다. Gaps 는 "명시적으로 관측하지 **않은** 것"을 담는 칸이다.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1·§2 — 실행하지 않은 명령에 귀속된 PASS 는 그 자체로 위반이다.

즉 "측정으로 확인"이 "판독으로 추정"으로 바뀌었는데 판정서가 그대로 PASS 를 내면, 그것은 이미 기존 독트린 위반이다. 새 규칙이 필요한 상황이 아니다.

**Gap**: 그러나 이 의무는 전부 **정책 층**이다. 에이전트가 "돌렸다고 주장하는 명령"과 "인박스가 기록한 거부"를 대조하는 기계는 없다. 규약을 지키는 에이전트는 보고하고, 지키지 않는 에이전트는 조용히 통과한다 — 그리고 리드는 그 차이를 못 본다. 카드가 지목한 위험은 정확히 여기에 남아 있다.

---

## F4 — 기록은 있으나 지금 형태로는 탐지에 못 쓴다

**Claim**: 거부 줄의 `event_key` 가 만능 버킷이라 다른 실패와 구별되지 않는다.

**Evidence**: 기록된 키는 `tool_failure:Bash:UnknownFailure`. 분류기(`post_tool_failure.go:classifyError`)는 순서대로 매칭하는데, 가드 거부문에는 `timeout` / `permission denied` / `sandbox` / `exit status` 어느 토큰도 없다. 의미상 정책 거부이므로 `PermissionDenied` 가 맞는 자리인데, 매처가 문자열 `"permission denied"` 만 보기 때문에 catch-all 로 떨어진다.

**함의**: 2730줄이 남아 있어도 키만으로는 골라낼 수 없다. 지금은 `summary` 본문을 문자열로 긁어야 하고, 그건 상류 문구가 바뀌면 조용히 깨지는 종류의 탐지다.

---

## 권고 — (c) 를 두 조각으로

카드는 "(c) 가 가장 값싸 보이지만 확인 필요"라고 했다. 확인 결과 **예상보다 더 싸다**: 기록 기제를 새로 만들 필요가 없다. 빠진 것은 두 가지뿐이다.

1. **이름 붙이기** — 가드 거부를 고유 범주로 분류해 `event_key` 만으로 골라낼 수 있게 한다. 안정적인 토큰(`isolated in the worktree` + `Refusing to run`)을 매처에 넣는다. 상류 문구 의존이라는 취약점이 남으므로, 그 취약점 자체를 테스트로 고정한다.
2. **판정서에 접붙이기** — 감사 창(window) 동안 그 세션에 귀속된 거부 줄이 하나라도 있으면, 판정서의 Gaps 절이 그 퇴화를 명시해야 한다. 리드가 증거를 *읽는* 기존 규율(`kanban-dispatch.md` § Completion is read, never trusted)에 그대로 얹히는 읽기 작업이다.

(a) 는 폐기하지 않는다 — 배경 에이전트가 살아 있는 동안 세션을 옮기지 않는 것은 공짜이고 옳다. 다만 강제할 수단이 없고, 오늘 하루에만 최소 2회 발생했으므로 **유일한 답이 될 수 없다**.

---

## Gaps — 관측하지 않은 것

- **카드 원형의 재현을 하지 않았다.** 배경 에이전트를 띄우고 세션을 옮기는 절차는 카드가 경고한 대로 세션 안정성을 흔든다. 대신 **같은 가드의 형제 거부**(cross-tree `git -C`)로 기록 경로를 측정했다. 두 거부문은 같은 가드가 같은 판단(세션 앵커 대 명령 대상)으로 내는 것이지만, 배경 에이전트 경로에서 `PostToolUseFailure` 가 동일하게 발화하는지는 **직접 보지 않았다**. (c) 착수 전 이 한 가지는 반드시 실측해야 한다.
- 인박스 기록이 세션 id 를 담지 않는다는 점 — 위 권고 2의 "그 세션에 귀속된"을 무엇으로 판정할지는 미검증이다. 기록된 필드는 `timestamp` / `event_key` / `summary` / `source` / `v` 뿐이고, summary 안의 워크트리 경로가 유일한 실마리다.
- plan-auditor 가 같은 형태를 맞았을 것이라는 lane-3 의 예상은 확인하지 않았다.

## Residual-risk

- 탐지를 상류 문구에 매다는 이상, Claude Code 가 거부문을 바꾸면 조용히 멎는다. 이름 붙이기의 구조적 한계다.
- 인박스는 primary 체크아웃 한 곳에 모인다. 워크트리 세션의 거부도 거기로 가므로 수집은 되지만, 워크트리가 폐기되면 경로 문자열만 남고 대조 대상이 사라진다.

---

## 부수 관측 — 별 카드 소관

1. **통합 창 락의 브랜치 기록 어긋남** (카드 본문이 이미 지목). 세션 앵커를 브랜치로 기록해 실제 병합 대상과 갈린다. 같은 뿌리(세션 앵커 대 실제 대상)지만 축이 다르다.
2. **가드의 복합 명령 거부 범위.** `{ wc; grep; tail; }` 처럼 git 이 전혀 없는 복합 명령도 "too complex to verify" 로 거부되며, 거부문은 그것을 "git operations" 라 부른다. 측정됨 (2026-09-12T10:54:13Z, 인박스 9165행). 과다 매칭으로 보이나 이 카드의 축이 아니다.

🗿 MoAI

---

# 2차 — (c) 구현 (리드 판정 2026-09-12, t529 안에서 진행)

리드가 (c) 2조각을 이 카드 안에서 진행하도록 판정했다. 순서: Gaps 실측 → 1조각(분류) → 2조각(규약).

## G1 — 남겨둔 Gap 을 닫았다: 배경 에이전트의 거부도 훅에 닿는다

**Claim**: 서브에이전트가 맞은 가드 거부도 `PostToolUseFailure` 를 거쳐 인박스에 기록된다.

**Evidence** — 카드가 경고한 불안정 절차(세션 이동)는 쓰지 않고, **행위자 축만 분리**했다. 가드와 거부 사유는 그대로 두고 실패 주체만 메인 세션 → 서브에이전트로 바꿨다. 귀속을 위해 내가 쓴 적 없는 형태(`--git-dir`)를 시켰다.

```
10:59:48Z  wc -l .moai/lessons-inbox.jsonl → 9172        (프로브 전)
           서브에이전트: git --git-dir=<primary>/.git status --porcelain
           → "This session is isolated in the worktree .../t529, but this command
              redirects git to the shared checkout via --git-dir. Refusing to run it — ..."
11:00:1xZ  wc -l .moai/lessons-inbox.jsonl → 9173        (프로브 후, +1)
```

새 줄 (tail -1):

```json
{"timestamp":"2026-09-12T11:00:13Z","event_key":"tool_failure:Bash:UnknownFailure",
 "summary":"This session is isolated in the worktree .../t529, but this command redirects git to the shared checkout via --git-dir. Refusing to run it — a worktree-isola…",
 "source":"tool:Bash","v":1}
```

**Baseline-attribution**: 이 트리에서, 프로브 창(10:59:48Z 시작) 전후로 같은 파일을 두 번 세어 얻은 차이. 정확히 +1 이고, 본문의 `--git-dir` 은 이 창에서 서브에이전트만 쓴 형태다.

**부수 확인 — 귀속 정보는 없다.** 그 줄의 필드는 `timestamp` / `event_key` / `summary` / `source` / `v` 뿐이다. 세션 id 도 에이전트 id 도 없고, summary 안의 워크트리 경로가 유일한 실마리다. 1차 Gaps 에 적었던 우려가 실측으로 확인됐다.

## G2 — 1조각: 거부에 고유 이름을 붙였다

`internal/hook/post_tool_failure.go`:

- `WorktreeGuardRefusal` 범주 추가.
- 매처를 `classifyError` 의 **첫 분기**로 배치. 이유 둘: ① OOM 분기가 맨몸 부분문자열 `"137"` 을 보는데 이 저장소는 워크트리를 카드 id 로 짓는다 — `.claude/worktrees/t137` 안의 거부는 그 뒤 어디에 두든 OOMKilled 로 분류된다. ② 거부는 "돌았는데 잘못됐다"가 아니라 "돌지 않았다"라 아래 범주들과 주장 자체가 다르다.
- 앵커는 `isolated in the worktree` **하나**. 뒷절(`Refusing to run it`)까지 요구하지 않은 것은 의도다 — 카드 원문이 인용한 변종은 첫 절까지만 인용돼 있어, 뒷절을 요구하면 **카드가 다루는 바로 그 시나리오를 놓칠 위험**이 있고 그걸 반증할 캡처가 없다.
- 메시지가 증거상 귀결을 말한다: 돌지 않았다 / 이에 기댄 검증은 pass 가 아니라 gap 이다.

## G3 — 뮤턴트를 실제로 죽는 것까지 관측했다

`verification-completeness` §1.1 은 red 를 실제로 봐야 완성이라 한다. 셋 다 관측했다.

| 뮤턴트 | 주입 | 관측된 red |
|---|---|---|
| 범주 자체가 없음 | 구현 전 (RED) | 컴파일 실패 — `undefined: WorktreeGuardRefusal` ×9 |
| 가드 분기를 OOM 뒤로 이동 | 순서 교체 | `--- FAIL: TestClassifyError_GuardRefusal_BeatsOOM` (그 하나만) |
| 앵커를 일반 토큰으로 확대 (`refusing to run`) | 상수 치환 | `--- FAIL: .../cwd_resolved_(card)` + `--- FAIL: ..._AnchorOnly` |

세 번째가 특히 의미 있다 — 앵커를 넓히자 **카드 자신의 변종이 가장 먼저 죽었다.** G2 의 설계 판단이 장식이 아니라 하중을 받고 있다는 직접 증거다.

GREEN 재확인:

```
$ go test ./internal/hook/ -run 'TestClassifyError|TestFormatMessage|TestPostToolFailure|TestPostToolUseFailure' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	0.641s
$ gofmt -l internal/hook/   → (없음)
$ go vet ./internal/hook/   → exit 0
```

## G4 — 선재 red 1건, 스왑으로 귀속했다

`go test ./internal/hook/ -count=1` 전량은 FAIL 이다. `--- FAIL` 은 **정확히 1건**:

```
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.55s)
    session_start_parallel_test.go:97: Handle blocked 550.516875ms waiting for advisory scan; expected deferred (non-blocking) return
```

고립 재실행도 일관되게 실패(515.6ms / 556.0ms)해 경합 탓이 아니다. 그래서 **대조군**을 만들었다 — `git show HEAD:internal/hook/post_tool_failure.go` 로 되돌리고 새 테스트 파일을 치운 뒤 같은 환경에서 같은 테스트를 돌렸다:

```
(base, 내 변경 없음)  --- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.54s)
                        Handle blocked 541.189833ms ... expected deferred
```

**base 에서도 같은 범위로 실패한다 → 선재 red, 내 변경과 무관.** 부하: 측정 시 16.87 / 대조군 시 24.51. 이 테스트가 이 머신에서 환경 민감하고 CI 에서는 재현되지 않는다는 기록이 `SPEC-KANBAN-RECORD-SESSION-KEY-001/progress.md` 에 이미 있고, 그 카드도 같은 스왑 방법으로 귀속했다.

## G5 — 2조각: 판정서 Gaps 규약

`verification-claim-integrity.md` §3.1 신설 — 거부된 명령에 기댄 검증은 Gaps 절에 **거부 사실과 대신 한 일**을 적는다. 두 가지를 명시했다: ① 거부가 판정을 막지는 않는다(대체 수단으로 PASS 를 내도 된다 — 대체했다는 사실을 숨기는 것이 금지다) ② 거부는 기계적으로도 기록되므로 Gaps 항목은 **작성자가 쓰지 않은 기록**과 대조 가능하다.

Template-First 준수 — 두 벌을 각각 썼다:

| 사본 | 경로 | 차이 |
|---|---|---|
| 로컬 | `.claude/rules/moai/core/verification-claim-integrity.md` | 카드 출처(t529) 포함 |
| 템플릿 | `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` | 중립화 — 카드 id·SPEC ID 0건 |

`make build` 실행(임베드 갱신; `catalog.yaml` 은 바이트 동일). 중립성·미러 가드:

```
$ go test ./internal/template/ -run 'Neutral|Leak|Mirror|Rules' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	3.630s
```

## Gaps — 이번에도 안 본 것

- **카드 원형(세션 이동 + 고정된 배경 에이전트)은 끝내 재현하지 않았다.** 행위자 축은 분리해 측정했지만, 세션이 이동해 앵커가 어긋나는 경로에서도 훅이 같은 줄을 남기는지는 직접 보지 않았다. 거부 주체·거부 경로가 같으므로 같으리라 보지만, 그것은 추론이다.
- **§3.1 규약의 준수 여부를 재는 기계는 만들지 않았다.** 지금은 Gaps 절과 인박스 기록을 사람이 대조해야 한다. 자동 대조는 G1 이 확인한 귀속 정보 부재(세션 id 없음) 때문에 현재 필드만으로는 불가능하다.
- 로컬 rule 사본은 `moai update` 의 wipe 대상이라 템플릿에서 복원된다. 즉 로컬판의 출처 문장은 update 후 사라진다 — 의도된 분기지만 명시해 둔다.

## Residual-risk

- 탐지가 상류 문구(`isolated in the worktree`)에 매달려 있다. Claude Code 가 문구를 바꾸면 분류가 조용히 UnknownFailure 로 되돌아간다. 테스트가 그 의존을 기록하지만 막지는 못한다.
- 첫 분기 배치는 OOM 의 `"137"` 충돌을 피하지만, 그 충돌 자체(경로·SHA 의 `137` 이 OOM 으로 분류되는 것)는 이 카드 범위 밖에 그대로 남아 있다.

🗿 MoAI
