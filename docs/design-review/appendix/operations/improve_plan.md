# `improve-plan`

operation `improve.plan` · 코드 `src/ops/improve_plan.rs` · SPEC-20261121-self-improvement-loop (조각 S2, REQ-12~REQ-15·REQ-26~REQ-31, AC-10~AC-14)

`improve-plan <project> --target <상대 경로> --item <JSON> [--item <JSON> …]` 는 프로젝트가 만든 스킬·에이전트 한 파일에 "배운 점" 항목 1~3개를 넣는 **제안만** 보여 주는 읽기 전용 명령이다. 아무것도 쓰지 않고(`writes:false`, 파일·폴더·장부 행·수정 시각 모두 그대로), 적용은 뒤 조각(S4)이 한다. 기준 시각이나 프로젝트 폴더는 결과에 영향을 주지 않는다.

**대상은 등급 A 뿐이다.** 프로젝트 상태 폴더 `manifests/generation.json` 의 `owned_files` 가 소유한 항목 가운데, `section_id` 가 `generation:project-skill:claude:` 또는 `generation:native-role:claude:` 로 시작하고 경로가 `.claude/skills/<moai-prj-…|prj-…>/SKILL.md` 또는 `.claude/agents/[moai/]prj-….md` 모양인 파일만 받는다. 그 밖은(매니페스트에 없는 파일, 설치된 `moai` 스킬, 설정·훅·`.mcp.json`, `CLAUDE.md`, `CLAUDE.local.md`, `..`가 든 경로, 절대 경로, 링크, 폴더) `not_allowed` 로 거절한다. `CLAUDE.local.md`(등급 B)는 조각 S4 가 연다. 경로를 따라가는 폴더 중 하나라도 심볼릭 링크면 따라가지 않고 거절한다. 허용 경로는 `src/generation.rs` 가 쓴 매니페스트의 기록에서만 읽는다(그 파일은 이 조각이 읽기만 한다).

**항목.** `--item` 은 JSON 한 개이고 반복한다. 한 번에 1~3개(retire 포함)다.

- `{"op":"add|replace","id":"<id>","rule":"<한 문장>","reason":"<한 줄>","source":["<신호 id>", …]}`
- `{"op":"retire","id":"<id>"}` (다른 필드를 쓰면 거절)
- `id` 는 `[a-z0-9][a-z0-9_-]` 로 40자 이하. `rule`·`reason` 은 1~200자, 한 줄이고 `|`·`<!--`·`-->` 를 담지 않는다. `source` 는 1~5개이고 각각 1~120자 한 줄(제안을 낳은 신호의 부류 키나 참조를 호출하는 쪽이 넘기고, 이 명령은 모양만 본다). 모르는 필드, 같은 `id` 두 번, JSON 이 아닌 값은 거절한다.

**병합(결정적 코드).** 대상 파일의 관리 블록 `<!-- moai:managed:begin id=moai-lessons -->` … `<!-- moai:managed:end id=moai-lessons -->` 안에서만 `id` 로 더하고(블록 맨 끝), 바꾸고(제자리), 뺀다. 블록이 없으면 `add` 가 파일 끝에 빈 줄 하나와 함께 블록을 만든다. 항목 한 줄의 모양은 다음과 같다. 블록 밖 바이트는 그대로다.

`- [<id>] <rule> | why: <reason> | source: <s1>, <s2> | status: trial | helped: 0 | hurt: 0`

`add` 는 이미 있는 `id`, `replace`·`retire` 는 없는 `id`, `retire` 는 `status: trial` 이 아닌 항목을 거절한다. 마커가 없거나 한 번씩만 있지 않으면 고치지 않고 거절한다.

**정적 게이트와 거절 사유.** 거절은 오류 종류 `invalid_input`, 메시지 `improve-plan refused: reasons=<사유,사유> · <사유>: <설명> …` 이다. 사유 단어는 모든 실패한 게이트의 것을 알파벳 순으로 모두 담는다(하나가 걸려도 다음 게이트를 계속 본다). 설명은 항목 글을 되풀이하지 않는다.

| 사유 | 걸리는 때 |
|------|-----------|
| `not_allowed` | 위 허용 목록 밖의 대상, 마커가 한 쌍이 아닌 블록 |
| `test_weakening` | 대상 경로가 테스트·평가 세트(경로 조각 `tests`·`test`·`evals`·`__tests__`, 파일 이름 `*_test.*`·`test_*`·`*.test.*`·`*.spec.*`)에 있음. 경로만 본다. 항목 글의 낱말 목록은 쓰지 않는다 |
| `items` | 항목이 0개나 4개 이상, 모양·길이 위반, 같은 `id`, 병합할 수 없는 `op` |
| `drift` | `rule`·`reason` 이 프로젝트에 없는 경로(확장자 rs·md·py·json·yml·yaml·toml·sh·txt 인 상대 경로)를 말하거나, 백틱 안의 `` `moai <낱말>` `` 의 낱말이 이 바이너리의 명령이 아님 |
| `secret` | `rule`·`reason`·`source`·`id` 가 `redact_secret_shapes` 로 바뀌는 비밀 모양을 담음. 결과와 오류에 그 글을 싣지 않는다 |
| `front_matter` | 대상이 `---` 로 시작하는 머리말에 `name` 과 `description`(1,024자 이하)을 갖지 않음, UTF-8 이 아님 |
| `budget` | 병합 뒤 줄 수가 한도를 넘음(스킬 500줄, 에이전트 파일 200줄. 에이전트 한도는 SPEC D10 에 없어 AC-12 가 200 으로 정한다), 대상이 읽기 한도 1 MiB 를 넘음 |

한도는 새 바이트 기준이다. N 줄은 통과하고 N+1 줄은 거절하며, 같은 제안이 시험 항목 하나를 `retire` 해서 N 줄로 돌아오면 통과한다. 이미 한도를 넘은 파일은 그 줄 수로 돌아오지 못하는 제안을 거절한다.

**결과.** `action:"plan"`, `grade:"A"`, `target`, `kind`(`project-skill`·`native-role`), `generation`(`unchanged`: 대상 바이트의 sha-256 이 매니페스트의 `baseline_sha256` 과 같음, `conflict`: 다르거나 기록이 없음), `old_sha256`·`new_sha256`, `old_lines`·`new_lines`·`budget_lines`, `items`(각 `op`·`id`·`line`), `plan_digest`(`sha256:<hex>`), `writes:false`.

**`plan_digest`.** 상대 대상 경로, 옛·새 바이트의 sha-256, 매니페스트의 `baseline_sha256`, 항목들을 도메인 `improve-plan@1` 아래에서 해시한다. 절대 경로·프로젝트 폴더 이름·시각은 입력이 아니다. 같은 입력이면 어느 폴더에서든 같은 값이고, 항목이나 대상 바이트나 기준 해시가 바뀌면 다른 값이다. 적용 조각(S4)이 이 값으로 제안과 적용을 묶는다.

**`generation_plan` 미리보기가 아닌 까닭.** 허용 경로를 정하는 함수는 `src/generation.rs` 안에서만 보이고 계획 함수는 프로필 답에서 완전한 요청을 받아야 한다. 이 조각은 그 파일을 읽기만 하므로 매니페스트 기록과 대상 파일의 실제 바이트에서 미리보기를 낸다.

**주의.** `source` 는 호출하는 쪽이 `improve-signals` 의 결과에서 골라 넘기고, 이 명령은 장부를 다시 읽지 않는다. 쓰지 않는 명령이라 `--permission-reference` 는 받지 않는다.
