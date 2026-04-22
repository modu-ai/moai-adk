---
title: /moai feedback
weight: 80
draft: false
---

MoAI-ADK에 피드백이나 버그 리포트를 제출하는 명령어입니다.

{{< callout type="info" >}}

**새로운 명령어 형식**

`/moai:9-feedback`는 이제 `/moai feedback`으로 변경되었습니다.

{{< /callout >}}

{{< callout type="info" >}}
**한 줄 요약**: `/moai feedback`은 MoAI-ADK 자체에 대한 개선 제안이나 버그 리포트를 **GitHub 이슈로 자동 생성**해주는 명령어입니다.
{{< /callout >}}

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai:feedback`을 입력하면 이 명령어를 바로 실행할 수 있습니다. `/moai`만 입력하면 사용 가능한 모든 서브커맨드 목록이 표시됩니다.
{{< /callout >}}

## 개요

MoAI-ADK를 사용하다가 버그를 발견하거나, 새로운 기능이 필요하거나, 개선 아이디어가 떠올랐을 때 이 명령어를 사용합니다. 직접 GitHub에 접속하여 이슈를 작성할 필요 없이, Claude Code 안에서 바로 피드백을 제출할 수 있습니다.

{{< callout type="info" >}}
**중요**: 이 명령어는 **여러분의 프로젝트 코드를 수정하는 명령어가 아닙니다**. MoAI-ADK 도구 자체에 대한 피드백을 개발팀에 전달하는 명령어입니다.
{{< /callout >}}

## 사용법

```bash
# 표준 형식
> /moai feedback

# 짧은 별칭
> /moai fb
> /moai bug
> /moai issue
```

명령어를 실행하면 피드백 유형을 선택하고, 내용을 입력하는 과정을 안내받습니다.

## 지원 플래그

| 플래그 | 설명 | 예시 |
|-------|------|------|
| `--type {bug,feature,question}` | 피드백 유형 직접 지정 | `/moai feedback --type bug` |
| `--title "<title>"` | 제목 직접 지정 | `/moai feedback --title "오류 보고"` |
| `--dry-run` | 이슈 생성 없이 내용만 확인 | `/moai feedback --dry-run` |

## 작동 방식

`/moai feedback`을 실행하면 다음 과정이 진행됩니다.

```mermaid
flowchart TD
    A["/moai feedback 실행"] --> B["피드백 유형 선택"]
    B --> C["내용 작성"]
    C --> D["현재 환경 정보<br/>자동 수집"]
    D --> E["GitHub 이슈<br/>자동 생성"]
    E --> F["이슈 URL 반환"]
```

### 자동 수집되는 정보

피드백 제출 시 다음 정보가 자동으로 포함되어, 개발팀이 문제를 더 빠르게 파악할 수 있습니다.

| 수집 항목 | 설명 | 예시 |
|-----------|------|------|
| MoAI-ADK 버전 | 현재 설치된 버전 | v10.8.0 |
| OS 정보 | 운영체제 및 버전 | macOS 15.2 |
| Claude Code 버전 | 사용 중인 Claude Code 버전 | 1.0.30 |
| 현재 SPEC | 작업 중인 SPEC ID | SPEC-AUTH-001 |
| 오류 로그 | 최근 발생한 오류 (있는 경우) | TypeError: ... |

## 피드백 유형

### 버그 리포트

MoAI-ADK 사용 중 발생한 오류나 예상과 다른 동작을 보고합니다.

```bash
> /moai feedback
# 유형 선택: 버그 리포트
# 제목: /moai run 실행 시 특성화 테스트가 생성되지 않음
# 설명: SPEC-AUTH-001에 대해 /moai run을 실행했는데,
#        PRESERVE 단계에서 특성화 테스트가 생성되지 않고
#        바로 IMPROVE 단계로 넘어갑니다.
# 재현 방법: /moai run SPEC-AUTH-001 실행
```

### 기능 요청

MoAI-ADK에 추가되었으면 하는 새로운 기능을 제안합니다.

```bash
> /moai feedback
# 유형 선택: 기능 요청
# 제목: /moai loop에 특정 파일만 대상으로 하는 옵션 추가
# 설명: /moai loop 실행 시 전체 프로젝트가 아닌 특정 디렉토리나
#        파일만 대상으로 할 수 있으면 좋겠습니다.
# 예시: /moai loop --path src/auth/
```

### 개선 제안

기존 기능의 개선 아이디어를 제안합니다.

```bash
> /moai feedback
# 유형 선택: 개선 제안
# 제목: /moai fix 실행 결과에 수정 전후 diff 표시
# 설명: /moai fix가 자동 수정한 내용을 diff 형태로
#        보여주면 어떤 변경이 있었는지 한눈에 파악할 수 있겠습니다.
```

## 에이전트 위임 체인

`/moai feedback` 명령어의 에이전트 위임 흐름입니다:

```mermaid
flowchart TD
    User["사용자 요청"] --> Orchestrator["MoAI 오케스트레이터"]
    Orchestrator --> Collect["환경 정보 수집"]

    Collect --> Info1["MoAI-ADK 버전"]
    Collect --> Info2["OS 정보"]
    Collect --> Info3["Claude Code 버전"]
    Collect --> Info4["현재 SPEC"]
    Collect --> Info5["오류 로그"]

    Info1 --> Format["이슈 포맷팅"]
    Info2 --> Format
    Info3 --> Format
    Info4 --> Format
    Info5 --> Format

    Format --> GitHub["manager-quality 에이전트<br/>GitHub 이슈 생성"]
    GitHub --> Complete["이슈 URL 반환"]
```

**에이전트 역할:**

| 에이전트 | 역할 | 주요 작업 |
|----------|------|----------|
| **MoAI 오케스트레이터** | 피드백 프로세스 안내 |
| **manager-quality** | GitHub 연동 | 이슈 생성, URL 반환 |

## 실전 예시

### 상황: 명령어 실행 중 예상치 못한 오류 발생

```bash
# 오류가 발생한 상황
> /moai "결제 기능 구현" --branch
# Error: Branch creation failed - permission denied

# 피드백 제출
> /moai feedback
```

MoAI 오케스트레이터가 피드백 유형, 제목, 설명을 차례로 물어봅니다. 답변을 입력하면 자동으로 GitHub 이슈가 생성되고, 이슈 URL이 반환됩니다.

```
GitHub 이슈가 생성되었습니다:
https://github.com/anthropics/moai-adk/issues/1234

개발팀이 확인 후 답변드리겠습니다.
```

{{< callout type="info" >}}
**피드백은 언제든 환영합니다!** 사소한 불편 사항이라도 피드백을 제출해주시면 MoAI-ADK 개선에 큰 도움이 됩니다.
{{< /callout >}}

## 자주 묻는 질문

### Q: 피드백 내용을 수정하거나 삭제할 수 있나요?

네, GitHub에서 직접 이슈를 수정하거나 닫을 수 있습니다. 이슈 URL이 제공되므로 언제든 접근할 수 있습니다.

### Q: 같은 문제를 여러 번 보고해도 되나요?

GitHub에서 중복 이슈를 확인하므로 걱정하지 않아도 됩니다. 이미 보고된 문제라면 기존 이슈로 안내해줍니다.

### Q: 피드백에 대한 응답은 언제 받을 수 있나요?

개발팀이 확인 후 주마다 이슈에 댓글을 달아드립니다. 복잡한 문제의 경우 해결까지 시간이 걸 수 있습니다.

### Q: `/moai feedback`와 GitHub 직접 이슈 생성의 차이는 무엇인가요?

`/moai feedback`는 환경 정보를 자동으로 수집하여 개발팀이 문제를 더 빠르게 파악할 수 있게 해줍니다. 수동으로 이슈를 생성하는 것보다 더 효율적입니다.

## 관련 문서

- [/moai - 완전 자율 자동화](/utility-commands/moai)
- [/moai loop - 반복 수정 루프](/utility-commands/moai-loop)
- [/moai fix - 일회성 자동 수정](/utility-commands/moai-fix)
