# SPEC-CC-DOCS-ALIGNMENT-001 — 진행 기록 (progress.md)

> Tier M · cycle_type = tdd · doc-alignment (5 공식 Claude Code 문서 대비 규칙·문서 정합)
> 33 findings (M1 workflows 6 / M2 skills 8 / M3 hooks-doc 5 / M4 goal 6 / M5 sub-agents 8)

---

## §E.1 Run-phase Summary

5개 공식 Claude Code 문서(workflows / skills / hooks-guide / goal / sub-agents)를 moai-adk 규칙·문서와 대조해 발견한 33건의 정합 결함을 해소했다. 대부분 template-managed 규칙 파일(`internal/template/templates/.claude/...` SOURCE + `.claude/...` MIRROR 이중 편집)이며, `make build`로 embedded 템플릿을 재생성하고 `go test ./internal/template/...`(neutrality + internal-content-leak + mirror-parity gate)로 회귀 0을 검증했다. LOCAL-ONLY 2파일(`CLAUDE.local.md`, `internal/hook/CLAUDE.md`)은 미러 없이 직접 편집했다.

대표 수정: M1 `workflow`→`ultracode` 트리거 리네임(v2.1.160) + `/workflows` 관리/`args` 입력 문서화 · M2 비존재 `type: skill` 필드 제거 + description 한도 1,536 합산캡 통일 · M3 `if`-field v2.1.84→v2.1.85 정정 + Stop block-cap/exec-form/self-gate caveat · M4 native `/loop` 구분 + auto mode 페어링 · M5 `maxContextSize`→`maxTurns` 복원 + `SubagentStart`/fork-subagent 문서화 + `claude-code-guide` built-in 모호성 해소.

**변경 파일 (31)**: template SOURCE 12 + deployed MIRROR 14 + LOCAL-ONLY 2 + SPEC 3 (spec/plan/acceptance). Go source 0건 (`internal/hook/*.go`는 sibling SPEC-HOOK-EVENT-REGISTRY-001 소유, 미접촉).

## §E.2 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: 91674bc19
run_status: implemented
ac_pass_count: 33
ac_fail_count: 0
template_gate: pass            # go test ./internal/template/... -count=1 → ok
neutrality_leak_count: 0       # grep -rc 'SPEC-CC-DOCS-ALIGNMENT' internal/template/templates/ → 0
make_build: exit-0             # embedded 템플릿 재생성, catalog hash idempotent
neutrality_split_files: 4      # CLAUDE.md / agent-authoring.md / archived-agent-rejection.md / agent-common-protocol.md (source↔mirror DIFFER 보존, 수정영역만 정합)
go_source_files_touched: 0     # internal/hook/*.go 미접촉 (sibling SPEC 소유)
preserve_list_post_run_count: 0  # 무관 dirty 파일 미접촉
l1_worktree: agent-abdf3cbbcae2ae687  # 런타임 자율 L1 worktree, orchestrator FF 통합
m1_to_m5_commit_strategy: single-cohesive  # Tier M — M1-M5 단일 run 커밋
```

## §E — Phase 0.95 Mode Selection

- **Input parameters**: tier=M, scope=31 files, domain count=1 (documentation-alignment, 동일 rules/template-mirror 도메인), file language mix=100% markdown(+2 LOCAL-ONLY md), concurrency benefit=LOW (sequential doc edits with mirror dependency), Agent Teams prereqs=미충족.
- **Decision**: sub-agent (Mode 5)
- **Justification**: 단일 도메인 markdown doc-alignment 작업으로, template source→make build→mirror 순서 의존성이 있어 sequential 처리가 필요하다. Mode 6(workflow)은 ≥30 파일이나 "단일 uniform mechanical transform"이 아니라 33개 서로 다른 정합 수정이므로 미충족(Finding A4 + §C.3). GATE-2 사용자 승인(SPEC B 먼저 → A) 하류에서 실행.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-03
sync_commit_sha: PLACEHOLDER_SYNC
sync_status: implemented
changelog_entry: changed         # CHANGELOG.md [Unreleased] ### Changed — 1 entry (doc 정합 = 기존 문서 수정)
readme_touched: false            # 내부 규칙/문서 정합 — user-facing README 무관
docs_site_touched: false         # 내부 .claude/rules 정합 — docs-site 페이지 무관
status_transition: in-progress -> implemented
authored_by: orchestrator-direct # bounded 내부 doc-alignment + active multi-session race 회피 (L_orchestrator_direct_sync_tier_m); Authored-By-Agent trailer 생략 → ownership lint silent SKIP
```

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: 2026-06-03
mx_commit_sha: PLACEHOLDER_MX
mx_status: completed
status_transition: implemented -> completed
four_phase_close: true           # plan + run + sync + Mx 완료
authored_by: orchestrator-direct # implemented->completed: matrix permits manager-docs OR orchestrator (Mx chore)
close_subject_full_id: true      # close commit subject가 full SPEC-ID 명명 (DRIFT-LEGACY-CONVENTION full-ID mandate)
```
