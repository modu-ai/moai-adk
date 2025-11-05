# MoAI-ADK 한국어 문서

이 디렉터리는 Nextra 기반으로 구축한 MoAI-ADK 온라인 문서를 담고 있습니다. Vercel에 배포하면 `/` 경로에서 바로 한국어 설명서를 제공할 수 있습니다.

## 개발 스크립트

```bash
# 개발 서버 실행 (http://localhost:3000)
npm run dev

# 정적 빌드 생성
npm run build

# 빌드 결과 확인
npm start
```

## 문서 구조

- `pages/` – MDX 문서 및 네비게이션 메타 (`meta.json`)
- `theme.config.js` – 로고, 헤더, 푸터 등 글로벌 설정
- `public/` – 파비콘 및 OG 이미지

문서 콘텐츠는 다음 자료를 출처로 삼습니다.

- `README.ko.md`
- `src/moai_adk/` 패키지 코드
- `.claude/commands`, `.claude/agents`, `.claude/skills`, `.claude/hooks`, `.claude/output-styles`

## 배포

1. Vercel에서 새 프로젝트를 만들고 이 디렉터리를 연결합니다.
2. Build Command: `npm run build`
3. Output Directory: `.next`
4. 환경 변수는 별도 설정이 필요하지 않습니다.

문서 수정 후에는 PR을 통해 리뷰하고, `/alfred:3-sync`로 문서 변경 사항을 프로젝트 기록에 남겨주세요.
