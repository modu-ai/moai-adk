# CSS 문서화 완전 가이드 인덱스

> **MoAI-ADK 문서 스타일 시스템 완전 문서화**
>
> 이 인덱스는 모든 CSS 관련 가이드를 한곳에서 찾을 수 있도록 정리했습니다.

## 문서 구성

### 1. 주 문서: CLAUDE.md
**위치**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md`

**섹션**: "CSS 아키텍처 및 스타일 시스템"

**내용**:
- 전체 CSS 계층 구조 설명
- 주요 CSS 파일 상세 분석
- 400+ 줄의 포괄적인 가이드
- Material 테마와의 관계
- 색상 팔레트 정의
- 문제 해결 방법

**언제 참고**: 깊이 있는 이해가 필요할 때, 전체 맥락을 알고 싶을 때

---

### 2. 빠른 참고서: CSS-STYLING-GUIDE.md
**위치**: `/Users/goos/MoAI/MoAI-ADK/.moai/CSS-STYLING-GUIDE.md`

**내용**:
- 핵심 규칙 3가지 (파일 관리, 우선순위, 테마)
- 자주 하는 작업 예제
- 자주 발생하는 문제 대응
- 색상 팔레트 (테이블)
- 스타일 추가 체크리스트

**언제 참고**: 빠르게 답을 찾고 싶을 때, 작은 수정을 할 때

---

### 3. 아키텍처 다이어그램: CSS-ARCHITECTURE-DIAGRAM.md
**위치**: `/Users/goos/MoAI/MoAI-ADK/.moai/CSS-ARCHITECTURE-DIAGRAM.md`

**내용**:
- 전체 스타일 흐름 (시각화)
- CSS 우선순위 계층 (다이어그램)
- 파일 관계도
- 레이아웃 구조 (데스크톱/태블릿/모바일)
- 테마 전환 플로우
- CSS 선택자 우선순위 예제
- 수정 시나리오

**언제 참고**: 시스템을 이해하고 싶을 때, 전체 구조를 시각화하고 싶을 때

---

### 4. 문제 해결: CSS-TROUBLESHOOTING.md
**위치**: `/Users/goos/MoAI/MoAI-ADK/.moai/CSS-TROUBLESHOOTING.md`

**내용**:
- 진단 플로우차트
- 색상 문제 해결 (5가지 시나리오)
- 레이아웃 문제 해결 (5가지 시나리오)
- 폰트/텍스트 문제 해결
- 반응형 문제 해결
- 일반 진단
- 성능 진단
- 빠른 체크리스트

**언제 참고**: 문제가 발생했을 때, 단계별로 진단하고 해결하고 싶을 때

---

## 빠른 선택 가이드

### "텍스트 색상을 다크 모드에서 더 밝게 하고 싶어요"

1. **빠른 해결**: CSS-STYLING-GUIDE.md → "자주 하는 작업" → "텍스트 색상 변경"
2. **깊이 있는 이해**: CLAUDE.md → "CSS 우선순위" + "색상 팔레트"
3. **문제 해결**: CSS-TROUBLESHOOTING.md → "색상 진단" → "다크 모드 색상"

### "모바일에서 콘텐츠가 중앙 정렬되지 않아요"

1. **진단**: CSS-TROUBLESHOOTING.md → "반응형 진단"
2. **해결**: CSS-STYLING-GUIDE.md → "자주 발생하는 문제" 또는 CSS-ARCHITECTURE-DIAGRAM.md → "모바일 레이아웃"
3. **이해**: CLAUDE.md → "레이아웃 최적화" 섹션

### "새로운 스타일을 추가하고 싶은데 어디에 추가해야 하나요?"

1. **규칙 확인**: CSS-STYLING-GUIDE.md → "핵심 규칙 3가지"
2. **체크리스트**: CSS-STYLING-GUIDE.md → "스타일 추가 체크리스트"
3. **상세 설명**: CLAUDE.md → "주요 CSS 섹션별 설명"

### "CSS 시스템을 완전히 이해하고 싶어요"

1. **개요**: CSS-ARCHITECTURE-DIAGRAM.md → "전체 스타일 흐름"
2. **상세**: CLAUDE.md → "CSS 아키텍처 및 스타일 시스템" 전체
3. **심화**: CLAUDE.md → "Material 테마의 주요 CSS 변수"

### "스타일이 적용되지 않아요"

1. **빠른 체크**: CSS-TROUBLESHOOTING.md → "빠른 체크리스트"
2. **상세 진단**: CSS-TROUBLESHOOTING.md → "일반 진단" → "스타일이 전혀 적용 안 됨"

---

## 주요 개념 매트릭스

| 개념 | CLAUDE.md | Guide | Diagram | Troubleshooting |
|------|-----------|-------|---------|-----------------|
| 파일 구조 | ✅ 상세 | ✅ 간결 | ✅ 다이어그램 | ✅ 진단 |
| CSS 우선순위 | ✅ 상세 | ✅ 간결 | ✅ 시각화 | ✅ 예제 |
| 라이트/다크 모드 | ✅ 상세 | ✅ 예제 | ✅ 흐름도 | ✅ 문제 |
| 색상 팔레트 | ✅ 표 | ✅ 표 | | ✅ 진단 |
| 레이아웃 | ✅ 코드 | ✅ CSS | ✅ 시각화 | ✅ 문제 |
| 반응형 | ✅ 상세 | ✅ 예제 | ✅ 다이어그램 | ✅ 문제 |
| 문제 해결 | ✅ 표 | ✅ 시나리오 | | ✅ 플로우 |
| 성능 | ✅ 권장사항 | | | ✅ 진단 |

---

## 파일 변경 시 참고

### CSS 파일 추가/수정 체크리스트

```
1. 파일 선택
   ❓ 어느 파일을 수정할 건가요?
   → CSS-STYLING-GUIDE.md: "핵심 규칙 3가지"

2. 스타일 작성
   ❓ 어떻게 작성해야 하나요?
   → CSS-STYLING-GUIDE.md: "자주 하는 작업"

3. 라이트/다크 모드
   ❓ 다크 모드도 처리해야 하나요?
   → CLAUDE.md: "라이트/다크 모드 최적화"

4. 반응형 확인
   ❓ 모바일에서도 테스트했나요?
   → CSS-ARCHITECTURE-DIAGRAM.md: "레이아웃 구조"

5. 문제 진단
   ❓ 스타일이 적용 안 되었다면?
   → CSS-TROUBLESHOOTING.md: "빠른 체크리스트"

6. 최종 확인
   ❓ 완벽한가요?
   → CSS-STYLING-GUIDE.md: "스타일 추가 체크리스트"
```

---

## 핸드북 요약

### 규칙 1: 파일 관리
- `/docs/stylesheets/extra.css` **만** 수정
- 설정은 `mkdocs.yml`, 변수는 `overrides/main.html`

### 규칙 2: CSS 우선순위
1. 인라인 스타일 (style="...")
2. extra.css의 !important
3. extra.css의 일반 규칙
4. Material 테마 기본값
5. 브라우저 기본값

### 규칙 3: 테마 동기화
- 라이트 모드: `[data-md-color-scheme="default"]`
- 다크 모드: `[data-md-color-scheme="slate"]`
- **항상 함께 처리하기**

### 규칙 4: 반응형
- 데스크톱: `> 1220px` (70:30 레이아웃)
- 태블릿: `76.25em - 1220px` (조정)
- 모바일: `< 76.25em` (100% 너비)

---

## 색상 팔레트 (한눈에 보기)

### 라이트 모드 (Light)
```
배경: #ffffff    텍스트: #171612    코드: #f5f5f5
강조: #504F4B    테두리: #D5D4D0    헤더: #F7F6F2
```

### 다크 모드 (Dark - Slate)
```
배경: #171612    텍스트: #F7F6F2    코드: #2a2a2a
강조: #C9C8C4    테두리: #444444    헤더: #1A1916
```

---

## 일반적인 작업별 가이드

| 작업 | 참고 문서 | 시간 |
|------|---------|------|
| 링크 색상 변경 | Guide → 텍스트 색상 | 2분 |
| 코드 블록 스타일 | CLAUDE.md → E섹션 | 5분 |
| 테이블 수정 | CLAUDE.md → F섹션 | 5분 |
| 모바일 레이아웃 | Diagram → 모바일 구조 | 10분 |
| 새로운 요소 스타일링 | Guide → 스타일 추가 | 15분 |
| 문제 해결 | Troubleshooting → 진단 | 10-20분 |
| 시스템 전체 이해 | CLAUDE.md 전체 | 30분 |

---

## 문서 업데이트 히스토리

- **2025-11-10**: 초기 완전 문서화
  - CLAUDE.md에 CSS 아키텍처 섹션 추가 (400+ 줄)
  - CSS-STYLING-GUIDE.md 작성
  - CSS-ARCHITECTURE-DIAGRAM.md 작성
  - CSS-TROUBLESHOOTING.md 작성
  - CSS-DOCUMENTATION-INDEX.md 작성

---

## 추가 리소스

### Material for MkDocs 공식 문서
- https://squidfunk.github.io/mkdocs-material/

### CSS 학습 리소스
- https://developer.mozilla.org/en-US/docs/Web/CSS/
- https://web.dev/css-cascade/

### 색상 도구
- WebAIM Contrast Checker: https://webaim.org/resources/contrastchecker/
- Color Picker: https://htmlcolorcodes.com/

### 브라우저 개발자 도구
- F12 또는 Ctrl+Shift+I (Windows)
- Cmd+Option+I (Mac)

---

## 질문과 답변

**Q: 왜 이렇게 많은 문서가 필요한가요?**
A: CSS 시스템이 복잡하기 때문입니다:
- Material 테마 (기본 스타일)
- overrides/main.html (변수)
- extra.css (커스텀)
- 라이트/다크 모드 (2배)
- 반응형 (3배)

따라서 빠른 참고, 깊이 있는 설명, 시각화, 문제 해결 등 다양한 관점이 필요합니다.

**Q: 어디서 시작해야 하나요?**
A: 당신의 상황에 따라:
1. **급할 때**: CSS-STYLING-GUIDE.md
2. **이해하고 싶을 때**: CSS-ARCHITECTURE-DIAGRAM.md
3. **완벽히 배우고 싶을 때**: CLAUDE.md
4. **문제가 있을 때**: CSS-TROUBLESHOOTING.md

**Q: Material for MkDocs를 알 필요가 있나요?**
A: 선택사항입니다:
- Material 문서를 모르고도 이 가이드로 충분합니다
- Material의 CSS 변수를 이해하면 더 쉬워집니다
- 깊이 있게 하려면 Material 공식 문서를 보세요

---

**문서 관리자**: Alfred SuperAgent
**마지막 업데이트**: 2025-11-10
**상태**: 완성 (v1.0)
