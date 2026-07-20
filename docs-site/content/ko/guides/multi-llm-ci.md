---
title: "GitHub 연동 가이드"
description: "moai github 서브커맨드로 이슈를 파싱하고 SPEC과 연결하기"
draft: false
weight: 10
---

MoAI-ADK의 GitHub 연동 기능은 GitHub 이슈를 파싱하고 SPEC 문서와 연결하는
경량 CLI 도구를 제공합니다. 모든 명령은 로컬에 설치된 `gh` CLI를 통해
현재 리포지토리의 이슈 데이터를 가져옵니다.

> **범위 안내**: 이 페이지는 실제로 배포되는 `moai github` 서브커맨드와
> 함께 제공되는 GitHub Actions 자산만 다룹니다. 여러 LLM을 PR에 패널로
> 붙이는 "멀티 LLM 리뷰 패널"은 현재 배포 릴리스에 포함되어 있지 않습니다.

## 사전 요구사항

- MoAI-ADK 설치 (macOS · Linux · Windows)
- GitHub CLI (`gh`) 설치 및 인증 (`gh auth login`)
- GitHub 리포지토리

## moai github 서브커맨드

`moai github`는 두 개의 활성 서브커맨드를 제공합니다. 공통으로
`--dry-run` 플래그를 지원하여 실제 변경 없이 수행할 작업만 미리 볼 수
있습니다.

### 이슈 파싱: `moai github parse-issue`

```bash
moai github parse-issue 123
```

`gh` CLI로 지정한 번호의 이슈를 가져와 번호·제목·작성자·라벨·본문
요약·코멘트 수를 카드 형태로 출력합니다.

### SPEC 연결: `moai github link-spec`

```bash
moai github link-spec 123 SPEC-ISSUE-123
```

GitHub 이슈와 SPEC 문서 사이에 양방향 링크를 만들고, 그 매핑을
`.moai/github-spec-registry.json`에 저장합니다. SPEC ID는 저장 전에
형식 검증을 거칩니다.

```bash
# 실제 변경 없이 계획만 확인
moai github link-spec 123 SPEC-ISSUE-123 --dry-run
```

## 함께 배포되는 GitHub Actions 자산

`moai init`은 `.github/` 아래에 다음 두 자산을 배포합니다.

### Label Sync 워크플로우 (`.github/workflows/label-sync.yml`)

`.github/labels.yml`을 단일 진실 공급원으로 삼아 리포지토리 라벨을
동기화합니다.

- **트리거**: `workflow_dispatch` (수동, `dry_run` 입력 지원) 또는
  `.github/labels.yml` / 워크플로우 파일이 `main`에 push될 때 자동 실행
- **권한**: `issues: write`, `pull-requests: write`, `contents: read`
- **동작**: EndBug/label-sync 액션으로 `labels.yml` → 리포 라벨 반영

### detect-language 컴포지트 액션 (`.github/actions/detect-language/action.yml`)

리포지토리의 첫 번째 소스 파일 확장자를 기준으로 주 언어를 감지하여
`language` 출력값으로 내보냅니다.

- **지원 언어 (16개)**: Go, Python, TypeScript, JavaScript, Rust, Java,
  Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift
- **구현 메모**: `find ... -print -quit`로 첫 매치 후 즉시 종료하여
  `set -o pipefail` 환경에서 broken-pipe 실패를 피합니다

## 트러블슈팅

### `gh` 명령을 찾을 수 없을 때

`moai github` 서브커맨드는 로컬 `gh` CLI에 의존합니다. `gh --version`으로
설치를 확인하고 `gh auth login`으로 인증을 마치세요.

### 이슈를 가져오지 못할 때

현재 디렉터리가 대상 리포지토리의 작업 트리 안에 있는지, 그리고 `gh`가
해당 리포에 접근 권한이 있는지 확인하세요.

### SPEC ID 검증 실패

`link-spec`은 `SPEC-` 접두사를 따르는 유효한 SPEC ID만 받습니다. ID
형식을 확인한 뒤 재실행하세요.

## 다음 단계

- [CLI 레퍼런스 참조](/ko/workflow-commands/)
- [Workflow 설정 참조](/ko/advanced/settings-json/)
- [보안 정책 확인](/ko/advanced/security-notes/)
