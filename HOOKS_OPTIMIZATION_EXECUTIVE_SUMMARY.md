# Claude Code Hooks 최적화: 요약 보고서

**프로젝트**: MoAI-ADK v0.26.0
**분석 날짜**: 2025-11-19
**상태**: ✅ 분석 완료 - Phase 2 실행 준비 완료

---

## 핵심 발견사항

### 문제점 요약

MoAI-ADK의 Claude Code Hooks 시스템에서 **구조적 중복과 유지보수성 문제** 발견:

```
🔴 Critical Issues (3개)
  1. 중복 코드: moai/core/ + moai/shared/core/ 동시 존재
  2. 임포트 경로 혼동: 상대 경로 + 절대 경로 혼용
  3. Hook이 아닌 코드: spec_status_hooks.py (CLI 유틸리티)

🟠 Major Issues (4개)
  4. 빈 디렉토리: moai/handlers/, moai/shared/config/
  5. 과도한 Hook 크기: session_start__auto_cleanup.py (628 lines)
  6. SessionEnd Hook 위치 불확실
  7. 공식 스펙 준수 확인 미완료

🟡 Minor Issues (3개)
  8. __pycache__ Git 추적
  9. 불완전한 문서화
  10. 캐싱 전략 분산
```

### 영향도 분석

```
현재 상태의 문제점:
┌─────────────────────────────────────────┐
│  유지보수 비용                          │
│  ┌──────────────────────────┐           │
│  │ 변경 시 2곳 동시 수정    │ ← 중복   │
│  │ 임포트 경로 이해 어려움   │ ← 혼동  │
│  │ 테스트 범위 불명확      │ ← 복잡  │
│  └──────────────────────────┘           │
└─────────────────────────────────────────┘

최적화 후:
┌─────────────────────────────────────────┐
│  유지보수 비용                          │
│  ┌──────────────────────────┐           │
│  │ 변경 시 1곳만 수정      │ 50% 감소 │
│  │ 임포트 경로 명확         │ 100% 개선│
│  │ 테스트 범위 명확         │ 완전 자동│
│  └──────────────────────────┘           │
└─────────────────────────────────────────┘
```

---

## 제안 솔루션

### 6단계 최적화 계획

| Phase | 목표 | 기간 | 노력 | 영향 | 위험 |
|-------|------|------|------|------|------|
| **1** | 분석 기획 | ✅ 완료 | 2일 | - | - |
| **2** | 코드 중복 제거 | ⏳ 대기 | 5일 | 🔴 Critical | 🟠 Medium |
| **3** | 구조 정리 | ⏳ 대기 | 3일 | 🟠 Major | 🟡 Low |
| **4** | Hook 분할 | ⏳ 선택 | 10일 | 🟡 Medium | 🟡 Medium |
| **5** | 문서화 검증 | ⏳ 대기 | 5일 | 🟡 Low | 🟡 Low |
| **6** | 테스트 배포 | ⏳ 대기 | 5일 | 🔴 Critical | 🟡 Low |
| **합계** | - | - | **33일** | - | - |

### Phase 2: 즉시 실행 (가장 영향 큼, 위험 관리 가능)

```python
# 핵심 변경 3가지

# 1️⃣ 중복 코드 제거
❌ moai/core/                    # 제거
❌ moai/utils/timeout.py          # 제거
✅ moai/shared/core/ 통합         # 유지

# 2️⃣ 임포트 경로 통일
❌ from core import ...           # 제거
❌ from utils.timeout import ...   # 제거
✅ from moai_adk.hooks.shared.core.config_cache import ...  # 절대 경로

# 3️⃣ Hook 검증
✅ session_start__show_project_info.py  (작동 확인)
✅ session_start__auto_cleanup.py       (작동 확인)
✅ session_start__config_health_check.py (작동 확인)
⏳ 나머지 6개 Hook                      (테스트 예정)
```

---

## 기대 효과

### 정량적 개선

```
지표                         현재      최적화 후  개선율
──────────────────────────────────────────────────
추적 파일 수                 39개      34개      -13%
총 코드 라인                 4,847L    4,667L    -4%
중복 코드                    180L      0L        -100%
디렉토리 계층                4단계     3단계     -25%
임포트 패턴                  3가지     1가지     -67%
```

### 정성적 개선

| 항목 | 현재 | 최적화 후 | 개선도 |
|------|------|----------|--------|
| **유지보수성** | 68점 | 82점 | ⬆️ +14점 |
| **코드 복잡도** | 높음 | 중간 | ⬇️ 개선 |
| **개발자 이해도** | 낮음 | 높음 | ⬆️ 개선 |
| **테스트 용이성** | 중간 | 높음 | ⬆️ 개선 |
| **SSOT 준수** | 70% | 100% | ⬆️ 개선 |

### 비용 절감

```
현재: 변경 발생 시
  - 중복 코드 수정: 2곳 검수 (30분/회)
  - 임포트 경로 디버깅: 1-2시간/이슈
  - 테스트 범위 파악: 불명확 (위험)
  ──────────────────────
  월간 추정 비용: 15-20시간

최적화 후:
  - 중복 코드 수정: 1곳 검수 (15분/회)
  - 임포트 경로: 자동 검증 (0분)
  - 테스트 범위: 명확한 문서 (5분)
  ──────────────────────
  월간 추정 비용: 5-7시간

절감: 10-15시간/월 (50-60% 감소)
```

---

## 위험 관리

### 주요 위험과 완화 전략

| 위험 | 가능성 | 영향 | 완화 전략 |
|------|--------|------|----------|
| **Hook 실행 실패** | 중간 | 높음 | 각 단계 SessionStart 테스트 + CI/CD |
| **임포트 오류** | 중간 | 높음 | mypy 타입 검사 + Python 문법 검증 |
| **캐시 손상** | 낮음 | 중간 | 캐시 로직 단위 테스트 |
| **SSOT 위반** | 낮음 | 높음 | 템플릿/로컬 동기화 검증 스크립트 |

### 롤백 계획

```bash
# 각 Phase 완료 후 Git 태그
git tag "hooks-phase-2-complete"
git tag "hooks-phase-3-complete"

# 문제 발생 시 즉시 롤백
git reset --hard hooks-phase-2-complete
```

---

## 실행 순서 (권장)

### 주간 계획

```
Week 1: Phase 2 (코드 중복 제거)
  ├─ Mon-Tue: 임포트 경로 전환
  ├─ Wed: 중복 코드 통합
  ├─ Thu: Hook 테스트
  └─ Fri: CI/CD 통과 확인

Week 2: Phase 3 (구조 정리)
  ├─ Mon: 빈 디렉토리 제거
  ├─ Tue: spec_status_hooks.py 이동
  ├─ Wed: __pycache__ 정리
  └─ Thu-Fri: 통합 테스트

Week 3: Phase 4 (선택사항)
  ├─ 필요시만 Hook 분할 실행
  └─ 또는 Phase 5로 진행

Week 4-5: Phase 5-6 (문서화 + 배포)
  ├─ Phase 5: 표준 헤더 추가 + 검증 스크립트
  └─ Phase 6: 최종 테스트 + 릴리스
```

---

## 의사결정 포인트

### 필수 결정 사항

```
✅ Phase 2 실행: 즉시 (가장 영향 큼)
   - 5일 소요
   - Critical 이슈 해결
   - 위험 관리 가능

✅ Phase 3 실행: Phase 2 직후 (높은 의존성)
   - 3일 소요
   - Major 이슈 해결
   - 매우 낮은 위험

❓ Phase 4 실행: 선택사항 (향후 결정 가능)
   - 10일 소요
   - Hook 파일 분할
   - 성능 최적화 목적
   - 추천: 나중에 결정

✅ Phase 5-6: 필수 (품질 보증)
   - 10일 소요
   - 문서화 + 최종 테스트
   - 릴리스 품질 보증
```

---

## 다음 단계

### 즉시 조치 (오늘-내일)

```markdown
## 1. 리포트 검토
- [ ] HOOKS_ANALYSIS_REPORT.md 검토
- [ ] HOOKS_OPTIMIZATION_PLAN.md 검토
- [ ] HOOKS_OPTIMIZATION_METRICS.json 데이터 확인

## 2. 의사결정
- [ ] Phase 2-3 실행 승인
- [ ] Phase 4 연기/실행 결정
- [ ] 담당자 할당

## 3. 준비
- [ ] 개발 브랜치 생성: feature/SPEC-HOOKS-001
- [ ] 백업 생성: .moai/backup/hooks-pre-optimization
- [ ] 테스트 환경 구성
```

### Phase 2 실행 (2-6일)

```markdown
## SPEC 작성 (1-2시간)
- /moai:1-plan "Claude Code Hooks 구조 최적화"
  → SPEC-HOOKS-001 생성
- /clear (토큰 초기화)

## 구현 (3-4일)
- /moai:2-run SPEC-HOOKS-001
  1. 중복 코드 제거 (moai/core → shared/core)
  2. 임포트 경로 통일 (절대 경로)
  3. Hook 검증 (SessionStart 우선)

## 테스트 및 문서화 (1-2일)
- 모든 Hook 실행 테스트
- CI/CD 통과 확인
- 변경 로그 작성

## 배포 (1일)
- /moai:3-sync auto SPEC-HOOKS-001
- PR 생성 및 검토
- main 브랜치 병합
```

---

## 성공 기준

### Phase 2 완료 기준

```
✅ 모든 import 경로가 절대 경로 (moai_adk.hooks.shared.*)
✅ moai/core/ 디렉토리 제거
✅ 중복 코드 0개 확인 (그렘린 스캔)
✅ 모든 Hook 실행 테스트 통과
✅ mypy 타입 검사 통과
✅ Python 문법 검사 통과 (py_compile)
✅ SessionStart Hook 3개 모두 작동 확인
✅ CI/CD 파이프라인 녹색
```

### Phase 3 완료 기준

```
✅ moai/handlers/ 제거
✅ moai/shared/config/ 제거
✅ spec_status_hooks.py → src/moai_adk/cli/ 이동
✅ __pycache__ 정리 및 .gitignore 업데이트
✅ Git 상태: 모든 변경 커밋됨
✅ 템플릿 동기화 확인
```

### 전체 완료 기준 (Phase 6)

```
✅ 모든 Hook 실행 테스트 통과
✅ 캐시 동작 확인 (git-info, version-check)
✅ 에러 복구 테스트 (graceful degradation)
✅ 타임아웃 처리 (5초 제한) 검증
✅ 문서화 완료 (모든 Hook에 표준 헤더)
✅ CI/CD 파이프라인 구성 완료
✅ Release notes v0.27.0 작성
✅ 모든 QA 게이트 통과
```

---

## 추가 리소스

### 상세 문서

1. **HOOKS_ANALYSIS_REPORT.md** (80KB)
   - 현재 상태 상세 분석
   - 공식 스펙 준수 검증
   - 문제점 상세 설명

2. **HOOKS_OPTIMIZATION_PLAN.md** (50KB)
   - 6단계 실행 계획
   - Phase별 상세 작업 절차
   - 코드 변경 예시
   - 테스트 체크리스트

3. **HOOKS_OPTIMIZATION_METRICS.json** (30KB)
   - 정량적 데이터
   - Phase별 예상 결과
   - 위험 평가

### 실행 중 참고

```
╔══════════════════════════════════════════╗
║  필수 참고 문서 (Phase 2 실행 시)        ║
╠══════════════════════════════════════════╣
║  1. HOOKS_OPTIMIZATION_PLAN.md           ║
║     └─ "Phase 2: 코드 중복 제거" 섹션  ║
║                                          ║
║  2. HOOKS_OPTIMIZATION_METRICS.json      ║
║     └─ "files_affected" 목록 참조      ║
║                                          ║
║  3. HOOKS_ANALYSIS_REPORT.md             ║
║     └─ "구조 문제점" 섹션 참조          ║
╚══════════════════════════════════════════╝
```

---

## 최종 권장사항

### 🎯 실행 방침

```
1️⃣  Phase 2-3 즉시 시작 (다음 스프린트)
    - 가장 높은 ROI (위험 대비 효과)
    - 개발자 생산성 향상
    - SSOT 준수 개선

2️⃣  Phase 4 연기 (필요시만)
    - Hook 분할은 후순위
    - 현재 기능 문제 없음
    - 미래 리팩토링 시 고려

3️⃣  Phase 5-6 필수 실행
    - 품질 보증 목적
    - 릴리스 v0.27.0 조건
    - 자동화 기반 구성

4️⃣  주간 단위 실행
    - Week 1: Phase 2 (임포트 경로)
    - Week 2: Phase 3 (정리)
    - Week 3-4: Phase 5-6 (문서 + 배포)
```

### ⚠️ 중요 주의사항

```
❌ 피해야 할 것:
  - Phase 1-6을 동시에 실행 (의존성 있음)
  - Hook 실행 테스트 없이 진행
  - 템플릿-로컬 동기화 확인 생략
  - 위험 완화 계획 무시

✅ 해야 할 것:
  - Phase별로 순차적 실행
  - 각 단계마다 SessionStart Hook 테스트
  - 변경 전/후 SSOT 검증
  - CI/CD 파이프라인 활용
```

---

## 승인 및 추적

| 항목 | 상태 | 담당 | 날짜 |
|------|------|------|------|
| 분석 리포트 | ✅ 완료 | cc-manager | 2025-11-19 |
| 최적화 계획 | ✅ 완료 | cc-manager | 2025-11-19 |
| 의사결정 (Phase 2-3) | ⏳ 대기 | PO/PM | - |
| Phase 2 실행 | ⏳ 대기 | Dev Team | - |
| Phase 2 검증 | ⏳ 대기 | QA | - |
| Phase 3-6 | ⏳ 계획 중 | Dev Team | - |

---

**문서 작성**: 2025-11-19
**버전**: 1.0.0
**상태**: 🔵 검토 대기 중 (Phase 2-3 승인 필요)

**다음 회의**: Phase 2 실행 계획 최종 결정
