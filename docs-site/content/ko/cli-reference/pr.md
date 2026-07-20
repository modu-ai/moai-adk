---
title: moai pr PR 감시
weight: 96
draft: false
---

`moai pr` 은 CI/CD 워크플로우에서 pull request를 감시·관리하는 커맨드입니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai pr watch <PR_NUMBER>` | PR의 CI 체크 감시 (또는 `--abort` 로 활성 감시 중단) |

## moai pr watch

```bash
moai pr watch 123
```

지정한 PR 번호에 대해 `gh pr checks` 를 모니터링합니다.

| 플래그 | 설명 |
|--------|------|
| `--abort` | 활성 CI 감시 루프 중단 |
| `--report` | 해당 PR 번호에 대한 머지 준비 보고서 방출 |
| `--branch <name>` | 보고서 컨텍스트용 브랜치 이름 (기본: main) |

## 예시

```bash
# PR CI 체크 감시
moai pr watch 42

# 활성 감시 중단
moai pr watch 42 --abort

# 머지 준비 보고서
moai pr watch 42 --report
```

## 관련 문서

- [moai github](/ko/cli-reference/github)
- [CLI 개요](/ko/getting-started/cli)
