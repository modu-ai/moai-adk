# 로컬 RC1 사용 키트 — 내 Mac의 다른 프로젝트에서 MoAI rc.1 써 보기 (#2262)

- 대상: 운영자 본인의 개발 Mac(Darwin/arm64), Claude Code·Codex(P04). 공개 배포가 아닙니다.
- 작성 2026-11-09. 이 문서의 실측은 모두 **임시 HOME·임시 프로젝트**에서 한 것입니다. 사용자 실제 HOME·실제 프로젝트 적용과 실제 host 실행은 이 카드에서 하지 않았습니다.

| 단계 | 이 카드에서 | 상태 |
| --- | --- | --- |
| 1 계획(이 문서) | 작성 | 완료 |
| 2 스테이징 빌드 | 실행 | passed |
| 3 격리 시험(임시 HOME·임시 프로젝트) | 실행 | passed. 처음 실측에서 결함 1건 발견(#2268), #2291 재실측에서 해소 확인(2절) |
| 4 사용자 실제 프로젝트 적용 | 실행하지 않음 | 사용자 확인 대상. 명령만 아래에 둡니다 |
| 5 사용자 실제 HOME 적용(`update-apply`) | 실행하지 않음 | 사용자 확인 대상. #2268(PR #2278) 병합 뒤에는 기존 비관리 `moai` skill 폴더가 있으면 `update-plan`이 `blocked`, `update-apply`는 0 write(3절) |
| 6 실제 host(Claude Code·Codex) 실행 | 실행하지 않음 | T2 not_run(#2014) |
| 7 공개 배포(tag·release) | 실행하지 않음 | blocked(사용자 승인) |

## 1. 스테이징 빌드

| 항목 | 값 |
| --- | --- |
| source | 브랜치 `issue/2262-rc1-release-prep` `7b702b255bc976aa88c6f3deb51287566d23e7f9` = `v2026.11.0` tip `5db3069a94effe4ae330137dc96ff3a01c5be222` + 이 카드의 문서 커밋(코드 차이 0) |
| 버전 | `moai-cli --version` → `moai-cli 2026.11.0-rc.1` |
| 바이너리 | `$CARGO_TARGET_DIR/release/moai-cli`, 이번 빌드는 `/tmp/moai-rc-target-agent-9/release/moai-cli` |
| sha256 | `72e614a196a7336ae6c43fcc0a62b83d5e01ed7b4c237795042e0e1b2dce5d62`(2026-11-09T05:15:50Z 빌드) |

절차(저장소의 깨끗한 checkout에서 실행합니다):

```sh
git fetch origin v2026.11.0
git worktree add /tmp/moai-rc1-stage origin/v2026.11.0     # 정확한 SHA를 쓰려면 origin/v2026.11.0 대신 SHA
cd /tmp/moai-rc1-stage && git rev-parse HEAD                # source SHA를 기록
CARGO_INCREMENTAL=0 CARGO_TARGET_DIR=/tmp/moai-rc1-stage-target cargo build --offline --locked --release
/tmp/moai-rc1-stage-target/release/moai-cli --version
shasum -a 256 /tmp/moai-rc1-stage-target/release/moai-cli   # 바이너리 hash를 기록
```

로컬 시험 빌드마다 번호를 붙이려면 빌드할 때 `MOAI_LOCAL_BUILD=<n>`을 주세요(#2722 AC-5, 리더 결정 6855056793). `MOAI_LOCAL_BUILD=3 cargo build …`로 만든 바이너리는 `moai-cli --version`과 상태줄에 `2026.11.0-rc.1+local.3`으로 보입니다. `+local.<n>`은 semver 빌드 메타데이터라서 릴리스 태그 `rc.<n>`과 섞이지 않습니다. `<n>`은 1부터 4294967295까지의 십진수만 받습니다. 부호, 앞자리 0, 공백이 들어가면 빌드가 멈추고, 비우거나 주지 않으면 지금처럼 메타데이터 없이 빌드됩니다. 설치·갱신 기록의 패키지 버전에는 번호를 붙이지 않습니다.

바이너리는 PATH에 넣지 말고 경로로 부르세요. 이름이 `moai`가 아니라서 Go `moai`와 섞이지 않습니다. `~/.moai-cli/bin/`에 설치하는 것은 04-A installer의 몫이고, 이 키트는 설치하지 않습니다.

## 2. 격리 시험 (실측)

재현: `bash tools/local-rc1-isolation.sh <BIN> <repo root> <out dir>`. 스크립트는 4절과 같은 순서(project → generation → agents)로 돌고, `WF_REF` 우회는 없앴다(#2291). 실측은 2026-11-09, `v2026.11.0` tip `496cdedbb` 위 #2291 브랜치의 debug 빌드로 했다. 실제 HOME 쓰기 0은 `sandbox-exec -p '(version 1)(allow default)(deny file-write* (subpath "<실제 HOME>"))'` 아래에서 실행해 확인했다(같은 프로필로 HOME 아래 `touch`를 시도하면 `Operation not permitted`). 시험은 `mktemp -d`로 만든 HOME과 프로젝트에서 돌렸고, 실제 HOME은 넘기지 않았습니다. 실제 기계를 흉내 내려고 다음을 먼저 심었습니다: `~/.claude/skills/moai/SKILL.md`와 `~/.agents/skills/moai/SKILL.md`(RC1이 아닌 기존 skill), 다른 skill 하나, `~/.moai/`, `~/.codex/AGENTS.md`, 프로젝트의 `.moai/`, `.claude/settings.json`, 사용자 내용이 든 `AGENTS.md`, `AGENTS.override.md`.

| 순서 | 명령 | 결과 |
| --- | --- | --- |
| a | `moai-cli project-plan <P>` (stdin: profile 요청) → `project-apply <P> --plan-digest … --permission-reference …` (stdin: plan) | exit 0 / 0. 새 파일 `.moai-cli/project.yaml`, `manifests/profile-baseline.json`, receipt |
| b | `moai-cli generation-plan <P>` (stdin: 4절 `generation-request.json`, `workflow_id` `kit-demo`, `workflow_reference` `.moai-cli/workflows/kit-demo/`) → `generation-apply` | exit 0 / 0. `state: applied`, 새 파일 `.moai-cli/workflows/kit-demo/SKILL.md`·`references/workflow.md`, `manifests/generation.json`. `AGENTS.md`의 사용자 줄(`keep this line`)은 그대로이고 그 아래 `<!-- moai:managed:begin id=moai-workflow -->` 블록이 1개 추가됨. 원본은 `.moai-cli/receipts/agents-apply-….before/AGENTS.md`에 보관 |
| c | `moai-cli agents-plan <P>` (stdin: 같은 `workflow_reference`) → `agents-apply` | exit 0 / 0. `AGENTS.md` action `unchanged`, apply `state: unchanged`(generation이 이미 맞춘 블록을 다시 맞춰도 바뀌지 않음). host note: `AGENTS.override.md present; preserved; …` |
| d | `moai-cli update-plan <H> <repo>/skills/moai` | exit 0. 두 host 모두 `preflight: blocked`, `blocked_reason: unmanaged existing package: 1 file(s) present in the projection root that install receipt … does not own (SKILL.md)` |
| e | `moai-cli update-apply <H> <repo>/skills/moai --plan-digest … --permission-reference …` (stdin: plan) | exit 0, `status: blocked`, 두 host 모두 `files: []`, `reload_basis: no files written` |
| f | `moai-cli recovery-plan <H>` | exit 0, 두 host `receipt_status: none`, `files: []` |

보존 확인(작업 전후 `find … | shasum -a 256` 비교): 실행 전부터 있던 파일 중 바뀌거나 사라진 것은 HOME 0건이고, 프로젝트는 `AGENTS.md` 1건뿐입니다(관리 블록 추가). 새로 생긴 HOME 파일도 0건입니다. `AGENTS.override.md`, `.claude/settings.json`, `.moai/`, `~/.moai/`, `~/.codex/AGENTS.md`, 기존 `moai` skill 두 개, 다른 skill은 모두 그대로였습니다. 블록 되돌리기(`apply-recover`)는 이번 실측 순서에 넣지 않았습니다. 명령은 4절 3)에 있고, 이전 실측(#2274 전)에서 `state: reverted`와 원본 hash 복귀를 확인했습니다.

## 3. 사용자 실제 HOME에 대해 알아야 할 사실

- 2026-11-09 읽기 전용 `ls` 결과, 사용자 HOME에는 이미 `~/.claude/skills/moai`(2026-10-22)와 `~/.agents/skills/moai`(2026-10-29)가 있습니다. RC1이 투영하려는 경로와 정확히 같습니다.
- #2268(PR #2278) 병합 뒤에는 이런 HOME에서 `update-plan`이 두 host를 `blocked`(`unmanaged existing package … (SKILL.md)`)로 보고하고, `update-apply`는 아무 파일도 쓰지 않습니다(위 d·e). 이전처럼 기존 폴더에 RC1 파일이 섞여 들어가는 일은 없습니다.
- 규칙: `update-plan` 결과에 `blocked`나 `conflict`가 하나라도 있으면 apply하지 않습니다.
- 안전한 대안(모두 사용자가 직접 결정하고 실행하며, 자동이 아닙니다):
  1. **별도 설치 위치로 시험**: 실제 host를 띄우지 않는 확인은 위 2절처럼 임시 HOME에서만 합니다.
  2. **백업한 뒤 교체**: 사용자가 기존 `~/.claude/skills/moai`, `~/.agents/skills/moai`를 다른 이름으로 옮겨 보관합니다. 그다음 `update-plan`에 `blocked`·`conflict`가 0인 것을 확인한 뒤에만 apply합니다. 기존 skill은 Go 세대 MoAI가 쓰고 있을 수 있으니, 옮기면 그쪽 동작이 바뀝니다.

## 4. 사용자 실제 프로젝트 적용 순서 (사용자 확인 대상 — 이 카드는 실행하지 않음)

`<P>`는 대상 프로젝트, `BIN`은 1절의 바이너리입니다. 단계마다 plan을 사람이 읽고 확인한 다음에만 apply합니다.

```sh
BIN=/tmp/moai-rc1-stage-target/release/moai-cli
# 0) 시작 전 상태를 기록(복구 대조용)
(cd <P> && git status --short && find . -path ./.git -prune -o -type f -print0 | sort -z | xargs -0 shasum -a 256 > /tmp/moai-kit-before.sha256)

# 1) SETUP: profile 미리보기 → 확인 → 적용
"$BIN" project-plan <P> < profile-request.json > project-plan.json      # 읽기 전용. result의 파일별 action을 확인
"$BIN" project-apply <P> --plan-digest <project-plan.json의 result.plan_digest> --permission-reference <내 확인 메모> < project-plan.result.json

# 2) 맞춤 workflow + AGENTS.md 관리 블록: 미리보기 → 확인 → 적용 (#2274)
#    workflow_reference는 generation이 직접 만드는 .moai-cli/workflows/<id>/ — 미리 만들 필요 없음
"$BIN" generation-plan <P> < generation-request.json > generation-plan.json   # action이 conflict면 멈춤
"$BIN" generation-apply <P> --plan-digest <…> --permission-reference <…> < generation-plan.result.json
# (나중에 AGENTS.md 블록만 다시 맞출 때) 같은 workflow_reference로
"$BIN" agents-plan <P> < agents-request.json > agents-plan.json
"$BIN" agents-apply <P> --plan-digest <…> --permission-reference <…> < agents-plan.result.json

# 3) 되돌리기: apply 결과의 receipt id로
"$BIN" apply-recover <P> --receipt <receipt id> --permission-reference <…>
```

- `profile-request.json` 예: `{"schema_version":2,"project_id":"<id>","purpose":"<목적>","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"<rev>"}`
- `generation-request.json` 예: `{"protocol_version":1,"language":"ko","workflow_reference":".moai-cli/workflows/<id>/","local_file":"none","workflow_id":"<id>","native_roles":[],"skill_projections":[]}`
- `agents-request.json`: `{"protocol_version":1,"language":"ko","workflow_reference":".moai-cli/workflows/<id>/","local_file":"none"}`. `agents-plan`을 단독으로 부를 때는 이 디렉터리가 이미 있어야 합니다(generation-apply 뒤). `local_file`은 `none|ignored|tracked`만 받습니다.
- `apply-recover`는 apply 뒤에 바뀐 파일은 건너뛰고 덮지 않습니다(`skipped`에 나옵니다).
- 2단계의 G2(일반 프로젝트에서 막힘)는 #2274로 해소: generation이 자기 출력 위치를 `workflow_reference`로 받으면 그 디렉터리를 미리 요구하지 않습니다. `.moai-cli/` 같은 우회 값은 쓰지 마세요(AGENTS.md가 엉뚱한 곳을 가리킴).

## 5. 보존 규칙

- Go `moai` 바이너리, `.moai/`, `~/.moai/`는 읽기만 하고 덮지도 지우지도 않습니다. 격리 시험에서 바이트가 그대로인 것을 확인했습니다.
- 개인 `AGENTS.md`의 사용자 내용과 `AGENTS.override.md`는 덮지 않습니다. RC1은 `AGENTS.md`에 표시된 관리 블록만 넣고 고칩니다.
- 기존 skill과 설정(`~/.claude/skills/*`, `~/.agents/skills/*`, `.claude/settings.json`)은 덮지 않습니다. 같은 경로에 다른 내용이 있으면 `conflict`로 보고하고 원본을 보존합니다. RC1 receipt가 소유하지 않은 기존 skill 폴더가 있으면 그 host 전체가 `blocked`이고 apply는 0 write입니다(#2268).

## 6. host별 instruction·skill 경로 (공식 문서, 2026-11-09 GET 확인)

| 항목 | Claude Code | Codex |
| --- | --- | --- |
| 프로젝트 instruction | `CLAUDE.md`. `AGENTS.md`는 작업 디렉터리와 그 위에 `CLAUDE.md`·`.claude/CLAUDE.md`·`CLAUDE.local.md`가 **없을 때만** 읽습니다(v2.1.277+). `~/.claude/CLAUDE.md`는 이 판단에 들어가지 않습니다 | 프로젝트 루트부터 현재 디렉터리까지 디렉터리마다 `AGENTS.override.md` → `AGENTS.md` → fallback 이름 순으로 **한 디렉터리에 하나만** 읽고, 루트부터 이어 붙입니다 |
| `AGENTS.override.md` | **읽지 않음**(`AGENTS.local.md`, `.agents/` 아래도 읽지 않음) | 같은 디렉터리에 있으면 그 디렉터리의 `AGENTS.md`는 **무시됩니다** |
| 전역 instruction | `~/.claude/CLAUDE.md` | `~/.codex/AGENTS.override.md`가 있으면 그것, 없으면 `~/.codex/AGENTS.md`(`CODEX_HOME` 기본 `~/.codex`) |
| 사용자 skill | `~/.claude/skills/<name>/SKILL.md` | `$HOME/.agents/skills` |
| 프로젝트 skill | `.claude/skills/<name>/SKILL.md`(하위 디렉터리 `.claude/skills/`도 있음) | 현재 디렉터리부터 저장소 루트까지 각 `.agents/skills` |
| 같은 이름 | 대소문자만 다르면 같은 이름으로 보고 로컬 skill을 씁니다 | 합치지 않고 둘 다 선택지에 나올 수 있습니다 |

이 표에서 바로 나오는 주의점:
- 대상 프로젝트에 `AGENTS.override.md`가 있으면, Codex는 그 디렉터리의 `AGENTS.md`(RC1 관리 블록 포함)를 읽지 않습니다.
- 대상 프로젝트에 `CLAUDE.md`가 있으면, Claude Code는 `AGENTS.md`를 읽지 않습니다. `CLAUDE.md`에서 import하거나 `/config`의 Project instructions 설정이 필요합니다.
- 공식 문서에 없는 이름이 자동으로 인식된다고 가정하지 않습니다.

출처(HTTP 200, 2026-11-09T05:19Z): https://code.claude.com/docs/en/memory , https://code.claude.com/docs/en/skills , https://learn.chatgpt.com/docs/agent-configuration/agents-md , https://learn.chatgpt.com/codex/build-skills

## 7. gap

| # | 내용 | 상태 |
| --- | --- | --- |
| G1 | `update-plan`이 루트 `SKILL.md` conflict를 `preflight: ok`로 두고, `update-apply`가 기존 폴더에 나머지 파일을 씀 | #2268(PR #2278)에서 해소: 비관리 기존 폴더면 `update-plan` `blocked`, `update-apply` 0 write(2절 d·e 재실측) |
| G2 | `agents-plan`은 프로젝트 안에 `workflow_reference` 디렉터리가 있어야 하는데, 일반 프로젝트에 그것을 만드는 RC1 단계가 없음(시험 fixture도 손으로 만듦) | #2274에서 해소: `generation-plan`이 자기 출력 위치(`.moai-cli/workflows/<id>/`)를 받으면 미리 요구하지 않음(시험 `tests/generation_own_reference.rs`). 단독 `agents-plan`은 계속 거부 |
| G3 | `recovery-plan`·`recovery-apply`는 `pending` host만 되돌림. `committed` 설치를 지우는 명령은 없음 | 설계 범위. 기록만 |
| G4 | 실제 host가 투영된 skill을 로드하는지 | not_run(T2, #2014) |
| G5 | 리더 쪽 스테이징 빌드 결과와 대조 | 아직 받지 못함 — 받으면 1절에 인용 |
