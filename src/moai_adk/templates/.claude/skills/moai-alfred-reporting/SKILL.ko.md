---
name: moai-alfred-reporting
description: 보고서 생성 표준, 출력 형식 규칙, 하위 에이전트 보고서 예시
tier: alfred
freedom: medium
tags: [reporting, formatting, documentation, output, style]
---

# 보고서 스타일

**핵심 규칙**: 화면 출력(사용자 대면)과 내부 문서(파일)를 구분하십시오.

## 출력 형식 규칙

- **사용자 화면 출력**: 일반 텍스트 (마크다운 문법 사용 금지)
- **내부 문서** (`.moai/docs/`, `.moai/reports/` 파일): 마크다운 형식
- **코드 주석 및 Git 커밋**: 사용자 구성 언어, 명확한 구조

## 사용자 화면 출력 (일반 텍스트)

**사용자와 직접 대화/프롬프트 시:**

일반 텍스트 형식을 사용하십시오 (마크다운 헤더, 표, 특수 서식 금지):

예시:
```
병합 충돌 감지됨:

근본 원인:
- 커밋 c054777b가 develop에서 언어 감지를 제거함
- 병합 커밋 e18c7f98(develop → main)이 해당 라인을 다시 도입함

영향 범위:
- .claude/hooks/alfred/shared/handlers/session.py
- src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py

제안 조치:
- detect_language() import 및 호출 제거
- 언어 표시 라인 삭제
- 두 파일 동기화
```

## 내부 문서 (마크다운 형식)

**`.moai/docs/`, `.moai/reports/`, `.moai/analysis/`에 파일 생성 시:**

적절한 구조의 마크다운 형식을 사용하십시오:

```markdown
## 🎊 작업 완료 보고서

### 구현 결과
- ✅ 기능 A 구현 완료
- ✅ 테스트 작성 및 통과
- ✅ 문서 동기화 완료

### 품질 지표
| 항목 | 결과 |
|------|--------|
| 테스트 커버리지 | 95% |
| 린팅 | 통과 |

### 다음 단계
1. `/alfred:3-sync` 실행
2. PR 생성 및 검토
3. main 브랜치로 병합
```

## ❌ 금지된 보고서 출력 패턴

**다음 방식으로 보고서를 포장하지 마십시오:**

```bash
# ❌ 잘못된 예시 1: Bash 명령어 포장
cat << 'EOF'
## 보고서
...내용...
EOF

# ❌ 잘못된 예시 2: Python 포장
python -c "print('''
## 보고서
...내용...
''')"

# ❌ 잘못된 예시 3: echo 사용
echo "## 보고서"
echo "...내용..."
```

## 보고서 작성 가이드라인

### 1. 마크다운 형식
- 섹션 분리를 위해 헤더(`##`, `###`) 사용
- 구조화된 정보를 표로 표시
- 항목을 불릿 포인트로 나열
- 상태 표시기를 위해 이모지 사용 (✅, ❌, ⚠️, 🎊, 📊)

### 2. 보고서 길이 관리
- 짧은 보고서(<500 문자): 한 번에 출력
- 긴 보고서(>500 문자): 섹션별로 분할
- 요약으로 시작, 세부 사항으로 이어짐

### 3. 구조화된 섹션
```markdown
## 🎯 핵심 성과
- 핵심 성취 사항

## 📊 통계 요약
| 항목 | 결과 |

## ⚠️ 중요 참고사항
- 사용자가 알아야 할 정보

## 🚀 다음 단계
1. 추천 조치
```

### 4. 언어 설정
- 사용자의 `conversation_language` 사용
- 코드/기술 용어는 영어로 유지
- 설명/안내는 사용자 언어로 사용

## 하위 에이전트 보고서 예시

### spec-builder (SPEC 생성 완료)
```markdown
## 📋 SPEC 생성 완료

### 생성된 문서
- ✅ `.moai/specs/SPEC-XXX-001/spec.md`
- ✅ `.moai/specs/SPEC-XXX-001/plan.md`
- ✅ `.moai/specs/SPEC-XXX-001/acceptance.md`

### EARS 검증 결과
- ✅ 모든 요구사항이 EARS 형식을 준수함
- ✅ @TAG 체인 생성됨
```

### tdd-implementer (구현 완료)
```markdown
## 🚀 TDD 구현 완료

### 구현 파일
- ✅ `src/feature.py` (코드 작성)
- ✅ `tests/test_feature.py` (테스트 작성)

### 테스트 결과
| 단계 | 상태 |
|-------|--------|
| RED | ✅ 실패 확인됨 |
| GREEN | ✅ 구현 성공 |
| REFACTOR | ✅ 리팩토링 완료 |

### 품질 지표
- 테스트 커버리지: 95%
- 린팅: 0개 이슈
```

### doc-syncer (문서 동기화 완료)
```markdown
## 📚 문서 동기화 완료

### 업데이트된 문서
- ✅ `README.md` - 사용 예시 추가됨
- ✅ `.moai/docs/architecture.md` - 구조 업데이트됨
- ✅ `CHANGELOG.md` - v0.8.0 항목 추가됨

### @TAG 검증
- ✅ SPEC → CODE 연결 검증됨
- ✅ CODE → TEST 연결 검증됨
- ✅ TEST → DOC 연결 검증됨
```

## 적용 시점

**다음 순간에 보고서를 직접 출력해야 함:**

1. **명령 완료** (항상)
   - `/alfred:0-project` 완료
   - `/alfred:1-plan` 완료
   - `/alfred:2-run` 완료
   - `/alfred:3-sync` 완료

2. **하위 에이전트 작업 완료** (대부분)
   - spec-builder: SPEC 생성 완료
   - tdd-implementer: 구현 완료
   - doc-syncer: 문서 동기화 완료
   - tag-agent: TAG 검증 완료

3. **품질 검증 완료**
   - TRUST 5 검증 통과
   - 테스트 실행 완료
   - 린팅/타입 검사 통과

4. **Git 작업 완료**
   - 커밋 생성 후
   - PR 생성 후
   - 병합 완료 후

**예외: 보고서가 필요 없는 경우**
- 간단한 조회/읽기 작업
- 중간 단계 (불완료 작업)
- 사용자가 "빠르게"라고 명시적으로 요청할 때

## Bash 도구 사용 예외

**Bash 도구는 다음 경우에만 허용됨:**

1. **실제 시스템 명령어**
   - 파일 작업 (`touch`, `mkdir`, `cp`)
   - Git 작업 (`git add`, `git commit`, `git push`)
   - 패키지 설치 (`pip`, `npm`, `uv`)
   - 테스트 실행 (`pytest`, `npm test`)

2. **환경 설정**
   - 권한 변경 (`chmod`)
   - 환경 변수 (`export`)
   - 디렉토리 이동 (`cd`)

3. **정보 조회 (파일 내용 제외)**
   - 시스템 정보 (`uname`, `df`)
   - 프로세스 상태 (`ps`, `top`)
   - 네트워크 상태 (`ping`, `curl`)

**파일 내용에는 Read 도구 사용:**
```markdown
❌ Bash: cat file.txt
✅ Read: Read(file_path="/absolute/path/file.txt")
```