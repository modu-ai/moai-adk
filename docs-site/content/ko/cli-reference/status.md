---
title: 프로젝트 상태
weight: 20
draft: false
---

`moai status` 명령어로 현재 프로젝트의 초기화 상태, SPEC 개수, 설정 파일 현황을 한눈에 확인합니다. 플래그가 없는 읽기 전용 명령어입니다.

## 사용법

```bash
moai status
```

플래그 없이 실행하면 프로젝트 상태를 박스 형태로 출력합니다.

## 출력 내용

### 초기화된 프로젝트

`.moai/` 디렉토리가 존재하는 프로젝트에서 실행하면 다음 정보가 표시됩니다.

| 항목 | 설명 |
|------|------|
| **Project** | 프로젝트 이름 (현재 디렉토리명) |
| **ADK** | 설치된 MoAI-ADK 버전 |
| **Config** | 설정 파일 경로 (`.moai/config/sections`) |
| **SPECs** | `.moai/specs/` 하위의 SPEC 디렉토리 개수 |
| **Configs** | `.moai/config/sections/` 의 YAML 파일 개수 |

하단에 초기화 상태와 SPEC 개수를 나타내는 상태 표시가 함께 출력됩니다.

### 미초기화 프로젝트

`.moai/` 디렉토리가 없는 경우, "Not initialized" 상태 표시와 함께 `moai init` 실행을 안내합니다.

## BODP 브랜치 알림

프로젝트가 Git 저장소인 경우, 현재 브랜치가 BODP (Branch-Oriented Development Practice) 규약을 벗어나면 stderr에 알림이 출력됩니다. 이 알림은 분산된 1인 OSS 워크플로우에서 브랜치 네이밍 규칙을 상기시키는 장치입니다.

알림은 자동으로 출력되며, Git이 설치되지 않았거나 현재 디렉토리가 Git 저장소가 아닌 경우 조용히 생략됩니다.

## 관련 명령어

| 명령어 | 설명 |
|--------|------|
| `moai doctor` | 시스템 진단 및 환경 검증 (상세 확인) |
| `moai inventory` | 활성 세션, 워크트리, 하네스 통합 조회 |
| `moai init` | 프로젝트 초기화 (미초기화 시 실행) |

## 참고

- [CLI 레퍼런스](./cli) — 전체 CLI 명령어
- [moai inventory](./inventory) — 활성 자원 통합 조회
- [초기 설정](./init-wizard) — 프로젝트 초기화 마법사
