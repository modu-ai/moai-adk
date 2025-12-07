# 장르 템플릿: architecture-overview (아키텍처 개요형)

## 📋 용도

시스템, 프레임워크 전체 구조 설명

**구조**: High-Level Diagram (15%) → Components (40%) → Data Flow (20%) → Integration Points (15%) → Trade-offs (10%)

**특징**:
- 전체 그림 우선
- 컴포넌트별 상세
- 데이터 흐름 명확
- 설계 결정 근거

---

## 🏗️ 5가지 구성 요소

### 1. 문서 구조

```
├─ High-Level Diagram (15%, 전체 구조)
│  ├─ 시스템 다이어그램
│  └─ 주요 컴포넌트
│
├─ Components (40%, 컴포넌트 설명)
│  ├─ Component A (10%)
│  │  ├─ 역할
│  │  └─ 인터페이스
│  ├─ Component B (10%)
│  └─ Component C... (가변)
│
├─ Data Flow (20%, 데이터 흐름)
│  ├─ 흐름 설명
│  └─ 시퀀스 다이어그램
│
├─ Integration Points (15%, 통합 지점)
│  ├─ 외부 시스템
│  └─ API 경계
│
└─ Trade-offs (10%, 설계 결정)
   ├─ 선택한 이유
   └─ 대안 비교

총 섹션: 5개
```

### 2. 문체

**어조**: 아키텍처 설명

**종결어미**: "-다", "-니다"

**능동태**: 85% 이상

**문장 길이**: 평균 20-25단어

### 3. 내용 전개

1. **High-Level**: 전체 그림
2. **Components**: 각 부분 역할
3. **Data Flow**: 어떻게 연결
4. **Integration**: 외부 연동
5. **Trade-offs**: 왜 이렇게

### 4. 조건

**글자 수**: 2,000-2,500자

**다이어그램**: 2개 이상 필수

**필수 요소**:
- 전체 아키텍처 다이어그램
- 각 컴포넌트 역할
- 데이터 흐름도

### 5. 형식

```markdown
# MoAI-ADK 아키텍처 개요

## High-Level Diagram

\`\`\`mermaid
graph TB
    User[사용자] --> CLI[CLI Interface]
    CLI --> Alfred[Mr.Alfred Agent]
    Alfred --> Agents[35 Specialized Agents]
    Agents --> Skills[135 Skills]
    Skills --> Tools[Tools & MCP]
\`\`\`

MoAI-ADK는 4계층 아키텍처를 따릅니다.

## Components

### CLI Interface (사용자 인터페이스)

**역할**: 사용자 명령어 수신 및 결과 표시

**인터페이스**:
- `/moai:0-project` - 프로젝트 초기화
- `/moai:1-plan` - SPEC 생성
- `/moai:2-run` - TDD 구현

### Mr.Alfred (슈퍼 에이전트)

**역할**: 작업 분석 및 전문 에이전트 위임

**책임**:
- 사용자 의도 파악
- 적절한 에이전트 선택
- 실행 결과 통합

## Data Flow

\`\`\`mermaid
sequenceDiagram
    User->>CLI: /moai:1-plan
    CLI->>Alfred: 요청 전달
    Alfred->>spec-builder: SPEC 생성 위임
    spec-builder->>Alfred: SPEC 반환
    Alfred->>CLI: 결과 전달
    CLI->>User: SPEC 출력
\`\`\`

## Integration Points

- **Context7 MCP**: 공식 문서 조회
- **GitHub API**: 이슈/PR 관리
- **Notion API**: 문서화 자동화

## Trade-offs

**선택**: 4계층 아키텍처

**이유**:
- 명확한 책임 분리
- 확장 용이성
- 테스트 가능성

**대안**: 모놀리식 에이전트 (기각 - 유지보수 어려움)
```

---

**마지막 수정**: 2025-11-25
**버전**: 1.0.0
