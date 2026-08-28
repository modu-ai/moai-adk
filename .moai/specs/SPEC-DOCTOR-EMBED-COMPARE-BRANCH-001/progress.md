# progress.md — SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001

카드 t356 · 브랜치 `WT-embed-compare-branch` · 기준점 `c6aa61346` (= origin/develop)

## §E.1 Plan-phase Audit-Ready Signal

- Tier S 산출물 작성 완료: `spec.md` · `plan.md` · `acceptance.md` · `progress.md`
- SPEC ID 정규식 자가검사 실행됨 (`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`) → `PASS`
- ID 유일성: `.moai/specs/SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001` 부재 확인 (700개 SPEC 대조)
- status: `draft` — run-phase 진입 전 Implementation Kickoff Approval 필요
- 미해결 `[NEEDS CLARIFICATION]` 없음

## §E.2 Run-phase Evidence

축자 명령·출력의 정본은 primary 체크아웃의 `.moai/reports/t356/verdict.md`이다
(워크트리 사본은 gitignored이며 폐기 시 유실). 아래는 AC별 판정 요약이다.

| AC | 상태 | 검증 명령 | 관측 출력 |
|---|---|---|---|
| AC-DECB-001 | PASS | `go test ./internal/cli/ -run TestAgentEmitEmbed_ComparisonErrorFails -v -count=1` | `--- PASS: TestAgentEmitEmbed_ComparisonErrorFails (0.00s)`, exit 0 |
| AC-DECB-002 | PASS | 동일 (테스트 내부 단언) | mutant 실행에서 관측된 메시지가 `comparison failed: read committed manager-git.toml: … is a directory` — 형제 3접두 미포함 |
| AC-DECB-003 | PASS | mutant(`:146` `CheckFail`→`CheckOK`) 상태에서 동일 명령 | exit 1, `--- FAIL: TestAgentEmitEmbed_ComparisonErrorFails (0.00s)` |
| AC-DECB-004 | PASS | 원복 후 동일 명령 | exit 0, `--- PASS: TestAgentEmitEmbed_ComparisonErrorFails` (패키지 요약 `ok …`는 근거로 미채택) |
| AC-DECB-005 | PASS | `git diff --stat -- internal/cli/doctor_agentemit_embed.go` | 빈 출력 (diff 0줄) |
| AC-DECB-006 | PASS | `go test ./internal/cli/ -run 'TestAgentEmitEmbed' -v -count=1` + `go test ./internal/cli/... -count=1` | 매치 12건 전부 PASS(형제 3종 포함); 패키지 전체 exit 0 |
| AC-DECB-007 | PASS | `ls /Users/goos/MoAI/moai-adk-go/.moai/reports/t356/verdict.md` | 존재 (8188 bytes), 워크트리 사본과 `cmp` 일치 |
| AC-DECB-008 | PASS | `git diff -- internal/cli/doctor_agentemit_embed_test.go \| grep -c '^+func '` | `1` — `+func TestAgentEmitEmbed_ComparisonErrorFails(t *testing.T) {`; 새 헬퍼 0건 |

불변식:

| 불변식 | 상태 | 근거 |
|---|---|---|
| 프로덕션 코드 무변경 (REQ-DECB-005) | PASS | `git diff --stat` = 테스트 파일 1개 49행 추가만 |
| 새 픽스처 헬퍼 0 (REQ-DECB-006) | PASS | `+func ` 1행이 곧 테스트 선언 |
| `go vet ./internal/cli/` 무경고 | PASS | exit 0, 무출력 |
| `gofmt -l` 빈 출력 | PASS | 무출력 |

미검증(Gap)은 verdict.md § Gaps 참조 — 원격 CI 판정, 로컬 전체 스위트, Windows 크로스 빌드,
`golangci-lint`는 이 run-phase에서 관측하지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha: pending-backfill-run
run_status: audit-ready
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 1   # internal/cli/doctor_agentemit_embed_test.go 이외 전부 무변경
l44_pre_commit_fetch: not-run     # 통합 창 동결 — push 없음, 커밋 직전 HEAD/브랜치 재판독으로 대체
l44_post_push_fetch: not-run      # push 없음
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass                    # go vet ./internal/cli/ exit 0 (이 트리)
  windows: not-measured           # 테스트 파일 1개, 플랫폼 분기 없음 — 미측정
total_run_phase_files: 1
m1_to_mN_commit_strategy: single-commit  # Tier S 단일 마일스톤 M1
```

## §E.4 Sync-phase Audit-Ready Signal

증거의 정본은 primary 체크아웃의 `.moai/reports/t356/verdict.md`이다. 그 파일은 executor의
run-phase 관측(§ Claim ~ § Residual-risk)과, 오케스트레이터가 executor 보고와 **별개로 직접
심고 원복하며 재관측한** 기록(§ 오케스트레이터 독립 재관측 R1-R5)을 함께 담고 있다. sync-phase는
그 기록을 읽었을 뿐 mutation을 다시 돌리지 않았다 — 재실행하지 않았다는 사실을 여기 남긴다.

```yaml
sync_complete_at: 2026-08-28
sync_commit_sha: 72052f998
sync_status: audit-ready
b12_self_test_a: skipped-changelog-out-of-scope   # grep -c 'SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001' CHANGELOG.md → 0 (관측), 그러나 방출 자체가 SPEC §Out of Scope
b12_self_test_b: pass                             # acceptance.md AC 토큰 8건 == §E.2 판정 행 8건
b12_self_test_c: pass                             # 주장된 경로 2건 실재 확인 (아래 evidence_paths_verified)
changelog_entry_position: none                    # 의도적 미방출 — 사용자 표면 변화 0, 약속은 t346 c2b51293e가 이미 적재
evidence_paths_verified:
  - /Users/goos/MoAI/moai-adk-go/.moai/reports/t356/verdict.md
  - internal/cli/doctor_agentemit_embed_test.go
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed
  plan_md: n/a-no-frontmatter-block
  acceptance_md: n/a-no-frontmatter-block
  progress_md: n/a-no-frontmatter-block
  updated_field_refreshed: 2026-08-28
mx_tag_validation: no-op                          # 신규 최상위 선언이 테스트 함수 1개, 프로덕션 diff 0 — @MX 대상 없음
docs_surfaces_touched: none                       # CHANGELOG / README / docs-site / codemaps 전부 무변경
canary_compliance_check: n/a                      # 이 SPEC은 전방위 정책을 정의하지 않음
push_state: not-pushed                            # 통합 창 동결 — 커밋만
```

미검증(sync-phase 자체의 Gap):

- **원격 CI 판정 없음.** push하지 않았으므로 darwin/windows 매트릭스 판정은 통합 시점의 몫이다.
- **mutation 재실행 없음.** R1-R2는 오케스트레이터가 이미 관측한 기록을 읽었을 뿐, sync-phase에서
  mutant를 다시 심지 않았다.
- **`golangci-lint` 여전히 미실행.** run-phase와 동일하게 `go vet` + `gofmt`까지만 관측됐다.

잔여 위험: verdict.md § Residual-risk의 Glob-디렉터리 결합이 그대로 남아 있다. 그 성질이 수리되면
이 테스트가 조용히 `:146`을 지키지 않게 될 수 있으며, 수리 자체는 SPEC이 별도 카드로 미룬 항목이다.
