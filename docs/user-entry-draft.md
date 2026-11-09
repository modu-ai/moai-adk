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

호스트 세션은 역할 없이 엽니다 — `moai-cli cc|glm|codex`는 평범한 host 세션이고, 세션 안에서 `/moai-cli --leader-mode`·`/moai-cli --agent-mode`로 MoAI 역할을 정합니다(호환 별칭 `--mode leader|agent`). 리더·에이전트 론처 플래그 `-l|-a`(`moai-cli -l|-a`, `moai-cli cc|glm|codex -l|-a`)는 deprecated입니다: 사용하면 경고가 뜨고 기존처럼 동작하지만, 새 세션은 역할 없이 열고 `/moai-cli --leader-mode`·`--agent-mode`를 쓰는 것을 권합니다.

### 수동 명령 (명시적 fallback)

자연어 진입이 불가능한 환경에서만 내부 실행 도구를 직접 씁니다: `moai-cli generation-plan`, `moai-cli update-plan` 등(각 명령은 plan/apply 2단계이며 plan digest 없이는 아무것도 쓰지 않습니다). 일상 사용 경로가 아닙니다.

### 3.x 명령에 대해 (fallback)

기존 `moai init`·`moai update`(Go 3.x)는 구 세대의 명령입니다. RC1을 이미 설치했다면 3.x 명령을 새 절차로 쓰지 마세요 — 두 세대의 업데이트 통로는 분리되어 있습니다(지원표와 릴리스 체크리스트 참조). 마이그레이션 도구는 별도 제공 예정입니다.

## 세션 론처 (저장소 개발용, 현행 — DR-08 #2145)

`-l|-a` 형태는 #2164부터 deprecated입니다(위의 호스트 세션과 역할 참조) — 경고가 출력되고 동작은 유지됩니다. 아래 표는 오늘 기준 동작을 기록합니다.

`moai-cli`는 저장소 개발 세션을 띄우는 론처를 갖고 있다. 사용자 진입(`/moai`·`$moai`)과는 별개인 개발 도구다. 아래 표는 `src/main.rs`의 `Cli`(45–56행), `LaunchOpts`(58–79행), `HostLaunch`(81–96행), `Command::Cc|Glm|Codex`(519–524행)와 `src/session.rs` 머리말(1–19행)을 기준으로, 2026-11-09에 base `604792666`의 빌드로 `--dry-run`을 실행해 확인했다.

| 형태 | 동작 |
| --- | --- |
| `moai-cli cc\|glm\|codex` (역할 없음) | 평범한 host 세션. 역할·세션 이름을 붙이지 않고 `--dangerously-skip-permissions`도 붙이지 않는다(#2053). Claude 계열은 `--model opus --effort medium --advisor fable` 기본값을 쓴다. |
| `moai-cli cc\|glm\|codex -l` (`--leader`) | 리더 세션. 이름 `leader`. |
| `moai-cli cc\|glm\|codex -a` (`--agent`) | 에이전트 세션. 이름은 `--name`, 없으면 `agent-1`과 안내 note(자동 번호는 HOLD). |
| `moai-cli -l` / `moai-cli -a [cc\|glm\|codex]` | 위와 같은 요청의 짧은 형태. 기본 host는 `cc`. |
| `--name N` | 세션 이름. |
| `--model M` | 주 모델. Claude 계열 기본 `opus`, codex는 지정했을 때만 `-m M`. |
| `--effort E` | Claude effort(`low\|medium\|high\|xhigh\|max\|ultracode`), 기본 `medium`. |
| `--no-advisor` | `--advisor fable`을 붙이지 않는다(Claude 계열). |
| `--permissions` | `-l\|-a`에서 `--dangerously-skip-permissions`를 붙이지 않는다(권한 확인 유지). |
| `--dry-run` | host를 띄우지 않고 조립한 명령을 JSON으로 출력한다. |
| `[PROMPT]` | 첫 프롬프트 텍스트(codex만). |

host별 argv:

- `cc`: `claude --name <name> --model opus --effort medium --advisor fable --dangerously-skip-permissions` (`-l|-a`의 기본).
- `glm`: `cc`와 같은 Claude Code argv에, 환경 변수 `ANTHROPIC_BASE_URL`·`API_TIMEOUT_MS`를 더해 z.ai로 보낸다. `ANTHROPIC_AUTH_TOKEN`이 셸에 있어야 하며, 론처는 있는지만 확인하고 값은 출력·저장하지 않는다. 없으면 `blocked:` 메시지와 exit 1로 끝난다.
- `codex`: `codex -s workspace-write -a on-request [-m <model>] [PROMPT]`. Codex에는 세션 이름 플래그가 없어 이름은 첫 프롬프트(`MoAI session name: <name>`)로 전한다. Codex 세션에는 RC1의 세션 간 메시지 채널(ListAgents·SendMessage)이 없다. advisor와 permission skip은 Claude 계열에만 붙는다.

`-l|-a`에 `--dangerously-skip-permissions`를 기본으로 붙이는 것은 결정 기록(`docs/decisions-2026-11-08.md` 18:0x "임시 기본, 추후 설정에서 선택")과 일치한다. 20:3x·20:5x 결정은 역할 없는 경로와 실측에 새 bypass를 더하지 않는다는 것이고, 실물도 역할 없는 경로에는 붙이지 않으며 `--permissions`로 끌 수 있다. 다만 이 체크아웃의 `.claude/settings.local.json`이 `bypassPermissions`이면 론처 플래그와 관계없이 권한 확인이 생략된다(결정 기록 20:3x 훅 범위 확인).

## 채택 시 확인 항목

- [x] 3.x 명령을 필수 단계로 안내하는 문장 0건(`grep -nE 'moai (init|update)\b' docs/*.md README.md | grep -v 'fallback\|3\.x'` = 0)
- [x] `/moai`·`$moai` 예시가 진입의 기본
- [ ] README 본문 반영 — README는 기존 사용자 영역 유지(이 초안 채택 시 별도 편집)
