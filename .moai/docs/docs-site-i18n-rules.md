# docs-site 4개국어 문서 동기화 규칙

> Externalized verbatim from CLAUDE.local.md §17 on 2026-05-20 (v2.20.0-rc1 release-readiness consolidation). Original section authored over multiple iterations 2026-04~05.

---


`docs-site/`는 `adk.mo.ai.kr` 공식 사용자 문서. moai-adk-go 개발자는 코드 변경 시 이 섹션을 반드시 준수.

### §17.1 URL 표준 & 블랙리스트

공식 URL은 단 하나:

| 리소스 | URL |
|--------|-----|
| 문서 홈페이지 | `https://adk.mo.ai.kr` |
| GitHub | `https://github.com/modu-ai/moai-adk` |
| Discord | `https://discord.gg/moai-adk` |
| NPM | `https://www.npmjs.com/package/moai-adk` |

### [HARD] 금지 URL (사용 시 CI 실패)

- `docs.moai-ai.dev` (구 주소)
- `adk.moai.com` (오타)
- `adk.moai.kr` (오타: 정확히 `adk.mo.ai.kr`)

### URL 변경 체크리스트 (docs-site 내)

- `docs-site/hugo.yaml` → `baseURL`, `params.og`
- `docs-site/static/robots.txt` → `Sitemap:` 라인
- `docs-site/static/llms.txt` → 문서 링크
- `docs-site/layouts/partials/head.html` → 메타 태그
- `docs-site/vercel.json` → redirects

### §17.2 Markdown/Hextra 작성 규칙

**[HARD] 강조 표기와 괄호 사이 공백 필수** (Nextra 시대 유래, Hextra/Goldmark 검증 완료 후 유지 여부 재평가):

- 잘못: `**바이브코딩(Vibe Coding)**`
- 올바름: `**바이브코딩** (Vibe Coding)`

**[HARD] Mermaid 방향: TD only**

- 사용: `flowchart TD`, `graph TB`
- 금지: `flowchart LR`, `graph LR`

이유: 모바일 가독성 + 사이드바 좁은 화면 대응.

### §17.3 4개국어 동기화 규칙

### [HARD] Canonical source는 ko

번역 체인: `ko` → `en` → `zh`/`ja` 병렬. ko-only 머지 금지.

### 지원 locale (4개 고정)

- `ko` (한국어, 기본) / `en` (영어) / `ja` (일본어) / `zh` (중국어 간체)

### [HARD] 동시 업데이트 의무

기능 추가/변경/제거로 ko 문서가 바뀌면 **동일 PR 내에서 en/zh/ja 모두 반영**. 예외: 5,000 단어 이상 대규모 콘텐츠는 ko 머지 후 48h 이내 en, 추가 72h 이내 zh/ja. 해당 파일 front matter에 `translation_status: pending` 필수.

### PR 체크리스트

- [ ] ko/en/zh/ja 4개 locale 해당 파일 모두 수정
- [ ] 각 locale front matter의 `title`, `description`, `weight` 대응
- [ ] `CHANGELOG.md` 반영
- [ ] 버전 스냅샷 필요 여부 판단 (§17.4)
- [ ] Mermaid 노드 라벨 번역 (구문/방향 미변경)
- [ ] Anchor 링크가 번역된 heading slug 가리킴
- [ ] §17.1 금지 URL 재도입 없음
- [ ] 본문에 이모지 유입 없음

### 불일치 탐지 스크립트

`scripts/docs-i18n-check.sh` (Phase 5 산출) 실행 — pre-commit + CI에서 자동:

- 4개 locale 간 파일 개수/경로 일치
- 각 파일 front matter `title` 존재 (비어있지 않음)
- H1 heading 존재
- MoAI 용어집 (glossary) 준수 — "MoAI-ADK" 등은 모든 locale에서 불번역 유지

### §17.4 버전 스냅샷

| 릴리스 타입 | 스냅샷 생성 | 대상 |
|---------------|----------------|------|
| Major (vX.0.0) | YES | 4 locale 전체 `content/{locale}/v{X}/` 동결 복사 |
| Minor (v2.Y.0) | YES | 동일 |
| Patch (v2.12.Z) | NO | `latest`만 업데이트 |

릴리스 태그 푸시 → `manager-git` 트리거 → `scripts/docs-release-snapshot.go` 실행 → 과거 버전 URL 자동 생성.

사이트 상단 배너: 과거 버전 조회 시 "Viewing v2.X. Go to latest" 자동 표시.

### §17.5 실행 주체

- **기본**: `manager-docs` subagent (`/moai sync` 워크플로우)
- **번역**: `manager-docs` → en 완료 확인 → zh/ja 병렬 위임 (expert-docs)
- **검증**: `plan-auditor` — §17.3 스크립트 결과로 PR pass/fail
- **스냅샷**: `manager-git` — 릴리스 태그 자동 후처리

docs-site 파일을 직접 수정하는 에이전트 (`expert-frontend`, `manager-docs` 등)는 `isolation: worktree` 사용 권장.

### §17.6 빌드/배포 체크리스트

- [ ] `cd docs-site && hugo --minify` 성공 (zero warnings)
- [ ] `docs-site/public/sitemap.xml` 생성 확인
- [ ] Vercel Preview URL 수동 검증 — 4개 locale 홈 페이지
- [ ] Mermaid 다이어그램 렌더링 (light/dark mode)
- [ ] `robots.txt`, `llms.txt` 도메인 일치
- [ ] 언어 스위처 작동 (locale A → locale B navigation)

### [HARD] Vercel 프로젝트 바인딩

- 프로젝트 ID: `prj_EZaVdfE3gJeXVbizafBEECpniINP` (moai-docs에서 계승, 절대 변경 금지)
- 도메인: `adk.mo.ai.kr` → production branch 배포
- Root Directory: `docs-site/` (Vercel Dashboard 설정)
- Framework Preset: Hugo
- Node.js version: Hugo는 Node 불요, 단 `vercel.json`에서 Hugo 버전 명시 필수

---

