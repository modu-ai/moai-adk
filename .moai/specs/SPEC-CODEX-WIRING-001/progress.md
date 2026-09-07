# Progress — SPEC-CODEX-WIRING-001

## §F Phase 4 Mode Selection

- Kickoff: Implementation Kickoff Approval **통과 (2026-08-24, 운영자 게이트 — 리드 배차
  "t88 run 진입 승인됨")**. Phase 1 verdict: review-3 PASS 0.86 (skip 조건 3종 모두 만족하나
  delta 감사로 최신판정 존재).

Input parameters:
- tier: M
- scope: 신규 패키지 internal/codexwiring(+테스트) + internal/cli 4파일 배선 + annotation 4줄 —
  약 10-14 파일
- domain count: 3 (Go CLI internals / 산출물 렌더링·병합 엔진 / MCP 서버 등록 annotation)
- file language mix: Go ~95% (산출물 검증 명령은 shell/python 일회성)
- concurrency benefit: LOW — 코딩 중심 구현, M1→M2→M3→M4→M5 데이터 계약 의존 사슬
- Agent Teams prereqs: 미요청 (명시적 --team 없음)

Mode evaluation table:
| Mode | 선택 | 근거 |
|---|---|---|
| direct | 아니오 | 신규 패키지 + 다중 파일 구현 — 자격 없음 |
| serial | **선택** | 코딩 중심(Anthropic coding-task 병렬성 주의), 마일스톤 순차 의존 |
| fanout | 아니오 | 단일 도메인 Go 구현, 연구 팬아웃 이점 없음, 쓰기 에이전트 동시 실행 금지 |
| sweep | 아니오 | ~30 파일 미만, 기계적 단일 변환 아님 |

Decision: serial

Justification: Tier M 코딩 중심 구현으로 마일스톤 간 데이터 계약(생성기 코어 → 플래그 배선 →
런타임 seam → doctor → 검증)이 순차 의존한다. 경계 케이스 없음(domain 3은 문턱이나 concurrency
benefit LOW가 tie-breaker로 serial 확정). 단일 manager-develop(cycle_type=tdd)에게 마일스톤
순차 위임, 마일스톤별 커밋.

Boundary Case: 없음.

통합 경로(리드 지시 2026-08-24): v3.1.3 배포가 t204 게이트 보류 중이므로 본 카드 산출물은
**main으로 가는 카드 PR로만 정리 — release 브랜치 통합 없음**.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-23
- artifacts: spec.md / plan.md / acceptance.md / progress.md (Tier M — 3 + progress)
- lint: `moai spec lint .moai/specs/SPEC-CODEX-WIRING-001/spec.md` → exit 0 (2026-08-23 관측, v0.2.0 재확인)
- REQ 14 / AC 14 (MUST 13 + SHOULD 1 — Tier M 상한 16/16 이내)
- v0.3.0 (2026-08-24): 운영자 지시 statusline 범위 추가(카드 t88 본문 개정). 리드 배차
  메시지(main rebase 915c310de 후 착수)로 진행. §A.6 사실 기준 신설 — /statusline 문서·
  config reference·sample·소스 StatusLineItem enum(정식 29종+별칭 7종) 확정, 카드 제안 5종을
  정식 토큰(model-with-reasoning·context-remaining·git-branch·current-dir·thread-id)으로
  확정. REQ-CW-013·014 신설, AC 14건. 신규 토큰 base-0 실측: statusLineAllowlist 0·17827 0
  (status_line 단독은 46hit로 채택 제외 — 산출물 grep으로만 사용)
- grep-AC 토큰 base-0 실측: acceptance.md §A 서두에 기록
- plan-audit review-1 (2026-08-23): FAIL 0.86 — D1 차단(annotation 실태 오측 정정:
  명시 15/21·미선언 6·유효 불일치 4·실효 승인 집합 10) + D2-D6 경미 → v0.2.0으로 전량 적용.
  D1의 4도구 annotation 수정은 plan M2 + PRESERVE 협정 예외로 확정
- plan-audit review-2 (2026-08-23): PASS 1.00 — D1-D6 전량 해소 확인(delta 재감사).
- plan-audit review-3 (2026-08-24): PASS 0.86 — v0.3.0 statusline 범위 개정 delta 재감사
  (운영자 지시 범위 추가로 아티팩트 해시 무효화 → 계약 내 iteration 3/절대상한 3으로 판정
  발행). 경미 D1-D3(미존재 영어 로케일 README 중복 열거 제거·rebase 후 낡은 사실 갱신·별칭 7종
  정정) → v0.3.1로
  전량 적용. 이후 추가 개정 재감사는 사용자 override 경유만. 보고서: review-3.md

## §E.2 Run-phase Evidence

Run-phase executed 2026-08-24, serial mode, manager-develop (cycle_type=tdd), worktree t88 @ branch WT-codex-wiring.

### RED evidence (TDD invariant i — verbatim pre-GREEN output)

- **M1 codexwiring core** — first test run against an empty package:

  ```
  # github.com/modu-ai/moai-adk/internal/codexwiring [github.com/modu-ai/moai-adk/internal/codexwiring.test]
  internal/codexwiring/wire_test.go:17:69: undefined: Result
  internal/codexwiring/configtoml_test.go:13:9: undefined: EnsureMCPTable
  internal/codexwiring/hooks_test.go:33:19: undefined: RenderHooks
  ...
  FAIL	github.com/modu-ai/moai-adk/internal/codexwiring [build failed]
  ```

- **M2 annotation guard** — `go test ./internal/cli/ -run 'TestMoaiMCPServer_AnnotationsMatchCatalog' -count=1` BEFORE the 4 annotation lines (proving the base-tree 4-tool gap is real):

  ```
  --- FAIL: TestMoaiMCPServer_AnnotationsMatchCatalog (0.00s)
      mcp_annotation_guard_test.go:84: tool "audit_cache": effective read-only=false but catalog WriteCapable=false (REQ-CW-011 equivalence: effective read-only ⟺ WriteCapable=false)
      mcp_annotation_guard_test.go:84: tool "codex_audit": effective read-only=false but catalog WriteCapable=false
      mcp_annotation_guard_test.go:84: tool "glm_audit": effective read-only=false but catalog WriteCapable=false
      mcp_annotation_guard_test.go:84: tool "audit_multi": effective read-only=false but catalog WriteCapable=false
      mcp_annotation_guard_test.go:90: 4 tool(s) violate the effective-annotation ⟺ catalog equivalence: [audit_cache audit_multi codex_audit glm_audit]
  FAIL
  ```

  Post-fix GREEN: `--- PASS: TestMoaiMCPServer_AnnotationsMatchCatalog`.

- **M2 --agent flag** — `undefined: resolveAgentWiring / agentWiringCodex / agentWiringBoth / agentWiringClaude` (build failed) before implementation; all 6 tests GREEN after.
- **M2 update wiring** — `undefined: refreshCodexWiringBestEffortAt` (build failed); GREEN after.
- **M3 harness mode** — all 10 tests FAIL before implementation (flag absent); GREEN after (one intermediate fix: tests use `ParseFlags` to trigger cobra's persistent-flag merge, mirroring the real invocation).
- **M4 doctor** — build failed (`checkCodexWiring`/`codexWiringLookPath` undefined); 7 tests GREEN after.

### AC binary matrix (E1)

All commands run 2026-08-24 against the final tree (M5 head, pre-M5-commit). Scratch e2e root: `/tmp/cw-e2e.hhJW53` (binary built `go build -o $SCRATCH/bin/moai ./cmd/moai`; invocations via a HOME-isolating wrapper so the real `~/.claude` was never touched; every project dir under the scratch root). `moai` below = the scratch binary.

| AC | Status | Command (verbatim) | Observed output |
|---|---|---|---|
| AC-CW-001 | PASS | `moai init --help \| grep -c -- '--agent'` / `moai init /tmp/cw-e2e.hhJW53/work/cw-invalid --agent gemini --non-interactive` | `1`; exit `1`, stderr `Invalid --agent value "gemini": must be one of: claude, codex, both.` |
| AC-CW-002 | PASS | jq event keys / handler count / harness grep / version grep / SessionEnd timeout over `proj/.codex/hooks.json` after `moai init proj --agent codex --non-interactive --no-hooks` | keys `PostToolUse,PreToolUse,SessionEnd,SessionStart,Stop,UserPromptSubmit`; HANDLERS=6, HARNESS=6, VERSION_KEY=0, SESSIONEND_TIMEOUT=3 |
| AC-CW-003 | PASS | `grep -c 'default_tools_approval_mode' proj/.codex/config.toml`; `grep -c -e enabled_tools -e disabled_tools …`; `python3 -c "import tomllib;tomllib.load(open(…,'rb'))"` | `1`; `0`; exit 0 (`TOML_OK`) · 2026-09-02 t254 재측정: 원 기록형 `-cE 'a\|b'`는 리터럴 파이프 매치라 0이 구조적 — 다중 `-e`로 교체, 베어 대안 렌더형 재측정 = 동일 0 (바이너리 v3.1.2-1308-g65196a5a7, develop 2660bcd09) |
| AC-CW-004 | PASS | flag-absent `moai init proj2 --non-interactive --no-hooks` → `test ! -f proj2/.codex/hooks.json` ×2 + `jq -r '.mcpServers.moai.command' proj2/.mcp.json` | NO_HOOKS + NO_CONFIG; `moai` |
| AC-CW-005 | PASS | `moai init proj3 --agent both --non-interactive --no-hooks` → file existence + `.mcp.json` entry | `.codex/hooks.json` + `.codex/config.toml` exist; `MCP3=moai` |
| AC-CW-006 | PASS | `shasum -a 256` hooks.json + config.toml before/after re-run (`moai init proj --agent codex … --force`) | `SHA_IDENTICAL` (empty diff) |
| AC-CW-007 | PASS | pre-seeded user hooks entry + `[mcp_servers.other]` + user-modified `[mcp_servers.moai]` + user `status_line = ["model"]` → `--agent codex` wire → greps | OWN_HOOK=1; OTHER_TABLE=1; USER_MOAI_TABLE_KEPT=1; canonical approval NOT injected (=0 — user table byte-invariant); USER_STATUSLINE_KEPT=1 |
| AC-CW-008 | PASS | clause 1: `grep -c 'codex /hooks'` on init stdout → `1`. clause 2: tamper Stop→`user-removed`, `moai update --yes --templates-only` → `grep -c '/hooks to re-trust'` = `1`, `moai hook stop --harness codex` restored = `1`. clause 3: no-change rerun → re-trust grep = `0` | all three clauses observed |
| AC-CW-009 | PASS | `moai update --yes --templates-only` in proj2 (claude) → still no `.codex/*` wiring (NO_HOOKS/NO_CONFIG); in proj (wired) → `shasum` diff empty | `CLAUDE_NO_HOOKS`,`CLAUDE_NO_CONFIG`; `HOOKS_UNCHANGED` |
| AC-CW-010 | PASS | `go test ./internal/cli/ -run 'TestMoaiMCPServer_AnnotationsMatchCatalog' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` (guard GREEN post-fix; RED evidence above proves the teeth) |
| AC-CW-011 | PASS | `go test ./internal/cli/ -run 'HarnessCodex' -count=1` | `ok … 0.858s` — 10 tests covering (a) block substitution + default reason, (b) discard record event/key/content_length, (c) mismatch + unadapted rejection, (d) exit-code-2 ExitCoder passthrough |
| AC-CW-012 | PASS | `moai doctor` in proj → `grep -c 'Codex Wiring'` = `3` (≥1); tampered → warn row `hooks.json differs from the last generated content (sidecar hash mismatch) — run codex /hooks to re-trust the changed hooks` (`/hooks to re-trust` grep = 1 in PLAIN doctor); proj2 → `ok not wired (claude-only project) — skipped`, doctor exit 0 `Fail 0`; `git diff -- internal/template/templates/.codex/` empty | all four clauses observed |
| AC-CW-013 | PASS | `python3 -c "import tomllib;c=tomllib.load(...);print(c['tui']['status_line']==['model-with-reasoning','context-remaining','git-branch','current-dir','thread-id'])"` → `True`; user `status_line = ["model"]` byte-kept (=1); `go test ./internal/codexwiring/ -run StatusLine` → ok (9 tests: canonical five, ⊆ allowlist, 7 aliases 0hit, allowlist=29 canonical, 3 merge branches, nested-table precision, idempotency) | all observed |
| AC-CW-014 | DELEGATED-TO-SYNC | SHOULD — README/docs-site mention of openai/codex#17827 is sync-phase output (plan M5 `[v0.3.0]` note; REQ-CW-014). Run-phase did NOT edit README/docs-site. Base-0 token state: `17827` grep 0hit (acceptance.md §A). | delegation recorded, not executed |

### E2 builds

- `go build ./...` → exit 0 (`BUILD_OK`)
- `GOOS=windows GOARCH=amd64 go vet ./internal/...` → exit 0 (`WINVET_EXIT=0`, no output)

### E3 coverage

- `go test ./internal/codexwiring/ -count=1 -cover` → `ok github.com/modu-ai/moai-adk/internal/codexwiring 0.511s coverage: 87.2% of statements` (≥ 85% DoD)

### E4 subagent boundary grep

- `grep -rn 'AskUserQuestion' internal/codexwiring/ internal/cli/doctor_codex.go 2>/dev/null | grep -v _test` → 0 matches (no output)

### E5 lint

- `golangci-lint run ./internal/codexwiring/...` → `0 issues.`
- `golangci-lint run ./internal/cli/...` → `0 issues.` (full package — NEW-vs-baseline separation moot: zero total)

### E6 commits (branch WT-codex-wiring, NOT pushed — B9-modified: main protected, no push/PR/branch ops)

- M1 `90439c59c` feat: codexwiring generator core (+ spec.md `draft → in-progress` on this commit)
- M2 `30c1387c4` feat: --agent flag + annotation 4-tool fix + guard
- M3 `20ec045c3` feat: --harness codex runtime mode
- M4 `09d34fcc0` feat: doctor Codex Wiring diagnostic
- M5 (this commit) feat: acceptance verification + e2e + placement/self-heal fixes

### Run-phase discovered defects fixed in-place (both within SPEC scope)

1. **Update refresh skipped on "Up to date" updates** — the initial placement of `refreshCodexWiringBestEffort` sat AFTER the `syncSkipped` early return, so an up-to-date update never refreshed the wiring (AC-CW-008 clause 2 failed e2e: re-trust count 0). Fixed by moving the call BEFORE the early return (the refresh does not depend on template redeploy); regression test `TestRunUpdate_CodexRefreshRunsEvenWhenSyncSkips` (RED→GREEN) pins the placement.
2. **Sidecar baseline silently lost** — a force-reinit resets `.moai/state` while an unchanged hooks.json survives, and the sidecar write was gated on HooksWritten only → doctor's divergence detection degraded to "no baseline" (observed live in the scratch e2e). Fixed with self-healing: when the sidecar is absent but the on-disk hooks byte-match the render, the baseline is re-recorded. Regression test `TestWireReestablishesMissingSidecar` (RED→GREEN).

### PRESERVE verification (plan §D)

- `git diff -- internal/template/templates/.codex/` → empty; `git diff --stat internal/template/templates/` → empty (no template additions; Template-First not triggered)
- `git diff --stat internal/codexadapter/` → empty (adapter consumed, never modified)
- `git diff 78c21b526 --stat -- internal/cli/mcp_server.go` → `14 insertions(+)`, 0 deletions — exactly the 4 `WithReadOnlyHintAnnotation(true)` lines + their comments (review-1 D1 PRESERVE exception); codex_task/codex_job_*/glm_* handler logic untouched
- existing init flags + flag-absent behavior: covered by `TestRunInit_AgentAbsentLeavesNoCodexFiles` + the full `TestRunInit` regression sweep (ok)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
run_complete_at: 2026-08-24
m1_to_mN_commit_strategy: one-commit-per-milestone (M1..M5, Conventional Commits, SPEC-ID scope)
m1_commit_sha: 90439c59c
m2_commit_sha: 30c1387c4
m3_commit_sha: 20ec045c3
m4_commit_sha: 09d34fcc0
m5_commit_sha: fdf2a0b96
run_commit_sha: fdf2a0b96
ac_total: 14
ac_pass_count: 13
ac_fail_count: 0
ac_delegated_should: 1  # AC-CW-014 → sync-phase (by design, not a failure)
preserve_list_post_run_count: 0  # violations; templates+adapter empty diff, mcp_server.go = 14 pure insertions
new_warnings_or_lints_introduced: 0  # golangci-lint: 0 issues on both changed packages
coverage_codexwiring: 87.2%  # go test ./internal/codexwiring/... -cover, threshold 85
cross_platform_build:
  darwin_arm64_go_build: exit-0
  windows_amd64_go_vet: exit-0
boundary_grep_askuserquestion: 0
l44_pre_commit_fetch: not-executed  # no push performed (B9-modified: main protected; branch left unpushed for the card PR flow)
l44_post_push_fetch: not-executed
total_run_phase_files: 16  # 7 codexwiring source+test, init.go, update.go, update_codex_wiring.go(+test), hook.go, hook_harness_codex.go(+test), doctor.go, doctor_codex.go(+test), mcp_server.go, init_agent_flag_test.go, mcp_annotation_guard_test.go, 3 doctor goldens
evidence_paths:
  - .moai/specs/SPEC-CODEX-WIRING-001/progress.md
scratch_e2e_root: /tmp/cw-e2e.hhJW53  # transient by design; verbatim outputs quoted in §E.2
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-08-24
sync_commit_sha: "de5d4a2de"  # D3 exemption — backfilled in a follow-up commit after the sync commit lands
b12_self_test_a_changelog_pre_emission_grep: 0  # grep -c 'SPEC-CODEX-WIRING-001' CHANGELOG.md before emission — no duplicate
b12_self_test_b_ac_count: 14  # distinct AC IDs in acceptance.md == CHANGELOG entry scope (13 MUST PASS + AC-CW-014 SHOULD delegated to sync, discharged here)
b12_self_test_c_path_verification: pass  # ls-verified: internal/codexwiring, README.md/.ko/.ja/.zh, CHANGELOG.md
ac_cw014_grep_17827_hits: 4  # README.md 1 + README.ko.md 1 + README.ja.md 1 + README.zh.md 1; docs-site intentionally untouched (later card)
changelog_entry_position: "[Unreleased] > Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (single sync commit, 3-phase close merged; spec.md is the only artifact of the four carrying a frontmatter block)"
  plan_acceptance_progress_md: "no frontmatter block — n/a"
status_transition_commit_strategy: single sync commit carries plan->run->sync close; no separate Mx chore commit; no §E.5 section
canary_compliance_check:
  applicable: false  # SPEC defines no forward-looking policy canaried by its own sync tests
evidence_paths:
  - CHANGELOG.md  # [Unreleased] Added, SPEC-CODEX-WIRING-001 entry
  - README.ko.md  # canonical: Codex statusline limitation sentence (AC-CW-014)
  - README.md
  - README.ja.md
  - README.zh.md
  - .moai/specs/SPEC-CODEX-WIRING-001/spec.md  # frontmatter status close
  - .moai/specs/SPEC-CODEX-WIRING-001/progress.md  # this section
code_changes_in_sync_phase: 0  # internal/ untouched; docs + SPEC artifacts only
```

