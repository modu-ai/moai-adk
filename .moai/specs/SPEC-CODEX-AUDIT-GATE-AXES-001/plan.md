# SPEC-CODEX-AUDIT-GATE-AXES-001 — 구현 계획

카드 **t686** · Tier **M** · 클래스 **C** · 기준 트리 develop @ `f67d2193f`

마일스톤은 뒤집힐 가능성이 큰 결정부터 적었다. §B 의 결정 두 건은 Implementation Kickoff 에서 운영자 확인을 받아야 한다.

## §A Context

spec.md §A 참조. 세 축은 서로 독립이며, 한 축이 막혀도 나머지는 진행할 수 있다.

## §B 열린 결정 (Kickoff 게이트에서 확정)

### B.1 [NEEDS CLARIFICATION: 축 (b) 집행 메커니즘 선택]

질문: 감사 도구 호출 증거가 없는 판정은 PASS 가 될 수 없어야 하는가? 그렇다면 어디서 막는가?

이 SPEC 의 입장은 "required 게이트가 명시된 프로젝트에서는 PASS 가 될 수 없어야 한다"이다(REQ-CAG-008). 제보자의 관측(건너뛴 3회 PASS 0.92 → 강제 호출 후 FAIL 0.79)이 곧 그 이유다. 남는 문제는 집행 지점이다.

**기존 opt-in Stop-hook(`multi_review_gate.go`)과의 관계.** 이 훅은 세션의 가장 최근 `audit_multi` 수렴 결과를 읽어 required FAIL 일 때만 막는다. 상태 파일이 없으면 ALLOW 한다(결정 순서 4). 감사 도구를 한 번도 부르지 않은 세션은 상태 파일이 없으므로 정확히 이 분기로 통과한다. 게다가 기본 off, 메인 세션 turn-end 범위, `audit_multi` 전용이라 단일 `codex_audit` 호출은 보지 못한다. 즉 이 훅은 "부른 감사가 실패했는가"를 볼 뿐 "감사를 불렀는가"는 보지 않는다. 아래 옵션은 모두 이 빈틈을 메우는 방식의 차이다.

| 옵션 | 무엇을 하는가 | 장점 | 단점 |
|------|--------------|------|------|
| **B-1 서버 발급 영수증 + 판정 수용 지점 검사 (권장 기본안)** | MCP 서버가 `codex_audit` / `audit_multi` 호출마다 영수증(서버 생성 id, 대상 트리, 시각, verdict)을 상태 저장소에 기록하고 결과에 영수증 id 를 싣는다. 감사 보고서는 영수증 id 를 인용한다. 판정을 받아들이는 기계 지점(run-phase 감사 게이트의 plan-audit 보고서 판독, sync 게이트)이 required 일 때 인용 영수증이 저장소에 없거나 대상 트리가 다르면 PASS 를 거부한다 | Claude Code 훅 런타임의 미확인 동작에 기대지 않는다. 사후 감사 가능. 에이전트가 텍스트로 위조할 수 없다(서버 저장소 대조). 단일·다중 경로 모두 포괄 | 검출이 판정 **후**다 — 에이전트는 이미 PASS 를 냈고 재감사가 필요하다. 보고서 형식 의무가 늘어 에이전트 본문(템플릿 미러 포함) 수정이 필요. 수용 지점을 빠뜨리면 구멍이 남는다 |
| **B-2 SubagentStop 훅 차단** | plan-auditor / sync-auditor 가 종료하려 할 때, 해당 서브에이전트 수명 동안 codex 감사 호출 기록이 없으면 `decision: block` 으로 종료를 막고 이유를 돌려준다 | 건너뛰기를 **그 자리에서** 막아 에이전트가 스스로 호출하게 만든다. 재감사 왕복이 없다 | MCP 도구 호출을 특정 서브에이전트 수명에 귀속시키는 방법이 이 트리에서 검증되지 않았다(훅 입력의 agent 식별자 제공 여부 미관측). 훅이 꺼진 환경(`disableAllHooks`)에서는 무력. 차단 루프 상한 관리 필요 |
| **B-3 기존 다중 리뷰 Stop-hook 확장** | `multi_review_gate.go` 의 "상태 파일 없음 → ALLOW" 를 required 일 때 BLOCK 으로 바꾸고 단일 `codex_audit` 기록도 읽게 한다 | 기존 훅·배선 재사용, 코드량 최소 | 기본 off 라 대부분 프로젝트에서 효과 없음. 메인 세션 turn-end 범위라 서브에이전트의 판정 시점과 어긋난다. 변경이 있는 모든 turn-end 에 발화해 감사와 무관한 작업까지 막을 수 있다 |
| B-4 메커니즘 없음 (문서만) | 에이전트 본문에 "required 이면 반드시 호출"만 강화 | 비용 없음 | 제보된 결함(자발적 호출에 의존)을 그대로 둔다. REQ-CAG-009 불충족 — 채택하면 REQ-CAG-008~011 을 삭제해야 한다 |

권장 기본안은 **B-1** 이다. 근거: REQ-CAG-009(런타임이 기록한 증거)를 이 트리에서 이미 확인된 표면(MCP 서버 상태 기록, run-phase 감사 게이트)만으로 충족한다. B-2 가 예방력은 더 강하지만 호출-서브에이전트 귀속이 관측되지 않은 전제에 기대므로, 선택하려면 run-phase 첫 단계에서 그 전제를 실측해야 한다. B-1 과 B-2 는 배타적이지 않다 — B-1 을 먼저 착지하고 B-2 를 후속 카드로 올리는 조합도 가능하다.

**이 결정은 권장일 뿐 확정이 아니다.** 레인이 리드/운영자에게 올려 Kickoff 에서 고른다.

### B.2 [NEEDS CLARIFICATION: 축 (c) 제보자 auth 형태의 확보 경로]

이슈 본문과 댓글 어디에도 제보자의 `auth.json` 형태나 `codex login status` 출력이 없다(spec.md §A.2 F9). 카드는 "제보자 구성 형태로 재현 먼저"를 요구하지만 그 형태는 현재 관측되지 않았다. 선택지:

- **C-1 (권장)**: 리드가 이슈에 형태(값 제외, 키 구조와 `codex login status` 한 줄)를 요청한다. 회신 전까지 축 (c)는 보류하고 (a)(b)만 진행한다.
- C-2: codex-cli 0.149.0 의 auth 저장 형식과 `login status` 출력 문자열을 릴리스 소스에서 실측해 후보 형태를 만들고, 각 후보의 characterization 테스트를 추가한다. 이때 테스트 이름과 보고서에 "제보자 관측이 아니라 0.149.0 소스 판독에서 유도한 형태"라고 명시한다.
- C-3: 축 (c)를 이 SPEC 에서 떼어 별도 카드로 넘긴다.

현 시점의 가설(관측 아님, run-phase 에서 확인 대상):
- H1: auth.json 이 없고(예: 자격증명을 파일 밖에 저장) `login status` 가 API key 로그인 시 `api key` 뒤에 다른 문구를 붙여 출력해, 줄 끝 앵커 문법에 걸리지 않는다.
- H2: `auth_mode` 값이 `chatgpt` / `apikey` 외의 토큰이다.
- H3: `CODEX_HOME` 해석 결과가 codex 가 실제로 쓰는 경로와 다르다(root 사용자 홈).

어느 가설도 관측되기 전에는 분류기를 고치지 않는다(REQ-CAG-012).

### B.3 축 (a) "차단"의 표현 — 결정 기록 (권장안 채택, 이의 시 Kickoff 에서 변경)

단일 MCP 도구 결과에서 "차단"을 무엇으로 표현할지 세 후보를 비교했다.

| 후보 | 평가 |
|------|------|
| **`verdict: fail` + `gate_unmet` 유지 + `isError: false` (채택)** | 두 감사자 본문은 `verdict` 로 판정하므로 fail 이 그대로 FAIL 판정으로 이어진다. 수렴 엔진의 overall=fail 선례(t580 운영자 결정 "명시적 required 는 fail-closed")와 같은 방향. `gate_unmet` 의 존재가 "리뷰 후 실패"와 "게이트 미충족"을 가른다 |
| `isError: true` | 기존 도구 에러 경로는 구조화 내용을 싣지 않고 조기 반환한다. 감사자는 이를 도구 고장으로 읽어 재시도하거나 무시할 수 있고, "fail-open 은 구조화 결과로 남는다"는 기존 계약 테스트와 충돌한다 |
| verdict 는 inconclusive 로 두고 새 필드(예: blocking=true) 추가 | 제보된 결함 그 자체 — 새 필드를 읽지 않는 소비자에게는 여전히 통과로 보인다 |

F4·F5 에 따라 수렴 엔진과 codex Stop-hook 은 이 핸들러를 거치지 않으므로 채택안은 두 감사자 에이전트에만 닿는다(REQ-CAG-006 으로 고정).

## §C Pre-flight

- 워크트리 `.claude/worktrees/t686`, 브랜치 `WT-codex-audit-gate`. 모든 git 은 `git -C <worktree>`.
- 변경 전 기준선: `go test ./internal/cli/ -run 'CodexAudit|RequiredGate|MultiReviewGate|CodexReviewGate|CodexAuth' -count=1` 를 run-phase 첫 단계에서 실행해 초록 기준선을 기록한다(이 plan 세션은 테스트를 실행하지 않았다).
- 전체 스위트는 로컬에서 돌리지 않는다 — 영향 패키지만, 전체 판정은 CI.

## §D Constraints

spec.md §C 참조. 추가로:
- 템플릿 파일을 고치면 `make build` 전에 `internal/template/templates/.claude/agents/moai/*.md` 수정분에 대해 `make agents-emit` 으로 codex 에이전트 방출본을 재생성한다(방출본 손편집 금지).
- 템플릿 본문에 SPEC ID·카드 ID·날짜·이슈 번호를 넣지 않는다.

## §E Self-Verification (run-phase 가 채울 항목)

- E1 AC 매트릭스 PASS/FAIL (acceptance.md)
- E2 `go vet ./internal/cli/... ./internal/runtime/...`
- E3 영향 패키지 테스트 + 커버리지
- E4 `golangci-lint run ./internal/cli/... ./internal/runtime/...`
- E5 `make agents-emit-check` 및 템플릿 중립성 점검
- E6 미푸시 상태 기록(레인은 push 하지 않는다)

## §F Milestones (우선순위 순)

### M1 — 축 (b) 결정 반영 설계 고정 · Priority High

- §B.1 에서 확정된 옵션에 맞춰 증거 저장소의 위치·레코드 필드·수용 지점을 고정한다.
- B-2 가 선택되면 먼저 훅 입력에 서브에이전트 식별자가 오는지 실측하고, 오지 않으면 운영자에게 되돌린다.
- 산출: progress.md §E.2 에 설계 확정 기록.

### M2 — 축 (a) RED → GREEN · Priority High

- RED: required+inconclusive 가 `verdict: fail` 을 내야 한다는 테스트(바이너리 부재, RPC 실패 두 원인). 기존 `codex_audit_gate_unmet_test.go` 의 "verdict 는 inconclusive 유지" 단언은 새 계약으로 뒤집힌다 — 이 단언 변경을 커밋 메시지에 명시.
- 역방향 회귀: off / advisory / 키 부재 / 설정 파일 손상 각각에서 직렬화 결과가 변경 전과 바이트 동일함을 고정(골든 비교).
- GREEN: 단일 핸들러의 주석 함수가 required 일 때 verdict 를 fail 로 올리고 summary 에 원인을 병기.
- 수렴 엔진·codex Stop-hook 기존 테스트가 무변경으로 통과함을 확인.

### M3 — 축 (b) 구현 · Priority High

- 선택 옵션 구현. B-1 이면: 서버 측 영수증 기록 → 결과에 영수증 id → 수용 지점 검사(required 일 때만) → 저장소 판독 실패의 gap 보고.
- 에이전트 본문(로컬 + 템플릿 미러)에 영수증 인용 의무 추가.

### M4 — 축 (c) 재현 우선 · Priority Medium (§B.2 확정 후)

- 관측된 형태로 characterization 테스트 작성, 변경 전 분류기에서 `unknown` 을 내며 실패함을 확인.
- 최소 수리 후 통과 확인. 판독 불가·충돌·자격증명 없음은 `unknown` 유지 테스트 유지/추가.
- C-1 회신이 없으면 이 마일스톤은 보류로 기록하고 SPEC 은 (a)(b)만으로 sync 할 수 있다.

### M5 — 문서·미러·빌드 · Priority Low (기계적)

- `codex_audit` 도구 설명 문구 갱신.
- plan-auditor / sync-auditor 본문 문장 갱신(로컬 + 템플릿 미러), `make agents-emit`, `make build`.

## §G Anti-Patterns

- 배포 기본값(codex required)을 opt-in 으로 읽어 모든 기존 프로젝트를 fail-closed 로 뒤집기.
- 제보자 auth 형태를 관측하지 않은 채 가설로 분류기를 고치고 "재현했다"고 보고하기.
- 에이전트가 쓴 텍스트(보고서 문구)만으로 호출 증거를 판정하기.
- 축 (b) 구현 중 수렴 엔진이나 codex Stop-hook 을 "김에" 고치기.

## §H Cross-References

### 템플릿 미러 쌍 (M3·M5 에서 닿을 수 있음)

| 로컬 | 템플릿 |
|------|--------|
| `.claude/agents/moai/plan-auditor.md` | `internal/template/templates/.claude/agents/moai/plan-auditor.md` |
| `.claude/agents/moai/sync-auditor.md` | `internal/template/templates/.claude/agents/moai/sync-auditor.md` |

두 템플릿 파일을 고치면 `make agents-emit` 로 `internal/template/templates/.codex/agents/moai/{plan-auditor,sync-auditor}.toml` 을 재생성하고 `make build` 한다. `workflow.yaml` 템플릿에는 `audit.gates` 블록이 없으므로(현재 트리 실측) 게이트 설명 추가가 필요할 때만 `internal/template/templates/.moai/config/sections/workflow.yaml` 를 고친다 — 필요 여부는 M5 에서 판단.

### run-phase 가 닿을 파일 (예상 목록)

- 축 (a): `internal/cli/mcp_codex.go`, `internal/cli/codex_audit_gate_unmet_test.go`, `internal/cli/mcp_server.go`(도구 설명)
- 축 (b) B-1 기준: `internal/cli/mcp_codex.go`, `internal/cli/mcp_convergence.go`(영수증 기록만, 수렴 판정 무변경), 신규 영수증 저장소 파일과 테스트(`internal/cli/` 내), `internal/runtime/audit_gate.go` 및 테스트(수용 지점 검사), 에이전트 본문 미러 4파일 + 방출본 2파일
- 축 (c): `internal/cli/mcp_codex.go`, `internal/cli/codex_auth_ladder_test.go`

### 선례

- `.moai/specs/SPEC-AUDIT-GATE-INTEGRITY-001` — "게이트가 실제로 게이트하지 않는다" 병리, BLOCKING 의 판정 배선
- `.moai/specs/SPEC-WF-AUDIT-GATE-001` — run-phase 감사 게이트(INCONCLUSIVE 는 PASS 가 아니다)
- `.moai/specs/SPEC-CODEX-REVIEW-TARGET-001` — #1632 #0 수리, 이 SPEC 의 직전 형제
