---
title: moai doctor 진단
weight: 60
draft: false
---

`moai doctor` 는 종합 시스템 진단을 실행합니다. Claude Code 설정, 의존성, 프로젝트 구조, 언어별 개발 도구, 환경을 검사하고 문제에 대한 수정 제안을 제공합니다.

## 개요

```bash
moai doctor [OPTIONS]
```

## 플래그

| 플래그 | 설명 |
|--------|------|
| `-v, --verbose` | 상세 진단 정보 (도구 버전, 언어 감지 결과) 표시 |
| `--fix` | 감지된 문제에 대한 수정 제안 |
| `--export` | 진단 결과를 JSON 파일로 내보내기 |
| `--check <tool>` | 특정 검사만 실행 (예: git, go, config) |

## 하위 명령어

`moai doctor` 는 특정 영역을 깊이 진단하는 하위 명령어를 제공합니다.

| 명령어 | 설명 |
|--------|------|
| `moai doctor config` | 설정 진단 — 병합된 설정을 provenance 와 함께 검사 |
| `moai doctor hook` | 27개 훅 이벤트 커버리지 표 표시 |
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

관련: [프로젝트 상태](/cli-reference/status) · [CLI 개요](/getting-started/cli)
