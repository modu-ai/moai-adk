---
id: SPEC-20261120-drop-omp
title: "Drop OMP support: the launcher, the detection, the update steps and the docs offer Claude Code, GLM and Codex only"
status: draft
revision: 1
created: 2026-11-20
---

# Drop OMP support: the launcher, the detection, the update steps and the docs offer Claude Code, GLM and Codex only

## Intent

사용자 결정(2026-11-19): moai-adk 의 공식 지원 호스트는 Claude Code·GLM·Codex 세 개로 좁히고 OMP(oh-my-pi) 지원을 없애요. 지금 소스 191·시험 220·문서 145 개 파일이 OMP 를 언급해요. 없애는 것: `moai omp` 런처, 설치 스크립트의 호스트 감지, update_models 와 project_hosts 의 OMP 항목(hooks·config·agents), OMP 에이전트 설치(src/omp_agents.rs), OMP 시험·고정값, 문서·help 표기. 남기는 것: 옛 프로필의 `omp` 호스트 항목은 알림 한 줄과 함께 무시하고, 예전에 OMP 용으로 쓴 파일은 `moai update` 가 '관리하지 않음' 으로 한 번 보여 주고 지우지 않아요(사용자 데이터는 건드리지 않아요). 스킬 글(SKILL.md·workflows)의 OMP 표기는 C-150 이 착지한 뒤 별개 조각에서 빼요(리더 배차).

## Non-goals

- 스킬 글(SKILL.md, skills/moai/workflows)의 OMP 문구 — C-150 과 겹쳐 그 뒤 조각이에요.
- 사용자 홈의 옛 OMP 파일 삭제·수정 — 보여 주기만 해요.
- GLM·Codex·Claude Code 지원의 어떤 변화도 없어요.

## Requirements

- REQ-1: The command `moai omp` (with any role or name flags) shall print one line that says OMP support was removed and exit with a non-zero code, without composing or starting a session.
- REQ-2: Every supported-host list of this repository shall offer claude, glm and codex only, with no update step, host table or survey offering OMP.
- REQ-3: A profile whose hosts include omp shall plan and apply with the OMP entries ignored and one notice naming the ignored host, changing none of the other hosts' entries.
- REQ-4: The command `moai update` shall list a file it managed for OMP as unmanaged once, leaving every file of a removed host undeleted and unchanged.
- REQ-5: The OMP-only source and its tests and fixtures shall be removed, with the tests that remain passing again after their OMP cases are rewritten for the remaining hosts or dropped.
- REQ-6: The skill texts (SKILL.md and skills/moai/workflows) shall keep their OMP wording unchanged until the C-150 follow-up slice lands.

## Acceptance

- AC-1 (REQ-1, REQ-2): `moai omp` 한 줄 알림·종료 코드, 세 호스트만 남은 열거·감지·갱신 표.
  verify: `cargo test --locked --offline --test omp_removal_cli`
- AC-2 (REQ-3): hosts [claude, omp] 프로필이 무시 알림 한 줄과 함께 계획·적용.
  verify: `cargo test --locked --offline --test project_hosts_cli omp_ignored`
- AC-3 (REQ-4): 옛 OMP 관리 파일의 '관리하지 않음' 한 번 표시·보존.
  verify: `cargo test --locked --offline --test home_install_apply_cli unmanaged`
- AC-4 (REQ-5): OMP 전용 파일 삭제 뒤 빌드·clippy·전 시험 대상·등재 검사 통과.
  verify: `cargo build --locked --offline && cargo clippy --locked --offline --all-targets -- -D warnings`
- AC-5 (REQ-6): 스킬 글의 OMP 문구 무변경.
  manual: 문서 읽기

## Scope

- write: `src/session.rs`, `src/main.rs`, `src/omp_agents.rs`, `src/agent_defs.rs`, `src/hooks.rs`,
  `src/project_hosts.rs`, `src/update_models.rs`, `src/home_install_apply.rs`, `install.sh`, `install.ps1`,
  `tests/omp_agents.rs`, `tests/omp_launcher.rs`, `tests/omp_host_survey_test.py`, `docs/**`,
  `.moai/specs/SPEC-20261120-drop-omp/spec.md`
- read-only: `skills/**`(C-150 뒤 조각), `src/forge/**`, `tests/common/**`
- forbidden: 사용자 홈의 파일 삭제·수정, 스킬 글 변경, GLM·Codex·Claude Code 동작 변화

## Plan

1. 조각 1(호스트 중심·런처): HOSTS 열거(session·agent_defs·hooks)에서 omp 제거, `moai omp` 한 줄 알림·실패 종료, session.rs 의 OMP 경로(argv 구성·환경·역할 안내) 제거, src/omp_agents.rs 와 tests/omp_* 삭제, 직접 이웃 시험.
2. 조각 2(설치·갱신): install.sh/ps1 감지 목록, update_models 의 OMP 모델 줄, `moai update` 의 '관리하지 않음' 한 번 표시·보존.
3. 조각 3(프로필·에이전트): project_hosts 의 omp hooks/config/agents 경로 제거, [claude, omp] 프로필 무시 알림, OMP 에이전트 설치 경로.
4. 조각 4(문서·전수): docs·help 의 OMP 표기, 남은 시험·고정값 정리, 기준 3 의 전수 검사.
5. (C-150 뒤) 조각 5: 스킬 글의 OMP 문구 제거 — 리더가 범위를 주면.

## Risks

- 넓은 삭제라 이웃 카드와 파일이 겹쳐요 — 조각마다 그 조각이 쓰는 파일만 범위에 넣고(card scope 확장), 겹치면 리더가 순서를 정해요.
- 옛 프로필을 가진 사용자의 갱신이 갑자기 실패하지 않아야 해요 — 무시 알림 한 줄로 통과(기준 2).
- 'omp' 낱말이 우연히 다른 뜻으로 쓰인 곳(예: compress)이 있으면 지우지 않아요 — 대소문자·문맥 확인.

## Result

- 조각 1(호스트 중심·런처, 2026-11-20T00:5xZ, agent-18): session.rs HOSTS 세 개로, omp argv·키·역할 안내·saved_note 만능칸 제거, agent_defs HOSTS 두 개로(render_omp 은 조각 3 까지 유지), main.rs `moai omp` 한 줄 제거 알림·종료, help.rs 열거·GROUPS·예시·한글 about 제거, tests/omp_launcher.rs 삭제. 파급 고정값: help 골든 재작성(MOAI_HELP_GOLDEN_WRITE), protocol 고정값 재작성(MOAI_UPDATE_FIXTURES), agents_home_cli(호스트 반복·개수 15→10·10→5·14→9·omp 심링크 시험 삭제), launcher_reads_models(omp 저장 칸·사례·omp 시작 시험 삭제, AC-5 낯선 도구 칸은 omp 유지), launcher_agent_entry(omp 철자 목록), launcher_auto_init(omp 행), output_names_moai·cli_name_moai·product_cli(오류 안내문), agent_defs_package(HOSTS 핀), help_layout(합격 단어 목록).
- 시험: omp_removal_cli 2(RED 2 failed 먼저 관측)·이웃 통합 12종·lib session 28 전부 통과.
- 남은 것: 조각 2(설치·갱신·관리하지 않음), 조각 3(project_hosts·omp_agents·프로필 무시 알림), 조각 4(docs·전수), (C-150 뒤) 조각 5(스킬 글).
