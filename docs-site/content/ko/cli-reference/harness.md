---
title: moai harness 하네스
weight: 80
draft: false
---

`moai harness` 는 SPEC 복잡도 라우팅과 하네스 학습 서브시스템을 관리하는 통합 커맨드 트리입니다. 라우팅, 검증, 라이프사이클, 제안 관리, v4 하네스 라이프사이클, 관찰 원장(ledger) 하위 명령어를 제공합니다.

공통 플래그로 `--project-root <path>` (기본: 현재 디렉터리)를 받습니다.

## 라우팅 verb

| 명령어 | 설명 |
|--------|------|
| `moai harness route --spec <id>` | SPEC을 minimal/standard/thorough 하네스 레벨로 라우팅 |
| `moai harness validate` | harness.yaml을 스키마·불변조건 기준으로 검증 |

`route` 는 `--json` (JSON 출력), `--path <harness.yaml>`, `--base-dir <dir>` 플래그를 받습니다.

## 라이프사이클 verb

| 명령어 | 설명 |
|--------|------|
| `moai harness status` | 관찰/티어/진화 요약 표시 |
| `moai harness apply` | 대기 중인 제안을 오케스트레이터에 반환 (또는 `--execute` 로 Go apply 경로 실행) |
| `moai harness rollback <date>` | 지정 날짜의 스냅샷 복원 |
| `moai harness disable` | 학습 서브시스템 비활성화 (`learning.enabled: false`) |

## 제안 관리 verb

| 명령어 | 설명 |
|--------|------|
| `moai harness mute` | 제안 카테고리 음소거 (workflow.yaml) |
| `moai harness mute-list` | 현재 음소거된 카테고리 출력 |
| `moai harness unmute` | 카테고리를 음소거 목록에서 제거 |
| `moai harness verify` | 하네스 결정성 검증 |

## v4 하네스 라이프사이클 verb

| 명령어 | 설명 |
|--------|------|
| `moai harness list` | 모든 v4 하네스 나열 (이름 + 도메인 + 진입 커맨드) |
| `moai harness edit <name>` | v4 하네스 manifest + specialist 편집 경로 표시 |
| `moai harness remove <name>` | v4 하네스 원자적 제거 (커맨드 + workflow + specialist + skill + manifest) |
| `moai harness doctor` | 하네스 설치 상태 진단 |

`list`, `edit` 는 `--json` 플래그를 받습니다.

## 관찰 원장 (ledger)

`moai harness ledger` 는 라우팅 관찰 원장을 관리합니다.

| 명령어 | 설명 |
|--------|------|
| `moai harness ledger record` | dispatch 시점 라우팅 결정 기록 (pending row) |
| `moai harness ledger evidence` | pending row에 기계 증거 ref (또는 위임 항목) 추가 |
| `moai harness ledger list` | 필터로 최종 원장 row 나열 |

> `ledger record` 와 `ledger evidence` 는 결과(outcome)를 위조할 수 없도록 `--outcome` 플래그를 노출하지 않습니다. 결과는 기계 증거에서 도출됩니다.

## 관련 문서

- [하네스 자가 진화](/ko/advanced/self-evolving)
- [Harness v4 Builder 심화 가이드](/ko/advanced/harness-v4-builder)
- [하네스 프로필과 평가](/ko/advanced/harness-profiles)
- [CLI 개요](/ko/getting-started/cli)
