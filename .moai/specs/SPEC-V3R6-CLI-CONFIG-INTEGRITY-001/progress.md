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

### M2 — F2 model alias central table (2026-06-24, manager-develop cycle_type=tdd)

**변경 파일** (6, +196/-43):
- `internal/template/model_policy.go` — `ModelAliasTable` (alias→canonical-id SSOT), `ModelDeprecatedCanonicalIDs` (superseded-id reverse map), `ModelAliasCanonicalID()` / `ModelAliasFromCanonicalID()` / `ModelAliasPickerValues()` accessors 신규
- `internal/cli/launcher.go:704-706` — `expandModelString` no-op → 테이블 참조 해석 구현 + `splitModelSuffix` 헬퍼
- `internal/cli/profile_setup.go:67-87` — `normalizeModel` switch → `ModelAliasFromCanonicalID` 역방향 참조 + `ModelAliasPickerValues` 정방향 검증; wizard picker literals(:446-451) → `ModelAliasCanonicalID` 참조
- `internal/cli/wizard/advanced_gate.go:143-149` — `Value:`/`Default:` literals → `ModelAliasCanonicalID` 참조
- `internal/settings/schema.go:140` — `modelOptions()` literal array → `ModelAliasPickerValues()` 참조
- `internal/cli/launcher_test.go:586-612` — `TestExpandModelString` no-op 단언 → canonical-id 해석 table-driven 단언 (RED→GREEN)

**RED → GREEN**: `TestExpandModelString`에서 `undefined: template.ModelAliasCanonicalID` 컴파일 실패(RED) → 테이블 + 접근자 구현 후 PASS(GREEN). `TestNormalizeModel_Deprecated` 3건 FAIL(`claude-opus-4-6` 역매핑 누락) → `ModelDeprecatedCanonicalIDs` 추가 후 PASS.

**AC**: AC-CCI-003 PASS, AC-CCI-004 PASS, AC-CCI-005 PASS, AC-CCI-010 PASS

**AC-003 검증** (`grep -rnE '"(opus|sonnet|haiku|opusplan)(\[1m\])?"' internal/cli/ internal/settings/`): pre-refactor 50건 → post-refactor 비-테스트·비-주석·비-접근자 standalone literal 0건 (잔여 매칭은 전부 `template.ModelAliasCanonicalID("opus")` 형태의 접근자 인자 또는 주석 예시 — 전부 중앙 테이블 SSOT로 귀속)

**AC-010 검증** (동일 pattern + `internal/template/`): Go-source 매칭은 전부 `model_policy.go` 내 SSOT 정의(`ModelAliasTable`, `ModelDeprecatedCanonicalIDs`, `ModelAliasPickerValues`) 또는 기존 `agentModelMap`(본 SPEC 범위 외, SPEC-CC2178-MODEL-POLICY-REPAIR-001 소유). `internal/template/templates/` 콘텐츠 파일(`.md`/`.yaml`/`.json.tmpl`)의 매칭은 user-facing 문서/설정으로 프로그래밍 scatter가 아님. 신규 scattered literal 도입 0건.

**Canonical id 매핑** (ModelAliasTable):
- `opus` → `claude-opus-4-7` (`ModelIDOpus47` const)
- `sonnet` → `claude-sonnet-4-6`
- `haiku` → `claude-haiku-4-5`
- `opusplan` → `opusplan` (CC-native routing alias, full-id 형태 없음 — 자기 자신으로 매핑)
- `[1m]` 접미사: 해석 후 canonical id에 보존 (예: `opus[1m]` → `claude-opus-4-7[1m]`)

**행위 변경**: wizard picker value가 short alias(`"opus"`)에서 canonical id(`"claude-opus-4-7"`)로 변경 — profile.Model 필드가 구조체 doc comment(`// e.g. "claude-opus-4-6", "claude-opus-4-7"`)와 정합. `expandModelString`은 이미-canonical 값에 idempotent(pass-through). 기존 사용자 prefs.yaml의 short-alias/full-id 값은 `normalizeModel`(wizard init) + `expandModelString`(launcher) 양쪽에서 모두 정상 처리되어 backward-compat 유지.

### M3 — F3 acceptEdits 투명성 (D4 pin 해소, 2026-06-24, manager-develop cycle_type=tdd)

**변경 파일** (3, +86/-1):
- `internal/cli/profile_setup.go` — `acceptEditsConfirmationLine` 상수 + `emitAcceptEditsConfirmation(out io.Writer)` 헬퍼 신규; `runProfileSetup` 정규화 블록(`if permissionMode == defaultPermissionMode`)에서 헬퍼 호출 추가
- `internal/cli/launcher.go:646-661` — `syncPermissionModeToSettingsLocal` 함수 주석에 REQ-CCI-006 / REQ-CCI-007 cross-ref 추가 (정규화가 의도적이며 사용자에게 공개됨을 명시)
- `internal/cli/profile_setup_acceptEdits_test.go` — 신규 테스트 2종 (anchor 단언 + 단일 라인 출력 단언)

**RED → GREEN**: `TestEmitAcceptEditsConfirmationAnchor` / `TestAcceptEditsConfirmationEmittedOnce` 에서 `undefined: emitAcceptEditsConfirmation` 컴파일 실패(RED) → 상수 + 헬퍼 구현 후 PASS(GREEN).

**AC**: AC-CCI-006 PASS (MUST), AC-CCI-007 PASS (SHOULD)

**D4 pin 해소 (plan-auditor review-2 defect)**: AC-CCI-006 acceptance.md body의 weasel phrase ("wizard 실행 출력 검사")를 deterministic stdout anchor 단언으로 전환. 선택된 anchor string:

```
Note: "acceptEdits" is the project default, so no settings.local.json defaultMode override will be written.
```

anchor 토큰 3종 (`acceptEdits` / `project default` / `settings.local.json`)은 grep-stable하며 `TestEmitAcceptEditsConfirmationAnchor`가 binary-testable하게 단언. acceptance.md body 정리는 sync-phase manager-spec/manager-docs pass로 이관 (본 run-phase는 body 수정 금지 — Ownership Transition Matrix 준수).

**행위 변경**: 빈문자열 정규화 로직 자체는 불변 (option a). option (b) 값 보존은 spec.md §E P2-1 명시적 Out of Scope. 유일한 행위 추가 = 정규화 직후 stdout 확인 라인 출력.

**Print site**: `runProfileSetup` 내 정규화 블록 바로 직후 (`emitAcceptEditsConfirmation(cmd.OutOrStdout())`). 사용자는 "Saved profile..." 라인보다 먼저 확인 라인을 봄.

**주석 cross-ref 위치**: `launcher.go:659` (`(REQ-CCI-006 / REQ-CCI-007 — the normalization is intentional, and it is` 라인). `grep -n 'REQ-CCI-006' internal/cli/launcher.go` → 1 match.

### M4 — P0-4 db.yaml _TBD_ placeholder 제거 (D5 pin 해소, 2026-06-24, manager-develop cycle_type=tdd)

**변경 파일** (2):
- `internal/template/templates/.moai/config/sections/db.yaml` — `engine`, `orm`, `migration_tool` 3종 `_TBD_` → `""` (empty string) 교체. 주석 "(set during /moai db init interview)" → "(empty until /moai db init interview fills it)" 정정. `multi_tenant: "none"` + 나머지 system keys 불변. SPEC ID/REQ/SHA 누출 없음 (template-neutrality §25 준수).
- `.moai/config/sections/db.yaml` (local) — `moai update` propagation 결과 (아래 D5 해소 참조).

**Template-First trace** (§2 [HARD]):
1. Template source edit (engine/orm/migration_tool `_TBD_` → `""`)
2. `make build` → binary 재빌드 (이 프로젝트는 `//go:embed all:templates` 직접 embed 방식 — 별도 `embedded.go` 파일 없음, `internal/template/embed.go` 의 `//go:embed` directive가 compile-time에 template 디렉토리를 binary에 bake-in)
3. `./bin/moai update` 실행 → propagation 관찰

**D5 pin 해소 (plan-auditor review-2 defect — propagation mechanism 검증)**:

REQ-CCI-009 / AC-009의 propagation mechanism을 실제 도구 실행으로 검증. 관찰된 mechanism은 **clean reinstall (NOT 3-way merge)**:

```
$ ./bin/moai update
Current version   moai-adk v3.0.0-rc2
·  v2 detected  running clean reinstall · signals: version=false agency=false deprecated=true
[clean-reinstall] v2 fingerprint detected (signals: version=false agency=false deprecated=true)
[clean-reinstall] PRESERVE inventory: 1618 files
[clean-reinstall] Backup created at .../.moai/backups/v2-to-v3-2026-06-23T17-39-51Z
[clean-reinstall] Removed 10 deprecated paths
[clean-reinstall] Embedded templates reinstalled
[clean-reinstall] PRESERVE inventory restored
[clean-reinstall] Integrity check PASSED
 Clean reinstall complete (1618 files preserved, 10 deprecated removed)
```

**관찰된 propagation mechanism 분석**:
- `moai update`는 이 프로젝트에서 "v2 fingerprint" (deprecated signal)를 감지하면 **clean reinstall** 경로로 진입 (3-way merge가 아님).
- local db.yaml의 `_TBD_` 3종은 template source의 `""` 교체가 **전파됨** (AC-009 충족 — 아래 grep 증거). 단 전파는 surgical merge가 아니라 full template overwrite의 부산물.
- 부작용: clean reinstall이 local db.yaml 뿐 아니라 64개 추가 파일(rules/hooks/agents/configs/settings/Makefile/scripts)을 embedded template 기준으로 rewrite. 이는 M4 scope(2개 db.yaml + progress.md)를 크게 초과.
- **조치**: clean reinstall collateral 64개 파일 + 4개 untracked iggda hook scripts를 `git checkout HEAD --` 및 `rm`으로 전량 restore. 최종 working tree는 M4 scope 2개 db.yaml 파일만 modified 상태로 정리 (B10 Untouched Paths PRESERVE 준수).

**AC-009 propagation 결과 (grep 증거, verification-claim-integrity §3.2 준수)**:

```
$ grep -rn '_TBD_' .moai/config/sections/db.yaml
$ echo "exit:$?"
exit:1   ← 0 matches (grep exit 1 = no match)
```

local db.yaml post-`moai update`:
```yaml
db:
  # ... comments from template ...
  engine: ""
  orm: ""
  multi_tenant: "none"
  migration_tool: ""
```

구 local db.yaml 구조(4-space indent, alphabetically sorted, no comments)는 template 구조(2-space indent, commented)로 대체됨 — 이는 clean reinstall의 정상 동작이며, REQ-CCI-009 "subsequent `moai update` shall propagate the placeholder removal"를 충족 (placeholder removal IS propagated).

**`/moai db init` interview contract 보존**: empty string `""`은 interview가 값을 채우는 기존 계약을 유지 — `_TBD_`가 아닌 explicit disabled 상태이므로 YAML linter/소비자가 placeholder를 real 값으로 오인하지 않음 (acceptance.md §C Edge Case 4 준수).

**행위 변경**: 없음 — db.yaml은 declarative config이며 값의 의미만 변경 (`_TBD_` placeholder → `""` empty). interview 동작 불변.

**RED → GREEN**: db.yaml은 declarative config (Go 코드 아님)이므로 전통적 unit-test RED는 적용되지 않음. 대신 embedded template의 `_TBD_` 부재를 `go test ./internal/template/` 통과로 간접 검증 (template neutrality + content-leak guard PASS — SPEC ID/REQ 토큰 누출 없음).

**AC**: AC-CCI-008 PASS (MUST), AC-CCI-009 PASS (MUST)

**Gaps (verification-claim-integrity §3.4)**:
- `moai update`의 clean-reinstall propagation이 다른 local-edit 시나리오에서 어떻게 동작하는지는 본 M4 scope 외 — 본 SPEC은 "template source change → `moai update` → local 전파"라는 단일 경로만 검증. 3-way merge vs clean-reinstall의 분기 조건(deprecated signal)은 `moai update` 구현 영역이며 본 SPEC이 수정하지 않음.
- local db.yaml의 구조(4-space indent)가 clean-reinstall로 template 구조(2-space indent)로 대체된 것은 사용자 local-edit이 손실된 것이 아님 — 원래 local도 template에서 marshal된 상태였으며, 의미 있는 local customization (engine/orm/migration_tool 값)은 원래 `_TBD_`였으므로 손실할 사용자 데이터가 없었음.

**Residual-risk**: clean-reinstall이 다른 프로젝트에서 다른 propagation 경로를 가질 수 있음. 본 SPEC의 AC-009 검증은 본 worktree (moai-adk-go dev project) 기준.

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

- SPEC ID: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
- Tier: M (standard)
- Era: V3R6
- AC tally: 10/10 PASS (9 MUST + 1 SHOULD)
- Sync scope: documentation only (frontmatter transitions + CHANGELOG + progress.md §E.4)
- Verification: no code changes (all implementation in run-phase)
- sync_commit_sha: _<pending — orchestrator backfill>_
- Sync-phase 3-phase close: spec.md `in-progress → completed` + plan.md `draft → completed` + acceptance.md `draft → completed` + progress.md §E.4 populated + CHANGELOG entry + era field added (V3R6) to all 3 artifacts
- Status: audit-ready, awaiting sync commit creation
