# /alfred:tag-renumber

브랜치 충돌이나 도메인 재구성 시 필요한 TAG 재번호 작업을 안전하게 수행합니다.

> **Note**: 복잡한 재번호 전략이 필요할 때 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 제공합니다. 사용자 승인이 필요한 경우 `AskUserQuestion` 도구를 통해 TUI 선택 메뉴를 제공합니다.

## 사용법

```bash
/alfred:tag-renumber --domain <DOMAIN>
/alfred:tag-renumber --all
/alfred:tag-renumber --dry-run --domain <DOMAIN>
```

## 옵션

- `--domain <DOMAIN>`: 특정 도메인만 재번호
- `--all`: 모든 도메인 재번호
- `--dry-run`: 실행 시뮬레이션만 수행
- `--from <N>`: 특정 번호부터 재번호 시작
- `--backup`: 재번호 전 자동 백업 생성

## 설명

브랜치 간 TAG ID 충돌이나 도메인 구조 변경 시, 충돌을 피하기 위해 일괄 재번호를 수행합니다. 모든 참조(topline, Relates, 주석)를 자동으로 업데이트하고 완전한 추적성을 유지합니다.

## 사용 시나리오

### 1. 브랜치 병합 충돌 해결
```bash
# develop 브랜치와 feature 브랜치가 동시에 AUTH-021을 예약한 경우
/alfred:tag-renumber --domain AUTH --dry-run
# 결과 확인 후 실제 실행
/alfred:tag-renumber --domain AUTH
```

### 2. 도메인 재구성
```bash
# 패키지 구조 변경으로 도메인 전체 재번호 필요 시
/alfred:tag-renumber --all
```

### 3. 특정 번호부터 재설정
```bash
# AUTH 도메인을 050번부터 재설정
/alfred:tag-renumber --domain AUTH --from 50
```

## 처리 절차

### 1단계: 분석 (Analysis)
- 충돌 상황 분석
- 재번호 대상 파일 목록 생성
- 영향 범위 평가

### 2단계: 백업 (Backup)
- 현재 ledger 스냅샷 생성
- 영향 파일 백업 (선택적)

### 3단계: 재번호 (Renumber)
- 카운터 재설정
- TAG ID 재할당
- 모든 참조 업데이트

### 4단계: 검증 (Verification)
- 참조 무결성 확인
- 체인 연결성 검증
- 원장 일관성 확인

## 🧠 복잡한 재번호 전략

### @sequential-thinking MCP 활용

TAG 재번호 시 다음 복잡한 상황에서는 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 수행합니다:

#### 복잡한 재번호가 필요한 경우

1. **대규모 도메인 재구성**
   - 여러 도메인의 동시 재번호가 필요할 때
   - 50개 이상의 TAG가 영향을 받는 대규모 재번호
   - 여러 팀이 동시에 작업 중인 도메인 재번호

2. **충돌 해결 복잡성**
   - 여러 브랜치에서 동일 ID 범위를 사용하는 경우
   - 원격 저장소와 로컬 저장소의 ID 불일치
   - 병합 충돌로 인한 TAG ID 중복

3. **전략적 재번호 결정**
   - 도메인 분리/병합에 따른 재번호
   - 프로젝트 구조 변경에 따른 재정렬
   - 표준화 정책 적용에 따른 대규모 재번호

#### @sequential-thinking 통합 패턴

```bash
# 복잡한 재번호 전략 분석
/alfred:tag-renumber --domain AUTH --complex-strategy

# 재번호 영향 분석 및 위험 평가
/alfred:tag-renumber --impact-analysis --all-domains
```

**분석 과정**:
1. **충돌 분석**: 현재 ID 사용 패턴과 충돌 지점 식별
2. **영향 평가**: 재번호가 코드, 테스트, 문서에 미치는 영향
3. **의존성 맵핑**: TAG 간 의존 관계와 연쇄 영향 분석
4. **위험 평가**: 재번호 실패 시 영향과 복구 전략
5. **실행 계획**: 단계별 재번호 실행 전략 수립

### AskUserQuestion 통합

복잡한 재번호 결정이 필요할 때 사용자의 승인을 받기 위해 `AskUserQuestion`을 사용합니다:

#### 재번호 전략 선택 예시

```bash
# 도메인 재번호 전략 선택
AUTH 도메인 재번호 방식을 선택하세요:

[ ] 점진적 재번호: 안전성 확인하며 순차 재번호
[ ] 일괄 재번호: 모든 TAG를 동시에 재번호
[ ] 분할 재번호: 하위 도메인별로 나누어 재번호
[ ] 전문가 상담: TAG 시스템 전문가 컨설팅
```

#### 위험 승인 요청 예시

```bash
# 고위험 재번호 승인 요청
경고: 이 재번호는 다음에 영향을 미칩니다:
- 영향 도메인: AUTH, USER, PAY (3개)
- 영향 TAG: 67개 (SPEC: 23, TEST: 22, CODE: 22)
- 영향 파일: 89개
- 예상 시간: 1-2시간
- 롤백 복잡도: 높음

진행 방법을 선택하세요:
[ ] 재번호 계획 재검토
[ ] 영향 범위 축소
[ ] 안전 장치 강화
[ ] 재번호 실행 승인
```

## 재번호 규칙

### ID 재할당
```
기존: @SPEC:AUTH-021, @TEST:AUTH-021, @CODE:AUTH-021
신규: @SPEC:AUTH-025, @TEST:AUTH-025, @CODE:AUTH-025
```

### 자동 업데이트 대상
- **Topline TAG**: 모든 파일의 상단 TAG
- **Relates 참조**: `# Relates: @TAG:DOMAIN-NNN`
- **주석 참조**: `# SPEC:@SPEC:DOMAIN-NNN`
- **문서 내부**: 모든 TAG 패턴

### 제외 대상
- 외부 시스템 참조 (API 문서 등)
- 생성된 빌드 결과물
- 이미 DEPRECATED/MIGRATED된 TAG

## 예시 실행

### Dry Run 모드
```bash
/alfred:tag-renumber --domain AUTH --dry-run

📋 재번호 시뮬레이션 결과:
- 대상 도메인: AUTH
- 현재 TAG: 12개 (AUTH-001 ~ AUTH-012)
- 재번호 필요: 4개 (AUTH-009 ~ AUTH-012)
- 영향 파일: 18개

🔄 재번호 계획:
  @SPEC:AUTH-009 → @SPEC:AUTH-013
  @TEST:AUTH-009 → @TEST:AUTH-013
  @CODE:AUTH-009 → @CODE:AUTH-013
  ...

⚠️ 실제 실행하려면 --dry-run 옵션을 제거하세요.
```

### 실제 실행
```bash
/alfred:tag-renumber --domain AUTH

🔄 재번호 진행 중...
✅ 백업 생성: .moai/tags/backup_renumber_20251106_103022/
✅ 카운터 재설정: AUTH (001 → 013)
✅ TAG 재할당: 12개
✅ 파일 업데이트: 18개
✅ 원장 동기화: 완료
✅ 참조 검증: 완료

📊 최종 결과:
- 재번호된 TAG: 12개
- 업데이트된 파일: 18개
- 소요 시간: 2.3초
- 상태: 완전 성공
```

## 안전장치

### 자동 백업
```bash
# 재번호 전 항상 백업 생성
.moai/tags/backup_renumber_YYYYMMDD_HHMMSS/
├── ledger.jsonl.backup
├── index.json.backup
├── counters.json.backup
└── affected_files/
    ├── src/auth/service.py.backup
    ├── tests/test_auth.py.backup
    └── ...
```

### 롤백 기능
```bash
# 재번호 취소 (최근 백업으로 복구)
/alfred:tag-audit --restore backup_renumber_20251106_103022
```

### 충돌 방지
- 파일 잠금 메커니즘으로 동시 실행 방지
- 진행 상황 실시간 표시
- 중단 시 롤백 자동 제안

## 영향 분석 리포트

```bash
/alfred:tag-renumber --domain AUTH --detailed

📈 상세 영향 분석:
- 재번호 대상: 12개 TAG
- 영향 파일: 18개
  - 코드 파일: 8개
  - 테스트 파일: 6개
  - 문서 파일: 4개

🔗 체인 상태:
- 완전 체인: 8개 (유지)
- 부분 체인: 4개 (재번호 후 완전 예상)

⚠️ 주의사항:
- 외부 API 문서 참조 1건 확인 필요
- CI/CD 파이프라인 업데이트 필요
```

## 관련 명령어

- `/alfred:tag-audit`: 재번호 전후 상태 비교
- `/alfred:tag-migrate`: 개별 TAG 마이그레이션
- `/alfred:3-sync`: 재번호 후 전체 동기화

## 주의사항

- **브랜치 병합 전 실행**: 충돌을 미리 해결하기 위해 병합 전에 실행 권장
- **CI/CD 고려**: 재번호 후 파이프라인 업데이트 필요
- **팀 동기화**: 대규모 재번호 시 팀원들에게 사전 통보 필요
- **외부 참조**: 외부 시스템이나 문서의 참조는 수동 업데이트 필요

## 성공 기준

- **데이터 무결성**: 모든 참조 정확하게 업데이트
- **추적성 유지**: 모든 이력 보존
- **안전성**: 롤백 기능으로 복구 가능
- **성능**: 대규모 프로젝트에서도 1분 내 완료
- **검증**: 재번호 후 100% 참조 무결성 확인