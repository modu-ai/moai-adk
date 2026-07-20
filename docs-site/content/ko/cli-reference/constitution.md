---
title: moai constitution 헌법
weight: 84
draft: false
---

`moai constitution` 은 zone 레지스트리(FROZEN/EVOLVABLE zone 성문화)를 조회하고 검증합니다. 규칙의 어떤 부분이 동결(FROZEN)되어 함부로 바꿀 수 없고 어떤 부분이 진화 가능(EVOLVABLE)한지를 관리하는 커맨드 트리입니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai constitution list` | zone 레지스트리 항목 나열 |
| `moai constitution guard` | FROZEN zone 위반 검사 (CI 통합용) |
| `moai constitution amend` | 5-layer 안전 게이트를 통과하는 헌법 개정 제안 |
| `moai constitution validate` | 소스 파일 대비 zone 레지스트리 드리프트·불변조건 검증 |

## moai constitution list

```bash
moai constitution list
moai constitution list --zone frozen --format json
```

| 플래그 | 설명 |
|--------|------|
| `--zone <frozen\|evolvable>` | zone 필터 |
| `--file <path>` | 파일 경로 필터 (부분 일치) |
| `--format <table\|json>` | 출력 형식 |

## moai constitution guard

```bash
moai constitution guard --violations CONST-V3R2-001,CONST-V3R2-002
```

변경된 규칙 ID 목록을 받아 FROZEN zone 위반을 검사합니다. CI 통합용입니다.

| 플래그 | 설명 |
|--------|------|
| `--violations <ids>` | 변경된 규칙 ID 목록 (쉼표 구분 또는 반복 플래그) |

## moai constitution amend

```bash
moai constitution amend --rule CONST-V3R2-001 --before "..." --after "..." --evidence "..."
```

FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight 5-layer 안전 게이트를 통과해야 적용됩니다.

| 플래그 | 설명 |
|--------|------|
| `--rule <id>` | 규칙 ID (CONST-V3R2-NNN) [필수] |
| `--before <text>` | 현재 조항 텍스트 [필수] |
| `--after <text>` | 새 조항 텍스트 [필수] |
| `--evidence <text>` | 개정 근거 (Frozen zone 필수) |
| `--dry-run` | 파일 수정 없이 시뮬레이션만 |

## moai constitution validate

```bash
moai constitution validate
```

레지스트리 각 항목의 조항이 소스 파일에 존재하는지 확인하고, zone_class enum·canary_gate 불변조건을 검증하며 드리프트를 보고합니다.

| 플래그 | 설명 |
|--------|------|
| `--strict` | 엄격 모드 (모든 검사 강제) |
| `--fail-on-warning` | 경고를 오류로 처리 (`--strict` 포함) |
| `--format <text\|json>` | 출력 형식 |

종료 코드: 0=정상, 1=드리프트/오류, 2=치명(소스 파일 없음).

## 관련 문서

- [CLI 개요](/ko/getting-started/cli)
