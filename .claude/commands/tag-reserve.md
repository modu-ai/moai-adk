# /alfred:tag-reserve

TAG를 수동으로 예약하여 SPEC-First 원칙을 유지하면서도 유연성을 확보합니다.

> **Note**: 복잡한 예약 전략이 필요할 때 `@sequential-thinking` MCP를 사용하여 체계적인 분석을 제공합니다. 사용자 승인이 필요한 경우 `AskUserQuestion` 도구를 통해 TUI 선택 메뉴를 제공합니다.

## 사용법

```bash
/alfred:tag-reserve <DOMAIN>
```

## 설명

지정된 도메인에 대한 새로운 SPEC TAG를 예약하고, 최소한의 SPEC 스텁을 자동으로 생성합니다.

### 예시

```bash
/alfred:tag-reserve AUTH
# 결과: @SPEC:AUTH-021 예약 및 .moai/specs/AUTH-021/spec.md 생성
```

## 처리 절차

1. **도메인 유효성 검사**: 지정된 도메인이 정책에 정의되어 있는지 확인
2. **카운터 증가**: 해당 도메인의 다음 번호를 확보 (파일 잠금으로 경쟁 방지)
3. **TAG ID 생성**: `@SPEC:{DOMAIN}-{NNN}` 형식으로 고유 ID 생성
4. **SPEC 스텁 생성**: `.moai/specs/{DOMAIN}-{NNN}/spec.md`에 템플릿 기반 스텁 생성
5. **원장 기록**: ledger.jsonl에 RESERVE 작업 기록
6. **인덱스 갱신**: index.json에 예약 정보 반영

## 생성되는 파일

```
.moai/specs/AUTH-021/spec.md
```

## 리턴 값

- **성공**: 예약된 TAG ID와 생성된 SPEC 파일 경로
- **실패**: 오류 메시지와 해결 제안

## 관련 명령어

- `/alfred:1-plan --reserve`: 자동 예약 (코드 우선 시)
- `/alfred:tag-audit`: 예약 상태 확인
- `/alfred:tag-renumber`: 도메인별 재번호

## 주의사항

- 예약된 TAG는 24시간 후 경고, 72시간 후 자동 만료 (RESCIND)
- 동일 도메인의 여러 예약은 자동으로 다음 번호를 부여
- 브랜치 병합 시 충돌 가능성 있으므로 `/alfred:tag-renumber` 사용 권장

## 예시 출력

```
✅ TAG 예약 완료:
- TAG ID: @SPEC:AUTH-021
- SPEC 파일: .moai/specs/AUTH-021/spec.md
- 만료 시간: 2025-11-09 10:30:00 (72시간 후)

⚠️ 24시간 내에 상세 내용을 채워주세요.
```