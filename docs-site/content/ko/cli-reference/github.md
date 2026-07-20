---
title: moai github GitHub 연동
weight: 92
draft: false
---

`moai github` 은 GitHub 이슈 파싱, SPEC 링크, 워크플로우 자동화 커맨드를 제공합니다.

공통 플래그로 `--dry-run` (변경 없이 수행 내용만 표시)을 받습니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai github parse-issue <number>` | GitHub 이슈 파싱 및 내용 표시 |
| `moai github link-spec <issue-number> <spec-id>` | GitHub 이슈와 SPEC 문서를 양방향 연결 |

## moai github parse-issue

```bash
moai github parse-issue 123
```

GitHub 이슈를 파싱해 내용을 표시합니다.

## moai github link-spec

```bash
moai github link-spec 123 SPEC-AUTH-001
```

GitHub 이슈와 SPEC 문서 사이에 양방향 링크를 생성합니다.

## 예시

```bash
# 이슈 파싱
moai github parse-issue 42

# 이슈를 SPEC에 연결 (미리보기)
moai github link-spec 42 SPEC-AUTH-001 --dry-run

# 실제 연결
moai github link-spec 42 SPEC-AUTH-001
```

## 관련 문서

- [moai pr](/ko/cli-reference/pr) — PR CI 감시
- [CLI 개요](/ko/getting-started/cli)
