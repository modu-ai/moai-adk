# /alfred:tag-audit

TAG 시스템의 건강 상태를 진단하고, 무결성을 검증하며, 필요한 조치를 제안합니다.

> **Note**: 복잡한 전략 결정이 필요할 때 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 제공합니다. 사용자 승인이 필요한 경우 `AskUserQuestion` 도구를 통해 TUI 선택 메뉴를 제공합니다.

## 사용법

```bash
/alfred:tag-audit [--options]
```

## 옵션

- `--full-scan`: 전체 프로젝트 스캔 (기본값: 변경 파일만)
- `--rebuild`: ledger 손상 시 전체 재구성
- `--period=7d`: 특정 기간 동안의 변경만 분석
- `--domain=AUTH`: 특정 도메인만 분석
- `--detailed`: 상세 분석 리포트 생성

## 설명

TAG 시스템의 전반적인 건강 상태를 진단하고, KPI를 보고하며, 문제가 있는 항목을 식별하여 해결책을 제안합니다.

## 진단 항목

### 1. 체인 완전성 (Chain Integrity)
- **Full**: SPEC → TEST → CODE → DOC 모두 연결
- **SpecOnly**: SPEC만 존재
- **SpecTest**: SPEC와 TEST만 연결
- **SpecTestCode**: SPEC, TEST, CODE 연결 (DOC 누락)
- **Orphan**: 연결이 끊어진 상태

### 2. 중복 분석 (Duplication Analysis)
- **Primary 중복**: 동일 ID의 primary 파일이 2개 이상
- **ID 충돌**: 서로 다른 내용의 동일 ID
- **Relates 중복**: 불필요한 중복 참조

### 3. 만료 관리 (Expiration Management)
- **예약 만료**: 72시간 초과된 예약
- **경고 대상**: 24시간 이내 만료 예정
- **고아 예약**: 연결되지 않은 예약

### 4. 도메인 일관성 (Domain Consistency)
- **경로-도메인 불일치**: 파일 경로와 TAG 도메인 불일치
- **도메인 분포**: 각 도메인별 TAG 분포 현황

### 5. 품질 지표 (Quality Metrics)
- **형식 준수**: @TYPE:DOMAIN-NNN 형식 준수율
- **위치 준수**: topline 위치 규칙 준수율
- **Relates 사용**: Related 파일에서의 Relates 사용률

## KPI 보고

### 기본 KPI
```
📊 TAG 시스템 건강 지표
- 전체 TAG: 156개
- 완전 체인: 89개 (57.1%) 🎯 목표: 90%
- 고아 TAG: 3개 (1.9%) 🎯 목표: 0%
- 중복 TAG: 0개 (0%) ✅
- 예약 TAG: 8개 (5.1%)
```

### 도메인별 분포
```
📈 도메인별 분포
- AUTH: 45개 (28.8%)
- CORE: 38개 (24.4%)
- USER: 32개 (20.5%)
- PAY: 28개 (17.9%)
- DOCS: 13개 (8.3%)
```

### 상태별 분포
```
🔄 상태별 분포
- active: 125개 (80.1%)
- reserved: 23개 (14.7%)
- deprecated: 6개 (3.8%)
- migrated: 2개 (1.3%)
- rescinded: 0개 (0%)
```

## 🧠 복잡한 전략 및 추론

### @sequential-thinking MCP 활용

TAG 시스템 감사 시 다음 복잡한 상황에서는 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 수행합니다:

#### 복잡한 결정이 필요한 경우

1. **대규모 마이그레이션 전략**
   - 여러 도메인의 동시 재구성이 필요할 때
   - 영향 분석이 50개 이상의 TAG에 미칠 때
   - 롤백 계획이 필요한 복잡한 변경일 때

2. **다중 기준 의사결정**
   - 무결성 vs. 호환성 trade-off가 필요할 때
   - 성능 vs. 안전성 우선순위 결정이 필요할 때
   - 단기 수정 vs. 장기 구조 개선 선택이 필요할 때

3. **위험-이익 분석**
   - 데이터 손실 가능성이 있는 변경 시
   - 기존 워크플로우에 큰 영향을 미치는 변경 시
   - 팀 생산성에 영향을 미치는 변경 시

#### @sequential-thinking 통합 패턴

```bash
# 복잡한 분석이 필요한 감사
/alfred:tag-audit --detailed --complex-analysis

# 자동 수정 전략 제안 요청
/alfred:tag-audit --strategy-recommendation
```

**분석 과정**:
1. **문제 식별**: 현재 상태와 목표 상태 간 격차 분석
2. **영향 평가**: 변경이 미치는 영향의 범위와 깊이 평가
3. **대안 수립**: 가능한 해결책과 각각의 장단점 분석
4. **위험 평가**: 각 대안의 위험도와 성공 확률 평가
5. **권장 사항**: 최적의 전략과 실행 계획 제안

### AskUserQuestion 통합

복잡한 전략 결정이 필요할 때 사용자의 승인을 받기 위해 `AskUserQuestion`을 사용합니다:

#### 전략 선택 메뉴 예시

```bash
# 대규모 TAG 재구성 전략 선택
[ ] 보수적 접근: 단계적 마이그레이션 (안전성 우선)
[ ] 중도 접근: 부분 병렬 처리 (균형 접근)
[ ] 공격적 접근: 일괄 처리 (속도 우선)
[ ] 맞춤 전략: 직접 계획 수립
```

#### 위험 승인 요청 예시

```bash
# 고위험 변경 승인 요청
경고: 이 변경은 156개의 TAG에 영향을 미칩니다
- 예상 영향: 높음
- 롤백 복잡도: 중간
- 소요 시간: 2-3시간

진행하시겠습니까?
[ ] 변경 계획 검토 후 재결정
[ ] 위험 감소 전략 적용
[ ] 전문가 상담 후 결정
[ ] 변경 실행 승인
```

## 자동 수정 제안

### 1단계: 자동 수정 가능
```bash
# 자동 수정 실행
/alfred:tag-audit --auto-fix
```

- README의 topline TAG 제거 및 Relates 변환
- 중복 primary 관련 자동 수정 제안
- 만료 예약 RESCIND 처리

### 2단계: 수동 검토 필요
- 도메인 변경이 필요한 항목
- 복잡한 체인 재구성
- 마이그레이션이 필요한 대규모 변경

## 예시 출력

### 기본 리포트
```
🔍 TAG 시스템 감사 결과

⚠️ 주요 문제 (3건):
1. 고아 TAG: @TEST:USER-005 (연결된 SPEC 없음)
   조치: /alfred:1-plan으로 SPEC 생성 또는 /alfred:tag-migrate

2. 도메인 불일치: src/auth/user_service.py (TAG: @CODE:USER-012)
   조치: /alfred:tag-migrate @CODE:USER-012 @CODE:AUTH-015

3. 만료 예약: @SPEC:PAY-008 (72시간 경과)
   조치: /alfred:tag-rescind @SPEC:PAY-008

💡 자동 수정 제안 (2건):
1. README.md topline TAG 제거 가능
2. tests/test_auth.py 중복 Relates 정리 가능

✅ 전체 건강 지표: 양호 (KPI 75% 달성)
```

### 상세 분석 모드
```bash
/alfred:tag-audit --detailed --domain=AUTH
```

## 복구 기능

### Ledger 재구성
```bash
# Ledger 손상 시 전체 재구성
/alfred:tag-audit --rebuild
```

- 기존 파일 스캔으로 ledger 재생성
- 인덱스 재구성
- 스냅샷 복구 기능

### 스냅샷 관리
```bash
# 최근 스냅샷 목록
/alfred:tag-audit --snapshots

# 특정 스냅샷으로 복구
/alfred:tag-audit --restore snapshot_20251106_103022
```

## 관련 명령어

- `/alfred:tag-reserve`: TAG 예약
- `/alfred:tag-migrate`: TAG 마이그레이션
- `/alfred:tag-renumber`: 재번호
- `/alfred:3-sync`: 전체 동기화 검증

## 주기적 실행 권장

```bash
# 매일 자동 실행 (cron 등)
0 9 * * * /alfred:tag-audit --period=24h

# 주간 상세 보고
0 9 * * 1 /alfred:tag-audit --detailed --full-scan
```

## 성공 기준

- **전체 TAG**: 정확한 파악
- **체인 분석**: 모든 연결 상태 확인
- **문제 식별**: 95% 이상의 문제 탐지
- **수정 제안**: 실용 가능한 해결책 제공
- **복구 지원**: 데이터 복구 기능 제공