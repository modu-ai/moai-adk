# 사용자 첫 진입 개편안 (08-D #2015 초안 — 게시 전)

목표: "한 번 추가하고, 기존 앱에서 쓴다"(설계 00-page §22). 자연어 진입이 기본이고 수동 명령은 명시적 fallback이다. 3.x Go 세대 명령을 차세대 필수 단계로 소개하지 않는다(REQ-08D-06). docs-site 배포는 범위 밖 — 이 파일은 README·문서 사이트가 채택할 초안이다.

## 초안 본문

### MoAI를 프로젝트에 한 번 추가하기

설치 후 프로젝트 루트에서 호스트 세션을 열고 이렇게 말합니다:

- Claude Code: `/moai 이 프로젝트 현재 상태를 읽고 설정해 줘`

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

호스트 세션은 역할 없이 엽니다 — `moai cc|glm [--name <name>]`은 평범한 host 세션이고, 세션 안에서 `/moai agent`(레인)나 `/moai leader`(조정 창)를 입력해 MoAI 역할 모드로 들어갑니다(사용자 선택 2026-11-22, SPEC-20261124-moai-leader-route; 옛 `--mode-…` 표기와 `/moai work` 낱말은 쓰지 않습니다). 이 둘은 모드 전환이라 뒤에 지시 글을 받지 않고, 지시는 크로스 세션 메시지로 옵니다. 리더·에이전트 론처 플래그 `-l|--leader`·`-a|--agent`는 `moai cc|glm` 뒤에서 다시 쓰입니다(사용자 결정 2026-11-25, SPEC-20261125-agent-team-session-names; 2026-11-09의 D-1 결정을 뒤집습니다): `-l`은 세션 `<프로젝트 이름>-leader`, `-a`는 세션 `<프로젝트 이름>-agent-<영어 이름>`(SPEC-20261125-session-names-project-prefix)을 권한 우회로 열고 첫 입력으로 `/moai leader`·`/moai agent`를 넘기며, `--name`이 이름을 바꿉니다. 역할 플래그가 없는 실행의 권한 모드는 auto mode(`--permission-mode auto`)로 시작합니다(사용자 결정 2026-11-23, SPEC-20261123-launcher-auto-mode; 사용자 설정 `launcher.role_access`로 `full`(우회)이나 `prompt`로 바꿀 수 있습니다 — 아래 세션 론처 문단, Issue #2413). 역할 플래그 없이 연 세션은 `/moai agent`(레인)·`/moai leader`(조정 창)를 입력해 역할로 들어갑니다.

### 수동 명령 (명시적 fallback)

자연어 진입이 불가능한 환경에서만 내부 실행 도구를 직접 씁니다: `moai generation-plan`, `moai update-plan` 등(각 명령은 plan/apply 2단계이며 plan digest 없이는 아무것도 쓰지 않습니다). 일상 사용 경로가 아닙니다.

### 3.x 명령에 대해 (fallback)

기존 `moai init`·`moai update`(Go 3.x)는 구 세대의 명령입니다. RC1을 이미 설치했다면 3.x 명령을 새 절차로 쓰지 마세요 — 두 세대의 업데이트 통로는 분리되어 있습니다(지원표와 릴리스 체크리스트 참조). 마이그레이션 도구는 별도 제공 예정입니다.

## 세션 론처 (저장소 개발용, 현행 — DR-08 #2145)

`-l|-a`는 다시 쓰입니다(사용자 결정 2026-11-25, SPEC-20261125-agent-team-session-names; 2026-11-09의 D-1 폐기 결정을 뒤집습니다). 역할 플래그가 없는 실행의 기본 권한 모드는 호스트마다 다릅니다: `moai cc`는 auto mode, `moai glm`은 bypass(`--dangerously-skip-permissions`)입니다(사용자 결정 2026-11-23의 auto 기본값을 `cc`에 남기고, `glm`은 사용자 결정 2026-11-24, SPEC-20261124-launcher-retire-role-flags D-7). 기본값은 사용자 전용 파일 `~/.moai-cli/config/launcher.yaml`의 `launcher.role_access` 키(`auto|full|prompt`)로 바꿀 수 있다(Issue #2413, 아래). 아래 표는 오늘 기준 동작을 기록합니다.

`moai`(호환 별칭 `moai-cli`)는 저장소 개발 세션을 띄우는 론처를 갖고 있다. 사용자 진입(`/moai`)과는 별개인 개발 도구다. 아래 표는 `src/main.rs`의 `Cli`(45–56행), `LaunchOpts`(58–79행), `HostLaunch`(81–96행), `Command::Cc|Glm`(519–524행)와 `src/session.rs` 머리말을 기준으로 한다. 2026-11-09에 base `604792666` 빌드로 처음 확인했고, 권한·이름 규칙은 2026-11-10 #2359 빌드의 `--dry-run`과 시험으로 다시 확인했다.

| 형태 | 동작 |
| --- | --- |
| `moai cc\|glm` (역할 없음) | 평범한 host 세션. 역할을 붙이지 않고, `--name`이 없으면 세션 이름도 붙이지 않는다(#2053). 권한 관련 플래그도 붙이지 않아, Claude Code는 사용자·프로젝트 설정의 모드로 시작한다(#2359). Claude 계열은 `--model opus`를 쓰고, effort 기본값은 호스트마다 다르다: `cc`는 `medium`, `glm`은 `high`(사용자 결정 2026-11-24, SPEC-20261124-launcher-retire-role-flags). `--advisor`는 붙이지 않는다(아래 `launcher.advisor`). |
| `moai cc\|glm --name N` | 세션 이름을 `N`으로 붙인다. 그 밖에는 역할 없는 형태와 같고, 역할은 세션 안에서 `/moai agent`(레인)·`/moai leader`(조정 창)로 들어간다. 권한은 `cc`가 auto mode, `glm`이 bypass(호스트 기본값; 아래 `launcher.role_access`). |
| `moai cc\|glm -l` (`--leader`) | 조정 창. 세션 이름 `leader-<프로젝트 이름>`(프로필의 `project_id`, 무작위 형태면 기본 체크아웃 폴더 이름, SPEC-20261127-launcher-session-names), 권한 우회(`--dangerously-skip-permissions`, `--permissions`로 확인 모드 유지), 첫 입력 `/moai leader`. 그 이름의 살아 있는 `.live` 잠금이 이미 있으면 거절한다. `--name`이 이름을 바꾼다. advisor는 붙이지 않는다. |
| `moai cc\|glm -a` (`--agent`) | 레인. 세션 이름 `agent-<영어 이름>`(런처 목록에서 이 클론의 잠금도 등록 항목도 없고 이 컴퓨터의 다른 런처도 쥐지 않은 이름 가운데 무작위로 고른 하나, 목록이 다 차면 `--name`을 안내하며 거절), 권한 우회, 첫 입력 `/moai agent`. 모델·effort는 `--model`·`--effort` > 저장된 agent 선택 > 기본 표의 agent 행(`claude`는 `claude-sonnet-5-5`·`high`, `glm`은 `glm-5.3-flash`·`max`). `cc`에서는 `--advisor opus`도 붙는다(`--advisor`·`--no-advisor`·`launcher.advisor`가 먼저; 주 모델이 더 높으면 기본 advisor는 빠지고 이유 한 줄이 남는다). |
| `-l`과 `-a`를 함께, 또는 역할 플래그와 PROMPT를 함께 | 한 문장으로 거절한다. host도 잠금도 파일도 만들지 않는다. |
| `--name N` | 세션 이름. |
| `--model M` | 주 모델. Claude 계열 기본 `opus`. |
| `--effort E` | Claude effort(`low\|medium\|high\|xhigh\|max\|ultracode`). 기본은 호스트마다 다르다: `cc`는 `medium`, `glm`은 `high`. 순서는 `--effort` > 저장된 `launcher.models.<host>` 값 > 호스트 기본값이다. |
| `--advisor A` | Claude 계열에 `--advisor A`를 붙인다(`fable\|opus\|sonnet` 또는 `claude-*` 전체 id). 기본은 붙이지 않으며, 사용자 설정 `launcher.advisor`(아래)보다 우선한다. 사용 크레딧으로 과금될 수 있다. |
| `--no-advisor` | 사용자 설정 `launcher.advisor`가 켠 advisor를 이번 실행에서 끈다. `--advisor`와 함께 주면 거부한다. |
| `--permissions` | 확인 모드를 유지한다. Claude 계열은 권한 플래그(`--permission-mode auto`, `--dangerously-skip-permissions`)를 하나도 붙이지 않아 host가 자기 설정의 모드로 시작한다. 사용자 설정 `launcher.role_access`(아래)보다 우선한다. |
| `--dry-run` | host를 띄우지 않고 조립한 명령을 JSON으로 출력한다. |
| `[PROMPT]` | 첫 프롬프트 텍스트. 받는 host가 없어, Claude 계열에는 넘기지 않았다는 note가 붙는다. |

host별 argv:

- `cc`: `claude --name <name> --model opus --effort medium --permission-mode auto` (`cc`의 기본값; `--permissions` 또는 `launcher.role_access: prompt`(아래)면 이 플래그가 빠지고, `launcher.role_access: full`이면 `--dangerously-skip-permissions`로 바뀐다). 역할 없는 형태는 `claude --model opus --effort medium --permission-mode auto`로, 이름만 없다. `glm`은 z.ai 게이트웨이 뒤라 auto mode를 고르면(`launcher.role_access: auto`) 분류기 요금 안내가 뜰 수 있어, 그때만 노트가 `docs/troubleshooting.md` 5절을 가리킨다(설정은 바꾸지 않는다). `glm`의 기본은 `--permission-mode auto` 대신 `--dangerously-skip-permissions`이고, bypass가 막는 것과 안 막는 것은 같은 문서 8절에 있다. `--advisor A` 또는 `launcher.advisor`(아래)가 있으면 `--effort` 뒤에 `--advisor A`가 붙는다.
- `glm`: `cc`와 같은 Claude Code argv에서 effort만 `--effort high`로 바꾸고(`glm`의 effort 기본값은 `high`, `cc`는 `medium`), Go 도구와 같은 환경 변수 묶음(`ANTHROPIC_BASE_URL`·`API_TIMEOUT_MS`, 모델 별칭 `ANTHROPIC_DEFAULT_*_MODEL`(opus·sonnet·fable은 `glm-5.4[1m]`(사용자 결정 2026-11-23, SPEC-20261123-glm-54-model), haiku는 `glm-5.3-flash`), 1M 컨텍스트 선언 `CLAUDE_CODE_AUTO_COMPACT_WINDOW`·`CLAUDE_CODE_MAX_CONTEXT_TOKENS`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, z.ai MCP용 `Z_AI_API_KEY`)을 더해 z.ai로 보낸다. 토큰은 셸의 `ANTHROPIC_AUTH_TOKEN`을 쓰고, 없으면 키 파일 `~/.moai/.env.glm`(Go 도구와 함께 쓰는 전역 저장소), 이어서 `~/.moai-cli/.env.glm`에서(둘 다 읽기만 하며, 65536바이트 이하의 일반 파일일 때만 연다) `ANTHROPIC_AUTH_TOKEN=` 또는 `GLM_API_KEY=` 줄을 찾아 host 프로세스에만 넘긴다(#2359). 론처는 값을 출력·저장하지 않고 키 파일을 만들거나 고치지 않는다. 어디에도 없으면 `blocked:` 메시지와 exit 1로 끝난다.

권한 모드는 사용자 결정 2026-11-23(SPEC-20261123-launcher-auto-mode)을 따르고, `-l|-a` 세션은 사용자 결정 2026-11-25(SPEC-20261125-agent-team-session-names)로 다시 권한 우회가 기본이고, `--permissions`가 확인 모드를 유지한다. `launcher.role_access`는 역할 플래그가 없는 실행의 권한 모드를 정한다. 모든 실행은 host와 프로젝트 `.claude/settings.json`의 `permissions` 키에 관계없이 호스트 기본(`cc`는 auto mode `--permission-mode auto`, `glm`은 bypass `--dangerously-skip-permissions`; 사용자 결정 2026-11-24)으로 시작하고, `--permissions`를 주면 권한 플래그를 하나도 붙이지 않아 host가 자기 설정의 모드로 뜬다. 기본값은 `session::ACCESS_DEFAULT` 한 곳에 있고, 사용자 결정 Issue #2413 옵션 A(댓글 6778672356, 설계 6781812681)로 사용자 전용 파일 `~/.moai-cli/config/launcher.yaml`의 `launcher.role_access` 키(`auto|full|prompt`)가 이 기본값을 고른다. `full`은 우회(`--dangerously-skip-permissions`)이고 `cc`에서는 사용자가 직접 고를 때만 쓰이며 `glm`의 기본이기도 하고, `prompt`는 권한 플래그를 붙이지 않는다. 이 파일은 프로젝트 profile `.yaml`과 같은 명시적 JSON 부분집합 규약(`serde_json`)으로만 읽으며 일반 YAML 파서는 쓰지 않는다. 우선순위는 CLI `--permissions` > 사용자 키 > 호스트 기본값(`cc`는 `auto`, `glm`은 `full`)이고 프로젝트 설정은 이 선택에 관여하지 않으며, 역할 없는 세션도 같은 순서로 정한다. 파일이나 키가 없으면 호스트 기본으로 시작하고, 홈 디렉터리는 `HOME`, Windows에서 `HOME`이 없거나 비었으면 `USERPROFILE`로 정하며(`key_vault::home` 규약) 둘 다 없을 때만 사용자 파일이 없는 것으로 본다. 값을 알 수 없으면 — 모르는 값, 문자열이 아닌·`null`인 `role_access`, `null`이거나 객체가 아닌 `launcher`·루트, 잘못된 JSON 문법, 읽을 수 없거나 일반 파일이 아닌 기존 파일(디렉터리·심볼릭 링크·FIFO) — host 시작 앞에서 실행을 거부하며, 설정 내용이나 원본 오류 값을 그대로 출력하지 않고 `--permissions`를 줬더라도 조용히 `auto`로 빠지는 fallback은 없다. 상세 계약과 예시는 docs/rust-product-path.md의 해당 문단이 정한다. 옛 `session` 하위 명령(Claude 전용)은 바뀌지 않았다.

advisor는 기본으로 붙이지 않는다(사용자 결정 2026-11-12, Issue #2476). 2026-11-08 결정(#2053 HOLD-L1 = A, #2062) 가운데 `--advisor fable` 기본값은 이 결정으로 대체되었고, `--model opus` 기본값과 `cc`의 `--effort medium` 기본값은 그대로다(`glm`의 effort 기본값은 `high`). Fable advisor가 사용 크레딧으로 과금되는 계정에서 크레딧을 설정하지 않았으면 host가 시작하자마자 exit 1로 끝났기 때문이다. advisor를 켜는 방법은 두 가지이고 우선순위는 CLI `--advisor A` > 같은 사용자 전용 파일의 `launcher.advisor` 키 > 없음이다. 키 값은 `none`(없는 것과 같다)·`fable`·`opus`·`sonnet`·`claude-*` 전체 id 중 하나다(예: `{"launcher":{"advisor":"fable"}}`). `--no-advisor`는 키가 켠 advisor를 이번 실행에서 끄고, `--advisor`와 함께 주면 거부한다. 키는 Claude 계열에만 적용된다. `glm`은 advisor를 빼고 그렇다는 note를 남긴다(advisor는 Anthropic API 전용). advisor가 주 모델 이상이어야 한다는 기존 검사도 그대로라, 기본 `opus`에서 `launcher.advisor: sonnet`은 거부된다. 잘못된 값(모르는 문자열·대소문자가 다른 값·빈 문자열·`null`·숫자·배열·객체, 깨진 JSON 문법)은 `role_access`와 같은 방식으로 host 시작 앞에서 거부하고, `--advisor`·`--no-advisor`를 줬더라도 검사하며, 값은 출력하지 않는다. 프로젝트 설정으로는 켤 수 없다. advisor를 붙인 실행의 notes에는 사용 크레딧으로 과금될 수 있다는 줄이, 키에서 온 경우에는 그 출처가 함께 나온다. 상세 계약은 docs/rust-product-path.md의 #2476 문단이 정한다.

`agent-N` 번호(#2361, #2971)는 폐기했습니다(사용자 결정 D-1 2026-11-09와 D-4, SPEC-20261124-launcher-retire-role-flags). 런처는 번호를 매기지 않습니다. 이름은 `--name`이나 역할 플래그가 붙입니다(`-l`은 `leader-<프로젝트 이름>`, `-a`는 `agent-<영어 이름>`의 목록에서 비어 있는 이름 가운데 무작위로 고른 하나, SPEC-20261127-launcher-session-names; 이 컴퓨터에서 이름이 겹치지 않도록 `<HOME>/.moai/session-names/<이름>.live` 파일의 배타 잠금을 클론의 잠금보다 먼저 쥡니다). 역할 플래그도 `--name`도 없으면 이름도 붙이지 않습니다. 리더나 Orca 프리셋이 레인을 띄울 때는 `-a`나 `--name agent-<영어 이름>`을 줍니다. 이름이 겹치는 것은 `session-register`가 막고(한 이름을 임대가 끝나지 않은 다른 instance가 쥐고 있으면 거부합니다), `--name`을 준 세션은 host가 끝날 때까지 `<clone의 공통 git 디렉터리>/.moai-cli/sessions/<name>.live` 파일의 배타 잠금을 쥐어 살아 있음을 알립니다(SPEC-20261120-launcher-liveness). 폴더 위치와 `session-register` 항목은 그대로입니다. 옛 번호 규칙(살아 있는 가장 큰 번호에 1을 더함, 번호 1의 경고 줄)은 이 SPEC 이전의 동작입니다.

## 채택 시 확인 항목

- [x] 3.x 명령을 필수 단계로 안내하는 문장 0건(`grep -nE 'moai (init|update)\b' docs/*.md README.md | grep -v 'fallback\|3\.x'` = 0)
- [x] `/moai` 예시가 진입의 기본
- [ ] README 본문 반영 — README는 기존 사용자 영역 유지(이 초안 채택 시 별도 편집)
