---
id: SPEC-20261122-report-rules
title: "A report is a file first: the md and html files are the record, the Claude artifact is a private extra when the session can publish, and long tables and repair items follow the documented limits"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# A report is a file first: the md and html files are the record, the Claude artifact is a private extra when the session can publish, and long tables and repair items follow the documented limits

## Intent

사용자 결정(2026-11-22, 보고서 문구 ⑤): "html + md 가 기록이고, 아티팩트는 될 때(비공개)". GLM, API 키, `-p`, ZDR 에서는 아티팩트가 안 돼요. 게시가 실패해도 보고서는 실패하지 않아요. 부모가 `SPEC-20261121-flow-skeleton` 인 까닭: 흐름의 레인·리더·사용자가 서로 읽는 증거가 보고서 파일이고(AGENTS.md "Reports go to `~/.moai/reports/<project>/`"), 이 SPEC 은 그 파일의 모양과 공개 규칙만 정해요.

moai 는 보고서를 어떻게 쓸지에 대한 명시적 규칙이 아직 없어요. 지금은 사용자의 기억 한 줄(`html + md + Claude 아티팩트 셋을 준다`)이 규칙을 대신해요. 이 SPEC 은 공식 문서의 한도를 규칙으로 옮겨요: 아티팩트가 안 되는 조건(문서 `artifacts` Availability 표), 터미널이 표를 200줄에서 자름(`troubleshooting`), 딥 링크의 길이와 모양(`deep-links`), Edit·Write 훅이 있는 세션의 게시 방식(`errors`), 보고서 폴더의 짝 파일과 리비전 표시(`claude-security` 의 보고서 폴더가 보여 주는 모양).

규칙은 두 곳에 놓여요. 에이전트가 읽는 글(스킬의 내부 참고 글 하나)과, 보고서 한 벌이 그 규칙을 지켰는지 읽기만 하는 검사 연산 하나(`report-check`)예요. moai 는 아티팩트를 직접 게시하지 않아요: 게시는 세션의 `Artifact` 도구가 하고, 이 SPEC 은 언제 어떻게 부르는지만 정해요. 프로젝트의 `.claude/` 에는 아무것도 쓰지 않아요(사용자 결정 2026-11-22).

## Non-goals

- 아티팩트를 공유하거나 공개로 바꾸지 않아요. 게시된 것은 비공개이고 공유는 사용자가 정해요.
- html 로 바꾸는 프로그램(pandoc 등)을 moai 가 설치하거나 부르지 않아요. 바꾸는 방법은 글에 있고 쓰는 것은 에이전트예요.
- Orca 아티팩트와 Orca 봇은 쓰지 않아요(사용자 결정 2026-11-22).
- 이슈·PR 댓글 보고(`forge-report`)는 이 SPEC 의 대상이 아니에요.
- 게시 가능 여부를 미리 알아내려고 환경을 뒤지거나 인증 정보를 읽지 않아요.

## Requirements

- REQ-1: The report guidance shall name the md file in the user's reports folder as the record and the html file beside it as its rendering · failure: none — the Claude artifact is a third copy and never the record.
- REQ-2: When the session offers the Artifact tool, the report step shall publish the html page once after both files are written and give the private link · failure: none — a refusal is handled by REQ-3.
- REQ-3 (REQ-2): IF the publish is refused or the Artifact tool is absent, THEN the report step shall give the two file paths and one line that says no artifact was made and why.
- REQ-4: The report step shall write both files before it asks for a publish · failure: none — a failed publish then costs nothing and a refusal never fails the report, and no probe of the account or the environment comes first.
- REQ-5: When a report holds a table of more than 200 rows, the report step shall write the table to a file beside the report and show its row count and path in the terminal · failure: none — the terminal shows only the first 200 rows.
- REQ-6: When the session has a PreToolUse hook on Edit or Write, the report step shall change the report file with the Edit or Write tool and then publish the same path as its own call · failure: none — the host refuses a one-step update there.
- REQ-7: When a report lists findings, the report step shall write a `.jsonl` file beside it with one finding per line and the same ids as the md · failure: none — a report without findings has no `.jsonl`.
- REQ-8: The report shall carry one revision line that holds the head commit it was made at or `UNVERSIONED` · failure: none — a repair item whose files changed since that commit is shown as stale, not as a repair.
- REQ-9: When a repair item names a folder, the report step shall add a `claude-cli://open` link whose `q` is URL-encoded and at most 5,000 characters and whose `cwd` is an absolute path without `..` · failure: none — the md holds the link in a code block and the html holds it as a link.
- REQ-10 (REQ-9): IF a repair item's prompt passes 5,000 characters or holds a value shaped like a credential, THEN the report step shall write that item without a link.
- REQ-11: When a report folder lies inside a git work tree, the report step shall write a `.gitignore` in it that ignores every file · failure: none — an existing `.gitignore` is kept as it is.
- REQ-12: The `report-check` operation shall list each broken rule of REQ-5 and REQ-7 to REQ-11 for one report file with the line it is on · failure: none — it reads files only and writes nothing.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4, REQ-6): 에이전트 참고 글이 기록(md·html)과 아티팩트의 순서, 거절 한 줄, 게시 전에 두 파일을 쓰는 것, Edit·Write 훅 세션의 두 단계 게시를 이름으로 말하고, 원본 목록의 sha256 이 글과 맞아요 · verify: `cargo test --locked --offline --test report_rules_skill the_reference_names_the_record_the_refusal_line_and_the_two_step_publish`
- AC-2 (REQ-5, REQ-7, REQ-8, REQ-11, REQ-12): 경계값: 표 200줄은 통과하고 201줄은 지적되고, 검사 결과가 줄 번호를 갖고, 리비전 줄이 없으면 지적하고, `.jsonl` 짝의 아이디가 md 와 어긋나면 지적하고, 작업 트리 안 폴더에 `.gitignore` 가 없으면 지적하고, 아무것도 쓰지 않아요 · verify: `cargo test --locked --offline --test report_check_lib the_check_lists_each_broken_rule_with_its_line_and_writes_nothing`
- AC-3 (REQ-9, REQ-10): 딥 링크 경계값: `q` 5,000 자는 통과하고 5,001 자는 링크 없는 항목이 되고, `cwd` 가 상대 경로나 `..` 면 지적하고, 비밀 모양 값이 있으면 링크 없음 · verify: `cargo test --locked --offline --test report_check_lib a_repair_link_follows_the_documented_limits`
- AC-4 (REQ-12): 연산이 `moai` 명령줄 도움말과 운영 페이지에 올라 있고, 보고서가 없으면 이유와 함께 거절해요 · verify: `cargo test --locked --offline --test report_check_cli the_operation_is_listed_and_refuses_a_missing_report`

## Scope

- write: `.moai/specs/SPEC-20261122-report-rules/**`(이 조각). 뒤 조각에서 카드 범위에 더해요: `skills/moai/references/internal/reports.md`(새 글), `skills/moai/references/source-manifest.json`, `skills/moai/references/internal/core-rules.md`(한 줄 가리킴), `src/report_check.rs`, `src/ops/report_check.rs`, `src/lib.rs`(모듈 한 줄), 운영 페이지 `docs/design-review/appendix/operations/report_check.md`, 시험 파일 셋
- read-only: 사용자의 보고서 폴더, `AGENTS.md`, `SPEC-20261121-flow-skeleton`
- forbidden: 아티팩트 공유·삭제, 프로젝트 `.claude/` 쓰기, 인증 정보 읽기, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰기, 실제 HOME

## Plan

1. 조각 1(이번): 이 SPEC 과 아래 Evidence(문서 인용).
2. 조각 A: 참고 글 `references/internal/reports.md` 와 시험(REQ-1 에서 REQ-4, REQ-6, AC-1). 원본 목록과 core-rules 의 한 줄 가리킴.
3. 조각 B: `report-check` 의 읽기(REQ-5, REQ-7 에서 REQ-11, REQ-12, AC-2, AC-3). 줄 번호와 규칙 이름을 가진 결과. 표 줄 수, 딥 링크 길이, 비밀 모양, `.jsonl` 아이디 비교, 리비전 줄, `.gitignore`.
4. 조각 C: 명령줄 연결, 운영 페이지, 도움말(AC-4).

## Risks

- 문서가 말하지 않는 것: 게시 실패가 세션에 영향이 없는지(오류 글 "Nothing was published" 와 "Usually nothing: Claude rewrites the file and publishes again" 만 있어요), 같은 파일 경로가 같은 아티팩트를 고르는지(문서는 "같은 URL" 이라고만 해요), 클라우드 세션에서 게시가 되는지(문서가 막지도 허락하지도 않아요). 규칙은 이 셋을 가정하지 않고, 게시를 한 번만 시도하고 거절은 줄로 남겨요.
- 문서가 어떤 훅도 "편집 뒤 게시" 로 권하지 않아요. 오히려 PreToolUse 훅이 있으면 한 번에 고치며 게시하는 길을 거절해요(`CLAUDE_CODE_ARTIFACT_PUBLISH_EDITS=0` 이 끄는 방법). 그래서 REQ-6 은 두 단계 게시이고 `PostToolUse` 훅은 쓰지 않아요.
- 딥 링크의 `q` 는 주소에 들어가서 브라우저 기록과 로그에 남을 수 있어요. 그래서 비밀 모양 값이 있으면 링크를 만들지 않아요(REQ-10). 문서는 `claude-cli://` 가 GitHub 의 마크다운에서 링크로 보이지 않는다고 해서(REQ-9 의 코드 블록), 보고서가 어디서 읽히든 주소가 보이도록 해요.
- `repo` 매개변수는 그 폴더를 한 번 `claude` 로 연 적이 있어야 풀려요. 그래서 `cwd` 를 써요.
- 문서의 `errors` 페이지는 643 KB 이고 조사 에이전트가 관련 절만 읽었어요(Evidence 의 읽은 범위 참고). 나머지에 이 규칙과 어긋나는 문장이 있을 가능성은 낮지만 검사하지 않았어요.
- 문서 줄은 버전마다 달라질 수 있어서 인용 표에 읽은 날짜가 있어요. 한도(200줄, 5,000자, 16 MiB)는 글에 숫자로 적되 검사 코드의 상수 한 곳에 모아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기록은 md 와 html 이에요. 아티팩트는 세션이 게시할 수 있을 때의 비공개 세 번째 사본이에요. | 사용자 결정 2026-11-22(문구 ⑤)와 "html, md, claude 아티팩트 셋만 제공" 결정. | 카드 열림 |
| D-2 | 게시 가능 여부를 미리 알아내지 않아요. 한 번 불러 보고, 거절이나 도구 없음이면 한 줄로 이유를 남겨요. | 문서에 가능 여부를 묻는 길이 없고, 거절 글이 이유를 말해요(`errors`, `artifacts`). | 카드 열림 |
| D-3 | 터미널 표 한도는 200줄이에요. 넘으면 파일로 보내고 줄 수와 경로를 터미널에 보여 줘요. | 문서 `troubleshooting`: 200줄만 그리고 `… N more rows not shown`. | 카드 열림 |
| D-4 | 딥 링크는 `cwd` 와 `q` 만 써요. `q` 5,000자 이하, URL 인코딩, 줄바꿈 `%0A`. | 문서 `deep-links`. 세션이나 파일 매개변수는 문서에 없어요. | 카드 열림 |
| D-5 | 보고서 폴더의 짝(`.md`, `.jsonl`, 리비전 줄, `.gitignore`)은 공식 `claude-security` 보고서 폴더의 모양을 따라요. | 문서 `claude-security`, "Read the scan results". | 카드 열림 |
| D-6 | `report-check` 는 읽기만 하고 고치지 않아요. 고치는 것은 보고서를 쓴 에이전트예요. | 사용자 규칙: 검사는 읽기, 쓰기는 따로. | 카드 열림 |

### Evidence

문서는 2026-11-22 에 조사 에이전트가 읽었어요(UTC 06:39 에서 06:40, `.md` 꼴, 원문 인용). 조각 A 의 참고 글이 이 표를 줄 단위 출처로 써요.

| 규칙 | 인용 | 출처 |
| --- | --- | --- |
| 아티팩트가 안 되는 세션 | "Sessions using an API key, gateway token, or cloud-provider credential cannot publish." | `artifacts.md`, Availability, Authentication |
| 모델 제공자 | "Anthropic API. Not available on Amazon Bedrock, Google Cloud's Agent Platform, or Microsoft Foundry." | `artifacts.md`, Availability, Model provider |
| 조직 정책 | "Customer-managed encryption keys (CMEK), HIPAA, and Zero Data Retention are not enabled for the organization." | `artifacts.md`, Availability, Organization policy |
| 비대화형 | "Off by default in the Agent SDK, the GitHub Action, and MCP-server contexts, when you run Claude Code with -p from your own terminal or scripts" | `artifacts.md`, Availability, Surface |
| 끄는 설정 | "Set CLAUDE_CODE_DISABLE_ARTIFACT=1" / `"enableArtifact": false` / "Add Artifact to permissions.deny" | `artifacts.md`, Disable artifacts |
| 콘솔 로그인 | "the 'Anthropic Console account' option does not provide claude.ai credentials." | `errors.md`, Artifacts need a claude.ai login |
| 비공개와 같은 주소 | "publishes from your session to a private URL on claude.ai." / "Claude edits the underlying file and publishes again to the same URL." | `artifacts.md`, 첫 문단, Update an artifact |
| 크기와 파일 | "The rendered page must be 16 MiB or smaller." / "The published file must be .html, .htm, or .md, and must decode as UTF-8" | `artifacts.md`, Page constraints |
| 게시 실패 | "so Claude Code refused the publish before uploading anything" / "Nothing was published." | `errors.md`, The source file is not valid UTF-8 text |
| 훅이 있는 세션 | "Claude Code doesn't allow one-step updates while your session has a PreToolUse hook or mod on Edit or Write." / "Edit the file, then publish it without edits." | `errors.md`, Refusing to apply edits: a hook is configured for file edits in this session |
| 한 번에 고침 끄기 | `CLAUDE_CODE_ARTIFACT_PUBLISH_EDITS` set to `0` | `artifacts.md`, Turn off one-step artifact updates |
| 터미널 표 | "A Markdown table with more than 200 rows renders its first 200 rows followed by a `… N more rows not shown` line." / "For a table too large to read in the terminal, ask Claude to write it to a file instead." | `troubleshooting.md`, Large tables are cut off in the terminal |
| 딥 링크 모양 | "A deep link starts with `claude-cli://open`, followed by optional query parameters." | `deep-links.md` |
| `q` | "Text to pre-fill in the prompt box. URL-encode the value. Use `%0A` for line breaks in multi-line prompts. Maximum 5,000 characters." | `deep-links.md`, parameters |
| `cwd` | "Absolute path to use as the working directory. Network and UNC paths are rejected, and so are paths that contain `..` segments or invisible or bidirectional control characters." | `deep-links.md`, parameters |
| `repo` | "Run `claude` inside the clone once so Claude Code records its path" | `deep-links.md`, parameters |
| 보내지 않음 | "The prompt is populated but not sent until you press Enter." | `deep-links.md` |
| GitHub 마크다운 | "GitHub-rendered Markdown does not allow `claude-cli://`, so a deep link in a GitHub README, issue, or wiki shows only its label with no clickable link." / "put the deep link in a code block" | `deep-links.md` |
| 보고서 폴더의 짝 | "`CLAUDE-SECURITY-RESULTS.md`: the report" / "`CLAUDE-SECURITY-RESULTS.jsonl`: the same findings in machine-readable form, one JSON object per line" / "`CLAUDE-SECURITY-REVISION-<commit>.json`: the revision stamp" / "A scan outside version control stamps `UNVERSIONED` in place of the commit" | `claude-security.md`, Read the scan results |
| 보고서 폴더의 `.gitignore` | "it carries its own `.gitignore`, so a stray `git add` never sweeps a report into a commit." | `claude-security.md`, Read the scan results |
| 오래된 보고서 | "findings whose code has since changed are skipped with a note, and the plugin offers a fresh scan instead of patching from a stale report." / "Patches are never applied automatically." | `claude-security.md`, Fix findings |

읽지 않은 것(조사 에이전트 보고): `errors.md` 는 643 KB 라서 관련 절만 읽었고, `claude-security.md` 와 `deep-links.md` 와 `troubleshooting.md` 는 해당 절을 읽었어요. 클라우드 세션의 게시 가능 여부, 일반 최소 버전, 표 너비 한도, 딥 링크의 세션·파일 매개변수와 버전 조건은 문서에 없어요.

## Result

(status draft — written when the slice lands)
