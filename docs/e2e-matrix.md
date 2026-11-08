# E2E matrix (08-C #2014, RC1)

이 문서는 08-C E2E receipt(`schemas/e2e-receipt.schema.json`, REQ-08C-01)의 조합별 상태를 기록한다. 이 표만이 08-D #2015 지원표(`docs/host-support-matrix.md`)의 E2E 입력이다. 열·형식은 지원표 소유 Issue 02-C #1993이 정한 것을 따른다(host | surface | item | capability_status | unknown_class | 내용 | source); receipt 필드(os·arch·host·host_exact_version·mode·skill_digest·moai_cli_sha·result)는 각 행의 receipt JSON으로 남긴다. 실행하지 않은 조합은 숨기지 않고 그대로 적는다.

## 상태 어휘

`capability_status`는 지원표(02-C #1993)와 같은 모델을 따른다. `certified`(실제 앱 E2E)만 "지원"으로 표기할 수 있다.

| 값 | 이 표에서의 뜻 |
| --- | --- |
| `probed` | T1 결정적 fixture가 이 조합에서 실측 receipt를 남겼다 |
| `unknown` | 아직 실측하지 않았다(`unknown_class` = `시험 대기`) |

P12 결정(리더 범위 댓글 2026-11-08, 사용자 결정 적용): RC1의 조합은 **개발 Mac 단일**(aarch64-apple-darwin, macOS)이다. 다른 OS·arch·cloud/worktree mode는 전부 not_run으로 남고 OS matrix workflow는 RC1에서 만들지 않는다(08-D 소유).

## T1 — 결정적 fixture (`tests/e2e_host_fixture.rs`)

실행 명령(실제 시험 파일·시험명과 일치 — AC 명령 인용 전 확인 의무):

```sh
CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target \
HOME=$(mktemp -d) cargo test --offline --locked --test e2e_host_fixture -- --nocapture
```

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| none (T1 fixture) | cli | e2e_setup_update_interrupt_restore_preserves_user_edits (AC-08C-1) | probed | | 임시 HOME·프로젝트에서 project-plan/apply(project-plan, project-apply, 01-B) + update-apply v1(04-B 양 host 설치) 후, 실제 중단(읽기 전용 projection 디렉터리로 apply가 트랜잭션 중간 실패) → recovery-plan/recovery-apply가 known-good에서 대체 파일 복구, 사용자 편집(routes.json) 바이트 보존, 성공 host(codex)는 committed v2 유지, 가짜 Go 상태 digest 불변 | cargo test 실행 기록(아래 receipt) |
| none (T1 fixture) | cli | e2e_digest_mismatch_is_fail_closed (AC-08C-2) | probed | | plan digest 계산 뒤 stage 바이트 변조 → update-apply 거절, HOME 트리 전체 바이트 동일 | cargo test 실행 기록 |
| none (T1 fixture) | cli | e2e_second_update_is_diff_zero (HA-02) | probed | | 동일 stage 재요청: 모든 action이 unchanged/preserved, apply 후 HOME digest 0 diff | cargo test 실행 기록 |
| none (T1 fixture) | cli | e2e_never_writes_outside_temp_home (AC-08C-7) | probed | | sandbox-exec deny-home(실제 HOME 쓰기 전면 거절) 아래에서 update-apply 성공 — HOME 리다이렉션이 사라져도 사용자 상태를 못 쓴다. 가짜 Go 상태 digest 불변. 실제 `~/.moai` 대량 해시 대조는 동시 변동으로 신뢰 불가하여 이 방식으로 갈음(리더 허용 방식) | cargo test 실행 기록 |
| none (T1 fixture) | cli | e2e_clean_clone_carries_portable_read_plan_assets (AC-08C-8, T1 절반) | probed | | binary 없는 clean clone에 skills/moai 전체(SKILL.md frontmatter·routes.json·routes가 참조하는 workflow 전부)가 존재 — READ/PLAN 포터블 자산 완결 | cargo test 실행 기록 |

T1 receipt 필드(위 5행 공통): os=Darwin, arch=aarch64, host=none, mode=local, skill_digest=SKILL.md sha256(실행 시점 기록), moai_cli_sha=빌드 바이너리 sha256, ran_at=실행 시각 실측. raw 결과는 `cargo test` 출력 그대로가 detail이며 자격증명이 없다.

## T2 — 실제 host 로딩 (`tests/e2e/host-claude.sh`, `tests/e2e/host-codex.sh`)

레인은 실행하지 않는다(리더 범위: 기존 인증만 쓰는 운영자 1회 실측, 체크아웃 밖). 명령과 기대 결과는 스크립트에 receipt 형식으로 준비돼 있고, 운영자 실행 전까지 전부 not_run이다. 스크립트는 자격증명 값의 존재만 확인하고 값은 어떤 파일에도 쓰지 않는다.

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| Claude Code | cli | skill_static_validation (AC-08C-5) | probed | | `claude plugin validate <tmp>/.claude/skills` — 자격증명 불필요, 레인 실행 가능. 실행은 AC 명령 절에서 수행하고 receipt로 기록 | 이 문서의 AC 명령 절 |
| Claude Code | cli | explicit_invocation_setup_route (AC-08C-3) | unknown | 시험 대기 | `claude --bare -p '/moai …' --add-dir <tmp> --output-format json --permission-mode dontAsk --permission-prompts none` — 기대: `.result`가 문자열로 SETUP route 진입과 로드된 skill을 보고(JSON 구조만 단언, 자연어 단언 없음). not_run — 운영자 실측 대기 | tests/e2e/host-claude.sh |
| Claude Code | cli | permission_denial_blocked (AC-08C-6) | unknown | 시험 대기 | 동일 옵션으로 apply 요청 → `permission_denials` 기록 + 대상 트리 shasum 전후 동일. not_run — 운영자 실측 대기 | tests/e2e/host-claude.sh |
| Codex | cli | explicit_invocation_setup_route (AC-08C-4) | unknown | 시험 대기 | `codex exec --json --ephemeral --ignore-user-config --sandbox read-only -o last.md '$moai …'` — 기대: JSONL에 `"type":"turn.completed"` 존재 + last message 비어 있지 않음. not_run — 운영자 실측 대기 | tests/e2e/host-codex.sh |
| Claude Code | cli | clean_clone_read_plan (AC-08C-8, T2 절반) | unknown | 시험 대기 | binary 없는 clean clone에서 `claude --bare -p '/moai 현재 저장소 읽기' --output-format json` — not_run — 운영자 실측 대기(T1 절반은 위 표 probed) | 이 문서의 AC 명령 절 |

## OS·mode 조합 (P12)

| os | arch | host | mode | result | 비고 |
| --- | --- | --- | --- | --- | --- |
| macOS (Darwin) | aarch64 | none (T1) | local | T1 표 참조 — 5시험 probed | RC1 유일 실측 조합 |
| macOS (Darwin) | aarch64 | Claude Code | local | AC-08C-3·6 not_run | 운영자 실측 대기 |
| macOS (Darwin) | aarch64 | Codex | local | AC-08C-4 not_run | 운영자 실측 대기 |
| ubuntu-24.04 / windows-2025 / arm 변형 | — | — | local | not_run | P12 HOLD — RC1에서 workflow 만들지 않음(08-D에 not_run 반영) |
| — | — | — | worktree / cloud | not_run | RC1 범위 밖 |

## AC 명령 절 (receipt 인용용 실측 명령)

AC-08C-5(레인 실행 가능 정적 검증)는 임시 프로젝트에 skill을 복사한 뒤 로컬 검증을 돌린다 — 자격증명 불필요:

```sh
TMP=$(mktemp -d /tmp/e2e-ac5-XXXXXX) && mkdir -p "$TMP/.claude/skills" \
  && cp -R skills/moai "$TMP/.claude/skills/moai" \
  && claude plugin validate "$TMP/.claude/skills"; echo "exit=$?"; rm -rf "$TMP"
```

AC-08C-7의 Go 바이너리 해시는 읽기만 한다(실행 금지): `shasum -a 256 "$(command -v moai)"` — 실행 전후 동일.

AC-08C-9(독립 AUDIT)는 리더 통합 게이트가 진행한다.
