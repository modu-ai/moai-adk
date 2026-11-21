---
id: SPEC-20261121-remove-typesafe
title: "Remove TypeSafe and the Jev skills from the shipped assets"
status: approved
revision: 1
created: 2026-11-21
approval: {by: user, at: 2026-11-21, basis: "AskUserQuestion selection: 승인 (Recommended)", digest: "sha256:14c81a4072f9e86d6c35c43202c97cab756e3ce071f797388e84c7c119fa709e"}
---

# Remove TypeSafe and the Jev skills from the shipped assets

## Intent

사용자가 "TypeSafe 는 더 이상 쓰지 않는다. 제거"와 "결정 모델 모두 제거"를 지시했어요. 앞선 SPEC 두 개(drop-decision-question, retire-decide)가 인터뷰 질문과 `src/decide/` 코드를 지웠고, 이번에는 제품에 실려 나가는 남은 부분을 지워요. 읽어서 확인한 남은 곳은 다음이에요. (1) `templates/bundle-packs/` 의 스킬 3개 — `moai-fetch-web`(Jev 점수 매기기에 기대는 웹 발췌 도구, TypeSafe 인증 필요), `moai-jev-skill-suggestion`, `moai-ref-jev-question-design` — 와 `packs.json` 의 항목 3개. (2) 같은 묶음의 에이전트 `manager-todo.md` 안의 Jev 상담 규칙(`## Jev Decision Boundary` 와 선택 순서 문장). (3) 예전 세대 비밀 파일 이름 `.env.typesafe` 를 예시로 쓰는 `src/home_layout.rs`, `src/home_retire.rs`, `tests/home_retire_cli.rs` 의 주석·시험 입력 — 파일 이름 규칙(`.env*` 패턴)은 그대로 두고 예시 이름만 중립 이름으로 바꿔요. 일부러 남기는 것: 예전 프로필의 `decision_model` 값을 계속 받아 주는지 확인하는 `tests/project_decide_model.rs`(데이터 호환), Go 세대 목록을 분류한 `tools/legacy_verdicts.py`, 사용자 PATH 의 예전 shim 본문을 흉내 낸 `src/shim_check.rs` 시험 글, Go 코드 모음 fixture, `docs/` 의 결정 기록·조사 문서(과거 기록), Go 저장소 `moai-adk-go`(다른 제품), 사용자 실제 HOME 의 파일(`~/.moai/.env.typesafe`, `~/.moai/cache/jev` — 이 작업은 만지지 않아요).

## Requirements

- REQ-1: When the pack catalog is listed or a pack is installed, the catalog shall carry none of the skills `moai-fetch-web`, `moai-jev-skill-suggestion` and `moai-ref-jev-question-design` · failure: none — no new failure; an unknown name is refused as before.
- REQ-2: When the files under `templates/bundle-packs/` are read, the folder shall hold no file for those three skills · failure: none — a file removal; the CLI test and the find check show any leftover.
- REQ-3: When the `manager-todo` agent text is read, the text shall hold no Jev consultation rule, and the queue rules it does state stay as they are · failure: none — a text edit; the grep checks show any leftover.
- REQ-4: When the catalog entry of `manager-todo` is read, the entry shall carry the sha256 of the edited file · failure: none — the CLI test reads the catalog and the file together.
- REQ-5: When an earlier-generation secret file is classified, `src/home_layout.rs` and `src/home_retire.rs` shall keep the `.env*` rule with its retire-by-rename decision and no longer name TypeSafe in comments or tests · failure: none — behavior is unchanged; the existing tests pin it.

## Acceptance

- AC-1 (REQ-1, REQ-2): 목록에서 세 이름이 사라지고 파일도 없음. verify: `grep -c -E 'moai-fetch-web|moai-jev-skill-suggestion|moai-ref-jev-question-design' templates/bundle-packs/packs.json` (base 3 이상, head 0); `find templates/bundle-packs -iname '*jev*' -o -iname '*fetch-web*'` (head 비어 있음); `cargo test --locked --offline --test bundle_user_assets_cli`
- AC-2 (REQ-3): `grep -c -i -E 'jev|typesafe' templates/bundle-packs/.claude/agents/moai/manager-todo.md` 가 head 에서 0. 큐 규칙 문단(`Serial-cycle contract`, `Judgment Sub-Role`, `Model/effort escalation`)이 그대로 있음. verify: `grep -c -E '^## (Judgment Sub-Role|Model/effort escalation)' ...manager-todo.md` 가 2
- AC-3 (REQ-4): `packs.json` 의 manager-todo 항목 sha256 이 파일의 sha256 과 같음. verify: `shasum -a 256 templates/bundle-packs/.claude/agents/moai/manager-todo.md` 와 `grep -A5 '"name": "manager-todo"' templates/bundle-packs/packs.json`
- AC-4 (REQ-5): 소스 시험이 통과하고 해당 파일들에 `typesafe` 가 없음. verify: `cargo test --locked --offline --lib home_`; `cargo test --locked --offline --test home_retire_cli`; `grep -c -i typesafe src/home_layout.rs src/home_retire.rs tests/home_retire_cli.rs` 가 각각 0
- AC-5: 실제 코드 호출처가 없음을 한 번 더 확인하고 가벼운 검사 통과. verify: `git grep -n -i -E 'typesafe|\bjev\b' -- src tests templates` 의 결과가 `tests/project_decide_model.rs` 와 `src/shim_check.rs` 의 시험 글뿐; `git diff --check`; `cargo fmt --check`; `moai check affected <base>..<head>` PASS

## Scope

- write: `templates/bundle-packs/packs.json`, `templates/bundle-packs/.claude/skills/moai-fetch-web/`, `templates/bundle-packs/.claude/skills/moai-jev-skill-suggestion/`, `templates/bundle-packs/.claude/skills/moai-ref-jev-question-design/`, `templates/bundle-packs/.claude/agents/moai/manager-todo.md`, `src/home_layout.rs`, `src/home_retire.rs`, `tests/home_retire_cli.rs`, `.moai/specs/SPEC-20261121-remove-typesafe/spec.md`
- read-only: `build.rs`, `src/user_assets.rs`, `tests/bundle_user_assets_cli.rs`, `tools/repo_scans.json`, `tools/legacy_verdicts.py`, `tests/project_decide_model.rs`, `src/shim_check.rs`, `docs/`
- forbidden: 실제 HOME·키 보관소·네트워크, Go 저장소 `moai-adk-go`, 카드 번호를 SPEC id·브랜치·커밋에 사용, push

## Interview

- Q1: 남은 TypeSafe 와 결정 모델을 어떻게 할까요? → A: "TypeSafe 더이상 사용하지 않는다. 제거", "결정모드 모두 다 제거 해라" (사용자, 2026-11-21 대화)
- Q2: 위 범위(스킬 3개·packs.json 항목·manager-todo 의 Jev 규칙 제거, .env.typesafe 예시 이름 교체; 호환 시험·Go 세대 분류표·과거 기록·Go 저장소·실제 HOME 은 유지)로 승인할까요? → A: 승인 (Recommended)

## Plan

1. 세 스킬 폴더를 `git rm` 으로 지우고 `packs.json` 에서 항목 3개(commands 묶음의 moai-fetch-web, ops-tools 묶음의 나머지 둘)를 빼요. ops-tools 설명에서 "Jev," 를 빼요.
2. `manager-todo.md` 에서 frontmatter 설명의 Jev 구절, 선택 순서 문장의 Jev 부분, `## Jev Decision Boundary` 절을 지우고 그 문장들을 "기록된 우선순위와 큐 순서" 기준으로 다시 써요. 큐 상태·완료·승인 규칙은 건드리지 않아요. 그 뒤 파일 sha256 을 다시 재서 `packs.json` 항목에 넣어요.
3. `src/home_layout.rs`, `src/home_retire.rs`, `tests/home_retire_cli.rs` 에서 `.env.typesafe` 예시를 `.env.legacy` 로 바꾸고 주석에서 TypeSafe 를 빼요. `.env*` 행과 결정(`SecretRenameOnly`)은 그대로예요.
4. `bundle_user_assets_cli`, `home_` lib 시험, `home_retire_cli`, 가벼운 검사, `moai check affected` 를 돌려요.
5. 독립 감사 뒤 병합해요.
