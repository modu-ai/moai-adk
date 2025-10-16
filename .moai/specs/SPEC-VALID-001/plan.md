# SPEC-VALID-001 구현 계획

> SPEC 메타데이터 검증 자동화 구현 전략

---

## 구현 우선순위 (4단계)

### Phase 1: Core Validator 구현 (우선순위: Critical)

**목표**: 핵심 검증 로직 구현 및 단위 테스트

**구현 범위**:
1. **spec-validator.ts** 생성
   - `validateRequiredFields()`: 7개 필수 필드 존재 확인
   - `validateFieldFormats()`: 필드 형식 검증 (version, author, dates, enums)
   - `validateHistory()`: HISTORY 섹션 검증
   - `validateDirectoryName()`: 디렉토리 명명 규칙 검증

2. **metadata-parser.ts** 생성
   - `parseYamlFrontMatter()`: YAML Front Matter 파싱
   - `extractMetadata()`: 메타데이터 추출 및 구조화
   - `extractHistory()`: HISTORY 섹션 파싱

**기술 스택**:
- TypeScript (strict mode)
- YAML parser (safe parsing only)
- 정규 표현식 (필드 형식 검증)

**의존성**:
- 없음 (독립적 구현 가능)

**완료 조건**:
- [ ] 12개 검증 함수 구현 완료
- [ ] 단위 테스트 작성 (정상 6개, 오류 6개)
- [ ] 테스트 커버리지 ≥ 90%

---

### Phase 2: /alfred:1-spec 통합 (우선순위: High)

**목표**: spec-builder 에이전트와 검증 시스템 통합

**구현 범위**:
1. **spec-builder.ts 수정**
   - SPEC 문서 생성 직후 `validateSpec()` 호출
   - 검증 실패 시 작업 중단 및 오류 메시지 출력
   - 검증 통과 시 다음 단계 진행 (Git 커밋)

2. **에러 메시지 포맷팅**
   - 구조화된 오류 메시지 템플릿 작성
   - 심각도별 아이콘 표시 (❌ Critical, ⚠️ Warning)

**통합 시나리오**:
```
/alfred:1-spec "기능명"
  ↓
SPEC 문서 3개 생성 (spec.md, plan.md, acceptance.md)
  ↓
validateSpec() 호출 ← **새로운 단계**
  ↓
검증 통과 → Git 커밋 → Draft PR 생성
검증 실패 → 오류 메시지 출력 → 작업 중단
```

**의존성**:
- Phase 1 완료 (Core Validator)

**완료 조건**:
- [ ] spec-builder와 검증 시스템 통합 완료
- [ ] 검증 실패 시 작업 중단 확인
- [ ] 오류 메시지 가독성 확인

---

### Phase 3: 중복 ID 및 순환 의존성 검증 (우선순위: High)

**목표**: Grep 도구를 활용한 TAG 무결성 검증

**구현 범위**:
1. **duplicate-checker.ts** 생성
   - `checkDuplicateId()`: Grep으로 중복 SPEC ID 검색
   - `scanExistingSpecs()`: .moai/specs/ 디렉토리 전체 스캔

2. **dependency-checker.ts** 생성
   - `checkCircularDependency()`: depends_on 필드 순환 의존성 감지
   - `buildDependencyGraph()`: SPEC 간 의존성 그래프 생성

**검증 명령 예시**:
```bash
# 중복 ID 검색
rg "@SPEC:VALID-001" -n .moai/specs/

# 의존성 체인 검색
rg "depends_on:" -A 5 .moai/specs/SPEC-*/spec.md
```

**의존성**:
- Phase 1 완료 (Core Validator)

**완료 조건**:
- [ ] 중복 ID 검증 로직 구현
- [ ] 순환 의존성 감지 알고리즘 구현
- [ ] 통합 테스트 작성

---

### Phase 4: 문서화 및 사용자 가이드 (우선순위: Medium)

**목표**: 검증 시스템 사용 방법 문서화

**구현 범위**:
1. **검증 오류 메시지 카탈로그 작성**
   - 각 오류 유형별 예시 및 해결 방법
   - 자주 묻는 질문 (FAQ)

2. **development-guide.md 업데이트**
   - SPEC 메타데이터 검증 섹션 추가
   - 검증 실패 시 대응 방법 문서화

3. **@DOC:VALID-001 작성**
   - 검증 시스템 아키텍처 다이어그램
   - API 레퍼런스 (검증 함수 목록)

**의존성**:
- Phase 1, 2, 3 모두 완료

**완료 조건**:
- [ ] 오류 메시지 카탈로그 작성 완료
- [ ] development-guide.md 업데이트 완료
- [ ] @DOC:VALID-001 작성 완료

---

## 아키텍처 설계

### 모듈 구조

```
moai-adk-ts/src/core/spec/
├── validator/
│   ├── spec-validator.ts       # 메인 검증 로직
│   ├── metadata-parser.ts      # YAML 파싱
│   ├── duplicate-checker.ts    # 중복 ID 검증
│   └── dependency-checker.ts   # 순환 의존성 검증
├── index.ts                    # Public API
└── types.ts                    # 타입 정의

tests/core/spec/
├── spec-validator.test.ts      # 단위 테스트
├── metadata-parser.test.ts     # 단위 테스트
└── integration.test.ts         # 통합 테스트
```

### 데이터 흐름

```
SPEC 문서 (spec.md)
  ↓
metadata-parser.ts (YAML 파싱)
  ↓
spec-validator.ts (필드 검증)
  ↓
duplicate-checker.ts (중복 ID 검증)
  ↓
dependency-checker.ts (순환 의존성 검증)
  ↓
검증 결과 (성공/실패)
```

---

## 리스크 및 대응 방안

### 리스크 1: YAML 파싱 오류

**발생 가능성**: Medium
**영향도**: High

**대응 방안**:
- Safe YAML 파서 사용 (임의 코드 실행 방지)
- 파싱 실패 시 구체적 오류 메시지 제공
- Fallback: 정규 표현식으로 필수 필드만 추출

### 리스크 2: 성능 저하 (대규모 프로젝트)

**발생 가능성**: Low
**영향도**: Medium

**대응 방안**:
- Grep 도구 활용 (빠른 파일 검색)
- 캐싱 전략 (최근 검증 결과 캐시)
- 병렬 처리 (여러 SPEC 동시 검증)

### 리스크 3: 기존 SPEC 문서 비호환성

**발생 가능성**: High
**영향도**: Medium

**대응 방안**:
- Migration 스크립트 제공
- 경고 모드 지원 (검증 실패 시 경고만 출력)
- 점진적 마이그레이션 가이드 제공

---

## 기술적 접근 방법

### 1. YAML Front Matter 파싱

**도구**: `js-yaml` (safe mode)

**파싱 전략**:
```typescript
import yaml from 'js-yaml';

function parseYamlFrontMatter(content: string): Metadata {
  const match = content.match(/^---\n([\s\S]+?)\n---/);
  if (!match) {
    throw new Error('YAML Front Matter not found');
  }

  const metadata = yaml.load(match[1], { schema: yaml.SAFE_SCHEMA });
  return metadata as Metadata;
}
```

### 2. 필드 형식 검증 (정규 표현식)

**검증 패턴**:
```typescript
const patterns = {
  version: /^\d+\.\d+\.\d+$/,                    // Semantic Versioning
  author: /^@[A-Z][a-zA-Z0-9-]+$/,               // @GitHub ID
  date: /^\d{4}-\d{2}-\d{2}$/,                   // YYYY-MM-DD
  status: /^(draft|active|completed|deprecated)$/,
  priority: /^(critical|high|medium|low)$/,
};
```

### 3. 중복 ID 검증 (Grep)

**검증 명령**:
```bash
rg "@SPEC:{ID}" -n .moai/specs/ --files-with-matches
```

**결과 해석**:
- 결과가 비어있음 → 중복 없음 → 생성 가능
- 결과가 있음 → 중복 발견 → 오류 반환

---

## 테스트 전략

### 단위 테스트 (12개)

**정상 케이스 (6개)**:
1. 모든 필수 필드 존재
2. version 형식 올바름 (0.0.1)
3. author 형식 올바름 (@Goos)
4. 날짜 형식 올바름 (2025-10-16)
5. status enum 올바름 (draft)
6. HISTORY 섹션 존재

**오류 케이스 (6개)**:
1. 필수 필드 누락 (id 누락)
2. version 형식 오류 (0.0.1a)
3. author 형식 오류 (Goos)
4. 날짜 형식 오류 (2025/10/16)
5. status 잘못된 값 (pending)
6. HISTORY 섹션 누락

### 통합 테스트 (3개)

1. `/alfred:1-spec` 실행 → 검증 통과 → Git 커밋 성공
2. `/alfred:1-spec` 실행 → 검증 실패 → 작업 중단
3. 중복 ID 생성 시도 → 검증 실패 → 오류 메시지

---

## 다음 단계 안내

Phase 1 완료 후:
```bash
/alfred:2-build VALID-001
```

Phase 4 완료 후:
```bash
/alfred:3-sync
```

---

**문서 상태**: Draft
**마지막 업데이트**: 2025-10-16
