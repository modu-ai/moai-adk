# t174 §A 측정 기록 — init 자율성 마법사 수리 전제 재검증

> 카드 t174 plan-phase §A. lane-5 조사 주장 3건을 본 트리(origin/main @
> 1519f2660)에서 재실측했다. **3건 전부 확인** — 판정 근거 포함.

## ① applyAutonomyTierFromWizard — 프로덕션 호출자 0 ✓

- 정의: `internal/cli/init_autonomy_wizard.go:34` (함수 시그니처 `flagChanged bool, flagValue string, res *wizard.WizardResult, opts *project.InitOptions`)
- 비테스트 매치: 3개 전부 주석/정의 (`autonomy_bundle.go:5` 주석, `init_autonomy_wizard.go:4`·`:19` 주석, `:34` 정의) — **호출 부재**
- 테스트 호출자: `init_autonomy_wizard_test.go` + `autonomy_bundle_test.go` (테스트만 살아있는 단위)
- 맥락: `internal/core/project/autonomy_bundle.go`의 "Gap 1 closure" 주석이 이 함수가 마법사의 자율성 티어 선택을 반영한다고 서술 — 즉 **의도된 기능의 배선 누락** 형태 (마법사는 티어를 수집하는데 반영이 안 됨).

## ② --branch-guard·--worktree-auto-* 4플래그 — 등록+적용 계층 존재, 미호출 ✓

- 등록: `internal/cli/init.go:116-119` (Bool 4종, help에 "Enable ... (workflow.branch_guard.enabled)" 식으로 **설정 키까지 안내** — 사용자 대면 계약)
- 적용 계층: `internal/cli/init_workflow_flags.go` — `applyWorkflowBranchGuardFlags`(:36, Changed→GetBool→적용), `runWorkflowConfigStep`(:68), `buildWorkflowToggleEdits`(:106), `readWorkflowToggleDefaults`(:142) + 프롬프트 보조
- **프로덕션 호출자: 0** — init.go:114 주석 1건이 함수명 언급뿐, 호출 부재. 테스트(`init_workflow_flags_test.go`)는 플래그→섹션 적용을 검증.
- 판정 근거: 등록 + 완전한 적용 계층 + 테스트 + 설정 키 실재(예: `Workflow.BranchGuard.Enabled` — t173 카드에서 확인한 로더) = **배선 복구가 의도 부합** 형태. 단, 마법사 경로(runWorkflowConfigStep는 프롬프트 포함)와 플래그 경로(applyWorkflowBranchGuardFlags)의 관계 정리 필요 — 비대화형 경로에서 프롬프트가 돌면 안 됨.

## ③ writeWorkflowAuditYAML — 프로덕션 호출자 0 ✓

- 정의: `internal/core/project/initializer_audit.go:37` (`sectionsDir, opts, result` → audit + codex review-gate 블록 기록)
- 비테스트 매치: 전부 주석 — `initializer.go:80`("When false, writeWorkflowAuditYAML MUST NOT ..."), `init.go:215`("writeWorkflowAuditYAML persists the block exclusively on ...")
- 테스트: `initializer_audit_test.go`가 함수 직접 호출로 동작 검증 (AC-MCP-020 인용)
- 카드가 지적한 "결함으로 미기록": 실측 결과 결함이 아니라 **호출 부재** — initializer의 어딘가에서 호출해야 하는데 그 호출이 없음. 주석들이 존재한다는 것은 배선이 있었거나 예정이었다는 뜻.
- 판정 근거: 주석 2곳이 "MUST/exclusively" 서술 = 의도된 계약. 테스트가 함수 수준 동작을 보증. **배선 복구** 형태 — 단, 호출 지점(initializer.go의 어느 단계, opts의 어늌 필터 조건)을 초기화 흐름에서 정확히 특정해야 함.

## 종합

3건 모두 "죽은 코드가 아니라 배선이 끊긴 의도된 기능"의 형태 — 정의·테스트·주석 계약은 존재하고 프로덕션 호출만 없다. 단, 이것이 제거가 절대 아님을 뜻하지는 않음: (a) 마법사 흐름 자체가 현재 init 경로에서 도달 가능한지 (b) 자율성 티어 플래그가 실제 소비자(config 로더)와 연결되는지 (c) audit 블록의 소비자(무엇이 workflow_audit.yaml을 읽는지) — 세 가지가 각 항목의 "복구 가치"를 결정. plan이 이 3질문에 답하고 항목별 처분을 세운다.

## Gaps

- 마법사(runWizard 계열)의 init 경로 도달성 미실측 — plan에서 확인 필요
- audit 블록의 소비자(리더) 부재 여부 미확인 — 기록만 있고 읽는 자가 없다면 처분이 뒤집힐 수 있음
