# 부록 3 — 인터페이스/명령 대응표 (DR-07 #2143)

출처: synthesis S11 + `src/main.rs` OPERATIONS(전수 — tip `604792666` 기준 **47개**, 아래 목록이 전부다). 문서가 그린 인터페이스명(`result.admit` 등)은 RC1에 없고, RC1 OPERATIONS는 문서 어디에도 대응 표가 없었다 — 이 표가 그 대응표다.

## 문서 인터페이스명 → RC1 OPERATIONS

| 문서 인터페이스명 (출처) | RC1 OPERATIONS | 대응 관계 |
| --- | --- | --- |
| `result.admit` ([43fa] 등) | `plan.admit` (`plan-admit`) | admission 판정 — RC1은 `ready/stale/blocked`+사유 코드로 응답 |
| `evidence.admit` | `plan.admit` + `run-verify` receipt | evidence 연결은 plan pin(`PinnedContract`) + verify receipt로 분리 |
| `verify.accept` | `run-verify` (`run.verify`) | verify receipt(6종 상태) 반환 |
| `owner.handoff` | `claim-handoff` (`claim.handoff`) | handoff 기록·인계 |
| `decision.propose` | `decision-record` (`decision.record`) | 결정 원장 append |
| `sync.*` | 없음 (RC1 미출하 — SYNC는 workflow 문서 절차) | `skills/moai/workflows/sync.md` 절차만; 상태 기계는 부록 1 |
| `artifact.pin` | `plan.admit` `PinnedContract` + `update` rollback_manifest | artifact pin은 plan pin/manifest로 실현 |
| `dispatch.*` | `forge-read`/`forge-report` + `effect-intend`/`effect-confirm`/`effects-reconcile` | dispatch는 forge 읽기·outbox effect 3-op로 분리 실현 |
| — (문서 인터페이스명 없음) | `codex-projection-plan` (`codex.projection_plan`) | 대응 없음 — [83f] §6 host adapter projection을 측정하는 op이며 문서 인터페이스 어휘가 이를 이름으로 부르지 않는다 |
| — (문서 인터페이스명 없음) | `skill-source-plan` / `skill-source-apply` (`skill_source.plan`·`skill_source.apply`) / `skill-read-receipt` (`skill_source.read_receipt`) | 대응 없음 — 03-B #1996 신설(license review·skill 설치·read receipt); 문서 인터페이스 어휘와 [43fa] §18 명령표 어느 쪽에도 없는 영역 |

## RC1 OPERATIONS 전수 (tip `604792666`, 47개)

`package-inspect` `project-plan` `project-apply` `package-resolve` `agents-plan` `agents-apply` `apply-recover` `baseline-inventory` `identity` `forge-read` `forge-report` `plan-admit` `claude-doctor` `codex-projection-plan` `codex-doctor` `claim-acquire` `claim-show` `claim-release` `claim-handoff` `effect-intend` `effect-confirm` `effects-reconcile` `todo-promote-preview` `todo-promote-apply` `run-verify` `snapshot-describe` `skill-source-plan` `generation-plan` `generation-apply` `skill-source-apply` `skill-read-receipt` `context-bundle` `context-verify` `decision-record` `audit-reduce` `install-inspect` `worktree-plan` `worktree-apply` `scope-check` `pr-body-render` `github-plan` `github-apply` `update-plan` `update-apply` `recovery-plan` `recovery-apply` `pr-confirm`

(추출: `git show 604792666:src/main.rs`의 `const OPERATIONS` 블록 파싱 — `_`를 포함한 operation명(`codex.projection_plan`·`skill_source.*`)이 누락되지 않도록 `[a-z._]`로 매칭. envelope 계약: `{protocol_version, operation, ok, result|error{kind, message, retryable, next_actions?}}`, exit 0/1/2, `kind` 14종 — `src/contract.rs`·`docs/protocol-v1.md`)

## 문서별 내부 명령 체계 → RC1

| 문서 | 문서의 명령 체계 | RC1 대응 |
| --- | --- | --- |
| [25f] | `setup` / `project init` / `maintenance update` / `doctor` / `capabilities` | `project-plan/apply`(setup) · `update-plan/apply`(maintenance update) · `claude-doctor`·`codex-doctor`(doctor) · `capabilities`는 **RC1에 없음**(S11: "공통 명령은 capabilities뿐인데 RC1에 없다") |
| [e08] | `spec plan` / `task status` / `skill show` | `agents-plan` + `plan-admit`(spec plan) · `claim-show`(task status) · skill show 없음(`install-inspect`·`context-bundle`이 부분 대체) |
| [3d9] | `task submit/cancel` / `spec verify` / `artifacts list` | `claim-acquire`/`claim-release`(task) · `run-verify`(spec verify) · artifacts list 없음 |
| [43fa] §18 | 문서 명령 전부 | RC1 OPERATIONS 47개로 교체(S11 결정) |

교체 규칙(S11 결정): 문서의 명령 표는 `src/main.rs` OPERATIONS 표로 교체하거나 "RC1 미출하 계약안" 배너를 단다. 번호·이름으로 누락 검사를 하지 않는다(S8 — 같은 F 번호가 다른 기능).
