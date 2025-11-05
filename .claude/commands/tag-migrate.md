# /alfred:tag-migrate

기존 TAG를 새로운 TAG로 마이그레이션하여 추적성을 유지하면서 변경을 관리합니다.

> **Note**: 복잡한 마이그레이션 전략이 필요할 때 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 제공합니다. 사용자 승인이 필요한 경우 `AskUserQuestion` 도구를 통해 TUI 선택 메뉴를 제공합니다.

## 사용법

```bash
/alfred:tag-migrate <OLD_TAG> <NEW_TAG>
```

## 설명

기존의 TAG ID를 새로운 TAG ID로 변경하고, 관련된 모든 파일들의 참조를 자동으로 업데이트합니다. 삭제 대신 마이그레이션 기록을 통해 완전한 추적성을 유지합니다.

### 예시

```bash
/alfred:tag-migrate @CODE:AUTH-009 @CODE:AUTH-021
```

## 처리 절차

1. **기존 TAG 확인**: OLD_TAG가 존재하는지 유효성 검사
2. **새 TAG 생성**: NEW_TAG가 충돌 없는지 확인하고 필요시 예약
3. **영향 분석**: OLD_TAG를 참조하는 모든 파일 목록 생성
4. **파일 업데이트**: 관련 파일들의 topline TAG와 Relates 참조 업데이트
5. **원장 기록**: MIGRATE 작업을 ledger.jsonl에 기록
6. **인덱스 갱신**: index.json에서 관계 업데이트
7. **이력 보존**: OLD_TAG는 "migrated" 상태로 유지

## 마이그레이션 유형

### 도메인 변경
```bash
/alfred:tag-migrate @CODE:AUTH-009 @CODE:USER-001
# AUTH 도메인에서 USER 도메인으로 변경
```

### 기능 분리
```bash
/alfred:tag-migrate @CODE:AUTH-009 @CODE:AUTH-021
# 하나의 기능을 두 개로 분리
```

### 기능 병합
```bash
/alfred:tag-migrate @CODE:AUTH-009 @CODE:AUTH-001
# 두 기능을 하나로 병합 (기존 것은 마이그레이션)
```

## 자동 업데이트 대상

- **Topline TAG**: 파일 상단의 `@TAG:DOMAIN-NNN` 형식
- **Relates 참조**: `# Relates: @TAG:DOMAIN-NNN` 형식
- **주석 참조**: `# SPEC:@SPEC:DOMAIN-NNN` 형식
- **문서 내부**: 모든 TAG 참조 패턴

## 리턴 값

- **성공**: 마이그레이션된 파일 목록과 변경 사항 요약
- **실패**: 오류 메시지와 충돌 현황

## 예시 출력

```
✅ TAG 마이그레이션 완료:
- 이전: @CODE:AUTH-009
- 신규: @CODE:AUTH-021

📝 업데이트된 파일 (3개):
- src/auth/service.py (topline)
- tests/test_auth.py (Relates)
- docs/auth.md (내부 참조)

📊 원장 기록 완료:
- MIGRATE: @CODE:AUTH-009 -> @CODE:AUTH-021
- 상태: AUTH-009는 'migrated', AUTH-021는 'active'
```

## 관련 명령어

- `/alfred:tag-audit`: 마이그레이션 현황 확인
- `/alfred:tag-renumber`: 일괄 재번호 (브랜치 충돌 시)
- `/alfred:tag-deprecated`: 폐기 처리 (마이그레이션 대안)

## 🧠 복잡한 마이그레이션 전략

### @sequential-thinking MCP 활용

TAG 마이그레이션 시 다음 복잡한 상황에서는 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 수행합니다:

#### 복잡한 마이그레이션이 필요한 경우

1. **대규모 연쇄 마이그레이션**
   - 10개 이상의 TAG가 연속적으로 마이그레이션 필요할 때
   - 여러 도메인에 걸친 동시 마이그레이션이 필요할 때
   - 의존성 그래프 재구성이 필요할 때

2. **위험도 높은 마이그레이션**
   - 핵심 비즈니스 로직 관련 TAG 마이그레이션
   - 프로덕션 환경에 직접 영향을 미치는 TAG
   - 롤백이 복잡한 구조적 변경

3. **전략적 마이그레이션 결정**
   - 기능 분리 vs 병합 전략 선택
   - 도메인 재구성 마이그레이션
   - 레거시 시스템 현대화 마이그레이션

#### @sequential-thinking 통합 패턴

```bash
# 복잡한 마이그레이션 전략 분석
/alfred:tag-migrate @CODE:AUTH-009 @CODE:AUTH-021 --complex-strategy

# 마이그레이션 영향 분석 요청
/alfred:tag-migrate --impact-analysis @CODE:AUTH-009
```

**분석 과정**:
1. **의존성 분석**: 마이그레이션 대상 TAG와 연관된 모든 의존성 맵핑
2. **영향 평가**: 변경이 코드, 테스트, 문서에 미치는 영향 평가
3. **위험 식별**: 잠재적 위험 요소와 실패 시나리오 분석
4. **롤백 계획**: 마이그레이션 실패 시 복구 전략 수립
5. **실행 계획**: 단계별 마이그레이션 실행 계획 수립

### AskUserQuestion 통합

복잡한 마이그레이션 결정이 필요할 때 사용자의 승인을 받기 위해 `AskUserQuestion`을 사용합니다:

#### 마이그레이션 전략 선택 예시

```bash
# 기능 분리 마이그레이션 전략
@CODE:AUTH-009를 다음과 같이 분리합니다:
- @CODE:AUTH-021: 인증 로직
- @CODE:AUTH-022: 권한 로직
- @CODE:AUTH-023: 세션 관리

어떤 전략을 선택하시겠습니까?
[ ] 단계적 분리: 안정성 확인하며 순차 분리
[ ] 동시 분리: 모든 기능을 동시에 분리
[ ] 보수적 접근: 테스트 완료 후 분리
[ ] 전문가 상담: 백엔드 전문가 컨설팅
```

#### 위험 승인 요청 예시

```bash
# 고위험 마이그레이션 승인 요청
경고: 이 마이그레이션은 다음에 영향을 미칩니다:
- 영향 파일: 23개 (코드: 12, 테스트: 8, 문서: 3)
- 예상 시간: 1-2시간
- 롤백 복잡도: 높음
- 다른 작업 영향: 있음 (진행 중인 AUTH 도메인 작업)

진행 방법을 선택하세요:
[ ] 마이그레이션 계획 재검토
[ ] 영향 범위 축소
[ ] 안전 장치 강화
[ ] 마이그레이션 실행 승인
```

## 주의사항

- 마이그레이션은 되돌릴 수 없으니 신중하게 실행
- 브랜치 병합 전에 충돌 여부 확인 필요
- 마이그레이션 후에는 반드시 테스트 실행 권장
- 대규모 마이그레이션은 Batch Guard가 동작할 수 있음

## 원장 기록 형식

```json
{
  "ts": "2025-11-06T10:30:00Z",
  "op": "MIGRATE",
  "from": "@CODE:AUTH-009",
  "to": "@CODE:AUTH-021",
  "reason": "기능 분리",
  "actor": "user:goos",
  "affected_files": ["src/auth/service.py", "tests/test_auth.py"]
}
```