---
title: moai session 세션 레지스트리
weight: 65
draft: false
---

`moai session` 은 `.moai/state/active-sessions.json` 의 다중 세션 조율 레지스트리를 관리합니다. 여러 Claude Code 세션이 같은 프로젝트에서 동시에 작업할 때 발생하는 경쟁(race)을 완화하기 위한 도구입니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai session register <session_id> <spec_id> <phase>` | 새 활성 세션 등록 |
| `moai session heartbeat <session_id>` | 기존 세션의 last_heartbeat 갱신 (idempotent) |
| `moai session deregister <session_id>` | 세션 제거 (idempotent) |
| `moai session list` | 활성 세션 나열 (`--filter-spec` 필터 가능) |
| `moai session purge` | stale 항목 제거 (기본: 마지막 heartbeat 후 30분 초과) |
| `moai session current` | 이 오케스트레이터의 세션 UUID 출력 |
| `moai session doctor` | 레지스트리가 비어 있는 원인 진단 |

대부분의 하위 명령어는 `--json` 플래그로 기계 판독 출력을 지원합니다.

## moai session list

```bash
moai session list
moai session list --filter-spec SPEC-AUTH-001
```

| 플래그 | 설명 |
|--------|------|
| `--json` | 기계 판독 JSON 출력 (오케스트레이터 pre-spawn 검사 형식) |
| `--filter-spec <id>` | 해당 spec_id와 일치하는 항목만 반환 |

## moai session purge

```bash
moai session purge
```

| 플래그 | 설명 |
|--------|------|
| `--json` | JSON 출력 |
| `--threshold-minutes <n>` | stale heartbeat 컷오프 (분, 기본 30) |

## moai session current

```bash
moai session current
```

오케스트레이터 자신의 세션 UUID를 출력합니다. 런타임이 세션 ID를 노출하지 않으면 canonical fallback 문자열을 반환합니다.

| 플래그 | 설명 |
|--------|------|
| `--json` | JSON 출력 |
| `--show-fallback` | canonical fallback 문자열만 출력 (paste-ready resume 생성용) |

## moai session doctor

```bash
moai session doctor
```

다중 세션 조율 레지스트리가 비어 있는 이유를 진단합니다 (write-path 진단).

| 플래그 | 설명 |
|--------|------|
| `--json` | JSON 출력 |

## 사용 맥락

이 레지스트리는 오케스트레이터가 구현 에이전트를 spawn하기 전 동시 세션 경쟁을 감지하는 데 사용됩니다. `moai session list --json --filter-spec <SPEC-ID>` 가 다른 세션 항목을 반환하면 오케스트레이터는 진행을 멈추고 사용자에게 확인합니다.

## 관련 문서

- [moai inventory](/ko/cli-reference/inventory) — 세션·워크트리·하네스 통합 조회
- [CLI 개요](/ko/getting-started/cli)
