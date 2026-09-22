# Card t590 — codex 훅 계약 전수 조사·수리 — 2026-09-09

트리: `.claude/worktrees/t590` @ `957a0ee44` (branch `WT-codex-hook-contract`, base = develop `817990f67`).
기원: 타 프로젝트(mo.ai.kr)에서 운영자가 접수한 진단 보고 — Codex가 MoAI 훅 출력을 "invalid ... JSON output"으로 거부.

## Claim (수리된 결함 3건)

1. **PreToolUse allow 계약 위반 (High, 보고서 주장 — 본 카드에서 독립 재확인)**: Codex 파서는 `updatedInput` 없는 `permissionDecision:allow`를 거부한다(codex-rs `hooks/src/engine/output_parser.rs` `unsupported_pre_tool_use_hook_specific_output`: "unsupported permissionDecision:allow"). MoAI의 안전 경로 허용(`NewSafeDefaultOutput` — permission_mode가 default/plan이 아니면 전부 allow, codex 환경에서는 permission_mode가 항상 공백이라 **모든** 안전 경로가 해당)이 정확히 이 형태를 출력해, Codex가 훅 출력을 무효로 처리하고 사전 검사가 적용되지 않았다.
2. **ask 계약 위반**: `permissionDecision:ask`는 Codex에서 전면 거부("unsupported permissionDecision:ask"). MoAI의 PreToolUse 핸들러는 실제로 ask를 배출한다(`internal/hook/pre_tool.go:513,583`).
3. **hooks.json PreToolUse 이중 등록**: MoAI 관리 네임스페이스 `.codex/hooks/moai/` 안을 가리키는 유산 래퍼 등록(타 프로젝트 실측: `handle-pre-tool.sh`, 내부에서 `--harness codex` 없이 moai 재호출 = Claude형 출력)이 "user"로 보존되고 직접 호출이 추가로 심겨 같은 이벤트에 핸들러 2개. 유산 이력은 moai-adk 소스에서 발생점 미발견(`git log -S handle-pre-tool` 0건 — 외부 유입 유산).

## Evidence (본 카드에서 직접 실행, 이 트리에서 빌드한 바이너리)

**계약 원문 확보**: `output_parser.rs` 617행 전수 판독(/tmp/t590-codex-output_parser.rs). PreToolUse 유효형: `{}` · deny+비공백 reason · allow+updatedInput · legacy decision:block+비공백 reason. 무효형: allow 무 updatedInput · ask · defer(역직렬화 실패) · reason 무 decision · 빈 사유 deny · universal continue:false/stopReason/suppressOutput.

**RED (수리 전, 바이너리 재현 — 보고서 재현의 본 트리 재실행)**:

```
$ printf '{"...","permission_mode":"bypassPermissions"}' | /tmp/t590-moai hook pre-tool --harness codex
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}   exit=0   ← Codex 거부형
$ (permission_mode=default) → {}   ← 유효 대조군
```

유닛 RED: 신규 계약 테스트 8종 중 6종 FAIL(`TestPreToolUseAllowWithoutUpdatedInputDropped` 등) — 고정 tree 817990f67.

**GREEN (수리 후)**:

- 유닛: `go test ./internal/codexadapter/ ./internal/codexwiring/ ./internal/cli/ -count=1` → 3패키지 `ok` (cli 527.882s, -timeout 600s 이내)
- 바이너리 재관측(같은 3페이로드): bypass → `{}` + stderr "codex-adapter: dropped \"hookSpecificOutput\" on PreToolUse (59 bytes): permissionDecision:allow is not accepted..." · default → `{}` · 빈 모드 → `{}` + 공지. stdout은 Codex 소비용 순수 JSON 유지, 열화 공지는 stderr.
- M2 종단(스크래치 `/tmp/t590-scratch/scratch`): 유산 래퍼 + 사용자 항목 시드 → `$ moai update --add-codex --yes` → 렌더된 PreToolUse = `my-own-pre-hook`(사용자, 보존) + `moai hook pre-tool --harness codex`(직접 1개) — 래퍼 제거 확인.
- 품질: codexadapter 87.8% · codexwiring 89.6% (≥85) · `golangci-lint run` 0 issues · `go vet` 통과.
- 커밋: `e84bd3162`(M1 어댑터) → `957a0ee44`(M2 wiring 정리) → 본 sync 커밋(CHANGELOG + docs 4로케일 + 본 보고서).

## Baseline-attribution

모든 측정은 본 실행에서, 본 트리(817990f67 → 957a0ee44)에서 수행. 계약 근거는 codex-rs main 브랜치 `output_parser.rs`(이동 ref — 규약 자체가 대상이라 핀 생략, 보고 시점 판독). 바이너리 2종(/tmp/t590-moai 수리 전, /tmp/t590-moai-fixed 수리 후) 모두 해당 시점 트리에서 `go build`로 직접 생성.

## Gaps (관측하지 못한 것)

- 보고서가 본 실사 이벤트의 원본 hook stdout/permission_mode를 보존하지 않았다고 기재 — 본 카드도 그 원본 이벤트는 관측 못 했다. 대신 동일 형태를 바이너리로 재현해 독립 확정.
- 사용자 프로젝트의 유산 래퍼 `.sh` 실물은 미확보 — 어느 세대의 moai/수작업이 심었는지 발생점 미확정(moai-adk 소스 이력 0건으로 moai 현재·과거 생성 아님 소스 근거만 확보). 본 수리는 발생점과 무관히 네임스페이스 소유로 정리한다.
- `permission-request`·compact 계열은 EventTable 비적응(등록 자체 없음)이라 본 수리 범위 밖 — Codex 파서의 PermissionRequest 계약(decision.behavior형)과 MoAI `NewPermissionRequestOutput`의 정합은 향후 해당 이벤트 적응 카드의 몫.
- `NewSuppressOutput`은 호출부 0건 확인 — suppressOutput을 inertKeys에 넣지 않았다(예측 분기 금지). 향후 호출부가 생기면 그때 측정해 추가.

## Residual-risk

- Codex 파서 규약은 이동 표적(버전마다 변할 수 있음) — 본 수리는 0.153.4 파서 소스 기준이며, golden 테스트가 아니라 "계약 서술 + 변환 고정" 테스트라 Codex 측 변경은 테스트로 자동 탐지되지 않는다. (구조적 한계 — Codex 버전 펌핑 시 재판독 필요)
- deny 경로의 `permissionDecisionReason`이 사용자 데이터라면 기본 사유 채움은 사유를 대체하지 않고 빈 사유만 메운다 — deny 자체는 유지되므로 안전성 손실 없음.
- 어댑터 열화 공지가 stderr로 나가지만 Codex가 이를 사용자에게 표시하지 않을 수 있음 — 진장 기록(`.moai/state/codex-wiring/` 진단 싱크)이 남는 쪽이 진짜 관측면.

## 후속 후보 (운영자 판단용)

- PermissionRequest 이벤트 적응 시 MoAI `NewPermissionRequestOutput`(behavior형)과 Codex 계약 정합 검증 필요.
- `output_parser.rs` 규약을 go 테스트에 golden으로 재단착하는 대신, codex 버전 펌핑 체크리스트에 "hooks 파서 재판독" 1행 추가(경량).
