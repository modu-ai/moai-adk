# SPEC-PLUGIN-002: 구현 계획

## 📋 개요

본 문서는 `/alfred:8-project` 커맨드 강화를 위한 상세한 구현 계획을 정의합니다.

---

## 🎯 우선순위별 마일스톤

### 1차 목표: 기본 템플릿 복사 및 렌더링

- **scripts/render-template.cjs 스크립트 작성**
  - Mustache 템플릿 렌더링 엔진 구현
  - 변수 매핑 및 파일 출력 로직
  - 에러 핸들링 및 로깅

- **Phase 1: 환경 분석 강화**
  - `.moai/` 디렉토리 존재 여부 확인
  - `${CLAUDE_PLUGIN_ROOT}` 환경변수 검증
  - Node.js 설치 여부 확인

- **Phase 2: 템플릿 복사 구현**
  - `${CLAUDE_PLUGIN_ROOT}/templates/.moai` → `.moai` 복사
  - Mustache 파일 탐지 및 렌더링 실행
  - 렌더링 완료 후 `.mustache` 파일 정리

### 2차 목표: Git 초기화 및 .gitignore 생성

- **Git 초기화 프롬프트**
  - 사용자 확인 메시지 표시
  - `git init` 실행
  - `.gitignore` 템플릿 복사 또는 생성

- **프로젝트 문서 생성/갱신**
  - product.md, structure.md, tech.md 생성
  - 기존 문서가 있을 경우 보존 전략 적용

### 3차 목표: 에러 처리 및 검증

- **경로 검증**
  - `${CLAUDE_PLUGIN_ROOT}` 유효성 확인
  - 템플릿 디렉토리 존재 확인
  - 쓰기 권한 검증

- **변수 검증**
  - PROJECT_NAME 정규화 (kebab-case)
  - PROJECT_MODE 값 검증 (personal|team)
  - LOCALE 값 검증 (ko|en|ja|zh)

- **롤백 전략**
  - 초기화 실패 시 부분 복사 파일 정리
  - 백업 복원 메커니즘 (선택적)

---

## 🛠️ 기술적 접근 방법

### 1. render-template.cjs 스크립트 설계

**파일 경로**: `scripts/render-template.cjs`

**핵심 기능**:
```javascript
// 1. 템플릿 파일 읽기
// 2. Mustache 렌더링
// 3. 출력 파일 작성
// 4. 에러 핸들링

const Mustache = require('mustache');
const fs = require('fs');
const path = require('path');

function renderTemplate(templatePath, outputPath, variables) {
  // 템플릿 파일 읽기
  const template = fs.readFileSync(templatePath, 'utf-8');

  // Mustache 렌더링
  const rendered = Mustache.render(template, variables);

  // 출력 파일 작성
  fs.writeFileSync(outputPath, rendered, 'utf-8');

  console.log(`✅ Rendered: ${outputPath}`);
}
```

**실행 인터페이스**:
```bash
node scripts/render-template.cjs \
  --template .moai/config.json.mustache \
  --output .moai/config.json \
  --var PROJECT_NAME="MoAI-ADK" \
  --var PROJECT_MODE="team" \
  --var PROJECT_DESCRIPTION="MoAI Agentic Development Kit" \
  --var LOCALE="ko"
```

**의존성**:
```json
{
  "dependencies": {
    "mustache": "^4.2.0"
  }
}
```

### 2. Mustache 변수 매핑 테이블

| 변수명                  | 기본값           | 설명                        | 예시                    |
| ----------------------- | ---------------- | --------------------------- | ----------------------- |
| `{{PROJECT_NAME}}`      | 현재 디렉토리명  | 프로젝트 이름               | "MoAI-ADK"              |
| `{{PROJECT_MODE}}`      | "personal"       | personal 또는 team          | "team"                  |
| `{{PROJECT_DESCRIPTION}}` | ""              | 프로젝트 설명 (한 줄)       | "SPEC-First TDD Framework" |
| `{{LOCALE}}`            | "ko"             | 언어 설정 (ko, en, ja, zh)  | "ko"                    |

**변수 수집 전략**:
1. 현재 디렉토리명으로 PROJECT_NAME 추론
2. Git 감지 결과로 PROJECT_MODE 제안 (Git 있으면 team, 없으면 personal)
3. package.json 또는 README.md에서 PROJECT_DESCRIPTION 추출
4. 시스템 언어 설정으로 LOCALE 추론

### 3. 커맨드 로직 흐름도

```
/alfred:8-project 실행
  ↓
Phase 1: 환경 분석
  ├─→ .moai/ 존재 확인
  ├─→ ${CLAUDE_PLUGIN_ROOT} 검증
  ├─→ Node.js 설치 확인
  └─→ Git 설치 확인
  ↓
사용자 확인 ("진행", "수정", "중단")
  ↓
Phase 2-1: 템플릿 복사
  ├─→ mkdir -p .moai/project .moai/memory .moai/specs
  ├─→ cp -r ${CLAUDE_PLUGIN_ROOT}/templates/.moai/* .moai/
  └─→ 복사 완료 확인
  ↓
Phase 2-2: 변수 수집
  ├─→ PROJECT_NAME 입력/추론
  ├─→ PROJECT_MODE 선택 (personal/team)
  ├─→ PROJECT_DESCRIPTION 입력
  └─→ LOCALE 선택 (ko/en/ja/zh)
  ↓
Phase 2-3: Mustache 렌더링
  ├─→ config.json.mustache → config.json
  ├─→ CLAUDE.md.mustache → CLAUDE.md (선택적)
  └─→ .mustache 파일 정리
  ↓
Phase 2-4: Git 초기화 (선택적)
  ├─→ 사용자 확인: "Git 초기화 하시겠습니까?"
  ├─→ git init (yes일 경우)
  ├─→ .gitignore 생성/복사
  └─→ 초기 커밋 (선택적)
  ↓
최종 보고
  ├─→ 생성된 파일 목록 표시
  ├─→ 다음 단계 안내 (/alfred:1-spec)
  └─→ 완료 메시지
```

---

## 🏗️ 아키텍처 설계 방향

### 1. 플러그인 경로 체계

**환경변수 활용**:
- `${CLAUDE_PLUGIN_ROOT}`: Claude Code가 자동 설정하는 플러그인 루트 경로
- 예: `/Users/user/.claude/plugins/moai-adk`

**템플릿 디렉토리 구조**:
```
${CLAUDE_PLUGIN_ROOT}/
  templates/
    .moai/
      config.json.mustache      # 필수 렌더링 대상
      CLAUDE.md.mustache        # 선택적 렌더링 대상
      project/
        product.md
        structure.md
        tech.md
      memory/
        development-guide.md
        spec-metadata.md
      specs/
        (빈 디렉토리)
      reports/
        (빈 디렉토리)
```

### 2. 에러 처리 시나리오

**시나리오 1: ${CLAUDE_PLUGIN_ROOT} 미설정**
```
❌ CRITICAL: 플러그인 루트 경로를 찾을 수 없습니다
  → ${CLAUDE_PLUGIN_ROOT} 환경변수가 설정되지 않았습니다
  → 해결: Claude Code 플러그인 설치 상태를 확인하세요
```

**시나리오 2: .moai/ 디렉토리 이미 존재**
```
⚠️ WARNING: .moai/ 디렉토리가 이미 존재합니다
  → 갱신 모드로 전환합니다
  → 기존 product.md, structure.md, tech.md는 보존됩니다
```

**시나리오 3: Node.js 미설치**
```
⚠️ WARNING: Node.js가 설치되지 않았습니다
  → Mustache 렌더링을 건너뜁니다
  → 수동 렌더링 가이드:
    1. Node.js 설치: https://nodejs.org
    2. npm install mustache
    3. node scripts/render-template.cjs --help
```

**시나리오 4: 템플릿 파일 누락**
```
❌ CRITICAL: 필수 템플릿 파일이 없습니다
  → ${CLAUDE_PLUGIN_ROOT}/templates/.moai/config.json.mustache
  → 해결: 플러그인을 재설치하거나 템플릿 파일을 수동으로 복사하세요
```

**시나리오 5: 특수문자 프로젝트명**
```
⚠️ WARNING: 프로젝트명에 특수문자가 포함되어 있습니다
  → 입력: "My Project (2025)"
  → 정규화: "my-project-2025"
  → 계속 진행하시겠습니까? (y/n)
```

---

## ⚠️ 리스크 및 대응 방안

### 리스크 1: Claude가 Mustache를 직접 렌더링할 수 없음

**영향도**: High
**발생 확률**: 확정
**대응 방안**:
- render-template.cjs Node.js 스크립트로 렌더링 위임
- Bash 도구로 `node scripts/render-template.cjs` 실행
- Node.js 미설치 시 대체 전략: 기본값으로 config.json 생성

### 리스크 2: 템플릿 파일 누락 또는 손상

**영향도**: Medium
**발생 확률**: Low
**대응 방안**:
- Phase 1에서 템플릿 파일 무결성 검증
- 누락 시 기본 템플릿 문자열로 대체
- 사용자에게 플러그인 재설치 안내

### 리스크 3: 권한 에러 (.moai/ 디렉토리 생성 실패)

**영향도**: High
**발생 확률**: Low
**대응 방안**:
- `mkdir -p .moai/` 실행 전 쓰기 권한 확인
- 권한 에러 시 `chmod` 명령어 안내
- 사용자 홈 디렉토리 대체 경로 제안

### 리스크 4: 기존 파일 덮어쓰기 충돌

**영향도**: Medium
**발생 확률**: Medium
**대응 방안**:
- 기존 파일 존재 시 백업 생성 (예: `config.json.backup-YYYYMMDD-HHMMSS`)
- 사용자에게 덮어쓰기 여부 확인 프롬프트 표시
- "Legacy Context" 섹션에 기존 내용 보존

---

## 🔍 의존성

### 내부 의존성

- **@SPEC:PLUGIN-001**: Claude Code 플러그인 구조 설계 (필수)
- **project-manager 에이전트**: 프로젝트 문서 생성 로직 재사용

### 외부 의존성

- **Node.js**: render-template.cjs 실행 환경
- **mustache npm 패키지**: 템플릿 렌더링 라이브러리
- **Git**: 선택적 Git 초기화 기능

---

## 📊 성공 기준

### 기능 테스트

- [ ] 신규 프로젝트에서 `.moai/` 디렉토리 생성 확인
- [ ] config.json.mustache → config.json 렌더링 성공
- [ ] 변수 치환 정확성 (PROJECT_NAME, PROJECT_MODE, PROJECT_DESCRIPTION, LOCALE)
- [ ] Git 초기화 선택적 실행 확인
- [ ] 에러 시나리오 처리 (Node.js 미설치, 템플릿 누락 등)

### 품질 기준

- [ ] render-template.cjs 스크립트 테스트 커버리지 ≥85%
- [ ] 에러 메시지 명확성 검증
- [ ] 커맨드 실행 시간 ≤5초

### 사용성 기준

- [ ] 사용자 입력 최소화 (기본값 제공)
- [ ] 진행 상태 실시간 표시
- [ ] 완료 메시지 및 다음 단계 안내 제공

---

## 🚀 다음 단계

1. **render-template.cjs 스크립트 작성** (@CODE:PLUGIN-002:SCRIPT)
2. **8-project.md 커맨드 로직 강화** (@CODE:PLUGIN-002)
3. **테스트 케이스 작성** (@TEST:PLUGIN-002)
4. **문서 동기화** (@DOC:PLUGIN-002)

---

_이 문서는 SPEC-PLUGIN-002의 구현 가이드로, `/alfred:2-build SPEC-PLUGIN-002` 실행 시 참조됩니다._
