# @SPEC:DOCS-001 인수 기준

## 1. Acceptance Overview (인수 개요)

본 문서는 SPEC-DOCS-001의 완료 기준과 검증 시나리오를 정의합니다.

**완료 조건 (Definition of Done)**:
- VitePress 관련 모든 파일 제거 완료
- docs 9개 카테고리 구조 생성 완료
- README.md 간소화 완료
- GitHub Pages 정상 렌더링 확인

---

## 2. Acceptance Scenarios (인수 시나리오)

### Scenario 1: VitePress 제거 및 순수 마크다운 전환

**Given**: 프로젝트에 VitePress 기반 문서 시스템이 존재하는 상태
**When**: VitePress 관련 파일을 제거하고 순수 마크다운으로 전환
**Then**:
- `.vitepress/` 디렉토리가 존재하지 않음
- `package.json`에 `vitepress` 의존성이 없음
- `npm install` 실행 시 vitepress 관련 에러가 발생하지 않음
- `npm run build` 정상 실행됨

**검증 방법**:
```bash
# 디렉토리 존재 확인
test ! -d .vitepress && echo "✅ .vitepress 제거 완료" || echo "❌ .vitepress 여전히 존재"

# package.json 의존성 확인
grep -q "vitepress" package.json && echo "❌ vitepress 의존성 존재" || echo "✅ vitepress 의존성 제거 완료"

# npm 스크립트 확인
grep -q "docs:dev\|docs:build" package.json && echo "❌ docs 스크립트 존재" || echo "✅ docs 스크립트 제거 완료"

# 빌드 테스트
npm install && npm run build
```

---

### Scenario 2: README.md 간소화 및 docs 연결

**Given**: 기존 README.md에 장황한 설명과 VitePress 안내가 포함된 상태
**When**: README.md를 간소화하고 docs 링크를 추가
**Then**:
- README.md 길이가 50-70 LOC 이내
- 프로젝트 한 줄 소개 포함
- 빠른 시작 섹션 포함 (3-5줄)
- `docs/index.md` 링크 포함
- `CLAUDE.md` 링크 포함
- 라이선스 정보 포함

**검증 방법**:
```bash
# LOC 확인
wc -l README.md | awk '{if ($1 <= 70) print "✅ README.md 길이 적절:", $1, "LOC"; else print "❌ README.md 너무 김:", $1, "LOC"}'

# 필수 섹션 확인
grep -q "빠른 시작\|Quick Start" README.md && echo "✅ 빠른 시작 섹션 존재" || echo "❌ 빠른 시작 섹션 없음"
grep -q "docs/index.md" README.md && echo "✅ docs 링크 존재" || echo "❌ docs 링크 없음"
grep -q "CLAUDE.md" README.md && echo "✅ CLAUDE.md 링크 존재" || echo "❌ CLAUDE.md 링크 없음"
grep -q "MIT\|License" README.md && echo "✅ 라이선스 정보 존재" || echo "❌ 라이선스 정보 없음"
```

---

### Scenario 3: GitHub Pages 호환 구조 및 카테고리별 인덱스

**Given**: docs 디렉토리가 비어있거나 VitePress 구조인 상태
**When**: 9개 카테고리 기반 docs 구조를 생성하고 GitHub Pages 설정 추가
**Then**:
- `docs/` 디렉토리에 9개 카테고리 존재
  - getting-started, concepts, alfred, cli, api, guides, agents, examples, contributing
- 각 카테고리에 `index.md` 존재
- `docs/index.md` 메인 허브 존재
- `docs/_config.yml` 존재
- GitHub Pages 설정 완료 (Repository Settings)
- 모든 마크다운 파일이 웹에서 정상 렌더링

**검증 방법**:
```bash
# 카테고리 디렉토리 확인
for dir in getting-started concepts alfred cli api guides agents examples contributing; do
  test -d "docs/$dir" && echo "✅ docs/$dir 존재" || echo "❌ docs/$dir 없음"
done

# index.md 확인
for dir in getting-started concepts alfred cli api guides agents examples contributing; do
  test -f "docs/$dir/index.md" && echo "✅ docs/$dir/index.md 존재" || echo "❌ docs/$dir/index.md 없음"
done

# 메인 허브 확인
test -f "docs/index.md" && echo "✅ docs/index.md 존재" || echo "❌ docs/index.md 없음"

# Jekyll 설정 확인
test -f "docs/_config.yml" && echo "✅ docs/_config.yml 존재" || echo "❌ docs/_config.yml 없음"

# 로컬 Jekyll 테스트
cd docs && bundle exec jekyll serve --detach
sleep 3
curl -s http://localhost:4000/index.html > /dev/null && echo "✅ 로컬 렌더링 성공" || echo "❌ 로컬 렌더링 실패"
pkill -f jekyll
```

---

## 3. Quality Gates (품질 게이트)

### 3.1 문서 무결성

**기준**:
- 기존 CLAUDE.md, development-guide.md의 핵심 콘텐츠가 누락 없이 docs로 이전됨
- 모든 내부 링크가 정상 작동함
- 외부 링크가 정상 작동함

**검증 스크립트**:
```bash
# 내부 링크 검증 (상대 경로)
rg '\[.*\]\(\.\.?/.*\.md\)' -n docs/ | while read -r line; do
  file=$(echo "$line" | cut -d: -f1)
  link=$(echo "$line" | grep -oP '\(\.\.?/.*\.md\)' | tr -d '()')
  dir=$(dirname "$file")
  target="$dir/$link"
  test -f "$target" && echo "✅ $target 존재" || echo "❌ $target 깨진 링크"
done

# 외부 링크 검증 (http/https)
rg '\[.*\]\(https?://.*\)' -n docs/ | while read -r line; do
  link=$(echo "$line" | grep -oP 'https?://[^\)]+')
  curl -s -o /dev/null -w "%{http_code}" "$link" | grep -q "200" && echo "✅ $link 정상" || echo "❌ $link 에러"
done
```

### 3.2 콘텐츠 품질

**기준**:
- 각 카테고리 `index.md`에 최소 3개 이상의 문서 링크 포함
- 각 문서에 명확한 제목(H1)과 설명 포함
- 코드 블록에 언어 지정 (syntax highlighting)

**검증 방법**:
```bash
# index.md 링크 개수 확인
for idx in docs/*/index.md; do
  count=$(grep -c '\[.*\](.*\.md)' "$idx")
  test $count -ge 3 && echo "✅ $idx: $count 링크" || echo "⚠️ $idx: $count 링크 (3개 미만)"
done

# H1 제목 확인
for md in docs/**/*.md; do
  grep -q '^# ' "$md" && echo "✅ $md: H1 존재" || echo "❌ $md: H1 없음"
done

# 코드 블록 언어 지정 확인
rg '```[^a-z]' -n docs/ && echo "⚠️ 언어 미지정 코드 블록 발견" || echo "✅ 모든 코드 블록 언어 지정 완료"
```

### 3.3 GitHub Pages 호환성

**기준**:
- Jekyll이 모든 마크다운 파일을 정상 렌더링
- 네비게이션 링크가 웹에서 정상 작동
- 모바일 반응형 레이아웃 지원

**검증 방법**:
```bash
# GitHub Pages 빌드 로그 확인
gh api repos/:owner/:repo/pages/builds/latest --jq '.status' | grep -q "built" && echo "✅ GitHub Pages 빌드 성공" || echo "❌ GitHub Pages 빌드 실패"

# 웹 접근성 확인 (배포 후)
curl -s https://<username>.github.io/<repo>/docs/ > /dev/null && echo "✅ 웹 접근 가능" || echo "❌ 웹 접근 불가"
```

---

## 4. Test Coverage (테스트 커버리지)

### 4.1 구조 검증 테스트

**테스트 파일**: `tests/docs/structure.test.ts`

```typescript
// @TEST:DOCS-001 | SPEC: SPEC-DOCS-001.md
import { describe, it, expect } from 'vitest'
import { existsSync, readdirSync } from 'fs'
import { join } from 'path'

describe('@TEST:DOCS-001 문서 구조 검증', () => {
  const docsRoot = join(__dirname, '../../docs')
  const categories = [
    'getting-started', 'concepts', 'alfred', 'cli', 'api',
    'guides', 'agents', 'examples', 'contributing'
  ]

  it('docs 디렉토리가 존재해야 한다', () => {
    expect(existsSync(docsRoot)).toBe(true)
  })

  it('9개 카테고리 디렉토리가 존재해야 한다', () => {
    categories.forEach(cat => {
      expect(existsSync(join(docsRoot, cat))).toBe(true)
    })
  })

  it('각 카테고리에 index.md가 존재해야 한다', () => {
    categories.forEach(cat => {
      expect(existsSync(join(docsRoot, cat, 'index.md'))).toBe(true)
    })
  })

  it('docs/index.md 메인 허브가 존재해야 한다', () => {
    expect(existsSync(join(docsRoot, 'index.md'))).toBe(true)
  })

  it('docs/_config.yml이 존재해야 한다', () => {
    expect(existsSync(join(docsRoot, '_config.yml'))).toBe(true)
  })
})
```

### 4.2 링크 무결성 테스트

**테스트 파일**: `tests/docs/links.test.ts`

```typescript
// @TEST:DOCS-001 | SPEC: SPEC-DOCS-001.md
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'fs'
import { join, dirname, resolve } from 'path'
import { glob } from 'glob'

describe('@TEST:DOCS-001 링크 무결성 검증', () => {
  const markdownFiles = glob.sync('docs/**/*.md')

  it('모든 내부 링크가 유효해야 한다', () => {
    const brokenLinks: string[] = []

    markdownFiles.forEach(file => {
      const content = readFileSync(file, 'utf-8')
      const linkRegex = /\[.*?\]\((\.\.?\/[^)]+\.md)\)/g
      let match

      while ((match = linkRegex.exec(content)) !== null) {
        const link = match[1]
        const targetPath = resolve(dirname(file), link)
        if (!existsSync(targetPath)) {
          brokenLinks.push(`${file}: ${link}`)
        }
      }
    })

    expect(brokenLinks).toEqual([])
  })
})
```

### 4.3 VitePress 제거 검증 테스트

**테스트 파일**: `tests/docs/cleanup.test.ts`

```typescript
// @TEST:DOCS-001 | SPEC: SPEC-DOCS-001.md
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'fs'
import { join } from 'path'

describe('@TEST:DOCS-001 VitePress 제거 검증', () => {
  it('.vitepress 디렉토리가 존재하지 않아야 한다', () => {
    expect(existsSync('.vitepress')).toBe(false)
  })

  it('package.json에 vitepress 의존성이 없어야 한다', () => {
    const pkg = JSON.parse(readFileSync('package.json', 'utf-8'))
    expect(pkg.dependencies?.vitepress).toBeUndefined()
    expect(pkg.devDependencies?.vitepress).toBeUndefined()
  })

  it('package.json에 docs 스크립트가 없어야 한다', () => {
    const pkg = JSON.parse(readFileSync('package.json', 'utf-8'))
    expect(pkg.scripts?.['docs:dev']).toBeUndefined()
    expect(pkg.scripts?.['docs:build']).toBeUndefined()
    expect(pkg.scripts?.['docs:preview']).toBeUndefined()
  })
})
```

---

## 5. Rollback Plan (롤백 계획)

### 롤백 트리거 조건

- GitHub Pages 빌드가 3회 연속 실패
- 내부 링크 20% 이상 깨짐
- 사용자 피드백에서 치명적 버그 보고

### 롤백 절차

1. **VitePress 복구**:
   ```bash
   git revert HEAD~3..HEAD  # 최근 3개 커밋 되돌리기
   npm install
   npm run docs:dev         # VitePress 정상 작동 확인
   ```

2. **백업 복원**:
   ```bash
   cp -r .moai/backup/docs/* docs/
   cp .moai/backup/README.md README.md
   ```

3. **GitHub Pages 설정 복원**:
   - Repository Settings → Pages
   - Source: GitHub Actions
   - Workflow: `.github/workflows/deploy-docs.yml` 재활성화

---

## 6. Definition of Done (완료 정의)

**SPEC-DOCS-001은 다음 조건을 모두 충족해야 완료로 간주됩니다**:

- [ ] VitePress 관련 파일 완전 제거 확인
- [ ] docs 9개 카테고리 구조 생성 및 index.md 작성 완료
- [ ] README.md 간소화 완료 (50-70 LOC)
- [ ] 모든 내부 링크 검증 통과 (0개 깨진 링크)
- [ ] 로컬 Jekyll 서버 정상 렌더링 확인
- [ ] GitHub Pages 설정 완료 및 웹 접근 가능
- [ ] 3개 테스트 스위트 통과 (structure, links, cleanup)
- [ ] TRUST 5원칙 준수 (테스트 커버리지 ≥ 85%)
- [ ] @TAG 체인 무결성 (`@SPEC:DOCS-001` → `@TEST:DOCS-001` → `@CODE:DOCS-001`)
