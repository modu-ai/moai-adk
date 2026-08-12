---
title: moai github GitHub 연동
weight: 92
draft: false
---

`moai github` 는 GitHub 이슈 파싱, SPEC 링크, 워크플로우 자동화 커맨드를 제공합니다.

이슈 트래커와 SPEC 문서가 서로 분리돼 있으면 맥락이 끊어지기 쉽기 때문에, 이 커맨드는 GitHub 이슈 본문에서 힌트를 뽑아 SPEC 으로 연결해 줍니다. 에이전트(스스로 일하는 AI)가 이슈를 해석할 때 일관된 입력을 받도록 정규화된 형태를 제공하므로, 자동화 파이프라인의 첫 단계로 활용하기에 적합합니다.

공통 플래그로 `--dry-run` (아무것도 바꾸지 않고 수행할 내용만 표시)을 받습니다.

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
