# SPEC-CORE-GIT-001 수락 기준

## Given-When-Then 테스트 시나리오

### 시나리오 1: Git 저장소 확인
**Given**: Git 저장소 내에 있음
**When**: `GitManager().is_repo()` 호출
**Then**:
- [ ] True 반환
- [ ] 에러 없이 실행

### 시나리오 2: 브랜치 생성
**Given**: develop 브랜치가 존재함
**When**: `create_branch("feature/SPEC-AUTH-001", "develop")` 호출
**Then**:
- [ ] 새 브랜치가 생성됨
- [ ] 현재 브랜치가 전환됨
- [ ] `git branch` 명령으로 확인 가능

### 시나리오 3: 커밋 생성
**Given**: 변경된 파일이 있음
**When**: `commit("🔴 RED: Add test", [".moai/specs/SPEC-AUTH-001/spec.md"])` 호출
**Then**:
- [ ] 파일이 스테이징됨
- [ ] 커밋이 생성됨
- [ ] `git log -1` 명령으로 확인 가능

### 시나리오 4: TDD 커밋 메시지 (한국어)
**Given**: locale이 "ko"로 설정됨
**When**: `format_commit_message("red", "테스트 추가", "ko")` 호출
**Then**:
- [ ] "🔴 RED: 테스트 추가" 반환

### 시나리오 5: Draft PR 생성
**Given**: feature 브랜치가 푸시됨
**When**: `create_draft_pr("SPEC-AUTH-001", "인증 시스템 구현")` 호출
**Then**:
- [ ] gh CLI가 실행됨
- [ ] PR URL이 반환됨
- [ ] GitHub에서 Draft PR 확인 가능

---

## 품질 게이트 기준

### 1. Git 작업 완성도
- [ ] 브랜치 생성/전환 성공
- [ ] 커밋 및 푸시 성공
- [ ] PR 생성 성공

### 2. 에러 처리
- [ ] 잘못된 브랜치명 처리
- [ ] Git 저장소 아닐 때 처리
- [ ] gh CLI 없을 때 처리

### 3. locale 지원
- [ ] 한국어 커밋 메시지
- [ ] 영어 커밋 메시지
- [ ] 기본값 fallback

---

## 검증 방법 및 도구

### 자동화 테스트
```python
# tests/core/git/test_manager.py
import pytest
from moai_adk.core.git import GitManager

def test_is_repo(tmp_path):
    # Git 저장소 초기화
    (tmp_path / ".git").mkdir()
    manager = GitManager(str(tmp_path))
    assert manager.is_repo() == True

def test_commit_message_format():
    msg = format_commit_message("red", "Add test", "ko")
    assert msg == "🔴 RED: Add test"
```

### 수동 검증
1. **브랜치 생성**: `git branch` 확인
2. **커밋 로그**: `git log -1 --oneline` 확인
3. **PR 생성**: GitHub UI에서 Draft PR 확인

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] GitManager 클래스 구현
- [ ] 브랜치/커밋 유틸리티 구현
- [ ] TDD 커밋 메시지 포맷팅
- [ ] 모든 테스트 통과

### 선택 조건
- [ ] PR 템플릿 자동 적용
- [ ] Git hooks 설정

### 문서화
- [ ] GitManager API 문서
- [ ] 예제 코드 추가
