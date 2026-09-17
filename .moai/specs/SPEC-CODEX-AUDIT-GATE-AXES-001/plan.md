# SPEC-CODEX-AUDIT-GATE-AXES-001 — 구현 계획

카드 **t686** · Tier **M** · 클래스 **C** · 기준 트리 develop @ `f67d2193f` · v0.2.0

마일스톤은 뒤집힐 가능성이 큰 결정부터 적었다.

## §A Context

spec.md §A 참조. 세 축은 서로 독립이며, 축 (c)는 입력 대기라 (a)(b)만으로 run·sync 가 가능하다.

## §B 결정 기록

### B.1 축 (b) 집행 방식 — 결정됨: B-1 서버 발급 영수증 (운영자, 2026-09-18)

질문이었던 것: 감사 도구 호출 증거가 없는 판정은 PASS 가 될 수 없어야 하는가, 어디서 막는가.

**답: required 게이트가 명시된 프로젝트에서는 PASS 가 될 수 없다.** 제보자의 관측(건너뛴 3회 PASS 0.92 → 강제 호출 후 FAIL 0.79)이 이유다.

**기존 opt-in Stop-hook(`multi_review_gate.go`)과의 관계.** 이 훅은 세션의 가장 최근 `audit_multi` 수렴 결과를 읽어 required FAIL 일 때만 막고, 상태 파일이 없으면 ALLOW 한다(`multi_review_gate.go:44-79`). 감사 도구를 부르지 않은 세션은 상태 파일이 없으므로 이 분기로 통과한다. 기본 off, 메인 세션 turn-end, `audit_multi` 전용이라 단일 `codex_audit` 는 보지 못한다. 즉 "부른 감사가 실패했는가"만 보고 "감사를 불렀는가"는 보지 않는다.

검토했던 선택지(기록용):

| 옵션 | 요지 | 처분 |
|------|------|------|
| **B-1 서버 발급 영수증** | MCP 서버가 호출마다 영수증을 기록, 감사 보고가 인용, required 일 때 서버가 모르는 영수증의 PASS 거부 | **채택.** 증거는 런타임이 쓴다. 단, 검사가 실제로 돌 기계 표면이 필요하다 — §B.4 |
| B-2 SubagentStop 훅이 호출 기록 부재 시 종료 차단 | 훅 입력에 `agent_id`·`agent_transcript_path`·`last_assistant_message` 가 선언돼 있다(`internal/hook/types.go:238-241`, 런타임 페이로드 미관측). 전사 파일은 에이전트 텍스트가 아닌 런타임 기록이라 호출-귀속 없이도 증거원이 된다 | 독립 옵션으로는 미채택. 다만 §B.4 의 권장 표면이 이 훅 지점을 B-1 의 검사 실행 장소로 쓴다 |
| B-3 다중 리뷰 Stop-hook 확장 | 상태 없음 → required 일 때 BLOCK | 미채택: 기본 off, 메인 세션 범위, 단일 도구 미포괄 |
| B-4 문서만 | 본문 강화 | 미채택: 제보된 결함을 남김 |

### B.2 축 (c) 입력 확보 — 결정됨: C-1 (운영자, 2026-09-18)

- 리드가 이슈에 제보자의 auth 형태(값 제외, 키 구조와 `codex login status` 한 줄)를 요청한다. 레인은 게시하지 않는다.
- 축 (c)는 **입력 대기**. M4 는 연기되며 Kickoff 차단 요소가 아니다.
- 회신이 없으면 SPEC 은 (a)+(b)로 sync 하고, (c)는 AC-CAG-013~014 를 기록된 gap 으로 닫는다.
- 가설(관측 아님, 회신 후 확인 대상): H1 auth.json 부재 + `login status` 줄이 줄 끝 앵커 문법(`mcp_codex.go:2024`)에 걸리지 않는 문구, H2 `auth_mode` 가 알려진 두 값 외의 토큰, H3 `CODEX_HOME` 해석 불일치.

### B.3 축 (a) 차단 표현 — 결정됨: `verdict: fail` + `gate_unmet` + `isError: false` (이의 없음)

| 후보 | 평가 |
|------|------|
| **`verdict: fail` + `gate_unmet` 유지 + `isError: false` (채택)** | 감사자는 `verdict` 로 판정한다. 수렴 엔진의 overall=fail 선례(t580)와 같은 방향. `gate_unmet` 존재가 "리뷰 후 실패"와 "게이트 미충족"을 가른다 |
| `isError: true` | 도구 에러 경로는 구조화 내용 없이 조기 반환한다. 감사자가 도구 고장으로 읽을 수 있고, "fail-open 은 구조화 결과" 계약과 충돌 |
| inconclusive 유지 + 새 필드 | 제보된 결함 그 자체 — 새 필드를 읽지 않는 소비자에게 여전히 통과로 보인다 |

### B.4 영수증 거부 검사가 실제로 실행되는 기계 표면 — **Kickoff 에서 확정할 유일한 결정**

운영자 문구는 "run·sync 게이트에서 거부"다. 그런데 트리를 재어 보면 그 게이트는 기계 지점이 아니다.

실측:
- run-phase plan-audit 게이트: `internal/runtime/audit_gate.go:199` `GateConfig.Invoke` 의 유일한 호출은 같은 파일 :307 `TeamModeInvoke` 의 자기 위임이다. 프로덕션 호출자 없음. 게이트는 오케스트레이터 산문이다.
- sync 게이트 `sync-phase-quality-gate.sh`: 헤더(:14-19)상 vet/build 실패는 **기본 차단**이고 `MOAI_SYNC_GATE_BLOCKING=0` 이 opt-out 이다(감사 보고서의 "=1 일 때만 차단"은 현재 트리와 다르다). 다만 메인 세션 Stop 훅이고, 검사 대상은 vet/build 이며 감사 판정이 아니다. plan 단계는 덮지 못한다.
- SubagentStop 훅: settings.json 에 배선돼 있고(`.claude/settings.json:201-207`, timeout 5), 핸들러 `internal/hook/subagent_stop.go:38` 이 `HookOutput` 을 돌려주며 그 출력은 최상위 `decision: "block"` 을 실을 수 있다(`internal/hook/types.go:369-372`). 입력에 `agent_type`(:230), `agent_id`·`agent_transcript_path`·`last_assistant_message`(:238-241)가 선언돼 있다. `agent_type` 은 이미 라우팅 원장이 에이전트 식별에 쓴다(`internal/hook/routing_ledger.go:232`). 현재 핸들러는 차단하지 않는다.
- PreToolUse 거부 틀: `DecisionDeny` 를 돌려주는 가드가 여럿 있다(`internal/hook/branch_guard.go:536`, `integration_lock_guard.go:102`, `agent_stop_guard.go:399`).
- 감사자 도구: sync-auditor 는 Write 가 없다(`.claude/agents/moai/sync-auditor.md:9`), plan-auditor 는 있다(`plan-auditor.md:7`).

후보 표면:

| 표면 | 무엇을 하는가 | 장점 | 단점 |
|------|--------------|------|------|
| **S1 SubagentStop 검사 (권장)** | plan-auditor / sync-auditor 종료 시 `last_assistant_message` 에서 판정과 인용 영수증 id 를 읽고 영수증 저장소와 대조. 조건 미충족이면 `decision: block` + 이유로 감사자를 계속 돌게 한다. `stop_hook_active` 로 재진입하면 다시 막지 않고 거부 기록을 저장소에 남기며 `systemMessage` 로 "영수증 없는 PASS — 수용 불가"를 띄운다 | 두 감사자 모두 포괄(Write 유무 무관). 판정이 오케스트레이터에 닿기 **전에** 돈다. 기존 배선·출력 필드 재사용. 로컬 파일 대조라 5초 timeout 안 | 입력 필드는 선언만 있고 런타임 페이로드 미관측 — M1 첫 작업으로 실측해야 한다. 훅 비활성 환경에서는 무력. 재진입 후에는 차단이 아니라 기록+경고라서, 오케스트레이터가 경고를 무시하면 뚫린다 |
| S2 PreToolUse 가드: plan-audit 보고서 Write 거부 | `.moai/reports/plan-audit/*.md` 에 PASS 보고서를 쓰려는데 영수증이 없으면 deny | 기존 deny 틀 재사용, 파일 경로로 대상이 분명 | sync-auditor 는 보고서를 쓰지 않으므로 sync 를 덮지 못한다. 보고서를 안 쓰고 판정만 돌려주면 우회 |
| S3 `moai` CLI 검증 동사 + 게이트 스크립트 호출 | 예: 영수증 검증 CLI 를 sync-phase 품질 게이트 스크립트와 run 진입 절차가 호출 | 사람도 수동 실행 가능, 표면이 명시적 | sync 게이트는 메인 세션 turn-end 에 돌아 판정 시점과 어긋나고, run 게이트에는 호출할 기계 지점이 없다(F10) — 결국 오케스트레이터 자발성에 기댄다 |

**권장: S1.** 운영자 문구의 "run·sync 게이트"는 plan-audit 판정(run 진입 전제)과 sync-audit 판정을 뜻하는 것으로 읽었고, S1 은 두 판정이 생성되는 순간에 검사한다. 이 해석이 운영자 문구에서 기계적으로 도출되지는 않으므로 **Kickoff 에서 확인받는다.** S1 이 M1 실측에서 성립하지 않으면(페이로드에 `agent_type` 또는 `last_assistant_message` 가 오지 않음) 레인은 구현하지 않고 S2+S3 조합안으로 되돌린다.

## §C Pre-flight

- 워크트리 `.claude/worktrees/t686`, 브랜치 `WT-codex-audit-gate`. git 은 `git -C <worktree>`.
- 변경 전 기준선: run-phase 첫 단계에서 `go test ./internal/cli/ -run 'CodexAudit_|CodexBlankReview_|Converge_|RunMultiAudit_|AuditMulti_|ReviewGate_|MultiReviewGate' -count=1` 초록을 기록한다(plan 세션은 테스트 미실행).
- 전체 스위트는 로컬에서 돌리지 않는다.

## §D Constraints

spec.md §C 참조. 템플릿 에이전트 본문을 고치면 `make agents-emit` 후 `make build`. 방출본 손편집 금지.

## §E Self-Verification (run-phase 가 채울 항목)

- E1 AC 매트릭스, E2 `go vet ./internal/cli/... ./internal/hook/...`, E3 영향 패키지 테스트+커버리지, E4 `golangci-lint run ./internal/cli/... ./internal/hook/...`, E5 `make agents-emit-check`·템플릿 중립성, E6 미푸시 기록.

## §F Milestones (우선순위 순)

### M1 — §B.4 표면 전제 실측 · Priority High

- 실제 Claude Code 세션에서 plan-auditor 를 한 번 띄워 SubagentStop 페이로드의 `agent_type`·`agent_id`·`agent_transcript_path`·`last_assistant_message` 존재를 관측하고 progress.md §E.2 에 원문(비밀값 제외)을 남긴다.
- 성립하지 않으면 멈추고 리드에게 되돌린다(S2+S3 조합안).

### M2 — 골든 캡처 (구현 전 커밋) · Priority High

- 변경 전 코드에서 비-required 5경우(off / advisory / 키 부재 / 설정 파일 부재 / YAML 손상) × 바이너리 부재 조건의 `codex_audit` 직렬화 결과를 캡처한다. `build_commit`·`build_lag`(`mcp_codex.go:1672`, :1694)는 고정 자리표시자로 정규화한다.
- 골든 파일과 비교 테스트를 **축 (a) 구현 커밋보다 앞선 별도 커밋**으로 올린다.

### M3 — 축 (a) RED → GREEN · Priority High

- RED: required + {바이너리 부재, RPC 실패, 빈 출력} → `verdict: fail`.
- 뒤집히는 단언 목록(커밋 메시지에 명시): `codex_audit_gate_unmet_test.go:32` 계열(`TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive`), `codex_blank_review_test.go:408-420`(`TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput`, :419-420 의 `verdict == inconclusive`). 후자는 SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 REQ-CBR-009 의 AC 이며, 주석 단언은 유지하고 verdict 단언만 `fail` 로 바꾼다.
- GREEN: 단일 핸들러가 required 일 때 verdict 를 fail 로 올리고 summary 에 원인 병기.

### M4 — 축 (b) 영수증 + S1 검사 · Priority High (§B.4 확정 후)

- 서버 측 영수증 기록(항상), 결과의 영수증 id 필드(required 일 때만), 수렴 결과 동일.
- SubagentStop 검사: required 트리 + 감사자 PASS 일 때만 동작, 네 가지 거부 조건과 판독 불가 gap.
- 에이전트 본문(로컬+템플릿)에 "PASS 보고 시 영수증 id 인용" 의무 추가.

### M5 — 축 (c) · 연기 (입력 대기, Kickoff 비차단)

- 회신 수령 시: 관측 형태로 RED characterization → 최소 수리 → unknown 유지 사례 재확인.
- 회신 없음: AC-CAG-013~014 를 gap 으로 기록하고 종료.

### M6 — 문서·미러·빌드 · Priority Low

- `codex_audit` 도구 설명(`mcp_server.go:269`), 감사자 본문, cross-model-audit 스킬(`SKILL.md:243-251` — "handlers the convergence engine reuses" 문구 정정 포함), 미러, `make agents-emit`, `make build`.

## §G Anti-Patterns

- 배포 기본값(codex required)을 opt-in 으로 읽기.
- 제보자 auth 형태를 관측하지 않은 채 가설로 분류기를 고치고 "재현했다"고 보고하기.
- 에이전트가 쓴 텍스트만으로 영수증 존재를 판정하기(저장소 대조 필수).
- 골든을 구현 후 캡처하기(동어반복).
- 수렴 판정이나 codex Stop-hook 을 "김에" 고치기.

## §H Cross-References

### 템플릿 미러 쌍

| 로컬 | 템플릿 |
|------|--------|
| `.claude/agents/moai/plan-auditor.md` | `internal/template/templates/.claude/agents/moai/plan-auditor.md` |
| `.claude/agents/moai/sync-auditor.md` | `internal/template/templates/.claude/agents/moai/sync-auditor.md` |
| `.claude/skills/moai-ref-cross-model-audit/SKILL.md` | `internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md` |

감사자 템플릿 본문 수정 시 `make agents-emit` 로 `internal/template/templates/.codex/agents/moai/{plan-auditor,sync-auditor}.toml` 재생성 → `make build`.

### run-phase 가 닿을 파일 (예상)

- 축 (a): `internal/cli/mcp_codex.go`, `internal/cli/mcp_server.go`(도구 설명), `internal/cli/codex_audit_gate_unmet_test.go`, `internal/cli/codex_blank_review_test.go`, 신규 골든 파일+비교 테스트(`internal/cli/testdata/` 아래)
- 축 (b): `internal/cli/mcp_codex.go`, `internal/cli/mcp_convergence.go`(영수증 필드만), 신규 영수증 저장소 파일+테스트(`internal/cli/`), `internal/hook/subagent_stop.go`+테스트, 필요 시 저장소 판독 공용 패키지; 에이전트 본문 미러 4파일 + 방출본 2파일
- 축 (c, 연기): `internal/cli/mcp_codex.go`, `internal/cli/codex_auth_ladder_test.go`
- 문서: cross-model-audit 스킬 미러 2파일

### 선례와 덮어쓰는 계약

- SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 REQ-CBR-009 (spec.md:163-165), Out of Scope (:221-222) — spec.md §A.4
- SPEC-CODEX-REVIEW-TARGET-001 spec.md:207 — spec.md §A.4
- SPEC-AUDIT-GATE-INTEGRITY-001 — "게이트가 실제로 게이트하지 않는다" 병리
- SPEC-WF-AUDIT-GATE-001 — INCONCLUSIVE 는 PASS 가 아니다
