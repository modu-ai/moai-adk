---
id: SPEC-20261121-style-banner
title: "Output style banners read MoAI 🗿"
status: draft
revision: 1
created: 2026-11-21
---

# Output style banners read MoAI 🗿

## Intent

사용자 지시 2026-11-21(리더 배차): 출력 스타일 배너 첫 줄을 `<이모지> MoAI-Easy ★ <이름> ─…`/`MoAI ★` 에서 `<이모지> MoAI 🗿 <이름> ─…` 로 바꿔요. 배포되는 스타일 참조·스타일 템플릿·그 문구를 못 박은 시험·스캔 표·매니페스트를 함께 고쳐요. 사용자의 `~/.claude/CLAUDE.md` 는 리더가 이미 바꿨어요.

## Non-goals

- 스타일 이름(`MoAI-Easy`), 스타일 파일명(`output-styles/MoAI-Easy.md`), `settings.json` 의 `outputStyle` 값, NOTICE·문서의 역사 기술은 바꾸지 않아요.
- 배너 이외의 진행판·아이콘 규칙은 바꾸지 않아요.
- 사용자 HOME 을 쓰지 않아요(설치된 쪽은 다음 `moai update` 가 새 소스에서 다시 적용해요).

## Requirements

- REQ-1: The shipped output style reference shall state the banner first line as `<emoji> MoAI 🗿 <name> ─…` with the 46 `─` closing line unchanged.
- REQ-2: The shipped style templates shall show their banner examples with `MoAI 🗿` where they showed `MoAI-Easy ★` or `MoAI ★`.
- REQ-3: The repository scan table shall reject the old banner wordings in shipped paths.
- REQ-4: The style name `MoAI-Easy` and its install paths shall stay unchanged.

## Acceptance

- AC-1 (REQ-1): 배포 스킬 참조의 스타일 소스 파일(`references` 바로 아래 `output-style.md`)의 배너 문장이 `MoAI 🗿` 이고 옛 문장이 없어요.
  verify: `cargo test --offline --locked --test style_banner_scan`
- AC-2 (REQ-2): `templates/claude/output-styles/moai-easy.md` 와 `moai.md` 의 배너 예시가 `MoAI 🗿` 이고 옛 문장이 없어요.
  verify: `cargo test --offline --locked --test style_banner_scan`
- AC-3 (REQ-3): 스캔이 `templates/`·`skills/moai/` 에서 옛 문장을 찾으면 실패해요.
  verify: `python3 -m unittest tests.repo_scans_test` 와 위 스캔
- AC-4 (REQ-4): 스타일 이름과 설치 경로는 그대로예요.
  verify: `rg -n 'MoAI-Easy ★|MoAI ★' skills/moai templates` 결과 없음과 `git diff --stat` 에 이름 파일 없음

## Scope

- write: `.moai/specs/SPEC-20261121-style-banner/**`, `skills/moai/references/*/output-style.md`(한 줄), `templates/claude/output-styles/moai-easy.md`(한 줄), `templates/claude/output-styles/moai.md`(두 줄), `tests/style_banner_scan.rs`(새 파일), `tools/repo_scans.json`(항목 하나), `skills/moai/references/source-manifest.json`(해시 갱신)
- read-only: `src/home_style.rs`(렌더러, 소스 문장만 읽음), `NOTICE.md`
- forbidden: 스타일 이름·파일명·`outputStyle` 값 바꾸기, 사용자 HOME 쓰기, 배너 외 문구 바꾸기

## Plan

1. 이 SPEC 을 먼저 커밋해요.
2. RED: `tests/style_banner_scan.rs` 두 시험이 옛 문장에서 실패하는 것을 눈으로 봐요.
3. GREEN: 네 줄만 고치고 `python3 -I tools/source_manifest.py write` 로 매니페스트를 갱신해요.
4. 확인: 위 verify 명령과 `tests.repo_scans_test`·`tests.source_manifest_test`·`tests.skill_references_test`·`product_cli skill_package`, `git diff --check`, fmt, clippy `--all-targets`, trailer-check.

## Risks

- 스캔 표에 항목을 더하면 `tests/repo_scans_test.py` 의 완전성 시험이 걸려요: 규칙 파일·범위를 표 형식대로 같이 더해요.
- `MoAI ★` 은 `MoAI-Easy ★` 의 부분 문자열이 아니에요: 두 패턴을 따로 잡아요.
- 매니페스트는 공용 파일이라 다른 변경과 부딪힐 수 있어요: 이 카드의 예약에 넣고 착지 순서에 따라 다시 맞춰요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 배너 문구만 바꾸고 스타일 이름·파일명은 그대로 둔다 | 지시는 첫 줄 표기에 관한 것이고 이름은 설치·설정 식별자예요 | 문장을 되돌리는 커밋 |
| D-2 | 옛 문장 검사는 새 스캔 파일과 표 항목으로 한다 | 기존 스캔 파일은 다른 카드가 쥐고 있어요 | 항목 제거 |
