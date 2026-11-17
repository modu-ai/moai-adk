# `spec`

operation `spec` · 코드 `src/ops/spec.rs`, `src/spec_doc.rs`, `templates/spec/full.md`, `templates/spec/minimal.md` · SDD 생애 주기 계획(2026-11-17) SDD-1·SDD-2, SPEC-20261117-spec-tooling

`spec <동작>` 은 프로젝트의 SPEC 문서 `<project>/.moai/specs/<SPEC-ID>/spec.md` 를 만들고, 검사하고, 목록으로 보이고, 하나를 JSON 으로 보여 준다. 낱말 하나에 동작이 중첩돼 있고 봉투의 `operation` 은 항상 `spec`, 동작은 결과의 `action` 으로 나간다. **어떤 동작도 프로그램을 실행하거나 네트워크에 나가지 않는다**: `verify:` 명령은 읽기만 하고 돌리지 않는다. 각 동작의 도움말 첫 문장은 `Writes one new SPEC.` 또는 `Reads only.` 이다.

| 동작 | 파일을 | 입력 | 결과 |
| --- | --- | --- | --- |
| `spec new SLUG --title T [--minimal] [--parent ID] [--date YYYY-MM-DD] [--project .]` | `.moai/specs/<ID>/` 폴더와 그 안의 `spec.md` 하나만 씀. `.moai/`·`.moai/specs/` 가 없으면 그것도 만든다. 그 밖에는 아무것도 쓰지 않음 | 슬러그는 소문자·숫자 낱말을 `-` 하나로 이은 1–48자, 제목은 한 줄 1–200**글자**(바이트가 아님; 제어 문자·`"`·`\` 거절), `--parent` 는 SPEC id(자기 자신 거절, 있는지는 보지 않음), `--date` 는 실제 달력 날짜(기본: 오늘 UTC) | `id`(`SPEC-<yyyymmdd>-<slug>`), `path`(프로젝트 기준), `status: "draft"`, `revision: 1`, `form`(`minimal`\|`full`), `created`, `writes: true`. 폴더가 이미 있으면 `conflict` 이고 아무것도 안 씀. 입력 거절은 `spec new refused: …`(`invalid_input`) |
| `spec check (ID \| --all) [--closure] [--project .]` | 읽기만 | `ID` 는 새 id 또는 Go 세대 id 모양이어야 하고(아니면 거절, 입력을 되풀이하지 않음), 폴더가 없으면 `not_found`. `--all` 은 `.moai/specs/` 의 폴더와 링크 전부(일반 파일은 건너뜀) | 오류가 하나도 없을 때만 종료 0: `files[]`(`id`·`path`·`errors`·`warnings`·`findings[]`), `errors: 0`, `warnings`, `closure`, `specs`(`present`\|`absent`). 오류가 있으면 종료 1, `spec check refused: N error(s) and M warning(s) in K SPEC(s): <경로>:<줄> <코드>; …`(`invalid_input`, 처음 50개까지). 폴더가 없으면 빈 목록과 종료 0 |
| `spec list [--project .]` | 읽기만 | — | `rows[]`(폴더 이름 순: `id`·`folder`·`status`·`title`·`parent`·`legacy`·`errors`·`warnings`), `count`, `specs` |
| `spec show ID [--project .]` | 읽기만 | `check` 의 `ID` 와 같음 | `path`, `spec`(앞머리 값, `form`, `sections`, `requirements[]`, `acceptance[]`(`refs`·`verify`·`manual`·`unknown`), `scope`(`write`·`read_only`·`forbidden`), `result[]`, `digest`, `findings[]`). 오류가 있어도 보여 준다 |

## `spec.md` 형식 (`src/spec_doc.rs`)

- **앞머리**: 첫 줄 `---` 와 닫는 `---` 사이의 `key: value` 줄. YAML 이 아니고 이 모양만 읽는다. 키는 `id`, `title`, `status`, `revision`(1 이상 정수, 앞자리 0 거절), `created`(`YYYY-MM-DD`), `parent`(선택, SPEC id), `approval`(들여 쓴 `by`·`at`·`basis`·`digest`). 모르는 키는 경고 `fm_unknown_key` 이고 오류가 아니다. **`cards:` 는 오류 `fm_cards`**: 개인 카드 번호는 공유되는 SPEC 에 들어가지 않는다(연결은 카드 쪽에 둔다).
- **id**: 새 모양 `SPEC-<yyyymmdd>-<slug>`(실제 달력 날짜, 슬러그 48자 이하). Go 세대 `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` 는 읽기만 하고 쓰지 않는다. 폴더 이름은 id 와 같아야 한다(`id_folder`).
- **상태**: `draft`·`approved`·`completed`·`superseded`·`rejected`·`archived`. `approved`·`completed`·`superseded` 는 `approval` 네 칸이 모두 있어야 한다(`approval_missing`, `approval_field`; `at` 은 `YYYY-MM-DD` 로 시작). Go 세대 상태(`planned`·`plan_ready` → `draft`, `ready-for-run`·`in-progress`·`implemented`·`implemented-pending-sync`·`sync-started` → `approved`, `backfilled` → `completed`, `retired` → `archived`)는 새 낱말로 읽고(`status_as`) `status_legacy` 로 알린다: Go id 에서는 경고, 새 id 에서는 오류.
- **절**(`##` 제목, 영어): `Intent`·`Requirements`·`Acceptance`·`Scope`·`Plan` 은 언제나 필요하다(최소 SPEC). `Non-goals` 나 `Risks` 가 하나라도 있으면 완전한 SPEC(`form: full`)이고 둘 다 필요하다. `Decisions`·`Result` 는 선택. 모르는 `##` 는 경고 `section_unknown`, 두 번 나온 절은 `section_duplicate`, 비어 있는 `Intent`·`Plan`·`Non-goals`·`Risks` 는 `section_empty`. 울타리 코드(```` ``` ````·`~~~`) 안의 제목은 글이다. 본문 언어는 작성자 마음이다.
- **요구**: `- REQ-<n>: <GEARS 문장>` 을 `spec::parse_gears` 로 읽는다(영어 키워드 `Where`/`While`/`When … shall …`, 본문은 한국어도 됨). EARS 의 원치 않는 동작 형태 `- REQ-<n> (REQ-…): IF <조건>, THEN <주어> shall <동작>` 도 받는다(`kind: "unwanted"`; `IF`·`THEN` 은 대소문자 무관, 쉼표는 없어도 됨; `When <조건>, <주어> shall <동작>` 으로 바꿔 같은 `parse_gears` 규칙을 거친다). 문장이 아니면 `req_gears`(이유만 말하고 글은 되풀이하지 않음; 비밀 모양·개인 절대 경로도 여기서 거절), 번호가 겹치면 `req_duplicate`, 하나도 없으면 `req_missing`. GEARS 문장은 통과 조건이다(권고가 아님).
- **실패 경로 규칙**(사용자 결정 2026-11-17): `When` 트리거가 있는 GEARS 요구마다, 같은 SPEC 안에 그 REQ 를 괄호로 가리키는(제대로 읽힌) IF/THEN 요구가 있거나, 그 REQ 줄에 `failure: none` 메모가 있어야 한다(`· failure: none — 이유` 처럼; 메모는 본문에서 빠진다). 없으면 **오류** `req_failure_path`: 실패 경로가 없는 트리거는 아무도 쓰지 않은 요구라 어떤 AC 도 그것을 확인할 수 없고, 면제는 메모 한 줄이면 되기 때문이다(Go 세대 SPEC 에서는 다른 규칙처럼 경고). IF/THEN 요구가 없는 REQ 를 가리키면 `req_ref_unknown`. `Where`·`While` 만 있는 요구에는 이 규칙이 없다.
- **수용 기준**: `- AC-<n> (REQ-…): <관찰 결과> · verify: `<명령>`` 또는 `manual: <누가 무엇을>` 또는 `unknown: <이유>`. 꼬리표는 같은 줄(앞이 공백이나 `·`)이나 바로 아래 줄 첫머리에 둔다. 없으면 `ac_no_verify`, `unknown:` 은 경고 `ac_unknown`, 번호가 겹치면 `ac_duplicate`, 없는 REQ 를 가리키면 `ac_ref_unknown`.
- **범위**: `- write: `a`, `b``·`- read-only: …`·`- forbidden: …`. 경로가 없는 꼬리표 줄(`write:`)은 그 아래 목록의 종류를 정하고, 꼬리표 없는 목록 항목은 `write` 다. 절대 경로(`/`·`\`·`~`·`C:`)는 `scope_absolute`, `..` 구간은 `scope_parent`, `write` 경로가 없으면 `scope_empty`.
- **digest**: 요구·수용 기준·범위에 대한 `spec::content_digest`(`show` 의 `digest`). 승인할 때 `approval.digest` 에 적고 PR 본문에도 되풀이한다. 승인 뒤 요구·기준·범위가 바뀌면 `approval_digest` 오류 — `revision` 을 올리고 다시 승인한다. 다른 글(의도·계획 등)은 digest 에 들지 않는다. 파일이 자기 git blob id 를 담을 수는 없어서 blob id 가 아니라 내용 digest 다.
- **마감(`--closure`)**: 상태가 `completed`(`closure_status`), `## Result` 가 있음(`closure_result`), 모든 AC 가 `- AC-<n>: met …` 또는 `withdrawn …`(`closure_open`), Result 가 없는 AC 를 가리키지 않음(`closure_unknown`).
- **Go 세대 SPEC**(Go id): 읽기·앞머리 블록·id·상태 코드만 오류로 남고 나머지는 경고로 내려간다(마감 검사는 예외 없이 오류). 1,000개가 넘는 옛 SPEC 이 `check --all` 을 막지 않게 하려는 것이다.
- **읽기 한도**: `spec.md` 는 일반 파일이어야 하고(링크는 따라가지 않음: `O_NOFOLLOW`, 연 파일과 `lstat` 의 장치·inode 비교), 256 KiB(262,144 바이트)까지, UTF-8. 넘으면 `read_too_large`, 링크·폴더는 `read_not_file`, 없으면 `read_missing`, UTF-8 이 아니면 `read_utf8`. SPEC 폴더나 `.moai`·`.moai/specs` 가 링크여도 들어가지 않는다.

- 템플릿은 바이너리에 들어 있다(`include_str!`, `templates/spec/`). 키와 제목은 영어, 자리 표시(`<…>`)는 작성자가 채운다. 새 템플릿은 `check` 를 오류 없이 통과한다.
- `spec new` 는 `.moai/` 가 없는 프로젝트에도 `.moai/specs/<ID>/` 를 만든다. 그러면 그 체크아웃의 `state_root::classify` 는 표식 없는 `.moai/` 라서 `unknown` 이 된다(SDD-3 이 `specs/` 를 추적할 때 함께 정할 일). `.moai` 가 링크이거나 `interrupted`·`refuse` 면 거절한다.
- 아직 없는 동작: `take`(SDD-7), `close`(SDD-8), 승인 기록 쓰기, 카드 연결(SDD-4).
