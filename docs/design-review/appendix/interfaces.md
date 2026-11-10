# 부록 3 — 인터페이스/명령 대응표 (DR-07 #2143)

출처: synthesis S11 + `src/main.rs` OPERATIONS(전수는 아래 "RC1 OPERATIONS 전수" 절의 목록이 전부다. 이름 집합은 `src/main.rs`의 시험 `operations_doc_matches_code`가 코드 배열과 대조한다). 문서가 그린 인터페이스명(`result.admit` 등)은 RC1에 없고, RC1 OPERATIONS는 문서 어디에도 대응 표가 없었다 — 이 표가 그 대응표다.

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

## RC1 OPERATIONS 전수

- `agents-apply`
- `agents-plan`
- `apply-recover`
- `archive-apply`
- `archive-plan`
- `archive-restore`
- `archive-search`
- `atlas-tag-scan`
- `audit-reduce`
- `baseline-inventory`
- `claim-acquire`
- `claim-handoff`
- `claim-release`
- `claim-show`
- `claim-verify`
- `claude-doctor`
- `claude-settings-apply`
- `claude-settings-plan`
- `codex-doctor`
- `codex-projection-apply`
- `codex-projection-plan`
- `context-bundle`
- `context-verify`
- `cost-report`
- `decision-append`
- `decision-record`
- `effect-confirm`
- `effect-intend`
- `effects-reconcile`
- `flow-gate-verdict`
- `flow-wip-check`
- `forge-audit-status`
- `forge-links-probe`
- `forge-read`
- `forge-report`
- `generation-apply`
- `generation-plan`
- `github-apply`
- `github-plan`
- `hook-dispatch`
- `hook-receipt`
- `hooks-apply`
- `hooks-plan`
- `identity`
- `index-build`
- `index-query`
- `install-inspect`
- `lane-handoff-check`
- `lane-status`
- `layout-apply`
- `layout-plan`
- `leader-tick`
- `package-inspect`
- `package-resolve`
- `parity-check`
- `plan-admit`
- `pr-body-render`
- `pr-confirm`
- `project-apply`
- `project-plan`
- `project-skill-apply`
- `project-skill-plan`
- `project-skill-revert`
- `queue-next`
- `queue-query`
- `recovery-apply`
- `recovery-plan`
- `run-verify`
- `scope-check`
- `search`
- `skill-read-receipt`
- `skill-source-apply`
- `skill-source-plan`
- `snapshot-describe`
- `update-apply`
- `update-plan`
- `worktree-apply`
- `worktree-plan`

(알파벳순, 한 줄에 operation 하나, 개수는 적지 않는다. `src/main.rs`의 시험 `operations_doc_matches_code`가 이 목록과 OPERATIONS 배열의 이름 집합을 비교해 빠진 것·남는 것·중복·순서 어긋남을 이름으로 보고한다 — op를 더하거나 빼는 PR은 알파벳 위치에 자기 줄만 넣거나 뺀다. 이력: tip `604792666` 47개에서 시작해 rc.2 카드들이 더했고, #2241에서 이 형식으로 바꿨다; `lane-handoff-check`은 A10 #2264 수리 r2에서 더했다. envelope 계약: `{protocol_version, operation, ok, result|error{kind, message, retryable, next_actions?}}`, exit 0/1/2, `kind` 14종 — `src/contract.rs`·`docs/protocol-v1.md`)

envelope 예외 하위 명령: `statusline`(#2350). Claude Code가 `statusLine` 명령의 stdout을 그대로 화면에 보여 주므로 JSON envelope 대신 상태줄 원문을 출력하고 항상 exit 0으로 끝난다. OPERATIONS에 들어가지 않으므로 위 목록에도 없다. 실행기 `cc`·`glm`·`codex`·`session`도 같은 이유로 목록 밖이다. 숨은 시험용 옵션 `--render-budget-ms`는 렌더 예산(기본 800ms, Go와 같음)을 바꾼다. golden 시험이 부하와 무관하게 결정적으로 돌도록 큰 값을 넘기고, `0`은 이미 지난 예산(git을 시작하지 않음)이다. `src/main.rs`의 시험 `statusline_is_the_documented_envelope_exception`이 이 문단과 코드를 대조한다.

## 문서별 내부 명령 체계 → RC1

| 문서 | 문서의 명령 체계 | RC1 대응 |
| --- | --- | --- |
| [25f] | `setup` / `project init` / `maintenance update` / `doctor` / `capabilities` | `project-plan/apply`(setup) · `update-plan/apply`(maintenance update) · `claude-doctor`·`codex-doctor`(doctor) · `capabilities`는 **RC1에 없음**(S11: "공통 명령은 capabilities뿐인데 RC1에 없다") |
| [e08] | `spec plan` / `task status` / `skill show` | `agents-plan` + `plan-admit`(spec plan) · `claim-show`(task status) · skill show 없음(`install-inspect`·`context-bundle`이 부분 대체) |
| [3d9] | `task submit/cancel` / `spec verify` / `artifacts list` | `claim-acquire`/`claim-release`(task) · `run-verify`(spec verify) · artifacts list 없음 |
| [43fa] §18 | 문서 명령 전부 | RC1 OPERATIONS 전수 목록(위 절)으로 교체(S11 결정) |

교체 규칙(S11 결정): 문서의 명령 표는 `src/main.rs` OPERATIONS 표로 교체하거나 "RC1 미출하 계약안" 배너를 단다. 번호·이름으로 누락 검사를 하지 않는다(S8 — 같은 F 번호가 다른 기능).
