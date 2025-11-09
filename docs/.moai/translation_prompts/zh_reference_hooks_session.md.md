Translate the following Korean markdown document to Chinese (Simplified).

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/hooks/session.md
**Target Language:** Chinese (Simplified)
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/zh/reference/hooks/session.md

**Content to Translate:**

# SessionStart Hook 상세 가이드

Claude Code 세션 시작 시 자동으로 실행되는 Hook입니다.

## :bullseye: 목적

세션 시작 시 다음을 자동으로 수행합니다:

- ✅ 프로젝트 상태 확인
- ✅ 세션 로그 분석
- ✅ 설정 검증
- ✅ 의존성 상태 확인
- ✅ Git 저장소 상태 확인

## 📝 실행 내용

```bash
#!/bin/bash
# SessionStart Hook

# 1. 프로젝트 메타데이터 로드
config=$(cat .moai/config.json)
project_name=$(echo $config | jq -r '.project.name')

# 2. 의존성 확인
python3 --version
uv --version
git --version

# 3. Git 상태 확인
current_branch=$(git branch --show-current)
commits_ahead=$(git rev-list --count HEAD..origin/main)

# 4. 세션 로그 분석
last_session=$(ls -t ~/.claude/projects/*/session-*.json | head -1)
# 로그 분석 결과 출력

# 5. 문제 감지 시 경고
if [ "$commits_ahead" -gt 10 ]; then
    echo "⚠️ 메인에서 10개 이상 커밋 앞서있습니다"
fi
```

## :mag: 세션 로그 분석

SessionStart Hook은 이전 세션의 로그를 분석합니다:

### 분석 항목

| 항목          | 분석 내용             |
| ------------- | --------------------- |
| **Tool 사용** | 가장 자주 사용된 도구 |
| **오류 패턴** | 반복되는 오류         |
| **성능**      | 평균 실행 시간        |
| **효율성**    | 성공률                |

### 분석 결과

```
📊 MoAI-ADK 세션 메타분석

Tool 사용 TOP 5:
1. Bash (git status) - 45회
2. Read (파일 읽기) - 38회
3. Edit (파일 수정) - 22회
4. Grep (검색) - 18회
5. Write (파일 작성) - 15회

⚠️ 오류 패턴:
- "File not found": 3회
- "Permission denied": 1회

:bullseye: 개선 제안:
- Glob 사용으로 파일 검색 효율화
- 경로 확인 후 작업 수행
```

## ⚙️ 설정 검증

```bash
# .moai/config.json 검증
- 프로젝트 메타데이터 확인
- 언어 설정 확인
- TRUST 5 원칙 설정 확인

# .claude/settings.json 검증
- Hook 활성화 상태
- 권한 설정 일관성
- MCP 서버 설정
```

## 📋 문제 감지

SessionStart Hook이 감지하는 문제들:

```
:x: 심각한 문제 (작업 중단)
- .moai/ 디렉토리 손상
- config.json 파싱 오류
- Git 저장소 손상

⚠️ 경고 (진행하되 주의)
- 10개 이상 커밋 앞서있음
- 테스트 미통과 파일 있음
- 커버리지 목표 미달

💡 정보성 메시지
- 새 버전 출시
- 추천 설정 변경
- 성능 최적화 제안
```

## 🔄 Hook 체인

```
SessionStart
    ├─→ config.json 로드
    ├─→ 의존성 확인
    ├─→ Git 상태 확인
    ├─→ 세션 로그 분석
    └─→ 문제 감지 및 보고
         └─→ 심각 → 중단
         └─→ 경고 → 진행 + 메시지
         └─→ 정보 → 진행 + 팁
```

______________________________________________________________________

**다음**: [Tool Hooks](tool.md) 또는 [Hooks 개요](index.md)


**Instructions:**
- Translate the content above to Chinese (Simplified)
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
