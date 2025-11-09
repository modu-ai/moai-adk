# 다크/라이트 테마 & 배포 검증 가이드

MoAI-ADK 문서는 `Next.js + Nextra` 기반으로 빌드되며, `docs/theme.config.cjs`와 `docs/stylesheets` 토큰이 다크·라이트 테마를 결정합니다. 아래 체크리스트를 따라 배포 전/후 품질을 빠르게 검증하세요.

## 1. 테마 토글 체크리스트

| 단계 | 명령/행동 | 기대 결과 | 관련 파일 |
| --- | --- | --- | --- |
| 1 | `npm run --prefix docs dev` 실행 후 브라우저에서 `/` 접속 | 기본값이 시스템 테마와 동일하게 시작 | `docs/theme.config.cjs` (`nextThemes.defaultTheme`) |
| 2 | 우측 상단 테마 토글을 라이트→다크로 전환 | 본문 대비 12+:1 확보, 표/다이어그램 경계 유지 | `docs/index.md#7-다크라이트-테마-가이드라인` |
| 3 | `/guides/statusline` 등 코드 블록 포함 페이지 확인 | 라이트: `#f3f4f6`, 다크: `#1e1e2f` 배경 유지 | `docs/stylesheets/prism.css` |
| 4 | 모바일 뷰(DevTools)에서 토글 반복 | 토글 상태가 `localStorage.moai-adk-docs-theme`에 저장 | `theme.config.cjs` (`nextThemes.storageKey`) |

## 2. 색상 토큰 빠른 점검

- `surface`: 라이트 `#f8fafc`, 다크 `#0f172a`
- `card/block`: 라이트 `#ffffff`, 다크 `#1e293b`
- `accent`: 라이트 `#4f46e5`, 다크 `#818cf8`
- `warning/success badges`: 본색 유지 + 20% 투명도 배경

페이지 내에서 색이 다르게 보이면 `docs/stylesheets/*.css` 또는 컴포넌트별 inline 스타일을 확인하세요.

## 3. 빌드 및 Vercel 배포 절차

```bash
# 1) 정적 빌드 (필수)
npm run --prefix docs build

# 2) Vercel 프리뷰 배포 (선택)
cd docs
npx vercel --prebuilt --confirm

# 3) 프로덕션 프로모트
npx vercel --prod --confirm
```

- `--prebuilt` 옵션을 사용하면 `npm run build` 결과를 그대로 업로드하므로 동일한 산출물을 검증할 수 있습니다.
- 배포 후 `https://<project>.vercel.app`의 `/?theme=dark` / `/?theme=light` 쿼리로 강제 모드를 테스트하세요.

## 4. 문제 발생 시 진단 포인트

- `next build` 실패 → `docs/pages` 또는 `docs/src/**` 경로를 확인하고, 필요한 경우 `docs/pages/index.mdx`가 올바르게 `../index.md`를 가져오는지 점검합니다.
- 테마 토글 불가 → `localStorage.moai-adk-docs-theme` 초기화 후 다시 시도하거나, `nextThemes` 설정을 확인합니다.
- Vercel 배포 실패 → `.vercel` 폴더 삭제 후 `vercel login` → `vercel link`를 통해 프로젝트를 다시 매핑합니다.
- CSS 누락 → `docs/stylesheets` 내 전역 CSS를 `next.config.cjs`에서 import하고 있는지 확인합니다.
