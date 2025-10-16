# SPEC-VALID-001 수락 기준

> SPEC 메타데이터 검증 자동화 - 상세 테스트 시나리오

---

## Given-When-Then 테스트 시나리오

### 정상 케이스 (6개)

#### TC-001: 모든 필수 필드가 올바르게 작성된 경우

**Given**:
- SPEC 문서가 7개 필수 필드를 모두 포함함
- 각 필드가 올바른 형식으로 작성됨
- HISTORY 섹션이 존재함

**When**:
- `/alfred:1-spec "사용자 인증"` 실행
- spec-validator가 메타데이터를 검증

**Then**:
- 검증이 통과됨
- 성공 메시지가 출력됨: `✅ SPEC-AUTH-001 메타데이터 검증 통과`
- Git 커밋이 자동으로 생성됨
- Draft PR이 생성됨

---

#### TC-002: 선택 필드까지 모두 작성된 경우

**Given**:
- SPEC 문서가 7개 필수 필드 + 9개 선택 필드를 모두 포함함
- 모든 필드가 올바른 형식으로 작성됨
- HISTORY 섹션이 존재함

**When**:
- spec-validator가 메타데이터를 검증

**Then**:
- 검증이 통과됨
- 선택 필드 검증도 성공함
- depends_on 관계가 그래프로 시각화됨

---

#### TC-003: 복합 도메인 ID가 올바르게 작성된 경우

**Given**:
- SPEC ID가 `UPDATE-REFACTOR-001` (하이픈 2개)
- 디렉토리명이 `SPEC-UPDATE-REFACTOR-001/`
- 모든 필수 필드가 올바름

**When**:
- spec-validator가 디렉토리명과 ID를 검증

**Then**:
- 검증이 통과됨
- 경고 메시지 없음

---

#### TC-004: 의존성 관계가 올바르게 작성된 경우

**Given**:
- SPEC A가 `depends_on: [USER-001, AUTH-001]`을 포함함
- USER-001과 AUTH-001 SPEC이 이미 존재함
- 순환 의존성이 없음

**When**:
- dependency-checker가 의존성 그래프를 생성

**Then**:
- 검증이 통과됨
- 의존성 그래프가 정상적으로 생성됨
- 실행 순서가 자동 계산됨: USER-001 → AUTH-001 → SPEC A

---

#### TC-005: 날짜 형식이 정확한 경우

**Given**:
- `created: 2025-10-16`
- `updated: 2025-10-16`
- YYYY-MM-DD 형식 준수

**When**:
- spec-validator가 날짜 필드를 검증

**Then**:
- 검증이 통과됨
- 날짜 유효성 확인 (실제 존재하는 날짜)

---

#### TC-006: HISTORY 섹션이 올바르게 작성된 경우

**Given**:
- HISTORY 섹션이 존재함
- 최소 1개 버전 항목 존재 (v0.0.1)
- AUTHOR, REASON 필드 포함

**When**:
- spec-validator가 HISTORY 섹션을 파싱

**Then**:
- 검증이 통과됨
- 버전 히스토리가 정상적으로 추출됨

---

### 오류 케이스 (6개)

#### TC-007: 필수 필드 누락 (id 누락)

**Given**:
- SPEC 문서에 `id` 필드가 없음
- 다른 필드는 모두 존재함

**When**:
- spec-validator가 메타데이터를 검증

**Then**:
- 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ SPEC 메타데이터 검증 실패: 필수 필드 'id' 누락
    → 추가 필요: id: AUTH-001
    → 참고: .moai/memory/spec-metadata.md
  ```
- 작업이 중단됨 (Git 커밋 생성 안 됨)

---

#### TC-008: version 형식 오류

**Given**:
- `version: 0.0.1a` (잘못된 형식)
- 올바른 형식: `0.0.1`

**When**:
- spec-validator가 version 필드를 검증

**Then**:
- 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ version 필드 형식 오류: 0.0.1a
    → 올바른 형식: MAJOR.MINOR.PATCH (예: 0.0.1)
    → 현재 값을 0.0.1로 수정하세요
  ```

---

#### TC-009: author 형식 오류

**Given**:
- `author: Goos` (`@` 누락)
- 올바른 형식: `@Goos`

**When**:
- spec-validator가 author 필드를 검증

**Then**:
- 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ author 필드 형식 오류: Goos
    → 올바른 형식: @{GitHub ID} (예: @Goos)
    → 현재 값을 @Goos로 수정하세요
  ```

---

#### TC-010: 중복 SPEC ID 발견

**Given**:
- 새로운 SPEC ID가 `AUTH-001`
- `.moai/specs/SPEC-AUTH-001/` 디렉토리가 이미 존재함

**When**:
- duplicate-checker가 중복 ID를 검색
- `rg "@SPEC:AUTH-001" -n .moai/specs/` 실행

**Then**:
- 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ 중복 SPEC ID 발견: AUTH-001
    → 기존 SPEC: .moai/specs/SPEC-AUTH-001/spec.md
    → 다른 ID를 사용하거나 기존 SPEC을 업데이트하세요
    → 권장: AUTH-002, AUTH-TOKEN-001
  ```

---

#### TC-011: 순환 의존성 발견

**Given**:
- SPEC A: `depends_on: [B]`
- SPEC B: `depends_on: [C]`
- SPEC C: `depends_on: [A]` ← 순환 의존성

**When**:
- dependency-checker가 의존성 그래프를 생성

**Then**:
- 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ 순환 의존성 발견: A → B → C → A
    → 다음 중 하나의 의존성을 제거하세요:
      - SPEC A의 depends_on에서 B 제거
      - SPEC C의 depends_on에서 A 제거
  ```

---

#### TC-012: HISTORY 섹션 누락

**Given**:
- SPEC 문서에 HISTORY 섹션이 없음

**When**:
- spec-validator가 HISTORY 섹션을 검증

**Then**:
- 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ HISTORY 섹션 누락
    → 추가 필요:
      ## HISTORY
      ### v0.0.1 (2025-10-16)
      - **INITIAL**: [설명]
      - **AUTHOR**: @Goos
  ```

---

## 품질 게이트 기준

### 필수 조건 (Definition of Done)

- [ ] 모든 정상 케이스 (TC-001 ~ TC-006) 테스트 통과
- [ ] 모든 오류 케이스 (TC-007 ~ TC-012) 테스트 통과
- [ ] 테스트 커버리지 ≥ 90%
- [ ] 검증 수행 시간 ≤ 500ms (단일 SPEC 기준)
- [ ] 중복 ID 검색 시간 ≤ 1초 (.moai/specs/ 전체 스캔)
- [ ] 오류 메시지 가독성 확인 (사용자 피드백)

### 성능 벤치마크

| 항목 | 목표 | 측정 방법 |
|------|------|-----------|
| 필수 필드 검증 | ≤ 50ms | 단위 테스트 시간 측정 |
| 필드 형식 검증 | ≤ 100ms | 정규 표현식 실행 시간 |
| 중복 ID 검색 | ≤ 1초 | Grep 명령 실행 시간 |
| 순환 의존성 검증 | ≤ 2초 | 그래프 알고리즘 실행 시간 |
| 전체 검증 | ≤ 3초 | 통합 테스트 실행 시간 |

---

## 보안 검증

### 보안 테스트 케이스

#### STC-001: YAML 임의 코드 실행 방지

**Given**:
- 악의적인 YAML 코드가 포함된 SPEC 문서
- 예: `!!python/object/apply:os.system ["rm -rf /"]`

**When**:
- metadata-parser가 YAML을 파싱

**Then**:
- Safe YAML 파서가 코드 실행을 차단함
- 파싱 오류 메시지가 출력됨
- 시스템 파일이 삭제되지 않음

---

#### STC-002: 경로 탐색 공격 방지

**Given**:
- 디렉토리명이 `../../etc/passwd`
- 악의적인 경로 탐색 시도

**When**:
- spec-validator가 디렉토리명을 검증

**Then**:
- 경로 검증이 실패함
- 오류 메시지가 출력됨:
  ```
  ❌ 잘못된 디렉토리명: ../../etc/passwd
    → 허용된 형식: SPEC-{ID}/ (.moai/specs/ 내부만 허용)
  ```

---

## 통합 시나리오

### 전체 워크플로우 테스트

**Scenario**: `/alfred:1-spec` 실행부터 Draft PR 생성까지

**Steps**:
1. 사용자가 `/alfred:1-spec "사용자 인증"` 실행
2. spec-builder가 SPEC 문서 3개 생성 (spec.md, plan.md, acceptance.md)
3. spec-validator가 자동으로 메타데이터 검증 수행
4. 검증 통과 시:
   - Git 브랜치 생성 (feature/SPEC-AUTH-001)
   - Git 커밋 생성 (🔴 RED: ...)
   - Draft PR 생성 (develop ← feature/SPEC-AUTH-001)
5. 검증 실패 시:
   - 오류 메시지 출력
   - 작업 중단
   - 사용자에게 수정 방법 안내

**Expected Result**:
- 검증 통과 시: Draft PR URL 반환
- 검증 실패 시: 구체적인 오류 메시지 및 수정 가이드

---

## 회귀 테스트

### 기존 SPEC 호환성 검증

**목표**: 기존에 작성된 SPEC 문서들이 새로운 검증 시스템과 호환되는지 확인

**테스트 대상**:
- `.moai/specs/SPEC-INIT-003/spec.md`
- `.moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md`
- 기타 모든 기존 SPEC 문서

**검증 항목**:
- [ ] 기존 SPEC이 검증을 통과하는지 확인
- [ ] 검증 실패 시 Migration 가이드 제공
- [ ] 경고 모드에서 정상 동작하는지 확인

---

## 사용자 수락 테스트 (UAT)

### 실사용 시나리오

**시나리오 1**: 처음으로 SPEC 작성하는 사용자
- 오류 메시지가 이해하기 쉬운지 확인
- 수정 방법이 명확한지 확인

**시나리오 2**: 여러 SPEC을 동시에 작성하는 사용자
- 중복 ID 검증이 정확한지 확인
- 의존성 관계가 올바르게 표시되는지 확인

**시나리오 3**: 기존 SPEC을 수정하는 사용자
- 버전 증가가 자동으로 제안되는지 확인
- HISTORY 섹션이 자동 업데이트되는지 확인

---

## 다음 단계

검증 완료 후:
```bash
/alfred:2-build VALID-001  # TDD 구현
/alfred:3-sync             # 문서 동기화
```

---

**문서 상태**: Draft
**마지막 업데이트**: 2025-10-16
