# 부록 2 — 경로 표 (DR-07 #2143)

출처: synthesis S3(경로 이름 4종) + RC1 소스 상수. **RC1 출하** 열은 tip `604792666` 소스에서 확인(`src/lib.rs`·`src/claim.rs`·`src/install.rs`·`src/update.rs`), **설계 목표(P09)** 열은 상위 §13·[43fa] §20·[83f] §1이 그린 목표 경로다. P09가 미결인 이상 두 열은 같아지지 않는다.

| 대상 | RC1 출하 (tip `604792666` 실측) | 설계 목표 (P09 미결) |
| --- | --- | --- |
| 홈 설치 receipt | `~/.moai-cli/installations/<install_id>.json` | [83f] `<state_root>/installations/`; 상위 §13 `installs/<host>/<version>/` |
| 홈 호스트 receipt·known-good | `~/.moai-cli/installations/<install_id>/hosts/moai/<host>.json`, `…/hosts/moai/<host>/known-good/<hex>` (#2286: skill 폴더 이름 `moai`를 따름) | [83f] `<state_root>/skills/moai/<version>/`(버전별 디렉터리 모델) |
| 홈 바이너리 | `~/.moai-cli/bin/moai-cli`(결정 A, 미출하 — 04-A) | 상위 §13 `installs/<host>/<version>/` 하위 |
| 프로젝트 상태 | `<project>/.moai-cli/{project.yaml, manifests/, receipts/, forge/, decisions/, evidence/, worktrees/}` | [487]·[43fa] `projects/<id>/{registry.json, contracts, artifacts, indexes, runs, retention}` |
| claim journal | `<common dir>/.moai-cli/claims/` (`src/claim.rs` journal_root — clone 공유) | 상위 §13 `workspaces/<id>/journal.jsonl` |
| 상태 어휘 스키마 | `schemas/protocol/v1/` | 상위 §6 schema (policy_digest 단일 필드) |

## 용어 정의 (synthesis S3 — 4종, 한 절에서 정의)

- **`<state_root>`**: [83f]이 가정하는 홈 상태 루트. RC1 출하물에는 이 이름의 디렉터리가 없다 — `~/.moai-cli/`이 실물이다.
- **`<moai_home>`**: Core Brain이 그리는 홈 루트 별칭. RC1 출하물과 동일한 실체는 없다.
- **`~/.moai-cli`**: RC1이 실제로 만드는 홈 상태 루트(04-A 결정 경로 A). `~/.moai`(Go 세대)와 점 하나로 구분되며 **Go에도 RC1에도 없는 `~/moai/`(점 없음)과 혼동 금지**.
- **`workspace_root`**: 상위 §13 `workspaces/<id>/` 트리 루트. RC1은 worktree를 `git worktree` + `.moai-cli/worktrees/`(레인 규칙)로 관리하며 이 이름의 루트는 없다.

규칙(synthesis §4-1): 경로 정의는 상위 §13 한 곳에서 두 열로만 하고, 나머지 문서는 이 부록을 참조만 한다.
