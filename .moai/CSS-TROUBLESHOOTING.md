# CSS 스타일 문제 해결 가이드

> **CSS 스타일 관련 문제를 진단하고 해결하는 단계별 가이드**

## 진단 플로우차트

```
문제 발생?
│
├─ 색상 관련 문제?
│  ├─ YES → [색상 진단] 섹션으로
│  └─ NO  → 다음으로
│
├─ 레이아웃 관련 문제?
│  ├─ YES → [레이아웃 진단] 섹션으로
│  └─ NO  → 다음으로
│
├─ 타이포그래피 관련 문제?
│  ├─ YES → [폰트/텍스트 진단] 섹션으로
│  └─ NO  → 다음으로
│
└─ 모바일/반응형 문제?
   ├─ YES → [반응형 진단] 섹션으로
   └─ NO  → [일반 진단] 섹션으로
```

## 1. 색상 진단 (Color Issues)

### 문제: 라이트 모드에서 텍스트가 안 보임

**증상**:
```
라이트 모드에서 텍스트가 배경과 구분이 안 됨
버튼/링크 색상이 옅음
```

**원인 분석**:
```
┌─────────────────────────────────┐
│ 원인 1: 색상 대비 부족          │
│ (배경: 밝음, 텍스트: 밝음)      │
└─────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────┐
│ 원인 2: CSS 변수 적용 안 됨      │
│ (하드코딩된 색상 사용)          │
└─────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────┐
│ 원인 3: Material 기본값 오버     │
│ (우선순위 문제)                  │
└─────────────────────────────────┘
```

**해결 단계**:

```
1단계: 어떤 요소인지 확인
   → F12 개발자 도구에서 요소 검사
   → HTML 요소 이름 확인 (ex: .md-content a)

2단계: 현재 색상 확인
   → 개발자 도구의 "Computed" 탭에서 색상 확인
   → 어느 파일에서 적용되었는지 확인

3단계: 대비율 확인
   → 배경색과 텍스트 색상 기록
   → WebAIM Contrast Checker 사용
   → (최소 4.5:1 권장, WCAG AA 기준)

4단계: 수정 방법 선택
   ├─ CSS 변수 사용 권장
   │  color: var(--md-default-fg-color)
   │
   └─ 또는 직접 색상 지정
      [data-md-color-scheme="default"] .my-element {
        color: #171612;
      }
      [data-md-color-scheme="slate"] .my-element {
        color: #F7F6F2;
      }
```

**검증**:
```
[ ] 라이트 모드에서 텍스트 가독성 확인
[ ] 다크 모드에서 텍스트 가독성 확인
[ ] 컬러 블라인드 시뮬레이터로 확인
    (Chrome DevTools > Rendering > Emulate CSS media feature)
```

### 문제: 다크 모드에서 색상이 라이트 모드와 같음

**증상**:
```
다크 모드 전환해도 색상이 변하지 않음
일부 요소만 색상이 변함
```

**원인**:
```
다크 모드 CSS 규칙이 없음
│
├─ 하드코딩된 색상 사용
│  (#171612처럼 직접 입력)
│
└─ CSS 변수를 라이트 모드에서만 정의
   ([data-md-color-scheme="slate"] 규칙 누락)
```

**해결책**:

```css
/* ❌ 나쁜 예: 다크 모드 규칙 없음 */
.my-element {
  color: #171612;  /* 항상 검은색 */
}

/* ✅ 좋은 예: 라이트/다크 모드 모두 정의 */
[data-md-color-scheme="default"] .my-element {
  color: #171612;  /* 라이트 모드: 검은색 */
}

[data-md-color-scheme="slate"] .my-element {
  color: #F7F6F2;  /* 다크 모드: 밝은 색 */
}

/* 또는 CSS 변수 사용 (권장) */
.my-element {
  color: var(--md-default-fg-color);  /* 자동 전환 */
}
```

**검증**:
```
[ ] 라이트 모드에서 색상 확인
[ ] 테마 전환 버튼 클릭
[ ] 다크 모드에서 색상 변경 확인
[ ] 새로고침 후에도 유지되는지 확인
```

### 문제: 특정 요소만 색상이 안 바뀜

**증상**:
```
대부분 요소는 다크 모드에서 색상이 변하지만
특정 버튼/링크/텍스트만 안 변함
```

**원인**:
```
!important로 인라인 스타일이 우선됨
│
├─ 요소에 style="color: red" 직접 지정됨
│
├─ 더 구체적인 선택자가 있음
│  (ex: .md-content.specific .element)
│
└─ CSS 변수 적용 안 됨
```

**해결책**:

```css
/* 진단: 개발자 도구에서 확인 */
F12 → 요소 검사 → Styles 탭
→ 색상이 어디서 적용되는지 보기

/* 해결책 1: !important 추가 */
[data-md-color-scheme="slate"] .my-element {
  color: #F7F6F2 !important;
}

/* 해결책 2: CSS 변수를 !important로 */
.my-element {
  color: var(--md-default-fg-color) !important;
}

/* 해결책 3: 더 구체적인 선택자 */
[data-md-color-scheme="slate"]
.md-content .specific .element {
  color: #F7F6F2;
}
```

## 2. 레이아웃 진단 (Layout Issues)

### 문제: 콘텐츠가 너무 좁음 (70% 너비)

**증상**:
```
화면이 크지만 콘텐츠가 화면의 절반만 차지
우측 TOC가 너무 많은 공간 차지
```

**원인**:
```
.md-main__inner 레이아웃이 70:30으로 고정됨
│
├─ TOC(우측 사이드바)가 필요 없을 수도 있음
│
└─ 콘텐츠가 길어서 스크롤이 많음
```

**진단**:

```
1. 개발자 도구 → 요소 검사
2. .md-sidebar--secondary 찾기
3. display 속성 확인
4. width: 30% 확인
```

**해결책**:

```css
/* TOC 숨기기 */
.md-sidebar--secondary {
  display: none !important;
}

/* 콘텐츠 100% 너비로 만들기 */
.md-main__inner .md-content {
  flex: 0 0 100% !important;
  width: 100% !important;
}

/* 특정 페이지에서만 적용 */
body.specific-page .md-sidebar--secondary {
  display: none !important;
}
```

### 문제: 헤더가 화면 너비와 안 맞음

**증상**:
```
헤더가 콘텐츠보다 더 넓음
또는 더 좁음
양쪽에 여백이 다름
```

**원인**:
```
.md-container와 .md-header__inner의 max-width가 다름
│
├─ 콘텐츠: max-width: 1220px
│
└─ 헤더: max-width: 다른 값
```

**해결책**:

```css
/* 모두 동일한 크기로 설정 */
.md-container,
.md-header__inner,
.md-footer__inner,
.md-main {
  max-width: 1220px !important;
  margin: 0 auto !important;
  padding: 0 1rem !important;
}
```

**검증**:
```
[ ] 개발자 도구 → 요소 검사
[ ] 각 섹션의 max-width 확인
[ ] 모두 1220px인지 확인
[ ] 화면을 다양한 크기로 조정
```

### 문제: 테이블이 깨져 보임

**증상**:
```
테이블 셀이 겹침
헤더와 데이터 정렬이 안 맞음
테이블이 grid 레이아웃처럼 보임
```

**원인**:
```
Material 테마의 grid 레이아웃과 충돌
│
├─ display: grid 설정이 우선됨
│
└─ display: table 설정이 없음
```

**해결책**:

```css
/* table 레이아웃 강제 */
.md-typeset table {
  display: table !important;
  width: 100% !important;
  max-width: 100% !important;
  border-collapse: collapse;
}

.md-typeset thead {
  display: table-header-group !important;
}

.md-typeset tbody {
  display: table-row-group !important;
}

.md-typeset tr {
  display: table-row !important;
}

.md-typeset th,
.md-typeset td {
  display: table-cell !important;
  padding: 12px 16px !important;
}
```

## 3. 폰트/텍스트 진단 (Typography Issues)

### 문제: 이탤릭체가 나타남

**증상**:
```
텍스트가 기울어져 보임 (이탤릭)
제목이나 강조 텍스트가 특히 심함
```

**원인**:
```
마크다운의 *텍스트*가 <em> 또는 <i> 태그로 변환됨
│
├─ 브라우저가 <em>, <i>을 기울인 글씨로 표시
│
└─ font-style: italic이 적용됨
```

**확인**:

```
F12 → 요소 검사 → <em> 또는 <i> 찾기
→ Computed 탭에서 font-style 확인
→ italic이 보이면 확인됨
```

**해결책**:

```css
/* 모든 기울임 제거 */
*,
*::before,
*::after,
em,
i,
.md-content em,
.md-content i {
  font-style: normal !important;
}

/* 또는 특정 요소만 */
.md-content em {
  font-style: normal !important;
  font-weight: 600;  /* 대신 굵기로 강조 */
}
```

**원인**: 한글 렌더링에서 이탤릭이 제대로 작동하지 않기 때문

### 문제: 폰트가 깨짐 (한글/일본어)

**증상**:
```
한글이 네모칸으로 표시됨
특정 문자가 정렬이 안 맞음
웹폰트가 로드되지 않음
```

**원인**:
```
overrides/main.html에서 폰트를 로드하지 않음
│
├─ @font-face 정의 누락
│
└─ 폰트 파일 경로 오류
```

**해결책**:

```html
<!-- overrides/main.html 헤더에 추가 -->
<link href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;600;700&display=swap" rel="stylesheet">

<style>
  :root {
    --font-ko-text: 'Noto Sans KR', system-ui, sans-serif;
  }

  .md-typeset {
    font-family: var(--font-ko-text);
  }
</style>
```

**검증**:
```
[ ] F12 → Network 탭에서 폰트 파일 로드 확인
[ ] 한글/일본어 텍스트 렌더링 확인
[ ] 특수 문자도 확인
```

## 4. 반응형 진단 (Responsive Issues)

### 문제: 모바일에서 레이아웃이 깨짐

**증상**:
```
텍스트가 화면 너비를 초과함
좌측 메뉴가 안 열림
콘텐츠가 겹침
```

**원인**:
```
@media query 규칙이 없거나 잘못됨
│
├─ 중단점이 76.25em이 아님
│
├─ flex-direction이 column이 아님
│
└─ width 설정이 100%가 아님
```

**진단**:

```
1. 브라우저를 좁혀서 확인
2. F12 → 개발자 도구 상단의 "Toggle device toolbar" 클릭
3. iPhone 등을 선택
4. 어느 지점부터 깨지는지 확인
```

**해결책**:

```css
/* 76.25em (1220px) 이하에서만 적용 */
@media screen and (max-width: 76.25em) {
  /* 수직 스택 레이아웃 */
  .md-main__inner {
    flex-direction: column !important;
  }

  /* 콘텐츠 100% 너비 */
  .md-main__inner .md-content {
    flex: 0 0 100% !important;
    width: 100% !important;
    padding: 1rem !important;
  }

  /* TOC를 아래로 이동 */
  .md-sidebar--secondary {
    flex: 0 0 100% !important;
    border-left: none !important;
    border-top: 1px solid var(...);
  }
}
```

**검증**:
```
[ ] iPhone (375px) - 확인
[ ] iPad (768px) - 확인
[ ] 태블릿 (1024px) - 확인
[ ] 데스크톱 (1220px+) - 확인
[ ] 초광화면 (1920px+) - 확인
```

### 문제: 메뉴 버튼이 안 나타남 (모바일)

**증상**:
```
모바일에서 "☰" 메뉴 아이콘이 없음
네비게이션에 접근할 수 없음
```

**원인**:
```
.md-sidebar--primary가 display: none 상태
│
└─ 모바일에서도 표시 규칙이 없음
```

**확인**:

```css
/* 데스크톱 */
.md-sidebar--primary {
  display: none !important;  /* 숨김 */
}

/* 모바일에서는 버튼으로 표시되어야 함 */
@media screen and (max-width: 76.25em) {
  .md-sidebar--primary {
    display: block !important;  /* 표시 */
    position: fixed !important;
    left: -100% !important;     /* 처음엔 화면 밖 */
  }
}
```

## 5. 일반 진단 (General Troubleshooting)

### 스타일이 전혀 적용 안 됨

**증상**:
```
변경한 CSS가 반영되지 않음
오래된 버전이 계속 보임
```

**원인**:
```
1. 브라우저 캐시
2. 파일 저장 안 됨
3. 잘못된 파일 수정
4. 선택자 오류
```

**해결 단계**:

```
1단계: 파일 저장 확인
   → /docs/stylesheets/extra.css 저장 확인
   → Ctrl+S (Windows) 또는 Cmd+S (Mac)

2단계: 캐시 제거
   → mkdocs serve 중지 (Ctrl+C)
   → 브라우저 Ctrl+Shift+Delete
   → "캐시된 이미지 및 파일 삭제"
   → mkdocs serve 재시작

3단계: 선택자 확인
   → F12 개발자 도구
   → 요소 검사
   → 정확한 클래스명 확인
   → 복사해서 CSS에 붙여넣기

4단계: 문법 확인
   → CSS 문법 오류 확인
   → ; 누락 확인
   → { } 괄호 쌍 확인
```

### 선택자가 작동하지 않음

**증상**:
```
CSS 규칙을 작성했지만 적용 안 됨
F12에서 선택자가 회색으로 보임
```

**원인**:
```
선택자가 실제 HTML과 다름
│
├─ 클래스명 오타
│  (.md-conent vs .md-content)
│
├─ 존재하지 않는 요소
│  (.my-element가 HTML에 없음)
│
└─ 더 강한 규칙이 우선됨
   (!important 누락)
```

**확인**:

```
1. F12 → 요소 검사
2. 실제 HTML 구조 확인
3. 정확한 클래스/ID 확인
4. 복사해서 CSS에 붙여넣기
```

**예제**:

```css
/* ❌ 나쁜 예 */
.my-element { color: red; }  /* HTML에 없음 */

/* ✅ 좋은 예 */
.md-content h1 { color: red; }  /* 실제 HTML 구조 */

/* 더 강한 선택자 필요 */
[data-md-color-scheme="default"] .md-content h1 {
  color: red !important;
}
```

## 6. 성능 진단 (Performance Issues)

### 문제: 페이지 로딩이 느림

**증상**:
```
mkdocs serve를 시작할 때 오래 걸림
페이지를 열 때 렌더링이 느림
```

**원인**:
```
과도한 !important 사용
│
├─ 브라우저가 우선순위 재계산
│
├─ 복잡한 선택자
│  (.a .b .c .d .e { })
│
└─ 과도한 미디어 쿼리
```

**해결책**:

```css
/* ❌ 나쁜 예: 과도한 중첩 */
.md-main__inner > .md-content > div > p > span {
  color: red;
}

/* ✅ 좋은 예: 간단한 선택자 */
.md-content span {
  color: red;
}

/* ❌ 나쁜 예: 과도한 !important */
.element { color: red !important; }
.element { border: 1px solid !important; }
.element { padding: 10px !important; }

/* ✅ 좋은 예: 필요한 경우만 */
.element {
  color: red;
  border: 1px solid;
  padding: 10px;
}

[data-md-color-scheme="slate"] .element {
  color: white !important;  /* Material 오버라이드만 */
}
```

## 빠른 체크리스트

```
CSS 문제 해결 체크리스트
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] 파일 저장 확인
    [ ] /docs/stylesheets/extra.css 저장됨
    [ ] VS Code에서 점(.) 사라짐

[ ] 캐시 제거
    [ ] mkdocs serve 중지
    [ ] 브라우저 캐시 제거
    [ ] mkdocs serve 재시작

[ ] 선택자 확인
    [ ] F12 → 요소 검사
    [ ] 정확한 클래스명/ID
    [ ] 존재하는 요소인가?

[ ] CSS 문법 확인
    [ ] { } 괄호 쌍 확인
    [ ] ; 세미콜론 확인
    [ ] 색상 형식 확인 (#RRGGBB)

[ ] 우선순위 확인
    [ ] !important 필요한가?
    [ ] 다른 규칙이 우선되나?
    [ ] Material 테마 오버라이드 필요한가?

[ ] 라이트/다크 모드 확인
    [ ] 라이트 모드 스타일
    [ ] [data-md-color-scheme="slate"] 규칙
    [ ] 테마 전환 테스트

[ ] 반응형 확인
    [ ] 데스크톱 확인
    [ ] 태블릿 확인 (F12 Toggle device)
    [ ] 모바일 확인

[ ] 최종 테스트
    [ ] mkdocs serve 재시작
    [ ] 브라우저 새로고침 (Ctrl+Shift+R)
    [ ] 여러 브라우저 확인
```

---

**더 복잡한 문제?**
→ CLAUDE.md의 전체 CSS 아키텍처 섹션을 참고하세요.
