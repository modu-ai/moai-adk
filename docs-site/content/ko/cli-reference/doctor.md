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

## Home Disk Usage 진단 {{< new-badge v3.1.1 >}}

`moai doctor` 전체 진단에는 **Home Disk Usage** 항목이 함께 나옵니다. `~/.moai` 홈 디렉터리가 얼마나 찼는지를 보고하는 **권고(advisory)** 성격의 검사라, 임계값을 넘어도 다른 명령을 막지 않습니다.

| 보고 항목 | 내용 |
|-----------|------|
| 전체 크기 | `~/.moai` 총 용량과 상위 3개 항목 |
| 프로필별 내역 | `claude-profiles/<프로필>` 각각의 크기와 범주 분해 |
| 릴리즈 개수 | `releases/`에 남아 있는 바이너리 수와 현재 버전 |
| 정리 가능량 | `moai clean --home`이 실제로 지울 수 있는 추정 바이트 |
| `~/.claude` | 크기만 보고 — 어떤 경로로도 정리 대상이 아님 |

정리 가능량이 임계값(컴파일된 기본값 500 MB)을 넘으면 상태가 WARN으로 바뀌고 `moai clean --home`(기본 dry-run)을 권합니다. 그 아래면 OK로 남습니다. `~/.moai`가 아예 없으면 "보고할 것 없음"으로 OK 처리됩니다.

이 추정치는 `moai clean --home`이 쓰는 것과 **같은 스캐너**를 호출하므로, doctor가 말하는 숫자와 clean이 실제로 지우는 목록이 어긋나지 않습니다. 자세한 내용은 [홈 디렉터리 위생](/ko/advanced/home-hygiene)에 있습니다.

## 종료 코드

스크립트나 CI 래퍼에서 `moai doctor` 를 부를 때는 요약 줄이 아니라 종료 코드를 읽습니다.

| 종료 코드 | 의미 |
|-----------|------|
| `0` | Fail 항목 없음. Warn 은 권고라 종료 코드를 바꾸지 않습니다 |
| `1` | 한 건 이상이 Fail. 요약의 `Fail N` 이 그대로 반영됩니다 |

Constitution Registry 항목은 레지스트리가 파싱되는지만 보지 않고 `moai constitution validate` 와 **같은 드리프트 검사**를 돌립니다. 따라서 같은 체크아웃에서 doctor 가 ok 라고 하는데 validate 가 실패하는 일은 없습니다. `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 로 우회하면 doctor 도 구조 검사 판정으로 돌아갑니다.

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
