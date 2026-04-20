---
title: 시작하기
weight: 20
draft: false
---
# 시작하기

AI Agency의 첫 프로젝트를 시작하는 완벽한 가이드입니다.

## 전제조건

- Node.js 20+ 설치
- 브랜드 가이드라인 또는 디자인 시스템 (기존 프로젝트인 경우)
- 프로젝트 브리프 또는 요구사항 문서
- (선택사항) 대체 텍스트/디자인 샘플

## Step 1: 브랜드 컨텍스트 설정

먼저 AI Agency가 참고할 **브랜드 가이드라인**을 정의합니다.

### 디렉토리 구조

```
my-website/
├── .agency/
│   ├── context/
│   │   ├── brand.md           # 브랜드 아이덴티티
│   │   ├── design-system.md   # 디자인 시스템
│   │   ├── copy-guidelines.md # 카피 스타일 가이드
│   │   └── audience.md        # 대상 고객 정보
│   ├── briefs/
│   │   └── initial-brief.md   # 프로젝트 브리프
│   └── learnings/
│       └── rules.yaml         # 학습된 규칙들
├── src/
│   ├── pages/
│   ├── components/
│   └── styles/
└── output/
    └── website/               # 생성된 웹사이트
```

### 브랜드 가이드라인 작성 예시

**`.agency/context/brand.md`**

```markdown
# 우리 브랜드

## 브랜드 정체성
- 회사명: Acme Corp
- 슬로건: "혁신하는 기술, 신뢰하는 선택"
- 핵심 가치: 혁신, 신뢰, 투명성

## 시각 아이덴티티
- 주 색상: #007AFF (파란색)
- 보조 색상: #34C759 (초록색), #FF3B30 (빨간색)
- 폰트: Inter (헤딩), SF Pro (본문)
- 톤: 모던하고 전문적이며 접근하기 쉬운 느낌

## 타겟 고객
- 기술 스타트업 CEO 및 CTO
- 나이: 25-45세
- 관심사: 비용 절감, 효율성, 보안
```

### 설계 시스템 정의

**`.agency/context/design-system.md`**

```markdown
# 디자인 시스템

## 레이아웃
- 최대 너비: 1200px
- 그리드: 12 컬럼
- 여백: 16px 기본 단위
- 모바일: 375px 최소 너비

## 컴포넌트
- 버튼: 3가지 크기 (sm, md, lg)
- 입력 필드: 텍스트, 선택, 다중선택
- 카드: 섬네일 + 텍스트 레이아웃
- 네비게이션: 헤더 고정, 반응형 메뉴

## 모션
- 전환(Transition): 200ms ease-in-out
- 호버(Hover): 스케일 1.05
- 로딩: 부드러운 페이드 효과
```

## Step 2: 프로젝트 브리프 작성

**`.agency/briefs/initial-brief.md`**

```markdown
# 프로젝트 브리프

## 목표
Acme Corp의 새로운 제품 랜딩 페이지 제작

## 페이지 구성
1. 히어로 섹션: 제품 소개
2. 기능 섹션: 3가지 주요 기능 설명
3. 사용 사례: 고객 사례 3개
4. 가격 플랜: 3가지 가격대
5. CTA 섹션: 가입 유도
6. 푸터: 링크, 소셜, 연락처

## 특별 요구사항
- 성능: LCP < 2.5초
- 접근성: WCAG AA 준수
- SEO: 주요 키워드 포함
- 모바일: 완전 반응형

## 데드라인
2주 이내
```

## Step 3: 프로젝트 초기화

```bash
# 프로젝트 디렉토리 생성
mkdir my-website
cd my-website

# Agency 초기화
agency init

# 위에서 작성한 브리프 파일들을 .agency/ 디렉토리에 추가
# 브리프 로드
agency load-brief .agency/briefs/initial-brief.md
```

## Step 4: 첫 빌드 실행

### "Just do it" 모드 (추천)

자동으로 최적의 과정을 진행합니다:

```bash
agency build
```

이 명령어는 다음을 자동으로 수행합니다:
1. **Planner**: 브리프와 가이드라인 분석
2. **Copywriter**: 각 섹션의 텍스트 작성
3. **Designer**: 디자인 목업 생성
4. **Builder**: 웹사이트 구현
5. **Evaluator**: 품질 평가
6. **Learner**: 규칙 저장

진행 상황은 실시간으로 표시됩니다.

### Step-by-step 모드 (세밀한 제어)

각 단계를 개별적으로 제어할 수 있습니다:

```bash
# 1단계: 전략 수립
agency plan

# 2단계: 카피 작성
agency write-copy

# 3단계: 디자인 생성
agency design

# 4단계: 웹사이트 빌드
agency build

# 5단계: 평가 및 개선
agency evaluate

# 6단계: 규칙 저장
agency learn
```

## Step 5: 결과 검토 및 피드백

빌드가 완료되면 생성된 웹사이트를 확인합니다:

```bash
# 로컬 서버에서 미리보기
agency serve

# 브라우저에서 http://localhost:3000 방문
```

### 평가 항목 체크리스트

- [ ] 시각 아이덴티티: 브랜드 가이드라인 준수
- [ ] 텍스트: 문법 정확, 톤 일관성
- [ ] 레이아웃: 정렬, 여백, 읽기 흐름
- [ ] 기능성: 버튼, 링크, 폼 작동
- [ ] 반응형: 모바일, 태블릿, 데스크탑
- [ ] 성능: 로딩 속도, 이미지 최적화
- [ ] SEO: 메타데이터, 제목, 설명

### 피드백 제출

마음에 들지 않는 부분에 대해 피드백을 제공합니다:

```bash
# 대화형 피드백 입력
agency feedback

# 또는 파일로 피드백 저장
agency feedback --file .agency/feedback.md
```

**피드백 작성 가이드:**
- 구체적이어야 함: "더 나은 색상" ❌ → "메인 색상을 더 밝게" ✓
- 이유를 포함: "왜 그렇게 생각하나요?"
- 우선순위 표시: 필수 / 개선 / 선택

## Step 6: 진화 (Evolution)

AI Agency는 피드백을 학습합니다:

```bash
# 피드백 학습 및 재빌드
agency evolve

# 또는 전체 재평가 사이클
agency build --with-feedback
```

이 과정에서 Learner 에이전트는:
1. Evaluator의 평가 분석
2. 개선 패턴 인식
3. 새로운 휴리스틱 생성
4. 다음 프로젝트에 적용할 규칙으로 저장

## 기본 커맨드 요약

| 커맨드 | 설명 |
|--------|------|
| `agency init` | 프로젝트 초기화 |
| `agency build` | 웹사이트 생성 (전체 파이프라인) |
| `agency serve` | 로컬 미리보기 |
| `agency feedback` | 피드백 입력 |
| `agency evolve` | 피드백 기반 진화 |
| `agency status` | 프로젝트 상태 확인 |
| `agency deploy` | 배포 |

## 다음 단계

- [에이전트 & 스킬](/ko/agency/agents-and-skills) - 각 에이전트의 역할 이해
- [자기진화 시스템](/ko/agency/self-evolution) - 진화 메커니즘 상세
- [커맨드 레퍼런스](/ko/agency/command-reference) - 모든 커맨드 설명서
