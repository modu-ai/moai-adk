---
title: moai doctor 진단
weight: 60
draft: false
---

`moai doctor` 는 시스템 전반을 한 번에 진단합니다. Claude Code 설정, 의존성, 프로젝트 구조, 언어별 개발 도구, 환경을 차례로 검사하고, 문제를 찾으면 고칠 방법까지 함께 알려 줍니다.

에이전트(스스로 일하는 AI)가 세팅을 무심히 건드리면서 설정 드리프트가 생기는 일이 잦기 때문에, 한 번에 전체 상태를 보여 주는 단일 진단 커맨드가 필요합니다. 하네스(harness) 게이트와 SPEC 라이프사이클이 모두 `moai` CLI 를 거쳐 동작하므로, 어느 한 축이 깨져 있으면 다른 커맨드의 메시지만으로는 원인을 짚기 어렵습니다. 따라서 `moai doctor` 는 개별 커맨드가 실패할 때 가장 먼저 실행해 볼 1차 진단 도구입니다.

## 개요

```bash
moai doctor [OPTIONS]
```

## 플래그

| 플래그 | 설명 |
|--------|------|
| `-v, --verbose` | 상세 진단 정보 (도구 버전, 언어 감지 결과) 표시 |
| `--fix` | 감지된 문제의 수정 방법 제안 |
| `--export` | 진단 결과를 JSON 파일로 내보내기 |
| `--check <tool>` | 특정 검사만 실행 (예: git, go, config) |

## 하위 명령어

특정 영역만 깊이 들여다볼 때 쓰는 하위 명령어도 있습니다.

| 명령어 | 설명 |
|--------|------|
| `moai doctor config` | 설정 진단 — 병합된 설정을 provenance 와 함께 검사 |
| `moai doctor hook` | 27개 훅 이벤트 커버리지 표시 |
| `moai doctor permission` | 권한 해석 진단 |
| `moai doctor sandbox` | 샌드박스 백엔드 가용성 진단 |

`moai doctor config` 는 다시 `dump`(병합 설정 덤프)와 `diff <tier-a> <tier-b>`(두 설정 티어 비교) 를 제공합니다.

## 예시

```bash
# 전체 진단
moai doctor

# 상세 진단
moai doctor --verbose

# 진단 결과 내보내기
moai doctor --export diagnostics.json

# 특정 영역 진단
moai doctor hook          # 훅 커버리지 표
moai doctor permission    # 권한 해석
moai doctor sandbox       # 샌드박스 백엔드
```

---

관련: [프로젝트 상태](/ko/cli-reference/status) · [CLI 개요](/ko/getting-started/cli)
