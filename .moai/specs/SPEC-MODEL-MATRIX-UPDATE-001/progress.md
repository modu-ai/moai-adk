---
spec: SPEC-MODEL-MATRIX-UPDATE-001
tier: M
created: 2026-09-30
author: manager-spec
---

# progress.md — SPEC-MODEL-MATRIX-UPDATE-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC-ID: SPEC-MODEL-MATRIX-UPDATE-001 (Bash regex 검증 PASS, 2026-09-30 — `[[ "SPEC-MODEL-MATRIX-UPDATE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS` → PASS; 기존 카탈로그에 동일 ID 부재, MODEL-MATRIX-* 선행 4종은 전부 superseded로 동명이족 아님)
- Tier: M — 근거: config(defaults/closed_sets/audit_models) + cli(mcp_claude/mcp_codex) + settings(schema_sections) + template(workflow.yaml/llm.yaml/settings.json.tmpl 불변) + web(i18n) + 룰 미러 + 테스트 8+파일. 파일 수 12-15, 마일스톤 5. 단일 도메인 결정(모델 핀 값)이 지배하는 Tier M.
- 아티팩트: spec.md / plan.md / acceptance.md / progress.md 4종 (Tier M 집합; research.md 불요 — 검증표가 plan.md §A에 귀속됨)
- 이슈 소지: **해소됨(2026-09-30 운영자 게이트)** — [NC-1/2/3] 전건 확정되어 plan.md 결정 기록(DR-1/2/3)으로 전환됐다. NC-2는 권장안 기각·전면 삭제 재정, NC-1은 claude-opus-5-5 확정(D7 접기 포함), NC-3은 변경 없음 확정. live 판단 대기 마커 없음.
- 운영자 지시 대체 기록: REQ-AMP-005(Audit.Codex Go 기본 EMPTY 중립성)를 REQ-MMU-001로 대체 — spec.md §C.1.
- plan-audit iter-1: **FAIL 0.875** (`.moai/reports/t1368/plan-audit.md` — MP-7 미해결 마커 게이트 + D2-D6). Wave-1 반영 완료(2026-09-30): D2 AC 19→15 병합·재번호(커버리지 손실 없음), D3 보장-RED 테스트 2파일 AC 편입(10파일), D4 i18n 4 로케일 + 감사자 미발견 동급 effort 설명 표면(spec.md §B REQ-MMU-002에 직접 관측 근거 기록), D5 무효 주석 2표면, D6 llm.yaml 환상 앵커 제거(verify-only 재분류), D8 related_specs 산문 이관. **D1은 wave-2에서 해소**(아래 4차 개정 행 — 결정 기록 DR-1/2/3 전환).
- plan-audit iter-2: **CONDITIONAL 0.875** — 8건 전부 관측 증거로 해소, 잔여 N1(소비자 열거 4파일 누락)만 기계 판정 가능. Wave-3 반영(2026-09-30): N1 M3 항목 6 + AC-MMU-015 13→17파일(단정 재작성 3파일 구분 명시), N2 AC-MMU-011/E8 grep-B 범위 cli/→internal/ 확대, N3 템플릿 llm.yaml context_windows 주석 오버라이드 경로 명시(M3 항목 7), N4 본 §E.1의 스테일 "운영자 응답 대기" 행 갱신. 제3의 감사 라운드 없음 — 레인 기계 검증 후 kickoff.
- 검증 예산: go build + grep/sed만 (go test ./... 금지 — plan.md §D).

## §E.2 Run-phase Evidence

> Run commits: M1 `36d2926fd` → M2 `31bf718d5` → M3 `15b6c1ced` → M4 `b24170567` on WT-model-matrix-update (base d61e7800b). Verbatim command+output evidence: `.moai/reports/t1368/run-evidence.md`. All commands were run in this run, against this tree, env-scrubbed (`unset MOAI_KANBAN … && <cmd>`), `-count=1`, targeted `-run` selectors only (no `go test ./...` — lane discipline; CI owns package-wide verdicts).

| AC | Status | Verification (command → observed output) |
|---|---|---|
| AC-MMU-001 | PASS | grep template+local workflow.yaml codex pin → `templates:117 model: gpt-6.1-sol` + `local:27 model: gpt-6.1-sol`; E7 residual scan `gpt-5.6-sol` → **0행** |
| AC-MMU-002 | PASS | grep defaults.go:1206 → `Model: "gpt-6.1-sol"` (Codex pin in NewDefaultWorkflowConfig.Audit) |
| AC-MMU-003 | PASS | mcp_codex.go:62 `codexAuditDefaultModel = "gpt-6.1-sol"`; model_backend_default_test.go:39 assertion flipped to the fallback pin; `go test ./internal/cli/ -run '...TestCodexResolution...'` → `ok 2.395s` |
| AC-MMU-004 | PASS | codexServableModelPrefixes untouched (mcp_codex.go:156 unchanged — not in any run commit's diff) |
| AC-MMU-005 | PASS | mcp_claude.go:20 `claudeAuditDefaultModel = "claude-opus-5-5"` + :21 effort `medium`; template/local workflow.yaml claude pins match; `grep -n "sonnet" internal/cli/mcp_claude.go` → **exit 1 (0 lines)** |
| AC-MMU-006 | PASS | catalogue.md:70 `defaults to claude-opus-5-5/medium` in BOTH local + template mirror, same commit b24170567; `cmp` → BYTE_PARITY_OK; `TestRuleTemplateMirror` → `ok 0.457s` |
| AC-MMU-007 | PASS | i18n.js model.desc/effort.desc updated in **all 4 locales** (en :586/:588, ko :1481/:1483, ja :2248/:2250, zh :3015/:3017 — 8 rows, same commit) |
| AC-MMU-008 | PASS | `grep -A 2 "func glmDefaultTierEffort"` → `return template.GLMStateMax` (all tiers); rationale comment rewritten; `TestGLMEffortTierDefaults` → `ok 0.893s` |
| AC-MMU-009 | PASS | template llm.yaml effort collapse-map block byte-unchanged (verify-only — not in any run commit's diff); context_windows comment updated per M3 item 7 |
| AC-MMU-010 | PASS | E5 grep → `return []string{DefaultGLM53Flash, DefaultGLM53}`; withdrawal+deletion record comment at closed_sets.go:63-75; defaults.go preservation comment replaced by deletion record |
| AC-MMU-011 | PASS | E8a old-constant scan → **0**; E8b `Models\.Opus\|Models\.Sonnet\|Models\.Haiku` scan → **0** (internal/ scope; mid-run 3 comment hits reworded — deletion record now cites id literals, not identifiers) |
| AC-MMU-012 | PASS | resolveGLMTierSlot (glm.go) implements fallback + one-line stderr warning at the resolved-value consumption point; warning observed verbatim in test batch stderr; `TestResolveGLMModels_WarnsOnRemovedId` asserts exactly-one-line naming id/slot/default |
| AC-MMU-013 | PASS | E6 re-measure: memory.go `"glm-5.3-flash": 1_000_000` + `"glm-5.3": 1_000_000` retained; settings.json.tmpl:417 `"model": "sonnet"` byte-identical to pre-flight (d61e7800b) |
| AC-MMU-014 | PASS | E7 `grep -rn "gpt-5.6-sol" internal/ .moai/config/` → **0행** (all 7 pin test files + fixtures retargeted) |
| AC-MMU-015 | PASS | 17-file test sweep landed across M1-M3 (7 pin files in M1; schema/glm_tier in M2; remaining 8 + the 3 assertion-REWRITE files glm_autocompact/glm_max_context/cg_mode_hardening in M3, rewritten to post-deletion behavior incl. override-path survival tests) |
| AC-MMU-016 | PASS | E1 `go build ./...` → exit 0; `make build` → MAKE_BUILD_EXIT=0 (agents-emit-check green, catalog.yaml hashes unchanged, embeds regenerated); GOOS=windows build → exit 0 |

Gate summary: 16/16 AC PASS, 0 FAIL, 0 PASS-WITH-DEBT. Lint: `golangci-lint run` over the 6 changed packages → `0 issues.` (v2.1.6, CI version). No-regression guard `TestNoConsumerCallPathShips` (internal/jevmeasure, actual home) → `ok 0.569s`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-30
run_commit_sha: 2266fbaf7
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 2
l44_pre_commit_fetch: not-applicable (isolated worktree lane; no primary-checkout fetch performed — worktree-only writes)
l44_post_push_fetch: pending-push (lane does not push; leader batches develop push — push-time fetch owned by the leader)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: pass (go build ./... exit 0)
cross_platform_build.windows: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 38
m1_to_mN_commit_strategy: per-milestone commits M1..M4 (value changes and their test updates ride the same commit per REQ-MMU-007)
evidence_path: .moai/reports/t1368/run-evidence.md
e8_first_audit_spawn_observation: documented-residual (out-of-tree execution; see Gaps in run-evidence.md — in-tree evidence is launcher_test.go:984-985 accepting the full-id forms)
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs가 sync 커밋 착지 시 기입>_
