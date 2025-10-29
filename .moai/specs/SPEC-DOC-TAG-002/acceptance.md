# @SPEC:DOC-TAG-002 인수 기준

## 개요

이 문서는 **Phase 2: @DOC 태그 자동 생성 - 에이전트 통합**의 인수 기준(Acceptance Criteria)을 정의합니다.

**테스트 범위**:
- Phase 1.5: TAG 할당 체크 (/alfred:3-sync)
- Phase 2.5: @DOC TAG 자동 생성 (doc-syncer)
- moai-foundation-tags Skill 업데이트
- CLI 유틸리티 (선택사항)

---

## 테스트 시나리오

### 시나리오 1: Phase 1.5 TAG 할당 체크

**GIVEN**: `/alfred:3-sync` 명령 실행, 태그 없는 10개 문서 파일 존재

**WHEN**: 시스템이 TAG 할당 체크 실행 (Phase 1.5)

**THEN**:
- ✅ "10개 파일에 @DOC TAG 제안됩니다" 메시지 표시
- ✅ 제안된 TAG 목록이 신뢰도별로 정렬됨 (HIGH → MEDIUM → LOW)
- ✅ 각 제안에 파일 경로, TAG ID, SPEC 매핑, 신뢰도 점수 포함
- ✅ 사용자가 승인/거부 선택 가능한 `AskUserQuestion` 표시
- ✅ 선택 옵션: Y (모두 생성), n (건너뛰기), h (HIGH만), m (수동 선택)

**검증 방법**:
```bash
# 태그 없는 10개 파일 생성
mkdir -p docs/auth docs/api docs/guides docs/misc
touch docs/auth/{login,register,password-reset}.md
touch docs/api/{endpoints,authentication}.md
touch docs/guides/{setup,troubleshooting}.md
touch docs/misc/{notes,old}.md
touch docs/README.md

# /alfred:3-sync 실행
/alfred:3-sync

# 예상 출력 확인:
# - "10개 파일에 @DOC TAG 제안됩니다"
# - HIGH: 7개, MEDIUM: 2개, LOW: 1개
# - AskUserQuestion 표시
```

**성공 기준**:
- 모든 태그 없는 파일 감지
- 신뢰도 점수 정확도 > 90%
- 사용자 상호작용 정상 작동

---

### 시나리오 2: Phase 2.5 자동 TAG 생성

**GIVEN**: doc-syncer Phase 2.5 실행, 사용자가 TAG 생성 승인 ("Y" 선택)

**WHEN**: 시스템이 `insert_tag_to_markdown()` 호출

**THEN**:
- ✅ 각 문서 파일 헤더에 @DOC 태그 삽입
- ✅ TAG 포맷 정확:
  - SPEC 매핑 있음: `# @DOC:DOMAIN-NNN | Chain: @SPEC:SOURCE-ID -> @DOC:DOMAIN-NNN`
  - SPEC 매핑 없음: `# @DOC:DOMAIN-NNN`
- ✅ 파일 백업 생성: `{file}.backup`
- ✅ 삽입 성공 후 원본 삭제, 백업 보관
- ✅ TAG 인벤토리 업데이트: `.moai/memory/tag-inventory.json`
- ✅ 삽입 성공/실패 로그 기록

**검증 방법**:
```bash
# Phase 2.5 실행 (Phase 1.5에서 "Y" 승인 후)
# doc-syncer Phase 2.5 자동 실행

# 파일 내용 확인
cat docs/auth/login.md
# 예상 출력:
# # @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001
#
# # 로그인 가이드
# ...

# 백업 파일 확인
ls docs/auth/login.md.backup

# TAG 인벤토리 확인
cat .moai/memory/tag-inventory.json | grep "DOC:AUTH-001"
# 예상 출력:
# {
#   "tag_id": "DOC:AUTH-001",
#   "file_path": "docs/auth/login.md",
#   "tag_type": "DOC",
#   "created_at": "2025-10-29T..."
# }
```

**성공 기준**:
- 100% TAG 삽입 성공 (HIGH 신뢰도 파일)
- 백업 파일 생성 확인
- TAG 인벤토리 정확성 100%
- 파일 포맷 손상 없음

---

### 시나리오 3: 사용자 거부 시 파일 무수정

**GIVEN**: TAG 제안 표시, 사용자가 거부 선택 ("n")

**WHEN**: 시스템이 거부 응답 처리

**THEN**:
- ✅ 어떤 문서 파일도 수정하지 않음
- ✅ Phase 2.5 생략
- ✅ 기존 워크플로우 계속 진행 (Phase 1 → Phase 2)
- ✅ "TAG 생성 건너뜀" 메시지 표시

**검증 방법**:
```bash
# 태그 없는 파일 생성
echo "# Test Doc" > docs/test.md

# /alfred:3-sync 실행, "n" 선택
/alfred:3-sync
# 입력: n

# 파일 내용 확인 (변경 없음)
cat docs/test.md
# 예상 출력:
# # Test Doc
# (TAG 없음)

# 백업 파일 없음 확인
ls docs/test.md.backup
# 예상 출력: No such file or directory
```

**성공 기준**:
- 파일 수정 없음 (100%)
- 백업 파일 생성 없음
- 워크플로우 정상 진행

---

### 시나리오 4: 신뢰도 점수 표시

**GIVEN**: 여러 SPEC-DOC 매핑 후보 존재

**WHEN**: `suggest_tag_for_file()` 신뢰도 점수 계산

**THEN**:
- ✅ 각 제안에 신뢰도 표시 (0.0~1.0 또는 HIGH/MEDIUM/LOW)
- ✅ HIGH (0.8-1.0): 도메인 완전 일치, 파일명 일치
- ✅ MEDIUM (0.5-0.8): 도메인 부분 일치, 키워드 일치
- ✅ LOW (<0.5): 매핑 불확실
- ✅ LOW 신뢰도는 사용자 확인 요청

**검증 방법**:
```bash
# HIGH 신뢰도 파일 생성
mkdir -p docs/auth
echo "# Login Guide" > docs/auth/login.md

# MEDIUM 신뢰도 파일 생성
mkdir -p docs/guides
echo "# Setup Guide" > docs/guides/setup.md

# LOW 신뢰도 파일 생성
mkdir -p docs/misc
echo "# Old Notes" > docs/misc/notes.md

# /alfred:3-sync 실행
/alfred:3-sync

# 예상 출력 확인:
# HIGH 신뢰도 (1개):
# - docs/auth/login.md → @DOC:AUTH-001 (SPEC:AUTH-001, 0.95)
#
# MEDIUM 신뢰도 (1개):
# - docs/guides/setup.md → @DOC:GUIDE-001 (0.65)
#
# LOW 신뢰도 (1개):
# - docs/misc/notes.md → @DOC:MISC-001 (0.32) [수동 개입 필요]
```

**성공 기준**:
- 신뢰도 점수 정확도 > 85%
- HIGH/MEDIUM/LOW 분류 정확
- LOW 신뢰도 경고 표시

---

### 시나리오 5: Skill 문서 완성도

**GIVEN**: moai-foundation-tags Skill 업데이트 완료

**WHEN**: 사용자가 @DOC TAG 관련 도움말 조회

**THEN**:
- ✅ SKILL.md에 명확한 예제와 설명
- ✅ @DOC TAG 포맷 및 체인 참조 섹션
- ✅ SPEC-DOC 매핑 예제
- ✅ 신뢰도 점수 설명
- ✅ Phase 1.5/2.5 통합 설명
- ✅ examples.md에 실제 사용 예시
- ✅ 문제 해결 섹션 포함 (고아 DOC, 중복 ID)

**검증 방법**:
```bash
# Skill 파일 확인
cat .claude/skills/moai-foundation-tags/SKILL.md | grep "@DOC TAG"
cat .claude/skills/moai-foundation-tags/examples.md | grep "예제"

# 섹션 존재 확인:
# - @DOC TAG: 문서 태그 자동 생성
# - @DOC TAG 포맷
# - SPEC-DOC 매핑
# - 신뢰도 점수 기준
# - Phase 1.5/2.5 통합
# - @DOC TAG 자동 생성 예제
# - 일반적인 문제 해결
```

**성공 기준**:
- 모든 필수 섹션 포함
- 예제 3개 이상
- 문제 해결 가이드 포함

---

### 시나리오 6: 백업 파일 관리

**GIVEN**: TAG 삽입 진행 중

**WHEN**: 파일 백업 생성 및 삽입 성공

**THEN**:
- ✅ 파일 수정 전 백업 생성: `{file}.backup`
- ✅ 삽입 성공 시 원본 삭제
- ✅ 백업 파일 보관 (추후 복구 가능)
- ✅ 삽입 실패 시 백업 복원

**검증 방법**:
```bash
# 백업 생성 확인
echo "# Test Doc" > docs/test.md
# TAG 삽입 실행 (내부적으로)
# 백업 확인
ls docs/test.md.backup

# 삽입 성공 시 원본 삭제 확인
cat docs/test.md
# 예상 출력: @DOC TAG 포함

# 백업 파일 존재 확인
cat docs/test.md.backup
# 예상 출력: 원본 내용 (TAG 없음)
```

**성공 기준**:
- 100% 백업 생성
- 삽입 성공 시 원본 삭제
- 삽입 실패 시 백업 복원

---

### 시나리오 7: TAG 인벤토리 업데이트

**GIVEN**: TAG 삽입 성공

**WHEN**: TAG 인벤토리 업데이트

**THEN**:
- ✅ `.moai/memory/tag-inventory.json` 파일 업데이트
- ✅ 각 TAG에 대한 메타데이터 기록:
  - `tag_id`: @DOC:DOMAIN-NNN
  - `file_path`: 파일 경로
  - `tag_type`: "DOC"
  - `created_at`: 생성 시간 (ISO 8601)
  - `chain_ref`: SPEC 체인 참조 (선택사항)
- ✅ 중복 TAG 방지 (기존 TAG 확인)

**검증 방법**:
```bash
# TAG 인벤토리 확인
cat .moai/memory/tag-inventory.json

# 예상 출력:
# {
#   "tags": [
#     {
#       "tag_id": "DOC:AUTH-001",
#       "file_path": "docs/auth/login.md",
#       "tag_type": "DOC",
#       "created_at": "2025-10-29T12:34:56Z",
#       "chain_ref": "@SPEC:AUTH-001 -> @DOC:AUTH-001"
#     },
#     ...
#   ]
# }

# 중복 TAG 테스트
# 동일한 TAG를 다시 삽입 시도
# 예상 결과: 중복 경고 표시, 삽입 중단
```

**성공 기준**:
- 100% TAG 인벤토리 업데이트
- 메타데이터 정확성 100%
- 중복 TAG 방지

---

### 시나리오 8: 기존 워크플로우 호환성

**GIVEN**: Phase 1.5/2.5 추가 후

**WHEN**: 기존 doc-syncer 및 /alfred:3-sync 워크플로우 실행

**THEN**:
- ✅ Phase 0 (프로젝트 분석) 정상 작동
- ✅ Phase 1.5 (TAG 할당 체크) 추가, 선택적 실행
- ✅ Phase 1 (doc-syncer 호출) 정상 작동
- ✅ Phase 2.5 (@DOC TAG 생성) 추가, 선택적 실행
- ✅ Phase 2 (TAG 검증) 정상 작동
- ✅ Phase 3 (최종 리포트) 정상 작동
- ✅ 사용자가 Phase 1.5/2.5 거부 시 기존 플로우 그대로 진행

**검증 방법**:
```bash
# 기존 워크플로우 테스트 (Phase 1.5/2.5 생략)
/alfred:3-sync
# 입력: n (TAG 생성 거부)

# 예상 결과:
# - Phase 0 실행
# - Phase 1.5 실행 → 사용자 거부
# - Phase 1 실행 (Phase 2.5 생략)
# - Phase 2 실행
# - Phase 3 실행

# 전체 워크플로우 테스트 (Phase 1.5/2.5 포함)
/alfred:3-sync
# 입력: Y (TAG 생성 승인)

# 예상 결과:
# - Phase 0 실행
# - Phase 1.5 실행 → 사용자 승인
# - Phase 1 실행 (Phase 2.5 포함)
# - Phase 2 실행
# - Phase 3 실행
```

**성공 기준**:
- 기존 워크플로우 변경 없음
- Phase 1.5/2.5 선택적 실행
- 후진 호환성 100%

---

### 시나리오 9: CLI 유틸리티 - tag-generate (선택사항)

**GIVEN**: CLI 유틸리티 설치 완료

**WHEN**: `moai-adk tag-generate docs/` 실행

**THEN**:
- ✅ 태그 없는 파일 스캔
- ✅ TAG 제안 생성 (신뢰도 점수 포함)
- ✅ 신뢰도별 분류 표시 (HIGH/MEDIUM/LOW)
- ✅ 사용자 승인 요청
- ✅ 승인 시 TAG 삽입
- ✅ 삽입 성공/실패 요약 표시

**검증 방법**:
```bash
# CLI 명령 실행
moai-adk tag-generate docs/

# 예상 출력:
# Found 10 untagged files
#
# HIGH confidence (7):
#   - docs/auth/login.md → @DOC:AUTH-001 (SPEC:AUTH-001, 0.95)
#   ...
#
# MEDIUM confidence (2):
#   - docs/guides/setup.md → @DOC:GUIDE-001 (0.65)
#   ...
#
# LOW confidence (1):
#   - docs/misc/notes.md → @DOC:MISC-001 (0.32) [manual intervention needed]
#
# Generate TAGs for all files? [Y/n]: Y
#
# ✅ TAG generation completed: 10/10
```

**성공 기준**:
- CLI 명령 정상 작동
- 출력 형식 일관성
- 승인/거부 기능

---

### 시나리오 10: CLI 유틸리티 - tag-validate (선택사항)

**GIVEN**: CLI 유틸리티 설치 완료, 일부 문서에 @DOC TAG 존재

**WHEN**: `moai-adk tag-validate docs/` 실행

**THEN**:
- ✅ 태그된 파일 스캔
- ✅ 고아 DOC 탐지 (관련 SPEC 없음)
- ✅ 중복 TAG 탐지
- ✅ 검증 리포트 표시

**검증 방법**:
```bash
# 고아 DOC 생성
echo "# @DOC:ORPHAN-001\n\n# Orphaned Doc" > docs/orphan.md

# 중복 TAG 생성
echo "# @DOC:AUTH-001\n\n# Duplicate Doc" > docs/duplicate.md

# CLI 명령 실행
moai-adk tag-validate docs/

# 예상 출력:
# Found 2 tagged files
#
# ⚠️  Orphaned DOCs (1):
#   - @DOC:ORPHAN-001 (docs/orphan.md)
#     → No related SPEC found
#
# ❌ Duplicate TAGs (1):
#   - @DOC:AUTH-001
#     - docs/auth/login.md
#     - docs/duplicate.md
```

**성공 기준**:
- 고아 DOC 탐지 정확도 100%
- 중복 TAG 탐지 정확도 100%
- 검증 리포트 명확성

---

### 시나리오 11: CLI 유틸리티 - tag-map (선택사항)

**GIVEN**: CLI 유틸리티 설치 완료, SPEC-DOC 매핑 존재

**WHEN**: `moai-adk tag-map AUTH-001 docs/` 실행

**THEN**:
- ✅ 관련 DOC 파일 표시
- ✅ 각 DOC의 TAG ID, 신뢰도, 체인 참조 표시
- ✅ SPEC-DOC 매핑 요약

**검증 방법**:
```bash
# CLI 명령 실행
moai-adk tag-map AUTH-001 docs/

# 예상 출력:
# SPEC-DOC mapping for @SPEC:AUTH-001:
#   - docs/auth/login.md
#     TAG: @DOC:AUTH-001
#     Confidence: 0.95
#     Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001
#   - docs/auth/register.md
#     TAG: @DOC:AUTH-002
#     Confidence: 0.92
#     Chain: @SPEC:AUTH-001 -> @DOC:AUTH-002
```

**성공 기준**:
- SPEC-DOC 매핑 정확도 100%
- 출력 형식 일관성
- 신뢰도 점수 표시

---

## 성공 기준 (Definition of Done)

### 통합 검증

- ✅ **doc-syncer Phase 2.5 완전 작동**
  - 태그 없는 파일 스캔
  - TAG ID 생성
  - SPEC-DOC 매핑
  - TAG 삽입 (백업 생성)
  - TAG 인벤토리 업데이트

- ✅ **/alfred:3-sync Phase 1.5 TAG 체크 기능**
  - 태그 없는 파일 확인
  - TAG 제안 생성
  - 신뢰도별 분류
  - 사용자 상호작용 (`AskUserQuestion`)

- ✅ **사용자 승인 전 파일 수정 안 함**
  - 사용자 거부 시 파일 무수정
  - Phase 2.5 생략 옵션

- ✅ **TAG 삽입 성공 및 포맷 정확**
  - SPEC 체인 참조 포함
  - 백업 파일 생성/관리
  - TAG 인벤토리 업데이트

- ✅ **파일 백업 생성 및 보관**
  - 삽입 전 백업 생성
  - 삽입 성공 시 원본 삭제, 백업 보관
  - 삽입 실패 시 백업 복원

---

### 워크플로우 호환성

- ✅ **기존 doc-syncer 기능 변경 없음**
  - Phase 1 (상태 분석) 정상 작동
  - Phase 2 (문서 동기화) 정상 작동
  - Phase 3 (품질 검증) 정상 작동

- ✅ **기존 /alfred:3-sync 단계 그대로 유지**
  - Phase 0 (프로젝트 분석) 정상 작동
  - Phase 2 (TAG 검증) 정상 작동
  - Phase 3 (최종 리포트) 정상 작동

- ✅ **후진 호환성 100%**
  - Phase 1.5/2.5 생략 시 기존 플로우 그대로
  - 이전 SPEC 실행 가능

---

### 문서화

- ✅ **Skill 업데이트 완료**
  - SKILL.md에 @DOC 섹션 추가
  - examples.md에 @DOC 예제 추가
  - Phase 1.5/2.5 통합 설명

- ✅ **@DOC TAG 예제 추가**
  - 도메인 기반 매핑 예제
  - 신뢰도 낮은 경우 수동 매핑 예제
  - /alfred:3-sync TAG 제안 플로우 예제

- ✅ **통합 가이드 작성**
  - Phase 1.5/2.5 사용 가이드
  - 문제 해결 가이드 (고아 DOC, 중복 ID)

---

### CLI 유틸리티 (선택사항)

- ✅ **tag-generate 명령 기능**
  - 태그 없는 파일 스캔
  - TAG 제안 생성 (신뢰도 점수)
  - 사용자 승인 후 TAG 삽입

- ✅ **tag-validate 명령 기능**
  - 고아 DOC 탐지
  - 중복 TAG 탐지
  - 검증 리포트 생성

- ✅ **tag-map 명령 기능**
  - SPEC-DOC 매핑 표시
  - 신뢰도 점수 표시
  - 체인 참조 표시

- ✅ **CLI 테스트 통과**
  - 단위 테스트 90% 이상 커버리지
  - 통합 테스트 100% 통과

---

## Phase 2 완료 체크리스트

### 필수 항목

- [ ] **doc-syncer Phase 2.5 구현**
  - [ ] Phase 2.5 섹션 추가 (`.claude/agents/alfred/doc-syncer.md`)
  - [ ] 워크플로우 다이어그램 업데이트
  - [ ] 태그 없는 파일 스캔 기능
  - [ ] TAG ID 생성 기능
  - [ ] SPEC-DOC 매핑 기능
  - [ ] TAG 삽입 기능 (백업 생성)
  - [ ] TAG 인벤토리 업데이트 기능

- [ ] **/alfred:3-sync Phase 1.5 구현**
  - [ ] Phase 1.5 섹션 추가 (`.claude/commands/alfred/3-sync.md`)
  - [ ] 워크플로우 다이어그램 업데이트
  - [ ] TAG 제안 생성 기능
  - [ ] 신뢰도별 분류 기능
  - [ ] 사용자 상호작용 (`AskUserQuestion`)

- [ ] **moai-foundation-tags Skill 업데이트**
  - [ ] SKILL.md에 @DOC 섹션 추가
  - [ ] examples.md에 @DOC 예제 추가
  - [ ] Phase 1.5/2.5 통합 설명
  - [ ] 문제 해결 가이드

- [ ] **통합 테스트**
  - [ ] Phase 1.5 단위 테스트
  - [ ] Phase 2.5 단위 테스트
  - [ ] Phase 1.5 → Phase 2.5 통합 테스트
  - [ ] /alfred:3-sync E2E 테스트
  - [ ] 기존 워크플로우 호환성 테스트

- [ ] **문서화**
  - [ ] Skill 문서 업데이트 확인
  - [ ] 예제 3개 이상 추가 확인
  - [ ] 문제 해결 섹션 확인

---

### 선택 항목

- [ ] **CLI 유틸리티 개발**
  - [ ] `tag_commands.py` 구현 (~200 LOC)
  - [ ] `tag-generate` 명령
  - [ ] `tag-validate` 명령
  - [ ] `tag-map` 명령

- [ ] **CLI 테스트 작성**
  - [ ] `test_tag_commands.py` 구현 (~150 LOC)
  - [ ] `tag-generate` 테스트
  - [ ] `tag-validate` 테스트
  - [ ] `tag-map` 테스트

- [ ] **CLI 문서화**
  - [ ] CLI 참고 문서 작성 (`docs/cli-reference.md`)
  - [ ] 사용 예제 추가

---

### 품질 검증

- [ ] **코드 품질**
  - [ ] ruff 린팅 통과
  - [ ] mypy 타입 체크 통과
  - [ ] 테스트 커버리지 90% 이상

- [ ] **코드 리뷰**
  - [ ] doc-syncer 수정 리뷰 완료
  - [ ] /alfred:3-sync 수정 리뷰 완료
  - [ ] Skill 업데이트 리뷰 완료

- [ ] **E2E 테스트**
  - [ ] /alfred:3-sync로 전체 플로우 테스트
  - [ ] TAG 생성 → 삽입 → 인벤토리 업데이트 검증
  - [ ] 기존 워크플로우 호환성 검증

---

### 최종 확인

- [ ] **모든 변경사항 커밋**
  - [ ] PR #103 또는 새 PR 생성
  - [ ] 커밋 메시지 형식 준수
  - [ ] @SPEC:DOC-TAG-002 태그 포함

- [ ] **기존 워크플로우 기능 확인**
  - [ ] Phase 1.5/2.5 생략 시 기존 플로우 정상 작동
  - [ ] 후진 호환성 100%

- [ ] **문서 업데이트 확인**
  - [ ] Skill 문서 완성도 100%
  - [ ] 예제 명확성 100%
  - [ ] 문제 해결 가이드 완성

---

## Definition of Done (완료 정의)

Phase 2가 완료되려면 다음 기준을 모두 충족해야 합니다:

### 1. 기능 완성도

- ✅ **모든 테스트 통과**
  - 단위 테스트 100% 통과
  - 통합 테스트 100% 통과
  - E2E 테스트 100% 통과

- ✅ **ruff 린팅 통과**
  - 코드 스타일 준수
  - 타입 힌트 100%

- ✅ **코드 리뷰 완료**
  - doc-syncer 수정 승인
  - /alfred:3-sync 수정 승인
  - Skill 업데이트 승인

### 2. 문서화

- ✅ **문서 업데이트**
  - Skill 문서 완성
  - 예제 3개 이상
  - 문제 해결 가이드

### 3. 품질 검증

- ✅ **/alfred:3-sync로 E2E 테스트**
  - TAG 제안 생성 검증
  - TAG 삽입 검증
  - 인벤토리 업데이트 검증
  - 기존 워크플로우 호환성 검증

### 4. 사용자 승인

- ✅ **실제 프로젝트에서 테스트**
  - MoAI-ADK 프로젝트 자체에서 /alfred:3-sync 실행
  - TAG 생성 결과 검증
  - 사용자 경험 만족도 확인

---

**완료 조건**: 위 모든 체크리스트 항목이 ✅로 체크되어야 Phase 2 완료

---

**END OF ACCEPTANCE CRITERIA**
