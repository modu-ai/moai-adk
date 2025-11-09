# Skills Integration Guide for Agents

**Context7-Based Skills 통합 가이드**

> **Updated**: 2025-11-10
> **Version**: 1.0.0
> **Target Agents**: ui-ux-expert, frontend-expert

---

## 📋 생성된 Skills 목록

### 1. moai-icons-vector (NEW!)
**담당**: frontend-expert (주), ui-ux-expert (보조)

**호출 시점**:
- SPEC에서 `icon`, `vector icon`, `lucide`, `heroicons`, `radix icon` 키워드 감지
- 아이콘 선택 및 통합 필요
- 아이콘 컴포넌트 설계 필요
- 접근성 아이콘 구현 필요

**Skill 호출 예시**:
```python
# frontend-expert 에이전트에서
Skill("moai-icons-vector")

# 아이콘 사용 필요 시 자동 로드
# "icon button", "vector icon", "lucide" 감지 시 자동 로드
```

**제공 콘텐츠**:
- Lucide (1000+ icons), Heroicons (300+ icons), Radix Icons (150+ icons) 비교
- 각 라이브러리 설치 및 기본 사용법
- Icon button 컴포넌트 (TypeScript)
- 동적 아이콘 로딩 패턴
- 접근성 (ARIA labels, semantic HTML)
- 애니메이션 효과
- 아이콘 테마 시스템
- 성능 최적화 (tree-shaking, 동적 import)
- 실제 React 예제 (TSX)

---

### 2. moai-lang-html-css
**담당**: ui-ux-expert (주), frontend-expert (보조)

**호출 시점**:
- SPEC에서 `semantic`, `html`, `accessibility`, `a11y`, `wcag`, `form`, `navigation` 키워드 감지
- HTML 구조 설계 필요
- 접근성(WCAG 2.1 AA) 구현 가이드 필요
- 시맨틱 마크업 검증 필요

**Skill 호출 예시**:
```python
# Alfred 또는 ui-ux-expert 에이전트에서
Skill("moai-lang-html-css")

# 특정 주제 포함 호출
Skill("moai-lang-html-css")
# 사용자 프롬프트에 "semantic HTML", "accessibility", "WCAG" 포함 시 자동 로드
```

**제공 콘텐츠**:
- 시맨틱 HTML5 요소 레퍼런스
- WCAG 2.1 AA 접근성 체크리스트
- 폼 설계 패턴
- 반응형 CSS 설계
- 디자인 토큰 CSS 변수 설정
- 포커스 관리 및 키보드 네비게이션
- 실제 예제 코드

---

### 2. moai-lang-tailwind-css
**담당**: frontend-expert (주), ui-ux-expert (보조)

**호출 시점**:
- SPEC에서 `tailwind`, `utility-first`, `css framework`, `styling`, `responsive` 키워드 감지
- Tailwind CSS 설정 필요
- 디자인 토큰 구현 필요
- 성능 최적화 (PurgeCSS, 번들 최소화) 필요

**Skill 호출 예시**:
```python
# frontend-expert 에이전트에서
Skill("moai-lang-tailwind-css")

# tailwind.config.js 생성 시 자동 로드
# Core Web Vitals 최적화 필요 시 자동 로드
```

**제공 콘텐츠**:
- Tailwind CSS v4.0+ 최신 설정
- 디자인 토큰 구현 (colors, spacing, typography)
- 반응형 그리드 및 레이아웃 패턴
- Dark mode 구현
- 커스텀 variants 및 plugins
- 성능 최적화 전략
- 모바일-퍼스트 디자인 패턴
- 실제 React 컴포넌트 예제

---

### 3. moai-lib-shadcn-ui
**담당**: frontend-expert (주), ui-ux-expert (보조)

**호출 시점**:
- SPEC에서 `shadcn`, `shadcn/ui`, `component library`, `radix ui` 키워드 감지
- React 컴포넌트 아키텍처 설계 필요
- UI 컴포넌트 구현 필요
- Tailwind + React 통합 필요

**Skill 호출 예시**:
```python
# frontend-expert 에이전트에서
Skill("moai-lib-shadcn-ui")

# React 프로젝트 with Tailwind 시 자동 로드
# 컴포넌트 설계 단계 자동 로드
```

**제공 콘텐츠**:
- shadcn/ui v2.0+ 설치 및 설정
- 20+ 컴포넌트 사용 패턴 (Button, Card, Dialog, Form, etc.)
- TypeScript 타입 안전성
- Radix UI 접근성 패턴
- 폼 검증 및 에러 처리
- 데이터 테이블 구현
- 커스텀 컴포넌트 composition (asChild)
- Dark mode 지원
- 실제 React 예제 (TSX)

---

## 🎯 에이전트별 Skill 통합 전략

### ui-ux-expert 에이전트

**Skill 자동 호출 트리거**:
```python
# 사용자 요청에서 다음 키워드 감지 시:
keywords = [
    "html", "semantic", "accessibility", "a11y", "wcag",
    "form", "navigation", "landmark", "aria", "keyboard",
    "focus", "tab order", "color contrast", "skip link"
]

if any(keyword in user_request.lower() for keyword in keywords):
    Skill("moai-lang-html-css")
```

**Skill 사용 시나리오**:

1. **SPEC 분석 단계**
   ```
   사용자: "대시보드 UI 설계 필요 (WCAG 2.1 AA 준수)"
   ui-ux-expert: Skill("moai-lang-html-css") 호출
                  → 접근성 체크리스트 제공
                  → 시맨틱 마크업 구조 제시
   ```

2. **접근성 검증**
   ```
   사용자: "폼 접근성을 WCAG 2.1 AA로 검증하고 싶음"
   ui-ux-expert: Skill("moai-lang-html-css") 호출
                  → 폼 레이블 연결 패턴
                  → 에러 메시지 표시 패턴
                  → 색상 대비 검증 가이드
   ```

3. **디자인 시스템 수립**
   ```
   사용자: "HTML/CSS 기반 디자인 시스템 구축"
   ui-ux-expert: Skill("moai-lang-html-css") 호출
                  → CSS 변수 설정 예제
                  → 시맨틱 HTML 구조
                  → 반응형 디자인 패턴
   ```

---

### frontend-expert 에이전트

**Skill 자동 호출 트리거**:
```python
# 사용자 요청에서 다음 키워드 감지 시:
tailwind_keywords = [
    "tailwind", "utility-first", "responsive", "tailwind css",
    "design tokens", "custom config", "plugins", "dark mode",
    "performance", "purge", "bundle size"
]

shadcn_keywords = [
    "shadcn", "shadcn/ui", "component library", "radix ui",
    "button", "card", "dialog", "form", "data table",
    "accessible components", "react components"
]

icon_keywords = [
    "icon", "icons", "vector icon", "lucide", "heroicons",
    "radix icons", "icon button", "icon library", "svg icons"
]

if any(keyword in user_request.lower() for keyword in tailwind_keywords):
    Skill("moai-lang-tailwind-css")

if any(keyword in user_request.lower() for keyword in shadcn_keywords):
    Skill("moai-lib-shadcn-ui")

if any(keyword in user_request.lower() for keyword in icon_keywords):
    Skill("moai-icons-vector")
```

**Skill 사용 시나리오**:

1. **Tailwind CSS 프로젝트 설정**
   ```
   사용자: "React + Tailwind 프로젝트 초기화"
   frontend-expert: Skill("moai-lang-tailwind-css") 호출
                     → tailwind.config.js 설정
                     → 디자인 토큰 구성
                     → CSS 변수 연동
   ```

2. **shadcn/ui 컴포넌트 구현**
   ```
   사용자: "로그인 폼 컴포넌트 구현 (shadcn/ui 사용)"
   frontend-expert: Skill("moai-lib-shadcn-ui") 호출
                     → 설치 방법 및 설정
                     → Form, Input, Button 컴포넌트
                     → 검증 및 에러 처리 패턴
   ```

3. **성능 최적화**
   ```
   사용자: "Tailwind CSS 번들 크기 최적화"
   frontend-expert: Skill("moai-lang-tailwind-css") 호출
                     → PurgeCSS 설정
                     → 동적 클래스명 피하기
                     → 번들 분석 도구
   ```

4. **아이콘 선택 및 구현**
   ```
   사용자: "로그인 폼에 아이콘 추가 (Lucide 또는 Heroicons?)"
   frontend-expert: Skill("moai-icons-vector") 호출
                     → Lucide vs Heroicons 비교
                     → Icon button 컴포넌트 패턴
                     → 접근성 (aria-label) 구현
   ```

5. **접근성 준수**
   ```
   사용자: "shadcn/ui 컴포넌트 접근성 검증"
   frontend-expert: Skill("moai-lib-shadcn-ui") 호출 (보조)
                     Skill("moai-lang-html-css") 호출 (ui-ux-expert 연동)
                     → Radix UI 접근성
                     → WCAG 2.1 AA 준수
   ```

---

## 📝 에이전트 프롬프트 통합 방법

### 방법 1: Task 호출 시 자동 포함

에이전트가 Task로 호출될 때, 프롬프트에 다음을 추가:

```python
Task(
    subagent_type="ui-ux-expert",
    prompt="""
    디자보드 UI 설계를 시작하겠습니다.

    다음 Skills을 활용하세요:
    - Skill("moai-lang-html-css") - 시맨틱 HTML 및 접근성 가이드
    - 필요시 Skill("moai-lang-tailwind-css") - 스타일링

    WCAG 2.1 AA 접근성을 만족하는 시맨틱 마크업을 제공해주세요.
    """
)
```

### 방법 2: 에이전트 내부 자동 로드

Alfred 또는 에이전트가 사용자 요청을 분석하여 자동 로드:

```python
# ui-ux-expert 에이전트 내부 로직
user_request = "accessible form design with WCAG 2.1 AA"

if "accessible" in user_request and "wcag" in user_request:
    # Skill 자동 호출
    Skill("moai-lang-html-css")
    # → 접근성 체크리스트 및 예제 자동 로드
```

### 방법 3: 사용자 명시적 호출

사용자가 직접 에이전트에 요청:

```
사용자: "ui-ux-expert, shadcn/ui 컴포넌트 설계해줄래?"

ui-ux-expert:
  1. Skill("moai-lib-shadcn-ui") 호출
  2. Skill("moai-lang-html-css") 호출 (접근성 검증)
  3. 컴포넌트 설계 제시
```

---

## ✅ 검증 체크리스트

### Skills 생성 검증
- ✅ moai-lang-html-css: 완성 (140+ 라인)
- ✅ moai-lang-tailwind-css: 완성 (180+ 라인)
- ✅ moai-lib-shadcn-ui: 완성 (200+ 라인)
- ✅ moai-icons-vector: 완성 (220+ 라인)

### 콘텐츠 검증
- ✅ Context7 공식 문서 기반
- ✅ 최신 버전 (HTML5, Tailwind v4, shadcn/ui v2)
- ✅ 실제 동작하는 예제 코드
- ✅ 베스트 프랙티스 포함
- ✅ 접근성 (WCAG 2.1 AA) 포함
- ✅ 레퍼런스 링크 포함

### 에이전트 통합 검증
- ✅ 호출 트리거 정의
- ✅ 사용 시나리오 문서화
- ✅ 에이전트별 역할 명확화

---

## 🚀 다음 단계

1. **Skills 배포**
   - `.claude/skills/` 디렉토리에 3개 Skill 파일 확인
   - Alfred가 자동으로 로드 가능한지 테스트

2. **에이전트 프롬프트 업데이트**
   - ui-ux-expert, frontend-expert 프롬프트에 Skill 호출 추가
   - 자동 트리거 규칙 설정

3. **테스트 및 검증**
   ```bash
   # Skill 로드 테스트
   Task(subagent_type="ui-ux-expert", prompt="시맨틱 HTML 가이드 필요")
   # → Skill("moai-lang-html-css") 자동 로드 확인

   Task(subagent_type="frontend-expert", prompt="Tailwind CSS 설정 필요")
   # → Skill("moai-lang-tailwind-css") 자동 로드 확인
   ```

4. **버전 관리**
   - Skills 파일에 버전 번호 추가 (v1.0.0)
   - Context7 업데이트 시 동기화

---

## 📚 관련 문서

- **ui-ux-expert 지침**: CLAUDE.md - Alfred UI/UX Expert
- **frontend-expert 지침**: CLAUDE.md - Alfred Frontend Expert
- **Skill 개발 가이드**: Skill("moai-cc-skills")
- **Context7 통합**: Skill("moai-jit-docs-enhanced")

---

## 🔗 Skills 파일 위치

```
/Users/goos/MoAI/MoAI-ADK/
└── .claude/
    └── skills/
        ├── moai-lang-html-css.md (ui-ux-expert용)
        ├── moai-lang-tailwind-css.md (frontend-expert용)
        └── moai-lib-shadcn-ui.md (frontend-expert용)
```

**자동 로드**: Alfred가 사용자 요청 분석 후 위 Skills을 자동으로 호출합니다.
