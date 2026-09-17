# SPEC-CODEX-AUDIT-GATE-AXES-001 — 구현 계획

카드 **t686** · Tier **M** · 클래스 **C** · 기준 트리 develop @ `f67d2193f` · v0.3.1

마일스톤은 뒤집힐 가능성이 큰 결정부터 적었다. 축 (c)는 카드 t870 으로 분리되어 이 계획에 없다.

## §A Context

spec.md §A 참조. 축 (a)와 (b)는 독립이다. M2·M3(골든 + 축 a)는 M1 결과와 무관하게 먼저 진행할 수 있다.

## §B 결정 기록

모든 결정이 확정되었다. 남은 것은 M1 의 실측 정지 규칙(§F M1)뿐이다.

### B.1 축 (b) 집행 방식 — 결정: B-1 서버 발급 영수증 (운영자, 2026-09-18)

required 게이트가 명시된 프로젝트에서는 감사 호출 증거가 없는 PASS 가 수용될 수 없다. 근거는 제보자 관측(건너뛴 3회 PASS 0.92 → 강제 호출 후 FAIL 0.79).

기존 opt-in Stop-hook(`multi_review_gate.go:44-79`)과의 관계: 이 훅은 가장 최근 `audit_multi` 수렴 결과를 읽어 required FAIL 일 때만 막고 상태 파일이 없으면 ALLOW 한다. 감사를 부르지 않은 세션은 상태 파일이 없어 통과한다. 기본 off, 메인 세션 범위, `audit_multi` 전용이다. 이 SPEC 은 그 훅을 바꾸지 않고, "감사를 불렀는가"를 검사하는 별도 경로를 추가한다.

기록용 선택지: B-2 SubagentStop 단독 차단(→ S1 의 실행 장소로 흡수), B-3 다중 리뷰 Stop-hook 확장(기본 off·메인 세션 범위로 미채택), B-4 문서만(결함 잔존으로 미채택).

### B.2 축 (c) — 카드 t870 으로 분리 (운영자, 2026-09-18)

제보자 auth 형태 회신 대기. 이 SPEC 의 요구사항·AC·마일스톤·파일 목록에서 제거했다.

### B.3 축 (a) 차단 표현 — 결정: `verdict: fail` + `gate_unmet` + `isError: false`

| 후보 | 평가 |
|------|------|
| **`verdict: fail` + `gate_unmet` + `isError: false` (채택)** | 감사자는 `verdict` 로 판정. 수렴 엔진 overall=fail 선례(t580)와 같은 방향. `gate_unmet` 존재가 "리뷰 후 실패"와 "게이트 미충족"을 가른다 |
| `isError: true` | 도구 에러 경로는 구조화 내용 없이 조기 반환 — 도구 고장으로 읽힐 수 있고 fail-open 구조화 결과 계약과 충돌 |
| inconclusive + 새 필드 | 새 필드를 읽지 않는 소비자에게 여전히 통과로 보인다 |

### B.4 검사 표면 — 결정: S1 SubagentStop + PreToolUse 소비자 (운영자 K1·N2, 2026-09-18)

실측(spec.md §A.2 F8-F11):
- run-phase plan-audit 런타임 게이트는 호출자가 없다(`internal/runtime/audit_gate.go:199`, :307).
- `sync-phase-quality-gate.sh` 는 vet/build 를 기본 차단하지만(:14-19) 메인 세션 Stop 훅이고 감사 판정을 보지 않는다.
- SubagentStop 은 배선·차단 출력 가능(`.claude/settings.json:201-207`, `internal/hook/subagent_stop.go:38`, `internal/hook/types.go:369-372`).
- PreToolUse 는 `Agent|Task` 매처로 배선되어 있고 핸들러가 스폰 시점에 deny 를 돌려줄 수 있다(`internal/hook/pre_tool.go:632-640`, 가드 선례 `internal/hook/agent_model_guard.go:237`).

흐름:
1. **SubagentStart**(plan-auditor / sync-auditor): 시작 표식 기록(§B.5).
2. **MCP 서버**: `codex_audit` / `audit_multi`(codex 참여) 호출마다 영수증 기록.
3. **SubagentStop**: 최종 메시지의 판정 줄(§B.6)을 읽고 영수증 검사.
   - PASS 이고 검사 실패 + `stop_hook_active: false` → `decision: "block"` + `rejections/<agent_type>--<spec_id>.json` 영속.
   - 해석 가능한 판정 줄 없음(누락·형식 불일치) + `stop_hook_active: false` → `decision: "block"` + `rejections/<agent_type>--unknown-spec.json` 영속. 판정 줄을 빼는 것으로 검사를 우회할 수 없다.
   - 위 두 조건 + `stop_hook_active: true` → 차단 없음, 거부 기록 유지, `systemMessage` 경고.
   - PASS 이고 검사 통과 → **같은 역할(`agent_type`)의 모든 거부 기록 제거** — 다른 SPEC 의 기록과 `unknown-spec` 기록 포함(운영자 결정 K4). 사람이 기록 파일을 직접 지우는 것도 허용된다.
   - FAIL → 검사 없음, 기록 변경 없음.
4. **PreToolUse `Agent|Task`**(기존 스폰 가드 경로, `pre_tool.go:632` 블록에 형제로 추가): 대상이 `manager-develop`(run 진입) / `manager-docs`(sync 진입) / `manager-git`(sync 후 PR) 이고 트리에 거부 기록이 하나라도 있으면 `AUDIT_RECEIPT_VIOLATION` 으로 deny.

에이전트 모델 가드와 달리 이 deny 에는 별도 opt-in 플래그가 없다. 활성 조건은 raw `workflow.audit.gates.codex == required` 자체다 — 그 값을 쓴 것이 곧 opt-in 이다.

선택하지 않은 표면: S2 PreToolUse 보고서 Write 거부(sync-auditor 는 Write 가 없어 sync 미포괄), S3 CLI 검증 동사(run 진입에 호출할 기계 지점 없음).

### B.5 저장소, 스키마, 상관, 트리 판정

**트리 루트(canonical tree root)**
- MCP 쪽: `project_root` 인자를 `EvalSymlinks` 로 정규화한 값(`mcp_project_root.go:179`). 인자가 없으면 서버 폴백 해석(`mcp_project_root.go:100-104`) 결과를 같은 방식으로 정규화하고, 영수증에 `root_source` 로 폴백 여부를 기록한다.
- 훅 쪽: 훅 입력 `cwd`(`types.go:212`) 에서 `git -C <cwd> rev-parse --show-toplevel` 결과를 `EvalSymlinks` 로 정규화. git 실패 시 정규화한 `cwd`. `CLAUDE_PROJECT_DIR` 는 쓰지 않는다 — 워크트리 세션에서도 primary 체크아웃을 가리키기 때문이다.
- "다른 트리" 판정: 영수증의 `tree_root` 와 감사자 시작 표식의 `tree_root` 가 바이트 동일하지 않으면 다른 트리다. 따라서 `project_root` 없이 호출돼 primary 체크아웃으로 폴백한 영수증은 워크트리 감사자에게 "다른 트리"로 거부된다(의도된 동작).

**저장 위치** — 모두 `<tree_root>/.moai/state/audit-receipts/` 아래:

| 레코드 | 경로 | 필드 |
|--------|------|------|
| 영수증 | `receipts/<receipt_id>.json` | `receipt_id`, `tool`(`codex_audit`\|`audit_multi`), `tree_root`, `root_source`(`argument`\|`fallback`), `created_at`(RFC 3339 UTC, 나노초), `codex_verdict`, `gate_unmet` |
| 시작 표식 | `starts/<agent_id>.json` | `agent_id`, `agent_type`, `session_id`, `tree_root`, `started_at`(RFC 3339 UTC, 나노초) |
| 거부 기록 | `rejections/<agent_type>--<spec_id>.json` | `agent_type`, `spec_id`(판정 줄 해석 불가 시 `unknown-spec`), `agent_id`, `cause`, `cited_receipts`, `rejected_at`, `reentry_warned`(bool) |

`receipt_id` 는 서버가 생성하는 `rcpt-` 접두 불투명 토큰(`rcpt-[a-z0-9]{20,40}`)이다. 쓰기는 임시 파일 후 원자 rename.

**상관**: SubagentStart 와 SubagentStop 은 같은 `agent_id` 로 같은 감사자 인스턴스를 잇는다. Stop 은 `starts/<agent_id>.json` 을 읽는다. 표식이 없으면 검사 실패 원인 "start marker missing" 이다. Stop 처리 후 표식 파일은 제거한다.

**시험 격리**: 저장소 루트는 패키지 수준 seam 으로 교체 가능해야 하며, 모든 신규 테스트는 `t.TempDir()` 트리를 `project_root`/`cwd` 로 넘기고 seam 을 `t.Cleanup` 으로 복원한다. `project_root` 없이 도구를 부르는 기존 테스트가 실제 저장소 트리에 쓰지 않도록, 폴백 루트 해석도 테스트에서 임시 디렉터리로 고정한다(AC-CAG-013 에서 누출 검사).

**읽기 전용 주석과의 관계**: `codex_audit` 의 읽기 전용 힌트(`mcp_server.go:278`)와 `WriteCapable: false`(`internal/mcp/catalog.go:51`)는 유지한다. 영수증은 사용자 작업 트리 내용이 아니라 런타임 상태이며, `audit_multi` 가 같은 분류에서 이미 수렴 결과를 `.moai/state` 에 영속한다(`mcp_convergence.go:749`)는 선례를 따른다.

### B.6 판정 줄 문법 (감사자 최종 메시지)

감사자 본문은 최종 응답의 **마지막 비어 있지 않은 줄**을 다음 형식으로 낸다.

```
AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL> spec=<SPEC-ID> receipts=<receipt_id>[,<receipt_id>...]
```

- 정규식: `^AUDIT-VERDICT: (PASS|PASS-WITH-DEBT|FAIL) spec=(SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}) receipts=(none|rcpt-[a-z0-9]{20,40}(,rcpt-[a-z0-9]{20,40})*)[ \t\r]*$`
- 후보 줄은 앞쪽 공백을 제거하고, 끝의 공백·탭·`\r` 은 정규식이 허용한다(CRLF 메시지에서도 유효한 줄이 누락으로 읽히지 않게).
- `PASS` 와 `PASS-WITH-DEBT` 는 PASS 부류로 검사한다. `receipts=none` 은 "인용 없음"이다.
- 검사 통과 조건(PASS 부류): 인용된 영수증 중 **하나 이상**이 (1) 저장소에 있고, (2) `tree_root` 가 시작 표식과 같고, (3) `created_at` 이 시작 표식 `started_at` 이후이며, (4) `tool` 이 `codex_audit` 이거나 codex 가 참여한 `audit_multi` 다. 모두 실패하면 원인은 다음 순서에서 처음 걸린 것으로 보고한다: 시작 표식 없음(조건 (2)(3)의 전제) → 인용 없음 → 저장소에 없음 → 다른 트리 → 시작 이전.
- 판정 줄이 없거나 형식이 맞지 않으면 required 트리에서 "verdict line missing" 이며, REQ-CAG-011 에 따라 첫 종료를 막고 `unknown-spec` 거부 기록을 남긴다.

## §C Pre-flight

- 워크트리 `.claude/worktrees/t686`, 브랜치 `WT-codex-audit-gate`. git 은 `git -C <worktree>`.
- 변경 전 기준선: run-phase 첫 단계에서 `go test ./internal/cli/ -run 'CodexAudit_|CodexBlankReview_|Converge_|RunMultiAudit_|AuditMulti_|ReviewGate_|MultiReviewGate' -count=1` 과 `go test ./internal/hook/ -run 'SubagentStart|SubagentStop|PreTool|AgentModel' -count=1` 초록을 기록(plan 세션은 테스트 미실행).
- [HARD] 로컬 검증은 `-run` 선택자를 붙인 패키지 단위 실행만 한다. `internal/cli`·`internal/hook` 전체 스위트는 로컬에서 돌리지 않으며, 전체 스위트 판정은 리드 push 후 CI 다. 모든 `-run` 실행은 스윕 수가 0 이 아님을 `-v` 출력의 `=== RUN` 줄 수로 확인한다.

## §D Constraints

spec.md §C 참조. 템플릿 에이전트 본문을 고치면 `make agents-emit` 후 `make build`. 방출본 손편집 금지. 영수증 쓰기와 읽기 전용 주석의 관계는 §B.5.

## §E Self-Verification (run-phase 가 채울 항목)

E1 AC 매트릭스, E2 `go vet ./internal/cli/... ./internal/hook/...`, E3 영향 패키지의 `-run` 선택 테스트+커버리지(전체 스위트는 CI), E4 `golangci-lint run ./internal/cli/... ./internal/hook/...`, E5 `make agents-emit-check`·템플릿 중립성, E6 미푸시 기록.

## §F Milestones (우선순위 순)

### M1 — 훅 페이로드 실측과 정지 규칙 · Priority High

- 실제 Claude Code 세션에서 plan-auditor 를 한 번 띄워 SubagentStart·SubagentStop 페이로드를 캡처하고(비밀값 제외) progress.md §E.2 에 필드 존재를 기록한다: `agent_type`, `agent_id`(양쪽), `cwd`, `last_assistant_message`(Stop), `stop_hook_active`(Stop).
- **정지 규칙 [HARD]**: 관측된 SubagentStart 또는 SubagentStop 페이로드에 `agent_type`, `last_assistant_message`, 또는 두 이벤트를 잇는 `agent_id` 중 하나라도 없으면, 축 (b) 구현을 시작하지 않고 S2+S3 조합안으로 되돌린다는 보고를 리드에게 먼저 올린다. 축 (a)(M2·M3)는 계속할 수 있다.

### M2 — 골든 캡처 (구현 전 커밋) · Priority High

- 변경 전 코드에서 비-required 5경우(off / advisory / 키 부재 / 설정 파일 부재 / YAML 손상) × 바이너리 부재의 `codex_audit` 직렬화 결과를 캡처. `build_commit`·`build_lag`(`mcp_codex.go:1672`, :1694)는 고정 자리표시자로 정규화.
- 골든 파일과 비교 테스트를 축 (a) 구현 커밋보다 앞선 별도 커밋으로 올린다.

### M3 — 축 (a) RED → GREEN · Priority High

- RED: required + {바이너리 부재, RPC 실패, 빈 출력} → `verdict: fail`.
- 뒤집히는 단언(커밋 메시지에 명시): `codex_audit_gate_unmet_test.go:32` `TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive`, `codex_blank_review_test.go:408-420` `TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput`(주석 단언 유지, verdict 단언만 `fail`).
- GREEN: 단일 핸들러가 required 일 때 verdict 를 fail 로 올리고 summary 에 원인 병기.

### M4 — 축 (b) 구현 · Priority High (M1 통과 후)

- 영수증 저장소(쓰기·읽기·seam), MCP 서버 영수증 기록과 required 한정 필드.
- SubagentStart 시작 표식, SubagentStop 검사·거부 영속·재진입 경고·수용 시 제거.
- PreToolUse `Agent|Task` 가드: `manager-develop` / `manager-docs` / `manager-git` 스폰 deny.
- 에이전트 본문(로컬+템플릿)에 §B.6 판정 줄 의무 추가.

### M5 — 문서·미러·빌드 · Priority Low

- `codex_audit` 도구 설명(`mcp_server.go:269`), 감사자 본문, cross-model-audit 스킬(`SKILL.md:243-251`, "handlers the convergence engine reuses" 문구 정정), 미러, `make agents-emit`, `make build`.

## §G Anti-Patterns

- 배포 기본값(codex required)을 opt-in 으로 읽기.
- 에이전트가 쓴 텍스트만으로 영수증 존재를 판정하기(저장소 대조 필수).
- 훅에서 `CLAUDE_PROJECT_DIR` 로 트리를 판정하기(워크트리에서 primary 를 가리킨다).
- 골든을 구현 후 캡처하기.
- M1 페이로드 실측 없이 축 (b)를 구현하기.
- 수렴 판정이나 codex Stop-hook 을 "김에" 고치기.

## §H Cross-References

### 템플릿 미러 쌍

| 로컬 | 템플릿 |
|------|--------|
| `.claude/agents/moai/plan-auditor.md` | `internal/template/templates/.claude/agents/moai/plan-auditor.md` |
| `.claude/agents/moai/sync-auditor.md` | `internal/template/templates/.claude/agents/moai/sync-auditor.md` |
| `.claude/skills/moai-ref-cross-model-audit/SKILL.md` | `internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md` |

감사자 템플릿 본문 수정 시 `make agents-emit` 로 `internal/template/templates/.codex/agents/moai/{plan-auditor,sync-auditor}.toml` 재생성 → `make build`. 훅 배선(settings.json)은 SubagentStart·SubagentStop·PreToolUse `Agent|Task` 가 이미 있으므로 바꾸지 않는다.

### run-phase 가 닿을 파일 (예상)

- 축 (a): `internal/cli/mcp_codex.go`, `internal/cli/mcp_server.go`(도구 설명), `internal/cli/codex_audit_gate_unmet_test.go`, `internal/cli/codex_blank_review_test.go`, 신규 골든+비교 테스트(`internal/cli/testdata/` 아래)
- 축 (b): `internal/cli/mcp_codex.go`, `internal/cli/mcp_convergence.go`(영수증 필드만), 신규 영수증 저장소 패키지 파일+테스트(MCP 서버와 훅이 함께 쓰므로 `internal/` 아래 공용 위치), `internal/hook/subagent_start.go`, `internal/hook/subagent_stop.go`, `internal/hook/pre_tool.go`, 신규 가드 파일+테스트(`internal/hook/`), 에이전트 본문 미러 4파일 + 방출본 2파일
- 문서: cross-model-audit 스킬 미러 2파일

### 선례와 덮어쓰는 계약

- SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 REQ-CBR-009 (spec.md:163-165), Out of Scope (:221-222) — spec.md §A.4
- SPEC-CODEX-REVIEW-TARGET-001 spec.md:207 — spec.md §A.4
- SPEC-AUDIT-GATE-INTEGRITY-001, SPEC-WF-AUDIT-GATE-001
