# SPEC-DOCS-003 구현 계획

> **VitePress → MkDocs Material 마이그레이션**
>
> 본 문서는 SPEC-DOCS-003의 구현 계획을 정의합니다.

---

## 전략 개요

### 마이그레이션 접근 방식
- **단계적 전환**: VitePress 유지하면서 MkDocs 병렬 구축 후 전환
- **자동화 우선**: 스크립트로 반복 작업 자동화
- **검증 기반**: 각 단계마다 검증 후 진행

### 우선순위 원칙
1. **핵심 기능 우선**: 빌드 및 배포 성공
2. **호환성 검증**: 모든 링크 및 다이어그램 정상 작동
3. **성능 최적화**: 빌드 시간 단축

---

## Phase 1: MkDocs Material 설정 및 기본 구조

### 1.1 Python 가상환경 설정
**작업 내용**:
```bash
# Python 버전 확인 (3.9 이상)
python3 --version

# 가상환경 생성 (선택사항)
python3 -m venv .venv
source .venv/bin/activate  # macOS/Linux
# .venv\Scripts\activate   # Windows
```

**산출물**:
- `.venv/` 디렉토리 (선택사항, .gitignore에 추가)

### 1.2 requirements.txt 생성
**작업 내용**:
프로젝트 루트에 `requirements.txt` 생성:
```txt
mkdocs>=1.5.0
mkdocs-material>=9.5.0
mkdocs-mermaid2-plugin>=1.1.0
pymdown-extensions>=10.0
```

**산출물**:
- `requirements.txt` 파일

### 1.3 의존성 설치
**작업 내용**:
```bash
pip install -r requirements.txt
```

**검증**:
```bash
mkdocs --version
# → mkdocs, version 1.5.x
```

### 1.4 mkdocs.yml 초기 설정
**작업 내용**:
프로젝트 루트에 `mkdocs.yml` 생성 (최소 설정):
```yaml
site_name: MoAI-ADK Documentation
site_description: SPEC-First TDD 개발 프레임워크
site_author: MoAI Team

theme:
  name: material
  language: ko

plugins:
  - search:
      lang:
        - ko
        - en
```

**검증**:
```bash
mkdocs serve
# → http://127.0.0.1:8000 접속 확인
```

**산출물**:
- `mkdocs.yml` (기본 설정)

---

## Phase 2: Markdown 파일 변환

### 2.1 기존 문서 구조 분석
**작업 내용**:
```bash
# VitePress 문서 파일 목록 조회
find docs -name "*.md" -type f | wc -l
# → 22개 파일 확인

# 디렉토리 구조 확인
tree docs -L 2
```

**산출물**:
- 문서 파일 목록 (`docs-inventory.txt`)

### 2.2 내부 링크 패턴 분석
**작업 내용**:
```bash
# VitePress 절대 경로 링크 검색
rg '\[.*\]\(/guides/' docs/
rg '\[.*\]\(/api/' docs/

# 링크 패턴 추출
rg -o '\[.*\]\(/[^)]+\)' docs/ > links-inventory.txt
```

**산출물**:
- `links-inventory.txt` (링크 패턴 목록)

### 2.3 링크 변환 스크립트 작성
**작업 내용**:
Python 스크립트 작성 (`scripts/convert-links.py`):
```python
#!/usr/bin/env python3
import re
import sys
from pathlib import Path

def convert_vitepress_to_mkdocs(md_file):
    """VitePress 링크를 MkDocs 형식으로 변환"""
    content = md_file.read_text(encoding='utf-8')

    # 절대 경로 → 상대 경로 + .md 확장자
    # [text](/guides/page) → [text](guides/page.md)
    pattern = r'\[([^\]]+)\]\(/([^)]+)\)'
    replacement = r'[\1](\2.md)'

    converted = re.sub(pattern, replacement, content)

    # 변경사항이 있으면 파일 업데이트
    if converted != content:
        md_file.write_text(converted, encoding='utf-8')
        print(f"✅ Converted: {md_file}")
    else:
        print(f"⏭️  Skipped: {md_file}")

def main():
    docs_dir = Path('docs')
    md_files = docs_dir.rglob('*.md')

    for md_file in md_files:
        convert_vitepress_to_mkdocs(md_file)

if __name__ == '__main__':
    main()
```

**실행**:
```bash
python scripts/convert-links.py
```

**산출물**:
- `scripts/convert-links.py` (링크 변환 스크립트)
- 변환된 Markdown 파일 (22개)

### 2.4 Mermaid 다이어그램 검증
**작업 내용**:
```bash
# Mermaid 코드 블록 검색
rg '```mermaid' docs/
```

**검증**:
- Mermaid 코드 블록이 ` ```mermaid ` 형식인지 확인
- 닫는 태그(` ``` `)가 있는지 확인

**산출물**:
- Mermaid 다이어그램 목록

---

## Phase 3: 네비게이션 재구성

### 3.1 VitePress config.mts 분석
**작업 내용**:
```bash
# VitePress 설정 파일 읽기
cat docs/.vitepress/config.mts
```

**분석 대상**:
- `sidebar` 구조
- 각 항목의 `text`, `link` 값
- 중첩 레벨

**산출물**:
- 네비게이션 구조 다이어그램

### 3.2 mkdocs.yml nav 섹션 작성
**작업 내용**:
VitePress의 `sidebar` 구조를 mkdocs.yml의 `nav`로 수동 변환:

**예시**:
```yaml
nav:
  - 홈: index.md
  - 시작하기:
    - Getting Started: guides/getting-started.md
    - Installation: guides/installation.md
    - Quick Start: guides/quick-start.md
  - Alfred 커맨드:
    - 개요: guides/alfred-commands.md
    - SPEC 작성: guides/spec-builder.md
    - TDD 구현: guides/code-builder.md
    - 문서 동기화: guides/doc-syncer.md
  - 가이드:
    - TDD 워크플로우: guides/tdd-workflow.md
    - TAG 시스템: guides/tag-system.md
    - TRUST 원칙: guides/trust-principles.md
  - API 레퍼런스:
    - spec-builder: api/spec-builder.md
    - code-builder: api/code-builder.md
    - doc-syncer: api/doc-syncer.md
    - tag-agent: api/tag-agent.md
    - git-manager: api/git-manager.md
```

**검증**:
```bash
mkdocs serve
# → 네비게이션 구조 확인
```

**산출물**:
- `mkdocs.yml` (nav 섹션 완성)

### 3.3 사이드바 구조 테스트
**작업 내용**:
- 브라우저에서 http://127.0.0.1:8000 접속
- 사이드바의 모든 링크 클릭
- 404 오류 없는지 확인

**산출물**:
- 네비게이션 테스트 결과

---

## Phase 4: Vercel 배포 설정

### 4.1 vercel.json 생성
**작업 내용**:
프로젝트 루트에 `vercel.json` 생성:
```json
{
  "buildCommand": "pip install -r requirements.txt && mkdocs build",
  "outputDirectory": "site",
  "framework": null,
  "devCommand": "mkdocs serve",
  "installCommand": "pip install -r requirements.txt"
}
```

**산출물**:
- `vercel.json` 파일

### 4.2 .gitignore 업데이트
**작업 내용**:
`.gitignore`에 MkDocs 빌드 산출물 추가:
```gitignore
# MkDocs
site/
.venv/
```

**산출물**:
- `.gitignore` (업데이트)

### 4.3 로컬 빌드 테스트
**작업 내용**:
```bash
# 빌드 시간 측정
time mkdocs build

# 빌드 산출물 확인
ls -lh site/

# 정적 HTML 테스트
cd site
python -m http.server 8001
# → http://127.0.0.1:8001 접속 확인
```

**목표**:
- 빌드 시간 < 30초
- site/ 디렉토리에 HTML 파일 생성

**산출물**:
- `site/` 디렉토리 (빌드 산출물)

### 4.4 Vercel 배포 (Dry Run)
**작업 내용**:
```bash
# Vercel CLI 설치 (선택사항)
npm install -g vercel

# Vercel 로그인
vercel login

# 배포 테스트 (프로덕션 아님)
vercel --prod=false
```

**검증**:
- 배포 URL 접속
- 빌드 로그 확인
- 빌드 시간 측정

**산출물**:
- Vercel 배포 URL (테스트)

---

## Phase 5: 검증 및 최적화

### 5.1 빌드 시간 벤치마크
**작업 내용**:
```bash
# VitePress 빌드 시간 측정 (비교 기준)
time npm run docs:build

# MkDocs 빌드 시간 측정
time mkdocs build
```

**목표**:
- MkDocs 빌드 시간이 VitePress의 50% 이하

**산출물**:
- 빌드 시간 비교 표

### 5.2 404 링크 검사 (Strict Mode)
**작업 내용**:
```bash
# MkDocs strict mode 빌드
mkdocs build --strict
```

**예상 결과**:
- 깨진 링크가 있으면 빌드 실패
- 오류 메시지로 문제 파일 식별

**산출물**:
- 링크 무결성 검증 보고서

### 5.3 검색 기능 테스트
**작업 내용**:
1. `mkdocs serve` 실행
2. 브라우저에서 검색창 열기 (Ctrl/Cmd + K)
3. 다음 키워드로 테스트:
   - "Alfred" (영문)
   - "명세" (한글)
   - "TDD" (약어)
   - "@SPEC" (특수문자)

**검증 기준**:
- 검색 결과가 즉시 표시
- 한글 검색이 정상 작동
- 검색 결과 클릭 시 해당 페이지로 이동

**산출물**:
- 검색 기능 테스트 결과

### 5.4 Mermaid 다이어그램 렌더링 테스트
**작업 내용**:
1. Mermaid 다이어그램이 포함된 페이지 접속
2. 라이트 모드 / 다크 모드 전환
3. 다이어그램이 정상 렌더링되는지 확인

**검증 기준**:
- 다이어그램이 이미지로 렌더링됨
- 다크 모드에서도 가독성 유지

**산출물**:
- Mermaid 렌더링 테스트 결과

### 5.5 모바일 반응형 테스트
**작업 내용**:
1. 브라우저 개발자 도구 열기
2. 모바일 뷰포트로 전환 (iPhone, Android)
3. 네비게이션 메뉴 작동 확인

**검증 기준**:
- 모바일 햄버거 메뉴 정상 작동
- 검색 기능 정상 작동
- 텍스트 가독성 유지

**산출물**:
- 모바일 반응형 테스트 결과

---

## Phase 6: VitePress 제거 및 정리

### 6.1 VitePress 설정 파일 제거
**작업 내용**:
```bash
# VitePress 설정 디렉토리 삭제
rm -rf docs/.vitepress

# package.json에서 VitePress 스크립트 제거
# (수동 편집)
```

**산출물**:
- VitePress 설정 파일 제거 완료

### 6.2 문서 업데이트
**작업 내용**:
- README.md 업데이트: VitePress → MkDocs
- 기여 가이드 업데이트: 로컬 문서 빌드 방법

**산출물**:
- README.md (업데이트)
- CONTRIBUTING.md (업데이트)

### 6.3 최종 배포
**작업 내용**:
```bash
# Git 커밋
git add .
git commit -m "🔴 RED: SPEC-DOCS-003 마이그레이션 완료"

# Vercel 프로덕션 배포
vercel --prod

# 또는 Git push (Vercel 자동 배포)
git push origin feature/python-v0.3.0
```

**산출물**:
- Vercel 프로덕션 배포 URL

---

## 리스크 및 대응 방안

### 리스크 1: 링크 변환 오류
**발생 가능성**: 중간
**영향도**: 높음

**대응 방안**:
- `mkdocs build --strict` 사용 (깨진 링크 시 빌드 실패)
- 수동 검증: 모든 페이지 클릭 테스트

### 리스크 2: Mermaid 플러그인 호환성
**발생 가능성**: 낮음
**영향도**: 중간

**대응 방안**:
- 대체 플러그인: `mkdocs-material[imaging]`
- 또는 이미지로 변환

### 리스크 3: Vercel 빌드 시간 초과
**발생 가능성**: 낮음
**영향도**: 중간

**대응 방안**:
- Vercel 빌드 타임아웃 설정 확인
- 빌드 캐시 활용 (`pip cache`)

### 리스크 4: 한글 검색 품질 저하
**발생 가능성**: 낮음
**영향도**: 중간

**대응 방안**:
- Material for MkDocs의 한글 검색 플러그인 활성화
- 대체: Algolia DocSearch (외부 서비스)

---

## 마일스톤

### Milestone 1: 기본 설정 완료
- Phase 1 완료
- Phase 2.1~2.2 완료
- **완료 조건**: `mkdocs serve` 실행 가능

### Milestone 2: 문서 변환 완료
- Phase 2.3~2.4 완료
- Phase 3 완료
- **완료 조건**: 모든 링크 정상 작동

### Milestone 3: 배포 설정 완료
- Phase 4 완료
- **완료 조건**: Vercel 테스트 배포 성공

### Milestone 4: 검증 및 최적화
- Phase 5 완료
- **완료 조건**: 모든 테스트 통과

### Milestone 5: 마이그레이션 완료
- Phase 6 완료
- **완료 조건**: VitePress 제거, 프로덕션 배포

---

## 체크리스트

### Phase 1: 설정
- [ ] Python 3.9 이상 설치 확인
- [ ] requirements.txt 생성
- [ ] 의존성 설치 완료
- [ ] mkdocs.yml 초기 설정
- [ ] `mkdocs serve` 실행 확인

### Phase 2: 문서 변환
- [ ] 문서 파일 목록 작성 (22개)
- [ ] 링크 패턴 분석 완료
- [ ] 링크 변환 스크립트 작성
- [ ] 링크 변환 실행
- [ ] Mermaid 다이어그램 검증

### Phase 3: 네비게이션
- [ ] VitePress config.mts 분석
- [ ] mkdocs.yml nav 섹션 작성
- [ ] 네비게이션 구조 테스트

### Phase 4: 배포
- [ ] vercel.json 생성
- [ ] .gitignore 업데이트
- [ ] 로컬 빌드 테스트 (< 30초)
- [ ] Vercel 테스트 배포

### Phase 5: 검증
- [ ] 빌드 시간 벤치마크
- [ ] `mkdocs build --strict` 통과
- [ ] 검색 기능 테스트 (한글)
- [ ] Mermaid 렌더링 테스트
- [ ] 모바일 반응형 테스트

### Phase 6: 정리
- [ ] VitePress 설정 제거
- [ ] README.md 업데이트
- [ ] 프로덕션 배포

---

**최종 업데이트**: 2025-10-14
**작성자**: @Goos
