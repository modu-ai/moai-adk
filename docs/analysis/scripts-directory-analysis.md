# templates/.moai/scripts 디렉토리 분석 보고서

**분석일**: 2025-10-01
**분석 대상**: `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/templates/.moai/scripts`

---

## 📊 요약

### 주요 발견사항
- ✅ 총 11개 파일 (10개 .ts + 1개 README.md)
- 🔴 **3개 파일 중복 확인** (실제 src 구현과 중복)
- ⚠️ **7개 파일 템플릿 예시** (사용되지 않음)
- 📦 모든 파일이 npm 패키지에 포함됨

### 권장 조치
**삭제 권장**: sync-analyzer.ts (완전 중복)
**보존 권장**: 템플릿 예시 파일들 (사용자 참고용)

---

## 📁 파일별 상세 분석

### 🔴 중복 파일 (삭제 권장)

#### 1. **sync-analyzer.ts** (326 라인)
- **상태**: 🔴 **완전 중복**
- **src 대응**: `src/scripts/sync-analyzer.ts` (329 라인)
- **차이점**:
  - templates 버전: `import { join } from 'path';`
  - src 버전: `import { logger } from '../utils/winston-logger.js';`
  - src 버전이 더 완전하고 실제 사용됨
- **사용 여부**: ❌ 어디에도 import 안됨
- **권장**: 🗑️ **삭제**

```bash
# 삭제 명령어
rm templates/.moai/scripts/sync-analyzer.ts
```

---

### ⚠️ 부분 중복 파일 (검토 필요)

#### 2. **spec-validator.ts** (387 라인)
- **상태**: ⚠️ 부분 중복
- **src 대응**: `src/core/spec-validator.ts` (533 라인)
- **차이점**:
  - src 버전이 더 완전한 구현 (146 라인 더 많음)
  - templates 버전은 단순화된 템플릿
- **사용 여부**: ❌ 어디에도 import 안됨
- **권장**: 🔵 **템플릿으로 보존** (사용자 참고용)

#### 3. **trust-checker.ts** (829 라인)
- **상태**: ⚠️ 부분 중복
- **src 대응**: `src/scripts/trust-principles-checker.ts` (493 라인)
- **차이점**:
  - templates 버전이 더 상세한 예시 (336 라인 더 많음)
  - 이름도 다름 (trust-checker vs trust-principles-checker)
- **사용 여부**: ❌ 어디에도 import 안됨
- **권장**: 🔵 **템플릿으로 보존** (사용자 참고용)

---

### ✅ 템플릿 예시 파일 (보존 권장)

#### 4. **debug-analyzer.ts** (25K, 약 750 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: 디버깅 분석 및 문제 진단 예시
- **src 대응**: 없음
- **권장**: 🔵 **보존** (사용자 프로젝트에서 참고)

#### 5. **detect-language.ts** (9.2K, 약 275 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: 프로젝트 언어 자동 감지 예시
- **src 대응**: 없음
- **권장**: 🔵 **보존** (사용자 프로젝트에서 활용)

#### 6. **doc-syncer.ts** (17K, 약 510 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: 문서 동기화 및 Living Document 관리 예시
- **src 대응**: 없음
- **권장**: 🔵 **보존** (에이전트 참고용)

#### 7. **project-init.ts** (3.9K, 약 117 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: 프로젝트 초기화 예시
- **src 대응**: 없음 (실제 초기화는 CLI로 처리)
- **권장**: 🔵 **보존** (사용자 프로젝트 커스터마이징용)

#### 8. **spec-builder.ts** (11K, 약 330 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: SPEC 문서 생성 예시
- **src 대응**: 없음
- **권장**: 🔵 **보존** (사용자 SPEC 생성 참고용)

#### 9. **tdd-runner.ts** (13K, 약 390 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: TDD Red-Green-Refactor 실행 예시
- **src 대응**: 없음
- **권장**: 🔵 **보존** (TDD 워크플로우 참고용)

#### 10. **test-analyzer.ts** (18K, 약 540 라인)
- **상태**: ✅ 순수 템플릿
- **용도**: 테스트 분석 및 품질 측정 예시
- **src 대응**: 없음
- **권장**: 🔵 **보존** (품질 측정 참고용)

#### 11. **README.md** (3.7K)
- **상태**: ✅ 문서
- **용도**: 스크립트 디렉토리 사용 가이드
- **권장**: ✅ **필수 보존**

---

## 📊 통계

### 파일 현황
| 상태 | 개수 | 비율 | 조치 |
|------|------|------|------|
| 🔴 완전 중복 | 1 | 10% | 삭제 |
| ⚠️ 부분 중복 | 2 | 20% | 검토 후 결정 |
| ✅ 템플릿 예시 | 7 | 70% | 보존 |
| **합계** | **10** | **100%** | - |

### 용량 분석
```
총 용량: 약 148K
- 삭제 권장: 7.9K (5.3%)
- 보존 권장: 140.1K (94.7%)
```

---

## 🎯 권장 조치

### 즉시 조치 (필수)

#### 1. sync-analyzer.ts 삭제
```bash
cd /Users/goos/MoAI/MoAI-ADK/moai-adk-ts
rm templates/.moai/scripts/sync-analyzer.ts
```

**이유**:
- src/scripts/sync-analyzer.ts와 완전 중복
- 실제 사용되는 것은 src 버전
- 혼란을 방지하고 유지보수 부담 감소

---

### 선택적 조치 (검토 필요)

#### 2. spec-validator.ts 검토
**옵션 A: 삭제**
- src/core/spec-validator.ts가 더 완전한 구현
- 중복 제거로 혼란 방지

**옵션 B: 보존**
- 사용자 프로젝트 커스터마이징 참고용
- README.md에 명확한 설명 추가

**권장**: 보존 (템플릿 용도 명시)

#### 3. trust-checker.ts 검토
**옵션 A: 통합**
- templates 버전의 상세한 예시를 src로 병합
- 하나의 완전한 구현 유지

**옵션 B: 보존**
- 더 상세한 예시로서 가치 있음
- 사용자 커스터마이징 참고용

**권장**: 보존 (상세 예시로 가치)

---

### 문서 개선 (권장)

#### README.md 업데이트
```markdown
# 중요: 템플릿 vs 실제 구현

이 디렉토리의 스크립트는 **사용자 프로젝트 참고용 템플릿**입니다.

## 실제 MoAI-ADK 구현
- spec-validator → `moai-adk/src/core/spec-validator.ts` 사용
- sync-analyzer → **삭제됨** (src/scripts/sync-analyzer.ts 사용)
- trust-checker → `moai-adk/src/scripts/trust-principles-checker.ts` 사용

## 템플릿 활용
다른 스크립트들은 순수 템플릿 예시입니다:
- debug-analyzer.ts
- detect-language.ts
- doc-syncer.ts
- project-init.ts
- spec-builder.ts
- tdd-runner.ts
- test-analyzer.ts

사용자 프로젝트에 복사하여 커스터마이징하세요.
```

---

## 🔍 심층 분석

### templates/.moai/scripts의 역할

#### 현재 상태
- 📦 npm 패키지에 포함됨 (`package.json: "files": ["templates"]`)
- ❌ MoAI-ADK 자체에서는 사용 안 함
- ✅ 사용자 프로젝트 초기화 시 복사됨

#### 의도된 용도
1. **사용자 가이드**: 프로젝트별 자동화 스크립트 예시
2. **참고 구현**: SPEC-First TDD 워크플로우 템플릿
3. **커스터마이징 기반**: 프로젝트 특성에 맞게 수정 가능

---

## 📝 결론

### 요약
templates/.moai/scripts는 **사용자 프로젝트를 위한 템플릿 디렉토리**로서 대부분 정상입니다.

### 필수 조치
- ✅ **sync-analyzer.ts 삭제** (완전 중복)

### 권장 조치
- ✅ **README.md 업데이트** (템플릿 용도 명시)
- ✅ **나머지 파일 보존** (템플릿 가치 유지)

### 예상 효과
- 🎯 중복 제거로 혼란 감소
- 📚 명확한 문서로 사용성 향상
- 💾 용량 절감: 약 8K (미미하지만 정리 효과)

---

**분석 담당**: cc-manager Agent
**검토 필요**: sync-analyzer.ts 삭제 승인