---
id: SPEC-20261122-project-scale
title: "The project setup tells a small project from a large tree or a monorepo and offers Claude Code's own large-codebase setup, written only on the user's pick"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# The project setup tells a small project from a large tree or a monorepo and offers Claude Code's own large-codebase setup, written only on the user's pick

## Intent

사용자 요구(2026-11-22): moai-adk 는 프로젝트 크기와 상관없이 Claude Code 의 자체 기능만으로 잘 돌아야 해요. 첫 큰 코드베이스는 사용자의 다른 프로젝트예요. `/moai project` 는 규모를 작은 프로젝트(`small`), 큰 단일 트리(`large_tree`), 패키지가 있는 모노레포(`monorepo`)로 읽고, 사용자가 고르면 공식 문서 `large-codebases` 의 설정을 미리보기로 보여 주고 승인 뒤에만 써요. 부모가 `SPEC-20261121-flow-skeleton` 인 까닭: 그 SPEC 이 재설계 전체의 마스터이고 각 단계를 자식 SPEC 으로 나눠 실어요. 이 SPEC 은 그 가운데 프로젝트 쪽 하네스(`/moai project`)가 프로젝트 크기에 맞는지를 다루는 한 단계예요. 이 SPEC 은 규모를 읽고 고르고 미리보기·적용하는 틀이고, 어떤 설정을 무엇으로 쓰는지는 자식 `SPEC-20261122-project-scale-settings` 가 정해요. 코드는 뒤 조각이 써요.

설정 하나하나는 Claude Code 자신의 키와 파일이에요. moai 만의 규모 키나 새 형식은 만들지 않아요(프로필에 사용자의 선택 하나만 더해요). 기본으로는 아무것도 쓰지 않아요: 프로젝트의 `.claude/` 아래 파일은 사용자가 골라서 쓰는 맞춤 하네스예요(사용자 결정 2026-11-22: 프로젝트 `.claude/rules` 등은 moai 가 기본으로 쓰지 않아요).

두 가지가 설계를 정해요. (1) 슬라이스 0 측정(보고서와 증거는 moai 홈의 보고서 폴더 `project-scale-probe-2026-11-22`, Claude Code 2.1.330): 루트에서 시작하면 영역의 `CLAUDE.md`, 영역의 `.claude/rules/` 폴더, `paths:` 가 있는 규칙은 그 영역의 파일을 건드릴 때까지 로드되지 않고, 건드리면 그 영역 것만 들어와요; 패키지 폴더에서 시작하면 그 패키지의 `CLAUDE.md` 와 중첩 규칙은 시작 때 로드되지만 루트의 `paths:` 규칙은 로드되지 않아요; `paths:` 가 없는 규칙은 언제나 로드돼요. (2) 공식 문서도 블로그도 "큰 코드베이스" 의 숫자 기준을 주지 않아요. 기준은 moai 의 기본값이고(D-1) 사용자의 선택이 감지를 이겨요.

## Non-goals

- 사용자의 기존 파일을 덮어쓰지 않아요: 이미 있는 것은 충돌로 남겨 두고 말해요.
- 언어 서버 바이너리, 플러그인, 마켓플레이스를 설치하지 않아요.
- 관리형 설정, 조직 배포, MCP 서버, 훅 설정, 비용 상한을 다루지 않아요.
- 저장소 설정(브랜치 보호 등)과 사용자 설정 폴더에 쓰지 않아요.
- 규모 숫자를 프로필에 사실처럼 저장하지 않아요: 저장하는 것은 사용자의 선택이에요.

## Requirements

- REQ-1: When `project-detect` runs, the detector shall report `scale` with a `kind` of `small`, `large_tree` or `monorepo` and the numbers it read · failure: none — the kind is a pure function of those numbers and is never stored as a fact the user did not pick.
- REQ-2: When a workspace marker is present or three or more manifest folders sit below the root, the detector shall read `monorepo` and list the package folders · failure: none — the markers are `pnpm-workspace.yaml`, `lerna.json`, `nx.json`, `turbo.json`, `rush.json`, `go.work`, a `workspaces` key in the root `package.json` and a `[workspace]` table in the root `Cargo.toml`.
- REQ-3: When no monorepo rule holds and the repository tracks 5,000 files or more, the detector shall read `large_tree` · failure: none — any other project reads `small`.
- REQ-4: The detector shall count tracked files from the entry count in the header of the git index without starting a program · failure: none — a missing or unreadable index counts the files of its bounded walk and says the count is a lower bound.
- REQ-5: When the scale is read, the PROJECT route shall ask one selection that names the detected kind first and the other two after it · failure: none — a cancelled or empty answer offers nothing and writes nothing.
- REQ-6: When the user picks `monorepo` or `large_tree`, the planner shall return a preview of every file it would create or change with the documentation line each setting comes from · failure: none — a pick of `small` plans nothing and the preview says so.
- REQ-7: The planner shall write no file · failure: none — only the apply writes, and only after the user approves the plan's digest.
- REQ-8 (REQ-6): IF a target file exists and its content was not written by this setup, THEN the plan shall keep it and report a conflict for that row.
- REQ-9 (REQ-7): IF the project changed since the plan or the digest does not match, THEN the apply shall refuse and write nothing.
- REQ-10: The preview shall carry one paragraph that says where to start Claude, that a package's settings file is not inherited from its parent and that a plan Claude writes in plan mode survives compaction · failure: none — the paragraph is text and writes nothing.
- REQ-11 (REQ-6): IF a planned value looks like a credential or a path leaves the repository, THEN the planner shall refuse the whole plan and name the value's key.
- REQ-12: When the user picks a scale, the profile shall keep the pick in an additive `scale` key so that `moai update` plans from it without asking again · failure: none — a project without the key shows no scale step.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-6, REQ-7): 임시 고정물 셋(작은 프로젝트, 추적 파일 5,000개 이상의 큰 단일 트리, 작업 공간 표지가 있는 모노레포)이 서로 다른 미리보기 셋을 내고, 키는 공식 설정 이름만 쓰고, 계획이 끝난 뒤 트리가 그대로예요 · verify: `cargo test --locked --offline --test project_scale_plan_lib three_scales_give_three_previews_and_write_nothing`
- AC-2 (REQ-2, REQ-3, REQ-4): 경계값: 4,999 와 5,000 개, 매니페스트 폴더 2 개와 3 개, 작업 공간 표지 각각, 인덱스가 없을 때의 하한 표시 · verify: `cargo test --locked --offline --test project_scale_detect_lib the_thresholds_and_markers_decide_the_kind`
- AC-3 (REQ-8, REQ-9): 이미 있는 파일과 설정 키는 충돌로 남고 바이트가 그대로예요; 계획 뒤 프로젝트가 바뀌면 적용이 거절해요 · verify: `cargo test --locked --offline --test project_scale_plan_lib an_existing_file_is_a_conflict_and_a_changed_project_is_refused`
- AC-4 (REQ-11): 비밀 모양 값이나 저장소를 벗어나는 경로가 하나라도 있으면 계획 전체가 거절되고 그 키 이름을 말해요 · verify: `cargo test --locked --offline --test project_scale_plan_lib a_credential_or_an_outside_path_refuses_the_whole_plan`
- AC-5 (REQ-5, REQ-10, REQ-12): 인터뷰 행(감지된 규모가 첫 선택지), 위치 안내 문단, 프로필의 `scale` 키와 갱신 단계의 재계획 · verify: `cargo test --locked --offline --test project_scale_cli the_pick_is_kept_and_update_plans_from_it`

## Scope

- write: `.moai/specs/SPEC-20261122-project-scale/**` 와 `.moai/specs/SPEC-20261122-project-scale-settings/**`(이 조각), 뒤 조각에서 카드 범위에 더해요: `src/project_detect.rs`(규모 읽기), 새 모듈 `src/project_scale.rs` 와 연산 둘과 문서 페이지, `src/profile.rs`(`scale` 키), `src/project_update.rs`(단계 하나), 질문 행과 워크플로 절(스킬 원본 목록 포함), 시험 파일. 설정 쓰기 규칙은 자식 SPEC 의 범위예요.
- read-only: the slice 0 probe report and its evidence folder (in the reports folder of the user's moai home), `SPEC-20261122-settings-write-check`, `SPEC-20261122-secret-file-deny-rules`, `SPEC-20261122-worktree-place`, `SPEC-20261122-claude-md-project-layer`
- forbidden: 사용자의 Claude 설정 폴더에 쓰기, 저장소 설정 쓰기, 플러그인 설치 실행, 프로젝트 이름을 코드에 박기, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰기, 실제 HOME

## Plan

1. 조각 1(이번): 슬라이스 0 측정과 이 SPEC 둘, 문서 인용(자식의 Evidence).
2. 조각 A: `project-detect` 의 `scale`(REQ-1 에서 REQ-4, AC-2). 인덱스 헤더 읽기는 버전 태그를 읽는 규칙과 같은 링크 워크트리 규칙을 따라요.
3. 조각 B: 자식 SPEC 의 쓰기 규칙 값 검사. `SPEC-20261122-settings-write-check` 의 쓰기 길을 쓰고 값 검사만 더해요.
4. 조각 C: 계획과 적용(REQ-6 에서 REQ-11, AC-1, AC-3, AC-4). 자기 계획·적용 연산, 자기 기준선 매니페스트와 영수증(CI 계획기와 같은 엔진 모양).
5. 조각 D: 인터뷰 행, 워크플로 절, 프로필 `scale` 키, 갱신 단계(REQ-5, REQ-10, REQ-12, AC-5).

### Interview rows

한 번에 최대 세 질문, 모두 선택, 권장 답이 먼저예요. 읽을 수 있는 것은 묻지 않고 확인만 해요.

| id | 묻는 때 | 질문 | 선택지 | 바뀌는 계획 |
| --- | --- | --- | --- | --- |
| `scale` | 프로필에 `scale` 이 없을 때 | 이 프로젝트는 어느 쪽에 가까운가요? | 감지한 종류(권장), 나머지 둘 | 아래 모든 행의 유무 |
| `scale_packages` | 모노레포 | 평소 일하지 않는 패키지가 있나요? | 패키지 목록에서 여러 개, 없음(권장) | `claudeMdExcludes` |
| `scale_generated` | 추적되는 생성·외부 폴더가 보일 때 | Claude 가 열지 않았으면 하는 폴더가 있나요? | 찾은 폴더(여러 개, 권장), 없음 | `Read` 거절 규칙 |
| `scale_start` | 모노레포 | Claude 를 어디서 시작해요? | 저장소 맨 위(권장), 패키지 폴더 | 패키지 설정, 위치 안내 문단 |
| `scale_worktree` | 큰 트리 또는 모노레포 | 작업 폴더(워크트리)를 필요한 폴더만 받게 할까요? | 네(권장), 아니요 | `sparsePaths`, `symlinkDirectories`, `.gitignore` 한 줄 |
| `scale_plugin` | 지원되는 언어가 감지될 때 | 언어 서버 플러그인을 켤까요? | 감지한 언어의 플러그인(권장), 켜지 않음 | `enabledPlugins` 와 설치 명령 글 |

## Risks

- 슬라이스 0 의 지연 로딩 측정은 `@` 언급으로 파일을 건드렸어요(가짜 모델은 도구를 부르지 못해요). 문서는 읽기·쓰기·고치기 도구를 말해요. 두 길의 `load_reason` 이 같은지는 도구 호출을 낼 수 있는 가짜 모델로 확인해야 해요. 확인 전까지 SPEC 은 문서의 문장과 측정을 둘 다 인용해요.
- 문서 줄은 버전마다 달라질 수 있어요. 인용 데이터 파일이 읽은 날짜와 Claude Code 버전을 갖고, 오래되면 미리보기가 날짜를 보여 줘요.
- 감지는 지금 루트 매니페스트만 읽고 깊이 3, 5,000 개에서 멈춰요. 큰 저장소를 위해 인덱스 헤더의 개수를 쓰고, 읽지 못하면 하한이라고 밝혀요.
- 큰 저장소의 인덱스 헤더는 추적 파일 수일 뿐 소스 줄 수가 아니에요. 기준은 파일 수이고 사용자 선택이 이겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기본 기준: 추적 파일 5,000 개 이상이면 `large_tree`, 작업 공간 표지가 있거나 매니페스트 폴더가 3 개 이상이면 `monorepo`, 그 밖은 `small`. 사용자의 선택이 감지를 이겨요. | 문서와 블로그는 숫자 기준을 주지 않아요(블로그를 끝까지 읽었어요). 이 컴퓨터의 두 저장소를 읽었어요: 추적 파일 2,355 개와 매니페스트 2 개, 추적 파일 25,494 개와 매니페스트 32 개. | 카드 열림 |
| D-2 | 쓰기는 모두 사용자의 선택이고 기본은 없어요. 프로젝트 `.claude/` 아래 파일은 맞춤 하네스예요. | 사용자 결정 2026-11-22(프로젝트 `.claude/rules` 등은 사용자가 필요할 때만). | 카드 열림 |
| D-3 | 파일 수는 인덱스 헤더에서 읽고 프로그램을 시작하지 않아요. | 감지는 프로그램을 시작하지 않는 규칙이에요(`project_detect.rs` 머리말). | 카드 열림 |

## Result

(status draft — written when the slice lands)
