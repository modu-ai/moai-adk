---
title: moai epic 에픽 진행률
weight: 100
draft: false
description: "여러 SPEC 으로 나뉜 에픽의 마일스톤 진행률을 디스크에서 계산하는 moai epic status 커맨드."
---

`moai epic status` 는 하나의 에픽이 지금 어디까지 왔는지를 **디스크에 있는 것만 읽어서** 계산합니다. 별도의 진행률 저장소를 두지 않고, 그때그때 SPEC 문서를 훑어 마일스톤 지도를 다시 만듭니다. 읽기 전용이라 어떤 파일도 고치지 않습니다.

에픽 하나가 SPEC 여러 개로 쪼개지면 "몇 개가 끝났는지"가 금세 헷갈립니다. 각 SPEC 을 열어 status 를 확인하는 대신, 이 커맨드가 한 줄로 답합니다.

## 개요

```bash
moai epic status <prefix> [OPTIONS]
```

`<prefix>` 는 필수 인자로, 에픽을 식별하는 SPEC-ID 접두사입니다. 예를 들어 `KANBAN` 을 주면 `.moai/specs/SPEC-KANBAN-*/spec.md` 를 대상으로 삼습니다.

## 무엇을 읽는가

세 가지를 읽어 마일스톤 지도를 만듭니다.

1. **SPEC frontmatter** — 각 `spec.md` 의 `status` 값. 마일스톤이 끝났는지 판단하는 근거입니다.
2. **제목의 마일스톤 표식** — SPEC 제목에 적힌 `(TOKEN Mx)` 형태의 표식. 어느 SPEC 이 어느 마일스톤인지 이어 줍니다.
3. **디자인 리포트(선택)** — 있으면 정규 마일스톤 목록의 출처로 씁니다. 자동 탐색하며 `--design-report` 로 직접 지정할 수 있습니다.

표식이 없는 SPEC 은 버려지지 않고 `untracked_specs` 로 따로 보고됩니다 — 조용히 빠지면 "그 SPEC 은 없다"로 읽히기 때문입니다.

## 플래그

| 플래그 | 설명 |
|--------|------|
| `--json` | 고정 형태의 JSON 을 stdout 으로 출력 |
| `--design-report <path>` | 디자인 리포트 자동 탐색을 대신할 경로 지정 |
| `--marker <token>` | 추론된 에픽 토큰을 직접 지정 (예: `BAS`) |
| `--base-dir <path>` | 프로젝트 루트 (기본값: 현재 작업 디렉터리) |

## 예시

기본 출력은 사람이 읽는 진행 보드입니다.

```bash
$ moai epic status KANBAN
🎯 KANBAN ▓▓▓▓▓░░░░░ 2/4 (50%)
Epic progress:   KANBAN
  🟢 M0 M0                            SPEC-KANBAN-RENAME-001 (completed)
  ⬜ M1 M1                             SPEC-KANBAN-BOOTSTRAP-001 (draft)
  ⬜ M2 M2                             SPEC-KANBAN-WORKTREE-001 (draft)
  🟢 M3 M3                            SPEC-KANBAN-BOARD-001 (completed)
```

표식이 하나도 없으면 그 사실을 그대로 적고, 대신 걸린 SPEC 을 나열합니다.

```bash
$ moai epic status DESIGN-DOCS
🎯 DESIGN-DOCS — 2 SPEC(s) matched, none carrying a (TOKEN Mx) milestone marker
Epic progress:   DESIGN-DOCS
untracked_specs: SPEC-DESIGN-DOCS-001, SPEC-DESIGN-DOCS-V31-001
```

`--json` 은 스크립트에서 쓰기 좋은 고정 형태를 냅니다.

```bash
$ moai epic status KANBAN --json
{
  "epic": "KANBAN",
  "epic_token": "KANBAN",
  "milestones": [
    {
      "id": "M0",
      "label": "M0",
      "status": "done",
      "covered": true,
      "spec_id": "SPEC-KANBAN-RENAME-001",
      "spec_status": "completed",
      "sync_commit_sha": "144573336d07da19f4b8a50aa26c38db2704afb5"
    }
  ],
  "done": 2,
  "total": 4,
  "pct": 50,
  "extra_mx": [],
  "untracked_specs": ["SPEC-KANBAN-TODO-CLI-001"],
  "baseline_attribution": "3b9b3bf9959669c4bfc43da313e25bca61f910a2"
}
```

## JSON 필드

| 필드 | 설명 |
|------|------|
| `epic` | 인자로 준 접두사 |
| `epic_token` | 제목 표식에서 찾아낸 에픽 토큰. 찾지 못하면 빈 문자열 |
| `milestones` | 마일스톤 배열. 각 항목은 id · 라벨 · 상태 · 담당 SPEC · 그 SPEC 의 status · sync 커밋 SHA |
| `done` / `total` / `pct` | 완료 수, 전체 수, 백분율 |
| `extra_mx` | 정규 목록에 없는데 표식만 있는 마일스톤 |
| `untracked_specs` | 접두사에는 걸렸지만 마일스톤 표식이 없는 SPEC |
| `baseline_attribution` | 계산 시점의 git 커밋 SHA |

`baseline_attribution` 이 함께 나오는 이유는 이 값이 **어느 시점의 트리를 읽은 결과인지** 를 남겨 두기 위해서입니다. 진행률만 적어 두면 언제 잰 값인지 알 수 없습니다.

## 읽기 전용

이 커맨드는 관측만 합니다. SPEC 의 status 를 바꾸지 않고, 진행률을 어디에도 저장하지 않으며, 실행할 때마다 디스크에서 다시 계산합니다. 그래서 결과가 실제 파일과 어긋날 일이 없습니다.

계산 결과는 웹 콘솔의 모니터 영역에도 에픽 패널로 나타납니다. [MoAI Web Console](/ko/advanced/moai-web-console/) 을 참조하세요.

---

관련: [moai spec 문서 관리](/ko/cli-reference/spec/) · [MoAI Web Console](/ko/advanced/moai-web-console/)
