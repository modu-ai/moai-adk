# Codex adapter 계약 (02-C #1993)

공통 skill package(`skills/moai/`, 02-A #1991)를 OpenAI Codex(Desktop 앱, CLI, IDE 확장)에 연결하는 adapter의 계약이다. RC1에서는 projection을 계획하고 설치 상태를 진단하는 데까지만 한다. 사용자 홈에 설치하거나 UPDATE하는 일은 EP-05에서 한다. 이 Issue의 코드는 아무 파일도 쓰지 않는다. 확인일은 모두 2026-11-08이다. 지원표 행은 `docs/support-matrix.md`에 있다(Codex 조합은 운영자 실측 전까지 not_run으로 표기된다).

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

원문을 확인하지 못한 항목도 있다. hooks 페이지의 handler 종류(`command`·`mcp_tool` 지원, `prompt`·`agent` 무시), "exit 0과 빈 출력은 계속", PreToolUse 오류·timeout 때의 동작, `features.hooks` 기본값이다. 긴 페이지라 이 부분은 요약만 얻었다. 그래서 표와 코드에 쓰지 않았고, Issue 본문의 해당 근거는 이번 조회로 다시 확인되지 않은 상태다.

로컬 관측은 `codex --version` → `codex-cli 0.163.0`과 `codex --help`의 subcommand 목록(`exec`, `doctor`, `features`, `plugin` 등)뿐이다. 둘 다 모델을 호출하지 않는다.

## Operation

두 operation 모두 01-C envelope(`{protocol_version, operation, ok, result|error}`)을 쓰고 쓰기가 없다. `result` 키는 `docs/rust-product-path.md`의 operation 표에 있다.

### `codex.projection_plan` (`codex-projection-plan PACKAGE --target DIR`)

- `--target`: 절대 경로여야 하고, 마지막 이름이 `moai`여야 하고, package 바깥이어야 한다. 없으면 `usage`, 조건에 어긋나면 `invalid_input`이다.
- 파일 계획: package의 모든 파일을 sha256이 같은 복사(`action: copy`)로 계획한다. 그래서 `SKILL.md`는 02-A canonical body와 바이트가 같다(REQ-02C-01). `skill_md_identical`은 고정값이 아니라, 계획한 `SKILL.md`의 sha256을 package에서 다시 읽은 바이트와 비교한 결과다. 계획 전에 `SKILL.md` frontmatter가 `name: moai`와 `description` 두 키만 갖는지 확인하고, 다른 키가 있거나 둘 중 하나가 없으면 `invalid_input`으로 거부한다. package 안에 `agents/`가 있으면 거부한다. `agents/openai.yaml`은 adapter metadata이지 package 파일이 아니기 때문이다.
- `agents/openai.yaml`: `interface.display_name`·`short_description`·`default_prompt`만 쓴다. `policy.allow_implicit_invocation`은 `--allow-implicit-invocation`을 줬을 때만 쓴다. 주지 않으면 키가 없어 Codex 기본값(`true`)이 그대로 유지된다(REQ-02C-02). 값은 큰따옴표 YAML 문자열로 escape한다(PyYAML `safe_load`로 다시 읽어 확인). 제어 문자가 있거나 256자를 넘으면 거부한다.
- 설정 요청: `--status-line ID`(반복 가능)와 `--disable-bundled-skills`는 `config_changes: {status: "hold", requested, unverified_items, writes: []}`로 돌려준다. 모든 status-line ID는 `unverified_items`에 들어간다(REQ-02C-07). 요청이 없으면 `status: none_requested`다.
- 결과: `hooks.generated: []`, `files_created: []`, `host_loaded: false`, `execution_verified: false`.

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
| HA-01-CODEX, HA-10-SPIKE | not_run(레인). 운영자 실측 여부는 사용자 결정 대기 |

## Adapter gap

분류는 02-D와 같다: `구현 필요`, `대안 있음`, `미검증`.

| # | 책임 | 분류 | 근거 | 후속 제안 |
| --- | --- | --- | --- | --- |
| C1 | trust 판정과 `blocked_by_trust` 보고 | 구현 필요 | config reference `projects.<path>.trust_level`. TOML 파서 HOLD | 02-B가 TOML 표면을 정한 뒤 doctor가 trust를 읽는 후속 Issue. 그때도 trust를 바꾸지는 않음 |
| C2 | `project_doc_fallback_filenames`·`project_doc_max_bytes`를 config에서 읽기 | 대안 있음 | 지금은 `--fallback`·`--max-bytes` 인자 | C1과 같은 후속 Issue |
| C3 | `tui.status_line` 항목 ID | 미검증 | config reference에 ID 목록 없음 | 운영자 실측(설치본의 `/statusline` 또는 schema)으로 ID를 확인한 뒤 의미 단위 병합(02-B 규칙) |
| C4 | hook 등록(`hooks.json` 또는 inline, 한 곳만) | 대안 있음 | RC1 결정 0개 | opt-in hook이 필요해지면 별도 Issue에서 위치를 정하고 hash trust 안내 |
| C5 | projection 설치(사용자 홈/project) | 구현 필요 | plan만 있음 | EP-05 설치 Issue. 위치 HOLD 결정 뒤 |
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
