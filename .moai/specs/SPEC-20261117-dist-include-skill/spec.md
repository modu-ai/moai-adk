---
id: SPEC-20261117-dist-include-skill
title: 릴리스 archive가 skill 패키지와 agent를 binary 옆에 실어요
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-installer-skill-agents
---

## Intent

`moai update`는 skill 패키지와 기본 agent를 binary 옆 버전 폴더에서 읽어요(`<version>/skills/moai`,
`<version>/agents`; `src/update.rs` `installed_package_dir`, `installed_agents_dir`). 설치 스크립트(SPEC-20261117-
installer-skill-agents)는 archive에서 binary 옆에 있는 `skills/`와 `agents/`를 버전 폴더로 복사해요. 그런데
`dist-workspace.toml`에는 `include`가 없어서, 지금 릴리스 archive에는 binary와 cargo-dist가 자동으로 넣는 README·
LICENSE뿐이에요. 그래서 새로 설치하면 `moai update`가 `no skill package installed with this binary`로 멈춰요(조사 F1).

cargo-dist 설정 문서(https://axodotdev.github.io/cargo-dist/book/reference/config.html, 2026-11-17T17:47Z에 읽음):
`include`는 "additional files or directories to copy into the root of all archives and installers"이고, "Paths are
relative to the config file. Globs are not supported."예요. package-local 설정이라 workspace `[dist]`에 둘 수 있어요.

## Non-goals

- 릴리스 workflow를 만들거나 dist target 목록을 바꾸는 일(지금 `aarch64-apple-darwin`만, 별도 사용자 선택).
- 설치 스크립트 쪽(SPEC-20261117-installer-skill-agents).

## Requirements

- REQ-1: The `[dist]` section of `dist-workspace.toml` shall list exactly `skills/` and `agents/` under `include`.
- REQ-2: Each `include` entry shall be a relative path without a glob that names a real folder of the repository.
- REQ-3: The `skills/` folder shall hold only the `moai` skill package with its `SKILL.md`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): verify: `python3 -m unittest -v tests/dist_include_test.py` → 4 tests OK.
- AC-2 (REQ-1): verify: `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py` → OK.

## Scope

- `dist-workspace.toml`, `tests/dist_include_test.py`, 이 SPEC.

## Plan

1. 시험 먼저: `include` 목록과 그 폴더들을 읽는 Python 시험.
2. `dist-workspace.toml`에 `include = ["skills/", "agents/"]`.

## Risks

- cargo-dist가 이 기계에 없어서 실제 archive는 만들어 보지 않았어요. archive 모양(`moai-cli-<triple>/` 아래 binary,
  `skills/moai`, `agents`)은 문서의 "root of all archives"에 기대고, 설치 스크립트 시험이 같은 모양의 archive로
  설치를 확인해요. 첫 릴리스 빌드에서 archive 목록을 확인해야 해요.
- `skills/` 아래의 모든 것이 실려요. 다른 skill 폴더를 더하면 REQ-3 시험이 막아요.
