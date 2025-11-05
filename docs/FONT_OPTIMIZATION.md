# MoAI-ADK 다국어 웹폰트 최적화 전략

## 개요

MoAI-ADK 문서 사이트의 다국어(한국어, 영어, 일본어, 중국어) 웹폰트 최적화 전략입니다. 각 언어별 최적의 폰트 조합과 성능 최적화를 구현했습니다.

## 폰트 구성

### 1. 한국어 (Primary)
- **Pretendard**: 현대적인 한국어 폰트 (CDN 로드)
- **보조 폰트**: Noto Sans KR, Apple SD Gothic Neo, Malgun Gothic
- **코드 폰트**: JetBrains Mono, D2Coding

### 2. 영어
- **Inter**: 현대적이고 가독성 높은 영문 폰트 (Google Fonts)
- **보조 폰트**: Roboto, Helvetica Neue, Arial
- **코드 폰트**: JetBrains Mono, Fira Code

### 3. 일본어
- **Noto Sans JP**: Google이 제공하는 최적의 일본어 폰트
- **보조 폰트**: Hiragino Kaku Gothic ProN, Meiryo
- **코드 폰트**: JetBrains Mono, GenEi Gothic K

### 4. 중국어
- **Noto Sans SC**: 중국어 간체 최적화 폰트
- **보조 폰트**: PingFang SC, Microsoft YaHei
- **코드 폰트**: JetBrains Mono, Consolas

## 성능 최적화 전략

### 1. Font Loading Strategy
- **Preconnect**: Google Fonts CDN에 사전 연결
- **font-display: swap**: FOUT/FOIT 방지
- **Dynamic Subset**: Pretendard dynamic-subset 사용
- **WOFF2 우선**: 현대적인 브라우저 지원

### 2. Loading 최적화
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
```

### 3. 폰트 스와핑 전략
```css
.md-typeset {
  font-display: swap;
}
```

## 구현 방법

### 파일 구조
```
docs/
├── mkdocs.yml                # 메인 설정 파일
├── overrides/
│   ├── main.html            # 커스텀 HTML 템플릿
│   └── .gitkeep             # 디렉토리 유지 파일
└── src/                     # 마크다운 소스
```

### Custom CSS 변수
```css
:root {
  --font-ko-text: 'Pretendard', 'Noto Sans KR', ...;
  --font-en-text: 'Inter', 'Roboto', ...;
  --font-ja-text: 'Noto Sans JP', ...;
  --font-zh-text: 'Noto Sans SC', ...;
}
```

### 언어별 폰트 적용
```css
html[lang="ko"] .md-typeset {
  font-family: var(--font-ko-text) !important;
}

html[lang="en"] .md-typeset {
  font-family: var(--font-en-text) !important;
}
```

## 렌더링 최적화

### 1. Line Height 조정
- **한국어/일본어/중국어**: 1.6 (문자 간격 확보)
- **영어**: 1.5 (가독성 최적화)

### 2. Typography 최적화
- **헤딩**: font-weight 600, letter-spacing -0.025em
- **본문**: 언어별 최적 line-height
- **코드**: ligature 비활성화 (개발자 가독성)

## 테스트 방법

### 1. 로컬 테스트
```bash
cd docs
uv run mkdocs serve
# http://127.0.0.1:8000 접속
```

### 2. 언어별 테스트
- 한국어: `/` (기본)
- 영어: `/en/`
- 일본어: `/ja/`
- 중국어: `/zh/`

### 3. 성능 측정
- Chrome DevTools Network 탭
- Lighthouse Performance 점수
- Web Vitals (LCP, FID, CLS)

## 배포 최적화

### Vercel 배포 고려사항
- CDN 캐싱 활용
- 폰트 파일 압축
- HTTP/2 Server Push
- 브라우저 캐싱 전략

### 빌드 최적화
```bash
# Minification 활성화
uv run mkdocs build --clean
```

## 주요 특징

### 1. 자동 언어 감지
- HTML lang 속성 기반 폰트 전환
- 사용자 언어 설정 자동 적용

### 2. Fallback 체계
- 각 언어별 3단계 폰트 fallback
- 시스템 폰트 최종 fallback

### 3. 성능 우선
- Critical CSS 인라인
- Non-blocking font loading
- Progressive enhancement

## 유지보수

### 1. 버전 관리
- Pretendard: v1.3.9 (최신 버전 유지)
- Google Fonts: 최신 버전 자동 적용

### 2. 모니터링
- 폰트 로딩 시간 추적
- 렌더링 이슈 감지
- 사용자 피드백 수집

### 3. 최적화 주기
- 분기별 폰트 성능 검토
- 새로운 폰트 기술 적용 검토
- 사용자 경험 개선

## 참고 자료

- [Pretendard GitHub](https://github.com/orioncactus/pretendard)
- [Google Fonts](https://fonts.google.com/)
- [Web Font Loading Best Practices](https://web.dev/font-best-practices/)
- [MkDocs Material Customization](https://squidfunk.github.io/mkdocs-material/customization/)