---
title: moai ast-grep / ast-edit 구조 검색·치환
weight: 77
draft: false
---

`moai ast-grep` 은 코드를 구문 트리 단위로 스캔하고, `moai ast-edit` 은 매칭된 코드를 실제로 치환합니다. 텍스트 기반 `grep` 과 달리 구문 구조로 매칭하므로 공백·줄바꿈·변수명 차이에 흔들리지 않습니다.

두 커맨드는 [ast-grep](https://ast-grep.github.io/) CLI(`sg`)를 사용하며, `sg` 가 없을 때 동작이 의도적으로 다릅니다. `ast-grep` 은 CI 게이트로도 쓰이는 검사 커맨드라 안내를 stderr 로 출력하고 **0 이 아닌 코드로 종료**합니다. 실행되지 않은 스캔을 "이상 없음" 으로 읽어서는 안 되기 때문입니다. `ast-edit` 은 치환 커맨드이고 "적용할 것이 없음" 은 정상적인 무동작이므로 안내만 출력하고 0 으로 종료합니다. `sg` 설치는 [ast-grep 퀵스타트](https://ast-grep.github.io/guide/quick-start.html) 를 참고하세요. 설치 여부는 `moai doctor` 로도 확인할 수 있습니다.

> **읽기와 쓰기가 분리돼 있습니다.** `ast-grep` 은 파일을 절대 수정하지 않고, `ast-edit` 은 수정합니다. 별도 커맨드이므로 `Bash(moai ast-grep:*)` 권한을 허용해도 쓰기 권한까지 열리지 않습니다.

## moai ast-grep — 스캔 (읽기 전용)

| 플래그 | 설명 |
|--------|------|
| `--format` | 출력 형식: `text`(기본) · `json` · `sarif` |
| `--lang` | 지정한 언어만 스캔 (예: `go`, `python`, `typescript`) |
| `--severity` | 표시할 최소 심각도 (`error` · `warning` · `info`) |
| `--rules-dir` | 룰 디렉터리 경로 (기본 `.moai/config/astgrep-rules`) |
| `--dry` | 적용될 룰 목록만 출력하고 실제 스캔은 생략 |

```bash
# 프로젝트 전체 스캔
moai ast-grep ./

# Go 코드만, SARIF로 출력 (GitHub code scanning 업로드용)
moai ast-grep --format=sarif --lang=go ./internal/

# error 심각도만 표시
moai ast-grep --severity=error ./
```

## moai ast-edit — 치환 (파일 수정)

`--dry` 없이 실행하면 **파일을 직접 수정합니다.** 먼저 `--dry` 로 무엇이 바뀌는지 확인한 뒤 적용하세요.

| 플래그 | 설명 |
|--------|------|
| `--dry` | 파일을 수정하지 않고 변경될 내용만 출력 |
| `--pattern` | 매칭할 ast-grep 패턴 (`--rewrite` 와 함께 사용) |
| `--rewrite` | 치환할 패턴 (`--pattern` 과 함께 사용) |
| `--rule` | 지정한 ID의 룰만 적용 (룰 모드) |
| `--lang` | 대상 코드의 언어 |
| `--rules-dir` | 룰 디렉터리 경로 (기본 `.moai/config/astgrep-rules`) |
| `--format` | 출력 형식: `text`(기본) · `json` |

### 패턴 모드

`--pattern` 과 `--rewrite` 를 함께 지정하면 매칭된 모든 코드를 치환합니다. 둘 중 하나만 지정하면 오류로 거부됩니다.

```bash
# 먼저 미리보기
moai ast-edit --dry --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/

# 확인 후 실제 적용
moai ast-edit --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/
```

### 룰 모드

`--pattern` 없이 실행하면 룰 디렉터리를 읽어 `fix:` 필드를 선언한 룰만 적용합니다. `fix:` 가 없는 룰은 탐지 전용이므로 건너뛰고 그 개수를 알려줍니다.

```bash
# 모든 fix: 룰 적용 (미리보기)
moai ast-edit --dry ./internal/

# 특정 룰만 적용
moai ast-edit --rule my-rule-id ./internal/
```

기본 배포 룰셋(`go/hardcoding`, `security/credentials`, `security/crypto`, `security/injection`)은 모두 **탐지 전용**입니다. 자동 치환이 의미를 바꾸거나 컴파일을 깨뜨릴 수 있어 의도적으로 `fix:` 를 넣지 않았습니다. 자동 치환이 필요하면 프로젝트 룰에 직접 `fix:` 를 선언하세요.

## 룰 파일 위치

두 커맨드 모두 기본적으로 `.moai/config/astgrep-rules/` 를 읽습니다. `sgconfig.yml` 이 룰 디렉터리 목록을 정의합니다.

## 관련 문서

- [moai loop](/ko/cli-reference/loop) — 진단 기반 반복 수정 루프
- [CLI 개요](/ko/getting-started/cli)
