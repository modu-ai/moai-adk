# progress.md — SPEC-V3R6-CLI-CONFIG-INTEGRITY-001

> Plan-phase skeleton. §E.2/§E.3/§E.4 content is owned by manager-develop (run-phase) and manager-docs (sync-phase) per REQ-ARR-002/REQ-ARR-003.

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
- Tier: M (standard)
- Era: V3R6
- Status: in-progress (M1 run-phase entered 2026-06-24)
- Files: spec.md / plan.md / acceptance.md / progress.md (4-file plan-phase set)
- REQ count: 10 (REQ-CCI-001 ~ REQ-CCI-010)
- AC count: 10 (AC-CCI-001 ~ AC-CCI-010; 9 MUST + 1 SHOULD)
- Evidence verification: all 4 defects verified by direct file:line read (verification-claim-integrity §1.1 surface 3 conformance)
- Out of Scope: 9 items (P1-1 ~ P2-4 + F1 rename), each with `### Out of Scope — <topic>` H3 + bullet
- Plan audit: iter-1 PASS-WITH-DEBT 0.81 → iter-2 PASS-WITH-DEBT 0.83 (D2/D3/D6 fixed, D4/D5 run-phase deferred)

## §E.2 Run-phase Evidence

### M1 — F1 `update -c` help/doc 명확화 (2026-06-24, manager-develop cycle_type=tdd)

**변경 파일** (4, +61/-6):
- `internal/cli/update.go:88` — flag help string을 `"Re-run the init wizard to edit project configuration (no template sync; bare 'moai update' syncs templates)"`로 정정
- `internal/cli/update.go` doc comment (~105-134) — "Reconfigure vs template sync" 섹션 추가, token anchor(`if editConfig { ... }` block) 참조 (line-number-drift-asymmetry 방지)
- `internal/cli/update_mode_test.go` — `TestUpdateCmd_ConfigFlagHelpDisambiguatesReconfigure` 신규 (AC-CCI-001)
- `README.md` / `README.ko.md` — interactive 섹션 + command table에 bare vs `-c` 구분

**행위 변경**: 없음 — `if editConfig { return runInitWizard(cmd, true) }` 블록 byte-identical (doc comment 확장으로 166→179로 drift, 로직 불변)

**RED → GREEN**: 구 help string에서 `TestUpdateCmd_ConfigFlagHelpDisambiguatesReconfigure` FAIL → 신규 help string에서 PASS

**AC**: AC-CCI-001 PASS, AC-CCI-002 PASS

### M2 — F2 model alias central table
_<pending>_

### M3 — F3 acceptEdits 투명성 (D4 pin 대상)
_<pending>_

### M4 — P0-4 db.yaml _TBD_ (D5 처리 대상)
_<pending>_

## §E.3 Run-phase Audit-Ready Signal

### M1 검증 (2026-06-24, orchestrator independent verification batch)

reconcile: manager-develop 격리 worktree(agent-a2146fbd8081ad291) → shared checkout patch 적용 (dry-run PASS 후, base e0a798353 이후 4개 파일 미변경으로 충돌 0)

- V1 빌드(linux): `go build ./...` → exit 0
- V2 빌드(windows cross): `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- V3 go vet: `go vet ./internal/cli/...` → CLEAN
- V4 golangci-lint: `golangci-lint run ./internal/cli/...` → 0 issues
- V5 신규 테스트: `TestUpdateCmd_ConfigFlagHelpDisambiguatesReconfigure` → PASS
- V6 전체 internal/cli: `go test ./internal/cli/ -skip TestEnableTeamMode_NoAPIKey` → ok 12.034s

**Gaps**: `TestEnableTeamMode_NoAPIKey`는 M1 미적용 shared checkout에서도 동일 실패 확인 → pre-existing 환경 실패 (GLM credential), 본 SPEC과 무관 (verification-claim-integrity §3.4 준수)

**Residual-risk**: README command table의 존재하지 않는 `--project` flag 표기 (EN line 646 / KO 674)는 M1 scope 외 — 별도 cleanup 후보

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync>_
