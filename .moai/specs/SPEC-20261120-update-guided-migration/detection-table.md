# 감지 통합 표: `moai update` 가 지금 감지하고 바꾸는 것 전부

기준 커밋: 로컬 `v2026.11.0` 의 `6a8fa9cacf4634f772b3c13dc628d0c1e4ec7450`. 코드는 아직 바뀌지 않았어요. 이 표는 읽기 전용 조사 에이전트 3개(최상위 흐름, HOME 쪽, 프로젝트 쪽)가 파일:줄 인용으로 낸 결과를 합친 거예요. 원문은 저장소 밖 리포트 폴더(`~/.moai/reports/moai-adk/c201/`)에 있어요.

## 증거의 등급

- **직접 다시 읽음(V1~V7)**:
  - V1 `moai update` 는 `~/.moai-cli → ~/.moai` 이동을 부르지 않아요. `home_install` 은 `update.rs:1001`, `:1070` 에서 이전 설치 id 를 읽을 뿐이에요.
  - V2 프로젝트 상태 폴더 상수 `PROJECT_STATE_DIR` 는 아직 `.moai-cli` 예요(`lib.rs:216`, `:239`). `.moai` 로의 이동은 `go-state-plan` 의 `legacy_into_new` 안에만 있어요(`go_state.rs`, 3곳).
  - V3 `hooks::HOSTS` 는 `["claude","codex","omp"]` 예요(`hooks.rs:88`). OMP 제거 카드는 아직 팁에 없어요.
  - V4 `moai update` 는 프로젝트 파일을 계획하지 않아요(`update_tux.rs:41`).
  - V5 옵션은 `--confirm`, `--confirm-cleanup`, `--confirm-home`, `--confirm-switch`, `--revert-switch`, `--allow-older`, `--json` 이에요(`main.rs:922-951`).
  - V6 확인 실행의 순서는 skills → agents → routing → style → bin_names → hooks → keys → cleanup → home → switch 예요(`update.rs:2028-2135`).
  - V7 빌드한 바이너리의 `moai update --help` 가 같은 옵션을 보여 줘요.
- **조사 에이전트가 읽은 그대로 옮김(미검증)**: 그 밖의 모든 줄 번호와 동작 설명. 설계 결정에 걸리는 곳(아래 "빈틈" G1, G2, G6)은 SLICE 2 의 RED 테스트가 먼저 확인해요.
- 돌려 본 것은 없어요. 코드와 테스트 코드를 읽었을 뿐이에요.

## 한눈에: 세 개의 따로 노는 세계

1. **HOME 층의 `moai update`**: 미리보기 한 번 → 질문 하나(`[y/N]`) → 일곱 가지(skills, agents, routing, style, bin_names, hooks, keys)가 묶인 한 `--confirm`. cleanup, Go HOME 정리(home retire), 이름 전환(switch)은 각자 토큰이 있어서 따로, 한 번에 하나씩 새 미리보기를 보고 돌려요. 모델 질문은 적용 뒤 stderr 로 나와요.
2. **HOME 이동 `~/.moai-cli → ~/.moai`**: 코드는 있지만(`home-install-plan/apply/revert`) `moai update` 가 부르지 않아요. 사람이 따로 돌려야 해요.
3. **프로젝트 층**: `moai project-update-plan <project>` 한 번(읽기 전용)이 모든 단계의 미리보기를 주고, 적용은 단계마다 인쇄된 명령을 `/moai update` 세션이 하나씩 돌려요(단계마다 새 미리보기). 프로젝트용 `update` 적용 명령은 없어요.

## 통합 표

열: **묶음**은 새 미리보기의 분류(바뀜 / 그대로 / 선택 필요), **보호**는 되돌리기 어렵거나 HOME·금고를 건드려 질문에 이름을 대야 하는 것(guarded), **순서**는 아래 "적용 순서"의 번호예요.

### HOME 층 (셸 `moai update` 의 영역)

| 순서 | 항목 | 지금의 감지(줄) | 묶음 | 보호 | 지금의 적용 | 지금의 되돌리기 |
|---|---|---|---|---|---|---|
| H1 | Go 세대 `~/.moai` 정리: 최상위 항목을 `~/.moai-go-retired/<UTC>/` 로 이름 바꿔 옮김. 삭제 없음 | `~/.moai` 가 go/mixed/rust-first 이고 옮길 것이 있을 때(`update_home.rs:49-62`) | 바뀜 | 예(HOME) | `moai update --confirm-home home:<16hex>` 만(터미널 y 나 `--confirm` 으로는 안 돼요, `update.rs:1698-1711`) | `home-retire-revert --home <H> --stamp <S> --permission-reference <ref>`. 옮긴 항목이 그 뒤 바뀌었으면 그대로 두고 `revert_partial` |
| H2 | HOME 폴더 이동 `~/.moai-cli → ~/.moai`: 항목별 이름 바꾸기, 버전 폴더와 `path/` 링크 마지막, 끝에 마커(`moai.json` 스키마 2). 남는 것: `projects`, `worktrees`, `bin`, 표에 없는 이름, 링크 | 이전 폴더가 있고 상태가 none/empty/go/mixed(또는 끝난 retire 기록이 있는 unknown)(`home_install.rs:385-395`) | 바뀜 | 예(HOME·금고) | **`moai update` 는 부르지 않음(V1).** `home-install-apply --home --now --plan-digest --permission-reference [--project]`. 토큰 플래그 없음. 이 클론의 레지스트리에 살아 있는 세션이 있거나 읽을 수 없으면 아무것도 쓰지 않고 중단(`home_install_apply.rs:506-522`) | `home-install-revert --home [--stamp] --permission-reference`. 마커 해시가 다르면 거절 |
| H3 | 새 HOME 만들기: `~/.moai-cli` 도 `~/.moai` 도 없으면 `~/.moai` 와 마커(스키마 1)를 만듦 | 이전 폴더 없음 + 판정 none(`home_root.rs:37-40`) | 바뀜 | 아니오 | `update.rs:2361`, 스킬 복사 직전에 자동. 토큰 없음 | 없음(기존 상태가 있으면 아무것도 안 써요) |
| H4 | 스킬 복사본 `~/.claude/skills/moai`, `~/.agents/skills/moai`(Codex): 파일마다 만듦/갱신/그대로/편집본 보존/충돌 | 3자 비교(`apply.rs:47-55`). 막힘: 링크·쓰기 불가, pending 영수증, 설치본이 더 새 버전, 영수증이 소유하지 않는 기존 `moai` 폴더 | 바뀜(충돌은 선택 필요) | 아니오 | `moai update --confirm <digest>`(일곱 가지 묶음) | **명령 없음**. 결과 문장만 `recovery-plan / recovery-apply for a host left pending`. 덮어쓰기 전 원본은 `known-good` 에 보관(수동 복구용) |
| H5 | 베이스 에이전트 `moai-<role>.<ext>`: claude `~/.claude/agents`, codex `~/.codex/agents`(omp `~/.omp/agent/agents` 는 OMP 카드 전까지) | 스킬이 막히지 않은 호스트만(`update.rs:1583-1589`). 편집본은 덮어쓰지 않음 | 바뀜 | 아니오 | 같은 `--confirm` | `moai agents-home-revert <HOME> --receipt <id> --permission-reference <ref>` |
| H6 | routing 블록 `moai-routing`: `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md` | 패키지에 `references/home/routing.md` 있고 스킬이 막히지 않음. 편집한 블록은 conflict | 바뀜(conflict 는 선택 필요) | 아니오 | 같은 `--confirm` | `moai home-routing-revert <HOME> --receipt <id> --permission-reference <ref>` |
| H7 | style 블록: `~/.claude/output-styles/MoAI-Easy.md` 와 `settings.json` 의 `outputStyle`(키가 없을 때만), Codex `AGENTS.md` 의 `moai-output-style` 블록 | 패키지에 `references/home/output-style.md` 있음. 다른 값은 conflict | 바뀜(conflict 는 선택 필요) | 아니오 | 같은 `--confirm` | `moai home-style-revert <HOME> --receipt <id> --permission-reference <ref>` |
| H8 | 명령 이름 `moai`: `bin/moai` 를 `bin/moai-cli` 의 하드 링크로(같은 빌드 별도 복사본은 `moai.replaced-<UTC>`) | `<state>/versions/<v>/bin/` 안에서 실행 중일 때만(`bin_names.rs:91-187`) | 바뀜 | 아니오 | 같은 `--confirm` | **명령 없음**(`result.revert` 에도 `REVERT_ORDER` 에도 없음) |
| H9 | 현재 프로젝트 훅 갱신: 이미 MoAI 훅이 설치된 호스트만 `.claude/settings.json`, `.codex/hooks.json`, `.omp/extensions/moai-hooks.ts`. 새로 설치하지 않음 | cwd 가 git 작업 트리이고 홈이 아님(`update_hooks.rs:129-173`) | 바뀜(사용자가 바꾼 항목은 conflict) | 아니오 | 같은 `--confirm`(digest 가 프로젝트 경로를 묶어요) | `moai hooks-revert <PROJECT> --receipt <id> --permission-reference <ref>` |
| H10 | glm 키 가져오기: `~/.moai/.env.glm`·`~/.moai-cli/.env.glm` → 금고 `vault/_shared/glm.key`. 원본 유지, 값 비출력 | 파일이 있고 금고에 항목이 없을 때. 링크·FIFO·64KiB 초과·0600 보다 느슨하면 그대로 둠 | 바뀜(그대로 두는 경우는 그대로) | 예(금고) | 같은 `--confirm` | `moai keys-revert <HOME> --receipt keys-… --permission-reference <ref>` |
| H11 | Go 잔재 보관: `<state>/cleanup/<stamp>/` 로 이동(삭제 없음) | `update_cleanup::plan` 에 확인 토큰이 있을 때(계획 오류는 버려져요, `update.rs:1644-1646`) | 바뀜 | 예(되돌리기 어려움 아님, HOME 항목) | `--confirm-cleanup <token>` 만. `--confirm-home` 과 한 번에 못 씀(`update.rs:1951-1955`) | `cleanup-revert-plan <HOME> --stamp <S>` 후 `cleanup-revert` |
| H12 | 이름 전환 스크립트 `<state>/path/moai`: 알려진 v1 을 v2 로 교체(옛 파일은 `moai.before-<UTC>`) | 알려진 v1 해시일 때만. 모르는 파일은 그대로(`update/switch.rs:155-196`) | 바뀜 | 아니오 | `--confirm-switch switch:<16hex>`. 터미널 y 로는 안 돼요 | `moai-cli update --revert-switch <STAMP>`(단독) |
| H13 | 모델·강도 저장: `<state>/config/launcher.yaml` | 적용 뒤, 사람 터미널일 때 질문(`update.rs:2247-2257`) | 선택 필요 | 아니오 | 질문 뒤 `Save these model settings? [y/N]` | `moai models-revert <HOME> --receipt <models-id> --permission-reference <ref>`. 결과의 `revert` 에는 없음 |
| — | 자기 갱신(바이너리 교체) | 호출처 없음(`self_update.rs:15-16`) | 그대로 | — | 연결 안 됨 | — |

### 프로젝트 층 (`/moai update` 가 돌리는 `moai project-update-plan` 의 영역)

phase 는 backup → install → merge 예요. 한 phase 의 상태는 가장 나쁜 step 이에요(unchanged < skip < ready < needs_decision < conflict < blocked).

| 순서 | 항목(step) | 지금의 감지(줄) | 묶음 | 보호 | 지금의 적용 | 지금의 되돌리기 |
|---|---|---|---|---|---|---|
| P1 | `go-settings`: `.claude/settings.json` 의 Go 도구가 쓴 `statusLine`·`subagentStatusLine`·`plansDirectory` 정리(한 글자라도 다르면 사용자 값이라 그대로) | 상태 go 일 때만(`go_settings_clean.rs:18-22`) | 바뀜(held 는 선택 필요) | 아니오 | `go-settings-apply --plan-digest <d> --permission-reference <r> [project]` | `go-settings-revert --receipt go-settings-… --permission-reference <r> [project]`. 파일이 apply 직후 상태일 때만 |
| P2 | `go-state`: ① Go `.moai/` → `.moai-backup`(이름이 차 있으면 `-<UTC>`) + 새 `.moai/`, 또는 ② 프로젝트 `.moai-cli` → `.moai`(`legacy_into_new`, 18개 항목만; `worktrees`·`reports`·`backups` 는 남김), 또는 ③ Go 폴더는 두고 증명된 `.claude/` 파일만 이동. 하위: `state_moves`, `restores`(사용자 데이터 복사), `root_files`(`CLAUDE.md`, `AGENTS.md`, `AGENTS.local.md` 백업으로), `counterparts`·`after_move`(읽기 전용 설명) | `state_root::classify` 가 go/none 일 때만. rust/mixed/unknown/refuse/interrupted 는 아무것도 안 움직임(`go_state.rs:2710-2769`). 링크 항목이 하나라도 있으면 전부 막힘 | 바뀜(blocked/backup_ignore 는 선택 필요) | **예**(프로젝트 상태 폴더 이동) | `go-state-apply <project> --plan-digest <d> --permission-reference <r>`. 항상 needs_decision. 영수증 pending → committed. 마커 마지막 | `go-state-revert-plan <project> --backup <UTC>` 후 `go-state-revert <project> --backup <UTC> --plan-digest <d> --permission-reference <r>`. 하나라도 자리에 다시 생겼으면 전체 거절. 생성 step 의 `AGENTS.md` 가 있으면 거절 |
| P3 | `profile`: `.moai/project.yaml` 등 이 세대의 프로필·관리 파일 | 프로필 delta | 바뀜(conflict 는 선택 필요) | 아니오 | `project-apply <p> --plan-digest <d> …`(stdin 은 계획) | apply 결과가 인쇄 |
| P4 | `generation`: 프로젝트 에이전트·스킬 사본, `AGENTS.md` 관리 블록(`moai-workflow`). 복원 못 하는 에이전트·스킬은 계획하지 않고 `/moai project` 로 | 관리 블록이 없으면 conflict. 선택지 `append-managed`(추천) 등 | 바뀜 / 선택 필요 | 아니오 | `generation-apply <p> --plan-digest <d> --permission-reference <r>`(충돌 선택지마다 따로 인쇄) | `apply-recover --receipt <agents_receipt>` 후 `--receipt <generation receipt>` |
| P5 | `claude-settings`: MoAI 가 소유한 키(`hooks` 제외)만. 사용자가 둔 키는 conflict, 교체는 `--confirm-replace <token>` | 소유 키 delta | 바뀜 / 선택 필요 | 아니오 | `claude-settings-apply … --plan-digest <d>`(stdin 은 소유 키) | 인쇄된 `cp` 명령(원본은 `settings-backups/`) |
| P6 | `hooks`: ① Go 훅 → Rust 훅(`go-hooks-apply`: 스크립트는 영수증으로 이동, `hooks` 키 교체), ② Go 설정이 백업으로 간 뒤엔 `hooks-apply --host claude` | 훅 판정은 `.claude/hooks/moai/<이름>` 한 토큰. 다른 도구 훅이 섞이면 blocked. PATH 의 첫 `moai` 가 Go 판이면 쓰기 전에 거절(`go_hooks.rs:477-493`) | 바뀜 / 선택 필요 | 아니오 | `go-hooks-apply --plan-digest <d> --permission-reference <r> [project]`, 또는 `hooks-apply <p> --host claude …` | `go-hooks-revert --receipt go-hooks-… --permission-reference <r> [project]`, 또는 `hooks-revert <p> --receipt <id> …`. 첫 설치는 되돌릴 것 없음 |
| P7 | `github` 템플릿 | 옵트인한 프로젝트만. 없으면 skip | 바뀜 / 그대로 | 아니오 | `github-apply <p> --plan-digest <d> --permission-reference <r>` | apply 결과가 인쇄 |
| P8 | `codex-hooks`: `.codex/hooks.json` | 프로필 hosts 에 codex. 사용자가 편집한 파일은 conflict(`keep` 이 기본) | 바뀜 / 선택 필요 | 아니오 | `hooks-apply <p> --host codex …` | `hooks-revert` 가 있는데도 `undo_order` 는 "no revert command" 라고 적음(모순) |
| P9 | `codex-config`: `.codex/config.toml` 의 `tui.status_line = "moai"` | 충돌이면 `keep`/`diff`/`replace` | 바뀜 / 선택 필요 | 아니오 | `codex-config-apply <p> --plan-digest <d> …` | **명령 없음** |
| P10 | `codex-skill`: `~/.agents/skills/moai` 가 없을 때만 프로젝트 `.agents/skills/moai` | HOME 에 있으면 skip. HOME 없으면 needs_decision | 바뀜 / 그대로 | 아니오 | `project-skill-apply <p> --plan-digest <d> …` | **명령 없음**(`project_skill_revert` ops 는 있으나 `undo_order` 에 없어요) |
| P11 | `agents-local-merge`: `AGENTS.local.md` 본문을 `AGENTS.md` 관리 블록 밖으로(커밋되는 파일에 들어간다는 안내 필요, 비밀 의심 줄은 번호만 보이고 막힘) | merge 후보가 있을 때. install 뒤에만 | **선택 필요** | 아니오 | `agents-merge-apply <p> --plan-digest <d> --permission-reference <r>` | `agents-merge-revert <project> --receipt <id> --permission-reference <ref>` |
| P12 | `claude-settings-merge`: Go 때 사용자가 둔 `settings.json` 값을 키별로 되돌려 줌(`mergeable`/`kept_in_backup`/`go_generated`/`blocked`) | 백업에서 읽음 | **선택 필요**(키별 고르기) | 아니오 | `claude-settings-apply … --plan-digest <d>`(고른 키만 stdin) | 인쇄된 `cp` 명령 |
| P13 | OMP 호스트: `omp-hooks`(`.omp/extensions/moai-hooks.ts`), `omp-agents`(`.omp/agents/moai-<role>.md`), `omp-config`(항상 skip) | 팁에서는 살아 있는 호스트(V3). 제거 카드가 들어가면 `omp-hooks` 는 skip + 알림 한 줄 | 제거 카드 전: 바뀜. 후: **그대로**(더는 관리 안 함으로 목록만) | 아니오 | `hooks-apply --host omp`, `agents-project-apply …` | `undo_order` 에 없음. **"더는 관리하지 않는 파일" 목록을 만드는 코드는 없어요**(`leftovers()` 는 `.moai-cli/`, `CLAUDE.md`, `.moai/.update.lock` 셋뿐) |

## 적용 순서와 이유

의존이 있는 곳만 고정하고, 나머지는 기존 코드 순서를 지켜요.

1. **H1 Go HOME 정리 → H2 HOME 이동**: Go 항목이 남아 있으면 이동이 `go_items_not_retired` 로 막히고, 정리 뒤 `~/.moai` 는 `unknown` 이라 끝난 정리 기록이 있어야 이동이 풀려요(`home_install.rs:480-495`).
2. **HOME 이동 직후 다시 조사**: 영수증과 상태 폴더 경로가 `~/.moai-cli` 에서 `~/.moai` 로 바뀌어서, 이동 전에 계산한 digest 와 경로는 낡았어요. 이동을 뒤로 미루면 갓 쓴 영수증까지 옮겨지고요.
3. **H3~H12(기존 순서)**: H3 새 HOME → 스킬 → 에이전트 → routing → style → 명령 이름 → 현재 프로젝트 훅 → 키 → cleanup → 이름 전환. cleanup 과 H1 은 토큰이 따로라 새 미리보기를 사이에 두고 돌려요.
4. **프로젝트 층(프로젝트 안일 때)**: P1 → P2 → (다시 조사) → P3 → P4 → P5 → (다시 조사) → P6 → P7~P10 → P11 → P12. 이유: go-settings 에 명령이 있으면 go-state 가 `re-run after the go-settings apply`, backup 에 할 일이 있으면 install 전부가 `re-run after the backup phase`, profile 이 ready 면 generation 이, claude-settings 에 명령이 있으면 hooks 가, generation 이 ready 면 merge 가 기다려요.
5. **프로젝트 훅(P6)은 HOME 의 이름 전환(H12) 뒤에**: Go 훅 교체는 PATH 의 첫 `moai` 가 Go 판이면 거절해요.
6. **프로젝트의 `codex-skill`(P10)은 HOME 스킬(H4) 뒤에**: HOME 에 Codex 스킬이 있으면 프로젝트 사본은 필요 없어요.
7. **H13 모델 질문**: 마지막(지금처럼). 저장 결과는 요약에 넣어요(지금의 마지막 화면은 모델 결과를 안 보여 줘요).

되돌리기 순서는 적용의 거꾸로예요: P12 → … → P1 → H12 → … → H2 → H1. 단계마다 새 권한 참조가 필요해요. 되돌리기 명령이 없는 단계는 요약에 "되돌리는 명령 없음"이라고 말해요(H4 스킬 사본, H8 명령 이름, P9, P10 그리고 상태에 따라 P13).

## 감지 통합 결과의 모양(제안)

읽기 전용 감지가 아래 모양을 줘요. 새 쓰기 코드는 없고, 기존 계획 함수들을 한 번에 불러 모아요.

```
survey: {
  survey_digest,                 // 아래 components 의 digest 를 묶음. 하나라도 바뀌면 달라져요
  components: [ { id, digest } ],// 기존 계획 하나당 하나(user layer, home retire, home install, cleanup, switch, project-update-plan)
  groups: { will_change: [item], kept: [item], needs_pick: [item] },
  order: [item id],              // 위 "적용 순서"
  nothing_to_apply: bool
}
item: { id, layer: home|project, title(쉬운 말), count{files,folders}, guarded: bool,
        depends_on: [id], blocked_by?: reason,
        apply: { part, command? }, revert: { command } | { none: reason } }
```

HOME 이동의 크기는 지금 개수와 모드만 있고 바이트는 없어요(`home_install.rs:514-523`). "크기"는 항목·파일 개수로 줄여요.

## 빈틈(설계가 메워야 하는 곳)

- **G1 HOME 이동이 `moai update` 밖이에요(V1).** 그리고 `home-install-apply` 는 이 클론의 레지스트리에 살아 있는 세션이 있으면 중단해요. `/moai update` 를 돌리는 세션 자신이 등록된 세션이면 자기 이동을 스스로 막을 수 있어요. 어느 쪽에서 돌릴지 정해야 해요(리더 결정 요청 1).
- **G2 프로젝트 층에는 `update` 적용 명령이 없어요.** 단계마다 인쇄된 명령을 세션이 돌려요. 통합 적용은 이 루프를 안에서 돌려야 하고(단계마다 새 미리보기, 첫 실패에서 멈춤), 셸 `moai update` 가 프로젝트까지 다룰지, 스킬만 다룰지 정해야 해요(리더 결정 요청 2).
- **G3 한 번에 안 되는 조합**: `--confirm-home` 과 `--confirm-cleanup` 동시 불가, `--confirm-switch` 와 `--revert-switch` 동시 불가. 한 질문 뒤 여러 번의 실행으로 풀어요.
- **G4 되돌리기 명령이 없는 것**: H4(스킬 사본은 pending 호스트 복구뿐), H8, P9, P10. `undo_order` 가 `codex-hooks`·`omp-hooks` 를 "no revert command" 라고 적지만 `hooks-revert` 가 있어요(모순).
- **G5 진행 표시가 없어요.** 지금 `moai update` 흐름에는 스피너도 단계 줄도 없어요(`update_tux.rs`). 단계별 `n/N` 줄과 done/skipped/failed 는 새로 만들어야 해요.
- **G6 OMP 잔재 목록을 만드는 코드가 없어요**(P13). 제거 카드 뒤에는 `.omp/…` 에 이전 버전이 쓴 파일이 남는데, 그것을 "더는 관리하지 않음"으로 보여 주는 감지가 새로 필요해요. 지우지는 않아요.
- **G7 `~/.moai-cli/projects` 는 이동 때 남아요**(`home_install.rs:74-76`). `layout::default_home_root` 는 새 루트를 따르므로, 이동 뒤 `layout.json` 이 있던 프로젝트가 조용히 프로젝트 쪽 저장소로 돌아갈 수 있어 보여요(코드만 읽은 추정, 테스트 없음). SLICE 2 의 RED 테스트가 먼저 확인해요.
- **G8 `update.rs`·`home_routing.rs`·`home_style.rs` 의 주석이 낡았어요**(`~/.moai-cli` 라고 적혀 있지만 실제 경로는 마커에 따라 달라져요). 새 흐름의 안내문은 코드가 계산한 경로를 써야 해요.
- **G9 적용 결과의 `"writes": false`**(쓴 뒤에도), 단독 cleanup/home/switch 적용이 `mode:"preview"`·`applied:false`, `/reload-skills` 와 `/reload` 두 가지 안내, 모델 도구 3개(질문)와 4개(보고서)의 차이. 새 요약은 하나의 기준으로 말해요.
- **G10 모델 결과와 `models-revert` 가 마지막 화면에 안 나와요.** 질문 줄밖에 저장 여부를 알 길이 없어요.
- **G11 `workflows/update.md`·`SKILL.md` 는 다른 카드(프로젝트·update 경계)가 다시 쓰는 중이에요.** 이 카드의 스킬 문장 조각은 그 카드가 들어간 뒤에 해요. 테스트(`update_route`, `mode_workflow_steps`)가 문장을 고정해서 같은 커밋에서 같이 바꿔요.
- **G12 선행 카드와 겹침**: 프로젝트 `.moai-cli → .moai` 마무리(C-175)와 HOME 합치기(C-199)는 아직 시작 전이에요. 이 표의 H2·P2 는 지금 있는 코드(`home-install-*`, `go-state` 의 `legacy_into_new`)를 기준으로 했고, 그 카드들이 같은 판별 코드(`state_root`)로 바꾸면 이 표의 "감지" 칸만 바뀌어요.
