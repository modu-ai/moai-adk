# SPEC-CH08-001 수용 기준 (Acceptance Criteria)

**SPEC**: Claude Code Plugins & Migration
**버전**: 1.0.0-dev
**검증 일자**: 2025-11-05
**상태**: 검증 대기 중

---

## 📋 수용 기준 체크리스트

### 1. 문서 구성 (9개 섹션)

#### 8-1: Plugin Architecture Overview
- [ ] 플러그인의 정의 명확 (1-2 문장)
- [ ] Output Styles vs Plugins 비교표 포함
- [ ] 플러그인 생명주기 다이어그램 포함
- [ ] 5개 v1.0 플러그인 개요 포함
- [ ] 분량: 900-1100 단어

**검증 방법**:
```
단어 수 확인: wc -w ch08-section-1.md
다이어그램 확인: Mermaid/PlantUML 렌더링 테스트
링크 확인: grep "http" ch08-section-1.md
```

---

#### 8-2: Plugin.json Schema Deep Dive
- [ ] plugin.json 전체 구조 예제 포함
- [ ] 필드별 설명 (metadata, commands, agents, hooks, permissions)
- [ ] 필수(required) vs 선택(optional) 명시
- [ ] 실제 예제 3개 포함 (PM, UI/UX, Backend)
- [ ] 스키마 검증 규칙 설명
- [ ] 분량: 1400-1600 단어

**검증 방법**:
```json
// 예제 1: Minimal Plugin
{
  "id": "test-minimal",
  "name": "Test Minimal",
  "version": "1.0.0",
  "commands": []
}
// 필수 필드만 포함

// 예제 2: Full-Featured Plugin
// 모든 선택 필드 포함

// 예제 3: 실제 v1.0 Plugin
// PM 또는 UI/UX 플러그인
```

---

#### 8-3: Command Development Patterns
- [ ] Command 문법 설명 (Markdown 형식)
- [ ] 인자 처리 (required, optional, flags) 예제
- [ ] 명령어 실행 흐름도 포함
- [ ] `/init-pm` 상세 예제 포함
- [ ] 출력 형식 정의
- [ ] 분량: 950-1050 단어

**검증 방법**:
```bash
# 실제 명령어 문법이 유효한지 확인
/init-pm my-project        # required arg
/init-pm my-project --skip-charter  # optional flag
```

---

#### 8-4: Agent Orchestration
- [ ] Agent의 역할 명확 (command 실행의 핵심)
- [ ] agent.md 템플릿 예제 포함
- [ ] Tool 권한 제어 방식 설명
- [ ] Sub-agent 호출 패턴 예제
- [ ] 여러 Agent 간 조율 예제
- [ ] 분량: 1150-1250 단어

**검증 방법**:
```markdown
# Agent Invocation Flow

User Input: /init-pm my-project
↓
Command Handler: Parse arguments
↓
Agent Invocation: Task(subagent_type="pm-agent")
↓
Agent Execution: Generate templates
↓
Output: Generated files + report
```

---

#### 8-5: Hook Lifecycle & Events
- [ ] 4가지 Hook 타입 설명 (SessionStart, PreToolUse, PostToolUse, SessionEnd)
- [ ] Hook 실행 순서도 포함
- [ ] 각 Hook의 timeout 및 priority 설명
- [ ] 실제 Hook 구현 예제 3개
- [ ] 조건부 Hook 실행 (conditions 필드)
- [ ] 분량: 1150-1250 단어

**검증 방법**:
```javascript
// 테스트 가능한 Hook 예제
{
  "preToolUse": {
    "name": "enforceDeniedTools",
    "priority": 100,
    "timeout": 1000
  }
}
// timeout이 0 < timeout <= 5000 범위인지 확인
```

---

#### 8-6: Skill Integration for Plugins
- [ ] Skill 정의 (재사용 가능한 지식)
- [ ] 플러그인 Skill 구조 (SKILL.md, examples.md, reference.md)
- [ ] Agent에서 Skill 로딩 방식
- [ ] 예제: `moai-plugin-scaffolding` Skill
- [ ] 예제: `moai-plugin-testing-patterns` Skill
- [ ] 분량: 950-1050 단어

**검증 방법**:
```bash
# Skill 파일 구조 확인
.claude/skills/moai-plugin-scaffolding/
├── SKILL.md (400-500 단어)
├── examples.md (코드 예제 3개)
└── reference.md (API 레퍼런스)
```

---

#### 8-7: Permission Model & Security
- [ ] Deny-by-Default 원칙 설명
- [ ] allowedTools vs deniedTools 예제
- [ ] 최소 권한 (Least Privilege) 적용
- [ ] 런타임 권한 검증 (PreToolUse Hook)
- [ ] 보안 모범 사례 5개 이상
- [ ] 분량: 950-1050 단어

**검증 방법**:
```json
// 안전한 권한 설정
{
  "permissions": {
    "allowedTools": ["Read", "Write"],
    "deniedTools": ["Bash", "DeleteFile"]
  }
}
// - 쓰기 권한이 필요한가?
// - 실행 권한이 필요한가?
// - 파일 삭제 권한이 필요한가?
```

---

#### 8-8: Migration Path (Output Styles → Plugins)
- [ ] Output Styles의 역사 및 한계 설명
- [ ] v1.0에서 Plugin으로 이행 이유
- [ ] 단계별 마이그레이션 체크리스트
- [ ] Output Style → Plugin 변환 예제 2개
- [ ] 호환성 문제 및 해결책
- [ ] 분량: 750-850 단어

**검증 방법**:
```markdown
# Migration Checklist

- [ ] Remove `.claude/output-style.json`
- [ ] Create plugin directories
- [ ] Define plugin.json
- [ ] Migrate styling to skills
- [ ] Update settings.json
- [ ] Test plugin installation

모두 체크되어야 마이그레이션 완료
```

---

#### 8-9: FAQ & Troubleshooting
- [ ] 최소 13개 Q&A 포함
  - 설치/제거 FAQ (5개)
  - 명령어 개발 FAQ (3개)
  - Hook 개발 FAQ (2개)
  - 보안/권한 FAQ (3개)
- [ ] 각 Q&A에 해결책 포함
- [ ] 실패 사례 + 해결책 최소 3개
- [ ] 분량: 450-550 단어

**검증 방법**:
```markdown
Q: How do I install a plugin?
A: [명확한 답변 + 예제 + 스크린샷]

Q: Why is my command not executing?
A: [트러블슈팅 단계]

Q: Can I override permissions?
A: [보안 정책 설명]
```

---

### 2. Hands-on Labs (4개 랩)

#### Lab 8A: 간단한 Command-Only Plugin
- [ ] 초보자도 5분 내 완료 가능
- [ ] 단계별 명시적 지시사항
- [ ] 예상되는 출력 샘플 제공
- [ ] 성공 기준 명확 (플러그인 설치 및 명령어 실행)
- [ ] 분량: 450-550 단어

**검증 기준**:
```bash
# Lab 완료 후
$ /hello-world
"Hello, Plugin!"
# 출력이 일치하면 ✅ Pass
```

---

#### Lab 8B: Agent가 포함된 Plugin
- [ ] 초보자도 10분 내 완료 가능
- [ ] Agent 조율 로직 명시
- [ ] Skill 통합 예제 포함
- [ ] 성공 기준 명확 (SPEC 파일 생성)
- [ ] 분량: 600-700 단어

**검증 기준**:
```bash
# Lab 완료 후
$ /generate-spec test-project
# .moai/specs/SPEC-TEST-001/ 디렉토리 생성되면 ✅ Pass
```

---

#### Lab 8C: Hook 등록 및 테스트
- [ ] 초보자도 8분 내 완료 가능
- [ ] Hook 실행 순서 확인 가능
- [ ] SessionStart 및 PreToolUse Hook 포함
- [ ] 성공 기준 명확 (Hook 실행 로그)
- [ ] 분량: 550-650 단어

**검증 기준**:
```bash
# Lab 완료 후
# SessionStart Hook 실행 로그 확인
[SessionStart] Plugin initialized
[PreToolUse] Checking permissions...
# 로그가 나타나면 ✅ Pass
```

---

#### Lab 8D: 권한 모델 구현
- [ ] 초보자도 10분 내 완료 가능
- [ ] allowedTools/deniedTools 설정
- [ ] 권한 위반 시나리오 테스트
- [ ] 성공 기준 명확 (권한 검증 동작)
- [ ] 분량: 550-650 단어

**검증 기준**:
```bash
# Lab 완료 후
# 허용된 도구는 실행 가능
/read-file config.json  # ✅ Pass

# 금지된 도구는 실행 불가
/delete-file config.json  # ❌ Blocked
# "Tool 'DeleteFile' denied" 메시지 출력 → ✅ Pass
```

---

### 3. 예제 및 자료

#### 코드 예제 (최소 15개)
- [ ] plugin.json 예제 3개
- [ ] command.md 템플릿 2개
- [ ] agent.md 예제 2개
- [ ] hooks.json 예제 2개
- [ ] Skill 구조 예제 2개
- [ ] 권한 설정 예제 2개

**검증 방법**:
```bash
# 모든 코드 예제가 유효한 JSON/JavaScript/Markdown인지 확인
jq . plugin.json  # JSON 유효성
grep "^#" agent.md | wc -l  # Markdown 헤더 확인
```

---

#### 다이어그램 (최소 8개)
- [ ] 플러그인 아키텍처 다이어그램 1개
- [ ] 플러그인 생명주기 다이어그램 1개
- [ ] 명령어 실행 흐름도 1개
- [ ] Agent 조율 흐름도 1개
- [ ] Hook 실행 순서도 1개
- [ ] 권한 모델 다이어그램 1개
- [ ] 마이그레이션 흐름 다이어그램 1개
- [ ] 비교표 (Output Styles vs Plugins) 1개

**검증 방법**:
```markdown
모든 다이어그램이 다음 중 하나로 표현되어야 함:
- Mermaid (markdown에 포함 가능)
- PlantUML (uml 형식)
- ASCII Art (마크다운 코드 블록)
- 스크린샷 (실제 실행 결과)
```

---

#### 스크린샷 (최소 4개)
- [ ] Plugin 설치 과정 스크린샷
- [ ] 명령어 실행 결과 스크린샷
- [ ] Hook 실행 로그 스크린샷
- [ ] 오류 메시지 예제 스크린샷

**검증 방법**:
```bash
# 모든 스크린샷이 최신 CLI/UI를 반영하고 있는지 확인
# 가독성이 좋은 크기 (800x600 이상)인지 확인
```

---

### 4. 품질 기준

#### 문법 및 철자
- [ ] 오타 0개 (맞춤법 검사)
- [ ] 일관된 용어 사용
  - ✅ moai-alfred (O)
  - ❌ moai-cc (X)
  - ✅ Plugin (O)
  - ❌ plugin (X - 문장 시작이 아닌 경우)
- [ ] 문장 구조 명확

**검증 방법**:
```bash
# 철자 검사
aspell check ch08.md

# moai-cc 참조 확인 (0개여야 함)
grep -i "moai-cc" ch08*.md | wc -l
```

---

#### 링크 및 참조
- [ ] 모든 링크 작동 (404 없음)
- [ ] 내부 링크 일관성 (다른 SPEC 참조)
- [ ] 외부 링크 신뢰성 (공식 문서만)

**검증 방법**:
```bash
# 링크 유효성 검사
curl -s https://example.com 2>&1 | head -1

# 내부 링크 확인
grep "\[.*\](.*SPEC.*)" ch08*.md | wc -l
```

---

#### 일관성
- [ ] 용어 통일 (Plugin vs Plugin, 플러그인)
- [ ] 포맷 통일 (코드 블록, 테이블, 리스트)
- [ ] 참조 스타일 통일

**검증 방법**:
```bash
# 용어 빈도 확인
grep -io "Plugin\|플러그인" ch08*.md | sort | uniq -c

# 테이블 형식 확인
grep "^|" ch08*.md | head -5
```

---

### 5. 최종 검증 체크리스트

#### 문서 통계
```
📊 전체 분량: 9,800 단어 ± 200단어 범위
- 8-1: 1,050 단어 ✓
- 8-2: 1,500 단어 ✓
- 8-3: 1,000 단어 ✓
- 8-4: 1,200 단어 ✓
- 8-5: 1,200 단어 ✓
- 8-6: 1,000 단어 ✓
- 8-7: 1,000 단어 ✓
- 8-8: 800 단어 ✓
- 8-9: 500 단어 ✓
- Labs: 2,400 단어 ✓
```

#### 실행 요소
```
📊 총 구성 요소: 20개 이상

코드 예제:
✓ plugin.json: 3개
✓ command.md: 2개
✓ agent.md: 2개
✓ hooks.json: 2개
✓ 권한 설정: 2개
✓ 마이그레이션: 2개

다이어그램: 8개
스크린샷: 4개
Lab: 4개
FAQ: 13개
```

---

## ✅ 승인 기준 (Go/No-Go)

### GO 조건
- ✅ 모든 섹션 완성 (8-1 ~ 8-9)
- ✅ 모든 Lab 실행 가능 (8A-D)
- ✅ 분량 9,600 ~ 10,000 단어
- ✅ 코드 예제 15개 이상
- ✅ 다이어그램 8개 이상
- ✅ 오타/문법 오류 0개
- ✅ 링크 100% 작동
- ✅ 용어 일관성 100%

### NO-GO 조건
- ❌ 섹션 미완성
- ❌ Lab 실행 불가
- ❌ 분량 부족 (<9,600 단어)
- ❌ 오타/문법 오류 5개 이상
- ❌ 링크 미작동 (404)
- ❌ moai-cc 참조 존재
- ❌ 코드 예제 테스트 실패

---

## 🔍 검증 담당자

| 역할 | 담당자 | 검증 내용 |
|------|--------|---------|
| 구조/내용 | Alfred SuperAgent | 섹션 완성도, 분량, 일관성 |
| 실행성 | QA Team | Lab 실행 가능성, 예제 유효성 |
| 품질 | Editor | 오타, 문법, 용어 통일 |
| 기술 | Tech Review | 코드 정확성, 최신 기술 반영 |

---

## 📅 검증 일정

| 날짜 | 검증 내용 | 승인 담당 |
|------|---------|---------|
| 2025-11-05 | 초안 완성 및 자체 검수 | Alfred |
| 2025-11-05 | Lab 실행 테스트 | QA |
| 2025-11-06 | 최종 검수 및 승인 | Tech Review |

---

**검증 완료 시 상태**: ACCEPTED ✅
**릴리스 준비**: READY FOR CH08 CHAPTER WRITING
