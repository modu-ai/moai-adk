# SPEC-INSTALLER-REFACTOR-001 인수 기준

## ✅ Acceptance Criteria

### AC-001: phase-executor.ts LOC 제한 준수

```gherkin
Given: phase-executor.ts가 358 LOC일 때
When: BackupManager, DirectoryBuilder, GitInitializer로 분리
Then: phase-executor.ts ≤ 200 LOC
  And: 각 신규 클래스 ≤ 100 LOC
  And: 모든 함수 ≤ 50 LOC
```

---

### AC-002: template-processor.ts LOC 제한 준수

```gherkin
Given: template-processor.ts가 371 LOC일 때
When: TemplatePathResolver, TemplateRenderer로 분리
Then: template-processor.ts ≤ 200 LOC
  And: TemplatePathResolver ≤ 150 LOC
  And: TemplateRenderer ≤ 100 LOC
```

---

### AC-003: 의존성 주입 패턴 적용

```gherkin
Given: 새로운 클래스가 다른 클래스에 의존할 때
When: 생성자를 통해 의존성 주입
Then: mock 객체로 대체 가능해야 한다
  And: 테스트에서 DI를 활용할 수 있어야 한다
```

---

### AC-004: 기존 테스트 통과

```gherkin
Given: 리팩토링 완료 후
When: 기존 테스트 스위트 실행
Then: 모든 테스트가 통과해야 한다
  And: 기존 API 동작이 변경되지 않아야 한다
```

---

### AC-005: 성능 저하 제한

```gherkin
Given: 리팩토링 전후 성능 측정
When: 동일한 설치 작업 수행
Then: 성능 저하 < 5%
  And: 메모리 사용량 증가 < 10MB
```

---

### AC-006: 테스트 커버리지 유지

```gherkin
Given: 신규 클래스 생성 완료
When: 테스트 커버리지 측정
Then: 각 클래스 커버리지 ≥ 85%
  And: 전체 커버리지가 감소하지 않아야 한다
```

---

## 📊 테스트 커버리지 목표

| 파일 | 목표 |
|------|------|
| backup-manager.ts | ≥ 85% |
| directory-builder.ts | ≥ 85% |
| git-initializer.ts | ≥ 85% |
| template-path-resolver.ts | ≥ 85% |
| template-renderer.ts | ≥ 85% |

---

## ✅ 전체 체크리스트

### 기능 요구사항
- [ ] AC-001: phase-executor.ts LOC ≤ 200
- [ ] AC-002: template-processor.ts LOC ≤ 200
- [ ] AC-003: DI 패턴 적용
- [ ] AC-004: 기존 테스트 통과
- [ ] AC-005: 성능 저하 < 5%
- [ ] AC-006: 커버리지 ≥ 85%

### LOC 제한
- [ ] 모든 파일 ≤ 300 LOC
- [ ] 모든 함수 ≤ 50 LOC
- [ ] 모든 함수 매개변수 ≤ 5개

---

## 🎯 인수 승인 기준

**모든 AC 통과** + **LOC 제한 준수** + **테스트 커버리지 85% 이상**
