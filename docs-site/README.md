# MoAI CLI 문서 사이트

Rust `moai-cli`용 한국어·영어 제품 문서입니다. Astro 7 + Starlight 기본 테마, Markdown/MDX, 빌드 시 생성하는 Pagefind 검색을 사용합니다. 서버 어댑터, 외부 검색 API, 분석 서비스나 인증 서비스는 필요하지 않습니다. 기존 Hugo 사이트와 배포 프로젝트는 변경하지 않습니다.

## 로컬 실행

**Node.js 24.x와 npm**이 필요합니다. 저장소 루트에서 다음 명령을 실행합니다. `package-lock.json`을 함께 사용하여 의존성을 재현합니다.

```sh
cd docs-site
npm ci
npm run dev
```

개발 서버가 출력하는 주소에서 `/ko/`와 `/en/`를 엽니다. 기본 언어는 한국어입니다. `/`는 `astro.config.mjs`의 명시적 redirect로 `/ko/`에 연결합니다. 서버 어댑터가 없는 정적 빌드에서는 meta-refresh HTML을 생성하며, HTTP 리디렉션 상태 코드를 보장하지 않습니다.

```sh
npm run check
npm run build
npm run preview
```

`check`는 Astro 진단, `build`는 정적 결과물 `dist/` 생성, `preview`는 빌드 결과의 로컬 확인에 사용합니다. **Pagefind 검색은 개발 서버가 아니라 프로덕션 빌드 후 preview에서 확인**합니다. 내용 변경 후에는 다시 빌드해야 검색 색인에도 반영됩니다.

Astro 7의 preview는 백그라운드 프로세스로 실행될 수 있습니다. 확인을 마친 뒤 같은 디렉터리에서 `npm run preview -- stop`으로 종료합니다.

## 문서 구조와 번역

`src/content/docs/ko/`와 `src/content/docs/en/`에 같은 상대 경로로 문서를 작성합니다. 현재 각 언어의 여섯 페이지는 다음과 같습니다.

- `index.mdx`
- `guides/getting-started.md`
- `guides/claude-code.md`
- `guides/codex.md`
- `reference/cli.md`
- `reference/safety.md`

제목·설명·본문을 함께 번역하고 파일명과 slug는 일치시킵니다. 파일 추가·삭제 시 양쪽 언어와 `astro.config.mjs`의 사이드바를 함께 갱신합니다. 사이드바 slug에는 언어 접두사를 넣지 않습니다. 그룹 라벨은 `translations`로, 페이지 라벨은 각 페이지의 frontmatter 제목으로 현지화합니다. UI는 Starlight의 기본 번역을 사용하되, 영어 코드 블록의 세 문자열은 `src/content/i18n/en.json`에 명시합니다. Starlight 0.42.10의 기본 영어 사전에 이 키들이 없어 한국어 기본 언어로 fallback되기 때문입니다. `i18nLoader`와 `i18nSchema`로 해당 번역을 읽으며, 한국어는 기본 번역을 유지합니다.

홈은 Starlight의 `CardGrid`·`LinkCard`를 쓰는 MDX이며 나머지는 Markdown입니다. `src/content.config.ts`에서 공식 `docsLoader`와 `docsSchema`를 사용합니다. 제품 동작은 현재 Rust 구현과 저장소의 지원표를 기준으로 쓰며, 게시되지 않은 설치 URL이나 미검증 플랫폼 지원을 추가하지 않습니다.

## Vercel 신규 프로젝트 배포 제안 — 미실행

다음은 **별도 승인 후 새 프로젝트에만 적용할 제안**입니다. 이 구현은 배포를 실행하거나 승인하지 않았고, Vercel 프로젝트 연결·설정·도메인·기존 Hugo 배포를 변경하지 않습니다.

- 기존 저장소를 가져오되 기존 Hugo 프로젝트와 분리된 **새 Vercel 프로젝트**를 만듭니다.
- Root Directory: `docs-site`
- Framework Preset: `Astro`
- Node.js: `24.x`
- Install Command: `npm ci`
- Build Command: `npm run build`
- Output Directory: `dist`
- 정적 배포로 구성하며 서버 어댑터는 추가하지 않습니다.

승인된 새 프로젝트에서만 배포를 진행하고, 결과의 양쪽 언어·언어 전환·검색·모바일 탐색을 확인합니다. 운영 도메인은 지정하지 않았습니다. 기존 Hugo preview를 바꾸는 브랜치 override, `vercel.json`, `.vercel` 연결 파일은 만들지 않습니다.

## 구성 근거

- [Starlight 수동 설정](https://starlight.astro.build/manual-setup/)
- [다국어](https://starlight.astro.build/guides/i18n/), [사이드바](https://starlight.astro.build/guides/sidebar/)
- [기본 Pagefind 검색](https://starlight.astro.build/guides/site-search/)
- [Astro 정적 배포](https://docs.astro.build/en/guides/deploy/)
- [Astro redirects 설정](https://docs.astro.build/en/reference/configuration-reference/#redirects)

직접 의존성은 정확한 버전으로 고정합니다. TypeScript는 `@astrojs/check@0.9.12`의 peer 범위(5.x 또는 6.x)에 맞는 `5.9.3`을 사용하며, 현재 latest인 7.x를 무조건 선택하지 않습니다.
