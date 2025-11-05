# @SPEC:DOCS-001: VitePress 문서 사이트 구축 - 수락 기준

## Given-When-Then 테스트 시나리오

이 문서는 SPEC-DOCS-001의 수락 기준을 Given-When-Then 형식으로 정의합니다.

---

## 시나리오 1: VitePress 기본 설정 및 빌드

### Given (전제 조건)
- Node.js ≥18 또는 Bun ≥1.2가 설치되어 있음
- MoAI-ADK 프로젝트 루트 디렉토리에서 작업 중
- package.json에 VitePress 의존성이 추가됨

### When (실행 조건)
- 개발자가 `bun run docs:dev` 명령을 실행함
- 브라우저에서 `http://localhost:5173`을 접속함

### Then (예상 결과)
- VitePress 개발 서버가 정상적으로 시작됨
- 5173 포트가 열리고 "VitePress server is running" 메시지 출력
- 브라우저에서 홈페이지(index.md)가 정상적으로 렌더링됨
- Alfred 로고와 타이틀이 표시됨
- Sidebar 네비게이션이 8개 섹션으로 구성됨

### 검증 방법
```bash
# 개발 서버 시작
bun run docs:dev

# 기대 출력:
# vite v5.x.x dev server running at:
# > Local: http://localhost:5173/
# > Network: use --host to expose
```

---

## 시나리오 2: 홈페이지 콘텐츠 검증

### Given
- VitePress 개발 서버가 실행 중
- `docs/index.md` 파일이 작성됨
- README.md 1-85줄의 내용이 마이그레이션됨

### When
- 사용자가 `http://localhost:5173`을 방문함

### Then
- **Hero 섹션**:
  - 타이틀: "MoAI-ADK (Agentic Development Kit)"
  - 서브타이틀: "SPEC-First TDD Development with Alfred SuperAgent"
  - CTA 버튼: "Getting Started", "View on GitHub"
- **Features 섹션**:
  - 일관성(Consistency) 카드
  - 품질(Quality) 카드
  - 추적성(Traceability) 카드
  - 범용성(Universality) 카드
- **Quick Links 섹션**:
  - Getting Started 링크
  - Concepts 링크
  - Examples 링크

### 검증 방법
```bash
# 홈페이지 렌더링 확인 (Visual Check)
open http://localhost:5173

# index.md 내용 확인
cat docs/index.md | grep "MoAI-ADK"
cat docs/index.md | grep "Alfred SuperAgent"
```

---

## 시나리오 3: Sidebar 네비게이션 동작

### Given
- VitePress 개발 서버가 실행 중
- `docs/.vitepress/config.ts`에 Sidebar 설정이 완료됨
- 8개 디렉토리에 페이지가 작성됨

### When
- 사용자가 Sidebar에서 "Getting Started" 링크를 클릭함

### Then
- 페이지가 SPA 방식으로 전환됨 (새로고침 없음)
- `docs/guide/getting-started.md` 내용이 렌더링됨
- Sidebar에서 "Getting Started" 링크가 하이라이트됨 (파란색)
- URL이 `/guide/getting-started`로 변경됨

### 검증 방법
```bash
# config.ts Sidebar 설정 확인
cat docs/.vitepress/config.ts | grep "sidebar"

# 예상 구조:
# sidebar: [
#   { text: 'Guide', items: [...] },
#   { text: 'Concepts', items: [...] },
#   ...
# ]
```

---

## 시나리오 4: 검색 기능 동작

### Given
- VitePress local search provider가 활성화됨
- 최소 5개 페이지가 작성됨 (Phase 1 완료)
- 검색 인덱스가 빌드됨

### When
- 사용자가 `Cmd/Ctrl + K`를 누름
- 검색창에 "Alfred"를 입력함

### Then
- 검색창이 모달 형태로 열림
- "Alfred" 키워드가 포함된 페이지 목록이 표시됨
  - index.md (Alfred SuperAgent)
  - concepts/alfred-superagent.md
  - guide/what-is-moai-adk.md
- 검색 결과 표시 시간 < 500ms

### 검증 방법
```bash
# config.ts 검색 설정 확인
cat docs/.vitepress/config.ts | grep "search"

# 예상 설정:
# search: {
#   provider: 'local'
# }
```

---

## 시나리오 5: 핫 리로드 기능

### Given
- VitePress 개발 서버가 실행 중
- 브라우저에서 특정 페이지를 열람 중
- `docs/guide/getting-started.md` 파일이 편집기에서 열려 있음

### When
- 개발자가 `getting-started.md` 파일의 내용을 수정함
- 파일을 저장함 (Cmd/Ctrl + S)

### Then
- 3초 이내에 브라우저가 자동으로 리로드됨
- 수정된 내용이 즉시 반영됨
- 개발자가 수동으로 새로고침할 필요 없음
- 콘솔에 "page reload" 메시지 출력 (선택적)

### 검증 방법
```bash
# 파일 수정 후 핫 리로드 확인 (Manual Test)
echo "\n## Test Section" >> docs/guide/getting-started.md

# 브라우저에서 자동 리로드 확인 (Visual Check)
# 예상: 3초 이내 페이지 업데이트
```

---

## 시나리오 6: 프로덕션 빌드 성공

### Given
- VitePress가 설치됨
- 모든 Phase 1 페이지가 작성됨 (5개)
- 내부 링크가 유효함
- 이미지 파일이 `docs/public/`에 존재함

### When
- 개발자가 `bun run docs:build` 명령을 실행함

### Then
- 빌드 프로세스가 시작됨
- 모든 Markdown 파일이 HTML로 변환됨
- CSS/JS 번들링이 완료됨
- `docs/.vitepress/dist/` 디렉토리가 생성됨
- 빌드 에러 0개
- 빌드 완료 시간 < 30초

### 검증 방법
```bash
# 프로덕션 빌드 실행
bun run docs:build

# 기대 출력:
# building client + server bundles...
# ✓ built in XXXms
# build complete. dist folder is ready to be deployed.

# dist 폴더 확인
ls -la docs/.vitepress/dist/
# 예상: index.html, assets/, guide/, concepts/, ...
```

---

## 시나리오 7: 링크 유효성 검증

### Given
- 모든 페이지가 작성됨
- 내부 링크가 Markdown 형식으로 작성됨
  - `[Getting Started](/guide/getting-started)`
  - `[SPEC Template](../reference/spec-template)`

### When
- 개발자가 `bun run docs:build` 명령을 실행함
- VitePress가 링크 유효성을 검증함

### Then
- 모든 내부 링크가 유효함 (404 없음)
- 외부 링크는 검증하지 않음 (선택적)
- 깨진 링크가 발견되면 빌드 경고 출력
- 빌드는 계속 진행됨 (중단하지 않음)

### 검증 방법
```bash
# 빌드 후 링크 검증
bun run docs:build 2>&1 | grep "404"

# 기대 출력: (없음)
# 깨진 링크가 있으면:
# ✖ [404] /guide/non-existent-page
```

---

## 시나리오 8: Dark Mode 전환

### Given
- VitePress Dark Mode가 활성화됨 (기본 설정)
- 사용자가 라이트 모드에서 페이지를 열람 중

### When
- 사용자가 Navbar의 테마 토글 버튼을 클릭함

### Then
- 페이지가 다크 모드로 전환됨
- 배경색이 어두운 색으로 변경됨
- 텍스트 색상이 밝은 색으로 변경됨
- 사용자 선택이 localStorage에 저장됨
- 다음 방문 시 선택한 테마가 유지됨

### 검증 방법
```bash
# config.ts 테마 설정 확인
cat docs/.vitepress/config.ts | grep "appearance"

# 예상 설정:
# appearance: true  # 또는 'dark' | 'light'
```

---

## 시나리오 9: 콘텐츠 소스 비율 검증

### Given
- Phase 4 완료 (58개 페이지 작성)
- README.md에서 330줄 활용 목표
- development-guide.md에서 235줄 활용 목표
- 신규 작성 10% 목표

### When
- 개발자가 콘텐츠 소스 추적 스크립트를 실행함
- 각 페이지의 Front Matter에 `source` 필드가 명시됨

### Then
- README.md 활용률 ≥ 30% (330줄 / 1097줄)
- development-guide.md 활용률 ≥ 60% (235줄 / 391줄)
- 신규 작성 비율 ≤ 10%
- 비율이 기준 미달 시 경고 출력

### 검증 방법
```bash
# Front Matter 예시 (각 페이지 상단)
# ---
# title: Getting Started
# source: README.md:87-156
# ---

# 콘텐츠 소스 집계 스크립트 (예시)
rg "^source: README" docs/ | wc -l  # README 소스 페이지 수
rg "^source: development-guide" docs/ | wc -l  # dev-guide 소스 페이지 수
```

---

## 시나리오 10: 검색 인덱스 커버리지

### Given
- 모든 페이지가 작성됨 (58개)
- VitePress local search가 활성화됨
- 빌드가 완료됨

### When
- 사용자가 다양한 키워드로 검색함:
  - "Alfred" - 10개 이상 결과
  - "TRUST" - 5개 이상 결과
  - "@TAG" - 5개 이상 결과
  - "SPEC" - 20개 이상 결과

### Then
- 모든 키워드가 적절한 페이지를 반환함
- 검색 결과가 관련성 순으로 정렬됨
- 제목 매칭이 본문 매칭보다 우선순위 높음
- 검색 결과 표시 시간 < 500ms

### 검증 방법
```bash
# 검색 인덱스 파일 확인
ls docs/.vitepress/dist/assets/search*.js

# 인덱스 크기 확인 (예상: < 500KB)
du -h docs/.vitepress/dist/assets/search*.js
```

---

## 품질 게이트 기준

### 빌드 품질
- ✅ VitePress 빌드 성공 (에러 0개)
- ✅ 경고 < 5개
- ✅ 빌드 시간 < 30초
- ✅ dist 폴더 크기 < 10MB

### 링크 품질
- ✅ 깨진 내부 링크 0개
- ✅ 모든 이미지 경로 유효
- ✅ 외부 링크 응답 시간 < 3초 (선택적)

### 콘텐츠 품질
- ✅ 맞춤법 검사 통과
- ✅ 코드 예제 문법 오류 0개
- ✅ 페이지당 파일 크기 < 50KB (권장)
- ✅ 이미지 최적화 (WebP 변환 권장)

### 성능 품질
- ✅ 페이지 로딩 시간 < 2초
- ✅ 검색 속도 < 500ms
- ✅ 핫 리로드 < 3초
- ✅ Lighthouse 성능 점수 ≥ 90

### 접근성 품질
- ✅ 모든 이미지에 alt 텍스트 존재
- ✅ 제목 계층 구조 올바름 (h1 → h2 → h3)
- ✅ 대비율 AAA 등급 (다크 모드 포함)
- ✅ 키보드 네비게이션 가능

---

## Definition of Done (완료 조건)

### Phase 1 완료 조건
- ✅ 5개 핵심 페이지 작성 (index, getting-started, what-is-moai-adk, spec-first-tdd, faq)
- ✅ VitePress 개발 서버 정상 동작
- ✅ VitePress 빌드 성공
- ✅ Sidebar 기본 구조 완성

### Phase 2 완료 조건
- ✅ 11개 페이지 추가 작성 (concepts 5개, installation 6개)
- ✅ Sidebar 구성 완료 (guide, concepts, installation)
- ✅ 내부 링크 유효성 검증
- ✅ 검색 기능 정상 동작

### Phase 3 완료 조건
- ✅ 29개 페이지 추가 작성 (CLI 10개, 언어 10개, 예제 4개, 레퍼런스 5개)
- ✅ 모든 CLI 명령어 문서화
- ✅ 언어별 가이드 10개 완료
- ✅ 실전 예제 4개 완료

### Phase 4 완료 조건
- ✅ 13개 페이지 추가 작성 (문제 해결 4개, 고급 4개, 기여 5개)
- ✅ 모든 섹션 완료 (8개 디렉토리)
- ✅ 콘텐츠 소스 비율 검증 (README 30%, dev-guide 60%, 신규 10%)
- ✅ 전체 품질 게이트 통과

---

## 수락 테스트 체크리스트

### 기능 테스트
- [ ] 홈페이지 렌더링 정상
- [ ] Sidebar 네비게이션 동작
- [ ] 검색 기능 정상 동작
- [ ] 핫 리로드 기능 정상
- [ ] Dark Mode 전환 정상
- [ ] 내부 링크 모두 유효
- [ ] 외부 링크 응답 정상 (선택적)

### 빌드 테스트
- [ ] 개발 서버 정상 시작 (`docs:dev`)
- [ ] 프로덕션 빌드 성공 (`docs:build`)
- [ ] 빌드 에러 0개
- [ ] 빌드 경고 < 5개
- [ ] dist 폴더 생성 확인

### 콘텐츠 테스트
- [ ] Phase 1 페이지 5개 작성
- [ ] Phase 2 페이지 11개 작성
- [ ] Phase 3 페이지 29개 작성
- [ ] Phase 4 페이지 13개 작성
- [ ] 총 58개 페이지 작성 완료
- [ ] 콘텐츠 소스 비율 검증 통과

### 품질 테스트
- [ ] 맞춤법 검사 통과
- [ ] 코드 예제 문법 확인
- [ ] 이미지 최적화 완료
- [ ] Lighthouse 점수 ≥ 90
- [ ] 접근성 검증 통과

---

**작성일**: 2025-10-06
**버전**: 0.1.0
**관련 SPEC**: @SPEC:DOCS-001
