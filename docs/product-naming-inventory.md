# 제품 표기 인벤토리 — 사용자 대면 표기 'MoAI' 통일 (#2363)

- 측정 일시: 2026-11-10T05:10:10Z(counts-all.txt mtime, KST 14:10:10) · 측정자: agent-9-t69 (claim `Moai-Claim: 2363#6759493149`, epoch=2)
- 측정 대상: branch `issue/2363-moai-name` @ base `8c57326151e54b908d920fc6b1f17249ad3cb1de` (= 그 시점 `origin/v2026.11.0` tip)
- 패턴: `rg -in 'moai[ _-]cli'` — 제외: `.moai-cli/**`(상태·원장), `.claude/**`, `target/**`, `references/**`, `adapters/**`, `experiments/**`, `Cargo.lock`
- 측정 명령: `rg -in 'moai[ _-]cli' --count-matches`(파일별 매치 수), 브랜드형은 `rg -n 'MoAI[ _-]?CLI|MOAI[ _-]?CLI|Moai[ _-]?CLI|moai CLI'`

## 요약

- 총계: **1,556매치 / 177파일**
- 브랜드형('MoAI CLI'·'MOAI CLI'·'Moai CLI'·'moai CLI'): **2건** — `tests/fixtures/atlas-tag/go-corpus.txt`(테스트 픽스처 코퍼스, Go 어노테이션 원문)
- 무구분자 변형('moaicli'): 0건
- `.github/` 템플릿: 0건(이슈 폼은 이미 'MoAI: …' 표기, PR 템플릿에 제품명 없음)
- 루트 `README.md`: 이 저장소에 없음

## 변경 (7줄 / 6파일 + 본 문서)

| 파일:줄 | 현재 | 변경 후 |
|---|---|---|
| `src/claim.rs:1194` | processes that bypass moai-cli | processes that bypass MoAI — 사용자 메시지. 테스트·픽스처 고정 없음 확인 |
| `docs/protocol-v1.md:3` | moai-cli의 stdin/stdout 계약 | MoAI의 stdin/stdout 계약 |
| `docs/protocol-v1.md:53` | moai-cli가 직접 내는 메시지 | MoAI가 직접 내는 메시지 |
| `docs/local-rc1-kit.md:1` | …에서 moai-cli rc.1 써 보기 | …에서 MoAI rc.1 써 보기 |
| `docs/release-attestation.md:1` | # moai-cli 릴리스 attestation… | # MoAI 릴리스 attestation… — `tests/release_suite.rs` needle(`sha256.sum`, `--latest=false`, `/releases/latest`, tag 리터럴 2종)은 다른 줄이라 단언 불변 |
| `docs/release-notes/2026.11.0-rc.1.md:1` | # moai-cli 2026.11.0-rc.1 릴리스 노트 | # MoAI 2026.11.0-rc.1 릴리스 노트 — 본문 9행은 'MoAI ADK의 … CLI `moai-cli`' 표기 유지(산출물명) |
| `docs/benchmarks/baseline-2026-11-09.md:3` | moai-cli 내부 연산 | MoAI 내부 연산 |
| `docs/product-naming-inventory.md` | (신설) | 본 문서 — 전수 표 |

## 유지 사유 (버킷 — 겹침 없는 분할, 합계 1,556)

| 버킷 | 매치 | 사유 |
|---|---|---|
| 시험·픽스처(tests/** 중 보호 파일 제외) | 476 | 단언·기록 데이터 — `tests/fixtures/github/issues-open.json`의 Issue 제목 인용, go-corpus 'moai CLI' 2건 포함 |
| 보호 파일(PR #2360 소유 6파일) | 467 | `docs/rust-product-path.md` 182, `src/main.rs` 171, `tests/product_cli.rs` 96, `docs/user-entry-draft.md` 14, `tests/launcher_roleless.rs` 3, `src/session.rs` 1. `tests/launcher_glm_key_file.rs`는 0건. #2016 관련 파일(`src/migrate.rs`·`tests/migration_fixture.rs`)은 이 base에 부재 — 제외 조항 공(空) 적용 |
| 기록물(decisions·design-review·e2e-receipts) | 271 | 결정 원장 40, 설계 검토 스냅샷 220, 측정 receipt 11 — 이력 기록이므로 재작성하지 않음 |
| src(보호 파일 제외) | 108 | 대부분 `.moai-cli` 경로·명령·문서 주석. 유일한 사용자 메시지가 위 변경 1행 |
| docs(살아있는 문서, 보호·기록물 제외) | 187 | 경로·명령·tag 참조가 대부분. 위 변경 2~7행이 전부 |
| scripts·tools | 16 | 개발 도구 주석·픽스처 문자열 — 사용자 대면 아님 |
| AGENTS.local.md | 22 | 지시 파일 — 설치 경로·명령 맥락 |
| skills/** | 5 | #2286/PR #2297 보류 결정 — 이번 라운드 미편집(내용은 명령 참조) |
| schemas/e2e-receipt.schema.json | 3 | receipt 필드 `moai_cli_sha` + 기계 판독 description |
| Cargo.toml | 1 | `name = "moai-cli"` — 패키지·바이너리명 불변(Issue 범위 밖) |

식별자 유지 공통: `~/.moai-cli/`·`.moai-cli/` 경로, `moai-cli <subcommand>` 명령, `pub const PRODUCT = "moai-cli"`(`src/install.rs:46`), `moai-cli-v*` tag/ruleset(Issue가 범위 밖으로 명시), 생성 훅 명령 문자열(`src/hooks.rs:73-74` — 출력 계약).

## 전수 파일별 집계

| 파일 | 매치 | 분류 |
|---|---|---|
| `docs/rust-product-path.md` | 182 | 보호(PR #2360) |
| `src/main.rs` | 171 | 보호(PR #2360) |
| `tests/product_cli.rs` | 96 | 보호(PR #2360) |
| `docs/design-review/fixes/top-5490c4ee.md` | 42 | 기록물 |
| `docs/decisions-2026-11-08.md` | 40 | 기록물 |
| `tests/install_cli.rs` | 35 | 시험·픽스처 |
| `tests/layout_cli.rs` | 31 | 시험·픽스처 |
| `tests/hooks_cli.rs` | 28 | 시험·픽스처 |
| `docs/release-checklist.md` | 26 | 표면 문서(식별자) |
| `docs/design-review/review-group3-install-host-plan-run-release.md` | 25 | 기록물 |
| `docs/design-review/review-group1-rc1-plan.md` | 25 | 기록물 |
| `docs/local-rc1-kit.md` | 24 | 표면·변경 |
| `tests/run_verify.rs` | 22 | 시험·픽스처 |
| `AGENTS.local.md` | 22 | 지시 파일 |
| `tests/release_suite.rs` | 21 | 시험·픽스처 |
| `docs/release-attestation.md` | 21 | 표면·변경 |
| `tests/claim_cli.rs` | 20 | 시험·픽스처 |
| `tests/archive_privacy_suite.rs` | 18 | 시험·픽스처 |
| `tests/archive_move_fixture_test.rs` | 17 | 시험·픽스처 |
| `tests/apply_receipt.rs` | 17 | 시험·픽스처 |
| `tests/handoff_cli.rs` | 15 | 시험·픽스처 |
| `src/layout.rs` | 15 | 표면(식별자) |
| `docs/design-review/review-group2-core-brain-sync-owner.md` | 15 | 기록물 |
| `tests/generation_own_reference.rs` | 14 | 시험·픽스처 |
| `docs/user-entry-draft.md` | 14 | 보호(PR #2360) |
| `docs/design-review/synthesis-improvement-report-2026-11-09.md` | 14 | 기록물 |
| `docs/baseline-rc1.md` | 14 | 표면 문서(식별자) |
| `tests/forge_report_cli.rs` | 13 | 시험·픽스처 |
| `src/install.rs` | 13 | 표면(식별자) |
| `docs/support-matrix.md` | 13 | 표면 문서(식별자) |
| `docs/release-notes/2026.11.0-rc.1.md` | 13 | 표면·변경 |
| `src/agents.rs` | 12 | 표면(식별자) |
| `docs/design-review/gap-analysis-rc1-vs-design-2026-11-08.md` | 12 | 기록물 |
| `docs/design-review/fixes/group3-25fa9b23.md` | 12 | 기록물 |
| `tests/claim_verify_cli.rs` | 11 | 시험·픽스처 |
| `tests/archive_plan_fixture_test.rs` | 11 | 시험·픽스처 |
| `docs/hooks-generation.md` | 11 | 표면 문서(식별자) |
| `tests/forge_cli.rs` | 9 | 시험·픽스처 |
| `tests/context_cli.rs` | 9 | 시험·픽스처 |
| `docs/home-layout-p09.md` | 9 | 표면 문서(식별자) |
| `docs/e2e-receipts/receipts.jsonl` | 9 | 기록물 |
| `docs/design-review/review-page_5490c4ee-2026-11-08.md` | 9 | 기록물 |
| `docs/design-review/fixes/group3-487fa473.md` | 9 | 기록물 |
| `docs/design-review/appendix/paths.md` | 9 | 기록물 |
| `tests/worktree_cli.rs` | 8 | 시험·픽스처 |
| `tests/todo_promotion_cli.rs` | 8 | 시험·픽스처 |
| `docs/design-review/autonomy-modes-spec.md` | 8 | 기록물 |
| `tools/host_live_harness.py` | 7 | 도구 |
| `tests/workflow_e2e.rs` | 7 | 시험·픽스처 |
| `tests/profile_apply_engine.rs` | 7 | 시험·픽스처 |
| `docs/design-review/fixes/group3-83f162ef.md` | 7 | 기록물 |
| `tests/leader_cli.rs` | 6 | 시험·픽스처 |
| `tests/generation_receipt_order.rs` | 6 | 시험·픽스처 |
| `tests/audit_admission.rs` | 6 | 시험·픽스처 |
| `src/hooks.rs` | 6 | 표면(식별자) |
| `src/generation.rs` | 6 | 표면(식별자) |
| `src/claim.rs` | 6 | 표면·변경 |
| `docs/webhook-ingest-research-2026-11-09.md` | 6 | 표면 문서(식별자) |
| `docs/design-review/fixes/plan-43faf222.md` | 6 | 기록물 |
| `tests/update_known_good.rs` | 5 | 시험·픽스처 |
| `tests/update_cli.rs` | 5 | 시험·픽스처 |
| `tests/queue_cli.rs` | 5 | 시험·픽스처 |
| `tests/github_profile_cli.rs` | 5 | 시험·픽스처 |
| `tests/github_cli.rs` | 5 | 시험·픽스처 |
| `tests/generation_manifest_single_write.rs` | 5 | 시험·픽스처 |
| `tests/e2e_host_fixture.rs` | 5 | 시험·픽스처 |
| `tests/claude_settings_cli.rs` | 5 | 시험·픽스처 |
| `tests/claim_record_cli.rs` | 5 | 시험·픽스처 |
| `scripts/benchmark.sh` | 5 | 도구 |
| `docs/host-support-matrix.md` | 5 | 표면 문서(식별자) |
| `docs/github-plan-proposal.md` | 5 | 표면 문서(식별자) |
| `docs/github-inventory-rc1.md` | 5 | 표면 문서(식별자) |
| `docs/claude-adapter-contract.md` | 5 | 표면 문서(식별자) |
| `tools/local-rc1-isolation.sh` | 4 | 도구 |
| `tests/worktree_branch_cli.rs` | 4 | 시험·픽스처 |
| `tests/update_known_good_replace.rs` | 4 | 시험·픽스처 |
| `tests/recovery_cli.rs` | 4 | 시험·픽스처 |
| `tests/codex_config.rs` | 4 | 시험·픽스처 |
| `tests/codex_adapter.rs` | 4 | 시험·픽스처 |
| `tests/claude_recovery_test.py` | 4 | 시험·픽스처 |
| `tests/claude_preview_test.py` | 4 | 시험·픽스처 |
| `tests/apply_recover_partial.rs` | 4 | 시험·픽스처 |
| `tests/agents_apply_linkage.rs` | 4 | 시험·픽스처 |
| `src/update.rs` | 4 | 표면(식별자) |
| `src/sources.rs` | 4 | 표면(식별자) |
| `src/profile.rs` | 4 | 표면(식별자) |
| `src/archive.rs` | 4 | 표면(식별자) |
| `src/apply.rs` | 4 | 표면(식별자) |
| `docs/session-hub-research-2026-11-08.md` | 4 | 표면 문서(식별자) |
| `docs/p07-decision.md` | 4 | 표면 문서(식별자) |
| `tests/sources_cli.rs` | 3 | 시험·픽스처 |
| `tests/plan_admission.rs` | 3 | 시험·픽스처 |
| `tests/launcher_roleless.rs` | 3 | 보호(PR #2360) |
| `tests/generation_ref_slash.rs` | 3 | 시험·픽스처 |
| `tests/e2e/host-codex.sh` | 3 | 시험·픽스처 |
| `tests/e2e/host-claude.sh` | 3 | 시험·픽스처 |
| `tests/archive_move_cli.rs` | 3 | 시험·픽스처 |
| `src/self_update.rs` | 3 | 표면(식별자) |
| `src/project_skills.rs` | 3 | 표면(식별자) |
| `src/lib.rs` | 3 | 표면(식별자) |
| `src/forge/report.rs` | 3 | 표면(식별자) |
| `src/claude.rs` | 3 | 표면(식별자) |
| `schemas/e2e-receipt.schema.json` | 3 | 스키마 필드 |
| `docs/host-live-session-checklist.md` | 3 | 표면 문서(식별자) |
| `docs/design-review/statusline-rust-stage1-2026-11-10.md` | 3 | 기록물 |
| `docs/design-review/personal-coding-core-2026-11-09.md` | 3 | 기록물 |
| `docs/design-review/fixes/group3-e0874134.md` | 3 | 기록물 |
| `docs/design-review/fixes/group3-3d917728.md` | 3 | 기록물 |
| `docs/design-review/assignments-schema.md` | 3 | 기록물 |
| `docs/codex-adapter-contract.md` | 3 | 표면 문서(식별자) |
| `tests/update_unmanaged_failclosed.rs` | 2 | 시험·픽스처 |
| `tests/struct_constructors_test.rs` | 2 | 시험·픽스처 |
| `tests/spec_contract.rs` | 2 | 시험·픽스처 |
| `tests/fixtures/protocol/v1/current/21-baseline-inventory-ok.json` | 2 | 시험·픽스처 |
| `tests/fixtures/protocol/v1/current/17-agents-apply-ok.json` | 2 | 시험·픽스처 |
| `tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json` | 2 | 시험·픽스처 |
| `tests/fixtures/github/issues-open.json` | 2 | 시험·픽스처 |
| `tests/fixtures/atlas-tag/go-corpus.txt` | 2 | 시험·픽스처 |
| `tests/claude_transaction_boundary_test.py` | 2 | 시험·픽스처 |
| `src/worktree.rs` | 2 | 표면(식별자) |
| `src/github_templates.rs` | 2 | 표면(식별자) |
| `src/forge/outbox.rs` | 2 | 표면(식별자) |
| `docs/rc-isolation-decision.md` | 2 | 표면 문서(식별자) |
| `docs/protocol-v1.md` | 2 | 표면·변경 |
| `docs/pr-contract.md` | 2 | 표면 문서(식별자) |
| `docs/lane-loop-contract.md` | 2 | 표면 문서(식별자) |
| `docs/e2e-receipts/T1-cargo-test-output.txt` | 2 | 기록물 |
| `docs/e2e-matrix.md` | 2 | 표면 문서(식별자) |
| `docs/design-review/review-group4-past-designs.md` | 2 | 기록물 |
| `docs/design-review/mcp-design-prep-2026-11-09.md` | 2 | 기록물 |
| `docs/design-review/host-session-events-2026-11-09.md` | 2 | 기록물 |
| `docs/design-review/fixes/group2-136c8203bf2881919eb1f543f72f87ea.md` | 2 | 기록물 |
| `docs/design-review/fixes/group2-0d4538a27fb88191a19ba261bd5a3442.md` | 2 | 기록물 |
| `tests/sync_marker_form.rs` | 1 | 시험·픽스처 |
| `tests/support_matrix_test.py` | 1 | 시험·픽스처 |
| `tests/promote_digest_order.rs` | 1 | 시험·픽스처 |
| `tests/parity_cli.rs` | 1 | 시험·픽스처 |
| `tests/ontology_schema.rs` | 1 | 시험·픽스처 |
| `tests/launcher_permissions_setting.rs` | 1 | 시험·픽스처 |
| `tests/host_live_harness_test.py` | 1 | 시험·픽스처 |
| `tests/gitlab_forge_contract.rs` | 1 | 시험·픽스처 |
| `tests/github_inventory_test.py` | 1 | 시험·픽스처 |
| `tests/github_inventory_collect.py` | 1 | 시험·픽스처 |
| `tests/forge_links_cli.rs` | 1 | 시험·픽스처 |
| `tests/forge_audit_status_cli.rs` | 1 | 시험·픽스처 |
| `tests/flow_cli.rs` | 1 | 시험·픽스처 |
| `tests/fixtures/protocol/v1/current/16-agents-plan-create.json` | 1 | 시험·픽스처 |
| `tests/fixtures/protocol/v1/current/09-project-apply-created.json` | 1 | 시험·픽스처 |
| `tests/e2e_receipt_not_run_test.py` | 1 | 시험·픽스처 |
| `tests/common/mod.rs` | 1 | 시험·픽스처 |
| `tests/claude_update_test.py` | 1 | 시험·픽스처 |
| `tests/baseline_git_env.rs` | 1 | 시험·픽스처 |
| `tests/atlas_tag_cli.rs` | 1 | 시험·픽스처 |
| `src/session.rs` | 1 | 보호(PR #2360) |
| `src/scope.rs` | 1 | 표면(식별자) |
| `src/recovery.rs` | 1 | 표면(식별자) |
| `src/leader.rs` | 1 | 표면(식별자) |
| `src/identity.rs` | 1 | 표면(식별자) |
| `src/hosts/codex_config.rs` | 1 | 표면(식별자) |
| `src/forge/mod.rs` | 1 | 표면(식별자) |
| `src/forge/gitlab.rs` | 1 | 표면(식별자) |
| `src/decision.rs` | 1 | 표면(식별자) |
| `src/baseline.rs` | 1 | 표면(식별자) |
| `skills/moai/workflows/update.md` | 1 | skill(#2297 보류) |
| `skills/moai/workflows/sync.md` | 1 | skill(#2297 보류) |
| `skills/moai/workflows/setup.md` | 1 | skill(#2297 보류) |
| `skills/moai/workflows/mode-agent.md` | 1 | skill(#2297 보류) |
| `skills/moai/references/source-manifest.json` | 1 | skill(#2297 보류) |
| `docs/rc-release-review.md` | 1 | 표면 문서(식별자) |
| `docs/p02-decision.md` | 1 | 표면 문서(식별자) |
| `docs/design-review/flow-control-2026-11-09.md` | 1 | 기록물 |
| `docs/design-review/fixes/group3-66acadc2.md` | 1 | 기록물 |
| `docs/code-atlas-name-collision.md` | 1 | 표면 문서(식별자) |
| `docs/claude-preview.md` | 1 | 표면 문서(식별자) |
| `docs/claude-owned-update.md` | 1 | 표면 문서(식별자) |
| `docs/benchmarks/baseline-2026-11-09.md` | 1 | 표면·변경 |
| `Cargo.toml` | 1 | 패키지명 |

## 변경 후 재검증 (실측)

- 편집 뒤 동일 패턴 재검색(`rg -in 'moai[ _-]cli' --count-matches` → counts-post): **1,572매치 / 176파일**. 원 표면은 정확히 **−7**(src/claim.rs 6→5, docs/protocol-v1.md 2→0, docs/local-rc1-kit.md 24→23, docs/release-attestation.md 21→20, docs/release-notes/2026.11.0-rc.1.md 13→12, docs/benchmarks/baseline-2026-11-09.md 1→0)이고, 유일한 신규 매치 파일은 본 문서 자체(**+23**)다. 나머지 170파일은 파일별 매치 수가 1건도 변하지 않았다(`diff counts-all.txt counts-post.txt` — 변동 줄은 위 6파일과 신규 1파일뿐).
- 독립 2차 계수: `rg -io 'moai[ _-]cli'`(기본 엔진)와 `rg -io --pcre2 '(?i)moai[\s_-]cli'`(PCRE2)가 **모두 1,572**로 일치.
- AC2(바이너리명·설치 경로·상태 폴더·PATH 불변): 커밋 diff에 `Cargo.toml`·`PRODUCT` 상수·경로 리터럴 변화 없음을 확인(변경 파일 목록은 6개 수정 + 본 문서 1개 신설뿐).
- AC3(envelope·operation 계약 불변): diff 범위에 `tests/fixtures/protocol/v1/**`·operation 이름·생성 훅 명령 문자열(`src/hooks.rs`)이 없음을 확인.
