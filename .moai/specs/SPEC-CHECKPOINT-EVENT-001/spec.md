---
id: CHECKPOINT-EVENT-001
version: 0.1.0
status: completed
created: 2025-10-15
updated: 2025-10-16
author: @Goos
priority: high
category: feature
labels:
  - checkpoint
  - safety-net
  - git
  - event-driven
---

# @SPEC:CHECKPOINT-EVENT-001: Event-Driven Checkpoint 시스템

## HISTORY

### v0.1.0 (2025-10-16)
- **COMPLETED**: Event-Driven Checkpoint 시스템 구현 완료
- **IMPLEMENTED**: checkpoint.py, event_detector.py, branch_manager.py
- **TESTED**: 테스트 커버리지 85% 달성
- **VERIFIED**: @CODE 태그 체인 무결성 확인
- **AUTHOR**: @Goos
- **COMMITS**:
  - 3b8c7bc: 🟢 GREEN: Claude Code Hooks 기반 구현
  - c3c48ac: 📝 DOCS: 문서 동기화

### v0.0.1 (2025-10-15)
- **INITIAL**: Event-Driven Checkpoint 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: 기존 시간 기반(5분 간격) checkpoint → 이벤트 기반 전환으로 불필요한 태그 생성 방지

---

## 개요

AI가 실수로 코드를 삭제하거나 잘못된 변경을 했을 때 빠르게 복구할 수 있도록 **위험한 작업 전에만** checkpoint를 생성하는 시스템.

**핵심 철학**:
- 시간이 아닌 "위험한 작업 전"에만 생성
- 태그 대신 local branch 사용 (원격 오염 방지)
- Git reflog를 1차 안전망으로 활용

---

## Ubiquitous Requirements (기본 요구사항)

- 시스템은 위험한 작업 전에 자동으로 checkpoint를 생성해야 한다
- 시스템은 checkpoint를 local branch로 관리해야 한다
- 시스템은 checkpoint의 원격 저장소 push를 차단해야 한다
- 시스템은 최대 10개의 checkpoint를 유지해야 한다

---

## Event-driven Requirements (이벤트 기반)

### 위험 작업 정의

**WHEN** 다음 작업이 감지되면, 시스템은 checkpoint를 생성해야 한다:

1. **대규모 파일 삭제**
   - WHEN 10개 이상의 파일이 삭제될 때
   - WHEN `git rm` 명령이 실행될 때
   - WHEN `rm -rf` 패턴이 감지될 때

2. **복잡한 리팩토링**
   - WHEN 클래스 분리 작업이 시작될 때
   - WHEN 함수 추출/이동이 대규모로 발생할 때
   - WHEN 파일 이름 변경이 10개 이상 발생할 때

3. **병합 작업**
   - WHEN `git merge` 명령이 실행될 때
   - WHEN 충돌 해결이 필요할 때

4. **외부 스크립트 실행**
   - WHEN 사용자 정의 스크립트가 실행될 때
   - WHEN Bash tool로 복잡한 명령어가 실행될 때

5. **중요 파일 수정**
   - WHEN `CLAUDE.md`, `config.json` 등 중요 파일이 수정될 때
   - WHEN `.moai/memory/*.md` 파일이 수정될 때

### Checkpoint 생성 규칙

**WHEN** checkpoint 생성이 트리거되면:
- 시스템은 `before-{operation}-{timestamp}` 형식의 local branch를 생성해야 한다
- 시스템은 현재 working directory 상태를 커밋해야 한다
- 시스템은 checkpoint 메타데이터를 `.moai/checkpoints.log`에 기록해야 한다

**예시**:
```
before-delete-20251015-143000
before-refactor-processor-20251015-143500
before-merge-feature-20251015-144000
```

---

## State-driven Requirements (상태 기반)

**WHILE** checkpoint가 10개를 초과하면:
- 시스템은 가장 오래된 checkpoint부터 자동 삭제해야 한다
- 시스템은 삭제 전에 7일 이상 경과한 checkpoint만 삭제해야 한다

**WHILE** checkpoint 복구 작업 중일 때:
- 시스템은 복구 전에 현재 상태를 새로운 checkpoint로 저장해야 한다
- 시스템은 복구 후 변경사항을 명확히 보고해야 한다

---

## Optional Features (선택적 기능)

**WHERE** 사용자가 명시적으로 요청하면:
- 시스템은 수동 checkpoint 생성을 허용할 수 있다
- 시스템은 checkpoint 목록 조회 기능을 제공할 수 있다
- 시스템은 특정 checkpoint로 복구 기능을 제공할 수 있다

---

## Constraints (제약사항)

- **IF** checkpoint가 생성되면, 원격 저장소로 push되어서는 안 된다
- **IF** working directory가 clean하지 않으면, checkpoint 생성 전에 변경사항을 커밋해야 한다
- **IF** checkpoint 복구가 실패하면, 시스템은 이전 상태로 롤백하고 에러를 보고해야 한다
- checkpoint branch 이름은 `before-*` 접두사를 가져야 한다
- checkpoint는 local branch로만 존재해야 하며, remote tracking branch를 가져서는 안 된다

---

## 기술 스택

- **언어**: Python 3.11+
- **Git 라이브러리**: GitPython
- **설정 파일**: `.moai/config.json`
- **로그 파일**: `.moai/checkpoints.log`

---

## 파일 구조

```
src/moai_adk/core/git/
├── checkpoint.py              # Checkpoint 생성/복구/관리
├── event_detector.py          # 위험 작업 감지
└── branch_manager.py          # Local branch 관리

tests/unit/
├── test_checkpoint.py         # Checkpoint 테스트
├── test_event_detector.py     # 이벤트 감지 테스트
└── test_branch_manager.py     # Branch 관리 테스트

.moai/
├── config.json                # 설정 (event-driven 활성화)
└── checkpoints.log            # Checkpoint 이력
```

---

## 성공 기준

- [ ] 10개 이상 파일 삭제 시 자동 checkpoint 생성
- [ ] 대규모 리팩토링 시 자동 checkpoint 생성
- [ ] Checkpoint는 local branch로만 존재 (원격 push 차단)
- [ ] 최대 10개 checkpoint 유지 (FIFO)
- [ ] Checkpoint 복구 기능 동작
- [ ] 테스트 커버리지 85% 이상

---

## 참고 문서

- Git Reflog: https://git-scm.com/docs/git-reflog
- GitPython: https://gitpython.readthedocs.io/
- MoAI-ADK development-guide.md

---

**작성**: 2025-10-15
**작성자**: @Goos
**우선순위**: High
