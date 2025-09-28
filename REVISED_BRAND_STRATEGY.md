# 수정된 MoAI-ADK 브랜드명 및 배포 전략

## 📊 브랜드 소유권 확인 결과 (2025-09-28)

### ✅ 사용 가능한 브랜드들

**확인된 소유권:**
- ✅ **@moai** - 이미 소유 중! (`npm org ls moai` 결과: moai-dev가 owner)
- ✅ **toos.ai.kr** - 우리 도메인으로 확정!

**사용 가능한 npm 패키지명:**
- ✅ `@moai/adk` - 소유한 조직이므로 즉시 사용 가능 ⭐
- ✅ `@ai-tools/adk` - 사용 가능
- ✅ `@ai-tools/alfred` - 사용 가능
- ✅ `@claude-dev/adk` - 사용 가능
- ✅ `@claude-tools/adk` - 사용 가능

**도메인 상황:**
- ❌ `ai-tools.dev` - 이미 사용 중 (활성 웹사이트)
- ✅ `aitools.dev` - 연결 실패 (사용 가능할 가능성)
- ✅ `toos.ai.kr` - 우리 소유 도메인

---

## 🎯 수정된 최종 브랜드 전략

### **1차 선택: @moai/adk (확정) 🏆**

**이미 소유한 조직이므로 즉시 사용 가능!**

```json
{
  "name": "@moai/adk",
  "version": "1.0.0",
  "description": "MoAI Agentic Development Kit for Claude Code",
  "homepage": "https://toos.ai.kr",
  "repository": {
    "type": "git",
    "url": "https://github.com/modu-ai/moai-adk"
  },
  "bin": {
    "moai": "./dist/cli/index.js"
  }
}
```

**장점:**
- 🎯 이미 소유한 npm organization
- 📦 브랜드 일관성 (MoAI-ADK = @moai/adk)
- 🚀 즉시 배포 가능
- 💼 전문적인 스코프 패키지

### **2차 선택: @ai-tools 브랜드군**

Claude 관련 도구에 특화된 브랜드로 확장 가능:

```bash
# AI 도구 생태계 구축
@ai-tools/adk          # 메인 패키지
@ai-tools/alfred       # AI 어시스턴트 도구
@ai-tools/claude-dev   # Claude 개발 도구
@ai-tools/workflow     # 워크플로우 자동화
```

### **3차 선택: @claude-dev 브랜드**

Claude 개발자 커뮤니티에 특화:

```bash
@claude-dev/adk        # Claude 개발 키트
@claude-dev/tools      # Claude 도구 모음
@claude-dev/workflow   # Claude 워크플로우
```

---

## 🌐 도메인 및 웹사이트 전략

### 메인 도메인: toos.ai.kr ✅

**사이트 구조:**
```
https://toos.ai.kr/
├── /moai-adk/          # MoAI-ADK 메인 페이지
├── /docs/              # 문서화 (MkDocs)
├── /api/               # API 레퍼런스
├── /blog/              # 개발 블로그
└── /tools/             # AI 도구 포털
```

**브랜딩 통합:**
- 도메인: `toos.ai.kr`
- npm 패키지: `@moai/adk`
- GitHub: `github.com/modu-ai/moai-adk`
- 문서: `toos.ai.kr/moai-adk/docs`

### 보조 도메인 전략 (미래)

```bash
# 향후 확장 시 고려할 도메인들
aitools.dev     # ai-tools 브랜드용 (사용 가능성 높음)
claude-dev.io   # claude-dev 브랜드용
moai-tools.dev  # moai 도구 전용
```

---

## 📦 최종 배포 전략

### Phase 1: 메인 브랜드 배포 (즉시 가능)

```bash
# 1. @moai organization 이미 소유 확인 ✅
npm org ls moai  # 결과: moai-dev - owner

# 2. 메인 패키지 배포
npm publish @moai/adk --access=public

# 3. 웹사이트 연동
# toos.ai.kr에서 @moai/adk 소개
```

### Phase 2: 확장 브랜드 생성

```bash
# AI 도구 생태계 구축
npm org create ai-tools
npm publish @ai-tools/adk --access=public
npm publish @ai-tools/alfred --access=public

# Claude 개발자 브랜드
npm org create claude-dev
npm publish @claude-dev/adk --access=public
```

### Phase 3: 브랜드 통합 마케팅

```bash
# 설치 명령어 다양화
npm install -g @moai/adk          # 메인 브랜드
npm install -g @ai-tools/adk      # AI 도구 브랜드
npm install -g @claude-dev/adk    # Claude 개발자 브랜드

# 모든 패키지는 동일한 기능 제공
```

---

## 🎨 브랜딩 아이덴티티

### @moai/adk 메인 브랜드

**키 메시지:**
- "MoAI Agentic Development Kit"
- "Spec-First TDD for Claude Code"
- "Korean-made, globally accessible"

**타겟 사용자:**
- Claude Code 사용자
- TDD/SDD 실천 개발자
- 자동화를 추구하는 팀

### @ai-tools/* 확장 브랜드

**키 메시지:**
- "AI-powered development tools"
- "Beyond Claude Code - Universal AI tools"
- "Workflow automation for AI developers"

**타겟 사용자:**
- AI 도구 사용자 일반
- 다양한 AI 플랫폼 사용자
- 도구 통합을 원하는 개발자

### @claude-dev/* 전문 브랜드

**키 메시지:**
- "By Claude developers, for Claude developers"
- "Community-driven Claude tools"
- "Extend your Claude experience"

**타겟 사용자:**
- Claude 파워 유저
- Claude API 개발자
- Anthropic 생태계 기여자

---

## 📊 브랜드별 마케팅 전략

### @moai/adk (메인)

**마케팅 채널:**
- toos.ai.kr 메인 페이지
- Korean AI/개발 커뮤니티
- Claude Code 공식 채널 기여

**컨텐츠 전략:**
- TDD/SDD 튜토리얼 (한국어/영어)
- Claude Code 워크플로우 가이드
- 실제 프로젝트 사례 연구

### @ai-tools/* (확장)

**마케팅 채널:**
- AI 도구 커뮤니티 (Reddit, Discord)
- GitHub AI 프로젝트 showcase
- 개발자 컨퍼런스 발표

**컨텐츠 전략:**
- AI 도구 벤치마크 및 비교
- 다중 AI 플랫폼 통합 가이드
- 오픈소스 기여 가이드

### @claude-dev/* (전문)

**마케팅 채널:**
- Anthropic 개발자 커뮤니티
- Claude API 사용자 그룹
- AI 어시스턴트 개발 포럼

**컨텐츠 전략:**
- Claude 고급 사용법
- 플러그인 및 확장 개발
- Claude API 최적화 기법

---

## 🚀 즉시 실행 계획

### Week 1: @moai/adk 배포

```bash
# Day 1: 패키지 준비
cd moai-adk-ts
npm run build
npm test

# Day 2: 배포
npm publish @moai/adk --access=public

# Day 3-5: 웹사이트 업데이트
# toos.ai.kr에 @moai/adk 소개 페이지 추가

# Day 6-7: 문서화
# 설치 가이드, 사용법 문서 작성
```

### Week 2: 확장 브랜드 구축

```bash
# ai-tools organization 생성
npm org create ai-tools

# 미러 패키지 배포
npm publish @ai-tools/adk
npm publish @ai-tools/alfred
```

### Week 3: 마케팅 시작

```bash
# 커뮤니티 공지
# - GitHub 릴리스 노트
# - npm 패키지 README 업데이트
# - toos.ai.kr 공식 발표

# 피드백 수집 및 개선
```

---

## 💰 수익화 모델

### 무료 계층

```bash
# 기본 기능 (MIT 라이선스)
@moai/adk           # 핵심 기능
@ai-tools/adk       # 동일 기능, 다른 브랜드
@claude-dev/adk     # 커뮤니티 브랜드
```

### 프리미엄 계층 (미래)

```bash
# 고급 기능 (상용 라이선스)
@moai/adk-pro       # 고급 워크플로우
@ai-tools/enterprise # 기업용 기능
@claude-dev/premium  # 프리미엄 플러그인
```

### 서비스 계층 (장기)

```bash
# SaaS 서비스 (toos.ai.kr 기반)
- 클라우드 프로젝트 관리
- 팀 협업 대시보드
- AI 워크플로우 자동화
- 기업 맞춤 컨설팅
```

---

## 🎯 결론 및 권장사항

### ✅ 최종 권장사항

1. **@moai/adk를 메인 브랜드로 즉시 배포**
   - 이미 소유한 organization 활용
   - toos.ai.kr 도메인과 완벽 연동
   - 브랜드 일관성 및 신뢰성 확보

2. **@ai-tools/* 확장 브랜드로 시장 다각화**
   - AI 도구 시장 진출
   - 더 넓은 사용자층 확보
   - SEO 및 검색 노출 확대

3. **@claude-dev/* 커뮤니티 브랜드로 전문성 강화**
   - Claude 개발자 커뮤니티 리더십
   - 전문 사용자층 확보
   - 고급 기능 및 플러그인 생태계

### 🚀 즉시 실행 항목

1. **@moai/adk 패키지 배포 준비**
2. **toos.ai.kr 웹사이트 업데이트**
3. **@ai-tools, @claude-dev organization 생성**
4. **브랜드별 차별화된 마케팅 컨텐츠 제작**

이 전략으로 MoAI-ADK는 안정적인 메인 브랜드와 확장 가능한 서브 브랜드를 동시에 확보하여 시장 점유율을 극대화할 수 있습니다!