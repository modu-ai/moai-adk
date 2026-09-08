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

### M2 — AGENTS.md 5 섹션 + universal 이동 (REQ-UAC-009~012)

- **RED (관측)**: §D.5 EV-6/EV-8 (착지 전 5 섹션 앵커 0건·품질 게이트 포인터 0건 — plan-phase 기록, 본 run에서 §C로 재확인).
- **GREEN (구현)**: plan §D6 제목 그대로 5 섹션 저작(§8 Codex Web Console / §9 Hook Event Coverage — EventTable 실측 기반 12종 중 8종 adapted·4종 미적용·claude 전용 이벤트 델타 설명 / §10 Configuration Map — §6 universal 본문 흡수 / §11 moai CLI Verbs 표 / §12 Status Line Tokens — 포인터 우선, 토큰 목록 미복제). 템플릿 CLAUDE.md §6 는 제목+한 줄 포인터만 잔류(18 제목 불변, `@AGENTS.md` import 유지). LEARNED-WORKFLOW 미러 없음(curator·merge 보호 목록 불변).
- **바이트**: 템플릿 AGENTS.md 15,415 → 18,582 B (+3,167 B — §D6 목표 ≤6.5 KB 이내, 하드 상한 24,576 B 대비 5,994 B 여유). `TestCodexContractByteCeiling` green (루트+미러 쌍). 루트 AGENTS.md 는 15,415 B 유지 — 쌍 동일성 강제 가드가 없어(§C 첫 측정: `internal/` 내 참조는 바이트 상한 측정뿐) plan §H 위험 1 문서화 상태로 다음 `moai update` 흡수를 기다린다.
- **판정 (AC-UAC-009~012)**: 전부 PASS — §D.5 EV-18. 템플릿 무결성: `go test ./internal/template -count=1` ok (36.0s, leak 검사·마커 존재 포함), `make build`(agents-emit-check 선행) green.

### M3 — init --force 안내 + 고정 테스트 (REQ-UAC-013~014)

- **RED (관측)**: §D.5 EV-12 (테스트 컴파일 실패 — `undefined: addCodexReinitGuidance`).
- **GREEN (구현)**: `runInit` 이 executor 가 상담하는 **동일한 validator** 로 already-initialized 를 사전 탐지하고, `--force` + `agentWiringSelection != claude` + `!probe.Valid` 조합에서 `executor.Execute` **전에** 안내 1행 출력(§D.5 EV-19 — 2행 위치, 재초기화는 exit 0 으로 그대로 진행: redirect-not-block). claude 선택·신규 프로젝트는 무음.
- **판정 (AC-UAC-013~014)**: PASS — §D.5 EV-19/EV-20. 고정 테스트 6종 인벤터리 불변, `init_agent_flag_test.go` 테스트 함수 8개 비감소(신규 테스트는 별도 파일 `init_add_codex_guidance_test.go` 3종).

### 감사 이관 노트 판정 (§E.1 carryover)

`TestUpdateAddCodex_ValidationRefusalFailsLoud` 의 타임아웃 시나리오: **바인더블하며 이미 바인드됨 — 별도 타임아웃 불요**. 판정 근거: 거부 시나리오는 순수 로컬(t.TempDir + in-process `Wire` 호출, 네트워크·서브프로세스 없음)이고, 표준 스위트 런에서 반복 관측 시 모두 즉시 완료(패밀리 런 0.975s, `-count=1` 재실행 1.123s — §D.5 EV-20). 행(hang) 경로가 존재하지 않아 타임아웃 추가는 불요로 판정했다.

### 스모크 바이너리 근거

- AC-001~008 스모크: `make build` 산출 바이너리(VERSION=list, M1 코드 상태). templates 미변경 구간이라 임베드 동일.
- AC-013 스모크: M3 코드 상태 `go build -o bin/moai ./cmd/moai` 바이너리(버전 문자열 dev — ldflags 없음; 기능 코드는 M3 커밋 `0cda08931`과 동일).
- 스모크 격리: `MOAI_SKIP_BINARY_UPDATE=1`(shouldSkipBinaryUpdate 의 자체 가드) — 바이너리 자가 갱신 re-exec 가 /tmp 스크래치 스모크를 탈선시키는 것을 막기 위함.

### 최종 검증 (plan §E)

- `go test ./internal/cli -count=1 -timeout=20m` → `ok  github.com/modu-ai/moai-adk/internal/cli  525.580s`, exit 0 — **post-M3 최종 트리 전체 패키지, 직접 관측**. 첫 시도는 go 기본 10분 타임아웃(601.2s panic, 테스트 실패 아님)으로 판정 불 산출 → 타임아웃 상향 후 재측정. `go test ./internal/codexwiring ./internal/config -count=1` → ok (2.1s / 3.5s). 참조: M1 시점 3패키지 동시 런도 ok (cli 564.9s).
- `go vet` exit 0. `golangci-lint run ./internal/cli/... ./internal/codexwiring/...` → 0 issues.
- `GOOS=windows` / `GOOS=linux go build ./...` → exit 0 (plan §E2).
- 경계 grep (plan §E4): `rg -c "add-codex" internal/cli/update.go` = 10 ≥ 1 · `rg -c "MOAI:LEARNED-WORKFLOW" internal/template/templates/AGENTS.md` = 0.
- 커밋: M1 `c26b7fddb` / M2 `8be0e637f` / M3 `0cda08931` + 문서 커밋 1. push 없음(레인 규율 — develop push 는 리드 일괄).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-09
run_commit_sha: 0cda08931   # M-final(문서). M1 c26b7fddb / M2 8be0e637f / M3 0cda08931
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a   # 레인 규율상 push 없음 — 원격 fetch/push 는 리드 일괄 소관
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
  linux_amd64: pass
total_run_phase_files: 12   # 구현 9(update.go, update_codex_wiring.go, update_version.go, update_version_test.go, update_add_codex_test.go, templates/AGENTS.md, templates/CLAUDE.md, init.go, init_add_codex_guidance_test.go) + SPEC 문서 3(spec.md status 전환, progress.md, acceptance.md §D.5 장부 추가)
m1_to_mn_commit_strategy: 마일스톤당 1커밋(M1/M2/M3) + 문서 1커밋 = 총 4커밋, Conventional Commits + 카드 id t589 본문 기재
```

### run-phase 발견 (sync-phase 인계)

1. **AC-UAC-003 판정식 전제 갱신 권고**(manager-spec 소관): 템플릿 배포가 `init --agent claude`에서도 `.codex/agents/moai`를 만든다(SPEC-CODEX-BODY-NEUTRALITY-001 이후 기존 동작). 리터럴 식 `test -e .codex`는 init 직후 상태에서 항상 exists — 실질 계약은 배선 파일 3종(hooks.json/config.toml/sidecar) 부재로 판정해 PASS 확인했다.
2. **clean-reinstall 제2 배선 자리**: plan §D1 ":507 옆" 외에 clean-reinstall 성공 블록에 동일 형태(플래그 가드+거부 전파)의 호출을 추가했다 — 그 early return 뒤에서는 v2 시대 프로젝트(이 동사의 주 대상)가 배선을 영영 받지 못한다(§D.5 EV-14 각주).
3. **RED 측정 슬립**: M1/M3 RED 관측에서 exit code 를 파이프가 소비했다(EV-11/12에 정직 기재). 같은 명령의 GREEN 재실행은 exit 0 직접 관측. 다음 run-phase 에서는 RED 관측을 파이프 없이 찍는다.

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-09-09
- sync_commit_sha: "55b5d5b8e" — backfill 커밋에서 실제 short SHA로 교체 (자기 참조 위험 회피, spec-frontmatter-schema.md D3 면제)
- artifacts: CHANGELOG.md 진입 ([Unreleased] → Added 최상단, t572 진입 위) / docs-site `cli-reference/update.md` 4개 로케일 (ko/en/ja/zh — 플래그 레퍼런스 표 1행 + `--add-codex` 전용 절 1개, 동일 구조) / progress.md §E.4 (본 절) / spec.md frontmatter (`status: in-progress → completed` + `updated: 2026-09-09` — `status` + `updated` 필드만, 본문 무변경)
- 소유 전환: in-progress → implemented → completed (단일 sync 커밋 3-phase close — 스키마 행렬의 manager-docs 소유 행)
- AC: 14/14 PASS (SSOT = acceptance.md AC-UAC-001..014, 판정 근거는 §E.2 + acceptance.md §D.5 EV-13~20)
- sync-audit: PASS (harmonic 96.1 — Functionality 100 · Security 100 · Craft 90 · Consistency 95, blocking 0, Tier M 0.80 상회) — 판정 실물 `.moai/reports/t589/sync-audit.md` (감사 트리 55b5d5b8e, 본 backfill 커밋에서 반출). 선택 발견 2건은 병합 전 수리 강제 없음(F-A1 second-seat 가드 보강 · F-A2 dry-run 출력 스타일 — 후속 카드 후보)
- B12 사전 점검: (a) 중복 grep `grep -c 'SPEC-UPDATE-ADD-CODEX-001' CHANGELOG.md` = 0 → 진행 / (b) AC 토큰 수 일치 — `grep -oE 'AC-UAC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` = 14, 진입이 참조하는 AC 수 14와 일치 / (c) 진입 내 경로 실존 — `.moai/specs/SPEC-UPDATE-ADD-CODEX-001/spec.md` 존재 확인
- 트레일러: 본 sync 커밋에 `Authored-By-Agent: manager-docs` 부착
- canary 준수: 배포 템플릿 파일 미변경 (sync 쓰기 표면은 CHANGELOG.md · docs-site 4파일 · SPEC 아티팩트 2종뿐) — 템플릿 무결성 판정은 run-phase M2 (EV-18) 유지, `make build` 불요
- docs-site/README 조사 결과 (sync 발견): doctor.md 4로케일의 수정 지시문 표(`moai init --agent codex`)는 Go 상수(`internal/cli/doctor_codex.go:52` `initCodexAdvice`)가 출력하는 문자열을 그대로 문서화한 것으로 현행 유지가 정확 — 코드 쪽 지시문을 `moai update --add-codex` 로 바꿀지는 별도 후속 카드 소관 (sync 범위 밖, Go 소스 변경 금지). README 4종(t341 상태줄 한계 서술은 여전히 정확)·init.md(--agent 미문서화)·`.moai/project/*.md` 는 사실 오류 없음 — 무변경
