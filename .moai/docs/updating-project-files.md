---
title: 프로젝트 문서 업데이트 가이드
description: .moai/project 폴더의 product.md, structure.md, tech.md 파일 관리 및 업데이트 방법
version: 1.0.0
created_at: 2025-11-04
last_updated: 2025-11-04
language: Korean
---

# 프로젝트 문서 업데이트 가이드

> **대상**: MoAI-ADK 개발자 및 프로젝트 소유자
> **용도**: `.moai/project/` 폴더의 프로젝트 정의 문서 관리 및 업데이트
> **관련 명령**: `/alfred:0-project` (자동 생성), 수동 편집

---

## 📋 목차

1. [프로젝트 문서 구조](#프로젝트-문서-구조)
2. [각 문서의 역할](#각-문서의-역할)
3. [문서 업데이트 방법](#문서-업데이트-방법)
4. [자동 생성 vs 수동 편집](#자동-생성-vs-수동-편집)
5. [편집 시 주의사항](#편집-시-주의사항)
6. [동기화 및 버전 관리](#동기화-및-버전-관리)
7. [문제 해결](#문제-해결)

---

## 프로젝트 문서 구조

### 폴더 위치

```
프로젝트 루트/
├── .moai/
│   ├── project/                 ← 프로젝트 정의 문서
│   │   ├── product.md           (제품 비전 및 목표)
│   │   ├── structure.md         (기술 구조)
│   │   └── tech.md              (기술 스택)
│   ├── config.json
│   ├── specs/
│   └── reports/
├── CLAUDE.md                    (개발 가이드)
└── README.md                    (사용자 문서)
```

### 파일 계층 구조

```
README.md (사용자용)
    ↓
.moai/project/ (개발팀용)
    ├── product.md (비즈니스)
    ├── structure.md (아키텍처)
    └── tech.md (기술 상세)
    ↓
CLAUDE.md (개발자용)
```

---

## 각 문서의 역할

### 📄 product.md: 제품 정의 및 비전

**목적**: 프로젝트의 비즈니스 목표, 핵심 기능, 사용자 페르소나를 정의

**포함 내용**:
- 프로젝트 이름 및 간단한 설명
- 핵심 목표 (3-5개)
- 주요 기능 목록
- 타겟 사용자 그룹
- 성공 지표

**작성자**: 프로젝트 소유자, 프로덕트 매니저

**예시**:
```markdown
# MyProject: 분산 작업 추적 시스템

## 목표
- 팀 작업 투명성 증대
- 실시간 협업 지원
- 자동화된 리포팅

## 핵심 기능
1. 실시간 작업 상태 업데이트
2. 자동 진행률 계산
3. 팀 대시보드

## 타겟 사용자
- 리모트 팀 리더
- 프로젝트 매니저
```

### 🏗️ structure.md: 기술 구조

**목적**: 프로젝트의 아키텍처, 컴포넌트, 데이터 흐름을 설명

**포함 내용**:
- 아키텍처 다이어그램 (ASCII 또는 Mermaid)
- 주요 컴포넌트 및 역할
- 데이터 흐름
- 시스템 계층 구조
- 확장성 전략

**작성자**: 아키텍트, 시니어 개발자

**예시**:
```markdown
# 기술 구조

## 아키텍처 다이어그램

```
Frontend (React) → API Gateway → Backend (FastAPI) → Database (PostgreSQL)
                                        ↓
                                  Message Queue (Redis)
```

## 주요 컴포넌트
- **Frontend**: React 대시보드
- **Backend**: FastAPI REST API
- **Database**: PostgreSQL
- **Cache**: Redis

## 데이터 흐름
1. 클라이언트 → API 요청
2. API → 데이터베이스 조회
3. 캐시에서 자주 사용하는 데이터 반환
```

### 🔧 tech.md: 기술 스택 및 의존성

**목적**: 프로젝트에 사용된 모든 기술, 라이브러리, 도구를 문서화

**포함 내용**:
- 프로그래밍 언어 및 버전
- 프레임워크 및 라이브러리 (주요 의존성)
- 데이터베이스 및 저장소
- 배포 및 인프라
- 개발 도구체인

**작성자**: 개발팀, DevOps 팀

**예시**:
```markdown
# 기술 스택

## 프로그래밍 언어
- Python 3.11+ (Backend)
- TypeScript 5.0+ (Frontend)
- SQL (Database)

## 프레임워크 및 라이브러리
- Backend: FastAPI 0.100.0, SQLAlchemy 2.0
- Frontend: React 18.2, Next.js 14.0
- Testing: pytest 7.4, Vitest 1.0

## 데이터베이스 및 저장소
- PostgreSQL 15 (주 데이터베이스)
- Redis 7.0 (캐시)
- S3 호환 저장소 (파일 저장)

## 배포 및 인프라
- Docker & Docker Compose
- Kubernetes 1.28
- GitHub Actions (CI/CD)

## 개발 도구
- Git (버전 관리)
- VS Code / PyCharm (IDE)
- Make (작업 자동화)
- uv (Python 패키지 관리)
```

---

## 문서 업데이트 방법

### ✅ 방법 1: /alfred:0-project 명령 (권장)

**언제 사용**: 프로젝트의 전반적인 상태를 반영하고 싶을 때

**장점**:
- 🤖 자동 분석으로 일관성 있는 문서 생성
- ✅ 구조화된 템플릿 사용
- 🔄 기존 파일과의 병합 지원
- 📊 메타데이터 자동 갱신

**사용 방법**:
```bash
# Claude Code에서 실행
/alfred:0-project

# 또는 CLI에서 실행
moai-adk update
```

**결과**:
- product.md, structure.md, tech.md 자동 업데이트
- 기존 "Project Information" 섹션 유지
- 메타데이터 자동 갱신

### ✍️ 방법 2: 수동 편집 (빠른 업데이트용)

**언제 사용**: 작은 내용 수정, 빠른 업데이트가 필요할 때

**장점**:
- ⚡ 빠른 수정 가능
- 🎯 특정 섹션만 업데이트
- 💡 세부 조정 가능

**단계**:

1. 파일 열기
   ```bash
   vim .moai/project/product.md
   # 또는 VS Code 등 에디터 사용
   ```

2. 내용 편집
   - 목표, 기능, 기술 스택 등 업데이트
   - 마크다운 포맷 유지

3. 저장하기
   ```bash
   git add .moai/project/
   git commit -m "docs: Update product/structure/tech metadata"
   ```

4. 검증 (선택사항)
   ```bash
   # 파일이 유효한 마크다운인지 확인
   cat .moai/project/product.md | head -20
   ```

---

## 자동 생성 vs 수동 편집

### 비교표

| 측면 | /alfred:0-project | 수동 편집 |
|------|------------------|---------|
| **속도** | 중간 (분석 필요) | 빠름 |
| **정확도** | 높음 (자동 분석) | 의존함 (사람이 작성) |
| **일관성** | 매우 높음 | 낮을 수 있음 |
| **시간** | 5-10분 | 1-5분 |
| **기존 내용 유지** | ✅ 병합됨 | ✅ 수동 관리 |
| **메타데이터** | 자동 갱신 | 수동 갱신 |
| **용도** | 전반적 업데이트 | 빠른 수정 |

### 권장 사용 패턴

```
프로젝트 초기 설정
    ↓
/alfred:0-project 실행 (자동 생성)
    ↓
작은 변경사항 발생
    ↓
수동 편집 (빠른 수정)
    ↓
분기마다 또는 주요 변경 후
    ↓
/alfred:0-project 다시 실행 (전반적 동기화)
```

---

## 편집 시 주의사항

### ⚠️ 주의할 점

#### 1. 마크다운 포맷 유지

**잘못된 예**:
```
목표
제품을 만드는 것

기능
1번 기능
2번 기능
```

**올바른 예**:
```markdown
## 목표
- 제품을 만드는 것
- 사용자 만족도 높이기

## 기능
1. 첫 번째 기능
2. 두 번째 기능
```

#### 2. 메타데이터 헤더 보존

**반드시 유지**:
```markdown
---
title: product.md
description: 제품 정의
version: 1.0.0
created_at: 2025-11-04
---
```

#### 3. 최상위 제목은 하나만

**잘못된 예**:
```markdown
# 제품 정의
텍스트...

# 핵심 목표
텍스트...
```

**올바른 예**:
```markdown
# MyProject 제품 정의

## 핵심 목표
텍스트...

## 주요 기능
텍스트...
```

#### 4. 코드 블록은 백틱으로 감싸기

```markdown
# 기술 스택

## 사용 기술
- Python (백엔드)
- React (프론트엔드)

## 명령어 예시
`python -m pip install package`
```

#### 5. 민감한 정보 제외

**포함하지 마세요**:
- ❌ API 키, 비밀번호
- ❌ 내부 IP 주소
- ❌ 직원 개인정보
- ❌ 기밀 비즈니스 정보

---

## 동기화 및 버전 관리

### 파일 동기화 규칙

#### 규칙 1: 정기적인 동기화

```bash
# 매월 또는 분기마다
/alfred:0-project

# 또는 주요 변경 후 즉시
git add .moai/project/
git commit -m "docs: Sync project documentation"
```

#### 규칙 2: Git으로 버전 관리

```bash
# 커밋으로 변경 이력 남기기
git log --oneline .moai/project/

# 예시 출력
commit a1b2c3d docs: Update tech stack to v2.0
commit e4f5g6h docs: Add new microservices architecture
```

#### 규칙 3: 백업 유지

```bash
# 주요 변경 전 백업
cp .moai/project/product.md .moai/project/product.md.backup

# 편집
vim .moai/project/product.md

# 필요 시 복구
cp .moai/project/product.md.backup .moai/project/product.md
```

### 메타데이터 갱신

편집한 후 메타데이터 갱신:

```markdown
---
title: product.md
description: 제품 정의
version: 1.0.1            ← 버전 증가
created_at: 2025-11-04
last_updated: 2025-11-05  ← 수정 날짜 갱신
---
```

---

## 문제 해결

### Q1: 파일을 편집했는데 `/alfred:0-project`를 실행하면 내용이 지워집니다

**A**: `/alfred:0-project`는 템플릿 기반으로 문서를 재생성합니다.

**해결책**:
1. **중요한 내용은 따로 백업**
   ```bash
   cp .moai/project/product.md /tmp/product.backup.md
   ```

2. **/alfred:0-project 실행** (자동으로 병합 시도)

3. **수동으로 내용 복구** (필요시)
   ```bash
   # 백업에서 필요한 내용만 가져오기
   diff /tmp/product.backup.md .moai/project/product.md
   ```

**권장**: 수동 편집 후에는 `/alfred:0-project`를 다시 실행하지 마세요. 다음 분기 또는 주요 구조 변경 시에만 실행하세요.

### Q2: 마크다운이 제대로 렌더링되지 않습니다

**A**: 마크다운 포맷 검증

**확인해야 할 항목**:
- [ ] 제목에 `#` 사용 (예: `# 제목`, `## 부제목`)
- [ ] 목록은 `-` 또는 `*` 사용
- [ ] 코드는 백틱으로 감싸기 (`` ` `)
- [ ] 링크 포맷: `[텍스트](URL)`

**검증 명령**:
```bash
# VS Code에서 열기
code .moai/project/product.md

# 또는 온라인 마크다운 뷰어 사용
# https://www.markdownlivepreview.com/
```

### Q3: 어떤 파일을 먼저 편집해야 합니까?

**A**: 이 순서를 권장합니다:

1. **product.md** (먼저)
   - 프로젝트의 "무엇"과 "왜"를 정의
   - 다른 문서의 기반이 됨

2. **structure.md** (두 번째)
   - product.md의 목표를 달성하기 위한 "어떻게"
   - 아키텍처 설계

3. **tech.md** (마지막)
   - structure.md를 구현하기 위한 구체적 기술
   - 라이브러리, 프레임워크 선택

### Q4: 여러 사람이 함께 편집할 수 있습니까?

**A**: 가능하지만 충돌 방지가 필요합니다.

**권장 방법**:
```bash
# 1. 최신 버전으로 업데이트
git pull origin develop

# 2. 수정 사항 적용
vim .moai/project/product.md

# 3. 다른 변경사항 확인
git status

# 4. 커밋하기
git add .moai/project/
git commit -m "docs: Update product definition - add new features"

# 5. 푸시하기
git push origin develop
```

**충돌 발생 시**:
```bash
# 모두의 변경사항 병합
git merge --no-ff origin/develop

# 충돌 해결
vim .moai/project/product.md  # 파일 직접 수정

# 병합 완료
git add .moai/project/
git commit -m "docs: Merge product documentation"
```

---

## 정리

| 상황 | 추천 방법 | 시간 |
|------|---------|------|
| 프로젝트 처음 생성 | `/alfred:0-project` | 5-10분 |
| 빠른 수정 | 수동 편집 | 1-5분 |
| 정기적 동기화 | `/alfred:0-project` | 5-10분 |
| 구조 재설계 | 수동 편집 후 `/alfred:0-project` | 15-30분 |
| 팀 협업 | Git + 수동 편집 | 10-20분 |

---

**문서 버전**: 1.0.0
**마지막 업데이트**: 2025-11-04
**작성자**: Alfred (MoAI-ADK SuperAgent)

🤖 Generated with Claude Code
Co-Authored-By: 🎩 Alfred@MoAI
