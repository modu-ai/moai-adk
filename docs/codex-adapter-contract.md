# Codex adapter 계약 (02-C #1993)

공통 skill package(`skills/moai/`, 02-A #1991; #2286에서 skill 이름 `moai`로 확정, 실행 파일은 `moai-cli`)를 OpenAI Codex(Desktop 앱, CLI, IDE 확장)에 연결하는 adapter의 계약이다. RC1에서는 projection을 계획하고 설치 상태를 진단하는 데까지만 했다. rc.2 I4 #2241이 계획한 projection을 호출자가 정한 대상에 설치하는 `codex.projection_apply`를 더했다(아래 Operation 절). 설치 위치는 여전히 정하지 않는다. 02-C 범위의 확인일은 2026-11-08이고, I4가 다시 확인한 항목은 2026-11-09(#2286 개명 갱신 동일일)로 따로 적는다. 지원표 행은 `docs/host-support-matrix.md`의 Codex 절에 있다.

## 결정 (리더 RC1 임시, 2026-11-08, Issue 본문 r30)

| 항목 | 결정 | 영향 |
| --- | --- | --- |
| TOML 읽기 | 파서를 쓰지 않는다. TOML 읽기·쓰기 표면과 의존성 결정은 02-B #1992 소유다(rc.2 #2240에서 쓰기용 파서 도입, 아래 "config.toml 생성") | doctor는 `config.toml`을 읽지 않는다. trust, inline `[hooks]`, `tui.status_line`은 `not_evaluated`로 보고하고 `trusted`라고 쓰지 않는다. fallback 이름과 max-bytes는 인자로 받는다. doctor 읽기 승격은 후속 Issue에서 한다 |
| hook | projection은 생성하지 않는다. 생성은 `hooks-plan`/`hooks-apply --host codex`(rc.2 I1 #2238)가 맡는다 | projection은 `.codex/hooks.json`도, inline `[hooks]`도 만들지 않는다. `hooks-apply`가 만든 `.codex/hooks.json`은 Codex hash trust 검토(`/hooks`)를 거쳐야 실행된다. 근거: `docs/hooks-generation.md` |
| projection 위치 | `--target`을 꼭 받는다. 기본값은 없다 | `~/.agents/skills/moai`(사용자 홈)와 `<repo>/.agents/skills/moai`(project) 중 고르는 일은 HOLD로 남는다. 두 후보는 이 문서에만 적는다 |
| E2E | 레인에서는 not_run | Desktop `$moai`는 GUI가 필요하다. CLI `$moai`는 실제 호출이 OpenAI로 prompt를 보내 외부 전송과 비용이 생긴다. 운영자 실측 여부는 사용자 결정이다 |
| 시험 위치 | `tests/codex_adapter.rs` 신규 | `tests/product_cli.rs` 꼬리 충돌을 피한다 |

not_run은 Desktop 발견·호출이 실패했다는 뜻이 아니다. 그래서 HA-10의 "Desktop 실패 → RC blocked" 조건에 해당하지 않는다. 실측해서 실패가 나오면 그때 RC를 blocked로 보고하고, Claude-only로 축소하지 않는다.

## 출처와 원문 인용

| 페이지 | 인용 |
| --- | --- |
| https://learn.chatgpt.com/docs/build-skills | "For repositories, Codex scans `.agents/skills` in every directory from your current working directory up to the repository root." / "The `SKILL.md` file must include `name` and `description`." / "If two skills share the same `name`, Codex doesn’t merge them; both can appear in skill selectors." / "Codex supports symlinked skill folders and follows the symlink target when scanning these locations." / "Codex detects skill changes automatically. If an update doesn’t appear, restart Codex." / "`allow_implicit_invocation` (default: `true`)" / "In Codex CLI or the IDE extension, run `/skills` or type `$` to mention a skill." |
| https://learn.chatgpt.com/docs/agent-configuration/agents-md | "Codex reads `AGENTS.override.md` if it exists. Otherwise, Codex reads `AGENTS.md`." / "Codex uses only the first non-empty file at this level." / "Starting at the project root (typically the Git root), Codex walks down to your current working directory." / "Codex includes at most one file per directory." / "Codex concatenates files from the root down, joining them with blank lines." / "Codex skips empty files" / Troubleshoot: "Ensure instruction files contain content; Codex ignores empty files." / "stops adding files once the combined size reaches the limit defined by `project_doc_max_bytes` (32 KiB by default)." / "Files closer to your current directory override earlier guidance because they appear later in the combined prompt." |
| https://learn.chatgpt.com/docs/config-file/config-reference | `projects.<path>.trust_level`: "Mark a project or worktree as trusted or untrusted (`"trusted"` \| `"untrusted"`)." / "Untrusted projects skip project-scoped `.codex/` layers, including project-local config, hooks, and rules." / `tui.status_line`: "Ordered list of TUI footer status-line item identifiers. `null` disables the status line." (ID 목록은 없음) / `project_doc_max_bytes`: "Maximum bytes read from `AGENTS.md` when building project instructions." (기본값 표기 없음) / `personality`: "Deprecated compatibility setting." |
| https://learn.chatgpt.com/docs/hooks | "If a single layer contains both `hooks.json` and inline `[hooks]`, Codex merges them and warns at startup." / "Higher-precedence config layers don't replace lower-precedence hooks." / "Non-managed hooks must be reviewed and trusted before they run." |
| https://learn.chatgpt.com/docs/app/slash-commands | "You can also explicitly invoke skills by typing `$` in the chat composer." / "Enabled skills also appear in the slash command list." (명령 표에 `/skills` 없음) |

### `agents/openai.yaml` 스키마 근거 (I4 #2241, 2026-11-09 재확인)

https://learn.chatgpt.com/docs/build-skills 를 2026-11-09에 다시 읽었다(HTTP 200, 원문).

- 위치와 목적: skill 폴더 트리에서 `agents/`는 `SKILL.md` 옆에 있고, `openai.yaml`에 "Optional: appearance and dependencies"라고 표시돼 있다. 목적은 "Add `agents/openai.yaml` to configure UI metadata in the ChatGPT desktop app, to set invocation policy, and to declare tool dependencies"다.
- 키: 페이지에는 따로 된 필드 표가 없고 예시 블록 하나가 있다. `interface`의 `display_name`·`short_description`·`icon_small`·`icon_large`·`brand_color`·`default_prompt`, `policy`의 `allow_implicit_invocation`, `dependencies.tools`(항목마다 `type`·`value`·`description`·`transport`·`url`)다.
- 기본값: "`allow_implicit_invocation` (default: `true`)" / "When `false`, Codex won't implicitly invoke the skill based on user prompt; explicit `$skill` invocation still works."
- 필수 여부: 절 제목이 "Optional metadata"이고, 필수 키는 적혀 있지 않다. 필수 필드는 SKILL.md의 것뿐이다: "The `SKILL.md` file must include `name` and `description`."
- 위치 후보: REPO `$CWD/.agents/skills`·`$CWD/../.agents/skills`·`$REPO_ROOT/.agents/skills`, USER `$HOME/.agents/skills`, ADMIN `/etc/codex/skills`. "Codex detects skill changes automatically. If an update doesn't appear, restart Codex."

MoAI가 쓰는 키는 이 중 `interface.display_name`·`short_description`·`default_prompt`와, 플래그를 줬을 때의 `policy.allow_implicit_invocation`뿐이다. 아이콘·색상·`dependencies`는 쓰지 않는다. 02-C 결정(REQ-02C-02)은 바뀌지 않았다.

원문을 확인하지 못한 항목도 있다. hooks 페이지의 handler 종류(`command`·`mcp_tool` 지원, `prompt`·`agent` 무시), "exit 0과 빈 출력은 계속", PreToolUse 오류·timeout 때의 동작, `features.hooks` 기본값이다. 긴 페이지라 이 부분은 요약만 얻었다. 그래서 표와 코드에 쓰지 않았고, Issue 본문의 해당 근거는 이번 조회로 다시 확인되지 않은 상태다.

로컬 관측은 `codex --version` → `codex-cli 0.163.0`과 `codex --help`의 subcommand 목록(`exec`, `doctor`, `features`, `plugin` 등)뿐이다. 둘 다 모델을 호출하지 않는다.

## Operation

세 operation 모두 01-C envelope(`{protocol_version, operation, ok, result|error}`)을 쓴다. 쓰는 것은 `codex.projection_apply`뿐이다. `result` 키는 `docs/rust-product-path.md`의 operation 표에 있다.

### `codex.projection_plan` (`codex-projection-plan PACKAGE --target DIR`)

- `--target`: 절대 경로여야 하고, 마지막 이름이 `moai`여야 하고, package 바깥이어야 한다. 없으면 `usage`, 조건에 어긋나면 `invalid_input`이다.
- 파일 계획: package의 모든 파일을 sha256이 같은 복사(`action: copy`)로 계획한다. 그래서 `SKILL.md`는 02-A canonical body와 바이트가 같다(REQ-02C-01). `skill_md_identical`은 고정값이 아니라, 계획한 `SKILL.md`의 sha256을 package에서 다시 읽은 바이트와 비교한 결과다. 계획 전에 `SKILL.md` frontmatter가 `name: moai`와 `description` 두 키만 갖는지 확인하고, 다른 키가 있거나 둘 중 하나가 없으면 `invalid_input`으로 거부한다. package 안에 `agents/`가 있으면 거부한다. `agents/openai.yaml`은 adapter metadata이지 package 파일이 아니기 때문이다.
- `agents/openai.yaml`: `interface.display_name`·`short_description`·`default_prompt`만 쓴다. `policy.allow_implicit_invocation`은 `--allow-implicit-invocation`을 줬을 때만 쓴다. 주지 않으면 키가 없어 Codex 기본값(`true`)이 그대로 유지된다(REQ-02C-02). 값은 큰따옴표 YAML 문자열로 escape한다(PyYAML `safe_load`로 다시 읽어 확인). 제어 문자가 있거나 256자를 넘으면 거부한다.
- 설정 요청: `--status-line ID`(반복 가능)와 `--disable-bundled-skills`는 `config_changes: {status: "hold", requested, unverified_items, writes: []}`로 돌려준다. 모든 status-line ID는 `unverified_items`에 들어간다(REQ-02C-07). 요청이 없으면 `status: none_requested`다.
- 대상 상태(I4 #2241): 파일마다 `target_action`이 붙는다. 대상에 파일이 없으면 `create`, 같은 바이트면 `unchanged`, 다른 바이트면 `conflict`다. 대상에 있지만 projection이 소유하지 않는 파일은 `preserved`에 적고 건드리지 않는다. 대상이 파일이거나 대상 아래에 symlink가 있으면 `invalid_input`으로 거부한다. 설치 뒤 다시 계획하면 모든 파일이 `unchanged`이므로, 이 계획이 곧 설치 검증이다.
- source manifest(REQ-I4-01): package의 `references/source-manifest.json` `files` 항목과 package 파일을 양쪽으로 대조한다. manifest 자신은 대조하지 않는다. manifest에는 sha256이 접두사 없는 hex로 적혀 있다. 결과 `source_manifest.status`는 `verified`·`mismatch`·`missing`·`unreadable` 중 하나이고, `mismatches`에 경로와 이유("no manifest entry", "sha256 differs from the manifest", "listed in the manifest but absent")를 담는다. plan은 보고만 하고 거부하지 않는다. 거부는 apply가 한다.
- `plan_digest`: 대상 경로, `package_digest`, 파일별 (경로, sha256, `target_action`), `preserved`, manifest 상태를 묶는다. 대상이 바뀌거나 `openai.yaml` 선택이 바뀌면 digest도 바뀐다.
- 결과: `hooks.generated: []`, `files_created: []`, `host_loaded: false`, `execution_verified: false`.

### `codex.projection_apply` (`codex-projection-apply PACKAGE --target DIR --plan-digest D --permission-reference R`, I4 #2241)

- 입력: plan과 같은 package·`--target`·`openai.yaml` 옵션, 그리고 plan의 `plan_digest`와 permission reference다. `--status-line`·`--disable-bundled-skills`는 받지 않는다. 설정 쓰기는 HOLD다(`config_changes.status: hold`, `writes: []`).
- 첫 바이트를 쓰기 전에 다음을 순서대로 확인한다.
  1. permission reference가 비었거나 `null`·`placeholder`이면 `permission_reference_missing`
  2. 새로 계산한 plan의 `plan_digest`가 주어진 값과 다르면 `stale_plan`
  3. `source_manifest.status`가 `verified`가 아니면 `conflict`("source manifest digest mismatch … nothing applied"). REQ-I4-01: digest가 manifest와 다른 복사본은 쓰지 않는다.
  4. `conflict` 파일이 하나라도 있으면 `conflict`("conflict in codex projection: … nothing applied")
- 쓰기: `create` 파일만 `apply::publish_exclusive`로 만든다. 이미 있는 파일은 덮어쓰지 않는다. 쓰기 직전에 package 바이트를 다시 읽어 plan의 sha256과 비교하고, 다르면 `stale_plan`이다. 첫 파일을 만든 뒤 실패하면 `partial`이다.
- 검증: 계획한 모든 파일을 다시 읽어 plan의 sha256과 비교한다. 모두 같으면 `verified: true`다. 하나라도 다르면 `conflict`("codex projection digest mismatch after write")다.
- 재설치: 설치된 대상에 새 plan으로 다시 실행하면 `files_created: []`이고, 파일 바이트와 수정 시각이 바뀌지 않는다.
- 위치: `--target`에 기본값이 없다. 홈 `~/.agents/skills/moai`와 프로젝트 `<repo>/.agents/skills/moai` 중 무엇을 쓸지는 결정 대기(P04 / C5)이고, 결과의 `location`에 그렇게 적는다. 시험은 두 경로 모두 임시 fixture로만 한다.
- 결과: `host_loaded: false`, `execution_verified: false`. Codex가 설치본을 실제로 읽는지는 C6 운영자 실측 몫이다.

### `codex.doctor` (`codex-doctor PROJECT [--cwd DIR] [--codex-home DIR] [--home DIR] [--fallback NAME]... [--max-bytes N]`)

- `PROJECT`는 project root이고, 호출자가 정한다(Codex는 "typically the Git root"라고만 한다). `--cwd`는 그 안이어야 한다(기본은 root). `--codex-home`의 기본은 `$CODEX_HOME`, 없으면 `~/.codex`다. `--home`의 기본은 `$HOME`이다. 모두 읽기만 한다.
- `agents.global`: `--codex-home`의 `AGENTS.override.md` → `AGENTS.md` 중 첫 비어 있지 않은 파일이다.
- `agents.project_chain`과 `included`: root부터 cwd까지 디렉터리마다 `AGENTS.override.md` → `AGENTS.md` → `--fallback` 이름 순으로, 비어 있지 않은 첫 파일 하나를 고른다. 빈 파일(공백만 있는 파일 포함)은 건너뛴다. 근거는 페이지의 "Codex skips empty files"와 Troubleshoot의 "Codex ignores empty files"다. 빈 파일 다음에 같은 디렉터리의 다음 이름으로 넘어가는 것은 **추론**이다. 페이지는 이를 전역 단계("the first non-empty file at this level")에만 명시하고, project 디렉터리에 대해서는 적지 않는다.
- `project_bytes`는 빈 줄로 이은 합이고, 이것을 `max_bytes`와 비교해 `over_budget`과 처음 넘는 파일 `over_budget_at`을 보고한다. 전역 파일이 같은 예산에 들어가는지, 넘는 파일을 자르는지 빼는지는 문서에 없어 `budget_note`로 남긴다.
- `agents_override_created: false`다. 개인 보충용 override를 만들지 않는다(REQ-02C-05).
- `trust`: `{status: "not_evaluated", reason, consequence}`. config.toml을 읽지 않으므로 untrusted여서 생기는 `blocked_by_trust`를 판정할 수 없고, trust 상태를 바꾸지도 않는다(REQ-02C-06).
- `codex_layer`: `.codex/config.toml`과 `.codex/hooks.json`이 있는지만 본다. `inline_hooks`와 `status_line`은 `not_evaluated`다.
- `skills.moai-cli`(key renamed by #2286): cwd → root의 `.agents/skills/moai`, `$HOME/.agents/skills/moai`, `/etc/codex/skills/moai` 중 있는 것을 모두 보여 준다. 둘 이상이면 병합되지 않는다는 경고를 낸다(REQ-02C-04). Go `moai` skill 디렉터리는 읽지 않는다(#2286 REQ-3).
- 결과: `writes: []`, `host_loaded: false`, `execution_verified: false`.

## AC와 시험 (Issue 본문 r30)

| AC | 시험(`cargo test --offline --locked --test codex_adapter <이름>`) |
| --- | --- |
| AGENTS-BUDGET | `codex_agents_chain_plain`, `codex_agents_chain_override`, `codex_agents_chain_fallback`, `codex_agents_chain_over_budget`, `codex_agents_chain_skips_empty_files`(감사 지적으로 추가) |
| TRUST-NO-CLAIM | `codex_doctor_never_claims_trust` |
| CONFIG-HOLD-NO-WRITE | `codex_config_requests_report_hold_without_writes` |
| STATUS-FIELD-PARITY | `codex_statusline_items_are_unverified` |
| HA-07-DUPLICATE(재정의) | `codex_hooks_not_generated` |
| PROJECTION-PLAN-NO-WRITE | `codex_projection_plan_writes_nothing`, `codex_openai_yaml_only_explicit_fields` |
| REQ-02C-01 | `codex_projection_requires_host_neutral_frontmatter`(감사 비차단 지적으로 추가) |
| REQ-02C-04 | `codex_duplicate_moai_skill_is_warned` |
| 오류 kind | `codex_messages_have_a_deliberate_kind` |
| I4 AC-1 설치→검증→재설치 no-op(홈·프로젝트 두 fixture) | `codex_projection_apply_installs_verifies_and_reinstalls_as_noop` |
| I4 REQ-I4-01 manifest 불일치 시 쓰지 않음(수정·미등록·누락·manifest 없음) | `codex_projection_apply_refuses_a_manifest_mismatch_and_writes_nothing` |
| I4 덮어쓰지 않음·소유하지 않은 파일 보존 | `codex_projection_apply_never_overwrites_and_keeps_unowned_files` |
| I4 게이트(permission·stale·yaml 변경·파일/symlink 대상) | `codex_projection_apply_gates` |
| I4 오류 kind | `codex_projection_apply_messages_have_a_deliberate_kind` |
| HA-01-CODEX, HA-10-SPIKE | not_run(레인). 운영자 실측 여부는 사용자 결정 대기 |

## Adapter gap

분류는 02-D와 같다: `구현 필요`, `대안 있음`, `미검증`.

| # | 책임 | 분류 | 근거 | 후속 제안 |
| --- | --- | --- | --- | --- |
| C1 | trust 판정과 `blocked_by_trust` 보고 | 구현 필요 | config reference `projects.<path>.trust_level`(2026-11-09 재확인: 위치는 user config만 명시). 파서는 #2240에서 들어옴(아래 "config.toml 생성") | doctor가 user config의 trust를 **읽는** 후속 Issue. 쓰기는 하지 않음: #2240은 trust를 선언(`trust.declaration`)만 하고 홈 쓰기는 blocked |
| C2 | `project_doc_fallback_filenames`·`project_doc_max_bytes`를 config에서 읽기 | 대안 있음 | 쓰기는 #2240(managed 키). doctor는 아직 `--fallback`·`--max-bytes` 인자 | doctor가 project `.codex/config.toml`을 읽는 후속 Issue(C1과 함께) |
| C3 | `tui.status_line` 항목 ID | 닫힘(#2928 S-5) | config reference에는 ID 목록이 없고, openai/codex `rust-v0.169.0`(`a41fe5b443652cbdbee57717fbb9fd78d8d8b32a`) `codex-rs/tui/src/bottom_pane/status_line_setup.rs:57-169`의 `StatusLineItem`에서 41개(별칭 포함)를 확인(2026-11-14) | `codex_config::CODEX_STATUS_LINE_ITEMS`로 검증. 목록 밖 항목은 그대로 쓰되 항목마다 `unknown_item`으로 보고(아래 "상태줄 (#2928 S-5)"). Codex 버전이 바뀌면 같은 파일을 다시 읽어 목록을 갱신 |
| C4 | hook 등록(`hooks.json` 또는 inline, 한 곳만) | 대안 있음 | RC1 결정 0개 → rc.2 I1 #2238에서 `.codex/hooks.json` 하나로 생성 | `hooks-apply --host codex`가 `.codex/hooks.json`만 쓴다(inline `[hooks]`는 쓰지 않음). 실행 전 hash trust 검토(`/hooks`) 안내는 apply 결과 `trust_required`에 있다 |
| C5 | projection 설치(사용자 홈/project) | 대안 있음 | I4 #2241: `codex-projection-apply`가 호출자가 정한 대상에 설치·검증한다. 기본 위치는 없다 | 위치 결정(P04) 뒤 기본값과 UPDATE 경로(`update.rs`의 `~/.agents/skills/moai`)와의 관계를 정하는 후속 Issue |
| C6 | Desktop·CLI·IDE 명시 호출 실측 receipt(version·표면·읽힌 경로) | 미검증 | REQ-02C-03 | 운영자 실측(HA-01, HA-10). CLI는 외부 전송을 승인한 뒤 |
| C7 | hooks 페이지의 handler 종류·exit 의미 원문 확인 | 미검증 | 이번 조회에서 요약만 얻음 | 지원표 다음 갱신 때 원문을 다시 조회 |

## 결정이 필요한 선택지

- **projection 위치**: HOLD. A `~/.agents/skills/moai`(D24 기본 후보) / B `<repo>/.agents/skills/moai`(동명 skill 중복 노출 위험).
- **운영자 E2E**: HOLD. Desktop GUI 실측과 CLI `$moai` 실제 호출(외부 전송·비용)을 할지, 누가 할지.
- **TOML 파서**: #2240에서 `toml_edit =0.25.3`(MIT OR Apache-2.0, 서식 보존)으로 들였다. doctor 읽기(C1·C2)는 후속 Issue다.

## config.toml 생성 (Issue #2240, rc.2 I3)

`moai_cli::hosts::codex_config::{config_plan, config_apply}`가 프로젝트 `.codex/config.toml`만 다룬다. #2237의 `codex-config-plan PROJECT` / `codex-config-apply PROJECT --plan-digest D --permission-reference R`가 이 라이브러리를 연결한다. stdin은 아래 managed dotted key의 JSON object이며 두 명령 모두 64 KiB 상한이다. plan은 무쓰기이고 apply는 같은 입력·digest·permission reference를 다시 검증한다. HOME(Windows에서는 USERPROFILE도 지원)과 같은 project, 또는 project의 `.codex`가 CODEX_HOME인 대상은 거부한다. CODEX_HOME이 설정됐지만 경로를 확인할 수 없는 경우도 거부한다. HOME·USERPROFILE이 모두 없거나 그 경로를 확인할 수 없으면 home 비교를 할 수 없으므로 plan·apply 모두 쓰기 전에 거부한다. 설정 키·기본값·상한은 기존 writer와 같고 trust·host load는 판정하지 않는다.

Claude `settings.json`과 같은 baseline/current/desired 3-way(REQ-I3-01): 소유 receipt는 `.moai-cli/manifests/codex-config.json`, 사용자가 MoAI가 둔 값을 바꾼 키는 `conflict`로 아무것도 쓰지 않고, 사용자가 이미 같은 값을 갖고 있던 키는 소유하지 않는다. managed 키 밖의 키·표·주석·빈 줄·순서는 그대로다(`toml_edit`). 바꾼 키도 앞뒤 주석을 유지한다. `tests/codex_config_cli.rs`는 합성 HOME/project에서 CLI 왕복·거부 경로를 검사하며 실제 host 설정이나 실세션은 사용하지 않는다. P04/C5와 MCP 결정은 이 CLI 연결로 해제되지 않는다.

| managed 키 | 받는 값 | 근거 |
| --- | --- | --- |
| `tui.status_line` | 문자열 1–32개(각 1–128자, 제어 문자 없음), 또는 preset 이름 `"moai"`(#2928 S-5, 다른 문자열은 거부). `null`은 받지 않음(TOML에 null 없음) | Codex 0.169.0 소스의 ID 목록(C3)과 대조. 목록 밖 항목만 `status_line_unverified`·`unknown_item` |
| `skills.bundled.enabled` | bool | 기본 `true`, trusted project에서 `false` 가능 |
| `project_doc_fallback_filenames` | 파일 이름 1–32개(`/`·`\`·`.`·`..` 거부) | 기본값 미기재 |
| `project_doc_max_bytes` | 정수 1–1,048,576 | 기본 32 KiB(AGENTS.md 페이지). 상한은 문서에 없어 MoAI가 1 MiB로 정함 |

거부: `projects.*`(trust는 user config, 홈 쓰기 blocked), 프로젝트 파일에서 Codex가 무시하는 키(`openai_base_url`, `chatgpt_base_url`, `apps_mcp_product_sku`, `model_provider`, `model_providers`, `notify`, `profile`, `profiles`, `experimental_realtime_ws_base_url`, `otel`), 그 밖의 키, 홈 디렉터리와 같은 project, symlink `.codex/`·`.moai-cli/`, 256 KiB를 넘는 파일·receipt·결과, 파싱되지 않거나 부모가 표가 아닌 파일(`requires preservation review`). 오류 문구의 kind는 `src/hosts/codex_config.rs` 자체 시험(`every_message_has_a_deliberate_kind`, `contract::classify` 단언)이 확인한다(D8 #2254 규칙: 공용 배열에 행을 더하지 않음).

| 공식 문서 | 확인한 내용 | 확인일 |
| --- | --- | --- |
| https://learn.chatgpt.com/docs/config-file/config-basic | user `~/.codex/config.toml`, project `.codex/config.toml` override, "loads project `.codex/` layers only when you trust the project", project root→cwd 순서(가까운 쪽 우선) | 2026-11-09 |
| https://learn.chatgpt.com/docs/config-file/config-reference | `tui.status_line` `array<string> \| null`(ID 미열거), `skills.bundled.enabled` `boolean` 기본 `true`, `project_doc_fallback_filenames` `array<string>`, `project_doc_max_bytes` `number`, `projects.<path>.trust_level` `"trusted"\|"untrusted"`(위치는 user config만 명시), 프로젝트에서 무시되는 키 10개 | 2026-11-09 |
| https://learn.chatgpt.com/docs/agent-configuration/agents-md | `project_doc_max_bytes` 기본 32 KiB, fallback 이름은 디렉터리마다 `AGENTS.override.md`·`AGENTS.md` 다음 | 2026-11-09 |

## 상태줄 (#2928 S-5)

사용자 지시(Issue #2928 comment 6877102675): Claude Code(CC) 상태줄 요소 중 Codex가 보여 줄 수 있는 것을 CC와 같은 순서로 놓는다. 새 명령은 만들지 않고(결정 C1 (가)) 위 `codex-config-plan`/`codex-config-apply`로 준다. Codex가 못 보여 주는 요소는 아래 표에 '불가 + 근거'로만 적고 우회(외부 스크립트 등)는 하지 않는다(결정 C2 (가)).

Codex CLI 0.169.0의 바닥줄은 내장 항목 ID를 고르고 순서만 정하는 방식이다(`/statusline` 또는 `tui.status_line`). CC의 `statusLine.command`처럼 외부 명령을 실행하는 방법은 문서에 없다. 그래서 한 줄 목록이고(세 줄 배치 불가), 아래 '불가' 요소는 채울 수 없다.

**preset.** 요청에서 `{"tui.status_line": "moai"}`는 아래 10개 목록(`codex_config::MOAI_STATUS_LINE_PRESET`)으로 바뀐 뒤 평소처럼 3-way 판정을 받는다. 명시 목록도 그대로 받는다.

```toml
[tui]
status_line = ["model", "reasoning", "codex-version", "project-name", "pull-request-number", "git-branch", "branch-changes", "thread-name", "context-remaining", "weekly-limit"]
```

세션 한 번만 쓰려면 Codex 안에서 `/statusline model reasoning codex-version project-name pull-request-number git-branch branch-changes thread-name context-remaining weekly-limit`.

**plan 결과.** `status_line`은 `before`(지금 파일의 목록, 없으면 `null`), `after`(적용 뒤 목록 — conflict면 쓰지 않으므로 `before`와 같음), `preset`(`"moai"` 또는 `null`), `items`(요청한 항목마다 `{"id", "status": "known" | "unknown_item"}`)를 담는다. `status_line_unverified`는 목록 밖 항목만 `items`에 담고, 그런 항목이 있을 때만 `reason`이 문자열이다.

**목록 밖 항목.** 그대로 쓰고(지금까지와 같음), 항목마다 `unknown_item`으로 보고하며 빼지 않는다. Codex 0.169.0은 해석하지 못한 항목을 경고 "Ignored invalid status line item"과 함께 건너뛰고(`codex-rs/tui/src/chatwidget/status_surfaces.rs:121-138`, `:467-468`), 더 새 Codex가 아는 ID일 수 있어서 거부하지 않는다.

**사용자 값과 되돌리기.** 프로젝트 `.codex/config.toml`에 사용자가 이미 `tui.status_line`을 둔 경우는 덮어쓰지 않는다: plan은 `conflict`(`before`=`after`=사용자 목록), apply는 아무것도 쓰지 않는다. 사용자 config(`~/.codex/config.toml`, CODEX_HOME)의 목록은 이 명령이 쓰지 않으므로 그대로 남는다. MoAI가 넣은 목록을 되돌리려면 `tui.status_line`을 뺀 요청으로 plan → apply한다: receipt가 MoAI 소유로 기록한 키만 지워지고 파일은 원래 바이트로 돌아온다. 별도 `*-revert` 명령은 두지 않았다(codex config는 사용자 값을 바꾸지 않으므로 보관할 이전 값이 없다).

**이름표(leader/agent).** `thread-name`은 Codex 대화 이름을 보여 주고, 이름이 없으면 칸을 생략한다(`status_line_setup.rs` `ThreadName`: "Current thread name, omitted when unnamed"). 이름은 Codex 안에서 `/rename leader`(leader 세션) 또는 `/rename agent`(그 밖의 세션)로 정한다(S-2 규칙). 실행 플래그나 설정 키로는 정할 수 없다. 지금 화면의 'Set MoAI session name' 자리는 이 이름이 정해지면 채워진다.

| CC 줄 | CC 요소 | Codex 0.169.0 | 항목 ID / 근거 |
| --- | --- | --- | --- |
| 1 | 모델 | 가능 | `model` |
| 1 | effort | 가능 | `reasoning`(또는 `model-with-reasoning`) |
| 1 | 캐시/시간 | 불가 | 해당 항목 없음(토큰 수 항목 `used-tokens` 등만 있음) — S1 |
| 1 | 호스트 버전 | 가능 | `codex-version` |
| 1 | moai 버전 | 불가 | 외부 명령 실행 없음, ID 목록만 받음 — D2, S1 |
| 2 | 저장소 | 일부 | `project-name`(또는 `current-dir`) |
| 2 | ahead/behind | 불가 | 해당 항목 없음 — `branch-changes`는 기본 가지 대비 커밋 차이 |
| 2 | PR | 가능 | `pull-request-number`(없으면 생략) |
| 2 | 가지 | 가능 | `git-branch` |
| 2 | 변경 | 뜻이 다름 | `branch-changes` = 기본 가지 대비 커밋된 차이, 작업 트리의 미커밋 변경이 아님 |
| 2 | worktree | 불가 | 해당 항목 없음(`current-dir` 경로로 짐작만) |
| 3 | 🏷️ 이름표 | 가능 | `thread-name`(또는 `thread-title`) — 값은 `/rename leader`·`/rename agent`, 플래그·설정 키로는 못 정함 |
| 3 | CTX % | 가능 | `context-remaining`(또는 `context-used`) |
| 3 | Week % | 가능 | `weekly-limit` |

| 출처 | 확인한 내용 | 확인일 |
| --- | --- | --- |
| D1 https://learn.chatgpt.com/docs/developer-commands?surface=cli | "/statusline · … Pick and reorder footer items (model/context/limits/git/tokens/session) and persist in config.toml." · "To set the footer for this session only, list the items in display order." | 2026-11-14 |
| D2 https://learn.chatgpt.com/docs/config-file/config-reference | "`tui.status_line` · array<string> \| null · Ordered list of TUI footer status-line item identifiers. `null` disables the status line." | 2026-11-14 |
| S1 openai/codex `rust-v0.169.0` (`a41fe5b443652cbdbee57717fbb9fd78d8d8b32a`) `codex-rs/tui/src/bottom_pane/status_line_setup.rs:57-169` | `StatusLineItem` 변형과 별칭 = ID 41개(`CODEX_STATUS_LINE_ITEMS`). `status-line-use-theme-colors`는 picker의 색 토글이지 항목이 아님(같은 파일 :43, :309-316) | 2026-11-14 |
| S2 같은 태그 `codex-rs/tui/src/chatwidget/status_surfaces.rs:121-138`, `:467-468` | 해석하지 못한 항목은 "Ignored invalid status line item" 경고와 함께 건너뜀 | 2026-11-14 |
| S4 같은 태그 `codex-rs/features/src/lib.rs` | `tui_extensions_statusline`("Enable experimental MCP status line providers.", UnderDevelopment, 기본 꺼짐) — 지금은 쓸 수 없음 | 2026-11-14 |

조사 원문: Issue #2928 comment 6877416640. 시험: `tests/codex_config.rs`(preset 순서 골든, plan 무쓰기 트리 비교, 이 키만 바뀌는 바이트 비교, 빼면 원래 바이트, 알려진/모르는 ID), `tests/codex_config_cli.rs`(임시 CODEX_HOME에서 CLI preset 왕복, 사용자 config 바이트 유지). 실제 Codex 화면 캡처와 `codex -c 'tui.status_line=[...]'`·첫 프롬프트 `/rename` 동작은 not_run.

미확인: 세 페이지 모두 "last updated" 없음, `project_doc_fallback_filenames` 기본값, `project_doc_max_bytes` 상한, project 파일 안의 `trust_level` 취급. host 실측(`host_loaded`, `execution_verified`)은 not_run.

## 미실행·위험

- not_run:
  - Codex Desktop·CLI·IDE 실측: 발견, `$moai`·`/skills` 호출, AGENTS 포함, trust 화면
  - Codex cloud 환경 hook
  - implicit invocation 선택률
  - Windows·Linux
- untrusted project에서는 `.codex/` 전체가 무시된다. 그래서 파일을 만든 것만으로 활성화됐다고 볼 수 없다. doctor가 trust를 판정하지 못하는 동안 이 위험은 남는다.
- 같은 이름의 `moai` skill이 홈과 project에 함께 있으면 둘 다 selector에 나온다. doctor는 경고만 하고 둘 중 하나를 고르지 않는다.
- Codex 문서 페이지에는 version과 날짜가 없다. 그래서 stale 판정은 조회일로만 한다.
