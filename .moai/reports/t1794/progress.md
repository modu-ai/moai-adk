# Progress — t1794

- **Card**: t1794 (Class B·P1, 브레인 결정 10-15) — 영수증 없는 감사 PASS 정리(가드 거부 4+1건)
- **Run / lane**: tmvgvn / lane-34 (10-15 원조사) → lane-13 (10-16 재배차, 착지)
- **Tree**: .moai/worktrees/t1794 (branch WT-receipt-cleanup, base e04ebd5ca)
- **Updated**: 2026-10-16T12:2xZ (lane-13, 착지 시점)

## 경과 요약

1. **lane-34 (10-15)**: 원인 규명 — 가드(`internal/hook/audit_receipt_guard.go`)가 1차 저장소의 거부 기록을 보고 manager-git 스폰을 차단. 해제 규칙 K4(같은 역할 증명 PASS 1회) 확인, "리더 실행 대기" 기록 후에도 작업을 이어감: 1차 체크아웃 대상 재감사 2건을 직접 수행(아래 보고서)하고 plan 재감사의 blocking D1을 docs 커밋 `1fa732ef9`(16:24)으로 수리.
2. **lane-13 (10-16 재배차)**: 상태 재검증 → **게이트 전환 발견** → 수용 관측 → 착지.

## Claim (lane-13, 이번 세션의 주장)

| # | 주장 | 판정 |
|---|---|---|
| C1 | 1차 트리 재감사 시도 영수증 3건(10-15 12:52–12:57)은 모두 PASS 증명 불가(fail/no-verdict) — K4 해제는 한 번도 일어나지 않았고 거부 6건(원 4+unknown+T1799 10-15 14:15 신규)이 잔존 | 확인 |
| C2 | 1차 체크아웃의 감사 게이트가 codex→claude로 전환됨(커밋 안 됨, 정밀 3키 diff: `audit.model`/`gates.claude↔codex`/`codex.review_gate.enabled`) → `CodexGateRequired(1차)=false` → 영수증 가드가 1차 트리에서 구조적으로 해제. 시점: 10-15 14:15(T1799 거부 기록 시점) 이후 | 확인 |
| C3 | 카드 수용 기준 달성: 1차 체크아웃 세션에서 manager-git 스폰 관측 — `AUDIT_RECEIPT_VIOLATION` 0 (스폰 정상 발사, 읽기전용 보고 완료) | 관측됨 |
| C4 | sync 재감사 F1[High·blocking] 회귀는 main에서 이미 수리됨 — card t1800 `d1481b321`(WT-t1794-f1-class, 본 카드 감사의 후속 카드). 감사 지정 확인 명령 `grep -c '10,000'` = 배포본 1 / 템플릿 1 (main 기준) | 확인 |
| C5 | plan 재감사 blocking D1은 lane-34 커밋 `1fa732ef9`으로 수리 완료(HISTORY v0.1.6, R-1 narrowing) | 확인 |

## Evidence (명령 + 축자 출력, 본 세션)

1. 스토어 열람 `ls .moai/state/audit-receipts/{receipts,rejections}/` (1차) → receipts 4건(신규 3: rcpt-a658b9c5=codex fail, rcpt-5fe3332a/rcpt-e1a8901f=gate_unmet "required…no verdict (fail-open inconclusive)") / rejections 6건(plan-auditor×1, sync-auditor×5 포함 SPEC-NOSPEC-T1799-000, rejected_at 2026-10-15T05:15:11Z).
2. `sed -n '25,70p' .moai/config/sections/workflow.yaml` (1차, 작업트리) → `audit.model: claude`, `gates: {claude: required, codex: off, glm: off}`.
3. `git diff .moai/config/sections/workflow.yaml` (1차) → 3키만 변경: `model: codex→claude`, `claude: off→required / codex: required→off`, `codex.review_gate.enabled: true→false` (템플릿 churn 아닌 정밀 전환). 이력 맥락: `78f89e22a "adopt codex-single audit gates per operator D8"` — 게이트는 운영자 결정 관리 영역.
4. `moai verify audit-plan` (1차) → `model: claude (config)`, backends claude=required(explicit)/codex=off(explicit)/glm=off, `cross_model_active: false`, `enforced_required: ["claude"]`.
5. 가드 소스(1차 HEAD): `internal/hook/audit_receipt_guard.go` `auditReceiptScope` → `CodexGateRequired(store)`가 false면 가드 무의견(해제·차단 모두 없음). `ClearRejectionsForRoleInTreeKind`(store.go:1155)는 파일 삭제 — 잔여 파일 6건 = 미해제 상태 그 자체.
6. **수용 관측**: 1차 체크아웃 세션(cwd=1차)에서 `Agent(manager-git)` 무기명 스폰 — PreToolUse 거부 없이 정상 발사, 에이전트가 읽기전용 보고 완료: `WT-main-integration-20261015` @ 76106777f 존재·main 비조상(미통합·미push, 카드 전제와 일치), main–origin/main 차이 1,179 커밋, main 선두 `9334f79cf`(t1800).
7. F1 수리 확인: `git show main:.claude/rules/moai/core/hooks-system.md | grep -c '10,000'` → 1; 템플릿 동일 → 1; `git log main --oneline -1 -- .claude/rules/moai/core/hooks-system.md` → `d1481b321 docs(rules): restore the per-additionalContext 10,000-char cap row and the 50K total-stdout clause in hooks-system (card t1800)`.
8. F3 잔존 확인: `git show main:internal/factory/integration_merge_step_test.go` L367 — `exec.Command("git","rev-parse","HEAD^{tree}")` 에 `cmd.Dir` 없음(미수리, 선택).
9. codex 헬스: `codex_task` 프로브(본 워크트리) → `PROBE-OK` (10-16 기준 codex 응답 정상 — 10-15 no-verdict는 일시적).

## Baseline-attribution

- 위 1–9는 모두 본 세션(lane-13, 10-16 21:0x–21:2x KST)에서 1차 체크아웃(HEAD `478285ab4`, detached) 또는 `main`(=`9334f79cf`) 기준으로 실행·관측.
- 판정 빌드: MCP moai 서버 v3.2.0-rc.36-t1822c(commit ee981a932, 1차 HEAD의 후손 — §2.2 상태된 바와 같이 binary가 ref보다 새 lace는 본 측정에 영향 없음: 판정은 파일 읽기·프로세스 관측).
- 이전 세션(lane-34) 산출물 — 재감사 보고 2건(본 디렉터리로 호이스트)과 커밋 `1fa732ef9` — 은 그 보고서들의 Baseline-attribution 절이 담당(audited_sha `e04ebd5ca` = 당시 main).

## Gaps (명시적으로 관측하지 않은 것)

- 게이트 전환(codex→claude)의 **주체·승인 경로**는 관측하지 못함 — diff의 정밀성과 D8 선례, 전환 시점(어제 재감사 실패 직후)이 "의도된 전환"을 강하게 시사하나 커밋되지 않아(1차 미커밋 dirt) 작성자를 특정하지 못했다. 리더/운영자 확인 필요(완료 보고에 명시).
- K4 루트북(감사관 2스폰)은 실행하지 않았음 — 현재 게이트 상태에서 구조적으로 불가(가드 해제)이자 불필요(차단 소멸). 미실행은 선택이 아니라 구조적 결과.
- 재감사 보고서 2건의 재검증(테스트 재실행)은 하지 않음 — 보고서 내용을 있는 그대로 인용했고, F1 수리만 main 기준으로 독립 확인.
- 거부 6건 파일은 그대로 남음(게이트 off 동안 inert). 삭제하지 않았음 — 수동 삭제는 비인가 개입.

## Residual-risk

- **게이트 전환의 내구성**: 1차 workflow.yaml 변경이 미커밋 dirt 상태. 커밋 전에 revert되면 영수증 가드가 재개되고 거부 6건이 다시 manager-git 스폰을 차단. 그때의 해법(K4 해제 절차, 인라인): ① `gates.codex: required` 상태의 1차 체크아웃 세션에서 ② sync-auditor·plan-auditor 각 1회 무기명 스폰 — 각 감사관은 스스로 `codex_audit`/`audit_multi`(project_root=1차 체크아웃)를 호출해 1차 저장소에 영수증(gate_unmet 없고 inconclusive 아닌 판정)을 발급받은 뒤 ③ 감사 결과가 실제로 PASS일 때만 마지막 줄을 `AUDIT-VERDICT: PASS spec=<SPEC-ID> receipts=<rcpt-…>`로 끝낸다 — 영수증이 gate_unmet 없고 inconclusive 아닌 것만으로는 부족하다(주의: `CheckCitedReceiptsSince`는 fail 판정 영수증도 기계적으로 인용을 허용하므로, 실제 판정이 fail이면 PASS 줄을 써서는 안 된다). FAIL이면 `AUDIT-VERDICT: FAIL …`로 끝낸다(FAIL은 영수증 불요) — 결함 수리·재감사가 PASS를 증명하기 전까지 해당 역할의 거부 상태를 유지한다(해제하지 않는다). 같은 역할의 증명 PASS 1회가 그 트리·역할의 영수증류 거부를 전부 해제한다(`internal/hook/audit_receipt_guard.go` checkAuditorStop, 운영자 결정 K4).
- **F2[Medium·프로세스]**: 1차 체크아웃이 낡은 미커밋 변경(workflow.yaml·llm.yaml·harness 등)을 계속 보유 — ec193e477류 "pending 커밋" 재발 시도 F1류 회귀 재발 위험. 리더 정리 소관(감사 지적 그대로).
- **F3[Low·선택]**(t1479 소관)·**F4[Low·선택]**(F-11 grep 표류) 미수리 — 선택 등급이라 본 카드 범위에서 제외.
- 영수증 가드 off 상태에서 신규 감사 PASS의 영수증 증명은 1차 트리에서 강제되지 않음(claude 단일 백엔드 체제의 설계된 결과).

## 산출물

- `.moai/reports/t1794/progress.md` (본 파일)
- `.moai/reports/t1794/plan-reaudit-primary.md` (lane-34, 10-15 13:01 — 1차 트리 호이스트)
- `.moai/reports/t1794/sync-reaudit-primary.md` (lane-34, 10-15 13:26 — 1차 트리 호이스트)
- `.moai/reports/t1794/sync-audit.md` (lane-13 sync 감사 r1: FAIL — codex 필수 레그 fail, F1 금단 포인터·F2 증거 명령 정밀도, r2 파일에서 회차 이력)
- 커밋 `1fa732ef9` (lane-34, plan 재감사 D1 수리 — 본 카드 브랜치 유일 소스 커밋)
- sync 감사 r1 수리: F1(K4 절차 인라인 추가)·F2(증거 명령 전체 경로 수정) — r1 보고서 fix_scope 그대로
