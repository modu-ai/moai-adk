# 빠른 시작 가이드

**@DOC:QUICK-START-001** | **최종 업데이트**: 2025-11-05 | **소요 시간**: 5분

---

## <span class="material-icons">target</span> 목표

이 가이드를 통해 MoAI-ADK 온라인 문서 시스템을 완벽하게 설정하고 실행하는 방법을 배웁니다.

---

## <span class="material-icons">rocket_launch</span> 1단계: 시스템 요구사항

### 필수 요구사항
- **Python**: 3.13 이상
- **UV**: 최신 버전 권장
- **Git**: 최신 버전
- **브라우저**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+

### 선택적 요구사항
- **Vercel CLI**: 자동 배포를 위한 선택적 도구
- **Node.js**: v18+ (일부 빌드 도구에 필요)

---

## ⚡ 2단계: 프로젝트 설정 (30초)

### UV 설치
```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# 설치 확인
uv --version
```

### 프로젝트 복제 및 설정
```bash
# 1. 프로젝트 복제
git clone https://github.com/moai-adk/MoAI-ADK.git
cd MoAI-ADK

# 2. 의존성 설치 (자동)
uv sync

# 3. 개발 서버 실행
uv run dev
```

### 빠른 확인
```bash
# 서버 상태 확인
curl http://127.0.0.1:8080

# 빌드 상태 확인
uv run build
```

---

## <span class="material-icons">palette</span> 3단계: 문서 시스템 구축

### MkDocs 설정
```bash
# MkDocs 기본 설정 확인
uv run mkdocs --help

# 프로젝트 구조 생성
mkdir -p docs/{getting-started,alfred,commands,development,advanced,api,contributing}

# 테마 설정
uv run mkdocs new .
```

### 다국어 설정 추가
```yaml
# mkdocs.yml
site_name: MoAI-ADK Documentation
nav:
  - 홈: index.md
  - 빠른 시작: getting-started/
  - Alfred: alfred/
  - 명령어: commands/
  - 개발: development/
  - 고급 기능: advanced/
  - API: api/
  - 기여: contributing/

theme:
  name: material
  language: ko
  palette:
    - media: "(prefers-color-scheme: light)"
      scheme: default
      primary: blue
      accent: indigo
    - media: "(prefers-color-scheme: dark)"
      scheme: slate
      primary: blue
      accent: indigo
```

---

## <span class="material-icons">search</span> 4단계: 검색 및 내비게이션

### 검색 시스템 활성화
```bash
# 의존성 추가
uv add mkdocs-search

# 설정 업데이트
uv run mkdocs build --strict
```

### 실시간 검색 테스트
1. 개발 서버 실행: `uv run dev`
2. 브라우저에서 http://127.0.0.1:8080 접속
3. 검색창에 "MoAI" 입력
4. 실시간 검색 결과 확인

---

## 🌍 5단계: 다국어 설정

### 언어 파일 생성
```bash
# 영어
echo "# English Documentation" > docs/getting-started/index-en.md

# 일본어
echo "# 日本語ドキュメント" > docs/getting-started/index-ja.md

# 중국어
echo "# 中文文档" > docs/getting-started/index-zh.md
```

### 언어 전환 테스트
```bash
# 다국어 문서 빌드
uv run build

# 결과 확인
ls -la site/getting-started/
```

---

## <span class="material-icons">analytics</span> 6단계: 배포 준비

### 로컬 빌드
```bash
# 정적 사이트 빌드
uv run build

# 결과 확인
ls -la site/

# 파일 사이즈 확인
du -sh site/
```

### Vercel 배포
```bash
# 1. Vercel CLI 설치
npm i -g vercel

# 2. 로그인
vercel login

# 3. 프로젝트 배포
vercel --prod

# 4. 배포 확인
vercel ls
```

---

## 🧪 7단계: 테스트 및 검증

### 자동화 테스트
```bash
# 1. 문서 유효성 검사
uv run validate

# 2. 링크 검사
uv run check-links

# 3. 빌드 테스트
uv run build --strict
```

### 수동 테스트 체크리스트
- [ ] 모킹 페이지 정상적으로 표시
- [ ] 다크/라이트 모드 전환 동작
- [ ] 검색 기능 정상 작동
- [ ] 모바일 반응형 디자인 확인
- [ ] 다국어 문서 접근성 확인

---

## 📋 완성 확인 체크리스트

### 시스템 상태
- [ ] UV 설치 완료
- [ ] 프로젝트 복제 성공
- [ ] 의존성 설치 성공
- [ ] 개발 서버 실행 성공
- [ ] 문서 빌드 성공

### 기능 확인
- [ ] 모킹 페이지 정상 표시
- [ ] 검색 기능 정상 작동
- [ ] 다국어 지원 확인
- [ ] 반응형 디자인 확인
- [ ] 다크/라이트 모드 전환 확인

### 배포 확인
- [ ] 로컬 빌드 성공
- [ ] Vercel 배포 성공
- [ ] 도메인 접속 확인
- [ ] SSL 인증서 확인
- [ ] CDN 성능 확인

---

## <span class="material-icons">rocket_launch</span> 다음 단계

### 1. 커스터마이징
- 디자인 시스템 수정
- 새로운 언어 추가
- 커스텀 컴포넌트 개발

### 2. 콘텐츠 추가
- API 문서 생성
- 튜토리얼 작성
- 고급 가이드 추가

### 3. 프로덕션 배포
- 자동 배포 설정
- 모니터링 도구 연결
- 성능 최적화

---

## 🐛 문제 해결

### 일반적인 문제

#### UV 설치 오류
```bash
# 캐시 정리
uv cache clean

# 재설치
pip install --upgrade uv
```

#### 빌드 오류
```bash
# 캐시 정리
rm -rf site/ .doit_db/

# 의존성 재설치
uv sync --force

# 재빌드
uv run build
```

#### 서버 시작 오류
```bash
# 포트 변경
uv run dev --port 3000

# 로그 확인
uv run dev --verbose
```

---

## 📞 지원

### 공식 문서
- **주소**: https://adk.mo.ai.kr
- **상태**: 24/7 운영
- **업데이트**: 실시간 동기화

### 개발 지원
- **GitHub Issues**: [기술 문제 제기](https://github.com/moai-adk/MoAI-ADK/issues)
- **GitHub Discussions**: [질의응답](https://github.com/moai-adk/MoAI-ADK/discussions)
- **이메일**: support@mo.ai.kr

---

*최종 업데이트: 2025-11-05 | 버전: v0.17.0 | 상태: 100% 완료*