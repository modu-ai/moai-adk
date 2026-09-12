# SPEC Review Report: SPEC-MOAI-GATEWAY-001

Iteration: 6 — 운영자가 허용한 변경분 한정 재감사
Verdict: PASS
Overall Score: 0.92 (변경분 점수; Tier L 기준 0.85)

Reasoning context ignored per M1 Context Isolation.

작성자의 설명과 과거 감사의 결론을 현재 근거로 사용하지 않았다. iter5 보고서는 재검사할 항목을 정하는 데만 사용했고, 두 dispatch brief는 승인된 수정 범위를 확인하는 데만 사용했다. 판정 입력은 최종 0.7.0 문서와 아래 기준선의 코드다. 이번 결과는 계획 변경분의 PASS이며, 제품 구현·클라이언트 호환성·실계정 인증의 PASS가 아니다.

## Claim

G5-B1·G5-B2·G5-B3은 최종 문서에서 해소되었다. GLM 슬롯 키를 다시 넣는 순서와 catalog 등록 전제까지 교차 확인했으며, 그 변경 때문에 새로 생긴 차단 결함은 찾지 못했다. G5-A1~A4도 반영되었다. 선택 개선 사항 한 건은 아래 G6-A1로 남긴다.

읽기 전에 정한 실패 가설은 다음과 같다.

- 상속 키를 지우지 않거나, launcher가 추가한 값까지 나중에 지우는 순서 오류
- GLM credential의 무조건 삭제 또는 상속 credential의 재사용
- GLM 슬롯 값과 catalog의 route ID가 서로 연결되지 않는 누락
- PICKER 착지 전에 코어가 노출될 수 있게 남은 출시 계약
- 공유 `teammateMode`의 위험을 닫았다고 다시 주장하는 문구
- 바뀐 AC가 다른 분기만 검사하거나 실측하지 않은 client 동작을 PASS로 취급하는 경우

## Must-Pass Results

- **MP-1 PASS — 번호와 추적 구조.** 현재 파일에서 REQ 정의 26개가 001~026으로 연속이고 중복이 없다. 폐기 묘비 007·020을 제외한 live REQ 24개와 live AC 24개를 측정했다. AC 헤더가 인용한 live REQ의 누락과 잘못된 참조가 모두 없다. E3.
- **MP-2 PASS — 이번에 바뀐 요구사항 계층.** `spec.md:517` REQ-018, `:528` REQ-019, `:545` REQ-021은 Ubiquitous/While 의무이며, `:647` REQ-022는 유효한 호환 기간 안의 기존 부정형 EARS다. AC는 Given-When-Then 검증 계층으로 판정했다. 이번 감사에서 변경되지 않은 REQ 전체의 의미를 처음부터 재심사하지 않았다. 구조 검사 E2는 현재 문서 전체에 실행했다.
- **MP-3 PASS — 현재 frontmatter.** `spec.md:2-14`에서 canonical 필드 12개와 `tier: L`을 직접 확인했다. `version: "0.7.0"`, `status: draft`, 날짜 둘, `priority: P1`, `phase: "v3.3.0 target"`, `lifecycle: spec-anchored`, 문자열 `tags`가 있다. 이 트리 바이너리의 lint 출력은 E2다.
- **MP-4 N/A — 단일 저장소 구현.** `spec.md:11`은 `internal/gateway, internal/cli, internal/kanban, internal/hook`을 대상으로 한다. 범용 언어 template을 바꾸는 SPEC이 아니다.
- **MP-5 PASS — 변경분의 형제 SPEC 경계.** E3의 현행 경로 검사에서 참조된 retired/superseded/archived SPEC은 없다. 네 형제 SPEC은 아직 없고 `spec.md:756`, `:767`, `:780`, `:798`에서 제안으로 명시한다. PROXY 식별자는 `:834`의 과거 보고서 경로다. 존재하지 않는 PICKER를 이미 착지했다고 판정하지 않았다. `plan.md:176`과 `design.md:808`은 그 착지를 기다리는 조건이다.
- **MP-6 PASS — 변경분의 플랫폼 경계.** 현재 `syscall` 언급 여섯 곳은 보존할 기존 POSIX 경로와 그 이력이다(E5). `spec.md:439-444`는 POSIX와 Windows 종료 경로를 구분하고, 같은 프로세스 요구사항 절의 REQ-009(`:465-468`)은 Windows에 spawn-and-wait를 요구하며 **"POSIX 계약을 그대로 가져다 쓰지 않는다"**고 명시한다. 이 명시적 플랫폼 예외에 근거한 판정이다. 단순히 `//go:build` 문자열이 없다는 것만으로 Windows에도 POSIX 호출을 요구한다고 해석하지 않았다. 이번 수정은 새 syscall을 도입하지 않는다. Windows 실행 결과는 검사하지 않았다.
- **MP-7 PASS — clarification 표지.** 현재 `plan.md`와 `research.md`에 `[NEEDS CLARIFICATION` 표지가 각각 0개다(E3). M0 운반 키 결정과 M1 실측 게이트가 완료되었다는 뜻은 아니다.
- **MP-8 N/A — release-blocking으로 분류된 AC 없음.** 현재 `acceptance.md`에는 개별 AC의 `release-blocking` 분류와 RED-now 셀이 없다(E3 및 본문 판독). `AC-MG-006`의 Windows release PR 판정은 향후 실행 위치를 정한 조항이며 RED-now 시험을 인용한 셀이 아니다. 재실행할 RED-now 명령은 없으므로 재현 PASS를 주장하지 않는다.

## Category Scores

점수는 승인된 변경분과 그 직접 회귀에 적용했다. 과거 전체 감사 점수를 새로운 전체 측정치로 복제하지 않았다.

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 1.00 | 1.00 | `spec.md:580-604`의 정리→후주입, launcher별 Z.AI/슬롯 구분; `:653-657`의 잔여 위험 |
| Completeness | 1.00 | 1.00 | `spec.md:530-539` GLM catalog와 출시 조건; `plan.md:142-143`, `:176-180`; `design.md:726-739`, `:808-812` |
| Testability | 1.00 | 1.00 | `acceptance.md:267-285`는 env 값·키 유무, `:298-308`은 tmux 밖 값·대조군, `:444-453`은 route와 mock 수신을 이진 판정 |
| Traceability | 0.75 | 0.75 | REQ/AC 헤더 연결은 전부 존재(E3). 다만 새 프로세스 env 의무의 14키 중 7키는 오염 픽스처에 직접 들어가지 않는다(G6-A1) |

조화평균: `4 / (1/1 + 1/1 + 1/1 + 1/0.75) = 0.923076…`, 표기 0.92. iter5 0.83보다 낮아지지 않았으므로 STOP 신호는 없다.

## Defects Found

D1. **G6-A1 — 상속 env 픽스처가 정리 집합 전체를 오염시키지 않음** — `acceptance.md:260-266` — Severity: minor — Class: optional — Confidence: high.

REQ-021(`spec.md:580-581`)의 상속 정리 의무는 14키 전부다. 파일 env 픽스처(`acceptance.md:249-256`)는 14키 전부를 넣지만, 별도로 추가된 **프로세스 env** 픽스처는 그 집합 가운데 BASE_URL, AUTH_TOKEN, OPUS·SONNET·HAIKU, MAX_CONTEXT_TOKENS, BACKUP_AUTH_TOKEN의 7키만 넣는다. `Z_AI_API_KEY`는 별도 추가 키다. FABLE, DISABLE_EXPERIMENTAL_BETAS, API_TIMEOUT_MS, AUTO_COMPACT_WINDOW, DISABLE_NONESSENTIAL_TRAFFIC, TEAMMATE_DISPLAY, STATUSLINE_CONTEXT_SIZE는 상속 표지가 없다. 따라서 문서에 열거된 픽스처만 사용하면 이 일곱 키의 개별 정리 누락을 직접 구분할 수 없다. 이 지적은 구현 결함이나 런타임 유출을 관측했다는 주장이 아니다.

선택 개선: 새 REQ/AC를 만들지 말고 기존 프로세스 env 픽스처에 동일한 14키 집합을 모두 서로 다른 상속 표지로 넣는다. 기존 값 부재 판정과 GLM 후주입 값 비교를 그대로 적용하면 된다. G5-B1이 요구했던 대표 오염 env·정리 순서·credential 대조는 이미 판정 가능하므로 이번 변경분 PASS를 막지는 않는다.

## Regression Check

| 이전 항목 | 판정 | 현재 근거 |
|---|---|---|
| G5-B1 상속 GLM 키 | RESOLVED | `spec.md:580-604`: 14키 선삭제, cc·gpt의 Z_AI 삭제, glm의 저장소 값, 운반 키 두 경우, 네 GLM 슬롯 후주입. `design.md:143-167`, `:546-563`. `acceptance.md:260-285`의 오염 표지·순서·저장소·슬롯 대조 |
| G5-B2 core-only 출시 | RESOLVED | `spec.md:534-539`가 빈 기본값의 중간 상태를 명시. `plan.md:176-180`, `design.md:808-812`가 PICKER 착지 전 코어 노출·출시를 금지. `acceptance.md:29-32`도 제외 경계와 연결 |
| G5-B3 in-process 과장 | RESOLVED | `spec.md:653-657`가 닫지 못한다고 명시하고 공유 파일, 읽기 시점 미측정, stale tmux 키 세 가지를 열거. `design.md:663-677`, `acceptance.md:418-420`도 같은 경계 |
| G5-A1 tmux 밖 판정 | RESOLVED | `acceptance.md:298-308`: TMUX 없는 auto→in-process 및 gateway 아닌 대조군 auto |
| G5-A2 인식기 변형 | RESOLVED | `acceptance.md:59-67` 아홉 변형에 max_tokens 0, stream null, 단일 system이 있다. `:69-75`는 형식 검사와의 관계 및 도달 대조군을 명시 |
| G5-A3 낡은 포인터 | RESOLVED | `research.md:923-924`는 결정 10 측정을 PICKER 형제로 연결 |
| G5-A4 증거 위치 | RESOLVED | `plan.md:108-115`는 원본 fixture와 게이트 보고서 경로를 명시. E4에서 해당 보고서·fixture 경로의 ignore 결과가 빈 출력/exit 1 |

GLM 슬롯 보정의 직접 회귀도 다음과 같이 확인했다.

- 코드의 기존 대응은 `glm.go:367-370`의 OPUS←High, SONNET←Medium, HAIKU←Low, FABLE←Fable이다. `spec.md:595-597`와 같다.
- `spec.md:530-532`와 `plan.md:142-143`은 설정된 네 tier ID를 Z.AI route에 등록하며 동일 ID의 여러 tier는 하나로 묶도록 한다. 슬롯만 추가하고 catalog 등록을 빠뜨린 계약이 아니다.
- `acceptance.md:271-285`는 서로 다른 tier 값으로 누락·상속값 재사용·순서 뒤집힘·tier 교환을 구별한다. `:444-448`은 각 ID의 registry 해석과 Z.AI mock 수신을 따로 판정한다.
- `spec.md:730-733`과 `design.md:728`은 Claude Code가 별칭을 슬롯 값으로 바꾸는 동작을 **미측정**으로 남긴다. 이 감사는 실제 subagent 요청의 model 값을 관측한 것으로 취급하지 않았다.
- `design.md:735-739`는 GLM OPUS 슬롯이 picker Default 행에 미칠 수 있는 영향을 PICKER에 남기며 코어의 출시 조건과 모순되지 않는다.

## Evidence

모든 명령은 이번 감사에서 실행했다. 아래 인용 출력은 요약문이 아니라 관측한 출력이다. 코드·문서 판독은 위에 적은 현재 파일 행을 사용했다.

### E1 — 기준선

```text
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified branch --show-current
WT-unified-gateway
```

각 exit code: 0.

### E2 — 현재 트리 바이너리의 구조 검사

실행 작업 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.

```text
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

Exit code: 0. 바이너리는 부모 세션이 이 HEAD에서 빌드해 제공했다. 감사자는 그 빌드를 다시 실행하지 않았고 위 lint는 직접 실행했다. 구조 lint는 의미·런타임 검증의 대체 근거가 아니다.

### E3 — 현재 파일의 번호·참조·표지·해시 검사

실행 명령: `python3 - <<'PY'`의 단일 read-only 스크립트. 검사에 사용한 식은 다음과 같다. `s`는 E1 작업 트리의 `.moai/specs/SPEC-MOAI-GATEWAY-001`이다.

```python
rq = re.findall(r'^\*\*REQ-MG-(\d{3})\*\*', t, re.M)
live = re.findall(r'^\*\*REQ-MG-(\d{3})\*\* \(', t, re.M)
ac = re.findall(r'^\*\*AC-MG-(\d{3})\*\* \(', a, re.M)
# AC 정의 헤더만 골라 REQ-MG-(\d{3})를 모으고 live 집합과 양방향 차집합 비교
# plan.md/research.md의 '[NEEDS CLARIFICATION' 개수
# acceptance.md의 r'release[- ]blocking|출시 차단 AC|릴리스 차단 AC' 개수
# spec.md의 SPEC 참조별 파일 존재와 status, 여섯 파일 byte 수와 sha256
```

관측 출력(exit 0):

```text
REQ definitions: 26 sequential: True unique: True
Live REQ: 24 live AC: 24 uncovered: [] invalid AC refs: []
plan.md NEEDS CLARIFICATION count: 0
research.md NEEDS CLARIFICATION count: 0
Explicit release-blocking classifications: 0
D7 SPEC-MOAI-CG-RETIRE-001 not found
D7 SPEC-MOAI-GATEWAY-001 draft
D7 SPEC-MOAI-GATEWAY-PICKER-001 not found
D7 SPEC-MOAI-GATEWAY-TEAMMATE-001 not found
D7 SPEC-MOAI-GPT-AUTH-001 not found
D7 SPEC-MOAI-PROXY-001 not found
spec.md 74402 8e139af052b3e97bc631f8b24dd55fe273950fe79283d17f1a036406470eaf4d
plan.md 47802 845ad4ac4f503f2a01f910d385bd808f0ebe5b7c6f28c71172a08d387c4b5244
acceptance.md 46356 558ec342df50a5375da0d1807aae780e1cf7ec637b779f4a4ca7dacda2b41b40
design.md 72838 1a1f3a047de8f632607819c551e44f3dbacccf1c86e350fb8aa4e2b1bc7c1f53
research.md 72555 89db1a59b3ead01f24432b23377838b4b970f65ea904007be7f7caa29fbf0d78
progress.md 5468 7b7e3185c2876e8a0c7ff7f045e60bb024ce7a1128c260627e754b7f5ffb9650
```

### E4 — 보고서와 fixture의 추적 가능 경로

```text
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified check-ignore .moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter6.md internal/gateway/testdata/validation-capture/claude-code-2.1.268/README.md
```

Stdout: 빈 출력. Exit code: 1. ignore되지 않았다는 뜻이며 아직 생성되지 않은 fixture가 존재한다는 뜻은 아니다.

### E5 — syscall 언급의 현재 위치

```text
$ rg -n 'syscall|//go:build|cross-platform exemption|EXCL.*syscall' /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
35:  - **D2 — 프로세스 모델은 분리된 detached child이며 `syscall.Exec`은 보존한다.**
37:    `syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))`는 현재 프로세스를
96:  `MOAI_SESSION_PID` 각인·signal·PTY·job control은 POSIX에서 `syscall.Exec` 보존으로부터
98:  경로에서 나온다. (0.2.0 원문은 다섯 보장이 모두 `syscall.Exec`에서 파생된다고 적었으며,
439:POSIX에서는 Claude child를 띄우는 `syscall.Exec` 호출을 보존하며, POSIX의 이 보장들은 그
707:- POSIX `syscall.Exec` 보존은 협상 대상이 아니다. `MOAI_SESSION_PID` 각인이 여기에
```

Exit code: 0. 플랫폼 예외 문장의 판독은 MP-6에 따로 적었다.

## Baseline-attribution

- 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- 브랜치·HEAD: E1. SPEC은 아직 untracked이며 내용 식별은 E3의 여섯 해시로 고정했다.
- 감사 일자: 2026-09-11. 대상 버전: 0.7.0. Tier L 다섯 문서와 `progress.md`를 변경 관련 절 중심으로 읽었다.
- 이번 감사는 SPEC/code를 수정하거나 Git 상태를 변경하지 않았다. 새로 작성한 것은 이 보고서뿐이다.
- 원격 develop과의 일치는 측정하지 않았다. 이전 진행 기록의 `0 0`을 현재 원격 상태로 재사용하지 않는다.

## Gaps

- M0의 실제 Claude 구독 OAuth와 gateway 세션 토큰 공존, settings env 우선순위, 운반 키 결정은 실행하지 않았다. 그 결과가 음성이면 계획에 정한 게이트로 돌아가야 한다.
- 실제 Claude Code의 모델 별칭→슬롯 해석, subagent/teammate의 model 값, MCP 헤더 확장과 실계정 Z.AI 왕복은 실행하지 않았다.
- M1 원본 캡처, 억제 플래그 없는 요청 목록, Go gateway를 통한 연속 provider 전환은 실행하지 않았다. `plan.md:119`의 버전별 재측정 의무는 남아 있다.
- 제품 구현 테스트, Windows 실행, 전체 저장소 회귀 시험은 실행하지 않았다. 이번 요청은 계획 변경분 감사다.
- 감사 모델 설정 검색에서 명시적인 `audit_model` 설정을 찾지 못했으며, 외부 codex/GLM second opinion은 실행하지 않았다. 독립된 본 감사 한 건의 판정이다.
- 이전 0.6.0 원본이 untracked이므로 byte 단위 before/after diff로 모든 변경 여부를 증명하지 못했다. 승인된 항목과 최종 문서/코드의 일치만 판정했다.

## Residual-risk

- G6-A1의 대표 키 픽스처는 전체 정리 집합의 개별 누락까지 직접 잡지 않는다.
- GLM 슬롯과 MCP 키를 launcher별로 고정하는 설계는 클라이언트의 실제 적용 방식에 의존한다. 현재 문서의 미측정 표시는 이 불확실성을 없애지 않는다.
- 공유 `teammateMode`와 stale tmux env 경로는 의도적으로 완전 차단하지 않았다. 이를 안전하다고 다시 설명하면 이번 PASS의 근거가 무너진다.
- PICKER는 제안 상태다. 이번 계획 PASS는 형제 SPEC 착지 전 코어 노출·출시를 허용하지 않는다.

## Recommendation

승인된 run 단계의 M0 측정부터 진행할 수 있다. 이 판정으로 M0/M1의 실측 게이트를 건너뛰거나 출시 조건을 해제해서는 안 된다. 선택 사항 G6-A1은 기존 픽스처 안에서 보강할 수 있으며, 별도 요구사항이나 새 SPEC을 만들 필요는 없다.

## Iteration History

- iter4: FAIL 0.80, STOP. 범위 분리 후 추가 감사가 승인되었다(과거 보고서의 이력).
- iter5: FAIL 0.83. B1~B3과 A1~A4를 이번 재검사 목록으로 사용했다.
- iter6: API 제한으로 중단된 감사를 재개했고, 현재 0.7.0 내용에 대해 PASS 0.92를 수출했다. 구현 판정은 아직 없다.
