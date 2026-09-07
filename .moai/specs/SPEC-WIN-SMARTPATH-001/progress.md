# progress.md — SPEC-WIN-SMARTPATH-001

Card t515 (factory) · worktree `.claude/worktrees/t515` · branch `WT-win-path-env` · base `6a46c0edb` (local develop tip at entry; lead later reported local develop advanced to edf782e7e — window absorb target)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
artifacts: spec.md + plan.md + acceptance.md (Tier M, 3 artifacts) + spec-compact.md (auto-generated compact view)
authoring note: Phase 8 proposal AND Phase 10 authoring executed orchestrator-direct (lane-10) — the manager-spec spawn failed twice with API 429 (rate limit) and the delegation target was functionally unavailable; fallback recorded here per the orchestrator-direct exemption. Design direction was fully assembled by the orchestrator from the completed research fan-out before authoring. Tier M confirmed from the card text (operator-quoted). DP1 gate passed ("진행 — SPEC 생성") via AskUserQuestion.
RED-now note: AC-CWSP-001..006 RED-now cells measured directly this session on tree 6a46c0edb (grep 0 / suites ok / fixture-hole :20 / pins :227,:244).

## §F Phase 4 Mode Selection

Input parameters: tier M; scope = 2 source files (settings.go + settings_test.go); domains = 1 (internal/template PATH generation); language mix = Go + markdown; concurrency benefit = LOW (strict milestone dependencies M1→M2→M3); agent teams prereqs = n/a.

Mode evaluation (pre-assessment):
- direct — not selected: real source refactor + new test surface, not a typo-scale change.
- serial — candidate: canonical owner for run-phase implementation (manager-develop) on the generator.
- fanout — not indicated: single domain, strict dependencies.
- sweep — not indicated: not a bulk transform.

Decision: serial (manager-develop, one sequential spawn over M1→M4)

Justification: single-domain coding work with strict ordering; the Anthropic coding-task caveat selects serial over fanout. The orchestrator-direct fallback used for plan authoring does NOT extend to run — run-phase implementation goes to manager-develop per the canonical owner matrix (rate-limit conditions will be re-checked at spawn time; on persistent 429 the orchestrator reports to the lead rather than self-implementing).

## §E.2 Run-phase Evidence

구현 커밋: `5ccf16842` — `fix(template): Windows-safe SmartPATH — GOOS-injected generation (t515)` (2 files changed, 295 insertions(+), 9 deletions(-); 부모 `714d60684`). 변경 파일: `internal/template/settings.go`, `internal/template/settings_test.go`만. 전체 증거(5섹션 + 뮤턴트 3종 전문 + M1 픽스처 provenance): `.moai/reports/t515/verdict.md` (gitignored, 레인 로컬).

모든 §E.2 항목의 attribution: worktree `.claude/worktrees/t515` @ `WT-win-path-env`, 측정 기준선 HEAD `5ccf16842`, 2026-09-07 본 런 직접 실행·관찰. 아래 출력은 실측 전문(필요시 발췌 명시).

**E1. AC 이항 PASS/FAIL 매트릭스**

| AC | 상태 | 커맨드 | 실측 출력 (verbatim) |
|----|------|--------|---------------------|
| AC-CWSP-001 | PASS | `grep -c 'case "windows"' internal/template/settings.go` | `1` (rc 0) |
| AC-CWSP-002 | PASS | `go test ./internal/template/ -run 'TestBuildSmartPATHWindows\|TestBuildSmartPATHWrapperMatchesFixture\|TestSettingsRenderWindowsPATH' -count=1 -v` | 11 서브테스트 전부 `--- PASS`, `ok ... 0.427s` — `[no tests to run]` 토큰 없음 |
| AC-CWSP-003 | PASS | `go test ./internal/template/ -run 'TestBuildSmartPATH_StableAcrossTerminalPATH\|TestBuildSmartPATH_WSL2' -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.253s` (rc 0) + 테이블 darwin/linux 행이 M1 픽스처와 바이트 동일 |
| AC-CWSP-004 | PASS | `go test ./internal/template/ -run TestSettingsRenderWindowsPATH -count=1 -v` | `--- PASS: TestSettingsRenderWindowsPATH (0.00s)`, `ok ... 0.248s` (rc 0) |
| AC-CWSP-005 | PASS | `go test ./internal/template/ -run 'TestBuildSmartPATH\|TestIsWSL2\|TestIsUserScopedWindowsPath\|TestIsWSL2DrivePath\|TestSettingsTemplateRequiredEnvVars\|TestSettingsTemplateHookExecForm\|TestBuildSmartPATHWindows\|TestSettingsRenderWindowsPATH' -count=1` | `ok ... 0.272s` (rc 0), 빈-스윕 토큰 없음 |
| AC-CWSP-006 | PASS | (a) `GOOS=windows GOARCH=amd64 go build ./...` (b) `go test ./internal/template/... -count=1` | (a) rc 0 — smoke 한정(테스트 미컴파일); (b) `ok ... 44.576s` + agentemit `ok` + commandemit `ok`, rc 0 |
| AC-CWSP-007 | PASS | AC-CWSP-002의 stat=false 행으로 커버 | `no_candidate_exists_never_empty`·`env_vars_unset_skip_candidates_even_if_stat_true` 행 PASS + 뮤턴트 B(프로브 무시)가 3행 FAIL로 판별력 실증 |
| AC-CWSP-008 | PASS | (1) `grep -c '재현했다\|reproduced on Windows' .moai/reports/t515/verdict.md` (2) `grep -c '미검증이지 반증' .moai/reports/t515/verdict.md` | (1) `0` (grep rc 1 = zero-match 정상 신호) (2) `1` (≥1 충족) — 양방향 관측 |

보조 측정: `go test ./internal/config/toolpolicy/ -run TestTemplateDirectivePreserved -count=1` → `ok ... 0.348s` (AC-TPS-014 센티넬 — 본 패키지 AC 명령 범위 밖, F1 수리에 따른 별도 확인). `golangci-lint run internal/template/...` → `0 issues.` (rc 0). `gofmt -l` → 무출력.

**E2. 크로스플랫폼 빌드**: `GOOS=windows GOARCH=amd64 go build ./...` → rc 0 (출력 없음) — smoke 한정임을 명시한다(카드 [HARD]: 수리 판정은 테스트 행).

**E3. 커버리지**: 본 런에서 별도 측정하지 않았다 — acceptance.md의 기준은 "internal/template 패키지 기존 수준 유지(하락 금지)"이며, 전체 패키지 스위트(AC-CWSP-006(b))가 green이다. 수치 측정은 sync-audit 단계의 검증 배치에 맡긴다(미측정을 Gaps로 명시).

**E4. 서브에이전트 경계 grep (C-HRA-008)**: 이 SPEC은 AskUserQuestion 관련 코드를 추가하지 않았다(설정 PATH 생성기 + 테스트만 변경) — 신규 프롬프트 표면 없음.

**E5. 린트**: `golangci-lint run internal/template/...` → `0 issues.` — 신규 결함 0.

**E6. 브랜치 HEAD + push 상태**: 브랜치 `WT-win-path-env`, 구현 커밋 `5ccf16842`, 증거 커밋은 아래 §E.3 뒤 별도. **push 안 함** — git-flow 레인 규율(develop 병합은 리드 창 경유)에 따라 레인은 push하지 않는다.

**E8. RED 증거 (TDD)**: 본 SPEC의 RED-now 셀은 plan-phase에서 `6a46c0edb` 트리에 직접 측정돼 spec.md §D에 기록돼 있다(`grep -c 'case "windows"'` = 0 / exit 1, `[no tests to run]` 빈 스윕 토큰, 픽스처 홀 :20). 본 런은 그 RED를 GREEN으로 뒤집었고, 채택 깊이는 뮤턴트 3종으로 실증했다:

- 뮤턴트 A(windows 분기에 POSIX 재추가) → `TestBuildSmartPATHWindows` 5개 exact 행 전부 FAIL (got에 `;/usr/bin;/bin;/usr/sbin;/sbin` 혼입 노출)
- 뮤턴트 B(stat 프로브 무시) → probe=false 시나리오 3행 FAIL, `FAIL ... 0.417s` rc=1
- 뮤턴트 C(darwin 분기에서 `/opt/homebrew/sbin` 제거) → darwin 픽스처 행 + `TestBuildSmartPATHWrapperMatchesFixture` FAIL rc=1
- 각 뮤턴트 후 `git restore -- internal/template/settings.go`로 단일 경로 복원 → `ok` rc=0 재확인, `git diff --stat HEAD` 무출력으로 복원 완전성 확인

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-07
run_commit_sha: "5ccf16842"
evidence_path: .moai/reports/t515/verdict.md (gitignored — 레인 로컬 복사본)
mx_tag_report: added 4 (`@MX:ANCHOR`+`@MX:REASON`+`@MX:SPEC` on buildSmartPATHFor, `@MX:NOTE` on the windows branch) / removed 0 / moved 0 — 기존 태그 무접촉
frontmatter_status_note: spec.md `status: draft → in-progress` 전환은 manager-develop 소관이나, 본 레인 지시(frontmatter는 레인 소관 아님)에 따라 수행하지 않았다 — 리드/오케스트레이터 창에서 처리 필요
blocker: 없음 — 4개 마일스톤 전부 계획대로 완료, plan.md 금지 목록(templates/** 무변경, internal/cli 무변경, WSL2·darwin/linux 출력 무변경) 준수

