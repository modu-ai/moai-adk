---
id: SPEC-20261127-review-term-rename
title: "영어 낱말 audit 을 review 로 바꾼다 — 이름 목록, 옛 기록 읽기, 조각 순서"
status: draft
revision: 1
created: 2026-11-27
---

# 영어 낱말 audit 을 review 로 바꾼다 — 이름 목록, 옛 기록 읽기, 조각 순서

## Intent

한국 개발자는 이 행위를 리뷰라고 말해요. 사용자가 2026-11-27 에 "영어도 review로" 를 골랐고(한국어 글의 낱말은 이미 정해졌어요), 영어 쪽 이름은 아직 audit 이에요: 노트 첫 줄 `AUDIT`, 기본 에이전트 `moai-audit`, 명령 12개, 게이트 `audit-gate`, GitHub 상태 `moai/audit`. 카드 상태 `review` 와 `moai card review`, `review-package`, `review-check`, `REVIEW.md` 는 이미 같은 행위의 이름이라서, 이 이름 바꾸기는 한 행위에 낱말 하나를 남겨요. 이 SPEC 은 무엇이 바뀌는지의 목록(세어 둔 수), 옛 기록을 읽는 법, 옛 이름을 받을지, 조각의 순서, `moai update` 의 처리를 정해요. 게이트의 강도는 바꾸지 않아요.

### Facts

측정은 NeXT tip `aadb78141a` 에서 `git grep` 으로 했어요. 대소문자 무시 `audit` 이 들어간 추적 파일은 1127개, 경로 이름에 audit 이 든 파일은 79개예요. 아래 줄 수는 영역별 `audit` 이 든 줄이에요.

| 영역 | 파일 | 줄 | 이 SPEC 에서 |
| --- | --- | --- | --- |
| `tests/` | 349 | 3295 | 바꿔요(픽스처 옛 노트는 그대로) |
| `.moai/specs/` | 342 | 2232 | 지난 SPEC 은 안 바꿔요(역사) |
| `src/` | 208 | 1976 | 바꿔요 |
| `.moai-cli/` | 11 | 908 | 안 바꿔요(옛 목록) |
| `docs/` | 90 | 818 | 바꿔요 |
| `templates/` | 53 | 218 | `github/` 8개와 `claude/output-styles/` 3개("in check (audit)" 줄)만 바꿔요(`bundle-packs/` 42개 제외) |
| `skills/` | 36 | 180 | 바꿔요 |
| `tools/` | 10 | 59 | 바꿔요 |
| `.github/` | 10 | 40 | 바꿔요 |
| `.moai/evals/` | 5 | 34 | 바꿔요 |
| 루트 파일 | 7 | 32 | `CLAUDE.md` 10줄은 제외, 나머지는 하나씩 봐요 |
| `agents/` | 5 | 30 | 바꿔요 |
| `schemas/` | 1 | 3 | 바꿔요 |

낱말 모양은 `audit` 6772, `auditor` 612, `audited` 338, `audit_status` 209, `audits` 199, `audit-gate` 199, `audit-tier` 164, `audit_gate` 153, `audit-claim` 125 번이 많은 순서예요(`git grep -ow`).

| 이름의 종류 | 수 | 예 |
| --- | --- | --- |
| 명령 | 12 | `audit-claim`, `audit-claim-stale`, `audit-verdict-line`, `audit-gate`, `release-audit-summary`, `audit-clone-plan`, `audit-slot`, `audit-prep`, `audit-done`, `audit-delta`, `audit-reduce`, `forge-audit-status` (`src/help.rs`) |
| 그 밖의 명령 | 1 | `pr-self-audit-check` |
| 옵션 | 2 | `--self-audit`(5곳), `card assign --auditor`(1곳) |
| `operation` 값(JSON 출력) | 12 | `audit.claim`, `audit.gate`, `audit.slot`, `forge.audit_status`, `release.audit.summary` 등 명령마다 하나 |
| 소스의 따옴표 문자열 | 36 종 | `audit_floor`, `audit_note`, `audit_verdict`, `audit_reserved`, `auditor`, `audited_head`, `reserve_audit` 등 |
| 레저 저장 이름 | 2 | 카드 종류 값 `audit`(`src/ledger/schema.rs:26` 의 CHECK), 이벤트 종류 `card.audit_assigned` |
| 프로토콜 | 3 | 노트 첫 줄 `AUDIT`(대문자 480줄, 첫 줄을 읽는 곳 `src/bundle.rs:808`, `src/card/orchestrate.rs:298` 등 31개 소스 파일), PR 본문 `audit-tier:`(164번), 예약 줄 `moai-audit-claim:`·`moai-audit-withdraw:`·`moai-audit-takeover-request:` |
| GitHub | 3 | 상태 컨텍스트 `moai/audit`(소스 13줄, 테스트 26줄, 문서 45줄, `.github` 9줄), 워크플로·검사 이름 `audit-gate`, 룰셋이 요구하는 컨텍스트 |
| 스킬·에이전트 | 4 | 기본 에이전트 `moai-audit`(소스 45줄, 테스트 84줄, 스킬 28줄, 문서 22줄, 에이전트 8줄), 경로 `agents/audit.md`, `skills/moai/references/audit.md`, `skills/moai/workflows/audit.md`, 라우트 낱말 `AUDIT`(`routes.json` 2줄, `SKILL.md` 3줄) |
| 파일·폴더 이름 | 59 | `src` 21, `tests` 18, `docs` 10(`operations/` 쪽), `.moai/evals/audit/` 4, `agents` 1, `.github/workflows` 1, `skills` 2, `templates` 2(bundle-packs, 제외) — 지난 SPEC 폴더 18개와 `tests/fixtures/audit` 2개는 따로 셌어요 |
| 이슈 템플릿 | 8 | `templates/github/en` 4개와 `ko` 4개의 "독립 AUDIT verdict 수신" 줄 |
| 임시 폴더 이름 | 4 | `moai-audit-<pr>-<head8>`, `moai-audit-slots`, `moai-audit-delta-`, `moai-audit-receipts-` |

이름이 이미 `review` 인 것(충돌 검토 결과): `src/review.rs`(→ `src/audit.rs` 는 같은 이름으로 못 옮겨요), `src/ops/review_package.rs`, `src/ops/review_check.rs`, `REVIEW.md`, `tools/review_checklist.json`, `skills/moai/references/internal/review.md`, 카드 상태 `review`, `moai card review`. 노트나 프로토콜 줄이 `REVIEW ` 로 시작하는 곳은 지금 없어요(`git grep`).

## Non-goals

- 게이트 규칙(tier, 보호 경로의 기준, 독립성 규칙, 판정의 네 가지 FAIL 사유)의 변경.
- `CLAUDE.md`: 그 파일의 audit 낱말(10줄)은 사용자가 레인 창에서 직접 입력해야 열리는 별도 카드가 맡아요.
- 이미 쓰인 기록 다시 쓰기: 레저 행, 노트, 영수증, 판정 이벤트, `.moai/specs/` 의 지난 SPEC 폴더 18개 이름과 342개 파일의 글, `.moai-cli/` 의 옛 목록 11개 파일은 그대로 둬요.
- MoAI 의 행위가 아닌 일반 영어 audit: `templates/bundle-packs/` 41개 파일 201줄과 `packs.json` 5줄(제3자 팩 글과 목록), `design-audit.md`, `moai-ref-cross-model-audit` 은 건드리지 않아요.
- GitHub 룰셋(`26791615`) 편집: 레인은 룰셋을 고치지 않아요. 사용자의 단계예요.
- 옛 기록을 읽는 길의 제거: 읽기 호환은 끝나지 않아요.

## Requirements

<!--
Rule 1: exactly one `shall` per requirement line; a second behavior goes on its own REQ line (spec check refuses it as req_gears).
Rule 2: every When requirement needs an IF/THEN requirement that names it in parentheses, or its line ends with "· failure: none — reason" (spec check refuses it as req_failure_path).
-->
- REQ-1: When the rename lands, the shipped English guidance (skill text, base agent, documents, help text, GitHub templates) shall use `review` for the act of an independent check · failure: none — wording only, and no parser or gate reads these texts.
- REQ-2: When a person or a session types a command or option of the act, the command line shall offer it under its review name (the twelve commands, `pr-self-review-check`, `--self-review`, `card assign --reviewer`).
- REQ-3 (REQ-2): IF an earlier command or option name is typed, THEN the command line shall act as D-3 records.
- REQ-4: When a writer records a verdict note, a status context or an assignment, it shall write the review name only after the readers of REQ-5 and REQ-6 have landed · failure: none — the order is held by the slice order in Plan and by AC-3 reading both forms.
- REQ-5: When any reader parses a note, comment, ledger row or pull request body, it shall count a record written under the earlier name exactly as it counted it before the rename (`AUDIT` first line, `audit-tier:`, `moai-audit-claim`, `card.audit_assigned`, kind `audit`, `moai/audit`).
- REQ-6 (REQ-5): IF a reader finds a record that matches neither the earlier nor the review name, THEN it shall refuse closed as it does today, so a gate never passes on a record it did not read.
- REQ-7: When a head carries a verdict note under each name from one session, the gate shall judge by the newest of them · failure: none — newest-wins is the rule the gate already applies to repeated verdicts.
- REQ-8: When the renamed gate judges a head, it shall apply the same tiers, the same protected-path list (carrying the earlier and the new spelling of every renamed path), the same independence rule and the same strength as before · failure: none — the gate code is moved, not rewritten, and its existing case table runs once per name.
- REQ-9: When the gate posts its GitHub status, it shall post the context or contexts that D-4 records.
- REQ-10 (REQ-9): IF the ruleset still requires the earlier context, THEN the gate shall keep posting it beside the new one.
- REQ-11: When `moai update` finds an installed managed file that holds an earlier name (the `moai-audit` agent file, a skill file, a managed template), it shall move it to the new name, keep the earlier bytes in the receipt and delete nothing.
- REQ-12 (REQ-11): IF the installed file was edited by the user or is not MoAI's, THEN `moai update` shall leave it where it is and report a conflict.

## Acceptance

- AC-1 (REQ-1): 제품이 싣는 영어 글(스킬, 에이전트, 문서, 도움말, GitHub 템플릿)에 허용 목록 밖의 `audit` 이 없고, 허용 목록은 이유와 함께 한 곳에 있어요 · verify: `cargo test --locked --offline --test review_rename_scan`
- AC-2 (REQ-2, REQ-3): 도움말이 review 이름을 보이고, 옛 이름은 D-3 대로 동작해요(기본값이면 같은 결과와 stderr 한 줄) · verify: `cargo test --locked --offline --test review_rename_commands_cli`
- AC-3 (REQ-4, REQ-5, REQ-6, REQ-7): 옛 노트·새 노트·두 이름이 섞인 머리·읽을 수 없는 줄 픽스처에서 읽기가 같은 결과를 내고, 읽을 수 없는 줄은 닫힌 채 거절해요 · verify: `cargo test --locked --offline --test review_rename_reading`
- AC-4 (REQ-8): 게이트의 기존 판정 표가 옛 첫 줄과 새 첫 줄로 각각 한 번씩 같은 결과를 내요 · verify: `cargo test --locked --offline --test review_gate_cli`
- AC-5 (REQ-8): 보호 경로 목록이 바뀐 모든 경로의 옛 이름과 새 이름을 함께 가져서, 이름 바꾸기가 tier 를 낮추지 못해요 · verify: `cargo test --locked --offline --lib gate`
- AC-6 (REQ-9, REQ-10): 게이트가 D-4 대로 컨텍스트를 올리고, 룰셋이 옛 컨텍스트를 요구하는 동안 옛 것도 올려요 · verify: `cargo test --locked --offline --lib forge`
- AC-7 (REQ-11, REQ-12): 임시 프로젝트에서 설치된 옛 에이전트 파일이 옮겨지고 영수증이 옛 바이트를 가져요, 사용자가 고친 파일은 conflict 로 남아요 · verify: `cargo test --locked --offline --test project_update_review_rename_cli`

## Scope

- write: `src/**`, `tests/**`, `skills/moai/**`, `agents/**`, `docs/**`, `tools/**`, `schemas/worktree-report.schema.json`, `.moai/evals/audit/**`, `templates/github/**`, `templates/claude/output-styles/**`, `.github/workflows/audit-gate.yml`, `.github/pull_request_template.md`, `README.md`, `README.ko.md`, `CONTRIBUTING.md`
- read-only: `templates/bundle-packs/**`, `.moai-cli/**`, `.moai/specs/` 의 지난 SPEC
- forbidden: `CLAUDE.md`, `AGENTS.md`, GitHub 룰셋 편집, 레저·영수증·기록된 판정의 다시 쓰기

이 카드의 쓰기 범위는 이 SPEC 폴더뿐이에요. 위 목록은 승인 뒤 조각 카드가 각자 좁혀서 잡아요. `skills/moai/**` 를 건드리는 조각은 사용자 복사본 동기(`moai update` 와 `diff -rq`)를 같은 라운드에 해요.

## Plan

한 번에 랜딩하면 열린 카드의 브랜치와 모두 충돌해요: 측정 시점에 활성 카드 9개, 리뷰 대기 카드 42개, NeXT 에 안 머지된 로컬 브랜치 540개(대부분 지난 것)예요. 아래 순서는 읽는 쪽이 먼저고 쓰는 쪽이 나중인 불변식을 지켜요.

1. 슬라이스 1 — 읽기 호환(더하기만): 31개 소스 파일의 `AUDIT` 읽기, `audit-tier:`, 예약 줄 세 가지, `card.audit_assigned`, `moai/audit` 읽기에 새 이름을 더해요. 쓰는 곳은 안 바꿔요. tier 2(게이트 파서).
2. 슬라이스 2 — 안내 글과 설치 파일: 스킬·에이전트·문서·이슈 템플릿 8개·PR 템플릿·출력 스타일 3개의 영어 글, `agents/audit.md` → `agents/review.md`(이름 `moai-review`), 두 스킬 파일 이름, `moai update` 의 옮김(REQ-11, REQ-12), 보호 경로 목록에 새 경로를 더하기(옛 경로는 남겨요). tier 2(`agents/audit.md`, `skills/moai/workflows/audit.md` 가 floor 목록에 있어요).
3. 슬라이스 3 — 명령과 옵션: 12개 명령, `pr-self-audit-check`, 두 옵션, `operation` 값을 새 이름으로 내고 옛 이름은 D-3 대로 받아요. 도움말과 `docs/design-review/appendix/operations/` 10쪽을 함께 바꿔요.
4. 슬라이스 4 — 기계적 이름 바꾸기와 쓰는 쪽 전환: 파일·모듈 이름(`src` 21개, `tests` 18개; `src/audit.rs` 의 새 이름은 `src/review.rs` 와 안 부딪히게 이 슬라이스가 정해요), 식별자, 임시 폴더 이름, 노트 첫 줄 `REVIEW`, 상태 컨텍스트(D-4), 워크플로 이름. 이름 바꾸기 도구는 허용 목록을 쓰고, 끝난 뒤 옛 낱말 수가 허용 목록 몫만 남는지 세요. D-2 대로 다른 카드가 랜딩된 뒤, 릴리스 전에 해요.
5. 슬라이스 5 — 옛 이름 받기의 끝: D-3 의 기간이 지나면 옛 명령·옵션 이름만 닫아요(옛 기록 읽기는 닫지 않아요).

## Risks

- 읽기가 쓰기보다 늦으면 새 노트를 못 읽는 게이트가 닫힌 채 거절해요(착륙이 막혀요). 거꾸로 쓰기가 늦으면 아무 일도 없어요. 그래서 슬라이스 1 이 먼저예요(AC-3).
- 보호 경로 목록(`src/gate.rs` 의 `src/audit_gate.rs`, `src/forge/audit_status`, `agents/audit.md`, `skills/moai/workflows/audit.md`)과 `tools/protected_checks.json`, `tools/check_guard.py` 가 옛 경로만 가지면 이름을 바꾼 파일이 tier 2 를 벗어나요. 이전 SPEC(audit-gate rename)이 같은 구멍을 닫았어요. 새 경로를 더하고 옛 경로는 남겨요(AC-5).
- `review` 가 이미 쓰이는 곳(`src/review.rs`, 카드 상태 `review`, `review-package`)과 이름이 겹쳐요. 모듈 하나는 이름을 새로 정해야 하고, 카드 종류 `review` 와 카드 상태 `review` 는 헷갈릴 수 있어요(D-5 에서 종류 이름은 그대로 둬요).
- 기계적 바꾸기는 `audit B1`, `AUDIT #3117 E2` 같은 지난 기록을 인용한 주석(소스에 수십 곳)도 건드려요. 지난 기록의 인용은 그대로 둬야 해서 허용 목록이 필요해요.
- 옛 바이너리로 도는 레인이 `AUDIT` 첫 줄을 계속 써요. 읽기가 두 이름을 영원히 받으니 문제가 없어요.
- 이벤트 종류 열에 CHECK 가 있는지 이 조사에서 확인하지 못했어요(카드 종류 열은 `schema.rs:26` 에 있음). 슬라이스 1 이 먼저 읽어서 기록해요.
- `moai update` 의 옮김은 `src/home_style.rs`(MoAI-Easy → MoAI-Style 의 옮김)와 `src/agent_install.rs` 의 legacy 경로를 읽기만 했고, 에이전트 파일에 그대로 맞는지는 슬라이스 2 가 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이 SPEC 은 이름 바꾸기만 하고 게이트 규칙은 바꾸지 않는다 | 사용자 선택 "영어도 review로"(2026-11-27)와 한국어 낱말 결정의 기준 | SPEC 을 새 개정으로 |
| D-2 | 슬라이스 1·2·3 은 지금, 슬라이스 4 는 다른 카드가 랜딩된 뒤 릴리스 전에 한 번 | 레인 제안: 더하기만 하는 앞 슬라이스는 충돌이 적고 기계적 4번은 열린 브랜치가 없을 때 가장 싸요. 사용자가 승인 때 고름 | Interview 1 의 다른 안 |
| D-3 | 옛 명령·옵션 이름은 릴리스 한 번 동안 받고, 같은 결과에 stderr 한 줄(새 이름 안내)을 낸다 | 레인 제안: 레인 스크립트와 오래된 지시가 한 번에 안 깨져요. 사용자가 승인 때 고름 | Interview 2 의 다른 안 |
| D-4 | 상태 컨텍스트는 `moai/audit` 와 `moai/review` 를 함께 올리고, 룰셋 편집은 사용자의 단계다 | 레인 제안: 룰셋이 요구하는 컨텍스트가 하루라도 안 올라오면 릴리스 PR 이 막혀요. 사용자가 승인 때 고름 | Interview 3 의 다른 안 |
| D-5 | 저장된 이름(카드 종류 값 `audit`)은 그대로 두고, 새로 쓰는 이벤트·JSON 필드는 새 이름으로 쓰고 읽기는 둘 다 한다 | 레인 제안: 종류 열은 CHECK 가 걸려 있어 바꾸려면 레저 스키마 이동이 필요해요. 사용자가 승인 때 고름 | Interview 4 의 다른 안 |

## Interview

아직 사용자의 답이 없어서 `- Qn: … → A: …` 항목이 하나도 없어요(그래서 `moai spec check` 가 `interview_entry` 경고를 내요). 사용자가 고르면 그 선택을 항목으로 옮기고, 승인은 그다음 사용자의 몫이에요. 아래 네 질문이 사용자가 고를 것이고, 기본값(권장)이 Decisions 의 D-2 ~ D-5 예요.

1. 조각 경계(D-2). 선택지: (가) 한 번에 지금: 활성 카드 9개와 리뷰 대기 카드 42개의 브랜치가 거의 다 충돌해요. (나) 한 번에 릴리스 직전: 충돌은 거의 없지만 tier 2 기계적 diff 하나를 리뷰해야 하고, 읽기 호환이 늦게 들어와서 새 이름 위에 쌓은 카드가 먼저 막힐 수 있어요. (다, 권장) 앞 세 슬라이스는 지금, 이름 바꾸기와 쓰는 쪽 전환은 마지막에 한 번: 충돌이 적은 순서이고 읽기가 쓰기보다 먼저예요.
2. 옛 명령·옵션 이름(D-3). 선택지: (가, 권장) 릴리스 한 번 동안 받고 stderr 로 안내해요: 이름이 두 벌로 남고 시험이 짝으로 늘어요. (나) 받지 않고 새 이름을 알려 주며 거절해요: 레인 스크립트와 지난 지시가 한 번에 깨져요. (다) 영원히 받아요: 이름이 하나로 모이지 않아요.
3. GitHub 상태 컨텍스트(D-4). 선택지: (가, 권장) `moai/audit` 와 `moai/review` 를 함께 올리다가 사용자가 룰셋을 고친 뒤 옛 것을 멈춰요. (나) 한 번에 바꾸고 룰셋도 같은 때 고쳐요: 그 사이에 아무도 올리지 않는 컨텍스트를 기다리는 PR 이 생겨요. (다) GitHub 쪽 이름만 `moai/audit` 로 영원히 둬요: 안은 review, 밖은 audit 이에요.
4. 저장된 이름(D-5). 선택지: (가, 권장) 종류 값 `audit` 은 두고, 새 쓰기는 새 이름이며 읽기는 둘 다 해요. (나) 레저 스키마를 옮겨 종류 값도 `review` 로 해요: 게이트가 읽는 저장소의 마이그레이션이고, 카드 종류 `review` 와 상태 `review` 가 겹쳐요. (다) 저장 이름은 전부 옛 것으로 두어요: 새 이름이 레저에 안 나타나요.

`moai update` 의 옮김(REQ-11, REQ-12)은 사용자가 이미 고른 방식("update가 옮김", 옛 파일은 되돌리기 기록과 함께 보관, 지우지 않음)과 같은 모양이라 따로 묻지 않아요.

## Result

초안이에요. 구현은 승인 뒤 슬라이스 카드가 해요. 이 초안을 쓰는 동안 측정만 했고 코드는 바꾸지 않았어요.
