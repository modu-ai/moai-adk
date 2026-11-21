---
id: SPEC-20261121-detect-walk-breadth
title: "project-detect finds sub-package manifests in a large monorepo and says when its list is partial"
status: draft
revision: 1
created: 2026-11-21
---

# project-detect finds sub-package manifests in a large monorepo and says when its list is partial

## Intent

모노레포(`apps/*/package.json`, `apps/*/go.mod`, `packages/*/package.json` 등 56 개)에서 `moai project-detect .` 가 루트 `package.json` 하나만 찾아요. 같은 저장소를 읽기만 하는 실행으로 재현했고(2026-11-21), 원인은 탐색 예산이에요: 하위 manifest 탐색(`SPEC-20261119-nested-manifests`)은 폴더를 알파벳 순으로 깊이 먼저 훑으면서 폴더 항목 5000 개를 넘기면 조용히 멈춰요. `apps/` 보다 앞서는 숨김 폴더(`.moai-backup` 약 1.8 만 항목, `.moai-backups` 약 2.2 만, `.vercel` 약 10 만)에서 예산이 다 떨어져 `apps/`·`packages/` 는 읽히지도 못해요. 프로필 적용이 `detected.manifests` 를 55 개에서 1 개로 줄이는 것은 그 결과예요.

고치는 것: (1) 얕은 폴더를 전부 읽은 다음에 한 단계 더 깊은 폴더로 가요(예산이 모자라도 얕은 manifest 가 먼저 나와요), (2) 숨김 폴더(이름이 `.` 으로 시작)는 도구·캐시·백업 상태라서 읽지 않아요(같은 저장소에서 2085 항목으로 54 개를 찾아요), (3) 결과에 목록이 부분적인지(예산이 바닥났는지, 상한 64 개를 넘어 뺀 것이 있는지)를 알려요. 프로필 계획이 기록된 manifest 를 줄일 때 보여 주는 일은 다음 조각이 맡아요.

## Non-goals

- 프로필 계획(`project-plan`)이 줄어드는 manifest 를 보여 주는 일(다음 조각).
- 코드 파일 언어 탐색(`code_languages`)의 순서와 예산을 바꾸지 않아요.
- 깊이 한도(`MAX_DEPTH` = 3), 항목 예산(`MAX_ENTRIES` = 5000), 목록 상한(`MAX_ITEMS` = 64)을 바꾸지 않아요.
- 중첩 manifest 의 내용 읽기, 심볼릭 링크 따라가기, 프로세스 시작은 하지 않아요.
- `skills/**` 문구는 바꾸지 않아요.

## Requirements

- REQ-1: The manifest walk shall read every folder at one depth before it reads any folder one level deeper.
- REQ-2: The manifest walk shall not enter a folder whose name starts with a dot.
- REQ-3: When `project-detect` reports the facts, the command shall report beside `detected` a `manifest_walk` object holding `complete` (false when the walk stopped at `MAX_ENTRIES`) and `over_cap` (the number of manifests found that `detected.manifests` leaves out because of `MAX_ITEMS`).
- REQ-4 (REQ-3): IF `manifest_walk.complete` is false or `manifest_walk.over_cap` is more than zero, THEN `next_actions` shall carry a line saying the manifest list is partial and must not be recorded over an earlier list without showing the user what drops.

## Acceptance

- AC-1 (REQ-1): 알파벳 순으로 앞서는 큰 폴더(깊이 3 에 파일 6000 개)가 있어도 뒤따르는 `apps/x/package.json` 이 나열되고 `manifest_walk.complete` 는 false 예요.
  verify: `cargo test --locked --offline --test project_detect_walk_cli a_large_early_folder_does_not_hide_a_later_packages_manifest`
- AC-2 (REQ-2): `.cache/pkg/package.json`, `.vercel/x/package.json`, `.moai-backup/y/go.mod` 는 나열되지 않고 `apps/x/package.json` 은 나열돼요.
  verify: `cargo test --locked --offline --test project_detect_walk_cli hidden_folders_are_not_walked_for_manifests`
- AC-3 (REQ-3): 작은 프로젝트는 `manifest_walk` 가 `{"complete": true, "over_cap": 0}` 이고, 70 개 manifest 는 64 개가 나열되고 `over_cap` 은 6 이에요.
  verify: `cargo test --locked --offline --test project_detect_walk_cli the_manifest_walk_fact_reports_completeness_and_the_cap`
- AC-4 (REQ-4): 부분 목록이면 `next_actions` 에 그 한 줄이 있고, 완전한 목록이면 없어요.
  verify: `cargo test --locked --offline --test project_detect_walk_cli a_partial_list_adds_one_next_action_and_a_complete_one_adds_none`
- AC-5 (REQ-1, REQ-2): 기존 중첩 manifest 시험(깊이 3 은 찾고 4 는 안 찾음, 무시 목록, 링크, 상한, 도구 사실)이 그대로 통과해요.
  verify: `cargo test --locked --offline --test project_detect_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-detect-walk-breadth/spec.md`, `src/project_detect.rs`(`nested_manifests` 와 `walk_manifests`, `detect_cli`, `detect` 의 호출부), `src/ops/project_detect.rs`(설명 글), `tests/project_detect_walk_cli.rs`(새 파일)
- read-only: `tests/project_detect_cli.rs`, `skills/moai/workflows/project.md`
- forbidden: `code_languages`, 한도 상수, `skills/**`, 다른 카드가 쥔 파일

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: `tests/project_detect_walk_cli.rs` 에 AC-1~4 시험을 먼저 써서 실패하는 것을 봐요.
3. GREEN: `walk_manifests` 를 폴더 단위 줄(queue)로 바꿔 한 깊이씩 읽고, 숨김 폴더를 건너뛰고, 예산 소진을 `complete` 로 돌려줘요. `detect` 가 `over_cap` 을 세어 `detect_cli` 가 `manifest_walk` 와 안내 줄을 내요.
4. 확인: 새 시험 파일, `project_detect_cli` 전체, `--lib project_detect`, 이 출력을 읽는 이웃 시험, fmt, clippy, 변이, `moai check affected`.

## Risks

- 숨김 폴더 안의 진짜 패키지(`.devcontainer/package.json` 등)는 더 이상 나열되지 않아요: 도구·캐시·백업 폴더가 대부분이고, 예산을 먼저 써 버려 실제 패키지를 가리는 쪽이 더 흔한 피해예요. 루트의 manifest 는 숨김 폴더 규칙과 상관없이 그대로 읽어요.
- 예산은 여전히 항목 5000 개예요: 숨김이 아닌 큰 폴더(`reports/` 등)가 깊이 3 에 있으면 `complete` 가 false 로 나와요. 조용히 줄어드는 대신 알려요(REQ-3·REQ-4).
- `detect_cli` 출력에 키 하나가 더해져요: 이 출력을 문자 그대로 비교하는 이웃 시험이 있으면 같이 고쳐요(범위 안에서 찾아요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 숨김 폴더를 건너뛴다(`SKIP_DIRS` 이름 목록을 늘리지 않는다) | 보고된 저장소의 큰 폴더가 모두 숨김 폴더였고, 이름 목록은 새 도구마다 모자라요 | 건너뛰기 빼기 |
| D-2 | 한도 상수는 그대로 두고 알림을 더한다 | 한도를 올리면 큰 저장소가 느려져요. 알려 주면 사용자와 다음 조각이 결정해요 | 해당 없음 |
| D-3 | 알림은 `detected` 안이 아니라 옆의 `manifest_walk` 에 둔다 | `detected` 는 프로필에 기록되고 `deny_unknown_fields` 라서 값이 늘면 기록 모양이 바뀌어요 | 키 빼기 |

## Result

(구현 뒤에 채워요.)
