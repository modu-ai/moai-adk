# 일반적인 워크플로우

> Claude Code를 사용한 일반적인 워크플로우에 대해 알아보세요.

이 문서의 각 작업에는 Claude Code를 최대한 활용하는 데 도움이 되는 명확한 지침, 예제 명령어 및 모범 사례가 포함되어 있습니다.

## 새로운 코드베이스 이해하기

### 빠른 코드베이스 개요 얻기

새 프로젝트에 막 합류하여 구조를 빠르게 이해해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="프로젝트 루트 디렉토리로 이동">
    ```bash
    cd /path/to/project 
    ```
  </Step>

  <Step title="Claude Code 시작">
    ```bash
    claude 
    ```
  </Step>

  <Step title="고수준 개요 요청">
    ```
    > give me an overview of this codebase 
    ```
  </Step>

  <Step title="특정 구성 요소에 대해 더 자세히 알아보기">
    ```
    > explain the main architecture patterns used here 
    ```

    ```
    > what are the key data models?
    ```

    ```
    > how is authentication handled?
    ```

  </Step>
</Steps>

<Tip>
  팁:

- 광범위한 질문부터 시작한 다음 특정 영역으로 범위를 좁혀나가세요
- 프로젝트에서 사용되는 코딩 규칙과 패턴에 대해 물어보세요
- 프로젝트별 용어 사전을 요청하세요
  </Tip>

### 관련 코드 찾기

특정 기능이나 기능성과 관련된 코드를 찾아야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="Claude에게 관련 파일 찾기 요청">
    ```
    > find the files that handle user authentication 
    ```
  </Step>

  <Step title="구성 요소가 어떻게 상호 작용하는지에 대한 컨텍스트 얻기">
    ```
    > how do these authentication files work together? 
    ```
  </Step>

  <Step title="실행 흐름 이해">
    ```
    > trace the login process from front-end to database 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- 찾고 있는 것에 대해 구체적으로 설명하세요
- 프로젝트의 도메인 언어를 사용하세요
  </Tip>

---

## 효율적으로 버그 수정하기

오류 메시지가 발생했고 그 원인을 찾아 수정해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="Claude와 오류 공유">
    ```
    > I'm seeing an error when I run npm test 
    ```
  </Step>

  <Step title="수정 권장사항 요청">
    ```
    > suggest a few ways to fix the @ts-ignore in user.ts 
    ```
  </Step>

  <Step title="수정 적용">
    ```
    > update user.ts to add the null check you suggested 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- Claude에게 문제를 재현하는 명령어를 알려주고 스택 추적을 얻으세요
- 오류를 재현하는 단계를 언급하세요
- 오류가 간헐적인지 일관적인지 Claude에게 알려주세요
  </Tip>

---

## 코드 리팩토링

기존 코드를 최신 패턴과 관행을 사용하도록 업데이트해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="리팩토링할 레거시 코드 식별">
    ```
    > find deprecated API usage in our codebase 
    ```
  </Step>

  <Step title="리팩토링 권장사항 얻기">
    ```
    > suggest how to refactor utils.js to use modern JavaScript features 
    ```
  </Step>

  <Step title="안전하게 변경사항 적용">
    ```
    > refactor utils.js to use ES2024 features while maintaining the same behavior 
    ```
  </Step>

  <Step title="리팩토링 검증">
    ```
    > run tests for the refactored code 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- Claude에게 최신 접근 방식의 이점을 설명해 달라고 요청하세요
- 필요할 때 변경사항이 하위 호환성을 유지하도록 요청하세요
- 작고 테스트 가능한 단위로 리팩토링을 수행하세요
  </Tip>

---

## 전문 서브에이전트 사용

특정 작업을 더 효과적으로 처리하기 위해 전문 AI 서브에이전트를 사용하고 싶다고 가정해 보겠습니다.

<Steps>
  <Step title="사용 가능한 서브에이전트 보기">
    ```
    > /agents
    ```

    이것은 사용 가능한 모든 서브에이전트를 보여주고 새로운 것을 만들 수 있게 해줍니다.

  </Step>

  <Step title="서브에이전트 자동 사용">
    Claude Code는 적절한 작업을 전문 서브에이전트에게 자동으로 위임합니다:

    ```
    > review my recent code changes for security issues
    ```

    ```
    > run all tests and fix any failures
    ```

  </Step>

  <Step title="특정 서브에이전트 명시적 요청">
    ```
    > use the code-reviewer subagent to check the auth module
    ```

    ```
    > have the debugger subagent investigate why users can't log in
    ```

  </Step>

  <Step title="워크플로우를 위한 사용자 정의 서브에이전트 생성">
    ```
    > /agents
    ```

    그런 다음 "Create New subagent"를 선택하고 프롬프트에 따라 다음을 정의하세요:

    * 서브에이전트 유형 (예: `api-designer`, `performance-optimizer`)
    * 언제 사용할지
    * 어떤 도구에 액세스할 수 있는지
    * 전문 시스템 프롬프트

  </Step>
</Steps>

<Tip>
  팁:

- 팀 공유를 위해 `.claude/agents/`에 프로젝트별 서브에이전트를 생성하세요
- 자동 위임을 가능하게 하려면 설명적인 `description` 필드를 사용하세요
- 각 서브에이전트가 실제로 필요한 것으로만 도구 액세스를 제한하세요
- 자세한 예제는 [서브에이전트 문서](/ko/docs/claude-code/sub-agents)를 확인하세요
  </Tip>

---

## 테스트 작업

커버되지 않은 코드에 대한 테스트를 추가해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="테스트되지 않은 코드 식별">
    ```
    > find functions in NotificationsService.swift that are not covered by tests 
    ```
  </Step>

  <Step title="테스트 스캐폴딩 생성">
    ```
    > add tests for the notification service 
    ```
  </Step>

  <Step title="의미 있는 테스트 케이스 추가">
    ```
    > add test cases for edge conditions in the notification service 
    ```
  </Step>

  <Step title="테스트 실행 및 검증">
    ```
    > run the new tests and fix any failures 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- 엣지 케이스와 오류 조건을 다루는 테스트를 요청하세요
- 적절할 때 단위 테스트와 통합 테스트 모두를 요청하세요
- Claude에게 테스트 전략을 설명해 달라고 하세요
  </Tip>

---

## 풀 리퀘스트 생성

변경사항에 대해 잘 문서화된 풀 리퀘스트를 생성해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="변경사항 요약">
    ```
    > summarize the changes I've made to the authentication module 
    ```
  </Step>

  <Step title="Claude로 PR 생성">
    ```
    > create a pr 
    ```
  </Step>

  <Step title="검토 및 개선">
    ```
    > enhance the PR description with more context about the security improvements 
    ```
  </Step>

  <Step title="테스트 세부사항 추가">
    ```
    > add information about how these changes were tested 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- Claude에게 직접 PR을 만들어 달라고 요청하세요
- 제출하기 전에 Claude가 생성한 PR을 검토하세요
- Claude에게 잠재적 위험이나 고려사항을 강조해 달라고 요청하세요
  </Tip>

## 문서 처리

코드에 대한 문서를 추가하거나 업데이트해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="문서화되지 않은 코드 식별">
    ```
    > find functions without proper JSDoc comments in the auth module 
    ```
  </Step>

  <Step title="문서 생성">
    ```
    > add JSDoc comments to the undocumented functions in auth.js 
    ```
  </Step>

  <Step title="검토 및 개선">
    ```
    > improve the generated documentation with more context and examples 
    ```
  </Step>

  <Step title="문서 검증">
    ```
    > check if the documentation follows our project standards 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- 원하는 문서 스타일을 지정하세요 (JSDoc, docstrings 등)
- 문서에 예제를 요청하세요
- 공개 API, 인터페이스 및 복잡한 로직에 대한 문서를 요청하세요
  </Tip>

---

## 이미지 작업

코드베이스에서 이미지 작업을 해야 하고 Claude의 이미지 콘텐츠 분석 도움을 받고 싶다고 가정해 보겠습니다.

<Steps>
  <Step title="대화에 이미지 추가">
    다음 방법 중 하나를 사용할 수 있습니다:

    1. Claude Code 창에 이미지를 드래그 앤 드롭
    2. 이미지를 복사하고 ctrl+v로 CLI에 붙여넣기 (cmd+v는 사용하지 마세요)
    3. Claude에게 이미지 경로 제공. 예: "Analyze this image: /path/to/your/image.png"

  </Step>

  <Step title="Claude에게 이미지 분석 요청">
    ```
    > What does this image show?
    ```

    ```
    > Describe the UI elements in this screenshot
    ```

    ```
    > Are there any problematic elements in this diagram?
    ```

  </Step>

  <Step title="컨텍스트를 위해 이미지 사용">
    ```
    > Here's a screenshot of the error. What's causing it?
    ```

    ```
    > This is our current database schema. How should we modify it for the new feature?
    ```

  </Step>

  <Step title="시각적 콘텐츠에서 코드 제안 받기">
    ```
    > Generate CSS to match this design mockup
    ```

    ```
    > What HTML structure would recreate this component?
    ```

  </Step>
</Steps>

<Tip>
  팁:

- 텍스트 설명이 불분명하거나 번거로울 때 이미지를 사용하세요
- 더 나은 컨텍스트를 위해 오류 스크린샷, UI 디자인 또는 다이어그램을 포함하세요
- 대화에서 여러 이미지로 작업할 수 있습니다
- 이미지 분석은 다이어그램, 스크린샷, 목업 등과 함께 작동합니다
  </Tip>

---

## 파일 및 디렉토리 참조

@를 사용하여 Claude가 읽기를 기다리지 않고 파일이나 디렉토리를 빠르게 포함하세요.

<Steps>
  <Step title="단일 파일 참조">
    ```
    > Explain the logic in @src/utils/auth.js
    ```

    이것은 파일의 전체 내용을 대화에 포함합니다.

  </Step>

  <Step title="디렉토리 참조">
    ```
    > What's the structure of @src/components?
    ```

    이것은 파일 정보가 포함된 디렉토리 목록을 제공합니다.

  </Step>

  <Step title="MCP 리소스 참조">
    ```
    > Show me the data from @github:repos/owner/repo/issues
    ```

    이것은 @server:resource 형식을 사용하여 연결된 MCP 서버에서 데이터를 가져옵니다. 자세한 내용은 [MCP 리소스](/ko/docs/claude-code/mcp#use-mcp-resources)를 참조하세요.

  </Step>
</Steps>

<Tip>
  팁:

- 파일 경로는 상대 경로 또는 절대 경로일 수 있습니다
- @ 파일 참조는 파일의 디렉토리와 상위 디렉토리에 있는 CLAUDE.md를 컨텍스트에 추가합니다
- 디렉토리 참조는 내용이 아닌 파일 목록을 보여줍니다
- 단일 메시지에서 여러 파일을 참조할 수 있습니다 (예: "@file1.js and @file2.js")
  </Tip>

---

## 확장된 사고 사용

복잡한 아키텍처 결정, 도전적인 버그 또는 깊은 추론이 필요한 다단계 구현 계획 작업을 하고 있다고 가정해 보겠습니다.

<Steps>
  <Step title="컨텍스트 제공 및 Claude에게 사고 요청">
    ```
    > I need to implement a new authentication system using OAuth2 for our API. Think deeply about the best approach for implementing this in our codebase. 
    ```

    Claude는 코드베이스에서 관련 정보를 수집하고
    인터페이스에서 볼 수 있는 확장된 사고를 사용합니다.

  </Step>

  <Step title="후속 프롬프트로 사고 개선">
    ```
    > think about potential security vulnerabilities in this approach 
    ```

    ```
    > think harder about edge cases we should handle
    ```

  </Step>
</Steps>

<Tip>
  확장된 사고에서 최대 가치를 얻기 위한 팁:

확장된 사고는 다음과 같은 복잡한 작업에 가장 유용합니다:

- 복잡한 아키텍처 변경 계획
- 복잡한 문제 디버깅
- 새로운 기능에 대한 구현 계획 생성
- 복잡한 코드베이스 이해
- 다양한 접근 방식 간의 트레이드오프 평가

사고를 프롬프트하는 방식에 따라 다양한 수준의 사고 깊이가 나타납니다:

- "think"는 기본적인 확장된 사고를 트리거합니다
- "think more", "think a lot", "think harder" 또는 "think longer"와 같은 강화 문구는 더 깊은 사고를 트리거합니다

더 많은 확장된 사고 프롬프팅 팁은 [확장된 사고 팁](/ko/docs/build-with-claude/prompt-engineering/extended-thinking-tips)을 참조하세요.
</Tip>

<Note>
  Claude는 응답 위에 기울임꼴 회색 텍스트로 사고 과정을 표시합니다.
</Note>

---

## 이전 대화 재개

Claude Code로 작업을 하고 있었고 나중에 세션에서 중단한 부분부터 계속해야 한다고 가정해 보겠습니다.

Claude Code는 이전 대화를 재개하기 위한 두 가지 옵션을 제공합니다:

- `--continue`로 가장 최근 대화를 자동으로 계속
- `--resume`으로 대화 선택기 표시

<Steps>
  <Step title="가장 최근 대화 계속">
    ```bash
    claude --continue
    ```

    이것은 프롬프트 없이 가장 최근 대화를 즉시 재개합니다.

  </Step>

  <Step title="비대화형 모드에서 계속">
    ```bash
    claude --continue --print "Continue with my task"
    ```

    `--continue`와 함께 `--print`를 사용하여 가장 최근 대화를 비대화형 모드에서 재개하세요. 스크립트나 자동화에 완벽합니다.

  </Step>

  <Step title="대화 선택기 표시">
    ```bash
    claude --resume
    ```

    이것은 다음을 보여주는 대화형 대화 선택기를 표시합니다:

    * 대화 시작 시간
    * 초기 프롬프트 또는 대화 요약
    * 메시지 수

    화살표 키를 사용하여 탐색하고 Enter를 눌러 대화를 선택하세요.

  </Step>
</Steps>

<Tip>
  팁:

- 대화 기록은 컴퓨터에 로컬로 저장됩니다
- 가장 최근 대화에 빠르게 액세스하려면 `--continue`를 사용하세요
- 특정 과거 대화를 선택해야 할 때는 `--resume`을 사용하세요
- 재개할 때 계속하기 전에 전체 대화 기록을 볼 수 있습니다
- 재개된 대화는 원래와 동일한 모델 및 구성으로 시작됩니다

작동 방식:

1. **대화 저장**: 모든 대화는 전체 메시지 기록과 함께 자동으로 로컬에 저장됩니다
2. **메시지 역직렬화**: 재개할 때 컨텍스트를 유지하기 위해 전체 메시지 기록이 복원됩니다
3. **도구 상태**: 이전 대화의 도구 사용 및 결과가 보존됩니다
4. **컨텍스트 복원**: 대화는 이전의 모든 컨텍스트가 그대로 유지된 상태로 재개됩니다

예제:

```bash
# 가장 최근 대화 계속
claude --continue

# 특정 프롬프트로 가장 최근 대화 계속
claude --continue --print "Show me our progress"

# 대화 선택기 표시
claude --resume

# 비대화형 모드에서 가장 최근 대화 계속
claude --continue --print "Run the tests again"
```

</Tip>

---

## Git worktree로 병렬 Claude Code 세션 실행

Claude Code 인스턴스 간에 완전한 코드 격리를 통해 여러 작업을 동시에 수행해야 한다고 가정해 보겠습니다.

<Steps>
  <Step title="Git worktree 이해">
    Git worktree를 사용하면 동일한 저장소에서 여러 브랜치를 별도의 디렉토리로 체크아웃할 수 있습니다. 각 worktree는 격리된 파일이 있는 자체 작업 디렉토리를 가지면서 동일한 Git 기록을 공유합니다. [공식 Git worktree 문서](https://git-scm.com/docs/git-worktree)에서 자세히 알아보세요.
  </Step>

  <Step title="새 worktree 생성">
    ```bash
    # 새 브랜치로 새 worktree 생성 
    git worktree add ../project-feature-a -b feature-a

    # 또는 기존 브랜치로 worktree 생성
    git worktree add ../project-bugfix bugfix-123
    ```

    이것은 저장소의 별도 작업 복사본이 있는 새 디렉토리를 생성합니다.

  </Step>

  <Step title="각 worktree에서 Claude Code 실행">
    ```bash
    # worktree로 이동 
    cd ../project-feature-a

    # 이 격리된 환경에서 Claude Code 실행
    claude
    ```

  </Step>

  <Step title="다른 worktree에서 Claude 실행">
    ```bash
    cd ../project-bugfix
    claude
    ```
  </Step>

  <Step title="worktree 관리">
    ```bash
    # 모든 worktree 나열
    git worktree list

    # 완료되면 worktree 제거
    git worktree remove ../project-feature-a
    ```

  </Step>
</Steps>

<Tip>
  팁:

- 각 worktree는 자체적인 독립적인 파일 상태를 가지므로 병렬 Claude Code 세션에 완벽합니다
- 한 worktree에서 수행된 변경사항은 다른 worktree에 영향을 주지 않아 Claude 인스턴스가 서로 간섭하는 것을 방지합니다
- 모든 worktree는 동일한 Git 기록과 원격 연결을 공유합니다
- 장기 실행 작업의 경우 한 worktree에서 Claude가 작업하는 동안 다른 worktree에서 개발을 계속할 수 있습니다
- 각 worktree가 어떤 작업을 위한 것인지 쉽게 식별할 수 있도록 설명적인 디렉토리 이름을 사용하세요
- 프로젝트의 설정에 따라 각 새 worktree에서 개발 환경을 초기화하는 것을 잊지 마세요. 스택에 따라 다음이 포함될 수 있습니다:
  _ JavaScript 프로젝트: 종속성 설치 실행 (`npm install`, `yarn`)
  _ Python 프로젝트: 가상 환경 설정 또는 패키지 관리자로 설치 \* 기타 언어: 프로젝트의 표준 설정 프로세스 따르기
  </Tip>

---

## Claude를 유닉스 스타일 유틸리티로 사용

### 검증 프로세스에 Claude 추가

Claude Code를 린터나 코드 리뷰어로 사용하고 싶다고 가정해 보겠습니다.

**빌드 스크립트에 Claude 추가:**

```json
// package.json
{
    ...
    "scripts": {
        ...
        "lint:claude": "claude -p 'you are a linter. please look at the changes vs. main and report any issues related to typos. report the filename and line number on one line, and a description of the issue on the second line. do not return any other text.'"
    }
}
```

<Tip>
  팁:

- CI/CD 파이프라인에서 자동화된 코드 리뷰를 위해 Claude를 사용하세요
- 프로젝트와 관련된 특정 문제를 확인하도록 프롬프트를 사용자 정의하세요
- 다양한 유형의 검증을 위해 여러 스크립트를 만드는 것을 고려하세요
  </Tip>

### 파이프 인, 파이프 아웃

Claude에 데이터를 파이프하고 구조화된 형식으로 데이터를 다시 받고 싶다고 가정해 보겠습니다.

**Claude를 통해 데이터 파이프:**

```bash
cat build-error.txt | claude -p 'concisely explain the root cause of this build error' > output.txt
```

<Tip>
  팁:

- 파이프를 사용하여 Claude를 기존 셸 스크립트에 통합하세요
- 강력한 워크플로우를 위해 다른 Unix 도구와 결합하세요
- 구조화된 출력을 위해 --output-format 사용을 고려하세요
  </Tip>

### 출력 형식 제어

특히 Claude Code를 스크립트나 다른 도구에 통합할 때 Claude의 출력을 특정 형식으로 필요로 한다고 가정해 보겠습니다.

<Steps>
  <Step title="텍스트 형식 사용 (기본값)">
    ```bash
    cat data.txt | claude -p 'summarize this data' --output-format text > summary.txt
    ```

    이것은 Claude의 일반 텍스트 응답만 출력합니다 (기본 동작).

  </Step>

  <Step title="JSON 형식 사용">
    ```bash
    cat code.py | claude -p 'analyze this code for bugs' --output-format json > analysis.json
    ```

    이것은 비용과 지속 시간을 포함한 메타데이터가 있는 메시지의 JSON 배열을 출력합니다.

  </Step>

  <Step title="스트리밍 JSON 형식 사용">
    ```bash
    cat log.txt | claude -p 'parse this log file for errors' --output-format stream-json
    ```

    이것은 Claude가 요청을 처리하는 동안 실시간으로 일련의 JSON 객체를 출력합니다. 각 메시지는 유효한 JSON 객체이지만 연결된 경우 전체 출력은 유효한 JSON이 아닙니다.

  </Step>
</Steps>

<Tip>
  팁:

- Claude의 응답만 필요한 간단한 통합에는 `--output-format text`를 사용하세요
- 전체 대화 로그가 필요할 때는 `--output-format json`을 사용하세요
- 각 대화 턴의 실시간 출력에는 `--output-format stream-json`을 사용하세요
  </Tip>

---

## 사용자 정의 슬래시 명령 생성

Claude Code는 특정 프롬프트나 작업을 빠르게 실행하기 위해 생성할 수 있는 사용자 정의 슬래시 명령을 지원합니다.

자세한 내용은 [슬래시 명령](/ko/docs/claude-code/slash-commands) 참조 페이지를 참조하세요.

### 프로젝트별 명령 생성

모든 팀 구성원이 사용할 수 있는 프로젝트용 재사용 가능한 슬래시 명령을 생성하고 싶다고 가정해 보겠습니다.

<Steps>
  <Step title="프로젝트에 명령 디렉토리 생성">
    ```bash
    mkdir -p .claude/commands
    ```
  </Step>

  <Step title="각 명령에 대한 Markdown 파일 생성">
    ```bash
    echo "Analyze the performance of this code and suggest three specific optimizations:" > .claude/commands/optimize.md 
    ```
  </Step>

  <Step title="Claude Code에서 사용자 정의 명령 사용">
    ```
    > /optimize 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- 명령 이름은 파일명에서 파생됩니다 (예: `optimize.md`는 `/optimize`가 됩니다)
- 하위 디렉토리에서 명령을 구성할 수 있습니다 (예: `.claude/commands/frontend/component.md`는 설명에 "(project:frontend)"가 표시된 `/component`를 생성합니다)
- 프로젝트 명령은 저장소를 복제하는 모든 사람이 사용할 수 있습니다
- Markdown 파일 내용은 명령이 호출될 때 Claude에게 전송되는 프롬프트가 됩니다
  </Tip>

### \$ARGUMENTS로 명령 인수 추가

사용자로부터 추가 입력을 받을 수 있는 유연한 슬래시 명령을 생성하고 싶다고 가정해 보겠습니다.

<Steps>
  <Step title="$ARGUMENTS 플레이스홀더가 있는 명령 파일 생성">
    ```bash
    echo 'Find and fix issue #$ARGUMENTS. Follow these steps: 1.
    Understand the issue described in the ticket 2. Locate the relevant code in
    our codebase 3. Implement a solution that addresses the root cause 4. Add
    appropriate tests 5. Prepare a concise PR description' >
    .claude/commands/fix-issue.md 
    ```
  </Step>

  <Step title="이슈 번호와 함께 명령 사용">
    Claude 세션에서 인수와 함께 명령을 사용하세요.

    ```
    > /fix-issue 123
    ```

    이것은 프롬프트에서 \$ARGUMENTS를 "123"으로 바꿉니다.

  </Step>
</Steps>

<Tip>
  팁:

- \$ARGUMENTS 플레이스홀더는 명령 뒤에 오는 모든 텍스트로 바뀝니다
- 명령 템플릿의 어느 곳에나 \$ARGUMENTS를 배치할 수 있습니다
- 기타 유용한 응용: 특정 함수에 대한 테스트 케이스 생성, 구성 요소에 대한 문서 생성, 특정 파일의 코드 검토 또는 지정된 언어로 콘텐츠 번역
  </Tip>

### 개인 슬래시 명령 생성

모든 프로젝트에서 작동하는 개인 슬래시 명령을 생성하고 싶다고 가정해 보겠습니다.

<Steps>
  <Step title="홈 폴더에 명령 디렉토리 생성">
    ```bash
    mkdir -p ~/.claude/commands 
    ```
  </Step>

  <Step title="각 명령에 대한 Markdown 파일 생성">
    ```bash
    echo "Review this code for security vulnerabilities, focusing on:" >
    ~/.claude/commands/security-review.md 
    ```
  </Step>

  <Step title="개인 사용자 정의 명령 사용">
    ```
    > /security-review 
    ```
  </Step>
</Steps>

<Tip>
  팁:

- 개인 명령은 `/help`로 나열될 때 설명에 "(user)"를 표시합니다
- 개인 명령은 본인만 사용할 수 있으며 팀과 공유되지 않습니다
- 개인 명령은 모든 프로젝트에서 작동합니다
- 다양한 코드베이스에서 일관된 워크플로우를 위해 이것들을 사용할 수 있습니다
  </Tip>

---

## Claude의 기능에 대해 Claude에게 묻기

Claude는 자체 문서에 대한 내장 액세스 권한을 가지고 있으며 자체 기능과 제한사항에 대한 질문에 답할 수 있습니다.

### 예제 질문

```
> can Claude Code create pull requests?
```

```
> how does Claude Code handle permissions?
```

```
> what slash commands are available?
```

```
> how do I use MCP with Claude Code?
```

```
> how do I configure Claude Code for Amazon Bedrock?
```

```
> what are the limitations of Claude Code?
```

<Note>
  Claude는 이러한 질문에 대해 문서 기반 답변을 제공합니다. 실행 가능한 예제와 실습 데모는 위의 특정 워크플로우 섹션을 참조하세요.
</Note>

<Tip>
  팁:

- Claude는 사용 중인 버전에 관계없이 항상 최신 Claude Code 문서에 액세스할 수 있습니다
- 자세한 답변을 얻으려면 구체적인 질문을 하세요
- Claude는 MCP 통합, 엔터프라이즈 구성 및 고급 워크플로우와 같은 복잡한 기능을 설명할 수 있습니다
  </Tip>

---

## 다음 단계

<Card title="Claude Code 참조 구현" icon="code" href="https://github.com/anthropics/claude-code/tree/main/.devcontainer">
  개발 컨테이너 참조 구현을 복제하세요.
</Card>
