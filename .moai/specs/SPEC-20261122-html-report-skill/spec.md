---
id: SPEC-20261122-html-report-skill
title: "Every moai install carries one internal html-report skill: six templates, Pretendard and JetBrains Mono with mermaid by CDN in html files, Google Fonts and pre.mermaid in artifacts"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-report-rules
---

# Every moai install carries one internal html-report skill: six templates, Pretendard and JetBrains Mono with mermaid by CDN in html files, Google Fonts and pre.mermaid in artifacts

## Intent

사용자 결정(2026-11-22): moai-adk-go 의 `moai-domain-html-report` 스킬(SKILL.md 와 참고 글 넷, 템플릿 여섯)을 moai-adk 로 가져와 모든 html 보고서와 Claude 아티팩트의 바탕으로 써요. 이어진 결정 셋이 모양을 정해요.

- 전역 설치: "글로벌로 설치해서 모든 프로젝트에 사용해야 한다". moai 가 각 Claude 설정 폴더에 까는 패키지에 들어가요. 프로젝트 파일이 아니에요.
- 내부 스킬: "moai-adk 설치시 이게 포함이 되어야 한다. 사용자에게 /리스트에 나올 필요 없고 내부에서 사용하는 스킬이다". 머리말에 `user-invocable: false` 를 두고, 업데이트 미리보기가 moai 스킬과 함께 보여 주며, 시험이 둘 다 고정해요.
- 글꼴: html 파일은 Pretendard(jsDelivr)와 JetBrains Mono(Google Fonts)를 쓰고, mermaid 는 CDN 스크립트로 그려요. 아티팩트는 보고서 언어에 맞는 Noto Sans(KR, JP, SC, TC)를 Google Fonts 에서 받고, 다이어그램은 `pre.mermaid` 로 둬요.

부모가 `SPEC-20261122-report-rules` 인 까닭: 그 SPEC 이 보고서의 기록(md)과 그 옆의 html, 한 번의 비공개 게시를 정했고, 이 스킬은 그 html 과 아티팩트를 그리는 쪽이에요.

원본은 끝까지 읽었어요(2026-11-22, `/Users/goos/moai/moai-adk-go/.claude/skills/moai-domain-html-report`, 11개 파일, 165,402 바이트). 저장소의 `templates/bundle-packs/.claude/skills/moai-domain-html-report/` 는 원본과 같은 사본이고, 지금은 사용자가 고를 때만 까는 `creative` 묶음에 들어 있어요(`moai bundle add`). 리더 검토 보고서(`~/.moai/reports/moai-adk/html-report-skill-review-2026-11-22.md`, 91개 항목)도 끝까지 읽었어요. 리더가 설치본에서 고친 두 가지는 이 SPEC 이 패키지 쪽에 옮겨요. 하나는 아티팩트 글꼴 블록이 html 머리에 새는 문제(A6)이고, 다른 하나는 HTML 칸이 중괄호 두 개라서 태그가 글자로 보이는 문제(I1, P1, X1, R1)예요.

지금 `moai update` 는 `moai` 스킬 하나만 복사해요(`src/update.rs` 의 `SKILL_DIR = "moai"`). 미리보기도 스킬 하나만 보여 줄 수 있어요. 그래서 설치는 코드 변경이고, 그 파일들은 다른 카드가 잡고 있어서 조각 B 로 나눠요.

## Non-goals

- mustache 엔진을 moai 에 넣지 않아요. 템플릿을 읽고 값을 채워 완성 html 을 쓰는 것은 에이전트예요(원본도 엔진이 없어요).
- 아티팩트를 공유하거나 지우지 않아요. 게시 규칙은 부모 SPEC 이 정해요.
- 원본 저장소 moai-adk-go 는 고치지 않아요.
- 리더가 손으로 깐 설치본(`~/.claude/skills/moai-domain-html-report`)을 moai 가 지우지 않아요. 정리는 사용자가 골라요(Risks 참고).
- 원본에 템플릿이 없는 `editorial`·`legal` 모드는 가져오지 않아요.

## Requirements

- REQ-1: The moai package shall carry the html-report skill as a second folder under `skills/` beside the moai skill · failure: none — the release archive already ships the whole `skills/` folder.
- REQ-2: The html-report SKILL.md shall set `user-invocable` to false · failure: none — Claude still loads the skill when a report is rendered.
- REQ-3: An html report file shall load Pretendard from jsDelivr and JetBrains Mono from Google Fonts in every mode · failure: none — offline readers see the system fonts.
- REQ-4: When an html report file holds a diagram, the file shall carry it as `pre.mermaid` with one mermaid script from jsDelivr at an exact version · failure: none — REQ-11 handles a script that does not run.
- REQ-5: The artifact version of a report shall load stylesheets only from Google Fonts with the Noto Sans family of the report's language · failure: none — no other stylesheet host is named in the artifact text.
- REQ-6: The artifact version of a report shall carry each diagram as `pre.mermaid` without a mermaid script · failure: none — the artifact viewer renders it.
- REQ-7: Each template shall write an HTML field with three braces and every other field with two braces · failure: none — an outside text is escaped before it reaches an HTML field.
- REQ-8: The skill text shall name no Go-era settings path, no withdrawn host and no report folder inside the project · failure: none.
- REQ-9: The skill text shall follow the parent report rule in this order: the md record, the html beside it, one private publish · failure: a refused publish gives the two paths and one reason line.
- REQ-10: Each item of the leader review shall be fixed by a named slice or recorded in the Result with its reason · failure: none.
- REQ-11 (REQ-4): IF the mermaid script does not run, THEN the html report file shall still show the diagram text and its plain-language fallback.

설치(`moai update` 가 두 스킬을 복사하고 미리보기에 한 줄씩 보여 주는 것, 묶음 목록의 기본 이름, moai 가 깔지 않은 같은 이름 폴더)는 이 SPEC 을 부모로 하는 자식 SPEC `SPEC-20261122-html-report-install` 이 정해요(조각 B). 요구가 열둘을 넘지 않게 나눴어요.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-8): `skills/moai-html-report/` 가 SKILL.md, 참고 글 넷, 템플릿 여섯을 가지고, SKILL.md 머리말이 `user-invocable: false` 이며, 스킬 글 어디에도 `report.yaml`, `report.format`, `Codex`, `skill-routing doctrine`, 프로젝트 안 `.moai/reports/` 가 없어요. `tests/dist_include_test.py` 의 스킬 폴더 목록이 두 이름을 가져요 · verify: `cargo test --locked --offline --test html_report_skill the_package_carries_the_internal_skill_with_no_go_era_text`
- AC-2 (REQ-3, REQ-4, REQ-7, REQ-11): 여섯 템플릿의 html 머리가 Pretendard(jsDelivr)·JetBrains Mono(Google Fonts) 링크를 가지고 Noto 글꼴 링크나 Noto 글꼴 토큰은 갖지 않아요. `_html` 로 끝나는 칸과 `diagram_svg` 는 세 중괄호이고, 남은 칸은 두 중괄호예요. mermaid 를 쓰는 글은 정확한 버전 주소와 스크립트 없이 보이는 원문 자리를 말해요 · verify: `cargo test --locked --offline --test html_report_skill the_html_head_loads_pretendard_and_html_fields_use_three_braces`
- AC-3 (REQ-5, REQ-6): 아티팩트 글이 스타일시트 호스트를 `fonts.googleapis.com` 하나로, 글꼴 파일 호스트를 `fonts.gstatic.com` 하나로 말하고, 언어별 Noto Sans 네 가족(KR, JP, SC, TC)을 이름으로 말하며, 다이어그램은 스크립트 없는 `pre.mermaid` 라고 말해요 · verify: `cargo test --locked --offline --test html_report_skill the_artifact_text_names_google_fonts_noto_by_language_and_pre_mermaid`
- AC-4 (REQ-9, REQ-10): SKILL.md 의 전달 절이 기록 md, 옆 html, 한 번의 게시, 거절 한 줄 순서를 말하고 부모 SPEC 의 참고 글을 가리켜요. Result 가 검토 91개 항목마다 조각 이름이나 남긴 까닭을 가져요 · verify: `cargo test --locked --offline --test html_report_skill the_delivery_section_follows_the_report_rule`
  manual: 이 SPEC Result 의 검토 항목 표

## Scope

- write(조각 1, 이번): `.moai/specs/SPEC-20261122-html-report-skill/**`
- write(조각 A, 겹침 없음): `skills/moai-html-report/**`(새), `tests/html_report_skill.rs`(새), `tests/dist_include_test.py`
- write(조각 C): `skills/moai-html-report/**`, `tests/html_report_skill.rs`
- 조각 B 는 자식 SPEC 이 범위를 정해요: `src/update.rs`, `src/update_tux.rs`, `templates/bundle-packs/packs.json`, 설치 시험 셋(`tests/update_user_layer_cli.rs`, `tests/host_home_cli.rs`, `tests/product_cli.rs`), 묶음 사본. 모두 지금 다른 카드가 잡고 있어요.
- read-only: 원본 폴더, 리더 검토 보고서, `SPEC-20261122-report-rules`, `src/user_assets.rs`, `build.rs`, `dist-workspace.toml`
- forbidden: 실제 HOME 과 실제 설정 폴더 쓰기, 원본 저장소 쓰기, 아티팩트 공유·삭제, 카드 번호를 SPEC id·가지·커밋·PR·배포 파일에 쓰기, 네트워크(시험은 글만 읽어요)

## Plan

1. 조각 1(이번): 이 SPEC.
2. 조각 A: 원본을 `skills/moai-html-report/` 로 옮기고 다듬어요.
   - 머리말: `user-invocable: false`, 좁힌 `allowed-tools`, Go 시절 metadata 제거
   - 기본 글: 기록 md → 옆 html → 한 번의 게시. 보고서 폴더는 `~/.moai/reports/<project>/`, 날짜는 `date -u`
   - 그리는 순서 절: 템플릿 고르기 → 값 채우기(바깥 글 이스케이프) → 좌표 계산 → 주석 지우기 → html 쓰기 → 아티팩트 본 만들기
   - 필드 표를 템플릿 이름에 맞추기, 언어 규칙(대화 언어, `lang`)
   - 템플릿: A6 와 세 중괄호 고침, explainer·plan 의 html 글꼴을 Pretendard 로, mermaid 정확한 버전
   - 아티팩트 글: Google Fonts 만, 언어별 Noto Sans, `pre.mermaid`, 현재 아티팩트 계약(뼈대 없음, `color-scheme`, 허용 호스트 다섯과 버전 고정)
   - 시험 `tests/html_report_skill.rs`(AC-1 에서 AC-4)와 `tests/dist_include_test.py` 의 목록
3. 조각 B: 그 파일들을 잡은 카드(출력 스타일, 브라우저 참고 글, CLAUDE.md 배치)가 착지하거나 리더가 허락하면, 자식 SPEC `SPEC-20261122-html-report-install` 을 쓰고 구현해요. `moai update` 가 스킬 여러 개를 복사하고 미리보기에 한 줄씩 보여 주며, 영수증 이름 공간은 스킬마다예요. `base_names` 에 이름을 더하고, `creative` 묶음의 사본은 지우거나 새 이름을 가리켜요.
4. 조각 C: 검토의 나머지 항목(배치 B6, 대비 C8·C9, 접근성 C11, 인쇄 C5–C7, 고정 색 B4, SRI C1–C3, 각 템플릿의 D 항목)을 시험과 함께 고쳐요.

## Risks

- 이름을 바꾸면(Decisions D-1) 리더가 손으로 깐 `~/.claude/skills/moai-domain-html-report` 와 새 이름이 둘 다 남아요. moai 는 자기가 깔지 않은 폴더를 지우지 않으니, 정리는 사용자가 골라요. 조각 B 의 미리보기가 그 폴더를 "moai 가 깔지 않은 옛 사본" 으로 한 줄 보여 줘요.
- 조각 A 만 착지하면 릴리스 묶음의 `skills/` 에 새 폴더가 실리지만, `moai update` 가 아직 복사하지 않아서 사용자 설정 폴더에는 생기지 않아요. 조각 B 전까지 효과가 없는 상태이고, 해가 되지도 않아요.
- 묶음 사본과 기본 설치가 같은 이름이면 `bundle remove creative` 가 기본 설치 파일을 지울 수 있어요(`src/user_assets.rs` 가 `base_names` 밖의 이름을 지워요). 그래서 자식 SPEC 은 기본 이름 등록과 복사를 한 조각에 담아요.
- 머리말 허용 목록 시험(`tests/product_cli.rs` 의 `skill_package_frontmatter_is_host_neutral`)은 `skills/moai/SKILL.md` 만 읽어요. 이 스킬은 Claude Code 전용(사용자 결정: Claude 만 지원)이라 `user-invocable` 을 가져도 그 규칙과 부딪히지 않지만, 규칙을 넓히는 변경이 생기면 이 스킬을 빼야 해요.
- 글꼴과 mermaid 는 CDN 에서 오므로 오프라인에서는 시스템 글꼴과 다이어그램 원문으로 보여요. 원본의 "외부 의존 0" 설명은 사실대로 고쳐요(검토 A11).
- 시험은 글만 읽어요. 실제 브라우저에서의 모양(대비, 400px 배치)은 조각 C 에서 계산값과 규칙으로 고정하고, 사람 눈 확인은 하지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이름은 `moai-html-report` 예요. `moai-domain-*` 은 Go 시절 분류 이름이에요. | 리더 노트 "renamed and cleaned per the shipped skill layout rules", 작업 본문 "no Go-era names". 리더에게 결정 노트로 물어요(기본값 이 이름). | 카드 열림, 폴더 이름 하나 |
| D-2 | 아티팩트는 보고서 언어의 Noto Sans(KR, JP, SC, TC)를 Google Fonts 에서 받아요. html 파일은 Pretendard 예요. | 사용자 결정(2026-11-22T11:1xZ, 질문 도구로 고름). 아티팩트 CSP 는 스타일시트를 `fonts.googleapis.com` 에서만 받아요. | 카드 열림 |
| D-3 | html 파일의 제목도 Pretendard 예요. plan·explainer 의 명조 제목(Noto Serif KR)은 쓰지 않아요. | 사용자 결정 "Pretendard + JetBrains Mono for html files", 리더 노트(2026-11-22T11:2xZ). | 카드 열림 |
| D-4 | 인라인 스크립트는 작게 허용하고(explainer 의 탭), 외부 스크립트는 mermaid 하나만이에요. 아티팩트에는 외부 스크립트가 없어요. | 검토 B9·E8, 아티팩트 계약(인라인 허용, 외부는 허용 호스트와 정확한 버전). | 카드 열림 |
| D-5 | 브라우저를 자동으로 열지 않아요. 경로와 링크만 보여 주고, 사용자가 원할 때만 열어요. | 검토 E7, 부모 SPEC 의 전달 순서. | 카드 열림 |
| D-6 | 설치(조각 B)는 다른 카드가 잡은 파일을 쓰므로 그 카드들의 착지나 리더 허락을 기다려요. 조각 A 는 겹치는 파일이 없어요. | `moai card show` 로 읽은 예약(2026-11-22T10:1xZ): `src/update.rs`·`src/update_tux.rs` 는 출력 스타일 카드, `packs.json` 은 브라우저 참고 글 카드, 설치 시험 셋은 CLAUDE.md 배치 카드가 잡고 있어요. | 카드 열림 |

### Evidence

| 사실 | 출처 |
| --- | --- |
| `moai update` 는 스킬 하나만 복사해요 | `src/update.rs` `SKILL_DIR = "moai"`, `stage_package`, `skills_place` (끝점 `35ff0e7080` 에서 조사 에이전트가 읽음, 2026-11-22T09:38Z) |
| 릴리스 묶음은 `skills/` 전체를 실어요 | `dist-workspace.toml` `include = ["skills/", "agents/"]`, `install.sh` 의 `for part in skills agents` |
| 스킬 폴더 목록 고정 시험 | `tests/dist_include_test.py` `test_the_skills_folder_holds_only_the_moai_skill_package` |
| 묶음 제거는 `base_names` 밖을 지워요 | `src/user_assets.rs` 의 remove 경로, `templates/bundle-packs/packs.json` `base_names` |
| `user-invocable: false` 는 `/` 메뉴에서 숨기고 Claude 는 부를 수 있어요 | 공식 문서 `skills.md` 머리말 표(리더 검토가 2026-11-22T09:2xZ 에 읽음, 검토 F2) |

## Result

(status draft — written when the slices land)
