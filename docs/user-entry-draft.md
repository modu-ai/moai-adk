# 사용자 첫 진입 개편안 (08-D #2015 초안 — 게시 전)

목표: "한 번 추가하고, 기존 앱에서 쓴다"(설계 00-page §22). 자연어 진입이 기본이고 수동 명령은 명시적 fallback이다. 3.x Go 세대 명령을 차세대 필수 단계로 소개하지 않는다(REQ-08D-06). docs-site 배포는 범위 밖 — 이 파일은 README·문서 사이트가 채택할 초안이다.

## 초안 본문

### MoAI를 프로젝트에 한 번 추가하기

설치 후 프로젝트 루트에서 호스트 세션을 열고 이렇게 말합니다:

- Claude Code: `/moai 이 프로젝트 현재 상태를 읽고 설정해 줘`
- Codex: `$moai 이 프로젝트 현재 상태를 읽고 설정해 줘`

SETUP이 끝나면 같은 진입으로 READ → PLAN → RUN → VERIFY → AUDIT → SYNC가 이어집니다. 매번 절차를 외울 필요가 없습니다 — 요청하면 워크플로가 필요한 만큼만 로딩됩니다.

### 설치

```sh
# RC1 게시 후 — GitHub Releases의 `moai-cli-v<version>` prerelease 자산
# (installer가 ~/.moai-cli/bin/moai-cli 설치와 PATH 안내를 담당합니다)
```

- 기존 Go 세대(`moai` 3.x) 사용자: `~/.moai/` 작업 데이터는 그대로 보존되며 두 세대는 서로의 업데이트 대상이 아닙니다. 전환 절차는 별도 안내(04-D)가 제공될 때까지 기존 대로 `moai`를 쓰면 됩니다.
- 서명 검증: release artifact는 GitHub artifact attestation으로 검증합니다(`gh attestation verify <asset> -R modu-ai/moai-adk`). checksum은 무결성 재확인용이며 서명을 대체하지 않습니다.

### 요청 예시

| 하고 싶은 것 | 말하면 |
| --- | --- |
| 프로젝트 등록·설정 | `/moai 이 프로젝트 설정해 줘` |
| 현황 읽기 | `/moai 지금 저장소 상태 요약해 줘` |
| 요구 정리 | `/moai 이 기능 요구사항 정리해 줘` |
| 구현·검증 | `/moai 이 SPEC 구현하고 검증까지 해 줘` |

### 호스트 세션과 역할

호스트 세션은 역할 없이 엽니다 — `moai-cli cc|glm|codex`는 평범한 host 세션이고, 세션 안에서 `/moai --mode-leader`·`/moai --mode-agent`로 MoAI 역할을 정합니다(모드 표기는 이 두 개뿐, 사용자 결정 2026-11-10). 리더·에이전트 론처 플래그 `-l|-a`(`moai-cli -l|-a`, `moai-cli cc|glm|codex -l|-a`)는 세션 이름(`leader`, `agent-N`)을 붙이고 전체 권한(기본값)으로 여는 형태로 유지합니다(사용자 결정 2026-11-10, Issue #2359 댓글 6760976883; 기본값은 사용자 설정 `launcher.role_access`로 바꿀 수 있습니다 — 아래 세션 론처 문단, Issue #2413). #2164의 deprecated 경고는 없앴습니다.

### 수동 명령 (명시적 fallback)

자연어 진입이 불가능한 환경에서만 내부 실행 도구를 직접 씁니다: `moai-cli generation-plan`, `moai-cli update-plan` 등(각 명령은 plan/apply 2단계이며 plan digest 없이는 아무것도 쓰지 않습니다). 일상 사용 경로가 아닙니다.

### 3.x 명령에 대해 (fallback)

기존 `moai init`·`moai update`(Go 3.x)는 구 세대의 명령입니다. RC1을 이미 설치했다면 3.x 명령을 새 절차로 쓰지 마세요 — 두 세대의 업데이트 통로는 분리되어 있습니다(지원표와 릴리스 체크리스트 참조). 마이그레이션 도구는 별도 제공 예정입니다.

## 세션 론처 (저장소 개발용, 현행 — DR-08 #2145)

`-l|-a`는 세션 이름과 전체 권한(기본값)을 붙이는 형태로 유지합니다(사용자 결정 2026-11-10, Issue #2359 댓글 6760976883). 기본값은 사용자 전용 파일 `~/.moai-cli/config/launcher.yaml`의 `launcher.role_access` 키로 바꿀 수 있다(Issue #2413, 아래). 아래 표는 오늘 기준 동작을 기록합니다.

`moai-cli`는 저장소 개발 세션을 띄우는 론처를 갖고 있다. 사용자 진입(`/moai`·`$moai`)과는 별개인 개발 도구다. 아래 표는 `src/main.rs`의 `Cli`(45–56행), `LaunchOpts`(58–79행), `HostLaunch`(81–96행), `Command::Cc|Glm|Codex`(519–524행)와 `src/session.rs` 머리말을 기준으로 한다. 2026-11-09에 base `604792666` 빌드로 처음 확인했고, 권한·이름 규칙은 2026-11-10 #2359 빌드의 `--dry-run`과 시험으로 다시 확인했다.

| 형태 | 동작 |
| --- | --- |
| `moai-cli cc\|glm\|codex` (역할 없음) | 평범한 host 세션. 역할·세션 이름을 붙이지 않는다(#2053). 권한 관련 플래그도 붙이지 않아, Claude Code는 사용자·프로젝트 설정의 모드로, codex는 `~/.codex/config.toml`의 샌드박스·승인 설정으로 시작한다(#2359). Claude 계열은 `--model opus --effort medium` 기본값을 쓰고, `--advisor`는 붙이지 않는다(아래 `launcher.advisor`). |
| `moai-cli cc\|glm\|codex -l` (`--leader`) | 리더 세션. 이름 `leader`, 전체 권한(기본값; 아래 `launcher.role_access`). |
| `moai-cli cc\|glm\|codex -a` (`--agent`) | 에이전트 세션, 전체 권한(기본값; 아래 `launcher.role_access`). 이름은 `--name`, 없으면 프로젝트 번호에서 `agent-N`(아래). |
| `moai-cli -l` / `moai-cli -a [cc\|glm\|codex]` | 위와 같은 요청의 짧은 형태. 기본 host는 `cc`. |
| `--name N` | 세션 이름. |
| `--model M` | 주 모델. Claude 계열 기본 `opus`, codex는 지정했을 때만 `-m M`. |
| `--effort E` | Claude effort(`low\|medium\|high\|xhigh\|max\|ultracode`), 기본 `medium`. |
| `--advisor A` | Claude 계열에 `--advisor A`를 붙인다(`fable\|opus\|sonnet` 또는 `claude-*` 전체 id). 기본은 붙이지 않으며, 사용자 설정 `launcher.advisor`(아래)보다 우선한다. 사용 크레딧으로 과금될 수 있다. |
| `--no-advisor` | 사용자 설정 `launcher.advisor`가 켠 advisor를 이번 실행에서 끈다. `--advisor`와 함께 주면 거부한다. |
| `--permissions` | 역할 세션을 확인 모드로 연다. Claude 계열은 `--dangerously-skip-permissions`를 붙이지 않고, codex는 `-s workspace-write -a on-request`로 띄운다. 사용자 설정 `launcher.role_access`(아래)보다 우선한다. 역할 없는 세션에서는 바뀌는 것이 없다. |
| `--dry-run` | host를 띄우지 않고 조립한 명령을 JSON으로 출력한다. |
| `[PROMPT]` | 첫 프롬프트 텍스트(codex만). |

host별 argv:

- `cc`: `claude --name <name> --model opus --effort medium --dangerously-skip-permissions` (`-l|-a` 기본값; `--permissions` 또는 `launcher.role_access: prompt`(아래)면 이 플래그가 빠진다). 역할 없는 형태는 `claude --model opus --effort medium`로, 이름과 권한 플래그가 없다. `--advisor A` 또는 `launcher.advisor`(아래)가 있으면 `--effort` 뒤에 `--advisor A`가 붙는다.
- `glm`: `cc`와 같은 Claude Code argv에, Go 도구와 같은 환경 변수 묶음(`ANTHROPIC_BASE_URL`·`API_TIMEOUT_MS`, 모델 별칭 `ANTHROPIC_DEFAULT_*_MODEL`(opus·sonnet·fable은 `glm-5.3`, haiku는 `glm-5.3-flash`), 1M 컨텍스트 선언 `CLAUDE_CODE_AUTO_COMPACT_WINDOW`·`CLAUDE_CODE_MAX_CONTEXT_TOKENS`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, z.ai MCP용 `Z_AI_API_KEY`)을 더해 z.ai로 보낸다. 토큰은 셸의 `ANTHROPIC_AUTH_TOKEN`을 쓰고, 없으면 키 파일 `~/.moai/.env.glm`(Go 도구와 함께 쓰는 전역 저장소), 이어서 `~/.moai-cli/.env.glm`에서(둘 다 읽기만 하며, 65536바이트 이하의 일반 파일일 때만 연다) `ANTHROPIC_AUTH_TOKEN=` 또는 `GLM_API_KEY=` 줄을 찾아 host 프로세스에만 넘긴다(#2359). 론처는 값을 출력·저장하지 않고 키 파일을 만들거나 고치지 않는다. 어디에도 없으면 `blocked:` 메시지와 exit 1로 끝난다.
- `codex`: 역할 세션은 `codex -s danger-full-access -a never [-m <model>] [PROMPT]`(`--permissions` 또는 `launcher.role_access: prompt`(아래)면 `-s workspace-write -a on-request`), 역할 없는 세션은 `codex [-m <model>] [PROMPT]`로 `-s`·`-a`를 붙이지 않는다. Codex에는 세션 이름 플래그가 없어 이름은 첫 프롬프트(`MoAI session name: <name>`)로 전한다. Codex 세션에는 RC1의 세션 간 메시지 채널(ListAgents·SendMessage)이 없다. advisor는 Claude 계열에만 붙는다.

권한과 이름 규칙은 사용자 결정 2026-11-10, Issue #2359 댓글 6760976883을 따른다. 같은 날의 앞선 두 결정(역할 없는 세션도 전체 권한, `-l|-a` 제거)은 이 결정으로 대체되었다. `-l|-a` 세션은 host와 프로젝트 `.claude/settings.json`의 `permissions` 키에 관계없이 전체 권한(기본값)으로 시작하고(#2244 규칙 대체), `--permissions`를 주면 확인 모드로 뜬다. 역할 없는 세션에는 권한 플래그를 붙이지 않는다. codex에서는 이 차이가 중요하다. Codex는 CLI 플래그를 `~/.codex/config.toml`보다 우선하므로, 역할 세션은 사용자 설정과 관계없이 묻지 않도록 플래그를 명시하고, 역할 없는 세션은 플래그를 빼서 사용자 설정을 따르게 한다(전에는 `-s workspace-write -a on-request`를 강제로 붙여 사용자 설정을 덮어썼다). 역할 세션의 기본값은 `session::FULL_ACCESS_DEFAULT` 한 곳에 있고, 사용자 결정 Issue #2413 옵션 A(댓글 6778672356, 설계 6781812681)로 사용자 전용 파일 `~/.moai-cli/config/launcher.yaml`의 `launcher.role_access` 키(`full|prompt`)가 이 기본값을 고른다. 이 파일은 프로젝트 profile `.yaml`과 같은 명시적 JSON 부분집합 규약(`serde_json`)으로만 읽으며 일반 YAML 파서는 쓰지 않는다. 우선순위는 CLI `--permissions` > 사용자 키 > 기본값 `full`이고 프로젝트 설정은 이 선택에 관여하지 않으며, 역할 없는 세션은 두 값 모두에서 권한 플래그를 붙이지 않는다. 파일이나 키가 없으면 기존 출력·argv·notes가 바이트 그대로이고, 홈 디렉터리는 `HOME`, Windows에서 `HOME`이 없거나 비었으면 `USERPROFILE`로 정하며(`key_vault::home` 규약) 둘 다 없을 때만 사용자 파일이 없는 것으로 본다. 값을 알 수 없으면 — 모르는 값, 문자열이 아닌·`null`인 `role_access`, `null`이거나 객체가 아닌 `launcher`·루트, 잘못된 JSON 문법, 읽을 수 없거나 일반 파일이 아닌 기존 파일(디렉터리·심볼릭 링크·FIFO) — host 시작과 agent 번호 할당 앞에서 실행을 거부하며, 설정 내용이나 원본 오류 값을 그대로 출력하지 않고 `--permissions`를 줬더라도 `full`로 조용히 빠지는 fallback은 없다. 상세 계약과 예시는 docs/rust-product-path.md의 #2413 문단이 정한다. 옛 `session` 하위 명령(Claude 전용)은 바뀌지 않았다.

advisor는 기본으로 붙이지 않는다(사용자 결정 2026-11-12, Issue #2476). 2026-11-08 결정(#2053 HOLD-L1 = A, #2062) 가운데 `--advisor fable` 기본값은 이 결정으로 대체되었고, `--model opus --effort medium` 기본값은 그대로다. Fable advisor가 사용 크레딧으로 과금되는 계정에서 크레딧을 설정하지 않았으면 host가 시작하자마자 exit 1로 끝났기 때문이다. advisor를 켜는 방법은 두 가지이고 우선순위는 CLI `--advisor A` > 같은 사용자 전용 파일의 `launcher.advisor` 키 > 없음이다. 키 값은 `none`(없는 것과 같다)·`fable`·`opus`·`sonnet`·`claude-*` 전체 id 중 하나다(예: `{"launcher":{"advisor":"fable"}}`). `--no-advisor`는 키가 켠 advisor를 이번 실행에서 끄고, `--advisor`와 함께 주면 거부한다. 키는 Claude 계열에만 적용된다. `glm`은 advisor를 빼고 그렇다는 note를 남기며(advisor는 Anthropic API 전용), codex에는 붙지 않는다. advisor가 주 모델 이상이어야 한다는 기존 검사도 그대로라, 기본 `opus`에서 `launcher.advisor: sonnet`은 거부된다. 잘못된 값(모르는 문자열·대소문자가 다른 값·빈 문자열·`null`·숫자·배열·객체, 깨진 JSON 문법)은 `role_access`와 같은 방식으로 host 시작과 agent 번호 할당 앞에서 거부하고, `--advisor`·`--no-advisor`를 줬더라도 검사하며, 값은 출력하지 않는다. 프로젝트 설정으로는 켤 수 없다. advisor를 붙인 실행의 notes에는 사용 크레딧으로 과금될 수 있다는 줄이, 키에서 온 경우에는 그 출처가 함께 나온다. 상세 계약은 docs/rust-product-path.md의 #2476 문단이 정한다.

`agent-N` 번호(#2361 해결): `-a`에 `--name`이 없으면 프로젝트(한 clone과 그 worktree 전체)마다 하나인 번호에서 다음 값을 받는다. host가 달라도 번호를 함께 쓰고, 번호는 계속 늘어나기만 하며 끝난 세션의 번호도 다시 쓰지 않는다. 저장 위치는 `<clone의 공통 git 디렉터리>/.moai-cli/sessions/agent-seq`(마지막으로 준 번호 한 줄)로, claim journal 옆이라 모든 worktree가 같은 파일을 보고 git에는 나타나지 않는다. 옆의 `.lock`을 잡은 채 읽고 1을 더해 쓰므로, 동시에 여러 개를 띄워도 같은 번호가 두 번 나오지 않는다(나중 실행은 기다린다). `--dry-run`은 다음 번호를 보여 주기만 하고 쓰지 않으며, `--name`과 `-l`은 번호를 쓰지 않는다. 막힌 실행(토큰 없는 `glm`, 읽을 수 없는 `launcher.role_access` 등)도 번호를 쓰지 않는다. git 저장소 밖이거나 저장 파일이 숫자 한 줄이 아니거나 일반 파일이 아니면(FIFO는 열지 않는다) 오류로 끝나며 `agent-1`로 돌아가지 않는다. 이때는 `--name`으로 이름을 주면 된다.

## 채택 시 확인 항목

- [x] 3.x 명령을 필수 단계로 안내하는 문장 0건(`grep -nE 'moai (init|update)\b' docs/*.md README.md | grep -v 'fallback\|3\.x'` = 0)
- [x] `/moai`·`$moai` 예시가 진입의 기본
- [ ] README 본문 반영 — README는 기존 사용자 영역 유지(이 초안 채택 시 별도 편집)
