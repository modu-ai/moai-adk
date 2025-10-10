# @SPEC:DOCS-001 구현 계획

## 1. Implementation Strategy (구현 전략)

### 우선순위 기반 마일스톤

**1차 목표: VitePress 제거 및 기본 구조 생성**
- VitePress 관련 모든 파일 제거
- `docs/` 디렉토리에 9개 카테고리 기본 구조 생성
- 각 카테고리 `index.md` 생성 (목차 템플릿)

**2차 목표: 핵심 콘텐츠 마이그레이션**
- CLAUDE.md → `docs/alfred/` 분리
- development-guide.md → `docs/guides/`, `docs/concepts/` 분리
- README.md 간소화

**3차 목표: 링크 무결성 및 GitHub Pages 설정**
- 모든 내부 링크 검증 및 업데이트
- GitHub Pages 설정 파일 생성
- 로컬 렌더링 테스트

**최종 목표: 배포 및 검증**
- GitHub Pages 배포
- 웹 접근성 확인
- 문서 네비게이션 UX 검증

---

## 2. Technical Approach (기술적 접근 방법)

### 2.1 파일 제거 전략

**안전한 제거 순서**:
1. `.github/workflows/deploy-docs.yml` 제거 (CI/CD 충돌 방지)
2. `package.json` 스크립트 수정 (docs 관련 명령어 제거)
3. `package.json` 의존성 제거 (vitepress, vue)
4. `.vitepress/` 디렉토리 삭제

**검증 방법**:
```bash
npm install           # 의존성 에러 확인
npm run build         # 빌드 정상 작동 확인
git status            # 삭제 파일 목록 확인
```

### 2.2 docs 구조 생성 전략

**디렉토리 생성 스크립트**:
```bash
mkdir -p docs/{getting-started,concepts,alfred,cli,api/agents,guides,agents,examples,contributing}
```

**index.md 템플릿**:
```markdown
# [카테고리명]

> 카테고리 설명

## 문서 목록

- [문서1](./doc1.md)
- [문서2](./doc2.md)

## 관련 카테고리

- [다른 카테고리](../other-category/index.md)
```

### 2.3 콘텐츠 마이그레이션 전략

**CLAUDE.md → docs/alfred/**:
- `overview.md`: Alfred 페르소나, 9개 에이전트 생태계
- `commands.md`: `/alfred:1-spec`, `/alfred:2-build`, `/alfred:3-sync` 상세
- `orchestration.md`: 오케스트레이션 전략, 작업 분해 및 라우팅

**development-guide.md → docs/guides/ + docs/concepts/**:
- `docs/guides/development-workflow.md`: SPEC-TDD 워크플로우, 체크리스트
- `docs/concepts/spec-first-tdd.md`: SPEC 우선 철학
- `docs/concepts/trust-principles.md`: TRUST 5원칙 상세
- `docs/concepts/tag-system.md`: @TAG Lifecycle, TAG BLOCK 템플릿
- `docs/concepts/ears-methodology.md`: EARS 요구사항 작성법
- `docs/guides/context-engineering.md`: JIT Retrieval, Compaction

**콘텐츠 분리 기준**:
- **Concepts**: 이론, 원칙, 개념 설명
- **Guides**: 실용적 가이드, 워크플로우, 체크리스트
- **Alfred**: Alfred 관련 모든 내용
- **API**: 에이전트별 API 레퍼런스

### 2.4 README.md 간소화 전략

**제거 대상**:
- 장황한 설명 (→ docs로 이동)
- VitePress 관련 안내
- 상세한 워크플로우 (→ docs/guides로 이동)

**유지 대상**:
- 프로젝트 한 줄 소개
- 빠른 시작 (3-5줄)
- 핵심 링크 (docs, CLAUDE.md, getting-started)
- 라이선스 정보

**새 README.md 길이**: 50-70 LOC

### 2.5 GitHub Pages 설정

**_config.yml 생성**:
```yaml
theme: jekyll-theme-minimal
title: MoAI-ADK Documentation
description: SPEC-First TDD Development Kit with Alfred SuperAgent
show_downloads: false
```

**Repository 설정**:
1. Settings → Pages
2. Source: `Deploy from a branch`
3. Branch: `main` (또는 `develop`)
4. Folder: `/docs`
5. Save

**로컬 테스트** (Jekyll 로컬 서버):
```bash
cd docs
bundle exec jekyll serve
# http://localhost:4000 접근
```

---

## 3. Architecture Design (아키텍처 설계)

### 3.1 문서 네비게이션 흐름

```
README.md (프로젝트 소개)
    ↓
docs/index.md (메인 허브)
    ↓
9개 카테고리 index.md (목차)
    ↓
개별 문서 (상세 내용)
    ↓
관련 문서 (내부 링크)
```

### 3.2 링크 구조

**절대 경로 사용** (GitHub Pages 호환):
```markdown
[Alfred 가이드](/docs/alfred/index.md)
[TRUST 원칙](/docs/concepts/trust-principles.md)
```

**상대 경로 사용** (같은 카테고리 내):
```markdown
[설치 가이드](./installation.md)
[빠른 시작](./quick-start.md)
```

### 3.3 카테고리 간 의존성

```
getting-started ← concepts (개념 이해 필요)
concepts ← guides (실습 적용)
alfred → cli (Alfred가 CLI 사용)
guides → api (가이드에서 에이전트 참조)
```

---

## 4. Risk Analysis (리스크 분석)

### 4.1 링크 깨짐 리스크

**원인**: 파일 경로 변경, 파일명 변경
**완화 방안**:
- 마이그레이션 전 모든 링크 스캔: `rg '\[.*\]\(.*\.md\)' -n`
- 마이그레이션 후 링크 검증 스크립트 실행
- 로컬 렌더링 테스트 (Jekyll 서버)

### 4.2 콘텐츠 누락 리스크

**원인**: 마이그레이션 중 일부 섹션 누락
**완화 방안**:
- 원본 파일 백업 (`.moai/backup/`)
- 마이그레이션 전후 LOC 비교
- 체크리스트 기반 섹션 매핑 검증

### 4.3 GitHub Pages 렌더링 실패

**원인**: Jekyll 테마 비호환, 마크다운 문법 에러
**완화 방안**:
- 로컬 Jekyll 테스트 필수
- `_config.yml` 최소 설정 사용
- GitHub Pages 빌드 로그 모니터링

---

## 5. Implementation Checklist (구현 체크리스트)

### Phase 1: VitePress 제거
- [ ] `.github/workflows/deploy-docs.yml` 삭제
- [ ] `package.json` docs 스크립트 제거
- [ ] `package.json` vitepress 의존성 제거
- [ ] `.vitepress/` 디렉토리 삭제
- [ ] `npm install` 정상 실행 확인

### Phase 2: docs 구조 생성
- [ ] 9개 카테고리 디렉토리 생성
- [ ] 각 카테고리 `index.md` 생성
- [ ] `docs/index.md` 메인 허브 생성
- [ ] `docs/_config.yml` 생성

### Phase 3: 콘텐츠 마이그레이션
- [ ] CLAUDE.md → `docs/alfred/` 분리
- [ ] development-guide.md → `docs/guides/`, `docs/concepts/` 분리
- [ ] README.md 간소화
- [ ] 모든 내부 링크 업데이트

### Phase 4: 검증 및 배포
- [ ] 로컬 Jekyll 서버 테스트
- [ ] 링크 무결성 검증 (`rg '\[.*\]\(.*\.md\)' -n`)
- [ ] GitHub Pages 설정 완료
- [ ] 웹 접근성 확인 (https://<username>.github.io/<repo>/docs/)

---

## 6. Next Steps (다음 단계)

### After SPEC-DOCS-001 Completion

1. **`/alfred:2-build SPEC-DOCS-001`**
   - TDD 구현: 마이그레이션 스크립트 작성
   - 링크 검증 테스트 작성
   - 구조 검증 테스트 작성

2. **`/alfred:3-sync`**
   - 문서 동기화
   - TAG 체인 검증
   - PR Ready 전환

3. **GitHub Pages 모니터링**
   - 빌드 로그 확인
   - 렌더링 품질 점검
   - 사용자 피드백 수집
