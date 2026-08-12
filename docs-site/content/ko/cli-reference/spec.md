---
title: moai spec 문서 관리
weight: 35
draft: false
---

`moai spec` 은 `.moai/specs/` 디렉터리의 SPEC 문서를 관리합니다. 상태 갱신, 드리프트 감지, 수용 기준 조회, EARS/GEARS 린트, 원자적 종료, era 감사, 아카이빙 하위 명령어를 제공합니다.

SPEC 은 하네스(harness) 가 작업을 위임받는 단위이자, 관리자 에이전트가 진행 상황을 추적하는 단일 레코드이기 때문에, 상태가 프론트매터와 실제 git 히스토리 사이에서 어긋나면 전체 워크플로우의 신호가 깨집니다. 그래서 이 커맨드는 드리프트 감지·era 감사·원자적 종료 같은 기계적 검사를 제공해, SPEC 라이프사이클 전반에 걸쳐 일관성을 보존합니다. 따라서 사람이 손으로 프론트매터를 고치는 대신 이 커맨드를 경유하는 것이 권장됩니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai spec status` | SPEC 상태 갱신 또는 나열 |
| `moai spec drift` | frontmatter status와 git log 간 드리프트 감지 |
| `moai spec view <SPEC-ID>` | 수용 기준을 트리 구조로 조회 |
| `moai spec lint [spec.md...]` | EARS 준수 및 구조 유효성 린트 |
| `moai spec close <SPEC-ID>` | 원자적 4-phase 종료 (status: completed + progress.md backfill) |
| `moai spec audit` | SPEC era 분류 및 modern-era 상태 드리프트 감사 |
| `moai spec archive` | 종료된 SPEC을 `.moai/specs/` 밖으로 아카이브 |

## moai spec status

```bash
moai spec status <SPEC-ID> <new-status>   # 상태 갱신
moai spec status --list                   # 전체 SPEC 나열
moai spec status --sync-git               # git log에서 상태 동기화
```

| 플래그 | 설명 |
|--------|------|
| `--dry-run` | 쓰기 없이 변경 미리보기 |
| `--list` | 전체 SPEC과 상태 나열 |
| `--sync-git` | main의 git log에서 SPEC 상태 동기화 |
| `--yes` | `--sync-git` 비대화형 자동 확인 (CI/파이프 필수) |

## moai spec drift

```bash
moai spec drift
```

| 플래그 | 설명 |
|--------|------|
| `--json` | JSON 형식 출력 |
| `--exit-code-on-drift` | 드리프트 감지 시 종료 코드 1 |
| `--count` | 드리프트 개수만 출력 |
| `--no-cache` | HEAD-SHA 결과 캐시 우회 후 재계산 |

## moai spec lint

```bash
moai spec lint [spec.md...]
```

| 플래그 | 설명 |
|--------|------|
| `--json` | JSON 형식 출력 |
| `--sarif` | SARIF 2.1.0 형식 출력 |
| `--strict` | 경고를 오류로 처리 |
| `--format <fmt>` | 출력 형식 (table) |

## moai spec close

```bash
moai spec close SPEC-ID
```

단일 커밋으로 SPEC을 `status: completed` 까지 원자적으로 전환합니다.

| 플래그 | 설명 |
|--------|------|
| `--backfill-only` | progress.md backfill만 수행 |
| `--dry-run` | 커밋 없이 미리보기 |
| `--force` | 확인 없이 강제 종료 |
| `--json` | JSON 형식 출력 |

## moai spec audit

```bash
moai spec audit
```

`.moai/specs/SPEC-*/` 를 훑어 각 SPEC을 era 휴리스틱으로 분류하고, modern-era 상태 드리프트를 찾아냅니다.

| 플래그 | 설명 |
|--------|------|
| `--json` | JSON 형식 출력 |
| `--filter-era <era>` | era로 필터 |
| `--filter-spec <id>` | SPEC ID로 필터 |
| `--include-grandfathered` | grandfather era SPEC 포함 |
| `--strict` | 엄격 모드 |

## moai spec archive

```bash
moai spec archive --dry-run   # 대상 확인 (이동 없음)
moai spec archive --yes       # 계획 적용
```

grace 윈도우(기본 90일)보다 오래된 terminal SPEC을 아카이브합니다.

| 플래그 | 설명 |
|--------|------|
| `--dry-run` | 이동 없이 대상 집합 보고 |
| `--yes` | 이동 확정 (적용에 필수) |
| `--grace-days <n>` | grace 윈도우 일수 (0 = 기본값 90) |
| `--json` | 계획을 JSON으로 출력 |

## 관련 문서

- [SPEC 기반 개발](/ko/core-concepts/spec-based-dev)
- [CLI 개요](/ko/getting-started/cli)
