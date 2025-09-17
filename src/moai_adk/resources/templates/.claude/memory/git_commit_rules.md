# Git 커밋 메시지 규칙

## 형식
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## 타입
- feat/fix/docs/style/refactor/test/chore/ci/build/perf/revert

## 설명 규칙
- 명령형 현재형, 첫 글자 소문자, 끝 마침표 금지, ≤ 50자, 구체적

## 본문/푸터
- 본문 72자 줄바꿈, 무엇/왜를 설명
- BREAKING CHANGE, Closes #123, Co-authored-by 지원

## 예시
```
feat(auth): add OAuth2 Google login
fix: resolve memory leak in user session cleanup
docs(api): update authentication endpoints
```

## 체크리스트
- [ ] 타입 승인 목록 사용
- [ ] 50자 이하 요약
- [ ] 명령형 사용, 마침표 없음
- [ ] 맥락이 명확함
