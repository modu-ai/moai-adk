# CSS 스타일 시스템 빠른 참고서

> **문서 유지보수 담당자를 위한 빠른 참고 가이드**
>
> 전체 설명은 `CLAUDE.md`의 "CSS 아키텍처 및 스타일 시스템" 섹션을 참고하세요.

## 핵심 규칙 3가지

### 1. 파일 관리: 하나의 파일만 수정

```
❌ 하면 안 됨                          ✅ 해야 할 일
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
mkdocs.yml 수정                        설정만 변경 (플러그인, 테마)
overrides/main.html 수정               색상 변수만 수정
여러 CSS 파일 수정                     /docs/stylesheets/extra.css만 수정
```

### 2. CSS 우선순위 명확히 이해

```
Material 테마를 오버라이드하려면:
  → extra.css에서 !important 사용

Material 변수를 사용하려면:
  → var(--md-xxx-color) 형태로 사용

Material 기본값을 따르려면:
  → !important 없이 일반 CSS 규칙
```

### 3. 라이트/다크 모드는 항상 함께

```css
/* ❌ 라이트 모드만 스타일 */
.md-content a {
  color: #171612;
}

/* ✅ 라이트 + 다크 모드 함께 */
.md-content a {
  color: var(--md-accent-fg-color);
}

[data-md-color-scheme="slate"] .md-content a {
  color: #C9C8C4;
}
```

## 자주 하는 작업

### 텍스트 색상 변경

```css
/* 현재 스키마(라이트/다크)에 맞게 자동 적용 */
.my-element {
  color: var(--md-default-fg-color);
}
```

**주요 변수**:
- `--md-default-fg-color` - 본문 텍스트
- `--md-accent-fg-color` - 강조색 (링크, 제목)
- `--md-code-fg-color` - 코드 텍스트

### 배경색 변경

```css
/* Material 테마 배경색 사용 */
.my-element {
  background-color: var(--md-default-bg-color);
}

/* 또는 수동으로 지정 */
[data-md-color-scheme="default"] .my-element {
  background-color: #ffffff;
}

[data-md-color-scheme="slate"] .my-element {
  background-color: #171612;
}
```

**주요 변수**:
- `--md-default-bg-color` - 본문 배경
- `--md-code-bg-color` - 코드 배경
- `--md-accent-bg-color--light` - 강조 배경 (연함)

### 테두리 색상 변경

```css
.my-element {
  border: 1px solid var(--md-default-fg-color--lightest);
}
```

### 반응형 디자인 (모바일 대응)

```css
/* 데스크톱: 70% 콘텐츠, 30% 사이드바 */
.md-content { width: 70%; }

/* 모바일: 100% 콘텐츠 */
@media screen and (max-width: 76.25em) {
  .md-content { width: 100%; }
}
```

**주요 중단점**:
- `> 1220px` - 데스크톱 (70:30 레이아웃)
- `76.25em - 1220px` - 태블릿 (조정)
- `< 76.25em` - 모바일 (100% 너비)

## 자주 발생하는 문제

### 문제: 다크 모드에서 텍스트가 안 보임

**원인**: 라이트 모드만 색상을 지정했을 때

**해결책**:
```css
/* ❌ 나쁜 예 */
.my-element { color: #171612; }

/* ✅ 좋은 예 */
.my-element { color: var(--md-default-fg-color); }
```

### 문제: Material 테마 기본값이 적용되지 않음

**원인**: !important가 없거나 선택자 특이도가 낮음

**해결책**:
```css
/* ❌ 나쁜 예 */
.md-header { background-color: #F7F6F2; }

/* ✅ 좋은 예 */
[data-md-color-scheme="default"] .md-header {
  background-color: #F7F6F2 !important;
}
```

### 문제: 모바일에서 레이아웃이 깨짐

**원인**: 반응형 CSS 규칙이 없음

**해결책**:
```css
@media screen and (max-width: 76.25em) {
  .md-main__inner {
    flex-direction: column !important;
  }

  .md-content {
    width: 100% !important;
  }
}
```

### 문제: 코드 블록에 이상한 테두리가 있음

**원인**: Material 테마 기본값 적용

**해결책**:
```css
.md-content .highlight {
  border: none !important;
}
```

## 색상 팔레트 (빠른 참고)

### 라이트 모드

| 요소 | 색상 | 용도 |
|------|------|------|
| 배경 | `#ffffff` | 페이지 배경 |
| 텍스트 | `#171612` | 본문 텍스트 |
| 헤더 | `#F7F6F2` | 헤더/탭 배경 |
| 코드 배경 | `#f5f5f5` | 코드 블록 배경 |
| 강조색 | `#504F4B` | 링크, 제목 |
| 테두리 | `#D5D4D0` | 구분선 |

### 다크 모드

| 요소 | 색상 | 용도 |
|------|------|------|
| 배경 | `#171612` | 페이지 배경 |
| 텍스트 | `#F7F6F2` | 본문 텍스트 |
| 헤더 | `#1A1916` | 헤더/탭 배경 |
| 코드 배경 | `#2a2a2a` | 코드 블록 배경 |
| 강조색 | `#C9C8C4` | 링크, 제목 |
| 테두리 | `#444444` | 구분선 |

## 스타일 추가 체크리스트

새로운 스타일을 추가할 때 확인하세요:

```
[ ] 파일 위치: /docs/stylesheets/extra.css에만 추가
[ ] 라이트 모드: [data-md-color-scheme="default"] 규칙 추가
[ ] 다크 모드: [data-md-color-scheme="slate"] 규칙 추가
[ ] 색상 변수: var(--md-xxx-color) 활용
[ ] 반응형: @media (max-width: 76.25em) 규칙 추가
[ ] 테스트: 라이트/다크/모바일 모두 확인
[ ] 주석: "왜" 이 스타일이 필요한지 설명
```

## 빠른 명령어

### 로컬에서 문서 확인

```bash
cd docs
mkdocs serve
# http://localhost:8000 에서 확인
```

### CSS 린트 (선택사항)

```bash
# CSS 문법 검사
stylelint stylesheets/extra.css
```

## 참고 자료

| 문서 | 위치 | 목적 |
|------|------|------|
| 전체 CSS 가이드 | `CLAUDE.md` - "CSS 아키텍처" | 상세한 설명 |
| CSS 파일 | `/docs/stylesheets/extra.css` | 실제 스타일 |
| 템플릿 | `/docs/overrides/main.html` | 색상 변수, 폰트 |
| 설정 | `/docs/mkdocs.yml` | 테마 설정 |

## 요청하기

이 문서를 개선하고 싶으신가요?

1. 문제점을 발견했으면 GitHub Issue 생성
2. CSS 개선 제안은 PR 생성
3. 색상 팔레트 변경은 CLAUDE.md 업데이트 후 팀과 공유

---

**마지막 업데이트**: 2025-11-10
**관리자**: Alfred SuperAgent
