# 제품 표기 인벤토리 — 사용자 대면 표기 'MoAI' 통일 (#2363)

- **이 판은 감사 FAIL(6762643101) 수리로 전면 재생성된 것**이다. 이전 판(1,556→1,572 재검증)은 base 병합 이전 트리의 스냅샷이라 후보 head와 어긋났고, 그 불일치가 감사 FAIL의 사유다. 경과는 맨 아래 부록에 보존한다.
- 측정 트리: **`346f3cf3e75afc808a1042a5d77f7c03c568f830`**(본 수리 커밋의 부모 — branch `issue/2363-moai-name`, origin/v2026.11.0 `e0f2fb0854` 병합 완료 상태; 측정 직전 base가 한 번 더 이동해 재병합·재측정했다)
- 측정 일시: 2026-11-10T10:09:44Z ± · 측정자: agent-9-t69 (claim `Moai-Claim: 2363#6762753812`, epoch=3; 최초 claim 2363#6759493149 epoch=2)
- 패턴: `moai[ _-]cli`(대소문자 무시, 매치 단위) — 제외: `.moai-cli/**`(상태·원장), `.claude/**`, `target/**`, `.git/**`, `references/**`, `adapters/**`, `experiments/**`, `Cargo.lock`
- 두 개의 독립 계수(서로 다른 엔진):
  1. `git grep -Iio 'moai[ _-]cli'`(tracked 파일 스캔) → **1,684매치 / 189파일** — 본 문서의 전수 표는 이 값
  2. BSD `grep -RIioE 'moai[ _-]cli'`(모든 파일, 위 디렉터리 제외) → **1,683매치** — 1매치 차이는 엔진 파일 범위 차이(tracked 한정 vs 무시 파일 포함)이며 교차 확인으로 충분하다. `rg --count-matches` 병행 실행은 이 문서 작성 시점에 미완료(3차 확인 not_run — 2개 독립 계수로 충분).

- 참조 점검(#2405, 2026-11-10, tip `b87e3d7d9c`): 전수 표 189행 중 **6행**이 지금은 없는 경로를 가리킨다.
  - `tests/omp_adapter.rs`는 PR #2381에서 삭제됐다.
  - `skills/moai/…` 5개는 PR #2297이 `skills/moai-cli/…`로 옮겼다.
  - 행과 매치 수는 측정 트리의 값이라 그대로 두고, 각 행의 마지막 칸에 현재 상태를 적었다.
  - 측정은 두 방법으로 했다(Python 파싱, `grep`+`test -e`). 결과가 같았다.
- 재점검(#2405, 2026-11-10, tip `33bfb5e3ca` 병합 뒤): **8행**이 없는 경로를 가리킨다. 위 6행에 2행이 더해졌다.
  - `tests/todo_promotion_cli.rs`와 `tests/promote_digest_order.rs`는 PR #2406(`b0c590c7c9`, 로컬 TODO 승격 코드 제거)에서 삭제됐다.
  - 처리 방식은 위와 같다. 측정은 같은 두 방법으로 했고 결과가 같았다.

## 요약 (측정 트리 기준)

- 총계: **1,684매치 / 189파일**(방법 1; 방법 2는 1,683)
- 브랜드형('MoAI CLI'·'MOAI CLI'·'Moai CLI'·'moai CLI'): **5건 = 본 문서의 측정 패턴 인용 3 + `tests/fixtures/atlas-tag/go-corpus.txt`(테스트 픽스처 코퍼스) 2** — 제품 표면의 브랜드형은 0
- 무구분자 변형('moaicli'): **1건 = 본 문서의 패턴 인용 1** — 제품 표면 0
- `.github/` 템플릿: 0건(이슈 폼은 'MoAI: …' 표기) / 루트 `README.md`: 없음

## 변경 목록 (커밋 8393a7f72 명칭 통일 7줄 + cda99c15 배너 + 본 수리 문서)

| 파일:줄 | 현재 | 변경 후 |
|---|---|---|
| `src/claim.rs:1194` | processes that bypass moai-cli | processes that bypass MoAI — 사용자 메시지(테스트·픽스처 고정 없음 확인) |
| `docs/protocol-v1.md:3` | moai-cli의 stdin/stdout 계약 | MoAI의 stdin/stdout 계약 |
| `docs/protocol-v1.md:53` | moai-cli가 직접 내는 메시지 | MoAI가 직접 내는 메시지 |
| `docs/local-rc1-kit.md:1` | …에서 moai-cli rc.1 써 보기 | …에서 MoAI rc.1 써 보기 |
| `docs/release-attestation.md:1` | # moai-cli 릴리스 attestation… | # MoAI 릴리스 attestation… — `tests/release_suite.rs` needle과 무관한 줄 확인 |
| `docs/release-notes/2026.11.0-rc.1.md:1` | # moai-cli … 릴리스 노트 | # MoAI … 릴리스 노트(본문 9행의 산출물명 `moai-cli`는 유지) |
| `docs/benchmarks/baseline-2026-11-09.md:3` | moai-cli 내부 연산 | MoAI 내부 연산 |
| `src/banner.rs`·`tests/banner_golden.rs`·`tests/fixtures/banner/*.golden` | (신설, #2363 추가 배차) | 배너 문구는 사용자 지정 `◆ MoAI-ADK …` 그대로 — 개명 예외이자 요구 자체 |
| `docs/product-naming-inventory.md` | (신설, 본 문서) | 전수 표 — 본 문서가 패턴 문자열을 인용해 자기매치 4건을 만든다(위 요약에 분류) |

## 유지 사유 (버킷 — 측정 트리 기준, 합계 1,628)

| 버킷 | 매치 | 사유 |
|---|---|---|
| 보호 파일(PR #2360 소유, 미병합) | 481 | `docs/rust-product-path.md` 182, `src/main.rs` 177, `tests/product_cli.rs` 97, `docs/user-entry-draft.md` 14, `src/session.rs` 1, launcher 시험 등 — 소유 PR이 open인 동안 편집하지 않음(단, 배너 call-site 2곳은 리더 추가 배차 예외를 PR 본문에 명시) |
| 시험·픽스처(tests/** 중 보호 파일 제외) | 550 | 단언·기록 데이터 — Issue 제목 인용, 프로토콜 fixture, go-corpus 등 |
| 기록물(decisions·design-review·e2e-receipts) | 283 | 결정 원장·설계 검토 스냅샷·측정 receipt — 이력 기록이므로 재작성하지 않음 |
| docs(살아있는 문서, 보호·기록물·본 문서 제외) | 184 | 경로·명령·tag 참조가 대부분; 변경 목록 2~7행이 이 버킷에서 나온 전부 |
| src(보호 파일 제외) | 111 | `.moai-cli` 경로·명령·문서 주석 — 사용자 메시지는 변경 목록 1행뿐 |
| 기타(지시 파일·스키마·gitignore·Cargo.toml) | 28 | AGENTS.local.md 22(설치 경로·명령 맥락), schemas 3(`moai_cli_sha` 필드), .gitignore 2, Cargo.toml 1(`name = "moai-cli"`) |
| skills/**(#2297 대기) | 7 | base의 구경로 스킬 파일이 이미 `moai-cli claim-verify` 등 명령 문자열을 포함 — PR #2297(감사 PASS, 병합 대기)이 경로를 moai-cli로 통일 |
| scripts·tools | 16 | 개발 도구 주석·픽스처 문자열 — 사용자 대면 아님 |
| 본 문서(product-naming-inventory.md) | 24 | 측정 방법론·버킷 기술 문자열 — 재검색 시 자기매치로 정상 등장 |

식별자 유지 공통: `~/.moai-cli/`·`.moai-cli/` 경로, `moai-cli <subcommand>` 명령, `pub const PRODUCT = "moai-cli"`(`src/install.rs:46`), `moai-cli-v*` tag/ruleset(Issue 범위 밖), 생성 훅 명령 문자열(`src/hooks.rs` — 출력 계약).

## 전수 파일별 집계 (방법 1, 189파일)

| 파일 | 매치 | 분류 |
|---|---|---|
| `docs/rust-product-path.md` | 183 | PROTECTED(2360-open) |
| `src/main.rs` | 182 | PROTECTED(2360-open) |
| `tests/product_cli.rs` | 97 | PROTECTED(2360-open) |
| `docs/design-review/fixes/top-5490c4ee.md` | 42 | RECORDS |
| `docs/decisions-2026-11-08.md` | 40 | RECORDS |
| `tests/install_cli.rs` | 35 | TESTS |
| `tests/migration_fixture.rs` | 31 | TESTS |
| `tests/layout_cli.rs` | 31 | TESTS |
| `tests/hooks_cli.rs` | 30 | TESTS |
| `docs/release-checklist.md` | 26 | DOCS_REST |
| `docs/design-review/review-group3-install-host-plan-run-release.md` | 25 | RECORDS |
| `docs/design-review/review-group1-rc1-plan.md` | 25 | RECORDS |
| `docs/product-naming-inventory.md` | 24 | DOCS_REST |
| `docs/local-rc1-kit.md` | 23 | DOCS_REST |
| `tests/run_verify.rs` | 22 | TESTS |
| `AGENTS.local.md` | 22 | OTHER |
| `tests/release_suite.rs` | 21 | TESTS |
| `tests/handoff_cli.rs` | 20 | TESTS |
| `tests/claim_cli.rs` | 20 | TESTS |
| `docs/release-attestation.md` | 20 | DOCS_REST |
| `tests/archive_privacy_suite.rs` | 18 | TESTS |
| `tests/archive_move_fixture_test.rs` | 17 | TESTS |
| `tests/apply_receipt.rs` | 17 | TESTS |
| `src/layout.rs` | 15 | SRC_REST |
| `docs/design-review/review-group2-core-brain-sync-owner.md` | 15 | RECORDS |
| `tests/generation_own_reference.rs` | 14 | TESTS |
| `docs/user-entry-draft.md` | 14 | PROTECTED(2360-open) |
| `docs/hooks-generation.md` | 14 | DOCS_REST |
| `docs/design-review/synthesis-improvement-report-2026-11-09.md` | 14 | RECORDS |
| `docs/baseline-rc1.md` | 14 | DOCS_REST |
| `tests/forge_report_cli.rs` | 13 | TESTS |
| `src/install.rs` | 13 | SRC_REST |
| `docs/support-matrix.md` | 13 | DOCS_REST |
| `docs/design-review/fixes/group3-487fa473.md` | 13 | RECORDS |
| `docs/design-review/fixes/group3-25fa9b23.md` | 13 | RECORDS |
| `tests/claim_verify_cli.rs` | 12 | TESTS |
| `src/agents.rs` | 12 | SRC_REST |
| `docs/release-notes/2026.11.0-rc.1.md` | 12 | DOCS_REST |
| `docs/design-review/gap-analysis-rc1-vs-design-2026-11-08.md` | 12 | RECORDS |
| `tests/archive_plan_fixture_test.rs` | 11 | TESTS |
| `tests/worktree_cli.rs` | 9 | TESTS |
| `tests/hook_mode_cli.rs` | 9 | TESTS |
| `tests/forge_cli.rs` | 9 | TESTS |
| `tests/context_cli.rs` | 9 | TESTS |
| `docs/home-layout-p09.md` | 9 | DOCS_REST |
| `docs/e2e-receipts/receipts.jsonl` | 9 | RECORDS |
| `docs/design-review/review-page_5490c4ee-2026-11-08.md` | 9 | RECORDS |
| `docs/design-review/fixes/plan-43faf222.md` | 9 | RECORDS |
| `docs/design-review/fixes/group3-83f162ef.md` | 9 | RECORDS |
| `docs/design-review/appendix/paths.md` | 9 | RECORDS |
| `tests/todo_promotion_cli.rs` | 8 | TESTS — 측정 뒤 삭제됨(`b0c590c7c9`, PR #2406 로컬 TODO 승격 제거; #2405 확인) |
| `tests/lane_handoff_cli.rs` | 8 | TESTS |
| `docs/design-review/autonomy-modes-spec.md` | 8 | RECORDS |
| `tools/host_live_harness.py` | 7 | SCRIPTSTOOLS |
| `tests/workflow_e2e.rs` | 7 | TESTS |
| `tests/profile_apply_engine.rs` | 7 | TESTS |
| `tests/queue_cli.rs` | 6 | TESTS |
| `tests/leader_cli.rs` | 6 | TESTS |
| `tests/journal_lock_busy_cli.rs` | 6 | TESTS |
| `tests/generation_receipt_order.rs` | 6 | TESTS |
| `tests/audit_admission.rs` | 6 | TESTS |
| `src/hooks.rs` | 6 | SRC_REST |
| `src/generation.rs` | 6 | SRC_REST |
| `docs/webhook-ingest-research-2026-11-09.md` | 6 | DOCS_REST |
| `tests/update_known_good.rs` | 5 | TESTS |
| `tests/update_cli.rs` | 5 | TESTS |
| `tests/github_profile_cli.rs` | 5 | TESTS |
| `tests/github_cli.rs` | 5 | TESTS |
| `tests/generation_manifest_single_write.rs` | 5 | TESTS |
| `tests/e2e_host_fixture.rs` | 5 | TESTS |
| `tests/claude_settings_cli.rs` | 5 | TESTS |
| `tests/claim_record_cli.rs` | 5 | TESTS |
| `src/claim.rs` | 5 | SRC_REST |
| `scripts/benchmark.sh` | 5 | SCRIPTSTOOLS |
| `docs/host-support-matrix.md` | 5 | DOCS_REST |
| `docs/github-plan-proposal.md` | 5 | DOCS_REST |
| `docs/github-inventory-rc1.md` | 5 | DOCS_REST |
| `docs/claude-adapter-contract.md` | 5 | DOCS_REST |
| `tools/local-rc1-isolation.sh` | 4 | SCRIPTSTOOLS |
| `tests/worktree_branch_cli.rs` | 4 | TESTS |
| `tests/update_known_good_replace.rs` | 4 | TESTS |
| `tests/secret_shapes_content.rs` | 4 | TESTS |
| `tests/recovery_cli.rs` | 4 | TESTS |
| `tests/codex_config.rs` | 4 | TESTS |
| `tests/codex_adapter.rs` | 4 | TESTS |
| `tests/claude_recovery_test.py` | 4 | TESTS |
| `tests/claude_preview_test.py` | 4 | TESTS |
| `tests/banner_golden.rs` | 4 | TESTS |
| `tests/apply_recover_partial.rs` | 4 | TESTS |
| `tests/agents_apply_linkage.rs` | 4 | TESTS |
| `src/update.rs` | 4 | SRC_REST |
| `src/sources.rs` | 4 | SRC_REST |
| `src/profile.rs` | 4 | SRC_REST |
| `src/archive.rs` | 4 | SRC_REST |
| `src/apply.rs` | 4 | SRC_REST |
| `docs/session-hub-research-2026-11-08.md` | 4 | DOCS_REST |
| `docs/p07-decision.md` | 4 | DOCS_REST |
| `tests/sources_cli.rs` | 3 | TESTS |
| `tests/plan_admission.rs` | 3 | TESTS |
| `tests/launcher_roleless.rs` | 3 | PROTECTED(2360-open) |
| `tests/generation_ref_slash.rs` | 3 | TESTS |
| `tests/e2e/host-codex.sh` | 3 | TESTS |
| `tests/e2e/host-claude.sh` | 3 | TESTS |
| `tests/archive_move_cli.rs` | 3 | TESTS |
| `src/self_update.rs` | 3 | SRC_REST |
| `src/project_skills.rs` | 3 | SRC_REST |
| `src/lib.rs` | 3 | SRC_REST |
| `src/forge/report.rs` | 3 | SRC_REST |
| `src/claude.rs` | 3 | SRC_REST |
| `skills/moai/workflows/mode-agent.md` | 3 | SKILLS(2297-pending) — 현재 `skills/moai-cli/workflows/mode-agent.md`(`36739b90da`, PR #2297; #2405 확인) |
| `schemas/e2e-receipt.schema.json` | 3 | OTHER |
| `docs/host-live-session-checklist.md` | 3 | DOCS_REST |
| `docs/design-review/statusline-rust-stage1-2026-11-10.md` | 3 | RECORDS |
| `docs/design-review/personal-coding-core-2026-11-09.md` | 3 | RECORDS |
| `docs/design-review/fixes/group3-e0874134.md` | 3 | RECORDS |
| `docs/design-review/fixes/group3-3d917728.md` | 3 | RECORDS |
| `docs/design-review/fixes/group2-0d4538a27fb88191a19ba261bd5a3442.md` | 3 | RECORDS |
| `docs/design-review/assignments-schema.md` | 3 | RECORDS |
| `docs/codex-adapter-contract.md` | 3 | DOCS_REST |
| `tests/update_unmanaged_failclosed.rs` | 2 | TESTS |
| `tests/struct_constructors_test.rs` | 2 | TESTS |
| `tests/spec_contract.rs` | 2 | TESTS |
| `tests/fixtures/protocol/v1/current/21-baseline-inventory-ok.json` | 2 | TESTS |
| `tests/fixtures/protocol/v1/current/17-agents-apply-ok.json` | 2 | TESTS |
| `tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json` | 2 | TESTS |
| `tests/fixtures/github/issues-open.json` | 2 | TESTS |
| `tests/fixtures/atlas-tag/go-corpus.txt` | 2 | TESTS |
| `tests/claude_transaction_boundary_test.py` | 2 | TESTS |
| `src/worktree.rs` | 2 | SRC_REST |
| `src/migrate.rs` | 2 | SRC_REST |
| `src/github_templates.rs` | 2 | SRC_REST |
| `src/forge/outbox.rs` | 2 | SRC_REST |
| `docs/rc-isolation-decision.md` | 2 | DOCS_REST |
| `docs/pr-contract.md` | 2 | DOCS_REST |
| `docs/lane-loop-contract.md` | 2 | DOCS_REST |
| `docs/e2e-receipts/T1-cargo-test-output.txt` | 2 | RECORDS |
| `docs/e2e-matrix.md` | 2 | DOCS_REST |
| `docs/design-review/review-group4-past-designs.md` | 2 | RECORDS |
| `docs/design-review/mcp-design-prep-2026-11-09.md` | 2 | RECORDS |
| `docs/design-review/host-session-events-2026-11-09.md` | 2 | RECORDS |
| `docs/design-review/fixes/group3-66acadc2.md` | 2 | RECORDS |
| `docs/design-review/fixes/group2-136c8203bf2881919eb1f543f72f87ea.md` | 2 | RECORDS |
| `.gitignore` | 2 | OTHER |
| `tests/sync_marker_form.rs` | 1 | TESTS |
| `tests/support_matrix_test.py` | 1 | TESTS |
| `tests/promote_digest_order.rs` | 1 | TESTS — 측정 뒤 삭제됨(`b0c590c7c9`, PR #2406 로컬 TODO 승격 제거; #2405 확인) |
| `tests/plan_input_limit_cli.rs` | 1 | TESTS |
| `tests/parity_cli.rs` | 1 | TESTS |
| `tests/ontology_schema.rs` | 1 | TESTS |
| `tests/omp_adapter.rs` | 1 | TESTS — 측정 뒤 삭제됨(`571c2e6476`, PR #2381 OMP 제거; #2405 확인) |
| `tests/launcher_permissions_setting.rs` | 1 | PROTECTED(2360-open) |
| `tests/host_live_harness_test.py` | 1 | TESTS |
| `tests/gitlab_forge_contract.rs` | 1 | TESTS |
| `tests/github_inventory_test.py` | 1 | TESTS |
| `tests/github_inventory_collect.py` | 1 | TESTS |
| `tests/git_env_scrub.rs` | 1 | TESTS |
| `tests/forge_links_cli.rs` | 1 | TESTS |
| `tests/forge_audit_status_cli.rs` | 1 | TESTS |
| `tests/flow_cli.rs` | 1 | TESTS |
| `tests/fixtures/protocol/v1/current/16-agents-plan-create.json` | 1 | TESTS |
| `tests/fixtures/protocol/v1/current/09-project-apply-created.json` | 1 | TESTS |
| `tests/e2e_receipt_not_run_test.py` | 1 | TESTS |
| `tests/common/mod.rs` | 1 | TESTS |
| `tests/claude_update_test.py` | 1 | TESTS |
| `tests/baseline_git_env.rs` | 1 | TESTS |
| `tests/atlas_tag_cli.rs` | 1 | TESTS |
| `src/session.rs` | 1 | PROTECTED(2360-open) |
| `src/scope.rs` | 1 | SRC_REST |
| `src/recovery.rs` | 1 | SRC_REST |
| `src/leader.rs` | 1 | SRC_REST |
| `src/lane_handoff.rs` | 1 | SRC_REST |
| `src/identity.rs` | 1 | SRC_REST |
| `src/hosts/codex_config.rs` | 1 | SRC_REST |
| `src/forge/mod.rs` | 1 | SRC_REST |
| `src/forge/gitlab.rs` | 1 | SRC_REST |
| `src/decision.rs` | 1 | SRC_REST |
| `src/baseline.rs` | 1 | SRC_REST |
| `src/banner.rs` | 1 | SRC_REST |
| `skills/moai/workflows/update.md` | 1 | SKILLS(2297-pending) — 현재 `skills/moai-cli/workflows/update.md`(PR #2297; #2405 확인) |
| `skills/moai/workflows/sync.md` | 1 | SKILLS(2297-pending) — 현재 `skills/moai-cli/workflows/sync.md`(PR #2297; #2405 확인) |
| `skills/moai/workflows/setup.md` | 1 | SKILLS(2297-pending) — 현재 `skills/moai-cli/workflows/setup.md`(PR #2297; #2405 확인) |
| `skills/moai/references/source-manifest.json` | 1 | SKILLS(2297-pending) — 현재 `skills/moai-cli/references/source-manifest.json`(PR #2297; #2405 확인) |
| `docs/rc-release-review.md` | 1 | DOCS_REST |
| `docs/p02-decision.md` | 1 | DOCS_REST |
| `docs/design-review/flow-control-2026-11-09.md` | 1 | RECORDS |
| `docs/code-atlas-name-collision.md` | 1 | DOCS_REST |
| `docs/claude-preview.md` | 1 | DOCS_REST |
| `docs/claude-owned-update.md` | 1 | DOCS_REST |
| `Cargo.toml` | 1 | OTHER |

## 변경 후 재검증 (측정 트리 기준)

- 변경한 7줄은 head 트리에서 'MoAI' 표기로 확인(7건 grep 실측, 2026-11-10T10:0xZ) — 대상 표기 0. 남은 항목은 위 버킷 사유로 전부 커버.
- AC2(바이너리명·설치 경로·상태 폴더·PATH 불변): 커밋 diff에 `Cargo.toml`·`PRODUCT` 상수·경로 리터럴 변화 없음 — 변경 파일은 명칭 6+문서 1, 배너 5, 본 문서.
- AC3(envelope·operation 계약 불변): diff 범위에 `tests/fixtures/protocol/v1/**`·operation 이름·생성 훅 명령 문자열 없음.

## 부록 — 측정 경과 (각 시점의 트리가 다르다)

| 시점(UTC) | 트리 | 총매치 | 비고 |
|---|---|---|---|
| 2026-11-10T05:10Z | 8c57326151(개명 커밋 전 base) | 1,556/177파일 | 이전 판의 baseline |
| 2026-11-10T05:4xZ | 8393a7f72(명칭 커밋, 병합 전) | 1,572/176파일 | −7 편집 표면 +23 본 문서(구판) |
| 2026-11-10T07:5xZ | e0f1ef3983(배너+55f383eaa8 병합) | 1,603 | 감사자 실측 — 구판 문서 수치와의 어긋남이 FAIL 사유 |
| 2026-11-10T10:03Z | fad15f59d6(0c13195747 병합, 첫 수리판) | 1,628/183파일(git grep)·1,627(BSD grep) | base 추가 이동(e0f2fb0854)으로 대체 |
| 2026-11-10T10:09Z | **346f3cf3e7**(e0f2fb0854 병합, **본 판**) | **1,684/189파일**(git grep)·1,683(BSD grep) | 본 문서 수치 |

병합이 base 측 `moai-cli` 문자열을 계속 늘린다(#2345·#2335·#2344 문서와 git_env_scrub 등 신규 시험) — 총매치의 증가 자체는 제품명 위반이 아니라 위 지시·시험 문서의 명령 참조이며, 파일별 분류표가 이를 전부 사유와 함께 분류한다.
