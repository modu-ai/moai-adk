# Progress — SPEC-UPDATE-ADD-CODEX-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-09
- plan-audit 판정: **PASS (1.00)** — iteration 2/3 (iter1 PASS-with-fixes 0.86 → F1/F2 수리 → 델타 재감사 PASS). 보고서: `.moai/reports/t589/plan-audit.md` § Re-Audit. Skip-eligible 요건(PASS + Tier M 0.80 이상) 충족 — artifact-hash 대상은 v0.2.0 최종 상태
- artifacts: spec.md, plan.md, acceptance.md, progress.md — Tier M (3 artifacts) + progress (v0.2.0)
- authored: 2026-09-09 by manager-spec, worktree `.claude/worktrees/t589` @ `5caddeb2d`
- iter1 차단 수리 이력: F1(REQ-UAC-005 기계 판정 — §D.4 #1 구체 테스트명 + plan §D1 거부 전파 구속) / F2(AC-UAC-011 `@AGENTS.md` import grep 추가) / F5(§A t585 공유 표면 조율 문장)

**수용된 부채 (전부 optional — run-phase 블로커 아님, 감사 M6):**

1. **F3** — 감사 인용 `.moai/reports/init-tui-audit-20260909.md` 가 primary 체크아웃 전용 비추적 파일이라 이 워크트리에서 미해결 경로(원문 확인·§7 검증은 primary에서 수행됨). 후속: 워크트리 증거 사본 또는 plan §I primary 경로 명기
2. **F4** — 미세 좌표 오차: `RefreshWiring` 은 `wire.go:51-56`(표기 :57은 내부 return 행), `init_agent_flag_test.go` 실측 163행(표기 161). 하중을 지는 좌표 아님
3. **F6** — plan §D1 "같은 조건부 자리" 표기 잔존(실제 update.go:499-507 슬롯은 조기 반환 이전의 무조건 호출 밴드). 단 D1 의 하드 전파 조항 + ungated 직접 호출 구속이 배치 모호성의 실위험을 흡수
4. **문구 잔여** — acceptance.md §D.3 REQ-UAC-005 행의 "간접" 라벨(F1 승격 후 기계 판정 기준과 비모순적 잔여)

**run-phase 이관 노트 (감사):** `TestUpdateAddCodex_ValidationRefusalFailsLoud` 의 바인더블 타임아웃 시나리오 존재는 미검증이다 — run-phase 에서 불가능 판정 시 해당 테스트에 타임아웃을 추가한다.

## §E.2 Run-phase Evidence

측정 트리: `.claude/worktrees/t589` @ `5caddeb2d` (branch `WT-add-codex-verb`), run-phase 2026-09-09 by manager-develop. §C 재측정: plan 기대값과 전부 일치(HEAD `5caddeb2d`, `add-codex` 0 matches, AGENTS.md 15,415 B, 스킬 16종, `## ` 7 섹션, 고정 테스트 6종 인벤터리 동일). B1 클린 재측정: `os.Rename` 후보는 `internal/cli/update/deploy/deploy.go:280`·`internal/cli/memory.go:334` 등 — M3는 백업 메커니즘을 변경하지 않음.

### M1 — update --add-codex (REQ-UAC-001~008)

- **RED (관측)**: `go test ./internal/cli -run TestUpdateAddCodex` → 컴파일 실패 (`undefined: addCodexWiringAt`, `undefined: emitAddCodexDryRunPreview`, `too many arguments in call to validateUpdateVersionConflicts`); `./bin/moai update --add-codex` → `Unknown flag: --add-codex.` (EV-1 재관측). 장부: acceptance.md §D.5 EV-11~
- **GREEN (구현)**: `addCodexWiringAt`(ungated `codexwiring.Wire` 직접 호출, `ErrValidationRefused` 하드 전파 — plan §D1), `emitAddCodexDryRunPreview`(REQ-UAC-006), `validateUpdateVersionConflicts`에 `--check × --add-codex` 상호배타 확장(REQ-UAC-007), 플래그 등록+help(REQ-UAC-008), runUpdate 배선 2소관 — :507 게이트 리프레시 옆 + clean-reinstall 성공 블록(early return이 배선 자리보다 앞서 v2 시대 프로젝트 — 이 동사의 주 대상 — 이 무음 no-op가 되는 것을 막음; D1 단일 위치 표기에 대한 측정 기반 추가, F6 예고 영역).
- **판정 (AC-UAC-001~008)**: 전부 PASS — verbatim 출력은 acceptance.md §D.5 EV-11~17 참조. 스모크 바이너리는 `make build` 산출(`bin/moai`, VERSION=list). 바이너리 자가 갱신이 스모크를 탈선시키는 것을 막기 위해 코드베이스 자체 격리 가드 `MOAI_SKIP_BINARY_UPDATE=1`(reexecNewBinary 루프 방지용, update.go `shouldSkipBinaryUpdate`)을 스모크에 사용 — 배선 판정 대상과 무관.
- **테스트**: `go test ./internal/cli/ ./internal/codexwiring/ ./internal/config/ -count=1` → ok/ok/ok (cli 564.9s, exit 0). 신규 `TestUpdateAddCodex_*` 10종 + 기존 `TestWireValidationRefusalWritesNothing` green 유지. `go vet` exit 0.

### 측정 기반 편차 (M1)

1. **AC-UAC-003 판정식의 전제 갱신 필요(동기화 단계 권고)**: 이 트리의 `moai init --agent claude`는 템플릿 배포로 `.codex/agents/moai`를 **이미 생성한다**(SPEC-CODEX-BODY-NEUTRALITY-001 이후 기존 동작 — 배선 파일 hooks.json/config.toml과 무관). AC-UAC-003 리터럴 식(`test -e .codex` absent 불변)은 init 직후 상태에서 성립하지 않음. 실질 계약(REQ-UAC-002 — update 동사가 배선을 만들지 않음)은 `hooks.json`·`config.toml`·sidecar 부재로 판정해 PASS 확인. AC 식 수정은 manager-spec 소관 — 본 run-phase에서는 발견만 기록.
2. **clean-reinstall 경로 배선 추가(위 GREEN 항)**: plan §D1 "호출 위치 :507 바로 옆"에 대한 추가 소관. 첫 update가 DeprecatedPaths 시그널로 clean-reinstall로 분기하는 것은 이 트리 실측(스크래치 3회 재현)이며, 그 early return 뒤에서는 REQ-UAC-001의 산출물이 영영 만들어지지 않는다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — M3 종료 시 기입>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
