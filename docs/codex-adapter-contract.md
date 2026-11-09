# Codex adapter 계약 (02-C #1993)

공통 skill package(`skills/moai/`, 02-A #1991)를 OpenAI Codex(Desktop 앱, CLI, IDE 확장)에 연결하는 adapter의 계약이다. RC1에서는 projection을 계획하고 설치 상태를 진단하는 데까지만 했다. rc.2 I4 #2241이 계획한 projection을 호출자가 정한 대상에 설치하는 `codex.projection_apply`를 더했다(아래 Operation 절). 설치 위치는 여전히 정하지 않는다. 02-C 범위의 확인일은 2026-11-08이고, I4가 다시 확인한 항목은 2026-11-09로 따로 적는다. 지원표 행은 `docs/host-support-matrix.md`의 Codex 절에 있다.

## 결정 (리더 RC1 임시, 2026-11-08, Issue 본문 r30)

| 항목 | 결정 | 영향 |
| --- | --- | --- |
| TOML 읽기 | 파서를 쓰지 않는다. TOML 읽기·쓰기 표면과 의존성 결정은 02-B #1992 소유다 | `config.toml`을 읽지 않는다. trust, inline `[hooks]`, `tui.status_line`은 `not_evaluated`로 보고하고 `trusted`라고 쓰지 않는다. fallback 이름과 max-bytes는 인자로 받는다. 02-B가 파서를 들이면 승격은 후속 Issue에서 한다 |
| hook | 0개. 아무것도 생성하지 않는다 | `.codex/hooks.json`도, inline `[hooks]`도 만들지 않는다. hash trust 검토가 필요 없다 |
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
- `skills.moai`: cwd → root의 `.agents/skills/moai`, `$HOME/.agents/skills/moai`, `/etc/codex/skills/moai` 중 있는 것을 모두 보여 준다. 둘 이상이면 병합되지 않는다는 경고를 낸다(REQ-02C-04).
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
| C1 | trust 판정과 `blocked_by_trust` 보고 | 구현 필요 | config reference `projects.<path>.trust_level`. TOML 파서 HOLD | 02-B가 TOML 표면을 정한 뒤 doctor가 trust를 읽는 후속 Issue. 그때도 trust를 바꾸지는 않음 |
| C2 | `project_doc_fallback_filenames`·`project_doc_max_bytes`를 config에서 읽기 | 대안 있음 | 지금은 `--fallback`·`--max-bytes` 인자 | C1과 같은 후속 Issue |
| C3 | `tui.status_line` 항목 ID | 미검증 | config reference에 ID 목록 없음 | 운영자 실측(설치본의 `/statusline` 또는 schema)으로 ID를 확인한 뒤 의미 단위 병합(02-B 규칙) |
| C4 | hook 등록(`hooks.json` 또는 inline, 한 곳만) | 대안 있음 | RC1 결정 0개 | opt-in hook이 필요해지면 별도 Issue에서 위치를 정하고 hash trust 안내 |
| C5 | projection 설치(사용자 홈/project) | 대안 있음 | I4 #2241: `codex-projection-apply`가 호출자가 정한 대상에 설치·검증한다. 기본 위치는 없다 | 위치 결정(P04) 뒤 기본값과 UPDATE 경로(`update.rs`의 `~/.agents/skills/moai`)와의 관계를 정하는 후속 Issue |
| C6 | Desktop·CLI·IDE 명시 호출 실측 receipt(version·표면·읽힌 경로) | 미검증 | REQ-02C-03 | 운영자 실측(HA-01, HA-10). CLI는 외부 전송을 승인한 뒤 |
| C7 | hooks 페이지의 handler 종류·exit 의미 원문 확인 | 미검증 | 이번 조회에서 요약만 얻음 | 지원표 다음 갱신 때 원문을 다시 조회 |

## 결정이 필요한 선택지

- **projection 위치**: HOLD. A `~/.agents/skills/moai`(D24 기본 후보) / B `<repo>/.agents/skills/moai`(동명 skill 중복 노출 위험).
- **운영자 E2E**: HOLD. Desktop GUI 실측과 CLI `$moai` 실제 호출(외부 전송·비용)을 할지, 누가 할지.
- **TOML 파서**: 02-B #1992에서 정한다. 정해지면 C1·C2·C3을 승격하는 후속 Issue를 만든다.

## 미실행·위험

- not_run:
  - Codex Desktop·CLI·IDE 실측: 발견, `$moai`·`/skills` 호출, AGENTS 포함, trust 화면
  - Codex cloud 환경 hook
  - implicit invocation 선택률
  - Windows·Linux
- untrusted project에서는 `.codex/` 전체가 무시된다. 그래서 파일을 만든 것만으로 활성화됐다고 볼 수 없다. doctor가 trust를 판정하지 못하는 동안 이 위험은 남는다.
- 같은 이름의 `moai` skill이 홈과 project에 함께 있으면 둘 다 selector에 나온다. doctor는 경고만 하고 둘 중 하나를 고르지 않는다.
- Codex 문서 페이지에는 version과 날짜가 없다. 그래서 stale 판정은 조회일로만 한다.
